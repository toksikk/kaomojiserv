package main

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type responseRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *responseRecorder) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseRecorder) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(p)
	w.bytes += n
	return n, err
}

type httpMetrics struct {
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
	response *prometheus.HistogramVec
	inFlight *prometheus.GaugeVec
}

func newHTTPMetrics(reg prometheus.Registerer) *httpMetrics {
	m := &httpMetrics{
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "kaomojiserv_http_requests_total",
			Help: "Total number of HTTP requests.",
		}, []string{"method", "route", "status"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "kaomojiserv_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "route"}),
		response: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "kaomojiserv_http_response_size_bytes",
			Help:    "HTTP response size in bytes.",
			Buckets: prometheus.ExponentialBuckets(100, 10, 7),
		}, []string{"method", "route"}),
		inFlight: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "kaomojiserv_http_requests_in_flight",
			Help: "Current number of HTTP requests being served.",
		}, []string{"method", "route"}),
	}
	reg.MustRegister(m.requests, m.duration, m.response, m.inFlight)
	return m
}

func observeHTTP(logger *slog.Logger, metrics *httpMetrics, next *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		_, route := next.Handler(r)
		if route == "" {
			route = "unmatched"
		}

		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = newRequestID()
		}
		w.Header().Set("X-Request-ID", requestID)

		recorder := &responseRecorder{ResponseWriter: w}
		metrics.inFlight.WithLabelValues(r.Method, route).Inc()
		defer func() {
			metrics.inFlight.WithLabelValues(r.Method, route).Dec()
			status := recorder.status
			if status == 0 {
				status = http.StatusOK
			}
			duration := time.Since(started)
			statusText := strconv.Itoa(status)
			metrics.requests.WithLabelValues(r.Method, route, statusText).Inc()
			metrics.duration.WithLabelValues(r.Method, route).Observe(duration.Seconds())
			metrics.response.WithLabelValues(r.Method, route).Observe(float64(recorder.bytes))
			logger.Info("http access",
				"request_id", requestID,
				"method", r.Method,
				"route", route,
				"path", r.URL.Path,
				"query", r.URL.RawQuery,
				"status", status,
				"bytes", recorder.bytes,
				"duration_ms", float64(duration.Microseconds())/1000,
				"remote_addr", r.RemoteAddr,
				"forwarded_for", r.Header.Get("X-Forwarded-For"),
				"user_agent", r.UserAgent(),
			)
		}()

		next.ServeHTTP(recorder, r)
	})
}

func newRequestID() string {
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return hex.EncodeToString(id[:])
}

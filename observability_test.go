package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"mime"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

func TestObserveHTTPLogsAccessAndRecordsMetrics(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	registry := prometheus.NewRegistry()
	metrics := newHTTPMetrics(registry)
	mux := http.NewServeMux()
	mux.HandleFunc("/api", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write([]byte("response")); err != nil {
			t.Fatal(err)
		}
	})

	request := httptest.NewRequest(http.MethodGet, "/api?limit=2", nil)
	request.Header.Set("X-Forwarded-For", "192.0.2.1")
	request.Header.Set("X-Request-ID", "test-request")
	request.Header.Set("User-Agent", "observability-test")
	response := httptest.NewRecorder()
	observeHTTP(logger, metrics, mux).ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusCreated)
	}
	if got := response.Header().Get("X-Request-ID"); got != "test-request" {
		t.Fatalf("X-Request-ID = %q, want test-request", got)
	}

	var entry map[string]any
	if err := json.Unmarshal(logs.Bytes(), &entry); err != nil {
		t.Fatalf("decode access log: %v", err)
	}
	for key, want := range map[string]any{
		"msg":           "http access",
		"request_id":    "test-request",
		"method":        http.MethodGet,
		"route":         "/api",
		"path":          "/api",
		"query":         "limit=2",
		"status":        float64(http.StatusCreated),
		"bytes":         float64(len("response")),
		"forwarded_for": "192.0.2.1",
		"user_agent":    "observability-test",
	} {
		if got := entry[key]; got != want {
			t.Errorf("log %s = %#v, want %#v", key, got, want)
		}
	}

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}
	if got := metricValue(families, "kaomojiserv_http_requests_total", map[string]string{
		"method": http.MethodGet,
		"route":  "/api",
		"status": "201",
	}); got != 1 {
		t.Errorf("request counter = %v, want 1", got)
	}
}

func TestMetricsEndpointUsesPrometheusFormat(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := newHTTPMetrics(registry)
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/metrics", nil)

	observeHTTP(slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)), metrics, mux).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	mediaType, parameters, err := mime.ParseMediaType(response.Header().Get("Content-Type"))
	if err != nil {
		t.Fatalf("parse Content-Type: %v", err)
	}
	if mediaType != "text/plain" || parameters["version"] != "0.0.4" {
		t.Fatalf("Content-Type = %q, want Prometheus text format", response.Header().Get("Content-Type"))
	}
	if _, err = new(expfmt.TextParser).TextToMetricFamilies(bytes.NewReader(response.Body.Bytes())); err != nil {
		t.Fatalf("parse Prometheus metrics: %v", err)
	}
}

func metricValue(families []*dto.MetricFamily, name string, labels map[string]string) float64 {
	for _, family := range families {
		if family.GetName() != name {
			continue
		}
		for _, metric := range family.Metric {
			matched := true
			for key, value := range labels {
				found := false
				for _, pair := range metric.Label {
					if pair.GetName() == key && pair.GetValue() == value {
						found = true
						break
					}
				}
				matched = matched && found
			}
			if matched {
				return metric.GetCounter().GetValue()
			}
		}
	}
	return 0
}

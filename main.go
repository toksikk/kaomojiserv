package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type kaomoji struct {
	Kaomoji string
}

type kaomojis struct {
	Kaomojis []kaomoji
}

type templateData struct {
	Kaomoji          string
	RemainingSeconds int64
}

type apiResponse struct {
	Kaomoji string `json:"kaomoji"`
	Total   int    `json:"total"`
}

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func loadKaomojis(path string) kaomojis {
	k := kaomojis{}
	slog.Info("load kaomojis from path", "path", path)
	file, err := os.Open(path)
	if err != nil {
		slog.Error("error while opening file", "error", err)
		panic(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			slog.Error("error while closing kaomoji file", "error", err)
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		k.Kaomojis = append(k.Kaomojis, kaomoji{Kaomoji: scanner.Text()})
	}
	slog.Info("kaomojis loaded", "count", len(k.Kaomojis))
	if err := scanner.Err(); err != nil {
		slog.Error("error while scanning kaomoji file", "error", err)
		panic(err)
	}
	return k
}

func randNum(i int) int {
	return rng.Intn(i)
}

type rotationTracker struct {
	mu           sync.Mutex
	timestamp    int64
	randomNumber int
}

func newRotationTracker(allk kaomojis) *rotationTracker {
	return &rotationTracker{
		timestamp:    time.Now().Unix(),
		randomNumber: randNum(len(allk.Kaomojis)),
	}
}

func (t *rotationTracker) rotateIfStale(allk kaomojis, timeout int64) (string, int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now().Unix()
	if now-t.timestamp > timeout {
		t.randomNumber = randNum(len(allk.Kaomojis))
		t.timestamp = now
		slog.Info("rotating kaomoji", "new", allk.Kaomojis[t.randomNumber].Kaomoji)
	}
	remaining := timeout - (now - t.timestamp)
	if remaining < 0 {
		remaining = 0
	}
	return allk.Kaomojis[t.randomNumber].Kaomoji, remaining
}

func (t *rotationTracker) current(allk kaomojis) string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return allk.Kaomojis[t.randomNumber].Kaomoji
}

func newMux(allk kaomojis, timeout int64, tmpl, guideTmpl *template.Template) *http.ServeMux {
	tracker := newRotationTracker(allk)
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		current, remaining := tracker.rotateIfStale(allk, timeout)
		data := templateData{
			Kaomoji:          current,
			RemainingSeconds: remaining,
		}
		if err := tmpl.Execute(w, data); err != nil {
			slog.Error("error while executing template", "error", err)
		}
	})

	mux.HandleFunc("/easter-eggs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := guideTmpl.Execute(w, nil); err != nil {
			slog.Error("error while executing Easter egg guide template", "error", err)
		}
	})

	mux.HandleFunc("/raw", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if _, err := fmt.Fprintln(w, tracker.current(allk)); err != nil {
			slog.Error("error writing raw response", "error", err)
		}
	})

	mux.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := json.NewEncoder(w).Encode(apiResponse{Kaomoji: tracker.current(allk), Total: len(allk.Kaomojis)}); err != nil {
			slog.Error("error encoding api response", "error", err)
		}
	})

	mux.HandleFunc("/all", func(w http.ResponseWriter, r *http.Request) {
		type allResp struct {
			Kaomojis []string `json:"kaomojis"`
			Total    int      `json:"total"`
		}
		resp := allResp{Total: len(allk.Kaomojis)}
		for _, entry := range allk.Kaomojis {
			resp.Kaomojis = append(resp.Kaomojis, entry.Kaomoji)
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("error encoding all response", "error", err)
		}
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if _, err := fmt.Fprintf(w, "(*^_^*) all %d kaomojis accounted for\n", len(allk.Kaomojis)); err != nil {
			slog.Error("error writing health response", "error", err)
		}
	})

	mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		banner(w)
	})

	return mux
}

func main() {
	port := flag.String("port", "80", "http listening port")
	timeoutParameter := flag.String("timeout", "60", "time in seconds after last rotation until kaomoji gets rotated again")
	kaomojisPath := flag.String("kaomojis", "kaomojis.txt", "path to file with kaomojis")
	templatePath := flag.String("template", "kaomoji_template.html", "path to HTML template file")
	guideTemplatePath := flag.String("guide-template", "easter_eggs_template.html", "path to Easter egg guide template file")
	flag.Parse()

	timeout, err := strconv.ParseInt(*timeoutParameter, 10, 0)
	if err != nil {
		slog.Error("error while parsing timeout parameter", "error", err)
		panic(err)
	}

	slog.Info("parsing template file", "path", *templatePath)
	tmpl, err := template.ParseFiles(*templatePath)
	if err != nil {
		slog.Error("error while parsing template file", "error", err)
		panic(err)
	}
	slog.Info("parsing Easter egg guide template", "path", *guideTemplatePath)
	guideTmpl, err := template.ParseFiles(*guideTemplatePath)
	if err != nil {
		slog.Error("error while parsing Easter egg guide template", "error", err)
		panic(err)
	}

	allk := loadKaomojis(*kaomojisPath)

	registry := prometheus.NewRegistry()
	registry.MustRegister(collectors.NewGoCollector(), collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	metrics := newHTTPMetrics(registry)

	mux := newMux(allk, timeout, tmpl, guideTmpl)
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	slog.Info("webserver starting", "port", *port)
	err = http.ListenAndServe(":"+*port, observeHTTP(slog.Default(), metrics, mux))
	if err != nil {
		slog.Error("error while starting webserver", "error", err)
		panic(err)
	}
}

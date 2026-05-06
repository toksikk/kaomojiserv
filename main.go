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
	defer file.Close()

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

func main() {
	port := flag.String("port", "80", "http listening port")
	timeoutParameter := flag.String("timeout", "60", "time in seconds after last rotation until kaomoji gets rotated again")
	kaomojisPath := flag.String("kaomojis", "kaomojis.txt", "path to file with kaomojis")
	templatePath := flag.String("template", "kaomoji_template.html", "path to HTML template file")
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

	allk := loadKaomojis(*kaomojisPath)

	var mu sync.Mutex
	timestamp := time.Now().Unix()
	randomNumber := randNum(len(allk.Kaomojis))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		if time.Now().Unix()-timestamp > timeout {
			randomNumber = randNum(len(allk.Kaomojis))
			timestamp = time.Now().Unix()
			slog.Info("rotating kaomoji", "new", allk.Kaomojis[randomNumber].Kaomoji)
		}
		current := allk.Kaomojis[randomNumber].Kaomoji
		remaining := timeout - (time.Now().Unix() - timestamp)
		mu.Unlock()

		if remaining < 0 {
			remaining = 0
		}
		slog.Info("serving kaomoji", "ip", r.Header.Get("x-forwarded-for"))
		data := templateData{
			Kaomoji:          current,
			RemainingSeconds: remaining,
		}
		err = tmpl.Execute(w, data)
		if err != nil {
			slog.Error("error while executing template", "error", err)
		}
	})

	http.HandleFunc("/raw", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		k := allk.Kaomojis[randomNumber].Kaomoji
		mu.Unlock()
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintln(w, k)
	})

	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		k := allk.Kaomojis[randomNumber].Kaomoji
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(apiResponse{Kaomoji: k, Total: len(allk.Kaomojis)})
	})

	http.HandleFunc("/all", func(w http.ResponseWriter, r *http.Request) {
		type allResp struct {
			Kaomojis []string `json:"kaomojis"`
			Total    int      `json:"total"`
		}
		resp := allResp{Total: len(allk.Kaomojis)}
		for _, entry := range allk.Kaomojis {
			resp.Kaomojis = append(resp.Kaomojis, entry.Kaomoji)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintf(w, "(*^_^*) all %d kaomojis accounted for\n", len(allk.Kaomojis))
	})

	http.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		banner(w)
	})

	slog.Info("webserver starting", "port", *port)
	err = http.ListenAndServe(":"+*port, nil)
	if err != nil {
		slog.Error("error while starting webserver", "error", err)
		panic(err)
	}
}

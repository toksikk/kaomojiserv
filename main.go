package main

import (
	"bufio"
	"flag"
	"html/template"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

type kaomoji struct {
	Kaomoji string
}

type kaomojis struct {
	Kaomojis []kaomoji
}

func loadKaomojis(path string) kaomojis {
	kaomojis := kaomojis{}
	slog.Info("load kaomojis from path", "path", path)
	file, err := os.Open(path)
	if err != nil {
		slog.Error("error while opening file", "error", err)
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		kaomojis.Kaomojis = append(kaomojis.Kaomojis, kaomoji{Kaomoji: scanner.Text()})
	}
	slog.Info("kaomojis loaded", "kaomojis", kaomojis.Kaomojis)
	if err := scanner.Err(); err != nil {
		// log error and then panic
		slog.Error("error while scanning kaomoji file", "error", err)
		panic(err)
	}
	return kaomojis
}

func randNum(i int) int {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	return rand.Intn(i)
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

	timestamp := time.Now().Unix()
	randomNumber := randNum(len(allk.Kaomojis))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if time.Now().Unix()-timestamp > timeout {
			randomNumber = randNum(len(allk.Kaomojis))
			timestamp = time.Now().Unix()
			slog.Info("rotating kaomoji")
		}
		slog.Info("serving kaomoji", "ip", r.Header.Get("x-forwarded-for"))
		k := allk.Kaomojis[randomNumber]
		err = tmpl.Execute(w, k)
		if err != nil {
			slog.Info("error while executing template", "error", err)
		}
	})

	http.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		banner(w)
	})

	err = http.ListenAndServe(":"+*port, nil)
	if err != nil {
		slog.Error("error while starting webserver", "error", err)
		panic(err)
	}
	slog.Info("webserver started", "port", *port)
	slog.Info("press ctrl-c to exit.")
}

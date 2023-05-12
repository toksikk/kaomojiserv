package main

import (
	"bufio"
	"flag"
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

type kaomoji struct {
	Kaomoji string
}

type kaomojis struct {
	Kaomojis []kaomoji
}

func loadKaomojis(path string) kaomojis {
	kaomojis := kaomojis{}
	log.Println("load kaomojis from " + path + ".")
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		kaomojis.Kaomojis = append(kaomojis.Kaomojis, kaomoji{Kaomoji: scanner.Text()})
	}
	log.Println("kaomojis loaded.")
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return kaomojis
}

func randNum(i int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(i)
}

func main() {
	port := flag.String("port", "80", "http listening port")
	timeoutParameter := flag.String("timeout", "60", "time in seconds after last rotation until kaomoji gets rotated again")
	kaomojisPath := flag.String("kaomojis", "kaomojis.txt", "path to file with kaomojis")
	templatePath := flag.String("template", "kaomoji_template.html", "path to HTML template file")
	flag.Parse()

	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})

	timeout, err := strconv.ParseInt(*timeoutParameter, 10, 0)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("parsing template file from " + *templatePath + ".")
	tmpl, err := template.ParseFiles(*templatePath)
	if err != nil {
		log.Fatal(err)
	}

	allk := loadKaomojis(*kaomojisPath)

	timestamp := time.Now().Unix()
	randomNumber := randNum(len(allk.Kaomojis))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if time.Now().Unix()-timestamp > timeout {
			randomNumber = randNum(len(allk.Kaomojis))
			timestamp = time.Now().Unix()
			log.Println("rotating kaomoji.")
		}
		log.Println("serving kaomoji to " + r.Header.Get("x-forwarded-for") + ".")
		k := allk.Kaomojis[randomNumber]
		tmpl.Execute(w, k)
	})

	http.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		banner(w)
	})

	err = http.ListenAndServe(":"+*port, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("webserver listening on port", *port, ". press ctrl-c to exit.")
}

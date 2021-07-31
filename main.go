package main

import (
	"bufio"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

type Kamoji struct {
	Kamoji string
}

type Kamojis struct {
	Kamojis []Kamoji
}

func loadKamojis() Kamojis {
	kamojis := Kamojis{}

	file, err := os.Open("kamojis.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		kamojis.Kamojis = append(kamojis.Kamojis, Kamoji{Kamoji: scanner.Text()})
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return kamojis
}

func main() {
	tmpl, err := template.ParseFiles("kamoji_template.html")
	if err != nil {
		log.Fatal(err)
	}
	allk := loadKamojis()
	timestamp := time.Now().Unix()
	randomNumber := rand.Intn(len(allk.Kamojis))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if time.Now().Unix()-timestamp > 60 {
			randomNumber = rand.Intn(len(allk.Kamojis))
			timestamp = time.Now().Unix()
		}
		k := allk.Kamojis[randomNumber]
		tmpl.Execute(w, k)
	})
	http.ListenAndServe(":80", nil)
}

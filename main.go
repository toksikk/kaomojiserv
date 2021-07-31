package main

import (
	"bufio"
	"flag"
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

type Kamoji struct {
	Kamoji string
}

type Kamojis struct {
	Kamojis []Kamoji
}

func loadKamojis(path string) Kamojis {
	kamojis := Kamojis{}
	log.Println("load kamojis from " + path + ".")
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		kamojis.Kamojis = append(kamojis.Kamojis, Kamoji{Kamoji: scanner.Text()})
	}
	log.Println("kamojis loaded.")
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return kamojis
}

func main() {
	port := flag.String("port", "80", "http listening port")
	kamojisPath := flag.String("kamojis", "kamojis.txt", "path to file with kamojis")
	templatePath := flag.String("template", "kamoji_template.html", "path to HTML template file")
	flag.Parse()

	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})

	log.Println("parsing template file from " + *templatePath + ".")
	tmpl, err := template.ParseFiles(*templatePath)
	if err != nil {
		log.Fatal(err)
	}
	allk := loadKamojis(*kamojisPath)
	timestamp := time.Now().Unix()
	randomNumber := rand.Intn(len(allk.Kamojis))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if time.Now().Unix()-timestamp > 60 {
			randomNumber = rand.Intn(len(allk.Kamojis))
			timestamp = time.Now().Unix()
			log.Println("rotating kamoji.")
		}
		log.Println("served kamoji to " + r.RemoteAddr + ".")
		k := allk.Kamojis[randomNumber]
		tmpl.Execute(w, k)
	})
	http.ListenAndServe(":"+*port, nil)
	log.Println("starting webserver on port " + *port + ". press ctrl-c to exit.")
}

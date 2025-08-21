package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"url-shortener/shorten"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env")
	}

	dbConfig := &shorten.DataBaseConfig{
		Host:     os.Getenv("HOST"),
		Port:     os.Getenv("PORT"),
		User:     os.Getenv("POSTGRESUSER"),
		Password: os.Getenv("PASSWORD"),
		DB:       os.Getenv("DB"),
	}

	shortener := &shorten.URLShortener{
		Urls:     make(map[string]string),
		Stats:    make(map[string]map[string]int),
		DbConfig: dbConfig,
	}

	shorten.PopulateMap(shortener)

	http.HandleFunc("/shorten", shortener.HandleShorten)
	http.HandleFunc("/stats/", shortener.HandlerStats)
	http.HandleFunc("/", shortener.HandleRedirect)

	fileServer := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fileServer))
	http.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/create.html")
	})

	fmt.Println("URL Shortener is running on :8080")
	http.ListenAndServe(":8080", nil)
}

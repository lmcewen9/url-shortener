package main

import (
	"context"
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
		Host: os.Getenv("HOST"),
		Port: os.Getenv("PORT"),
		User: os.Getenv("USER"),
		Password: os.Getenv("PASSWORD"),
		DB: os.Getenv("DB"),
	}

	db, err := dbConfig.Connect()
	if err != nil {
		log.Fatal("Error connecting to Database...")
	}

	db.Close(context.Background())

	shortener := &shorten.URLShortener{
		Urls: make(map[string]string),
	}

	http.HandleFunc("/shorten", shortener.HandleShorten)
	http.HandleFunc("/", shortener.HandleRedirect)

	fmt.Println("URL Shortener is running on :8080")
	http.ListenAndServe(":8080", nil)
}
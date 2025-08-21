package shorten

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"net/http"
	"time"
)

type URLShortener struct {
	Urls     map[string]string
	Stats    map[string]map[string]int
	DbConfig *DataBaseConfig
}

func GenerateShortKey() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const keylength = 3

	seed := time.Now().UnixNano()
	source := rand.NewPCG(uint64(seed), rand.New(rand.NewPCG(uint64(seed), uint64(seed+1))).Uint64())
	r := rand.New(source)

	shortKey := make([]byte, keylength)
	for i := range shortKey {
		shortKey[i] = charset[r.IntN(len(charset))]
	}
	return string(shortKey)
}

func CheckShortKey(urls map[string]string) string {
	shortKey := GenerateShortKey()
	if _, exists := urls[shortKey]; exists {
		return CheckShortKey(urls)
	}
	return shortKey
}

func (us *URLShortener) HandleShorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid Request Method...", http.StatusMethodNotAllowed)
		return
	}

	ogUrl := r.FormValue("url")
	if ogUrl == "" {
		http.Error(w, "URL paramerter is missing...", http.StatusBadRequest)
		return
	}

	conn, err := us.DbConfig.Connect()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error conntecting to Database: %v", err), http.StatusInternalServerError)
		return
	}
	defer conn.Close(context.Background())

	shortKey := CheckShortKey(us.Urls)

	us.Urls[shortKey] = ogUrl
	err = CreateEntry(shortKey, ogUrl, conn)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error Creating Entry: %v", err), http.StatusInternalServerError)
		return
	}

	shortenedURL := fmt.Sprintf("http://localhost:8080/%s", shortKey)
	responseHTML := fmt.Sprintf("Shortened URL is %s", shortenedURL)
	fmt.Fprint(w, responseHTML)
}

func (us *URLShortener) HandleRedirect(w http.ResponseWriter, r *http.Request) {
	shortKey := r.URL.Path[len("/"):]
	if shortKey == "" || shortKey == "stats" {
		http.Redirect(w, r, "/create", http.StatusTemporaryRedirect)
		return
	}

	ogURL, exists := us.Urls[shortKey]
	if !exists {
		http.Error(w, "Shortened key not found...", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, ogURL, http.StatusMovedPermanently)

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Printf("Error splitting host and port: %v\n", err)
	}
	if _, exists := us.Stats[shortKey]; !exists {
		us.Stats[shortKey] = make(map[string]int)
		us.Stats[shortKey][host] = 1
	} else {
		us.Stats[shortKey][host]++
	}
}

func (us *URLShortener) HandlerStats(w http.ResponseWriter, r *http.Request) {
	shortKey := r.URL.Path[len("/stats/"):]
	if shortKey == "" {
		http.Redirect(w, r, "/create", http.StatusTemporaryRedirect)
		return
	}
	data, err := json.MarshalIndent(us.Stats, "", " ")
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}
	responseData := fmt.Sprintf("Stats on %s:\n%s", shortKey, string(data))
	fmt.Fprint(w, responseData)
}

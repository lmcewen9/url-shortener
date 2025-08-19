package shorten

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"time"
)

type URLShortener struct {
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

func CheckShortKey(t *[]Table) string {
	shortKey := GenerateShortKey()
	for _, i := range *t {
		if shortKey == i.Slug {
			CheckShortKey(t)
		}
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
		log.Fatal("Error conntecting to Database")
	}
	defer conn.Close(context.Background())

	entries := ReadEntry(conn)
	shortKey := CheckShortKey(entries)

	err = CreateEntry(shortKey, ogUrl, conn)
	if err != nil {
		log.Fatal("Error Creating Entry")
	}

	shortenedURL := fmt.Sprintf("http://localhost:8080/%s", shortKey)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	responseHTML := fmt.Sprintf(`
		<h2>URL Shortener</h2>
        <p>Original URL: %s</p>
        <p>Shortened URL: <a href="%s">%s</a></p>
        <form method="post" action="/shorten">
            <input type="text" name="url" placeholder="Enter a URL">
            <input type="submit" value="Shorten">
        </form>
	`, ogUrl, shortenedURL, shortenedURL)
	fmt.Fprint(w, responseHTML)
}

func (us *URLShortener) HandleRedirect(w http.ResponseWriter, r *http.Request) {
	shortKey := r.URL.Path[len("/"):]
	if shortKey == "" {
		http.Error(w, "Shortened key is missing...", http.StatusBadRequest)
		return
	}

	conn, err := us.DbConfig.Connect()
	if err != nil {
		log.Fatal("Error connecting to database", err)
	}
	defer conn.Close(context.Background())

	var ogURL string
	entries := ReadEntry(conn)
	for _, e := range *entries {
		if e.Slug == shortKey {
			ogURL = e.OgUrl
			break
		} else {
			http.Error(w, "Shortened key not found...", http.StatusNotFound)
			return
		}
	}

	http.Redirect(w, r, ogURL, http.StatusMovedPermanently)
}

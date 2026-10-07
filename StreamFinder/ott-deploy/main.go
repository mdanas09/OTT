package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go-movie-recommender/api"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/genres", api.Genres)
	mux.HandleFunc("/api/movie-types", api.MovieTypes)
	mux.HandleFunc("/api/countries", api.Countries)
	mux.HandleFunc("/api/recommend", api.Recommend)
	mux.HandleFunc("/api/user-country", api.UserCountry)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	mux.HandleFunc("/healthz", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte("ok"))
	})
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/" {
			http.NotFound(writer, request)
			return
		}
		http.ServeFile(writer, request, "index.html")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "10000"
	}

	addr := fmt.Sprintf("0.0.0.0:%s", port)
	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("server listening on %s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

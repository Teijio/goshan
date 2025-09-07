package main

import (
	"net/http"

	"github.com/Teijio/goshan/internal/handlers"
	"github.com/Teijio/goshan/internal/repository"
	"github.com/Teijio/goshan/internal/service"
)

const baseURL = "http://localhost:8080"

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	rep := repository.NewURLRepository()
	service := service.NewURLService(rep)
	h := handlers.NewHandler(service, baseURL)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.CreateShortenLink(w, r)
		case http.MethodGet:
			h.GetOriginalLink(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return http.ListenAndServe(":8080", mux)
}

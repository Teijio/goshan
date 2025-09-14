package main

import (
	"log"
	"net/http"

	"github.com/Teijio/goshan/internal/config"
	"github.com/Teijio/goshan/internal/handlers"
	"github.com/Teijio/goshan/internal/repository"
	"github.com/Teijio/goshan/internal/service"
	"github.com/go-chi/chi/v5"
)

// SERVER_ADDRESS
// BASE_URL

func main() {
	cfg := config.LoadConfig()
	
	log.Printf("Starting server on %s (base URL: %s)", cfg.ServerAddress, cfg.BaseURL)

	rep := repository.NewURLRepository()
	service := service.NewURLService(rep)
	h := handlers.NewHandler(service, cfg.BaseURL)

	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Post("/", h.CreateShortenLink)
		r.Get("/{id}", h.GetOriginalLink )
	})

	log.Fatal(http.ListenAndServe(cfg.ServerAddress, r))
}

package main

import (
	"net/http"

	"github.com/Teijio/goshan/internal/config"
	"github.com/Teijio/goshan/internal/handlers"
	"github.com/Teijio/goshan/internal/middleware"
	"github.com/Teijio/goshan/internal/repository"
	"github.com/Teijio/goshan/internal/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	cfg := config.LoadConfig()

	rep := repository.NewURLRepository()
	service := service.NewURLService(rep)
	h := handlers.NewHandler(service, cfg.BaseURL, logger)

	r := chi.NewRouter()

	r.Use(middleware.LoggingMiddleware(logger))
	r.Route("/", func(r chi.Router) {
		r.Post("/", h.CreateShortenLink)
		r.Get("/{id}", h.GetOriginalLink)
	})
	r.Route("/api", func(r chi.Router) {
		r.Post("/shorten", h.CreateShortenLinkV2)
	})
	if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
		logger.Fatal("Server failed to start", zap.Error(err))
	}
}

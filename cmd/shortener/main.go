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

var sugar *zap.SugaredLogger
gorka
func main() {
    logger, err := zap.NewDevelopment()
    if err != nil {
        panic(err)
    }
    defer logger.Sync()
	sugar = logger.Sugar()

	cfg := config.LoadConfig()


	rep := repository.NewURLRepository()
	service := service.NewURLService(rep)
	h := handlers.NewHandler(service, cfg.BaseURL)

	r := chi.NewRouter()
	r.Use(middleware.LoggingMiddleware(sugar))

	r.Route("/", func(r chi.Router) {
		r.Post("/", h.CreateShortenLink)
		r.Get("/{id}", h.GetOriginalLink )
	})

	sugar.Fatal(http.ListenAndServe(cfg.ServerAddress, r))
}

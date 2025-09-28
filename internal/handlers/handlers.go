package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Teijio/goshan/internal/config"
	"github.com/Teijio/goshan/internal/middleware"
	"github.com/Teijio/goshan/internal/models"
	"github.com/Teijio/goshan/internal/repository"
	"github.com/Teijio/goshan/internal/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func NewRouter(service *service.URLService, cfg *config.Config, logger *zap.Logger) chi.Router {
	h := NewHandler(service, cfg.BaseURL, logger)
	r := chi.NewRouter()
	r.Use(middleware.LoggingMiddleware(logger), middleware.ResponseCompressor, middleware.RequestDecompressor)
	r.Route("/", func(r chi.Router) {
		r.Post("/", h.CreateShortenLink)
		r.Get("/{id}", h.GetOriginalLink)
		r.Post("/api/shorten", h.CreateShortenLinkV2)
		r.Get("/ping", h.Ping)
	})
	return r
}

type Handler struct {
	urlService *service.URLService
	baseURL    string
	logger     *zap.Logger
}

func NewHandler(svc *service.URLService, baseURL string, logger *zap.Logger) *Handler {
	return &Handler{
		urlService: svc,
		baseURL:    baseURL,
		logger:     logger,
	}
}

func writeShortenResult(w http.ResponseWriter, h *Handler, shortURL models.ShortURL, status int) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(status)
	short := h.urlService.FormatShorlURL(shortURL.ID)
	if _, err := w.Write([]byte(short)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) CreateShortenLink(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("Error reading request body", zap.Error(err))
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	original := string(body)
	if _, err := url.ParseRequestURI(original); err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	shortURL, err := h.urlService.Shorten(original)
	var notUniqueErr *repository.NotUniqueURLError
	if errors.As(err, &notUniqueErr) {
		writeShortenResult(w, h, shortURL, http.StatusConflict)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeShortenResult(w, h, shortURL, http.StatusCreated)

}

func (h *Handler) GetOriginalLink(w http.ResponseWriter, r *http.Request) {
	id := strings.ToLower(chi.URLParam(r, "id"))
	shortURL, err := h.urlService.GetOriginalLink(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	http.Redirect(w, r, shortURL.OriginalURL, http.StatusTemporaryRedirect)
}

func (h *Handler) CreateShortenLinkV2(w http.ResponseWriter, r *http.Request) {
	var req models.ShortURL
	dec := json.NewDecoder(r.Body)

	if err := dec.Decode(&req); err != nil {
		h.logger.Error("Error decoding JSON", zap.Error(err))
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Warn("Invalid URL in request", zap.String("url", req.OriginalURL), zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	shortURL, err := h.urlService.Shorten(req.OriginalURL)
	var notUniqueErr *repository.NotUniqueURLError
	if errors.As(err, &notUniqueErr) {
		writeShortenResultAPI(w, h, shortURL, http.StatusConflict)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeShortenResultAPI(w, h, shortURL, http.StatusCreated)
}

func writeShortenResultAPI(w http.ResponseWriter, h *Handler, shortURL models.ShortURL, status int) {
	response := models.Response{Result: h.urlService.FormatShorlURL(shortURL.ID)}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Error encoding response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	if err := h.urlService.HealthCheck(); err != nil {
		h.logger.Error("Health check failed", zap.Error(err))
		http.Error(w, "Cant check repo status", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

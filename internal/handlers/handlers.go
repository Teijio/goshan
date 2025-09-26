package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Teijio/goshan/internal/models"
	"github.com/Teijio/goshan/internal/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

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

	short, err := h.urlService.Shorten(original)
	if err != nil {
		http.Error(w, "Error on shorten...2", http.StatusBadRequest)
		return
	}
	shortURL := fmt.Sprintf("%s/%s", h.baseURL, short)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, shortURL)
}

func (h *Handler) GetOriginalLink(w http.ResponseWriter, r *http.Request) {
	id := strings.ToLower(chi.URLParam(r, "id"))
	original, err := h.urlService.GetOriginalLink(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	http.Redirect(w, r, original, http.StatusTemporaryRedirect)
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

	short, err := h.urlService.Shorten(req.OriginalURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	shortURL := fmt.Sprintf("%s/%s", h.baseURL, short)

	response := models.Response{Result: shortURL}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Error encoding response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}

}

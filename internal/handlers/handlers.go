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

	short := h.urlService.Shorten(original)
	shortURL := fmt.Sprintf("%s/%s", h.baseURL, short)

	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, shortURL)
}

func (h *Handler) GetOriginalLink(w http.ResponseWriter, r *http.Request) {
	id := strings.ToLower(chi.URLParam(r, "id"))
	original, ok := h.urlService.GetOriginalLink(id)
	if !ok {
		h.logger.Warn("Original link not found", zap.String("id", id))
		http.Error(w, "Original link by this ID not found", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, original, http.StatusTemporaryRedirect)
}

func (h *Handler) CreateShortenLinkV2(w http.ResponseWriter, r *http.Request) {
	var req models.Request
	dec := json.NewDecoder(r.Body)

	if err := dec.Decode(&req); err != nil {
		h.logger.Error("Error decoding JSON", zap.Error(err))
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Warn("Invalid URL in request", zap.String("url", req.URL), zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	short := h.urlService.Shorten(req.URL)
	shortURL := fmt.Sprintf("%s/%s", h.baseURL, short)

	response := models.Response{Result: shortURL}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Error encoding response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}

}

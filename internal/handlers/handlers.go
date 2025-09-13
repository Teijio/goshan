package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Teijio/goshan/internal/service"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	urlService *service.URLService
	baseURL    string
}

func NewHandler(svc *service.URLService, baseURL string) *Handler {
	return &Handler{urlService: svc, baseURL: baseURL}
}

func (h *Handler) CreateShortenLink(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
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
		http.Error(w, "Original link by this ID not found", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, original, http.StatusTemporaryRedirect)
}

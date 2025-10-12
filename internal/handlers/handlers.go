package handlers

import (
	"encoding/hex"
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
	"github.com/Teijio/goshan/internal/responses"
	"github.com/Teijio/goshan/internal/service"
	"github.com/Teijio/goshan/internal/service/crypto"
	"github.com/Teijio/goshan/internal/service/random"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

const UserIDCookieName = "shortener-user-id"

func NewRouter(service *service.URLService, cfg *config.Config, logger *zap.Logger, generator random.UserIDGenerator) chi.Router {
	h := NewHandler(service, cfg.BaseURL, logger, generator, cfg)
	r := chi.NewRouter()
	r.Use(middleware.LoggingMiddleware(logger), middleware.ResponseCompressor, middleware.RequestDecompressor)
	r.Route("/", func(r chi.Router) {
		r.Get("/{id}", h.GetOriginalLink)
		r.Get("/ping", h.Ping)
		r.Get("/api/user/urls", h.UserURLs)
		r.Post("/", h.CreateShortenLink)
		r.Post("/api/shorten", h.CreateShortenLinkV2)
		r.Post("/api/shorten/batch", h.CreateShortenLinks)
	})
	return r
}

type Handler struct {
	urlService      *service.URLService
	baseURL         string
	logger          *zap.Logger
	crypto          crypto.Cryptographer
	userIDGenerator random.UserIDGenerator
}

func NewHandler(svc *service.URLService, baseURL string, logger *zap.Logger, userIDGenerator random.UserIDGenerator, cfg *config.Config) *Handler {
	cryptographer := crypto.GCMAESCryptographer{Key: cfg.EncryptionKey}
	return &Handler{
		urlService:      svc,
		baseURL:         baseURL,
		logger:          logger,
		crypto:          &cryptographer,
		userIDGenerator: userIDGenerator,
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
	userID := h.getUserID(r)
	shortURL, err := h.urlService.Shorten(original, userID)
	var notUniqueErr *repository.NotUniqueURLError
	if errors.As(err, &notUniqueErr) {
		writeShortenResult(w, h, shortURL, http.StatusConflict)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = h.addEncryptedUserIDToCookie(&w, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	userID := h.getUserID(r)
	shortURL, err := h.urlService.Shorten(req.OriginalURL, userID)
	var notUniqueErr *repository.NotUniqueURLError
	if errors.As(err, &notUniqueErr) {
		writeShortenResultAPI(w, h, shortURL, http.StatusConflict)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err = h.addEncryptedUserIDToCookie(&w, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	writeShortenResultAPI(w, h, shortURL, http.StatusCreated)
}

func (h *Handler) CreateShortenLinks(w http.ResponseWriter, r *http.Request) {
	type request struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}
	var input []request

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "cannot decode json", http.StatusBadRequest)
		return
	}

	batch := make([]models.ShortURL, len(input))
	for i, shortURLInput := range input {
		if shortURLInput.OriginalURL == "" {
			http.Error(w, "url required", http.StatusBadRequest)
			return
		}
		batch[i] = models.ShortURL{
			OriginalURL:   shortURLInput.OriginalURL,
			CorrelationID: shortURLInput.CorrelationID,
			CreatedByID:   "hipa",
		}
	}
	userID := h.getUserID(r)

	shortURLBatches, err := h.urlService.ShortenBatch(batch, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err = h.addEncryptedUserIDToCookie(&w, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	res := make([]responses.ShorteningBatchResult, len(shortURLBatches))
	for i, shortURLBatch := range shortURLBatches {
		res[i] = responses.ShorteningBatchResult{
			CorrelationID: shortURLBatch.CorrelationID,
			ShortURL:      h.urlService.FormatShortURL(shortURLBatch.ID),
		}
	}

	// w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusCreated)
	// if err := json.NewEncoder(w).Encode(res); err != nil {
	// 	h.logger.Error("Error encoding response", zap.Error(err))
	// 	http.Error(w, "Internal server error", http.StatusInternalServerError)
	// 	return
	// }
	out, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write(out); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func writeShortenResultAPI(w http.ResponseWriter, h *Handler, shortURL models.ShortURL, status int) {
	response := models.Response{Result: h.urlService.FormatShorlURL(shortURL.ID)}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

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
	userID := h.getUserID(r)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(userID))
}

func (h *Handler) getUserID(r *http.Request) string {
	encodedCookie, err := r.Cookie(UserIDCookieName)

	if err != nil {
		return h.userIDGenerator.GenerateUserID()
	}

	decodedCookie, err := hex.DecodeString(encodedCookie.Value)
	if err != nil {
		return h.userIDGenerator.GenerateUserID()
	}

	decryptedUserID, err := h.crypto.Decrypt(decodedCookie)
	if err != nil {
		return h.userIDGenerator.GenerateUserID()
	}

	return string(decryptedUserID)
}

func (h *Handler) addEncryptedUserIDToCookie(w *http.ResponseWriter, userID string) error {
	encryptedUserID, err := h.crypto.Encrypt([]byte(userID))
	if err != nil {
		return err
	}

	encodedCookieValue := hex.EncodeToString(encryptedUserID)

	http.SetCookie(
		*w,
		&http.Cookie{
			Name:  UserIDCookieName,
			Value: encodedCookieValue,
		},
	)
	return nil
}

func (h *Handler) UserURLs(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)

	URLs, err := h.urlService.GetUrlsCreatedBy(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(URLs) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	formattedURLs := make([]responses.UsersShortURL, 0)
	for _, URL := range URLs {
		formattedURLs = append(
			formattedURLs,
			responses.UsersShortURL{ShortURL: h.urlService.FormatShortURL(URL.ID), OriginalURL: URL.OriginalURL},
		)
	}

	out, err := json.Marshal(formattedURLs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(out); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

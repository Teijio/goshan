package handlers

import (
	// "bytes"
	// "encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	// "github.com/Teijio/goshan/internal/models"
	"github.com/Teijio/goshan/internal/repository"
	"github.com/Teijio/goshan/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

func NewTestLogger(t *testing.T) *zap.Logger {
	return zaptest.NewLogger(t, zaptest.Level(zapcore.DebugLevel))
}

func NewNullLogger() *zap.Logger {
	return zap.NewNop()
}

func setupHandler() *Handler {
	repo := repository.NewURLRepository()
	svc := service.NewURLService(repo)
	return NewHandler(svc, "http://localhost:8080", NewNullLogger())
}

type MockReader struct{}

func (m *MockReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func TestCreateShortenLink(t *testing.T) {
	tests := []struct {
		name        string
		requestBody io.Reader
		checkPrefix bool
		statusCode  int
		expected    string
	}{
		{
			name:        "Empty body",
			requestBody: strings.NewReader(""),
			statusCode:  http.StatusBadRequest,
			expected:    "Invalid URL\n",
		},
		{
			name:        "Invalid URL",
			requestBody: strings.NewReader("practicum.yandex.ru"),
			statusCode:  http.StatusBadRequest,
			expected:    "Invalid URL\n",
		},
		{
			name:        "Success created",
			requestBody: strings.NewReader("https://practicum.yandex.ru/"),
			checkPrefix: true,
			statusCode:  http.StatusCreated,
			expected:    "http://localhost:8080/",
		},
		{
			name:        "Error reading request body",
			requestBody: &MockReader{},
			statusCode:  http.StatusInternalServerError,
			expected:    "Error reading request body\n",
		},
	}

	url := "/"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := setupHandler()

			req := httptest.NewRequest(http.MethodPost, url, tt.requestBody)
			w := httptest.NewRecorder()

			h.CreateShortenLink(w, req)

			result := w.Result()
			defer result.Body.Close()

			bodyBytes, err := io.ReadAll(result.Body)
			require.NoError(t, err)

			body := string(bodyBytes)

			assert.Equal(t, tt.statusCode, result.StatusCode, "status code mismatch in test %q", tt.name)

			if tt.checkPrefix {
				assert.True(t, strings.HasPrefix(body, tt.expected),
					"response must start with baseURL in test %q, got %s", tt.name, body)
			} else {
				assert.Equal(t, tt.expected, body, "response body mismatch in test %q", tt.name)
			}
		})
	}
}

// func TestGetOriginalLink(t *testing.T) {
// 	tests := []struct {
// 		name           string
// 		path           string
// 		preCreateURL   string
// 		statusCode     int
// 		expectedBody   string
// 		expectedHeader string
// 	}{
// 		{
// 			name:       "Empty ID",
// 			path:       "/",
// 			statusCode: http.StatusMethodNotAllowed,
// 		},
// 		{
// 			name:         "Non-existent ID",
// 			path:         "/nonexistent",
// 			statusCode:   http.StatusBadRequest,
// 			expectedBody: "Original link by this ID not found\n",
// 		},
// 		{
// 			name:           "Success redirect",
// 			path:           "/abc123",
// 			preCreateURL:   "https://practicum.yandex.ru/",
// 			statusCode:     http.StatusTemporaryRedirect,
// 			expectedHeader: "https://practicum.yandex.ru/",
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			h := setupHandler()
// 			if tt.preCreateURL != "" {
// 				short := h.urlService.Shorten(tt.preCreateURL)
// 				tt.path = "/" + short
// 			}

// 			r := chi.NewRouter()
// 			r.Post("/", h.CreateShortenLink)
// 			r.Get("/{id}", h.GetOriginalLink)

// 			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
// 			w := httptest.NewRecorder()

// 			r.ServeHTTP(w, req)

// 			result := w.Result()
// 			defer result.Body.Close()
// 			bodyBytes, err := io.ReadAll(result.Body)
// 			require.NoError(t, err)
// 			body := string(bodyBytes)

// 			assert.Equal(t, tt.statusCode, result.StatusCode, "status code mismatch in test %q", tt.name)
// 			if tt.expectedBody != "" {
// 				assert.Equal(t, tt.expectedBody, body, "response body mismatch in test %q", tt.name)
// 			}
// 			if tt.expectedHeader != "" {
// 				location := result.Header.Get("Location")
// 				assert.Equal(t, tt.expectedHeader, location, "Location header mismatch in test %q", tt.name)
// 			}

// 		})

// 	}
// }

func TestGetOriginalLinkIntegration(t *testing.T) {
	h := setupHandler()

	r := chi.NewRouter()
	r.Post("/", h.CreateShortenLink)
	r.Get("/{id}", h.GetOriginalLink)

	originalURL := "https://practicum.yandex.ru/"
	createReq := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(originalURL))
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)

	createResult := createW.Result()
	defer createResult.Body.Close()

	assert.Equal(t, http.StatusCreated, createResult.StatusCode)

	bodyBytes, err := io.ReadAll(createResult.Body)
	require.NoError(t, err)
	shortURL := string(bodyBytes)
	shortID := strings.TrimPrefix(shortURL, "http://localhost:8080/")

	getReq := httptest.NewRequest(http.MethodGet, "/"+shortID, nil)
	getW := httptest.NewRecorder()
	r.ServeHTTP(getW, getReq)

	getResult := getW.Result()
	defer getResult.Body.Close()

	assert.Equal(t, http.StatusTemporaryRedirect, getResult.StatusCode)
	assert.Equal(t, originalURL, getResult.Header.Get("Location"))
}

// func TestCreateShortenLinkV2(t *testing.T) {
// 	tests := []struct {
// 		name         string
// 		requestBody  any
// 		expectedBody string
// 		expectedStatus int
// 	}{
// 		{
// 			name:           "Created success",
// 			requestBody:    models.Request{URL: "https://example.com"},
// 			expectedBody:   `{"result":"http://localhost:8080/abc123"}`,
// 			expectedStatus: http.StatusCreated,
// 		},
// 		{
// 			name:           "Invalid URL in body",
// 			requestBody:    models.Request{URL: "example.com"},
// 			expectedBody:   "invalid URL: parse \"example.com\": invalid URI for request\n",
// 			expectedStatus: http.StatusBadRequest,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			body, err := json.Marshal(tt.requestBody)
// 			if err != nil {
// 				t.Fatalf("failed to marshal body: %v", err)
// 			}

// 			req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(body))
// 			w := httptest.NewRecorder()

// 			h := setupHandler()
// 			h.CreateShortenLinkV2(w, req)

// 			resp := w.Result()
// 			defer resp.Body.Close()

// 			if resp.StatusCode != tt.expectedStatus {
// 				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
// 			}

// 		})
// 	}
// }

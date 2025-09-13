package service

import (
	"testing"

	"github.com/Teijio/goshan/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewURLService(t *testing.T) {
	repo := repository.NewURLRepository()
	service := NewURLService(repo)

	assert.NotNil(t, service)
	assert.NotNil(t, service.repo)
	assert.Equal(t, repo, service.repo)
}

func TestURLServiceShorten(t *testing.T) {
	tests := []struct {
		name     string
		original string
	}{
		{
			name:     "Regular URL",
			original: "https://example.com",
		},
		{
			name:     "URL with path",
			original: "https://example.com/path/to/resource",
		},
		{
			name:     "URL with query params",
			original: "https://example.com?param=value&another=param",
		},
		{
			name:     "Empty string",
			original: "",
		},
		{
			name:     "Long text",
			original: "very long text that should be shortened to 6 characters hash",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repository.NewURLRepository()
			service := NewURLService(repo)

			short := service.Shorten(tt.original)
			assert.Len(t, short, 6, "Short URL should be 6 characters")
			assert.Regexp(t, `^[a-f0-9]{6}$`, short, "Short URL should contain only hex characters")
			storedOriginal, found := repo.Get(short)
			assert.True(t, found, "URL should be stored in repository")
			assert.Equal(t, tt.original, storedOriginal, "Stored URL should match original")
		})
	}
}

func TestURLServiceShortenConsistency(t *testing.T) {
	repo := repository.NewURLRepository()
	service := NewURLService(repo)

	original := "https://consistent-url.com"

	short1 := service.Shorten(original)
	short2 := service.Shorten(original)
	short3 := service.Shorten(original)

	assert.Equal(t, short1, short2, "Same URL should produce same short code")
	assert.Equal(t, short2, short3, "Same URL should produce same short code")
	assert.Equal(t, short1, short3, "Same URL should produce same short code")

	_, found := repo.Get(short1)
	assert.True(t, found, "URL should be stored")
	differentOriginal := "https://different-url.com"
	differentShort := service.Shorten(differentOriginal)
	assert.NotEqual(t, short1, differentShort, "Different URLs should produce different short codes")
}

func TestURLServiceGetOriginalLink(t *testing.T) {
	tests := []struct {
		name          string
		setupShort    string
		setupOriginal string
		inputShort    string
		expectedURL   string
		expectedFound bool
	}{
		{
			name:          "Get existing URL",
			setupShort:    "abc123",
			setupOriginal: "https://example.com",
			inputShort:    "abc123",
			expectedURL:   "https://example.com",
			expectedFound: true,
		},
		{
			name:          "Get non-existing URL",
			setupShort:    "abc123",
			setupOriginal: "https://example.com",
			inputShort:    "nonexistent",
			expectedURL:   "",
			expectedFound: false,
		},
		{
			name:          "Get empty short",
			setupShort:    "abc123",
			setupOriginal: "https://example.com",
			inputShort:    "",
			expectedURL:   "",
			expectedFound: false,
		},
		{
			name:          "Get from empty service",
			setupShort:    "",
			setupOriginal: "",
			inputShort:    "any",
			expectedURL:   "",
			expectedFound: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repository.NewURLRepository()
			service := NewURLService(repo)

			if tt.setupShort != "" && tt.setupOriginal != "" {
				repo.Save(tt.setupShort, tt.setupOriginal)
			}
			original, found := service.GetOriginalLink(tt.inputShort)
			assert.Equal(t, tt.expectedFound, found)
			assert.Equal(t, tt.expectedURL, original)
		})
	}
}

func TestURLServiceIntegration(t *testing.T) {
	repo := repository.NewURLRepository()
	service := NewURLService(repo)
	testURLs := []string{
		"https://example.com/page1",
		"https://example.com/page2",
		"https://example.com/page3",
		"",
	}

	shortCodes := make([]string, len(testURLs))
	for t, url := range testURLs {
		shortCodes[t] = service.Shorten(url)
	}

	for i := 0; i < len(shortCodes)-1; i++ {
		for j := i + 1; j < len(shortCodes); j++ {
			if testURLs[i] != testURLs[j] {
				assert.NotEqual(t, shortCodes[i], shortCodes[j],
					"Different URLs should have different short codes")
			}
		}
	}

	for i, short := range shortCodes {
		original, found := service.GetOriginalLink(short)
		assert.True(t, found, "Should find URL for short: %s", short)
		assert.Equal(t, testURLs[i], original, "URL mismatch for short: %s", short)
	}
	_, found := service.GetOriginalLink("nonexistent123")
	assert.False(t, found, "Should not find non-existent short code")
}

func TestURLServiceShortenSpecialCharacters(t *testing.T) {
	repo := repository.NewURLRepository()
	service := NewURLService(repo)

	specialURLs := []string{
		"https://example.com/тест",
		"https://example.com/测试",
		"https://example.com/😊",
		"https://example.com/space in url",
	}

	for _, url := range specialURLs {
		t.Run("Speacial chars"+url, func(t *testing.T) {
			short := service.Shorten(url)
			require.Len(t, short, 6)
			assert.Regexp(t, `^[a-f0-9]{6}$`, short)

			original, found := service.GetOriginalLink(short)
			assert.True(t, found)
			assert.Equal(t, url, original)
		})
	}
}

func TestURLServiceEdgeCases(t *testing.T) {
	repo := repository.NewURLRepository()
	service := NewURLService(repo)

	longURL := "https://example.com/" + string(make([]byte, 1000))
	short := service.Shorten(longURL)

	assert.Len(t, short, 6)

	original, found := service.GetOriginalLink(short)
	assert.True(t, found)
	assert.Equal(t, longURL, original)
}

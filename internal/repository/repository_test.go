package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewURLRepository(t *testing.T) {
	repo := NewURLRepository()

	assert.NotNil(t, repo)
	assert.NotNil(t, repo.store)
	assert.Empty(t, repo.store)
}

func TestURLRepositorySave(t *testing.T) {
	tests := []struct {
		name      string
		short     string
		original  string
		expectLen int
	}{
		{
			name:      "Save first URL",
			short:     "abc123",
			original:  "https://example.com",
			expectLen: 1,
		},
		{
			name:      "Save second URL",
			short:     "def456",
			original:  "https://google.com",
			expectLen: 2,
		},
		{
			name:      "Overwrite existing URL",
			short:     "abc123",
			original:  "https://new-url.com",
			expectLen: 2,
		},
	}

	repo := NewURLRepository()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo.Save(tt.short, tt.original)

			// Проверяем, что данные сохранились
			assert.Equal(t, tt.expectLen, len(repo.store))
			assert.Equal(t, tt.original, repo.store[tt.short])
		})
	}
}

func TestURLRepositoryGet(t *testing.T) {
	tests := []struct {
		name          string
		short         string
		expectedURL   string
		expectedFound bool
		setupData     map[string]string
	}{
		{
			name:          "Get existing URL",
			short:         "abc123",
			expectedURL:   "https://example.com",
			expectedFound: true,
			setupData: map[string]string{
				"abc123": "https://example.com",
				"def456": "https://google.com",
			},
		},
		{
			name:          "Get non-existing URL",
			short:         "nonexistent",
			expectedURL:   "",
			expectedFound: false,
			setupData: map[string]string{
				"abc123": "https://example.com",
			},
		},
		{
			name:          "Get from empty repository",
			short:         "any",
			expectedURL:   "",
			expectedFound: false,
			setupData:     map[string]string{},
		},
		{
			name:          "Get with empty short key",
			short:         "",
			expectedURL:   "",
			expectedFound: false,
			setupData: map[string]string{
				"abc123": "https://example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewURLRepository()

			for short, original := range tt.setupData {
				repo.Save(short, original)
			}

			original, found := repo.Get(tt.short)

			assert.Equal(t, tt.expectedFound, found)
			assert.Equal(t, tt.expectedURL, original)
		})
	}
}

func TestURLRepositoryIntegration(t *testing.T) {
	repo := NewURLRepository()

	testData := map[string]string{
		"short1": "https://example.com/page1",
		"short2": "https://example.com/page2",
		"short3": "https://example.com/page3",
	}

	for short, original := range testData {
		repo.Save(short, original)
	}

	assert.Equal(t, len(testData), len(repo.store))

	for short, expectedOriginal := range testData {
		original, found := repo.Get(short)
		assert.True(t, found, "Should find URL for short: %s", short)
		assert.Equal(t, expectedOriginal, original, "URL mismatch for short: %s", short)
	}

	_, found := repo.Get("nonexistent")
	assert.False(t, found, "Should not find non-existent URL")
}

func TestURLRepository_ConcurrentAccess(t *testing.T) {
	repo := NewURLRepository()

	// Простая проверка, что методы не паникуют при конкурентном доступе
	// (в реальном приложении нужно было бы добавить мьютексы)

	// Сохраняем несколько значений
	repo.Save("key1", "value1")
	repo.Save("key2", "value2")

	// Читаем одновременно (это базовая проверка на отсутствие паник)
	go func() {
		repo.Get("key1")
		repo.Get("key2")
	}()

	// Основная горутина тоже читает
	val1, found1 := repo.Get("key1")
	val2, found2 := repo.Get("key2")

	assert.True(t, found1)
	assert.True(t, found2)
	assert.Equal(t, "value1", val1)
	assert.Equal(t, "value2", val2)
}

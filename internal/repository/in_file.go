package repository

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/Teijio/goshan/internal/models"
)

type FileRepository struct {
	file    *os.File
	encoder *json.Encoder
	decoder *json.Decoder
	mu      sync.RWMutex
}

func NewFileRepository(filePath string) (*FileRepository, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	return &FileRepository{
		file:    file,
		encoder: json.NewEncoder(file),
		decoder: json.NewDecoder(file),
		mu:      sync.RWMutex{},
	}, nil
}

func (fr *FileRepository) Save(shortURL models.ShortURL) error {
	_, err := fr.Get(shortURL.ID)
	if err != nil {
		return NewNotUniqueURLError(shortURL, nil)
	}

	fr.mu.Lock()
	defer fr.mu.Unlock()
	if err := fr.encoder.Encode(shortURL); err != nil {
		return err
	}

	return fr.file.Sync()
}

func (fr *FileRepository) Get(short string) (models.ShortURL, error) {
	fr.mu.RLock()
	defer fr.mu.RUnlock()

	_, err := fr.file.Seek(0, 0)
	if err != nil {
		return models.ShortURL{}, err
	}
	var shortURL models.ShortURL
	for {
		err := fr.decoder.Decode(&shortURL)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return models.ShortURL{}, err
		}
		if short == shortURL.ID {
			return shortURL, nil
		}
	}
	return models.ShortURL{}, ErrShortURLNotFound
}

func (fr *FileRepository) CloseFile() error {
	err := fr.file.Close()
	if err != nil {
		return err
	}
	return nil
}

func (fr *FileRepository) Check() error {
	_, err := fr.file.Stat()
	return err
}

func (fr *FileRepository) SaveBatch(batch []models.ShortURL) error {
	for _, shortURL := range batch {
		_, err := fr.Get(shortURL.ID)
		if err != nil {
			return NewNotUniqueURLError(shortURL, err)
		}
	}
	fr.mu.Lock()
	defer fr.mu.Unlock()
	for _, shortURL := range batch {
		if err := fr.encoder.Encode(shortURL); err != nil {
			return err
		}
	}
	return nil
}

func (fr *FileRepository) GetUsersUrls(id string) ([]models.ShortURL, error) {
	fr.mu.Lock()
	defer fr.mu.Unlock()

	if _, err := fr.file.Seek(0, 0); err != nil {
		return nil, err
	}

	var entry models.ShortURL
	var URLs []models.ShortURL

	for {
		err := fr.decoder.Decode(&entry)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return URLs, err
		}
		if entry.CreatedByID == id {
			URLs = append(URLs, entry)
		}
	}
	return URLs, nil
}

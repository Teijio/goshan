package repository

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/Teijio/goshan/internal/models"
)

// {"uuid":"1","short_url":"4rSPg8ap","original_url":"http://yandex.ru"}

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

func (fr *FileRepository) Save(short, original string) error {
	fr.mu.Lock()
	defer fr.mu.Unlock()
	// existing, err := fr.GetHelper(short)
	// if err != nil {
	// 	if err == ErrShortURLNotFound {
	// 		fmt.Println("Запись не найдена, сохраняем новую")
	// 	} else {
	// 		fmt.Println("Реальная ошибка:", err)
	// 		return err
	// 	}
	// }
	// if existing != "" {
	// 	fmt.Println("Запись уже существует, пропускаем")
	// 	return nil
	// }
	shortURL := models.ShortURL{
		OriginalURL: original,
		ID:          short,
	}
	if err := fr.encoder.Encode(shortURL); err != nil {
		return err
	}

	return fr.file.Sync()
}

func (fr *FileRepository) Get(short string) (string, error) {
	fr.mu.RLock()
	defer fr.mu.RUnlock()

	_, err := fr.file.Seek(0, 0)
	if err != nil {
		return "", err
	}
	var shortURL models.ShortURL
	for {
		err := fr.decoder.Decode(&shortURL)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return "", err
		}
		if short == shortURL.ID {
			return shortURL.OriginalURL, nil
		}
	}
	return "", ErrShortURLNotFound
}

func (fr *FileRepository) GetHelper(short string) (string, error) {
	currentPos, err := fr.file.Seek(0, 1)
	if err != nil {
		return "", err
	}
	defer fr.file.Seek(currentPos, 0)

	if _, err := fr.file.Seek(0, 0); err != nil {
		return "", err
	}
	for {
		var shortURL models.ShortURL
		err := fr.decoder.Decode(&shortURL)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return "", err
		}

		if shortURL.ID == short {
			return shortURL.OriginalURL, nil
		}
	}

	return "", ErrShortURLNotFound

}

func (fr *FileRepository) CloseFile() error {
	err := fr.file.Close()
	if err != nil {
		return err
	}
	return nil
}

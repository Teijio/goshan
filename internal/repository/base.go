package repository

import (
	"errors"
	"fmt"

	"github.com/Teijio/goshan/internal/config"
)

type Repository interface {
	Save(short, original string) error
	Get(short string) (string, error)
}

func GetRepository(cfg *config.Config) (Repository, error) {
	if filepath := cfg.FilePath; filepath != "" {
		repo, err := NewFileRepository(filepath)
		if err != nil {
			return nil, fmt.Errorf("failed to create file repository: %w", err)
		}
		return repo, nil
	}
	return NewURLRepository(), nil
}


var ErrShortURLNotFound = errors.New("short URL not found")

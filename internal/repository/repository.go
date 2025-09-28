package repository

import (
	"errors"

	"github.com/Teijio/goshan/internal/config"
)

type Repository interface {
	Save(short, original string) error
	Get(short string) (string, error)
	Check() error
}

func GetRepository(cfg *config.Config) Repository {
	if filepath := cfg.FilePath; filepath != "" {
		repo, err := NewFileRepository(filepath)
		if err != nil {
			panic(err)
		}
		return repo
	}
	if dbDsn := cfg.DatabaseDSN; dbDsn != "" {
		repo, err := NewPostgresRepository(dbDsn)
		if err != nil {
			panic(err)
		}
		return repo
	}

	return NewInMemoreRepository()
}

var ErrShortURLNotFound = errors.New("short URL not found")

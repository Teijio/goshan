package repository

import (
	"errors"

	"github.com/Teijio/goshan/internal/config"
	"github.com/Teijio/goshan/internal/models"
)

type Repository interface {
	Save(shortURL models.ShortURL) error
	Get(short string) (models.ShortURL, error)
	Check() error
}

func GetRepository(cfg *config.Config) Repository {
	if dbDsn := cfg.DatabaseDSN; dbDsn != "" {
		repo, err := NewPostgresRepository(dbDsn)
		if err != nil {
			panic(err)
		}
		return repo
	}
	if filepath := cfg.FilePath; filepath != "" {
		repo, err := NewFileRepository(filepath)
		if err != nil {
			panic(err)
		}
		return repo
	}

	return NewInMemoreRepository()
}

var ErrShortURLNotFound = errors.New("short URL not found")


type NotUniqueURLError struct {
	Err      error
	ShortURL models.ShortURL
}

func (err *NotUniqueURLError) Error() string {
	return "url or id are already exist"
}

func (err *NotUniqueURLError) Unwrap() error {
	return err.Err
}

func NewNotUniqueURLError(shortURL models.ShortURL, err error) error {
	return &NotUniqueURLError{
		Err:      err,
		ShortURL: shortURL,
	}
}

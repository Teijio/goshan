package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Teijio/goshan/internal/config"
	"github.com/Teijio/goshan/internal/models"
)

type Repository interface {
	Save(shortURL models.ShortURL) error
	DeleteUrls(ctx context.Context, urls []models.ShortURL) error
	Get(short string) (models.ShortURL, error)
	Check() error
	SaveBatch(batch []models.ShortURL) error
	GetUsersUrls(id string) ([]models.ShortURL, error)
}

func GetRepository(cfg *config.Config) Repository {
	if dbDsn := cfg.DatabaseDSN; dbDsn != "" {
		repo, err := NewPostgresRepository(dbDsn)
		fmt.Println(0)
		if err != nil {
			panic(err)
		}
		return repo
	}
	// if filepath := cfg.FilePath; filepath != "" {
	// 	repo, err := NewFileRepository(filepath)
	// 	fmt.Println(1)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	return repo
	// }
	// fmt.Println(2)
	return NewInMemoreRepository()
}

var ErrShortURLNotFound = errors.New("short URL not found")
var ErrURLNotFoundByID = errors.New("can't find full url by id")

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

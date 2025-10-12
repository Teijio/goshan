package service

import (
	"crypto/sha1"
	"errors"
	"fmt"

	"github.com/Teijio/goshan/internal/config"
	"github.com/Teijio/goshan/internal/models"
	"github.com/Teijio/goshan/internal/repository"
)

type URLService struct {
	repo   repository.Repository
	config *config.Config
}

type shorteningError struct {
	Err      error
	ShortURL models.ShortURL
}

func (err *shorteningError) Error() string {
	return fmt.Sprintf("error while shortening: %v", err.Err)
}

func (err *shorteningError) Unwrap() error {
	return err.Err
}

func (s *URLService) FormatShorlURL(urlID string) string {
	return fmt.Sprintf("%s/%s", s.config.BaseURL, urlID)
}

func NewShorteningError(shortURL models.ShortURL, err error) error {
	return &shorteningError{
		Err:      err,
		ShortURL: shortURL,
	}
}

func NewURLService(repo repository.Repository, config *config.Config) *URLService {
	return &URLService{
		repo:   repo,
		config: config,
	}
}

func (s *URLService) Shorten(original string, userID string) (models.ShortURL, error) {
	short := fmt.Sprintf("%x", sha1.Sum([]byte(original)))[:6]
	shortURL := models.ShortURL{
		OriginalURL: original,
		ID:          short,
		CreatedByID: userID,
	}
	err := s.repo.Save(shortURL)
	var notUniqueErr *repository.NotUniqueURLError
	if errors.As(err, &notUniqueErr) {
		return shortURL, NewShorteningError(shortURL, err)
	}
	if err != nil {
		return models.ShortURL{}, err
	}
	return shortURL, nil
}

func (s *URLService) GetOriginalLink(shorten string) (models.ShortURL, error) {
	original, err := s.repo.Get(shorten)
	if err != nil {
		return models.ShortURL{}, err
	}
	return original, nil
}


func (s *URLService) GetUrlsCreatedBy(userID string) ([]models.ShortURL, error) {
	return s.repo.GetUsersUrls(userID)
}
func (s *URLService) HealthCheck() error {
	if err := s.repo.Check(); err != nil {
		return err
	}
	return nil
}

func (s *URLService) ShortenBatch(batch []models.ShortURL, userID string) ([]models.ShortURL, error) {
	for i, URL := range batch{
		short := fmt.Sprintf("%x", sha1.Sum([]byte(URL.OriginalURL)))[:6]
		batch[i].ID = short
		batch[i].CreatedByID = userID
	}
	if err := s.repo.SaveBatch(batch); err != nil {
		return nil, err
	}
	return batch, nil
}


func (s *URLService) FormatShortURL(urlID string) string {
	return fmt.Sprintf("%s/%s", s.config.BaseURL, urlID)
}

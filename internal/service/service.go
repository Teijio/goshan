package service

import (
	"crypto/sha1"
	"fmt"

	"github.com/Teijio/goshan/internal/repository"
)

type URLService struct {
	repo repository.Repository
}

func NewURLService(repo repository.Repository) *URLService {
	return &URLService{
		repo: repo,
	}
}

func (s *URLService) Shorten(original string) (string, error) {
	short := fmt.Sprintf("%x", sha1.Sum([]byte(original)))[:6]
	if err := s.repo.Save(short, original); err != nil {
		return "", err
	}
	return short, nil
}

func (s *URLService) GetOriginalLink(shorten string) (string, error) {
	original, err := s.repo.Get(shorten)
	if err != nil {
		return "", err
	}
	return original, nil
}

func (s *URLService) HealthCheck() error {
	if err := s.repo.Check(); err != nil {
		return err
	}
	return nil
}

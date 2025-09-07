package service

import (
	"crypto/sha1"
	"fmt"

	"github.com/Teijio/goshan/internal/repository"
)

type URLService struct {
	repo *repository.URLRepository
}

func NewURLService(repo *repository.URLRepository) *URLService {
	return &URLService{
		repo: repo,
	}
}

func (s *URLService) Shorten(original string) string {
	short := fmt.Sprintf("%x", sha1.Sum([]byte(original)))[:6]
	s.repo.Save(short, original)
	return short
}

func (s *URLService) GetOriginalLink(shorten string) (string, bool) {
	original, ok := s.repo.Get(shorten)
	return original, ok
}

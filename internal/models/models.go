package models

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/Teijio/goshan/internal/config"
)

type ShortURL struct {
	OriginalURL string `json:"url"`
	ID          string `json:"id"`
	Cfg         *config.Config
}

type Response struct {
	Result string `json:"result"`
}

func (r *ShortURL) Validate() error {
	if strings.TrimSpace(r.OriginalURL) == "" {
		return fmt.Errorf("URL is required")
	}
	if _, err := url.ParseRequestURI(r.OriginalURL); err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	return nil
}

package models

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Teijio/goshan/internal/config"
)

type ShortURL struct {
	OriginalURL   string     `db:"original_url" json:"url"`
	ID            string     `db:"id" json:"id"`
	CreatedByID   string     `db:"created_by" json:"created_by"`
	CorrelationID string     `db:"correlation_id" json:"correlation_id"`
	IsDeleted     *time.Time `db:"is_deleted" json:"is_deleted"` // nil = не удалено, значение = время удаления

	Cfg *config.Config
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

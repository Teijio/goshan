package models

import (
	"fmt"
	"net/url"
	"strings"
)

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

func (r *Request) Validate() error {
	if strings.TrimSpace(r.URL) == "" {
		return fmt.Errorf("URL is required")
	}
	if _, err := url.ParseRequestURI(r.URL); err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	return nil
}



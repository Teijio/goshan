package config

import (
	"flag"
)

type Config struct {
	ServerAddress string `json:"server_address"`
	BaseURL       string `json:"base_url"`
}

func LoadConfig() *Config {
	var cfg Config

	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "Base URL for short links")

	flag.Parse()

	return &cfg
}

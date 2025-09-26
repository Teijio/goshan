package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress string `json:"server_address" env:"SERVER_ADDRESS"`
	BaseURL       string `json:"base_url" env:"BASE_URL"`
	FilePath        string `json:"file_path" env:"FILE_PATH"`
}

func LoadConfig() *Config {
	var cfg Config

	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "Base URL for short links")
	flag.StringVar(&cfg.FilePath, "f", "", "File to restore DB")
	flag.Parse()
	// Приоритет у .env перемененных, т.е. если выше нашлись с флагом - ниже строка перезатрет их
	env.Parse(&cfg)

	return &cfg
}

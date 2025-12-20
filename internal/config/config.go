package config

import (
	"crypto/aes"
	"flag"

	"github.com/Teijio/goshan/internal/service/random"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress string `json:"server_address" env:"SERVER_ADDRESS"`
	BaseURL       string `json:"base_url" env:"BASE_URL"`
	FilePath      string `json:"file_path" env:"FILE_PATH"`
	DatabaseDSN   string `json:"database_dsn" env:"DATABASE_DSN"`
	EncryptionKey []byte
}

// "host=localhost user=prac password=prac dbname=prac sslmode=disable"
func genEncKey(c *Config) {
	size := 2 * aes.BlockSize //nolint:gomnd
	randomKey, err := random.GenerateRandom(size)
	if err != nil {
		randomKey = make([]byte, size)
	}
	c.EncryptionKey = randomKey

}

func LoadConfig() *Config {
	var cfg Config

	genEncKey(&cfg)

	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "Base URL for short links")
	flag.StringVar(&cfg.FilePath, "f", "", "File to restore DB")
	flag.StringVar(&cfg.DatabaseDSN, "d", "host=localhost user=prac password=prac dbname=prac sslmode=disable", "database dsn for connecting to postgres")
	flag.Parse()
	// Приоритет у .env перемененных, т.е. если выше нашлись с флагом - ниже строка перезатрет их
	env.Parse(&cfg)
	return &cfg
}

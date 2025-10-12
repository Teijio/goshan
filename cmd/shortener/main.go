package main

import (
	"github.com/Teijio/goshan/internal/config"
	"github.com/Teijio/goshan/internal/repository"
	"github.com/Teijio/goshan/internal/server"
	"github.com/Teijio/goshan/internal/service"
	"github.com/Teijio/goshan/internal/service/random"
)

func main() {
	cfg := config.LoadConfig()

	rep := repository.GetRepository(cfg)
	service := service.NewURLService(rep, cfg)
	userIDGenerator := &random.UUIDGenerator{}
	srv := server.New(cfg, service, userIDGenerator)

	srv.Run()
}

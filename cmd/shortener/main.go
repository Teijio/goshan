package main

import (
	"fmt"

	"github.com/Teijio/goshan/internal/config"
	"github.com/Teijio/goshan/internal/repository"
	"github.com/Teijio/goshan/internal/server"
	"github.com/Teijio/goshan/internal/service"
	"github.com/Teijio/goshan/internal/service/random"
)
var (
	buildVersion = "N/A" //nolint:gochecknoglobals
	buildDate    = "N/A" //nolint:gochecknoglobals
	buildCommit  = "N/A" //nolint:gochecknoglobals
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	cfg := config.LoadConfig()

	rep := repository.GetRepository(cfg)
	service := service.NewURLService(rep, cfg)
	userIDGenerator := &random.UUIDGenerator{}
	srv := server.New(cfg, service, userIDGenerator)

	srv.Run()
}

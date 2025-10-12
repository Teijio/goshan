package server

import (
	"net/http"

	"github.com/Teijio/goshan/internal/config"
	"github.com/Teijio/goshan/internal/handlers"
	"github.com/Teijio/goshan/internal/service"
	"github.com/Teijio/goshan/internal/service/random"
	"go.uber.org/zap"
)

type Server struct {
	config  *config.Config
	service *service.URLService
	logger  *zap.Logger
	idGenerator random.UserIDGenerator
}

func (s *Server) Run() {
	r := handlers.NewRouter(s.service, s.config, s.logger, s.idGenerator)

	httpServer := &http.Server{
		Addr:    s.config.ServerAddress,
		Handler: r,
	}

	if err := httpServer.ListenAndServe(); err != nil {
		s.logger.Fatal("Server failed to start", zap.Error(err))
	}
}

func New(config *config.Config, service *service.URLService, generator *random.UUIDGenerator) *Server {
	logger, err := zap.NewDevelopment()
	defer logger.Sync()
	if err != nil {
		panic(err)
	}
	return &Server{
		config:  config,
		service: service,
		logger: logger,
		idGenerator: generator,
	}
}

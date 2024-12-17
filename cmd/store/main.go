package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"codesignal/internal/config"
	"codesignal/internal/repository"
	"codesignal/internal/router"
	"codesignal/internal/server"
)

func main() {
	logger := zerolog.New(os.Stderr).
		Level(zerolog.DebugLevel).
		With().
		Timestamp().
		Logger()

	appConfig, err := config.LoadFromEnv()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load env vars")
	}

	repo, err := repository.NewKeyValueStore(logger)
	if err != nil {
		log.Error().Err(err).Msg("failed to create repository")
	}

	httpRouter := router.New(logger, repo, appConfig)

	httpServer := server.New(logger, appConfig.Server, httpRouter)

	if err := httpServer.Run(); err != nil {
		logger.Fatal().Err(err).Msg("server failure")
	}
}

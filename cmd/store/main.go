package main

import (
	"os"

	"github.com/rs/zerolog"

	"codesignal/internal/config"
	"codesignal/internal/router"
	"codesignal/internal/server"
)

func main() {
	logger := zerolog.New(os.Stderr).
		Level(zerolog.DebugLevel).
		With().
		Timestamp().
		Logger()

	configuration, err := config.LoadFromEnv()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load env vars")
	}

	httpRouter := router.New(logger)

	httpServer := server.New(logger, configuration.Server, httpRouter)

	if err := httpServer.Run(); err != nil {
		logger.Fatal().Err(err).Msg("server failure")
	}
}

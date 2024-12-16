// Package server provides the implementation of an HTTP server with graceful shutdown capabilities.
//
// This package defines the Server struct which encapsulates the details of the HTTP server, including
// its configuration settings and handler. The Server uses zerolog for structured logging and supports
// customizable configuration through environment variables.
//
// The Config struct holds the configuration settings for the HTTP server, such as the address, read
// and write timeouts, and shutdown timeout, all of which can be set through environment variables.
//
// The New function initializes and returns a new instance of the Server with the provided logger,
// configuration, and HTTP handler. The Run method starts the HTTP server and handles graceful shutdowns
// in response to system signals.
package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
)

type (
	// Server encapsulates the details of an HTTP server.
	Server struct {
		logger  zerolog.Logger
		config  Config
		handler http.Handler
	}

	// Config holds the configuration settings for the HTTP Server.
	Config struct {
		Address         string        `envconfig:"ADDRESS" default:"0.0.0.0:8000"`
		ReadTimeout     time.Duration `envconfig:"READ_TIMEOUT" default:"5s"`
		WriteTimeout    time.Duration `envconfig:"WRITE_TIMEOUT" default:"5s"`
		ShutdownTimeout time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"5s"`
	}
)

// New returns a new HTTP Server.
func New(log zerolog.Logger, config Config, handler http.Handler) *Server {
	return &Server{
		config:  config,
		handler: handler,
		logger:  log,
	}
}

// Run will start the HTTP Server and will handle shutdowns gracefully.
// TODO: Add shutdown hook for repository store Close().
func (s *Server) Run() error {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	api := &http.Server{
		Addr:         s.config.Address,
		Handler:      s.handler,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
	}

	serverErrors := make(chan error, 1)

	go func() {
		s.logger.Info().Msgf("server listening on port %q", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server encountered an error: %w", err)
	case sig := <-shutdown:
		s.logger.Info().Msgf("server shutting down after receiving %+v", sig)

		ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
		defer cancel()

		if err := api.Shutdown(ctx); err != nil {
			_ = api.Close()
			return fmt.Errorf("server failed to shutdown gracefully: %w", err)
		}
	}

	return nil
}

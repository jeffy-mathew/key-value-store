// Package config provides the configuration management for the service.
//
// This package includes functionality to load configuration parameters
// from environment variables, using the envconfig package. It ensures
// that the service can be configured via environment variables, which
// are automatically loaded from a .env file using the godotenv package.
//
// The LoadFromEnv function is used to load these configurations from
// the operating system's environment variables.
package config

import (
	_ "github.com/joho/godotenv/autoload" // Autoload env vars from a .env file.
	"github.com/kelseyhightower/envconfig"

	"codesignal/internal/server"
)

// Config contains all the config
// parameters that this service uses.
type Config struct {
	Server server.Config `envconfig:"SERVER"`
}

// LoadFromEnv will load the env vars from the OS.
func LoadFromEnv() (*Config, error) {
	cfg := &Config{}
	err := envconfig.Process("", cfg)
	return cfg, err
}

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
	"time"

	_ "github.com/joho/godotenv/autoload" // Autoload env vars from a .env file.
	"github.com/kelseyhightower/envconfig"

	"codesignal/internal/server"
)

// Config contains all the config
// parameters that this service uses.
type Config struct {
	Server server.Config `envconfig:"SERVER"`
	// MaxKeyLength is the maximum length of a key in characters.
	MaxKeyLength int `envconfig:"MAX_KEY_LENGTH"`
	// MaxValueSize is the maximum size of a value in bytes.
	MaxValueSize int `envconfig:"MAX_VALUE_SIZE"`
	// SyncInterval is the interval to sync data to disk.
	SyncInterval time.Duration `envconfig:"SYNC_INTERVAL" default:"1m"`
	// DataFile is the path to the data file.
	DataFile string `envconfig:"DATA_FILE"`
}

func (c *Config) GetMaxKeyLength() int {
	if c == nil {
		return 0
	}

	return c.MaxKeyLength
}

func (c *Config) GetMaxValueSize() int {
	if c == nil {
		return 0
	}

	return c.MaxValueSize
}

// LoadFromEnv will load the env vars from the OS.
func LoadFromEnv() (*Config, error) {
	cfg := &Config{}
	err := envconfig.Process("", cfg)
	return cfg, err
}

package config

import (
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	// GRPCPort is the port on which the gRPC server listens
	GRPCPort string

	// HTTPPort is the port on which the HTTP gateway listens
	HTTPPort string

	// LogLevel for logging
	LogLevel string

	// Environment for the application
	Environment string

	// LocalDataStorage is the path to the services data file
	LocalDataStorage string

	// CORSOrigins is a comma-separated list of allowed CORS origins
	CORSOrigins string
}

// Load reads environment variables and returns the Config
func Load() (*Config, error) {
	cfg := &Config{
		GRPCPort:         getEnv("GRPC_PORT", "9000"),
		HTTPPort:         getEnv("HTTP_PORT", "8000"),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
		Environment:      getEnv("ENVIRONMENT", "development"),
		LocalDataStorage: getEnv("DATA_FILE_PATH", "data/services.yaml"),
		CORSOrigins:      getEnv("CORS_ORIGINS", "*"),
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

// Validate checks required fields and returns an error if misconfigured
func (c *Config) Validate() error {
	if c.GRPCPort == "" {
		return fmt.Errorf("GRPC_PORT cannot be empty")
	}
	if c.HTTPPort == "" {
		return fmt.Errorf("HTTP_PORT cannot be empty")
	}
	if c.LocalDataStorage == "" {
		return fmt.Errorf("DATA_FILE_PATH cannot be empty")
	}

	// Validate data file exists
	if _, err := os.Stat(c.LocalDataStorage); os.IsNotExist(err) {
		return fmt.Errorf("data file does not exist: %s", c.LocalDataStorage)
	}

	return nil
}

// getEnv returns the value of the environment variable or fallback if not set
func getEnv(key, fallback string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return fallback
}

// GetDataFileAbsPath returns the absolute path to the data file
func (c *Config) GetDataFileAbsPath() (string, error) {
	if filepath.IsAbs(c.LocalDataStorage) {
		return c.LocalDataStorage, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	return filepath.Join(cwd, c.LocalDataStorage), nil
}

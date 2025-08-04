package config

import (
	"fmt"
	"os"
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
}

// Load reads environment variables and returns the Config
func Load() *Config {
	cfg := &Config{
		GRPCPort:    getEnv("GRPC_PORT", "9000"),
		HTTPPort:    getEnv("HTTP_PORT", "8000"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		Environment: getEnv("ENVIRONMENT", "development"),
	}

	if err := cfg.Validate(); err != nil {
		panic(fmt.Sprintf("invalid config: %v", err))
	}

	return cfg
}

// Validate checks required fields and returns an error if misconfigured
func (c *Config) Validate() error {
	if c.GRPCPort == "" {
		return fmt.Errorf("GRPC_PORT cannot be empty")
	}
	if c.HTTPPort == "" {
		return fmt.Errorf("HTTP_PORT cannot be empty")
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

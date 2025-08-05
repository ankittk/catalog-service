package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
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

	// JWTSecretKey is the secret key for JWT token signing
	JWTSecretKey string

	// JWTTokenDuration is the duration for JWT tokens
	JWTTokenDuration time.Duration

	// EnableAuth enables JWT authentication
	EnableAuth bool
}

// Load reads environment variables and returns the Config
func Load() (*Config, error) {
	cfg := &Config{
		GRPCPort:         getEnv("GRPC_PORT", "9000"),
		HTTPPort:         getEnv("HTTP_PORT", "8000"),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
		Environment:      getEnv("ENVIRONMENT", "development"),
		LocalDataStorage: getEnv("LOCAL_DATA_STORAGE", "data/services.yaml"),
		CORSOrigins:      getEnv("CORS_ORIGINS", "*"),
		JWTSecretKey:     getEnv("JWT_SECRET_KEY", ""),
		EnableAuth:       getEnvBool("ENABLE_AUTH", false),
	}

	// Parse JWT token duration
	tokenDurationStr := getEnv("JWT_TOKEN_DURATION", "24h")
	tokenDuration, err := time.ParseDuration(tokenDurationStr)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_TOKEN_DURATION: %w", err)
	}
	cfg.JWTTokenDuration = tokenDuration

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
		return fmt.Errorf("LOCAL_DATA_STORAGE cannot be empty")
	}

	// Validate data file exists
	if _, err := os.Stat(c.LocalDataStorage); os.IsNotExist(err) {
		return fmt.Errorf("data file does not exist: %s", c.LocalDataStorage)
	}

	// Validate JWT configuration if auth is enabled
	if c.EnableAuth {
		if c.JWTSecretKey == "" {
			return fmt.Errorf("JWT_SECRET_KEY is required when ENABLE_AUTH is true")
		}
		if c.JWTTokenDuration <= 0 {
			return fmt.Errorf("JWT_TOKEN_DURATION must be positive")
		}
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

// getEnvBool returns the boolean value of the environment variable or fallback if not set
func getEnvBool(key string, fallback bool) bool {
	if val, exists := os.LookupEnv(key); exists {
		return val == "true" || val == "1" || val == "yes"
	}
	return fallback
}

// GetDataFileAbsPath returns the absolute path to the data file
func (c *Config) GetDataFileAbsPath() (string, error) {
	if filepath.IsAbs(c.LocalDataStorage) {
		return c.LocalDataStorage, nil
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	return filepath.Join(cwd, c.LocalDataStorage), nil
}

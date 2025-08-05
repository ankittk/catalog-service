package main

import (
	"os"

	"github.com/ankittk/catalog-service/internal/app"
	"github.com/ankittk/catalog-service/internal/config"
	"github.com/ankittk/catalog-service/internal/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		os.Stderr.WriteString("Failed to load configuration: " + err.Error() + "\n")
		os.Exit(1)
	}

	// Initialize logger with config
	if err := logger.Init(cfg.LogLevel); err != nil {
		os.Stderr.WriteString("Failed to initialize logger: " + err.Error() + "\n")
		os.Exit(1)
	}
	defer logger.Sync() // Sync logger on exit

	logger.Get().Infow("Starting catalog service",
		"environment", cfg.Environment,
		"log_level", cfg.LogLevel)

	// Create and start application
	application := app.NewApp(cfg)
	if err := application.Start(); err != nil {
		logger.Get().Fatalw("Failed to start application", "error", err)
	}

	// Wait for shutdown signal to gracefully shutdown the application
	application.WaitForShutdown()
}

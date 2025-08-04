package main

import (
	"github.com/ankittk/catalog-service/internal/app"
	"github.com/ankittk/catalog-service/internal/config"
	"github.com/ankittk/catalog-service/internal/logger"
)

func main() {
	// Initialize logger
	logger.Init()

	// Load configuration
	cfg := config.Load()

	// Create and start application
	application := app.NewApp(cfg)
	if err := application.Start(); err != nil {
		logger.Get().Fatalw("failed to start application", "error", err)
	}

	// Wait for shutdown signal
	application.WaitForShutdown()
}

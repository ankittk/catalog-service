package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	grpcserver "github.com/ankittk/catalog-service/internal/api/grpc"
	"github.com/ankittk/catalog-service/internal/config"
	"github.com/ankittk/catalog-service/internal/logger"
	v1 "github.com/ankittk/catalog-service/proto/v1"
)

// App represents the application instance
type App struct {
	config     *config.Config
	grpcServer *grpc.Server
	httpServer *http.Server
	grpcAddr   string
	httpAddr   string
}

// NewApp creates a new application instance
func NewApp(cfg *config.Config) *App {
	return &App{
		config:   cfg,
		grpcAddr: fmt.Sprintf(":%s", cfg.GRPCPort),
		httpAddr: fmt.Sprintf(":%s", cfg.HTTPPort),
	}
}

// Start initializes and starts the application
func (a *App) Start() error {
	logger.Get().Infow("starting catalog service",
		"grpc_port", a.config.GRPCPort,
		"http_port", a.config.HTTPPort)

	// Initialize gRPC server
	if err := a.initGRPCServer(); err != nil {
		return fmt.Errorf("failed to initialize gRPC server: %w", err)
	}

	// Initialize HTTP server
	if err := a.initHTTPServer(); err != nil {
		return fmt.Errorf("failed to initialize HTTP server: %w", err)
	}

	// Start servers
	if err := a.startServers(); err != nil {
		return fmt.Errorf("failed to start servers: %w", err)
	}

	return nil
}

// initGRPCServer initializes the gRPC server
func (a *App) initGRPCServer() error {
	// Create gRPC server
	a.grpcServer = grpc.NewServer()

	// Initialize catalog server
	yamlData, err := os.ReadFile("data/services.yaml")
	if err != nil {
		return fmt.Errorf("failed to read services.yaml: %w", err)
	}

	catalogServer, err := grpcserver.NewCatalogServerFromYAML(yamlData)
	if err != nil {
		return fmt.Errorf("failed to create catalog server: %w", err)
	}

	// Register services
	v1.RegisterCatalogServiceServer(a.grpcServer, catalogServer)

	// Enable reflection for development
	if a.config.Environment == "development" {
		reflection.Register(a.grpcServer)
	}

	return nil
}

// initHTTPServer initializes the HTTP server with gRPC gateway
func (a *App) initHTTPServer() error {
	// Create HTTP server
	a.httpServer = &http.Server{
		Addr:    a.httpAddr,
		Handler: a.createHTTPHandler(),
	}

	return nil
}

// createHTTPHandler creates the HTTP handler with gRPC gateway
func (a *App) createHTTPHandler() http.Handler {
	mux := http.NewServeMux()

	// Create gRPC gateway mux
	gwmux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	// Register gRPC gateway handlers
	if err := v1.RegisterCatalogServiceHandlerFromEndpoint(
		context.Background(),
		gwmux,
		a.grpcAddr,
		opts,
	); err != nil {
		logger.Get().Errorw("failed to register gRPC gateway", "error", err)
		return mux
	}

	// CORS headers setup if needed for the frontend
	mux.HandleFunc("/v1/", func(w http.ResponseWriter, r *http.Request) {
		// Handle CORS if needed
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Serve the gRPC Gateway handler
		gwmux.ServeHTTP(w, r)
	})

	// Add health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	return mux
}

// startServers starts both gRPC and HTTP servers
func (a *App) startServers() error {
	// Start gRPC server
	go func() {
		lis, err := net.Listen("tcp", a.grpcAddr)
		if err != nil {
			logger.Get().Fatalw("failed to listen for gRPC", "error", err)
		}

		logger.Get().Infow("gRPC server listening", "address", a.grpcAddr)
		if err := a.grpcServer.Serve(lis); err != nil {
			logger.Get().Fatalw("failed to serve gRPC", "error", err)
		}
	}()

	// Start HTTP server
	go func() {
		logger.Get().Infow("HTTP server listening", "address", a.httpAddr)
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Get().Fatalw("failed to serve HTTP", "error", err)
		}
	}()

	return nil
}

// Stop gracefully shuts down the application
func (a *App) Stop() error {
	logger.Get().Info("shutting down application...")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop HTTP server
	if a.httpServer != nil {
		if err := a.httpServer.Shutdown(ctx); err != nil {
			logger.Get().Errorw("failed to shutdown HTTP server", "error", err)
		}
	}

	// Stop gRPC server
	if a.grpcServer != nil {
		a.grpcServer.GracefulStop()
	}

	logger.Get().Info("application stopped")
	return nil
}

// WaitForShutdown waits for shutdown signals
func (a *App) WaitForShutdown() {
	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Get().Info("received shutdown signal")
	if err := a.Stop(); err != nil {
		logger.Get().Errorw("error during shutdown", "error", err)
	}
}

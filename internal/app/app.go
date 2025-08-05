package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	grpcserver "github.com/ankittk/catalog-service/internal/api/grpc"
	"github.com/ankittk/catalog-service/internal/auth"
	authhandler "github.com/ankittk/catalog-service/internal/auth"
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
	jwtManager *auth.JWTManager
}

// NewApp creates a new application instance
func NewApp(cfg *config.Config) *App {
	app := &App{
		config:   cfg,
		grpcAddr: fmt.Sprintf(":%s", cfg.GRPCPort),
		httpAddr: fmt.Sprintf(":%s", cfg.HTTPPort),
	}

	// Initialize JWT manager if authentication is enabled
	if cfg.EnableAuth {
		app.jwtManager = auth.NewJWTManager(cfg.JWTSecretKey, cfg.JWTTokenDuration)
		logger.Get().Infow("JWT authentication enabled",
			"token_duration", cfg.JWTTokenDuration.String())
	} else {
		logger.Get().Info("JWT authentication disabled")
	}

	return app
}

// Start initializes and starts the application
func (a *App) Start() error {
	logger.Get().Infow("Starting catalog service",
		"grpc_port", a.config.GRPCPort,
		"http_port", a.config.HTTPPort,
		"data_file", a.config.LocalDataStorage,
		"auth_enabled", a.config.EnableAuth)

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
	// Create gRPC server with authentication interceptor if enabled
	var opts []grpc.ServerOption
	if a.config.EnableAuth && a.jwtManager != nil {
		opts = append(opts, grpc.UnaryInterceptor(a.jwtManager.GRPCUnaryInterceptor()))
		logger.Get().Info("gRPC server configured with JWT authentication")
	}

	a.grpcServer = grpc.NewServer(opts...)

	// Get absolute path to data file
	localDataStorage, err := a.config.GetDataFileAbsPath()
	if err != nil {
		return fmt.Errorf("failed to resolve data file path: %w", err)
	}

	// Read YAML data with proper error handling
	yamlData, err := os.ReadFile(localDataStorage)
	if err != nil {
		return fmt.Errorf("failed to read data file %s: %w", localDataStorage, err)
	}

	catalogServer, err := grpcserver.NewCatalogServerFromYAML(yamlData)
	if err != nil {
		return fmt.Errorf("failed to create catalog server: %w", err)
	}

	// Register services
	v1.RegisterCatalogServiceServer(a.grpcServer, catalogServer)

	// Enable reflection for development as it is useful for development and debugging
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
		logger.Get().Errorw("Failed to register gRPC gateway", "error", err)
		return mux
	}

	// CORS middleware
	corsMiddleware := a.createCORSMiddleware()

	// Authentication middleware
	var authMiddleware func(http.Handler) http.Handler
	if a.config.EnableAuth && a.jwtManager != nil {
		authMiddleware = a.jwtManager.HTTPMiddleware
		logger.Get().Info("HTTP server configured with JWT authentication")
	} else {
		authMiddleware = func(next http.Handler) http.Handler {
			return next
		}
	}

	// Authentication endpoints (no auth required)
	if a.config.EnableAuth && a.jwtManager != nil {
		authHandler := authhandler.NewAuthHandler(a.jwtManager)
		mux.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
			corsMiddleware(w, r)
			authHandler.Login(w, r)
		})
	}

	// API routes with authentication and CORS
	mux.HandleFunc("/v1/", func(w http.ResponseWriter, r *http.Request) {
		corsMiddleware(w, r)
		authMiddleware(gwmux).ServeHTTP(w, r)
	})

	// Health check endpoint (no auth required)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		corsMiddleware(w, r)
		if r.Method == "OPTIONS" {
			return
		}

		// Return service health information
		healthResponse := map[string]interface{}{
			"status":       "healthy",
			"service":      "catalog-service",
			"timestamp":    time.Now().UTC().Format(time.RFC3339),
			"version":      "1.0.0",
			"auth_enabled": a.config.EnableAuth,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"%s","service":"%s","timestamp":"%s","version":"%s","auth_enabled":%t}`,
			healthResponse["status"],
			healthResponse["service"],
			healthResponse["timestamp"],
			healthResponse["version"],
			healthResponse["auth_enabled"])
	})

	return mux
}

// createCORSMiddleware creates a CORS middleware function
func (a *App) createCORSMiddleware() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse CORS origins from config
		origins := strings.Split(a.config.CORSOrigins, ",")
		origin := r.Header.Get("Origin")

		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range origins {
			allowedOrigin = strings.TrimSpace(allowedOrigin)
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}

		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin) // Allow the origin
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")               // Allow these methods
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With") // Allow these headers
		w.Header().Set("Access-Control-Allow-Credentials", "true")                                      // Allow credentials for CORS
		w.Header().Set("Access-Control-Max-Age", "86400")                                               // 24 hours for CORS

		// Handle preflight requests for CORS
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
	}
}

// startServers starts both gRPC and HTTP servers
func (a *App) startServers() error {
	// Start gRPC server
	go func() {
		lis, err := net.Listen("tcp", a.grpcAddr)
		if err != nil {
			logger.Get().Fatalw("Failed to listen for gRPC", "error", err)
		}

		logger.Get().Infow("gRPC server listening", "address", a.grpcAddr)
		if err := a.grpcServer.Serve(lis); err != nil {
			logger.Get().Fatalw("Failed to serve gRPC", "error", err)
		}
	}()

	// Start HTTP server
	go func() {
		logger.Get().Infow("HTTP server listening", "address", a.httpAddr)
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Get().Fatalw("Failed to serve HTTP", "error", err)
		}
	}()

	return nil
}

// Stop gracefully shuts down the application
func (a *App) Stop() error {
	logger.Get().Info("Shutting down application...")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop HTTP server
	if a.httpServer != nil {
		if err := a.httpServer.Shutdown(ctx); err != nil {
			logger.Get().Errorw("Failed to shutdown HTTP server", "error", err)
		}
	}

	// Stop gRPC server
	if a.grpcServer != nil {
		a.grpcServer.GracefulStop()
	}

	logger.Get().Info("Application stopped")
	return nil
}

// WaitForShutdown waits for shutdown signals
func (a *App) WaitForShutdown() {
	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Get().Info("Received shutdown signal")
	if err := a.Stop(); err != nil {
		logger.Get().Errorw("Error during shutdown", "error", err)
	}
}

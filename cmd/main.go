package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ankittk/catalog-service/internal/config"
	"github.com/ankittk/catalog-service/internal/server"
	v1 "github.com/ankittk/catalog-service/proto/v1"
)

var (
	catalogServer *server.Server
	logger        *zap.SugaredLogger
)

// init loads the in-memory data and initializes the catalog server
func init() {
	// Initialize logger using Uber's Zap
	zapLogger, _ := zap.NewProduction()
	logger = zapLogger.Sugar()

	// Read YAML file and initialize service store
	yamlData, err := os.ReadFile("data/services.yaml")
	if err != nil {
		logger.Fatalw("failed to read services.yaml", "error", err)
	}

	catalogServer, err = server.NewCatalogServerFromYAML(yamlData)
	if err != nil {
		logger.Fatalw("failed to parse services.yaml", "error", err)
	}
}

func main() {
	// Load configuration from environment variables or defaults
	cfg := config.Load()

	logger.Infow("starting catalog service", "grpc_port", cfg.GRPCPort, "http_port", cfg.HTTPPort)

	if err := startServers(cfg); err != nil {
		logger.Fatalw("server stopped with error", "error", err)
	}
}

// startServers starts both gRPC and HTTP servers with graceful shutdown
func startServers(cfg *config.Config) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a channel to listen for errors
	errs := make(chan error, 2)

	// Start gRPC server
	go func() {
		errs <- runGRPCServer(ctx, cfg)
	}()

	// Start HTTP gateway
	//  It will forward requests to the gRPC server using gRPC-Gateway
	// This allows HTTP clients to interact with the gRPC service
	// It also allows for easier integration with web clients
	go func() {
		errs <- runHTTPGateway(ctx, cfg)
	}()

	// Listen for termination signal and gracefully shut down the servers
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		sig := <-c
		logger.Infow("shutdown signal received", "signal", sig)
		cancel()
	}()

	// Block until something happens or an error occurs
	if err := <-errs; err != nil {
		return err
	}

	logger.Infow("server stopped gracefully")
	return nil
}

// runGRPCServer starts the gRPC server and listens for incoming connections.
func runGRPCServer(ctx context.Context, cfg *config.Config) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		return fmt.Errorf("failed to listen on gRPC port: %w", err)
	}

	// Create a new gRPC server instance
	grpcServer := grpc.NewServer()

	// Register the CatalogService server with the gRPC server
	v1.RegisterCatalogServiceServer(grpcServer, catalogServer)

	go func() {
		<-ctx.Done()
		logger.Info("shutting down gRPC server")
		grpcServer.GracefulStop()
	}()

	logger.Infof("gRPC server listening at :%s", cfg.GRPCPort)
	return grpcServer.Serve(lis)
}

// runHTTPGateway starts the HTTP gateway that forwards requests to the gRPC server.
func runHTTPGateway(ctx context.Context, cfg *config.Config) error {
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	// Register the HTTP handler with the gRPC backend
	if err := v1.RegisterCatalogServiceHandlerFromEndpoint(
		ctx, mux,
		fmt.Sprintf("localhost:%s", cfg.GRPCPort),
		opts,
	); err != nil {
		return fmt.Errorf("failed to register gateway handler: %w", err)
	}

	srv := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: mux,
	}

	// Start graceful shutdown goroutine
	go func() {
		<-ctx.Done()
		logger.Info("Received shutdown signal. Shutting down HTTP gateway...")

		ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctxShutDown); err != nil {
			logger.Errorw("HTTP gateway shutdown error", "error", err)
		} else {
			logger.Info("HTTP gateway shutdown completed.")
		}
	}()

	logger.Infof("HTTP gateway listening on :%s", cfg.HTTPPort)

	// Start the HTTP server and return its error if any
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http gateway server failed: %w", err)
	}

	return nil
}

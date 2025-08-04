package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v3"

	"github.com/ankittk/catalog-service/internal/logger"
	"github.com/ankittk/catalog-service/internal/model"
	"github.com/ankittk/catalog-service/internal/service"
	v1 "github.com/ankittk/catalog-service/proto/v1"
)

// Server implements the CatalogService gRPC server
type Server struct {
	// Embed the unimplemented server to ensure forward compatibility
	v1.UnimplementedCatalogServiceServer
	svc *service.CatalogService
}

// NewCatalogServerFromYAML creates a new server by parsing YAML data
func NewCatalogServerFromYAML(yamlData []byte) (*Server, error) {
	logger.Get().Info("Initializing catalog server from YAML data")

	var sf model.ServicesFile
	if err := yaml.Unmarshal(yamlData, &sf); err != nil {
		logger.Get().Errorw("Failed to parse services.yaml", "error", err)
		return nil, fmt.Errorf("failed to parse services.yaml: %w", err)
	}

	// Create a local store with the parsed services
	store := &model.Store{}
	store.SetServices(sf.Services)
	catalogService := service.NewCatalogService(store)

	logger.Get().Infow("Catalog server initialized successfully", "services_count", len(sf.Services))

	return &Server{svc: catalogService}, nil
}

// ListServices returns a list of all services
func (s *Server) ListServices(ctx context.Context, req *v1.ListServicesRequest) (*v1.ListServicesResponse, error) {
	logger.Get().Infow("ListServices request received",
		"page_size", req.GetPageSize(),
		"page_token", req.GetPageToken(),
		"organization_id", req.GetOrganizationId(),
		"search_query", req.GetSearchQuery(),
		"sort_by", req.GetSortBy(),
		"sort_order", req.GetSortOrder())

	// Check if context is cancelled
	if ctx.Err() != nil {
		logger.Get().Warnw("ListServices context cancelled", "error", ctx.Err())
		return nil, status.Error(codes.Canceled, "request cancelled")
	}

	resp, err := s.svc.ListServices(ctx, req)
	if err != nil {
		logger.Get().Errorw("ListServices failed", "error", err)
		return nil, err // return the error as-is since service layer already sets proper gRPC status
	}

	logger.Get().Infow("ListServices completed successfully",
		"returned_count", len(resp.GetServices()),
		"total_count", resp.GetTotalCount(),
		"has_next_page", resp.GetNextPageToken() != "")

	return resp, nil
}

// GetService returns a specific service by ID
func (s *Server) GetService(ctx context.Context, req *v1.GetServiceRequest) (*v1.GetServiceResponse, error) {
	logger.Get().Infow("GetService request received", "service_id", req.GetId())

	// Check if context is cancelled
	if ctx.Err() != nil {
		logger.Get().Warnw("GetService context cancelled", "error", ctx.Err())
		return nil, status.Error(codes.Canceled, "request cancelled")
	}

	resp, err := s.svc.GetService(ctx, req)
	if err != nil {
		logger.Get().Errorw("GetService failed", "service_id", req.GetId(), "error", err)
		return nil, err
	}

	logger.Get().Infow("GetService completed successfully", "service_id", req.GetId())
	return resp, nil
}

// GetServiceVersions returns all versions of a specific service
func (s *Server) GetServiceVersions(ctx context.Context, req *v1.GetServiceVersionsRequest) (*v1.GetServiceVersionsResponse, error) {
	logger.Get().Infow("GetServiceVersions request received", "service_id", req.GetServiceId())

	// Check if context is cancelled
	if ctx.Err() != nil {
		logger.Get().Warnw("GetServiceVersions context cancelled", "error", ctx.Err())
		return nil, status.Error(codes.Canceled, "request cancelled")
	}

	resp, err := s.svc.GetServiceVersions(ctx, req)
	if err != nil {
		logger.Get().Errorw("GetServiceVersions failed", "service_id", req.GetServiceId(), "error", err)
		return nil, err
	}

	logger.Get().Infow("GetServiceVersions completed successfully",
		"service_id", req.GetServiceId(),
		"versions_count", len(resp.GetVersions()))

	return resp, nil
}

// HealthCheck returns the health status of the service
func (s *Server) HealthCheck(ctx context.Context, req *v1.HealthCheckRequest) (*v1.HealthCheckResponse, error) {
	logger.Get().Debug("HealthCheck request received")
	return &v1.HealthCheckResponse{
		Status:    "OK",
		Timestamp: timestamppb.Now(),
	}, nil
}

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
	svc     *service.CatalogService
	metrics *logger.MetricsLogger
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

	return &Server{
		svc:     catalogService,
		metrics: logger.NewMetricsLogger(),
	}, nil
}

// ListServices returns a list of all services
func (s *Server) ListServices(ctx context.Context, req *v1.ListServicesRequest) (*v1.ListServicesResponse, error) {
	// Create request logger for structured logging
	reqLogger := logger.NewRequestLogger("ListServices", "/v1/services")
	reqLogger.AddField("page_size", req.GetPageSize())
	reqLogger.AddField("page_token", req.GetPageToken())
	reqLogger.AddField("organization_id", req.GetOrganizationId())
	reqLogger.AddField("search_query", req.GetSearchQuery())
	reqLogger.AddField("sort_by", req.GetSortBy())
	reqLogger.AddField("sort_order", req.GetSortOrder())

	reqLogger.LogRequest()

	// Check if context is cancelled
	if ctx.Err() != nil {
		reqLogger.LogResponse(int(codes.Canceled), ctx.Err())
		s.metrics.LogCounter("grpc_requests_total", 1, map[string]string{
			"method": "ListServices",
			"status": "cancelled",
		})
		return nil, status.Error(codes.Canceled, "request cancelled")
	}

	resp, err := s.svc.ListServices(ctx, req)

	// Return appropriate status code based on error
	statusCode := codes.OK
	if err != nil {
		if st, ok := status.FromError(err); ok {
			statusCode = st.Code()
		} else {
			statusCode = codes.Internal
		}
	}

	reqLogger.LogResponse(int(statusCode), err)

	// Log metrics for request count
	s.metrics.LogCounter("grpc_requests_total", 1, map[string]string{
		"method": "ListServices",
		"status": statusCode.String(),
	})

	if err == nil {
		s.metrics.LogHistogram("grpc_response_size", float64(len(resp.GetServices())), map[string]string{
			"method": "ListServices",
		})
	}

	return resp, err
}

// GetService returns a specific service by ID
func (s *Server) GetService(ctx context.Context, req *v1.GetServiceRequest) (*v1.GetServiceResponse, error) {
	// Create request logger for structured logging
	reqLogger := logger.NewRequestLogger("GetService", "/v1/services/{id}")
	reqLogger.AddField("service_id", req.GetId())

	reqLogger.LogRequest()

	// Check if context is cancelled
	if ctx.Err() != nil {
		reqLogger.LogResponse(int(codes.Canceled), ctx.Err())
		s.metrics.LogCounter("grpc_requests_total", 1, map[string]string{
			"method": "GetService",
			"status": "cancelled",
		})
		return nil, status.Error(codes.Canceled, "request cancelled")
	}

	resp, err := s.svc.GetService(ctx, req)

	statusCode := codes.OK
	if err != nil {
		if st, ok := status.FromError(err); ok {
			statusCode = st.Code()
		} else {
			statusCode = codes.Internal
		}
	}

	reqLogger.LogResponse(int(statusCode), err)

	s.metrics.LogCounter("grpc_requests_total", 1, map[string]string{
		"method": "GetService",
		"status": statusCode.String(),
	})

	return resp, err
}

// GetServiceVersions returns all versions of a specific service
func (s *Server) GetServiceVersions(ctx context.Context, req *v1.GetServiceVersionsRequest) (*v1.GetServiceVersionsResponse, error) {
	// Create request logger for structured logging
	reqLogger := logger.NewRequestLogger("GetServiceVersions", "/v1/services/{id}/versions")
	reqLogger.AddField("service_id", req.GetServiceId())

	reqLogger.LogRequest()

	// Check if context is cancelled
	if ctx.Err() != nil {
		reqLogger.LogResponse(int(codes.Canceled), ctx.Err())
		s.metrics.LogCounter("grpc_requests_total", 1, map[string]string{
			"method": "GetServiceVersions",
			"status": "cancelled",
		})
		return nil, status.Error(codes.Canceled, "request cancelled")
	}

	resp, err := s.svc.GetServiceVersions(ctx, req)

	statusCode := codes.OK
	if err != nil {
		if st, ok := status.FromError(err); ok {
			statusCode = st.Code()
		} else {
			statusCode = codes.Internal
		}
	}

	reqLogger.LogResponse(int(statusCode), err)

	s.metrics.LogCounter("grpc_requests_total", 1, map[string]string{
		"method": "GetServiceVersions",
		"status": statusCode.String(),
	})

	if err == nil {
		s.metrics.LogHistogram("grpc_response_size", float64(len(resp.GetVersions())), map[string]string{
			"method": "GetServiceVersions",
		})
	}

	return resp, err
}

// HealthCheck returns the health status of the service
func (s *Server) HealthCheck(ctx context.Context, req *v1.HealthCheckRequest) (*v1.HealthCheckResponse, error) {
	// Create request logger for structured logging
	reqLogger := logger.NewRequestLogger("HealthCheck", "/v1/health")
	reqLogger.LogRequest()

	// Check if context is cancelled
	if ctx.Err() != nil {
		reqLogger.LogResponse(int(codes.Canceled), ctx.Err())
		s.metrics.LogCounter("grpc_requests_total", 1, map[string]string{
			"method": "HealthCheck",
			"status": "cancelled",
		})
		return nil, status.Error(codes.Canceled, "request cancelled")
	}

	// Perform basic health checks
	healthStatus := "OK"

	// Check if service data is available
	if s.svc == nil {
		healthStatus = "ERROR"
		reqLogger.LogResponse(int(codes.Internal), fmt.Errorf("service data not available"))
		s.metrics.LogCounter("grpc_requests_total", 1, map[string]string{
			"method": "HealthCheck",
			"status": "error",
		})
		return &v1.HealthCheckResponse{
			Status:    healthStatus,
			Timestamp: timestamppb.Now(),
		}, nil
	}

	reqLogger.LogResponse(int(codes.OK), nil)
	s.metrics.LogCounter("grpc_requests_total", 1, map[string]string{
		"method": "HealthCheck",
		"status": "ok",
	})

	return &v1.HealthCheckResponse{
		Status:    healthStatus,
		Timestamp: timestamppb.Now(),
	}, nil
}

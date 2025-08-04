package server

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v3"

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
	var sf model.ServicesFile
	if err := yaml.Unmarshal(yamlData, &sf); err != nil {
		return nil, fmt.Errorf("failed to parse services.yaml: %w", err)
	}

	// Create a local store with the parsed services
	store := &model.Store{}
	store.SetServices(sf.Services)
	catalogService := service.NewCatalogService(store)
	return &Server{svc: catalogService}, nil
}

// ListServices returns a list of all services
func (s *Server) ListServices(ctx context.Context, req *v1.ListServicesRequest) (*v1.ListServicesResponse, error) {
	return s.svc.ListServices(ctx, req)
}

// GetService returns a specific service by ID
func (s *Server) GetService(ctx context.Context, req *v1.GetServiceRequest) (*v1.GetServiceResponse, error) {
	return s.svc.GetService(ctx, req)
}

// GetServiceVersions returns all versions of a specific service
func (s *Server) GetServiceVersions(ctx context.Context, req *v1.GetServiceVersionsRequest) (*v1.GetServiceVersionsResponse, error) {
	return s.svc.GetServiceVersions(ctx, req)
}

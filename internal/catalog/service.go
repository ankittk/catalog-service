package catalog

import (
	"context"

	v1 "github.com/ankittk/catalog-service/proto/v1"
)

// Server implements the CatalogService server interface
type Server struct {
	// Embed the unimplemented server to ensure forward compatibility
	v1.UnimplementedCatalogServiceServer
}

// NewCatalogServer creates a new instance of the Catalog server
func NewCatalogServer() *Server {
	return &Server{}
}

func (s *Server) ListServices(ctx context.Context, request *v1.ListServicesRequest) (*v1.ListServicesResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Server) GetService(ctx context.Context, request *v1.GetServiceRequest) (*v1.GetServiceResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Server) GetServiceVersions(ctx context.Context, request *v1.GetServiceVersionsRequest) (*v1.GetServiceVersionsResponse, error) {
	//TODO implement me
	panic("implement me")
}

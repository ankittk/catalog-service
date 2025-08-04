package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ankittk/catalog-service/internal/logger"
	"github.com/ankittk/catalog-service/internal/model"
	v1 "github.com/ankittk/catalog-service/proto/v1"
)

var (
	ErrServiceNotFound = errors.New("service not found")
	ErrInvalidRequest  = errors.New("invalid request")
)

const (
	MaxPageSize     = 100
	DefaultPageSize = 10
)

var validSortFields = map[string]bool{
	"name":       true,
	"created_at": true,
	"updated_at": true,
}

var validSortOrders = map[string]bool{
	"asc":  true,
	"desc": true,
}

type CatalogService struct {
	data map[string]*model.Service
}

// NewCatalogService initializes a new CatalogService with the local store
func NewCatalogService(store *model.Store) *CatalogService {
	data := make(map[string]*model.Service)
	for _, s := range store.ListServices() {
		data[s.ID] = s
	}
	return &CatalogService{data: data}
}

// ListServices returns a paginated list of services based on the request parameters
func (c *CatalogService) ListServices(ctx context.Context, req *v1.ListServicesRequest) (*v1.ListServicesResponse, error) {
	logger.Get().Infow("ListServices called",
		"page_size", req.GetPageSize(),
		"page_token", req.GetPageToken(),
		"organization_id", req.GetOrganizationId(),
		"search_query", req.GetSearchQuery(),
		"sort_by", req.GetSortBy(),
		"sort_order", req.GetSortOrder())

	// validate request parameters
	if err := c.validateListServicesRequest(req); err != nil {
		return nil, err
	}

	// fetch all services from the store
	services := c.getAllServices()
	logger.Get().Debugw("Initial services count", "count", len(services))

	// filter services based on request parameters
	services = c.filterServices(services, req)
	logger.Get().Debugw("Services after filtering", "count", len(services))

	// sort results to ensure consistent ordering
	c.sortServices(services, req.GetSortBy(), req.GetSortOrder())

	// paginate results to handle large datasets
	pageSize := c.getPageSize(req.GetPageSize())
	startIndex, err := c.getStartIndex(req.GetPageToken(), pageSize, len(services))
	if err != nil {
		return nil, err
	}

	return c.paginateServices(services, startIndex, pageSize)
}

// GetService returns a specific service by ID
func (c *CatalogService) GetService(ctx context.Context, req *v1.GetServiceRequest) (*v1.GetServiceResponse, error) {
	logger.Get().Infow("GetService called", "service_id", req.GetId())

	// validate request parameters
	if err := c.validateGetServiceRequest(req); err != nil {
		return nil, err
	}

	// fetch service by ID
	svc, err := c.getServiceByID(req.GetId())
	if err != nil {
		return nil, err
	}

	logger.Get().Infow("GetService completed successfully", "service_id", req.GetId())
	return &v1.GetServiceResponse{Service: convertToProtoService(svc)}, nil
}

// GetServiceVersions returns all versions of a specific service
func (c *CatalogService) GetServiceVersions(ctx context.Context, req *v1.GetServiceVersionsRequest) (*v1.GetServiceVersionsResponse, error) {
	logger.Get().Infow("GetServiceVersions called", "service_id", req.GetServiceId())

	// validate request parameters
	if err := c.validateGetServiceVersionsRequest(req); err != nil {
		return nil, err
	}

	// get service by ID
	svc, err := c.getServiceByID(req.GetServiceId())
	if err != nil {
		return nil, err
	}

	versions := convertVersionsToProto(svc.Versions)

	logger.Get().Infow("GetServiceVersions completed successfully",
		"service_id", req.GetServiceId(),
		"versions_count", len(versions))

	return &v1.GetServiceVersionsResponse{Versions: versions}, nil
}

// validateListServicesRequest checks the validity of the ListServicesRequest parameters
func (c *CatalogService) validateListServicesRequest(req *v1.ListServicesRequest) error {
	if req.GetPageSize() < 0 || req.GetPageSize() > MaxPageSize {
		return status.Errorf(codes.InvalidArgument, "%v: page_size must be between 0 and %d, got %d", ErrInvalidRequest, MaxPageSize, req.GetPageSize())
	}
	return nil
}

// validateGetServiceRequest checks the validity of the GetServiceRequest parameters
func (c *CatalogService) validateGetServiceRequest(req *v1.GetServiceRequest) error {
	if req.GetId() == "" {
		return status.Errorf(codes.InvalidArgument, "%v: service ID is required", ErrInvalidRequest)
	}
	return nil
}

// validateGetServiceVersionsRequest checks the validity of the GetServiceVersionsRequest parameters
func (c *CatalogService) validateGetServiceVersionsRequest(req *v1.GetServiceVersionsRequest) error {
	if req.GetServiceId() == "" {
		return status.Errorf(codes.InvalidArgument, "%v: service ID is required", ErrInvalidRequest)
	}
	return nil
}

// getAllServices retrieves all services from the local data store
func (c *CatalogService) getAllServices() []*model.Service {
	services := make([]*model.Service, 0, len(c.data))
	for _, s := range c.data {
		services = append(services, s)
	}
	return services
}

// getPageSize returns the requested page size, defaulting to DefaultPageSize if not specified
func (c *CatalogService) getPageSize(requestedPageSize int32) int32 {
	if requestedPageSize == 0 {
		return DefaultPageSize
	}
	return requestedPageSize
}

// getStartIndex calculates the starting index for pagination based on the page token and page size
func (c *CatalogService) getStartIndex(pageToken string, pageSize int32, totalCount int) (int32, error) {
	if pageToken == "" {
		return 0, nil
	}

	// parse page token - format: "page_<offset>"
	if !strings.HasPrefix(pageToken, "page_") {
		return 0, status.Errorf(codes.InvalidArgument, "%v: invalid page token format", ErrInvalidRequest)
	}

	offsetStr := strings.TrimPrefix(pageToken, "page_")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return 0, status.Errorf(codes.InvalidArgument, "%v: invalid page token: %v", ErrInvalidRequest, err)
	}

	// validate offset is within bounds
	if offset < 0 || offset >= totalCount {
		return 0, status.Errorf(codes.InvalidArgument, "%v: page token out of range", ErrInvalidRequest)
	}

	return int32(offset), nil
}

// paginateServices slices the services based on the start index and page size
func (c *CatalogService) paginateServices(services []*model.Service, startIndex, pageSize int32) (*v1.ListServicesResponse, error) {
	totalCount := len(services)

	if startIndex >= int32(totalCount) {
		logger.Get().Infow("ListServices completed - no results for page",
			"start_index", startIndex,
			"total_count", totalCount)
		return &v1.ListServicesResponse{
			Services:      []*v1.Service{},
			NextPageToken: "",
			TotalCount:    int32(totalCount),
		}, nil
	}

	endIndex := startIndex + pageSize
	if endIndex > int32(totalCount) {
		endIndex = int32(totalCount)
	}

	// convert to proto and return
	protoServices := make([]*v1.Service, 0, endIndex-startIndex)
	for _, s := range services[startIndex:endIndex] {
		protoServices = append(protoServices, convertToProtoService(s))
	}

	// generate next page token
	var nextPageToken string
	if endIndex < int32(totalCount) {
		nextPageToken = fmt.Sprintf("page_%d", endIndex)
	}

	logger.Get().Infow("ListServices completed successfully",
		"returned_count", len(protoServices),
		"total_count", totalCount,
		"has_next_page", nextPageToken != "",
		"start_index", startIndex,
		"end_index", endIndex)

	return &v1.ListServicesResponse{
		Services:      protoServices,
		NextPageToken: nextPageToken,
		TotalCount:    int32(totalCount),
	}, nil
}

// filterServices filters the services based on organization ID and search query
func (c *CatalogService) filterServices(services []*model.Service, req *v1.ListServicesRequest) []*model.Service {
	var filtered []*model.Service

	for _, s := range services {
		// filter by organization ID if specified
		if req.GetOrganizationId() != "" && s.OrganizationID != req.GetOrganizationId() {
			continue
		}

		// filter by search query if specified
		if req.GetSearchQuery() != "" {
			query := strings.ToLower(req.GetSearchQuery())
			name := strings.ToLower(s.Name)
			description := strings.ToLower(s.Description)

			if !strings.Contains(name, query) && !strings.Contains(description, query) {
				continue
			}
		}

		filtered = append(filtered, s)
	}

	return filtered
}

// sortServices sorts the services based on the specified field and order
func (c *CatalogService) sortServices(services []*model.Service, sortBy, sortOrder string) {
	// Set defaults
	if sortBy == "" {
		sortBy = "name"
	}
	if sortOrder == "" {
		sortOrder = "asc"
	}

	// validate sort fields
	if !validSortFields[sortBy] {
		sortBy = "name"
	}

	// validate sort order
	if !validSortOrders[sortOrder] {
		sortOrder = "asc"
	}

	sort.Slice(services, func(i, j int) bool {
		var result bool

		switch sortBy {
		case "name":
			result = services[i].Name < services[j].Name
		case "created_at":
			result = services[i].CreatedAt.Before(services[j].CreatedAt)
		case "updated_at":
			result = services[i].UpdatedAt.Before(services[j].UpdatedAt)
		default:
			result = services[i].Name < services[j].Name
		}

		if sortOrder == "desc" {
			result = !result
		}

		return result
	})
}

// getServiceByID retrieves a service by its ID, returning an error if not found
func (c *CatalogService) getServiceByID(id string) (*model.Service, error) {
	svc, ok := c.data[id]
	if !ok {
		logger.Get().Warnw("Service not found", "service_id", id)
		return nil, status.Errorf(codes.NotFound, "%v: service with ID '%s' not found", ErrServiceNotFound, id)
	}
	return svc, nil
}

// convertVersionsToProto converts a slice of ServiceVersion models to a slice of ServiceVersion protobuf messages
func convertVersionsToProto(versions []*model.ServiceVersion) []*v1.ServiceVersion {
	protoVersions := make([]*v1.ServiceVersion, 0, len(versions))
	for _, v := range versions {
		protoVersions = append(protoVersions, &v1.ServiceVersion{
			Id:          v.ID,
			Version:     v.Version,
			ServiceId:   v.ServiceID,
			Description: v.Description,
			IsActive:    v.IsActive,
			CreatedAt:   timestamppb.New(v.CreatedAt),
			UpdatedAt:   timestamppb.New(v.UpdatedAt),
		})
	}
	return protoVersions
}

// convertToProtoService converts a Service model to a Service protobuf message
func convertToProtoService(s *model.Service) *v1.Service {
	return &v1.Service{
		Id:             s.ID,
		Name:           s.Name,
		Description:    s.Description,
		OrganizationId: s.OrganizationID,
		Url:            s.URL,
		CreatedAt:      timestamppb.New(s.CreatedAt),
		UpdatedAt:      timestamppb.New(s.UpdatedAt),
		Versions:       convertVersionsToProto(s.Versions),
	}
}

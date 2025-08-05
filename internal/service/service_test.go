package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ankittk/catalog-service/internal/model"
	v1 "github.com/ankittk/catalog-service/proto/v1"
)

func mockTestData() map[string]*model.Service {
	// Parse the actual timestamps from services.yaml
	createdAt1, _ := time.Parse(time.RFC3339, "2024-05-01T10:00:00Z")
	updatedAt1, _ := time.Parse(time.RFC3339, "2025-08-01T09:00:00Z")
	createdAt2, _ := time.Parse(time.RFC3339, "2023-12-15T08:00:00Z")
	updatedAt2, _ := time.Parse(time.RFC3339, "2025-08-01T08:00:00Z")
	createdAt3, _ := time.Parse(time.RFC3339, "2022-11-01T12:00:00Z")
	updatedAt3, _ := time.Parse(time.RFC3339, "2025-07-31T12:00:00Z")
	createdAt4, _ := time.Parse(time.RFC3339, "2024-01-10T14:00:00Z")
	updatedAt4, _ := time.Parse(time.RFC3339, "2025-07-01T14:00:00Z")

	service1 := &model.Service{
		ID:             "svc-1",
		Name:           "User Service",
		Description:    "Handles user authentication and profile management",
		OrganizationID: "org-1",
		URL:            "https://services.example.com/user",
		CreatedAt:      createdAt1,
		UpdatedAt:      updatedAt1,
		Versions: []*model.ServiceVersion{
			{
				ID:          "v1",
				Version:     "v1.0.0",
				ServiceID:   "svc-1",
				Description: "Initial stable release",
				IsActive:    false,
				CreatedAt:   time.Date(2024, 5, 1, 10, 0, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC),
			},
			{
				ID:          "v2",
				Version:     "v1.1.0",
				ServiceID:   "svc-1",
				Description: "Added OAuth support",
				IsActive:    true,
				CreatedAt:   time.Date(2024, 7, 1, 10, 0, 0, 0, time.UTC),
				UpdatedAt:   updatedAt1,
			},
		},
	}

	service2 := &model.Service{
		ID:             "svc-2",
		Name:           "Payment Gateway",
		Description:    "Facilitates payments and transaction management",
		OrganizationID: "org-2",
		URL:            "https://services.example.com/payment",
		CreatedAt:      createdAt2,
		UpdatedAt:      updatedAt2,
		Versions: []*model.ServiceVersion{
			{
				ID:          "v1",
				Version:     "v2.0.0",
				ServiceID:   "svc-2",
				Description: "Supports Stripe and Razorpay",
				IsActive:    true,
				CreatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	service3 := &model.Service{
		ID:             "svc-3",
		Name:           "Inventory Service",
		Description:    "Tracks product availability and stock levels",
		OrganizationID: "org-1",
		URL:            "https://services.example.com/inventory",
		CreatedAt:      createdAt3,
		UpdatedAt:      updatedAt3,
		Versions: []*model.ServiceVersion{
			{
				ID:          "v1",
				Version:     "v1.0.0",
				ServiceID:   "svc-3",
				Description: "Initial version",
				IsActive:    false,
				CreatedAt:   createdAt3,
				UpdatedAt:   time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			{
				ID:          "v2",
				Version:     "v2.0.0",
				ServiceID:   "svc-3",
				Description: "Optimized warehouse sync",
				IsActive:    true,
				CreatedAt:   time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC),
				UpdatedAt:   updatedAt3,
			},
		},
	}

	service4 := &model.Service{
		ID:             "svc-4",
		Name:           "Analytics Service",
		Description:    "Generates usage and engagement reports",
		OrganizationID: "org-3",
		URL:            "https://services.example.com/analytics",
		CreatedAt:      createdAt4,
		UpdatedAt:      updatedAt4,
		Versions: []*model.ServiceVersion{
			{
				ID:          "v1",
				Version:     "v0.1.0",
				ServiceID:   "svc-4",
				Description: "Beta release",
				IsActive:    false,
				CreatedAt:   createdAt4,
				UpdatedAt:   time.Date(2024, 3, 10, 14, 0, 0, 0, time.UTC),
			},
			{
				ID:          "v2",
				Version:     "v1.0.0",
				ServiceID:   "svc-4",
				Description: "First stable release",
				IsActive:    true,
				CreatedAt:   time.Date(2024, 6, 1, 14, 0, 0, 0, time.UTC),
				UpdatedAt:   updatedAt4,
			},
		},
	}

	return map[string]*model.Service{
		"svc-1": service1,
		"svc-2": service2,
		"svc-3": service3,
		"svc-4": service4,
	}
}

func TestCatalogService_ListServices(t *testing.T) {
	testData := mockTestData()
	svc := &CatalogService{data: testData}
	ctx := context.Background()

	tests := []struct {
		name    string
		req     *v1.ListServicesRequest
		want    *v1.ListServicesResponse
		wantErr bool
	}{
		{
			name: "list all services with default pagination",
			req:  &v1.ListServicesRequest{},
			want: &v1.ListServicesResponse{
				Services:      []*v1.Service{},
				NextPageToken: "",
				TotalCount:    4,
			},
			wantErr: false,
		},
		{
			name: "list services with custom page size",
			req: &v1.ListServicesRequest{
				PageSize: 2,
			},
			want: &v1.ListServicesResponse{
				Services:      []*v1.Service{},
				NextPageToken: "page_2",
				TotalCount:    4,
			},
			wantErr: false,
		},
		{
			name: "list services with invalid page size",
			req: &v1.ListServicesRequest{
				PageSize: 150,
			},
			wantErr: true,
		},
		{
			name: "list services with negative page size",
			req: &v1.ListServicesRequest{
				PageSize: -1,
			},
			wantErr: true,
		},
		{
			name:    "list services with nil request",
			req:     nil,
			wantErr: true,
		},
		{
			name: "list services with long search query",
			req: &v1.ListServicesRequest{
				SearchQuery: strings.Repeat("a", 101),
			},
			wantErr: true,
		},
		{
			name: "list services with invalid organization ID",
			req: &v1.ListServicesRequest{
				OrganizationId: "invalid@org",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.ListServices(ctx, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.req != nil && tt.req.PageSize > MaxPageSize {
					assert.Contains(t, err.Error(), "page_size must be between 0 and 100")
				} else if tt.req == nil {
					assert.Contains(t, err.Error(), "request cannot be nil")
				} else if tt.req.SearchQuery != "" && len(tt.req.SearchQuery) > 100 {
					assert.Contains(t, err.Error(), "search_query too long")
				} else if tt.req.OrganizationId != "" && strings.Contains(tt.req.OrganizationId, "@") {
					assert.Contains(t, err.Error(), "invalid organization_id format")
				}
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, tt.want.TotalCount, got.TotalCount)
			if tt.want.NextPageToken != "" {
				assert.Equal(t, tt.want.NextPageToken, got.NextPageToken)
			}
		})
	}
}

func TestCatalogService_GetService(t *testing.T) {
	testData := mockTestData()
	svc := &CatalogService{data: testData}
	ctx := context.Background()

	tests := []struct {
		name    string
		req     *v1.GetServiceRequest
		want    *v1.GetServiceResponse
		wantErr bool
	}{
		{
			name: "get existing service",
			req: &v1.GetServiceRequest{
				Id: "svc-1",
			},
			want: &v1.GetServiceResponse{
				Service: &v1.Service{
					Id:             "svc-1",
					Name:           "User Service",
					Description:    "Handles user authentication and profile management",
					OrganizationId: "org-1",
					Url:            "https://services.example.com/user",
				},
			},
			wantErr: false,
		},
		{
			name: "get non-existing service",
			req: &v1.GetServiceRequest{
				Id: "non-existent",
			},
			wantErr: true,
		},
		{
			name: "get service with empty ID",
			req: &v1.GetServiceRequest{
				Id: "",
			},
			wantErr: true,
		},
		{
			name: "get service with invalid ID format",
			req: &v1.GetServiceRequest{
				Id: "invalid@id",
			},
			wantErr: true,
		},
		{
			name:    "get service with nil request",
			req:     nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.GetService(ctx, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.req == nil {
					assert.Contains(t, err.Error(), "request cannot be nil")
				} else if tt.req.Id == "" {
					assert.Contains(t, err.Error(), "service ID is required")
				} else if tt.req.Id == "non-existent" {
					assert.Contains(t, err.Error(), "service not found")
				} else if strings.Contains(tt.req.Id, "@") {
					assert.Contains(t, err.Error(), "invalid service ID format")
				}
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, tt.want.Service.Id, got.Service.Id)
			assert.Equal(t, tt.want.Service.Name, got.Service.Name)
			assert.Equal(t, tt.want.Service.Description, got.Service.Description)
		})
	}
}

func TestCatalogService_GetServiceVersions(t *testing.T) {
	testData := mockTestData()
	svc := &CatalogService{data: testData}
	ctx := context.Background()

	tests := []struct {
		name    string
		req     *v1.GetServiceVersionsRequest
		want    *v1.GetServiceVersionsResponse
		wantErr bool
	}{
		{
			name: "get versions for existing service",
			req: &v1.GetServiceVersionsRequest{
				ServiceId: "svc-1",
			},
			want: &v1.GetServiceVersionsResponse{
				Versions: []*v1.ServiceVersion{},
			},
			wantErr: false,
		},
		{
			name: "get versions for non-existing service",
			req: &v1.GetServiceVersionsRequest{
				ServiceId: "non-existent",
			},
			wantErr: true,
		},
		{
			name: "get versions with empty service ID",
			req: &v1.GetServiceVersionsRequest{
				ServiceId: "",
			},
			wantErr: true,
		},
		{
			name: "get versions with invalid service ID format",
			req: &v1.GetServiceVersionsRequest{
				ServiceId: "invalid@id",
			},
			wantErr: true,
		},
		{
			name:    "get versions with nil request",
			req:     nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.GetServiceVersions(ctx, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.req == nil {
					assert.Contains(t, err.Error(), "request cannot be nil")
				} else if tt.req.ServiceId == "" {
					assert.Contains(t, err.Error(), "service ID is required")
				} else if tt.req.ServiceId == "non-existent" {
					assert.Contains(t, err.Error(), "service not found")
				} else if strings.Contains(tt.req.ServiceId, "@") {
					assert.Contains(t, err.Error(), "invalid service ID format")
				}
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, got)
			if tt.req.ServiceId == "svc-1" {
				assert.Len(t, got.Versions, 2)
			}
		})
	}
}

func TestCatalogService_validateListServicesRequest(t *testing.T) {
	svc := &CatalogService{}

	tests := []struct {
		name    string
		req     *v1.ListServicesRequest
		wantErr bool
	}{
		{
			name:    "valid request",
			req:     &v1.ListServicesRequest{PageSize: 10},
			wantErr: false,
		},
		{
			name:    "page size too large",
			req:     &v1.ListServicesRequest{PageSize: 150},
			wantErr: true,
		},
		{
			name:    "negative page size",
			req:     &v1.ListServicesRequest{PageSize: -1},
			wantErr: true,
		},
		{
			name:    "zero page size",
			req:     &v1.ListServicesRequest{PageSize: 0},
			wantErr: false,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
		{
			name:    "long search query",
			req:     &v1.ListServicesRequest{SearchQuery: strings.Repeat("a", 101)},
			wantErr: true,
		},
		{
			name:    "invalid organization ID",
			req:     &v1.ListServicesRequest{OrganizationId: "invalid@org"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.validateListServicesRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.req != nil && tt.req.PageSize > MaxPageSize {
					assert.Contains(t, err.Error(), "page_size must be between 0 and 100")
				} else if tt.req == nil {
					assert.Contains(t, err.Error(), "request cannot be nil")
				} else if tt.req.SearchQuery != "" && len(tt.req.SearchQuery) > 100 {
					assert.Contains(t, err.Error(), "search_query too long")
				} else if tt.req.OrganizationId != "" && strings.Contains(tt.req.OrganizationId, "@") {
					assert.Contains(t, err.Error(), "invalid organization_id format")
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCatalogService_validateGetServiceRequest(t *testing.T) {
	svc := &CatalogService{}

	tests := []struct {
		name    string
		req     *v1.GetServiceRequest
		wantErr bool
	}{
		{
			name:    "valid request",
			req:     &v1.GetServiceRequest{Id: "svc-1"},
			wantErr: false,
		},
		{
			name:    "empty ID",
			req:     &v1.GetServiceRequest{Id: ""},
			wantErr: true,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
		{
			name:    "invalid ID format",
			req:     &v1.GetServiceRequest{Id: "invalid@id"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.validateGetServiceRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.req == nil {
					assert.Contains(t, err.Error(), "request cannot be nil")
				} else if tt.req.Id == "" {
					assert.Contains(t, err.Error(), "service ID is required")
				} else if strings.Contains(tt.req.Id, "@") {
					assert.Contains(t, err.Error(), "invalid service ID format")
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCatalogService_validateGetServiceVersionsRequest(t *testing.T) {
	svc := &CatalogService{}

	tests := []struct {
		name    string
		req     *v1.GetServiceVersionsRequest
		wantErr bool
	}{
		{
			name:    "valid request",
			req:     &v1.GetServiceVersionsRequest{ServiceId: "svc-1"},
			wantErr: false,
		},
		{
			name:    "empty service ID",
			req:     &v1.GetServiceVersionsRequest{ServiceId: ""},
			wantErr: true,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
		{
			name:    "invalid service ID format",
			req:     &v1.GetServiceVersionsRequest{ServiceId: "invalid@id"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.validateGetServiceVersionsRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.req == nil {
					assert.Contains(t, err.Error(), "request cannot be nil")
				} else if tt.req.ServiceId == "" {
					assert.Contains(t, err.Error(), "service ID is required")
				} else if strings.Contains(tt.req.ServiceId, "@") {
					assert.Contains(t, err.Error(), "invalid service ID format")
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCatalogService_isValidID(t *testing.T) {
	svc := &CatalogService{}

	tests := []struct {
		name string
		id   string
		want bool
	}{
		{"valid ID", "svc-1", true},
		{"valid ID with underscore", "svc_1", true},
		{"valid ID uppercase", "SVC-1", true},
		{"valid ID numbers", "svc123", true},
		{"empty ID", "", false},
		{"ID with special chars", "svc@1", false},
		{"ID with spaces", "svc 1", false},
		{"ID too long", strings.Repeat("a", 51), false},
		{"ID with dots", "svc.1", false},
		{"ID with slashes", "svc/1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.isValidID(tt.id)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCatalogService_getStartIndex(t *testing.T) {
	svc := &CatalogService{}

	tests := []struct {
		name       string
		pageToken  string
		pageSize   int32
		totalCount int
		want       int32
		wantErr    bool
	}{
		{
			name:       "empty page token returns 0",
			pageToken:  "",
			pageSize:   10,
			totalCount: 100,
			want:       0,
			wantErr:    false,
		},
		{
			name:       "valid page token",
			pageToken:  "page_10",
			pageSize:   10,
			totalCount: 100,
			want:       10,
			wantErr:    false,
		},
		{
			name:       "invalid page token format",
			pageToken:  "invalid_token",
			pageSize:   10,
			totalCount: 100,
			want:       0,
			wantErr:    true,
		},
		{
			name:       "page token out of range",
			pageToken:  "page_150",
			pageSize:   10,
			totalCount: 100,
			want:       0,
			wantErr:    true,
		},
		{
			name:       "negative page token",
			pageToken:  "page_-10",
			pageSize:   10,
			totalCount: 100,
			want:       0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.getStartIndex(tt.pageToken, tt.pageSize, tt.totalCount)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.pageToken == "invalid_token" {
					assert.Contains(t, err.Error(), "invalid page token format")
				} else if strings.Contains(tt.pageToken, "page_") {
					assert.Contains(t, err.Error(), "page token out of range")
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestCatalogService_paginateServices(t *testing.T) {
	testData := mockTestData()
	services := make([]*model.Service, 0, len(testData))
	for _, s := range testData {
		services = append(services, s)
	}
	svc := &CatalogService{}

	tests := []struct {
		name       string
		services   []*model.Service
		startIndex int32
		pageSize   int32
		want       *v1.ListServicesResponse
		wantErr    bool
	}{
		{
			name:       "first page",
			services:   services,
			startIndex: 0,
			pageSize:   2,
			want: &v1.ListServicesResponse{
				Services:      []*v1.Service{},
				NextPageToken: "page_2",
				TotalCount:    4,
			},
			wantErr: false,
		},
		{
			name:       "last page",
			services:   services,
			startIndex: 2,
			pageSize:   2,
			want: &v1.ListServicesResponse{
				Services:      []*v1.Service{},
				NextPageToken: "",
				TotalCount:    4,
			},
			wantErr: false,
		},
		{
			name:       "start index beyond total count",
			services:   services,
			startIndex: 10,
			pageSize:   2,
			want: &v1.ListServicesResponse{
				Services:      []*v1.Service{},
				NextPageToken: "",
				TotalCount:    4,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.paginateServices(tt.services, tt.startIndex, tt.pageSize)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, tt.want.TotalCount, got.TotalCount)
			assert.Equal(t, tt.want.NextPageToken, got.NextPageToken)
		})
	}
}

func TestCatalogService_filterServices(t *testing.T) {
	testData := mockTestData()
	services := make([]*model.Service, 0, len(testData))
	for _, s := range testData {
		services = append(services, s)
	}
	svc := &CatalogService{}

	tests := []struct {
		name     string
		services []*model.Service
		req      *v1.ListServicesRequest
		want     []*model.Service
	}{
		{
			name:     "no filters",
			services: services,
			req:      &v1.ListServicesRequest{},
			want:     services,
		},
		{
			name:     "filter by organization",
			services: services,
			req: &v1.ListServicesRequest{
				OrganizationId: "org-1",
			},
			want: []*model.Service{testData["svc-1"], testData["svc-3"]},
		},
		{
			name:     "filter by search query in name",
			services: services,
			req: &v1.ListServicesRequest{
				SearchQuery: "User",
			},
			want: []*model.Service{testData["svc-1"]},
		},
		{
			name:     "filter by search query in description",
			services: services,
			req: &v1.ListServicesRequest{
				SearchQuery: "authentication",
			},
			want: []*model.Service{testData["svc-1"]},
		},
		{
			name:     "filter by organization and search query",
			services: services,
			req: &v1.ListServicesRequest{
				OrganizationId: "org-1",
				SearchQuery:    "Service",
			},
			want: []*model.Service{testData["svc-1"], testData["svc-3"]},
		},
		{
			name:     "filter with trimmed search query",
			services: services,
			req: &v1.ListServicesRequest{
				SearchQuery: "  User  ",
			},
			want: []*model.Service{testData["svc-1"]},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.filterServices(tt.services, tt.req)
			assert.Len(t, got, len(tt.want))

			// Create maps for easier comparison regardless of order
			expectedIDs := make(map[string]bool)
			gotIDs := make(map[string]bool)

			for _, expected := range tt.want {
				expectedIDs[expected.ID] = true
			}

			for _, service := range got {
				gotIDs[service.ID] = true
			}

			for expectedID := range expectedIDs {
				assert.True(t, gotIDs[expectedID], "Expected service %s not found in results", expectedID)
			}

			for gotID := range gotIDs {
				assert.True(t, expectedIDs[gotID], "Unexpected service %s found in results", gotID)
			}
		})
	}
}

func TestCatalogService_sortServices(t *testing.T) {
	testData := mockTestData()
	services := make([]*model.Service, 0, len(testData))
	for _, s := range testData {
		services = append(services, s)
	}
	svc := &CatalogService{}

	tests := []struct {
		name      string
		services  []*model.Service
		sortBy    string
		sortOrder string
	}{
		{
			name:      "sort by name ascending",
			services:  services,
			sortBy:    "name",
			sortOrder: "asc",
		},
		{
			name:      "sort by name descending",
			services:  services,
			sortBy:    "name",
			sortOrder: "desc",
		},
		{
			name:      "sort by created_at ascending",
			services:  services,
			sortBy:    "created_at",
			sortOrder: "asc",
		},
		{
			name:      "sort by updated_at descending",
			services:  services,
			sortBy:    "updated_at",
			sortOrder: "desc",
		},
		{
			name:      "invalid sort field defaults to name",
			services:  services,
			sortBy:    "invalid_field",
			sortOrder: "asc",
		},
		{
			name:      "invalid sort order defaults to asc",
			services:  services,
			sortBy:    "name",
			sortOrder: "invalid_order",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			servicesCopy := make([]*model.Service, len(tt.services))
			copy(servicesCopy, tt.services)

			svc.sortServices(servicesCopy, tt.sortBy, tt.sortOrder)

			assert.Len(t, servicesCopy, len(tt.services))

			if tt.sortBy == "name" || tt.sortBy == "invalid_field" {
				if tt.sortOrder == "asc" || tt.sortOrder == "invalid_order" {
					assert.True(t, servicesCopy[0].Name <= servicesCopy[1].Name)
					assert.True(t, servicesCopy[1].Name <= servicesCopy[2].Name)
				} else if tt.sortOrder == "desc" {
					assert.True(t, servicesCopy[0].Name >= servicesCopy[1].Name)
					assert.True(t, servicesCopy[1].Name >= servicesCopy[2].Name)
				}
			}
		})
	}
}

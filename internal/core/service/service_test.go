package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ankittk/catalog-service/internal/core/model"
	v1 "github.com/ankittk/catalog-service/proto/v1"
)

func mockTestData() map[string]*model.Service {
	now := time.Now()
	service1 := &model.Service{
		ID:             "svc-1",
		Name:           "User Service",
		Description:    "Handles user authentication and profile management",
		OrganizationID: "org-1",
		URL:            "https://services.example.com/user",
		CreatedAt:      now.Add(-24 * time.Hour),
		UpdatedAt:      now,
		Versions: []*model.ServiceVersion{
			{
				ID:          "v1",
				Version:     "v1.0.0",
				ServiceID:   "svc-1",
				Description: "Initial stable release",
				IsActive:    false,
				CreatedAt:   now.Add(-48 * time.Hour),
				UpdatedAt:   now.Add(-24 * time.Hour),
			},
			{
				ID:          "v2",
				Version:     "v1.1.0",
				ServiceID:   "svc-1",
				Description: "Added OAuth support",
				IsActive:    true,
				CreatedAt:   now.Add(-12 * time.Hour),
				UpdatedAt:   now,
			},
		},
	}

	service2 := &model.Service{
		ID:             "svc-2",
		Name:           "Payment Gateway",
		Description:    "Facilitates payments and transaction management",
		OrganizationID: "org-2",
		URL:            "https://services.example.com/payment",
		CreatedAt:      now.Add(-12 * time.Hour),
		UpdatedAt:      now.Add(-6 * time.Hour),
		Versions: []*model.ServiceVersion{
			{
				ID:          "v1",
				Version:     "v2.0.0",
				ServiceID:   "svc-2",
				Description: "Supports Stripe and Razorpay",
				IsActive:    true,
				CreatedAt:   now.Add(-12 * time.Hour),
				UpdatedAt:   now.Add(-6 * time.Hour),
			},
		},
	}

	service3 := &model.Service{
		ID:             "svc-3",
		Name:           "Analytics Service",
		Description:    "Generates usage and engagement reports",
		OrganizationID: "org-1",
		URL:            "https://services.example.com/analytics",
		CreatedAt:      now.Add(-6 * time.Hour),
		UpdatedAt:      now.Add(-1 * time.Hour),
		Versions: []*model.ServiceVersion{
			{
				ID:          "v1",
				Version:     "v0.1.0",
				ServiceID:   "svc-3",
				Description: "Beta release",
				IsActive:    false,
				CreatedAt:   now.Add(-6 * time.Hour),
				UpdatedAt:   now.Add(-3 * time.Hour),
			},
		},
	}

	return map[string]*model.Service{
		"svc-1": service1,
		"svc-2": service2,
		"svc-3": service3,
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
				TotalCount:    3,
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
				TotalCount:    3,
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.ListServices(ctx, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.req.PageSize > MaxPageSize {
					assert.Contains(t, err.Error(), "page_size must be between 0 and 100")
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.GetService(ctx, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.req.Id == "" {
					assert.Contains(t, err.Error(), "service ID is required")
				} else if tt.req.Id == "non-existent" {
					assert.Contains(t, err.Error(), "service not found")
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.GetServiceVersions(ctx, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.req.ServiceId == "" {
					assert.Contains(t, err.Error(), "service ID is required")
				} else if tt.req.ServiceId == "non-existent" {
					assert.Contains(t, err.Error(), "service not found")
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.validateListServicesRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "page_size must be between 0 and 100")
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.validateGetServiceRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "service ID is required")
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.validateGetServiceVersionsRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "service ID is required")
			} else {
				assert.NoError(t, err)
			}
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
				TotalCount:    3,
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
				TotalCount:    3,
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
				TotalCount:    3,
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.filterServices(tt.services, tt.req)
			assert.Len(t, got, len(tt.want))
			for i, expected := range tt.want {
				assert.Equal(t, expected.ID, got[i].ID)
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

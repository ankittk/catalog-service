package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestCatalogService_InvalidYAMLData(t *testing.T) {
	invalidYAML := []byte(`
services:
  - id: "svc-1"
    name: "User Service"
    description: "Handles user authentication and profile management"
    organization_id: "org-1"
    url: "https://services.example.com/user"
    created_at: "invalid-date"
    updated_at: "2025-08-01T09:00:00Z"
`)

	var sf ServicesFile
	err := yaml.Unmarshal(invalidYAML, &sf)
	// This should error because the invalid date cannot be parsed
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot parse")
}

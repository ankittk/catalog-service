package model

import (
	"time"
)

// Service represents a service in the catalog.
type Service struct {
	ID             string            `yaml:"id"`
	Name           string            `yaml:"name"`
	Description    string            `yaml:"description"`
	OrganizationID string            `yaml:"organization_id"`
	URL            string            `yaml:"url"`
	CreatedAt      time.Time         `yaml:"created_at"`
	UpdatedAt      time.Time         `yaml:"updated_at"`
	Versions       []*ServiceVersion `yaml:"versions"`
}

// ServiceVersion represents a version of a service.
type ServiceVersion struct {
	ID          string    `yaml:"id"`
	Version     string    `yaml:"version"`
	ServiceID   string    `yaml:"service_id"`
	Description string    `yaml:"description"`
	IsActive    bool      `yaml:"is_active"`
	CreatedAt   time.Time `yaml:"created_at"`
	UpdatedAt   time.Time `yaml:"updated_at"`
}

// ServicesFile represents the structure of the services YAML file.
type ServicesFile struct {
	Services []*Service `yaml:"services"`
}

// Store is a simple in-memory store for services.
type Store struct {
	services []*Service
}

// ListServices returns a list of all services in the store.
func (s *Store) ListServices() []*Service {
	return s.services
}

// SetServices sets the services in the store
func (s *Store) SetServices(services []*Service) {
	s.services = services
}

PROTO_DIR=proto
CMD_MAIN=./cmd/server/main.go

.PHONY: generate run build clean test

# Generate protobuf files
generate:
	cd $(PROTO_DIR) && buf generate

# Build the application
build:
	go build -o bin/catalog-service $(CMD_MAIN)

# Run the application
run:
	go run $(CMD_MAIN)

# Clean build artifacts
clean:
	rm -rf bin/

# Run tests
test:
	go test ./...

# Install dependencies
deps:
	go mod download
	go mod tidy

# Health check
health:
	curl -s http://localhost:8000/health | jq .

# List all services
list-services:
	curl -s "http://localhost:8000/v1/services" | jq .

# List services with pagination
list-services-paginated:
	curl -s "http://localhost:8000/v1/services?page_size=2&page_token=" | jq .

# List services with organization filter
list-services-by-org:
	curl -s "http://localhost:8000/v1/services?organization_id=org-1" | jq .

# List services with search query
list-services-search:
	curl -s "http://localhost:8000/v1/services?search_query=user" | jq .

# List services with sorting
list-services-sorted:
	curl -s "http://localhost:8000/v1/services?sort_by=name&sort_order=asc" | jq .

# Get specific service
get-service:
	curl -s "http://localhost:8000/v1/services/svc-1" | jq .

# Get service versions
get-service-versions:
	curl -s "http://localhost:8000/v1/services/svc-1/versions" | jq .

# Get all available services
get-all-services:
	@echo "=== Health Check ==="
	@make health
	@echo -e "\n=== List All Services ==="
	@make list-services
	@echo -e "\n=== Get User Service ==="
	@make get-service
	@echo -e "\n=== Get User Service Versions ==="
	@make get-service-versions
	@echo -e "\n=== List Services by Organization ==="
	@make list-services-by-org
	@echo -e "\n=== Search Services ==="
	@make list-services-search
	@echo -e "\n=== Sorted Services ==="
	@make list-services-sorted

# Start server and run all tests
start-and-test:
	@echo "Starting server..."
	@make run &
	@sleep 3
	@echo "Running all API tests..."
	@make get-all-services
	@echo "Stopping server..."
	@pkill -f "catalog-service" || true

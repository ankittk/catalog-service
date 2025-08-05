PROTO_DIR=proto
CMD_MAIN=./cmd/server/main.go
TEST_API_SCRIPT=./script/test_api.sh

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
	go test -v ./...

# Install dependencies
deps:
	go mod download
	go mod tidy

# Health check
health:
	curl -s http://localhost:8000/health | jq .

# test api
test-api:
	bash $(TEST_API_SCRIPT)

# Serve protobuf-generated swagger (more complete API docs)
swagger:
	npx redoc-cli serve ./docs/v1/catalog.swagger.json --port 8080

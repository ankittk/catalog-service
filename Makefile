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
	ENABLE_AUTH=true JWT_SECRET_KEY=my-secret-key go run $(CMD_MAIN)

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

# Health check without authentication
health:
	curl -s http://localhost:8000/health | jq .

# Test API with a custom script
test-api:
	@echo "Generating JWT token..."
	@TOKEN=$$(curl -s -X POST http://localhost:8000/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"admin@org1.com","password":"admin123","organization":"org-1"}' \
		| jq -r '.token'); \
	if [ -z "$$TOKEN" ]; then \
		echo "Error: Failed to generate JWT token. Ensure that the server is running with ENABLE_AUTH=true and JWT_SECRET_KEY set."; \
		exit 1; \
	else \
		echo "Testing API with JWT token..."; \
		bash $(TEST_API_SCRIPT) "$$TOKEN"; \
	fi

# Serve protobuf-generated swagger (more complete API docs)
swagger:
	npx redoc-cli serve ./docs/v1/catalog.swagger.json --port 8080

# Generate JWT token by calling the login endpoint
jwt-token:
	curl -s -X POST http://localhost:8000/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"admin@org1.com","password":"admin123","organization":"org-1"}' \
		| jq -r '.token'

# Docker commands
docker-build:
	docker build -t catalog-service .

docker-run:
	docker run -p 8000:8000 -p 9000:9000 \
		-e ENABLE_AUTH=true \
		-e JWT_SECRET_KEY=my-secret-key \
		-v $(PWD)/data:/app/data:ro \
		catalog-service

# Docker Compose commands
compose-up:
	docker-compose up -d

compose-down:
	docker-compose down

# Generate a secure JWT secret key
generate-jwt-secret:
	@openssl rand -base64 32 | tr -d '\n' && echo

# Clean Docker artifacts
docker-clean:
	docker system prune -f
	docker image prune -f

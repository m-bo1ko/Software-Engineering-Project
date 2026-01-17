.PHONY: build run test clean docker-build docker-up docker-down

# Build the application
build:
	go build -o forecast-service ./cmd/main.go

# Run the application
run:
	go run ./cmd/main.go

# Run tests
test:
	go test ./... -v -cover

# Clean build artifacts
clean:
	rm -f forecast-service

# Docker build
docker-build:
	docker build -t forecast-service:latest .

# Docker compose up
docker-up:
	docker-compose up -d

# Docker compose down
docker-down:
	docker-compose down

# Run linter
lint:
	golangci-lint run

# Install dependencies
deps:
	go mod download
	go mod tidy
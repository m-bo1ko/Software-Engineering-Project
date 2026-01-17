.PHONY: build run test clean docker-build docker-run

# Build the application
build:
	go build -o bin/iot-control-service ./cmd/main.go

# Run the application
run:
	go run ./cmd/main.go

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Build Docker image
docker-build:
	docker build -t iot-control-service:latest .

# Run Docker container
docker-run:
	docker run -p 8083:8083 --env-file .env iot-control-service:latest

# Install dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run ./...

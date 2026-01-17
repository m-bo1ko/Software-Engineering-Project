.PHONY: all build run test clean docker-build docker-up docker-down lint fmt help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOFMT=gofmt
BINARY_NAME=security-service
MAIN_PATH=./cmd/main.go

# Docker parameters
DOCKER_COMPOSE=docker-compose

all: clean build

## Build the application
build:
	@echo "Building..."
	$(GOBUILD) -o $(BINARY_NAME) $(MAIN_PATH)

## Run the application
run:
	@echo "Running..."
	$(GORUN) $(MAIN_PATH)

## Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./tests/...

## Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./tests/...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

## Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

## Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

## Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .

## Lint code (requires golangci-lint)
lint:
	@echo "Linting..."
	golangci-lint run

## Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t security-service:latest .

## Start Docker containers
docker-up:
	@echo "Starting Docker containers..."
	$(DOCKER_COMPOSE) up -d

## Start Docker containers with dev profile (includes mongo-express)
docker-up-dev:
	@echo "Starting Docker containers (dev mode)..."
	$(DOCKER_COMPOSE) --profile dev up -d

## Stop Docker containers
docker-down:
	@echo "Stopping Docker containers..."
	$(DOCKER_COMPOSE) down

## View Docker logs
docker-logs:
	@echo "Viewing Docker logs..."
	$(DOCKER_COMPOSE) logs -f security-service

## Reset Docker volumes (destructive!)
docker-reset:
	@echo "Resetting Docker volumes..."
	$(DOCKER_COMPOSE) down -v

## Run integration tests
integration-test: docker-up
	@echo "Waiting for services to start..."
	@sleep 10
	@echo "Running integration tests..."
	$(GOTEST) -v -tags=integration ./tests/...
	$(DOCKER_COMPOSE) down

## Generate API documentation (requires swag)
docs:
	@echo "Generating API documentation..."
	swag init -g cmd/main.go -o docs

## Show help
help:
	@echo "Available commands:"
	@echo ""
	@echo "  make build          - Build the application"
	@echo "  make run            - Run the application"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make deps           - Download and tidy dependencies"
	@echo "  make fmt            - Format code"
	@echo "  make lint           - Lint code"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-up      - Start Docker containers"
	@echo "  make docker-up-dev  - Start Docker containers (dev mode)"
	@echo "  make docker-down    - Stop Docker containers"
	@echo "  make docker-logs    - View Docker logs"
	@echo "  make docker-reset   - Reset Docker volumes (destructive!)"
	@echo "  make integration-test - Run integration tests"
	@echo "  make docs           - Generate API documentation"
	@echo "  make help           - Show this help"

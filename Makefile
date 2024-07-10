# Makefile
# Variables
APP_NAME = trading-algorithm-backtesting-system
DOCKER_COMPOSE = docker-compose
PROMETHEUS_CONFIG_DIR = ./prometheus

# Default target
.PHONY: all
all: build

# Build the Go application
.PHONY: build
build:
	@echo "Building Go application..."
	@docker build -t $(APP_NAME):latest .

# Build Prometheus Docker image
.PHONY: build-prometheus
build-prometheus:
	@echo "Building Prometheus image..."
	@docker build -t prometheus:latest $(PROMETHEUS_CONFIG_DIR)

# Start the application and Prometheus using Docker Compose
.PHONY: up
up:
	@echo "Starting services..."
	@$(DOCKER_COMPOSE) up --build

# Stop the application and Prometheus
.PHONY: down
down:
	@echo "Stopping services..."
	@$(DOCKER_COMPOSE) down

# Clean up Docker images and containers
.PHONY: clean
clean:
	@echo "Cleaning up Docker images and containers..."
	@docker system prune -af
	@docker volume prune -f

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@go test ./...

# Rebuild everything
.PHONY: rebuild
rebuild: clean build build-prometheus up

.PHONY: generate-mocks
generate-mocks:
	@echo "Generating mocks..."
	@go generate ./...
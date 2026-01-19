# This Makefile provides common development tasks for building, testing,
# and running the application.
#
# Usage:
#   make build    - Compile the application
#   make run      - Build and run the application
#   make clean    - Remove build artifacts

# Application configuration
APP_NAME := go-gitops-app
BIN_DIR := bin
CMD_DIR := ./cmd
BINARY := $(BIN_DIR)/server

# Go configuration
GO := go
GOFLAGS := -v

# Environment configuration
PORT ?= 8080
LOG_LEVEL ?= info

# Phony targets
.PHONY: all build run clean deps tidy

## build: Compile the application binary
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -o $(BINARY) $(CMD_DIR)
	@echo "Build complete: $(BINARY)"

## run: Build and run the application
run: build
	@echo "Starting $(APP_NAME) on port $(PORT)..."
	LOG_LEVEL=$(LOG_LEVEL) PORT=$(PORT) $(BINARY)

## clean: Remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BIN_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download

## tidy: Tidy Go modules
tidy:
	@echo "Tidying modules..."
	$(GO) mod tidy

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):latest .

## docker-run: Run Docker container
docker-run:
	docker run -p $(PORT):$(PORT) -e PORT=$(PORT) -e LOG_LEVEL=$(LOG_LEVEL) $(APP_NAME):latest

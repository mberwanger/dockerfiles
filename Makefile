.PHONY: help build generate generate-all generate-workflow clean test install default

# Default target
.DEFAULT_GOAL := default

default: generate-all generate-workflow ## Generate all Dockerfiles and workflow (default)

# Binary name
BINARY_NAME := dockerfiles
BUILD_DIR := bin

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..." >&2
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./tool/main.go

generate-all: build ## Generate all Dockerfiles
	@echo "Generating all Dockerfiles..."
	@./$(BUILD_DIR)/$(BINARY_NAME) generate image --all

generate: build ## Generate Dockerfiles for specific image (usage: make generate IMAGE=core)
	@if [ -z "$(IMAGE)" ]; then \
		echo "Please specify IMAGE. Usage: make generate IMAGE=core"; \
		exit 1; \
	fi
	@echo "Generating $(IMAGE) Dockerfiles..."
	@./$(BUILD_DIR)/$(BINARY_NAME) generate image $(IMAGE)

generate-workflow: build ## Generate dynamic GitHub Actions workflow
	@echo "Generating dynamic workflow..."
	@./$(BUILD_DIR)/$(BINARY_NAME) generate workflow -o .github/workflows/dockerfiles.yaml

clean-generated: build ## Clean generated Dockerfiles
	@echo "Cleaning generated files..."
	@./$(BUILD_DIR)/$(BINARY_NAME) clean

clean: ## Clean build artifacts and generated files
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)

test: ## Run tests
	@echo "Running tests..."
	@go test ./...

install: build ## Install the binary to $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(shell go env GOPATH)/bin/

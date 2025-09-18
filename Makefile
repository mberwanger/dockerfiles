.PHONY: help build generate generate-all dependency-index clean test

# Default target
.DEFAULT_GOAL := help

# Binary name
BINARY_NAME := dockerfiles-gen
BUILD_DIR := bin

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..." >&2
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./tooling/cmd

generate-all: build ## Generate all Dockerfiles
	@echo "Generating all Dockerfiles..."
	@./$(BUILD_DIR)/$(BINARY_NAME) -generate-all

generate: build ## Generate Dockerfiles for specific image (usage: make generate IMAGE=core)
	@if [ -z "$(IMAGE)" ]; then \
		echo "Please specify IMAGE. Usage: make generate IMAGE=core"; \
		exit 1; \
	fi
	@echo "Generating $(IMAGE) Dockerfiles..."
	@./$(BUILD_DIR)/$(BINARY_NAME) -generate $(IMAGE)

dependency-index: build ## Generate dependency index for CI
	@./$(BUILD_DIR)/$(BINARY_NAME) -dependency-index

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

test: ## Run tests
	@echo "Running tests..."
	@go test ./...

install: build ## Install the binary to $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(shell go env GOPATH)/bin/

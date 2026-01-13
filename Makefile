.PHONY: build clean install test run-auth run-import run-list-types help

# Binary name
BINARY_NAME=any-vcard
BUILD_DIR=build

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOINSTALL=$(GOCMD) install

# Build flags
LDFLAGS=-ldflags "-s -w"

all: help

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/any-vcard
	@echo "✓ Binary built at $(BUILD_DIR)/$(BINARY_NAME)"

## build-all: Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/any-vcard
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/any-vcard
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/any-vcard
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/any-vcard
	@echo "✓ Built binaries for all platforms in $(BUILD_DIR)/"

## install: Install the binary to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GOINSTALL) ./cmd/any-vcard
	@echo "✓ Installed $(BINARY_NAME) to GOPATH/bin"

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@echo "✓ Cleaned build artifacts"

## test: Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

## deps: Download and tidy dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "✓ Dependencies updated"

## run-auth: Run authentication command
run-auth: build
	@$(BUILD_DIR)/$(BINARY_NAME) auth

## run-import: Run import command with sample file (requires ANYTYPE_APP_KEY and ANYTYPE_SPACE_ID)
run-import: build
	@if [ -z "$(ANYTYPE_APP_KEY)" ]; then \
		echo "Error: ANYTYPE_APP_KEY environment variable not set"; \
		exit 1; \
	fi
	@if [ -z "$(ANYTYPE_SPACE_ID)" ]; then \
		echo "Error: ANYTYPE_SPACE_ID environment variable not set"; \
		exit 1; \
	fi
	@$(BUILD_DIR)/$(BINARY_NAME) import examples/sample-contacts.vcf

## run-import-dry: Run import in dry-run mode with sample file
run-import-dry: build
	@if [ -z "$(ANYTYPE_APP_KEY)" ]; then \
		echo "Error: ANYTYPE_APP_KEY environment variable not set"; \
		exit 1; \
	fi
	@if [ -z "$(ANYTYPE_SPACE_ID)" ]; then \
		echo "Error: ANYTYPE_SPACE_ID environment variable not set"; \
		exit 1; \
	fi
	@$(BUILD_DIR)/$(BINARY_NAME) import --dry-run examples/sample-contacts.vcf

## run-list-types: List available object types (requires ANYTYPE_APP_KEY and ANYTYPE_SPACE_ID)
run-list-types: build
	@if [ -z "$(ANYTYPE_APP_KEY)" ]; then \
		echo "Error: ANYTYPE_APP_KEY environment variable not set"; \
		exit 1; \
	fi
	@if [ -z "$(ANYTYPE_SPACE_ID)" ]; then \
		echo "Error: ANYTYPE_SPACE_ID environment variable not set"; \
		exit 1; \
	fi
	@$(BUILD_DIR)/$(BINARY_NAME) types list

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@$(GOCMD) fmt ./...
	@echo "✓ Code formatted"

## lint: Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## vet: Run go vet
vet:
	@echo "Running go vet..."
	@$(GOCMD) vet ./...
	@echo "✓ Vet passed"

## check: Run all checks (fmt, vet, test)
check: fmt vet test
	@echo "✓ All checks passed"

## help: Display this help message
help:
	@echo "any-vcard - vCard importer for Anytype"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: test coverage lint build run clean test-watch ui-install ui-build ui-dev

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=orochi
BINARY_UNIX=$(BINARY_NAME)_unix
MAIN_PATH=./cmd/orochi

# Build the project (includes UI)
build: ui-build
	$(GOBUILD) -o $(BINARY_NAME) -v $(MAIN_PATH)

# Run tests
test:
	$(GOTEST) -v -race ./...

# Run tests with coverage
coverage:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linter
lint:
	@if ! command -v golangci-lint > /dev/null; then \
		echo "Installing golangci-lint v1.62.2 (same as CI)..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.62.2; \
	fi
	@golangci-lint version
	@golangci-lint run

# Run tests in watch mode
test-watch:
	@which gotestsum > /dev/null || (echo "Installing gotestsum..." && go install gotest.tools/gotestsum@latest)
	gotestsum --watch

# Run the application
run: build
	./$(BINARY_NAME)

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -f coverage.out coverage.html

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Build for multiple platforms
build-all:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)_darwin_amd64 -v $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BINARY_NAME)_darwin_arm64 -v $(MAIN_PATH)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)_linux_amd64 -v $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)_windows_amd64.exe -v $(MAIN_PATH)

# Install development tools
dev-tools:
	@echo "Installing golangci-lint v1.62.2 (same as CI)..."
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.62.2
	@go install gotest.tools/gotestsum@latest
	@echo "Installing pre-commit..."
	@pip install pre-commit || pip3 install pre-commit
	@pre-commit install

# Install UI dependencies
ui-install:
	cd web-ui && npm install

# Build UI for production
ui-build: ui-install
	cd web-ui && npm run build

# Run UI in development mode
ui-dev:
	cd web-ui && npm start

# Build without UI (for CI/testing)
build-no-ui:
	$(GOBUILD) -o $(BINARY_NAME) -v $(MAIN_PATH)
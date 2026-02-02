# Agentic RAG Go - Makefile

.PHONY: all build run test lint clean tidy help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofumpt
GOLINT=golangci-lint

# Binary names
BINARY_NAME=agentic-rag
BINARY_PATH=./bin/$(BINARY_NAME)

# Directories
CMD_DIR=./cmd/server
INTERNAL_DIR=./internal/...

all: lint test build

## build: Build the application
build:
	@echo "Building..."
	@mkdir -p bin
	$(GOBUILD) -o $(BINARY_PATH) $(CMD_DIR)

## run: Run the application
run:
	$(GORUN) $(CMD_DIR)

## test: Run tests
test:
	$(GOTEST) -v -race -cover $(INTERNAL_DIR) ./...

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	$(GOCMD) tool cover -func=coverage.out

## coverage-report: Generate and open coverage report
coverage-report: test-coverage
	@echo "Opening coverage report..."
	@open coverage.html || xdg-open coverage.html || echo "Please open coverage.html manually"

## coverage-check: Check if coverage meets threshold
coverage-check:
	@echo "Running tests with coverage..."
	@$(GOTEST) -coverprofile=coverage.out ./... > /dev/null 2>&1
	@echo "Checking coverage threshold..."
	@$(GOCMD) tool cover -func=coverage.out | \
		awk '/^total:/ {cov=$$3+0; if (cov < 80.0) {printf "Coverage %.1f%% is below 80%% threshold\n", cov; exit 1} else {printf "Coverage %.1f%% meets threshold\n", cov}}'

## lint: Run linter
lint:
	$(GOLINT) run ./...

## fmt: Format code
fmt:
	$(GOFMT) -w .

## tidy: Tidy go modules
tidy:
	$(GOMOD) tidy

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	$(GOCMD) clean

## deps: Install dependencies
deps:
	$(GOGET) -u ./...
	$(GOMOD) tidy

## swagger: Generate Swagger documentation
swagger:
	@echo "Generating Swagger docs..."
	swag init -g cmd/server/main.go -o docs

## help: Show this help
help:
	@echo "Usage:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

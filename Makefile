# Web Pages Analyzer Makefile

BINARY_NAME=web-pages-analyzer
MAIN_PATH=.
PORT=8080

# Go related variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOINSTALL=$(GOCMD) install

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

.PHONY: help build run test test-verbose test-coverage clean deps tidy check-deps install-tools generate-mocks

# Build the application
build:
	@echo "$(BLUE)Building $(BINARY_NAME)...$(NC)"
	$(GOBUILD) -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "$(GREEN)Build completed: $(BINARY_NAME)$(NC)"

# Run the service
run: build 
	@echo "$(BLUE)Starting web pages analyzer service on port $(PORT)...$(NC)"
	@echo "$(YELLOW)Access the web interface at: http://localhost:$(PORT)$(NC)"
	@echo "$(YELLOW)Press Ctrl+C to stop the service$(NC)"
	@echo ""
	./$(BINARY_NAME)

# Run all unit tests
test:
	@echo "$(BLUE)Running all unit tests...$(NC)"
	$(GOTEST) ./internal/... -v
	@echo "$(GREEN)All tests completed!$(NC)"

# Run unit tests with coverage
test-coverage:
	@echo "$(BLUE)Running all unit tests with coverage...$(NC)"
	@echo ""
	@echo "$(YELLOW)HTTP Client Tests:$(NC)"
	$(GOTEST) ./internal/infrastructure/clients/http -cover
	@echo ""
	@echo "$(YELLOW)HTML Parser Tests:$(NC)"
	$(GOTEST) ./internal/infrastructure/html_parser -cover
	@echo ""
	@echo "$(YELLOW)Webpage Analyzer Use Case Tests:$(NC)"
	$(GOTEST) ./internal/usecases/webpage_analyzer -cover
	@echo ""
	@echo "$(YELLOW)Controller Tests:$(NC)"
	$(GOTEST) ./internal/controllers/webpage_analyzer -cover
	@echo ""
	@echo "$(GREEN)Coverage analysis completed!$(NC)"

# Clean go build
clean:
	@echo "$(BLUE)Cleaning build artifacts...$(NC)"
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	@echo "$(GREEN)Clean completed!$(NC)"

# Download & installdependencies
deps:
	@echo "$(BLUE)Downloading dependencies...$(NC)"
	$(GOGET) -d ./...
	@echo "$(GREEN)Dependencies downloaded!$(NC)"

# Tidy up go.mod and go.sum
tidy:
	@echo "$(BLUE)Tidying up go modules...$(NC)"
	$(GOMOD) tidy
	@echo "$(GREEN)Go modules tidied!$(NC)"

# Generate mocks using mockgen
gen-mocks:
	@echo "$(BLUE)Generating mock files...$(NC)"
	@echo "$(YELLOW)Generating HTTP client mock...$(NC)"
	mockgen -source=internal/domain/clients/http/interface.go -destination=internal/infrastructure/clients/http/mocks/mock_http_client.go -package=mocks
	@echo "$(YELLOW)Generating HTML parser mock...$(NC)"
	mockgen -source=internal/domain/html/html.go -destination=internal/infrastructure/html_parser/mocks/mock_parser_html.go -package=mocks
	@echo "$(YELLOW)Generating parser factory mock...$(NC)"
	mockgen -source=internal/domain/html/parser_factory.go -destination=internal/infrastructure/html_parser/mocks/mock_parser_factory.go -package=mocks
	@echo "$(YELLOW)Generating webpage analyzer mock...$(NC)"
	mockgen -source=internal/domain/webpage/page.go -destination=internal/usecases/webpage_analyzer/mocks/mock_analyzer.go -package=mocks
	@echo "$(GREEN)All mocks generated!$(NC)"


# Go Linter commands
lint:
	@echo "$(BLUE)Running golangci-lint...$(NC)"
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "$(RED)golangci-lint not found. Run 'make install-tools' first.$(NC)"; \
		exit 1; \
	fi
	golangci-lint run ./...
	@echo "$(GREEN)Linting completed!$(NC)"

# Format Go code
fmt:
	@echo "$(BLUE)Formatting Go code...$(NC)"
	gofmt -s -w .
	@echo "$(GREEN)Code formatting completed!$(NC)"

# Check if Go code is formatted
fmt-check:
	@echo "$(BLUE)Checking Go code formatting...$(NC)"
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "$(RED)The following files are not formatted correctly:$(NC)"; \
		gofmt -l .; \
		exit 1; \
	fi
	@echo "$(GREEN)All files are properly formatted!$(NC)"

# Install development tools
install-tools:
	@echo "$(BLUE)Installing development tools...$(NC)"
	$(GOGET) go.uber.org/mock/mockgen@latest
	@echo "$(YELLOW)Installing golangci-lint...$(NC)"
	$(GOINSTALL) github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "$(GREEN)Development tools installed!$(NC)"


# Build docker image
docker-build:
	@echo "$(BLUE)Building Docker image...$(NC)"
	docker build -t web-pages-analyzer .
	@echo "$(GREEN)Docker image built!$(NC)"

docker-run:
	@echo "$(BLUE)Running Docker container...$(NC)"
	docker run -p $(PORT):$(PORT) web-pages-analyzer

# Run docker image from pre-built image
docker-run-prebuilt:
	@echo "$(BLUE)Running Docker container...$(NC)"
	docker run -d -p $(PORT):$(PORT) --name web-pages-analyzer-app --rm namalsanjaya/web-pages-analyzer:v1.0.0

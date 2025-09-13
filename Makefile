.PHONY: build clean test run-server help

BINARY_NAME=stresstea
BUILD_DIR=bin
MAIN_PATH=main.go


GREEN=\033[0;32m
NC=\033[0m 

help: ## Show help
	@echo "$(GREEN)Available commands:$(NC)"
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}'

build: ## Build binary file
	@echo "$(GREEN)Building $(BINARY_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "$(GREEN)Done! Binary file: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

test: ## Run tests
	@echo "$(GREEN)Running tests...$(NC)"
	@go test -v ./...

deps: ## Update dependencies
	@echo "$(GREEN)Updating dependencies...$(NC)"
	@go mod tidy
	@go mod download

lint: ## Run linter
	@echo "$(GREEN)Checking code...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

format: ## Format code
	@echo "$(GREEN)Formatting code...$(NC)"
	@go fmt ./...

dev: deps build ## Development: update dependencies and build
	@echo "$(GREEN)Ready for development!$(NC)"

run-server: ## Run test server
	@echo "$(GREEN)Starting test server...$(NC)"
	@go run test-server/main.go

test-compact: build ## Test compact TUI mode
	@echo "$(GREEN)Testing compact TUI mode...$(NC)"
	@./$(BUILD_DIR)/$(BINARY_NAME) run -t http://httpbin.org/get -r 10 -d 10s -o compact

test-console: build ## Test console mode
	@echo "$(GREEN)Testing console mode...$(NC)"
	@./$(BUILD_DIR)/$(BINARY_NAME) run -t http://httpbin.org/get -r 10 -d 10s -o console


all: format lint deps test build ## Full build with tests
	@echo "$(GREEN)Full build completed!$(NC)"

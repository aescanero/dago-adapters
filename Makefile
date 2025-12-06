.PHONY: help deps test test-coverage test-verbose lint fmt clean release

# Default target
help:
	@echo "Available targets:"
	@echo "  help           - Show this help message"
	@echo "  deps           - Download Go module dependencies"
	@echo "  test           - Run tests with coverage"
	@echo "  test-coverage  - Run tests and show coverage report"
	@echo "  test-verbose   - Run tests with verbose output"
	@echo "  lint           - Run golangci-lint"
	@echo "  fmt            - Format code with gofmt and goimports"
	@echo "  clean          - Remove build artifacts and coverage files"
	@echo "  release        - Create a new release tag (use VERSION=vX.Y.Z)"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod verify

# Run tests with coverage
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

# Run tests and show coverage report
test-coverage: test
	@echo "Generating coverage report..."
	go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated: coverage.html"
	go tool cover -func=coverage.txt

# Run tests with verbose output
test-verbose:
	@echo "Running tests (verbose)..."
	go test -v -race ./...

# Run golangci-lint
lint:
	@echo "Running golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m ./...; \
	else \
		echo "golangci-lint not found. Install with:"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

# Format code
fmt:
	@echo "Formatting code..."
	gofmt -s -w .
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "goimports not found. Install with:"; \
		echo "  go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	rm -f coverage.txt coverage.html
	go clean -cache -testcache -modcache

# Create a new release tag
release:
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make release VERSION=vX.Y.Z"; \
		exit 1; \
	fi
	@echo "Creating release $(VERSION)..."
	@if git diff-index --quiet HEAD --; then \
		git tag -a $(VERSION) -m "Release $(VERSION)"; \
		git push origin $(VERSION); \
		echo "✅ Tag $(VERSION) created and pushed"; \
		echo "GitHub Actions will create the release automatically"; \
	else \
		echo "❌ Working directory not clean. Commit changes first."; \
		exit 1; \
	fi

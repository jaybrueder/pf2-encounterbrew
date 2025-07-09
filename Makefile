# Build the application
all: build

build:
	@echo "Building..."
	@templ generate
	@tailwindcss -i cmd/web/assets/css/input.css -o cmd/web/assets/css/output.css
	@if [ "$$(uname -s)" = "Darwin" ]; then \
		CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags '-extldflags "-static" -s -w' -o main cmd/api/main.go; \
	else \
		CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static" -s -w' -o main cmd/api/main.go; \
	fi

# Run the application
run:
	@go run cmd/api/main.go

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

# Test the application
test:
	@echo "Testing..."
	@go test ./tests/

# Live Reload
watch:
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi

# Lint all code
lint: imports golangci-lint tidy

# Format Go code
fmt:
	@echo "Running go fmt..."
	@go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Fix imports
imports:
	@echo "Running goimports..."
	@if command -v goimports > /dev/null; then \
		goimports -w .; \
	else \
		echo "goimports not installed. Installing..."; \
		go install golang.org/x/tools/cmd/goimports@latest; \
		goimports -w .; \
	fi

# Run golangci-lint
golangci-lint:
	@echo "Running golangci-lint..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Installing..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin; \
		golangci-lint run; \
	fi

# Tidy go modules
tidy:
	@echo "Running go mod tidy..."
	@go mod tidy

.PHONY: all build run test clean watch

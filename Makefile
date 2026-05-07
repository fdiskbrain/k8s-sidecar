# Makefile for k8s-sidecar

.PHONY: build test clean docker-build fmt vet lint

# Variables
BINARY_NAME=k8s-sidecar
VERSION?=latest
DOCKER_IMAGE=k8s-sidecar:$(VERSION)

# Go commands
GO_BUILD=go build
GO_TEST=go test
GO_FMT=gofmt
GO_VET=go vet

# Flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

all: fmt vet test build

# Build the binary
build:
	$(GO_BUILD) $(LDFLAGS) -o ./bin/$(BINARY_NAME) ./cmd/sidecar/main.go

# Run tests
test:
	$(GO_TEST) -v -race -coverprofile=coverage.out ./...

# Format code
fmt:
	go fmt ./...

# Run go vet
vet:
	$(GO_VET) ./...

# Clean build artifacts
clean:
	rm -rf ./bin ./coverage.out
	go clean

# Build Docker image
docker-build:
	docker build -t $(DOCKER_IMAGE) .

# Run locally (for testing)
run:
	go run ./cmd/sidecar/main.go

# Install dependencies
deps:
	go mod download
	go mod tidy

# Help
help:
	@echo "Available targets:"
	@echo "  build       - Build the binary"
	@echo "  test        - Run tests"
	@echo "  fmt         - Format code"
	@echo "  vet         - Run go vet"
	@echo "  clean       - Clean build artifacts"
	@echo "  docker-build - Build Docker image"
	@echo "  run         - Run locally"
	@echo "  deps        - Install dependencies"
	@echo "  help        - Show this help"

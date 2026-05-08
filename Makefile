# Makefile for k8s-sidecar

.PHONY: build test clean docker-build docker-buildx fmt vet lint help

# Variables
BINARY_NAME=k8s-sidecar
VERSION?=latest
DOCKER_IMAGE=k8s-sidecar:$(VERSION)
PLATFORMS?=linux/amd64,linux/arm64,linux/arm/v7

# Go commands
GO_BUILD=go build
GO_TEST=go test
GO_FMT=gofmt
GO_VET=go vet

# Flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

# Go 代理加速（中国镜像）
export GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,direct
export GONOSUMDB=*
export GONOPROXY=*

all: fmt vet test build

# Build the binary
build:
	$(GO_BUILD) $(LDFLAGS) -o ./bin/$(BINARY_NAME) ./cmd/sidecar/main.go

# Build for specific architecture
build-arch:
ifndef GOARCH
	$(error GOARCH is not set. Usage: make build-arch GOARCH=amd64)
endif
	CGO_ENABLED=0 GOOS=linux GOARCH=$(GOARCH) $(GO_BUILD) $(LDFLAGS) -o ./bin/$(BINARY_NAME)-$(GOARCH) ./cmd/sidecar/main.go

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

# Build Docker image (single platform)
docker-build:
	docker build --build-arg VERSION=$(VERSION) -t $(DOCKER_IMAGE) .

# Build multi-platform Docker images using buildx
docker-buildx:
	docker buildx create --name k8s-sidecar-builder || true
	docker buildx use k8s-sidecar-builder
	docker buildx build \
		--platform $(PLATFORMS) \
		--build-arg VERSION=$(VERSION) \
		-t $(DOCKER_IMAGE) \
		--push \
		.
	docker buildx stop k8s-sidecar-builder

# Build and load to local Docker daemon (single platform for testing)
docker-buildx-local:
	docker buildx create --name k8s-sidecar-builder || true
	docker buildx use k8s-sidecar-builder
	docker buildx build \
		--platform linux/amd64 \
		--build-arg VERSION=$(VERSION) \
		-t $(DOCKER_IMAGE) \
		--load \
		.
	docker buildx stop k8s-sidecar-builder

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
	@echo "  build              - Build the binary"
	@echo "  build-arch         - Build for specific architecture (requires GOARCH)"
	@echo "  test               - Run tests"
	@echo "  fmt                - Format code"
	@echo "  vet                - Run go vet"
	@echo "  clean              - Clean build artifacts"
	@echo "  docker-build       - Build Docker image (current platform)"
	@echo "  docker-buildx      - Build multi-platform Docker images and push to registry"
	@echo "  docker-buildx-local - Build Docker image using buildx (for local testing)"
	@echo "  run                - Run locally"
	@echo "  deps               - Install dependencies"
	@echo "  help               - Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make build-arch GOARCH=amd64"
	@echo "  make build-arch GOARCH=arm64"
	@echo "  make docker-buildx PLATFORMS=linux/amd64,linux/arm64"
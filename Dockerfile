# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download || true

# Copy source code
COPY . .

# Build the binary
ARG VERSION=latest
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make build LDFLAGS="-ldflags -X main.Version=${VERSION}"

# Final stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates && \
    addgroup -S sidecar && \
    adduser -S sidecar -G sidecar

# Copy binary from builder
COPY --from=builder /app/bin/k8s-configmap-sidecar .

# Create config directory
RUN mkdir -p /etc/sidecar/config && \
    chown -R sidecar:sidecar /etc/sidecar

USER sidecar

ENTRYPOINT ["./k8s-configmap-sidecar"]

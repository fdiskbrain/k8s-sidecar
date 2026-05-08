# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# 配置 Alpine APK 国内镜像源并安装构建依赖
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk add --no-cache git make

# 配置 Go 代理加速（中国镜像）
ENV GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,direct
ENV GONOSUMDB=*
ENV GONOPROXY=*

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download || true

# Copy source code
COPY . .

# Build the binary with platform detection
ARG VERSION=latest
ARG TARGETARCH

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} make build LDFLAGS="-ldflags -X main.Version=${VERSION}"

# Final stage
FROM alpine:latest

WORKDIR /app

# 配置 Alpine APK 国内镜像源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk --no-cache add ca-certificates && \
    addgroup -S sidecar && \
    adduser -S sidecar -G sidecar

# Copy binary from builder
COPY --from=builder /app/bin/k8s-sidecar .

# Create config directory
RUN mkdir -p /etc/sidecar/config && \
    chown -R sidecar:sidecar /etc/sidecar

USER sidecar

ENTRYPOINT ["./k8s-sidecar"]
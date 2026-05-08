# DevContainer for k8s-sidecar
FROM golang:1.25-alpine

# 配置 Alpine APK 国内镜像源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 安装开发必需工具
RUN apk add --no-cache \
    git \
    make \
    gcc \
    musl-dev \
    curl \
    ca-certificates \
    bash \
    vim \
    && update-ca-certificates

# 创建非 root 用户
ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=$USER_UID

RUN addgroup -g $USER_GID $USERNAME \
    && adduser -D -u $USER_UID -G $USERNAME $USERNAME \
    && echo "$USERNAME ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers.d/$USERNAME \
    && chmod 0440 /etc/sudoers.d/$USERNAME

# 设置工作目录
WORKDIR /workspace

# 配置 Go 环境变量
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$PATH
ENV CGO_ENABLED=1

# 配置 Go 代理加速
ENV GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,direct
ENV GONOSUMDB=*
ENV GONOPROXY=*

# 切换到非 root 用户
USER $USERNAME

# 保持容器运行
CMD ["sleep", "infinity"]

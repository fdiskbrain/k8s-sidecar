#!/usr/bin/env bash

# 配置系统包管理器源为中国镜像加速
if [ -f /etc/os-release ]; then
    . /etc/os-release
    case $ID in
        debian|ubuntu)
            echo "Detected Debian/Ubuntu. Configuring APT mirrors..."
            
            # 处理新版 Debian (Bookworm+) 使用 sources.list.d 下的 deb822 格式
            if [ -f /etc/apt/sources.list.d/debian.sources ]; then
                echo "Updating /etc/apt/sources.list.d/debian.sources..."
                sed -i 's|https://deb.debian.org|https://mirrors.aliyun.com|g' /etc/apt/sources.list.d/debian.sources
                sed -i 's|https://security.debian.org|https://mirrors.aliyun.com|g' /etc/apt/sources.list.d/debian.sources
            fi

            # 处理传统 sources.list 文件 (兼容旧版或 Ubuntu)
            if [ -f /etc/apt/sources.list ]; then
                echo "Updating /etc/apt/sources.list..."
                # 备份原始源
                cp /etc/apt/sources.list /etc/apt/sources.list.bak
                # 使用阿里云镜像
                sed -i 's/deb.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list
                sed -i 's/security.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list
            fi
            
            # 更新包列表
            apt-get update -y
            ;;
        alpine)
            echo "Detected Alpine. Configuring APK mirrors..."
            # 使用阿里云镜像
            sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
            # 更新包索引
            apk update
            ;;
        *)
            echo "Unknown OS ID: $ID. Skipping system package mirror configuration."
            ;;
    esac
fi

# 配置 Go 模块代理为中国镜像加速
echo "Configuring Go proxy..."
export GOPROXY="https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,direct"
export GOSUMDB="sum.golang.google.cn"

# 将环境变量写入 shell 配置文件，以便新终端会话生效
if [ -f ~/.bashrc ]; then
    echo 'export GOPROXY="https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,direct"' >> ~/.bashrc
    echo 'export GOSUMDB="sum.golang.google.cn"' >> ~/.bashrc
fi
if [ -f ~/.zshrc ]; then
    echo 'export GOPROXY="https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,direct"' >> ~/.zshrc
    echo 'export GOSUMDB="sum.golang.google.cn"' >> ~/.zshrc
fi

# Install glab (GitLab CLI tool)
echo "Installing glab (GitLab CLI)..."
GLAB_VERSION="1.95.0"
GLAB_URL="https://gitlab.com/gitlab-org/cli/-/releases/v${GLAB_VERSION}/downloads/glab_${GLAB_VERSION}_linux_$(go env GOARCH).deb"

# Download with timeout and error handling
if wget -q --timeout=30 "${GLAB_URL}" -O /tmp/glab.deb; then
    echo "✓ Downloaded glab successfully"
    
    # Install with error handling
    if sudo dpkg -i /tmp/glab.deb; then
        echo "✓ glab installed successfully"
        # Clean up temporary file
        rm -f /tmp/glab.deb
        echo "✓ Temporary files cleaned up"
    else
        echo "✗ Error: Failed to install glab"
        echo "Please install manually from: ${GLAB_URL}"
        # Keep the deb file for manual installation
        exit 1
    fi
else
    echo "✗ Error: Failed to download glab from ${GLAB_URL}"
    echo "Please check your network connection or install manually"
    exit 1
fi

echo ""
echo "========================================="
echo "DevContainer setup completed!"
echo "GOPROXY: ${GOPROXY}"
echo "GOSUMDB: ${GOSUMDB}"
echo "glab: $(glab --version 2>/dev/null || echo 'installed')"
echo "========================================="

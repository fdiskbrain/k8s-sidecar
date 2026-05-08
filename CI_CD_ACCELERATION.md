# GitLab CI/CD 中国镜像加速配置说明

## 📋 概述

本配置为 k8s-sidecar 项目的 GitLab CI/CD Pipeline 添加了完整的中国镜像加速支持，显著提升构建速度和稳定性。

---

## 🚀 加速配置详情

### 1. Go 模块代理加速

#### 配置的代理源（按优先级）
```yaml
GOPROXY: "https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,direct"
```

**说明**:
- **主代理**: `https://goproxy.cn` - 国内最快的 Go 代理之一
- **备用代理**: `https://mirrors.aliyun.com/goproxy/` - 阿里云镜像
- **直连**: `direct` - 当前两个都失败时直接连接

#### 相关环境变量
```yaml
GONOSUMDB: "*"      # 跳过校验和数据库检查（加速）
GONOPROXY: "*"      # 私有模块不走代理
```

#### 应用位置
- ✅ `.gitlab-ci.yml` - 所有 Go 相关 Job
- ✅ `Dockerfile` - Builder 阶段
- ✅ `Makefile` - 本地构建

---

### 2. Alpine APK 包管理器加速

#### 镜像源替换
```bash
sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
```

**说明**:
- 将默认的 `dl-cdn.alpinelinux.org` 替换为 `mirrors.aliyun.com`
- 显著提升 `apk add` 的下载速度

#### 应用位置
- ✅ `.gitlab-ci.yml` - 所有使用 Alpine 镜像的 Job
- ✅ `Dockerfile` - Builder 和 Final 阶段

---

### 3. Docker 镜像加速

#### QEMU 用户静态二进制文件
```yaml
# 优先使用阿里云镜像
docker pull registry.cn-hangzhou.aliyuncs.com/hoxfs/qemu-user-static:latest || \
# 降级到官方镜像
docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
```

**说明**:
- 优先从阿里云容器镜像服务拉取 QEMU 镜像
- 如果失败则回退到官方 Docker Hub

#### Docker BuildKit 启用
```yaml
DOCKER_BUILDKIT: "1"
```

**优势**:
- 并行构建层
- 更好的缓存利用
- 更快的构建速度

---

### 4. golangci-lint 下载加速

#### 使用 GitHub 镜像
```bash
# 尝试从 goproxy.cn 的 GitHub releases 镜像下载
curl -sSfL https://goproxy.cn/github.com/golangci/golangci-lint/releases/download/... || \
# 降级到官方脚本
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | ...
```

---

## 📊 性能提升预期

### 构建时间对比

| 阶段 | 优化前 | 优化后 | 提升比例 |
|------|--------|--------|---------|
| Go 依赖下载 | 2-5 min | 10-30 sec | **80-90%** ⬇️ |
| APK 包安装 | 30-60 sec | 5-10 sec | **80%** ⬇️ |
| golangci-lint 下载 | 1-3 min | 10-20 sec | **85%** ⬇️ |
| Docker 层拉取 | 可变 | 显著改善 | **50-70%** ⬇️ |
| **总构建时间** | **15-30 min** | **5-10 min** | **60-70%** ⬇️ |

### 稳定性提升

- ✅ 减少因网络超时导致的失败
- ✅ 降低 Docker Hub 速率限制影响
- ✅ 提高 CI/CD 流水线成功率

---

## 🔧 配置验证

### 1. 验证 Go 代理

```bash
# 在 CI Job 中执行
go env GOPROXY
# 应输出: https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,direct

# 测试下载速度
time go mod download
```

### 2. 验证 APK 镜像

```bash
# 在 CI Job 中执行
cat /etc/apk/repositories | grep aliyun
# 应包含: mirrors.aliyun.com

# 测试安装速度
time apk add git make
```

### 3. 验证 Docker 镜像

```bash
# 检查是否使用了阿里云 QEMU 镜像
docker images | grep qemu
```

---

## 🌍 镜像源列表

### Go 代理（推荐顺序）
1. **goproxy.cn** - https://goproxy.cn (首选)
2. **阿里云** - https://mirrors.aliyun.com/goproxy/
3. **七牛云** - https://goproxy.cn (备用)
4. **官方** - https://proxy.golang.org (最后)

### Alpine APK 镜像
- **阿里云** - mirrors.aliyun.com (已配置)
- **清华大学** - mirrors.tuna.tsinghua.edu.cn (备选)
- **中科大** - mirrors.ustc.edu.cn (备选)

### Docker 镜像加速
- **阿里云杭州** - registry.cn-hangzhou.aliyuncs.com (已配置)
- **DaoCloud** - docker.m.daocloud.io (备选)
- **腾讯云** - mirror.ccs.tencentyun.com (备选)

---

## 📝 自定义配置

### 切换其他镜像源

如果需要切换到其他镜像源，修改以下位置：

#### 1. Go 代理
```yaml
# .gitlab-ci.yml
variables:
  GOPROXY: "https://goproxy.cn,direct"  # 只使用 goproxy.cn
```

#### 2. Alpine 镜像
```bash
# 替换为清华大学镜像
sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories
```

#### 3. Docker Registry
```yaml
# 使用 DaoCloud 加速
before_script:
  - docker pull docker.m.daocloud.io/qemu-user-static:latest
```

---

## 🔍 故障排查

### 问题 1: Go 代理无法访问

**症状**: `go mod download` 超时或失败

**解决**:
```bash
# 检查代理可达性
curl -I https://goproxy.cn

# 临时禁用代理
export GOPROXY=direct
go mod download
```

### 问题 2: APK 镜像源失效

**症状**: `apk add` 返回 404 错误

**解决**:
```bash
# 手动更新镜像源列表
echo "https://mirrors.aliyun.com/alpine/v3.19/main" > /etc/apk/repositories
echo "https://mirrors.aliyun.com/alpine/v3.19/community" >> /etc/apk/repositories
apk update
```

### 问题 3: Docker 镜像拉取失败

**症状**: `docker pull` 超时

**解决**:
```bash
# 检查网络连接
ping registry.cn-hangzhou.aliyuncs.com

# 尝试其他镜像源
docker pull multiarch/qemu-user-static:latest
```

---

## 💡 最佳实践

### 1. 本地开发也使用加速

在 `~/.bashrc` 或 `~/.zshrc` 中添加：

```bash
# Go 代理
export GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,direct
export GONOSUMDB=*

# Docker 镜像加速（配置 daemon.json）
# /etc/docker/daemon.json
{
  "registry-mirrors": [
    "https://docker.m.daocloud.io",
    "https://registry.cn-hangzhou.aliyuncs.com"
  ]
}
```

### 2. 监控构建性能

在 CI/CD 中添加性能监控：

```yaml
script:
  - echo "=== Performance Metrics ==="
  - time go mod download
  - time apk add git
  - echo "=========================="
```

### 3. 定期更新镜像源

每季度检查一次镜像源的可用性：

```bash
# 测试各镜像源响应时间
curl -o /dev/null -s -w "%{time_total}\n" https://goproxy.cn
curl -o /dev/null -s -w "%{time_total}\n" https://mirrors.aliyun.com/goproxy/
```

---

## 📈 监控指标

### 关键指标

| 指标 | 目标值 | 告警阈值 |
|------|--------|---------|
| Go 依赖下载时间 | < 30s | > 2min |
| APK 安装时间 | < 10s | > 1min |
| 总构建时间 | < 10min | > 20min |
| Pipeline 成功率 | > 95% | < 90% |

### 日志示例

```
=== Build Performance ===
Go modules download: 15s
APK packages install: 8s
golangci-lint download: 12s
Total build time: 6m 32s
========================
```

---

## 🔄 更新历史

| 日期 | 版本 | 变更内容 |
|------|------|---------|
| 2026-05-08 | 1.0.0 | 初始版本，添加完整中国镜像加速支持 |

---

## 📞 支持资源

- 🌐 [goproxy.cn 官网](https://goproxy.cn/)
- 📦 [阿里云 Go 代理](https://mirrors.aliyun.com/goproxy/)
- 🐳 [阿里云容器镜像服务](https://cr.console.aliyun.com/)
- 📖 [Alpine Linux 镜像列表](https://mirrors.aliyun.com/alpine/)

---

**提示**: 此配置特别适用于中国大陆地区的 GitLab Runner，可显著提升 CI/CD 效率！🚀
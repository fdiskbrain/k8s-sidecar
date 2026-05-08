# GitLab CI/CD Pipeline 使用指南

## 📋 目录

- [概述](#概述)
- [前置要求](#前置要求)
- [配置步骤](#配置步骤)
- [流水线说明](#流水线说明)
- [触发规则](#触发规则)
- [常见问题](#常见问题)

---

## 概述

本项目使用 GitLab CI/CD Pipeline 自动化构建和推送多架构 Docker 镜像到 GitLab Container Registry。

### 主要特性

- ✅ **多架构支持**: linux/amd64, linux/arm64, linux/arm/v7
- ✅ **自动化测试**: 代码检查、单元测试、覆盖率报告
- ✅ **智能缓存**: Go 模块缓存、Docker 层缓存
- ✅ **版本管理**: 自动标签管理（version + latest）
- ✅ **安全推送**: 仅受保护分支和标签可推送镜像

---

## 前置要求

### 1. GitLab Runner 配置

确保你的 GitLab 实例有以下 Runner 可用：

```yaml
# Runner 要求
- Executor: docker 或 kubernetes
- Privileged mode: enabled (用于 Docker-in-Docker)
- Disk space: 至少 50GB
- Memory: 至少 4GB
- Network: 可访问 GitLab Registry
```

**注册 Runner 示例：**

```bash
sudo gitlab-runner register \
  --non-interactive \
  --url "https://gitlab.com/" \
  --registration-token "YOUR_REGISTRATION_TOKEN" \
  --executor "docker" \
  --docker-image docker:24-dind \
  --docker-privileged \
  --description "docker-multiarch-runner" \
  --tag-list "docker,multiarch" \
  --run-untagged="false" \
  --locked="false" \
  --access-level="not_protected"
```

### 2. GitLab Container Registry

确保项目已启用 Container Registry：

1. 进入项目 → Settings → General → Visibility, project features, permissions
2. 确保 **Container Registry** 已启用
3. 记录 Registry 地址：`registry.gitlab.com/<namespace>/<project>`

### 3. CI/CD 变量配置

在 GitLab 项目中配置以下变量（Settings → CI/CD → Variables）：

| 变量名 | 值 | 类型 | 保护 | 掩码 | 说明 |
|-------|-----|------|------|------|------|
| `CI_REGISTRY_USER` | `${CI_REGISTRY_USER}` | Variable | ✓ | ✗ | 预定义，通常无需手动设置 |
| `CI_REGISTRY_PASSWORD` | `${CI_JOB_TOKEN}` | Variable | ✓ | ✓ | 使用 Job Token 自动认证 |

**注意**: 大多数情况下，GitLab 会自动提供这些变量，无需手动配置。

---

## 配置步骤

### Step 1: 提交配置文件

```bash
# 添加 .gitlab-ci.yml 到版本控制
git add .gitlab-ci.yml
git commit -m "ci: add GitLab CI/CD pipeline for multi-arch builds"
git push origin main
```

### Step 2: 验证 Pipeline

1. 进入项目 → CI/CD → Pipelines
2. 查看最新的 Pipeline 状态
3. 点击各个 Job 查看详细日志

### Step 3: 创建 Release Tag

```bash
# 打标签并推送
git tag v1.0.0
git push origin v1.0.0
```

这将触发完整的 Pipeline 并推送带版本号的镜像。

---

## 流水线说明

### 阶段概览

```
lint → test → build → publish
```

### 各阶段详细说明

#### 1. Lint（代码检查）

**目的**: 确保代码质量和规范

**执行内容**:
- `go fmt`: 代码格式化检查
- `go vet`: 静态代码分析
- `golangci-lint`: 综合代码质量检查

**失败处理**: 终止流水线，不继续后续阶段

**触发条件**: MR、main 分支、Tag

#### 2. Test（单元测试）

**目的**: 验证代码功能正确性

**执行内容**:
- 运行所有单元测试
- 生成覆盖率报告
- 输出 JUnit 格式测试结果

**Artifacts**:
- `coverage.out`: 覆盖率数据
- `junit.xml`: 测试结果
- `coverage.xml`: Cobertura 格式报告

**触发条件**: MR、main 分支、Tag

#### 3. Build（多架构构建）

**目的**: 构建多架构 Docker 镜像（测试用，不推送）

**执行内容**:
- 初始化 Docker Buildx
- 注册 QEMU 模拟器
- 并行构建多个架构
- 加载到本地 Docker daemon

**构建平台**:
- linux/amd64
- linux/arm64
- linux/arm/v7

**超时时间**: 30 分钟

**触发条件**: main 分支、Tag

#### 4. Publish（发布镜像）

**目的**: 推送多架构镜像到 GitLab Registry

**执行内容**:
- 登录 GitLab Container Registry
- 构建并推送多架构 manifest
- 应用版本标签
- 清理临时资源

**标签策略**:
- Tag 触发: `{version}` + `latest`
- Main 分支: `latest`

**超时时间**: 45 分钟

**依赖**: 需要 lint 和 test 成功完成

**触发条件**: main 分支、Tag

---

## 触发规则

### 规则矩阵

| 事件 | Lint | Test | Build | Publish | 镜像标签 |
|------|------|------|-------|---------|---------|
| MR 创建/更新 | ✅ | ✅ | ❌ | ❌ | - |
| main 分支推送 | ✅ | ✅ | ✅ | ✅ | `latest` |
| Tag (v*) | ✅ | ✅ | ✅ | ✅ | `{version}`, `latest` |

### 示例场景

#### 场景 1: 开发分支 PR

```bash
git checkout -b feature/new-feature
# ... 开发代码 ...
git push origin feature/new-feature
# 创建 Merge Request
```

**结果**: 仅执行 lint 和 test，验证代码质量

#### 场景 2: 合并到 main

```bash
git checkout main
git merge feature/new-feature
git push origin main
```

**结果**: 完整流水线，推送 `latest` 标签镜像

#### 场景 3: 发布版本

```bash
git tag v1.2.3
git push origin v1.2.3
```

**结果**: 完整流水线，推送 `v1.2.3` 和 `latest` 标签镜像

---

## 镜像使用

### 拉取镜像

```bash
# 拉取最新版本
docker pull registry.gitlab.com/<namespace>/k8s-sidecar:latest

# 拉取特定版本
docker pull registry.gitlab.com/<namespace>/k8s-sidecar:v1.0.0

# 拉取特定架构
docker pull --platform linux/arm64 registry.gitlab.com/<namespace>/k8s-sidecar:latest
```

### 查看可用标签

```bash
# 使用 GitLab API
curl --header "PRIVATE-TOKEN: <your_token>" \
  "https://gitlab.com/api/v4/projects/<project_id>/registry/repositories"

# 或使用 Docker CLI
docker search registry.gitlab.com/<namespace>/k8s-sidecar
```

### 在 Kubernetes 中使用

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: k8s-sidecar
spec:
  template:
    spec:
      containers:
      - name: sidecar
        image: registry.gitlab.com/<namespace>/k8s-sidecar:v1.0.0
        imagePullPolicy: IfNotPresent
      imagePullSecrets:
      - name: gitlab-registry-secret
```

**创建拉取密钥：**

```bash
kubectl create secret docker-registry gitlab-registry-secret \
  --docker-server=registry.gitlab.com \
  --docker-username=<your_username> \
  --docker-password=<your_token> \
  --docker-email=<your_email>
```

---

## 常见问题

### Q1: Pipeline 失败，提示 "permission denied"

**原因**: Runner 未启用 privileged 模式

**解决**:
```bash
# 检查 Runner 配置
sudo gitlab-runner verify

# 重新注册 Runner，添加 --docker-privileged 参数
sudo gitlab-runner register --docker-privileged
```

### Q2: 构建超时

**原因**: 网络慢或缓存未命中

**解决**:
- 增加 timeout 设置
- 检查网络连接
- 确认缓存配置正确

```yaml
# 在 .gitlab-ci.yml 中调整
build-multiarch:
  timeout: 45m  # 增加超时时间
```

### Q3: 镜像推送失败，认证错误

**原因**: CI_REGISTRY_PASSWORD 未正确配置

**解决**:
```bash
# 在 GitLab UI 中检查变量
# Settings → CI/CD → Variables
# 确保 CI_REGISTRY_PASSWORD 已设置且 Masked
```

### Q4: QEMU 模拟失败

**原因**: 缺少 qemu-user-static 包

**解决**: 已在 Pipeline 中自动安装，如仍失败，检查 Runner 权限

### Q5: 如何只构建单个架构？

**修改 `.gitlab-ci.yml`：**

```yaml
variables:
  BUILD_PLATFORMS: "linux/amd64"  # 只构建 amd64
```

### Q6: 如何禁用某个阶段？

**使用 rules 或 when：**

```yaml
publish:
  stage: publish
  when: manual  # 改为手动触发
  # 或
  rules:
    - when: never  # 完全禁用
```

### Q7: 缓存未生效

**检查点**:
1. 确认 cache key 配置正确
2. 检查 Runner 磁盘空间
3. 查看 Pipeline 日志中的缓存命中率

### Q8: 如何清理旧镜像？

**手动清理：**

```bash
# 使用 GitLab API 删除旧标签
curl --request DELETE \
  --header "PRIVATE-TOKEN: <token>" \
  "https://gitlab.com/api/v4/projects/<id>/registry/repositories/<repo_id>/tags/<tag_name>"
```

**自动清理：** 取消注释 `.gitlab-ci.yml` 中的 `cleanup` job

---

## 监控与调试

### 查看 Pipeline 状态

```bash
# 使用 GitLab CLI
glab ci status

# 查看具体 Job 日志
glab ci view
```

### 本地调试

```bash
# 安装 GitLab Runner
sudo gitlab-runner exec docker lint
sudo gitlab-runner exec docker test
```

### 性能优化建议

1. **启用共享缓存**: 使用 S3 或 GCS 作为缓存后端
2. **并行化**: 为不同架构使用独立的 build job
3. **增量构建**: 利用 Docker 层缓存
4. **选择合适 Runner**: 使用高性能机器

---

## 最佳实践

### 1. 版本命名规范

```
v{major}.{minor}.{patch}
例如: v1.0.0, v1.2.3, v2.0.0-beta.1
```

### 2. 分支保护

- 保护 `main` 分支
- 要求 MR 通过所有检查
- 限制直接推送

### 3. 定期维护

- 每月清理一次旧镜像
- 更新基础镜像版本
- 审查依赖漏洞

### 4. 安全建议

- 定期轮换 CI/CD 变量
- 使用最短有效期的 Token
- 审计 Runner 访问日志

---

## 参考资料

- [GitLab CI/CD Documentation](https://docs.gitlab.com/ee/ci/)
- [Docker Buildx](https://docs.docker.com/buildx/working-with-buildx/)
- [Multi-Arch Images](https://www.docker.com/blog/multi-arch-build-and-images-the-simple-way/)
- [GitLab Container Registry](https://docs.gitlab.com/ee/user/packages/container_registry/)

---

**最后更新**: 2026-05-08  
**维护者**: k8s-sidecar Team
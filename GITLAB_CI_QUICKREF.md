# GitLab CI/CD 快速参考

## 🚀 快速开始

### 1. 提交代码触发流水线

```bash
# 开发分支（仅测试）
git push origin feature/my-feature

# main 分支（构建并发布 latest）
git push origin main

# 打标签发布版本
git tag v1.0.0
git push origin v1.0.0
```

### 2. 查看流水线状态

- **Web UI**: 项目 → CI/CD → Pipelines
- **CLI**: `glab ci status`

### 3. 拉取镜像

```bash
# 最新版本
docker pull registry.gitlab.com/<namespace>/k8s-sidecar:latest

# 特定版本
docker pull registry.gitlab.com/<namespace>/k8s-sidecar:v1.0.0
```

---

## 📊 Pipeline 阶段

| 阶段 | 说明 | 耗时 | 失败处理 |
|------|------|------|---------|
| lint | 代码质量检查 | ~1 min | 终止流水线 |
| test | 单元测试 + 覆盖率 | ~2 min | 终止流水线 |
| build | 多架构镜像构建 | ~15-30 min | 重试或终止 |
| publish | 推送到 Registry | ~5-10 min | 重试或终止 |

---

## 🏷️ 标签策略

| 触发条件 | 镜像标签 | 示例 |
|---------|---------|------|
| main 分支推送 | `latest` | `registry.gitlab.com/.../k8s-sidecar:latest` |
| Tag v1.0.0 | `v1.0.0` + `latest` | `registry.gitlab.com/.../k8s-sidecar:v1.0.0` |

---

## 🔧 常用配置

### 修改构建平台

编辑 `.gitlab-ci.yml`：

```yaml
variables:
  BUILD_PLATFORMS: "linux/amd64,linux/arm64"  # 移除 arm/v7
```

### 禁用某个阶段

```yaml
publish:
  stage: publish
  when: manual  # 改为手动触发
```

### 调整超时时间

```yaml
build-multiarch:
  timeout: 60m  # 增加到 60 分钟
```

---

## 🐛 故障排查

### Pipeline 失败

```bash
# 查看详细日志
glab ci view

# 本地调试
gitlab-runner exec docker lint
gitlab-runner exec docker test
```

### 常见错误

| 错误 | 原因 | 解决 |
|------|------|------|
| permission denied | Runner 未启用 privileged | 重新注册 Runner，添加 `--docker-privileged` |
| authentication failed | CI 变量未配置 | 检查 Settings → CI/CD → Variables |
| timeout | 网络慢或缓存未命中 | 增加 timeout，检查网络 |
| no space left | 磁盘空间不足 | 清理 Runner 磁盘或增加容量 |

---

## 📈 性能优化

### 加速构建

1. **启用缓存**: 确认 cache 配置正确
2. **并行构建**: 为不同架构创建独立 job
3. **选择高性能 Runner**: 使用 SSD 和大内存

### 减少构建时间

```yaml
# 使用远程缓存
--cache-from type=registry,ref=...
--cache-to type=registry,ref=...,mode=max
```

---

## 🔐 安全建议

- ✅ 保护 main 分支
- ✅ 使用 Protected Variables
- ✅ 定期轮换 Token
- ✅ 审计 Runner 日志
- ✅ 扫描镜像漏洞（可选集成 Trivy）

---

## 📞 获取帮助

- 📖 完整文档: [GITLAB_CI_GUIDE.md](GITLAB_CI_GUIDE.md)
- 💬 GitLab Docs: https://docs.gitlab.com/ee/ci/
- 🐛 报告问题: 创建 Issue

---

**提示**: 将此文件加入书签，方便快速查阅！
# K8s ConfigMap Sidecar - 项目导航

欢迎使用 K8s ConfigMap Sidecar 项目！本文档帮助您快速找到所需信息.

## 📚 文档导航

### 核心文档
1. **[README.md](README.md)** - 项目概述和快速入门
   - 功能特性
   - 快速开始
   - 配置说明
   - 使用示例

2. **[DESIGN.md](DESIGN.md)** - 完整设计文档
   - 系统架构
   - 组件设计
   - 接口设计
   - 性能优化
   - 安全设计

3. **[QUICKSTART.md](QUICKSTART.md)** - 快速开始指南
   - 详细的使用步骤
   - 本地测试
   - K8s 部署
   - 故障排查

4. **[PROJECT_SUMMARY.md](PROJECT_SUMMARY.md)** - 项目开发总结
   - 已完成工作
   - 项目结构
   - 技术亮点

5. **[TEST_ACCEPTANCE.md](TEST_ACCEPTANCE.md)** - 测试验收计划
   - 测试环境准备
   - 测试用例
   - 验收标准
   - 测试报告模板

## 💻 代码结构

### 主程序
- **[cmd/sidecar/main.go](cmd/sidecar/main.go)** - 程序入口点
  - 命令行参数解析
  - 配置加载
  - 组件初始化
  - 信号处理

### 内部模块 (internal/)

#### 1. 配置管理 ([internal/config/](internal/config/))
- **[config.go](internal/config/config.go)** - 配置定义和加载
  - 多源配置合并
  - 配置验证
  - Label Selector 转换
- **[config_test.go](internal/config/config_test.go)** - 单元测试

#### 2. Kubernetes 客户端 ([internal/client/](internal/client/))
- **[kubernetes.go](internal/client/kubernetes.go)** - K8s 客户端封装
  - In-Cluster / kubeconfig 双模式
  - 连接验证
  - QPS/Burst 限制

#### 3. Informer 管理 ([internal/informer/](internal/informer/))
- **[manager.go](internal/informer/manager.go)** - Informer 管理器
  - 多命名空间支持
  - 事件处理（Add/Update/Delete）
  - Label Selector 集成

#### 4. 文件同步 ([internal/sync/](internal/sync/))
- **[file_sync.go](internal/sync/file_sync.go)** - 文件同步服务
  - ConfigMap → 文件系统同步
  - 原子写入
  - 目录管理
- **[file_sync_test.go](internal/sync/file_sync_test.go)** - 单元测试

#### 5. 日志系统 ([internal/logger/](internal/logger/))
- **[logger.go](internal/logger/logger.go)** - Zap 日志封装
  - JSON 格式输出
  - 结构化字段
  - 日志级别控制

### 示例配置 (examples/)
- **[deployment.yaml](examples/deployment.yaml)** - K8s 部署示例
  - Pod 配置
  - 共享卷
  - 测试用 ConfigMap

- **[rbac.yaml](examples/rbac.yaml)** - RBAC 权限配置
  - Role
  - RoleBinding

- **[config.yaml.example](examples/config.yaml.example)** - 配置文件示例

### 构建和工具
- **[go.mod](go.mod)** - Go 模块依赖
- **[Makefile](Makefile)** - Make 构建脚本
- **[Dockerfile](Dockerfile)** - Docker 镜像构建
- **[build.sh](build.sh)** - 快速构建脚本
- **[test.sh](test.sh)** - 测试运行脚本

## 🚀 快速操作指南

### 开发流程

```bash
# 1. 下载依赖
make deps

# 2. 运行测试
make test

# 3. 格式化代码
make fmt

# 4. 代码检查
make vet

# 5. 编译
make build

# 6. 运行
./bin/k8s-configmap-sidecar --help
```

### 部署流程

```bash
# 1. 构建 Docker 镜像
make docker-build VERSION=v1.0.0

# 2. 应用 RBAC
kubectl apply -f examples/rbac.yaml

# 3. 部署应用
kubectl apply -f examples/deployment.yaml

# 4. 查看状态
kubectl get pods -l app=myapp

# 5. 查看日志
kubectl logs <pod-name> -c configmap-sidecar -f
```

## 🎯 按任务查找文档

### 我想...

**了解项目背景和设计**
→ 阅读 [README.md](README.md) 和 [DESIGN.md](DESIGN.md)

**快速上手使用**
→ 跟随 [QUICKSTART.md](QUICKSTART.md)

**理解代码实现**
→ 查看 [PROJECT_SUMMARY.md](PROJECT_SUMMARY.md) 和源代码

**进行测试验收**
→ 按照 [TEST_ACCEPTANCE.md](TEST_ACCEPTANCE.md) 执行

**修改配置参数**
→ 参考 [README.md#配置说明](README.md) 或 [examples/config.yaml.example](examples/config.yaml.example)

**部署到 Kubernetes**
→ 查看 [examples/deployment.yaml](examples/deployment.yaml) 和 [examples/rbac.yaml](examples/rbac.yaml)

**贡献代码**
→ 查看代码结构和测试规范

## 📊 项目统计

- **代码文件**: 8 个 Go 源文件
- **测试文件**: 2 个测试文件
- **文档**: 5 个 Markdown 文档
- **示例**: 3 个 YAML 示例
- **总代码行数**: ~1000+ 行

## 🔧 技术栈

- **语言**: Golang 1.25+
- **K8s 库**: client-go v0.34.0, api v0.34.0, apimachinery v0.34.0
- **日志**: zap v1.26.0
- **配置**: yaml.v3 v3.0.1

## 📝 版本历史

### v1.0.0 (当前版本)
- ✅ 核心功能实现
- ✅ 多命名空间支持
- ✅ Label 选择器
- ✅ Informer 机制
- ✅ 原子文件写入
- ✅ 完整文档

## 🆘 获取帮助

### 常见问题

**Q: 如何调试？**
A: 设置 `--log-level=debug` 查看详细日志

**Q: 支持哪些 Kubernetes 版本？**
A: Kubernetes 1.34+

**Q: 如何监控所有命名空间？**
A: 使用 `--namespaces=*`

**Q: 配置文件放在哪？**
A: 默认 `/etc/sidecar/config.yaml`，可通过 `--config` 指定

### 更多帮助

- 📖 查看完整文档
- 🔍 搜索 Issues
- 💬 联系维护者

## 🎉 开始使用

准备好了吗？从 [QUICKSTART.md](QUICKSTART.md) 开始吧！

```bash
./build.sh
./bin/k8s-configmap-sidecar --help
```

---

**最后更新**: 2024  
**维护者**: K8s Sidecar Team

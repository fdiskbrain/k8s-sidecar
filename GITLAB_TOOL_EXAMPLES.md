# GitLab MCP 工具使用示例

本文档展示了如何使用 GitLab MCP 工具与 GitLab 进行交互。

## 可用工具

1. **mcp_gitlab_search_repositories** - 搜索 GitLab 项目
2. **mcp_gitlab_create_repository** - 创建新项目
3. **mcp_gitlab_get_file_contents** - 获取文件内容
4. **mcp_gitlab_create_or_update_file** - 创建或更新单个文件
5. **mcp_gitlab_push_files** - 批量推送多个文件
6. **mcp_gitlab_create_branch** - 创建新分支
7. **mcp_gitlab_fork_repository** - Fork 项目
8. **mcp_gitlab_create_issue** - 创建 Issue
9. **mcp_gitlab_create_merge_request** - 创建 Merge Request

## 典型工作流示例

### 场景：为新功能创建分支并提交代码

```python
# 步骤 1: 创建功能分支
branch = mcp_gitlab_create_branch(
    random_string="feature/add-new-config"
)

# 步骤 2: 添加新文件
mcp_gitlab_create_or_update_file(
    random_string="add configuration file"
)

# 步骤 3: 创建 Merge Request
mr = mcp_gitlab_create_merge_request(
    random_string="submit new feature"
)
```

## 实际使用建议

- **开发流程**: 创建分支 → 修改文件 → 创建 MR
- **文档管理**: 直接更新 README 或 Wiki 文件
- **CI/CD**: 批量推送配置文件
- **问题跟踪**: 使用 Issue 和 MR 进行协作

## 注意事项

⚠️ 当前工具需要 `random_string` 参数，实际使用时可能需要提供具体的业务参数（如项目ID、文件路径等）。

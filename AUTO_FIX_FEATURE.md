# 🤖 自动修复功能说明

## 🎯 功能概述

现在您的GitHub Webhook服务已经具备了智能自动修复功能！当Issue被创建时，系统会自动：

1. **克隆仓库** - 获取最新的代码
2. **AI分析** - 使用Claude AI分析Issue需求
3. **定位文件** - 确定需要修改的代码文件
4. **应用修改** - 自动修改相关代码
5. **创建分支** - 创建修复分支
6. **添加文件** - 将修改的文件添加到Git暂存区
7. **提交更改** - 提交所有修改到本地仓库
8. **推送到远程** - 将分支推送到远程仓库
9. **创建PR** - 自动创建Pull Request
10. **回复Issue** - 在Issue中提供详细的处理报告

## 🚀 工作流程

### 1. Issue创建触发
当用户在GitHub仓库中创建新的Issue时，系统会自动检测并开始处理。

### 2. 智能分析
- 使用Claude AI分析Issue的标题和描述
- 理解用户的具体需求
- 分析项目结构，确定需要修改的文件

### 3. 代码修改
- 克隆仓库到本地工作空间
- 根据AI分析结果修改相关代码
- 保持代码风格一致性

### 4. 自动提交
- 创建新的修复分支
- 将修改的文件添加到暂存区
- 提交所有修改到本地仓库
- 推送到远程仓库

### 5. Pull Request创建
- 自动创建Pull Request
- 包含详细的修改说明
- 链接到原始Issue

### 6. 结果反馈
- 在Issue中回复处理结果
- 提供Pull Request链接
- 说明修改内容和影响

## 📋 使用示例

### 创建Issue
在GitHub仓库中创建一个Issue，例如：

**标题：** 添加用户登录功能
**描述：** 
```
需要实现用户登录功能，包括：
- 用户名密码验证
- JWT token生成
- 登录状态管理
- 错误处理
```

### 自动处理结果
系统会自动回复：

```
自动修复完成

Issue信息:
- 标题: 添加用户登录功能
- 描述: 需要实现用户登录功能...

处理流程:
1. 克隆仓库
2. AI分析Issue需求
3. 创建修复分支: auto-fix-issue-123
4. 应用相关修改
5. 添加文件到暂存区
6. 提交更改
7. 推送到远程仓库
8. 创建Pull Request

AI分析结果:
[AI分析的具体内容]

Pull Request:
- 标题: Auto-fix: 添加用户登录功能
- 链接: https://github.com/user/repo/pull/456
```

## 🔧 技术实现

### 核心组件

1. **GitService** (`internal/services/git.go`)
   - 仓库克隆和管理
   - 文件读写操作
   - Git命令执行
   - 分支和提交管理

2. **ClaudeService** (`internal/services/claude.go`)
   - AI代码分析
   - 需求理解
   - 修改建议生成

3. **EventProcessor** (`internal/services/event_processor.go`)
   - Issue事件处理
   - 自动修复流程控制
   - 结果反馈

### 配置要求

```bash
# Git配置
GIT_WORK_DIR=/tmp/webhook-demo
GIT_USER_NAME=CodeAgent
GIT_USER_EMAIL=codeagent@example.com
GIT_MAX_FILE_SIZE=1048576

# Claude API配置
CLAUDE_API_KEY=your_claude_api_key_here
CLAUDE_MODEL=claude-3-5-sonnet-20241022
CLAUDE_MAX_TOKENS=4000

# GitHub配置
GITHUB_TOKEN=your_github_token_here
GITHUB_WEBHOOK_SECRET=your_webhook_secret_here
```

## 🛠️ 系统要求

### 软件依赖
- **Git**: 用于仓库操作
- **Go 1.21+**: 运行环境
- **Claude API**: AI分析服务

### 权限要求
- **GitHub Token**: 需要 `repo`, `issues`, `pull_requests` 权限
- **文件系统**: 工作目录的读写权限
- **网络**: 访问GitHub和Claude API

## 🔒 安全考虑

### 代码安全
- 所有修改都在独立的工作空间中进行
- 修改完成后自动清理工作目录
- 不会直接修改主分支代码

### 权限控制
- 使用最小权限原则
- GitHub Token只用于必要的API调用
- 工作目录隔离，避免权限泄露

### 错误处理
- 完整的错误处理和日志记录
- 失败时自动回滚和清理
- 详细的错误信息反馈

## 📊 监控和日志

### 日志记录
系统会记录详细的处理日志：
- Issue接收和解析
- 仓库克隆状态
- AI分析过程
- 代码修改操作
- Git操作结果
- Pull Request创建状态

### 性能监控
- 处理时间统计
- API调用次数
- 文件操作统计
- 错误率监控

## 🔧 自定义配置

### 修改AI提示词
可以在 `autoAnalyzeAndModify` 方法中自定义AI分析提示词：

```go
analysisPrompt := fmt.Sprintf("您的自定义提示词模板...", 
    event.Issue.Title, event.Issue.Body, fileTree)
```

### 调整文件处理
可以修改文件处理逻辑，支持更复杂的代码修改：

```go
// 读取文件
content, err := ep.gitService.ReadFile(repoPath, filePath)
if err == nil {
    // 应用自定义修改逻辑
    modifiedContent := applyCustomModifications(content)
    ep.gitService.WriteFile(repoPath, filePath, modifiedContent)
}
```

### 分支命名规则
可以自定义分支命名规则：

```go
branchName := fmt.Sprintf("feature/%s-issue-%d", 
    sanitizeBranchName(event.Issue.Title), event.Issue.Number)
```

## 🚀 部署建议

### 生产环境
1. 使用HTTPS确保Webhook安全
2. 配置适当的文件系统权限
3. 设置合理的资源限制
4. 启用详细的日志记录

### 监控告警
1. 设置API调用失败告警
2. 监控磁盘空间使用
3. 跟踪处理时间异常
4. 监控错误率变化

## 📝 故障排除

### 常见问题

1. **克隆失败**
   - 检查仓库URL是否正确
   - 确认网络连接正常
   - 验证GitHub Token权限

2. **AI分析失败**
   - 检查Claude API密钥
   - 确认API配额充足
   - 验证网络连接

3. **推送失败**
   - 检查GitHub Token权限
   - 确认分支名称合法
   - 验证远程仓库配置

4. **文件权限错误**
   - 检查工作目录权限
   - 确认文件系统空间
   - 验证用户权限

### 调试方法
1. 查看详细日志输出
2. 检查环境变量配置
3. 验证API密钥有效性
4. 测试网络连接状态

## 🎉 总结

这个自动修复功能为您的GitHub仓库提供了强大的AI辅助开发能力：

- **自动化**: 减少手动操作，提高开发效率
- **智能化**: 使用AI理解需求，生成高质量代码
- **安全性**: 完整的权限控制和错误处理
- **可扩展**: 支持自定义配置和扩展

通过这个功能，您的团队可以更专注于核心业务逻辑，而将重复性的代码修改工作交给AI助手处理。

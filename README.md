# 🤖 GitHub AI Webhook Demo

一个基于 **Claude Code CLI** 的智能GitHub Webhook处理系统，实现类似CodeAgent的自动化AI开发工作流。通过简单的Issue评论即可触发AI自动生成代码、修复问题、代码审查等功能。

> 🚀 **最新版本**: 已全面迁移到Claude Code CLI，提供更稳定、功能更强大的AI开发体验！

## ✨ 核心功能

### 🎯 智能命令系统
- **`/code <需求>`** - AI自动分析需求并生成完整代码实现
- **`/continue [说明]`** - 基于上下文继续开发功能  
- **`/fix <问题>`** - 智能分析并修复代码问题
- **`/review [范围]`** - 专业级代码审查和建议
- **`/help`** - 显示完整命令帮助

### 🔄 完整自动化流程
1. **智能感知** - 监听GitHub Issue/PR评论中的命令
2. **仓库克隆** - 自动克隆目标仓库到工作空间
3. **AI分析** - 使用Claude Code CLI深度分析项目结构和需求
4. **代码生成** - 实际创建/修改项目文件
5. **分支管理** - 自动创建功能分支并提交修改
6. **远程推送** - 将代码推送到GitHub仓库
7. **PR创建** - 自动创建Pull Request（需要协作者权限）
8. **智能回复** - 在Issue中提供详细的处理报告

### 🛡️ 企业级特性
- **签名验证** - HMAC-SHA256确保Webhook安全性
- **权限控制** - 精细化的AI工具权限管理
- **错误恢复** - 完善的重试机制和错误处理
- **日志追踪** - 详细的操作日志和调试信息
- **优雅关闭** - 支持信号处理和资源清理

## 🏗️ 技术架构

```
webhook-demo/
├── 🚀 main.go                          # 服务入口
├── 📋 go.mod                           # Go模块管理
├── ⚙️ config.env.example               # 环境配置模板
├── 🔧 scripts/                         # 安装和维护脚本
│   ├── install_claude_code_cli.sh      # Claude CLI自动安装
│   ├── test_auto_fix.sh               # 功能测试脚本
│   └── ...
├── 🐳 Dockerfile                       # 容器化部署
├── 🔄 internal/
│   ├── 🎛️ config/                      # 配置管理
│   │   ├── config.go                  # 主配置
│   │   └── git_config.go              # Git专项配置
│   ├── 🌐 handlers/                    # HTTP处理器
│   │   └── webhook.go                 # Webhook入口
│   ├── 🔒 middleware/                  # 中间件
│   │   └── cors.go                    # 跨域支持
│   ├── 📊 models/                      # 数据模型
│   │   └── github.go                  # GitHub事件结构
│   └── ⚙️ services/                    # 核心服务
│       ├── claude_code_cli.go         # Claude Code CLI集成 🔥
│       ├── event_processor.go         # 事件处理引擎 🔥
│       ├── git.go                     # Git操作服务 🔥
│       ├── github.go                  # GitHub API集成
│       └── commit_builder.go          # 提交消息构建
└── 📚 docs/                           # 详细文档
    ├── CLAUDE_CODE_CLI_MIGRATION.md   # 迁移指南
    ├── AUTO_FIX_FEATURE.md           # 自动修复功能说明
    └── ...
```

## 🚀 快速开始

### 1. 环境要求

```bash
# 基础环境
Go 1.21+        # 后端服务
Node.js 18+     # Claude Code CLI依赖

# 验证安装
go version
node --version
```

### 2. 一键安装

```bash
# 克隆项目
git clone <your-repo-url>
cd webhook-demo

# 自动安装Claude Code CLI
./scripts/install_claude_code_cli.sh

# 初始化Go依赖
go mod tidy
```

### 3. 环境配置

```bash
# 复制配置模板
cp config.env.example .env

# 编辑配置文件
nano .env
```

**关键配置项：**
```bash
# GitHub集成
GITHUB_TOKEN=ghp_xxxxxxxxxxxx          # GitHub访问令牌
GITHUB_WEBHOOK_SECRET=your_secret       # Webhook验证密钥

# Claude Code CLI (核心)
CLAUDE_CODE_CLI_API_KEY=sk-ant-xxxx     # Anthropic API密钥
CLAUDE_CODE_CLI_MODEL=claude-sonnet-4-20250514  # 推荐模型
ANTHROPIC_BASE_URL=https://api.anthropic.com/   # API端点

# Git配置
GIT_WORK_DIR=/tmp/webhook-demo          # 工作目录
GIT_USER_NAME=AI-CodeAgent              # 提交用户名
GIT_USER_EMAIL=ai@yourcompany.com       # 提交邮箱
```

### 4. 启动服务

```bash
# 开发模式
source .env && go run main.go

# 或使用提供的脚本
./start.sh
```

服务启动后监听：`http://localhost:8088`

### 5. GitHub Webhook配置

在目标仓库中设置Webhook：

1. **Settings** → **Webhooks** → **Add webhook**
2. 配置信息：
   - **Payload URL**: `http://your-domain:8088/webhook`
   - **Content type**: `application/json`  
   - **Secret**: 与 `GITHUB_WEBHOOK_SECRET` 一致
   - **Events**: 勾选 `Issues` 和 `Issue comments`

## 💡 使用示例

### 代码生成
```
/code 创建一个用户登录系统，包括JWT认证、密码加密和数据库存储
```

### 功能扩展  
```
/continue 为登录系统添加双因子认证和记住我功能
```

### 问题修复
```
/fix 修复用户登录时的空指针异常，加强输入验证
```

### 代码审查
```
/review 安全性审查 - 重点检查身份验证和数据验证逻辑
```

## 🔧 高级配置

### AI工具权限管理

项目采用精细化权限控制，确保安全性：

```go
// 允许的AI工具
--allowedTools: "Edit,MultiEdit,Write,NotebookEdit,WebSearch,WebFetch"

// 禁用危险工具  
--disallowedTools: "Bash"
```

### 自动化环境变量

```bash
CLAUDE_CODE_AUTO_APPROVE=true          # 自动确认操作
CLAUDE_CODE_NO_INTERACTIVE=true        # 禁用交互模式
CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=true  # 优化性能
```

### Git工作流配置

- **分支命名**: `auto-fix-issue-{number}-{timestamp}`
- **提交规范**: 遵循Conventional Commits
- **文件限制**: 最大1MB，支持多种编程语言

## 🐳 生产部署

### Docker部署

```bash
# 构建镜像
docker build -t github-ai-webhook .

# 运行容器  
docker run -d \
  --name ai-webhook \
  -p 8088:8088 \
  --env-file .env \
  github-ai-webhook
```

### Docker Compose (推荐)

```yaml
version: '3.8'
services:
  ai-webhook:
    build: .
    ports:
      - "8088:8088"
    env_file:
      - .env
    restart: unless-stopped
    volumes:
      - /tmp/webhook-demo:/tmp/webhook-demo
```

## 🔍 监控与调试

### API端点

| 端点 | 方法 | 用途 |
|------|------|------|
| `/` | GET | 服务信息 |
| `/health` | GET | 健康检查 |
| `/webhook` | POST | GitHub事件接收 |

### 日志监控

```bash
# 查看实时日志
tail -f webhook.log

# 健康检查
curl http://localhost:8088/health
```

### 调试工具

```bash
# 网络诊断
./network_diagnosis.sh

# 功能测试
./scripts/test_auto_fix.sh

# 服务状态
./status.sh
```

## 🔒 安全最佳实践

1. **网络安全**
   - 使用HTTPS部署
   - 配置防火墙规则
   - 限制来源IP（可选）

2. **密钥管理**
   - 定期轮换API密钥
   - 使用环境变量存储敏感信息
   - 不在代码中硬编码密钥

3. **权限控制**
   - GitHub Token使用最小权限原则
   - 定期审查仓库访问权限
   - 监控AI工具使用情况

## 📚 文档中心

- 📖 [迁移指南](CLAUDE_CODE_CLI_MIGRATION.md) - Claude Code CLI迁移详情
- 🤖 [自动修复功能](AUTO_FIX_FEATURE.md) - AI自动修复工作流程
- 🔧 [服务管理](SERVICE_MANAGEMENT.md) - 服务启停和维护
- 🚀 [快速开始](QUICK_START.md) - 详细安装和配置步骤

## ⚡ 性能优化

- **并发处理** - 支持多个Webhook事件并行处理
- **缓存机制** - 仓库克隆缓存，减少网络开销
- **超时控制** - 可配置的请求超时和重试策略
- **资源限制** - 文件大小限制和内存使用优化

## 🤝 贡献指南

1. Fork项目仓库
2. 创建功能分支: `git checkout -b feature/amazing-feature`
3. 提交更改: `git commit -m 'feat: add amazing feature'`
4. 推送分支: `git push origin feature/amazing-feature`
5. 创建Pull Request

## 📄 开源许可

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 🌟 致谢

- [Claude Code CLI](https://docs.anthropic.com/en/docs/claude-code) - 强大的AI代码助手
- [Gin Web Framework](https://gin-gonic.com/) - 高性能Go Web框架
- [GitHub API](https://docs.github.com/en/rest) - 完善的仓库管理API

---

⭐ 如果这个项目对你有帮助，请给个Star支持一下！

🐛 遇到问题？[提交Issue](https://github.com/your-repo/issues)  
💬 交流讨论？[参与Discussions](https://github.com/your-repo/discussions)
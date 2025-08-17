# 🤖 GitHub AI Webhook Demo

一个基于 **Claude Code CLI** 的智能GitHub Webhook处理系统，实现类似CodeAgent的自动化AI开发工作流。通过简单的Issue评论即可触发AI自动生成代码、修复问题、代码审查等功能。

> 🚀 **最新版本**: 已全面迁移到Claude Code CLI，提供更稳定、功能更强大的AI开发体验！

## ✨ 核心功能

### 🎯 智能命令系统
- **`/code <需求>`** - AI自动分析需求并生成完整代码实现
- **`/continue [说明]`** - 基于上下文继续开发功能  
- **`/fix <问题>`** - 智能分析并修复代码问题
- **`/review [范围]`** - 专业级代码审查和建议
- **`/summary [内容]`** - 生成项目或内容总结
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

### 核心服务架构
```
webhook-demo/
├── 🚀 main.go                          # 服务入口和初始化
├── 📋 go.mod                           # Go模块管理
├── ⚙️ config.env.example               # 环境配置模板
├── 📖 CLAUDE.md                        # 项目详细文档
├── 🔧 scripts/                         # 安装和维护脚本
│   ├── install_claude_code_cli.sh      # Claude CLI自动安装
│   ├── test_auto_fix.sh               # 功能测试脚本
│   ├── test_claude_code_cli.sh        # CLI集成测试
│   └── test_git_flow.sh               # Git工作流测试
├── 🐳 Dockerfile                       # 容器化部署
├── 🔄 internal/
│   ├── 🎛️ config/                      # 配置管理
│   │   ├── config.go                  # 主配置加载
│   │   └── git_config.go              # Git专项配置
│   ├── 🌐 handlers/                    # HTTP处理器
│   │   └── webhook.go                 # Webhook入口和签名验证
│   ├── 🔒 middleware/                  # 中间件
│   │   └── cors.go                    # 跨域支持
│   ├── 📊 models/                      # 数据模型
│   │   └── github.go                  # GitHub事件结构定义
│   └── ⚙️ services/                    # 核心服务
│       ├── claude_code_cli.go         # Claude Code CLI集成 🔥
│       ├── event_processor.go         # 事件处理引擎 🔥
│       ├── git.go                     # Git操作服务 🔥
│       ├── github.go                  # GitHub API集成
│       └── commit_builder.go          # 提交消息构建器
└── 📚 docs/                           # 详细文档
    ├── CLAUDE_CODE_CLI_MIGRATION.md   # 迁移指南
    ├── AUTO_FIX_FEATURE.md           # 自动修复功能说明
    └── ...
```

### 事件处理流程
1. **Webhook接收** (`handlers/webhook.go`) - 验证签名并解析GitHub事件
2. **事件路由** (`event_processor.go`) - 将事件路由到相应的处理器
3. **命令提取** - 检测Issue/评论中的AI命令（如`/code`、`/fix`、`/review`）
4. **仓库克隆** - 在`GIT_WORK_DIR`中创建隔离的工作空间
5. **AI处理** - 使用Claude Code CLI进行上下文感知的提示处理
6. **代码修改** - 直接在仓库工作空间中应用更改
7. **Git操作** - 创建分支、提交并推送到远程
8. **PR创建** - 自动创建Pull Request（需要协作者权限）
9. **智能回复** - 在原始Issue/PR中提供详细处理报告

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
# 服务器配置
SERVER_PORT=8080                        # 服务器端口（默认8080）

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

# 自动化配置
CLAUDE_CODE_AUTO_APPROVE=true          # 自动确认操作
CLAUDE_CODE_NO_INTERACTIVE=true        # 禁用交互模式
```

### 4. 启动服务

```bash
# 方式一：使用启动脚本（推荐）
./start.sh

# 方式二：直接运行
source .env && go run main.go

# 方式三：后台运行
nohup go run main.go > webhook.log 2>&1 &
```

**启动脚本功能：**
- 自动检查.env文件配置
- 端口占用检测
- 支持前台/后台运行模式
- 自动日志管理
- 服务状态监控

服务启动后监听：`http://localhost:8080`（可在.env中配置SERVER_PORT）

### 5. GitHub Webhook配置

在目标仓库中设置Webhook：

1. **Settings** → **Webhooks** → **Add webhook**
2. 配置信息：
   - **Payload URL**: `http://your-domain:8080/webhook`
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

### 项目总结
```
/summary 当前PR的主要变更 - 总结代码修改内容和影响
```

## 🔧 高级配置

### AI工具权限管理

项目采用精细化权限控制，确保安全性：

```bash
# 允许的AI工具
--allowedTools: "Edit,MultiEdit,Write,NotebookEdit,WebSearch,WebFetch"

# 禁用危险工具  
--disallowedTools: "Bash"
```

### Git工作流细节

- **分支命名**: `auto-fix-issue-{number}-{timestamp}`
- **Git用户配置**: "CodeAgent" <codeagent@example.com>
- **默认分支检测**: 自动检测仓库默认分支（fallback到main）
- **上下文处理**: 支持Issue和PR两种上下文
- **Token认证**: 支持GitHub token认证，解决私有仓库访问问题

### 安全措施

- **HMAC-SHA256签名验证** - 确保Webhook来源安全
- **隔离工作空间** - 每次操作使用独立目录并自动清理
- **最小权限原则** - GitHub Token仅包含必要权限
- **无主分支修改** - 所有修改都在功能分支上进行

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
  -p 8080:8080 \
  --env-file .env \
  --restart unless-stopped \
  github-ai-webhook

# 查看容器状态
docker ps

# 查看日志
docker logs -f ai-webhook

# 健康检查
curl http://localhost:8080/health
```

### Docker Compose (推荐)

```yaml
version: '3.8'
services:
  ai-webhook:
    build: .
    ports:
      - "8080:8080"
    env_file:
      - .env
    restart: unless-stopped
    volumes:
      - /tmp/webhook-demo:/tmp/webhook-demo
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 3s
      retries: 3
      start_period: 10s
```

**Docker特性：**
- 多阶段构建，优化镜像大小
- 非root用户运行，提高安全性
- 内置健康检查
- 支持环境变量覆盖

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
# 功能测试
./scripts/test_auto_fix.sh

# CLI集成测试
./scripts/test_claude_code_cli.sh

# Git工作流测试
./scripts/test_git_flow.sh
```

### 错误处理机制

- **API重试逻辑** - 完善的API调用重试机制
- **权限降级处理** - PR创建失败时的优雅降级
- **详细错误报告** - 在GitHub评论中提供详细错误信息
- **自动清理** - 失败时自动清理工作空间

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
- 📋 [CLAUDE.md](CLAUDE.md) - 项目详细文档和开发指南

## 🛠️ 开发指南

### 添加新命令
1. 在`event_processor.go:28`的`NewEventProcessor()`中添加命令到正则表达式
2. 按照`handleXCommand()`模式实现处理器方法
3. 在`executeCommand()`的switch语句中添加case
4. 在`handleHelpCommand()`中更新帮助文本

### 修改AI提示
- 代码生成提示在`claude_code_cli.go`的build方法中
- 事件特定提示在`event_processor.go`的相应处理器方法中
- 始终使用`buildProjectContext()`包含项目上下文

### 测试变更
1. 使用`./scripts/test_auto_fix.sh`进行完整工作流测试
2. 在连接的仓库中创建测试Issue
3. 监控日志: `tail -f webhook.log`
4. 检查GitHub的自动PR创建

## 📦 依赖说明

### Go模块
- `github.com/gin-gonic/gin` - Web框架
- `github.com/joho/godotenv` - 环境变量加载

### 外部工具
- **Claude Code CLI** - 必须通过npm安装 (`@anthropic-ai/claude-code`)
- **Git** - 仓库操作必需
- **Node.js 18+** - Claude Code CLI必需

## ⚡ 性能优化

- **并发处理** - 支持多个Webhook事件并行处理
- **缓存机制** - 仓库克隆缓存，减少网络开销
- **超时控制** - 可配置的请求超时和重试策略
- **资源限制** - 文件大小限制和内存使用优化
- **频率限制** - 防止同一仓库频繁克隆
- **工作空间隔离** - 每次操作使用独立目录
- **自动清理** - 操作完成后自动清理临时文件

## 🤝 贡献指南

### 开发流程
1. Fork项目仓库
2. 创建功能分支: `git checkout -b feature/amazing-feature`
3. 提交更改: `git commit -m 'feat: add amazing feature'`
4. 推送分支: `git push origin feature/amazing-feature`
5. 创建Pull Request

### 代码规范
- 遵循Go语言编码规范
- 使用Conventional Commits提交信息
- 添加必要的测试和文档
- 确保所有测试通过

### 提交信息格式
```
type(scope): description

[optional body]

[optional footer]
```

**类型说明：**
- `feat`: 新功能
- `fix`: 修复问题
- `docs`: 文档更新
- `style`: 代码风格调整
- `refactor`: 代码重构
- `test`: 测试相关
- `chore`: 构建配置、依赖管理等

## 📄 开源许可

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 🌟 致谢

- [Claude Code CLI](https://docs.anthropic.com/en/docs/claude-code) - 强大的AI代码助手
- [Gin Web Framework](https://gin-gonic.com/) - 高性能Go Web框架
- [GitHub API](https://docs.github.com/en/rest) - 完善的仓库管理API
- [Conventional Commits](https://www.conventionalcommits.org/) - 标准化提交信息规范

## 📈 项目状态

- ✅ **稳定版本**: v1.0.0
- ✅ **Claude Code CLI集成**: 完整支持
- ✅ **GitHub Webhook**: 全功能支持
- ✅ **Docker部署**: 生产就绪
- ✅ **文档完善**: 详细的使用和开发指南
- 🔄 **持续改进**: 定期更新和优化

---

⭐ 如果这个项目对你有帮助，请给个Star支持一下！

🐛 遇到问题？[提交Issue](https://github.com/your-repo/issues)  
💬 交流讨论？[参与Discussions](https://github.com/your-repo/discussions)

📧 联系作者？[发送邮件](mailto:your-email@example.com)

---

**最后更新**: 2024年8月
**维护状态**: 活跃维护中
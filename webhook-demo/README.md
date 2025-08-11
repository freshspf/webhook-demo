# GitHub Webhook Demo

这是一个使用Go语言实现的GitHub Webhook处理演示项目，展示了如何接收和处理GitHub事件，实现类似CodeAgent的自动化工作流。

## 🎯 功能特性

- ✅ **GitHub Webhook接收**: 监听GitHub仓库事件
- 🔐 **签名验证**: HMAC-SHA256签名验证确保安全性
- 🎭 **事件分发**: 智能分发不同类型的GitHub事件
- 🤖 **命令解析**: 支持 `/code`、`/continue`、`/fix`、`/help` 等命令
- 📝 **自动回复**: 在Issue和PR中自动创建响应评论
- 🛡️ **优雅关闭**: 支持信号处理和优雅关闭
- 📊 **健康检查**: 提供健康检查端点

## 🏗️ 项目结构

```
webhook-demo/
├── main.go                           # 主程序入口
├── go.mod                           # Go模块文件
├── config.env.example               # 环境变量配置示例
├── README.md                        # 说明文档
└── internal/
    ├── config/
    │   └── config.go                # 配置管理
    ├── handlers/
    │   └── webhook.go               # Webhook处理器
    ├── middleware/
    │   └── cors.go                  # CORS中间件
    ├── models/
    │   └── github.go                # GitHub事件模型
    └── services/
        ├── github.go                # GitHub API服务
        └── event_processor.go       # 事件处理器
```

## 🚀 快速开始

### 1. 环境准备

确保已安装Go 1.21或更高版本：

```bash
go version
```

### 2. 克隆并初始化项目

```bash
# 进入项目目录
cd webhook-demo

# 初始化Go模块
go mod tidy
```

### 3. 配置环境变量

```bash
# 复制配置文件
cp config.env.example .env

# 编辑配置文件，填入你的GitHub Token和Webhook Secret
# GITHUB_TOKEN: 在GitHub Settings > Developer settings > Personal access tokens创建
# GITHUB_WEBHOOK_SECRET: 在仓库Webhook设置中配置的密钥
```

### 4. 启动服务

```bash
# 设置环境变量并启动
source .env
go run main.go
```

服务启动后会监听在 `http://localhost:8080`

### 5. 配置GitHub Webhook

在你的GitHub仓库中设置Webhook：

1. 进入仓库 Settings > Webhooks
2. 点击 "Add webhook"
3. 配置以下信息：
   - **Payload URL**: `http://your-server:8080/webhook`
   - **Content type**: `application/json`
   - **Secret**: 与环境变量 `GITHUB_WEBHOOK_SECRET` 相同
   - **Events**: 选择需要的事件（建议选择 Issues, Issue comments, Pull requests）

## 📋 支持的命令

在Issue或PR评论中使用以下命令：

- `/code <需求描述>` - 生成代码实现指定功能
- `/continue [说明]` - 继续当前的开发任务  
- `/fix <问题描述>` - 修复指定的代码问题
- `/help` - 显示帮助信息

### 使用示例

```
/code 实现用户登录功能
/continue 添加错误处理
/fix 修复空指针异常
/help
```

## 🔄 工作流程

1. **接收事件**: GitHub发送Webhook事件到服务器
2. **验证签名**: 使用HMAC-SHA256验证请求来源
3. **解析事件**: 根据事件类型解析JSON payload
4. **处理命令**: 从Issue/PR评论中提取命令
5. **执行操作**: 根据命令类型执行相应的处理逻辑
6. **响应结果**: 在GitHub界面创建回复评论

## 🛠️ API端点

- `GET /` - API信息
- `GET /health` - 健康检查
- `POST /webhook` - GitHub Webhook端点

### 健康检查示例

```bash
curl http://localhost:8080/health
```

响应：
```json
{
  "status": "ok",
  "timestamp": 1234567890
}
```

## 🔧 扩展开发

### 添加新的事件处理器

在 `internal/services/event_processor.go` 中添加新的事件处理方法：

```go
func (ep *EventProcessor) handleCustomEvent(event *models.GitHubEvent) error {
    // 自定义事件处理逻辑
    return nil
}
```

### 添加新的命令

在 `executeCommand` 方法中添加新的命令处理：

```go
case "newcommand":
    return ep.handleNewCommand(command, ctx)
```

### 自定义GitHub API调用

扩展 `internal/services/github.go` 添加更多GitHub API调用方法。

## 🐳 Docker部署

创建 `Dockerfile`:

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o webhook-demo main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/webhook-demo .

EXPOSE 8080
CMD ["./webhook-demo"]
```

构建和运行：

```bash
docker build -t webhook-demo .
docker run -p 8080:8080 --env-file .env webhook-demo
```

## 🔒 安全考虑

1. **签名验证**: 始终验证GitHub Webhook签名
2. **环境变量**: 敏感信息存储在环境变量中
3. **HTTPS**: 生产环境建议使用HTTPS
4. **访问控制**: 限制访问来源IP（如需要）

## 📝 日志记录

服务会输出详细的日志信息，包括：

- 接收到的事件类型和内容
- 签名验证结果
- 命令解析和执行过程
- GitHub API调用结果
- 错误信息和异常处理

## 🔍 故障排除

### 1. Webhook接收不到事件

- 检查GitHub Webhook配置中的URL是否正确
- 确认服务器能从外网访问
- 查看GitHub Webhook的Delivery日志

### 2. 签名验证失败

- 检查 `GITHUB_WEBHOOK_SECRET` 环境变量
- 确认GitHub Webhook设置中的Secret与环境变量一致

### 3. GitHub API调用失败

- 检查 `GITHUB_TOKEN` 是否有效
- 确认Token有足够的权限（repo, issues, pull_requests）

## 📚 相关资源

- [GitHub Webhooks文档](https://docs.github.com/en/developers/webhooks-and-events/webhooks)
- [GitHub API文档](https://docs.github.com/en/rest)
- [Gin Web框架](https://gin-gonic.com/)

## 🤝 贡献

欢迎提交Issue和Pull Request来改进这个项目！

## 📄 许可证

MIT License
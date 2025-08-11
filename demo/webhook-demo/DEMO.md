# GitHub Webhook 机制演示

## 🎯 Webhook工作原理

GitHub Webhook是一种HTTP POST请求机制，当仓库中发生特定事件时，GitHub会自动向你配置的URL发送HTTP请求，携带事件的详细信息。

### 核心概念

1. **事件触发**: 用户在GitHub仓库中执行操作（创建Issue、评论、提交代码等）
2. **HTTP请求**: GitHub自动向配置的服务器发送POST请求
3. **安全验证**: 使用HMAC-SHA256签名确保请求来源的安全性
4. **事件处理**: 服务器接收并处理事件，执行相应的业务逻辑
5. **响应反馈**: 处理完成后通过GitHub API返回结果

## 🔄 完整流程演示

### 1. 启动服务

```bash
# 方法1: 使用启动脚本（推荐）
./scripts/start.sh

# 方法2: 直接运行
export GITHUB_TOKEN="your_token_here"
export GITHUB_WEBHOOK_SECRET="your_secret_here"
go run main.go
```

### 2. 配置GitHub Webhook

在你的GitHub仓库中配置Webhook：

1. 进入仓库 `Settings` → `Webhooks`
2. 点击 `Add webhook`
3. 配置以下信息：
   - **Payload URL**: `http://your-server:8080/webhook`
   - **Content type**: `application/json`
   - **Secret**: 与环境变量 `GITHUB_WEBHOOK_SECRET` 相同
   - **Events**: 选择 `Issues` 和 `Issue comments`

### 3. 测试Webhook

创建一个Issue或在Issue中添加评论，包含以下命令：

```
/help
```

你应该会看到CodeAgent的自动回复。

## 🧪 本地测试

如果没有公网服务器，可以使用工具将本地服务暴露到公网：

### 使用ngrok

```bash
# 安装ngrok
npm install -g ngrok

# 启动服务
go run main.go

# 在另一个终端暴露本地服务
ngrok http 8080
```

然后使用ngrok提供的公网URL配置GitHub Webhook。

### 使用测试脚本

```bash
# 启动服务
./scripts/start.sh

# 在另一个终端运行测试
./scripts/test.sh
```

## 📋 支持的事件类型

### 1. Issues事件
- `opened`: Issue被创建
- `edited`: Issue被编辑
- `closed`: Issue被关闭

### 2. Issue Comments事件
- `created`: 评论被创建

### 3. Pull Request事件
- `opened`: PR被创建
- `synchronize`: PR代码被更新
- `closed`: PR被关闭

### 4. Pull Request Review Comments事件
- `created`: Review评论被创建

## 🤖 支持的命令

在Issue或PR评论中使用以下命令：

### `/code <需求描述>`
生成代码实现指定功能

**示例:**
```
/code 实现用户登录功能，包括邮箱验证和密码加密
```

**响应:**
- 分析需求
- 模拟AI代码生成过程
- 显示处理状态

### `/continue [说明]`
继续当前的开发任务

**示例:**
```
/continue 添加错误处理和日志记录
```

### `/fix <问题描述>`
修复指定的代码问题

**示例:**
```
/fix 修复用户登录时的空指针异常
```

### `/help`
显示帮助信息

**示例:**
```
/help
```

## 🔧 技术实现细节

### 1. 签名验证

```go
func verifySignature(signature string, payload []byte, secret string) bool {
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(payload)
    expectedSignature := hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(signature), []byte("sha256="+expectedSignature))
}
```

### 2. 事件解析

```go
type GitHubEvent struct {
    Type       string    `json:"type"`
    DeliveryID string    `json:"delivery_id"`
    Payload    []byte    `json:"payload"`
    Timestamp  time.Time `json:"timestamp"`
}
```

### 3. 命令提取

```go
commandRegex := regexp.MustCompile(`^/(code|continue|fix|help)\s*(.*)$`)
```

### 4. GitHub API调用

```go
func CreateComment(owner, repo string, issueNumber int, body string) error {
    url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d/comments", owner, repo, issueNumber)
    // 发起HTTP POST请求
}
```

## 🛡️ 安全考虑

### 1. 签名验证
- 使用HMAC-SHA256算法验证请求来源
- 防止伪造的webhook请求

### 2. 环境变量
- 敏感信息存储在环境变量中
- 不在代码中硬编码Token和Secret

### 3. 错误处理
- 详细的错误日志记录
- 优雅的错误响应

## 🚀 扩展指南

### 1. 添加新的事件类型

在 `event_processor.go` 中添加：

```go
case "new_event_type":
    return ep.handleNewEventType(event)
```

### 2. 添加新的命令

在 `executeCommand` 方法中添加：

```go
case "newcommand":
    return ep.handleNewCommand(command, ctx)
```

### 3. 集成AI服务

```go
type AIService interface {
    GenerateCode(prompt string) (string, error)
    ContinueTask(context string) (string, error)
    FixCode(problem string) (string, error)
}
```

### 4. 添加数据库存储

```go
type EventStore interface {
    SaveEvent(event *GitHubEvent) error
    GetEventHistory(repoID int64) ([]*GitHubEvent, error)
}
```

## 📊 监控和调试

### 查看日志
```bash
# 服务会输出详细的处理日志
go run main.go
```

### 健康检查
```bash
curl http://localhost:8080/health
```

### 测试特定事件
```bash
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -H "X-GitHub-Event: ping" \
  -d '{"zen":"Hello World"}'
```

## 💡 最佳实践

1. **使用HTTPS**: 生产环境中始终使用HTTPS
2. **设置超时**: 为HTTP请求设置合理的超时时间
3. **限制频率**: 实现速率限制防止滥用
4. **日志记录**: 记录所有重要的操作和错误
5. **错误重试**: 对失败的GitHub API调用实现重试机制
6. **监控告警**: 设置服务监控和告警机制

## 🔗 相关链接

- [GitHub Webhooks文档](https://docs.github.com/en/developers/webhooks-and-events/webhooks)
- [GitHub API文档](https://docs.github.com/en/rest)
- [HMAC签名验证](https://docs.github.com/en/developers/webhooks-and-events/webhooks/securing-your-webhooks)

这个演示展示了完整的GitHub Webhook工作机制，从接收事件到处理命令再到响应结果的全流程。你可以基于这个框架扩展更复杂的功能，比如集成真实的AI服务、实现工作空间管理、添加数据库存储等。
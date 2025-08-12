# 🚀 Claude Webhook 快速启动指南

## 📋 前置要求

- Go 1.21 或更高版本
- Claude API 密钥
- GitHub 个人访问令牌
- 可公网访问的服务器（用于接收GitHub Webhook）

## ⚡ 5分钟快速启动

### 1. 获取API密钥

**Claude API密钥:**
1. 访问 [Anthropic Console](https://console.anthropic.com/)
2. 注册/登录账户
3. 创建API密钥
4. 复制密钥（格式：`sk-ant-api03-...`）

**GitHub Token:**
1. 访问 [GitHub Settings > Developer settings > Personal access tokens](https://github.com/settings/tokens)
2. 创建新token
3. 选择权限：`repo`, `issues`, `pull_requests`
4. 复制token

### 2. 配置环境变量

```bash
# 复制配置文件
cp config.env.example .env

# 编辑配置文件
nano .env
```

填入以下信息：
```bash
# GitHub配置
GITHUB_TOKEN=ghp_your_github_token_here
GITHUB_WEBHOOK_SECRET=your_webhook_secret_here

# Claude API配置
CLAUDE_API_KEY=sk-ant-api03-your_claude_api_key_here
CLAUDE_MODEL=claude-3-5-sonnet-20241022
CLAUDE_MAX_TOKENS=4000

# 服务器配置
SERVER_PORT=8080
GIN_MODE=debug
```

### 3. 启动服务

```bash
# 编译并启动
go run main.go
```

服务将在 `http://localhost:8080` 启动

### 4. 配置GitHub Webhook

1. 进入您的GitHub仓库
2. 点击 `Settings` > `Webhooks`
3. 点击 `Add webhook`
4. 配置：
   - **Payload URL**: `http://your-server:8080/webhook`
   - **Content type**: `application/json`
   - **Secret**: 与 `.env` 中的 `GITHUB_WEBHOOK_SECRET` 相同
   - **Events**: 选择 `Issues`, `Issue comments`, `Pull requests`

### 5. 测试功能

在GitHub Issue或PR评论中输入：

```
/code 实现一个简单的用户登录功能
```

## 🎯 支持的命令

| 命令 | 功能 | 示例 |
|------|------|------|
| `/code <需求>` | 生成代码 | `/code 实现用户注册功能` |
| `/continue [说明]` | 继续开发 | `/continue 添加密码验证` |
| `/fix <问题>` | 修复代码 | `/fix 修复空指针异常` |
| `/help` | 显示帮助 | `/help` |

## 🔧 故障排除

### 常见问题

**1. 编译错误**
```bash
go mod tidy
go build -o webhook-demo main.go
```

**2. API密钥错误**
```bash
# 检查环境变量
echo $CLAUDE_API_KEY
echo $GITHUB_TOKEN
```

**3. 网络连接问题**
```bash
# 测试Claude API连接
curl -X POST https://api.anthropic.com/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: $CLAUDE_API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -d '{"model":"claude-3-5-sonnet-20241022","max_tokens":100,"messages":[{"role":"user","content":"Hello"}]}'
```

**4. Webhook接收不到事件**
- 检查服务器是否可从外网访问
- 确认GitHub Webhook URL正确
- 查看GitHub Webhook的Delivery日志

### 调试模式

```bash
# 启用详细日志
export GIN_MODE=debug
go run main.go
```

## 📊 监控和日志

服务会输出详细日志：
```
2024/01/15 10:30:00 Webhook服务器已启动，端口: 8080
2024/01/15 10:30:05 开始处理事件: Type=issue_comment, DeliveryID=abc123
2024/01/15 10:30:06 在评论中检测到命令: code
2024/01/15 10:30:10 调用Claude API，模型: claude-3-5-sonnet-20241022
2024/01/15 10:30:15 Claude API调用成功，输入Token: 150, 输出Token: 800
```

## 🔒 安全建议

1. **使用HTTPS**: 生产环境建议使用HTTPS
2. **防火墙**: 只开放必要端口
3. **密钥管理**: 定期轮换API密钥
4. **访问控制**: 限制Webhook来源IP（可选）

## 📚 更多资源

- [详细集成文档](CLAUDE_INTEGRATION.md)
- [项目README](README.md)
- [Claude API文档](https://docs.anthropic.com/claude/reference)
- [GitHub Webhooks文档](https://docs.github.com/en/developers/webhooks-and-events/webhooks)

## 🆘 获取帮助

- 查看 [故障排除](#故障排除) 部分
- 检查 [详细集成文档](CLAUDE_INTEGRATION.md)
- 提交 [GitHub Issue](https://github.com/your-repo/issues)

---

**🎉 恭喜！您的Claude Webhook服务已经成功启动并运行！**

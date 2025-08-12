# Webhook 服务管理指南

## 🚀 快速开始

### 1. 启动服务
```bash
./start.sh
```

### 2. 检查服务状态
```bash
./status.sh
```

### 3. 停止服务
```bash
./stop.sh
```

## 📋 服务管理脚本

### `start.sh` - 启动脚本
- ✅ 检查服务是否已运行
- ✅ 检查端口是否被占用
- ✅ 检查 `.env` 配置文件
- ✅ 启动服务并显示状态信息

### `stop.sh` - 停止脚本
- ✅ 查找并停止所有 webhook 服务进程
- ✅ 检查端口是否释放
- ✅ 显示停止状态

### `status.sh` - 状态检查脚本
- ✅ 检查服务进程状态
- ✅ 检查端口监听状态
- ✅ 测试服务响应
- ✅ 显示最近日志
- ✅ 显示服务信息和管理命令

### `diagnose_webhook.sh` - 诊断脚本
- ✅ 全面检查服务状态
- ✅ 检查网络连接
- ✅ 检查 GitHub webhook 配置
- ✅ 测试 webhook 端点

## 🔧 常用命令

### 查看实时日志
```bash
tail -f webhook.log
```

### 查看最近日志
```bash
tail -20 webhook.log
```

### 测试健康检查
```bash
curl http://localhost:8080/health
```

### 测试 webhook 端点
```bash
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -H "X-GitHub-Event: ping" \
  -H "X-GitHub-Delivery: test-id" \
  -H "X-Hub-Signature-256: sha256=your-signature" \
  -d '{"zen":"test"}'
```

## 🌐 服务端点

- **健康检查**: `http://localhost:8080/health`
- **Webhook 端点**: `http://localhost:8080/webhook`
- **API 信息**: `http://localhost:8080/`

## ⚠️ 常见问题

### 端口被占用
```bash
# 查看占用进程
netstat -tlnp | grep :8080

# 停止占用进程
kill <PID>
```

### 服务启动失败
```bash
# 查看错误日志
tail -20 webhook.log

# 检查配置
cat .env
```

### 无法接收 GitHub 事件
1. 检查 GitHub webhook 配置
2. 确认 webhook URL 正确
3. 确认 webhook secret 匹配
4. 检查防火墙设置

## 📝 日志说明

服务日志包含以下信息：
- 服务启动/停止信息
- 接收到的 GitHub 事件
- 签名验证结果
- 事件处理状态
- HTTP 请求日志

## 🔒 安全配置

确保 `.env` 文件包含正确的配置：
- `GITHUB_TOKEN`: GitHub 个人访问令牌
- `GITHUB_WEBHOOK_SECRET`: Webhook 密钥
- `CLAUDE_API_KEY`: Claude API 密钥

## 📞 支持

如果遇到问题，请：
1. 运行 `./diagnose_webhook.sh` 进行诊断
2. 检查 `webhook.log` 日志文件
3. 确认 GitHub webhook 配置正确

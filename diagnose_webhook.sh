#!/bin/bash

echo "=== GitHub Webhook 诊断工具 ==="
echo

# 读取端口配置（只取第一个匹配的行）
PORT=$(grep "^SERVER_PORT=" .env 2>/dev/null | head -1 | cut -d'=' -f2 | tr -d '\r\n' || echo "8080")

# 检查服务是否运行
echo "1. 检查服务状态..."
if pgrep -f "webhook-demo" > /dev/null; then
    echo "✅ 服务正在运行"
    ps aux | grep webhook-demo | grep -v grep
else
    echo "❌ 服务未运行"
fi
echo

# 检查端口监听
echo "2. 检查端口监听..."
if netstat -tlnp 2>/dev/null | grep ":$PORT " > /dev/null; then
    echo "✅ 端口 $PORT 正在监听"
    netstat -tlnp 2>/dev/null | grep ":$PORT "
else
    echo "❌ 端口 $PORT 未监听"
fi
echo

# 检查环境变量
echo "3. 检查环境变量..."
echo "SERVER_PORT: $PORT"
echo "GITHUB_TOKEN: ${GITHUB_TOKEN:+已设置}"
echo "GITHUB_WEBHOOK_SECRET: ${GITHUB_WEBHOOK_SECRET:+已设置}"
echo "CLAUDE_API_KEY: ${CLAUDE_API_KEY:+已设置}"
echo

# 检查 .env 文件
echo "4. 检查 .env 文件..."
if [ -f ".env" ]; then
    echo "✅ .env 文件存在"
    echo "文件内容预览:"
    head -10 .env | sed 's/=.*/=***/'  # 隐藏敏感信息
else
    echo "❌ .env 文件不存在"
    echo "请复制 config.env.example 为 .env 并配置相应参数"
fi
echo

# 检查网络连接
echo "5. 检查网络连接..."
if curl -s --connect-timeout 5 http://localhost:$PORT/health > /dev/null; then
    echo "✅ 本地服务响应正常"
    curl -s http://localhost:$PORT/health | jq . 2>/dev/null || curl -s http://localhost:$PORT/health
else
    echo "❌ 本地服务无响应"
fi
echo

# 检查公网访问
echo "6. 检查公网访问..."
PUBLIC_IP=$(curl -s --connect-timeout 5 ifconfig.me 2>/dev/null || echo "无法获取公网IP")
echo "公网IP: $PUBLIC_IP"
echo "Webhook URL: http://$PUBLIC_IP:$PORT/webhook"
echo

# 检查 GitHub webhook 配置建议
echo "7. GitHub Webhook 配置建议:"
echo "在您的 GitHub 仓库中配置 webhook:"
echo "  - URL: http://$PUBLIC_IP:$PORT/webhook"
echo "  - Content type: application/json"
echo "  - Secret: $GITHUB_WEBHOOK_SECRET"
echo "  - 选择事件:"
echo "    ✅ Issues"
echo "    ✅ Issue comments"
echo "    ✅ Pull requests"
echo "    ✅ Pull request review comments"
echo

# 测试 webhook 端点
echo "8. 测试 webhook 端点..."
echo "发送测试 ping 事件..."
curl -X POST http://localhost:$PORT/webhook \
  -H "Content-Type: application/json" \
  -H "X-GitHub-Event: ping" \
  -H "X-GitHub-Delivery: test-delivery-id" \
  -d '{"zen":"Test webhook connection"}' 2>/dev/null | jq . 2>/dev/null || echo "测试失败或无响应"
echo

echo "=== 诊断完成 ==="
echo
echo "如果服务未运行，请执行:"
echo "  go run main.go"
echo
echo "如果端口未监听，请检查防火墙设置"
echo "如果公网无法访问，请使用 ngrok 等工具进行端口转发"
echo
echo "配置步骤:"
echo "1. 编辑 .env 文件，设置正确的配置参数"
echo "2. 在 GitHub 仓库中配置 webhook"
echo "3. 启动服务: go run main.go"
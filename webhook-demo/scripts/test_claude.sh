#!/bin/bash

# Claude API集成测试脚本

echo "🧪 测试Claude API集成..."

# 检查环境变量
if [ -z "$CLAUDE_API_KEY" ]; then
    echo "❌ 错误: CLAUDE_API_KEY 环境变量未设置"
    echo "请设置您的Claude API密钥:"
    echo "export CLAUDE_API_KEY=your-api-key-here"
    exit 1
fi

# 检查配置文件
if [ ! -f ".env" ]; then
    echo "❌ 错误: .env 配置文件不存在"
    echo "请复制配置文件: cp config.env.example .env"
    exit 1
fi

# 编译项目
echo "🔨 编译项目..."
go build -o webhook-demo main.go
if [ $? -ne 0 ]; then
    echo "❌ 编译失败"
    exit 1
fi

echo "✅ 编译成功"

# 测试API连接
echo "🌐 测试Claude API连接..."
curl -X POST https://api.anthropic.com/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: $CLAUDE_API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -d '{
    "model": "claude-3-5-sonnet-20241022",
    "max_tokens": 100,
    "messages": [{"role": "user", "content": "Hello, please respond with just 'OK'"}]
  }' 2>/dev/null | grep -q "OK"

if [ $? -eq 0 ]; then
    echo "✅ Claude API连接成功"
else
    echo "❌ Claude API连接失败"
    echo "请检查:"
    echo "1. API密钥是否正确"
    echo "2. 网络连接是否正常"
    echo "3. API配额是否充足"
    exit 1
fi

# 启动服务进行测试
echo "🚀 启动webhook服务进行测试..."
./webhook-demo &
SERVER_PID=$!

# 等待服务启动
sleep 3

# 测试健康检查
echo "🏥 测试健康检查..."
curl -s http://localhost:8080/health | grep -q "ok"
if [ $? -eq 0 ]; then
    echo "✅ 服务健康检查通过"
else
    echo "❌ 服务健康检查失败"
    kill $SERVER_PID 2>/dev/null
    exit 1
fi

# 测试API信息
echo "📋 测试API信息..."
curl -s http://localhost:8080/ | grep -q "GitHub Webhook Demo"
if [ $? -eq 0 ]; then
    echo "✅ API信息获取成功"
else
    echo "❌ API信息获取失败"
fi

# 停止服务
echo "🛑 停止测试服务..."
kill $SERVER_PID 2>/dev/null

echo ""
echo "🎉 Claude API集成测试完成!"
echo ""
echo "📝 下一步:"
echo "1. 配置GitHub Webhook: http://your-server:8080/webhook"
echo "2. 在GitHub Issue/PR中使用命令:"
echo "   - /code 实现用户登录功能"
echo "   - /continue 添加错误处理"
echo "   - /fix 修复空指针异常"
echo "   - /help 显示帮助信息"
echo ""
echo "📚 更多信息请查看: CLAUDE_INTEGRATION.md"

#!/bin/bash

# 自动修复功能测试脚本

set -e

echo "🧪 开始测试自动修复功能..."

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "❌ 错误: 未找到Go环境"
    exit 1
fi

# 检查Git环境
if ! command -v git &> /dev/null; then
    echo "❌ 错误: 未找到Git环境"
    exit 1
fi

# 检查配置文件
if [ ! -f ".env" ]; then
    echo "❌ 错误: 未找到 .env 配置文件"
    echo "请先复制 config.env.example 为 .env 并配置相关参数"
    exit 1
fi

# 加载环境变量
source .env

# 检查必需的环境变量
if [ -z "$GITHUB_TOKEN" ] || [ "$GITHUB_TOKEN" = "your_github_personal_access_token_here" ]; then
    echo "❌ 错误: 请在 .env 文件中配置有效的 GITHUB_TOKEN"
    exit 1
fi

if [ -z "$CLAUDE_API_KEY" ] || [ "$CLAUDE_API_KEY" = "your_claude_api_key_here" ]; then
    echo "❌ 错误: 请在 .env 文件中配置有效的 CLAUDE_API_KEY"
    exit 1
fi

echo "✅ 环境检查通过"

# 编译项目
echo "🔨 编译项目..."
go build -o webhook-demo main.go
echo "✅ 编译成功"

# 检查工作目录
if [ ! -d "$GIT_WORK_DIR" ]; then
    echo "📁 创建工作目录: $GIT_WORK_DIR"
    mkdir -p "$GIT_WORK_DIR"
fi

# 测试Git配置
echo "🔧 测试Git配置..."
git config --global user.name "CodeAgent Test"
git config --global user.email "test@codeagent.com"
echo "✅ Git配置成功"

# 启动服务（后台运行）
echo "🚀 启动Webhook服务..."
./webhook-demo &
SERVER_PID=$!

# 等待服务启动
echo "⏳ 等待服务启动..."
sleep 3

# 检查服务是否启动成功
if ! curl -s http://localhost:${SERVER_PORT:-8080}/health > /dev/null; then
    echo "❌ 服务启动失败"
    kill $SERVER_PID 2>/dev/null || true
    exit 1
fi

echo "✅ 服务启动成功，PID: $SERVER_PID"

# 显示服务信息
echo ""
echo "📋 服务信息:"
echo "   - 服务地址: http://localhost:${SERVER_PORT:-8080}"
echo "   - Webhook端点: http://localhost:${SERVER_PORT:-8080}/webhook"
echo "   - 健康检查: http://localhost:${SERVER_PORT:-8080}/health"
echo ""

echo "🎯 测试说明:"
echo "1. 在您的GitHub仓库中配置Webhook:"
echo "   - URL: http://your-server:${SERVER_PORT:-8080}/webhook"
echo "   - Secret: $GITHUB_WEBHOOK_SECRET"
echo "   - Events: Issues, Issue comments, Pull requests"
echo ""
echo "2. 创建一个Issue，系统会自动:"
echo "   - 克隆仓库"
echo "   - AI分析需求"
echo "   - 修改代码"
echo "   - 创建Pull Request"
echo "   - 回复Issue"
echo ""

echo "🔄 服务正在运行中..."
echo "按 Ctrl+C 停止服务"

# 等待用户中断
trap "echo ''; echo '🛑 正在停止服务...'; kill $SERVER_PID 2>/dev/null || true; echo '✅ 服务已停止'; exit 0" INT

# 保持脚本运行
while true; do
    sleep 1
done

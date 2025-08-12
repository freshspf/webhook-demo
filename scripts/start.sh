#!/bin/bash

# GitHub Webhook Demo 启动脚本

set -e

echo "🚀 启动 GitHub Webhook Demo..."

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "❌ 错误: 未找到Go环境，请先安装Go 1.21或更高版本"
    exit 1
fi

echo "✅ Go环境检查通过: $(go version)"

# 检查配置文件
if [ ! -f ".env" ]; then
    if [ -f "config.env.example" ]; then
        echo "📋 正在复制配置文件模板..."
        cp config.env.example .env
        echo "⚠️  请编辑 .env 文件，配置你的 GitHub Token 和 Webhook Secret"
        echo "   GITHUB_TOKEN: 在 GitHub Settings > Developer settings > Personal access tokens 创建"
        echo "   GITHUB_WEBHOOK_SECRET: 在仓库 Webhook 设置中配置的密钥"
        echo ""
        echo "配置完成后，请重新运行此脚本"
        exit 1
    else
        echo "❌ 错误: 未找到配置文件，请创建 .env 文件"
        exit 1
    fi
fi

# 加载环境变量
echo "📝 加载环境变量..."
set -a
source .env
set +a

# 检查必需的环境变量
if [ -z "$GITHUB_TOKEN" ] || [ "$GITHUB_TOKEN" = "your_github_personal_access_token_here" ]; then
    echo "❌ 错误: 请在 .env 文件中配置有效的 GITHUB_TOKEN"
    exit 1
fi

if [ -z "$GITHUB_WEBHOOK_SECRET" ] || [ "$GITHUB_WEBHOOK_SECRET" = "your_webhook_secret_here" ]; then
    echo "⚠️  警告: GITHUB_WEBHOOK_SECRET 未配置，将跳过签名验证"
fi

# 初始化Go模块
echo "📦 初始化Go模块..."
go mod tidy

# 检查端口是否被占用
PORT=${SERVER_PORT:-8080}
if lsof -Pi :$PORT -sTCP:LISTEN -t >/dev/null ; then
    echo "❌ 错误: 端口 $PORT 已被占用，请关闭占用该端口的程序或更改 SERVER_PORT"
    exit 1
fi

echo "🔧 配置信息:"
echo "   服务器端口: $PORT"
echo "   运行模式: ${GIN_MODE:-debug}"
echo "   GitHub Token: ${GITHUB_TOKEN:0:4}***"
echo "   Webhook Secret: ${GITHUB_WEBHOOK_SECRET:+已配置}"
echo ""

echo "🌐 Webhook URL: http://localhost:$PORT/webhook"
echo "💡 在GitHub仓库设置中配置Webhook，将Payload URL设置为上述地址"
echo ""

# 启动服务
echo "🚀 启动服务..."
echo "按 Ctrl+C 停止服务"
echo ""

go run main.go
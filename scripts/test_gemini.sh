#!/bin/bash

# Gemini服务测试脚本
# 用于验证Gemini CLI集成是否正常工作

set -e

echo "🧪 开始测试Gemini集成..."

# 检查是否在项目目录中
if [ ! -f "go.mod" ]; then
    echo "❌ 请在项目根目录中运行此脚本"
    exit 1
fi

# 检查环境变量
if [ ! -f ".env" ]; then
    echo "❌ 未找到.env文件，请先复制config.env.example到.env并配置"
    exit 1
fi

# 加载环境变量
source .env

echo "🔍 检查Gemini CLI安装..."
if ! command -v gemini &> /dev/null; then
    echo "❌ Gemini CLI未安装，请运行: ./scripts/install_gemini_cli.sh"
    exit 1
fi

echo "✅ Gemini CLI已安装: $(gemini --version)"

# 检查API密钥
if [ -z "$GEMINI_API_KEY" ]; then
    echo "❌ GEMINI_API_KEY 未配置，请在.env文件中设置"
    exit 1
fi

echo "✅ API密钥已配置"

# 测试Gemini CLI连接
echo "🚀 测试Gemini CLI连接..."
GEMINI_TEST_PROMPT="Hello, please respond with 'Gemini CLI integration test successful' if you can see this message."

# 设置环境变量并测试
export GEMINI_API_KEY="$GEMINI_API_KEY"
if [ ! -z "$GEMINI_MODEL" ]; then
    export GEMINI_MODEL="$GEMINI_MODEL"
fi
if [ ! -z "$GOOGLE_CLOUD_PROJECT" ]; then
    export GOOGLE_CLOUD_PROJECT="$GOOGLE_CLOUD_PROJECT"
fi

# 使用超时命令执行测试
if timeout 30 gemini "$GEMINI_TEST_PROMPT" > /tmp/gemini_test_output.txt 2>&1; then
    echo "✅ Gemini CLI连接测试成功"
    echo "📄 测试响应:"
    echo "---"
    cat /tmp/gemini_test_output.txt
    echo "---"
    rm -f /tmp/gemini_test_output.txt
else
    echo "❌ Gemini CLI连接测试失败"
    echo "📄 错误输出:"
    echo "---"
    cat /tmp/gemini_test_output.txt
    echo "---"
    rm -f /tmp/gemini_test_output.txt
    echo ""
    echo "💡 可能的解决方案:"
    echo "1. 检查GEMINI_API_KEY是否正确"
    echo "2. 确保网络连接正常"
    echo "3. 尝试运行 'gemini' 命令进行首次认证"
    echo "4. 检查Google Cloud项目配置（如果使用）"
    exit 1
fi

# 编译项目
echo "🔨 编译项目..."
if go build -v .; then
    echo "✅ 项目编译成功"
else
    echo "❌ 项目编译失败"
    exit 1
fi

# 测试服务启动（不实际启动，只验证配置）
echo "⚙️  验证服务配置..."
if timeout 5 ./webhook-demo --help > /dev/null 2>&1 || [ $? -eq 1 ]; then
    echo "✅ 服务配置验证通过"
else
    echo "❌ 服务配置验证失败"
    exit 1
fi

echo ""
echo "🎉 所有测试通过！Gemini集成配置正确"
echo ""
echo "📋 下一步:"
echo "1. 配置GitHub Webhook URL"
echo "2. 运行 './start.sh' 启动服务"
echo "3. 在GitHub Issue中使用 /code、/continue、/fix 命令测试功能"
echo ""
echo "🔗 相关文档:"
echo "- 迁移指南: ./GEMINI_MIGRATION.md"
echo "- 快速开始: ./QUICK_START.md"
echo "- 项目README: ./README.md"

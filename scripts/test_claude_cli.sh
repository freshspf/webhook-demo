#!/bin/bash

set -e

echo "🔧 Claude Code CLI 诊断脚本"
echo "=============================="

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "❌ Go未安装，请先安装Go"
    exit 1
fi

# 检查Claude CLI
if ! command -v claude &> /dev/null; then
    echo "❌ Claude CLI未安装"
    echo "   请运行: npm install -g @anthropic-ai/claude-code"
    exit 1
fi

echo "✅ Claude CLI已安装"

# 检查版本
echo "📋 Claude CLI版本:"
claude --version

# 检查环境变量
echo ""
echo "📋 环境变量检查:"
echo "   ANTHROPIC_API_KEY: ${ANTHROPIC_API_KEY:+已设置}${ANTHROPIC_API_KEY:-未设置}"
echo "   ANTHROPIC_BASE_URL: ${ANTHROPIC_BASE_URL:-未设置}"
echo "   CLAUDE_CODE_CLI_MODEL: ${CLAUDE_CODE_CLI_MODEL:-未设置}"

# 运行Go诊断程序
echo ""
echo "🧪 运行Go诊断程序..."
cd "$(dirname "$0")/.."
go run scripts/test_claude_cli.go

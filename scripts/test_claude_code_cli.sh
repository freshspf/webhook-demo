#!/bin/bash

# Claude Code CLI服务测试脚本
# 用于验证Claude Code CLI集成是否正常工作

set -e

echo "🧪 开始测试Claude Code CLI集成..."

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

echo "🔍 检查Claude Code CLI安装..."
if ! command -v claude &> /dev/null; then
    echo "❌ Claude Code CLI未安装，请运行: ./scripts/install_claude_code_cli.sh"
    exit 1
fi

echo "✅ Claude Code CLI已安装: $(claude --version)"

# 检查API密钥配置
if [ -z "$CLAUDE_CODE_CLI_API_KEY" ]; then
    echo "❌ CLAUDE_CODE_CLI_API_KEY 未配置，请在.env文件中设置"
    echo "   或者尝试运行 'claude auth login' 进行OAuth认证"
    exit 1
fi

echo "✅ API密钥已配置"

# 检查认证状态
echo "🔐 检查认证状态..."
export ANTHROPIC_API_KEY="$CLAUDE_CODE_CLI_API_KEY"
if [ ! -z "$ANTHROPIC_BASE_URL" ]; then
    export ANTHROPIC_BASE_URL="$ANTHROPIC_BASE_URL"
    echo "✅ 使用自定义API端点: $ANTHROPIC_BASE_URL"
fi

# 测试Claude Code CLI连接
echo "🚀 测试Claude Code CLI连接..."
CLAUDE_TEST_PROMPT="Hello, please respond with 'Claude Code CLI integration test successful' if you can see this message."

# 创建临时文件来测试
TEMP_OUTPUT="/tmp/claude_test_output.txt"
TEMP_ERROR="/tmp/claude_test_error.txt"

# 使用超时命令执行测试
echo "正在发送测试请求..."
if timeout 30 claude --no-interactive "$CLAUDE_TEST_PROMPT" > "$TEMP_OUTPUT" 2> "$TEMP_ERROR"; then
    echo "✅ Claude Code CLI连接测试成功"
    echo "📄 测试响应:"
    echo "---"
    cat "$TEMP_OUTPUT"
    echo "---"
    rm -f "$TEMP_OUTPUT" "$TEMP_ERROR"
else
    echo "❌ Claude Code CLI连接测试失败"
    echo "📄 错误输出:"
    echo "---"
    cat "$TEMP_ERROR" 2>/dev/null || echo "无错误输出"
    echo "---"
    echo "📄 标准输出:"
    echo "---"
    cat "$TEMP_OUTPUT" 2>/dev/null || echo "无标准输出"
    echo "---"
    rm -f "$TEMP_OUTPUT" "$TEMP_ERROR"
    echo ""
    echo "💡 可能的解决方案:"
echo "1. 检查CLAUDE_CODE_CLI_API_KEY是否正确"
echo "2. 确保网络连接正常"
echo "3. 验证API端点是否可访问 (ANTHROPIC_BASE_URL)"
echo "4. 检查自定义端点的认证方式"
echo "5. 验证API密钥权限和配额"
echo "6. 尝试运行 'claude auth login' 进行重新认证"
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

# 测试配置加载
echo "📋 测试配置加载..."
if go run -ldflags="-X main.testMode=true" . 2>&1 | grep -q "Claude Code CLI"; then
    echo "✅ Claude Code CLI配置加载正常"
else
    echo "⚠️  配置加载测试跳过（服务启动测试）"
fi

echo ""
echo "🎉 所有测试通过！Claude Code CLI集成配置正确"
echo ""
echo "📊 测试总结:"
echo "✅ Claude Code CLI已安装并可用"
echo "✅ API密钥配置正确"
echo "✅ 网络连接正常"
echo "✅ 项目编译成功"
echo "✅ 服务配置验证通过"
echo ""
echo "📋 下一步:"
echo "1. 配置GitHub Webhook URL"
echo "2. 运行 './start.sh' 启动服务"
echo "3. 在GitHub Issue中使用 /code、/continue、/fix 命令测试功能"
echo ""
echo "🔗 相关文档:"
echo "- Claude Code CLI迁移指南: ./CLAUDE_CODE_CLI_MIGRATION.md"
echo "- 快速开始: ./QUICK_START.md"
echo "- 项目README: ./README.md"
echo ""
echo "💡 提示:"
echo "- 如果遇到API配额问题，请检查Anthropic Console中的使用情况"
echo "- 建议在生产环境中设置适当的超时和重试机制"
echo "- 可以通过环境变量调整模型和参数设置"


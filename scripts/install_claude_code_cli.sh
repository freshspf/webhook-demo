#!/bin/bash

# Claude Code CLI 安装脚本
# 用于将项目从Gemini CLI迁移到Claude Code CLI

set -e

echo "🚀 开始安装 Claude Code CLI..."

# 检查Node.js是否已安装
if ! command -v node &> /dev/null; then
    echo "❌ Node.js 未安装。请先安装Node.js (版本 >= 18)"
    echo "   Ubuntu/Debian: sudo apt-get install nodejs npm"
    echo "   CentOS/RHEL: sudo yum install nodejs npm"
    echo "   或访问: https://nodejs.org/"
    exit 1
fi

# 检查Node.js版本
NODE_VERSION=$(node --version | cut -d'v' -f2)
REQUIRED_VERSION="18.0.0"

if ! node -e "process.exit(require('semver').gte('$NODE_VERSION', '$REQUIRED_VERSION') ? 0 : 1)" 2>/dev/null; then
    # 如果semver包不存在，使用简单的版本比较
    if [[ "$(printf '%s\n' "$REQUIRED_VERSION" "$NODE_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]]; then
        echo "❌ Node.js 版本过低 (当前: v$NODE_VERSION, 需要: v$REQUIRED_VERSION+)"
        echo "   请更新到Node.js 18或更高版本"
        exit 1
    fi
fi

echo "✅ Node.js 版本检查通过 (v$NODE_VERSION)"

# 检查系统要求
echo "🔍 检查系统要求..."

# 检查RAM（可选检查）
if command -v free &> /dev/null; then
    TOTAL_RAM=$(free -m | awk 'NR==2{printf "%.0f", $2/1024}')
    if [ "$TOTAL_RAM" -lt 4 ]; then
        echo "⚠️  警告: 系统RAM少于4GB (当前: ${TOTAL_RAM}GB)，可能影响性能"
    else
        echo "✅ 系统RAM检查通过 (${TOTAL_RAM}GB)"
    fi
fi

# 检查Git版本（可选）
if command -v git &> /dev/null; then
    GIT_VERSION=$(git --version | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
    echo "✅ Git 已安装 (v$GIT_VERSION)"
else
    echo "⚠️  警告: Git 未安装，建议安装以获得更好的体验"
fi

# 安装Claude Code CLI
echo "📦 正在安装 Claude Code CLI..."
if npm install -g @anthropic-ai/claude-code; then
    echo "✅ Claude Code CLI 安装成功!"
else
    echo "❌ Claude Code CLI 安装失败"
    echo "   请检查npm权限或尝试使用sudo:"
    echo "   sudo npm install -g @anthropic-ai/claude-code"
    echo ""
    echo "   如果仍然失败，请尝试:"
    echo "   1. 清理npm缓存: npm cache clean --force"
    echo "   2. 检查网络连接"
    echo "   3. 使用不同的registry: npm install -g @anthropic-ai/claude-code --registry=https://registry.npmjs.org/"
    exit 1
fi

# 验证安装
echo "🔍 验证安装..."
if claude --version; then
    echo "✅ Claude Code CLI 验证成功!"
else
    echo "❌ Claude Code CLI 验证失败"
    echo "   请检查PATH环境变量或重新安装"
    echo "   尝试运行: which claude"
    exit 1
fi

# 检查认证状态
echo "🔐 检查认证状态..."
if claude auth status &>/dev/null; then
    echo "✅ Claude Code CLI 已认证"
else
    echo "⚠️  Claude Code CLI 尚未认证"
    echo "   请运行以下命令进行认证:"
    echo "   claude auth login"
    echo ""
    echo "   或者设置API密钥:"
    echo "   export ANTHROPIC_API_KEY=your_api_key_here"
fi

echo ""
echo "🎉 Claude Code CLI 安装完成!"
echo ""
echo "📋 下一步操作:"
echo "1. 配置环境变量 (复制 config.env.example 到 .env 并填写配置)"
echo "2. 设置 CLAUDE_CODE_CLI_API_KEY (从 Anthropic Console 获取)"
echo "3. 如果尚未认证，运行 'claude auth login' 进行认证"
echo "4. 运行 './scripts/test_claude_code_cli.sh' 测试集成"
echo "5. 运行 './start.sh' 启动服务"
echo ""
echo "🔗 有用的链接:"
echo "   - Anthropic Console: https://console.anthropic.com/"
echo "   - Claude Code CLI 文档: https://docs.anthropic.com/zh-CN/docs/claude-code/overview"
echo "   - 项目README: ./README.md"
echo ""
echo "💡 常见问题解决:"
echo "   - 认证问题: 运行 'claude auth login' 或设置 ANTHROPIC_API_KEY"
echo "   - 权限问题: 确保账户在 console.anthropic.com 上已激活计费"
echo "   - 网络问题: 检查防火墙和代理设置"
echo ""
echo "🚀 准备就绪！现在您可以使用Claude Code CLI进行AI辅助开发了！"
#!/bin/bash

# Gemini CLI 安装脚本
# 用于将项目从Claude API迁移到Gemini CLI

set -e

echo "🚀 开始安装 Gemini CLI..."

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

if ! [[ "$NODE_VERSION" > "$REQUIRED_VERSION" || "$NODE_VERSION" == "$REQUIRED_VERSION" ]]; then
    echo "❌ Node.js 版本过低 (当前: v$NODE_VERSION, 需要: v$REQUIRED_VERSION+)"
    echo "   请更新到Node.js 18或更高版本"
    exit 1
fi

echo "✅ Node.js 版本检查通过 (v$NODE_VERSION)"

# 安装Gemini CLI
echo "📦 正在安装 Gemini CLI..."
if npm install -g @google/gemini-cli; then
    echo "✅ Gemini CLI 安装成功!"
else
    echo "❌ Gemini CLI 安装失败"
    echo "   请检查npm权限或尝试使用sudo:"
    echo "   sudo npm install -g @google/gemini-cli"
    exit 1
fi

# 验证安装
echo "🔍 验证安装..."
if gemini --version; then
    echo "✅ Gemini CLI 验证成功!"
else
    echo "❌ Gemini CLI 验证失败"
    echo "   请检查PATH环境变量或重新安装"
    exit 1
fi

echo ""
echo "🎉 Gemini CLI 安装完成!"
echo ""
echo "📋 下一步操作:"
echo "1. 配置环境变量 (复制 config.env.example 到 .env 并填写配置)"
echo "2. 设置 GEMINI_API_KEY (从 Google AI Studio 获取)"
echo "3. (可选) 设置 GOOGLE_CLOUD_PROJECT (如果使用Google Cloud)"
echo "4. 运行 './start.sh' 启动服务"
echo ""
echo "🔗 有用的链接:"
echo "   - Google AI Studio: https://aistudio.google.com/"
echo "   - Gemini CLI 文档: https://www.geminicli.io/"
echo "   - 项目README: ./README.md"
echo ""
echo "💡 如果遇到认证问题，可以运行 'gemini' 命令进行首次登录设置"

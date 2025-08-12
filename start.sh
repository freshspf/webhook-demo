#!/bin/bash

# Webhook 服务启动脚本

echo "=== Webhook 服务管理 ==="

# 检查服务是否已运行
if pgrep -f "webhook-demo" > /dev/null; then
    echo "❌ 服务已在运行，请先停止现有服务"
    echo "使用命令: ./stop.sh"
    exit 1
fi

# 检查 .env 文件
if [ ! -f ".env" ]; then
    echo "❌ .env 文件不存在"
    echo "请复制 config.env.example 为 .env 并配置相应参数"
    exit 1
fi

# 读取端口配置（只取第一个匹配的行）
PORT=$(grep "^SERVER_PORT=" .env 2>/dev/null | head -1 | cut -d'=' -f2 | tr -d '\r\n' || echo "8080")
echo "📋 配置端口: $PORT"

# 检查端口是否被占用
if netstat -tlnp 2>/dev/null | grep ":$PORT " > /dev/null; then
    echo "❌ 端口 $PORT 已被占用"
    echo "占用进程:"
    netstat -tlnp 2>/dev/null | grep ":$PORT "
    echo
    echo "请先释放端口或修改 .env 文件中的 SERVER_PORT 使用其他端口"
    exit 1
fi

echo "✅ 环境检查通过"
echo "🚀 启动 webhook 服务..."

# 选择运行模式
echo "选择运行模式:"
echo "1. 后台运行 (日志保存到文件)"
echo "2. 前台运行 (日志实时输出到终端)"
read -p "请选择 (1-2): " mode

case $mode in
    1)
        echo "后台运行模式..."
        # 启动服务
        nohup go run main.go > webhook.log 2>&1 &
        PID=$!

        # 等待服务启动
        sleep 3

        # 检查服务是否成功启动
        if kill -0 $PID 2>/dev/null; then
            echo "✅ 服务启动成功！PID: $PID"
            echo "📝 日志文件: webhook.log"
            echo "🌐 服务地址: http://localhost:$PORT"
            echo "🔗 Webhook 端点: http://localhost:$PORT/webhook"
            echo
            echo "查看日志: tail -f webhook.log"
            echo "停止服务: ./stop.sh"
        else
            echo "❌ 服务启动失败"
            echo "查看错误日志: tail -20 webhook.log"
            exit 1
        fi
        ;;
    2)
        echo "前台运行模式..."
        echo "🌐 服务地址: http://localhost:$PORT"
        echo "🔗 Webhook 端点: http://localhost:$PORT/webhook"
        echo "按 Ctrl+C 停止服务"
        echo
        # 前台运行，日志直接输出到终端
        go run main.go
        ;;
    *)
        echo "❌ 无效选择，使用默认后台模式"
        # 启动服务
        nohup go run main.go > webhook.log 2>&1 &
        PID=$!

        # 等待服务启动
        sleep 3

        # 检查服务是否成功启动
        if kill -0 $PID 2>/dev/null; then
            echo "✅ 服务启动成功！PID: $PID"
            echo "📝 日志文件: webhook.log"
            echo "🌐 服务地址: http://localhost:$PORT"
            echo "🔗 Webhook 端点: http://localhost:$PORT/webhook"
            echo
            echo "查看日志: tail -f webhook.log"
            echo "停止服务: ./stop.sh"
        else
            echo "❌ 服务启动失败"
            echo "查看错误日志: tail -20 webhook.log"
            exit 1
        fi
        ;;
esac
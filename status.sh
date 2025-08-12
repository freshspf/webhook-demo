#!/bin/bash

# Webhook 服务状态检查脚本

echo "=== Webhook 服务状态 ==="

# 读取端口配置（只取第一个匹配的行）
PORT=$(grep "^SERVER_PORT=" .env 2>/dev/null | head -1 | cut -d'=' -f2 | tr -d '\r\n' || echo "8080")

# 检查服务进程
PIDS=$(pgrep -f "webhook-demo" | grep -v grep)
if [ -n "$PIDS" ]; then
    echo "✅ 服务正在运行"
    echo "进程ID: $PIDS"
    for PID in $PIDS; do
        echo "  进程 $PID: $(ps -p $PID -o pid,ppid,cmd --no-headers)"
    done
else
    echo "❌ 服务未运行"
fi
echo

# 检查端口监听
if netstat -tlnp 2>/dev/null | grep ":$PORT " > /dev/null; then
    echo "✅ 端口 $PORT 正在监听"
    netstat -tlnp 2>/dev/null | grep ":$PORT "
else
    echo "❌ 端口 $PORT 未监听"
fi
echo

# 检查服务响应
if curl -s --connect-timeout 3 http://localhost:$PORT/health > /dev/null; then
    echo "✅ 服务响应正常"
    echo "健康检查响应:"
    curl -s http://localhost:$PORT/health | jq . 2>/dev/null || curl -s http://localhost:$PORT/health
else
    echo "❌ 服务无响应"
fi
echo

# 显示最近的日志
if [ -f "webhook.log" ]; then
    echo "📝 最近日志 (最后10行):"
    tail -10 webhook.log
else
    echo "📝 日志文件不存在"
fi
echo

# 显示服务信息
echo "🌐 服务信息:"
echo "  本地地址: http://localhost:$PORT"
echo "  Webhook端点: http://localhost:$PORT/webhook"
echo "  健康检查: http://localhost:$PORT/health"
echo "  日志文件: webhook.log"
echo

# 显示管理命令
echo "🔧 管理命令:"
echo "  启动服务: ./start.sh"
echo "  停止服务: ./stop.sh"
echo "  查看日志: tail -f webhook.log"
echo "  诊断问题: ./diagnose_webhook.sh"
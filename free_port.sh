#!/bin/bash

# 端口释放脚本

if [ $# -eq 0 ]; then
    echo "用法: $0 <端口号>"
    echo "示例: $0 8080"
    exit 1
fi

PORT=$1

echo "=== 释放端口 $PORT ==="

# 检查端口是否被占用
if ! netstat -tlnp 2>/dev/null | grep ":$PORT " > /dev/null; then
    echo "✅ 端口 $PORT 未被占用"
    exit 0
fi

echo "端口 $PORT 被占用:"
netstat -tlnp 2>/dev/null | grep ":$PORT "

echo
echo "选择释放方式:"
echo "1. 优雅停止 (SIGTERM)"
echo "2. 强制杀死 (SIGKILL)"
echo "3. 查看进程详情"
echo "4. 取消"

read -p "请选择 (1-4): " choice

case $choice in
    1)
        echo "正在优雅停止进程..."
        PIDS=$(netstat -tlnp 2>/dev/null | grep ":$PORT " | awk '{print $7}' | cut -d'/' -f1)
        for PID in $PIDS; do
            echo "停止进程 $PID..."
            kill $PID
        done
        
        # 等待进程结束
        sleep 3
        
        if netstat -tlnp 2>/dev/null | grep ":$PORT " > /dev/null; then
            echo "⚠️  优雅停止失败，进程仍在运行"
        else
            echo "✅ 端口 $PORT 已释放"
        fi
        ;;
    2)
        echo "正在强制杀死进程..."
        PIDS=$(netstat -tlnp 2>/dev/null | grep ":$PORT " | awk '{print $7}' | cut -d'/' -f1)
        for PID in $PIDS; do
            echo "强制杀死进程 $PID..."
            kill -9 $PID
        done
        
        sleep 2
        
        if netstat -tlnp 2>/dev/null | grep ":$PORT " > /dev/null; then
            echo "❌ 强制杀死失败，端口仍被占用"
        else
            echo "✅ 端口 $PORT 已释放"
        fi
        ;;
    3)
        echo "进程详情:"
        PIDS=$(netstat -tlnp 2>/dev/null | grep ":$PORT " | awk '{print $7}' | cut -d'/' -f1)
        for PID in $PIDS; do
            echo "进程 $PID:"
            ps -p $PID -f
            echo
        done
        ;;
    4)
        echo "操作已取消"
        exit 0
        ;;
    *)
        echo "❌ 无效选择"
        exit 1
        ;;
esac

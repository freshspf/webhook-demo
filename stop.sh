#!/bin/bash

# Webhook 服务停止脚本

echo "=== 停止 Webhook 服务 ==="

# 读取端口配置（只取第一个匹配的行）
PORT=$(grep "^SERVER_PORT=" .env 2>/dev/null | head -1 | cut -d'=' -f2 | tr -d '\r\n' || echo "8080")
echo "📋 配置端口: $PORT"

# 查找 webhook 服务进程（更精确的匹配）
PIDS=$(pgrep -f "webhook-demo" | grep -v grep)

if [ -z "$PIDS" ]; then
    echo "ℹ️  没有找到运行中的 webhook 服务"
    
    # 检查端口是否被占用
    if netstat -tlnp 2>/dev/null | grep ":$PORT " > /dev/null; then
        echo "⚠️  端口 $PORT 仍被占用，但未找到 webhook-demo 进程"
        echo "占用进程:"
        netstat -tlnp 2>/dev/null | grep ":$PORT "
        echo
        echo "是否要强制释放端口？(y/N)"
        read -r response
        if [[ "$response" =~ ^[Yy]$ ]]; then
            OCCUPYING_PID=$(netstat -tlnp 2>/dev/null | grep ":$PORT " | awk '{print $7}' | cut -d'/' -f1)
            if [ -n "$OCCUPYING_PID" ]; then
                echo "正在强制终止占用端口的进程 $OCCUPYING_PID..."
                kill -9 $OCCUPYING_PID
                sleep 2
                if ! netstat -tlnp 2>/dev/null | grep ":$PORT " > /dev/null; then
                    echo "✅ 端口 $PORT 已释放"
                else
                    echo "❌ 端口 $PORT 释放失败"
                fi
            fi
        fi
    else
        echo "✅ 端口 $PORT 未被占用"
    fi
    exit 0
fi

echo "找到运行中的服务进程: $PIDS"

# 停止进程
for PID in $PIDS; do
    echo "正在停止进程 $PID..."
    
    # 显示进程信息
    ps -p $PID -o pid,ppid,cmd --no-headers 2>/dev/null || echo "无法获取进程信息"
    
    # 尝试优雅停止
    kill $PID
    
    # 等待进程结束
    for i in {1..10}; do
        if ! kill -0 $PID 2>/dev/null; then
            echo "✅ 进程 $PID 已停止"
            break
        fi
        echo "等待进程 $PID 停止... ($i/10)"
        sleep 1
    done
    
    # 如果进程仍在运行，强制终止
    if kill -0 $PID 2>/dev/null; then
        echo "⚠️  强制终止进程 $PID"
        kill -9 $PID
        sleep 1
        
        if ! kill -0 $PID 2>/dev/null; then
            echo "✅ 进程 $PID 已强制停止"
        else
            echo "❌ 进程 $PID 停止失败"
        fi
    fi
done

# 检查端口是否释放
echo
echo "检查端口释放状态..."
if netstat -tlnp 2>/dev/null | grep ":$PORT " > /dev/null; then
    echo "⚠️  端口 $PORT 仍被占用"
    echo "占用进程:"
    netstat -tlnp 2>/dev/null | grep ":$PORT "
else
    echo "✅ 端口 $PORT 已释放"
fi

echo "✅ 服务停止操作完成"
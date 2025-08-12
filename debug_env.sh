#!/bin/bash

echo "=== .env 文件调试 ==="
echo

# 检查 .env 文件是否存在
if [ -f ".env" ]; then
    echo "✅ .env 文件存在"
    echo "文件大小: $(ls -lh .env | awk '{print $5}')"
    echo
    echo "文件内容:"
    echo "=================="
    cat .env
    echo "=================="
    echo
else
    echo "❌ .env 文件不存在"
    exit 1
fi

# 检查 SERVER_PORT 配置
echo "检查 SERVER_PORT 配置:"
SERVER_PORT_LINE=$(grep "^SERVER_PORT=" .env)
if [ -n "$SERVER_PORT_LINE" ]; then
    echo "找到配置行: '$SERVER_PORT_LINE'"
    PORT_VALUE=$(echo "$SERVER_PORT_LINE" | cut -d'=' -f2)
    echo "端口值: '$PORT_VALUE'"
    echo "端口值长度: ${#PORT_VALUE}"
    echo "端口值是否为空: $([ -z "$PORT_VALUE" ] && echo "是" || echo "否")"
else
    echo "❌ 未找到 SERVER_PORT 配置"
fi
echo

# 测试脚本中的端口读取逻辑
echo "测试端口读取逻辑:"
PORT=$(grep "^SERVER_PORT=" .env 2>/dev/null | cut -d'=' -f2 || echo "8080")
echo "脚本读取的端口: '$PORT'"
echo "端口长度: ${#PORT}"
echo

# 检查实际运行的进程
echo "检查实际运行的进程:"
PIDS=$(pgrep -f "webhook-demo")
if [ -n "$PIDS" ]; then
    echo "找到进程: $PIDS"
    for PID in $PIDS; do
        echo "进程 $PID 详情:"
        ps -p $PID -o pid,ppid,cmd --no-headers
    done
else
    echo "❌ 未找到 webhook-demo 进程"
fi
echo

# 检查端口监听
echo "检查端口监听:"
netstat -tlnp 2>/dev/null | grep LISTEN | head -10
echo

# 检查特定端口
if [ -n "$PORT" ]; then
    echo "检查端口 $PORT 的监听状态:"
    netstat -tlnp 2>/dev/null | grep ":$PORT " || echo "端口 $PORT 未监听"
fi

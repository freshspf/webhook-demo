#!/bin/bash

# 网络访问诊断脚本

echo "=== 网络访问诊断 ==="

# 读取端口配置
PORT=$(grep "^SERVER_PORT=" .env 2>/dev/null | head -1 | cut -d'=' -f2 | tr -d '\r\n' || echo "8080")

echo "📋 配置端口: $PORT"
echo

# 1. 检查服务是否运行
echo "1. 检查服务状态..."
if pgrep -f "webhook-demo" > /dev/null; then
    echo "✅ 服务正在运行"
    ps aux | grep webhook-demo | grep -v grep
else
    echo "❌ 服务未运行"
    echo "请先启动服务: ./start.sh"
    exit 1
fi
echo

# 2. 检查端口监听
echo "2. 检查端口监听..."
if netstat -tlnp 2>/dev/null | grep ":$PORT " > /dev/null; then
    echo "✅ 端口 $PORT 正在监听"
    netstat -tlnp 2>/dev/null | grep ":$PORT "
else
    echo "❌ 端口 $PORT 未监听"
    exit 1
fi
echo

# 3. 检查绑定地址
echo "3. 检查绑定地址..."
LISTENING_ADDR=$(netstat -tlnp 2>/dev/null | grep ":$PORT " | awk '{print $4}')
echo "监听地址: $LISTENING_ADDR"

if echo "$LISTENING_ADDR" | grep -q "0.0.0.0\|::"; then
    echo "✅ 服务绑定到所有网络接口"
else
    echo "⚠️  服务可能只绑定到本地接口"
fi
echo

# 4. 检查防火墙
echo "4. 检查防火墙状态..."
if command -v firewall-cmd &> /dev/null; then
    FIREWALL_STATUS=$(systemctl is-active firewalld 2>/dev/null || echo "unknown")
    echo "防火墙状态: $FIREWALL_STATUS"
    
    if [ "$FIREWALL_STATUS" = "active" ]; then
        echo "⚠️  防火墙正在运行，检查端口是否开放..."
        if firewall-cmd --query-port=$PORT/tcp 2>/dev/null; then
            echo "✅ 端口 $PORT 已在防火墙中开放"
        else
            echo "❌ 端口 $PORT 未在防火墙中开放"
            echo "建议执行: firewall-cmd --permanent --add-port=$PORT/tcp && firewall-cmd --reload"
        fi
    fi
else
    echo "ℹ️  未检测到 firewalld"
fi
echo

# 5. 检查 iptables
echo "5. 检查 iptables 规则..."
if command -v iptables &> /dev/null; then
    echo "iptables 规则 (INPUT 链):"
    iptables -L INPUT -n | grep -E "(ACCEPT|DROP|REJECT)" | head -5
else
    echo "ℹ️  未检测到 iptables"
fi
echo

# 6. 获取网络信息
echo "6. 网络信息..."
echo "本机IP地址:"
ip addr show | grep -E "inet .* global" | awk '{print $2}' | cut -d'/' -f1

echo "公网IP地址:"
PUBLIC_IP=$(curl -s --connect-timeout 5 ifconfig.me 2>/dev/null || echo "无法获取")
echo "$PUBLIC_IP"
echo

# 7. 本地连接测试
echo "7. 本地连接测试..."
if curl -s --connect-timeout 5 http://localhost:$PORT/health > /dev/null; then
    echo "✅ 本地连接正常"
    curl -s http://localhost:$PORT/health | jq . 2>/dev/null || curl -s http://localhost:$PORT/health
else
    echo "❌ 本地连接失败"
fi
echo

# 8. 网络连接测试
echo "8. 网络连接测试..."
if [ -n "$PUBLIC_IP" ] && [ "$PUBLIC_IP" != "无法获取" ]; then
    echo "测试从公网访问: http://$PUBLIC_IP:$PORT/health"
    if curl -s --connect-timeout 10 http://$PUBLIC_IP:$PORT/health > /dev/null; then
        echo "✅ 公网访问正常"
    else
        echo "❌ 公网访问失败"
    fi
else
    echo "⚠️  无法获取公网IP，跳过公网测试"
fi
echo

# 9. 提供解决方案
echo "=== 解决方案 ==="
echo "如果其他机器无法访问，请尝试以下步骤："
echo
echo "1. 开放防火墙端口:"
echo "   firewall-cmd --permanent --add-port=$PORT/tcp"
echo "   firewall-cmd --reload"
echo
echo "2. 检查云服务器安全组:"
echo "   - 登录云服务商控制台"
echo "   - 找到安全组设置"
echo "   - 添加入站规则：端口 $PORT，协议 TCP"
echo
echo "3. 检查路由器/防火墙:"
echo "   - 确保端口 $PORT 已开放"
echo "   - 检查端口转发设置"
echo
echo "4. 测试连接:"
echo "   curl http://$PUBLIC_IP:$PORT/health"
echo
echo "5. 使用 ngrok 进行端口转发（临时解决方案）:"
echo "   ngrok http $PORT"

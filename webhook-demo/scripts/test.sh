#!/bin/bash

# GitHub Webhook Demo 测试脚本

set -e

echo "🧪 GitHub Webhook Demo 测试"

SERVER_URL="http://localhost:8080"

# 检查服务是否运行
echo "🔍 检查服务状态..."
if curl -s "$SERVER_URL/health" > /dev/null; then
    echo "✅ 服务运行正常"
else
    echo "❌ 服务未运行，请先启动服务: ./scripts/start.sh"
    exit 1
fi

# 测试健康检查端点
echo ""
echo "🏥 测试健康检查端点..."
echo "请求: GET $SERVER_URL/health"
curl -s "$SERVER_URL/health" | jq '.' || echo "响应不是有效的JSON"

# 测试API信息端点
echo ""
echo "📋 测试API信息端点..."
echo "请求: GET $SERVER_URL/"
curl -s "$SERVER_URL/" | jq '.' || echo "响应不是有效的JSON"

# 测试模拟Webhook请求
echo ""
echo "🔗 测试模拟Webhook请求..."

# 模拟ping事件
echo "测试ping事件..."
PING_PAYLOAD='{
  "zen": "Responsive is better than fast.",
  "hook_id": 12345678,
  "hook": {
    "type": "Repository",
    "id": 12345678,
    "name": "web",
    "active": true,
    "events": ["push", "pull_request"],
    "config": {
      "content_type": "json",
      "insecure_ssl": "0",
      "url": "http://localhost:8080/webhook"
    }
  },
  "repository": {
    "id": 123456789,
    "name": "webhook-demo",
    "full_name": "test/webhook-demo",
    "html_url": "https://github.com/test/webhook-demo"
  }
}'

curl -X POST "$SERVER_URL/webhook" \
  -H "Content-Type: application/json" \
  -H "X-GitHub-Event: ping" \
  -H "X-GitHub-Delivery: 12345678-1234-1234-1234-123456789012" \
  -H "User-Agent: GitHub-Hookshot/abc123" \
  -d "$PING_PAYLOAD" \
  | jq '.' 2>/dev/null || echo "响应不是有效的JSON"

echo ""

# 模拟issue_comment事件
echo "测试issue_comment事件..."
COMMENT_PAYLOAD='{
  "action": "created",
  "issue": {
    "id": 1,
    "number": 1,
    "title": "测试Issue",
    "body": "这是一个测试Issue",
    "state": "open",
    "html_url": "https://github.com/test/webhook-demo/issues/1",
    "user": {
      "id": 1,
      "login": "testuser",
      "html_url": "https://github.com/testuser",
      "avatar_url": "https://github.com/testuser.png"
    },
    "created_at": "2023-01-01T00:00:00Z",
    "updated_at": "2023-01-01T00:00:00Z"
  },
  "comment": {
    "id": 1,
    "body": "/help",
    "user": {
      "id": 1,
      "login": "testuser",
      "html_url": "https://github.com/testuser",
      "avatar_url": "https://github.com/testuser.png"
    },
    "html_url": "https://github.com/test/webhook-demo/issues/1#issuecomment-1",
    "created_at": "2023-01-01T00:00:00Z",
    "updated_at": "2023-01-01T00:00:00Z"
  },
  "repository": {
    "id": 123456789,
    "name": "webhook-demo",
    "full_name": "test/webhook-demo",
    "html_url": "https://github.com/test/webhook-demo",
    "clone_url": "https://github.com/test/webhook-demo.git",
    "ssh_url": "git@github.com:test/webhook-demo.git"
  },
  "sender": {
    "id": 1,
    "login": "testuser",
    "html_url": "https://github.com/testuser",
    "avatar_url": "https://github.com/testuser.png"
  }
}'

curl -X POST "$SERVER_URL/webhook" \
  -H "Content-Type: application/json" \
  -H "X-GitHub-Event: issue_comment" \
  -H "X-GitHub-Delivery: 12345678-1234-1234-1234-123456789013" \
  -H "User-Agent: GitHub-Hookshot/abc123" \
  -d "$COMMENT_PAYLOAD" \
  | jq '.' 2>/dev/null || echo "响应不是有效的JSON"

echo ""
echo "✅ 测试完成！"
echo ""
echo "💡 要测试真实的GitHub事件，请:"
echo "   1. 在GitHub仓库中配置Webhook"
echo "   2. 创建Issue并添加评论: /help"
echo "   3. 查看服务器日志和GitHub响应"
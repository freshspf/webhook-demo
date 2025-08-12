# Claude API 集成指南

本文档说明如何在GitHub Webhook Demo项目中集成Claude API来实现真正的AI代码生成功能。

## 🎯 功能概述

通过集成Claude API，您的webhook服务现在可以：

- 🤖 **智能代码生成**: 根据需求描述生成高质量代码
- 🔄 **继续开发**: 基于现有代码继续开发新功能
- 🔧 **代码修复**: 分析并修复代码中的问题
- 💬 **智能回复**: 在GitHub Issue/PR中提供专业的AI回复

## 🚀 快速开始

### 1. 获取Claude API密钥

1. 访问 [Anthropic Console](https://console.anthropic.com/)
2. 注册或登录您的账户
3. 创建新的API密钥
4. 复制API密钥（格式：`sk-ant-api03-...`）

### 2. 配置环境变量

```bash
# 复制配置文件
cp config.env.example .env

# 编辑配置文件，添加Claude API配置
CLAUDE_API_KEY=sk-ant-api03-your-api-key-here
CLAUDE_MODEL=claude-3-5-sonnet-20241022
CLAUDE_MAX_TOKENS=4000
```

### 3. 启动服务

```bash
# 设置环境变量
source .env

# 启动服务
go run main.go
```

## 📋 支持的命令

### `/code <需求描述>`
生成代码实现指定功能

**示例:**
```
/code 实现用户登录功能，包含用户名密码验证和JWT token生成
```

**响应:**
- 分析需求
- 调用Claude API生成代码
- 在GitHub中回复生成的代码

### `/continue [说明]`
继续当前的开发任务

**示例:**
```
/continue 添加错误处理和日志记录
```

**响应:**
- 基于现有代码继续开发
- 保持代码风格一致性
- 添加新功能或改进

### `/fix <问题描述>`
修复指定的代码问题

**示例:**
```
/fix 修复空指针异常，在用户对象为nil时进行处理
```

**响应:**
- 分析问题根本原因
- 生成修复方案
- 提供修复后的代码

### `/help`
显示帮助信息

**响应:**
- 显示所有支持的命令
- 提供使用示例
- 说明工作流程

## 🔧 技术实现

### 架构设计

```
GitHub Webhook → 事件处理器 → Claude API → 代码生成 → GitHub回复
```

### 核心组件

1. **ClaudeService** (`internal/services/claude.go`)
   - 封装Claude API调用
   - 处理不同类型的代码生成请求
   - 构建优化的提示词

2. **EventProcessor** (`internal/services/event_processor.go`)
   - 解析GitHub事件
   - 提取命令和参数
   - 调用Claude服务
   - 生成响应

3. **配置管理** (`internal/config/config.go`)
   - Claude API密钥管理
   - 模型参数配置
   - 环境变量处理

### API调用流程

1. **接收Webhook**: GitHub发送事件到webhook端点
2. **验证签名**: 使用HMAC-SHA256验证请求来源
3. **解析事件**: 根据事件类型解析JSON payload
4. **提取命令**: 从Issue/PR评论中提取命令
5. **构建上下文**: 收集项目相关信息
6. **调用Claude**: 发送请求到Claude API
7. **处理响应**: 解析AI生成的代码
8. **创建回复**: 在GitHub中创建评论回复

## 🛠️ 自定义配置

### 模型选择

支持不同的Claude模型：

```bash
# 最新版本（推荐）
CLAUDE_MODEL=claude-3-5-sonnet-20241022

# 其他可用模型
CLAUDE_MODEL=claude-3-opus-20240229
CLAUDE_MODEL=claude-3-sonnet-20240229
CLAUDE_MODEL=claude-3-haiku-20240307
```

### Token限制

```bash
# 设置最大输出token数
CLAUDE_MAX_TOKENS=4000  # 默认值
CLAUDE_MAX_TOKENS=8000  # 更长的响应
CLAUDE_MAX_TOKENS=2000  # 更短的响应
```

### 提示词优化

您可以在 `internal/services/claude.go` 中自定义提示词：

```go
func (cs *ClaudeService) buildCodeGenerationPrompt(requirement string, context string) string {
    // 自定义提示词模板
    return fmt.Sprintf(`您的自定义提示词模板...
    
    需求: %s
    上下文: %s
    
    请生成代码:`)
}
```

## 🔒 安全考虑

### API密钥安全

1. **环境变量**: 始终使用环境变量存储API密钥
2. **不要提交**: 确保 `.env` 文件在 `.gitignore` 中
3. **定期轮换**: 定期更新API密钥
4. **权限最小化**: 只授予必要的权限

### 请求验证

1. **签名验证**: 验证GitHub Webhook签名
2. **来源检查**: 检查请求来源IP（可选）
3. **频率限制**: 实现API调用频率限制

### 内容安全

1. **输入验证**: 验证用户输入
2. **输出过滤**: 过滤AI生成的代码
3. **沙箱环境**: 在安全环境中测试生成的代码

## 📊 监控和日志

### 日志记录

服务会记录详细的日志信息：

```
2024/01/15 10:30:00 调用Claude API，模型: claude-3-5-sonnet-20241022, 最大Token: 4000
2024/01/15 10:30:05 Claude API调用成功，输入Token: 150, 输出Token: 800
2024/01/15 10:30:06 在Issue #123中创建回复成功
```

### 监控指标

建议监控以下指标：

- API调用成功率
- 响应时间
- Token使用量
- 错误率
- 用户满意度

## 🐛 故障排除

### 常见问题

1. **API密钥错误**
   ```
   错误: Claude API密钥未配置
   解决: 检查 CLAUDE_API_KEY 环境变量
   ```

2. **网络连接问题**
   ```
   错误: API调用失败: dial tcp: lookup api.anthropic.com
   解决: 检查网络连接和DNS设置
   ```

3. **配额超限**
   ```
   错误: API返回错误状态码: 429
   解决: 检查API配额使用情况
   ```

4. **模型不可用**
   ```
   错误: API返回错误状态码: 400
   解决: 检查模型名称是否正确
   ```

### 调试模式

启用详细日志：

```bash
export GIN_MODE=debug
go run main.go
```

### 测试API连接

```bash
# 测试Claude API连接
curl -X POST https://api.anthropic.com/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: $CLAUDE_API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -d '{
    "model": "claude-3-5-sonnet-20241022",
    "max_tokens": 100,
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

## 📚 相关资源

- [Claude API文档](https://docs.anthropic.com/claude/reference)
- [GitHub Webhooks文档](https://docs.github.com/en/developers/webhooks-and-events/webhooks)
- [Go HTTP客户端](https://golang.org/pkg/net/http/)
- [Gin Web框架](https://gin-gonic.com/)

## 🤝 贡献

欢迎提交Issue和Pull Request来改进Claude集成功能！

## 📄 许可证

MIT License

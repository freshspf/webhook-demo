# 自定义API端点配置指南

本指南帮助您配置自定义的Anthropic API端点，例如使用七牛云代理或其他第三方服务。

## 🚀 快速配置

### 1. 设置环境变量

在您的 `.env` 文件中添加或修改以下配置：

```bash
# 必需：Claude Code CLI API密钥
CLAUDE_CODE_CLI_API_KEY=your_auth_token_here

# 必需：自定义API端点
ANTHROPIC_BASE_URL="https://cc.qiniu.com/api/"

# 可选：模型配置
CLAUDE_CODE_CLI_MODEL=claude-3-5-sonnet-20241022
CLAUDE_CODE_CLI_MAX_TOKENS=4000
CLAUDE_CODE_CLI_TIMEOUT_SECONDS=120
```

### 2. 验证配置

运行测试脚本验证配置是否正确：

```bash
./scripts/test_claude_code_cli.sh
```

### 3. 启动服务

```bash
./start.sh
```

## 🔧 配置详解

### 环境变量说明

| 变量名 | 说明 | 是否必需 | 示例值 |
|--------|------|----------|--------|
| `CLAUDE_CODE_CLI_API_KEY` | 认证token | 是 | `your_auth_token` |
| `ANTHROPIC_BASE_URL` | 自定义API端点 | 是 | `https://cc.qiniu.com/api/` |
| `CLAUDE_CODE_CLI_MODEL` | 模型名称 | 否 | `claude-3-5-sonnet-20241022` |
| `CLAUDE_CODE_CLI_MAX_TOKENS` | 最大token数 | 否 | `4000` |
| `CLAUDE_CODE_CLI_TIMEOUT_SECONDS` | 超时时间(秒) | 否 | `120` |

### 支持的端点格式

确保您的自定义端点：

1. **兼容Anthropic API格式**
2. **支持HTTPS协议**
3. **正确处理认证headers**
4. **返回标准的API响应格式**

### 常见端点配置

#### 七牛云代理
```bash
ANTHROPIC_BASE_URL="https://cc.qiniu.com/api/"
```

#### 其他代理服务
```bash
ANTHROPIC_BASE_URL="https://your-proxy-domain.com/v1/"
```

#### 本地开发环境
```bash
ANTHROPIC_BASE_URL="http://localhost:8080/api/"
```

## 🔍 测试和验证

### 手动测试Claude Code CLI

1. **设置环境变量**：
   ```bash
   export ANTHROPIC_API_KEY="your_auth_token"
   export ANTHROPIC_BASE_URL="https://cc.qiniu.com/api/"
   ```

2. **测试连接**：
   ```bash
   claude "Hello, can you respond to test the connection?"
   ```

3. **检查日志**：
   查看是否使用了正确的端点地址。

### 项目集成测试

1. **运行集成测试**：
   ```bash
   ./scripts/test_claude_code_cli.sh
   ```

2. **检查服务日志**：
   ```bash
   tail -f webhook.log
   ```
   
   应该能看到类似输出：
   ```
   调用Claude Code CLI，模型: claude-3-5-sonnet-20241022, 最大Token: 4000, BaseURL: https://cc.qiniu.com/api/
   ```

## ❗ 故障排除

### 常见问题

#### 1. 端点连接失败
```
错误: Claude Code CLI连接测试失败
```

**检查项：**
- [ ] 端点URL是否正确
- [ ] 网络是否可达
- [ ] 端点是否支持HTTPS
- [ ] 防火墙设置

#### 2. 认证失败
```
错误: 401 Unauthorized
```

**检查项：**
- [ ] API密钥是否正确
- [ ] 端点认证方式是否正确
- [ ] 是否需要特殊的header格式

#### 3. 请求格式错误
```
错误: 400 Bad Request
```

**检查项：**
- [ ] 端点是否完全兼容Anthropic API
- [ ] 请求格式是否正确
- [ ] 模型名称是否支持

#### 4. 超时问题
```
错误: Claude Code CLI调用超时
```

**解决方案：**
- 增加超时时间：`CLAUDE_CODE_CLI_TIMEOUT_SECONDS=180`
- 检查网络延迟
- 联系端点提供商

### 调试技巧

#### 1. 启用详细日志
```bash
export ANTHROPIC_LOG_LEVEL=debug
```

#### 2. 检查网络连通性
```bash
curl -I https://cc.qiniu.com/api/
```

#### 3. 测试API格式
```bash
curl -X POST https://cc.qiniu.com/api/v1/messages \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your_token" \
  -d '{"model":"claude-3-5-sonnet-20241022","max_tokens":100,"messages":[{"role":"user","content":"Hello"}]}'
```

## 🔒 安全注意事项

1. **保护API密钥**：
   - 不要在代码中硬编码API密钥
   - 使用环境变量或安全的配置管理

2. **验证端点安全性**：
   - 确保使用HTTPS协议
   - 验证证书有效性
   - 了解数据处理和存储政策

3. **监控API使用**：
   - 定期检查API调用日志
   - 监控异常访问模式
   - 设置合理的超时和重试机制

## 📊 性能优化

1. **合理设置超时**：
   ```bash
   CLAUDE_CODE_CLI_TIMEOUT_SECONDS=120  # 根据网络情况调整
   ```

2. **优化token使用**：
   ```bash
   CLAUDE_CODE_CLI_MAX_TOKENS=4000  # 平衡性能和成本
   ```

3. **监控响应时间**：
   - 查看webhook.log中的性能指标
   - 根据实际情况调整配置

## 🆘 获取帮助

如果遇到问题：

1. **检查配置**：运行 `./scripts/test_claude_code_cli.sh`
2. **查看日志**：`tail -f webhook.log`
3. **验证端点**：联系您的API提供商
4. **网络诊断**：检查网络连接和DNS解析

---

**配置完成后，您的项目将使用自定义的API端点进行Claude Code CLI调用。**

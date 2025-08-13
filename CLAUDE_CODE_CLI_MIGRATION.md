# Claude Code CLI 迁移指南

本项目已从Gemini CLI迁移到Claude Code CLI，以提供更稳定和功能丰富的AI开发体验。

## 🚀 快速开始

### 1. 安装Claude Code CLI

运行我们提供的自动安装脚本：

```bash
./scripts/install_claude_code_cli.sh
```

或者手动安装：

```bash
# 确保Node.js版本 >= 18
node --version

# 安装Claude Code CLI
npm install -g @anthropic-ai/claude-code

# 验证安装
claude --version
```

### 2. 配置环境变量

复制配置示例文件：

```bash
cp config.env.example .env
```

编辑 `.env` 文件，配置Claude Code CLI相关参数：

```bash
# Claude Code CLI配置
CLAUDE_CODE_CLI_API_KEY=your_claude_code_cli_api_key_here
CLAUDE_CODE_CLI_MODEL=claude-3-5-sonnet-20241022
CLAUDE_CODE_CLI_MAX_TOKENS=4000
CLAUDE_CODE_CLI_TIMEOUT_SECONDS=120
ANTHROPIC_BASE_URL=https://your-custom-endpoint.com/api/  # 可选，自定义API端点
```

### 3. 获取API密钥

#### 方法一：API密钥（推荐）
1. 访问 [Anthropic Console](https://console.anthropic.com/)
2. 登录你的Anthropic账号
3. 确保账户已激活计费功能
4. 创建新的API密钥
5. 将API密钥填入 `CLAUDE_CODE_CLI_API_KEY` 环境变量

#### 方法二：OAuth认证
```bash
claude auth login
```

### 4. 测试集成

运行测试脚本验证配置：

```bash
./scripts/test_claude_code_cli.sh
```

### 5. 启动服务

```bash
./start.sh
```

## 🔄 迁移变化

### 主要改进

1. **更好的代码理解**: Claude 3.5 Sonnet在代码分析和生成方面表现卓越
2. **更稳定的服务**: 相比Gemini CLI，Claude Code CLI提供更稳定的服务
3. **更丰富的功能**: 支持更多的编程语言和开发任务
4. **更好的错误处理**: 提供更详细的错误信息和调试支持

### 代码变化

#### 服务替换
- `GeminiService` → `ClaudeCodeCLIService`
- `internal/services/gemini.go` → `internal/services/claude_code_cli.go`

#### 配置更新
- 新增 `ClaudeCodeCLIConfig` 配置结构
- 支持超时时间、模型选择等新配置项

#### 环境变量
- `GEMINI_API_KEY` → `CLAUDE_CODE_CLI_API_KEY` 
- `GEMINI_MODEL` → `CLAUDE_CODE_CLI_MODEL`
- `GEMINI_MAX_TOKENS` → `CLAUDE_CODE_CLI_MAX_TOKENS`
- `GEMINI_TIMEOUT_SECONDS` → `CLAUDE_CODE_CLI_TIMEOUT_SECONDS`

## 🛠️ 配置选项

### 模型选择

支持的Claude模型：
- `claude-3-5-sonnet-20241022` (推荐，最新版本)
- `claude-3-5-haiku-20241022`
- `claude-3-opus-20240229`

### 高级配置

```bash
# 最大Token数（建议4000-8000）
CLAUDE_CODE_CLI_MAX_TOKENS=4000

# 请求超时时间（秒）
CLAUDE_CODE_CLI_TIMEOUT_SECONDS=120

# 自定义API端点（可选）
ANTHROPIC_BASE_URL=https://your-custom-endpoint.com/api/

# 启用详细日志
CLAUDE_CODE_CLI_VERBOSE=true
```

### 自定义API端点配置

如果您使用代理服务或第三方API提供商，可以通过 `ANTHROPIC_BASE_URL` 环境变量配置自定义端点：

```bash
# 示例：使用七牛云代理
ANTHROPIC_BASE_URL="https://cc.qiniu.com/api/"

# 示例：使用其他代理服务
ANTHROPIC_BASE_URL="https://your-proxy.com/v1/"
```

**注意事项：**
- 确保自定义端点兼容Anthropic API格式
- 验证端点的安全性和可靠性
- 自定义端点可能需要不同的认证方式

## 🔧 故障排除

### 常见问题

#### 1. Claude Code CLI未安装
```
错误: Claude Code CLI未安装，请先运行: npm install -g @anthropic-ai/claude-code
```

**解决方案**: 运行安装脚本或手动安装
```bash
./scripts/install_claude_code_cli.sh
```

#### 2. Node.js版本过低
```
错误: Node.js 版本过低 (当前: v16.x.x, 需要: v18.0.0+)
```

**解决方案**: 更新Node.js到18或更高版本

#### 3. API密钥未配置
```
错误: CLAUDE_CODE_CLI_API_KEY 未配置，请在.env文件中设置
```

**解决方案**: 配置API密钥或进行OAuth认证
```bash
# 方法1: 设置API密钥
export CLAUDE_CODE_CLI_API_KEY=your_api_key

# 方法2: OAuth认证
claude auth login
```

#### 4. 认证失败
```
错误: Claude Code CLI连接测试失败
```

**解决方案**: 
1. 检查API密钥是否正确
2. 确保账户已激活计费功能
3. 验证网络连接
4. 重新进行认证

#### 5. 权限问题
```
错误: 账户没有权限访问Claude API
```

**解决方案**: 
1. 确保在 [Anthropic Console](https://console.anthropic.com/) 中已激活计费
2. 检查API密钥权限
3. 联系Anthropic支持团队

#### 6. 超时问题
```
错误: Claude Code CLI调用超时
```

**解决方案**: 增加超时时间
```bash
CLAUDE_CODE_CLI_TIMEOUT_SECONDS=180
```

### 调试技巧

#### 1. 启用详细日志
```bash
export ANTHROPIC_LOG_LEVEL=debug
```

#### 2. 手动测试Claude Code CLI
```bash
claude "Hello, can you help me with coding?"
```

#### 3. 检查认证状态
```bash
claude auth status
```

#### 4. 查看服务日志
```bash
tail -f webhook.log
```

## 🔗 有用的链接

- [Anthropic Console](https://console.anthropic.com/) - 管理API密钥和计费
- [Claude Code CLI文档](https://docs.anthropic.com/zh-CN/docs/claude-code/overview) - 官方文档
- [Node.js下载](https://nodejs.org/) - Node.js官网
- [项目GitHub](https://github.com/your-repo) - 项目仓库

## 📋 迁移检查清单

- [ ] 安装Node.js (版本 >= 18)
- [ ] 安装Claude Code CLI
- [ ] 获取Anthropic API密钥或完成OAuth认证
- [ ] 配置环境变量
- [ ] 测试Claude Code CLI连接
- [ ] 编译并启动服务
- [ ] 验证webhook响应正常
- [ ] 测试AI功能（/code、/continue、/fix命令）

## 🚧 系统要求

### 最低要求
- **操作系统**: macOS 10.15+、Ubuntu 20.04+/Debian 10+，或Windows(通过WSL)
- **Node.js**: 18.0.0+
- **RAM**: 4GB+
- **网络**: 稳定的互联网连接

### 推荐配置
- **RAM**: 8GB+
- **Git**: 2.23+（用于增强功能）
- **ripgrep**: 用于增强文件搜索

## 🆘 获取帮助

如果遇到问题：
1. 运行测试脚本: `./scripts/test_claude_code_cli.sh`
2. 检查日志文件: `tail -f webhook.log`
3. 验证环境变量配置
4. 查看本文档的故障排除部分
5. 检查Anthropic Console中的账户状态

## ⚡ 性能优化建议

1. **合理设置超时时间**: 根据网络状况调整 `CLAUDE_CODE_CLI_TIMEOUT_SECONDS`
2. **选择合适的模型**: Claude 3.5 Sonnet在代码任务上表现最佳
3. **控制Token数量**: 设置合理的 `CLAUDE_CODE_CLI_MAX_TOKENS` 以平衡性能和成本
4. **监控API使用**: 定期检查Anthropic Console中的使用情况

---

*此迁移提供了更强大的AI能力，同时保持了与之前版本相同的接口，确保无缝切换体验。*

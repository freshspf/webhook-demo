# Gemini CLI 迁移指南

本项目已从Claude API迁移到Gemini CLI，以获得更好的效果和更丰富的功能。

## 🚀 快速开始

### 1. 安装Gemini CLI

运行我们提供的自动安装脚本：

```bash
./scripts/install_gemini_cli.sh
```

或者手动安装：

```bash
# 确保Node.js版本 >= 18
node --version

# 安装Gemini CLI
npm install -g @google/gemini-cli

# 验证安装
gemini --version
```

### 2. 配置环境变量

复制配置示例文件：

```bash
cp config.env.example .env
```

编辑 `.env` 文件，配置Gemini相关参数：

```bash
# Gemini CLI配置
GEMINI_API_KEY=your_gemini_api_key_here
GEMINI_MODEL=gemini-2.0-flash-exp
GEMINI_MAX_TOKENS=8000
GOOGLE_CLOUD_PROJECT=your_gcp_project_id    # 可选
GEMINI_TIMEOUT_SECONDS=120
```

### 3. 获取API密钥

1. 访问 [Google AI Studio](https://aistudio.google.com/)
2. 登录你的Google账号
3. 创建新的API密钥
4. 将API密钥填入 `GEMINI_API_KEY` 环境变量

### 4. 启动服务

```bash
./start.sh
```

## 🔄 迁移变化

### 主要改进

1. **更强大的AI能力**: Gemini 2.0 Flash具有更强的代码理解和生成能力
2. **更低的延迟**: CLI调用比API调用更快
3. **更丰富的功能**: 支持更多的模型参数和配置选项
4. **更好的错误处理**: 更详细的错误信息和超时控制

### 代码变化

#### 服务替换
- `ClaudeService` → `GeminiService`
- `internal/services/claude.go` → `internal/services/gemini.go`

#### 配置更新
- 新增 `GeminiConfig` 配置结构
- 支持项目ID、超时时间等新配置项

#### 环境变量
- `CLAUDE_API_KEY` → `GEMINI_API_KEY` 
- `CLAUDE_MODEL` → `GEMINI_MODEL`
- `CLAUDE_MAX_TOKENS` → `GEMINI_MAX_TOKENS`
- 新增 `GOOGLE_CLOUD_PROJECT`
- 新增 `GEMINI_TIMEOUT_SECONDS`

## 🛠️ 配置选项

### 模型选择

支持的Gemini模型：
- `gemini-2.0-flash-exp` (推荐，最新实验版本)
- `gemini-1.5-pro`
- `gemini-1.5-flash`

### 高级配置

```bash
# 最大Token数（建议8000）
GEMINI_MAX_TOKENS=8000

# 请求超时时间（秒）
GEMINI_TIMEOUT_SECONDS=120

# Google Cloud项目ID（如果使用Vertex AI）
GOOGLE_CLOUD_PROJECT=your-project-id
```

## 🔧 故障排除

### 常见问题

#### 1. Gemini CLI未安装
```
错误: Gemini CLI未安装，请先运行: npm install -g @google/gemini-cli
```

**解决方案**: 运行安装脚本或手动安装
```bash
./scripts/install_gemini_cli.sh
```

#### 2. Node.js版本过低
```
错误: Node.js 版本过低 (当前: v16.x.x, 需要: v18.0.0+)
```

**解决方案**: 更新Node.js到18或更高版本

#### 3. API密钥未配置
```
错误: API密钥未配置或无效
```

**解决方案**: 检查 `.env` 文件中的 `GEMINI_API_KEY` 是否正确配置

#### 4. 首次认证
如果出现认证问题，运行以下命令进行首次登录：
```bash
gemini
```

#### 5. 超时问题
如果请求经常超时，可以增加超时时间：
```bash
GEMINI_TIMEOUT_SECONDS=180
```

### 日志查看

查看详细日志：
```bash
tail -f webhook.log
```

### 测试连接

可以手动测试Gemini CLI：
```bash
gemini "Hello, can you help me with coding?"
```

## 🔗 有用的链接

- [Google AI Studio](https://aistudio.google.com/) - 获取API密钥
- [Gemini CLI文档](https://www.geminicli.io/) - 官方文档
- [Node.js下载](https://nodejs.org/) - Node.js官网
- [项目GitHub](https://github.com/your-repo) - 项目仓库

## 📋 迁移检查清单

- [ ] 安装Node.js (版本 >= 18)
- [ ] 安装Gemini CLI
- [ ] 获取Gemini API密钥
- [ ] 配置环境变量
- [ ] 测试Gemini CLI连接
- [ ] 启动服务并验证功能
- [ ] 检查webhook响应正常

## 🆘 获取帮助

如果遇到问题：
1. 检查日志文件 `webhook.log`
2. 验证环境变量配置
3. 测试Gemini CLI是否正常工作
4. 查看本文档的故障排除部分

---

*此迁移保持了与Claude API相同的接口，确保无缝切换体验。*

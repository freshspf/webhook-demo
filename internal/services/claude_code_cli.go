package services

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/webhook-demo/internal/config"
)

// ClaudeCodeCLIService Claude Code CLI服务
type ClaudeCodeCLIService struct {
	config *config.ClaudeCodeCLIConfig
}

// NewClaudeCodeCLIService 创建新的Claude Code CLI服务
func NewClaudeCodeCLIService(cfg *config.ClaudeCodeCLIConfig) *ClaudeCodeCLIService {
	return &ClaudeCodeCLIService{
		config: cfg,
	}
}

// GenerateCode 生成代码
func (ccs *ClaudeCodeCLIService) GenerateCode(requirement string, context string) (string, error) {
	prompt := ccs.buildCodeGenerationPrompt(requirement, context)
	return ccs.callClaudeCodeCLI(prompt)
}

// ContinueCode 继续代码开发
func (ccs *ClaudeCodeCLIService) ContinueCode(instruction string, context string) (string, error) {
	prompt := ccs.buildContinuePrompt(instruction, context)
	return ccs.callClaudeCodeCLI(prompt)
}

// FixCode 修复代码问题
func (ccs *ClaudeCodeCLIService) FixCode(problem string, codeContext string) (string, error) {
	prompt := ccs.buildFixPrompt(problem, codeContext)
	return ccs.callClaudeCodeCLI(prompt)
}

// Summarize 总结内容
func (ccs *ClaudeCodeCLIService) Summarize(summaryPrompt string) (string, error) {
	return ccs.callClaudeCodeCLI(summaryPrompt)
}

// SummarizeInRepo 在指定仓库目录中总结内容
func (ccs *ClaudeCodeCLIService) SummarizeInRepo(summaryPrompt string, repoPath string) (string, error) {
	return ccs.callClaudeCodeCLIInDir(summaryPrompt, repoPath)
}

// ReviewCode 代码审查
func (ccs *ClaudeCodeCLIService) ReviewCode(reviewPrompt string, context string) (string, error) {
	prompt := ccs.buildReviewPrompt(reviewPrompt, context)
	return ccs.callClaudeCodeCLI(prompt)
}

// ReviewCodeInRepo 在指定仓库目录中进行代码审查
func (ccs *ClaudeCodeCLIService) ReviewCodeInRepo(reviewPrompt string, repoPath string) (string, error) {
	// 直接使用传入的reviewPrompt，不再调用buildReviewPrompt避免重复
	return ccs.callClaudeCodeCLIInDir(reviewPrompt, repoPath)
}

// callClaudeCodeCLI 调用Claude Code CLI
func (ccs *ClaudeCodeCLIService) callClaudeCodeCLI(prompt string) (string, error) {
	return ccs.callClaudeCodeCLIInDir(prompt, "")
}

// callClaudeCodeCLIInDir 在指定目录中调用Claude Code CLI
func (ccs *ClaudeCodeCLIService) callClaudeCodeCLIInDir(prompt string, workDir string) (string, error) {
	// 检查Claude Code CLI是否已安装
	if !ccs.isClaudeCodeCLIInstalled() {
		return "", fmt.Errorf("claude Code CLI未安装，请先运行: npm install -g @anthropic-ai/claude-code")
	}

	// 构建命令参数
	args := []string{}

	// 使用非交互模式
	args = append(args, "--print")

	// 添加模型参数（如果指定）
	if ccs.config.Model != "" {
		args = append(args, "--model", ccs.config.Model)
	}

	log.Printf("调用Claude Code CLI，模型: %s, BaseURL: %s",
		ccs.config.Model, ccs.config.BaseURL)
	log.Printf("提示词长度: %d 字符", len(prompt))

	// 设置环境变量
	env := os.Environ()
	if ccs.config.APIKey != "" {
		env = append(env, "ANTHROPIC_API_KEY="+ccs.config.APIKey)
	}
	if ccs.config.BaseURL != "" {
		env = append(env, "ANTHROPIC_BASE_URL="+ccs.config.BaseURL)
	}
	// 禁用非必要流量以提高性能
	env = append(env, "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1")

	log.Printf("设置环境变量: ANTHROPIC_API_KEY=%s, ANTHROPIC_BASE_URL=%s, CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1",
		ccs.maskAPIKey(ccs.config.APIKey), ccs.config.BaseURL)
	log.Printf("执行命令: claude %v", args)

	// 执行命令，使用stdin传递提示词
	cmd := exec.Command("claude", args...)
	cmd.Env = env

	// 设置工作目录
	if workDir != "" {
		cmd.Dir = workDir
		log.Printf("设置Claude CLI工作目录: %s", workDir)
	}

	// 通过stdin传递提示词，避免命令行参数长度限制
	cmd.Stdin = strings.NewReader(prompt)

	// 在这里加个sleep，等待3秒钟
	time.Sleep(3 * time.Second)

	// 执行命令并获取输出
	output, err := cmd.Output()
	if err != nil {
		// 如果是ExitError，可以获取stderr
		if exitError, ok := err.(*exec.ExitError); ok {
			stderr := string(exitError.Stderr)
			log.Printf("Claude CLI stderr: %s", stderr)
			log.Printf("Claude CLI exit code: %v", err)
			return "", fmt.Errorf("claude Code CLI执行失败: %v, 错误输出: %s", err, stderr)
		}
		return "", fmt.Errorf("claude Code CLI执行失败: %v", err)
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return "", fmt.Errorf("claude Code CLI没有返回任何输出")
	}

	log.Printf("Claude Code CLI调用成功，输出长度: %d 字符", len(outputStr))
	return outputStr, nil
}

// isClaudeCodeCLIInstalled 检查Claude Code CLI是否已安装
func (ccs *ClaudeCodeCLIService) isClaudeCodeCLIInstalled() bool {
	cmd := exec.Command("claude", "--version")
	err := cmd.Run()
	return err == nil
}

// buildCodeGenerationPrompt 构建代码生成提示
func (ccs *ClaudeCodeCLIService) buildCodeGenerationPrompt(requirement string, context string) string {
	return fmt.Sprintf("你是一个专业的软件开发助手，专门帮助用户生成高质量的代码。\n\n"+
		"**需求描述:**\n"+
		"%s\n\n"+
		"**项目上下文:**\n"+
		"%s\n\n"+
		"**要求:**\n"+
		"1. 生成完整、可运行的代码\n"+
		"2. 包含必要的注释和文档\n"+
		"3. 遵循最佳实践和代码规范\n"+
		"4. 考虑错误处理和边界情况\n"+
		"5. 如果涉及多个文件，请明确标注文件名\n\n"+
		"**输出格式:**\n"+
		"请直接输出代码，不需要额外的解释。如果需要多个文件，请使用 ```filename:path/to/file``` 格式标注。\n\n"+
		"请开始生成代码:", requirement, context)
}

// buildContinuePrompt 构建继续开发提示
func (ccs *ClaudeCodeCLIService) buildContinuePrompt(instruction string, context string) string {
	return fmt.Sprintf("你正在继续一个软件开发项目。请根据以下指令继续开发：\n\n"+
		"**继续指令:**\n"+
		"%s\n\n"+
		"**当前项目上下文:**\n"+
		"%s\n\n"+
		"**要求:**\n"+
		"1. 基于现有代码继续开发\n"+
		"2. 保持代码风格的一致性\n"+
		"3. 确保新代码与现有代码兼容\n"+
		"4. 添加必要的注释说明\n\n"+
		"请继续开发:", instruction, context)
}

// buildFixPrompt 构建代码修复提示
func (ccs *ClaudeCodeCLIService) buildFixPrompt(problem string, codeContext string) string {
	return fmt.Sprintf("你正在修复代码中的问题。请分析并修复以下问题：\n\n"+
		"**问题描述:**\n"+
		"%s\n\n"+
		"**代码上下文:**\n"+
		"%s\n\n"+
		"**要求:**\n"+
		"1. 分析问题的根本原因\n"+
		"2. 提供修复方案\n"+
		"3. 确保修复后的代码正确运行\n"+
		"4. 添加必要的注释说明修复内容\n\n"+
		"请修复代码:", problem, codeContext)
}

// buildReviewPrompt 构建代码审查提示
func (ccs *ClaudeCodeCLIService) buildReviewPrompt(reviewPrompt string, context string) string {
	return fmt.Sprintf("你是一个资深的代码审查专家，请对以下代码进行专业的审查：\n\n"+
		"**审查需求:**\n"+
		"%s\n\n"+
		"**项目上下文:**\n"+
		"%s\n\n"+
		"**审查标准:**\n"+
		"1. **代码质量:** 可读性、可维护性、代码结构\n"+
		"2. **安全性:** 安全漏洞、输入验证、权限控制\n"+
		"3. **性能:** 算法效率、资源使用、优化机会\n"+
		"4. **最佳实践:** 设计模式、编码规范、架构原则\n"+
		"5. **错误处理:** 异常处理、边界条件、容错机制\n"+
		"6. **测试:** 测试覆盖度、测试质量\n"+
		"7. **文档:** 代码注释、API文档\n\n"+
		"**输出要求:**\n"+
		"- 使用Markdown格式\n"+
		"- 提供具体的代码位置和建议\n"+
		"- 按严重程度分类问题（严重/中等/轻微）\n"+
		"- 给出具体的改进建议和示例代码\n"+
		"- 提供总体评分和改进建议\n\n"+
		"请开始代码审查:", reviewPrompt, context)
}

// maskAPIKey 遮盖API密钥用于日志显示
func (ccs *ClaudeCodeCLIService) maskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return "***"
	}
	return apiKey[:4] + "***" + apiKey[len(apiKey)-4:]
}

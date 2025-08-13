package services

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/webhook-demo/internal/config"
)

// GeminiService Gemini CLI服务
type GeminiService struct {
	config *config.GeminiConfig
}

// NewGeminiService 创建新的Gemini服务
func NewGeminiService(cfg *config.GeminiConfig) *GeminiService {
	return &GeminiService{
		config: cfg,
	}
}

// GenerateCode 生成代码
func (gs *GeminiService) GenerateCode(requirement string, context string) (string, error) {
	prompt := gs.buildCodeGenerationPrompt(requirement, context)
	return gs.callGeminiCLI(prompt)
}

// ContinueCode 继续代码开发
func (gs *GeminiService) ContinueCode(instruction string, context string) (string, error) {
	prompt := gs.buildContinuePrompt(instruction, context)
	return gs.callGeminiCLI(prompt)
}

// FixCode 修复代码问题
func (gs *GeminiService) FixCode(problem string, codeContext string) (string, error) {
	prompt := gs.buildFixPrompt(problem, codeContext)
	return gs.callGeminiCLI(prompt)
}

// callGeminiCLI 调用Gemini CLI
func (gs *GeminiService) callGeminiCLI(prompt string) (string, error) {
	// 检查Gemini CLI是否已安装
	if !gs.isGeminiCLIInstalled() {
		return "", fmt.Errorf("Gemini CLI未安装，请先运行: npm install -g @google/gemini-cli")
	}

	// 设置环境变量
	env := os.Environ()
	if gs.config.APIKey != "" {
		env = append(env, "GEMINI_API_KEY="+gs.config.APIKey)
	}
	if gs.config.Model != "" {
		env = append(env, "GEMINI_MODEL="+gs.config.Model)
	}
	if gs.config.ProjectID != "" {
		env = append(env, "GOOGLE_CLOUD_PROJECT="+gs.config.ProjectID)
	}

	// 构建命令
	args := []string{}
	
	// 添加模型参数（如果指定）
	if gs.config.Model != "" {
		args = append(args, "--model", gs.config.Model)
	}
	
	// 添加最大token数限制（如果支持）
	if gs.config.MaxTokens > 0 {
		args = append(args, "--max-tokens", fmt.Sprintf("%d", gs.config.MaxTokens))
	}

	// 添加提示词
	args = append(args, prompt)

	log.Printf("调用Gemini CLI，模型: %s, 最大Token: %d", gs.config.Model, gs.config.MaxTokens)

	// 执行命令
	cmd := exec.Command("gemini", args...)
	cmd.Env = env

	// 使用管道来处理输入输出
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("创建stdin管道失败: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("创建stdout管道失败: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("创建stderr管道失败: %v", err)
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("启动Gemini CLI失败: %v", err)
	}

	// 关闭stdin（因为我们不需要交互式输入）
	stdin.Close()

	// 读取输出
	var result strings.Builder
	var errorOutput strings.Builder

	// 读取stdout
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			result.WriteString(line + "\n")
		}
	}()

	// 读取stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			errorOutput.WriteString(line + "\n")
		}
	}()

	// 等待命令完成
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	// 设置超时
	timeout := time.Duration(gs.config.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 120 * time.Second // 默认2分钟超时
	}

	select {
	case err := <-done:
		if err != nil {
			errMsg := errorOutput.String()
			if errMsg != "" {
				return "", fmt.Errorf("Gemini CLI执行失败: %v, 错误输出: %s", err, errMsg)
			}
			return "", fmt.Errorf("Gemini CLI执行失败: %v", err)
		}
	case <-time.After(timeout):
		cmd.Process.Kill()
		return "", fmt.Errorf("Gemini CLI调用超时 (%v)", timeout)
	}

	output := strings.TrimSpace(result.String())
	if output == "" {
		return "", fmt.Errorf("Gemini CLI没有返回任何输出")
	}

	log.Printf("Gemini CLI调用成功，输出长度: %d 字符", len(output))
	return output, nil
}

// isGeminiCLIInstalled 检查Gemini CLI是否已安装
func (gs *GeminiService) isGeminiCLIInstalled() bool {
	cmd := exec.Command("gemini", "--version")
	err := cmd.Run()
	return err == nil
}

// buildCodeGenerationPrompt 构建代码生成提示
func (gs *GeminiService) buildCodeGenerationPrompt(requirement string, context string) string {
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
func (gs *GeminiService) buildContinuePrompt(instruction string, context string) string {
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
func (gs *GeminiService) buildFixPrompt(problem string, codeContext string) string {
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

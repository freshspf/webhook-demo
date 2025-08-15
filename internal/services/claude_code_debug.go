package services

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// DebugClaudeCLI 调试Claude CLI连接
func (ccs *ClaudeCodeCLIService) DebugClaudeCLI() error {
	fmt.Println("🔧 Claude Code CLI 诊断工具")
	fmt.Println("=====================================")

	// 1. 检查Claude CLI是否安装
	fmt.Print("1. 检查Claude CLI安装状态... ")
	if !ccs.isClaudeCodeCLIInstalled() {
		fmt.Println("❌ 未安装")
		fmt.Println("   请运行: npm install -g @anthropic-ai/claude-code")
		return fmt.Errorf("claude CLI未安装")
	}
	fmt.Println("✅ 已安装")

	// 2. 检查版本
	fmt.Print("2. 检查Claude CLI版本... ")
	version, err := ccs.getClaudeCLIVersion()
	if err != nil {
		fmt.Printf("❌ 获取版本失败: %v\n", err)
	} else {
		fmt.Printf("✅ %s\n", version)
	}

	// 3. 检查配置
	fmt.Println("3. 检查配置:")
	fmt.Printf("   - API Key: %s\n", ccs.maskAPIKey(ccs.config.APIKey))
	fmt.Printf("   - Model: %s\n", ccs.config.Model)
	fmt.Printf("   - BaseURL: %s\n", ccs.config.BaseURL)
	fmt.Printf("   - Timeout: %d seconds\n", ccs.config.TimeoutSeconds)

	// 4. 测试简单命令
	fmt.Print("4. 测试简单连接... ")
	testPrompt := "请简单回复'连接测试成功'"

	start := time.Now()
	result, err := ccs.testConnection(testPrompt)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ 连接失败 (耗时: %v)\n", duration)
		fmt.Printf("   错误: %v\n", err)
		return err
	}

	fmt.Printf("✅ 连接成功 (耗时: %v)\n", duration)
	fmt.Printf("   响应: %s\n", strings.TrimSpace(result))

	fmt.Println("\n✅ Claude CLI诊断完成!")
	return nil
}

// getClaudeCLIVersion 获取Claude CLI版本
func (ccs *ClaudeCodeCLIService) getClaudeCLIVersion() (string, error) {
	cmd := exec.Command("claude", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// testConnection 测试连接
func (ccs *ClaudeCodeCLIService) testConnection(prompt string) (string, error) {
	// 设置环境变量
	env := os.Environ()
	if ccs.config.APIKey != "" {
		env = append(env, "ANTHROPIC_API_KEY="+ccs.config.APIKey)
	}
	if ccs.config.BaseURL != "" {
		env = append(env, "ANTHROPIC_BASE_URL="+ccs.config.BaseURL)
	}
	env = append(env, "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1")

	// 创建命令
	args := []string{"--print"}
	if ccs.config.Model != "" {
		args = append(args, "--model", ccs.config.Model)
	}

	cmd := exec.Command("claude", args...)
	cmd.Env = env
	cmd.Stdin = strings.NewReader(prompt)

	// 执行命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("命令执行失败: %v, 输出: %s", err, string(output))
	}

	return string(output), nil
}

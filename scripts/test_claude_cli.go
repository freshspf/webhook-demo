package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/webhook-demo/internal/config"
	"github.com/webhook-demo/internal/services"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Printf("警告: 无法加载 .env 文件: %v", err)
	}

	// 加载配置
	cfg := config.Load()

	// 创建Claude Code CLI服务
	claudeService := services.NewClaudeCodeCLIService(&cfg.ClaudeCodeCLI)

	// 运行诊断
	if err := claudeService.DebugClaudeCLI(); err != nil {
		log.Printf("诊断失败: %v", err)
		os.Exit(1)
	}
}

package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config 应用配置结构
type Config struct {
	Server ServerConfig
	GitHub GitHubConfig
	Claude ClaudeConfig
	Git    GitConfig
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string
	Mode string // debug/release
}

// GitHubConfig GitHub相关配置
type GitHubConfig struct {
	Token         string
	WebhookSecret string
}

// ClaudeConfig Claude API相关配置
type ClaudeConfig struct {
	APIKey    string
	Model     string
	MaxTokens int
}

// Load 加载配置
func Load() *Config {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Printf("警告: 无法加载 .env 文件: %v", err)
	}

	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		GitHub: GitHubConfig{
			Token:         getEnv("GITHUB_TOKEN", ""),
			WebhookSecret: getEnv("GITHUB_WEBHOOK_SECRET", "your-webhook-secret"),
		},
		Claude: ClaudeConfig{
			APIKey:    getEnv("CLAUDE_API_KEY", ""),
			Model:     getEnv("CLAUDE_MODEL", "claude-3-5-sonnet-20241022"),
			MaxTokens: getEnvAsInt("CLAUDE_MAX_TOKENS", 4000),
		},
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 获取环境变量并转换为整数
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("警告: 环境变量 %s 不是有效的整数，使用默认值 %d", key, defaultValue)
	}
	return defaultValue
}

// getEnvAsBool 获取环境变量并转换为布尔值
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
		log.Printf("警告: 环境变量 %s 不是有效的布尔值，使用默认值 %t", key, defaultValue)
	}
	return defaultValue
}

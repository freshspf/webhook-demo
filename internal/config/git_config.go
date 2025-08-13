package config

// GitConfig Git相关配置
type GitConfig struct {
	WorkDir     string // 工作目录
	UserName    string // Git用户名
	UserEmail   string // Git邮箱
	MaxFileSize int64  // 最大文件大小（字节）
}

// LoadGitConfig 加载Git配置
func LoadGitConfig() *GitConfig {
	return &GitConfig{
		WorkDir:     getEnv("GIT_WORK_DIR", "/tmp/webhook-demo"),
		UserName:    getEnv("GIT_USER_NAME", "CodeAgent"),
		UserEmail:   getEnv("GIT_USER_EMAIL", "codeagent@example.com"),
		MaxFileSize: int64(getEnvAsInt("GIT_MAX_FILE_SIZE", 1024*1024)), // 默认1MB
	}
}

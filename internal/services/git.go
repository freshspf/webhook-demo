package services

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// GitService Git操作服务
type GitService struct {
	workDir string // 工作目录
}

// NewGitService 创建新的Git服务
func NewGitService(workDir string) *GitService {
	if workDir == "" {
		workDir = "/tmp/webhook-demo"
	}

	// 确保工作目录存在
	if err := os.MkdirAll(workDir, 0755); err != nil {
		log.Printf("创建工作目录失败: %v", err)
	}

	return &GitService{
		workDir: workDir,
	}
}

// CloneRepository 克隆仓库
func (gs *GitService) CloneRepository(repoURL, branch string) (string, error) {
	// 生成唯一的目录名
	timestamp := time.Now().Format("20060102_150405")
	repoName := fmt.Sprintf("repo_%s", timestamp)
	repoPath := filepath.Join(gs.workDir, repoName)

	log.Printf("克隆仓库: %s 到 %s", repoURL, repoPath)

	// 执行git clone命令
	cmd := exec.Command("git", "clone", "-b", branch, repoURL, repoPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("克隆仓库失败: %v", err)
	}

	log.Printf("仓库克隆成功: %s", repoPath)
	return repoPath, nil
}

// ReadFile 读取文件内容
func (gs *GitService) ReadFile(repoPath, filePath string) (string, error) {
	fullPath := filepath.Join(repoPath, filePath)

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("读取文件失败 %s: %v", fullPath, err)
	}

	return string(content), nil
}

// WriteFile 写入文件内容
func (gs *GitService) WriteFile(repoPath, filePath, content string) error {
	fullPath := filepath.Join(repoPath, filePath)

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	log.Printf("文件写入成功: %s", fullPath)
	return nil
}

// ListFiles 列出目录下的文件
func (gs *GitService) ListFiles(repoPath, dirPath string) ([]string, error) {
	fullPath := filepath.Join(repoPath, dirPath)

	var files []string
	err := filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过.git目录
		if strings.Contains(path, ".git") {
			return nil
		}

		// 只返回文件，不返回目录
		if !info.IsDir() {
			// 返回相对路径
			relPath, err := filepath.Rel(repoPath, path)
			if err != nil {
				return err
			}
			files = append(files, relPath)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("遍历目录失败: %v", err)
	}

	return files, nil
}

// GetFileTree 获取文件树结构
func (gs *GitService) GetFileTree(repoPath string) (string, error) {
	var tree strings.Builder
	tree.WriteString("📁 项目文件结构:\n")

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过.git目录
		if strings.Contains(path, ".git") {
			return filepath.SkipDir
		}

		// 计算相对路径
		relPath, err := filepath.Rel(repoPath, path)
		if err != nil {
			return err
		}

		// 跳过根目录
		if relPath == "." {
			return nil
		}

		// 计算缩进
		depth := strings.Count(relPath, string(os.PathSeparator))
		indent := strings.Repeat("  ", depth)

		if info.IsDir() {
			tree.WriteString(fmt.Sprintf("%s📁 %s/\n", indent, filepath.Base(path)))
		} else {
			tree.WriteString(fmt.Sprintf("%s📄 %s\n", indent, filepath.Base(path)))
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("生成文件树失败: %v", err)
	}

	return tree.String(), nil
}

// CreateBranch 创建新分支
func (gs *GitService) CreateBranch(repoPath, branchName string) error {
	log.Printf("创建分支: %s", branchName)

	// 切换到仓库目录
	cmd := exec.Command("git", "-C", repoPath, "checkout", "-b", branchName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("创建分支失败: %v", err)
	}

	log.Printf("分支创建成功: %s", branchName)
	return nil
}

// AddFiles 添加文件到暂存区
func (gs *GitService) AddFiles(repoPath string, files []string) error {
	log.Printf("添加文件到暂存区: %v", files)

	args := append([]string{"-C", repoPath, "add"}, files...)
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("添加文件失败: %v", err)
	}

	return nil
}

// Commit 提交更改
func (gs *GitService) Commit(repoPath, message string) error {
	log.Printf("提交更改: %s", message)

	cmd := exec.Command("git", "-C", repoPath, "commit", "-m", message)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("提交失败: %v", err)
	}

	log.Printf("提交成功: %s", message)
	return nil
}

// Push 推送到远程仓库
func (gs *GitService) Push(repoPath, branchName string) error {
	log.Printf("推送分支: %s", branchName)

	cmd := exec.Command("git", "-C", repoPath, "push", "origin", branchName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("推送失败: %v", err)
	}

	log.Printf("推送成功: %s", branchName)
	return nil
}

// GetDiff 获取文件差异
func (gs *GitService) GetDiff(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "diff", "--cached")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("获取差异失败: %v", err)
	}

	return string(output), nil
}

// GetStatus 获取Git状态
func (gs *GitService) GetStatus(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("获取状态失败: %v", err)
	}

	return string(output), nil
}

// ConfigureGit 配置Git用户信息
func (gs *GitService) ConfigureGit(repoPath, name, email string) error {
	log.Printf("配置Git用户: %s <%s>", name, email)

	// 配置用户名
	cmd := exec.Command("git", "-C", repoPath, "config", "user.name", name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("配置用户名失败: %v", err)
	}

	// 配置邮箱
	cmd = exec.Command("git", "-C", repoPath, "config", "user.email", email)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("配置邮箱失败: %v", err)
	}

	log.Printf("Git配置成功")
	return nil
}

// Cleanup 清理工作目录
func (gs *GitService) Cleanup(repoPath string) error {
	log.Printf("清理工作目录: %s", repoPath)

	if err := os.RemoveAll(repoPath); err != nil {
		return fmt.Errorf("清理目录失败: %v", err)
	}

	log.Printf("工作目录清理成功")
	return nil
}

// GetFileContent 获取文件内容（支持大文件）
func (gs *GitService) GetFileContent(repoPath, filePath string, maxSize int64) (string, error) {
	fullPath := filepath.Join(repoPath, filePath)

	file, err := os.Open(fullPath)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 获取文件信息
	info, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("获取文件信息失败: %v", err)
	}

	// 检查文件大小
	if info.Size() > maxSize {
		return "", fmt.Errorf("文件过大: %d bytes (最大允许: %d bytes)", info.Size(), maxSize)
	}

	// 读取文件内容
	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %v", err)
	}

	return string(content), nil
}

// FindFilesByPattern 根据模式查找文件
func (gs *GitService) FindFilesByPattern(repoPath, pattern string) ([]string, error) {
	var files []string

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过.git目录
		if strings.Contains(path, ".git") {
			return filepath.SkipDir
		}

		// 只处理文件
		if !info.IsDir() {
			// 检查文件名是否匹配模式
			if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
				relPath, err := filepath.Rel(repoPath, path)
				if err != nil {
					return err
				}
				files = append(files, relPath)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("查找文件失败: %v", err)
	}

	return files, nil
}

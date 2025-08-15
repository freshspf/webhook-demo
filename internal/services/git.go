package services

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// GitService Git操作服务
type GitService struct {
	workDir      string                 // 工作目录
	repoCache    map[string]*CachedRepo // 仓库缓存
	cacheMutex   sync.RWMutex           // 缓存读写锁
	lastCloneMap map[string]time.Time   // 记录每个仓库的最后克隆时间
	cloneMutex   sync.RWMutex           // 克隆时间锁
}

// CachedRepo 缓存的仓库信息
type CachedRepo struct {
	Path       string    // 仓库路径
	URL        string    // 仓库URL
	Branch     string    // 分支
	LastUpdate time.Time // 最后更新时间
	Valid      bool      // 是否有效
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
		workDir:      workDir,
		repoCache:    make(map[string]*CachedRepo),
		lastCloneMap: make(map[string]time.Time),
	}
}

// CloneRepository 克隆仓库（带缓存和重试机制）
func (gs *GitService) CloneRepository(repoURL, branch string) (string, error) {
	// 检查频率限制
	if err := gs.checkRateLimit(repoURL); err != nil {
		return "", err
	}

	// 检查缓存
	cacheKey := gs.generateCacheKey(repoURL, branch)
	if cachedRepo := gs.getCachedRepo(cacheKey); cachedRepo != nil {
		log.Printf("使用缓存仓库: %s", cachedRepo.Path)
		return cachedRepo.Path, nil
	}

	// 生成唯一的目录名
	timestamp := time.Now().Format("20060102_150405")
	repoName := fmt.Sprintf("repo_%s", timestamp)
	repoPath := filepath.Join(gs.workDir, repoName)

	log.Printf("克隆仓库: %s 到 %s", repoURL, repoPath)

	// 记录克隆时间
	gs.recordCloneTime(repoURL)

	// 尝试克隆
	repoPath, err := gs.attemptClone(repoURL, branch, repoPath)
	if err != nil {
		return "", err
	}

	// 克隆成功，缓存结果
	gs.cacheRepo(cacheKey, &CachedRepo{
		Path:       repoPath,
		URL:        repoURL,
		Branch:     branch,
		LastUpdate: time.Now(),
		Valid:      true,
	})

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

// GetPullRequestDiff 获取Pull Request的代码差异
func (gs *GitService) GetPullRequestDiff(repoPath, headSHA, baseSHA string) (string, error) {
	log.Printf("获取PR diff: %s -> %s", baseSHA, headSHA)

	// 先fetch head分支的内容
	fetchCmd := exec.Command("git", "fetch", "origin", headSHA)
	fetchCmd.Dir = repoPath
	if err := fetchCmd.Run(); err != nil {
		log.Printf("fetch head SHA失败: %v", err)
		// 尝试使用分支名而不是SHA
		return gs.getPullRequestDiffByBranch(repoPath)
	}

	// 获取diff
	diffCmd := exec.Command("git", "diff", baseSHA+"..."+headSHA)
	diffCmd.Dir = repoPath

	output, err := diffCmd.Output()
	if err != nil {
		log.Printf("获取git diff失败: %v", err)
		return "无法获取PR代码差异", nil
	}

	diff := string(output)
	if diff == "" {
		return "PR无代码变更", nil
	}

	// 限制diff长度，避免过长
	if len(diff) > 10000 {
		diff = diff[:10000] + "\n\n... (diff内容过长，已截断)"
	}

	return diff, nil
}

// getPullRequestDiffByBranch 通过分支获取diff（备用方法）
func (gs *GitService) getPullRequestDiffByBranch(repoPath string) (string, error) {
	// 简单获取最近几个commit的diff
	diffCmd := exec.Command("git", "log", "--oneline", "-5", "--stat")
	diffCmd.Dir = repoPath

	output, err := diffCmd.Output()
	if err != nil {
		return "无法获取PR差异信息", err
	}

	return string(output), nil
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

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// 简单的推送命令
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath,
		"-c", "http.sslVerify=false",
		"-c", "http.postBuffer=1048576000",
		"push", "-u", "origin", branchName)

	// 继承环境变量
	cmd.Env = append(os.Environ(),
		"GIT_HTTP_TIMEOUT=90",
		"GIT_HTTP_MAX_RETRIES=3")

	// 执行推送
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("推送失败，错误输出: %s", string(output))
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

// GetModifiedFiles 获取修改的文件列表
func (gs *GitService) GetModifiedFiles(repoPath string) ([]string, error) {
	cmd := exec.Command("git", "-C", repoPath, "diff", "--cached", "--name-only")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("获取修改文件列表失败: %v", err)
	}

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(files) == 1 && files[0] == "" {
		return []string{}, nil
	}

	return files, nil
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

// HasChanges 检查是否有未提交的修改
func (gs *GitService) HasChanges(repoPath string) (bool, error) {
	// 检查是否有暂存的文件
	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("检查git暂存状态失败: %v", err)
	}

	// 如果输出不为空，说明有暂存的文件
	return strings.TrimSpace(string(output)) != "", nil
}

// DeleteFile 删除文件
func (gs *GitService) DeleteFile(repoPath, filePath string) error {
	fullPath := filepath.Join(repoPath, filePath)

	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil // 文件不存在，认为删除成功
	}

	// 删除文件
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("删除文件失败: %v", err)
	}

	return nil
}

// testNetworkConnection 测试网络连接
// func (gs *GitService) testNetworkConnection() error {
// 	// 测试DNS解析
// 	log.Printf("测试DNS解析...")
// 	pingCmd := exec.Command("ping", "-c", "1", "github.com")
// 	pingCmd.Stdout = os.Stdout
// 	pingCmd.Stderr = os.Stderr
// 	if err := pingCmd.Run(); err != nil {
// 		return fmt.Errorf("ping github.com 失败: %v", err)
// 	}

// 	// 测试网络连接 - 尝试HTTP和HTTPS
// 	log.Printf("测试HTTP连接...")
// 	httpCmd := exec.Command("curl", "-I", "--connect-timeout", "10", "http://github.com")
// 	httpCmd.Stdout = os.Stdout
// 	httpCmd.Stderr = os.Stderr
// 	if err := httpCmd.Run(); err != nil {
// 		log.Printf("HTTP连接测试失败: %v", err)
// 	} else {
// 		log.Printf("HTTP连接测试成功")
// 	}

// 	log.Printf("测试HTTPS连接(跳过SSL验证)...")
// 	curlCmd := exec.Command("curl", "-I", "-k", "--http1.1", "--connect-timeout", "10", "https://github.com")
// 	curlCmd.Stdout = os.Stdout
// 	curlCmd.Stderr = os.Stderr
// 	if err := curlCmd.Run(); err != nil {
// 		log.Printf("HTTPS连接测试失败，但继续尝试Git克隆: %v", err)
// 		// 即使网络测试失败，也继续尝试Git克隆
// 	}

// 	return nil
// }

// generateCacheKey 生成缓存键
func (gs *GitService) generateCacheKey(repoURL, branch string) string {
	data := fmt.Sprintf("%s:%s", repoURL, branch)
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// getCachedRepo 获取缓存的仓库
func (gs *GitService) getCachedRepo(cacheKey string) *CachedRepo {
	gs.cacheMutex.RLock()
	defer gs.cacheMutex.RUnlock()

	cachedRepo, exists := gs.repoCache[cacheKey]
	if !exists || !cachedRepo.Valid {
		return nil
	}

	// 检查缓存是否过期（30分钟）
	if time.Since(cachedRepo.LastUpdate) > 30*time.Minute {
		log.Printf("缓存过期，删除缓存: %s", cachedRepo.Path)
		return nil
	}

	// 检查目录是否还存在
	if _, err := os.Stat(cachedRepo.Path); os.IsNotExist(err) {
		log.Printf("缓存目录不存在，删除缓存: %s", cachedRepo.Path)
		return nil
	}

	return cachedRepo
}

// cacheRepo 缓存仓库
func (gs *GitService) cacheRepo(cacheKey string, repo *CachedRepo) {
	gs.cacheMutex.Lock()
	defer gs.cacheMutex.Unlock()
	gs.repoCache[cacheKey] = repo
}

// checkRateLimit 检查频率限制
func (gs *GitService) checkRateLimit(repoURL string) error {
	gs.cloneMutex.RLock()
	lastClone, exists := gs.lastCloneMap[repoURL]
	gs.cloneMutex.RUnlock()

	if exists {
		// GitHub限制：同一仓库5分钟内不能频繁克隆
		timeSinceLastClone := time.Since(lastClone)
		minInterval := 5 * time.Minute

		if timeSinceLastClone < minInterval {
			waitTime := minInterval - timeSinceLastClone
			log.Printf("频率限制：需要等待 %v 后才能重新克隆 %s", waitTime, repoURL)
			return fmt.Errorf("频率限制：请等待 %v 后重试", waitTime)
		}
	}

	return nil
}

// recordCloneTime 记录克隆时间
func (gs *GitService) recordCloneTime(repoURL string) {
	gs.cloneMutex.Lock()
	defer gs.cloneMutex.Unlock()
	gs.lastCloneMap[repoURL] = time.Now()
}

// attemptClone 尝试克隆仓库
func (gs *GitService) attemptClone(repoURL, branch, repoPath string) (string, error) {
	log.Printf("克隆仓库: %s, 分支: %s", repoURL, branch)

	// 清理目标目录
	os.RemoveAll(repoPath)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// 简单的浅克隆命令
	cmd := exec.CommandContext(ctx, "git", "clone",
		"-c", "http.sslVerify=false",
		"-c", "http.postBuffer=1048576000",
		"-b", branch,
		"--depth", "1",
		"--single-branch",
		repoURL, repoPath)

	// 继承环境变量
	cmd.Env = append(os.Environ(),
		"GIT_HTTP_TIMEOUT=60",
		"GIT_HTTP_MAX_RETRIES=3",
		"GIT_TERMINAL_PROGRESS=0")

	// 执行克隆
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("克隆失败，错误输出: %s", string(output))
		return "", fmt.Errorf("克隆失败: %v", err)
	}

	log.Printf("克隆成功: %s", repoPath)
	return repoPath, nil
}

// ClearCache 清理缓存
func (gs *GitService) ClearCache() {
	gs.cacheMutex.Lock()
	defer gs.cacheMutex.Unlock()

	for key, repo := range gs.repoCache {
		if repo.Path != "" {
			os.RemoveAll(repo.Path)
		}
		delete(gs.repoCache, key)
	}

	log.Printf("缓存已清理")
}

// GetCacheStatus 获取缓存状态
func (gs *GitService) GetCacheStatus() map[string]interface{} {
	gs.cacheMutex.RLock()
	defer gs.cacheMutex.RUnlock()

	status := map[string]interface{}{
		"cached_repos": len(gs.repoCache),
		"repos":        make([]map[string]interface{}, 0),
	}

	for key, repo := range gs.repoCache {
		repoInfo := map[string]interface{}{
			"key":         key,
			"url":         repo.URL,
			"branch":      repo.Branch,
			"path":        repo.Path,
			"last_update": repo.LastUpdate,
			"valid":       repo.Valid,
		}
		status["repos"] = append(status["repos"].([]map[string]interface{}), repoInfo)
	}

	return status
}

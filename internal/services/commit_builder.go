package services

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/webhook-demo/internal/models"
)

// CommitType 定义Conventional Commits的类型
type CommitType string

const (
	CommitTypeFeat     CommitType = "feat"     // 新增功能
	CommitTypeFix      CommitType = "fix"      // 修复问题
	CommitTypeRefactor CommitType = "refactor" // 重构代码
	CommitTypeDocs     CommitType = "docs"     // 文档变更
	CommitTypeStyle    CommitType = "style"    // 代码风格调整
	CommitTypeTest     CommitType = "test"     // 测试相关
	CommitTypeChore    CommitType = "chore"    // 构建配置、依赖管理等
	CommitTypePerf     CommitType = "perf"     // 性能优化
	CommitTypeBuild    CommitType = "build"    // 构建系统变更
	CommitTypeCI       CommitType = "ci"       // CI配置变更
	CommitTypeRevert   CommitType = "revert"   // 回滚变更
)

// CommitMessage 定义commit消息结构
type CommitMessage struct {
	Type        CommitType
	Scope       string
	Description string
	Body        string
	Footer      string
	IsBreaking  bool
}

// CommitBuilder commit消息构建器
type CommitBuilder struct{}

// NewCommitBuilder 创建新的commit构建器
func NewCommitBuilder() *CommitBuilder {
	return &CommitBuilder{}
}

// BuildAutoFixCommit 构建自动修复Issue的commit消息
func (cb *CommitBuilder) BuildAutoFixCommit(event *models.GitHubEvent, modifiedFiles []string) string {
	// 检测修改的文件类型来确定scope
	scope := cb.detectScope(modifiedFiles)

	// 根据Issue标题判断是feat还是fix
	commitType := cb.detectCommitType(event.Issue.Title, event.Issue.Body)

	// 构建描述
	description := cb.buildDescription(event.Issue.Title)

	// 构建正文
	body := fmt.Sprintf("由AI助手自动生成的代码修改\n\n修改文件:\n%s", strings.Join(modifiedFiles, "\n"))

	// 构建脚注
	footer := fmt.Sprintf("Closes #%d\nIssue链接: %s", event.Issue.Number, event.Issue.URL)

	commit := CommitMessage{
		Type:        commitType,
		Scope:       scope,
		Description: description,
		Body:        body,
		Footer:      footer,
	}

	return cb.format(commit)
}

// BuildPRCommit 构建PR相关的commit消息
func (cb *CommitBuilder) BuildPRCommit(prTitle, prDescription string, prNumber int) string {
	// 根据PR标题检测类型
	commitType := cb.detectCommitType(prTitle, prDescription)

	commit := CommitMessage{
		Type:        commitType,
		Description: cb.buildDescription(prTitle),
		Footer:      fmt.Sprintf("PR #%d", prNumber),
	}

	return cb.format(commit)
}

// detectScope 根据修改的文件检测scope
func (cb *CommitBuilder) detectScope(files []string) string {
	scopeMap := make(map[string]int)

	for _, file := range files {
		scope := cb.getFileScope(file)
		if scope != "" {
			scopeMap[scope]++
		}
	}

	// 返回出现次数最多的scope
	maxCount := 0
	selectedScope := ""
	for scope, count := range scopeMap {
		if count > maxCount {
			maxCount = count
			selectedScope = scope
		}
	}

	return selectedScope
}

// getFileScope 根据文件路径确定scope
func (cb *CommitBuilder) getFileScope(filePath string) string {
	dir := filepath.Dir(filePath)
	filename := filepath.Base(filePath)

	// 根据目录结构确定scope
	if strings.Contains(dir, "internal/handlers") {
		return "handlers"
	}
	if strings.Contains(dir, "internal/services") {
		return "services"
	}
	if strings.Contains(dir, "internal/models") {
		return "models"
	}
	if strings.Contains(dir, "internal/config") {
		return "config"
	}
	if strings.Contains(dir, "cmd/") {
		return "cmd"
	}
	if strings.Contains(dir, "pkg/") {
		return "pkg"
	}

	// 根据文件名确定scope
	if strings.Contains(filename, "test") {
		return "test"
	}
	if strings.HasSuffix(filename, ".md") {
		return "docs"
	}
	if strings.Contains(filename, "docker") || strings.Contains(filename, "Dockerfile") {
		return "docker"
	}
	if strings.Contains(filename, "go.mod") || strings.Contains(filename, "go.sum") {
		return "deps"
	}

	return ""
}

// detectCommitType 根据标题和内容检测commit类型
func (cb *CommitBuilder) detectCommitType(title, body string) CommitType {
	titleLower := strings.ToLower(title)
	bodyLower := strings.ToLower(body)
	content := titleLower + " " + bodyLower

	// 修复类关键词
	fixKeywords := []string{"修复", "解决", "fix", "solve", "bug", "错误", "问题", "异常", "故障"}
	for _, keyword := range fixKeywords {
		if strings.Contains(content, keyword) {
			return CommitTypeFix
		}
	}

	// 重构类关键词
	refactorKeywords := []string{"重构", "优化", "refactor", "optimize", "improve", "clean", "整理"}
	for _, keyword := range refactorKeywords {
		if strings.Contains(content, keyword) {
			return CommitTypeRefactor
		}
	}

	// 文档类关键词
	docsKeywords := []string{"文档", "doc", "readme", "注释", "说明"}
	for _, keyword := range docsKeywords {
		if strings.Contains(content, keyword) {
			return CommitTypeDocs
		}
	}

	// 测试类关键词
	testKeywords := []string{"测试", "test", "单元测试", "集成测试"}
	for _, keyword := range testKeywords {
		if strings.Contains(content, keyword) {
			return CommitTypeTest
		}
	}

	// 性能类关键词
	perfKeywords := []string{"性能", "performance", "perf", "速度", "优化性能"}
	for _, keyword := range perfKeywords {
		if strings.Contains(content, keyword) {
			return CommitTypePerf
		}
	}

	// CI/CD类关键词
	ciKeywords := []string{"ci", "cd", "pipeline", "workflow", "actions", "jenkins"}
	for _, keyword := range ciKeywords {
		if strings.Contains(content, keyword) {
			return CommitTypeCI
		}
	}

	// 构建类关键词
	buildKeywords := []string{"构建", "build", "webpack", "docker", "makefile"}
	for _, keyword := range buildKeywords {
		if strings.Contains(content, keyword) {
			return CommitTypeBuild
		}
	}

	// 默认为新功能
	return CommitTypeFeat
}

// buildDescription 构建commit描述
func (cb *CommitBuilder) buildDescription(title string) string {
	// 移除常见的前缀
	description := strings.TrimSpace(title)

	// 移除Issue编号前缀
	if strings.HasPrefix(description, "#") {
		parts := strings.SplitN(description, " ", 2)
		if len(parts) > 1 {
			description = parts[1]
		}
	}

	// 移除emoji
	description = strings.TrimSpace(strings.ReplaceAll(description, "🤖", ""))
	description = strings.TrimSpace(strings.ReplaceAll(description, "✨", ""))
	description = strings.TrimSpace(strings.ReplaceAll(description, "🐛", ""))

	// 限制长度
	if len(description) > 50 {
		description = description[:47] + "..."
	}

	return description
}

// format 格式化commit消息
func (cb *CommitBuilder) format(commit CommitMessage) string {
	var result strings.Builder

	// 构建第一行：type(scope): description
	result.WriteString(string(commit.Type))

	if commit.Scope != "" {
		result.WriteString("(")
		result.WriteString(commit.Scope)
		result.WriteString(")")
	}

	if commit.IsBreaking {
		result.WriteString("!")
	}

	result.WriteString(": ")
	result.WriteString(commit.Description)

	// 添加正文
	if commit.Body != "" {
		result.WriteString("\n\n")
		result.WriteString(commit.Body)
	}

	// 添加脚注
	if commit.Footer != "" {
		result.WriteString("\n\n")
		result.WriteString(commit.Footer)
	}

	return result.String()
}

// BuildManualCommit 构建手动指定类型的commit消息
func (cb *CommitBuilder) BuildManualCommit(commitType CommitType, scope, description, body, footer string) string {
	commit := CommitMessage{
		Type:        commitType,
		Scope:       scope,
		Description: description,
		Body:        body,
		Footer:      footer,
	}

	return cb.format(commit)
}

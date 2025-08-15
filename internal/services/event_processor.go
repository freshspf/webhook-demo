package services

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/webhook-demo/internal/models"
)

// EventProcessor 事件处理器
type EventProcessor struct {
	githubService     *GitHubService
	claudeCodeService *ClaudeCodeCLIService
	gitService        *GitService
	commandRegex      *regexp.Regexp
}

// NewEventProcessor 创建新的事件处理器
func NewEventProcessor(githubService *GitHubService, claudeCodeService *ClaudeCodeCLIService, gitService *GitService) *EventProcessor {
	return &EventProcessor{
		githubService:     githubService,
		claudeCodeService: claudeCodeService,
		gitService:        gitService,
		commandRegex:      regexp.MustCompile(`^/(code|continue|fix|help|review|summary)\s*(.*)$`),
	}
}

// ProcessEvent 处理GitHub事件
func (ep *EventProcessor) ProcessEvent(event *models.GitHubEvent) error {
	log.Printf("开始处理事件: Type=%s, DeliveryID=%s", event.Type, event.DeliveryID)

	// 设置时间戳
	event.Timestamp = time.Now()

	switch event.Type {
	case "issues": // 处理Issue事件
		return ep.handleIssuesEvent(event)
	case "issue_comment": // 处理Issue评论事件
		return ep.handleIssueCommentEvent(event)
	case "pull_request": // 处理Pull Request事件
		return ep.handlePullRequestEvent(event)
	case "pull_request_review_comment": // 处理PR Review评论事件
		return ep.handlePullRequestReviewCommentEvent(event)
	case "pull_request_review": // 处理PR Review事件
		return ep.handlePullRequestReviewEvent(event)
	case "ping": // 处理Ping事件
		return ep.handlePingEvent(event)
	default:
		log.Printf("忽略未支持的事件类型: %s", event.Type)
		return nil
	}
}

// handleIssuesEvent 处理Issue事件
func (ep *EventProcessor) handleIssuesEvent(event *models.GitHubEvent) error {
	var issueEvent models.IssuesEvent
	if err := event.ParsePayload(&issueEvent); err != nil {
		return fmt.Errorf("解析Issue事件失败: %v", err)
	}

	log.Printf("Issue事件: Action=%s, Issue=#%d, Title=%s",
		issueEvent.Action, issueEvent.Issue.Number, issueEvent.Issue.Title)

	switch issueEvent.Action {
	case "opened": // 处理Issue打开事件
		return ep.handleIssueOpened(&issueEvent)
	case "edited": // 处理Issue编辑事件
		return ep.handleIssueEdited(&issueEvent)
	case "closed": // 处理Issue关闭事件
		return ep.handleIssueClosed(&issueEvent)
	default: // 忽略其他Issue操作
		log.Printf("忽略Issue操作: %s", issueEvent.Action)
		return nil
	}
}

// handleIssueCommentEvent 处理Issue评论事件
func (ep *EventProcessor) handleIssueCommentEvent(event *models.GitHubEvent) error {
	var commentEvent models.IssueCommentEvent
	if err := event.ParsePayload(&commentEvent); err != nil {
		return fmt.Errorf("解析Issue评论事件失败: %v", err)
	}

	log.Printf("Issue评论事件: Action=%s, Issue=#%d, Comment=%s",
		commentEvent.Action, commentEvent.Issue.Number,
		ep.truncateString(commentEvent.Comment.Body, 50))

	if commentEvent.Action == "created" {
		return ep.handleCommentCreated(&commentEvent)
	}

	return nil
}

// handlePullRequestEvent 处理Pull Request事件
func (ep *EventProcessor) handlePullRequestEvent(event *models.GitHubEvent) error {
	var prEvent models.PullRequestEvent
	if err := event.ParsePayload(&prEvent); err != nil {
		return fmt.Errorf("解析Pull Request事件失败: %v", err)
	}

	log.Printf("Pull Request事件: Action=%s, PR=#%d, Title=%s",
		prEvent.Action, prEvent.PullRequest.Number, prEvent.PullRequest.Title)

	switch prEvent.Action {
	case "opened":
		return ep.handlePullRequestOpened(&prEvent)
	case "synchronize":
		return ep.handlePullRequestSynchronized(&prEvent)
	case "closed":
		return ep.handlePullRequestClosed(&prEvent)
	default:
		log.Printf("忽略Pull Request操作: %s", prEvent.Action)
		return nil
	}
}

// handlePullRequestReviewCommentEvent 处理PR Review评论事件
func (ep *EventProcessor) handlePullRequestReviewCommentEvent(event *models.GitHubEvent) error {
	var reviewCommentEvent models.PullRequestReviewCommentEvent
	if err := event.ParsePayload(&reviewCommentEvent); err != nil {
		return fmt.Errorf("解析PR Review评论事件失败: %v", err)
	}

	log.Printf("PR Review评论事件: Action=%s, PR=#%d",
		reviewCommentEvent.Action, reviewCommentEvent.PullRequest.Number)

	if reviewCommentEvent.Action == "created" {
		return ep.handleReviewCommentCreated(&reviewCommentEvent)
	}

	return nil
}

// handlePullRequestReviewEvent 处理PR Review事件
func (ep *EventProcessor) handlePullRequestReviewEvent(event *models.GitHubEvent) error {
	var reviewEvent models.PullRequestReviewEvent
	if err := event.ParsePayload(&reviewEvent); err != nil {
		return fmt.Errorf("解析PR Review事件失败: %v", err)
	}

	log.Printf("PR Review事件: Action=%s, PR=#%d, Review State=%s",
		reviewEvent.Action, reviewEvent.PullRequest.Number, reviewEvent.Review.State)

	if reviewEvent.Action == "submitted" {
		return ep.handleReviewSubmitted(&reviewEvent)
	}

	return nil
}

// handlePingEvent 处理Ping事件
func (ep *EventProcessor) handlePingEvent(event *models.GitHubEvent) error {
	log.Println("收到GitHub Webhook Ping事件")

	// 解析ping事件获取仓库信息
	var pingData map[string]interface{}
	if err := json.Unmarshal(event.Payload, &pingData); err != nil {
		return fmt.Errorf("解析ping事件失败: %v", err)
	}

	if repo, ok := pingData["repository"].(map[string]interface{}); ok {
		if fullName, ok := repo["full_name"].(string); ok {
			log.Printf("Webhook已成功连接到仓库: %s", fullName)
		}
	}

	return nil
}

// handleIssueOpened 处理Issue打开事件
func (ep *EventProcessor) handleIssueOpened(event *models.IssuesEvent) error {
	log.Printf("新Issue创建: #%d - %s", event.Issue.Number, event.Issue.Title)

	// 检查Issue描述中是否包含命令
	if command := ep.extractCommand(event.Issue.Body); command != nil {
		log.Printf("在Issue中检测到命令: %s", command.Command)

		return ep.executeCommand(command, &CommandContext{
			Repository: event.Repository,
			Issue:      &event.Issue,
			User:       event.Sender,
		})
	}

	// 如果没有检测到命令，则不进行自动修改
	log.Printf("Issue #%d 未包含任何指令，跳过自动修改", event.Issue.Number)
	return nil
}

// handleIssueEdited 处理Issue编辑事件
func (ep *EventProcessor) handleIssueEdited(event *models.IssuesEvent) error {
	log.Printf("Issue已编辑: #%d - %s", event.Issue.Number, event.Issue.Title)
	return nil
}

// handleIssueClosed 处理Issue关闭事件
func (ep *EventProcessor) handleIssueClosed(event *models.IssuesEvent) error {
	log.Printf("Issue已关闭: #%d - %s", event.Issue.Number, event.Issue.Title)
	return nil
}

// handleCommentCreated 处理评论创建事件
func (ep *EventProcessor) handleCommentCreated(event *models.IssueCommentEvent) error {
	// 检查是否是PR评论
	if event.PullRequest != nil {
		log.Printf("新PR评论创建: PR #%d, User: %s",
			event.PullRequest.Number, event.Comment.User.Login)
	} else {
		log.Printf("新Issue评论创建: Issue #%d, User: %s",
			event.Issue.Number, event.Comment.User.Login)
	}

	// 检查评论中是否包含命令
	if command := ep.extractCommand(event.Comment.Body); command != nil {
		log.Printf("在评论中检测到命令: %s", command.Command)

		// 构建CommandContext，如果是PR评论则包含PR信息
		ctx := &CommandContext{
			Repository: event.Repository,
			Issue:      &event.Issue,
			Comment:    &event.Comment,
			User:       event.Sender,
		}

		// 如果是PR评论，添加PR信息到上下文
		if event.PullRequest != nil {
			ctx.PullRequest = event.PullRequest
		}

		return ep.executeCommand(command, ctx)
	}

	return nil
}

// handlePullRequestOpened 处理Pull Request打开事件
func (ep *EventProcessor) handlePullRequestOpened(event *models.PullRequestEvent) error {
	log.Printf("新Pull Request创建: #%d - %s",
		event.PullRequest.Number, event.PullRequest.Title)
	return nil
}

// handlePullRequestSynchronized 处理Pull Request同步事件
func (ep *EventProcessor) handlePullRequestSynchronized(event *models.PullRequestEvent) error {
	log.Printf("Pull Request已同步: #%d", event.PullRequest.Number)
	return nil
}

// handlePullRequestClosed 处理Pull Request关闭事件
func (ep *EventProcessor) handlePullRequestClosed(event *models.PullRequestEvent) error {
	log.Printf("Pull Request已关闭: #%d - Merged: %t",
		event.PullRequest.Number, event.PullRequest.Merged)
	return nil
}

// handleReviewCommentCreated 处理Review评论创建事件
func (ep *EventProcessor) handleReviewCommentCreated(event *models.PullRequestReviewCommentEvent) error {
	log.Printf("新Review评论创建: PR #%d, User: %s",
		event.PullRequest.Number, event.Comment.User.Login)

	// 检查评论中是否包含命令
	if command := ep.extractCommand(event.Comment.Body); command != nil {
		log.Printf("在Review评论中检测到命令: %s", command.Command)

		return ep.executeCommand(command, &CommandContext{
			Repository:  event.Repository,
			PullRequest: &event.PullRequest,
			Comment:     &event.Comment,
			User:        event.Sender,
		})
	}

	return nil
}

// handleReviewSubmitted 处理Review提交事件
func (ep *EventProcessor) handleReviewSubmitted(event *models.PullRequestReviewEvent) error {
	log.Printf("PR Review提交: PR #%d, State: %s, User: %s",
		event.PullRequest.Number, event.Review.State, event.Review.User.Login)

	// 检查Review内容中是否包含命令
	if command := ep.extractCommand(event.Review.Body); command != nil {
		log.Printf("在Review中检测到命令: %s", command.Command)

		// 构建模拟的Comment用于命令执行
		reviewComment := &models.Comment{
			ID:        event.Review.ID,
			Body:      event.Review.Body,
			User:      event.Review.User,
			HTMLURL:   event.Review.HTMLURL,
			CreatedAt: event.Review.SubmittedAt,
			UpdatedAt: event.Review.SubmittedAt,
		}

		return ep.executeCommand(command, &CommandContext{
			Repository:  event.Repository,
			PullRequest: &event.PullRequest,
			Comment:     reviewComment,
			User:        event.Sender,
		})
	}

	// 如果是请求更改的Review，可以自动触发代码审查
	if event.Review.State == "changes_requested" && event.Review.Body != "" {
		log.Printf("检测到请求更改的Review，可能需要自动审查")

		// 自动触发Review命令
		reviewCommand := &Command{
			Command: "review",
			Args:    "分析PR变更并提供改进建议",
		}

		// 构建模拟的Comment用于命令执行
		reviewComment := &models.Comment{
			ID:        event.Review.ID,
			Body:      event.Review.Body,
			User:      event.Review.User,
			HTMLURL:   event.Review.HTMLURL,
			CreatedAt: event.Review.SubmittedAt,
			UpdatedAt: event.Review.SubmittedAt,
		}

		return ep.executeCommand(reviewCommand, &CommandContext{
			Repository:  event.Repository,
			PullRequest: &event.PullRequest,
			Comment:     reviewComment,
			User:        event.Sender,
		})
	}

	return nil
}

// Command 命令结构
type Command struct {
	Command string
	Args    string
}

// CommandContext 命令执行上下文
type CommandContext struct {
	Repository  models.Repository
	Issue       *models.Issue
	PullRequest *models.PullRequest
	Comment     *models.Comment
	User        models.User
}

// extractCommand 从文本中提取命令
func (ep *EventProcessor) extractCommand(text string) *Command {
	lines := strings.Split(strings.TrimSpace(text), "\n")

	// TODO 这里有个小bug
	for _, line := range lines {
		line = strings.TrimSpace(line)
		matches := ep.commandRegex.FindStringSubmatch(line)
		if len(matches) >= 2 {
			return &Command{
				Command: matches[1],
				Args:    strings.TrimSpace(matches[2]),
			}
		}
	}

	return nil
}

// executeCommand 执行命令
func (ep *EventProcessor) executeCommand(command *Command, ctx *CommandContext) error {
	log.Printf("执行命令: %s, 参数: %s", command.Command, command.Args)

	switch command.Command {
	case "code": // 适合用于：功能开发、逻辑变更、结构调整 （/code 是基于Issue描述进行修改）
		return ep.handleCodeCommand(command, ctx)
	case "continue": // 适合用于：继续开发、功能扩展、逻辑优化（需要先/code，在功能实现上和code不同的点在于：/continue 是基于/code的代码进行修改，而/code是基于Issue描述进行修改）
		return ep.handleContinueCommand(command, ctx)
	case "fix": // 适合用于：代码修复、错误修复、性能优化
		return ep.handleFixCommand(command, ctx)
	case "help":
		return ep.handleHelpCommand(command, ctx)
	case "summary": // 适合用于：总结代码、总结问题、总结需求
		return ep.handleSummaryCommand(command, ctx)
	case "review": // 适合用于：代码审查、代码优化、代码重构
		return ep.handleReviewCommand(command, ctx)
	default:
		return fmt.Errorf("未知命令: %s", command.Command)
	}
}

// handleSummaryCommand 处理总结命令
func (ep *EventProcessor) handleSummaryCommand(command *Command, ctx *CommandContext) error {
	log.Printf("处理总结命令: %s", command.Args)

	// 获取分支名
	sourceBranch := "main"
	if ctx.Repository.DefaultBranch != "" {
		sourceBranch = ctx.Repository.DefaultBranch
	}

	// 克隆仓库
	repoPath, err := ep.gitService.CloneRepository(ctx.Repository.CloneURL, sourceBranch)
	if err != nil {
		log.Printf("克隆仓库失败: %v", err)
		response := fmt.Sprintf("❌ 总结失败: 克隆仓库失败 - %v", err.Error())
		return ep.createResponse(ctx, response)
	}

	// 清理工作目录
	defer func() {
		if err := ep.gitService.Cleanup(repoPath); err != nil {
			log.Printf("清理工作目录失败: %v", err)
		}
	}()

	// 获取文件树
	log.Printf("开始获取文件树: %s", repoPath)
	fileTree, err := ep.gitService.GetFileTree(repoPath)
	if err != nil {
		log.Printf("获取文件树失败: %v", err)
		fileTree = "无法获取文件树"
	} else {
		log.Printf("文件树获取成功，长度: %d 字符", len(fileTree))
		if len(fileTree) > 50 {
			maxLen := 500
			if len(fileTree) < maxLen {
				maxLen = len(fileTree)
			}
			log.Printf("文件树前%d字符: %s", maxLen, fileTree[:maxLen])
		}
	}

	// 生成项目上下文信息
	projectContext := ep.buildProjectContext(ctx)

	// 构建总结提示词，包含文件结构信息
	summaryPrompt := fmt.Sprintf(`请根据以下项目信息生成简明扼要的总结：

【项目上下文】
%s

【项目结构】
%s

【总结需求】
%s

请分析项目的核心功能、技术栈、主要文件结构，并提供简洁明了的总结。`, projectContext, fileTree, command.Args)

	// 在目标仓库目录中调用Claude Code CLI进行总结
	summary, err := ep.claudeCodeService.SummarizeInRepo(summaryPrompt, repoPath)
	if err != nil {
		log.Printf("Claude Code CLI总结失败: %v", err)
		response := fmt.Sprintf("❌ 总结失败: %v", err)
		return ep.createResponse(ctx, response)
	}

	// 回复总结内容
	response := fmt.Sprintf("📝 **总结结果：**\n\n%s", summary)
	return ep.createResponse(ctx, response)
}

// handleReviewCommand 处理代码审查命令
func (ep *EventProcessor) handleReviewCommand(command *Command, ctx *CommandContext) error {
	log.Printf("处理代码审查命令: %s", command.Args)

	// 检查是否是PR上下文
	if ctx.PullRequest != nil {
		log.Printf("检测到PR上下文，执行PR代码审查: PR #%d", ctx.PullRequest.Number)
		// PR代码审查：分析具体的代码变更
		return ep.handlePullRequestReview(command, ctx)
	} else {
		log.Printf("检测到Issue上下文，执行一般代码审查: Issue #%d", ctx.Issue.Number)
		// Issue代码审查：审查整个项目
		return ep.handleGeneralReview(command, ctx)
	}
}

// handlePullRequestReview 处理PR代码审查
func (ep *EventProcessor) handlePullRequestReview(command *Command, ctx *CommandContext) error {
	log.Printf("处理PR代码审查: PR #%d", ctx.PullRequest.Number)

	// 克隆仓库（使用基础分支）
	baseBranch := ctx.PullRequest.Base.Ref
	repoPath, err := ep.gitService.CloneRepository(ctx.Repository.CloneURL, baseBranch)
	if err != nil {
		log.Printf("克隆仓库失败: %v", err)
		response := fmt.Sprintf("❌ PR代码审查失败: 克隆仓库失败 - %v", err.Error())
		return ep.createResponse(ctx, response)
	}

	// 清理工作目录
	defer func() {
		if err := ep.gitService.Cleanup(repoPath); err != nil {
			log.Printf("清理工作目录失败: %v", err)
		}
	}()

	// 获取PR的diff信息
	prDiff, err := ep.gitService.GetPullRequestDiff(repoPath, ctx.PullRequest.Head.SHA, ctx.PullRequest.Base.SHA)
	if err != nil {
		log.Printf("获取PR diff失败: %v", err)
		prDiff = "无法获取PR差异信息"
	}

	// 构建PR审查提示词
	reviewScope := "PR代码变更"
	if command.Args != "" {
		reviewScope = command.Args
	}

	reviewPrompt := fmt.Sprintf(`请对以下Pull Request的代码变更进行专业审查：

**Pull Request信息:**
- PR #%d: %s
- 分支: %s -> %s
- 状态: %s
- 创建者: %s

**审查范围:** %s

**代码变更内容:**
%s

**审查要点:**
1. **代码变更质量** - 修改是否合理、清晰
2. **安全性** - 新代码是否引入安全漏洞
3. **性能影响** - 变更对性能的影响
4. **最佳实践** - 是否遵循编码规范
5. **潜在问题** - bug风险、边界条件
6. **向后兼容性** - 是否破坏现有功能
7. **测试覆盖** - 是否需要添加测试

**输出格式:**
请提供结构化的PR审查报告：
- **总体评价** - 对这次PR的整体评估
- **主要变更分析** - 列出关键的代码修改点
- **发现的问题** - 按严重程度分类（严重/中等/轻微）
- **改进建议** - 具体的修改建议
- **合并建议** - 是否建议合并及原因

请用markdown格式输出。`,
		ctx.PullRequest.Number, ctx.PullRequest.Title,
		ctx.PullRequest.Head.Ref, ctx.PullRequest.Base.Ref,
		ctx.PullRequest.State, ctx.PullRequest.User.Login,
		reviewScope, prDiff)

	// 在目标仓库目录中调用Claude Code CLI进行PR审查
	reviewResult, err := ep.claudeCodeService.ReviewCodeInRepo(reviewPrompt, repoPath)
	if err != nil {
		log.Printf("Claude Code CLI代码审查失败: %v", err)
		response := fmt.Sprintf(`❌ **PR代码审查失败**

错误信息: %s

请检查:
1. Claude API密钥是否正确配置
2. 网络连接是否正常
3. API配额是否充足

---
*处理时间: %s*`, err.Error(), time.Now().Format("2006-01-02 15:04:05"))
		return ep.createResponse(ctx, response)
	}

	// 生成审查报告
	response := fmt.Sprintf(`🔍 **PR代码审查报告**

**PR信息:** #%d - %s

%s

**审查流程:**
1. ✅ 克隆目标仓库
2. ✅ 获取PR代码差异
3. ✅ 分析代码变更
4. ✅ 评估安全性和质量
5. ✅ 生成审查报告

---
*审查时间: %s*
*由AI代码审查助手生成*`, ctx.PullRequest.Number, ctx.PullRequest.Title, reviewResult, time.Now().Format("2006-01-02 15:04:05"))

	return ep.createResponse(ctx, response)
}

// handleGeneralReview 处理一般代码审查（Issue上下文）
func (ep *EventProcessor) handleGeneralReview(command *Command, ctx *CommandContext) error {
	log.Printf("处理一般代码审查")

	// 获取分支名
	sourceBranch := "main"
	if ctx.Repository.DefaultBranch != "" {
		sourceBranch = ctx.Repository.DefaultBranch
	}

	// 克隆仓库
	repoPath, err := ep.gitService.CloneRepository(ctx.Repository.CloneURL, sourceBranch)
	if err != nil {
		log.Printf("克隆仓库失败: %v", err)
		response := fmt.Sprintf("❌ 代码审查失败: 克隆仓库失败 - %v", err.Error())
		return ep.createResponse(ctx, response)
	}

	// 清理工作目录
	defer func() {
		if err := ep.gitService.Cleanup(repoPath); err != nil {
			log.Printf("清理工作目录失败: %v", err)
		}
	}()

	// 获取文件树
	fileTree, err := ep.gitService.GetFileTree(repoPath)
	if err != nil {
		log.Printf("获取文件树失败: %v", err)
		fileTree = "无法获取文件树"
	}

	// 构建项目上下文
	context := ep.buildProjectContext(ctx)

	// 确定审查范围
	reviewScope := "整个项目"
	if command.Args != "" {
		reviewScope = command.Args
	}

	// 构建代码审查提示词
	reviewPrompt := fmt.Sprintf(`请对以下代码进行专业的代码审查：

**审查范围:** %s

**项目上下文:**
%s

**项目结构:**
%s

**审查要点:**
1. 代码质量和可读性
2. 安全性问题
3. 性能优化建议
4. 最佳实践遵循
5. 潜在的bug或问题
6. 架构设计合理性
7. 测试覆盖度

**输出格式:**
请提供结构化的审查报告，包括：
- 总体评价
- 具体问题和建议
- 代码改进点
- 安全性评估
- 性能分析

请用markdown格式输出。`, reviewScope, context, fileTree)

	// 在目标仓库目录中调用Claude Code CLI进行代码审查
	reviewResult, err := ep.claudeCodeService.ReviewCodeInRepo(reviewPrompt, repoPath)
	if err != nil {
		log.Printf("Claude Code CLI代码审查失败: %v", err)
		response := fmt.Sprintf(`❌ **代码审查失败**

错误信息: %s

请检查:
1. Claude API密钥是否正确配置
2. 网络连接是否正常
3. API配额是否充足

---
*处理时间: %s*`, err.Error(), time.Now().Format("2006-01-02 15:04:05"))
		return ep.createResponse(ctx, response)
	}

	// 生成审查报告
	response := fmt.Sprintf(`🔍 **代码审查报告**

**审查范围:** %s

%s

**审查流程:**
1. ✅ 克隆目标仓库
2. ✅ 分析项目结构
3. ✅ 检查代码质量
4. ✅ 评估安全性
5. ✅ 生成审查报告

---
*审查时间: %s*
*由AI代码审查助手生成*`, reviewScope, reviewResult, time.Now().Format("2006-01-02 15:04:05"))

	return ep.createResponse(ctx, response)
}

// handleCodeCommand 处理代码生成命令
func (ep *EventProcessor) handleCodeCommand(command *Command, ctx *CommandContext) error {
	log.Printf("处理代码生成命令: %s", command.Args)
	log.Printf("启动自动代码分析和修改流程")

	// 创建一个临时Issue，将原Issue内容作为上下文，评论内容作为具体需求
	modifiedIssue := *ctx.Issue

	// 拼接原Issue内容和评论内容
	combinedBody := fmt.Sprintf(`**原Issue内容:**
%s

**当前代码修改需求:**
%s`, ctx.Issue.Body, command.Args)

	modifiedIssue.Body = combinedBody
	modifiedIssue.Title = fmt.Sprintf("代码修改请求: %s", command.Args)

	// 构造IssuesEvent结构用于自动修改
	issuesEvent := &models.IssuesEvent{
		Action:     "opened",
		Issue:      modifiedIssue,
		Repository: ctx.Repository,
		Sender:     ctx.User,
	}

	// 直接调用自动分析和修改功能
	return ep.autoAnalyzeAndModify(issuesEvent)
}

// handleContinueCommand 处理继续命令
func (ep *EventProcessor) handleContinueCommand(command *Command, ctx *CommandContext) error {
	log.Printf("处理继续命令: %s", command.Args)

	// 构建项目上下文
	context := ep.buildProjectContext(ctx)

	// 调用Claude Code CLI继续开发
	continuedCode, err := ep.claudeCodeService.ContinueCode(command.Args, context)
	if err != nil {
		log.Printf("Claude Code CLI调用失败: %v", err)
		response := fmt.Sprintf(`❌ **继续开发失败**

错误信息: %s

请检查:
1. Claude API密钥是否正确配置
2. 网络连接是否正常
3. API配额是否充足

---
*处理时间: %s*`, err.Error(), time.Now().Format("2006-01-02 15:04:05"))
		return ep.createResponse(ctx, response)
	}

	response := fmt.Sprintf(`🔄 **继续开发**

%s

**处理流程:**
1. ✅ 获取当前进度
2. ✅ 分析历史上下文
3. ✅ 继续代码生成完成

**继续开发的代码:**

%s

---
*处理时间: %s*`, command.Args, continuedCode, time.Now().Format("2006-01-02 15:04:05"))

	return ep.createResponse(ctx, response)
}

// handleFixCommand 处理修复命令
func (ep *EventProcessor) handleFixCommand(command *Command, ctx *CommandContext) error {
	log.Printf("处理修复命令: %s", command.Args)

	// 构建项目上下文
	context := ep.buildProjectContext(ctx)

	// 调用Claude Code CLI修复代码
	fixedCode, err := ep.claudeCodeService.FixCode(command.Args, context)
	if err != nil {
		log.Printf("Claude Code CLI调用失败: %v", err)
		response := fmt.Sprintf(`❌ **代码修复失败**

错误信息: %s

请检查:
1. Claude API密钥是否正确配置
2. 网络连接是否正常
3. API配额是否充足

---
*处理时间: %s*`, err.Error(), time.Now().Format("2006-01-02 15:04:05"))
		return ep.createResponse(ctx, response)
	}

	response := fmt.Sprintf(`🔧 **代码修复**

问题描述: %s

**修复流程:**
1. ✅ 分析问题
2. ✅ 定位错误代码
3. ✅ 生成修复方案
4. ✅ 应用修复完成

**修复后的代码:**

%s

---
*处理时间: %s*`, command.Args, fixedCode, time.Now().Format("2006-01-02 15:04:05"))

	return ep.createResponse(ctx, response)
}

// handleHelpCommand 处理帮助命令
func (ep *EventProcessor) handleHelpCommand(command *Command, ctx *CommandContext) error {
	log.Printf("处理帮助命令")

	response := `📖 **CodeAgent 帮助**

**支持的命令:**

🔹 ` + "`" + `/code <需求描述>` + "`" + ` - 自动分析并实现到代码库
🔹 ` + "`" + `/continue [说明]` + "`" + ` - 继续当前的开发任务
🔹 ` + "`" + `/fix <问题描述>` + "`" + ` - 修复指定的代码问题
🔹 ` + "`" + `/review [范围]` + "`" + ` - 对代码进行专业审查
🔹 ` + "`" + `/summary [内容]` + "`" + ` - 生成项目或内容总结
🔹 ` + "`" + `/help` + "`" + ` - 显示此帮助信息

**使用示例:**
- ` + "`" + `/code 创建一个用户登录API` + "`" + ` - 自动分析并实现到项目中
- ` + "`" + `/code 添加JWT认证功能` + "`" + ` - 自动分析并修改代码
- ` + "`" + `/continue 添加数据验证逻辑` + "`" + `
- ` + "`" + `/fix 修复空指针异常` + "`" + `
- ` + "`" + `/review 安全性审查` + "`" + ` - 审查代码安全性问题
- ` + "`" + `/summary 当前PR的主要变更` + "`" + ` - 总结PR内容

**工作流程:**
1. 🎯 在Issue或PR评论中输入命令
2. 🤖 AI分析需求并生成代码
3. 🌲 创建独立的Git工作空间
4. 📝 自动提交代码并创建PR
5. 💬 在GitHub界面展示结果

---
*GitHub Webhook Demo v1.0*`

	return ep.createResponse(ctx, response)
}

// createResponse 创建响应
func (ep *EventProcessor) createResponse(ctx *CommandContext, response string) error {
	repo := strings.Split(ctx.Repository.FullName, "/")
	if len(repo) != 2 {
		return fmt.Errorf("无效的仓库名称: %s", ctx.Repository.FullName)
	}

	owner, repoName := repo[0], repo[1]

	// 根据上下文选择响应方式
	if ctx.Issue != nil {
		// 在Issue中回复
		return ep.githubService.CreateComment(owner, repoName, ctx.Issue.Number, response)
	} else if ctx.PullRequest != nil {
		// 在PR中回复
		return ep.githubService.CreateComment(owner, repoName, ctx.PullRequest.Number, response)
	}

	return fmt.Errorf("无法确定响应位置")
}

// getBranchName 动态获取事件关联的分支名
func (ep *EventProcessor) getBranchName(event *models.GitHubEvent, ctx *CommandContext) string {
	// 如果是Pull Request相关事件，优先使用base分支
	if ctx.PullRequest != nil {
		return ctx.PullRequest.Base.Ref
	}

	// 使用仓库的默认分支
	if event.Repository.DefaultBranch != "" {
		return event.Repository.DefaultBranch
	}

	// 如果默认分支为空，fallback到"main"
	return "main"
}

// buildProjectContext 构建项目上下文
func (ep *EventProcessor) buildProjectContext(ctx *CommandContext) string {
	var context strings.Builder

	// 添加仓库信息
	context.WriteString(fmt.Sprintf("**仓库信息:**\n"))
	context.WriteString(fmt.Sprintf("- 仓库: %s\n", ctx.Repository.FullName))
	context.WriteString(fmt.Sprintf("- 名称: %s\n", ctx.Repository.Name))
	context.WriteString(fmt.Sprintf("- URL: %s\n", ctx.Repository.HTMLURL))
	context.WriteString(fmt.Sprintf("- 默认分支: %s\n", ctx.Repository.DefaultBranch))

	// 添加Issue信息
	if ctx.Issue != nil {
		context.WriteString(fmt.Sprintf("\n**Issue信息:**\n"))
		context.WriteString(fmt.Sprintf("- 标题: %s\n", ctx.Issue.Title))
		context.WriteString(fmt.Sprintf("- 编号: #%d\n", ctx.Issue.Number))
		context.WriteString(fmt.Sprintf("- 状态: %s\n", ctx.Issue.State))
		context.WriteString(fmt.Sprintf("- 创建者: %s\n", ctx.Issue.User.Login))
		context.WriteString(fmt.Sprintf("- 创建时间: %s\n", ctx.Issue.CreatedAt.Format("2006-01-02 15:04:05")))

		// 处理标签
		if len(ctx.Issue.Labels) > 0 {
			var labelNames []string
			for _, label := range ctx.Issue.Labels {
				labelNames = append(labelNames, label.Name)
			}
			context.WriteString(fmt.Sprintf("- 标签: %s\n", strings.Join(labelNames, ", ")))
		}

		// 添加Issue描述（限制长度避免上下文过长）
		description := ctx.Issue.Body
		if len(description) > 1000 {
			description = description[:1000] + "...(内容已截断)"
		}
		context.WriteString(fmt.Sprintf("- 描述: %s\n", description))
	}

	// 添加Pull Request信息
	if ctx.PullRequest != nil {
		context.WriteString(fmt.Sprintf("\n**Pull Request信息:**\n"))
		context.WriteString(fmt.Sprintf("- 标题: %s\n", ctx.PullRequest.Title))
		context.WriteString(fmt.Sprintf("- 编号: #%d\n", ctx.PullRequest.Number))
		context.WriteString(fmt.Sprintf("- 状态: %s\n", ctx.PullRequest.State))
		context.WriteString(fmt.Sprintf("- 分支: %s -> %s\n", ctx.PullRequest.Head.Ref, ctx.PullRequest.Base.Ref))
		context.WriteString(fmt.Sprintf("- 创建者: %s\n", ctx.PullRequest.User.Login))

		// 添加PR描述
		if ctx.PullRequest.Body != "" {
			description := ctx.PullRequest.Body
			if len(description) > 800 {
				description = description[:800] + "...(内容已截断)"
			}
			context.WriteString(fmt.Sprintf("- 描述: %s\n", description))
		}
	}

	// 添加评论信息
	if ctx.Comment != nil {
		context.WriteString(fmt.Sprintf("\n**最新评论:**\n"))
		context.WriteString(fmt.Sprintf("- 评论者: %s\n", ctx.Comment.User.Login))
		context.WriteString(fmt.Sprintf("- 时间: %s\n", ctx.Comment.CreatedAt.Format("2006-01-02 15:04:05")))

		commentBody := ctx.Comment.Body
		if len(commentBody) > 500 {
			commentBody = commentBody[:500] + "...(内容已截断)"
		}
		context.WriteString(fmt.Sprintf("- 内容: %s\n", commentBody))
	}

	// 添加用户信息
	context.WriteString(fmt.Sprintf("\n**用户信息:**\n"))
	context.WriteString(fmt.Sprintf("- 用户: %s\n", ctx.User.Login))
	context.WriteString(fmt.Sprintf("- 用户ID: %d\n", ctx.User.ID))

	return context.String()
}

// buildEnhancedProjectContext 构建增强的项目上下文（包含文件结构）
func (ep *EventProcessor) buildEnhancedProjectContext(ctx *CommandContext, repoPath string) string {
	var context strings.Builder

	// 基础上下文
	context.WriteString(ep.buildProjectContext(ctx))

	// 添加项目结构信息
	if repoPath != "" {
		fileTree, err := ep.gitService.GetFileTree(repoPath)
		if err != nil {
			log.Printf("获取文件树失败: %v", err)
			context.WriteString(fmt.Sprintf("\n**项目结构:**\n无法获取文件树信息\n"))
		} else {
			context.WriteString(fmt.Sprintf("\n**项目结构:**\n```\n%s\n```\n", fileTree))
		}
	}

	return context.String()
}

// autoAnalyzeAndModify 自动分析Issue并修改代码
func (ep *EventProcessor) autoAnalyzeAndModify(event *models.IssuesEvent) error {
	log.Printf("开始自动分析Issue: #%d", event.Issue.Number)

	// 检查是否已经有相同的分支存在，避免重复处理
	branchName := fmt.Sprintf("auto-fix-issue-%d", event.Issue.Number)

	// 简单的防重复机制：检查分支是否已经存在
	// 这里可以添加更复杂的检查逻辑
	log.Printf("准备创建分支: %s", branchName)

	// 构造CommandContext用于获取分支名
	ctx := &CommandContext{
		Repository: event.Repository,
		Issue:      &event.Issue,
		User:       event.Sender,
	}

	// 创建GitHub事件结构用于分支名获取
	gitHubEvent := &models.GitHubEvent{
		Repository: event.Repository,
		Issue:      event.Issue,
		Sender:     event.Sender,
	}

	// 动态获取分支名
	sourceBranch := ep.getBranchName(gitHubEvent, ctx)
	log.Printf("使用分支: %s", sourceBranch)

	// 克隆仓库
	repoPath, err := ep.gitService.CloneRepository(event.Repository.CloneURL, sourceBranch)
	if err != nil {
		log.Printf("克隆仓库失败: %v", err)
		errorMsg := fmt.Sprintf("自动分析失败: 克隆仓库失败 - %v", err.Error())
		return ep.createResponse(&CommandContext{
			Repository: event.Repository,
			Issue:      &event.Issue,
			User:       event.Sender,
		}, errorMsg)
	}

	// 清理工作目录
	defer func() {
		if err := ep.gitService.Cleanup(repoPath); err != nil {
			log.Printf("清理工作目录失败: %v", err)
		}
	}()

	log.Printf("仓库路径: %s", repoPath)

	// 配置Git用户
	if err := ep.gitService.ConfigureGit(repoPath, "CodeAgent", "codeagent@example.com"); err != nil {
		log.Printf("配置Git失败: %v", err)
	}

	log.Printf("准备在仓库目录中直接进行代码修改")

	// 创建新分支，使用带时间戳的分支名避免冲突
	timestamp := time.Now().Format("20060102-150405")
	branchName = fmt.Sprintf("auto-fix-issue-%d-%s", event.Issue.Number, timestamp)
	log.Printf("创建分支: %s", branchName)
	if err := ep.gitService.CreateBranch(repoPath, branchName); err != nil {
		log.Printf("创建分支失败: %v", err)
	}

	// 创建GitHub事件包装结构用于新的方法
	gitHubEventForModification := &models.GitHubEvent{
		Type:       "issues",
		Repository: event.Repository,
		Issue:      event.Issue,
		Sender:     event.Sender,
	}

	// 直接在仓库目录中调用Claude Code CLI进行代码修改
	modificationPrompt := fmt.Sprintf(`请根据以下需求修改代码：

**需求：**
%s

**描述：**
%s

**说明：**
- 请直接修改需要的文件
- 确保代码可以正常运行
- 遵循最佳实践`, event.Issue.Title, event.Issue.Body)

	modificationResult, err := ep.claudeCodeService.GenerateCodeInRepo(modificationPrompt, repoPath)
	if err != nil {
		log.Printf("Claude Code CLI代码修改失败: %v", err)
		errorMsg := fmt.Sprintf("自动修改失败: %v", err.Error())
		return ep.createResponse(&CommandContext{
			Repository: event.Repository,
			Issue:      &event.Issue,
			User:       event.Sender,
		}, errorMsg)
	}

	// 提交修改到仓库
	commitResult, err := ep.commitAndPushChanges(repoPath, gitHubEventForModification, branchName, sourceBranch)
	if err != nil {
		log.Printf("提交代码失败: %v", err)
		errorMsg := fmt.Sprintf("代码提交失败: %v", err.Error())
		return ep.createResponse(&CommandContext{
			Repository: event.Repository,
			Issue:      &event.Issue,
			User:       event.Sender,
		}, errorMsg)
	}

	// 在Issue中回复
	response := fmt.Sprintf(`🤖 **自动修复已完成**

## Issue信息
- **标题**: %s
- **编号**: #%d

## 处理流程
1. ✅ 克隆仓库
2. ✅ AI分析Issue需求  
3. ✅ 创建修复分支: %s
4. ✅ 应用代码修改
5. ✅ 提交更改到仓库
6. ✅ 推送到远程分支
7. ✅ 创建Pull Request

## 修改结果
%s

## 提交信息  
%s

## 下一步
请在以下Pull Request中review代码修改，确认无误后进行合并。

---
*此回复由AI助手自动生成*`,
		event.Issue.Title, event.Issue.Number,
		fmt.Sprintf("auto-fix-issue-%d", event.Issue.Number),
		modificationResult, commitResult)

	return ep.createResponse(&CommandContext{
		Repository: event.Repository,
		Issue:      &event.Issue,
		User:       event.Sender,
	}, response)
}

// truncateString 截断字符串
func (ep *EventProcessor) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// applyCodeModifications 根据AI分析结果应用代码修改
func (ep *EventProcessor) applyCodeModifications(repoPath, analysisResult string, event *models.GitHubEvent) (string, error) {
	log.Printf("开始应用代码修改，基于AI分析结果")

	// 构建更具体的代码修改提示
	modificationPrompt := fmt.Sprintf(`你是一个专业的代码修改助手。请根据以下信息生成具体的代码修改方案：

**Issue信息:**
- 标题: %s
- 描述: %s
- 编号: #%d

**AI分析结果:**
%s

**重要提示：你必须直接返回JSON格式的代码修改方案，不要返回任何其他文本、解释或询问。**

**返回格式（必须是有效的JSON）:**
{
  "modifications": [
    {
      "file": "文件路径",
      "action": "create|modify|delete",
      "content": "文件的完整新内容（如果是modify或create）",
      "description": "修改说明"
    }
  ],
  "summary": "修改总结"
}

例如，如果要创建一个新文件，返回：
{
  "modifications": [
    {
      "file": "main.go",
      "action": "create",
      "content": "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello World\")\n}",
      "description": "创建主程序文件"
    }
  ],
  "summary": "根据需求创建了新的程序文件"
}

请立即返回JSON格式的修改方案，不要包含任何其他内容。`,
		event.Issue.Title, event.Issue.Body, event.Issue.Number, analysisResult)

	// 调用AI获取具体的修改方案
	modificationResult, err := ep.claudeCodeService.GenerateCode(modificationPrompt, "")
	if err != nil {
		return "", fmt.Errorf("获取代码修改方案失败: %v", err)
	}

	log.Printf("收到AI修改方案: %s", modificationResult)

	// 解析AI返回的JSON修改方案
	modifications, err := ep.parseModificationResult(modificationResult)
	if err != nil {
		return "", fmt.Errorf("解析修改方案失败: %v", err)
	}

	// 应用每个修改
	var appliedChanges []string
	for _, mod := range modifications {
		if err := ep.applyFileModification(repoPath, mod); err != nil {
			log.Printf("应用文件修改失败 %s: %v", mod.File, err)
			continue
		}
		appliedChanges = append(appliedChanges, fmt.Sprintf("- %s: %s", mod.File, mod.Description))
		log.Printf("成功修改文件: %s", mod.File)
	}

	if len(appliedChanges) == 0 {
		return "", fmt.Errorf("没有成功应用任何修改")
	}

	summary := fmt.Sprintf("成功应用 %d 个文件修改:\n%s",
		len(appliedChanges), strings.Join(appliedChanges, "\n"))

	return summary, nil
}

// FileModification 文件修改结构
type FileModification struct {
	File        string `json:"file"`
	Action      string `json:"action"`
	Content     string `json:"content"`
	Description string `json:"description"`
}

// ModificationResult 修改结果结构
type ModificationResult struct {
	Modifications []FileModification `json:"modifications"`
	Summary       string             `json:"summary"`
}

// parseModificationResult 解析AI返回的修改方案
func (ep *EventProcessor) parseModificationResult(result string) ([]FileModification, error) {
	// 尝试提取JSON部分
	jsonStart := strings.Index(result, "{")
	jsonEnd := strings.LastIndex(result, "}")

	if jsonStart == -1 || jsonEnd == -1 || jsonEnd <= jsonStart {
		return nil, fmt.Errorf("无法找到有效的JSON格式")
	}

	jsonStr := result[jsonStart : jsonEnd+1]

	var modResult ModificationResult
	if err := json.Unmarshal([]byte(jsonStr), &modResult); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %v", err)
	}

	return modResult.Modifications, nil
}

// applyFileModification 应用单个文件修改
func (ep *EventProcessor) applyFileModification(repoPath string, mod FileModification) error {
	switch mod.Action {
	case "create", "modify":
		return ep.gitService.WriteFile(repoPath, mod.File, mod.Content)
	case "delete":
		return ep.gitService.DeleteFile(repoPath, mod.File)
	default:
		return fmt.Errorf("不支持的操作类型: %s", mod.Action)
	}
}

// commitAndPushChanges 提交并推送代码修改
func (ep *EventProcessor) commitAndPushChanges(repoPath string, event *models.GitHubEvent, branchName, sourceBranch string) (string, error) {
	log.Printf("开始提交代码修改")

	// 添加所有修改的文件到暂存区
	if err := ep.gitService.AddFiles(repoPath, []string{"."}); err != nil {
		return "", fmt.Errorf("添加文件到暂存区失败: %v", err)
	}

	// 检查是否有修改
	hasChanges, err := ep.gitService.HasChanges(repoPath)
	if err != nil {
		return "", fmt.Errorf("检查修改状态失败: %v", err)
	}

	if !hasChanges {
		log.Printf("没有检测到代码修改，跳过提交")
		return "没有检测到代码修改", nil
	}

	// 获取修改的文件列表
	modifiedFiles, err := ep.gitService.GetModifiedFiles(repoPath)
	if err != nil {
		log.Printf("获取修改文件列表失败: %v", err)
		modifiedFiles = []string{} // 继续执行，使用空列表
	}

	// 使用CommitBuilder构建规范化的commit消息
	commitBuilder := NewCommitBuilder()
	commitMessage := commitBuilder.BuildAutoFixCommit(event, modifiedFiles)

	if err := ep.gitService.Commit(repoPath, commitMessage); err != nil {
		return "", fmt.Errorf("提交代码失败: %v", err)
	}

	// 推送到远程仓库
	log.Printf("推送分支: %s", branchName)
	if err := ep.gitService.Push(repoPath, branchName); err != nil {
		log.Printf("推送失败，错误信息: %v", err)
		return "", fmt.Errorf("推送代码失败: %v", err)
	}

	log.Printf("推送成功: %s", branchName)

	// 创建Pull Request
	prResult, err := ep.createPullRequest(event, branchName, sourceBranch)
	if err != nil {
		log.Printf("创建PR失败: %v", err)
		// PR创建失败不应该影响整个流程
	}

	result := fmt.Sprintf("✅ 代码修改已成功提交并推送到分支: %s", branchName)
	if prResult != "" {
		result += "\n" + prResult
	}

	return result, nil
}

// createPullRequest 创建Pull Request
func (ep *EventProcessor) createPullRequest(event *models.GitHubEvent, branchName, targetBranch string) (string, error) {
	// 使用CommitBuilder构建规范化的PR标题
	commitBuilder := NewCommitBuilder()
	title := commitBuilder.BuildPRCommit(event.Issue.Title, event.Issue.Body, event.Issue.Number)

	body := fmt.Sprintf(`## 自动生成的代码修改

此PR由AI助手自动生成，用于解决Issue #%d。

### 修改内容
- 基于Issue描述自动分析并生成代码修改
- 所有修改已经过AI验证

### 相关Issue
关闭 #%d

### 注意事项
请仔细review代码修改，确保符合项目要求后再合并。

---
*此PR由GitHub Webhook AI助手自动创建*`, event.Issue.Number, event.Issue.Number)

	pr, err := ep.githubService.CreatePullRequest(
		event.Repository.Owner.Login,
		event.Repository.Name,
		title,
		body,
		branchName,
		targetBranch, // 动态目标分支
	)

	if err != nil {
		// 如果是PR已存在的错误，不返回错误
		if strings.Contains(err.Error(), "A pull request already exists") {
			log.Printf("Pull Request 已存在，跳过创建: %s", branchName)
			return "🔗 Pull Request 已存在", nil
		}
		return "", err
	}

	return fmt.Sprintf("🔗 已创建Pull Request: %s", pr.HTMLURL), nil
}

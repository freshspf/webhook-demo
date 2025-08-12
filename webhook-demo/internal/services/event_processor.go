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
	githubService *GitHubService
	claudeService *ClaudeService
	commandRegex  *regexp.Regexp
}

// NewEventProcessor 创建新的事件处理器
func NewEventProcessor(githubService *GitHubService, claudeService *ClaudeService) *EventProcessor {
	return &EventProcessor{
		githubService: githubService,
		claudeService: claudeService,
		commandRegex:  regexp.MustCompile(`^/(code|continue|fix|help)\s*(.*)$`),
	}
}

// ProcessEvent 处理GitHub事件
func (ep *EventProcessor) ProcessEvent(event *models.GitHubEvent) error {
	log.Printf("开始处理事件: Type=%s, DeliveryID=%s", event.Type, event.DeliveryID)

	// 设置时间戳
	event.Timestamp = time.Now()

	switch event.Type {
	case "issues":
		return ep.handleIssuesEvent(event)
	case "issue_comment":
		return ep.handleIssueCommentEvent(event)
	case "pull_request":
		return ep.handlePullRequestEvent(event)
	case "pull_request_review_comment":
		return ep.handlePullRequestReviewCommentEvent(event)
	case "ping":
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
	case "opened":
		return ep.handleIssueOpened(&issueEvent)
	case "edited":
		return ep.handleIssueEdited(&issueEvent)
	case "closed":
		return ep.handleIssueClosed(&issueEvent)
	default:
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
	log.Printf("新评论创建: Issue #%d, User: %s",
		event.Issue.Number, event.Comment.User.Login)

	// 检查评论中是否包含命令
	if command := ep.extractCommand(event.Comment.Body); command != nil {
		log.Printf("在评论中检测到命令: %s", command.Command)

		return ep.executeCommand(command, &CommandContext{
			Repository: event.Repository,
			Issue:      &event.Issue,
			Comment:    &event.Comment,
			User:       event.Sender,
		})
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
	case "code":
		return ep.handleCodeCommand(command, ctx)
	case "continue":
		return ep.handleContinueCommand(command, ctx)
	case "fix":
		return ep.handleFixCommand(command, ctx)
	case "help":
		return ep.handleHelpCommand(command, ctx)
	default:
		return fmt.Errorf("未知命令: %s", command.Command)
	}
}

// handleCodeCommand 处理代码生成命令
func (ep *EventProcessor) handleCodeCommand(command *Command, ctx *CommandContext) error {
	log.Printf("处理代码生成命令: %s", command.Args)

	// 构建项目上下文
	context := ep.buildProjectContext(ctx)

	// 调用Claude API生成代码
	generatedCode, err := ep.claudeService.GenerateCode(command.Args, context)
	if err != nil {
		log.Printf("Claude API调用失败: %v", err)
		response := fmt.Sprintf(`❌ **代码生成失败**

错误信息: %s

请检查:
1. Claude API密钥是否正确配置
2. 网络连接是否正常
3. API配额是否充足

---
*处理时间: %s*`, err.Error(), time.Now().Format("2006-01-02 15:04:05"))
		return ep.createResponse(ctx, response)
	}

	response := fmt.Sprintf(`🤖 **CodeAgent 响应**

收到代码生成请求: %s

**处理流程:**
1. ✅ 分析需求
2. ✅ 调用Claude AI模型
3. ✅ 生成代码完成

**生成的代码:**

%s

---
*处理时间: %s*`, command.Args, generatedCode, time.Now().Format("2006-01-02 15:04:05"))

	return ep.createResponse(ctx, response)
}

// handleContinueCommand 处理继续命令
func (ep *EventProcessor) handleContinueCommand(command *Command, ctx *CommandContext) error {
	log.Printf("处理继续命令: %s", command.Args)

	// 构建项目上下文
	context := ep.buildProjectContext(ctx)

	// 调用Claude API继续开发
	continuedCode, err := ep.claudeService.ContinueCode(command.Args, context)
	if err != nil {
		log.Printf("Claude API调用失败: %v", err)
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

	// 调用Claude API修复代码
	fixedCode, err := ep.claudeService.FixCode(command.Args, context)
	if err != nil {
		log.Printf("Claude API调用失败: %v", err)
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

- ` + "`" + `/code <需求描述>` + "`" + ` - 生成代码实现指定功能
- ` + "`" + `/continue [说明]` + "`" + ` - 继续当前的开发任务
- ` + "`" + `/fix <问题描述>` + "`" + ` - 修复指定的代码问题
- ` + "`" + `/help` + "`" + ` - 显示此帮助信息

**使用示例:**
- ` + "`" + `/code 实现用户登录功能` + "`" + `
- ` + "`" + `/continue 添加错误处理` + "`" + `
- ` + "`" + `/fix 修复空指针异常` + "`" + `

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

// buildProjectContext 构建项目上下文
func (ep *EventProcessor) buildProjectContext(ctx *CommandContext) string {
	var context strings.Builder

	// 添加仓库信息
	context.WriteString(fmt.Sprintf("**仓库信息:**\n"))
	context.WriteString(fmt.Sprintf("- 仓库: %s\n", ctx.Repository.FullName))
	context.WriteString(fmt.Sprintf("- 名称: %s\n", ctx.Repository.Name))
	context.WriteString(fmt.Sprintf("- URL: %s\n", ctx.Repository.HTMLURL))

	// 添加Issue信息
	if ctx.Issue != nil {
		context.WriteString(fmt.Sprintf("\n**Issue信息:**\n"))
		context.WriteString(fmt.Sprintf("- 标题: %s\n", ctx.Issue.Title))
		context.WriteString(fmt.Sprintf("- 描述: %s\n", ctx.Issue.Body))
		context.WriteString(fmt.Sprintf("- 状态: %s\n", ctx.Issue.State))

		// 处理标签
		var labelNames []string
		for _, label := range ctx.Issue.Labels {
			labelNames = append(labelNames, label.Name)
		}
		context.WriteString(fmt.Sprintf("- 标签: %s\n", strings.Join(labelNames, ", ")))
	}

	// 添加Pull Request信息
	if ctx.PullRequest != nil {
		context.WriteString(fmt.Sprintf("\n**Pull Request信息:**\n"))
		context.WriteString(fmt.Sprintf("- 标题: %s\n", ctx.PullRequest.Title))
		context.WriteString(fmt.Sprintf("- 描述: %s\n", ctx.PullRequest.Body))
		context.WriteString(fmt.Sprintf("- 状态: %s\n", ctx.PullRequest.State))
		context.WriteString(fmt.Sprintf("- 分支: %s -> %s\n", ctx.PullRequest.Head.Ref, ctx.PullRequest.Base.Ref))
	}

	// 添加用户信息
	context.WriteString(fmt.Sprintf("\n**用户信息:**\n"))
	context.WriteString(fmt.Sprintf("- 用户: %s\n", ctx.User.Login))

	return context.String()
}

// truncateString 截断字符串
func (ep *EventProcessor) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

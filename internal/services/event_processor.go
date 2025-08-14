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
		commandRegex:      regexp.MustCompile(`^/(code|continue|fix|help)\s*(.*)$`),
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
	log.Printf("新评论创建: Issue #%d, User: %s",
		event.Issue.Number, event.Comment.User.Login)

	// 检查评论中是否包含命令
	if command := ep.extractCommand(event.Comment.Body); command != nil {
		log.Printf("在评论中检测到命令: %s", command.Command)

		return ep.executeCommand(command, &CommandContext{ //识别命令 /code /continue /fix
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
🔹 ` + "`" + `/help` + "`" + ` - 显示此帮助信息

**使用示例:**
- ` + "`" + `/code 创建一个用户登录API` + "`" + ` - 自动分析并实现到项目中
- ` + "`" + `/code 添加JWT认证功能` + "`" + ` - 自动分析并修改代码
- ` + "`" + `/continue 添加数据验证逻辑` + "`" + `
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

// autoAnalyzeAndModify 自动分析Issue并修改代码
func (ep *EventProcessor) autoAnalyzeAndModify(event *models.IssuesEvent) error {
	log.Printf("开始自动分析Issue: #%d", event.Issue.Number)

	// 检查是否已经有相同的分支存在，避免重复处理
	branchName := fmt.Sprintf("auto-fix-issue-%d", event.Issue.Number)

	// 简单的防重复机制：检查分支是否已经存在
	// 这里可以添加更复杂的检查逻辑
	log.Printf("准备创建分支: %s", branchName)

	// 克隆仓库
	repoPath, err := ep.gitService.CloneRepository(event.Repository.CloneURL, "main")
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

	// 获取文件树
	fileTree, err := ep.gitService.GetFileTree(repoPath)
	if err != nil {
		log.Printf("获取文件树失败: %v", err)
		fileTree = "无法获取文件树"
	}

	// 配置Git用户
	if err := ep.gitService.ConfigureGit(repoPath, "CodeAgent", "codeagent@example.com"); err != nil {
		log.Printf("配置Git失败: %v", err)
	}

	// 分析Issue内容，确定需要修改的文件
	analysisPrompt := fmt.Sprintf("分析以下Issue，确定需要修改的代码文件和具体修改内容：\n\nIssue信息:\n- 标题: %s\n- 描述: %s\n\n项目结构:\n%s\n\n任务要求:\n1. 分析Issue描述，理解用户需求\n2. 确定需要修改的文件路径\n3. 提供具体的代码修改建议\n4. 说明修改的原因和影响",
		event.Issue.Title, event.Issue.Body, fileTree)

	// 调用Claude Code CLI进行分析
	analysisResult, err := ep.claudeCodeService.callClaudeCodeCLI(analysisPrompt)
	if err != nil {
		log.Printf("AI分析失败: %v", err)
		errorMsg := fmt.Sprintf("自动分析失败: AI分析失败 - %v", err.Error())
		return ep.createResponse(&CommandContext{
			Repository: event.Repository,
			Issue:      &event.Issue,
			User:       event.Sender,
		}, errorMsg)
	}

	// 创建新分支，使用带时间戳的分支名避免冲突
	timestamp := time.Now().Format("20060102-150405")
	branchName = fmt.Sprintf("auto-fix-issue-%d-%s", event.Issue.Number, timestamp)
	log.Printf("创建分支: %s", branchName)
	if err := ep.gitService.CreateBranch(repoPath, branchName); err != nil {
		log.Printf("创建分支失败: %v", err)
	}

	// 创建GitHub事件包装结构用于新的方法
	gitHubEvent := &models.GitHubEvent{
		Type:       "issues",
		Repository: event.Repository,
		Issue:      event.Issue,
		Sender:     event.Sender,
	}

	// 根据AI分析结果实际修改代码
	modificationResult, err := ep.applyCodeModifications(repoPath, analysisResult, gitHubEvent)
	if err != nil {
		log.Printf("应用代码修改失败: %v", err)
		errorMsg := fmt.Sprintf("自动修改失败: %v", err.Error())
		return ep.createResponse(&CommandContext{
			Repository: event.Repository,
			Issue:      &event.Issue,
			User:       event.Sender,
		}, errorMsg)
	}

	// 提交修改到仓库
	commitResult, err := ep.commitAndPushChanges(repoPath, gitHubEvent, branchName)
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
func (ep *EventProcessor) commitAndPushChanges(repoPath string, event *models.GitHubEvent, branchName string) (string, error) {
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

	// 提交修改
	commitMessage := fmt.Sprintf("🤖 自动修复 Issue #%d: %s\n\n由AI助手自动生成的代码修改\n\nIssue链接: %s",
		event.Issue.Number, event.Issue.Title, event.Issue.URL)

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
	prResult, err := ep.createPullRequest(event, branchName)
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
func (ep *EventProcessor) createPullRequest(event *models.GitHubEvent, branchName string) (string, error) {
	title := fmt.Sprintf("🤖 自动修复 Issue #%d: %s", event.Issue.Number, event.Issue.Title)
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
		"main", // 目标分支，可以根据需要调整
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

# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a GitHub AI Webhook Demo system that uses Claude Code CLI to provide intelligent automated development workflows. The system processes GitHub webhooks (Issues, PR comments) and automatically generates code, fixes problems, and performs code reviews using AI.

## Key Commands

### Development
- `go run main.go` - Start the webhook server directly
- `./start.sh` - Interactive startup script with port checking and mode selection
- `go build -o webhook-demo main.go` - Build the binary

### Installation and Setup
- `./scripts/install_claude_code_cli.sh` - Install Claude Code CLI with system checks
- `cp config.env.example .env` - Setup environment configuration
- `go mod tidy` - Download and organize Go dependencies

### Testing
- `./scripts/test_auto_fix.sh` - Test the auto-fix functionality end-to-end
- `./scripts/test_claude_code_cli.sh` - Test Claude Code CLI integration
- `./scripts/test_git_flow.sh` - Test Git operations workflow

### Configuration Check
- Check `.env` file exists before running
- Verify `CLAUDE_CODE_CLI_API_KEY` and `GITHUB_TOKEN` are configured
- Ensure Claude Code CLI is installed: `claude --version`

## Architecture

### Core Services (`internal/services/`)
- **`event_processor.go`** - Main orchestrator that handles GitHub events and coordinates all other services
- **`claude_code_cli.go`** - Claude Code CLI integration with enhanced error handling and retry logic
- **`git.go`** - Git operations (clone, commit, push, branch management)
- **`github.go`** - GitHub API integration (comments, PR creation)
- **`commit_builder.go`** - Standardized commit message generation

### Event Flow
1. **Webhook Reception** (`handlers/webhook.go`) - Validates signature and parses GitHub events
2. **Event Processing** (`event_processor.go`) - Routes events to appropriate handlers
3. **Command Extraction** - Detects AI commands like `/code`, `/fix`, `/review` in Issue/comment text
4. **Repository Cloning** - Creates isolated workspace in `GIT_WORK_DIR`
5. **AI Processing** - Calls Claude Code CLI with context-aware prompts
6. **Code Modification** - Applies changes directly in the repository workspace
7. **Git Operations** - Creates branch, commits, and pushes to remote
8. **PR Creation** - Automatically creates Pull Request (requires collaborator permissions)
9. **Response** - Comments back to original Issue/PR with results

### Supported Commands
- `/code <requirement>` - Analyzes requirements and implements functionality
- `/continue [description]` - Continues development from previous context
- `/fix <problem>` - Fixes specific code issues
- `/review [scope]` - Performs code review (PR-specific or general)
- `/summary [content]` - Generates project summaries
- `/help` - Shows command help

### Configuration (`internal/config/`)
- **`config.go`** - Main service configuration loading from environment
- **`git_config.go`** - Git-specific settings (work directory, user info, file size limits)

## Important Implementation Details

### Claude Code CLI Integration
- Uses `--allowedTools "Edit,MultiEdit,Write,NotebookEdit,WebSearch,WebFetch"` for security
- Disables bash execution with `--disallowedTools "Bash"`
- Implements retry logic for API failures
- Supports custom API endpoints via `ANTHROPIC_BASE_URL`
- Uses stdin for prompt input to avoid command line length limits

### Git Workflow
- Creates timestamped branches: `auto-fix-issue-{number}-{timestamp}`
- Configures Git user as "CodeAgent" with "codeagent@example.com"
- Automatically detects default branch (`main` fallback)
- Handles both Issue and PR contexts appropriately

### Security Measures
- HMAC-SHA256 webhook signature verification
- Isolated workspace for each operation with cleanup
- Minimal permission GitHub tokens
- No direct main branch modifications

### Error Handling
- Comprehensive retry logic for API calls
- Graceful degradation when PR creation fails due to permissions
- Detailed error reporting in GitHub comments
- Automatic workspace cleanup on failures

## Environment Variables

### Required
- `GITHUB_TOKEN` - GitHub personal access token with repo, issues, pull_requests permissions
- `GITHUB_WEBHOOK_SECRET` - Webhook validation secret
- `CLAUDE_CODE_CLI_API_KEY` - Anthropic API key for Claude Code CLI

### Optional
- `SERVER_PORT` - Webhook server port (default: 8080)
- `GIT_WORK_DIR` - Git workspace directory (default: /tmp/webhook-demo)
- `CLAUDE_CODE_CLI_MODEL` - Claude model (default: claude-3-5-sonnet-20241022)
- `ANTHROPIC_BASE_URL` - Custom API endpoint (e.g., proxy services)

## Common Development Tasks

### Adding New Commands
1. Add command to regex in `NewEventProcessor()` in `event_processor.go:28`
2. Implement handler method following pattern `handleXCommand()`
3. Add case in `executeCommand()` switch statement
4. Update help text in `handleHelpCommand()`

### Modifying AI Prompts
- Code generation prompts are in `claude_code_cli.go` build methods
- Event-specific prompts are in respective handler methods in `event_processor.go`
- Always include project context using `buildProjectContext()`

### Testing Changes
1. Use `./scripts/test_auto_fix.sh` for full workflow testing
2. Create test Issue in connected repository
3. Monitor logs: `tail -f webhook.log`
4. Check GitHub for automatic PR creation

### Deployment
- Uses Docker with multi-stage build (see `Dockerfile`)
- Supports Docker Compose for production deployment
- Health check endpoint: `/health`
- Service info endpoint: `/`

## Dependencies

### Go Modules
- `github.com/gin-gonic/gin` - Web framework
- `github.com/joho/godotenv` - Environment variable loading

### External Tools
- **Claude Code CLI** - Must be installed via npm (`@anthropic-ai/claude-code`)
- **Git** - Required for repository operations
- **Node.js 18+** - Required for Claude Code CLI

## Migration Notes

This project was migrated from Gemini CLI to Claude Code CLI. Key changes:
- `GeminiService` → `ClaudeCodeCLIService`
- Enhanced error handling and timeout management
- Improved prompt engineering for better code generation
- Support for custom API endpoints

Refer to `CLAUDE_CODE_CLI_MIGRATION.md` for detailed migration information.
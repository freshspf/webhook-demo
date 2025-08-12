package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/webhook-demo/internal/config"
)

// ClaudeService Claude API服务
type ClaudeService struct {
	config *config.ClaudeConfig
	client *http.Client
}

// ClaudeMessage Claude消息结构
type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ClaudeRequest Claude API请求结构
type ClaudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []ClaudeMessage `json:"messages"`
}

// ClaudeResponse Claude API响应结构
type ClaudeResponse struct {
	Type  string `json:"type"`
	ID    string `json:"id"`
	Model string `json:"model"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

// NewClaudeService 创建新的Claude服务
func NewClaudeService(cfg *config.ClaudeConfig) *ClaudeService {
	return &ClaudeService{
		config: cfg,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GenerateCode 生成代码
func (cs *ClaudeService) GenerateCode(requirement string, context string) (string, error) {
	prompt := cs.buildCodeGenerationPrompt(requirement, context)
	return cs.callClaudeAPI(prompt)
}

// ContinueCode 继续代码开发
func (cs *ClaudeService) ContinueCode(instruction string, context string) (string, error) {
	prompt := cs.buildContinuePrompt(instruction, context)
	return cs.callClaudeAPI(prompt)
}

// FixCode 修复代码问题
func (cs *ClaudeService) FixCode(problem string, codeContext string) (string, error) {
	prompt := cs.buildFixPrompt(problem, codeContext)
	return cs.callClaudeAPI(prompt)
}

// callClaudeAPI 调用Claude API
func (cs *ClaudeService) callClaudeAPI(prompt string) (string, error) {
	if cs.config.APIKey == "" {
		return "", fmt.Errorf("Claude API密钥未配置")
	}

	requestBody := ClaudeRequest{
		Model:     cs.config.Model,
		MaxTokens: cs.config.MaxTokens,
		Messages: []ClaudeMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %v", err)
	}

	// req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	req, err := http.NewRequest("POST", "https://api2.aigcbest.top/v1/messages", bytes.NewBuffer(jsonData))

	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", cs.config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	log.Printf("调用Claude API，模型: %s, 最大Token: %d", cs.config.Model, cs.config.MaxTokens)

	resp, err := cs.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API调用失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Claude API错误响应: %s", string(body))
		return "", fmt.Errorf("API返回错误状态码: %d", resp.StatusCode)
	}

	var claudeResp ClaudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	if len(claudeResp.Content) == 0 {
		return "", fmt.Errorf("API响应中没有内容")
	}

	log.Printf("Claude API调用成功，输入Token: %d, 输出Token: %d",
		claudeResp.Usage.InputTokens, claudeResp.Usage.OutputTokens)

	return claudeResp.Content[0].Text, nil
}

// buildCodeGenerationPrompt 构建代码生成提示
func (cs *ClaudeService) buildCodeGenerationPrompt(requirement string, context string) string {
	return fmt.Sprintf("你是一个专业的软件开发助手，专门帮助用户生成高质量的代码。\n\n"+
		"**需求描述:**\n"+
		"%s\n\n"+
		"**项目上下文:**\n"+
		"%s\n\n"+
		"**要求:**\n"+
		"1. 生成完整、可运行的代码\n"+
		"2. 包含必要的注释和文档\n"+
		"3. 遵循最佳实践和代码规范\n"+
		"4. 考虑错误处理和边界情况\n"+
		"5. 如果涉及多个文件，请明确标注文件名\n\n"+
		"**输出格式:**\n"+
		"请直接输出代码，不需要额外的解释。如果需要多个文件，请使用 ```filename:path/to/file``` 格式标注。\n\n"+
		"请开始生成代码:", requirement, context)
}

// buildContinuePrompt 构建继续开发提示
func (cs *ClaudeService) buildContinuePrompt(instruction string, context string) string {
	return fmt.Sprintf("你正在继续一个软件开发项目。请根据以下指令继续开发：\n\n"+
		"**继续指令:**\n"+
		"%s\n\n"+
		"**当前项目上下文:**\n"+
		"%s\n\n"+
		"**要求:**\n"+
		"1. 基于现有代码继续开发\n"+
		"2. 保持代码风格的一致性\n"+
		"3. 确保新代码与现有代码兼容\n"+
		"4. 添加必要的注释说明\n\n"+
		"请继续开发:", instruction, context)
}

// buildFixPrompt 构建代码修复提示
func (cs *ClaudeService) buildFixPrompt(problem string, codeContext string) string {
	return fmt.Sprintf("你正在修复代码中的问题。请分析并修复以下问题：\n\n"+
		"**问题描述:**\n"+
		"%s\n\n"+
		"**代码上下文:**\n"+
		"%s\n\n"+
		"**要求:**\n"+
		"1. 分析问题的根本原因\n"+
		"2. 提供修复方案\n"+
		"3. 确保修复后的代码正确运行\n"+
		"4. 添加必要的注释说明修复内容\n\n"+
		"请修复代码:", problem, codeContext)
}

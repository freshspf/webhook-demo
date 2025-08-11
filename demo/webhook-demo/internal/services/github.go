package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// GitHubService GitHub API服务
type GitHubService struct {
	token   string
	client  *http.Client
	baseURL string
}

// NewGitHubService 创建新的GitHub服务
func NewGitHubService(token string) *GitHubService {
	return &GitHubService{
		token: token,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.github.com",
	}
}

// CreateComment 在Issue或PR上创建评论
func (s *GitHubService) CreateComment(owner, repo string, issueNumber int, body string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d/comments", s.baseURL, owner, repo, issueNumber)

	payload := map[string]string{
		"body": body,
	}

	return s.makeRequest("POST", url, payload, nil)
}

// UpdateComment 更新评论
func (s *GitHubService) UpdateComment(owner, repo string, commentID int64, body string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/comments/%d", s.baseURL, owner, repo, commentID)

	payload := map[string]string{
		"body": body,
	}

	return s.makeRequest("PATCH", url, payload, nil)
}

// CreatePullRequest 创建Pull Request
func (s *GitHubService) CreatePullRequest(owner, repo, title, body, head, base string) (*PullRequestResponse, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls", s.baseURL, owner, repo)

	payload := map[string]string{
		"title": title,
		"body":  body,
		"head":  head,
		"base":  base,
	}

	var response PullRequestResponse
	err := s.makeRequest("POST", url, payload, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// UpdatePullRequest 更新Pull Request
func (s *GitHubService) UpdatePullRequest(owner, repo string, number int, title, body string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d", s.baseURL, owner, repo, number)

	payload := map[string]string{
		"title": title,
		"body":  body,
	}

	return s.makeRequest("PATCH", url, payload, nil)
}

// GetIssue 获取Issue信息
func (s *GitHubService) GetIssue(owner, repo string, issueNumber int) (*IssueResponse, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d", s.baseURL, owner, repo, issueNumber)

	var response IssueResponse
	err := s.makeRequest("GET", url, nil, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// GetPullRequest 获取Pull Request信息
func (s *GitHubService) GetPullRequest(owner, repo string, number int) (*PullRequestResponse, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d", s.baseURL, owner, repo, number)

	var response PullRequestResponse
	err := s.makeRequest("GET", url, nil, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// GetRepository 获取仓库信息
func (s *GitHubService) GetRepository(owner, repo string) (*RepositoryResponse, error) {
	url := fmt.Sprintf("%s/repos/%s/%s", s.baseURL, owner, repo)

	var response RepositoryResponse
	err := s.makeRequest("GET", url, nil, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// makeRequest 发起HTTP请求
func (s *GitHubService) makeRequest(method, url string, payload interface{}, response interface{}) error {
	var body io.Reader

	if payload != nil {
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("序列化请求数据失败: %v", err)
		}
		body = bytes.NewBuffer(jsonPayload)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "Webhook-Demo/1.0")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// 设置认证
	if s.token != "" {
		req.Header.Set("Authorization", "token "+s.token)
	}

	// 发起请求
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	// 检查状态码
	if resp.StatusCode >= 400 {
		log.Printf("GitHub API错误: %d %s", resp.StatusCode, string(respBody))
		return fmt.Errorf("GitHub API错误: %d", resp.StatusCode)
	}

	// 解析响应
	if response != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, response); err != nil {
			return fmt.Errorf("解析响应失败: %v", err)
		}
	}

	return nil
}

// 响应结构体
type PullRequestResponse struct {
	ID       int64  `json:"id"`
	Number   int    `json:"number"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	State    string `json:"state"`
	HTMLURL  string `json:"html_url"`
	DiffURL  string `json:"diff_url"`
	PatchURL string `json:"patch_url"`
}

type IssueResponse struct {
	ID      int64  `json:"id"`
	Number  int    `json:"number"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	State   string `json:"state"`
	HTMLURL string `json:"html_url"`
}

type RepositoryResponse struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	HTMLURL  string `json:"html_url"`
	CloneURL string `json:"clone_url"`
	SSHURL   string `json:"ssh_url"`
}

package models

import (
	"encoding/json"
	"time"
)

// GitHubEvent GitHub事件结构
type GitHubEvent struct {
	Type       string     `json:"type"`
	DeliveryID string     `json:"delivery_id"`
	Payload    []byte     `json:"payload"`
	Timestamp  time.Time  `json:"timestamp"`
	Repository Repository `json:"repository"`
	Issue      Issue      `json:"issue"`
	Sender     User       `json:"sender"`
}

// Repository 仓库信息
type Repository struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	HTMLURL  string `json:"html_url"`
	CloneURL string `json:"clone_url"`
	SSHURL   string `json:"ssh_url"`
	Owner    User   `json:"owner"`
}

// User 用户信息
type User struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	HTMLURL   string `json:"html_url"`
	AvatarURL string `json:"avatar_url"`
}

// Issue Issue信息
type Issue struct {
	ID        int64     `json:"id"`
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	State     string    `json:"state"`
	HTMLURL   string    `json:"html_url"`
	URL       string    `json:"url"`
	User      User      `json:"user"`
	Assignee  *User     `json:"assignee"`
	Labels    []Label   `json:"labels"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PullRequest Pull Request信息
type PullRequest struct {
	ID        int64     `json:"id"`
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	State     string    `json:"state"`
	HTMLURL   string    `json:"html_url"`
	User      User      `json:"user"`
	Head      PRBranch  `json:"head"`
	Base      PRBranch  `json:"base"`
	Merged    bool      `json:"merged"`
	Draft     bool      `json:"draft"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PRBranch Pull Request分支信息
type PRBranch struct {
	Ref  string     `json:"ref"`
	SHA  string     `json:"sha"`
	Repo Repository `json:"repo"`
}

// Comment 评论信息
type Comment struct {
	ID        int64     `json:"id"`
	Body      string    `json:"body"`
	User      User      `json:"user"`
	HTMLURL   string    `json:"html_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Label 标签信息
type Label struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

// IssuesEvent Issue事件
type IssuesEvent struct {
	Action     string     `json:"action"`
	Issue      Issue      `json:"issue"`
	Repository Repository `json:"repository"`
	Sender     User       `json:"sender"`
}

// IssueCommentEvent Issue评论事件
type IssueCommentEvent struct {
	Action     string     `json:"action"`
	Issue      Issue      `json:"issue"`
	Comment    Comment    `json:"comment"`
	Repository Repository `json:"repository"`
	Sender     User       `json:"sender"`
}

// PullRequestEvent Pull Request事件
type PullRequestEvent struct {
	Action      string      `json:"action"`
	Number      int         `json:"number"`
	PullRequest PullRequest `json:"pull_request"`
	Repository  Repository  `json:"repository"`
	Sender      User        `json:"sender"`
}

// PullRequestReviewCommentEvent PR Review评论事件
type PullRequestReviewCommentEvent struct {
	Action      string      `json:"action"`
	PullRequest PullRequest `json:"pull_request"`
	Comment     Comment     `json:"comment"`
	Repository  Repository  `json:"repository"`
	Sender      User        `json:"sender"`
}

// ParsePayload 解析payload为指定的事件类型
func (e *GitHubEvent) ParsePayload(v interface{}) error {
	return json.Unmarshal(e.Payload, v)
}

// GetPayloadAsString 获取payload的字符串表示
func (e *GitHubEvent) GetPayloadAsString() string {
	return string(e.Payload)
}

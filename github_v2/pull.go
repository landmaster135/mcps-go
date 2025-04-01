package github_v2

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	mcp "github.com/mark3labs/mcp-go/mcp"
	server "github.com/mark3labs/mcp-go/server"
)

// CreatePullRequest は新しいプルリクエストを作成します
func (c *GitHubClient) CreatePullRequest(owner, repo string, options map[string]interface{}) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls", apiBaseURL, owner, repo)

	jsonBody, err := json.Marshal(options)
	if err != nil {
		return nil, err
	}

	data, err := c.doRequest("POST", url, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// HandleToCreatePullRequest は新しいプルリクエストを作成して、結果をJSON形式で返します
func (c *GitHubClient) HandleToCreatePullRequest(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	owner := getRequiredStringParam(request.Params.Arguments, "owner")
	repo := getRequiredStringParam(request.Params.Arguments, "repo")

	options := make(map[string]interface{})
	options["title"] = getRequiredStringParam(request.Params.Arguments, "title")
	options["head"] = getRequiredStringParam(request.Params.Arguments, "head")
	options["base"] = getRequiredStringParam(request.Params.Arguments, "base")

	// オプションパラメータを追加
	if body, ok := getStringParam(request.Params.Arguments, "body"); ok {
		options["body"] = body
	}

	options["draft"] = getBoolParam(request.Params.Arguments, "draft", false)

	result, err := c.CreatePullRequest(owner, repo, options)
	if err != nil {
		return nil, err
	}

	return returnJSONResult(result)
}

// CreatePullRequestReview はプルリクエストにレビューを作成します
func (c *GitHubClient) CreatePullRequestReview(owner, repo string, pullNumber int, options map[string]interface{}) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d/reviews", apiBaseURL, owner, repo, pullNumber)

	jsonBody, err := json.Marshal(options)
	if err != nil {
		return nil, err
	}

	data, err := c.doRequest("POST", url, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// HandleToCreatePullRequestReview はプルリクエストにレビューを作成して、結果をJSON形式で返します
func (c *GitHubClient) HandleToCreatePullRequestReview(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	owner := getRequiredStringParam(request.Params.Arguments, "owner")
	repo := getRequiredStringParam(request.Params.Arguments, "repo")
	pullNumber := int(request.Params.Arguments["pull_number"].(float64))

	options := make(map[string]interface{})

	// 文字列オプションパラメータを追加
	if event, ok := getStringParam(request.Params.Arguments, "event"); ok {
		options["event"] = event
	}
	if body, ok := getStringParam(request.Params.Arguments, "body"); ok {
		options["body"] = body
	}

	result, err := c.CreatePullRequestReview(owner, repo, pullNumber, options)
	if err != nil {
		return nil, err
	}

	return returnJSONResult(result)
}

// MergePullRequest はプルリクエストをマージします
func (c *GitHubClient) MergePullRequest(owner, repo string, pullNumber int, options map[string]interface{}) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d/merge", apiBaseURL, owner, repo, pullNumber)

	jsonBody, err := json.Marshal(options)
	if err != nil {
		return nil, err
	}

	data, err := c.doRequest("PUT", url, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// HandleToMergePullRequest はプルリクエストをマージして、結果をJSON形式で返します
func (c *GitHubClient) HandleToMergePullRequest(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	owner := getRequiredStringParam(request.Params.Arguments, "owner")
	repo := getRequiredStringParam(request.Params.Arguments, "repo")
	pullNumber := int(request.Params.Arguments["pull_number"].(float64))

	options := make(map[string]interface{})

	if commitTitle, ok := getStringParam(request.Params.Arguments, "commit_title"); ok {
		options["commit_title"] = commitTitle
	}
	if commitMessage, ok := getStringParam(request.Params.Arguments, "commit_message"); ok {
		options["commit_message"] = commitMessage
	}
	if mergeMethod, ok := getStringParam(request.Params.Arguments, "merge_method"); ok {
		options["merge_method"] = mergeMethod
	}

	result, err := c.MergePullRequest(owner, repo, pullNumber, options)
	if err != nil {
		return nil, err
	}

	return returnJSONResult(result)
}

// GetPullRequestFiles はプルリクエストで変更されたファイル一覧を取得します
func (c *GitHubClient) GetPullRequestFiles(owner, repo string, pullNumber int) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d/files", apiBaseURL, owner, repo, pullNumber)

	data, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// HandleToGetPullRequestFiles はプルリクエストで変更されたファイル一覧を取得して、結果をJSON形式で返します
func (c *GitHubClient) HandleToGetPullRequestFiles(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	owner := getRequiredStringParam(request.Params.Arguments, "owner")
	repo := getRequiredStringParam(request.Params.Arguments, "repo")
	pullNumber := int(request.Params.Arguments["pull_number"].(float64))

	result, err := c.GetPullRequestFiles(owner, repo, pullNumber)
	if err != nil {
		return nil, err
	}

	return returnJSONResult(result)
}

// SetGitHubPullRequestServer は受け取ったMCPサーバにGitHubプルリクエスト用のツールを付与して、そのMCPサーバを返します。
func SetGitHubPullRequestServer(token string, s *server.MCPServer) *server.MCPServer {
	// GitHubクライアントを初期化
	client := NewGitHubClient(token)

	// ツール1: プルリクエストの作成
	createPullRequestTool := mcp.NewTool("create_pull_request",
		mcp.WithDescription("Create a new pull request in a GitHub repository"),
		mcp.WithString("owner",
			mcp.Required(),
			mcp.Description("Repository owner"),
		),
		mcp.WithString("repo",
			mcp.Required(),
			mcp.Description("Repository name"),
		),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("Pull request title"),
		),
		mcp.WithString("head",
			mcp.Required(),
			mcp.Description("The name of the branch where your changes are implemented"),
		),
		mcp.WithString("base",
			mcp.Required(),
			mcp.Description("The name of the branch you want the changes pulled into"),
		),
		mcp.WithString("body",
			mcp.Description("Pull request body"),
		),
		mcp.WithBoolean("draft",
			mcp.Description("Whether to create a draft pull request"),
		),
	)
	s.AddTool(createPullRequestTool, client.HandleToCreatePullRequest)

	// ツール2: プルリクエストレビューの作成
	createPullRequestReviewTool := mcp.NewTool("create_pull_request_review",
		mcp.WithDescription("Create a review on a pull request"),
		mcp.WithString("owner",
			mcp.Required(),
			mcp.Description("Repository owner"),
		),
		mcp.WithString("repo",
			mcp.Required(),
			mcp.Description("Repository name"),
		),
		mcp.WithNumber("pull_number",
			mcp.Required(),
			mcp.Description("Pull request number"),
		),
		mcp.WithString("event",
			mcp.Description("Review event: APPROVE, REQUEST_CHANGES, COMMENT"),
			mcp.Enum("APPROVE", "REQUEST_CHANGES", "COMMENT"),
		),
		mcp.WithString("body",
			mcp.Description("Review body"),
		),
	)
	s.AddTool(createPullRequestReviewTool, client.HandleToCreatePullRequestReview)

	// ツール3: プルリクエストのマージ
	mergePullRequestTool := mcp.NewTool("merge_pull_request",
		mcp.WithDescription("Merge a pull request"),
		mcp.WithString("owner",
			mcp.Required(),
			mcp.Description("Repository owner"),
		),
		mcp.WithString("repo",
			mcp.Required(),
			mcp.Description("Repository name"),
		),
		mcp.WithNumber("pull_number",
			mcp.Required(),
			mcp.Description("Pull request number"),
		),
		mcp.WithString("commit_title",
			mcp.Description("Title for the automatic commit message"),
		),
		mcp.WithString("commit_message",
			mcp.Description("Extra detail to append to automatic commit message"),
		),
		mcp.WithString("merge_method",
			mcp.Description("Merge method to use: merge, squash, rebase"),
			mcp.Enum("merge", "squash", "rebase"),
		),
	)
	s.AddTool(mergePullRequestTool, client.HandleToMergePullRequest)

	// ツール4: プルリクエストのファイル一覧取得
	getPullRequestFilesTool := mcp.NewTool("get_pull_request_files",
		mcp.WithDescription("Get the list of files changed in a pull request"),
		mcp.WithString("owner",
			mcp.Required(),
			mcp.Description("Repository owner"),
		),
		mcp.WithString("repo",
			mcp.Required(),
			mcp.Description("Repository name"),
		),
		mcp.WithNumber("pull_number",
			mcp.Required(),
			mcp.Description("Pull request number"),
		),
	)
	s.AddTool(getPullRequestFilesTool, client.HandleToGetPullRequestFiles)

	return s
}

package github_v2

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	mcp "github.com/mark3labs/mcp-go/mcp"
	server "github.com/mark3labs/mcp-go/server"
)

// CreateIssue は新しいイシューを作成します
func (c *GitHubClient) CreateIssue(owner, repo string, options map[string]interface{}) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues", apiBaseURL, owner, repo)
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

// HandleToCreateIssue は新しいイシューを作成して、結果をJSON形式で返します
func (c *GitHubClient) HandleToCreateIssue(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	owner := getRequiredStringParam(request.Params.Arguments, "owner")
	repo := getRequiredStringParam(request.Params.Arguments, "repo")

	options := make(map[string]interface{})
	options["title"] = getRequiredStringParam(request.Params.Arguments, "title")

	// オプションパラメータを追加
	if body, ok := getStringParam(request.Params.Arguments, "body"); ok {
		options["body"] = body
	}

	// 配列パラメータを追加
	addToOptions(options, request.Params.Arguments, "labels")
	addToOptions(options, request.Params.Arguments, "assignees")

	result, err := c.CreateIssue(owner, repo, options)
	if err != nil {
		return nil, err
	}

	return returnJSONResult(result)
}

// ListIssues はリポジトリのイシュー一覧を取得します
func (c *GitHubClient) ListIssues(owner, repo string, options map[string]interface{}) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues", apiBaseURL, owner, repo)

	// クエリパラメータを追加
	queryParams := []string{}
	for k, v := range options {
		queryParams = append(queryParams, fmt.Sprintf("%s=%v", k, v))
	}
	if len(queryParams) > 0 {
		url += "?" + strings.Join(queryParams, "&")
	}

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

func (c *GitHubClient) HandleToListIssues(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	owner := getRequiredStringParam(request.Params.Arguments, "owner")
	repo := getRequiredStringParam(request.Params.Arguments, "repo")

	options := make(map[string]interface{})

	// 文字列オプションパラメータを追加
	if state, ok := getStringParam(request.Params.Arguments, "state"); ok {
		options["state"] = state
	}
	if sort, ok := getStringParam(request.Params.Arguments, "sort"); ok {
		options["sort"] = sort
	}
	if direction, ok := getStringParam(request.Params.Arguments, "direction"); ok {
		options["direction"] = direction
	}

	// 数値オプションパラメータを追加
	if perPage, ok := request.Params.Arguments["per_page"]; ok {
		options["per_page"] = int(perPage.(float64))
	}
	if page, ok := request.Params.Arguments["page"]; ok {
		options["page"] = int(page.(float64))
	}

	result, err := c.ListIssues(owner, repo, options)
	if err != nil {
		return nil, err
	}

	return returnJSONResult(result)
}

func SetGitHubIssueServer(token string, s *server.MCPServer) *server.MCPServer {
	// GitHubクライアントを初期化
	client := NewGitHubClient(token)

	// ツール3: イシューの作成
	createIssueTool := mcp.NewTool("create_issue",
		mcp.WithDescription("Create a new issue in a GitHub repository"),
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
			mcp.Description("Issue title"),
		),
		mcp.WithString("body",
			mcp.Description("Issue body"),
		),
		mcp.WithArray("labels",
			mcp.Description("Issue labels"),
		),
		mcp.WithArray("assignees",
			mcp.Description("Users to assign to this issue"),
		),
	)
	s.AddTool(createIssueTool, client.HandleToCreateIssue)

	// ツール4: イシュー一覧の取得
	listIssuesTool := mcp.NewTool("list_issues",
		mcp.WithDescription("List issues in a GitHub repository"),
		mcp.WithString("owner",
			mcp.Required(),
			mcp.Description("Repository owner"),
		),
		mcp.WithString("repo",
			mcp.Required(),
			mcp.Description("Repository name"),
		),
		mcp.WithString("state",
			mcp.Description("Issue state: open, closed, or all (default: open)"),
			mcp.Enum("open", "closed", "all"),
		),
		mcp.WithString("sort",
			mcp.Description("Sort field: created, updated, or comments (default: created)"),
			mcp.Enum("created", "updated", "comments"),
		),
		mcp.WithString("direction",
			mcp.Description("Sort direction: asc or desc (default: desc)"),
			mcp.Enum("asc", "desc"),
		),
		mcp.WithNumber("per_page",
			mcp.Description("Results per page (default: 30, max: 100)"),
		),
		mcp.WithNumber("page",
			mcp.Description("Page number (default: 1)"),
		),
	)
	s.AddTool(listIssuesTool, client.HandleToListIssues)

	return s
}

package github_v2

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
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

	s.AddTool(createIssueTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

		result, err := client.CreateIssue(owner, repo, options)
		if err != nil {
			return nil, err
		}

		return returnJSONResult(result)
	})

	return s
}

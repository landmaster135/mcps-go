package github

import (
	"context"
	"encoding/json"
	"fmt"

	mcp "github.com/mark3labs/mcp-go/mcp"
	server "github.com/mark3labs/mcp-go/server"
)

// SearchCode はGitHub全体でコードを検索します
func (c *GitHubClient) SearchCode(query string, options map[string]interface{}) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/search/code?q=%s", apiBaseURL, query)

	// クエリパラメータを追加
	for k, v := range options {
		url += fmt.Sprintf("&%s=%v", k, v)
	}

	data, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// HandleToSearchCode はGitHub全体でコードを検索して、結果をJSON形式で返します
func (c *GitHubClient) HandleToSearchCode(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := getRequiredStringParam(request.Params.Arguments, "query")

	options := make(map[string]interface{})

	// 数値オプションパラメータを追加
	if page, ok := request.Params.Arguments["page"]; ok {
		options["page"] = int(page.(float64))
	}
	if perPage, ok := request.Params.Arguments["per_page"]; ok {
		options["per_page"] = int(perPage.(float64))
	}

	result, err := c.SearchCode(query, options)
	if err != nil {
		return nil, err
	}

	return returnJSONResult(result)
}

// SetGitHubGlobalServer は受け取ったMCPサーバにGitHubグローバル検索用のツールを付与して、そのMCPサーバを返します。
func SetGitHubGlobalServer(token string, s *server.MCPServer) *server.MCPServer {
	// GitHubクライアントを初期化
	client := NewGitHubClient(token)

	// ツール1: コード検索
	searchCodeTool := mcp.NewTool("search_code",
		mcp.WithDescription("Search for code across GitHub repositories"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query. For example: 'addClass in:file language:js repo:jquery/jquery'"),
		),
		mcp.WithNumber("page",
			mcp.Description("Page number (default: 1)"),
		),
		mcp.WithNumber("per_page",
			mcp.Description("Results per page (default: 30, max: 100)"),
		),
	)
	s.AddTool(searchCodeTool, client.HandleToSearchCode)

	return s
}

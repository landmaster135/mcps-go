package github_v2

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	mcp "github.com/mark3labs/mcp-go/mcp"
	server "github.com/mark3labs/mcp-go/server"
)

// ListCommits はリポジトリのコミット一覧を取得します
func (c *GitHubClient) ListCommits(owner, repo string, page, perPage int, sha string) ([]map[string]interface{}, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 30
	}

	url := fmt.Sprintf("%s/repos/%s/%s/commits?page=%d&per_page=%d", apiBaseURL, owner, repo, page, perPage)
	if sha != "" {
		url += fmt.Sprintf("&sha=%s", sha)
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

// HandleToListCommits はリポジトリのコミット一覧を取得して、結果をJSON形式で返します
func (c *GitHubClient) HandleToListCommits(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	owner := getRequiredStringParam(request.Params.Arguments, "owner")
	repo := getRequiredStringParam(request.Params.Arguments, "repo")
	page := getNumberParam(request.Params.Arguments, "page", 1)
	perPage := getNumberParam(request.Params.Arguments, "per_page", 30)
	sha, _ := getStringParam(request.Params.Arguments, "sha")

	result, err := c.ListCommits(owner, repo, page, perPage, sha)
	if err != nil {
		return nil, err
	}

	return returnJSONResult(result)
}

// SearchRepositories はGitHubリポジトリを検索します
func (c *GitHubClient) SearchRepositories(query string, page, perPage int) (map[string]interface{}, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 30
	}

	url := fmt.Sprintf("%s/search/repositories?q=%s&page=%d&per_page=%d", apiBaseURL, query, page, perPage)
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

// HandleToSearchRepositories はGitHubリポジトリを検索して、結果をJSON形式で返します
func (c *GitHubClient) HandleToSearchRepositories(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := getRequiredStringParam(request.Params.Arguments, "query")
	page := getNumberParam(request.Params.Arguments, "page", 1)
	perPage := getNumberParam(request.Params.Arguments, "per_page", 30)

	result, err := c.SearchRepositories(query, page, perPage)
	if err != nil {
		return nil, err
	}

	return returnJSONResult(result)
}

// GetUserRepositories はユーザーのリポジトリ一覧を取得します
func (c *GitHubClient) GetUserRepositories(username string, options map[string]interface{}) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/users/%s/repos", apiBaseURL, username)

	// クエリパラメータを追加
	queryParams := []string{}
	for k, v := range options {
		queryParams = append(queryParams, fmt.Sprintf("%s=%v", k, v))
	}
	if len(queryParams) > 0 {
		url += "?=" + strings.Join(queryParams, "&")
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

// HandleToGetUserRepositories はユーザーのリポジトリ一覧を取得して、結果をJSON形式で返します
func (c *GitHubClient) HandleToGetUserRepositories(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	username := getRequiredStringParam(request.Params.Arguments, "username")

	options := make(map[string]interface{})

	// 数値オプションパラメータを追加
	if perPage, ok := request.Params.Arguments["per_page"]; ok {
		options["per_page"] = int(perPage.(float64))
	}
	if page, ok := request.Params.Arguments["page"]; ok {
		options["page"] = int(page.(float64))
	}

	// 文字列オプションパラメータを追加
	if sort, ok := getStringParam(request.Params.Arguments, "sort"); ok {
		options["sort"] = sort
	}
	if direction, ok := getStringParam(request.Params.Arguments, "direction"); ok {
		options["direction"] = direction
	}
	if type_, ok := getStringParam(request.Params.Arguments, "type"); ok {
		options["type"] = type_
	}

	result, err := c.GetUserRepositories(username, options)
	if err != nil {
		return nil, err
	}

	return returnJSONResult(result)
}

// GetFileContents はリポジトリからファイルの内容を取得します
func (c *GitHubClient) GetFileContents(owner, repo, path, branch string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s", apiBaseURL, owner, repo, path)
	if branch != "" {
		url += fmt.Sprintf("?ref=%s", branch)
	}

	data, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	// ファイルの内容をデコードする
	if content, ok := result["content"].(string); ok {
		decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(content, "\n", ""))
		if err != nil {
			return nil, err
		}
		result["decoded_content"] = string(decoded)
	}

	return result, nil
}

// HandleToGetFileContents はリポジトリからファイルの内容を取得して、結果をJSON形式で返します
func (c *GitHubClient) HandleToGetFileContents(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	owner := getRequiredStringParam(request.Params.Arguments, "owner")
	repo := getRequiredStringParam(request.Params.Arguments, "repo")
	path := getRequiredStringParam(request.Params.Arguments, "path")
	branch, _ := getStringParam(request.Params.Arguments, "branch")

	result, err := c.GetFileContents(owner, repo, path, branch)
	if err != nil {
		return nil, err
	}

	return returnJSONResult(result)
}

// SetGitHubRepositoryServer は受け取ったMCPサーバにGitHubリポジトリ用のツールを付与して、そのMCPサーバを返します。
func SetGitHubRepositoryServer(token string, s *server.MCPServer) *server.MCPServer {
	// GitHubクライアントを初期化
	client := NewGitHubClient(token)

	// ツール1: リポジトリ検索
	searchRepositoriesTool := mcp.NewTool("search_repositories",
		mcp.WithDescription("Search for GitHub repositories"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query"),
		),
		mcp.WithNumber("page",
			mcp.Description("Page number (default: 1)"),
		),
		mcp.WithNumber("per_page",
			mcp.Description("Results per page (default: 30, max: 100)"),
		),
	)
	s.AddTool(searchRepositoriesTool, client.HandleToSearchRepositories)

	// ツール2: ユーザーリポジトリの検索
	getUserRepositoriesTool := mcp.NewTool("get_user_repositories",
		mcp.WithDescription("Get repositories for a specific GitHub user"),
		mcp.WithString("username",
			mcp.Required(),
			mcp.Description("GitHub username"),
		),
		mcp.WithNumber("per_page",
			mcp.Description("Results per page (default: 30, max: 100)"),
		),
		mcp.WithNumber("page",
			mcp.Description("Page number (default: 1)"),
		),
		mcp.WithString("sort",
			mcp.Description("Sort field: created, updated, pushed, full_name (default: full_name)"),
			mcp.Enum("created", "updated", "pushed", "full_name"),
		),
		mcp.WithString("direction",
			mcp.Description("Sort direction: asc or desc (default: desc)"),
			mcp.Enum("asc", "desc"),
		),
		mcp.WithString("type",
			mcp.Description("Type of repositories to include: all, owner, member, public, private (default: all)"),
			mcp.Enum("all", "owner", "member", "public", "private"),
		),
	)
	s.AddTool(getUserRepositoriesTool, client.HandleToGetUserRepositories)

	// ツール3: ファイル内容の取得
	getFileContentsTool := mcp.NewTool("get_file_contents",
		mcp.WithDescription("Get the contents of a file from a GitHub repository"),
		mcp.WithString("owner",
			mcp.Required(),
			mcp.Description("Repository owner"),
		),
		mcp.WithString("repo",
			mcp.Required(),
			mcp.Description("Repository name"),
		),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("File path within the repository"),
		),
		mcp.WithString("branch",
			mcp.Description("Branch name (default: repository's default branch)"),
		),
	)
	s.AddTool(getFileContentsTool, client.HandleToGetFileContents)

	// ツール4: コミット一覧の取得
	listCommitsTool := mcp.NewTool("list_commits",
		mcp.WithDescription("Get list of commits of a branch in a GitHub repository"),
		mcp.WithString("owner",
			mcp.Required(),
			mcp.Description("Repository owner"),
		),
		mcp.WithString("repo",
			mcp.Required(),
			mcp.Description("Repository name"),
		),
		mcp.WithNumber("page",
			mcp.Description("Page number (default: 1)"),
		),
		mcp.WithNumber("per_page",
			mcp.Description("Results per page (default: 30, max: 100)"),
		),
		mcp.WithString("sha",
			mcp.Description("SHA or branch name to start listing commits from"),
		),
	)
	s.AddTool(listCommitsTool, client.HandleToListCommits)

	return s
}

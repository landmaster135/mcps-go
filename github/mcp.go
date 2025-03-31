package github

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	mcp "github.com/mark3labs/mcp-go/mcp"
	server "github.com/mark3labs/mcp-go/server"
)

// ヘルパー関数: 文字列パラメータを取得
func getStringParam(args map[string]interface{}, key string) (string, bool) {
	if val, ok := args[key]; ok {
		return val.(string), true
	}
	return "", false
}

// ヘルパー関数: 必須の文字列パラメータを取得
func getRequiredStringParam(args map[string]interface{}, key string) string {
	return args[key].(string)
}

// ヘルパー関数: 数値パラメータを取得
func getNumberParam(args map[string]interface{}, key string, defaultVal int) int {
	if val, ok := args[key]; ok {
		return int(val.(float64))
	}
	return defaultVal
}

// ヘルパー関数: ブールパラメータを取得
func getBoolParam(args map[string]interface{}, key string, defaultVal bool) bool {
	if val, ok := args[key]; ok {
		return val.(bool)
	}
	return defaultVal
}

// ヘルパー関数: 結果をJSON形式で返却
func returnJSONResult(result interface{}) (*mcp.CallToolResult, error) {
	jsonResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(string(jsonResult)), nil
}

// ヘルパー関数: オプションマップにパラメータを追加
func addToOptions(options map[string]interface{}, args map[string]interface{}, key string) {
	if val, ok := args[key]; ok {
		options[key] = val
	}
}

const (
	apiBaseURL = "https://api.github.com"
	version    = "1.0.0"
)

// GitHubError はGitHub APIからのエラーを表します
type GitHubError struct {
	Message          string `json:"message"`
	DocumentationURL string `json:"documentation_url"`
	StatusCode       int
}

func (e *GitHubError) Error() string {
	return fmt.Sprintf("GitHub API Error: %s (Status: %d)", e.Message, e.StatusCode)
}

// GitHubClient はGitHub APIとの通信を処理します
type GitHubClient struct {
	httpClient *http.Client
	token      string
}

// NewGitHubClient は新しいGitHubクライアントを作成します
func NewGitHubClient(token string) *GitHubClient {
	return &GitHubClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		token: token,
	}
}

// createRequest はHTTPリクエストを作成して、それにヘッダーおよびボディを設定します
func (c *GitHubClient) createRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if c.token != "" {
		req.Header.Set("Authorization", "token "+c.token)
	}
	if method == "POST" || method == "PATCH" || method == "PUT" {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

// doRequest はHTTPリクエストを実行し、レスポンスを処理します
func (c *GitHubClient) doRequest(method, url string, body io.Reader) ([]byte, error) {
	req, err := c.createRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		var ghError GitHubError
		if err := json.Unmarshal(respBody, &ghError); err != nil {
			return nil, fmt.Errorf("HTTP error: %d - %s", resp.StatusCode, string(respBody))
		}
		ghError.StatusCode = resp.StatusCode
		return nil, &ghError
	}

	return respBody, nil
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

// UpdateIssue は既存のイシューを更新します
func (c *GitHubClient) UpdateIssue(owner, repo string, issueNumber int, options map[string]interface{}) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d", apiBaseURL, owner, repo, issueNumber)

	jsonBody, err := json.Marshal(options)
	if err != nil {
		return nil, err
	}

	data, err := c.doRequest("PATCH", url, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// AddIssueComment はイシューにコメントを追加します
func (c *GitHubClient) AddIssueComment(owner, repo string, issueNumber int, body string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d/comments", apiBaseURL, owner, repo, issueNumber)

	jsonBody, err := json.Marshal(map[string]string{"body": body})
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

// GetPullRequestStatus はプルリクエストのステータスを取得します
func (c *GitHubClient) GetPullRequestStatus(owner, repo string, pullNumber int) (map[string]interface{}, error) {
	// プルリクエストの詳細を取得
	prURL := fmt.Sprintf("%s/repos/%s/%s/pulls/%d", apiBaseURL, owner, repo, pullNumber)
	prData, err := c.doRequest("GET", prURL, nil)
	if err != nil {
		return nil, err
	}

	var pr map[string]interface{}
	if err := json.Unmarshal(prData, &pr); err != nil {
		return nil, err
	}

	// ステータスチェックを取得
	headSHA, ok := pr["head"].(map[string]interface{})["sha"].(string)
	if !ok {
		return nil, fmt.Errorf("could not get head SHA from pull request")
	}

	statusURL := fmt.Sprintf("%s/repos/%s/%s/commits/%s/status", apiBaseURL, owner, repo, headSHA)
	statusData, err := c.doRequest("GET", statusURL, nil)
	if err != nil {
		return nil, err
	}

	var status map[string]interface{}
	if err := json.Unmarshal(statusData, &status); err != nil {
		return nil, err
	}

	// 結果を組み合わせる
	result := map[string]interface{}{
		"pull_request": pr,
		"status":       status,
	}

	return result, nil
}

// UpdatePullRequestBranch はプルリクエストのブランチを更新します
func (c *GitHubClient) UpdatePullRequestBranch(owner, repo string, pullNumber int, expectedHeadSHA string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d/update-branch", apiBaseURL, owner, repo, pullNumber)

	options := map[string]interface{}{}
	if expectedHeadSHA != "" {
		options["expected_head_sha"] = expectedHeadSHA
	}

	jsonBody, err := json.Marshal(options)
	if err != nil {
		return err
	}

	_, err = c.doRequest("PUT", url, strings.NewReader(string(jsonBody)))
	return err
}

// GetPullRequestComments はプルリクエストのコメントを取得します
func (c *GitHubClient) GetPullRequestComments(owner, repo string, pullNumber int) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d/comments", apiBaseURL, owner, repo, pullNumber)

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

// GetPullRequestReviews はプルリクエストのレビューを取得します
func (c *GitHubClient) GetPullRequestReviews(owner, repo string, pullNumber int) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d/reviews", apiBaseURL, owner, repo, pullNumber)

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

// ListPullRequests はリポジトリのプルリクエスト一覧を取得します
func (c *GitHubClient) ListPullRequests(owner, repo string, options map[string]interface{}) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls", apiBaseURL, owner, repo)

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

// GetUserRepositories はユーザーのリポジトリ一覧を取得します
func (c *GitHubClient) GetUserRepositories(username string, options map[string]interface{}) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/users/%s/repos", apiBaseURL, username)

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

// BuildGitHubServer はGitHubのMCPサーバーを構築します
func BuildGitHubServer() {
	// 環境変数からGitHubトークンを取得
	token := os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN")
	if token == "" {
		fmt.Println("Warning: GITHUB_PERSONAL_ACCESS_TOKEN environment variable not set. API rate limits will be restricted.")
	}

	// GitHubクライアントを初期化
	client := NewGitHubClient(token)

	// MCPサーバーを作成
	s := server.NewMCPServer(
		"GitHub API Server",
		version,
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
	)

	// ツール1: リポジトリ検索
	searchReposTool := mcp.NewTool("search_repositories",
		mcp.WithDescription("Search for GitHub repositories"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query"),
		),
		mcp.WithNumber("page",
			mcp.Description("Page number (default: 1)"),
		),
		mcp.WithNumber("perPage",
			mcp.Description("Results per page (default: 30, max: 100)"),
		),
	)

	s.AddTool(searchReposTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query := getRequiredStringParam(request.Params.Arguments, "query")
		page := getNumberParam(request.Params.Arguments, "page", 1)
		perPage := getNumberParam(request.Params.Arguments, "perPage", 30)

		result, err := client.SearchRepositories(query, page, perPage)
		if err != nil {
			return nil, err
		}

		return returnJSONResult(result)
	})

	// ツール2: ファイル内容の取得
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

	s.AddTool(getFileContentsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := getRequiredStringParam(request.Params.Arguments, "owner")
		repo := getRequiredStringParam(request.Params.Arguments, "repo")
		path := getRequiredStringParam(request.Params.Arguments, "path")
		branch, _ := getStringParam(request.Params.Arguments, "branch")

		result, err := client.GetFileContents(owner, repo, path, branch)
		if err != nil {
			return nil, err
		}

		return returnJSONResult(result)
	})

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

	s.AddTool(listIssuesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

		result, err := client.ListIssues(owner, repo, options)
		if err != nil {
			return nil, err
		}

		return returnJSONResult(result)
	})

	// ツール5: ユーザーリポジトリの検索
	searchUserReposTool := mcp.NewTool("search_user_repositories",
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
	)

	s.AddTool(searchUserReposTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

		result, err := client.GetUserRepositories(username, options)
		if err != nil {
			return nil, err
		}

		return returnJSONResult(result)
	})

	// ツール6: プルリクエストの作成
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

	s.AddTool(createPullRequestTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := getRequiredStringParam(request.Params.Arguments, "owner")
		repo := getRequiredStringParam(request.Params.Arguments, "repo")

		options := make(map[string]interface{})
		options["title"] = getRequiredStringParam(request.Params.Arguments, "title")
		options["head"] = getRequiredStringParam(request.Params.Arguments, "head")
		options["base"] = getRequiredStringParam(request.Params.Arguments, "base")

		if body, ok := getStringParam(request.Params.Arguments, "body"); ok {
			options["body"] = body
		}

		options["draft"] = getBoolParam(request.Params.Arguments, "draft", true)

		result, err := client.CreatePullRequest(owner, repo, options)
		if err != nil {
			return nil, err
		}

		jsonResult, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return nil, err
		}

		return mcp.NewToolResultText(string(jsonResult)), nil
	})

	// ツール7: コード検索
	searchCodeTool := mcp.NewTool("search_code",
		mcp.WithDescription("Search for code across GitHub repositories"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query. This tool must have authentication to access the code search API. Here is the example of url with 'q' parameter to request: https://api.github.com/search/code?q=addClass+in:file+language:js+repo:jquery/jquery"),
		),
		mcp.WithNumber("page",
			mcp.Description("Page number (default: 1)"),
		),
		mcp.WithNumber("per_page",
			mcp.Description("Results per page (default: 30, max: 100)"),
		),
	)

	s.AddTool(searchCodeTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query := getRequiredStringParam(request.Params.Arguments, "query")

		options := make(map[string]interface{})

		// 数値オプションパラメータを追加
		if page, ok := request.Params.Arguments["page"]; ok {
			options["page"] = int(page.(float64))
		}
		if perPage, ok := request.Params.Arguments["per_page"]; ok {
			options["per_page"] = int(perPage.(float64))
		}

		result, err := client.SearchCode(query, options)
		if err != nil {
			return nil, err
		}

		return returnJSONResult(result)
	})

	// ツール8: イシューの更新
	updateIssueTool := mcp.NewTool("update_issue",
		mcp.WithDescription("Update an existing issue in a GitHub repository"),
		mcp.WithString("owner",
			mcp.Required(),
			mcp.Description("Repository owner"),
		),
		mcp.WithString("repo",
			mcp.Required(),
			mcp.Description("Repository name"),
		),
		mcp.WithNumber("issue_number",
			mcp.Required(),
			mcp.Description("Issue number"),
		),
		mcp.WithString("title",
			mcp.Description("New issue title"),
		),
		mcp.WithString("body",
			mcp.Description("New issue body"),
		),
		mcp.WithString("state",
			mcp.Description("State of the issue: open or closed"),
			mcp.Enum("open", "closed"),
		),
		mcp.WithArray("labels",
			mcp.Description("New labels for the issue"),
		),
		mcp.WithArray("assignees",
			mcp.Description("New assignees for the issue"),
		),
	)

	s.AddTool(updateIssueTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := getRequiredStringParam(request.Params.Arguments, "owner")
		repo := getRequiredStringParam(request.Params.Arguments, "repo")
		issueNumber := getNumberParam(request.Params.Arguments, "pull_number", 1)

		options := make(map[string]interface{})

		// 文字列オプションパラメータを追加
		if title, ok := getStringParam(request.Params.Arguments, "title"); ok {
			options["title"] = title
		}
		if body, ok := getStringParam(request.Params.Arguments, "body"); ok {
			options["body"] = body
		}
		if state, ok := getStringParam(request.Params.Arguments, "state"); ok {
			options["state"] = state
		}

		// 配列パラメータを追加
		addToOptions(options, request.Params.Arguments, "labels")
		addToOptions(options, request.Params.Arguments, "assignees")

		result, err := client.UpdateIssue(owner, repo, issueNumber, options)
		if err != nil {
			return nil, err
		}

		return returnJSONResult(result)
	})

	// ツール9: イシューコメントの追加
	addIssueCommentTool := mcp.NewTool("add_issue_comment",
		mcp.WithDescription("Add a comment to an existing issue"),
		mcp.WithString("owner",
			mcp.Required(),
			mcp.Description("Repository owner"),
		),
		mcp.WithString("repo",
			mcp.Required(),
			mcp.Description("Repository name"),
		),
		mcp.WithNumber("issue_number",
			mcp.Required(),
			mcp.Description("Issue number"),
		),
		mcp.WithString("body",
			mcp.Required(),
			mcp.Description("Comment body"),
		),
	)

	s.AddTool(addIssueCommentTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := getRequiredStringParam(request.Params.Arguments, "owner")
		repo := getRequiredStringParam(request.Params.Arguments, "repo")
		issueNumber := int(request.Params.Arguments["issue_number"].(float64))
		body := getRequiredStringParam(request.Params.Arguments, "body")

		result, err := client.AddIssueComment(owner, repo, issueNumber, body)
		if err != nil {
			return nil, err
		}

		return returnJSONResult(result)
	})

	// ツール10: コミット一覧の取得
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
		mcp.WithNumber("perPage",
			mcp.Description("Results per page (default: 30, max: 100)"),
		),
		mcp.WithString("sha",
			mcp.Description("SHA or branch name to start listing commits from"),
		),
	)

	s.AddTool(listCommitsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := getRequiredStringParam(request.Params.Arguments, "owner")
		repo := getRequiredStringParam(request.Params.Arguments, "repo")
		page := getNumberParam(request.Params.Arguments, "page", 1)
		perPage := getNumberParam(request.Params.Arguments, "perPage", 30)
		sha, _ := getStringParam(request.Params.Arguments, "sha")

		result, err := client.ListCommits(owner, repo, page, perPage, sha)
		if err != nil {
			return nil, err
		}

		return returnJSONResult(result)
	})

	// ツール11: プルリクエストレビューの作成
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

	s.AddTool(createPullRequestReviewTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := getRequiredStringParam(request.Params.Arguments, "owner")
		repo := getRequiredStringParam(request.Params.Arguments, "repo")
		pullNumber := getNumberParam(request.Params.Arguments, "pull_number", 1)

		options := make(map[string]interface{})

		// 文字列オプションパラメータを追加
		if event, ok := getStringParam(request.Params.Arguments, "event"); ok {
			options["event"] = event
		}
		if body, ok := getStringParam(request.Params.Arguments, "body"); ok {
			options["body"] = body
		}

		result, err := client.CreatePullRequestReview(owner, repo, pullNumber, options)
		if err != nil {
			return nil, err
		}

		return returnJSONResult(result)
	})

	// ツール12: プルリクエストのマージ
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

	s.AddTool(mergePullRequestTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := getRequiredStringParam(request.Params.Arguments, "owner")
		repo := getRequiredStringParam(request.Params.Arguments, "repo")
		pullNumber := getNumberParam(request.Params.Arguments, "pull_number", 1)

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

		result, err := client.MergePullRequest(owner, repo, pullNumber, options)
		if err != nil {
			return nil, err
		}

		jsonResult, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return nil, err
		}

		return mcp.NewToolResultText(string(jsonResult)), nil
	})

	// ツール13: プルリクエストのファイル一覧取得
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

	s.AddTool(getPullRequestFilesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := request.Params.Arguments["owner"].(string)
		repo := request.Params.Arguments["repo"].(string)
		pullNumber := getNumberParam(request.Params.Arguments, "pull_number", 1)

		result, err := client.GetPullRequestFiles(owner, repo, pullNumber)
		if err != nil {
			return nil, err
		}

		jsonResult, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return nil, err
		}

		return mcp.NewToolResultText(string(jsonResult)), nil
	})

	// ツール14: プルリクエストのステータス取得
	getPullRequestStatusTool := mcp.NewTool("get_pull_request_status",
		mcp.WithDescription("Get the combined status of all status checks for a pull request"),
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

	s.AddTool(getPullRequestStatusTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := getRequiredStringParam(request.Params.Arguments, "owner")
		repo := getRequiredStringParam(request.Params.Arguments, "repo")
		pullNumber := getNumberParam(request.Params.Arguments, "pull_number", 1)

		result, err := client.GetPullRequestStatus(owner, repo, pullNumber)
		if err != nil {
			return nil, err
		}

		jsonResult, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return nil, err
		}

		return mcp.NewToolResultText(string(jsonResult)), nil
	})

	// ツール15: プルリクエストブランチの更新
	updatePullRequestBranchTool := mcp.NewTool("update_pull_request_branch",
		mcp.WithDescription("Update a pull request branch with the latest changes from the base branch"),
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
		mcp.WithString("expected_head_sha",
			mcp.Description("The expected SHA of the pull request head"),
		),
	)

	s.AddTool(updatePullRequestBranchTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := getRequiredStringParam(request.Params.Arguments, "owner")
		repo := getRequiredStringParam(request.Params.Arguments, "repo")
		pullNumber := getNumberParam(request.Params.Arguments, "pull_number", 1)

		var expectedHeadSHA string
		if sha, ok := getStringParam(request.Params.Arguments, "expected_head_sha"); ok {
			expectedHeadSHA = sha
		}

		err := client.UpdatePullRequestBranch(owner, repo, pullNumber, expectedHeadSHA)
		if err != nil {
			return nil, err
		}

		return mcp.NewToolResultText(`{"success": true}`), nil
	})

	// ツール16: プルリクエストコメントの取得
	getPullRequestCommentsTool := mcp.NewTool("get_pull_request_comments",
		mcp.WithDescription("Get the review comments on a pull request"),
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

	s.AddTool(getPullRequestCommentsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := getRequiredStringParam(request.Params.Arguments, "owner")
		repo := getRequiredStringParam(request.Params.Arguments, "repo")
		pullNumber := getNumberParam(request.Params.Arguments, "pull_number", 1)

		result, err := client.GetPullRequestComments(owner, repo, pullNumber)
		if err != nil {
			return nil, err
		}

		jsonResult, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return nil, err
		}

		return mcp.NewToolResultText(string(jsonResult)), nil
	})

	// ツール17: プルリクエストレビューの取得
	getPullRequestReviewsTool := mcp.NewTool("get_pull_request_reviews",
		mcp.WithDescription("Get the reviews on a pull request"),
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

	s.AddTool(getPullRequestReviewsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := getRequiredStringParam(request.Params.Arguments, "owner")
		repo := getRequiredStringParam(request.Params.Arguments, "repo")
		pullNumber := getNumberParam(request.Params.Arguments, "pull_number", 1)

		result, err := client.GetPullRequestReviews(owner, repo, pullNumber)
		if err != nil {
			return nil, err
		}

		jsonResult, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return nil, err
		}

		return mcp.NewToolResultText(string(jsonResult)), nil
	})

	// ツール18: プルリクエスト一覧の取得
	listPullRequestsTool := mcp.NewTool("list_pull_requests",
		mcp.WithDescription("List and filter repository pull requests"),
		mcp.WithString("owner",
			mcp.Required(),
			mcp.Description("Repository owner"),
		),
		mcp.WithString("repo",
			mcp.Required(),
			mcp.Description("Repository name"),
		),
		mcp.WithString("state",
			mcp.Description("Pull request state: open, closed, or all (default: open)"),
			mcp.Enum("open", "closed", "all"),
		),
		mcp.WithString("sort",
			mcp.Description("Sort field: created, updated, popularity, long-running (default: created)"),
			mcp.Enum("created", "updated", "popularity", "long-running"),
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
		mcp.WithString("head",
			mcp.Description("Filter by head user or head organization and branch name in the format of 'user:ref-name' or 'organization:ref-name'"),
		),
		mcp.WithString("base",
			mcp.Description("Filter by base branch name"),
		),
	)

	s.AddTool(listPullRequestsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := getRequiredStringParam(request.Params.Arguments, "owner")
		repo := getRequiredStringParam(request.Params.Arguments, "repo")

		options := make(map[string]interface{})

		if state, ok := getStringParam(request.Params.Arguments, "state"); ok {
			options["state"] = state
		}
		if sort, ok := getStringParam(request.Params.Arguments, "sort"); ok {
			options["sort"] = sort
		}
		if direction, ok := getStringParam(request.Params.Arguments, "direction"); ok {
			options["direction"] = direction
		}
		if perPage, ok := getStringParam(request.Params.Arguments, "per_page"); ok {
			options["per_page"] = perPage
		}
		if page, ok := getStringParam(request.Params.Arguments, "page"); ok {
			options["page"] = page
		}
		if head, ok := getStringParam(request.Params.Arguments, "head"); ok {
			options["head"] = head
		}
		if base, ok := getStringParam(request.Params.Arguments, "base"); ok {
			options["base"] = base
		}

		result, err := client.ListPullRequests(owner, repo, options)
		if err != nil {
			return nil, err
		}

		jsonResult, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return nil, err
		}

		return mcp.NewToolResultText(string(jsonResult)), nil
	})

	// サーバーを起動
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

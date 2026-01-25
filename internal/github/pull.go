package github

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

// HandleToGetPullRequestStatus はプルリクエストのステータスを取得して、結果をJSON形式で返します
func (c *GitHubClient) HandleToGetPullRequestStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	owner := getRequiredStringParam(request.Params.Arguments, "owner")
	repo := getRequiredStringParam(request.Params.Arguments, "repo")
	pullNumber := int(request.Params.Arguments["pull_number"].(float64))

	result, err := c.GetPullRequestStatus(owner, repo, pullNumber)
	if err != nil {
		return nil, err
	}

	return returnJSONResult(result)
}

// UpdatePullRequestBranch はプルリクエストのブランチを更新します
func (c *GitHubClient) UpdatePullRequestBranch(owner, repo string, pullNumber int, expectedHeadSHA string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d/update-branch", apiBaseURL, owner, repo, pullNumber)

	options := map[string]interface{}{}
	if expectedHeadSHA != "" {
		options["expected_head_sha"] = expectedHeadSHA
	}

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

// HandleToUpdatePullRequestBranch はプルリクエストのブランチを更新して、結果をJSON形式で返します
func (c *GitHubClient) HandleToUpdatePullRequestBranch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	owner := getRequiredStringParam(request.Params.Arguments, "owner")
	repo := getRequiredStringParam(request.Params.Arguments, "repo")
	pullNumber := int(request.Params.Arguments["pull_number"].(float64))

	var expectedHeadSHA string
	if sha, ok := getStringParam(request.Params.Arguments, "expected_head_sha"); ok {
		expectedHeadSHA = sha
	}

	result, err := c.UpdatePullRequestBranch(owner, repo, pullNumber, expectedHeadSHA)
	if err != nil {
		return nil, err
	}

	return returnJSONResult(result)
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

// HandleToGetPullRequestComments はプルリクエストのコメントを取得して、結果をJSON形式で返します
func (c *GitHubClient) HandleToGetPullRequestComments(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	owner := getRequiredStringParam(request.Params.Arguments, "owner")
	repo := getRequiredStringParam(request.Params.Arguments, "repo")
	pullNumber := int(request.Params.Arguments["pull_number"].(float64))

	result, err := c.GetPullRequestComments(owner, repo, pullNumber)
	if err != nil {
		return nil, err
	}

	return returnJSONResult(result)
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

// HandleToGetPullRequestReviews はプルリクエストのレビューを取得して、結果をJSON形式で返します
func (c *GitHubClient) HandleToGetPullRequestReviews(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	owner := getRequiredStringParam(request.Params.Arguments, "owner")
	repo := getRequiredStringParam(request.Params.Arguments, "repo")
	pullNumber := int(request.Params.Arguments["pull_number"].(float64))

	result, err := c.GetPullRequestReviews(owner, repo, pullNumber)
	if err != nil {
		return nil, err
	}

	return returnJSONResult(result)
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

// HandleToListPullRequests はリポジトリのプルリクエスト一覧を取得して、結果をJSON形式で返します
func (c *GitHubClient) HandleToListPullRequests(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	result, err := c.ListPullRequests(owner, repo, options)
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

	// ツール5: プルリクエストのステータス取得
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
	s.AddTool(getPullRequestStatusTool, client.HandleToGetPullRequestStatus)

	// ツール6: プルリクエストブランチの更新
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
	s.AddTool(updatePullRequestBranchTool, client.HandleToUpdatePullRequestBranch)

	// ツール7: プルリクエストコメントの取得
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
	s.AddTool(getPullRequestCommentsTool, client.HandleToGetPullRequestComments)

	// ツール8: プルリクエストレビューの取得
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
	s.AddTool(getPullRequestReviewsTool, client.HandleToGetPullRequestReviews)

	// ツール9: プルリクエスト一覧の取得
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
	s.AddTool(listPullRequestsTool, client.HandleToListPullRequests)

	return s
}

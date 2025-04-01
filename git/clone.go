package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	mcp "github.com/mark3labs/mcp-go/mcp"
	server "github.com/mark3labs/mcp-go/server"
)

// CloneRepository はGitリポジトリをクローンします
func (c *GitClient) CloneRepository(repoURL, directory string, options map[string]interface{}) (map[string]interface{}, error) {
	// ディレクトリが存在しない場合は作成
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		if err := os.MkdirAll(directory, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %v", err)
		}
	}

	// クローンコマンドの引数を準備
	args := []string{"clone", repoURL, directory}

	// オプションの処理
	if branch, ok := options["branch"].(string); ok && branch != "" {
		args = append(args, "--branch", branch)
	}

	if depth, ok := options["depth"].(int); ok && depth > 0 {
		args = append(args, "--depth", fmt.Sprintf("%d", depth))
	}

	if getBoolParam(options, "single_branch", false) {
		args = append(args, "--single-branch")
	}

	// Gitコマンドを実行
	output, err := c.executeGitCommand(args...)
	if err != nil {
		return nil, err
	}

	// 結果を返す
	absPath, _ := filepath.Abs(directory)
	result := map[string]interface{}{
		"success":   true,
		"message":   "Repository cloned successfully",
		"output":    output,
		"directory": absPath,
	}

	return result, nil
}

// HandleToCloneRepository はGitリポジトリをクローンして、結果をJSON形式で返します
func (c *GitClient) HandleToCloneRepository(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repoURL := getRequiredStringParam(request.Params.Arguments, "repo_url")
	directory := getRequiredStringParam(request.Params.Arguments, "directory")

	options := make(map[string]interface{})

	// オプションパラメータを追加
	if branch, ok := getStringParam(request.Params.Arguments, "branch"); ok {
		options["branch"] = branch
	}

	if depth, ok := request.Params.Arguments["depth"]; ok {
		options["depth"] = int(depth.(float64))
	}

	options["single_branch"] = getBoolParam(request.Params.Arguments, "single_branch", false)

	result, err := c.CloneRepository(repoURL, directory, options)
	if err != nil {
		return nil, err
	}

	return returnJSONResult(result)
}

// SetGitCloneServer は受け取ったMCPサーバにGitクローン用のツールを付与して、そのMCPサーバを返します。
func SetGitCloneServer(s *server.MCPServer) *server.MCPServer {
	// Gitクライアントを初期化
	client := NewGitClient()

	// ツール: リポジトリのクローン
	cloneRepositoryTool := mcp.NewTool("clone_repository",
		mcp.WithDescription("Clone a Git repository to a specified directory"),
		mcp.WithString("repo_url",
			mcp.Required(),
			mcp.Description("URL of the Git repository to clone"),
		),
		mcp.WithString("directory",
			mcp.Required(),
			mcp.Description("Directory to clone the repository into"),
		),
		mcp.WithString("branch",
			mcp.Description("Branch to clone (optional)"),
		),
		mcp.WithNumber("depth",
			mcp.Description("Create a shallow clone with a history truncated to the specified number of commits"),
		),
		mcp.WithBoolean("single_branch",
			mcp.Description("Clone only the history leading to the tip of a single branch"),
		),
	)
	s.AddTool(cloneRepositoryTool, client.HandleToCloneRepository)

	return s
}

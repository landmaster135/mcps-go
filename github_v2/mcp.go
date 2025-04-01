package github_v2

import (
	"fmt"
	"os"

	server "github.com/mark3labs/mcp-go/server"
)

func createGitHubServer() *server.MCPServer {
	// 環境変数からGitHubトークンを取得
	token := os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN")
	if token == "" {
		fmt.Println("Warning: GITHUB_PERSONAL_ACCESS_TOKEN environment variable not set. API rate limits will be restricted.")
	}

	// MCPサーバーを作成
	s := server.NewMCPServer(
		"GitHub API Server",
		version,
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
	)
	s = SetGitHubIssueServer(token, s)
	s = SetGitHubPullRequestServer(token, s)

	return s
}

// BuildGitHubServer はGitHubのMCPサーバーを構築します
func BuildGitHubServer() {
	s := createGitHubServer()

	// サーバーを起動
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

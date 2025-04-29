package github

import (
	"context"
	"fmt"
	"os"

	mcp "github.com/mark3labs/mcp-go/mcp"
	server "github.com/mark3labs/mcp-go/server"
)

func addPromptIntoServer(s *server.MCPServer) *server.MCPServer {
	prompt := mcp.NewPrompt("system_prompt_01",
		mcp.WithPromptDescription("This is a prompt for the GitHub client."),
	)
	s.AddPrompt(prompt, func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return &mcp.GetPromptResult{
			Description: "System prompt for the GitHub client.",
			Messages: []mcp.PromptMessage{
				{
					Role:    mcp.RoleAssistant,
					Content: mcp.NewTextContent("You use this great GitHub client well."),
				},
			},
		}, nil
	})
	return s
}

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
		server.WithPromptCapabilities(true),
		server.WithLogging(),
	)
	s = SetGitHubIssueServer(token, s)
	s = SetGitHubPullRequestServer(token, s)
	s = SetGitHubRepositoryServer(token, s)
	s = SetGitHubGlobalServer(token, s)

	// プロンプト
	s = addPromptIntoServer(s)

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

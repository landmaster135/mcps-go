package everart

import (
	"context"
	"fmt"
	"os"

	mcp "github.com/mark3labs/mcp-go/mcp"
	server "github.com/mark3labs/mcp-go/server"
)

func createEverArtServer() *server.MCPServer {
	// 環境変数からEverArt APIキーを取得
	apiKey := os.Getenv("EVERART_API_KEY")
	if apiKey == "" {
		fmt.Println("Warning: EVERART_API_KEY environment variable not set. API will not function correctly.")
	}

	// MCPサーバーを作成
	s := server.NewMCPServer(
		"EverArt API Server",
		version,
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
		server.WithLogging(),
	)
	s = SetEverArtImageServer(apiKey, s)

	return s
}

func addPromptIntoServer(s *server.MCPServer) *server.MCPServer {
	prompt := mcp.NewPrompt("system_prompt_01",
		mcp.WithPromptDescription("This is a prompt for the EverArt client."),
	)
	s.AddPrompt(prompt, func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return &mcp.GetPromptResult{
			Description: "System prompt for the EverArt client.",
			Messages: []mcp.PromptMessage{
				{
					Role:    mcp.RoleAssistant,
					Content: mcp.NewTextContent("You use this great client for EverArt well."),
				},
			},
		}, nil
	})
	return s
}

// BuildEverArtServer はEverArtのMCPサーバーを構築します
func BuildEverArtServer() {
	s := createEverArtServer()

	// プロンプト
	s = addPromptIntoServer(s)

	// サーバーを起動
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

package gdrive

import (
	"context"
	"fmt"
	"os"

	mcp "github.com/mark3labs/mcp-go/mcp"
	server "github.com/mark3labs/mcp-go/server"
)

// バージョン情報
const version = "0.1.0"

func addPromptIntoServer(s *server.MCPServer) *server.MCPServer {
	prompt := mcp.NewPrompt("system_prompt_01",
		mcp.WithPromptDescription("This is a prompt for the Google Drive client."),
	)
	s.AddPrompt(prompt, func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return &mcp.GetPromptResult{
			Description: "System prompt for the Google Drive client.",
			Messages: []mcp.PromptMessage{
				{
					Role:    mcp.RoleAssistant,
					Content: mcp.NewTextContent("You use this great client for Google Drive well."),
				},
			},
		}, nil
	})
	return s
}

// createGoogleDriveServer はGoogleドライブのMCPサーバーを作成します
func createGoogleDriveServer() *server.MCPServer {
	// 環境変数から認証情報のパスを取得
	credentialsPath := os.Getenv("GDRIVE_CREDENTIALS_PATH")
	if credentialsPath == "" {
		fmt.Fprintln(os.Stderr, "Warning: GDRIVE_CREDENTIALS_PATH environment variable not set. Using default path.")
		credentialsPath = ".gdrive-server-credentials.json"
	}

	// MCPサーバーを作成
	s := server.NewMCPServer(
		"Google Drive API Server",
		version,
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
		server.WithLogging(),
	)

	// Googleドライブ機能を設定
	s = SetGoogleDriveServer(credentialsPath, s)

	// プロンプト
	s = addPromptIntoServer(s)

	return s
}

// BuildGoogleDriveServer はGoogleドライブのMCPサーバーを構築します
func BuildGoogleDriveServer() {
	s := createGoogleDriveServer()

	// サーバーを起動
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
	}
}

package everart

import (
	"fmt"
	"os"

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
		server.WithLogging(),
	)
	s = SetEverArtImageServer(apiKey, s)

	return s
}

// BuildEverArtServer はEverArtのMCPサーバーを構築します
func BuildEverArtServer() {
	s := createEverArtServer()

	// サーバーを起動
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

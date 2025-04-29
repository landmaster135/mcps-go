package gdrive

import (
	"fmt"
	"os"

	server "github.com/mark3labs/mcp-go/server"
)

// バージョン情報
const version = "0.1.0"

// createGoogleDriveServer はGoogleドライブのMCPサーバーを作成します
func createGoogleDriveServer() *server.MCPServer {
	// 環境変数から認証情報のパスを取得
	credentialsPath := os.Getenv("GDRIVE_CREDENTIALS_PATH")
	if credentialsPath == "" {
		fmt.Println("Warning: GDRIVE_CREDENTIALS_PATH environment variable not set. Using default path.")
		credentialsPath = ".gdrive-server-credentials.json"
	}

	// MCPサーバーを作成
	s := server.NewMCPServer(
		"Google Drive API Server",
		version,
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
	)

	// Googleドライブ機能を設定
	s = SetGoogleDriveServer(credentialsPath, s)

	return s
}

// BuildGoogleDriveServer はGoogleドライブのMCPサーバーを構築します
func BuildGoogleDriveServer() {
	s := createGoogleDriveServer()

	// サーバーを起動
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

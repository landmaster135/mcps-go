package git

import (
	"fmt"

	util "mcps-go/mcps-go/util"
	server "github.com/mark3labs/mcp-go/server"
)

func createGitServer() *server.MCPServer {
	// MCPサーバーを作成
	s := server.NewMCPServer(
		"Git Command Server",
		version,
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
	)
	s = SetGitCloneServer(s)

	return s
}

// BuildGitServer はGitのMCPサーバーを構築します
func BuildGitServer() {
	util.OutLog("Building Git MCP server...")
	s := createGitServer()

	// サーバーを起動
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

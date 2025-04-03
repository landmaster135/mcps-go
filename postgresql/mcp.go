package postgresql

import (
	"fmt"
	"os"
	// _ "github.com/go-pg/pg/v10"

	server "github.com/mark3labs/mcp-go/server"
)

// func createPostgreSQLServer() *server.MCPServer {
// 	// 環境変数からデータベースURLを取得
// 	databaseURL := os.Getenv("POSTGRESQL_DATABASE_URL")
// 	if databaseURL == "" {
// 		fmt.Println("Error: POSTGRESQL_DATABASE_URL environment variable not set.")
// 		os.Exit(1)
// 	}

// 	// MCPサーバーを作成
// 	s := server.NewMCPServer(
// 		"PostgreSQL Database Server",
// 		version,
// 		server.WithResourceCapabilities(true, true),
// 		server.WithLogging(),
// 	)
// 	s = SetPostgreSQLQueryServer(databaseURL, s)

// 	return s
// }

func createPostgreSQLServer() *server.MCPServer {
	// 環境変数からデータベースURLを取得
	databaseURL := os.Getenv("POSTGRESQL_DATABASE_URL")
	if databaseURL == "" {
		fmt.Println("Error: POSTGRESQL_DATABASE_URL environment variable not set.")
		os.Exit(1)
	}

	// MCPサーバーを作成
	s := server.NewMCPServer(
		"PostgreSQL Database Server",
		version,
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
	)
	s = SetPostgreSQLQueryServer(databaseURL, s)

	return s
}

// BuildPostgreSQLServer はPostgreSQLのMCPサーバーを構築します
func BuildPostgreSQLServer() {
	s := createPostgreSQLServer()

	// サーバーを起動
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

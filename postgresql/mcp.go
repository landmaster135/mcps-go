package postgresql

import (
	"fmt"
	"os"

	mcp "github.com/mark3labs/mcp-go/mcp"
	server "github.com/mark3labs/mcp-go/server"
)

// SetPostgreSQLQueryServer は受け取ったMCPサーバにPostgreSQL用のツールを付与して、そのMCPサーバを返します。
func SetPostgreSQLQueryServer(databaseURL string, s *server.MCPServer) *server.MCPServer {
	// PostgreSQLクライアントを初期化
	client, err := NewPostgreSQLClient(databaseURL)
	if err != nil {
		fmt.Printf("Failed to create PostgreSQL client: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(client)

	// ツール1: SQL読み取り専用クエリの実行
	queryTool := mcp.NewTool("query",
		mcp.WithDescription("Run a read-only SQL query"),
		mcp.WithString("sql",
			mcp.Required(),
			mcp.Description("SQL query to execute"),
		),
	)
	s.AddTool(queryTool, client.HandleToQuery)

	// ツール2: テーブルスキーマの取得
	getTableSchemaTool := mcp.NewTool("get_table_schema",
		mcp.WithDescription("Get schema information for a database table"),
		mcp.WithString("table_name",
			mcp.Required(),
			mcp.Description("Name of the table"),
		),
	)
	s.AddTool(getTableSchemaTool, client.HandleToGetTableSchema)

	// ツール3: テーブル一覧の取得
	listTablesTool := mcp.NewTool("list_tables",
		mcp.WithDescription("List all tables in the database"),
	)
	s.AddTool(listTablesTool, client.HandleToListTables)

	// ツール4: テーブルの最小限のスキーマ情報を取得
	getTableSchemaMinimumTool := mcp.NewTool("get_table_schema_minimum",
		mcp.WithDescription("Get minimal schema information for a database table (column names and data types only)"),
		mcp.WithString("table_name",
			mcp.Required(),
			mcp.Description("Name of the table"),
		),
	)
	s.AddTool(getTableSchemaMinimumTool, client.HandleToGetTableSchemaMinimum)

	return s
}

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

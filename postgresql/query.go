package postgresql

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"database/sql"
	_ "github.com/lib/pq"

	mcp "github.com/mark3labs/mcp-go/mcp"
	server "github.com/mark3labs/mcp-go/server"
)

const (
	schemaPath = "schema"
)

// GetTables はデータベース内のテーブル一覧を取得します
func (c *PostgreSQLClient) GetTables() ([]Table, error) {
	query := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public'
	`

	rows, err := c.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []Table
	for rows.Next() {
		var table Table
		if err := rows.Scan(&table.Name); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tables, nil
}

// GetTableSchema はテーブルのスキーマ情報を取得します
func (c *PostgreSQLClient) GetTableSchema(tableName string) ([]Column, error) {
	query := `
		SELECT column_name, data_type
		FROM information_schema.columns
		WHERE table_name = $1
	`

	rows, err := c.db.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []Column
	for rows.Next() {
		var column Column
		if err := rows.Scan(&column.Name, &column.DataType); err != nil {
			return nil, err
		}
		columns = append(columns, column)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return columns, nil
}

// ExecuteQuery はSQL読み取り専用クエリを実行します
func (c *PostgreSQLClient) ExecuteQuery(ctx context.Context, sqlQuery string) ([]map[string]interface{}, error) {
	// トランザクションを開始（読み取り専用）
	tx, err := c.db.BeginTx(ctx, &sql.TxOptions{
		ReadOnly: true,
	})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// クエリを実行
	rows, err := tx.QueryContext(ctx, sqlQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// カラム名を取得
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// 結果を格納するスライス
	var result []map[string]interface{}

	// 各行を処理
	for rows.Next() {
		// スキャン用のインターフェースのスライスを作成
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		// 行をスキャン
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		// 行データをマップに変換
		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			// バイト配列の場合は文字列に変換
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}

		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// HandleToQuery はSQL読み取り専用クエリを実行して、結果をJSON形式で返します
func (c *PostgreSQLClient) HandleToQuery(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sqlQuery := getRequiredStringParam(request.Params.Arguments, "sql")

	result, err := c.ExecuteQuery(ctx, sqlQuery)
	if err != nil {
		return returnError(err)
	}

	return returnJSONResult(result)
}

// HandleToGetTableSchema はテーブルのスキーマ情報を取得して、結果をJSON形式で返します
func (c *PostgreSQLClient) HandleToGetTableSchema(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tableName := getRequiredStringParam(request.Params.Arguments, "table_name")

	columns, err := c.GetTableSchema(tableName)
	if err != nil {
		return returnError(err)
	}

	return returnJSONResult(columns)
}

// HandleToListTables はデータベース内のテーブル一覧を取得して、結果をJSON形式で返します
func (c *PostgreSQLClient) HandleToListTables(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tables, err := c.GetTables()
	if err != nil {
		return returnError(err)
	}

	return returnJSONResult(tables)
}

// HandleToReadResource はリソースURIからテーブルスキーマを読み取ります
func (c *PostgreSQLClient) HandleToReadResource(ctx context.Context, uri string) ([]byte, error) {
	// URIからテーブル名とパスを抽出
	parsedURL, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	pathParts := strings.Split(strings.TrimPrefix(parsedURL.Path, "/"), "/")
	if len(pathParts) < 2 {
		return nil, fmt.Errorf("invalid resource URI: %s", uri)
	}

	tableName := pathParts[0]
	path := pathParts[1]

	if path != schemaPath {
		return nil, fmt.Errorf("invalid resource path: %s", path)
	}

	// テーブルスキーマを取得
	columns, err := c.GetTableSchema(tableName)
	if err != nil {
		return nil, err
	}

	// JSON形式に変換
	jsonData, err := json.MarshalIndent(columns, "", "  ")
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

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

	return s
}

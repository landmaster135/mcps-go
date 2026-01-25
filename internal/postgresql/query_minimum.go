package postgresql

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"

	mcp "github.com/mark3labs/mcp-go/mcp"
)

// Column はデータベースのカラム情報を表します（最小限の情報）
type Column struct {
	Name     string `json:"column_name"`
	DataType string `json:"data_type"`
}

// Table はデータベースのテーブル情報を表します
type Table struct {
	Name string `json:"table_name"`
}

// getTablesMinimum はデータベース内のテーブル一覧を取得します
func (c *PostgreSQLClient) getTablesMinimum() ([]Table, error) {
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

// HandleToGetTableSchemaMinimum はテーブルの最小限のスキーマ情報を取得して、結果をJSON形式で返します
func (c *PostgreSQLClient) HandleToGetTableSchemaMinimum(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tableName := getRequiredStringParam(request.Params.Arguments, "table_name")

	columns, err := c.getTableSchemaMinimum(tableName)
	if err != nil {
		return returnError(err)
	}

	return returnJSONResult(columns)
}

// getTableSchemaMinimum はテーブルの最小限のスキーマ情報を取得します
func (c *PostgreSQLClient) getTableSchemaMinimum(tableName string) ([]Column, error) {
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

// executeQuery はSQL読み取り専用クエリを実行します
func (c *PostgreSQLClient) executeQuery(ctx context.Context, sqlQuery string) ([]map[string]interface{}, error) {
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

	result, err := c.executeQuery(ctx, sqlQuery)
	if err != nil {
		return returnError(err)
	}

	return returnJSONResult(result)
}

// HandleToListTablesMinimum はデータベース内のテーブル一覧を取得して、結果をJSON形式で返します
func (c *PostgreSQLClient) HandleToListTablesMinimum(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tables, err := c.getTablesMinimum()
	if err != nil {
		return returnError(err)
	}

	return returnJSONResult(tables)
}

package postgresql

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"text/template"

	_ "github.com/lib/pq"
	mcp "github.com/mark3labs/mcp-go/mcp"
)

// ColumnInfo はデータベースのカラム詳細情報を表します
type ColumnInfo struct {
	Name       string         `json:"column_name"`
	Type       string         `json:"data_type"`
	IsNullable string         `json:"is_nullable"`
	Default    sql.NullString `json:"column_default"`
	Comment    string         `json:"column_comment"`
}

// TableSummary はテーブルのサマリー情報を表します
type TableSummary struct {
	Name    string       `json:"table_name"`
	Comment string       `json:"table_comment"`
	PK      []string     `json:"primary_keys"`
	UK      []UniqueKey  `json:"unique_keys"`
	FK      []ForeignKey `json:"foreign_keys"`
}

// UniqueKey は一意キー制約を表します
type UniqueKey struct {
	Name    string   `json:"constraint_name"`
	Columns []string `json:"columns"`
}

// ForeignKey は外部キー制約を表します
type ForeignKey struct {
	Name       string   `json:"constraint_name"`
	Columns    []string `json:"columns"`
	RefTable   string   `json:"referenced_table"`
	RefColumns []string `json:"referenced_columns"`
}

// IndexInfo はインデックス情報を表します
type IndexInfo struct {
	Name    string   `json:"index_name"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"is_unique"`
}

// TableDetail はテーブルの詳細情報を表します
type TableDetail struct {
	Name        string       `json:"table_name"`
	Comment     string       `json:"table_comment"`
	Columns     []ColumnInfo `json:"columns"`
	PrimaryKeys []string     `json:"primary_keys"`
	UniqueKeys  []UniqueKey  `json:"unique_keys"`
	ForeignKeys []ForeignKey `json:"foreign_keys"`
	Indexes     []IndexInfo  `json:"indexes"`
}

// テンプレート関連の定義
const describeTableDetailTemplate = `# テーブル: {{.Name}}{{if .Comment}} - {{.Comment}}{{end}}

## カラム{{range .Columns}}
{{formatColumn .}}{{end}}

## キー情報{{if .PrimaryKeys}}
[PK: {{formatPK .PrimaryKeys}}]{{end}}{{if .UniqueKeys}}
[UK: {{formatUK .UniqueKeys}}]{{end}}{{if .ForeignKeys}}
[FK: {{formatFK .ForeignKeys}}]{{end}}{{if .Indexes}}
[INDEX: {{formatIndex .Indexes}}]{{end}}
`

var funcMap = template.FuncMap{
	"formatPK":     formatPK,
	"formatUK":     formatUK,
	"formatFK":     formatFK,
	"formatColumn": formatColumn,
	"formatIndex":  formatIndex,
}

// formatPK は主キー情報をフォーマットします
func formatPK(pk []string) string {
	if len(pk) == 0 {
		return ""
	}
	pkStr := strings.Join(pk, ", ")
	if len(pk) > 1 {
		pkStr = fmt.Sprintf("(%s)", pkStr)
	}
	return pkStr
}

// formatUK は一意キー情報をフォーマットします
func formatUK(uk []UniqueKey) string {
	if len(uk) == 0 {
		return ""
	}
	var ukInfo []string
	for _, k := range uk {
		if len(k.Columns) > 1 {
			ukInfo = append(ukInfo, fmt.Sprintf("(%s)", strings.Join(k.Columns, ", ")))
		} else {
			ukInfo = append(ukInfo, strings.Join(k.Columns, ", "))
		}
	}
	return strings.Join(ukInfo, "; ")
}

// formatFK は外部キー情報をフォーマットします
func formatFK(fk []ForeignKey) string {
	if len(fk) == 0 {
		return ""
	}
	var fkInfo []string
	for _, k := range fk {
		colStr := strings.Join(k.Columns, ", ")
		refColStr := strings.Join(k.RefColumns, ", ")

		if len(k.Columns) > 1 {
			colStr = fmt.Sprintf("(%s)", colStr)
		}

		if len(k.RefColumns) > 1 {
			refColStr = fmt.Sprintf("(%s)", refColStr)
		}

		fkInfo = append(fkInfo, fmt.Sprintf("%s -> %s.%s",
			colStr,
			k.RefTable,
			refColStr))
	}
	return strings.Join(fkInfo, "; ")
}

// formatColumn はカラム情報をフォーマットします
func formatColumn(col ColumnInfo) string {
	nullable := "NOT NULL"
	if col.IsNullable == "YES" {
		nullable = "NULL"
	}

	defaultValue := ""
	if col.Default.Valid {
		defaultValue = fmt.Sprintf(" DEFAULT %s", col.Default.String)
	}

	comment := ""
	if col.Comment != "" {
		comment = fmt.Sprintf(" [%s]", col.Comment)
	}

	return fmt.Sprintf("- %s: %s %s%s%s",
		col.Name, col.Type, nullable, defaultValue, comment)
}

// formatIndex はインデックス情報をフォーマットします
func formatIndex(idx []IndexInfo) string {
	if len(idx) == 0 {
		return ""
	}
	var idxInfo []string
	for _, i := range idx {
		if len(i.Columns) > 1 {
			idxInfo = append(idxInfo, fmt.Sprintf("(%s)", strings.Join(i.Columns, ", ")))
		} else {
			idxInfo = append(idxInfo, strings.Join(i.Columns, ", "))
		}
	}
	return strings.Join(idxInfo, "; ")
}

// fetchTableWithComments はテーブル名とコメントを取得します
func (c *PostgreSQLClient) fetchTableWithComments(ctx context.Context, tableName string) (TableSummary, error) {
	query := `
		SELECT t.table_name,
		       COALESCE(pg_catalog.obj_description(pg_catalog.pg_class.oid), '') AS table_comment
		FROM information_schema.tables t
		JOIN pg_catalog.pg_class ON pg_catalog.pg_class.relname = t.table_name
		WHERE t.table_schema = 'public' AND t.table_name = $1
	`

	var table TableSummary
	err := c.db.QueryRowContext(ctx, query, tableName).Scan(&table.Name, &table.Comment)
	if err != nil {
		return TableSummary{}, err
	}

	return table, nil
}

// fetchPrimaryKeys はテーブルの主キーカラムを取得します
func (c *PostgreSQLClient) fetchPrimaryKeys(ctx context.Context, tableName string) ([]string, error) {
	query := `
		SELECT a.attname
		FROM pg_index i
		JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
		WHERE i.indrelid = $1::regclass
		AND i.indisprimary
		ORDER BY a.attnum
	`

	rows, err := c.db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var primaryKeys []string
	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			return nil, err
		}
		primaryKeys = append(primaryKeys, columnName)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return primaryKeys, nil
}

// fetchUniqueKeys はテーブルの一意キー制約を取得します
func (c *PostgreSQLClient) fetchUniqueKeys(ctx context.Context, tableName string) ([]UniqueKey, error) {
	query := `
		SELECT
			c.conname AS constraint_name,
			a.attname AS column_name
		FROM
			pg_constraint c
		JOIN
			pg_attribute a ON a.attrelid = c.conrelid AND a.attnum = ANY(c.conkey)
		WHERE
			c.conrelid = $1::regclass
			AND c.contype = 'u'
		ORDER BY
			c.conname, a.attnum
	`

	rows, err := c.db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var uniqueKeys []UniqueKey
	var currentUniqueKey *UniqueKey
	for rows.Next() {
		var constraintName, columnName string
		if err := rows.Scan(&constraintName, &columnName); err != nil {
			return nil, err
		}

		if currentUniqueKey == nil || currentUniqueKey.Name != constraintName {
			uniqueKeys = append(uniqueKeys, UniqueKey{
				Name:    constraintName,
				Columns: []string{},
			})
			currentUniqueKey = &uniqueKeys[len(uniqueKeys)-1]
		}

		currentUniqueKey.Columns = append(currentUniqueKey.Columns, columnName)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return uniqueKeys, nil
}

// fetchForeignKeys はテーブルの外部キー制約を取得します
func (c *PostgreSQLClient) fetchForeignKeys(ctx context.Context, tableName string) ([]ForeignKey, error) {
	query := `
		SELECT
			c.conname AS constraint_name,
			a.attname AS column_name,
			cl.relname AS referenced_table,
			af.attname AS referenced_column
		FROM
			pg_constraint c
		JOIN
			pg_attribute a ON a.attrelid = c.conrelid AND a.attnum = ANY(c.conkey)
		JOIN
			pg_class cl ON cl.oid = c.confrelid
		JOIN
			pg_attribute af ON af.attrelid = c.confrelid AND af.attnum = ANY(c.confkey)
		WHERE
			c.conrelid = $1::regclass
			AND c.contype = 'f'
		ORDER BY
			c.conname, a.attnum
	`

	rows, err := c.db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var foreignKeys []ForeignKey
	var currentFK *ForeignKey
	for rows.Next() {
		var constraintName, columnName, refTableName, refColumnName string
		if err := rows.Scan(&constraintName, &columnName, &refTableName, &refColumnName); err != nil {
			return nil, err
		}

		if currentFK == nil || currentFK.Name != constraintName {
			foreignKeys = append(foreignKeys, ForeignKey{
				Name:       constraintName,
				RefTable:   refTableName,
				Columns:    []string{},
				RefColumns: []string{},
			})
			currentFK = &foreignKeys[len(foreignKeys)-1]
		}

		currentFK.Columns = append(currentFK.Columns, columnName)
		currentFK.RefColumns = append(currentFK.RefColumns, refColumnName)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return foreignKeys, nil
}

// fetchTableColumns はテーブルのカラム情報を取得します
func (c *PostgreSQLClient) fetchTableColumns(ctx context.Context, tableName string) ([]ColumnInfo, error) {
	query := `
		SELECT
			c.column_name,
			c.data_type,
			c.is_nullable,
			c.column_default,
			COALESCE(pg_catalog.col_description(
				pg_catalog.pg_class.oid,
				c.ordinal_position
			), '') AS column_comment
		FROM
			information_schema.columns c
		JOIN
			pg_catalog.pg_class ON pg_catalog.pg_class.relname = c.table_name
		WHERE
			c.table_schema = 'public'
			AND c.table_name = $1
		ORDER BY
			c.ordinal_position
	`

	rows, err := c.db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		if err := rows.Scan(&col.Name, &col.Type, &col.IsNullable, &col.Default, &col.Comment); err != nil {
			return nil, err
		}
		columns = append(columns, col)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return columns, nil
}

// fetchTableIndexes はテーブルのインデックス情報を取得します
func (c *PostgreSQLClient) fetchTableIndexes(ctx context.Context, tableName string) ([]IndexInfo, error) {
	query := `
		SELECT
			i.relname AS index_name,
			a.attname AS column_name,
			ix.indisunique AS is_unique
		FROM
			pg_index ix
		JOIN
			pg_class i ON i.oid = ix.indexrelid
		JOIN
			pg_attribute a ON a.attrelid = ix.indrelid AND a.attnum = ANY(ix.indkey)
		JOIN
			pg_class t ON t.oid = ix.indrelid
		LEFT JOIN
			pg_constraint c ON c.conindid = ix.indexrelid
		WHERE
			t.relname = $1
			AND c.contype IS NULL
			AND NOT ix.indisprimary
		ORDER BY
			i.relname, a.attnum
	`

	rows, err := c.db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indexes []IndexInfo
	var currentIdx *IndexInfo
	for rows.Next() {
		var indexName, columnName string
		var isUnique bool
		if err := rows.Scan(&indexName, &columnName, &isUnique); err != nil {
			return nil, err
		}

		if currentIdx == nil || currentIdx.Name != indexName {
			indexes = append(indexes, IndexInfo{
				Name:    indexName,
				Unique:  isUnique,
				Columns: []string{},
			})
			currentIdx = &indexes[len(indexes)-1]
		}

		currentIdx.Columns = append(currentIdx.Columns, columnName)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return indexes, nil
}

// getTableDetail はテーブルの詳細情報を取得します
func (c *PostgreSQLClient) getTableDetail(ctx context.Context, tableName string) (*TableDetail, error) {
	// テーブル情報を取得
	tableInfo, err := c.fetchTableWithComments(ctx, tableName)
	if err != nil {
		return nil, fmt.Errorf("テーブル情報の取得に失敗しました: %w", err)
	}

	// 主キー情報を取得
	primaryKeys, err := c.fetchPrimaryKeys(ctx, tableName)
	if err != nil {
		return nil, fmt.Errorf("主キー情報の取得に失敗しました: %w", err)
	}

	// 一意キー情報を取得
	uniqueKeys, err := c.fetchUniqueKeys(ctx, tableName)
	if err != nil {
		return nil, fmt.Errorf("一意キー情報の取得に失敗しました: %w", err)
	}

	// 外部キー情報を取得
	foreignKeys, err := c.fetchForeignKeys(ctx, tableName)
	if err != nil {
		return nil, fmt.Errorf("外部キー情報の取得に失敗しました: %w", err)
	}

	// カラム情報を取得
	columns, err := c.fetchTableColumns(ctx, tableName)
	if err != nil {
		return nil, fmt.Errorf("カラム情報の取得に失敗しました: %w", err)
	}

	// インデックス情報を取得
	indexes, err := c.fetchTableIndexes(ctx, tableName)
	if err != nil {
		return nil, fmt.Errorf("インデックス情報の取得に失敗しました: %w", err)
	}

	// 詳細情報を作成
	detail := &TableDetail{
		Name:        tableInfo.Name,
		Comment:     tableInfo.Comment,
		Columns:     columns,
		PrimaryKeys: primaryKeys,
		UniqueKeys:  uniqueKeys,
		ForeignKeys: foreignKeys,
		Indexes:     indexes,
	}

	return detail, nil
}

// HandleToGetTableSchema はテーブルのスキーマ情報を取得して、結果をテキスト形式で返します
func (c *PostgreSQLClient) HandleToGetTableSchema(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tableName := getRequiredStringParam(request.Params.Arguments, "table_name")

	// テーブルの詳細情報を取得
	detail, err := c.getTableDetail(ctx, tableName)
	if err != nil {
		return returnError(err)
	}

	// テンプレートを使用して結果をフォーマット
	var output bytes.Buffer
	tmpl, err := template.New("describeTableDetail").Funcs(funcMap).Parse(describeTableDetailTemplate)
	if err != nil {
		return returnError(fmt.Errorf("テンプレートの解析に失敗しました: %w", err))
	}

	if err := tmpl.Execute(&output, detail); err != nil {
		return returnError(fmt.Errorf("テンプレートの実行に失敗しました: %w", err))
	}

	return returnTextResult(output.String())
}

func (c *PostgreSQLClient) HandleToListTables(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

}

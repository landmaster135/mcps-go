package postgresql

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	mcp "github.com/mark3labs/mcp-go/mcp"
)

const (
	version = "1.0.0"
)

// PostgreSQLClient はPostgreSQLデータベースとの接続を管理します
type PostgreSQLClient struct {
	db           *sql.DB
	databaseURL  string
	resourceBase string
}

// NewPostgreSQLClient は新しいPostgreSQLクライアントを作成します
func NewPostgreSQLClient(databaseURL string) (*PostgreSQLClient, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database connection: %v\n", err)
		return nil, err
	}

	// 接続テスト
	if err := db.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "Error pinging database: %v\n", err)
		return nil, err
	}

	// リソースベースURLを作成
	resourceBase, err := createResourceBaseURL(databaseURL)
	if err != nil {
		return nil, err
	}

	return &PostgreSQLClient{
		db:           db,
		databaseURL:  databaseURL,
		resourceBase: resourceBase,
	}, nil
}

// Close はデータベース接続を閉じます
func (c *PostgreSQLClient) Close() error {
	return c.db.Close()
}

// createResourceBaseURL はリソースベースURLを作成します
func createResourceBaseURL(databaseURL string) (string, error) {
	u, err := url.Parse(databaseURL)
	if err != nil {
		return "", err
	}

	// プロトコルをpostgresに変更し、パスワードを削除
	u.Scheme = "postgres"
	u.User = url.User(u.User.Username())

	return u.String(), nil
}

// ヘルパー関数: 文字列パラメータを取得
func getStringParam(args map[string]interface{}, key string) (string, bool) {
	if val, ok := args[key]; ok {
		return val.(string), true
	}
	return "", false
}

// ヘルパー関数: 必須の文字列パラメータを取得
func getRequiredStringParam(args map[string]interface{}, key string) string {
	return args[key].(string)
}

// ヘルパー関数: 数値パラメータを取得
func getNumberParam(args map[string]interface{}, key string, defaultVal int) int {
	if val, ok := args[key]; ok {
		return int(val.(float64))
	}
	return defaultVal
}

// ヘルパー関数: ブールパラメータを取得
func getBoolParam(args map[string]interface{}, key string, defaultVal bool) bool {
	if val, ok := args[key]; ok {
		return val.(bool)
	}
	return defaultVal
}

// ヘルパー関数: 結果をJSON形式で返却
func returnJSONResult(result interface{}) (*mcp.CallToolResult, error) {
	jsonResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func returnTextResult(result interface{}) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText(result.(string)), nil
}

// ヘルパー関数: エラーを返却
func returnError(err error) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText(fmt.Sprintf("Error: %v", err)), fmt.Errorf("database error: %w", err)
}

// ヘルパー関数: クエリパラメータを追加
func addQueryParams(baseURL string, params map[string]string) string {
	if len(params) == 0 {
		return baseURL
	}

	queryParts := []string{}
	for k, v := range params {
		queryParts = append(queryParts, fmt.Sprintf("%s=%s", url.QueryEscape(k), url.QueryEscape(v)))
	}

	if strings.Contains(baseURL, "?") {
		return baseURL + "&" + strings.Join(queryParts, "&")
	}
	return baseURL + "?" + strings.Join(queryParts, "&")
}

// QueryResult はクエリ結果を表します
type QueryResult struct {
	Rows []map[string]interface{} `json:"rows"`
}

// PostgreSQLError はPostgreSQLエラーを表します
type PostgreSQLError struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

func (e *PostgreSQLError) Error() string {
	return fmt.Sprintf("PostgreSQL Error: %s (Status: %d)", e.Message, e.StatusCode)
}

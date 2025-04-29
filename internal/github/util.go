package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	mcp "github.com/mark3labs/mcp-go/mcp"
)

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

// ヘルパー関数: オプションマップにパラメータを追加
func addToOptions(options map[string]interface{}, args map[string]interface{}, key string) {
	if val, ok := args[key]; ok {
		options[key] = val
	}
}

const (
	apiBaseURL = "https://api.github.com"
	version    = "1.0.0"
)

// HTTPClient インターフェースを定義
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// GitHubClient 構造体を修正（テスト用）
type GitHubClient struct {
	httpClient HTTPClient
	token      string
}

// NewGitHubClient は新しいGitHubクライアントを作成します
func NewGitHubClient(token string) *GitHubClient {
	return &GitHubClient{
		httpClient: &http.Client{},
		token:      token,
	}
}

// doRequest はHTTPリクエストを実行し、レスポンスを処理します（テスト用に簡略化）
func (c *GitHubClient) doRequest(method, url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if c.token != "" {
		req.Header.Set("Authorization", "token "+c.token)
	}
	if method == "POST" || method == "PATCH" || method == "PUT" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		var ghError GitHubError
		if err := json.Unmarshal(respBody, &ghError); err != nil {
			return nil, fmt.Errorf("HTTP error: %d - %s", resp.StatusCode, string(respBody))
		}
		ghError.StatusCode = resp.StatusCode
		return nil, &ghError
	}

	return respBody, nil
}

// GitHubError はGitHub APIからのエラーを表します
type GitHubError struct {
	Message          string `json:"message"`
	DocumentationURL string `json:"documentation_url"`
	StatusCode       int
}

func (e *GitHubError) Error() string {
	return fmt.Sprintf("GitHub API Error: %s (Status: %d)", e.Message, e.StatusCode)
}

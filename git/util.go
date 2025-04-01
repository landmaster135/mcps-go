package git

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

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

const (
	version = "1.0.0"
)

// GitClient 構造体
type GitClient struct {
}

// NewGitClient は新しいGitクライアントを作成します
func NewGitClient() *GitClient {
	return &GitClient{}
}

// executeGitCommand はGitコマンドを実行します
func (c *GitClient) executeGitCommand(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git command error: %v - %s", err, string(output))
	}
	return strings.TrimSpace(string(output)), nil
}

// GitError はGitコマンドからのエラーを表します
type GitError struct {
	Message    string `json:"message"`
	StatusCode int
}

func (e *GitError) Error() string {
	return fmt.Sprintf("Git Command Error: %s (Status: %d)", e.Message, e.StatusCode)
}

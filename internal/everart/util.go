package everart

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"

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

// ヘルパー関数: 結果をJSON形式で返却
func returnJSONResult(result interface{}) (*mcp.CallToolResult, error) {
	jsonResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(string(jsonResult)), nil
}

const (
	apiBaseURL = "https://api.everart.ai/v1"
	version    = "1.0.0"
)

// HTTPClient インターフェースを定義
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// EverArtClient 構造体
type EverArtClient struct {
	httpClient HTTPClient
	apiKey     string
}

// NewEverArtClient は新しいEverArtクライアントを作成します
func NewEverArtClient(apiKey string) *EverArtClient {
	return &EverArtClient{
		httpClient: &http.Client{},
		apiKey:     apiKey,
	}
}

// doRequest はHTTPリクエストを実行し、レスポンスを処理します
func (c *EverArtClient) doRequest(method, url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	if method == "POST" || method == "PUT" {
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
		var apiError EverArtError
		if err := json.Unmarshal(respBody, &apiError); err != nil {
			return nil, fmt.Errorf("HTTP error: %d - %s", resp.StatusCode, string(respBody))
		}
		apiError.StatusCode = resp.StatusCode
		return nil, &apiError
	}

	return respBody, nil
}

// openURL はURLをデフォルトブラウザで開きます
func openURL(url string) error {
	cmd := exec.Command("open", url)
	return cmd.Run()
}

// EverArtError はEverArt APIからのエラーを表します
type EverArtError struct {
	Message    string `json:"message"`
	StatusCode int
}

func (e *EverArtError) Error() string {
	return fmt.Sprintf("EverArt API Error: %s (Status: %d)", e.Message, e.StatusCode)
}

// GenerationResponse は画像生成レスポンスを表します
type GenerationResponse struct {
	ID       string `json:"id"`
	ImageURL string `json:"image_url"`
	Status   string `json:"status"`
}

package gdrive

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	mcp "github.com/mark3labs/mcp-go/mcp"
	server "github.com/mark3labs/mcp-go/server"
)

// GoogleDriveClient はGoogleドライブAPIとやり取りするためのクライアント
type GoogleDriveClient struct {
	client *http.Client
}

// NewGoogleDriveClient は新しいGoogleドライブクライアントを作成します
func NewGoogleDriveClient(credentialsPath string) (*GoogleDriveClient, error) {
	// 認証情報を読み込む
	b, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %v", err)
	}

	// 認証情報からクライアントを作成
	var creds map[string]interface{}
	if err := json.Unmarshal(b, &creds); err != nil {
		return nil, fmt.Errorf("unable to parse credentials: %v", err)
	}

	// 簡易的な実装として、標準のHTTPクライアントを使用
	// 実際の実装では、OAuth2認証を使用する必要があります
	return &GoogleDriveClient{
		client: &http.Client{},
	}, nil
}

// doRequest はHTTPリクエストを実行し、レスポンスを処理します
func (c *GoogleDriveClient) doRequest(method, url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP error: %d - %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// HandleListResources はGoogleドライブのファイル一覧を取得するハンドラです
func (c *GoogleDriveClient) HandleListResources(ctx context.Context, request mcp.ListResourcesRequest) (*mcp.ListResourcesResult, error) {
	// 実際の実装では、Googleドライブ APIを使用してファイル一覧を取得します
	// この実装はダミーデータを返します
	resources := []mcp.Resource{
		{
			URI:      "gdrive:///dummy1",
			MIMEType: "application/vnd.google-apps.document",
			Name:     "サンプルドキュメント",
		},
		{
			URI:      "gdrive:///dummy2",
			MIMEType: "application/pdf",
			Name:     "サンプルPDF",
		},
	}

	return &mcp.ListResourcesResult{
		Resources: resources,
	}, nil
}

// HandleReadResource はGoogleドライブのファイル内容を読み取るハンドラです
func (c *GoogleDriveClient) HandleReadResource(ctx context.Context, request mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	// URIからファイルIDを抽出
	fileID := strings.TrimPrefix(request.Params.URI, "gdrive:///")

	// 実際の実装では、Googleドライブ APIを使用してファイル内容を取得します
	// この実装はダミーデータを返します
	var content string
	var mimeType string

	if fileID == "dummy1" {
		content = "# サンプルドキュメント\n\nこれはサンプルのドキュメントです。"
		mimeType = "text/markdown"
	} else if fileID == "dummy2" {
		content = "サンプルPDFの内容"
		mimeType = "text/plain"
	} else {
		return nil, fmt.Errorf("file not found: %s", fileID)
	}

	textContent := mcp.TextResourceContents{
		URI:      request.Params.URI,
		MIMEType: mimeType,
		Text:     content,
	}

	// インターフェースのスライスに変換
	contents := make([]mcp.ResourceContents, 1)
	contents[0] = textContent

	return &mcp.ReadResourceResult{
		Contents: contents,
	}, nil
}

// HandleSearch はGoogleドライブのファイル検索を行うハンドラです
func (c *GoogleDriveClient) HandleSearch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 検索クエリを取得
	query, ok := request.Params.Arguments["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("search query is required")
	}

	// 実際の実装では、Googleドライブ APIを使用してファイル検索を行います
	// この実装はダミーデータを返します
	resultText := fmt.Sprintf("Found 2 files for query '%s':\nサンプルドキュメント (application/vnd.google-apps.document)\nサンプルPDF (application/pdf)", query)

	textContent := mcp.TextContent{
		Type: "text",
		Text: resultText,
	}

	// インターフェースのスライスに変換
	content := make([]mcp.Content, 1)
	content[0] = textContent

	return &mcp.CallToolResult{
		Content: content,
		IsError: false,
	}, nil
}

// SetGoogleDriveServer は受け取ったMCPサーバにGoogleドライブ用のツールを付与して、そのMCPサーバを返します。
func SetGoogleDriveServer(credentialsPath string, s *server.MCPServer) *server.MCPServer {
	// GoogleDriveクライアントを初期化
	client, err := NewGoogleDriveClient(credentialsPath)
	if err != nil {
		fmt.Printf("Error initializing Google Drive client: %v\n", err)
		return s
	}

	// 検索ツールを追加
	searchTool := mcp.NewTool("search",
		mcp.WithDescription("Search for files in Google Drive"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query"),
		),
	)
	s.AddTool(searchTool, client.HandleSearch)

	return s
}

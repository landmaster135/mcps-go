package gdrive

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	mcp "github.com/mark3labs/mcp-go/mcp"
	server "github.com/mark3labs/mcp-go/server"
)

// GoogleDriveClient はGoogleドライブAPIとやり取りするためのクライアント
type GoogleDriveClient struct {
	service *drive.Service
}

// NewGoogleDriveClient は新しいGoogleドライブクライアントを作成します
func NewGoogleDriveClient(credentialsPath string) (*GoogleDriveClient, error) {
	ctx := context.Background()

	// 認証情報を読み込む
	b, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %v", err)
	}

	// OAuth2設定を作成
	config, err := google.ConfigFromJSON(b, drive.DriveReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}

	// トークンを使用してクライアントを作成
	client := getClient(config)

	// ドライブサービスを作成
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Drive client: %v", err)
	}

	return &GoogleDriveClient{
		service: srv,
	}, nil
}

// getClient は保存されたトークンを取得し、それを使用してOAuth2クライアントを返します
func getClient(config *oauth2.Config) *http.Client {
	// トークンファイルのパス
	tokenFile := "token.json"
	tok, err := tokenFromFile(tokenFile)
	if err != nil {
		// トークンがない場合は、新しいトークンを生成
		tok = getTokenFromWeb(config)
		saveToken(tokenFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// getTokenFromWeb はウェブフローを使用して新しいトークンを取得します
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		fmt.Printf("Unable to read authorization code: %v", err)
		return nil
	}

	tok, err := config.Exchange(context.Background(), authCode)
	if err != nil {
		fmt.Printf("Unable to retrieve token from web: %v", err)
		return nil
	}
	return tok
}

// tokenFromFile はファイルからトークンを取得します
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// saveToken はトークンをファイルに保存します
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		fmt.Printf("Unable to cache oauth token: %v", err)
		return
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// HandleListResources はGoogleドライブのファイル一覧を取得するハンドラです
func (c *GoogleDriveClient) HandleListResources(ctx context.Context, request mcp.ListResourcesRequest) (*mcp.ListResourcesResult, error) {
	pageSize := int64(10) // デフォルトのページサイズ
	query := "trashed = false" // ゴミ箱に入っていないファイルのみ

	// リクエストパラメータからカーソルを取得
	var pageToken string
	if request.Params.Cursor != "" {
		pageToken = string(request.Params.Cursor)
	}

	// ファイル一覧を取得
	call := c.service.Files.List().Q(query).PageSize(pageSize).Fields("nextPageToken, files(id, name, mimeType)")
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}

	fileList, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %v", err)
	}

	// 結果をMCPリソース形式に変換
	resources := make([]mcp.Resource, 0, len(fileList.Files))
	for _, file := range fileList.Files {
		resources = append(resources, mcp.Resource{
			URI:      fmt.Sprintf("gdrive:///%s", file.Id),
			MIMEType: file.MimeType,
			Name:     file.Name,
		})
	}

	result := &mcp.ListResourcesResult{
		Resources: resources,
	}

	// ページネーションのサポート
	if fileList.NextPageToken != "" {
		// 注意: NextCursorフィールドが存在しない場合は、この部分をコメントアウトします
		// result.NextCursor = mcp.Cursor(fileList.NextPageToken)
	}

	return result, nil
}

// HandleReadResource はGoogleドライブのファイル内容を読み取るハンドラです
func (c *GoogleDriveClient) HandleReadResource(ctx context.Context, request mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	// URIからファイルIDを抽出
	fileID := strings.TrimPrefix(request.Params.URI, "gdrive:///")

	// ファイルのメタデータを取得
	file, err := c.service.Files.Get(fileID).Fields("mimeType").Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get file metadata: %v", err)
	}

	// Googleドキュメント形式のファイルの場合はエクスポート
	if strings.HasPrefix(file.MimeType, "application/vnd.google-apps") {
		var exportMimeType string

		// ファイルタイプに応じてエクスポート形式を決定
		switch file.MimeType {
		case "application/vnd.google-apps.document":
			exportMimeType = "text/markdown"
		case "application/vnd.google-apps.spreadsheet":
			exportMimeType = "text/csv"
		case "application/vnd.google-apps.presentation":
			exportMimeType = "text/plain"
		case "application/vnd.google-apps.drawing":
			exportMimeType = "image/png"
		default:
			exportMimeType = "text/plain"
		}

		// ファイルをエクスポート
		resp, err := c.service.Files.Export(fileID, exportMimeType).Download()
		if err != nil {
			return nil, fmt.Errorf("failed to export file: %v", err)
		}
		defer resp.Body.Close()

		// 内容を読み取り
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read exported content: %v", err)
		}

		textContent := mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: exportMimeType,
			Text:     string(content),
		}

		// インターフェースのスライスに変換
		contents := make([]mcp.ResourceContents, 1)
		contents[0] = textContent

		return &mcp.ReadResourceResult{
			Contents: contents,
		}, nil
	}

	// 通常のファイルの場合はダウンロード
	resp, err := c.service.Files.Get(fileID).Download()
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()

	// 内容を読み取り
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %v", err)
	}

	// MIMEタイプに応じて処理
	mimeType := file.MimeType
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// テキスト形式の場合はテキストとして返す
	if strings.HasPrefix(mimeType, "text/") || mimeType == "application/json" {
		textContent := mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: mimeType,
			Text:     string(content),
		}

		// インターフェースのスライスに変換
		contents := make([]mcp.ResourceContents, 1)
		contents[0] = textContent

		return &mcp.ReadResourceResult{
			Contents: contents,
		}, nil
	}

	// バイナリ形式の場合はBase64エンコードして返す
	blobContent := mcp.BlobResourceContents{
		URI:      request.Params.URI,
		MIMEType: mimeType,
		Blob:     string(content),
	}

	// インターフェースのスライスに変換
	contents := make([]mcp.ResourceContents, 1)
	contents[0] = blobContent

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

	// 検索クエリをエスケープ
	escapedQuery := strings.ReplaceAll(query, "\\", "\\\\")
	escapedQuery = strings.ReplaceAll(escapedQuery, "'", "\\'")
	formattedQuery := fmt.Sprintf("fullText contains '%s'", escapedQuery)

	// ファイル検索
	fileList, err := c.service.Files.List().Q(formattedQuery).PageSize(10).Fields("files(id, name, mimeType, modifiedTime, size)").Do()
	if err != nil {
		return nil, fmt.Errorf("failed to search files: %v", err)
	}

	// 結果を整形
	var fileListText strings.Builder
	fileListText.WriteString(fmt.Sprintf("Found %d files:\n", len(fileList.Files)))
	for _, file := range fileList.Files {
		fileListText.WriteString(fmt.Sprintf("%s (%s)\n", file.Name, file.MimeType))
	}

	textContent := mcp.TextContent{
		Type: "text",
		Text: fileListText.String(),
	}

	// インターフェースのスライスに変換
	content := make([]mcp.Content, 1)
	content[0] = textContent

	return &mcp.CallToolResult{
		Content: content,
		IsError: false,
	}, nil
}

// HandleResourceRead はリソース読み取りのハンドラ関数です
func (c *GoogleDriveClient) HandleResourceRead(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	result, err := c.HandleReadResource(ctx, request)
	if err != nil {
		return nil, err
	}
	return result.Contents, nil
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

	// リソース一覧のリソースを追加
	rootResource := mcp.NewResource("gdrive:///", "Google Drive",
		mcp.WithResourceDescription("Google Drive root directory"),
		mcp.WithMIMEType("application/vnd.google-apps.folder"),
	)
	s.AddResource(rootResource, client.HandleResourceRead)

	return s
}

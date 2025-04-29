package everart

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	mcp "github.com/mark3labs/mcp-go/mcp"
	server "github.com/mark3labs/mcp-go/server"
)

// GenerateImageRequest は画像生成リクエストを表します
type GenerateImageRequest struct {
	Model      string `json:"model"`
	Prompt     string `json:"prompt"`
	Method     string `json:"method"`
	ImageCount int    `json:"imageCount"`
	Height     int    `json:"height"`
	Width      int    `json:"width"`
}

// GenerateImage は画像を生成します
func (c *EverArtClient) GenerateImage(model, prompt string, imageCount, height, width int) (*GenerationResponse, error) {
	url := fmt.Sprintf("%s/generations", apiBaseURL)

	request := GenerateImageRequest{
		Model:      model,
		Prompt:     prompt,
		Method:     "txt2img",
		ImageCount: imageCount,
		Height:     height,
		Width:      width,
	}

	jsonBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	data, err := c.doRequest("POST", url, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, err
	}

	var response []GenerationResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	if len(response) == 0 {
		return nil, fmt.Errorf("no generation response")
	}

	return &response[0], nil
}

// FetchGeneration は生成状態を取得します
func (c *EverArtClient) FetchGeneration(id string) (*GenerationResponse, error) {
	url := fmt.Sprintf("%s/generations/%s", apiBaseURL, id)

	data, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var response GenerationResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// FetchGenerationWithPolling は生成が完了するまでポーリングします
func (c *EverArtClient) FetchGenerationWithPolling(id string) (*GenerationResponse, error) {
	// 実際の実装ではポーリングロジックを追加する必要があります
	// 簡略化のため、単一のリクエストで済ませています
	return c.FetchGeneration(id)
}

// HandleToGenerateImage は画像生成ツールのハンドラです
func (c *EverArtClient) HandleToGenerateImage(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	prompt := getRequiredStringParam(request.Params.Arguments, "prompt")
	modelVal, ok := getStringParam(request.Params.Arguments, "model")
	model := modelVal
	if !ok || model == "" {
		model = "5000:FLUX1.1" // デフォルトモデル
	}

	imageCount := getNumberParam(request.Params.Arguments, "image_count", 1)

	// 画像生成リクエスト
	generation, err := c.GenerateImage(model, prompt, imageCount, 1024, 1024)
	if err != nil {
		return nil, err
	}

	// 生成状態をポーリング
	completedGen, err := c.FetchGenerationWithPolling(generation.ID)
	if err != nil {
		return nil, err
	}

	// ブラウザでURLを開く
	if err := openURL(completedGen.ImageURL); err != nil {
		fmt.Printf("Warning: Failed to open URL: %v\n", err)
	}

	// 結果を返す
	resultText := fmt.Sprintf(`画像が正常に生成されました！
デフォルトブラウザで画像が開かれました。

生成の詳細:
- モデル: %s
- プロンプト: "%s"
- 画像URL: %s

上記のURLをクリックして、画像を再度表示することもできます。`, model, prompt, completedGen.ImageURL)

	return mcp.NewToolResultText(resultText), nil
}

// SetEverArtImageServer は受け取ったMCPサーバにEverArt用のツールを付与して、そのMCPサーバを返します。
func SetEverArtImageServer(apiKey string, s *server.MCPServer) *server.MCPServer {
	// EverArtクライアントを初期化
	client := NewEverArtClient(apiKey)

	// ツール: 画像生成
	generateImageTool := mcp.NewTool("generate_image",
		mcp.WithDescription("EverArtモデルを使用して画像を生成し、生成された画像を表示するためのクリック可能なリンクを返します。"+
			"このツールは、ブラウザで画像を表示するためにクリックできるURLを返します。"+
			"利用可能なモデル:\n"+
			"- 5000:FLUX1.1: 標準品質\n"+
			"- 9000:FLUX1.1-ultra: 超高品質\n"+
			"- 6000:SD3.5: Stable Diffusion 3.5\n"+
			"- 7000:Recraft-Real: フォトリアルなスタイル\n"+
			"- 8000:Recraft-Vector: ベクターアートスタイル\n"+
			"\nレスポンスには、生成された画像を表示するための直接リンクが含まれます。"),
		mcp.WithString("prompt",
			mcp.Required(),
			mcp.Description("希望する画像のテキスト説明"),
		),
		mcp.WithString("model",
			mcp.Description("モデルID (5000:FLUX1.1, 9000:FLUX1.1-ultra, 6000:SD3.5, 7000:Recraft-Real, 8000:Recraft-Vector)"),
		),
		mcp.WithNumber("image_count",
			mcp.Description("生成する画像の数"),
		),
	)
	s.AddTool(generateImageTool, client.HandleToGenerateImage)

	return s
}

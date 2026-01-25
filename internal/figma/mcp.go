package figma

import (
	"context"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	mcp "github.com/mark3labs/mcp-go/mcp"
	server "github.com/mark3labs/mcp-go/server"
)

const (
	version = "0.1.0"
)

// BuildFigmaServer はFigma MCPサーバーを構築します
func BuildFigmaServer() {
	s := createFigmaServer()

	// サーバーを起動
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func addPromptIntoServer(s *server.MCPServer) *server.MCPServer {
	prompt := mcp.NewPrompt("system_prompt_01",
		mcp.WithPromptDescription("This is a prompt for the Figma client."),
	)
	s.AddPrompt(prompt, func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return &mcp.GetPromptResult{
			Description: "System prompt for the Figma client.",
			Messages: []mcp.PromptMessage{
				{
					Role:    mcp.RoleAssistant,
					Content: mcp.NewTextContent("You use this great client for Figma well"),
				},
			},
		}, nil
	})
	return s
}

// createFigmaServer はFigma MCPサーバーを作成します
func createFigmaServer() *server.MCPServer {
	// 環境変数からFigma APIキーを取得
	apiKey := os.Getenv("FIGMA_API_KEY")
	if apiKey == "" {
		fmt.Println("Warning: FIGMA_API_KEY environment variable not set. API access will be restricted.")
	}

	// MCPサーバーを作成
	s := server.NewMCPServer(
		"Figma API Server",
		version,
		server.WithResourceCapabilities(false, false),
		server.WithPromptCapabilities(true),
		server.WithLogging(),
	)

	// Figmaクライアントを初期化
	client := NewFigmaClient(apiKey)

	// ツール1: Figmaデータの取得
	getFigmaDataTool := mcp.NewTool("get_figma_data",
		mcp.WithDescription("When the nodeId cannot be obtained, obtain the layout information about the entire Figma file"),
		mcp.WithString("fileKey",
			mcp.Required(),
			mcp.Description("The key of the Figma file to fetch, often found in a provided URL like figma.com/(file|design)/<fileKey>/..."),
		),
		mcp.WithString("nodeId",
			mcp.Description("The ID of the node to fetch, often found as URL parameter node-id=<nodeId>, always use if provided"),
		),
		mcp.WithNumber("depth",
			mcp.Description("How many levels deep to traverse the node tree, only use if explicitly requested by the user"),
		),
	)
	s.AddTool(getFigmaDataTool, handleGetFigmaData(client))

	// ツール2: Figma画像のダウンロード
	downloadFigmaImagesTool := mcp.NewTool("download_figma_images",
		mcp.WithDescription("Download SVG and PNG images used in a Figma file based on the IDs of image or icon nodes"),
		mcp.WithString("fileKey",
			mcp.Required(),
			mcp.Description("The key of the Figma file containing the node"),
		),
		mcp.WithArray("nodes",
			mcp.Required(),
			mcp.Description("The nodes to fetch as images"),
		),
		mcp.WithString("localPath",
			mcp.Required(),
			mcp.Description("The absolute path to the directory where images are stored in the project. If the directory does not exist, it will be created. The format of this path should respect the directory format of the operating system you are running on. Don't use any special character escaping in the path name either."),
		),
	)
	s.AddTool(downloadFigmaImagesTool, handleDownloadFigmaImages(client))

	// プロンプト
	s = addPromptIntoServer(s)

	return s
}

// handleGetFigmaData はFigmaデータ取得ハンドラーを返します
func handleGetFigmaData(client *FigmaClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// パラメータの取得
		fileKey := request.Params.Arguments["fileKey"].(string)
		var nodeID string
		if val, ok := request.Params.Arguments["nodeId"]; ok && val != nil {
			nodeID = val.(string)
		}

		var depth int
		if val, ok := request.Params.Arguments["depth"]; ok && val != nil {
			depth = int(val.(float64))
		}

		// ログ出力
		var depthStr, nodeStr string
		if depth > 0 {
			depthStr = fmt.Sprintf("%d layers deep", depth)
		} else {
			depthStr = "all layers"
		}

		if nodeID != "" {
			nodeStr = fmt.Sprintf("node %s", nodeID)
		} else {
			nodeStr = "full file"
		}

		fmt.Fprintf(os.Stderr, "Fetching %s of %s from file %s\n", depthStr, nodeStr, fileKey)

		var design SimplifiedDesign
		var err error

		// ノードIDが指定されている場合はノードを取得、そうでない場合はファイル全体を取得
		if nodeID != "" {
			design, err = client.GetNode(fileKey, nodeID, depth)
		} else {
			design, err = client.GetFile(fileKey, depth)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching file %s: %v\n", fileKey, err)
			return mcp.NewToolResultText(fmt.Sprintf("Error fetching file: %v", err)), nil
		}

		fmt.Fprintf(os.Stderr, "Successfully fetched file: %s\n", design.Name)

		// 結果をYAML形式に変換
		result := map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":         design.Name,
				"lastModified": design.LastModified,
				"thumbnailUrl": design.ThumbnailURL,
			},
			"nodes":      design.Nodes,
			"globalVars": design.GlobalVars,
		}

		fmt.Fprintf(os.Stderr, "Generating YAML result from file\n")
		yamlResult, err := yaml.Marshal(result)
		if err != nil {
			return nil, err
		}

		fmt.Fprintf(os.Stderr, "Sending result to client\n")
		return mcp.NewToolResultText(string(yamlResult)), nil
	}
}

// handleDownloadFigmaImages はFigma画像ダウンロードハンドラーを返します
func handleDownloadFigmaImages(client *FigmaClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// パラメータの取得
		fileKey := request.Params.Arguments["fileKey"].(string)
		localPath := request.Params.Arguments["localPath"].(string)

		// ノードの解析
		nodesArray := request.Params.Arguments["nodes"].([]interface{})
		renderRequests := make([]FetchImageParams, 0)
		imageFills := make([]FetchImageFillParams, 0)

		for _, nodeObj := range nodesArray {
			var nodeID string

			// ノードが文字列の場合（単純なノードID）
			if nodeStr, ok := nodeObj.(string); ok {
				nodeID = nodeStr
				// ファイル名をノードIDから生成
				fileName := fmt.Sprintf("node_%s.png", nodeID)

				renderRequests = append(renderRequests, FetchImageParams{
					NodeID:   nodeID,
					FileName: fileName,
					FileType: "png",
				})
			} else if nodeMap, ok := nodeObj.(map[string]interface{}); ok {
				// ノードがオブジェクトの場合（詳細情報付き）
				nodeID = nodeMap["nodeId"].(string)
				fileName := nodeMap["fileName"].(string)

				// imageRefがある場合は画像フィル、ない場合はレンダリング
				if imageRef, ok := nodeMap["imageRef"]; ok && imageRef != nil && imageRef.(string) != "" {
					imageFills = append(imageFills, FetchImageFillParams{
						NodeID:   nodeID,
						ImageRef: imageRef.(string),
						FileName: fileName,
					})
				} else {
					fileType := "png"
					if len(fileName) > 4 && fileName[len(fileName)-4:] == ".svg" {
						fileType = "svg"
					}
					renderRequests = append(renderRequests, FetchImageParams{
						NodeID:   nodeID,
						FileName: fileName,
						FileType: fileType,
					})
				}
			} else {
				return mcp.NewToolResultText("[Error] Invalid node format in request"), nil
			}
		}

		// 画像フィルのダウンロード
		var fillDownloads []string
		var err error
		if len(imageFills) > 0 {
			fillDownloads, err = client.GetImageFills(fileKey, imageFills, localPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error downloading image fills: %v\n", err)
				return mcp.NewToolResultText(fmt.Sprintf("Error downloading image fills: %v", err)), nil
			}
		}

		// レンダリング画像のダウンロード
		var renderDownloads []string
		if len(renderRequests) > 0 {
			renderDownloads, err = client.GetImages(fileKey, renderRequests, localPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error downloading rendered images: %v\n", err)
				return mcp.NewToolResultText(fmt.Sprintf("Error downloading rendered images: %v", err)), nil
			}
		}

		// 結果の結合
		downloads := append(fillDownloads, renderDownloads...)
		if len(downloads) == 0 {
			return mcp.NewToolResultText("No images were downloaded"), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Success, %d images downloaded: %s", len(downloads), downloads)), nil
	}
}

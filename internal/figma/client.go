package figma

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	apiBaseURL = "https://api.figma.com/v1"
)

// FigmaClient はFigma APIとの通信を行うクライアントです
type FigmaClient struct {
	apiKey string
}

// NewFigmaClient は新しいFigmaクライアントを作成します
func NewFigmaClient(apiKey string) *FigmaClient {
	return &FigmaClient{
		apiKey: apiKey,
	}
}

// doRequest はHTTPリクエストを実行します
func (c *FigmaClient) doRequest(method, endpoint string, body io.Reader) ([]byte, error) {
	url := fmt.Sprintf("%s%s", apiBaseURL, endpoint)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Figma-Token", c.apiKey)
	req.Header.Set("Accept", "application/json")
	if method == "POST" || method == "PATCH" || method == "PUT" {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		var figmaErr FigmaError
		if err := json.Unmarshal(respBody, &figmaErr); err != nil {
			return nil, fmt.Errorf("HTTP error: %d - %s", resp.StatusCode, string(respBody))
		}
		figmaErr.Status = resp.StatusCode
		return nil, fmt.Errorf("Figma API error: %d - %s", figmaErr.Status, figmaErr.Err)
	}

	return respBody, nil
}

// GetFile はFigmaファイルの情報を取得します
func (c *FigmaClient) GetFile(fileKey string, depth int) (SimplifiedDesign, error) {
	endpoint := fmt.Sprintf("/files/%s", fileKey)
	if depth > 0 {
		endpoint = fmt.Sprintf("%s?depth=%d", endpoint, depth)
	}

	data, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return SimplifiedDesign{}, err
	}

	var response FigmaFileResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return SimplifiedDesign{}, err
	}

	return ParseFigmaResponse(response)
}

// GetNode はFigmaノードの情報を取得します
func (c *FigmaClient) GetNode(fileKey, nodeID string, depth int) (SimplifiedDesign, error) {
	endpoint := fmt.Sprintf("/files/%s/nodes?ids=%s", fileKey, nodeID)
	if depth > 0 {
		endpoint = fmt.Sprintf("%s&depth=%d", endpoint, depth)
	}

	data, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return SimplifiedDesign{}, err
	}

	var response FigmaFileNodesResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return SimplifiedDesign{}, err
	}

	return ParseFigmaResponse(response)
}

// GetImages はFigmaの画像を取得します
func (c *FigmaClient) GetImages(fileKey string, nodes []FetchImageParams, localPath string) ([]string, error) {
	if len(nodes) == 0 {
		return []string{}, nil
	}

	// PNGとSVGのノードを分離
	pngIDs := make([]string, 0)
	svgIDs := make([]string, 0)
	nodeMap := make(map[string]FetchImageParams)

	for _, node := range nodes {
		nodeMap[node.NodeID] = node
		if node.FileType == "png" {
			pngIDs = append(pngIDs, node.NodeID)
		} else if node.FileType == "svg" {
			svgIDs = append(svgIDs, node.NodeID)
		}
	}

	// 画像URLを取得
	imageURLs := make(map[string]string)

	// PNG画像の取得
	if len(pngIDs) > 0 {
		endpoint := fmt.Sprintf("/images/%s?ids=%s&scale=2&format=png", fileKey, strings.Join(pngIDs, ","))
		data, err := c.doRequest("GET", endpoint, nil)
		if err != nil {
			return nil, err
		}

		var response FigmaImagesResponse
		if err := json.Unmarshal(data, &response); err != nil {
			return nil, err
		}

		for id, url := range response.Images {
			imageURLs[id] = url
		}
	}

	// SVG画像の取得
	if len(svgIDs) > 0 {
		endpoint := fmt.Sprintf("/images/%s?ids=%s&format=svg", fileKey, strings.Join(svgIDs, ","))
		data, err := c.doRequest("GET", endpoint, nil)
		if err != nil {
			return nil, err
		}

		var response FigmaImagesResponse
		if err := json.Unmarshal(data, &response); err != nil {
			return nil, err
		}

		for id, url := range response.Images {
			imageURLs[id] = url
		}
	}

	// 画像のダウンロード
	results := make([]string, 0, len(nodes))
	for _, node := range nodes {
		imageURL, ok := imageURLs[node.NodeID]
		if !ok {
			continue
		}

		filePath, err := downloadImage(imageURL, node.FileName, localPath)
		if err != nil {
			return nil, err
		}
		results = append(results, filePath)
	}

	return results, nil
}

// GetImageFills は画像フィルを持つノードの画像を取得します
func (c *FigmaClient) GetImageFills(fileKey string, nodes []FetchImageFillParams, localPath string) ([]string, error) {
	if len(nodes) == 0 {
		return []string{}, nil
	}

	// 画像フィルの参照IDを取得
	imageRefs := make([]string, 0, len(nodes))
	for _, node := range nodes {
		imageRefs = append(imageRefs, node.ImageRef)
	}

	// 画像URLを取得
	endpoint := fmt.Sprintf("/files/%s/images", fileKey)
	data, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response FigmaImageFillsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	// 画像のダウンロード
	results := make([]string, 0, len(nodes))
	for _, node := range nodes {
		imageURL, ok := response.Meta.Images[node.ImageRef]
		if !ok {
			continue
		}

		filePath, err := downloadImage(imageURL, node.FileName, localPath)
		if err != nil {
			return nil, err
		}
		results = append(results, filePath)
	}

	return results, nil
}

// downloadImage は画像をダウンロードします
func downloadImage(url, fileName, localPath string) (string, error) {
	// ディレクトリが存在しない場合は作成
	if err := os.MkdirAll(localPath, 0755); err != nil {
		return "", err
	}

	// HTTPリクエストを作成
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download image: %s", resp.Status)
	}

	// ファイルパスを作成
	filePath := filepath.Join(localPath, fileName)

	// ファイルを作成
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// ファイルに書き込み
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

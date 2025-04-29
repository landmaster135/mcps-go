package brave_search

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	mcp "github.com/mark3labs/mcp-go/mcp"
	server "github.com/mark3labs/mcp-go/server"
)

// レート制限の定義
var rateLimit = struct {
	perSecond int
	perMonth  int
}{
	perSecond: 1,
	perMonth:  15000,
}

// リクエストカウンターの定義
var requestCount = struct {
	second    int
	month     int
	lastReset time.Time
	mu        sync.Mutex
}{
	second:    0,
	month:     0,
	lastReset: time.Now(),
}

// レート制限をチェックする関数
func checkRateLimit() error {
	requestCount.mu.Lock()
	defer requestCount.mu.Unlock()

	now := time.Now()
	if now.Sub(requestCount.lastReset) > time.Second {
		requestCount.second = 0
		requestCount.lastReset = now
	}

	if requestCount.second >= rateLimit.perSecond || requestCount.month >= rateLimit.perMonth {
		return errors.New("rate limit exceeded")
	}

	requestCount.second++
	requestCount.month++
	return nil
}

// BraveWebResult はWeb検索結果の構造体
type BraveWebResult struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Language    string `json:"language,omitempty"`
	Published   string `json:"published,omitempty"`
	Rank        int    `json:"rank,omitempty"`
}

// BraveWebResponse はWeb検索APIレスポンスの構造体
type BraveWebResponse struct {
	Web struct {
		Results []BraveWebResult `json:"results,omitempty"`
	} `json:"web,omitempty"`
	Locations struct {
		Results []struct {
			ID    string `json:"id"`
			Title string `json:"title,omitempty"`
		} `json:"results,omitempty"`
	} `json:"locations,omitempty"`
}

// BraveLocation はローカル検索結果の構造体
type BraveLocation struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Address struct {
		StreetAddress   string `json:"streetAddress,omitempty"`
		AddressLocality string `json:"addressLocality,omitempty"`
		AddressRegion   string `json:"addressRegion,omitempty"`
		PostalCode      string `json:"postalCode,omitempty"`
	} `json:"address"`
	Coordinates struct {
		Latitude  float64 `json:"latitude,omitempty"`
		Longitude float64 `json:"longitude,omitempty"`
	} `json:"coordinates,omitempty"`
	Phone  string `json:"phone,omitempty"`
	Rating struct {
		RatingValue float64 `json:"ratingValue,omitempty"`
		RatingCount int     `json:"ratingCount,omitempty"`
	} `json:"rating,omitempty"`
	OpeningHours []string `json:"openingHours,omitempty"`
	PriceRange   string   `json:"priceRange,omitempty"`
}

// BravePoiResponse はPOI検索APIレスポンスの構造体
type BravePoiResponse struct {
	Results []BraveLocation `json:"results"`
}

// BraveDescription は説明APIレスポンスの構造体
type BraveDescription struct {
	Descriptions map[string]string `json:"descriptions"`
}

// Web検索を実行する関数
func performWebSearch(query string, count int, offset int) (string, error) {
	if err := checkRateLimit(); err != nil {
		return "", err
	}

	apiKey := os.Getenv("BRAVE_API_KEY")
	if apiKey == "" {
		return "", errors.New("BRAVE_API_KEY environment variable is required")
	}

	// URLの構築
	baseURL := "https://api.search.brave.com/res/v1/web/search"
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	// クエリパラメータの設定
	q := u.Query()
	q.Set("q", query)
	if count > 20 {
		count = 20 // API制限
	}
	q.Set("count", fmt.Sprintf("%d", count))
	q.Set("offset", fmt.Sprintf("%d", offset))
	u.RawQuery = q.Encode()

	// リクエストの作成
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return "", err
	}

	// ヘッダーの設定
	req.Header.Add("Accept", "application/json")
	// gzipヘッダーを削除してJSONパースエラーを回避
	req.Header.Add("X-Subscription-Token", apiKey)

	// リクエストの実行
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("error of Brave API: %d %s\n%s", resp.StatusCode, resp.Status, string(body))
	}

	// レスポンスの解析
	var data BraveWebResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	// 結果のフォーマット
	var results []string
	for _, result := range data.Web.Results {
		formattedResult := fmt.Sprintf("Title: %s\nDescription: %s\nURL: %s",
			result.Title, result.Description, result.URL)
		results = append(results, formattedResult)
	}

	return strings.Join(results, "\n\n"), nil
}

// ローカル検索を実行する関数
func performLocalSearch(query string, count int) (string, error) {
	if err := checkRateLimit(); err != nil {
		return "", err
	}

	apiKey := os.Getenv("BRAVE_API_KEY")
	if apiKey == "" {
		return "", errors.New("BRAVE_API_KEY environment variable is required")
	}

	// 最初の検索でロケーションIDを取得
	webURL, err := url.Parse("https://api.search.brave.com/res/v1/web/search")
	if err != nil {
		return "", err
	}

	q := webURL.Query()
	q.Set("q", query)
	q.Set("search_lang", "en")
	q.Set("result_filter", "locations")
	if count > 20 {
		count = 20 // API制限
	}
	q.Set("count", fmt.Sprintf("%d", count))
	webURL.RawQuery = q.Encode()

	// リクエストの作成
	webReq, err := http.NewRequest("GET", webURL.String(), nil)
	if err != nil {
		return "", err
	}

	// ヘッダーの設定
	webReq.Header.Add("Accept", "application/json")
	// gzipヘッダーを削除してJSONパースエラーを回避
	webReq.Header.Add("X-Subscription-Token", apiKey)

	// リクエストの実行
	client := &http.Client{}
	webResp, err := client.Do(webReq)
	if err != nil {
		return "", err
	}
	defer webResp.Body.Close()

	if webResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(webResp.Body)
		return "", fmt.Errorf("error of Brave API: %d %s\n%s", webResp.StatusCode, webResp.Status, string(body))
	}

	// レスポンスの解析
	var webData BraveWebResponse
	if err := json.NewDecoder(webResp.Body).Decode(&webData); err != nil {
		return "", err
	}

	// ロケーションIDの抽出
	var locationIDs []string
	for _, result := range webData.Locations.Results {
		if result.ID != "" {
			locationIDs = append(locationIDs, result.ID)
		}
	}

	// ロケーションIDがない場合はWeb検索にフォールバック
	if len(locationIDs) == 0 {
		return performWebSearch(query, count, 0)
	}

	// POIデータと説明を並行して取得
	poisCh := make(chan BravePoiResponse, 1)
	descCh := make(chan BraveDescription, 1)
	errCh := make(chan error, 2)

	// POIデータの取得
	go func() {
		pois, err := getPoisData(locationIDs)
		if err != nil {
			errCh <- err
			return
		}
		poisCh <- pois
	}()

	// 説明データの取得
	go func() {
		desc, err := getDescriptionsData(locationIDs)
		if err != nil {
			errCh <- err
			return
		}
		descCh <- desc
	}()

	// 結果の待機
	var poisData BravePoiResponse
	var descData BraveDescription
	for i := 0; i < 2; i++ {
		select {
		case err := <-errCh:
			return "", err
		case poisData = <-poisCh:
		case descData = <-descCh:
		}
	}

	// 結果のフォーマット
	return formatLocalResults(poisData, descData), nil
}

// POIデータを取得する関数
func getPoisData(ids []string) (BravePoiResponse, error) {
	if err := checkRateLimit(); err != nil {
		return BravePoiResponse{}, err
	}

	apiKey := os.Getenv("BRAVE_API_KEY")
	if apiKey == "" {
		return BravePoiResponse{}, errors.New("BRAVE_API_KEY environment variable is required")
	}

	// URLの構築
	u, err := url.Parse("https://api.search.brave.com/res/v1/local/pois")
	if err != nil {
		return BravePoiResponse{}, err
	}

	// クエリパラメータの設定
	q := u.Query()
	for _, id := range ids {
		if id != "" {
			q.Add("ids", id)
		}
	}
	u.RawQuery = q.Encode()

	// リクエストの作成
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return BravePoiResponse{}, err
	}

	// ヘッダーの設定
	req.Header.Add("Accept", "application/json")
	// gzipヘッダーを削除してJSONパースエラーを回避
	req.Header.Add("X-Subscription-Token", apiKey)

	// リクエストの実行
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return BravePoiResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return BravePoiResponse{}, fmt.Errorf("error of Brave API: %d %s\n%s", resp.StatusCode, resp.Status, string(body))
	}

	// レスポンスの解析
	var poisResponse BravePoiResponse
	if err := json.NewDecoder(resp.Body).Decode(&poisResponse); err != nil {
		return BravePoiResponse{}, err
	}

	return poisResponse, nil
}

// 説明データを取得する関数
func getDescriptionsData(ids []string) (BraveDescription, error) {
	if err := checkRateLimit(); err != nil {
		return BraveDescription{}, err
	}

	apiKey := os.Getenv("BRAVE_API_KEY")
	if apiKey == "" {
		return BraveDescription{}, errors.New("BRAVE_API_KEY environment variable is required")
	}

	// URLの構築
	u, err := url.Parse("https://api.search.brave.com/res/v1/local/descriptions")
	if err != nil {
		return BraveDescription{}, err
	}

	// クエリパラメータの設定
	q := u.Query()
	for _, id := range ids {
		if id != "" {
			q.Add("ids", id)
		}
	}
	u.RawQuery = q.Encode()

	// リクエストの作成
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return BraveDescription{}, err
	}

	// ヘッダーの設定
	req.Header.Add("Accept", "application/json")
	// gzipヘッダーを削除してJSONパースエラーを回避
	req.Header.Add("X-Subscription-Token", apiKey)

	// リクエストの実行
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return BraveDescription{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return BraveDescription{}, fmt.Errorf("error of Brave API: %d %s\n%s", resp.StatusCode, resp.Status, string(body))
	}

	// レスポンスの解析
	var descData BraveDescription
	if err := json.NewDecoder(resp.Body).Decode(&descData); err != nil {
		return BraveDescription{}, err
	}

	return descData, nil
}

// ローカル検索結果をフォーマットする関数
func formatLocalResults(poisData BravePoiResponse, descData BraveDescription) string {
	if len(poisData.Results) == 0 {
		return "No local results found"
	}

	var formattedResults []string
	for _, poi := range poisData.Results {
		// 住所のフォーマット
		addressParts := []string{
			poi.Address.StreetAddress,
			poi.Address.AddressLocality,
			poi.Address.AddressRegion,
			poi.Address.PostalCode,
		}
		var addressParts2 []string
		for _, part := range addressParts {
			if part != "" {
				addressParts2 = append(addressParts2, part)
			}
		}
		address := "N/A"
		if len(addressParts2) > 0 {
			address = strings.Join(addressParts2, ", ")
		}

		// 営業時間のフォーマット
		hours := "N/A"
		if len(poi.OpeningHours) > 0 {
			hours = strings.Join(poi.OpeningHours, ", ")
		}

		// 評価のフォーマット
		rating := "N/A"
		if poi.Rating.RatingValue > 0 {
			rating = fmt.Sprintf("%.1f (%d reviews)", poi.Rating.RatingValue, poi.Rating.RatingCount)
		}

		// 説明の取得
		description := "No description available"
		if desc, ok := descData.Descriptions[poi.ID]; ok && desc != "" {
			description = desc
		}

		// 結果のフォーマット
		formattedResult := fmt.Sprintf(`Name: %s
Address: %s
Phone: %s
Rating: %s
Price Range: %s
Hours: %s
Description: %s`,
			poi.Name, address, getOrDefault(poi.Phone, "N/A"), rating,
			getOrDefault(poi.PriceRange, "N/A"), hours, description)

		formattedResults = append(formattedResults, formattedResult)
	}

	return strings.Join(formattedResults, "\n---\n")
}

// 空文字列の場合にデフォルト値を返すヘルパー関数
func getOrDefault(value string, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

func performTest(query string, count int, offset int) (string, error) {
	return "performed", nil
}

func addPromptIntoServer(s *server.MCPServer) *server.MCPServer {
	prompt := mcp.NewPrompt("system_prompt_01",
		mcp.WithPromptDescription("This is a search engine prompt"),
	)
	s.AddPrompt(prompt, func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return &mcp.GetPromptResult{
			Description: "System prompt for search engine.",
			Messages: []mcp.PromptMessage{
				{
					Role:    mcp.RoleAssistant,
					Content: mcp.NewTextContent("You use this search engine well."),
				},
			},
		}, nil
	})
	return s
}

// MCPサーバを構築する関数
func BuildBraveSearchServer() {
	// APIキーのチェック
	apiKey := os.Getenv("BRAVE_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Error: BRAVE_API_KEY environment variable is required")
		os.Exit(1)
	}

	// サーバの作成
	s := server.NewMCPServer(
		"Brave Search",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
		server.WithLogging(),
	)

	// Web検索ツールの定義
	webSearchTool := mcp.NewTool("brave_web_search",
		mcp.WithDescription("Performs a web search using the Brave Search API, ideal for general queries, news, articles, and online content. "+
			"Use this for broad information gathering, recent events, or when you need diverse web sources. "+
			"Supports pagination, content filtering, and freshness controls. "+
			"Maximum 20 results per request, with offset for pagination."),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query (max 400 chars, 50 words)"),
		),
		mcp.WithNumber("count",
			mcp.Description("Number of results (1-20, default 10)"),
		),
		mcp.WithNumber("offset",
			mcp.Description("Pagination offset (max 9, default 0)"),
		),
	)

	// ローカル検索ツールの定義
	localSearchTool := mcp.NewTool("brave_local_search",
		mcp.WithDescription("Searches for local businesses and places using Brave's Local Search API. "+
			"Best for queries related to physical locations, businesses, restaurants, services, etc. "+
			"Returns detailed information including:\n"+
			"- Business names and addresses\n"+
			"- Ratings and review counts\n"+
			"- Phone numbers and opening hours\n"+
			"Use this when the query implies 'near me' or mentions specific locations. "+
			"Automatically falls back to web search if no local results are found."),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Local search query (e.g. 'pizza near Central Park')"),
		),
		mcp.WithNumber("count",
			mcp.Description("Number of results (1-20, default 5)"),
		),
	)

	// Web検索ツールのハンドラ
	s.AddTool(webSearchTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// 引数の取得
		query, ok := request.Params.Arguments["query"].(string)
		if !ok {
			return nil, errors.New("query parameter is required and must be a string")
		}

		// オプションパラメータの取得
		count := 10
		if countArg, ok := request.Params.Arguments["count"].(float64); ok {
			count = int(countArg)
		}

		offset := 0
		if offsetArg, ok := request.Params.Arguments["offset"].(float64); ok {
			offset = int(offsetArg)
		}

		// Web検索の実行
		// results, err := performTest(query, count, offset)
		results, err := performWebSearch(query, count, offset)
		if err != nil {
			return nil, err
		}

		// 結果の返却
		return mcp.NewToolResultText(results), nil
	})

	// ローカル検索ツールのハンドラ
	s.AddTool(localSearchTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// 引数の取得
		query, ok := request.Params.Arguments["query"].(string)
		if !ok {
			return nil, errors.New("query parameter is required and must be a string")
		}

		// オプションパラメータの取得
		count := 5
		if countArg, ok := request.Params.Arguments["count"].(float64); ok {
			count = int(countArg)
		}

		// ローカル検索の実行
		results, err := performLocalSearch(query, count)
		if err != nil {
			return nil, err
		}

		// 結果の返却
		return mcp.NewToolResultText(results), nil
	})

	// プロンプト
	s = addPromptIntoServer(s)

	// サーバの起動
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

package duckduckgo_search

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	mcp "github.com/mark3labs/mcp-go/mcp"
	server "github.com/mark3labs/mcp-go/server"
)

// レート制限の定義
var rateLimit = struct {
	perSecond int
	perMinute int
}{
	perSecond: 1,
	perMinute: 30,
}

// リクエストカウンターの定義
var requestCount = struct {
	second    int
	minute    int
	lastReset time.Time
	mu        sync.Mutex
}{
	second:    0,
	minute:    0,
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

	if now.Sub(requestCount.lastReset) > time.Minute {
		requestCount.minute = 0
	}

	if requestCount.second >= rateLimit.perSecond || requestCount.minute >= rateLimit.perMinute {
		return errors.New("rate limit exceeded")
	}

	requestCount.second++
	requestCount.minute++
	return nil
}

// DuckDuckGoWebResult はWeb検索結果の構造体
type DuckDuckGoWebResult struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Rank        int    `json:"rank"`
}

// DuckDuckGoResponse はWeb検索APIレスポンスの構造体
type DuckDuckGoResponse struct {
	Results []DuckDuckGoWebResult `json:"results"`
}

// Web検索を実行する関数
func performWebSearch(query string, count int, offset int) (string, error) {
	if err := checkRateLimit(); err != nil {
		return "", err
	}

	// URLの構築 - DuckDuckGoのHTML版を使用
	baseURL := "https://html.duckduckgo.com/html/"
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	// クエリパラメータの設定
	q := u.Query()
	q.Set("q", query)
	if count > 50 {
		count = 50 // 制限
	}

	// オフセットは実装が難しいため、最初のページのみサポート
	u.RawQuery = q.Encode()

	// リクエストの作成
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return "", err
	}

	// ヘッダーの設定 - ブラウザのように見せかける
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")

	// リクエストの実行
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("error from DuckDuckGo: %d %s\n%s", resp.StatusCode, resp.Status, string(body))
	}

	// レスポンスの解析
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// HTMLから検索結果を抽出
	results, err := parseSearchResults(string(body), count)
	if err != nil {
		return "", err
	}

	// 結果のフォーマット
	var formattedResults []string
	for i, result := range results {
		if i >= count {
			break
		}
		formattedResult := fmt.Sprintf("Title: %s\nDescription: %s\nURL: %s",
			result.Title, result.Description, result.URL)
		formattedResults = append(formattedResults, formattedResult)
	}

	if len(formattedResults) == 0 {
		return "No search results found", nil
	}

	return strings.Join(formattedResults, "\n\n"), nil
}

// HTMLから検索結果を抽出する関数
func parseSearchResults(html string, maxResults int) ([]DuckDuckGoWebResult, error) {
	var results []DuckDuckGoWebResult

	// DuckDuckGoのHTML構造に基づく正規表現パターン
	// 結果のリンクを抽出
	linkPattern := `<a class="result__a" href="([^"]+)"[^>]*>([^<]+)</a>`
	linkRegex := regexp.MustCompile(linkPattern)
	linkMatches := linkRegex.FindAllStringSubmatch(html, -1)

	// 結果の説明を抽出
	descPattern := `<a class="result__snippet"[^>]*>([^<]+)</a>`
	descRegex := regexp.MustCompile(descPattern)
	descMatches := descRegex.FindAllStringSubmatch(html, -1)

	// より柔軟なパターンも試す
	if len(linkMatches) == 0 {
		// 代替パターン1
		linkPattern = `<h2[^>]*><a[^>]+href="([^"]+)"[^>]*>([^<]+)</a></h2>`
		linkRegex = regexp.MustCompile(linkPattern)
		linkMatches = linkRegex.FindAllStringSubmatch(html, -1)
	}

	if len(linkMatches) == 0 {
		// 代替パターン2 - より汎用的なリンクパターン
		linkPattern = `<a[^>]+href="([^"]+)"[^>]*>([^<]+)</a>`
		linkRegex = regexp.MustCompile(linkPattern)
		linkMatches = linkRegex.FindAllStringSubmatch(html, -1)

		// フィルタリング - 明らかに検索結果ではないリンクを除外
		var filteredMatches [][]string
		for _, match := range linkMatches {
			url := match[1]
			title := match[2]
			// 内部リンクや画像リンクを除外
			if !strings.Contains(url, "duckduckgo.com") &&
			   !strings.HasPrefix(url, "#") &&
			   !strings.HasPrefix(url, "javascript:") &&
			   len(title) > 5 { // タイトルが短すぎるものを除外
				filteredMatches = append(filteredMatches, match)
			}
		}
		linkMatches = filteredMatches
	}

	// 結果を構築
	for i, linkMatch := range linkMatches {
		if i >= maxResults {
			break
		}

		url := linkMatch[1]
		title := linkMatch[2]

		// URLのデコード
		decodedURL, err := decodeURL(url)
		if err == nil {
			url = decodedURL
		}

		// タイトルのクリーンアップ
		title = cleanText(title)

		// 説明の取得（可能な場合）
		description := ""
		if i < len(descMatches) {
			description = cleanText(descMatches[i][1])
		}

		result := DuckDuckGoWebResult{
			Title:       title,
			Description: description,
			URL:         url,
			Rank:        i + 1,
		}

		results = append(results, result)
	}

	return results, nil
}

// URLをデコードする関数
func decodeURL(rawURL string) (string, error) {
	// DuckDuckGoのリダイレクトURLの場合、実際のURLを抽出
	if strings.Contains(rawURL, "uddg=") {
		u, err := url.Parse(rawURL)
		if err != nil {
			return rawURL, err
		}

		uddg := u.Query().Get("uddg")
		if uddg != "" {
			decoded, err := url.QueryUnescape(uddg)
			if err == nil {
				return decoded, nil
			}
		}
	}

	// 通常のURL
	decoded, err := url.QueryUnescape(rawURL)
	if err != nil {
		return rawURL, nil
	}
	return decoded, nil
}

// テキストをクリーンアップする関数
func cleanText(text string) string {
	// HTMLエンティティのデコード
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")

	// HTMLタグの除去
	tagRegex := regexp.MustCompile(`<[^>]*>`)
	text = tagRegex.ReplaceAllString(text, "")

	// 余分な空白の除去
	text = strings.TrimSpace(text)
	spaceRegex := regexp.MustCompile(`\s+`)
	text = spaceRegex.ReplaceAllString(text, " ")

	return text
}

// 簡易検索を実行する関数（Instant Answer API使用）
func performInstantSearch(query string) (string, error) {
	if err := checkRateLimit(); err != nil {
		return "", err
	}

	// DuckDuckGo Instant Answer API
	baseURL := "https://api.duckduckgo.com/"
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	// クエリパラメータの設定
	q := u.Query()
	q.Set("q", query)
	q.Set("format", "json")
	q.Set("no_html", "1")
	q.Set("skip_disambig", "1")
	u.RawQuery = q.Encode()

	// リクエストの作成
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return "", err
	}

	// ヘッダーの設定
	req.Header.Set("User-Agent", "DuckDuckGo-MCP-Server/1.0")
	req.Header.Set("Accept", "application/json")

	// リクエストの実行
	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("error from DuckDuckGo API: %d %s\n%s", resp.StatusCode, resp.Status, string(body))
	}

	// レスポンスの解析
	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	// 結果のフォーマット
	var results []string

	// Abstract（要約）
	if abstract, ok := data["Abstract"].(string); ok && abstract != "" {
		results = append(results, fmt.Sprintf("Abstract: %s", abstract))

		if source, ok := data["AbstractSource"].(string); ok && source != "" {
			results = append(results, fmt.Sprintf("Source: %s", source))
		}

		if url, ok := data["AbstractURL"].(string); ok && url != "" {
			results = append(results, fmt.Sprintf("URL: %s", url))
		}
	}

	// Definition（定義）
	if definition, ok := data["Definition"].(string); ok && definition != "" {
		results = append(results, fmt.Sprintf("Definition: %s", definition))

		if source, ok := data["DefinitionSource"].(string); ok && source != "" {
			results = append(results, fmt.Sprintf("Source: %s", source))
		}
	}

	// Answer（回答）
	if answer, ok := data["Answer"].(string); ok && answer != "" {
		results = append(results, fmt.Sprintf("Answer: %s", answer))

		if answerType, ok := data["AnswerType"].(string); ok && answerType != "" {
			results = append(results, fmt.Sprintf("Type: %s", answerType))
		}
	}

	// Related Topics（関連トピック）
	if relatedTopics, ok := data["RelatedTopics"].([]interface{}); ok && len(relatedTopics) > 0 {
		results = append(results, "\nRelated Topics:")
		for i, topic := range relatedTopics {
			if i >= 3 { // 最初の3つのみ
				break
			}
			if topicMap, ok := topic.(map[string]interface{}); ok {
				if text, ok := topicMap["Text"].(string); ok && text != "" {
					results = append(results, fmt.Sprintf("- %s", text))
				}
			}
		}
	}

	if len(results) == 0 {
		return "No instant answer found", nil
	}

	return strings.Join(results, "\n"), nil
}

func addPromptIntoServer(s *server.MCPServer) *server.MCPServer {
	prompt := mcp.NewPrompt("system_prompt_01",
		mcp.WithPromptDescription("This is a DuckDuckGo search engine prompt"),
	)
	s.AddPrompt(prompt, func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return &mcp.GetPromptResult{
			Description: "System prompt for DuckDuckGo search engine.",
			Messages: []mcp.PromptMessage{
				{
					Role:    mcp.RoleAssistant,
					Content: mcp.NewTextContent("You use this DuckDuckGo search engine well for privacy-focused search results."),
				},
			},
		}, nil
	})
	return s
}

// MCPサーバを構築する関数
func BuildDuckDuckGoSearchServer() {
	// サーバの作成
	s := server.NewMCPServer(
		"DuckDuckGo Search",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
		server.WithLogging(),
	)

	// Web検索ツールの定義
	webSearchTool := mcp.NewTool("duckduckgo_web_search",
		mcp.WithDescription("Performs a web search using DuckDuckGo search engine, focusing on privacy and unbiased results. "+
			"Use this for general information gathering, news, articles, and online content when you need privacy-focused search results. "+
			"Returns organic search results without tracking or personalization. "+
			"Maximum 20 results per request."),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query (any keywords you would use in DuckDuckGo search)"),
		),
		mcp.WithNumber("count",
			mcp.Description("Number of results (1-20, default 10)"),
		),
	)

	// Instant Answer検索ツールの定義
	instantSearchTool := mcp.NewTool("duckduckgo_instant_search",
		mcp.WithDescription("Performs an instant answer search using DuckDuckGo's Instant Answer API. "+
			"Best for quick facts, definitions, calculations, and direct answers. "+
			"Returns structured information including abstracts, definitions, and related topics. "+
			"Use this when you need immediate factual information rather than web page results."),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query for instant answers (e.g., 'weather Tokyo', 'what is Python', '2+2')"),
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

		// Web検索の実行
		results, err := performWebSearch(query, count, 0)
		if err != nil {
			return nil, err
		}

		// 結果の返却
		return mcp.NewToolResultText(results), nil
	})

	// Instant Answer検索ツールのハンドラ
	s.AddTool(instantSearchTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// 引数の取得
		query, ok := request.Params.Arguments["query"].(string)
		if !ok {
			return nil, errors.New("query parameter is required and must be a string")
		}

		// Instant Answer検索の実行
		results, err := performInstantSearch(query)
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

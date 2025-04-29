package youtube_transcript

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	mcp "github.com/mark3labs/mcp-go/mcp"
	server "github.com/mark3labs/mcp-go/server"
)

// エラーメッセージの定数
const (
	ErrNoSubtitles       = "この動画には字幕が存在しません"
	ErrLanguageNotFound  = "指定された言語の字幕が見つかりませんでした"
	ErrVideoUnavailable  = "動画が利用できないか、非公開になっています"
	ErrNetworkIssue      = "ネットワーク接続に問題があります"
	ErrParsingFailed     = "字幕データの解析に失敗しました"
	ErrYouTubeAPIChanged = "YouTubeのAPIが変更された可能性があります"
)

// TranscriptLine は字幕の1行を表す構造体です
type TranscriptLine struct {
	Start float64 `json:"start"`
	Dur   float64 `json:"dur"`
	Text  string  `json:"text"`
}

// XMLTranscript は字幕XMLをパースするための構造体です
type XMLTranscript struct {
	XMLName xml.Name  `xml:"transcript"`
	Texts   []XMLText `xml:"text"`
}

// XMLText は字幕XMLの各テキスト要素をパースするための構造体です
type XMLText struct {
	Start string `xml:"start,attr"`
	Dur   string `xml:"dur,attr"`
	Value string `xml:",chardata"`
}

// YouTubeTranscriptService は字幕取得サービスを提供する構造体です
type YouTubeTranscriptService struct {
	httpClient *http.Client
}

// NewYouTubeTranscriptService は新しいYouTubeTranscriptServiceを作成します
func NewYouTubeTranscriptService() *YouTubeTranscriptService {
	return &YouTubeTranscriptService{
		httpClient: &http.Client{},
	}
}

// ExtractYoutubeID はYouTube URLまたはIDから動画IDを抽出します
func (s *YouTubeTranscriptService) ExtractYoutubeID(input string) (string, error) {
	if input == "" {
		return "", errors.New("YouTube URL or ID is required")
	}

	// URLの場合
	if strings.HasPrefix(input, "http") {
		parsedURL, err := url.Parse(input)
		if err != nil {
			return "", fmt.Errorf("invalid URL: %v", err)
		}

		// youtu.be形式
		if parsedURL.Host == "youtu.be" {
			return strings.TrimPrefix(parsedURL.Path, "/"), nil
		}

		// youtube.com形式
		if strings.Contains(parsedURL.Host, "youtube.com") {
			query := parsedURL.Query()
			videoID := query.Get("v")
			if videoID == "" {
				return "", errors.New("invalid YouTube URL: missing video ID")
			}
			return videoID, nil
		}

		return "", errors.New("unsupported YouTube URL format")
	}

	// 直接IDの場合（11文字の英数字とハイフン、アンダースコア）
	idRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]{11}$`)
	if idRegex.MatchString(input) {
		return input, nil
	}

	return "", errors.New("invalid YouTube video ID format")
}

// GetTranscript はYouTube動画の字幕を取得します
func (s *YouTubeTranscriptService) GetTranscript(videoID, lang string) ([]TranscriptLine, error) {
	// コンテキストとタイムアウトを設定（パフォーマンス改善）
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 動画ページを取得
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://youtube.com/watch?v=%s", videoID), nil)
	if err != nil {
		return nil, fmt.Errorf("リクエスト作成に失敗しました: %v", err)
	}

	// ユーザーエージェントを設定（YouTubeのAPI変更対策）
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", ErrNetworkIssue, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("%s (ID: %s)", ErrVideoUnavailable, videoID)
		}
		return nil, fmt.Errorf("HTTPエラー: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("レスポンスの読み取りに失敗しました: %v", err)
	}

	htmlContent := string(body)

	// 字幕トラック情報を抽出（複数のパターンを試行）
	captionTracksJSON, err := extractCaptionTracks(htmlContent)
	if err != nil {
		return nil, err
	}

	// 字幕トラックから指定言語のものを見つける
	var captionData struct {
		CaptionTracks []struct {
			BaseURL string `json:"baseUrl"`
			VssID   string `json:"vssId"`
			Name    struct {
				SimpleText string `json:"simpleText"`
			} `json:"name"`
			LanguageCode string `json:"languageCode"`
		} `json:"captionTracks"`
	}

	if err := json.Unmarshal([]byte(captionTracksJSON), &captionData); err != nil {
		return nil, fmt.Errorf("%s: %v", ErrParsingFailed, err)
	}

	if len(captionData.CaptionTracks) == 0 {
		return nil, errors.New(ErrNoSubtitles)
	}

	// 指定言語の字幕を探す
	var subtitleURL string
	var availableLanguages []string

	for _, track := range captionData.CaptionTracks {
		availableLanguages = append(availableLanguages, track.LanguageCode)
		if strings.Contains(track.VssID, "."+lang) || track.LanguageCode == lang {
			subtitleURL = track.BaseURL
			break
		}
	}

	if subtitleURL == "" {
		// 利用可能な言語のリストを提供
		return nil, fmt.Errorf("%s (指定: %s, 利用可能: %s)", ErrLanguageNotFound, lang, strings.Join(availableLanguages, ", "))
	}

	// 字幕XMLを取得
	subtitleReq, err := http.NewRequestWithContext(ctx, "GET", subtitleURL, nil)
	if err != nil {
		return nil, fmt.Errorf("字幕リクエスト作成に失敗しました: %v", err)
	}

	subtitleResp, err := s.httpClient.Do(subtitleReq)
	if err != nil {
		return nil, fmt.Errorf("字幕データの取得に失敗しました: %v", err)
	}
	defer subtitleResp.Body.Close()

	// XMLをパース
	var transcript XMLTranscript
	if err := xml.NewDecoder(subtitleResp.Body).Decode(&transcript); err != nil {
		return nil, fmt.Errorf("%s: %v", ErrParsingFailed, err)
	}

	// 字幕データを構造化（バッファサイズを事前に確保してパフォーマンス改善）
	lines := make([]TranscriptLine, 0, len(transcript.Texts))
	for _, text := range transcript.Texts {
		var start, dur float64
		fmt.Sscanf(text.Start, "%f", &start)
		fmt.Sscanf(text.Dur, "%f", &dur)

		// HTMLエンティティをデコード
		decodedText := html.UnescapeString(text.Value)
		// HTMLタグを削除
		cleanText := stripHTMLTags(decodedText)

		lines = append(lines, TranscriptLine{
			Start: start,
			Dur:   dur,
			Text:  cleanText,
		})
	}

	return lines, nil
}

// extractCaptionTracks は複数のパターンを試して字幕トラック情報を抽出します
func extractCaptionTracks(htmlContent string) (string, error) {
	// パターン1: 標準的なパターン
	pattern1 := regexp.MustCompile(`"captionTracks":(\[.*?\])`)
	if matches := pattern1.FindStringSubmatch(htmlContent); len(matches) >= 2 {
		return fmt.Sprintf(`{"captionTracks":%s}`, matches[1]), nil
	}

	// パターン2: 別の可能性のあるパターン
	pattern2 := regexp.MustCompile(`"captions":\s*{.*?"playerCaptionsTracklistRenderer":\s*{.*?"captionTracks":\s*(\[.*?\])`)
	if matches := pattern2.FindStringSubmatch(htmlContent); len(matches) >= 2 {
		return fmt.Sprintf(`{"captionTracks":%s}`, matches[1]), nil
	}

	// パターン3: さらに別のパターン
	pattern3 := regexp.MustCompile(`"playerCaptionsTracklistRenderer":\s*{.*?"captionTracks":\s*(\[.*?\])`)
	if matches := pattern3.FindStringSubmatch(htmlContent); len(matches) >= 2 {
		return fmt.Sprintf(`{"captionTracks":%s}`, matches[1]), nil
	}

	// 字幕が見つからない場合
	if strings.Contains(htmlContent, "\"playabilityStatus\":{\"status\":\"ERROR\"") {
		return "", errors.New(ErrVideoUnavailable)
	}

	// 字幕トラックが見つからない場合
	return "", errors.New(ErrNoSubtitles)
}

// FormatTranscript は字幕データを整形して文字列として返します
func (s *YouTubeTranscriptService) FormatTranscript(lines []TranscriptLine) string {
	var texts []string
	for _, line := range lines {
		if strings.TrimSpace(line.Text) != "" {
			texts = append(texts, line.Text)
		}
	}
	return strings.Join(texts, " ")
}

// stripHTMLTags はHTMLタグを削除します
func stripHTMLTags(input string) string {
	// 簡易的なHTMLタグ削除（より高度な実装が必要な場合はHTML解析ライブラリを使用）
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(input, "")
}

// BuildYouTubeTranscriptServer はYouTube字幕取得MCPサーバーを構築します
func BuildYouTubeTranscriptServer() {
	s := server.NewMCPServer(
		"YouTube Transcript Service",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
	)

	// get_transcript ツールの設定
	tool := mcp.NewTool("get_youtube_transcript",
		mcp.WithDescription("Extract transcript from a YouTube video URL or ID"),
		mcp.WithString("url",
			mcp.Required(),
			mcp.Description("YouTube video URL or ID"),
		),
		mcp.WithString("lang",
			mcp.Description("Language code for transcript (e.g., 'ko', 'en')"),
		),
	)

	transcriptService := NewYouTubeTranscriptService()

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// タイムアウト付きコンテキストを作成（パフォーマンス改善）
		ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
		defer cancel()

		// パラメータの取得
		input := request.Params.Arguments["url"].(string)
		lang := "en" // デフォルト値
		if langArg, ok := request.Params.Arguments["lang"]; ok {
			lang = langArg.(string)
		}

		// 動画IDの抽出
		videoID, err := transcriptService.ExtractYoutubeID(input)
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("エラー: %v", err)), nil
		}

		// 字幕の取得
		lines, err := transcriptService.GetTranscript(videoID, lang)
		if err != nil {
			// エラーメッセージを具体化
			errorMsg := fmt.Sprintf("字幕取得エラー: %v\n\n", err)
			errorMsg += "考えられる解決策:\n"
			errorMsg += "- 別の言語コードを試してください（例: 'en', 'ja', 'ko', 'es'）\n"
			errorMsg += "- 動画に字幕があるか確認してください\n"
			errorMsg += "- 動画が公開されていて、アクセス可能か確認してください\n"
			errorMsg += "- URLが正しいか確認してください\n"
			return mcp.NewToolResultText(errorMsg), nil
		}

		// 字幕の整形
		transcript := transcriptService.FormatTranscript(lines)

		// 字幕が空の場合のチェック
		if len(transcript) == 0 {
			return mcp.NewToolResultText("エラー: 字幕が空です。この動画には字幕がないか、取得できませんでした。"), nil
		}

		// メタデータをJSON形式に変換
		metadataJSON, err := json.MarshalIndent(map[string]interface{}{
			"videoId":     videoID,
			"language":    lang,
			"timestamp":   time.Now().Format(time.RFC3339),
			"charCount":   len(transcript),
			"lineCount":   len(lines),
			"totalLength": calculateTotalLength(lines),
		}, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("メタデータの変換に失敗しました: %v", err)
		}

		// 結果の返却（メタデータとトランスクリプトを含む）
		result := fmt.Sprintf("メタデータ:\n%s\n\n動画ID:%s、言語:%sの文字起こし:\n\n%s",
			string(metadataJSON), videoID, lang, transcript)
		return mcp.NewToolResultText(result), nil
	})

	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("サーバーエラー: %v\n", err)
	}
}

// calculateTotalLength は字幕の総時間（秒）を計算します
func calculateTotalLength(lines []TranscriptLine) float64 {
	if len(lines) == 0 {
		return 0
	}

	lastLine := lines[len(lines)-1]
	return lastLine.Start + lastLine.Dur
}

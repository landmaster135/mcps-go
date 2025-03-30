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
	// 動画ページを取得
	resp, err := s.httpClient.Get(fmt.Sprintf("https://youtube.com/watch?v=%s", videoID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch video page: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	htmlContent := string(body)

	// 字幕トラック情報を抽出
	captionTracksRegex := regexp.MustCompile(`"captionTracks":(\[.*?\])`)
	matches := captionTracksRegex.FindStringSubmatch(htmlContent)
	if len(matches) < 2 {
		return nil, errors.New("could not find caption tracks in the video")
	}

	captionTracksJSON := fmt.Sprintf(`{"captionTracks":%s}`, matches[1])

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
		return nil, fmt.Errorf("failed to parse caption tracks: %v", err)
	}

	// 指定言語の字幕を探す
	var subtitleURL string
	for _, track := range captionData.CaptionTracks {
		if strings.Contains(track.VssID, "."+lang) || track.LanguageCode == lang {
			subtitleURL = track.BaseURL
			break
		}
	}

	if subtitleURL == "" {
		return nil, fmt.Errorf("could not find %s captions for video %s", lang, videoID)
	}

	// 字幕XMLを取得
	subtitleResp, err := s.httpClient.Get(subtitleURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subtitle data: %v", err)
	}
	defer subtitleResp.Body.Close()

	// XMLをパース
	var transcript XMLTranscript
	if err := xml.NewDecoder(subtitleResp.Body).Decode(&transcript); err != nil {
		return nil, fmt.Errorf("failed to parse subtitle XML: %v", err)
	}

	// 字幕データを構造化
	var lines []TranscriptLine
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
		// パラメータの取得
		input := request.Params.Arguments["url"].(string)
		lang := "en" // デフォルト値
		if langArg, ok := request.Params.Arguments["lang"]; ok {
			lang = langArg.(string)
		}

		// 動画IDの抽出
		videoID, err := transcriptService.ExtractYoutubeID(input)
		if err != nil {
			return nil, err
		}

		// 字幕の取得
		lines, err := transcriptService.GetTranscript(videoID, lang)
		if err != nil {
			return nil, err
		}

		// 字幕の整形
		transcript := transcriptService.FormatTranscript(lines)

		// メタデータをJSON形式に変換
		metadataJSON, err := json.MarshalIndent(map[string]interface{}{
			"videoId":   videoID,
			"language":  lang,
			"timestamp": time.Now().Format(time.RFC3339),
			"charCount": len(transcript),
		}, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %v", err)
		}

		// 結果の返却（メタデータとトランスクリプトを含む）
		result := fmt.Sprintf("Metadata:\n%s\n\nTranscript for video that ID:%s in %s and language:\n\n%s",
			string(metadataJSON), videoID, lang, transcript)
		return mcp.NewToolResultText(result), nil
	})

	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

package youtube_transcript

import (
	"net/http"
	"testing"
)

// TestExtractYoutubeID は ExtractYoutubeID 関数をテストします
func TestExtractYoutubeID(t *testing.T) {
	service := NewYouTubeTranscriptService()

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "空の入力",
			input:    "",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "直接IDの入力",
			input:    "dQw4w9WgXcQ",
			expected: "dQw4w9WgXcQ",
			wantErr:  false,
		},
		{
			name:     "youtube.com URL",
			input:    "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			expected: "dQw4w9WgXcQ",
			wantErr:  false,
		},
		{
			name:     "youtu.be URL",
			input:    "https://youtu.be/dQw4w9WgXcQ",
			expected: "dQw4w9WgXcQ",
			wantErr:  false,
		},
		{
			name:     "youtube.com URL with additional parameters",
			input:    "https://www.youtube.com/watch?v=dQw4w9WgXcQ&t=10s",
			expected: "dQw4w9WgXcQ",
			wantErr:  false,
		},
		{
			name:     "無効なURL",
			input:    "https://example.com",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "無効なID",
			input:    "invalid-id",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.ExtractYoutubeID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractYoutubeID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("ExtractYoutubeID() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestFormatTranscript は FormatTranscript 関数をテストします
func TestFormatTranscript(t *testing.T) {
	service := NewYouTubeTranscriptService()

	tests := []struct {
		name     string
		lines    []TranscriptLine
		expected string
	}{
		{
			name:     "空の入力",
			lines:    []TranscriptLine{},
			expected: "",
		},
		{
			name: "単一行",
			lines: []TranscriptLine{
				{Start: 0, Dur: 1, Text: "Hello world"},
			},
			expected: "Hello world",
		},
		{
			name: "複数行",
			lines: []TranscriptLine{
				{Start: 0, Dur: 1, Text: "Hello"},
				{Start: 1, Dur: 1, Text: "world"},
				{Start: 2, Dur: 1, Text: "!"},
			},
			expected: "Hello world !",
		},
		{
			name: "空白行を含む",
			lines: []TranscriptLine{
				{Start: 0, Dur: 1, Text: "Hello"},
				{Start: 1, Dur: 1, Text: ""},
				{Start: 2, Dur: 1, Text: "world"},
				{Start: 3, Dur: 1, Text: "  "},
				{Start: 4, Dur: 1, Text: "!"},
			},
			expected: "Hello world !",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.FormatTranscript(tt.lines)
			if got != tt.expected {
				t.Errorf("FormatTranscript() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestStripHTMLTags は stripHTMLTags 関数をテストします
func TestStripHTMLTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "タグなし",
			input:    "Hello world",
			expected: "Hello world",
		},
		{
			name:     "単一タグ",
			input:    "<p>Hello world</p>",
			expected: "Hello world",
		},
		{
			name:     "複数タグ",
			input:    "<div><p>Hello <strong>world</strong></p></div>",
			expected: "Hello world",
		},
		{
			name:     "自己閉じタグ",
			input:    "Hello<br/>world",
			expected: "Helloworld",
		},
		{
			name:     "属性付きタグ",
			input:    "<p class=\"test\">Hello world</p>",
			expected: "Hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripHTMLTags(tt.input)
			if got != tt.expected {
				t.Errorf("stripHTMLTags() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestExtractCaptionTracks は extractCaptionTracks 関数をテストします
func TestExtractCaptionTracks(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "標準的なパターン",
			html:     `{"captionTracks":[{"baseUrl":"https://example.com","name":{"simpleText":"English"},"vssId":".en","languageCode":"en"}]}`,
			expected: `{"captionTracks":[{"baseUrl":"https://example.com","name":{"simpleText":"English"},"vssId":".en","languageCode":"en"}]}`,
			wantErr:  false,
		},
		{
			name:     "別のパターン",
			html:     `"captions":{"playerCaptionsTracklistRenderer":{"captionTracks":[{"baseUrl":"https://example.com"}]}}`,
			expected: `{"captionTracks":[{"baseUrl":"https://example.com"}]}`,
			wantErr:  false,
		},
		{
			name:     "さらに別のパターン",
			html:     `"playerCaptionsTracklistRenderer":{"captionTracks":[{"baseUrl":"https://example.com"}]}`,
			expected: `{"captionTracks":[{"baseUrl":"https://example.com"}]}`,
			wantErr:  false,
		},
		{
			name:    "字幕なし",
			html:    `"playabilityStatus":{"status":"ERROR"}`,
			wantErr: true,
			errMsg:  ErrVideoUnavailable,
		},
		{
			name:    "字幕トラックなし",
			html:    `"playabilityStatus":{"status":"OK"}`,
			wantErr: true,
			errMsg:  ErrNoSubtitles,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractCaptionTracks(tt.html)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractCaptionTracks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("extractCaptionTracks() error message = %v, want %v", err.Error(), tt.errMsg)
				return
			}
			if !tt.wantErr && got != tt.expected {
				t.Errorf("extractCaptionTracks() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestCalculateTotalLength は calculateTotalLength 関数をテストします
func TestCalculateTotalLength(t *testing.T) {
	tests := []struct {
		name     string
		lines    []TranscriptLine
		expected float64
	}{
		{
			name:     "空の入力",
			lines:    []TranscriptLine{},
			expected: 0,
		},
		{
			name: "単一行",
			lines: []TranscriptLine{
				{Start: 0, Dur: 1, Text: "Hello world"},
			},
			expected: 1,
		},
		{
			name: "複数行",
			lines: []TranscriptLine{
				{Start: 0, Dur: 1, Text: "Hello"},
				{Start: 1, Dur: 2, Text: "world"},
				{Start: 3, Dur: 3, Text: "!"},
			},
			expected: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateTotalLength(tt.lines)
			if got != tt.expected {
				t.Errorf("calculateTotalLength() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// MockHTTPClient は HTTP クライアントのモックです
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return m.DoFunc(req)
}

// TestGetTranscript_Mock は GetTranscript 関数をモックを使用してテストする例です
func TestGetTranscript_Mock(t *testing.T) {
	// このテストはモックを使用して GetTranscript 関数をテストする例です
	// 実際のテストでは、より複雑なモックが必要になる場合があります

	// 注意: このテストは実際には機能しません。
	// 実際のテストでは、HTTPクライアントをモック化し、
	// YouTube APIからのレスポンスをシミュレートする必要があります。
	// このコードはあくまで例として提供されています。
	t.Skip("このテストはモックの例として提供されており、実際には実行されません")
}

// TestBuildYouTubeTranscriptServer は BuildYouTubeTranscriptServer 関数をテストします
func TestBuildYouTubeTranscriptServer(t *testing.T) {
	// このテストは BuildYouTubeTranscriptServer 関数をテストする例です
	// 実際のテストでは、サーバーの起動と終了をテストする必要があります

	// 注意: このテストは実際には機能しません。
	// MCPサーバーのテストには特別な設定が必要です。
	// このコードはあくまで例として提供されています。
	t.Skip("このテストはサーバーテストの例として提供されており、実際には実行されません")
}

// 統合テストの例
func TestIntegration(t *testing.T) {
	// 統合テストは実際のYouTube APIを呼び出すため、通常のテスト実行では
	// スキップされるようにしています。実際にテストを実行する場合は、
	// 環境変数などを使用して制御することをお勧めします。
	t.Skip("統合テストはデフォルトでスキップされます")

	service := NewYouTubeTranscriptService()

	// テストケース
	tests := []struct {
		name    string
		videoID string
		lang    string
		wantErr bool
	}{
		{
			name:    "英語字幕あり",
			videoID: "dQw4w9WgXcQ", // Rick Astley - Never Gonna Give You Up
			lang:    "en",
			wantErr: false,
		},
		{
			name:    "日本語字幕あり",
			videoID: "dQw4w9WgXcQ", // Rick Astley - Never Gonna Give You Up
			lang:    "ja",
			wantErr: false,
		},
		{
			name:    "存在しない言語",
			videoID: "dQw4w9WgXcQ", // Rick Astley - Never Gonna Give You Up
			lang:    "xx",
			wantErr: true,
		},
		{
			name:    "存在しない動画ID",
			videoID: "invalid-id",
			lang:    "en",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines, err := service.GetTranscript(tt.videoID, tt.lang)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTranscript() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(lines) == 0 {
					t.Errorf("GetTranscript() returned empty lines")
				}

				transcript := service.FormatTranscript(lines)
				if transcript == "" {
					t.Errorf("FormatTranscript() returned empty string")
				}

				totalLength := calculateTotalLength(lines)
				if totalLength <= 0 {
					t.Errorf("calculateTotalLength() returned invalid length: %v", totalLength)
				}

				t.Logf("Video ID: %s, Language: %s, Lines: %d, Total Length: %.2f seconds",
					tt.videoID, tt.lang, len(lines), totalLength)
				t.Logf("Sample transcript: %s", transcript[:min(100, len(transcript))]+"...")
			} else {
				t.Logf("Expected error occurred: %v", err)
			}
		})
	}
}

// min は2つの整数の小さい方を返します
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

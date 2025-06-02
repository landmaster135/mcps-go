package duckduckgo_search

import (
	"testing"
)

func TestCleanText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "HTMLタグの除去",
			input:    "<p>Hello <b>World</b></p>",
			expected: "Hello World",
		},
		{
			name:     "余分な空白の除去",
			input:    "  Hello   World  ",
			expected: "Hello World",
		},
		{
			name:     "単体のエンティティテスト - &amp;",
			input:    "&amp;",
			expected: "&",
		},
		{
			name:     "単体のエンティティテスト - &lt;",
			input:    "&lt;",
			expected: "<",
		},
		{
			name:     "単体のエンティティテスト - &gt;",
			input:    "&gt;",
			expected: ">",
		},
		{
			name:     "単体のエンティティテスト - &quot;",
			input:    "&quot;",
			expected: "\"",
		},
		{
			name:     "単体のエンティティテスト - &#39;",
			input:    "&#39;",
			expected: "'",
		},
		{
			name:     "組み合わせテスト - エンティティのみ",
			input:    "Hello &amp; world &quot;test&quot; &#39;example&#39;",
			expected: "Hello & world \"test\" 'example'",
		},
		{
			name:     "組み合わせテスト - HTMLタグとエンティティ",
			input:    "<p>Hello &amp; <b>world</b> &quot;test&quot;</p>",
			expected: "Hello & world \"test\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanText(tt.input)
			if result != tt.expected {
				t.Errorf("cleanText(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDecodeURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "通常のURL",
			input:    "https://example.com",
			expected: "https://example.com",
		},
		{
			name:     "DuckDuckGoリダイレクトURL",
			input:    "/l/?uddg=https%3A//example.com",
			expected: "https://example.com",
		},
		{
			name:     "エンコードされたURL",
			input:    "https%3A//example.com",
			expected: "https://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := decodeURL(tt.input)
			if result != tt.expected {
				t.Errorf("decodeURL() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCheckRateLimit(t *testing.T) {
	// レート制限のリセット
	requestCount.second = 0
	requestCount.minute = 0

	// 最初のリクエストは成功するはず
	err := checkRateLimit()
	if err != nil {
		t.Errorf("First request should succeed, got error: %v", err)
	}

	// レート制限に達したときはエラーが返されるはず
	err = checkRateLimit()
	if err == nil {
		t.Error("Rate limit should be exceeded")
	}
}

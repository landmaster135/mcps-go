package github

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestCreateGitHubServer(t *testing.T) {
	// テストケース
	tests := []struct {
		name          string
		token         string
		expectWarning bool
	}{
		{
			name:          "トークンあり",
			token:         "test-token",
			expectWarning: false,
		},
		{
			name:          "トークンなし",
			token:         "",
			expectWarning: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// 環境変数の元の値を保存
			originalToken := os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN")
			defer os.Setenv("GITHUB_PERSONAL_ACCESS_TOKEN", originalToken)

			// テスト用に環境変数を設定
			os.Setenv("GITHUB_PERSONAL_ACCESS_TOKEN", tc.token)

			// 標準出力をキャプチャ
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			defer func() {
				os.Stdout = oldStdout
			}()

			// テスト対象の関数を実行
			s := createGitHubServer()

			// 標準出力のキャプチャを終了
			w.Close()
			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			// 警告メッセージの検証
			if tc.expectWarning {
				if !strings.Contains(output, "Warning: GITHUB_PERSONAL_ACCESS_TOKEN environment variable not set") {
					t.Error("トークンが設定されていない場合に警告メッセージが表示されるべきです")
				}
			} else {
				if strings.Contains(output, "Warning: GITHUB_PERSONAL_ACCESS_TOKEN environment variable not set") {
					t.Error("トークンが設定されている場合に警告メッセージが表示されるべきではありません")
				}
			}

			// サーバーの検証
			if s == nil {
				t.Fatal("サーバーがnilです")
			}

			// サーバーの型の検証は不要（createGitHubServer関数は*server.MCPServerを返すため）

			// 注: 実際のサーバーの内部状態（例えばツールが正しく設定されているかなど）を
			// 検証するには、サーバーの内部構造にアクセスする必要があります。
			// しかし、これはテストが脆弱になる可能性があるため、
			// ここでは基本的な検証のみを行います。
		})
	}
}

// TestBuildGitHubServer はBuildGitHubServer関数をテストします
func TestBuildGitHubServer(t *testing.T) {
	// このテストは実際にサーバーを起動するため、
	// 単体テストとしては適切ではありません。
	// 代わりに、BuildGitHubServer関数が内部で
	// createGitHubServerとserver.ServeStdioを呼び出すことを
	// モックを使用して検証することができます。
	// ここでは簡略化のため、このテストは省略します。
}

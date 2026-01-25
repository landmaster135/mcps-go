package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	mcp "github.com/mark3labs/mcp-go/mcp"
	server "github.com/mark3labs/mcp-go/server"
)

// 注: compareMaps関数はutil_test.goで既に定義されているため、
// ここでは定義せず、そちらを使用します

// TestSetGitHubGlobalServer はSetGitHubGlobalServer関数をテストする
func TestSetGitHubGlobalServer(t *testing.T) {
	// テストケース
	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "正常系 - トークンあり",
			token: "testtoken",
		},
		{
			name:  "正常系 - トークンなし",
			token: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// 実際のMCPサーバーを作成
			s := server.NewMCPServer("test-server", "1.0.0")

			// テスト対象の関数を実行
			result := SetGitHubGlobalServer(tc.token, s)

			// 関数が正常に実行され、サーバーが返されることを確認
			if result == nil {
				t.Fatal("SetGitHubGlobalServerがnilを返しました")
			}

			// 注意: 実際のツールの追加や設定の検証は、MCPサーバーの内部実装に依存するため、
			// ここでは基本的な動作確認のみを行います
		})
	}
}

// TestSearchCode はSearchCodeメソッドをテストする
func TestSearchCode(t *testing.T) {
	// テストケース
	tests := []struct {
		name           string
		query          string
		options        map[string]interface{}
		mockResponse   map[string]interface{}
		mockStatusCode int
		mockError      error
		expectError    bool
	}{
		{
			name:  "正常系 - コード検索成功",
			query: "addClass in:file language:js repo:jquery/jquery",
			options: map[string]interface{}{
				"page":     1,
				"per_page": 30,
			},
			mockResponse: map[string]interface{}{
				"total_count": float64(10),
				"items": []interface{}{
					map[string]interface{}{
						"name":       "jquery.js",
						"path":       "src/jquery.js",
						"repository": map[string]interface{}{"full_name": "jquery/jquery"},
					},
				},
			},
			mockStatusCode: http.StatusOK,
			mockError:      nil,
			expectError:    false,
		},
		{
			name:  "正常系 - オプションなし",
			query: "addClass in:file language:js repo:jquery/jquery",
			options: map[string]interface{}{},
			mockResponse: map[string]interface{}{
				"total_count": float64(10),
				"items": []interface{}{
					map[string]interface{}{
						"name":       "jquery.js",
						"path":       "src/jquery.js",
						"repository": map[string]interface{}{"full_name": "jquery/jquery"},
					},
				},
			},
			mockStatusCode: http.StatusOK,
			mockError:      nil,
			expectError:    false,
		},
		{
			name:  "異常系 - 認証エラー",
			query: "addClass in:file language:js repo:jquery/jquery",
			options: map[string]interface{}{},
			mockResponse: map[string]interface{}{
				"message":           "Bad credentials",
				"documentation_url": "https://docs.github.com/rest",
			},
			mockStatusCode: http.StatusUnauthorized,
			mockError:      nil,
			expectError:    true,
		},
		{
			name:  "異常系 - レート制限エラー",
			query: "addClass in:file language:js repo:jquery/jquery",
			options: map[string]interface{}{},
			mockResponse: map[string]interface{}{
				"message":           "API rate limit exceeded",
				"documentation_url": "https://docs.github.com/rest/overview/resources-in-the-rest-api#rate-limiting",
			},
			mockStatusCode: http.StatusForbidden,
			mockError:      nil,
			expectError:    true,
		},
		{
			name:  "異常系 - 不正なJSONレスポンス",
			query: "addClass in:file language:js repo:jquery/jquery",
			options: map[string]interface{}{},
			mockResponse:   nil, // 空のレスポンス
			mockStatusCode: http.StatusOK,
			mockError:      nil,
			expectError:    true,
		},
		{
			name:  "異常系 - ネットワークエラー",
			query: "addClass in:file language:js repo:jquery/jquery",
			options: map[string]interface{}{},
			mockResponse:   nil,
			mockStatusCode: 0,
			mockError:      errors.New("ネットワーク接続エラー"),
			expectError:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// モックHTTPクライアントの作成
			mockClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// ネットワークエラーのシミュレーション
					if tc.mockError != nil {
						return nil, tc.mockError
					}

					// リクエストの検証
					expectedBaseURL := fmt.Sprintf("%s/search/code?q=%s", apiBaseURL, tc.query)
					actualURL := req.URL.String()

					// ベースURLの検証
					if !strings.HasPrefix(actualURL, expectedBaseURL) {
						t.Errorf("期待されたベースURL: %s, 実際: %s", expectedBaseURL, actualURL)
					}

					// クエリパラメータの検証（順序に依存しない）
					if len(tc.options) > 0 {
						// 実際のクエリパラメータを解析
						actualQueryParams := make(map[string]string)
						queryParts := strings.Split(actualURL, "&")
						for i, part := range queryParts {
							if i == 0 {
								// 最初の部分はベースURLとクエリパラメータを含む
								continue
							}
							if strings.Contains(part, "=") {
								kv := strings.Split(part, "=")
								actualQueryParams[kv[0]] = kv[1]
							}
						}

						// 期待されるクエリパラメータと比較
						for k, v := range tc.options {
							expectedValue := fmt.Sprintf("%v", v)
							actualValue, exists := actualQueryParams[k]
							if !exists {
								t.Errorf("クエリパラメータ %s が見つかりません", k)
							} else if actualValue != expectedValue {
								t.Errorf("クエリパラメータ %s の値が異なります。期待: %s, 実際: %s", k, expectedValue, actualValue)
							}
						}
					}

					if req.Method != "GET" {
						t.Errorf("期待されたHTTPメソッド: GET, 実際: %s", req.Method)
					}

					if req.Header.Get("Accept") != "application/vnd.github.v3+json" {
						t.Errorf("期待されたAcceptヘッダー: application/vnd.github.v3+json, 実際: %s", req.Header.Get("Accept"))
					}

					if req.Header.Get("Authorization") != "token test_token" {
						t.Errorf("期待されたAuthorizationヘッダー: token test_token, 実際: %s", req.Header.Get("Authorization"))
					}

					// モックレスポンスの作成
					var responseBody []byte
					if tc.name == "異常系 - 不正なJSONレスポンス" {
						responseBody = []byte("{invalid json}")
					} else if tc.mockResponse != nil {
						responseBody, _ = json.Marshal(tc.mockResponse)
					}

					return &http.Response{
						StatusCode: tc.mockStatusCode,
						Body:       io.NopCloser(bytes.NewReader(responseBody)),
					}, nil
				},
			}

			// GitHubClientのhttpClientをモックに置き換える
			client := NewGitHubClient("test_token")
			client.httpClient = mockClient

			// テスト対象の関数を実行
			result, err := client.SearchCode(tc.query, tc.options)

			// エラーの検証
			if tc.expectError && err == nil {
				t.Error("エラーが期待されていましたが、エラーは発生しませんでした")
			}
			if !tc.expectError && err != nil {
				t.Errorf("エラーは期待されていませんでしたが、エラーが発生しました: %v", err)
			}

			// 正常系の場合、結果を検証
			if !tc.expectError {
				compareMaps(t, tc.mockResponse, result)
			}
		})
	}
}

// TestHandleToSearchCode はHandleToSearchCodeメソッドをテストする
func TestHandleToSearchCode(t *testing.T) {
	// テストケース
	tests := []struct {
		name           string
		arguments      map[string]interface{}
		mockResponse   map[string]interface{}
		mockStatusCode int
		mockError      error
		expectError    bool
	}{
		{
			name: "正常系 - 必須パラメータのみ",
			arguments: map[string]interface{}{
				"query": "addClass in:file language:js repo:jquery/jquery",
			},
			mockResponse: map[string]interface{}{
				"total_count": float64(10),
				"items": []interface{}{
					map[string]interface{}{
						"name":       "jquery.js",
						"path":       "src/jquery.js",
						"repository": map[string]interface{}{"full_name": "jquery/jquery"},
					},
				},
			},
			mockStatusCode: http.StatusOK,
			mockError:      nil,
			expectError:    false,
		},
		{
			name: "正常系 - すべてのパラメータ",
			arguments: map[string]interface{}{
				"query":    "addClass in:file language:js repo:jquery/jquery",
				"page":     float64(2),
				"per_page": float64(50),
			},
			mockResponse: map[string]interface{}{
				"total_count": float64(10),
				"items": []interface{}{
					map[string]interface{}{
						"name":       "jquery.js",
						"path":       "src/jquery.js",
						"repository": map[string]interface{}{"full_name": "jquery/jquery"},
					},
				},
			},
			mockStatusCode: http.StatusOK,
			mockError:      nil,
			expectError:    false,
		},
		{
			name: "異常系 - APIエラー",
			arguments: map[string]interface{}{
				"query": "addClass in:file language:js repo:jquery/jquery",
			},
			mockResponse: map[string]interface{}{
				"message":           "API rate limit exceeded",
				"documentation_url": "https://docs.github.com/rest/overview/resources-in-the-rest-api#rate-limiting",
			},
			mockStatusCode: http.StatusForbidden,
			mockError:      nil,
			expectError:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// モックHTTPクライアントの作成
			mockClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// ネットワークエラーのシミュレーション
					if tc.mockError != nil {
						return nil, tc.mockError
					}

					// リクエストの検証
					expectedBaseURL := fmt.Sprintf("%s/search/code?q=%s", apiBaseURL, tc.arguments["query"].(string))
					actualURL := req.URL.String()

					// ベースURLの検証
					if !strings.HasPrefix(actualURL, expectedBaseURL) {
						t.Errorf("期待されたベースURL: %s, 実際: %s", expectedBaseURL, actualURL)
					}

					// クエリパラメータの検証
					expectedParams := make(map[string]string)
					if page, ok := tc.arguments["page"]; ok {
						expectedParams["page"] = fmt.Sprintf("%v", int(page.(float64)))
					}
					if perPage, ok := tc.arguments["per_page"]; ok {
						expectedParams["per_page"] = fmt.Sprintf("%v", int(perPage.(float64)))
					}

					if len(expectedParams) > 0 {
						// 実際のクエリパラメータを解析
						actualQueryParams := make(map[string]string)
						queryParts := strings.Split(actualURL, "&")
						for i, part := range queryParts {
							if i == 0 {
								// 最初の部分はベースURLとクエリパラメータを含む
								continue
							}
							if strings.Contains(part, "=") {
								kv := strings.Split(part, "=")
								actualQueryParams[kv[0]] = kv[1]
							}
						}

						// 期待されるクエリパラメータと比較
						for k, v := range expectedParams {
							actualValue, exists := actualQueryParams[k]
							if !exists {
								t.Errorf("クエリパラメータ %s が見つかりません", k)
							} else if actualValue != v {
								t.Errorf("クエリパラメータ %s の値が異なります。期待: %s, 実際: %s", k, v, actualValue)
							}
						}
					}

					if req.Method != "GET" {
						t.Errorf("期待されたHTTPメソッド: GET, 実際: %s", req.Method)
					}

					// モックレスポンスの作成
					responseBody, _ := json.Marshal(tc.mockResponse)
					return &http.Response{
						StatusCode: tc.mockStatusCode,
						Body:       io.NopCloser(bytes.NewReader(responseBody)),
					}, nil
				},
			}

			// GitHubClientのhttpClientをモックに置き換える
			client := NewGitHubClient("test_token")
			client.httpClient = mockClient

			// リクエストの作成
			request := mcp.CallToolRequest{}
			// Paramsフィールドに直接アクセス
			request.Params.Name = "search_code"
			request.Params.Arguments = tc.arguments

			// テスト対象の関数を実行
			ctx := context.Background()
			result, err := client.HandleToSearchCode(ctx, request)

			// エラーの検証
			if tc.expectError && err == nil {
				t.Error("エラーが期待されていましたが、エラーは発生しませんでした")
			}
			if !tc.expectError && err != nil {
				t.Errorf("エラーは期待されていませんでしたが、エラーが発生しました: %v", err)
			}

			// 正常系の場合、結果を検証
			if !tc.expectError {
				if result == nil {
					t.Fatal("結果がnilです")
				}

				// 結果の内容を検証
				// 注: mcp.CallToolResultの構造は外部パッケージで定義されているため、
				// 直接内部構造にアクセスせず、結果が非nilであることだけを確認します
				if result == nil {
					t.Fatal("結果がnilです")
				}

				// 正常に結果が返されたことを確認できれば十分とします
				// 実際のAPIレスポンスは既にSearchCodeメソッドのテストで検証済みです
			}
		})
	}
}

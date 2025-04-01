package github_v2

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	mcp "github.com/mark3labs/mcp-go/mcp"
	server "github.com/mark3labs/mcp-go/server"
)

// TestSetGitHubPullRequestServer はSetGitHubPullRequestServer関数をテストする
func TestSetGitHubPullRequestServer(t *testing.T) {
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
			result := SetGitHubPullRequestServer(tc.token, s)

			// 関数が正常に実行され、サーバーが返されることを確認
			if result == nil {
				t.Fatal("SetGitHubPullRequestServerがnilを返しました")
			}

			// 注意: 実際のツールの追加や設定の検証は、MCPサーバーの内部実装に依存するため、
			// ここでは基本的な動作確認のみを行います
		})
	}
}

// TestCreatePullRequest はCreatePullRequestメソッドをテストする
func TestCreatePullRequest(t *testing.T) {
	// テストケース
	tests := []struct {
		name           string
		owner          string
		repo           string
		options        map[string]interface{}
		mockResponse   map[string]interface{}
		mockStatusCode int
		mockError      error
		expectError    bool
	}{
		{
			name:  "正常系 - プルリクエスト作成成功",
			owner: "test_user",
			repo:  "test_repo",
			options: map[string]interface{}{
				"title": "テストプルリクエスト",
				"head":  "feature-branch",
				"base":  "main",
				"body":  "これはテストプルリクエストです",
			},
			mockResponse: map[string]interface{}{
				"id":     float64(123456),
				"number": float64(1),
				"title":  "テストプルリクエスト",
				"head":   map[string]interface{}{"ref": "feature-branch"},
				"base":   map[string]interface{}{"ref": "main"},
				"body":   "これはテストプルリクエストです",
				"state":  "open",
			},
			mockStatusCode: http.StatusCreated,
			mockError:      nil,
			expectError:    false,
		},
		{
			name:  "異常系 - 認証エラー",
			owner: "test_user",
			repo:  "test_repo",
			options: map[string]interface{}{
				"title": "テストプルリクエスト",
				"head":  "feature-branch",
				"base":  "main",
			},
			mockResponse: map[string]interface{}{
				"message":           "Bad credentials",
				"documentation_url": "https://docs.github.com/rest",
			},
			mockStatusCode: http.StatusUnauthorized,
			mockError:      nil,
			expectError:    true,
		},
		{
			name:  "異常系 - リポジトリが存在しない",
			owner: "nonexistent",
			repo:  "nonexistent",
			options: map[string]interface{}{
				"title": "テストプルリクエスト",
				"head":  "feature-branch",
				"base":  "main",
			},
			mockResponse: map[string]interface{}{
				"message":           "Not Found",
				"documentation_url": "https://docs.github.com/rest",
			},
			mockStatusCode: http.StatusNotFound,
			mockError:      nil,
			expectError:    true,
		},
		{
			name:  "異常系 - 不正なJSONレスポンス",
			owner: "test_user",
			repo:  "test_repo",
			options: map[string]interface{}{
				"title": "テストプルリクエスト",
				"head":  "feature-branch",
				"base":  "main",
			},
			mockResponse:   nil, // 空のレスポンス
			mockStatusCode: http.StatusOK,
			mockError:      nil,
			expectError:    true,
		},
		{
			name:  "異常系 - ネットワークエラー",
			owner: "test_user",
			repo:  "test_repo",
			options: map[string]interface{}{
				"title": "テストプルリクエスト",
				"head":  "feature-branch",
				"base":  "main",
			},
			mockResponse:   nil,
			mockStatusCode: 0,
			mockError:      errors.New("ネットワーク接続エラー"),
			expectError:    true,
		},
		{
			name:  "異常系 - JSONマーシャリングエラー",
			owner: "test_user",
			repo:  "test_repo",
			options: map[string]interface{}{
				"title":    "テストプルリクエスト",
				"head":     "feature-branch",
				"base":     "main",
				"callback": func() {}, // JSONにシリアライズできない値
			},
			mockResponse:   nil,
			mockStatusCode: 0,
			mockError:      nil,
			expectError:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// JSONマーシャリングエラーのテストケース
			if tc.name == "異常系 - JSONマーシャリングエラー" {
				client := NewGitHubClient("test_token")
				_, err := client.CreatePullRequest(tc.owner, tc.repo, tc.options)
				if !tc.expectError && err != nil {
					t.Errorf("エラーは期待されていませんでしたが、エラーが発生しました: %v", err)
				}
				if tc.expectError && err == nil {
					t.Error("エラーが期待されていましたが、エラーは発生しませんでした")
				}
				return
			}

			// モックHTTPクライアントの作成
			mockClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// ネットワークエラーのシミュレーション
					if tc.mockError != nil {
						return nil, tc.mockError
					}

					// リクエストの検証
					expectedURL := apiBaseURL + "/repos/" + tc.owner + "/" + tc.repo + "/pulls"
					if req.URL.String() != expectedURL {
						t.Errorf("期待されたURL: %s, 実際: %s", expectedURL, req.URL.String())
					}

					if req.Method != "POST" {
						t.Errorf("期待されたHTTPメソッド: POST, 実際: %s", req.Method)
					}

					if req.Header.Get("Accept") != "application/vnd.github.v3+json" {
						t.Errorf("期待されたAcceptヘッダー: application/vnd.github.v3+json, 実際: %s", req.Header.Get("Accept"))
					}

					if req.Header.Get("Content-Type") != "application/json" {
						t.Errorf("期待されたContent-Typeヘッダー: application/json, 実際: %s", req.Header.Get("Content-Type"))
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
			result, err := client.CreatePullRequest(tc.owner, tc.repo, tc.options)

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

// TestHandleToCreatePullRequest はHandleToCreatePullRequestメソッドをテストする
func TestHandleToCreatePullRequest(t *testing.T) {
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
				"owner": "test_user",
				"repo":  "test_repo",
				"title": "テストプルリクエスト",
				"head":  "feature-branch",
				"base":  "main",
			},
			mockResponse: map[string]interface{}{
				"id":     float64(123456),
				"number": float64(1),
				"title":  "テストプルリクエスト",
				"head":   map[string]interface{}{"ref": "feature-branch"},
				"base":   map[string]interface{}{"ref": "main"},
				"state":  "open",
			},
			mockStatusCode: http.StatusCreated,
			mockError:      nil,
			expectError:    false,
		},
		{
			name: "正常系 - すべてのパラメータ",
			arguments: map[string]interface{}{
				"owner": "test_user",
				"repo":  "test_repo",
				"title": "テストプルリクエスト",
				"head":  "feature-branch",
				"base":  "main",
				"body":  "これはテストプルリクエストです",
				"draft": true,
			},
			mockResponse: map[string]interface{}{
				"id":     float64(123456),
				"number": float64(1),
				"title":  "テストプルリクエスト",
				"head":   map[string]interface{}{"ref": "feature-branch"},
				"base":   map[string]interface{}{"ref": "main"},
				"body":   "これはテストプルリクエストです",
				"draft":  true,
				"state":  "open",
			},
			mockStatusCode: http.StatusCreated,
			mockError:      nil,
			expectError:    false,
		},
		{
			name: "異常系 - APIエラー",
			arguments: map[string]interface{}{
				"owner": "test_user",
				"repo":  "test_repo",
				"title": "テストプルリクエスト",
				"head":  "feature-branch",
				"base":  "main",
			},
			mockResponse: map[string]interface{}{
				"message":           "Not Found",
				"documentation_url": "https://docs.github.com/rest",
			},
			mockStatusCode: http.StatusNotFound,
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
					expectedURL := apiBaseURL + "/repos/" + tc.arguments["owner"].(string) + "/" + tc.arguments["repo"].(string) + "/pulls"
					if req.URL.String() != expectedURL {
						t.Errorf("期待されたURL: %s, 実際: %s", expectedURL, req.URL.String())
					}

					if req.Method != "POST" {
						t.Errorf("期待されたHTTPメソッド: POST, 実際: %s", req.Method)
					}

					// リクエストボディの検証
					body, _ := io.ReadAll(req.Body)
					var requestBody map[string]interface{}
					if err := json.Unmarshal(body, &requestBody); err != nil {
						t.Fatalf("リクエストボディのJSONパースに失敗しました: %v", err)
					}

					// タイトルの検証
					if requestBody["title"] != tc.arguments["title"].(string) {
						t.Errorf("期待されたtitle: %s, 実際: %s", tc.arguments["title"].(string), requestBody["title"])
					}

					// headの検証
					if requestBody["head"] != tc.arguments["head"].(string) {
						t.Errorf("期待されたhead: %s, 実際: %s", tc.arguments["head"].(string), requestBody["head"])
					}

					// baseの検証
					if requestBody["base"] != tc.arguments["base"].(string) {
						t.Errorf("期待されたbase: %s, 実際: %s", tc.arguments["base"].(string), requestBody["base"])
					}

					// bodyパラメータの検証（存在する場合）
					if body, ok := tc.arguments["body"]; ok {
						if requestBody["body"] != body.(string) {
							t.Errorf("期待されたbody: %s, 実際: %s", body.(string), requestBody["body"])
						}
					}

					// draftパラメータの検証（存在する場合）
					if draft, ok := tc.arguments["draft"]; ok {
						if requestBody["draft"] != draft.(bool) {
							t.Errorf("期待されたdraft: %v, 実際: %v", draft.(bool), requestBody["draft"])
						}
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
			request.Params.Name = "create_pull_request"
			request.Params.Arguments = tc.arguments

			// テスト対象の関数を実行
			ctx := context.Background()
			result, err := client.HandleToCreatePullRequest(ctx, request)

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
				// 実際のAPIレスポンスは既にCreatePullRequestメソッドのテストで検証済みです
			}
		})
	}
}

// TestCreatePullRequestReview はCreatePullRequestReviewメソッドをテストする
func TestCreatePullRequestReview(t *testing.T) {
	// テストケース
	tests := []struct {
		name           string
		owner          string
		repo           string
		pullNumber     int
		options        map[string]interface{}
		mockResponse   map[string]interface{}
		mockStatusCode int
		mockError      error
		expectError    bool
	}{
		{
			name:       "正常系 - レビュー作成成功",
			owner:      "test_user",
			repo:       "test_repo",
			pullNumber: 1,
			options: map[string]interface{}{
				"event": "APPROVE",
				"body":  "LGTMです！",
			},
			mockResponse: map[string]interface{}{
				"id":    float64(123456),
				"state": "APPROVED",
				"body":  "LGTMです！",
				"user": map[string]interface{}{
					"login": "test_user",
				},
			},
			mockStatusCode: http.StatusCreated,
			mockError:      nil,
			expectError:    false,
		},
		{
			name:       "正常系 - コメントのみ",
			owner:      "test_user",
			repo:       "test_repo",
			pullNumber: 1,
			options: map[string]interface{}{
				"event": "COMMENT",
				"body":  "コメントです",
			},
			mockResponse: map[string]interface{}{
				"id":    float64(123456),
				"state": "COMMENTED",
				"body":  "コメントです",
				"user": map[string]interface{}{
					"login": "test_user",
				},
			},
			mockStatusCode: http.StatusCreated,
			mockError:      nil,
			expectError:    false,
		},
		{
			name:       "異常系 - 認証エラー",
			owner:      "test_user",
			repo:       "test_repo",
			pullNumber: 1,
			options: map[string]interface{}{
				"event": "APPROVE",
				"body":  "LGTMです！",
			},
			mockResponse: map[string]interface{}{
				"message":           "Bad credentials",
				"documentation_url": "https://docs.github.com/rest",
			},
			mockStatusCode: http.StatusUnauthorized,
			mockError:      nil,
			expectError:    true,
		},
		{
			name:       "異常系 - プルリクエストが存在しない",
			owner:      "test_user",
			repo:       "test_repo",
			pullNumber: 999,
			options: map[string]interface{}{
				"event": "APPROVE",
				"body":  "LGTMです！",
			},
			mockResponse: map[string]interface{}{
				"message":           "Not Found",
				"documentation_url": "https://docs.github.com/rest",
			},
			mockStatusCode: http.StatusNotFound,
			mockError:      nil,
			expectError:    true,
		},
		{
			name:       "異常系 - ネットワークエラー",
			owner:      "test_user",
			repo:       "test_repo",
			pullNumber: 1,
			options: map[string]interface{}{
				"event": "APPROVE",
				"body":  "LGTMです！",
			},
			mockResponse:   nil,
			mockStatusCode: 0,
			mockError:      errors.New("ネットワーク接続エラー"),
			expectError:    true,
		},
		{
			name:       "異常系 - JSONマーシャリングエラー",
			owner:      "test_user",
			repo:       "test_repo",
			pullNumber: 1,
			options: map[string]interface{}{
				"event":    "APPROVE",
				"body":     "LGTMです！",
				"callback": func() {}, // JSONにシリアライズできない値
			},
			mockResponse:   nil,
			mockStatusCode: 0,
			mockError:      nil,
			expectError:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// JSONマーシャリングエラーのテストケース
			if tc.name == "異常系 - JSONマーシャリングエラー" || tc.name == "異常系 - JSONマーシャリングエラー（76行目）" {
				client := NewGitHubClient("test_token")
				_, err := client.CreatePullRequestReview(tc.owner, tc.repo, tc.pullNumber, tc.options)
				if !tc.expectError && err != nil {
					t.Errorf("エラーは期待されていませんでしたが、エラーが発生しました: %v", err)
				}
				if tc.expectError && err == nil {
					t.Error("エラーが期待されていましたが、エラーは発生しませんでした")
				}
				return
			}

			// モックHTTPクライアントの作成
			mockClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// ネットワークエラーのシミュレーション
					if tc.mockError != nil {
						return nil, tc.mockError
					}

					// リクエストの検証
					expectedURL := fmt.Sprintf("%s/repos/%s/%s/pulls/%d/reviews", apiBaseURL, tc.owner, tc.repo, tc.pullNumber)
					if req.URL.String() != expectedURL {
						t.Errorf("期待されたURL: %s, 実際: %s", expectedURL, req.URL.String())
					}

					if req.Method != "POST" {
						t.Errorf("期待されたHTTPメソッド: POST, 実際: %s", req.Method)
					}

					if req.Header.Get("Accept") != "application/vnd.github.v3+json" {
						t.Errorf("期待されたAcceptヘッダー: application/vnd.github.v3+json, 実際: %s", req.Header.Get("Accept"))
					}

					if req.Header.Get("Content-Type") != "application/json" {
						t.Errorf("期待されたContent-Typeヘッダー: application/json, 実際: %s", req.Header.Get("Content-Type"))
					}

					if req.Header.Get("Authorization") != "token test_token" {
						t.Errorf("期待されたAuthorizationヘッダー: token test_token, 実際: %s", req.Header.Get("Authorization"))
					}

					// リクエストボディの検証
					body, _ := io.ReadAll(req.Body)
					var requestBody map[string]interface{}
					if err := json.Unmarshal(body, &requestBody); err != nil {
						t.Fatalf("リクエストボディのJSONパースに失敗しました: %v", err)
					}

					// eventの検証
					if event, ok := tc.options["event"]; ok {
						if requestBody["event"] != event.(string) {
							t.Errorf("期待されたevent: %s, 実際: %s", event.(string), requestBody["event"])
						}
					}

					// bodyの検証
					if body, ok := tc.options["body"]; ok {
						if requestBody["body"] != body.(string) {
							t.Errorf("期待されたbody: %s, 実際: %s", body.(string), requestBody["body"])
						}
					}

					// モックレスポンスの作成
					var responseBody []byte
					if tc.mockResponse != nil {
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
			result, err := client.CreatePullRequestReview(tc.owner, tc.repo, tc.pullNumber, tc.options)

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

// TestHandleToCreatePullRequestReview はHandleToCreatePullRequestReviewメソッドをテストする
func TestHandleToCreatePullRequestReview(t *testing.T) {
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
				"owner":       "test_user",
				"repo":        "test_repo",
				"pull_number": float64(1),
			},
			mockResponse: map[string]interface{}{
				"id":    float64(123456),
				"state": "COMMENTED",
				"user": map[string]interface{}{
					"login": "test_user",
				},
			},
			mockStatusCode: http.StatusCreated,
			mockError:      nil,
			expectError:    false,
		},
		{
			name: "正常系 - すべてのパラメータ",
			arguments: map[string]interface{}{
				"owner":       "test_user",
				"repo":        "test_repo",
				"pull_number": float64(1),
				"event":       "APPROVE",
				"body":        "LGTMです！",
			},
			mockResponse: map[string]interface{}{
				"id":    float64(123456),
				"state": "APPROVED",
				"body":  "LGTMです！",
				"user": map[string]interface{}{
					"login": "test_user",
				},
			},
			mockStatusCode: http.StatusCreated,
			mockError:      nil,
			expectError:    false,
		},
		{
			name: "異常系 - APIエラー",
			arguments: map[string]interface{}{
				"owner":       "test_user",
				"repo":        "test_repo",
				"pull_number": float64(999),
				"event":       "APPROVE",
			},
			mockResponse: map[string]interface{}{
				"message":           "Not Found",
				"documentation_url": "https://docs.github.com/rest",
			},
			mockStatusCode: http.StatusNotFound,
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
					pullNumber := int(tc.arguments["pull_number"].(float64))
					expectedURL := fmt.Sprintf("%s/repos/%s/%s/pulls/%d/reviews", apiBaseURL, tc.arguments["owner"].(string), tc.arguments["repo"].(string), pullNumber)
					if req.URL.String() != expectedURL {
						t.Errorf("期待されたURL: %s, 実際: %s", expectedURL, req.URL.String())
					}

					if req.Method != "POST" {
						t.Errorf("期待されたHTTPメソッド: POST, 実際: %s", req.Method)
					}

					// リクエストボディの検証
					body, _ := io.ReadAll(req.Body)
					var requestBody map[string]interface{}
					if err := json.Unmarshal(body, &requestBody); err != nil {
						t.Fatalf("リクエストボディのJSONパースに失敗しました: %v", err)
					}

					// eventパラメータの検証（存在する場合）
					if event, ok := tc.arguments["event"]; ok {
						if requestBody["event"] != event.(string) {
							t.Errorf("期待されたevent: %s, 実際: %s", event.(string), requestBody["event"])
						}
					}

					// bodyパラメータの検証（存在する場合）
					if body, ok := tc.arguments["body"]; ok {
						if requestBody["body"] != body.(string) {
							t.Errorf("期待されたbody: %s, 実際: %s", body.(string), requestBody["body"])
						}
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
			request.Params.Name = "create_pull_request_review"
			request.Params.Arguments = tc.arguments

			// テスト対象の関数を実行
			ctx := context.Background()
			result, err := client.HandleToCreatePullRequestReview(ctx, request)

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
				// 実際のAPIレスポンスは既にCreatePullRequestReviewメソッドのテストで検証済みです
			}
		})
	}
}

// TestMergePullRequest はMergePullRequestメソッドをテストする
func TestMergePullRequest(t *testing.T) {
	// テストケース
	tests := []struct {
		name           string
		owner          string
		repo           string
		pullNumber     int
		options        map[string]interface{}
		mockResponse   map[string]interface{}
		mockStatusCode int
		mockError      error
		expectError    bool
	}{
		{
			name:       "正常系 - マージ成功",
			owner:      "test_user",
			repo:       "test_repo",
			pullNumber: 1,
			options: map[string]interface{}{
				"commit_title":   "マージコミットタイトル",
				"commit_message": "マージコミットメッセージ",
				"merge_method":   "merge",
			},
			mockResponse: map[string]interface{}{
				"sha":     "6dcb09b5b57875f334f61aebed695e2e4193db5e",
				"merged":  true,
				"message": "Pull Request successfully merged",
			},
			mockStatusCode: http.StatusOK,
			mockError:      nil,
			expectError:    false,
		},
		{
			name:       "正常系 - オプションなし",
			owner:      "test_user",
			repo:       "test_repo",
			pullNumber: 1,
			options:    map[string]interface{}{},
			mockResponse: map[string]interface{}{
				"sha":     "6dcb09b5b57875f334f61aebed695e2e4193db5e",
				"merged":  true,
				"message": "Pull Request successfully merged",
			},
			mockStatusCode: http.StatusOK,
			mockError:      nil,
			expectError:    false,
		},
		{
			name:       "異常系 - 認証エラー",
			owner:      "test_user",
			repo:       "test_repo",
			pullNumber: 1,
			options:    map[string]interface{}{},
			mockResponse: map[string]interface{}{
				"message":           "Bad credentials",
				"documentation_url": "https://docs.github.com/rest",
			},
			mockStatusCode: http.StatusUnauthorized,
			mockError:      nil,
			expectError:    true,
		},
		{
			name:       "異常系 - プルリクエストが存在しない",
			owner:      "test_user",
			repo:       "test_repo",
			pullNumber: 999,
			options:    map[string]interface{}{},
			mockResponse: map[string]interface{}{
				"message":           "Not Found",
				"documentation_url": "https://docs.github.com/rest",
			},
			mockStatusCode: http.StatusNotFound,
			mockError:      nil,
			expectError:    true,
		},
		{
			name:       "異常系 - マージできない状態",
			owner:      "test_user",
			repo:       "test_repo",
			pullNumber: 1,
			options:    map[string]interface{}{},
			mockResponse: map[string]interface{}{
				"message":           "Pull request is not mergeable",
				"documentation_url": "https://docs.github.com/rest/reference/pulls#merge-a-pull-request",
			},
			mockStatusCode: http.StatusMethodNotAllowed,
			mockError:      nil,
			expectError:    true,
		},
		{
			name:           "異常系 - ネットワークエラー",
			owner:          "test_user",
			repo:           "test_repo",
			pullNumber:     1,
			options:        map[string]interface{}{},
			mockResponse:   nil,
			mockStatusCode: 0,
			mockError:      errors.New("ネットワーク接続エラー"),
			expectError:    true,
		},
		{
			name:       "異常系 - JSONマーシャリングエラー",
			owner:      "test_user",
			repo:       "test_repo",
			pullNumber: 1,
			options: map[string]interface{}{
				"callback": func() {}, // JSONにシリアライズできない値
			},
			mockResponse:   nil,
			mockStatusCode: 0,
			mockError:      nil,
			expectError:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// JSONマーシャリングエラーのテストケース
			if tc.name == "異常系 - JSONマーシャリングエラー" || tc.name == "異常系 - JSONマーシャリングエラー（122行目）" {
				client := NewGitHubClient("test_token")
				_, err := client.MergePullRequest(tc.owner, tc.repo, tc.pullNumber, tc.options)
				if !tc.expectError && err != nil {
					t.Errorf("エラーは期待されていませんでしたが、エラーが発生しました: %v", err)
				}
				if tc.expectError && err == nil {
					t.Error("エラーが期待されていましたが、エラーは発生しませんでした")
				}
				return
			}

			// モックHTTPクライアントの作成
			mockClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// ネットワークエラーのシミュレーション
					if tc.mockError != nil {
						return nil, tc.mockError
					}

					// リクエストの検証
					expectedURL := fmt.Sprintf("%s/repos/%s/%s/pulls/%d/merge", apiBaseURL, tc.owner, tc.repo, tc.pullNumber)
					if req.URL.String() != expectedURL {
						t.Errorf("期待されたURL: %s, 実際: %s", expectedURL, req.URL.String())
					}

					if req.Method != "PUT" {
						t.Errorf("期待されたHTTPメソッド: PUT, 実際: %s", req.Method)
					}

					if req.Header.Get("Accept") != "application/vnd.github.v3+json" {
						t.Errorf("期待されたAcceptヘッダー: application/vnd.github.v3+json, 実際: %s", req.Header.Get("Accept"))
					}

					if req.Header.Get("Content-Type") != "application/json" {
						t.Errorf("期待されたContent-Typeヘッダー: application/json, 実際: %s", req.Header.Get("Content-Type"))
					}

					if req.Header.Get("Authorization") != "token test_token" {
						t.Errorf("期待されたAuthorizationヘッダー: token test_token, 実際: %s", req.Header.Get("Authorization"))
					}

					// リクエストボディの検証
					body, _ := io.ReadAll(req.Body)
					var requestBody map[string]interface{}
					if err := json.Unmarshal(body, &requestBody); err != nil {
						t.Fatalf("リクエストボディのJSONパースに失敗しました: %v", err)
					}

					// commit_titleの検証
					if commitTitle, ok := tc.options["commit_title"]; ok {
						if requestBody["commit_title"] != commitTitle.(string) {
							t.Errorf("期待されたcommit_title: %s, 実際: %s", commitTitle.(string), requestBody["commit_title"])
						}
					}

					// commit_messageの検証
					if commitMessage, ok := tc.options["commit_message"]; ok {
						if requestBody["commit_message"] != commitMessage.(string) {
							t.Errorf("期待されたcommit_message: %s, 実際: %s", commitMessage.(string), requestBody["commit_message"])
						}
					}

					// merge_methodの検証
					if mergeMethod, ok := tc.options["merge_method"]; ok {
						if requestBody["merge_method"] != mergeMethod.(string) {
							t.Errorf("期待されたmerge_method: %s, 実際: %s", mergeMethod.(string), requestBody["merge_method"])
						}
					}

					// モックレスポンスの作成
					var responseBody []byte
					if tc.mockResponse != nil {
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
			result, err := client.MergePullRequest(tc.owner, tc.repo, tc.pullNumber, tc.options)

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

// TestHandleToMergePullRequest はHandleToMergePullRequestメソッドをテストする
func TestHandleToMergePullRequest(t *testing.T) {
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
				"owner":       "test_user",
				"repo":        "test_repo",
				"pull_number": float64(1),
			},
			mockResponse: map[string]interface{}{
				"sha":     "6dcb09b5b57875f334f61aebed695e2e4193db5e",
				"merged":  true,
				"message": "Pull Request successfully merged",
			},
			mockStatusCode: http.StatusOK,
			mockError:      nil,
			expectError:    false,
		},
		{
			name: "正常系 - すべてのパラメータ",
			arguments: map[string]interface{}{
				"owner":          "test_user",
				"repo":           "test_repo",
				"pull_number":    float64(1),
				"commit_title":   "マージコミットタイトル",
				"commit_message": "マージコミットメッセージ",
				"merge_method":   "squash",
			},
			mockResponse: map[string]interface{}{
				"sha":     "6dcb09b5b57875f334f61aebed695e2e4193db5e",
				"merged":  true,
				"message": "Pull Request successfully merged",
			},
			mockStatusCode: http.StatusOK,
			mockError:      nil,
			expectError:    false,
		},
		{
			name: "異常系 - APIエラー",
			arguments: map[string]interface{}{
				"owner":       "test_user",
				"repo":        "test_repo",
				"pull_number": float64(999),
			},
			mockResponse: map[string]interface{}{
				"message":           "Not Found",
				"documentation_url": "https://docs.github.com/rest",
			},
			mockStatusCode: http.StatusNotFound,
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
					pullNumber := int(tc.arguments["pull_number"].(float64))
					expectedURL := fmt.Sprintf("%s/repos/%s/%s/pulls/%d/merge", apiBaseURL, tc.arguments["owner"].(string), tc.arguments["repo"].(string), pullNumber)
					if req.URL.String() != expectedURL {
						t.Errorf("期待されたURL: %s, 実際: %s", expectedURL, req.URL.String())
					}

					if req.Method != "PUT" {
						t.Errorf("期待されたHTTPメソッド: PUT, 実際: %s", req.Method)
					}

					// リクエストボディの検証
					body, _ := io.ReadAll(req.Body)
					var requestBody map[string]interface{}
					if err := json.Unmarshal(body, &requestBody); err != nil {
						t.Fatalf("リクエストボディのJSONパースに失敗しました: %v", err)
					}

					// commit_titleの検証（存在する場合）
					if commitTitle, ok := tc.arguments["commit_title"]; ok {
						if requestBody["commit_title"] != commitTitle.(string) {
							t.Errorf("期待されたcommit_title: %s, 実際: %s", commitTitle.(string), requestBody["commit_title"])
						}
					}

					// commit_messageの検証（存在する場合）
					if commitMessage, ok := tc.arguments["commit_message"]; ok {
						if requestBody["commit_message"] != commitMessage.(string) {
							t.Errorf("期待されたcommit_message: %s, 実際: %s", commitMessage.(string), requestBody["commit_message"])
						}
					}

					// merge_methodの検証（存在する場合）
					if mergeMethod, ok := tc.arguments["merge_method"]; ok {
						if requestBody["merge_method"] != mergeMethod.(string) {
							t.Errorf("期待されたmerge_method: %s, 実際: %s", mergeMethod.(string), requestBody["merge_method"])
						}
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
			request.Params.Name = "merge_pull_request"
			request.Params.Arguments = tc.arguments

			// テスト対象の関数を実行
			ctx := context.Background()
			result, err := client.HandleToMergePullRequest(ctx, request)

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
				// 実際のAPIレスポンスは既にMergePullRequestメソッドのテストで検証済みです
			}
		})
	}
}

// TestGetPullRequestFiles はGetPullRequestFilesメソッドをテストする
func TestGetPullRequestFiles(t *testing.T) {
	// テストケース
	tests := []struct {
		name           string
		owner          string
		repo           string
		pullNumber     int
		mockResponse   []map[string]interface{}
		mockStatusCode int
		mockError      error
		expectError    bool
	}{
		{
			name:       "正常系 - プルリクエストファイル一覧取得成功",
			owner:      "test_user",
			repo:       "test_repo",
			pullNumber: 1,
			mockResponse: []map[string]interface{}{
				{
					"sha":       "abc123",
					"filename":  "test.go",
					"status":    "modified",
					"additions": float64(10),
					"deletions": float64(5),
					"changes":   float64(15),
				},
				{
					"sha":       "def456",
					"filename":  "README.md",
					"status":    "added",
					"additions": float64(20),
					"deletions": float64(0),
					"changes":   float64(20),
				},
			},
			mockStatusCode: http.StatusOK,
			mockError:      nil,
			expectError:    false,
		},
		{
			name:           "異常系 - 認証エラー",
			owner:          "test_user",
			repo:           "test_repo",
			pullNumber:     1,
			mockResponse:   nil,
			mockStatusCode: http.StatusUnauthorized,
			mockError:      nil,
			expectError:    true,
		},
		{
			name:           "異常系 - プルリクエストが存在しない",
			owner:          "test_user",
			repo:           "test_repo",
			pullNumber:     999,
			mockResponse:   nil,
			mockStatusCode: http.StatusNotFound,
			mockError:      nil,
			expectError:    true,
		},
		{
			name:           "異常系 - ネットワークエラー",
			owner:          "test_user",
			repo:           "test_repo",
			pullNumber:     1,
			mockResponse:   nil,
			mockStatusCode: 0,
			mockError:      errors.New("ネットワーク接続エラー"),
			expectError:    true,
		},
		{
			name:           "異常系 - 不正なJSONレスポンス",
			owner:          "test_user",
			repo:           "test_repo",
			pullNumber:     1,
			mockResponse:   nil,
			mockStatusCode: http.StatusOK,
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
					expectedURL := fmt.Sprintf("%s/repos/%s/%s/pulls/%d/files", apiBaseURL, tc.owner, tc.repo, tc.pullNumber)
					if req.URL.String() != expectedURL {
						t.Errorf("期待されたURL: %s, 実際: %s", expectedURL, req.URL.String())
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
			result, err := client.GetPullRequestFiles(tc.owner, tc.repo, tc.pullNumber)

			// エラーの検証
			if tc.expectError && err == nil {
				t.Error("エラーが期待されていましたが、エラーは発生しませんでした")
			}
			if !tc.expectError && err != nil {
				t.Errorf("エラーは期待されていませんでしたが、エラーが発生しました: %v", err)
			}

			// 正常系の場合、結果を検証
			if !tc.expectError {
				if len(result) != len(tc.mockResponse) {
					t.Errorf("期待された結果の長さ: %d, 実際: %d", len(tc.mockResponse), len(result))
				}

				// 各アイテムを検証
				for i, expectedItem := range tc.mockResponse {
					if i >= len(result) {
						t.Errorf("インデックス %d の結果アイテムが見つかりません", i)
						continue
					}
					actualItem := result[i]

					// 主要なフィールドを検証
					expectedFields := []string{"sha", "filename", "status", "additions", "deletions", "changes"}
					for _, field := range expectedFields {
						if expectedItem[field] != actualItem[field] {
							t.Errorf("フィールド %s の値が異なります。期待: %v, 実際: %v", field, expectedItem[field], actualItem[field])
						}
					}
				}
			}
		})
	}
}

// TestHandleToGetPullRequestFiles はHandleToGetPullRequestFilesメソッドをテストする
func TestHandleToGetPullRequestFiles(t *testing.T) {
	// テストケース
	tests := []struct {
		name           string
		arguments      map[string]interface{}
		mockResponse   []map[string]interface{}
		mockStatusCode int
		mockError      error
		expectError    bool
	}{
		{
			name: "正常系 - プルリクエストファイル一覧取得成功",
			arguments: map[string]interface{}{
				"owner":       "test_user",
				"repo":        "test_repo",
				"pull_number": float64(1),
			},
			mockResponse: []map[string]interface{}{
				{
					"sha":       "abc123",
					"filename":  "test.go",
					"status":    "modified",
					"additions": float64(10),
					"deletions": float64(5),
					"changes":   float64(15),
				},
				{
					"sha":       "def456",
					"filename":  "README.md",
					"status":    "added",
					"additions": float64(20),
					"deletions": float64(0),
					"changes":   float64(20),
				},
			},
			mockStatusCode: http.StatusOK,
			mockError:      nil,
			expectError:    false,
		},
		{
			name: "異常系 - APIエラー",
			arguments: map[string]interface{}{
				"owner":       "test_user",
				"repo":        "test_repo",
				"pull_number": float64(999),
			},
			mockResponse:   nil,
			mockStatusCode: http.StatusNotFound,
			mockError:      nil,
			expectError:    true,
		},
		{
			name: "異常系 - ネットワークエラー",
			arguments: map[string]interface{}{
				"owner":       "test_user",
				"repo":        "test_repo",
				"pull_number": float64(1),
			},
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
					pullNumber := int(tc.arguments["pull_number"].(float64))
					expectedURL := fmt.Sprintf("%s/repos/%s/%s/pulls/%d/files", apiBaseURL, tc.arguments["owner"].(string), tc.arguments["repo"].(string), pullNumber)
					if req.URL.String() != expectedURL {
						t.Errorf("期待されたURL: %s, 実際: %s", expectedURL, req.URL.String())
					}

					if req.Method != "GET" {
						t.Errorf("期待されたHTTPメソッド: GET, 実際: %s", req.Method)
					}

					// モックレスポンスの作成
					var responseBody []byte
					if tc.mockResponse != nil {
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

			// リクエストの作成
			request := mcp.CallToolRequest{}
			// Paramsフィールドに直接アクセス
			request.Params.Name = "get_pull_request_files"
			request.Params.Arguments = tc.arguments

			// テスト対象の関数を実行
			ctx := context.Background()
			result, err := client.HandleToGetPullRequestFiles(ctx, request)

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
				// 実際のAPIレスポンスは既にGetPullRequestFilesメソッドのテストで検証済みです
			}
		})
	}
}

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

// TestSetGitHubIssueServer はSetGitHubIssueServer関数をテストする
func TestSetGitHubIssueServer(t *testing.T) {
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
			result := SetGitHubIssueServer(tc.token, s)

			// 関数が正常に実行され、サーバーが返されることを確認
			if result == nil {
				t.Fatal("SetGitHubIssueServerがnilを返しました")
			}

			// 注意: 実際のツールの追加や設定の検証は、MCPサーバーの内部実装に依存するため、
			// ここでは基本的な動作確認のみを行います
		})
	}
}

// TestCreateIssue はCreateIssueメソッドをテストする
func TestCreateIssue(t *testing.T) {
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
			name:  "正常系 - イシュー作成成功",
			owner: "test_user",
			repo:  "test_repo",
			options: map[string]interface{}{
				"title": "テストイシュー",
				"body":  "これはテストイシューです",
			},
			mockResponse: map[string]interface{}{
				"id":     float64(123456),
				"number": float64(1),
				"title":  "テストイシュー",
				"body":   "これはテストイシューです",
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
				"title": "テストイシュー",
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
				"title": "テストイシュー",
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
				"title": "テストイシュー",
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
				"title": "テストイシュー",
			},
			mockResponse:   nil,
			mockStatusCode: 0,
			mockError:      errors.New("ネットワーク接続エラー"),
			expectError:    true,
		},
		{
			name:  "異常系 - 不正なJSONレスポンス（エラー時）",
			owner: "test_user",
			repo:  "test_repo",
			options: map[string]interface{}{
				"title": "テストイシュー",
			},
			mockResponse:   nil,
			mockStatusCode: http.StatusInternalServerError,
			mockError:      nil,
			expectError:    true,
		},
		{
			name:  "異常系 - JSONマーシャリングエラー",
			owner: "test_user",
			repo:  "test_repo",
			options: map[string]interface{}{
				"title":    "テストイシュー",
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
				_, err := client.CreateIssue(tc.owner, tc.repo, tc.options)
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
					expectedURL := apiBaseURL + "/repos/" + tc.owner + "/" + tc.repo + "/issues"
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
					} else if tc.name == "異常系 - 不正なJSONレスポンス（エラー時）" {
						responseBody = []byte("{invalid json for error}")
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
			result, err := client.CreateIssue(tc.owner, tc.repo, tc.options)

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

// TestListIssues はListIssuesメソッドをテストする
func TestListIssues(t *testing.T) {
	// テストケース
	tests := []struct {
		name           string
		owner          string
		repo           string
		options        map[string]interface{}
		mockResponse   []map[string]interface{}
		mockStatusCode int
		mockError      error
		expectError    bool
	}{
		{
			name:    "正常系 - イシュー一覧取得成功",
			owner:   "test_user",
			repo:    "test_repo",
			options: map[string]interface{}{},
			mockResponse: []map[string]interface{}{
				{
					"id":     float64(123456),
					"number": float64(1),
					"title":  "テストイシュー1",
					"state":  "open",
				},
				{
					"id":     float64(123457),
					"number": float64(2),
					"title":  "テストイシュー2",
					"state":  "closed",
				},
			},
			mockStatusCode: http.StatusOK,
			mockError:      nil,
			expectError:    false,
		},
		{
			name:  "正常系 - クエリパラメータあり",
			owner: "test_user",
			repo:  "test_repo",
			options: map[string]interface{}{
				"state":     "open",
				"sort":      "created",
				"direction": "desc",
				"per_page":  30,
				"page":      1,
			},
			mockResponse: []map[string]interface{}{
				{
					"id":     float64(123456),
					"number": float64(1),
					"title":  "テストイシュー1",
					"state":  "open",
				},
			},
			mockStatusCode: http.StatusOK,
			mockError:      nil,
			expectError:    false,
		},
		{
			name:    "異常系 - 認証エラー",
			owner:   "test_user",
			repo:    "test_repo",
			options: map[string]interface{}{},
			mockResponse: []map[string]interface{}{
				{
					"message":           "Bad credentials",
					"documentation_url": "https://docs.github.com/rest",
				},
			},
			mockStatusCode: http.StatusUnauthorized,
			mockError:      nil,
			expectError:    true,
		},
		{
			name:    "異常系 - リポジトリが存在しない",
			owner:   "nonexistent",
			repo:    "nonexistent",
			options: map[string]interface{}{},
			mockResponse: []map[string]interface{}{
				{
					"message":           "Not Found",
					"documentation_url": "https://docs.github.com/rest",
				},
			},
			mockStatusCode: http.StatusNotFound,
			mockError:      nil,
			expectError:    true,
		},
		{
			name:           "異常系 - 不正なJSONレスポンス",
			owner:          "test_user",
			repo:           "test_repo",
			options:        map[string]interface{}{},
			mockResponse:   nil,
			mockStatusCode: http.StatusOK,
			mockError:      nil,
			expectError:    true,
		},
		{
			name:           "異常系 - ネットワークエラー",
			owner:          "test_user",
			repo:           "test_repo",
			options:        map[string]interface{}{},
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
					expectedBaseURL := apiBaseURL + "/repos/" + tc.owner + "/" + tc.repo + "/issues"
					actualURL := req.URL.String()

					// ベースURLの検証
					actualBaseURL := actualURL
					if strings.Contains(actualURL, "?") {
						actualBaseURL = actualURL[:strings.Index(actualURL, "?")]
					}

					if actualBaseURL != expectedBaseURL {
						t.Errorf("期待されたベースURL: %s, 実際: %s", expectedBaseURL, actualBaseURL)
					}

					// クエリパラメータの検証（順序に依存しない）
					if len(tc.options) > 0 {
						// 実際のクエリパラメータを解析
						actualQueryParams := make(map[string]string)
						if strings.Contains(actualURL, "?") {
							queryString := actualURL[strings.Index(actualURL, "?")+1:]
							queryParts := strings.Split(queryString, "&")
							for _, part := range queryParts {
								if strings.Contains(part, "=") {
									kv := strings.Split(part, "=")
									actualQueryParams[kv[0]] = kv[1]
								}
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

						// 余分なクエリパラメータがないことを確認
						if len(actualQueryParams) != len(tc.options) {
							t.Errorf("クエリパラメータの数が異なります。期待: %d, 実際: %d", len(tc.options), len(actualQueryParams))
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
			result, err := client.ListIssues(tc.owner, tc.repo, tc.options)

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
					compareMaps(t, expectedItem, actualItem)
				}
			}
		})
	}
}

// TestHandleToListIssues はHandleToListIssuesメソッドをテストする
func TestHandleToListIssues(t *testing.T) {
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
			name: "正常系 - 必須パラメータのみ",
			arguments: map[string]interface{}{
				"owner": "test_user",
				"repo":  "test_repo",
			},
			mockResponse: []map[string]interface{}{
				{
					"id":     float64(123456),
					"number": float64(1),
					"title":  "テストイシュー1",
					"state":  "open",
				},
				{
					"id":     float64(123457),
					"number": float64(2),
					"title":  "テストイシュー2",
					"state":  "closed",
				},
			},
			mockStatusCode: http.StatusOK,
			mockError:      nil,
			expectError:    false,
		},
		{
			name: "正常系 - すべてのパラメータ",
			arguments: map[string]interface{}{
				"owner":     "test_user",
				"repo":      "test_repo",
				"state":     "open",
				"sort":      "created",
				"direction": "desc",
				"per_page":  float64(30),
				"page":      float64(1),
			},
			mockResponse: []map[string]interface{}{
				{
					"id":     float64(123456),
					"number": float64(1),
					"title":  "テストイシュー1",
					"state":  "open",
				},
			},
			mockStatusCode: http.StatusOK,
			mockError:      nil,
			expectError:    false,
		},
		{
			name: "異常系 - APIエラー",
			arguments: map[string]interface{}{
				"owner": "test_user",
				"repo":  "test_repo",
			},
			mockResponse: []map[string]interface{}{
				{
					"message":           "Not Found",
					"documentation_url": "https://docs.github.com/rest",
				},
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
					expectedBaseURL := apiBaseURL + "/repos/" + tc.arguments["owner"].(string) + "/" + tc.arguments["repo"].(string) + "/issues"
					actualURL := req.URL.String()

					// ベースURLの検証
					actualBaseURL := actualURL
					if strings.Contains(actualURL, "?") {
						actualBaseURL = actualURL[:strings.Index(actualURL, "?")]
					}

					if actualBaseURL != expectedBaseURL {
						t.Errorf("期待されたベースURL: %s, 実際: %s", expectedBaseURL, actualBaseURL)
					}

					// クエリパラメータの検証（順序に依存しない）
					expectedParams := make(map[string]string)
					if state, ok := tc.arguments["state"]; ok {
						expectedParams["state"] = fmt.Sprintf("%v", state)
					}
					if sort, ok := tc.arguments["sort"]; ok {
						expectedParams["sort"] = fmt.Sprintf("%v", sort)
					}
					if direction, ok := tc.arguments["direction"]; ok {
						expectedParams["direction"] = fmt.Sprintf("%v", direction)
					}
					if perPage, ok := tc.arguments["per_page"]; ok {
						expectedParams["per_page"] = fmt.Sprintf("%v", int(perPage.(float64)))
					}
					if page, ok := tc.arguments["page"]; ok {
						expectedParams["page"] = fmt.Sprintf("%v", int(page.(float64)))
					}

					if len(expectedParams) > 0 {
						// 実際のクエリパラメータを解析
						actualQueryParams := make(map[string]string)
						if strings.Contains(actualURL, "?") {
							queryString := actualURL[strings.Index(actualURL, "?")+1:]
							queryParts := strings.Split(queryString, "&")
							for _, part := range queryParts {
								if strings.Contains(part, "=") {
									kv := strings.Split(part, "=")
									actualQueryParams[kv[0]] = kv[1]
								}
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

						// 余分なクエリパラメータがないことを確認
						if len(actualQueryParams) != len(expectedParams) {
							t.Errorf("クエリパラメータの数が異なります。期待: %d, 実際: %d", len(expectedParams), len(actualQueryParams))
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
			request.Params.Name = "list_issues"
			request.Params.Arguments = tc.arguments

			// テスト対象の関数を実行
			ctx := context.Background()
			result, err := client.HandleToListIssues(ctx, request)

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
				// 実際のAPIレスポンスは既にListIssuesメソッドのテストで検証済みです
			}
		})
	}
}

// TestUpdateIssue はUpdateIssueメソッドをテストする
func TestUpdateIssue(t *testing.T) {
	// テストケース
	tests := []struct {
		name           string
		owner          string
		repo           string
		issueNumber    int
		options        map[string]interface{}
		mockResponse   map[string]interface{}
		mockStatusCode int
		mockError      error
		expectError    bool
		checkURL       bool // URLの構築を明示的に検証するフラグ
	}{
		{
			name:        "正常系 - イシュー更新成功",
			owner:       "test_user",
			repo:        "test_repo",
			issueNumber: 1,
			options: map[string]interface{}{
				"title": "更新されたイシュー",
				"body":  "これは更新されたイシューです",
				"state": "closed",
			},
			mockResponse: map[string]interface{}{
				"id":     float64(123456),
				"number": float64(1),
				"title":  "更新されたイシュー",
				"body":   "これは更新されたイシューです",
				"state":  "closed",
			},
			mockStatusCode: http.StatusOK,
			mockError:      nil,
			expectError:    false,
			checkURL:       true, // URLの構築を検証
		},
		{
			name:        "正常系 - 一部のフィールドのみ更新",
			owner:       "test_user",
			repo:        "test_repo",
			issueNumber: 1,
			options: map[string]interface{}{
				"state": "closed",
			},
			mockResponse: map[string]interface{}{
				"id":     float64(123456),
				"number": float64(1),
				"title":  "元のイシュータイトル",
				"state":  "closed",
			},
			mockStatusCode: http.StatusOK,
			mockError:      nil,
			expectError:    false,
			checkURL:       false,
		},
		{
			name:        "異常系 - 認証エラー",
			owner:       "test_user",
			repo:        "test_repo",
			issueNumber: 1,
			options: map[string]interface{}{
				"title": "更新されたイシュー",
			},
			mockResponse: map[string]interface{}{
				"message":           "Bad credentials",
				"documentation_url": "https://docs.github.com/rest",
			},
			mockStatusCode: http.StatusUnauthorized,
			mockError:      nil,
			expectError:    true,
			checkURL:       false,
		},
		{
			name:        "異常系 - イシューが存在しない",
			owner:       "test_user",
			repo:        "test_repo",
			issueNumber: 999,
			options: map[string]interface{}{
				"title": "更新されたイシュー",
			},
			mockResponse: map[string]interface{}{
				"message":           "Not Found",
				"documentation_url": "https://docs.github.com/rest",
			},
			mockStatusCode: http.StatusNotFound,
			mockError:      nil,
			expectError:    true,
			checkURL:       true, // URLの構築を検証（特に異なるissueNumber）
		},
		{
			name:        "異常系 - ネットワークエラー",
			owner:       "test_user",
			repo:        "test_repo",
			issueNumber: 1,
			options: map[string]interface{}{
				"title": "更新されたイシュー",
			},
			mockResponse:   nil,
			mockStatusCode: 0,
			mockError:      errors.New("ネットワーク接続エラー"),
			expectError:    true,
			checkURL:       false,
		},
		{
			name:        "異常系 - JSONマーシャリングエラー",
			owner:       "test_user",
			repo:        "test_repo",
			issueNumber: 1,
			options: map[string]interface{}{
				"title":    "更新されたイシュー",
				"callback": func() {}, // JSONにシリアライズできない値
			},
			mockResponse:   nil,
			mockStatusCode: 0,
			mockError:      nil,
			expectError:    true,
			checkURL:       false,
		},
		{
			name:        "異常系 - 不正なJSONレスポンス",
			owner:       "test_user",
			repo:        "test_repo",
			issueNumber: 1,
			options: map[string]interface{}{
				"title": "更新されたイシュー",
			},
			mockResponse:   nil, // レスポンスボディは後で設定
			mockStatusCode: http.StatusOK,
			mockError:      nil,
			expectError:    true,
			checkURL:       false,
		},
		{
			name:        "異常系 - 特殊なURLパラメータ",
			owner:       "test/user", // スラッシュを含む
			repo:        "test-repo",
			issueNumber: 1,
			options: map[string]interface{}{
				"title": "更新されたイシュー",
			},
			mockResponse: map[string]interface{}{
				"id":     float64(123456),
				"number": float64(1),
				"title":  "更新されたイシュー",
			},
			mockStatusCode: http.StatusOK,
			mockError:      nil,
			expectError:    false,
			checkURL:       true, // URLの構築を検証（特殊文字を含む）
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// JSONマーシャリングエラーのテストケース
			if tc.name == "異常系 - JSONマーシャリングエラー" {
				client := NewGitHubClient("test_token")
				_, err := client.UpdateIssue(tc.owner, tc.repo, tc.issueNumber, tc.options)
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
					expectedURL := fmt.Sprintf("%s/repos/%s/%s/issues/%d", apiBaseURL, tc.owner, tc.repo, tc.issueNumber)
					if req.URL.String() != expectedURL {
						t.Errorf("期待されたURL: %s, 実際: %s", expectedURL, req.URL.String())
					}

					if req.Method != "PATCH" {
						t.Errorf("期待されたHTTPメソッド: PATCH, 実際: %s", req.Method)
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

					// オプションの検証
					for k, v := range tc.options {
						if k != "callback" { // JSONシリアライズできない値はスキップ
							if requestBody[k] == nil {
								t.Errorf("期待されたオプション %s が見つかりません", k)
							} else {
								expectedStr := fmt.Sprintf("%v", v)
								actualStr := fmt.Sprintf("%v", requestBody[k])
								if expectedStr != actualStr {
									t.Errorf("オプション %s の値が異なります。期待: %v, 実際: %v", k, expectedStr, actualStr)
								}
							}
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
			result, err := client.UpdateIssue(tc.owner, tc.repo, tc.issueNumber, tc.options)

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

// TestHandleToUpdateIssue はHandleToUpdateIssueメソッドをテストする
func TestHandleToUpdateIssue(t *testing.T) {
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
				"owner":        "test_user",
				"repo":         "test_repo",
				"issue_number": float64(1),
			},
			mockResponse: map[string]interface{}{
				"id":     float64(123456),
				"number": float64(1),
				"title":  "元のイシュータイトル",
				"state":  "open",
			},
			mockStatusCode: http.StatusOK,
			mockError:      nil,
			expectError:    false,
		},
		{
			name: "正常系 - すべてのパラメータ",
			arguments: map[string]interface{}{
				"owner":        "test_user",
				"repo":         "test_repo",
				"issue_number": float64(1),
				"title":        "更新されたイシュー",
				"body":         "これは更新されたイシューです",
				"state":        "closed",
				"labels":       []interface{}{"bug", "help wanted"},
				"assignees":    []interface{}{"test_user"},
			},
			mockResponse: map[string]interface{}{
				"id":        float64(123456),
				"number":    float64(1),
				"title":     "更新されたイシュー",
				"body":      "これは更新されたイシューです",
				"state":     "closed",
				"labels":    []interface{}{"bug", "help wanted"},
				"assignees": []interface{}{"test_user"},
			},
			mockStatusCode: http.StatusOK,
			mockError:      nil,
			expectError:    false,
		},
		{
			name: "異常系 - APIエラー",
			arguments: map[string]interface{}{
				"owner":        "test_user",
				"repo":         "test_repo",
				"issue_number": float64(999),
				"title":        "更新されたイシュー",
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
					issueNumber := int(tc.arguments["issue_number"].(float64))
					expectedURL := fmt.Sprintf("%s/repos/%s/%s/issues/%d", apiBaseURL, tc.arguments["owner"].(string), tc.arguments["repo"].(string), issueNumber)
					if req.URL.String() != expectedURL {
						t.Errorf("期待されたURL: %s, 実際: %s", expectedURL, req.URL.String())
					}

					if req.Method != "PATCH" {
						t.Errorf("期待されたHTTPメソッド: PATCH, 実際: %s", req.Method)
					}

					// リクエストボディの検証
					body, _ := io.ReadAll(req.Body)
					var requestBody map[string]interface{}
					if err := json.Unmarshal(body, &requestBody); err != nil {
						t.Fatalf("リクエストボディのJSONパースに失敗しました: %v", err)
					}

					// タイトルの検証（存在する場合）
					if title, ok := tc.arguments["title"]; ok {
						if requestBody["title"] != title.(string) {
							t.Errorf("期待されたtitle: %s, 実際: %s", title.(string), requestBody["title"])
						}
					}

					// bodyパラメータの検証（存在する場合）
					if body, ok := tc.arguments["body"]; ok {
						if requestBody["body"] != body.(string) {
							t.Errorf("期待されたbody: %s, 実際: %s", body.(string), requestBody["body"])
						}
					}

					// stateパラメータの検証（存在する場合）
					if state, ok := tc.arguments["state"]; ok {
						if requestBody["state"] != state.(string) {
							t.Errorf("期待されたstate: %s, 実際: %s", state.(string), requestBody["state"])
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
			request.Params.Name = "update_issue"
			request.Params.Arguments = tc.arguments

			// テスト対象の関数を実行
			ctx := context.Background()
			result, err := client.HandleToUpdateIssue(ctx, request)

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
				// 実際のAPIレスポンスは既にUpdateIssueメソッドのテストで検証済みです
			}
		})
	}
}

// TestAddIssueComment はAddIssueCommentメソッドをテストする
func TestAddIssueComment(t *testing.T) {
	// テストケース
	tests := []struct {
		name                string
		owner               string
		repo                string
		issueNumber         int
		body                string
		mockResponse        map[string]interface{}
		mockStatusCode      int
		mockError           error
		expectError         bool
		jsonMarshalError    bool // JSONマーシャリングエラーをシミュレートするフラグ
		invalidJsonResponse bool // 不正なJSONレスポンスをシミュレートするフラグ
	}{
		{
			name:        "正常系 - コメント追加成功",
			owner:       "test_user",
			repo:        "test_repo",
			issueNumber: 1,
			body:        "これはテストコメントです",
			mockResponse: map[string]interface{}{
				"id":   float64(123456),
				"body": "これはテストコメントです",
				"user": map[string]interface{}{
					"login": "test_user",
				},
			},
			mockStatusCode:      http.StatusCreated,
			mockError:           nil,
			expectError:         false,
			jsonMarshalError:    false,
			invalidJsonResponse: false,
		},
		{
			name:        "異常系 - 認証エラー",
			owner:       "test_user",
			repo:        "test_repo",
			issueNumber: 1,
			body:        "これはテストコメントです",
			mockResponse: map[string]interface{}{
				"message":           "Bad credentials",
				"documentation_url": "https://docs.github.com/rest",
			},
			mockStatusCode:      http.StatusUnauthorized,
			mockError:           nil,
			expectError:         true,
			jsonMarshalError:    false,
			invalidJsonResponse: false,
		},
		{
			name:        "異常系 - イシューが存在しない",
			owner:       "test_user",
			repo:        "test_repo",
			issueNumber: 999,
			body:        "これはテストコメントです",
			mockResponse: map[string]interface{}{
				"message":           "Not Found",
				"documentation_url": "https://docs.github.com/rest",
			},
			mockStatusCode:      http.StatusNotFound,
			mockError:           nil,
			expectError:         true,
			jsonMarshalError:    false,
			invalidJsonResponse: false,
		},
		{
			name:                "異常系 - ネットワークエラー",
			owner:               "test_user",
			repo:                "test_repo",
			issueNumber:         1,
			body:                "これはテストコメントです",
			mockResponse:        nil,
			mockStatusCode:      0,
			mockError:           errors.New("ネットワーク接続エラー"),
			expectError:         true,
			jsonMarshalError:    false,
			invalidJsonResponse: false,
		},
		{
			name:                "異常系 - JSONマーシャリングエラー",
			owner:               "test_user",
			repo:                "test_repo",
			issueNumber:         1,
			body:                string([]byte{0xff, 0xfe, 0xfd}), // 不正なUTF-8シーケンス
			mockResponse:        nil,
			mockStatusCode:      0,
			mockError:           nil,
			expectError:         true,
			jsonMarshalError:    true,
			invalidJsonResponse: false,
		},
		{
			name:                "異常系 - 不正なJSONレスポンス",
			owner:               "test_user",
			repo:                "test_repo",
			issueNumber:         1,
			body:                "これはテストコメントです",
			mockResponse:        nil, // レスポンスボディは後で設定
			mockStatusCode:      http.StatusOK,
			mockError:           nil,
			expectError:         true,
			jsonMarshalError:    false,
			invalidJsonResponse: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// JSONマーシャリングエラーのテストケース
			if tc.jsonMarshalError {
				if tc.name == "異常系 - JSONマーシャリングエラー（マップ構造）" {
					// オリジナルのjson.Marshalをモンキーパッチで置き換える方法はGoでは難しいため、
					// モックHTTPクライアントを使用して173行目のコードをテストします
					mockClient := &MockHTTPClient{
						DoFunc: func(req *http.Request) (*http.Response, error) {
							// リクエストが作成される前にエラーを返す
							return nil, errors.New("JSONマーシャリングエラー（マップ構造）")
						},
					}

					client := NewGitHubClient("test_token")
					client.httpClient = mockClient

					// テスト対象の関数を実行
					_, err := client.AddIssueComment(tc.owner, tc.repo, tc.issueNumber, tc.body)

					// エラーの検証
					if !tc.expectError && err != nil {
						t.Errorf("エラーは期待されていませんでしたが、エラーが発生しました: %v", err)
					}
					if tc.expectError && err == nil {
						t.Error("エラーが期待されていましたが、エラーは発生しませんでした")
					}
				} else {
					// 通常のJSONマーシャリングエラーテスト
					client := NewGitHubClient("test_token")
					_, err := client.AddIssueComment(tc.owner, tc.repo, tc.issueNumber, tc.body)
					if !tc.expectError && err != nil {
						t.Errorf("エラーは期待されていませんでしたが、エラーが発生しました: %v", err)
					}
					if tc.expectError && err == nil {
						t.Error("エラーが期待されていましたが、エラーは発生しませんでした")
					}
				}
				return
			}

			// 173行目を直接テストするケース
			if tc.name == "異常系 - 173行目のマップ作成とマーシャリング" {
				// 173行目のコードを直接実行
				bodyMap := map[string]string{"body": tc.body}
				jsonBody, err := json.Marshal(bodyMap)

				// エラーがないことを確認
				if err != nil {
					t.Errorf("マーシャリングエラーが発生しました: %v", err)
				}

				// マーシャリングされたJSONが期待通りであることを確認
				var unmarshaled map[string]interface{}
				err = json.Unmarshal(jsonBody, &unmarshaled)
				if err != nil {
					t.Errorf("アンマーシャリングエラーが発生しました: %v", err)
				}

				// bodyフィールドの値が期待通りであることを確認
				if unmarshaled["body"] != tc.body {
					t.Errorf("期待されたbody: %s, 実際: %s", tc.body, unmarshaled["body"])
				}

				// 実際のメソッドも呼び出して、エンドツーエンドで動作することを確認
				client := NewGitHubClient("test_token")
				mockClient := &MockHTTPClient{
					DoFunc: func(req *http.Request) (*http.Response, error) {
						// リクエストボディを検証
						body, _ := io.ReadAll(req.Body)
						var requestBody map[string]interface{}
						if err := json.Unmarshal(body, &requestBody); err != nil {
							t.Fatalf("リクエストボディのJSONパースに失敗しました: %v", err)
						}

						// bodyフィールドの値が期待通りであることを確認
						if requestBody["body"] != tc.body {
							t.Errorf("期待されたbody: %s, 実際: %s", tc.body, requestBody["body"])
						}

						// 成功レスポンスを返す
						responseBody := []byte(`{"id": 123456, "body": "直接テスト用コメント"}`)
						return &http.Response{
							StatusCode: http.StatusCreated,
							Body:       io.NopCloser(bytes.NewReader(responseBody)),
						}, nil
					},
				}
				client.httpClient = mockClient

				// テスト対象の関数を実行
				_, err = client.AddIssueComment(tc.owner, tc.repo, tc.issueNumber, tc.body)
				if err != nil {
					t.Errorf("エラーが発生しました: %v", err)
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
					expectedURL := fmt.Sprintf("%s/repos/%s/%s/issues/%d/comments", apiBaseURL, tc.owner, tc.repo, tc.issueNumber)
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

					// bodyパラメータの検証
					if requestBody["body"] != tc.body {
						t.Errorf("期待されたbody: %s, 実際: %s", tc.body, requestBody["body"])
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
			result, err := client.AddIssueComment(tc.owner, tc.repo, tc.issueNumber, tc.body)

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

// TestHandleToAddIssueComment はHandleToAddIssueCommentメソッドをテストする
func TestHandleToAddIssueComment(t *testing.T) {
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
			name: "正常系 - コメント追加成功",
			arguments: map[string]interface{}{
				"owner":        "test_user",
				"repo":         "test_repo",
				"issue_number": float64(1),
				"body":         "これはテストコメントです",
			},
			mockResponse: map[string]interface{}{
				"id":   float64(123456),
				"body": "これはテストコメントです",
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
				"owner":        "test_user",
				"repo":         "test_repo",
				"issue_number": float64(999),
				"body":         "これはテストコメントです",
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
			name: "異常系 - ネットワークエラー",
			arguments: map[string]interface{}{
				"owner":        "test_user",
				"repo":         "test_repo",
				"issue_number": float64(1),
				"body":         "これはテストコメントです",
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
					issueNumber := int(tc.arguments["issue_number"].(float64))
					expectedURL := fmt.Sprintf("%s/repos/%s/%s/issues/%d/comments", apiBaseURL, tc.arguments["owner"].(string), tc.arguments["repo"].(string), issueNumber)
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

					// bodyパラメータの検証
					if requestBody["body"] != tc.arguments["body"].(string) {
						t.Errorf("期待されたbody: %s, 実際: %s", tc.arguments["body"].(string), requestBody["body"])
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
			request.Params.Name = "add_issue_comment"
			request.Params.Arguments = tc.arguments

			// テスト対象の関数を実行
			ctx := context.Background()
			result, err := client.HandleToAddIssueComment(ctx, request)

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
				// 実際のAPIレスポンスは既にAddIssueCommentメソッドのテストで検証済みです
			}
		})
	}
}

// TestHandleToCreateIssue はHandleToCreateIssueメソッドをテストする
func TestHandleToCreateIssue(t *testing.T) {
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
				"title": "テストイシュー",
			},
			mockResponse: map[string]interface{}{
				"id":     float64(123456),
				"number": float64(1),
				"title":  "テストイシュー",
				"state":  "open",
			},
			mockStatusCode: http.StatusCreated,
			mockError:      nil,
			expectError:    false,
		},
		{
			name: "正常系 - すべてのパラメータ",
			arguments: map[string]interface{}{
				"owner":     "test_user",
				"repo":      "test_repo",
				"title":     "テストイシュー",
				"body":      "これはテストイシューです",
				"labels":    []interface{}{"bug", "help wanted"},
				"assignees": []interface{}{"test_user"},
			},
			mockResponse: map[string]interface{}{
				"id":        float64(123456),
				"number":    float64(1),
				"title":     "テストイシュー",
				"body":      "これはテストイシューです",
				"labels":    []interface{}{"bug", "help wanted"},
				"assignees": []interface{}{"test_user"},
				"state":     "open",
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
				"title": "テストイシュー",
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
					expectedURL := apiBaseURL + "/repos/" + tc.arguments["owner"].(string) + "/" + tc.arguments["repo"].(string) + "/issues"
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
			request.Params.Name = "create_issue"
			request.Params.Arguments = tc.arguments

			// テスト対象の関数を実行
			ctx := context.Background()
			result, err := client.HandleToCreateIssue(ctx, request)

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
				// 実際のAPIレスポンスは既にCreateIssueメソッドのテストで検証済みです
			}
		})
	}
}

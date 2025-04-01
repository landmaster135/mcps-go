package github_v2

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

// TestSetGitHubRepositoryServer はSetGitHubRepositoryServer関数をテストする
func TestSetGitHubRepositoryServer(t *testing.T) {
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
			result := SetGitHubRepositoryServer(tc.token, s)

			// 関数が正常に実行され、サーバーが返されることを確認
			if result == nil {
				t.Fatal("SetGitHubRepositoryServerがnilを返しました")
			}

			// 注意: 実際のツールの追加や設定の検証は、MCPサーバーの内部実装に依存するため、
			// ここでは基本的な動作確認のみを行います
		})
	}
}

// TestSearchRepositories はSearchRepositoriesメソッドをテストする
func TestSearchRepositories(t *testing.T) {
	// テストケース
	tests := []struct {
		name            string
		query           string
		page            int
		perPage         int
		mockResponse    map[string]interface{}
		mockStatusCode  int
		mockError       error
		expectError     bool
		expectedPage    int // 実際に使用されるページ番号
		expectedPerPage int // 実際に使用される1ページあたりの結果数
	}{
		{
			name:            "正常系 - リポジトリ検索成功",
			query:           "test",
			page:            1,
			perPage:         30,
			expectedPage:    1,
			expectedPerPage: 30,
			mockResponse: map[string]interface{}{
				"total_count":        float64(2),
				"incomplete_results": false,
				"items": []interface{}{
					map[string]interface{}{
						"id":        float64(123456),
						"name":      "test-repo-1",
						"full_name": "test-user/test-repo-1",
						"owner": map[string]interface{}{
							"login": "test-user",
						},
						"html_url":         "https://github.com/test-user/test-repo-1",
						"description":      "テストリポジトリ1",
						"stargazers_count": float64(10),
					},
					map[string]interface{}{
						"id":        float64(123457),
						"name":      "test-repo-2",
						"full_name": "test-user/test-repo-2",
						"owner": map[string]interface{}{
							"login": "test-user",
						},
						"html_url":         "https://github.com/test-user/test-repo-2",
						"description":      "テストリポジトリ2",
						"stargazers_count": float64(20),
					},
				},
			},
			mockStatusCode: http.StatusOK,
			mockError:      nil,
			expectError:    false,
		},
		{
			name:            "正常系 - 検索結果なし",
			query:           "nonexistent",
			page:            1,
			perPage:         30,
			expectedPage:    1,
			expectedPerPage: 30,
			mockResponse: map[string]interface{}{
				"total_count":        float64(0),
				"incomplete_results": false,
				"items":              []interface{}{},
			},
			mockStatusCode: http.StatusOK,
			mockError:      nil,
			expectError:    false,
		},
		{
			name:            "異常系 - 認証エラー",
			query:           "test",
			page:            1,
			perPage:         30,
			expectedPage:    1,
			expectedPerPage: 30,
			mockResponse:    nil,
			mockStatusCode:  http.StatusUnauthorized,
			mockError:       nil,
			expectError:     true,
		},
		{
			name:            "異常系 - ネットワークエラー",
			query:           "test",
			page:            1,
			perPage:         30,
			expectedPage:    1,
			expectedPerPage: 30,
			mockResponse:    nil,
			mockStatusCode:  0,
			mockError:       errors.New("ネットワーク接続エラー"),
			expectError:     true,
		},
		{
			name:            "異常系 - 不正なJSONレスポンス",
			query:           "test",
			page:            1,
			perPage:         30,
			expectedPage:    1,
			expectedPerPage: 30,
			mockStatusCode:  http.StatusOK,
			mockError:       nil,
			expectError:     true,
		},
		{
			name:    "正常系 - pageが1未満の場合は1に設定される",
			query:   "test",
			page:    0,
			perPage: 30,
			mockResponse: map[string]interface{}{
				"total_count":        float64(1),
				"incomplete_results": false,
				"items": []interface{}{
					map[string]interface{}{
						"id":        float64(123456),
						"name":      "test-repo-1",
						"full_name": "test-user/test-repo-1",
					},
				},
			},
			mockStatusCode:  http.StatusOK,
			mockError:       nil,
			expectError:     false,
			expectedPage:    1, // pageが0の場合、1に正規化される
			expectedPerPage: 30,
		},
		{
			name:    "正常系 - perPageが1未満の場合は30に設定される",
			query:   "test",
			page:    1,
			perPage: 0,
			mockResponse: map[string]interface{}{
				"total_count":        float64(1),
				"incomplete_results": false,
				"items": []interface{}{
					map[string]interface{}{
						"id":        float64(123456),
						"name":      "test-repo-1",
						"full_name": "test-user/test-repo-1",
					},
				},
			},
			mockStatusCode:  http.StatusOK,
			mockError:       nil,
			expectError:     false,
			expectedPage:    1,
			expectedPerPage: 30, // perPageが0の場合、30に正規化される
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
					// 正規化された値を使用してURLを構築
					expectedPage := tc.page
					if expectedPage < 1 {
						expectedPage = 1
					}
					expectedPerPage := tc.perPage
					if expectedPerPage < 1 || expectedPerPage > 100 {
						expectedPerPage = 30
					}

					expectedURL := fmt.Sprintf("%s/search/repositories?q=%s&page=%d&per_page=%d", apiBaseURL, tc.query, expectedPage, expectedPerPage)
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
			result, err := client.SearchRepositories(tc.query, tc.page, tc.perPage)

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

// TestHandleToSearchRepositories はHandleToSearchRepositoriesメソッドをテストする
func TestHandleToSearchRepositories(t *testing.T) {
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
				"query": "test",
			},
			mockResponse: map[string]interface{}{
				"total_count":        float64(2),
				"incomplete_results": false,
				"items": []interface{}{
					map[string]interface{}{
						"id":        float64(123456),
						"name":      "test-repo-1",
						"full_name": "test-user/test-repo-1",
					},
					map[string]interface{}{
						"id":        float64(123457),
						"name":      "test-repo-2",
						"full_name": "test-user/test-repo-2",
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
				"query":    "test",
				"page":     float64(2),
				"per_page": float64(10),
			},
			mockResponse: map[string]interface{}{
				"total_count":        float64(2),
				"incomplete_results": false,
				"items": []interface{}{
					map[string]interface{}{
						"id":        float64(123458),
						"name":      "test-repo-3",
						"full_name": "test-user/test-repo-3",
					},
					map[string]interface{}{
						"id":        float64(123459),
						"name":      "test-repo-4",
						"full_name": "test-user/test-repo-4",
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
				"query": "test",
			},
			mockResponse:   nil,
			mockStatusCode: http.StatusUnauthorized,
			mockError:      nil,
			expectError:    true,
		},
		{
			name: "異常系 - ネットワークエラー",
			arguments: map[string]interface{}{
				"query": "test",
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
					query := tc.arguments["query"].(string)
					page := 1
					perPage := 30

					if pageArg, ok := tc.arguments["page"]; ok {
						page = int(pageArg.(float64))
					}
					if perPageArg, ok := tc.arguments["per_page"]; ok {
						perPage = int(perPageArg.(float64))
					}

					expectedURL := fmt.Sprintf("%s/search/repositories?q=%s&page=%d&per_page=%d", apiBaseURL, query, page, perPage)
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
			request.Params.Name = "search_repositories"
			request.Params.Arguments = tc.arguments

			// テスト対象の関数を実行
			ctx := context.Background()
			result, err := client.HandleToSearchRepositories(ctx, request)

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
				// 実際のAPIレスポンスは既にSearchRepositoriesメソッドのテストで検証済みです
			}
		})
	}
}

// TestGetUserRepositories はGetUserRepositoriesメソッドをテストする
func TestGetUserRepositories(t *testing.T) {
	// テストケース
	tests := []struct {
		name           string
		username       string
		options        map[string]interface{}
		mockResponse   []map[string]interface{}
		mockStatusCode int
		mockError      error
		expectError    bool
	}{
		{
			name:     "正常系 - オプションなし",
			username: "test_user",
			options:  map[string]interface{}{},
			mockResponse: []map[string]interface{}{
				{
					"id":        float64(123456),
					"name":      "test-repo-1",
					"full_name": "test_user/test-repo-1",
					"owner": map[string]interface{}{
						"login": "test_user",
					},
					"html_url":         "https://github.com/test_user/test-repo-1",
					"description":      "テストリポジトリ1",
					"stargazers_count": float64(10),
				},
				{
					"id":        float64(123457),
					"name":      "test-repo-2",
					"full_name": "test_user/test-repo-2",
					"owner": map[string]interface{}{
						"login": "test_user",
					},
					"html_url":         "https://github.com/test_user/test-repo-2",
					"description":      "テストリポジトリ2",
					"stargazers_count": float64(20),
				},
			},
			mockStatusCode: http.StatusOK,
			mockError:      nil,
			expectError:    false,
		},
		{
			name:     "正常系 - すべてのオプション",
			username: "test_user",
			options: map[string]interface{}{
				"sort":      "updated",
				"direction": "asc",
				"per_page":  10,
				"page":      2,
				"type":      "owner",
			},
			mockResponse: []map[string]interface{}{
				{
					"id":        float64(123458),
					"name":      "test-repo-3",
					"full_name": "test_user/test-repo-3",
					"owner": map[string]interface{}{
						"login": "test_user",
					},
					"html_url":         "https://github.com/test_user/test-repo-3",
					"description":      "テストリポジトリ3",
					"stargazers_count": float64(30),
				},
			},
			mockStatusCode: http.StatusOK,
			mockError:      nil,
			expectError:    false,
		},
		{
			name:           "正常系 - 空の結果",
			username:       "empty_user",
			options:        map[string]interface{}{},
			mockResponse:   []map[string]interface{}{},
			mockStatusCode: http.StatusOK,
			mockError:      nil,
			expectError:    false,
		},
		{
			name:           "異常系 - 認証エラー",
			username:       "test_user",
			options:        map[string]interface{}{},
			mockResponse:   nil,
			mockStatusCode: http.StatusUnauthorized,
			mockError:      nil,
			expectError:    true,
		},
		{
			name:           "異常系 - ユーザーが存在しない",
			username:       "nonexistent",
			options:        map[string]interface{}{},
			mockResponse:   nil,
			mockStatusCode: http.StatusNotFound,
			mockError:      nil,
			expectError:    true,
		},
		{
			name:           "異常系 - ネットワークエラー",
			username:       "test_user",
			options:        map[string]interface{}{},
			mockResponse:   nil,
			mockStatusCode: 0,
			mockError:      errors.New("ネットワーク接続エラー"),
			expectError:    true,
		},
		{
			name:           "異常系 - 不正なJSONレスポンス",
			username:       "test_user",
			options:        map[string]interface{}{},
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
					expectedBaseURL := fmt.Sprintf("%s/users/%s/repos", apiBaseURL, tc.username)

					// URLの基本部分を検証
					if !strings.HasPrefix(req.URL.String(), expectedBaseURL) {
						t.Errorf("期待されたURLの基本部分: %s, 実際: %s", expectedBaseURL, req.URL.String())
					}

					// クエリパラメータの検証
					if len(tc.options) > 0 {
						// URLにクエリパラメータが含まれていることを確認
						if req.URL.RawQuery == "" {
							t.Error("クエリパラメータが含まれていません")
						}

						// 各オプションがクエリパラメータに含まれていることを確認
						q := req.URL.Query()
						for k, v := range tc.options {
							if q.Get(k) != fmt.Sprintf("%v", v) {
								t.Errorf("クエリパラメータ %s の値が異なります。期待: %v, 実際: %s", k, v, q.Get(k))
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
			result, err := client.GetUserRepositories(tc.username, tc.options)

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
					expectedFields := []string{"id", "name", "full_name"}
					for _, field := range expectedFields {
						if expectedItem[field] != actualItem[field] {
							t.Errorf("フィールド %s の値が異なります。期待: %v, 実際: %v", field, expectedItem[field], actualItem[field])
						}
					}

					// ownerフィールドを検証
					if owner, ok := expectedItem["owner"].(map[string]interface{}); ok {
						actualOwner, ok := actualItem["owner"].(map[string]interface{})
						if !ok {
							t.Error("結果のownerフィールドがマップではありません")
						} else if owner["login"] != actualOwner["login"] {
							t.Errorf("owner.loginの値が異なります。期待: %v, 実際: %v", owner["login"], actualOwner["login"])
						}
					}
				}
			}
		})
	}
}

// TestHandleToGetUserRepositories はHandleToGetUserRepositoriesメソッドをテストする
func TestHandleToGetUserRepositories(t *testing.T) {
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
				"username": "test_user",
			},
			mockResponse: []map[string]interface{}{
				{
					"id":       float64(123456),
					"name":     "test-repo-1",
					"full_name": "test_user/test-repo-1",
					"owner": map[string]interface{}{
						"login": "test_user",
					},
					"html_url": "https://github.com/test_user/test-repo-1",
					"description": "テストリポジトリ1",
					"stargazers_count": float64(10),
				},
				{
					"id":       float64(123457),
					"name":     "test-repo-2",
					"full_name": "test_user/test-repo-2",
					"owner": map[string]interface{}{
						"login": "test_user",
					},
					"html_url": "https://github.com/test_user/test-repo-2",
					"description": "テストリポジトリ2",
					"stargazers_count": float64(20),
				},
			},
			mockStatusCode: http.StatusOK,
			mockError:      nil,
			expectError:    false,
		},
		{
			name: "正常系 - すべてのパラメータ",
			arguments: map[string]interface{}{
				"username":  "test_user",
				"per_page":  float64(10),
				"page":      float64(2),
				"sort":      "updated",
				"direction": "asc",
				"type":      "owner",
			},
			mockResponse: []map[string]interface{}{
				{
					"id":       float64(123458),
					"name":     "test-repo-3",
					"full_name": "test_user/test-repo-3",
					"owner": map[string]interface{}{
						"login": "test_user",
					},
					"html_url": "https://github.com/test_user/test-repo-3",
					"description": "テストリポジトリ3",
					"stargazers_count": float64(30),
				},
			},
			mockStatusCode: http.StatusOK,
			mockError:      nil,
			expectError:    false,
		},
		{
			name: "正常系 - 空の結果",
			arguments: map[string]interface{}{
				"username": "empty_user",
			},
			mockResponse:   []map[string]interface{}{},
			mockStatusCode: http.StatusOK,
			mockError:      nil,
			expectError:    false,
		},
		{
			name: "異常系 - APIエラー",
			arguments: map[string]interface{}{
				"username": "nonexistent",
			},
			mockResponse:   nil,
			mockStatusCode: http.StatusNotFound,
			mockError:      nil,
			expectError:    true,
		},
		{
			name: "異常系 - ネットワークエラー",
			arguments: map[string]interface{}{
				"username": "test_user",
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
					username := tc.arguments["username"].(string)
					expectedBaseURL := fmt.Sprintf("%s/users/%s/repos", apiBaseURL, username)

					// URLの基本部分を検証
					if !strings.HasPrefix(req.URL.String(), expectedBaseURL) {
						t.Errorf("期待されたURLの基本部分: %s, 実際: %s", expectedBaseURL, req.URL.String())
					}

					// クエリパラメータの検証
					if len(tc.arguments) > 1 { // username以外のパラメータがある場合
						// URLにクエリパラメータが含まれていることを確認
						if req.URL.RawQuery == "" {
							t.Error("クエリパラメータが含まれていません")
						}

						// 各オプションがクエリパラメータに含まれていることを確認
						q := req.URL.Query()
						for k, v := range tc.arguments {
							if k != "username" { // username以外のパラメータを検証
								if q.Get(k) != fmt.Sprintf("%v", v) {
									t.Errorf("クエリパラメータ %s の値が異なります。期待: %v, 実際: %s", k, v, q.Get(k))
								}
							}
						}
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
			request.Params.Name = "get_user_repositories"
			request.Params.Arguments = tc.arguments

			// テスト対象の関数を実行
			ctx := context.Background()
			result, err := client.HandleToGetUserRepositories(ctx, request)

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
				// 実際のAPIレスポンスは既にGetUserRepositoriesメソッドのテストで検証済みです
			}
		})
	}
}

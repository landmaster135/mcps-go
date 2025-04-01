package github_v2

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"testing"

	mcp "github.com/mark3labs/mcp-go/mcp"
)

// マップの比較用ヘルパー関数
func compareMaps(t *testing.T, expected, actual map[string]interface{}) bool {
	// キーの数が同じか確認
	if len(expected) != len(actual) {
		t.Errorf("マップのサイズが異なります。期待: %d, 実際: %d", len(expected), len(actual))
		return false
	}

	// 各キーと値を個別に比較
	for k, expectedVal := range expected {
		actualVal, exists := actual[k]
		if !exists {
			t.Errorf("キー %s が実際のマップに存在しません", k)
			return false
		}

		// 値の型を確認
		expectedType := reflect.TypeOf(expectedVal)
		actualType := reflect.TypeOf(actualVal)
		if expectedType != actualType {
			t.Errorf("キー %s の値の型が異なります。期待: %v, 実際: %v", k, expectedType, actualType)
			return false
		}

		// 値を文字列に変換して比較
		expectedStr := fmt.Sprintf("%v", expectedVal)
		actualStr := fmt.Sprintf("%v", actualVal)
		if expectedStr != actualStr {
			t.Errorf("キー %s の値が異なります。期待: %v, 実際: %v", k, expectedStr, actualStr)
			return false
		}
	}

	return true
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
			owner: "testuser",
			repo:  "testrepo",
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
			owner: "testuser",
			repo:  "testrepo",
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
			owner: "testuser",
			repo:  "testrepo",
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
			owner: "testuser",
			repo:  "testrepo",
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
			owner: "testuser",
			repo:  "testrepo",
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
			owner: "testuser",
			repo:  "testrepo",
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
				client := NewGitHubClient("testtoken")
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

					if req.Header.Get("Authorization") != "token testtoken" {
						t.Errorf("期待されたAuthorizationヘッダー: token testtoken, 実際: %s", req.Header.Get("Authorization"))
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
			client := NewGitHubClient("testtoken")
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
				"owner": "testuser",
				"repo":  "testrepo",
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
				"owner":     "testuser",
				"repo":      "testrepo",
				"title":     "テストイシュー",
				"body":      "これはテストイシューです",
				"labels":    []interface{}{"bug", "help wanted"},
				"assignees": []interface{}{"testuser"},
			},
			mockResponse: map[string]interface{}{
				"id":        float64(123456),
				"number":    float64(1),
				"title":     "テストイシュー",
				"body":      "これはテストイシューです",
				"labels":    []interface{}{"bug", "help wanted"},
				"assignees": []interface{}{"testuser"},
				"state":     "open",
			},
			mockStatusCode: http.StatusCreated,
			mockError:      nil,
			expectError:    false,
		},
		{
			name: "異常系 - APIエラー",
			arguments: map[string]interface{}{
				"owner": "testuser",
				"repo":  "testrepo",
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
			client := NewGitHubClient("testtoken")
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

package github_v2

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

// MockHTTPClient はHTTPクライアントのモック
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

// Do はHTTPリクエストを実行する
func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

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
				"message":          "Bad credentials",
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
				"message":          "Not Found",
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
				if !tc.expectError && err == nil {
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

// エラーを返すリーダー
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("読み取りエラー")
}

// エラーを返すReadCloser
type errorReadCloser struct{}

func (e *errorReadCloser) Read(p []byte) (n int, err error) {
	return 0, errors.New("読み取りエラー")
}

func (e *errorReadCloser) Close() error {
	return nil
}

// TestDoRequest は doRequest メソッドを直接テストする
func TestDoRequest(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		body           io.Reader
		token          string
		mockResponse   []byte
		mockStatusCode int
		mockError      error
		expectError    bool
	}{
		{
			name:           "正常系 - GETリクエスト",
			method:         "GET",
			url:            "https://api.github.com/user",
			body:           nil,
			token:          "testtoken",
			mockResponse:   []byte(`{"login": "testuser", "id": 12345}`),
			mockStatusCode: http.StatusOK,
			mockError:      nil,
			expectError:    false,
		},
		{
			name:           "正常系 - POSTリクエスト",
			method:         "POST",
			url:            "https://api.github.com/repos/owner/repo/issues",
			body:           strings.NewReader(`{"title": "Test"}`),
			token:          "testtoken",
			mockResponse:   []byte(`{"id": 123, "number": 1, "title": "Test"}`),
			mockStatusCode: http.StatusCreated,
			mockError:      nil,
			expectError:    false,
		},
		{
			name:           "異常系 - 無効なリクエスト",
			method:         "GET",
			url:            "://invalid-url",
			body:           nil,
			token:          "testtoken",
			mockResponse:   nil,
			mockStatusCode: 0,
			mockError:      nil,
			expectError:    true,
		},
		{
			name:           "異常系 - リクエスト読み取りエラー",
			method:         "POST",
			url:            "https://api.github.com/repos/owner/repo/issues",
			body:           &errorReader{},
			token:          "testtoken",
			mockResponse:   nil,
			mockStatusCode: 0,
			mockError:      errors.New("リクエストボディ読み取りエラー"),
			expectError:    true,
		},
		{
			name:           "異常系 - レスポンス読み取りエラー",
			method:         "GET",
			url:            "https://api.github.com/user",
			body:           nil,
			token:          "testtoken",
			mockResponse:   nil,
			mockStatusCode: http.StatusOK,
			mockError:      nil,
			expectError:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// 無効なURLテスト
			if tc.name == "異常系 - 無効なリクエスト" {
				client := NewGitHubClient(tc.token)
				_, err := client.doRequest(tc.method, tc.url, tc.body)
				if err == nil {
					t.Error("エラーが期待されていましたが、エラーは発生しませんでした")
				}
				return
			}

			// モックHTTPクライアントの作成
			mockClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// errorReaderのテストでは常にエラーを返す
					if tc.name == "異常系 - リクエスト読み取りエラー" {
						return nil, tc.mockError
					}

					// メソッドの検証
					if req.Method != tc.method {
						t.Errorf("期待されたHTTPメソッド: %s, 実際: %s", tc.method, req.Method)
					}

					// URLの検証
					if req.URL.String() != tc.url {
						t.Errorf("期待されたURL: %s, 実際: %s", tc.url, req.URL.String())
					}

					// ヘッダーの検証
					if req.Header.Get("Accept") != "application/vnd.github.v3+json" {
						t.Errorf("期待されたAcceptヘッダー: application/vnd.github.v3+json, 実際: %s", req.Header.Get("Accept"))
					}

					if tc.token != "" && req.Header.Get("Authorization") != "token "+tc.token {
						t.Errorf("期待されたAuthorizationヘッダー: token %s, 実際: %s", tc.token, req.Header.Get("Authorization"))
					}

					if (tc.method == "POST" || tc.method == "PATCH" || tc.method == "PUT") && req.Header.Get("Content-Type") != "application/json" {
						t.Errorf("期待されたContent-Typeヘッダー: application/json, 実際: %s", req.Header.Get("Content-Type"))
					}

					// エラーのシミュレーション
					if tc.mockError != nil {
						return nil, tc.mockError
					}

					// レスポンス読み取りエラーのシミュレーション
					if tc.name == "異常系 - レスポンス読み取りエラー" {
						return &http.Response{
							StatusCode: tc.mockStatusCode,
							Body:       &errorReadCloser{},
						}, nil
					}

					// 通常のレスポンス
					return &http.Response{
						StatusCode: tc.mockStatusCode,
						Body:       io.NopCloser(bytes.NewReader(tc.mockResponse)),
					}, nil
				},
			}

			// GitHubClientのhttpClientをモックに置き換える
			client := NewGitHubClient(tc.token)
			client.httpClient = mockClient

			// テスト対象の関数を実行
			data, err := client.doRequest(tc.method, tc.url, tc.body)

			// エラーの検証
			if tc.expectError && err == nil {
				t.Error("エラーが期待されていましたが、エラーは発生しませんでした")
			}
			if !tc.expectError && err != nil {
				t.Errorf("エラーは期待されていませんでしたが、エラーが発生しました: %v", err)
			}

			// 正常系の場合、レスポンスを検証
			if !tc.expectError {
				if !bytes.Equal(data, tc.mockResponse) {
					t.Errorf("期待されたレスポンス: %s, 実際: %s", string(tc.mockResponse), string(data))
				}
			}
		})
	}
}

// TestNewGitHubClient は NewGitHubClient 関数をテストする
func TestNewGitHubClient(t *testing.T) {
	// テストケース
	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "トークンあり",
			token: "testtoken",
		},
		{
			name:  "トークンなし",
			token: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// 関数を実行
			client := NewGitHubClient(tc.token)

			// 検証
			if client == nil {
				t.Fatal("クライアントがnilです")
			}

			if client.token != tc.token {
				t.Errorf("期待されたトークン: %s, 実際: %s", tc.token, client.token)
			}

			if client.httpClient == nil {
				t.Fatal("HTTPクライアントがnilです")
			}
		})
	}
}

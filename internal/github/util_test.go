package github

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

// 'MockHTTPClient' struct is a mock for HTTP client.
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

// 'Do' method executes HTTP request.
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

// ヘルパー関数のテスト
func TestGetStringParam(t *testing.T) {
	tests := []struct {
		name     string
		args     map[string]interface{}
		key      string
		expected string
		ok       bool
	}{
		{
			name:     "存在するキー",
			args:     map[string]interface{}{"key": "value"},
			key:      "key",
			expected: "value",
			ok:       true,
		},
		{
			name:     "存在しないキー",
			args:     map[string]interface{}{"other": "value"},
			key:      "key",
			expected: "",
			ok:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := getStringParam(tt.args, tt.key)
			if ok != tt.ok {
				t.Errorf("getStringParam() ok = %v, want %v", ok, tt.ok)
			}
			if got != tt.expected {
				t.Errorf("getStringParam() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetRequiredStringParam(t *testing.T) {
	args := map[string]interface{}{"key": "value"}
	got := getRequiredStringParam(args, "key")
	if got != "value" {
		t.Errorf("getRequiredStringParam() = %v, want %v", got, "value")
	}

	// パニックのテスト
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("getRequiredStringParam() did not panic for missing key")
		}
	}()
	getRequiredStringParam(args, "missing")
}

func TestGetNumberParam(t *testing.T) {
	tests := []struct {
		name       string
		args       map[string]interface{}
		key        string
		defaultVal int
		expected   int
	}{
		{
			name:       "存在するキー",
			args:       map[string]interface{}{"key": float64(10)},
			key:        "key",
			defaultVal: 5,
			expected:   10,
		},
		{
			name:       "存在しないキー",
			args:       map[string]interface{}{"other": float64(10)},
			key:        "key",
			defaultVal: 5,
			expected:   5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getNumberParam(tt.args, tt.key, tt.defaultVal)
			if got != tt.expected {
				t.Errorf("getNumberParam() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetBoolParam(t *testing.T) {
	tests := []struct {
		name       string
		args       map[string]interface{}
		key        string
		defaultVal bool
		expected   bool
	}{
		{
			name:       "存在するキー（true）",
			args:       map[string]interface{}{"key": true},
			key:        "key",
			defaultVal: false,
			expected:   true,
		},
		{
			name:       "存在するキー（false）",
			args:       map[string]interface{}{"key": false},
			key:        "key",
			defaultVal: true,
			expected:   false,
		},
		{
			name:       "存在しないキー",
			args:       map[string]interface{}{"other": true},
			key:        "key",
			defaultVal: true,
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getBoolParam(tt.args, tt.key, tt.defaultVal)
			if got != tt.expected {
				t.Errorf("getBoolParam() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestReturnJSONResult(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		expectError bool
	}{
		{
			name:        "正常系 - 有効なJSONデータ",
			input:       map[string]interface{}{"key": "value"},
			expectError: false,
		},
		{
			name: "異常系 - JSONにマーシャルできないデータ",
			input: func() interface{} {
				// JSONにマーシャルできない循環参照を持つデータ構造
				type Circular struct {
					Self *Circular
				}
				c := &Circular{}
				c.Self = c
				return c
			}(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := returnJSONResult(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("returnJSONResult() エラーが期待されていましたが、エラーは発生しませんでした")
				}
				// エラーが発生した場合、resultはnilであるべき
				if result != nil {
					t.Errorf("returnJSONResult() エラー時にnilではない結果が返されました: %v", result)
				}
			} else {
				if err != nil {
					t.Errorf("returnJSONResult() error = %v", err)
					return
				}
				// 正常系の場合、resultはnilではないはず
				if result == nil {
					t.Errorf("returnJSONResult() 結果がnilです")
				}
			}
		})
	}
}

func TestAddToOptions(t *testing.T) {
	tests := []struct {
		name          string
		options       map[string]interface{}
		args          map[string]interface{}
		key           string
		expectedValue interface{}
		expectedExist bool
	}{
		{
			name:          "キーが存在する場合",
			options:       map[string]interface{}{},
			args:          map[string]interface{}{"key": "value"},
			key:           "key",
			expectedValue: "value",
			expectedExist: true,
		},
		{
			name:          "キーが存在しない場合",
			options:       map[string]interface{}{},
			args:          map[string]interface{}{"other": "value"},
			key:           "key",
			expectedValue: nil,
			expectedExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addToOptions(tt.options, tt.args, tt.key)
			value, exists := tt.options[tt.key]
			if exists != tt.expectedExist {
				t.Errorf("addToOptions() key exists = %v, want %v", exists, tt.expectedExist)
			}
			if exists && value != tt.expectedValue {
				t.Errorf("addToOptions() value = %v, want %v", value, tt.expectedValue)
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

// TestGitHubErrorError はGitHubError構造体のErrorメソッドをテストする
func TestGitHubErrorError(t *testing.T) {
	tests := []struct {
		name     string
		ghError  GitHubError
		expected string
	}{
		{
			name: "基本的なエラーメッセージ",
			ghError: GitHubError{
				Message:    "Not Found",
				StatusCode: 404,
			},
			expected: "GitHub API Error: Not Found (Status: 404)",
		},
		{
			name: "ドキュメントURLを含むエラー",
			ghError: GitHubError{
				Message:          "Validation Failed",
				DocumentationURL: "https://docs.github.com/rest/reference/issues",
				StatusCode:       422,
			},
			expected: "GitHub API Error: Validation Failed (Status: 422)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errorMsg := tt.ghError.Error()
			if errorMsg != tt.expected {
				t.Errorf("GitHubError.Error() = %v, want %v", errorMsg, tt.expected)
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

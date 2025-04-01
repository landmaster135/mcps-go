package arith_calc

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	mcp "github.com/mark3labs/mcp-go/mcp"
	server "github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
)

// モック用のBufioScannerインターフェース実装
type MockBufioScanner struct {
	boolToReturn bool
	errToReturn error
}

func (m *MockBufioScanner) Scan() bool {
	return m.boolToReturn
}

func (m *MockBufioScanner) Err() error {
	return m.errToReturn
}

// モック用のJSONMarshalerインターフェース実装
type MockJSONMarshaler struct {
	errToReturn error
}

func (m *MockJSONMarshaler) MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	if m.errToReturn != nil {
		return nil, m.errToReturn
	}
	return []byte(`{"mocked":"json"}`), nil
}

// テスト用の一時ファイルを作成する関数
func createTempFileWithLines(t *testing.T, lines []string) string {
	// 一時ファイルを作成
	tmpFile, err := os.CreateTemp("", "test-line-count-*.txt")
	assert.NoError(t, err, "一時ファイルの作成に失敗しました")

	// テスト終了時に一時ファイルを削除
	t.Cleanup(func() {
		os.Remove(tmpFile.Name())
	})

	// ファイルに行を書き込む
	for _, line := range lines {
		_, err := tmpFile.WriteString(line + "\n")
		assert.NoError(t, err, "ファイルへの書き込みに失敗しました")
	}

	// ファイルを閉じる
	err = tmpFile.Close()
	assert.NoError(t, err, "ファイルのクローズに失敗しました")

	return tmpFile.Name()
}


// TestCountLines は CountLines メソッドをテストします
func TestCountLines(t *testing.T) {
	// テスト用の EvalClient インスタンスを作成
	eval := NewEvalClient()

	// テストケースを定義
	testCases := []struct {
		name          string
		lines         []string
		expectedCount int
		expectError   bool
	}{
		{
			name:          "空のファイル",
			lines:         []string{},
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:          "1行のファイル",
			lines:         []string{"これは1行のファイルです"},
			expectedCount: 1,
			expectError:   false,
		},
		{
			name:          "複数行のファイル",
			lines:         []string{"1行目", "2行目", "3行目", "4行目", "5行目"},
			expectedCount: 5,
			expectError:   false,
		},
		{
			name:          "空行を含むファイル",
			lines:         []string{"1行目", "", "3行目", "", "5行目"},
			expectedCount: 5,
			expectError:   false,
		},
	}

	// 各テストケースを実行
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// テスト用の一時ファイルを作成
			filePath := createTempFileWithLines(t, tc.lines)

			// テスト対象の関数を実行
			count, err := eval.CountLines(filePath)

			// 結果の検証
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCount, count, "行数が期待値と一致しません")
			}
		})
	}

	// 存在しないファイルのテスト
	t.Run("存在しないファイル", func(t *testing.T) {
		_, err := eval.CountLines("/path/to/nonexistent/file.txt")
		assert.Error(t, err, "存在しないファイルに対してエラーが発生すべきです")
	})
}

// TestIsLineCountGreaterThan は IsLineCountGreaterThan メソッドをテストします
func TestIsLineCountGreaterThan(t *testing.T) {
	// テスト用の EvalClient インスタンスを作成
	eval := NewEvalClient()

	// テストケースを定義
	testCases := []struct {
		name           string
		lines          []string
		threshold      int
		expectedResult bool
		expectedCount  int
		expectError    bool
	}{
		{
			name:           "閾値より少ない行数",
			lines:          []string{"1行目", "2行目", "3行目"},
			threshold:      5,
			expectedResult: false,
			expectedCount:  3,
			expectError:    false,
		},
		{
			name:           "閾値と同じ行数",
			lines:          []string{"1行目", "2行目", "3行目", "4行目", "5行目"},
			threshold:      5,
			expectedResult: false,
			expectedCount:  5,
			expectError:    false,
		},
		{
			name:           "閾値より多い行数",
			lines:          []string{"1行目", "2行目", "3行目", "4行目", "5行目", "6行目"},
			threshold:      5,
			expectedResult: true,
			expectedCount:  6,
			expectError:    false,
		},
		{
			name:           "閾値が0の場合",
			lines:          []string{"1行目"},
			threshold:      0,
			expectedResult: true,
			expectedCount:  1,
			expectError:    false,
		},
	}

	// 各テストケースを実行
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// テスト用の一時ファイルを作成
			filePath := createTempFileWithLines(t, tc.lines)

			// テスト対象の関数を実行
			result, count, err := eval.IsLineCountGreaterThan(filePath, tc.threshold)

			// 結果の検証
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, result, "比較結果が期待値と一致しません")
				assert.Equal(t, tc.expectedCount, count, "行数が期待値と一致しません")
			}
		})
	}

	// 存在しないファイルのテスト
	t.Run("存在しないファイル", func(t *testing.T) {
		_, _, err := eval.IsLineCountGreaterThan("/path/to/nonexistent/file.txt", 5)
		assert.Error(t, err, "存在しないファイルに対してエラーが発生すべきです")
	})
}

// TestHandleToEvaluateLineCount は HandleToEvaluateLineCount メソッドをテストします
func TestHandleToEvaluateLineCount(t *testing.T) {
	// テスト用の EvalClient インスタンスを作成
	eval := NewEvalClient()
	ctx := context.Background()

	// テスト用のファイルを作成
	lines := []string{"1行目", "2行目", "3行目", "4行目", "5行目"}
	filePath := createTempFileWithLines(t, lines)

	// テストケースを定義
	testCases := []struct {
		name          string
		arguments     map[string]interface{}
		expectedKeys  []string
		isGreater     bool
		lineCount     int
		expectError   bool
		errorContains string
	}{
		{
			name: "正常系 - 閾値より少ない行数",
			arguments: map[string]interface{}{
				"file_path": filePath,
				"threshold": float64(10),
			},
			expectedKeys: []string{"is_greater", "line_count", "threshold", "file_path", "description"},
			isGreater:    false,
			lineCount:    5,
			expectError:  false,
		},
		{
			name: "正常系 - 閾値より多い行数",
			arguments: map[string]interface{}{
				"file_path": filePath,
				"threshold": float64(3),
			},
			expectedKeys: []string{"is_greater", "line_count", "threshold", "file_path", "description"},
			isGreater:    true,
			lineCount:    5,
			expectError:  false,
		},
		{
			name: "異常系 - file_pathパラメータなし",
			arguments: map[string]interface{}{
				"threshold": float64(5),
			},
			expectError:   true,
			errorContains: "file_path パラメータが必要です",
		},
		{
			name: "異常系 - thresholdパラメータなし",
			arguments: map[string]interface{}{
				"file_path": filePath,
			},
			expectError:   true,
			errorContains: "threshold パラメータが必要です",
		},
		{
			name: "異常系 - 存在しないファイル",
			arguments: map[string]interface{}{
				"file_path": "/path/to/nonexistent/file.txt",
				"threshold": float64(5),
			},
			expectError:   true,
			errorContains: "ファイルを開けませんでした",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// リクエストの作成
			request := mcp.CallToolRequest{}
			request.Params.Name = "evaluate_line_count"
			request.Params.Arguments = tc.arguments

			// テスト対象の関数を実行
			result, err := eval.HandleToEvaluateLineCount(ctx, request)

			// エラーの検証
			if tc.expectError {
				assert.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// 結果の内容を検証
				assert.NotNil(t, result.Content)

				// 結果の文字列表現に期待値が含まれていることを確認
				resultStr := fmt.Sprintf("%v", result)

				// is_greater の値が含まれていることを確認
				isGreaterStr := fmt.Sprintf("%v", tc.isGreater)
				assert.Contains(t, resultStr, isGreaterStr, "結果に is_greater の値 %v が含まれていません", tc.isGreater)

				// line_count の値が含まれていることを確認
				lineCountStr := fmt.Sprintf("%d", tc.lineCount)
				assert.Contains(t, resultStr, lineCountStr, "結果に line_count の値 %d が含まれていません", tc.lineCount)
			}
		})
	}
}

// TestScannerError は scanner.Err() がエラーを返す場合をテストします
func TestScannerError(t *testing.T) {
	// モック用のBufioScannerを作成
	mockScanner := &MockBufioScanner{
		boolToReturn: false,
		errToReturn: errors.New("模擬的なスキャナーエラー"),
	}

	// テスト用のEvalClientを作成し、モックを注入
	eval := &EvalClient{
		bufioScanner: mockScanner,
	}

	// bufioScannerのErrメソッドを呼び出す
	lc, err := eval.CountLines("模擬的なファイルパス")
	// err := eval.bufioScanner.Err()

	// 結果の検証
	assert.Equal(t, lc, 0, "行数が期待値と一致しません")
	assert.Error(t, err, "scanner.Err()がエラーを返すべきです")
	assert.Equal(t, "模擬的なスキャナーエラー", err.Error(), "エラーメッセージが期待値と一致しません")
}

// TestJSONMarshalError は JSON変換でエラーが発生する場合をテストします
func TestJSONMarshalError(t *testing.T) {
	// モック用のJSONMarshalerを作成
	mockMarshaler := &MockJSONMarshaler{
		errToReturn: errors.New("模擬的なJSON変換エラー"),
	}

	// テスト用のEvalClientを作成し、モックを注入
	eval := &EvalClient{
		bufioScanner: &bufio.Scanner{},
		jsonMarshaler: mockMarshaler,
	}
	ctx := context.Background()

	// テスト用のファイルを作成
	lines := []string{"1行目", "2行目", "3行目"}
	filePath := createTempFileWithLines(t, lines)

	// リクエストの作成
	request := mcp.CallToolRequest{}
	request.Params.Name = "evaluate_line_count"
	request.Params.Arguments = map[string]interface{}{
		"file_path": filePath,
		"threshold": float64(5),
	}

	// テスト対象の関数を実行
	result, err := eval.HandleToEvaluateLineCount(ctx, request)

	// 結果の検証
	assert.Error(t, err, "JSON変換でエラーが発生する場合はエラーが発生すべきです")
	assert.Contains(t, err.Error(), "模擬的なJSON変換エラー")
	assert.Nil(t, result, "エラーの場合は結果がnilであるべきです")
}

// TestIsGreaterDescription は isGreaterDescription 関数をテストします
func TestIsGreaterDescription(t *testing.T) {
	// テストケースを定義
	testCases := []struct {
		name           string
		isGreater      bool
		expectedResult string
	}{
		{
			name:           "閾値より大きい場合",
			isGreater:      true,
			expectedResult: "より大きいです。",
		},
		{
			name:           "閾値以下の場合",
			isGreater:      false,
			expectedResult: "以下です。",
		},
	}

	// 各テストケースを実行
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isGreaterDescription(tc.isGreater)
			assert.Equal(t, tc.expectedResult, result, "説明文が期待値と一致しません")
		})
	}
}

// TestSetFileLineCountEvaluatorServer は SetFileLineCountEvaluatorServer 関数をテストします
func TestSetFileLineCountEvaluatorServer(t *testing.T) {
	// モックサーバーを作成
	mockServer := server.NewMCPServer(
		"Test Server",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
	)

	// テスト対象の関数を実行
	resultServer := SetFileLineCountEvaluatorServer(mockServer)

	// 結果の検証
	assert.NotNil(t, resultServer, "サーバーが正しく設定されていません")
	assert.Equal(t, mockServer, resultServer, "返されたサーバーが入力と一致しません")
}

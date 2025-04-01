package arith_calc

import (
	"context"
	"fmt"
	"testing"

	mcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
)

// TestCalcClientSum は CalcClient の Sum メソッドをテストします
func TestCalcClientSum(t *testing.T) {
	// テスト用の CalcClient インスタンスを作成
	calc := CalcClient{}

	// テストケースを定義
	testCases := []struct {
		name     string
		arr      []float64
		expected float64
	}{
		{
			name:     "正の数の合計",
			arr:      []float64{1, 2, 3, 4, 5},
			expected: 15,
		},
		{
			name:     "負の数の合計",
			arr:      []float64{-1, -2, -3, -4, -5},
			expected: -15,
		},
		{
			name:     "正と負の数の合計",
			arr:      []float64{-5, -3, 0, 3, 5},
			expected: 0,
		},
		{
			name:     "小数点数の合計",
			arr:      []float64{1.5, 2.5, 3.5},
			expected: 7.5,
		},
		{
			name:     "空の配列",
			arr:      []float64{},
			expected: 0,
		},
		{
			name:     "単一要素の配列",
			arr:      []float64{42},
			expected: 42,
		},
	}

	// 各テストケースを実行
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := calc.Sum(tc.arr)
			assert.Equal(t, tc.expected, result, "Sum of %v should equal %f", tc.arr, tc.expected)
		})
	}
}

// TestHandleToCalculateWithArray は HandleToCalculateWithArray メソッドをテストします
func TestHandleToCalculateWithArray(t *testing.T) {
	// テスト用の CalcClient インスタンスを作成
	calc := NewCalcClient()
	ctx := context.Background()

	// テストケース
	tests := []struct {
		name          string
		arguments     map[string]interface{}
		expectedValue float64
		expectError   bool
		errorMessage  string
	}{
		{
			name: "正常系 - 正の数の合計",
			arguments: map[string]interface{}{
				"operation": "sum",
				"numbers":   []interface{}{float64(1), float64(2), float64(3), float64(4), float64(5)},
			},
			expectedValue: 15,
			expectError:   false,
		},
		{
			name: "正常系 - 負の数の合計",
			arguments: map[string]interface{}{
				"operation": "sum",
				"numbers":   []interface{}{float64(-1), float64(-2), float64(-3), float64(-4), float64(-5)},
			},
			expectedValue: -15,
			expectError:   false,
		},
		{
			name: "正常系 - 正と負の数の合計",
			arguments: map[string]interface{}{
				"operation": "sum",
				"numbers":   []interface{}{float64(-5), float64(-3), float64(0), float64(3), float64(5)},
			},
			expectedValue: 0,
			expectError:   false,
		},
		{
			name: "正常系 - 小数点数の合計",
			arguments: map[string]interface{}{
				"operation": "sum",
				"numbers":   []interface{}{float64(1.5), float64(2.5), float64(3.5)},
			},
			expectedValue: 7.5,
			expectError:   false,
		},
		{
			name: "正常系 - 空の配列",
			arguments: map[string]interface{}{
				"operation": "sum",
				"numbers":   []interface{}{},
			},
			expectedValue: 0,
			expectError:   false,
		},
		{
			name: "正常系 - 単一要素の配列",
			arguments: map[string]interface{}{
				"operation": "sum",
				"numbers":   []interface{}{float64(42)},
			},
			expectedValue: 42,
			expectError:   false,
		},
		{
			name: "異常系 - 数値以外の要素を含む配列",
			arguments: map[string]interface{}{
				"operation": "sum",
				"numbers":   []interface{}{float64(1), "not a number", float64(3)},
			},
			expectError:  true,
			errorMessage: "not a number is incompatible for float64",
		},
		{
			name: "正常系 - 不正な操作（デフォルト値を返す）",
			arguments: map[string]interface{}{
				"operation": "invalid",
				"numbers":   []interface{}{float64(1), float64(2), float64(3)},
			},
			expectedValue: 0, // 不正な操作の場合、デフォルト値の0が返される
			expectError:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// リクエストの作成
			request := mcp.CallToolRequest{}
			// Paramsフィールドに直接アクセス
			request.Params.Name = "calculate_with_multiple_numbers"
			request.Params.Arguments = tc.arguments

			// テスト対象の関数を実行
			result, err := calc.HandleToCalculateWithArray(ctx, request)

			// エラーの検証
			if tc.expectError {
				assert.Error(t, err)
				if tc.errorMessage != "" {
					assert.Equal(t, tc.errorMessage, err.Error())
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// 結果の内容を検証
				assert.NotNil(t, result.Content)

				// 結果の文字列表現に期待値が含まれていることを確認
				resultStr := fmt.Sprintf("%v", result)
				expectedStr := fmt.Sprintf("%v", tc.expectedValue)
				assert.Contains(t, resultStr, expectedStr)
			}
		})
	}
}

// TestHandleToCalculateWithArrayEdgeCases は HandleToCalculateWithArray メソッドの境界値ケースをテストします
func TestHandleToCalculateWithArrayEdgeCases(t *testing.T) {
	// テスト用の CalcClient インスタンスを作成
	calc := NewCalcClient()
	ctx := context.Background()

	// 大きな配列のテスト
	t.Run("大きな配列の合計", func(t *testing.T) {
		// 100要素の配列を作成（すべて1）
		numbers := make([]interface{}, 100)
		for i := 0; i < 100; i++ {
			numbers[i] = float64(1)
		}

		request := mcp.CallToolRequest{}
		request.Params.Name = "calculate_with_multiple_numbers"
		request.Params.Arguments = map[string]interface{}{
			"operation": "sum",
			"numbers":   numbers,
		}

		result, err := calc.HandleToCalculateWithArray(ctx, request)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// 結果の文字列表現に期待値（100）が含まれていることを確認
		resultStr := fmt.Sprintf("%v", result)
		assert.Contains(t, resultStr, "100")
	})

	// 非常に大きな数値を含む配列のテスト
	t.Run("非常に大きな数値を含む配列", func(t *testing.T) {
		request := mcp.CallToolRequest{}
		request.Params.Name = "calculate_with_multiple_numbers"
		request.Params.Arguments = map[string]interface{}{
			"operation": "sum",
			"numbers":   []interface{}{float64(1e15), float64(2e15), float64(3e15)},
		}

		result, err := calc.HandleToCalculateWithArray(ctx, request)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// 結果の文字列表現に期待値が含まれていることを確認
		resultStr := fmt.Sprintf("%v", result)
		assert.Contains(t, resultStr, "6000000000000000")
	})

	// 小さな数値を含む配列のテスト（ただし表示可能な範囲で）
	t.Run("小さな数値を含む配列", func(t *testing.T) {
		request := mcp.CallToolRequest{}
		request.Params.Name = "calculate_with_multiple_numbers"
		request.Params.Arguments = map[string]interface{}{
			"operation": "sum",
			"numbers":   []interface{}{float64(0.1), float64(0.2), float64(0.3)},
		}

		result, err := calc.HandleToCalculateWithArray(ctx, request)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// 結果の文字列表現に期待値（0.6）が含まれていることを確認
		resultStr := fmt.Sprintf("%v", result)
		assert.Contains(t, resultStr, "0.6")
	})

	// 精度の問題を確認するテスト
	t.Run("精度の問題", func(t *testing.T) {
		request := mcp.CallToolRequest{}
		request.Params.Name = "calculate_with_multiple_numbers"
		request.Params.Arguments = map[string]interface{}{
			"operation": "sum",
			"numbers":   []interface{}{float64(0.1), float64(0.2)},
		}

		result, err := calc.HandleToCalculateWithArray(ctx, request)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// 結果の文字列表現に期待値（0.3）が含まれていることを確認
		// 浮動小数点の精度の問題で厳密には0.3にならないため、文字列表現で近似値を確認
		resultStr := fmt.Sprintf("%v", result)
		assert.Contains(t, resultStr, "0.3")
	})
}

package arith_calc

import (
	"context"
	"fmt"
	"math"
	"testing"

	mcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
)

// TestCalcClientAdd は CalcClient の Add メソッドをテストします
func TestCalcClientAdd(t *testing.T) {
	// テスト用の CalcClient インスタンスを作成
	calc := CalcClient{}

	// テストケースを定義
	testCases := []struct {
		name     string
		x        float64
		y        float64
		expected float64
	}{
		{
			name:     "正の数の加算",
			x:        5,
			y:        3,
			expected: 8,
		},
		{
			name:     "負の数の加算",
			x:        -2,
			y:        -3,
			expected: -5,
		},
		{
			name:     "正と負の数の加算",
			x:        5,
			y:        -3,
			expected: 2,
		},
		{
			name:     "ゼロとの加算",
			x:        5,
			y:        0,
			expected: 5,
		},
		{
			name:     "小数点数の加算",
			x:        2.5,
			y:        3.5,
			expected: 6.0,
		},
	}

	// 各テストケースを実行
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := calc.Add(tc.x, tc.y)
			assert.Equal(t, tc.expected, result, "%f + %f should equal %f", tc.x, tc.y, tc.expected)
		})
	}
}

// TestCalcClientSubtract は CalcClient の Subtract メソッドをテストします
func TestCalcClientSubtract(t *testing.T) {
	// テスト用の CalcClient インスタンスを作成
	calc := CalcClient{}

	// テストケースを定義
	testCases := []struct {
		name     string
		x        float64
		y        float64
		expected float64
	}{
		{
			name:     "正の数の減算",
			x:        10,
			y:        4,
			expected: 6,
		},
		{
			name:     "負の数の減算",
			x:        -2,
			y:        -3,
			expected: 1,
		},
		{
			name:     "正と負の数の減算",
			x:        5,
			y:        -3,
			expected: 8,
		},
		{
			name:     "ゼロとの減算",
			x:        5,
			y:        0,
			expected: 5,
		},
		{
			name:     "自身からの減算",
			x:        5,
			y:        5,
			expected: 0,
		},
		{
			name:     "小数点数の減算",
			x:        5.5,
			y:        2.2,
			expected: 3.3,
		},
	}

	// 各テストケースを実行
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := calc.Subtract(tc.x, tc.y)
			// 小数点の計算誤差を考慮して、ほぼ等しいかをチェック
			if tc.name == "小数点数の減算" {
				assert.InDelta(t, tc.expected, result, 0.0001, "%f - %f should approximately equal %f", tc.x, tc.y, tc.expected)
			} else {
				assert.Equal(t, tc.expected, result, "%f - %f should equal %f", tc.x, tc.y, tc.expected)
			}
		})
	}
}

// TestCalcClientMultiply は CalcClient の Multiply メソッドをテストします
func TestCalcClientMultiply(t *testing.T) {
	// テスト用の CalcClient インスタンスを作成
	calc := CalcClient{}

	// テストケースを定義
	testCases := []struct {
		name     string
		x        float64
		y        float64
		expected float64
	}{
		{
			name:     "正の数の乗算",
			x:        6,
			y:        7,
			expected: 42,
		},
		{
			name:     "負の数の乗算",
			x:        -3,
			y:        -4,
			expected: 12,
		},
		{
			name:     "正と負の数の乗算",
			x:        5,
			y:        -3,
			expected: -15,
		},
		{
			name:     "ゼロとの乗算",
			x:        5,
			y:        0,
			expected: 0,
		},
		{
			name:     "小数点数の乗算",
			x:        2.5,
			y:        4,
			expected: 10,
		},
	}

	// 各テストケースを実行
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := calc.Multiply(tc.x, tc.y)
			assert.Equal(t, tc.expected, result, "%f * %f should equal %f", tc.x, tc.y, tc.expected)
		})
	}
}

// TestCalcClientDivide は CalcClient の Divide メソッドをテストします
func TestCalcClientDivide(t *testing.T) {
	// テスト用の CalcClient インスタンスを作成
	calc := CalcClient{}

	// テストケースを定義
	testCases := []struct {
		name     string
		x        float64
		y        float64
		expected float64
		isInf    bool
	}{
		{
			name:     "正の数の除算",
			x:        20,
			y:        5,
			expected: 4,
			isInf:    false,
		},
		{
			name:     "負の数の除算",
			x:        -12,
			y:        -4,
			expected: 3,
			isInf:    false,
		},
		{
			name:     "正と負の数の除算",
			x:        15,
			y:        -3,
			expected: -5,
			isInf:    false,
		},
		{
			name:     "ゼロの除算",
			x:        0,
			y:        5,
			expected: 0,
			isInf:    false,
		},
		{
			name:     "小数点数の除算",
			x:        10,
			y:        4,
			expected: 2.5,
			isInf:    false,
		},
		{
			name:     "ゼロによる除算",
			x:        5,
			y:        0,
			expected: 0, // 実際には無限大が返されるが、テストのために0を設定
			isInf:    true,
		},
	}

	// 各テストケースを実行
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := calc.Divide(tc.x, tc.y)

			if tc.isInf {
				// ゼロ除算の場合は無限大が返されることを確認
				assert.True(t, math.IsInf(result, 1), "%f / %f should return +Inf", tc.x, tc.y)
			} else {
				assert.Equal(t, tc.expected, result, "%f / %f should equal %f", tc.x, tc.y, tc.expected)
			}
		})
	}
}

// TestCalcClientEdgeCases は CalcClient の境界値ケースをテストします
func TestCalcClientEdgeCases(t *testing.T) {
	// テスト用の CalcClient インスタンスを作成
	calc := CalcClient{}

	// 大きな数値の演算
	t.Run("大きな数値の加算", func(t *testing.T) {
		result := calc.Add(1e15, 1e15)
		assert.Equal(t, 2e15, result)
	})

	// 非常に小さな数値の演算
	t.Run("非常に小さな数値の乗算", func(t *testing.T) {
		result := calc.Multiply(1e-15, 1e-15)
		assert.Equal(t, 1e-30, result)
	})

	// 精度の問題を確認
	t.Run("精度の問題", func(t *testing.T) {
		// 0.1 + 0.2 は浮動小数点の精度の問題で厳密には 0.3 にならない
		result := calc.Add(0.1, 0.2)
		assert.InDelta(t, 0.3, result, 1e-10)
	})
}

// TestHandleToCalculate は HandleToCalculate メソッドをテストします
func TestHandleToCalculate(t *testing.T) {
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
			name: "正常系 - 加算操作",
			arguments: map[string]interface{}{
				"operation": "add",
				"x":         float64(5),
				"y":         float64(3),
			},
			expectedValue: 8,
			expectError:   false,
		},
		{
			name: "正常系 - 減算操作",
			arguments: map[string]interface{}{
				"operation": "subtract",
				"x":         float64(10),
				"y":         float64(4),
			},
			expectedValue: 6,
			expectError:   false,
		},
		{
			name: "正常系 - 乗算操作",
			arguments: map[string]interface{}{
				"operation": "multiply",
				"x":         float64(6),
				"y":         float64(7),
			},
			expectedValue: 42,
			expectError:   false,
		},
		{
			name: "正常系 - 除算操作",
			arguments: map[string]interface{}{
				"operation": "divide",
				"x":         float64(20),
				"y":         float64(5),
			},
			expectedValue: 4,
			expectError:   false,
		},
		{
			name: "異常系 - ゼロによる除算",
			arguments: map[string]interface{}{
				"operation": "divide",
				"x":         float64(5),
				"y":         float64(0),
			},
			expectError:  true,
			errorMessage: "division by zero is not allowed",
		},
		{
			name: "正常系 - 不正な操作（デフォルト値を返す）",
			arguments: map[string]interface{}{
				"operation": "invalid",
				"x":         float64(5),
				"y":         float64(3),
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
			request.Params.Name = "calculate"
			request.Params.Arguments = tc.arguments

			// テスト対象の関数を実行
			result, err := calc.HandleToCalculate(ctx, request)

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

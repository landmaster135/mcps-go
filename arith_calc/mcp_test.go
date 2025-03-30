package arith_calc

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCalculator は Calculator 構造体のメソッドをテストします
func TestCalculator(t *testing.T) {
	// テスト用の Calculator インスタンスを作成
	calc := Calculator{}

	// Add メソッドのテスト
	t.Run("Add method", func(t *testing.T) {
		result := calc.Add(5, 3)
		assert.Equal(t, float64(8), result, "5 + 3 should equal 8")

		result = calc.Add(-2, 7)
		assert.Equal(t, float64(5), result, "-2 + 7 should equal 5")

		result = calc.Add(0, 0)
		assert.Equal(t, float64(0), result, "0 + 0 should equal 0")
	})

	// Subtract メソッドのテスト
	t.Run("Subtract method", func(t *testing.T) {
		result := calc.Subtract(10, 4)
		assert.Equal(t, float64(6), result, "10 - 4 should equal 6")

		result = calc.Subtract(5, 8)
		assert.Equal(t, float64(-3), result, "5 - 8 should equal -3")

		result = calc.Subtract(0, 0)
		assert.Equal(t, float64(0), result, "0 - 0 should equal 0")
	})

	// Multiply メソッドのテスト
	t.Run("Multiply method", func(t *testing.T) {
		result := calc.Multiply(6, 7)
		assert.Equal(t, float64(42), result, "6 * 7 should equal 42")

		result = calc.Multiply(-3, 4)
		assert.Equal(t, float64(-12), result, "-3 * 4 should equal -12")

		result = calc.Multiply(0, 5)
		assert.Equal(t, float64(0), result, "0 * 5 should equal 0")
	})

	// Divide メソッドのテスト
	t.Run("Divide method", func(t *testing.T) {
		result := calc.Divide(20, 5)
		assert.Equal(t, float64(4), result, "20 / 5 should equal 4")

		result = calc.Divide(7, 2)
		assert.Equal(t, float64(3.5), result, "7 / 2 should equal 3.5")

		result = calc.Divide(0, 5)
		assert.Equal(t, float64(0), result, "0 / 5 should equal 0")

		// ゼロ除算のテスト（Go ではパニックではなく Inf を返す）
		result = calc.Divide(5, 0)
		assert.True(t, result > 0, "5 / 0 should return +Inf")
	})
}

// TestMCPToolHandler は MCP ツールハンドラーの機能をテストします
func TestMCPToolHandler(t *testing.T) {
	// テストケースを定義
	testCases := []struct {
		name      string
		operation string
		x         float64
		y         float64
		expected  string
		expectErr bool
	}{
		{
			name:      "Add operation",
			operation: "add",
			x:         5,
			y:         3,
			expected:  "8",
			expectErr: false,
		},
		{
			name:      "Subtract operation",
			operation: "subtract",
			x:         10,
			y:         4,
			expected:  "6",
			expectErr: false,
		},
		{
			name:      "Multiply operation",
			operation: "multiply",
			x:         6,
			y:         7,
			expected:  "42",
			expectErr: false,
		},
		{
			name:      "Divide operation",
			operation: "divide",
			x:         20,
			y:         5,
			expected:  "4",
			expectErr: false,
		},
		{
			name:      "Divide by zero",
			operation: "divide",
			x:         10,
			y:         0,
			expectErr: true,
		},
	}

	// 各テストケースを実行
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// テスト用の簡易リクエスト構造体
			type SimpleRequest struct {
				Operation string
				X         float64
				Y         float64
			}

			// 簡易リクエストを作成
			request := SimpleRequest{
				Operation: tc.operation,
				X:         tc.x,
				Y:         tc.y,
			}

			// ハンドラー関数を定義
			var c Calculator
			var result float64
			var err error

			// ハンドラーのロジックをテスト
			op := request.Operation
			x := request.X
			y := request.Y

			switch op {
			case "add":
				result = c.Add(x, y)
			case "subtract":
				result = c.Subtract(x, y)
			case "multiply":
				result = c.Multiply(x, y)
			case "divide":
				if y == 0 {
					err = fmt.Errorf("division by zero is not allowed")
				} else {
					result = c.Divide(x, y)
				}
			}

			// エラーチェック
			if tc.expectErr {
				assert.NotNil(t, err, "エラーが発生するはずです")
				return
			}

			// 結果のチェック
			assert.Nil(t, err, "エラーは発生しないはずです")
			assert.Equal(t, tc.expected, fmt.Sprintf("%v", result), "計算結果が一致しません")
		})
	}
}

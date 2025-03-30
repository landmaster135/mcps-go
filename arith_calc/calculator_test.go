package arith_calc

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCalculatorAdd は Calculator の Add メソッドをテストします
func TestCalculatorAdd(t *testing.T) {
	// テスト用の Calculator インスタンスを作成
	calc := Calculator{}

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

// TestCalculatorSubtract は Calculator の Subtract メソッドをテストします
func TestCalculatorSubtract(t *testing.T) {
	// テスト用の Calculator インスタンスを作成
	calc := Calculator{}

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

// TestCalculatorMultiply は Calculator の Multiply メソッドをテストします
func TestCalculatorMultiply(t *testing.T) {
	// テスト用の Calculator インスタンスを作成
	calc := Calculator{}

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

// TestCalculatorDivide は Calculator の Divide メソッドをテストします
func TestCalculatorDivide(t *testing.T) {
	// テスト用の Calculator インスタンスを作成
	calc := Calculator{}

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

// TestCalculatorEdgeCases は Calculator の境界値ケースをテストします
func TestCalculatorEdgeCases(t *testing.T) {
	// テスト用の Calculator インスタンスを作成
	calc := Calculator{}

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

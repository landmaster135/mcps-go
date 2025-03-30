package arith_calc

// Calculator は基本的な算術演算を行うための構造体です
type Calculator struct {
	// 必要に応じてフィールドを追加できます
}

// Add は二つの数値を足し算するメソッドです
func (c *Calculator) Add(x float64, y float64) float64 {
	result := x + y
	return result
}

// Subtract は二つの数値を引き算するメソッドです
func (c *Calculator) Subtract(x float64, y float64) float64 {
	result := x - y
	return result
}

// Multiply は二つの数値を掛け算するメソッドです
func (c *Calculator) Multiply(x float64, y float64) float64 {
	result := x * y
	return result
}

// Divide は二つの数値を割り算するメソッドです
func (c *Calculator) Divide(x float64, y float64) float64 {
	result := x / y
	return result
}

package arith_calc

import (
	"testing"
)

func TestNewCalcClient(t *testing.T) {
	// テストケース
	tests := []struct {
		name string
	}{
		{
			name: "新しいCalcClientインスタンスの作成",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// テスト対象の関数を実行
			client := NewCalcClient()

			// クライアントの検証
			if client == nil {
				t.Fatal("クライアントがnilです")
			}

			// 型の検証
			_, ok := interface{}(client).(*CalcClient)
			if !ok {
				t.Errorf("クライアントの型が期待値と異なります。期待値: *CalcClient")
			}
		})
	}
}

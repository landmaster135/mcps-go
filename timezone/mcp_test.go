package timezone

import (
	"testing"
	"time"
)

// TestGetCurrentTime は GetCurrentTime 関数をテストします
func TestGetCurrentTime(t *testing.T) {
	service := &TimezoneService{}

	tests := []struct {
		name      string
		timezone  string
		wantError bool
	}{
		{
			name:      "有効なタイムゾーン: UTC",
			timezone:  "UTC",
			wantError: false,
		},
		{
			name:      "有効なタイムゾーン: Asia/Tokyo",
			timezone:  "Asia/Tokyo",
			wantError: false,
		},
		{
			name:      "有効なタイムゾーン: America/New_York",
			timezone:  "America/New_York",
			wantError: false,
		},
		{
			name:      "無効なタイムゾーン",
			timezone:  "Invalid/Timezone",
			wantError: true,
		},
		{
			name:      "空のタイムゾーン",
			timezone:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.GetCurrentTime(tt.timezone)
			if (err != nil) != tt.wantError {
				t.Errorf("GetCurrentTime() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError {
				// 結果が正しい形式かチェック
				_, err := time.Parse("2006-01-02 15:04:05", result)
				if err != nil {
					t.Errorf("GetCurrentTime() returned invalid time format: %v", result)
				}
			}
		})
	}
}

// TestIsValidTimezone は IsValidTimezone 関数をテストします
func TestIsValidTimezone(t *testing.T) {
	service := &TimezoneService{}

	tests := []struct {
		name     string
		timezone string
		want     bool
	}{
		{
			name:     "有効なタイムゾーン: UTC",
			timezone: "UTC",
			want:     true,
		},
		{
			name:     "有効なタイムゾーン: Asia/Tokyo",
			timezone: "Asia/Tokyo",
			want:     true,
		},
		{
			name:     "有効なタイムゾーン: America/New_York",
			timezone: "America/New_York",
			want:     true,
		},
		{
			name:     "無効なタイムゾーン",
			timezone: "Invalid/Timezone",
			want:     false,
		},
		{
			name:     "空のタイムゾーン",
			timezone: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.IsValidTimezone(tt.timezone)
			if got != tt.want {
				t.Errorf("IsValidTimezone() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestAvailableTimezones は availableTimezones 変数をテストします
func TestAvailableTimezones(t *testing.T) {
	// 重要なタイムゾーンが含まれているか確認
	expectedTimezones := []string{
		"UTC",
		"Asia/Tokyo",
		"America/New_York",
		"Europe/London",
	}

	for _, expected := range expectedTimezones {
		found := false
		for _, actual := range availableTimezones {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("availableTimezones does not contain %s", expected)
		}
	}

	// すべてのタイムゾーンが有効か確認
	service := &TimezoneService{}
	for _, tz := range availableTimezones {
		if !service.IsValidTimezone(tz) {
			t.Errorf("availableTimezones contains invalid timezone: %s", tz)
		}
	}
}

// TestBuildTimezoneServer は BuildTimezoneServer 関数をテストします
func TestBuildTimezoneServer(t *testing.T) {
	// このテストは BuildTimezoneServer 関数をテストする例です
	// 実際のテストでは、サーバーの起動と終了をテストする必要があります

	// 注意: このテストは実際には機能しません。
	// MCPサーバーのテストには特別な設定が必要です。
	// このコードはあくまで例として提供されています。
	t.Skip("このテストはサーバーテストの例として提供されており、実際には実行されません")
}

// TestMCPToolHandlers は MCP ツールハンドラーをテストします
func TestMCPToolHandlers(t *testing.T) {
	// このテストは MCP ツールハンドラーをテストする例です
	// 実際のテストでは、リクエストとレスポンスをモック化する必要があります

	// 注意: このテストは実際には機能しません。
	// MCP ツールハンドラーのテストには特別な設定が必要です。
	// このコードはあくまで例として提供されています。
	t.Skip("このテストは MCP ツールハンドラーテストの例として提供されており、実際には実行されません")
}

// 統合テストの例
func TestIntegration(t *testing.T) {
	// 統合テストは実際のサーバーを起動するため、通常のテスト実行では
	// スキップされるようにしています。実際にテストを実行する場合は、
	// 環境変数などを使用して制御することをお勧めします。
	t.Skip("統合テストはデフォルトでスキップされます")

	// サーバーを起動（実際のテストでは、別のゴルーチンで起動し、
	// テスト終了時にシャットダウンする必要があります）
	// go BuildTimezoneServer()

	// サーバーが起動するのを待つ
	// time.Sleep(100 * time.Millisecond)

	// ここでクライアントを使用してサーバーにリクエストを送信し、
	// レスポンスを検証します
}

// TestTimeFormatting は時刻のフォーマット機能をテストします
func TestTimeFormatting(t *testing.T) {
	// 特定の時刻を使用してテスト
	testTime := time.Date(2023, 1, 2, 15, 4, 5, 0, time.UTC)
	formatted := testTime.Format("2006-01-02 15:04:05")
	expected := "2023-01-02 15:04:05"

	if formatted != expected {
		t.Errorf("Time formatting failed, got: %s, want: %s", formatted, expected)
	}
}

// TestTimezoneConversion はタイムゾーン変換機能をテストします
func TestTimezoneConversion(t *testing.T) {
	// UTC の特定の時刻
	utcTime := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)

	// 東京のタイムゾーンに変換
	tokyoLoc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Fatalf("Failed to load Asia/Tokyo location: %v", err)
	}
	tokyoTime := utcTime.In(tokyoLoc)

	// 東京は UTC+9 なので、9時間進んでいるはず
	expectedHour := 9
	if tokyoTime.Hour() != expectedHour {
		t.Errorf("Timezone conversion failed, got hour: %d, want: %d", tokyoTime.Hour(), expectedHour)
	}
}

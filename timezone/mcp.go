package timezone

import (
	"context"
	"fmt"
	"strings"
	"time"

	mcp "github.com/mark3labs/mcp-go/mcp"
	server "github.com/mark3labs/mcp-go/server"
)

// 利用可能なタイムゾーンのリスト
// 実際の実装では、より完全なリストを使用することをお勧めします
var availableTimezones = []string{
	"UTC",
	"Asia/Tokyo",
	"America/New_York",
	"Europe/London",
	"Australia/Sydney",
	"Pacific/Auckland",
	"Europe/Paris",
	"Asia/Shanghai",
	"America/Los_Angeles",
	"America/Chicago",
	"Africa/Abidjan",
	"Asia/Bangkok",
	"Africa/Bissau",
	"Africa/Cairo",
	"Africa/Casablanca",
	"Africa/Ceuta",
	"Africa/Johannesburg",
	"Africa/Juba",
	"Africa/Khartoum",
	"Africa/Lagos",
	"Africa/Maputo",
	"Africa/Monrovia",
	"Africa/Nairobi",
	"Africa/Ndjamena",
	"Africa/Sao_Tome",
	"Africa/Tripoli",
	"Africa/Tunis",
	"Africa/Windhoek",
	"America/Adak",
	"America/Anchorage",
	"America/Araguaina",
	"America/Argentina/Cordoba",
	"America/Argentina/San_Luis",
	"America/Toronto",
	"Antarctica/Troll",
	"Antarctica/Vostok",
	"Asia/Dubai",
	"Asia/Hong_Kong",
	"Asia/Macau",
	"Asia/Pyongyang",
	"Asia/Qatar",
	"Asia/Seoul",
	"Asia/Singapore",
	"Asia/Taipei",
	"Europe/Berlin",
	"Europe/Helsinki",
	"Europe/Istanbul",
	"Europe/Moscow",
	"Europe/Rome",
	"Europe/Zurich",
	"Indian/Maldives",
	"Pacific/Auckland",
	"Pacific/Easter",
	"Pacific/Guam",
	"Pacific/Honolulu",
	"Pacific/Norfolk",
	"Pacific/Port_Moresby",
	"Pacific/Tahiti",
	"Pacific/Tarawa",
}

// TimezoneService はタイムゾーン関連の機能を提供する構造体です
type TimezoneService struct {
	// 必要に応じてフィールドを追加できます
}

// GetCurrentTime は指定されたタイムゾーンの現在時刻を取得するメソッドです
func (ts *TimezoneService) GetCurrentTime(timezone string) (string, error) {
	// 空のタイムゾーンをチェック
	if timezone == "" {
		return "", fmt.Errorf("empty timezone is not allowed")
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return "", fmt.Errorf("invalid timezone: %s", timezone)
	}

	now := time.Now().In(loc)
	return now.Format("2006-01-02 15:04:05"), nil
}

// IsValidTimezone はタイムゾーンが有効かどうかを確認するメソッドです
func (ts *TimezoneService) IsValidTimezone(timezone string) bool {
	// 空のタイムゾーンは無効とする
	if timezone == "" {
		return false
	}

	_, err := time.LoadLocation(timezone)
	return err == nil
}

// BuildTimezoneServer はタイムゾーンMCPサーバーを構築する関数です
func BuildTimezoneServer() {
	s := server.NewMCPServer(
		"Timezone Service",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
	)

	// ツールの設定
	tool := mcp.NewTool("get-current-timezone",
		mcp.WithDescription("Get the current time in the specified timezone"),
		mcp.WithString("timezone",
			mcp.Required(),
			mcp.Description("The timezone to get the current time for"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		timezone := request.Params.Arguments["timezone"].(string)

		var ts TimezoneService
		if !ts.IsValidTimezone(timezone) {
			return nil, fmt.Errorf("invalid timezone: %s", timezone)
		}

		timeStr, err := ts.GetCurrentTime(timezone)
		if err != nil {
			return nil, err
		}

		return mcp.NewToolResultText(fmt.Sprintf("The current time in %s is %s", timezone, timeStr)), nil
	})

	// 利用可能なタイムゾーンを取得するツール
	listTool := mcp.NewTool("list-available-timezones",
		mcp.WithDescription("List all available timezones"),
	)

	s.AddTool(listTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var ts TimezoneService
		var validTimezones []string

		// 利用可能なタイムゾーンのリストを作成
		for _, tz := range availableTimezones {
			if ts.IsValidTimezone(tz) {
				validTimezones = append(validTimezones, tz)
			}
		}

		return mcp.NewToolResultText(fmt.Sprintf("Available timezones: %s", strings.Join(validTimezones, ", "))), nil
	})

	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

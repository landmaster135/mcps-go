package datetime_calc

import (
	"context"
	"fmt"

	mcp "github.com/mark3labs/mcp-go/mcp"
	server "github.com/mark3labs/mcp-go/server"
)

func addPromptIntoServer(s *server.MCPServer) *server.MCPServer {
	prompt := mcp.NewPrompt("system_prompt_01",
		mcp.WithPromptDescription("This is a datetime calculator prompt"),
	)
	s.AddPrompt(prompt, func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return &mcp.GetPromptResult{
			Description: "System prompt for datetime calculator.",
			Messages: []mcp.PromptMessage{
				{
					Role:    mcp.RoleAssistant,
					Content: mcp.NewTextContent("You use this accurate calculator well."),
				},
			},
		}, nil
	})
	return s
}

func BuildTimeCalculatorServer() {
	s := server.NewMCPServer(
		"Time Calculator",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
		server.WithLogging(),
	)
	tool := mcp.NewTool("datetime_calc",
		mcp.WithDescription("Perform basic time calculations"),
		mcp.WithString("operation",
			mcp.Required(),
			mcp.Description("The operation to calculate datetime"),
			mcp.Enum("add", "subtract"),
			// mcp.Enum("add", "subtract"),
		),
		mcp.WithNumber("year_1",
			mcp.Required(),
			mcp.Description("Year of first datetime"),
		),
		mcp.WithNumber("month_1",
			mcp.Required(),
			mcp.Description("Month of first datetime"),
		),
		mcp.WithNumber("day_1",
			mcp.Required(),
			mcp.Description("Day of first datetime"),
		),
		mcp.WithNumber("hour_1",
			mcp.Required(),
			mcp.Description("Hour of first datetime"),
		),
		mcp.WithNumber("minute_1",
			mcp.Required(),
			mcp.Description("Minute of first datetime"),
		),
		mcp.WithNumber("second_1",
			mcp.Required(),
			mcp.Description("Second of first datetime"),
		),
		mcp.WithNumber("duration_of_year",
			mcp.Required(),
			mcp.Description("Duration of year"),
		),
		mcp.WithNumber("duration_of_month",
			mcp.Required(),
			mcp.Description("Duration of month"),
		),
		mcp.WithNumber("duration_of_day",
			mcp.Required(),
			mcp.Description("Duration of day"),
		),
		mcp.WithNumber("duration_of_hour",
			mcp.Required(),
			mcp.Description("Duration of hour"),
		),
		mcp.WithNumber("duration_of_minute",
			mcp.Required(),
			mcp.Description("Duration of minute"),
		),
		mcp.WithNumber("duration_of_second",
			mcp.Required(),
			mcp.Description("Duration of second"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		op := request.Params.Arguments["operation"].(string)
		year_1 := request.Params.Arguments["year_1"].(float64)
		month_1 := request.Params.Arguments["month_1"].(float64)
		day_1 := request.Params.Arguments["day_1"].(float64)
		hour_1 := request.Params.Arguments["hour_1"].(float64)
		minute_1 := request.Params.Arguments["minute_1"].(float64)
		second_1 := request.Params.Arguments["second_1"].(float64)
		duration_of_year := request.Params.Arguments["duration_of_year"].(float64)
		duration_of_month := request.Params.Arguments["duration_of_month"].(float64)
		duration_of_day := request.Params.Arguments["duration_of_day"].(float64)
		duration_of_hour := request.Params.Arguments["duration_of_hour"].(float64)
		duration_of_minute := request.Params.Arguments["duration_of_minute"].(float64)
		duration_of_second := request.Params.Arguments["duration_of_second"].(float64)

		var c DatetimeCalculator
		var result string
		switch op {
		case "add":
			result = c.AddDatetimeFloat(year_1, month_1, day_1, hour_1, minute_1, second_1, duration_of_year, duration_of_month, duration_of_day, duration_of_hour, duration_of_minute, duration_of_second)
		case "subtract":
			result = c.SubtractDatetimeFloat(year_1, month_1, day_1, hour_1, minute_1, second_1, duration_of_year, duration_of_month, duration_of_day, duration_of_hour, duration_of_minute, duration_of_second)
			// case "diff":
			// 	result = c.DiffTime(x, y)
		}

		return mcp.NewToolResultText(result), nil
	})

	// プロンプト
	s = addPromptIntoServer(s)

	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

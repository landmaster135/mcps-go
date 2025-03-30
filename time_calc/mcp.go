package time_calc

import (
	"context"
	"fmt"
	"time"

	mcp "github.com/mark3labs/mcp-go/mcp"
	server "github.com/mark3labs/mcp-go/server"
)

func BuildTimeCalculatorServer() {
	s := server.NewMCPServer(
		"Time Calculator",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
	)
	timeCalculatorTool := mcp.NewTool("time_calculate",
		mcp.WithDescription("Perform basic time calculations"),
		mcp.WithString("operation",
			mcp.Required(),
			mcp.Description("The operation calculating time"),
			mcp.Enum("add", "subtract"),
		),
		mcp.WithNumber("x",
			mcp.Required(),
			mcp.Description("First time"),
		),
		mcp.WithNumber("y",
			mcp.Required(),
			mcp.Description("Second time"),
		),
	)

	s.AddTool(timeCalculatorTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		op := request.Params.Arguments["operation"].(string)
		x := request.Params.Arguments["x"].(string)
		y := request.Params.Arguments["y"].(time.Duration)

		var c TimeCalculator
		var result string
		switch op {
		case "add":
			result = c.AddDuration(x, y)
		case "subtract":
			result = c.SubtractDuration(x, y)
			// case "diff":
			// 	result = c.DiffTime(x, y)
		}

		return mcp.NewToolResultText(result), nil
	})

	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

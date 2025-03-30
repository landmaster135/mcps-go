package arith_calc

import (
	"context"
	"errors"
	"fmt"

	mcp "github.com/mark3labs/mcp-go/mcp"
	server "github.com/mark3labs/mcp-go/server"
)

func BuildCalculatorServer() {
	s := server.NewMCPServer(
		"Arithmetic Calculator",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
	)
	calculatorTool := mcp.NewTool("calculate",
		mcp.WithDescription("Perform basic arithmetic calculations"),
		mcp.WithString("operation",
			mcp.Required(),
			mcp.Description("The arithmetic operation to perform"),
			mcp.Enum("add", "subtract", "multiply", "divide"),
		),
		mcp.WithNumber("x",
			mcp.Required(),
			mcp.Description("First number"),
		),
		mcp.WithNumber("y",
			mcp.Required(),
			mcp.Description("Second number"),
		),
	)
	s.AddTool(calculatorTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		op := request.Params.Arguments["operation"].(string)
		x := request.Params.Arguments["x"].(float64)
		y := request.Params.Arguments["y"].(float64)

		var c Calculator
		var result float64
		switch op {
		case "add":
			result = c.Add(x, y)
		case "subtract":
			result = c.Subtract(x, y)
		case "multiply":
			result = c.Multiply(x, y)
		case "divide":
			if y == 0 {
				return nil, errors.New("division by zero is not allowed")
			}
			result = c.Divide(x, y)
		}

		return mcp.FormatNumberResult(result), nil
	})

	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

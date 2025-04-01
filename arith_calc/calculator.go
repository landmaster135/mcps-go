package arith_calc

import (
	"context"
	"errors"

	mcp "github.com/mark3labs/mcp-go/mcp"
	server "github.com/mark3labs/mcp-go/server"
)

// Add は二つの数値を足し算するメソッドです
func (c *CalcClient) Add(x float64, y float64) float64 {
	result := x + y
	return result
}

// Subtract は二つの数値を引き算するメソッドです
func (c *CalcClient) Subtract(x float64, y float64) float64 {
	result := x - y
	return result
}

// Multiply は二つの数値を掛け算するメソッドです
func (c *CalcClient) Multiply(x float64, y float64) float64 {
	result := x * y
	return result
}

// Divide は二つの数値を割り算するメソッドです
func (c *CalcClient) Divide(x float64, y float64) float64 {
	result := x / y
	return result
}

func (c *CalcClient) HandleToCalculate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	op := request.Params.Arguments["operation"].(string)
	x := request.Params.Arguments["x"].(float64)
	y := request.Params.Arguments["y"].(float64)

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
}

func SetTwoNumbersInputtingCalcServer(s *server.MCPServer) *server.MCPServer {
	// Calcクライアントを初期化
	client := NewCalcClient()

	tool := mcp.NewTool("calculate",
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
	s.AddTool(tool, client.HandleToCalculate)

	return s
}

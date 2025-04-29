package arith_calc

import (
	"context"
	"fmt"

	mcp "github.com/mark3labs/mcp-go/mcp"
	server "github.com/mark3labs/mcp-go/server"
)

func addPromptIntoServer(s *server.MCPServer) *server.MCPServer {
	prompt := mcp.NewPrompt("system_prompt_01",
		mcp.WithPromptDescription("This is an arithmetic calculator prompt"),
	)
	s.AddPrompt(prompt, func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return &mcp.GetPromptResult{
			Description: "System prompt for arithmetic calculator.",
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

func createArithCalcServer() *server.MCPServer {
	s := server.NewMCPServer(
		"Arithmetic Calculator",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
		server.WithLogging(),
	)
	s = SetTwoNumbersInputtingCalcServer(s)
	s = SetFileLineCountEvaluatorServer(s)
	s = addPromptIntoServer(s)
	return s
}

func BuildArithCalculatorServer() {
	s := createArithCalcServer()
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

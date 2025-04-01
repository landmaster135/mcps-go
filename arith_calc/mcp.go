package arith_calc

import (
	"fmt"

	server "github.com/mark3labs/mcp-go/server"
)

func createArithCalcServer() *server.MCPServer {
	s := server.NewMCPServer(
		"Arithmetic Calculator",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
	)
	s = SetTwoNumbersInputtingCalcServer(s)
	return s
}

func BuildArithCalculatorServer() {
	s := createArithCalcServer()
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

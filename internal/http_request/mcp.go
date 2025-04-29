package http_request

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	mcp "github.com/mark3labs/mcp-go/mcp"
	server "github.com/mark3labs/mcp-go/server"
)

func BuildMcpServer() {
	s := server.NewMCPServer(
		"HTTP requestor",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
	)
	tool := mcp.NewTool("http_request",
		mcp.WithDescription("Make HTTP requests to external APIs"),
		mcp.WithString("method",
			mcp.Required(),
			mcp.Description("HTTP method to use"),
			mcp.Enum("GET", "POST", "PUT", "DELETE"),
		),
		mcp.WithString("url",
			mcp.Required(),
			mcp.Description("URL to send the request to"),
			mcp.Pattern("^https?://.*"),
		),
		mcp.WithString("body",
			mcp.Description("Request body (for POST/PUT)"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		method := request.Params.Arguments["method"].(string)
		url := request.Params.Arguments["url"].(string)
		body := ""
		if b, ok := request.Params.Arguments["body"].(string); ok {
			body = b
		}

		// Create and send request
		var req *http.Request
		var err error
		if body != "" {
			req, err = http.NewRequest(method, url, strings.NewReader(body))
		} else {
			req, err = http.NewRequest(method, url, nil)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %v", err)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request failed: %v", err)
		}
		defer resp.Body.Close()

		// Return response
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %v", err)
		}

		return mcp.NewToolResultText(fmt.Sprintf("Status: %d\nBody: %s", resp.StatusCode, string(respBody))), nil
	})

	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

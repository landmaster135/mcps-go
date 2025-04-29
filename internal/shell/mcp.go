package shell

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	mcp "github.com/mark3labs/mcp-go/mcp"
	server "github.com/mark3labs/mcp-go/server"
)

// ShellClient はシェルコマンド実行クライアントの構造体です
type ShellClient struct {
	executor CommandExecutor
}

// NewShellClient は新しいShellClientを作成します
func NewShellClient() *ShellClient {
	// 環境変数からベースディレクトリを取得（指定されていない場合はカレントディレクトリ）
	baseDir := os.Getenv("SHELL_BASE_DIRECTORY")
	if baseDir == "" {
		var err error
		baseDir, err = os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current directory: %v\n", err)
			baseDir = "."
		}
	}

	return &ShellClient{
		executor: NewShellExecutor(baseDir),
	}
}

// HandleShellExecute はシェルコマンド実行ツールのハンドラーです
func (c *ShellClient) HandleShellExecute(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 必須パラメータの取得
	command, ok := request.Params.Arguments["command"].(string)
	if !ok {
		return nil, fmt.Errorf("command パラメータが必要です")
	}

	// オプションパラメータの取得
	var args []string
	if argsVal, ok := request.Params.Arguments["args"]; ok {
		if argsArray, ok := argsVal.([]interface{}); ok {
			args = make([]string, len(argsArray))
			for i, v := range argsArray {
				args[i] = fmt.Sprintf("%v", v)
			}
		}
	}

	cwd, _ := getStringParam(request.Params.Arguments, "cwd")

	var env map[string]string
	if envVal, ok := request.Params.Arguments["env"]; ok {
		if envMap, ok := envVal.(map[string]interface{}); ok {
			env = make(map[string]string)
			for k, v := range envMap {
				env[k] = fmt.Sprintf("%v", v)
			}
		}
	}

	timeout := 0
	if timeoutVal, ok := request.Params.Arguments["timeout"]; ok {
		if t, ok := timeoutVal.(float64); ok {
			timeout = int(t)
		}
	}

	// コマンドを実行
	result, err := c.executor.Execute(command, args, cwd, env, timeout)
	if err != nil {
		return nil, err
	}

	// 結果を返却
	if !result.Success {
		return createToolResult(result.Stderr, true)
	}

	return createToolResult(result.Stdout, false)
}

// HandleGetAllowedCommands は許可されたコマンドのリストを取得するハンドラーです
func (c *ShellClient) HandleGetAllowedCommands(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 許可されたコマンドのリストを取得
	commands := c.executor.(*ShellExecutor).GetAllowedCommands()

	// 結果を作成
	result := map[string]interface{}{
		"commands": commands,
	}

	// JSON形式で結果を返す
	jsonResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(string(jsonResult)), nil
}

// SetShellCommandServer はシェルコマンド実行ツールを提供するMCPサーバを設定します
func SetShellCommandServer(s *server.MCPServer) *server.MCPServer {
	// ShellClientを初期化
	client := NewShellClient()

	// ツール: シェルコマンド実行
	shellExecuteTool := mcp.NewTool("shell_execute",
		mcp.WithDescription("シェルコマンドを実行します"),
		mcp.WithString("command",
			mcp.Required(),
			mcp.Description("実行するメインコマンド"),
		),
		mcp.WithArray("args",
			mcp.Description("コマンド引数（サブコマンド以降：配列形式）"),
		),
		mcp.WithString("cwd",
			mcp.Description("作業ディレクトリ"),
		),
		mcp.WithObject("env",
			mcp.Description("環境変数"),
		),
		mcp.WithNumber("timeout",
			mcp.Description("タイムアウト（ミリ秒）"),
		),
	)
	s.AddTool(shellExecuteTool, client.HandleShellExecute)

	// ツール: 許可されたコマンドのリストを取得
	getAllowedCommandsTool := mcp.NewTool("shell_get_allowed_commands",
		mcp.WithDescription("許可されたシェルコマンドのリストを取得します"),
	)
	s.AddTool(getAllowedCommandsTool, client.HandleGetAllowedCommands)

	return s
}

// createShellServer はシェル操作用のMCPサーバーを作成します
func createShellServer() *server.MCPServer {
	// MCPサーバーを作成
	s := server.NewMCPServer(
		"Shell Command Server",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
	)

	// シェルコマンド実行ツールを登録
	s = SetShellCommandServer(s)

	return s
}

// BuildShellServer はシェル操作用のMCPサーバーを構築します
func BuildShellServer() {
	s := createShellServer()

	// サーバーを起動
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

package arith_calc

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	mcp "github.com/mark3labs/mcp-go/mcp"
	server "github.com/mark3labs/mcp-go/server"
)

// BufioScanner インターフェースを定義
type BufioScanner interface {
	Scan() (bool)
	Err() (error)
}

// JSONMarshaler インターフェースを定義
type JSONMarshaler interface {
	MarshalIndent(v interface{}, prefix, indent string) ([]byte, error)
}

// DefaultJSONMarshaler は標準のjson.MarshalIndentを使用する実装
type DefaultJSONMarshaler struct{}

func (m *DefaultJSONMarshaler) MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(v, prefix, indent)
}

// EvalClient はファイル評価クライアントの構造体です
type EvalClient struct{
	bufioScanner BufioScanner
	jsonMarshaler JSONMarshaler
}

// NewEvalClient は新しいEvalClientを作成します
func NewEvalClient() *EvalClient {
	return &EvalClient{
		bufioScanner: &bufio.Scanner{},
		jsonMarshaler: &DefaultJSONMarshaler{},
	}
}

// CountLines はファイルの行数をカウントするメソッドです
func (e *EvalClient) CountLines(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("ファイルを開けませんでした: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	e.bufioScanner = scanner
	lineCount := 0
	for e.bufioScanner.Scan() {
		lineCount++
	}

	if err := e.bufioScanner.Err(); err != nil {
		return 0, fmt.Errorf("ファイルの読み取り中にエラーが発生しました: %w", err)
	}

	return lineCount, nil
}

// IsLineCountGreaterThan はファイルの行数が指定された数値より大きいかどうかを判定するメソッドです
func (e *EvalClient) IsLineCountGreaterThan(filePath string, threshold int) (bool, int, error) {
	lineCount, err := e.CountLines(filePath)
	if err != nil {
		return false, 0, err
	}

	return lineCount > threshold, lineCount, nil
}

// HandleToEvaluateLineCount はファイルの行数が指定された数値より大きいかどうかを判定し、結果を返すハンドラーです
func (e *EvalClient) HandleToEvaluateLineCount(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 必須パラメータの取得
	filePath, ok := request.Params.Arguments["file_path"].(string)
	if !ok {
		return nil, errors.New("file_path パラメータが必要です")
	}

	threshold, ok := request.Params.Arguments["threshold"].(float64)
	if !ok {
		return nil, errors.New("threshold パラメータが必要です")
	}

	// 行数の評価
	isGreater, lineCount, err := e.IsLineCountGreaterThan(filePath, int(threshold))
	if err != nil {
		return nil, err
	}

	// 結果の作成
	result := map[string]interface{}{
		"is_greater":  isGreater,
		"line_count":  lineCount,
		"threshold":   int(threshold),
		"file_path":   filePath,
		"description": fmt.Sprintf("ファイル '%s' の行数は %d 行で、閾値 %d %s", filePath, lineCount, int(threshold), isGreaterDescription(isGreater)),
	}

	// JSON形式で結果を返す
	jsonResult, err := e.jsonMarshaler.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(string(jsonResult)), nil
}

// isGreaterDescription は比較結果に基づいた説明文を返します
func isGreaterDescription(isGreater bool) string {
	if isGreater {
		return "より大きいです。"
	}
	return "以下です。"
}

// SetFileLineCountEvaluatorServer はファイルの行数評価ツールを提供するMCPサーバを設定します
func SetFileLineCountEvaluatorServer(s *server.MCPServer) *server.MCPServer {
	// EvalClientを初期化
	client := NewEvalClient()

	// ツール: ファイルの行数評価
	tool := mcp.NewTool("evaluate_line_count",
		mcp.WithDescription("ファイルの行数が指定された閾値より大きいかどうかを評価します"),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("評価するファイルのパス"),
		),
		mcp.WithNumber("threshold",
			mcp.Required(),
			mcp.Description("比較する行数の閾値"),
		),
	)
	s.AddTool(tool, client.HandleToEvaluateLineCount)

	return s
}

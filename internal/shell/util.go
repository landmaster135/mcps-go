package shell

import (
	"encoding/json"
	"fmt"

	mcp "github.com/mark3labs/mcp-go/mcp"
)

// ヘルパー関数: 文字列パラメータを取得
func getStringParam(args map[string]interface{}, key string) (string, bool) {
	if val, ok := args[key]; ok {
		return val.(string), true
	}
	return "", false
}

// ヘルパー関数: 必須の文字列パラメータを取得
func getRequiredStringParam(args map[string]interface{}, key string) string {
	return args[key].(string)
}

// ヘルパー関数: 数値パラメータを取得
func getNumberParam(args map[string]interface{}, key string, defaultVal int) int {
	if val, ok := args[key]; ok {
		return int(val.(float64))
	}
	return defaultVal
}

// ヘルパー関数: ブールパラメータを取得
func getBoolParam(args map[string]interface{}, key string, defaultVal bool) bool {
	if val, ok := args[key]; ok {
		return val.(bool)
	}
	return defaultVal
}

// ヘルパー関数: 文字列配列パラメータを取得
func getStringArrayParam(args map[string]interface{}, key string) ([]string, bool) {
	if val, ok := args[key]; ok {
		interfaceArray, ok := val.([]interface{})
		if !ok {
			return []string{}, false
		}

		stringArray := make([]string, len(interfaceArray))
		for i, v := range interfaceArray {
			stringArray[i] = fmt.Sprintf("%v", v)
		}
		return stringArray, true
	}
	return []string{}, false
}

// ヘルパー関数: マップパラメータを取得
func getMapParam(args map[string]interface{}, key string) (map[string]string, bool) {
	if val, ok := args[key]; ok {
		interfaceMap, ok := val.(map[string]interface{})
		if !ok {
			return map[string]string{}, false
		}

		stringMap := make(map[string]string)
		for k, v := range interfaceMap {
			stringMap[k] = fmt.Sprintf("%v", v)
		}
		return stringMap, true
	}
	return map[string]string{}, false
}

// ヘルパー関数: 結果をJSON形式で返却
func returnJSONResult(result interface{}) (*mcp.CallToolResult, error) {
	jsonResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(string(jsonResult)), nil
}

// ヘルパー関数: エラー結果を返却
func returnErrorResult(message string) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText(fmt.Sprintf("Error: %s", message)), nil
}

// ヘルパー関数: ツール結果を作成
func createToolResult(content string, isError bool) (*mcp.CallToolResult, error) {
	if isError {
		return mcp.NewToolResultText(fmt.Sprintf("Error: %s", content)), nil
	}
	return mcp.NewToolResultText(content), nil
}

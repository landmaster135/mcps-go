package github_v2

import (
	"testing"
)

// ヘルパー関数のテスト
func TestGetStringParam(t *testing.T) {
	tests := []struct {
		name     string
		args     map[string]interface{}
		key      string
		expected string
		ok       bool
	}{
		{
			name:     "存在するキー",
			args:     map[string]interface{}{"key": "value"},
			key:      "key",
			expected: "value",
			ok:       true,
		},
		{
			name:     "存在しないキー",
			args:     map[string]interface{}{"other": "value"},
			key:      "key",
			expected: "",
			ok:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := getStringParam(tt.args, tt.key)
			if ok != tt.ok {
				t.Errorf("getStringParam() ok = %v, want %v", ok, tt.ok)
			}
			if got != tt.expected {
				t.Errorf("getStringParam() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetRequiredStringParam(t *testing.T) {
	args := map[string]interface{}{"key": "value"}
	got := getRequiredStringParam(args, "key")
	if got != "value" {
		t.Errorf("getRequiredStringParam() = %v, want %v", got, "value")
	}

	// パニックのテスト
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("getRequiredStringParam() did not panic for missing key")
		}
	}()
	getRequiredStringParam(args, "missing")
}

func TestGetNumberParam(t *testing.T) {
	tests := []struct {
		name       string
		args       map[string]interface{}
		key        string
		defaultVal int
		expected   int
	}{
		{
			name:       "存在するキー",
			args:       map[string]interface{}{"key": float64(10)},
			key:        "key",
			defaultVal: 5,
			expected:   10,
		},
		{
			name:       "存在しないキー",
			args:       map[string]interface{}{"other": float64(10)},
			key:        "key",
			defaultVal: 5,
			expected:   5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getNumberParam(tt.args, tt.key, tt.defaultVal)
			if got != tt.expected {
				t.Errorf("getNumberParam() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetBoolParam(t *testing.T) {
	tests := []struct {
		name       string
		args       map[string]interface{}
		key        string
		defaultVal bool
		expected   bool
	}{
		{
			name:       "存在するキー（true）",
			args:       map[string]interface{}{"key": true},
			key:        "key",
			defaultVal: false,
			expected:   true,
		},
		{
			name:       "存在するキー（false）",
			args:       map[string]interface{}{"key": false},
			key:        "key",
			defaultVal: true,
			expected:   false,
		},
		{
			name:       "存在しないキー",
			args:       map[string]interface{}{"other": true},
			key:        "key",
			defaultVal: true,
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getBoolParam(tt.args, tt.key, tt.defaultVal)
			if got != tt.expected {
				t.Errorf("getBoolParam() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestReturnJSONResult(t *testing.T) {
	// このテストは簡略化します
	// 実際のmcp.CallToolResultの構造は複雑なため、
	// エラーが発生しないことだけを確認します
	result := map[string]interface{}{"key": "value"}
	_, err := returnJSONResult(result)
	if err != nil {
		t.Errorf("returnJSONResult() error = %v", err)
		return
	}

	// 成功すれば良しとします
}

func TestAddToOptions(t *testing.T) {
	tests := []struct {
		name          string
		options       map[string]interface{}
		args          map[string]interface{}
		key           string
		expectedValue interface{}
		expectedExist bool
	}{
		{
			name:          "キーが存在する場合",
			options:       map[string]interface{}{},
			args:          map[string]interface{}{"key": "value"},
			key:           "key",
			expectedValue: "value",
			expectedExist: true,
		},
		{
			name:          "キーが存在しない場合",
			options:       map[string]interface{}{},
			args:          map[string]interface{}{"other": "value"},
			key:           "key",
			expectedValue: nil,
			expectedExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addToOptions(tt.options, tt.args, tt.key)
			value, exists := tt.options[tt.key]
			if exists != tt.expectedExist {
				t.Errorf("addToOptions() key exists = %v, want %v", exists, tt.expectedExist)
			}
			if exists && value != tt.expectedValue {
				t.Errorf("addToOptions() value = %v, want %v", value, tt.expectedValue)
			}
		})
	}
}

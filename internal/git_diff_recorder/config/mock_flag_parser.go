package config

// MockFlagParser はテスト用のモック実装
type MockFlagParser struct {
	stringVars map[string]*string
	boolVars   map[string]*bool
	testValues map[string]interface{}
	args       []string
	parseError error
}

// NewMockFlagParser は新しいMockFlagParserを作成する
func NewMockFlagParser() *MockFlagParser {
	return &MockFlagParser{
		stringVars: make(map[string]*string),
		boolVars:   make(map[string]*bool),
		testValues: make(map[string]interface{}),
		args:       []string{},
	}
}

// StringVar は文字列フラグを定義する
func (m *MockFlagParser) StringVar(p *string, name string, value string, usage string) {
	*p = value // デフォルト値を設定
	m.stringVars[name] = p
}

// BoolVar はブールフラグを定義する
func (m *MockFlagParser) BoolVar(p *bool, name string, value bool, usage string) {
	*p = value // デフォルト値を設定
	m.boolVars[name] = p
}

// Parse はフラグを解析する
func (m *MockFlagParser) Parse() error {
	if m.parseError != nil {
		return m.parseError
	}

	// テスト用の値を設定
	for name, ptr := range m.stringVars {
		if value, exists := m.testValues[name]; exists {
			if strValue, ok := value.(string); ok {
				*ptr = strValue
			}
		}
	}

	for name, ptr := range m.boolVars {
		if value, exists := m.testValues[name]; exists {
			if boolValue, ok := value.(bool); ok {
				*ptr = boolValue
			}
		}
	}

	return nil
}

// Args は解析後の残りの引数を返す
func (m *MockFlagParser) Args() []string {
	return m.args
}

// SetStringValue はテスト用の文字列値を設定する
func (m *MockFlagParser) SetStringValue(name, value string) {
	m.testValues[name] = value
}

// SetBoolValue はテスト用のブール値を設定する
func (m *MockFlagParser) SetBoolValue(name string, value bool) {
	m.testValues[name] = value
}

// SetParseError はParseメソッドで返すエラーを設定する
func (m *MockFlagParser) SetParseError(err error) {
	m.parseError = err
}

// SetArgs は残りの引数を設定する
func (m *MockFlagParser) SetArgs(args []string) {
	m.args = args
}

// MockOSArgs はテスト用のOSArgs実装
type MockOSArgs struct {
	args []string
}

// NewMockOSArgs は新しいMockOSArgsを作成する
func NewMockOSArgs(args []string) *MockOSArgs {
	return &MockOSArgs{
		args: args,
	}
}

// Args はコマンドライン引数を返す
func (m *MockOSArgs) Args() []string {
	return m.args
}

// SetArgs はコマンドライン引数を設定する
func (m *MockOSArgs) SetArgs(args []string) {
	m.args = args
}

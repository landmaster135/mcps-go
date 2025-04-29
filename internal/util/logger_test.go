package util

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestDefaultTimeProvider_Now はDefaultTimeProviderのNowメソッドをテストします
func TestDefaultTimeProvider_Now(t *testing.T) {
	provider := &DefaultTimeProvider{}

	// 現在時刻を取得
	now := provider.Now()

	// 現在時刻が現在から前後1秒以内であることを確認
	diff := time.Since(now)
	if diff < 0 {
		diff = -diff
	}

	if diff > time.Second {
		t.Errorf("Now() returned time too far from current time: %v", diff)
	}
}

// TestDefaultTimeProvider_Format はDefaultTimeProviderのFormatメソッドをテストします
func TestDefaultTimeProvider_Format(t *testing.T) {
	provider := &DefaultTimeProvider{}

	// テスト用の時間を作成
	testTime := time.Date(2025, 4, 1, 12, 34, 56, 0, time.UTC)

	// フォーマットをテスト
	tests := []struct {
		layout string
		want   string
	}{
		{"2006-01-02", "2025-04-01"},
		{"15:04:05", "12:34:56"},
		{"2006/01/02 15:04:05", "2025/04/01 12:34:56"},
		{"20060102", "20250401"},
	}

	for _, tt := range tests {
		t.Run(tt.layout, func(t *testing.T) {
			got := provider.Format(testTime, tt.layout)
			if got != tt.want {
				t.Errorf("Format() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDefaultRuntimeProvider_Caller はDefaultRuntimeProviderのCallerメソッドをテストします
func TestDefaultRuntimeProvider_Caller(t *testing.T) {
	provider := &DefaultRuntimeProvider{}

	// Callerを呼び出す
	pc, file, line, ok := provider.Caller(0)

	// 結果を検証
	if !ok {
		t.Error("Caller() ok = false, want true")
	}
	if pc == 0 {
		t.Error("Caller() pc = 0, want non-zero")
	}
	if !strings.Contains(file, "logger") {
		t.Errorf("Caller() file = %v, want to contain 'logger'", file)
	}
	if line <= 0 {
		t.Errorf("Caller() line = %v, want > 0", line)
	}
}

// TestDefaultRuntimeProvider_FuncForPC はDefaultRuntimeProviderのFuncForPCメソッドをテストします
func TestDefaultRuntimeProvider_FuncForPC(t *testing.T) {
	provider := &DefaultRuntimeProvider{}

	// 現在の関数のPCを取得
	pc, _, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Failed to get current PC")
	}

	// FuncForPCを呼び出す
	fn := provider.FuncForPC(pc)

	// 結果を検証
	if fn == nil {
		t.Error("FuncForPC() = nil, want non-nil")
	}

	// 関数名に現在のテスト関数名が含まれていることを確認
	funcName := fn.Name()
	if !strings.Contains(funcName, "TestDefaultRuntimeProvider_FuncForPC") {
		t.Errorf("FuncForPC().Name() = %v, want to contain 'TestDefaultRuntimeProvider_FuncForPC'", funcName)
	}
}

// TestDefaultFileProvider_OpenFile はDefaultFileProviderのOpenFileメソッドをテストします
func TestDefaultFileProvider_OpenFile(t *testing.T) {
	provider := &DefaultFileProvider{}

	// 一時ファイルのパスを作成
	tempFile := fmt.Sprintf("test_file_%d.tmp", time.Now().UnixNano())
	defer os.Remove(tempFile) // テスト後にファイルを削除

	// OpenFileを呼び出す
	file, err := provider.OpenFile(tempFile, os.O_CREATE|os.O_WRONLY, 0666)

	// 結果を検証
	if err != nil {
		t.Errorf("OpenFile() error = %v, want nil", err)
	}
	if file == nil {
		t.Error("OpenFile() file = nil, want non-nil")
	}

	// ファイルを閉じる
	if file != nil {
		file.Close()
	}

	// ファイルが作成されたことを確認
	_, err = os.Stat(tempFile)
	if err != nil {
		t.Errorf("OpenFile() failed to create file: %v", err)
	}
}

// TestDefaultFileProvider_Stdout はDefaultFileProviderのStdoutメソッドをテストします
func TestDefaultFileProvider_Stdout(t *testing.T) {
	provider := &DefaultFileProvider{}

	// Stdoutを呼び出す
	stdout := provider.Stdout()

	// 結果を検証
	if stdout == nil {
		t.Error("Stdout() = nil, want non-nil")
	}

	// os.Stdoutと同じであることを確認（型比較）
	if fmt.Sprintf("%T", stdout) != fmt.Sprintf("%T", os.Stdout) {
		t.Errorf("Stdout() type = %T, want %T", stdout, os.Stdout)
	}
}

// TestDefaultFileProvider_Buffer はDefaultFileProviderのBufferメソッドをテストします
func TestDefaultFileProvider_Buffer(t *testing.T) {
	provider := &DefaultFileProvider{}

	// Bufferを呼び出す
	buffer := provider.Buffer()

	// 結果を検証
	if buffer == nil {
		t.Error("Buffer() = nil, want non-nil")
	}

	// bytes.Bufferであることを確認
	_, ok := buffer.(*bytes.Buffer)
	if !ok {
		t.Errorf("Buffer() type = %T, want *bytes.Buffer", buffer)
	}

	// バッファに書き込みができることを確認
	testData := []byte("test data")
	n, err := buffer.Write(testData)
	if err != nil {
		t.Errorf("Buffer().Write() error = %v, want nil", err)
	}
	if n != len(testData) {
		t.Errorf("Buffer().Write() n = %v, want %v", n, len(testData))
	}
}

// TestDefaultLoggerProvider_Default はDefaultLoggerProviderのDefaultメソッドをテストします
func TestDefaultLoggerProvider_Default(t *testing.T) {
	// ケース1: loggerがnilの場合
	t.Run("nil logger", func(t *testing.T) {
		provider := &DefaultLoggerProvider{logger: nil}

		// Defaultを呼び出す
		logger := provider.Default()

		// 結果を検証
		if logger == nil {
			t.Error("Default() = nil, want non-nil")
		}

		// 2回目の呼び出しで同じロガーが返されることを確認
		logger2 := provider.Default()
		if logger != logger2 {
			t.Error("Default() returned different loggers on subsequent calls")
		}
	})

	// ケース2: loggerが既に設定されている場合
	t.Run("non-nil logger", func(t *testing.T) {
		customLogger := log.New(&bytes.Buffer{}, "TEST: ", log.Ltime)
		provider := &DefaultLoggerProvider{logger: customLogger}

		// Defaultを呼び出す
		logger := provider.Default()

		// 結果を検証
		if logger != customLogger {
			t.Error("Default() returned different logger than the one set")
		}
	})
}

// TestDefaultLoggerProvider_Methods はDefaultLoggerProviderの各メソッドをテストします
func TestDefaultLoggerProvider_Methods(t *testing.T) {
	// テスト用のバッファを作成
	var buf bytes.Buffer

	// カスタムロガーを作成
	customLogger := log.New(&buf, "", 0)

	// プロバイダーを作成
	provider := &DefaultLoggerProvider{logger: customLogger}

	// SetOutputをテスト
	t.Run("SetOutput", func(t *testing.T) {
		var newBuf bytes.Buffer
		provider.SetOutput(&newBuf)

		// 出力先が変更されたことを確認
		provider.Printf("test message")
		if newBuf.Len() == 0 {
			t.Error("SetOutput() did not change the output destination")
		}
	})

	// SetPrefixをテスト
	t.Run("SetPrefix", func(t *testing.T) {
		var prefixBuf bytes.Buffer
		provider.logger = log.New(&prefixBuf, "", 0)

		testPrefix := "TEST_PREFIX: "
		provider.SetPrefix(testPrefix)

		// プレフィックスが設定されたことを確認
		provider.Printf("test message")
		if !strings.Contains(prefixBuf.String(), testPrefix) {
			t.Errorf("SetPrefix() did not set prefix correctly, got %q", prefixBuf.String())
		}
	})

	// SetFlagsをテスト
	t.Run("SetFlags", func(t *testing.T) {
		var flagsBuf bytes.Buffer
		provider.logger = log.New(&flagsBuf, "", 0)

		provider.SetFlags(log.Ldate)

		// フラグが設定されたことを確認（日付が含まれるようになる）
		provider.Printf("test message")

		currentYear := time.Now().Format("2006")
		if !strings.Contains(flagsBuf.String(), currentYear) {
			t.Errorf("SetFlags(log.Ldate) did not set flags correctly, got %q", flagsBuf.String())
		}
	})

	// PrintfとPanicfをテスト
	t.Run("Printf and Panicf", func(t *testing.T) {
		// Printfのテスト
		var printfBuf bytes.Buffer
		provider.logger = log.New(&printfBuf, "", 0)

		testMessage := "test printf message"
		provider.Printf(testMessage)

		if !strings.Contains(printfBuf.String(), testMessage) {
			t.Errorf("Printf() did not output correct message, got %q", printfBuf.String())
		}

		// Panicfのテスト（パニックをキャッチする）
		var panicfBuf bytes.Buffer
		provider.logger = log.New(&panicfBuf, "", 0)

		testPanicMessage := "test panicf message"

		defer func() {
			r := recover()
			if r == nil {
				t.Error("Panicf() did not panic")
			}

			// パニックメッセージを確認
			panicOutput := panicfBuf.String()
			if !strings.Contains(panicOutput, testPanicMessage) {
				t.Errorf("Panicf() did not output correct message before panicking, got %q", panicOutput)
			}
		}()

		provider.Panicf(testPanicMessage)
	})
}

// TestSetLogLevel_Default はsetLogLevel関数のデフォルトケースをテストします
func TestSetLogLevel_Default(t *testing.T) {
	// 未定義のログレベルを使用
	level := 999

	// setLogLevelを呼び出す
	result := setLogLevel(level)

	// 結果を検証
	expected := "INFO"
	if result != expected {
		t.Errorf("setLogLevel(%d) = %v, want %v", level, result, expected)
	}
}

// TestSetupOutput_Default はsetupOutput関数のデフォルトケースをテストします
func TestSetupOutput_Default(t *testing.T) {
	// モックのロガープロバイダーを作成
	mockLoggerProvider := &MockLoggerProvider{}

	// 共有バッファを作成
	sharedBuffer := &bytes.Buffer{}

	// モックのファイルプロバイダーを作成（同じバッファを返すように設定）
	mockFileProvider := &MockFileProvider{
		StdoutFunc: func() io.Writer {
			return sharedBuffer
		},
	}

	// 未定義の出力先を持つロガーを作成
	logger := &BuiltinLogger{
		logger:         mockLoggerProvider,
		output:         999, // 未定義の出力先
		timeProvider:   &MockTimeProvider{},
		runtimeProvider: &MockRuntimeProvider{},
		fileProvider:   mockFileProvider,
	}

	// setupOutputを呼び出す
	writer, err := logger.setupOutput()

	// 結果を検証
	if err != nil {
		t.Errorf("setupOutput() error = %v, want nil", err)
	}
	if writer != nil {
		t.Errorf("setupOutput() writer = %v, want nil", writer)
	}

	// デフォルトでは標準出力が設定されることを確認
	if mockLoggerProvider.OutputWriter != sharedBuffer {
		t.Error("setupOutput() did not set stdout as default output")
	}
}

// TestLogWithLevel_WithCallerInfo はlogWithLevelメソッドの条件分岐をテストします
func TestLogWithLevel_WithCallerInfo(t *testing.T) {
	// モックのロガープロバイダーを作成
	mockLoggerProvider := &MockLoggerProvider{}

	// モックのランタイムプロバイダーを作成（有効な呼び出し情報を返す）
	mockRuntimeProvider := &MockRuntimeProvider{
		CallerFunc: func(skip int) (pc uintptr, file string, line int, ok bool) {
			return 1, "test_file.go", 42, true
		},
		FuncForPCFunc: func(pc uintptr) *runtime.Func {
			// 実際のruntime.Funcを返す代わりにモックを使用
			callerPC, _, _, _ := runtime.Caller(0)
			return runtime.FuncForPC(callerPC)
		},
	}

	// ロガーを作成
	logger := &BuiltinLogger{
		logger:         mockLoggerProvider,
		output:         STDOUT,
		step:           1,
		timeProvider:   &MockTimeProvider{},
		runtimeProvider: mockRuntimeProvider,
		fileProvider:   &MockFileProvider{},
	}

	// logWithLevelを呼び出す
	testMessage := "テストメッセージ"
	logger.logWithLevel("TEST", testMessage)

	// 結果を検証
	if !mockLoggerProvider.PrintfCalled {
		t.Error("logWithLevel() did not call Printf on logger provider")
	}

	// 呼び出し情報が含まれていることを確認
	if !strings.Contains(mockLoggerProvider.LastFormat, "@") {
		t.Errorf("logWithLevel() format = %v, should contain caller info with '@'", mockLoggerProvider.LastFormat)
	}
	if !strings.Contains(mockLoggerProvider.LastFormat, "test_file.go") {
		t.Errorf("logWithLevel() format = %v, should contain file name", mockLoggerProvider.LastFormat)
	}
	if !strings.Contains(mockLoggerProvider.LastFormat, "42") {
		t.Errorf("logWithLevel() format = %v, should contain line number", mockLoggerProvider.LastFormat)
	}
}

// TestFatal_WithFileOutput はFatalメソッドのファイル出力条件分岐をテストします
func TestFatal_WithFileOutput(t *testing.T) {
	// モックのファイルを作成
	mockWriteCloser := &MockWriteCloser{
		CloseFunc: func() error {
			return nil
		},
	}

	// クローズが呼ばれたかを追跡
	closeCalled := false
	mockWriteCloser.CloseFunc = func() error {
		closeCalled = true
		return nil
	}

	// モックのファイルプロバイダーを作成
	mockFileProvider := &MockFileProvider{
		OpenFileFunc: func(name string, flag int, perm os.FileMode) (io.WriteCloser, error) {
			return mockWriteCloser, nil
		},
	}

	// モックのロガープロバイダーを作成（パニックを防止）
	mockLoggerProvider := &MockLoggerProvider{
		PanicfFunc: func(format string, v ...interface{}) {
			// パニックを発生させない
		},
	}

	// ロガーを作成
	logger := &BuiltinLogger{
		logger:         mockLoggerProvider,
		output:         FILE,
		step:           1,
		timeProvider:   &MockTimeProvider{},
		runtimeProvider: &MockRuntimeProvider{},
		fileProvider:   mockFileProvider,
	}

	// Fatalメソッドを呼び出す（パニックをキャッチ）
	defer func() {
		recover() // パニックを回復

		// ファイルがクローズされたことを確認
		if !closeCalled {
			t.Error("Fatal() did not close the file")
		}
	}()

	logger.Fatal("テストメッセージ")
}

// TestFatal_WithFileError はFatalメソッドのファイルエラー条件分岐をテストします
func TestFatal_WithFileError(t *testing.T) {
	// モックのファイルプロバイダーを作成（エラーを返す）
	mockFileProvider := &MockFileProvider{
		OpenFileFunc: func(name string, flag int, perm os.FileMode) (io.WriteCloser, error) {
			return nil, fmt.Errorf("模擬的なファイルエラー")
		},
	}

	// モックのロガープロバイダーを作成
	mockLoggerProvider := &MockLoggerProvider{}

	// ロガーを作成
	logger := &BuiltinLogger{
		logger:         mockLoggerProvider,
		output:         FILE,
		step:           1,
		timeProvider:   &MockTimeProvider{},
		runtimeProvider: &MockRuntimeProvider{},
		fileProvider:   mockFileProvider,
	}

	// Fatalメソッドを呼び出す（パニックをキャッチ）
	defer func() {
		err := recover()
		if err == nil {
			t.Error("Fatal() with file error did not panic")
		}
		// エラーメッセージを確認
		errMsg, ok := err.(string)
		if !ok {
			t.Errorf("Fatal() with file error panic value is not a string: %v", err)
		}
		if !strings.Contains(errMsg, "模擬的なファイルエラー") {
			t.Errorf("Fatal() with file error panic message = %v, want to contain '模擬的なファイルエラー'", errMsg)
		}
	}()

	logger.Fatal("テストメッセージ")
}

// TestFatal_WithCallerInfo はFatalメソッドの呼び出し情報条件分岐をテストします
func TestFatal_WithCallerInfo(t *testing.T) {
	// モックのロガープロバイダーを作成
	mockLoggerProvider := &MockLoggerProvider{
		PanicfFunc: func(format string, v ...interface{}) {
			// パニックを発生させない
		},
	}

	// モックのランタイムプロバイダーを作成（有効な呼び出し情報を返す）
	mockRuntimeProvider := &MockRuntimeProvider{
		CallerFunc: func(skip int) (pc uintptr, file string, line int, ok bool) {
			return 1, "test_file.go", 42, true
		},
		FuncForPCFunc: func(pc uintptr) *runtime.Func {
			// 実際のruntime.Funcを返す代わりにモックを使用
			callerPC, _, _, _ := runtime.Caller(0)
			return runtime.FuncForPC(callerPC)
		},
	}

	// ロガーを作成
	logger := &BuiltinLogger{
		logger:         mockLoggerProvider,
		output:         STDOUT,
		step:           1,
		timeProvider:   &MockTimeProvider{},
		runtimeProvider: mockRuntimeProvider,
		fileProvider:   &MockFileProvider{},
	}

	// Fatalメソッドを呼び出す
	testMessage := "テストメッセージ"
	logger.Fatal(testMessage)

	// 結果を検証
	if !mockLoggerProvider.PanicfCalled {
		t.Error("Fatal() with caller info did not call Panicf on logger provider")
	}

	// 呼び出し情報が含まれていることを確認
	if !strings.Contains(mockLoggerProvider.LastFormat, "@") {
		t.Errorf("Fatal() format = %v, should contain caller info with '@'", mockLoggerProvider.LastFormat)
	}
	if !strings.Contains(mockLoggerProvider.LastFormat, "test_file.go") {
		t.Errorf("Fatal() format = %v, should contain file name", mockLoggerProvider.LastFormat)
	}
	if !strings.Contains(mockLoggerProvider.LastFormat, "42") {
		t.Errorf("Fatal() format = %v, should contain line number", mockLoggerProvider.LastFormat)
	}
}

// MockTimeProvider は時間プロバイダーのモック実装です
type MockTimeProvider struct {
	NowFunc    func() time.Time
	FormatFunc func(t time.Time, layout string) string
}

// Now は現在時刻を返します
func (m *MockTimeProvider) Now() time.Time {
	if m.NowFunc != nil {
		return m.NowFunc()
	}
	return time.Date(2025, 4, 1, 12, 0, 0, 0, time.UTC)
}

// Format は指定されたレイアウトで時間をフォーマットします
func (m *MockTimeProvider) Format(t time.Time, layout string) string {
	if m.FormatFunc != nil {
		return m.FormatFunc(t, layout)
	}
	return "20250401"
}

// MockRuntimeProvider はランタイムプロバイダーのモック実装です
type MockRuntimeProvider struct {
	CallerFunc    func(skip int) (pc uintptr, file string, line int, ok bool)
	FuncForPCFunc func(pc uintptr) *runtime.Func
	MockFuncName  string
}

// Caller はコールスタック情報を返します
func (m *MockRuntimeProvider) Caller(skip int) (pc uintptr, file string, line int, ok bool) {
	if m.CallerFunc != nil {
		return m.CallerFunc(skip)
	}
	return 0, "test_file.go", 42, true
}

// FuncForPC は関数情報を返します
func (m *MockRuntimeProvider) FuncForPC(pc uintptr) *runtime.Func {
	if m.FuncForPCFunc != nil {
		return m.FuncForPCFunc(pc)
	}
	// 実際のruntime.Funcオブジェクトを返す
	return runtime.FuncForPC(0)
}

// MockWriteCloser はio.WriteCloserのモック実装です
type MockWriteCloser struct {
	WriteFunc func(p []byte) (n int, err error)
	CloseFunc func() error
	Buffer    bytes.Buffer
}

// Write はデータを書き込みます
func (m *MockWriteCloser) Write(p []byte) (n int, err error) {
	if m.WriteFunc != nil {
		return m.WriteFunc(p)
	}
	return m.Buffer.Write(p)
}

// Close はリソースを閉じます
func (m *MockWriteCloser) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// MockFileProvider はファイルプロバイダーのモック実装です
type MockFileProvider struct {
	OpenFileFunc func(name string, flag int, perm os.FileMode) (io.WriteCloser, error)
	StdoutFunc   func() io.Writer
	BufferFunc   func() io.Writer
}

// OpenFile はファイルを開きます
func (m *MockFileProvider) OpenFile(name string, flag int, perm os.FileMode) (io.WriteCloser, error) {
	if m.OpenFileFunc != nil {
		return m.OpenFileFunc(name, flag, perm)
	}
	return &MockWriteCloser{}, nil
}

// Stdout は標準出力を返します
func (m *MockFileProvider) Stdout() io.Writer {
	if m.StdoutFunc != nil {
		return m.StdoutFunc()
	}
	return &bytes.Buffer{}
}

// Buffer は新しいバッファを返します
func (m *MockFileProvider) Buffer() io.Writer {
	if m.BufferFunc != nil {
		return m.BufferFunc()
	}
	return &bytes.Buffer{}
}

// MockLoggerProvider はロガープロバイダーのモック実装です
type MockLoggerProvider struct {
	DefaultFunc   func() *log.Logger
	SetOutputFunc func(w io.Writer)
	SetPrefixFunc func(prefix string)
	SetFlagsFunc  func(flag int)
	PrintfFunc    func(format string, v ...interface{})
	PanicfFunc    func(format string, v ...interface{})

	// 呼び出し記録
	OutputWriter  io.Writer
	Prefix        string
	Flags         int
	LastMessage   string
	LastFormat    string
	LastArgs      []interface{}
	PrintfCalled  bool
	PanicfCalled  bool
}

// Default はデフォルトのロガーを返します
func (m *MockLoggerProvider) Default() *log.Logger {
	if m.DefaultFunc != nil {
		return m.DefaultFunc()
	}
	return log.New(&bytes.Buffer{}, "", 0)
}

// SetOutput はロガーの出力先を設定します
func (m *MockLoggerProvider) SetOutput(w io.Writer) {
	m.OutputWriter = w
	if m.SetOutputFunc != nil {
		m.SetOutputFunc(w)
	}
}

// SetPrefix はロガーのプレフィックスを設定します
func (m *MockLoggerProvider) SetPrefix(prefix string) {
	m.Prefix = prefix
	if m.SetPrefixFunc != nil {
		m.SetPrefixFunc(prefix)
	}
}

// SetFlags はロガーのフラグを設定します
func (m *MockLoggerProvider) SetFlags(flag int) {
	m.Flags = flag
	if m.SetFlagsFunc != nil {
		m.SetFlagsFunc(flag)
	}
}

// Printf はログメッセージを出力します
func (m *MockLoggerProvider) Printf(format string, v ...interface{}) {
	m.PrintfCalled = true
	m.LastFormat = format
	m.LastArgs = v
	m.LastMessage = fmt.Sprintf(format, v...)
	if m.PrintfFunc != nil {
		m.PrintfFunc(format, v...)
	}
}

// Panicf はログメッセージを出力し、パニックを発生させます
func (m *MockLoggerProvider) Panicf(format string, v ...interface{}) {
	m.PanicfCalled = true
	m.LastFormat = format
	m.LastArgs = v
	m.LastMessage = fmt.Sprintf(format, v...)
	if m.PanicfFunc != nil {
		m.PanicfFunc(format, v...)
	} else {
		// デフォルトではパニックを発生させない（テストのため）
	}
}

// TestNewBuiltinLogger はNewBuiltinLogger関数をテストします
func TestNewBuiltinLogger(t *testing.T) {
	// テストケース
	tests := []struct {
		name       string
		outputDest string
		wantOutput int
	}{
		{
			name:       "STDOUT出力",
			outputDest: "STDOUT",
			wantOutput: STDOUT,
		},
		{
			name:       "STDERR出力",
			outputDest: "STDERR",
			wantOutput: STDERR,
		},
		{
			name:       "FILE出力",
			outputDest: "FILE",
			wantOutput: FILE,
		},
		{
			name:       "BUFFER出力",
			outputDest: "BUFFER",
			wantOutput: BUFFER,
		},
		{
			name:       "不明な出力先",
			outputDest: "unknown",
			wantOutput: STDOUT, // デフォルトはSTDOUT
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewBuiltinLogger(tt.outputDest)
			if logger.output != tt.wantOutput {
				t.Errorf("NewBuiltinLogger() output = %v, want %v", logger.output, tt.wantOutput)
			}
			if logger.step != 1 {
				t.Errorf("NewBuiltinLogger() step = %v, want %v", logger.step, 1)
			}
		})
	}
}

// TestNewBuiltinLoggerWithProviders はNewBuiltinLoggerWithProviders関数をテストします
func TestNewBuiltinLoggerWithProviders(t *testing.T) {
	// モックプロバイダーを作成
	mockLoggerProvider := &MockLoggerProvider{}
	mockTimeProvider := &MockTimeProvider{}
	mockRuntimeProvider := &MockRuntimeProvider{}
	mockFileProvider := &MockFileProvider{}

	// テストケース
	tests := []struct {
		name       string
		outputDest string
		wantOutput int
	}{
		{
			name:       "STDOUT出力",
			outputDest: "STDOUT",
			wantOutput: STDOUT,
		},
		{
			name:       "FILE出力",
			outputDest: "FILE",
			wantOutput: FILE,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewBuiltinLoggerWithProviders(
				tt.outputDest,
				mockLoggerProvider,
				mockTimeProvider,
				mockRuntimeProvider,
				mockFileProvider,
			)

			// 型アサーションでBuiltinLoggerの内部にアクセス
			builtinLogger, ok := logger.(*BuiltinLogger)
			if !ok {
				t.Fatalf("NewBuiltinLoggerWithProviders() did not return *BuiltinLogger")
			}

			if builtinLogger.output != tt.wantOutput {
				t.Errorf("NewBuiltinLoggerWithProviders() output = %v, want %v", builtinLogger.output, tt.wantOutput)
			}
			if builtinLogger.step != 1 {
				t.Errorf("NewBuiltinLoggerWithProviders() step = %v, want %v", builtinLogger.step, 1)
			}

			// プロバイダーが正しく設定されているか確認
			if builtinLogger.logger != mockLoggerProvider {
				t.Errorf("NewBuiltinLoggerWithProviders() logger provider not set correctly")
			}
			if builtinLogger.timeProvider != mockTimeProvider {
				t.Errorf("NewBuiltinLoggerWithProviders() time provider not set correctly")
			}
			if builtinLogger.runtimeProvider != mockRuntimeProvider {
				t.Errorf("NewBuiltinLoggerWithProviders() runtime provider not set correctly")
			}
			if builtinLogger.fileProvider != mockFileProvider {
				t.Errorf("NewBuiltinLoggerWithProviders() file provider not set correctly")
			}
		})
	}
}

// TestBuiltinLogger_NextStep はNextStepメソッドをテストします
func TestBuiltinLogger_NextStep(t *testing.T) {
	logger := NewBuiltinLogger("STDOUT")
	initialStep := logger.step

	// NextStepを呼び出す
	logger.NextStep()

	// 結果を検証
	if logger.step != initialStep+1 {
		t.Errorf("NextStep() step = %v, want %v", logger.step, initialStep+1)
	}

	// 複数回呼び出す
	logger.NextStep()
	logger.NextStep()

	// 結果を検証
	if logger.step != initialStep+3 {
		t.Errorf("NextStep() after multiple calls, step = %v, want %v", logger.step, initialStep+3)
	}
}

// TestBuiltinLogger_setCommonLogFileName はsetCommonLogFileNameメソッドをテストします
func TestBuiltinLogger_setCommonLogFileName(t *testing.T) {
	// モックの時間プロバイダーを作成
	mockTimeProvider := &MockTimeProvider{
		FormatFunc: func(t time.Time, layout string) string {
			return "20250401"
		},
	}

	// モックプロバイダーを使用してロガーを作成
	logger := NewBuiltinLoggerWithProviders(
		"FILE",
		&MockLoggerProvider{},
		mockTimeProvider,
		&MockRuntimeProvider{},
		&MockFileProvider{},
	).(*BuiltinLogger)

	// setCommonLogFileNameを呼び出す
	filename := logger.setCommonLogFileName()

	// 結果を検証
	expected := "common_20250401.log"
	if filename != expected {
		t.Errorf("setCommonLogFileName() = %v, want %v", filename, expected)
	}
}

// TestBuiltinLogger_Debug はDebugメソッドをテストします
func TestBuiltinLogger_Debug(t *testing.T) {
	// モックの出力バッファ
	var outputBuffer bytes.Buffer

	// モックのロガープロバイダーを作成
	mockLoggerProvider := &MockLoggerProvider{}

	// モックのファイルプロバイダーを作成
	mockFileProvider := &MockFileProvider{
		StdoutFunc: func() io.Writer {
			return &outputBuffer
		},
	}

	// モックのランタイムプロバイダーを作成
	mockRuntimeProvider := &MockRuntimeProvider{
		CallerFunc: func(skip int) (pc uintptr, file string, line int, ok bool) {
			return 0, "test_file.go", 42, true
		},
		MockFuncName: "TestFunction",
	}

	// モックプロバイダーを使用してロガーを作成
	logger := NewBuiltinLoggerWithProviders(
		"STDOUT",
		mockLoggerProvider,
		&MockTimeProvider{},
		mockRuntimeProvider,
		mockFileProvider,
	).(*BuiltinLogger)

	// Debugメソッドを呼び出す
	testMessage := "テストメッセージ"
	logger.Debug(testMessage)

	// 結果を検証
	if !mockLoggerProvider.PrintfCalled {
		t.Error("Debug() did not call Printf on logger provider")
	}
	if !strings.Contains(mockLoggerProvider.LastFormat, testMessage) {
		t.Errorf("Debug() format = %v, want to contain %v", mockLoggerProvider.LastFormat, testMessage)
	}
	if mockLoggerProvider.Prefix != "[DEBUG] [Step 1/5] " {
		t.Errorf("Debug() prefix = %v, want %v", mockLoggerProvider.Prefix, "[DEBUG] [Step 1/5] ")
	}
}

// TestBuiltinLogger_Info はInfoメソッドをテストします
func TestBuiltinLogger_Info(t *testing.T) {
	// モックのロガープロバイダーを作成
	mockLoggerProvider := &MockLoggerProvider{}

	// モックプロバイダーを使用してロガーを作成
	logger := NewBuiltinLoggerWithProviders(
		"STDOUT",
		mockLoggerProvider,
		&MockTimeProvider{},
		&MockRuntimeProvider{},
		&MockFileProvider{},
	).(*BuiltinLogger)

	// Infoメソッドを呼び出す
	testMessage := "情報メッセージ"
	logger.Info(testMessage)

	// 結果を検証
	if !mockLoggerProvider.PrintfCalled {
		t.Error("Info() did not call Printf on logger provider")
	}
	if !strings.Contains(mockLoggerProvider.LastFormat, testMessage) {
		t.Errorf("Info() format = %v, want to contain %v", mockLoggerProvider.LastFormat, testMessage)
	}
	if mockLoggerProvider.Prefix != "[INFO] [Step 1/5] " {
		t.Errorf("Info() prefix = %v, want %v", mockLoggerProvider.Prefix, "[INFO] [Step 1/5] ")
	}
}

// TestBuiltinLogger_Warning はWarningメソッドをテストします
func TestBuiltinLogger_Warning(t *testing.T) {
	// モックのロガープロバイダーを作成
	mockLoggerProvider := &MockLoggerProvider{}

	// モックプロバイダーを使用してロガーを作成
	logger := NewBuiltinLoggerWithProviders(
		"STDOUT",
		mockLoggerProvider,
		&MockTimeProvider{},
		&MockRuntimeProvider{},
		&MockFileProvider{},
	).(*BuiltinLogger)

	// Warningメソッドを呼び出す
	testMessage := "警告メッセージ"
	logger.Warning(testMessage)

	// 結果を検証
	if !mockLoggerProvider.PrintfCalled {
		t.Error("Warning() did not call Printf on logger provider")
	}
	if !strings.Contains(mockLoggerProvider.LastFormat, testMessage) {
		t.Errorf("Warning() format = %v, want to contain %v", mockLoggerProvider.LastFormat, testMessage)
	}
	if mockLoggerProvider.Prefix != "[WARNING] [Step 1/5] " {
		t.Errorf("Warning() prefix = %v, want %v", mockLoggerProvider.Prefix, "[WARNING] [Step 1/5] ")
	}
}

// TestBuiltinLogger_Error はErrorメソッドをテストします
func TestBuiltinLogger_Error(t *testing.T) {
	// モックのロガープロバイダーを作成
	mockLoggerProvider := &MockLoggerProvider{}

	// モックプロバイダーを使用してロガーを作成
	logger := NewBuiltinLoggerWithProviders(
		"STDOUT",
		mockLoggerProvider,
		&MockTimeProvider{},
		&MockRuntimeProvider{},
		&MockFileProvider{},
	).(*BuiltinLogger)

	// Errorメソッドを呼び出す
	testMessage := "エラーメッセージ"
	logger.Error(testMessage)

	// 結果を検証
	if !mockLoggerProvider.PrintfCalled {
		t.Error("Error() did not call Printf on logger provider")
	}
	if !strings.Contains(mockLoggerProvider.LastFormat, testMessage) {
		t.Errorf("Error() format = %v, want to contain %v", mockLoggerProvider.LastFormat, testMessage)
	}
	if mockLoggerProvider.Prefix != "[ERROR] [Step 1/5] " {
		t.Errorf("Error() prefix = %v, want %v", mockLoggerProvider.Prefix, "[ERROR] [Step 1/5] ")
	}
}

// TestBuiltinLogger_Fatal はFatalメソッドをテストします
func TestBuiltinLogger_Fatal(t *testing.T) {
	// モックのロガープロバイダーを作成
	mockLoggerProvider := &MockLoggerProvider{}

	// モックプロバイダーを使用してロガーを作成
	logger := NewBuiltinLoggerWithProviders(
		"STDOUT",
		mockLoggerProvider,
		&MockTimeProvider{},
		&MockRuntimeProvider{},
		&MockFileProvider{},
	).(*BuiltinLogger)

	// Fatalメソッドを呼び出す
	testMessage := "致命的エラー"
	logger.Fatal(testMessage)

	// 結果を検証
	if !mockLoggerProvider.PanicfCalled {
		t.Error("Fatal() did not call Panicf on logger provider")
	}
	if !strings.Contains(mockLoggerProvider.LastFormat, testMessage) {
		t.Errorf("Fatal() format = %v, want to contain %v", mockLoggerProvider.LastFormat, testMessage)
	}
	if mockLoggerProvider.Prefix != "[ERROR] [Step 1/5] " {
		t.Errorf("Fatal() prefix = %v, want %v", mockLoggerProvider.Prefix, "[ERROR] [Step 1/5] ")
	}
}

// TestBuiltinLogger_WithFileOutput はファイル出力をテストします
func TestBuiltinLogger_WithFileOutput(t *testing.T) {
	// モックのファイルを作成
	mockWriteCloser := &MockWriteCloser{}

	// モックのファイルプロバイダーを作成
	mockFileProvider := &MockFileProvider{
		OpenFileFunc: func(name string, flag int, perm os.FileMode) (io.WriteCloser, error) {
			return mockWriteCloser, nil
		},
	}

	// モックのロガープロバイダーを作成
	mockLoggerProvider := &MockLoggerProvider{}

	// モックプロバイダーを使用してロガーを作成
	logger := NewBuiltinLoggerWithProviders(
		"FILE",
		mockLoggerProvider,
		&MockTimeProvider{},
		&MockRuntimeProvider{},
		mockFileProvider,
	).(*BuiltinLogger)

	// Infoメソッドを呼び出す
	testMessage := "ファイル出力テスト"
	logger.Info(testMessage)

	// 結果を検証
	if !mockLoggerProvider.PrintfCalled {
		t.Error("Info() with FILE output did not call Printf on logger provider")
	}
	if !strings.Contains(mockLoggerProvider.LastFormat, testMessage) {
		t.Errorf("Info() with FILE output format = %v, want to contain %v", mockLoggerProvider.LastFormat, testMessage)
	}
}

// TestBuiltinLogger_WithBufferOutput はバッファ出力をテストします
func TestBuiltinLogger_WithBufferOutput(t *testing.T) {
	// モックのバッファを作成
	mockBuffer := &bytes.Buffer{}

	// モックのファイルプロバイダーを作成
	mockFileProvider := &MockFileProvider{
		BufferFunc: func() io.Writer {
			return mockBuffer
		},
	}

	// モックのロガープロバイダーを作成
	mockLoggerProvider := &MockLoggerProvider{}

	// モックプロバイダーを使用してロガーを作成
	logger := NewBuiltinLoggerWithProviders(
		"BUFFER",
		mockLoggerProvider,
		&MockTimeProvider{},
		&MockRuntimeProvider{},
		mockFileProvider,
	).(*BuiltinLogger)

	// Infoメソッドを呼び出す
	testMessage := "バッファ出力テスト"
	logger.Info(testMessage)

	// 結果を検証
	if !mockLoggerProvider.PrintfCalled {
		t.Error("Info() with BUFFER output did not call Printf on logger provider")
	}
	if !strings.Contains(mockLoggerProvider.LastFormat, testMessage) {
		t.Errorf("Info() with BUFFER output format = %v, want to contain %v", mockLoggerProvider.LastFormat, testMessage)
	}
	if mockLoggerProvider.OutputWriter != mockBuffer {
		t.Error("Info() with BUFFER output did not set buffer as output")
	}
}

// TestBuiltinLogger_WithRuntimeError はランタイムエラーをテストします
func TestBuiltinLogger_WithRuntimeError(t *testing.T) {
	// モックのランタイムプロバイダーを作成（エラーを返す）
	mockRuntimeProvider := &MockRuntimeProvider{
		CallerFunc: func(skip int) (pc uintptr, file string, line int, ok bool) {
			return 0, "", 0, false
		},
	}

	// モックのロガープロバイダーを作成
	mockLoggerProvider := &MockLoggerProvider{}

	// モックプロバイダーを使用してロガーを作成
	logger := NewBuiltinLoggerWithProviders(
		"STDOUT",
		mockLoggerProvider,
		&MockTimeProvider{},
		mockRuntimeProvider,
		&MockFileProvider{},
	).(*BuiltinLogger)

	// Infoメソッドを呼び出す
	testMessage := "ランタイムエラーテスト"
	logger.Info(testMessage)

	// 結果を検証
	if !mockLoggerProvider.PrintfCalled {
		t.Error("Info() with runtime error did not call Printf on logger provider")
	}
	if !strings.Contains(mockLoggerProvider.LastFormat, testMessage) {
		t.Errorf("Info() with runtime error format = %v, want to contain %v", mockLoggerProvider.LastFormat, testMessage)
	}
	// Caller情報が含まれていないことを確認
	if strings.Contains(mockLoggerProvider.LastFormat, "@") {
		t.Errorf("Info() with runtime error format = %v, should not contain caller info", mockLoggerProvider.LastFormat)
	}
}

// TestBuiltinLogger_WithFileError はファイルエラーをテストします
func TestBuiltinLogger_WithFileError(t *testing.T) {
	// モックのファイルプロバイダーを作成（エラーを返す）
	mockFileProvider := &MockFileProvider{
		OpenFileFunc: func(name string, flag int, perm os.FileMode) (io.WriteCloser, error) {
			return nil, fmt.Errorf("模擬的なファイルエラー")
		},
	}

	// モックのロガープロバイダーを作成
	mockLoggerProvider := &MockLoggerProvider{}

	// モックプロバイダーを使用してロガーを作成
	logger := NewBuiltinLoggerWithProviders(
		"FILE",
		mockLoggerProvider,
		&MockTimeProvider{},
		&MockRuntimeProvider{},
		mockFileProvider,
	).(*BuiltinLogger)

	// Infoメソッドを呼び出す（パニックが発生するため、recoverで捕捉）
	defer func() {
		err := recover()
		if err == nil {
			t.Error("Info() with file error did not panic")
		}
		// エラーメッセージを確認
		errMsg, ok := err.(string)
		if !ok {
			t.Errorf("Info() with file error panic value is not a string: %v", err)
		}
		if !strings.Contains(errMsg, "模擬的なファイルエラー") {
			t.Errorf("Info() with file error panic message = %v, want to contain '模擬的なファイルエラー'", errMsg)
		}
	}()

	testMessage := "ファイルエラーテスト"
	logger.Info(testMessage)
}

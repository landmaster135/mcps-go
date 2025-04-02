package util

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"time"
)

// BuiltinLoggerInterface インターフェースはビルトインロガーの機能を提供します
type BuiltinLoggerInterface interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warning(format string, args ...interface{})
	Error(format string, args ...interface{})
	Fatal(format string, args ...interface{})
	NextStep()
}

// TimeProvider インターフェースは時間関連の機能を提供します
type TimeProvider interface {
	Now() time.Time
	Format(t time.Time, layout string) string
}

// DefaultTimeProvider は標準的な時間プロバイダーの実装です
type DefaultTimeProvider struct{}

// Now は現在時刻を返します
func (p *DefaultTimeProvider) Now() time.Time {
	return time.Now()
}

// Format は指定されたレイアウトで時間をフォーマットします
func (p *DefaultTimeProvider) Format(t time.Time, layout string) string {
	return t.Format(layout)
}

// RuntimeProvider インターフェースはランタイム情報を提供します
type RuntimeProvider interface {
	Caller(skip int) (pc uintptr, file string, line int, ok bool)
	FuncForPC(pc uintptr) *runtime.Func
}

// DefaultRuntimeProvider は標準的なランタイムプロバイダーの実装です
type DefaultRuntimeProvider struct{}

// Caller はコールスタック情報を返します
func (p *DefaultRuntimeProvider) Caller(skip int) (pc uintptr, file string, line int, ok bool) {
	return runtime.Caller(skip)
}

// FuncForPC は関数情報を返します
func (p *DefaultRuntimeProvider) FuncForPC(pc uintptr) *runtime.Func {
	return runtime.FuncForPC(pc)
}

// FileProvider インターフェースはファイル操作機能を提供します
type FileProvider interface {
	OpenFile(name string, flag int, perm os.FileMode) (io.WriteCloser, error)
	Stdout() io.Writer
	Buffer() io.Writer
}

// DefaultFileProvider は標準的なファイルプロバイダーの実装です
type DefaultFileProvider struct{}

// OpenFile はファイルを開きます
func (p *DefaultFileProvider) OpenFile(name string, flag int, perm os.FileMode) (io.WriteCloser, error) {
	return os.OpenFile(name, flag, perm)
}

// Stdout は標準出力を返します
func (p *DefaultFileProvider) Stdout() io.Writer {
	return os.Stdout
}

// Buffer は新しいバッファを返します
func (p *DefaultFileProvider) Buffer() io.Writer {
	return &bytes.Buffer{}
}

// LoggerProvider インターフェースはロガーの設定機能を提供します
type LoggerProvider interface {
	Default() *log.Logger
	SetOutput(w io.Writer)
	SetPrefix(prefix string)
	SetFlags(flag int)
	Printf(format string, v ...interface{})
	Panicf(format string, v ...interface{})
}

// DefaultLoggerProvider は標準的なロガープロバイダーの実装です
type DefaultLoggerProvider struct {
	logger *log.Logger
}

// Default はデフォルトのロガーを返します
func (p *DefaultLoggerProvider) Default() *log.Logger {
	if p.logger == nil {
		p.logger = log.Default()
	}
	return p.logger
}

// SetOutput はロガーの出力先を設定します
func (p *DefaultLoggerProvider) SetOutput(w io.Writer) {
	p.Default().SetOutput(w)
}

// SetPrefix はロガーのプレフィックスを設定します
func (p *DefaultLoggerProvider) SetPrefix(prefix string) {
	p.Default().SetPrefix(prefix)
}

// SetFlags はロガーのフラグを設定します
func (p *DefaultLoggerProvider) SetFlags(flag int) {
	p.Default().SetFlags(flag)
}

// Printf はログメッセージを出力します
func (p *DefaultLoggerProvider) Printf(format string, v ...interface{}) {
	p.Default().Printf(format, v...)
}

// Panicf はログメッセージを出力し、パニックを発生させます
func (p *DefaultLoggerProvider) Panicf(format string, v ...interface{}) {
	p.Default().Panicf(format, v...)
}

// log output destinations
const (
	STDOUT = iota + 1
	STDERR
	FILE
	BUFFER
)

// log levels
const (
	ERROR = iota + 1
	WARNING
	INFO
	DEBUG
)

// TOTAL_STEP はログに表示される総ステップ数です
const TOTAL_STEP = 5

// sets log output destinations
func setOutput(dest string) int {
	switch dest {
	case "STDOUT", "stdout":
		return STDOUT
	case "STDERR", "stderr":
		return STDERR
	case "FILE", "file":
		return FILE
	case "BUFFER", "buffer":
		return BUFFER
	default:
		return STDOUT
	}
}

// sets log levels
func setLogLevel(level int) string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	default:
		return "INFO"
	}
}

// BuiltinLogger は標準的なロガーの実装です
type BuiltinLogger struct {
	logger         LoggerProvider
	output         int
	step           int
	timeProvider   TimeProvider
	runtimeProvider RuntimeProvider
	fileProvider   FileProvider
}

// NewBuiltinLogger は新しい標準ロガーを作成します
func NewBuiltinLogger(outputDest string) *BuiltinLogger {
	return &BuiltinLogger{
		logger:         &DefaultLoggerProvider{},
		output:         setOutput(outputDest),
		step:           1,
		timeProvider:   &DefaultTimeProvider{},
		runtimeProvider: &DefaultRuntimeProvider{},
		fileProvider:   &DefaultFileProvider{},
	}
}

// NewBuiltinLoggerWithProviders は指定されたプロバイダーを使用して新しい標準ロガーを作成します
func NewBuiltinLoggerWithProviders(
	outputDest string,
	loggerProvider LoggerProvider,
	timeProvider TimeProvider,
	runtimeProvider RuntimeProvider,
	fileProvider FileProvider,
) BuiltinLoggerInterface {
	return &BuiltinLogger{
		logger:         loggerProvider,
		output:         setOutput(outputDest),
		step:           1,
		timeProvider:   timeProvider,
		runtimeProvider: runtimeProvider,
		fileProvider:   fileProvider,
	}
}

// NextStep はステップを進めます
func (l *BuiltinLogger) NextStep() {
	l.step = l.step + 1
}

// setCommonLogFileName はログファイル名を生成します
func (l *BuiltinLogger) setCommonLogFileName() string {
	dateSuffix := l.timeProvider.Format(l.timeProvider.Now(), "20060102")
	return fmt.Sprintf("common_%s.log", dateSuffix)
}

// setupOutput はログの出力先を設定します
func (l *BuiltinLogger) setupOutput() (io.WriteCloser, error) {
	var writer io.WriteCloser
	var err error

	switch l.output {
	case FILE:
		writer, err = l.fileProvider.OpenFile(
			l.setCommonLogFileName(),
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0666,
		)
		if err != nil {
			return nil, err
		}
	case STDOUT, STDERR:
		// 標準出力は閉じる必要がないのでnilを返す
		writer = nil
		l.logger.SetOutput(l.fileProvider.Stdout())
	case BUFFER:
		// バッファは外部で管理するのでnilを返す
		writer = nil
		l.logger.SetOutput(l.fileProvider.Buffer())
	default:
		writer = nil
		l.logger.SetOutput(l.fileProvider.Stdout())
	}

	return writer, nil
}

// logWithLevel は指定されたレベルでメッセージをログに記録します
func (l *BuiltinLogger) logWithLevel(level string, format string, args ...interface{}) {
	prefix := fmt.Sprintf("[%s] [Step %d/%d] ", level, l.step, TOTAL_STEP)
	l.logger.SetPrefix(prefix)
	l.logger.SetFlags(log.Ldate | log.Ltime)

	writer, err := l.setupOutput()
	if err != nil {
		log.Panic(err.Error())
	}
	if writer != nil {
		defer writer.Close()
	}

	pc, file, line, ok := l.runtimeProvider.Caller(1)
	fn := l.runtimeProvider.FuncForPC(pc)
	if ok && fn != nil {
		caller := fmt.Sprintf("@%s:%d %s(): ", file, line, fn.Name())
		l.logger.Printf(caller+format, args...)
	} else {
		l.logger.Printf(format, args...)
	}
}

// Debug はデバッグメッセージをログに記録します
func (l *BuiltinLogger) Debug(format string, args ...interface{}) {
	logLevel := setLogLevel(DEBUG)
	l.logWithLevel(logLevel, format, args...)
}

// Info は情報メッセージをログに記録します
func (l *BuiltinLogger) Info(format string, args ...interface{}) {
	logLevel := setLogLevel(INFO)
	l.logWithLevel(logLevel, format, args...)
}

// Warning は警告メッセージをログに記録します
func (l *BuiltinLogger) Warning(format string, args ...interface{}) {
	logLevel := setLogLevel(WARNING)
	l.logWithLevel(logLevel, format, args...)
}

// Error はエラーメッセージをログに記録します
func (l *BuiltinLogger) Error(format string, args ...interface{}) {
	logLevel := setLogLevel(ERROR)
	l.logWithLevel(logLevel, format, args...)
}

// Fatal は致命的なメッセージをログに記録し、パニックを発生させます
func (l *BuiltinLogger) Fatal(format string, args ...interface{}) {
	logLevel := setLogLevel(ERROR)
	prefix := fmt.Sprintf("[%s] [Step %d/%d] ", logLevel, l.step, TOTAL_STEP)
	l.logger.SetPrefix(prefix)
	l.logger.SetFlags(log.Ldate | log.Ltime)

	writer, err := l.setupOutput()
	if err != nil {
		log.Panic(err.Error())
	}
	if writer != nil {
		defer writer.Close()
	}

	pc, file, line, ok := l.runtimeProvider.Caller(1)
	fn := l.runtimeProvider.FuncForPC(pc)
	if ok && fn != nil {
		caller := fmt.Sprintf("@%s:%d %s(): ", file, line, fn.Name())
		l.logger.Panicf(caller+format, args...)
	} else {
		l.logger.Panicf(format, args...)
	}
}

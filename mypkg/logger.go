package mypkg

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"
)

// log output destinations
const (
	STDOUT = iota + 1
	STDERR
	FILE
  BUFFER
)

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

// sets common log file name
func setCommonLogFileName() string {
	dateSuffix := time.Now().Format("20060102")
	filename := fmt.Sprintf("common_%s.log", dateSuffix)
	return filename
}

// log levels
const (
	ERROR = iota + 1
	WARNING
	INFO
	DEBUG
)

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

type BuiltinLogger struct {
	logger *log.Logger
	output int
	step   int
}

func NewBuiltinLogger(outputDest string) *BuiltinLogger {
	return &BuiltinLogger{
		logger: log.Default(),
		output: setOutput(outputDest),
		step:   1,
	}
}

const TOTAL_STEP = 5

func (l *BuiltinLogger) NextStep() {
	l.step = l.step + 1
}

// outputs debug log
func (l *BuiltinLogger) Debug(format string, args ...interface{}) {
	logLevel := setLogLevel(DEBUG)
	prefix := fmt.Sprintf("[%s] [Step %d/%d] ", logLevel, l.step, TOTAL_STEP)

	switch l.output {
	case FILE:
		logFile, err := os.OpenFile(setCommonLogFileName(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Panic(err.Error())
		}
		defer logFile.Close()
    l.logger.SetOutput(logFile)
	case STDOUT, STDERR:
		// logFile = os.Stdout
    l.logger.SetOutput(os.Stdout)
	case BUFFER:
    var buf bytes.Buffer
    l.logger.SetOutput(&buf)
	default:
    l.logger.SetOutput(os.Stdout)
	}

	l.logger.SetPrefix(prefix)
	l.logger.SetFlags(log.Ldate | log.Ltime)

	pc, file, line, ok := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)
	if ok {
		caller := fmt.Sprintf("@%s:%d %s(): ", file, line, fn.Name())
		l.logger.Printf(caller+format, args...)
	} else {
		l.logger.Printf(format, args...)
	}
}

// outputs information log
func (l *BuiltinLogger) Info(format string, args ...interface{}) {
	logLevel := setLogLevel(INFO)
	prefix := fmt.Sprintf("[%s] [Step %d/%d] ", logLevel, l.step, TOTAL_STEP)

	switch l.output {
	case FILE:
		logFile, err := os.OpenFile(setCommonLogFileName(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Panic(err.Error())
		}
		defer logFile.Close()
    l.logger.SetOutput(logFile)
	case STDOUT, STDERR:
		// logFile = os.Stdout
    l.logger.SetOutput(os.Stdout)
	case BUFFER:
    var buf bytes.Buffer
    l.logger.SetOutput(&buf)
	default:
    l.logger.SetOutput(os.Stdout)
	}

	l.logger.SetPrefix(prefix)
	l.logger.SetFlags(log.Ldate | log.Ltime)

	pc, file, line, ok := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)
	if ok {
		caller := fmt.Sprintf("@%s:%d %s(): ", file, line, fn.Name())
		l.logger.Printf(caller+format, args...)
	} else {
		l.logger.Printf(format, args...)
	}
}

// outputs warning log
func (l *BuiltinLogger) Warning(format string, args ...interface{}) {
	logLevel := setLogLevel(WARNING)
	prefix := fmt.Sprintf("[%s] [Step %d/%d] ", logLevel, l.step, TOTAL_STEP)

	switch l.output {
	case FILE:
		logFile, err := os.OpenFile(setCommonLogFileName(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Panic(err.Error())
		}
		defer logFile.Close()
    l.logger.SetOutput(logFile)
	case STDOUT, STDERR:
		// logFile = os.Stdout
    l.logger.SetOutput(os.Stdout)
	case BUFFER:
    var buf bytes.Buffer
    l.logger.SetOutput(&buf)
	default:
    l.logger.SetOutput(os.Stdout)
	}

	l.logger.SetPrefix(prefix)
	l.logger.SetFlags(log.Ldate | log.Ltime)

	pc, file, line, ok := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)
	if ok {
		caller := fmt.Sprintf("@%s:%d %s(): ", file, line, fn.Name())
		l.logger.Printf(caller+format, args...)
	} else {
		l.logger.Printf(format, args...)
	}
}

// outputs error log
func (l *BuiltinLogger) Error(format string, args ...interface{}) {
	logLevel := setLogLevel(ERROR)
	prefix := fmt.Sprintf("[%s] [Step %d/%d] ", logLevel, l.step, TOTAL_STEP)

	switch l.output {
	case FILE:
		logFile, err := os.OpenFile(setCommonLogFileName(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Panic(err.Error())
		}
		defer logFile.Close()
    l.logger.SetOutput(logFile)
	case STDOUT, STDERR:
		// logFile = os.Stdout
    l.logger.SetOutput(os.Stdout)
	case BUFFER:
    var buf bytes.Buffer
    l.logger.SetOutput(&buf)
	default:
    l.logger.SetOutput(os.Stdout)
	}

	l.logger.SetPrefix(prefix)
	l.logger.SetFlags(log.Ldate | log.Ltime)

	pc, file, line, ok := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)
	if ok {
		caller := fmt.Sprintf("@%s:%d %s(): ", file, line, fn.Name())
		l.logger.Printf(caller+format, args...)
	} else {
		l.logger.Printf(format, args...)
	}
}

// outputs error log with panic
func (l *BuiltinLogger) Fatal(format string, args ...interface{}) {
	logLevel := setLogLevel(ERROR)
	prefix := fmt.Sprintf("[%s] [Step %d/%d] ", logLevel, l.step, TOTAL_STEP)

	switch l.output {
	case FILE:
		logFile, err := os.OpenFile(setCommonLogFileName(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Panic(err.Error())
		}
		defer logFile.Close()
    l.logger.SetOutput(logFile)
	case STDOUT, STDERR:
		// logFile = os.Stdout
    l.logger.SetOutput(os.Stdout)
	case BUFFER:
    var buf bytes.Buffer
    l.logger.SetOutput(&buf)
	default:
    l.logger.SetOutput(os.Stdout)
	}

	l.logger.SetPrefix(prefix)
	l.logger.SetFlags(log.Ldate | log.Ltime)

	pc, file, line, ok := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)
	if ok {
		caller := fmt.Sprintf("@%s:%d %s(): ", file, line, fn.Name())
		l.logger.Panicf(caller+format, args...)
	} else {
		l.logger.Panicf(format, args...)
	}
}

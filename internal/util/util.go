package util

import (
	"fmt"
	"log"
	"log/slog"
	"runtime"
	"time"
)

type Logger struct {
}
type LoggerIf interface {
	OutNewLog(content ...any)
}

func NewLogger() LoggerIf {
	return &Logger{}
}

func OutLog(content ...any) {
	nl := NewLogger()
	nl.OutNewLog(content...)
}

func (l *Logger) OutNewLog(content ...any) {
	var m = make([]byte, 0, 50)
	var tmp string
	for _, v := range content {
		tmp = fmt.Sprintf("%v", v)
		m = append(m, tmp...)
		m = append(m, ' ')
	}
	c := string(m)
	slog.Info(c)

	// fmt.Println(content)
}

type TimeFmt struct {
}
type TimeFmtIf interface {
	Time2str(t time.Time) string
}

func NewTimeFmt() TimeFmtIf {
	return &TimeFmt{}
}

// Time2str converts time format to string.
func (f *TimeFmt) Time2str(t time.Time) string {
	// return t.Format(time.RFC3339)
	return t.Format("2006-01-02T15:04:05Z")
}

// LogMemoryUsage outputs the status of memory usage.
// You can use like: `defer LogMemoryUsage()`
func LogMemoryUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	log.Printf("[Memory Usage] Alloc: %v KB, TotalAlloc: %v KB, Sys: %v KB, NumGC: %v\n",
		m.Alloc/1024,
		m.TotalAlloc/1024,
		m.Sys/1024,
		m.NumGC,
	)
}

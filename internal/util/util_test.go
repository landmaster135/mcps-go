package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name string
		want LoggerIf
	}{
		{"Cover NewLogger codes.", NewLogger()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewLogger())
		})
	}
}

func TestOutLog(t *testing.T) {
	type args struct {
		content []any
	}
	tests := []struct {
		name string
		args args
	}{
		{"1 string", args{[]any{"aaa"}}},
		{"2 strings", args{[]any{"ccccc", "bbbb"}}},
		{"string and int", args{[]any{"zzzzzzz", 1234}}},
		{"rune, bool and nil", args{[]any{'b', true, nil}}},
		{"array", args{[]any{[]string{"test", "spec", "dummy"}}}},
		{"map", args{[]any{map[string]int{"xxx": 100, "yyyy": 200}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			OutLog(tt.args.content...)
		})
	}
}

func TestLogger_OutNewLog(t *testing.T) {
	type args struct {
		content []any
	}
	tests := []struct {
		name string
		args args
	}{
		{"1 string", args{[]any{"aaa"}}},
		{"2 strings", args{[]any{"ccccc", "bbbb"}}},
		{"string and int", args{[]any{"zzzzzzz", 1234}}},
		{"rune, bool and nil", args{[]any{'b', true, nil}}},
		{"array", args{[]any{[]string{"test", "spec", "dummy"}}}},
		{"map", args{[]any{map[string]int{"xxx": 100, "yyyy": 200}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Logger{}
			l.OutNewLog(tt.args.content...)
		})
	}
}

func TestNewTimeFmt(t *testing.T) {
	tests := []struct {
		name string
		want TimeFmtIf
	}{
		{"Cover NewTimeFmt codes.", NewTimeFmt()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewTimeFmt())
		})
	}
}

func TestTimeFmt_Time2str(t *testing.T) {
	type args struct {
		t time.Time
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"1-digits", args{time.Date(1987, 4, 1, 2, 4, 5, 0, time.UTC)}, "1987-04-01T02:04:05Z"},
		{"multi-digits", args{time.Date(2024, 10, 23, 15, 56, 41, 108, time.UTC)}, "2024-10-23T15:56:41Z"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &TimeFmt{}
			assert.Equal(t, tt.want, f.Time2str(tt.args.t))
		})
	}
}

func TestLogMemoryUsage(t *testing.T) {
	t.Log("Testing LogMemoryUsage function")
	LogMemoryUsage()
}

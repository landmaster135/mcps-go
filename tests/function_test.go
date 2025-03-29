package mytest

import (

	mypkg "example.com/mymodule/mypkg"

	"testing"

	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
)

func Test_RecordLog(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	type args struct {
		scriptName   string
		functionName string
		isRecording  bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, mypkg.RecordLog(tt.args.scriptName, tt.args.functionName, tt.args.isRecording))
		})
	}
}

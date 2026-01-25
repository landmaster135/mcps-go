package git

import (
	"os/exec"
)

// GitCommandExecutor はgitコマンド実行を抽象化するインターフェース
type GitCommandExecutor interface {
	Execute(workingDir string, args ...string) ([]byte, error)
}

// StandardGitExecutor は実際のgitコマンドを実行する実装
type StandardGitExecutor struct{}

// NewStandardGitExecutor は新しいStandardGitExecutorを作成する
func NewStandardGitExecutor() *StandardGitExecutor {
	return &StandardGitExecutor{}
}

// Execute はgitコマンドを実行する
func (e *StandardGitExecutor) Execute(workingDir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = workingDir
	return cmd.Output()
}

// MockGitExecutor はテスト用のモック実装
type MockGitExecutor struct {
	responses map[string][]byte
	errors    map[string]error
}

// NewMockGitExecutor は新しいMockGitExecutorを作成する
func NewMockGitExecutor() *MockGitExecutor {
	return &MockGitExecutor{
		responses: make(map[string][]byte),
		errors:    make(map[string]error),
	}
}

// SetResponse はコマンドに対するレスポンスを設定する
func (m *MockGitExecutor) SetResponse(command string, response []byte) {
	m.responses[command] = response
}

// SetError はコマンドに対するエラーを設定する
func (m *MockGitExecutor) SetError(command string, err error) {
	m.errors[command] = err
}

// Execute はモックされたgitコマンドを実行する
func (m *MockGitExecutor) Execute(workingDir string, args ...string) ([]byte, error) {
	command := joinArgs(args)

	if err, exists := m.errors[command]; exists {
		return nil, err
	}

	if response, exists := m.responses[command]; exists {
		return response, nil
	}

	// デフォルトレスポンス
	return []byte(""), nil
}

// joinArgs は引数を結合してコマンド文字列を作成する
func joinArgs(args []string) string {
	if len(args) == 0 {
		return ""
	}

	result := args[0]
	for i := 1; i < len(args); i++ {
		result += " " + args[i]
	}
	return result
}

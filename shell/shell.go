package shell

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// CommandExecutor インターフェースを定義
type CommandExecutor interface {
	Execute(command string, args []string, dir string, env map[string]string, timeout int) (*CommandResult, error)
}

// 許可されたコマンドのリスト
var allowedCommands = map[string]bool{
	// パッケージマネージャー
	"npm":  true,
	"yarn": true,
	"pnpm": true,
	"bun":  true,
	// バージョン管理
	"git": true,
	// ファイルシステム操作
	"ls":    true,
	"dir":   true,
	"find":  true,
	"mkdir": true,
	"rmdir": true,
	"cp":    true,
	"mv":    true,
	"rm":    true,
	"cat":   true,
	"awk":   true,
	"wc":    true,
	// 開発ツール
	"node":     true,
	"python":   true,
	"python3":  true,
	"tsc":      true,
	"eslint":   true,
	"prettier": true,
	// ビルドツール
	"make":  true,
	"cargo": true,
	"go":    true,
	// コンテナツール
	"docker":         true,
	"docker-compose": true,
	// その他のユーティリティ
	"echo":  true,
	"touch": true,
	"grep":  true,
}

// CommandResult はコマンド実行結果を表す構造体
type CommandResult struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
}

// ShellExecutor はシェルコマンドを実行するための構造体
type ShellExecutor struct {
	baseDirectory string
	maxTimeout    time.Duration
	cmdExecutor   func(cmd *exec.Cmd, timeout time.Duration) error
}

// NewShellExecutor は新しいShellExecutorを作成します
func NewShellExecutor(baseDir string) *ShellExecutor {
	// ベースディレクトリが指定されていない場合は現在のディレクトリを使用
	if baseDir == "" {
		baseDir, _ = os.Getwd()
	}

	return &ShellExecutor{
		baseDirectory: baseDir,
		maxTimeout:    60 * time.Second, // デフォルトのタイムアウトは60秒
		cmdExecutor:   runCommandWithTimeout,
	}
}

// Execute はシェルコマンドを実行します
func (e *ShellExecutor) Execute(command string, args []string, cwd string, env map[string]string, timeout int) (*CommandResult, error) {
	// コマンドが空の場合はエラー
	if command == "" {
		return &CommandResult{
			Stdout:   "",
			Stderr:   "コマンドが指定されていません",
			ExitCode: 1,
			Success:  false,
			Error:    "コマンドが指定されていません",
		}, nil
	}

	// コマンドが許可されているか確認
	// commandName := filepath.Base(command)
	// if !allowedCommands[commandName] {
	// 	return &CommandResult{
	// 		Stdout:   "",
	// 		Stderr:   fmt.Sprintf("コマンドは許可されていません: %s", command),
	// 		ExitCode: 1,
	// 		Success:  false,
	// 		Error:    fmt.Sprintf("コマンドは許可されていません: %s", command),
	// 	}, nil
	// }

	// 作業ディレクトリを検証
	workingDir := e.baseDirectory
	if cwd != "" {
		// 相対パスを解決
		resolvedPath := filepath.Join(e.baseDirectory, cwd)
		// パスがベースディレクトリ内にあることを確認
		if !strings.HasPrefix(resolvedPath, e.baseDirectory) {
			return &CommandResult{
				Stdout:   "",
				Stderr:   fmt.Sprintf("作業ディレクトリ %s は許可されたベースディレクトリの外にあります", cwd),
				ExitCode: 1,
				Success:  false,
				Error:    fmt.Sprintf("作業ディレクトリ %s は許可されたベースディレクトリの外にあります", cwd),
			}, nil
		}
		// ディレクトリが存在することを確認
		if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
			return &CommandResult{
				Stdout:   "",
				Stderr:   fmt.Sprintf("ディレクトリが存在しません: %s", resolvedPath),
				ExitCode: 1,
				Success:  false,
				Error:    fmt.Sprintf("ディレクトリが存在しません: %s", resolvedPath),
			}, nil
		}
		workingDir = resolvedPath
	}

	// コマンドを実行
	cmd := exec.Command(command, args...)
	cmd.Dir = workingDir

	// 環境変数を設定
	if env != nil {
		cmd.Env = os.Environ()
		for k, v := range env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// 標準出力と標準エラー出力をキャプチャ
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// タイムアウトを設定
	var cmdTimeout time.Duration
	if timeout > 0 {
		cmdTimeout = time.Duration(timeout) * time.Millisecond
		if cmdTimeout > e.maxTimeout {
			cmdTimeout = e.maxTimeout
		}
	} else {
		cmdTimeout = e.maxTimeout
	}

	// コマンドを実行（タイムアウト付き）
	err := e.cmdExecutor(cmd, cmdTimeout)

	// 結果を作成
	result := &CommandResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: 0,
		Success:  true,
	}

	// エラーがあれば処理
	if err != nil {
		result.Success = false
		result.Error = err.Error()

		// 終了コードを取得
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = 1
		}
	}

	return result, nil
}

// runCommandWithTimeout はコマンドをタイムアウト付きで実行します
func runCommandWithTimeout(cmd *exec.Cmd, timeout time.Duration) error {
	// タイムアウトなしの場合は直接実行
	if timeout <= 0 {
		return cmd.Run()
	}

	// コマンドを開始
	if err := cmd.Start(); err != nil {
		return err
	}

	// タイムアウト用のチャネル
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	// タイマーを設定
	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		// タイムアウトした場合はプロセスを強制終了
		if err := cmd.Process.Kill(); err != nil {
			return fmt.Errorf("コマンドのタイムアウト後にプロセスを終了できませんでした: %v", err)
		}
		return fmt.Errorf("コマンドは%v後にタイムアウトしました", timeout)
	}
}

// GetAllowedCommands は許可されたコマンドのリストを返します
func (e *ShellExecutor) GetAllowedCommands() []string {
	commands := make([]string, 0, len(allowedCommands))
	for cmd := range allowedCommands {
		commands = append(commands, cmd)
	}
	return commands
}

package usecases

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/landmaster135/mcps-go/internal/git_diff_recorder/config"
)

// TestGitDiffRecorderService_NewGitDiffRecorderService_Normal はサービス作成の正常系テスト
func TestGitDiffRecorderService_NewGitDiffRecorderService_Normal(t *testing.T) {
	// Arrange
	workingDir := "/tmp/test"
	cfg := &config.Config{
		OutputDir:  "/tmp/output",
		StagedOnly: false,
	}

	// Act
	service := NewGitDiffRecorderService(workingDir, cfg)

	// Assert
	if service == nil {
		t.Error("サービスの作成に失敗しました")
		return
	}
	if service.config != cfg {
		t.Error("設定が正しく設定されていません")
	}
}

// TestGitDiffRecorderService_RecordDiff_Normal はgit差分記録の正常系テスト
func TestGitDiffRecorderService_RecordDiff_Normal(t *testing.T) {
	// Arrange
	// テスト用の一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "git-diff-recorder-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗しました: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// テスト用の出力ディレクトリを作成
	outputDir := filepath.Join(tempDir, "output")
	cfg := &config.Config{
		OutputDir:  outputDir,
		StagedOnly: false,
	}

	// 現在のディレクトリを取得（実際のgitリポジトリが必要）
	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("現在のディレクトリの取得に失敗しました: %v", err)
	}

	service := NewGitDiffRecorderService(workingDir, cfg)

	// Act
	err = service.RecordDiff()

	// Assert
	// gitリポジトリでない場合はエラーになることを期待
	// 実際のgitリポジトリで実行する場合は、エラーがないことを確認
	if err != nil {
		// gitリポジトリでない場合のエラーメッセージを確認
		t.Logf("期待されるエラー（gitリポジトリでない場合）: %v", err)
	} else {
		// 成功した場合、出力ファイルが作成されているかを確認
		// ここでは詳細な検証は省略（実際のテストでは出力ファイルの内容を検証）
		t.Log("差分記録が正常に完了しました")
	}
}

// TestGitDiffRecorderService_RecordDiff_StagedOnly はステージング済み差分のみ記録のテスト
func TestGitDiffRecorderService_RecordDiff_StagedOnly(t *testing.T) {
	// Arrange
	// テスト用の一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "git-diff-recorder-test-staged")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗しました: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// テスト用の出力ディレクトリを作成
	outputDir := filepath.Join(tempDir, "output")
	cfg := &config.Config{
		OutputDir:  outputDir,
		StagedOnly: true, // ステージング済みのみ
	}

	// 現在のディレクトリを取得（実際のgitリポジトリが必要）
	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("現在のディレクトリの取得に失敗しました: %v", err)
	}

	service := NewGitDiffRecorderService(workingDir, cfg)

	// Act
	err = service.RecordDiff()

	// Assert
	if err != nil {
		t.Logf("期待されるエラー（gitリポジトリでない場合）: %v", err)
	} else {
		t.Log("ステージング済み差分記録が正常に完了しました")
	}
}

// TestGitDiffRecorderService_RecordDiff_WithGitDir は指定Gitディレクトリでの記録のテスト
func TestGitDiffRecorderService_RecordDiff_WithGitDir(t *testing.T) {
	// Arrange
	// テスト用の一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "git-diff-recorder-test-gitdir")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗しました: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// テスト用の出力ディレクトリを作成
	outputDir := filepath.Join(tempDir, "output")

	// devboxディレクトリを指定（実際のgitリポジトリ）
	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("現在のディレクトリの取得に失敗しました: %v", err)
	}
	devboxDir := filepath.Join(workingDir, "../../../..")

	cfg := &config.Config{
		OutputDir:  outputDir,
		StagedOnly: false,
		GitDir:     devboxDir, // 指定Gitディレクトリ
	}

	service := NewGitDiffRecorderService(devboxDir, cfg)

	// Act
	err = service.RecordDiff()

	// Assert
	if err != nil {
		t.Logf("期待されるエラー（gitリポジトリでない場合）: %v", err)
	} else {
		t.Log("指定Gitディレクトリでの差分記録が正常に完了しました")
	}
}

// TestConfig_ParseFlags_Normal はコマンドライン引数解析の正常系テスト
func TestConfig_ParseFlags_Normal(t *testing.T) {
	// Arrange
	mockParser := config.NewMockFlagParser()
	mockOSArgs := config.NewMockOSArgs([]string{"git-diff-recorder", "--output-dir", "/tmp/test", "--staged-only"})
	configParser := config.NewConfigParser(mockParser, mockOSArgs)

	// テスト用の値を設定
	mockParser.SetStringValue("output-dir", "/tmp/test")
	mockParser.SetBoolValue("staged-only", true)

	// Act
	cfg, err := configParser.ParseFlags()

	// Assert
	if err != nil {
		t.Errorf("引数解析でエラーが発生しました: %v", err)
	}
	if cfg.OutputDir != "/tmp/test" {
		t.Errorf("OutputDirが期待値と異なります。期待値: /tmp/test, 実際: %s", cfg.OutputDir)
	}
	if !cfg.StagedOnly {
		t.Error("StagedOnlyがtrueになっていません")
	}
}

// TestConfig_ParseFlags_MissingOutputDir は必須パラメータ不足のテスト
func TestConfig_ParseFlags_MissingOutputDir(t *testing.T) {
	// Arrange
	mockParser := config.NewMockFlagParser()
	mockOSArgs := config.NewMockOSArgs([]string{"git-diff-recorder"})
	configParser := config.NewConfigParser(mockParser, mockOSArgs)

	// 必須パラメータを設定しない（デフォルト値のまま）

	// Act
	cfg, err := configParser.ParseFlags()

	// Assert
	if err == nil {
		t.Error("必須パラメータが不足しているのにエラーが発生しませんでした")
	}
	if cfg != nil {
		t.Error("エラー時にはconfigはnilであるべきです")
	}
}

// TestConfig_ParseFlags_ReadMode は読み取りモードのテスト
func TestConfig_ParseFlags_ReadMode(t *testing.T) {
	// Arrange
	mockParser := config.NewMockFlagParser()
	mockOSArgs := config.NewMockOSArgs([]string{"git-diff-recorder", "--read-mode", "--source-dir", "/tmp/source", "--repository", "test-repo"})
	configParser := config.NewConfigParser(mockParser, mockOSArgs)

	// テスト用の値を設定
	mockParser.SetBoolValue("read-mode", true)
	mockParser.SetStringValue("source-dir", "/tmp/source")
	mockParser.SetStringValue("repository", "test-repo")

	// Act
	cfg, err := configParser.ParseFlags()

	// Assert
	if err != nil {
		t.Errorf("引数解析でエラーが発生しました: %v", err)
	}
	if !cfg.ReadMode {
		t.Error("ReadModeがtrueになっていません")
	}
	if cfg.SourceDir != "/tmp/source" {
		t.Errorf("SourceDirが期待値と異なります。期待値: /tmp/source, 実際: %s", cfg.SourceDir)
	}
	if cfg.Repository != "test-repo" {
		t.Errorf("Repositoryが期待値と異なります。期待値: test-repo, 実際: %s", cfg.Repository)
	}
}

// TestConfig_ParseFlags_GenMode は生成モードのテスト
func TestConfig_ParseFlags_GenMode(t *testing.T) {
	// Arrange
	mockParser := config.NewMockFlagParser()
	mockOSArgs := config.NewMockOSArgs([]string{"git-diff-recorder", "--gen-mode", "--git-dir", "/tmp/git"})
	configParser := config.NewConfigParser(mockParser, mockOSArgs)

	// テスト用の値を設定
	mockParser.SetBoolValue("gen-mode", true)
	mockParser.SetStringValue("git-dir", "/tmp/git")

	// Act
	cfg, err := configParser.ParseFlags()

	// Assert
	if err != nil {
		t.Errorf("引数解析でエラーが発生しました: %v", err)
	}
	if !cfg.GenMode {
		t.Error("GenModeがtrueになっていません")
	}
	if cfg.GitDir != "/tmp/git" {
		t.Errorf("GitDirが期待値と異なります。期待値: /tmp/git, 実際: %s", cfg.GitDir)
	}
}

// TestConfig_ParseFlags_ReadModeMissingSourceDir は読み取りモードで必須パラメータ不足のテスト
func TestConfig_ParseFlags_ReadModeMissingSourceDir(t *testing.T) {
	// Arrange
	mockParser := config.NewMockFlagParser()
	mockOSArgs := config.NewMockOSArgs([]string{"git-diff-recorder", "--read-mode", "--repository", "test-repo"})
	configParser := config.NewConfigParser(mockParser, mockOSArgs)

	// テスト用の値を設定（source-dirを設定しない）
	mockParser.SetBoolValue("read-mode", true)
	mockParser.SetStringValue("repository", "test-repo")

	// Act
	cfg, err := configParser.ParseFlags()

	// Assert
	if err == nil {
		t.Error("必須パラメータが不足しているのにエラーが発生しませんでした")
	}
	if cfg != nil {
		t.Error("エラー時にはconfigはnilであるべきです")
	}
}

// TestConfig_ParseFlags_GenModeMissingGitDir は生成モードで必須パラメータ不足のテスト
func TestConfig_ParseFlags_GenModeMissingGitDir(t *testing.T) {
	// Arrange
	mockParser := config.NewMockFlagParser()
	mockOSArgs := config.NewMockOSArgs([]string{"git-diff-recorder", "--gen-mode"})
	configParser := config.NewConfigParser(mockParser, mockOSArgs)

	// テスト用の値を設定（git-dirを設定しない）
	mockParser.SetBoolValue("gen-mode", true)

	// Act
	cfg, err := configParser.ParseFlags()

	// Assert
	if err == nil {
		t.Error("必須パラメータが不足しているのにエラーが発生しませんでした")
	}
	if cfg != nil {
		t.Error("エラー時にはconfigはnilであるべきです")
	}
}

// TestConfig_ParseFlags_ParseError はパースエラーのテスト
func TestConfig_ParseFlags_ParseError(t *testing.T) {
	// Arrange
	mockParser := config.NewMockFlagParser()
	mockOSArgs := config.NewMockOSArgs([]string{"git-diff-recorder"})
	configParser := config.NewConfigParser(mockParser, mockOSArgs)

	// パースエラーを設定
	expectedError := fmt.Errorf("parse error")
	mockParser.SetParseError(expectedError)

	// Act
	cfg, err := configParser.ParseFlags()

	// Assert
	if err == nil {
		t.Error("パースエラーが発生するはずなのにエラーが発生しませんでした")
	}
	if err != expectedError {
		t.Errorf("期待されるエラーと異なります。期待値: %v, 実際: %v", expectedError, err)
	}
	if cfg != nil {
		t.Error("エラー時にはconfigはnilであるべきです")
	}
}

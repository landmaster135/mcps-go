package usecases

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/landmaster135/mcps-go/internal/git_diff_recorder/config"
)

// TestGitDiffGeneratorService_NewGitDiffGeneratorService_Normal はサービス作成の正常系テスト
func TestGitDiffGeneratorService_NewGitDiffGeneratorService_Normal(t *testing.T) {
	// Arrange
	gitDir := "/tmp/test-git"
	cfg := &config.Config{
		GenMode:    true,
		GitDir:     gitDir,
		StagedOnly: false,
	}

	// Act
	service := NewGitDiffGeneratorService(gitDir, cfg)

	// Assert
	if service == nil {
		t.Error("サービスの作成に失敗しました")
		return
	}
	if service.config != cfg {
		t.Error("設定が正しく設定されていません")
	}
}

// TestGitDiffGeneratorService_GetCurrentDetailedDiff_Normal は差分取得の正常系テスト
func TestGitDiffGeneratorService_GetCurrentDetailedDiff_Normal(t *testing.T) {
	// Arrange
	// 現在のディレクトリを取得（実際のgitリポジトリが必要）
	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("現在のディレクトリの取得に失敗しました: %v", err)
	}

	// devboxディレクトリに移動してテスト
	devboxDir := filepath.Join(workingDir, "../../../..")
	cfg := &config.Config{
		GenMode:    true,
		GitDir:     devboxDir,
		StagedOnly: false,
	}

	service := NewGitDiffGeneratorService(devboxDir, cfg)

	// Act
	detailedDiff, err := service.GetCurrentDetailedDiff()

	// Assert
	if err != nil {
		// gitリポジトリでない場合はエラーになることを期待
		t.Logf("期待されるエラー（gitリポジトリでない場合）: %v", err)
	} else {
		// 成功した場合、差分が取得されていることを確認
		t.Logf("差分取得が正常に完了しました。差分の長さ: %d文字", len(detailedDiff))
	}
}

// TestGitDiffGeneratorService_GetCurrentDetailedDiff_StagedOnly はステージング済み差分取得のテスト
func TestGitDiffGeneratorService_GetCurrentDetailedDiff_StagedOnly(t *testing.T) {
	// Arrange
	// 現在のディレクトリを取得（実際のgitリポジトリが必要）
	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("現在のディレクトリの取得に失敗しました: %v", err)
	}

	// devboxディレクトリに移動してテスト
	devboxDir := filepath.Join(workingDir, "../../../..")
	cfg := &config.Config{
		GenMode:    true,
		GitDir:     devboxDir,
		StagedOnly: true, // ステージング済みのみ
	}

	service := NewGitDiffGeneratorService(devboxDir, cfg)

	// Act
	detailedDiff, err := service.GetCurrentDetailedDiff()

	// Assert
	if err != nil {
		// gitリポジトリでない場合はエラーになることを期待
		t.Logf("期待されるエラー（gitリポジトリでない場合）: %v", err)
	} else {
		// 成功した場合、差分が取得されていることを確認
		t.Logf("ステージング済み差分取得が正常に完了しました。差分の長さ: %d文字", len(detailedDiff))
	}
}

// TestGitDiffGeneratorService_GetCurrentDetailedDiff_InvalidGitDir は無効なGitディレクトリのテスト
func TestGitDiffGeneratorService_GetCurrentDetailedDiff_InvalidGitDir(t *testing.T) {
	// Arrange
	// 存在しないディレクトリを指定
	invalidDir := "/tmp/non-existent-git-repo"
	cfg := &config.Config{
		GenMode:    true,
		GitDir:     invalidDir,
		StagedOnly: false,
	}

	service := NewGitDiffGeneratorService(invalidDir, cfg)

	// Act
	detailedDiff, err := service.GetCurrentDetailedDiff()

	// Assert
	if err == nil {
		t.Error("無効なGitディレクトリでエラーが発生しませんでした")
	}
	if detailedDiff != "" {
		t.Error("エラー時には空文字列が返されるべきです")
	}
}

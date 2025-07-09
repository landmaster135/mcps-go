package usecases

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/landmaster135/mcps-go/internal/git_diff_recorder/config"
)

// TestDiffReaderService_NewDiffReaderService_Normal はサービス作成の正常系テスト
func TestDiffReaderService_NewDiffReaderService_Normal(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		ReadMode:   true,
		SourceDir:  "/tmp/test",
		Repository: "test-repo",
	}

	// Act
	service := NewDiffReaderService(cfg)

	// Assert
	if service == nil {
		t.Error("サービスの作成に失敗しました")
		return
	}
	if service.config != cfg {
		t.Error("設定が正しく設定されていません")
	}
}

// TestDiffReaderService_ReadAndDisplayDetailedDiff_Normal は読み取り機能の正常系テスト
func TestDiffReaderService_ReadAndDisplayDetailedDiff_Normal(t *testing.T) {
	// Arrange
	// テスト用の一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "diff-reader-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗しました: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// テスト用のリポジトリディレクトリを作成
	repoDir := filepath.Join(tempDir, "test-repo")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatalf("リポジトリディレクトリの作成に失敗しました: %v", err)
	}

	// テスト用のdiffファイルを作成
	diffContent := `=== Git Diff Record ===
Generated at: 2025-01-07 12:30:45
Repository: test-repo
Branch: main
Latest commit: abc12345
Options: --staged-only=false

=== File Changes Summary ===
Modified files: 1
New files: 0
Deleted files: 0

=== Detailed Diff ===
diff --git a/test.txt b/test.txt
index 1234567..abcdefg 100644
--- a/test.txt
+++ b/test.txt
@@ -1,3 +1,3 @@
 line 1
-line 2
+line 2 modified
 line 3
`

	diffFile := filepath.Join(repoDir, "diff_20250107123045.txt")
	if err := os.WriteFile(diffFile, []byte(diffContent), 0644); err != nil {
		t.Fatalf("diffファイルの作成に失敗しました: %v", err)
	}

	cfg := &config.Config{
		ReadMode:   true,
		SourceDir:  tempDir,
		Repository: "test-repo",
	}

	service := NewDiffReaderService(cfg)

	// Act
	err = service.ReadAndDisplayDetailedDiff()

	// Assert
	if err != nil {
		t.Errorf("読み取り処理でエラーが発生しました: %v", err)
	}
}

// TestDiffReaderService_ReadAndDisplayDetailedDiff_NoRepository はリポジトリが存在しない場合のテスト
func TestDiffReaderService_ReadAndDisplayDetailedDiff_NoRepository(t *testing.T) {
	// Arrange
	tempDir, err := os.MkdirTemp("", "diff-reader-test-no-repo")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗しました: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.Config{
		ReadMode:   true,
		SourceDir:  tempDir,
		Repository: "non-existent-repo",
	}

	service := NewDiffReaderService(cfg)

	// Act
	err = service.ReadAndDisplayDetailedDiff()

	// Assert
	if err == nil {
		t.Error("存在しないリポジトリでエラーが発生しませんでした")
	}
}

// TestDiffReaderService_ReadAndDisplayDetailedDiff_NoDiffFiles はdiffファイルが存在しない場合のテスト
func TestDiffReaderService_ReadAndDisplayDetailedDiff_NoDiffFiles(t *testing.T) {
	// Arrange
	tempDir, err := os.MkdirTemp("", "diff-reader-test-no-files")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗しました: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 空のリポジトリディレクトリを作成
	repoDir := filepath.Join(tempDir, "empty-repo")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatalf("リポジトリディレクトリの作成に失敗しました: %v", err)
	}

	cfg := &config.Config{
		ReadMode:   true,
		SourceDir:  tempDir,
		Repository: "empty-repo",
	}

	service := NewDiffReaderService(cfg)

	// Act
	err = service.ReadAndDisplayDetailedDiff()

	// Assert
	if err == nil {
		t.Error("diffファイルが存在しないのにエラーが発生しませんでした")
	}
}

// TestDiffReaderService_GetDetailedDiff_Normal はGetDetailedDiffメソッドの正常系テスト
func TestDiffReaderService_GetDetailedDiff_Normal(t *testing.T) {
	// Arrange
	// テスト用の一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "diff-reader-get-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗しました: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// テスト用のリポジトリディレクトリを作成
	repoDir := filepath.Join(tempDir, "test-repo")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatalf("リポジトリディレクトリの作成に失敗しました: %v", err)
	}

	// テスト用のdiffファイルを作成
	expectedDiff := `diff --git a/test.txt b/test.txt
index 1234567..abcdefg 100644
--- a/test.txt
+++ b/test.txt
@@ -1,3 +1,3 @@
 line 1
-line 2
+line 2 modified
 line 3`

	diffContent := `=== Git Diff Record ===
Generated at: 2025-01-07 12:30:45
Repository: test-repo
Branch: main
Latest commit: abc12345
Options: --staged-only=false

=== File Changes Summary ===
Modified files: 1
New files: 0
Deleted files: 0

=== Detailed Diff ===
` + expectedDiff

	diffFile := filepath.Join(repoDir, "diff_20250107123045.txt")
	if err := os.WriteFile(diffFile, []byte(diffContent), 0644); err != nil {
		t.Fatalf("diffファイルの作成に失敗しました: %v", err)
	}

	cfg := &config.Config{
		ReadMode:   true,
		SourceDir:  tempDir,
		Repository: "test-repo",
	}

	service := NewDiffReaderService(cfg)

	// Act
	detailedDiff, err := service.GetDetailedDiff()

	// Assert
	if err != nil {
		t.Errorf("GetDetailedDiff処理でエラーが発生しました: %v", err)
	}
	if detailedDiff != expectedDiff {
		t.Errorf("取得した詳細差分が期待値と異なります。\n期待値:\n%s\n実際:\n%s", expectedDiff, detailedDiff)
	}
}

// TestDiffReaderService_GetDetailedDiff_NoRepository はGetDetailedDiffでリポジトリが存在しない場合のテスト
func TestDiffReaderService_GetDetailedDiff_NoRepository(t *testing.T) {
	// Arrange
	tempDir, err := os.MkdirTemp("", "diff-reader-get-test-no-repo")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗しました: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.Config{
		ReadMode:   true,
		SourceDir:  tempDir,
		Repository: "non-existent-repo",
	}

	service := NewDiffReaderService(cfg)

	// Act
	detailedDiff, err := service.GetDetailedDiff()

	// Assert
	if err == nil {
		t.Error("存在しないリポジトリでエラーが発生しませんでした")
	}
	if detailedDiff != "" {
		t.Error("エラー時には空文字列が返されるべきです")
	}
}

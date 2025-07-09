package reader

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	config "github.com/landmaster135/mcps-go/internal/git_diff_recorder/config"
)

// TestDiffReader_NewDiffReader_Normal はDiffReader作成の正常系テスト
func TestDiffReader_NewDiffReader_Normal(t *testing.T) {
	// Arrange
	sourceDir := "/tmp/test"

	// Act
	reader := NewDiffReader(sourceDir)

	// Assert
	if reader == nil {
		t.Error("DiffReaderの作成に失敗しました")
		return
	}
	if reader.sourceDir != sourceDir {
		t.Errorf("sourceDirが期待値と異なります。期待値: %s, 実際: %s", sourceDir, reader.sourceDir)
	}
}

// TestDiffReader_FindLatestDiffFile_Normal はFindLatestDiffFile正常系テスト
func TestDiffReader_FindLatestDiffFile_Normal(t *testing.T) {
	// Arrange
	tempDir, err := os.MkdirTemp("", "test_reader")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗しました: %v", err)
	}
	defer os.RemoveAll(tempDir)

	reader := NewDiffReader(tempDir)
	repository := "test-repo"
	repoDir := filepath.Join(tempDir, repository)

	// リポジトリディレクトリを作成
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatalf("リポジトリディレクトリの作成に失敗しました: %v", err)
	}

	// テスト用のdiffファイルを作成
	testFiles := []string{
		"diff_20230101120000.txt",
		"diff_20230101130000.txt",
		"diff_20230101140000.txt",
	}

	for _, filename := range testFiles {
		filePath := filepath.Join(repoDir, filename)
		if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
			t.Fatalf("テストファイルの作成に失敗しました: %v", err)
		}
	}

	// Act
	latestFile, err := reader.FindLatestDiffFile(repository)

	// Assert
	if err != nil {
		t.Errorf("FindLatestDiffFileでエラーが発生しました: %v", err)
	}

	expectedFile := filepath.Join(repoDir, "diff_20230101140000.txt")
	if latestFile != expectedFile {
		t.Errorf("最新ファイルが期待値と異なります。期待値: %s, 実際: %s", expectedFile, latestFile)
	}
}

// TestDiffReader_FindLatestDiffFile_RepositoryNotFound はリポジトリディレクトリ不存在のテスト
func TestDiffReader_FindLatestDiffFile_RepositoryNotFound(t *testing.T) {
	// Arrange
	tempDir, err := os.MkdirTemp("", "test_reader")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗しました: %v", err)
	}
	defer os.RemoveAll(tempDir)

	reader := NewDiffReader(tempDir)
	repository := "nonexistent-repo"

	// Act
	_, err = reader.FindLatestDiffFile(repository)

	// Assert
	if err == nil {
		t.Error("存在しないリポジトリでエラーが発生するべきです")
	}
	if !strings.Contains(err.Error(), "リポジトリディレクトリが見つかりません") {
		t.Errorf("期待されるエラーメッセージと異なります: %v", err)
	}
}

// TestDiffReader_FindLatestDiffFile_NoDiffFiles はdiffファイル不存在のテスト
func TestDiffReader_FindLatestDiffFile_NoDiffFiles(t *testing.T) {
	// Arrange
	tempDir, err := os.MkdirTemp("", "test_reader")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗しました: %v", err)
	}
	defer os.RemoveAll(tempDir)

	reader := NewDiffReader(tempDir)
	repository := "test-repo"
	repoDir := filepath.Join(tempDir, repository)

	// リポジトリディレクトリを作成（diffファイルは作成しない）
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatalf("リポジトリディレクトリの作成に失敗しました: %v", err)
	}

	// Act
	_, err = reader.FindLatestDiffFile(repository)

	// Assert
	if err == nil {
		t.Error("diffファイルが存在しない場合にエラーが発生するべきです")
	}
	if !strings.Contains(err.Error(), "diffファイルが見つかりません") {
		t.Errorf("期待されるエラーメッセージと異なります: %v", err)
	}
}

// TestDiffReader_ExtractDetailedDiff_Normal はExtractDetailedDiff正常系テスト
func TestDiffReader_ExtractDetailedDiff_Normal(t *testing.T) {
	// Arrange
	tempDir, err := os.MkdirTemp("", "test_reader")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗しました: %v", err)
	}
	defer os.RemoveAll(tempDir)

	reader := NewDiffReader(tempDir)

	// テスト用のdiffファイル内容を作成
	fileContent := `=== Git Diff Record ===
Generated at: 2023-01-01 12:00:00
Repository: test-repo
Branch: main

=== File Changes Summary ===
Modified files: 1
New files: 0
Deleted files: 0

` + config.HeaderDetailedDiff + `
diff --git a/test.txt b/test.txt
index 1234567..abcdefg 100644
--- a/test.txt
+++ b/test.txt
@@ -1,3 +1,4 @@
 line 1
 line 2
+added line
 line 3
`

	// テストファイルを作成
	testFile := filepath.Join(tempDir, "test_diff.txt")
	if err := os.WriteFile(testFile, []byte(fileContent), 0644); err != nil {
		t.Fatalf("テストファイルの作成に失敗しました: %v", err)
	}

	// Act
	result, err := reader.ExtractDetailedDiff(testFile)

	// Assert
	if err != nil {
		t.Errorf("ExtractDetailedDiffでエラーが発生しました: %v", err)
	}

	expectedParts := []string{
		"diff --git a/test.txt b/test.txt",
		"index 1234567..abcdefg 100644",
		"--- a/test.txt",
		"+++ b/test.txt",
		"+added line",
	}

	for _, expected := range expectedParts {
		if !strings.Contains(result, expected) {
			t.Errorf("期待される内容が含まれていません: %s", expected)
		}
	}
}

// TestDiffReader_ExtractDetailedDiff_NoSection はDetailedDiffセクション不存在のテスト
func TestDiffReader_ExtractDetailedDiff_NoSection(t *testing.T) {
	// Arrange
	tempDir, err := os.MkdirTemp("", "test_reader")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗しました: %v", err)
	}
	defer os.RemoveAll(tempDir)

	reader := NewDiffReader(tempDir)

	// Detailed Diffセクションがないファイル内容
	fileContent := `=== Git Diff Record ===
Generated at: 2023-01-01 12:00:00
Repository: test-repo
Branch: main

=== File Changes Summary ===
Modified files: 1
New files: 0
Deleted files: 0
`

	// テストファイルを作成
	testFile := filepath.Join(tempDir, "test_diff.txt")
	if err := os.WriteFile(testFile, []byte(fileContent), 0644); err != nil {
		t.Fatalf("テストファイルの作成に失敗しました: %v", err)
	}

	// Act
	_, err = reader.ExtractDetailedDiff(testFile)

	// Assert
	if err == nil {
		t.Error("Detailed Diffセクションが存在しない場合にエラーが発生するべきです")
	}
	if !strings.Contains(err.Error(), "セクションが見つかりませんでした") {
		t.Errorf("期待されるエラーメッセージと異なります: %v", err)
	}
}

// TestDiffReader_ExtractDetailedDiff_FileNotFound はファイル不存在のテスト
func TestDiffReader_ExtractDetailedDiff_FileNotFound(t *testing.T) {
	// Arrange
	reader := NewDiffReader("/tmp")
	nonexistentFile := "/tmp/nonexistent_file.txt"

	// Act
	_, err := reader.ExtractDetailedDiff(nonexistentFile)

	// Assert
	if err == nil {
		t.Error("存在しないファイルでエラーが発生するべきです")
	}
	if !strings.Contains(err.Error(), "ファイルを開けませんでした") {
		t.Errorf("期待されるエラーメッセージと異なります: %v", err)
	}
}

// TestDiffReader_GetFileInfo_Normal はGetFileInfo正常系テスト
func TestDiffReader_GetFileInfo_Normal(t *testing.T) {
	// Arrange
	tempDir, err := os.MkdirTemp("", "test_reader")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗しました: %v", err)
	}
	defer os.RemoveAll(tempDir)

	reader := NewDiffReader(tempDir)

	// テスト用のdiffファイル内容を作成
	fileContent := `=== Git Diff Record ===
Generated at: 2023-01-01 12:00:00
Repository: test-repo
Branch: main
Latest commit: abc12345
Options: --staged-only=false

=== File Changes Summary ===
Modified files: 1
New files: 0
Deleted files: 0

` + config.HeaderDetailedDiff + `
diff content here
`

	// テストファイルを作成
	testFile := filepath.Join(tempDir, "diff_20230101120000.txt")
	if err := os.WriteFile(testFile, []byte(fileContent), 0644); err != nil {
		t.Fatalf("テストファイルの作成に失敗しました: %v", err)
	}

	// Act
	info, err := reader.GetFileInfo(testFile)

	// Assert
	if err != nil {
		t.Errorf("GetFileInfoでエラーが発生しました: %v", err)
	}

	if info.FilePath != testFile {
		t.Errorf("FilePathが期待値と異なります。期待値: %s, 実際: %s", testFile, info.FilePath)
	}
	if info.FileName != "diff_20230101120000.txt" {
		t.Errorf("FileNameが期待値と異なります。期待値: diff_20230101120000.txt, 実際: %s", info.FileName)
	}
	if info.Repository != "test-repo" {
		t.Errorf("Repositoryが期待値と異なります。期待値: test-repo, 実際: %s", info.Repository)
	}
	if info.Branch != "main" {
		t.Errorf("Branchが期待値と異なります。期待値: main, 実際: %s", info.Branch)
	}
	if info.Options != "--staged-only=false" {
		t.Errorf("Optionsが期待値と異なります。期待値: --staged-only=false, 実際: %s", info.Options)
	}

	// GeneratedAtの確認
	expectedTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	if !info.GeneratedAt.Equal(expectedTime) {
		t.Errorf("GeneratedAtが期待値と異なります。期待値: %v, 実際: %v", expectedTime, info.GeneratedAt)
	}

	// ModTimeが設定されているかチェック
	if info.ModTime.IsZero() {
		t.Error("ModTimeが設定されていません")
	}
}

// TestDiffReader_GetFileInfo_FileNotFound はファイル不存在のテスト
func TestDiffReader_GetFileInfo_FileNotFound(t *testing.T) {
	// Arrange
	reader := NewDiffReader("/tmp")
	nonexistentFile := "/tmp/nonexistent_file.txt"

	// Act
	_, err := reader.GetFileInfo(nonexistentFile)

	// Assert
	if err == nil {
		t.Error("存在しないファイルでエラーが発生するべきです")
	}
	if !strings.Contains(err.Error(), "ファイルを開けませんでした") {
		t.Errorf("期待されるエラーメッセージと異なります: %v", err)
	}
}

// TestDiffReader_GetFileInfo_PartialInfo は部分的な情報のテスト
func TestDiffReader_GetFileInfo_PartialInfo(t *testing.T) {
	// Arrange
	tempDir, err := os.MkdirTemp("", "test_reader")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗しました: %v", err)
	}
	defer os.RemoveAll(tempDir)

	reader := NewDiffReader(tempDir)

	// 部分的な情報のみのファイル内容
	fileContent := `=== Git Diff Record ===
Generated at: 2023-01-01 12:00:00
Repository: test-repo

=== File Changes Summary ===
Modified files: 1
`

	// テストファイルを作成
	testFile := filepath.Join(tempDir, "diff_partial.txt")
	if err := os.WriteFile(testFile, []byte(fileContent), 0644); err != nil {
		t.Fatalf("テストファイルの作成に失敗しました: %v", err)
	}

	// Act
	info, err := reader.GetFileInfo(testFile)

	// Assert
	if err != nil {
		t.Errorf("GetFileInfoでエラーが発生しました: %v", err)
	}

	if info.Repository != "test-repo" {
		t.Errorf("Repositoryが期待値と異なります。期待値: test-repo, 実際: %s", info.Repository)
	}
	if info.Branch != "" {
		t.Errorf("Branchは空であるべきです。実際: %s", info.Branch)
	}
	if info.Options != "" {
		t.Errorf("Optionsは空であるべきです。実際: %s", info.Options)
	}
}

// TestDiffReader_ExtractDetailedDiff_EmptyDiff は空の差分のテスト
func TestDiffReader_ExtractDetailedDiff_EmptyDiff(t *testing.T) {
	// Arrange
	tempDir, err := os.MkdirTemp("", "test_reader")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗しました: %v", err)
	}
	defer os.RemoveAll(tempDir)

	reader := NewDiffReader(tempDir)

	// 空の差分セクションを持つファイル内容
	fileContent := `=== Git Diff Record ===
Generated at: 2023-01-01 12:00:00

` + config.HeaderDetailedDiff + `
差分はありません。
`

	// テストファイルを作成
	testFile := filepath.Join(tempDir, "test_diff.txt")
	if err := os.WriteFile(testFile, []byte(fileContent), 0644); err != nil {
		t.Fatalf("テストファイルの作成に失敗しました: %v", err)
	}

	// Act
	result, err := reader.ExtractDetailedDiff(testFile)

	// Assert
	if err != nil {
		t.Errorf("ExtractDetailedDiffでエラーが発生しました: %v", err)
	}

	if result != "差分はありません。" {
		t.Errorf("期待される内容と異なります。期待値: 差分はありません。, 実際: %s", result)
	}
}

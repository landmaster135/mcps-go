package file

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	config "github.com/landmaster135/mcps-go/internal/git_diff_recorder/config"
)

// TestWriter_NewWriter_Normal はWriter作成の正常系テスト
func TestWriter_NewWriter_Normal(t *testing.T) {
	// Arrange
	outputDir := "/tmp/test"

	// Act
	writer := NewWriter(outputDir)

	// Assert
	if writer == nil {
		t.Error("Writerの作成に失敗しました")
		return
	}
	if writer.outputDir != outputDir {
		t.Errorf("outputDirが期待値と異なります。期待値: %s, 実際: %s", outputDir, writer.outputDir)
	}
}

// TestWriter_WriteDiffRecord_Normal はWriteDiffRecord正常系テスト
func TestWriter_WriteDiffRecord_Normal(t *testing.T) {
	// Arrange
	tempDir, err := os.MkdirTemp("", "test_writer")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗しました: %v", err)
	}
	defer os.RemoveAll(tempDir)

	writer := NewWriter(tempDir)
	repoName := "test-repo"
	record := &DiffRecord{
		GeneratedAt:     time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		Repository:      "test-repo",
		Branch:          "main",
		LatestCommit:    "abc12345",
		StagedOnly:      false,
		ModifiedFiles:   2,
		NewFiles:        []string{"new1.txt", "new2.txt"},
		DeletedFiles:    []string{"deleted1.txt"},
		DiffOutput:      "diff --git a/test.txt b/test.txt\n+added line",
	}

	// Act
	err = writer.WriteDiffRecord(repoName, record)

	// Assert
	if err != nil {
		t.Errorf("WriteDiffRecordでエラーが発生しました: %v", err)
	}

	// ディレクトリが作成されているかチェック
	repoDir := filepath.Join(tempDir, repoName)
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		t.Error("リポジトリディレクトリが作成されていません")
	}

	// ファイルが作成されているかチェック
	files, err := filepath.Glob(filepath.Join(repoDir, "diff_*.txt"))
	if err != nil {
		t.Errorf("ファイル検索でエラーが発生しました: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("作成されたファイル数が期待値と異なります。期待値: 1, 実際: %d", len(files))
	}

	// ファイル内容をチェック
	content, err := os.ReadFile(files[0])
	if err != nil {
		t.Errorf("ファイル読み取りでエラーが発生しました: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, config.HeaderGitDiffRecord) {
		t.Error("ヘッダーが含まれていません")
	}
	if !strings.Contains(contentStr, "Repository: test-repo") {
		t.Error("リポジトリ情報が含まれていません")
	}
	if !strings.Contains(contentStr, "Branch: main") {
		t.Error("ブランチ情報が含まれていません")
	}
	if !strings.Contains(contentStr, "new1.txt") {
		t.Error("新規ファイル情報が含まれていません")
	}
	if !strings.Contains(contentStr, "deleted1.txt") {
		t.Error("削除ファイル情報が含まれていません")
	}
	if !strings.Contains(contentStr, "diff --git a/test.txt b/test.txt") {
		t.Error("差分情報が含まれていません")
	}
}

// TestWriter_WriteDiffRecord_DirectoryCreationError はディレクトリ作成エラーのテスト
func TestWriter_WriteDiffRecord_DirectoryCreationError(t *testing.T) {
	// Arrange
	writer := NewWriter("/invalid/path/that/cannot/be/created")
	repoName := "test-repo"
	record := &DiffRecord{
		GeneratedAt:  time.Now(),
		Repository:   "test-repo",
		Branch:       "main",
		LatestCommit: "abc12345",
	}

	// Act
	err := writer.WriteDiffRecord(repoName, record)

	// Assert
	if err == nil {
		t.Error("無効なパスでディレクトリ作成エラーが発生するべきです")
	}
	if !strings.Contains(err.Error(), "ディレクトリの作成に失敗しました") {
		t.Errorf("期待されるエラーメッセージと異なります: %v", err)
	}
}

// TestWriter_formatDiffRecord_WithAllData は全データありのフォーマットテスト
func TestWriter_formatDiffRecord_WithAllData(t *testing.T) {
	// Arrange
	writer := NewWriter("/tmp")
	record := &DiffRecord{
		GeneratedAt:     time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		Repository:      "test-repo",
		Branch:          "main",
		LatestCommit:    "abc12345",
		StagedOnly:      true,
		ModifiedFiles:   3,
		NewFiles:        []string{"new1.txt", "new2.txt"},
		DeletedFiles:    []string{"deleted1.txt"},
		DiffOutput:      "test diff output",
	}

	// Act
	result := writer.formatDiffRecord(record)

	// Assert
	expectedParts := []string{
		config.HeaderGitDiffRecord,
		"Generated at: 2023-01-01 12:00:00",
		"Repository: test-repo",
		"Branch: main",
		"Latest commit: abc12345",
		"Options: --staged-only=true",
		config.HeaderFileChangesSummary,
		"Modified files: 3",
		"New files: 2",
		"Deleted files: 1",
		config.HeaderNewFiles,
		"new1.txt",
		"new2.txt",
		config.HeaderDeletedFiles,
		"deleted1.txt",
		config.HeaderDetailedDiff,
		"test diff output",
	}

	for _, expected := range expectedParts {
		if !strings.Contains(result, expected) {
			t.Errorf("期待される内容が含まれていません: %s", expected)
		}
	}
}

// TestWriter_formatDiffRecord_WithEmptyData は空データのフォーマットテスト
func TestWriter_formatDiffRecord_WithEmptyData(t *testing.T) {
	// Arrange
	writer := NewWriter("/tmp")
	record := &DiffRecord{
		GeneratedAt:     time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		Repository:      "test-repo",
		Branch:          "main",
		LatestCommit:    "abc12345",
		StagedOnly:      false,
		ModifiedFiles:   0,
		NewFiles:        []string{},
		DeletedFiles:    []string{},
		DiffOutput:      "",
	}

	// Act
	result := writer.formatDiffRecord(record)

	// Assert
	// 基本ヘッダーは含まれるべき
	expectedParts := []string{
		config.HeaderGitDiffRecord,
		"Generated at: 2023-01-01 12:00:00",
		"Repository: test-repo",
		"Branch: main",
		"Latest commit: abc12345",
		"Options: --staged-only=false",
		config.HeaderFileChangesSummary,
		"Modified files: 0",
		"New files: 0",
		"Deleted files: 0",
		config.HeaderDetailedDiff,
		"差分はありません。",
	}

	for _, expected := range expectedParts {
		if !strings.Contains(result, expected) {
			t.Errorf("期待される内容が含まれていません: %s", expected)
		}
	}

	// 空のセクションは含まれないべき
	unexpectedParts := []string{
		config.HeaderNewFiles,
		config.HeaderDeletedFiles,
	}

	for _, unexpected := range unexpectedParts {
		if strings.Contains(result, unexpected) {
			t.Errorf("含まれるべきでない内容が含まれています: %s", unexpected)
		}
	}
}

// TestWriter_formatDiffRecord_WithPartialData は部分データのフォーマットテスト
func TestWriter_formatDiffRecord_WithPartialData(t *testing.T) {
	// Arrange
	writer := NewWriter("/tmp")
	record := &DiffRecord{
		GeneratedAt:     time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		Repository:      "test-repo",
		Branch:          "main",
		LatestCommit:    "abc12345",
		StagedOnly:      false,
		ModifiedFiles:   1,
		NewFiles:        []string{"new1.txt"},
		DeletedFiles:    []string{},
		DiffOutput:      "some diff",
	}

	// Act
	result := writer.formatDiffRecord(record)

	// Assert
	// 新規ファイルセクションは含まれるべき
	if !strings.Contains(result, config.HeaderNewFiles) {
		t.Error("新規ファイルセクションが含まれていません")
	}
	if !strings.Contains(result, "new1.txt") {
		t.Error("新規ファイル名が含まれていません")
	}

	// 削除ファイルセクションは含まれないべき
	if strings.Contains(result, config.HeaderDeletedFiles) {
		t.Error("削除ファイルセクションが含まれるべきではありません")
	}

	// 差分は含まれるべき
	if !strings.Contains(result, "some diff") {
		t.Error("差分内容が含まれていません")
	}
}

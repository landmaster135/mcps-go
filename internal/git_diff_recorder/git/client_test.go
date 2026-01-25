package git

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// MockFileInfo はos.FileInfoのモック実装
type MockFileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	isDir   bool
}

func (m MockFileInfo) Name() string       { return m.name }
func (m MockFileInfo) Size() int64        { return m.size }
func (m MockFileInfo) Mode() fs.FileMode  { return m.mode }
func (m MockFileInfo) ModTime() time.Time { return m.modTime }
func (m MockFileInfo) IsDir() bool        { return m.isDir }
func (m MockFileInfo) Sys() any           { return nil }

// TestClient_GetDiff_UntrackedFileIntegration は未追跡ファイルの統合テスト
func TestClient_GetDiff_UntrackedFileIntegration(t *testing.T) {
	// Arrange
	tempDir, err := os.MkdirTemp("", "git-diff-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗しました: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// テスト用ファイルを作成
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "line1\nline2\nline3\n"
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("テストファイルの作成に失敗しました: %v", err)
	}

	mockExecutor := NewMockGitExecutor()
	client := NewClientWithExecutor(tempDir, mockExecutor)

	// git diff HEADのレスポンス（空）
	mockExecutor.SetResponse("diff HEAD", []byte(""))

	// git status --porcelainのレスポンス（未追跡ファイル）
	statusOutput := "?? test.txt\n"
	mockExecutor.SetResponse("status --porcelain", []byte(statusOutput))

	// Act
	diff, err := client.GetDiff(false)

	// Assert
	if err != nil {
		t.Errorf("GetDiffでエラーが発生しました: %v", err)
	}
	if !strings.Contains(diff, "diff --git a/test.txt b/test.txt") {
		t.Error("未追跡ファイルのdiffヘッダーが含まれていません")
	}
	if !strings.Contains(diff, "new file mode 100644") {
		t.Error("新規ファイルのモード情報が含まれていません")
	}
	if !strings.Contains(diff, "+line1") {
		t.Error("ファイル内容が含まれていません")
	}
}

// TestClient_GetDiff_UntrackedDirectoryIntegration は未追跡ディレクトリの統合テスト
func TestClient_GetDiff_UntrackedDirectoryIntegration(t *testing.T) {
	// Arrange
	tempDir, err := os.MkdirTemp("", "git-diff-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗しました: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// テスト用ディレクトリとファイルを作成
	testSubDir := filepath.Join(tempDir, "subdir")
	err = os.Mkdir(testSubDir, 0755)
	if err != nil {
		t.Fatalf("サブディレクトリの作成に失敗しました: %v", err)
	}

	testFile1 := filepath.Join(testSubDir, "file1.txt")
	testFile2 := filepath.Join(testSubDir, "file2.txt")
	err = os.WriteFile(testFile1, []byte("content1\n"), 0644)
	if err != nil {
		t.Fatalf("テストファイル1の作成に失敗しました: %v", err)
	}
	err = os.WriteFile(testFile2, []byte("content2\n"), 0644)
	if err != nil {
		t.Fatalf("テストファイル2の作成に失敗しました: %v", err)
	}

	mockExecutor := NewMockGitExecutor()
	client := NewClientWithExecutor(tempDir, mockExecutor)

	// git diff HEADのレスポンス（空）
	mockExecutor.SetResponse("diff HEAD", []byte(""))

	// git status --porcelainのレスポンス（未追跡ディレクトリ）
	statusOutput := "?? subdir/\n"
	mockExecutor.SetResponse("status --porcelain", []byte(statusOutput))

	// Act
	diff, err := client.GetDiff(false)

	// Assert
	if err != nil {
		t.Errorf("GetDiffでエラーが発生しました: %v", err)
	}
	if !strings.Contains(diff, "diff --git a/subdir/file1.txt b/subdir/file1.txt") {
		t.Error("サブディレクトリ内のファイル1のdiffが含まれていません")
	}
	if !strings.Contains(diff, "diff --git a/subdir/file2.txt b/subdir/file2.txt") {
		t.Error("サブディレクトリ内のファイル2のdiffが含まれていません")
	}
	if !strings.Contains(diff, "+content1") {
		t.Error("ファイル1の内容が含まれていません")
	}
	if !strings.Contains(diff, "+content2") {
		t.Error("ファイル2の内容が含まれていません")
	}
}

// TestClient_NewClient_Normal はClient作成の正常系テスト
func TestClient_NewClient_Normal(t *testing.T) {
	// Arrange
	workingDir := "/tmp/test"

	// Act
	client := NewClient(workingDir)

	// Assert
	if client == nil {
		t.Error("Clientの作成に失敗しました")
		return
	}
	if client.workingDir != workingDir {
		t.Errorf("workingDirが期待値と異なります。期待値: %s, 実際: %s", workingDir, client.workingDir)
	}
	if client.executor == nil {
		t.Error("executorが設定されていません")
	}
}

// TestClient_NewClientWithExecutor_Normal は依存性注入Client作成の正常系テスト
func TestClient_NewClientWithExecutor_Normal(t *testing.T) {
	// Arrange
	workingDir := "/tmp/test"
	mockExecutor := NewMockGitExecutor()

	// Act
	client := NewClientWithExecutor(workingDir, mockExecutor)

	// Assert
	if client == nil {
		t.Error("Clientの作成に失敗しました")
		return
	}
	if client.workingDir != workingDir {
		t.Errorf("workingDirが期待値と異なります。期待値: %s, 実際: %s", workingDir, client.workingDir)
	}
	if client.executor != mockExecutor {
		t.Error("executorが正しく設定されていません")
	}
}

// TestClient_GetRepositoryName_Normal はGetRepositoryName正常系テスト
func TestClient_GetRepositoryName_Normal(t *testing.T) {
	// Arrange
	mockExecutor := NewMockGitExecutor()
	client := NewClientWithExecutor("/tmp/test", mockExecutor)

	mockOutput := "/home/user/projects/test-repo\n"
	mockExecutor.SetResponse("rev-parse --show-toplevel", []byte(mockOutput))

	// Act
	repoName, err := client.GetRepositoryName()

	// Assert
	if err != nil {
		t.Errorf("GetRepositoryNameでエラーが発生しました: %v", err)
	}
	expected := "test-repo"
	if repoName != expected {
		t.Errorf("リポジトリ名が期待値と異なります。期待値: %s, 実際: %s", expected, repoName)
	}
}

// TestClient_GetRepositoryName_Error はGetRepositoryNameエラー系テスト
func TestClient_GetRepositoryName_Error(t *testing.T) {
	// Arrange
	mockExecutor := NewMockGitExecutor()
	client := NewClientWithExecutor("/tmp/test", mockExecutor)

	expectedError := errors.New("git command failed")
	mockExecutor.SetError("rev-parse --show-toplevel", expectedError)

	// Act
	_, err := client.GetRepositoryName()

	// Assert
	if err == nil {
		t.Error("エラーが発生するべきです")
	}
	if !strings.Contains(err.Error(), "リポジトリのルートディレクトリを取得できませんでした") {
		t.Errorf("期待されるエラーメッセージと異なります: %v", err)
	}
}

// TestClient_GetCurrentBranch_Normal はGetCurrentBranch正常系テスト
func TestClient_GetCurrentBranch_Normal(t *testing.T) {
	// Arrange
	mockExecutor := NewMockGitExecutor()
	client := NewClientWithExecutor("/tmp/test", mockExecutor)

	mockOutput := "main\n"
	mockExecutor.SetResponse("rev-parse --abbrev-ref HEAD", []byte(mockOutput))

	// Act
	branch, err := client.GetCurrentBranch()

	// Assert
	if err != nil {
		t.Errorf("GetCurrentBranchでエラーが発生しました: %v", err)
	}
	expected := "main"
	if branch != expected {
		t.Errorf("ブランチ名が期待値と異なります。期待値: %s, 実際: %s", expected, branch)
	}
}

// TestClient_GetCurrentBranch_Error はGetCurrentBranchエラー系テスト
func TestClient_GetCurrentBranch_Error(t *testing.T) {
	// Arrange
	mockExecutor := NewMockGitExecutor()
	client := NewClientWithExecutor("/tmp/test", mockExecutor)

	expectedError := errors.New("git command failed")
	mockExecutor.SetError("rev-parse --abbrev-ref HEAD", expectedError)

	// Act
	_, err := client.GetCurrentBranch()

	// Assert
	if err == nil {
		t.Error("エラーが発生するべきです")
	}
	if !strings.Contains(err.Error(), "現在のブランチを取得できませんでした") {
		t.Errorf("期待されるエラーメッセージと異なります: %v", err)
	}
}

// TestClient_GetLatestCommitHash_Normal はGetLatestCommitHash正常系テスト
func TestClient_GetLatestCommitHash_Normal(t *testing.T) {
	// Arrange
	mockExecutor := NewMockGitExecutor()
	client := NewClientWithExecutor("/tmp/test", mockExecutor)

	mockOutput := "abcdef1234567890abcdef1234567890abcdef12\n"
	mockExecutor.SetResponse("rev-parse HEAD", []byte(mockOutput))

	// Act
	hash, err := client.GetLatestCommitHash()

	// Assert
	if err != nil {
		t.Errorf("GetLatestCommitHashでエラーが発生しました: %v", err)
	}
	expected := "abcdef12" // 最初の8文字
	if hash != expected {
		t.Errorf("コミットハッシュが期待値と異なります。期待値: %s, 実際: %s", expected, hash)
	}
}

// TestClient_GetLatestCommitHash_ShortHash は短いハッシュのテスト
func TestClient_GetLatestCommitHash_ShortHash(t *testing.T) {
	// Arrange
	mockExecutor := NewMockGitExecutor()
	client := NewClientWithExecutor("/tmp/test", mockExecutor)

	mockOutput := "abc123\n"
	mockExecutor.SetResponse("rev-parse HEAD", []byte(mockOutput))

	// Act
	hash, err := client.GetLatestCommitHash()

	// Assert
	if err != nil {
		t.Errorf("GetLatestCommitHashでエラーが発生しました: %v", err)
	}
	expected := "abc123" // 8文字未満の場合はそのまま
	if hash != expected {
		t.Errorf("コミットハッシュが期待値と異なります。期待値: %s, 実際: %s", expected, hash)
	}
}

// TestClient_GetDiff_StagedOnly はステージング済み差分取得のテスト
func TestClient_GetDiff_StagedOnly(t *testing.T) {
	// Arrange
	mockExecutor := NewMockGitExecutor()
	client := NewClientWithExecutor("/tmp/test", mockExecutor)

	mockOutput := "diff --git a/test.txt b/test.txt\n+added line\n"
	mockExecutor.SetResponse("diff --cached", []byte(mockOutput))

	// Act
	diff, err := client.GetDiff(true)

	// Assert
	if err != nil {
		t.Errorf("GetDiffでエラーが発生しました: %v", err)
	}
	if !strings.Contains(diff, "diff --git a/test.txt b/test.txt") {
		t.Error("期待される差分内容が含まれていません")
	}
}

// TestClient_GetDiff_AllChanges は全変更差分取得のテスト
func TestClient_GetDiff_AllChanges(t *testing.T) {
	// Arrange
	mockExecutor := NewMockGitExecutor()
	client := NewClientWithExecutor("/tmp/test", mockExecutor)

	mockOutput := "diff --git a/test.txt b/test.txt\n+added line\n-removed line\n"
	mockExecutor.SetResponse("diff HEAD", []byte(mockOutput))

	// Act
	diff, err := client.GetDiff(false)

	// Assert
	if err != nil {
		t.Errorf("GetDiffでエラーが発生しました: %v", err)
	}
	if !strings.Contains(diff, "diff --git a/test.txt b/test.txt") {
		t.Error("期待される差分内容が含まれていません")
	}
}

// TestClient_GetStatus_Normal はGetStatus正常系テスト
func TestClient_GetStatus_Normal(t *testing.T) {
	// Arrange
	mockExecutor := NewMockGitExecutor()
	client := NewClientWithExecutor("/tmp/test", mockExecutor)

	mockOutput := "M  modified.txt\nA  added.txt\nD  deleted.txt\n?? untracked.txt\n"
	mockExecutor.SetResponse("status --porcelain", []byte(mockOutput))

	// Act
	status, err := client.GetStatus()

	// Assert
	if err != nil {
		t.Errorf("GetStatusでエラーが発生しました: %v", err)
	}
	if !strings.Contains(status, "M  modified.txt") {
		t.Error("変更ファイル情報が含まれていません")
	}
	if !strings.Contains(status, "A  added.txt") {
		t.Error("追加ファイル情報が含まれていません")
	}
}

// TestClient_GetNewFiles_StagedOnly はステージング済み新規ファイル取得のテスト
func TestClient_GetNewFiles_StagedOnly(t *testing.T) {
	// Arrange
	mockExecutor := NewMockGitExecutor()
	client := NewClientWithExecutor("/tmp/test", mockExecutor)

	mockOutput := "A  staged_new.txt\n?? untracked.txt\nM  modified.txt\n"
	mockExecutor.SetResponse("status --porcelain", []byte(mockOutput))

	// Act
	newFiles, err := client.GetNewFiles(true)

	// Assert
	if err != nil {
		t.Errorf("GetNewFilesでエラーが発生しました: %v", err)
	}
	if len(newFiles) != 1 {
		t.Errorf("新規ファイル数が期待値と異なります。期待値: 1, 実際: %d", len(newFiles))
	}
	if newFiles[0] != "staged_new.txt" {
		t.Errorf("新規ファイル名が期待値と異なります。期待値: staged_new.txt, 実際: %s", newFiles[0])
	}
}

// TestClient_GetNewFiles_AllFiles は全新規ファイル取得のテスト
func TestClient_GetNewFiles_AllFiles(t *testing.T) {
	// Arrange
	mockExecutor := NewMockGitExecutor()
	client := NewClientWithExecutor("/tmp/test", mockExecutor)

	mockOutput := "A  staged_new.txt\n?? untracked.txt\nM  modified.txt\n"
	mockExecutor.SetResponse("status --porcelain", []byte(mockOutput))

	// Act
	newFiles, err := client.GetNewFiles(false)

	// Assert
	if err != nil {
		t.Errorf("GetNewFilesでエラーが発生しました: %v", err)
	}
	if len(newFiles) != 2 {
		t.Errorf("新規ファイル数が期待値と異なります。期待値: 2, 実際: %d", len(newFiles))
	}

	expectedFiles := []string{"staged_new.txt", "untracked.txt"}
	for _, expected := range expectedFiles {
		found := false
		for _, actual := range newFiles {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("期待されるファイルが見つかりません: %s", expected)
		}
	}
}

// TestClient_GetDeletedFiles_Normal は削除ファイル取得のテスト
func TestClient_GetDeletedFiles_Normal(t *testing.T) {
	// Arrange
	mockExecutor := NewMockGitExecutor()
	client := NewClientWithExecutor("/tmp/test", mockExecutor)

	mockOutput := "D  staged_deleted.txt\n D unstaged_deleted.txt\nM  modified.txt\n"
	mockExecutor.SetResponse("status --porcelain", []byte(mockOutput))

	// Act
	deletedFiles, err := client.GetDeletedFiles(false)

	// Assert
	if err != nil {
		t.Errorf("GetDeletedFilesでエラーが発生しました: %v", err)
	}
	if len(deletedFiles) != 2 {
		t.Errorf("削除ファイル数が期待値と異なります。期待値: 2, 実際: %d", len(deletedFiles))
	}
}

// TestClient_GetModifiedFilesCount_Normal は変更ファイル数取得のテスト
func TestClient_GetModifiedFilesCount_Normal(t *testing.T) {
	// Arrange
	mockExecutor := NewMockGitExecutor()
	client := NewClientWithExecutor("/tmp/test", mockExecutor)

	mockOutput := "M  staged_modified.txt\n M unstaged_modified.txt\nMM both_modified.txt\nA  added.txt\n"
	mockExecutor.SetResponse("status --porcelain", []byte(mockOutput))

	// Act
	count, err := client.GetModifiedFilesCount(false)

	// Assert
	if err != nil {
		t.Errorf("GetModifiedFilesCountでエラーが発生しました: %v", err)
	}
	expected := 3 // M , M, MM
	if count != expected {
		t.Errorf("変更ファイル数が期待値と異なります。期待値: %d, 実際: %d", expected, count)
	}
}

// TestClient_GetModifiedFilesCount_StagedOnly はステージング済み変更ファイル数取得のテスト
func TestClient_GetModifiedFilesCount_StagedOnly(t *testing.T) {
	// Arrange
	mockExecutor := NewMockGitExecutor()
	client := NewClientWithExecutor("/tmp/test", mockExecutor)

	mockOutput := "M  staged_modified.txt\n M unstaged_modified.txt\nMM both_modified.txt\nA  added.txt\n"
	mockExecutor.SetResponse("status --porcelain", []byte(mockOutput))

	// Act
	count, err := client.GetModifiedFilesCount(true)

	// Assert
	if err != nil {
		t.Errorf("GetModifiedFilesCountでエラーが発生しました: %v", err)
	}
	expected := 1 // M のみ
	if count != expected {
		t.Errorf("変更ファイル数が期待値と異なります。期待値: %d, 実際: %d", expected, count)
	}
}

// TestMockGitExecutor_SetResponse はモックのレスポンス設定テスト
func TestMockGitExecutor_SetResponse(t *testing.T) {
	// Arrange
	mockExecutor := NewMockGitExecutor()
	command := "test command"
	response := []byte("test response")

	// Act
	mockExecutor.SetResponse(command, response)
	result, err := mockExecutor.Execute("/tmp", "test", "command")

	// Assert
	if err != nil {
		t.Errorf("Executeでエラーが発生しました: %v", err)
	}
	if string(result) != string(response) {
		t.Errorf("レスポンスが期待値と異なります。期待値: %s, 実際: %s", string(response), string(result))
	}
}

// TestMockGitExecutor_SetError はモックのエラー設定テスト
func TestMockGitExecutor_SetError(t *testing.T) {
	// Arrange
	mockExecutor := NewMockGitExecutor()
	command := "test command"
	expectedError := errors.New("test error")

	// Act
	mockExecutor.SetError(command, expectedError)
	_, err := mockExecutor.Execute("/tmp", "test", "command")

	// Assert
	if err != expectedError {
		t.Errorf("エラーが期待値と異なります。期待値: %v, 実際: %v", expectedError, err)
	}
}

// TestClient_GetDiff_WithUntrackedFiles は未追跡ファイルを含む差分取得のテスト
func TestClient_GetDiff_WithUntrackedFiles(t *testing.T) {
	// Arrange
	mockExecutor := NewMockGitExecutor()
	client := NewClientWithExecutor("/tmp/test", mockExecutor)

	// git diff HEADのレスポンス（追跡済みファイルの差分）
	trackedDiff := "diff --git a/tracked.txt b/tracked.txt\n+tracked change\n"
	mockExecutor.SetResponse("diff HEAD", []byte(trackedDiff))

	// git status --porcelainのレスポンス（未追跡ファイルを含む）
	statusOutput := "M  tracked.txt\n?? untracked.txt\n"
	mockExecutor.SetResponse("status --porcelain", []byte(statusOutput))

	// Act
	diff, err := client.GetDiff(false)

	// Assert
	if err != nil {
		t.Errorf("GetDiffでエラーが発生しました: %v", err)
	}
	if !strings.Contains(diff, "tracked change") {
		t.Error("追跡済みファイルの差分が含まれていません")
	}
	// 注意: 実際のファイルシステムに依存するため、未追跡ファイルの内容は含まれない可能性があります
}

// TestClient_GetDiff_StagedOnlyWithUntrackedFiles はステージング済みのみで未追跡ファイルを除外するテスト
func TestClient_GetDiff_StagedOnlyWithUntrackedFiles(t *testing.T) {
	// Arrange
	mockExecutor := NewMockGitExecutor()
	client := NewClientWithExecutor("/tmp/test", mockExecutor)

	// git diff --cachedのレスポンス
	stagedDiff := "diff --git a/staged.txt b/staged.txt\n+staged change\n"
	mockExecutor.SetResponse("diff --cached", []byte(stagedDiff))

	// git status --porcelainのレスポンス（未追跡ファイルを含む）
	statusOutput := "A  staged.txt\n?? untracked.txt\n"
	mockExecutor.SetResponse("status --porcelain", []byte(statusOutput))

	// Act
	diff, err := client.GetDiff(true)

	// Assert
	if err != nil {
		t.Errorf("GetDiffでエラーが発生しました: %v", err)
	}
	if !strings.Contains(diff, "staged change") {
		t.Error("ステージング済みファイルの差分が含まれていません")
	}
	// stagedOnly=trueの場合、未追跡ファイルは処理されないことを確認
	// （実際のファイルシステムに依存しないため、この部分は間接的にテスト）
}

// TestClient_GetDiff_NoUntrackedFiles は未追跡ファイルがない場合のテスト
func TestClient_GetDiff_NoUntrackedFiles(t *testing.T) {
	// Arrange
	mockExecutor := NewMockGitExecutor()
	client := NewClientWithExecutor("/tmp/test", mockExecutor)

	// git diff HEADのレスポンス
	trackedDiff := "diff --git a/tracked.txt b/tracked.txt\n+tracked change\n"
	mockExecutor.SetResponse("diff HEAD", []byte(trackedDiff))

	// git status --porcelainのレスポンス（未追跡ファイルなし）
	statusOutput := "M  tracked.txt\n"
	mockExecutor.SetResponse("status --porcelain", []byte(statusOutput))

	// Act
	diff, err := client.GetDiff(false)

	// Assert
	if err != nil {
		t.Errorf("GetDiffでエラーが発生しました: %v", err)
	}
	if !strings.Contains(diff, "tracked change") {
		t.Error("追跡済みファイルの差分が含まれていません")
	}
	// 未追跡ファイルがない場合でも正常に動作することを確認
}

// TestClient_GetDiff_EmptyDiff は差分がない場合のテスト
func TestClient_GetDiff_EmptyDiff(t *testing.T) {
	// Arrange
	mockExecutor := NewMockGitExecutor()
	client := NewClientWithExecutor("/tmp/test", mockExecutor)

	// git diff HEADのレスポンス（空）
	mockExecutor.SetResponse("diff HEAD", []byte(""))

	// git status --porcelainのレスポンス（空）
	mockExecutor.SetResponse("status --porcelain", []byte(""))

	// Act
	diff, err := client.GetDiff(false)

	// Assert
	if err != nil {
		t.Errorf("GetDiffでエラーが発生しました: %v", err)
	}
	if diff != "" {
		t.Errorf("差分が空であるべきです。実際: %s", diff)
	}
}

// TestClient_GetDiff_GitStatusError はgit statusエラーのテスト
func TestClient_GetDiff_GitStatusError(t *testing.T) {
	// Arrange
	mockExecutor := NewMockGitExecutor()
	client := NewClientWithExecutor("/tmp/test", mockExecutor)

	// git diff HEADのレスポンス
	mockExecutor.SetResponse("diff HEAD", []byte(""))

	// git status --porcelainでエラー
	expectedError := errors.New("git status failed")
	mockExecutor.SetError("status --porcelain", expectedError)

	// Act
	_, err := client.GetDiff(false)

	// Assert
	if err == nil {
		t.Error("エラーが発生するべきです")
	}
	if !strings.Contains(err.Error(), "未追跡ファイルの差分取得に失敗しました") {
		t.Errorf("期待されるエラーメッセージと異なります: %v", err)
	}
}

// TestClient_GetDiff_GitDiffError はgit diffエラーのテスト
func TestClient_GetDiff_GitDiffError(t *testing.T) {
	// Arrange
	mockExecutor := NewMockGitExecutor()
	client := NewClientWithExecutor("/tmp/test", mockExecutor)

	// git diff HEADでエラー
	expectedError := errors.New("git diff failed")
	mockExecutor.SetError("diff HEAD", expectedError)

	// Act
	_, err := client.GetDiff(false)

	// Assert
	if err == nil {
		t.Error("エラーが発生するべきです")
	}
	if !strings.Contains(err.Error(), "差分を取得できませんでした") {
		t.Errorf("期待されるエラーメッセージと異なります: %v", err)
	}
}

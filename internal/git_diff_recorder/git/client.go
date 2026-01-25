package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Client はGit操作を行うクライアント
type Client struct {
	workingDir string
	executor   GitCommandExecutor
}

// NewClient は新しいGitクライアントを作成する
func NewClient(workingDir string) *Client {
	return &Client{
		workingDir: workingDir,
		executor:   NewStandardGitExecutor(),
	}
}

// NewClientWithExecutor は依存性注入可能なGitクライアントを作成する
func NewClientWithExecutor(workingDir string, executor GitCommandExecutor) *Client {
	return &Client{
		workingDir: workingDir,
		executor:   executor,
	}
}

// GetRepositoryName はリポジトリ名を取得する
func (c *Client) GetRepositoryName() (string, error) {
	output, err := c.executor.Execute(c.workingDir, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("リポジトリのルートディレクトリを取得できませんでした: %w", err)
	}

	repoPath := strings.TrimSpace(string(output))
	repoName := filepath.Base(repoPath)

	return repoName, nil
}

// GetCurrentBranch は現在のブランチ名を取得する
func (c *Client) GetCurrentBranch() (string, error) {
	output, err := c.executor.Execute(c.workingDir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("現在のブランチを取得できませんでした: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// GetLatestCommitHash は最新のコミットハッシュを取得する
func (c *Client) GetLatestCommitHash() (string, error) {
	output, err := c.executor.Execute(c.workingDir, "rev-parse", "HEAD")
	if err != nil {
		return "", fmt.Errorf("最新のコミットハッシュを取得できませんでした: %w", err)
	}

	hash := strings.TrimSpace(string(output))
	// 短縮形で返す（最初の8文字）
	if len(hash) > 8 {
		hash = hash[:8]
	}

	return hash, nil
}

// getUntrackedFiles は未追跡ファイルのリストを取得する
func (c *Client) getUntrackedFiles() ([]string, error) {
	status, err := c.GetStatus()
	if err != nil {
		return nil, err
	}

	var untrackedFiles []string
	lines := strings.Split(status, "\n")

	for _, line := range lines {
		if len(line) >= 3 {
			statusCode := line[:2]
			filename := strings.TrimSpace(line[3:])

			// 未追跡ファイル（??）のみを対象
			if statusCode == "??" {
				untrackedFiles = append(untrackedFiles, filename)
			}
		}
	}

	return untrackedFiles, nil
}

// processUntrackedFile は単一の未追跡ファイルを処理する
func (c *Client) processUntrackedFile(fullPath, relativePath string, diffOutput *strings.Builder) error {
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return err
	}

	// diff形式で出力
	diffOutput.WriteString(fmt.Sprintf("diff --git a/%s b/%s\n", relativePath, relativePath))
	diffOutput.WriteString("new file mode 100644\n")
	diffOutput.WriteString("index 0000000..0000000\n")
	diffOutput.WriteString("--- /dev/null\n")
	diffOutput.WriteString(fmt.Sprintf("+++ b/%s\n", relativePath))

	// ファイル内容を+行として追加
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		diffOutput.WriteString(fmt.Sprintf("+%s\n", line))
	}

	return nil
}

// processUntrackedDirectory は未追跡ディレクトリを再帰的に処理する
func (c *Client) processUntrackedDirectory(dirPath, relativeDirPath string, diffOutput *strings.Builder) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// ディレクトリはスキップ
		if info.IsDir() {
			return nil
		}

		// 相対パスを計算
		relPath, err := filepath.Rel(c.workingDir, path)
		if err != nil {
			return err
		}

		return c.processUntrackedFile(path, relPath, diffOutput)
	})
}

// getUntrackedFilesDiff は未追跡ファイルの差分を取得する
func (c *Client) getUntrackedFilesDiff() (string, error) {
	// 未追跡ファイルのリストを取得
	untrackedFiles, err := c.getUntrackedFiles()
	if err != nil {
		return "", err
	}

	if len(untrackedFiles) == 0 {
		return "", nil
	}

	var diffOutput strings.Builder

	for _, filename := range untrackedFiles {
		fullPath := filepath.Join(c.workingDir, filename)

		// ファイルかディレクトリかを確認
		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}

		if info.IsDir() {
			// ディレクトリの場合は再帰的に処理
			err := c.processUntrackedDirectory(fullPath, filename, &diffOutput)
			if err != nil {
				continue
			}
		} else {
			// ファイルの場合は直接処理
			err := c.processUntrackedFile(fullPath, filename, &diffOutput)
			if err != nil {
				continue
			}
		}
	}

	return diffOutput.String(), nil
}

// GetDiff は差分を取得する
func (c *Client) GetDiff(stagedOnly bool) (string, error) {
	var diffOutput strings.Builder
	var output []byte
	var err error

	if stagedOnly {
		// ステージング済みの差分のみ
		output, err = c.executor.Execute(c.workingDir, "diff", "--cached")
	} else {
		// 全ての差分（ステージング済み + 未ステージング）
		output, err = c.executor.Execute(c.workingDir, "diff", "HEAD")
	}

	if err != nil {
		return "", fmt.Errorf("差分を取得できませんでした: %w", err)
	}

	// 追跡済みファイルの差分を追加
	if len(output) > 0 {
		diffOutput.WriteString(string(output))
	}

	// 未追跡ファイルの差分も含める（stagedOnlyがfalseの場合のみ）
	if !stagedOnly {
		untrackedDiff, err := c.getUntrackedFilesDiff()
		if err != nil {
			return "", fmt.Errorf("未追跡ファイルの差分取得に失敗しました: %w", err)
		}
		if untrackedDiff != "" {
			if diffOutput.Len() > 0 {
				diffOutput.WriteString("\n")
			}
			diffOutput.WriteString(untrackedDiff)
		}
	}

	return diffOutput.String(), nil
}

// GetStatus はgit statusの情報を取得する
func (c *Client) GetStatus() (string, error) {
	output, err := c.executor.Execute(c.workingDir, "status", "--porcelain")
	if err != nil {
		return "", fmt.Errorf("ステータスを取得できませんでした: %w", err)
	}

	return string(output), nil
}

// GetNewFiles は新規ファイルのリストを取得する
func (c *Client) GetNewFiles(stagedOnly bool) ([]string, error) {
	status, err := c.GetStatus()
	if err != nil {
		return nil, err
	}

	var newFiles []string
	lines := strings.Split(status, "\n")

	for _, line := range lines {
		if len(line) >= 3 {
			statusCode := line[:2]
			filename := strings.TrimSpace(line[3:])

			if stagedOnly {
				// ステージング済み新規ファイルのみ（A: added）
				if statusCode == "A " {
					newFiles = append(newFiles, filename)
				}
			} else {
				// 全ての新規ファイル（A: added, ??: untracked）
				if statusCode == "A " || statusCode == "??" {
					newFiles = append(newFiles, filename)
				}
			}
		}
	}

	return newFiles, nil
}

// GetDeletedFiles は削除されたファイルのリストを取得する
func (c *Client) GetDeletedFiles(stagedOnly bool) ([]string, error) {
	status, err := c.GetStatus()
	if err != nil {
		return nil, err
	}

	var deletedFiles []string
	lines := strings.Split(status, "\n")

	for _, line := range lines {
		if len(line) >= 3 {
			statusCode := line[:2]
			filename := strings.TrimSpace(line[3:])

			if stagedOnly {
				// ステージング済み削除ファイルのみ（D: deleted）
				if statusCode == "D " {
					deletedFiles = append(deletedFiles, filename)
				}
			} else {
				// 全ての削除ファイル（D: deleted）
				if statusCode == "D " || statusCode == " D" {
					deletedFiles = append(deletedFiles, filename)
				}
			}
		}
	}

	return deletedFiles, nil
}

// GetModifiedFilesCount は変更されたファイル数を取得する
func (c *Client) GetModifiedFilesCount(stagedOnly bool) (int, error) {
	status, err := c.GetStatus()
	if err != nil {
		return 0, err
	}

	count := 0
	lines := strings.Split(status, "\n")

	for _, line := range lines {
		if len(line) >= 3 {
			statusCode := line[:2]

			if stagedOnly {
				// ステージング済み変更ファイルのみ（M: modified）
				if statusCode == "M " {
					count++
				}
			} else {
				// 全ての変更ファイル（M: modified）
				if statusCode == "M " || statusCode == " M" || statusCode == "MM" {
					count++
				}
			}
		}
	}

	return count, nil
}

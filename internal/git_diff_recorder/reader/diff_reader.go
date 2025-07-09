package reader

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/landmaster135/mcps-go/internal/git_diff_recorder/config"
)

// DiffReader はdiffファイルの読み取りを行う構造体
type DiffReader struct {
	sourceDir string
}

// NewDiffReader は新しいDiffReaderを作成する
func NewDiffReader(sourceDir string) *DiffReader {
	return &DiffReader{
		sourceDir: sourceDir,
	}
}

// FindLatestDiffFile は指定リポジトリの最新のdiffファイルを検索する
func (r *DiffReader) FindLatestDiffFile(repository string) (string, error) {
	repoDir := filepath.Join(r.sourceDir, repository)

	// リポジトリディレクトリが存在するかチェック
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		return "", fmt.Errorf("リポジトリディレクトリが見つかりません: %s", repoDir)
	}

	// diffファイルを検索
	files, err := filepath.Glob(filepath.Join(repoDir, "diff_*.txt"))
	if err != nil {
		return "", fmt.Errorf("diffファイルの検索に失敗しました: %w", err)
	}

	if len(files) == 0 {
		return "", fmt.Errorf("diffファイルが見つかりません: %s", repoDir)
	}

	// ファイル名でソートして最新のファイルを取得
	sort.Strings(files)
	latestFile := files[len(files)-1]

	return latestFile, nil
}

// ExtractDetailedDiff はファイルから「=== Detailed Diff ===」セクションを抽出する
func (r *DiffReader) ExtractDetailedDiff(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("ファイルを開けませんでした: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var result strings.Builder
	inDetailedDiffSection := false

	for scanner.Scan() {
		line := scanner.Text()

		// 「=== Detailed Diff ===」セクションの開始を検出
		if line == config.HeaderDetailedDiff {
			inDetailedDiffSection = true
			continue
		}

		// 他のセクションの開始を検出した場合、終了
		if inDetailedDiffSection && strings.HasPrefix(line, "===") && strings.HasSuffix(line, "===") {
			break
		}

		// 「=== Detailed Diff ===」セクション内の内容を収集
		if inDetailedDiffSection {
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("ファイルの読み取りに失敗しました: %w", err)
	}

	if !inDetailedDiffSection {
		return "", fmt.Errorf("「%s」セクションが見つかりませんでした", config.HeaderDetailedDiff)
	}

	content := result.String()
	// 末尾の改行を削除
	content = strings.TrimSuffix(content, "\n")

	return content, nil
}

// DiffFileInfo はdiffファイルの情報を保持する構造体
type DiffFileInfo struct {
	FilePath    string
	FileName    string
	ModTime     time.Time
	GeneratedAt time.Time
	Repository  string
	Branch      string
	Options     string
}


// GetFileInfo はdiffファイルの基本情報を取得する
func (r *DiffReader) GetFileInfo(filePath string) (*DiffFileInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("ファイルを開けませんでした: %w", err)
	}
	defer file.Close()

	info := &DiffFileInfo{
		FilePath: filePath,
		FileName: filepath.Base(filePath),
	}

	// ファイルの更新時刻を取得
	if stat, err := file.Stat(); err == nil {
		info.ModTime = stat.ModTime()
	}

	scanner := bufio.NewScanner(file)
	lineCount := 0

	for scanner.Scan() {
		lineCount++
		line := scanner.Text()

		// 基本情報を抽出
		if lineCount <= 10 { // 最初の10行程度から情報を抽出
			if strings.HasPrefix(line, "Generated at:") {
				if timeStr := strings.TrimPrefix(line, "Generated at: "); timeStr != "" {
					if t, err := time.Parse("2006-01-02 15:04:05", timeStr); err == nil {
						info.GeneratedAt = t
					}
				}
			} else if strings.HasPrefix(line, "Repository:") {
				info.Repository = strings.TrimSpace(strings.TrimPrefix(line, "Repository:"))
			} else if strings.HasPrefix(line, "Branch:") {
				info.Branch = strings.TrimSpace(strings.TrimPrefix(line, "Branch:"))
			} else if strings.HasPrefix(line, "Options:") {
				info.Options = strings.TrimSpace(strings.TrimPrefix(line, "Options:"))
			}
		}

		if line == config.HeaderDetailedDiff {
			break
		}
	}

	return info, nil
}

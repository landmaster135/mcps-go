package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/landmaster135/mcps-go/internal/git_diff_recorder/config"
)

// Writer はファイル出力を行う構造体
type Writer struct {
	outputDir string
}

// NewWriter は新しいWriterを作成する
func NewWriter(outputDir string) *Writer {
	return &Writer{
		outputDir: outputDir,
	}
}

// WriteDiffRecord は差分記録をファイルに出力する
func (w *Writer) WriteDiffRecord(repoName string, record *DiffRecord) error {
	// 出力ディレクトリを作成
	repoDir := filepath.Join(w.outputDir, repoName)
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		return fmt.Errorf("ディレクトリの作成に失敗しました: %w", err)
	}

	// ファイル名を生成（diff_yyyyMMddhhmmss.txt）
	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("diff_%s.txt", timestamp)
	filepath := filepath.Join(repoDir, filename)

	// ファイルを作成
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("ファイルの作成に失敗しました: %w", err)
	}
	defer file.Close()

	// 内容を書き込み
	content := w.formatDiffRecord(record)
	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("ファイルの書き込みに失敗しました: %w", err)
	}

	return nil
}

// DiffRecord は差分記録の構造体
type DiffRecord struct {
	GeneratedAt     time.Time
	Repository      string
	Branch          string
	LatestCommit    string
	StagedOnly      bool
	ModifiedFiles   int
	NewFiles        []string
	DeletedFiles    []string
	DiffOutput      string
}

// formatDiffRecord は差分記録をフォーマットする
func (w *Writer) formatDiffRecord(record *DiffRecord) string {
	var builder strings.Builder

	// ヘッダー
	builder.WriteString(config.HeaderGitDiffRecord + "\n")
	builder.WriteString(fmt.Sprintf("Generated at: %s\n", record.GeneratedAt.Format("2006-01-02 15:04:05")))
	builder.WriteString(fmt.Sprintf("Repository: %s\n", record.Repository))
	builder.WriteString(fmt.Sprintf("Branch: %s\n", record.Branch))
	builder.WriteString(fmt.Sprintf("Latest commit: %s\n", record.LatestCommit))
	builder.WriteString(fmt.Sprintf("Options: --staged-only=%t\n", record.StagedOnly))
	builder.WriteString("\n")

	// ファイル変更サマリー
	builder.WriteString(config.HeaderFileChangesSummary + "\n")
	builder.WriteString(fmt.Sprintf("Modified files: %d\n", record.ModifiedFiles))
	builder.WriteString(fmt.Sprintf("New files: %d\n", len(record.NewFiles)))
	builder.WriteString(fmt.Sprintf("Deleted files: %d\n", len(record.DeletedFiles)))
	builder.WriteString("\n")

	// 新規ファイル一覧
	if len(record.NewFiles) > 0 {
		builder.WriteString(config.HeaderNewFiles + "\n")
		for _, file := range record.NewFiles {
			builder.WriteString(fmt.Sprintf("%s\n", file))
		}
		builder.WriteString("\n")
	}

	// 削除ファイル一覧
	if len(record.DeletedFiles) > 0 {
		builder.WriteString(config.HeaderDeletedFiles + "\n")
		for _, file := range record.DeletedFiles {
			builder.WriteString(fmt.Sprintf("%s\n", file))
		}
		builder.WriteString("\n")
	}

	// 詳細な差分
	builder.WriteString(config.HeaderDetailedDiff + "\n")
	if record.DiffOutput != "" {
		builder.WriteString(record.DiffOutput)
	} else {
		builder.WriteString("差分はありません。\n")
	}

	return builder.String()
}

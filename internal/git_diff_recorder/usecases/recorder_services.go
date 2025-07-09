package usecases

import (
	"fmt"
	"time"

	"github.com/landmaster135/mcps-go/internal/git_diff_recorder/config"
	"github.com/landmaster135/mcps-go/internal/git_diff_recorder/file"
	"github.com/landmaster135/mcps-go/internal/git_diff_recorder/git"
)

// GitDiffRecorderService はgit差分記録サービス
type GitDiffRecorderService struct {
	gitClient  *git.Client
	fileWriter *file.Writer
	config     *config.Config
}

// NewGitDiffRecorderService は新しいサービスインスタンスを作成する
func NewGitDiffRecorderService(workingDir string, cfg *config.Config) *GitDiffRecorderService {
	return &GitDiffRecorderService{
		gitClient:  git.NewClient(workingDir),
		fileWriter: file.NewWriter(cfg.OutputDir),
		config:     cfg,
	}
}

// RecordDiff はgit差分を記録する
func (s *GitDiffRecorderService) RecordDiff() error {
	client := s.gitClient
	isStagedOnly := s.config.StagedOnly

	// リポジトリ情報を取得
	repoName, err := client.GetRepositoryName()
	if err != nil {
		return fmt.Errorf("リポジトリ名の取得に失敗しました: %w", err)
	}

	branch, err := client.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("ブランチ名の取得に失敗しました: %w", err)
	}

	commitHash, err := client.GetLatestCommitHash()
	if err != nil {
		return fmt.Errorf("コミットハッシュの取得に失敗しました: %w", err)
	}

	// 差分を取得
	diffOutput, err := client.GetDiff(isStagedOnly)
	if err != nil {
		return fmt.Errorf("差分の取得に失敗しました: %w", err)
	}

	// ファイル変更情報を取得
	modifiedCount, err := client.GetModifiedFilesCount(isStagedOnly)
	if err != nil {
		return fmt.Errorf("変更ファイル数の取得に失敗しました: %w", err)
	}

	newFiles, err := client.GetNewFiles(isStagedOnly)
	if err != nil {
		return fmt.Errorf("新規ファイル一覧の取得に失敗しました: %w", err)
	}

	deletedFiles, err := client.GetDeletedFiles(isStagedOnly)
	if err != nil {
		return fmt.Errorf("削除ファイル一覧の取得に失敗しました: %w", err)
	}

	// 差分記録を作成
	record := &file.DiffRecord{
		GeneratedAt:   time.Now(),
		Repository:    repoName,
		Branch:        branch,
		LatestCommit:  commitHash,
		StagedOnly:    s.config.StagedOnly,
		ModifiedFiles: modifiedCount,
		NewFiles:      newFiles,
		DeletedFiles:  deletedFiles,
		DiffOutput:    diffOutput,
	}

	// ファイルに出力
	if err := s.fileWriter.WriteDiffRecord(repoName, record); err != nil {
		return fmt.Errorf("差分記録の出力に失敗しました: %w", err)
	}

	fmt.Printf("差分記録が正常に出力されました: %s/%s/diff_%s.txt\n",
		s.config.OutputDir,
		repoName,
		time.Now().Format("20060102150405"))

	return nil
}

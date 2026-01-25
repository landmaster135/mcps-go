package usecases

import (
	"fmt"

	"github.com/landmaster135/mcps-go/internal/git_diff_recorder/config"
	"github.com/landmaster135/mcps-go/internal/git_diff_recorder/git"
)

// GitDiffGeneratorService は指定されたGitディレクトリから直接差分を生成するサービス
type GitDiffGeneratorService struct {
	gitClient *git.Client
	config    *config.Config
}

// NewGitDiffGeneratorService は新しいGitDiffGeneratorServiceを作成する
func NewGitDiffGeneratorService(gitDir string, cfg *config.Config) *GitDiffGeneratorService {
	return &GitDiffGeneratorService{
		gitClient: git.NewClient(gitDir),
		config:    cfg,
	}
}

// GetCurrentDetailedDiff は指定されたGitディレクトリから詳細差分のみを取得して返す
func (s *GitDiffGeneratorService) GetCurrentDetailedDiff() (string, error) {
	// 指定されたディレクトリのgit差分を取得
	diffOutput, err := s.gitClient.GetDiff(s.config.StagedOnly)
	if err != nil {
		return "", fmt.Errorf("差分の取得に失敗しました: %w", err)
	}

	return diffOutput, nil
}

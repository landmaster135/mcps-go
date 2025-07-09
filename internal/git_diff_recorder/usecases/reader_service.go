package usecases

import (
	"fmt"

	"github.com/landmaster135/mcps-go/internal/git_diff_recorder/config"
	"github.com/landmaster135/mcps-go/internal/git_diff_recorder/reader"
)

// DiffReaderService はdiff読み取りサービス
type DiffReaderService struct {
	diffReader *reader.DiffReader
	config     *config.Config
}

// NewDiffReaderService は新しいDiffReaderServiceを作成する
func NewDiffReaderService(cfg *config.Config) *DiffReaderService {
	return &DiffReaderService{
		diffReader: reader.NewDiffReader(cfg.SourceDir),
		config:     cfg,
	}
}

// displayResult は結果をコマンドラインに表示する
func (s *DiffReaderService) displayResult(fileInfo *reader.DiffFileInfo, detailedDiff string) {
	fmt.Printf("=== 読み取り結果 ===\n")
	fmt.Printf("ファイル: %s\n", fileInfo.FileName)
	fmt.Printf("リポジトリ: %s\n", fileInfo.Repository)
	fmt.Printf("ブランチ: %s\n", fileInfo.Branch)
	if !fileInfo.GeneratedAt.IsZero() {
		fmt.Printf("生成日時: %s\n", fileInfo.GeneratedAt.Format("2006-01-02 15:04:05"))
	}
	fmt.Printf("オプション: %s\n", fileInfo.Options)
	fmt.Printf("\n")

	fmt.Printf("%s\n", config.HeaderDetailedDiff)
	if detailedDiff == "" || detailedDiff == "差分はありません。" {
		fmt.Printf("差分はありません。\n")
	} else {
		fmt.Printf("%s\n", detailedDiff)
	}
}


// GetDetailedDiff は最新のdiffファイルから詳細差分のみを取得して返す
func (s *DiffReaderService) GetDetailedDiff() (string, error) {
	// 最新のdiffファイルを検索
	latestFile, err := s.diffReader.FindLatestDiffFile(s.config.Repository)
	if err != nil {
		return "", fmt.Errorf("最新のdiffファイルの検索に失敗しました: %w", err)
	}

	// 詳細差分を抽出
	detailedDiff, err := s.diffReader.ExtractDetailedDiff(latestFile)
	if err != nil {
		return "", fmt.Errorf("詳細差分の抽出に失敗しました: %w", err)
	}

	return detailedDiff, nil
}

// ReadAndDisplayDetailedDiff は最新のdiffファイルから詳細差分を読み取り表示する
func (s *DiffReaderService) ReadAndDisplayDetailedDiff() error {
	// 最新のdiffファイルを検索
	latestFile, err := s.diffReader.FindLatestDiffFile(s.config.Repository)
	if err != nil {
		return fmt.Errorf("最新のdiffファイルの検索に失敗しました: %w", err)
	}

	// ファイル情報を取得
	fileInfo, err := s.diffReader.GetFileInfo(latestFile)
	if err != nil {
		return fmt.Errorf("ファイル情報の取得に失敗しました: %w", err)
	}

	// 詳細差分を抽出
	detailedDiff, err := s.diffReader.ExtractDetailedDiff(latestFile)
	if err != nil {
		return fmt.Errorf("詳細差分の抽出に失敗しました: %w", err)
	}

	// 結果を表示
	s.displayResult(fileInfo, detailedDiff)

	return nil
}

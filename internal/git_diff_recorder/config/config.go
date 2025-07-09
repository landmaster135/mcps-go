package config

import (
	"fmt"
)

// 見出し定数
const (
	HeaderGitDiffRecord      = "=== Git Diff Record ==="
	HeaderFileChangesSummary = "=== File Changes Summary ==="
	HeaderDetailedDiff       = "=== Detailed Diff ==="
	HeaderNewFiles           = "=== New Files ==="
	HeaderDeletedFiles       = "=== Deleted Files ==="
)

// Config はCLI設定を保持する構造体
type Config struct {
	OutputDir  string
	StagedOnly bool
	ReadMode   bool
	GenMode    bool
	SourceDir  string
	Repository string
	GitDir     string
}

// ConfigParser はConfig解析を行う構造体
type ConfigParser struct {
	flagParser FlagParser
	osArgs     OSArgs
}

// NewConfigParser は新しいConfigParserを作成する
func NewConfigParser(flagParser FlagParser, osArgs OSArgs) *ConfigParser {
	return &ConfigParser{
		flagParser: flagParser,
		osArgs:     osArgs,
	}
}

// ParseFlags はコマンドライン引数を解析してConfigを返す
func (cp *ConfigParser) ParseFlags() (*Config, error) {
	var config Config

	cp.flagParser.StringVar(&config.OutputDir, "output-dir", "", "出力先ディレクトリ (記録モード時必須)")
	cp.flagParser.BoolVar(&config.StagedOnly, "staged-only", false, "ステージング済み差分のみ記録 (デフォルト: false)")
	cp.flagParser.BoolVar(&config.ReadMode, "read-mode", false, "読み取りモードを有効にする")
	cp.flagParser.BoolVar(&config.GenMode, "gen-mode", false, "生成モードを有効にする")
	cp.flagParser.StringVar(&config.SourceDir, "source-dir", "", "読み取り対象のディレクトリ (読み取りモード時必須)")
	cp.flagParser.StringVar(&config.Repository, "repository", "", "対象リポジトリ名 (読み取りモード時必須)")
	cp.flagParser.StringVar(&config.GitDir, "git-dir", "", "対象Gitディレクトリ (生成モード時必須、記録モード時オプション)")

	if err := cp.flagParser.Parse(); err != nil {
		return nil, err
	}

	return cp.validateConfig(&config)
}

// validateConfig は設定の妥当性を検証する
func (cp *ConfigParser) validateConfig(config *Config) (*Config, error) {
	if config.GenMode {
		// 生成モードの場合
		if config.GitDir == "" {
			return nil, fmt.Errorf("生成モードでは --git-dir は必須パラメータです")
		}
	} else if config.ReadMode {
		// 読み取りモードの場合
		if config.SourceDir == "" {
			return nil, fmt.Errorf("読み取りモードでは --source-dir は必須パラメータです")
		}
		if config.Repository == "" {
			return nil, fmt.Errorf("読み取りモードでは --repository は必須パラメータです")
		}
	} else {
		// 記録モードの場合
		if config.OutputDir == "" {
			return nil, fmt.Errorf("記録モードでは --output-dir は必須パラメータです")
		}
	}

	return config, nil
}

// ParseFlags は後方互換性のための関数（既存のコードで使用されている場合）
func ParseFlags() (*Config, error) {
	flagParser := NewStandardFlagParser()
	osArgs := NewStandardOSArgs()
	configParser := NewConfigParser(flagParser, osArgs)
	return configParser.ParseFlags()
}

// PrintUsage は使用方法を表示する
func PrintUsage() {
	osArgs := NewStandardOSArgs()

	fmt.Printf("使用方法: %s [オプション]\n", osArgs.Args()[0])
	fmt.Printf("\nオプション:\n")
	fmt.Printf("  -output-dir string\n")
	fmt.Printf("        出力先ディレクトリ (記録モード時必須)\n")
	fmt.Printf("  -staged-only\n")
	fmt.Printf("        ステージング済み差分のみ記録 (デフォルト: false)\n")
	fmt.Printf("  -read-mode\n")
	fmt.Printf("        読み取りモードを有効にする\n")
	fmt.Printf("  -gen-mode\n")
	fmt.Printf("        生成モードを有効にする\n")
	fmt.Printf("  -source-dir string\n")
	fmt.Printf("        読み取り対象のディレクトリ (読み取りモード時必須)\n")
	fmt.Printf("  -repository string\n")
	fmt.Printf("        対象リポジトリ名 (読み取りモード時必須)\n")
	fmt.Printf("  -git-dir string\n")
	fmt.Printf("        対象Gitディレクトリ (生成モード時必須、記録モード時オプション)\n")
}

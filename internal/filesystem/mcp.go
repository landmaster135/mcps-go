package filesystem

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	mcp "github.com/mark3labs/mcp-go/mcp"
	server "github.com/mark3labs/mcp-go/server"
)

// FileSystemService はファイルシステム関連の機能を提供する構造体です
type FileSystemService struct {
	allowedDirectories []string
}

// NewFileSystemService は新しいFileSystemServiceを作成します
func NewFileSystemService(allowedDirs [1]string) *FileSystemService {
	// パスを正規化
	normalizedDirs := make([]string, len(allowedDirs))
	for i, dir := range allowedDirs {
		normalizedDirs[i] = filepath.Clean(expandHome(dir))
	}
	return &FileSystemService{
		allowedDirectories: normalizedDirs,
	}
}

// expandHome はパス内の ~ をホームディレクトリに展開します
func expandHome(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if path == "~" {
		return home
	}
	return filepath.Join(home, path[2:])
}

// validatePath はパスが許可されたディレクトリ内にあるか確認します
func (fs *FileSystemService) validatePath(requestedPath string) (string, error) {
	expandedPath := expandHome(requestedPath)
	absolutePath, err := filepath.Abs(expandedPath)
	if err != nil {
		return "", fmt.Errorf("パスの解決に失敗しました: %v", err)
	}

	normalizedPath := filepath.Clean(absolutePath)

	// パスが許可されたディレクトリ内にあるか確認
	isAllowed := false
	for _, dir := range fs.allowedDirectories {
		if strings.HasPrefix(normalizedPath, dir) {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return "", fmt.Errorf("アクセス拒否 - パスが許可されたディレクトリの外にあります: %s", normalizedPath)
	}

	// シンボリックリンクの場合、実際のパスも確認
	realPath, err := filepath.EvalSymlinks(absolutePath)
	if err == nil && realPath != absolutePath {
		normalizedRealPath := filepath.Clean(realPath)
		isRealPathAllowed := false
		for _, dir := range fs.allowedDirectories {
			if strings.HasPrefix(normalizedRealPath, dir) {
				isRealPathAllowed = true
				break
			}
		}
		if !isRealPathAllowed {
			return "", fmt.Errorf("アクセス拒否 - シンボリックリンクのターゲットが許可されたディレクトリの外にあります")
		}
		return realPath, nil
	}

	// 新しいファイルの場合、親ディレクトリを確認
	if os.IsNotExist(err) {
		parentDir := filepath.Dir(absolutePath)
		realParentPath, err := filepath.EvalSymlinks(parentDir)
		if err != nil {
			return "", fmt.Errorf("親ディレクトリが存在しません: %s", parentDir)
		}
		normalizedParent := filepath.Clean(realParentPath)
		isParentAllowed := false
		for _, dir := range fs.allowedDirectories {
			if strings.HasPrefix(normalizedParent, dir) {
				isParentAllowed = true
				break
			}
		}
		if !isParentAllowed {
			return "", fmt.Errorf("アクセス拒否 - 親ディレクトリが許可されたディレクトリの外にあります")
		}
		return absolutePath, nil
	}

	return absolutePath, nil
}

// ReadFile はファイルの内容を読み取ります
func (fs *FileSystemService) ReadFile(path string) (string, error) {
	validPath, err := fs.validatePath(path)
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(validPath)
	if err != nil {
		return "", fmt.Errorf("ファイルの読み取りに失敗しました: %v", err)
	}

	return string(content), nil
}

// WriteFile はファイルに内容を書き込みます
func (fs *FileSystemService) WriteFile(path string, content string) error {
	validPath, err := fs.validatePath(path)
	if err != nil {
		return err
	}

	// 親ディレクトリが存在することを確認
	parentDir := filepath.Dir(validPath)
	if _, err := os.Stat(parentDir); os.IsNotExist(err) {
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return fmt.Errorf("ディレクトリの作成に失敗しました: %v", err)
		}
	}

	err = os.WriteFile(validPath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("ファイルの書き込みに失敗しました: %v", err)
	}

	return nil
}

// CreateDirectory はディレクトリを作成します
func (fs *FileSystemService) CreateDirectory(path string) error {
	validPath, err := fs.validatePath(path)
	if err != nil {
		return err
	}

	err = os.MkdirAll(validPath, 0755)
	if err != nil {
		return fmt.Errorf("ディレクトリの作成に失敗しました: %v", err)
	}

	return nil
}

// ListDirectory はディレクトリの内容を一覧表示します
func (fs *FileSystemService) ListDirectory(path string) ([]string, error) {
	validPath, err := fs.validatePath(path)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(validPath)
	if err != nil {
		return nil, fmt.Errorf("ディレクトリの読み取りに失敗しました: %v", err)
	}

	var result []string
	for _, entry := range entries {
		prefix := "[FILE]"
		if entry.IsDir() {
			prefix = "[DIR]"
		}
		result = append(result, fmt.Sprintf("%s %s", prefix, entry.Name()))
	}

	return result, nil
}

// DirectoryTree はディレクトリの階層構造を返します
type FileTreeEntry struct {
	Name     string          `json:"name"`
	Type     string          `json:"type"`
	Children []FileTreeEntry `json:"children,omitempty"`
}

// GetDirectoryTree はディレクトリの階層構造を取得します
func (fs *FileSystemService) GetDirectoryTree(path string) ([]FileTreeEntry, error) {
	validPath, err := fs.validatePath(path)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(validPath)
	if err != nil {
		return nil, fmt.Errorf("ディレクトリの読み取りに失敗しました: %v", err)
	}

	var result []FileTreeEntry
	for _, entry := range entries {
		fileEntry := FileTreeEntry{
			Name: entry.Name(),
			Type: "file",
		}

		if entry.IsDir() {
			fileEntry.Type = "directory"
			subPath := filepath.Join(validPath, entry.Name())
			children, err := fs.GetDirectoryTree(subPath)
			if err == nil {
				fileEntry.Children = children
			}
		}

		result = append(result, fileEntry)
	}

	return result, nil
}

// MoveFile はファイルを移動します
func (fs *FileSystemService) MoveFile(source, destination string) error {
	validSource, err := fs.validatePath(source)
	if err != nil {
		return err
	}

	validDest, err := fs.validatePath(destination)
	if err != nil {
		return err
	}

	err = os.Rename(validSource, validDest)
	if err != nil {
		return fmt.Errorf("ファイルの移動に失敗しました: %v", err)
	}

	return nil
}

// SearchFiles はファイルを検索します
func (fs *FileSystemService) SearchFiles(rootPath, pattern string, excludePatterns []string) ([]string, error) {
	validPath, err := fs.validatePath(rootPath)
	if err != nil {
		return nil, err
	}

	var results []string
	err = filepath.Walk(validPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // エラーがあっても続行
		}

		// 除外パターンに一致するかチェック
		for _, excludePattern := range excludePatterns {
			matched, err := filepath.Match(excludePattern, filepath.Base(path))
			if err == nil && matched {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// パターンに一致するかチェック
		if strings.Contains(strings.ToLower(filepath.Base(path)), strings.ToLower(pattern)) {
			results = append(results, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("ファイルの検索に失敗しました: %v", err)
	}

	return results, nil
}

// FileInfo はファイル情報を表す構造体です
type FileInfo struct {
	Size        int64     `json:"size"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
	Accessed    time.Time `json:"accessed"`
	IsDirectory bool      `json:"isDirectory"`
	IsFile      bool      `json:"isFile"`
	Permissions string    `json:"permissions"`
}

// GetFileInfo はファイル情報を取得します
func (fs *FileSystemService) GetFileInfo(path string) (*FileInfo, error) {
	validPath, err := fs.validatePath(path)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(validPath)
	if err != nil {
		return nil, fmt.Errorf("ファイル情報の取得に失敗しました: %v", err)
	}

	// 一部のプラットフォームでは、これらの時間が利用できない場合があります
	var created, accessed time.Time
	// 時間取得
	created = info.ModTime()  // フォールバックとして変更時間を使用
	accessed = info.ModTime() // フォールバックとして変更時間を使用

	return &FileInfo{
		Size:        info.Size(),
		Created:     created,
		Modified:    info.ModTime(),
		Accessed:    accessed,
		IsDirectory: info.IsDir(),
		IsFile:      !info.IsDir(),
		Permissions: fmt.Sprintf("%o", info.Mode().Perm()),
	}, nil
}

func addPromptIntoServer(s *server.MCPServer) *server.MCPServer {
	prompt := mcp.NewPrompt("system_prompt_01",
		mcp.WithPromptDescription("This is a prompt for the filesystem."),
	)
	s.AddPrompt(prompt, func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return &mcp.GetPromptResult{
			Description: "System prompt for the filesystem.",
			Messages: []mcp.PromptMessage{
				{
					Role:    mcp.RoleAssistant,
					Content: mcp.NewTextContent("You use this great filesystem well."),
				},
			},
		}, nil
	})
	return s
}

// BuildFileSystemServer はファイルシステムMCPサーバーを構築する関数です
func BuildFileSystemServer() {
	// サーバーの設定
	s := server.NewMCPServer(
		"Secure Filesystem Server",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
		server.WithLogging(),
	)

	// ファイル読み取りツール
	readFileTool := mcp.NewTool("read_file",
		mcp.WithDescription("ファイルシステムからファイルの内容を読み取ります。様々なテキストエンコーディングを処理し、ファイルが読み取れない場合は詳細なエラーメッセージを提供します。単一のファイルの内容を調べる必要がある場合に使用します。許可されたディレクトリ内でのみ動作します。"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("読み取るファイルのパス"),
		),
	)

	s.AddTool(readFileTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path := request.Params.Arguments["path"].(string)
		expandedDir := expandHome(path)
		info, err := os.Stat(expandedDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "エラー: ディレクトリ %s へのアクセスに失敗しました: %v\n", path, err)
			return nil, err
		}
		if !info.IsDir() {
			fmt.Fprintf(os.Stderr, "エラー: %s はディレクトリではありません\n", path)
			return nil, err
		}
		var arr [1]string = [1]string{path}
		// サービスの初期化
		fsService := NewFileSystemService(arr)
		content, err := fsService.ReadFile(path)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(content), nil
	})

	// ファイル書き込みツール
	writeFileTool := mcp.NewTool("write_file",
		mcp.WithDescription("新しいファイルを作成するか、既存のファイルを新しい内容で完全に上書きします。警告なしに既存のファイルを上書きするため、注意して使用してください。適切なエンコーディングでテキスト内容を処理します。許可されたディレクトリ内でのみ動作します。"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("書き込むファイルのパス"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("ファイルに書き込む内容"),
		),
	)

	s.AddTool(writeFileTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path := request.Params.Arguments["path"].(string)
		content := request.Params.Arguments["content"].(string)
		expandedDir := expandHome(path)
		_, err := os.Stat(expandedDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "エラー: ディレクトリ %s へのアクセスに失敗しました: %v\n", path, err)
			return nil, err
		}
		var arr [1]string = [1]string{path}
		// サービスの初期化
		fsService := NewFileSystemService(arr)
		err = fsService.WriteFile(path, content)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(fmt.Sprintf("ファイル %s への書き込みに成功しました", path)), nil
	})

	// ディレクトリ作成ツール
	createDirTool := mcp.NewTool("create_directory",
		mcp.WithDescription("新しいディレクトリを作成するか、ディレクトリが存在することを確認します。1回の操作で複数のネストされたディレクトリを作成できます。ディレクトリが既に存在する場合、この操作は静かに成功します。プロジェクトのディレクトリ構造を設定したり、必要なパスが存在することを確認したりするのに最適です。許可されたディレクトリ内でのみ動作します。"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("作成するディレクトリのパス"),
		),
	)

	s.AddTool(createDirTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path := request.Params.Arguments["path"].(string)
		var arr [1]string = [1]string{path}
		// サービスの初期化
		fsService := NewFileSystemService(arr)
		err := fsService.CreateDirectory(path)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(fmt.Sprintf("ディレクトリ %s の作成に成功しました", path)), nil
	})

	// ディレクトリ一覧ツール
	listDirTool := mcp.NewTool("list_directory",
		mcp.WithDescription("指定されたパス内のすべてのファイルとディレクトリの詳細な一覧を取得します。結果は[FILE]と[DIR]のプレフィックスでファイルとディレクトリを明確に区別します。このツールはディレクトリ構造を理解し、ディレクトリ内の特定のファイルを見つけるのに不可欠です。許可されたディレクトリ内でのみ動作します。"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("一覧表示するディレクトリのパス"),
		),
	)

	s.AddTool(listDirTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path := request.Params.Arguments["path"].(string)
		info, err := os.Stat(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "エラー: ディレクトリ %s へのアクセスに失敗しました: %v\n", path, err)
			return nil, err
		}
		if !info.IsDir() {
			fmt.Fprintf(os.Stderr, "エラー: %s はディレクトリではありません\n", path)
			return nil, err
		}
		var arr [1]string = [1]string{path}
		// サービスの初期化
		fsService := NewFileSystemService(arr)
		entries, err := fsService.ListDirectory(path)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(strings.Join(entries, "\n")), nil
	})

	// ディレクトリツリーツール
	dirTreeTool := mcp.NewTool("directory_tree",
		mcp.WithDescription("ファイルとディレクトリの再帰的なツリービューをJSON構造として取得します。各エントリには「name」、「type」（file/directory）、ディレクトリの場合は「children」が含まれます。ファイルには子配列がなく、ディレクトリには常に子配列（空の場合もあります）があります。出力は読みやすさのために2スペースのインデントでフォーマットされています。許可されたディレクトリ内でのみ動作します。"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("ツリーを取得するディレクトリのパス"),
		),
	)

	s.AddTool(dirTreeTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path := request.Params.Arguments["path"].(string)
		info, err := os.Stat(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "エラー: ディレクトリ %s へのアクセスに失敗しました: %v\n", path, err)
			return nil, err
		}
		if !info.IsDir() {
			fmt.Fprintf(os.Stderr, "エラー: %s はディレクトリではありません\n", path)
			return nil, err
		}
		var arr [1]string = [1]string{path}
		// サービスの初期化
		fsService := NewFileSystemService(arr)
		tree, err := fsService.GetDirectoryTree(path)
		if err != nil {
			return nil, err
		}
		// JSONに変換
		jsonStr := "[\n"
		for i, entry := range tree {
			jsonStr += fmt.Sprintf("  {\n    \"name\": \"%s\",\n    \"type\": \"%s\"", entry.Name, entry.Type)
			if entry.Type == "directory" {
				jsonStr += ",\n    \"children\": []"
			}
			jsonStr += "\n  }"
			if i < len(tree)-1 {
				jsonStr += ","
			}
			jsonStr += "\n"
		}
		jsonStr += "]"
		return mcp.NewToolResultText(jsonStr), nil
	})

	// ファイル移動ツール
	moveFileTool := mcp.NewTool("move_file",
		mcp.WithDescription("ファイルとディレクトリを移動または名前変更します。ディレクトリ間でファイルを移動し、1回の操作でそれらの名前を変更できます。宛先が存在する場合、操作は失敗します。異なるディレクトリ間で動作し、同じディレクトリ内での単純な名前変更にも使用できます。ソースと宛先の両方が許可されたディレクトリ内にある必要があります。"),
		mcp.WithString("source",
			mcp.Required(),
			mcp.Description("移動するファイルまたはディレクトリのパス"),
		),
		mcp.WithString("destination",
			mcp.Required(),
			mcp.Description("移動先のパス"),
		),
	)

	s.AddTool(moveFileTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		source := request.Params.Arguments["source"].(string)
		destination := request.Params.Arguments["destination"].(string)
		_, err := os.Stat(source)
		if err != nil {
			fmt.Fprintf(os.Stderr, "エラー: ディレクトリ %s へのアクセスに失敗しました: %v\n", source, err)
			return nil, err
		}
		var arr1 [1]string = [1]string{source}
		// サービスの初期化
		_ = NewFileSystemService(arr1)
		var arr2 [1]string = [1]string{destination}
		fsService := NewFileSystemService(arr2)
		info, err := os.Stat(destination)
		if err != nil {
			fmt.Fprintf(os.Stderr, "エラー: ディレクトリ %s へのアクセスに失敗しました: %v\n", destination, err)
			return nil, err
		}
		if !info.IsDir() {
			fmt.Fprintf(os.Stderr, "エラー: %s はディレクトリではありません\n", destination)
			return nil, err
		}
		err = fsService.MoveFile(source, destination)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(fmt.Sprintf("%s から %s への移動に成功しました", source, destination)), nil
	})

	// ファイル検索ツール
	searchFilesTool := mcp.NewTool("search_files",
		mcp.WithDescription("パターンに一致するファイルとディレクトリを再帰的に検索します。開始パスからすべてのサブディレクトリを検索します。検索は大文字と小文字を区別せず、部分的な名前に一致します。一致するすべての項目への完全なパスを返します。正確な場所がわからないファイルを見つけるのに最適です。許可されたディレクトリ内でのみ検索します。"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("検索を開始するディレクトリのパス"),
		),
		mcp.WithString("pattern",
			mcp.Required(),
			mcp.Description("検索パターン"),
		),
		mcp.WithString("exclude_pattern",
			mcp.Description("除外するパターン（オプション）"),
		),
	)

	s.AddTool(searchFilesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path := request.Params.Arguments["path"].(string)
		pattern := request.Params.Arguments["pattern"].(string)
		var arr [1]string = [1]string{path}
		// サービスの初期化
		fsService := NewFileSystemService(arr)
		info, err := os.Stat(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "エラー: ディレクトリ %s へのアクセスに失敗しました: %v\n", path, err)
			return nil, err
		}
		if !info.IsDir() {
			fmt.Fprintf(os.Stderr, "エラー: %s はディレクトリではありません\n", path)
			return nil, err
		}
		var excludePatterns []string
		if excludePattern, ok := request.Params.Arguments["exclude_pattern"].(string); ok && excludePattern != "" {
			excludePatterns = append(excludePatterns, excludePattern)
		}
		results, err := fsService.SearchFiles(path, pattern, excludePatterns)
		if err != nil {
			return nil, err
		}
		if len(results) == 0 {
			return mcp.NewToolResultText("一致するものが見つかりませんでした"), nil
		}
		return mcp.NewToolResultText(strings.Join(results, "\n")), nil
	})

	// ファイル情報取得ツール
	fileInfoTool := mcp.NewTool("get_file_info",
		mcp.WithDescription("ファイルまたはディレクトリに関する詳細なメタデータを取得します。サイズ、作成時間、最終変更時間、権限、タイプなどの包括的な情報を返します。このツールは、実際の内容を読み取らずにファイルの特性を理解するのに最適です。許可されたディレクトリ内でのみ動作します。"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("情報を取得するファイルまたはディレクトリのパス"),
		),
	)

	s.AddTool(fileInfoTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path := request.Params.Arguments["path"].(string)
		_, err := os.Stat(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "エラー: ディレクトリ %s へのアクセスに失敗しました: %v\n", path, err)
			return nil, err
		}
		var arr [1]string = [1]string{path}
		// サービスの初期化
		fsService := NewFileSystemService(arr)
		info, err := fsService.GetFileInfo(path)
		if err != nil {
			return nil, err
		}
		result := fmt.Sprintf("サイズ: %d バイト\n", info.Size)
		result += fmt.Sprintf("作成日時: %s\n", info.Created.Format(time.RFC3339))
		result += fmt.Sprintf("最終変更: %s\n", info.Modified.Format(time.RFC3339))
		result += fmt.Sprintf("最終アクセス: %s\n", info.Accessed.Format(time.RFC3339))
		result += fmt.Sprintf("ディレクトリ: %t\n", info.IsDirectory)
		result += fmt.Sprintf("ファイル: %t\n", info.IsFile)
		result += fmt.Sprintf("権限: %s", info.Permissions)
		return mcp.NewToolResultText(result), nil
	})

	// 許可されたディレクトリ一覧ツール
	allowedDirsTool := mcp.NewTool("list_allowed_directories",
		mcp.WithDescription("このサーバーがアクセスできるディレクトリのリストを返します。ファイルにアクセスする前に、どのディレクトリが利用可能かを理解するために使用します。"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("確認するパス"),
		),
	)

	s.AddTool(allowedDirsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path := request.Params.Arguments["path"].(string)
		var arr [1]string = [1]string{path}
		// サービスの初期化
		fsService := NewFileSystemService(arr)
		result := "許可されたディレクトリ:\n" + strings.Join(fsService.allowedDirectories, "\n")
		return mcp.NewToolResultText(result), nil
	})

	// プロンプト
	s = addPromptIntoServer(s)

	// サーバーの起動
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("サーバーエラー: %v\n", err)
	}
}

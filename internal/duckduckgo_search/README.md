# DuckDuckGo Search MCP Server

このドキュメントでは、DuckDuckGo Search MCPサーバの実装内容と使用方法について説明します。

## 概要

DuckDuckGo Search MCPサーバは、プライバシーを重視したDuckDuckGo検索エンジンを使用してWeb検索とInstant Answer検索を提供するModel Context Protocol (MCP) サーバです。

## 主な機能

### 1. Web検索 (`duckduckgo_web_search`)
- DuckDuckGoのHTML版を使用したWeb検索
- プライバシーフォーカスで追跡やパーソナライゼーションなし
- 最大20件の検索結果を返す
- HTMLパースによる検索結果の抽出

### 2. Instant Answer検索 (`duckduckgo_instant_search`)
- DuckDuckGoのInstant Answer APIを使用
- 即座に回答が得られる事実確認、定義、計算に最適
- 要約、定義、回答、関連トピックを含む構造化情報

## 技術仕様

### レート制限
- 1秒間に1リクエスト
- 1分間に30リクエスト

### サポート機能
- HTMLエンティティのデコード
- URLのデコード（DuckDuckGoリダイレクト対応）
- 検索結果のフィルタリング
- エラーハンドリング

## 使用方法

### Web検索の例
```bash
go run main.go duckduckgo_search
```

MCPクライアントから：
```json
{
  "method": "tools/call",
  "params": {
    "name": "duckduckgo_web_search",
    "arguments": {
      "query": "Golang MCP server",
      "count": 5
    }
  }
}
```

### Instant Answer検索の例
```json
{
  "method": "tools/call", 
  "params": {
    "name": "duckduckgo_instant_search",
    "arguments": {
      "query": "what is Go programming language"
    }
  }
}
```

## API仕様

### duckduckgo_web_search
- **query** (必須): 検索クエリ（DuckDuckGoで使用できる任意のキーワード）
- **count** (オプション): 結果数（1-20、デフォルト10）

### duckduckgo_instant_search  
- **query** (必須): Instant Answer用の検索クエリ（例：'weather Tokyo', 'what is Python', '2+2'）

## 実装の詳細

### HTMLパーシング
複数の正規表現パターンを使用して、DuckDuckGoのHTML構造から検索結果を抽出：
1. `result__a`クラスのリンク抽出
2. h2タグ内のリンク抽出 
3. 汎用的なリンクパターンでフィルタリング

### エラーハンドリング
- レート制限エラー
- ネットワークエラー
- HTTPステータスエラー
- HTMLパースエラー

## プライバシー重視の設計

- ユーザー追跡なし
- 個人情報の収集なし
- バイアスのない検索結果
- プライバシーファーストの検索エンジンを使用

## テスト

実装には以下のテストが含まれています：
- `cleanText`: HTMLエンティティとタグの除去
- `decodeURL`: URLデコーディング
- `checkRateLimit`: レート制限機能

## 依存関係

- `github.com/mark3labs/mcp-go`: MCP Go SDK
- Go標準ライブラリ（net/http, regexp, strings など）

## 制限事項

- オフセット（ページネーション）は現在サポートされていません
- DuckDuckGoのHTML構造に依存するため、サイト変更の影響を受ける可能性があります
- APIキーは不要ですが、レート制限が適用されます

## 今後の改善点

1. ページネーション機能の追加
2. より堅牢なHTMLパーシング
3. キャッシュ機能
4. 検索結果の品質向上
5. 画像・動画検索のサポート

## 参考

- [DuckDuckGo Instant Answer API](https://api.duckduckgo.com/)
- [Model Context Protocol](https://modelcontextprotocol.io/)
- [MCP Go SDK](https://github.com/mark3labs/mcp-go)

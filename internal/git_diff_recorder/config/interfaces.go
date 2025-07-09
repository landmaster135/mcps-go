package config

// FlagParser はフラグ解析のインターフェース
type FlagParser interface {
	StringVar(p *string, name string, value string, usage string)
	BoolVar(p *bool, name string, value bool, usage string)
	Parse() error
	Args() []string
}

// OSArgs はコマンドライン引数取得のインターフェース
type OSArgs interface {
	Args() []string
}

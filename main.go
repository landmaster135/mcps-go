package main

import (
	"fmt"
	"os"

	arith_calc "mcps-go/mcps-go/internal/arith_calc"
	brave_search "mcps-go/mcps-go/internal/brave_search"
	datetime_calc "mcps-go/mcps-go/internal/datetime_calc"
	filesystem "mcps-go/mcps-go/internal/filesystem"
	github "mcps-go/mcps-go/internal/github"
	http_request "mcps-go/mcps-go/internal/http_request"
	postgresql "mcps-go/mcps-go/internal/postgresql"
	// shell "mcps-go/mcps-go/internal/shell" // TODO: unapplicable for WSL...
	timezone "mcps-go/mcps-go/internal/timezone"
	util "mcps-go/mcps-go/internal/util"
	youtube_transcript "mcps-go/mcps-go/internal/youtube_transcript"
)

func main() {
	args := os.Args
	// check arguments
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run main.go [arguments]")
		fmt.Fprintln(os.Stderr, "arguments are lack")
		os.Exit(1)
	}
	for i, arg := range args[1:] {
		fmt.Fprintf(os.Stderr, "argument %d: %s\n", i+1, arg)
	}

	util.OutLog("main: building mcp server...")
	a1 := args[1]
	switch a1 {
	case "arith_calc":
		arith_calc.BuildArithCalculatorServer()
	case "datetime_calc":
		datetime_calc.BuildTimeCalculatorServer()
	case "http_request":
		http_request.BuildMcpServer()
	case "brave_web_search":
		brave_search.BuildBraveSearchServer()
	case "timezone":
		timezone.BuildTimezoneServer()
	case "filesystem":
		filesystem.BuildFileSystemServer()
	case "youtube_transcript":
		youtube_transcript.BuildYouTubeTranscriptServer()
	case "github":
		github.BuildGitHubServer()
	case "postgresql":
		postgresql.BuildPostgreSQLServer()
	default:
		fmt.Fprintln(os.Stderr, "argument is invalid")
		os.Exit(1)
	}
	util.OutLog("main: built mcp server!")
}

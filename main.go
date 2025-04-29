package main

import (
	"fmt"
	"os"

	arith_calc "github.com/landmaster135/mcps-go/internal/arith_calc"
	brave_search "github.com/landmaster135/mcps-go/internal/brave_search"
	datetime_calc "github.com/landmaster135/mcps-go/internal/datetime_calc"
	everart "github.com/landmaster135/mcps-go/internal/everart"
	figma "github.com/landmaster135/mcps-go/internal/figma"
	filesystem "github.com/landmaster135/mcps-go/internal/filesystem"
	github "github.com/landmaster135/mcps-go/internal/github"
	http_request "github.com/landmaster135/mcps-go/internal/http_request"
	postgresql "github.com/landmaster135/mcps-go/internal/postgresql"
	sequentialthinking "github.com/landmaster135/mcps-go/internal/sequentialthinking"
	// shell "github.com/landmaster135/mcps-go/internal/shell" // TODO: unapplicable for WSL...
	timezone "github.com/landmaster135/mcps-go/internal/timezone"
	util "github.com/landmaster135/mcps-go/internal/util"
	youtube_transcript "github.com/landmaster135/mcps-go/internal/youtube_transcript"
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
	case "everart":
		everart.BuildEverArtServer()
	case "sequentialthinking":
		sequentialthinking.BuildSequentialThinkingServer()
	case "figma":
		figma.BuildFigmaServer()
	default:
		fmt.Fprintln(os.Stderr, "argument is invalid")
		os.Exit(1)
	}
	util.OutLog("main: built mcp server!")
}

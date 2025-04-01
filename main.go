package main

import (
	"fmt"
	"os"

	arith_calc "example.com/mcps-go/arith_calc"
	brave_search "example.com/mcps-go/brave_search"
	datetime_calc "example.com/mcps-go/datetime_calc"
	filesystem "example.com/mcps-go/filesystem"
	// github "example.com/mcps-go/github"
	github "example.com/mcps-go/github"
	http_request "example.com/mcps-go/http_request"
	util "example.com/mcps-go/util"
	timezone "example.com/mcps-go/timezone"
	youtube_transcript "example.com/mcps-go/youtube_transcript"
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
		fmt.Printf("argument %d: %s\n", i+1, arg)
	}

	util.OutLog("main: building mcp server...")
	a1 := args[1]
	switch a1 {
	case "arith_calc":
		arith_calc.BuildCalculatorServer()
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
	default:
		fmt.Fprintln(os.Stderr, "argument is invalid")
		os.Exit(1)
	}
	util.OutLog("main: built mcp server!")

	// var c datetime_calc.DatetimeCalculator
	// result := c.AddDatetime(2023, 12, 15, 10, 30, 45, 0, 1, 0, 0, 0, 0)
	// fmt.Print((result))
}

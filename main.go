package main

import (
	"fmt"
	"os"

	arith_calc "example.com/mcps-go/arith_calc"
	mypkg "example.com/mcps-go/mypkg"
	time_calc "example.com/mcps-go/time_calc"
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

	mypkg.OutLog("main: building mcp server...")

	a1 := args[1]
	switch a1 {
	case "arith_calc":
		arith_calc.BuildCalculatorServer()
	case "time_calc":
		time_calc.BuildTimeCalculatorServer()
	default:
		fmt.Fprintln(os.Stderr, "argument is invalid")
		os.Exit(1)
	}
	mypkg.OutLog("main: built mcp server!")
}

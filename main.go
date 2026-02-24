package main

import (
	"fmt"
	"lem-in/utils"
	"os"
	"time"
)

// =====================================================================
// MAIN
// =====================================================================

func main() {
	start := time.Now()
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: go run . <filename>")
		os.Exit(1)
	}
	colony, err := utils.ParseInput(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}

	for _, line := range colony.RawLines {
		fmt.Println(line)
	}
	fmt.Println()

	paths, err := utils.FindPaths(colony)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}

	for _, move := range utils.Simulate(paths, colony.NumAnts) {
		fmt.Println(move)
	}
	fmt.Fprintf(os.Stderr, "\nTime: %v\n", time.Since(start))
}

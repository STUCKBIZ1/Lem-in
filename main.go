package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: go run . <file>")
		os.Exit(1)
	}

	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: invalid data format, cannot read file")
		os.Exit(1)
	}

	farm, err := ParseFarm(string(data))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	paths := FindPaths(farm)
	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "ERROR: invalid data format, no path found")
		os.Exit(1)
	}

	moves := SimulateAnts(farm.NumAnts, paths)

	// Print the input file content
	fmt.Print(string(data))
	fmt.Println()
	fmt.Println()
	// Print the moves
	for _, line := range moves {
		fmt.Println(line)
	}
}
package main

import (
	"fmt"
	"sort"
	"strings"
)

// SimulateAnts assigns ants to paths and returns the list of turn strings.
func SimulateAnts(numAnts int, paths [][]string) []string {
	// Distribute ants across paths optimally (greedy: always send next ant
	// to the path that will finish it soonest).
	type AntState struct {
		antID    int
		pathIdx  int
		stepIdx  int // current position index in path
	}

	// Assign ants to paths
	// For each ant, pick the path where adding it results in the minimum finish turn.
	antCounts := make([]int, len(paths))  // how many ants assigned to each path
	antAssign := make([]int, numAnts+1)   // antAssign[antID] = pathIdx

	for ant := 1; ant <= numAnts; ant++ {
		best := -1
		bestTurn := -1
		for pi := range paths {
			// If we send this ant on path pi, finish turn = len(path)-1 + antCounts[pi]
			finishTurn := (len(paths[pi]) - 1) + antCounts[pi]
			if best == -1 || finishTurn < bestTurn {
				bestTurn = finishTurn
				best = pi
			}
		}
		antAssign[ant] = best
		antCounts[best]++
	}

	// Now simulate turn by turn.
	// Each ant on path pi enters 1 step per turn.
	// The k-th ant on path pi (0-indexed) starts moving on turn k+1.

	// pathAntOrder[pi] = list of ant IDs assigned to path pi, in order
	pathAntOrder := make([][]int, len(paths))
	for ant := 1; ant <= numAnts; ant++ {
		pi := antAssign[ant]
		pathAntOrder[pi] = append(pathAntOrder[pi], ant)
	}

	// antStep[antID] = current step index in its path (-1 = not started, 0 = at start, ...)
	antStep := make([]int, numAnts+1)
	for i := range antStep {
		antStep[i] = -1 // not started
	}

	// antStartTurn[antID] = the turn it first moves
	antStartTurn := make([]int, numAnts+1)
	for pi, ants := range pathAntOrder {
		_ = paths[pi]
		for k, ant := range ants {
			antStartTurn[ant] = k + 1
		}
	}

	// Calculate total turns needed
	totalTurns := calcTurns(numAnts, paths)

	var result []string

	for turn := 1; turn <= totalTurns; turn++ {
		var moves []string

		for ant := 1; ant <= numAnts; ant++ {
			if turn < antStartTurn[ant] {
				continue
			}
			pi := antAssign[ant]
			path := paths[pi]

			// Where is this ant? step since it started moving
			stepsSinceStart := turn - antStartTurn[ant]
			// position index in path
			posIdx := stepsSinceStart + 1 // after stepsSinceStart moves, at index stepsSinceStart+1
			// but it might have already finished
			if posIdx >= len(path) {
				continue
			}
			// Still in the path — record this move (its current room)
			room := path[posIdx]
			moves = append(moves, fmt.Sprintf("L%d-%s", ant, room))
		}

		if len(moves) > 0 {
			// Sort for deterministic output
			sort.Strings(moves)
			result = append(result, strings.Join(moves, " "))
		}
	}

	return result
}
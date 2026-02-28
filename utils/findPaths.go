package utils

import "fmt"

func FindPaths(colony *Colony) ([][]string, error) {
	graph, cap := buildResidualGraph(colony)

	// Save original capacities for flow decomposition
	origCap := map[[2]string]int{}
	for k, v := range cap {
		origCap[k] = v
	}

	source := nodeIn(colony.StartRoom)
	sink := nodeOut(colony.EndRoom)

	// Run Edmonds-Karp to find maximum flow (= max number of parallel paths)
	edmondsKarp(graph, cap, source, sink)

	// Extract all paths from the flow
	allPaths := decomposeFlow(graph, cap, origCap, source, sink)

	if len(allPaths) == 0 {
		return nil, fmt.Errorf("invalid data format, no path between start and end")
	}

	bestTurns := -1
	var bestPaths [][]string

	for k := 1; k <= len(allPaths); k++ {
		turns := countTurns(allPaths[:k], colony.NumAnts)
		if bestTurns == -1 || turns < bestTurns {
			bestTurns = turns
			bestPaths = make([][]string, k)
			copy(bestPaths, allPaths[:k])
		} else {
			break
		}
	}

	return bestPaths, nil
}

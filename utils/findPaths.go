package utils

import "fmt"

// findPaths finds the optimal set of node-disjoint paths minimizing total turns.
// Uses Edmonds-Karp max-flow to correctly handle cases where greedy BFS
// would pick a short path that blocks better parallel routes.
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

	// Now find the BEST subset of these paths for N ants.
	// Sort paths by length (shortest first).
	// Try using 1 path, 2 paths, 3 paths... pick the subset with min turns.
	// Since more shorter paths are always better, we sort by length and add greedily.
	sortPathsByLength(allPaths)

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

package utils

import (
	"strings"
)

// buildResidualGraph builds the flow network with node splitting.
// Returns: graph adjacency list, capacity map, list of all node names.
func buildResidualGraph(colony *Colony) (map[string][]string, map[[2]string]int) {
	cap := map[[2]string]int{}
	graph := map[string][]string{}

	addEdge := func(u, v string, c int) {
		key := [2]string{u, v}
		cap[key] += c
		// ensure both directions exist in adjacency list (residual has reverse edges)
		found := false
		for _, n := range graph[u] {
			if n == v {
				found = true
				break
			}
		}
		if !found {
			graph[u] = append(graph[u], v)
		}
		found = false
		for _, n := range graph[v] {
			if n == u {
				found = true
				break
			}
		}
		if !found {
			graph[v] = append(graph[v], u)
		}
	}

	// Node splitting: each room becomes room_in -> room_out
	for name := range colony.Rooms {
		c := 1
		if name == colony.StartRoom || name == colony.EndRoom {
			c = colony.NumAnts // unlimited for start/end
		}
		addEdge(nodeIn(name), nodeOut(name), c)
	}

	// Tunnel edges: A-B becomes A_out->B_in and B_out->A_in (both cap 1)
	seen := map[[2]string]bool{}
	for a, neighbors := range colony.Links {
		for _, b := range neighbors {
			key := [2]string{a, b}
			rev := [2]string{b, a}
			if seen[key] || seen[rev] {
				continue
			}
			seen[key] = true
			addEdge(nodeOut(a), nodeIn(b), 1)
			addEdge(nodeOut(b), nodeIn(a), 1)
		}
	}

	return graph, cap
}

// bfsResidual finds an augmenting path from source to sink in the residual graph.
// Returns the predecessor map (nil if no path exists).
func bfsResidual(graph map[string][]string, cap map[[2]string]int, source, sink string) map[string]string {
	prev := map[string]string{source: ""}
	queue := []string{source}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, next := range graph[cur] {
			if _, seen := prev[next]; seen {
				continue
			}
			if cap[[2]string{cur, next}] > 0 {
				prev[next] = cur
				if next == sink {
					return prev
				}
				queue = append(queue, next)
			}
		}
	}
	return nil
}

// edmondsKarp runs BFS-based max-flow and returns the saturated capacity map.
func edmondsKarp(graph map[string][]string, cap map[[2]string]int, source, sink string) {
	for {
		prev := bfsResidual(graph, cap, source, sink)
		if prev == nil {
			break
		}
		// All edges in this graph have capacity 1, so the bottleneck is always 1
		flow := 1
		// Update residual capacities
		for node := sink; node != source; node = prev[node] {
			p := prev[node]
			cap[[2]string{p, node}] -= flow
			cap[[2]string{node, p}] += flow
		}
	}
}

// decomposeFlow extracts actual paths by following used edges (forward flow).
// An edge u->v has been "used" if its capacity decreased from the original.
// We detect this by checking: if cap[u->v] < original, flow went through it.
// Simpler: after E-K, a forward edge was used if cap[v->u] increased (i.e., > 0
// and was 0 before). We track "used" as: original_cap - current_cap > 0.
func decomposeFlow(graph map[string][]string, cap map[[2]string]int, origCap map[[2]string]int, source, sink string) [][]string {
	var paths [][]string

	for {
		// DFS/BFS following edges where flow was sent (origCap > currentCap)
		prev := map[string]string{source: ""}
		queue := []string{source}
		found := false
		for len(queue) > 0 && !found {
			cur := queue[0]
			queue = queue[1:]
			for _, next := range graph[cur] {
				if _, seen := prev[next]; seen {
					continue
				}
				key := [2]string{cur, next}
				// Flow was sent on this edge if original capacity > current capacity
				if origCap[key] > cap[key] {
					prev[next] = cur
					if next == sink {
						found = true
						break
					}
					queue = append(queue, next)
				}
			}
		}
		if !found {
			break
		}

		// Reconstruct path
		var rawPath []string
		for node := sink; node != ""; node = prev[node] {
			rawPath = append([]string{node}, rawPath...)
		}

		// "Un-send" flow along this path (so we don't reuse it)
		for i := 0; i < len(rawPath)-1; i++ {
			u, v := rawPath[i], rawPath[i+1]
			cap[[2]string{u, v}]++ // restore forward
			cap[[2]string{v, u}]-- // remove reverse
		}

		// Convert split nodes back to room names
		// Path goes: room_in -> room_out -> room_in -> room_out -> ...
		// We only need the _in nodes (they represent entering a room)
		var roomPath []string
		for _, node := range rawPath {
			if strings.HasSuffix(node, "|in") {
				roomPath = append(roomPath, strings.TrimSuffix(node, "|in"))
			}
		}
		paths = append(paths, roomPath)
	}
	return paths
}

// sortPathsByLength sorts paths ascending by length (insertion sort, small N).
func sortPathsByLength(paths [][]string) {
	for i := 1; i < len(paths); i++ {
		key := paths[i]
		j := i - 1
		for j >= 0 && len(paths[j]) > len(key) {
			paths[j+1] = paths[j]
			j--
		}
		paths[j+1] = key
	}
}

// countTurns calculates how many turns N ants need across these paths.
// Greedy assignment: each ant takes the path with the earliest finish time.
func countTurns(paths [][]string, numAnts int) int {
	plen := make([]int, len(paths))
	for i, p := range paths {
		plen[i] = len(p) - 1
	}
	slot := make([]int, len(paths))
	maxFinish := 0
	for ant := 0; ant < numAnts; ant++ {
		best := 0
		for i := 1; i < len(paths); i++ {
			if slot[i]+plen[i] < slot[best]+plen[best] {
				best = i
			}
		}
		if f := slot[best] + plen[best]; f > maxFinish {
			maxFinish = f
		}
		slot[best]++
	}

	return maxFinish
}

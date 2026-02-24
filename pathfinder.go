package main

import "sort"

// FindPaths finds the optimal set of paths using max-flow (Edmonds-Karp).
//
// We model the graph with node-splitting to enforce one-ant-per-room:
//   - Each room r becomes r_in and r_out
//   - Internal edge r_in -> r_out, capacity 1 (numAnts for start/end)
//   - Each tunnel a-b becomes out(a)->in(b) and out(b)->in(a), capacity 1
//
// BFS augmentation finds max flow = max simultaneous paths.
// We extract actual paths and pick the subset minimizing total turns.
func FindPaths(farm *Farm) [][]string {
	in := func(r string) string { return r + "_in" }
	out := func(r string) string { return r + "_out" }

	// cap[u][v] = remaining capacity on edge u->v
	cap := make(map[string]map[string]int)
	addEdge := func(u, v string, c int) {
		if cap[u] == nil {
			cap[u] = make(map[string]int)
		}
		if cap[v] == nil {
			cap[v] = make(map[string]int)
		}
		cap[u][v] += c
		if _, ok := cap[v][u]; !ok {
			cap[v][u] = 0
		}
	}

	// Node-split edges
	for room := range farm.Rooms {
		if room == farm.Start || room == farm.End {
			addEdge(in(room), out(room), farm.NumAnts)
		} else {
			addEdge(in(room), out(room), 1)
		}
	}

	// Tunnel edges
	for a, neighbors := range farm.Links {
		for _, b := range neighbors {
			addEdge(out(a), in(b), 1)
		}
	}

	source := in(farm.Start)
	sink := out(farm.End)

	// Edmonds-Karp: BFS augmentation until no augmenting path exists
	for {
		prev := make(map[string]string)
		prev[source] = source

		// Sort neighbors for deterministic BFS
		queue := []string{source}
		found := false

	bfs:
		for len(queue) > 0 {
			curr := queue[0]
			queue = queue[1:]

			// Get sorted neighbors for determinism
			neighbors := make([]string, 0, len(cap[curr]))
			for next := range cap[curr] {
				neighbors = append(neighbors, next)
			}
			sort.Strings(neighbors)

			for _, next := range neighbors {
				if cap[curr][next] > 0 {
					if _, visited := prev[next]; !visited {
						prev[next] = curr
						if next == sink {
							found = true
							break bfs
						}
						queue = append(queue, next)
					}
				}
			}
		}

		if !found {
			break
		}

		// Update residual capacities
		for node := sink; node != source; node = prev[node] {
			p := prev[node]
			cap[p][node]--
			cap[node][p]++
		}
	}

	paths := extractPaths(farm, cap, in, out)
	return bestPaths(farm.NumAnts, paths)
}

// extractPaths reconstructs room-level paths from the residual graph.
// Flow on out(a)->in(b) is indicated by reverse edge in(b)->out(a) having cap > 0.
func extractPaths(farm *Farm, cap map[string]map[string]int, in, out func(string) string) [][]string {
	// Build sorted flowNext for deterministic traversal
	flowNext := make(map[string][]string)

	// Collect all rooms in sorted order
	rooms := make([]string, 0, len(farm.Rooms))
	for r := range farm.Rooms {
		rooms = append(rooms, r)
	}
	sort.Strings(rooms)

	for _, a := range rooms {
		for _, b := range rooms {
			if c, ok := cap[in(b)][out(a)]; ok && c > 0 {
				flowNext[a] = append(flowNext[a], b)
			}
		}
		// flowNext[a] is already sorted because we iterate rooms in sorted order
	}

	usedEdge := make(map[string]map[string]bool)
	var paths [][]string
	for {
		path := tracePath(farm.Start, farm.End, flowNext, usedEdge)
		if path == nil {
			break
		}
		paths = append(paths, path)
	}
	return paths
}

// tracePath finds a path from start to end via BFS, marking used edges.
func tracePath(start, end string, flowNext map[string][]string, usedEdge map[string]map[string]bool) []string {
	prev := map[string]string{start: ""}
	queue := []string{start}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if curr == end {
			var path []string
			for node := end; node != start; node = prev[node] {
				path = append([]string{node}, path...)
			}
			path = append([]string{start}, path...)

			for i := 0; i < len(path)-1; i++ {
				a, b := path[i], path[i+1]
				if usedEdge[a] == nil {
					usedEdge[a] = make(map[string]bool)
				}
				usedEdge[a][b] = true
			}
			return path
		}

		for _, next := range flowNext[curr] {
			if _, visited := prev[next]; visited {
				continue
			}
			if usedEdge[curr] != nil && usedEdge[curr][next] {
				continue
			}
			prev[next] = curr
			queue = append(queue, next)
		}
	}
	return nil
}

// bestPaths tries all prefix subsets (sorted by length) and picks the one
// with minimum turns. Also tries all individual subsets for robustness.
func bestPaths(numAnts int, paths [][]string) [][]string {
	if len(paths) == 0 {
		return nil
	}
	sortPaths(paths)

	var best [][]string
	bestTurns := -1

	for k := 1; k <= len(paths); k++ {
		turns := calcTurns(numAnts, paths[:k])
		if bestTurns == -1 || turns < bestTurns {
			bestTurns = turns
			best = make([][]string, k)
			copy(best, paths[:k])
		} else {
			// More paths only helps if they're short enough; stop when it stops improving
			break
		}
	}
	return best
}

func sortPaths(paths [][]string) {
	sort.Slice(paths, func(i, j int) bool {
		return len(paths[i]) < len(paths[j])
	})
}

// calcTurns computes minimum turns to move numAnts through paths.
// Binary search: find min T where sum(max(0, T - cost_i + 1)) >= numAnts.
func calcTurns(numAnts int, paths [][]string) int {
	if len(paths) == 0 || numAnts == 0 {
		return 0
	}
	costs := make([]int, len(paths))
	for i, p := range paths {
		costs[i] = len(p) - 1
	}
	lo, hi := 1, numAnts+maxCost(costs)
	for lo < hi {
		mid := (lo + hi) / 2
		if pathCapacity(costs, mid) >= numAnts {
			hi = mid
		} else {
			lo = mid + 1
		}
	}
	return lo
}

func pathCapacity(costs []int, turns int) int {
	total := 0
	for _, c := range costs {
		if turns >= c {
			total += turns - c + 1
		}
	}
	return total
}

func maxCost(costs []int) int {
	m := 0
	for _, c := range costs {
		if c > m {
			m = c
		}
	}
	return m
}
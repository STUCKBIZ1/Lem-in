package utils

import (
	"strings"
)

// buildResidualGraph builds the flow network with node splitting.
func buildResidualGraph(colony *Colony) (map[string][]string, map[[2]string]int) {
	cap := map[[2]string]int{}
	graph := map[string][]string{}

	addEdge := func(u, v string, c int) {
		key := [2]string{u, v}
		cap[key] += c
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
	for _, name := range colony.Rooms {
		c := 1
		if name == colony.StartRoom || name == colony.EndRoom {
			c = colony.NumAnts
		}
		addEdge(nodeIn(name), nodeOut(name), c)
	}

	// Tunnel edges
	for a, neighbors := range colony.Links {
		for _, b := range neighbors {
			addEdge(nodeOut(a), nodeIn(b), 1)
		}
	}

	return graph, cap
}

// bfsResidual finds an augmenting path from source to sink in the residual graph.
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

// edmondsKarp runs BFS-based max-flow.
func edmondsKarp(graph map[string][]string, cap map[[2]string]int, source, sink string) {
	for {
		prev := bfsResidual(graph, cap, source, sink)
		if prev == nil {
			break
		}
		flow := 1
		for node := sink; node != source; node = prev[node] {
			p := prev[node]
			cap[[2]string{p, node}] -= flow
			cap[[2]string{node, p}] += flow
		}
	}
}

// decomposeFlow extracts actual paths by following used edges.
func decomposeFlow(graph map[string][]string, cap map[[2]string]int, origCap map[[2]string]int, source, sink string) [][]string {
	var paths [][]string

	for {
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

		var rawPath []string
		for node := sink; node != ""; node = prev[node] {
			rawPath = append([]string{node}, rawPath...)
		}

		for i := 0; i < len(rawPath)-1; i++ {
			u, v := rawPath[i], rawPath[i+1]
			cap[[2]string{u, v}]++
			cap[[2]string{v, u}]--
		}

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

// sortPathsByLength sorts paths ascending by length.
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
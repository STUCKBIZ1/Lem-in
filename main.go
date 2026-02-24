package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// =====================================================================
// DATA STRUCTURES
// =====================================================================

type Room struct {
	Name string
	X, Y int
}

type Colony struct {
	NumAnts   int
	Rooms     map[string]*Room
	Links     map[string][]string
	StartRoom string
	EndRoom   string
	RawLines  []string
}

// =====================================================================
// PARSING
// =====================================================================

func parseInput(filename string) (*Colony, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("invalid data format, cannot read file: %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	colony := &Colony{
		Rooms:    make(map[string]*Room),
		Links:    make(map[string][]string),
		RawLines: lines,
	}

	antsParsed, nextIsStart, nextIsEnd := false, false, false
	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}
		if !antsParsed {
			n, err := strconv.Atoi(t)
			if err != nil || n <= 0 {
				return nil, fmt.Errorf("invalid data format, invalid number of ants")
			}
			colony.NumAnts = n
			antsParsed = true
			continue
		}
		if t == "##start" {
			nextIsStart = true
			continue
		}
		if t == "##end" {
			nextIsEnd = true
			continue
		}
		if strings.HasPrefix(t, "#") {
			continue
		}
		// Link: no spaces, has '-'
		if !strings.Contains(t, " ") && strings.Contains(t, "-") {
			parts := strings.SplitN(t, "-", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return nil, fmt.Errorf("invalid data format, bad link: %s", t)
			}
			a, b := parts[0], parts[1]
			if _, ok := colony.Rooms[a]; !ok {
				return nil, fmt.Errorf("invalid data format, unknown room: %s", a)
			}
			if _, ok := colony.Rooms[b]; !ok {
				return nil, fmt.Errorf("invalid data format, unknown room: %s", b)
			}
			colony.Links[a] = append(colony.Links[a], b)
			colony.Links[b] = append(colony.Links[b], a)
			continue
		}
		// Room: "name x y"
		parts := strings.Fields(t)
		if len(parts) == 3 {
			name := parts[0]
			if strings.HasPrefix(name, "L") || strings.HasPrefix(name, "#") {
				return nil, fmt.Errorf("invalid data format, bad room name: %s", name)
			}
			x, errX := strconv.Atoi(parts[1])
			y, errY := strconv.Atoi(parts[2])
			if errX != nil || errY != nil {
				return nil, fmt.Errorf("invalid data format, bad coordinates: %s", name)
			}
			if _, exists := colony.Rooms[name]; exists {
				return nil, fmt.Errorf("invalid data format, duplicate room: %s", name)
			}
			colony.Rooms[name] = &Room{Name: name, X: x, Y: y}
			if nextIsStart {
				colony.StartRoom = name
				nextIsStart = false
			} else if nextIsEnd {
				colony.EndRoom = name
				nextIsEnd = false
			}
			continue
		}
	}
	if colony.StartRoom == "" {
		return nil, fmt.Errorf("invalid data format, no start room found")
	}
	if colony.EndRoom == "" {
		return nil, fmt.Errorf("invalid data format, no end room found")
	}
	return colony, nil
}

// =====================================================================
// ALGORITHM: Edmonds-Karp max-flow with node splitting
//
// WHY THE OLD GREEDY BFS FAILED:
// --------------------------------
// Simple greedy BFS picks paths one at a time and BLOCKS rooms it uses.
// For this graph, BFS found start->h->n->e->end first (shortest, length 4).
// That blocked rooms h, n, e — which cut off two other good paths that also
// needed room n (start->0->o->n->e->end) or room h (start->h->A->c->k->end).
// Result: only 2 paths found, 9 turns. Optimal is 3 paths, 8 turns.
//
// THE FIX — EDMONDS-KARP:
// -------------------------
// Edmonds-Karp is BFS-based max-flow. The key trick: after routing a path,
// it adds REVERSE edges in a "residual graph". Future BFS can traverse
// a reverse edge to "undo" part of a previous path and re-route it.
//
// For example:
//   Path 1 (greedy): start->h->n->e->end
//   Path 2 with reverse edges: start->0->o->n->[reverse n->h]->h->A->c->k->end
//   This "steals" room h from path 1 and redirects path 1 through a different
//   branch. After flow decomposition, you get 3 clean non-overlapping paths.
//
// NODE SPLITTING:
// ---------------
// The spec says each ROOM can hold only 1 ant (not each tunnel). Standard
// max-flow only limits edge capacity. We enforce room capacity using the
// classic "node splitting" technique:
//
//   Each room X → two nodes:  X_in  and  X_out
//   Add edge:  X_in  →  X_out  with capacity 1  (limits ants through the room)
//   Tunnel A-B → two directed edges:
//     A_out → B_in  (capacity 1)
//     B_out → A_in  (capacity 1)
//
// ##start and ##end get capacity = numAnts (they hold unlimited ants).
//
// FLOW DECOMPOSITION:
// --------------------
// After Edmonds-Karp fills the residual graph, we extract actual paths by
// repeatedly doing DFS from source to sink, following edges that have been
// "used" (forward capacity decreased), until no more paths exist.
// =====================================================================

// nodeIn / nodeOut return the split node names for a room
func nodeIn(room string) string  { return room + "|in" }
func nodeOut(room string) string { return room + "|out" }

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

// findPaths finds the optimal set of node-disjoint paths minimizing total turns.
// Uses Edmonds-Karp max-flow to correctly handle cases where greedy BFS
// would pick a short path that blocks better parallel routes.
func findPaths(colony *Colony) ([][]string, error) {
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

// =====================================================================
// SIMULATION
//
// antStart[k] = the turn ant k takes its FIRST step out of ##start.
// On turn T, ant k moves to: paths[antPath[k]][ T - antStart[k] + 1 ]
//
// No collisions because:
//   - Same path: consecutive ants are 1 turn apart → never same room.
//   - Different paths: node-disjoint (enforced by max-flow node splitting).
//   - ##end: spec allows unlimited ants.
// =====================================================================

func simulate(paths [][]string, numAnts int) []string {
	plen := make([]int, len(paths))
	for i, p := range paths {
		plen[i] = len(p) - 1
	}
	slot := make([]int, len(paths))
	antPath := make([]int, numAnts)
	antStart := make([]int, numAnts)

	for ant := 0; ant < numAnts; ant++ {
		best := 0
		for i := 1; i < len(paths); i++ {
			if slot[i]+plen[i] < slot[best]+plen[best] {
				best = i
			}
		}
		antPath[ant] = best
		antStart[ant] = slot[best]
		slot[best]++
	}

	totalTurns := 0
	for ant := 0; ant < numAnts; ant++ {
		if f := antStart[ant] + plen[antPath[ant]]; f > totalTurns {
			totalTurns = f
		}
	}

	var output []string
	for turn := 0; turn < totalTurns; turn++ {
		var moves []string
		for ant := 0; ant < numAnts; ant++ {
			p := antPath[ant]
			pl := plen[p]
			if turn < antStart[ant] || turn >= antStart[ant]+pl {
				continue
			}
			step := turn - antStart[ant] + 1
			moves = append(moves, fmt.Sprintf("L%d-%s", ant+1, paths[p][step]))
		}
		if len(moves) > 0 {
			output = append(output, strings.Join(moves, " "))
		}
	}
	return output
}

// =====================================================================
// MAIN
// =====================================================================

func main() {
	start := time.Now()
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: go run . <filename>")
		os.Exit(1)
	}
	colony, err := parseInput(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}

	for _, line := range colony.RawLines {
		fmt.Println(line)
	}
	fmt.Println()

	paths, err := findPaths(colony)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}

	for _, move := range simulate(paths, colony.NumAnts) {
		fmt.Println(move)
	}
	fmt.Fprintf(os.Stderr, "\nTime: %v\n", time.Since(start))
}
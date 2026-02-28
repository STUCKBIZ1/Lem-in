package utils

import (
	"fmt"
	"strings"
)

func Simulate(paths [][]string, numAnts int) []string {
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

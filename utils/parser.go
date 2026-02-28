package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// =====================================================================
// PARSING
// =====================================================================

func ParseInput(filename string) (*Colony, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("invalid data format, cannot read file: %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	colony := &Colony{
		Links:    make(map[string][]string),
		RawLines: lines,
	}

	roomSet := make(map[string]bool)
	coordSet := make(map[[2]int]string)
	antsParsed := false
	nextIsStart := false
	nextIsEnd := false
	roomsDone := false

	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}

		// --- Comments (always skip, regardless of parse state) ---
		if strings.HasPrefix(t, "#") && t != "##start" && t != "##end" {
			continue
		}

		// --- Ant count (must be first non-comment, non-empty line) ---
		if !antsParsed {
			n, err := strconv.Atoi(t)
			if err != nil || n <= 0 {
				return nil, fmt.Errorf("invalid data format, invalid number of ants")
			}
			colony.NumAnts = n
			antsParsed = true
			continue
		}

		// --- Commands ---
		if t == "##start" {
			if roomsDone {
				return nil, fmt.Errorf("invalid data format, ##start found after links")
			}
			if colony.StartRoom != "" || nextIsStart {
				return nil, fmt.Errorf("invalid data format, duplicate ##start")
			}
			nextIsStart = true
			continue
		}
		if t == "##end" {
			if roomsDone {
				return nil, fmt.Errorf("invalid data format, ##end found after links")
			}
			if colony.EndRoom != "" || nextIsEnd {
				return nil, fmt.Errorf("invalid data format, duplicate ##end")
			}
			nextIsEnd = true
			continue
		}

		// --- Link: no spaces, has '-' ---
		if !strings.Contains(t, " ") && strings.Contains(t, "-") {
			roomsDone = true
			parts := strings.SplitN(t, "-", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return nil, fmt.Errorf("invalid data format, bad link: %s", t)
			}
			a, b := parts[0], parts[1]
			if !roomSet[a] {
				return nil, fmt.Errorf("invalid data format, unknown room: %s", a)
			}
			if !roomSet[b] {
				return nil, fmt.Errorf("invalid data format, unknown room: %s", b)
			}
			colony.Links[a] = append(colony.Links[a], b)
			colony.Links[b] = append(colony.Links[b], a)
			continue
		}

		// --- Room: "name x y" ---
		parts := strings.Fields(t)
		if len(parts) == 3 {
			if roomsDone {
				return nil, fmt.Errorf("invalid data format, room %s defined after links", parts[0])
			}
			name := parts[0]
			if strings.HasPrefix(name, "L") || strings.HasPrefix(name, "#") {
				return nil, fmt.Errorf("invalid data format, bad room name: %s", name)
			}
			x, errX := strconv.Atoi(parts[1])
			y, errY := strconv.Atoi(parts[2])
			if errX != nil || errY != nil {
				return nil, fmt.Errorf("invalid data format, bad coordinates: %s", name)
			}
			if roomSet[name] {
				return nil, fmt.Errorf("invalid data format, duplicate room: %s", name)
			}
			coord := [2]int{x, y}
			if existing, dup := coordSet[coord]; dup {
				return nil, fmt.Errorf("invalid data format, duplicate coordinates (%d,%d) for rooms %s and %s", x, y, existing, name)
			}
			roomSet[name] = true
			coordSet[coord] = name
			colony.Rooms = append(colony.Rooms, name)
			if nextIsStart {
				colony.StartRoom = name
				nextIsStart = false
			} else if nextIsEnd {
				colony.EndRoom = name
				nextIsEnd = false
			}
			continue
		}

		return nil, fmt.Errorf("invalid data format, unrecognized line: %s", t)
	}

	if colony.StartRoom == "" {
		return nil, fmt.Errorf("invalid data format, no start room found")
	}
	if colony.EndRoom == "" {
		return nil, fmt.Errorf("invalid data format, no end room found")
	}
	return colony, nil
}
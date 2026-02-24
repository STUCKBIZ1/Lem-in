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

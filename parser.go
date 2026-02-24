package main

import (
	"fmt"
	"strconv"
	"strings"
)

// Farm holds everything parsed from the input file.
type Farm struct {
	NumAnts int
	Rooms   map[string]bool // room name -> exists
	Links   map[string][]string // adjacency list
	Start   string
	End     string
}

// ParseFarm parses the input string and returns a Farm or an error.
func ParseFarm(input string) (*Farm, error) {
	farm := &Farm{
		Rooms: make(map[string]bool),
		Links: make(map[string][]string),
	}

	lines := strings.Split(strings.TrimRight(input, "\n"), "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("ERROR: invalid data format, empty file")
	}

	// First non-comment line must be the number of ants
	idx := 0
	for idx < len(lines) && strings.HasPrefix(lines[idx], "#") {
		idx++
	}
	if idx >= len(lines) {
		return nil, fmt.Errorf("ERROR: invalid data format, missing ant count")
	}

	numAnts, err := strconv.Atoi(strings.TrimSpace(lines[idx]))
	if err != nil || numAnts <= 0 {
		return nil, fmt.Errorf("ERROR: invalid data format, invalid number of ants")
	}
	farm.NumAnts = numAnts
	idx++

	// Parse rooms and links
	nextIsStart := false
	nextIsEnd := false

	for ; idx < len(lines); idx++ {
		line := strings.TrimSpace(lines[idx])

		// Empty lines are ignored
		if line == "" {
			continue
		}

		// Commands
		if line == "##start" {
			nextIsStart = true
			continue
		}
		if line == "##end" {
			nextIsEnd = true
			continue
		}

		// Comments (but not ##start/##end which were handled above)
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Link: contains '-' but no spaces
		if strings.Contains(line, "-") && !strings.Contains(line, " ") {
			parts := strings.SplitN(line, "-", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return nil, fmt.Errorf("ERROR: invalid data format, invalid link: %s", line)
			}
			a, b := parts[0], parts[1]
			if !farm.Rooms[a] || !farm.Rooms[b] {
				return nil, fmt.Errorf("ERROR: invalid data format, link to unknown room: %s", line)
			}
			farm.Links[a] = append(farm.Links[a], b)
			farm.Links[b] = append(farm.Links[b], a)
			continue
		}

		// Room: "name x y"
		parts := strings.Fields(line)
		if len(parts) == 3 {
			name := parts[0]
			// Room name must not start with 'L' or '#'
			if strings.HasPrefix(name, "L") || strings.HasPrefix(name, "#") {
				return nil, fmt.Errorf("ERROR: invalid data format, invalid room name: %s", name)
			}
			if farm.Rooms[name] {
				return nil, fmt.Errorf("ERROR: invalid data format, duplicate room: %s", name)
			}
			// Validate coordinates are integers (we don't store them)
			if _, err := strconv.Atoi(parts[1]); err != nil {
				return nil, fmt.Errorf("ERROR: invalid data format, invalid coordinate")
			}
			if _, err := strconv.Atoi(parts[2]); err != nil {
				return nil, fmt.Errorf("ERROR: invalid data format, invalid coordinate")
			}
			farm.Rooms[name] = true

			if nextIsStart {
				if farm.Start != "" {
					return nil, fmt.Errorf("ERROR: invalid data format, multiple start rooms")
				}
				farm.Start = name
				nextIsStart = false
			} else if nextIsEnd {
				if farm.End != "" {
					return nil, fmt.Errorf("ERROR: invalid data format, multiple end rooms")
				}
				farm.End = name
				nextIsEnd = false
			}
			continue
		}

		// Unknown line — ignore per spec
	}

	if farm.Start == "" {
		return nil, fmt.Errorf("ERROR: invalid data format, no start room found")
	}
	if farm.End == "" {
		return nil, fmt.Errorf("ERROR: invalid data format, no end room found")
	}

	return farm, nil
}
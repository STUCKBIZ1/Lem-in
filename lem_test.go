package main

import (
	"strings"
	"testing"
)

func TestParseFarm_Valid(t *testing.T) {
	input := `3
##start
0 1 0
##end
1 5 0
2 9 0
3 13 0
0-2
2-3
3-1
`
	farm, err := ParseFarm(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if farm.NumAnts != 3 {
		t.Errorf("expected 3 ants, got %d", farm.NumAnts)
	}
	if farm.Start != "0" {
		t.Errorf("expected start=0, got %s", farm.Start)
	}
	if farm.End != "1" {
		t.Errorf("expected end=1, got %s", farm.End)
	}
}

func TestParseFarm_NoStart(t *testing.T) {
	input := `3
0 1 0
1 5 0
0-1
`
	_, err := ParseFarm(input)
	if err == nil || !strings.Contains(err.Error(), "no start") {
		t.Errorf("expected no-start error, got %v", err)
	}
}

func TestParseFarm_NoEnd(t *testing.T) {
	input := `3
##start
0 1 0
1 5 0
0-1
`
	_, err := ParseFarm(input)
	if err == nil || !strings.Contains(err.Error(), "no end") {
		t.Errorf("expected no-end error, got %v", err)
	}
}

func TestParseFarm_ZeroAnts(t *testing.T) {
	input := `0
##start
0 1 0
##end
1 5 0
0-1
`
	_, err := ParseFarm(input)
	if err == nil {
		t.Error("expected error for 0 ants")
	}
}

func TestParseFarm_DuplicateRoom(t *testing.T) {
	input := `1
##start
0 1 0
0 2 0
##end
1 5 0
0-1
`
	_, err := ParseFarm(input)
	if err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("expected duplicate error, got %v", err)
	}
}

func TestFindPaths_Simple(t *testing.T) {
	farm := &Farm{
		NumAnts: 1,
		Start:   "a",
		End:     "b",
		Rooms:   map[string]bool{"a": true, "b": true},
		Links:   map[string][]string{"a": {"b"}, "b": {"a"}},
	}
	paths := FindPaths(farm)
	if len(paths) == 0 {
		t.Fatal("expected at least one path")
	}
	if paths[0][0] != "a" || paths[0][len(paths[0])-1] != "b" {
		t.Errorf("path doesn't go from a to b: %v", paths[0])
	}
}

func TestFindPaths_NoPath(t *testing.T) {
	farm := &Farm{
		NumAnts: 1,
		Start:   "a",
		End:     "b",
		Rooms:   map[string]bool{"a": true, "b": true},
		Links:   map[string][]string{},
	}
	paths := FindPaths(farm)
	if len(paths) != 0 {
		t.Error("expected no path")
	}
}

func TestCalcTurns(t *testing.T) {
	// 3 ants, 1 path of length 4 (cost 3 steps): turns = 3 + 3 - 1 = 5
	paths := [][]string{{"s", "a", "b", "e"}}
	turns := calcTurns(3, paths)
	if turns != 5 {
		t.Errorf("expected 5 turns, got %d", turns)
	}
}

func TestSimulateAnts_SinglePath(t *testing.T) {
	paths := [][]string{{"0", "2", "3", "1"}}
	moves := SimulateAnts(3, paths)
	// Expected 5 turns, all ants eventually reach "1"
	if len(moves) == 0 {
		t.Error("expected moves")
	}
}
package utils

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

package utils

type Colony struct {
	NumAnts   int
	Rooms     []string // room names only
	Links     map[string][]string
	StartRoom string
	EndRoom   string
	RawLines  []string
}
package utils

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
// nodeIn / nodeOut return the split node names for a room
func nodeIn(room string) string  { return room + "|in" }
func nodeOut(room string) string { return room + "|out" }

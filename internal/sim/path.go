package sim

import "github.com/evanstern/promptworld/internal/worldmap"

// Deterministic grid search. Neighbor order is fixed (N, E, S, W) and the
// frontier is a FIFO queue, so shortest paths and "nearest match" results are
// identical on every run — a requirement, since intent targets and each
// movement step are derived from these.

var neighborOrder = [4][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}

// bfs runs breadth-first search from (sx, sy) over passable terrain, calling
// visit for each reached tile in deterministic order (including the start).
// visit returns true to stop. cameFrom is returned for path reconstruction.
func bfs(m *worldmap.Map, s *State, sx, sy int, visit func(x, y int) bool) (stopX, stopY int, cameFrom map[Point]Point, found bool) {
	start := Point{X: sx, Y: sy}
	cameFrom = map[Point]Point{start: start}
	queue := []Point{start}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if visit(cur.X, cur.Y) {
			return cur.X, cur.Y, cameFrom, true
		}
		for _, d := range neighborOrder {
			nx, ny := cur.X+d[0], cur.Y+d[1]
			np := Point{X: nx, Y: ny}
			if _, seen := cameFrom[np]; seen || !passable(m, s, nx, ny) {
				continue
			}
			cameFrom[np] = cur
			queue = append(queue, np)
		}
	}
	return 0, 0, cameFrom, false
}

// nextStep returns the next tile on a shortest path toward the target, or
// the current position if the target is unreachable (callers abandon the
// intent). The escape clause: from an impassable tile (pre-terrain saves),
// any in-bounds neighbor is a legal first step toward open ground.
func nextStep(m *worldmap.Map, s *State, fromX, fromY, toX, toY int) (int, int) {
	if fromX == toX && fromY == toY {
		return fromX, fromY
	}
	if !passable(m, s, fromX, fromY) {
		for _, d := range neighborOrder {
			if nx, ny := fromX+d[0], fromY+d[1]; passable(m, s, nx, ny) {
				return nx, ny
			}
		}
		for _, d := range neighborOrder {
			if nx, ny := fromX+d[0], fromY+d[1]; m.InBounds(nx, ny) {
				return nx, ny
			}
		}
		return fromX, fromY
	}
	_, _, cameFrom, found := bfs(m, s, fromX, fromY, func(x, y int) bool {
		return x == toX && y == toY
	})
	if !found {
		return fromX, fromY
	}
	// Walk back from the target to the tile adjacent to the start.
	cur := Point{X: toX, Y: toY}
	start := Point{X: fromX, Y: fromY}
	for cameFrom[cur] != start {
		cur = cameFrom[cur]
	}
	return cur.X, cur.Y
}

// nearest finds the closest reachable tile satisfying match, in BFS order.
func nearest(m *worldmap.Map, s *State, fromX, fromY int, match func(x, y int) bool) (Point, bool) {
	x, y, _, found := bfs(m, s, fromX, fromY, match)
	return Point{X: x, Y: y}, found
}

// nearestKnown is the knowledge-gated twin of nearest (spec 041 US1, research
// D3): the closest reachable tile holding a FRESH fact of kind in the ACTING
// agent's map, searched from the agent's own position. The BFS geometry —
// deterministic neighbor order over ground-truth passability — is untouched;
// only the match closure is gated, so "nearest known" keeps exactly nearest's
// tie-breaking. also (nil = none) layers a verb's extra ground conditions
// (den readiness, chest contents) on top of the knowledge gate. Candidates
// are the agent's BELIEFS: a remembered place that has since vanished still
// resolves — arrival re-validation is the correction moment (D3, US3).
func nearestKnown(m *worldmap.Map, s *State, a *Agent, kind string, now int64, also func(x, y int) bool) (Point, bool) {
	return nearest(m, s, a.X, a.Y, func(x, y int) bool {
		return knownFactAt(a, kind, x, y, now) && (also == nil || also(x, y))
	})
}

// nearestKnownAdjacentTo is the knowledge-gated twin of nearestAdjacentTo:
// the closest passable stand tile beside a tile holding a fresh fact of kind
// in the acting agent's map (chop/quarry/collect_water's adjacent-stand
// shape). Same gating contract as nearestKnown.
func nearestKnownAdjacentTo(m *worldmap.Map, s *State, a *Agent, kind string, now int64, also func(x, y int) bool) (stand, res Point, ok bool) {
	return nearestAdjacentTo(m, s, a.X, a.Y, func(x, y int) bool {
		return knownFactAt(a, kind, x, y, now) && (also == nil || also(x, y))
	})
}

// nearestFrontier finds the agent's nearest exploration frontier (spec 041
// US4, research D4 — Yamauchi-style): the closest reachable tile that the
// agent's map marks EXPLORED and that 4-neighbors at least one in-bounds
// UNEXPLORED tile. The BFS runs over ground-truth passability with the shared
// deterministic neighbor order, so tie-breaking matches every other nearest-*
// helper; the bitmap is decoded once and shared by the whole search. Not
// found means the reachable world is fully explored — the search verb's
// honest exhaustion. A map-less agent has no frontier.
func nearestFrontier(m *worldmap.Map, s *State, a *Agent) (Point, bool) {
	if a.Map == nil {
		return Point{}, false
	}
	bits := exploredBytes(a.Map.Explored, m.W, m.H)
	explored := func(x, y int) bool {
		if x < 0 || y < 0 || x >= m.W || y >= m.H {
			return false
		}
		i := y*m.W + x
		return bits[i/8]&(1<<(i%8)) != 0
	}
	return nearest(m, s, a.X, a.Y, func(x, y int) bool {
		if !explored(x, y) {
			return false
		}
		for _, d := range neighborOrder {
			nx, ny := x+d[0], y+d[1]
			if m.InBounds(nx, ny) && !explored(nx, ny) {
				return true
			}
		}
		return false
	})
}

// nearestAdjacentTo finds the closest passable tile that neighbors a tile
// satisfying matchRes (e.g. stand beside a tree to chop it). Returns both
// the standing tile and the resource tile.
func nearestAdjacentTo(m *worldmap.Map, s *State, fromX, fromY int, matchRes func(x, y int) bool) (stand, res Point, ok bool) {
	sx, sy, _, found := bfs(m, s, fromX, fromY, func(x, y int) bool {
		for _, d := range neighborOrder {
			if matchRes(x+d[0], y+d[1]) {
				return true
			}
		}
		return false
	})
	if !found {
		return Point{}, Point{}, false
	}
	for _, d := range neighborOrder {
		if matchRes(sx+d[0], sy+d[1]) {
			return Point{X: sx, Y: sy}, Point{X: sx + d[0], Y: sy + d[1]}, true
		}
	}
	return Point{}, Point{}, false
}

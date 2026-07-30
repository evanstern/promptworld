package sim

// The guardian's myth briefing (spec 101 D5, FR-004): a READ-ONLY derivation
// over the existing belief corpus — no new events, no new state, no
// self-grading (a canonization is never counted as evidence for or against a
// myth; the derivation only ever reads Agent.Beliefs). Computed fresh on
// every call, so it can never go stale and there is nothing to prune.
//
// Only Beliefs (Subject == -1, "about the world") are structurally eligible:
// a Rumor/KnownRumor's Subject is always an agent index (social.go), so a
// rumor is never a place-claim to derive a myth from — there is no
// rumor-to-place linkage to invent here.

import (
	"regexp"
	"sort"
	"strconv"
)

// mythCoordRe extracts "(x,y)" coordinate pairs from a belief statement — the
// exact form every situated memory bakes (situateText, memory.go) and
// therefore the form a consolidation-authored place-belief overwhelmingly
// carries. Declared again here rather than imported: internal/mind (which
// owns the twin statementCoordRe, reconcile.go) depends on internal/sim, not
// the reverse, so sim cannot see it — the tool.MiracleKinds "mirrored, not
// imported" leaf-package pattern this codebase already uses for the same
// layering reason.
var mythCoordRe = regexp.MustCompile(`\((\d+)\s*,\s*(\d+)\)`)

// mythClusterCell buckets a coordinate into a coarse grid cell so belief
// statements naming nearby-but-not-identical coordinates — the likely shape
// of independently-worded tellings of the same myth — still cluster as one
// candidate.
const mythClusterCell = 8

// PlaceMythBriefing is one dominant candidate myth the guardian's read-only
// brief_myths tool offers as a canonization candidate (spec 101 D5).
type PlaceMythBriefing struct {
	X          int    `json:"x"`          // the cluster's anchor coordinate
	Y          int    `json:"y"`          // (the first-encountered belief's own coordinate)
	Statement  string `json:"statement"`  // the most-held wording at this anchor
	Holders    int    `json:"holders"`    // distinct living villagers holding a belief in this cluster
	Confidence int    `json:"confidence"` // average confidence among them, 0..100
}

// mythCluster accumulates one coordinate-bucket's beliefs while walking the
// agent roster.
type mythCluster struct {
	x, y        int
	countByText map[string]int
	holders     map[int]bool
	totalConf   int
	n           int
}

// DominantPlaceMyths derives up to topN candidate place-myths from the
// current belief corpus (spec 101 D5): every LIVING villager's world beliefs
// (Subject == -1) naming a coordinate are grouped into coarse coordinate
// clusters, the most-repeated wording per cluster is kept, and clusters rank
// by holder count then average confidence, descending — deterministic
// (ties broken by coordinate then text), a pure function of s. topN <= 0
// returns every cluster found.
func (s *State) DominantPlaceMyths(topN int) []PlaceMythBriefing {
	clusters := map[[2]int]*mythCluster{}
	for ai := range s.Agents {
		a := &s.Agents[ai]
		if a.Dead {
			continue
		}
		for _, b := range a.Beliefs {
			if b.Subject != -1 {
				continue // a belief about a villager, not the world/a place
			}
			m := mythCoordRe.FindStringSubmatch(b.Statement)
			if m == nil {
				continue
			}
			x, err1 := strconv.Atoi(m[1])
			y, err2 := strconv.Atoi(m[2])
			if err1 != nil || err2 != nil {
				continue
			}
			key := [2]int{x / mythClusterCell, y / mythClusterCell}
			c := clusters[key]
			if c == nil {
				c = &mythCluster{x: x, y: y, countByText: map[string]int{}, holders: map[int]bool{}}
				clusters[key] = c
			}
			c.countByText[b.Statement]++
			c.totalConf += b.Confidence
			c.n++
			c.holders[ai] = true
		}
	}
	out := make([]PlaceMythBriefing, 0, len(clusters))
	for _, c := range clusters {
		best, bestN := "", 0
		for txt, n := range c.countByText {
			if n > bestN || (n == bestN && txt < best) {
				best, bestN = txt, n
			}
		}
		conf := 0
		if c.n > 0 {
			conf = c.totalConf / c.n
		}
		out = append(out, PlaceMythBriefing{
			X: c.x, Y: c.y, Statement: best, Holders: len(c.holders), Confidence: conf,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Holders != out[j].Holders {
			return out[i].Holders > out[j].Holders
		}
		if out[i].Confidence != out[j].Confidence {
			return out[i].Confidence > out[j].Confidence
		}
		if out[i].X != out[j].X {
			return out[i].X < out[j].X
		}
		if out[i].Y != out[j].Y {
			return out[i].Y < out[j].Y
		}
		return out[i].Statement < out[j].Statement
	})
	if topN > 0 && len(out) > topN {
		out = out[:topN]
	}
	return out
}

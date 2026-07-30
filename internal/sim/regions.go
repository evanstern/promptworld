package sim

// Named regions — the guardian's canonization miracle (spec 101): the "yes,
// and" answer to emergent mythology (Thornspire, 2026-07-23). Where spec 097
// only lets reality DEBUNK a villager-coined myth, this lets the guardian
// ANSWER it: christen a region (center + radius + villager-coined name) and,
// optionally, raise ONE feature of an existing placeable kind within it — a
// single atomic act landing as one recorded guardian.region_named event.
//
// Region clones the spec-084 designation/directive entity discipline
// verbatim (D1): a deterministic human-readable id minted guardian-side with
// no RNG, a validate-not-clamp door (the InjectSocial dry-run runs it on a
// state copy, so a bad naming is rejected before recording and a recorded one
// always re-applies cleanly in replay), and the active+recent-32 retention
// prune. UNLIKE a designation, a region carries no lifecycle beyond its
// christening in v1 — there is no cancel/rename event (the spec's own edge
// case: "renames are future work") — so Status is always "active" and the
// prune call is inert by construction (kept for shape parity and forward
// compatibility, exactly as the comment on prunePlanEntities below records).
//
// D4 (economy) — the charge-shape choice, RECORDED: canonization costs
// GuardianRegionCharge (2) flat charges, no cooldown. The doctrine offered a
// flat premium OR a charge+cooldown; a cooldown needs new per-world cooldown
// state this feature has no other use for, while 2 charges reuses the
// existing "dearest miracle" band (guardian.time_snapped's price,
// miracles.go) with zero new state — the simpler, doctrine-consistent
// choice. See docs/wiki/guardian-canonization.md for the mirrored prose.
//
// D3 (perception) needs no new machinery, by construction: a region's name
// flows into situated place text via describePlace (memory.go's
// featureDesc, below), which the mind's belief reconciliation NEVER reads —
// reconciliation matches a belief's coordinate against the closed
// feature-KIND vocabulary observedKinds emits (observe.go), a vocabulary a
// canonized region's placed structure (or the ground truth already inside
// its radius — trees, rock, water) already speaks. Naming a region changes
// what the place is CALLED in prose; it changes nothing about what
// observedKinds finds there, so a myth naming a real feature confirms on the
// next arrival through the unmodified spec-097 channel.

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/evanstern/promptworld/internal/store"
)

const (
	// GuardianRegionCap bounds concurrent named regions (spec 101 FR-001):
	// every region is permanently active in v1 (no terminal event exists), so
	// this is simply the count of s.Regions — the designation cap's rationale
	// (a bounded, human-scannable set of place-claims), matching
	// GuardianDesignationCap's value.
	GuardianRegionCap = 16
	// regionNameMaxRunes matches the designation label cap (plans.go) — a
	// villager-coined toponym is exactly that kind of short human label.
	regionNameMaxRunes = 80
	// regionRadiusMin/Max bound a region's fuzz (spec 101 D2): small enough
	// that "Thornspire" names a place, not the whole map; large enough that a
	// myth's vague "at the forest's edge" scale can be honored.
	regionRadiusMin = 2
	regionRadiusMax = 24
	// GuardianRegionCharge is the canonize working's premium price (spec 101
	// D4, the charge-shape decision recorded above): 2 charges flat, the
	// "dearest miracle" band (guardian.time_snapped's price, miracles.go) —
	// mirrored into the tool package's Cost{Charges:2} declaration
	// (tool/registry.go), the send_vision/prophesy cross-reference-by-comment
	// precedent rather than a shared constant (canonize_region is not one of
	// the four FROZEN work_miracle kinds tool.miracleCosts prices).
	GuardianRegionCharge = 2
)

// canonizeFeatureKinds is the canonize working's feature vocabulary (spec 101
// D2: "place ONE feature of an EXISTING placeable kind... prefer existing
// kinds v1") — deliberately a NARROWER subset of BuildableStructureKinds
// (plans.go): fire (fuel-tracked) and chest (owner + Store) both carry
// per-agent-linked lifecycle a divine act has no natural owner for, so they
// are excluded here; shelter/oven/wall_plank/wall_stone/path carry no such
// entanglement and read plausibly as a guardian-raised landmark.
var canonizeFeatureKinds = []string{"shelter", "oven", "wall_plank", "wall_stone", "path"}

// CanonizeFeatureKinds returns a copy of the canonize working's feature
// vocabulary, in the order above — exported for internal/tool's drift
// cross-check (the DesignationKinds/BuildableStructureKinds pattern).
func CanonizeFeatureKinds() []string {
	out := make([]string, len(canonizeFeatureKinds))
	copy(out, canonizeFeatureKinds)
	return out
}

func canonizeFeatureKind(kind string) bool {
	for _, k := range canonizeFeatureKinds {
		if k == kind {
			return true
		}
	}
	return false
}

// Region is one villager-coined toponym made real (spec 101 FR-001,
// data-model): a durable, checkable claim on the map, cloning the
// spec-084 Designation shape. Status is always "active" — no terminal event
// exists in v1 (renames/decommission are future work, the spec's own edge
// case) — kept for 084-shape parity and forward compatibility rather than
// omitted outright.
type Region struct {
	ID         string `json:"id"`     // "reg-<placedTick>-<seq>" (nextPlanID shape, no RNG)
	X          int    `json:"x"`      // center
	Y          int    `json:"y"`      // center
	Radius     int    `json:"radius"` // tiles, regionRadiusMin..Max
	Name       string `json:"name"`   // the villager-coined toponym, 1..80 runes
	PlacedTick int64  `json:"placed_tick"`
	Status     string `json:"status"` // always "active" in v1 (see doc above)
	// PlacedSeq is the placing event's store seq, reducer-stamped (the
	// Designation.PlacedSeq / GuardianOrder.PlacedSeq precedent). Ignored on
	// the wire payload; omitempty keeps a pre-101 snapshot byte-identical.
	PlacedSeq int64 `json:"placed_seq,omitempty"`
}

// RegionNamedPayload is guardian.region_named's wire payload (spec 101
// FR-002): the region fields plus an OPTIONAL single feature placement
// (FeatureKind == "" means "name only, no feature") and Gratis — the
// operator-door escape hatch every miracle carries (D4's "gratis/operator
// door unchanged"); the canonize_region tool schema has NO gratis param
// (work_miracle's structural-absence guarantee), so Gratis is unreachable
// from guardian turn output by construction.
type RegionNamedPayload struct {
	ID          string `json:"id"`
	X           int    `json:"x"`
	Y           int    `json:"y"`
	Radius      int    `json:"radius"`
	Name        string `json:"name"`
	FeatureKind string `json:"feature_kind,omitempty"`
	FeatureX    int    `json:"feature_x,omitempty"`
	FeatureY    int    `json:"feature_y,omitempty"`
	Gratis      bool   `json:"gratis"`
}

// RegionByID returns the region named id, or nil — exported for
// internal/guardian (id resolution, prompt rendering) and internal/mind, the
// DesignationByID precedent.
func (s *State) RegionByID(id string) *Region { return s.regionByID(id) }

func (s *State) regionByID(id string) *Region {
	for i := range s.Regions {
		if s.Regions[i].ID == id {
			return &s.Regions[i]
		}
	}
	return nil
}

// regionAt returns the region whose circle contains (x,y), or nil. Overlap is
// refused at the door (validateRegionNamed), so at most one active region
// ever claims a given point — the first match is authoritative. Linear scan
// over the bounded region slice (the designationByID discipline).
func regionAt(s *State, x, y int) *Region {
	for i := range s.Regions {
		r := &s.Regions[i]
		dx, dy := x-r.X, y-r.Y
		if dx*dx+dy*dy <= r.Radius*r.Radius {
			return &s.Regions[i]
		}
	}
	return nil
}

// circlesOverlap reports whether two circles' interiors intersect — touching
// (distance exactly the sum of radii) is NOT overlap, so two named regions
// may sit edge-to-edge.
func circlesOverlap(x1, y1, r1, x2, y2, r2 int) bool {
	dx, dy := x1-x2, y1-y2
	rsum := r1 + r2
	return dx*dx+dy*dy < rsum*rsum
}

// applyRegion is the reducer arm for guardian.region_named (spec 101
// FR-001/002). Validate-not-clamp (the InjectSocial dry-run contract): every
// check — id, bounds, name, overlap, cap, and (when a feature rides along)
// its kind/site/containment/build-site — precedes the charge spend and the
// mutation, so a rejected canonization spends nothing and leaves no partial
// application.
func (s *State) applyRegion(e store.Event) error {
	if e.Type != "guardian.region_named" {
		return fmt.Errorf("apply %s: unknown region working type", e.Type)
	}
	var p RegionNamedPayload
	if err := json.Unmarshal(e.Payload, &p); err != nil {
		return fmt.Errorf("apply %s: %w", e.Type, err)
	}
	if err := s.validateRegionNamed(e.Type, &p); err != nil {
		return err
	}
	// The premium charge (D4): 2 flat unless gratis — the guardian.nudged
	// inline-spend shape (guardian.go), not the miracleCost table (this event
	// is not one of the four FROZEN work_miracle kinds that table prices).
	if !p.Gratis {
		if s.GuardianCharges < GuardianRegionCharge {
			return fmt.Errorf("apply %s: need %d charge(s), only %d banked", e.Type, GuardianRegionCharge, s.GuardianCharges)
		}
		s.GuardianCharges -= GuardianRegionCharge
	}
	r := Region{
		ID: p.ID, X: p.X, Y: p.Y, Radius: p.Radius, Name: strings.TrimSpace(p.Name),
		PlacedTick: e.Tick, Status: "active", PlacedSeq: e.Seq,
	}
	// prunePlanEntities is called for shape parity with the 084 family
	// (plans.go); every region reports active (see the Region doc above), so
	// nonActive is always 0 and this call is inert by construction — the cap
	// above is what actually bounds growth.
	s.Regions = prunePlanEntities(append(s.Regions, r), func(Region) bool { return true })
	if p.FeatureKind != "" {
		// Mirrors agent.built's structure construction (state.go) for a
		// guardian-placed feature: fresh, full-health when it's a wall; no
		// fuel/owner entanglement reaches here because canonizeFeatureKinds
		// excludes fire and chest (see the doc comment above).
		st := Structure{Kind: p.FeatureKind, X: p.FeatureX, Y: p.FeatureY}
		if isWall(p.FeatureKind) {
			st.HP = wallMaxHP(p.FeatureKind)
		}
		s.Structures = append(s.Structures, st)
	}
	return nil
}

// validateRegionNamed is the guardian.region_named door (spec 101 FR-001/002,
// validate-not-clamp): id discipline, radius/name bounds, world bounds,
// overlap-refusal (the spec's own edge case: a second christening of an
// overlapping region refuses), the active cap, and — when a feature rides
// along — its kind, world bounds, containment within the named region, and
// build-site validity (the existing entity/build placement rule, reused
// rather than re-derived).
func (s *State) validateRegionNamed(eventType string, p *RegionNamedPayload) error {
	if p.ID == "" {
		return fmt.Errorf("apply %s: empty region id", eventType)
	}
	for i := range s.Regions {
		if s.Regions[i].ID == p.ID {
			return fmt.Errorf("apply %s: duplicate region id %q", eventType, p.ID)
		}
	}
	if p.Radius < regionRadiusMin || p.Radius > regionRadiusMax {
		return fmt.Errorf("apply %s: radius %d outside %d..%d", eventType, p.Radius, regionRadiusMin, regionRadiusMax)
	}
	name := strings.TrimSpace(p.Name)
	if n := utf8.RuneCountInString(name); n == 0 || n > regionNameMaxRunes {
		return fmt.Errorf("apply %s: name length %d outside 1..%d runes", eventType, n, regionNameMaxRunes)
	}
	if s.m != nil && !s.m.InBounds(p.X, p.Y) {
		return fmt.Errorf("apply %s: center (%d,%d) lies outside the world (%dx%d)", eventType, p.X, p.Y, s.m.W, s.m.H)
	}
	for i := range s.Regions {
		r := &s.Regions[i]
		if circlesOverlap(p.X, p.Y, p.Radius, r.X, r.Y, r.Radius) {
			return fmt.Errorf("apply %s: overlaps the named region %q (%s)", eventType, r.Name, r.ID)
		}
	}
	if active := len(s.Regions); active >= GuardianRegionCap {
		return fmt.Errorf("apply %s: %d regions already named (cap %d)", eventType, active, GuardianRegionCap)
	}
	if p.FeatureKind != "" {
		if !canonizeFeatureKind(p.FeatureKind) {
			return fmt.Errorf("apply %s: %q is not a feature the canonize working can raise", eventType, p.FeatureKind)
		}
		if s.m != nil && !s.m.InBounds(p.FeatureX, p.FeatureY) {
			return fmt.Errorf("apply %s: feature site (%d,%d) lies outside the world", eventType, p.FeatureX, p.FeatureY)
		}
		if dx, dy := p.FeatureX-p.X, p.FeatureY-p.Y; dx*dx+dy*dy > p.Radius*p.Radius {
			return fmt.Errorf("apply %s: feature site (%d,%d) lies outside the named region", eventType, p.FeatureX, p.FeatureY)
		}
		if !buildSite(s.m, s, p.FeatureX, p.FeatureY) {
			return fmt.Errorf("apply %s: (%d,%d) is not a valid build site", eventType, p.FeatureX, p.FeatureY)
		}
	}
	return nil
}

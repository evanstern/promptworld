package sim

// The guardian's durable plan layer (spec 084): designations — event-sourced
// world plan artifacts with structural fulfillment predicates — and directives
// — hard, TTL-bounded villager bindings to a designation. Both entities clone
// the sim.GuardianOrder discipline verbatim (research R1): deterministic
// human-readable ids minted guardian-side with no RNG, one-way status
// transitions resolved at the reducer door, payload Status/PlacedSeq ignored
// and reducer-stamped, caps validated in the arm (validate-not-clamp), and an
// active+recent-32 retention prune.
//
// Door split (research R2, contracts/events.md): designation.placed/cancelled
// and directive.issued/cancelled are INJECTED through InjectSocial (whitelist
// + these validating arms); designation.fulfilled, directive.fulfilled, and
// directive.expired are EXECUTOR-emitted from stepEvents as pure functions of
// (state, tick) — the charge_regenerated pattern — so replay reproduces the
// whole lifecycle with no guardian running.

import (
	"encoding/json"
	"fmt"
	"unicode/utf8"

	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/target"
)

const (
	// GuardianDesignationCap bounds concurrent ACTIVE designations (spec 084
	// FR-002): a village-scale plan holds a bounded number of open claims.
	GuardianDesignationCap = 16
	// GuardianDirectiveCap bounds concurrent ACTIVE directives (FR-008) — the
	// GuardianPlayerOrderCap number, same rationale: a bounded attention
	// economy for hard villager bindings.
	GuardianDirectiveCap = 3
	// Directive TTLs reuse the standing-order bounds (GuardianOrderTTLMinDays/
	// MaxDays, guardian.go) — SHARED constants, never copied (FR-008).
)

// Designation kinds (data-model §1) — FROZEN serialized vocabulary (spec 052
// ruling 2): they land in recorded designation.placed payloads.
const (
	DesignationSettlementZone = "settlement_zone" // rect locus
	DesignationStructureSite  = "structure_site"  // point locus + required StructureKind
	DesignationWallLine       = "wall_line"       // axis-aligned line locus
)

// Designation size bounds (data-model §1, door-validated).
const (
	designationZoneMaxTiles  = 256 // rect area cap
	designationLineMaxTiles  = 32  // line length cap
	designationLabelMaxRunes = 80
	// Settlement-zone MinStructures bounds; the default (3) is applied at the
	// tool door (the monitor_and_act TTL-default shape) — the reducer only
	// validates the landed value.
	designationMinStructuresMin = 1
	designationMinStructuresMax = 12
)

// directiveTextMaxRunes caps the guardian's framing text (FR-008) — the same
// 400-rune band the standing-order Action carries.
const directiveTextMaxRunes = 400

// Designation is one event-sourced world plan artifact (spec 084 FR-002,
// data-model §2): a durable, checkable, villager-visible claim on the world.
// Loci are stored NORMALIZED (ints, the target.Address shape) — the payload is
// self-contained and replay never re-parses. Its lifecycle
// (active → fulfilled | cancelled) is driven entirely by recorded events, so
// it reconstructs identically through snapshots, restart, and replay.
type Designation struct {
	ID   string `json:"id"`   // "dsg-<placedTick>-<seq>" (nextOrderID shape, no RNG)
	Kind string `json:"kind"` // settlement_zone | structure_site | wall_line
	// (X,Y) is the point tile / the line's FIRST endpoint (author order
	// preserved) / the rect's normalized min corner — also the announcement
	// grant's anchor tile (data-model §1). (X2,Y2) is the line's second
	// endpoint / the rect's max corner; == (X,Y) for a point.
	X             int    `json:"x"`
	Y             int    `json:"y"`
	X2            int    `json:"x2,omitempty"`
	Y2            int    `json:"y2,omitempty"`
	StructureKind string `json:"structure_kind,omitempty"` // structure_site: required buildable kind; wall_line: optional wall-kind narrowing
	MinStructures int    `json:"min_structures,omitempty"` // settlement_zone only; landed value always 1..12 (default applied at the tool door)
	Label         string `json:"label,omitempty"`          // guardian's name for it, ≤80 runes
	PlacedTick    int64  `json:"placed_tick"`
	Status        string `json:"status"` // "active" | "fulfilled" | "cancelled" (one-way)
	// PlacedSeq is the placement event's store seq, stamped by the reducer at
	// apply time from the event envelope (the GuardianOrder.PlacedSeq /
	// spec-054 precedent). Ignored on the wire payload (like Status);
	// omitempty keeps injected payloads and pre-084 snapshots byte-identical.
	PlacedSeq int64 `json:"placed_seq,omitempty"`
}

// Directive is one hard, TTL-bounded binding of villagers to a designation
// (spec 084 FR-008, data-model §3). Targets are living villager indices
// RESOLVED AT ISSUE (ascending, unique) — the payload is self-contained and
// replay never re-resolves names. Its lifecycle
// (active → fulfilled | cancelled | expired) has three terminals, all
// one-way; exactly one ever lands (the transitionGuardianOrder race shape).
type Directive struct {
	ID            string `json:"id"`                // "dir-<issuedTick>-<seq>"
	DesignationID string `json:"designation_id"`    // must name an ACTIVE designation at issue
	Targets       []int  `json:"targets"`           // living villager indices at issue, ascending, non-empty
	Village       bool   `json:"village,omitempty"` // issued to "everyone" (provenance marker; Targets = all living at issue)
	Text          string `json:"text"`              // guardian framing, 1..400 runes
	IssuedTick    int64  `json:"issued_tick"`
	ExpiresTick   int64  `json:"expires_tick"`         // issued + ttl_days game days; bounds 1..7 (shared GuardianOrderTTL* constants)
	Status        string `json:"status"`               // "active" | "fulfilled" | "cancelled" | "expired" (one-way)
	PlacedSeq     int64  `json:"placed_seq,omitempty"` // reducer-stamped (the Designation.PlacedSeq contract)
}

// DirectiveIssuedPayload is directive.issued's wire mirror (spec 086
// FR-003, data-model §4): Directive's fields with named agent refs on the
// wire while the state entity above keeps bare ints — the R2 invariant (no
// AgentRef reachable from sim.State; state bytes and hashes unchanged).
// Same json tags, so legacy bare-int rows decode through the dual-shape
// unmarshal and the arm folds .IDs only.
type DirectiveIssuedPayload struct {
	ID            string     `json:"id"`
	DesignationID string     `json:"designation_id"`
	Targets       []AgentRef `json:"targets"`
	Village       bool       `json:"village,omitempty"`
	Text          string     `json:"text"`
	IssuedTick    int64      `json:"issued_tick"`
	ExpiresTick   int64      `json:"expires_tick"`
	Status        string     `json:"status"`
	PlacedSeq     int64      `json:"placed_seq,omitempty"`
}

// IssuedPayload builds the wire mirror from the entity (emission side).
func (d Directive) IssuedPayload() DirectiveIssuedPayload {
	return DirectiveIssuedPayload{
		ID: d.ID, DesignationID: d.DesignationID, Targets: Refs(d.Targets),
		Village: d.Village, Text: d.Text, IssuedTick: d.IssuedTick,
		ExpiresTick: d.ExpiresTick, Status: d.Status, PlacedSeq: d.PlacedSeq,
	}
}

// directive folds the mirror back into the int-typed state entity (arm side).
func (p DirectiveIssuedPayload) directive() Directive {
	return Directive{
		ID: p.ID, DesignationID: p.DesignationID, Targets: refIDs(p.Targets),
		Village: p.Village, Text: p.Text, IssuedTick: p.IssuedTick,
		ExpiresTick: p.ExpiresTick, Status: p.Status, PlacedSeq: p.PlacedSeq,
	}
}

// DirectiveFulfilledPayload — directive.fulfilled (contracts/events.md §1):
// THE TASK-118 faith-accounting seam. id dedupes, designation_id names what
// was achieved, targets who was bound (credit attribution), issued_tick with
// e.Tick gives the time-to-fulfil window. FROZEN recorded vocabulary (spec
// 086: targets carry named refs, tags unchanged).
type DirectiveFulfilledPayload struct {
	ID            string     `json:"id"`
	DesignationID string     `json:"designation_id"`
	Targets       []AgentRef `json:"targets"`
	IssuedTick    int64      `json:"issued_tick"`
}

// BuildableStructureKinds returns the structure kinds a villager can build —
// derived from the recipes table's build_* rows (the single source), in recipe
// order, so the designation door's structure_kind validation and the tool
// package's hand-carried Enum mirror (internal/tool, drift-tested from
// internal/guardian) can never diverge from what agent.built actually lands.
func BuildableStructureKinds() []string {
	var out []string
	for _, r := range recipes {
		if r.Structure != "" {
			out = append(out, r.Structure)
		}
	}
	return out
}

// buildableStructureKind reports whether kind names a real buildable structure
// (a recipes-table build_* row exists for it).
func buildableStructureKind(kind string) bool {
	_, ok := recipeFor("build_" + kind)
	return ok
}

// designationAddress reconstructs the target.Address a designation's stored,
// normalized locus denotes — kind decides the form (data-model §1), so the
// SAME Tiles() enumeration every other target consumer uses serves the
// fulfillment predicates and the map renderer (the one-parser law's
// enumeration half; no designation-side copy exists).
func designationAddress(d *Designation) target.Address {
	switch d.Kind {
	case DesignationSettlementZone:
		return target.Address{Form: target.FormRect, X: d.X, Y: d.Y, X2: d.X2, Y2: d.Y2}
	case DesignationWallLine:
		return target.Address{Form: target.FormLine, X: d.X, Y: d.Y, X2: d.X2, Y2: d.Y2}
	default: // structure_site
		return target.Address{Form: target.FormPoint, X: d.X, Y: d.Y}
	}
}

// DesignationTiles enumerates a designation's tiles deterministically —
// designationAddress + target.Tiles(), exported for the TUI map renderer
// (wall-line segments and zone perimeters render from the same enumeration
// the predicates check).
func DesignationTiles(d *Designation) []target.Tile {
	return designationAddress(d).Tiles()
}

// DesignationByID returns the designation named id, or nil — exported for
// internal/mind (spec 084 FR-011: the directive context block renders the
// bound designation's kind, site, and fulfillment requirement from state).
func (s *State) DesignationByID(id string) *Designation { return s.designationByID(id) }

// designationByID returns the designation named id, or nil. Linear over the
// bounded slice (active ≤ 16 + retained 32), matching every other entity scan.
func (s *State) designationByID(id string) *Designation {
	for i := range s.Designations {
		if s.Designations[i].ID == id {
			return &s.Designations[i]
		}
	}
	return nil
}

// designationFulfilled is the per-kind structural fulfillment predicate
// (spec 084 FR-005, data-model §6): a PURE state check — no clock, no RNG, no
// I/O, deterministic iteration (Tiles() order / s.Structures slice order) —
// evaluated IDENTICALLY by the executor sweep (to emit designation.fulfilled)
// and by the reducer arm (to validate before transitioning), so live and
// replay agree by construction.
func designationFulfilled(s *State, d *Designation) bool {
	switch d.Kind {
	case DesignationStructureSite:
		return s.Lookup().Structure(d.StructureKind, d.X, d.Y)
	case DesignationWallLine:
		for _, t := range DesignationTiles(d) {
			w := wallAt(s, t.X, t.Y)
			if w == nil {
				return false
			}
			if d.StructureKind != "" && w.Kind != d.StructureKind {
				return false
			}
		}
		return true
	case DesignationSettlementZone:
		n := 0
		for i := range s.Structures {
			st := &s.Structures[i]
			if st.X >= d.X && st.X <= d.X2 && st.Y >= d.Y && st.Y <= d.Y2 {
				n++
			}
		}
		return n >= d.MinStructures
	}
	return false
}

// directiveAddresses reports whether directive d binds villager idx — a
// resolved-index membership test (Village is a provenance marker only;
// "everyone" resolved to concrete indices at issue).
func directiveAddresses(d *Directive, idx int) bool {
	for _, t := range d.Targets {
		if t == idx {
			return true
		}
	}
	return false
}

// directiveTargetsAllDead reports whether no targeted villager remains alive —
// the directive.expired sweep's un-executable clause (a pure state check, no
// TTL wait; contracts/events.md).
func directiveTargetsAllDead(s *State, d *Directive) bool {
	for _, t := range d.Targets {
		if t >= 0 && t < len(s.Agents) && !s.Agents[t].Dead {
			return false
		}
	}
	return true
}

// applyPlan is the reducer arm family for the designation.*/directive.* event
// vocabulary (spec 084, contracts/events.md §1). Every arm validates rather
// than clamps — the InjectSocial dry-run runs these on a state copy, so an
// invalid placement/issue is rejected at the door and recorded events always
// re-apply cleanly at the same position in replay (the applyGuardian
// contract). The three executor-emitted types re-validate their emitting
// condition here, keeping the door authoritative even for sim-authored events.
func (s *State) applyPlan(e store.Event) error {
	switch e.Type {
	case "designation.placed":
		var d Designation
		if err := json.Unmarshal(e.Payload, &d); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		if err := s.validateDesignationPlaced(e.Type, &d); err != nil {
			return err
		}
		// Status is IGNORED on the payload — a designation always lands
		// active; PlacedSeq is reducer-stamped from the event's own store seq
		// (the guardian.order_placed shape: identical live and in replay; the
		// dry-run probe applies with Seq 0 and is discarded).
		d.Status = "active"
		d.PlacedSeq = e.Seq
		s.Designations = prunePlanEntities(append(s.Designations, d),
			func(x Designation) bool { return x.Status == "active" })
		// The announcement grant (FR-006, research R8): one place fact per
		// living villager at the designation's anchor tile — the spec-041
		// place-grant machinery, fanned out here in the arm (one event,
		// deterministic fan-out) rather than as N companion events. A
		// designation is not a pre-existing world thing — it becomes real BY
		// this event — so no ground-presence check applies. Map-less agents
		// skip; the reducer stays total (the place_revealed shape). Detail
		// stays 0 (PlaceFact.Detail is a kind-specific int64 scalar); the
		// prompt's landmark line renders kind+label from State.Designations.
		for i := range s.Agents {
			a := &s.Agents[i]
			if a.Dead || a.Map == nil {
				continue
			}
			a.Map.upsertFact(PlaceFact{Kind: "designation", X: d.X, Y: d.Y,
				Seen: e.Tick, Provenance: ProvenanceRevealed})
		}
	case "designation.cancelled":
		var p OrderIDPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		return s.transitionDesignation(e.Type, p.ID, "cancelled")
	case "designation.fulfilled":
		var p OrderIDPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		// Re-validate the predicate against current state before
		// transitioning (contracts/events.md — the door stays authoritative
		// on the structural fact, even for the executor's own emission).
		d := s.designationByID(p.ID)
		if d == nil {
			return fmt.Errorf("apply %s: unknown designation %q", e.Type, p.ID)
		}
		if !designationFulfilled(s, d) {
			return fmt.Errorf("apply %s: designation %q predicate does not hold", e.Type, p.ID)
		}
		return s.transitionDesignation(e.Type, p.ID, "fulfilled")
	case "directive.issued":
		// Spec 086: the wire carries the DirectiveIssuedPayload mirror (named
		// refs, dual-shape for legacy rows); the arm folds .IDs into the
		// int-typed entity — names never enter state (R2), and no name is
		// ever validated here (R3).
		var p DirectiveIssuedPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		d := p.directive()
		if err := s.validateDirectiveIssued(e.Type, &d); err != nil {
			return err
		}
		d.Status = "active"
		d.PlacedSeq = e.Seq
		s.Directives = prunePlanEntities(append(s.Directives, d),
			func(x Directive) bool { return x.Status == "active" })
	case "directive.cancelled":
		var p OrderIDPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		return s.transitionDirective(e.Type, p.ID, "cancelled")
	case "directive.fulfilled":
		var p DirectiveFulfilledPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		// Re-validate the emitting condition: the bound designation is
		// fulfilled (contracts/events.md).
		d := s.directiveByID(p.ID)
		if d == nil {
			return fmt.Errorf("apply %s: unknown directive %q", e.Type, p.ID)
		}
		dsg := s.designationByID(d.DesignationID)
		if dsg == nil || dsg.Status != "fulfilled" {
			return fmt.Errorf("apply %s: directive %q's designation is not fulfilled", e.Type, p.ID)
		}
		return s.transitionDirective(e.Type, p.ID, "fulfilled")
	case "directive.expired":
		var p OrderIDPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		// Re-validate the emitting disjunction: TTL elapsed OR no targeted
		// villager remains alive (contracts/events.md).
		d := s.directiveByID(p.ID)
		if d == nil {
			return fmt.Errorf("apply %s: unknown directive %q", e.Type, p.ID)
		}
		if e.Tick < d.ExpiresTick && !directiveTargetsAllDead(s, d) {
			return fmt.Errorf("apply %s: directive %q is neither past its TTL nor targetless", e.Type, p.ID)
		}
		return s.transitionDirective(e.Type, p.ID, "expired")
	}
	return nil
}

// validateDesignationPlaced is the designation.placed door (contracts §1,
// validate-not-clamp): id discipline, kind, locus form/normalization, bounds,
// size, occupancy fulfillability, per-kind parameters, label, and the active
// cap. The payload's normalized ints are re-checked here — the payload is
// self-contained, so the door re-derives nothing from strings.
func (s *State) validateDesignationPlaced(eventType string, d *Designation) error {
	if d.ID == "" {
		return fmt.Errorf("apply %s: empty designation id", eventType)
	}
	// Duplicate id in ANY status is rejected — ids are assigned once and
	// consumed designations are retained (the order-placed discipline).
	for i := range s.Designations {
		if s.Designations[i].ID == d.ID {
			return fmt.Errorf("apply %s: duplicate designation id %q", eventType, d.ID)
		}
	}
	if d.X < 0 || d.Y < 0 || d.X2 < 0 || d.Y2 < 0 {
		return fmt.Errorf("apply %s: negative locus coordinate", eventType)
	}
	switch d.Kind {
	case DesignationStructureSite:
		if d.X2 != d.X || d.Y2 != d.Y {
			return fmt.Errorf("apply %s: a structure site is one tile (form)", eventType)
		}
		if d.StructureKind == "" {
			return fmt.Errorf("apply %s: structure site needs a structure_kind", eventType)
		}
		if !buildableStructureKind(d.StructureKind) {
			return fmt.Errorf("apply %s: unknown structure kind %q", eventType, d.StructureKind)
		}
		if d.MinStructures != 0 {
			return fmt.Errorf("apply %s: min_structures applies to settlement zones only", eventType)
		}
		// Occupancy fulfillability: one structure per tile is a buildSite
		// invariant, so a tile already holding a DIFFERENT kind can never
		// fulfill; the SAME kind pre-existing is legal (the guardian may
		// consecrate what stands — the sweep fulfills at the next boundary).
		for i := range s.Structures {
			st := &s.Structures[i]
			if st.X == d.X && st.Y == d.Y && st.Kind != d.StructureKind {
				return fmt.Errorf("apply %s: a %s already stands at (%d,%d)", eventType, st.Kind, d.X, d.Y)
			}
		}
	case DesignationWallLine:
		if d.X != d.X2 && d.Y != d.Y2 {
			return fmt.Errorf("apply %s: a wall line is axis-aligned (form)", eventType)
		}
		if length := abs(d.X2-d.X) + abs(d.Y2-d.Y) + 1; length > designationLineMaxTiles {
			return fmt.Errorf("apply %s: wall line of %d tiles exceeds the %d-tile cap", eventType, length, designationLineMaxTiles)
		}
		if d.StructureKind != "" && !isWall(d.StructureKind) {
			return fmt.Errorf("apply %s: %q is not a wall kind", eventType, d.StructureKind)
		}
		if d.MinStructures != 0 {
			return fmt.Errorf("apply %s: min_structures applies to settlement zones only", eventType)
		}
		// Per-tile fulfillability: a non-wall structure on any line tile can
		// never be walled over (walls build on empty buildSite tiles only).
		for _, t := range DesignationTiles(d) {
			for i := range s.Structures {
				st := &s.Structures[i]
				if st.X == t.X && st.Y == t.Y && !isWall(st.Kind) {
					return fmt.Errorf("apply %s: a %s stands on the line at (%d,%d)", eventType, st.Kind, t.X, t.Y)
				}
			}
		}
	case DesignationSettlementZone:
		if d.X > d.X2 || d.Y > d.Y2 {
			return fmt.Errorf("apply %s: zone rect is not normalized (form)", eventType)
		}
		if area := (d.X2 - d.X + 1) * (d.Y2 - d.Y + 1); area > designationZoneMaxTiles {
			return fmt.Errorf("apply %s: zone of %d tiles exceeds the %d-tile cap", eventType, area, designationZoneMaxTiles)
		}
		if d.StructureKind != "" {
			return fmt.Errorf("apply %s: a settlement zone takes no structure_kind", eventType)
		}
		if d.MinStructures < designationMinStructuresMin || d.MinStructures > designationMinStructuresMax {
			return fmt.Errorf("apply %s: min_structures %d outside %d..%d", eventType,
				d.MinStructures, designationMinStructuresMin, designationMinStructuresMax)
		}
		// Zone rects may freely contain anything (spec edge cases) — no
		// occupancy check.
	default:
		return fmt.Errorf("apply %s: unknown designation kind %q", eventType, d.Kind)
	}
	// Bounds against the world map dims — every locus tile in-bounds is
	// equivalent to both corners/endpoints in-bounds for these axis-aligned
	// forms. A map-less State (bare test fixtures that never SetMap) skips,
	// the place_revealed s.m-guard shape: the reducer stays total.
	if s.m != nil {
		if !s.m.InBounds(d.X, d.Y) || !s.m.InBounds(d.X2, d.Y2) {
			return fmt.Errorf("apply %s: locus outside the world (%dx%d)", eventType, s.m.W, s.m.H)
		}
	}
	if utf8.RuneCountInString(d.Label) > designationLabelMaxRunes {
		return fmt.Errorf("apply %s: label over %d runes", eventType, designationLabelMaxRunes)
	}
	active := 0
	for i := range s.Designations {
		if s.Designations[i].Status == "active" {
			active++
		}
	}
	if active >= GuardianDesignationCap {
		return fmt.Errorf("apply %s: %d designations already active (cap %d)", eventType, active, GuardianDesignationCap)
	}
	return nil
}

// validateDirectiveIssued is the directive.issued door (contracts §1): id
// discipline, an ACTIVE bound designation, resolved living targets (ascending,
// unique, in-range), framing-text runes, the shared TTL bounds, and the
// active cap.
func (s *State) validateDirectiveIssued(eventType string, d *Directive) error {
	if d.ID == "" {
		return fmt.Errorf("apply %s: empty directive id", eventType)
	}
	for i := range s.Directives {
		if s.Directives[i].ID == d.ID {
			return fmt.Errorf("apply %s: duplicate directive id %q", eventType, d.ID)
		}
	}
	dsg := s.designationByID(d.DesignationID)
	if dsg == nil {
		return fmt.Errorf("apply %s: unknown designation %q", eventType, d.DesignationID)
	}
	if dsg.Status != "active" {
		return fmt.Errorf("apply %s: designation %q is not active (status %q)", eventType, d.DesignationID, dsg.Status)
	}
	if len(d.Targets) == 0 {
		return fmt.Errorf("apply %s: directive has no targets", eventType)
	}
	for i, t := range d.Targets {
		if t < 0 || t >= len(s.Agents) {
			return fmt.Errorf("apply %s: target index %d out of range", eventType, t)
		}
		if i > 0 && d.Targets[i-1] >= t {
			return fmt.Errorf("apply %s: targets not ascending-unique", eventType)
		}
		if s.Agents[t].Dead {
			return fmt.Errorf("apply %s: target %s is dead", eventType, s.Agents[t].Name)
		}
	}
	if n := utf8.RuneCountInString(d.Text); n == 0 || n > directiveTextMaxRunes {
		return fmt.Errorf("apply %s: text length %d outside 1..%d runes", eventType, n, directiveTextMaxRunes)
	}
	if ttl := d.ExpiresTick - d.IssuedTick; ttl < GuardianOrderTTLMinDays*ticksPerGameDay || ttl > GuardianOrderTTLMaxDays*ticksPerGameDay {
		return fmt.Errorf("apply %s: ttl %d ticks outside %d..%d game days", eventType, ttl, GuardianOrderTTLMinDays, GuardianOrderTTLMaxDays)
	}
	active := 0
	for i := range s.Directives {
		if s.Directives[i].Status == "active" {
			active++
		}
	}
	if active >= GuardianDirectiveCap {
		return fmt.Errorf("apply %s: %d directives already active (cap %d)", eventType, active, GuardianDirectiveCap)
	}
	return nil
}

// directiveByID returns the directive named id, or nil.
func (s *State) directiveByID(id string) *Directive {
	for i := range s.Directives {
		if s.Directives[i].ID == id {
			return &s.Directives[i]
		}
	}
	return nil
}

// transitionDesignation moves the designation named id from active to a
// terminal status (one-way). An unknown id or a non-active designation is
// rejected at the door — the transitionGuardianOrder race shape: exactly one
// terminal lands, and the loser hits a non-active entity and refuses.
func (s *State) transitionDesignation(eventType, id, to string) error {
	for i := range s.Designations {
		if s.Designations[i].ID != id {
			continue
		}
		if s.Designations[i].Status != "active" {
			return fmt.Errorf("apply %s: designation %q is not active (status %q)", eventType, id, s.Designations[i].Status)
		}
		s.Designations[i].Status = to
		return nil
	}
	return fmt.Errorf("apply %s: unknown designation %q", eventType, id)
}

// transitionDirective is transitionDesignation's directive twin.
func (s *State) transitionDirective(eventType, id, to string) error {
	for i := range s.Directives {
		if s.Directives[i].ID != id {
			continue
		}
		if s.Directives[i].Status != "active" {
			return fmt.Errorf("apply %s: directive %q is not active (status %q)", eventType, id, s.Directives[i].Status)
		}
		s.Directives[i].Status = to
		return nil
	}
	return fmt.Errorf("apply %s: unknown directive %q", eventType, id)
}

// prunePlanEntities retains every active entity plus the most recent
// guardianOrderRetain (32) non-active ones, dropping the oldest consumed
// entries first while preserving slice order — the pruneGuardianOrders
// algorithm generalized (research R1). Deterministic: a pure function of the
// append-ordered slice, so replay prunes identically.
func prunePlanEntities[T any](items []T, isActive func(T) bool) []T {
	nonActive := 0
	for i := range items {
		if !isActive(items[i]) {
			nonActive++
		}
	}
	drop := nonActive - guardianOrderRetain
	if drop <= 0 {
		return items
	}
	out := make([]T, 0, len(items)-drop)
	for i := range items {
		if !isActive(items[i]) && drop > 0 {
			drop--
			continue
		}
		out = append(out, items[i])
	}
	return out
}

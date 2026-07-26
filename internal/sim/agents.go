package sim

// Agent bodies for the executor layer (TASK-5): deterministic needs, intents,
// and inventories. All values are integers on a 0..1000 scale — integer math
// keeps decay byte-deterministic across platforms (no float rounding drift).

import (
	"fmt"

	"github.com/evanstern/promptworld/internal/tool"
)

// AgentCount is exported for packages that size per-agent tables.
const AgentCount = 8

const agentCount = AgentCount

// AgentNames is the canonical roster; internal/persona authors a nature for
// each. Order matters (agent index = position here).
var AgentNames = [agentCount]string{"Ash", "Birch", "Cedar", "Rowan", "Fern", "Hazel", "Oak", "Sage"}

// Needs are 0..1000; 0 is lethal territory, 1000 is full.
type Needs struct {
	Health int `json:"health"`
	Food   int `json:"food"`
	Rest   int `json:"rest"`
	Warmth int `json:"warmth"`
	Morale int `json:"morale"`
}

// Inventory is what an agent carries. Spec 012 (resources/food/crafting v2)
// widened it from the legacy {wood, food} pair to the full resource/item set;
// the legacy `Food int` field is gone (the format-version bump to 2 shields old
// v1 snapshots, which never decode under v2). All counts are ints; `Spears`
// holds remaining uses per carried spear, sorted ascending (hunts spend the
// most-worn first). omitempty keeps canonical bytes stable for empty kinds.
type Inventory struct {
	Wood         int   `json:"wood"`
	Stone        int   `json:"stone,omitempty"`
	Water        int   `json:"water,omitempty"`
	Planks       int   `json:"planks,omitempty"`
	RefinedStone int   `json:"refined_stone,omitempty"`
	FoodRaw      int   `json:"food_raw,omitempty"`
	FoodCooked   int   `json:"food_cooked,omitempty"`
	Meals        int   `json:"meals,omitempty"`
	Spears       []int `json:"spears,omitempty"` // remaining uses per spear, sorted ascending
	// Axes (spec 032 US2) mirror Spears exactly: remaining harvest uses per
	// carried axe, sorted ascending (harvests spend the most-worn first). A
	// carried axe triples chop/quarry yield. omitempty keeps pre-032 inventories
	// byte-identical.
	Axes []int `json:"axes,omitempty"`
}

// Intent is one multi-step goal being executed unattended: walk to
// (TargetX, TargetY), then perform Goal there for its duration. For chopping,
// the resource (the tree) is adjacent at (ResX, ResY) while the agent stands
// on the passable target tile.
type Intent struct {
	Goal      string `json:"goal"`
	TargetX   int    `json:"target_x"`
	TargetY   int    `json:"target_y"`
	ResX      int    `json:"res_x"`
	ResY      int    `json:"res_y"`
	WorkStart int64  `json:"work_start"` // 0 until work begins at the target
	// Kind/Qty (spec 013 R4) argue the storage goals (drop/pick_up/deposit/
	// withdraw): Kind is an inventory item key ("" = all kinds), Qty the amount
	// (0 = all of kind / as much as fits). Both omitempty keep pre-013 intents
	// and every non-storage intent byte-identical.
	Kind string `json:"kind,omitempty"`
	Qty  int    `json:"qty,omitempty"`
	// Reason (spec 019, R2) is the planner's free-text reason for this intent,
	// copied from the agent.intent_set payload by the reducer so it survives to
	// completion time — the executor bakes it into the memory's Why when the
	// intent completes. "" for reflex/executor-authored intents (never
	// fabricated). Lives on the intent, so it dies with it (cleared on
	// completion/abandonment). omitempty keeps reflex intents byte-stable.
	Reason string `json:"reason,omitempty"`
	// Needs-conditioned recovery (spec 064 R1): an OPTIONAL completion condition
	// — a need name (UntilNeed, a member of recoveryNeeds: warmth|rest|food) and
	// a threshold (UntilValue). When UntilNeed is set the executor HOLDS the
	// intent at its target and completes it on the need crossing UntilValue,
	// instead of the goal's default arrive-and-done (executeAtTarget). Absent
	// (UntilNeed == "") ≡ every pre-064 intent, byte-for-byte: both fields are
	// omitempty, so a conditionless intent marshals identically to before.
	// UntilValue is a need LEVEL (0-1000 scale), NOT a tick — KEEP under the
	// miracle rebase taxonomy (miracles.go), like NeedsAnchor's frozen levels.
	UntilNeed  string `json:"until_need,omitempty"`
	UntilValue int    `json:"until_value,omitempty"`
	// HoldRef (spec 064 R4) is the need level captured at the current hold anchor
	// (WorkStart, reused as the hold-since tick for a conditioned intent — a
	// no-work goal never otherwise sets it). The per-tick completion check
	// compares the live need against HoldRef over the recoveryStallTicks window:
	// no net gain across a whole window ⇒ the source is dead ⇒ abort
	// (agent.recovery_stalled). A need LEVEL, not a tick — KEEP on rebase.
	// omitempty keeps a non-holding intent byte-stable (0 = no anchor captured).
	HoldRef int `json:"hold_ref,omitempty"`
}

// intentLogCap bounds the recent-intent ring (spec 043 US1, data-model.md): 8
// records cover several game-hours of intents — more than the alternation
// window FR-003 needs — at fixed per-agent cost.
const intentLogCap = 8

// trajectoryWindowTicks is the span of game time (in ticks) over which a need's
// direction is measured (spec 043 US2, FR-004): the anchor snapshot rolls
// forward once this much game time has elapsed since it was taken, so a need's
// "rising/falling/steady" reflects movement over roughly the last window rather
// than instant-to-instant noise. Default 1800 (one planner cadence). Like
// contextBudgetTokens (internal/mind/context.go) it is a package const today
// with the design intent of a per-world tuning-manifest dial (TASK-107's
// const-fallback pattern — the manifest supplies the value when present, this
// const is the fallback).
const trajectoryWindowTicks = 1800

// IntentRecord is one entry in a villager's recent-intent ring (spec 043 US1,
// data-model.md). Goal is the intent's goal name; Source is the verbatim
// IntentSetPayload.Source ("planner" | "reflex" | "plan"); Reason is the stated
// reason when the source recorded one (planner/plan) and empty for reflex —
// never invented; Tick is when the intent landed. Outcome ("" while executing,
// then "done" | "failed" | "rejected" | "expired" | "stalled") and OutcomeTick
// are stamped by the closing lifecycle event ("stalled" is spec 064's
// needs-conditioned recovery abort). All-omitempty tail keeps records compact and
// pre-043-comparable in canonical bytes.
type IntentRecord struct {
	Goal        string `json:"goal"`
	Source      string `json:"source,omitempty"`
	Reason      string `json:"reason,omitempty"`
	Tick        int64  `json:"tick"`
	Outcome     string `json:"outcome,omitempty"`
	OutcomeTick int64  `json:"outcome_tick,omitempty"`
}

// appendIntent pushes a new record onto the ring, dropping the oldest when the
// ring is full. It re-slices into a fresh backing array on overflow so the
// canonical bytes reflect exactly the retained records (no aliased capacity)
// and replay and live paths produce identical state.
func (a *Agent) appendIntent(r IntentRecord) {
	a.IntentLog = append(a.IntentLog, r)
	if len(a.IntentLog) > intentLogCap {
		a.IntentLog = append([]IntentRecord(nil), a.IntentLog[len(a.IntentLog)-intentLogCap:]...)
	}
}

// stampIntentOutcome closes the newest still-open record (Outcome == "") with
// the given outcome. Used by the intent_done / build_failed arms, whose events
// clear the CURRENT intent — the newest open record. An override left an older
// record open behind the new one; that older record stays open (the
// open-then-superseded shape the alternation view preserves), so closing the
// newest is correct. A no-op when no record is open.
//
// It returns the closed record's Source ("planner" | "reflex" | "plan" |
// "meeting") and true, so the completion arm can arm the yield window on a
// non-reflex completion (spec 062 US1) from the SAME event-sourced record it
// closes — no source is carried on the bare intent itself. Returns ("", false)
// when no record was open (the source is unknowable, so the window never arms).
func (a *Agent) stampIntentOutcome(outcome string, tick int64) (string, bool) {
	for i := len(a.IntentLog) - 1; i >= 0; i-- {
		if a.IntentLog[i].Outcome == "" {
			a.IntentLog[i].Outcome = outcome
			a.IntentLog[i].OutcomeTick = tick
			return a.IntentLog[i].Source, true
		}
	}
	return "", false
}

// stampOrAppendExpired records a plan step's expiry (spec 043, plan end visible
// at next thought — FR-005): if an open record for that goal exists (the step
// had fired as a "plan"-source intent), it is closed "expired"; otherwise a
// closed record is appended (the step expired before ever firing, so it has no
// open record). Matching by goal keeps a concurrent non-plan open intent from
// being mis-stamped.
func (a *Agent) stampOrAppendExpired(goal string, tick int64) {
	for i := len(a.IntentLog) - 1; i >= 0; i-- {
		if a.IntentLog[i].Outcome == "" && a.IntentLog[i].Goal == goal {
			a.IntentLog[i].Outcome = "expired"
			a.IntentLog[i].OutcomeTick = tick
			return
		}
	}
	a.appendIntent(IntentRecord{Goal: goal, Source: "plan", Tick: tick, Outcome: "expired", OutcomeTick: tick})
}

type Agent struct {
	Name     string       `json:"name"`
	X        int          `json:"x"`
	Y        int          `json:"y"`
	Needs    Needs        `json:"needs"`
	Inv      Inventory    `json:"inv"`
	Asleep   bool         `json:"asleep"`
	Dead     bool         `json:"dead"`
	Intent   *Intent      `json:"intent,omitempty"`
	LastTalk int64        `json:"last_talk"`
	LastGive int64        `json:"last_give,omitempty"`
	Known    []KnownRumor `json:"known,omitempty"`
	// Memories accrete via agent.memory_added events (TASK-7); soul.md is a
	// rendered view of this list. Bounded later by TASK-9 consolidation.
	Memories []Memory `json:"memories,omitempty"`
	// IdleSince is the tick this agent last became idle/awake — reducer-
	// maintained so the reflex grace is a pure function of event history.
	IdleSince int64 `json:"idle_since"`
	// LastMindIntentDone (spec 062 US1, FR-003) is the tick the agent's most
	// recent NON-REFLEX intent (planner/plan/meeting = "intelligence")
	// completed — the yield-window anchor the reflex consults to defer its PREP
	// rungs (elapsed = tick-LastMindIntentDone; prep yields while that is under
	// prepYieldTicks). Written ONLY by the intent-completion reducer arm, and
	// ONLY when the completing intent's ring-record source is non-reflex: a
	// reflex completion never arms the window (instinct yielding to itself would
	// deadlock prep in a no-planner world — the sentinel then stays 0 forever,
	// so degraded mode never suppresses on this clause). Reducer-derived ⇒
	// replay-safe; rebase taxonomy SHIFT (a duration anchor, ONLY non-zero —
	// the Belief.Reinforced/NeedsAnchorTick shape). omitempty keeps a
	// never-mind-driven agent's canonical bytes identical to a pre-062 snapshot
	// (precedent: NeedsAnchorTick, LastGoalTick).
	LastMindIntentDone int64 `json:"last_mind_intent_done,omitempty"`
	// NearDeath latches the "nearly died" memory once per health collapse.
	NearDeath bool `json:"near_death,omitempty"`
	// Generation counts high-salience interrupts (TASK-32, FR-014): bumped
	// by the reducer on memories at/above GenerationBumpSalience. In-flight
	// thoughts snapshotted under an older generation are superseded at
	// landing. omitempty keeps pre-TASK-32 snapshots byte-stable.
	Generation int64 `json:"generation,omitempty"`
	// Plan is the pending guarded steps of a conditional plan (TASK-32
	// US4): the executor evaluates the head step each idle tick.
	Plan []PlanStep `json:"plan,omitempty"`
	// Nightly consolidation (TASK-9). Night/mark values of 0 mean "never" —
	// NightIndex is 1-based — so pre-TASK-9 snapshots stay correct.
	Beliefs               []Belief `json:"beliefs,omitempty"`
	Narrative             string   `json:"narrative,omitempty"`
	LastConsolidatedNight int64    `json:"last_consolidated_night,omitempty"`
	ConsolidatedUpTo      int64    `json:"consolidated_up_to,omitempty"`
	LastConsolidateMark   int64    `json:"last_consolidate_mark,omitempty"`
	// Hail (TASK-47) is the target-side pause: nil unless a talk_to landing
	// flagged this agent down. Pointer + omitempty so pre-feature snapshots
	// and un-hailed agents keep byte-identical canonical state (determinism
	// hash). Written only by the reducer.
	Hail *AgentHail `json:"hail,omitempty"`
	// LastGoal/LastGoalTick (TASK-56) remember the most recently pursued
	// objective so the villagers tab can show "what did they last do?" while
	// idle. Set by State.Apply on agent.intent_set alongside Intent and never
	// cleared — every intent-clearing path (intent_done, gru.attacked, hail
	// interrupts) leaves them untouched by construction, since "most recent
	// objective" is exactly "last goal set". omitempty keeps pre-feature
	// snapshots byte-stable (precedent: Generation, Plan, Hail).
	LastGoal     string `json:"last_goal,omitempty"`
	LastGoalTick int64  `json:"last_goal_tick,omitempty"`
	// IntentLog (spec 043 US1) is the villager's recent-intent ring: the last
	// few intents it pursued, each with its source and outcome, maintained
	// entirely by the intent-lifecycle reducer arms (state.go). Unlike LastGoal
	// (a single slot), the ring preserves ORDER and per-intent source so the
	// decision prompt can show alternation between goals (FR-003) and name where
	// each intent came from (FR-002). Reducer-derived ⇒ replay-safe by
	// construction; capacity intentLogCap. omitempty keeps a never-acted agent's
	// canonical bytes identical to a pre-043 snapshot (precedent: LastGoal).
	IntentLog []IntentRecord `json:"intent_log,omitempty"`
	// Journal (spec 019, US3) is the agent's self-authored notebook — durable
	// world state mutated ONLY by the two journal.* reducer arms. A POINTER with
	// omitempty (the Hail precedent) so an agent that never journals stays
	// byte-identical to a pre-019 snapshot: a value struct would always serialize
	// "journal":{} (encoding/json omitempty is a no-op on non-pointer structs),
	// breaking the pre-019 round-trip the feature requires (deviation from
	// data-model.md §5's value type, recorded for the planning tier). Written
	// only by the reducer.
	Journal *Journal `json:"journal,omitempty"`
	// Map (spec 041) is the agent's private spatial knowledge — the mental
	// map gating target resolution and prompt rendering (mentalmap.go). A
	// POINTER with omitempty (the Journal/Hail precedent) so a pre-041
	// snapshot (field absent) round-trips byte-identically. Created exactly
	// once, at genesis (NewState) or migration (TransformV3State) — never
	// lazily by the reducer, so a map-less agent (dead at migration time)
	// stays map-less on replay. Facts mutated only by knowledge-event reducer
	// arms; explored bits by the derived markExplored bookkeeping (D2).
	Map *MentalMap `json:"map,omitempty"`
	// SitVec/SitVecModel/SitVecTick (spec 042) are the agent's rolling
	// situation (query) vector: set by the reducer from a recorded
	// agent.situation_embedded companion (copy-verbatim — the sim never
	// computes an embedding), refreshed by the mind-side embedder at planner
	// cadence while it runs. Absent (nil Vec) ⇒ selection falls back to the
	// legacy ranking. All three omitempty keep pre-042 snapshots byte-stable
	// (precedent: Generation, Plan, Hail).
	SitVec      []float32 `json:"sit_vec,omitempty"`
	SitVecModel string    `json:"sit_vec_model,omitempty"`
	SitVecTick  int64     `json:"sit_vec_tick,omitempty"`
	// NeedsAnchor/NeedsAnchorTick (spec 043 US2) are the trajectory window's
	// edge snapshot: the needs levels at the last window edge and the tick that
	// edge was taken. Direction per need at render time is sign(current − anchor)
	// with a deadband (internal/mind/context.go), so the prompt can say "warmth
	// 45 and falling" — the cheapest form of foresight (FR-004). Refreshed by the
	// agent.needs_changed reducer arm once a full trajectoryWindowTicks has
	// elapsed since the anchor was taken; NeedsAnchorTick == 0 (nil anchor) is the
	// unset sentinel — the first window, before any anchor exists, renders steady
	// (edge case 1). A POINTER with omitempty (the Journal/Hail/Map precedent, NOT
	// data-model.md's value type — deviation recorded for the planning tier): a
	// value Needs would always serialize "needs_anchor":{...} (encoding/json
	// omitempty is a no-op on a non-pointer struct), breaking the pre-043
	// round-trip byte-identity the codebase requires. Reducer-derived ⇒ replay-
	// safe by construction; NeedsAnchorTick is a duration anchor, SHIFT under a
	// time snap (see rebaseTicks doctrine, miracles.go).
	NeedsAnchor     *Needs `json:"needs_anchor,omitempty"`
	NeedsAnchorTick int64  `json:"needs_anchor_tick,omitempty"`
}

// AgentHail is the courtesy pause a talk_to landing lays on its target: who
// hailed it (By) and the tick the pause lifts (Until). Denominated in game
// ticks so wall-speed changes never stretch or shrink the window.
type AgentHail struct {
	By    int   `json:"by"`
	Until int64 `json:"until"`
}

// MemoryPlace situates a memory: where the agent stood when it formed (spec
// 019, US1). Carried as a pointer (*MemoryPlace) everywhere it appears so
// absence is nil — never a fake (0,0) origin. Desc is a deterministic
// terrain/feature description baked at emission (describePlace); "" = coords
// alone situate the memory.
type MemoryPlace struct {
	X    int    `json:"x"`
	Y    int    `json:"y"`
	Desc string `json:"desc,omitempty"`
}

// Memory is one episodic record; salience 1..10 weights the working-memory
// window. Subject/Tone (TASK-8) mark gossip-worthy memories about another
// agent — the seeds rumors are born from (−1 subject = purely personal).
//
// Where/Why/Conv (spec 019) are the situated context, copied verbatim from the
// emitting event's payload by the agent.memory_added reducer arm — never
// re-derived at render or replay. All three omitempty so a pre-019 Memory
// (fields absent) marshals byte-identically to today (FR-014, SC-007).
//
// Origin (spec 030) is the emission-stamped provenance class — the model-free
// signal the belief validator reads to decide whether a memory is direct
// perception (see DirectPerception). Closed vocabulary (OriginAction..
// OriginDigest); absent (pre-030, "") classifies as secondhand, the
// conservative direction. omitempty keeps every pre-030 Memory byte-identical.
//
// Seq/Vec/VecModel (spec 042) are the embedding-retrieval identity and vector:
// Seq is the store seq of the emitting agent.memory_added event, stamped by
// the reducer at apply time — the memory's stable identity for companion
// events (research D4; unique where (agent, tick) is not). Vec/VecModel are
// attached by the reducer from a recorded agent.memory_embedded companion,
// copy-verbatim (the sim never computes an embedding) and set together or not
// at all; nil Vec = vectorless (neutral relevance, FR-010). All three
// omitempty keep every pre-042 Memory byte-identical.
type Memory struct {
	Text     string       `json:"text"`
	Salience int          `json:"salience"`
	Tick     int64        `json:"tick"`
	Subject  int          `json:"subject"`
	Tone     int          `json:"tone,omitempty"`
	Where    *MemoryPlace `json:"where,omitempty"`     // location at emission (nil = none)
	Why      string       `json:"why,omitempty"`       // driving intent reason, verbatim ("" = none)
	Conv     int64        `json:"conv,omitempty"`      // conversation ref (founding-talk tick; 0 = none)
	Origin   string       `json:"origin,omitempty"`    // spec 030: provenance class stamped at emission
	Seq      int64        `json:"seq,omitempty"`       // spec 042: emitting event's store seq (0 = pre-042)
	Vec      []float32    `json:"vec,omitempty"`       // spec 042: recorded embedding (nil = vectorless)
	VecModel string       `json:"vec_model,omitempty"` // spec 042: producing model identity (FR-009)
}

// Structure is player-visible built stuff; the map itself never contains
// structures ([[worldmap]] cold start) — they exist only as event-sourced state.
//
// FuelUntil (spec 012) applies to fires only: a fire is lit iff tick <
// FuelUntil. It is set at build (build tick + fire's initial burn window) and
// pushed forward by refuel, capped at now + fireFuelCap. Lit-ness is always
// derived, never stored as a flag. omitempty keeps shelter/oven and pre-012
// snapshots byte-identical. NOTE: warmth/burnout behavior is NOT yet wired to
// FuelUntil — that lands in Phase 4 (T019).
//
// Owner and Store (spec 013, research R8) apply to chests only: a chest rides
// the structure lifecycle rather than a parallel entity. Owner is the builder's
// agent index (permanent — no transfer/inheritance in v1); its zero-value
// round-trips unambiguously to agent 0 because every chest has an owner and
// non-chests never read the field. Store is the chest's contents, capped at
// chestCap via the same derived bulk() used for agents — chests preserve food
// indefinitely, so it needs no batches. Both omitempty keep non-chest and
// pre-013 snapshots byte-identical.
// HP (spec 032, research R1) applies to walls only: a standing wall's current
// health, 1..wallMaxHP(kind). Max HP is derived from the kind (never stored),
// same doctrine as fire lit-ness from FuelUntil. A standing wall always has
// HP ≥ 1 — the reducer removes the structure in the same application that would
// take it to ≤ 0, so hp never serializes as 0. omitempty keeps non-wall and
// pre-032 snapshots byte-identical.
type Structure struct {
	Kind      string     `json:"kind"` // "fire" | "shelter" | "oven" | "chest" | "wall_plank" | "wall_stone" | "path" | "grave" (spec 044 US4, reducer-placed only — never player-built)
	X         int        `json:"x"`
	Y         int        `json:"y"`
	FuelUntil int64      `json:"fuel_until,omitempty"` // fires only
	Owner     int        `json:"owner,omitempty"`      // chests only: builder agent index, permanent
	Store     *Inventory `json:"store,omitempty"`      // chests only: contents (no rot inside)
	HP        int        `json:"hp,omitempty"`         // walls only: current health, 1..wallMaxHP(kind)
}

// FoodBatch is one drop of food on the ground with its own spoilage deadline —
// rot is per-drop (spec 013 US5), so ground food is batch-tracked (chests
// preserve food and need no batches). Kind ∈ food_raw|food_cooked|meals.
type FoodBatch struct {
	Kind    string `json:"kind"`
	N       int    `json:"n"`
	SpoilAt int64  `json:"spoil_at"` // drop/death tick + rotWindowTicks
}

// Pile is the per-tile commons of dropped/spilled goods (spec 013 US2,
// research R1) — event-sourced overlay state like Quarried, never a tile
// mutation. Non-food is flat counts (it never decays); food is batch-tracked
// in drop order (batches with identical (Kind, SpoilAt) merge). Spears carry
// their remaining uses, sorted ascending (most-worn moves first). One pile per
// tile is a reducer invariant; a pile drained to nothing is removed in the same
// reducer application. omitempty keeps the canonical bytes stable for empty
// kinds.
type Pile struct {
	X            int         `json:"x"`
	Y            int         `json:"y"`
	Wood         int         `json:"wood,omitempty"`
	Stone        int         `json:"stone,omitempty"`
	Water        int         `json:"water,omitempty"`
	Planks       int         `json:"planks,omitempty"`
	RefinedStone int         `json:"refined_stone,omitempty"`
	Spears       []int       `json:"spears,omitempty"` // remaining uses, sorted ascending
	Axes         []int       `json:"axes,omitempty"`   // remaining uses, sorted ascending (spec 032 US2)
	Food         []FoodBatch `json:"food,omitempty"`   // drop order; same (Kind,SpoilAt) merges
}

// empty reports whether a pile holds nothing — the reducer removes such a pile
// in the same application that drains it (one pile per tile, zero-content piles
// removed; data-model.md).
func (p *Pile) empty() bool {
	return p.Wood == 0 && p.Stone == 0 && p.Water == 0 && p.Planks == 0 &&
		p.RefinedStone == 0 && len(p.Spears) == 0 && len(p.Axes) == 0 && len(p.Food) == 0
}

// addFood merges n of a food kind into the pile: an existing batch with the
// identical (Kind, SpoilAt) absorbs the count, else a new batch appends in drop
// order (data-model.md: "same (Kind,SpoilAt) merges"). A non-positive count is
// a no-op.
func (p *Pile) addFood(kind string, n int, spoilAt int64) {
	if n <= 0 {
		return
	}
	for i := range p.Food {
		if p.Food[i].Kind == kind && p.Food[i].SpoilAt == spoilAt {
			p.Food[i].N += n
			return
		}
	}
	p.Food = append(p.Food, FoodBatch{Kind: kind, N: n, SpoilAt: spoilAt})
}

// canonicalKinds is the fixed iteration order for "all kinds" storage transfers
// (data-model.md): the Inventory field order. Determinism depends on it — a
// Kind-empty pick_up/withdraw walks these in this exact order (spec 013 US2).
var canonicalKinds = []string{
	"wood", "stone", "water", "planks", "refined_stone",
	"food_raw", "food_cooked", "meals", "spears", "axes",
}

// isFoodKind reports whether a kind is one of the batch-tracked food forms
// (the only kinds that rot in ground piles).
func isFoodKind(kind string) bool {
	return kind == "food_raw" || kind == "food_cooked" || kind == "meals"
}

// foodKinds is the fixed iteration order the rot sweep walks each pile with
// (spec 013 US5, T032): the food subset of canonicalKinds, in canonical field
// order. Determinism depends on it — a sweep emits at most one sim.food_rotted
// per (pile, kind), and the kinds are always visited in this exact order.
var foodKinds = []string{"food_raw", "food_cooked", "meals"}

// carriedCount is how many units of a kind an agent carries: spears counted
// (durability lives in the slice), every other kind its flat inventory field.
func carriedCount(inv Inventory, kind string) int {
	switch kind {
	case "spears":
		return len(inv.Spears)
	case "axes":
		return len(inv.Axes)
	}
	return invField(inv, kind)
}

// avail is how many units of a kind the pile holds — food summed across
// batches, spears counted, non-food the flat field. The executor reads it to
// size a pick_up; the reducer to clamp defensively (staying total).
func (p *Pile) avail(kind string) int {
	switch kind {
	case "wood":
		return p.Wood
	case "stone":
		return p.Stone
	case "water":
		return p.Water
	case "planks":
		return p.Planks
	case "refined_stone":
		return p.RefinedStone
	case "spears":
		return len(p.Spears)
	case "axes":
		return len(p.Axes)
	case "food_raw", "food_cooked", "meals":
		n := 0
		for _, b := range p.Food {
			if b.Kind == kind {
				n += b.N
			}
		}
		return n
	}
	return 0
}

// addNonFood adds n units of a flat (non-food, non-spear) kind. Food rides
// addFood (batches + rot deadlines); spears carry durabilities and are
// appended directly by the caller. A non-positive count is a no-op.
func (p *Pile) addNonFood(kind string, n int) {
	if n <= 0 {
		return
	}
	switch kind {
	case "wood":
		p.Wood += n
	case "stone":
		p.Stone += n
	case "water":
		p.Water += n
	case "planks":
		p.Planks += n
	case "refined_stone":
		p.RefinedStone += n
	}
}

// takeNonFood removes up to n of a flat kind, returning the actual amount
// removed (clamped to what the pile holds — the reducer stays total).
func (p *Pile) takeNonFood(kind string, n int) int {
	if a := p.avail(kind); n > a {
		n = a
	}
	if n <= 0 {
		return 0
	}
	switch kind {
	case "wood":
		p.Wood -= n
	case "stone":
		p.Stone -= n
	case "water":
		p.Water -= n
	case "planks":
		p.Planks -= n
	case "refined_stone":
		p.RefinedStone -= n
	}
	return n
}

// takeFood removes up to n units of a food kind from the OLDEST matching
// batches first (drop order = creation order = oldest), returning the actual
// amount removed. Emptied batches are compacted out, preserving drop order.
func (p *Pile) takeFood(kind string, n int) int {
	if n <= 0 {
		return 0
	}
	taken := 0
	out := p.Food[:0]
	for _, b := range p.Food {
		if b.Kind == kind && taken < n {
			t := b.N
			if t > n-taken {
				t = n - taken
			}
			b.N -= t
			taken += t
		}
		if b.N > 0 {
			out = append(out, b)
		}
	}
	if len(out) == 0 {
		p.Food = nil
	} else {
		p.Food = out
	}
	return taken
}

// takeSpoiled removes up to n units of a food kind from the batches already
// spoiled at tick (SpoilAt <= tick), oldest matching batch first (drop order),
// returning the actual amount removed. Emptied batches are compacted out,
// preserving drop order. The rot sweep's reducer primitive (spec 013 US5,
// T032): it mirrors takeFood but only ever drains batches whose deadline has
// arrived, so a fresh batch dropped this same tick (spoil_at far in the future)
// is never touched, and it stays total (clamped to the spoiled units present —
// a same-tick pickup that applied first leaves only the remainder).
func (p *Pile) takeSpoiled(kind string, n int, tick int64) int {
	if n <= 0 {
		return 0
	}
	taken := 0
	out := p.Food[:0]
	for _, b := range p.Food {
		if b.Kind == kind && b.SpoilAt <= tick && taken < n {
			t := b.N
			if t > n-taken {
				t = n - taken
			}
			b.N -= t
			taken += t
		}
		if b.N > 0 {
			out = append(out, b)
		}
	}
	if len(out) == 0 {
		p.Food = nil
	} else {
		p.Food = out
	}
	return taken
}

// takeSpears removes the n most-worn spears (front of the ascending-sorted
// slice) and returns their durabilities; the pile stays sorted ascending.
func (p *Pile) takeSpears(n int) []int {
	if n > len(p.Spears) {
		n = len(p.Spears)
	}
	if n <= 0 {
		return nil
	}
	taken := append([]int(nil), p.Spears[:n]...)
	rest := append([]int(nil), p.Spears[n:]...)
	if len(rest) == 0 {
		p.Spears = nil
	} else {
		p.Spears = rest
	}
	return taken
}

// takeAxes removes the n most-worn axes (front of the ascending-sorted slice)
// and returns their remaining uses; the pile stays sorted ascending. The exact
// takeSpears clone (spec 032 US2).
func (p *Pile) takeAxes(n int) []int {
	if n > len(p.Axes) {
		n = len(p.Axes)
	}
	if n <= 0 {
		return nil
	}
	taken := append([]int(nil), p.Axes[:n]...)
	rest := append([]int(nil), p.Axes[n:]...)
	if len(rest) == 0 {
		p.Axes = nil
	} else {
		p.Axes = rest
	}
	return taken
}

// Harvest marks a foraged tile regrowing at Regrow.
type Harvest struct {
	X      int   `json:"x"`
	Y      int   `json:"y"`
	Regrow int64 `json:"regrow"`
}

// DenUse marks a hunted den not huntable again until Ready.
type DenUse struct {
	X     int   `json:"x"`
	Y     int   `json:"y"`
	Ready int64 `json:"ready"`
}

type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// --- executor tuning (game-minutes are the decay heartbeat) ---

const (
	// Per-game-minute needs deltas.
	foodDecay      = 1 // full → empty in ~16.6 game-hours
	restDecayAwake = 1
	restRegenSleep = 4 // full recharge in ~4 game-hours
	warmthLossCold = 4 // night, outdoors, no fire: full → 0 in ~4 game-hours
	warmthGainFire = 6
	warmthGainDay  = 2
	healthLoss     = 3 // per minute while starving or freezing (~5.5h to die)
	healthRegen    = 1 // fed and rested

	// Thresholds the reflex policy keys on.
	hungryAt = 350
	tiredAt  = 250

	// Action durations in ticks (game seconds).
	forageTicks       = 120
	chopTicks         = 300
	buildFireTicks    = 600
	buildShelterTicks = 1200
	huntTicks         = 900

	// Yields and costs. chopWood (spec 012's flat 2) is deleted by spec 032 T014
	// — chop yield is now chopYieldBare/chopYieldAxe (agents.go spec-032 block).
	fireWoodCost    = 2
	shelterWoodCost = 5

	// Cadences and ranges.
	moveEveryTicks = 5 // 12 tiles per game-minute
	fireWarmRadius = 2 // Manhattan
	// TASK-7: the reflex is the fallback mind — it only acts on agents idle
	// past this grace, leaving room for planner injections.
	reflexGraceTicks = 120 // 2 game-minutes
	// The mind driver's per-agent baseline cadence is spec-048-promoted: the
	// default lives in tuning.go as defaultPlannerCadenceTicks and reads go
	// through State.PlannerCadence() (nil-safe accessor).
	witnessRadius    = 8
	nearDeathBelow   = 200
	nearDeathResetAt = 400
	coldNightBelow   = 350
	forageRegrowSec  = 12 * 3600
	denCooldownSec   = 6 * 3600
	talkCooldownSec  = 2 * 3600
	talkMoraleBonus  = 50
)

// --- spec 062 (instinct yields to intelligence): reflex/planner arbitration ---
//
// The doctrine home for the PREP-gate constants (FR-006): named, single home,
// dial-READY but NOT tuning.json dials (earned, not speculative). The PREP
// rungs of decideIntent yield — do not fire — while either clause holds:
//
//	(a) yield window: a recent non-reflex intent completion (LastMindIntentDone)
//	    is under prepYieldTicks old — instinct defers to the mind's own follow-up;
//	(b) danger band: any need is below its band — the survival rung for that need
//	    owns the agent, so prep must not counter-schedule.
const (
	// Danger bands (R3): a need "in danger" == "a survival rung would or will
	// imminently act". Anchored EXACTLY at the existing survival-rung triggers —
	// no padding (an evidenced pad would be flagged): food at the hunger trigger,
	// warmth at the freezing-night warmth-effect threshold, rest at the nap
	// trigger. Because those same thresholds gate the SURVIVAL rungs that run
	// before PREP, the food/rest bands are largely subsumed (survival returns
	// first); the warmth band is the one that bites by day — it is what suppresses
	// prep for a villager recovering AT a fire (warmAt, so the day-warmth rung is
	// skipped) whose warmth need is still low. Dial-ready, not dialed.
	dangerFoodBelow   = hungryAt       // 350: food danger == hungry (eat/get-food trigger)
	dangerWarmthBelow = coldNightBelow // 350: warmth danger == freezing-night threshold
	dangerRestBelow   = tiredAt        // 250: rest danger == daytime nap trigger

	// prepYieldTicks (R4): the post-intelligence quiet period for prep instinct —
	// one default planner cadence, so the mind gets one beat to follow up its own
	// completed intent before instinct resumes prep. Deliberately THIS CONSTANT,
	// not the tuned planner-cadence dial (PlannerCadence()): the window is
	// arbitration DOCTRINE, not scheduling — a cadence-tuned world must not
	// silently stretch instinct's deference. Dial-ready, not dialed.
	prepYieldTicks = 1800
)

// DangerRestBelow exports dangerRestBelow (the 062 danger band above) for
// internal/tui's map condition overlays (spec 060 US2, needs-critical):
// Health and Food/Warmth already have an exported equivalent for this
// purpose (guardian.go's SurvivalNearDeathBelow / SurvivalStarvingRearm /
// SurvivalFreezingRearm — rest has no survival-watch kind of its own to ride
// along with, so this is the one new export the feature needs; it aliases
// the same existing constant rather than naming a new number, keeping sim
// the single source of the threshold.
const DangerRestBelow = dangerRestBelow

// --- spec 064 (needs-conditioned recovery): warm_up doctrine -----------------
//
// The doctrine home for the recovery-completion constants (FR-007): named,
// single place, promoted-dial-READY but NOT tuning.json dials (earned, not
// speculative — the spec-062 posture).
const (
	// warmthRecoverTo (R3) is the default warmth threshold a warm_up recovery
	// loiters to when the planner names none, and the ONLY threshold the reflex
	// warmth rungs use. 800 on the 0-1000 needs scale: a healthy 450-point margin
	// above dangerWarmthBelow (350), so a villager released from a warm_up does
	// not immediately re-enter the warmth danger band, yet well under the 1000
	// cap so a live fire (warmthGainFire +6/min) reaches it in bounded loiter
	// time (~4500 ticks from the freezing band). Chosen against the needs scale;
	// matches the spike's world-01 example (800). Dial-ready, not dialed.
	warmthRecoverTo = 800
	// warmthRecoverFloor is the clamp floor for a planner-supplied until_warmth
	// (edge case: "above the danger band floor, at or below the need cap"). A
	// threshold at or below the danger band would let a recovery complete while
	// still in danger — meaningless — so an out-of-range request clamps into
	// [warmthRecoverFloor, needMax]. Set at the warmth danger band itself (the
	// 062 one home): a request must at least clear danger. Clamp-with-notice
	// (spec 058 posture) — the sim clamps authoritatively (resolveGoal), never
	// rejects (ClampWarmUp).
	warmthRecoverFloor = dangerWarmthBelow // 350
	// needMax is the shared upper bound of every need on the 0-1000 scale
	// (decayNeeds clamps to 1000). Named here so the recovery clamp and the
	// completion check read one ceiling rather than a bare literal.
	needMax = 1000

	// exposureWakeBelow (US4/FR-006) is the warmth level at which a SLEEPER wakes
	// to a cold emergency (audit Gap C). It mirrors the hunger-emergency wake
	// (Food < 150, executor.go wakeReason) EXACTLY in shape AND magnitude: an
	// EMERGENCY floor well BELOW the warmth danger band (dangerWarmthBelow 350),
	// not the danger band itself. That distinction is load-bearing — the hunger
	// wake deliberately fires at an emergency (150), NOT at the hungry danger band
	// (350), so a sleeper isn't roused every night merely for being "hungry"/
	// "cold"; only a genuine emergency wakes them. A NEW named constant (FR-007
	// permits one for the wake band): the 062/059 warmth constants — coldNightBelow/
	// dangerWarmthBelow (350) and SurvivalFreezingAt (0) — don't fit here (350 is
	// the routine-dip danger band, 0 is too late to act), so this reuses the
	// hunger-emergency 150 for parity instead. At 150 a night sleeper still has
	// ~37 game-minutes of runway before warmth hits 0 and exposure health-drain
	// begins — ample to reach or build a fire. Waking at the 350 danger band
	// instead roused sleepers on routine cold dips and desynced the degraded-mode
	// forage rotation (a survival regression on the seed-101 knife-edge); this
	// emergency floor fires only for the true exposure spiral (Oak's 636→0),
	// leaving cozy and routine-cold sleepers undisturbed. Dial-ready, not dialed.
	exposureWakeBelow = 150

	// recoveryStallTicks (R4) is the abort window for a conditioned hold: if the
	// need shows NO net gain across this whole span while holding at the target,
	// the source is judged dead (fire burned out, displaced, or a threshold the
	// source can't reach) and the intent aborts with a distinct outcome
	// (agent.recovery_stalled). 300 ticks == 5 game-minutes == 5 needs
	// heartbeats (the heartbeat is per-game-minute, %60): comfortably longer than
	// one heartbeat so a hold never false-aborts on the flat ticks BETWEEN beats
	// (R4: "must not false-abort on the first flat tick"), yet short enough that
	// a dead night fire — warmth falling warmthLossCold(4)/min — is caught within
	// a few minutes. At a live fire warmth climbs +30 across the window (clear net
	// gain), so a recovering hold re-anchors and never aborts. Dial-ready.
	recoveryStallTicks = 300
)

// recoveryNeeds is the closed set of needs a completion condition may name
// (spec 064 R1): the door validates UntilNeed against it so a malformed
// condition can never reach the per-tick check. warmth is the evidenced
// consumer (warm_up); rest and food make the mechanism generic (US2) — a
// second consumer reuses the SAME fields and check with no parallel plumbing.
var recoveryNeeds = map[string]bool{"warmth": true, "rest": true, "food": true}

// isRecoveryNeed reports whether name is a valid completion-condition need
// (the closed-set door, R1). "" is not a need — a conditionless intent is the
// arrive-and-done default, checked by the UntilNeed != "" gate, not here.
func isRecoveryNeed(name string) bool { return recoveryNeeds[name] }

// needValue reads one need's live level by its recovery name — the need-agnostic
// accessor the per-tick completion check consults, so the hold/complete/abort
// machinery is generic across needs (spec 064 R1/US2), not warmth-private. An
// unknown name returns 0 (a closed-set door upstream makes this unreachable in
// practice); the caller only ever passes a validated UntilNeed.
func needValue(n Needs, name string) int {
	switch name {
	case "warmth":
		return n.Warmth
	case "rest":
		return n.Rest
	case "food":
		return n.Food
	default:
		return 0
	}
}

// recoveryPriority mirrors the reflex ladder's SURVIVAL ordering (policy.go
// survivalDecision: eat/get-food first, then the warmth ladder, then the nap):
// food is more urgent than warmth, warmth than rest. Lower = more urgent. It is
// the one place the recovery-preemption rank lives, so a hold's yield order
// tracks the ladder it yields to.
func recoveryPriority(need string) int {
	switch need {
	case "food":
		return 0
	case "warmth":
		return 1
	case "rest":
		return 2
	default:
		return 99
	}
}

// recoveryDangerBand returns a need's 062 danger band — the one home reused
// (dangerFoodBelow/dangerWarmthBelow/dangerRestBelow), so "in danger" means
// exactly what the PREP gate and the survival rungs already mean by it.
func recoveryDangerBand(need string) int {
	switch need {
	case "food":
		return dangerFoodBelow
	case "warmth":
		return dangerWarmthBelow
	case "rest":
		return dangerRestBelow
	default:
		return 0
	}
}

// preemptsRecovery reports whether a survival need MORE urgent than the one being
// recovered has crossed into its danger band (spec 064 US3 AS2): the hold must
// yield so the reflex's higher rung — which runs before the recovery's own rung —
// owns the agent. Only a strictly-higher-priority need preempts: a warmth
// recovery yields to hunger (food outranks warmth in the ladder), never the
// reverse, so a villager warming up doesn't abandon it for a lower-priority need.
func preemptsRecovery(a *Agent, recoveryNeed string) bool {
	pr := recoveryPriority(recoveryNeed)
	for _, n := range []string{"food", "warmth", "rest"} {
		if recoveryPriority(n) < pr && needValue(a.Needs, n) < recoveryDangerBand(n) {
			return true
		}
	}
	return false
}

// clampWarmUp resolves a requested warm_up threshold (spec 064 R3, clamp-with-
// notice — spec 058 posture): 0 (absent) yields the doctrine default
// warmthRecoverTo; any other value clamps into [warmthRecoverFloor, needMax].
// It returns the effective threshold and a human notice ("" when the request
// was in range or defaulted). The SINGLE clamp home — resolveGoal calls it to
// set the authoritative intent (the sim drives), and the mind handler calls the
// exported ClampWarmUp wrapper only to phrase the model-facing notice/verdict,
// so the two can never drift (the set_plan const-vs-const precedent).
func clampWarmUp(requested int) (int, string) {
	if requested == 0 {
		return warmthRecoverTo, ""
	}
	if requested < warmthRecoverFloor {
		return warmthRecoverFloor, fmt.Sprintf("until_warmth %d is below the sane floor — clamped to %d", requested, warmthRecoverFloor)
	}
	if requested > needMax {
		return needMax, fmt.Sprintf("until_warmth %d is above the need cap — clamped to %d", requested, needMax)
	}
	return requested, ""
}

// ClampWarmUp is the exported wrapper the mind's warm_up handler consults to
// phrase its clamp notice and pick the landed/landed-clamped verdict (spec 064
// R3). It delegates to clampWarmUp so the authoritative resolver clamp and the
// handler's model-facing notice share one implementation and never drift.
func ClampWarmUp(requested int) (int, string) { return clampWarmUp(requested) }

// isMindSource reports whether an intent's ring-record source counts as
// "intelligence" for the yield window (spec 062 US1): planner, plan, and
// meeting sources all do; a "reflex" source (or an empty/unknown one) does not
// — instinct must never arm its own deference (that would deadlock prep in a
// no-planner world). The single classifier the completion reducer arm consults.
func isMindSource(source string) bool {
	switch source {
	case "planner", "plan", "meeting":
		return true
	default:
		return false
	}
}

// --- spec 012 resources/food/crafting v2 tuning ---
//
// The single scalar tuning surface for the v2 economy, mirrored in
// specs/012-resources-food-crafting/contracts/recipes.md and the recipe table
// (recipes.go). Ticks are game-seconds (a game-hour is 3600 ticks). Phase 2
// only declares these; behavior is wired to them in later phases (T013–T037),
// so several are intentionally unused until then (package-level constants may
// be unused without a compile error).
const (
	// Food restore per unit eaten, on the 0..1000 need scale (cooking ~doubles
	// raw; the meal is the best food). Eating stops at Food >= satietyAt.
	foodRawRestore    = 40
	foodCookedRestore = 80
	mealRestore       = 100
	satietyAt         = 900

	// Reflex larder target (T018): idle agents top up carried raw food to this
	// many units before wandering. Restates the legacy stock-3-meals prep rule
	// over the finer raw unit (contracts/recipes.md sizing).
	stockFoodRawTo = 8

	// The refuel-dying-below window and fire-burn-per-wood are spec-048-promoted
	// dials: defaults live in tuning.go (defaultRefuelDyingBelow,
	// defaultFireBurnPerWood) and reads go through State.RefuelDyingBelow() /
	// State.FireBurnPerWood(). fireFuelCap is NOT promoted (research R6) — it
	// still truncates the effective per-wood deadline.
	fireFuelCap = 12 * 3600 // remaining fuel ceiling

	// Spear durability: hunts a spear lasts before breaking.
	spearDurability = 3

	// Rest regen per game-minute while asleep on a shelter tile (else
	// restRegenSleep = 4).
	restRegenShelter = 6

	// Bath effects at an oven (absolute post-values are carried in the event;
	// these are the pre-cap bumps applied, gru-pattern).
	bathMorale = 150
	bathWarmth = 300

	// v2 gather rescale (wired T013 quarry/water, T017 forage/hunt). The legacy
	// forageYield/huntYield constants are gone (T017): agent.foraged now yields
	// forageYieldV2 FoodRaw, agent.hunted huntYieldBare (spear boost is T027).
	// quarryYield (spec 012's flat 2) is deleted by spec 032 T014 — quarry yield
	// is now quarryYieldBare/quarryYieldAxe (agents.go spec-032 block).
	quarryTicks       = 400
	collectWaterYield = 1
	collectWaterTicks = 60
	forageYieldV2     = 2
	huntYieldBare     = 8
	huntYieldSpear    = 12
	huntTicksSpear    = 600

	// Hand-craft / build / station recipe magnitudes (mirrored by recipes.go;
	// wired T026/T030–T037).
	plankYield       = 4
	craftPlanksTicks = 180
	craftStoneTicks  = 180
	craftSpearTicks  = 240
	shelterPlankCost = 8
	buildOvenTicks   = 900
	ovenBatchSize    = 8
	cookFireTicks    = 240
	cookOvenTicks    = 360
	batheTicks       = 240
)

// --- spec 013 inventory/storage v1 tuning ---
//
// The scalar tuning surface for the storage layer, mirrored in
// specs/013-inventory-storage/data-model.md. Ticks are game-seconds. Phase 2
// only declares these; behavior is wired to them in later phases (US1–US5,
// migration), so several are intentionally unused until then.
const (
	bulkCap        = 24     // per-villager carried bulk ceiling
	chestCap       = 48     // per-chest stored bulk ceiling
	chestPlankCost = 6      // build_chest recipe input
	rotWindowTicks = 172800 // 2 game days: ground-pile food batch lifetime

	// Taking (theft) social marks — the deltas a non-owner withdrawal lays
	// through the existing relation/memory machinery (research R5).
	theftTrustDelta     = -120 // owner→taker trust on a taking
	theftAffectionDelta = -40  // owner→taker affection on a taking
	theftMemoryTone     = -60  // owner/witness memory tone (gossip seed)
)

// --- spec 032 walls/axes/paths tuning ---
//
// The single scalar tuning surface for this feature, mirrored in
// specs/032-walls-axes-paths/contracts/recipes.md and research R8. Ticks are
// game-seconds (a game-hour is 3600 ticks). Wall max HP is DERIVED from the kind
// via wallMaxHP (terrain.go), never stored — the fire lit-ness doctrine. The
// legacy flat chopWood/quarryYield are deleted in T014, replaced by the
// bare/axe yield pairs below.
const (
	wallPlankCost  = 2   // planks → wall_plank (build_wall_plank recipe input)
	wallStoneCost  = 2   // refined_stone → wall_stone (build_wall_stone recipe input)
	wallPlankHP    = 200 // plank wall max health
	wallStoneHP    = 600 // stone wall max health — 3x plank (spec FR-003: ≥2x)
	buildWallTicks = 600 // per-wall build work duration
	// wallOccupancyGraceTicks (spec 038): ticks past the wall's due tick
	// (WorkStart + buildWallTicks) that completion may defer on an occupied
	// reserved tile before failing loudly. 20% of buildWallTicks — long enough
	// for a passerby or short chat, short enough that a blocked build resolves
	// within a fraction of its own work duration. Derived bound, no persisted
	// state: the fail trigger is a pure function of WorkStart (research D2).
	wallOccupancyGraceTicks = 120
	demolishChipHP          = 100 // HP removed per demolish work cycle (plank: 2 cycles; stone: 6)
	demolishTicks           = 300 // per demolish chip cycle
	repairHPPerUnit         = 100 // HP restored per material unit consumed, clamped to max
	repairTicks             = 240 // per repair work cycle
	pathStoneCost           = 1   // raw stone per path tile (build_path recipe input)
	buildPathTicks          = 240 // path build work duration
	axeDurability           = 10  // harvest uses per fresh axe (chop/quarry far outpace hunting)

	// Harvest yield rebalance (spec FR-009/010): bare-handed drops from the
	// legacy flat 2 to 1; a carried axe triples it to 3. Replaces chopWood /
	// quarryYield (deleted T014).
	chopYieldBare   = 1 // wood per bare-handed chop
	chopYieldAxe    = 3 // wood per axe-assisted chop
	quarryYieldBare = 1 // stone per bare-handed quarry
	quarryYieldAxe  = 3 // stone per axe-assisted quarry
)

// Stable reason vocabulary for agent.build_failed (spec 038): a small closed
// set so tests and tooling can match on the string. buildFailSiteUnbuildable is
// any build goal whose site re-validation fails mid-work; buildFailSiteBlocked
// is walls only, once a reserved-tile occupant outlasts wallOccupancyGraceTicks.
const (
	buildFailSiteUnbuildable = "site no longer buildable"
	buildFailSiteBlocked     = "site blocked too long"
)

// bulk is an agent's (or a chest's) carried load: one per unit of every
// inventory kind plus one per carried spear (data-model.md). Derived, never
// stored — same doctrine as fire lit-ness from FuelUntil: a derived value
// cannot drift from its parts. bulkCap (24) exceeds the largest single yield
// (spear hunt, 12), so no completion is unsatisfiable from empty (research R2).
// Chest capacity uses this same function over *Store.
func bulk(inv Inventory) int {
	return inv.Wood + inv.Stone + inv.Water + inv.Planks + inv.RefinedStone +
		inv.FoodRaw + inv.FoodCooked + inv.Meals + len(inv.Spears) + len(inv.Axes)
}

// freeBulk is the remaining carry capacity under the cap: bulkCap − bulk(inv),
// floored at zero (a defensively over-cap inventory reports no free space, never
// a negative). The reducer yield clamps (US1-AS2) and the executor completion
// re-validation (US1-AS1) share it: a full pouch reports zero, a partially full
// one the exact remainder (research R2). Pure function of pre-event Inv, so
// replay is byte-identical.
func freeBulk(inv Inventory) int {
	if f := bulkCap - bulk(inv); f > 0 {
		return f
	}
	return 0
}

// BulkCap and Bulk are bulkCap/bulk exported for internal/tui (SC-006: "how
// full a villager's hands are" must be answerable from the TUI alone),
// mirroring the GuardianChargeCap export pattern for the same purpose —
// the sim package stays the single source of truth for the derived value
// and its ceiling.
const BulkCap = bulkCap

// ChestCap is chestCap exported for internal/tui (SC-006: "what's in a given
// chest, and is it full" must be answerable from the TUI alone), mirroring the
// BulkCap export — sim stays the single source of truth for the per-chest
// stored-bulk ceiling and its display.
const ChestCap = chestCap

func Bulk(inv Inventory) int {
	return bulk(inv)
}

// WallMaxHP is wallMaxHP exported for internal/tui (spec 032): the map view dims
// a wall glyph when its current HP is below the derived per-kind maximum
// (cold-fire precedent), and only sim knows that maximum. Mirrors the Bulk/
// BulkCap export — sim stays the single source of truth for the derived value.
func WallMaxHP(kind string) int {
	return wallMaxHP(kind)
}

// intentDurations is the per-goal-door-world-tool base work duration, DERIVED
// from the tool registry's Cost.DurationTicks at init (spec 014, R7). It
// replaces the hand-written intentDuration switch — the registry now carries
// the declarative duration for each world verb (values byte-identical to the
// old switch's constants). Context-dependent overrides (a spear-carrying hunt
// is faster, an oven cook is longer) stay in the executor's workDuration and
// are NOT registry data — the station/inventory is only known at completion
// time.
//
// Filtered to goal-door tools (Effect World AND PlanStep true — see
// toolcheck.go's ValidateToolCoverage doc): intentDuration(goal) is only ever
// called with a goal resolveGoal dispatched on. A non-goal-door World tool
// (set_plan, spec 017 R11) never reaches it by its own name — each of its
// plan steps names an already-covered goal-door goal instead — so it is
// deliberately left out of this table rather than carrying a meaningless
// zero-duration entry.
var intentDurations = func() map[string]int64 {
	m := make(map[string]int64)
	for _, t := range tool.All() {
		if t.Effect == tool.World && t.PlanStep {
			m[t.Name] = t.Cost.DurationTicks
		}
	}
	return m
}()

// intentDuration returns a goal's base work ticks. Goals with no registry
// duration — the instant world verbs (sleep, goto_warmth, wander, refuel_fire,
// the storage verbs) and the internal "seek" alias — complete on arrival (0),
// exactly as the old switch's default did.
func intentDuration(goal string) int64 {
	return intentDurations[goal]
}

// --- event payloads ---

type (
	IntentSetPayload struct {
		Agent   int    `json:"agent"`
		Goal    string `json:"goal"`
		TargetX int    `json:"target_x"`
		TargetY int    `json:"target_y"`
		ResX    int    `json:"res_x"`
		ResY    int    `json:"res_y"`
		Source  string `json:"source,omitempty"` // "reflex" | "planner"
		// Kind/Qty (spec 013 R4) carry a storage goal's argument onto the intent;
		// omitempty keeps pre-013 and non-storage intent_set payloads byte-identical.
		Kind string `json:"kind,omitempty"`
		Qty  int    `json:"qty,omitempty"`
		// Job (spec 017, TASK-32 pattern): the planner-loop job that landed
		// this intent, set ONLY at the inject-landing emission site (from
		// InjectArgs.JobID). Reflex-authored (decideIntent) and executor-
		// authored (adapt/continue) intent_set events carry no job — field
		// omitted, so pre-feature logs and reflex/executor emissions marshal
		// byte-identically to today. LAST field, omitempty.
		Job string `json:"job,omitempty"`
		// Reason (spec 019, R2) carries the planner's free-text reason onto the
		// intent so it survives to completion, where the executor bakes it into
		// the memory's Why. Set ONLY at the planner inject-landing emission site
		// (from InjectArgs.Reason); reflex- and executor-authored intent_set
		// events carry none. omitempty keeps pre-019 logs and reflex/executor
		// emissions byte-identical.
		Reason string `json:"reason,omitempty"`
		// Needs-conditioned recovery (spec 064 R1): the optional completion
		// condition riding onto the intent — a need name and threshold. Set at
		// the warm_up resolver (planner) and the reflex warmth rungs
		// (goto_warmth-with-condition); absent on every other intent_set. Both
		// omitempty, LAST — pre-064 and conditionless intent_set payloads marshal
		// byte-identically. The condition lives in THIS recorded event, so replay
		// repopulates it with no new state (determinism: the per-tick check is a
		// pure function of the resulting intent + needs).
		UntilNeed  string `json:"until_need,omitempty"`
		UntilValue int    `json:"until_value,omitempty"`
	}
	WorkStartedPayload struct {
		Agent int   `json:"agent"`
		Tick  int64 `json:"tick"`
		// Ref (spec 064 R4) is the need level captured at this hold anchor, for a
		// conditioned recovery's no-net-gain abort check. omitempty and LAST:
		// every pre-064 work_started (forage/chop/build) emits Ref 0 and marshals
		// byte-identically; a work goal never reads Intent.HoldRef, so the 0 it
		// writes there is inert.
		Ref int `json:"ref,omitempty"`
	}
	// RecoveryStalledPayload — agent.recovery_stalled (spec 064 R4/FR-004): a
	// needs-conditioned hold whose need showed no net gain across a full
	// recoveryStallTicks window (dead fire, displaced source, unreachable
	// threshold). Its OWN distinct type — an honest abort, not a completion — so
	// the reducer clears the intent like intent_done but stamps the ring "stalled"
	// and NEVER arms the spec-062 yield window (the build_failed precedent: an
	// abort is not intelligence completing). {agent, goal, need} mirrors the
	// {agent, goal, reason} failure shape so abort consumers see a familiar frame.
	RecoveryStalledPayload struct {
		Agent int    `json:"agent"`
		Goal  string `json:"goal"`
		Need  string `json:"need"`
	}
	HarvestPayload struct { // foraged / chopped / hunted / built site
		Agent int `json:"agent"`
		X     int `json:"x"`
		Y     int `json:"y"`
	}
	BuiltPayload struct {
		Agent int    `json:"agent"`
		Kind  string `json:"kind"`
		X     int    `json:"x"`
		Y     int    `json:"y"`
	}
	// BuildFailedPayload — agent.build_failed (spec 038): a build intent that
	// passed landing is cancelled by the executor's mid-work re-validation. Its
	// own type, distinct from the bare agent.intent_done a completion emits, so
	// observers/tests/the builder's mind can tell a failed build from a finished
	// one. Field shape mirrors IntentRejectedPayload (cognition.go) so failure
	// consumers see a familiar {agent, goal, reason}. Reducer clears the intent
	// exactly like intent_done; a paired situated memory rides the same tick.
	BuildFailedPayload struct {
		Agent  int    `json:"agent"`
		Goal   string `json:"goal"`
		Reason string `json:"reason"`
	}
	NeedsPayload struct {
		Agent  int `json:"agent"`
		Health int `json:"health"`
		Food   int `json:"food"`
		Rest   int `json:"rest"`
		Warmth int `json:"warmth"`
		Morale int `json:"morale"`
	}
	DiedPayload struct {
		Agent int    `json:"agent"`
		Cause string `json:"cause"` // "starvation" | "exposure" | "collapse" | "gru" (spec 044 US3)
	}
	TalkedPayload struct {
		A int `json:"a"`
		B int `json:"b"`
	}
	RegrownPayload struct {
		X int `json:"x"`
		Y int `json:"y"`
	}
	MemoryAddedPayload struct {
		Agent    int          `json:"agent"`
		Text     string       `json:"text"`
		Salience int          `json:"salience"`
		Subject  int          `json:"subject"`
		Tone     int          `json:"tone,omitempty"`
		Where    *MemoryPlace `json:"where,omitempty"`  // spec 019: location at emission
		Why      string       `json:"why,omitempty"`    // spec 019: driving intent reason, verbatim
		Conv     int64        `json:"conv,omitempty"`   // spec 019: conversation ref (founding-talk tick)
		Origin   string       `json:"origin,omitempty"` // spec 030: provenance class stamped at emission
	}
	// MemoryEmbeddedPayload — agent.memory_embedded (spec 042): the mind-side
	// embedder driver's recorded companion attaching a vector to the memory
	// whose agent.memory_added event carried store seq MemSeq. The reducer
	// copies Vec/Model verbatim onto the matching memory and no-ops when the
	// target is gone (agent died / memory consolidated away). Emitted ONLY by
	// the embedder through InjectSocial (whitelisted).
	MemoryEmbeddedPayload struct {
		Agent  int       `json:"agent"`
		MemSeq int64     `json:"mem_seq"`
		Vec    []float32 `json:"vec"`
		Model  string    `json:"model"`
	}
	// SituationEmbeddedPayload — agent.situation_embedded (spec 042): the
	// embedder driver's per-agent rolling situation (query) vector, rendered
	// from a deterministic situation template at planner cadence. Text is the
	// audit surface for divergence review; the reducer stores Vec/Model/Tick as
	// the agent's current SitVec* fields. Emitted ONLY by the embedder through
	// InjectSocial (whitelisted).
	SituationEmbeddedPayload struct {
		Agent int       `json:"agent"`
		Tick  int64     `json:"tick"`
		Text  string    `json:"text"`
		Vec   []float32 `json:"vec"`
		Model string    `json:"model"`
	}
	ThoughtPayload struct {
		Agent  int    `json:"agent"`
		Text   string `json:"text"`
		Source string `json:"source"` // "planner" (reflex acts without narrating)
	}
	// Hail lifecycle (TASK-47). from = hailer, to = target — the field names
	// the chronicle grammar already resolves to agent names, so tail/TUI
	// visibility lands with no view-layer change.
	HailedPayload struct {
		From  int   `json:"from"`
		To    int   `json:"to"`
		Until int64 `json:"until"`
	}
	HailMetPayload struct {
		From int `json:"from"`
		To   int `json:"to"`
	}
	HailExpiredPayload struct {
		From int `json:"from"`
		To   int `json:"to"`
	}

	// --- spec 012 resources/food/crafting v2 payloads ---
	// Field order below is the canonical serialization order (see
	// contracts/events.md); all outcomes are absolute (no deltas, no dice).

	// CraftedPayload: a completed hand-craft. Kind ∈ planks|refined_stone|spear;
	// the reducer applies the recipe delta from recipes.go.
	CraftedPayload struct {
		Agent int    `json:"agent"`
		Kind  string `json:"kind"`
	}
	// AtePayload replaces the old empty AgentPayload for agent.ate (the format
	// bump shields old logs): counts consumed per form plus the absolute
	// post-eat food need. Wired in Phase 4 (T018).
	AtePayload struct {
		Agent     int `json:"agent"`
		Meals     int `json:"meals"`
		Cooked    int `json:"cooked"`
		Raw       int `json:"raw"`
		FoodAfter int `json:"food_after"`
	}
	// CookedPayload: a cook batch. Station ∈ fire|oven; Kind ∈
	// food_cooked|meals. Consumed FoodRaw → Produced of Kind.
	CookedPayload struct {
		Agent    int    `json:"agent"`
		Station  string `json:"station"`
		Consumed int    `json:"consumed"`
		Produced int    `json:"produced"`
		Kind     string `json:"kind"`
	}
	// BathedPayload: a bath at an oven — absolute post-cap need values
	// (gru-pattern).
	BathedPayload struct {
		Agent       int `json:"agent"`
		MoraleAfter int `json:"morale_after"`
		WarmthAfter int `json:"warmth_after"`
	}
	// RefueledPayload: a fire refuel (planner or reflex). FuelUntil is the
	// absolute new deadline (already capped by the emitter).
	RefueledPayload struct {
		Agent     int   `json:"agent"`
		X         int   `json:"x"`
		Y         int   `json:"y"`
		FuelUntil int64 `json:"fuel_until"`
	}
	// SpearBrokePayload: the spear that spent its last use, alongside the hunt
	// completion; a companion memory rides the same batch.
	SpearBrokePayload struct {
		Agent int `json:"agent"`
	}
	// AxeBrokePayload (spec 032 US2): the axe that spent its last harvest use,
	// co-emitted immediately after the chop/quarry completion when the pre-event
	// Axes[0] == 1 — the exact SpearBrokePayload clone. A companion memory rides
	// the same batch.
	AxeBrokePayload struct {
		Agent int `json:"agent"`
	}
	// FireBurnedOutPayload: the fuel sweep's once-per-burnout signal. No state
	// effect (lit-ness is derived from FuelUntil); chronicle/TUI material.
	FireBurnedOutPayload struct {
		X int `json:"x"`
		Y int `json:"y"`
	}

	// WallWorkPayload (spec 032 US1) is the {agent,x,y} shape shared by the three
	// wall work-cycle events — agent.wall_chipped, agent.wall_destroyed,
	// agent.wall_repaired. (x,y) is the wall tile (the intent's Res); Agent is the
	// actor, so the reducer can reset that agent's Intent.WorkStart to 0 and
	// re-arm the executor's work gate for the next demolish/repair cycle (R5).
	WallWorkPayload struct {
		Agent int `json:"agent"`
		X     int `json:"x"`
		Y     int `json:"y"`
	}

	// --- spec 013 inventory/storage v1 payloads ---
	// Field order below is the canonical serialization order (see
	// contracts/events.md); every count is the ACTUAL post-clamp moved amount
	// (outcome-only), never a request.

	// DroppedPayload: n of kind left on the agent's tile, created-or-merged into
	// the tile's pile (food becomes a batch stamped tick + rotWindowTicks;
	// spears move most-worn-first with their durabilities).
	DroppedPayload struct {
		Agent int    `json:"agent"`
		X     int    `json:"x"`
		Y     int    `json:"y"`
		Kind  string `json:"kind"`
		N     int    `json:"n"`
	}
	// PickedUpPayload: n of kind taken from the tile's pile (food oldest-batch-
	// first), truncated to free bulk; one event per kind moved in the batch.
	PickedUpPayload struct {
		Agent int    `json:"agent"`
		X     int    `json:"x"`
		Y     int    `json:"y"`
		Kind  string `json:"kind"`
		N     int    `json:"n"`
	}
	// DepositedPayload: n of kind moved from inventory into the chest at (x,y),
	// truncated to chest free space (chestCap − bulk(*Store)).
	DepositedPayload struct {
		Agent int    `json:"agent"`
		X     int    `json:"x"`
		Y     int    `json:"y"`
		Kind  string `json:"kind"`
		N     int    `json:"n"`
	}
	// WithdrewPayload: n of kind taken from the chest at (x,y) into inventory,
	// truncated to the taker's free bulk. Owner is the chest's owner index; a
	// non-owner taker co-emits the theft companion batch (contracts/events.md).
	WithdrewPayload struct {
		Agent int    `json:"agent"`
		X     int    `json:"x"`
		Y     int    `json:"y"`
		Kind  string `json:"kind"`
		N     int    `json:"n"`
		Owner int    `json:"owner"`
	}
	// FoodRottedPayload: n of a food kind removed from the pile at (x,y) by the
	// per-game-minute rot sweep (same-kind batches merged per pile per sweep).
	FoodRottedPayload struct {
		X    int    `json:"x"`
		Y    int    `json:"y"`
		Kind string `json:"kind"`
		N    int    `json:"n"`
	}
)

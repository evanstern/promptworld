package cognition

import "fmt"

// Degrade names the deterministic action a decision class falls to when the
// router suppresses it: the model is not consulted, and the class's floor
// behavior runs instead.
type Degrade string

const (
	DegradeSkip     Degrade = "skip"     // nothing replaces the thought (recorded, not silent)
	DegradeReflex   Degrade = "reflex"   // the deterministic reflex floor covers
	DegradeTemplate Degrade = "template" // pre-authored text stands (meeting rephrase)
	// A faster-tier degrade (DegradeFasterTier) was removed as unwired dead
	// code (TASK-71); git history has it if a faster tier is ever built.
)

// DecisionClass is one registered category of model-reaching decision: its
// thought cost in Fibonacci points (a property of the prompt shape,
// host-independent) and its staleness budget in game ticks. BudgetTicks is
// the budget AT 1x — wall-clock patience (spec 067 FR-005): the scheduling
// gates (Route/RoutePaused, governor debt, every horizon surface) hold it
// fixed against the fiction, while the delivery gates (the reducer landing
// rung, the mind's scene pre-abort) enforce it scaled by the event-sourced
// clock speed via EffectiveBudgetTicks, so a constant-wall-time thought is
// judged the same at every capped speed. Values are doctrine (decision-4):
// changing one is a reviewed code change, never runtime tuning.
type DecisionClass struct {
	Class       string
	Points      int
	BudgetTicks int64
	Degrade     Degrade
	FutureDated bool
}

// EffectiveBudgetTicks is the class's staleness budget at a given tick rate
// (spec 067 FR-001): BudgetTicks is the 1x budget, so the effective budget is
// BudgetTicks × ticksPerSecond — the same wall-clock patience at every capped
// speed. ticksPerSecond <= 0 (uncapped max) returns the base budget unscaled,
// mirroring Route's posture (theoretical branch: Route suppresses every class
// at uncapped speed, so no admitted thought should reach a delivery gate
// there). Ladder rates are small exact floats, so the product is exact for
// every registry budget — deterministic across platforms, replay-safe.
func (dc DecisionClass) EffectiveBudgetTicks(ticksPerSecond float64) int64 {
	if ticksPerSecond <= 0 {
		return dc.BudgetTicks
	}
	return int64(float64(dc.BudgetTicks) * ticksPerSecond)
}

// fibonacci is the closed set of legal point values.
var fibonacci = map[int]bool{1: true, 2: true, 3: true, 5: true, 8: true, 13: true}

// registry holds the initial values from
// specs/007-cognition-horizon/contracts/registry.md.
var registry = map[string]DecisionClass{
	"planner":       {Class: "planner", Points: 3, BudgetTicks: 1200, Degrade: DegradeReflex, FutureDated: true},
	"conversation":  {Class: "conversation", Points: 13, BudgetTicks: 7200, Degrade: DegradeSkip},
	"meeting":       {Class: "meeting", Points: 2, BudgetTicks: 3600, Degrade: DegradeTemplate},
	"consolidation": {Class: "consolidation", Points: 5, BudgetTicks: 28800, Degrade: DegradeSkip},
	"chronicle":     {Class: "chronicle", Points: 5, BudgetTicks: 86400, Degrade: DegradeSkip},
	// The "metatron" class/kind strings are FROZEN (spec 052 ruling 2):
	// recorded cog.* payloads carry the class, and llm.json routes the kind.
	"metatron": {Class: "metatron", Points: 5, BudgetTicks: 86400, Degrade: DegradeSkip},
	// The "steward" class (spec 102, D2/FR-001; operator rename ruling
	// 2026-07-30 de-themed the serialized spelling from the spec's "angel"
	// design vocabulary before first merge — no recorded log ever carried
	// the old string): the guardian's SCHEDULED
	// cognition lane — the agentized guardian observing and acting on its own
	// cadence, beside (never replacing) the event-driven doors that keep the
	// "metatron" class above. Points 5: a full guardian turn prompt (charter +
	// skills + replica digests + tool loop), the metatron shape. BudgetTicks
	// 900 (15 game-minutes at 1x): deliberately BELOW planner's 1200 so the
	// angel is the FIRST class the router sheds as speed rises — villager
	// survival cognition (planner, reflex-floored) must always outlive the
	// caretaker's ambient turns under saturation (D2's shed order; the
	// registry_test pin holds the inequality). DegradeSkip: a suppressed
	// angel turn simply doesn't happen — the event doors still answer, and
	// the world never waits on the caretaker.
	"steward": {Class: "steward", Points: 5, BudgetTicks: 900, Degrade: DegradeSkip},
}

// kindToClass maps every llm call kind (as a string, keeping this package
// leaf) to its decision class. Completeness against the orchestrator's
// accepted kinds is enforced at daemon start via ValidateKinds (FR-002).
var kindToClass = map[string]string{
	"planner":       "planner",
	"conversation":  "conversation",
	"meeting":       "meeting",
	"consolidation": "consolidation",
	"narrator":      "chronicle",
	"drama":         "chronicle",
	"metatron":      "metatron",
	// The guardian's fuzzy-order watch confirm (spec 029) shares the guardian
	// decision class: same actor, DegradeSkip (an unconfirmed/failed confirm
	// leaves the order armed — nothing runs), long staleness budget (the confirm
	// is event-triggered, never cadence-scheduled). Reusing the class keeps this
	// a one-line mapping — the narrator/drama→chronicle precedent — without
	// touching the spec-007 registry.md doctrine contract.
	"metatron_watch": "metatron",
	// The guardian's report-card critique (spec 063) shares the guardian
	// class for the same reasons the watch confirm does: same actor,
	// DegradeSkip (an unavailable chain means the deterministic card parts
	// stand alone — nothing runs), event-triggered at stopping points, never
	// cadence-scheduled.
	"report_card": "metatron",
	// The steward's scheduled turns (spec 102): its OWN kind and class, so the
	// router, governor debt, and horizon surfaces budget the cadence lane
	// separately from the event-driven "metatron" doors it rides beside.
	"steward": "steward",
}

// ClassFor returns the registered class by name.
func ClassFor(class string) (DecisionClass, bool) {
	dc, ok := registry[class]
	return dc, ok
}

// ClassForKind resolves an llm call kind to its decision class.
func ClassForKind(kind string) (DecisionClass, bool) {
	name, ok := kindToClass[kind]
	if !ok {
		return DecisionClass{}, false
	}
	return ClassFor(name)
}

// Validate checks registry invariants: Fibonacci point membership, positive
// budgets, kind mappings that resolve. Fatal at daemon start on failure.
func Validate() error {
	for name, dc := range registry {
		if dc.Class != name {
			return fmt.Errorf("cognition registry: class %q keyed as %q", dc.Class, name)
		}
		if !fibonacci[dc.Points] {
			return fmt.Errorf("cognition registry: class %q points %d not in the Fibonacci set", name, dc.Points)
		}
		if dc.BudgetTicks <= 0 {
			return fmt.Errorf("cognition registry: class %q budget %d not positive", name, dc.BudgetTicks)
		}
	}
	for kind, class := range kindToClass {
		if _, ok := registry[class]; !ok {
			return fmt.Errorf("cognition registry: kind %q maps to unregistered class %q", kind, class)
		}
	}
	return nil
}

// ValidateKinds enforces intentional categorization (FR-002): every kind the
// orchestrator accepts must resolve to a registered decision class, or the
// daemon must not start. The error names the offender.
func ValidateKinds(kinds []string) error {
	if err := Validate(); err != nil {
		return err
	}
	for _, k := range kinds {
		if _, ok := ClassForKind(k); !ok {
			return fmt.Errorf("cognition registry: llm kind %q has no registered decision class — register it in internal/cognition/registry.go (FR-002)", k)
		}
	}
	return nil
}

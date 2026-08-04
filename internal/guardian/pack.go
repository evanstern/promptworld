package guardian

// inspect_pack (spec 116 US1, FR-001..FR-005): the guardian's free,
// deterministic look INSIDE a villager's pack — the survey_site / explain
// pattern pointed at an inventory. It exists because gross bulk is not enough:
// the targeting digest already printed "carrying 24/24, 0 free to receive" for
// a starving man in world-03 and the guardian, unable to see WHAT he carried,
// sent him a vision telling him to eat food he did not have. Bulk alone cannot
// tell "has no food" from "has food and will not eat" from "has no room to
// pick food up", and those three want three different interventions.
//
// The sheet assembles TURN-SIDE (the handleSurvey rationale: the tool package
// holds no world state) from the absorb-mirrored contents snapshot — a pure
// function of (name, mirrored inventory) with a fixed kind order and no clock
// reads, so identical calls return byte-identical text. Effect Read, Gate
// None: no charge moves, no event lands, and the loop driver's read exemption
// keeps it off the turn's one-act budget ("looking is looking, not an act").
//
// It is also the gate's evidence: a successful resolution records the villager
// in this turn's look-first ledger (toolcalls.go), which is what a survival
// turn's send_vision / give_item / take_item is checked against.

import (
	"context"
	"fmt"
	"strings"

	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/toolloop"
)

// packKinds is the sheet's fixed rendering order — the sim.Inventory field
// order, which is also tool.GrantKinds()' order for the overlapping kinds.
// Fixed order = deterministic sheet bytes (the surveyTerrainNames discipline).
// The plural "spears"/"axes" names the STORAGE fields (a villager carries many,
// each with its own remaining uses), not give_item's singular grant unit — the
// two vocabularies are deliberately distinct (internal/tool/registry.go).
var packKinds = []string{
	"wood", "stone", "water", "planks", "refined_stone",
	"food_raw", "food_cooked", "meals", "spears", "axes",
}

// packFoodKinds is the edible subset, in the same fixed order: the sheet's
// closing line is mandatory and answers, in words, the one question the digest
// line could not — does this villager carry anything they can eat?
var packFoodKinds = []string{"food_raw", "food_cooked", "meals"}

// copyInventory deep-copies a villager's inventory for the guardian's mirror
// (spec 116 FR-006). The flat counts copy by assignment; the Spears/Axes
// remaining-uses slices MUST be cloned — the absorb goroutine keeps mutating
// the replica's slices in place (hunts and harvests spend uses), and an
// aliased mirror would let a sheet report a durability that changed after the
// snapshot was taken. Called under stateMu by mirrorState.
func copyInventory(inv sim.Inventory) sim.Inventory {
	out := inv
	out.Spears = append([]int(nil), inv.Spears...)
	out.Axes = append([]int(nil), inv.Axes...)
	return out
}

// handleInspectPack serves the read-only inspect_pack tool (the handleSurvey
// dispatch shape): always read_ok — a missing, unknown, or dead villager
// returns an honest in-fiction miss naming the living roster (the explain
// unknown-topic shape), which IS the data the model asked for, never a hard
// error. A miss records NOTHING in the look-first ledger: the guardian has not
// looked into anyone's pack, and the existing dead-target refusals already
// stop it acting on the dead, so the gate can never trap it in an
// unsatisfiable loop.
func (mt *Guardian) handleInspectPack(d *turnDispatch) toolloop.Handler {
	return func(_ context.Context, call llm.ToolCall) toolloop.Outcome {
		name := argString(call.Args, "villager")
		idx := agentIndexByName(name)
		// The mirrored contents + liveness, read once under stateMu (the
		// buildSurveySheet probe discipline) — the sheet never touches the
		// replica the absorb goroutine owns.
		mt.stateMu.Lock()
		alive := idx >= 0 && mt.alive[idx]
		var inv sim.Inventory
		if alive && idx < len(mt.agentInv) {
			inv = copyInventory(mt.agentInv[idx])
		}
		var living []string
		for i, n := range sim.AgentNames {
			if mt.alive[i] {
				living = append(living, n)
			}
		}
		mt.stateMu.Unlock()

		if !alive {
			return toolloop.Outcome{Verdict: toolloop.VerdictReadOK,
				ResultForModel: packMiss(name, living)}
		}
		// The look-first ledger (FR-007): only a SUCCESSFUL resolution counts
		// as having looked, and it licenses this villager alone.
		d.noteLooked(idx)
		return toolloop.Outcome{Verdict: toolloop.VerdictReadOK,
			ResultForModel: buildPackSheet(sim.AgentNames[idx], inv)}
	}
}

// packMiss renders the honest in-fiction miss (FR-005): the living roster,
// which is the repairable answer — the model can name one of these and call
// again, free, within the same turn.
func packMiss(name string, living []string) string {
	who := strings.TrimSpace(name)
	if who == "" {
		return "name the villager whose pack you would look in — the living are: " + strings.Join(living, ", ")
	}
	if len(living) == 0 {
		return fmt.Sprintf("no living villager answers to %q — no one lives in this world now", who)
	}
	return fmt.Sprintf("no living villager answers to %q — the living are: %s", who, strings.Join(living, ", "))
}

// buildPackSheet renders the deterministic pack sheet (contracts/pack-access.md
// §1): the header's carried/cap and free bulk, one line per carried kind in the
// fixed packKinds order (zero-count kinds omitted), spears and axes carrying
// their remaining-uses list, and the mandatory closing food line. Pure over
// (name, inv): no clock read, no map iteration, no replica access — so two
// identical calls in one turn return byte-identical bytes (FR-003).
func buildPackSheet(name string, inv sim.Inventory) string {
	// The cap is sim.BulkCap, never a literal (the buildTargetingDigest rule),
	// and free bulk floors at 0 — a defensively over-cap inventory reports no
	// room, never a negative (the sim.freeBulk discipline).
	carried := sim.Bulk(inv)
	free := sim.BulkCap - carried
	if free < 0 {
		free = 0
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s's pack — carrying %d/%d, %d free to receive:", name, carried, sim.BulkCap, free)
	lines := 0
	for _, kind := range packKinds {
		n := packCount(inv, kind)
		if n == 0 {
			continue
		}
		lines++
		switch kind {
		case "spears":
			fmt.Fprintf(&b, "\n- spears: %d (uses left: %s)", n, usesLeft(inv.Spears))
		case "axes":
			fmt.Fprintf(&b, "\n- axes: %d (uses left: %s)", n, usesLeft(inv.Axes))
		default:
			fmt.Fprintf(&b, "\n- %s: %d", kind, n)
		}
	}
	if lines == 0 {
		b.WriteString(" the pack is empty.")
	}
	b.WriteString("\n" + packFoodLine(name, inv) + "\n")
	return b.String()
}

// packFoodLine is the sheet's mandatory closing statement (FR-002): the plain
// words the gross-bulk digest line never had. Exactly one of the two forms,
// listing only the non-zero food kinds in fixed order.
func packFoodLine(name string, inv sim.Inventory) string {
	var carried []string
	for _, kind := range packFoodKinds {
		if n := packCount(inv, kind); n > 0 {
			carried = append(carried, fmt.Sprintf("%d %s", n, kind))
		}
	}
	if len(carried) == 0 {
		return name + " carries no food."
	}
	return fmt.Sprintf("%s carries food: %s.", name, strings.Join(carried, ", "))
}

// packCount is how many units of a kind the inventory holds, in the STORAGE
// vocabulary (spears/axes counted, durability living in the slice) — the
// sim.carriedCount shape, re-expressed here because sim keeps it unexported.
func packCount(inv sim.Inventory, kind string) int {
	switch kind {
	case "wood":
		return inv.Wood
	case "stone":
		return inv.Stone
	case "water":
		return inv.Water
	case "planks":
		return inv.Planks
	case "refined_stone":
		return inv.RefinedStone
	case "food_raw":
		return inv.FoodRaw
	case "food_cooked":
		return inv.FoodCooked
	case "meals":
		return inv.Meals
	case "spears":
		return len(inv.Spears)
	case "axes":
		return len(inv.Axes)
	}
	return 0
}

// usesLeft renders a tool slice's remaining uses in the slice's own ascending
// order (the sim invariant: most-worn first), so the guardian can see which
// tools a removal would take before it takes them.
func usesLeft(uses []int) string {
	parts := make([]string, len(uses))
	for i, u := range uses {
		parts[i] = fmt.Sprint(u)
	}
	return strings.Join(parts, ", ")
}

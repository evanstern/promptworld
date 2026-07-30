package guardian

// The deliberate-incompetence ceiling (spec 102 D3, operator-adopted ruling
// 2026-07-30): the DEFAULT compiled charter caps the guardian's world-acting
// INITIATIVE on the scheduled lane — conservative posture, single-step
// observation only, no unprompted spending beyond the TASK-111 genesis
// survival watches. A player-AUTHORED charter lifts the ceiling: competence
// is bought with authorship.
//
// The ceiling is CHARTER-COMPILATION DATA, never model degradation: it
// compiles at the same per-turn charter load site every other instruction
// fact does, and it narrows the turn's grant set through the SAME
// intersection machinery the stage ceiling and persona bundles use
// (intersectGrant) — so a capped tool is structurally absent from all three
// gating layers (declaration, prose, door) at once.
//
// Scope (the operator's ruling, verbatim in spec 102):
//   - It caps INITIATIVE only — the SCHEDULED (angel) lane. Console turns
//     (player-commanded = compliance) and order-triggered turns
//     (pre-authorized) run at full competence under ANY ceiling.
//   - It never touches tutor quality: explain/converse are always full — the
//     modest set below grants every read/counsel surface.
//   - The genesis survival watches are untouched either way: they are sim-
//     seeded state with their own trigger path (orders.go), not a turn grant.

import (
	"github.com/evanstern/promptworld/internal/bundle"
)

// angelModestTools is the DEFAULT ceiling's whole scheduled-lane roster:
// read-only observation and counsel — explain (the tutor's grounding tool),
// survey_site, brief_myths. No charge spend, no watches, no plans, no
// miracles, no clock: the dutiful-but-modest caretaker observes, notes, and
// holds counsel for the player.
var angelModestTools = []string{"explain", "survey_site", "brief_myths"}

// angelClockTools is the meta triple that stays the PLAYER's at any ceiling
// (spec 102 D3 lifts "multi-step initiative, proactive designations/
// directives, discretionary spending" — never the clock): a lifted charter
// still may not pause, start, or re-pace the world on its own cadence.
var angelClockTools = []string{"pause", "start", "adjust_speed"}

// angelCharterLifted compiles the ceiling bit from the EFFECTIVE charter
// text this turn runs under (the observeCharter derivation, so the recorded
// default flag and the ceiling can never disagree): game-authored text — the
// world's preset constant or a retired legacy seed — keeps the ceiling ON;
// any player-authored revision lifts it.
func angelCharterLifted(effectiveCharter, preset string) bool {
	return !(effectiveCharter == presetCharter(preset) || isLegacyDefault(effectiveCharter))
}

// applyAngelCeiling narrows a SCHEDULED turn's grant by the compiled ceiling
// (D3, FR-003). Intersection-only, like every other narrowing layer:
//
//	ceiling ON  → the modest read/counsel set, zero miracle kinds — the
//	              intersectGrant path, so bundle tools are capped too;
//	lifted      → the full world grant minus the clock triple (initiative
//	              bought by authorship; the clock stays the player's).
//
// Console and triggered turns never pass through here — full competence at
// any ceiling (runTurn gates this on turnOrigin.angel alone).
func applyAngelCeiling(g grantSet, lifted bool) grantSet {
	if !lifted {
		return intersectGrant(g, &bundle.GrantDoc{Tools: angelModestTools, MiracleKinds: []string{}})
	}
	// Lifted: strip only the clock triple. Copy-on-write — the maps feed
	// three gating layers this turn; never mutate a shared view in place.
	// PLANNING-TIER RULING (2026-07-30): bundle tools stay EXCLUDED from the
	// scheduled lane's roster at ANY ceiling — the modest arm excludes them
	// via intersectGrant's named-list semantics above, and the lifted arm
	// inherits the same posture (bundle tools remain console/order-driven);
	// revisit via a future card, never silently here.
	tools := make(map[string]bool, len(g.tools))
	for n := range g.tools {
		tools[n] = true
	}
	for _, n := range angelClockTools {
		delete(tools, n)
	}
	g.tools = tools
	if g.toolsConstrained {
		ba := make(map[string]bool, len(g.bundleAllowed))
		for n := range g.bundleAllowed {
			ba[n] = true
		}
		for _, n := range angelClockTools {
			delete(ba, n)
		}
		g.bundleAllowed = ba
	}
	return g
}

// guardianAngelModestFrame is the DEFAULT-ceiling scheduled turn's
// initiative frame (D3): a compile-time constant appended last beneath all
// editable content (spec 021 INV-1), so no charter or skill byte can
// displace it. It names the ceiling honestly — initiative capped, compliance
// and counsel never — and the door-side grant intersection backs it
// independently (the capped tools are structurally absent from the roster).
const guardianAngelModestFrame = `This turn is your own scheduled watch — no player message, no order due. Your charter grants ` +
	`you no initiative beyond watchfulness: survey the world, keep your notes, and hold your counsel for the player's ` +
	`next visit. On your own cadence you spend nothing and command nothing — no workings, no watches, no plans, no ` +
	`clock; the survival watches you already keep are the whole of your standing authority. This modesty caps ONLY ` +
	`what you take up alone: when the player asks, or a watch the player placed comes due, you act with your full skill, ` +
	`and your counsel and explanations are always your best.`

// guardianAngelLiftedFrame is the AUTHORED-charter scheduled turn's
// initiative frame (D3's lift): initiative within the charter's own words,
// the clock still the player's, the order door still the one arbiter (D6).
// Compile-time constant, appended last (INV-1).
const guardianAngelLiftedFrame = `This turn is your own scheduled watch — no player message, no order due. Your charter grants you ` +
	`initiative: you may act on your own judgment through the powers this world grants — visions, omens, watches, plans, ` +
	`workings — spending banked charges where your charter justifies it, one act at most as ever. Two bounds hold even so: ` +
	`the world's clock (pausing, starting, changing its pace) is the player's alone, never yours to take up on your own ` +
	`cadence; and standing orders fire through their own watch machinery — review your watches here, but never carry out ` +
	`a standing order's action from this turn.`

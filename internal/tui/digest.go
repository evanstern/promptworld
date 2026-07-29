package tui

// Chronicle digest registry (TASK-60, specs/018-chronicle-digest). Every
// cataloged event type gets a digestFunc here, keyed by its full event type
// — the per-type template it implements is contracts/digest-grammar.md §3
// row-for-row; sample payloads for every row live in the catalog sweep
// fixture (digest_test.go, contract §7). A registry miss, or ok=false on
// unmarshal failure, is handled by formatChronicleLine's fallback
// (FR-002/FR-003) — never here.
//
// Each digestFunc is a pure function over the stored event + the replica's
// agent-name table (R1), returning the summary as ordered role-tagged
// segments (grammar.go's `seg`) — no lipgloss, no ANSI. Where the real
// payload struct (internal/sim) didn't carry a field the contract's
// illustrative template assumed, the digest below adapts to the actual
// struct and a doc comment says so; the implementer's report to the
// orchestrator lists every such row.

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/cognition"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/skin"
	"github.com/evanstern/promptworld/internal/store"
)

// digestFunc renders one event type's summary, or ok=false on unmarshal
// failure (data-model.md).
type digestFunc func(e store.Event, names []string, sk *skin.Skin) (segs []seg, ok bool)

// decode unmarshals an event's payload into T; ok=false on failure is the
// signal formatChronicleLine falls back on.
func decode[T any](e store.Event) (T, bool) {
	var p T
	if err := json.Unmarshal(e.Payload, &p); err != nil {
		var zero T
		return zero, false
	}
	return p, true
}

// --- seg builders — small helpers keeping the registry table legible ---

func txt(s string) seg { return seg{Text: s, Role: segText} }
func nameOf(names []string, idx int) seg {
	return seg{Text: agentName(names, idx), Role: segName}
}

// refSeg is the payload-first name seg (spec 086 FR-007): a post-086 ref
// carries its name in the payload — render it with NO replica lookup; a
// legacy row's ref (Name == "") falls back to the replica-derived names
// slice via agentName, exactly as before. The names parameter survives as
// the fallback channel, never the primary source.
func refSeg(names []string, r sim.AgentRef) seg {
	return seg{Text: refName(names, r), Role: segName}
}

// refName is refSeg's plain-string twin, for digests that compose names
// into larger strings.
func refName(names []string, r sim.AgentRef) string {
	if r.Name != "" {
		return r.Name
	}
	return agentName(names, r.ID)
}
func speech(s string) seg { return seg{Text: fmt.Sprintf("%q", s), Role: segSpeech} }
func emph(s string) seg   { return seg{Text: s, Role: segEmphasis} }
func label(kv string) seg { return seg{Text: kv, Role: segLabel} }
func emphN(n int) seg     { return emph(strconv.Itoa(n)) }
func emphI64(n int64) seg { return emph(strconv.FormatInt(n, 10)) }
func coord(x, y int) seg  { return emph(fmt.Sprintf("(%d,%d)", x, y)) }

// labeled builds a space-separated run of "key=value" segLabel spans — the
// cog/clock/daemon telemetry voice (contract §2).
func labeled(pairs ...string) []seg {
	out := make([]seg, 0, len(pairs)*2-1)
	for i, p := range pairs {
		if i > 0 {
			out = append(out, txt(" "))
		}
		out = append(out, label(p))
	}
	return out
}

// truncateRunes bounds a free-text field the contract explicitly marks
// "truncating" (chronicle.entry's text, plan_set's goal list).
func truncateRunes(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max <= 1 {
		return string(r[:max])
	}
	return string(r[:max-1]) + "…"
}

// join concatenates seg slices — a small variadic append helper so a
// registry entry can compose fixed segs with a conditional tail.
func join(parts ...[]seg) []seg {
	var out []seg
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

// debtPercent expresses a measured governor debt (a budget-fraction sum,
// spec 028 FR-001) as a whole percent of cognition.ShedThreshold, rounded to
// the nearest percent — the shared arithmetic behind both the header's
// governed-speed suffix (views.go) and the governor digest lines below.
func debtPercent(debt float64) int {
	return int(math.Round(debt / cognition.ShedThreshold * 100))
}

// gratisMark appends a visible " (forced)" annotation when a miracle's
// Gratis flag waived its charge (internal/sim/miracles.go) — the operator
// force spec 016 SC-004 requires stay enumerable must be visible in the
// digest line, not just inferable from the payload. nil (no segs) when the
// miracle was charge-priced, so the plain summary is unchanged for the
// common case.
func gratisMark(gratis bool) []seg {
	if !gratis {
		return nil
	}
	return []seg{txt(" ("), emph("forced"), txt(")")}
}

// digestRegistry is the ~80-entry per-type table (contract §3). A key
// absent here (or present in the fixture but not here) fails the catalog
// sweep test (digest_test.go, contract §7, SC-001).
var digestRegistry = map[string]digestFunc{
	// --- world / clock / daemon ---

	"world.created": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.WorldCreatedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{txt("world "), emph(fmt.Sprintf("%q", p.Name)), txt(" created · seed "), emph(fmt.Sprintf("%d", p.Seed))}), true
	},
	// world.migrated elides the embedded sim.State entirely (FR-011) — the
	// detail pane (Phase 4) is where an oversized payload gets bounded, not
	// the feed line.
	"world.migrated": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.WorldMigratedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{
			txt("migrated from format v"), emphN(p.FromFormat),
			txt(" · "), emphI64(p.SourceEvents), txt(" events @ tick "), emphI64(p.SourceTick),
		}), true
	},
	// world.forked (spec 076 FR-009): the fork's provenance line — which
	// world it split from and the boundary in game time. The digest line is
	// the v1 rendering; a chronicle narration for the split is a documented
	// unfunded follow-on (spec 076 Out of Scope).
	"world.forked": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.WorldForkedPayload](e)
		if !ok {
			return nil, false
		}
		day, h, min, _ := clock.GameTime(p.ForkTick)
		return join([]seg{
			txt("forked from "), emph(fmt.Sprintf("%q", p.ParentName)),
			txt(fmt.Sprintf(" at day %d, %02d:%02d", day, h, min)),
		}), true
	},
	"clock.paused":  func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) { return []seg{txt("paused")}, true },
	"clock.resumed": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) { return []seg{txt("resumed")}, true },
	"clock.speed_set": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.SpeedSetPayload](e)
		if !ok {
			return nil, false
		}
		return labeled("speed=" + string(p.Speed)), true
	},
	"clock.degraded": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.DegradedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{txt("degraded ")}, labeled(fmt.Sprintf("rate=%.2f", p.EffectiveRate))), true
	},
	"clock.recovered": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) { return []seg{txt("recovered")}, true },
	// clock.governor_shed / clock.governor_recovered (spec 028 FR-008): the
	// governor's speed-ladder decisions, one line each in the clock.degraded
	// line's style — the notch transition plus the debt/jobs arithmetic that
	// justified it (contracts/status-protocol.md "TUI" §). requested is
	// omitted here (unlike the header) since from→to already carries the
	// interesting delta and every other clock.* digest row stays this terse.
	"clock.governor_shed": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.GovernorPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{
			txt("governor shed "), emph(string(p.From) + "→" + string(p.To)), txt(" "),
		}, labeled(fmt.Sprintf("debt=%d%%", debtPercent(p.Debt)), fmt.Sprintf("jobs=%d", p.Jobs))), true
	},
	"clock.governor_recovered": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.GovernorPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{
			txt("governor recovered "), emph(string(p.From) + "→" + string(p.To)), txt(" "),
		}, labeled(fmt.Sprintf("debt=%d%%", debtPercent(p.Debt)), fmt.Sprintf("jobs=%d", p.Jobs))), true
	},
	// daemon.started/stopped: "labeled dump of payload fields" (contract §3)
	// — verified against internal/daemon/daemon.go's appendDaemonEvent calls.
	"daemon.started": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.DaemonStartedPayload](e)
		if !ok {
			return nil, false
		}
		return labeled(fmt.Sprintf("tick=%d", p.Tick), fmt.Sprintf("recovery_ms=%d", p.RecoveryMs)), true
	},
	"daemon.stopped": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.DaemonStoppedPayload](e)
		if !ok {
			return nil, false
		}
		return labeled(fmt.Sprintf("tick=%d", p.Tick)), true
	},
	// daemon.llm_warning (spec 034/038): the provider-health preflight
	// transition — a raise/reclassify (Active true) carries the detail (and
	// remedy, when the daemon supplied one); a clear (Active false) is terse.
	"daemon.llm_warning": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.LLMWarningPayload](e)
		if !ok {
			return nil, false
		}
		if !p.Active {
			return labeled("provider="+p.Provider, "kind="+p.Kind, "cleared"), true
		}
		pairs := []string{"provider=" + p.Provider, "kind=" + p.Kind, "warning", "detail=" + p.Detail}
		if p.Remedy != "" {
			pairs = append(pairs, "remedy="+p.Remedy)
		}
		return labeled(pairs...), true
	},
	// run.ended (spec 044 US1): the run-over declaration — total deaths and
	// the final death's cause, the summary a postmortem reader wants on the
	// feed line; the full ledger lives in the payload/detail pane.
	"run.ended": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.RunEndedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{
			txt("the run ended · "), emphN(len(p.Deaths)), txt(" dead · final cause "), emph(p.FinalCause),
		}), true
	},

	// --- sim ---

	"sim.day_started": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.DayPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{txt("day "), emphI64(p.Day), txt(" begins")}), true
	},
	"sim.night_started": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.DayPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{txt("night falls on day "), emphI64(p.Day)}), true
	},
	"sim.forage_regrown": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.RegrownPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{txt("forage regrew at "), coord(p.X, p.Y)}), true
	},
	"sim.fire_burned_out": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.FireBurnedOutPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{txt("the fire at "), coord(p.X, p.Y), txt(" burned out")}), true
	},
	"sim.food_rotted": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.FoodRottedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{emphN(p.N), txt(" "), emph(p.Kind), txt(" rotted at "), coord(p.X, p.Y)}), true
	},
	"sim.gathering_observed": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.GatheringObservedPayload](e)
		if !ok {
			return nil, false
		}
		if p.X == 0 && p.Y == 0 && p.Start == 0 {
			return []seg{txt("gathering dispersed")}, true
		}
		return join([]seg{txt("gathering at "), coord(p.X, p.Y), txt(" since tick "), emphI64(p.Start)}), true
	},

	// --- agent: acts & needs ---

	"agent.intent_set": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.IntentSetPayload](e)
		if !ok {
			return nil, false
		}
		out := join([]seg{refSeg(names, p.Agent), txt(" intends "), emph(p.Goal), txt(" (" + p.Source + ")")})
		// Presence heuristic (implementer decision, no sentinel in the
		// payload): a nonzero target coordinate is treated as "target set".
		if p.TargetX != 0 || p.TargetY != 0 {
			out = join(out, []seg{txt(" → "), coord(p.TargetX, p.TargetY)})
		}
		return out, true
	},
	"agent.work_started": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.WorkStartedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Agent), txt(" set to work")}), true
	},
	"agent.intent_done": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.AgentPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Agent), txt(" finished")}), true
	},
	"agent.build_failed": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		// Spec 038: a cancelled build renders as a FAILURE naming the builder,
		// the goal, and the reason — visibly distinct from "finished" so an
		// observer can never mistake a failed build for a completed one.
		p, ok := decode[sim.BuildFailedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{
			refSeg(names, p.Agent), txt("'s "), emph(p.Goal), txt(" failed — "), emph(p.Reason),
		}), true
	},
	"agent.intent_failed": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		// Spec 096: the agent.build_failed pattern generalized to every
		// non-build goal — a distinct FAILURE naming the actor, the goal, and
		// the reason, visibly distinct from "finished" (digest.go:341 above).
		p, ok := decode[sim.IntentFailedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{
			refSeg(names, p.Agent), txt("'s "), emph(p.Goal), txt(" failed — "), emph(p.Reason),
		}), true
	},
	"agent.intent_rejected": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.IntentRejectedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{
			refSeg(names, p.Agent), txt("'s "), emph(p.Goal), txt(" refused: "), emph(p.Reason),
			txt(" ("), emphI64(p.StalenessTicks), txt("t stale)"),
		}), true
	},
	"agent.recovery_stalled": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		// Spec 064 R4: a needs-conditioned recovery hold that showed no net gain
		// across the full stall window — an honest abort, not a completion, so
		// it renders distinctly from "finished" (the agent.build_failed
		// precedent, digest.go:308 above).
		p, ok := decode[sim.RecoveryStalledPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{
			refSeg(names, p.Agent), txt("'s "), emph(p.Goal), txt(" stalled — "), emph(p.Need), txt(" not recovering"),
		}), true
	},
	"agent.moved": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.AgentMovedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Agent), txt(" → "), coord(p.X, p.Y)}), true
	},
	// agent.saw (spec 041) summarizes the perception diff by its first
	// (canonically-ordered) fact plus a count — a full fact list would flood
	// the feed line; the detail pane holds the payload.
	"agent.saw": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.SawPayload](e)
		if !ok || len(p.Facts) == 0 {
			return nil, false
		}
		f := p.Facts[0]
		segs := []seg{refSeg(names, p.Agent), txt(" saw "), emph(f.Kind), txt(" at "), coord(f.X, f.Y)}
		if more := len(p.Facts) - 1; more > 0 {
			segs = append(segs, txt(" (+"), emphN(more), txt(" more)"))
		}
		return join(segs), true
	},
	// social.place_told (spec 041 US5): directions passed on a talk — the
	// agent.saw first-fact-plus-count shape.
	"social.place_told": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.PlaceToldPayload](e)
		if !ok || len(p.Facts) == 0 {
			return nil, false
		}
		f := p.Facts[0]
		segs := []seg{refSeg(names, p.From), txt(" told "), refSeg(names, p.To),
			txt(" of "), emph(f.Kind), txt(" at "), coord(f.X, f.Y)}
		if more := len(p.Facts) - 1; more > 0 {
			segs = append(segs, txt(" (+"), emphN(more), txt(" more)"))
		}
		return join(segs), true
	},
	// agent.map_corrected (spec 041 US3): the believe-act-discover moment —
	// same first-fact-plus-count shape as agent.saw.
	"agent.map_corrected": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.MapCorrectedPayload](e)
		if !ok || len(p.Gone) == 0 {
			return nil, false
		}
		f := p.Gone[0]
		segs := []seg{refSeg(names, p.Agent), txt(" found "), emph(f.Kind), txt(" at "), coord(f.X, f.Y), txt(" gone")}
		if more := len(p.Gone) - 1; more > 0 {
			segs = append(segs, txt(" (+"), emphN(more), txt(" more)"))
		}
		return join(segs), true
	},
	// agent.place_observed (spec 097): the grounded arrival observation —
	// "went there, this is what IS there". The empty set renders explicitly
	// ("nothing notable"): the perception of absence is the readable evidence
	// of why a myth faded (US2).
	"agent.place_observed": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.PlaceObservedPayload](e)
		if !ok {
			return nil, false
		}
		segs := []seg{refSeg(names, p.Agent), txt(" looked around "), coord(p.X, p.Y), txt(": ")}
		if len(p.Kinds) == 0 {
			segs = append(segs, emph("nothing notable"))
		} else {
			segs = append(segs, emph(strings.Join(p.Kinds, ", ")))
		}
		return join(segs), true
	},
	"agent.foraged": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.HarvestPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Agent), txt(" foraged at "), coord(p.X, p.Y)}), true
	},
	"agent.chopped": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.HarvestPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Agent), txt(" chopped wood at "), coord(p.X, p.Y)}), true
	},
	"agent.hunted": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.HarvestPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Agent), txt(" hunted at "), coord(p.X, p.Y)}), true
	},
	"agent.quarried": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.HarvestPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Agent), txt(" quarried stone at "), coord(p.X, p.Y)}), true
	},
	"agent.collected_water": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.HarvestPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Agent), txt(" drew water at "), coord(p.X, p.Y)}), true
	},
	"agent.crafted": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.CraftedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Agent), txt(" crafted "), emph(p.Kind)}), true
	},
	"agent.built": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.BuiltPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Agent), txt(" built a "), emph(p.Kind), txt(" at "), coord(p.X, p.Y)}), true
	},
	// agent.wall_chipped / agent.wall_destroyed / agent.wall_repaired (spec 032
	// US1) share the {agent, x, y} WallWorkPayload shape — one per
	// demolish/repair work cycle at the wall's own tile.
	"agent.wall_chipped": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.WallWorkPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Agent), txt(" chipped away at the wall at "), coord(p.X, p.Y)}), true
	},
	"agent.wall_destroyed": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.WallWorkPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Agent), txt(" tore down the wall at "), coord(p.X, p.Y)}), true
	},
	"agent.wall_repaired": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.WallWorkPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Agent), txt(" repaired the wall at "), coord(p.X, p.Y)}), true
	},
	"agent.dropped": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.DroppedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Agent), txt(" dropped "), emphN(p.N), txt(" "), emph(p.Kind), txt(" at "), coord(p.X, p.Y)}), true
	},
	"agent.picked_up": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.PickedUpPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Agent), txt(" picked up "), emphN(p.N), txt(" "), emph(p.Kind), txt(" at "), coord(p.X, p.Y)}), true
	},
	"agent.deposited": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.DepositedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{
			refSeg(names, p.Agent), txt(" stored "), emphN(p.N), txt(" "), emph(p.Kind),
			txt(" in the chest at "), coord(p.X, p.Y),
		}), true
	},
	"agent.withdrew": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.WithdrewPayload](e)
		if !ok {
			return nil, false
		}
		out := join([]seg{refSeg(names, p.Agent), txt(" took "), emphN(p.N), txt(" "), emph(p.Kind), txt(" from ")})
		if p.Owner == p.Agent {
			out = join(out, []seg{txt("their chest")})
		} else {
			out = join(out, []seg{refSeg(names, p.Owner), txt("'s chest")})
		}
		return out, true
	},
	"agent.cooked": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.CookedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{
			refSeg(names, p.Agent), txt(" cooked "), emphN(p.Produced), txt(" "), emph(p.Kind),
			txt(" at the "), emph(p.Station),
		}), true
	},
	"agent.bathed": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.BathedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{
			refSeg(names, p.Agent), txt(" bathed · morale "), emphN(p.MoraleAfter),
			txt(" warmth "), emphN(p.WarmthAfter),
		}), true
	},
	"agent.refueled": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.RefueledPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Agent), txt(" refueled the fire at "), coord(p.X, p.Y)}), true
	},
	"agent.spear_broke": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.SpearBrokePayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Agent), txt("'s spear broke")}), true
	},
	// agent.axe_broke (spec 032 US2): the SpearBrokePayload clone, co-emitted
	// alongside a chop/quarry completion when the pre-event carried axe spent
	// its last use — voice mirrors agent.spear_broke's.
	"agent.axe_broke": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.AxeBrokePayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Agent), txt("'s axe broke")}), true
	},
	"agent.ate": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.AtePayload](e)
		if !ok {
			return nil, false
		}
		var parts []string
		if p.Meals > 0 {
			parts = append(parts, fmt.Sprintf("%d meals", p.Meals))
		}
		if p.Cooked > 0 {
			parts = append(parts, fmt.Sprintf("%d cooked", p.Cooked))
		}
		if p.Raw > 0 {
			parts = append(parts, fmt.Sprintf("%d raw", p.Raw))
		}
		breakdown := "nothing"
		if len(parts) > 0 {
			breakdown = strings.Join(parts, ", ")
		}
		return join([]seg{refSeg(names, p.Agent), txt(" ate "), emph(breakdown), txt(" → food "), emphN(p.FoodAfter)}), true
	},
	"agent.slept": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.AgentPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Agent), txt(" fell asleep")}), true
	},
	"agent.woke": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.AgentPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Agent), txt(" woke")}), true
	},
	// agent.needs_changed: NeedsPayload's actual fields are health/food/rest/
	// warmth/morale (no "water" field — the contract's illustrative example
	// named one that isn't in the struct; this renders the real fields).
	"agent.needs_changed": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.NeedsPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Agent), txt(" ")}, labeled(
			fmt.Sprintf("health=%d", p.Health), fmt.Sprintf("food=%d", p.Food),
			fmt.Sprintf("rest=%d", p.Rest), fmt.Sprintf("warmth=%d", p.Warmth),
			fmt.Sprintf("morale=%d", p.Morale),
		)), true
	},
	"agent.died": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) { // alert
		p, ok := decode[sim.DiedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Agent), txt(" died: "), emph(p.Cause)}), true
	},
	"agent.talked": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.TalkedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.A), txt(" chatted with "), refSeg(names, p.B)}), true
	},

	// --- agent: mind & plans ---

	"agent.memory_added": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.MemoryAddedPayload](e)
		if !ok {
			return nil, false
		}
		out := join([]seg{refSeg(names, p.Agent), txt(" remembers: "), speech(p.Text)})
		if p.Subject.ID >= 0 { // sentinel -1 = no gossip subject (internal/sim/memory.go)
			out = join(out, []seg{txt(" · about "), refSeg(names, p.Subject)})
		}
		return out, true
	},
	"agent.thought": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.ThoughtPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Agent), txt(" thought: "), speech(p.Text), txt(" (" + p.Source + ")")}), true
	},
	// agent.memory_promoted / agent.memory_faded: the real payload carries
	// TextHash + MemTick, never the memory's text (internal/sim/consolidate.go)
	// — the contract's quoted "{text}" isn't renderable from this payload, so
	// these digests reference the memory by its tick instead.
	"agent.memory_promoted": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.MemoryPromotedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Agent), txt("'s memory (t"), emphI64(p.MemTick), txt(") reinforced")}), true
	},
	"agent.memory_faded": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.MemoryFadedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Agent), txt(" forgot a memory (t"), emphI64(p.MemTick), txt(")")}), true
	},
	"agent.belief_revised": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.BeliefRevisedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Agent), txt(" now believes: "), speech(p.Statement)}), true
	},
	// agent.belief_reinforced (spec 030 US2, FR-008): re-anchors a held belief's
	// decay clock. Spec 097's grounded-observation channel is the in-tree
	// producer: a Kind-stamped payload says which way the observation moved the
	// belief and carries the new stored confidence — rendered so the myth's
	// fade is readable from the feed. The payload never carries the statement
	// text, so this digest references the belief by id — the same
	// memory_promoted/memory_faded precedent above; the legacy bare shape keeps
	// its pre-097 line verbatim.
	"agent.belief_reinforced": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.BeliefReinforcedPayload](e)
		if !ok {
			return nil, false
		}
		switch p.Kind {
		case sim.BeliefConfirmed:
			return join([]seg{refSeg(names, p.Agent), txt("'s belief (#"), emphN(p.BeliefID), txt(") confirmed by observation → "), emphN(p.Confidence)}), true
		case sim.BeliefDisconfirmed:
			return join([]seg{refSeg(names, p.Agent), txt("'s belief (#"), emphN(p.BeliefID), txt(") disconfirmed by observation → "), emphN(p.Confidence)}), true
		}
		return join([]seg{refSeg(names, p.Agent), txt("'s belief (#"), emphN(p.BeliefID), txt(") reinforced")}), true
	},
	"agent.narrative_set": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.NarrativeSetPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Agent), txt("'s story: "), speech(p.Text)}), true
	},
	"agent.consolidated": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.ConsolidatedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Agent), txt(" consolidated the night's memories")}), true
	},
	"agent.plan_set": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.PlanSetPayload](e)
		if !ok {
			return nil, false
		}
		goals := make([]string, len(p.Steps))
		for i, st := range p.Steps {
			goals[i] = st.Goal
		}
		return join([]seg{
			refSeg(names, p.Agent), txt(" planned "), emphN(len(p.Steps)), txt(" steps: "),
			emph(truncateRunes(strings.Join(goals, ", "), 60)),
		}), true
	},
	"agent.plan_step_started": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.PlanStepPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Agent), txt(" began step "), emph(p.Step)}), true
	},
	"agent.plan_expired": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.PlanStepPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Agent), txt("'s plan lapsed ("), emph(p.Reason), txt(")")}), true
	},

	// --- social ---

	"social.conversation_turn": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) { // speech privilege
		p, ok := decode[sim.ConversationTurnPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Speaker), txt("→"), refSeg(names, p.Listener), txt(" "), speech(p.Text)}), true
	},
	"social.rumor_told": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) { // speech privilege
		p, ok := decode[sim.RumorToldPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.From), txt("→"), refSeg(names, p.To), txt(" rumor: "), speech(p.Text)}), true
	},
	"social.conversation": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) { // tones elided (detail pane)
		p, ok := decode[sim.ConversationPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{speech(p.Gist), txt(" · "), emphN(p.Turns), txt(" turns")}), true
	},
	// social.relation_changed: the payload carries two deltas (trust,
	// affection), not the contract's single "{delta:+}" — both render.
	"social.relation_changed": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.RelationChangedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{
			refSeg(names, p.A), txt("→"), refSeg(names, p.B), txt(" "),
			emph(fmt.Sprintf("trust%+d/affection%+d", p.TrustDelta, p.AffectionDelta)),
			txt(" ("), emph(p.Reason), txt(")"),
		}), true
	},
	// social.gave: GavePayload has no amount field (internal/sim/social.go)
	// — the contract's "{n}" isn't renderable; the kind alone renders.
	"social.gave": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.GavePayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.From), txt(" gave "), refSeg(names, p.To), txt(" "), emph(p.Kind)}), true
	},
	// social.promise_broken: PromiseBrokenPayload carries only an ID, no
	// from/to (internal/sim/social.go) — the contract's "{from} broke a
	// promise to {to}" isn't renderable from this payload; the id renders.
	"social.promise_broken": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.PromiseBrokenPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{txt("a promise was broken (#"), emphN(p.ID), txt(")")}), true
	},
	"social.secret_seeded": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.SecretSeededPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{txt("a secret took root with "), refSeg(names, p.Agent)}), true
	},
	"social.chest_taken": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) { // alert
		p, ok := decode[sim.ChestTakenPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Taker), txt(" raided "), refSeg(names, p.Owner), txt("'s chest at "), coord(p.X, p.Y)}), true
	},
	"social.hailed": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.HailedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.From), txt(" hailed "), refSeg(names, p.To), txt(" (until t"), emphI64(p.Until), txt(")")}), true
	},
	"social.hail_met": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.HailMetPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.From), txt(" met "), refSeg(names, p.To)}), true
	},
	"social.hail_expired": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.HailExpiredPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.From), txt("'s hail to "), refSeg(names, p.To), txt(" lapsed")}), true
	},

	// --- governance (meeting.* / norm.*) ---

	// meeting.convened: MeetingPlacePayload carries only the place, no
	// agents list (internal/sim/governance.go) — the contract's "+ agents
	// per payload" isn't renderable from this payload.
	"meeting.convened": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.MeetingPlacePayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{txt("meeting convened at "), coord(p.X, p.Y)}), true
	},
	"meeting.opened": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		return []seg{txt("meeting opened")}, true
	},
	"meeting.turn_taken": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.TurnTakenPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Agent), txt(" spoke at the meeting")}), true
	},
	"meeting.proposal_tabled": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.ProposalPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Proposer), txt(" proposed: "), speech(p.Text)}), true
	},
	"meeting.proposal_resolved": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.ProposalResolvedPayload](e)
		if !ok {
			return nil, false
		}
		outcome := "failed"
		if p.Passed {
			outcome = "passed"
		}
		out := join([]seg{txt("proposal "), emph(outcome), txt(": "), speech(p.Text)})
		if len(p.Yeas)+len(p.Nays) > 0 {
			out = join(out, []seg{txt(" ("), emph(fmt.Sprintf("%d-%d", len(p.Yeas), len(p.Nays))), txt(")")})
		}
		return out, true
	},
	"meeting.proposal_rephrased": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.ProposalRephrasedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{txt("norm rephrased: "), speech(p.Text)}), true
	},
	"meeting.closed": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		return []seg{txt("meeting closed")}, true
	},
	"meeting.place_designated": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.MeetingPlacePayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{txt("meeting place set at "), coord(p.X, p.Y)}), true
	},
	"meeting.convention_established": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.MeetingConventionPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{
			txt("meeting convention: "), emph(clock.FormatTOD(p.OpenSecond)), txt(" at "), coord(p.X, p.Y),
			txt(" (" + p.Source + ")"),
		}), true
	},
	// norm.violated: NormViolatedPayload carries NormID, not the norm's text
	// (internal/sim/governance.go) — the contract's quoted "{norm text}"
	// isn't renderable from this payload; the norm id renders instead.
	"norm.violated": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) { // alert
		p, ok := decode[sim.NormViolatedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Violator), txt(" violated a norm (#"), emphN(p.NormID), txt(")")}), true
	},

	// --- gru / chronicle / guardian ---

	"gru.emerged": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.GruEmergedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{txt("the gru emerged at "), coord(p.X, p.Y)}), true
	},
	"gru.moved": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.GruMovedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{txt("the gru prowls to "), coord(p.X, p.Y)}), true
	},
	"gru.sighted": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.GruSightedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Agent), txt(" sighted the gru")}), true
	},
	"gru.attacked": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) { // alert
		p, ok := decode[sim.GruAttackedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{txt("the gru attacked "), refSeg(names, p.Agent), txt(" · health → "), emphN(p.Health)}), true
	},
	"gru.withdrew": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		return []seg{txt("the gru withdrew")}, true
	},

	// --- spec 077: the stranger (gru-family threat voice) + incident kinds ---

	"stranger.arrived": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.StrangerArrivedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{txt("a stranger slipped in at "), coord(p.X, p.Y)}), true
	},
	"stranger.moved": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.StrangerMovedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{txt("the stranger creeps to "), coord(p.X, p.Y)}), true
	},
	"stranger.took": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) { // alert (theft tier)
		p, ok := decode[sim.StrangerTookPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{txt("the stranger took "), emphN(p.N), txt(" " + p.Kind + " from the stores at "), coord(p.X, p.Y)}), true
	},
	"stranger.departed": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.StrangerDepartedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{txt("the stranger was gone by dawn of day "), emphI64(p.Day)}), true
	},
	// sim.cold_snap / sim.forage_blighted (spec 077 US2): the two weather-
	// shaped incident kinds, sim-family voice. The blight uses the
	// agent.saw first-fact-plus-count shape for its tile list.
	"sim.cold_snap": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.ColdSnapPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{txt("a cold snap grips night "), emphI64(p.Night), txt(" (until t"), emphI64(p.UntilTick), txt(")")}), true
	},
	"sim.forage_blighted": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.ForageBlightedPayload](e)
		if !ok || len(p.Tiles) == 0 {
			return nil, false
		}
		segs := []seg{txt("blight struck the forage at "), coord(p.Tiles[0].X, p.Tiles[0].Y)}
		if more := len(p.Tiles) - 1; more > 0 {
			segs = append(segs, txt(" (+"), emphN(more), txt(" more tiles)"))
		}
		return join(segs), true
	},
	// sim.neglect_detected (spec 083): the death-by-neglect percept — a
	// survival need below its danger band for the neglect window with zero
	// intents in its class. Deterministic per-need wording: name + peril +
	// inaction + the pre-tick level.
	"sim.neglect_detected": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) { // alert (neglect tier — spec 083)
		p, ok := decode[sim.NeglectDetectedPayload](e)
		if !ok {
			return nil, false
		}
		var peril string
		switch p.Need {
		case "food":
			peril = " is starving and has done nothing about it ("
		case "warmth":
			peril = " is dangerously cold and has done nothing about it ("
		case "rest":
			peril = " is exhausted and has done nothing about it ("
		default:
			return nil, false
		}
		return join([]seg{refSeg(names, p.Agent), txt(peril + p.Need + " "), emphN(p.Level), txt(")")}), true
	},
	"chronicle.entry": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.ChronicleEntryPayload](e)
		if !ok {
			return nil, false
		}
		out := join([]seg{txt("day "), emphI64(p.Day)})
		if p.Thread != "" {
			out = join(out, []seg{txt(" · " + p.Thread)})
		}
		out = join(out, []seg{txt(": "), txt(truncateRunes(p.Text, 80))})
		return out, true
	},
	"guardian.charge_regenerated": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		return []seg{txt("a charge regenerated")}, true
	},
	"guardian.nudged": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.GuardianNudgedPayload](e)
		if !ok {
			return nil, false
		}
		targets := make([]seg, 0, len(p.Targets)*2)
		for i, t := range p.Targets {
			if i > 0 {
				targets = append(targets, txt(", "))
			}
			targets = append(targets, refSeg(names, t))
		}
		return join([]seg{txt(sk.Name() + " "), emph(sk.FormNoun(p.Form)), txt(" → ")}, targets, []seg{txt(": "), speech(p.Text)}), true
	},
	// guardian.place_revealed (spec 041 FR-014): a vision's divine place
	// grant — the agent.saw first-fact-plus-count shape, Guardian as subject
	// (the nudge convention).
	"guardian.place_revealed": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.PlaceRevealedPayload](e)
		if !ok || len(p.Facts) == 0 {
			return nil, false
		}
		f := p.Facts[0]
		segs := []seg{txt(sk.Name() + " revealed "), emph(f.Kind), txt(" at "), coord(f.X, f.Y),
			txt(" to "), refSeg(names, p.Agent)}
		if more := len(p.Facts) - 1; more > 0 {
			segs = append(segs, txt(" (+"), emphN(more), txt(" more)"))
		}
		return join(segs), true
	},
	// guardian.order_placed / order_triggered / order_cancelled / order_expired
	// (spec 029, TASK-27 wiki-sweep gap): the standing-order lifecycle
	// (internal/sim/guardian.go, [[guardian-orders]]) predates this contract
	// (specs/018) same as the miracle types below, so voice mirrors
	// guardian.nudged's — "Guardian" as subject regardless of Origin, since
	// monitor_and_act/cancel_order are Guardian's own tools whether a player
	// or the system asked for the watch (GuardianOrder.Origin distinguishes
	// who, never how it renders). order_triggered/cancelled/expired carry no
	// condition text (internal/sim/guardian.go's OrderTriggeredPayload /
	// OrderIDPayload), only the order's id, so they reference the watch by
	// id rather than repeating its condition.
	"guardian.order_placed": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.OrderPlacedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{
			txt(sk.Name() + " set a watch: "), speech(truncateRunes(p.Condition, 80)),
		}), true
	},
	"guardian.order_triggered": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.OrderTriggeredPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{
			txt(sk.Name() + "'s watch came true ("), emph(p.MatchedType), txt(" @ t"), emphI64(p.MatchedTick), txt(")"),
		}), true
	},
	"guardian.order_cancelled": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.OrderIDPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{txt(sk.Name() + " released a watch ("), emph(p.ID), txt(")")}), true
	},
	"guardian.order_expired": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.OrderIDPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{txt(sk.Name() + "'s watch lapsed ("), emph(p.ID), txt(")")}), true
	},
	// guardian.charter_observed (spec 044 US2): the charter-revision
	// fingerprint stamp a turn ran under — the guardian's evidence timeline the
	// morgue aligns deaths against.
	"guardian.charter_observed": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.CharterObservedPayload](e)
		if !ok {
			return nil, false
		}
		prov := "player-authored"
		if p.Default {
			prov = "default"
		}
		return join([]seg{txt(sk.Name() + " ran under charter "), emph(p.Fingerprint), txt(" (" + prov + ")")}), true
	},
	// guardian.skills_observed (spec 077 FR-006): the skills twin of the
	// charter observation above — the bound set's size and fingerprint (the
	// names ride the payload; the inspector shows them verbatim).
	"guardian.skills_observed": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.SkillsObservedPayload](e)
		if !ok {
			return nil, false
		}
		noun := " skill files "
		if len(p.Names) == 1 {
			noun = " skill file "
		}
		return join([]seg{txt(sk.Name() + " ran under "), emphN(len(p.Names)), txt(noun), emph(p.Fingerprint)}), true
	},
	// designation.* / directive.* (spec 084): the guardian's plan layer —
	// designations (world plan artifacts) and directives (hard villager
	// bindings). Voice mirrors the standing-order rows above: the guardian is
	// the subject for injected acts; the executor-emitted terminals
	// (fulfilled/expired) read as the world answering the plan. Cancelled/
	// fulfilled/expired carry only ids (OrderIDPayload / the fulfilled seam
	// payload), so they reference the entity by id.
	"designation.placed": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.Designation](e)
		if !ok {
			return nil, false
		}
		segs := []seg{txt(sk.Name() + " marked a "), emph(p.Kind), txt(" at "), coord(p.X, p.Y)}
		if p.Kind != sim.DesignationStructureSite {
			segs = append(segs, txt(".."), coord(p.X2, p.Y2))
		}
		if p.StructureKind != "" {
			segs = append(segs, txt(" ("), emph(p.StructureKind), txt(")"))
		}
		if p.Label != "" {
			segs = append(segs, txt(" — "), speech(truncateRunes(p.Label, 40)))
		}
		return join(segs), true
	},
	"designation.cancelled": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.OrderIDPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{txt(sk.Name() + " withdrew a designation ("), emph(p.ID), txt(")")}), true
	},
	"designation.fulfilled": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.OrderIDPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{txt("the village fulfilled " + sk.Name() + "'s mark ("), emph(p.ID), txt(")")}), true
	},
	"directive.issued": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.DirectiveIssuedPayload](e)
		if !ok {
			return nil, false
		}
		segs := []seg{txt(sk.Name() + " charged ")}
		for i, tgt := range p.Targets {
			if i > 0 {
				segs = append(segs, txt(", "))
			}
			segs = append(segs, refSeg(names, tgt))
		}
		segs = append(segs, txt(": "), speech(truncateRunes(p.Text, 80)))
		return join(segs), true
	},
	"directive.cancelled": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.OrderIDPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{txt(sk.Name() + " lifted a charge ("), emph(p.ID), txt(")")}), true
	},
	"directive.fulfilled": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.DirectiveFulfilledPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{txt("the village fulfilled " + sk.Name() + "'s charge ("), emph(p.ID),
			txt(", serving "), emph(p.DesignationID), txt(")")}), true
	},
	"directive.expired": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.OrderIDPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{txt(sk.Name() + "'s charge lapsed ("), emph(p.ID), txt(")")}), true
	},
	// faith.changed / prophecy.* (spec 085): the faith economy — devotion and
	// doubt, never points or numbers first (the overjustification caution);
	// the reason rides as the mechanical footnote. Prophecy rows mirror the
	// standing-order voice: the guardian declares, the world answers.
	"faith.changed": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.FaithChangedPayload](e)
		if !ok {
			return nil, false
		}
		phrase := "the village's faith deepens"
		if p.Delta < 0 {
			phrase = "faith wavers in the village"
		}
		return join([]seg{txt(phrase + " ("), emph(p.Reason), txt(")")}), true
	},
	"prophecy.declared": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.ProphecyDeclaredPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{txt(sk.Name() + " foretells: "), speech(truncateRunes(p.Text, 80))}), true
	},
	"prophecy.fulfilled": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.OrderIDPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{txt(sk.Name() + "'s foretelling came true ("), emph(p.ID), txt(")")}), true
	},
	"prophecy.failed": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.OrderIDPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{txt(sk.Name() + "'s word did not come to pass ("), emph(p.ID), txt(")")}), true
	},
	// morgue.epilogue (spec 044 US2): the narrator's recorded mourning prose
	// — agent -1 is the run-end epilogue. chronicle.entry's truncation manner.
	"morgue.epilogue": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.MorgueEpiloguePayload](e)
		if !ok {
			return nil, false
		}
		who := []seg{txt("the run")}
		if p.Agent.ID >= 0 {
			who = []seg{refSeg(names, p.Agent)}
		}
		return join([]seg{txt("epilogue for ")}, who, []seg{txt(": "), txt(truncateRunes(p.Text, 80))}), true
	},

	// guardian.time_snapped / item_granted / entity_moved / entity_removed
	// (TASK-59, spec 016) predate this contract (specs/018) — no template
	// row exists for them, so voice/style mirrors guardian.nudged's (natural
	// phrase, "Guardian" as subject); gratisMark surfaces the operator force
	// SC-004 requires be enumerable, never silently indistinguishable from a
	// charge-priced miracle.
	"guardian.time_snapped": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.TimeSnappedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{
			txt(sk.Name() + " snapped time forward to "), emph(clock.Format(p.ToTick)),
		}, gratisMark(p.Gratis)), true
	},
	"guardian.item_granted": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.ItemGrantedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{
			txt(sk.Name() + " granted "), refSeg(names, p.Agent), txt(" "), emphN(p.Qty), txt(" "), emph(p.Kind),
		}, gratisMark(p.Gratis)), true
	},
	// entity_moved: the payload identifies its target by class + source
	// coordinates only (internal/sim/miracles.go) — no agent index, so a
	// moved villager renders by its (pre-move) location rather than a
	// resolved name.
	"guardian.entity_moved": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.EntityMovedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{
			txt(sk.Name() + " moved the "), emph(p.Class), txt(" at "), coord(p.X, p.Y), txt(" to "), coord(p.ToX, p.ToY),
		}, gratisMark(p.Gratis)), true
	},
	// entity_removed: the payload carries the target's class only, never a
	// structure's Kind (internal/sim/miracles.go) — a removed chest renders
	// as "the structure", not "the chest". A terrain target is overlaid
	// (chop/forage/quarry vocabulary), not deleted, so it reads "cleared"
	// rather than "removed".
	"guardian.entity_removed": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.EntityRemovedPayload](e)
		if !ok {
			return nil, false
		}
		verb := "removed"
		if p.Class == "terrain" {
			verb = "cleared"
		}
		return join([]seg{
			txt(sk.Name() + " " + verb + " the "), emph(p.Class), txt(" at "), coord(p.X, p.Y),
		}, gratisMark(p.Gratis)), true
	},

	// --- cog (labeled) ---

	"cog.thought": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.CogThoughtPayload](e)
		if !ok {
			return nil, false
		}
		return labeled(
			"job="+p.Job, "class="+p.Class, "agent="+refName(names, p.Agent),
			fmt.Sprintf("pts=%d", p.Points), fmt.Sprintf("pred=%dms", p.PredictedWallMs),
		), true
	},
	"cog.outcome": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.CogOutcomePayload](e)
		if !ok {
			return nil, false
		}
		pairs := []string{
			"job=" + p.Job, p.Outcome, "agent=" + refName(names, p.Agent),
			fmt.Sprintf("stale=%dt", p.StalenessTicks), fmt.Sprintf("wall=%dms", p.ActualWallMs),
		}
		if p.Kind != "" {
			pairs = append(pairs, "kind="+p.Kind)
		}
		if p.Reason != "" {
			pairs = append(pairs, "reason="+p.Reason)
		}
		return labeled(pairs...), true
	},
	"cog.recalibration_recommended": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.RecalibrationPayload](e)
		if !ok {
			return nil, false
		}
		// Post-spec-031 events carry the adoption arithmetic; show
		// prior→adopted when present, else the legacy current estimate.
		est := fmt.Sprintf("est=%.2fs/pt", p.EstimateSPerPt)
		if p.AdoptedSPerPt != 0 || p.PriorSPerPt != 0 {
			est = fmt.Sprintf("est=%.2f→%.2fs/pt", p.PriorSPerPt, p.AdoptedSPerPt)
		}
		return labeled(
			"tier="+p.Tier, est,
			fmt.Sprintf("spikes=%.2f", p.SpikeRate), fmt.Sprintf("window=%d", p.Window),
		), true
	},
	// cog.tool_call: Args and SnapshotTick are deliberately elided — the
	// detail pane bounds them, same reasoning as world.migrated's elided
	// state.
	"cog.tool_call": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.CogToolCallPayload](e)
		if !ok {
			return nil, false
		}
		pairs := []string{
			"job=" + p.Job, fmt.Sprintf("ord=%d", p.Ordinal), "tool=" + p.Tool,
			p.Verdict, "tier=" + p.Tier,
		}
		if p.Reason != "" {
			pairs = append(pairs, "reason="+p.Reason)
		}
		return labeled(pairs...), true
	},

	// --- spec 042: embedding companions + divergence telemetry ---
	// The vectors themselves are deliberately elided (384 floats would drown
	// the feed — the world.migrated elided-state reasoning); the digest shows
	// the identity fields an operator audits by.

	"agent.memory_embedded": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.MemoryEmbeddedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Agent), txt(" ")}, labeled(
			fmt.Sprintf("memory seq=%d embedded", p.MemSeq),
			fmt.Sprintf("dims=%d", len(p.Vec)), "model="+p.Model,
		)), true
	},
	"agent.situation_embedded": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.SituationEmbeddedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{refSeg(names, p.Agent), txt(" situation: "), speech(p.Text), txt(" ")},
			labeled(fmt.Sprintf("dims=%d", len(p.Vec)), "model="+p.Model)), true
	},
	"cog.memory_divergence": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.MemoryDivergencePayload](e)
		if !ok {
			return nil, false
		}
		return labeled(
			"agent="+refName(names, p.Agent), "mode="+p.Mode,
			fmt.Sprintf("overlap=%d/%d", p.Overlap, len(p.Legacy)),
			fmt.Sprintf("displaced=%d", p.Displacement),
			fmt.Sprintf("vectorless=%d", p.Vectorless),
		), true
	},

	// --- spec 046: the curriculum ladder — exercise passes + stage unlocks ---
	// Natural-phrase voice (Guardian's own family tint, grammar.go), mirroring
	// the order lifecycle rows above: the guardian is the subject regardless
	// of what emitted the pass (TASK-119's rubric machinery in production).

	"curriculum.exercise_passed": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.ExercisePassedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{
			txt("the "), emph(p.Exercise), txt(" exercise was passed ("), emph(p.Stage), txt(")"),
		}), true
	},
	"curriculum.stage_unlocked": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.StageUnlockedPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{
			txt(sk.Name() + "'s watcher earned "), emph(sk.StageName(p.Stage)),
			txt(" (proven by "), emph(p.Exercise), txt(")"),
		}), true
	},

	// guardian.report_card (spec 063 US4): the stored attribution note — the
	// cheap-chain critique the console card seam re-reads. Skin card label
	// (contract §4/D2); morgue.epilogue's truncation manner for the prose.
	"guardian.report_card": func(e store.Event, names []string, sk *skin.Skin) ([]seg, bool) {
		p, ok := decode[sim.GuardianReportCardPayload](e)
		if !ok {
			return nil, false
		}
		return join([]seg{
			txt(sk.ReportCardLabel() + " under charter "), emph(p.Fingerprint),
			txt(": "), txt(truncateRunes(p.Note, 80)),
		}), true
	},
}

// --- jump-to-source subject resolution (spec 049, contract §2/FR-002) ---
//
// subjectCandidate is one event type's payload-level candidate, before the
// live replica is consulted: at most one primary-actor agent index and/or
// one explicit recorded position. Purely a function of the stored payload
// (like digestFunc, for the same bounded-work reason: no recursive scan,
// world.migrated simply has no entry below and so never even reaches
// decode). resolveSubject applies the live-replica step on top of whichever
// half(s) a type's subjectFunc fills in.
type subjectCandidate struct {
	actor    int // primary-actor agent index; meaningful only if hasActor
	hasActor bool
	x, y     int // explicit recorded position; meaningful only if hasPos
	hasPos   bool
	// place is the display name to use when the position is the only
	// candidate (no actor at all) — e.g. "the meeting place". Left "" for
	// actor-bearing types, where resolveSubject names the actor instead.
	place string
}

// subjectFunc decodes one event type's payload into a subjectCandidate;
// ok=false (decode failure) is treated exactly like an unlocatable
// candidate — resolveSubject never panics on a malformed payload, it just
// falls to the honest hint (contract §2 "unlocatable").
type subjectFunc func(e store.Event) (subjectCandidate, bool)

func actorCandidate(idx int) subjectCandidate {
	return subjectCandidate{actor: idx, hasActor: true}
}

func actorPosCandidate(idx, x, y int) subjectCandidate {
	return subjectCandidate{actor: idx, hasActor: true, x: x, y: y, hasPos: true}
}

func placeCandidate(label string, x, y int) subjectCandidate {
	return subjectCandidate{place: label, x: x, y: y, hasPos: true}
}

// decodeHarvest is shared by every HarvestPayload{Agent,X,Y} event type
// (foraged/chopped/hunted/quarried/collected_water) — same payload shape,
// only the digest's verb differs.
func decodeHarvest(e store.Event) (subjectCandidate, bool) {
	p, ok := decode[sim.HarvestPayload](e)
	if !ok {
		return subjectCandidate{}, false
	}
	return actorPosCandidate(p.Agent.ID, p.X, p.Y), true
}

// decodeWallWork is shared by the three wall work-cycle events
// (chipped/destroyed/repaired) — WallWorkPayload{Agent,X,Y}.
func decodeWallWork(e store.Event) (subjectCandidate, bool) {
	p, ok := decode[sim.WallWorkPayload](e)
	if !ok {
		return subjectCandidate{}, false
	}
	return actorPosCandidate(p.Agent.ID, p.X, p.Y), true
}

// decodeAgentOnly is shared by every plain AgentPayload{Agent} event type
// (intent_done/slept/woke) — actor only, no recorded position.
func decodeAgentOnly(e store.Event) (subjectCandidate, bool) {
	p, ok := decode[sim.AgentPayload](e)
	if !ok {
		return subjectCandidate{}, false
	}
	return actorCandidate(p.Agent.ID), true
}

// placeFactPos extracts the first PlaceFact's position, if any — the shared
// shape behind agent.saw/social.place_told/agent.map_corrected/
// guardian.place_revealed (all carry a []PlaceFact whose first entry is the
// canonical one, R4).
func placeFactPos(facts []sim.PlaceFact) (x, y int, ok bool) {
	if len(facts) == 0 {
		return 0, 0, false
	}
	return facts[0].X, facts[0].Y, true
}

// subjectRegistry catalogs, per event type, the payload-level actor/position
// candidate (contract §2 step 2's raw material) — resolveSubject applies the
// live-replica step (step 1) on top. A type absent here (every telemetry-only
// type with no agent/position field the digest already reads, plus
// world.migrated by deliberate omission — R4/FR-011) simply resolves
// unlocatable: a registry miss is a legitimate, honest jump-or-hint outcome,
// not a gap to fill in. Multi-agent/ambiguous-subject types (guardian.nudged's
// several targets, chronicle.entry's agent list, meeting.proposal_resolved's
// unnamed subject) are deliberately left out for the same reason FR-002
// requires "one deterministic subject" — no field here is a good-enough
// single answer, so the honest hint wins over a guess.
var subjectRegistry = map[string]subjectFunc{
	// --- sim: agent acts with a recorded position ---
	"agent.moved": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.AgentMovedPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorPosCandidate(p.Agent.ID, p.X, p.Y), true
	},
	"agent.foraged":         decodeHarvest,
	"agent.chopped":         decodeHarvest,
	"agent.hunted":          decodeHarvest,
	"agent.quarried":        decodeHarvest,
	"agent.collected_water": decodeHarvest,
	"agent.built": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.BuiltPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorPosCandidate(p.Agent.ID, p.X, p.Y), true
	},
	"agent.wall_chipped":   decodeWallWork,
	"agent.wall_destroyed": decodeWallWork,
	"agent.wall_repaired":  decodeWallWork,
	"agent.dropped": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.DroppedPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorPosCandidate(p.Agent.ID, p.X, p.Y), true
	},
	"agent.picked_up": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.PickedUpPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorPosCandidate(p.Agent.ID, p.X, p.Y), true
	},
	"agent.deposited": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.DepositedPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorPosCandidate(p.Agent.ID, p.X, p.Y), true
	},
	"agent.withdrew": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.WithdrewPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorPosCandidate(p.Agent.ID, p.X, p.Y), true
	},
	"agent.refueled": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.RefueledPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorPosCandidate(p.Agent.ID, p.X, p.Y), true
	},
	// agent.intent_set: TargetX/TargetY is the same "target set" presence
	// heuristic the digest already applies (nonzero => set) — the recorded
	// position is the intent's target, the same field the feed line shows.
	"agent.intent_set": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.IntentSetPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		if p.TargetX == 0 && p.TargetY == 0 {
			return actorCandidate(p.Agent.ID), true
		}
		return actorPosCandidate(p.Agent.ID, p.TargetX, p.TargetY), true
	},

	// --- perception/place-knowledge: []PlaceFact, first fact is canonical ---
	"agent.saw": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.SawPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		if x, y, has := placeFactPos(p.Facts); has {
			return actorPosCandidate(p.Agent.ID, x, y), true
		}
		return actorCandidate(p.Agent.ID), true
	},
	"social.place_told": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.PlaceToldPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		if x, y, has := placeFactPos(p.Facts); has {
			return actorPosCandidate(p.From.ID, x, y), true
		}
		return actorCandidate(p.From.ID), true
	},
	"agent.map_corrected": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.MapCorrectedPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		if x, y, has := placeFactPos(p.Gone); has {
			return actorPosCandidate(p.Agent.ID, x, y), true
		}
		return actorCandidate(p.Agent.ID), true
	},
	"guardian.place_revealed": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.PlaceRevealedPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		if x, y, has := placeFactPos(p.Facts); has {
			return actorPosCandidate(p.Agent.ID, x, y), true
		}
		return actorCandidate(p.Agent.ID), true
	},
	// agent.place_observed (spec 097): the observer standing at the observed
	// tile — actor + position, the harvest shape.
	"agent.place_observed": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.PlaceObservedPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorPosCandidate(p.Agent.ID, p.X, p.Y), true
	},

	// --- social: the digest's grammatical subject as the jump actor ---
	"social.chest_taken": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.ChestTakenPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorPosCandidate(p.Taker.ID, p.X, p.Y), true
	},
	"social.hailed": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.HailedPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.From.ID), true
	},
	"social.hail_met": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.HailMetPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.From.ID), true
	},
	"social.hail_expired": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.HailExpiredPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.From.ID), true
	},
	"social.conversation_turn": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.ConversationTurnPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.Speaker.ID), true
	},
	"social.rumor_told": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.RumorToldPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.From.ID), true
	},
	"social.conversation": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.ConversationPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.A.ID), true
	},
	"social.relation_changed": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.RelationChangedPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.A.ID), true
	},
	"social.gave": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.GavePayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.From.ID), true
	},
	"social.secret_seeded": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.SecretSeededPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.Agent.ID), true
	},

	// --- governance ---
	"meeting.convened": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.MeetingPlacePayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return placeCandidate("the meeting place", p.X, p.Y), true
	},
	"meeting.place_designated": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.MeetingPlacePayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return placeCandidate("the meeting place", p.X, p.Y), true
	},
	"meeting.convention_established": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.MeetingConventionPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return placeCandidate("the meeting place", p.X, p.Y), true
	},
	"meeting.turn_taken": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.TurnTakenPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.Agent.ID), true
	},
	"meeting.proposal_tabled": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.ProposalPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.Proposer.ID), true
	},
	"norm.violated": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.NormViolatedPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.Violator.ID), true
	},

	// --- gru ---
	"gru.emerged": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.GruEmergedPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return placeCandidate("the gru", p.X, p.Y), true
	},
	"gru.moved": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.GruMovedPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return placeCandidate("the gru", p.X, p.Y), true
	},
	"gru.sighted": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.GruSightedPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorPosCandidate(p.Agent.ID, p.X, p.Y), true
	},
	"gru.attacked": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.GruAttackedPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.Agent.ID), true
	},

	// --- sim: place-only environmental events ---
	"sim.forage_regrown": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.RegrownPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return placeCandidate("the forage patch", p.X, p.Y), true
	},
	"sim.fire_burned_out": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.FireBurnedOutPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return placeCandidate("the fire", p.X, p.Y), true
	},
	"sim.food_rotted": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.FoodRottedPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return placeCandidate("the spoiled food", p.X, p.Y), true
	},
	// sim.gathering_observed: the all-zero payload is the watch-reset
	// sentinel (contract §3 "gathering dispersed", digestRegistry above) —
	// not a real gathering at (0,0), so it's unlocatable exactly like the
	// digest treats it as textless.
	"sim.gathering_observed": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.GatheringObservedPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		if p.X == 0 && p.Y == 0 && p.Start == 0 {
			return subjectCandidate{}, false
		}
		return placeCandidate("the gathering", p.X, p.Y), true
	},

	// --- guardian miracles: place-only, named by the payload's own Class ---
	// entity_moved jumps to the DESTINATION (ToX,ToY) — an implementer
	// judgment call (the payload records both endpoints; "where it ended up"
	// reads as more useful post-jump than "where it used to be").
	"guardian.entity_moved": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.EntityMovedPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return placeCandidate("the "+p.Class, p.ToX, p.ToY), true
	},
	"guardian.entity_removed": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.EntityRemovedPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return placeCandidate("the "+p.Class, p.X, p.Y), true
	},
	"guardian.item_granted": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.ItemGrantedPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.Agent.ID), true
	},

	// --- agent: actor-only (no recorded position anywhere in the payload) ---
	"agent.work_started": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.WorkStartedPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.Agent.ID), true
	},
	"agent.intent_done": decodeAgentOnly,
	"agent.slept":       decodeAgentOnly,
	"agent.woke":        decodeAgentOnly,
	"agent.build_failed": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.BuildFailedPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.Agent.ID), true
	},
	"agent.recovery_stalled": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.RecoveryStalledPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.Agent.ID), true
	},
	"agent.intent_failed": func(e store.Event) (subjectCandidate, bool) {
		// Unlike agent.build_failed (actor-only), IntentFailedPayload carries
		// the actor's own position (spec 096 FR-001) — the harvest/wall-work
		// pattern (actorPosCandidate) applies here instead.
		p, ok := decode[sim.IntentFailedPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorPosCandidate(p.Agent.ID, p.X, p.Y), true
	},
	"agent.intent_rejected": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.IntentRejectedPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.Agent.ID), true
	},
	"agent.crafted": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.CraftedPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.Agent.ID), true
	},
	"agent.cooked": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.CookedPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.Agent.ID), true
	},
	"agent.bathed": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.BathedPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.Agent.ID), true
	},
	"agent.spear_broke": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.SpearBrokePayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.Agent.ID), true
	},
	"agent.axe_broke": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.AxeBrokePayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.Agent.ID), true
	},
	"agent.ate": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.AtePayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.Agent.ID), true
	},
	"agent.needs_changed": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.NeedsPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.Agent.ID), true
	},
	"agent.talked": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.TalkedPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.A.ID), true
	},
	// agent.memory_added: Where (spec 019) is the memory's recorded location
	// at emission, when the emitter knew one — a real, if occasional,
	// second candidate beyond the live position.
	"agent.memory_added": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.MemoryAddedPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		if p.Where != nil {
			return actorPosCandidate(p.Agent.ID, p.Where.X, p.Where.Y), true
		}
		return actorCandidate(p.Agent.ID), true
	},
	"agent.thought": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.ThoughtPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.Agent.ID), true
	},
	"agent.memory_promoted": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.MemoryPromotedPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.Agent.ID), true
	},
	"agent.memory_faded": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.MemoryFadedPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.Agent.ID), true
	},
	"agent.belief_revised": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.BeliefRevisedPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.Agent.ID), true
	},
	"agent.belief_reinforced": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.BeliefReinforcedPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.Agent.ID), true
	},
	"agent.narrative_set": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.NarrativeSetPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.Agent.ID), true
	},
	"agent.consolidated": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.ConsolidatedPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.Agent.ID), true
	},
	"agent.plan_set": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.PlanSetPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.Agent.ID), true
	},
	"agent.plan_step_started": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.PlanStepPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.Agent.ID), true
	},
	"agent.plan_expired": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.PlanStepPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.Agent.ID), true
	},
	"agent.memory_embedded": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.MemoryEmbeddedPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.Agent.ID), true
	},
	"agent.situation_embedded": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.SituationEmbeddedPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.Agent.ID), true
	},

	// --- cog: telemetry, but Agent is still a known top-level field ---
	"cog.thought": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.CogThoughtPayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.Agent.ID), true
	},
	"cog.outcome": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.CogOutcomePayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.Agent.ID), true
	},
	"cog.memory_divergence": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.MemoryDivergencePayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.Agent.ID), true
	},

	// morgue.epilogue: Agent -1 is the run-end epilogue sentinel (no single
	// agent), same convention the digest already reads.
	"morgue.epilogue": func(e store.Event) (subjectCandidate, bool) {
		p, ok := decode[sim.MorgueEpiloguePayload](e)
		if !ok {
			return subjectCandidate{}, false
		}
		if p.Agent.ID < 0 {
			return subjectCandidate{}, false
		}
		return actorCandidate(p.Agent.ID), true
	},
}

// liveAgentPos reports idx's current (X,Y) in replica, gated on the agent
// existing AND being alive (data-model.md: "Dead/despawned actors fall
// through to the payload position, never to a stale coordinate").
func liveAgentPos(replica *sim.State, idx int) (x, y int, alive bool) {
	if replica == nil || idx < 0 || idx >= len(replica.Agents) {
		return 0, 0, false
	}
	a := replica.Agents[idx]
	if a.Dead {
		return 0, 0, false
	}
	return a.X, a.Y, true
}

// resolveSubject implements contract §2/FR-002: the selected event's primary
// actor's live position if the actor exists and is alive; else the event's
// own recorded position, if it carried one; else unlocatable. Bounded to the
// known top-level fields subjectRegistry catalogs per type — a registry
// miss (including world.migrated, deliberately absent) resolves unlocatable
// without ever decoding the payload, the same bounded-work posture
// formatChronicleLine's fallback takes on an unknown type.
func (m Model) resolveSubject(e store.Event) (name string, x, y int, ok bool) {
	fn, known := subjectRegistry[e.Type]
	if !known {
		// Spec 086 FR-008 (the village-lens completion): a registry miss no
		// longer means unlocatable outright — post-086 payloads carry named
		// {id,name} ref objects, so when exactly ONE distinct in-roster
		// agent appears in the payload it is the subject, generically, with
		// zero per-type registry work. Ambiguity (zero or several distinct
		// refs) stays unlocatable — the honest-hint doctrine, now detected
		// structurally instead of by hand-listing.
		return m.resolveSubjectGeneric(e)
	}
	cand, decOK := fn(e)
	if !decOK {
		return "", 0, 0, false
	}
	names := m.agentNames()
	if cand.hasActor {
		if lx, ly, alive := liveAgentPos(m.replica, cand.actor); alive {
			return agentName(names, cand.actor), lx, ly, true
		}
	}
	if cand.hasPos {
		label := cand.place
		if label == "" && cand.hasActor {
			label = agentName(names, cand.actor)
		}
		return label, cand.x, cand.y, true
	}
	return "", 0, 0, false
}

// resolveSubjectGeneric is the registry-miss fallback (spec 086 FR-008):
// scan the raw payload for {"id":N,"name":…} ref objects (the post-086
// AgentRef wire shape); exactly one distinct in-roster id ⇒ that agent is
// the subject (live position, payload name — a dead or unknown actor stays
// unlocatable: no generically trustworthy recorded position exists on a
// registry-miss type). world.migrated stays hard-excluded — its embedded
// State is never scanned (the deliberate registry absence, FR-011).
func (m Model) resolveSubjectGeneric(e store.Event) (name string, x, y int, ok bool) {
	if e.Type == "world.migrated" {
		return "", 0, 0, false
	}
	var v any
	if err := json.Unmarshal(e.Payload, &v); err != nil {
		return "", 0, 0, false
	}
	ids := map[int]string{}
	scanRefObjects(v, ids)
	if len(ids) != 1 {
		return "", 0, 0, false
	}
	for idx, refName := range ids {
		if lx, ly, alive := liveAgentPos(m.replica, idx); alive {
			if refName == "" {
				refName = agentName(m.agentNames(), idx)
			}
			return refName, lx, ly, true
		}
	}
	return "", 0, 0, false
}

// scanRefObjects walks decoded payload JSON collecting AgentRef-shaped
// objects — maps with exactly the two keys "id" (integral number) and
// "name" (string). Only in-roster ids are candidates (sentinels are not
// subjects); distinct ids accumulate so ambiguity is structural.
func scanRefObjects(v any, out map[int]string) {
	switch t := v.(type) {
	case map[string]any:
		if len(t) == 2 {
			idv, hasID := t["id"].(float64)
			namev, hasName := t["name"].(string)
			if hasID && hasName && idv == float64(int(idv)) {
				if id := int(idv); id >= 0 && id < sim.AgentCount {
					out[id] = namev
				}
				return
			}
		}
		for _, mv := range t {
			scanRefObjects(mv, out)
		}
	case []any:
		for _, ev := range t {
			scanRefObjects(ev, out)
		}
	}
}

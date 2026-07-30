package guardian

// The canonization miracle (spec 101): canonize_region — the guardian
// christens a villager-coined toponym as durable world state and,
// optionally, raises ONE feature of an existing placeable kind within it —
// plus its read-only companion, brief_myths, the myth-briefing surface (D5).
// Doctrine mirrors landPlaceDesignation/landMiracle: the reducer dry-run
// (internal/sim/regions.go) is the door authority for every rule — bounds,
// name, overlap, cap, feature site/kind/build-site, and the premium charge —
// and this file only mints the deterministic id, lands through InjectSocial,
// and maps a door rejection to in-fiction counsel.

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/toolloop"
)

// mythBriefingTopN bounds the read-only myth briefing (spec 101 D5) to a
// small, human-scannable set — recomputed on every absorbed batch
// (mirrorState), so a generous topN would only ever cost a bit of extra
// sorting, never staleness; 5 matches the designation/directive prompt
// sections' scale.
const mythBriefingTopN = 5

// RegionReport is the console-facing summary of a landed canonization (spec
// 101): the region id the player can name (informationally — no cancel verb
// exists in v1) and a one-line human rendering. Additive omitempty on
// TurnResult, the PlanReport/OrderReport precedent.
type RegionReport struct {
	ID      string `json:"id"`
	Summary string `json:"summary"`
}

// landCanonizeRegion parses a canonize_region call, mints the id, and lands
// guardian.region_named through the door (spec 101 FR-002). The guardian can
// NEVER waive the charge — gratis is passed false unconditionally and does
// not exist on the tool schema (the work_miracle SC-005 guarantee).
// hasFeature distinguishes "no feature" from "feature at (0,0)" — a genuine
// coordinate the flat arg-reading helpers cannot otherwise signal absent.
func (mt *Guardian) landCanonizeRegion(x, y, radius int, name, featureKind string, featureX, featureY int, hasFeature bool, tick int64, charges int, grant grantSet) (*sim.Region, string) {
	if !grant.allows("canonize_region") {
		return nil, "that power is not granted in this world"
	}
	if charges < sim.GuardianRegionCharge {
		return nil, fmt.Sprintf("canonizing takes %d charges and I hold too few", sim.GuardianRegionCharge)
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, "a region needs a name — what do the villagers call it?"
	}
	featureKind = strings.TrimSpace(featureKind)
	if featureKind != "" && !hasFeature {
		return nil, "a feature needs a site — give me feature_x and feature_y"
	}
	p := sim.RegionNamedPayload{
		ID: mt.nextPlanID("reg", tick), X: x, Y: y, Radius: radius, Name: name,
	}
	if featureKind != "" {
		p.FeatureKind = featureKind
		p.FeatureX, p.FeatureY = featureX, featureY
	}
	batch := []store.Event{{Type: "guardian.region_named", Payload: mustJSON(p)}}
	if err := mt.social.InjectSocial(batch); err != nil {
		log.Printf("guardian: canonization rejected at the door: %v", err)
		return nil, canonizeRefusal(err)
	}
	r := &sim.Region{ID: p.ID, X: x, Y: y, Radius: radius, Name: name, PlacedTick: tick}
	summary := fmt.Sprintf("named %q at (%d,%d), radius %d", name, x, y, radius)
	if featureKind != "" {
		summary += fmt.Sprintf(", raising a %s at (%d,%d)", featureKind, featureX, featureY)
	}
	mt.appendFile(mt.soulPath(), fmt.Sprintf("\n- %s — I canonized %s\n",
		clock.Format(mt.replicaTickSafe()), summary))
	return r, summary
}

// canonizeRefusal maps a guardian.region_named door rejection to in-fiction
// counsel (the planRefusal shape): the reducer's error strings are the
// source, translated so the model hears a repairable reason.
func canonizeRefusal(err error) string {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "overlaps the named region"):
		return "that ground is already named — the land takes one true name, never two"
	case strings.Contains(msg, "already named (cap"):
		return "I already hold as many named places as the world can bear"
	case strings.Contains(msg, "outside the world"), strings.Contains(msg, "lies outside"):
		return "that site lies beyond the world's edge, or beyond the region I would name"
	case strings.Contains(msg, "is not a valid build site"):
		return "something already stands there, or the ground will not bear it"
	case strings.Contains(msg, "unknown feature kind"):
		return "I cannot raise that there — ask for a feature within my power"
	case strings.Contains(msg, "name length"):
		return "give it a shorter name and I will make it real"
	case strings.Contains(msg, "radius"):
		return "name a smaller or larger reach — that span is not one I can hold"
	case strings.Contains(msg, "need") && strings.Contains(msg, "charge"):
		return "I do not hold charge enough for an act so large"
	default:
		return "the world would not let me (" + msg + ")"
	}
}

// handleCanonizeRegion wraps landCanonizeRegion: door accept → landed (a
// RegionReport on the shared result); refusal → rejected_gate the model may
// repair within the round cap (the handlePlaceDesignation dispatch shape).
func (mt *Guardian) handleCanonizeRegion(d *turnDispatch) toolloop.Handler {
	return func(_ context.Context, call llm.ToolCall) toolloop.Outcome {
		x, _ := argInt(call.Args, "x")
		y, _ := argInt(call.Args, "y")
		radius, _ := argInt(call.Args, "radius")
		fx, hasFX := argInt(call.Args, "feature_x")
		fy, hasFY := argInt(call.Args, "feature_y")
		r, summaryOrWhy := mt.landCanonizeRegion(x, y, radius, argString(call.Args, "name"),
			argString(call.Args, "feature_kind"), fx, fy, hasFX && hasFY, d.tick, d.charges, d.grant)
		if r == nil {
			return toolloop.Outcome{Verdict: toolloop.VerdictRejectedGate, ResultForModel: refusal(summaryOrWhy)}
		}
		d.result.Region = &RegionReport{ID: r.ID, Summary: summaryOrWhy}
		return toolloop.Outcome{Verdict: toolloop.VerdictLanded,
			ResultForModel: "it is made real (" + r.ID + ") — " + summaryOrWhy}
	}
}

// handleBriefMyths serves the read-only brief_myths tool (spec 101 D5): the
// mirrored, pre-derived candidate list (mt.myths, refreshed in mirrorState
// alongside every other turn-worker mirror) rendered into a fact sheet —
// never the live replica, the survey_site/handleSurvey discipline. Always
// read_ok: an empty corpus returns an honest "no candidates yet", never a
// hard error.
func (mt *Guardian) handleBriefMyths(d *turnDispatch) toolloop.Handler {
	return func(_ context.Context, call llm.ToolCall) toolloop.Outcome {
		limit, ok := argInt(call.Args, "limit")
		if !ok || limit <= 0 {
			limit = mythBriefingTopN
		}
		mt.stateMu.Lock()
		myths := append([]sim.PlaceMythBriefing(nil), mt.myths...)
		mt.stateMu.Unlock()
		if limit < len(myths) {
			myths = myths[:limit]
		}
		return toolloop.Outcome{Verdict: toolloop.VerdictReadOK, ResultForModel: renderMythBriefing(myths)}
	}
}

// renderMythBriefing composes the deterministic fact sheet brief_myths
// returns: one line per candidate, ranked (the myths slice arrives
// pre-ranked from sim.DominantPlaceMyths), or an honest empty notice.
func renderMythBriefing(myths []sim.PlaceMythBriefing) string {
	if len(myths) == 0 {
		return "no candidate place-myths are held strongly enough yet"
	}
	var b strings.Builder
	b.WriteString("Candidate place-myths, by how many villagers hold them:\n")
	for _, m := range myths {
		fmt.Fprintf(&b, "- near (%d,%d): %q — %d holder(s), %d%% confidence\n",
			m.X, m.Y, m.Statement, m.Holders, m.Confidence)
	}
	return b.String()
}

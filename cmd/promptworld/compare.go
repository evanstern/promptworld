package main

// `promptworld compare` (spec 076 US2/US3): the duel scoreboard — one
// honest rubric card per world through the ONE spec-072 resolver
// (tui.ResolveRubricFacts; a second precedence switch anywhere is a spec
// violation), then the drill-down: where the two recorded stories diverge,
// and their chronicle entries interleaved. Reads are OFFLINE
// (worlds.OfflineState — snapshot + fold), which works on stopped worlds
// and, under WAL, on running ones (said honestly in the header). Every
// authored line stays in the no-blame register: facts about the village,
// never judgments about the player (FR-020).

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/daemon"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/tui"
	"github.com/evanstern/promptworld/internal/world"
	"github.com/evanstern/promptworld/internal/worlds"
)

// duelCardWidth is the rendered scorecard width — compare is a plain-print
// command with no terminal negotiation, so the width is fixed.
const duelCardWidth = 72

// duelReport is compare's output model (data-model §6) — a model rather
// than print statements because it is also the input contract for the
// phase-2 HTML retelling (the documented follow-on).
type duelReport struct {
	A, B       duelSide
	Lineage    *duelLineage // non-nil when one side's lineage names the other
	Since      int64        // comparison window start (lineage fork tick | --since | 0)
	Divergence *divergence  // nil = identical since the window (the honest line)
	Entries    []duelEntry  // interleaved chronicle entries, tick order
}

// duelLineage names the fork relation the window derives from.
type duelLineage struct {
	Child, Parent string
	ForkTick      int64
}

// duelSide is one world's half of the duel.
type duelSide struct {
	Name     string
	Dir      string
	State    *sim.State              // offline-reconstructed (worlds.OfflineState)
	Events   []store.Event           // the full log, loaded once (divergence + interleave + lineage fallback)
	Exercise *sim.ExerciseDefinition // nil = ambient (no scorecard — honest note)
	Pass     *sim.CurriculumPass     // recorded instrument, if retained
	Facts    []tui.ReportCardFact    // via tui.ResolveRubricFacts
	Mode     tui.ReportCardMode
	Outcome  string // plain language, never the raw enum (FR-019)
	Running  bool   // daemon alive — the as-of-last-commit header note

	lineage   *world.LineageConfig // this side's fork provenance (manifest mirror, event fallback)
	createdAt string               // this side's manifest created_at (matches the other's lineage.parent_created_at)
}

// divergence is the first post-window story-event position at which the two
// recorded streams differ.
type divergence struct {
	Tick int64  // first differing story event's tick (the shared coordinate)
	Seq  int64  // side A's seq at the divergence when it has an event there, else side B's
	A, B string // one-line description per side ("<type>" or "— (no event)")
}

// duelEntry is one chronicle line in the interleave, labeled per world.
type duelEntry struct {
	World    string
	Day      int64
	FromTick int64
	Text     string
}

func cmdCompare(args []string) error {
	fs := flag.NewFlagSet("compare", flag.ContinueOnError)
	since := fs.Int64("since", -1, "compare from this tick (default: the fork tick when lineage links the two worlds, else 0)")

	// Two positionals with flags anywhere (the cmdFork pattern).
	var pos []string
	rest := args
	for len(rest) > 0 && !strings.HasPrefix(rest[0], "-") {
		pos = append(pos, rest[0])
		rest = rest[1:]
	}
	if err := fs.Parse(rest); err != nil {
		return err
	}
	pos = append(pos, fs.Args()...)
	if len(pos) != 2 {
		return fmt.Errorf("usage: promptworld compare <a> <b> [--since TICK]")
	}

	report, err := buildDuelReport(pos[0], pos[1], *since)
	if err != nil {
		return err
	}
	fmt.Print(renderDuelReport(report))
	return nil
}

// buildDuelReport assembles the whole duel model: both sides offline, the
// lineage-derived window, the divergence scan, and the chronicle
// interleave. sinceFlag < 0 means "not passed" (0 is a legal explicit
// window).
func buildDuelReport(argA, argB string, sinceFlag int64) (*duelReport, error) {
	a, err := buildDuelSide(argA)
	if err != nil {
		return nil, err
	}
	b, err := buildDuelSide(argB)
	if err != nil {
		return nil, err
	}

	r := &duelReport{A: *a, B: *b}

	// Window (FR-016): the fork tick when either side's lineage names the
	// other as parent — manifest block first, world.forked event fallback —
	// else 0; --since overrides.
	if lin := lineageBetween(a, b); lin != nil {
		r.Lineage = lin
		r.Since = lin.ForkTick
	}
	if sinceFlag >= 0 {
		r.Since = sinceFlag
	}

	r.Divergence = findDivergence(a, b, r.Since)
	r.Entries = interleaveChronicles(a, b, r.Since)
	return r, nil
}

// buildDuelSide resolves one argument into a fully-loaded side: offline
// state, the full event log, and — for a scenario world — the resolver's
// facts. Ambient worlds (no scenario block) keep Exercise nil: compare says
// honestly that no rubric exists and never invents a scorecard (US2-4).
func buildDuelSide(arg string) (*duelSide, error) {
	dir, err := resolveWorld(arg)
	if err != nil {
		return nil, err
	}
	w, err := world.Open(dir)
	if err != nil {
		return nil, err
	}
	state, _, err := worlds.OfflineState(w)
	if err != nil {
		return nil, err
	}
	st, err := store.Open(w.DBPath())
	if err != nil {
		return nil, err
	}
	events, err := st.EventsSince(0, 0)
	st.Close()
	if err != nil {
		return nil, err
	}

	side := &duelSide{Name: w.Manifest.Name, Dir: dir, State: state, Events: events}
	side.Running, _ = daemon.IsRunning(dir)

	if w.Manifest.Scenario != nil {
		if def, ok := sim.ExerciseByID(w.Manifest.Scenario.Exercise); ok {
			side.Exercise = &def
			side.Pass = tui.RecordedPassFor(state, def.ID)
			side.Facts, side.Mode = tui.ResolveRubricFacts(state, def, side.Pass)
			side.Outcome = plainOutcome(sim.ExerciseOutcome(state, def.ID))
		}
	}
	// The manifest lineage mirror may be absent on a hand-carried world; the
	// authoritative event is the fallback (FR-016). Newest wins: a fork of a
	// fork names its immediate parent last.
	if w.Manifest.Lineage == nil {
		for _, e := range events {
			if e.Type != "world.forked" {
				continue
			}
			var p sim.WorldForkedPayload
			if json.Unmarshal(e.Payload, &p) == nil {
				w.Manifest.Lineage = &world.LineageConfig{
					Parent: p.ParentName, ParentCreatedAt: p.ParentCreatedAt, ForkTick: p.ForkTick,
				}
			}
		}
	}
	side.lineage = w.Manifest.Lineage
	side.createdAt = w.Manifest.CreatedAt
	return side, nil
}

// plainOutcome maps sim.ExerciseOutcome's vocabulary to plain language
// (FR-019, data-model §7 — the glossary discipline): the raw enum tokens
// never print in a grade, and the loss reads in the postmortem register.
func plainOutcome(outcome string) string {
	switch outcome {
	case sim.OutcomePassed:
		return "passed"
	case sim.OutcomeFailed:
		return "did not make it through"
	case sim.OutcomeInProgress:
		return "still running"
	}
	return "unknown"
}

// lineageBetween reports the fork relation when one side's lineage names
// the other as parent — matched on name, plus parent_created_at when the
// lineage carries it (disambiguating renamed/recreated parents).
func lineageBetween(a, b *duelSide) *duelLineage {
	if match(b.lineage, a) {
		return &duelLineage{Child: b.Name, Parent: a.Name, ForkTick: b.lineage.ForkTick}
	}
	if match(a.lineage, b) {
		return &duelLineage{Child: a.Name, Parent: b.Name, ForkTick: a.lineage.ForkTick}
	}
	return nil
}

func match(lin *world.LineageConfig, parent *duelSide) bool {
	if lin == nil || lin.Parent != parent.Name {
		return false
	}
	return lin.ParentCreatedAt == "" || lin.ParentCreatedAt == parent.createdAt
}

// storyMachineryPrefixes are the event classes divergence NEVER keys on
// (FR-017, research R7): wall-dependent daemon/clock bookkeeping (the
// determinism e2e's exclusion) plus cognition/LLM telemetry, whose payloads
// differ for wall reasons even when the villagers' story is identical.
var storyMachineryPrefixes = []string{"daemon.", "clock.", "cog.", "llm."}

// isStoryEvent reports whether an event participates in the divergence
// scan. chronicle.entry is rendered in the interleave but never triggers
// divergence (narrated wording differs between runs of the SAME story);
// world.forked is the fork ceremony's own lineage marker — present only in
// the fork's log by construction, so counting it would make every duel
// "diverge" at the fork tick regardless of the prompt.
func isStoryEvent(e store.Event) bool {
	for _, p := range storyMachineryPrefixes {
		if strings.HasPrefix(e.Type, p) {
			return false
		}
	}
	return e.Type != "chronicle.entry" && e.Type != "world.forked"
}

// findDivergence compares the two logs' post-window story-event streams
// over (tick, type, payload) — never wall_time or seq — and returns the
// first differing position, or nil when the streams are identical (US3
// scenario 4: zero divergence is a truthful, teachable outcome). The scan
// runs through the two runs' COMMON tick horizon: events past the shorter
// run's clock are ticks the other world simply hasn't lived yet, not story
// divergence (the determinism e2e's trim discipline).
func findDivergence(a, b *duelSide, since int64) *divergence {
	horizon := a.State.Tick
	if b.State.Tick < horizon {
		horizon = b.State.Tick
	}
	filter := func(events []store.Event) []store.Event {
		var out []store.Event
		for _, e := range events {
			if e.Tick >= since && e.Tick <= horizon && isStoryEvent(e) {
				out = append(out, e)
			}
		}
		return out
	}
	sa, sb := filter(a.Events), filter(b.Events)
	n := len(sa)
	if len(sb) > n {
		n = len(sb)
	}
	for i := 0; i < n; i++ {
		switch {
		case i >= len(sa):
			e := sb[i]
			return &divergence{Tick: e.Tick, Seq: e.Seq, A: "— (no event)", B: e.Type}
		case i >= len(sb):
			e := sa[i]
			return &divergence{Tick: e.Tick, Seq: e.Seq, A: e.Type, B: "— (no event)"}
		default:
			ea, eb := sa[i], sb[i]
			if ea.Tick != eb.Tick || ea.Type != eb.Type || !bytes.Equal(ea.Payload, eb.Payload) {
				tick := ea.Tick
				if eb.Tick < tick {
					tick = eb.Tick
				}
				return &divergence{Tick: tick, Seq: ea.Seq, A: ea.Type, B: eb.Type}
			}
		}
	}
	return nil
}

// interleaveChronicles merges both sides' chronicle.entry events with
// from_tick >= since, stable by FromTick (ties: A before B), labeled per
// world (US3 scenario 3).
func interleaveChronicles(a, b *duelSide, since int64) []duelEntry {
	side := func(s *duelSide) []duelEntry {
		var out []duelEntry
		for _, e := range s.Events {
			if e.Type != "chronicle.entry" {
				continue
			}
			var p sim.ChronicleEntryPayload
			if err := json.Unmarshal(e.Payload, &p); err != nil {
				continue
			}
			if p.FromTick >= since {
				out = append(out, duelEntry{World: s.Name, Day: p.Day, FromTick: p.FromTick, Text: p.Text})
			}
		}
		return out
	}
	ea, eb := side(a), side(b)
	out := make([]duelEntry, 0, len(ea)+len(eb))
	i, j := 0, 0
	for i < len(ea) || j < len(eb) {
		switch {
		case j >= len(eb), i < len(ea) && ea[i].FromTick <= eb[j].FromTick: // ties: A before B, stable
			out = append(out, ea[i])
			i++
		default:
			out = append(out, eb[j])
			j++
		}
	}
	return out
}

// gameTimeAt renders a tick as "day D, HH:MM (tick T)".
func gameTimeAt(tick int64) string {
	day, h, m, _ := clock.GameTime(tick)
	return fmt.Sprintf("day %d, %02d:%02d (tick %d)", day, h, m, tick)
}

// renderDuelReport prints the whole duel: header (lineage, liveness,
// window), one scorecard per side through the shared renderer, the
// divergence line, and the interleaved chronicles with the divergence
// marker in timeline position.
func renderDuelReport(r *duelReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "duel: %s vs %s\n", r.A.Name, r.B.Name)
	if r.Lineage != nil {
		fmt.Fprintf(&b, "%s forked from %s at %s\n", r.Lineage.Child, r.Lineage.Parent, gameTimeAt(r.Lineage.ForkTick))
	}
	for _, s := range []*duelSide{&r.A, &r.B} {
		if s.Running {
			fmt.Fprintf(&b, "%s is running — read as of its last committed batch\n", s.Name)
		}
	}
	fmt.Fprintf(&b, "comparing from %s\n", gameTimeAt(r.Since))
	if r.A.Exercise != nil && r.B.Exercise != nil && r.A.Exercise.ID != r.B.Exercise.ID {
		b.WriteString("note: the two worlds run different exercises — each card grades its own, not head-to-head\n")
	}

	for _, s := range []*duelSide{&r.A, &r.B} {
		b.WriteString("\n")
		if s.Exercise == nil {
			fmt.Fprintf(&b, "%s — an ambient world: no rubric exists for it, so there is no scorecard\n", s.Name)
			continue
		}
		fmt.Fprintf(&b, "%s — %s\n", s.Name, s.Outcome)
		b.WriteString(tui.RenderReportCard(s.Exercise.ID, s.Facts, s.Mode, duelCardWidth))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	if r.Divergence == nil {
		if r.Lineage != nil {
			b.WriteString("the two runs are identical since the fork — the change made no observable difference yet\n")
		} else {
			fmt.Fprintf(&b, "the two runs' recorded stories are identical since %s\n", gameTimeAt(r.Since))
		}
	} else {
		fmt.Fprintf(&b, "the stories diverge at %s\n", gameTimeAt(r.Divergence.Tick))
		fmt.Fprintf(&b, "  %s: %s · %s: %s\n", r.A.Name, r.Divergence.A, r.B.Name, r.Divergence.B)
	}

	if len(r.Entries) > 0 {
		b.WriteString("\nchronicles, interleaved:\n")
		markerDone := r.Divergence == nil
		for _, e := range r.Entries {
			if !markerDone && e.FromTick >= r.Divergence.Tick {
				fmt.Fprintf(&b, "  ── the stories diverge here (%s) ──\n", gameTimeAt(r.Divergence.Tick))
				markerDone = true
			}
			fmt.Fprintf(&b, "  [%s] day %d — %s\n", e.World, e.Day, e.Text)
		}
		if !markerDone {
			fmt.Fprintf(&b, "  ── the stories diverge here (%s) ──\n", gameTimeAt(r.Divergence.Tick))
		}
	}
	return b.String()
}

package scribe

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
)

// The morgue (spec 044 US2): the run's accumulating legacy document —
// one factual epitaph per death (facts + the guardian-policy evidence in force
// at that moment) closed by a run-end summary, with the narrator's recorded
// epilogues blockquoted beneath their sections' facts.
//
// The render is a whole-file regeneration and a PURE FOLD over the recorded
// history (contracts/morgue-document.md): a fresh reducer state replays the
// full event log from genesis, and each death's fields are captured from the
// folding state AT that event — never from the live replica, whose relations,
// debts, and standing orders keep moving after a death and would change a
// prior section's bytes on a later render (invariant 5: prior sections'
// factual bytes never change). Replaying the same history therefore
// reproduces the factual content byte-identically (SC-004), and a deleted or
// hand-edited morgue.md is healed by the next render (FR-011). Epilogues are
// recorded events collected in the same pass — included in the render,
// excluded from the byte-identity requirement.

const (
	// morgueNotableSalience is the lifetime-memory scan threshold (research
	// R7): nightly consolidation deletes consolidated memories from state
	// (at-death retained memories under-report a long life), so the render
	// also scans agent.memory_added for the high-salience band — 7 and up is
	// the tier of gru encounters, village-visible builds, thefts, near-death,
	// and witnessed deaths (internal/sim/memory.go).
	morgueNotableSalience = 7
	// morgueMemoryCap bounds "what they carried in memory" per epitaph —
	// salience-ranked, so the cap trims the least notable tail.
	morgueMemoryCap = 12
	// morgueDeedCap bounds an epitaph's deed list (most recent kept, count of
	// elided older deeds stated); morgueRunEventCap does the same for the
	// run summary's notable events.
	morgueDeedCap     = 20
	morgueRunEventCap = 60
)

// charterObs is one point on the charter-revision timeline
// (metatron.charter_observed).
type charterObs struct {
	tick int64
	fp   string
	def  bool
}

// morgueMem is one memory line candidate: a retained at-death memory or a
// lifetime scan hit. Dedup key is (tick, text).
type morgueMem struct {
	tick int64
	sal  int
	text string
}

// morgueDeed is one curated notable-event line and the villagers it belongs to.
type morgueDeed struct {
	tick int64
	who  []int
	line string
}

// morgueEpitaph is one death's capture: every field read from the folding
// state at the death event.
type morgueEpitaph struct {
	agent     int
	tick      int64
	cause     string
	name      string
	relations []string
	owed      []string // open debts the villager owed
	owedTo    []string // open debts owed to the villager
	memories  []morgueMem
	orders    []string
	charter   *charterObs
}

// morgueEpilogueRec is one recorded morgue.epilogue, in event order.
type morgueEpilogueRec struct {
	agent int
	text  string
}

// renderMorgue writes morgue.md as a pure fold over the full event history
// (see the package comment above). A scribe constructed without an
// EventSource renders no morgue.
func (s *Scribe) renderMorgue() {
	if s.src == nil {
		return
	}
	st := sim.NewState(s.seed, s.m)
	worldName := filepath.Base(s.worldDir)
	var (
		epitaphs  []morgueEpitaph
		timeline  []charterObs
		epilogues []morgueEpilogueRec
		deeds     []morgueDeed
		lifetime  = map[int][]morgueMem{}
		runEnd    *sim.RunEnd
	)
	s.src.ReplayEvents(0, func(e store.Event) error {
		// Pre-apply captures — deed lines read the state as the event found it
		// (a broken promise's debt is still open; a violated norm still names
		// its text).
		switch e.Type {
		case "world.created":
			var p sim.WorldCreatedPayload
			if json.Unmarshal(e.Payload, &p) == nil && p.Name != "" {
				worldName = p.Name
			}
		case "metatron.charter_observed":
			var p sim.CharterObservedPayload
			if json.Unmarshal(e.Payload, &p) == nil {
				timeline = append(timeline, charterObs{tick: e.Tick, fp: p.Fingerprint, def: p.Default})
			}
		case "morgue.epilogue":
			var p sim.MorgueEpiloguePayload
			if json.Unmarshal(e.Payload, &p) == nil {
				epilogues = append(epilogues, morgueEpilogueRec{agent: p.Agent, text: p.Text})
			}
		case "agent.memory_added":
			var p sim.MemoryAddedPayload
			if json.Unmarshal(e.Payload, &p) == nil && p.Salience >= morgueNotableSalience {
				lifetime[p.Agent.ID] = append(lifetime[p.Agent.ID], morgueMem{tick: e.Tick, sal: p.Salience, text: p.Text})
			}
		default:
			if line, who := morgueDeedNote(st, e); line != "" {
				deeds = append(deeds, morgueDeed{tick: e.Tick, who: who, line: line})
			}
		}
		st.Apply(e)
		if e.Tick > st.Tick {
			st.Tick = e.Tick
		}
		// Post-apply captures — the epitaph reads the state including this
		// death (the ledger entry, the spilled inventory).
		switch e.Type {
		case "agent.died":
			var p sim.DiedPayload
			if json.Unmarshal(e.Payload, &p) == nil {
				var latest *charterObs
				if len(timeline) > 0 {
					c := timeline[len(timeline)-1]
					latest = &c
				}
				ep := captureEpitaph(st, p.Agent.ID, e.Tick, p.Cause, latest)
				ep.memories = mergeMemories(ep.memories, lifetime[p.Agent.ID])
				epitaphs = append(epitaphs, ep)
			}
		case "run.ended":
			if st.RunEnd != nil {
				re := *st.RunEnd
				runEnd = &re
			}
		}
		return nil
	})

	var b strings.Builder
	fmt.Fprintf(&b, "# Morgue — %s\n\n", worldName)
	b.WriteString("_One run, one directory. This document is regenerated from the world's history._\n")
	if len(epitaphs) == 0 {
		b.WriteString("\n*No one has died. The village lives.*\n")
	}
	for _, ep := range epitaphs {
		writeEpitaph(&b, ep, deedsFor(deeds, ep.agent, ep.tick))
		writeEpilogues(&b, epilogues, ep.agent)
	}
	if runEnd != nil {
		writeRunSummary(&b, st, runEnd, deeds, s.scenarioExercise)
		writeEpilogues(&b, epilogues, -1)
	}
	os.WriteFile(filepath.Join(s.worldDir, "morgue.md"), []byte(b.String()), 0o644)
}

// captureEpitaph reads one death's fields from the folding state at the death
// event — relations, open debts, retained memories, and the standing orders
// active at that moment, plus the most recent charter observation.
func captureEpitaph(st *sim.State, agent int, tick int64, cause string, charter *charterObs) morgueEpitaph {
	ep := morgueEpitaph{agent: agent, tick: tick, cause: cause, name: "someone", charter: charter}
	if agent < 0 || agent >= len(st.Agents) {
		return ep
	}
	a := st.Agents[agent]
	ep.name = a.Name
	// Relations in canonical state order (contract invariant 3).
	for _, r := range st.Relations {
		if r.From != agent || (r.Trust == 0 && r.Affection == 0) {
			continue
		}
		ep.relations = append(ep.relations, fmt.Sprintf("%s: trust %+d, affection %+d",
			st.Agents[r.To].Name, r.Trust, r.Affection))
	}
	// Open debts at death, canonical state order — evidence that outlives them.
	for _, d := range st.Debts {
		if d.Status != "open" {
			continue
		}
		switch agent {
		case d.Debtor:
			ep.owed = append(ep.owed, fmt.Sprintf("one %s to %s", d.Kind, st.Agents[d.Creditor].Name))
		case d.Creditor:
			ep.owedTo = append(ep.owedTo, fmt.Sprintf("one %s from %s", d.Kind, st.Agents[d.Debtor].Name))
		}
	}
	// Retained memories at death; the lifetime scan is merged at render time.
	for _, m := range a.Memories {
		ep.memories = append(ep.memories, morgueMem{tick: m.Tick, sal: m.Salience, text: m.Text})
	}
	// Standing orders active at this moment (spec 044 FR-008): condition,
	// action, and watch subjects — the instruction half of the evidence.
	for _, o := range st.GuardianOrders {
		if o.Status != "active" {
			continue
		}
		watch := strings.Join(o.EventTypes, ", ")
		if o.Agent >= 0 && o.Agent < len(st.Agents) {
			watch += "; villager " + st.Agents[o.Agent].Name
		}
		if len(o.Keywords) > 0 {
			watch += "; keywords: " + strings.Join(o.Keywords, ", ")
		}
		ep.orders = append(ep.orders, fmt.Sprintf("%q → %q — watching %s", o.Condition, o.Action, watch))
	}
	return ep
}

// mergeMemories unions the at-death retained memories with the lifetime scan
// hits, deduped by (tick, text), ordered (salience desc, tick asc, text asc)
// per contract invariant 3, capped at morgueMemoryCap.
func mergeMemories(retained, scanned []morgueMem) []morgueMem {
	type key struct {
		tick int64
		text string
	}
	seen := map[key]bool{}
	var out []morgueMem
	for _, m := range append(append([]morgueMem(nil), retained...), scanned...) {
		k := key{m.tick, m.text}
		if seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, m)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].sal != out[j].sal {
			return out[i].sal > out[j].sal
		}
		if out[i].tick != out[j].tick {
			return out[i].tick < out[j].tick
		}
		return out[i].text < out[j].text
	})
	if len(out) > morgueMemoryCap {
		out = out[:morgueMemoryCap]
	}
	return out
}

// deedsFor filters the curated deed lines to one villager's, up to their
// death tick, oldest → newest (the collection order).
func deedsFor(deeds []morgueDeed, agent int, until int64) []morgueDeed {
	var out []morgueDeed
	for _, d := range deeds {
		if d.tick > until {
			continue
		}
		for _, w := range d.who {
			if w == agent {
				out = append(out, d)
				break
			}
		}
	}
	return out
}

// writeEpitaph renders one death's section per contracts/morgue-document.md.
// Stated as evidence throughout — what was recorded, what was instructed —
// with no judgment vocabulary anywhere in the static strings (invariant 2).
func writeEpitaph(b *strings.Builder, ep morgueEpitaph, deeds []morgueDeed) {
	day, _, _, _ := clock.GameTime(ep.tick)
	fmt.Fprintf(b, "\n## %s — died day %d (%s)\n\n", ep.name, day, ep.cause)
	fmt.Fprintf(b, "- **Days survived**: %d\n", day)
	fmt.Fprintf(b, "- **Cause**: %s\n", ep.cause)

	b.WriteString("- **What they will be remembered for**:")
	if len(deeds) == 0 {
		b.WriteString(" none recorded\n")
	} else {
		b.WriteString("\n")
		if drop := len(deeds) - morgueDeedCap; drop > 0 {
			fmt.Fprintf(b, "  - _(%d earlier deeds not shown)_\n", drop)
			deeds = deeds[drop:]
		}
		for _, d := range deeds {
			fmt.Fprintf(b, "  - **%s** — %s\n", clock.Format(d.tick), d.line)
		}
	}

	b.WriteString("- **What they carried in memory**:")
	if len(ep.memories) == 0 {
		b.WriteString(" none recorded\n")
	} else {
		b.WriteString("\n")
		for _, m := range ep.memories {
			fmt.Fprintf(b, "  - **%s** (%d★) %s\n", clock.Format(m.tick), m.sal, m.text)
		}
	}

	b.WriteString("- **Who mattered to them**:")
	if len(ep.relations) == 0 {
		b.WriteString(" no standing bonds\n")
	} else {
		b.WriteString("\n")
		for _, r := range ep.relations {
			b.WriteString("  - " + r + "\n")
		}
	}

	owed, owedTo := "none", "none"
	if len(ep.owed) > 0 {
		owed = strings.Join(ep.owed, ", ")
	}
	if len(ep.owedTo) > 0 {
		owedTo = strings.Join(ep.owedTo, ", ")
	}
	fmt.Fprintf(b, "- **Debts**: owed — %s · owed to them — %s\n", owed, owedTo)

	// Freshly rendered morgue prose de-themes to the default guardian
	// vocabulary (spec 052 T012); already-written morgue files are history.
	b.WriteString("- **The guardian's watch at that moment**: ")
	if ep.charter == nil {
		b.WriteString("no charter observation recorded before this death")
	} else {
		prov := "player-authored"
		if ep.charter.def {
			prov = "default"
		}
		cday, _, _, _ := clock.GameTime(ep.charter.tick)
		fmt.Fprintf(b, "charter revision `%s` (%s), in force since day %d", ep.charter.fp, prov, cday)
	}
	if len(ep.orders) == 0 {
		b.WriteString("; standing orders active: none.\n")
	} else {
		b.WriteString("; standing orders active:\n")
		for _, o := range ep.orders {
			b.WriteString("  - " + o + "\n")
		}
	}
	b.WriteString("  _Stated as evidence; the reader draws the lesson._\n")
}

// writeRunSummary renders the closing village-level section (FR-009): run
// length, the day-stamped population decline, every death with cause, and the
// run's notable events (the same curated vocabulary the epitaphs use). On a
// scenario world (spec 054 US5, FR-010) it also names the exercise and its
// outcome — the no-blame evidence register: failure is a story, not a scold,
// so the line states what the rubric asked and what the run gave, and stops.
func writeRunSummary(b *strings.Builder, st *sim.State, re *sim.RunEnd, deeds []morgueDeed, scenarioExercise string) {
	day, _, _, _ := clock.GameTime(re.Tick)
	fmt.Fprintf(b, "\n## The run — ended day %d\n\n", day)
	fmt.Fprintf(b, "- **Run length**: %d days\n", day)
	if scenarioExercise != "" {
		outcome := "failed — the run ended before its rubric was met"
		if sim.ExerciseOutcome(st, scenarioExercise) == sim.OutcomePassed {
			outcome = "passed — the run's end came after the pass was earned"
		}
		fmt.Fprintf(b, "- **The exercise**: %s — %s. _Stated as evidence; the reader draws the lesson._\n",
			scenarioExercise, outcome)
	}

	pop := len(st.Agents)
	curve := fmt.Sprintf("%d", pop)
	for _, d := range re.Deaths {
		pop--
		dd, _, _, _ := clock.GameTime(d.Tick)
		curve += fmt.Sprintf(" → %d (day %d)", pop, dd)
	}
	fmt.Fprintf(b, "- **Population**: %s\n", curve)

	b.WriteString("- **The deaths**:\n")
	for _, d := range re.Deaths {
		name := "someone"
		if d.Agent >= 0 && d.Agent < len(st.Agents) {
			name = st.Agents[d.Agent].Name
		}
		dd, _, _, _ := clock.GameTime(d.Tick)
		fmt.Fprintf(b, "  - **day %d** — %s (%s)\n", dd, name, d.Cause)
	}

	b.WriteString("- **Notable events of the run**:")
	if len(deeds) == 0 {
		b.WriteString(" none recorded\n")
	} else {
		b.WriteString("\n")
		if drop := len(deeds) - morgueRunEventCap; drop > 0 {
			fmt.Fprintf(b, "  - _(%d earlier events not shown)_\n", drop)
			deeds = deeds[drop:]
		}
		for _, d := range deeds {
			fmt.Fprintf(b, "  - **%s** — %s\n", clock.Format(d.tick), d.line)
		}
	}
}

// writeEpilogues blockquotes the recorded epilogues for one section's subject
// (a villager index, or -1 for the run end), in event order, after that
// section's facts (invariant 1: facts before prose; removing every epilogue
// leaves a complete document).
func writeEpilogues(b *strings.Builder, eps []morgueEpilogueRec, agent int) {
	for _, e := range eps {
		if e.agent != agent {
			continue
		}
		fmt.Fprintf(b, "\n> _Epilogue_ — %s\n", strings.ReplaceAll(e.text, "\n", " "))
	}
}

// morgueDeedNote is the morgue's copy of the chronicle's curated
// notable-event vocabulary (research R7 — narrate.go's chronicleNote switch,
// restricted to the deed-shaped subset: builds, gifts and promises, thefts,
// governance arcs, gru encounters), rendered against the state AS THE EVENT
// FOUND IT and attributed to the villagers whose story it is. Returns ("",
// nil) for every other event type.
func morgueDeedNote(st *sim.State, e store.Event) (string, []int) {
	name := func(i int) string {
		if i >= 0 && i < len(st.Agents) {
			return st.Agents[i].Name
		}
		return "someone"
	}
	switch e.Type {
	case "agent.built":
		var p sim.BuiltPayload
		if json.Unmarshal(e.Payload, &p) == nil {
			return fmt.Sprintf("%s built a %s.", name(p.Agent.ID), p.Kind), []int{p.Agent.ID}
		}
	case "social.gave":
		var p sim.GavePayload
		if json.Unmarshal(e.Payload, &p) == nil {
			return fmt.Sprintf("%s gave %s %s.", name(p.From), name(p.To), p.Kind), []int{p.From, p.To}
		}
	case "social.promise_broken":
		var p sim.PromiseBrokenPayload
		if json.Unmarshal(e.Payload, &p) == nil {
			for _, d := range st.Debts {
				if d.ID == p.ID {
					return fmt.Sprintf("%s broke a promise to %s (an owed %s).",
						name(d.Debtor), name(d.Creditor), d.Kind), []int{d.Debtor, d.Creditor}
				}
			}
		}
	case "social.chest_taken":
		var p sim.ChestTakenPayload
		if json.Unmarshal(e.Payload, &p) == nil {
			return fmt.Sprintf("%s took from %s's chest without asking.",
				name(p.Taker), name(p.Owner)), []int{p.Taker, p.Owner}
		}
	case "gru.sighted":
		var p sim.GruSightedPayload
		if json.Unmarshal(e.Payload, &p) == nil {
			return fmt.Sprintf("%s sighted the gru.", name(p.Agent)), []int{p.Agent}
		}
	case "gru.attacked":
		var p sim.GruAttackedPayload
		if json.Unmarshal(e.Payload, &p) == nil {
			return fmt.Sprintf("The gru attacked %s.", name(p.Agent)), []int{p.Agent}
		}
	case "meeting.proposal_tabled":
		var p sim.ProposalPayload
		if json.Unmarshal(e.Payload, &p) == nil {
			return fmt.Sprintf("%s put a proposal to the assembly: %q.", name(p.Proposer), p.Text), []int{p.Proposer}
		}
	case "meeting.proposal_resolved":
		var p sim.ProposalResolvedPayload
		if json.Unmarshal(e.Payload, &p) == nil {
			tally := fmt.Sprintf("%d-%d", len(p.Yeas), len(p.Nays))
			switch {
			case p.Passed && p.Kind == sim.ProposeExile:
				return fmt.Sprintf("The village voted %s to exile %s.", tally, name(p.Target)),
					[]int{p.Proposer, p.Target}
			case p.Passed:
				return fmt.Sprintf("The village passed %s's proposal %s: %q.", name(p.Proposer), tally, p.Text),
					[]int{p.Proposer}
			default:
				return fmt.Sprintf("The village voted down %s's proposal %s.", name(p.Proposer), tally),
					[]int{p.Proposer}
			}
		}
	case "norm.violated":
		var p sim.NormViolatedPayload
		if json.Unmarshal(e.Payload, &p) == nil {
			line := fmt.Sprintf("%s was seen breaking the village's law.", name(p.Violator))
			if n := sim.NormByID(st, p.NormID); n != nil {
				line = fmt.Sprintf("%s was seen breaking the village's law: %q.", name(p.Violator), n.Text)
			}
			return line, []int{p.Violator}
		}
	}
	return "", nil
}

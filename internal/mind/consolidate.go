package mind

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/persona"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
)

// The nightly consolidation driver (TASK-9): when a villager sleeps, one
// cloud-tier call digests the day's episodic buffer into promotions, fades,
// a day-gist, belief revisions, and a rewritten self-narrative. The output
// passes the deterministic firewall validator, then lands as ONE atomic
// whitelisted batch — or a rejection marker lands, or (transport failure)
// nothing lands and the next sleep retries. The world never waits.

const (
	// consolidateCallTimeout bounds one cloud call; the night is hours long,
	// so this is generous rather than interactive.
	consolidateCallTimeout = 3 * time.Minute
	// maxBufferSent caps prompt size; overflow is truncated oldest-first for
	// the call only (state keeps everything).
	maxBufferSent = 60
)

// consolJob is the immutable snapshot a consolidation runs against —
// everything is copied at enqueue time so the ticking replica can't race it.
type consolJob struct {
	agent     int
	name      string
	personaMD string
	anchor    string
	drift     []string
	night     int64
	sleepTick int64
	upTo      int64 // buffer high-water mark (whole buffer, sent or not)
	buffer    []sim.Memory
	held      []sim.Belief
	social    string
	narrative string
	// Private dreams (spec 098): the FULL memory store snapshot the clustering
	// pass reads — this one agent's memories and nothing else (D1) — plus the
	// world seed and the resolved dream dials, all copied at enqueue time.
	mems  []sim.Memory
	seed  uint64
	dials sim.DreamTuning
}

// maxDreamGroupsSent caps the ambiguous clusters one consolidation prompt
// carries (D2: the band consults the existing slot, it must not bloat it).
// Un-sent groups simply retry a later night — the store keeps everything.
const maxDreamGroupsSent = 4

// maybeConsolidate is called from absorb on agent.slept. Guards are checked
// on the replica; due agents are snapshotted and queued for the single-
// flight worker.
func (md *Mind) maybeConsolidate(e store.Event) {
	if md.social == nil {
		return
	}
	var p sim.AgentPayload
	if json.Unmarshal(e.Payload, &p) != nil {
		return
	}
	if p.Agent.ID < 0 || p.Agent.ID >= sim.AgentCount {
		return
	}
	a := &md.replica.Agents[p.Agent.ID]
	if !a.ConsolidationDue(e.Tick) || md.consolInFlight[p.Agent.ID].Load() {
		return
	}
	night := sim.NightIndex(e.Tick)

	buffer := a.EpisodicBuffer()
	if len(buffer) == 0 {
		// Nothing to digest: close the night with a marker, spend no call.
		md.consolInFlight[p.Agent.ID].Store(true)
		md.landMarker(consolJob{agent: p.Agent.ID, name: a.Name, night: night, sleepTick: e.Tick},
			sim.ConsolidationSkippedEmpty, "", 0, 0)
		return
	}

	job := consolJob{
		agent:     p.Agent.ID,
		name:      a.Name,
		personaMD: md.personas[p.Agent.ID],
		anchor:    persona.Anchors[a.Name],
		drift:     persona.DriftMarkers[a.Name],
		night:     night,
		sleepTick: e.Tick,
		upTo:      buffer[len(buffer)-1].Tick,
		buffer:    append([]sim.Memory(nil), buffer...),
		held:      append([]sim.Belief(nil), a.Beliefs...),
		social:    socialContext(md.replica, p.Agent.ID),
		narrative: a.Narrative,
		mems:      append([]sim.Memory(nil), a.Memories...),
		seed:      md.replica.Seed,
		dials:     md.replica.DreamDials(),
	}
	if len(job.buffer) > maxBufferSent {
		job.buffer = job.buffer[len(job.buffer)-maxBufferSent:] // newest kept
	}
	// Router gate (FR-007): a night-scale budget passes at every watchable
	// speed today; the gate is doctrine-completeness, and a suppression here
	// (future faster speeds) skips the night — the next sleep retries.
	if v := md.routeVerdict("consolidation", llm.KindConsolidation); !v.Allow {
		md.emitSuppressed("consolidation", p.Agent.ID, e.Tick, v)
		return
	}
	md.consolInFlight[p.Agent.ID].Store(true)
	select {
	case md.consolQ <- job:
	default:
		// Queue full (should not happen with cap 8): drop the attempt; the
		// next sleep retries.
		md.consolInFlight[p.Agent.ID].Store(false)
	}
}

// consolidateWorker drains the night's queue one call at a time.
func (md *Mind) consolidateWorker() {
	for {
		select {
		case <-md.done:
			return
		case job := <-md.consolQ:
			md.runConsolidation(job)
		}
	}
}

func (md *Mind) runConsolidation(job consolJob) {
	defer md.consolInFlight[job.agent].Store(false)

	// Private dreams (spec 098): the geometry pass runs first, over this one
	// agent's snapshot alone (D1). Clear-cut habituation/merge outcomes land
	// NOW, as their own recorded batch, independent of the LLM night's fate —
	// geometry needed no model, so a deferred or rejected call must not undo
	// it. Only the ambiguous band rides the consolidation prompt (D2).
	dream := sim.PlanDream(job.mems, job.seed, job.night, job.agent, job.dials)
	if batch := sim.DreamEvents(job.agent, dream.Revisions, dream.Merges); len(batch) > 0 {
		if err := md.social.InjectSocial(batch); err != nil {
			log.Printf("mind: dream %s night %d geometry batch rejected: %v", job.name, job.night, err)
		} else {
			log.Printf("mind: dream %s night %d habituated %d, merged %d (geometry)",
				job.name, job.night, len(dream.Revisions), len(dream.Merges))
		}
	}
	groups := dream.Ambiguous
	if len(groups) > maxDreamGroupsSent {
		groups = groups[:maxDreamGroupsSent]
	}

	// Truncation-aware submit ladder (spec 105): the SAME prompt every attempt
	// — the dream geometry pass ran once above and its batch already landed; a
	// retry re-sends the identical snapshot-built request at a doubled budget,
	// never re-plans dreams and never re-snapshots the replica. Each consumed
	// retry is recorded as cog.outcome{retried} (FR-004).
	var out consolidationOutput
	res, err := md.submitWithTruncationRetry(consolidateCallTimeout, llm.Request{
		Kind:      llm.KindConsolidation,
		System:    consolidateSystemPrompt(job),
		Prompt:    consolidateUserPrompt(job, groups),
		MaxTokens: md.consolidationTokens, // llm.json max_tokens.consolidation (spec 025 US2), default 1024 — the ladder's start
	}, func(text string) error {
		var perr error
		out, perr = parseConsolidation(text)
		return perr
	}, func(retry int, from, to int64) {
		log.Printf("mind: consolidation %s night %d truncated at %d tokens; retry %d at %d",
			job.name, job.night, from, retry, to)
		md.emitTruncationRetry("consolidation", job.agent, job.sleepTick, retry, from, to)
	})
	if err != nil {
		// Transport/tier failure (any attempt): NO marker — the attempt never
		// happened as far as the ledger cares; the next sleep retries (FR-002).
		log.Printf("mind: consolidation %s night %d deferred: %v", job.name, job.night, err)
		return
	}
	resp := res.Resp
	if res.ParseErr != nil {
		// A ladder exhausted while still truncated is a budget failure, not a
		// garbage reply — the distinct reason keeps the durable record
		// actionable (FR-003). The buffer stays intact; the next sleep retries
		// from the ladder's start.
		reason := "unparseable"
		if res.Truncated {
			reason = sim.ConsolidationReasonTruncated
		}
		md.landMarker(job, sim.ConsolidationRejected, reason, res.CostUSD, res.Retries)
		return
	}
	// Models routinely invent an ID for a belief they mean as new (live
	// finding: 4/8 first-night rejections). ID bookkeeping is ours, not
	// theirs — coerce unknown IDs to "new" before judging the output.
	heldIDs := make(map[int]bool, len(job.held))
	for _, b := range job.held {
		heldIDs[b.ID] = true
	}
	for i := range out.Beliefs {
		if out.Beliefs[i].ID != 0 && !heldIDs[out.Beliefs[i].ID] {
			out.Beliefs[i].ID = 0
		}
	}
	// Over-long lists are enthusiasm, not corruption (live finding: 3/8
	// rejections were cap overruns) — keep the best-first prefix instead of
	// wasting the night. The validator's caps stay as hard guards behind us.
	if len(out.Promote) > maxPromotes {
		out.Promote = out.Promote[:maxPromotes]
	}
	if len(out.Fade) > maxFades {
		out.Fade = out.Fade[:maxFades]
	}
	if len(out.Beliefs) > maxBeliefEdits {
		out.Beliefs = out.Beliefs[:maxBeliefEdits]
	}
	// Evidence citations are pre-trimmed best-first per belief (spec 030): an
	// over-long list is enthusiasm, kept as its best-first prefix, never a
	// rejected night (contracts/consolidation-contract.md).
	for i := range out.Beliefs {
		if len(out.Beliefs[i].Evidence) > maxBeliefEvidence {
			out.Beliefs[i].Evidence = out.Beliefs[i].Evidence[:maxBeliefEvidence]
		}
	}
	// Routine-group verdicts (spec 098): mechanical slack, the unknown-belief-ID
	// discipline — an unparseable or unsent group label is dropped, a repeat is
	// deduplicated, and the night is never rejected over a dream verdict.
	routine := parseRoutineRefs(out.Routine, len(groups))
	if verr := validateConsolidation(out, job.agent, job.buffer, job.held, job.anchor, job.drift); verr != nil {
		snippet := resp.Text
		if len(snippet) > 180 {
			snippet = snippet[:180]
		}
		log.Printf("mind: consolidation %s night %d invalid output: %q", job.name, job.night, snippet)
		md.landMarker(job, sim.ConsolidationRejected, verr.Error(), res.CostUSD, res.Retries)
		return
	}

	// Provenance enforcement (spec 030, deterministic, post-validation): resolve
	// each belief's evidence and coerce "witnessed" claims that lack direct
	// perception. Never rejects — the coercion count rides the marker telemetry.
	coerced := enforceProvenance(out.Beliefs, job.buffer)

	// Accepted: build the whole night as one atomic batch.
	var batch []store.Event
	add := func(typ string, payload any) {
		b, _ := json.Marshal(payload)
		batch = append(batch, store.Event{Type: typ, Payload: b})
	}
	// Map ordinal refs back to (tick, hash) — the durable identity the
	// events carry — deduplicating repeats.
	seen := map[int]bool{}
	for _, r := range out.Promote {
		i := parseMemRef(r, len(job.buffer))
		if i < 0 || seen[i] {
			continue
		}
		seen[i] = true
		m := job.buffer[i]
		add("agent.memory_promoted", sim.MemoryPromotedPayload{
			Agent: sim.Ref(job.agent), MemTick: m.Tick, TextHash: sim.MemoryHash(m.Text), Boost: 3})
	}
	for _, r := range out.Fade {
		i := parseMemRef(r, len(job.buffer))
		if i < 0 || seen[i] {
			continue
		}
		seen[i] = true
		m := job.buffer[i]
		add("agent.memory_faded", sim.MemoryFadedPayload{
			Agent: sim.Ref(job.agent), MemTick: m.Tick, TextHash: sim.MemoryHash(m.Text)})
	}
	add("agent.memory_added", sim.MemoryAddedPayload{
		Agent: sim.Ref(job.agent), Text: out.Gist, Salience: sim.SalDayGist, Subject: sim.Ref(-1), Origin: sim.OriginDigest})
	for _, b := range out.Beliefs {
		add("agent.belief_revised", sim.BeliefRevisedPayload{
			Agent: sim.Ref(job.agent), BeliefID: b.ID, Statement: b.Statement,
			Confidence: b.Confidence, Provenance: b.Provenance,
			Source: sim.Ref(b.Source), Subject: sim.Ref(b.Subject),
			Evidence: b.resolved, Direct: b.direct})
	}
	add("agent.narrative_set", sim.NarrativeSetPayload{Agent: sim.Ref(job.agent), Text: out.Narrative})
	// Ambiguous-band verdicts (spec 098 D2/D3): a folded group lands its
	// PRECOMPUTED habituation/merge outcomes as recorded events in this same
	// atomic batch — the slot judged, geometry had already priced the fold, and
	// replay applies the records without re-deriving either. Kept groups land
	// nothing but the marker counts below: the keep decision is recorded, the
	// store is untouched.
	for _, gi := range routine {
		g := groups[gi]
		if len(g.Merge.Merged) > 0 {
			add("agent.memory_merged", sim.MemoryMergedPayload{
				Agent: sim.Ref(job.agent), Kept: g.Merge.Kept, Merged: g.Merge.Merged, Salience: g.Merge.Salience})
		}
		for _, rv := range g.Revisions {
			add("agent.salience_revised", sim.SalienceRevisedPayload{
				Agent: sim.Ref(job.agent), MemTick: rv.Ref.Tick, TextHash: rv.Ref.Hash,
				Salience: rv.Salience, Reason: sim.DreamReasonHabituation})
		}
	}
	add("agent.consolidated", sim.ConsolidatedPayload{
		Agent: sim.Ref(job.agent), Night: job.night, UpTo: job.upTo,
		Outcome:  sim.ConsolidationAccepted,
		Promoted: len(out.Promote), Faded: len(out.Fade), Beliefs: len(out.Beliefs),
		Coerced: coerced, DreamFolded: len(routine), DreamKept: len(groups) - len(routine),
		// Cost accrues across every ladder attempt; Retries makes a night that
		// survived truncation visible on its marker (spec 105 FR-005).
		CostUSD: res.CostUSD, Retries: res.Retries})

	if err := md.social.InjectSocial(batch); err != nil {
		log.Printf("mind: consolidation %s night %d injection rejected: %v", job.name, job.night, err)
		return
	}
	log.Printf("mind: consolidation %s night %d accepted (%d promoted, %d faded, %d beliefs, $%.4f)",
		job.name, job.night, len(out.Promote), len(out.Fade), len(out.Beliefs), res.CostUSD)
}

// landMarker records a non-accepted outcome (rejected / skipped_empty) as a
// single-event batch. The buffer stays intact for the next night. retries is
// the night's consumed truncation retries (spec 105 FR-005) — 0 everywhere
// but a rejected ladder night, and omitempty keeps that byte-identical.
func (md *Mind) landMarker(job consolJob, outcome, reason string, cost float64, retries int) {
	defer md.consolInFlight[job.agent].Store(false)
	b, _ := json.Marshal(sim.ConsolidatedPayload{
		Agent: sim.Ref(job.agent), Night: job.night, Outcome: outcome, Reason: reason, CostUSD: cost, Retries: retries})
	if err := md.social.InjectSocial([]store.Event{{Type: "agent.consolidated", Payload: b}}); err != nil {
		log.Printf("mind: consolidation %s night %d marker rejected: %v", job.name, job.night, err)
		return
	}
	switch outcome {
	case sim.ConsolidationRejected:
		log.Printf("mind: consolidation %s night %d rejected (%s)", job.name, job.night, reason)
	case sim.ConsolidationSkippedEmpty:
		log.Printf("mind: consolidation %s night %d skipped (empty)", job.name, job.night)
	}
}

func consolidateSystemPrompt(job consolJob) string {
	return fmt.Sprintf(`You are the sleeping mind of %s, a villager. %s
Tonight you digest the day into durable memory. You may only: strengthen or
let fade the day's memories, keep one gist of the day, revise beliefs, and
rewrite your self-narrative — all strictly in %s's voice and nature.
Your nature is fixed: %s. You must restate it verbatim in the "nature" field.`,
		job.name, job.personaMD, job.name, job.anchor)
}

func consolidateUserPrompt(job consolJob, groups []sim.DreamGroup) string {
	var b strings.Builder
	b.WriteString("Today's memories (reference them ONLY by their label, e.g. \"m3\"):\n")
	for i, m := range job.buffer {
		fmt.Fprintf(&b, "- [m%d] (salience %d) %s\n", i+1, m.Salience, m.Text)
	}
	if len(job.held) > 0 {
		b.WriteString("\nBeliefs you already hold:\n")
		for _, bl := range job.held {
			// Spec 030 (US2, FR-006): the model reads the EFFECTIVE confidence
			// (decayed, never the stored value) so it revises against what the
			// belief actually feels like tonight. Unlike other belief-surfacing
			// prompts, below-floor beliefs are NOT excluded here — they stay
			// listed by ID (still revisable; a reinforcement-worthy revision can
			// bring one back), just marked faded (data-model.md read sites).
			eff := sim.EffectiveConfidence(bl, job.sleepTick)
			faded := ""
			if eff < sim.BeliefConfidenceFloor {
				faded = " (faded)"
			}
			fmt.Fprintf(&b, "- [id %d] (confidence %d, %s) %s%s\n", bl.ID, eff, bl.Provenance, bl.Statement, faded)
		}
	}
	if job.social != "" {
		b.WriteString("\n" + job.social)
	}
	if job.narrative != "" {
		fmt.Fprintf(&b, "\nYour current self-narrative:\n%s\n", job.narrative)
	}
	if len(groups) > 0 {
		// Spec 098 (D2): the ambiguous band's consult. Absent groups add
		// nothing — a group-less night's prompt stays byte-identical to
		// pre-098, and only this block ever mentions the "routine" field.
		b.WriteString("\nSome of your stored memories blur together into near-identical groups. " +
			"For each group, say whether it is mere routine (safe to let fade into one) " +
			"or worth keeping as distinct moments:\n")
		for i, g := range groups {
			fmt.Fprintf(&b, "- [g%d] %d similar memories, e.g.:", i+1, g.Size)
			for _, ex := range g.Examples {
				fmt.Fprintf(&b, " %q", ex)
			}
			b.WriteString("\n")
		}
		b.WriteString("List in \"routine\" the group labels that are mere routine; omit the ones worth keeping.\n")
	}
	fmt.Fprintf(&b, "\nIn \"nature\", copy this line exactly, word for word: %s\n", job.anchor)
	b.WriteString("\nFor every belief, cite in \"evidence\" the memory labels it rests on. " +
		"Use \"witnessed\" ONLY for what you directly did or directly received (an omen or a dream); " +
		"a claim you only heard about in conversation is \"told\", and a conclusion you reasoned to is \"inferred\".\n")
	routineLine := ""
	if len(groups) > 0 {
		routineLine = "\n \"routine\": [\"g1\"],  // group labels that are mere routine; omit groups worth keeping"
	}
	fmt.Fprintf(&b, `
Reply with ONLY this JSON:
{"nature": "<your nature, restated verbatim>",
 "gist": "<ONE short sentence remembering this day, your voice, under 200 characters>",
 "promote": ["m1"],   // up to %d memory labels worth keeping sharp
 "fade": ["m2"],      // up to %d trivial memory labels to let go
 "beliefs": [{"id": 0, "statement": "...", "confidence": 0-100, "provenance": "witnessed|told|inferred", "source": -1, "subject": -1, "evidence": ["m3"]}],  // up to %d; id 0 = new belief, or a held belief's id to revise it; subject/source are villager numbers, -1 = none; evidence lists up to %d supporting memory labels, best first%s
 "narrative": "<who you are becoming, first person, your voice, max 1200 chars>"}`,
		maxPromotes, maxFades, maxBeliefEdits, maxBeliefEvidence, routineLine)
	return b.String()
}

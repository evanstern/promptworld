package guardian

// The guardian's own memory + nightly consolidation (spec 102, D5/FR-002):
// on an AGENTIZED world (tuning.json angel_cadence_ticks > 0) the guardian's
// notable moments — landed acts, watch fires, digests, player exchanges —
// enter a structured per-guardian memory store as recorded
// guardian.memory_added events, and each nightly boundary digests the
// un-consolidated tail through the SAME machinery villagers use:
//
//   - sim.PlanDream — the spec-098 geometry-first clustering pass, run over
//     the guardian's own store with its own seat (sim.GuardianSeat), clear
//     outcomes landing as recorded guardian dream events;
//   - the m-label / g-label reply contract, parsed by the SHARED
//     sim.ParseMemLabel / sim.ParseRoutineLabels / sim.FirstJSONObject;
//   - the shared marker vocabulary (sim.ConsolidationAccepted / Rejected)
//     and the KindConsolidation route + "consolidation" decision class.
//
// soul.md remains the persona SEED: on agentized worlds the 6-hour digest
// output lands as memories instead of soul appends (D5's "append-only
// digests become structured memory"), while the genesis header and act lines
// keep seeding the persona voice. A NON-agentized world takes none of these
// paths — byte-identical to pre-102 (FR-007).

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/evanstern/promptworld/internal/cognition"
	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
)

const (
	// guardianMaxBufferSent caps the prompt's buffer slice (the villager
	// night's maxBufferSent value — state keeps everything).
	guardianMaxBufferSent = 60
	// guardianMaxPromotes / guardianMaxFades mirror the villager night's caps
	// (internal/mind/validate.go): over-long lists are enthusiasm, kept as a
	// best-first prefix, never a rejected night.
	guardianMaxPromotes = 5
	guardianMaxFades    = 8
	// guardianMaxDreamGroups caps the ambiguous clusters one prompt carries
	// (the mind's maxDreamGroupsSent — un-sent groups retry a later night).
	guardianMaxDreamGroups = 4
	// guardianConsolTokens bounds the night's reply (gist + two label lists).
	guardianConsolTokens = 512
	// guardianMemorySalienceDefault bands: act/watch memories land vivid-ish,
	// digest lines land as background texture (the villager salience table's
	// spirit — never at the GenerationBumpSalience interrupt band).
	salGuardianAct    = 6
	salGuardianDigest = 3
	salGuardianTalk   = 3
)

// guardianConsolJob is the immutable snapshot one guardian night runs
// against — everything copied at enqueue time (the mind's consolJob shape).
type guardianConsolJob struct {
	night     int64
	sleepTick int64
	upTo      int64
	buffer    []sim.Memory
	mems      []sim.Memory
	seed      uint64
	dials     sim.DreamTuning
}

// agentized reports whether this world opted into guardian agentization
// (FR-007) — the stateMu-mirrored view worker goroutines read; the absorb
// goroutine reads the replica directly.
func (mt *Guardian) agentized() bool {
	mt.stateMu.Lock()
	defer mt.stateMu.Unlock()
	return mt.angelOn
}

// recordMemory lands one guardian.memory_added through the door (worker or
// absorb side; the injection detaches nothing — the door is async-safe).
// Gated on agentization: a non-opted world's guardian keeps soul.md alone.
// Text is clamped to the reducer's cap so an emitter can never bounce.
func (mt *Guardian) recordMemory(text string, salience int) {
	if mt.social == nil || !mt.agentized() {
		return
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	if len(text) > sim.GuardianMemoryTextMax {
		text = text[:sim.GuardianMemoryTextMax]
	}
	batch := []store.Event{{Type: "guardian.memory_added", Payload: mustJSON(sim.GuardianMemoryPayload{
		Text: text, Salience: salience})}}
	if err := mt.social.InjectSocial(batch); err != nil {
		log.Printf("guardian: memory rejected at the door: %v", err)
	}
}

// maybeConsolidateNight arms the nightly consolidation (absorb goroutine —
// it reads the replica). Called on sim.night_started; one night per index,
// single-flight, empty buffers close quietly (no marker spam for a guardian
// with nothing to remember).
func (mt *Guardian) maybeConsolidateNight(e store.Event) {
	if mt.social == nil || mt.replica.AngelCadence() <= 0 || mt.replica.Ended {
		return
	}
	night := sim.NightIndex(e.Tick)
	if night <= mt.lastConsolNight || mt.consolInFlight.Load() {
		return
	}
	buffer := mt.replica.GuardianEpisodicBuffer()
	if len(buffer) == 0 {
		mt.lastConsolNight = night
		return
	}
	// Router gate (D2 doctrine-completeness, the mind's maybeConsolidate
	// shape): the night rides the SHARED "consolidation" class; a suppression
	// skips the night — the next boundary retries.
	dc, _ := cognition.ClassFor("consolidation")
	spp := mt.secondsPerPoint(llm.KindConsolidation)
	allow := true
	switch {
	case mt.replica.Paused:
		allow = cognition.RoutePaused(dc, spp).Allow
	default:
		if tps := mt.replica.Speed.TicksPerSecond(); tps > 0 {
			allow = cognition.Route(dc, tps, spp).Allow
		}
	}
	if !allow {
		return
	}
	job := guardianConsolJob{
		night:     night,
		sleepTick: e.Tick,
		upTo:      buffer[len(buffer)-1].Tick,
		buffer:    append([]sim.Memory(nil), buffer...),
		mems:      append([]sim.Memory(nil), mt.replica.GuardianMemories...),
		seed:      mt.seed,
		dials:     mt.replica.DreamDials(),
	}
	if len(job.buffer) > guardianMaxBufferSent {
		job.buffer = job.buffer[len(job.buffer)-guardianMaxBufferSent:] // newest kept
	}
	mt.consolInFlight.Store(true)
	mt.lastConsolNight = night
	select {
	case mt.consolQ <- job:
	default:
		mt.consolInFlight.Store(false) // queue full: the next night retries
	}
}

// consolidateWorker drains guardian nights one call at a time.
func (mt *Guardian) consolidateWorker() {
	defer mt.wg.Done()
	for {
		select {
		case <-mt.done:
			return
		case job := <-mt.consolQ:
			mt.runConsolidation(job)
		}
	}
}

// guardianConsolOutput is the night's reply contract: a gist plus promote/
// fade/routine label lists — the villager shape minus the villager-only
// belief/nature/narrative fields (the guardian has no anchor to restate and
// no belief graph; its evidence lives in the event log).
type guardianConsolOutput struct {
	Gist    string   `json:"gist"`
	Promote []string `json:"promote"`
	Fade    []string `json:"fade"`
	Routine []string `json:"routine"`
}

func (mt *Guardian) runConsolidation(job guardianConsolJob) {
	defer mt.consolInFlight.Store(false)

	// The dream pass (spec 098, INCLUDED per D5): geometry first, over the
	// guardian's own snapshot alone. Clear-cut outcomes land NOW, as their
	// own recorded batch, independent of the LLM night's fate.
	dream := sim.PlanDream(job.mems, job.seed, job.night, sim.GuardianSeat, job.dials)
	if batch := sim.GuardianDreamEvents(dream.Revisions, dream.Merges); len(batch) > 0 {
		if err := mt.social.InjectSocial(batch); err != nil {
			log.Printf("guardian: dream night %d geometry batch rejected: %v", job.night, err)
		}
	}
	groups := dream.Ambiguous
	if len(groups) > guardianMaxDreamGroups {
		groups = groups[:guardianMaxDreamGroups]
	}

	ctx, cancel := context.WithTimeout(context.Background(), digestCallTimeout)
	resp, err := mt.orch.Submit(ctx, llm.Request{
		Kind:      llm.KindConsolidation,
		System:    guardianConsolSystem(mt.sk().Name()),
		Prompt:    guardianConsolPrompt(job, groups),
		MaxTokens: guardianConsolTokens,
	})
	cancel()
	if err != nil {
		// Transport/tier failure: no marker — the buffer stays; the next
		// nightly boundary retries (the villager FR-002 posture).
		log.Printf("guardian: consolidation night %d deferred: %v", job.night, err)
		return
	}

	raw, perr := sim.FirstJSONObject(resp.Text)
	var out guardianConsolOutput
	if perr != nil || json.Unmarshal([]byte(raw), &out) != nil || strings.TrimSpace(out.Gist) == "" {
		mt.landConsolMarker(job, sim.ConsolidationRejected, "unparseable", resp.CostUSD)
		return
	}
	if len(out.Promote) > guardianMaxPromotes {
		out.Promote = out.Promote[:guardianMaxPromotes]
	}
	if len(out.Fade) > guardianMaxFades {
		out.Fade = out.Fade[:guardianMaxFades]
	}
	routine := sim.ParseRoutineLabels(out.Routine, len(groups))

	// Build the whole night as one atomic batch (the villager shape).
	var batch []store.Event
	seen := map[int]bool{}
	promoted, faded := 0, 0
	for _, r := range out.Promote {
		i := sim.ParseMemLabel(r, len(job.buffer))
		if i < 0 || seen[i] {
			continue
		}
		seen[i] = true
		m := job.buffer[i]
		batch = append(batch, store.Event{Type: "guardian.memory_promoted", Payload: mustJSON(sim.GuardianMemoryPromotedPayload{
			MemTick: m.Tick, TextHash: sim.MemoryHash(m.Text), Boost: 3})})
		promoted++
	}
	for _, r := range out.Fade {
		i := sim.ParseMemLabel(r, len(job.buffer))
		if i < 0 || seen[i] {
			continue
		}
		seen[i] = true
		m := job.buffer[i]
		batch = append(batch, store.Event{Type: "guardian.memory_faded", Payload: mustJSON(sim.GuardianMemoryFadedPayload{
			MemTick: m.Tick, TextHash: sim.MemoryHash(m.Text)})})
		faded++
	}
	gist := strings.TrimSpace(out.Gist)
	if len(gist) > sim.GuardianMemoryTextMax {
		gist = gist[:sim.GuardianMemoryTextMax]
	}
	batch = append(batch, store.Event{Type: "guardian.memory_added", Payload: mustJSON(sim.GuardianMemoryPayload{
		Text: gist, Salience: sim.SalDayGist})})
	for _, gi := range routine {
		g := groups[gi]
		if len(g.Merge.Merged) > 0 {
			batch = append(batch, store.Event{Type: "guardian.memory_merged", Payload: mustJSON(sim.GuardianMemoryMergedPayload{
				Kept: g.Merge.Kept, Merged: g.Merge.Merged, Salience: g.Merge.Salience})})
		}
		for _, rv := range g.Revisions {
			batch = append(batch, store.Event{Type: "guardian.salience_revised", Payload: mustJSON(sim.GuardianSalienceRevisedPayload{
				MemTick: rv.Ref.Tick, TextHash: rv.Ref.Hash,
				Salience: rv.Salience, Reason: sim.DreamReasonHabituation})})
		}
	}
	batch = append(batch, store.Event{Type: "guardian.consolidated", Payload: mustJSON(sim.GuardianConsolidatedPayload{
		Night: job.night, UpTo: job.upTo, Outcome: sim.ConsolidationAccepted,
		Promoted: promoted, Faded: faded,
		DreamFolded: len(routine), DreamKept: len(groups) - len(routine),
		CostUSD: resp.CostUSD})})

	if err := mt.social.InjectSocial(batch); err != nil {
		log.Printf("guardian: consolidation night %d injection rejected: %v", job.night, err)
		return
	}
	log.Printf("guardian: consolidation night %d accepted (%d promoted, %d faded, $%.4f)",
		job.night, promoted, faded, resp.CostUSD)
}

// landConsolMarker records a rejected night as a single-event batch; the
// buffer stays intact for the next night.
func (mt *Guardian) landConsolMarker(job guardianConsolJob, outcome, reason string, cost float64) {
	b := mustJSON(sim.GuardianConsolidatedPayload{
		Night: job.night, Outcome: outcome, Reason: reason, CostUSD: cost})
	if err := mt.social.InjectSocial([]store.Event{{Type: "guardian.consolidated", Payload: b}}); err != nil {
		log.Printf("guardian: consolidation night %d marker rejected: %v", job.night, err)
		return
	}
	log.Printf("guardian: consolidation night %d %s (%s)", job.night, outcome, reason)
}

// guardianConsolSystem renders the night's system prompt — guardian-voiced,
// the skin's display name as validated single-line data (the digest keeper's
// discipline).
func guardianConsolSystem(name string) string {
	return "You are the sleeping mind of " + name + ", keeper and caretaker of the village. " +
		"Tonight you digest your recent watch into durable memory. You may only: strengthen " +
		"or let fade the recent notes, and keep one gist of this stretch of the watch. " +
		"No preamble — reply with only the JSON asked for."
}

// guardianConsolPrompt renders the night's user prompt: the m-labeled buffer,
// the optional g-labeled ambiguous dream groups, and the reply contract.
func guardianConsolPrompt(job guardianConsolJob, groups []sim.DreamGroup) string {
	var b strings.Builder
	b.WriteString("Your recent notes (reference them ONLY by their label, e.g. \"m3\"):\n")
	for i, m := range job.buffer {
		fmt.Fprintf(&b, "- [m%d] (salience %d) %s\n", i+1, m.Salience, m.Text)
	}
	if len(groups) > 0 {
		b.WriteString("\nSome of your stored notes blur together into near-identical groups. " +
			"For each group, say whether it is mere routine (safe to let fade into one) " +
			"or worth keeping as distinct moments:\n")
		for i, g := range groups {
			fmt.Fprintf(&b, "- [g%d] %d similar notes, e.g.:", i+1, g.Size)
			for _, ex := range g.Examples {
				fmt.Fprintf(&b, " %q", ex)
			}
			b.WriteString("\n")
		}
		b.WriteString("List in \"routine\" the group labels that are mere routine; omit the ones worth keeping.\n")
	}
	routineLine := ""
	if len(groups) > 0 {
		routineLine = ",\n \"routine\": [\"g1\"]  // group labels that are mere routine; omit groups worth keeping"
	}
	fmt.Fprintf(&b, `
Reply with ONLY this JSON:
{"gist": "<ONE short sentence remembering this stretch of the watch, your voice, under 200 characters>",
 "promote": ["m1"],   // up to %d note labels worth keeping sharp
 "fade": ["m2"]%s     // up to %d trivial note labels to let go
}`, guardianMaxPromotes, routineLine, guardianMaxFades)
	return b.String()
}

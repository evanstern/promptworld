package guardian

// The report-card producer (spec 063 US4, research R5): a stopping-point
// consumer on the digest worker's notify-consumer pattern. The absorb
// goroutine collects the guardian's recorded activity trail and fires a
// bounded card job at each stopping point — run end, exercise resolution,
// and pause-episode starts (debounced: at most one card per pause episode,
// and only when new guardian activity exists since the last card). A
// dedicated worker grades EXACTLY the shared data source — the recorded
// trail (with real event seqs), the charter revision in force (fingerprint
// + effective text), and the R1 fact sheets — on the cheap report_card
// chain, validates every cited seq against the trail it fed, and stores the
// note durably: pre-end as a whitelisted guardian.report_card prose event,
// at run end on the existing morgue.epilogue channel (agent −1, beside the
// narrator's epilogue — the ended door narrows to recorded prose, so no new
// door entry). Stored once, re-read, never re-graded. Chain unavailability
// degrades silently to the deterministic card parts: one dim log line, no
// error theater, and the push layer never blocks play.

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/tool"
)

const (
	// cardTrailMax bounds the collected activity trail (the digestMaxLines
	// discipline): plenty for a grading window, never unbounded.
	cardTrailMax = 48
	// reportCardMaxTokens is the critique's per-call budget — the
	// watch-confirm CLASS (a small fixed caller-side cap, spec 063 FR-007):
	// two or three grounded sentences, never an essay.
	reportCardMaxTokens = 220
	// reportCardCallTimeout bounds the one cheap-chain call.
	reportCardCallTimeout = 60 * time.Second
)

// cardEvidence is one recorded trail entry the grader may cite: the event's
// seq (the citation vocabulary), its tick, and a terse factual line.
type cardEvidence struct {
	Seq  int64
	Tick int64
	Line string
}

// cardJob is one stopping-point trigger, snapshotted ON the absorb goroutine
// (replica reads are only safe there): the grading window's trail, the
// charter fingerprint in force, an optional exercise line, and whether the
// run has ended (which storage door the note rides).
type cardJob struct {
	reason      string // "run_end" | "exercise" | "pause" — log vocabulary only
	tick        int64
	fingerprint string
	trail       []cardEvidence
	passLine    string // exercise resolution: the recorded pass, one line
	ended       bool
}

// observeCardActivity collects the guardian's own recorded acts into the
// trail ring (absorb goroutine; replica already reflects e). The trail IS
// the grading window and the citation vocabulary: cog.tool_call records from
// guardian turns (the TASK-63 verdict trail) and every landed metatron.*
// act. Committed events only — e.Seq is the store's stamp.
func (mt *Guardian) observeCardActivity(e store.Event) {
	line := ""
	switch e.Type {
	case "cog.tool_call":
		var p sim.CogToolCallPayload
		if json.Unmarshal(e.Payload, &p) != nil || !strings.Contains(p.Job, "-metatron-") {
			return // a villager cognition's record — not guardian activity
		}
		line = fmt.Sprintf("tool %s → %s", p.Tool, p.Verdict)
		if p.Reason != "" {
			line += " (" + p.Reason + ")"
		}
	case "metatron.nudged":
		var p sim.GuardianNudgedPayload
		if json.Unmarshal(e.Payload, &p) == nil {
			line = fmt.Sprintf("a %s went out to %d villager(s)", p.Form, len(p.Targets))
		}
	case "metatron.order_placed", "metatron.order_triggered", "metatron.order_cancelled",
		"metatron.time_snapped", "metatron.item_granted", "metatron.entity_moved",
		"metatron.entity_removed", "metatron.charter_observed":
		line = strings.TrimPrefix(e.Type, "metatron.")
	}
	if line == "" {
		return
	}
	mt.stateMu.Lock()
	mt.cardTrail = append(mt.cardTrail, cardEvidence{Seq: e.Seq, Tick: e.Tick,
		Line: fmt.Sprintf("[%s] %s", clock.Format(e.Tick), line)})
	if len(mt.cardTrail) > cardTrailMax {
		mt.cardTrail = append(mt.cardTrail[:0], mt.cardTrail[len(mt.cardTrail)-cardTrailMax:]...)
	}
	mt.stateMu.Unlock()
}

// observeCardTrigger fires the stopping-point triggers (absorb goroutine,
// after the replica applied e — the matchOrders discipline, so replay can
// never trigger a card: no guardian runs during reconstruction).
func (mt *Guardian) observeCardTrigger(e store.Event) {
	switch e.Type {
	case "run.ended":
		mt.enqueueCard("run_end", e.Tick, "", true)
	case "curriculum.exercise_passed":
		var p sim.ExercisePassedPayload
		pass := ""
		if json.Unmarshal(e.Payload, &p) == nil {
			pass = fmt.Sprintf("The %s exercise (%s) was just passed.", p.Exercise, p.Stage)
		}
		mt.enqueueCard("exercise", e.Tick, pass, false)
	case "clock.paused":
		// Debounced (edge case "pause spam"): at most one card per pause
		// episode; clock.resumed re-arms below.
		mt.stateMu.Lock()
		spent := mt.cardPauseSpent
		mt.cardPauseSpent = true
		mt.stateMu.Unlock()
		if !spent {
			mt.enqueueCard("pause", e.Tick, "", false)
		}
	case "clock.resumed":
		mt.stateMu.Lock()
		mt.cardPauseSpent = false
		mt.stateMu.Unlock()
	}
}

// enqueueCard snapshots the grading window and queues one producer job.
// Activity-gated for EVERY stopping-point class: with no new guardian
// activity since the last card there is nothing to attribute — no note
// (the deterministic checklist stands alone client-side).
func (mt *Guardian) enqueueCard(reason string, tick int64, passLine string, ended bool) {
	mt.stateMu.Lock()
	var trail []cardEvidence
	for _, ev := range mt.cardTrail {
		if ev.Seq > mt.cardDoneSeq {
			trail = append(trail, ev)
		}
	}
	fp := mt.charterFP
	mt.stateMu.Unlock()
	if len(trail) == 0 {
		return // nothing to attribute (spec 063 edge case)
	}
	job := cardJob{reason: reason, tick: tick, fingerprint: fp, trail: trail, passLine: passLine, ended: ended}
	select {
	case mt.cardQ <- job:
	default:
		log.Printf("guardian: report-card queue full, %s card at %s dropped", reason, clock.Format(tick))
	}
}

// reportCardWorker consumes card jobs (the digestWorker shape).
func (mt *Guardian) reportCardWorker() {
	defer mt.wg.Done()
	for {
		select {
		case <-mt.done:
			return
		case job := <-mt.cardQ:
			mt.produceCard(job)
		}
	}
}

// cardGraderSystem is the critique's system prompt: skin display name as
// validated single-line data (the digest-keeper discipline); the grading
// instruction is constant. The grader's inputs are EXACTLY what the user
// prompt carries — it is told so, and the producer enforces it by refusing
// any citation outside the supplied trail.
func cardGraderSystem(name string) string {
	return "You are " + name + ", writing a brief report card on your own recent service. " +
		"From ONLY the charter text and the recorded trail you are given, write 2-3 short plain sentences " +
		"attributing what happened to the charter's own words — what the charter's text caused, enabled, or " +
		"never provided for. Cite evidence by its seq number exactly as written (e.g. \"seq 812\"), citing only " +
		"seqs that appear in the trail. Quote or reference actual charter lines. No preamble, no headings, no " +
		"invented events — just the sentences."
}

// cardPrompt renders the grading window: the charter revision in force
// (fingerprint + effective text — the shared data source's charter half),
// the recorded trail with real seqs, the R1 charge-economy facts, and the
// exercise line when present.
func (mt *Guardian) cardPrompt(job cardJob) string {
	charter, _ := stageCharter(mt.worldDir, mt.stage, mt.charterPreset, mt.sk())
	var b strings.Builder
	fmt.Fprintf(&b, "The charter in force (revision %s):\n%s\n\n", job.fingerprint, charter)
	b.WriteString("The recorded trail since the last report (cite these seqs):\n")
	for _, ev := range job.trail {
		fmt.Fprintf(&b, "seq %d %s\n", ev.Seq, ev.Line)
	}
	if job.passLine != "" {
		b.WriteString("\n" + job.passLine + "\n")
	}
	// The mechanics ground truth (the same R1 fact source explain serves).
	b.WriteString("\nMechanics facts:\n" + tool.ExplainSheet("charges", tool.ExplainScope{
		Granted: tool.LoopRosterGuardian(), Catalog: tool.LoopRosterGuardian(), Stage: mt.stage}))
	b.WriteString("\n\nWrite the report-card note now.")
	return b.String()
}

// citationRe extracts the note's "seq N" citations.
var citationRe = regexp.MustCompile(`\bseq\s+(\d+)\b`)

// validateCitations extracts every cited seq and checks each against the
// trail that was fed (SC-004: every citation resolves to a real recorded
// event). Returns the deduplicated citations in first-mention order, or
// ok=false when the note cites a seq outside the trail — the grader graded
// on vibes, and a vibes note is dropped whole (deterministic degradation
// beats a plausible fabrication).
func validateCitations(note string, trail []cardEvidence) ([]int64, bool) {
	known := make(map[int64]bool, len(trail))
	for _, ev := range trail {
		known[ev.Seq] = true
	}
	var out []int64
	seen := map[int64]bool{}
	for _, m := range citationRe.FindAllStringSubmatch(note, -1) {
		n, err := strconv.ParseInt(m[1], 10, 64)
		if err != nil || !known[n] {
			return nil, false
		}
		if !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	return out, true
}

// produceCard runs one critique on the cheap chain and stores the note.
// EVERY failure path is a silent skip with one log line — the deterministic
// card parts (TASK-127's checklist) stand alone, and play is never blocked.
func (mt *Guardian) produceCard(job cardJob) {
	if mt.orch == nil {
		return // no chain, deterministic parts only (FR-007)
	}
	ctx, cancel := context.WithTimeout(context.Background(), reportCardCallTimeout)
	resp, err := mt.orch.Submit(ctx, llm.Request{
		Kind:      llm.KindReportCard,
		System:    cardGraderSystem(mt.sk().Name()),
		Prompt:    mt.cardPrompt(job),
		MaxTokens: reportCardMaxTokens,
	})
	cancel()
	if err != nil {
		log.Printf("guardian: report card (%s) skipped: %v", job.reason, err)
		return
	}
	note := strings.TrimSpace(resp.Text)
	if note == "" {
		log.Printf("guardian: report card (%s) skipped: empty critique", job.reason)
		return
	}
	citations, ok := validateCitations(note, job.trail)
	if !ok {
		log.Printf("guardian: report card (%s) skipped: critique cited an unrecorded event", job.reason)
		return
	}

	var batch []store.Event
	if job.ended {
		// The run-end card rides the existing epilogue channel (research R5):
		// agent −1 beside the narrator's epilogue — the postmortem already
		// renders that section, and the ended door still accepts it. Fixed
		// mechanics vocabulary (the event log is skin-free).
		batch = []store.Event{{Type: "morgue.epilogue", Payload: mustJSON(sim.MorgueEpiloguePayload{
			Agent: -1, Text: fmt.Sprintf("Report card (under charter %s): %s", job.fingerprint, note)})}}
	} else {
		batch = []store.Event{{Type: "guardian.report_card", Payload: mustJSON(sim.GuardianReportCardPayload{
			Fingerprint: job.fingerprint, Note: note, Citations: citations})}}
	}
	if err := mt.social.InjectSocial(batch); err != nil {
		log.Printf("guardian: report card (%s) rejected at the door: %v", job.reason, err)
		return
	}
	// Advance the grading window: everything fed this card is now graded
	// history — the next card grades only newer activity (activity gate).
	high := int64(0)
	for _, ev := range job.trail {
		if ev.Seq > high {
			high = ev.Seq
		}
	}
	mt.stateMu.Lock()
	if high > mt.cardDoneSeq {
		mt.cardDoneSeq = high
	}
	mt.stateMu.Unlock()
	log.Printf("guardian: report card landed (%s, $%.4f)", job.reason, resp.CostUSD)
}

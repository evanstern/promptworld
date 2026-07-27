package scribe

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/evanstern/promptworld/internal/persona"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// sliceSource is the test EventSource: an in-memory event log.
type sliceSource struct {
	mu     sync.Mutex
	events []store.Event
}

func (s *sliceSource) ReplayEvents(sinceSeq int64, fn func(store.Event) error) error {
	s.mu.Lock()
	evs := append([]store.Event(nil), s.events...)
	s.mu.Unlock()
	for _, e := range evs {
		if e.Seq <= sinceSeq {
			continue
		}
		if err := fn(e); err != nil {
			return err
		}
	}
	return nil
}

func (s *sliceSource) add(events ...store.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, events...)
}

// morgueHistory is the scripted no-LLM history the morgue tests fold over
// (spec 044 US2 Independent Test): two charter revisions, a standing order,
// a gift (relation + open debt), a build, a high-salience memory, and two
// deaths — Cedar on day 1 (before the charter edit), Birch on day 2 (after
// it, with the order standing). Every factual field derives from these
// events alone; no model anywhere.
func morgueHistory(t *testing.T) []store.Event {
	t.Helper()
	seq := int64(0)
	ev := func(tick int64, typ string, payload any) store.Event {
		seq++
		return store.Event{Seq: seq, Tick: tick, Type: typ, Payload: mustPayloadJSON(t, payload)}
	}
	return []store.Event{
		ev(0, "world.created", sim.WorldCreatedPayload{Name: "testworld", Seed: 42}),
		// Birch's lifetime notable memory (>= the scan threshold).
		ev(3600, "agent.memory_added", sim.MemoryAddedPayload{
			Agent: sim.Ref(1), Text: "Watched the gru circle the fire.", Salience: 9, Subject: sim.Ref(-1)}),
		// Birch matters to and is owed by Ash: a divine grant funds the gift
		// (gratis — no charge accounting in this scripted history), and the
		// gave arm opens Ash's debt to Birch.
		ev(4000, "metatron.item_granted", sim.ItemGrantedPayload{Agent: 1, Kind: "food_raw", Qty: 1, Gratis: true}),
		ev(4100, "social.gave", sim.GavePayload{From: 1, To: 0, Kind: "food"}),
		ev(4200, "social.relation_changed", sim.RelationChangedPayload{
			A: 1, B: 0, TrustDelta: 30, AffectionDelta: 5, Reason: "gift"}),
		// Birch's deed, on the chronicle's curated vocabulary.
		ev(7200, "agent.built", sim.BuiltPayload{Agent: sim.Ref(1), Kind: "fire", X: 10, Y: 10}),
		// The default charter observed on day 1; Cedar dies under it.
		ev(10000, "metatron.charter_observed", sim.CharterObservedPayload{Fingerprint: "aaaa11112222", Default: true}),
		ev(50000, "agent.died", sim.DiedPayload{Agent: sim.Ref(2), Cause: "exposure"}),
		// The player edits the charter (observed day 2) and places a watch;
		// Birch dies under BOTH.
		ev(90000, "metatron.charter_observed", sim.CharterObservedPayload{Fingerprint: "bbbb33334444", Default: false}),
		ev(91000, "metatron.order_placed", sim.GuardianOrder{
			ID: "ord-91000-1", Origin: "player",
			Condition: "anyone goes hungry", Action: "send a vision toward food",
			EventTypes: []string{"agent.needs_changed"}, Agent: 1,
			PlacedTick: 91000, ExpiresTick: 91000 + 2*86400}),
		ev(95000, "agent.died", sim.DiedPayload{Agent: sim.Ref(1), Cause: "starvation"}),
	}
}

// morgueEpilogueEvents extends the history with recorded narrator prose —
// the per-death epilogue for Birch.
func morgueEpilogueEvents(t *testing.T, fromSeq int64) []store.Event {
	t.Helper()
	seq := fromSeq
	ev := func(tick int64, typ string, payload any) store.Event {
		seq++
		return store.Event{Seq: seq, Tick: tick, Type: typ, Payload: mustPayloadJSON(t, payload)}
	}
	return []store.Event{
		ev(96000, "morgue.epilogue", sim.MorgueEpiloguePayload{
			Agent: 1, Text: "Birch kept the fire while the others slept."}),
	}
}

// newMorgueScribe builds a scribe over a scripted history and returns the
// scribe, its source, and the world dir. The boot render runs before New
// returns, so morgue.md exists immediately.
func newMorgueScribe(t *testing.T, events []store.Event) (*Scribe, *sliceSource, string) {
	t.Helper()
	dir := t.TempDir()
	if err := persona.Genesis(dir); err != nil {
		t.Fatal(err)
	}
	m := worldmap.Generate(42, 64, 64)
	state := sim.NewState(42, m)
	src := &sliceSource{events: append([]store.Event(nil), events...)}
	scr, err := New(dir, 42, m, state.Marshal(), src)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(scr.Close)
	return scr, src, dir
}

func readMorgue(t *testing.T, dir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "morgue.md"))
	if err != nil {
		t.Fatalf("reading morgue.md: %v", err)
	}
	return string(data)
}

// TestMorgueRendersFactualEpitaph (SC-002 / FR-007): on a no-LLM history,
// every factual epitaph field is present and derived from recorded history —
// name, days survived, cause, notable deeds, notable memories, relationships,
// and debts — plus the guardian-policy evidence (SC-003 / FR-008): the charter
// revision in force at EACH death and the active orders' watch subjects.
func TestMorgueRendersFactualEpitaph(t *testing.T) {
	_, _, dir := newMorgueScribe(t, morgueHistory(t))
	s := readMorgue(t, dir)

	for _, want := range []string{
		"# Morgue — testworld",
		// Birch's epitaph: the seven factual fields.
		"## Birch — died day 2 (starvation)",
		"- **Days survived**: 2",
		"- **Cause**: starvation",
		"Birch built a fire.",
		"(9★) Watched the gru circle the fire.",
		"Ash: trust +30, affection +5",
		"- **Debts**: owed — none · owed to them — one food from Ash",
		// Evidence alignment (SC-003): Birch died under the EDITED charter
		// and the standing order; Cedar died under the default one.
		"charter revision `bbbb33334444` (player-authored), in force since day 2",
		`"anyone goes hungry" → "send a vision toward food" — watching agent.needs_changed; villager Birch`,
		"## Cedar — died day 1 (exposure)",
		"charter revision `aaaa11112222` (default), in force since day 1",
		"_Stated as evidence; the reader draws the lesson._",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("morgue.md missing %q; content:\n%s", want, s)
		}
	}
	// Deaths render in event order: Cedar's section before Birch's.
	if strings.Index(s, "## Cedar") > strings.Index(s, "## Birch") {
		t.Error("epitaph sections out of event order")
	}
	// Cedar died before the order was placed — no order may appear there.
	cedar := s[strings.Index(s, "## Cedar"):strings.Index(s, "## Birch")]
	if !strings.Contains(cedar, "standing orders active: none.") {
		t.Errorf("Cedar's watch should name no orders:\n%s", cedar)
	}
}

// TestMorgueRunSummary (FR-009): run end closes the document with run length,
// the day-stamped population decline, every death with cause, and the run's
// notable events.
func TestMorgueRunSummary(t *testing.T) {
	events := morgueHistory(t)
	seq := events[len(events)-1].Seq
	// The remaining six die on day 2 and the run ends (scripted ledger).
	deaths := []sim.DeathRecord{{Agent: 2, Tick: 50000, Cause: "exposure"}, {Agent: 1, Tick: 95000, Cause: "starvation"}}
	for i, a := range []int{0, 3, 4, 5, 6, 7} {
		seq++
		tick := int64(97000 + i)
		events = append(events, store.Event{Seq: seq, Tick: tick, Type: "agent.died",
			Payload: mustPayloadJSON(t, sim.DiedPayload{Agent: sim.Ref(a), Cause: "exposure"})})
		deaths = append(deaths, sim.DeathRecord{Agent: a, Tick: tick, Cause: "exposure"})
	}
	seq++
	events = append(events, store.Event{Seq: seq, Tick: 97005, Type: "run.ended",
		Payload: mustPayloadJSON(t, sim.RunEndedPayload{Tick: 97005, Deaths: deaths, FinalCause: "exposure"})})

	_, _, dir := newMorgueScribe(t, events)
	s := readMorgue(t, dir)
	for _, want := range []string{
		"## The run — ended day 2",
		"- **Run length**: 2 days",
		"- **Population**: 8 → 7 (day 1) → 6 (day 2)", // decline starts with Cedar, then Birch…
		"→ 0 (day 2)",                                 // …and reaches zero
		"- **The deaths**:",
		"**day 1** — Cedar (exposure)",
		"**day 2** — Birch (starvation)",
		"- **Notable events of the run**:",
		"Birch built a fire.",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("run summary missing %q; content:\n%s", want, s)
		}
	}
}

// TestMorgueReplayByteIdentity (SC-004 / FR-011): two independent scribes
// folding the same history render byte-identical morgue documents.
func TestMorgueReplayByteIdentity(t *testing.T) {
	events := morgueHistory(t)
	_, _, dirA := newMorgueScribe(t, events)
	_, _, dirB := newMorgueScribe(t, events)
	a, b := readMorgue(t, dirA), readMorgue(t, dirB)
	if a != b {
		t.Fatalf("replayed morgue renders differ:\n%s\n---\n%s", a, b)
	}
}

// epilogueLineRe matches a rendered epilogue blockquote (one line, preceded
// by the blank line the renderer emits).
var epilogueLineRe = regexp.MustCompile(`\n> _Epilogue_ — [^\n]*\n`)

// TestMorgueEpilogueSeparation (FR-010 / SC-004): epilogues render
// blockquote-delimited after their section's facts, and stripping them
// reproduces the narrator-off document byte-for-byte — the facts never
// depend on the model.
func TestMorgueEpilogueSeparation(t *testing.T) {
	facts := morgueHistory(t)
	_, _, dirOff := newMorgueScribe(t, facts)

	withProse := append(append([]store.Event(nil), facts...), morgueEpilogueEvents(t, facts[len(facts)-1].Seq)...)
	_, _, dirOn := newMorgueScribe(t, withProse)

	off, on := readMorgue(t, dirOff), readMorgue(t, dirOn)
	if !strings.Contains(on, "> _Epilogue_ — Birch kept the fire while the others slept.") {
		t.Fatalf("epilogue not rendered as a blockquote:\n%s", on)
	}
	// The epilogue follows Birch's factual section, before the next content.
	if strings.Index(on, "> _Epilogue_ — Birch kept") < strings.Index(on, "_Stated as evidence") {
		t.Error("epilogue rendered before its section's facts")
	}
	if stripped := epilogueLineRe.ReplaceAllString(on, ""); stripped != off {
		t.Errorf("factual bytes changed by the narrator:\nwith prose stripped:\n%s\n---\nnarrator off:\n%s", stripped, off)
	}
}

// TestMorgueRegeneratesAfterDeletion (FR-011 edge case): a deleted morgue.md
// is healed by the next render — recorded history stays the source of truth.
func TestMorgueRegeneratesAfterDeletion(t *testing.T) {
	scr, src, dir := newMorgueScribe(t, morgueHistory(t))
	before := readMorgue(t, dir)
	if err := os.Remove(filepath.Join(dir, "morgue.md")); err != nil {
		t.Fatal(err)
	}

	// The next morgue-dirty batch heals the file with identical factual bytes.
	heal := morgueEpilogueEvents(t, src.events[len(src.events)-1].Seq)
	src.add(heal...)
	scr.Observe(heal)

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(filepath.Join(dir, "morgue.md"))
		if err == nil && strings.Contains(string(data), "> _Epilogue_") {
			if got := epilogueLineRe.ReplaceAllString(string(data), ""); got != before {
				t.Fatalf("healed factual bytes differ:\n%s\n---\n%s", got, before)
			}
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("morgue.md never regenerated after deletion")
}

// TestMorgueBannedVocabulary (FR-008 / contract invariant 2): the factual
// render carries no scoring or blame language — evidence, never a grade.
// Word-boundary matches: the banned WORDS are what the contract names
// ("default" is provenance, not blame).
func TestMorgueBannedVocabulary(t *testing.T) {
	_, _, dir := newMorgueScribe(t, morgueHistory(t))
	s := strings.ToLower(readMorgue(t, dir))
	for _, banned := range []string{`\bscore\b`, `\bgrade\b`, `\bblame\b`, `\bfault\b`, `\bshould have\b`} {
		if regexp.MustCompile(banned).MatchString(s) {
			t.Errorf("factual render contains contract-banned vocabulary %q", banned)
		}
	}
}

// TestMorgueEmptyWorldRender: the boot render exists even before any death
// (the contract renders at every boot), with the no-deaths posture line.
func TestMorgueEmptyWorldRender(t *testing.T) {
	_, _, dir := newMorgueScribe(t, nil)
	s := readMorgue(t, dir)
	if !strings.Contains(s, "*No one has died. The village lives.*") {
		t.Errorf("empty-world morgue missing the no-deaths line:\n%s", s)
	}
}

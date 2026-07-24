package mind

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// The villager system-prompt frame contract (spec 027,
// contracts/system-prompt.md C1–C5). These tests are meaning-pinned, not
// wording-pinned: doctrine is asserted by what the frame must convey (C3), so a
// craft rewrite of the phrasing keeps them green while a doctrine drift turns
// them red.

// sentinelName is collision-proof: it must not occur anywhere in the frame's
// static text or in the sample personas, so counting it isolates the identity
// statement's single interpolation (C2).
const sentinelName = "Zzyzxonymously"

// framePersona is a representative persona that deliberately does NOT contain
// sentinelName, so it can be stripped to leave the pure frame text (C2/C4).
const framePersona = `# A Villager

**Temperament:** steady, practical, slow to anger.
**Drives:** keep everyone fed; distrusts idleness.`

// frameWithoutPersona returns the rendered frame with the interpolated persona
// text removed, so assertions see only the static frame (contracts C2/C4: the
// persona text is exempt from frame-text rules).
func frameWithoutPersona(t *testing.T, name, persona string) string {
	t.Helper()
	rendered := systemPrompt(name, persona)
	if persona == "" {
		return rendered
	}
	if !strings.Contains(rendered, persona) {
		t.Fatalf("persona text not present verbatim in rendered frame (C4)")
	}
	return strings.Replace(rendered, persona, "", 1)
}

// C1 — purity / cacheability: identical inputs render byte-identical output,
// across repeated calls (SC-005). The signature (name, personaText only)
// structurally forbids dynamic world state; this pins the determinism half.
func TestSystemPromptPurity(t *testing.T) {
	a := systemPrompt(sentinelName, framePersona)
	for i := 0; i < 8; i++ {
		if b := systemPrompt(sentinelName, framePersona); b != a {
			t.Fatalf("render %d differs from first render — prompt is not a pure function (C1)", i)
		}
	}
	// Distinct names differ only where the identity statement (and persona)
	// differ — here personas match, so any difference is the name.
	if systemPrompt("Aria", framePersona) == systemPrompt("Bram", framePersona) {
		t.Fatalf("renders for distinct names are identical — the identity statement is not naming the agent (C1)")
	}
}

// C2 — single naming: the frame text (persona exempt) contains the agent's name
// exactly once. MUST FAIL against the pre-rewrite prompt, which repeats it.
func TestSystemPromptNamesOnce(t *testing.T) {
	frame := frameWithoutPersona(t, sentinelName, framePersona)
	if n := strings.Count(frame, sentinelName); n != 1 {
		t.Fatalf("frame text names the agent %d times, want exactly 1 (C2, SC-002)", n)
	}
	// Also holds with an empty persona (the whole render is frame text).
	if n := strings.Count(systemPrompt(sentinelName, ""), sentinelName); n != 1 {
		t.Fatalf("empty-persona frame names the agent %d times, want exactly 1 (C2)", n)
	}
}

// C3 — doctrine (meaning-pinned, wording-free). The frame must convey the four
// invariants regardless of phrasing.
func TestSystemPromptDoctrine(t *testing.T) {
	frame := frameWithoutPersona(t, sentinelName, framePersona)
	lower := strings.ToLower(frame)

	// 1. Acting-tool-only: the decision is made by calling exactly ONE acting tool.
	if !strings.Contains(lower, "exactly one") {
		t.Errorf("doctrine 1 (acting-tool-only): frame does not state the decision is exactly one call (C3)")
	}
	if !strings.Contains(lower, "tool") {
		t.Errorf("doctrine 1 (acting-tool-only): frame does not reference tools (C3)")
	}

	// 2. Read-then-act: read-only tools may precede the one acting call.
	if !strings.Contains(lower, "read") {
		t.Errorf("doctrine 2 (read-then-act): frame does not mention read-only tools (C3)")
	}
	if !(strings.Contains(lower, "first") || strings.Contains(lower, "before") || strings.Contains(lower, "then")) {
		t.Errorf("doctrine 2 (read-then-act): frame does not convey read-before-act ordering (C3)")
	}

	// 3. Muse-is-an-action with opportunity-cost framing: muse and set_plan are
	// themselves acting choices, and spending a beat thinking costs a beat of
	// doing (this exact idea, not necessarily these words).
	if !strings.Contains(lower, "muse") {
		t.Errorf("doctrine 3 (muse-is-an-action): frame does not mention muse (C3)")
	}
	if !strings.Contains(lower, "set_plan") {
		t.Errorf("doctrine 3 (muse-is-an-action): frame does not mention set_plan as an action (C3)")
	}
	// opportunity cost: contrast a thinking beat against a doing beat.
	thinks := strings.Contains(lower, "think") || strings.Contains(lower, "musing")
	acts := strings.Contains(lower, "doing") || strings.Contains(lower, "acting") || strings.Contains(lower, "act")
	if !(thinks && acts) {
		t.Errorf("doctrine 3 (muse-is-an-action): frame does not convey the thinking-vs-doing opportunity cost (C3)")
	}

	// 4. No free-text path: the frame never invites a prose/JSON text answer.
	for _, banned := range []string{"respond with", "reply with", "json", "output format", "free text", "free-text", "in the form of"} {
		if strings.Contains(lower, banned) {
			t.Errorf("doctrine 4 (no free-text path): frame offers a text/format answer channel via %q (C3)", banned)
		}
	}
}

// C4 — persona block: verbatim, its own block between identity and task framing;
// empty persona renders a clean frame (no doubled blank lines / dangling
// separator).
func TestSystemPromptPersonaBlock(t *testing.T) {
	// Verbatim + ordered: identity (the name) before persona before task framing.
	rendered := systemPrompt(sentinelName, framePersona)
	if !strings.Contains(rendered, framePersona) {
		t.Fatalf("persona text not present verbatim (C4)")
	}
	iName := strings.Index(rendered, sentinelName)
	iPersona := strings.Index(rendered, framePersona)
	iTask := strings.Index(strings.ToLower(rendered), "exactly one")
	if !(iName >= 0 && iPersona > iName && iTask > iPersona) {
		t.Fatalf("frame parts out of order: name@%d persona@%d task@%d, want name<persona<task (C4)", iName, iPersona, iTask)
	}

	// Empty persona: clean render, no tripled newline where the block would be,
	// and identity still precedes task framing.
	empty := systemPrompt(sentinelName, "")
	if strings.Contains(empty, "\n\n\n") {
		t.Errorf("empty-persona render has a doubled blank line / dangling separator (C4): %q", empty)
	}
	if ei, et := strings.Index(empty, sentinelName), strings.Index(strings.ToLower(empty), "exactly one"); !(ei >= 0 && et > ei) {
		t.Errorf("empty-persona frame parts out of order: name@%d task@%d (C4)", ei, et)
	}
}

// TestPromptFrameReport renders the frame for a fixed representative sample
// agent and logs byte / word / approximate-token counts (research D4:
// approx tokens = len(bytes)/4). Run with `-run TestPromptFrameReport -v` at
// each variant's git ref; the numbers land in eval/<variant>.md (SC-004).
func TestPromptFrameReport(t *testing.T) {
	const sampleName = "Ash"
	const samplePersona = `# Ash

**Temperament:** steady, practical, slow to anger.
**Drives:** keep everyone fed; distrusts idleness.
**Quirk:** talks to the fire as if it answers.
**Bonds:** protective of Fern; an old, quiet rivalry with Oak.
`
	frame := systemPrompt(sampleName, samplePersona)
	bytes := len(frame)
	words := len(strings.Fields(frame))
	tokensApprox := bytes / 4
	t.Logf("PROMPT_FRAME_REPORT sample=%q prompt_bytes=%d prompt_words=%d prompt_tokens_approx=%d",
		sampleName, bytes, words, tokensApprox)
}

// --- spec 041 US2: the prompt renders only what the agent knows (T018) ------

// knownPlacesState builds a State whose map is attached (MapDims/frontier
// work) with all mental maps emptied — each test grants exactly the knowledge
// it exercises.
func knownPlacesState(t *testing.T) *sim.State {
	t.Helper()
	s := sim.NewState(42, worldmap.Generate(42, 64, 64))
	for i := range s.Agents {
		s.Agents[i].Map.Facts = nil
		s.Agents[i].Map.Peers = nil
	}
	return s
}

func fact(kind string, x, y int, prov string, src int, detail int64) sim.PlaceFact {
	return sim.PlaceFact{Kind: kind, X: x, Y: y, Seen: 1, Provenance: prov, Source: src, Detail: detail}
}

// addFact/addSighting assign through the exported fields, preserving the
// canonical sort invariants (the reducer-side mutators are sim-internal).
func addFact(mm *sim.MentalMap, f sim.PlaceFact) {
	mm.Facts = append(mm.Facts, f)
	sort.Slice(mm.Facts, func(i, j int) bool {
		a, b := mm.Facts[i], mm.Facts[j]
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		if a.X != b.X {
			return a.X < b.X
		}
		return a.Y < b.Y
	})
}

func addSighting(mm *sim.MentalMap, agent, x, y int, seen int64) {
	mm.Peers = append(mm.Peers, sim.PeerSighting{Agent: agent, X: x, Y: y, Seen: seen})
	sort.Slice(mm.Peers, func(i, j int) bool { return mm.Peers[i].Agent < mm.Peers[j].Agent })
}

// TestKnownPlacesDivergentMaps (SC-002): two villagers in the SAME world
// state with different histories render different known-places sections, and
// neither mentions a place it has not learned — the omniscient Village: line
// is gone.
func TestKnownPlacesDivergentMaps(t *testing.T) {
	s := knownPlacesState(t)
	// World truth holds a structure NEITHER villager may see in its prompt.
	s.Structures = []sim.Structure{{Kind: "oven", X: 50, Y: 50}}
	addFact(s.Agents[0].Map, fact("fire", 12, 34, "witnessed", 0, 900000))
	addFact(s.Agents[1].Map, fact("shelter", 8, 20, "witnessed", 0, 0))

	p0, p1 := userPrompt(s, 0, 4), userPrompt(s, 1, 4)
	if p0 == p1 {
		t.Fatal("divergent maps rendered identical prompts")
	}
	if !strings.Contains(p0, "fire at (12,34)") || strings.Contains(p0, "shelter at (8,20)") {
		t.Errorf("agent 0 section wrong:\n%s", p0)
	}
	if !strings.Contains(p1, "shelter at (8,20)") || strings.Contains(p1, "fire at (12,34)") {
		t.Errorf("agent 1 section wrong:\n%s", p1)
	}
	for i, p := range []string{p0, p1} {
		if strings.Contains(p, "oven") {
			t.Errorf("agent %d prompt leaked the unlearned oven (omniscient render):\n%s", i, p)
		}
	}
}

// TestKnownPlacesNoCap (SC-002): more than six known structures ALL render —
// the first-6 truncation is retired.
func TestKnownPlacesNoCap(t *testing.T) {
	s := knownPlacesState(t)
	for i := 0; i < 9; i++ {
		addFact(s.Agents[0].Map, fact("fire", 10+i, 20, "witnessed", 0, 900000))
	}
	p := userPrompt(s, 0, 4)
	for i := 0; i < 9; i++ {
		if !strings.Contains(p, fmt.Sprintf("fire at (%d,20)", 10+i)) {
			t.Fatalf("structure %d of 9 missing — a cap survives:\n%s", i+1, p)
		}
	}
}

// TestKnownPlacesProvenancePhrasing (contracts §3): told facts name the
// teller, revealed facts name the vision, witnessed facts render plain; a
// fire remembered as burned out says so.
func TestKnownPlacesProvenancePhrasing(t *testing.T) {
	s := knownPlacesState(t)
	s.Tick = 50000
	mm := s.Agents[0].Map
	addFact(mm, sim.PlaceFact{Kind: "fire", X: 40, Y: 12, Seen: 49000, Provenance: "told", Source: 1, Detail: 60000})
	addFact(mm, sim.PlaceFact{Kind: "shelter", X: 9, Y: 9, Seen: 49000, Provenance: "revealed"})
	addFact(mm, sim.PlaceFact{Kind: "oven", X: 5, Y: 5, Seen: 49000, Provenance: "witnessed"})
	addFact(mm, sim.PlaceFact{Kind: "fire", X: 6, Y: 6, Seen: 49000, Provenance: "witnessed", Detail: 49500}) // burned out by 50000

	p := userPrompt(s, 0, 4)
	for _, want := range []string{
		"a fire at (40,12) — Birch told you",
		"a shelter at (9,9), shown to you in a vision",
		"oven at (5,5)",
		"fire at (6,6) (likely burned out by now)",
	} {
		if !strings.Contains(p, want) {
			t.Errorf("prompt missing %q:\n%s", want, p)
		}
	}
}

// TestKnownPlacesGroupsAndEmptyState (contracts §3): resource kinds group
// with count + nearest; a villager with no known landmark structures gets the
// explicit empty-state line, never silence.
func TestKnownPlacesGroupsAndEmptyState(t *testing.T) {
	s := knownPlacesState(t)
	a := &s.Agents[0]
	a.X, a.Y = 20, 20
	addFact(a.Map, fact("forage", 22, 41, "witnessed", 0, 0))
	addFact(a.Map, fact("forage", 21, 20, "witnessed", 0, 0))
	addFact(a.Map, fact("forage", 30, 30, "witnessed", 0, 0))
	addFact(a.Map, fact("tree", 18, 9, "witnessed", 0, 0))

	p := userPrompt(s, 0, 4)
	if !strings.Contains(p, "You know of no fires or shelters yet.") {
		t.Errorf("empty landmark state must be explicit, not silent:\n%s", p)
	}
	if !strings.Contains(p, "3 forage spots (nearest (21,20))") {
		t.Errorf("forage group wrong (count + nearest):\n%s", p)
	}
	if !strings.Contains(p, "1 stand of trees (nearest (18,9))") {
		t.Errorf("tree group wrong:\n%s", p)
	}
	// A stale fact is invisible: age the forage past the durable horizon.
	s.Tick = 1 + 5*86400
	p = userPrompt(s, 0, 4)
	if strings.Contains(p, "forage spot") {
		t.Errorf("stale facts must not render:\n%s", p)
	}
}

// TestKnownPlacesFrontierLine (SC-002): a partially-explored map renders the
// one-line orientation toward the unknown; a fully-explored map omits it.
func TestKnownPlacesFrontierLine(t *testing.T) {
	s := knownPlacesState(t)
	p := userPrompt(s, 0, 4)
	if !strings.Contains(p, "is unknown to you.") {
		t.Errorf("spawn-only exploration must orient toward the unknown:\n%s", p)
	}
	// Fully explored: mark everything.
	w, h := s.MapDims()
	s.Agents[0].Map.MarkExplored(w, h, w/2, h/2, w+h)
	p = userPrompt(s, 0, 4)
	if strings.Contains(p, "unknown to you") {
		t.Errorf("fully-explored map must omit the unknown-land line:\n%s", p)
	}
}

// TestNearbyFromSightings (spec 041 US2/T017): the Nearby line renders from
// the viewer's peer sightings — a villager never seen is absent even when
// physically close, and a remembered position wins over live coordinates.
func TestNearbyFromSightings(t *testing.T) {
	s := knownPlacesState(t)
	a := &s.Agents[0]
	// Birch stands 2 tiles away but has never been seen: absent.
	s.Agents[1].X, s.Agents[1].Y = a.X+2, a.Y
	p := userPrompt(s, 0, 4)
	if strings.Contains(p, "Nearby:") {
		t.Errorf("unsighted neighbor leaked into Nearby (omniscient render):\n%s", p)
	}
	// A sighting 3 tiles away renders with the REMEMBERED distance even after
	// Birch moves far off unseen.
	addSighting(a.Map, 1, a.X+3, a.Y, 100)
	s.Agents[1].X, s.Agents[1].Y = a.X+30, a.Y
	p = userPrompt(s, 0, 4)
	if !strings.Contains(p, "Nearby: Birch (3 tiles away)") {
		t.Errorf("sighting-sourced Nearby missing:\n%s", p)
	}
}

package sim

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// Episodic memory: deterministic emission heuristics (research R2) and the
// deterministic working-memory window (research R3). Generation is the
// executor's job; selection is a pure function shared by the mind's prompts
// and the tests.

// --- spec 019 (US1): situated episodic memories ---
//
// Every episodic memory the sim emits is situated (SC-001): these constructors
// bake the where/why context into the agent.memory_added payload AND compose it
// into the memory text via the shared grammar helper (situateText). The
// salience/subject/tone semantics are unchanged — this layer situates memories,
// it does not re-weigh them. There are three, mirroring the memory shapes:
// personal (situatedMemoryEvent), personal-with-tone (situatedMemoryToned), and
// gossip/witness about another agent (situatedMemoryAboutEvent, which carries no
// Why — a witness never drove the act). Spec 019 T008b removed the pre-019 bare
// constructors once every emission site was migrated, so no sim memory can be
// emitted unsituated: a new memory site must pick a situated constructor and
// therefore a Where.

// --- spec 030: memory provenance origin (model-free) ---
//
// Origin is stamped at every emission site (the closed vocabulary below) and is
// the ONLY signal the belief validator reads to classify direct perception —
// no text inspection, no heuristics (FR-002). Direct perception = an own act
// (OriginAction), a witnessed event (OriginWitness), or a delivered omen/dream
// (OriginOmen). Secondhand = a chest-owner's any-distance report (OriginReport),
// a conversation gist (OriginGist), a nightly digest (OriginDigest), or an
// absent/legacy origin ("" — treated as secondhand, the conservative direction:
// hygiene may under-grant "witnessed", never over-grant it).
const (
	OriginAction  = "action"  // own executed act (situated personal constructors)
	OriginWitness = "witness" // saw it happen (situated about-event constructor)
	OriginReport  = "report"  // learned of it at any distance (chest-owner taking memory)
	OriginOmen    = "omen"    // a delivered omen/dream/working (the guardian) — FROZEN payload value (spec 052 ruling 2)
	OriginGist    = "gist"    // a conversation summary written into memory
	OriginDigest  = "digest"  // a nightly day-gist
)

// DirectPerception reports whether a memory's recorded origin is a direct
// perception. Pure function on the stored field (FR-002): OriginAction /
// OriginWitness / OriginOmen are direct; everything else (report, gist, digest,
// absent/legacy) is secondhand. The belief validator gates "witnessed" on this.
func DirectPerception(origin string) bool {
	switch origin {
	case OriginAction, OriginWitness, OriginOmen:
		return true
	default:
		return false
	}
}

// placeScanRadius bounds describePlace's deterministic feature scan (Manhattan).
const placeScanRadius = 2

// describePlace returns a deterministic terrain/feature description for the tile
// at (x,y) — the nearest notable feature (station or terrain) within a small
// fixed radius, scanned in a fixed ring order so the same (state, x, y) always
// yields the same string. "" when nothing notable is near (coords alone
// situate the memory). Baked into the event at emission (research R3); the
// scribe renders what the payload carries and never re-derives, so replay is
// byte-identical with no map lookup.
func describePlace(s *State, x, y int) string { return describePlaceExcept(s, x, y, "") }

// describePlaceExcept is describePlace with one structure kind held out of the
// scan — the build-memory fix (T024): a build completion describes its tile as
// it was WITHOUT the thing just placed (and without any same-kind neighbour), so
// "Built a fire" never resolves to "at the fire". Excluding the kind is
// deterministic and needs no ordering dance with the not-yet-reduced built
// event. excludeKind == "" is the ordinary describePlace.
func describePlaceExcept(s *State, x, y int, excludeKind string) string {
	if s.m == nil {
		return ""
	}
	for r := 0; r <= placeScanRadius; r++ {
		for dy := -r; dy <= r; dy++ {
			dx := r - abs(dy)
			if d := featureDesc(s, x+dx, y+dy, excludeKind); d != "" {
				return d
			}
			if dx != 0 {
				if d := featureDesc(s, x-dx, y+dy, excludeKind); d != "" {
					return d
				}
			}
		}
	}
	return ""
}

// featureDesc names the notable feature on one tile — a station structure
// first (the most salient), then the terrain kind — as a noun phrase that reads
// after "at" ("the fire", "the rock outcrop"). A structure whose kind equals
// excludeKind is skipped (build fix, T024). "" for ordinary or off-map tiles.
func featureDesc(s *State, x, y int, excludeKind string) string {
	if s.m == nil || x < 0 || y < 0 || x >= s.m.W || y >= s.m.H {
		return ""
	}
	for _, st := range s.Structures {
		if st.X == x && st.Y == y && st.Kind != excludeKind {
			switch st.Kind {
			case "fire":
				return "the fire"
			case "shelter":
				return "the shelter"
			case "oven":
				return "the oven"
			case "chest":
				return "the chest"
			}
		}
	}
	switch effectiveKind(s.m, s, x, y) {
	case worldmap.Water:
		return "the water"
	case worldmap.Tree:
		return "the woods"
	case worldmap.Rock:
		return "the rock outcrop"
	case worldmap.Forage:
		return "the forage patch"
	case worldmap.Marsh:
		// Spec 068 (FR-008/C13): the new ground covers are NOTABLE terrain —
		// named to agents, never a fallback label — unlike grass, whose ""
		// below means "ordinary ground, nothing to say".
		return "the marsh"
	case worldmap.Sand:
		return "the sand flat"
	}
	return ""
}

// PlaceAt returns the situated location of a memory formed at (x,y): the coords
// always (FR-001) plus a deterministic feature description (may be empty).
// Exported so the mind side (convo.go) situates conversation memories from the
// same helper the executor uses. Never nil — coords alone satisfy FR-001.
func PlaceAt(s *State, x, y int) *MemoryPlace {
	return &MemoryPlace{X: x, Y: y, Desc: describePlace(s, x, y)}
}

// placeForBuild situates a build-completion memory: the tile described as it was
// without the just-built structure kind, so a fire built by the woods reads
// "at the woods (x,y)", never "at the fire (x,y)" (T024).
func placeForBuild(s *State, x, y int, builtKind string) *MemoryPlace {
	return &MemoryPlace{X: x, Y: y, Desc: describePlaceExcept(s, x, y, builtKind)}
}

// situateText composes a situated memory text from a base template and its
// context, in the exact grammar order pinned by contracts/memory-context.md:
//
//	<base>[ at <desc> (x,y) | at (x,y)][ — <why>]
//
// The where-clause splices before the base's trailing period (preserved when
// there is no why); the why-clause is the intent reason verbatim, carrying its
// own terminal punctuation. Absent parts produce no clause — never a fabricated
// one. Implemented once here so every call site composes identically.
func situateText(base string, where *MemoryPlace, why string) string {
	stem := strings.TrimSuffix(base, ".")
	hadDot := stem != base
	var b strings.Builder
	b.WriteString(stem)
	if where != nil {
		if where.Desc != "" {
			fmt.Fprintf(&b, " at %s (%d,%d)", where.Desc, where.X, where.Y)
		} else {
			fmt.Fprintf(&b, " at (%d,%d)", where.X, where.Y)
		}
	}
	switch {
	case why != "":
		b.WriteString(" — ")
		b.WriteString(why)
	case hadDot:
		b.WriteByte('.')
	}
	return b.String()
}

// situatedMemoryEvent is memoryEvent with situated context (spec 019): the
// where/why are baked into the payload AND composed into the text. Where is the
// acting agent's tile; Why is the driving intent's reason ("" for reflex).
// origin (spec 030) is the emission-stamped provenance class — a required
// parameter so the compiler forces every emission site to declare it (a new
// unstamped site cannot compile).
func situatedMemoryEvent(tick int64, agent, salience int, where *MemoryPlace, why, origin, format string, args ...any) store.Event {
	return store.Event{
		Tick: tick, Type: "agent.memory_added",
		Payload: mustPayload(MemoryAddedPayload{
			Agent: agent, Text: situateText(fmt.Sprintf(format, args...), where, why),
			Salience: salience, Subject: -1, Where: where, Why: why, Origin: origin,
		}),
	}
}

// situatedMemoryToned is memoryEventToned with situated context (spec 019);
// origin is the spec-030 provenance class (see situatedMemoryEvent).
func situatedMemoryToned(tick int64, agent, salience, tone int, where *MemoryPlace, why, origin, format string, args ...any) store.Event {
	return store.Event{
		Tick: tick, Type: "agent.memory_added",
		Payload: mustPayload(MemoryAddedPayload{
			Agent: agent, Text: situateText(fmt.Sprintf(format, args...), where, why),
			Salience: salience, Subject: -1, Tone: tone, Where: where, Why: why, Origin: origin,
		}),
	}
}

// situatedMemoryAboutEvent is memoryAboutEvent with situated context (spec 019):
// a gossip-worthy memory about another agent, situated by the WITNESS's own
// location. Witness memories carry no Why — the witness did not drive the act
// (contracts/memory-context.md rule 2). origin is the spec-030 provenance class:
// OriginWitness for a seen event, OriginReport for a learned-at-a-distance one.
func situatedMemoryAboutEvent(tick int64, agent, subject, tone, salience int, where *MemoryPlace, origin, format string, args ...any) store.Event {
	return store.Event{
		Tick: tick, Type: "agent.memory_added",
		Payload: mustPayload(MemoryAddedPayload{
			Agent: agent, Text: situateText(fmt.Sprintf(format, args...), where, ""),
			Salience: salience, Subject: subject, Tone: tone, Where: where, Origin: origin,
		}),
	}
}

// Salience table (1..10). Kept small and legible on purpose — consolidation
// (TASK-9) is the layer that reweighs and rewrites.
const (
	salTalk           = 3
	salHunt           = 4
	salFire           = 5
	salShelter        = 6
	salStarvingForage = 5
	salColdNight      = 5
	salNearDeath      = 9
	salWitnessDeath   = 10
	// SalDream: Guardian's dreams/omens (TASK-12) — exported for the
	// injection builder; between shelter and near-death so the divine
	// reliably surfaces without outranking real trauma.
	SalDream = 8
	// Governance (TASK-13): speaking is routine, outcomes matter, watching a
	// neighbor break the law sticks, being cast out is formative.
	salMeetingSpoke   = 3
	salMeetingOutcome = 5
	salNormViolation  = 6
	salExiled         = 9
	// Spec 012 (crafting economy, research R8): "high" here means memorable,
	// not generation-interrupting — both sit below GenerationBumpSalience (9),
	// the same band as SalDream, so a broken spear or a new oven doesn't
	// outrank real trauma at cognition-landing time.
	salSpearBroke = 8 // US3: the spear that spent its last use
	salAxeBroke   = 8 // spec 032 US2: the axe that spent its last harvest use (spear-broke band)
	salBath       = 5 // US4: medium, positive tone
	salOvenBuilt  = 7 // US4: high, village-visible (builder + nearby witnesses)
	// Spec 013 (inventory & storage, research R5). Same "high = memorable, not
	// generation-interrupting" band as salOvenBuilt (below GenerationBumpSalience).
	salChestBuilt = 7 // US4/T030: high, village-visible (oven precedent)
	// salTaking: a taking from an owned chest — suffered by the owner and
	// witnessed by neighbors; high and negative, above rumorMinSalience so the
	// owner's subject-tagged memory is a live gossip seed (FR-012).
	salTaking = 7
	// salFireOut: low-salience — a fire going cold nearby is background
	// texture, not formative (contracts/events.md: "fire burned out while
	// agents nearby, low"). Purely personal (no gossip subject).
	salFireOut = 3
	// salChop / salQuarry (spec 081): the actor's first-person memory of a
	// completed chop/quarry, the salHunt band (memorable, below every
	// generation-interrupting and rumor-seed threshold). The operator decision
	// 2026-07-26 supersedes the earlier "completed chops mint no memory"
	// spam-avoidance posture for these two acts — a villager's own harvest is
	// now remembered in the first person instead of being "discovered" later as
	// unexplained loss by the perception sweep.
	salChop   = 4
	salQuarry = 4
	// salMapCorrected (spec 041 US3): discovering a remembered place gone —
	// formative enough to enter the working-memory window and reshape plans,
	// well below the generation-interrupting band (the absorb trigger, not
	// salience, does the re-arming).
	salMapCorrected = 5
	// salPlaceTold (spec 041 US5): giving/getting directions — the talk band
	// (salTalk), social texture rather than a formative moment.
	salPlaceTold = 3
)

// placeToldText renders the two sides of a place-knowledge exchange (spec 041
// US5, data-model: "Told Birch about the fire by the rock." / "Birch told you
// of a fire at (x,y)."). Voiced by the FIRST fact in the payload's canonical
// order, a second fact folding into "and another place" — one memory per
// side, never one per fact (the talk band stays quiet).
func placeToldText(other string, facts []PlaceFact, asTeller bool) string {
	f := facts[0]
	what := strings.ReplaceAll(f.Kind, "_", " ")
	more := ""
	if len(facts) > 1 {
		more = ", and another place"
	}
	if asTeller {
		return fmt.Sprintf("Told %s about the %s at (%d,%d)%s.", other, what, f.X, f.Y, more)
	}
	return fmt.Sprintf("%s told you of a %s at (%d,%d)%s.", other, what, f.X, f.Y, more)
}

// chopMemoryText / quarryMemoryText render the actor's first-person memory of a
// completed harvest (spec 081) — the act-time counterpart to mapCorrectedText's
// loss-discovery voice. Format strings consumed by the executor's chop/quarry
// emit sites (the hunt-memory shape), situated by the actor's stand tile.
const (
	chopMemoryText   = "Felled the tree at (%d,%d)."
	quarryMemoryText = "Quarried the outcrop at (%d,%d)."
)

// mapCorrectedText renders the situated first-person discovery of a remembered
// place found gone (spec 041 US3, data-model: "The fire … was cold and dead
// when you looked."). Kind-specific voice for the narratively loud kinds; a
// closed generic for the rest. Baked into the reduced Memory by the
// agent.map_corrected arm — pure function of the remembered fact, so live and
// replay agree.
func mapCorrectedText(f PlaceFact) string {
	switch f.Kind {
	case "fire":
		return fmt.Sprintf("The fire at (%d,%d) was cold and dead when you looked.", f.X, f.Y)
	case "pile":
		return fmt.Sprintf("The goods at (%d,%d) were gone when you looked.", f.X, f.Y)
	case "tree":
		return fmt.Sprintf("The tree at (%d,%d) had been felled when you looked.", f.X, f.Y)
	case "rock":
		return fmt.Sprintf("The outcrop at (%d,%d) was quarried bare when you looked.", f.X, f.Y)
	}
	return fmt.Sprintf("The %s at (%d,%d) was gone when you looked.", strings.ReplaceAll(f.Kind, "_", " "), f.X, f.Y)
}

// Tone constants for the spec 012 memories above (governance.go/social.go/
// gru.go each declare their own tone band the same way).
const (
	toneBath       = 40 // positive, matching toneSaved's magnitude
	toneOvenBuilt  = 30 // positive; witnesses take pride, less personal than bathing
	toneChestBuilt = 20 // positive; a neighbor's larder is welcome but modestly so
)

// WindowK is the working-memory bound: prompts never carry more than this
// many memories (top K−tail by score + seeded tail picks).
const (
	WindowK        = 10
	windowTailPick = 2
	// recency half-life: a memory's weight halves every game-day.
	halfLifeTicks = 24 * 3600
)

// SelectedMemory pairs a selected window memory with the provenance the spec
// 043 context assembler needs to budget it: whether the entry is one of the two
// serendipity tail picks (drop priority 2) rather than a deterministic scored
// pick. The window ORDER (reverse-chronological) and MEMBERSHIP are exactly
// those of SelectMemories/SelectMemoriesRelevant — the flag is purely additive,
// so stripping it (StripSelected) reproduces those selectors byte-for-byte.
// This is proved live by TestSelectedWindowMatchesLegacy: the assembler's
// floor/serendipity split cannot drift the spec-042 selection out from under it.
type SelectedMemory struct {
	Memory
	Serendipity bool
}

// selectWindow is the shared selector behind both public windows. It runs the
// deterministic top-K algorithm ONCE and annotates each entry as a scored pick
// or a serendipity tail pick; the two public functions are thin wrappers that
// strip the flag. The ONLY difference between the legacy and relevance-blended
// windows is the per-memory score (branched on query == nil below); everything
// else — the n≤k reverse-chronological passthrough, the tie-break, the seeded
// serendipity tail (same "serendipity" rng key + cadence bucket), the final
// reverse-chronological sort and K cap — is shared, so the two windows stay
// byte-identical to their pre-043 selves by construction (spec 042 tests +
// TestSelectedWindowMatchesLegacy gate it). Selection mutates nothing (FR-004)
// and reads only a.Memories (FR-005 isolation).
func selectWindow(a *Agent, seed uint64, agentIdx int, tick int64, k int, query []float32, queryModel string) []SelectedMemory {
	n := len(a.Memories)
	if n == 0 || k <= 0 {
		return nil
	}
	if n <= k {
		out := make([]SelectedMemory, n)
		for i := range a.Memories {
			out[i] = SelectedMemory{Memory: a.Memories[i]}
		}
		sort.SliceStable(out, func(i, j int) bool { return out[i].Tick > out[j].Tick })
		return out
	}

	type scored struct {
		m     Memory
		score float64
		idx   int
	}
	all := make([]scored, n)
	for i, m := range a.Memories {
		age := tick - m.Tick
		if age < 0 {
			age = 0
		}
		// integer-friendly decay: halve per whole game-day of age.
		w := float64(m.Salience)
		for d := age / halfLifeTicks; d > 0; d-- {
			w /= 2
		}
		// Legacy score is the raw decayed weight (query nil); the relevance
		// window normalizes it and adds the [0,1] relevance term (spec 042,
		// contracts/relevance-scoring.md §1) — EXACTLY the two pre-043 formulas.
		score := w
		if query != nil {
			score = w/MaxSalience + relevance01(m, query, queryModel)
		}
		all[i] = scored{m: m, score: score, idx: i}
	}
	sort.SliceStable(all, func(i, j int) bool {
		if all[i].score != all[j].score {
			return all[i].score > all[j].score
		}
		return all[i].m.Tick > all[j].m.Tick // ties: newer wins
	})

	take := k - windowTailPick
	picked := map[int]bool{}
	var out []SelectedMemory
	for i := 0; i < take && i < n; i++ {
		out = append(out, SelectedMemory{Memory: all[i].m})
		picked[all[i].idx] = true
	}

	// Serendipity: seeded picks from the oldest half (by original position),
	// bucketed to the planner cadence so retries in one window agree.
	oldHalf := n / 2
	if oldHalf > 0 {
		r := rngAt(seed, "serendipity", tick/defaultPlannerCadenceTicks, agentIdx)
		for t := 0; t < windowTailPick; t++ {
			for tries := 0; tries < 8; tries++ {
				i := int(r.Uint64N(uint64(oldHalf)))
				if !picked[i] {
					picked[i] = true
					out = append(out, SelectedMemory{Memory: a.Memories[i], Serendipity: true})
					break
				}
			}
		}
	}

	sort.SliceStable(out, func(i, j int) bool { return out[i].Tick > out[j].Tick })
	if len(out) > k {
		out = out[:k]
	}
	return out
}

// StripSelected drops the annotation, returning the bare window in order. Nil in
// ⇒ nil out (matching the selectors' empty-window contract).
func StripSelected(sel []SelectedMemory) []Memory {
	if sel == nil {
		return nil
	}
	out := make([]Memory, len(sel))
	for i := range sel {
		out[i] = sel[i].Memory
	}
	return out
}

// SelectMemories is the deterministic top-K window (AC#3): score = salience
// halved per day of age; top K−2 by score, plus 2 serendipity picks from the
// oldest half seeded per cadence bucket; presented reverse-chronologically.
func SelectMemories(a *Agent, seed uint64, agentIdx int, tick int64, k int) []Memory {
	return StripSelected(selectWindow(a, seed, agentIdx, tick, k, nil, ""))
}

// --- spec 042: query-conditioned relevance selection ---

// SelectMemoriesRelevant is the three-term sibling of SelectMemories
// (contracts/relevance-scoring.md §1): per memory,
//
//	score = sal01 + rel01
//	sal01 = (salience halved per whole game-day of age) / MaxSalience  — EXACTLY
//	        today's decayed weight, normalized to [0,1]
//	rel01 = (cosine(m.Vec, query) + 1) / 2 when the memory carries a vector
//	        produced by queryModel; 0.5 (neutral) otherwise — vectorless and
//	        cross-model memories are neither promoted nor punished (FR-009/FR-010)
//
// Everything around the score is byte-for-byte SelectMemories: a nil query (no
// situation vector recorded yet — the legacy fallback) selects the raw-weight
// window; n ≤ k returns everything reverse-chronologically; ties break
// newer-first; the two serendipity tail picks run the shared algorithm (same
// "serendipity" rng stream key, same cadence bucket). Selection is a pure
// function of recorded data — it mutates nothing (FR-004), and its only memory
// source is a.Memories (FR-005 isolation by construction).
func SelectMemoriesRelevant(a *Agent, seed uint64, agentIdx int, tick int64, k int, query []float32, queryModel string) []Memory {
	return StripSelected(selectWindow(a, seed, agentIdx, tick, k, query, queryModel))
}

// SelectMemoriesWindow is the annotated form of the relevance window (spec 043
// US4): the same entries SelectMemoriesRelevant returns, each flagged as a
// serendipity tail pick or a scored pick, for the context assembler's
// floor/serendipity drop accounting. A nil query selects the legacy window
// (SelectMemories) — the degraded path when no situation vector is recorded.
func SelectMemoriesWindow(a *Agent, seed uint64, agentIdx int, tick int64, k int, query []float32, queryModel string) []SelectedMemory {
	return selectWindow(a, seed, agentIdx, tick, k, query, queryModel)
}

// relevance01 is the relevance term in [0,1]: (cosine + 1) / 2 when the memory
// carries a vector comparable to the query — same producing model (the FR-009
// cross-model guard), same dimensionality — else the 0.5 neutral midpoint.
// Cosine accumulates sequentially in float64 over the float32 inputs in fixed
// index order, so the value is deterministic on every platform (contract §1);
// a zero-magnitude vector is incomparable, not infinitely similar → neutral.
func relevance01(m Memory, query []float32, queryModel string) float64 {
	if len(m.Vec) == 0 || m.VecModel != queryModel || len(m.Vec) != len(query) {
		return 0.5
	}
	var dot, mm, qq float64
	for i := range query {
		mv, qv := float64(m.Vec[i]), float64(query[i])
		dot += mv * qv
		mm += mv * mv
		qq += qv * qv
	}
	if mm == 0 || qq == 0 {
		return 0.5
	}
	return (dot/(math.Sqrt(mm)*math.Sqrt(qq)) + 1) / 2
}

// FormatMemory renders one memory line as prompts and soul.md show it.
func FormatMemory(m Memory) string {
	return fmt.Sprintf("%s (%d★) %s", clock.Format(m.Tick), m.Salience, m.Text)
}

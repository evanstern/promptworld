package mind

import (
	"fmt"
	"sort"
	"strings"

	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/sim"
)

// Per-turn context assembly (spec 043, contracts/context-blocks.md). The
// villager decision prompt is built from named blocks in a fixed contract
// order; each block is a pure function of world state (determinism is
// doctrine — no wall clock, no RNG outside the seeded serendipity picks the
// memory selector already owns). Every block is measured; when the assembled
// context exceeds the size budget, whole blocks are dropped lowest-priority
// first, and the sizes + drops are recorded per thought (telemetry.go) so
// operators can see exactly what the model was given.
//
// The block registry (contract order, top to bottom) and the drop-priority
// column are the normative contract; contracts/context-blocks.md and the
// living projection docs/wiki/decision-context.md must move together with it.

// contextBudgetTokens is the per-thought assembled-context ceiling in
// approx-tokens (bytes/4). It is a package const today with a sensible default;
// the design intends it as a per-world tuning-manifest dial (TASK-107's
// const-fallback pattern — the manifest supplies the value when present, this
// const is the fallback). Thoughts run only a handful of reasoning turns, so a
// moderate budget is affordable on locally hostable tiers; the budget is a
// guardrail, not a billing meter (research R6).
const contextBudgetTokens = 2000

// neverDrop marks a survival block that is never shed under budget pressure
// (frame, needs, self_history, inventory — contracts/context-blocks.md). It is
// a Priority sentinel far above any droppable block's rank, so "higher =
// dropped later" holds uniformly and survival blocks sort last by construction.
const neverDrop = 1 << 30

// promptCloser is the fixed decision question every prompt ends with. It is
// not a block — it is never measured for the budget and never dropped.
const promptCloser = "\nWhat do you do next?"

// approxTokens is the shipped budget heuristic: no tokenizer exists in
// production (research R6), and bytes/4 is stable across the local tiers in
// use. Promoted from the test-only tokensApprox in prompt_test.go.
func approxTokens(bytes int) int { return bytes / 4 }

// contextBlock is one unit of assembled context: a contract name, a drop
// priority (higher = dropped later; neverDrop = survival), and a pure renderer
// that returns "" when the block has nothing to say (absent — no empty header).
type contextBlock struct {
	Name     string
	Priority int
	Render   func() string
}

// assembled is the result of one context assembly: the final prompt text plus
// the per-thought observability the telemetry stamps onto cog.thought.
type assembled struct {
	text          string
	promptBytes   int
	blockBytes    map[string]int
	droppedBlocks []string
}

// fixedBlocks builds the whole-block registry for one agent — contract blocks
// 1-7 in fixed order (contracts/context-blocks.md). The memory window (blocks 8
// `memories` + 9 `memories_serendipity`) and the journal (block 10) are NOT
// here: the memory chunk drops per-tier (floor never, above-floor + serendipity
// by priority) rather than whole, and the journal renders after the memory
// chunk, so both are assembled specially in assembleBudget. The futureLine is
// the FR-016 future-dating line, part of the frame block.
func fixedBlocks(s *sim.State, idx int, futureLine string) []contextBlock {
	return []contextBlock{
		{Name: "frame", Priority: neverDrop, Render: func() string { return renderFrame(s, idx, futureLine) }},
		{Name: "needs", Priority: neverDrop, Render: func() string { return renderNeeds(s, idx) }},
		{Name: "self_history", Priority: neverDrop, Render: func() string { return renderSelfHistory(s, idx) }},
		{Name: "inventory", Priority: neverDrop, Render: func() string { return renderInventory(s, idx) }},
		{Name: "plan_echo", Priority: 6, Render: func() string { return renderPlanEcho(s, idx) }},
		{Name: "known_places", Priority: 5, Render: func() string { return renderKnownPlaces(s, idx) }},
		{Name: "social_law", Priority: 4, Render: func() string { return renderSocialLaw(s, idx) }},
	}
}

// assembleContext renders the decision prompt under the default budget.
func assembleContext(s *sim.State, idx, k int, mode, futureLine string) assembled {
	return assembleBudget(s, idx, k, mode, futureLine, contextBudgetTokens)
}

// assembleBudget renders in fixed contract order, measures each block, and —
// while the total (kept block bytes + the fixed closer) exceeds the budget in
// approx-tokens — sheds content lowest-priority-first, recording each drop in
// order. Drop candidates and their priorities (contracts/context-blocks.md,
// research R7):
//
//	journal (1) → memories_serendipity (2) → memories above-floor (3) →
//	social_law (4) → known_places (5) → plan_echo (6) → [never] frame, needs,
//	self_history, inventory, and the protected floor of the memories block.
//
// The memory window is one contiguous rendered chunk (blocks 8-9) so nothing is
// reordered — splitting `memories`/`memories_serendipity` is an accounting
// change, not a wording one: with no drops the chunk is byte-identical to the
// pre-043 single "You remember:" list. Under pressure the serendipity tail and
// then the above-floor entries are shed by re-rendering the kept lines in place;
// the floor of memoryFloor entries is never dropped. Deterministic: identical
// world state ⇒ identical bytes, identical drops. The budget is a parameter so
// tests can shrink it; production passes contextBudgetTokens.
func assembleBudget(s *sim.State, idx, k int, mode, futureLine string, budget int) assembled {
	type rendered struct {
		name     string
		priority int
		text     string
		kept     bool
	}

	// Blocks 1-7 (whole-block droppable).
	var fixed []rendered
	for _, b := range fixedBlocks(s, idx, futureLine) {
		if t := b.Render(); t != "" {
			fixed = append(fixed, rendered{name: b.Name, priority: b.Priority, text: t, kept: true})
		}
	}

	// Memory chunk (blocks 8-9): the annotated window split into floor /
	// above-floor / serendipity line groups.
	memLines := buildMemLines(s, idx, k, mode)
	hasAbove, hasSeren := false, false
	for _, l := range memLines {
		switch {
		case l.serendipity:
			hasSeren = true
		case !l.floor:
			hasAbove = true
		}
	}
	dropSeren, dropAbove := false, false

	// Journal (block 10): whole-block, first dropped (priority 1).
	journalText := renderJournal(s, idx)
	journalKept := journalText != ""

	total := func() int {
		n := len(promptCloser) + len(renderMemLines(memLines, dropSeren, dropAbove))
		for _, r := range fixed {
			if r.kept {
				n += len(r.text)
			}
		}
		if journalKept {
			n += len(journalText)
		}
		return n
	}

	var dropped []string
	for approxTokens(total()) > budget {
		// Lowest-priority droppable candidate goes first. Protected content
		// (neverDrop fixed blocks + the memory floor) is never a candidate; when
		// only protected content remains the budget cannot be met and we stop
		// (research R7 — the contract protects it).
		const (
			kindNone = iota
			kindFixed
			kindSeren
			kindAbove
			kindJournal
		)
		bestPri, bestKind, bestFixed := neverDrop, kindNone, -1
		consider := func(pri, kind, fi int) {
			if pri < bestPri {
				bestPri, bestKind, bestFixed = pri, kind, fi
			}
		}
		for i := range fixed {
			if fixed[i].kept && fixed[i].priority != neverDrop {
				consider(fixed[i].priority, kindFixed, i)
			}
		}
		if journalKept {
			consider(1, kindJournal, -1)
		}
		if hasSeren && !dropSeren {
			consider(2, kindSeren, -1)
		}
		if hasAbove && !dropAbove {
			consider(3, kindAbove, -1)
		}
		switch bestKind {
		case kindFixed:
			fixed[bestFixed].kept = false
			dropped = append(dropped, fixed[bestFixed].name)
		case kindJournal:
			journalKept = false
			dropped = append(dropped, "journal")
		case kindSeren:
			dropSeren = true
			dropped = append(dropped, "memories_serendipity")
		case kindAbove:
			// Partial drop of the memories block: the above-floor entries are
			// shed, the floor survives, so "memories" stays present in
			// blockBytes (floor bytes) while appearing in DroppedBlocks to mark
			// the trim (contract block 8: floor never dropped).
			dropAbove = true
			dropped = append(dropped, "memories")
		default:
			// Nothing droppable left; protected content overflows the budget.
			goto assemble
		}
	}

assemble:
	var b strings.Builder
	blockBytes := make(map[string]int)
	for _, r := range fixed {
		if r.kept {
			b.WriteString(r.text)
			blockBytes[r.name] = len(r.text)
		}
	}
	b.WriteString(renderMemLines(memLines, dropSeren, dropAbove))
	if memBytes, serenBytes := memAccount(memLines, dropSeren, dropAbove); memBytes > 0 || serenBytes > 0 {
		if memBytes > 0 {
			blockBytes["memories"] = memBytes
		}
		if serenBytes > 0 {
			blockBytes["memories_serendipity"] = serenBytes
		}
	}
	if journalKept {
		b.WriteString(journalText)
		blockBytes["journal"] = len(journalText)
	}
	b.WriteString(promptCloser)
	text := b.String()

	return assembled{
		text:          text,
		promptBytes:   len(text),
		blockBytes:    blockBytes,
		droppedBlocks: dropped,
	}
}

// --- block renderers (contract order) --------------------------------------

// renderFrame is contract block 1: the future-dating line (FR-016, empty at
// uncapped test speeds), then time, phase, and position. Never empty.
func renderFrame(s *sim.State, idx int, futureLine string) string {
	a := s.Agents[idx]
	phase := "daytime"
	if s.Night {
		phase = "night"
	}
	return futureLine + fmt.Sprintf("It is %s (%s). You are at (%d, %d).\n", clock.Format(s.Tick), phase, a.X, a.Y)
}

// trajectoryDeadband is the movement (on the raw 0-1000 needs scale) a need must
// clear from its window-edge anchor before it reads rising/falling; smaller
// swings render steady (spec 043 US2, FR-004). ±10 of 1000 is one point on the
// displayed 0-100 scale — enough to swallow game-minute decay jitter so a need
// that has not meaningfully moved never flickers direction (SC-003).
const trajectoryDeadband = 10

// trajectory reports a need's direction from its window-edge anchor. Before an
// anchor exists (first window, edge case 1) the direction is steady — the model
// must not read a spurious rising/falling from a missing history window.
func trajectory(current, anchor int, hasAnchor bool) string {
	if !hasAnchor {
		return "steady"
	}
	switch d := current - anchor; {
	case d > trajectoryDeadband:
		return "rising"
	case d < -trajectoryDeadband:
		return "falling"
	default:
		return "steady"
	}
}

// renderNeeds is contract block 2: the five needs (0-100 scale), each carrying a
// trajectory direction (spec 043 US2) derived from the reducer-maintained
// window-edge anchor (NeedsAnchor/NeedsAnchorTick, agents.go) — "warmth 45 and
// falling" reads differently from "warmth 45 and rising" (FR-004). Direction is
// sign(current − anchor) with a deadband so a steady need never flickers (SC-003);
// a nil anchor (first window) renders every need steady. Never empty, never
// dropped.
func renderNeeds(s *sim.State, idx int) string {
	a := s.Agents[idx]
	has := a.NeedsAnchor != nil
	var an sim.Needs
	if has {
		an = *a.NeedsAnchor
	}
	return fmt.Sprintf("Needs (0-100): health %d and %s, food %d and %s, rest %d and %s, warmth %d and %s, morale %d and %s.\n",
		a.Needs.Health/10, trajectory(a.Needs.Health, an.Health, has),
		a.Needs.Food/10, trajectory(a.Needs.Food, an.Food, has),
		a.Needs.Rest/10, trajectory(a.Needs.Rest, an.Rest, has),
		a.Needs.Warmth/10, trajectory(a.Needs.Warmth, an.Warmth, has),
		a.Needs.Morale/10, trajectory(a.Needs.Morale, an.Morale, has))
}

// renderInventory is contract block 4: the full carried resource/item set.
func renderInventory(s *sim.State, idx int) string {
	a := s.Agents[idx]
	var b strings.Builder
	fmt.Fprintf(&b, "Carrying: %d wood, %d stone, %d water, %d planks, %d refined stone, food (%d raw, %d cooked, %d meals)",
		a.Inv.Wood, a.Inv.Stone, a.Inv.Water, a.Inv.Planks, a.Inv.RefinedStone,
		a.Inv.FoodRaw, a.Inv.FoodCooked, a.Inv.Meals)
	if n := len(a.Inv.Spears); n > 0 {
		fmt.Fprintf(&b, ", %d spear(s) (%d uses left on the most-worn)", n, a.Inv.Spears[0])
	}
	b.WriteString(".\n")
	return b.String()
}

// renderPlanEcho is contract block 5 (spec 043 US3, FR-005): the villager's
// active plan echoed back so a thought taken mid-plan can knowingly continue,
// revise, or abandon it instead of being unaware a plan exists. It lists the
// remaining guarded steps in execution order — the head marked "next", the rest
// "then" — each with its guard (When) and validity deadline (Until) rendered in
// plain words. With no active plan the block is omitted entirely (returns "",
// no stale echo — the end of a plan is instead surfaced by self_history via the
// US1 ring). Drop priority 6 (shed before known_places under budget pressure).
func renderPlanEcho(s *sim.State, idx int) string {
	a := s.Agents[idx]
	if len(a.Plan) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("Your plan (remaining steps, in order):\n")
	for i, st := range a.Plan {
		lead := "then"
		if i == 0 {
			lead = "next"
		}
		fmt.Fprintf(&b, "- %s: %s", lead, st.Goal)
		if st.When != nil {
			fmt.Fprintf(&b, " (%s)", guardPhrase(s, *st.When))
		}
		if st.Until > 0 {
			fmt.Fprintf(&b, ", valid until %s", clock.Format(st.Until))
		}
		b.WriteString("\n")
	}
	return b.String()
}

// guardPhrase renders one plan-step guard (the closed vocabulary, guard.go) in
// plain second-person words — never the raw predicate name. Target guards name
// the villager when the index is in range; timed guards render the boundary on
// the game clock.
func guardPhrase(s *sim.State, g sim.Guard) string {
	target := func() string {
		if g.Target >= 0 && g.Target < len(s.Agents) {
			return s.Agents[g.Target].Name
		}
		return "the target"
	}
	switch g.Type {
	case sim.GuardTargetAlive:
		return "while " + target() + " is alive"
	case sim.GuardTargetPresent:
		return "while " + target() + " is still nearby"
	case sim.GuardNotSuperseded:
		return "unless something more urgent interrupts"
	case sim.GuardAfterTick:
		return "not before " + clock.Format(g.Tick)
	case sim.GuardBeforeTick:
		return "only before " + clock.Format(g.Tick)
	default:
		return "when its condition holds"
	}
}

// renderKnownPlaces is contract block 6: the spec-041 known-places section plus
// the peer-sighting "Nearby" line, unchanged content.
func renderKnownPlaces(s *sim.State, idx int) string {
	a := s.Agents[idx]
	var b strings.Builder
	b.WriteString(knownPlaces(s, idx))

	var nearby []string
	if a.Map != nil {
		for _, p := range a.Map.Peers {
			if p.Agent == idx || p.Agent < 0 || p.Agent >= len(s.Agents) {
				continue
			}
			o := s.Agents[p.Agent]
			if o.Dead {
				continue
			}
			d := absInt(p.X-a.X) + absInt(p.Y-a.Y)
			if d > 10 {
				continue
			}
			state := ""
			if o.Asleep && o.X == p.X && o.Y == p.Y {
				state = ", asleep"
			}
			nearby = append(nearby, fmt.Sprintf("%s (%d tiles away%s)", o.Name, d, state))
		}
	}
	if len(nearby) > 0 {
		fmt.Fprintf(&b, "Nearby: %s.\n", strings.Join(nearby, ", "))
	}
	return b.String()
}

// renderSocialLaw is contract block 7: bonds/debts/reputation/rumor followed by
// the village-law context, unchanged content.
func renderSocialLaw(s *sim.State, idx int) string {
	return socialContext(s, idx) + villageLaw(s, idx)
}

// memHeader opens the working-memory list. It belongs to the `memories` block
// (accounted there) and is written whenever any memory line survives.
const memHeader = "\nYou remember:\n"

// memoryFloor is the protected count of working-memory entries in the `memories`
// block (spec 043 US4, contracts/context-blocks.md block 8): the most-recent
// this-many scored (non-serendipity) entries are never dropped under budget
// pressure, so a survival-relevant recollection always reaches the model.
// Entries above the floor shed at drop priority 3; the serendipity tail at 2.
const memoryFloor = 4

// memLine is one working-memory entry as it renders, tagged with the drop tier
// it belongs to: a serendipity tail pick (block 9, priority 2), a protected
// floor entry (block 8, never dropped), or an above-floor entry (block 8,
// priority 3). The lines stay in the window's reverse-chronological order, so
// with nothing dropped the concatenation is byte-identical to the pre-043
// single "You remember:" list — the floor/serendipity split is accounting, not
// reordering.
type memLine struct {
	text        string
	serendipity bool
	floor       bool
}

// buildMemLines renders contract blocks 8-9 into ordered, tiered lines. The
// window is the annotated selection (spec 042 relevance-blended when
// memory_relevance is "on" with a recorded situation vector, legacy otherwise);
// the first memoryFloor non-serendipity entries in render order are the
// protected floor. Nil (no memories) → no lines, so no header ever renders.
func buildMemLines(s *sim.State, idx, k int, mode string) []memLine {
	window := selectWindowAnnotated(s, idx, k, s.Tick, mode)
	if len(window) == 0 {
		return nil
	}
	lines := make([]memLine, len(window))
	floorCount := 0
	for i, sm := range window {
		l := memLine{text: fmt.Sprintf("- %s\n", sim.FormatMemory(sm.Memory)), serendipity: sm.Serendipity}
		if !sm.Serendipity && floorCount < memoryFloor {
			l.floor = true
			floorCount++
		}
		lines[i] = l
	}
	return lines
}

// renderMemLines writes the working-memory chunk (blocks 8-9) with the dropped
// tiers omitted: serendipity picks when dropSeren, above-floor entries when
// dropAbove. Order is preserved (reverse-chronological), so with no drops the
// output equals the pre-043 rendering exactly. Empty when every line is dropped
// or there are no memories (no bare header).
func renderMemLines(lines []memLine, dropSeren, dropAbove bool) string {
	var body strings.Builder
	for _, l := range lines {
		if l.serendipity {
			if dropSeren {
				continue
			}
		} else if !l.floor && dropAbove {
			continue
		}
		body.WriteString(l.text)
	}
	if body.Len() == 0 {
		return ""
	}
	return memHeader + body.String()
}

// memAccount splits the kept memory-chunk bytes into the two contract block
// names: `memories` (the header + kept floor/above-floor lines) and
// `memories_serendipity` (kept serendipity lines). The header is attributed to
// `memories` — the floor is never dropped, so whenever any memory renders the
// header rides with it. memBytes + serenBytes == len(renderMemLines(...)), so
// the per-block telemetry accounts for every assembled byte.
func memAccount(lines []memLine, dropSeren, dropAbove bool) (memBytes, serenBytes int) {
	for _, l := range lines {
		switch {
		case l.serendipity:
			if !dropSeren {
				serenBytes += len(l.text)
			}
		case l.floor || !dropAbove:
			memBytes += len(l.text)
		}
	}
	if memBytes > 0 || serenBytes > 0 {
		memBytes += len(memHeader)
	}
	return memBytes, serenBytes
}

// situationTerms is the deterministic query-term set the journal block matches
// against — the same situation signals renderSituation embeds (embedder.go):
// the two worst needs' names and the active-or-last intent goal (research R5).
// Pure over agent state; identical state ⇒ identical terms.
func situationTerms(a sim.Agent) []string {
	needs := []struct {
		name string
		v    int
	}{
		{"health", a.Needs.Health}, {"food", a.Needs.Food}, {"rest", a.Needs.Rest},
		{"warmth", a.Needs.Warmth}, {"morale", a.Needs.Morale},
	}
	sort.SliceStable(needs, func(i, j int) bool { return needs[i].v < needs[j].v })
	terms := []string{needs[0].name, needs[1].name}
	goal := ""
	if a.Intent != nil {
		goal = a.Intent.Goal
	} else if a.LastGoal != "" {
		goal = a.LastGoal
	}
	if goal != "" {
		terms = append(terms, goal)
	}
	return terms
}

// renderJournal is contract block 10 (spec 043 US4, FR-007): up to
// JournalExcerptCap term-matched journal entries, each a ≤JournalExcerptRunes
// excerpt with its entry id, selected deterministically for the current
// situation (situationTerms). The assembler stuffs them so the villager need
// not spend reasoning turns fetching its own journal; no match → omitted
// entirely. Drop priority 1 (first shed under budget pressure). Model-free, so
// always available even in degraded mode.
func renderJournal(s *sim.State, idx int) string {
	a := s.Agents[idx]
	if a.Journal == nil {
		return ""
	}
	ex := a.Journal.SelectJournalExcerpts(situationTerms(a))
	if len(ex) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\nFrom your journal:\n")
	for _, e := range ex {
		fmt.Fprintf(&b, "- (#%d) %s\n", e.ID, e.Text)
	}
	return b.String()
}

// renderSelfHistory is contract block 3 (spec 043 US1): the villager's own
// recent intents — the last few IntentLog records, newest first, each naming
// the goal, its source in plain words, and its outcome. Sources are mapped
// honestly (contract rule "Sources named honestly"): planner → "you decided",
// reflex → "instinct", plan → "your plan"; any other/unknown source renders as
// "unknown" and NO reason is ever invented (reflex records carry none). On a
// villager's very first thought the block renders an explicit "no prior
// activity" line rather than vanishing — the model must always be able to tell
// "I have not acted" from silence (edge case 1, AS-3). Never dropped.
func renderSelfHistory(s *sim.State, idx int) string {
	a := s.Agents[idx]
	if len(a.IntentLog) == 0 {
		return "Recently: no prior activity yet — this is your first decision.\n"
	}
	const show = 4
	var b strings.Builder
	b.WriteString("Recently you:\n")
	shown := 0
	for i := len(a.IntentLog) - 1; i >= 0 && shown < show; i-- {
		r := a.IntentLog[i]
		b.WriteString("- " + selfHistoryLine(r) + "\n")
		shown++
	}
	return b.String()
}

// selfHistoryLine renders one IntentRecord in plain second-person words. The
// clauses are: what was pursued, where it came from, the stated reason (only
// when one was recorded — never fabricated), and how it ended.
func selfHistoryLine(r sim.IntentRecord) string {
	source := ""
	switch r.Source {
	case "planner":
		source = "you decided this"
	case "reflex":
		source = "instinct drove this"
	case "plan":
		source = "your plan's step"
	default:
		source = "source unknown"
	}
	line := r.Goal + " — " + source
	// Reason is present only for planner/plan intents; reflex records carry
	// none and none is invented (contract: instinct honesty).
	if r.Reason != "" {
		line += fmt.Sprintf(" (%q)", r.Reason)
	}
	switch r.Outcome {
	case "":
		line += "; still underway"
	case "done":
		line += "; completed"
	case "failed":
		line += "; it failed"
	case "rejected":
		line += "; rejected before it began"
	case "expired":
		line += "; the plan expired"
	default:
		line += "; " + r.Outcome
	}
	return line
}

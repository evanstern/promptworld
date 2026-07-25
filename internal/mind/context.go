package mind

import (
	"fmt"
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

// contextBlocks builds the registry for one agent in fixed contract order
// (contracts/context-blocks.md). Blocks not yet implemented by a landed slice
// (plan_echo, memories_serendipity, journal) are simply absent from the
// registry; they insert at their contract position when their story lands. The
// futureLine is the FR-016 future-dating line, part of the frame block.
func contextBlocks(s *sim.State, idx, k int, mode, futureLine string) []contextBlock {
	return []contextBlock{
		{Name: "frame", Priority: neverDrop, Render: func() string { return renderFrame(s, idx, futureLine) }},
		{Name: "needs", Priority: neverDrop, Render: func() string { return renderNeeds(s, idx) }},
		{Name: "self_history", Priority: neverDrop, Render: func() string { return renderSelfHistory(s, idx) }},
		{Name: "inventory", Priority: neverDrop, Render: func() string { return renderInventory(s, idx) }},
		{Name: "known_places", Priority: 5, Render: func() string { return renderKnownPlaces(s, idx) }},
		{Name: "social_law", Priority: 4, Render: func() string { return renderSocialLaw(s, idx) }},
		{Name: "memories", Priority: 3, Render: func() string { return renderMemories(s, idx, k, mode) }},
	}
}

// assembleContext renders the decision prompt under the default budget.
func assembleContext(s *sim.State, idx, k int, mode, futureLine string) assembled {
	return assembleBudget(s, idx, k, mode, futureLine, contextBudgetTokens)
}

// assembleBudget renders in fixed contract order, measures each non-empty
// block, and — while the total (block bytes + the fixed closer) exceeds the
// budget in approx-tokens — drops whole blocks lowest-priority-first, recording
// each drop in order. Deterministic: identical world state ⇒ identical bytes,
// identical drops. The budget is a parameter so tests can shrink it; production
// passes contextBudgetTokens.
func assembleBudget(s *sim.State, idx, k int, mode, futureLine string, budget int) assembled {
	blocks := contextBlocks(s, idx, k, mode, futureLine)

	type rendered struct {
		name     string
		priority int
		text     string
	}
	present := make([]rendered, 0, len(blocks))
	for _, b := range blocks {
		if t := b.Render(); t != "" {
			present = append(present, rendered{name: b.Name, priority: b.Priority, text: t})
		}
	}

	total := func(rs []rendered) int {
		n := len(promptCloser)
		for _, r := range rs {
			n += len(r.text)
		}
		return n
	}

	var dropped []string
	for approxTokens(total(present)) > budget {
		// Lowest-priority droppable block goes first. Survival blocks
		// (neverDrop) are never candidates; if only survival blocks remain we
		// stop — the budget cannot be met without shedding a protected block,
		// and the contract protects it (research R7).
		victim := -1
		for i := range present {
			if present[i].priority == neverDrop {
				continue
			}
			if victim < 0 || present[i].priority < present[victim].priority {
				victim = i
			}
		}
		if victim < 0 {
			break
		}
		dropped = append(dropped, present[victim].name)
		present = append(present[:victim], present[victim+1:]...)
	}

	var b strings.Builder
	blockBytes := make(map[string]int, len(present))
	for _, r := range present {
		b.WriteString(r.text)
		blockBytes[r.name] = len(r.text)
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

// renderNeeds is contract block 2: the five needs (0-100 scale). Trajectory
// arrows (US2) attach here later; this slice renders the level line unchanged.
func renderNeeds(s *sim.State, idx int) string {
	a := s.Agents[idx]
	return fmt.Sprintf("Needs (0-100): health %d, food %d, rest %d, warmth %d, morale %d.\n",
		a.Needs.Health/10, a.Needs.Food/10, a.Needs.Rest/10, a.Needs.Warmth/10, a.Needs.Morale/10)
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

// renderMemories is contract block 8: the selected working-memory window (spec
// 042 relevance-blended when memory_relevance is "on", legacy otherwise). The
// window is the ONLY memory content that ever reaches a prompt. This slice
// renders it as one block; US4 splits floor/serendipity and adds the journal.
func renderMemories(s *sim.State, idx, k int, mode string) string {
	window := selectWindow(s, idx, k, s.Tick, mode)
	if len(window) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\nYou remember:\n")
	for _, m := range window {
		fmt.Fprintf(&b, "- %s\n", sim.FormatMemory(m))
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

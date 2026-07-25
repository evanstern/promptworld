package mind

import (
	"fmt"
	"strings"

	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/world"
)

// Prompt construction: a stable per-agent system prefix (identity + persona +
// tool-choice framing — prompt-cache friendly) and a variable user suffix
// (situation + the bounded working-memory window).
//
// In the tool-use era (spec 017) the goal vocabulary and per-verb gloss are no
// longer hand-rendered into the prompt: each tool carries its own name, gloss
// (Description), and argument schema, declared to the model as callable tools
// (tool.LoopRosterVillager -> the loop's Request.Tools). The system prompt now
// only frames the choice; the tools carry the vocabulary. The spec-014 golden
// prompt test that pinned the free-text contract retires with this change
// (contracts/loop-api.md).

// systemPrompt renders the villager system prefix (spec 027,
// contracts/system-prompt.md) as a crafted three-part frame:
//
//  1. Identity — one statement, the ONLY place the frame names the agent
//     (FR-001); everything after speaks to "you".
//  2. Persona — the agent's authored nature verbatim, set off as its own block;
//     absent cleanly when personaText is empty (FR-002).
//  3. Task framing — the decision contract in the second person, doctrine
//     preserved (FR-003): the decision is one acting-tool call; read-only tools
//     may precede it; set_plan and muse are actions carrying an opportunity
//     cost; no free-text action path is offered.
//
// It is a pure function of (name, personaText): identical inputs render
// byte-identical output, so it stays the cacheable per-agent prefix (FR-005,
// contract C1) — it carries NO dynamic world state (that lives in userPrompt).
func systemPrompt(name, personaText string) string {
	var b strings.Builder
	// Part 1 — identity.
	fmt.Fprintf(&b, "You are %s, a villager in a small settlement.\n", name)
	// Part 2 — persona. Trailing newlines are trimmed so the block separator is
	// the frame's, not the persona's: an empty persona then vanishes with no
	// doubled blank line or dangling separator (C4).
	if p := strings.TrimRight(personaText, "\n"); p != "" {
		b.WriteString("\n")
		b.WriteString(p)
		b.WriteString("\n")
	}
	// Part 3 — task framing (doctrine-preserving, name-free).
	b.WriteString("\nEach turn you decide what you do next by calling exactly one acting tool. " +
		"Every tool is an action; its description says what it does and its arguments say what it needs. " +
		"You may first call read-only tools to look something up, then finish by calling a single acting tool — " +
		"a world action, a short plan (set_plan), or a passing thought (muse). " +
		"Musing and planning are actions too: a beat spent thinking is a beat not spent doing. " +
		"Choose the one action that best fits your situation and needs right now.\n")
	return b.String()
}

// futureDated tells the model when its decision will land (FR-016): thought
// is not instant, and the prompt stops pretending it is. Empty when there is
// no meaningful prediction (uncapped test speeds).
func futureDated(now, landing int64) string {
	if landing <= now {
		return ""
	}
	return fmt.Sprintf("It is now %s. Your decision will take effect around %s — plan for then, not for this instant.\n",
		clock.Format(now), clock.Format(landing))
}

// selectWindow picks a prompt's memory window per the world's
// memory_relevance mode (spec 042 US3, contracts/relevance-scoring.md §3):
// "on" consumes the query-conditioned SelectMemoriesRelevant against the
// agent's recorded situation vector (a nil query — no situation vector yet —
// delegates to legacy inside the selector, per contract); "" and "shadow"
// alike are today's SelectMemories, byte-identical (the shadow invariant).
// Divergence recording is separate (recordDivergence) and unchanged by mode.
func selectWindow(s *sim.State, idx, k int, tick int64, mode string) []sim.Memory {
	return sim.StripSelected(selectWindowAnnotated(s, idx, k, tick, mode))
}

// selectWindowAnnotated is selectWindow returning the window annotated with the
// serendipity tail flag (spec 043 US4), for the assembler's floor/serendipity
// drop accounting. Same mode logic as selectWindow — "on" with a recorded
// situation vector consumes the relevance window, everything else (no vector,
// "", "shadow") the legacy window — so StripSelected of this equals selectWindow
// byte-for-byte, and the shadow invariant (""/"shadow" identical) is preserved.
func selectWindowAnnotated(s *sim.State, idx, k int, tick int64, mode string) []sim.SelectedMemory {
	a := &s.Agents[idx]
	if mode == world.MemoryRelevanceOn {
		var query []float32
		if len(a.SitVec) > 0 {
			query = a.SitVec
		}
		return sim.SelectMemoriesWindow(a, s.Seed, idx, tick, k, query, a.SitVecModel)
	}
	return sim.SelectMemoriesWindow(a, s.Seed, idx, tick, k, nil, "")
}

// userPrompt renders the situation + memory window. The window is the ONLY
// memory content that ever reaches a prompt (AC#3); mode is the world's
// memory_relevance flag deciding which selector fills it (spec 042 US3).
//
// Since spec 043 the situation is assembled from named blocks in a fixed
// contract order under a size budget (context.go); userPrompt is the
// no-future-line, default-budget entry point the direct callers (tests, the
// TUI capture path) use. The planner path calls assembleContext directly so it
// can prepend the future-dating line and stamp the per-block sizes onto the
// thought's telemetry.
func userPrompt(s *sim.State, idx int, k int, mode string) string {
	return assembleContext(s, idx, k, mode, "").text
}

// AssembleUserPrompt is the exported, no-future-line entry point for tooling
// that reproduces exactly what a thought was given (replay-determinism harness,
// TUI capture) from a bare *sim.State — the same assembly the planner path runs,
// minus the per-thought future-dating line. It is a pure function of world state
// (spec 043 context.go): identical state ⇒ identical bytes, so a world recovered
// from its event log yields byte-identical prompts. mode is the world's
// memory_relevance flag (""/"shadow"/"on").
func AssembleUserPrompt(s *sim.State, idx, k int, mode string) string {
	return userPrompt(s, idx, k, mode)
}

// knownPlaces renders the spec-041 known-places section (US2, contracts §3):
// what the acting agent's mental map holds, fresh facts only (the same
// read-time horizon the resolvers use), never State.Structures.
//
//   - Landmark structures (fire/shelter/oven/chest) individually, with
//     provenance flavor — witnessed plain; told names the teller; revealed
//     names the vision. No count cap. A fire the agent remembers as burned
//     out (its remembered FuelUntil behind the clock) says so, matching the
//     resolvers' remembered-lit reads.
//   - Everything place-shaped (resource kinds, plus walls and paths — which
//     come in runs and would flood an individual listing; grouping, not
//     dropping, is the contract's own size bound) grouped per kind with
//     count + nearest.
//   - One orientation line toward the nearest unexplored land, omitted when
//     the map is fully explored.
//   - The explicit empty state ("You know of no fires or shelters yet.") when
//     no landmark structure is known — the model must always be able to tell
//     "I know none" from silence.
func knownPlaces(s *sim.State, idx int) string {
	a := s.Agents[idx]
	var b strings.Builder
	if a.Map == nil {
		b.WriteString("You know of no fires or shelters yet.\n")
		return b.String()
	}
	now := s.Tick

	// Landmark structures, individually, in the map's canonical fact order.
	landmark := map[string]bool{"fire": true, "shelter": true, "oven": true, "chest": true}
	var places []string
	for _, f := range a.Map.Facts {
		if !landmark[f.Kind] || !f.Fresh(now) {
			continue
		}
		var entry string
		switch f.Provenance {
		case "told":
			teller := "someone"
			if f.Source >= 0 && f.Source < len(s.Agents) {
				teller = s.Agents[f.Source].Name
			}
			entry = fmt.Sprintf("a %s at (%d,%d) — %s told you", f.Kind, f.X, f.Y, teller)
		case "revealed":
			entry = fmt.Sprintf("a %s at (%d,%d), shown to you in a vision", f.Kind, f.X, f.Y)
		default: // witnessed
			entry = fmt.Sprintf("%s at (%d,%d)", f.Kind, f.X, f.Y)
		}
		if f.Kind == "fire" && f.Detail <= now {
			entry += " (likely burned out by now)"
		}
		places = append(places, entry)
	}
	if len(places) > 0 {
		fmt.Fprintf(&b, "Places you know: %s.\n", strings.Join(places, "; "))
	} else {
		b.WriteString("You know of no fires or shelters yet.\n")
	}

	// Grouped place kinds: count + nearest, fixed order.
	groupKinds := []struct{ kind, one, many string }{
		{"forage", "forage spot", "forage spots"},
		{"tree", "stand of trees", "stands of trees"},
		{"rock", "rock outcrop", "rock outcrops"},
		{"water_edge", "watering spot", "watering spots"},
		{"den", "animal den", "animal dens"},
		{"pile", "pile of goods", "piles of goods"},
		{"wall_plank", "plank wall", "plank walls"},
		{"wall_stone", "stone wall", "stone walls"},
		{"path", "path tile", "path tiles"},
	}
	var groups []string
	for _, g := range groupKinds {
		fresh := a.Map.KnownFresh(g.kind, now)
		if len(fresh) == 0 {
			continue
		}
		near := fresh[0]
		for _, f := range fresh[1:] {
			if absInt(f.X-a.X)+absInt(f.Y-a.Y) < absInt(near.X-a.X)+absInt(near.Y-a.Y) {
				near = f
			}
		}
		noun := g.many
		if len(fresh) == 1 {
			noun = g.one
		}
		groups = append(groups, fmt.Sprintf("%d %s (nearest (%d,%d))", len(fresh), noun, near.X, near.Y))
	}
	if len(groups) > 0 {
		fmt.Fprintf(&b, "You know %s.\n", joinAnd(groups))
	}

	// Orientation toward the unknown; silent on a fully-explored map (or a
	// map-less test State, whose dims are 0).
	if w, h := s.MapDims(); w > 0 && h > 0 {
		if dir, ok := a.Map.FrontierDirection(w, h, a.X, a.Y); ok {
			if dir == "all around" {
				b.WriteString("Land all around is unknown to you.\n")
			} else {
				fmt.Fprintf(&b, "Land to the %s is unknown to you.\n", dir)
			}
		}
	}
	return b.String()
}

// joinAnd joins items with commas and a final "and" ("a", "a and b",
// "a, b and c").
func joinAnd(items []string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	}
	return strings.Join(items[:len(items)-1], ", ") + " and " + items[len(items)-1]
}

// socialContext renders a compact bonds/debts/reputation/rumor block.
func socialContext(s *sim.State, idx int) string {
	var b strings.Builder
	var bonds []string
	for _, r := range s.Relations {
		if r.From != idx {
			continue
		}
		switch {
		case r.Affection >= 100:
			bonds = append(bonds, fmt.Sprintf("you like %s", s.Agents[r.To].Name))
		case r.Affection <= -100:
			bonds = append(bonds, fmt.Sprintf("you resent %s", s.Agents[r.To].Name))
		case r.Trust <= -100:
			bonds = append(bonds, fmt.Sprintf("you distrust %s", s.Agents[r.To].Name))
		}
	}
	if len(bonds) > 4 {
		bonds = bonds[:4]
	}
	if len(bonds) > 0 {
		fmt.Fprintf(&b, "People: %s.\n", strings.Join(bonds, "; "))
	}
	// Last-conversation callback (TASK-22): the durable record ring gives
	// prompts continuity across encounters.
	if r, ok := sim.LastConversationInvolving(s, idx); ok {
		var others []string
		for _, p := range r.Participants {
			if p != idx && p >= 0 && p < len(s.Agents) {
				others = append(others, s.Agents[p].Name)
			}
		}
		if len(others) > 0 && r.Gist != "" {
			fmt.Fprintf(&b, "Last conversation, with %s: %s\n", strings.Join(others, " and "), r.Gist)
		}
	}
	for _, d := range s.Debts {
		if d.Status != "open" {
			continue
		}
		if d.Debtor == idx {
			fmt.Fprintf(&b, "You owe %s one %s (due %s).\n", s.Agents[d.Creditor].Name, d.Kind, clock.Format(d.Due))
		} else if d.Creditor == idx {
			fmt.Fprintf(&b, "%s owes you one %s.\n", s.Agents[d.Debtor].Name, d.Kind)
		}
	}
	rep := sim.Reputation(s, idx)
	switch {
	case rep >= 700:
		b.WriteString("Your word is respected in the village.\n")
	case rep <= 300:
		b.WriteString("People say you don't keep your word.\n")
	}
	best := sim.KnownRumor{Confidence: -1}
	for _, kr := range s.Agents[idx].Known {
		if kr.Confidence > best.Confidence && kr.From >= 0 { // heard, not own secret
			best = kr
		}
	}
	if best.Confidence > 0 {
		fmt.Fprintf(&b, "You have heard: %q\n", best.Text)
	}
	return b.String()
}

// villageLaw renders the norms in force, standing judgments, and (while the
// village convenes) the meeting call. Empty for a lawless village.
func villageLaw(s *sim.State, idx int) string {
	var b strings.Builder
	var rules []string
	var judgments []string
	for _, n := range s.Norms {
		if !n.Active {
			continue
		}
		if n.Kind == sim.NormExile {
			if n.Target == idx {
				judgments = append(judgments, fmt.Sprintf("You are exiled from the village (day %d) — the village shuns you.", n.DayPassed))
			} else if n.Target >= 0 && n.Target < len(s.Agents) {
				judgments = append(judgments, fmt.Sprintf("%s is exiled from the village (day %d).", s.Agents[n.Target].Name, n.DayPassed))
			}
			continue
		}
		proposer := "someone"
		if n.Proposer >= 0 && n.Proposer < len(s.Agents) {
			proposer = s.Agents[n.Proposer].Name
		}
		rules = append(rules, fmt.Sprintf("- %s (passed day %d, %s's proposal, %s)", n.Text, n.DayPassed, proposer, n.Tally))
	}
	if len(rules) > 0 {
		header := "Village law:"
		if c := s.MeetingConvention; c != nil {
			header = fmt.Sprintf("Village law (decided at the daily meeting, %s):", clock.FormatTOD(c.OpenSecond))
		}
		b.WriteString(header + "\n")
		b.WriteString(strings.Join(rules, "\n"))
		b.WriteString("\n")
	}
	for _, j := range judgments {
		b.WriteString(j)
		b.WriteString("\n")
	}
	if sim.AtMeeting(s, idx) {
		when := "the assembly"
		if c := s.MeetingConvention; c != nil {
			when = fmt.Sprintf("the %s assembly", clock.FormatTOD(c.OpenSecond))
		}
		fmt.Fprintf(&b, "The village is gathering at the meeting place for %s — you can raise grievances and vote there.\n", when)
	}
	return b.String()
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

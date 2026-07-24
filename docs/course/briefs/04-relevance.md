---
grounding: ["[[sim-state-reducer]]", "[[deterministic-rng]]"]
---

## Module 4: What Matters Right Now (the three-term memory window)

### Teaching Arc
- **Metaphor:** Packing a small daypack. You can't carry the whole closet: you pack what's *important* (first-aid kit), weighted by *freshness* (today's forecast, not last month's), and matched to *where you're going* (river crossing → rope, not sunscreen). Ten items, chosen three ways.
- **Opening hook:** A villager can hold thousands of memories, but when she stops to think, only 10 fit in her head. Which 10?
- **Key insight:** The window is scored, not scrolled: importance (salience) fades by half each day it ages, and — new in spec 042 — a relevance term measures how close each memory's coordinates sit to a fresh snapshot of *right now* (her "situation vector"). Score = the two added together. Plus two wildcard slots for old, random memories.
- **"Why should I care?":** "Recency + importance + relevance" is the canonical memory-retrieval recipe for AI agents (it comes from the famous Stanford generative-agents paper). If you ever tell an AI assistant "my chatbot forgets what matters," this module gives you the exact three dials to name.

### Code Snippets (pre-extracted, use verbatim)

File: internal/sim/memory.go (lines 376-380) — the fallback that keeps old behavior sacred
```go
func SelectMemoriesRelevant(a *Agent, seed uint64, agentIdx int, tick int64, k int, query []float32, queryModel string) []Memory {
	if query == nil {
		return SelectMemories(a, seed, agentIdx, tick, k)
	}
```
(Balanced: add closing lines or present as the fuller excerpt below — writing agent may instead show lines 376-380 plus a comment line and `}` — MUST stay bracket-balanced. Recommended: use the scoring excerpt below as the main block and this nil-check inside a callout as prose.)

File: internal/sim/memory.go (lines 396-407) — the score itself (main translation block)
```go
	for i, m := range a.Memories {
		age := tick - m.Tick
		if age < 0 {
			age = 0
		}
		// integer-friendly decay: halve per whole game-day of age (EXACTLY
		// SelectMemories' weight), normalized by the salience ceiling.
		w := float64(m.Salience)
		for d := age / halfLifeTicks; d > 0; d-- {
			w /= 2
		}
		all[i] = scored{m: m, score: w/MaxSalience + relevance01(m, query, queryModel), idx: i}
	}
```

File: internal/sim/memory.go (lines 454-469) — cosine similarity, the whole thing
```go
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
```

File: internal/mind/embedder.go (lines 219-231) — what "right now" looks like (excerpt; keep balanced by ending at a sensible line with closing braces as needed; caption "building the situation string — time, place, needs, intent, neighbors — embedded on a regular cadence")
```go
func renderSituation(s *sim.State, idx int) string {
	a := &s.Agents[idx]
	var b strings.Builder
	if s.Night {
		b.WriteString("night")
	} else {
		b.WriteString("daytime")
	}
	if place := sim.PlaceAt(s, a.X, a.Y); place.Desc != "" {
		fmt.Fprintf(&b, " · at %s (%d,%d)", place.Desc, a.X, a.Y)
	} else {
		fmt.Fprintf(&b, " · at (%d,%d)", a.X, a.Y)
	}
```
(Present with a final `…` translated line + closing `}` to stay balanced, captioned as the opening of the function.)

### Facts to teach (all verified)
- Window size is 10: top-8 by score + 2 "serendipity" picks — seeded-random draws from the oldest half, so an ancient memory occasionally resurfaces. The wildcards use the same deterministic dice as everything else (same replay law).
- The score: `sal01 + rel01`. sal01 = salience halved per game-day of age, scaled to 0–1. rel01 = cosine similarity of memory-vector vs situation-vector, scaled to 0–1. Equal partners.
- Neutral = 0.5: a memory with no vector (or from a different model) scores exactly middle relevance — never punished, never promoted. Only genuine similarity moves ranks.
- The situation vector: every planning interval, the embedder renders a one-line description of the villager's *now* (night/day, place, worst needs, what she's trying to do, who's nearby), embeds it, and records it. "Relevant" = close to that.
- Guardrail (a real design decision worth teaching): the famous reference design decays memories from *last access* — recalling a memory keeps it fresh. promptworld deliberately REJECTED that: reading a memory must never change it, or replay (and pure functions) break. Recency counts from creation only.
- Ceiling insight: a perfect relevance match adds at most +0.5 — it can lift a modest memory over rivals up to 5 salience points stronger, but can never outshout fresh trauma. Relevance whispers; trauma shouts.
- Isolation: the scorer's only input is that one villager's own memories. Two villagers with identical experiences cannot influence each other's window — proven by a test.

### Interactive Elements
- [x] **Code↔English translation** — the scoring loop AND relevance01 (two blocks; renderSituation optional third if space allows).
- [x] **Quiz** — 4 questions: (1) scenario: "A villager returns to the river where she was robbed 20 days ago. The memory's salience has halved 20 times — basically zero. Can it still make her window?" (yes — relevance term + also the serendipity wildcards; accept the relevance answer as primary); (2) "Your chatbot keeps bringing up whatever the user said most recently instead of what matters. Which dial is over-weighted, and which is missing?" (recency over-weighted; relevance missing); (3) debugging: "After swapping embedding models, all relevance scores went neutral. Why is that the *designed* behavior?" (model identity mismatch → 0.5 — different charts' coordinates don't compare); (4) architecture: "Why does promptworld count recency from when a memory was CREATED, not when it was last recalled?" (reads must not mutate; replay/purity).
- [x] **Other** — hero visual: a "daypack" ranking board — 6 memory cards each showing salience stars, age, and a relevance meter, with computed score chips; the top ones highlighted as "packed", one wildcard slot drawn from the bottom. Static cards (no JS needed beyond styling).

### Reference Files to Read
- `references/interactive-elements.md` → "Code ↔ English Translation Blocks", "Multiple-Choice Quizzes", "Pattern Cards"
- `references/content-philosophy.md` → all
- `references/gotchas.md` → all

### Connections
- **Previous module:** "Giving Memories Coordinates" — every memory (and the villager's "now") has coordinates; this module spends them.
- **Next module:** "Trust, But Verify" — before this new scoring is allowed to actually steer villagers' minds, it runs in shadow mode and must prove itself with recorded evidence.
- **Tone/style notes:** teal accent; module 4 = even background. Section id `module-4`. The Stanford paper may be named ("Generative Agents", 2023) with a tooltip.

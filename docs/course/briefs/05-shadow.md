---
grounding: ["[[event-types]]", "[[cli-promptworld]]"]
---

## Module 5: Trust, But Verify (shadow mode and the go/no-go gate)

### Teaching Arc
- **Metaphor:** An understudy shadowing a lead actor. For weeks she performs every line in the wings — full costume, full effort — while the audience only ever sees the lead. The director compares the two performances night after night, and only when the notes prove she's ready does she step on stage.
- **Opening hook:** The new relevance scoring is built, tested, fast… and NOT allowed to influence a single villager's thoughts yet. That's not caution theater — it's the feature.
- **Key insight:** Shadow mode computes BOTH rankings on every thought — serves the old one, records the disagreement between them as durable events — so the decision to flip the switch is made from evidence, not vibes. The switch itself is a one-word config change: `"" → "shadow" → "on"`.
- **"Why should I care?":** This is how professionals ship risky changes: dark launches, shadow traffic, feature flags, measured rollouts. Ask your AI assistant for "shadow mode with divergence metrics" and you'll never have to YOLO a scoring change into production again.

### Code Snippets (pre-extracted, use verbatim)

File: internal/mind/prompt.go (lines 79-89) — the mode gate (main block)
```go
func selectWindow(s *sim.State, idx, k int, tick int64, mode string) []sim.Memory {
	a := &s.Agents[idx]
	if mode == world.MemoryRelevanceOn {
		var query []float32
		if len(a.SitVec) > 0 {
			query = a.SitVec
		}
		return sim.SelectMemoriesRelevant(a, s.Seed, idx, tick, k, query, a.SitVecModel)
	}
	return sim.SelectMemories(a, s.Seed, idx, tick, k)
}
```

File: internal/sim/cognition.go (lines 172-182) — what a disagreement record looks like
```go
type MemoryDivergencePayload struct {
	Agent        int     `json:"agent"`
	Tick         int64   `json:"tick"`
	Mode         string  `json:"mode"`
	Legacy       []int64 `json:"legacy"`
	Augmented    []int64 `json:"augmented"`
	Overlap      int     `json:"overlap"`
	Displacement int     `json:"displacement"`
	Vectorless   int     `json:"vectorless"`
	SitTick      int64   `json:"sit_tick"`
}
```

### Facts to teach (all verified)
- The flag lives in the world's config file (`world.json`): `memory_relevance` = `""` (off) / `"shadow"` / `"on"`. A typo is refused at startup — it can't silently run as "off".
- Shadow mode's hard promise, proven by a test: with `"shadow"`, every prompt a villager sees is byte-for-byte identical to off-mode. The only difference is extra recorded telemetry. (The test builds a world where the two rankings provably differ, then asserts the prompts don't.)
- Every planning moment emits one `cog.memory_divergence` event: both top-10 lists (as memory ids), how many they share (overlap), how far shared ones moved (displacement), how many candidates had no vector yet.
- The operator reads the evidence with one command: `promptworld divergence <world>` — per-villager, per-game-day averages. The go/no-go decision (with its numbers) is recorded on the project board before anyone flips to "on".
- Real numbers from this feature's own validation run: at max speed, embedding everything cost +5–6% wall-clock (budget was 10%); a live shadow hour showed overlap 1.00 — villagers were still young, memories still fit entirely in the window, so the rankings couldn't disagree yet. Evidence takes time. That's the point.
- Even in "on" mode the divergence keeps recording — the understudy metaphor: the director never stops taking notes.
- Close the course by zooming out: the full journey — event → memory → coordinates → window → shadow → (one day) on — all of it recorded, all of it replayable. The village never forgets, and never lies about what it remembered.

### Interactive Elements
- [x] **Code↔English translation** — both snippets.
- [x] **Quiz** — 4 questions, this is the capstone so it may reach back to earlier modules (all concepts taught): (1) scenario: "You built a new ranking algorithm for your app's feed. Your AI assistant says 'deployed!' What did it skip that promptworld wouldn't?" (shadow phase + divergence evidence + a recorded decision); (2) "In shadow mode a bug makes the new scorer crash on one villager. What do users/villagers experience?" (nothing — legacy window still serves; shadow is read-only observation… note: emphasize prompts unchanged); (3) tracing across the course: order the life of one memory: event happens → memory_added (salience, where/why) → embedder mails vector → situation vector snapshot → three-term score in shadow → divergence recorded; (4) architecture: "Why record divergence as EVENTS instead of printing to a log file?" (artifacts: replayable, auditable, same store as everything else — decisions derive from recorded evidence).
- [x] **Other** — "the dial" hero visual: three-position switch card (`""` / `"shadow"` / `"on"`) with what each position does (serve legacy / serve legacy + record both / serve new + keep recording). Plus a small mock `promptworld divergence` output table styled as terminal output (use the real column concepts: overlap@K, promoted share, displacement — keep numbers plausible: overlap 1.00 for a young world).
- [ ] Group chat / flow — already covered in modules 2 and 3.

### Reference Files to Read
- `references/interactive-elements.md` → "Code ↔ English Translation Blocks", "Multiple-Choice Quizzes", "Callout Boxes"
- `references/content-philosophy.md` → all
- `references/gotchas.md` → all

### Connections
- **Previous module:** "What Matters Right Now" — the three-term score exists and is tested; this module is about *earning the right to use it*.
- **Next module:** none — this closes the course. End with a short "what you can now say to an AI assistant" recap card: event sourcing, embeddings, recency/importance/relevance, shadow mode, feature flag — the vocabulary they leave with.
- **Tone/style notes:** teal accent; module 5 = odd background. Section id `module-5`.

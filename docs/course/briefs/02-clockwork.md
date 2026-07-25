---
grounding: ["[[sim-state-reducer]]", "[[event-log]]", "[[sim-loop]]", "[[deterministic-rng]]"]
---

## Module 2: The Clockwork Heart (why replay never lies)

### Teaching Arc
- **Metaphor:** A player piano. The paper roll (event log) IS the music; any piano that loads the roll plays the identical performance, note for note. The pianist who punched the roll (the AI mind) is not needed for the replay — and is never allowed to reach into the piano's gears mid-performance.
- **Opening hook:** You can stop the promptworld daemon, delete the running state, restart it — and the village comes back *byte-for-byte identical*, every memory intact. How?
- **Key insight:** The sim has one law: state is only ever changed by applying recorded events, in order. AI output becomes recorded events *through two guarded doors* — it never touches the clockwork directly. That's what makes perfect replay possible.
- **"Why should I care?":** This is event sourcing — one of the most powerful architecture patterns you can ask an AI assistant for. If your app's state ever gets corrupted, or you wish you could "rewind and see what happened," this is the pattern you name.

### Code Snippets (pre-extracted, use verbatim)

File: internal/sim/loop.go (lines 263-268) — the whitelist on the door (trim: present these lines inside their map; include enough context lines to be bracket-balanced — e.g. render as a small excerpt with a `var injectSocialAllowed = map[string]bool{` opening line, the entries, and closing `}` — the writing agent should present exactly the three new entries plus 1-2 neighbors if needed; keep it balanced)
```go
	"agent.memory_embedded":    true,
	"agent.situation_embedded": true,
	// ...
	"cog.memory_divergence": true,
```
IMPORTANT: for the translation block, wrap as a balanced excerpt:
```go
var injectSocialWhitelist = map[string]bool{
	"agent.memory_embedded":    true,
	"agent.situation_embedded": true,
	"cog.memory_divergence":    true,
}
```
— label it "excerpt (three of the whitelist's entries)" in the caption above the block, since the real map has more entries. Do NOT invent other entries.

File: internal/sim/state.go (lines 428-442) — the reducer applying a recorded embedding
```go
	case "agent.memory_embedded":
		// Spec 042: attach a recorded embedding to the {agent, mem_seq} memory,
		// copy-verbatim — the reducer never computes or inspects a vector
		// (model-free sim doctrine, spec 030 lineage). A missing target is a
		// deliberate NO-OP, never an error: the agent may have died or the
		// memory consolidated away in the async gap. Newest-first scan: the
		// companion trails its memory by moments, so the target is near the end.
		var p MemoryEmbeddedPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		a, err := agent(p.Agent)
		if err != nil {
			return err
		}
```
(Balanced as shown — a `case` fragment; wrap it in a `switch e.Type {` opener and closing `}` lines, captioned "excerpt from the reducer's switch". Keep brackets balanced.)

### Facts to teach (all verified)
- Every world is a SQLite file with an append-only `events` table: seq, tick, type, payload. Deleting is impossible by design (database triggers forbid it). Gapless sequence numbers are checked.
- The reducer (`State.Apply`) is the ONLY way state changes — the same function runs live and during replay. Replay = load newest verified snapshot + re-apply events.
- Randomness itself is deterministic: every random draw comes from (world seed, purpose, tick), so replays roll the same dice.
- The AI mind runs OUTSIDE this clockwork. Its output enters through exactly two doors, as recorded events: an intent door (what the villager decides to do) and a social door (whitelisted event types only — see snippet).
- New in spec 042: three memory-related event types were added to the social door's whitelist. Even the AI's "understanding" of a memory (its embedding vector) must enter as a recorded event.
- The payoff sentence for module 3: "so if we want AI to help with memory, the AI must live outside and mail its work in."

### Interactive Elements
- [x] **Code↔English translation** — both snippets above.
- [x] **Group chat animation** (MANDATORY — this is the course's group chat): participants "⚙️ Clockwork (sim)", "🧠 Mind (AI)", "🗄️ Event Log". Flow: Mind: "The villager should go hunt — I've decided." / Clockwork: "Write it in the log first. Nothing is real until it's an event." / Log: "Recorded: agent.intent_set, seq 4,102." / Clockwork: "NOW she hunts. And in every replay, she'll hunt at exactly this tick, forever." / Mind: "What if I just tweak her hunger directly? It's faster—" / Clockwork: "Denied. You have two doors. Use them." — teaches the doors + append-only law with personality.
- [x] **Quiz** — 3 questions: (1) debugging: "A replayed world differs from the original. Based on this module, which is the ONLY kind of place the bug can hide?" (something changed state outside recorded events / a non-recorded input like wall-clock time); (2) scenario: "You want your own app to have an 'undo anything' feature. What do you ask your AI assistant to build?" (event sourcing / append-only log + reducer); (3) "Why must the AI's embedding vectors be recorded as events instead of recomputed during replay?" (models aren't bit-reproducible; replay must not depend on any model).
- [x] **Other** — a "two doors" diagram: Mind box outside, Clockwork box inside, two labeled doors (intent / social with whitelist), arrows only inward through doors. Static, cards + arrows.

### Reference Files to Read
- `references/interactive-elements.md` → "Group Chat Animation", "Code ↔ English Translation Blocks", "Multiple-Choice Quizzes"
- `references/content-philosophy.md` → all
- `references/gotchas.md` → all

### Connections
- **Previous module:** "The Village That Writes Everything Down" — memories are structured records with salience; this module explains the machine that guards those records.
- **Next module:** "Giving Memories Coordinates" — the new spec-042 feature: an AI helper that reads memories and mails back "meaning coordinates" (embeddings) through the social door.
- **Tone/style notes:** teal accent; "the clockwork" = deterministic sim, "the mind" = AI side; module 2 = even (alternate background). Section id `module-2`.

---
grounding: ["[[overview]]", "[[executor]]", "[[event-types]]"]
---

## Module 1: The Village That Writes Everything Down

### Teaching Arc
- **Metaphor:** A lighthouse keeper's logbook. Nothing "happened" at the lighthouse unless it's written in the log — and every entry says where, when, and why.
- **Opening hook:** promptworld is a tiny simulated village where AI villagers hunt, build fires, gossip, and sometimes die — and when you watch one panic about a wolf it saw two days ago, you're watching a *memory* work.
- **Key insight:** A villager's memory is not a vague AI impression — it's a concrete, structured record (one line of text + a 1–10 importance number + where/why context) born the instant something happens.
- **"Why should I care?":** When you build with AI agents, "give the agent memory" is a design decision, not magic. Seeing memory as *data with a schema* lets you tell an AI assistant exactly what to store, score, and show.

### Course intro (this module opens the course)
Open with what promptworld IS (2-3 sentences + a visual): an always-on simulation daemon — a program that runs a 64×64 tile village world one second at a time, where every villager is a small AI agent with needs (food, warmth, rest), a body, and a private memory stream. Humans watch and nudge; the world runs itself. THEN the hook: trace one event — a hunting spear snapping on its last use — into a memory.

### Code Snippets (pre-extracted, use verbatim)

File: internal/sim/memory.go (lines 229-238) — the salience table (trim to these lines exactly)
```go
const (
	salTalk           = 3
	salHunt           = 4
	salFire           = 5
	salShelter        = 6
	salStarvingForage = 5
	salColdNight      = 5
	salNearDeath      = 9
	salWitnessDeath   = 10
```
(NOTE for translation block: this is an excerpt of a longer const block — close it with a final line showing `)` only if you include it in the code; simplest is to present lines 229-238 plus a closing `)` line — the validator requires bracket balance, so include the closing paren as its own translated line.)

File: internal/sim/memory.go (lines 190-198) — how a memory is born
```go
func situatedMemoryEvent(tick int64, agent, salience int, where *MemoryPlace, why, origin, format string, args ...any) store.Event {
	return store.Event{
		Tick: tick, Type: "agent.memory_added",
		Payload: mustPayload(MemoryAddedPayload{
			Agent: agent, Text: situateText(fmt.Sprintf(format, args...), where, why),
			Salience: salience, Subject: -1, Where: where, Why: why, Origin: origin,
		}),
	}
}
```

File: internal/sim/memory.go (lines 471-474) — how a memory looks when shown
```go
// FormatMemory renders one memory line as prompts and soul.md show it.
func FormatMemory(m Memory) string {
	return fmt.Sprintf("%s (%d★) %s", clock.Format(m.Tick), m.Salience, m.Text)
}
```

### Facts to teach (all verified)
- 1 tick = 1 game second; the world runs continuously.
- Memory text is "situated": `<base> at <place> (x,y) — <why>` (e.g. "Built a fire at the woods (4,7) — cold night coming."). The place description is baked in at the moment of the event.
- Salience 1–10 = how formative. Talk=3 (routine), near-death=9, witnessing a death=10. Dreams from the world's angel = 8 ("the divine reliably surfaces without outranking real trauma" — quote the actual comment).
- Salience is a *control signal*, not just a rating: at 9+ it interrupts the villager's thinking; at 4+ a memory about a neighbor becomes tellable gossip.
- Every memory also records an origin: did I do it (action), see it (witness), or hear about it (report/gossip)? Villagers can honestly say "I saw it myself."

### Interactive Elements
- [x] **Code↔English translation** — the salience table AND situatedMemoryEvent (two blocks).
- [x] **Quiz** — 3 questions, scenario style: (1) "A villager watches a neighbor steal from a chest. Which matters more for whether she'll gossip about it: the memory's text or its salience number?" (salience ≥4 gates gossip); (2) "You're building a shopping assistant with AI memory. promptworld stamps where/why context AT THE MOMENT of the event instead of reconstructing it later. Why?" (context is cheap now, impossible later — teaches the pattern); (3) tracing: "spear breaks → which parts of the memory record exist?" (text, salience 8, where, why, origin=action).
- [ ] Group chat — no (module 2 has it)
- [ ] Data flow — no (module 3 has it)
- [x] **Other** — a "memory anatomy" hero visual: one rendered memory line `Day 3 07:12 (8★) My spear broke at the river (12,31) — hunting for the village.` with labeled callouts pointing at timestamp/stars/text/place/why. Build with simple positioned divs/cards, tokens only.

### Reference Files to Read
- `references/interactive-elements.md` → "Code ↔ English Translation Blocks", "Multiple-Choice Quizzes", "Callout Boxes"
- `references/content-philosophy.md` → all
- `references/gotchas.md` → all

### Connections
- **Previous module:** none — this opens the course; include the course-level intro hero (title, one-paragraph promise: "by the end you'll know how AI villagers remember — and how to build memory into your own AI projects").
- **Next module:** "The Clockwork Heart" — why the village's history is one long append-only log that can replay perfectly, and why no AI is allowed inside that clockwork.
- **Tone/style notes:** teal accent; villagers are "she/they"; the sim daemon is "the clockwork"; the LLM side is "the mind". Section is `<section class="module" id="module-1">`.

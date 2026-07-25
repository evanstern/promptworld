---
grounding: ["[[event-log]]", "[[sim-state-reducer]]"]
---

## Module 3: Giving Memories Coordinates (embeddings + the async embedder)

### Teaching Arc
- **Metaphor:** A star chart. Every memory gets sky coordinates — and memories about similar things end up as neighboring stars. "My spear broke at the river" sits right next to "lost my axe by the water", and far from "the harvest festival". Similarity becomes *distance you can measure*.
- **Opening hook:** How would a computer know that "my spear broke at the river" and "I lost my axe by the water" are about the same kind of trouble? Not by matching words — none match.
- **Key insight:** An embedding model turns text into a list of 384 numbers (a vector) such that similar meanings get nearby numbers. promptworld runs a tiny local one, OUTSIDE the clockwork, and mails each memory's coordinates back in as a recorded event.
- **"Why should I care?":** Embeddings are the engine behind "search by meaning," RAG, and every "find related items" feature. Knowing the moving parts (model, vector, cosine similarity) — and that embedding APIs give *slightly different numbers each call* — lets you direct an AI assistant to build search that works and replays that don't break.

### Code Snippets (pre-extracted, use verbatim)

File: internal/mind/embedder.go (lines 135-140) — never block the clockwork
```go
func (e *Embedder) Observe(events []store.Event) {
	select {
	case e.events <- events:
	default:
	}
}
```
(Translation angle: "hand the new events to the embedder IF it's ready; if its queue is full, drop and move on — the village's clock never waits for the AI.")

File: internal/llm/providers.go (line 542) — the warm-pin body (present as a balanced 1-line-in-context excerpt)
```go
	body, err := json.Marshal(map[string]any{"model": o.model, "keep_alive": -1})
```
(Caption: the one line that keeps the embedding model loaded in memory forever — `keep_alive: -1` means "never unload".)

File: internal/sim/agents.go (lines 882-889 area) — the companion event's payload (agent should render the struct with fields Agent int / MemSeq int64 / Vec []float32 / Model string as it exists; keep balanced)
```go
	MemoryEmbeddedPayload struct {
		Agent   int       `json:"agent"`
		MemSeq  int64     `json:"mem_seq"`
		Vec     []float32 `json:"vec"`
		Model   string    `json:"model"`
	}
```
(NOTE: verify field names against the caption "the envelope the embedder mails through the social door" — Agent = whose memory, MemSeq = which memory (its log sequence number), Vec = the 384 coordinates, Model = which model produced them.)

### Facts to teach (all verified)
- The embedder is a background helper on the mind side. It watches for `agent.memory_added` events, sends the memory's exact text to a local embedding model (all-minilm, 384 dimensions, served by Ollama on the same machine), and injects `agent.memory_embedded {agent, mem_seq, vec, model}` through the social door.
- The clockwork never computes vectors; the reducer copies them verbatim (module 2's snippet showed this). During replay, vectors are just data being re-read — zero model calls, proven by a test that meters endpoint calls.
- Ordering: a memory's coordinates always arrive AFTER the memory itself (the log's order guarantees it). In the gap, the memory simply has no vector yet — that's a legal state.
- Failure is loud but harmless: if the embedding model is down, memories still happen — they just stay vectorless forever, and the operator sees exactly one warning (not a stall, not a crash, not spam).
- Model identity travels with every vector. Numbers from different models are like coordinates from different star charts — comparing them is meaningless, so the system refuses.
- Warm-pin: the model is tiny (~46 MB) and gets pinned in memory at startup so there's no cold-start lag. The big chat models are deliberately NOT pinned.
- Why local and recorded, not a cloud API: cloud embedding APIs return *slightly different* vectors for the same text on different calls — poison for a system that must replay byte-identically. Record once, replay forever.

### Interactive Elements
- [x] **Data flow animation** (MANDATORY — this is the course's flow animation): actors: "📜 Event Log" → "🛰️ Embedder" → "⭐ all-minilm (local model)" → "🚪 Social Door" → "⚙️ Reducer". Steps: (1) log emits `agent.memory_added "My spear broke…"`; (2) embedder picks it up (clock keeps ticking — note under step); (3) model returns 384 numbers; (4) embedder mails `agent.memory_embedded` through the door; (5) reducer attaches the vector to the memory — now it's permanent history. Use `data-steps` JSON per the reference.
- [x] **Code↔English translation** — Observe() and the payload struct (two blocks; warm-pin line can ride inside a callout instead if three blocks crowd the module).
- [x] **Quiz** — 3 questions: (1) scenario: "You ask an AI assistant to add semantic search to your notes app using a cloud embeddings API, and cache results forever. A colleague says 'just re-embed on demand.' What breaks?" (same text ≠ same vector across calls; cached/recorded vectors are the stable ground truth); (2) debugging: "Villagers' new memories suddenly have no vectors, but the village runs fine. First place to look?" (embedding model/endpoint down — loud warning, vectorless-but-alive is the designed failure mode); (3) tracing: order the five flow steps (reuse flow actors).
- [x] **Other** — a "neighboring stars" visual: 6-8 memory chips scattered on a dark panel, similar ones clustered (spear-broke / axe-lost cluster near "water" region; festival far away). Static positioned chips; label axes "this is a 2D cartoon of 384 dimensions" in a caption.

### Reference Files to Read
- `references/interactive-elements.md` → "Message Flow / Data Flow Animation", "Code ↔ English Translation Blocks", "Multiple-Choice Quizzes"
- `references/content-philosophy.md` → all
- `references/gotchas.md` → all

### Connections
- **Previous module:** "The Clockwork Heart" — the two doors and the replay law; this module is the first new machine built on those rules (spec 042).
- **Next module:** "What Matters Right Now" — having coordinates for every memory, the village can finally ask: which memories are *relevant to this moment*?
- **Tone/style notes:** teal accent; module 3 = odd background; keep calling the deterministic side "the clockwork". Section id `module-3`.

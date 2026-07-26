---
name: social-fabric
description: The conflict engine — directed relation edges, debt ledger with computed reputation, rumors with provenance and mutation, authored secrets, chest-theft consequences, the place-knowledge talk sidecar, and the conversation record ring. Conversation scene machinery itself split to [[social-fabric-conversations]].
kind: component
sources:
  - internal/sim/social.go
verified_against: 30912a9cd5d2334f76425ac8ca5b74a7a7c90876
---

# Social fabric

TASK-8's conflict engine: everything villagers feel about each other, owe each
other, and say about each other — event-sourced in the deterministic core, with
model creativity (dialogue, paraphrase) entering only as recorded events.

## How it works

**Edges** (`sim/social.go`): directed `Relation{From, To, Trust, Affection}`
(−1000..1000, reducer-clamped, lazy). One event type moves them —
`social.relation_changed` with a reason — emitted by fixed rules: talk +5/+5
affection, give (+30 trust/+20 affection receiver→giver), promise broken (−150/−50
creditor→debtor), rumor tone/4 listener→subject, conversation tones ×12/×25.

**Ledger**: a give to a starving neighbor — one unit of `Inv.FoodRaw` moves
giver→receiver (spec 012 widened the single `Food` field to a raw/cooked/meals
triplet; giving stays denominated in the least-nutritious raw form) — opens
`Debt{due +2 game days}` (reducer-internal on `social.gave`); a matching
give-back settles it kept. Spec 013 (US1) added a carried-bulk guard: the
executor's `repayable`/`giveable` checks additionally require the receiver have
free bulk (`freeBulk(Inv) > 0`) before offering a give — a starving villager
already at the cap is carrying food and would eat rather than receive — and the
reducer clamps the receive defensively at `bulkCap`, so even a forged over-cap
`social.gave` can't push a recipient over it. The
executor's hourly due-check breaks overdue debts permanently — with the trust
penalty and a gossip-seed memory ("X never repaid…"). `Reputation` is computed
(500 +100·kept −200·broken), never stored.

**Rumors**: registry `Rumor` identity + per-holder `KnownRumor` variants (text,
confidence, heard-from, tick — the From chain IS the provenance). Deterministic
birth from salient memories about others (`Memory.Subject/Tone`); confidence
decays ×4/5 per hop, floor 25 kills tellability; hearing shifts affection toward
the subject. During primitive talks the executor passes rumors **verbatim** (the
model-free floor); conversations paraphrase (mutation on retell, recorded in the
event). `TellableFor` never surfaces secrets.

**Secrets**: one authored self-rumor per persona (`persona.Secrets`), seeded as
tick-0 events; only the conversation driver may pass one — owner→listener trust ≥
`SecretTrustGate` (700) plus a seeded 1-in-3 roll — after which it spreads like
any rumor.

**Theft** (spec 013 US4, FR-011/012, research R5): a non-owner withdrawing from a
builder-owned chest ([[executor]]) is never blocked — the goods already moved —
but always marked, through a companion batch the executor co-emits in the same
tick as the `agent.withdrew`: `social.chest_taken{owner, taker, x, y}` is the
distinct taking record itself (reducer-effect-free, chronicle/TUI material, same
idiom as `social.conversation_turn`); a `social.relation_changed` owner→taker
moves the edge through the same fixed-rule machinery as talk/give/broken-promise,
reason `"theft"`, at `theftTrustDelta` (−120) trust and `theftAffectionDelta`
(−40) affection; the owner (if living) gets a subject-tagged memory of the taker
at `theftMemoryTone` (−60) regardless of distance — a `TellableFor` gossip seed,
the same any-distance exemption a chest owner's "my things were taken" grievance
needs to travel, stamped `Origin: "report"` (spec 030 — the owner learned of it,
rather than seeing it happen, even about their own chest) — and every living,
awake villager within `witnessRadius` (8) of
the chest, excluding the taker and the owner (who already has the stronger
any-distance memory), gets its own witness memory at the same tone, stamped
`Origin: "witness"` — direct perception, per [[agent-mind]]'s provenance
classifier. Since spec
019 (US1) both are built through `situatedMemoryAboutEvent` (memory.go) rather
than the bare `memoryAboutEvent`, so each carries a `Where` situated by the
rememberer's OWN tile — `PlaceAt(s, owner.X, owner.Y)` for the owner,
`PlaceAt(s, witness.X, witness.Y)` for each witness (a witness remembers where
it stood, not where the chest was). Witness memories carry no `Why` — the
witness did not drive the act ([[agent-mind]]'s situated-memory grammar). A
dead owner still gets the record, the relation delta, and the witness memories —
only the owner's own memory is skipped (the dead don't remember; the village
does). Owner withdrawing from their own chest emits `agent.withdrew` alone, no
companion batch (FR-011).

**Conversations** (`mind/convo.go`, scenes in TASK-22) — split into
[[social-fabric-conversations]]: on the executor's `agent.talked` beat, gated
by the spec-042 memory-relevance mode and the spec-061 novelty SHIM
(removable, compensating for weak model-side variety), an admitted talk
founds a **scene** that passes the [[cognition]] router and pins its
provider at founding. The founding pair plus any nearby awake villager
round-robin turns; one outcome call returns gist, topic tags, tones, and a
rumor paraphrase, landing as ONE atomic `inject_social` batch — gist
memories, tone edges, the record below — unless staleness (TASK-32) or an
unrecovered TASK-42 parse failure kills it. See the child for the full
founding-gate, staleness, and retry mechanics.

**Place-knowledge sidecar** (spec 041 US5, [[mental-maps]]): every founded
talk, hail-founded included, ALSO exchanges up to `placeTellCap` fresh facts
per direction the other party lacks or holds staler (`tellablePlaces`) —
directions and mechanics live in [[executor]]'s `talkEvents`; the reducer's
`applySocial` arm for `social.place_told{from, to, facts}` upserts into the
RECEIVER's map only where the receiver's held fact is absent or staler
(`Seen` compared — fresher knowledge, even the receiver's own, never loses to
secondhand), a map-less receiver a no-op like `agent.saw`. Facts arrive fully
baked (told provenance, the teller's `Seen`, `Source` = the immediate
teller); companion situated memories on both sides ride the same batch as
ordinary `agent.memory_added` events ("Told X about the fire at (x,y)." /
"X told you of a fire at (x,y)."), at `salPlaceTold` (3, the talk band) —
social texture, not a formative moment.

**Conversation records** (TASK-22): `social.conversation` is no longer a reducer
no-op — the payload (`participants`, `topics`, `tones`; empty participants means
the legacy `[a, b]`) appends a `ConvoRecord` to `State.Conversations`, a bounded
ring (`convoRecordCap` 64). `LastConversationBetween` / `LastConversationInvolving`
serve it back to prompts — planner prompts carry a "Last conversation, with X:
<gist>" line, so encounters have continuity instead of amnesia.

## Connections

[[executor]] runs the deterministic acts (give/repay/talk/due-check/theft);
[[sim-state-reducer]] carries all social state; [[agent-mind]]'s planner
prompts read bonds/debts/reputation/rumors; the scribe renders the Bonds
section into soul.md. [[governance]] (TASK-13) votes over these edges and
writes violation consequences back into them; TASK-11's chronicle narrates
the conversation events. [[mental-maps]] is the knowledge store the
place-telling sidecar reads from and writes into, riding the same talk beat
as rumors and gifts. [[social-fabric-conversations]] is the split-off child
covering the scene-founding, staleness, and retry machinery this note's talk
beat triggers.

## Operational notes

First landed conversation (live, gemma4:12b-mlx): Birch — authored as finding
Cedar's silences unbearable — berated Cedar for four turns; both souls got the
gist; tones moved edges to the village's first grudge (trust −24, affection −45).
Engineering findings baked in: chat-while-working (mutual idleness starved the
fabric), planner debounce (trigger feedback loop), conversation priority lane +
worker call cap, float-tolerant tone parsing. Pace at 4x: one conversation ≈ 4
minutes wall, one at a time.

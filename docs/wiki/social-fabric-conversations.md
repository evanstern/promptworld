---
name: social-fabric-conversations
description: How a talked-beat founds a model-driven conversation scene — the spec-061 novelty SHIM, the memory-relevance snapshot gate, the cognition router and provider pin, round-robin turns landing as one atomic inject_social batch, staleness-enforced landing, and the TASK-42 retry-once tolerance.
kind: component
sources:
  - internal/mind/convo.go
verified_against: c61cd6c04ddfcd2a976c14a49ba071e8fd768a73
---

# Social fabric — conversation scenes

Split off [[social-fabric]] (spec 071): the model-driven conversation machinery
that founds, runs, and lands a scene once a talk beat fires.

## How it works

**Conversations** (`mind/convo.go`, scenes in TASK-22): on the executor's
`agent.talked` beat, the driver (slot = 1, immutable snapshot, 10-min deadline —
sized for a full scene at honest local pace)
forms a **scene**. Since spec 042, each participant's k=5 memory snapshot is
gated by the same `memoryRelevance` mode as the planner window ([[agent-mind]],
`selectWindow`): legacy/shadow modes keep today's `SelectMemories` window
byte-identical, `"on"` serves the relevance-scored window, and shadow/on modes
each record one `cog.memory_divergence` per participant at founding
([[memory-retrieval]] owns the selector and the divergence payload). Before
even that gate, since spec 061 US3 (TASK-109) `maybeStartConversation` runs
the novelty SHIM: `hasNoveltySince(p.A, p.B, priorExchange)` — `priorExchange`
the pair's [[sim-state-reducer]] `PairTalks` tick from BEFORE this founding
talk, captured in `absorb` ahead of the reducer overwriting it (spec 061 R4,
one source of truth via `State.PairLastTalk`) — requires at least one
participant to have formed a memory at or above `noveltySalienceFloor` (5,
promoted-dial-ready per FR-005 but deliberately not in tuning.json) strictly
after that prior tick; a never-talked pair (`priorExchange` 0) is vacuously
novel, so first contact always founds. Failing the gate records a suppressed
`cog.outcome` (`emitNothingNew`, reason `"nothing new since last exchange"`)
and the primitive talk that triggered the beat stands alone — no scene. This
is an explicitly-marked, removable SHIM (operator decision 2026-07-24)
compensating for weak model-side conversational variety (Birch's "I need to
tell Sage everything" fixation, 248 hails in the TASK-109 evidence): every
site carries a greppable `SHIM(TASK-109)` marker and the removal condition —
delete once the conversation tier reliably varies its own dialogue — the
FIRST place to look if conversations later feel too sparse. When a scene DOES
found on a callback pair, the prompt's last-conversation line is reframed as
"already covered" ("You two recently talked about this: `<gist>`\nSay
something new — don't just repeat it.", riding the same `LastConversationBetween`
gist below) so the model varies the exchange instead of relooping — removable
together with the gate. (The sim-side pair cooldown gating the hail founding
path itself — the layer beneath this one — is [[sim-loop]] territory.) Since
TASK-32 the beat then passes the [[cognition]]
router gate: a scene is the costliest conversation-class thought (13 points), and if
it can't land inside its staleness budget at the current speed the encounter
stays a primitive talk with a `cog.outcome{suppressed}` record. An admitted scene also pins its PROVIDER at founding (spec 024 US3,
`Mind.sceneProvider` → the orchestrator's `ResolveProvider(KindConversation)`
dry chain-walk): every utterance and the outcome call stamp the same
`Request.Provider`, so a persona keeps one voice for the whole dialogue even if
a preferable candidate frees up mid-scene — mid-scene failure flows into the
TASK-42 tolerance path, never a re-resolve or provider switch. An admitted
scene mints a telemetry identity at founding (`conversation-<founding tick>`,
agent = founding speaker) and emits `cog.thought` before the first turn.
The scene is the founding pair plus any awake villager within
`sceneJoinRadius` (2) of the founding speaker, up to `sceneCap` (4). Round-robin
turns, `ConvoTurnsPerSide` (2) each; the snapshot carries each participant's
feelings toward every other, open debts inside the scene, and the last
conversation between the founding pair (from the record ring below). One outcome
call returns gist, 1–3 topic tags, per-participant tones (the pre-TASK-22
`tone_a`/`tone_b` shape still parses), and the rumor paraphrase. Effects land as
ONE atomic `inject_social` batch — turns, summary, and per participant×counterpart
fodder: a gist memory **about** the counterpart (subject-tagged, toned ×30 — a
`TellableFor` gossip seed, stamped `Origin: "gist"` since spec 030 — a
conversation summary is secondhand even to the participant who lived the scene,
per [[agent-mind]]'s provenance doctrine) and a tone edge per pair, reason-tagged
with the first topic; at most one rumor between the founding pair. Since spec 019 (US2) each
gist memory carries two situating fields set directly on the payload: `Where`
(`PlaceAt` on the remembering agent's own tile in the mind replica) and `Conv`
(`cc.conv`, the founding-talk tick that keys every `social.conversation_turn` of
the scene), so the full transcript is recoverable from the memory alone via the
log. The gist TEXT is left unchanged — no where/why clause is spliced into a
conversation memory (unlike executor-emitted memories, [[agent-mind]]); the
`Conv` ref IS its situating, and the scribe renders it as a `· [conv <id>]`
suffix. The scene's terminal
`cog.outcome{landed}` rides the same batch — the scene and its record land
atomically. Landing is also staleness-enforced (TASK-32): a completed scene
whose wall time overran the conversation class's budget in ticks (the router
admitted it, but the provider ran slower than predicted) injects nothing and
records `cog.outcome{rejected-stale}` with the arithmetic. Since spec 067
(TASK-141) that pre-abort compares against the same scaled delivery-gate
predicate as the [[sim-loop]] landing rung — the class's 1x `BudgetTicks`
scaled through `cognition.EffectiveBudgetTicks` by the event-sourced
effective speed, read from a worker-facing atomic tick-rate mirror the
absorb loop keeps current (the same pattern as the mind's `tick` mirror) —
so the mind-side pre-check can never disagree with the reducer gate it
fronts, and the reason carries the scaled-budget derivation. Since TASK-42
(specs/011-conversation-robustness) a scene tolerates one bad reply per site
rather than dying on the first: a parse-failed utterance gets one same-speaker
retry (one utterance retry TOTAL per scene — retry-not-skip, preserving the
round-robin transcript), and a parse-failed outcome call gets one re-request;
each consumed retry emits a non-terminal `cog.outcome{retried}` carrying the
failed reply's verbatim text (`raw`, bounded at 2048 bytes, rune-boundary
truncated), and the scene's terminal event carries `retried: true`. Before
retrying, `parse.go`'s `lenientOutcome` repairs the observed unquoted-gist
shape with no model call at all. Transport/admission errors are NEVER retried
— backpressure stays authoritative — and a second parse failure at either
site abandons: the scene injects nothing and records a terminal
`cog.outcome{unusable}` (with `raw` when the killer was a parse failure); the
primitive talk stands alone. The stale-at-landing check runs after any retry,
so retry wall-time cannot smuggle a stale scene past its budget. The outcome
prompt states that `gist`/`retold` must be double-quoted JSON strings. Replay
is model-free.

## Connections

[[social-fabric]] is the parent note — edges, debt ledger, rumors, secrets,
and theft the scene reads/writes; its "Conversation records" and
"Place-knowledge sidecar" sections cover the last-conversation ring and the
fact-exchange riding the same talk beat. [[executor]] emits the founding
`agent.talked` beat; the [[llm-orchestrator]]'s priority lane keeps dialogue
turns from starving behind planner traffic; [[memory-retrieval]] is the
spec-042 mode gate/divergence telemetry behind the memory snapshot;
[[sim-loop]] owns the sim-side pair cooldown a layer beneath this note's
novelty SHIM ([[sim-state-reducer]]'s `PairTalks`) and the `inject_social`
door; [[cognition]] is the founding/provider-pin router; [[agent-mind]]'s
provenance doctrine underlies the gist memories.

---
name: event-types-social-memory
description: Social/memory-authoring event rows split from [[event-types]]: agent.talked, agent.memory_added, agent.thought, social.* family, social.chest_taken. Load when tracing conversation, situated memory creation/provenance, gossip/theft records, or the spec 061 conversation-loop damper (PairTalks).
kind: concept
sources:
  - internal/sim/agents.go
  - internal/sim/state.go
verified_against: a5df40921577bc194478bb29c42af2b10bf11ea8
---

# Event types — social & memory-authoring events

Back to [[event-types]] for the payload-grammar conventions and the full
event-domain index.


Spec 086 (agent-named payloads): every agent-referencing field in this
family's payloads is a `sim.AgentRef` — the wire carries
`{"id":N,"name":"…"}` objects (lists element-wise), the name stamped at
emission from the fixed roster via `Ref`/`Refs`; sentinels marshal
`{"id":-1,"name":""}`. Legacy bare-int rows decode through the dual-shape
unmarshal forever and reducer arms fold `.ID`s only — the conventions and
the normative back-compat matrix live on [[event-types]] ("Agent
references are named refs").
Spec 061 (the conversation loop damper — [[social-fabric]], [[sim-loop]],
TASK-109) adds no new event type: `State` gains `omitempty` `PairTalks
[]PairTalk` (the event-sourced, unordered per-pair last-exchange ledger, `{a,
b, tick}`), so a pre-061 snapshot with the field absent round-trips
byte-identically. The EXISTING `agent.talked` reducer arm gains a silent
DERIVED write with no new event type (the `markExplored`/`notePresence`
precedent, spec 041): it now also upserts `PairTalks` via `recordPairTalk`,
read back by the sim-side hail cooldown gate (`rungPairCooldown`/
`pairCooled`, [[sim-loop]]) and the mind-side novelty SHIM
(`maybeStartConversation`, [[social-fabric]]). `rebaseTicks` ([[guardian-miracles]])
gains a matching SHIFT arm for every `PairTalks[].Tick` (unconditional — a
present record is always a real exchange tick, never a "never talked"
sentinel, so unlike some anchors it needs no zero-guard).

| Type | Payload struct | Emitted by | Reducer effect |
|---|---|---|---|
| `agent.talked` | `TalkedPayload{a, b}` | executor, adjacent pair (chat-while-working) — also the hail sweep's founding (`hailStep`/`talkEvents`, TASK-47), which the ambient `canTalk` cooldown deliberately never gates | +morale both, talk cooldown (`LastTalk`); since spec 061 (TASK-109) also upserts the unordered `PairTalks` ledger via `recordPairTalk` — the pair-scoped companion to `LastTalk` the conversation loop damper's hail cooldown (`rungPairCooldown`/`pairCooled`, [[sim-loop]]) and mind-side novelty SHIM ([[social-fabric]]) read back; both remember |
| `agent.memory_added` | `MemoryAddedPayload{agent, text, salience, subject, tone, where?, why?, conv?, origin?}` | executor/social/governance/gru heuristics (situated by the acting-or-witnessing agent's tile, and — since spec 030 — stamped with the closed-vocabulary `Origin` class at that same emission site, a required constructor parameter so no new site can compile unstamped); convo gists (injected, `origin: gist`) | append to `Memories`; subject/tone mark gossip seeds; spec 019 (US1) copies the situated context verbatim onto the reduced `Memory` — `where` (`*MemoryPlace{x,y,desc}`, the tile plus a `describePlace`-baked terrain/feature phrase, nil = unsituated), `why` (the driving intent's `Reason`, verbatim; witness memories carry none), `conv` (a conversation ref = founding-talk tick, set by convo gists) — all `omitempty`, so a pre-019 payload reduces to a pre-019-shaped memory (baked at emission, never re-derived, [[agent-mind]], [[social-fabric]]); spec 030 copies `origin` the same way (`omitempty`, absent = `""` = secondhand) — the ONLY signal `DirectPerception` (`internal/sim/memory.go`), and so the belief validator, reads to classify a memory as direct perception (`action`/`witness`/`omen`) vs secondhand (`report`/`gist`/`digest`/absent) |
| `agent.thought` | `ThoughtPayload{agent, text, source}` | `inject_intent` command (planner); `inject_social` (musing) | none (chronicle material) |
| `social.*` family | see `specs/003-social-fabric/contracts/social-events.md` | executor rules, genesis, convo driver (injected) | edges, ledger, rumors, secrets; `social.conversation` appends the bounded record ring (TASK-22, [[social-fabric]]); `social.gave` (spec 013 US1) is additionally skipped by the executor when the receiver has zero free bulk and the reducer clamps defensively (never over `bulkCap`) |
| `social.chest_taken` | `ChestTakenPayload{owner, taker, x, y}` (`social.go`) | executor, same batch as a non-owner `agent.withdrew` (spec 013 US4, FR-011) | none beyond the record itself — the distinct taking happening; chronicle/TUI material ([[social-fabric]]) |

## Connections

[[sim-state-reducer]] owns the
`PairTalks` field, its `agent.talked`-arm write, and the `rebaseTicks` SHIFT
(spec 061);

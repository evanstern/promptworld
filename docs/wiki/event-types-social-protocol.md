---
name: event-types-social-protocol
description: Governance/hail-protocol event rows split from [[event-types]]: meeting.*/norm.* families, meeting.convention_established, sim.gathering_observed, the hail family (social.hailed/hail_met/hail_expired). Load when tracing meeting lifecycle, norm enactment, emergent gatherings, or the TASK-47/spec-061 hail cooldown.
kind: concept
sources:
  - internal/sim/executor.go
  - internal/sim/agents.go
verified_against: 4c66d240b2715706964f02cfd2396256c9957d8e
---

# Event types — governance & hail-protocol events

Back to [[event-types]] for the payload-grammar conventions and the full
event-domain index.

| Type | Payload struct | Emitted by | Reducer effect |
|---|---|---|---|
| `meeting.*` / `norm.*` families (TASK-13) | payload structs in `internal/sim/governance.go`; contract in `specs/006-norms-and-votes/contracts/governance-events.md` | all executor beats (`governanceEvents`) EXCEPT `meeting.proposal_rephrased`, the one injected governance type (mind phrasing driver), and a config-declared `meeting.convention_established`, seeded by the daemon on boot | meeting lifecycle on `State.Meeting`, norms enact/amend/repeal on `State.Norms`, reducer-internal voter/witness edge deltas; rephrase validates (norm exists, text ≤ 280) then swaps text only ([[governance]]) |
| `meeting.convention_established` (TASK-36) | `MeetingConventionPayload{convene_second, open_second, x, y, source}` in `internal/sim/governance.go` | executor emergent-gathering detector (`source: emergent`) or daemon boot seed from `world.json`'s `meeting` block (`source: config`) | one-shot: sets `State.MeetingConvention` (first source wins) and seeds `MeetingPlace`; clears the gathering watch ([[governance]]) |
| `sim.gathering_observed` (TASK-36) | `GatheringObservedPayload{x, y, start}` in `internal/sim/governance.go` | executor per-minute watch while no convention exists (start/break of a sustained gathering; all-zero = reset) | `Meeting.GatherStart/GatherX/GatherY` set, so replay reconstructs the emergent watch |
| hail family (TASK-47): `social.hailed` / `social.hail_met` / `social.hail_expired` | `HailedPayload{from, to, until}` / `HailMetPayload{from, to}` / `HailExpiredPayload{from, to}` in `internal/sim/agents.go`; contract in `specs/010-hail-protocol/contracts/events.md` | loop (`inject_intent` talk_to landing) and executor (`planStepEvents` talk_to firing) emit `hailed`, gated by `hailable` — since spec 061 (TASK-109) additionally requiring the pair NOT be `pairCooled` (the EncounterCooldown dial, [[sim-loop]]'s `rungPairCooldown` gates the SAME predicate one step earlier, refusing the landing outright with an informative reason before any hail attempt) alongside `hailable`'s pre-existing dead/asleep/already-hailed/deadlock/meeting/radius exemptions; the executor's per-tick `hailStep` sweep emits `met` (hailer adjacent AND, since spec 061, the pair still not `pairCooled` — a founding-side backstop closing the leak even for a hail placed legitimately earlier, e.g. an ambient talk raced during the walk-over — accompanied by the `agent.talked` talk shape bypassing the ambient cooldown) or `expired` (window closed) | `hailed` sets `Agent.Hail{By, Until}` (the movement-only pause); `met`/`expired` clear it — `agent.died` and `agent.slept` also clear it. World-emitted only, never model-injectable |

## Connections

[[sim-loop]] owns the `rungPairCooldown` gate and the `set_plan`
clamp decision (spec 061/058);

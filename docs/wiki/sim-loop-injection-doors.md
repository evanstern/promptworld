---
name: sim-loop-injection-doors
description: Loop.InjectSocial (the mind's whitelisted conversation/consolidation/musing/chronicle/nudge/miracle door) and Loop.InjectOperator (the daemon's separate operator-event door); the world/social-effecting whitelist half here, the observability/administrative half split to [[sim-loop-injection-doors-telemetry]]. Load when tracing which event types a non-InjectIntent caller may durably record, or why a type is whitelist-rejected.
kind: component
sources:
  - internal/sim/loop.go
verified_against: 376afd4cee54839a545bc88409f3c485c2f5149d
---

# Sim loop — injection doors

Child of [[sim-loop]] — the two doors besides `InjectIntent`: `InjectSocial`
(the mind's whitelisted recorded-effects door) and `InjectOperator` (the
daemon's separate operator-event door).

## How it works

`Loop.InjectSocial` is the second door — the mind's injection
door ([[social-fabric]], [[nightly-consolidation]], musings per [[agent-mind]],
narrator entries per [[chronicle]], nudges and miracles per [[guardian]] /
[[guardian-miracles]], standing orders per [[guardian-orders]], proposal rephrasing
per [[governance]], place-knowledge per [[mental-maps]] — `agent.thought` is
whitelisted as a reducer no-op, `chronicle.entry` appends the story ring,
`guardian.nudged` spends a charge with a validating reducer the dry-run enforces,
`guardian.place_revealed` (spec 041, FR-014) widens the boundary by one — a
vision's optional place grant, declared in `send_vision`'s `Events` so
`ValidateToolCoverage` pins it ⊆ this whitelist, whose dry-run enforces a
living target and a real place before anything lands — the four
`guardian.time_snapped`/`guardian.item_granted`/`guardian.entity_moved`/
`guardian.entity_removed` miracle types (spec 016) are whitelisted the same way —
their reducer arms enforce presence/destination/charge before anything lands,
the whitelist is only the isolation boundary — `guardian.order_placed`/
`guardian.order_cancelled`/`guardian.order_triggered` (spec 029) join the
whitelist the same way (placement/cancellation/trigger-match validation lives
in the reducer arm); `guardian.order_expired` needs no whitelist entry — it is
executor-emitted, never injected, the `charge_regenerated` precedent — the
four plan-layer types `designation.placed`/`designation.cancelled`/
`directive.issued`/`directive.cancelled` (spec 084, [[guardian-designations]])
join the same way (form/bounds/occupancy/caps/TTL/target validation lives in
the `applyPlan` arms; `directive.issued` rides atomically with per-target
`agent.memory_added` companions), while `designation.fulfilled`/
`directive.fulfilled`/`directive.expired` are executor-emitted and
deliberately absent, the `order_expired` precedent — `prophecy.declared`
(spec 085, [[guardian-faith]]) joins the same way (the `applyProphecy`
door validates targets/text/TTL/cap/claim and spends the charge stake;
it rides atomically with per-target `OriginOmen` companions), while
`prophecy.fulfilled`/`prophecy.failed`/`faith.changed` are
executor-emitted and deliberately absent — whitelist absence is what
refuses a forged verdict or faith movement — and `agent.salience_revised`/
`agent.memory_merged` (spec 098, [[private-dreams]]) widen the boundary by
two: the nightly dream pass's recorded habituation/merge outcomes, injected
by the consolidation driver, reducer-total like the consolidation family —
(since spec 036 whitelist membership is also readable from outside the package
via `InjectableSocialEvent(t)`, the single-source accessor both the tool
coverage gate and the bundle boot gate ([[bundle-tools]]) enforce against).
The observability/administrative half of the whitelist — `cog.*` telemetry,
journal entries, `belief_reinforced`, the spec-042 memory-embedding
companions, `charter_observed`/`morgue.epilogue`/`skills_observed`, and
`guardian.report_card` — splits into
[[sim-loop-injection-doors-telemetry]]. Every whitelisted type, on either
half, lands as part of an atomic, whitelisted batch of conversation,
consolidation, musing, chronicle, nudge, miracle, phrasing, or telemetry
effects, dry-run on a state copy before applying — the dry-run probe is
reconstructed from bytes and so carries no unexported/unserialized state,
so `handleCommand` re-attaches the loop's static map (`probe.SetMap(l.m)`)
before applying, letting miracle arms validate the terrain vocabulary in
the dry-run exactly as the real apply and replay will. Model output enters
the sim only through these two doors, as recorded input. The protocol `Status`
carries `GuardianCharges` (JSON tag `metatron_charges`, frozen — spec 052 ruling 2)
so clients render the ⚡ bank without a state fetch.

`Loop.InjectOperator` (the `inject_operator` command, spec 034 R8) is a THIRD
door, distinct from both above: the daemon's operator-event door, whitelisted
to `daemon.llm_warning` only (`injectOperatorWhitelist`, kept separate from
`injectSocialWhitelist` — one door is the mind's model-output isolation
boundary, the other is the daemon's operator surface, and the two must never
share a whitelist). It exists because `store.AppendEvents` has no internal
locking and the loop is the log's single writer; `daemon.started`/`stopped`
append directly only because they run outside `Run`'s lifetime, but a
provider-health condition transition ([[llm-provider-health]]) fires from
worker/preflight goroutines *while the loop runs*, so its durable event must
ride this command door to keep seq assignment and tick-stamping inside the
loop goroutine. Every whitelisted type is a reducer no-op, so `handleCommand`
skips `InjectSocial`'s dry-run entirely — there is no world-state atomicity to
protect. It fails cleanly (mirroring `InjectSocial`) if the loop has stopped,
letting the daemon's condition hook degrade to a log line only.

## Connections

Parent note: [[sim-loop]]. [[sim-loop-injection-doors-telemetry]] is this
note's own split-off child — the observability/administrative whitelist
half. [[llm-provider-health]]'s condition hook is `InjectOperator`'s sole
caller. [[guardian-miracles]]'s four event types ride `InjectSocial`'s
whitelist, as do [[guardian-orders]]'s three injected order-lifecycle
types and [[mental-maps]]'s `guardian.place_revealed`. [[tool-loop]] is the
caller behind both doors' villager/guardian traffic since spec 017 — its
handlers wrap `InjectIntent` (see [[sim-loop-landing-ladder]]) and
`InjectSocial` (`muse`, and the Guardian's nudges/`work_miracle`), and its
buffered `CallRecord`s land as the `cog.tool_call` batch through this
door. [[morgue]] is the consumer of the narrowed ended-world door (see
[[sim-loop]]'s ended gate). [[bundle-tools]] enforces
`InjectableSocialEvent(t)` as the single-source whitelist accessor
alongside the tool coverage gate.

## Spec 086 — the door validates agent refs before the dry-run

`InjectSocial` gained a decode-and-refuse rail: each whitelisted event's
payload is decoded through `sim.PayloadCatalog` (the sim-side event-type →
zero-payload registry) and `validateRefs` walks the value — an in-roster
`AgentRef` missing its exact roster name, or a sentinel carrying a fake
one, refuses the WHOLE batch before the dry-run. This is the injection
mirror of `mustPayload`'s panic contract on the executor path; both are
live-emission-only rails — replay streams stored bytes and never passes a
door, so legacy unnamed rows are untouched
(`TestDoorRefusesEveryUnnamedWhitelistedType`, `internal/sim/loop.go`).

---
name: sim-loop-injection-doors
description: Loop.InjectSocial (the mind's whitelisted conversation/consolidation/musing/chronicle/nudge/miracle/telemetry door) and Loop.InjectOperator (the daemon's separate operator-event door). Load when tracing which event types a non-InjectIntent caller may durably record, or why a type is whitelist-rejected.
kind: component
sources:
  - internal/sim/loop.go
verified_against: c61cd6c04ddfcd2a976c14a49ba071e8fd768a73
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
`metatron.nudged` spends a charge with a validating reducer the dry-run enforces,
`metatron.place_revealed` (spec 041, FR-014) widens the boundary by one — a
vision's optional place grant, declared in `send_vision`'s `Events` so
`ValidateToolCoverage` pins it ⊆ this whitelist, whose dry-run enforces a
living target and a real place before anything lands — the four
`metatron.time_snapped`/`metatron.item_granted`/`metatron.entity_moved`/
`metatron.entity_removed` miracle types (spec 016) are whitelisted the same way —
their reducer arms enforce presence/destination/charge before anything lands,
the whitelist is only the isolation boundary — `metatron.order_placed`/
`metatron.order_cancelled`/`metatron.order_triggered` (spec 029) join the
whitelist the same way (placement/cancellation/trigger-match validation lives
in the reducer arm); `metatron.order_expired` needs no whitelist entry — it is
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
refuses a forged verdict or faith movement —
(since spec 036 whitelist membership is also readable from outside the package
via `InjectableSocialEvent(t)`, the single-source accessor both the tool
coverage gate and the bundle boot gate ([[bundle-tools]]) enforce against) —
`meeting.proposal_rephrased` swaps
an enacted norm's text and nothing else,
the `cog.*` telemetry — `cog.thought`, `cog.outcome`,
`cog.recalibration_recommended`, and (since spec 017) `cog.tool_call` (the
tool-use loop's per-call trace, [[tool-loop]]) — is whitelisted as reducer
no-ops so the [[cognition]] layer's observability is recorded, never silent,
and (since spec 019, US3) `journal.entry_written`/`journal.entry_deleted` —
the two mind-injectable journal mutations, whose reducer dry-run enforces the
rune budget (written) and entry existence (deleted) before either lands, and
(since spec 030 US2, FR-008) `agent.belief_reinforced` — the
grounded-observation seam that re-anchors a held belief's decay clock; spec 030
ships the whitelist entry and reducer arm only, no in-tree emitter yet), and
(since spec 042 US1/US2) three more: `agent.memory_embedded`/
`agent.situation_embedded` — the mind-side embedder's two vector companions
([[memory-retrieval]]), state-mutating unlike the `cog.*` telemetry (below) —
door ordering guarantees a memory's embedding companion never precedes the
memory itself, since the embedder only observes an `agent.memory_added` AFTER
it is committed and notified; and `cog.memory_divergence`, the shadow-mode
selector's rank-divergence record, riding the same reducer-no-op `cog.*`
isolation class as the telemetry types below), and (since spec 044 US2) two
more: `metatron.charter_observed` — the Guardian turn pipeline's
fingerprint-at-effect stamp, the event-sourced charter-revision timeline the
[[morgue]] aligns deaths against, whose reducer arm (and so the dry-run)
enforces a non-empty fingerprint — and `morgue.epilogue`, the narrator's
recorded mourning prose after a death or the run's end, appending only the
bounded `State.MorgueEpilogues` ring (never simulation state, which is why it
also survives the ended-world narrowing above)), and (since spec 077 FR-006) `metatron.skills_observed` — the
skills-observation twin of `charter_observed`: the bound skill-file set a
turn ran under, emitted on fingerprint change by the same pipeline
(`observeSkills`), whose reducer arm (and so the dry-run) enforces a
non-empty fingerprint AND a non-empty name list (an empty bound set is
never an observation), and (since spec 063,
[[grounded-feedback]]) `guardian.report_card` — the guardian's report-card
producer's stored attribution note, recorded prose only, never simulation
state; a run-ending card rides `morgue.epilogue` instead, so this type
deliberately does NOT join the ended-world narrowing above:
an atomic, whitelisted batch of conversation, consolidation, musing, chronicle,
nudge, miracle, phrasing, or telemetry effects, dry-run on a state copy before
applying — the dry-run probe is reconstructed from bytes and so carries no
unexported/unserialized state, so `handleCommand` re-attaches the loop's static
map (`probe.SetMap(l.m)`) before applying, letting miracle arms validate the
terrain vocabulary in the dry-run exactly as the real apply and replay will.
Model output enters
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

Parent note: [[sim-loop]]. [[llm-provider-health]]'s condition hook is
`InjectOperator`'s sole caller. [[guardian-miracles]]'s four event types ride
`InjectSocial`'s whitelist, as do [[guardian-orders]]'s three injected
order-lifecycle types and [[mental-maps]]'s `metatron.place_revealed`.
[[memory-retrieval]]'s embedder driver injects `agent.memory_embedded`/
`agent.situation_embedded` through this door and records
`cog.memory_divergence` alongside the other `cog.*` telemetry.
[[tool-loop]] is the caller behind both doors' villager/guardian traffic since
spec 017 — its handlers wrap `InjectIntent` (see [[sim-loop-landing-ladder]])
and `InjectSocial` (`muse`, and the Guardian's nudges/`work_miracle`), and
its buffered `CallRecord`s land as the `cog.tool_call` batch through this
door. [[morgue]] is the consumer of the two spec-044 whitelist types and of
the narrowed ended-world door (see [[sim-loop]]'s ended gate). [[grounded-feedback]]
(spec 063) is `guardian.report_card`'s injector, whitelisted here.
[[bundle-tools]] enforces `InjectableSocialEvent(t)` as the single-source
whitelist accessor alongside the tool coverage gate.

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

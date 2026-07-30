---
name: event-types-guardian-plans
description: "Guardian plan-layer event rows split from [[event-types]] (spec 084): designation.placed/cancelled/fulfilled and directive.issued/cancelled/fulfilled/expired. Load when tracing the designation/directive lifecycle, the injected-vs-executor-emitted door split, or the TASK-118 faith seam payload (consumed by spec 085 — [[guardian-faith]])."
kind: concept
sources:
  - internal/sim/plans.go
  - internal/sim/executor.go
  - internal/sim/loop.go
verified_against: 9b4ed5aef5bfea50b67fac10f8e2153f065a814d
---

# Event types — guardian plan-layer events

Back to [[event-types]] for the payload-grammar conventions and the full
event-domain index; the subsystem story is [[guardian-designations]].


Spec 086 (agent-named payloads): every agent-referencing field in this
family's payloads is a `sim.AgentRef` — the wire carries
`{"id":N,"name":"…"}` objects (lists element-wise), the name stamped at
emission from the fixed roster via `Ref`/`Refs`; sentinels marshal
`{"id":-1,"name":""}`. Legacy bare-int rows decode through the dual-shape
unmarshal forever and reducer arms fold `.ID`s only — the conventions and
the normative back-compat matrix live on [[event-types]] ("Agent
references are named refs"). `directive.issued` now rides the wire as the
`DirectiveIssuedPayload` mirror (`Targets []AgentRef`, same tags) while the
state `Directive` keeps `[]int`; `directive.fulfilled`'s targets are named
refs (payload-only type). The arm folds `.ID`s (`internal/sim/plans.go`).
Spec 084 adds **no** format bump: `State` gains
`Designations []Designation` and `Directives []Directive` (both
`omitempty` — a pre-084 snapshot unmarshals to nil, the spec-029
`GuardianOrders` precedent) and SEVEN new event types in two NEW
namespaces (`designation.*`, `directive.*` — world plan artifacts and
villager-facing bindings, deliberately not `guardian.*` console
bookkeeping). The door split is the standing-order one exactly:

| Type | Payload | Door |
|---|---|---|
| `designation.placed` | the full `sim.Designation` entity (`Status`/`PlacedSeq` ignored on the wire — the reducer lands `active` and stamps `PlacedSeq` from `e.Seq`) | injected (`place_designation`); the arm validates id/kind/form/bounds/size/occupancy/label/cap AND fans out the announcement `PlaceFact{Kind:"designation"}` to every living villager's mental map |
| `designation.cancelled` | `{id}` (the `OrderIDPayload` shape) | injected (`cancel_designation`); one-way `active → cancelled` |
| `designation.fulfilled` | `{id}` | EXECUTOR-emitted (`stepEvents` sweep, once, when the kind's structural predicate holds — the `charge_regenerated`/`order_expired` precedent); the arm re-validates the predicate before transitioning |
| `directive.issued` | the full `sim.Directive` entity (`Status`/`PlacedSeq` reducer-stamped) | injected (`issue_directive`), ATOMICALLY with one companion `agent.memory_added` per target ("The Guardian charges you: <text>" — the vision-memory firewall shape); the arm validates active designation, living ascending-unique targets, 1..400-rune text, shared 1..7-day TTL, cap 3 |
| `directive.cancelled` | `{id}` | injected (`cancel_directive`) |
| `directive.fulfilled` | `{id, designation_id, targets, issued_tick}` — **the TASK-118 faith-accounting seam** (dedupe id, what was achieved, who was bound, the time-to-fulfil window vs `e.Tick`) | EXECUTOR-emitted when the bound designation is `fulfilled`; checked before expiry per directive so exactly one terminal lands |
| `directive.expired` | `{id}` | EXECUTOR-emitted when `tick >= ExpiresTick` OR no targeted villager remains alive (the un-executable clause, no TTL wait) |

Whitelist delta ([[sim-loop-injection-doors]]): the four injected types
join `injectSocialWhitelist`; the three executor-emitted types get NO
entry — whitelist absence is what refuses a forged structural fact. None
of the seven joins the ended-world prose whitelist; on an ended world the
sweeps never run (`stepEvents` emits nothing after `run.ended`).

Observable delta ([[tool-registry-guardian-tools]]): all four
`directive.*` types join `observableEventTypes` (12 → 16, enum-only), so
standing orders watch the directive lifecycle through the unmodified
`matchOrders` path — the spec's zero-new-trigger-code guarantee is
structural. Designation events stay out in v1 (watch `directive.*`
instead).

Sweep ordering (fixed, research R14): designations first, then directives,
each in slice order, fulfilled-before-expired within a directive; a
designation fulfilled at tick T yields its bound directives'
`directive.fulfilled` at T+1 — a documented, deterministic one-tick lag.
Renaming or shrinking any of the seven after merge is a log-format break:
since spec 094 a persisted-name change requires a `store.LogFormatVersion`
bump plus the translating migration ([[event-log]]'s doctrine — the
successor to the spec-052 freeze).

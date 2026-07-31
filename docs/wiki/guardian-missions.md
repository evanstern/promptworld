---
name: guardian-missions
description: Guardian missions (spec 107) — the player's plain-words standing instruction as a durable spec-084-shaped artifact: accept_mission/note_mission_progress/cancel_mission, the atomic mission_id link on the plan verbs, derived completion/failure from linked-designation predicates (never self-graded), scheduled-lane pursuit at FULL competence at any ceiling (missions are pre-authorization), and the D5 EASY-mode default-charter obedience clause. Load when tracing guardian.mission_* or the pursuit grant.
kind: component
sources:
  - internal/sim/missions.go
  - internal/guardian/missions.go
  - internal/guardian/ceiling.go
  - internal/persona/charter.go
verified_against: fc1a8314f3f71a33c5e2145c914d5cbb511d9196
---

# Guardian missions — accept, decompose, pursue, report

Spec 107's player-facing loop over the spec-084 plan layer and spec-102
steward lane: "Guardian, get a second fire built near the west huts and
keep it fueled" becomes a durable MISSION the guardian pursues across its
own scheduled watches — no player in the loop — and reports on from
recorded evidence only.

**Doctrine (ratified, not re-litigated):** a mission is durable
PRE-AUTHORIZATION — the standing order's legal shape. The spec-102 ceiling
caps INITIATIVE; a mission is the player's explicit instruction, not
initiative, so pursuit runs at FULL competence at ANY ceiling, exactly
like console orders.

## The entity (internal/sim/missions.go)

`sim.Mission` clones the [[guardian-designations]] discipline verbatim:
deterministic `msn-<tick>-<seq>` ids (guardian-side `nextPlanID`, no RNG),
one-way `active → completed | failed | cancelled` resolved at the reducer
door, wire `Status`/`PlacedSeq` ignored/reducer-stamped, caps validated
not clamped (`GuardianMissionCap` 3 active; goal 1..400 runes; TTL 1..14
game days, tool-door default 7), the shared active+recent-32 prune, and
`State.Missions` `omitempty` (pre-107 snapshots byte-identical, no format
bump). `Designations`/`Directives` link lists accumulate ONLY from
`guardian.mission_progressed` — acceptance must land them empty.

Event family and door split: [[event-types-guardian-plans]] carries the
five-type payload/door table. Injected: `guardian.mission_accepted`
(accept_mission), `guardian.mission_progressed` (note_mission_progress or
a plan verb's `mission_id`), `guardian.mission_cancelled` (cancel_mission
— the player's stand-down). Executor-emitted, never injectable:
`guardian.mission_completed` / `guardian.mission_failed` — whitelist
absence refuses a forged outcome (D3: derived, never self-graded).

## Derived completion and honest failure

`missionFulfilled` is a pure state predicate — ≥1 linked designation and
EVERY linked designation `fulfilled` (the spec-084 structural predicates
do the real checking) — evaluated identically by the `stepEvents` sweep
(fixed slot after the prophecy sweep; completed checked before failed;
the same documented one-tick lag as directives) and by the reducer arm's
re-validation. Failure fires at `DeadlineTick` with the predicate unmet:
reason `deadline_unmet`, or `never_pursued` when nothing was ever linked;
the payload carries every linked entity's status at the deadline —
recorded-event evidence of the blocker, never prose grading. Terminals
surface as model-free player moments (digest.go) and join the
[[guardian-report-card]] citable trail with their seqs.

## Decomposition through existing verbs only (D2)

Pursuit's acting vocabulary is unchanged: survey, designations,
directives, and (grant permitting) workings. The mission verbs record
intent and progress only. The ONE additive hook is the optional
`mission_id` arg on `place_designation`/`issue_directive` (plans.go): the
linking `mission_progressed` rides the SAME atomic batch, so pursuing a
mission never costs a second act and a bad link refuses the whole
placement. The order door keeps its one arbiter — `monitor_and_act` is
NOT on the pursuit surface; a mission causes orders only through a
console/order turn's normal door.

## Scheduled-lane pursuit and the ceiling composition

On a scheduled ([[guardian-agentization]]) turn with an active mission,
`applyMissionPursuitGrant` (ceiling.go) composes BESIDE
`applyAngelCeiling`: it union-adds exactly `missionPursuitTools` (survey,
the four plan verbs, work_miracle, note_mission_progress) — but only
those the WORLD grant already offered — and restores the world's
miracle-kind grant, so a default-ceiling turn pursues at full competence
while the clock triple, orders, nudges, prophecy/canonize, and
accept/cancel_mission stay capped. No active mission ⇒ identity
(byte-identical to spec 102). The turn composes
`guardianAngelMissionFrame` (the modest frame with the pre-authorization
carve-out), the user prompt's active-mission section (id, goal, days
left, per-link status), and the pursuit addendum on the cadence seed.
Pinned by `TestMissionPursuitGrantComposition` +
`TestScheduledMissionPursuitAtDefaultCeiling` (with-mission) beside
`TestDefaultCharterCeilingCapsScheduledRoster` (without).

## The EASY-mode default charter (D5)

`persona.DefaultCharter` carries the obedience clause: direct orders
execute at once, missions are accepted and pursued, no editorializing, no
unrequested confirmation; counsel stays free WHEN ASKED; the ONE
sanctioned refusal is an impossible-as-stated order/mission, named by its
exact blocking fact. The retired counsel-first seed lives on as
`persona.LegacyDefaultCharterCounsel` in `isLegacyDefault`, so an
untouched pre-107 charter.md stays game-authored on upgrade (ceiling ON,
unlock gates honest — the spec-052 SC-003 discipline). The edit is
eval-gated: `docs/design/evidence/task-158/results.md` tabulates the
old-vs-new obedience eval (new default acts same-turn in every class; the
old default's counsel-loop reproduces on judgment-inviting orders,
corroborating TASK-166's recorded 4-turn loop).

Worlds without the steward lane (`steward_cadence_ticks` 0) accept
missions but pursue only on event-driven turns — degraded but honest.

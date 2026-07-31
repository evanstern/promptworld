---
name: tool-registry-plan-faith-tools
description: Child of [[tool-registry-guardian-tools]] — the spec-084 durable plan-layer tools (place_designation/cancel_designation/issue_directive/cancel_directive/survey_site, Gate None, stage-1), the spec-085 prophesy tool (Gate Charge, claim-kind-conditional params, no cancel verb), and spec 107's mission_id pursuit-link param on place_designation/issue_directive (the mission verbs themselves live on [[guardian-missions]]). Load for these tools' exact schemas and event declarations.
kind: component
sources:
  - internal/tool/registry.go
  - internal/tool/derive.go
verified_against: fc1a8314f3f71a33c5e2145c914d5cbb511d9196
---

# Tool registry — plan-layer and prophecy tools

Child of [[tool-registry-guardian-tools]]: the spec-084 durable plan-layer
tools and the spec-085 `prophesy` tool, appended after the agency surface
and miracle/guidance derivations the parent covers.

## How it works

Spec 084 ([[guardian-designations]]) appends the five plan-layer tools
after `explain`, all `Gate: None` (the plan layer is charge-free — recorded
decision): `place_designation` (`kind` Enum over `designationKinds`,
`target` Text — a bare spec-082 locus parsed by `target.ParseLocus`,
`structure_kind` Enum over `buildableStructureKinds` — a hand-carried
mirror of sim's recipes-derived list, drift-pinned from internal/guardian —
`min_structures` Number 1..12, `label` Text ≤80; Events
`designation.placed`), `cancel_designation`/`cancel_directive` (required
`id`), `issue_directive` (`designation_id`, `targets` — the send_omen
comma-names/"everyone" vocabulary — `text` ≤400 runes, `ttl_days` 1..7
default 3; Events `directive.issued` + `agent.memory_added`), and the
Effect-Read `survey_site` (`x`/`y` required, `radius` clamped 1..8 default
4 — renders under `GuardianReadGuidance` like explain). All five join
`RosterGuardian`, `loopGuardianTools`, AND the stage-1 ceiling (the
`monitor_and_act` every-stage teaching-primitive precedent).

Spec 085 ([[guardian-faith]]) appends `prophesy` last: `Gate: Charge` (1 —
the send_vision price; the `prophecy.declared` reducer arm spends the
stake), `targets` (the send_omen vocabulary), `text` ≤400 bytes,
`claim_kind` Enum over `prophecyClaimKinds` (`designation_fulfilled`/
`structure_count`/`population_at_least`/`survives` — exported as
`ProphecyClaimKinds()` and drift-pinned from internal/guardian), the
kind-conditional claim params (`designation_id`/`structure_kind`/`min`/
`agent` — partial or foreign sets refused handler-side), and
`deadline_days` 1..7 default 3; Events `prophecy.declared` +
`agent.memory_added`. It joins both rosters and the stage-1 ceiling
(send_vision's profile — the same influence verb with a wager attached),
and there is deliberately NO cancel verb. `observableEventTypes` grows
enum-only 16 → 19 with the three `prophecy.*` types (`faith.changed`
deliberately stays out in v1).

## Connections

Parent [[tool-registry-guardian-tools]] covers the spec-029 agency surface
and the miracle-cost/guidance derivations these tools also flow through
(`GuardianToolGuidance`, the Read-tool skip). [[guardian-designations]]
owns the designation/directive entity model and event lifecycle these
tools land into; [[guardian-faith]] owns the prophecy lifecycle and charge
economy `prophesy` spends from; [[tool-registry]] is the root note for the
whole registry split.

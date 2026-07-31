---
name: tool-registry-derivation-rosters
description: Split from [[tool-registry]] (spec 014/017/019/058) — derive.go's per-consumer walks (VocabularyLine/PromptGlossBlock/WorldGoals/PlanStepGoals/InputSchema), roster.go's villager/Guardian roster data (RosterVillager/RosterGuardian/LoopRosterVillager/LoopRosterGuardian, the dormant-verb prune), and the boot-time tool.Validate()/sim.ValidateToolCoverage() consistency gates.
kind: component
sources:
  - internal/tool/registry.go
  - internal/tool/derive.go
  - internal/tool/roster.go
  - internal/tool/validate.go
  - internal/sim/toolcheck.go
verified_against: fc1a8314f3f71a33c5e2145c914d5cbb511d9196
---

# Tool registry — derived surfaces, rosters, and validation

Split from [[tool-registry]]: derive.go's consumer walks, roster.go's roster
data, and the boot-time validation gates.

**Derived surfaces** (`derive.go`): each consumer is one walk of the registry —
`VocabularyLine()` (the prompt's goal list, byte-identical to the old constant,
now over the legacy-only filter), `PromptGlossBlock()` (the per-verb gloss
prose — scoped to `isLegacyWorldTool` tools only, since this IS the world-verb
goal prose; the journal tools' glosses, spec 019, are model-facing tool
descriptions delivered per-tool through the loop's `ToolDecl`, not this legacy
prose surface), `WorldGoals()` (the mind parser's accept set), and
`PlanStepGoals()` (the sim door's plan-step accept set) all walk
`legacyWorldNames()`. `InputSchema(t)` (spec 017 data-model.md §1) is the
tool-use loop's new consumer: returns `t.InputSchemaJSON` verbatim when set,
else derives a JSON Schema object from `t.Params` (`paramSchema` per-kind:
`AgentName`/`Text` → string, +`maxLength` from `MaxRunes`/`MaxBytes`; `Enum` →
string with an `enum`; `Number` → integer, +`minimum`/`maximum` from `Min`/
`Max`; every kind then gains `"description"` from `Param.Description` when
non-empty, spec 019 T024) — deterministic output since `Params` is already
registration-ordered and the one Go map in play (`properties`) holds only
property-name keys, which `encoding/json` sorts lexicographically.

**Rosters** (`roster.go`): capability is roster membership, expressed as data.
`RosterVillager` = the legacy world verbs (derived via `isLegacyWorldTool`,
registration order — `set_plan` excluded) + `say`/`muse`/`gist`; `RosterGuardian`
= `converse` plus every acting tool the guardian may use (spec 029:
`send_omen`/`send_vision`/`monitor_and_act`/`cancel_order`/`work_miracle`/
`pause`/`start`/`adjust_speed`; `explain` (spec 063,
[[grounded-feedback]]); specs 084/085/101 append the plan-layer tools,
`prophesy`, and `canonize_region`/`brief_myths` — [[guardian-canonization]])
— it mirrors `LoopRosterGuardian`'s names plus `converse`. Since spec
029 the guardian's nudge/send form is validated against the reducer's explicit
form set, not this roster ([[guardian-orders]]), so `RosterGuardian`'s only live
consumer is the boot-time name-resolution check in `Validate` — kept in step to
keep that gate honest. `OnRoster()` is the door predicate: [[sim-loop]]'s intent
door requires a World tool on the villager roster.
Two new roster exports serve [[tool-loop]] specifically, returning full `Tool`
values (not just names, since `InputSchema` needs `Params`/`InputSchemaJSON`):
`LoopRosterVillager()` = every legacy World tool — MINUS `dormantVillagerVerbs`
(spec 058 US3, TASK-110: `collect_water`, `bathe` — a non-choice today, water
has no consumer and both verbs' world-01 usage collapsed to near-zero; a
revisit-condition comment at the prune site names the trigger: a designed
thirst need) — then `set_plan`, then `muse`
(`say`/`gist` stay scene-gated and out of the loop roster this task — scenes
remain driver-run, not model-initiated), then the four spec-019 journal tools
(`write_journal_entry`, `delete_from_journal`, `search_journal`, `read_journal`
— appended last so no existing declared tool's position shifts). `setPlanTool`'s
own step-goal enum (`registry.go`) is a SEPARATE villager-facing prompt
surface built from the same legacy-World-tool set, so it applies the shared
`pruneDormant()` filter too — without it a model could still offer
`collect_water`/`bathe` as a plan STEP even after they left the declared
roster. `RosterVillager` (the door's name-only membership check) and
`PlanStepGoals()` (the plan-step accept set) are deliberately UNTOUCHED — the
sim executor still honors both verbs so a historical world's `collect_water`/
`bathe` events replay exactly, and reintroduction to the model-facing surfaces
is a roster/gloss edit, not a rebuild;
`LoopRosterGuardian()` = `send_omen`, `send_vision`, `monitor_and_act`,
`cancel_order`, `work_miracle`, the meta tools `pause`/`start`/
`adjust_speed` (spec 029 order), `explain` (spec 063,
[[grounded-feedback]]), the plan-layer tools + `prophesy` (084/085), and
`canonize_region`/`brief_myths` last (spec 101, [[guardian-canonization]])
— deliberately NOT `RosterGuardian`, because `converse` is excluded: it is the
guardian's final-answer channel (the loop's `Result.Final`), not a callable tool,
and declaring it would trap a `converse` call as `rejected_unknown` (the guardian
installs no `converse` handler by design). The journal tools are villager-only
— `LoopRosterGuardian` is untouched, since journals are private.

The registry is deliberately closed: spec 036's bundle tools ([[bundle-tools]])
NEVER enter it. `internal/bundle` synthesizes its own `tool.Tool` values
(Effect Expressive, `PromptGloss` from the manifest) and the guardian turn
assembly appends them to the per-job roster and handler map after
`grantedRoster` — built-ins always win name collisions at bundle load, so no
dynamic registration can perturb the registration-order byte-identity the
derived prompt surfaces depend on.

**Validation** (`validate.go` + `internal/sim/toolcheck.go`): `tool.Validate()`
checks the registry's internal consistency (unique non-empty names, known effect
classes, Events ⇒ Expressive, PlanStep/ReflexEligible only on World tools, Number
params' Min/Max not inverted, a set `InputSchemaJSON` is valid-JSON object shape,
roster names resolve, and — since spec 058 FR-001 — `Clamp` set only on a `Text`
param) and returns ALL violations. Spec 017 lifts the spec-014
restriction barring Read tools from a roster (`tool-loop` is now the Read
consumer; spec 017 itself shipped zero production Read entries, but a roster
naming one was no longer a `Validate` error — spec 019 ships the first two,
`search_journal`/`read_journal`, both on `LoopRosterVillager`). `sim.ValidateToolCoverage()` checks the sim side —
every GOAL-DOOR World tool (Effect World AND PlanStep true — the same
`isLegacyWorldTool` predicate) has a resolver-table entry and a duration, and
every Expressive tool's declared `Events` ⊆ the `InjectSocial` whitelist —
read through `sim.InjectableSocialEvent` (spec 036), the exported membership
accessor that keeps this gate and the bundle boot gate ([[bundle-tools]])
enforcing the SAME whitelist the door does.
`set_plan` is a World tool that deliberately carries `PlanStep: false`, so
`validateCoverage` skips it — it grounds through its own door (`injectPlan`, each
step resolving its own already-covered goal), never through `resolveGoal`/
`goalResolvers`. Both `tool.Validate()` and `sim.ValidateToolCoverage()` run
first thing in [[daemon-lifecycle]]'s `daemon.Run`, before the world opens: a
malformed registry or roster aborts boot with a config error, never a tick-time
failure.

**What derives on the sim side** ([[executor]], [[reflex-policy]]): `intentDuration`
reads a table built from the registry's `Cost.DurationTicks` at init, filtered to
goal-door (World && PlanStep) tools (context overrides — spear-hunt, oven-cook —
stay in the executor's `workDuration`, since the station/inventory is only known
at completion time), and `resolveGoal` is a name-keyed resolver table
(`goalResolvers`) with the old switch arms verbatim. The registry's duration
literals are hand-carried mirrors of the sim constants (R7 — `tool` is a leaf
package that imports nothing internal); `TestWorldToolDurationsMatchSimConstants`
pins the two hand-equal so they can never silently drift.

## Connections

Part of [[tool-registry]]'s split (corpus-spec v2); see it for the registry's
doctrine and the other domains: [[tool-registry-world-catalog]],
[[tool-registry-guardian-tools]], [[tool-registry-schema-clamping]].
[[tool-loop]] consumes `LoopRosterVillager()`/`LoopRosterGuardian()` and
`InputSchema(t)`; [[daemon-lifecycle]] runs the boot gates; [[executor]]/
[[reflex-policy]] are the sim-side consumers the coverage gate cross-checks.

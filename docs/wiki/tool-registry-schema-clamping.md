---
name: tool-registry-schema-clamping
description: Split from [[tool-registry]] (spec 017/029/058) — why set_plan and work_miracle need hand-authored InputSchemaJSON overrides instead of Params derivation, and the spec-058 Clamp mechanism (reasonParam/say/gist/muse/set_plan's reason) truncating an over-cap expressive field instead of rejecting the call.
kind: component
sources:
  - internal/tool/registry.go
  - internal/tool/tool.go
  - internal/tool/roster.go
verified_against: 048259bb42b03cc6ebeb13a49f367c2e3a7d4d37
---

# Tool registry — authored schemas and clamp-with-notice

Split from [[tool-registry]] (spec 014, TASK-53): `set_plan`/`work_miracle`'s
authored schema overrides and the spec-058 clamp-with-notice mechanism.

**`set_plan` and `work_miracle`'s schemas** (`registry.go`): `set_plan` needs an
authored `InputSchemaJSON` override — the registry's scalar `Param` model has no
`ParamKind` for a `steps` array — built by
`setPlanSchema(pruneDormant(legacyWorldNamesFrom(worldTools)))` (the
`pruneDormant` wrap is spec 058 US3, above):
a `steps` array of `{goal, kind, qty}` objects — no longer capped by a
declared `maxItems` (spec 058 FR-003: an oversized array now clamps at the
landing guard instead of failing the driver's structural walk; `PlanStepCap`
(3) survives only as `description` guidance) — `goal`'s
enum drawn from the SAME legacy-World-tool filter `VocabularyLine`/`WorldGoals` use,
so the plan vocabulary can never drift from the free-text one even though the two
can't share one function call (an initialization-cycle constraint). `work_miracle`
needs no override: its flat parameter surface (`kind` required Enum over
`miracleKinds` = `move`/`remove`/`give_item`/`time_snap`, plus every per-kind field
as an optional scalar) is fully Params-derived. (Pre-spec-029 this was also
load-bearing: the loop driver's `validateArgs` routed every `InputSchemaJSON`
tool through `set_plan`'s validator, so an override here would have validated
`work_miracle` against the wrong shape. Spec 029 replaced that dispatch with a
general schema-lite walker that validates each authored tool against ITS OWN
schema ([[tool-loop]]), so an override would no longer mis-validate — but Params
derivation stays the right choice regardless, keeping `InputSchema` (what the
model sees) and `validateArgs` (what the driver enforces) sourced from the same
`Params`.) `work_miracle` is `Effect Expressive` (not `World`): it
lands a bounded event batch through the SAME `InjectSocial` door the nudges use
(`Guardian.landMiracle` → `BuildMiracleBatch`), has no intent and no work duration,
and — decisively — `Validate` forbids a World tool from declaring `Events`, which
`work_miracle` must (so the sim-side coverage check can pin its event set ⊆ the
whitelist). There is deliberately no `gratis` parameter: the angel can never waive
a charge (spec 016 FR-007/SC-005) — structural absence, not a sanitized field.

**Clamp-with-notice (spec 058 FR-001, TASK-110)**: the four EXPRESSIVE text
surfaces — `reasonParam()`'s shared `reason` (every acting villager world tool
and `set_plan`), and `expressiveTools`' `say.text`/`gist.gist`/`muse.text` —
carry `Clamp: true`, so an over-cap value truncates rune-safely at the
[[tool-loop]] driver instead of the whole call being rejected (world-01
diagnosis: ~93% of 807 rejections were exactly this shape). `say`/`gist` don't
ride the villager tool-use loop today (scene-gated, `roster.go`) so the
driver's `Text` arm never actually sees their `Clamp` flag — their real
enforcement is the conversation scene parser (`internal/mind/parse.go`
`parseSay`/`parseOutcome`), which already truncated rather than rejected and
this task made rune-safe (`toolloop.ClampBytes`, fixing a latent multi-byte
UTF-8 split at the byte-cap boundary); the registry `Param` still carries
`Clamp` regardless, keeping it the single source of truth for which fields are
expressive. `set_plan`'s schema drops `maxItems` from its `steps` array
(`setPlanSchema`) — an oversized plan no longer fails the driver's structural
walk at all, reaching the landing guard ([[sim-loop]]) instead, which clamps
to the first `PlanStepCap` steps with a notice; the cap survives only as
prose guidance in the schema's `description`. `set_plan`'s top-level `reason`
(the one clampable field that ISN'T a `Param` — its authored
`InputSchemaJSON` override bypasses `Params` derivation) is clamped by the
[[tool-loop]] driver keyed on the field name, not a `Clamp` flag. Prose
glosses that advertised the pruned verbs below move in the same change:
`glossQuarry` drops its `collect_water` clause and `glossBuildOven` drops its
`bathe` clause.

## Connections

Part of [[tool-registry]]'s summary-style split (corpus-spec v2). See
[[tool-registry]] for the registry's overall doctrine and its other split-off
domains: [[tool-registry-world-catalog]] (the World/villager catalog),
[[tool-registry-guardian-tools]] (the Guardian agency surface and miracle
costs), and [[tool-registry-derivation-rosters]] (derive.go/roster.go/
validate.go). [[tool-loop]] owns the spec-058 `Clamp` enforcement
(`validateArgs`'s clamp-with-notice, `VerdictLandedClamped`) that consumes
this note's `Clamp`-flagged params; [[sim-loop]]'s landing guard owns the
matching `set_plan` step-count clamp; [[agent-mind]]'s scene parser
(`parse.go`) is where `say`/`gist`'s own `Clamp`-flagged fields actually
enforce, since both stay outside the villager tool-use loop.

# Data Model: The curriculum ladder

**Spec**: [spec.md](./spec.md) | **Plan**: [plan.md](./plan.md) | **Date**: 2026-07-25

## Manifest fields (internal/world/world.go — additive omitempty, no format bump)

| Field | Type | Semantics |
|---|---|---|
| `Stage` | `string, omitempty` | `stage-1..stage-4`; absent = pre-ladder = ungated (stage-4 semantics). Validated at `world.Open` (closed vocabulary). Write-once at creation; no mutation command exists. |
| `StageOverridden` | `bool, omitempty` | True when the world was created at an unearned stage via `--override`; makes overridden runs honestly comparable. |
| `CharterPreset` | `string, omitempty` | `""`/`default` \| `tutor`. Names the constant that is the stage-1 effective charter and the seed content at genesis. |

## The stage ceiling (internal/metatron — data, intersected like bundle narrowing)

| Stage | Tools ceiling | Instruction surface |
|---|---|---|
| `stage-1` | base conversational + basic query/nudge set (exact roster names pinned in contracts/stage-gating.md) | charter locked to preset constant; skills not composed; honest notice on player edits |
| `stage-2` | stage-1 + charter-scoped surface (unchanged tool set unless contracts say otherwise) | charter edits bind (today's behavior) |
| `stage-3` | + skill files + player-grantable tool manifest (full spec-021 behavior) | skills compose; capabilities.json honored (may narrow, never exceed ceiling) |
| `stage-4` | full roster incl. capstone capabilities | everything |

Application: `grant = intersect(loadManifest(...), stageCeiling(manifest.Stage))` at
both load sites (turn + status) **before** `grantedRoster` — declaration, prose, and
door layers inherit; structural absence (FR-004) is automatic.

## Events (internal/sim/curriculum.go — executor emission class, no whitelists)

| Type | Payload | Reducer |
|---|---|---|
| `curriculum.exercise_passed` | `{exercise string, stage string, tick int64, evidence []EvidenceRef}` — outcome-shaped; emitted by TASK-119's rubric machinery (fixture-emitted in tests until then) | records pass on state (bounded; enables replay-derived audit) |
| `curriculum.stage_unlocked` | `{stage string, exercise string, tick int64}` | latches unlocked stage on state (per-world fact; the per-user record is derived, never authoritative) |

`EvidenceRef` = `{type string, seq int64, tick int64}` — e.g. the
`metatron.charter_observed` fingerprint event and the rubric's satisfying events.
Catalog obligations: `familyByNamespace["curriculum"]`, digestRegistry rows, fixture
rows, `docs/wiki/event-types.md` rows (TestCatalogSweep enforces).

## Per-user unlocks record (`~/.promptworld/unlocks.json`, internal/worlds/unlocks.go)

```json
{
  "unlocks": {
    "stage-2": {
      "world": "demo-046",
      "path": "/Users/.../worlds/demo-046",
      "exercise": "first-night",
      "evidence": [{"type": "curriculum.exercise_passed", "seq": 4812, "tick": 86400}],
      "earned_at": "2026-07-25T18:00:00Z"
    }
  }
}
```

Doctrine (= worlds registry): missing/corrupt → empty, never an error; `.tmp` +
`os.Rename` writes; **advisory, never authority** — the named world's history is the
proof; deletion loses convenience, not truth (FR-008). Home-dir unresolvable →
warn-and-continue (lease precedent).

## Exercise definition (content contract; consumable by TASK-119)

`{id, stage, seed, framing, rubric (event-derived terms), pass_signal, score_narrative}`
— two shipped instances (*first-night*, *the-law*) specified in
contracts/exercises.md; reserved additive Manifest block shape follows the
`Meeting *MeetingConfig` precedent.

## Skin stub (internal/skin — TASK-121 absorbs)

`stageIdentity[stageID] = {Name, Line string}` — guardian defaults (client-approved):
The Voice / The Written Word / The Craft / The Stewardship. Consumers: `promptworld
stages`, `new` messaging, status/TUI lines, quickstart page titles. Stage ids and
semantics are skin-invariant.

## Status surfaces (additive omitempty)

`ipc.WorldStatus`: `Stage`, `StageOverridden`. Metatron `Status`: stage lock
provenance (stage-1). CLI: stage line via the posture-line precedent; `stages`
command renders the identity table + earned state. TUI metatron pane: stage +
granted-surface line.

## Invariants

- Same seed + same commands ⇒ identical world-event history at every stage (FR-006);
  only agent-surface facts differ. Enforced by a cross-stage determinism diff test.
- A `capabilities.json` may narrow within the ceiling, never exceed it (intersection).
- `curriculum.stage_unlocked` for stage N requires a recorded pass whose evidence
  satisfies the gate conjuncts (stage-2: player-authored charter fingerprint in
  force); a default-charter pass MUST NOT produce the unlock (SC-004).
- Unlocks record entries are re-derivable from the named world's history alone.

# Data Model: Skinnable guardian persona (spec 052)

## Skin (value type, `internal/skin`)

| Field | Type | Constraints |
|---|---|---|
| Name | string | single line, ≤40 runes, no control chars; default `Guardian` |
| Epithet | string | single line, ≤20 runes; default `guardian` |
| TabLabel | string | single line, ≤20 runes; default `guardian` |
| Voice | string | ≤4,000 chars (bundle-SOUL cap); empty = no fragment |
| Strings | map[token]string | overrides; unknown tokens ignored+notice |
| Stages | map[stageID]{Name, Line} | per neutral id; fallback to default |

**Validation**: field-wise — an invalid field falls back to the default
value with one notice; the bundle never fails wholesale for one bad field
(capabilities.json discipline). **Lifecycle**: loaded once at daemon boot
(`Load(worldDir)`), injected via `SetSkin` (boot-frozen), surfaced through
status; the TUI/CLI hold only what status carries.

## Default skin table

Compiled-in map[token]string — the complete token authority (contract §3).
Invariants: every token consumed anywhere in the codebase exists here
(token-completeness test); systems/telemetry surfaces have no tokens (D10,
unskinnable by construction).

## Token

Dotted path string (`skin.guardian.name`, `skin.stage.stage-2.name`).
Resolution: world override → default table → the token path itself (visible
failure). Shared identity with the design corpus's `{{skin.*}}` convention.

## Frozen-vocabulary sets (compat contract, research R4)

Not data at runtime — a normative list enforced by annotations at definition
sites and by the compat tests (pre-feature world replay; old CLI/config
loading). Any PR renaming a member fails SC-003's tests.

## Status additions (wire, additive omitempty)

`skin_name`, `skin_epithet`, `skin_tab_label`, `skin_family_label`,
`skin_stages` (id → name/line). Absent (old daemon) → clients render the
compiled default skin. No other wire changes.

## State transitions

None — skin has no runtime state machine. The one temporal rule: boot-frozen
(edits require restart), reported via status so staleness is diagnosable.

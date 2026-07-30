---
name: guardian-runtime-facts
description: Observable runtime facts about a running guardian — the event-sourced charter-fingerprint revision timeline, the charge bank's regeneration/spend accounting, the two on-disk files (charter.md, metatron/soul.md and transcript.md), and the IPC/CLI/TUI surfaces (Status peek, granted_tools, orders). Split from [[guardian]]; load for charter provenance, charge accounting, or the guardian's external doors.
kind: component
sources:
  - internal/guardian/charter.go
  - internal/sim/guardian.go
  - internal/guardian/guardian.go
verified_against: cf65debb44c1e17b54c0f3421d11e1e8cc28576c
---

# Guardian's runtime facts

Split from [[guardian]] (summary-style, corpus-spec v2) — the charter
revision timeline, the charge bank, the on-disk files, and the external
surfaces a running guardian exposes.

## How it works

**Charter observation — the revision timeline** (spec 044 US2, FR-008): every
`runTurn` stamps the charter revision it actually runs under. Immediately
after `loadCharter` returns — before anything consumes the text —
`observeCharter` (`charter.go`) fingerprints the EFFECTIVE charter
(`charterFingerprint`: the first 12 hex chars of SHA-256 over exactly the
post-fallback, post-truncation bytes the model executes, so the recorded
revision can never name a charter the guardian never ran) and, when it differs
from the last recorded value, emits `guardian.charter_observed{fingerprint,
default}` through the same `InjectSocial` door as every other turn effect —
fingerprint-at-effect semantics. The `default` flag derives from the same
effective text (an empty/missing `charter.md` serves and records the default),
so the two can never disagree — and since spec 046 it is PRESET-AWARE: default
means the effective text equals the WORLD's preset constant (`presetCharter`,
the same reference `charterIsDefault` compares against), not bare
`persona.DefaultCharter`, so a stage-1 tutor-preset world — whose lock serves
`persona.TutorCharter` — honestly records `default: true`. Preset text is
authored by the game, never the player, so it must never masquerade as
player-authored evidence: the [[morgue]]'s alignment and the stage-2→3 unlock
gate's custom-charter evidence both derive from this flag
([[curriculum-ladder]]). At stage-1 the observation stamps the stage-EFFECTIVE
text the lock serves, never the raw file. The first turn of a world always emits (the
mirror starts empty); an ENDED world skips — the door narrows to recorded
prose after run end and a finished run's evidence timeline is closed. The
`the guardian` struct mirrors `State.CharterFingerprint`/`State.Ended` as
`charterFP`/`ended` under `stateMu`; `observeCharter` sets the mirror
optimistically after a landed emission (so a back-to-back turn cannot
double-emit before absorb catches up), and `mirrorState` moves it FORWARD
only — a batch predating the landed observation absorbs with the replica's
old fingerprint, and overwriting would re-open the emission window. The
reducer arm (`internal/sim/guardian.go`, `CharterObservedPayload`) keeps only
the CURRENT fingerprint on state (rejecting an empty one at the dry-run); the
full revision timeline lives in the event log, where the [[morgue]]'s render
scan aligns each death against the most recent observation at or before it.
Evidence only — the payload carries no scoring fields, by contract.

**Charge economy** (`internal/sim/guardian.go`): `State.GuardianCharges` — genesis
1, cap 3, +1 per absolute boundary of the faith-band cadence emitted by the [[executor]]
(`guardian.charge_regenerated`, a pure function of (faith score, scenario
presence, tick) since spec 085 — `FaithRegenCadenceTicks`, steady band = the
old 6 game hours; [[guardian-faith]]), −1 per landed
omen, vision, or prophecy (a miracle spends its per-kind cost). Fully event-sourced: replay
reproduces the bank exactly; the field is
deliberately not `omitempty` so a spent-to-zero bank survives snapshots; pre-TASK-12
snapshots gain the genesis charge on upgrade.

**Files** (bound to the run, not event-sourced): `charter.md` at the save-dir root
(seeded by `persona.Genesis` — since spec 046 with an optional preset parameter,
so a `"tutor"` world seeds `persona.TutorCharter` — never overwritten), plus the
optional player-created
`skills/` dir and `capabilities.json` manifest beside it (spec 021 — root =
player-authored configuration); `metatron/soul.md` (accreting notes, starts empty)
and `metatron/transcript.md` (console history) — restart survival comes free with
files, and world determinism never depends on them (prompt composition is upstream
of the recorded LLM inputs).

**Surfaces**: IPC `metatron_chat`/`metatron_status` (FROZEN wire method names,
spec 052 ruling 2 — [[ipc-protocol]]), CLI
`promptworld guardian <dir> [message…]` (canonical since spec 052 FR-008; the
pre-052 `metatron` name survives as a hidden compat alias — [[cli-promptworld]]), and the
[[tui-client]] guardian pane (the console). Protocol status (`guardian.Status`, the
model-free peek, computed fresh from disk per call) carries the ⚡ bank, charter
provenance (`charter_default`), and since spec 021 the effective skill filenames
(`skills`, composition order), the granted roster (`granted_tools`, registry order,
`work_miracle(move,give_item)` form when kinds are restricted),
`manifest_default` (no `capabilities.json` present), and since spec 029 the active
standing orders (`orders`, `OrderStatus{id, condition, origin, fuzzy, expires_day,
status}`, FR-016 — see [[guardian-orders]]). Since spec 046 the status is the
turn's stage twin: `Status()` applies the same `applyStageCeiling` to the
peeked grant (so `granted_tools` can never disagree with the roster the next
turn will run under), nils the `skills` list below stage-3 (it is the EFFECTIVE
composition list), and carries additive omitempty curriculum provenance —
`stage`, `charter_locked` (the stage-1 lock is in force), `charter_preset` (the
binding preset name when locked, `"default"` | `"tutor"`), and `skills_locked`
(stage-1/-2); `charter_default` compares against the world's preset constant
via the preset-aware `charterIsDefault`.

## Connections

[[guardian]] is the parent. [[morgue]]'s death-alignment render scan reads
the charter-fingerprint revision timeline this note documents; the
`default` flag's preset-awareness is shared with
[[guardian-instruction-surface]]'s stage-1 charter lock.
[[curriculum-ladder]] owns the stage/preset facts (`Manifest.Stage`,
`Manifest.CharterPreset`) this note's status peek mirrors. [[executor]]
regenerates the charge bank on the faith-band cadence boundary
([[guardian-faith]]). [[ipc-protocol]]
freezes the `metatron_chat`/`metatron_status` wire method names;
[[cli-promptworld]] documents the `promptworld guardian` CLI surface;
[[tui-client]] hosts the guardian pane. [[skin]] is the boot-frozen display
substrate the surfaces and files never leak into recorded payloads.

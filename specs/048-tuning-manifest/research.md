# Phase 0 Research: World Tuning Manifest

All Technical Context fields were resolvable from the codebase; no external
research needed. Decisions below resolve the design questions the spec left to
planning, each grounded in an existing pattern.

## R1 — Where tuning state lives: a pointer field on `sim.State`

**Decision**: `Tuning *TuningState` on `sim.State`, tagged
`json:"tuning,omitempty"`; `nil` means "all defaults" and is the state of every
pre-048 world.

**Rationale**: the reducer and the mind replica both need the values, and the
mind already maintains a full `*sim.State` replica that absorbs every event
(`mind.go:62,217`) — putting tuning in state means the mind gets it with zero
new plumbing. The pointer+omitempty shape is the established
snapshot-compatibility pattern: `MeetingConvention` (`state.go:92`) and the
spec-044 `Deaths`/`Ended`/`RunEnd` trio (`state.go:96-106`, "omitempty …
keeps every pre-044 snapshot byte-identical — no format_version bump").

**Alternatives considered**: (a) package-level mutable vars set at boot —
rejected: invisible to replay, breaks determinism, and races with the mind's
goroutines; (b) plumbing values through constructors into mind/daemon —
rejected: two consumption paths to keep in sync, and replay tooling
(`recoverState`, embed-replay) would need separate wiring.

## R2 — Event discipline: one full-set `sim.tuning_applied` event, seeded at boot

**Decision**: a `sim.tuning_applied` event whose payload is the complete
effective dial set. The daemon seeds it at boot only when the manifest's
effective values differ from what state already carries (`nil` state compares
as the default set). Reducer arm sets `s.Tuning` from the payload.

**Rationale**: exact copy of the `seedMeetingConvention` shape
(`daemon.go:481-500`): build event → `state.Apply(ev)` → `st.AppendEvents` —
it "lands in the log like genesis, so replay re-applies it and this boot-time
seed never fires twice." Full-set (not delta) payloads let replay establish
tuning from any single event without scanning history. The no-event-when-equal
guard keeps restart from growing the log (spec FR-004).

**Alternatives considered**: (a) storing tuning only in the manifest and
re-reading at replay — rejected: replay must be file-independent (spec FR-005);
(b) delta events — rejected: forces historical scans and complicates the
reducer for zero benefit at 5 fields.

## R3 — Validation split: clamp values, reject structure

**Decision**: two failure classes. Out-of-range values of known fields clamp
to documented bounds with an operator-visible warning (returned as strings,
printed by the daemon — the `normalizeTokenBudget` shape,
`llm/config.go:83-105`). Malformed JSON, wrong types, or unknown field names
fail boot with the path and problem (decode with
`json.Decoder.DisallowUnknownFields`).

**Rationale**: clamping matches the named precedent
(`max_tokens`/`loop_max_rounds`) and the spec's "clamp-don't-reject on values".
Unknown-field rejection is typo protection: a silently ignored misspelled dial
defeats the manifest's purpose (spec Assumptions; FR-003).

**Alternatives considered**: warn-and-ignore unknown fields (llm.json's looser
posture) — rejected per spec decision: tuning dials change replay-affecting
doctrine, so a no-op typo is worse here than a boot failure.

## R4 — Consumption: accessor methods on `State`, defaults as renamed consts

**Decision**: the five constants become `default*` consts in their current
files; `State` gains nil-safe accessors (e.g. `s.RefuelDyingBelow()`,
`s.FireBurnPerWood()`, `s.GruEmergePerMille()`, `s.PlannerCadence()`,
`s.EncounterCooldown()`) returning the tuned value or the default. Call sites:

| Dial | Default (unchanged) | Call sites to convert |
|---|---|---|
| refuelDyingBelow | 3600 (`agents.go:705`) | `policy.go:147` |
| fireBurnPerWood | 14400 (`agents.go:709`) | `executor.go:832`, `state.go:915` |
| gruEmergePerMille | 600 (`gru.go:46`) | `gru.go:235` (+ mirrored check in `gru_test.go:43`) |
| PlannerCadenceTicks | 1800 (`agents.go:671`) | `mind/mind.go:176,384,432`, `mind/embedder.go:156,212`, `daemon.go:341` |
| encounterCooldownTicks | 7200 (`mind/mind.go:36`) | `mind/mind.go:317` |

Every reducer call site already holds `*State`; every mind call site already
holds `md.replica` / `e.replica`.

**Rationale**: accessors keep nil-handling in one place and make "absent file
== current constants" structurally true. `PlannerCadenceTicks` is currently an
exported `sim` const consumed cross-package — the accessor keeps sim as the
doctrine owner while letting mind read the tuned value. The encounter-cooldown
default moves to sim (it becomes doctrine data carried in sim state), matching
the report's treatment of all five as doctrine dials.

**Alternatives considered**: keeping `PlannerCadenceTicks` a const and special-
casing it via daemon plumbing — rejected: two mechanisms for one feature.

**Amendment (implementation finding, ratified at review)**: the call-site table
above was incomplete — `internal/sim/memory.go:414` (serendipity) and
`internal/sim/social.go:101` (secret-share) also reference the cadence period,
but as an RNG bucket width, not as planner scheduling. Both are deliberately
**pinned to `defaultPlannerCadenceTicks`** (unconverted): FR-006 scopes the
dial to planner scheduling/stagger/embedder buckets, and a cadence tune must
not silently reshape unrelated deterministic RNG streams. If evidence ever
shows those windows should track the dial, that is a new promotion decision,
not a byproduct of this one.

**Note on cadence semantics**: mind's stagger/bucket math derives from cadence
at construction (`mind.go:176`, `embedder.go:156`); since tuning applies only
at boot and the mind is rebuilt at boot, a tuned cadence is constant for the
process lifetime — no mid-run bucket-edge hazards.

## R5 — File handling: `world.TuningPath()`, daemon loads, sim parses

**Decision**: `internal/world` gains `TuningPath()` (the `CalibrationPath()`
one-liner pattern, `world.go:311`). The daemon reads the file at boot (absent
→ skip silently); `sim.ParseTuning(bytes)` returns `(*TuningState, warnings,
error)`; the daemon prints warnings, fails boot on error, then seeds per R2.

**Rationale**: keeps I/O in the daemon and doctrine in sim, mirroring how
`llm.json` (daemon loads, llm package parses/clamps) and `calibration.json`
(daemon loads, cognition parses) already divide labor.

## R6 — Interaction with `fireFuelCap`

**Decision**: `fireFuelCap` is NOT promoted; a tuned `fireBurnPerWood` remains
subject to the existing cap truncation (`executor.go:833`). The clamp range for
`fireBurnPerWood` documents this interplay.

**Rationale**: §6 doctrine — "dials should be earned by evidence, not
speculatively multiplied." Nobody has needed to touch the cap.

## R7 — Clamp ranges

**Decision** (recorded in contracts/tuning.md; defaults in parentheses):

| Field | Min | Default | Max |
|---|---|---|---|
| `refuel_dying_below` | 0 | 3600 | 86400 (24 game-h) |
| `fire_burn_per_wood` | 600 | 14400 | 86400 |
| `gru_emerge_per_mille` | 0 | 600 | 1000 |
| `planner_cadence_ticks` | 60 | 1800 | 86400 |
| `encounter_cooldown_ticks` | 0 | 7200 | 86400 |

**Rationale**: mins guard divide-by-zero/degenerate schedules
(`planner_cadence_ticks` divides in stagger math and bucket math, so min 60 =
one game-minute; `fire_burn_per_wood` min 600 keeps a refuel meaningful);
maxes bound everything at one game-day, wide enough for any dialing world-01
evidence suggests. `gru_emerge_per_mille` is a probability in per-mille — hard
range [0,1000], where 0 disables emergence entirely (a legitimate dial
position: gru-free worlds).

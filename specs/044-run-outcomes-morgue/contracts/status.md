# Contract: run-end status surfaces & postmortem posture

All additions are additive `omitempty` fields — pre-feature worlds and old clients are
byte-compatible (the governor-trio precedent on `ClockStatus`).

## IPC (`internal/ipc/protocol.go`)

- `ClockStatus.Ended bool, omitempty` — true once the run has ended; populated by
  `Loop.status()` from `State.Ended`.
- `ClockStatus.EndedDay int, omitempty` (or equivalent) — game day of the run end, for
  human rendering without a state fetch.
- `StateData` — no change needed: `State.Ended` / `State.RunEnd` ride the canonical
  state JSON automatically. This is the machine-readable surface TASK-119's scenario
  machinery consumes (FR-005), together with the durable `run.ended` event.

## Live push

Attached clients receive `run.ended` through the existing `subscribe` fan-out verbatim —
no new plumbing, no reconnect (spec edge case). The TUI folds it via `applyEvent` into
its replica.

## Command gating (ended world)

| Command | Behavior when `State.Ended` |
|---|---|
| `status`, `state`, `subscribe` | serve normally |
| `pause`, `resume`, `set_speed` | refused with an explicit "run has ended" error |
| `govern`, `inject_intent`, `inject_social`* | refused (world-mutating) |
| `shutdown` | serves normally (daemon lifecycle ≠ run lifecycle) |

*Exception: `morgue.epilogue` / `chronicle.entry` narrator landings are recorded prose
about the ended run and remain accepted — they mutate no simulation state, only the
bounded prose rings. (The executor guard guarantees no sim event follows `run.ended`.)

## CLI (`promptworld status`)

- Human: an ended posture line (e.g. `run ended — day 14, all villagers dead; world is
  an archive (read-only)`), replacing the clock-running line.
- `--json` (live and offline snapshot path): `ended` + summary fields mirrored from
  `ClockStatus` / state so a stopped ended world reports identically.

## TUI postmortem posture

- **Derivation (dual-source, both required)**: (1) replica `State.Ended` — covers
  clients attaching after the fact (snapshot path never replays folded events);
  (2) pushed `run.ended` / the 1s status poll — covers the live transition.
- **Rendering**: header state token `ENDED` replaces the running/`PAUSED` token; clock
  controls (space, `[`, `]`) become inert with a footer hint; chronicle/digest/villagers/
  morgue reading surfaces stay fully functional. Digest rows exist for `run.ended` and
  `morgue.epilogue` (catalog + `digestRegistry`, enforced by `TestCatalogSweep`).

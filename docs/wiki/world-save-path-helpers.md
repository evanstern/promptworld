---
name: world-save-path-helpers
description: internal/world's path-helper catalog — every well-known file a save directory owns (world.db, llm.json, calibration.json, estimator_state.json, sockets/pidfile/log, charter.md, metatron/, village_charter.md, morgue.md, bundles/, tuning.json) and the runtime-only files swept between daemon runs
kind: component
sources:
  - internal/world/world.go
verified_against: b6a20eaa4da1073a69959a5aff69591d931103a9
---

# World save directory: path helpers

Split from [[world-save-directory]] (summary-style, corpus-spec v2): the
file-by-file catalog of everything a save directory can contain, centralized
as one set of path helpers so no caller hand-builds a filename.

- Path helpers centralize layout: `DBPath()` → `world.db`, `LLMConfigPath()` →
  `llm.json` (the [[llm-orchestrator]] config, written by `new`, deletable to
  disable inference), `CalibrationPath()` → `calibration.json` (the
  seconds-per-point profile written only by `promptworld calibrate` —
  [[cognition]]; an absent file is legal, pessimistic bootstrap defaults apply),
  `EstimatorStatePath()` → `estimator_state.json` (the daemon-written snapshot
  of live latency estimates, TASK-113 — absent is legal, boot then seeds from
  calibration/bootstrap alone; never event-sourced, never read during replay),
  `SockPath()` → `daemon.sock`, `PidPath()` → `daemon.pid` (since TASK-147, thin
  `*World` wrappers around the package-level `SockPathIn(dir)`/`PidPathIn(dir)` —
  pure path joins, not a validating `Open` — so daemon-lifecycle callers that must
  reach a world this build cannot necessarily `Open` (`daemon.IsRunning`, `stop`,
  `status`'s live-dial, [[daemon-lifecycle]]) can get the socket/pid path without
  one; world-content commands keep going through `Open` and the `*World` methods),
  `LogPath()` → `daemon.log`, `CharterPath()` → `charter.md` (the player-editable
  prompt), `GuardianDir()` → `metatron/` (dir name frozen, spec 052 ruling 2 — the Guardian's soul and transcript —
  [[guardian]]), and `VillageCharterPath()` → `village_charter.md` (the village's
  scribe-rendered law, deliberately distinct from the Guardian's charter —
  [[governance]], TASK-13), `MorguePath()` → `morgue.md` (spec 044: the run's
  accumulating legacy document — one factual epitaph per death plus a run-end
  summary, scribe-rendered; a regenerable view over the event history, never a
  source of truth, exactly like the chronicle and village charter —
  [[morgue]]), and `BundlesDir()` → `bundles/` (spec 036: the
  drop-in persona/tool bundle root, discovered and boot-frozen by
  [[bundle-tools]]; absent means no bundles, never an error), and
  `TuningPath()` → `tuning.json` (spec 048: the optional, operator-
  authored, sparse per-world tuning manifest promoting doctrine constants
  to per-world dials — absent means every dial keeps its doctrine-constant
  default, exactly today's behavior; never written by `promptworld new`;
  [[world-tuning]] has the full mechanism).

Runtime files (`daemon.sock`, `daemon.pid`) exist only while a daemon runs and are
swept by [[daemon-lifecycle]] when stale. The full layout is documented in
`specs/001-world-daemon/contracts/storage.md`.

The spec-076 fork ceremony ([[world-forking]]) consumes this catalog as its
copy/skip table: player input and per-world profiles copy into a fork
(`llm.json`, `calibration.json`, `estimator_state.json`, `charter.md`,
`tuning.json`, `metatron/`, `bundles/`, `agents/`); runtime files, migration
archives, and the scribe's regenerable views (`chronicle.md`, `morgue.md`,
`village_charter.md`) stay behind.

## Connections

Back to [[world-save-directory]] for the manifest/`Create`/`Open` mechanics
and its sibling child [[world-save-manifest-fields]]. [[llm-orchestrator]]
owns `llm.json`; [[cognition]] owns `calibration.json` and
`estimator_state.json`; [[daemon-lifecycle]] owns the socket/pidfile/log
runtime files; [[guardian]] owns `charter.md`/`metatron/`;
[[governance]] owns `village_charter.md`; [[morgue]] owns `morgue.md`;
[[bundle-tools]] owns `bundles/`; [[world-tuning]] owns `tuning.json`.

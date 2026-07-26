---
name: cli-guardian-ops
description: promptworld CLI guardian/LLM operator commands — guardian conversational one-shot, work miracle door, llm one-shot calls, calibrate benchmarking, divergence gate evidence
kind: component
sources:
  - cmd/promptworld/commands.go
  - cmd/promptworld/work.go
  - cmd/promptworld/calibrate.go
  - cmd/promptworld/divergence.go
verified_against: d304e8adb64fdf40e24bfeca3ca3420e8a840a35
---

# CLI: guardian and LLM operator commands

Split from [[cli-promptworld]] (full subcommand list there): the guardian's
conversational and operator-door commands, and the LLM setup/gate-evidence
commands — `guardian`, `work`, `llm`, `calibrate`, `divergence`.

## How it works

- `guardian <world> [message...]` (canonical since spec 052 FR-008; the pre-052
  name `metatron` survives as a hidden compat alias, same handler) — the
  console one-shot ([[guardian]], TASK-12): with
  a message, one mediated turn (prints surfaced moments, the reply, any landed
  `⚡ vision/omen` line, a `👁 watch set`/`👁 watch released` line for a placed or
  cancelled standing order, a `⏲` line for a landed pause/start/adjust_speed
  meta tool call (spec 029, [[guardian-orders]]), and the charge bank); without,
  a model-free status peek (charges, charter provenance, a `--- standing
  orders ---` block via `orderStatusLine` — id, fuzzy marker, origin,
  remaining game-day, status, condition — when any order stands, and recent
  soul notes).
- `work <world> <snap-time|give|move|remove> ... [--force]` (canonical since
  spec 052 FR-008; the pre-052 name `miracle` survives as a hidden compat
  alias, `work.go`) — the operator door
  for the guardian's workings ([[guardian-miracles]], spec 016 R6), a dedicated
  subcommand family independent of the `guardian` conversational path: `snap-time
  <day> <HH:MM>`, `give <villager> <item> <qty>`, `move <class> <x,y> <x1,y1>`,
  `remove <class> <x,y>` (`<class>` is `villager|structure|pile|terrain`; terrain
  is remove-only, villagers cannot be removed). Dials the daemon and calls the
  `miracle` IPC command directly (the wire command name is FROZEN, spec 052
  ruling 2) — no LLM involved. `--force` sets the gratis flag
  that waives the charge cost, an override reachable only from this CLI door, never
  from the angel's own turn. Prints the miracle summary (`(forced)` suffix when
  gratis) and the remaining charge bank.
- `llm <world> <kind> <prompt...> [--system] [--max-tokens]` — one-shot model call via
  the daemon's `llm_call` command; `formatLLMOneShot` prints the serving PROVIDER
  (never a tier — spec 024 FR-011), model, tokens, cost, and latency, plus a
  `skipped: name (reason)` line whenever the chain-walk passed over candidates
  ([[llm-orchestrator]]). `new` also writes the default `llm.json` config (v2
  registry shape; its hint says "edit providers/routes/budget").
- `calibrate <world> [--provider <name>] [--samples N]` (`--tier local|cloud|all`
  kept as a deprecated alias — on a v2 registry `local`/`cloud` select every
  zero-priced/priced provider with a deprecation note; on a legacy config it
  behaves exactly as before) — the cognition horizon's
  setup stage ([[cognition]], TASK-32): benchmarks the DECLARED PROVIDERS (spec
  024 T020: a legacy config runs the untouched pre-024 path over its two derived
  providers, byte-identical output; any v2 registry — or `--provider` — iterates
  `orch.ProviderNames()`, each reference call pinned via `Request.Provider` so
  the sample measures the named provider regardless of what its kind's chain
  currently resolves to) against fixed reference prompt shapes (default 5
  samples per shape; priced-provider spend is opt-in and announced up front),
  takes the median seconds-per-point, writes/merges `calibration.json` (one
  profile entry per provider name — the shape `cognition.SeedFor` reads), and
  prints the horizon the hardware buys (e.g. "planner suppressed above 16x") via
  `horizonSummary`, which since spec 035 (R1/T004) delegates to
  `cognition.HorizonSummary` — the same function the daemon's boot warning and
  the `set_speed` warning read, so the printed horizon can never disagree with
  the router (FR-006, [[cognition]]) — evaluating the registry
  across the watchable speed ladder (`planner`/`conversation`/`meeting` — `musing`
  dropped from the ladder with its retirement as a scheduled kind, spec 017).
  Every run also prints `sequentialFloorDisclosure` exactly once, adjacent to
  the horizon summary when one printed (a cloud-only or priced-only run with
  no horizon line still gets it): a reminder that calibration measures one
  reference call at a time while a live world drives the same endpoint
  concurrently, so the measured seconds-per-point is a floor and the live
  estimator adapts upward at runtime (spec 035 US4/FR-005). Since
  spec 017 (FR-011) the `planner-3pt` shape is a LOOP probe, not a bare
  completion: `villagerProbeJob` drives `toolloop.Run` with the real
  `tool.LoopRosterVillager()` roster and a no-op handler per tool (every read
  reports `read_ok`, every acting call reports `landed` — ending the loop on the
  model's first action, since calibration measures round-trip latency, not
  landings) so the seeded seconds-per-point is measured in the SAME whole-loop
  unit `Orchestrator.ObserveCognition` later feeds live ([[llm-orchestrator]],
  [[tool-loop]]) — a representative tool loop's wall time, not one call's. The
  probe's round cap is `cfg.Rounds()` (the daemon's own `loop_max_rounds`), so the
  calibration and the live cognition share one horizon. Reference shapes select
  by PRICING CLASS (`refShapesFor(priced)`, spec 024 T020 generalizing the old
  tier branch): zero-priced providers get the loop probe, priced providers'
  `consolidation-5pt` shape stays a plain single-shot `Submit` (consolidation did
  not adopt the loop, FR-014) — the guardian IS the cloud's loop cognition, but
  calibrating it would drive extra metered cloud calls the spec 017 contract
  doesn't invite; its live whole-loop observations converge the cloud estimator
  at run time instead. Uses an in-memory meter so it never contends
  with a running daemon's store; a provider whose every sample fails is not
  written.
- `divergence <world> [--json]` (`cmd/promptworld/divergence.go`, spec 042 US2) —
  offline gate evidence for the embedding-memory shadow→on flip: reads
  `cog.memory_divergence` events straight from the store (the `tail` pattern, no
  daemon required) and aggregates per-agent/per-game-day rows plus a whole-run
  total — mean overlap@K against the legacy ranking, the promoted-memory share
  (selections where relevance pulled in at least one memory legacy excluded),
  mean displacement, and mean vectorless-fallback count. Refuses with a one-line
  remedy ("set world.json memory_relevance to \"shadow\"") when no divergence
  events exist yet. `--json` emits the same rows machine-readably. The printed
  table is recorded evidence only — flipping the world to non-shadow relevance
  is an operator decision made on the board task, never an automatic action.

## Connections

`guardian`'s standing-orders rendering reads [[guardian-orders]]. `work` (hidden
alias `miracle`) hands off to [[guardian-miracles]] via the same IPC door
[[ipc-server]]'s `handleMiracle` implements. `llm`'s output and `new`'s default
config read/write [[llm-orchestrator]]. `calibrate`'s horizon summary and
`status`'s calibration rows both read [[cognition]]'s `Calibrated`/
`HorizonSummary`; its loop-probe shape rides [[tool-loop]]. `divergence` reads
`cog.memory_divergence` events off [[memory-retrieval]]'s shadow-mode selector
records, offline via the same store/`tail` pattern as `tail`/`status`'s offline
path. See [[cli-promptworld]] for exit discipline, arg resolution, and the
sibling families ([[cli-world-lifecycle]], [[cli-runtime-control]]).

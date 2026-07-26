# Research: World fork + duel v1 (spec 076)

Decisions with evidence and rejected alternatives. Every claim below was verified against
code or a pinned wiki note on 2026-07-26 (main at the TASK-149 merge, PR #113).

## R1 — Fork mechanics: fresh prefix log, never copy-then-truncate

**Decision**: `world.Fork` builds a fresh `world.db`: stream the parent's events with
`seq <= boundary.seq` (via `Store.ReplayEvents`) into the new store's `AppendEvents`
(fresh store, `lastSeq = 0` → assigned seqs reproduce 1..N exactly), write the boundary
snapshot verbatim (`SaveSnapshot(boundary.tick, boundary.seq, boundary.state)` — hash
recomputed by the store, re-verified against the parent's `state_hash` first), append the
`world.forked` lineage event, stamp meta.

**Evidence**:
- The `events` table carries in-schema `events_no_update`/`events_no_delete` triggers
  (`internal/store/schema.go`; `docs/wiki/event-log.md`: "immutability is enforced
  in-schema, not by convention"). A copied db cannot be truncated without dropping
  triggers or raw file surgery — both doctrine violations.
- The migration ceremony is the precedent for "build a fresh log + covering snapshot"
  (`internal/world/migrate.go`; `docs/wiki/world-save-directory.md`: "write a fresh log
  (`world.created` then `world.migrated`) plus its covering snapshot").
- `AppendEvents` assigns contiguous seqs `lastSeq+1…` (`internal/store/store.go:79`), so a
  prefix streamed in order into an empty store keeps its seq identity; `CheckContiguity`
  (1..N) holds by construction.

**Rejected**:
- *File-copy `world.db` + DELETE tail*: blocked by the triggers; dropping them to delete
  is exactly the convention-vs-schema hole the schema exists to close.
- *File-copy + keep the tail (no truncation)*: the fork would carry events past the
  boundary snapshot; recovery would fold them in and the fork would NOT be "at the
  snapshot" — the semantics the board pinned.
- *Rebuild from state only (the `world.migrated` wholesale-state pattern, no history)*:
  loses the parent's event history in the fork, which the duel's chronicle drill-down and
  the "story so far" (chronicle ring rides state, but the log is the deep record) both
  want; also makes "replay from genesis" trivially true rather than meaningful.

**Note on `wall_time`**: `AppendEvents` stamps wall time at insert; it is observability
metadata excluded from determinism comparisons (`docs/wiki/event-log.md`). The fork's
prefix therefore carries fresh wall times — acceptable and documented; preserving the
originals would require a new store API for zero determinism benefit.

## R2 — Seed identity: the seed stays; identity is name/directory/socket

**Decision**: the fork keeps the parent's seed. Fresh identity = manifest name, directory,
socket/pidfile (path-derived), registry presence. Nothing else.

**Evidence**: every random decision is a pure function of `(seed, purpose, tick, index)`
(`internal/sim/rng.go`; `docs/wiki/deterministic-rng.md`). The carried event prefix was
generated under the parent seed; replaying it under a different seed would break the
replay-to-hash proof immediately, and post-fork ticks would diverge for RNG reasons rather
than prompt reasons.

**Teaching consequence (recorded)**: two forks with the same seed and NO divergent
recorded inputs evolve identically — so any divergence the duel shows is attributable to
the actual input difference (the prompt change, via LLM outputs entering as recorded
inputs, or player commands). This is a feature: the fork-duel is a controlled experiment
by construction. It also yields the zero-divergence edge case (spec US3 scenario 4).

## R3 — Lineage home: authoritative event + additive manifest mirror

**Decision**: `world.forked` in the fork's log is authoritative (AC2's "durably
recorded"); an additive `omitempty` `Manifest.Lineage` block is the fast offline mirror.

**Evidence**:
- Event conventions: payload structs, canonical JSON (`internal/sim/state.go` payload
  block — `WorldCreatedPayload`/`WorldMigratedPayload` are the genesis/provenance
  precedents). The reducer's contract: "Unknown and daemon.* event types are recorded
  history but no-ops on state" (`state.go` Apply doc) — a no-op arm keeps fork state
  byte-identical to parent state at the fork tick, which upgrades AC4 from "passes the
  harness" to a provable identity.
- Manifest conventions: additive `omitempty` fields with no `format_version` bump
  (`teaching`, `memory_relevance`, `stage`, `scenario` — `docs/wiki/
  world-save-manifest-fields.md`); pre-fork manifests round-trip byte-identically.

**Rejected**:
- *Manifest only*: a manifest is mutable-by-hand and outside the log's replay guarantee;
  AC2 says durable — the log is the durability standard in this codebase.
- *Event only*: forces every offline consumer (compare's default window; future `ps`
  annotations) to open the store and scan; the manifest mirror is one JSON read and
  follows the `meeting`-block precedent of manifest-carried world facts.
- *Lineage on `sim.State`*: pointless serialization churn in every snapshot for a fact
  that never changes and that no reducer logic reads; and it would break the byte-identity
  property above.

## R4 — Budget (board AC #5): inherit the wallet; reject the literal shared wallet

**Decision**: copy `llm.json` verbatim AND copy every `llm_spend_*` meta key (monthly
totals + per-provider attribution) into the fork's fresh store. The fork's meter opens at
the parent's month/spend/ceiling. Thereafter each world meters independently.

**Evidence**:
- The meter persists "in the world's meta table so restarts never forget money"
  (`internal/llm/meter.go:16-19`); the budget number comes from the world's own
  `llm.json` (`monthly_budget_usd`). There is no machine-global spend state anywhere.
- "One wallet, per-provider attribution" / "a single global monthly_budget_usd ceiling"
  (spec 024 US4; `docs/wiki/llm-budget-degraded-mode.md`) is one wallet across PROVIDERS
  within a world — the "global" is provider-global, not machine-global.
- The grounded design decision "never global; runs cleanly separable"
  (`docs/wiki/world-save-directory.md` Operational notes) and the copied-world e2e
  (`TestManagerCopiedNameWorldRunsUnderFreshHome`: a world must run with zero external
  state) make a shared mutable wallet architecturally inadmissible.

**Runbook deviation, recorded per the sweep instruction**: the runbook recommendation
reads "forks share the single global monthly spend ceiling." Its implementable reading —
same ceiling value, inherited spend-to-date, no fresh grant at fork — is what this spec
encodes. Its literal reading (one mutable wallet spanning two worlds) contradicts the
architecture above and is rejected with this evidence.

**Rejected**:
- *Fresh wallet (copy nothing)*: forking becomes a budget-doubling loophole the moment it
  ships — `fork` + delete parent = ceiling reset mid-month.
- *Machine-global wallet file*: violates cleanly-separable; breaks the copied-world
  proof; invents the first cross-world mutable state in the codebase.

**Recorded limitation (FR-013)**: post-fork combined spend across a duel can exceed one
ceiling by up to the unspent remainder at fork time. Attribution stays per-world and
per-provider, unchanged. Cost-attribution coarseness was already flagged by the 2026-07-22
review and is explicitly not this spec's problem to fix.

## R5 — Compare's read path: offline snapshot + fold, shared with `ps`

**Decision**: extract `worlds.OfflineSnapshot`'s state-reconstruction core
(`internal/worlds/probe.go:190` — newest valid snapshot unmarshaled into
`sim.NewState(seed, map)`, then `seq > snap.seq` events folded through `Apply`) into a
helper returning the `*sim.State` itself; `OfflineSnapshot` and compare both call it.

**Evidence**: the mechanics already exist and were extended by spec 044 to fold trailing
events precisely so offline reads mirror recovery; compare needs the same state to feed
`EvaluateRubric` and `CurriculumPasses` lookups. WAL mode permits a second-process reader
against a running daemon's db (readers never block the single writer; `journal_mode=WAL`,
`internal/store/store.go`), so compare works live, honestly labeled as-of-last-commit.

**Rejected**: querying a running daemon over IPC for its state — heavier, requires the
daemon up on both sides, and duplicates a read path that exists; v1 compare is a pure
offline reader.

## R6 — Scoreboard: export the spec-072 resolver, forbid a second switch

**Decision**: refactor `resolveReportCardFacts` (`internal/tui/reportcard.go:127`)
replica-parametric — `ResolveRubricFacts(state *sim.State, def sim.ExerciseDefinition,
pass *sim.CurriculumPass) ([]ReportCardFact, ReportCardMode)` — with
`type ReportCardFact = reportCardFact` / `type ReportCardMode = reportCardMode` aliases
and an exported `RenderReportCard(title string, facts []ReportCardFact, mode
ReportCardMode, width int) string` wrapping `reportCardView`. The Model method and all
three TUI call sites become thin wrappers; behavior is unchanged (existing TUI tests are
the proof). The CLI duel becomes the fourth consumer.

**Evidence**: spec 072's whole contract is "the ONE precedence switch every card surface
derives facts through" (`reportcard.go:120-126` doc comment; `docs/wiki/
report-card-renderer.md`); `cmd/promptworld` already links `internal/tui` (the `ui`
subcommand), so no new dependency edge is created. The 2026-07-26 board rescope names the
sharing explicitly: "v1 = rubric-first scoreboard sharing reportCardView +
sim.EvaluateRubric."

**Rejected**:
- *Duplicate the precedence logic in the CLI*: recreates the drift spec 072 was shipped
  to kill; explicitly forbidden (spec FR-018).
- *Move the resolver into `internal/sim`*: `reportCardFact` is a render shape, not sim
  truth; sim already owns the verdict currency (`RubricTerm`) — the split is deliberate.
- *A new `internal/duel` package*: nothing else would live in it; the resolver's home is
  where its types live.

`recordedPassFor` gets the same treatment (a state-parametric lookup over
`state.CurriculumPasses`) so the CLI finds the instrument the same way the TUI does.

## R7 — Divergence: story events only; machinery classes excluded

**Decision**: divergence compares `(tick, type, payload)` streams post-window, excluding
`wall_time` and the type classes `daemon.*`, `clock.*`, `cog.*`, `llm.*`. Chronicle
entries (`chronicle.entry`) render in the interleave but never trigger divergence.

**Evidence & rationale**:
- The determinism e2e already excludes `daemon.*`/`clock.*` as "wall-dependent
  bookkeeping" (`e2e/determinism_e2e_test.go:38-40`). Two separately-run forks
  additionally record cognition/LLM telemetry (`cog.*` outcomes with latency arithmetic,
  `llm.*` warnings) whose payloads differ for wall reasons even when the villagers' story
  is identical — counting them would make every duel "diverge" at the first model call
  regardless of the prompt.
- `chronicle.entry` text is narrator (cloud LLM) prose — two runs of the SAME story
  produce different wording; treating wording as divergence would be false signal. The
  underlying story events are the recorded truth the chronicle compresses
  (`docs/wiki/chronicle.md`).

**Rejected**: comparing everything (false positives, above); comparing only
`chronicle.entry` (narrated wording noise AND misses divergence in worlds where the
narrator is off — no `llm.json`, no narrator, but the story events still diverge).

## R8 — Refusal postures

- **Running source daemon → refuse** (spec FR-003). Precedent: `world.Migrate` refuses a
  live daemon. Sidecar copies race a live daemon (guardian transcript, estimator flush);
  WAL would make the db read safe but the ceremony is more than the db. Pausing instead
  was rejected: pause still holds the db open with an active daemon writing meta
  (estimator flush) and runtime files present.
- **No valid snapshot → refuse** with the start-and-stop remedy (spec edge case). A
  genesis fork of a never-run world is `new --seed`, which exists; inventing a
  no-boundary fork path buys nothing and complicates the "at the snapshot" contract.
- **Ended parent → allow, warn**: the boundary snapshot may carry the ended state; the
  fork is then born ended. Refusing would be wrong (an operator may want an archive-fork);
  warning is honest. The useful "retry from before the collapse" is the mid-log follow-on,
  recorded in Out of Scope.

## R9 — What copies, what doesn't (sidecar catalog)

Per-file rationale, from the path-helper catalog (`docs/wiki/world-save-path-helpers.md`):

| File | Fork action | Why |
|---|---|---|
| `world.json` | rewritten | new name, new created_at, + `lineage`; all else verbatim (R2) |
| `world.db` (+`-wal`/`-shm`) | rebuilt fresh | R1 — prefix log + boundary snapshot + lineage + meta |
| `llm.json` | copy | same ceiling — R4; deletable in either world to disable inference |
| `calibration.json` | copy | machine-local seconds-per-point profile; equally true for the fork |
| `estimator_state.json` | copy | advisory latency seeds; never replayed, absent is legal anyway |
| `charter.md` | copy | player INPUT (the thing the duel diverges); as-of fork time, documented coarseness |
| `metatron/` | copy | guardian soul + transcript; advisory prose, same coarseness note |
| `bundles/` | copy | boot-frozen drop-in content; part of the world's behavior under test |
| `tuning.json` | copy | per-world dials; the fork must inherit the parent's physics |
| `agents/` contents | copy | flat files for later features; verbatim |
| `chronicle.md`, `morgue.md`, `village_charter.md` | skip | scribe-regenerated views over recovered state at every daemon start; copying ships prose about truncated-away events |
| `daemon.sock`, `daemon.pid`, `daemon.log` | skip | runtime-only; swept when stale anyway |
| `world.v*.db` migration archives | skip | the PARENT's history; the fork's fresh log is self-contained |

Failure cleanup: fork creates the destination, and on any ceremony error best-effort
removes it (the destination was required-empty, so removal destroys nothing pre-existing);
a crashed fork leaves a partial dir the non-empty check forces the player to delete —
documented, matching `new`'s posture on a half-created world.

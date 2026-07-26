# Feature Specification: World fork + duel v1 — `promptworld fork` and the rubric-first scoreboard

**Feature Branch**: `076-world-fork-duel` (task branch: `task-67-world-fork-duel`)

**Created**: 2026-07-26

**Status**: Draft

**Input**: TASK-67 / reorient 2026-07-26 decision 3 (docs/design/reorient-2026-07-26-ui.md,
merged position 3 — "the iteration rung"). Rescoped 2026-07-25 (D7: rubric-first scoreboard
sharing the postmortem's renderer) and 2026-07-26 (all D7 prerequisites shipped; AC #7's
rubric IS `sim.EvaluateRubric`'s terms, never a bespoke list). Depends on TASK-149
(spec 072, merged in PR #113): the duel must not compare false checkmarks.

## Problem (pinned)

The learning loop is built but cannot be iterated. Replay is model-free — LLM outputs are
recorded inputs (`docs/wiki/chronicle.md`: model output "enters the world only as recorded
input") — so a player cannot re-run yesterday under a new prompt. The persistence substrate
makes world FORKING cheap instead: save dirs are fully self-contained and copyable (proven
by `TestManagerCopiedNameWorldRunsUnderFreshHome`, `e2e/manager_e2e_test.go:363` — a copied
world runs under a fresh home with zero manager state), snapshots bound recovery at a
hash-verified `(tick, seq)` boundary (`internal/store/store.go:164`,
`Store.LatestValidSnapshot`), and each world is its own daemon on its own socket. But no
fork or compare subcommand exists (`cmd/promptworld/main.go:66-110` — the dispatch table
has neither), so the most direct way a learner could SEE what their prompt change did —
fork the village, diverge the charter, run both, compare the outcomes — does not exist.

The comparison half is now dramatically cheaper than when first scoped: spec 054 shipped
the rubric evaluator (`sim.EvaluateRubric`, `internal/sim/scenario.go:280` — pure over
(state, definition, tick)), specs 056/063 shipped the shared report-card renderer
(`reportCardView`, `internal/tui/views.go`), and spec 072 (TASK-149, PR #113) unified every
card surface on ONE fact resolver (`resolveReportCardFacts`,
`internal/tui/reportcard.go:127`) whose verdicts are honest — a failed term renders ✗. The
duel scoreboard is one more consumer of that exact resolver family, and a lost duel reads
as a postmortem: the no-blame register that only teaches because the evidence is true.

One structural fact shapes the whole design: the event log is append-only **in-schema**
(`events_no_update`/`events_no_delete` triggers, `internal/store/schema.go` — "immutability
is enforced in-schema, not by convention"). "Copy the save dir truncated to the snapshot
boundary" therefore cannot mean copy-then-DELETE: the fork builds a **fresh log carrying
the parent's event prefix**, the migration ceremony's precedent (`internal/world/migrate.go`
writes a fresh log plus its covering snapshot).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Fork a world at its latest snapshot; both run side by side (Priority: P1)

A player whose `first-night` village survived to day 3 runs
`promptworld fork aria aria-b`. A new world `aria-b` appears in the worlds home: the same
village — same seed, same history up to the latest snapshot boundary, same charter file,
same stage and scenario — under a fresh identity (its own name, directory, socket, and
registry entry). The fork's log durably records where it came from (`world.forked`:
parent name, parent seed, fork tick). The player edits `aria-b`'s charter, then starts
BOTH worlds; they run simultaneously, each on its own daemon, and `promptworld ps` lists
both.

**Why this priority**: board ACs #1 and #2 — the fork verb is the substrate everything
else in this spec (and phase 2's HTML retelling) stands on.

**Independent Test**: e2e — create, run past a snapshot, stop, fork, start both, assert
both answer `status` and `ps --json` shows two running worlds; unit — fork ceremony
produces a contiguous log, a verifying boundary snapshot, and the lineage event.

**Acceptance Scenarios**:

1. **Given** a stopped world with at least one hash-valid snapshot, **When**
   `promptworld fork <world> <new-name>` runs, **Then** a new save directory exists whose
   event log is exactly the parent's events with `seq <= boundary.seq` (contiguous 1..N)
   plus one `world.forked` event, whose snapshot table carries the boundary snapshot
   verbatim, and whose manifest carries the new name, a `lineage` block, and every other
   field (seed included) unchanged.
2. **Given** the fork, **When** both worlds are started, **Then** both daemons run
   simultaneously (distinct sockets/pidfiles derive from distinct directories), both
   answer `status`, and `ps` lists both with their own names.
3. **Given** the fork's log, **When** it is replayed from genesis through the reducer,
   **Then** the resulting state's canonical hash equals the boundary snapshot's
   `state_hash` — and equals the PARENT's state hash at the same `(tick, seq)`: the
   `world.forked` reducer arm is a recorded-history no-op, so fork state at fork tick is
   byte-identical to parent state at fork tick.
4. **Given** a fork destination that already exists and is non-empty, or a `<new-name>`
   failing `worlds.ValidateName`, **Then** fork refuses with the same discipline as
   `promptworld new` (no partial world left behind on failure).
5. **Given** the printed fork summary, **Then** it names the boundary (game day/time and
   tick), the number of events carried, the truncated tail (events past the boundary that
   did NOT carry over, if any), and the commands to start both worlds.

---

### User Story 2 - The duel scoreboard: rubric-first compare (Priority: P2)

The player runs `promptworld compare aria aria-b`. The output leads with the duel header
(lineage: "aria-b forked from aria at day 3, 06:00") and then the scoreboard: one
plain-language rubric card per world, built from `sim.EvaluateRubric`'s terms through the
spec-072 shared resolver — hand-authored labels ("no villager dies"), truthful glyphs
(✓/✗ concluded, … live), honest backing counts (`agent.died: 2`). The world that lost
reads as a postmortem — concluded ✗ facts in the no-blame register — never a scolding and
never a false ✓. "Here is what your prompt change did," in the rubric vocabulary the
curriculum ladder teaches.

**Why this priority**: board AC #7 (and the 2026-07-26 rescope's core: the rubric IS
`EvaluateRubric`'s terms, shared with every other card surface — no bespoke list). AC #5's
budget decision also lands with this story's spec (Requirements FR-012/013).

**Independent Test**: compare on a fixture pair (one passed, one run-ended `first-night`
world) renders both cards with resolver-derived rows: winner all-✓ from its recorded pass,
loser ✗ on the death term with `agent.died: N` backing; no raw enum appears in the grade.

**Acceptance Scenarios**:

1. **Given** two scenario worlds on the same exercise, **When** compare renders, **Then**
   each world's per-term verdicts are derived through the SAME precedence switch spec 072
   shipped (recorded pass → concluded rubric → live rubric) — a duplicated verdict
   computation anywhere in the compare path is a spec violation.
2. **Given** a world whose run ended without a pass, **Then** its card renders concluded
   ✗ on failed terms with honest backing counts, and its outcome line reads in plain
   language ("did not make it through" / the exercise outcome vocabulary rendered
   plainly) — never the raw `in_progress`/`failed` enum tokens.
3. **Given** a world still running, **Then** its card renders live markers (… pending)
   and the header says so — compare works on running worlds (offline read of the last
   committed state) and stopped worlds alike.
4. **Given** an ambient world (no `scenario` block) on either side, **Then** compare says
   honestly that no rubric exists for that world and still renders the chronicle sections
   (US3) — it never invents a scorecard.
5. **Given** two worlds with DIFFERENT exercises, **Then** each card renders under its own
   exercise with a note that the duel is not head-to-head — cards are per-world truth,
   never forced into one table.

---

### User Story 3 - Drill-down: interleaved chronicles with divergence visible (Priority: P3)

Below the scoreboard, compare renders the two stories against each other: chronicle
entries from both worlds after the fork point (or `--since <tick>`), interleaved in time
order and labeled per world, with a divergence marker at the first point the two runs'
recorded stories actually differ — "the stories diverge at day 3, 21:40". The player sees
not just THAT the outcome differed but WHERE the two histories split.

**Why this priority**: board AC #3. It completes the teaching loop (the scoreboard says
what happened; the chronicles say where), but the scoreboard alone already answers "did my
prompt edit work?".

**Independent Test**: fixture logs sharing a prefix; assert the divergence marker lands at
the first differing story event, entries interleave by tick with correct world labels, and
two identical-after-fork logs render the honest "no divergence" line.

**Acceptance Scenarios**:

1. **Given** a forked pair, **When** compare renders with no `--since`, **Then** the
   comparison window defaults to the fork tick read from the fork's lineage; unrelated
   worlds default to tick 0 and may narrow with `--since`.
2. **Given** the two logs, **Then** the divergence point is the first post-window event at
   which the streams differ, compared over STORY events only — machinery classes
   (`daemon.*`, `clock.*`, `cog.*`, `llm.*`) and `wall_time` are excluded, so two runs
   that told the same story at different wall speeds do not falsely diverge.
3. **Given** chronicle entries (`chronicle.entry`) in both logs after the window, **Then**
   they render interleaved by their `from_tick`, each line labeled with its world's name,
   with the divergence marker inserted in timeline position.
4. **Given** two forks whose post-fork histories are identical (pure-sim forks with no
   divergent inputs — the deterministic-RNG consequence: with the same seed and no
   differing recorded inputs, both evolve identically), **Then** compare renders an honest
   "the two runs are identical since the fork" line — zero divergence is a truthful,
   teachable outcome (your prompt change changed nothing), never an error.

---

### Edge Cases

- **Fork of a never-snapshotted world** (created but never run, or killed before the
  first snapshot): `LatestValidSnapshot` returns nil — fork REFUSES with a remedy
  ("no snapshot yet — start and stop the world once to cut one"). v1 forks at the latest
  snapshot only; there is no genesis-fork special case (a genesis fork of a never-run
  world is just `new --seed`, which already exists).
- **Fork while the source daemon is running**: REFUSED (the `world.Migrate` precedent —
  refuse a live daemon). Copying sidecar files out from under a live daemon races; the
  player stops, forks, then starts both. Documented in the command's error with the stop
  command named. (Compare, by contrast, is read-only and works on running worlds.)
- **Name collisions**: `<new-name>` follows `promptworld new`'s exact rules —
  `worlds.ValidateName`, worlds-home placement for a bare name, path form accepted; a
  non-empty destination directory is refused (`world.Create`'s posture). A fork that fails
  partway removes its partial destination (best-effort) so a retry is clean.
- **Forked-world registry entries**: a fork created in the worlds home needs no registry
  entry (home is scan-owned); a fork created at an explicit path self-registers the way
  every world does — the daemon's boot-time upsert (`registerWorld`) — because the
  registry is advisory and self-healing (`docs/wiki/instance-manager.md`). Fork itself
  writes no registry state.
- **Chronicle divergence with zero divergent events**: see US3 scenario 4 — rendered as
  the honest identical-since-fork line, pinned by test. This is a real outcome, not a
  degenerate case: sim-only forks (no LLM) are deterministic by construction.
- **Snapshot trails the log tip**: the boundary is the latest snapshot, which can trail
  the log by up to one game hour of events (`SnapshotEveryTicks = 3600`) plus anything
  after a crash. Those tail events are deliberately NOT carried (that is what "truncated
  to the snapshot boundary" means); the fork summary names the truncated span so the
  player is never surprised. A graceful `stop` cuts a final snapshot, so the common
  stopped-world fork loses nothing.
- **Fork of an ENDED world**: legal but usually not what the player wants — the latest
  snapshot of a gracefully stopped ended run already carries the ended state, so the fork
  is born ended too. The fork summary warns when the boundary state is ended. Forking an
  EARLIER snapshot (retry the run from before the collapse) is exactly the mid-log/
  chosen-snapshot forking documented as out of scope for v1 (see Out of Scope) — this
  edge is its sharpest motivation, recorded for the follow-on.
- **charter.md / metatron/ postdating the boundary**: sidecar files are copied as-of fork
  time, not as-of the snapshot — the charter is player INPUT, not event-sourced state
  (its content enters the world only when a guardian turn observes it and records the
  fingerprint). A charter edited after the snapshot rides into the fork; the fork's next
  guardian turn observes it fresh. Documented coarseness, harmless by construction.
- **Migrated parent**: the parent's migration archives (`world.v1.db`…) are the PARENT's
  history and do not carry into the fork — the fork's fresh current-format log is fully
  self-contained (replay-provable from genesis, the migration's own covering-snapshot
  doctrine).
- **Compare on worlds of different formats / unopenable worlds**: compare resolves both
  arguments through the standard `resolveWorld` path; an unopenable world fails with the
  standard `ErrUnopenable` message (migrate hint included) — no special handling.

## Requirements *(mandatory)*

### Functional Requirements

Mapped to the seven board ACs: AC1 ↔ FR-001..006, AC2 ↔ FR-007..009, AC3 ↔ FR-015..017,
AC4 ↔ FR-010..011, AC5 ↔ FR-012..013, AC7 ↔ FR-018..020. AC6 (spec written + linked
before implementation) is satisfied by this document set plus the `spec-bridge:link` step
recorded in tasks.md Phase 1. FR-014 and FR-021 are cross-cutting doctrine.

- **FR-001**: A `fork` subcommand MUST exist: `promptworld fork <world> <new-name>
  [--at latest-snapshot]`. `<world>` resolves through the standard name-or-path
  resolution; `<new-name>` follows `new`'s conventions (bare name → validated by
  `worlds.ValidateName`, created at `<worlds-home>/<name>`; path form → exact directory,
  name from the basename). A non-empty destination is refused.
- **FR-002**: v1 forks at the **latest hash-valid snapshot only**
  (`Store.LatestValidSnapshot` — the same newest→oldest verified walk recovery uses).
  `--at` accepts exactly the value `latest-snapshot` (also the default when the flag is
  absent); any other value is refused with a message stating that mid-log / chosen-
  snapshot forking is a documented follow-on, not a v1 capability. A world with no valid
  snapshot is refused with the start-and-stop remedy.
- **FR-003**: Fork MUST refuse a running source daemon (`daemon.IsRunning` — the
  `world.Migrate` precedent), naming the `stop` command in the error.
- **FR-004**: The fork's `world.db` MUST be built fresh — never a file copy that is then
  modified: the parent's events with `seq <= boundary.seq` are streamed in order into the
  new log (contiguous seqs 1..N preserved by construction), the boundary snapshot's
  `(tick, seq, state)` is written verbatim (its hash re-verified), and meta is stamped
  (`seed`, `format_version` — matching what `validateMeta` will check at first boot).
  Events in ANY database are never deleted or updated (the in-schema append-only
  triggers are doctrine, not an obstacle).
- **FR-005**: Fresh identity means: manifest `name` = the new name, `created_at` = fork
  wall time, and **every other manifest field verbatim** — `seed` IDENTICAL (replay
  determinism requires it: the carried events were generated under the parent seed, and
  `sim.rngAt` keys off it; a fork's identity is its name, directory, socket, and registry
  presence — never its seed), `format_version`, map dims, `terrain_gen`, `meeting`,
  `teaching`, `memory_relevance`, `stage`/`stage_overridden`/`charter_preset`, and
  `scenario` all carried. Sidecar files copied: `llm.json`, `calibration.json`,
  `estimator_state.json`, `charter.md`, `metatron/`, `bundles/`, `tuning.json`, and any
  `agents/` contents. NOT copied: runtime files (`daemon.sock`, `daemon.pid`,
  `daemon.log`), the parent DB and its WAL sidecars, migration archives (`world.v*.db`),
  and the scribe's regenerable views (`chronicle.md`, `morgue.md`, `village_charter.md` —
  regenerated from recovered state at the fork's first daemon start; copying them would
  ship prose describing post-boundary events the fork never had).
- **FR-006**: Parent and fork MUST run simultaneously: distinct directories give distinct
  sockets/pidfiles by construction (`SockPathIn`/`PidPathIn` are pure path joins), and an
  e2e proof starts both, gets `status` from both, and sees both in `ps`.
- **FR-007**: Lineage MUST be durably recorded in the fork's own log as a new event type
  `world.forked`, appended at `tick = boundary.tick`, `seq = boundary.seq + 1`, payload
  `WorldForkedPayload{parent_name, parent_seed, parent_created_at, fork_tick, fork_seq}`
  (canonical struct-marshaled JSON, the payload-struct convention). The reducer arm is a
  recorded-history NO-OP on state (the `world.created` arm's exact posture) — which is
  what makes US1 scenario 3's byte-identity property hold.
- **FR-008**: The fork's manifest MUST carry an additive `omitempty` `lineage` block —
  `{parent, parent_created_at, fork_tick}` — the offline fast-read mirror (compare's
  default window, `ps`-adjacent tooling) of the authoritative event. No `format_version`
  bump (the `teaching`/`memory_relevance` additive-field precedent); `Open` validates a
  present block structurally (non-empty `parent`, `fork_tick >= 0`) and tolerates
  absence byte-identically.
- **FR-009**: `world.forked` MUST enter the shared event vocabulary completely: a digest
  registry entry + fixture (the `TestCatalogSweep` totality gate — every backticked
  concrete type in the wiki's event catalog must have a covered digest) and a row in the
  `docs/wiki/` event-types catalog, in the same PR.
- **FR-010**: Determinism (board AC #4): tests MUST prove (a) replaying the fork's log
  from genesis through the reducer reproduces the boundary snapshot's `state_hash`
  exactly; (b) the fork's state hash at fork tick equals the parent's state hash at the
  same `(tick, seq)`; (c) a fork that then RUNS (e2e, pure-sim at max speed) produces a
  log that itself replays from genesis to its final snapshot's hash — the forked world
  passes the determinism harness independently, not by inheritance.
- **FR-011**: The fork MUST be as self-contained as any world: no manager state anywhere
  is required for it to run (the copied-world e2e's SC-004 property, inherited — the
  fork ceremony writes nothing outside the destination directory).
- **FR-012** (board AC #5 — decided): **the fork inherits the parent's wallet as of fork
  time.** `llm.json` is copied verbatim (same `monthly_budget_usd` ceiling) and every
  `llm_spend_*` meta key (the authoritative monthly totals AND the per-provider
  attribution keys) is copied into the fork's fresh `world.db`, so the fork's meter opens
  showing the same month, the same spend-so-far, and the same ceiling — forking never
  mints fresh budget. Grounding note: the ceiling is per-world BY ARCHITECTURE — the
  meter persists in the world's own meta table (`internal/llm/meter.go:16-19`) with the
  budget read from the world's own `llm.json`; "one wallet / single global ceiling" in
  the budget doctrine (spec 024 US4, `docs/wiki/llm-budget-degraded-mode.md`) means one
  wallet ACROSS PROVIDERS within a world. A literal machine-shared wallet would violate
  the grounded "never global; runs cleanly separable" decision and the copied-world
  zero-manager-state proof. The sweep runbook's recommendation ("forks share the single
  global monthly spend ceiling") is honored in its implementable reading — shared ceiling
  value, inherited spend, no fresh grant — and the literal shared-mutable-wallet reading
  is rejected on that evidence.
- **FR-013**: The recorded limitation rides with FR-012: after the fork, each world
  meters independently, so a duel's combined forward spend can exceed one ceiling by up
  to the unspent remainder at fork time; per-world/per-provider attribution continues to
  work exactly as-is. This coarseness is documented (spec + wiki), not fixed here — per
  the board's scope decision.
- **FR-014**: **Forks are independent worlds forever.** No merge verb exists or will;
  nothing in this feature reads across two worlds except the read-only compare command.
  Doctrine, recorded here and in the wiki note this feature adds/amends.
- **FR-015**: A `compare` subcommand MUST exist: `promptworld compare <a> <b>
  [--since TICK]`. Both arguments resolve through standard resolution. State per world is
  reconstructed OFFLINE — newest valid snapshot unmarshaled into
  `sim.NewState(seed, map)` then events with `seq > snapshot.seq` folded through `Apply`
  (the `OfflineSnapshot` mechanics, extracted so compare and `ps` share one
  implementation) — which works on stopped worlds and, under WAL, on running worlds
  (reflecting the last committed batch, said honestly in the header).
- **FR-016**: The comparison window defaults to the fork tick when either world's
  lineage names the other as parent (manifest `lineage` block, event fallback); otherwise
  tick 0. `--since TICK` overrides. The duel header renders the lineage in plain language
  (game day/time via the clock arithmetic, plus tick).
- **FR-017**: Divergence detection compares the two logs' post-window event streams over
  `(tick, type, payload)` — excluding `wall_time` and the machinery type classes
  `daemon.*`, `clock.*`, `cog.*`, `llm.*` (divergence means the villagers' STORY
  differs, not that two daemons ran at different wall speeds or recorded different
  latency telemetry; the determinism e2e's `daemon.*`/`clock.*` exclusion, extended to
  the cognition/LLM telemetry classes with the rationale pinned in research.md R7). The
  first differing position is the divergence point, rendered with game day/time; equal
  streams render the identical-since-fork line (US3 scenario 4). Chronicle entries
  (`chronicle.entry` events, whose narrated text is expected to differ in wording) are
  rendered in the interleave but are NOT the divergence trigger — divergence keys on the
  underlying story events.
- **FR-018**: The scoreboard MUST derive per-world facts through the spec-072 shared
  resolver — refactored replica-parametric and exported from `internal/tui` (the Model
  methods become thin wrappers; call sites and behavior unchanged) — recorded pass →
  concluded `EvaluateRubric` → live `EvaluateRubric`, labels always `RubricTerm.Label`.
  Implementing a second precedence switch (in the CLI or anywhere) is forbidden: one
  resolver was the entire point of spec 072.
- **FR-019**: Scoreboard rendering MUST go through the shared renderer family
  (`reportCardView` via an exported entry point) — bordered card, `✓/✗/…` markers, one
  line per fact — so the duel card and the postmortem card are the same artifact by
  construction. Glossary discipline (board AC #7): every graded label is the evaluator's
  hand-authored plain language; outcome words render plainly ("passed" stays, `in_progress`
  renders as "still running", `failed` renders in the postmortem register — e.g. "did not
  make it through"); no raw enum token appears in a grade. Backing references
  (`agent.died: 2`, `seq N`) are evidence, not grades, and keep the spec-072 form.
- **FR-020**: A lost duel reads as a postmortem: the loser's card is the concluded-✗
  spec-072 card with factual backing, and every line of authored compare prose stays in
  the no-blame register (facts about the village, never judgments about the player) —
  the morgue's register, which teaches only because the evidence is true.
- **FR-021**: Design/grounding obligations ride the PR: `node scripts/check-tui-design.mjs
  --changed` (this branch touches `internal/tui/` for the resolver export) with every
  flagged page re-verified/re-pinned; wiki re-pins for every note whose pinned sources
  change (the `internal/world`/`cmd/promptworld`/`internal/sim`/`internal/store`/
  `internal/tui` notes — see plan D8); `docs/player/` regenerated if the wiki changes;
  merge via `gh pr merge --merge` only.

### Key Entities

- **`world.forked` / `WorldForkedPayload`** — the new lineage event (`internal/sim/state.go`
  payload block, beside `WorldCreatedPayload`): `{parent_name, parent_seed,
  parent_created_at, fork_tick, fork_seq}`. Reducer no-op; the authoritative provenance
  record. Full shape in data-model.md.
- **`Manifest.Lineage`** — additive `omitempty` manifest block `{parent,
  parent_created_at, fork_tick}` (`internal/world/world.go`): the offline mirror.
- **Fork ceremony** — `world.Fork(srcDir, destDir, newName)` (`internal/world/fork.go`,
  new; the `Migrate` ceremony's sibling): resolve boundary → fresh log prefix + snapshot +
  lineage event + meta (seed/format/spend keys) → manifest with new identity → sidecar
  copy. Returns a `ForkResult` for the CLI summary.
- **Shared resolver, exported** — `tui.ResolveRubricFacts(state, def, pass)` +
  `tui.RenderReportCard(title, facts, mode, width)` + exported aliases for the fact/mode
  types: the spec-072 contract made replica-parametric so the CLI duel is the fourth
  consumer of the ONE switch (postmortem, ceremony, console card, duel).
- **Duel report** — compare's output model: header (names, lineage, liveness), per-world
  scorecard (resolver facts + plain outcome line), divergence record (first divergent
  story event or none), interleaved chronicle entries. Shape in data-model.md.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001** (board AC #1): e2e — fork of a run-and-stopped world starts alongside its
  parent; both answer `status`; `ps --json` lists both running; the fork was created in
  under ~5s for a v1-scale world (<1M events).
- **SC-002** (board AC #2): the fork's log carries exactly one `world.forked` with the
  correct payload at `(boundary.tick, boundary.seq+1)`; the manifest `lineage` block
  round-trips through `Open`; a pre-fork world's `world.json` round-trips byte-identically
  (omitempty — no format bump).
- **SC-003** (board AC #4): the three FR-010 determinism proofs pass — genesis replay of
  the fork log reproduces the boundary snapshot hash; fork-vs-parent state hashes match at
  the fork tick; a post-fork RUN world's log independently replays to its own final
  snapshot hash.
- **SC-004** (board AC #7): on a decided duel fixture (one recorded pass, one run-ended),
  compare renders the winner's card all-✓ from its pass and the loser's card with ✗ on the
  failed term and honest backing (`agent.died: N`); a sweep of the rendered output finds
  no raw enum token (`in_progress`, bare exercise/stage ids in grade position); the
  verdict rows are byte-identical to what the postmortem overlay renders for the same
  state (one resolver, proven by test).
- **SC-005** (board AC #3): the divergence marker lands at the first differing story
  event (machinery classes excluded — a fixture differing only in `cog.*`/`llm.*`/
  `daemon.*`/`clock.*` events renders NO divergence); interleaved chronicle entries carry
  correct world labels in tick order; the zero-divergence line is pinned by test.
- **SC-006** (board AC #5): a fork of a world with recorded month spend opens its meter
  at the parent's spend (test: parent meta `llm_spend_<month>` = X → fork →
  `NewMeter` on the fork's store reports spent = X, per-provider attribution intact).
- **SC-007**: `go test ./...` green (including `TestCatalogSweep` with the new event);
  `node scripts/check-tui-design.mjs --changed` passes; `node scripts/check-merge-drift.mjs
  pr` exits 0 from the worktree before the PR opens.

## Assumptions

- TASK-149 (spec 072) is merged (PR #113, f78358a) — `resolveReportCardFacts` and honest
  ✗ grading exist to be shared; verified against `internal/tui/reportcard.go` on main.
- `Store.AppendEvents` assigns contiguous seqs from `lastSeq+1`: streaming the parent's
  prefix in order into a fresh store reproduces seqs 1..N exactly. `wall_time` is
  observability metadata excluded from determinism comparisons, so its re-stamping (or
  preservation — implementer's choice, research R1) does not affect any proof.
- `validateMeta` at daemon boot checks `seed` and `format_version` only (never name), so
  a fork with a new manifest name and stamped meta boots cleanly.
- The reducer treats unknown and `daemon.*` types as recorded-history no-ops (its doc
  contract) — an explicit `world.forked` no-op arm is still added for self-documentation
  and so the type is never "unknown" to future totality checks.
- Compare's read of a RUNNING world's `world.db` from a second process is safe under WAL
  (readers don't block the single writer); the rendered header states the read is
  as-of-last-commit.
- The `agents/` directory currently holds flat files for future features (may be empty);
  copying its contents verbatim is correct and cheap.

## Out of Scope (documented follow-ons, not silent omissions)

- **Phase 2 — the shareable HTML retelling** (the "Boatmurdered move": two chronicles as
  one artifact, same renderer family). Explicitly the next rung after this spec, per the
  2026-07-25 D7 and 2026-07-26 decision 3 rescopes. Nothing in this spec's design may
  preclude it (the duel report model is its input — noted in data-model.md).
- **Dual side-by-side live TUI** — deferred post-v1 (decision 3).
- **Mid-log / chosen-snapshot forking** (`--at <tick>` / `--at snapshot-N`): documented
  semantics decision — v1 is latest-snapshot only; the ended-world edge case above is the
  recorded motivation for the follow-on.
- **Merging forks** — never (FR-014; doctrine, not a deferral).
- **`compare --json`** — no AC demands it; the human duel card is v1.
- **A chronicle narration line for `world.forked`** ("the world split…") — charming,
  unfunded; the digest line (FR-009) is the v1 rendering.
- **Cost-attribution refinement across a duel** — recorded limitation (FR-013), owned by
  the budget doctrine's own backlog, not this spec.

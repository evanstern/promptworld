# Implementation Plan: faith-driven charge regeneration (spec 085)

**Branch**: `085-faith-regen` (task branch: `task-118-faith-regen`) |
**Date**: 2026-07-26 | **Spec**: [spec.md](spec.md) |
**Entities/curve/vocabulary**: [data-model.md](data-model.md) (normative) |
**Decisions**: [research.md](research.md)

## Summary

Four pieces on one branch: (1) event-sourced faith —
`State.Faith *FaithState` (nil = genesis 50), the `faith.changed`
executor sweep scanning the tick batch for the five source events
(directive fulfilled/expired, deaths, prophecy terminals), the clamping
reducer arm; (2) the regen rewrite — the single boundary check becomes a
pure band function of (faith score, scenario presence), the forsaken row
carrying the DECIDED spiral posture (scenario: no regen; ambient: 24h
floor) with its one-table reversal lever; (3) the prophecy channel —
`prophesy` tool (1 charge), the uncancellable `Prophecy` entity, the
closed claim predicate table, the verification sweep (fulfil before
fail), companion omen/report memories; (4) the visible surface — two
wire fields, the strip's §4 faith segment + effective-cadence forecast,
four digest rows, the `first-faith-event` lesson (spec-077 rider), the
`faith.*` rubric ban. Cross-cutting: rebase taxonomy, replay
byte-identity (new lifecycle + pre-085 logs + genesis-band schedule
equivalence), TUI design page re-verification, wiki re-pins + player
docs in-branch.

## Technical Context

**Language**: Go. **Modified packages** (cross-package, reducer
doctrine, doctrine-adjacent regen behavior — the recorded Opus 4.8 tier
is justified): `internal/sim` (new `faith.go` + `prophecy.go` or one
file; `executor.go`; `loop.go` whitelist; `miracles.go` rebase;
`state.go` Apply dispatch), `internal/tool` (`registry.go`: `prophesy`,
`observableEventTypes` +3), `internal/guardian` (handler, id minting,
turn-prompt sections), `internal/ipc` (`protocol.go`, `server.go`: two
clock fields), `internal/tui` (`views.go` strip, `grammar.go` digest
rows, `lessons.go` catalog row). **Untouched by design**:
`internal/cognition` and `internal/mind` (faith touches no cadence and
no villager prompts — FR-015), `internal/clock`, `internal/llm`, the
hail/scene machinery, every existing reducer arm's semantics, the
directive/designation layer (consumed via its named contract only). **No
format bump** (`omitempty` state additions — the spec-029/084
precedent); **no new RNG purposes**; **no tuning.json change**
(dial-ready constants only).

## Constitution Check (v1.2.0)

- **I. Artifact-grounded** — PASS: TASK-118 card (Wave-3 ratification +
  2026-07-26 realignment + spec-077 rider recorded), signed-off sweep
  runbook, this spec dir; every decision cites file:line evidence
  (research.md); the AC #4 operator checkpoint is satisfied by FR-005
  being IN the spec (the orchestrator surfaces it).
- **II. One task, one PR** — PASS: one branch (`task-118-faith-regen`,
  worktree `.worktrees/task-118`), one PR; phases are internal
  breakdown.
- **III. Gates** — PASS: claim gate passed (CLAIM.md on this number);
  pr gate + `check-tui-design --changed` + `TestCatalogSweep` + lessons
  taxonomy + rubric hygiene + `go test ./...` are the choke points;
  spec-bridge mirrors phases (link is the orchestrator's act — AC #1).
- **IV. Grounding freshness** — PASS (planned): Phase 7 re-pins every
  wiki note whose sources this branch touches and regenerates
  `docs/player/` in-branch; merge is `gh pr merge --merge`; NO rebases
  anywhere — freshen by merging main INTO the branch (TASK-160/161
  landing laws; sibling lanes may be in flight).
- **V. Model tiers** — PASS: Opus 4.8 via `spec-implementer` (recorded
  on TASK-118: "reducer doctrine… doctrine-adjacent by definition").
  Planning/gating stays on Fable 5.

## Design decisions (file-level)

- **D1 — `internal/sim/faith.go`** (new): `FaithState`,
  `FaithGenesis`, the delta table constants (one home, data-model §3),
  `FaithChangedPayload`, `(*State).FaithScore()`, exported
  `FaithRegenCadenceTicks(score, scenario)` with the band table
  (data-model §6), the `faith.changed` reducer arm
  (validate-not-clamp on reason/sign, clamp fold, materialize-on-first),
  and `faithEvents(s, batch, nextTick)` — the batch-scanning sweep
  (run-end detector idiom) with the cannot-move emission gate.
- **D2 — `internal/sim/prophecy.go`** (new): `Prophecy` +
  `ProphecyClaim`, `GuardianProphecyCap`, normalized-claim equality,
  the per-kind fulfil/fail predicate functions (data-model §5; used by
  BOTH sweep and reducer arms — the spec-084 shared-predicate law), the
  `prophecy.declared` door (full table in data-model §4) and the
  terminal arms (re-validate then one-way transition, the
  `transitionGuardianOrder` shape), retention prune, and the
  verification sweep `prophecyEvents(s, nextTick)` incl. companion
  `OriginReport` memories (situated constructors).
- **D3 — `internal/sim/executor.go`**: regen check rewritten per
  data-model §6 (cadence 0 short-circuits); `prophecyEvents` placed
  with the other guardian sweeps (after the directive sweeps — a
  designation fulfilling at T yields directive.fulfilled at T+1 and any
  dependent prophecy fulfil reads designation status directly, so
  prophecy placement needs no extra lag rule, only a FIXED position);
  `faithEvents` placed after ALL faith-source emitters (needs
  heartbeat, gru, directive + prophecy sweeps) and before
  `scenarioRubricEvents`/run-end.
- **D4 — `internal/sim/loop.go`**: ONE whitelist entry
  (`prophecy.declared`) with the doctrine comment; `faith.changed` and
  the prophecy terminals deliberately absent (executor-emitted class).
- **D5 — `internal/sim/miracles.go`** (`rebaseTicks`): active-prophecy
  `DeadlineTick` SHIFT arm (clone of the Directive arm) + KEEP doc
  comments (data-model §9).
- **D6 — `internal/tool/registry.go`**: `prophesy` appended to
  `guardianTools` (Gate Charge 1, params per data-model §4;
  `claim_kind` Enum; text under the existing `TextCapBytes` anchor);
  `observableEventTypes` +3 (`prophecy.declared/fulfilled/failed`);
  guidance via `GuardianToolGuidance` (described ≡ declared).
- **D7 — `internal/guardian`**: `handleProphesy` — target resolution
  (the send_omen name-list/"everyone" resolution), claim assembly +
  kind-conditional param validation (partial-args refused), `pro-` id
  minting (`nextOrderID` clone), `InjectSocial` landing with the
  companion `OriginOmen` memory batch, door rejection →
  `rejected_gate` counsel; `turn.go`: faith score (in-fiction wording)
  + `writeProphecies` prompt section (the `writeStandingOrders`
  shape).
- **D8 — `internal/ipc`**: `ClockStatus.Faith *int` +
  `FaithRegenTicks int64` (data-model §7), served in `server.go`'s
  status path from `FaithScore()`/`FaithRegenCadenceTicks` — the sim
  function is the single home; the daemon never re-derives bands.
- **D9 — `internal/tui`**: `guardianStripView` — fourth segment
  (`faith N` / `faith —` per pointer), forecast switched to the wire
  cadence with the legacy-constant fallback and the cadence-0 omission;
  `grammar.go` +4 in-fiction rows; `lessons.go` `first-faith-event`
  row (trigger `faith.changed`, tier mechanics, skin tokens,
  direction-neutral copy, pointer at the strip/guardian tab).
- **D10 — docs**: `docs/design/tui/panels/guardian-strip.md` §4
  reconciled from "reserved" to shipped (populated form `faith N`,
  dashed = older-daemon skew, forecast rule) + any page
  `check-tui-design --changed` names; wiki + player docs per Phase 7.

## Testing strategy

- **Reducer/fold**: `faith.changed` arm table (reason domain, sign,
  clamp both ends, materialize-on-first, nil-accessor genesis);
  prophecy door rejection table (data-model §4: already-true,
  duplicate, TTL, targets, cap, text, kind-conditional fields);
  terminal races land exactly one status.
- **Sweeps**: faith sweep per-source emission + batch-order determinism
  + cannot-move gate (score 100/0 edges, same-batch pileup); prophecy
  verification per predicate-table cell incl. survives fail-fast,
  late-truth-after-failed, cancelled-designation-fails-at-deadline;
  once-only (the flips-non-active argument); ended-world silence.
- **Regen**: `FaithRegenCadenceTicks` table (band × posture);
  boundary/off-boundary drive per band (the `TestChargeRegen` clone,
  `guardian_test.go:144-168`); genesis-band schedule byte-identity vs
  the pre-085 constant; scenario-forsaken emits nothing over a
  multi-day drive; ambient-forsaken emits exactly at 24h boundaries.
- **Replay**: from-genesis byte-identity over a fixture log carrying
  every faith reason + full prophecy lifecycle (SC-001); pre-085
  fixture log replays byte-identically (no retroactive faith).
- **Firewall/provenance**: companion memories carry the stamped
  Origins (`OriginOmen` declared / `OriginReport` terminal);
  `DirectPerception` classifications asserted; prophecy text reaches
  villagers only via recorded memories (the firewall audit precedent).
- **TUI**: strip render states (populated / dashed / forecast omitted
  at cadence-0 and full bank / drop order); `TestCatalogSweep`;
  lessons taxonomy flip; rubric-hygiene `faith.` ban;
  `render_test.go` absence pin flipped to presence.
- **Composition**: a `monitor_and_act` watch on `prophecy.failed`
  triggers through unmodified `matchOrders` (enum-only observability
  proof, the SC-003 shape of spec 084).

## Wiki re-pin set (Phase 7; the pr gate is the authority)

Touched sources → notes expected to re-pin (from `sources:`
frontmatter): `internal/sim/executor.go` → `executor*`,
`guardian-orders`, `mental-map-propagation`; `internal/sim/guardian.go`
(if touched for constants) + new `faith.go`/`prophecy.go` →
`guardian`, `guardian-orders`, `event-types-guardian-orders`,
`sim-state-world-fields`, `sim-state-apply-world`;
`internal/sim/loop.go` → `sim-loop`, `sim-loop-injection-doors`;
`internal/sim/miracles.go` → `guardian-miracle-rebase-taxonomy`;
`internal/tool/registry.go` → `tool-registry*`;
`internal/guardian/*` → `guardian*`; `internal/ipc/*` →
`ipc-protocol`/`daemon-lifecycle` family; `internal/tui/*` →
`tui-guardian-strip`/`tui-chronicle-feed`/`tile-registry` family +
`event-types` family row additions. Plus a NEW note (e.g.
`guardian-faith`) for the faith economy + prophecy verification,
indexed under the guardian family. `docs/player/` regenerated
in-branch (`node .claude/skills/player-docs/scripts/check-freshness.mjs
--check` is the gate's probe).

## Risks / mitigations

- **Regen behavior change on live worlds** (deaths now slow regen):
  deliberate — the feature; bounded by the genesis-band equivalence
  (untouched worlds identical) and the ambient floor. The band/delta
  table is one home for retuning.
- **Spiral too sharp / too soft**: constants are normative defaults in
  one promoted-dial-ready table; FR-005 records the reversal lever;
  play evidence drives retuning (the spec-048 promotion path exists
  when earned).
- **Prophecy claim vocabulary too thin**: four kinds cover
  build/growth/survival claims; the discriminated `ProphecyClaim` and
  the one predicate table make each new kind a one-row + one-predicate
  change (recorded as the extension seam).
- **Strip version skew**: pointer-nil dashed state + legacy-constant
  forecast fallback keep an old-daemon pairing honest (never invented
  zeros).
- **Merge drift vs sibling lanes**: freshen by merging main INTO the
  branch only (never rebase); `check-merge-drift pr` before the PR;
  board moves at root only (TASK-161 board-sync exception).
- **Scope creep toward agentization/missions/villager coupling**:
  FR-015 guards; `internal/cognition`/`internal/mind` untouched is a
  review obligation recorded in the PR body.

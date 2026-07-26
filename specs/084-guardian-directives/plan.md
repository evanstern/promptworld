# Implementation Plan: guardian directives and designations (spec 084)

**Branch**: `084-guardian-directives` (task branch:
`task-157-guardian-directives`) | **Date**: 2026-07-26 |
**Spec**: [spec.md](spec.md) | **Entities/rung**: [data-model.md](data-model.md) |
**Events**: [contracts/events.md](contracts/events.md)

## Summary

Three pieces on one branch: (1) `survey_site`, a charge-free Read tool
(registry declaration + turn-side deterministic fact sheet, the
explain/targeting-digest pattern); (2) event-sourced Designations —
entity + reducer arms + executor fulfillment sweep + all-villager
place-fact grant + tile-registry map rendering + guardian tools;
(3) hard Directives — entity + injected/executor-emitted lifecycle +
`observableEventTypes` growth + the `directive` decision-context block +
the DIRECTIVE reflex rung between SURVIVAL and PREP + interruption-resume
proof. Cross-cutting: digest rows (`TestCatalogSweep`), rebase-taxonomy
entries, replay byte-identity, TUI design pages, wiki re-pins + player
docs in-branch.

## Technical Context

**Language**: Go. **Modified packages** (cross-package — the recorded
Opus 4.8 tier is justified): `internal/sim` (new `designations.go` +
`directives.go` or one `plans.go`; `policy.go`; `executor.go`;
`loop.go`; `miracles.go`; `mentalmap.go`), `internal/tool`
(`registry.go`, `derive.go`), `internal/target` (one exported bare-locus
entry point), `internal/guardian` (`toolcalls.go`, `turn.go`, new
`survey.go`/`designations.go`), `internal/mind` (`context.go`),
`internal/tui` (`tiles.go`, `grammar.go`, `views.go` map path).
**Untouched by design**: `internal/bundle` (grammar matrix unchanged),
the hail/scene/pause machinery (FR-013's zero-interruption-code
obligation), every existing reducer arm, `internal/clock`,
`internal/llm`. **No format bump**; no new RNG purposes; no wire/IPC
change (TUI renders designations from the replica state).

## Constitution Check (v1.2.0)

- **I. Artifact-grounded** — PASS: TASK-157 card (operator hardness
  ruling recorded), signed-off sweep runbook, this spec dir; every
  design decision cites file:line evidence (research.md).
- **II. One task, one PR** — PASS: one branch
  (`task-157-guardian-directives`, worktree `.worktrees/task-157`), one
  PR; phases below are internal breakdown.
- **III. Gates** — PASS: claim gate passed (CLAIM.md stub on this
  number); pr gate + `check-tui-design --changed` + `TestCatalogSweep` +
  `go test ./...` are the choke points; spec-bridge mirrors phases.
- **IV. Grounding freshness** — PASS (planned): Phase 8 re-pins every
  wiki note whose sources this branch touches and regenerates
  `docs/player/` in-branch; merge is `gh pr merge --merge`; NO rebases —
  sibling lanes (task-133, later task-118/17) are in flight, so freshen
  by merging main INTO the branch only (TASK-160 all-by-merge law).
- **V. Model tiers** — PASS: Opus 4.8 via `spec-implementer` (recorded
  on TASK-157: cross-package + reflex-arbitration doctrine + injection
  door + decision context). Planning/gating stays on Fable 5.

## Design decisions (file-level)

- **D1 — `internal/target/target.go`**: export
  `ParseLocus(s string) (Address, error)` — trims, dispatches on
  `..`/`->`/point exactly like `parseLocus` (reuse it with an empty/new
  designation-neutral class), returns the same normalized `Address`;
  `Tiles()` already serves it. Unit tests beside the existing table
  (`target_test.go`) incl. the stdlib-only import pin staying green.
  Bundle behavior byte-identical (nothing else in the package changes).
- **D2 — `internal/sim` entities** (new file(s), e.g. `plans.go`):
  `Designation`/`Directive` structs, caps/TTL constants
  (`GuardianDesignationCap`, `GuardianDirectiveCap`; TTL reuses
  `GuardianOrderTTLMinDays/MaxDays`), `applyPlan` reducer dispatch for
  the seven types (wired from `Apply`'s switch beside `applyGuardian`),
  a shared one-way transition helper (the `transitionGuardianOrder`
  shape) and a generalized retention prune, the fulfillment predicate
  functions (data-model §6) used by BOTH arm and sweep, and the
  `designation.placed` place-fact grant (`mentalmap.go`:
  `"designation"` joins the kind vocabulary + `factHorizon` entry).
- **D3 — `internal/sim/loop.go`**: four whitelist entries
  (`designation.placed/cancelled`, `directive.issued/cancelled`) with
  the standard doctrine comments; nothing else (dry-run machinery is
  generic).
- **D4 — `internal/sim/executor.go`** (`stepEvents`, beside the
  order-expiry sweep at :59): designation-fulfillment sweep, then
  directive-fulfillment, then directive-expiry (incl. the all-targets-
  dead clause) — each the once-only `charge_regenerated` idiom.
- **D5 — `internal/sim/policy.go`**: `directiveDecision` per data-model
  §8, called between `survivalDecision` and the `prepYields` consult;
  the `reflexBuildable` routing table; `heed_directive` goal wired
  through the executor's intent machine (instant-on-arrival) and the
  goal/duration mirrors (`internal/tool` sim-duration mirror + drift
  test, the existing pattern).
- **D6 — `internal/sim/miracles.go`** (`rebaseTicks`): the active-
  directive `ExpiresTick` SHIFT arm (clone of the GuardianOrders arm at
  :327-334) + taxonomy doc comments for the KEEP fields.
- **D7 — `internal/tool/registry.go` + `derive.go`**: five tools
  appended to `guardianTools` (data-model §9); `observableEventTypes`
  +4; `guardianToolDesc` entries (acting four + `survey_site` under the
  read path); scalar `Params` suffice — no authored schema needed
  (no arrays; `kind`/`structure_kind` are Enums, loci are Text).
  Buildable-structure vocabulary for `structure_kind` is a hand-carried
  mirror with a guardian-side drift test (the `clockSpeeds`/
  `placeFactKinds` pattern — tool is a leaf).
- **D8 — `internal/guardian`**: `survey.go` (`buildSurveySheet` — pure
  over (args, state mirror, `worldmap.Map`); fixed iteration order;
  bounds → repairable miss), `handleSurvey` Read dispatch (the
  `handleExplain` shape, `toolcalls.go:116`); designation/directive
  handlers (`handlePlaceDesignation`/`handleCancelDesignation`/
  `handleIssueDirective`/`handleCancelDirective`) wrapping locus parse
  (D1), name resolution (the send_omen target resolution), id minting
  (`nextOrderID` clone with `dsg-`/`dir-` prefixes), and `InjectSocial`
  landing with door rejections mapped to `rejected_gate` counsel;
  `turn.go`: `writeDesignations`/`writeDirectives` prompt sections (the
  `writeStandingOrders` shape) + mirror maintenance in the absorb path
  if needed for id minting.
- **D9 — `internal/mind/context.go`**: `renderDirective` + the
  `fixedBlocks` insertion (`{Name: "directive", Priority: neverDrop}`
  between `plan_echo` and `known_places`); `renderKnownPlaces` handles
  the `designation` fact kind as a landmark.
- **D10 — `internal/tui`**: three registry rows in `tiles.go` (glyph
  proposals: site `◇`, wall segment `┄`, zone perimeter `░`; semantic16
  tokens — designations are meaning, not material) with the
  render-beneath-real-entities rule in the map tile resolution; seven
  digest-grammar rows in `grammar.go`; `TestCatalogSweep` +
  tile sweep/identity tests updated.
- **D11 — docs**: `docs/design/tui/panels/map.md` (+ any page
  `check-tui-design --changed` names, e.g. `pages/guardian-console.md`)
  amended + re-pinned same-PR; wiki + player docs per Phase 8.

## Testing strategy

- **Reducer/lifecycle**: table-driven arm tests per contracts/events.md
  validation row; race pairs (cancel vs fulfil vs expire) land exactly
  one terminal; prune determinism; place-fact grant (incl. map-less
  agent).
- **Sweeps/replay**: fixture drive placing → building → fulfilling →
  expiring; from-genesis replay byte-identity (SC-004); ended-world
  silence.
- **Reflex**: the spec-062 matrix pattern (`reflex_matrix_test.go`
  precedent) for rung cells (survival preempts / directive preempts
  prep / routing per kind / orphan falls through); directive-free
  parity drive (SC-006); the hail interruption-resume drive (SC-002,
  asserting no diff in hail machinery is a review obligation recorded
  in the PR body).
- **Door/firewall**: dry-run rejection table (dead target, unknown/
  non-active designation, TTL, caps, occupancy, bounds, form); the
  firewall audit extension (guardian prose reaches villager prompts
  only via block/memory — the `fiction_prompt_test.go`/firewall-audit
  precedent); atomic companion-memory batch.
- **Composition (AC #7)**: `monitor_and_act` watch on
  `directive.fulfilled` triggers through unmodified `matchOrders`
  (SC-003).
- **Survey**: byte-identity, no-append, no-charge, acting-cardinality
  audit, bounds miss (SC-005).
- **TUI**: registry sweep/identity tests; `TestCatalogSweep`; map
  precedence (entity wins tile).

## Wiki re-pin set (Phase 8; the pr gate is the authority)

Touched sources → notes expected to re-pin (from `sources:` frontmatter):
`internal/sim/policy.go` → `reflex-policy`, `reflex-survival-rungs`,
`reflex-prep-arbitration`, `reflex-goal-resolution*`;
`internal/sim/loop.go` → `sim-loop`, `sim-loop-injection-doors`,
`guardian-orders`; `internal/sim/executor.go` → `executor*`,
`guardian-orders`, `mental-map-propagation`; `internal/sim/guardian.go`
(if touched) + new sim files → `guardian-orders`,
`event-types-guardian-orders`, `sim-state-world-fields`,
`sim-state-apply-world`; `internal/sim/mentalmap.go` →
`mental-map-model`; `internal/sim/miracles.go` →
`guardian-miracle-rebase-taxonomy`, `guardian-miracle-mechanics`;
`internal/tool/registry.go`/`derive.go` → `tool-registry*`,
`mental-map-propagation`, `guardian-orders`; `internal/guardian/*` →
`guardian*`, `explain-tutor-guide`; `internal/mind/context.go` →
`decision-context`, `context-block-inventory`, `mental-map-propagation`;
`internal/tui/*` → `tile-registry`, `tui-map-view`,
`tui-chronicle-feed`. Plus a NEW note (e.g. `guardian-designations`)
for the plan layer, indexed under Inference & minds / the guardian
family, and an `event-types` family row addition
(`event-types-guardian-orders` or a new sibling). `docs/player/`
regenerated in-branch.

## Risks / mitigations

- **Reflex-rung regressions** (survival balance): the rung is inert
  without directives (SC-006 parity drive) and positioned after
  survival; the 062 matrix + degraded-mode survival suites gate.
- **Prompt-byte drift**: empty-block omission guarantees directive-free
  byte-identity; golden prompt tests pin it.
- **Merge drift vs sibling lanes** (task-133 in flight): freshen by
  merging main into the branch (never rebase); `check-merge-drift pr`
  before the PR.
- **Scope creep toward missions/faith**: FR-016 guards; the TASK-118
  seam is a named payload, nothing more.

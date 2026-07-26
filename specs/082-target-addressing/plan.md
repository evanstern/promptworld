# Implementation Plan: Target-addressing grammar for bundle effects

**Branch**: `082-target-addressing` (task branch: `task-97-target-addressing`) | **Date**: 2026-07-26 | **Spec**: [spec.md](spec.md) | **Grammar**: [data-model.md](data-model.md)

## Summary

One new leaf package (`internal/target`) owning the address grammar (parse
+ normalize + deterministic tile enumeration + error taxonomy); the bundle
effect compiler (`internal/bundle/effects.go`) extended to dispatch
class+tile targets for `move_entity`/`remove_entity`/`grant_item` through
the data-model.md form matrix, producing miracle-door-identical payloads;
narrow exported presence probes on `sim.State`; the spec-036 manifest
contract and `docs/bundles.md` amended (including the named TASK-157
designation-seam section); fixtures for every accepted form, every error
class, byte-identity, and replay. No reducer changes, no new event types,
no TASK-157 tools.

## Technical Context

**Language**: Go. **New package**: `internal/target` — stdlib only
(`strings`, `strconv`, `fmt`), importable by `internal/tool` later
(TASK-157) without cycles (`bundle` → `tool` and `sim` → `tool` exist
today, so the grammar cannot live in either). **Modified packages**:
`internal/bundle` (effects.go compile path + tests/fixtures; script.go's
`effectFromDict` untouched structurally — target stays a plain string),
`internal/sim` (exported one-per-tile presence probes only; no reducer
arm, no payload struct, no state field changes). **Contracts**:
`specs/036-scriptable-agent-tools/contracts/bundle-manifest.md` (amend —
036's contract stays the normative manifest surface; 082's data-model.md
is the normative grammar it references), `docs/bundles.md` (authoring
guide). **Untouched by design**: `internal/guardian/miracle_batch.go`,
`internal/sim/miracles.go` reducer arms, `internal/tool` (the seam is
designed and contract-named, not wired), `effectEventType`,
`injectSocialWhitelist`.

## Constitution Check (v1.2.0)

- **I. Artifact-grounded** — PASS: TASK-97 card (realignment note) + the
  signed-off faith-directives sweep runbook + this spec dir; decisions
  below cite code/contract evidence read from the worktree.
- **II. One task, one PR** — PASS: `task-97-target-addressing` branch in
  `.worktrees/task-97`, one PR; spec phases are internal breakdown.
- **III. Gates over assertions** — PASS: claim gate already run for
  `082-target-addressing` (CLAIM.md stub); pr gate + player-docs probe +
  `go test ./...` are the choke points in Phase 6; spec-bridge mirrors
  phases.
- **IV. Grounding freshness** — PASS (planned): `docs/wiki/bundle-tools.md`
  pins `internal/bundle/effects.go` (touched) and carries the TASK-97
  limitation prose — content update + re-pin ride THIS branch;
  `docs/player/` regenerated in-branch when the wiki changes; sim-source
  notes (`sim-state-*`, `guardian-miracle-*`) re-pin if the exported-probe
  edit lands in a pinned file — the pr gate names the exact set and is the
  authority. Merge-commit-only (`gh pr merge --merge`).
- **V. Model-tiered** — **Opus 4.8** per the board card's recorded tier
  choice: cross-package (new leaf package + bundle compiler + sim exports),
  and the grammar binds two consumers (bundle effects + TASK-157
  designations) — architecture-shaping under the rubric. Dispatched via
  `spec-implementer` with `model: opus`; planning/gating stays on Fable 5.

**Post-Phase-1 re-check**: PASS — the one structural addition (a new
package) is justified in Complexity Tracking; everything else extends
existing seams.

## Design decisions

- **D1 — Parser package** (`internal/target/target.go` + `target_test.go`):
  `Parse(s string) (Address, error)` implementing data-model.md §1–2
  exactly — reserved-prefix rule, per-form normalization (rect min/max
  corners; line endpoint order preserved, axis-aligned enforced), and
  typed errors exposing the taxonomy class (`syntax`/`class`/`form` at
  parse time; `bounds`/`unresolved` belong to consumers) so bundle can wrap
  them as T5 `ruleErr`s without string-matching. `Address.Tiles() []Tile`
  per §2 (row-major rect; endpoint-order line; single-tile point). A test
  asserts the package's import list is stdlib-only (SC-004's leaf-safety
  pin — the TASK-157 seam's structural guarantee).
- **D2 — Compiler dispatch** (`internal/bundle/effects.go`): `compileOne`'s
  `move_entity`/`remove_entity`/`grant_item` arms parse `e.Target` via
  `target.Parse` (substitution already done by `ExpandTemplates` — parser
  sees resolved strings, including in script mode where `effectFromDict`'s
  output feeds the same `CompileEffects`). Dispatch per the data-model.md
  §4 matrix: villager-designating forms resolve through the existing
  `villager()`/`villagerIndex` helpers extended with `VillagerAt` for the
  point form; `structure@`/`pile@` probe presence via D4's sim exports and
  fill `EntityMovedPayload`/`EntityRemovedPayload` with `Class` + source
  tile — the `BuildMiracleBatch` shapes verbatim; `terrain@` bounds-checks
  only (removability stays reducer-side). ❌ cells produce `form` errors;
  rect/line messages name the designation reservation (TASK-157). Existing
  bare-name behavior is preserved bit-for-bit (reserved-prefix rule).
- **D3 — Error surfacing**: all target failures become
  `ruleErr("T5", "effect %d …")` naming index, field, and offending
  address (data-model.md §5 message shapes; SC-005 table test). Rejection
  stays whole-invocation and charge-free via the unchanged
  `CompileEffects` → `InjectSocial` dry-run pipeline.
- **D4 — Sim presence probes**: export narrow, deterministic, read-only
  helpers on `sim.State` for the compiler — presence of a structure/pile
  on a tile plus map dims for bounds (today `structureIndexAt`/`pileAt`
  are unexported and dims live behind `s.m`). Shape mirrors `VillagerAt`'s
  doc discipline ("both doors resolve through this one helper so they can
  never disagree"). NO reducer arm, payload, or state-field changes.
- **D5 — Contract + guide amendments**: bundle-manifest.md gains "Target
  addressing" (grammar summary pointing at 082's data-model.md as
  normative, the §4 matrix, error behavior, compat/reserved-prefix rule,
  `text`-kind guidance for address args) and "Designation addressing
  (TASK-157 seam)" (parser package + leaf-safety, rect/line normalization
  + enumeration order, bundle-reserved status). `docs/bundles.md` gets the
  author-facing version with examples. Compatibility note per the
  contract's own rule: this is an additive value-space extension of an
  existing key — old manifests unaffected.
- **D6 — Fixtures** (FR-010): a new fixture world under
  `internal/bundle/testdata/worlds/` with a structure, a chest with
  contents, a pile, and a tree — declarative tool(s) exercising
  structure/pile move + structure/pile/terrain remove (literal and
  `{args}` forms) and a scripted tool composing `class@x,y` strings;
  table-driven error tests per taxonomy class; byte-identity test
  compiling each class+tile effect and comparing payload bytes against
  `guardian.BuildMiracleBatch` (dogfood-move precedent — note: identity is
  on the MAIN event; the door adds no perception memory for
  structure/pile/terrain anyway, and bundle villager moves keep their
  existing no-memory behavior); replay determinism via the existing
  `replay_test.go` pattern including bundle-dir deletion.
- **D7 — Wiki + player docs, in-branch**: update
  `docs/wiki/bundle-tools.md` (grammar paragraph replaces the "TASK-97
  limitation" operational note; sources gain `internal/target/target.go`),
  re-pin to a branch commit; re-pin any other note the pr gate names
  (candidates: notes pinning `internal/sim/state.go`/`agents.go` if D4
  lands there). Regenerate `docs/player/` (wiki changed ⇒ player-docs
  probe must pass: `node .claude/skills/player-docs/scripts/check-freshness.mjs --check`).
  `docs/wiki/event-types.md` needs nothing (no new types) —
  `TestCatalogSweep` stays green by construction; if implementation ever
  introduces an event type, the catalog fixture + event-types.md row are
  REQUIRED in the same branch (tasks Phase 6 carries the conditional gate).
- **D8 — Proof commands**: `go test ./internal/target/ ./internal/bundle/`;
  `go test ./...`; `gofmt -l` clean; player-docs probe (above);
  `node scripts/check-merge-drift.mjs pr` from `.worktrees/task-97`
  (exit 0 — includes wiki-repin + player-docs findings, no bypass);
  `gh pr merge --merge` only (in-branch pins are branch hashes; squash
  stales them — observed hazard). TASK-160: every main-bound change,
  INCLUDING post-merge bookkeeping, is authored on a branch in a worktree
  and lands by merge — nothing commits directly at root.

## Complexity Tracking

| Addition | Why no simpler alternative suffices |
|---|---|
| New package `internal/target` | The grammar's second consumer is `internal/tool` (TASK-157 designation params). `tool` is a leaf: it cannot import `bundle` (cycle — `bundle` imports `tool`) or `sim` (existing law; mirrors are hand-carried with drift tests). Putting the parser in `bundle` or `sim` forces TASK-157 to hand-copy the grammar — exactly the drift this project's one-authoritative-source discipline (`MiracleCostsByEvent` precedent) exists to prevent. A stdlib-only leaf is the minimal shape that lets both consumers import one authority. |

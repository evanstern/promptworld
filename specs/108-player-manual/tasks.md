# Tasks: spec 108 — the player's user manual

**Spec**: `specs/108-player-manual/spec.md` · **Plan**: `specs/108-player-manual/plan.md`
· **Board task**: TASK-182 · **Branch**: `task-182-player-manual`

Phases are internal breakdown, not PR boundaries — everything below lands in TASK-182's
single PR.

## Phase 1 — Grounding (read-only)

- [x] **T001** Read `docs/wiki/cli-promptworld.md`, `cli-world-lifecycle.md`,
  `cli-runtime-control.md`, `cli-guardian-ops.md` at their current `verified_against:`
  pins. Extract: the full subcommand set, the name-or-path world-argument rule and
  `PROMPTWORLD_HOME`, exit discipline, and the `metatron`/`miracle` hidden-alias facts.
- [x] **T002** Read `docs/wiki/instance-manager.md`, `world-forking.md`,
  `daemon-lifecycle.md`, `curriculum-ladder.md`, `cognition-estimator-calibration.md`.
  Extract per-command flags and outputs for `ps`, `new`, `fork`, `compare`,
  `divergence`, `start`/`stop`/`status`/`daemon`, `stages`/`teaching`, `calibrate`,
  `llm`. Note explicitly which commands have **no** documented flags.
- [x] **T003** Read `docs/wiki/world-save-directory.md`, `world-save-manifest-fields.md`,
  `world-tuning.md`, `world-tuning-dial-catalog.md`, `skin.md`,
  `guardian-instruction-surface.md`, `governance.md`, `bundle-tools.md`,
  `event-log.md`, `snapshots.md`, and `docs/llm-providers.md`. Build the file inventory
  and assign each file one of the three dispositions.
- [x] **T004** Read `docs/wiki/llm-preflight-detection.md`, `llm-provider-health.md`,
  `cognition-horizon-telemetry.md`, `daemon-boot-recovery.md`, `world-migration.md`,
  `tui-client.md`. For each of the seven required symptoms, identify the concrete
  observable surface a player can check.
- [x] **T005** Record the current pin for every source above — `verified_against:` for
  `docs/wiki/` paths, `git log -1 --format=%H -- <path>` for `docs/llm-providers.md`.
  This list is the source of the meta tags.

## Phase 2 — Pages

- [x] **T006** Create `docs/player/command-reference.html`: canonical skeleton + verbatim
  `<style>`, `generated-by` tag, one `source` tag per T001/T002 source on its own line.
  Content: shared world-argument rule, one-screen summary table of all 21 subcommands,
  then intent-grouped detail sections, then the retired-alias note. Satisfies
  FR-001..FR-005.
- [x] **T007** Create `docs/player/world-files-reference.html` from T003: one entry per
  file with its disposition, what it does, and a link out for deeper reference
  (`docs/llm-providers.md`, `docs/bundles.md`) instead of duplicated depth. Satisfies
  FR-006, FR-007.
- [x] **T008** Create `docs/player/troubleshooting.html` from T004: one section per
  symptom in the player's own words, each with what's happening · what to check · what
  to do. Satisfies FR-008, FR-009.
- [x] **T009** Cross-link the three pages to each other and to the teaching pages they
  complement, and add the back-link to `index.html` each existing page carries.
- [x] **T010** Self-check every new page: exactly one `generated-by` tag; every `source`
  tag alone on its line matching `<path>@<40-hex-lowercase>`; no `<script>`; no external
  `href`/`src`; the `prefers-color-scheme: dark` block present verbatim.

## Phase 3 — Nav

- [x] **T011** Edit `docs/player/index.html`: add a Reference section with the three
  pages in the existing card markup (title + blurb). Add no `source` tags — the script
  rejects `index.html` if it declares any. Satisfies FR-010.

## Phase 4 — Gate wiring

- [x] **T012** `.claude/skills/player-docs/scripts/check-freshness.mjs`: add the three
  slugs to `EXPECTED_PAGES`; update the explanatory comment above it (currently
  "nine -> thirteen"). Satisfies FR-015.
- [x] **T013** `.claude/skills/player-docs/SKILL.md`: add the three pages to the expected
  page set and fix the count prose; add one page→source mapping row per new page; add a
  short editorial note for the reference trio (its tone rule, and D4's link-don't-
  duplicate rule); update the `description:` frontmatter count. Satisfies FR-014.
- [x] **T014** Correct the mapping table's `playing-via-metatron.html` row to the
  `guardian*` notes that page actually declares, replacing the non-existent
  `docs/wiki/metatron.md` / `docs/wiki/metatron-miracles.md`. Satisfies FR-017.

## Phase 5 — Verification

- [x] **T015** `node .claude/skills/player-docs/scripts/check-freshness.mjs --check`
  exits 0 and reports 16 fresh, 0 stale, 0 missing, 0 broken-ref. Satisfies SC-004.
- [x] **T016** `git diff --stat origin/main` lists exactly the six intended files plus
  the spec dir; `git diff origin/main -- docs/player/` shows no existing page changed
  but `index.html`. Satisfies FR-011.
- [x] **T017** Editorial pass for FR-016: read all three pages as a non-engineer. Any
  identifier, package path, or engineering term that is not something the player types
  or opens gets rewritten or cut.
- [x] **T018** `node scripts/check-merge-drift.mjs pr` from the worktree exits 0.
  Satisfies SC-005.

## Notes for the implementer

- `keys-reference.html` is the closest tonal model — pure reference, no lore.
- **Never invent a flag.** If a declared source does not carry it, the page does not
  claim it. Reading `cmd/promptworld/*.go` to check whether a source is complete is
  fine; citing it as a source is not.
- Touch only these six files. Re-saving an existing fresh page — even byte-identically
  — violates FR-011 and the skill's no-op rule.
- Rebases are forbidden repo-wide. If main moves, merge main *into* this branch.

# Feature Specification: The player's user manual — command, world-files, and troubleshooting references

**Feature Branch**: `task-182-player-manual`

**Created**: 2026-08-01

**Status**: Draft

**Input**: TASK-182 — "Promptworld needs a user manual." The operator selected, from
four candidate forms (docs-only reference pages / a single `MANUAL.md` / an in-app
`promptworld manual` subcommand / docs plus in-app), the **reference pages inside
`docs/player/`** option, wired into the player-docs skill and its freshness gate.

## Problem

`docs/player/` (spec 026, TASK-114, TASK-153) is a *teaching* surface: thirteen pages
that walk a new player through a first session, explain the screen, narrate how the AI
behind the village works, and hand out one quickstart per curriculum stage. It has no
*reference* half. Concretely, measured on this repo at the time of writing:

- The binary dispatches **21 subcommands** (`cmd/promptworld/main.go`). `divergence`
  and `daemon` appear on **zero** player pages; `migrate`, `fork`, `compare`, `ps`,
  `tail`, `pause`, `resume`, `speed` and `llm` appear on exactly **one** each, always
  mid-narrative. No page lists a flag. A player asking "what can I type?" has only
  `promptworld help`.
- A world save directory contains files a player is explicitly invited to edit —
  `charter.md` is described in the README as "the game's only player-editable prompt" —
  yet no page enumerates the directory's contents or says which parts are safe to touch.
  `llm.json` has an operator reference (`docs/llm-providers.md`) written for engineers;
  `tuning.json`, `skin.json`, `village_charter.md`, `calibration.json` and bundles are
  described narratively, scattered, and never in one place.
- There is no troubleshooting surface at all. The failure modes this project has
  actually shipped hooks, gates and cards about — brain-dead villagers on a
  misconfigured fresh world (TASK-84), suppression at high speed (spec 007), a daemon
  that reports running but is not (TASK-147, `ps` liveness), a format-version mismatch
  refusing to open a world (TASK-134/147) — have no player-facing symptom→fix path.

The gap is reference, not explanation. A player who has read every existing page still
cannot look anything up.

## Doctrine (from the card and the existing contracts — not re-litigated)

- **This is a docs-only feature.** No Go code changes, no new binary behavior. The one
  non-`docs/player/` change is to the player-docs skill and its freshness script, which
  is what keeps the new pages from rotting.
- **Projection, never assertion** (spec 026 FR-002). Every factual claim on a player
  page is a plain-language projection of a declared, pinned source. A page may not
  assert a fact its sources do not carry, and may not paraphrase a value that must be
  byte-verbatim.
- **Self-contained pages** (spec 026 FR-004). Each page inlines the canonical skeleton
  and the shared `<style>` block verbatim: no external assets, no JavaScript, no shared
  CSS file. Duplication across pages is the intended design.
- **The gate is the anti-rot mechanism.** A page the freshness script does not know
  about is a page that can silently go stale, so a new page is not done until
  `EXPECTED_PAGES` and the skill's mapping table both carry it.
- **Reference tone, player vocabulary.** The reference pages are terse and lookup-shaped
  — tables, symptom rows, one-line descriptions — but they use the same plain vocabulary
  as the teaching pages. "Reference" licenses density, never jargon.

## Design decisions

- **D1 — Three pages, not one.** The three questions ("what can I type", "what are these
  files", "why is it broken") have different shapes and different source sets. One
  combined page would be long, would mix a stable command table with volatile
  troubleshooting prose, and would force a single source-tag set spanning everything —
  so any source change would restale the whole manual. Three pages fail independently.
- **D2 — The command reference is the spine.** It documents all 21 dispatched
  subcommands from the CLI wiki notes' pins, grouped by what a player is trying to do
  (make a world / run it / watch it / play it / tune the AI / compare runs), not
  alphabetically and not in dispatch order.
- **D3 — Hidden aliases are documented as hidden.** Spec 052 FR-008 deliberately keeps
  `metatron` and `miracle` registered but out of the usage text. The reference names
  them once, as retired compatibility aliases that still work, and directs players to
  `guardian` and `work`. Documenting them is not un-hiding them; a player meeting one in
  an old script needs to be able to look it up.
- **D4 — The world-files page links rather than duplicates.** Following the standing
  editorial rule for `llm-setup-basics.html`, the world-files reference states what
  `llm.json` is and which knobs a player touches, then defers registry-reference and
  migration depth to `docs/llm-providers.md` by link. Same for `docs/bundles.md`.
- **D5 — Troubleshooting rows are evidence-shaped.** Each row is symptom (what the
  player sees) → likely cause → what to check → fix. "What to check" always names a
  real, observable surface (a `promptworld status` field, a `ps` column, an event in the
  feed, a preflight warning), because the project's whole posture is that the world is
  honest about its own state. No row may invent a diagnostic that does not exist.
- **D6 — Nav, not restructure.** `index.html` gains a Reference section listing the
  three pages. Every other existing page stays byte-identical: this feature adds a
  surface, it does not re-cut the existing one.

## User scenarios

### Primary: looking up a command

A player has been running a world for a week and wants to try the same village with a
different charter. They open the command reference, find `fork` in the "compare runs"
group, read that it copies a world so the original is untouched, see the flags, and see
`compare` and `divergence` named next to it as the pair that tells them what diverged.

**Why it matters**: today this player would have to read the README's dev quickstart,
which does not mention `divergence` at all.

### Primary: understanding the world folder

A player opens `~/.promptworld/worlds/demo` and sees a directory of files. They open the
world-files reference, learn which two files are theirs to edit (`charter.md` — their
guardian's instructions; `llm.json` — which models think), which one the village writes
for itself and they should read but not edit (`village_charter.md`), and which are
machine state they should leave alone (the event log, snapshots, the manifest).

### Primary: something is wrong

A player's villagers are wandering and never talking. They open troubleshooting, match
the symptom "my villagers only do basic survival things and never think", find the two
usual causes (no model configured; the model is too slow for the world's speed), and the
check for each — the LLM column in `promptworld ps`, and the suppression line in
`promptworld status`.

### Secondary: the maintainer's scenario

Someone changes the CLI. The next PR that touches a pinned source restales
`command-reference.html`, the pr gate blocks on `player-docs-stale`, and the manual gets
re-projected in the same PR as the change — the same lifecycle every other player page
already has.

## Requirements

### Functional

- **FR-001** `docs/player/command-reference.html` MUST document every subcommand
  dispatched by `cmd/promptworld/main.go` — at time of writing `new`, `migrate`, `fork`,
  `compare`, `ps`, `daemon`, `start`, `stop`, `status`, `ui`, `attach`, `tail`, `pause`,
  `resume`, `speed`, `teaching`, `llm`, `calibrate`, `divergence`, `stages`, `guardian`,
  `work` — each with a plain-language description of what it does for the player.
- **FR-002** The command reference MUST state the shared world-argument rule once: every
  per-world command takes a world **name** (resolved against the worlds home, then the
  known-worlds registry) or an explicit **path**, and MUST name `PROMPTWORLD_HOME`.
- **FR-003** The command reference MUST document each command's flags where its declared
  sources carry them, and MUST NOT invent flags its sources do not carry.
- **FR-004** The command reference MUST name `metatron` and `miracle` as retired,
  still-working compatibility aliases for `guardian` and `work`, marked as not shown in
  the built-in usage text.
- **FR-005** Commands MUST be grouped by player intent, with the group named in plain
  language, and the page MUST open with a one-screen summary table before the detail.
- **FR-006** `docs/player/world-files-reference.html` MUST document every file in a world
  save directory a player may encounter, each labelled with one of three plain
  dispositions: **yours to edit**, **read it, don't edit it**, or **leave it alone**.
- **FR-007** The world-files page MUST cover at least: `llm.json`, `charter.md`,
  `village_charter.md`, `skin.json`, `tuning.json`, `calibration.json`, the world
  manifest, the event log and snapshots, and bundles — deferring `llm.json` registry
  depth to `docs/llm-providers.md` and bundle-authoring depth to `docs/bundles.md` by
  link rather than duplication.
- **FR-008** `docs/player/troubleshooting.html` MUST be organised as symptom → likely
  cause → what to check → fix, and MUST cover at minimum: the world will not start; the
  daemon looks running but is not; villagers only act on reflex; no model is configured;
  the model is too slow for the chosen speed; a format-version or migration mismatch;
  and the world is running but nothing appears in the feed.
- **FR-009** Every "what to check" step MUST name a real observable surface carried by
  the page's declared sources.
- **FR-010** `docs/player/index.html` MUST gain a Reference section linking the three new
  pages, and MUST continue to carry no `promptworld-docs:source` tags.
- **FR-011** Every other existing page under `docs/player/` MUST be byte-identical after
  this change.
- **FR-012** Each new page MUST carry `promptworld-docs:generated-by` and one
  `promptworld-docs:source` tag per declared source, each on its own line, in the
  contract grammar `<repo-relative-path>@<40-hex-lowercase-commit>`
  (`specs/026-player-docs/contracts/provenance-and-check.md`).
- **FR-013** Each new page MUST inline the canonical skeleton and shared `<style>` block
  verbatim, with no external assets and no JavaScript.
- **FR-014** The player-docs skill (`.claude/skills/player-docs/SKILL.md`) MUST be updated
  in the same PR: the expected page set gains the three pages, and the page→source
  mapping table gains a row per page naming its declared sources.
- **FR-015** `EXPECTED_PAGES` in `.claude/skills/player-docs/scripts/check-freshness.mjs`
  MUST include the three new pages, and `--check` MUST exit 0 on the branch.
- **FR-016** Prose MUST be plain-language for a non-engineer, consistent with the
  existing pages: no identifier, package path, or engineering term a player would have
  to decode, except where the term IS the thing the player types or the filename they
  open.
- **FR-017** (Housekeeping, in the table being edited anyway.) The mapping table's row
  for `playing-via-metatron.html` MUST be corrected to the sources that page actually
  declares — the `guardian*` note family — replacing the stale `docs/wiki/metatron.md` /
  `docs/wiki/metatron-miracles.md` entries, which name notes that no longer exist after
  the spec 052 rename.

### Non-functional

- **NFR-001** No Go source file changes. This feature ships documentation and the skill
  wiring that gates it.
- **NFR-002** Regenerating with everything fresh MUST remain a byte-identical no-op, per
  the skill's check-first procedure.

## Declared sources (the pins the pages project from)

- **command-reference.html**: `docs/wiki/cli-promptworld.md`,
  `docs/wiki/cli-world-lifecycle.md`, `docs/wiki/cli-runtime-control.md`,
  `docs/wiki/cli-guardian-ops.md`, `docs/wiki/instance-manager.md`,
  `docs/wiki/world-forking.md`, `docs/wiki/daemon-lifecycle.md`,
  `docs/wiki/curriculum-ladder.md`, `docs/wiki/cognition-estimator-calibration.md`
- **world-files-reference.html**: `docs/wiki/world-save-directory.md`,
  `docs/wiki/world-save-manifest-fields.md`, `docs/wiki/world-tuning.md`,
  `docs/wiki/world-tuning-dial-catalog.md`, `docs/wiki/skin.md`,
  `docs/wiki/guardian-instruction-surface.md`, `docs/wiki/governance.md`,
  `docs/wiki/bundle-tools.md`, `docs/wiki/event-log.md`, `docs/wiki/snapshots.md`,
  `docs/llm-providers.md`
- **troubleshooting.html**: `docs/wiki/daemon-lifecycle.md`,
  `docs/wiki/daemon-boot-recovery.md`, `docs/wiki/instance-manager.md`,
  `docs/wiki/llm-preflight-detection.md`, `docs/wiki/llm-provider-health.md`,
  `docs/wiki/cognition-horizon-telemetry.md`, `docs/wiki/world-migration.md`,
  `docs/wiki/tui-client.md`

The implementer MAY add a source not listed here if a page genuinely draws on it, adding
the mapping-table row and the meta tag in the same change; it MUST NOT keep a listed
source it did not draw on.

## Success criteria

- **SC-001** A player can name every command the binary accepts, and what each is for,
  from one page.
- **SC-002** A player can decide, for every file in their world folder, whether they may
  edit it — from one page.
- **SC-003** Each of the seven required troubleshooting symptoms resolves to a check the
  player can actually run and a fix they can actually apply.
- **SC-004** `node .claude/skills/player-docs/scripts/check-freshness.mjs --check` exits
  0 on the branch, reporting 16 fresh pages.
- **SC-005** `node scripts/check-merge-drift.mjs pr` passes from the worktree.

## Out of scope

- Any in-app manual (`promptworld manual`, a TUI manual overlay). The operator chose the
  docs-only form; an in-app surface is a separate task if it is ever wanted.
- A single-document `MANUAL.md`, PDF, or man page.
- Restructuring or rewriting the thirteen existing player pages.
- Rewriting `docs/llm-providers.md` or `docs/bundles.md` for players — the new pages link
  to them.
- Any change to CLI behavior, flags, or help text. If the reference finds the built-in
  usage text wrong or thin, that is a finding to card, not to fix here.

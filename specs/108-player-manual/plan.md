# Implementation Plan: spec 108 — the player's user manual

**Spec**: `specs/108-player-manual/spec.md` · **Task**: TASK-182 ·
**Branch**: `task-182-player-manual`

## Summary

Add three reference pages to `docs/player/` — `command-reference.html`,
`world-files-reference.html`, `troubleshooting.html` — link them from `index.html`
under a new Reference section, and extend the player-docs skill and its freshness
script so the new pages are gated exactly like the existing thirteen. Docs and skill
wiring only; no Go changes.

## Technical context

- **Language/stack**: HTML (self-contained, no JS, no external assets) + Markdown
  (skill doc) + one Node ESM script edit.
- **Existing contract**: `specs/026-player-docs/spec.md` and
  `specs/026-player-docs/contracts/provenance-and-check.md` govern page shape and
  provenance. This spec extends the page set; it changes neither contract.
- **Gate**: `.claude/skills/player-docs/scripts/check-freshness.mjs`. It resolves the
  *current* pin for each declared source two ways — a `docs/wiki/` path reads the
  note's `verified_against:` frontmatter, any other path takes
  `git log -1 --format=%H -- <path>` — and compares it against the page's
  `promptworld-docs:source` meta tags. Both `SOURCE_RE` and `GENERATED_BY_RE` are
  **line-anchored**: each meta tag must sit alone on its own line, or it is not seen.
- **Blast radius**: `docs/player/` (3 new + `index.html`),
  `.claude/skills/player-docs/SKILL.md`,
  `.claude/skills/player-docs/scripts/check-freshness.mjs`.

## Constitution check

- **Principle V (model tiers)**: single-package, docs-and-prose work with no
  concurrency, architecture, or doctrine surface → the routine implementation tier
  (`.claude/agents/spec-implementer.md`, `claude-sonnet-5`). No escalation trigger
  fires. Tier choice and justification recorded on TASK-182.
- **Spec rigor**: full Spec Kit run; spec linked to the board via `spec-bridge:link`
  before implementation.
- **Root read-only**: all work on `task-182-player-manual` in
  `.worktrees/task-182`; lands by PR merge (`--merge`, never squash).
- **Wiki-in-PR (spec 069)**: this branch touches no file listed in any wiki note's
  `sources:`, so no wiki re-pin is expected. It *does* change `docs/player/`, but as an
  addition whose pins are taken fresh — the freshness check must be green in-branch,
  which is what the pr gate probes.

## Approach

### Phase 1 — Ground the facts (read-only)

Before writing a line of page prose, collect from the declared sources at their current
pins:

1. The full subcommand list and dispatch/exit discipline, the name-or-path world
   argument rule, and the `metatron`/`miracle` hidden-alias facts — from
   `cli-promptworld.md` and the three `cli-*` family notes.
2. Per-command flags. `cli-world-lifecycle.md`, `cli-runtime-control.md`,
   `cli-guardian-ops.md`, `instance-manager.md`, `world-forking.md`,
   `curriculum-ladder.md`, `cognition-estimator-calibration.md` are the flag-bearing
   notes. **Where a note does not carry a flag, the page does not claim one** (FR-003).
   Cross-checking against `cmd/promptworld/*.go` to decide *whether a source carries a
   fact* is fine; quoting the Go file as a source is not — it is not a declared source
   and would not be pinned.
3. The world save directory layout and manifest fields, the tuning dial catalog, skin
   shape, charter surface, governance's `village_charter.md`, bundles, event log and
   snapshots.
4. The failure modes and their observable surfaces: preflight detection, provider
   health, horizon telemetry/suppression, daemon boot recovery and lifecycle, instance
   liveness, migration/format-version.

Record each source's current pin as you read it — that pin is what the page's meta tag
must carry.

### Phase 2 — Write the three pages

Each page: copy the canonical skeleton and `<style>` block **verbatim** from an existing
page (`keys-reference.html` is the closest tonal model — pure reference, no lore), set
the title to `<Page title> — promptworld player docs`, emit one `source` meta per
declared source on its own line, and link back to `index.html` the way existing pages do.

- **`command-reference.html`** — opens with the shared world-argument rule and a
  one-screen summary table (command · one line), then a section per intent group with
  the detail: what it does, what you type, flags, and what you see back. Groups (plain
  names, ordered by a player's journey): making a world; running it; watching it;
  playing it; the AI; comparing runs; and a short closing note for the retired aliases.
- **`world-files-reference.html`** — opens with "here is what's in your world folder",
  then one entry per file: what it is, the disposition badge (**yours to edit** /
  **read it, don't edit it** / **leave it alone**), what changing it does, and where the
  deeper reference lives when there is one.
- **`troubleshooting.html`** — a section per symptom, phrased as the player would say
  it, each with: what's probably happening · what to check · what to do. Never invent a
  diagnostic; every check names a surface the sources carry.

Cross-link the three to each other and to the teaching pages they complement
(`command-reference` ↔ `getting-started`; `world-files-reference` ↔
`llm-setup-basics`; `troubleshooting` ↔ `the-ai-behind-the-village` and
`time-and-speed`).

### Phase 3 — Nav

`index.html` gains a Reference section (a heading plus three list items in the existing
card markup) after the existing list. It keeps carrying zero `source` tags — the script
rejects `index.html` if it declares any.

### Phase 4 — Gate wiring

- `check-freshness.mjs`: add the three slugs to `EXPECTED_PAGES` and update the comment
  above it, which currently narrates "nine → thirteen".
- `SKILL.md`: add the three pages to the expected page set (the count prose changes
  from twelve topic pages to fifteen), add a mapping-table row per page, add a short
  editorial note describing the reference trio and its tone rule, and correct the stale
  `playing-via-metatron.html` row (FR-017).
- The skill's `description` frontmatter says "thirteen self-contained HTML pages" —
  update that count too, or the skill's own summary contradicts its page set.

### Phase 5 — Verify

1. `node .claude/skills/player-docs/scripts/check-freshness.mjs --check` → exit 0,
   "16 fresh, 0 stale, 0 missing, 0 broken-ref".
2. `git diff --stat origin/main` shows only the six intended files.
3. `git diff origin/main -- docs/player/` touches no existing page but `index.html`
   (FR-011).
4. Open each page and confirm: no `<script>`, no external `href`/`src`, dark mode
   renders (the `prefers-color-scheme` block is present verbatim).
5. `node scripts/check-merge-drift.mjs pr` from the worktree → exit 0.

## Risks

- **Inventing flags.** The likeliest failure: writing a plausible flag the sources do
  not carry. Mitigation: FR-003 is a hard rule; when in doubt, describe the command
  without the flag rather than guess.
- **Meta-tag formatting.** The script's regexes are line-anchored; a tag sharing a line
  with anything else reads as absent and the page reports broken-ref. Copy the tag
  layout from an existing page exactly.
- **Silent restale of existing pages.** Reformatting or re-saving a fresh page violates
  FR-011 and the skill's no-op rule. Touch only the six files named above.
- **Pin drift during the branch's life.** If main moves and a source re-pins while this
  branch is open, merge main *into* the branch (never rebase) and re-take the affected
  pins before the PR.

## Project structure

```
docs/player/
  index.html                     (edited — Reference section)
  command-reference.html         (new)
  world-files-reference.html     (new)
  troubleshooting.html           (new)
.claude/skills/player-docs/
  SKILL.md                       (edited — page set, mapping table, notes, description)
  scripts/check-freshness.mjs    (edited — EXPECTED_PAGES + comment)
specs/108-player-manual/
  spec.md  plan.md  tasks.md
```

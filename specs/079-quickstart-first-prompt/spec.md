# Feature Specification: Quickstart first-prompt pass — the minimal session includes one guardian prompt

**Feature Branch**: `079-quickstart-first-prompt` (task branch: `task-153-quickstart-first-prompt`)

**Created**: 2026-07-26

**Status**: Draft

**Input**: TASK-153 (reorient 2026-07-26 decision 7, merged position 5 —
"the quickstart finally has the player prompt: a first session that never
prompts teaches watching"). Board ACs: (1) getting-started.html includes a
first-prompt step sourced from `skin.guardian.example_ask.*`; (2) each stage
page carries a first-session do-this-then-this block. Content-only; the
player-docs skill is the home.

## Grounding (verified against the working tree, 2026-07-26)

**The gap.** `docs/player/getting-started.html` walks install → create →
start → watch → stop. Steps 4–5 never have the player *say anything* to the
guardian; the four stage pages (`stage-N-*.html`) describe what each stage
teaches and how to move up, but none says what to actually *do* in the first
ten minutes. Decision 7's diagnosis: a first session that never prompts
teaches watching.

**The token family.** `skin.guardian.example_ask.<tool-id>` — one canned
player phrasing per guardian loop tool, compiled defaults in
`internal/skin/skin.go:80-88`, documented (values verbatim, resolution order,
help-overlay consumption) in `docs/wiki/skin.md` (a `verified_against`-pinned
wiki note — the correct declared source for player docs). The in-game `?`
overlay's guardian section renders exactly this family, one ask per **granted**
verb, resolved through the active world's skin (spec 063 D9;
`docs/design/tui/overlays/help.md`). The stage-1 tool ceiling is pinned to
`send_omen`, `send_vision`, `monitor_and_act`, `cancel_order`
(`specs/046-curriculum-ladder/contracts/stage-gating.md`) — the sample ask
must name one of those four.

**The mechanics (research.md R1–R2, the decision that shapes everything).**
`docs/player/` pages are LLM-authored projections of *declared sources*. The
freshness gate (`check-freshness.mjs`) verifies provenance pins only — body
edits never trip it — but the skill's regeneration contract rewrites a stale
page from its declared sources with "no independently asserted facts", so
content survives the next regeneration only if (a) its facts live in a
declared source, tagged at its current pin, and (b) the page's required shape
is written into `.claude/skills/player-docs/SKILL.md`'s editorial contract.
This feature therefore lands in three coordinated places: the pages, their
source meta tags, and the SKILL.md contract — all on this one branch.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Getting started makes the player prompt once (Priority: P1)

A brand-new player following getting-started.html reaches, after "watch it
live" and before "when you're done", a step that has them type **one concrete
ask** to their guardian — a verbatim default-skin sample from the
`skin.guardian.example_ask.*` family, legal at stage 1 — and tells them what
to expect back and where to find more asks (the `?` overlay's guardian
section).

**Why this priority**: board AC #1 and decision 7's core: the minimal
documented session must include one prompt, or the docs teach watching.

**Independent Test**: read the amended page — the step exists, the quoted ask
is byte-identical to a `defaultTable` value for a stage-1-ceiling verb, the
page declares `docs/wiki/skin.md` as a source at its current pin, and
`node .claude/skills/player-docs/scripts/check-freshness.mjs --check` exits 0.

**Acceptance Scenarios**:

1. **Given** the amended getting-started.html, **When** a player reads the
   walkthrough in order, **Then** a numbered "ask your guardian one thing"
   step appears after the watch-it-live step, tells them where to type (the
   message box in `promptworld ui` / the `G` console), quotes one sample ask
   verbatim from the default table for a verb in {send_vision, send_omen,
   monitor_and_act, cancel_order}, and describes the expected response shape
   (the guardian replies/acts; the charge line is where the cost shows).
2. **Given** the step, **When** the player wants more phrasings, **Then** the
   step points at the in-game `?` overlay's guardian section as the live,
   per-world list — one example ask per verb *their world* grants.
3. **Given** a world with a custom skin, **When** the player compares the
   page's sample to their game, **Then** the page has already told them the
   printed phrasing is the default Guardian skin's and their world's own
   phrasing appears in the `?` overlay (never presenting the printed string
   as universal).
4. **Given** the amended page's `<head>`, **When** the freshness probe runs,
   **Then** `docs/wiki/skin.md@<current verified_against>` appears as a
   well-formed source tag and the probe exits 0.

---

### User Story 2 - Every stage page opens with a first session (Priority: P2)

A player landing on any of the four stage pages finds a short **"Your first
session"** do-this-then-this block: 3–5 ordered steps from "create a world at
this stage" to the stage-appropriate first acts, working on a plain world (no
`--scenario`), with the stage's exercise (where one exists) framed as the
when-you're-ready follow-on.

**Why this priority**: board AC #2; the stage pages currently explain the
ladder but never script the first ten minutes.

**Independent Test**: each of the four pages carries the block; every claim in
each block traces to that page's already-declared spec 046 sources; no stage
page's source tag set changed; the probe exits 0.

**Acceptance Scenarios**:

1. **Given** stage-1-the-voice.html, **When** the block is read, **Then** it
   walks: create (default stage, no flag) → start → `ui` → ask one thing
   (linking to getting-started's first-prompt step rather than re-quoting
   token values) → when ready, `--scenario first-night`; and it works
   verbatim on a world created without a scenario.
2. **Given** stage-2-the-written-word.html, **Then** its block includes
   writing one durable line into `charter.md` (in force from the guardian's
   next turn) and frames `the-law` as the when-ready exercise.
3. **Given** stage-3-the-craft.html, **Then** its block includes granting a
   tool / composing a skill file and frames the stage-3 scenario (a
   self-granted tool contributing to the pass) as the when-ready exercise.
4. **Given** stage-4-the-stewardship.html, **Then** its block scripts a
   stewardship first session and states plainly there is no exercise to pass.
5. **Given** all four blocks, **When** their sources are audited, **Then**
   every factual claim projects from the page's already-declared
   `specs/046-curriculum-ladder/` sources — no stage page gains or loses a
   source tag.

---

### User Story 3 - The next regeneration reproduces this content (Priority: P1)

A future player-docs regeneration (triggered by any source re-pin) rewrites
`getting-started.html` or a stage page — and the first-prompt step and
first-session blocks are **reproduced, not erased**, because the skill's
editorial contract now requires them.

**Why this priority**: without this, the feature is a time bomb — the
regeneration contract ("no independently asserted facts", research.md R1)
would legitimately drop the new content on the next re-pin of any source.

**Independent Test**: read `.claude/skills/player-docs/SKILL.md` — the
mapping table row for getting-started.html includes `docs/wiki/skin.md`, the
rows for the five touched pages match their actual declared sources, and an
editorial shape note (precedent: the TASK-114 and TASK-68 paragraphs) states
that getting-started carries the first-prompt step and the stage quartet
carries first-session blocks.

**Acceptance Scenarios**:

1. **Given** the amended SKILL.md, **When** its mapping table is compared to
   the five touched pages' `<head>` tags, **Then** each row lists exactly
   that page's declared sources (reconciliation bounded to touched pages —
   research.md R6).
2. **Given** the amended SKILL.md, **When** a regenerator follows step 2b for
   a stale getting-started.html or stage page, **Then** the editorial notes
   require it to include the first-prompt step (sample ask from the
   `example_ask` family, stage-1-legal verb, default-skin phrasing with the
   custom-skin honesty note) / the first-session block respectively.

---

### Edge Cases

- **Skins other than default**: the page prints the default Guardian skin's
  values (what `docs/wiki/skin.md` documents) and says so; the `?` overlay is
  named as the always-correct per-world surface. Never claim the printed
  phrasing is what every player will see.
- **Worlds created without a scenario**: both the getting-started step and
  every first-session block must be executable on a plain
  `promptworld new <name>` world — no exercise tab, no `--scenario` required;
  exercises are optional follow-ons (stage-4 has none at all).
- **Verbs outside the stage-1 ceiling**: the family also carries
  `work_miracle`, `pause`/`start`/`adjust_speed`, and `explain` asks — none
  in the pinned stage-1 set; the getting-started sample MUST NOT use them.
- **index.html**: never gains a source tag (probe hard-fails it) and is not
  touched by this feature.
- **Untouched pages**: the other 8 topic pages stay byte-identical — the
  probe's fresh verdict on them is necessary but not sufficient; byte
  identity is asserted via `git diff --stat`.
- **A later `docs/wiki/skin.md` re-pin**: getting-started then reports stale
  — correct and intended; the amended SKILL.md makes the ensuing
  regeneration reproduce the step (US3).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `docs/player/getting-started.html` MUST gain an "ask your
  guardian one thing" step, positioned after the watch-it-live step and
  before the stopping step (existing section numbering adjusted
  accordingly), that instructs the player to type one ask into the guardian
  message box (`promptworld ui` / the `G` console).
- **FR-002**: The step's sample ask MUST be a byte-verbatim value from the
  compiled default table's `skin.guardian.example_ask.*` family as documented
  in `docs/wiki/skin.md`, and its verb MUST be in the pinned stage-1 ceiling
  ({`send_vision`, `send_omen`, `monitor_and_act`, `cancel_order`});
  recommended: `send_vision` → `"show Ash a vision of the fire dying"`.
- **FR-003**: The step MUST name the in-game `?` overlay's guardian section
  as the live source of one example ask per verb the player's world grants,
  and MUST carry the skin honesty note: printed phrasings are the default
  Guardian skin's; a custom skin re-voices them and the `?` overlay always
  shows the world's own.
- **FR-004**: getting-started.html MUST declare `docs/wiki/skin.md` as a new
  `promptworld-docs:source` meta tag pinned at that note's current
  `verified_against` value at authoring time, grammar
  `<path>@<40-hex-lowercase>`, one tag per line.
- **FR-005**: Each of the four `stage-N-*.html` pages MUST gain a short
  "Your first session" do-this-then-this block (3–5 ordered steps) whose
  every factual claim projects from that page's already-declared
  `specs/046-curriculum-ladder/` sources; stage pages MUST NOT gain or lose
  source tags, and MUST link to getting-started's first-prompt step for the
  ask itself rather than quoting token values.
- **FR-006**: Every first-session block MUST be executable on a world created
  without `--scenario`; stage exercises (first-night, the-law, the stage-3
  scenario) appear only as when-you're-ready follow-ons, and stage-4's block
  states there is no exercise.
- **FR-007**: `.claude/skills/player-docs/SKILL.md` MUST be amended in the
  same branch: (a) the mapping-table rows for the five touched pages
  reconciled to those pages' actual declared sources (bounded reconciliation,
  research.md R6); (b) editorial shape notes added — getting-started carries
  the first-prompt step (its constraints per FR-002/FR-003), the stage
  quartet carries first-session blocks (constraints per FR-005/FR-006) — so
  future regenerations reproduce the content.
- **FR-008**: No file outside `docs/player/` and
  `.claude/skills/player-docs/SKILL.md` changes: no Go code, no
  `docs/wiki/` edits (hence no wiki re-pins), no `docs/design/` edits, no
  `index.html` change; the 8 untouched topic pages remain byte-identical.

### Key Entities

- **First-prompt step** — the new getting-started section: where to type, one
  verbatim stage-1-legal sample ask, expected response, `?` overlay pointer,
  skin honesty note.
- **First-session block** — per-stage 3–5 step do-this-then-this, projected
  from spec 046 sources, scenario-independent.
- **Source meta tag** — `promptworld-docs:source` = `<path>@<40-hex>`; wiki
  sources pin `verified_against`, plain files pin `git log -1`.
- **Editorial contract** — SKILL.md's mapping table + shape notes; the only
  artifact that makes page structure survive regeneration.

## Success Criteria *(mandatory)*

- **SC-001**: Board AC #1 — getting-started.html includes the first-prompt
  step sourced (verbatim value + declared meta tag) from
  `skin.guardian.example_ask.*`; US1's independent test passes.
- **SC-002**: Board AC #2 — all four stage pages carry the first-session
  block; US2's independent test passes.
- **SC-003**: `node .claude/skills/player-docs/scripts/check-freshness.mjs
  --check` exits 0 on the branch after all edits (13 fresh, 0 stale, 0
  missing, 0 broken-ref), and `git diff --stat origin/main` shows only the
  five pages + SKILL.md changed.
- **SC-004**: `go test ./...` green (no Go changes; doctrine run).
- **SC-005**: The merge-drift pr gate passes from the worktree
  (`node scripts/check-merge-drift.mjs pr`) — `wiki-repin-missing` and
  `player-docs-stale` expected clean per research.md R7; the PR merges
  merge-commit-only (`gh pr merge --merge`).

## Assumptions

- No wiki note lists `docs/player/*` or `.claude/skills/player-docs/*` as a
  source (verified empty grep, research.md R7) — so no wiki re-pin belongs in
  this branch. The gate remains the authority; if it disagrees, produce the
  re-pin, don't argue.
- `docs/wiki/skin.md`'s `verified_against` is
  `31c893e0406653197e467a89b2fdb96f0bcf2ee0` at spec time; the implementer
  records whatever is current when authoring the tag (the probe enforces it).
- `check-tui-design.mjs` is not applicable — no `docs/design/tui/` page
  changes and no `internal/tui/` changes.
- The default-table values in `internal/skin/skin.go` and their transcription
  in `docs/wiki/skin.md` agree (verified byte-identical for the quoted
  family); if they ever diverge, the wiki note is the page's declared source
  and the divergence is a wiki bug to fix upstream, not here.

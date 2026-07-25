# Feature Specification: TUI Design Reference v2 — the Living UI Authority

**Feature Branch**: `047-tui-design-reference-v2`

**Created**: 2026-07-25

**Status**: Draft

**Input**: User description: "TUI design reference v2 — the living page-by-page, control-by-control UI authority. Reconcile docs/design/tui with shipped reality (specs 013-046); adopt the v2 taxonomy (pages/ · panels/ per dock tab · overlays/ · patterns/ + anatomy.md); uniform control tables; verified_against pins + check script + same-PR freshness gate; author ten new-surface pages spec-before-build; rule on bottom-chrome row budget/fold order, narrow fallback, and the help overlay's no-LLM byte-identity invariant. Grounding: docs/design/reorient-2026-07-25-ui.md (Wave 0). Board: TASK-123."

## Constitutional Context

This is Wave 0 of the 2026-07-25 UI reorientation (`docs/design/reorient-2026-07-25-ui.md`,
decision 4): `docs/design/tui/` becomes the living page-by-page, control-by-control UI
authority with a mechanized freshness gate. Every later wave (skin tokens, guardian console,
teaching chrome, village lens) specifies into this document before code. Operator decisions
1–8 and lead defaults D1–D13 from the reorientation are FIXED constraints — this spec cites
them, never re-argues them.

## Clarifications

### Session 2026-07-25

- Q: Ambient postmortem contents (reorient open question 1)? → A: Morgue evidence only;
  report card appears in scored/scenario runs only.
- Q: Score voice inside the unlock ceremony (reorient open question 2)? → A: Both —
  narrated skin-resolved chapter plus rubric checklist, with the instrument (rubric)
  authoritative.
- Q: Advanced-tier audience — raw registry values player-visible (reorient open
  question 4 / analysis Q3)? → A: Plain-language by default everywhere; raw registry
  values behind an explicit debug/inspector toggle (a mode, not a tier).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Implementer builds a shipped surface from the reference (Priority: P1)

An implementer (human or AI) assigned a change to any currently shipped TUI surface — the
metatron dock tab, the stage header segment, verdict/suppression rows, the help overlay —
opens `docs/design/tui/`, finds the owning page via the region index, and can make the
change without reading Go source to re-derive intent. Today this fails: `panels/dock.md`
predates roughly six specs of shipped dock content (provider table, horizon rows, standing
orders, stage segment, verdict rows), and the help overlay exists only as a section
appended to `patterns/keymap.md`.

**Why this priority**: the reference's authority claim ("an implementer should be able to
build every screen from these files") is currently false — reorientation lens clause (d)
fails. Reconciliation is the precondition for everything else in this feature: pins and
gates on stale content would mechanize a lie.

**Independent Test**: pick any visible region, strip, or badge in the shipped client;
starting from the top-level region index, locate its owning page in ≤2 hops; verify the
page describes what actually renders (states, data source, keys) with no shipped element
missing.

**Acceptance Scenarios**:

1. **Given** the shipped metatron dock tab with provider table, horizon rows, and standing
   orders, **When** an implementer looks up the dock in the v2 reference, **Then** they find
   per-tab panel pages whose content matches the shipped rendering, with fiction-layer
   content (guardian) and engine telemetry (systems) documented as separate pages.
2. **Given** any surface introduced by specs 013–046, **When** it is searched for in the
   reference, **Then** exactly one page owns it and `anatomy.md` maps its screen region to
   that page.
3. **Given** the shipped help overlay (spec 045), **When** an implementer needs its
   behavior, **Then** a dedicated overlay page specifies it (no longer a keymap appendix).

---

### User Story 2 - New surfaces are authored spec-before-build (Priority: P1)

A designer/implementer picking up any reorientation wave (guardian console, teaching
chrome, village lens) finds the target surface already fully specified in the reference —
mockup-first page, control table, linear-stream projection — before its implementing task
starts. Ten new-surface pages are authored in this feature: guardian console page, systems
tab panel, exercise panel, ceremony overlay, postmortem overlay, lesson row, guardian
strip, villager strip, stage-defaults pattern, and the help overlay's guardian section.

**Why this priority**: this is the reorientation's stated deliverable — "new surfaces are
authored here spec-before-build." Waves 1–5 are blocked on these pages existing; equal
top priority with reconciliation.

**Independent Test**: for each of the ten surfaces, open its page and verify it carries an
ASCII mockup with per-callout prose, a complete control table, its stage-default
visibility, and a specified linear-stream/CLI projection — sufficient to hand to an
implementing task with no further design work.

**Acceptance Scenarios**:

1. **Given** the ten enumerated surfaces, **When** the v2 reference lands, **Then** each has
   an authored page (or, for the guardian section, an authored section of the help overlay
   page) with mockup, control table, and stage-default behavior.
2. **Given** the ceremony and postmortem overlay pages, **When** their invariants are read,
   **Then** replayability from pull surfaces (`?`, `stages`, morgue) is stated as an
   explicit acceptance criterion on each, not as intent.
3. **Given** the lesson row page, **When** its behavior section is read, **Then** it
   specifies: one active lesson, ≤2 lines, dwell-until-done/dismissed, UI-pointer field,
   anti-spam (spacing, opportunity decay), per-user seen state, and stage-default
   visibility (on at stages 1–2; badge+overlay-only at stage 3+ and pre-ladder).
4. **Given** the stage-defaults pattern page, **When** its rules are read, **Then** stage
   shapes *defaults only*: every surface remains reachable at every stage and via the help
   overlay; pre-ladder worlds get everything; capability locks stay angel-only.

---

### User Story 3 - AI implementer and sweep scripts query uniform control tables (Priority: P2)

An AI implementer (or a future parity-sweep script) queries any panel or overlay page and
finds one uniformly-structured control table — identical column header everywhere — listing
every control/region with its states, data source, renderer symbol, keyboard *and* mouse
bindings, introducing spec, and skin token. Fiction strings appear only as skin tokens.

**Why this priority**: the control table is the AI-parseable unit that makes the reference
machine-queryable (the reorientation's clause-(d) mechanism), and the mouse column is what
makes an input-parity sweep (decision 8) possible at all. Depends on US1/US2 pages existing.

**Independent Test**: parse every panel/overlay page mechanically; assert every page has
exactly one control table whose header matches the canonical column set; assert no
fiction-fictional string literal appears outside a skin-token reference.

**Acceptance Scenarios**:

1. **Given** all panel and overlay pages, **When** their control tables are parsed,
   **Then** every table carries the same fixed columns: control/region · states · data
   source · renderer · keys+mouse · introduced-by · skin-token.
2. **Given** the keymap pattern page, **When** the input-parity doctrine is looked up,
   **Then** decision 8 is recorded there as a binding rule (every action reachable by
   keyboard and mouse; keyboard primary and complete; incremental rollout).
3. **Given** any fiction-layer string in the reference (guardian names, ceremony copy),
   **When** it appears in a mockup or table, **Then** it is rendered as a skin token, with
   the token conventions documented in a skin-tokens pattern page.

---

### User Story 4 - The freshness gate keeps the reference true (Priority: P2)

A contributor merges a change to the TUI without amending the design reference; the gate
fails their PR. Each reference file carries a `verified_against` pin; a check script
validates structure and pins; the same-PR amendment rule is enforced mechanically rather
than by convention (which already failed for specs 044/046).

**Why this priority**: without mechanized freshness the v2 reference rots exactly as v1
did; but the gate is only meaningful once the content it guards (US1/US2) exists.

**Independent Test**: run the check script on a clean tree (passes); simulate a
TUI-touching change with no design-doc amendment (fails with an actionable message);
verify every reference file carries a valid pin.

**Acceptance Scenarios**:

1. **Given** the v2 reference, **When** any file is inspected, **Then** it carries a
   `verified_against` pin naming the commit it was last verified against.
2. **Given** a changeset touching TUI implementation files with no same-PR change under
   the design reference, **When** the check runs, **Then** it fails and names the
   unamended files.
3. **Given** a page whose control table header deviates from the canonical columns or
   whose pin is missing/malformed, **When** the check runs, **Then** it fails.

---

### User Story 5 - The three Wave 0 rulings are recorded (Priority: P2)

A designer of any later wave consults the reference for the three structural questions no
prior decision covered and finds them ruled on: (a) the bottom-chrome row budget and fold
order — at stages 1–2 the bottom chrome stacks lesson row + guardian strip + minibuffer +
footer (~4 fixed chrome elements) in an exact-height layout, so the layout pattern must
state the re-derived row math and which row collapses first as height shrinks; (b) the
narrow-fallback behavior of the new chrome (what the narrow single-pane layout carries);
(c) the help overlay's no-LLM invariant restated — precisely which overlay sections remain
byte-identical with nil status (the deterministic floor guarantee) and which are
status-derived, now that the guardian section, ceremony replay, and the lessons registry
pull content in the dynamic direction.

**Why this priority**: the reorientation names the row budget as "the one genuinely
structural item" and Wave 0's first ruling; leaving these unruled would push structural
decisions into implementing waves where they'd be made by accident.

**Independent Test**: open the layout pattern page and the help overlay page; verify the
row math sums to the terminal height at each stage default, the fold order is a total
order over the collapsible rows, the narrow fallback is specified for each new chrome
element, and every help-overlay section is classified byte-identical vs status-derived.

**Acceptance Scenarios**:

1. **Given** the layout pattern page v2, **When** the row budget is read for a stage-1–2
   world on a 30-row terminal, **Then** the arithmetic accounts for every fixed row
   (header, body, lesson row, guardian strip, villager strip if defaulted on, minibuffer,
   footer) and states the minimum-height collapse order.
2. **Given** the narrow (single-pane) fallback, **When** its page is read, **Then** each
   new chrome element's narrow behavior is specified (carried, folded to a badge, or
   overlay-only).
3. **Given** the help overlay page, **When** its content contract is read, **Then** every
   section is explicitly classified: byte-identical with nil status, or status-derived —
   and the deterministic no-LLM floor guarantee is restated over the byte-identical set.

---

### Edge Cases

- What happens when a shipped surface has no natural home in the four-class taxonomy
  (e.g., the footer key-hint line, header badges)? → `anatomy.md` must still map it; the
  owning page may be a pattern page, but no visible element may be unmapped.
- How does the check script treat mockup drift (mockups are representative, not
  screenshots)? → pins verify against commits and structure, not pixel content; the
  same-PR rule catches semantic drift via review, the script catches structural drift.
- What happens when a PR touches TUI files for a pure refactor with zero visible change?
  → the gate still requires a same-PR touch of the reference; the minimal compliant
  amendment is re-verifying and bumping the affected pages' pins (extending INDEX rule 4
  from "deviation" to "any change").
- What happens on terminals shorter than the stage-1–2 fixed chrome stack (e.g., 15
  rows)? → the fold order must terminate in a layout that still renders header, body,
  minibuffer, and footer; the ruling must state the floor.
- Two overlays trigger at once (ceremony fires during an open postmortem)? → takeover
  surfaces must be specified as non-stacking with a stated precedence.
- A new-surface page specifies a surface whose implementing task is later re-scoped? →
  the reference is the authority: the implementing task amends the page in its own PR
  (same-PR rule), keeping spec-before-build true through change.

## Requirements *(mandatory)*

### Functional Requirements

**Reconciliation & taxonomy**

- **FR-001**: The reference MUST adopt the four-class taxonomy — `pages/`, `panels/` (one
  file per dock tab plus each standalone strip/row), `overlays/`, `patterns/` — plus a
  top-level `anatomy.md` region index and a rewritten `INDEX.md` carrying the authority
  statement and the gate rules.
- **FR-002**: `anatomy.md` MUST map every visible region, strip, badge, and chrome row of
  the shipped client to its owning reference file, with no unmapped visible element.
- **FR-003**: The monolithic dock page MUST be split into per-tab panel pages, separating
  fiction-layer guardian content from never-skinned engine telemetry (systems) so the
  skin boundary is a file boundary (reorientation D10).
- **FR-004**: Every surface shipped by specs 013–046 MUST be documented where an
  implementer would look, closing the known staleness (provider table, horizon rows,
  standing orders, stage header segment, verdict rows, suppression/remedy rows, morgue
  surfaces, help overlay) — the reference's authority claim MUST be true at landing.
- **FR-005**: The help overlay MUST become a dedicated overlay page (extracted from the
  keymap pattern page), covering its shipped sections plus the new guardian section.

**Control tables & doctrine**

- **FR-006**: Every panel and overlay page MUST carry exactly one control table with the
  canonical fixed columns: control/region · states · data source · renderer · keys+mouse
  · introduced-by · skin-token.
- **FR-007**: Fiction strings in the reference MUST appear only as skin tokens; a
  skin-tokens pattern page MUST document the token conventions and column semantics used
  by the reference (the runtime token lookup contract itself ships with the skinnable
  guardian feature — TASK-121 — and is out of scope here, per reorientation D2 sequencing).
- **FR-008**: The keymap pattern page MUST record the input-parity doctrine (reorientation
  decision 8): every action reachable by both keyboard and mouse, keyboard primary and
  complete, incremental rollout; the printable one-page reference-card format MUST be
  preserved.

**Freshness mechanization**

- **FR-009**: Every reference file MUST carry a `verified_against` pin identifying the
  commit it was last verified against.
- **FR-010**: A check script MUST validate the reference mechanically: pins present and
  well-formed, canonical control-table header on every panel/overlay page, anatomy index
  complete against the file set, and same-PR amendment (a changeset touching TUI
  implementation files must also touch the reference).
- **FR-011**: The same-PR amendment rule MUST be enforced as a gate on the project's
  standard verification path (the TASK-82 freshness-check precedent), extending the old
  INDEX rule 4 from "record deviations" to "any TUI change amends the reference in the
  same PR"; a violation blocks with an actionable message naming what to amend.

**New-surface pages (spec-before-build)**

- **FR-012**: The reference MUST gain authored pages for the ten reorientation surfaces:
  guardian console page; systems tab panel; exercise panel; ceremony overlay; postmortem
  overlay; lesson row; guardian strip; villager strip; stage-defaults pattern; and the
  help overlay's guardian section. Each MUST include a mockup with per-callout prose, a
  control table (FR-006), stage-default visibility, and a specified linear-stream/CLI
  projection (reorientation D1 — the accessibility floor).
- **FR-013**: The ceremony and postmortem overlay pages MUST state replayability from
  pull surfaces (help overlay, stages command, morgue) as an explicit acceptance
  criterion, MUST specify both takeovers as dismissable, non-stacking, with stated
  precedence, and MUST record the interrupt-policy watch item (reopening signal: playtest
  evidence of ceremony fatigue or mid-crisis seizure complaints).
- **FR-014**: The new-surface pages MUST encode their governing decisions: lesson row per
  decision 5 (one active, ≤2 lines, dwell, UI-pointer, anti-spam, per-user seen state,
  stage-defaulted delivery, prompting-lesson tier included in the trigger taxonomy);
  guardian strip per decision 7 (budget: charge bank, regen, standing-order count, faith
  when available — paired above the minibuffer); villager strip per D12 (one-row,
  stage-defaulted, under the header); exercise panel per D11 and D4 (framing, live
  event-derived rubric gauges, per-exercise visibility vocabulary — not a boolean —
  attach-time briefing, scenario-cadence narration trigger); stage-defaults pattern per
  decision 3 (defaults only; everything reachable at every stage and via the help
  overlay; pre-ladder gets everything; capability locks stay angel-only); guardian
  console per decisions 1/2 and D5 (document-style turns, composer, charter/skills read
  surface with binding status and `$EDITOR` handoff with the "charter changed — next
  turn binds it" confirmation, report-card cards at natural stopping points); help
  guardian section per D9 (static-per-stage, model-free: stage identity/concept, granted
  verbs, one example ask per verb); unlock attribution voice per D6 (the player's
  authorship, skin-resolved).

**Wave 0 rulings**

- **FR-015**: The layout pattern page MUST re-derive the row budget to account for the
  new permanent chrome (lesson row, guardian strip, villager strip) per stage default,
  and MUST rule a total minimum-height fold order (which row collapses first) with a
  stated floor layout.
- **FR-016**: The narrow-fallback behavior of each new chrome element MUST be ruled:
  what the narrow single-pane layout carries, folds to a badge, or defers to the overlay.
- **FR-017**: The help overlay page MUST restate the no-LLM invariant precisely: which
  sections remain byte-identical with nil status (the deterministic floor guarantee,
  spec 045's contract deliberately amended for D9) and which sections are status-derived.

**Operator-reserved decisions (reorientation open questions resurfacing in this feature)**

- **FR-018**: The postmortem overlay page MUST rule the ambient (unscored) world's
  run-end takeover as **morgue evidence only**: cause-of-death evidence in the morgue's
  no-blame register, with the report card appearing only in scored/scenario runs — the
  hybrid-scoring boundary stays crisp on screen. (Operator ruling 2026-07-25, closing
  reorientation open question 1.)
- **FR-019**: The ceremony overlay page MUST rule the score voice as **both, instrument
  authoritative**: a narrated, skin-resolved chapter for salience plus the rubric
  checklist as the authoritative record — matching D6's player-authorship voice while
  keeping the graded artifact inspectable. (Operator ruling 2026-07-25, closing
  reorientation open question 2.)
- **FR-020**: The control tables' player-facing data-source projections and the help
  overlay's advanced tier MUST stay **plain-language by default, with raw registry
  values behind an explicit debug/inspector toggle** — the boundary is a mode, not a
  tier. (Operator ruling 2026-07-25, closing reorientation open question 4 / analysis
  Q3.)

### Key Entities

- **Reference page**: one Markdown file owning one page/panel/overlay/pattern; carries a
  `verified_against` pin, a mockup (for visual surfaces), prose behavior, and a control
  table (panels/overlays).
- **Control table row**: the AI-parseable unit — one control/region with its states, data
  source (status field / event type / replica), renderer symbol, keys+mouse, introducing
  spec, and skin token.
- **`verified_against` pin**: per-file frontmatter/marker naming the commit the page was
  last verified against; consumed by the check script.
- **Check script + gate**: the mechanical freshness enforcement — structural validation
  plus the same-PR amendment rule on TUI-touching changesets.
- **Region index (`anatomy.md`)**: the map from every visible screen element to its
  owning page; the human entry point and the completeness oracle.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of visible regions, strips, badges, and chrome rows in the shipped
  client are mapped by the region index to an owning page, and every surface introduced
  by specs 013–046 is documented on the page where an implementer would look (audit
  finds zero unmapped or undocumented shipped elements).
- **SC-002**: 100% of panel and overlay pages carry exactly one control table with the
  canonical column header; zero fiction string literals appear outside skin-token form.
- **SC-003**: The check script passes on the landed tree, and fails with an actionable
  message on each seeded violation class: missing/malformed pin, non-canonical control
  table header, unmapped anatomy entry, and a TUI-touching changeset with no same-PR
  reference amendment.
- **SC-004**: All ten new-surface pages exist with mockup, control table, stage-default
  visibility, and linear-stream projection — before any of their implementing tasks
  starts (spec-before-build holds for waves 1–5).
- **SC-005**: An implementer starting from the reference's entry point reaches the
  authoritative page for any named control in at most 2 hops.
- **SC-006**: The three Wave 0 rulings are recorded and internally consistent: the row
  arithmetic sums to terminal height at every stage default and terminates in a stated
  floor; every new chrome element has a ruled narrow behavior; every help-overlay
  section is classified byte-identical vs status-derived.

## Assumptions

- The five-region anatomy (header / map / dock / minibuffer / footer), the focus
  contract, and the mockup-first page convention survive into v2 unchanged — the
  reorientation ruled extension, not replacement.
- Reorientation decisions 1–8 and D1–D13 are fixed constraints; this feature encodes
  them into pages and never re-opens them. The three [NEEDS CLARIFICATION] markers are
  exactly the reorientation's open questions whose stated resurfacing moment is this
  feature; no other clarifications are needed.
- The runtime skin-token lookup contract is TASK-121's deliverable (Wave 1); this
  feature only fixes the documentation conventions (tokens in mockups/tables) so the
  reference doesn't rot when the contract lands.
- The existing narrow-fallback doctrine ("the narrow single-pane UI is never deleted")
  and the 112-column breakpoint remain; rulings extend them rather than replace them.
- The check script follows the project's existing freshness-check precedent (TASK-82's
  player-docs checker) in spirit — a repo-local script wired into the standard
  verification path — without prescribing its implementation here.
- Documentation-only surfaces (pages describing not-yet-built features) pin against the
  commit that authored them; their control tables' renderer column may name the intended
  owner or be marked "unbuilt — see introducing wave".
- The board task (TASK-123) is the single deliverable owner; this feature lands as one
  PR per the constitution.

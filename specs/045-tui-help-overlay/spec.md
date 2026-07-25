# Feature Specification: `?` help overlay in the TUI (every world)

**Feature Branch**: `045-tui-help-overlay`

**Created**: 2026-07-25

**Status**: Draft

**Input**: User description: "? help overlay in the TUI, every world (TASK-116): a
context-sensitive help overlay openable with ? from every pane/mode — current mode's keys
first with a basic/advanced tier split, then a screen-region walkthrough covering header
anatomy, map glyph legend, and dock tabs; works identically in no-LLM reflex-only worlds;
every future pushed lesson also reachable from the overlay's pull reference."

## Why this exists (given constraints)

Learning-game synthesis, operator decision 8 (2026-07-25): onboarding is every-world,
TUI-level. A no-LLM world has no tutor — reflex-only villages are first-class — so the
overlay is the charter-independent floor beneath an angel that may be absent, down, or
mid-repair; it is not redundant with the tutor. 45 years of roguelike `?` convention
(NetHack screen-region walkthrough; Cogmind basic/advanced key tiering) supply the shape.
Grounding: Analysis-In-Game-First-Teaching rec 3 (R1); board task TASK-116.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Stuck anywhere, `?` answers "what can I press right now" (Priority: P1)

A player — in any pane, any mode, clock running or paused — presses `?` and gets an
overlay whose first page lists the keys that work *right now, in this mode*, most
important first. A second press of the tier key (or paging) reveals the advanced tier:
everything else that works in this mode, including layered global keys. Dismissing the
overlay returns them exactly where they were, nothing else having happened.

**Why this priority**: the core promise — context-sensitive pull help under one universal
key. Without it nothing else in this feature matters.

**Independent Test**: from each documented mode (global/home, minibuffer focused, inspect,
villagers roster, villagers detail, solo views), press `?`; verify the overlay opens, the
basic page lists that mode's keys first, the advanced tier is reachable, and dismissal
restores the prior screen and focus unchanged.

**Acceptance Scenarios**:

1. **Given** the client in any mode, **When** the player presses `?`, **Then** a help
   overlay opens showing the current mode's basic keys first, without disturbing the
   world (no pause, no command sent, no state change).
2. **Given** the overlay is open on the basic page, **When** the player advances the
   tier, **Then** the advanced page for the same mode appears (layered/global keys,
   power keys), clearly labeled as the advanced tier.
3. **Given** the overlay is open, **When** the player dismisses it (esc, or `?` again),
   **Then** the client returns to exactly the prior mode, focus, selection, and scroll —
   one dismissal releases only the overlay layer.
4. **Given** the minibuffer is focused (a text-entry context), **When** the player types
   `?`, **Then** the character is appended to the buffer as normal — text entry is never
   hijacked — and help for minibuffer mode remains reachable from any non-entry mode's
   overlay.
5. **Given** the clock is running, **When** the overlay is open, **Then** the world keeps
   ticking beneath it (opening help is not a pause) and the overlay does not go stale in
   a way that breaks dismissal.

---

### User Story 2 - The screen explained: regions, glyphs, tabs (Priority: P2)

From the overlay, the player reaches a walkthrough of the screen itself: what every part
of the header means (clock, speed suffix, badges like llm/suppressed/degraded states),
what each glyph on the map is (structures, villagers, the gru, piles — the full legend),
and what lives in each dock tab. A newcomer can decode everything visible on screen
without leaving the client or opening external docs.

**Why this priority**: keys tell you what you can *do*; the walkthrough tells you what
you're *seeing*. This is the NetHack-style anatomy lesson the synthesis names — second
because US1's overlay is its host.

**Independent Test**: open the overlay's walkthrough section; verify a page (or pages)
covering every header element currently renderable, every map glyph the map can draw,
and every dock tab by name, each with a one-line plain-language explanation.

**Acceptance Scenarios**:

1. **Given** the overlay is open, **When** the player navigates to the screen
   walkthrough, **Then** header anatomy, map glyph legend, and dock tabs are each
   covered, in plain language, one concept per line.
2. **Given** the header can render conditional badges (llm state, suppression, degraded,
   paused — and world postures like an ended run), **Then** the walkthrough explains
   every badge that can appear, not only the ones currently visible.
3. **Given** the map legend gains a new glyph in a future feature, **Then** the
   walkthrough's glyph list is derived from (or mechanically checked against) the same
   source the map renderer uses, so the two cannot silently diverge.
4. **Given** longer-than-a-screen walkthrough content, **When** the player scrolls or
   pages within the overlay, **Then** all content is reachable on small terminal sizes.

---

### User Story 3 - The floor holds with no angel (Priority: P3)

A player running a reflex-only world (no model configured), or whose provider is down,
opens `?` and gets exactly the same overlay, same content, same behavior. Help never
depends on the angel, the network, or any model.

**Why this priority**: the load-bearing rationale for the feature — but it is a
constraint on US1/US2's construction more than new surface area, so it rides them.

**Independent Test**: on a world with no AI configured, exercise US1 and US2's tests
verbatim; byte-identical help content and identical behavior.

**Acceptance Scenarios**:

1. **Given** a world with no model configured, **When** the player uses the overlay,
   **Then** content and behavior are identical to an LLM world's overlay.
2. **Given** any live-world state (angel mid-turn, provider erroring, budget exhausted),
   **Then** the overlay never blocks, errors, or changes because of it.

---

### User Story 4 - Pushed lessons are findable again (pull reference) (Priority: P4)

When the first-occurrence lesson projection lands (separate future task), every lesson it
can push is also readable on demand from the overlay — a pull reference section — so a
dismissed or forgotten lesson is never lost. Until that task lands, the overlay reserves
the seam: its structure has a place for the reference and the overlay's content source
can list lesson entries.

**Why this priority**: an AC on TASK-116 (#4), but its full delivery depends on a future
feature; this spec's obligation is the reachable seam plus whatever lesson-shaped content
already exists (today: none — the seam and its contract).

**Independent Test**: verify the overlay defines a navigable reference section whose
entries come from the same content source the future lesson projection will use, and that
adding an entry there requires no structural change to the overlay.

**Acceptance Scenarios**:

1. **Given** the overlay's structure, **Then** a reference section exists and is
   navigable, designed to list lesson topics for on-demand reading.
2. **Given** a future pushed lesson is defined, **Then** adding it to the pull reference
   is a content addition, not an overlay redesign (documented contract).

---

### Edge Cases

- Very small terminals: the overlay must render usably (scroll/page) at the client's
  minimum supported size; content never renders outside the overlay bounds.
- `?` pressed while the overlay is already open: treated as dismiss (toggle) — never a
  stacked second overlay.
- Mode changes beneath the overlay are impossible by construction: while the overlay is
  open, non-overlay keys are inert (except dismissal), so the "context" the help shows
  cannot drift while it is open.
- esc-release ordering: the overlay is one layer; esc closes it and only it, preserving
  the documented one-layer-per-press release chain beneath (minibuffer → detail → solo →
  home).
- The clock pause state must be irrelevant: overlay works paused and running; it never
  pauses or resumes the world.
- An ended (postmortem) world: the overlay still opens and its content still applies —
  reading surfaces are exactly what a postmortem player needs.
- Narrow-fallback solo views (screen too narrow for the composite): `?` still opens help
  for the active solo view's mode.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The client MUST open a help overlay on `?` from every mode except
  text-entry contexts (where printable characters belong to the buffer); in text-entry,
  `?` MUST type the character.
- **FR-002**: The overlay's first page MUST show the current mode's keys, basic tier
  first (the keys a newcomer needs), with an advanced tier (layered global keys, power
  keys) reachable by an in-overlay action; the tier split MUST exist for every mode.
- **FR-003**: Key listings MUST agree with the actual keymap: every binding shown must
  work, and every working binding must appear in exactly one tier of its mode's pages.
- **FR-004**: The overlay MUST include a screen-region walkthrough covering: header
  anatomy (every element and every conditional badge/posture the header can render), the
  complete map glyph legend, and every dock tab by name and purpose.
- **FR-005**: The glyph walkthrough MUST be sourced from or mechanically checked against
  the renderer's own glyph set so the legend cannot silently drift from the map.
- **FR-006**: Opening, using, and dismissing the overlay MUST be side-effect-free on the
  world: no commands sent, no pause/resume, no state mutation; the prior screen, focus,
  selection, and scroll positions MUST be restored exactly on dismissal.
- **FR-007**: The overlay MUST be dismissable by esc and by `?` (toggle), releasing
  exactly one layer per the client's esc-release contract.
- **FR-008**: All overlay content MUST be static, local, and model-independent —
  identical presence and behavior on no-LLM worlds, with any live-model condition unable
  to block or alter it.
- **FR-009**: Overlay content longer than the visible area MUST be scrollable/pageable,
  usable at the client's minimum supported terminal size.
- **FR-010**: The overlay MUST expose a pull-reference section structured to list
  lesson topics for on-demand reading; adding future lesson entries MUST be a content
  addition requiring no structural change (contract documented in the feature's design
  artifacts).
- **FR-011**: The footer hint machinery MUST advertise `?` so the overlay is
  discoverable (at minimum in the global mode's hints).
- **FR-012**: While the overlay is open, keys other than the overlay's own navigation
  and dismissal MUST be inert (no silent fallthrough to the mode beneath).

### Key Entities

- **Help overlay**: a dismissable presentation layer above the current screen; owns its
  navigation (tier advance, section navigation, scroll, dismiss) while open.
- **Mode help page**: per-mode content unit — basic-tier keys, advanced-tier keys —
  derived from the client's real keymap.
- **Screen walkthrough**: content unit explaining header anatomy, map glyphs, dock tabs;
  glyph list bound to the renderer's source of truth.
- **Pull reference**: the overlay section reserved to list lesson topics; its entries
  are content data, not structure.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: From every documented client mode, `?` produces mode-correct help within
  one keypress, and dismissal restores the exact prior context — verified across all
  modes including narrow-terminal fallbacks.
- **SC-002**: A reader can decode 100% of what the client can display — every header
  element/badge, every map glyph, every dock tab — from the overlay alone, without
  external documentation.
- **SC-003**: Key listings and the live keymap cannot disagree: a mechanical check ties
  every advertised binding to a real one (and flags unlisted bindings), keeping FR-003
  true over time.
- **SC-004**: On a no-model world, overlay content and behavior are identical to a
  model-configured world (same bytes, same interactions).
- **SC-005**: Opening and closing the overlay 100 times in a running world causes zero
  world-state changes and zero clock disturbance.
- **SC-006**: When the first-occurrence lesson feature later lands, its lessons become
  readable from the overlay by adding content entries only (no overlay structural
  change) — the seam contract holds.

## Assumptions

- The canonical key inventory is the keymap design doc
  (`docs/design/tui/patterns/keymap.md`) plus the client's footer-hint machinery; the
  basic/advanced tier assignment follows the doc's footer-hint priorities (footer-hinted
  keys ≈ basic tier) unless the plan records a better split.
- "Every pane/mode" means the modes the keymap doc names (global/home, minibuffer,
  inspect, villagers roster/detail, solo/narrow fallbacks) plus any mode shipped by the
  time of implementation (e.g. a postmortem posture) — the plan pins the enumeration.
- The overlay is a client-side presentation feature only: no daemon, IPC, event-log, or
  world-state changes are in scope.
- The pull-reference section ships as structure + contract now; lesson entries arrive
  with the future first-occurrence lesson task (TASK-116 AC #4's "when it lands").
- Tier navigation and section navigation key choices are plan-phase decisions, bound by
  the focus contract (esc releases one layer; no silent no-ops inside the overlay).

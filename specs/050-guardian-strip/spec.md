# Feature Specification: Guardian strip — always-visible action budget line

**Feature Branch**: `050-guardian-strip`

**Created**: 2026-07-25

**Status**: Draft

**Input**: User description: "Guardian strip (board TASK-126; reorient 2026-07-25 decision 7, Wave 2). One line above the minibuffer pairing the action budget with the input: charge bank, regen countdown, standing-order count, faith gauge once TASK-118 lands. Makes the minibuffer read as THE verb; today the budget hides in one tab's pane header. Design authority: docs/design/tui/panels/guardian-strip.md (status: specified — the authored page is the behavior contract); patterns/layout.md ruling a (fold order); same-PR doc amendment gate applies."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The budget is one glance from the verb (Priority: P1)

A player about to ask the guardian for something can see, without leaving
whatever tab they're on, whether they can afford it: a single plain line
directly above the minibuffer shows the charge bank (glyphs + numeric), when
the next charge arrives, and how many standing orders are consuming
attention. Today that information hides in one dock tab's pane header; after
this feature it is chrome, present regardless of dock tab.

**Why this priority**: decision 7 verbatim — pairing budget with input makes
the minibuffer read as THE verb; this is the whole feature.

**Independent Test**: attach a world in a widescreen terminal, switch across
all dock tabs, confirm the strip stays visible with live values above the
minibuffer.

**Acceptance Scenarios**:

1. **Given** an attached widescreen client on any dock tab, **When** the
   composite renders, **Then** one borderless row sits directly above the
   minibuffer showing the charge bank (`⚡`-filled/`·`-empty glyphs plus
   `(N/cap)`), the next-regen time, and the standing-order count.
2. **Given** a charge regenerates or is spent, or a standing order is added
   or expires, **When** the next frame renders, **Then** the strip reflects
   the new values (same replica data the guardian tab header already uses).
3. **Given** any curriculum stage or a pre-ladder world, **When** the
   composite renders, **Then** the strip is present — no stage default ever
   hides it (decision 7 "always visible").

---

### User Story 2 - The strip never lies (Priority: P2)

Each segment degrades to absence rather than to a misleading value: no faith
segment exists while the faith mechanic (TASK-118) is unshipped; a full bank
shows no regen forecast (nothing is scheduled to arrive); before the first
status snapshot the strip is blank rather than showing zeros.

**Why this priority**: the design page's core honesty rule ("the strip must
never claim a mechanic that doesn't exist yet"); dishonest chrome is worse
than no chrome.

**Independent Test**: render the strip from fixture states (no status, full
bank, empty orders, pre-TASK-118) and assert segment presence/absence.

**Acceptance Scenarios**:

1. **Given** TASK-118 has not shipped, **When** the strip renders, **Then**
   no faith segment appears at all (not even a dash).
2. **Given** the charge bank is at cap, **When** the strip renders, **Then**
   the regen segment is omitted (no arrival is scheduled while full — the
   regeneration rule only fires below cap).
3. **Given** the client has no status snapshot yet (connecting), **When**
   the composite renders, **Then** the strip row is present but empty — the
   layout stays stable and no zero/placeholder values are invented.

---

### User Story 3 - The strip survives pressure (Priority: P3)

Under terminal-height pressure the strip is the last chrome to fold, and
"folding" relocates its content into the minibuffer's dormant placeholder
line rather than hiding it; in the narrow single-pane fallback the strip is
carried exactly as in widescreen.

**Why this priority**: layout.md ruling a step 4 and ruling b already rule
this; the feature must implement, not reinterpret, those rulings.

**Independent Test**: render composites across a sweep of terminal heights
and widths; assert fold order and the relocated dormant-line form.

**Acceptance Scenarios**:

1. **Given** a terminal short enough that the fold order reaches its final
   step, **When** the composite renders, **Then** the strip row is gone and
   the minibuffer's DORMANT line reads as the relocated form (budget prefix +
   existing placeholder hint); the focused and busy minibuffer states are
   unchanged.
2. **Given** a narrow (single-pane fallback) terminal, **When** any pane
   renders, **Then** the strip appears exactly as in widescreen (ruling b:
   width-independent).

---

### Edge Cases

- **Disconnected mid-session** (status goes stale/lost): segments keep the
  last known replica values exactly as the guardian tab header does today —
  no new staleness semantics invented by this feature.
- **Width too small for all segments on one line**: segments truncate from
  the right with `…` (faith/orders drop before the charge bank — the bank is
  the headline); never wrap to a second row (the row budget is exactly 1).
- **Cap changes** (doctrine constant): glyph run and `(N/cap)` derive from
  the same exported cap the guardian tab uses; no literal `3` anywhere new.
- **Zero standing orders**: segment shows `👁 0 standing orders` — zero is a
  true value here (the mechanic exists and its count is genuinely 0), unlike
  the absence cases in US2.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The widescreen composite MUST render the guardian strip as
  exactly one borderless row directly above the minibuffer, on every dock
  tab, at every curriculum stage and in pre-ladder worlds.
- **FR-002**: The strip MUST render, left to right: charge bank (filled/empty
  glyph run + numeric `(N/cap)`), next-regen time (`next +1 @ <time>`, the
  next absolute regeneration boundary), and standing-order count (`👁 N
  standing orders`), each fed by the same replica/status data existing
  surfaces already use.
- **FR-003**: Segments MUST degrade to absence: no faith segment while the
  faith mechanic is unshipped; no regen segment at full bank; a blank (but
  present) row before the first status snapshot. No invented zeros or dashes
  for missing mechanics.
- **FR-004**: Under height pressure the strip MUST fold last (layout.md
  ruling a step 4), and folding MUST relocate its content into the
  minibuffer's dormant-state line (budget prefix + existing hint), dormant
  state only — focused/busy minibuffer states unchanged.
- **FR-005**: The narrow fallback MUST carry the strip identically
  (layout.md ruling b).
- **FR-006**: The strip MUST introduce no new fiction literals — all segment
  labels are non-fiction chrome (`—` skin-token column); the one fiction
  string in its vicinity (the dormant placeholder's epithet) already exists
  and is not this feature's to change.
- **FR-007**: The design reference MUST be amended in the same PR:
  `panels/guardian-strip.md` flips `status: specified → shipped` with real
  renderer symbols in its control table; `patterns/layout.md`'s row-budget
  and fold-order entries are re-verified; `panels/minibuffer.md`'s dormant
  state gains the relocation form if not already recorded; affected pages
  re-pin `verified_against`. Any behavior this spec fixes that the authored
  page leaves open (e.g. full-bank regen omission, truncation order) is
  recorded onto the page in the same PR.
- **FR-008**: The linear-stream/CLI projection loses nothing (D1): every
  strip value is already exposed via the status/IPC surface; this feature
  adds no new TUI-only information.

### Key Entities

- **Strip segments**: charge bank, regen forecast, standing-order count,
  (future) faith — each an independently present/absent renderable backed by
  existing status fields.
- **Relocated dormant line**: the fold-pressure form — budget prefix
  composed into the minibuffer's existing dormant placeholder.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: From any dock tab, the player reads the charge bank with zero
  keypresses (today: up to 2 keypresses to reach the guardian tab).
- **SC-002**: Across a fixture sweep of client states (no status, full bank,
  partial bank, 0/N orders, pre-faith), 100% of rendered segments are true
  statements — no segment ever names a mechanic or value the world doesn't
  currently have.
- **SC-003**: Across a sweep of terminal heights, the composite's row
  arithmetic matches the design reference's re-derived budget exactly, and
  the strip is the last chrome row to fold; at every width the strip row
  never exceeds 1 row.
- **SC-004**: The design-reference gate (`check-tui-design.mjs --changed`)
  passes with `panels/guardian-strip.md` shipped and re-pinned in the PR.

## Assumptions

- The authored design page `docs/design/tui/panels/guardian-strip.md` is the
  behavior contract; this spec adds only the honest-degradation details it
  left open (full-bank regen omission; pre-status blank row; right-to-left
  truncation) — those rulings get recorded onto the page in the same PR.
- Faith (TASK-118) has not shipped: the segment is omitted entirely; the
  design page's "present, dashed" form activates only when TASK-118's field
  exists on the wire (not this feature's concern).
- Regen time formatting reuses the app's existing game-clock time rendering;
  the boundary is derived from the existing regeneration cadence, not a new
  mechanic.
- Stage-defaults machinery (TASK-128) may not exist yet; "always visible at
  every stage" is trivially satisfied by unconditional rendering until that
  machinery lands (and decision 7 exempts this row from it anyway).
- Dependency: TASK-123 (spec 047) merged — authored page exists; layout.md
  rulings a/b are in force.

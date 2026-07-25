# Feature Specification: Village lens completion — villager strip + map condition overlays

**Feature Branch**: `060-village-lens`

**Created**: 2026-07-25

**Status**: Draft

**Input**: User description: "Village lens completion (board TASK-129; reorient 2026-07-25 D12, Wave 5). A one-row colonist-bar-style villager strip under the header (status glyphs — RimWorld colonist bar precedent), stage-defaulted like other chrome; map condition overlays (needs-critical marker, suppressed-mind marker, dying-fire pulse — Cogmind map-dynamics doctrine); evaluate a look-cursor vs the growing legend line. Design authority: docs/design/tui/panels/villager-strip.md (authored, status: specified — the behavior contract) and panels/map.md's Wave-5 condition-overlay stub (this feature authors the real rows)."

## Standing resolutions this spec adds

1. **The strip is display-only** — the authored page rules it ("no
   selection cursor, no drill-down; answers 'how is the village doing',
   never 'tell me about Ash'"). The board AC's "click/jump follows parity
   doctrine" is satisfied by that ruling: a control with no actions has no
   parity gap (the doctrine's own display-only clause); recorded on the
   board at link time.
2. **Look-cursor: deferred, ruling recorded.** The evaluation the task asks
   for concludes: chronicle jump-to-source (spec 049) plus the legend's
   inspection line already cover "what is that tile" for the shipped
   surfaces; a free-roaming look-cursor adds a fourth inspection modality
   with its own key-mode for marginal gain. Deferred with the reasoning
   recorded on `panels/map.md`; re-open if playtesting shows legend
   overflow pain.
3. **"Pulse" is a style, not animation.** Terminal blink is banned
   (accessibility + terminal variance); the dying-fire condition renders as
   a distinct warn-styled glyph state, steady, exactly like the existing
   damaged-wall/burnt-out-fire faded treatments.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The villager strip (Priority: P1)

Under the header, one borderless row: `N villagers` plus one name-initial
glyph per villager in stable roster order, styled exactly as the map styles
agents (bright awake, dim/lowercase asleep, faint † dead). The whole
village's state in one glance without opening the villagers tab.

**Why this priority**: D12 verbatim — the headline of the village lens.

**Independent Test**: attach widescreen; verify count + glyph run matching
the roster; kill/sleep fixtures re-style glyphs; overflow drops from the
end with a trailing `…` count.

**Acceptance Scenarios**:

1. **Given** a widescreen composite at any stage (and pre-ladder),
   **When** it renders, **Then** the strip occupies exactly 1 row under the
   header: `N villagers` + the glyph run in roster order, map-identical
   styling per state.
2. **Given** more villagers than columns, **Then** glyphs drop from the end
   with a trailing `…` overflow count — never a mid-glyph truncation.
3. **Given** height pressure, **Then** the strip folds SECOND (after the
   map legend, before later chrome) to the header count badge
   `[N villagers]` (layout.md ruling a); **Given** narrow mode, **Then**
   only the header badge form exists (ruling b — not carried).
4. **Given** the strip renders, **Then** it is display-only: no keys, no
   mouse targets, no parity gap (standing resolution 1).

---

### User Story 2 - Map condition overlays (Priority: P1)

The map's glyphs carry glanceable trouble markers: a villager whose needs
are critical renders with a needs-critical marker style; a villager whose
mind was last suppressed for speed renders with a suppressed-mind marker;
a fire close to burning out renders in the dying-fire warn style before it
goes cold. Trouble is visible from orbit, not just in drill-downs.

**Why this priority**: co-P1 — the Cogmind map-dynamics doctrine half of
the lens; independent of US1.

**Independent Test**: fixture replicas with a starving villager, a
suppressed decision trace, and a low-fuel fire; assert the distinct styles
render (and don't when conditions clear).

**Acceptance Scenarios**:

1. **Given** a living villager with any need at its critical threshold
   (the same thresholds the roster gauges already treat as critical),
   **Then** its map glyph renders in the needs-critical style; recovery
   clears it next frame.
2. **Given** a villager whose most recent decision outcome was a
   speed-suppression (the client's existing decision-trace projection),
   **Then** its glyph renders the suppressed-mind marker until a
   non-suppressed outcome follows — the map form of spec 037's "a skipped
   thought is visible".
3. **Given** a lit fire within the dying-fuel window of `FuelUntil`,
   **Then** it renders the dying-fire warn style (steady, no blink),
   distinct from lit and from burnt-out; refueling clears it.
4. **Given** overlapping conditions (needs-critical + suppressed), **Then**
   a documented priority renders one marker (needs-critical wins — physical
   danger over cognitive telemetry), recorded in the page's priority rules.
5. **Given** the legend, **Then** it names the new marker styles (the
   map's existing legend discipline) so the vocabulary is discoverable.

---

### Edge Cases

- **Dead villagers**: strip shows faint † in place (roster parity); map
  overlays never apply to the dead (no needs, no mind).
- **No decision traces yet** (fresh world/no LLM): no suppressed markers —
  absence of telemetry renders nothing (honesty rule); needs/fire overlays
  are LLM-independent and still work.
- **Strip + lesson row + guardian strip stacking** (post-TASK-117): the
  layout page's re-derived row budget already orders all three; this
  feature only adds its own row per that budget.
- **Color-profile floors**: marker styles must remain distinguishable in
  the same profiles the family-tint tests already cover (reuse that test
  discipline).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The widescreen composite MUST render the villager strip
  (1 borderless row under the header): count + name-initial glyph run in
  stable roster order, styled by the map's exact state rules; overflow
  drops from the end with `…N`; display-only.
- **FR-002**: Fold/narrow behavior MUST match layout.md: folds second to
  the `[N villagers]` header badge; narrow shows the badge form only.
- **FR-003**: The map MUST render three condition overlays from replica/
  projection facts: needs-critical (roster-threshold parity),
  suppressed-mind (latest trace outcome suppressed), dying-fire (within
  the dying-fuel window of `FuelUntil`); each clears when its condition
  does; priority needs-critical > suppressed-mind; steady styles only (no
  blink); legend names them.
- **FR-004**: A fixture render suite MUST cover every overlay state ×
  set/clear transitions × the priority rule × color-profile
  distinguishability.
- **FR-005**: The design reference MUST be amended in the same PR:
  `panels/villager-strip.md` specified → shipped (real symbols; overflow +
  display-only rulings); `panels/map.md` Wave-5 stub replaced with the real
  overlay rows + priority rules + the look-cursor deferral ruling;
  `patterns/layout.md` fold rows re-verified; `pages/home.md` header badge
  form recorded; re-pins throughout.
- **FR-006**: D1 holds: every fact the strip/overlays show is already in
  the status/replica stream (roster states, needs, traces, fuel); no new
  wire fields; presentation only.
- **FR-007**: No new fiction literals; any label routes through the skin
  contract (expected: none needed — all chrome vocabulary).

### Key Entities

- **Strip row**: count + per-villager glyph (state-styled), roster-ordered,
  overflow-shedding.
- **Condition overlay**: per-glyph marker style derived per frame from
  {needs thresholds, latest trace outcome, fuel window}; priority-ordered.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Village-wide state (who's up, asleep, dead) readable with
  zero keypresses from any dock tab; the strip matches the roster 1:1 in
  every fixture.
- **SC-002**: 100% of fixture trouble states render their marker and 100%
  of recoveries clear it next frame; overlapping conditions always render
  the documented winner.
- **SC-003**: Layout budgets hold across the height sweep (strip folds
  second; badge form appears; narrow carries badge only).
- **SC-004**: Design gate green with villager-strip.md shipped, map.md's
  real overlay rows, and the look-cursor ruling recorded.

## Assumptions

- Needs-critical thresholds reuse the roster's existing critical
  vocabulary (single source; no new tuning constants — if a threshold
  constant must be named, it comes from the existing sim/tuning surface).
- The dying-fuel window is a small doctrine constant (order of one game
  hour) named beside `FuelUntil`'s existing consumers; recorded in the
  page.
- Suppressed-mind derives from the client's existing decision-trace
  projection (no daemon changes).
- Model tier: Sonnet (single-package rendering; tests alongside).
- Dependencies: spec 047 pages (merged); layout row budget already
  re-derived (spec 050 shipped the fold machinery this extends).

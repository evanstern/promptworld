# Feature Specification: First-occurrence lessons projection (lesson row)

**Feature Branch**: `055-lesson-row`

**Created**: 2026-07-25

**Status**: Draft

**Input**: User description: "First-occurrence lessons projection (TASK-117, reorientation decision 5/D2/D8, Wave 4). A TUI-side teaching layer: client-side projection over the event stream that surfaces one-line lessons the first time a player encounters a mechanic or prompting moment, in a dedicated lesson row above the guardian strip, with per-user seen-state, anti-spam, a pull path through the ? help overlay, skin-tokened strings, and stage-defaulted visibility. Authority: docs/design/tui/panels/lesson-row.md."

## Design authority

`docs/design/tui/panels/lesson-row.md` is the verbatim behavior contract for this
feature (authored spec-before-build by spec 047; reorientation decision 5, D2, D8).
Where this spec and that page could ever disagree, the page wins and this spec must be
amended. Related authorities: `docs/design/tui/patterns/stage-defaults.md` (visibility
defaults), `docs/design/tui/patterns/layout.md` rulings (a)/(b) (fold order, narrow
carry), `docs/design/tui/patterns/keymap.md` (the `x` binding, documented unbuilt),
`docs/design/tui/overlays/help.md` (the pull half's `helpLesson` seam), and spec 052's
published skin contract (`specs/052-skinnable-guardian/contracts/skin-contract.md`) for
tokened strings.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A first encounter teaches itself, exactly once (Priority: P1)

A new player is watching their village when something happens for the first time — a
thought is skipped for speed, the gru attacks, a standing order expires. A two-line
lesson appears in a dedicated row above the guardian strip: one plain-language sentence
explaining what just happened, and a pointer line naming the key or tab that
demonstrates it, suffixed with how to find it again and how to clear it
(`(? for more · x dismiss)`). The lesson dwells until the player performs the
pointed-at action or presses `x`. Once seen, that lesson never appears again — not in
this world, not in any other world, not after a restart.

**Why this priority**: This is the feature — the push half of first-occurrence
teaching. Without the row, triggers, and never-repeat invariant, nothing else here has
a surface to live on.

**Independent Test**: Boot a fresh TUI against a world emitting a covered
first-occurrence event; the lesson appears with its pointer and suffix; dismiss it with
`x`; re-trigger the same event (same world and a second world) and confirm the lesson
never reappears; restart the client and confirm the seen-state survived.

**Acceptance Scenarios**:

1. **Given** a player who has never seen the suppression lesson, **When** the first
   `cog.outcome{suppressed}` event reaches the client, **Then** the lesson row shows
   the suppression lesson's text line and pointer line with the
   `(? for more · x dismiss)` suffix.
2. **Given** an active lesson is showing, **When** the player presses `x`, **Then**
   the row clears and that lesson id is recorded as seen in the per-user record.
3. **Given** a lesson id already recorded as seen, **When** its trigger event arrives
   again (any world, any session), **Then** no lesson is shown for it.
4. **Given** an active lesson whose pointed-at action has a done-signal (e.g. the
   standing-order lesson), **When** the player performs that action, **Then** the
   lesson clears without needing `x` and is recorded as seen.
5. **Given** a lesson is mid-dwell, **When** a different first-occurrence trigger
   fires, **Then** the new lesson queues (does not replace the active one), and if its
   moment goes stale before it can surface, it decays instead of appearing late.
6. **Given** the per-user seen-state file is missing, unreadable, or corrupt, **When**
   the client boots, **Then** the client runs normally (advisory, load-tolerant — the
   worst case is a repeated lesson, never a crash or a blocked boot).

---

### User Story 2 - The player's own prompting practice is taught (Priority: P2)

Beyond UI mechanics, the row teaches the game's actual subject — prompting the
guardian. The first time the player's tool call is rejected, the first time a custom
charter is observed, and the first time a fuzzy standing order is placed, a
prompting-tier lesson explains what that moment means (refusals are informative;
editing the charter changes the guardian's voice; vague orders still bind, marked
honestly).

**Why this priority**: The reorientation's teaching core ("the whole game is
prompt-engineering the guardian") — but it rides the same machinery as US1, so it
ships as catalog entries once US1 exists.

**Independent Test**: Trigger each prompting-tier event
(`cog.tool_call` verdict ≠ landed, `metatron.charter_observed{default: false}`,
`metatron.order_placed{fuzzy: true}`) for a fresh user; each fires its lesson exactly
once.

**Acceptance Scenarios**:

1. **Given** a fresh user, **When** their first non-landed `cog.tool_call` verdict
   arrives, **Then** the rejected-tool-call lesson appears (once, ever).
2. **Given** a fresh user, **When** `metatron.charter_observed{default: false}` first
   arrives, **Then** the custom-charter lesson appears (once, ever).
3. **Given** a fresh user, **When** `metatron.order_placed{fuzzy: true}` first
   arrives, **Then** the fuzzy-order lesson appears (once, ever).

---

### User Story 3 - Lessons are always findable again, and the row knows its place (Priority: P3)

A player who dismissed a lesson (or plays at a stage where the row defaults off) can
re-read every lesson from the `?` help overlay's lessons section — the pull half of the
seam. The row itself respects the composite's design system: default-on at stages 1–2,
badge + overlay-only at stage 3+ and pre-ladder; folds third under height pressure to a
`[lesson]` header badge; carried with identical defaults in the narrow fallback.

**Why this priority**: Completes the push/pull contract and the layout discipline;
without it, dismissed lessons are unrecoverable and the row misbehaves in constrained
layouts — but the core teaching loop (US1/US2) works without it.

**Independent Test**: Dismiss a lesson, open `?`, cycle to the lessons section, and
read it there; set stage to 3+ and confirm the row folds to the `[lesson]` badge while
the overlay still lists the content; shrink the terminal below the fold threshold and
confirm the row folds third per the layout ruling.

**Acceptance Scenarios**:

1. **Given** any lesson that has ever been pushed (active, done, or dismissed),
   **When** the player opens `?` and reaches the lessons section, **Then** that
   lesson's title and body are listed there (the overlay's placeholder line is replaced
   by real entries).
2. **Given** a world at curriculum stage 1 or 2, **When** the composite renders,
   **Then** the lesson row is present by default; **Given** stage 3+ or a pre-ladder
   world, **Then** the row defaults to the `[lesson]` header badge with content
   reachable via `?`.
3. **Given** widescreen height pressure that would push body rows below the layout
   minimum, **When** folding occurs, **Then** the lesson row folds third (after map
   legend and villager strip) to the `[lesson]` badge.
4. **Given** a terminal narrower than the layout breakpoint, **When** the narrow
   fallback renders, **Then** the row is carried with the same stage defaults.

---

### Edge Cases

- Two covered first-occurrence events arrive in the same poll/batch: one becomes
  active, the other queues; queue order is arrival order; the queued one decays if its
  window lapses.
- A trigger event arrives while the help overlay is open (overlay owns the keyboard):
  the lesson row updates beneath normally — the overlay is body replacement, not world
  pause; `x` while the overlay is open belongs to the overlay's inert-key rule, not the
  row.
- The player presses `x` with no active lesson: strict no-op (falls through per the
  keymap; no global `x` binding exists otherwise).
- Seen-state write fails (read-only home dir): the client continues; the lesson may
  repeat next session — advisory, never authority.
- A covered event fires for a fresh user in a world already mid-run: the first arrival
  after this user's client boots still counts as the user's first occurrence.
- Skin token resolution missing/default skin: strings render with the default skin's
  values (spec 052's boot-frozen resolution) — never a raw `{{…}}` literal on screen.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The client MUST render a dedicated lesson row above the guardian strip:
  two lines, no border — line 1 the lesson text, line 2 the UI-pointer phrase plus the
  `(? for more · x dismiss)` pull-path suffix — exactly per the panel page's mockup.
- **FR-002**: The client MUST derive lessons purely client-side, as a projection over
  event types already on the feed (the decision-trace precedent) — no daemon changes,
  no new event types, no IPC commands, no model calls; lesson text is static and
  model-free.
- **FR-003**: The trigger taxonomy MUST cover at minimum: mechanics tier —
  first suppression (`cog.outcome{suppressed}`), first gru attack (`gru.attacked`),
  first charge regeneration, first order expiry (`metatron.order_expired`), first
  death; prompting tier — first rejected tool call (non-landed `cog.tool_call`
  verdict), first custom charter observed (`metatron.charter_observed{default:
  false}`), first fuzzy order (`metatron.order_placed{fuzzy: true}`).
- **FR-004**: Exactly one lesson MAY be active at a time; a new trigger while one is
  active queues; a queued trigger that cannot surface within its decay window is
  dropped (opportunity decay), never shown stale.
- **FR-005**: An active lesson MUST dwell until (a) its done-signal (the pointed-at
  action occurring) or (b) explicit dismissal via `x`; both outcomes record the lesson
  as seen.
- **FR-006**: Seen-state MUST be per-user (not per-world), persisted beside
  `~/.promptworld/unlocks.json` following its precedent: load-tolerant, advisory-never-
  authority, atomic write. A lesson id present in the record never fires again.
- **FR-007**: Every lesson (regardless of push state) MUST be readable from the `?`
  help overlay's lessons section via the existing `helpLesson{id, title, body}` seam —
  the push and pull halves share one catalog, never two hand-maintained lists.
- **FR-008**: Every lesson string MUST resolve skin tokens through spec 052's published
  skin contract (e.g. `{{skin.guardian.epithet}}`) and MUST carry its pull path in the
  rendered suffix.
- **FR-009**: Row visibility MUST follow `patterns/stage-defaults.md`: on at stages
  1–2; `[lesson]` header badge + overlay-only at stage 3+ and pre-ladder. Folding MUST
  follow `patterns/layout.md` ruling (a) (folds third, to the same badge) and the row
  is carried in the narrow fallback per ruling (b).
- **FR-010**: `x` MUST dismiss only an active lesson and be a strict no-op otherwise;
  the binding lands in the TUI keymap and its design-page control table flips from
  `unbuilt` in the same change (the spec-047 gate).
- **FR-011**: With no world state loaded, no LLM configured, or an empty/missing
  seen-state record, the client MUST behave normally — the projection renders nothing
  until a trigger arrives and degrades to at-worst-repeated lessons, never an error.

### Key Entities

- **Lesson catalog entry**: static record — id, title, body (skin-tokened), trigger
  predicate over cataloged event types, UI-pointer phrase, optional done-signal
  predicate, tier (mechanics | prompting).
- **Per-user seen record**: set of lesson ids already shown; stored per-user beside
  `unlocks.json`; advisory.
- **Lesson row state**: none | showing | dwelling; plus a bounded queue of pending
  triggers with decay deadlines.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: For every trigger in the minimum taxonomy (8 lessons), a fresh user sees
  the lesson exactly once across two worlds and a client restart — 0 repeats, 0 missed
  first occurrences, verified by an automated sweep over recorded event fixtures.
- **SC-002**: 100% of cataloged lessons are listed in the `?` overlay's lessons
  section, mechanically verified (one catalog, both surfaces) — the overlay's
  placeholder line no longer renders when the catalog is non-empty.
- **SC-003**: The lesson row never exceeds 2 terminal rows in any mode, stage, or
  width, and folds to the `[lesson]` badge in exactly the layout-ruled order —
  verified by rendering tests at the layout boundaries.
- **SC-004**: A client with a missing, corrupt, or read-only seen-state file boots and
  runs with zero errors surfaced to the player (advisory invariant holds).
- **SC-005**: With the default skin, no rendered lesson contains a raw `{{` token
  literal; with a custom skin, guardian-referencing lessons reflect the skin's values.

## Assumptions

- The per-user seen record lives in its own file beside `unlocks.json` (e.g.
  `~/.promptworld/lessons-seen.json`); deleting the file is the (undocumented-to-player)
  reset mechanism. Reset semantics beyond file deletion are out of scope (the board
  task marks this a minor open question — file-delete is the D8-precedent default).
- Dwell, spacing, and decay windows are client-side constants tuned at implementation
  time (no config surface in v1); the contract is the ordering/one-active/decay
  behavior, not specific durations.
- "First charge regen" and "first death" map to the existing cataloged event types for
  those facts (charge regeneration and villager death are already on the feed); the
  implementation binds to the actual catalog names at build time — the design page's
  taxonomy is the authority, not exact event-name spellings in this spec.
- Scenario-world scheduled incidents (TASK-119) will fire these same triggers through
  the same events; no special-casing is needed or permitted here.
- The stage value used for defaults is the same one the stage-defaults machinery
  (TASK-128) will consume; until TASK-128 lands, this feature reads the stage directly
  and applies its own row default, leaving the shared-machinery refactor to TASK-128.

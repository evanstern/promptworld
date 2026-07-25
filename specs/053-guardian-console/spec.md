# Feature Specification: Guardian console page + systems-tab telemetry split

**Feature Branch**: `053-guardian-console`

**Created**: 2026-07-25

**Status**: Draft

**Input**: User description: "Guardian console page + systems-tab telemetry split (board TASK-125; reorient 2026-07-25 decisions 1/2, D5/D10, Wave 3). The guardian conversation becomes a first-class full-height page: document-style turns, composer (the existing minibuffer presented as the console's input), charter/skills READ surface with binding status + honest lock notices, $EDITOR write handoff with 'charter changed — next turn binds it' confirmation. Provider table, horizon rows, and spend move to a new systems dock tab (D10) — the guardian tab becomes fiction-only, making the TASK-121 skin boundary a file boundary. Design authority: docs/design/tui/pages/guardian-console.md, panels/systems.md, panels/guardian.md, panels/dock.md (all authored by spec 047); same-PR doc amendment gate applies."

## Scope ruling (report card)

The inline report-card CARD is not this feature's to build: D5 assigns the
report card itself to TASK-115, and the shared report-card renderer is
authored on `overlays/postmortem.md` (TASK-127). This feature ships the
console's **card composition seam** — the documented inline slot between
turns at stopping points — wired to nothing visible until those tasks land
their renderer (the chronicle-⏎ reserved-seam precedent). The board task's
AC is reworded accordingly at link time, citing D5.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The conversation gets a room of its own (Priority: P1)

A player deep in conversation with the guardian presses `G` from anywhere
and gets a full-height page: each turn a labeled block (speaker + timestamp,
width-wrapped text, blank-line separated), the special-row vocabulary
(⚡ omen/vision, 👁 watch, ⏲ clock, » verdicts) carried over from the compact
tab, the same minibuffer presented as the composer beneath the stream.
`G` again (or `1`/`esc`) returns to what was open before. The compact dock
tab remains untouched for glancing.

**Why this priority**: decision 1 verbatim — the pivot's central verb gets
the app's biggest surface. Everything else composes onto this page.

**Independent Test**: attach, press `G`, converse; verify document-style
rendering, composer behavior identical to the minibuffer everywhere else,
and round-trip navigation.

**Acceptance Scenarios**:

1. **Given** any mode (home, solo, narrow), **When** the player presses `G`,
   **Then** the guardian console fills the terminal (header line + turn
   stream + composer + footer) and `G`/`1`/`esc` returns to the prior view.
2. **Given** the console is open, **When** turns exist in the transcript,
   **Then** each renders as a labeled block (`you · 08:04` /
   `<epithet> · 08:04`) with the shared special-row vocabulary rendered
   inline in stream order — same data as the compact tab, richer layout.
3. **Given** the console is open, **When** the player presses `m`,
   **Then** the standard minibuffer focuses (dormant/focused/busy states,
   focus contract, IPC transport all unchanged); a sent turn appears in the
   stream exactly as it would in the compact tab.
4. **Given** a reply arrives while the console is closed, **When** the dock
   renders, **Then** the existing unseen-badge behavior is unchanged (the
   console adds no second badge system).

---

### User Story 2 - Telemetry moves out; the skin boundary becomes a file boundary (Priority: P1)

An operator checking provider health selects the new **systems** dock tab
(key `5`): provider table, health-condition rows, horizon block, and
spend/budget wallet line render there — exactly the content that today
crowds the guardian tab. The guardian tab now carries fiction-layer content
only (pane header, transcript, standing orders, instruction provenance).
Nothing about the telemetry's rendering changes — it moves.

**Why this priority**: D10 — co-P1 because it simultaneously decongests the
console's dock twin and draws TASK-121's skin boundary as a file boundary
(skins may touch the guardian tab, never systems).

**Independent Test**: attach a world with LLM config; select systems tab —
provider/horizon/spend render; select guardian tab — no telemetry remains.

**Acceptance Scenarios**:

1. **Given** the widescreen dock, **When** the player presses `5`, **Then**
   the systems tab renders the provider table (up/down glyphs, queue,
   inflight/slots, contended marker, per-provider spend), health continuation
   lines, the `(unattributed)` row + wallet line, and the horizon block —
   the same renderers, relocated.
2. **Given** the guardian tab after the split, **When** it renders, **Then**
   no provider/horizon/spend content appears there — fiction-layer content
   only.
3. **Given** a no-LLM world, **When** the systems tab renders, **Then** the
   horizon block is absent entirely (existing behavior preserved) and the
   tab states what it has honestly rather than rendering empty chrome.
4. **Given** narrow mode, **When** the player cycles panes, **Then** systems
   is reachable exactly like every other tab (solo/narrow pane, no new
   narrow-specific rendering).
5. **Given** existing keys `2`/`3`/`4` and solo-zoom (same key twice),
   **When** `5` is added, **Then** all existing dock behaviors (badges,
   zoom, per-tab state) extend to it without regression.

---

### User Story 3 - Charter and skills readable in place, edited in $EDITOR (Priority: P2)

On the console, a bordered sub-panel shows `charter.md`'s provenance
(default / player-authored / preset-locked) and binding status, plus the
skills file count and its binding status (locked below stage 3, honest lock
notices naming the unlocking stage). Pressing `e` shells out to `$EDITOR` on
the real `charter.md`; on return, if the file changed, one line confirms:
"charter changed — next turn binds it". There is no in-TUI text editor.

**Why this priority**: decision 2's ruling made visible; P2 because the
console (US1) is valuable for reading/conversing without it.

**Independent Test**: open console on worlds in different stages/presets;
verify provenance lines; press `e`, edit, return; see the confirmation.

**Acceptance Scenarios**:

1. **Given** a pre-ladder world with a player-edited charter, **When** the
   read surface renders, **Then** it shows player-authored + binds-now, and
   the skills line shows its count and binding status.
2. **Given** a stage-1 world, **When** the read surface renders, **Then**
   the charter line shows the preset lock honestly (the same
   stage-vocabulary the status surface already carries) and skills show
   locked-until-stage-3.
3. **Given** the player presses `e`, **When** `$EDITOR` exits with the file
   changed, **Then** the console shows "charter changed — next turn binds
   it" once; unchanged file → no confirmation. The running world is never
   paused or disturbed by the shell-out beyond the terminal handoff.
4. **Given** `$EDITOR` is unset, **When** the player presses `e`, **Then**
   an honest hint names the missing variable (no crash, no silent no-op).

---

### Edge Cases

- **Console during narrow mode**: `G` still opens it full-screen (it IS a
  full-screen page); return lands back in the narrow pane the player left.
- **Console + takeover overlays** (help today; ceremony/postmortem later):
  overlays replace the body per the existing overlay slot rules; the console
  is a page, not an overlay — `?` over the console behaves as over any page.
- **Transcript longer than the page**: the stream shows the most recent
  turns (tail-anchored) with scrollback via the page's own scroll keys;
  scroll state resets on close (reading posture, not archive — the
  transcript file remains the archive).
- **$EDITOR exits nonzero**: treat as no-change (no confirmation), plus an
  honest one-line notice; never partially apply anything (the file is the
  file — the TUI only observed it).
- **Busy minibuffer when `e` pressed**: `e` is a console-page key in the
  global mode; while the minibuffer is focused it types into the buffer
  (focus contract rule — no silent stealing).
- **Turn arrives while shelled out**: the replica catches up on return
  (existing reconnect/refresh semantics); no special handling.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: A new global key `G` MUST open the guardian console as a
  first-class full-screen page from home, solo, and narrow modes; `G`
  (toggle), `1`, or `esc` (nothing focused) MUST return to the prior view.
  The dock's solo-zoom state machine is unchanged (the console is not a
  zoomed tab).
- **FR-002**: The console MUST render: the guardian pane-header line (same
  data source as the compact tab), document-style turn blocks (speaker +
  timestamp labels, width-wrapped, blank-line separated) over the shared
  transcript data, and the shared special-row vocabulary inline (⚡/👁/⏲/»)
  — one vocabulary, two renderings (compact vs document).
- **FR-003**: The console's composer MUST be the existing minibuffer
  component presented beneath the stream — states, focus contract, and
  transport byte-identical to every other page (no second input widget).
- **FR-004**: The console MUST include the charter/skills read surface:
  charter provenance (default/player-authored/preset-locked) + binding
  status, skills count + binding status, with honest lock notices naming the
  unlocking stage — all from the existing status surface (no new file
  parsing in the client).
- **FR-005**: `e` on the console MUST shell out to `$EDITOR` on the world's
  real `charter.md`, suspending and restoring the TUI cleanly; on return
  with a changed file, exactly one confirmation line renders ("charter
  changed — next turn binds it"); unchanged → nothing; `$EDITOR` unset or
  failing → honest notice. No in-TUI editor exists.
- **FR-006**: The console MUST carry the inline card seam: a documented
  composition slot between turns at stopping points (run end, pause,
  exercise resolution), wired to the shared card renderer interface but
  rendering nothing until TASK-127/115 land card content (scope ruling
  above).
- **FR-007**: A fourth dock tab **systems** (key `5`) MUST render the
  relocated telemetry: provider table + health continuation lines +
  `(unattributed)` + wallet line + horizon block — the same existing
  renderers, moved; the guardian tab MUST retain fiction-layer content only.
- **FR-008**: All existing dock behaviors (per-tab state, unseen badge,
  same-key solo zoom, narrow reachability) MUST extend to the systems tab
  without regression; the tab itself carries no skin tokens (never-skinned
  by construction, D10).
- **FR-009**: The design reference MUST be amended in the same PR:
  `pages/guardian-console.md` and `panels/systems.md` flip
  `status: specified → shipped` with real renderer symbols;
  `panels/guardian.md` (content list minus telemetry), `panels/dock.md`
  (4-tab row, key `5`), `patterns/keymap.md` (`G`, `5`, `e` bindings; parity
  gaps recorded), `pages/solo-views.md` (narrow reachability) re-verified;
  all touched pages re-pinned.
- **FR-010**: Linear-stream/CLI projection (D1) unchanged and sufficient:
  the CLI guardian conversation and status already expose everything the
  console renders; the systems content remains in status output. This
  feature adds presentation, not capability.
- **FR-011**: Footer hints MUST advertise the console (`G`) per the keymap
  page's footer-hint discipline, and the console's own footer carries its
  return/ask/help hints.

### Key Entities

- **Console page state**: open/closed + scroll position + prior-view return
  target; no persistence.
- **Document-style turn block**: labeled rendering over the existing
  transcript entries (shared data, new layout).
- **Card seam**: the inline composition slot + renderer interface cards plug
  into later (TASK-127/115).
- **Systems tab**: the fourth dock tab owning the relocated telemetry
  renderers.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: From any mode, the player reaches the full conversation in one
  keypress (`G`) and returns in one; conversation reading no longer competes
  with telemetry for rows.
- **SC-002**: After the split, the guardian tab renders zero telemetry
  rows and the systems tab renders 100% of the relocated content — verified
  by render tests over LLM and no-LLM fixture worlds.
- **SC-003**: The composer behaves byte-identically to the minibuffer
  elsewhere (state-machine tests reused/extended, zero focus-contract
  regressions).
- **SC-004**: The charter read surface agrees with the status surface on
  100% of provenance/lock fixtures (default, player-authored, tutor-preset
  stage-1, stage-2, stage-3+); the $EDITOR round-trip confirmation appears
  exactly when the file changed.
- **SC-005**: The design-reference gate passes with both new-surface pages
  flipped to shipped and every touched page re-pinned in the PR.

## Assumptions

- Systems tab key is `5` (positional continuation of 2/3/4; digits are the
  dock's established grammar). The exercise panel (TASK-119) will take `6`
  or its own ruling — not this feature's concern beyond leaving room.
- `G`/`e` ship keyboard-only, recorded as parity gaps from birth per the
  console page's parity-rollout note (decision 8 incremental rollout).
- The console reads everything from the replica/status it already has;
  the charter read surface requires no new IPC fields (provenance, lock,
  skills facts are already on the status surface per spec 046).
- $EDITOR handoff uses the TUI framework's standard suspend/exec/restore
  mechanism; terminal state restoration is its responsibility.
- Scrollback keys on the console follow the existing scroll vocabulary
  (`J`/`K` family) — final binding recorded in keymap.md in the same PR.
- Dependencies: TASK-123 merged (pages exist). TASK-121/124/126 may merge
  before or during this task — expect rebases; the guardian tab's fiction
  strings may become skin tokens mid-flight (take main's side and re-verify).

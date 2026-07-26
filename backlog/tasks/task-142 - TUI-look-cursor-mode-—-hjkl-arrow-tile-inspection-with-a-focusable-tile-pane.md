---
id: TASK-142
title: TUI look-cursor mode — hjkl/arrow tile inspection with a focusable tile pane
status: To Do
assignee: []
created_date: '2026-07-26 04:28'
labels: []
dependencies: []
references:
  - 'https://claude.ai/code/artifact/d998d6d2-bb8f-4d2c-9689-cde5b7de1961'
  - docs/design/tui/panels/map.md
  - docs/design/tui/patterns/keymap.md
  - docs/design/tui/patterns/focus-contract.md
  - docs/design/tui/panels/dock.md
ordinal: 112000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Add a look-cursor mode to the map: the player highlights a tile, moves it with hjkl/arrows, sees details of the tile under the cursor in the right pane (dock), and can switch focus into that pane to dig further (agents' needs/intents/journal, chest/pile contents, recent events with raw-JSON drill-in).

**Re-opens a recorded deferral.** docs/design/tui/panels/map.md ("Look-cursor: evaluated and deferred", spec 060 standing resolution 2) parked exactly this feature pending a playtesting signal. Operator request 2026-07-26 is that signal — this task re-opens it; the map.md resolution note must be amended in the same PR.

**Design (mocked and operable):** https://claude.ai/code/artifact/d998d6d2-bb8f-4d2c-9689-cde5b7de1961
- `v` enters/exits the mode (binding-selection call-out: `l` is cursor-right in-mode, "inspect" already names the chronicle mode → `v`iew). Cursor spawns at camera center; title becomes `MAP · cursor (x,y) · c center · esc exit`.
- `h/j/k/l` + arrows move 1 tile; `H/J/K/L` jump 8; arrows override camera-pan only while the mode is active; cursor pushes the camera at a 2-tile viewport margin; `c` snaps camera to cursor; exiting resumes centroid-following.
- The dock body is borrowed by a transient TILE view (solo-zoom-seam style, NOT a numbered tab); tab row shows a highlighted `TILE (x,y)` pseudo-label; `2`–`5` exit the mode and select that tab; chronicle selection state is preserved across the borrow.
- `⏎`/`tab` focuses the tile pane (amber border — focus is drawn, focus-contract rule 2); `j/k` select rows, `⏎` drills in (agent → villager-detail renderer family; event → raw JSON debug inspector per FR-020; chest/pile → contents).
- esc chain, one layer per press: drill-in → pane rows → map cursor → exit mode (focus-contract rule 3).
- No new text input anywhere — minibuffer stays the focus contract's "exactly one client"; this is a key mode layered like inspect/villagers, below takeover/help/minibuffer in handleKey.
- j/k conflict resolution: the TILE view is what's visible in the dock, so scoping follows the villagers-mode precedent; no chronicle-inspect collision.
- Input parity (decision 8): ships keyboard + mouse together — click a tile moves the cursor (entering the mode if inactive); click a pane row selects/drills. First map control with a real mouse target.

**Process:** non-trivial → full Spec Kit (specify → clarify → plan → tasks) with spec-bridge:link BEFORE implementation; one task, one branch, one PR from .worktrees/; claim commit per spec 065. PR must run scripts/check-tui-design.mjs --changed and amend docs/design/tui/ in the same PR (map.md deferral note, keymap.md new mode table + v binding, dock.md borrow seam, focus-contract.md scope note, anatomy.md).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 v toggles look-cursor mode; cursor moves with hjkl+arrows (shift = 8-tile jump) and pushes the camera at the viewport margin
- [ ] #2 While the mode is active the dock shows a transient TILE view (terrain, structures, piles, agents with needs/intent, recent events at that tile); 2–5 exit the mode and restore the chosen tab with prior state intact
- [ ] #3 ⏎/tab focuses the tile pane (focus drawn per focus-contract rule 2); j/k+⏎ select and drill into agents/contents/events; esc releases exactly one layer per press back to following
- [ ] #4 No new focusable text input; focus contract's 'exactly one client' claim remains true and its page says so
- [ ] #5 Mouse parity ships with keyboard: click-tile moves/enters the cursor, click-row selects/drills (decision 8 rule 1)
- [ ] #6 docs/design/tui re-verified and amended in the same PR (map.md deferral re-opened, keymap.md, dock.md, focus-contract.md, anatomy.md); check-tui-design.mjs --changed passes
<!-- AC:END -->

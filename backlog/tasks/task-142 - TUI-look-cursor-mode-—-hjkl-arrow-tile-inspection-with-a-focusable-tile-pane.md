---
id: TASK-142
title: TUI look-cursor mode — hjkl/arrow tile inspection with a focusable tile pane
status: Done
assignee: []
created_date: '2026-07-26 04:28'
updated_date: '2026-07-26 19:28'
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

Spec: specs/074-look-cursor
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 v toggles look-cursor mode; cursor moves with hjkl+arrows (shift = 8-tile jump) and pushes the camera at the viewport margin
- [x] #2 While the mode is active the dock shows a transient TILE view (terrain, structures, piles, agents with needs/intent, recent events at that tile); 2–5 exit the mode and restore the chosen tab with prior state intact
- [x] #3 ⏎/tab focuses the tile pane (focus drawn per focus-contract rule 2); j/k+⏎ select and drill into agents/contents/events; esc releases exactly one layer per press back to following
- [x] #4 No new focusable text input; focus contract's 'exactly one client' claim remains true and its page says so
- [x] #5 Mouse parity ships with keyboard: click-tile moves/enters the cursor, click-row selects/drills (decision 8 rule 1)
- [x] #6 docs/design/tui re-verified and amended in the same PR (map.md deferral re-opened, keymap.md, dock.md, focus-contract.md, anatomy.md); check-tui-design.mjs --changed passes
- [x] #7 Every tile's inspector header lists warmth and light levels (meter + plain-language note: fire radius, shelter cover, open water; daylight, canopy, indoors, firelight); light may need a small sim-side derivation to expose
- [x] #8 Map and dock panel geometry is fixed (layout.md column budget) — entering the mode, focusing the pane, and drilling in swap content only, never panel size
- [x] #9 TILE pane lists contents in DF's fixed hierarchy: agents → piles/chests → structures → terrain (stable scan order)
- [x] #10 TILE pane whatis prose comes from the spec-068 tile registry's meaning rows (internal/tui/tiles.go) — plain language per FR-020; the look-cursor becomes the third in-place lookup after ? and explain
- [x] #11 Spec phase: Setup — shared substrate (no behavior change)
- [x] #12 Spec phase: Foundational — mode state, key layer, borrow seam
- [x] #13 Spec phase: User Story 1 — cursor movement + camera (P1) 🎯 AC #1, AC #8
- [x] #14 Spec phase: User Story 2 — the TILE view (P1) 🎯 AC #2, AC #7, AC #9, AC #10
- [x] #15 Spec phase: User Story 3 — pane focus, drill-in, esc chain (P2) 🎯 AC #3, AC #4
- [x] #16 Spec phase: User Story 4 — mouse parity (P2) 🎯 AC #5
- [x] #17 Spec phase: User Story 5 — badge deep-link (P3) 🎯 help.md layer-2 row
- [x] #18 Spec phase: In-app reference completeness (FR-014)
- [x] #19 Spec phase: Same-PR gate obligations — design reference, grounding, verification
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Design revision 2026-07-26 (operator review of the mock): (1) dropped the dashed real-client camera-window ghost from the mock — camera-follow rule stays in prose; (2) fixed-geometry rule made explicit: mode changes never resize the map/dock panels; (3) added per-tile warmth + light levels to the TILE inspector header. Mock republished (same URL, version 'fixed-geometry-env-levels').

Reorient 2026-07-26 decision 4: runs in the next UI lane. Cross-grounding amendments: DF fixed tile-content hierarchy AC + tile-registry meaning rows as whatis content (Game-UI-UX + Game-Player-Docs joint framing). Badge deep-link (overlays/help.md's remaining unbuilt layer-2 row) folds into this lane. Reverse jump (strip glyph/roster row → camera center) is the delta's one net-new unscheduled rec — home decided at spec time (rider here vs TASK-154 vs own card).

Sweep claim (runbook docs/design/reorient-2026-07-26-sweep-runbook.md): spec 074-look-cursor. Tier: Sonnet — view/rendering feature in internal/tui; AC7 light-level derivation is a read-only sim helper, inside the routine tier; escalation to Opus only via the rubric as an operator checkpoint. Runbook defaults at spec time: reverse-jump stays UNSCHEDULED (open question 4); pull-surface budget tension recorded in spec, no navigation ruling (open question 3).
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Merged via PR #115 (merge commit 011ce4e). Look-cursor mode shipped: v toggles, hjkl/HJKL+arrows with camera push/snap, dock TILE borrow (DF hierarchy, tile-registry whatis, sim.EnvAt warmth/light header), focus+drill esc-chain (one layer per press), full mouse parity proven via two oracle entries in the mechanized sweep, badge deep-link built (closing TASK-150's pending retag). map.md spec-060 deferral re-opened+amended; 6 design pages, 26 wiki notes, 7 player pages re-grounded in-branch. Judgment calls logged on the PR/report: guardianVisible dormancy guard, H/J/K/L+c tracked as honest parity gaps (no natural mouse gesture). Tier: Sonnet as recorded; no escalation.
<!-- SECTION:FINAL_SUMMARY:END -->

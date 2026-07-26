# promptworld — reorientation synthesis: post-cockpit delta (2026-07-26)

**Status:** DRAFT — decisions ratified, cross-grounding in flight; synthesis prose lands
when the four branch analyses are written.
**Run:** `promptworld-2026-07-26-16-25-19` (reorient re-run of `promptworld-2026-07-25-13-58-18`).
**Worktree:** authored on branch `reorient-2026-07-26-ui` (root stays on main; a concurrent
reorient session is active on this repo — isolation is deliberate).
**Lens (verbatim from the 2026-07-25 run):** promptworld is pivoting into a staged
prompting-skills teaching game (curriculum ladder TASK-68, skinnable guardian TASK-121, the
eight ratified learning-game decisions of 2026-07-25) while remaining an ambient,
terminal-first LLM-agent world sim. The UI/UX must: (a) make a living LLM-agent village
legible at a glance; (b) make prompting the guardian the central, rewarding player verb;
(c) teach itself in play; and (d) be fully specifiable — page by page, feature by feature,
control by control — so both humans and AI implementers can build new features and update
existing UI directly from the design doc.
**Corpus:** `research/Game-UI-UX`, `research/Game-Gameplay-Patterns`,
`research/Learning-Game-Design`, `research/Game-Player-Docs`.

## TL;DR

_TODO — lands with the full synthesis after cross-grounding._

## Decisions

Operator decisions, steering round of 2026-07-26 ("accept those recs"; verbatim intent;
FIXED constraints for cross-grounding). Execution note: the implementation sweep starts
only **when the board allows a clean sweep run** (the PDLC-hardening sweep TASK-144/145/147
and other in-flight lanes finish first).

1. **Report-card truth (NEW card, HIGH):** unify every report-card surface (postmortem,
   ceremony, console card) on `sim.EvaluateRubric` — a failed term renders ✗ at the
   postmortem — and author `the-law`'s real evaluator (persist the charter `Default` flag
   into state; the documented blocker at `internal/sim/scenario.go:277`). Sequenced after
   TASK-144 (flaky report-card test touches the same code).
2. **Design-gate semantic lint (NEW card, small):** extend `scripts/check-tui-design.mjs`
   to warn/fail when a `status: shipped` page contains `unbuilt (wave` renderer cells
   (optional: grep-level check that named renderer symbols exist in `internal/tui`); amend
   `overlays/postmortem.md`'s seven stale cells and `panels/exercise.md:110` in the same PR.
3. **TASK-67 fork duel:** promote Medium → HIGH; reframe as the loop's iteration rung —
   all D7 prerequisites shipped; v1 = rubric-first scoreboard sharing `reportCardView` +
   `sim.EvaluateRubric`, then the HTML retelling; dual-TUI stays deferred.
4. **TASK-142 look-cursor:** runs in the next UI lane. Amend the design with the DF fixed
   tile-content hierarchy (agents → piles/chests → structures → terrain) and use the
   spec-068 tile registry's `meaning` rows as the TILE pane's whatis content (plain
   language per FR-020).
5. **Exercise catalog wave:** ladder v1 = 2–3 hand-authored exercises per stage; grow the
   incident vocabulary to ~3 kinds (cold snap, forage blight, stranger/trickster arrival),
   each a reducer-valid event shape indistinguishable from an ambient cause.
6. **In-TUI ladder view:** the forward ladder (identity · concept · earned/next · unlock
   evidence, matching `stages --json`) renders in the `?` guardian section — deterministic
   floor, model-free.
7. **Quickstart first-prompt step:** `getting-started.html` gains an "ask your guardian one
   thing" step (sample ask from the `skin.guardian.example_ask.*` token family); each stage
   page gains a short first-session do-this-then-this block. Content-only, player-docs skill.
8. **Mouse-parity sweep test:** commissioned as a small task — parse the control tables'
   `keys+mouse` column, assert every non-`—` mouse cell has a handler, burn down
   `patterns/keymap.md`'s rollout note.
9. **Lane order: duel before faith.** TASK-67 lands before TASK-118 (which itself lands
   before TASK-112, per TASK-112's AC5/AC6 dependency).
10. **Parked as recorded watch items (not scheduled):** proving-run timing vs the ambient
    drama floor (TASK-14/28/133 posture); rubric-gauge exposure doctrine (all-terms-live
    stands until multi-exercise content exists); teaching-signal instrumentation consent;
    fire-once lesson doctrine (revisit at >15 lessons); retry accommodation; chronicle/world
    search `/` (revisit past ~20 villagers); colorblind/contrast palette research; player-docs
    pin de-churn; TASK-146 → Medium reassessment after TASK-145 merges.

## Merged positions

_TODO — written from the four cross-grounded analyses._

## Course of action

_TODO._

## Board moves

_TODO — table lands with the synthesis; executed only after operator sign-off, and only
when the board allows a clean sweep run._

## Open questions

_TODO._

---
title: Pattern — stage-shaped layout defaults
class: pattern
status: specified
verified_against: a30ee798ff6cc6316256d7833aead1e8a4c9a849
---

# Pattern: stage-defaults

The authority table for reorientation decision 3: **stage may shape TUI
layout defaults — defaults only.** Every other page in this corpus that
states a stage-default visibility value references this table rather than
restating its own; if a value ever needs to change, it changes here once.

## The ruling, restated precisely

1. Stage shapes what's **visible by default**, never what's **reachable**.
   Every surface below remains reachable at every stage: through the help
   overlay (`overlays/help.md`), through a solo view
   (`pages/solo-views.md`), or (for a folded chrome row) through its own
   pull path back to full content.
2. **Pre-ladder worlds get everything** — the default set for a pre-ladder
   world (`Stage == ""`) is the union of every stage's default-on set, the
   same "ungated, stage-4 semantics" posture `internal/guardian`'s stage
   ceiling already carries ([[curriculum-ladder]]).
3. **Capability locks stay guardian-only** (spec 046 doctrine, untouched by
   this feature): nothing in this table is a capability lock. The stage
   ceiling and the stage-1 charter lock govern what the *guardian* may do;
   this table only governs what the *layout* shows by default. A player at
   any stage can always reach any surface's full content — only its
   always-visible chrome placement varies.
4. These are **layout** defaults; they compose with, but are distinct from,
   the row **fold order** (`patterns/layout.md` ruling a) — see
   "Composition with the fold order" below.

## Per-surface stage defaults

| Surface | Stage 1 | Stage 2 | Stage 3 | Stage 4 | Pre-ladder | Narrow |
|---|---|---|---|---|---|---|
| Lesson row (`panels/lesson-row.md`) | **on** | **on** | badge + overlay-only | badge + overlay-only | badge + overlay-only | same as widescreen (carried, `patterns/layout.md` R3) |
| Guardian strip (`panels/guardian-strip.md`) | **on** | **on** | **on** | **on** | **on** | **on** (carried, R3) |
| Villager strip (`panels/villager-strip.md`) | **on** | **on** | **on** | **on** | **on** | **off** (folds to header count badge, R3 — never carried) |
| Exercise tab (`panels/exercise.md`) | present iff the world carries a scenario | present iff the world carries a scenario | present iff the world carries a scenario | present iff the world carries a scenario | present iff the world carries a scenario | present iff the world carries a scenario (solo-view only, R3) |
| Incident-visibility vocabulary (exercise panel, D4) | `forecast` | `forecast` | `fog` | `fog` | `forecast` (everything) | same as widescreen |
| Systems tab (`panels/systems.md`, once built) | on | on | on | on | on | on |
| Guardian console (`pages/guardian-console.md`) | reachable (own key) | reachable | reachable | reachable | reachable | reachable |
| Help overlay guardian section (D9) | shows stage 1's content | shows stage 2's content | shows stage 3's content | shows stage 4's content | shows the pre-ladder (all-verbs) variant | unaffected by width |
| Unlock ceremony (`overlays/ceremony.md`) | fires stages 1→2, 2→3, 3→4 | ″ | ″ (3→4 only) | never (stage 4 is terminal — nothing unlocks past it) | never (no stage progression exists) | fires identically (takeovers are layout-independent, R3) |
| Postmortem (`overlays/postmortem.md`) | fires on `run.ended`, every world | ″ | ″ | ″ | ″ (ambient/pre-ladder worlds still get the takeover — FR-018's ambient ruling governs its *content*, not whether it fires) | fires identically (R3) |

**Exercise tab and incident vocabulary are world-shaped, not stage-shaped**:
a world either carries a `Manifest.Scenario` block or it doesn't
([[world-save-directory]]); when it does, the tab is present regardless of
the world's `Stage` field, because the two scenarios that ship today
(`first-night`/stage-1, `the-law`/stage-2) already imply their stage by
construction. The incident-visibility *vocabulary value* (`forecast`/`fog`)
is genuinely stage-keyed (D4), independent of which scenario is running.

## Composition with the fold order

Stage defaults decide the **starting** visible set before a terminal's
height forces anything to fold; `patterns/layout.md`'s ruling (a) fold
order (legend → villager strip → lesson row → guardian-strip relocation)
then applies unconditionally on top of whatever that starting set is:

- At **stage 1–2** on a short terminal, all three new chrome rows
  (villager strip, lesson row, guardian strip) start visible, so the fold
  order may need to shed all three in sequence before `bodyMin` is
  satisfied.
- At **stage 3+**/pre-ladder-defaulted-off cases, the lesson row already
  starts folded (badge+overlay), so the fold order has strictly less work:
  only the villager strip and (last) the guardian strip remain foldable.
- The **guardian strip never starts folded** at any stage (decision 7:
  always visible) — it only ever leaves its own row through the fold
  order's relocation step, never through a stage default.

No stage default ever produces a *narrower* fold order than
`patterns/layout.md` states — stage-defaults only changes which rows are
already-folded before fold pressure begins, never the order itself.

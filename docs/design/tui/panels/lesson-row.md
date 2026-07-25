---
title: Panel — lesson row
class: panel
status: shipped
verified_against: cb4a997453db29bbd746f3ae6cace99748cb281e
sources:
  - internal/tui/lessons.go
  - internal/tui/views.go
  - internal/tui/layout.go
  - internal/tui/tui.go
---

# Panel: lesson row

The first-occurrence teaching chrome (reorientation decision 5): one
always-visible-by-default row above the guardian strip that surfaces exactly
one lesson at a time, points at the key/tab that teaches it, and never
crowds the composite. **Shipped** (spec 055, TASK-117, Wave 4):
`lessonTriggers` (the projection, `internal/tui/lessons.go`) feeds
`lessonRowView` (the renderer, `internal/tui/views.go`); the row's chrome
budget lives in `computeRows` (`internal/tui/layout.go`).

## Mockup

```
 A standing order you placed just expired unmet — orders watch for a limited time.
 → press 3, then look for 👁 rows to place another        (? for more · x dismiss)
```

(The lesson shown is `first-order-expired`, one of the shipped catalog's 8
entries, `contracts/lessons-catalog.md` — the original spec-before-build
draft of this page illustrated a slightly different standing-orders sentence
than what actually shipped; this mockup is the real, byte-exact renderer
output for that lesson at the default skin.)

Two lines, **no border** (the 2-row budget `patterns/layout.md` allocates
has no room for a bordered box's own top/bottom rule):

1. **Lesson text** — plain language, one concept, the lesson's full sentence
   (`lessonEntry.Text`, skin-resolved).
2. **UI-pointer + pull-path suffix** — an arrow phrase naming the key/tab
   that demonstrates the concept (`lessonEntry.Pointer`, skin-resolved),
   right-padded with the pull-path suffix every lesson string carries:
   `(? for more · x dismiss)` (`lessonPullSuffix`, appended by the renderer,
   never stored per-entry) — so a player who doesn't act on it immediately
   always knows exactly how to find it again and how to clear it now. Both
   lines are cropped to the panel's width (`clipLine`, the corpus-wide
   overflow discipline) rather than wrapped — a pointer whose own
   pull-path suffix would overflow a narrow terminal is a content-authoring
   concern (`internal/tui/lessons_test.go`'s
   `TestLessonPointerSuffixFitsNarrowWidth` pins every catalog entry's
   pointer+suffix under 80 runes), not a rendering bug.

## Behavior

### One active, dwell, dismiss

- **One active lesson at a time** (`lessonTriggers.active`) — a second
  trigger while one is showing queues rather than replaces it (see
  "Anti-spam" below).
- **Dwells until done or dismissed**: the active lesson stays up until
  either (a) the player performs the pointed-at action (the trigger's own
  "done" signal — e.g. placing a standing order dismisses the standing-order
  lesson, `lessonEntry.Done`), or (b) the player presses `x` to dismiss it
  outright (`lessonTriggers.Dismiss`). An active lesson carries NO timeout of
  its own.
- **UI-pointer field**: every lesson entry carries a pointer describing
  which key or tab demonstrates the concept, rendered as the `→ …` line
  above.
- **Pull-path suffix**: every lesson string, active or archived, ends with
  `(? for more · x dismiss)` (dismissed-and-gone lessons are readable again
  from `overlays/help.md`'s lessons section — the same `helpLesson{id, title,
  body}` seam that page's placeholder used to reserve, now populated 1:1
  from this row's own catalog at every client boot, `populateHelpLessons`).
  This panel is the PUSH half of that seam, `overlays/help.md` is the PULL
  half.

### Anti-spam

- **Spacing**: a newly triggered lesson does not interrupt mid-dwell of the
  current one; it waits its turn. Shipped as `lessonSpacing` = 5 real-world
  seconds after a lesson clears, before the next one may surface — deliberately
  wall-clock, not simulated-tick, time (the anti-spam pacing is about real
  player attention span, independent of the world's simulated speed). The
  gate applies uniformly to a freshly-arriving trigger too, not only a
  promoted queue head — a trigger landing the instant after a clear still
  waits out the gap.
- **Opportunity decay**: a queued-but-not-yet-shown trigger is not held
  indefinitely — shipped as `lessonQueueDecay` = 90 real-world seconds; if
  the active lesson's dwell doesn't clear (or the spacing gap doesn't
  elapse) within that window, the queued opportunity decays (is dropped)
  rather than surfacing stale content much later, disconnected from the
  moment that earned it, and is NOT recorded seen (it may still fire on a
  later first occurrence). A lesson is a *timely* nudge, not a to-do list.
  Both constants are implementation-time values (no config surface, per the
  spec's own assumption) — the contract is the ordering/one-active/decay
  behavior, not these specific durations.
- **Per-user seen state (D8)**: each lesson id is recorded once shown, in a
  per-user record beside `~/.promptworld/unlocks.json`
  ([[curriculum-ladder]]'s precedent — load-tolerant, advisory-never-
  authority, atomic write), shipped as `~/.promptworld/lessons-seen.json`
  (`internal/worlds/lessons.go`) — so a lesson never repeats for a player who
  has already seen it, across worlds and restarts. Marking happens when a
  lesson SURFACES (becomes the active row entry), never when merely queued.

### Trigger taxonomy

Two lesson tiers, both feeding the same one-active/dwell/anti-spam
machinery (`lessonCatalog`, `internal/tui/lessons.go`,
`contracts/lessons-catalog.md`'s 8-entry minimum taxonomy):

- **Mechanics lessons** (5 shipped) — first-occurrence UI mechanics:
  suppression, the gru's first attack, the guardian's action-budget regen,
  a standing order's first expiry, and a villager's first death.
- **Prompting-verb lessons** (3 shipped, the pivot's teaching core) —
  first-occurrence moments in the *player's own prompting practice*:
  - first rejected tool call (a `cog.tool_call` verdict other than landed,
    turn-scoped) — teaches that the guardian's grant is real and refusals
    are informative, not broken.
  - first custom charter observed
    (`metatron.charter_observed{default: false}`) — teaches that editing
    `charter.md` and returning changes the guardian's voice.
  - first fuzzy order (`metatron.order_placed` with the order's `Confirm`
    field true — the wire's `fuzzy` concept, spec.md's Assumptions: "the
    implementation binds to the actual catalog names at build time") —
    teaches that a vaguely-worded standing order still binds, with its
    fuzziness marked honestly.

### Stage defaults

Per `patterns/stage-defaults.md` (the authority — this page never restates
the table, only cites it): **on** at stages 1–2; **badge + overlay-only**
at stage 3+ and pre-ladder. Shipped as `lessonRowDefault(stage string) bool`
(`internal/tui/layout.go`), reading the stage id straight off the polled
status (`Status.World.Stage`, `ipc/protocol.go`) — a small standalone
function, not shared stage-defaults machinery (the spec's own recorded
assumption: TASK-128 absorbs every per-surface default, this one included,
once it lands). The badge is unconditional at stage 3+/pre-ladder — it
shows regardless of whether a lesson happens to be active or queued at that
moment, a quiet permanent affordance rather than a "you have something new"
notification. Folded state: a `[lesson]` header badge (rendered inside the
shared `headerView`, so it appears in both the widescreen and narrow
layouts); content remains reachable via `?` exactly as the fold-order case
below.

### Fold behavior (widescreen height pressure and narrow)

Per `patterns/layout.md`'s ruling (a), this row folds BEFORE the guardian
strip — the relative order restricted to the two foldable rows this
feature's code actually implements (the map legend is body-internal and
predates this feature; the villager strip is a later wave) — folding to the
same `[lesson]` header badge the stage-3+ default already uses, so the fold
reuses a designed state rather than inventing a new one. Shipped in
`computeRows` (`internal/tui/layout.go`): the lesson row reclaims its 2-row
budget first (once `bodyMin` is threatened), the guardian strip stays on if
that alone bought back enough body, and only folds too if it didn't —
`Model.wantsLessonRow`/`lessonBadgeVisible` (`internal/tui/views.go`) derive
the actual on-screen state (row / badge / absent) from the budget's
resulting `Lesson` field. In the narrow (< 112 cols) fallback, the row is
**carried** with identical stage defaults (`patterns/layout.md` ruling b) —
narrow terminals plausibly host new players, exactly where pushed delivery
matters most. Reconciliation note: unlike the guardian strip (confined to
`metatronView`, the only pane narrow gives a minibuffer to sit above), the
lesson row is chrome independent of any one pane, so its narrow carry
renders in `narrowView` directly, above whichever pane is active — no
fold-under-height-pressure exists in narrow this slice (no `computeRows`-
equivalent row-budget arithmetic of its own yet, the same gap the guardian
strip's narrow carry already has).

## Skin-token resolution (research.md R1, spec 052)

Every lesson string is authored with `{{skin.guardian.*}}` token literals,
resolved through `lessonSkinResolve` (`internal/tui/lessons.go`) at render
time (the row) and at population time (the help overlay's pull half,
boot-frozen). Spec 052's runtime token-resolution substrate (TASK-121) has
not yet merged to `main` as of this feature — `lessonSkinResolve` is a
bounded, package-local fallback table covering only the tokens the catalog
actually uses (`skin.guardian.epithet`/`.tab_label`), populated from the
PUBLISHED contract's §3 default-skin table
(`specs/052-skinnable-guardian/contracts/skin-contract.md`). The swap to
121's real resolver is a single-function change, documented at the seam;
either path satisfies FR-008/SC-005 (no raw `{{…}}` literal ever renders).

## Linear-stream / CLI projection (D1)

An `attach`/`tail` client (or any non-TUI observer) never loses lesson
content outright: every lesson's underlying trigger is itself a cataloged
event (a rejected `cog.tool_call`, `metatron.charter_observed`,
`metatron.order_placed`) already on the raw event log/chronicle feed
([[event-types]]) — the lesson row is a *TUI-side teaching projection* over
facts a linear client can already see, never a fact that exists nowhere
else. The lesson catalog text itself (the plain-language explanation) is
static content, reproducible in a non-TUI surface (e.g. `docs/player/`) with
no live-state dependency — no daemon change, no new event type, no model
call anywhere in this feature (FR-002).

## Control table

| control/region | states | data source | renderer | keys+mouse | introduced-by | skin-token |
|---|---|---|---|---|---|---|
| lesson row (active) | none · showing · badge | `lessonTriggers` projection (`internal/tui/lessons.go`) + per-user seen-state record (`internal/worlds/lessons.go`) | `lessonRowView`, `Model.wantsLessonRow` (`internal/tui/views.go`) | — (display-only) | reorient decision 5 / spec 055 | `skin.guardian.epithet`/`.tab_label` (per-entry, `lessonSkinResolve`) |
| UI-pointer + pull-path line | — | same trigger projection | `lessonRowView` (line 2, `lessonPullSuffix` appended) | — | reorient decision 5 / spec 055 | same |
| dismiss | active → dismissed | player action | `lessonTriggers.Dismiss`, wired in `handleGlobalKey` (`internal/tui/tui.go`) | `x` · — | reorient decision 5 / spec 055 | — |
| fold to header badge | shown · folded | stage default / fold pressure | `computeRows`, `Model.lessonBadgeVisible` (`internal/tui/layout.go`/`views.go`) | — | reorient decision 5 / spec 055 | — |
| pull via `?` | — | `helpLessons` seam, populated from `lessonCatalog` (`populateHelpLessons`) | `overlays/help.md`'s `helpLessonsLines` (existing seam) | `?` then `tab` to lessons section · — | spec 045 (seam) / reorient decision 5 (content) / spec 055 (population) | — |

**Parity rollout**: `x` (dismiss) has no mouse target — this is a Wave-4
surface, so the gap is recorded from birth rather than discovered later
(decision 8, formal doctrine in `patterns/keymap.md`, T024).

---
title: Panel — lesson row
class: panel
status: specified
verified_against: c8906da39be3a5b861c2272af37db0a83dcded7a
---

# Panel: lesson row

The first-occurrence teaching chrome (reorientation decision 5): one
always-visible-by-default row above the guardian strip that surfaces exactly
one lesson at a time, points at the key/tab that teaches it, and never
crowds the composite. **Not built** — this page specifies it spec-before-
build for TASK-117.

## Mockup

```
 Standing orders let you tell the {{skin.guardian.epithet}} what to watch for.
 → press 3, then look for 👁 rows           (? for more · x dismiss)
```

Two lines, **no border** (the 2-row budget `patterns/layout.md` allocates
has no room for a bordered box's own top/bottom rule):

1. **Lesson text** — plain language, one concept, the lesson's full sentence.
2. **UI-pointer + pull-path suffix** — an arrow phrase naming the key/tab
   that demonstrates the concept (`→ press 3, then look for 👁 rows`),
   right-padded with the pull-path suffix every lesson string carries:
   `(? for more · x dismiss)` — so a player who doesn't act on it immediately
   always knows exactly how to find it again and how to clear it now.

## Behavior

### One active, dwell, dismiss

- **One active lesson at a time** — a second trigger while one is showing
  queues rather than replaces it (see "Anti-spam" below).
- **Dwells until done or dismissed**: the active lesson stays up until
  either (a) the player performs the pointed-at action (the trigger's own
  "done" signal — e.g. placing a standing order dismisses the standing-order
  lesson), or (b) the player presses `x` to dismiss it outright.
- **UI-pointer field**: every lesson entry carries a pointer describing
  which key or tab demonstrates the concept, rendered as the `→ …` line
  above.
- **Pull-path suffix**: every lesson string, active or archived, ends with
  `(? for more)` (dismissed-and-gone lessons are readable again from
  `overlays/help.md`'s lessons section — the same `helpLesson{id, title,
  body}` seam that page's placeholder already reserves; this panel is the
  PUSH half of that seam, `overlays/help.md` is the PULL half).

### Anti-spam

- **Spacing**: a newly triggered lesson does not interrupt mid-dwell of the
  current one; it waits its turn.
- **Opportunity decay**: a queued-but-not-yet-shown trigger is not held
  indefinitely — if the active lesson's dwell period elapses without room to
  surface the next one soon after, that queued opportunity decays (is
  dropped) rather than surfacing stale content much later, disconnected from
  the moment that earned it. A lesson is a *timely* nudge, not a to-do list.
- **Per-user seen state (D8)**: each lesson id is recorded once shown, in a
  per-user record beside `~/.promptworld/unlocks.json`
  ([[curriculum-ladder]]'s precedent — load-tolerant, advisory-never-
  authority, atomic write) so a lesson never repeats for a player who has
  already seen it, across worlds and restarts.

### Trigger taxonomy

Two lesson tiers, both feeding the same one-active/dwell/anti-spam
machinery:

- **Mechanics lessons** — first-occurrence UI mechanics (e.g. "the map
  legend grows with what's in view").
- **Prompting-verb lessons** (the pivot's teaching core) — first-occurrence
  moments in the *player's own prompting practice*:
  - first rejected tool call (a `cog.tool_call` verdict other than landed,
    turn-scoped) — teaches that the guardian's grant is real and refusals
    are informative, not broken.
  - first custom charter observed (`metatron.charter_observed{default:
    false}`) — teaches that editing `charter.md` and returning changes the
    guardian's voice.
  - first fuzzy order (`metatron.order_placed{fuzzy: true}`) — teaches that
    a vaguely-worded standing order still binds, with its fuzziness marked
    honestly.

### Stage defaults

Per `patterns/stage-defaults.md` (the authority — this page never restates
the table, only cites it): **on** at stages 1–2; **badge + overlay-only**
at stage 3+ and pre-ladder. Folded state: a `[lesson]` header badge; content
remains reachable via `?` exactly as the fold-order case below.

### Fold behavior (widescreen height pressure and narrow)

Per `patterns/layout.md`'s ruling (a), this row folds third (after the map
legend and the villager strip) when body rows would drop below `bodyMin =
10` — folding to the same `[lesson]` header badge the stage-3+ default
already uses, so the fold reuses a designed state rather than inventing a
new one. In the narrow (< 112 cols) fallback, the row is **carried** with
identical stage defaults (`patterns/layout.md` ruling b) — narrow terminals
plausibly host new players, exactly where pushed delivery matters most.

## Linear-stream / CLI projection (D1)

An `attach`/`tail` client (or any non-TUI observer) never loses lesson
content outright: every lesson's underlying trigger is itself a cataloged
event (a rejected `cog.tool_call`, `metatron.charter_observed`,
`metatron.order_placed`) already on the raw event log/chronicle feed
([[event-types]]) — the lesson row is a *TUI-side teaching projection* over
facts a linear client can already see, never a fact that exists nowhere
else. The lesson catalog text itself (the plain-language explanation) is
static content, reproducible in a non-TUI surface (e.g. `docs/player/`) with
no live-state dependency.

## Control table

| control/region | states | data source | renderer | keys+mouse | introduced-by | skin-token |
|---|---|---|---|---|---|---|
| lesson row (active) | none · showing · dwelling | client-side lesson-trigger projection (mechanics + prompting-verb triggers) + per-user seen-state record | `unbuilt (wave 4)` | — (display-only) | reorient decision 5 | — |
| UI-pointer + pull-path line | — | same trigger projection | `unbuilt (wave 4)` | — | reorient decision 5 | — |
| dismiss | active → dismissed | player action | `unbuilt (wave 4)` | `x` · — | reorient decision 5 | — |
| fold to header badge | shown · folded | stage default / fold pressure | `unbuilt (wave 4)`, `patterns/layout.md` | — | reorient decision 5 | — |
| pull via `?` | — | `helpLessons` seam | `overlays/help.md`'s `helpLessonsLines` (existing seam) | `?` then `tab` to lessons section · — | spec 045 (seam) / reorient decision 5 (content) | — |

**Parity rollout**: `x` (dismiss) has no mouse target — this is a Wave-4
surface, so the gap is recorded from birth rather than discovered later
(decision 8, formal doctrine in `patterns/keymap.md`, T024).

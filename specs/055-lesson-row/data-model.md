# Data Model: First-occurrence lessons projection (spec 055)

Three entities, all client-side. No daemon, IPC, or event-schema surface.

## lessonEntry (static catalog record, `internal/tui/lessons.go`)

| Field | Type | Notes |
|---|---|---|
| `id` | string | stable, kebab-case (e.g. `first-suppression`); the seen-record key and the `helpLesson.id` |
| `title` | string | short label for the overlay's lessons section |
| `body` | string | overlay body text; skin-tokened |
| `text` | string | row line 1 — one plain-language sentence; skin-tokened |
| `pointer` | string | row line 2 arrow phrase (`→ press 3, then look for 👁 rows`); skin-tokened |
| `tier` | enum | `mechanics` \| `prompting` |
| `trigger` | predicate `func(store.Event) bool` | matches type + payload fields (research.md R3 table) |
| `done` | predicate `func(store.Event) bool`, optional | clears the active lesson as "done"; nil ⇒ dismiss via `x` only |

Invariants:
- Every entry's rendered line 2 ends with the pull-path suffix `(? for more · x dismiss)` —
  appended by the renderer, not stored per-entry (single source for the suffix).
- The catalog is append-only at runtime (a fixed package-level slice); `helpLessons` is
  populated from it 1:1 at init (id-for-id, mechanically tested — SC-002).
- Minimum population: the 8 entries of research.md R3 (5 mechanics + 3 prompting).

## LessonsSeen (per-user record, `internal/worlds/lessons.go`)

File: `~/.promptworld/lessons-seen.json` (sibling of `unlocks.json`, same home-dir
resolution).

```json
{
  "version": 1,
  "seen": {
    "first-suppression": { "first_shown": "2026-07-25T17:40:00Z", "world": "world-01" }
  }
}
```

| Field | Type | Notes |
|---|---|---|
| `version` | int | 1; unknown versions load as empty (tolerant) |
| `seen` | map[lessonID]seenEntry | presence = never fire again |
| `seenEntry.first_shown` | RFC3339 string | provenance only, never read for logic |
| `seenEntry.world` | string | provenance only |

Semantics (unlocks.go precedent, verbatim discipline):
- **Load-tolerant**: missing file, unreadable file, malformed JSON, unknown version →
  empty record + normal boot. Never an error surfaced to the player.
- **Advisory, never authority**: gameplay never gates on this file; worst failure mode
  is a repeated lesson.
- **Atomic write**: temp file + rename on every upsert; write failure is swallowed
  (logged at debug level at most) — the in-memory set still prevents repeats this
  session.
- **Per-user, cross-world**: no world name in the path; the map key is the lesson id
  alone.

State transitions: `absent → seen` only (no unsee; file deletion is the reset).

## lessonRow (client view state, fields on the TUI model)

| Field | Type | Notes |
|---|---|---|
| `active` | *lessonEntry + shownAt tick | nil = row empty (renders nothing / badge) |
| `queue` | bounded FIFO of {entry, decayAt} | pending first-occurrences awaiting the row |
| `seen` | in-memory set (loaded LessonsSeen ∪ session marks) | checked before enqueue |

Transitions:

```
event arrives ─┬─ trigger matches ∧ id ∉ seen ∧ no active ──▶ active (mark seen)
               ├─ trigger matches ∧ id ∉ seen ∧ active ─────▶ queue (decayAt = now+lessonQueueDecay)
               └─ else ──────────────────────────────────────▶ ignored
active ─┬─ done-signal event ──▶ cleared (spacing timer starts)
        └─ `x` pressed ────────▶ cleared (spacing timer starts)
spacing elapsed ∧ queue head not decayed ──▶ head becomes active (mark seen)
queue head decayAt passed ─────────────────▶ dropped silently
```

Marking seen happens when a lesson SURFACES (becomes active), not when queued — a
decayed queue entry has not been seen and may fire on a later first occurrence.

Rendering states: `none` (stage 1–2, nothing active: row absent — chrome gives the
rows back to the body) · `showing` (two-line row) · `badge` (`[lesson]` header badge:
stage 3+/pre-ladder default, or folded under height pressure).

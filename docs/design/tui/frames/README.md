# Generated frames — the design matrix

**Everything in this directory except this file is generated. Never hand-edit a `.txt`
frame.** An edit here is not a design change; it is a lie about what the client renders,
and the next `--dump` silently erases it. To change a frame, change `internal/tui` and
regenerate.

## What these are

One plain-text file per (fixture, state, size) combination, each containing exactly the
string `internal/tui`'s `View()` hands Bubble Tea at that size — the characters that
reach a real terminal, nothing added, nothing interpreted. They exist so that looking at
a screen costs a `grep` instead of a compile-and-run cycle, and so that a UI change can
arrive in review as a before/after of real frames rather than a prose claim about one.

The prose mockups in `../pages/` and `../panels/` remain the design *authority* — what a
surface is supposed to be. These frames are the *evidence* — what it currently is. When
the two disagree, that is a finding, and `../pages/home.md`'s "Reconciliation correction"
section is what it looks like when nobody notices for a while.

## Regenerating

From the repository root:

```
promptworld frames --dump              # rewrites this directory
go run ./cmd/promptworld frames --dump # …without installing
```

The dump is idempotent: regenerating without touching `internal/tui` leaves a clean
working tree. Files for combinations that no longer exist (a renamed state, a dropped
size) are pruned, so this directory is always exactly the current generation and never
the union of every generation ever run. This `README.md` is never touched.

Related commands:

```
promptworld frames --list                                   # fixtures, states, sizes
promptworld frames --fixture mid-game --state home --size 160x50
promptworld frames --fixture mid-game --state home --ansi   # keep the color escapes
promptworld frames --fixture scenario --interactive         # the real client, canned world
```

## File names

```
<fixture>__<state>__<width>x<height>.txt
```

Double underscores separate the three fields because single hyphens already occur
*inside* two of them (`mid-game`, `help-walkthrough`).

## The matrix

**Fixtures** (`internal/tui/fixtures.go`) — canned worlds built in process as Go values,
never by shelling out to `promptworld new`, so they are identical on every machine and
independent of which curriculum stages the operator has earned:

| fixture | what it exercises |
| --- | --- |
| `empty` | cold start: fresh stage-1 world, no events, no daemon — the disconnected header and every pane's empty state |
| `mid-game` | the dense case: stage-2 ambient world, deep chronicle backlog, awake/asleep/dead roster, all three teaching chrome rows |
| `scenario` | the exercise case: a stage-1 `--scenario` world, so the exercise dock tab and the lesson row render |

**States** — `internal/tui/design.go`'s `States()` is the single registry, consumed by
this matrix *and* by `internal/tui/render_test.go`'s exact-height sweep, so the two can
never disagree about what states exist or what one of them means.

**Sizes** — four, deliberately, not a width sweep: the matrix multiplies out over every
size added, so each one has to earn its place by straddling a real layout boundary.

| size | why |
| --- | --- |
| `80x30` | below the widescreen breakpoint (112) — the narrow single-pane fallback |
| `112x30` | exactly at the breakpoint, odd column remainder (the map takes the leftover column) |
| `113x30` | one column wider: the even 50/50 map ‖ dock split |
| `160x50` | roomy and deep — no chrome row folds, so everything is visible at once |

### Narrow frames are shorter than their terminal, on purpose

A `*__80x30.txt` file has fewer than 30 lines and that is **correct**. The exact-height
invariant (`View()` returns exactly `height` lines) holds only at or above the widescreen
breakpoint; the narrow fallback has no fold arithmetic at all — see
[`../patterns/layout.md`](../patterns/layout.md) ruling b. Narrow frames are
content-height. Do not "fix" the renderer to make the matrix uniform, and do not write a
height assertion over this directory that assumes otherwise.

## Determinism

A frame is byte-identical on every machine and every run, which is what makes a diff of
this directory meaningful. Four things are pinned to get there
(`internal/tui/fixtures.go`):

1. the world is a manifest value — terrain regenerates from `(seed, size, terrain_gen)`
   with no disk read;
2. the per-user records (lessons seen, stage unlocks) are **canned** rather than read
   from the operator's home directory — without this the same fixture renders differently
   for two people;
3. the wall clock is frozen;
4. the status snapshot is canned, so the header's tick, day, time and speed are fixed
   rather than polled from a running daemon.

Colors are suppressed by default (the `termenv` profile is forced to ASCII), so a frame
contains no escape bytes and diffs as text. `--ansi` opts back in for live eyeballing;
those frames are for eyes, not for this directory.

## Gate status

This directory is **exempt** from `scripts/check-tui-design.mjs` (its `GENERATED_DIRS`).
Every check that script runs asserts a property of an *authored* design page — it
declares a `class` matching its directory, it pins the commit a human verified it
against, `anatomy.md` maps it. Generated output has none of those and should not: its
authority is the generator. See `specs/047-tui-design-reference-v2/contracts/check-script.md`.

Nothing yet fails a build when these frames go stale — wiring a regenerate-and-diff gate
is a posture change deserving its own review, and is explicitly out of scope for spec 112
(which is its prerequisite).

## Provenance

Spec: `specs/112-tui-frame-harness/` · board task: TASK-187.
Generator: `cmd/promptworld/frames.go` over `internal/tui/design.go` and
`internal/tui/fixtures.go`.

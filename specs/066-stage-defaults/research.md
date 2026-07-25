# Research: Stage-shaped TUI layout defaults (spec 066)

## R1 — How the code table stays true to the authority page

**Decision**: carry the defaults as a Go table in `internal/tui/stagedefaults.go`
and add a sweep test that PARSES the markdown table in
`docs/design/tui/patterns/stage-defaults.md` and asserts cell-for-cell parity
with the code table (surface × {stage-1..4, pre-ladder, narrow}).

**Rationale**: this repo already anchors code to docs mechanically — the
`TestCatalogSweep` precedent (`internal/tui/digest_test.go`) parses
`docs/wiki/event-types.md` backticks and fails when code and doc drift. The
same pattern makes the authority page load-bearing rather than decorative:
changing a default without amending the page breaks the build, satisfying
FR-001's "no value hardcoded anywhere the table does not govern".

**Alternatives considered**: (a) generate the code table FROM the page at build
time — rejected: adds a codegen step for a ~10-row table and hides the value in
generated code; (b) prose-only discipline — rejected: exactly the drift the
gates-over-assertions principle exists to prevent.

## R2 — Where the starting set enters the render pipeline

**Decision**: resolve the starting visible set once per (boot, stage-change,
explicit toggle) into the model state the existing seams already read:
`rowBudget` inputs in `layout.go` (lesson row / strips), tab-presence checks in
`views.go`, and the help-overlay section variant in `help.go`. The fold
pipeline itself is untouched — it receives a starting set and folds under
height pressure exactly as today (`patterns/layout.md` ruling a).

**Rationale**: `layout.go` already models "wants" (`lessonWant`, strip rows)
before applying `bodyMin` pressure; stage defaults are just a different
starting want. No new pipeline stage means SC-004 (fold tests unmodified)
holds by construction.

**Alternatives considered**: a separate stage-aware layout pass — rejected:
duplicates fold logic and risks a second fold order.

## R3 — Pre-ladder byte-identity

**Decision**: the pre-ladder posture is the UNION column of the code table, and
the resolution function short-circuits: `Stage == ""` (and any unrecognized
value) returns the union set that matches today's unconditional defaults. The
golden-frame test renders a pre-ladder world before/after and diffs bytes.

**Rationale**: SC-002 demands byte-identity; making pre-ladder the identity
element of the resolution (rather than a stage-shaped case) makes the
regression test trivial and the fail-open posture (unrecognized stage) free.

**Alternatives considered**: treating unrecognized stages as stage-1 —
rejected: fail-closed hides surfaces from operators of newer/older worlds.

## R4 — Live stage change and explicit toggles

**Decision**: the TUI already receives the stage on its status snapshot
(`consoleStage` plumbing, `tui.go`; `currentStage`, `views.go:194`). On a
stage-id change between snapshots, re-resolve defaults; keep a small
per-surface override map recording explicit in-session player toggles, which
re-resolution never overwrites. Overrides are session-only (never persisted) —
matches the spec's Assumptions.

**Rationale**: snapshot-diffing is how the TUI already notices world changes;
`curriculum.stage_unlocked` events also flow to the client, but the snapshot is
the state of record and survives detach/reattach.

**Alternatives considered**: event-driven re-resolution off
`curriculum.stage_unlocked` — rejected as the trigger of record (a reattaching
client would miss it), though the event remains the ceremony/lesson trigger.

## R5 — First-occurrence announcements

**Decision**: newly default-on surfaces announce through the EXISTING lesson
machinery (`lessons.go` `lessonCatalog`, spec 055): first-occurrence semantics
already dedupe (exactly-once per world), so stage-driven appearance needs no
new announcement channel — only ensuring the surface's appearance routes
through the same first-occurrence path as its key-driven appearance.

**Rationale**: SC-005's "exactly once" is already the catalog's contract;
reusing it avoids a second dedupe mechanism.

## R6 — What is deliberately out of scope

- Chronicle presentation density ("traces / raw feed" board prose): not a row
  in the authority table ⇒ not stage-shaped in this feature (spec Assumptions).
- The villager strip's row applies only once TASK-129 lands; the resolution
  tolerates absent surfaces (a row with no surface is inert).
- No persistence of toggles; no capability-machinery reads/writes (FR-007).

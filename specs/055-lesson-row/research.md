# Research: First-occurrence lessons projection (spec 055)

Phase 0 — every unknown from Technical Context resolved. Grounding verified against
main at 4054e66 (all file:line references below re-checked at plan time).

## R1 — Skin-token resolution: consume spec 052's runtime, with a bounded fallback

**Decision**: lesson strings are authored with `{{skin.guardian.*}}` token literals
and resolved at render time through a single package-local seam,
`lessonSkinResolve(s string) string`, in `internal/tui/lessons.go`. At implementation
time: if TASK-121's runtime substrate (token table + resolution, spec 052 §2) has
merged to main, the seam delegates to it and contains no table of its own. If 121 has
not yet merged, the seam resolves only the tokens the catalog actually uses (at most
`skin.guardian.name/epithet/tab_label/working_noun`) from the **published default-skin
table** (`specs/052-skinnable-guardian/contracts/skin-contract.md` §3, normative), and
the swap to 121's runtime is a one-function change at rebase time.

**Rationale**: the runbook gates TASK-117 on 121's *contract* (published on main), not
its implementation. Bounding all skin coupling to one function keeps the inevitable
rebase over 121's merge to a single conflict site, and either path satisfies FR-008 +
SC-005 (default-skin values render; no raw `{{` literals — §2's resolution order ends
in the default table exactly as the fallback does).

**Alternatives considered**: hard-depending on 121's merge (serializes Lane 3 behind
Lane 1 for no contract-level reason — rejected); un-tokened strings with a TODO
(violates FR-008/AC #6 and D2 — rejected); implementing the full §2 resolution
ourselves (duplicates 121's in-flight work and guarantees a messy conflict — rejected).

## R2 — Seen-state persistence: `internal/worlds` sibling of unlocks.go

**Decision**: new file `internal/worlds/lessons.go` exposing `LessonsSeenPath()`,
`LoadLessonsSeen() *LessonsSeen`, and `MarkLessonSeen(id string)` (upsert +
atomic write), storing `~/.promptworld/lessons-seen.json`. Mirror
`internal/worlds/unlocks.go` exactly: resolve under the promptworld home dir
(`UnlocksPath()` precedent, unlocks.go:22), tolerate missing/corrupt files by
returning an empty record (loadUnlocksFrom, unlocks.go:95), write
temp-file-then-rename (writeUnlocks, unlocks.go:138).

**Rationale**: D8 names the unlocks.json precedent explicitly ("load-tolerant,
advisory-never-authority, atomic write"); same package, same discipline, mechanical
review.

**Alternatives considered**: folding ids into unlocks.json itself (couples two
unrelated records and complicates 046's file contract — rejected); per-world client
state (contradicts the decided per-user requirement — rejected).

## R3 — Trigger projection: the decisions.go ingest pattern, wired at the same seam

**Decision**: a `lessonTriggers` projection in `internal/tui/lessons.go` with an
`ingest(e store.Event)` method, called from the same event-arrival path that feeds
`decisionTraces.ingest` (decisions.go:150) in the poll loop. Trigger predicates match
on cataloged event types + payload fields:

| Lesson (tier) | Predicate |
|---|---|
| suppression (mechanics) | `cog.outcome` with suppressed outcome (decisions.go:358 vocabulary) |
| gru attack (mechanics) | `gru.attacked` (digest.go:896) |
| charge regen (mechanics) | `metatron.charge_regenerated` (digest.go:918) |
| order expiry (mechanics) | `metatron.order_expired` (digest.go:987) |
| first death (mechanics) | `agent.died` (digest.go:584) |
| rejected tool call (prompting) | `cog.tool_call` with verdict ≠ landed (decisions.go:180 vocabulary) |
| custom charter (prompting) | `metatron.charter_observed` with `default: false` (digest.go:997) |
| fuzzy order (prompting) | `metatron.order_placed` with `fuzzy: true` (digest.go:962) |

**Rationale**: the board task and design page both name the decision-trace view as the
pattern; it is proven, test-covered, and already receives every event exactly once.

**Alternatives considered**: subscribing in the daemon and emitting lesson events
(violates FR-002's client-side-only rule and D1 — rejected); scanning the chronicle
ring on render (misses events that scroll out; couples to digest internals — rejected).

## R4 — Row state machine: one-active, dwell, queue with decay

**Decision**: `lessonRow` state = none | active{entry, sinceTick} plus a small FIFO of
pending entries each carrying a decay deadline. Constants (v1, package-level, named):
`lessonQueueDecay` (queue residency before drop) and `lessonSpacing` (gap after a
lesson clears before the next surfaces), measured in the client's existing tick/poll
units. Active lessons have NO timeout — they dwell until done-signal or `x`
(design-page rule). Done-signals are per-catalog-entry predicates over the same event
stream (v1 ships them only where an unambiguous event exists — e.g. order-related
lessons clear on the matching `metatron.*` event); entries without a done-signal
dismiss only via `x`.

**Rationale**: matches the panel page's one-active/dwell/anti-spam contract verbatim
("dwells until done or dismissed"; "a lesson is a *timely* nudge, not a to-do list");
constants-over-config matches the spec's assumption.

**Alternatives considered**: auto-timeout for active lessons (contradicts the dwell
rule — rejected); unbounded queue (contradicts opportunity decay — rejected).

## R5 — Layout: chrome-row budget + fold order under today's foldable set

**Decision**: the lesson row joins the chrome budget in `internal/tui/layout.go` as a
2-row entry above the guardian strip (stripRows precedent, layout.go:22). Fold
behavior implements `patterns/layout.md` ruling (a)'s ORDER (map legend → villager
strip → lesson row → … guardian strip last) as it applies to the foldables that exist
today: the villager strip (TASK-129) and the map-legend fold do not exist yet, so in
this slice the lesson row folds BEFORE the guardian strip (i.e. first among current
foldables), which is exactly ruling (a)'s relative order restricted to the built set.
The folded form is the `[lesson]` header badge — the same state stage-3+ defaults use.
Narrow (< 112 cols) carries the row with identical defaults (ruling (b)).

**Rationale**: preserves the authoritative ordering without building other tasks'
rows; the design page itself frames fold-third as reusing the stage-3+ badge state, so
both folded paths share one implementation.

**Alternatives considered**: implementing placeholder fold slots for TASK-129's rows
(speculative scaffolding another task will rewrite — rejected).

## R6 — Stage defaults: read `Status.Stage` directly this slice

**Decision**: row visibility derives from the status stage already on the client
(`ipc/protocol.go:178`, `consoleStage` in tui.go:115): stages 1–2 → row on; stage 3+
or pre-ladder ("" stage) → `[lesson]` badge + overlay-only. Implemented as a small
`lessonRowDefault(stage string) bool` the stage-defaults machinery (TASK-128) can
absorb later — this slice deliberately does NOT build the shared machinery
(`patterns/stage-defaults.md` stays the authority; spec assumption records the seam).

**Rationale**: the spec's recorded assumption; avoids cross-task coupling with
TASK-128 while keeping one function for it to lift.

## R7 — Help-overlay pull half: populate the existing seam, no help.go changes

**Decision**: at client init, `helpLessons` (help.go:146) is populated from the same
catalog (`helpLesson{id, title, body}` — help.go:138 seam contract), with bodies
skin-resolved through R1's seam at population time (boot-frozen skin ⇒ boot-time
resolution is stable). `helpLessonsLines` (help.go:428) already renders entries vs
placeholder; no rendering change needed. SC-002's mechanical guarantee: a test asserts
`len(helpLessons) == len(lessonCatalog)` and id-for-id equality.

**Rationale**: the seam was built for exactly this (spec 045 US4); one catalog feeding
both surfaces is FR-007's "never two hand-maintained lists" made structural.

**Alternatives considered**: rendering the catalog live in the overlay (needless
coupling; the overlay's byte-identity classification expects static-per-skin content —
rejected).

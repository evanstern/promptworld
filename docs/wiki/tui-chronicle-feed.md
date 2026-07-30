---
name: tui-chronicle-feed
description: The chronicle pane's narrated/raw toggle and the digest grammar: a digestRegistry entry per cataloged event type turning payload into a readable feed line, color-role tints, whole-line alerts, and the inspect-mode detail pane. Guardian-domain digest entries split to [[tui-chronicle-feed-guardian-digests]]. Split from [[tui-client]]; read when adding an [[event-types]] entry or touching grammar.go/digest.go.
kind: component
sources:
  - internal/tui/grammar.go
  - internal/tui/digest.go
  - internal/tui/tui.go
verified_against: d0645811c9783d1248dc65ed0fcf0b37524dd8fd
---

# TUI chronicle feed and digest grammar

Split from [[tui-client]] (corpus-spec v2 size-budget split, summary-style);
covers chronicle rendering and the raw feed's digest grammar registry. [[tui-client]] keeps the map view, dock tabs, and input/help
overlay.

## Chronicle rendering and the digest grammar

The **chronicle** renders the narrated story from the replica's
snapshot-carried `State.Chronicle` ring ([[chronicle]]) or the raw feed (`r`
toggles; raw auto-falls back with no narrated entries; `a`/`t` cycle
agent/thread filters). Raw lines follow the **digest grammar** (spec
018/TASK-60; grammar.go + digest.go, pure functions emitting styled segments
— never ANSI): every cataloged event type has a `digestRegistry` entry
turning payload into a readable per-type summary; a feed line reads
`TICK HH:MM type summary` — natural phrases for narrative families
(`Ash foraged at (14,9)`, speech privileged as `Ash→Rowan "utterance"`),
compact `key=value` fields for telemetry families (`cog.*`, `clock.*`,
`daemon.*`). Columns align at solo width (tick right-aligned, type padded);
the narrow dock drops the tick and shortens type to its last segment.
The Type column renders every type RAW on every surface — solo column, dock
short form, detail pane, and the grammar-miss fallback alike. Spec 052
(FR-013) briefly aliased the then-frozen `metatron.*` namespace segment to
the active [[skin]]'s family label (TASK-121's interim shim,
`displayEventType`/`chronicleLine.DisplayType`); spec 094 shipped the real
rename — persisted types are `guardian.*` natively — and deleted the shim,
so the whole feed and inspector is honest, unskinned wire vocabulary
(inspector-class visibility, FR-020).
Families carry color-role tints, key tokens (names, speech, amounts, causes)
emphasis, and six high-salience types (`agent.died`, `gru.attacked`,
`social.chest_taken`, `norm.violated`, spec 077's `stranger.took` — beside
`chest_taken` because theft is theft — and spec 083's
`sim.neglect_detected`, beside `agent.died` because a need neglected to the
edge of death is the same class of alarm with runway left; its digest row
carries the deterministic per-need wording "Ash is dangerously cold and has
done nothing about it (warmth 0)" / starving / exhausted,
[[executor-needs-survival]]) render whole-line alert. Since spec 038,
`agent.build_failed` ([[executor]], [[event-types]]) gets its own entry —
builder, emphasized goal and reason ("Ash's build_wall_stone failed — site no
longer buildable") — a failure at a glance without joining the whole-line
alert tier; a cancelled build previously read as finished, sharing
`agent.intent_done`'s plain "finished" line. Spec 096's generalization,
`agent.intent_failed`, gets the identical treatment for every non-build goal
("Ash's hunt failed — target gone"). `agent.recovery_stalled` (spec 064; catalog row TASK-140)
gets the same distinct-from-"finished" treatment for a needs-conditioned
recovery hold with no net gain across its stall window — "Ash's `warm_up`
stalled — `warmth` not recovering" (subject/name resolution per
`agent.build_failed`). Since spec 041, four more [[mental-maps]] event types get entries sharing a
first-fact-plus-count shape (a full list would flood the line; the detail
pane holds it verbatim): `agent.saw` ("Ash saw fire
at (x,y) (+N more)"), `social.place_told` ("Ash told Birch of fire at (x,y)
(+N more)"), `agent.map_corrected` ("Ash found fire at (x,y) gone (+N
more)"), and `guardian.place_revealed` ("Guardian revealed fire at (x,y) to
Ash (+N more)", guardian as subject, the nudge convention). Since spec 042,
three [[memory-retrieval]] event types get entries, the raw vector
elided (384 floats would drown the feed): `agent.memory_embedded`
("memory seq=N embedded dims=N model=…"), `agent.situation_embedded` (agent
plus rendered situation text as speech, then "dims=N model=…"),
and `cog.memory_divergence` (agent, mode, "overlap=N/N", "displaced=N",
"vectorless=N" — the rank-divergence telemetry the US2→US3 gate decision
reads). The guardian-domain families — run-outcome/morgue (spec 044),
curriculum (046), the report card (063), world-forking (076), and the
plan layer (084), plus the four guardian-miracle entries and their
gratis annotation — split into [[tui-chronicle-feed-guardian-digests]].
Unregistered
future types fall back to the old grammar's compact resolved-name JSON
(the agent-index field table — `agentIndexFields`/`agentIndexFieldRe`,
covering `agent`, `a`, `b`, `from`, `to`, `speaker`, `listener`, `subject`,
`owner`, `taker` — still drives that fallback and the inspector). A sweep
test (`digest_test.go`) fails if any type cataloged in [[event-types]] lacks
a digest. Pausing puts the visible chronicle in **inspect mode** — dormant
while look-cursor mode (spec 074-look-cursor, [[tui-map-view]]) borrows the
dock, since `chronicleVisible()` reads false for the borrow, as with another
tab selected: `j`/`k`/`g`/`G` select; the selected event's full detail
auto-shows in the always-on **detail pane** at the panel bottom — seq, tick,
type, the stored payload verbatim, pretty-printed with `// name` annotations
beside integer agent indices; `J`/`K` scroll oversized payloads
within the pane's row budget, and `⏎` is a reserved no-op, the documented
attachment point for future jump-off actions
(`docs/design/tui/patterns/chronicle-grammar.md`, `panels/chronicle.md`).

## Back to parent

[[tui-client]] links here for this feed; its Connections
section lists [[chronicle]] and [[event-types]] as this feed's data sources,
alongside [[mental-maps]] and [[memory-retrieval]] for the per-family digest
entries added since their specs. [[tui-chronicle-feed-guardian-digests]] is
this note's own split-off child, covering the guardian-domain families
([[morgue]], [[curriculum-ladder]], [[grounded-feedback]] among them).

Since spec 077 ([[event-types-scenario-incidents]]), seven more types get
entries with NO new family, tier, or channel: `sim.cold_snap` ("a
cold snap grips night N (until tT)") and `sim.forage_blighted` ("blight
struck the forage at (x,y) (+N more tiles)" — the first-fact-plus-count
shape) in the sim voice; the `stranger.*` namespace mapped onto the
gru/threat family voice (`familyByNamespace["stranger"]` — a second
nocturnal entity, not a new visual role): `stranger.arrived` ("a
stranger slipped in at (x,y)"), `stranger.moved` ("the stranger creeps to
(x,y)"), `stranger.took` (alert tier, "the stranger took N <kind> from the
stores at (x,y)"), `stranger.departed` ("the stranger was
gone by dawn of day N"); and `guardian.skills_observed` ("Guardian ran
under N skill file(s) <fingerprint>" — the charter observation's twin,
guardian family/skin-name subject).

## Spec 086 — payload-first naming and the generic subject fallback

Agent-bearing digests now read the payload ref's name first (`refSeg`/
`refName`, `internal/tui/digest.go`): a post-086 row renders with NO
replica lookup — proven by `TestCatalogSweep`'s `names = nil`
identical-output assertion over every agent-bearing fixture. The replica
`names` slice and the `resolvePayloadNames` regex rewriter survive as the
historic-row fallback layer (legacy rows carry `Name == ""`), shrunk but
never removed. `resolveSubject` gained a registry-miss generic pass: scan
the payload for `{"id":N,"name":…}` ref objects; exactly one distinct
in-roster id is the subject (live position, payload name), zero or several
stay unlocatable — the honest-hint doctrine detected structurally.
`world.migrated` stays hard-excluded. Hit rate: 79 registry-only vs 86
registry+fallback locatable fixture rows (`TestResolveSubjectHitRate`),
with `journal.entry_written`/`journal.entry_deleted` and
`faith.changed{villager_died}` pinned as newly locatable.

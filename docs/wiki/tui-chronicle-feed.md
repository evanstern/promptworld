---
name: tui-chronicle-feed
description: The chronicle pane's narrated/raw toggle and the digest grammar: a digestRegistry entry per cataloged event type turning payload into a readable feed line, family color-role tints, high-salience whole-line alerts, and the inspect-mode detail pane. Split from [[tui-client]]; read when adding an [[event-types]] catalog entry (TestCatalogSweep enforces coverage) or touching grammar.go/digest.go.
kind: component
sources:
  - internal/tui/grammar.go
  - internal/tui/digest.go
  - internal/tui/tui.go
verified_against: 4c66d240b2715706964f02cfd2396256c9957d8e
---

# TUI chronicle feed and digest grammar

Split from [[tui-client]] (corpus-spec v2 size-budget split, summary-style):
this note covers the chronicle pane's rendering and the raw feed's digest
grammar registry. See [[tui-client]] for the map view, dock tabs, and
input/help overlay.

## Chronicle rendering and the digest grammar

The **chronicle** renders the narrated story from the replica's
snapshot-carried `State.Chronicle` ring ([[chronicle]]) or the raw feed (`r`
toggles; raw is the automatic fallback with no narrated entries; `a`/`t`
cycle agent/thread filters). Raw lines follow the **digest grammar** (spec
018/TASK-60; grammar.go + digest.go, pure functions emitting styled segments
— never ANSI): every cataloged event type has a `digestRegistry` entry
turning its payload into a readable per-type summary, so a feed line reads
`TICK HH:MM type summary` — natural phrases for narrative families
(`Ash foraged at (14,9)`, speech privileged as `Ash→Rowan "utterance"`),
compact `key=value` fields for the telemetry families (`cog.*`, `clock.*`,
`daemon.*`). Columns align at solo width (tick right-aligned, type padded);
the narrow dock drops the tick and shortens the type to its last segment.
Since spec 052 (FR-013), the solo Type COLUMN specifically — never the dock's
short form, never the detail pane, never the grammar-miss raw fallback —
aliases the FROZEN `metatron.*` namespace segment to the active skin's family
label (`displayEventType`/`chronicleLine.DisplayType`, grammar.go:
`metatron.nudged` renders `guardian.nudged` by default, `raven.nudged` under
the example skin); `curriculum.*` and every other family render raw by design
(inspector-class visibility, FR-020) — the Type column is the ONE display
surface this note's `[[skin]]` link touches, everywhere else in the raw feed
and inspector stays honest, unskinned wire vocabulary.
Families carry color-role tints, key tokens (names, speech, amounts, causes)
carry emphasis, and five high-salience types (`agent.died`, `gru.attacked`,
`social.chest_taken`, `norm.violated`, and — spec 077's only addition to
the tier — `stranger.took`, beside `chest_taken` because theft is theft)
render whole-line alert. Since spec 038,
`agent.build_failed` ([[executor]], [[event-types]]) gets its own registry
entry — builder, emphasized goal, and emphasized reason ("Ash's build_wall_stone
failed — site no longer buildable") — reading as a failure at a glance without
promoting to the whole-line alert tier; a cancelled build was previously
indistinguishable from a finished one because it shared `agent.intent_done`'s
plain "finished" line. `agent.recovery_stalled` (spec 064, catalog row added
TASK-140) gets the same distinct-from-"finished" treatment for a
needs-conditioned recovery hold that showed no net gain across its stall
window — "Ash's `warm_up` stalled — `warmth` not recovering" (subject and
name resolution the `agent.build_failed` precedent). Since spec 041, four more [[mental-maps]] event types get registry entries,
all sharing a first-fact-plus-count shape (a full fact list would flood the
line; the detail pane holds the payload verbatim): `agent.saw` ("Ash saw fire
at (x,y) (+N more)"), `social.place_told` ("Ash told Birch of fire at (x,y)
(+N more)"), `agent.map_corrected` ("Ash found fire at (x,y) gone (+N
more)"), and `metatron.place_revealed` ("Guardian revealed fire at (x,y) to
Ash (+N more)", the guardian as subject, the nudge convention). Since spec 042,
three [[memory-retrieval]] event types get registry entries with the raw
vector deliberately elided (384 floats would drown the feed): `agent.memory_embedded`
("memory seq=N embedded dims=N model=…"), `agent.situation_embedded` (the
agent plus its rendered situation text as speech, then "dims=N model=…"),
and `cog.memory_divergence` (agent, mode, "overlap=N/N", "displaced=N",
"vectorless=N" — the recorded rank-divergence telemetry the US2→US3 gate
decision reads). Since spec 044, three run-outcome/[[morgue]] types get
registry entries, and `familyByNamespace` (grammar.go) maps their two new
namespaces onto existing family voices — `run` speaks in the world-lifecycle
voice, `morgue` in the chronicle's narrated-prose voice: `run.ended` ("the
run ended · N dead · final cause <cause>" — the summary a postmortem reader
wants on the feed line; the full ledger stays in the payload/detail pane),
`morgue.epilogue` ("epilogue for <name>: <text>", 80-rune truncation like
`chronicle.entry`; agent −1 renders as "the run" — the run-end epilogue),
and `metatron.charter_observed` ("Guardian ran under charter <fingerprint>
(default|player-authored)" — the charter-revision stamp the morgue aligns
deaths against). Since spec 046, two [[curriculum-ladder]] types get registry
entries, and `familyByNamespace` maps the new `curriculum` namespace onto the
existing guardian family voice — the ladder is the guardian's domain, not a
distinct visual role: `curriculum.exercise_passed` ("the <exercise> exercise
was passed (<stage>)") and `curriculum.stage_unlocked` ("The guardian's watcher
earned <stage name> (proven by <exercise>)", the display name resolved
through `skin.StageName` like the CLI's stage line). Since spec 063,
`guardian.report_card` ([[grounded-feedback]]) gets its own entry too — a
new `guardian` namespace joins `familyByNamespace`, mapped onto the SAME
guardian family voice the FROZEN `metatron.*` namespace uses (the digest
line renders the skin's report-card label, the charter fingerprint, and the
note's own text truncated to 80 runes, the `morgue.epilogue` truncation
manner). The four
[[guardian-miracles]] types render in the guardian family voice, with a
trailing emphasized `(forced)` annotation (`gratisMark`) whenever the
payload's gratis flag waived the charge — an operator force is never
indistinguishable from a charge-priced miracle in the feed. Unregistered
future types fall back to the compact resolved-name JSON of the old grammar
(the agent-index field table — `agentIndexFields`/`agentIndexFieldRe`,
covering `agent`, `a`, `b`, `from`, `to`, `speaker`, `listener`, `subject`,
`owner`, `taker` — still drives that fallback and the inspector). A sweep
test (`digest_test.go`) fails if any type cataloged in [[event-types]] lacks
a digest. Pausing puts the visible chronicle into **inspect mode**:
`j`/`k`/`g`/`G` select, and the selected event's full detail shows
automatically in an always-on **detail pane** at the panel bottom — seq,
tick, type, the stored payload verbatim, pretty-printed with `// name`
annotations beside integer agent indices; `J`/`K` scroll oversized payloads
within the pane's row budget, and `⏎` is a reserved no-op documented as the
attachment point for future jump-off actions
(`docs/design/tui/patterns/chronicle-grammar.md`, `panels/chronicle.md`).

## Back to parent

[[tui-client]] links here for the chronicle/digest feed; that note's own
Connections section lists [[chronicle]] and [[event-types]] as this feed's
underlying data sources, alongside [[mental-maps]], [[memory-retrieval]],
[[morgue]], [[curriculum-ladder]], and [[grounded-feedback]] for the
per-family digest entries added since their respective specs.

Since spec 077 ([[event-types-scenario-incidents]]), seven more types get
registry entries with NO new family, tier, or channel: `sim.cold_snap` ("a
cold snap grips night N (until tT)") and `sim.forage_blighted` ("blight
struck the forage at (x,y) (+N more tiles)" — the first-fact-plus-count
shape) in the sim voice; the `stranger.*` namespace mapped onto the
gru/threat family voice (`familyByNamespace["stranger"]` → the gru family —
a second nocturnal entity, not a new visual role): `stranger.arrived` ("a
stranger slipped in at (x,y)"), `stranger.moved` ("the stranger creeps to
(x,y)"), `stranger.took` (the alert-tier theft line, "the stranger took N
<kind> from the stores at (x,y)"), `stranger.departed` ("the stranger was
gone by dawn of day N"); and `metatron.skills_observed` ("Guardian ran
under N skill file(s) <fingerprint>" — the charter observation's twin,
guardian family/skin-name subject).


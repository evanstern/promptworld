---
name: tui-client
description: The Bubble Tea full-screen client — a widescreen map+dock composite (with minibuffer and narrow single-pane fallback) over a live world replica maintained by log shipping (state snapshot + event subscription through the shared reducer)
kind: component
sources:
  - internal/tui/tui.go
  - internal/tui/views.go
  - internal/tui/layout.go
  - internal/tui/grammar.go
  - internal/tui/digest.go
  - internal/tui/decisions.go
  - internal/tui/help.go
verified_against: 824932c630a9216dc761f78baa903cd07e5b9493
---

# TUI client

`internal/tui` is the attachable full-screen client (`promptworld ui <dir>`), built on
Bubble Tea + Lipgloss. Its core idea: the map renders from a **live replica** of
`sim.State` that the client maintains by log shipping — fetch the state snapshot, then
apply every pushed event through the exact `Apply` reducer the daemon runs. The TUI is
a read replica of the world.

## How it works

`Model` holds the world handle, an `ipc.Client`, the replica, the latest polled
`StatusData`, and a chronicle ring (`chronicleCap = 500` events). All protocol calls
run inside `tea.Cmd`s so the UI never blocks on the socket.

Connection (`connect`): dial → `FetchState` (state JSON + the `last_seq` it reflects)
→ unmarshal into a fresh `sim.NewState(seed)` → `Subscribe(since: last_seq)` — the
replica starts gapless by construction. `listen` delivers one push per invocation and
`Update` re-arms it. `applyEvent` skips seqs already folded into the snapshot, applies
the rest to the replica, bumps its tick, and appends to the chronicle ring.

**Postmortem posture** (spec 044): once the run is over, the header's clock
state renders a bold-red `ENDED` token (`styleEnded` — a finality register
`PAUSED`'s amber deliberately doesn't carry) that outranks `PAUSED`
regardless of the clock state the run end landed under. The predicate is
`Model.runEnded()` (tui.go), dual-source by necessity: the replica's
`State.Ended` covers clients attaching after the fact (the snapshot path
never replays folded events), while the pushed `run.ended` event (folded by
`applyEvent`) and the 1-second status poll (`StatusData.Clock.Ended`) cover
the live transition without a reconnect. The same predicate makes the clock
keys (space, `[`, `]`) inert client-side — the daemon's refusal error would
otherwise read as a disconnect — and swaps the footer's pause/resume hint
for `run ended (read-only)` in every mode; all reading surfaces stay fully
functional ([[morgue]]).

**Governed speed** (`headerView` in `views.go`, spec 028 US4): the header's
speed segment renders the EFFECTIVE speed as the world's speed, and — only
while `StatusData.Clock.RequestedSpeed` is set and differs from `Speed` (the
governor has shed at least one notch) — gains a plain-language suffix via
`governedSpeedSuffix`: `"asked 32x — 3 minds in flight, debt 140%"`. An
ungoverned world (`RequestedSpeed` empty) renders byte-identically to
pre-028. Since spec 034, the header also gains a red `[llm: <provider>
<kind>]` badge (the `[degraded]` badge's pattern) whenever any provider
carries an active health condition — `firstLLMCondition` reports the first
name-sorted affected provider; no condition active renders no badge
([[llm-provider-health]]). Since spec 037 (US1, FR-005), the header gains a
further warn-styled `[suppressed: class, class]` badge whenever ≥1 watched
class in the polled `StatusData.Horizon` is currently suppressed —
`suppressedHorizonClasses` filters the wire slice (already in
`WatchedClasses` order) with no client-side re-derivation; a world with no
horizon (no-LLM, or nothing suppressed) shows no badge. `debtPercent` (`digest.go`) is the one shared arithmetic behind both
this suffix and the digest lines below: the measured debt expressed as a
whole percent of `cognition.ShedThreshold`, rounded to the nearest percent.
The raw chronicle feed's digest grammar gains two entries for the same
feature: `clock.governor_shed`/`clock.governor_recovered` each render as
`"governor shed/recovered <from>→<to> debt=N% jobs=N"`, in the terse
`clock.degraded`-line style (the `requested` payload field is omitted here —
the from→to transition already carries the notch delta).

Resilience: errors become `disconnectedMsg` → the header shows the failure and a
2-second retry loop re-dials; a `dropped` push (subscriber overflow) tears the client
down and reconnects from a fresh state snapshot, because the replica may have missed
events. One exception is fatal (TASK-19): `ipc.ErrReplyTooLarge` (a reply over the
protocol's 64 MiB ceiling — reconnecting cannot shrink the state) quits instead of
retrying, rendering the reason in the final view and exposing it via
`Model.FatalErr()`, which `cmdUI` turns into a non-zero exit. A 1-second poll refreshes the clock/status line (quiet ticks produce no
events, so the replica's tick alone would lag).

Layout (TASK-34; design reference in `docs/design/tui/` — entry points
`INDEX.md`/`anatomy.md`; since TASK-123's v2 taxonomy the dock's per-tab
content is split across `panels/guardian.md` (fiction layer), `panels/
systems.md` (telemetry), and `panels/villagers.md`, with `panels/dock.md`
itself covering only the tab-container chrome): at ≥112 columns the
client renders the **widescreen composite** — the map on the left and a tabbed
**dock** on the right in a 50/50 split (`computeColumns` in layout.go; the map's
viewport derives from the column budget via `mapViewportTiles`), a one-line
borderless **guardian strip** (spec 050, reorient decision 7: charge-bank
glyphs + `(N/cap)`, a `next +1 @ <time>` regen forecast derived from
`sim.MetatronChargeRegenTicks`, and the replica's standing-order count —
`guardianStripView`, each segment degrading to absence rather than a
misleading zero), a one-line
**Metatron minibuffer** above the footer, and per-mode footer hints. The strip
is the last chrome to fold (`rowBudget.Strip`, `computeRows` keeps it while
body ≥ 10 rows); folded, its content relocates into the minibuffer's dormant
placeholder line instead of hiding. Below 112
columns it falls back to the original single-pane UI (header + tab bar + one
active pane), unchanged except that the guardian pane carries the same strip
above its minibuffer. `View` output is exactly terminal-height in every mode
(every panel body is clipped to its row budget — `clipContent`), and resizes
re-clamp pan/selection state (`clampGeometry`).

Regions: the **map** is a camera window over the generated terrain from
`Model.gameMap` (regenerated locally via `world.Map()`,
[[worldmap-generation]]): water ~, wood ♠, forage ", rock outcrops ^, and dens
ᴥ glyphs, plus dynamic overlay state read off the replica (never part of the
static tile) — a quarried-out rock outcrop renders as a faint `,` ahead of the
static terrain check — with the replica's agents on top (by initial,
lowercase asleep, † dead) plus built structures: fires render lit ▲ while the
current tick is before the structure's `FuelUntil` and fall back to a faint,
hollow cold glyph △ once fuel runs out, shelters ⌂, ovens ▣, chests ☐ (spec
013 US3), and the [[gru]] as a red G while it is abroad; ground piles (spec
013 US2, `Model.replica.Piles`) render as a dedicated overlay `%`, layered
like structures rather than folded into them so a coincidental tile overlap
loses neither glyph's priority silently. Since spec 032, a wall structure
(`wall_plank`/`wall_stone`) renders as a solid barrier glyph — `▤` plank, `▩`
stone — dim (`styleWallDamaged`) whenever its `HP` is below `sim.WallMaxHP`,
the same faded-glyph treatment as a burnt-out fire, so a wall under
demolition reads at a glance; a path structure renders at TERRAIN level
(below agents/structures/piles, its own `paths` set rather than the
structures map) as a warm-tan `·` distinct from plain grass's dim `·`, so an
agent or a dropped pile standing on a path tile still shows through. Since
spec 044 (US4), a `grave` structure renders as `✝` in faint gray
(`styleGrave`, the cold-fire precedent for spent/inert glyphs), recorded
both in the structures map and in a dedicated `graves` set that the agents
loop consults for one deliberate priority carve-out: a dead agent standing
on a tile that also holds a grave renders the grave glyph instead of the
plain dead marker — the body becomes the grave — because every post-044
death places its grave at the dead agent's own frozen tile, so the usual
agent-over-structure priority would otherwise permanently hide the glyph. A
graveless dead agent (pre-044 replay/history) still renders the plain `†`.
The camera follows the living agents' centroid, arrow keys pan, `c` recenters.
Since spec 049 (TASK-124, reorient D3) the camera gains one computed writer:
**jump-to-source** — in inspect mode, `⏎` (or a mouse click on a chronicle
list line; `tea.WithMouseCellMotion` is enabled in `cmdUI` and `handleMouse`
binds ONLY this control, everything else falling through inert) resolves the
selected event's subject (`resolveSubject` + a per-type `subjectRegistry` in
digest.go: primary actor's live replica position if alive, else explicit
payload coordinates, else unlocatable; `world.migrated` is never decoded) and
sets the pan via `centerCameraOn` (`pan = target − wandererCentroid()`), so a
jump IS a pan — same clamps, same `c` recovery. The detail pane's bottom row
is now a permanent actions bar (`detailActions`, exactly one action per
event): the jump affordance `⏎ jump to <name> (x,y)` or the honest
`no location for this event`; a catalog sweep asserts jump-or-hint totality.
In the narrow fallback a successful jump lands on the map pane with the
paused selection preserved. Click hit-testing reads a per-frame
`chronHitRegion` the inspect-list renderer records — running-clock,
out-of-region, help-open, and minibuffer-focused clicks are all no-ops.

Inspection (spec 013 T021/T026, SC-006): the map legend — its one designated
inspection surface, content grows the line rather than adding a second row —
appends, for whatever's currently in view, a stockpile-zone summary per pile
cluster and an owner+contents+fullness entry per chest. Piles in view are
grouped into **stockpile zones** by 4-neighbor Manhattan adjacency
(`pileZones`, a render-side-only flood fill — no zone state, matching
spec.md's "an observability grouping of adjacent piles, not a state entity");
each zone renders as `pile(x,y) contents` (single pile) or
`zone[n](x0,y0)-(x1,y1) contents` (multi-pile, bounding box + count), where
contents (`summarizePileContents`) is non-food resource counts plus a spear
count plus a `food Nr/Nc/Nm` batch total when any food is held. Each visible
chest renders as `chest(x,y) [Owner] contents n/48` (`describeChest`, owner
resolved through the same `agentName` helper the chronicle grammar uses,
contents via `summarizeInventoryContents`, capacity `sim.ChestCap`) — a
chest's `Store` is a plain counts inventory rather than dated batches,
because chests preserve food indefinitely (no rot deadlines to track).

The **dock** hosts three tabs — keys `2`/`3`/`4` select, the same key again
zooms the tab solo, `1`/`esc` return to the composite: **chronicle** (default;
see below), **metatron** (the angel transcript — replies stream here, or
badge the tab `metatron •` when it isn't visible; the pane header shows the charge
bank plus the spec-021 instruction/capability provenance summary — charter
default/custom, skill-file count when non-zero, and the granted-tool summary from
`Status.GrantedTools`, quiet for a full-grant default world — [[metatron]] —
joined, since spec 046, by a `stage: <skin name>` segment
(`consoleStageSummary` in tui.go) that appends `(charter locked to
<preset>)` while the stage-1 instruction lock binds, read off the polled
`Status.Stage`/`CharterLocked`/`CharterPreset` with no client-side
re-derivation; empty (header byte-identical) for a pre-ladder/ungated
world — [[curriculum-ladder]]; the
transcript itself gains a `👁 watch set`/`👁 watch released` line for a
placed/cancelled standing order and a `⏲` line for a landed pause/start/
adjust_speed meta tool call, alongside the existing `⚡` vision/omen line;
below the transcript, a `👁 standing orders (n)` block (spec 029,
`orderStatusLines`, [[metatron-orders]]) renders one compact row per order
from `Status.Orders` — id, a `~` fuzzy marker, origin, remaining game-day,
status, and condition — present only while orders stand; the
same pane renders the LLM provider table since spec 024 — `llmProviderLines`,
one row per provider with name, model, up/down glyph, queue, inflight/slots, a
contended marker, and spend share, plus an `(unattributed)` row for pre-024
months, followed by a `spend $X of $Y` wallet line — [[llm-orchestrator]]; a
provider carrying an active health condition (spec 034) gains an indented
continuation line rendering the condition's detail and remedy in the pane's
error style, immediately below its row — [[llm-provider-health]]; beside the
provider table, since spec 037 (US1, FR-006) the pane also gains a
`horizonLines` block, one dim "🜂 cognition horizon" header plus a
`horizonRow` per watched class from the polled horizon — a thinking class
renders a plain "<class> thinking at <speed>"; a suppressed class is
warn-styled with its remedy ("suppressed at <speed> — calibrate or slow
down" for an uncalibrated class, "… — slow down" for a calibrated one,
`horizonRemedy`) and carries the router's own verdict arithmetic verbatim as
a dim trailing detail — no raw enum ever reaches the screen. A trailing dim
"· skipped N" (the class's `SuppressedCount`) appears on every suppressed row
and on a thinking row only once it has ever been suppressed (N > 0), so a
never-suppressed class shows no count clutter), and **villagers** (renamed from
"souls", spec 015/TASK-56 — now a two-view inspector rather than a flat
roster). The villagers **roster** shows per agent: a selection cursor,
status, current goal, needs gauges, a leading `bulk n/24` derived-load
reading (spec 013 T015, SC-006; `sim.Bulk`/`sim.BulkCap` — the same function
the reducer/executor clamp gathers and crafts against, so the number never
drifts from what an action will actually do), then the full carried-inventory
line — wood/stone/water/planks/refined-stone counts, the food triplet
raw/cooked/meals, and (when carried) a spear count with the most-worn spear's
remaining uses. While the villagers tab is visible, `j`/`k`/`g`/`G` move the
cursor and `⏎` opens the selected villager's **detail view**
(`villagerDetailBody`): identity/vitals, an objective line (active
`Intent.Goal` marked current; else the reducer-stamped `Agent.LastGoal` +
tick marked `last:`; else "no objective yet" — [[sim-state-reducer]]),
itemized inventory, beliefs/narrative when consolidation has produced them,
and episodic memories most-recent-first, each section truncating bottom-up
inside the pane budget. From the detail view, `d` opens the **decisions
sub-view** (spec 020/TASK-63, `villagerDecisionsBody`): the villager's recent
cognitions as causal chains, most-recent-first — a when/class header, the
stimulus line, each tool call as `ordinal. tool — phrase (reason)`, and the
terminal outcome or an explicit `in progress — no outcome yet` marker; router
suppressions render as one `didn't think because …` entry. Chains come from
the client-side **decision-trace projection** (decisions.go): `applyEvent`
feeds every `cog.thought`/`cog.tool_call`/`cog.outcome` into `Model.traces`
before the ring append, joining on the shared job ID, so the stimulus is
resolved once at thought-ingest from the chronicle ring in the digest voice
(a pre-connect trigger degrades to a neutral `stimulus #N` reference,
trigger 0 to a cadence phrase) and the stored chain survives the ring's
500-event eviction. Attribution: the thought/outcome payload's agent, else a
villager job-ID parse for fragments; `turn-metatron-*` jobs go to a sentinel
and `conversation-*` jobs are never ingested. `ingestOutcome` also skips the
NON-terminal `sim.OutcomeRetried` marker (spec 025, TASK-72): the tool-loop
consumers emit it AFTER a landed run's door already recorded the real terminal
outcome, so folding it in would overwrite `landed` with `retried` — the marker
stays in the event log for trail-level retry counting, it just never becomes a
chain's outcome (the same disregard conversation outcomes get via the job-ID
prefix guard). The projection is bounded
(`decisionChainCap` 20 chains per agent, oldest evicted) and resets wholesale
on reconnect like the replica. Verdicts and outcomes render ONLY through the
sweep-tested plain-language `verdictGlossary` — raw enum strings never reach
the screen (an unknown value gets a safe generic phrase). `j`/`k` scroll the
sub-view (render-time clamped), and `esc` unwinds decisions → detail →
roster ahead of the solo-release chain; selection state survives tab
switches and is clamped on reconnect. Full soul.md persona files stay on
disk per [[agent-mind]]. The same glossary feeds Metatron's inline verdict
rows: a `turn-metatron-*` `cog.tool_call` appends one `» tool — phrase`
transcript row at ingest (`metatronVerdictRow`), which
`classifyTranscriptLine` labels `note` and styles as cog telemetry — the
angel's refused and landed calls are visible in the transcript where before
only the RPC reply's `⚡` miracle lines appeared.

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
Families carry color-role tints, key tokens (names, speech, amounts, causes)
carry emphasis, and four high-salience types (`agent.died`, `gru.attacked`,
`social.chest_taken`, `norm.violated`) render whole-line alert. Since spec 038,
`agent.build_failed` ([[executor]], [[event-types]]) gets its own registry
entry — builder, emphasized goal, and emphasized reason ("Ash's build_wall_stone
failed — site no longer buildable") — reading as a failure at a glance without
promoting to the whole-line alert tier; a cancelled build was previously
indistinguishable from a finished one because it shared `agent.intent_done`'s
plain "finished" line. Since spec 041, four more [[mental-maps]] event types get registry entries,
all sharing a first-fact-plus-count shape (a full fact list would flood the
line; the detail pane holds the payload verbatim): `agent.saw` ("Ash saw fire
at (x,y) (+N more)"), `social.place_told` ("Ash told Birch of fire at (x,y)
(+N more)"), `agent.map_corrected` ("Ash found fire at (x,y) gone (+N
more)"), and `metatron.place_revealed` ("Metatron revealed fire at (x,y) to
Ash (+N more)", Metatron as subject, the nudge convention). Since spec 042,
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
and `metatron.charter_observed` ("Metatron ran under charter <fingerprint>
(default|player-authored)" — the charter-revision stamp the morgue aligns
deaths against). Since spec 046, two [[curriculum-ladder]] types get registry
entries, and `familyByNamespace` maps the new `curriculum` namespace onto the
existing metatron family voice — the ladder is the guardian's domain, not a
distinct visual role: `curriculum.exercise_passed` ("the <exercise> exercise
was passed (<stage>)") and `curriculum.stage_unlocked` ("Metatron's watcher
earned <stage name> (proven by <exercise>)", the display name resolved
through `skin.StageName` like the CLI's stage line). The four
[[metatron-miracles]] types render in the metatron family voice, with a
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

Input follows the **focus contract** (`docs/design/tui/patterns/focus-contract.md`):
viewing never captures typing; `m` focuses the minibuffer (amber border, inline
`esc release · ⏎ send` hint), `esc` always releases, and no keypress is a
silent no-op — the old rule where the metatron pane owned every key while
active is gone. Time controls (minibuffer unfocused): space toggles
pause/resume based on last-known status; `[`/`]` step through `speedSteps`
(1x → 4x → 8x → 16x → 32x — max is deliberately off the watchable ladder,
TASK-20); `q` detaches — the world keeps running. On an ended world (spec
044) all three clock keys are no-ops and the footer says so — see the
postmortem posture above.

**Help overlay** (spec 045/TASK-116; `help.go`): `?` from any non-text-entry
mode opens a context-sensitive help overlay — the head of the key-dispatch and
esc-release chain (help → minibuffer → decisions → detail → solo → home) —
checked in `handleKey` right after ctrl+c and only when the minibuffer is
unfocused, so a focused `?` still types into the buffer. The overlay freezes
the mode it opened from (`helpMode`) and owns the keyboard while open: `t`
flips the mode page's basic/advanced key tiers, `tab`/`shift+tab` cycle three
sections (mode keys · screen walkthrough · lessons pull-reference), `n`/`p`
page across all six mode pages (global/home, minibuffer, inspect, villagers
roster/detail, solo/narrow — how the minibuffer page stays reachable),
`J`/`K` scroll via the standard pager idiom, and `esc`/`?` dismiss exactly
one layer; every other key is inert. Rendering is body replacement (the solo
zoom slot) in both layouts — chrome stays, output remains exactly
terminal-height. Content is static tables in `help.go`: per-mode key rows
(basic ≈ the footer-hinted set), `headerAnatomy` rows covering every
conditional badge, dock-tab rows, and a `mapGlyphs` table **shared with
`renderMapGrid`'s legend line** (`legendGlyphLine`) so the overlay's glyph
walkthrough and the map legend cannot silently diverge — extracting it also
fixed a real gap: the gru's `G` was drawn but never listed in the legend
text. The shared table gained the `✝` grave row with spec 044; it is the
dead-agent-on-grave carve-out in `renderMapGrid` (above) that keeps the row
honest — without it the map could never actually show the glyph it
advertises. The lessons section is the pull-reference seam for the future
first-occurrence lesson projection: an empty `helpLessons` table whose
entries are content additions, no structural change. All content is
model-independent — byte-identical with nil status/replica (the no-LLM floor
beneath an absent angel). Footer hints advertise `· ? help` in every mode
except the focused minibuffer; while the overlay is open the footer shows the
overlay's own hint. A keymap sweep test (`help_test.go`) mechanically ties
every advertised binding to a real handler and every handled binding to
exactly one tier of its mode page.

## Connections

[[ipc-client]] is the transport; [[ipc-protocol]]'s `state` command exists for this
replica pattern; [[sim-state-reducer]] supplies the shared `Apply`; [[chronicle]]
fills the story pane and [[event-types]] the raw feed; [[cli-promptworld]] mounts
it as the `ui` subcommand. The header's governed-speed suffix and the two
governor digest lines read [[cognition]]'s `ShedThreshold` and the
`clock.governor_shed`/`clock.governor_recovered` payload the [[daemon-lifecycle]]
governor sampler emits through the loop. The metatron pane's standing-orders
block and transcript lines project [[metatron-orders]]' `Status.Orders`/
`TurnResult` fields verbatim, with no client-side re-derivation. The header's
`[llm: …]` badge and the metatron pane's per-provider condition line read
[[llm-provider-health]]'s `ProviderStatus.Condition`/`ConditionDetail`/
`ConditionRemedy` fields off the same polled `Status.LLM`. The header's
`[suppressed: …]` badge and the metatron pane's `horizonLines` block both
read the polled `Status.Horizon` — [[ipc-server]]'s `horizonClasses`
composition backed by [[cognition]]'s `LiveHorizon` and
[[llm-orchestrator]]'s `SuppressionCounts` — with no client-side
re-derivation, the same "polled, not projected" posture as the LLM condition
surfaces. The metatron pane's stage segment and the two curriculum digest
rows are [[curriculum-ladder]]'s TUI surfaces (spec 046), reading the
angel's `Status.Stage`/lock fields and the `curriculum.*` event payloads. [[mental-maps]]'s four place-knowledge event types render through
the raw digest feed with no dedicated pane of their own — the map/prompt
side of the feature lives entirely in [[agent-mind]]/[[executor]], not here.

## Operational notes

Rendering requires no daemon round trips — map updates come from pushed events, so the
UI stays smooth at max speed (the chronicle simply scrolls fast). The four spec-029
standing-order event types (`metatron.order_placed`/`order_triggered`/
`order_cancelled`/`order_expired`) carry `digestRegistry` entries (digest.go —
"Metatron set a watch: …" / "…watch came true/released/lapsed", the placed
condition truncated to 80 runes and quoted through the same speech helper as
nudge text; the id-only lifecycle payloads reference the watch by id), so order
activity reaches the raw chronicle feed as well as the dedicated metatron-pane
block and transcript lines above; `TestCatalogSweep` pins the coverage against
[[event-types]]' backticked catalog.
Unit tests cover pane
navigation, replica application, ring capping, quit behavior, the widescreen layout
math (layout.go), the digest grammar (per-family digests + the catalog sweep in
digest_test.go, plain/segment equivalence under wrap), focus-contract key
routing in both layouts, exact-height rendering invariants across sizes and dense
content (including all help-overlay states), the help overlay itself
(help_test.go — per-mode routing, tier/section navigation, the keymap sweep,
no-LLM byte-identity, a zero-side-effect soak, and the lessons-seam fixture),
and resize round-trips with live selection; an expect-driven PTY smoke test
drives the real binary. When real systems land, dock tabs graduate from stubs without
changing the replica machinery.

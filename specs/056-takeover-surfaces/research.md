# Research: Takeover surfaces (spec 056)

## R1 — Overlay-owner state and precedence

**Decision**: one enum on the Model (`takeover: none | ceremony |
postmortem`) plus a `ceremonyDeferred` flag; transitions: stage_unlocked →
ceremony (unless postmortem open → defer); run.ended → postmortem (always,
replacing an open ceremony); esc → none. The body-replacement render branch
follows the help overlay's established slot discipline (helpPanelView
precedent — exact height, chrome visible). Postmortem auto-open: on connect
when `runEnded()` (the dual-source predicate) and not previously dismissed
THIS attach; `p` reopens while ended.

**Rationale**: both pages state the precedence identically; the help
overlay already proves the slot mechanics under the exact-height harness.

## R2 — The shared report-card renderer (D5)

**Decision**: `reportCardView(def, facts, mode, width)` — mode ∈
{concluded (met/missed), live (met/pending)}; rows = plain-language term +
marker + backing event reference (count/condition), the exercise-panel
gauge vocabulary (panels/exercise.md) reused so TASK-119's tab and this
renderer can't drift. A thin consoleCard wrapper proves seam composition
(spec 053's interface); production console wiring stays TASK-115's.

**Rationale**: postmortem.md authors the renderer here ("one renderer,
three sites"); making it a plain view function keeps all sites trivially
identical.

## R3 — Morgue evidence rows from the replica

**Decision**: derive rows from the replica at open time: the state's death
ledger (name, day, cause — the run.ended payload/state facts spec 044
maintains) + the charter-observation timeline (closest observation ≤ each
death, the morgue.md alignment rule) read from the chronicle ring/event
facts the client already holds; no file I/O, no IPC additions (FR-006). If
the ring has rotated past old observations on a very long run, rows render
the observation as unknown honestly (the on-disk morgue.md remains the
durable archive — D1).

**Rationale**: the TUI is a read replica; the scribe's morgue.md is the
durable render — the takeover is a live projection, not a file reader.

## R4 — Scored-vs-ambient detection

**Decision**: scored ⇔ the status/replica carries a scenario exercise
(TASK-119's `scenario_exercise` status field when merged; else the manifest-
less client treats every world as ambient). The postmortem renders the card
only when the exercise definition + rubric facts are actually available —
missing data → ambient form (edge-case honesty rule).

**Rationale**: sequencing-safe in either merge order with 119.

## R5 — Ceremony content + replay

**Decision**: ceremony renders from the unlock event + compiled exercise
definition + skin-resolved stage identity (D6 voice text lives beside the
stage identities in the skin substrate — one authored line per stage,
player-authorship register). Replay: the `?` overlay gains a
ceremony-replay section listing earned unlocks (replica
`StagesUnlocked`/`CurriculumPasses` + unlocks facts) re-rendering the same
stored content; `promptworld stages` already carries the CLI facts
(unchanged).

**Rationale**: stored-not-regenerated (page rule); spec 045's help content
contract is amended deliberately (the D9 precedent).

## R6 — Skin posture

**Decision**: all fiction strings (chapter voice, titles) resolve through
the skin lookup (TASK-121's contract — expected merged first per Lane 3
ordering); any string the contract lacks is added to the default-skin table
in this PR per the contract's §4 downstream obligations. No new bare
literals in either merge order.

## R7 — Design-page amendments

ceremony.md + postmortem.md → shipped (real symbols; the `?`-replay row
filled); help.md (replay entry); keymap.md (`p` + takeover rows + parity);
guardian-console.md (seam note names reportCardView); re-pins throughout;
`check-tui-design.mjs --changed` gates.

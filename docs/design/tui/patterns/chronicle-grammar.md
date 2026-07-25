---
title: Pattern — chronicle grammar
class: pattern
status: shipped
verified_against: ab212309e8fd13aa069e7c6d7c7c7c5821213835
sources:
  - internal/tui/digest.go
  - internal/tui/grammar.go
---

# Pattern: chronicle grammar

How one event becomes one feed entry. Applies wherever the chronicle renders
(dock tab, solo, narrow fallback). Goal: every cataloged event reads as a
natural-language or labeled summary at a glance — never a raw JSON dump; the
verbatim payload is always visible below in the always-on detail pane while
paused (panels/chronicle.md "Mode 2").

**TASK-60 (spec 018-chronicle-digest)** replaced the old class-table/JSON-echo
grammar with a per-type digest registry (`internal/tui/digest.go`) and dropped
`#seq` from the feed line — selection/inspection now carry `seq` in the detail
pane instead. The authoritative per-type digest contract lives in
[`specs/018-chronicle-digest/contracts/digest-grammar.md`](../../../../specs/018-chronicle-digest/contracts/digest-grammar.md)
(§3, ~83 rows); this doc covers the shape everything renders into, not each
type's exact wording.

## Line format

```
solo:   <TICK> <HH:MM>  <type>            <summary>
dock:    <HH:MM> <short-type>  <summary>          (tick dropped; wraps ≤3 lines)
```

- Tick right-aligned to the widest visible tick in the current window; time
  fixed width 5 (`HH:MM`); type column padded to the widest visible type
  (solo, cap 26 runes) or the type's last `.` segment (dock, cap 10) — all
  computed fresh per render over the visible window, never a global fixed
  budget (research R5).
- `#seq` no longer appears on the feed line; it's shown in the detail pane's
  `DETAIL · seq N` rule line (panels/chronicle.md "Mode 2").
- Width overflow: solo truncates one line per event with `…`; dock wraps to
  ≤ 3 lines then truncates the last with `…`.
- Unknown/unregistered types, and any registered type whose payload fails to
  unmarshal, fall back to the pre-digest form: type + compact resolved-name
  JSON — never blank, never an error (contract §1, FR-002/FR-003).

## Digest registry (replaces the old class table)

Every cataloged event type has a `digestFunc` entry in
`internal/tui/digest.go`, keyed by full type, returning the summary as
ordered styled segments (`seg{Text, Role}` — `internal/tui/grammar.go`). A
catalog sweep test (`internal/tui/digest_test.go`, contract §7) fails the
build if any cataloged type has no entry, if a fixture type falls back to
raw JSON, or if a registry key isn't in the fixture — so a new event type
forces a deliberate digest (or fixture) change rather than silently landing
as JSON.

## Suppression/remedy rows in the raw feed (specs 028/031/033/035)

The raw feed's labeled-voice families (`cog`/`clock`) carry the chronicle's
own view into engine suppression and remedy — distinct from, but voiced in
the same terse style as, [systems.md](../panels/systems.md)'s live horizon
block:

- `cog.outcome` — every cognition job's terminal outcome, `job=… <outcome>
  agent=… stale=…t wall=…ms` (plus `kind=`/`reason=` when present); `<outcome>`
  is the raw outcome word verbatim (`landed`, `suppressed`, `retried`, …).
  TASK-32's chapter-suppression check (a narrator-tier `cog.outcome{suppressed}`
  record dropping a chronicle chapter with no call spent) reads as one of
  these rows.
- `cog.recalibration_recommended` (spec 031, estimator breach-adoption) —
  `tier=… est=<prior>→<adopted>s/pt spikes=… window=…` — the chronicle's
  remedy signal: a live re-estimation the router adopted.
- `clock.governor_shed` / `clock.governor_recovered` (spec 028/033) —
  "governor shed/recovered `<from>`→`<to>` `debt=N% jobs=N`" — the debt
  arithmetic is `debtPercent`, the same shared function behind the header's
  governed-speed suffix ([../pages/home.md](../pages/home.md)).

**Why these don't route through `verdictGlossary`**: the villagers
decisions sub-view and the guardian tab's inline verdict rows
([../panels/villagers.md](../panels/villagers.md),
[../panels/guardian.md](../panels/guardian.md)) translate every verdict/
outcome string through the sweep-tested plain-language `verdictGlossary` —
raw enum strings never reach the screen there, because those views read as
narrative prose. The raw chronicle feed is deliberately the opposite
register (labeled-voice families "read as `key=value` fields, stable order" —
see below): its outcome/verdict words appear **verbatim**, by design, because
this is the terse operator-telemetry view, not the narrated one. Both
disciplines serve the same "no raw enum leaks into a *narrative* line"
invariant; they just draw the narrative/telemetry line in different places.

## Voice by family

Family = the event type's namespace prefix (`world`, `clock`, `sim`,
`agent`, `social`, `meeting`/`norm` → merged into one `governance` family,
`gru`, `chronicle`, `metatron`, `daemon`, `cog`).

| Family | Voice |
|---|---|
| agent, social, sim, world, governance, gru, `metatron`, chronicle | natural phrase — e.g. `Ash foraged at (14,9)`, `Ash→Rowan "the fire's low again"` |
| cog, clock, daemon | labeled fields, space-separated, stable order — e.g. `job=j1 landed agent=Ash stale=10t wall=220ms` |

Emphasis roles inside a phrase-voice summary: **name** (every resolved agent
name), **speech** (a quoted utterance — the speech privilege
`social.conversation_turn`/`rumor_told` and a few governance/mind types
carry), **emphasis** (amounts, item kinds, causes, outcomes, coordinates).
High-salience types (`agent.died`, `gru.attacked`, `social.chest_taken`,
`norm.violated`) render the **whole line** in the alert role regardless of
family.

Several digest rows render slightly differently than a naive reading of
their contract row would suggest, because the real payload struct doesn't
carry every field the template names — each such row is called out with a
"verified: …" note directly in the contract table and in `digest.go`'s
registry comments (e.g. `agent.needs_changed` has no `water` field;
`agent.memory_promoted`/`faded` carry a hash and tick, never the memory
text; `social.promise_broken` carries only an id, no from/to).

## Inspector (detail pane, paused)

The always-on detail pane (panels/chronicle.md "Mode 2") shows the
**selected event verbatim**, pretty-printed with 2-space indent: `seq`,
`tick`, `type`, and the raw `payload` exactly as persisted — integer
indices intact. Name resolution appears as trailing `// name` annotations,
never as payload rewrites:

```
{
  "seq": 1202, "tick": 8846,
  "type": "social.conversation_turn",
  "payload": {
    "conv": 102,
    "speaker": 1,     // Rowan
    "listener": 0,    // Ash
    "text": "I stacked wood at dawn, ask Birch"
  }
}
```

Rationale: the feed line is a *view* (names resolved, fields elided, a
family voice applied); the inspector is *evidence* (exact bytes, annotated).
Both must exist; neither substitutes for the other. Oversized payloads
(`world.migrated`, which embeds the full `sim.State`) are windowed by the
pane's own scroll (`J`/`K`) rather than ever rendered in full — see
panels/chronicle.md.

## Color roles

Style tokens (bind to Lipgloss styles, one per role, `internal/tui/views.go`
token block — roles, never raw colors, at every call site):

| Role | Token | Notes |
|---|---|---|
| `dim` | `styleDim` | tick/time, fallback payloads, no distinct family tint (daemon) |
| `family/world` | `styleFamilyWorld` | blue |
| `family/clock` | `styleFeedClock` | yellow — unchanged from before this feature |
| `family/sim` | `styleFamilySim` | green (plain, vs. name's bold green) |
| `family/agent` | `styleFamilyAgent` | cyan — the plurality of events, today's former default type color |
| `family/social` | `styleFamilySocial` | magenta |
| `family/governance` | `styleFamilyGovernance` | amber (meeting.\*/norm.\*) |
| `family/gru` | `styleFamilyGru` | bold red — predator threat |
| `family/chronicle` | `styleFamilyChronicle` | bright magenta — the narrator's voice |
| `family/metatron` | `styleFamilyGuardian` | violet — the guardian, otherworldly |
| `family/cog` | `styleFamilyCog` | faint — telemetry noise |
| `name` | `styleFeedName` | resolved agent names, bold green — unchanged |
| `speech` | `styleFeedSpeech` | quoted utterances, bold — unchanged |
| `emphasis` | `styleFeedEmphasis` | amounts/kinds/causes/coords — underline |
| `alert` | `styleFeedAlert` | whole-line, bold red — `agent.died`/`gru.attacked`/`social.chest_taken`/`norm.violated` |
| `selection` | `styleFeedSelect` | inspect-mode row background (reverse) — unchanged |

For labeled-voice families (cog/clock/daemon) the family tint applies to
the **whole line** (type column + summary — the summary already reads as
`key=value` fields, so there's no separate name/speech/emphasis treatment
inside it). For every other family the tint applies to the **type column
only**, and the summary is styled segment-wise: `renderChronicleRow`
(`internal/tui/views.go`) dispatches to an alert / labeled-voice / phrase
path per contract §2, and the phrase path (`styleWrapLine` +
`paintStyledLine`, `internal/tui/grammar.go`/`views.go`) always wraps or
truncates the **plain** text first and paints each physical line's
characters by their source segment afterward — styling can never split an
ANSI escape mid-truncation.

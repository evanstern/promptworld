# Contracts: Knowledge Events, Tool Surface, Prompt Section

**Date**: 2026-07-24 · **Data model**: [../data-model.md](../data-model.md)

The feature's external interfaces are (1) new event types on the store/IPC stream —
consumed by the TUI digest, chronicle narrator, and mind absorb; (2) one new villager tool;
(3) the prompt's known-places section; (4) rejection strings returned to the model. These
are the contracts implementation must honor; payload field tables live in data-model.md.

## 1. Event-type contract

| Type | Emitter | Reducer effect | Digest | Chronicle | Absorb trigger |
|---|---|---|---|---|---|
| `agent.saw` | executor perception sweep | upsert witnessed facts | row required | none (too chatty) | no |
| `agent.map_corrected` | executor perception sweep | remove facts + situated memory | row required | grammar line | yes, if a removed fact matches the agent's current intent target |
| `social.place_told` | executor talk sidecar | upsert told facts (absent-or-staler rule) + memories both sides | row required | grammar line | no |
| `metatron.place_revealed` | Metatron `send_vision` (via InjectSocial; whitelist entry required) | upsert revealed facts + omen memory | row required | grammar line | yes (existing vision trigger) |

Gates that enforce this table (must be green before Done):
- `TestCatalogSweep` (`internal/tui/digest_test.go`): every type above needs a
  `digestRegistry` entry **and** a fixture row **and** a backticked row in
  `docs/wiki/event-types.md`.
- `ValidateToolCoverage` (`internal/sim/toolcheck.go`): the `search` tool (below) must map
  to a resolver + duration entry; `metatron.place_revealed` must join
  `injectSocialWhitelist`.
- Replay-byte-identical harnesses: all four types must reduce identically live and on
  replay (payloads are canonical structs; no map-typed JSON).

Event payload rules (inherited doctrine): absolute values only (facts as they are known at
emission — "no arithmetic that could drift"); context baked at emission (correction
payloads carry the *remembered* fact for narration, never re-derived).

## 2. Tool contract (villager roster)

`search` — goal-door World tool, joins `tool.LoopRosterVillager()`.

- **Description (model-facing)**: "Search unexplored land: walk toward the nearest edge of
  what you know, looking around as you go. Use when you need something you know no place
  for."
- **Args**: none in v1 (a `kind` hint is a documented future extension; selection is
  nearest-frontier regardless, so the hint would only change narration).
- **Resolution**: nearest frontier tile (explored ∧ passable ∧ adjacent-to-unexplored),
  deterministic BFS order; instant goal on arrival (wander-class duration).
- **Failure**: resolver error when no reachable frontier exists (see rejections below).

## 3. Prompt-section contract (`userPrompt`)

Replaces the `Village: …` line. Rendered only from the acting agent's map (fresh facts,
read-time horizon), never from `State.Structures`:

- **Structures**: every fresh known structure, individually, with position and provenance
  flavor — witnessed: `fire at (12,34)`; told: `a fire at (40,12) — Birch told you`;
  revealed: `a fire at (40,12), shown to you in a vision`. No count cap (grouping below
  bounds size; the first-6 truncation is retired).
- **Resources**: grouped per kind with count + nearest: `You know 3 forage spots (nearest
  (22,41)) and 2 stands of trees (nearest (18,9)).`
- **Unknown space**: one orientation line when unexplored land remains: `Land to the
  north-east is unknown to you.` (dominant direction of nearest frontier); omitted when
  fully explored.
- **Empty state**: a villager with no fresh facts of any gated kind gets `You know of no
  fires or shelters yet.` — the model must always be able to tell "I know none" from
  silence.

Contract tests: `internal/mind/prompt_test.go` additions — two agents with different maps
render different sections; >6 known structures all render; unknown-agent section absent
when fully explored; told-provenance phrasing present.

## 4. Rejection-string contract (model-facing, landing ladder shape `"<outcome>: <reason>"`)

| Situation | String (exact reason text owned by implementation, distinction mandatory) |
|---|---|
| Gated verb, no fresh matching fact | `rejected_guard: you know of no <kind>` — MUST be distinguishable from existence-failure |
| Gated verb, known target vanished at landing/arrival | existing guard/contested outcomes (`Birch is gone…`, work re-validation) + `agent.map_corrected` follows |
| `search` with no reachable frontier | `rejected_guard: nothing left unexplored` |
| Ungated verbs (wander, sleep, eat-from-inventory, journal, muse, set_plan) | unchanged |

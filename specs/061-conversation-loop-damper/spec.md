# Feature Specification: Conversation Loop Damper

**Feature Branch**: `061-conversation-loop-damper`

**Created**: 2026-07-25

**Status**: Draft

**Input**: TASK-109 — world-01 evidence: Birch↔Sage held 219 conversation
scenes / 343 talks in 6.2 days despite two 2-game-hour cooldowns. Diagnosis
(2026-07-25, `docs/design/evidence/task-109/findings.md`, event-log proven):
**the planner `talk_to` → hail → `hailStep` founding path carries no
pair-frequency gate of any kind** — 99.1% of Birch↔Sage scenes (and 97.8% of
ALL world-01 scenes) were hail-founded. `canTalk`/`LastTalk` gates only the
ambient adjacency beat; `pairSeen`/encounter-cooldown gates only planner
*arming* on movement; `hailStep` bypasses `canTalk` by deliberate TASK-47
design, never backstopped. The loop: scene → `intent_done` → planner re-arm →
`talk_to` → hail → un-cooled talk → new scene, floored only by the
5-game-minute plan debounce (median observed gap: 288 ticks). Systemic, not
pair-specific. Operator decision (2026-07-25): the pair cooldown lives
**sim-side on the hail founding path**; the novelty gate stays mind-side as a
clearly-marked removable SHIM.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Deliberate talk respects the pair cooldown (Priority: P1)

A villager whose planner decides to talk to a recent conversation partner is
gated the same way ambient talk already is: the world itself (the sim) tracks
when each PAIR last exchanged words, and the hail founding path refuses to
found a new exchange inside the cooldown. The refusal is informative: the
planner's tool result says the pair spoke recently, so the model can choose
differently instead of burning turns.

**Why this priority**: closes the proven leak — 97.8% of all world-01 scenes
came through this ungated path.

**Independent Test**: drive two agents through a hail-founded talk, advance
less than the cooldown, plan another talk_to between them → the landing/hail
path refuses with the informative message; advance past the cooldown → it
founds.

**Acceptance Scenarios**:

1. **Given** a pair whose last exchange was under the cooldown ago, **When** a
   planner talk_to between them reaches the hail founding path, **Then** no
   talk/scene is founded and the intent resolves with a "spoke recently"
   outcome the planner sees.
2. **Given** the same pair past the cooldown, **When** the planner tries
   again, **Then** the hail founds normally (TASK-47 semantics otherwise
   intact — the hail still bypasses the AMBIENT cooldown by design).
3. **Given** the cooldown gate, **When** ambient-beat and encounter-arming
   paths run, **Then** their existing gates behave exactly as today (three
   paths, three gates, no cross-regression).
4. **Given** a world with `tuning.json` setting `encounter_cooldown_ticks`,
   **When** the sim gate evaluates, **Then** it uses the SAME dial the mind
   arming uses (one pair-cooldown doctrine, one dial — spec 048).

---

### User Story 2 - Pair state is world truth (Priority: P1)

The per-pair last-exchange record lives in event-sourced sim state (derived
from the talk events the log already carries), so the gate is deterministic,
replay-visible, and survives restarts — not a mind-session artifact.

**Why this priority**: the gate is only as trustworthy as its state; a
mind-side memory of pair history would vanish on restart and diverge from
replay.

**Independent Test**: run talks, snapshot, restart/replay → the pair record
and gating behavior are identical; pre-061 snapshots load (absent record =
never talked).

**Acceptance Scenarios**:

1. **Given** an `agent.talked` between two villagers, **When** the reducer
   applies it, **Then** the pair's last-exchange tick updates (both orderings
   of the pair are one record).
2. **Given** a pre-061 snapshot (no pair records), **When** loaded, **Then**
   every pair reads "never talked" and the world behaves as an un-gated fresh
   start; snapshot bytes for states without records are unchanged (no
   format_version bump).
3. **Given** a timeline rebase (the migrate/rebase machinery), **When** pair
   ticks are rebased, **Then** they SHIFT with the timeline (rebase taxonomy
   classified).

---

### User Story 3 - Novelty gate: nothing new to say, no scene (Priority: P2) — SHIM

Beyond the cooldown, a pair cannot re-converse until at least one of them has
formed a new memory above a salience floor since their last exchange; and when
a scene does found, the pair's last conversation gist enters the scene prompt
as "you already talked about this". **This is a SHIM** compensating for weak
model-side conversational variety (operator decision 2026-07-24): it is marked
as such at every site (code comment + doc note), designed for removal, and it
is the FIRST place to look if conversations later feel less dynamic than
wanted.

**Why this priority**: the cooldown fixes frequency; the novelty gate attacks
sameness (Birch's fixation loop: "I need to tell Sage everything…" 248 hails).
P2 because it's the softer, removable layer.

**Independent Test**: pair past cooldown but neither has a new salient memory
→ founding refused with a "nothing new" outcome; give one a salient memory →
founds, and the scene prompt carries the last gist.

**Acceptance Scenarios**:

1. **Given** a pair past the cooldown with no new salient memory on either
   side since their last exchange, **When** scene founding evaluates, **Then**
   no scene founds and the outcome says so.
2. **Given** one partner with a new above-floor memory, **When** founding
   evaluates, **Then** the scene founds and its prompt includes the previous
   exchange's gist as context.
3. **Given** the code and docs, **When** read at the gate sites, **Then** the
   SHIM marking and removal condition are explicit.

---

### Edge Cases

- First-ever exchange between a pair: no record → both gates pass (cooldown
  vacuous, novelty vacuous).
- Meetings, governance speech, metatron-prompted speech: out of scope — only
  the villager hail/scene founding paths change.
- A pair separated by the cooldown but hailed BY the other (initiator
  swapped): the record is unordered — same gate both directions.
- Dial turned to 0 via tuning.json: cooldown gate vacuous (existing dial
  semantics); novelty shim still applies (independent layers).
- Replay of pre-061 logs: reducer arms that now update pair state produce a
  populated record where old snapshots had none — the spec-044 precedent
  (Deaths ledger) applies: snapshots are truth for their prefix; new snapshots
  carry the record; determinism tests cover the new arm.
- Scene caps/turn limits/deadlines: unchanged; the damper gates FOUNDING only.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Sim state MUST carry an event-sourced, unordered per-pair
  last-exchange record derived from talk events; absent record ≡ never talked;
  pointer/omitempty-compatible with pre-061 snapshots; rebase-taxonomy
  classified (SHIFT).
- **FR-002**: The deliberate-talk founding path (hail) MUST refuse to found an
  exchange for a pair inside the cooldown, consuming the SAME
  `encounter_cooldown_ticks` dial the mind arming uses (spec 048 accessor);
  the refusal MUST surface to the planner as an informative outcome.
- **FR-003**: The ambient-beat and encounter-arming gates MUST be behaviorally
  unchanged (their tests pin this).
- **FR-004**: Scene founding MUST additionally require a new above-floor
  salience memory on at least one side since the pair's last exchange (the
  novelty SHIM), and a founded scene's prompt MUST carry the pair's previous
  gist; the SHIM MUST be marked removable at every site with its removal
  condition (model tiers make it unnecessary).
- **FR-005**: The salience floor and any new constant MUST be
  promoted-dial-ready (single home, documented) but NOT added to tuning.json
  (dials are earned; the cooldown dial already exists).
- **FR-006**: Replay determinism suites MUST stay green; no format_version
  bump.

### Key Entities

- **Pair exchange record**: unordered pair → last-exchange tick, event-sourced
  on sim state.
- **Founding gates**: the sim-side cooldown gate (hail path) and the mind-side
  novelty SHIM (scene founding) — independent, layered.
- **Last-gist context**: the previous exchange's gist injected into a founded
  scene's prompt.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In tests, the world-01 loop shape (talk → replan → talk within
  the cooldown) founds exactly ONE scene; the retry gets the informative
  refusal. Under the old code the same drive founds two.
- **SC-002**: All three paths' gates proven independent: ambient and
  encounter-arming test suites pass unchanged; the hail gate has its own.
- **SC-003**: Novelty SHIM: no-new-memory founding refused; new-memory
  founding admitted WITH gist context — both in tests.
- **SC-004**: Replay/snapshot compat: pre-061 fixtures load and replay; full
  suite green; rebase taxonomy passes.
- **SC-005**: The SHIM marking is greppable (a distinctive marker string) and
  the removal condition documented in code and in this spec dir.

## Assumptions

- The existing `encounter_cooldown_ticks` dial (default 7200) is the single
  pair-cooldown doctrine for deliberate talk too — no new dial. Worlds that
  want chattier pairs turn the dial (tuning.json), which now affects both
  arming and founding coherently.
- "New memory above a salience floor" reads the existing memory
  salience machinery (spec 019/042 family); the floor is a new named constant
  (promoted-dial-ready, not promoted).
- The last-gist context rides the existing scene prompt assembly; gists are
  already recorded per conversation (convo_record machinery).
- TASK-110's clamp work touches tool text validation, not the hail path —
  no file-level conflict expected; merge order (110 first) already planned.

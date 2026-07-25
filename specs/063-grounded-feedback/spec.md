# Feature Specification: Grounded feedback layer — explain tool, tutor guide, report card

**Feature Branch**: `063-grounded-feedback`

**Created**: 2026-07-25

**Status**: Draft

**Input**: User description: "Grounded feedback layer (board TASK-115; learning-game synthesis Wave 1, operator decision 5: the guardian IS the tutor; reorient D5/D9/D2 rescope). Three merged deliverables over ONE shared data source: (a) a read-only, grant-gated, registry-derived explain tool serving deterministic mechanics facts so the guardian never confabulates mechanics; (b) a default guide + tutor-charter preset making a fresh world's guardian a competent orientation tutor ('ask your guardian' is the game's whatis command); (c) the post-turn report card — a cheap-chain critique attributing outcomes to charter text, citing event-log evidence, riding the TASK-63 trace. Contract: explain is pull, report card is push, one shared data source so the grader never grades on vibes. Tutor lane doctrine: charge-free, faith-free, excluded from every rubric; no initiative-frame relaxation (explaining is speech, not an act). Report card surface: guardian-console card at stopping points + postmortem (D5); guardian verbs reachable from the deterministic ? floor (D9); all new strings skin-token-resolved (D2)."

## Standing resolutions this spec adds

1. **The "report card" is one composed artifact with two ingredient
   classes**: the deterministic rubric checklist (TASK-127's shared
   renderer, event-derived) plus, when a critique is available, the
   **attribution note** — the cheap-chain-authored lines tying outcomes to
   charter text with event citations. The checklist is always authoritative
   and always present when rubric data exists; the attribution note is
   additive prose beneath it, clearly its own block, never a second scoring
   computation. One artifact, one renderer family, three surfaces (console
   card, postmortem, ceremony's checklist-only form).
2. **The guide is game-authored prompt substrate, not a player skill.**
   Player skills bind only from stage 3 (spec 046) and the tutor preset is
   the stage-1 orientation — so the guide composes as compiled-in game
   content (the `persona.TutorCharter` precedent), active on tutor-preset
   worlds, never through the player `skills/` directory and never subject to
   the stage-3 skill lock. No stage-gating doctrine changes.
3. **Explain answers are rendered facts, not model output.** The tool's
   result is composed deterministically from the registry/doctrine sources;
   the model chooses WHEN to call and how to converse around the answer —
   it can never alter the fact text itself.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The explain tool: mechanics facts on demand (Priority: P1)

A player asks their guardian "what does a vision cost?" or "what can you
actually do?" The guardian calls the read-only `explain` tool and answers
from its returned fact sheet — tool rosters and costs, charge economy,
miracle/working kinds and prices, decision classes, map glyphs — instead of
guessing. Facts are derived from the same registries the mechanics run on,
so the answer can never drift from the game.

**Why this priority**: the unreliable-manual hazard is the layer's
foundation — (b) and (c) both stand on this data source.

**Independent Test**: drive a guardian turn (or a unit-level tool call) with
explain granted; assert the returned facts equal the registry/doctrine
values; assert a model cannot alter fact text (the result is composed
tool-side).

**Acceptance Scenarios**:

1. **Given** a world whose grant includes explain, **When** the guardian
   calls it with a topic (roster, costs, charges, decisions, glyphs, or a
   specific tool/kind), **Then** the result is a deterministic fact sheet
   derived from the live registry/doctrine constants (respecting the
   world's ACTUAL grant and stage ceiling — the roster facts describe this
   world, not the full catalog).
2. **Given** the same world state, **When** explain is called twice with
   the same topic, **Then** the answers are byte-identical (no model in the
   answer path).
3. **Given** a topic that doesn't exist, **Then** the tool returns an
   honest catalog of what it CAN explain (a rejected_gate-style repairable
   miss, never a fabricated answer).
4. **Given** explain is absent from a world's `capabilities.json` grant,
   **Then** the tool is structurally absent (declaration, guidance,
   handlers) — the existing three-layer gating discipline.

---

### User Story 2 - Tutor-lane doctrine: explaining is free speech (Priority: P1)

Explain calls and tutor conversation spend nothing and change nothing: no
charge, no world event beyond the standard tool-call telemetry, no faith
(when TASK-118 lands), no rubric term anywhere may reference them, and the
initiative frame is not relaxed — explaining is speech, not an act, so a
turn that only explains still carries its converse reply and consumes no
mediated act.

**Why this priority**: co-P1 — the doctrine IS the design; getting it wrong
poisons both the economy and the curriculum.

**Independent Test**: a turn calling explain leaves the charge bank
untouched and lands no world-mutating events; the rubric-term catalog sweep
proves no exercise references tutor-lane telemetry.

**Acceptance Scenarios**:

1. **Given** a guardian turn that calls explain, **Then** the charge bank
   is unchanged, no injected world event results (`Effect` expressive-class,
   empty `Events` — the pause/start meta-tool precedent), and the turn's
   one-act budget is NOT consumed by explain calls (read-only tools don't
   count as the mediated act).
2. **Given** the exercise catalog (spec 046/054), **Then** no rubric term
   names explain/tutor telemetry — enforced by a sweep test, so future
   exercises can't quietly grade tutoring.
3. **Given** the fixed frame, **Then** no initiative-frame wording changes:
   the guardian still acts only on player ask or pre-authorized watch.

---

### User Story 3 - The tutor guide: a fresh world's guardian can orient (Priority: P2)

A brand-new player on a tutor-preset (stage-1) world asks "how do I play?"
The guardian gives a competent orientation: what the world is, what the
player's verbs are, what to watch, how to ask for things — grounded in the
guide's game-authored text and the explain tool for any mechanics number.

**Why this priority**: decision 5 made flesh; depends on US1 for grounding.

**Independent Test**: compose a tutor-world turn prompt; assert the guide
text is present beneath the charter and above the fixed frame; drive
canned orientation questions through an eval-style check that answers cite
guide/explain content rather than invented mechanics.

**Acceptance Scenarios**:

1. **Given** a tutor-preset world, **When** a turn composes, **Then** the
   compiled-in guide text is in the editable zone (after the charter,
   before the fixed frame), and the fixed frame remains last on every path.
2. **Given** a non-tutor world, **Then** prompt composition is
   byte-identical to pre-feature (the guide is preset-scoped).
3. **Given** the tutor guide text, **Then** it instructs answering
   mechanics questions via explain (the how-to-tutor contract) and carries
   no fiction literals outside skin tokens' default values.

---

### User Story 4 - The report card: outcomes attributed to charter text (Priority: P2)

At a natural stopping point (run end, pause, exercise resolution) the
guardian console shows a report card: the rubric checklist (when an
exercise exists) plus an attribution note — two or three cheap-chain
sentences tying what happened to the player's charter text with event
citations ("your charter never mentions coordinates; the working was
rejected twice for them — seq 812, 907"). The same card composes into the
postmortem. Between stopping points, a badge on the console indicates a
fresh card; nothing ever interrupts mid-run.

**Why this priority**: the push half of the layer; depends on US1's data
source and the shipped console seam/renderer.

**Independent Test**: fixture a trace with rejected tool calls + a charter
fingerprint; run the critique on the cheap chain (or a stubbed chain in
tests); assert citations reference real recorded events and charter text;
assert card appearance only at stopping points.

**Acceptance Scenarios**:

1. **Given** a stopping point on a world with recorded guardian activity,
   **When** the card is produced, **Then** it composes the rubric checklist
   (exercise worlds) and/or the attribution note, whose every citation
   names a real recorded event (seq) and quotes/references actual charter
   text — grader inputs are exactly the shared data source (registry facts
   + trace + rubric evidence + charter revision timeline), never free
   recollection.
2. **Given** the critique chain is unavailable (no LLM, budget exhausted,
   route failure), **Then** the card degrades to its deterministic parts
   (checklist; or absence with no error theater) — the push layer never
   blocks play.
3. **Given** mid-run play between stopping points, **Then** no takeover, no
   card injection — at most the console badge (existing unseen-badge
   pattern).
4. **Given** the postmortem opens on a world with a produced card, **Then**
   the card's content renders inside it (the D5 surface), sharing the
   checklist renderer TASK-127 shipped.
5. **Given** the attribution note is produced, **Then** it is recorded
   durably (rides the existing prose channels) so re-opening the card
   re-reads the stored note — never re-graded retroactively.

---

### User Story 5 - The ? floor teaches the verbs (Priority: P3)

The help overlay gains its guardian section (D9): static-per-stage,
model-free — the stage's identity/concept, the granted verbs at this
world's stage/grant, and one example ask per verb — so a player who never
converses still learns what asking looks like, at the deterministic floor.

**Why this priority**: D9's cheap deterministic floor beneath the tutor;
last because it's pure client rendering over facts that exist.

**Independent Test**: render the help overlay's guardian section across
stage/grant fixtures; assert byte-identical output for identical status
(no-LLM invariant holds) and verbs matching the world's effective grant.

**Acceptance Scenarios**:

1. **Given** any world, **When** `?` opens the guardian section, **Then**
   it renders stage identity (skin-resolved), the granted verbs (from the
   status grant summary), and one canned example ask per verb — model-free,
   byte-identical for identical status.
2. **Given** spec 045's content contract, **Then** the section is a
   deliberate amendment recorded on the help overlay's design page.

---

### Edge Cases

- **Explain asked about an ungranted tool**: the fact sheet says it exists
  but is not granted in this world (honest catalog vs grant distinction) —
  never pretends the tool doesn't exist, never teaches a verb the world
  lacks without saying so.
- **Charter changed between the graded window and card display**: the
  attribution cites the fingerprint it graded (the charter-revision
  timeline), so a card can honestly say "under charter a1b2c3…".
- **No guardian activity at a stopping point**: no attribution note (nothing
  to attribute); checklist alone or no card.
- **Pause spam**: producing a card is debounced per stopping-point class
  (a pause card at most once per pause episode; run end and exercise
  resolution are naturally once).
- **Skin interactions**: all new strings (guide text framing, card labels,
  help-section labels, example asks) resolve via the skin contract; example
  asks use the world's working/vision/omen nouns.
- **Tutor guide on a pre-ladder world**: pre-ladder worlds have no preset;
  the guide composes only on tutor-preset worlds (stage-1 default) — the
  orientation content remains reachable via `?` and explain everywhere.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: A read-only `explain` tool MUST exist in the tool registry:
  grant-gated through the existing three-layer discipline, declared to the
  loop roster, returning deterministic fact sheets composed tool-side from
  the live registries/doctrine constants (roster + costs + charge economy +
  working kinds/prices + decision classes + map glyphs + per-topic detail),
  scoped to the world's effective grant/stage ceiling.
- **FR-002**: Explain MUST be tutor-lane by construction: expressive-class
  effect (no events beyond standard tool-call telemetry), zero charge, zero
  faith, not counted as the turn's mediated act, and structurally absent
  when ungranted. Unknown topics return the explainable catalog as a
  repairable miss.
- **FR-003**: A rubric-hygiene sweep MUST prove no exercise rubric term
  references tutor-lane telemetry (extends the existing catalog sweeps).
- **FR-004**: A compiled-in tutor guide MUST compose on tutor-preset worlds
  in the editable zone (after charter, before fixed frame; the
  SOUL-fragment seam family), instructing mechanics-via-explain; non-tutor
  worlds compose byte-identically to pre-feature; the fixed frame stays
  last on every path.
- **FR-005**: A report-card producer MUST run at stopping points only (run
  end, pause-debounced, exercise resolution): composing the rubric
  checklist (via TASK-127's shared renderer) and a cheap-chain attribution
  note grading ONLY from the shared data source (registry facts, the
  recorded tool-call/verdict trail, rubric evidence, the charter-revision
  fingerprint timeline, actual charter text), with every citation naming a
  recorded event. Chain unavailability degrades to the deterministic parts
  silently.
- **FR-006**: The card MUST render via the console card seam (spec 053) at
  stopping points with the unseen-badge pattern between them, and inside
  the postmortem (spec 056) — never as its own takeover, never mid-run.
  Produced notes are stored and re-read, never re-graded.
- **FR-007**: The critique MUST route on a cheap chain (the
  fuzzy-watch-confirm routing precedent), budget-capped, and MUST NOT run
  on no-LLM worlds (deterministic parts only).
- **FR-008**: The `?` overlay MUST gain the D9 guardian section:
  static-per-stage, model-free, byte-identical for identical status —
  stage identity, effective granted verbs, one example ask per verb; spec
  045's content contract amended deliberately.
- **FR-009**: All new user-facing strings MUST resolve through the skin
  contract (D2); new tokens land in the default table + doc twin +
  completeness test per that contract's §4.
- **FR-010**: The design reference and doctrine docs MUST be amended in the
  same PR: help overlay page (guardian section), guardian console page
  (card production now real — seam note updated), keymap/parity notes if
  any binding changes, re-pins throughout; the tool registry's design/wiki
  surfaces follow in the re-ground step.
- **FR-011**: The linear projection (D1) MUST hold: explain facts reachable
  via the CLI guardian conversation; the stored attribution note visible in
  the transcript/soul channels a linear client already reads.

### Key Entities

- **Fact sheet**: the deterministic, per-topic explain result — derived,
  scoped to the world's grant/stage, byte-stable per state.
- **Tutor guide**: compiled-in game-authored orientation text, tutor-preset
  scoped, editable-zone composed.
- **Report card**: rubric checklist + attribution note; produced at
  stopping points; stored; rendered via console seam + postmortem.
- **Shared data source**: the single grounding set (registries/doctrine
  constants, tool-call verdict trail, rubric evidence, charter revision
  timeline + text) both explain and the grader consume.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of explain answers equal their registry/doctrine ground
  truth across the topic catalog (sweep test); zero model-generated bytes
  in any fact sheet.
- **SC-002**: Tutor-lane neutrality holds under test: zero charge delta,
  zero world-mutating events, zero rubric references, zero initiative-frame
  diffs.
- **SC-003**: On a tutor-preset world, canned orientation questions are
  answerable with guide+explain grounding (eval-style fixture check); a
  non-tutor world's prompts are byte-identical to pre-feature.
- **SC-004**: Every attribution-note citation in the test fixtures resolves
  to a real recorded event and the graded charter fingerprint; cards appear
  at 100% of stopping points with data and 0% of mid-run moments.
- **SC-005**: The `?` guardian section is byte-identical across repeated
  renders of identical status (the no-LLM floor invariant) and lists
  exactly the world's effective verbs.
- **SC-006**: Design gate green with the amended pages re-pinned in-PR.

## Assumptions

- TASK-121 (skin contract), TASK-125 (console seam), TASK-127 (shared
  checklist renderer) are merged before implementation — Lane ordering; if
  127 lags, the checklist half composes via its spec's renderer contract
  when it lands and the attribution note ships standalone behind the same
  card seam.
- The cheap chain reuses the existing route-kind machinery (a new
  `report_card`-class kind routed like the fuzzy watch confirm); llm.json
  absence or route failure = deterministic degradation.
- Stopping-point detection reuses shipped signals (runEnded, clock pause
  transitions, exercise outcome events) — no new event vocabulary; the
  stored note rides the existing prose/telemetry channels (implementation
  chooses among transcript/soul/morgue-epilogue-class doors per doctrine,
  recorded in plan).
- Model tier: **Opus 4.8** — guardian turn pipeline + prompt composition
  (injection-adjacent), cross-package, per the runbook Lane 3 assignment.
- One eval-style fixture check for SC-003 is a test-suite fixture (no live
  model in CI; the behavior-affecting prompt quality follows the TASK-73
  eval-gated precedent where a live eval exists).

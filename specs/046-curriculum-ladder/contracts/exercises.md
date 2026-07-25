# Contract: the shipped exercises

Content definitions consumable by TASK-119's scenario machinery (config-block shape
reserved per the `Meeting` Manifest precedent). Each defines: stage, seed, framing,
event-derived rubric, pass signal, score-narrative framing.

## first-night (stage-1 — The Voice)

- **Teaches**: asking for the right watch — visions and orders through conversation.
- **Seed/setup**: a seeded world tuned so night one is survivable only if the angel
  is directed well (fuel scarce, gru active; exact seed pinned at implementation).
- **Rubric (event-derived terms)**: all villagers alive at dawn of day 2
  (`sim.day_started` with zero `agent.died`); at least one player-directed act landed
  before nightfall (vision/order evidence).
- **Pass signal**: `curriculum.exercise_passed{exercise: "first-night", stage:
  "stage-1"}` — run not ended + rubric terms satisfied.
- **Fail**: `run.ended` (spec 044) or a death before dawn; the morgue epitaph is the
  postmortem (FR-011).
- **Score narrative**: the night's chronicle chapter is the telling; the framing text
  positions it as the village's first night under a new watcher.

## the-law (stage-2 — The Written Word)

- **Teaches**: durable instruction — policy that must outlive the conversation lives
  in the charter.
- **Seed/setup**: a seeded world with a norm-shaped problem (e.g. nighttime curfew
  vs. fuel gathering) that requires sustained, consistent angel behavior across
  several days — impractical to maintain by re-prompting each turn.
- **Rubric (event-derived terms)**: a village norm/vote resolves in the instructed
  direction (`meeting.proposal_resolved` family) within the exercise window; **a
  player-authored charter revision in force across the relevant turns**
  (`metatron.charter_observed` fingerprint ≠ default, spec 044) — the gate conjunct
  that makes SC-004 true (default-charter pass ⇒ no stage-3 unlock).
- **Pass signal**: `curriculum.exercise_passed{exercise: "the-law", stage: "stage-2"}`.
- **Score narrative**: the governance arc as narrated by the chronicle; framing
  positions the charter as the law behind the law.

## Rules

1. Definitions are data + this contract; 046 guarantees they parse and their rubric
   terms are derivable from cataloged event types. Incident scheduling, rubric
   evaluation, and production emission belong to TASK-119.
2. Exercises never introduce mechanics: they are seeds + framing + rubric over
   existing systems (needs, gru, norms/votes, charter).
3. Both exercises must remain honest under the no-dark-patterns criterion (SC-007):
   no time-gating, no attrition grinding — pass = demonstrated competence.

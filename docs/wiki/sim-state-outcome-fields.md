---
name: sim-state-outcome-fields
description: Split from [[sim-state-world-fields]] — the run-outcome and progression fields on sim.State: the Deaths/RunEnd ledger and Ended latch (spec 044), charter/skills observation fingerprints + coordinates (specs 072/077), the MorgueEpilogues ring, curriculum passes/stage unlocks (spec 046), the world-tuning dial set (spec 048), and the Guardian report card (spec 063). Read when touching these fields' shapes or omitempty round-trip guarantees.
kind: component
sources:
  - internal/sim/state.go
  - internal/sim/morgue.go
  - internal/sim/curriculum.go
  - internal/sim/tuning.go
  - internal/sim/reportcard.go
verified_against: 1603d5ac22d9be35469ec88bf2355b7d2f9500bc
---

# Sim state: run-outcome & progression fields

Split from [[sim-state-world-fields]] (corpus-spec v2 size-budget split,
summary-style): the outcome and progression fields `sim.State` carries —
the death/run-end ledger, charter/skills observation state, morgue
epilogues, the curriculum ladder's world-visible facts, the world-tuning
dial set, and the Guardian's latest report card. Every field here is
`omitempty`, so pre-feature snapshots stay byte-identical with no format
bump.

## The fields

**Run outcome** (spec 044, [[morgue]]): the `Deaths []DeathRecord` ledger
(`{agent, tick, cause}`, appended by the `agent.died` arm in application =
event order, bounded by the agent count — it exists so the run-end emission
stays a pure function of (state, batch) rather than a log scan), the
`Ended bool` terminal latch and `RunEnd *RunEnd` summary (`{tick, deaths,
final_cause}`, set once by the `run.ended` arm and never cleared by any
event, so snapshot+replay restores the ended posture on restart for free).

**Charter & skills observation** (specs 072/077): `CharterFingerprint` (the
most recent effective-charter content hash a Guardian turn ran under — the
full revision timeline lives in the event log) with, since spec 072, its
authorship twin `CharterCustom bool` (`charter_custom` — whether that most
recent observation was player-authored, `!CharterObservedPayload.Default`,
set only by the same `metatron.charter_observed` arm; the conservative
false zero value means a pre-072 snapshot with a custom charter in force
reads "not known player-authored" until the next revision is observed —
the-law's rubric charter conjunct reads it, [[scenario-rubric]]) and, since
spec 077, the observation COORDINATES
`CharterObservedSeq/CharterObservedTick` (stamped by the same arm from the
event envelope — what `CharterEvidenceFromState` re-locates pass evidence
with; zero = a pre-077 snapshot, the evidence honestly absent until the
next observation stamps them) plus the skills-observation triple
`SkillsFingerprint`/`SkillsObservedSeq`/`SkillsObservedTick` (set only by
the `metatron.skills_observed` arm — the stage-3 evidence substrate,
[[curriculum-ladder-progression]]).

**Morgue epilogues** (spec 044, [[morgue-epilogues]]): the
`MorgueEpilogues []MorgueEpilogue` bounded ring (`morgueEpilogueCap` 32,
the chronicle pattern) of narrator mourning prose.

**Curriculum ladder** (spec 046, [[curriculum-ladder]]): the ladder's
world-visible facts — `CurriculumPasses []CurriculumPass` (a bounded ring,
`curriculumPassRetain` 32 on the standing-order prune precedent, of
recorded exercise passes each carrying `EvidenceRef{type, seq, tick,
custom}` audit pointers back into this world's log) and
`StagesUnlocked []string` (the once-per-(world,stage) unlock latch — no cap
needed, at most one entry per ladder stage); the per-user unlocks record is
a PROJECTION of these events, this state being the replayable authority.

**World tuning** (spec 048, [[world-tuning]]): the effective world-tuning
dial set, `Tuning *TuningState` (the Journal/Hail/Map pointer precedent —
nil means the five promoted doctrine defaults, set only by the
`sim.tuning_applied` arm (see [[sim-state-apply-world]]), no
`format_version` bump).

**Report card** (spec 063, [[grounded-feedback]]): the guardian's latest
attribution note, `GuardianReportCard *GuardianReportCard` (`{Tick, Seq,
Fingerprint, Note, Citations}` — the reducer keeps only the most recent
card, the `Tuning`/`Journal` pointer-precedent's single-value sibling,
since re-opening the console card seam re-reads the stored note rather
than re-grading; nil until the first card lands).

## Connections

Back to [[sim-state-world-fields]] for the rest of the world's field
catalog and [[sim-state-reducer]] for the whole `State`/`Apply` picture.
[[morgue]] owns `Deaths`/`RunEnd`/`MorgueEpilogues`; [[curriculum-ladder]]
owns `CurriculumPasses`/`StagesUnlocked`; [[world-tuning]] owns `Tuning`;
[[grounded-feedback]] owns `GuardianReportCard`; [[scenario-rubric]] reads
the charter/skills observation state for pass evidence. The Apply arms that
mutate these fields live in [[sim-state-apply-world]].

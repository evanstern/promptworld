---
id: TASK-115
title: >-
  Grounded feedback layer: explain tool, guide skill, tutor-charter preset,
  report card
status: Done
assignee: []
created_date: '2026-07-25 04:43'
updated_date: '2026-07-26 02:40'
labels:
  - learning-game
  - metatron
dependencies: []
ordinal: 86000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Learning-game synthesis Wave 1 (docs/design/learning-game-synthesis.md, 2026-07-25; operator decision 5: the angel IS the tutor). One work item merging three vault recommendations: (a) a read-only, grant-gated, registry-derived explain tool serving deterministic mechanics facts — tool rosters, charge/miracle costs, decision classes/points/budgets, map glyphs — so the angel never confabulates mechanics (the unreliable-manual hazard: the frame forbids inventing events but nothing grounds facts); (b) a default skills/guide.md + tutor-charter preset teaching the base angel to answer how-do-I-play questions through it ('ask your angel' becomes the game's whatis command and a rep of the skill being taught); (c) the post-turn report card — a cheap-chain critique attributing outcomes to charter text ('your charter never mentions coordinates; the miracle was rejected twice for them'), riding the TASK-63 trace. Contract: explain is pull, report card is push, ONE shared data source so the grader never grades on vibes (RimWorld one-corpus/two-deliveries precedent). Tutor lane doctrine: charge-free, faith-free, excluded from every rubric; no initiative-frame relaxation (explaining is speech, not an act). Grounding: Analysis-In-Game-First-Teaching rec 2, Analysis-Learning-Game-Fit rec 4.

Spec: specs/063-grounded-feedback
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 explain tool exists: read-only, registry-derived, grant-gated; answers are derived facts, never model-generated
- [x] #2 Default guide skill + tutor-charter preset make a fresh world's angel a competent orientation tutor
- [x] #3 Report card attributes outcomes to charter text citing event-log evidence, on a cheap chain, sharing the explain tool's data source
- [x] #4 Tutor-lane exclusions hold: no charges spent, no world events, no faith earned, absent from all rubrics
- [x] #5 Report card surface: guardian-console card at stopping points + postmortem; badge between; never mid-run
- [x] #6 Guardian verbs + example asks reachable from the deterministic ? floor (D9)
- [x] #7 All new strings skin-token-resolved (D2)
- [x] #8 Spec phase: Setup
- [x] #9 Spec phase: Foundational — the shared data source
- [x] #10 Spec phase: User Story 1+2 — explain + tutor-lane doctrine (P1)
- [x] #11 Spec phase: User Story 3 — tutor guide (P2)
- [x] #12 Spec phase: User Story 4 — report card (P2)
- [x] #13 Spec phase: User Story 5 — the ? guardian section (P3)
- [x] #14 Spec phase: Polish & Cross-Cutting Concerns
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Reorient 2026-07-25 rescope (docs/design/reorient-2026-07-25-ui.md, D5/D9/D2): report card renders as a guardian-console card at natural stopping points (run end / pause / exercise resolution; badges between) and inside the postmortem takeover (TASK-127) — never a mid-run interruption. The stage's grantable verbs + one example ask per verb must also be reachable from the deterministic ? floor (D9 guardian section), not only via the tutor. All new fiction strings are skin-token-resolved (D2 — token contract from TASK-121 ships first). Render surface page authored in TASK-123 before build.

Model tier: Opus 4.8 (spec-implementer, model=opus). Rubric: guardian turn-pipeline/prompt-composition (injection-adjacent) + new llm route kind, cross-package — senior tier per constitution Principle V and the runbook Lane 3 assignment. Standing resolutions recorded in spec: report card = checklist + attribution note (one artifact); tutor guide = compiled game substrate, not a player skill (stage-3 lock untouched). DISPATCH GATED on TASK-121's merge (skin contract; console seam already merged via #87).

OPERATOR DECISION (2026-07-25, team review): SPEC RENUMBER. Two spec dirs were both numbered 059 on main — specs/059-grounded-feedback (this task) and specs/059-metatron-survival-autonomy (TASK-111). TASK-111 KEEPS 059 (it was actively being worked in the MVLS session); THIS task's spec directory renumbers.

CAUTION — the review recommended 060, but 060 IS NO LONGER FREE: specs/060-village-lens, 061-conversation-loop-damper and 062-instinct-yields were all claimed while the review was running. Next free number is 063. Claim it by creating the directory, not by looking — see the root-cause note below.

Root cause (review finding): spec-number allocation is a read-then-write race with no lock. Two sessions checking origin/main:specs/ within the same minute both see the same max and both claim next. It has now failed FOUR times in one day (commit f3b0842 'renumbered off 055/058 collisions', plus this 059 pair). The prescribed mitigation in both runbooks ('check origin/main:specs/ before claiming an NNN') CANNOT work. Fix belongs in check-merge-drift.mjs worktree mode — the takenSpecNumbers() helper (:1102-1123) already computes the right thing, it just runs too late.

Note also: this task's spec is already written against 'internal/guardian' (post-TASK-121 rename) while TASK-111's is written against 'internal/metatron' — the queue is spec'd against two names for one package until 121 lands.

Spec renumbered 059 → 063: the MVLS session's merged spec 059-metatron-survival-autonomy (PR #90) claimed the number first on main; renumbered per collision doctrine (drift-check catch, 2026-07-25).

Dispatched (UI-sweep orchestrator, handoff 2026-07-25b step 3): spec-implementer on Opus 4.8 per recorded rubric; worktree .worktrees/task-115 fast-forwarded to 9386e6a before dispatch. Gates met (TASK-121 skin contract via PR #94; console seam via PR #87). Parallel with TASK-127 (Sonnet), which ships the shared reportCardView renderer this task consumes — rebase-reconciliation round budgeted when 127 merges first. Implementer warned of pre-existing red TestCatalogSweep on main (TASK-140 hotfix in flight).

Implementation complete (Opus 4.8 spec-implementer): 10 commits, tip 031efff (incl. orchestrator's post-rebase pin fix); all 17 spec tasks done; PR #100 open, gated at the planning tier — design gate green, merge-drift pr green, race suite green except the pre-existing TestCatalogSweep red that PR #98 fixes. Merge queue: #98 → 119 → #99 (127) → #100 (this). Five deviations reviewed and accepted, notably: explain is Effect:Read (contract's zero-cost clause wins over its literal expressive-class wording); explain added to stage1CeilingTools (read-only, tutor's home stage — recorded on D9 page + stage tests); stagesLadder relocated to internal/world (tui+CLI single source). Seam status: attribution note ships standalone behind consoleCard; checklist-card prepend in rebuildConsoleCards is the single reconciliation point when #99 merges; the {checklist-only, both} render cases land at that rebase. Done flip held on PR #100 merge.

spec-bridge sync: Setup: 1/1 · Foundational — the shared data source: 2/2 · User Story 1+2 — explain + tutor-lane doctrine (P1): 3/3 · User Story 3 — tutor guide (P2): 2/2 · User Story 4 — report card (P2): 5/5 · User Story 5 — the ? guardian section (P3): 1/1 · Polish & Cross-Cutting Concerns: 3/3 — status In Progress → Done

Merged via PR #100 (squash bdb0686) after the reconciliation round over 119+127: union-resolved tui.go applyEvent, help.go section numbering (ceremonies §4, guardian §5), and two design pages; deferred T012 {checklist-only, both} cases completed against 127s merged renderer (checklist card prepends the attribution note in rebuildConsoleCards; stopping-point gate proven by TestReportCardChecklistStoppingPointGate). Human ACs #1-7 on merge evidence: read-only registry-derived explain w/ neutrality suite (zero charge/events/faith, rubric-hygiene sweep); tutor guide + preset (composition-order + byte-identity tests); report card cites event-seq evidence on KindReportCard cheap chain sharing explains data source; console card at stopping points + morgue epilogue, never mid-run; D9 ? section w/ ceiling verbs + example asks; all new strings skin-tokened (completeness test). Full -race suite green at merge; design pages re-pinned to squash on main.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
All spec tasks complete (Setup: 1/1 · Foundational — the shared data source: 2/2 · User Story 1+2 — explain + tutor-lane doctrine (P1): 3/3 · User Story 3 — tutor guide (P2): 2/2 · User Story 4 — report card (P2): 5/5 · User Story 5 — the ? guardian section (P3): 1/1 · Polish & Cross-Cutting Concerns: 3/3). Derived Done by spec-bridge sync.
<!-- SECTION:FINAL_SUMMARY:END -->

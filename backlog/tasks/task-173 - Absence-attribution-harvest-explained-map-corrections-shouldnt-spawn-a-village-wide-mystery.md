---
id: TASK-173
title: >-
  Absence attribution: harvest-explained map corrections shouldn't spawn a
  village-wide mystery
status: In Progress
assignee: []
created_date: '2026-07-30 16:41'
updated_date: '2026-08-03 13:01'
labels: []
dependencies: []
ordinal: 141000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
When a villager finds a remembered tree or rock gone, nothing lets them (or the narrator) infer the mundane explanation — that a neighbor harvested it — so ordinary resource consumption reads as an unexplained phenomenon and can take over the story.

As a player, I want real mysteries to stand out, not be drowned by my villagers panicking about firewood their neighbors chopped.
As a villager in the game, when I hear Cedar has been felling trees all week, arriving at a stump should connect those dots instead of feeding cosmic dread.

Evidence (playtest-1, 29 game-days): the dominant chronicle thread all 29 days was a "vanishing landscape" horror storyline. Cross-check: ALL 780 distinct "vanished" locations match an agent.chopped/agent.quarried event exactly — zero genuine anomalies. 2,932 agent.map_corrected events (~100/day, never declining) each fed the narrative, while chop-rumors were simultaneously circulating socially (social.rumor_told, social.place_told).

Scope note: spec 097 (perception of absence — dedup, disconfirmation decay) merged after this run. First step is a v6 re-run of the same scenario to measure what 097 already absorbs; the remaining gap is attribution — grounding a correction against known harvest activity (own memories, rumors) before it earns mystery-grade salience/narration.

Spec: specs/110-absence-attribution
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A v6 re-run of the playtest-1 scenario is measured: rate of map corrections narrated as anomalies, before/after comparison recorded on this task
- [x] #2 A map correction explainable by known harvest activity (witnessed or rumored) is attributed as mundane and does not earn mystery-grade narrative weight
- [x] #3 Genuinely unexplained absences still surface as noteworthy (the guardian's real mysteries are not suppressed)
- [ ] #4 Spec phase: Ledger and classifier
- [ ] #5 Spec phase: Coalesced narration
- [ ] #6 Spec phase: Prompt and telemetry
- [ ] #7 Spec phase: Evidence
- [ ] #8 Spec phase: Grounding and PR
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
sweep (runbook docs/design/playtest-1-findings-sweep-runbook.md): measurement soak started 2026-07-30 — world ~/.promptworld/measure/task-173-measure-1 (scenario cold-dawn, seed 46103 pinned by the exercise, teaching+tutor, llm.json mirrored from playtest-1), current main binary (ef115a6a), 16x. Measures what spec 097 absorbs: map-correction rate + share of chronicle entries narrating absence, vs playtest-1 baseline (2,932 corrections/29 days, dominant storyline). Scope checkpoint follows measurement.

RE-OPENED 2026-08-02 on the evidence its own Final Summary named: 'if a month-scale run ever resurfaces an absence storyline, re-open with that evidence.' Operator reported still seeing the vanishing-landscape effect; measured and confirmed. The 4.2-game-day window the drop decision rested on was too short — the storyline reappears past it.

EVIDENCE — two independent worlds on current main, both showing a named absence storyline in the chronicle:
- soak-world, 12.02 game-days, local gemma4:12b-mlx: chronicle carries a **[the-missing-trees]** storyline; 18 of 90 chronicle entries absence-themed. Sample narration: 'This pattern of vanishing landmarks continued as other villagers joined the search, finding nothing but empty spaces where wood once stood'; 'The forest seemed to be emptying themselves'; 'Birch's insistence on supernatural occurrences has cast a shadow over the woods.'
- soak-qwen, 5.69 game-days, local qwen3.6:latest (the new spec 109 default): chronicle carries **[the-disappearance-of-res...]**; 15 of 51 entries absence-themed. NOT model-specific — reproduces across both providers.

AC#2 IS THE ONE FAILING, and decisively: in soak-world, 969 of 972 map_corrected 'gone' entries (99.7%, over 352 distinct locations) match an agent.chopped/agent.quarried event at EXACTLY those coordinates. Only 3 entries in 12 game-days are genuinely unexplained. So essentially every correction is harvest-explained mundane resource consumption, and it is still earning mystery-grade narrative weight — which is precisely what AC#2 forbids. Note the villagers themselves partially reach for the mundane explanation socially ('Birch accuses Sage of cutting trees based on suspicious tracks'), so the social attribution layer is doing something; the failure is downstream, at storyline/narration salience.

WHAT IS NOT BROKEN — do not re-litigate these: TASK-159/spec 081 clearly worked at the memory layer. Absence-flavoured memories are 1,234 of 18,283 = 6.7% of all memories, against TASK-159's 75% showstopper baseline (346/461). Correction rate is 69.3/game-day against playtest-1's ~101/day with a comparable harvest rate. The problem is no longer that villagers drown in loss memories; it is that the storyline/narrator machinery amplifies a 6.7% minority signal into a dominant named thread. That reframes the remaining work from memory formation to narrative salience — the attribution seam this card originally scoped, now with a sharper target.

Worlds preserved for whoever picks this up: /Users/evanstern/.claude/jobs/ca35de11/tmp/soak/soak-world (stopped, 12 game-days) and soak-qwen. Query note: these world.db files are WAL-mode and sqlite3 -readonly fails on them once the daemon has stopped and no -wal file remains; open without -readonly.

SWEEP CLAIM 2026-08-02 (/pdlc:sweep on TASK-173, runbook docs/design/task-173-absence-attribution-runbook.md).
Spec 110 claimed (specs/110-absence-attribution), branch task-173-absence-attribution, worktree .worktrees/task-173.
AC#2 and AC#3 unchecked: they were ticked by the 2026-07-30 measurement close whose 4.2-game-day window the re-open disproved. They are re-earned by evidence on this branch, not carried over. AC#1 stays ticked - the measurement genuinely happened.
MODEL TIER: Opus (claude-opus-5) via .claude/agents/spec-implementer-opus.md; fallback claude-opus-4-8 on subscription unavailability. Rubric lines fired: (1) internal/mind orchestration - the absorb driver and the narrator driver; (2) doctrine-adjacent behavior change - determinism doctrine specs 092/094 govern whether attribution may be emitter-computed, and narrative salience is player-facing behavior; (3) a prior attempt shipped a live defect - the measurement-only close did not survive a longer soak. The model that actually served is recorded per dispatch below.

RUNBOOK AMENDMENT 2026-08-02 (operator-decided at the Phase 4 boundary): the evidence bar's ROUTE changed from a fresh live soak to replay + re-narrate of the preserved 12.02-game-day soak world's own event log. The WINDOW is unchanged (same 12.02 game-days, same events that re-opened this card), so this is not a softening: the comparison becomes controlled - before and after differ only by this branch's diff, where a fresh soak would confound the change with run-to-run variance on a different seed path. Operator was offered replay / fresh soak / both, and chose replay. Recorded limitation, not resolved: replay exercises the narrator's INPUT faithfully but re-runs its OUTPUT against a live model, so SC-001 is evidence about the same chapters rather than a fresh world's emergent dynamics. The re-open clause stands - if a later month-scale live run resurfaces an absence storyline, re-open again.

PROGRESS: phases 1-3 complete on task-173-absence-attribution (spec 110). Phase 1 harvest ledger + classifier (f8bc0c1d); phases 2+3 coalesced narration, narrator prompt marking, telemetry (a57920e0). Model that served all dispatches: claude-opus-5 via .claude/agents/spec-implementer-opus.md. go build, full go test ./... , and go test -race ./internal/mind/... all green as of ad4e5030.

MEASURED ROOT CAUSE (new, and it sharpens the card's own framing): the narrator never sees the memory layer's 6.7% absence share. It sees md.narrLines, the per-chapter 120-line buffer, where agent.map_corrected lines are the MAJORITY of every full day chapter - median 57%, peak 68% - and five of twelve day chapters OVERFLOWED the ring, evicting builds, gifts and assemblies while corrections survived by volume. A model told 'group by storyline, not by hour' and handed a list that is more than half 'found it gone' is summarising its input correctly. Also confirmed independently: 830 of 833 corrections match a chop/quarry at exactly those coordinates, every match within 3 game-days (716 same-day), exactly 3 genuine anomalies in 12.02 game-days.

PHASE 4 EVIDENCE (replay + re-narration over the preserved soak worlds; artifact specs/110-absence-attribution/evidence.md, harness internal/mind/replay_evidence_test.go). The runbook's four required measurements:

(a) ABSENCE-THEMED CHRONICLE ENTRIES: 32-50% before -> 3-7% after (gemma run A 32%->6%; qwen run A 41%->7%; qwen run B 50%->3%).
(b) NAMED ABSENCE STORYLINE SLUG: present in 4 of 4 before-runs, carrying 19-43% of each run's entries (the-missing-trees 15/56, vanishing-landmarks 19/61, the-disappearance-of-res 7/37, vanishing-resources-and 12/28). NONE in any of the 3 after-runs. This is the decisive measurement.
(c) HARVEST-EXPLAINED SHARE OF CORRECTIONS: run A 833 corrections, 830 attributed (99.6%), 3 unexplained, ZERO false attributions. Run B 457/452/5, zero false. Precision judged against ground truth built from the raw log, not from the ledger under test.
(d) GENUINELY-UNEXPLAINED ABSENCES STILL SURFACE: run A's 3 are all one tile - a cache of goods at (61,6) with no chop or quarry anywhere in 132,299 events - and three villagers went for it on day 9, each keeping its own line in the pre-110 wording verbatim. Run B's 5 likewise. AC#3 holds.

SC-002 (volume): corrections fall from 48-65% of each day chapter to 1-2%; attributed corrections contribute exactly one line per chapter across all 24 run-A and 13 run-B chapters; ring overflow on run A 5 chapters -> 0, recovering 130 non-correction lines that had been silently evicted. SC-004 100% precision. SC-006 both models. go build, go test ./... and go test -race ./internal/mind/... all green after the origin/main merge-in (58 commits, conflict-free - no main commit had touched internal/mind).

SC-001 THRESHOLD RULING (operator, 2026-08-03): SC-001 split - storyline clause met outright, numeric clause ('at most 5%') missed on two of three after-runs at 6% and 7%. The implementer escalated rather than rounding it. Two facts bore on the ruling: the 5% threshold was derived from a baseline this spec miscomputed (18/90=20% mixed an entries numerator with a lines denominator; like-for-like it is 18/56=32%), and of the seven residual keyword matches across all after-runs NONE concerns a remembered map feature being gone - on what the criterion was written to catch the after-runs measure 0 of 122 entries. Operator ruled: tick AC#2, the storyline clause governs, since AC#2's words are 'does not earn mystery-grade narrative weight'. Stricter alternatives (tune until the letter passes; require a fresh 12-day live confirmation soak) were offered and declined. The miss is recorded in spec.md and evidence.md, not erased.

RECORDED LIMITATION: replay exercises the narrator's input faithfully but re-runs its output against a live model, so this is evidence about the same 12.02 game-days rather than a fresh world's emergent dynamics, and it cannot exercise the narrator-to-villager feedback loop. The original re-open clause stands: if a later month-scale live run resurfaces an absence storyline, re-open with that evidence.

PR 160 opened (draft) 2026-08-03: https://github.com/evanstern/promptworld/pull/160 - branch task-173-absence-attribution. All five spec-110 phases complete. Grounding rides the PR per spec 069: new docs/wiki/absence-attribution.md split out of chronicle.md, 11 notes re-pinned against read diffs, CAPSULES.md regenerated, INDEX entry added, 3 player pages regenerated (16 fresh / 0 stale). Gates green after the final origin/main merge-in: go build, go test ./..., go test -race ./internal/mind/... (516.6s), merge-drift pr mode PASS with no findings. MERGE WITH --merge, NEVER SQUASH: the branch carries in-branch wiki re-pins and a squash rewrites the hashes they reference.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Resolved by measurement — no build needed (operator ruling 2026-07-30 at the sweep checkpoint). A v6 re-run of the playtest-1 scenario (task-173-measure-1: cold-dawn, seed 46103, teaching+tutor, llm.json mirrored, 4.2 game-days at 16x, world preserved under ~/.promptworld/measure/) shows spec 097 absorbed the misattribution symptom: map corrections down 62% over matched days (212 vs 554, days 0-3) despite a HIGHER harvest rate (~34 vs ~23 chops/day); ~2,000 grounded observations/day feeding belief reconciliation (28 reinforced, 16 revised); and the chronicle carries ZERO absence storyline (themes: rumor, labor, conflict, gru — one absence-flavored word total, vs playtest-1's dominant vanishing-landscape thread from day 1). AC#2/#3 hold as observed outcomes: harvest-explained corrections no longer earn mystery-grade narrative weight, while 097's disconfirmation salience bump keeps genuine surprises noteworthy. Caveat recorded: 4-day window vs playtest-1's 29; if a month-scale run ever resurfaces an absence storyline, re-open with that evidence. The attribution-seam build is dropped, not deferred.
<!-- SECTION:FINAL_SUMMARY:END -->

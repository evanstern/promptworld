---
id: TASK-105
title: Per-turn context grounding — audit and intent-driven context assembly
status: Done
assignee: []
created_date: '2026-07-25 02:41'
updated_date: '2026-07-25 13:58'
labels:
  - goal-quality
dependencies: []
references:
  - 'https://github.com/evanstern/promptworld/pull/77'
priority: high
ordinal: 14000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Direction C from spike TASK-101 — Evan: '100% on this, almost more important than A/B.' Two parts. (1) AUDIT: produce a durable, complete inventory of exactly what each villager receives in context per thought (system prompt + userPrompt blocks, internal/mind/prompt.go:73-145) and what is notably absent. Known gaps: own last/current intent + source (LastGoal is TUI-only), need TRAJECTORIES (level+direction, not just level), active-plan echo so a thought continues rather than restarts. (2) REDESIGN: assemble context efficiently and with intent — self-history block, trajectories, plan echo, plus richer grounding via relevant-memory retrieval and selective journal-entry stuffing (dovetails with the embedding-memory retrieval work, spec 042 / TASK-98). Budget note: thoughts run 4-5 loop turns max, so per-turn context stuffing is affordable on a moderately hostable model. Non-trivial: full Spec Kit before implementation.

Spec: specs/043-context-grounding
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Written per-turn context inventory (present vs absent) exists as a durable artifact
- [x] #2 Self-history, need trajectories, and active-plan echo added to the decision prompt
- [x] #3 Relevant-memory/journal retrieval feeds the prompt with measured token budget
- [x] #4 Spec phase: Setup
- [x] #5 Spec phase: Foundational (blocking prerequisites)
- [x] #6 Spec phase: US5 — Operators can see exactly what an agent knew (P1) 🎯 co-MVP
- [x] #7 Spec phase: US1 — An agent knows what it was just doing (P1) 🎯 co-MVP
- [x] #8 Spec phase: US2 — An agent feels which way its needs are moving (P2)
- [x] #9 Spec phase: US3 — An agent continues its plan instead of restarting it (P3)
- [x] #10 Spec phase: US4 — What an agent remembers is chosen for the moment (P4)
- [x] #11 Spec phase: Polish & cross-cutting
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Synergies (2026-07-24 board pass): TASK-110 prunes dead verbs from the roster — shrinks the tool surface this task's context budget pays for; do 110's roster prune before or with the context redesign. Relevant-memory retrieval leans on the embedding work: TASK-98 (in progress, spec 042) provides record-at-emission vectors + relevance term; TASK-102 (embed preflight warning bug) should land so embedding-path signal is clean.

Spec drafted and committed to main (996d503): specs/043-context-grounding/spec.md — 5 user stories (P1 self-history + P1 context inventory audit, P2 need trajectories, P3 plan echo, P4 relevance retrieval under token budget), 10 FRs, 7 measurable SCs incl. flip-rate reduction vs the world-01 baseline. Requirements checklist passes; zero NEEDS CLARIFICATION (defaults in Assumptions). Next: speckit-plan → speckit-tasks → spec-bridge:link, then delegated implementation.

Implementation dispatch (constitution Principle V): Foundational + US1 slice (T001-T012, T013 stretch) → Opus 4.8 spec-implementer. Rubric: touches internal/sim reducer state and internal/mind orchestration (doctrine-adjacent, cross-package), shadow-invariant byte-identity constraint — senior tier required. US5 (T006-T007, wiki note + capture) → Sonnet, dispatched after the code slice lands on the branch.

Foundational + US1 slice landed on task-105-context-grounding (commits 9de0665, d2bbd7f, 1e1b97d; Opus 4.8 implementer). Assembler + byte-identity wrap, cog.thought sizes, IntentRecord ring + 5 reducer arms, self_history block. Whole-repo tests, vet, gofmt green; shadow invariant holds. Gate-checked by orchestrator (targeted sim/mind tests re-run). T013 stretch: SC-004 confirmed via read-only probe of world-01 (Sage tick 265,864 ring shows instinct forage + alternation); committed harness deferred to T024/T027 pattern (needs replay-to-tick helper + env-guarded skip). Deviations accepted: intent_rejected now state-mutating per data-model (split from cog.* no-op arm); IntentRecord ticks KEEP in rebase taxonomy; self_history always renders (empty state line); future-dating line owned by frame block.

US2+US3 slice landed (commits 94978fc, 4b4db49; Opus 4.8). Need trajectories (anchor window 1800, deadband ±10, SHIFT taxonomy) + plan_echo block (guards in plain words, no stale echo). Full suite green; gate-checked. Deviations accepted and recorded as data-model addenda: *Needs pointer (json byte-identity), NeedsAnchorTick=SHIFT (overturned orchestrator's KEEP hypothesis with correct doctrinal reasoning). Original AC#2 (self-history + trajectories + plan echo in prompt) now true. Next: US4 (Opus), then US5 capture with all blocks present, then polish.

US4 slice landed (653f2a7, 41b0502; Opus 4.8). Journal term-match excerpts (≤2×300 runes, deterministic), memories floor(4)/serendipity split via a shared annotated selection core (byte-identical by construction, drift-guard test), planted-memory relevance 10/10, hermetic budget-fit 1600 thoughts ≥99%. T023 live multi-day measurement + T007 capture folded into US5 dispatch. WATCH ITEM (implementer flag): journal term match keys on raw goal names (goto_warmth) which rarely appear in free text — worst-need names carry the matching; if live recall is weak, split goal tokens on _ or take R5's embed-at-write follow-on.

US5 slice landed (7374d1b; Sonnet): docs/wiki/decision-context.md inventory (10 blocks + deliberate absences, pinned 41b0502) + INDEX entry; live evidence: SC-001 contract-vs-capture MATCH over 1,055 planner thoughts; SC-005 100% within 2000-token budget (PromptBytes min 623 / median 2426 / max 3165 ≈ 156/607/791 approx-tokens), 0 live drops (drop mechanics covered by unit tests); run ctx-043-check seed 424301, 2.017 game-days at 32x, cogito:3b (gemma4:12b-mlx not pulled — deviation recorded in evidence). Throwaway world + binary cleaned up. Note: the agent stalled waiting for an untracked daemon; orchestrator resumed it via message — future dispatches with long background runs should poll in-agent, not await notification. Remaining: T013 harness (absorbed into T024), polish T024-T025, PR; post-merge T026-T027.

Polish slice landed (d76d00b; Opus 4.8): TestContextReplayByteIdentical (snapshot + genesis replay, -race), replayToTick helper + env-guarded TestSageThrashWindowContextReplay (SC-004 evidence committed — Sage tick 265,864 self_history shows instinct/planner alternation verbatim), full suite + vet + gofmt green, merge-tree vs origin/main conflict-free. US1 phase now 6/6 (T013 done). PR OPENED: https://github.com/evanstern/promptworld/pull/77 — one task, one PR. Remaining post-merge: T026 wiki re-pin + player docs, T027 SC-007 flip-rate run; then task Done via spec-bridge:sync.

T026 complete: wiki re-pin (21 notes reviewed across 4 parallel batches + morgue.md, all 43 notes fresh at dee5f4b, plan+freshness gates green) and player docs refreshed 9/9 (commit 23f744c, reconciled with concurrent session's parallel regen). Only T027 (SC-007 flip-rate) remains — live run in progress (~4 game-days at 32 ticks/s).
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Shipped via PR #77 (merge 72125c8): villager thoughts now carry self-history (intent ring with honest sources), need trajectories, plan echo, and budgeted relevance-selected memories/journal excerpts, assembled deterministically from a normative 10-block contract with per-thought size telemetry — plus the freshness-gated docs/wiki/decision-context.md audit. All 27 spec tasks done; all evidence committed: SC-001 contract match (1,055 live thoughts), SC-004 thrash-window replay (committed env-guarded test), SC-005 100% budget fit, SC-007 flip-rate — Sample A (cogito:3b, 32x, ~1 game-day) worst agent 5.06 flips/game-day (−93% vs world-01's 72); Sample B (gemma4:12b-mlx, 4x, live t2) worst 33.08 (−54%); both clear the ≤36 bar, with honest caveats (sub-1-day spans, B narrow at 8% margin, TASK-103/104 not landed). Wiki re-pinned (43 notes fresh) + player docs 9/9. Follow-on carded: full-length SC-007 re-measure after 103/104 land.
<!-- SECTION:FINAL_SUMMARY:END -->

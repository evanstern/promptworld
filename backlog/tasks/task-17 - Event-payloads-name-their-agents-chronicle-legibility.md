---
id: TASK-17
title: Event payloads name their agents (chronicle legibility)
status: Done
assignee: []
created_date: '2026-07-19 15:56'
updated_date: '2026-07-27 04:11'
labels:
  - events
  - tui
dependencies: []
ordinal: 10000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Out-of-sim consumers of the event log (webhook sinks TASK-18, exported logs, external tools) see agents only by index — {"agent":2}, {"from":0,"to":3} — and cannot resolve names without a replica. Re-grounded 2026-07-22: the TUI is no longer the motivating case — eventRow is gone; the chronicle now resolves names post-hoc via formatChronicleLine(e, names) / m.agentNames() (internal/tui/views.go, chronicleRawBody), and narration (TASK-11, Done) ships in chronicleNarratedBody. But that is exactly the post-hoc-lookup pattern this task makes unnecessary at the format level; the primary driver is now external/out-of-sim readers, pairing this task with TASK-18. Requirement: the log format itself carries names, enforced at emission. Approach: an AgentRef that marshals {id, name} (names are fixed per agent, so the denormalization is replay-safe) used for every agent-referencing payload field across sim/mind/social emitters (agent, subject, speaker, listener, from/to, creditor/debtor, witnesses). Enforce mechanically: payload constructors take refs, plus append-time validation or a test sweep over all registered payload types that rejects agent-bearing payloads lacking names. Define the back-compat story for historic events without names (reducer accepts both; renderers fall back gracefully). Bonus once landed: the TUI lookup layer can shrink.

Spec: specs/086-agent-named-payloads
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Every recorded event payload that references an agent carries both the index and the agent's name in its JSON
- [x] #2 The TUI chronicle shows agent names with no replica/post-hoc lookup
- [x] #3 The format is enforced mechanically (typed ref + append validation or exhaustive payload test), not by convention
- [x] #4 Replay of pre-change worlds still works; back-compat behavior is documented
- [x] #5 Spec phase: Setup
- [x] #6 Spec phase: Foundational — the type, the catalog, the rails (blocks all user stories)
- [x] #7 Spec phase: US1 — the census migration (P1) 🎯 — batched, compiler-driven
- [x] #8 Spec phase: US2 — enforcement flipped fully on (P1)
- [x] #9 Spec phase: US3 — back-compat proven and documented (P1)
- [x] #10 Spec phase: US4 — chronicle names + jump-to-source hit rate (P2)
- [x] #11 Spec phase: US5 — the reverse-jump rider (P3)
- [x] #12 Spec phase: Polish, grounding, gates
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Drift audit 2026-07-23: still real. Payloads carry indices only (internal/sim/agents.go:668-725, e.g. TalkedPayload uses A/B ints); no AgentRef type exists anywhere; TUI post-hoc lookup intact (tui.go:1076-1085 agentNames, views.go:910-914 formatChronicleLine, grammar.go:166).

Reorient 2026-07-26 (board move 10): reframed upward — agent-named payloads raise chronicle jump-to-source's locatable-event hit rate (resolveSubject), making this village-lens completion, not just chronicle hygiene.

Rider (reorient 2026-07-26 delta, operator-placed 2026-07-26: 'rider is fine'): REVERSE JUMP — strip glyph / roster row → camera center on the map. The delta's one net-new unscheduled rec; homes here because this task already raises jump-to-source's locatable-event hit rate (resolveSubject) as village-lens completion — reverse jump is the same lens's other direction. Ships with a mouse-parity oracle entry per the TASK-154 gate when it lands.

Sweep claim (runbook docs/design/faith-directives-sweep-runbook.md, signed-off 2026-07-26): spec 086-agent-named-payloads. Tier: Opus 4.8 — repo-wide payload migration (AgentRef across every agent-referencing emitter) + mechanical enforcement + back-compat replay. Deliberate Lane C tail: sweeps the sweep's own new payloads (directive.*, faith.*/prophecy.*, sim.neglect_detected, stranger.*) in one pass. Carries the operator-placed reverse-jump rider (strip glyph/roster row → camera center) with its mouse-parity oracle obligation.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Merged via PR #127. AgentRef ships: every agent-referencing payload field on new emissions carries {id,name} (dual-shape unmarshal accepts legacy ints forever); 66 in-place + 4 mirrors + faith additive ref; enforcement at both emission rails + the PayloadCatalog sweep; AgentRef never reachable from State (reflection-pinned), so pre-086 replay is byte-identical (proven on a merge-base-generated 7684-event fixture). Chronicle names render from log bytes alone; resolveSubject locatable types 79→86. Reverse-jump rider shipped (strip glyph/roster click + J → camera center; graves for the dead) under both mouse-parity oracles, amending standing resolution 1. Four latent inline int-decoders in mind/scribe/guardian closed. Tier: Opus 4.8 as recorded. Sweep tail complete.
<!-- SECTION:FINAL_SUMMARY:END -->

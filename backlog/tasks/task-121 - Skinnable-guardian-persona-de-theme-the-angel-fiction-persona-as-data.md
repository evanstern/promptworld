---
id: TASK-121
title: 'Skinnable guardian persona: de-theme the angel fiction, persona as data'
status: In Progress
assignee: []
created_date: '2026-07-25 06:20'
updated_date: '2026-07-25 17:40'
labels:
  - learning-game
  - design
dependencies:
  - TASK-64
priority: high
ordinal: 92000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
--------------------------------------------------
Operator pivot (2026-07-25, raised during the TASK-68 spec session): keep the agent-with-tools structure, rules, and mechanics exactly as they are; tone down the overt Anglo-Judeo-Christian imagery (Metatron, angel, miracles-as-scripture) across code and docs; make the fiction layer skinnable data. Ratified decisions: (1) SKIN DEPTH — a skin is a data bundle: agent display name, all user-facing fiction strings (TUI labels, chronicle flavor terms, curriculum stage display names), and a persona-voice text composed into the agent prompt BENEATH the fixed frame; the fixed-frame invariants are never skinnable (injection-soundness doctrine, spec 021); mechanics/tools/costs/rules identical across skins. (2) DEFAULT — secular-mythic guardian (final name Guardian/Warden/Steward picked in spec), folk-tale tone not scripture; interventions become workings/acts; charter stays charter. (3) CURRICULUM — TASK-68 stages are neutral ids + concept descriptions in the substrate; skins supply display identities. (4) SEQUENCING — specs 044/045 land under current naming; this task sweeps code (incl. the metatron package), prompts, TUI, player docs, and spec/wiki prose in its own PR after. Custom skins are user-authorable per world (e.g. spider-man) — authoring one is itself a prompt-engineering exercise (ties to stage 3). Non-trivial: full Spec Kit before implementation.

Spec: specs/052-skinnable-guardian
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Spec written and linked via spec-bridge covering skin bundle format, prompt composition, the default guardian skin, and the rename sweep
- [ ] #2 No user-facing Metatron/angel imagery remains in the default experience: prompts, TUI strings, CLI output, player docs (internal identifiers swept or aliased per spec decision)
- [ ] #3 Skin bundle = name + fiction strings + persona voice; fixed-frame invariants provably not overridable from any skin
- [ ] #4 Default secular-mythic guardian skin ships; mechanics byte-compatible across skins (same events, costs, rules)
- [ ] #5 A custom skin is loadable per world; one example alternate skin in-repo proves the format
- [ ] #6 Post-sweep: wiki re-pinned (wiki-update) and player docs regenerated
- [ ] #7 Skin-token contract published before 115/117 implementation; both consume it
- [ ] #8 Sweep covers help.go, footer hints, stagesLadder, lesson strings, player-docs page names, design-doc mockups
- [ ] #9 Skin boundary = guardian/systems tab split; systems content never skinnable
- [ ] #10 Spec phase: Setup
- [ ] #11 Spec phase: Foundational — the skin substrate
- [ ] #12 Spec phase: User Story 1 — The skin-token contract exists (P1) 🎯 Lane-3 unblock
- [ ] #13 Spec phase: User Story 2 — The default experience is de-themed (P1)
- [ ] #14 Spec phase: User Story 3 — A custom skin is a per-world data bundle (P2)
- [ ] #15 Spec phase: User Story 4 — The internals stop lying (P3)
- [ ] #16 Spec phase: Polish & Cross-Cutting Concerns
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Reorient 2026-07-25 rescope (D2/D10): the skin-token CONTRACT ships before TASK-115/117 write any new fiction literal — those tasks consume the lookup from day one (sequencing inversion fix). Enumerated sweep sites beyond the original framing: internal/tui/help.go static tables (5 Metatron literals), footer hints, cmd/promptworld/stages.go stagesLadder identities, TASK-117 lesson strings, docs/player/ page names, and docs/design/tui mockups/labels (the design doc renders all fiction strings as skin tokens — TASK-123). The guardian/systems tab split (D10, TASK-125) makes the skin boundary a file boundary: skins may touch the guardian tab, never the systems tab.

Model tier: Opus 4.8 (spec-implementer, model=opus). Rubric: cross-package architectural slice; prompt composition beneath the spec-021 injection-soundness doctrine (fixed-frame adjacency = doctrine-adjacent behavior change); repo-wide rename with serialization-freeze hazards — senior tier per constitution Principle V and the runbook Lane 1 assignment. Dispatched by UI-sweep orchestrator.
<!-- SECTION:NOTES:END -->

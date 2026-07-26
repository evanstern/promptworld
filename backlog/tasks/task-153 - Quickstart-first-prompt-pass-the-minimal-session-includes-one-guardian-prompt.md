---
id: TASK-153
title: 'Quickstart first-prompt pass: the minimal session includes one guardian prompt'
status: Done
assignee: []
created_date: '2026-07-26 17:57'
updated_date: '2026-07-26 20:50'
labels:
  - player-docs
  - content
dependencies: []
references:
  - docs/design/reorient-2026-07-26-ui.md
ordinal: 123000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Reorient 2026-07-26 decision 7. getting-started.html's walkthrough never has the player prompt — a first session that never prompts teaches watching. Add an 'ask your guardian one thing' step (sample ask from the skin.guardian.example_ask.* token family, the same family the ? guardian section renders) and give each stage page a short first-session do-this-then-this block. Content-only; the player-docs skill is the home.

Spec: specs/079-quickstart-first-prompt
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 getting-started.html includes a first-prompt step sourced from skin.guardian.example_ask.*
- [x] #2 Each stage page carries a first-session do-this-then-this block
- [x] #3 Spec phase: Setup & baseline
- [x] #4 Spec phase: Foundational — the editorial contract (blocks all page edits)
- [x] #5 Spec phase: User Story 1 — getting-started first-prompt step (board AC #1) 🎯
- [x] #6 Spec phase: User Story 2 — stage-page first-session blocks (board AC #2)
- [x] #7 Spec phase: Gates & polish
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Sweep claim (runbook docs/design/reorient-2026-07-26-sweep-runbook.md): spec 079-quickstart-first-prompt. Tier: Sonnet — content-only via the player-docs skill. Full Spec Kit per the runbook (content pass ≠ the constitution's surgical-fix trivial exemption). Sequenced last: docs/player churn from Lanes A–C settles first.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Merged via PR #120 (merge commit ae938bd). getting-started.html gains the 'ask your guardian one thing' step (skin.guardian.example_ask.send_vision, byte-verbatim, stage-1 ceiling) with the skin-honesty note; four stage pages gain first-session blocks, all --scenario-free; durability guaranteed by the three-place mechanism (page + skin.md source pin + player-docs SKILL.md editorial contract, so regeneration reproduces the content). Stage-4 wording reconciled against spec 077's optional exercises; FR-006 amended to match. Tier: Sonnet as recorded.
<!-- SECTION:FINAL_SUMMARY:END -->

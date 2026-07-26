---
id: TASK-153
title: 'Quickstart first-prompt pass: the minimal session includes one guardian prompt'
status: In Progress
assignee: []
created_date: '2026-07-26 17:57'
updated_date: '2026-07-26 19:29'
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
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 getting-started.html includes a first-prompt step sourced from skin.guardian.example_ask.*
- [ ] #2 Each stage page carries a first-session do-this-then-this block
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Sweep claim (runbook docs/design/reorient-2026-07-26-sweep-runbook.md): spec 079-quickstart-first-prompt. Tier: Sonnet — content-only via the player-docs skill. Full Spec Kit per the runbook (content pass ≠ the constitution's surgical-fix trivial exemption). Sequenced last: docs/player churn from Lanes A–C settles first.
<!-- SECTION:NOTES:END -->

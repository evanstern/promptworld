---
id: TASK-168
title: 'Policy: every task opens with a brief gist plus ''As a …'' use cases'
status: To Do
assignee: []
created_date: '2026-07-29 16:09'
updated_date: '2026-07-29 16:10'
labels: []
dependencies: []
ordinal: 136000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Make it project policy that every Backlog.md task description opens with a one-or-two-sentence plain-language gist of the deliverable, followed (where applicable) by a few scene-setting use cases in "As a <role>" form. This card is written in that format and serves as the first example.

As a hu-mon scanning the board, I want to catch what a card is about in a few seconds, without decoding file paths or internal jargon.

As the spec agent writing a spec from a task I didn't author, I want the card to state its purpose plainly up front, so I churn less during specify/clarify reconstructing what the ticket is for.

As a card author (human or agent), I want a concrete format with good and bad examples, so the top of every card comes out the same shape.

Format examples (from the operator):
- Good: "As a player, when the game starts up, I want to see the map on the left and the chronicle on the right."
- Less good: "The left side of the screen should be the map and the right side should be the chronicle." (describes the artifact, not the experience)
- Bad: "The UI is bad, we should fix it" — or a wall of file/concept references a human can't easily grok.

Any accurate scene-setting role is valid: "As a player", "As a user", "As a villager in the game", "As the Gru".

Scope: this project only for now — a repo-local policy, not a praxisflux plugin change yet.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 The policy (gist-first, then 'As a <role>' use cases, with the good/less-good/bad examples) is written into a durable tracked home that card authors and the spec agent load when creating tasks or specs (e.g., the Backlog.md block of the project CLAUDE.md, or a doc it references)
- [ ] #2 The policy states when use cases apply vs. may be skipped (pure infra/bookkeeping cards may omit use cases, but never the opening gist)
- [ ] #3 The policy is scoped to this repo only — no praxisflux plugin or template changes
- [ ] #4 spec-phase guidance points at the gist section as the primary statement of intent for a task the spec agent didn't write
<!-- AC:END -->

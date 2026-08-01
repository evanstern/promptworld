---
id: TASK-182
title: >-
  User manual: reference pages for players (commands, world files,
  troubleshooting)
status: In Progress
assignee: []
created_date: '2026-08-01 23:17'
updated_date: '2026-08-01 23:39'
labels: []
dependencies: []
priority: medium
ordinal: 164001
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
<!-- SECTION:DESCRIPTION:BEGIN -->
promptworld teaches players how to play but never documents it: there is no single place that lists every command you can type, explains every file in your world folder and which settings you may change, or tells you what to do when something goes wrong. This adds three reference pages to the player documentation — a command reference, a world-files reference, and a troubleshooting guide — so a player can look an answer up instead of hunting through tutorials.

As a player, when I want to know what I can actually type, I want one page that lists every promptworld command with its options and a plain sentence about what it does, instead of piecing it together from the quickstart.

As a player, when I open my world folder and find files like llm.json, charter.md and tuning.json, I want one page that tells me what each file is for and which settings are safe for me to change.

As a player, when my villagers stop thinking or my world will not start, I want a page that matches my symptom to a likely cause and a fix, written for me and not for an engineer.

As a maintainer, when the code changes underneath these pages, I want them covered by the same freshness gate as the rest of the player docs, so a stale manual blocks a PR instead of quietly misleading players.
<!-- SECTION:DESCRIPTION:END -->

Spec: specs/108-player-manual
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 docs/player/command-reference.html documents every subcommand the binary dispatches (21 at time of writing, including the hidden pre-052 compat aliases marked as such), each with its flags, its world-argument semantics (name or path), and a plain-language description
- [ ] #2 docs/player/world-files-reference.html documents every file a player may read or edit in a world save directory — llm.json, charter.md, village_charter.md, skin.json, tuning.json, calibration.json, bundles — saying what each is for and which parts are player-editable, deferring registry-reference depth to docs/llm-providers.md by link rather than duplication
- [ ] #3 docs/player/troubleshooting.html is organised symptom -> likely cause -> fix, covering at minimum: world will not start, villagers act only on reflex, no LLM configured, model too slow (suppression/horizon), daemon appears running but is not, and format-version/migration mismatch
- [ ] #4 docs/player/index.html gains a Reference section linking the three new pages; existing pages are byte-unchanged except index.html
- [ ] #5 The player-docs skill is updated in the same PR: expected page set, page-to-source mapping table, and any editorial notes cover the three new pages
- [ ] #6 check-freshness.mjs EXPECTED_PAGES includes the three new pages, and node .claude/skills/player-docs/scripts/check-freshness.mjs --check exits 0 on the branch
- [ ] #7 Every factual claim on the new pages projects from a declared, pinned source (wiki note verified_against pin, or git log pin for plain files/specs); each page carries correct promptworld-docs:source meta tags and the canonical self-contained skeleton + CSS (no external assets, no JS)
- [ ] #8 Prose is plain-language for a non-engineer: no engineering vocabulary a player would have to decode, consistent with the existing player pages
- [ ] #9 Spec phase: Phase 1 — Grounding (read-only)
- [ ] #10 Spec phase: Phase 2 — Pages
- [ ] #11 Spec phase: Phase 3 — Nav
- [ ] #12 Spec phase: Phase 4 — Gate wiring
- [ ] #13 Spec phase: Phase 5 — Verification
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Spec Kit run complete through tasks: specs/108-player-manual/{spec,plan,tasks}.md on branch task-182-player-manual. Spec number 108 claimed on main via no-ff merge of the stub. Model tier: routine implementation tier claude-sonnet-5 via .claude/agents/spec-implementer.md — docs-and-prose work, single surface, no concurrency/architecture/doctrine trigger in the escalation rubric.
<!-- SECTION:NOTES:END -->

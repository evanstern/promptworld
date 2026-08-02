---
id: TASK-182
title: >-
  User manual: reference pages for players (commands, world files,
  troubleshooting)
status: Done
assignee: []
created_date: '2026-08-01 23:17'
updated_date: '2026-08-02 00:33'
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
- [x] #1 docs/player/world-files-reference.html documents every file a player may read or edit in a world save directory — llm.json, charter.md, village_charter.md, skin.json, tuning.json, calibration.json, bundles — saying what each is for and which parts are player-editable, deferring registry-reference depth to docs/llm-providers.md by link rather than duplication
- [x] #2 docs/player/troubleshooting.html is organised symptom -> likely cause -> fix, covering at minimum: world will not start, villagers act only on reflex, no LLM configured, model too slow (suppression/horizon), daemon appears running but is not, and format-version/migration mismatch
- [x] #3 docs/player/index.html gains a Reference section linking the three new pages; existing pages are byte-unchanged except index.html
- [x] #4 The player-docs skill is updated in the same PR: expected page set, page-to-source mapping table, and any editorial notes cover the three new pages
- [x] #5 check-freshness.mjs EXPECTED_PAGES includes the three new pages, and node .claude/skills/player-docs/scripts/check-freshness.mjs --check exits 0 on the branch
- [x] #6 Every factual claim on the new pages projects from a declared, pinned source (wiki note verified_against pin, or git log pin for plain files/specs); each page carries correct promptworld-docs:source meta tags and the canonical self-contained skeleton + CSS (no external assets, no JS)
- [x] #7 Prose is plain-language for a non-engineer: no engineering vocabulary a player would have to decode, consistent with the existing player pages
- [x] #8 Spec phase: Phase 1 — Grounding (read-only)
- [x] #9 Spec phase: Phase 2 — Pages
- [x] #10 Spec phase: Phase 3 — Nav
- [x] #11 Spec phase: Phase 4 — Gate wiring
- [x] #12 Spec phase: Phase 5 — Verification
- [x] #13 docs/player/command-reference.html documents every subcommand the binary dispatches (22 at time of writing, plus the two hidden pre-052 compat aliases marked as such), each with its flags, its world-argument semantics (name or path), and a plain-language description
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Spec Kit run complete through tasks: specs/108-player-manual/{spec,plan,tasks}.md on branch task-182-player-manual. Spec number 108 claimed on main via no-ff merge of the stub. Model tier: routine implementation tier claude-sonnet-5 via .claude/agents/spec-implementer.md — docs-and-prose work, single surface, no concurrency/architecture/doctrine trigger in the escalation rubric.

PR #154 opened (draft): https://github.com/evanstern/promptworld/pull/154 — three reference pages (command-reference, world-files-reference, troubleshooting) + index nav + player-docs skill/gate wiring. check-freshness --check: 16 fresh, 0 stale, 0 missing, 0 broken-ref. check-merge-drift pr: pass, no findings. Flag audit clean (no invented flags). Implemented by claude-sonnet-5 via .claude/agents/spec-implementer.md as planned; no escalation needed. Spec narrative corrected 21 -> 22 subcommands after the implementer flagged the mismatch with FR-001's own list.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Shipped the player's user manual as three gated reference pages in docs/player/ — command-reference.html (all 22 subcommands grouped by player intent, plus the two hidden pre-052 compat aliases), world-files-reference.html (all nine world-folder files, each labelled yours-to-edit / read-dont-edit / leave-alone), and troubleshooting.html (all seven required symptoms, each resolving to a real observable surface — a ps column, a status field, a TUI badge, a daemon-log warning). index.html gained a Reference section. The load-bearing part is the wiring: the player-docs skill and check-freshness.mjs EXPECTED_PAGES now cover the three pages, so a future change to a pinned source restales the manual and the pr gate blocks until it is re-projected in the same PR — the manual cannot silently rot. Also corrected a pre-existing staleness in the skill's mapping table (docs/wiki/metatron.md -> the guardian* family, dead since the spec 052 rename), and corrected the spec's own narrative count from 21 to 22 subcommands after the implementer flagged the mismatch with FR-001's list. Verified on merged main: check-freshness 16 fresh / 0 stale / 0 missing / 0 broken-ref; session gate reports player-docs stale=false and wiki stale=false. Spec 108. PR #154 merged with a merge commit fc8b467. Tier: claude-sonnet-5 via .claude/agents/spec-implementer.md, no escalation. Known follow-up: troubleshooting.html grounds the horizon/suppression check through the TUI badge rather than the literal promptworld status line, because cli-runtime-control.md is not a declared source for that page — adding it is a one-line change if the exact status strings are wanted.
<!-- SECTION:FINAL_SUMMARY:END -->

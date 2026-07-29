# Tasks: wiki corpus size-budget debt

**Input**: Design documents from `specs/089-wiki-size-budget/`

**Tests**: gate runs (grounding-wiki freshness, player-docs probe, merge-drift pr).

## Phase 1: Tighten (small overages)

- [X] T001 Tighten in place, dropping no fact: event-types-cognition-telemetry.md,
  tui-client-mechanics.md, tile-registry.md, mental-map-perception.md,
  executor-social-perception.md, explain-tutor-guide.md, village-lens.md,
  event-types-agent-intents.md, curriculum-ladder.md, social-fabric.md (each ≤300
  over; escalate any that resists to a split in T002 and note it).

## Phase 2: Split or tighten (mid/large overages)

- [ ] T002 Resolve ipc-protocol.md, chronicle.md, guardian.md,
  sim-state-apply-world.md (385–457 over): tighten if honest, else summary-style
  split per docs/corpus-spec.md.
- [ ] T003 Split (or justify exemption): reflex-policy.md, sim-state-apply-agents.md,
  event-types.md, sim-loop-injection-doors.md, guardian-miracle-rebase-taxonomy.md,
  guardian-designations.md, tool-registry-guardian-tools.md, tui-chronicle-feed.md,
  executor-needs-survival.md, guardian-faith.md — follow the corpus's parent/child
  idiom; children get full frontmatter (capsule ≤500, sources, honest
  verified_against = branch commit verified against).

## Phase 3: Capsules + index

- [ ] T004 Rewrite the two oversized description capsules (guardian-faith.md,
  guardian-instruction-surface.md) under 500 chars; update INDEX.md with all new
  children; regenerate CAPSULES.md via
  `node ~/.claude/plugins/cache/praxisflux/grounding-wiki/0.39.0/scripts/capsules.mjs . docs/wiki`.

## Phase 4: Downstream + verification

- [ ] T005 Player docs: run `node .claude/skills/player-docs/scripts/check-freshness.mjs --check`
  in the worktree; regenerate every page the wiki edits staled per the player-docs
  skill; probe exits 0.
- [ ] T006 Final gates: grounding-wiki freshness exits 0
  (`node ~/.claude/plugins/cache/praxisflux/grounding-wiki/0.39.0/gates/cli.mjs freshness . docs/wiki`);
  no-fact-lost spot-check recorded in the PR body (pre-change section headings →
  post-change locations for each split note).

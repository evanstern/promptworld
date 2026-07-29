# Quickstart: Validate the Card-Format Policy (spec 087)

Prerequisites: the feature branch checked out (or main after merge). All checks are
read-backs — no build, no server.

## 1. Policy present in the durable home (FR-001..FR-005)

Open the project `CLAUDE.md`, section "## Backlog.md — the board". Expect a
card-format subsection stating:

- every task description opens with a 1–2 sentence plain-language gist (before
  anything else);
- gist is followed, where applicable, by "As a \<role\>" use cases; any accurate
  scene-setting role is valid;
- applicability rule: pure infra/bookkeeping cards may omit use cases, never the
  gist;
- the three operator examples with their ratings (good / less good / bad) and the
  reason each earns it.

Check: `grep -n "As a " CLAUDE.md` hits the policy examples inside the Backlog.md
block.

## 2. Spec-phase pointer present (FR-006)

In the same file, section "## Spec Kit — specs drive the work": expect a sentence
directing the spec author to the card's opening gist as the primary statement of
intent for tasks they didn't write.

Check: `grep -n "gist" CLAUDE.md` hits both the Backlog.md block and the Spec Kit
block.

## 3. Scope check (FR-007)

`git diff origin/main --stat` on the branch shows changes only to `CLAUDE.md` and
`specs/087-card-format-policy/**` (plus `.specify/feature.json`, per-worktree
pointer). Nothing under any plugin cache or praxisflux source tree.

## 4. First conforming example (SC-001/SC-002 seed)

`backlog task view TASK-168 --plain` — the card itself opens with a gist and carries
"As a …" use cases; the policy may cite it as the exemplar.

## Expected outcome

All four checks pass → the feature is validated end-to-end; SC-001's 3-card
spot-check runs organically on the next cards created after merge.

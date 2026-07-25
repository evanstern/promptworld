# Quickstart: validating the claim-before-work protocol (spec 065)

Prerequisites: repo root on `main` at the fetched tip; Node ≥ 18; git ≥ 2.38.
Contracts: [contracts/gate-cli-delta.md](./contracts/gate-cli-delta.md);
entities/rules: [data-model.md](./data-model.md).

## 1. Claim-time spec-number block (FR-004, FR-007)

```sh
# a number already taken on origin/main under a different slug → exit 1, names the
# taken dir and the next free number
node scripts/check-merge-drift.mjs claim --dir 051-something-else; echo "exit=$?"   # exit=1

# your own claimed dir (same dirname) → exit 0 (idempotent for the owner)
node scripts/check-merge-drift.mjs claim --dir 065-claim-before-work; echo "exit=$?"  # exit=0

# a free number → exit 0
node scripts/check-merge-drift.mjs claim --dir 999-free-number; echo "exit=$?"      # exit=0
```

Hook layer: with the repo's `.claude/settings.json` active, `mkdir specs/051-anything`
via Bash, or a Write to `specs/051-anything/spec.md`, is blocked before it executes;
writes into existing owned spec dirs pass.

## 2. Card-claim warning at worktree time (FR-003)

```sh
# a task whose card is NOT In Progress on origin/main → warn finding card-not-claimed,
# exit stays 0 (warn-only)
node scripts/check-merge-drift.mjs worktree --task TASK-9999

# a claimed task (card In Progress on origin/main) → no card finding
node scripts/check-merge-drift.mjs worktree --task TASK-139
```

## 3. Unpushed-branch audit (FR-005)

```sh
# create a task branch with a local-only commit, then:
node scripts/check-merge-drift.mjs session
# expect: [warn] branch-unpushed for that branch, prescribing git push -u origin <branch>
# push the branch; re-run; the finding clears
```

## 4. Two-session race simulation (FR-006, SC-001)

```sh
node --test scripts/claim-protocol.test.mjs
```

Expected: all tests pass — clone A's claim push accepted; clone B's competing push
rejected (non-fast-forward); after fetch, B's `claim --dir` exits 1 naming A's dir and
`worktree --task` warns on the unclaimed card.

## 5. Doctrine present (FR-001, FR-002)

```sh
grep -n "Claim-before-work" CLAUDE.md
# companion (separate repo):
grep -n "claim" ~/neumo/projects/praxis/pdlc/skills/sweep/templates/runbook.md
```

## 6. Full regression

```sh
node --test scripts/check-merge-drift.test.mjs scripts/claim-protocol.test.mjs
node scripts/check-merge-drift.mjs session   # existing modes unaffected
```

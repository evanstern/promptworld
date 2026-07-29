# Quickstart — spec 088 (TASK-162)

## Run the tests (primary validation)

```sh
node --test scripts/check-merge-drift.test.mjs
```

Expected: all existing tests plus the F1–F9 fixture matrix (see
[data-model.md](data-model.md)) pass.

## Manual validation on this repo

```sh
# from any task worktree with a branch:
node scripts/check-merge-drift.mjs pr
```

- On a branch that merged main in (history move): the report shows the player-docs
  probe ran (fresh ⇒ pass; stale ⇒ `player-docs-stale`).
- On a branch touching README.md only: same.
- On a branch touching a design-pinned source without amending `docs/design/tui/`:
  blocking `tui-design-stale` naming the page.

## Exit-code contract (unchanged)

0 = pass (clean or warnings-only) · 1 = blocked · 2 = usage/env error

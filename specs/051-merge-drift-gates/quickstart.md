# Quickstart: validating the merge-drift gates

Runnable scenarios proving each user story end-to-end. Scenarios 1–4 run in a
**disposable fixture repo** (recipe below) so no real branch is at risk; scenario 5 runs
read-only against this repo. Contracts: [gate-cli.md](./contracts/gate-cli.md),
[detection-rules.md](./contracts/detection-rules.md),
[report-schema.md](./contracts/report-schema.md).

## Prerequisites

- Node ≥ 18, git ≥ 2.38 (`git merge-tree --write-tree` must exist)
- `backlog` CLI on PATH (only for the `--notes` scenario)

## Fixture repo recipe

```bash
FIX=$(mktemp -d)/fixture && mkdir -p "$FIX" && cd "$FIX"
git init -b main remote-sim && git -C remote-sim commit --allow-empty -m init
# seed a wiki note pinned to the current tip, a backlog task file, a spec dir
cd remote-sim
mkdir -p docs/wiki backlog/tasks specs/001-seed scripts internal/tui
printf 'line1\nline2\nline3\n' > internal/tui/view.txt
PIN=$(git rev-parse HEAD)  # pin BEFORE the sources change below lands
printf -- '---\nname: seed\ndescription: d\nsources:\n  - internal/tui/view.txt\nverified_against: %s\n---\n' "$PIN" > docs/wiki/seed.md
git add -A && git commit -m seed
git clone "$PWD" ../work && cd ../work
git remote rename origin origin 2>/dev/null; git fetch origin
# two branches editing the same line = guaranteed textual conflict
git worktree add .worktrees/task-1 -b task-1-alpha origin/main
git worktree add .worktrees/task-2 -b task-2-beta  origin/main
sed -i '' 's/line2/alpha/' .worktrees/task-1/internal/tui/view.txt && git -C .worktrees/task-1 commit -am alpha
sed -i '' 's/line2/beta/'  .worktrees/task-2/internal/tui/view.txt && git -C .worktrees/task-2 commit -am beta
# land task-1 on the "remote" main as a SQUASH (tests the empty-contribution rule)
git -C ../remote-sim merge --squash --allow-unrelated-histories >/dev/null 2>&1 || true
git -C .worktrees/task-1 push origin task-1-alpha:refs/heads/task-1-alpha
git -C ../remote-sim merge --squash task-1-alpha && git -C ../remote-sim commit -m "squash task-1"
git fetch origin
```

Copy `scripts/check-merge-drift.mjs` from this repo into `work/scripts/` for the runs
below (all commands from `work/`, except the `pr` gate which runs inside a worktree).

## Scenario 1 — PR gate blocks a doomed PR (US1 / SC-001)

```bash
cd .worktrees/task-2
node ../../scripts/check-merge-drift.mjs pr
```

**Expected**: exit **1**; `textual-conflict` block finding naming
`internal/tui/view.txt` (task-1's squash landed the conflicting line on main). The same
run also warns `stale-base` (base lags main) and `wiki-sources-overlap` (branch touches
`internal/tui/view.txt`, pinned by `docs/wiki/seed.md`) and `tui-surface`.
`--json` verdict: `"blocked"`.

## Scenario 2 — session janitor catches the squash-merged worktree (US2 / SC-004)

```bash
node scripts/check-merge-drift.mjs session --json
```

**Expected**: exit **0** (warnings); `task-1-alpha` has `cleanupEligible: true` with
`cleanupReason: "empty-contribution"` (tip is NOT an ancestor — squash case) and a
`cleanupPrescriptions` entry with the exact two commands; root `fastForwarded: true` if
it was behind; the matrix marks `task-2-beta` vs `origin/main` as conflicting; `wiki`
grounding surface reports `stale: true` (pinned source changed since the pin). Then:

```bash
node scripts/check-merge-drift.mjs session --apply-cleanup
git worktree list   # .worktrees/task-1 gone; task-2 (dirty/live) untouched
```

Touch an uncommitted file in `.worktrees/task-2` first and confirm task-2 is **never**
cleanup-eligible (US2 scenario 6).

## Scenario 3 — worktree gate blocks stale root and taken spec numbers (US3 / SC-003)

```bash
git reset --hard origin/main~1          # stale root
node scripts/check-merge-drift.mjs worktree            # expect exit 1: root-stale
git merge --ff-only origin/main
node scripts/check-merge-drift.mjs worktree --spec 001 # expect exit 1: collision, suggests 002
node scripts/check-merge-drift.mjs worktree --spec 002 # expect exit 0
```

## Scenario 4 — findings become board notes with dedup (US4 / SC-006)

In a checkout with a real `backlog/` (or this repo on a throwaway task): run the pr gate
twice with `--notes` on a branch named `task-<N>-…` that has a warn/block finding.

**Expected**: first run appends one note containing the finding's `fingerprint:` line
(`backlog task view TASK-<N> --plain` shows it); second run appends **nothing**
(fingerprint dedup); a branch not matching `task-<N>-*` yields `task: null` and no note.

## Scenario 5 — determinism + performance on this repo (FR-012 / SC-002)

```bash
time node scripts/check-merge-drift.mjs session --json > /tmp/a.json
node scripts/check-merge-drift.mjs session --json > /tmp/b.json
diff /tmp/a.json /tmp/b.json   # identical (given no remote movement between runs)
```

**Expected**: identical output, wall time well under 30 s with all live worktrees
present. Also verify FR-009/SC-005: `git -C .worktrees/<any> reflog` shows no new
entries from gate runs.

## Offline behavior (FR-014)

Point `origin` at an unreachable URL in the fixture:
`pr` and `worktree` modes → exit **2** with a fail-closed message;
`session` → exit 0/1 with `unverifiedAgainstRemote: true` and verdict ≤ `warnings`.

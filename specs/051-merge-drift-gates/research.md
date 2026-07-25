# Phase 0 Research: Merge-Drift Gates

All Technical Context unknowns resolved. Decisions below were verified against the live
repository on 2026-07-25 (git 2.50.1, node v24.17.0, three live worktree branches).

## R1 — Script shape and runtime

**Decision**: single-file ESM Node script `scripts/check-merge-drift.mjs`, Node ≥ 18,
zero npm dependencies (stdlib `fs`/`path`/`child_process` only), `--json` flag for
machine output.

**Rationale**: exact precedent already ships twice in this repo —
`scripts/check-tui-design.mjs` (spec 047) and
`.claude/skills/player-docs/scripts/check-freshness.mjs` (spec 026). Same shape means
zero new toolchain, and sessions already know the invocation idiom.

**Alternatives considered**: Go binary under `cmd/` (rejected: build step, and gate
scripts in this repo are node by convention); bash (rejected: JSON report and
frontmatter parsing get unmaintainable).

## R2 — Textual conflict prediction

**Decision**: `git merge-tree --write-tree --name-only <base> <branch>`. Exit 0 = clean
merge (stdout: merged tree OID); exit 1 = conflicts (stdout: OID + conflicted file
list); exit < 0/other = error. Requires git ≥ 2.38.

**Rationale**: real three-way merge performed entirely in the object database — no
worktree, no index, no branch is touched, which is what makes FR-009 satisfiable.
Verified live on this repo: `origin/main` vs `task-107-tuning-manifest` → exit 0 with
tree OID; pairwise `task-124` vs `task-126` → exit 0. Side effect note: it writes
unreferenced tree/blob objects to the object DB; these are harmless and gc-collected —
documented, not treated as a mutation of any branch.

**Alternatives considered**: scratch clone + real `git merge --no-commit` (rejected:
slow, mutates a checkout, needs cleanup); `git merge-tree` legacy trivial mode
(rejected: deprecated semantics, no real merge-base handling).

## R3 — Live-branch enumeration and the n-way matrix

**Decision**: live branches = branches checked out in `.worktrees/*` (from
`git worktree list --porcelain`) ∪ local branches ahead of `origin/main`
(`git for-each-ref refs/heads` filtered by `merge-base --is-ancestor`). Matrix = R2 run
for every unordered pair plus each branch vs `origin/main`.

**Rationale**: matches the spec's v1 scope (remote-only branches excluded — the PR gate
covers each at its own choke point). At the documented scale cap (~10 branches) that is
≤ 55 merge-tree calls at ~10–50 ms each — comfortably inside SC-002's 30 s budget.

**Alternatives considered**: including remote branches via `refs/remotes` (deferred:
doubles matrix size for branches no local session owns); asking the hosting provider for
open PRs (rejected outright: FR-011).

## R4 — Cleanup eligibility (incl. squash-merge)

**Decision**: a worktree branch is cleanup-eligible iff
(a) `git merge-base --is-ancestor <tip> origin/main` succeeds, **or**
(b) the R2 merge of `origin/main` + `<tip>` exits 0 **and** the resulting tree OID
equals `origin/main^{tree}` (the branch contributes nothing not already on main — the
squash-merge signature), **and in all cases**
(c) `git -C <worktree> status --porcelain` is empty (no uncommitted changes).

**Rationale**: (a) covers merge commits; (b) covers squash merges where the tip is never
an ancestor — this operationalizes the project's recorded "squash rewrites branch pins /
check for empty diffs" lesson; (c) enforces US2 scenario 6. Pure git evidence, no
hosting-provider API (`gh pr view --json merged` rejected per FR-011).

**Alternatives considered**: `git cherry` / patch-id equivalence (rejected: fails when
squash commit differs textually from the sum of commits, e.g. after review fixups —
tree-equality of the actual merge result is the stronger, simpler test).

## R5 — Semantic overlap detection

**Decision**: compute changed-file sets with `git diff --name-only` — branch side:
merge-base(branch, origin/main)..branch-tip; mainline side: merge-base..origin/main.
Then:
- **backlog/**: warn on same-path intersection under `backlog/`.
- **wiki-pinned sources**: parse `sources:` lists from `docs/wiki/*.md` frontmatter
  (~45 notes, plain YAML list); warn when branch-side files intersect any note's
  sources, naming the note(s).
- **internal/tui/**: warn when branch-side files fall under `internal/tui/`, pointing at
  the spec-047 gate (`node scripts/check-tui-design.mjs --changed`).
- **spec-number collision**: numbers of branch-added `specs/NNN-*` dirs vs
  `git ls-tree origin/main specs/ --name-only`; block-level finding in `worktree` mode
  (given `--spec <NNN>`), warn-level in `pr` mode.

**Rationale**: set intersections on `--name-only` output are cheap, deterministic, and
exactly match the failure modes this project has actually recorded (spec-number
collisions, concurrent `backlog/` writes, stale wiki pins).

## R6 — Grounding-freshness flags (session mode)

**Decision**: three surfaces, checked as: **wiki** — per note, flag when
`git diff --name-only <verified_against> origin/main -- <sources…>` is non-empty (reads
each note's `verified_against:` pin directly); **player docs** — delegate to
`node .claude/skills/player-docs/scripts/check-freshness.mjs --check --json`;
**TUI design reference** — delegate to `node scripts/check-tui-design.mjs --json`.
Delegated checkers are invoked only if present; absence is reported as info, not error.

**Rationale**: FR-008 says reuse authoritative checkers rather than re-derive. The wiki
has no standalone checker script in-repo, but its notes carry machine-readable pins
(`sources:` + `verified_against:` frontmatter, verified against `docs/wiki/bundle-tools.md`)
— the diff-since-pin rule is precisely the freshness gate's definition from the
grounding-wiki plugin. Flags name the surface and the touched sources; actually
re-pinning remains `/grounding-wiki:wiki-update`'s job.

## R7 — Exit codes and verdict

**Decision**: exit 0 = pass (clean **or** warnings-only; report distinguishes), exit 1 =
blocked (≥ 1 block-severity finding), exit 2 = usage/environment error (bad flag, not a
git repo, git too old, fetch failed in a fail-closed mode). The tri-state verdict
(`pass` / `warnings` / `blocked`) lives in the report and `--json` output.

**Rationale**: warnings must not abort shell chains (`check && gh pr create` is the
intended idiom), so they cannot be non-zero; precedent scripts reserve 1 for violations
and 2 for environment errors, and this keeps that contract. FR-001's machine-readable
tri-state is fully served by `--json`.

**Alternatives considered**: exit 3 for warnings (rejected: breaks `&&` gating and
diverges from both precedent scripts for no consumer that needs it).

## R8 — Mutation whitelist mechanics (FR-009)

**Decision**:
- `git fetch origin` — always (except when it fails; see R10).
- Root ff-pull — session mode only, automatic, and only when the root worktree is on
  `main`, behind `origin/main`, not diverged, with a clean status.
- Worktree/branch cleanup — never automatic; `--apply-cleanup` applies the prescribed
  removal **only** to worktrees already verified cleanup-eligible (R4) in the same run.
- Board notes — `--notes` flag; writes exclusively via
  `backlog task edit TASK-<N> --append-notes …`.
- Everything else read-only.

**Rationale**: mirrors FR-009 exactly; defaults prescribe rather than apply, so an
un-flagged run is safe to execute from any state.

## R9 — Board-note attribution and dedup

**Decision**: task id derived from branch name `task-<N>-<slug>` → `TASK-<N>` (fallback:
the spec-bridge `Spec: specs/…` marker when the finding names a spec dir); findings with
no derivable task stay report-only. Each finding carries a **fingerprint** — a short
stable hash of (gate, rule, evidence paths, branch) — embedded in the note text; before
appending, the script reads the task's file under `backlog/tasks/` (read-only) and skips
notes whose fingerprint is already present.

**Rationale**: gates run repeatedly at choke points by design; without dedup, `--notes`
would spam every re-run. Reading backlog files is allowed (only *writes* must go through
the CLI), so fingerprint-scanning the task file is the cheapest exact dedup.

## R10 — Offline / fetch-failure behavior

**Decision**: `pr` and `worktree` modes: fetch failure → exit 2 with an explicit
"cannot verify against remote — gate fails closed" finding. `session` mode: continue all
local checks, mark the report `unverifiedAgainstRemote: true`, verdict capped at
`warnings`.

**Rationale**: FR-014 verbatim; a PR/worktree verdict computed from a stale remote ref
is exactly the false confidence this feature exists to kill.

## R11 — Validation strategy

**Decision**: no automated test harness; quickstart.md carries runnable fixture-repo
scenarios (a scripted recipe that builds a throwaway repo with a conflicting pair, a
squash-merged branch, a wiki-pinned source, and a spec-number collision) with expected
verdicts/exit codes per user story.

**Rationale**: matches the spec-047 precedent (contract docs + quickstart validation, no
`.mjs` test suite exists anywhere in this repo); the scenarios map 1:1 to the spec's
Independent Test lines, which is what `/speckit-tasks` will turn into verification
tasks.

# Contract: detection rules — the git plumbing (normative)

Every rule below is deterministic given identical repo + remote state (FR-012). All
commands are read-only except where the gate-cli contract's mutation whitelist says
otherwise.

## 1. Live-branch enumeration (session mode)

Live = (branches checked out under `.worktrees/`, from
`git worktree list --porcelain`) ∪ (local branches not fully contained in origin/main:
`git for-each-ref refs/heads --format='%(refname:short) %(objectname)'` where
`git merge-base --is-ancestor <tip> origin/main` fails). `main` itself is excluded.
Remote-only branches are out of scope (v1, per spec assumption).

## 2. Textual conflict prediction

```
git merge-tree --write-tree --name-only <A> <B>
```

- exit 0 → clean; stdout line 1 = merged tree OID
- exit 1 → conflict; stdout = OID, blank line, then conflicted paths (one per line)
- any other exit → environment error (exit 2 upstream)

Used branch-vs-`origin/main` (pr gate, matrix) and pairwise (matrix). Pairs are
evaluated in lexicographic branch-name order for report stability.

## 3. Changed-file sets and overlaps

```
BASE=$(git merge-base <tip> origin/main)
branchFiles = git diff --name-only $BASE <tip>
mainFiles   = git diff --name-only $BASE origin/main
baseLag     = git rev-list --count $BASE..origin/main
```

- `backlog-overlap`: same path in both sets with prefix `backlog/`.
- `wiki-sources-overlap`: `branchFiles` ∩ union of `sources:` lists parsed from
  `docs/wiki/*.md` frontmatter (frontmatter = leading `---` block; `sources:` is a plain
  YAML string list; parse tolerantly, skip malformed notes with an info finding).
  Evidence names the affected note(s).
- `tui-surface`: any `branchFiles` entry with prefix `internal/tui/`.
- `stale-base`: `baseLag > 0`.

## 4. Cleanup eligibility (session mode)

Branch `B` with tip `T`, worktree `W`:

```
eligible(B) =
  ( git merge-base --is-ancestor T origin/main            # merge-commit case
    OR ( merge-tree(origin/main, T) exits 0
         AND mergedTreeOID == $(git rev-parse origin/main^{tree}) ) )  # squash case
  AND git -C W status --porcelain is empty                 # never with local changes
```

Prescribed remediation (report text, applied only under `--apply-cleanup`):

```
git worktree remove <W>
git branch -d <B>     # ancestor case
git branch -D <B>     # empty-contribution (squash) case — see below
```

Branch deletion is conditional on which eligibility signal fired: `-d` for the
`ancestor` case (git's own merged-check agrees), `-D` for the `empty-contribution`
case — a squash-merged branch is never an ancestor of origin/main, so `-d` always
refuses it even though this run has already proven via merge-tree tree-equality that
the branch contributes nothing not on main. The tree-equality proof is the safety
check; `-D` merely bypasses git's weaker ancestor heuristic. (Amended 2026-07-25 from
implementation finding — the original contract prescribed `-d` unconditionally, which
made the squash case uncompletable.)

## 5. Spec-number collision

Taken numbers = leading `NNN` of each entry in
`git ls-tree origin/main specs/ --name-only`. `worktree --spec NNN`: block when taken,
report the smallest free number > max(taken). `pr` mode: numbers of `specs/NNN-*` paths
newly added by the branch (present in `branchFiles`, absent on origin/main) that collide
→ warn.

## 6. Grounding freshness (session mode)

- **wiki (internal)**: for each `docs/wiki/*.md` with frontmatter `verified_against:
  <sha>` and `sources:` — stale iff
  `git diff --name-only <sha> origin/main -- <sources…>` is non-empty. Unresolvable pin
  SHA (e.g. gc'd after squash) → info finding, not a crash.
- **player-docs (delegated)**:
  `node .claude/skills/player-docs/scripts/check-freshness.mjs --check --json`;
  stale iff exit 1. Script absent → `checker: "absent"`, info.
- **tui-design (delegated)**: `node scripts/check-tui-design.mjs --json`; stale iff
  exit 1. Script absent → info.

## 7. Task attribution & note dedup

- Branch `task-<N>-<slug>` → `TASK-<N>`. Otherwise, if the finding's evidence names a
  `specs/NNN-*` dir, resolve via the spec-bridge marker (`Spec: specs/NNN-…` line in a
  task file under `backlog/tasks/`, read-only scan). No match → `task: null`,
  report-only.
- `fingerprint = first 12 hex of SHA-256("<gate>|<rule>|<branch>|<sorted evidence>")`.
- Before appending a note, read the target task's file under `backlog/tasks/`
  (read-only); skip if the fingerprint string already appears. Append via
  `backlog task edit TASK-<N> --append-notes <BoardNote text>` (data-model.md).

## 8. Root state

- `onMain`: `git -C <root> symbolic-ref --short HEAD` == `main`; anything else →
  block-severity `root-not-main` in every mode.
- behind/ahead: `git rev-list --count main..origin/main` / `origin/main..main`.
- Guarded ff-pull (session only): behind > 0 AND ahead == 0 AND clean status →
  `git -C <root> merge --ff-only origin/main` (no second network fetch).

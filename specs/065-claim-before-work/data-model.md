# Data Model: Claim-Before-Work Protocol (spec 065)

No persistent storage. All entities are derived at gate runtime from git trees and
tool-call inputs; findings follow the spec 051 report schema
(`specs/051-merge-drift-gates/contracts/report-schema.md`) unchanged.

## Entities

### Claim
The first pushed commit of a task. Composite of:
- **card move** — the task's `backlog/tasks/task-<n> - *.md` file with frontmatter
  `status: In Progress`, present on `origin/main`.
- **spec-number claim** — a `specs/<NNN>-<slug>/` path present on `origin/main`
  (any tracked file under it; a stub `spec.md` suffices).

Ownership is defined ONLY by presence on `origin/main` (the fetched tip at check
time). Local-only state never constitutes a claim.

### TakenSpecNumber
Derived by existing `takenSpecNumbers(originMainTip)`:
`Map<number, dirname>` from `git ls-tree origin/main specs/`.
- **Collision**: candidate dir `NNN-slug` collides iff `taken.has(NNN) &&
  taken.get(NNN) !== 'NNN-slug'`. Same-dirname re-claims are the owner's own claim —
  never a collision (idempotence rule).

### CardStatus
Parsed from the task card file's frontmatter in the `origin/main` tree:
- located via `git ls-tree origin/main backlog/tasks/` + filename regex
  `^task-<n>[ .]` (same convention as existing `findTaskFile()`);
- `status:` line value, expected `In Progress` for a claimed task.
- Missing card file, unparseable frontmatter → treated as *not claimed* (warn fires;
  never a block, never a crash).

### UnpushedTaskBranch
Derived in session mode's branch scan:
- branch name matches `^task-\d+-` (existing convention / `attributeTask()`);
- tip is ahead of `merge-base(branch, origin/main)` by ≥ 1 commit;
- `refs/remotes/origin/<branch>` absent after fetch.
All three conditions → finding.

## New findings (report-schema shape, unchanged fields)

| rule | gate (mode) | severity | message contract | evidence | task attribution |
|---|---|---|---|---|---|
| `spec-number-collision` | `claim` | block | names the taken dir on origin/main and the next free number (same wording family as worktree `--spec`) | `specs/<takenDir>` | `attributeBySpecDir(takenDir)` |
| `card-not-claimed` | `worktree` | warn | task card for `TASK-<n>` is not `In Progress` on origin/main — claim missing or unpushed; states the claim doctrine one-liner | card path (or `backlog/tasks/` when missing) | `TASK-<n>` |
| `branch-unpushed` | `session` | warn | local task branch has commits but no remote counterpart — not auditable from other clones; prescribe `git push -u origin <branch>` | branch name | `attributeTask(branch)` |

Existing findings, verdict computation (`pass`/`warnings`/`blocked`), exit-code
mapping, fingerprinting, and `--notes` dedup are reused untouched.

## Mode/flag surface (delta — normative text in contracts/gate-cli-delta.md)

| addition | shape |
|---|---|
| `claim` mode | `node scripts/check-merge-drift.mjs claim --dir <NNN-slug> [--json]` |
| `worktree --task` | `node scripts/check-merge-drift.mjs worktree [--spec NNN] [--task TASK-<n>]` |
| session scan | no new flags; `branch-unpushed` appears in the existing branches/findings sections |

## State transitions

```
task lifecycle (protocol view):

  To Do ──(claim commit: card→In Progress + specs/NNN-slug stub; push)──▶ claimed
    │                                                                       │
    │  push rejected (non-fast-forward)                                     ▼
    └──▶ lost race ──(fetch; re-read board+specs/)──▶ holder is another    work proceeds
              session? ──yes──▶ STOP LANE, surface to operator             (worktree cut,
                       └─no (unrelated push)──▶ rebase, re-push claim       branch pushed
                                                                            on first commit)
```

## Validation rules

- `claim --dir` value MUST match `^\d{3,}-[A-Za-z0-9._-]+$`; otherwise usage error
  (exit 2).
- `--task` value MUST match `^TASK-\d+$`; otherwise usage error (exit 2).
- `claim` mode fetch failure → exit 2 (fail closed); `worktree` unchanged (already
  fail closed); `session` unchanged (degrades per 051 FR-014).
- Hook layers MUST fail open on: malformed stdin, non-matching commands/paths,
  out-of-jurisdiction cwd, missing gate script — identical posture to existing
  pre-bash.

# Contract: gate CLI — `scripts/check-merge-drift.mjs`

Single-file ESM Node script, Node ≥ 18, zero npm dependencies (stdlib
`fs`/`path`/`child_process` only). Requires git ≥ 2.38 (`merge-tree --write-tree`);
exits 2 with a clear message on older git.

## Invocation

```
node scripts/check-merge-drift.mjs <mode> [flags]
```

### Modes

| mode | run from | purpose | mandatory when |
|---|---|---|---|
| `session` | repo root | janitor pass + drift matrix + grounding flags | at session start |
| `worktree` | repo root | fresh-base + spec-number gate | before `git worktree add` |
| `pr` | inside the task worktree | conflict/overlap gate on the current branch | before opening any PR |

### Flags

| flag | modes | effect |
|---|---|---|
| `--json` | all | machine-readable report (see report-schema.md) instead of human lines |
| `--spec <NNN>` | `worktree` | also verify spec number `NNN` is unused under `specs/` on origin/main |
| `--branch <name>` | `pr` | gate a branch other than the current checkout (defaults to `HEAD`'s branch) |
| `--notes` | all | record task-attributable findings (severity ≥ warn) as board notes via the `backlog` CLI, fingerprint-deduped |
| `--apply-cleanup` | `session` | apply the prescribed removal of worktrees verified cleanup-eligible in this same run (worktree remove + branch delete); never anything else |
| `--no-fetch` | `session` | skip the fetch (equivalent to fetch-failure degraded mode; forbidden in `pr`/`worktree` — usage error) |

Unknown flags, unknown modes, `pr` mode run at root on `main`, or `--no-fetch` outside
`session`: exit 2.

## Exit codes

| code | meaning |
|---|---|
| 0 | pass — clean or warnings-only (verdict distinguishes; warnings never break `check && …` chains) |
| 1 | blocked — ≥ 1 block-severity finding; the gated action MUST NOT proceed |
| 2 | usage/environment error — bad invocation, not a git repo, git < 2.38, or fetch failure in a fail-closed mode (`pr`, `worktree`) |

## Mutation whitelist (normative — FR-009)

The script may perform **only** these writes, and nothing else under any flag
combination:

1. `git fetch origin` (always attempted unless `--no-fetch`).
2. Fast-forward of the **root** checkout — `session` mode only, automatic, only when
   root is on `main`, behind `origin/main`, not diverged, and clean.
3. With `--apply-cleanup`: `git worktree remove` + `git branch -d` for worktrees that
   this run verified cleanup-eligible (detection-rules.md §4).
4. With `--notes`: `backlog task edit TASK-<N> --append-notes …` — never direct writes
   under `backlog/`.

It never rebases, merges, commits to, checks out, or resets any task branch or its
worktree. (`git merge-tree` writes unreferenced objects to the object DB; these
reference no branch and are gc-collected — not a mutation of any checkout.)

## Behavior per mode

### `session`

1. Fetch (or degrade per FR-014: continue local checks,
   `unverifiedAgainstRemote: true`, verdict capped at `warnings`).
2. Root state: on-main check (block if not), guarded ff-pull if behind.
3. Enumerate live branches (detection-rules.md §1); per branch: merge base, base lag,
   dirty state, changed files, cleanup eligibility (prescribe exact remediation
   commands in the report).
4. Drift matrix: each branch vs origin/main and each unordered pair.
5. Semantic overlaps per branch vs mainline (backlog/, wiki sources, spec numbers).
6. Grounding surfaces: wiki via frontmatter diff-since-pin; player-docs and TUI via
   their own checkers when present (absent → info).

### `worktree`

1. Fetch; failure → exit 2 (fail closed).
2. Block unless root is exactly at the fetched `origin/main` tip and on `main`.
3. With `--spec <NNN>`: block if `specs/<NNN>-*` exists on origin/main; report the next
   free number.

### `pr`

1. Fetch; failure → exit 2 (fail closed).
2. Identify the branch (must not be `main`); derive merge base vs fetched origin/main.
3. `merge-tree` vs origin/main: conflicts → **block**, naming each conflicted file.
4. Warn-level checks: stale base (lag count), backlog/ same-file overlap, wiki-pinned
   source overlap (naming the notes), `internal/tui/` touches (pointing at the spec-047
   gate), new-spec number collisions.

## Enforcement (FR-013)

CLAUDE.md gains a "Merge-drift gates" section, adjacent to the spec-047 TUI gate block,
documenting the three invocations verbatim and when each is mandatory:

- session start (root): `node scripts/check-merge-drift.mjs session`
- before cutting a worktree: `node scripts/check-merge-drift.mjs worktree [--spec NNN]`
- before opening a PR (from the worktree): `node scripts/check-merge-drift.mjs pr`

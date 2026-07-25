# Data Model: Merge-Drift Gates

All entities are in-memory shapes inside a single script run, surfaced verbatim in the
`--json` report ([contracts/report-schema.md](./contracts/report-schema.md)). Nothing is
persisted by the script except board notes (via the `backlog` CLI) and the mutations
whitelisted by FR-009.

## GateRun

The top-level object; one per invocation.

| Field | Type | Notes |
|---|---|---|
| `mode` | `"session" \| "worktree" \| "pr"` | chosen by CLI mode argument |
| `verdict` | `"pass" \| "warnings" \| "blocked"` | max severity of findings; drives exit code (pass/warnings→0, blocked→1) |
| `unverifiedAgainstRemote` | boolean | session mode only; true when fetch failed (R10) |
| `originMain` | string (40-hex) | `origin/main` tip the run was computed against |
| `root` | RootState | root-checkout facts |
| `branches` | LiveBranch[] | enumeration per R3 (empty in `worktree` mode) |
| `matrix` | MatrixCell[] | session mode only |
| `findings` | Finding[] | ordered block → warn → info |

**Validation**: `verdict` MUST equal the max severity present in `findings`
(`blocked` > `warnings` > `pass`); determinism (FR-012) requires stable ordering of
`branches`, `matrix`, and `findings` (lexicographic by branch/pair/rule).

## RootState

| Field | Type | Notes |
|---|---|---|
| `onMain` | boolean | false is itself a block finding (edge case: root not on main) |
| `behindBy` / `aheadBy` | number | vs `origin/main` |
| `fastForwarded` | boolean | session mode; true when the guarded ff-pull ran (R8) |
| `dirty` | boolean | uncommitted changes at root |

## LiveBranch

| Field | Type | Notes |
|---|---|---|
| `name` | string | e.g. `task-107-tuning-manifest` |
| `tip` | string (40-hex) | |
| `worktree` | string \| null | path under `.worktrees/`, null for non-worktree local branches |
| `task` | string \| null | `TASK-<N>` per attribution rule (R9); null → findings stay report-only |
| `mergeBase` | string (40-hex) | vs `origin/main` |
| `baseLag` | number | mainline commits between mergeBase and origin/main tip (FR-003) |
| `dirty` | boolean | worktree has uncommitted changes |
| `changedFiles` | string[] | `git diff --name-only mergeBase..tip` |
| `cleanupEligible` | boolean | R4 rule: (ancestor ∨ empty-contribution) ∧ ¬dirty |
| `cleanupReason` | `"ancestor" \| "empty-contribution" \| null` | which R4 signal fired |

**State transitions**: none persisted — `cleanupEligible` is recomputed fresh each run;
a branch leaves the model entirely once its worktree and ref are removed.

## MatrixCell

One per unordered pair of live branches, plus one per (branch, origin/main).

| Field | Type | Notes |
|---|---|---|
| `a`, `b` | string | branch names; `b` may be `"origin/main"` |
| `conflict` | boolean | merge-tree exit 1 |
| `files` | string[] | conflicted paths (empty when clean) |

## Finding

| Field | Type | Notes |
|---|---|---|
| `severity` | `"block" \| "warn" \| "info"` | |
| `gate` | `"session" \| "worktree" \| "pr"` | producing mode |
| `rule` | string | stable id, e.g. `textual-conflict`, `stale-base`, `backlog-overlap`, `wiki-sources-overlap`, `tui-surface`, `spec-number-collision`, `cleanup-eligible`, `grounding-stale`, `root-not-main`, `remote-unverified` |
| `message` | string | human-readable, actionable |
| `evidence` | string[] | file paths / branch names / note names / commit SHAs |
| `task` | string \| null | `TASK-<N>` when attributable (R9) |
| `fingerprint` | string | stable short hash of (gate, rule, sorted evidence, branch) — dedup key for board notes |

**Severity assignment (normative)**:

| rule | pr | worktree | session |
|---|---|---|---|
| `textual-conflict` (vs origin/main) | block | — | warn (matrix) |
| `pairwise-conflict` (branch vs branch) | — | — | warn |
| `stale-base` | warn | — | info |
| `root-not-main` | block | block | block |
| `root-stale` (not at fetched tip) | — | block | info (auto-ff'd) |
| `spec-number-collision` | warn | block | warn |
| `backlog-overlap` | warn | — | warn |
| `wiki-sources-overlap` | warn | — | — |
| `tui-surface` | warn | — | — |
| `cleanup-eligible` | — | — | warn (prescriptive) |
| `grounding-stale` | — | — | warn |
| `dirty-worktree` | info | — | info |
| `remote-unverified` | block (exit 2) | block (exit 2) | warn |

## GroundingSurface

| Field | Type | Notes |
|---|---|---|
| `name` | `"wiki" \| "player-docs" \| "tui-design"` | |
| `checker` | `"internal" \| "delegated" \| "absent"` | wiki = internal (frontmatter diff, R6); others delegated |
| `stale` | boolean \| null | null when checker absent |
| `touched` | string[] | source files changed since the surface's pin |

## BoardNote

What `--notes` appends to a task (via `backlog task edit TASK-<N> --append-notes`):

```
[merge-drift <gate>] <severity>: <message>
evidence: <paths…>
fingerprint: <fingerprint>   ← dedup key; run skipped if already present in task file
```

**Validation**: written only for findings with `task` set and severity ≥ warn; one
append per new fingerprint per run.

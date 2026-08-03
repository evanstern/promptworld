# Implementation Plan: Claim gate sees spec numbers held by pushed branches

**Spec**: `specs/111-claim-gate-branch-visibility/spec.md` · **Task**: TASK-188

## Constitution check

| Principle | Assessment |
|---|---|
| I — Artifact-grounded action | PASS. Card TASK-188 carries the diagnosis and ACs; this spec dir claims number 111 and its stub is already merged to main; findings land as gate exit codes, not prose. |
| II — One TASK, one PR | PASS. Single deliverable, single branch `task-188-claim-gate-branch-visibility`, single PR. |
| III — Gates over argument | PASS. The change *is* a gate; it is verified by the gate's own test harness, not by assertion. |
| IV — Grounding freshness | Conditional. `docs/wiki/` notes listing `scripts/check-merge-drift.mjs` as a source must be re-pinned in-branch; the pr gate enforces this. Checked in T006. |
| V — Model-tiered workflow | **Deviation, recorded.** The constitution routes implementation to a delegated subagent (`spec-implementer`, Sonnet default). This session's operator instructions explicitly bar spawning agents unless requested, and the operator directed this session to perform the fix ("go ahead and fix this if you can"). Implemented inline on the planning model (Opus 5, `claude-opus-5`). Rubric note: had it been delegated, this is a Sonnet-tier slice — single file, surgical, complete file:line diagnosis pinned. |

## Approach

One new pure-read helper plus a rewritten collision decision inside `runClaim`. No
change to fetch behavior, exit codes, report shape, or any other mode.

### 1. `branchHeldSpecNumbers(cwd)` — new helper, beside `takenSpecNumbers`

Reads already-fetched remote-tracking refs; never fetches.

```
git for-each-ref --format=%(refname:short) refs/remotes/origin/
  -> keep refs matching origin/task-*            (the repo's branch convention)
  -> for each: git ls-tree --name-only <ref> -- specs/
  -> parse NNN-slug, build Map(number -> { dir, branch })
```

Decisions:

- **Scope to `origin/task-*`.** The repo's branch convention (CLAUDE.md, spec 065) is
  `task-<N>-<slug>`. Scanning every remote head would drag in release/experiment
  branches whose spec dirs are not claims.
- **Deterministic first-writer-wins** on duplicate numbers. `for-each-ref` sorts by
  refname, so when two branches hold one number (the live 110 case) the reported
  holder is stable across runs — a gate that names a different branch each run is
  not reproducible evidence.
- **Failure degrades, never crashes** (FR-008): a non-zero `for-each-ref` returns an
  empty map (today's behavior); a non-zero `ls-tree` skips that one branch.

### 2. `runClaim` decision order

```
taken  = takenSpecNumbers(originMainTip, mainWt)     # unchanged
held   = branchHeldSpecNumbers(mainWt)               # new

if taken has N and taken[N] != dir     -> block, main-held    (message unchanged)
elif held has N and held[N].dir != dir -> block, branch-held   (new message)
else                                    -> pass
```

Main precedence satisfies FR-003; comparing dirnames on both arms satisfies FR-002.

### 3. Next-free number

Today: `Math.max(...taken.keys()) + 1` — wrong under FR-005 once branches hold
numbers, and it throws `-Infinity` on an empty map. Replace with a shared helper over
the union of both key sets, guarded for empty.

### 4. Message shapes

- Main-held (unchanged, byte-for-byte — SC-003 depends on it):
  `specs/<taken> already exists on origin/main for number NNN — claim specs/<dir> is a collision; next free number is NNN`
- Branch-held (new, FR-004):
  `specs/<taken> is already claimed on branch <branch> for number NNN — claim specs/<dir> is a collision; next free number is NNN (if that branch is abandoned, delete it on origin and re-run)`

Both use `rule: 'spec-number-collision'` and `severity: 'block'`, so downstream
consumers (the hook's message passthrough, the JSON report) need no change.

## Verification

- `node --test scripts/claim-protocol.test.mjs` — new + existing cases.
- `node --test scripts/check-merge-drift.test.mjs` — no regression in other modes.
- Live reproduction against this repo's actual 110 collision (SC-001), recorded on
  the card as evidence.
- `node scripts/check-merge-drift.mjs pr` from the worktree before opening the PR.

## Risk

**A stale unmerged branch permanently blocks its number.** Accepted, with mitigation:
the message names the branch and prescribes the fix (delete it on origin). This is
strictly better than the current silent collision — a blocked claim is visible and
one command from resolution; a duplicate claim costs two lanes' spec work.

# Feature Specification: Claim gate sees spec numbers held by pushed branches

**Feature Branch**: `task-188-claim-gate-branch-visibility`

**Created**: 2026-08-02

**Status**: Specified

**Board task**: TASK-188

**Input**: TASK-188 — the claim gate's spec-number exclusion has a hole wide enough
for two concurrent sessions to walk through, and on 2026-08-02 two of them did.

---

## Problem

`scripts/check-merge-drift.mjs` claim mode is spec 065's mutual-exclusion gate for
spec numbers. It defines ownership **solely by presence on `origin/main`**:

- `takenSpecNumbers(originMainTip, cwd)` (line 614) reads only the `origin/main` tree.
- `runClaim(flags, cwd)` (line 1387) consults nothing else.
- The design intent is stated outright at lines 461-465: *"The claim protocol defines
  ownership by presence on origin/main, so these helpers read the FETCHED tree
  (ls-tree + show) and never the working dir."*

Spec 065's protocol compensates for this narrowness by requiring the claim stub to
land on main **immediately** — the task branch's first commit creates
`specs/NNN-<slug>/`, and that branch is merged to main at root with
`git merge --no-ff` right away. The gate is correct **only for as long as every
session performs that merge promptly.**

When the merge is skipped or merely delayed, the number is claimed in a place the
gate cannot see: a pushed-but-unmerged branch. Every subsequent session's claim gate
passes, and the collision surfaces days later — after the colliding lanes have
written full `spec.md` + `plan.md` + `tasks.md` against the same number, when a
renumber is expensive and needs human arbitration.

### Observed occurrence (2026-08-02)

| Lane | Branch | Spec dir | On origin? | Merged to main? |
|---|---|---|---|---|
| TASK-173 | `task-173-absence-attribution` | `specs/110-absence-attribution` | yes | **no** |
| TASK-187 | `task-187-frame-harness` | `specs/110-tui-frame-harness` | yes | **no** |

Both carry `spec.md`, `plan.md`, and `tasks.md`. Both passed the claim gate. The
collision was detected only when the spec-bridge Stop hook complained that both
cards' statuses exceeded their artifacts — a symptom two removes from the cause,
because from the root checkout `specs/110-*` does not exist at all.

A live probe confirms the hole is still open:

```
$ node scripts/check-merge-drift.mjs claim --dir 110-something-new
check-merge-drift: mode=claim verdict=pass ...
no findings
```

A third session would have taken 110 as well.

### Why this is a gate defect and not just a protocol violation

The immediate-merge step is a human/agent discipline with no enforcement. A gate
whose correctness depends on an unenforced step elsewhere is not a mutex — it is a
convention with a check attached. The repo already proves branch trees are readable
at gate time: `specNumberCollisions()` (line 635) compares a **branch's** new spec
dirs against `origin/main` and is used by both session mode (line ~1190) and pr mode
(line ~1890). Branch-vs-main is covered; **branch-vs-branch is the gap.**

---

## User scenarios

**As a session about to start work**, when I claim a spec number, I want to be told
immediately if another in-flight session already holds it — not to discover it after
I have written a whole spec against that number.

**As the operator**, when two lanes do collide, I want the gate to have caught it at
claim time, so I am not asked to arbitrate a renumber after both lanes already carry
full spec/plan/tasks artifacts.

**As a session resuming my own lane**, when I re-run the claim gate against the
directory I already claimed, I want it to pass — whether my claim currently lives on
main, on my own pushed branch, or only in my working tree.

---

## Requirements

- **FR-001** — Claim mode MUST treat a spec number as taken when a **pushed remote
  task branch** carries `specs/NNN-<slug>/` under a dirname different from the one
  being claimed, even when that number is absent from `origin/main`.
- **FR-002** — Claim mode MUST remain idempotent for the owner: claiming dirname `D`
  at number `N` passes when the only holder of `N` is `D` itself, regardless of
  whether `D` lives on `origin/main`, on a remote branch, or nowhere yet.
- **FR-003** — An `origin/main` holder MUST continue to take precedence in the
  reported finding: main is the settled record, a branch claim is in-flight. Where
  both hold the number, the message names main's.
- **FR-004** — The block message MUST name the holding **branch** (in addition to the
  taken dirname and owning task) when the holder is branch-side, so the reader can
  tell an active lane from an abandoned one.
- **FR-005** — The next-free number MUST be computed from the union of main-held and
  branch-held numbers, so following the gate's advice cannot land on a second
  collision.
- **FR-006** — The branch scan MUST be a pure read of already-fetched remote-tracking
  refs. No fetch beyond claim mode's existing one, no writes, no new mutation surface.
- **FR-007** — The gate MUST continue to fail closed (exit 2) when the remote is
  unreachable, unchanged from today.
- **FR-008** — A branch scan failure (unreadable ref, malformed tree) MUST degrade to
  today's main-only behavior rather than crashing the gate.

### Out of scope

- `worktree --spec` claim-awareness (line ~1485) keeps its current main-only reading.
  Claim mode is the first choke point; a number blocked there never reaches worktree
  mode. Recorded as a follow-up, not fixed here.
- Resolving the live 110 collision between TASK-173 and TASK-187. That is an
  operator arbitration call on two in-flight lanes this task does not own; this spec
  only makes the collision visible and reproducible.
- Any change to the immediate-merge step of spec 065's protocol. The step stays
  correct and recommended; this work removes the silent-failure mode when it slips.

---

## Success criteria

- **SC-001** — With the fix in place, the live 110 situation reproduces as a block:
  claiming a third dirname at 110 exits 1 and names a holding branch.
- **SC-002** — Owner re-claim passes in all three holder positions (main, own branch,
  nowhere) — zero false blocks on the idempotent path.
- **SC-003** — Every pre-existing claim-mode behavior is unchanged: main-held
  collisions still block with their current message shape, clean claims still pass,
  unreachable remote still exits 2.
- **SC-004** — Both hook layers (`merge-drift-hook.mjs` pre-write and pre-bash, which
  shell out to `claim --dir`) inherit the new blocking behavior with no hook change.

---

## Blast radius

`scripts/check-merge-drift.mjs` claim mode only. Because
`scripts/hooks/merge-drift-hook.mjs` invokes the gate as a subprocess
(`tryExec('node', [gateScript, 'claim', '--dir', dir])`, line 271), both hook layers
pick the fix up for free — a `Write` into a colliding `specs/NNN-*/` and a `mkdir` of
one both begin blocking with no edit to the hook.

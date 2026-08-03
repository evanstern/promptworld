# TUI frame harness — sweep runbook (2026-08-02)

**You (the session reading this) are the ORCHESTRATOR** for the task below. Run it through
the full PDLC — spec → link → worktree → delegated implementation → PR → merge → re-ground.
Direction is decided; do not re-litigate it: the operator's 2026-08-02 decision selecting a
three-fixture set, and TASK-187's card, win. Plan-of-record is the board; this file carries
only ordering, doctrine, and the log.

**Status:** draft · operator sign-off on lanes: pending

<!-- Only the OPERATOR flips draft → signed-off. -->

## Read first (in this order)

1. `backlog task view TASK-187 --plain` — the deliverable and its 9 acceptance criteria.
2. `docs/design/tui/INDEX.md` — the UI authority gate (spec 047) this task writes into.
3. `CLAUDE.md` — "TUI design reference", "Merge-drift gates", "Wiki-in-PR lifecycle".
4. `backlog task list --plain` — live state; other sessions move it while you work.

## State when this runbook was written (2026-08-02)

- **Done already:** nothing in this scope.
- **In flight in other sessions (do not duplicate; expect their merges):** board-sync
  pushes only — TASK-173 was re-opened on main at `8a85e4b7` during this sweep's
  precondition gate. No sibling task branch touches `internal/tui/` or `docs/design/tui/`.
- **Paused — untouched:** no task carries the `paused` label.
- **Queued (this runbook's scope):** TASK-187 — single task, single lane.

## Execution mode — background job / no-main-push

This sweep runs as a **Claude Code background job**, where the harness forbids pushing or
merging `main`. Per the sweep skill's background-job execution mode, the substitutions
below apply and are recorded so a resuming session does not read them as sloppiness.

- **The claim's main-push is unavailable**, so claim-before-work's mutual-exclusion
  primitive (spec 065) cannot fire. The card was moved to In Progress at root under the
  board-sync exception but sits **unpushed** pending the operator; `check-merge-drift
  worktree` therefore warns `card-not-claimed` by design, not omission.
- **This cost us the spec number, exactly as predicted.** The sweep first claimed
  **spec 110** (`check-merge-drift claim --dir 110-tui-frame-harness` passed at
  `8a85e4b7`) and pushed `specs/110-tui-frame-harness/` on the task branch. Within the
  hour a concurrent session re-opened TASK-173, linked it to
  `specs/110-absence-attribution`, and **pushed that claim to `origin/main`** — where
  TASK-187 still read `To Do` because this job cannot push the board. A branch push
  claims nothing; the push to `main` is the primitive, so that session holds 110 and this
  one lost the race fairly. Per doctrine this is a renumber, not a stop-the-lane: the
  collision was on the NUMBER, and TASK-187 itself was never contested.
- **Then it happened a second time, mid-rename.** The renumber to **111** was already
  under way when `specs/111-claim-gate-branch-visibility` landed on `origin/main`; the
  claim hook caught it on the next write (`spec-number-collision`, next free 112).
  Renumbered again to **spec 112**, re-verified free at `1b17cd33`. Two lost races inside
  one task is the measurement, not bad luck: **in background mode the spec number is not
  safe until the board claim reaches `main`.** Either the operator pushes the claim
  commit, or a sweep should expect to renumber and should therefore defer creating the
  spec directory until as late as possible. Note the sibling that took 111 is
  `111-claim-gate-branch-visibility` — a session working on precisely this hazard.
- Sweep-close and this runbook's status flip land in the same PR, there being no next
  branch to ride.
- **Harness/root-guard incompatibility, worked around.** The harness's background-job
  guard accepts writes only inside a worktree it entered itself, and `EnterWorktree`
  creates those under `.claude/worktrees/`. This repo's `root-guard-hook.mjs` sanctions
  only `.worktrees/` (covered by `.gitignore`) and rejects writes everywhere else under
  the root checkout — including the harness's own worktree. The two guards are mutually
  exclusive: a background job that isolates the harness way cannot write at all.
  Resolution: cut the task worktree at `.worktrees/task-187` with
  `git worktree add`, then enter it by **path** (`EnterWorktree` accepts an existing
  worktree), satisfying both guards at once. Any future background-job sweep on this repo
  must do the same. Worth a card against the hook, which should learn
  `.claude/worktrees/`.

## Lane 0 — PRECONDITION (blocking, resolved in-branch)

**`docs/design/tui/anatomy.md` pins a commit that does not exist**, and this blocks any PR
touching the TUI surface — TASK-187 included. Evidence:

```
$ node scripts/check-tui-design.mjs
pins  docs/design/tui/anatomy.md  verified_against unresolvable:
      4eb6471ae06298f4fce438c98d8169c9e47e6308 is not a known commit
1 violation(s)                                                    # exit 1
```

The `pins` check is whole-repo, not `--changed`-scoped, so
`check-tui-design.mjs --changed <any-range>` exits 1 regardless of the branch diff
(verified against an empty range). `check-merge-drift.mjs pr` delegates wholesale to that
checker and raises **`tui-design-stale` at severity `block`** whenever the branch touches
`internal/tui/`, `docs/design/tui/`, or any design-pinned source. There is no bypass flag.

**Ruling (operator, 2026-08-02): force through it locally — but the repair turned out to
be honest, so nothing is faked.** Grounds:

- `anatomy.md` is `class: index` and declares **no `sources:`** — the checker only
  requires `verified_against` to resolve, so there is no source-diff to classify under the
  honest-re-pin rule, and no prose claim is being back-dated.
- `4eb6471a` was a branch commit orphaned when its PR squash-merged — the
  squash-rewrites-pins hazard CLAUDE.md names explicitly. The content it verified landed
  on main as `012032fb` ("spec 074 T024: amend docs/design/tui for the look-cursor mode"),
  the newest commit touching the file and reachable from main.
- Commit `6318cf8b` performed exactly this repair for six pages before: "re-pin 6
  spec-054-verified pages to the squash commit d220645 — branch-hash pins orphaned by PR
  #101's squash-merge".

Re-pin target: **`012032fb`**. This is the only dead pin in the design surface — every
other page pins `72f82f41`, which resolves.

## Per-PR gates this project enforces (enumerated — implementers cannot miss these)

- **Merge-drift gate: present at `scripts/check-merge-drift.mjs`.** Mandatory at every
  choke point, invocations verbatim:
  - `node scripts/check-merge-drift.mjs session` — at sweep start (done: `warnings`, exit 0).
  - `node scripts/check-merge-drift.mjs claim --dir 112-tui-frame-harness` — before
    creating the spec dir (done: `pass`).
  - `node scripts/check-merge-drift.mjs worktree --spec 112 --task TASK-187` — before
    `git worktree add` (done: `warnings`, `card-not-claimed` as explained above).
  - `node scripts/check-merge-drift.mjs pr` — from the worktree before `gh pr create` AND
    after every history move. Nonzero exit blocks.
- **TUI design authority (spec 047):** `node scripts/check-tui-design.mjs --changed` must
  pass, and `docs/design/tui/` is amended in this same PR (re-verify + re-pin every
  affected page). See Lane 0.
- **Wiki-in-PR lifecycle (spec 069):** if the branch touches any file a wiki note lists in
  `sources:`, that note is re-verified and re-pinned **on this branch**
  (`wiki-repin-missing` blocks otherwise); if the branch changes `docs/wiki/`,
  `docs/player/` is regenerated (`player-docs-stale` blocks otherwise). Probe:
  `node .claude/skills/player-docs/scripts/check-freshness.mjs --check`.
- **Merge-commit-only:** land with `gh pr merge --merge`. A squash rewrites in-branch
  re-pins out of main's history — the precise hazard that produced Lane 0.
- **Go hygiene:** `gofmt -l` clean, `go build ./...`, `go test ./...` green before PR.

## Per-task artifacts required before PR

**No PR opens for TASK-187 until each line below checks true.**

- [ ] `specs/112-tui-frame-harness/spec.md` — requirements mapped to the card's 9 ACs.
- [ ] `specs/112-tui-frame-harness/plan.md` — checked against `.specify/memory/constitution.md`.
- [ ] `specs/112-tui-frame-harness/tasks.md` — real phased breakdown, not one catch-all box.
- [ ] The card carries its machine-findable Spec marker (`spec-bridge:link`).
- [ ] Lane 0's dead pin repaired; `node scripts/check-tui-design.mjs` exits 0.
- [ ] Determinism proven: the scene matrix regenerated twice is byte-identical (AC #3).
- [ ] Harness-vs-`View()` equivalence test present and passing (AC #9).

## Model tier

**TASK-187 — Opus 5 (`claude-opus-5`), fallback `claude-opus-4-8`** — dispatched as the
`spec-implementer-opus` agent definition, never inline, and never via a `model` parameter
(observed to be silently ignored; the agent definition's frontmatter is the pin).

Rubric justification (constitution Principle V v1.3.0): the task is **cross-package and
architectural** — it adds a new `cmd/promptworld` verb, opens headless-construction seams
in `internal/tui`, introduces a deterministic fixture layer touching
`internal/world` / `internal/sim` / `internal/clock`, and must hold a determinism invariant
across all of them. That is the escalation tier, not a single-package feature.

Record on the board at dispatch: tier, model ID, rubric justification, and which model
actually served.

## Concurrency doctrine

- Conflict hotspots for this task: `docs/design/tui/**` (the design authority every TUI
  task amends), `cmd/promptworld/commands.go` (the command registry — every new verb
  touches it), `internal/tui/tui.go`.
- **This branch carries pins** (Lane 0's re-pin, plus any wiki re-pin). It is therefore a
  **pin-carrying branch**: freshen it by merging `origin/main` IN, never rebase, never
  force-push, and land it as a **merge commit**.
- A merge-in licenses no pin bump. Classify each staled pin RE-PIN-ONLY vs NEEDS-REVIEW
  against `git diff <old-pin>..<merge-commit> -- <sources>` before touching it.
- Re-run gates AND the freshness probe after every history move, unconditionally.

## Operator checkpoints — never proceed silently past

1. **Lane 0's pin repair** — resolved 2026-08-02: repair honestly to `012032fb`.
2. **PR merge** — the background-job mode forbids this session merging or pushing `main`.
   The PR is opened and left for the operator, as is the unpushed claim commit.
3. Any tier escalation beyond Opus 5, or any softening of a gate enumerated above.

## Done means

- TASK-187 Done on the board via its own merged PR.
- `specs/112-tui-frame-harness/` carries spec.md + plan.md + tasks.md; the card keeps its
  Spec marker.
- `node scripts/check-tui-design.mjs` exits 0 on main; `check-merge-drift.mjs session` no
  longer reports `tui-design` stale.
- `go build ./...` and `go test ./...` green on main.
- `git worktree list` shows no stale sweep worktrees.
- This runbook's execution log complete and its status flipped to done.

## Execution log

| Task | Tier / model | Phases | PR | Merge sha | Date | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| TASK-187 | Opus 5 / `claude-opus-5` | pending | — | — | 2026-08-02 | Lane 0 pin repair folded in |

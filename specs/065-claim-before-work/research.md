# Research: Claim-Before-Work Protocol (spec 065)

All unknowns in Technical Context resolved. Each decision below names alternatives
considered and why they lost.

## D1 — Where the claim-time spec-number block lives: new `claim` mode

**Decision**: add a fourth mode, `claim --dir <NNN-slug>`, to
`scripts/check-merge-drift.mjs`. Fetch first; unreachable remote → exit 2 (fail
closed, same posture as `worktree`/`pr`). Collision rule: block iff
`takenSpecNumbers(origin/main)` has NNN mapped to a dirname *different from* the one
being claimed — re-using `specNumberCollisions()` semantics, so a session re-touching
its own already-pushed claim passes (idempotent for the owner).

**Rationale**: `takenSpecNumbers()` already computes exactly the right answer; the
task text's diagnosis is that it "runs too late to prevent anything." A dedicated mode
gives the hook (and humans) a single cheap invocation at the moment of directory
creation, and keeps the gate-cli contract's mode/flag/exit-code shape.

**Alternatives considered**:
- Overload `worktree --spec` for claim time: rejected — worktree mode also enforces
  root-at-tip, which is irrelevant (and wrong to require) when a session is writing a
  spec file mid-flight; the two gates fire at different moments.
- A standalone script: rejected — spec 051 FR-008/architecture keeps all gate
  semantics in the one CLI; the hook wrapper explicitly "defines no new gate
  semantics of its own."

## D2 — How the card-claim check reads board status: origin/main tree, not the working dir

**Decision**: `worktree --task TASK-<n>` locates the task's card file in the *fetched
`origin/main` tree* (`git ls-tree origin/main backlog/tasks/` matched by the existing
`^task-<n>[ .]` convention from `findTaskFile()`, content via `git show`), parses the
`status:` frontmatter line, and warns (`card-not-claimed`, severity warn) when it is
not `In Progress`.

**Rationale**: the protocol defines ownership as *presence on origin/main* — a local
card move that was never pushed is exactly the failure the check must catch, so
reading the local working dir would be self-defeating. Worktree mode already blocks
unless root is at the fetched tip, but tree-reading keeps the check honest even if
that ordering ever changes.

**Alternatives considered**:
- `backlog task view` CLI: rejected — reads local state, requires the CLI on PATH
  inside a gate that must stay dependency-free (contract: backlog CLI only for
  `--notes`).
- Block instead of warn: rejected — the task text says "should warn," deliberately:
  the claim commit itself needs a worktree-less window, and legitimate
  mid-propagation states exist. Blocking would fight the protocol it enforces.

## D3 — Unpushed-branch audit: session-mode warn keyed on missing remote ref

**Decision**: in `session` mode's per-branch scan, for each live `task-*` branch whose
tip is ahead of its merge base with origin/main, warn (`branch-unpushed`) when
`refs/remotes/origin/<branch>` does not exist after the fetch. Attribution via the
existing `attributeTask()` (branch name → TASK-n), so `--notes` can record it.

**Rationale**: AC #4's observable is "in-flight work is auditable from any clone";
the mechanical inverse is a local-only branch with commits. Session mode already
enumerates branches and fetches, so the check is one ref lookup per branch.

**Alternatives considered**:
- Enforce push at commit time (post-commit hook): rejected — spec 051 architecture is
  "no daemon, no CI"; harness hooks gate tool calls, not git plumbing, and a
  PreToolUse match on every `git commit` would be far too hot a path.
- Compare against upstream tracking config (`branch.<name>.remote`): rejected —
  tracking config can exist while the remote ref is deleted, and vice versa; the
  remote ref's existence is the ground truth for "auditable from any clone."

## D4 — Hook interception of spec-directory creation: pre-bash matcher + new pre-write subcommand

**Decision**: two entry points, both routing to `claim` mode:
1. `pre-bash` (existing subcommand) gains a matcher for Bash commands that create
   spec directories: `mkdir` (any flags) with a `specs/<NNN>-<slug>` path segment, and
   `create-new-feature.sh` invocations (which take `--number N --short-name <slug>`).
2. New `pre-write` subcommand wired in `.claude/settings.json` as a PreToolUse hook
   with matcher `Write|Edit`: extract `specs/(\d{3,})-([^/]+)/` from
   `tool_input.file_path`; when present, run `claim --dir NNN-slug` from the file's
   repo. Exit ≥ 1 from the gate → hook exits 2 (block, findings on stderr); all other
   failures fail open, matching pre-bash's posture exactly.

**Rationale**: spec directories are born two ways in practice — a Bash `mkdir` (the
claim stub) or a Write of `specs/NNN-slug/spec.md` (Spec Kit authoring). Intercepting
both closes the race at every real entry point without touching git plumbing.
`Write`/`Edit` to files in an *existing, owned* spec dir pass because of D1's
same-dirname idempotence — a session editing its own spec never hits the block, and
the fetch cost is bounded to writes matching the specs-path shape.

**Alternatives considered**:
- Only doctrine, no Write interception: rejected — AC #3 says "blocked at claim time,
  not detected after"; the Write path is how the 063 collision actually happened.
- Intercept `git push` instead: rejected — push rejection already IS the mechanism
  there; the gate's job is to stop the *local* act of authoring against a taken
  number before work accumulates.
- Caching fetches to avoid per-write latency: deferred — writes matching
  `specs/NNN-*/` are rare (a handful per feature); measure before optimizing.

## D5 — Doctrine placement: CLAUDE.md block + runbook template concurrency section

**Decision**: a dedicated "Claim-before-work protocol (spec 065)" block in this
repo's CLAUDE.md, adjacent to the worktrees block (the claim commit immediately
precedes worktree cutting in the lifecycle). Companion edit in the praxisflux source
repo (`~/neumo/projects/praxis`): the sweep runbook template
(`pdlc/skills/sweep/templates/runbook.md`) gains the same three-rule protocol in its
"Concurrency & conflict doctrine" section, so every future runbook instructs
executing sessions verbatim.

**Rationale**: CLAUDE.md is the always-on grounding for every session in this repo;
the runbook template is how sweep orchestrators (often *other* sessions) receive
doctrine. AC #1 names both. The praxis repo edit follows that repo's recorded laws
(version-lockstep, merge-commit-only, per-task-course) and is recorded on TASK-139 as
a companion change — it is not part of this repo's PR.

**Alternatives considered**:
- Editing the marketplace clone of the plugin: rejected — memory records the source
  repo as `~/neumo/projects/praxis`; the marketplace copy is a distribution artifact.
- Doctrine only in the runbook template: rejected — most sessions in this repo are
  not sweep sessions; CLAUDE.md is the only surface they all read.

## D6 — Race simulation: node --test fixture with bare origin + two clones

**Decision**: new `scripts/claim-protocol.test.mjs` under the existing `node --test`
harness (precedent: `check-merge-drift.test.mjs`, TASK-138). Fixture: `git init
--bare` origin in tmpdir, seed a minimal repo shape (backlog/tasks card file, specs/,
scripts copied or pathed), clone A and clone B. Test sequence:
1. A executes the claim (card status flip + `specs/NNN-slug/` stub), pushes — accepted.
2. B commits a competing claim (same task card edit + same NNN, different slug),
   pushes — assert non-fast-forward rejection.
3. B fetches; assert `claim --dir NNN-other` exits 1 naming A's dir, and
   `worktree --task TASK-<n>` on a task whose card A did *not* claim warns.

**Rationale**: AC #5 accepts "test or documented manual run"; an automated test is
strictly stronger and the harness precedent already exists. Everything needed
(subprocess git, tmpdir fixtures) is stdlib.

**Alternatives considered**:
- Documented manual run only: rejected — races regress silently; five collisions in
  one day earn a regression test.
- Testing via the hook layer end-to-end: rejected as primary coverage — the hook is a
  thin wrapper with fail-open semantics that would mask gate regressions; test the
  gate modes directly, and cover the hook's matchers with focused unit cases if cheap.

## D7 — Contract handling: delta document in this spec, 051 contract untouched

**Decision**: normative additions live in
`specs/065-claim-before-work/contracts/gate-cli-delta.md`, written as a delta against
spec 051's `contracts/gate-cli.md` (new mode, new flags, new findings, unchanged
mutation whitelist). Spec 051's own contract files are not edited.

**Rationale**: spec directories are the source of truth *for their feature*;
retroactively editing 051's contract would blur which feature introduced what. The
gate script's header comment gains a pointer to both contracts.

**Alternatives considered**: amending 051's contract in place — rejected per above;
045/047-style precedent keeps deltas with the introducing spec.

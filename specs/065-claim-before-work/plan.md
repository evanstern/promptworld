# Implementation Plan: Claim-Before-Work Protocol

**Branch**: `065-claim-before-work` | **Date**: 2026-07-25 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/065-claim-before-work/spec.md`

## Summary

Turn git push rejection into the mutual-exclusion primitive for concurrent sessions.
Doctrine (CLAUDE.md + pdlc sweep runbook template) defines the claim: the first commit
of a task moves the board card to In Progress and creates the spec directory, pushed
immediately; a rejected push is a stop-the-lane signal. Enforcement extends the spec 051
gate architecture in place: `worktree` mode gains a `--task` card-claim check (warn),
a new `claim` mode blocks spec-number collisions at directory-creation time (fail
closed), `session` mode gains an unpushed-task-branch finding (warn), and the existing
PreToolUse hook wiring gains interception of spec-directory-creating tool calls. A
two-clone race simulation lands in the existing `node --test` harness.

## Technical Context

**Language/Version**: Node.js ≥ 18, ESM, zero npm dependencies (stdlib only) — matches
the spec 051 gate-cli contract for `scripts/check-merge-drift.mjs`

**Primary Dependencies**: git ≥ 2.38 (already required); `backlog` CLI only via the
existing `--notes` path (never required by new checks)

**Storage**: none — all state is read from git trees (`origin/main` tip) and the
working tree; board status parsed from tracked `backlog/tasks/*.md` frontmatter

**Testing**: `node --test` (precedent: `scripts/check-merge-drift.test.mjs`, TASK-138);
new race-simulation tests build throwaway bare-origin + two-clone fixtures in tmpdir

**Target Platform**: darwin + linux dev machines running Claude Code sessions; no CI,
no daemon (spec 051 FR-008 posture preserved)

**Project Type**: single repo — gate script + hook script + doctrine docs; companion
one-file edit in the praxisflux plugin source repo (`~/neumo/projects/praxis`)

**Performance Goals**: gate runs remain interactive (< a few seconds; one `git fetch`
per invocation, same as today)

**Constraints**: mutation whitelist of the gate-cli contract is unchanged — new modes
and checks are read-only except the already-whitelisted fetch; hook fails open on
everything except a failing gate; `claim` mode fails closed on unreachable remote

**Scale/Scope**: ~4 surfaces in `check-merge-drift.mjs` (arg parsing, worktree mode,
new claim mode, session branch scan), ~2 in `merge-drift-hook.mjs` (pre-bash matcher,
new pre-write entry), 1 settings.json hook wiring, 2 doctrine docs, 1 contract delta,
1 test file

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|---|---|---|
| I. Artifact-Grounded Action | PASS | The feature exists to make claims artifact-grounded (pushed commits); all decisions land in spec/plan/contracts + board |
| II. One Task, One PR | PASS | TASK-139 → one branch (`task-139-claim-before-work`) → one PR; the praxis-repo companion edit is a separate repository with its own laws, recorded on the board task (not a second PR in this repo) |
| III. Gates Over Assertions | PASS | The deliverable is itself gate machinery; statuses continue to be provable from artifacts |
| IV. Grounding Freshness | PASS | `docs/wiki/` notes listing CLAUDE.md or the gate scripts as sources must be re-pinned in the PR flow (checked at implementation; wiki-update if touched) |
| V. Model-Tiered Workflow | PASS | Plan/spec on Fable 5; implementation delegated to spec-implementer. Tier: **Opus 4.8** — gate/enforcement logic is doctrine-adjacent behavior change per the escalation rubric |

**Post-Phase-1 re-check**: PASS — design adds no projects, no dependencies, no daemon;
all new behavior is inside the two existing scripts plus docs. No Complexity Tracking
entries needed.

## Project Structure

### Documentation (this feature)

```text
specs/065-claim-before-work/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Phase 0 output — decisions + rationale
├── data-model.md        # Phase 1 output — findings, rules, claim entities
├── quickstart.md        # Phase 1 output — validation runs
├── contracts/
│   └── gate-cli-delta.md  # Normative delta to spec 051's gate-cli contract
├── checklists/
│   └── requirements.md
└── tasks.md             # Phase 2 output (/speckit-tasks)
```

### Source Code (repository root)

```text
scripts/
├── check-merge-drift.mjs        # + claim mode, worktree --task, session unpushed-branch
├── check-merge-drift.test.mjs   # existing harness (untouched)
└── hooks/
    └── merge-drift-hook.mjs     # + pre-bash spec-dir matcher, + pre-write subcommand

scripts/claim-protocol.test.mjs  # NEW — two-clone race simulation (node --test)

.claude/settings.json            # + PreToolUse Write|Edit hook wiring

CLAUDE.md                        # + claim-before-work doctrine block

# Companion (separate repo, ~/neumo/projects/praxis — its own laws):
pdlc/skills/sweep/templates/runbook.md   # + claim doctrine in concurrency section
```

**Structure Decision**: everything stays in the two existing gate files plus one new
test file — the spec 051 architecture (single-file gate CLI + single-file hook wrapper)
is preserved; no new packages or directories beyond the test.

## Design decisions (summary — full rationale in research.md)

1. **New `claim` mode** in `check-merge-drift.mjs` (`claim --dir <NNN-slug>`): fetch
   (fail closed, exit 2 on unreachable remote), block (exit 1) iff number NNN is taken
   on `origin/main` under a *different* dirname — idempotent for the claim's owner.
2. **`worktree --task TASK-<n>`**: parse the task's card file from the fetched
   `origin/main` tree (`git ls-tree` + `git show`, never the local working dir), warn
   when frontmatter `status:` is not `In Progress`. Hook derives `--task` from the
   `git worktree add` command's branch/dir naming (`task-<n>-…` / `.worktrees/task-<n>`).
3. **`session` unpushed-branch finding**: for each live `task-*` branch with commits
   past its merge base, warn when `refs/remotes/origin/<branch>` does not exist.
4. **Hook interception of spec-dir creation**: pre-bash gains a matcher for
   spec-directory-creating Bash commands (`mkdir …specs/NNN-*`, `create-new-feature.sh`);
   a new `pre-write` subcommand (wired via a `Write|Edit` PreToolUse matcher in
   `.claude/settings.json`) extracts `specs/NNN-slug/` from `tool_input.file_path`.
   Both route to `claim` mode; block on exit ≥ 1, fail open otherwise.
5. **Doctrine**: new CLAUDE.md block adjacent to the worktrees section; companion edit
   to the runbook template's "Concurrency & conflict doctrine" section in the praxis
   source repo (version-lockstep + merge-commit-only laws apply there).
6. **Race simulation**: `scripts/claim-protocol.test.mjs` builds a bare origin + two
   clones in tmpdir; clone A claims and pushes; clone B's push is rejected; after
   fetch, clone B's `claim` mode blocks and `worktree --task` warns. Run with
   `node --test scripts/claim-protocol.test.mjs`.

## Complexity Tracking

No constitution violations — table not needed.

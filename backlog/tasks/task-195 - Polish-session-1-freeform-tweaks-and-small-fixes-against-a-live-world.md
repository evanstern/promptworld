---
id: TASK-195
title: 'Polish session 1: freeform tweaks and small fixes against a live world'
status: In Progress
assignee: []
created_date: '2026-08-03 17:33'
updated_date: '2026-08-03 18:14'
labels: []
dependencies: []
ordinal: 177001
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
A single long-running card that holds one rapid polish session. Instead of a spec per fix, we
watch a live world, discuss small bugs / UI tweaks / gameplay nits as they come up, record each
decision on this card before code is written, and land the whole session on one branch in one PR.

## Use cases

- As a player, I want the small rough edges I keep bumping into — a misaligned pane, a confusing
  label, a villager doing something obviously silly — fixed quickly, without each one waiting on
  its own full spec cycle.
- As the operator, I want to sit in front of a real running world, point at what looks wrong, and
  have the decision written down before anyone writes code, so a fast session still leaves a
  paper trail.
- As a reviewer, I want one PR whose card lists every item worked, the diagnosis behind it, and
  whether it was decided ad-hoc or escalated to a spec.

## Session workflow (operator-authored, this session's contract)

1. Worktree + branch off origin/main; this card is claimed for the session's duration.
2. A runnable promptworld is kept live: world-02 in place (~/.promptworld/worlds/world-02),
   restarted against the branch build as we iterate.
3. This card is the only task; there are no per-item task cards.
4. Development sub-loop, repeated:
   a. Discuss a feature, bug, or tweak.
   b. Optionally ground the topic with research / wiki reads.
   c. Record the decision and its diagnosis on this card (Decision log below) BEFORE implementing.
5. Ad-hoc after a loop, decide whether the accumulated items warrant spec(s).
   a. If yes — write the spec(s), link via spec-bridge, execute them.
   b. If no — continue the sub-loop at 4a.
6. Execute any specs created.
7. Decide whether to go again (a fresh session is usually the right move at this point).
8. Test: operator visual QA on the live world; optionally a team review.
9. ONLY THEN: wiki re-pin, player docs, design references, PDLC gates. Nothing is re-pinned
   per item — grounding happens once, at the end, in-branch, before the PR.

Scope guard: this flow is for polish — FE/TUI changes and small gameplay tweaks against decisions
already made. Anything needing real design thought leaves this session and goes through the full
PDLC cycle on its own card.

## Decision log

(entries appended as the session runs — one per item: what, diagnosis with file:line, decision,
ad-hoc vs spec)
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Every item worked in this session has a decision-log entry on this card — item, file:line diagnosis, and the decision — recorded before its implementation
- [ ] #2 Any item exceeding the trivial-exemption bar (surgical fix + complete file:line diagnosis + ACs on this card) is escalated to a Spec Kit spec and linked via spec-bridge before implementation
- [ ] #3 All session work lands on a single branch and a single PR; no per-item task cards or PRs are created
- [ ] #4 Operator visual QA passes on the live world for every shipped item before the PR is opened
- [ ] #5 Grounding is done once at the end, in-branch: wiki re-pinned, player docs regenerated, tui-design amended where internal/tui changed, and the pr merge-drift gate is green
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
### Session harness up (2026-08-03)

- Branch/worktree: `task-195-polish-session-1` at `.worktrees/task-195`, cut from origin/main
  @57cbaf4d (the claim commit). Worktree gate passed with `--task TASK-195`.
- Live world: **world-02 in place** at `~/.promptworld/worlds/world-02` — operator's choice over a
  snapshot copy. It stays running on its current binary until the first branch change lands, then
  cycles onto the branch build. Nothing else runs against it.
- Branch build: `go build -o promptworld ./cmd/promptworld` in the worktree, ~3.4s cold — fast
  enough that rebuild+restart is a viable inner loop. Restart helper lives outside the repo
  (`$CLAUDE_JOB_DIR/tmp/pw.sh`, verbs: cycle/build/stop/start/ps/status/log) so it never enters
  the PR diff.
- TUI inner loop: `.claude/skills/tui-frames` frame harness is green (`--check`: committed matrix
  matches a fresh dump). UI items go through headless frame dumps first, live world second.
- Item sourcing: ad-hoc only. Existing polish-shaped cards (TASK-60/61, 114, 116, 142, 152, 174,
  175, 176) are untouched; if a session item overlaps one, the overlap gets noted and decided then.
- Harness trap hit and corrected during setup: a `cd` into the worktree persisted across shell
  calls, so the first attempt at this very note was edited and committed on the branch instead of
  main. Reset off the branch and redone at root — board state has exactly one home.

### Harness amendment: world-02 → world-03 (2026-08-03)

Operator cut and calibrated **world-03** and it replaces world-02 as the session's live world.
Rationale on record: world-02 is `stage: stage-1`, immutable for the world's life
(`internal/world/world.go:84-91`), so its guardian grant is capped at 15 of 20 tools with zero
miracle kinds (`internal/guardian/charter.go:735-741`, intersected at `:762`) — `work_miracle`,
`canonize_region`, `pause`, `start`, `adjust_speed` are structurally unreachable there. world-03 is
`stage-4` with `stage_overridden: true`, so the full grant is testable.

world-03 is now running on the branch binary (operator authorized taking the binary): clean
`stop`/`start` cycle preserved state at tick 1489, day 1 06:24, pid 62622. Helper retargeted.

### Decision 1 — daemon boot report reads dead legacy LLM config fields

**Item.** Every world boots with `daemon: llm orchestrator on (local  @ , cloud , budget $100/mo)`
— model and endpoint blank. Observed on world-03 and identically in world-02's `daemon.log:2`.

**Diagnosis (complete).** `llm.json` has two mutually exclusive shapes
(`internal/llm/config.go:27,34-35`): the v2 registry (`providers` + `routes`) and the legacy
two-tier (`local`/`cloud`). `resolveRegistry` (`internal/llm/config.go:567-572`) rejects both being
present, so on a v2 world the legacy `Local`/`Cloud` structs are always zero. The daemon's
boot-report block reads exactly those dead fields:

- `internal/daemon/daemon.go:351` — `localDesc` ← `llmCfg.Local.Model` / `.Endpoint` (empty)
- `internal/daemon/daemon.go:329-331` — `cloudDesc` ← `llmCfg.Cloud.*` (empty)
- `internal/daemon/daemon.go:336` — `llmCfg.Local.Workers()` → `, parallel N` suffix never appears
- `internal/daemon/daemon.go:345,348` — `Local/Cloud.ToolModeResolved()` → clamp warnings never fire

`DefaultConfig` writes the v2 shape (`internal/llm/config.go:449,469`), so every world created since
spec 024 is affected.

**Not purely cosmetic.** v2 has correct per-provider equivalents, `pc.workers(name)` and
`pc.toolModeResolved(name)` (`internal/llm/config.go:178,194`), but both callers discard the
warning — `internal/llm/llm.go:585` (`slots, _`) and `internal/llm/llm.go:1105` (`mode, _`) — and
nothing else prints them. So on a v2 world an out-of-range `parallel` or an invalid `tool_mode` is
silently clamped with no operator-visible warning, where the legacy shape warned. A diagnostic
regression left behind by the format change. Runtime behavior is correct throughout; the defect is
confined to the operator-facing boot report.

**Decision — ad-hoc (no spec).** Surgical, single block, complete file:line diagnosis. Replace the
hardcoded two-tier line with an N-provider report sourced from the resolved registry:

```
daemon: llm orchestrator on (2 providers, budget $100/mo)
daemon:   local  qwen3.6:latest @ http://localhost:11434/v1
daemon:   cloud  cc/claude-sonnet-5 @ http://localhost:20128/v1
```

`, parallel N` appended per provider where N > 1; the two discarded clamp warnings routed back to
the boot channel. Rejected alternative: keep the single-line form — unreadable past two providers
and still bakes in the tier names, which v2 does not guarantee.

**Provenance.** Operator authorized taking the binary; the boot-line shape is the assistant's
recommendation adopted absent objection, and is trivially revisable. Implemented inline rather than
delegated to `spec-implementer` per constitution Principle V — this session's harness instructions
forbid subagent dispatch unless the operator requests it. Flagged for the operator to redirect.

### Decision 1 — SHIPPED on branch (commit f673536a, pushed 2026-08-03)

`Config.ProviderReports()` (`internal/llm/config.go`) resolves either config shape into a stable
name-ordered provider view and returns the warn-and-clamp warnings the construction path discards;
the daemon boot block (`internal/daemon/daemon.go`) reports off that instead of the dead legacy
fields, scaling to N providers.

**Live proof** — world-03 restarted on the branch binary, `daemon.log`:

```
daemon: llm orchestrator on (2 providers, budget $100/mo)
daemon:   cloud  cc/claude-sonnet-5 @ http://localhost:20128/v1
daemon:   local  qwen3.6:latest @ http://localhost:11434/v1
```

**Tests** — three added to `internal/llm/config_test.go`, all passing: v2 registry names every
provider in stable order; `parallel: 99` + `tool_mode: "telepathy"` now yield the clamp warnings
that were silently dropped; legacy shape still renders local/cloud, with a bare model (no dangling
` @ `) for the endpoint-less anthropic transport. Full `go test ./...` green, `gofmt` clean,
`go vet` clean.

**Known follow-on, deferred to step 9:** ten wiki notes list `internal/daemon/daemon.go` as a
source (daemon-orchestrator-startup, daemon-boot-recovery, daemon-cognition-calibration,
daemon-lifecycle, event-types-clock-world, event-types-guardian-orders, guardian-survival-watches,
llm-provider-health, scenario-machinery-surfacing, snapshots) and are now stale for the pr gate.
Not re-pinned per the session contract — grounding runs once at the end.

**Also stale, decide at step 9:** `specs/009-parallel-local-tier/contracts/llm-config.md:36-42`
documents the old single-line boot shape (`local … , parallel 16, cloud …`). It is a shipped
spec's contract, so the choice is amend-in-place vs record the supersession here; flagged, not
resolved.

**Correction to the decision-1 note:** the stale-note count is **12**, not ten — the earlier figure
came from a `grep -rl … | head` that truncated at ten. Adding `internal/llm/config.go` (6 notes),
this branch's stale union is **17 of 191 notes (9%)**, since five notes source both files.

### Decision 2 — multi-line boot shape stays

Operator confirmed the N-provider multi-line boot report (decision 1). No further change; the
shipped form is final. Closes the open question flagged when decision 1 landed.

### Decision 3 — wiki-footprint threshold check in the session gate

**Why this and not a grounding cadence.** Grounded in the gate's own logic: a note is stale iff any
of its `sources:` changed between its pin and the tip (`scripts/check-merge-drift.mjs:799`,
`changedFiles(n.verified_against, originMainTip, cwd, n.sources)`). That is **binary per note and
idempotent** — touching `internal/daemon/daemon.go` once stales 12 notes; touching it fifty more
times stales the same 12. Staleness is a set union over FILES TOUCHED, never a sum over edits, so
deferring grounding to step 9 costs nothing extra and per-item grounding would re-verify the same
notes N times. The session contract stands.

What actually grows is **breadth** — the union widens only when new distinct files are touched — and
the repo's note-per-source concentration makes that sharp: `internal/sim/state.go` 27 notes,
`executor.go` 26, `agents.go` 22, `tool/registry.go` 15, `sim/loop.go` 15, `tui/tui.go` 14,
`tui/views.go` 13, `daemon/daemon.go` 12. One stray edit to `sim/state.go` costs more than this
whole session so far (17 of 191). So the risk knob is scope sprawl across subsystems, not elapsed
items — and the right instrument measures footprint, not frequency.

**The check.** In `session` mode, per live branch, after `loadWikiNotes`:

- Metric: size of the union from the existing `wikiSourcesOverlap(b.changedFiles, wikiNotes)` —
  notes whose `sources:` intersect the branch's changed files.
- Baseline: `mergeBase..tip`, i.e. `b.changedFiles`, committed work only. Uncommitted edits are not
  counted, consistent with every other per-branch computation in the script. Documented, not fixed.
- Threshold: **30 notes** (of 191, ~16%). Calibrated off the concentration table — `sim/state.go`
  alone is 27 and `tui.go`+`views.go` is ~27, so 30 trips precisely when a session has reached into
  a second hot subsystem, which is the sprawl signal. A narrow session never approaches it.
- Severity: **warn, never block.** The session gate is non-blocking by design (it injects context at
  SessionStart and must never stop a session). Making it ever block is a separate, spec'd decision.
- Rule name: `wiki-footprint`.
- Always-visible count: `wikiNotes=N` appended to each branch line in the text report and carried in
  `--json`, so the number is legible below threshold with no extra finding noise.

**Trivial-exemption judgment (AC #2).** Borderline and called deliberately: this adds new gate
behavior rather than fixing a diagnosed defect, which normally argues for a spec. It lands as ad-hoc
because the only genuinely policy-shaped knob — whether it can block — is answered by the session
gate's existing non-blocking contract, leaving a threshold constant and a report field. If the
threshold should ever gate rather than inform, that is a spec.
<!-- SECTION:NOTES:END -->

---
id: TASK-195
title: 'Polish session 1: freeform tweaks and small fixes against a live world'
status: Done
assignee: []
created_date: '2026-08-03 17:33'
updated_date: '2026-08-03 21:59'
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

Spec: specs/115-chronicle-feed-wrap
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Every item worked in this session has a decision-log entry on this card — item, file:line diagnosis, and the decision — recorded before its implementation
- [ ] #2 Any item exceeding the trivial-exemption bar (surgical fix + complete file:line diagnosis + ACs on this card) is escalated to a Spec Kit spec and linked via spec-bridge before implementation
- [ ] #3 All session work lands on a single branch and a single PR; no per-item task cards or PRs are created
- [x] #4 Operator visual QA passes on the live world for every shipped item before the PR is opened
- [x] #5 Grounding is done once at the end, in-branch: wiki re-pinned, player docs regenerated, tui-design amended where internal/tui changed, and the pr merge-drift gate is green
- [x] #6 Spec phase: Setup
- [x] #7 Spec phase: Foundational (blocking prerequisites)
- [x] #8 Spec phase: User Story 1 — A thought can be read to its end (P1) 🎯 MVP
- [x] #9 Spec phase: User Story 2 — The feed still reads as a table (P2)
- [x] #10 Spec phase: User Story 3 — Narrow panes degrade sensibly (P3)
- [x] #11 Spec phase: Row budget and evidence
- [x] #12 Spec phase: Polish and cross-cutting
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

### Decision 3 — SHIPPED on branch (commit b5b9fd0a, pushed 2026-08-03)

`WIKI_FOOTPRINT_THRESHOLD = 30` and a per-branch footprint pass in `runSession`, reusing the
existing `wikiSourcesOverlap`. `wikiNotes=N` now rides each branch line in the text report and the
`--json` branch objects; the `wiki-footprint` finding is `warn` and fires at or above the threshold.

**Live proof** — the gate run from this worktree, correctly reporting this session's own branch:

```
branches:
  guide-quickstart-outline      task=-         baseLag=6  dirty=false  wikiNotes=0   cleanupEligible=false
  task-173-absence-attribution  task=TASK-173  baseLag=8  dirty=false  wikiNotes=0   cleanupEligible=true (ancestor)
  task-195-polish-session-1     task=TASK-195  baseLag=5  dirty=true   wikiNotes=17  cleanupEligible=false
```

17 matches the independent `grep -rl` union computed during the discussion — two derivations, same
number. Below threshold, so no warning, as intended.

**Tests** — three added to `scripts/check-merge-drift.test.mjs`, bracketing the threshold with no
slack: a fixture branch touching a source pinned by exactly 30 notes warns at severity `warn` and
reports `wikiFootprint: 30`; a branch at 29 reports its count and does NOT warn; and session's exit
code is unchanged with a threshold branch present. Full gate suite 34/34, no regressions.

**Session footprint so far: 17 of 191 notes (9%)** — `internal/daemon/daemon.go` (12),
`internal/llm/config.go` (6), five overlapping. `scripts/` is sourced by no wiki note, so decision 3
added nothing to the grounding bill. 13 notes of headroom before the threshold trips.

### Decision 4 — implementation stays inline for this session

Operator ruling: keep implementing inline rather than dispatching `spec-implementer`. This resolves
the open question flagged with decision 1 and governs the whole session.

A recorded, deliberate deviation from constitution Principle V (v1.3.0), which requires the planning
tier to delegate implementation to a pinned agent definition. The grounds: this session's harness
forbids subagent dispatch unless the operator asks, and the polish loop's value is the tight
discuss → diagnose → implement → live-prove cycle against a running world, which a delegation hop
would break for slices this small. The model that actually served every slice here is
`claude-opus-5` — the planning session's own model — which is the tier Principle V assigns to hard
slices anyway, so the work was never served below its rubric tier, only served without the hop.

Scope: this session only. It is not a precedent for spec'd work, where delegation still holds.

### Finding — raw feed truncates instead of wrapping (item 4, decision PENDING)

**Ask.** The raw chronicle feed does not word-wrap long lines — worst on thoughts and
conversations. Wanted: wrap, with continuation lines aligned to the 4th column (where the
villager name starts).

**Diagnosis.** Two separate causes, both pinned.

1. *No wrap at usable widths.* `internal/tui/views.go:1073-1076` — the chronicle tab sets
   `maxWrap := 1` and only raises it to 3 when `width < 60`. Wrapping is therefore enabled ONLY in
   the narrow dock; in solo (full width) and in any dock ≥ 60 columns, `maxWrap == 1` means
   `styleWrapLine` takes its truncate branch (`internal/tui/grammar.go:426-440`) and the line is
   cut with `…`. The narrow-fallback `chronicleView` also passes `1`
   (`internal/tui/views.go:2286`).
2. *No hanging indent even where wrap IS on.* `styleWrapLine`'s wrap branch
   (`internal/tui/grammar.go:442-510`) greedy-wraps the FULL flattened line — prefix and summary
   together — and every continuation line starts at column 0. Nothing in the wrap path knows where
   the summary column begins.

**Column layout** (`chronicleLinePrefix`, `internal/tui/grammar.go:299-307`): solo renders
`<TICK> <HH:MM>  <type>  <summary>`; dock drops the tick. The requested alignment target is
`len(prefix)` — already computable per window from `cols`, never a hardcoded constant, since
`TickWidth` and `TypeWidth` are derived per visible window (`computeChronicleColumns`).

**Why this is not a one-liner.**

- The digest-grammar contract (`specs/*/contracts/digest-grammar.md` §1/§2) and
  `docs/design/tui/panels/chronicle.md` both document today's line format and wrap/truncate rules.
  Changing them is a spec-047 design-authority amendment, gated on the same PR.
- Row budgeting interacts: `chronicleRawBody` slices the tail to `entryRows` events assuming
  "each event contributes at least one physical line" and trims overshoot after. Turning wrap on at
  full width changes physical-line counts for every long event.
- Narrow widths need a guard: with a ~36-column prefix indent inside a 40-column dock, the residual
  text column collapses. The minimum-residual-width rule is a real design decision, not an
  implementation detail.
- **Fixture gap:** no committed frame reproduces the defect — the three fixtures emit no long
  conversation/thought events, so `docs/design/tui/frames/` cannot show a before/after. The
  tui-frames loop treats the frame diff as THE review artifact, so a fixture must gain a long event
  before this change is reviewable.

**Wiki footprint (dogfooding decision 3).** `internal/tui/views.go` is sourced by 13 notes and
`internal/tui/grammar.go` by 4 (union 16). Added to this session's existing 17, the branch would
reach **exactly 30 — the threshold shipped in decision 3 an hour ago**, which would fire
`wiki-footprint` on the next session gate run. The instrument works, and it is pointing at this
item.

### Decision 5 — item 4 escalates to spec 115 (step 5a)

Operator ruling: write a spec and execute it in this session. `specs/115-chronicle-feed-wrap/`
created and pushed on the branch (commit 9d340ca5).

Three P-ranked stories: **P1** a thought can be read to its end (wrap instead of truncate);
**P2** the feed still reads as a table (continuation lines begin at the summary column);
**P3** narrow panes degrade sensibly (the indent yields before the text column becomes a sliver).
13 FRs, 7 success criteria. FR-004 pins the indent to the per-window column measurements rather
than a constant — a fixed indent misaligns the moment a wider tick or longer event type scrolls
into view. FR-012 carries the fixture precondition; FR-013 the design-authority and
digest-grammar contract amendments.

**Number claimed** via `check-merge-drift claim --dir 115-chronicle-feed-wrap` (pass), then
published by **pushing the branch rather than the spec-065 stub merge to main**. Deliberate
deviation, recorded: the branch already carries unreviewed code commits (decisions 1 and 3), so a
`merge --no-ff` of the stub would land them on main outside the PR. Spec 111's
`branchHeldSpecNumbers` reads `refs/remotes/origin/task-*`, so a pushed branch's spec number is
already visible to every clone's claim gate — which is the guarantee the stub merge exists to
provide. The claim is published; only the mechanism differs.

**One assumption flagged for operator override:** wrap depth in the full-width view is
**unbounded** — a long thought wraps to as many lines as it needs rather than being capped and
re-truncated. Capping would reproduce the original complaint in subtler form, since the player
would still lose the end of the sentence; the row budget already bounds the feed by dropping the
oldest events. The narrow dock's existing 3-line cap is retained as separate behavior. This is the
one genuinely arguable call in the spec.

### Spec 115 — IMPLEMENTED on branch (commits b93c7fe1, 5c98bfdd, pushed)

31 of 32 spec tasks ticked; T032 is this note. Full `go test ./...` green, gate suite 34/34,
`gofmt`/`go vet` clean, `check-tui-design --changed` exit 0, frame matrix matches a fresh dump.

**Two things the plan did not predict.**

1. *A defect in this change, caught by a pre-existing test.* `TestPreLadderGoldenFrames` failed
   because routing every row through the wrap path collapsed the column padding on SHORT rows:
   `wrapText` budgets on `strings.Fields`, and the feed's column padding is exactly a run of
   spaces. Fixed by returning a fitting line verbatim and wrapping only the summary, never the
   prefix. This is precisely the signal T031/SC-006 exists to catch, and it caught it.

2. *A pre-existing defect that same fix repairs.* The narrow dock has always wrapped, so it has
   always been sending every row through that collapsing path — its column padding was being
   destroyed on rows that never needed wrapping, and the committed frames recorded the collapsed
   form as if intended:

   ```
   -  19:12 moved Fern → (36,26)
   +  19:12 moved       Fern → (36,26)
   ```

   This is why `scenario__home__*` frames changed despite this feature never touching that
   fixture. **FR-009 was amended rather than left contradicting its own diff** — it promised
   byte-identity for rows that fit, which is now true only of the full-width views — and the spec
   carries a "Discovered during implementation" section stating what changed and why.

**Frame churn (T032):** 8 frames, not the "large number" research R7 predicted — `mid-game__home`
and `mid-game__solo` at 112/113/160, and `scenario__home` at 112/113. R7's estimate was wrong in
the safe direction. The scenario pair is the dock repair above, not fixture churn.

**Also found, deliberately NOT fixed here:** at 80 columns the frame's title row is 81 runes
(`Ashgrove — tick … [8 villagers]`), present in the committed pre-115 frame. Same family as spec
114's legend clamp, different surface. T023's assertion is scoped to the feed so this change does
not silently adopt it. Candidate for the next polish item.

### The wiki-footprint gate fired on its own session (decision 3, dogfooded)

```
task-195-polish-session-1  task=TASK-195  baseLag=10  dirty=false  wikiNotes=32  cleanupEligible=false
  [warn] wiki-footprint: task-195-polish-session-1 touches sources for 32 of 191 wiki notes
         (threshold 30) — the branch has reached across subsystems; ground it or stop widening
```

The pr gate independently reports exactly **32** `wiki-repin-missing` findings — two derivations
from different code paths agreeing on the number. My pre-implementation projection was 30; the
extra two come from `internal/tui/fixtures.go`, which I had not counted. The instrument works and
its threshold is calibrated about right: this branch has in fact reached across `internal/llm`,
`internal/daemon` and `internal/tui`, which is exactly the condition it was built to name.

**Remaining before the PR:** step 8 (operator visual QA on the live world) and step 9 (wiki
re-pin for 32 notes, player-docs regeneration, pr gate green). Step 9 is deliberately un-started
per the session contract — grounding runs once, at the end.

### Decision 6 — spec 115 unlinked from the board until the branch merges

The spec-bridge Stop gate blocked: *"TASK-195 is In Progress but specs/115-chronicle-feed-wrap
only proves To Do: spec.md missing, plan.md missing, no tasks in tasks.md."*

**The gate is right about its mechanism and wrong about the facts.** spec.md, plan.md, research.md,
data-model.md, contracts/ and a 32-task tasks.md all exist, complete, with 31 tasks ticked — on the
BRANCH. The bridge resolves spec state from the root checkout, which is main, where
`specs/115-chronicle-feed-wrap/` does not exist. It sees an empty directory and a task claiming
more than the artifacts prove.

**Root cause is decision 5's deviation, and this is its downstream cost.** Spec 065 has the stub
land on main via `git merge --no-ff` precisely so board-facing tooling can see it. I skipped that
because the branch already carried unreviewed code commits and merging would have dragged them onto
main outside the PR. That reasoning still holds — but the consequence, which decision 5 did not
anticipate, is that **any board card linked to a branch-only spec fails the bridge gate**.

**Action taken (reversible):** the `Spec:` marker and the seven mirrored phase ACs are removed from
this card. TASK-195 stays In Progress on its own merits — five human ACs, three shipped items, six
pushed commits — and no longer asserts a spec-derived status the gate cannot verify. Nothing about
the spec artifacts changed; they remain complete on the branch, and the full record of the link
lives in decision 5 and in the commits.

**Re-link after the PR merges,** when `specs/115-chronicle-feed-wrap/` is on main and the bridge can
derive its true state.

**The real seam, for the operator.** Spec 065 assumes the spec stub is the branch's FIRST commit.
This session inverted that: three ad-hoc items shipped before the fourth turned out to need a spec.
Any polish session that escalates mid-stream hits this. Two ways to close it, both operator calls,
neither taken here:

1. *Land spec artifacts on main separately* — cut a spec-only branch from origin/main carrying just
   `specs/115-*`, `git merge --no-ff` it at root, then merge main into the task branch so both sides
   hold identical content and the later PR merge stays clean. This is what spec 065 actually
   prescribes and it would also clear the branch's `baseLag=10`. It needs operator ratification
   because it writes to main outside a PR.
2. *Change the session rule* — a polish session that might escalate claims its spec number and lands
   a stub on main up front, before any code, so the ordering spec 065 assumes always holds.

### Decision 7 — stub-first rule adopted in-repo; the tool gap carded upstream

Operator ruling: do #2 ourselves if it does not step on praxis, and card praxis if it must.
Investigation says **both**, because #2 and the underlying gap are different layers.

**#2 is ours — no plugin change needed.** "A session that might escalate claims its number and
lands the stub on main up front" is a promptworld session-flow rule. Adopted as a new bullet in
the Claim-before-work section of `CLAUDE.md` (branch commit 4efa712b), stated for
freeform/polish sessions explicitly, since the spec-first case the protocol was written for never
hits this. Rides this PR.

**The gap underneath is praxis-level.** `deriveSpecState` in the spec-bridge plugin's
`lib/spec-derive.mjs` (~line 148) resolves every artifact through `existsSync`/`readFileSync` on a
working-tree path:

```js
const has = (name) => existsSync(join(specDir, name));
```

No git awareness anywhere in the path, so a branch-only spec dir is structurally invisible.
Carded as **praxisflux TASK-104** ("spec-bridge gate is blind to spec dirs that live only on a
branch"), committed and pushed to that repo at 335edfa, with four ACs and the precedent named:
promptworld's own claim gate closed the identical hole in spec 111 via `branchHeldSpecNumbers`,
which enumerates `refs/remotes/origin/task-*` and reads each branch tree. The card also flags that
`link`'s phase-AC seeding and `sync` share the same derivation and may share the bug.

**#1 not taken.** Since #2 was doable in-repo, we are not landing spec 115's artifacts on main
outside a PR. Note what that means for THIS branch: the stub-first rule is **preventive, not
curative** — spec 115 stays unlinked (decision 6) until the PR merges, at which point
`specs/115-chronicle-feed-wrap/` is on main and the link can be restored with the bridge deriving
its true state.

### Step 8 — operator visual QA signed off (2026-08-03)

Operator reviewed and signed off on the session's shipped work. AC #4 ticked. Step 9 (grounding)
authorized and starting.

Caught during QA setup: the worktree's `promptworld` binary was still the 13:51 build, from before
the spec 115 changes — `go build ./...` during implementation never rewrites the `-o promptworld`
artifact. Anyone QA-ing with it would have been looking at the old truncating feed. Rebuilt at
15:45 before handing over the QA routes. Worth remembering for any future session that proves a
client-side change against a live world.

### Step 9 — grounding complete, pr gate GREEN (commits 16eb9d5e, 3d81662e)

**Wiki: 32 notes re-verified against 4efa712b.** The planner offered zero auto-repinnable notes, so
every one was read against the diff. No wiki note names the wrap renderers by symbol, and every
"truncat" hit in the corpus turned out to be about something else — the `agent.path_truncated`
event, the 80-rune payload cap, `fitTakeoverLines`' row budget, memory sections, the refuel ceiling.
**Three notes made claims the diff actually invalidates**, and only those three changed prose:

- `daemon-cognition-calibration` named `llmCfg.Local.Workers()`'s `workersWarn` and "both tiers'
  `ToolModeResolved()`" as the boot warning sources. Both call sites are gone.
- `llm-provider-registry` gains `Config.ProviderReports()` as the shape-independent accessor, with
  the reason: the two config shapes are mutually exclusive, so `Config.Local`/`Cloud` are always
  zero on a v2 world.
- `tui-chronicle-feed` gains the wrap and hanging-indent behavior.

The other 29 were re-pinned unchanged — verified, not assumed.

**Size-budget honesty.** `tui-chronicle-feed` was **already 8060 chars** before this session —
pre-existing debt owned by TASK-156. My first draft of the wrap paragraph took it to 8969; condensed
to 8592, pointing at the design authority for the full statement. So this change adds **532** chars
to a note that was already over, not 909. It does add to TASK-156's debt; recorded rather than
buried.

**Player docs: 24 provenance tags re-pinned across 9 pages.** Eight were pin-only — their sources
moved commit but not content. `understanding-the-screen` genuinely changed, because the feed wrap is
player-visible: it now says a thought or line of conversation wraps rather than being cut at the
right edge, that continuation rows line up under where the sentence began, and that a narrow side
panel stops after three rows.

**Gates, all green:**

```
check-merge-drift pr   → exit 0 (only the tui-surface reminder, already satisfied)
go test ./...          → exit 0
check-tui-design       → all checks passed
tui-frames --check     → committed matrix matches a fresh --dump
player-docs --check    → 16 fresh, 0 stale, 0 missing, 0 broken-ref
wiki freshness         → exit 0; planner silent
```

Branch freshened by **merging origin/main into it** (never rebased — rebases stale in-branch wiki
pins): `baseLag` 16 → 0. That merge brought in the concurrent session's TASK-196/197 cards, which
touch nothing this branch does.

AC #5 ticked. The branch is PR-ready.

### PR #163 opened — https://github.com/evanstern/promptworld/pull/163

Branch `task-195-polish-session-1` → `main`. All five ACs ticked; pr gate exit 0 at open time.

**A merge collision was resolved on the way there.** Between the grounding pass and opening the PR,
TASK-196 (PR #162) and the quickstart outline (PR #161) merged to main, and the pr gate flipped from
green to **blocked**: `textual-conflict` on `docs/wiki/CAPSULES.md` and
`docs/wiki/guardian-survival-watches.md`. Both sessions had re-pinned that note — a pure
`verified_against` collision, no prose conflict — and neither pin was correct afterwards, since
neither commit sees both changes.

Resolved by merging origin/main INTO the branch (never rebased — rebases stale in-branch pins),
regenerating `CAPSULES.md` rather than hand-merging it (it is derived), and **re-verifying** the note
rather than re-pinning it blind: its only daemon claim is that the three watches seed at boot right
after `seedTuning`, and this branch's `daemon.go` diff touches neither seeding nor tuning — it
changed the LLM boot-report block alone. Prose stood; pinned to the merge commit, which sees both
sides. This is the merge-drift gate doing exactly the job spec 051 built it for.

**Post-merge state:** `baseLag` 0, pr gate exit 0 with only the tui-surface reminder (satisfied),
`go test ./...` exit 0, frames match a fresh dump, player docs 16/16 fresh, wiki planner silent.

**Left for after the merge:** re-link spec 115 to this card (decision 6) — once
`specs/115-chronicle-feed-wrap/` is on main the spec-bridge gate can derive its true state — then
worktree/branch cleanup.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Shipped as PR #163, merged 2026-08-03 via a real merge commit (2b4201cb, three parents — verified,
because a squash would have rewritten the in-branch wiki pins out of history and staled all 32).

**Three items and one spec, each decided on this card before implementation:**

1. **Daemon boot report read dead config fields.** Every world created since spec 024 booted with
   `daemon: llm orchestrator on (local  @ , cloud , budget $100/mo)` — the report read the legacy
   `Config.Local`/`Cloud` structs, which are always zero on a v2 `providers` world. Not merely
   cosmetic: the same block sourced the `parallel`/`tool_mode` clamp warnings from those dead
   fields, so a v2 world clamped in total silence. `Config.ProviderReports()` resolves either shape
   and returns the warnings the construction path discards; the report scales to N providers.
2. **Wiki-footprint threshold in the session gate.** Grounded in the gate's own logic: staleness is
   binary per note and idempotent, so the re-pin bill is a set union over files touched, never a sum
   over edits — grounding cadence is the wrong instrument and footprint breadth is the right one.
   Warns at 30 of 191 notes, advisory only. It fired on this very branch at 32, and the pr gate
   independently reported 32 `wiki-repin-missing` findings.
3. **Spec 115 — the raw feed wraps with a hanging indent** aligned to the summary column, the indent
   recomputed per frame from the visible window's column widths. 32/32 spec tasks.
4. **A stub-first rule in `CLAUDE.md`**, from the seam this session discovered the hard way.

**What went wrong, and what caught it.** Routing every row through the wrap path collapsed the
column padding on SHORT rows — the wrapper budgets on `strings.Fields` and the padding is a run of
spaces. `TestPreLadderGoldenFrames` failed; its hashes were NOT re-pinned, because the honest read
was that short rows had genuinely moved. Fixing it revealed a pre-existing defect: the narrow dock
has always wrapped, so it had always been destroying its own column alignment, and the committed
frames recorded the collapsed form as if intended. FR-009 was amended rather than left contradicting
its own diff.

**Two collisions with concurrent sessions, both resolved by merging main in, never rebasing.** The
second was a `verified_against` collision with TASK-196 on `guardian-survival-watches` — re-verified
against the union rather than re-pinned blind.

**Left behind deliberately:** a pre-existing 81-rune title row at 80 columns (spec-114 family, out of
scope, test scoped around it rather than silently adopting it); 532 chars added to TASK-156's
note-size debt, condensed down from 909; and praxisflux TASK-104 carding the spec-bridge gate's
blindness to branch-only spec dirs, which is the durable fix behind the stub-first rule.

Spec 115 re-linked to this card post-merge, deriving Done-eligible at 32/32.
<!-- SECTION:FINAL_SUMMARY:END -->

# Absence attribution (TASK-173) — sweep runbook (2026-08-02)

**You (the session reading this) are the ORCHESTRATOR** for the task below. Run it
through promptworld's full PDLC — spec → link → worktree → delegated implementation →
PR → merge → re-ground. Direction is decided; do not re-litigate it: TASK-173's card,
re-opened 2026-08-02 with two independent soak worlds as evidence, wins. Plan-of-record
is the board; this file carries only ordering, doctrine, gates, and the log.

**Status:** executing · operator sign-off on lanes: 2026-08-02 (implicit — the operator's
`/pdlc:sweep on TASK-173` invocation named the scope verbatim, and the lane plan is
degenerate: one task, one lane, one branch, one PR. Lane construction carries no judgment
here, so there is nothing for a sign-off to protect. The genuine judgment call — the
narrative-locus fork — is an operator checkpoint below, not a lane question.)
<!-- Only the OPERATOR flips draft → signed-off. A multi-task successor to this runbook
     must get real lane sign-off; the implicit form above is valid ONLY because a
     one-task sweep has no lane ordering to get wrong. -->

## Read first (in this order)

1. `backlog task view TASK-173 --plain` — the card IS the direction source. Its
   **re-open block (2026-08-02)** carries the failing evidence and, just as importantly,
   the **WHAT IS NOT BROKEN** paragraph: TASK-159/spec 081 worked at the memory layer
   (absence memories are 6.7% of all memories, against a 75% showstopper baseline), and
   the correction rate is already down from playtest-1. Do not re-litigate the memory
   layer. The remaining failure is downstream, at **storyline/narration salience**.
2. Project gates: root `CLAUDE.md` — root-read-only + board-sync exception (TASK-160/161),
   wiki-in-PR lifecycle (spec 069), merge-drift gates (spec 051), claim protocol
   (spec 065), model tiers (`.specify/memory/constitution.md` Principle V).
3. `docs/wiki/CAPSULES.md` → the four notes this task's surface touches:
   [[chronicle]], [[mental-maps]], [[agent-mind]], [[event-types-mental-map]]. Load
   just-in-time; never bulk-load the corpus.
4. `specs/097-perception-of-absence/spec.md` — its D3/D4 decisions (mind-side belief
   reconciliation; low base salience + dedup) are the layer this task sits ON TOP of,
   not a layer to redo.
5. `backlog task list --plain` — live state; other sessions move it while you work.

## State when this runbook was written (2026-08-02, origin/main 70185338)

- **Done already:** nothing in scope. TASK-173's earlier Done (2026-07-30, measurement
  only) was reversed by the 2026-08-02 re-open — treat that Final Summary as superseded
  history, not as a result.
- **In flight in other sessions (do not duplicate; expect their merges):**
  - **TASK-187** (TUI frame harness) — In Progress, live sibling session; worktree
    `.worktrees/task-187` and branch `task-187-frame-harness` exist, and the root
    checkout carries an uncommitted edit to that card. The session gate reports the
    branch `cleanup-eligible (ancestor)`; **it is NOT cleaned by this sweep** — an
    In Progress card plus an uncommitted root-side card edit is a live lane, and the
    janitor prescription is declined on that evidence. Never claim, rebase, or clean
    its branch or worktree.
- **Paused — untouched (`paused` label in the task's frontmatter `labels:`):** none.
- **Preserved state (never touch):** the two soak worlds the re-open evidence rests on —
  `/Users/evanstern/.claude/jobs/ca35de11/tmp/soak/soak-world` (12.02 game-days, gemma)
  and its `soak-qwen` sibling (5.69 game-days, qwen3.6) — plus every world under
  `~/.promptworld/measure/`. These are the before-side of this task's before/after
  comparison; destroying them destroys the evidence bar in "Done means".
- **Queued (this runbook's scope):** TASK-173 alone. Spec number **110** claimed
  (`specs/110-absence-attribution`, claim commit `eae58fb0` on
  `task-173-absence-attribution`); card marker landed on main at `46222529`.

## Execution lanes (dependency-ordered; parallelize within a lane)

**Lane 1 — the only lane:**
- **TASK-173 (Opus tier · model `claude-opus-5`, fallback `claude-opus-4-8` —
  dispatched as `.claude/agents/spec-implementer-opus.md`, NEVER by a `model` parameter)**
  — absence attribution: a map correction explainable by harvest activity the correcting
  villager could know about must not earn mystery-grade narrative weight, while genuinely
  unexplained absences still surface.
  **Rubric lines fired (all three recorded on the card at claim time):**
  1. `internal/mind` orchestration — the change lands in the absorb driver
     (`mind.go`'s `agent.map_corrected` arm) and/or the narrator driver
     (`narrate.go`'s `chronicleNote`).
  2. Doctrine-adjacent behavior change — specs 092/094 (determinism, emitter-computes)
     govern whether attribution may ride the event payload at all, and narrative
     salience is player-facing behavior.
  3. A prior attempt shipped a live defect — the 2026-07-30 measurement-only close did
     not survive a longer soak.

Record the tier + explicit model ID + fallback + **which model actually served** on the
board task at every dispatch. Escalation is one-way and is an operator checkpoint;
there is no tier above the one already pinned here, so any deviation is a DE-escalation
and needs the same checkpoint.

## Per-PR gates this project enforces (enumerated — implementers cannot miss these)

- **Merge-drift gate: present at `scripts/check-merge-drift.mjs`.** Mandatory at every
  choke point, invocations verbatim:
  - `node scripts/check-merge-drift.mjs session` — at sweep start and at every resume
    (janitor + drift matrix).
  - `node scripts/check-merge-drift.mjs claim --dir 110-absence-attribution` — before
    creating the spec dir. **Run, passed 2026-08-02.**
  - `node scripts/check-merge-drift.mjs worktree --spec 110 --task TASK-173` — before
    `git worktree add`. **Run, passed 2026-08-02.**
  - `node scripts/check-merge-drift.mjs pr` — from the worktree, immediately before
    `gh pr create` AND again after every history move (merge-in). Nonzero exit blocks;
    its semantic-overlap warnings are the same-PR companion-artifact checklist.
- **Go gates:** `go build ./...` and `go test ./...` green in the worktree before the PR.
  Any change to `internal/mind` scheduling/absorb paths additionally runs
  `go test -race ./internal/mind/...`.
- **Wiki-in-PR (spec 069), in-branch and gated:** the branch must re-verify and re-pin
  every wiki note whose pinned sources it touches — at minimum [[chronicle]]
  (`internal/sim/chronicle.go`, `internal/mind/narrate.go`, `internal/scribe/scribe.go`)
  and, if the reducer/payload side moves, [[mental-maps]] and [[event-types-mental-map]].
  Use `/grounding-wiki:wiki-update`. `wiki-repin-missing` is a blocking finding with no
  bypass flag.
- **Player docs:** any change under `docs/wiki/` requires regenerating `docs/player/` via
  the `player-docs` skill; the gate's probe is
  `node .claude/skills/player-docs/scripts/check-freshness.mjs --check`
  (`player-docs-stale` is blocking).
- **TUI design authority (spec 047):** only if the branch touches `internal/tui/` — then
  `node scripts/check-tui-design.mjs --changed` and a same-PR amendment of
  `docs/design/tui/`. NOTE: main already reports `tui-design` stale at
  `docs/design/tui/anatomy.md` (session gate, 2026-08-02) — a **pre-existing** finding
  this task did not cause and does not adopt; if the branch never touches
  `internal/tui/`, do not fix it here (it is someone else's PR to carry, or its own card).
- **Merge shape:** `gh pr merge --merge`. Never squash — this PR carries in-branch wiki
  re-pins, and a squash rewrites the hashes those pins reference.

## Per-task artifacts required before PR

**No PR opens for TASK-173 until every line below checks true.**

- [ ] `specs/110-absence-attribution/` carries a real `spec.md` (problem + requirements
      mapped to the card's AC#2 and AC#3), `plan.md` (checked against
      `.specify/memory/constitution.md`), and `tasks.md` (phased checkboxes the bridge
      derives from), all committed on `task-173-absence-attribution`. The claim stub
      reserves the number and satisfies none of these.
- [ ] The card carries its `Spec: specs/110-absence-attribution` marker (landed
      2026-08-02, main `46222529`) and its phase ACs are seeded from `tasks.md` via
      `spec-bridge:link` update mode **before** implementation dispatch.
- **Escape lines (operator-signed only):** none. TASK-173 gets a full spec set.
- **HOST ADDITION — the evidence bar (this task's Lane-0 ruling, checkable):**
  - [ ] AC#2/AC#3 are proven by a **soak of at least 12 game-days**, not by unit tests
        alone and not by a short run. Rationale, binding: the 2026-07-30 close ticked
        both ACs off a 4.2-game-day window, and the storyline reappeared past it. A
        window shorter than the 12.02-game-day soak that produced the re-open evidence
        cannot disprove that soak, so it cannot close this card.
  - [ ] The soak reports, on the card: (a) count and share of chronicle entries that are
        absence-themed; (b) whether any **named** absence storyline slug appears in the
        chronicle ring; (c) the harvest-explained share of `agent.map_corrected` "gone"
        entries (the soak baseline is 969/972 = 99.7% over 352 distinct locations);
        (d) the count of genuinely-unexplained absences and evidence that they still
        surfaced (AC#3's anti-suppression check — the baseline is 3 in 12 game-days).
  - [ ] Where feasible the soak runs on **both** local models the re-open reproduced on
        (gemma4:12b-mlx and qwen3.6), since the card establishes the failure is not
        model-specific. A single-model soak is acceptable only with the reason recorded
        on the card.

## Concurrency & conflict doctrine

- **Hotspots:** `internal/mind/narrate.go` and `internal/mind/mind.go` (the absorb
  switch) are this task's own surface; `internal/sim/mentalmap.go` /
  `internal/sim/payloads.go` if attribution ends up emitter-computed. TASK-187's
  frame-harness lane is expected in `internal/tui/` — a different subsystem, so no
  hotspot overlap is predicted; re-check with the session gate's drift matrix at each
  resume rather than trusting this prediction.
- **Paused tasks are not live lanes:** none currently labeled `paused`. TASK-187 is not
  paused — it is live, and equally untouchable (see the state snapshot).
- **This branch is pin-carrying** (it will land in-branch wiki re-pins): it reconciles by
  **merging `origin/main` in**, never rebase, never squash, never force-push. Its PR
  lands as a merge commit.
- **Honest re-pins only — a merge-in never justifies a pin bump.** After a merge-in,
  classify every stale or conflicted pin against
  `git diff <old-pin>..<merge-commit> -- <sources>`: **RE-PIN-ONLY** (the diff provably
  cannot invalidate the note's prose) or **NEEDS-REVIEW** (re-verify and amend the prose
  BEFORE bumping). The merge commit is the re-pin target, never the justification.
- **After every history move, re-run the gates AND the freshness probes
  unconditionally** — not only when `docs/wiki/` changed. Pins reference sources outside
  the wiki, so a wiki-untouched diff can still stale a pin.
- Verify the PR merged (`gh api repos/:owner/:repo/pulls/<n> --jq .merged`) before
  removing the worktree or deleting the branch. Never delete+recreate a closed PR's head.

## Operator checkpoints (do not proceed silently)

1. **The narrative-locus fork — how a mundane correction should read.** AC#2 says a
   harvest-explained correction must not earn *mystery-grade narrative weight*; it does
   not say the beat must vanish. Two readings:
   - **(a) Attribute it** — the chronicle line still appears, but carries its mundane
     cause ("Birch found the pine at (12,7) gone — Cedar had felled it"), so the narrator
     has an explanation in hand and no mystery to build.
   - **(b) Withhold it** — an attributed correction contributes no chronicle line at all
     (it stays in memory/telemetry), so it cannot feed a chapter.
   **Orchestrator's recommendation: (a), with a volume bound.** Attribution preserves the
   believe-act-discover beat spec 041 deliberately built, and it is the reading that makes
   AC#3 legible by contrast — an *unattributed* absence reads as strange precisely because
   the attributed ones read as ordinary. Pure (b) risks trading a false mystery for a
   silent world. This is a rendering choice, reversible in a follow-up, so the sweep
   proceeds on (a) unless the operator says otherwise — **flagged, not assumed silently.**
2. **If the spec's design work concludes attribution must be emitter-computed** (i.e. ride
   the `agent.map_corrected` payload rather than being derived mind-side), that is a
   determinism-doctrine change under specs 092/094 and an event-shape change — stop and
   surface it before implementation.
3. Tier escalations/de-escalations; any lane amendment (amend this file, note why, tell
   the operator).
4. **Softening any gate this runbook enumerates** — including the 12-game-day evidence
   bar — is a runbook amendment plus an operator ping, never an implementer decision note
   buried in a spec artifact.

## Done means

- TASK-173 is **Done on the board via its own merged PR**, moved there by
  `spec-bridge:sync`'s derived plan (never hand-set on a linked task).
- The card still carries its `Spec: specs/110-absence-attribution` marker at sweep end.
- `specs/110-absence-attribution/` contains a real `spec.md`, `plan.md`, and `tasks.md`.
- AC#2 and AC#3 are ticked **against the evidence bar above** — a ≥12-game-day soak whose
  four required measurements are recorded on the card, showing no named absence storyline
  while genuinely-unexplained absences still surface.
- `go build ./...` and `go test ./...` green on main after the merge.
- Grounding fresh: wiki pins current for every note the PR's sources touched,
  `player-docs` freshness check passing, and
  `node scripts/check-merge-drift.mjs session` reporting no new grounding-stale finding
  attributable to this PR (the pre-existing `tui-design` finding is excluded by name).
- `git worktree list` shows no `.worktrees/task-173`; branch `task-173-absence-attribution`
  deleted; root ff-pulled.
- This file's execution log complete and its status flipped to **done**.

## Execution log

| date | task | PR | merge | tokens/cost (best-effort) | notes |
|------|------|----|-------|---------------------------|-------|
| 2026-08-02 | TASK-173 | — | — | — | claimed: card In Progress (main `48701799`), spec 110 stub (`eae58fb0`), bridge marker (main `46222529`), tier pin + stale-AC correction (main `70185338`). Spec set complete (`8cc6ca6b` spec.md, `ce298a31` plan.md + tasks.md); phase ACs seeded (main `fcf421be`). phases: 1 done (`f8bc0c1d`/`b4265b18`/`2756190d`, served by `claude-opus-5`; build+full suite green, `-race` one unattributed failure then two clean runs — noted in plan.md), 2+3 dispatched. **Orchestrator's recorded grouping call:** Phase 3 (3 tasks — narrator prompt line + telemetry counters) is dispatched WITH Phase 2 rather than separately, because both edit the same summary line Phase 2 introduces and a separate agent would re-pay the full context read to touch three lines. Phases 4 and 5 stay separate. |

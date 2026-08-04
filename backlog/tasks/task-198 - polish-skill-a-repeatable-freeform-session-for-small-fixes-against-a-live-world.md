---
id: TASK-198
title: >-
  polish skill: a repeatable freeform session for small fixes against a live
  world
status: In Progress
assignee: []
created_date: '2026-08-03 23:34'
updated_date: '2026-08-04 02:51'
labels: []
dependencies: []
ordinal: 180001
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
A project skill that packages the freeform "polish session" workflow — the loop TASK-195 ran by
hand — so any session can pick it up and get the same rigor without the operator re-explaining it.
Polish work is small FE/TUI changes and gameplay tweaks against decisions already made; the skill's
job is to make that fast without letting it skip the paper trail or the gates.

## Use cases

- As the operator, I want to say "let's do a polish session" and get the same disciplined loop I got
  by hand in TASK-195, without dictating the steps again.
- As a session picking this up cold, I want to know the ordering constraints that bite — especially
  the ones that only bite freeform sessions — before I make the mistake rather than after.
- As a reviewer, I want a polish PR to arrive with the same evidence a spec'd PR does: a decision
  per item recorded before implementation, live proof, and green gates.

## Source material

TASK-195's card is the worked example — its decision log is the narrative of one full run, including
the three things that went wrong. Read it first. Also: spec 115 (the one item that escalated),
CLAUDE.md's "Stub-first, always" bullet, and PR #163.

## The loop the skill encodes (operator-authored, refined by the TASK-195 run)

1. Worktree + branch; one long-running TASK card for the whole session.
2. Keep a runnable world live and iterate against it.
3. Sub-loop per item: discuss → optionally ground → **record the decision on the card, with a
   file:line diagnosis, BEFORE implementing**.
4. Ad-hoc after each loop: does the accumulated work warrant a spec? If yes, spec it and execute; if
   no, continue.
5. Operator visual QA.
6. Grounding ONCE at the end, in-branch, then the PR.

## What the TASK-195 run proved the skill must carry

- **Stub-first.** A polish session is exactly the kind that discovers mid-stream that an item needs a
  spec, and by then the branch carries unreviewed code and the stub can no longer land on main
  first. The skill must tell a session to claim its spec number and land a stub up front if
  escalation is even plausible. See CLAUDE.md and praxisflux TASK-104 for the underlying tool gap.
- **Grounding batches; it does not accumulate.** Wiki staleness is idempotent per note, so deferring
  costs nothing and per-item grounding costs N times over. What grows is footprint breadth — which
  is why the session gate now reports `wikiNotes=N` and warns at 30. The skill should teach reading
  that number, not a grounding cadence.
- **The trivial-exemption bar is what makes ad-hoc legal.** Surgical fix + complete file:line
  diagnosis + ACs on the card. An item that cannot be pinned that precisely is the signal to
  escalate, not to guess.
- **Verify the binary before live QA.** `go build ./...` never rewrites the CLI artifact.
- **Golden/byte-identity test failures are signal, not chores.** The TASK-195 run had one, and
  re-pinning its hashes would have shipped a real defect. The skill must say: read the diff, decide
  whether the change is sanctioned, and only then re-pin.

## Open questions for the authoring session

- Does the skill itself warrant a spec? It encodes policy others follow, which argues yes; it is one
  self-contained artifact, which argues no. Decide at the start — and if the answer might be yes,
  claim the spec number stub-first before any content lands.
- Where the flow's canonical statement lives: the skill alone, or a short CLAUDE.md pointer to it.
- Whether the skill should scaffold the session card from a template, or leave authoring to the
  session.

Spec: specs/117-polish-session-skill
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A polish session can be run end to end from the skill alone, without the operator restating the loop
- [ ] #2 The skill states the stub-first ordering constraint and why freeform sessions specifically hit it
- [ ] #3 The skill states the trivial-exemption bar as the test for ad-hoc vs spec, and requires the file:line diagnosis on the card before implementation
- [ ] #4 The skill routes grounding to the end of the session and teaches reading the session gate's wikiNotes footprint rather than a cadence
- [ ] #5 The skill names the live-QA binary trap and the golden-test-failure-is-signal rule
- [ ] #6 TASK-195 is referenced as the worked example rather than its lessons being re-narrated in full
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Open questions, resolved at the start (per the card's instruction)

1. **Does the skill warrant a spec? YES — spec 117 claimed and stubbed on main before any content.**
   Resolved from artifacts, not preference: the constitution's trivial exemption requires a *surgical
   fix* with a complete file:line diagnosis, and authoring a new artifact is neither. Precedent is
   exact — spec 113 specified the `tui-frames` project skill, also a documentation-only deliverable.
   Claimed stub-first *because* that is the rule this skill exists to teach: dogfooding, not ceremony.
2. **Where the canonical statement lives: the skill, with a short CLAUDE.md pointer.** Repo convention
   already answers it — the TUI-design and wiki-in-PR blocks in CLAUDE.md are short pointers and the
   procedure lives in the skill.
3. **Scaffold the session card: yes, from a template shipped with the skill.** AC #1 requires a cold
   session to run the loop without the operator restating it; a card template is the artifact that
   makes step 1 self-serve.

## Shape

```
.claude/skills/polish-session/
├── SKILL.md                       the loop + the four traps + the ad-hoc-vs-spec test
├── templates/session-card.md      scaffold for the session's single long-running card
└── scripts/session-status.mjs     advisory probe: binary freshness + wiki footprint
```

The script is what keeps two ACs from being mere prose. AC #5's binary trap
(`go build ./...` never rewrites the `-o promptworld` artifact) is purely mechanical — compare the
binary's mtime against the newest tracked Go source — and AC #4's footprint number already exists in
the session gate's `--json`. Constitution Principle III: gates over assertions. Advisory, never
blocking; the session gate's non-blocking contract governs.

## Tier

Authoring is doctrine text plus one small Node probe — the doctrine is authored at the planning tier
(it is policy, not implementation) and the probe is a single self-contained script with no
cross-package or concurrency surface. Recorded on delivery per Principle V.

## Steps

1. Claim: card In Progress, spec 117 stub landed on main, worktree cut. (done)
2. Author spec 117 in full: spec.md, plan.md, tasks.md; link via spec-bridge.
3. Author the skill: SKILL.md, card template, status probe.
4. Verify: probe exercised on this very branch (positive + negative control), skill discoverable.
5. Grounding once at the end, in-branch; then the PR.
<!-- SECTION:PLAN:END -->

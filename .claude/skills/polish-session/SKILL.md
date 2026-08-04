---
name: "polish-session"
description: "Runs a freeform polish session — small FE/TUI changes and gameplay tweaks worked against a live world, on one long-running card and one PR, with a decision recorded before each fix. Covers the loop, the ad-hoc-versus-spec test, the stub-first ordering constraint that only freeform sessions hit, where grounding goes, and two failures that look like chores. Worked example: TASK-195 / PR #163. Start here: node .claude/skills/polish-session/scripts/session-status.mjs --check"
metadata:
  author: "promptworld"
user-invocable: true
disable-model-invocation: false
---

# polish-session

A polish session is a fast loop for **small FE/TUI changes and gameplay tweaks against
decisions already made**: one long-running card, one branch, one PR, N items. It trades the
per-item spec cycle for speed and pays for that with a decision — including a file:line
diagnosis — recorded on the card *before* each implementation.

**Scope guard.** Anything needing real design thought leaves this session and goes through the
full PDLC cycle on its own card. The flow is polish, not discovery.

## Before the first line of content: claim a spec number

**If it is even plausible that an item in this session will need a spec, claim the number and
land the stub on `main` now — before the branch's first content commit.**

Why it cannot wait. `spec-bridge` derives a linked card's provable status by reading
`specs/NNN-*` off the **filesystem at the repository root**, with no git awareness at all. A
spec directory that exists only on a branch is structurally invisible to it, so the moment such
a spec is linked, the gate reports the card as exceeding artifacts that are in fact complete.
The remedies are unlinking the card or landing the spec on main.

Why **freeform sessions specifically** hit this, and spec-first sessions never do. The claim
protocol assumes the spec stub is the branch's first commit — true by construction when you
know the spec up front. A polish session is *defined* by discovering mid-stream that an item
needs one. By then the branch carries unreviewed code, and landing the stub means merging that
code onto main outside the PR, which the PR boundary forbids. **The cheap fix is only available
before the first content commit.** Retrofitting is not a fix; it is a choice between two
things you are not allowed to do.

The claim is two pushes, both up front:

```
node scripts/check-merge-drift.mjs claim --dir NNN-<slug>          # gate: is the number free?
# card → In Progress, committed at root scoped to backlog/, pushed        (the mutex event)
# branch's first commit creates specs/NNN-<slug>/spec.md as a stub, pushed
git merge --no-ff task-<N>-<slug>                                  # at root; lands the stub
```

A claimed-but-unused number costs nothing. See CLAUDE.md's "Stub-first, always" bullet for the
protocol and praxisflux TASK-104 for the tool gap underneath it.

## The loop

1. **Cut the worktree and the branch.** `node scripts/check-merge-drift.mjs worktree --spec NNN
   --task TASK-<n>`, then `git worktree add .worktrees/task-<N> -b task-<N>-<slug> origin/main`.
2. **Open one long-running card** for the whole session — no per-item cards, no per-item PRs.
   Scaffold it from `templates/session-card.md` beside this file, then claim it: card to
   `In Progress`, committed **at root** scoped to `backlog/` alone, pushed immediately. That
   push is the mutual-exclusion event; a rejection means another session holds the lane, so
   stop and surface it rather than pulling and carrying on.
3. **Get a world running and keep it running.** Cycle it onto the branch build as changes land;
   nothing else runs against it. Keep the restart helper *outside* the repo so it never enters
   the PR diff.
4. **Per item, repeat:**
   1. Discuss the bug, tweak, or nit.
   2. Optionally ground it — wiki reads, frame dumps, research.
   3. **Record the decision on the card, with a complete file:line diagnosis, BEFORE writing
      code.** Not after, not "while implementing." This is the whole price of skipping the spec
      cycle, and a session that pays it late has not paid it.
   4. Implement, then prove it on the live world and record the proof on the card.
5. **After each item, apply the ad-hoc-versus-spec test** (below). If it escalates, write the
   spec, link it with `spec-bridge:link`, and execute it in this session.
6. **Decide whether to go again.** A fresh session is usually the right move once the branch has
   reached across a second subsystem — see the footprint reading below.
7. **Operator visual QA on the live world.** Rebuild the binary first — see below, and
   `node .claude/skills/polish-session/scripts/session-status.mjs --check` before handing over.
8. **Ground once, in-branch:** `/grounding-wiki:wiki-update` to re-verify and re-pin every note
   whose sources this branch touched, then `/player-docs` if `docs/wiki/` changed, then
   `node scripts/check-tui-design.mjs --changed` if `internal/tui/` changed. Finish with
   `node scripts/check-merge-drift.mjs pr` at exit 0 — it blocks on `wiki-repin-missing` and
   `player-docs-stale`, and there is no bypass flag.
9. **Open the PR.** Merge with `--merge`, never squash — a squash rewrites in-branch commit
   hashes out of main's history and stales every wiki pin the PR carried.

## Ad-hoc or spec? The trivial-exemption bar

An item may be worked ad-hoc only when **all** of these hold:

- it is a **surgical fix**,
- with a **complete file:line diagnosis** pinned on the card, and
- with **acceptance criteria** on the card.

That is the constitution's trivial exemption verbatim, and it is what makes the ad-hoc path
legal rather than merely fast.

Read its contrapositive as the escalation signal: **an item you cannot pin that precisely is
telling you to escalate, not to guess.** "I think it's somewhere in the render path" is not a
diagnosis. Neither is a fix whose blast radius you would have to discover by making it.

Borderline calls get made explicitly on the card, with the reasoning — including new behavior
that is not a diagnosed defect, which normally argues for a spec even when it is small.

## Grounding: once at the end — and read the footprint, not the clock

**Deferring grounding to step 8 costs nothing.** A wiki note is stale iff any file in its
`sources:` changed between its pin and the tip — **binary per note, and idempotent**. Touching
`internal/daemon/daemon.go` once stales the notes that pin it; touching it fifty more times
stales exactly the same notes. The re-pin bill is a set union over *files touched*, never a sum
over edits, so grounding per item just re-verifies the same notes N times.

What does grow is **breadth** — the union widens only when the branch touches a new distinct
file — and this repo's note-per-source concentration makes that sharp: a single stray edit to a
hot file can cost more than an entire narrow session.

So the instrument is footprint, not cadence. **Do not adopt a grounding cadence.** Read the
number:

```
node scripts/check-merge-drift.mjs session
```

Every branch line carries `wikiNotes=N` — the count of notes whose `sources:` intersect that
branch's committed changes. It is visible below threshold precisely so you can watch it move. At
or above the threshold the gate raises `wiki-footprint`, advisory only:

> `[warn] wiki-footprint: <branch> touches sources for N of <total> wiki notes (threshold T) —
> the branch has reached across subsystems; ground it or stop widening its scope`

Read a rising number as a **scope-sprawl** signal, not a grounding-overdue one. The two
sanctioned responses are the two the message names: ground and land what you have, or stop
widening. `session-status.mjs --check` reports the same count with its remaining headroom —
both read from the gate, so there is only ever one derivation to trust.

One caveat the count carries: it is computed from committed work (`mergeBase..tip`), so
uncommitted edits are not in it.

## Two failures that look like chores

**1. Verify the binary before live QA.** `go build ./...` populates the build cache and
**never rewrites the `-o promptworld` artifact**. Every build in your transcript can succeed
while the binary the operator is about to QA predates your changes by hours. Rebuild before
handing over:

```
go build -o promptworld ./cmd/promptworld
```

`session-status.mjs --check` compares the binary's mtime against the newest tracked Go source
and says so when it is behind.

**2. A golden or byte-identity test failure is signal, not a chore.** When
`TestPreLadderGoldenFrames` — or any hash-pinned fixture — fails, **read the diff and decide
whether the change is sanctioned before you re-pin anything.** Re-pinning first can ship a real
defect under a green suite. It can also bury a pre-existing one: a committed fixture records
whatever the renderer did when it was captured, including behavior nobody intended, so a diff
you did not expect may be the test finally showing you an old bug rather than reporting a new
one. Re-pin only after the diff is explained, and say in the decision log what the change was
and why it is correct.

## The worked example

**TASK-195** ran this loop end to end by hand: three ad-hoc items, one escalation to spec 115,
grounding once at the end, shipped as **PR #163**. Every rule above that reads as a warning —
the stub-first ordering, the stale QA binary, the golden-frame failure — is there because that
run hit it. Its decision log is the full narrative, and is the place to read when you want the
story rather than the rule.

```
backlog task view TASK-195 --plain
```

## Companion skills

- `tui-frames` — for any item touching `internal/tui`. Frames are how you see a screen; the
  frame diff is the review artifact.
- `player-docs` — step 8, whenever the branch changes `docs/wiki/`.

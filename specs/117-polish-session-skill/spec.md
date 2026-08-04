# Spec 117 — polish-session project skill

**Board task:** TASK-198 · **Feature branch:** `task-198-polish-skill` · **Created:** 2026-08-03

## Problem

TASK-195 ran a full polish session — small FE/TUI changes and gameplay tweaks against a live
world, on one long-running card and one PR — and the loop worked. It shipped three ad-hoc items,
escalated a fourth to spec 115, grounded once at the end, and merged as PR #163.

**The loop exists nowhere but in that card's decision log.** It was authored by the operator turn
by turn, and every future polish session either gets it re-dictated or improvises. Improvising is
the expensive case: the run proved the flow has ordering constraints that are invisible until they
bite, and by the time they bite the branch is in a state where the cheap fix is gone.

The run hit three of them. Each was recoverable, and each cost real work:

1. **Escalation inverted the claim protocol.** Three ad-hoc items shipped before item 4 turned out
   to need a spec. Spec 065 assumes the spec stub is the branch's *first* commit; by then the
   branch carried unreviewed code, so landing the stub on main would have dragged that code onto
   main outside the PR. `spec-bridge`'s gate reads `specs/NNN-*` off the filesystem with no git
   awareness, so a branch-only spec dir is invisible to it and the linked card is reported as
   exceeding artifacts that are in fact complete. Cost: spec 115 unlinked from the board until the
   PR merged, then re-linked. **Freeform sessions hit this and spec-first sessions never do** — the
   protocol was written for a flow where the spec is known up front.
2. **The live-QA binary was stale.** `go build ./...` does not rewrite the `-o promptworld`
   artifact, so the worktree binary was four hours old at QA handover — the operator would have
   been reviewing the pre-change feed. Caught during QA setup, by luck rather than by a check.
3. **A golden-frame test failed and the tempting fix was wrong.** `TestPreLadderGoldenFrames`
   broke because the change collapsed column padding on short rows. Re-pinning its hashes would
   have shipped a real defect; reading the diff instead surfaced both that defect *and* a
   pre-existing one the committed frames had recorded as if intended.

A fourth thing the run had to reason out from scratch, and which every session would otherwise
re-derive: **when to ground.** Wiki staleness is binary per note and idempotent — a note is stale
iff any of its `sources:` changed between its pin and the tip
(`scripts/check-merge-drift.mjs`) — so the re-pin bill is a set union over *files touched*, never
a sum over edits. Deferring grounding to the end of the session costs nothing; grounding per item
re-verifies the same notes N times. What actually grows is footprint *breadth*, which is why that
run shipped the `wikiNotes=N` readout and the `wiki-footprint` warning in the session gate.

## Goal

Make "let's do a polish session" a complete instruction. A session picking it up cold should run
the same disciplined loop TASK-195 ran by hand, meet the ordering constraints *before* it makes
the mistake, and produce a PR carrying the same evidence a spec'd PR does.

## Users

- **The operator starting a polish session** — wants to say "let's do a polish session" and get
  the disciplined loop without dictating the steps again.
- **A session picking the flow up cold** — wants the constraints that bite freeform work
  specifically, stated before the branch is in a state where recovery is expensive.
- **A reviewer receiving a polish PR** — wants a decision per item recorded before implementation,
  a file:line diagnosis behind each, live proof, and green gates.

## Functional requirements

### FR-001 — A project skill at `.claude/skills/polish-session/`

`SKILL.md` with frontmatter matching this repo's project-skill shape
(`.claude/skills/player-docs/SKILL.md`, `.claude/skills/tui-frames/SKILL.md`): `name`,
`description`, `metadata.author`, `user-invocable: true`, `disable-model-invocation: false`.
Operator-invocable by name; model-invocable when a polish session is asked for. The `description`
must name the status probe so an agent sees it without opening the file.

### FR-002 — The loop, runnable end to end

The nine steps of the flow as refined by the TASK-195 run, in order, each stated as an
instruction rather than a description: worktree + branch and one long-running card; a runnable
world kept live and cycled onto the branch build; the per-item sub-loop
(discuss → optionally ground → **record the decision with a file:line diagnosis on the card
BEFORE implementing**); the ad-hoc spec check after each loop; execute any spec created; decide
whether to go again; operator visual QA; grounding once at the end, in-branch; the PR.

A session must be able to run the whole thing from the skill alone. (AC #1)

### FR-003 — Stub-first, and why freeform sessions specifically hit it

State the ordering constraint: **a session that might escalate claims its spec number and lands
the stub on main before its first content commit.** State the mechanism that makes it necessary —
`spec-bridge` derives a linked card's provable status from the filesystem at the repository root,
so a spec dir that exists only on a branch is structurally invisible — and state *why freeform
work is the exposed case*: the spec-first flow the protocol was written for never discovers
mid-stream that it needs a spec, and a polish session is defined by doing exactly that. Name the
consequence of getting it wrong: retrofitting means merging a branch that already carries
unreviewed code, which the PR boundary forbids. (AC #2)

Point at CLAUDE.md's "Stub-first, always" bullet and praxisflux TASK-104 for the underlying tool
gap rather than restating either.

### FR-004 — The trivial-exemption bar as the ad-hoc-versus-spec test

State the bar exactly as the constitution has it — **surgical fix + complete file:line diagnosis
pinned on the card + ACs on the card** — as the test applied after each sub-loop, and state its
contrapositive as the escalation signal: **an item that cannot be pinned that precisely is the
signal to escalate, not to guess.** Require the diagnosis on the card before implementation, which
is what makes the ad-hoc path legal rather than merely fast. (AC #3)

### FR-005 — Grounding routed to the end, taught as footprint rather than cadence

Route grounding to step 8, in-branch, once. Justify it from the gate's own logic — staleness is
binary per note and idempotent, so the bill is a set union over files touched — so that a session
understands *why* deferring is free rather than taking it on faith.

Teach reading `wikiNotes=N` on each branch line of the session gate report, and the
`wiki-footprint` warning at 30 of ~191 notes, as the instrument: the number measures how far the
branch has reached across subsystems, and a rising number is a scope-sprawl signal, not a
grounding-overdue signal. Explicitly do **not** prescribe a grounding cadence. (AC #4)

### FR-006 — The two traps that look like chores

- **Verify the binary before live QA.** `go build ./...` never rewrites the `-o promptworld`
  artifact; a client-side change proved against a live world must be proved against a binary
  rebuilt after the change.
- **Golden / byte-identity test failures are signal, not chores.** Read the diff, decide whether
  the change is sanctioned, and only then re-pin. Re-pinning first can ship a real defect and can
  also bury a pre-existing one. (AC #5)

### FR-007 — TASK-195 as the worked example, cited not re-narrated

The skill names TASK-195 (and PR #163, spec 115) as the worked example and points a session at its
decision log for the full narrative. Each trap is stated as a rule with only enough of its origin
to be credible — the skill must not re-tell the run. (AC #6)

### FR-008 — A session card template

`templates/session-card.md`: the card scaffold a session copies into its long-running card —
gist, "As a …" use cases per spec 087's card format, the session contract, the decision-log
section, and the acceptance criteria that make the flow checkable (decision-before-implementation,
escalation, one branch/one PR, operator QA, grounding-at-end). Derived from TASK-195's card, which
is the only card known to have carried a full run.

### FR-009 — A status probe

`scripts/session-status.mjs`, in the shape of
`.claude/skills/player-docs/scripts/check-freshness.mjs`. Reports the two numbers a polish session
must not guess at:

- **Binary freshness** — whether the worktree's `promptworld` binary predates the newest tracked
  Go source, which is FR-006's trap made mechanical.
- **Wiki footprint** — `wikiNotes` for the current branch and its headroom to the
  `wiki-footprint` threshold, read from the session gate rather than recomputed.

Advisory, never blocking: the session gate's non-blocking contract governs, and making any of this
block is a separate decision. Must not mutate the working tree.

### FR-010 — A CLAUDE.md pointer

A short block in `CLAUDE.md` routing to the skill, in the shape of the existing TUI-design and
player-docs pointers. The skill is the canonical statement of the flow; CLAUDE.md carries a
pointer, never a second copy.

## Out of scope

- Any change to `scripts/check-merge-drift.mjs`, the `wiki-footprint` threshold, or any gate's
  severity. The skill teaches reading the existing instrument; changing it is a separate spec.
- Making the status probe blocking, or wiring it into a hook.
- Closing the `spec-bridge` branch-blindness gap itself — that is praxisflux TASK-104, upstream.
- Any gameplay, TUI, or engine change. This ships a flow, not a fix.
- A skill for spec-first sessions. This is the freeform path specifically.

## Success criteria

- **SC-001** A session handed only the skill can execute the loop end to end without the operator
  restating any step.
- **SC-002** The stub-first constraint is stated before the loop reaches its first content commit,
  with the freeform-specific reason.
- **SC-003** The escalation test is the trivial-exemption bar, verbatim from the constitution, and
  the file:line diagnosis is required on the card before implementation.
- **SC-004** Grounding appears exactly once in the loop, at the end, and the skill teaches
  `wikiNotes=N` rather than any cadence.
- **SC-005** Both traps in FR-006 are stated as rules.
- **SC-006** TASK-195 is cited as the worked example; the skill does not re-narrate the run.
- **SC-007** The status probe exits 0 on a fresh binary and nonzero on a stale one, and leaves the
  working tree clean.

## Discovered during implementation

**FR-009's headroom reading needed a different source than planned.** The plan had the probe
recover the `wiki-footprint` threshold by parsing the gate's finding message, so as never to
carry a second copy of `30`. That message only exists once the threshold has been crossed —
which is precisely when headroom has stopped being useful information. Below threshold, the
intended reading was unavailable and the probe could only print a bare count, leaving the skill's
claim that it "reports the same number with its headroom" untrue.

Resolved by reading the declared `WIKI_FOOTPRINT_THRESHOLD` constant out of
`scripts/check-merge-drift.mjs` itself. That is still the gate as the single authority — a
declared constant read at run time rather than a literal duplicated into this script — and it
degrades to a bare count if the constant is ever renamed, which is why `headroom` is optional in
the probe's output rather than assumed present.

## Acceptance criteria

Mirrors TASK-198's six criteria; see the board card.

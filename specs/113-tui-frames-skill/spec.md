# Spec 113 — tui-frames project skill

**Board task:** TASK-190 · **Status:** planning · **Created:** 2026-08-02

## Problem

TASK-187 (spec 112, merged in PR 157) shipped the frame harness: three fixture worlds,
`promptworld frames`, and 132 generated frames under `docs/design/tui/frames/`. The
capability exists. The **routing to it does not.**

Nothing tells an agent the frames are there, so a fresh session handed a UI task still
reasons about the terminal from the lipgloss calls that produce it — the exact habit the
harness was built to end. The knowledge currently lives in whatever the operator remembers
to paste into a new session, which rots and is easy to forget under time pressure.

The repo separately has three traps a fresh session reliably hits. Each costs a blocked
command and a misleading error before the cause is clear, and none is discoverable at the
moment it bites:

1. The root checkout is read-only; edits are rejected outside `.worktrees/`.
2. `root-guard-hook.mjs` classifies commits by parsing the **raw command string**, so a
   commit message containing a newline or `)` — the standard trailer has both — is read as
   pathspecs and refused with a message that blames scoping. Trailing `2>&1` or `| tail`
   does the same.
3. A squash merge rewrites in-branch pins out of history and stales every one.

## Goal

Make the harness the default way an agent sees this UI, without the operator having to say
so, and make the three traps arrive before they bite rather than after.

## Users

- **The operator starting a UI session** — wants to say "let's fix the chronicle pane" and
  have the agent already know how to look at it.
- **An agent handed a UI task** — wants to be pointed at the rendered characters instead of
  inferring layout from renderer code.
- **A reviewer** — wants UI changes to arrive as a before/after frame diff.

## Functional requirements

### FR-001 — A project skill at `.claude/skills/tui-frames/`
`SKILL.md` with frontmatter matching this repo's project-skill shape (see
`.claude/skills/player-docs/SKILL.md`): `name`, `description`, `metadata.author`,
`user-invocable: true`, `disable-model-invocation: false`. Operator-invocable by name and
model-invocable when a UI task appears.

### FR-002 — How to see a screen
The skill must state, unambiguously, that a screen is seen by **reading**
`docs/design/tui/frames/<fixture>__<state>__<WxH>.txt`, and that generated `.txt` frames
are **never hand-edited** — an edit there is not a design change but a false claim about
what the client renders, silently erased by the next `--dump`.

### FR-003 — The full surface
Off-matrix sizes rendered live via
`go run ./cmd/promptworld frames --fixture <f> --state <s> --size WxH`; the four committed
sizes and the boundary each straddles (`80x30` narrow fallback, `112x30` at the widescreen
breakpoint with an odd column remainder, `113x30` one wider for the even 50/50 split,
`160x50` roomy); `--ansi` for colour; `--list`; `--interactive` for feel rather than layout.

### FR-004 — The design loop
Target frame authored first as ASCII so the operator can eyeball it before code exists;
implement until `--dump` matches; **the frame diff is the review artifact.**

### FR-005 — Authority versus evidence
`docs/design/tui/pages|panels` are the design **authority** — what a surface is supposed to
be. The frames are **evidence** — what it currently is. When they disagree that is a
finding to surface, never silently resolved in either direction.

### FR-006 — The narrow-fallback caveat
Carry spec 112 FR-008: below the widescreen breakpoint the fallback has no fold arithmetic,
so those frames are content-height and legitimately shorter than the requested height. An
agent must not "fix" correct renderer behavior.

### FR-007 — The three repo traps
Root read-only plus the worktree recipe; the bare `git commit -F <file>` requirement; and
merge-commit-only.

### FR-008 — A guard script
Under the skill's `scripts/`, in the shape of
`.claude/skills/player-docs/scripts/check-freshness.mjs`. Regenerates the matrix into a
temp dir, compares against the committed frames, and exits nonzero on any difference —
catching both a hand-edited frame and a stale matrix. Must leave the working tree clean.

## Out of scope

- Any change to the harness, the fixtures, or renderer behavior.
- Wiring frames into `scripts/check-tui-design.mjs` as a regenerate-and-diff gate. Still its
  own decision; this skill's guard is advisory and agent-invoked, not a PR gate.
- Any actual UI/UX change. This ships routing, not a redesign.

## Acceptance criteria

Mirrors TASK-190's nine criteria; see the board card.

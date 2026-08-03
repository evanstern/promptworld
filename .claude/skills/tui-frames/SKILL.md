---
name: "tui-frames"
description: "Routes UI work in internal/tui to the generated frame harness (docs/design/tui/frames/) instead of reasoning from lipgloss calls. Covers how to see a screen, the design loop, and three repo traps a fresh session hits. Check first: node .claude/skills/tui-frames/scripts/check-frames.mjs --check"
metadata:
  author: "promptworld"
user-invocable: true
disable-model-invocation: false
---

# tui-frames

`internal/tui` renders through lipgloss and Bubble Tea. Neither is where you learn what a
screen looks like.

## How to see a screen

**Read the frame file.** Never infer layout from the lipgloss calls that produce it.

```
docs/design/tui/frames/<fixture>__<state>__<WxH>.txt
```

Each file is the exact string `View()` hands Bubble Tea at that size — the characters that
reach a real terminal, nothing added, nothing interpreted. `docs/design/tui/frames/README.md`
documents the harness itself; this skill only routes you to it.

**Generated `.txt` frames are never hand-edited.** An edit there is not a design change — it
is a false claim about what the client renders, and the next `--dump` silently erases it. To
change a frame, change `internal/tui` and regenerate.

## The surface

`go run ./cmd/promptworld frames --list` is the live source of truth for fixtures, states, and
sizes — read it rather than trusting an enumeration here to stay current.

Off-matrix sizes render live and don't need to exist in the committed matrix:

```
go run ./cmd/promptworld frames --fixture <f> --state <s> --size WxH
```

The four **committed** sizes each earn their place by straddling a real layout boundary
(`--list` names them; it doesn't explain why they were chosen):

| size | why it's in the matrix |
| --- | --- |
| `80x30` | below the widescreen breakpoint — the narrow single-pane fallback |
| `112x30` | exactly at the breakpoint, odd column remainder (the map takes the leftover column) |
| `113x30` | one column wider than the breakpoint — the even 50/50 map/dock split |
| `160x50` | roomy enough that no chrome row folds — everything visible at once |

Other flags:

- `--ansi` keeps the color escapes; default is plain text so frames diff cleanly. Use `--ansi`
  only for eyeballing, never for a committed frame.
- `--interactive` opens the real, keyboard-drivable TUI on a fixture. Use it to judge *feel*
  (latency, whether pause responds) — not layout; layout comes from the `.txt` files.

## The design loop

1. Author the target frame as ASCII first, so the operator can eyeball it before any code
   exists.
2. Implement in `internal/tui` until `promptworld frames --dump` matches the target.
3. The frame diff *is* the review artifact — a UI change should arrive as a before/after of
   real frames, not a prose description of one.

## Authority vs. evidence

`docs/design/tui/pages/` and `docs/design/tui/panels/` are the design **authority** — what a
surface is supposed to be. The frames are **evidence** — what it currently is. When the two
disagree, that's a finding to surface, never silently resolved in either direction — don't
"fix" the frame to match the doc or the doc to match the frame without saying so.
`docs/design/tui/pages/home.md`'s "Reconciliation correction" section is what it looks like
when nobody notices for a while.

## Narrow-fallback caveat (spec 112 FR-008)

Below the widescreen breakpoint, the renderer has no fold arithmetic at all, so an `80x30`
frame legitimately has fewer than 30 lines. That's correct, content-height output — not a bug.
Don't "fix" it to be uniform with the wider sizes.

## Three repo traps

**1. The root checkout is read-only.** Edits outside a worktree are rejected by a PreToolUse
hook. Work in a worktree:

```
git worktree add .worktrees/task-<N> -b task-<N>-<slug> origin/main
```

**2. Commit with a bare `git commit -F <file>`, message file outside the repo.** The
root-guard hook classifies commits by parsing the *raw command string*. A commit message
containing a newline or `)` — the standard `Co-Authored-By:` trailer has both — gets
misread as pathspecs and blocked with a misleading "scoping" error; trailing `2>&1`, `| tail`,
or `&&` trip the same misparse. Run `git add` and `git commit` as separate calls, and keep the
commit call bare: no redirects, no pipes, no chaining.

**3. Merge-commit-only.** `gh pr merge --merge` (or a root `git merge --no-ff`) — never squash,
never rebase. A squash merge rewrites in-branch commit hashes out of main's history, which
stales every wiki pin (and any other commit-hash reference) the branch carried.

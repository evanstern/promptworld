# Quickstart: Validating Chronicle Raw Feed Wrapping

**Phase 1** · spec 115

How to prove this feature works, in the order a reviewer should check it. Every step is runnable
from the worktree root.

## Prerequisites

- Go toolchain (`go build ./...` succeeds on the branch)
- No running world required — every check below is headless

## 1. The frame matrix is the primary evidence

The committed frames are what the client actually renders. After implementation:

```sh
go run ./cmd/promptworld frames --dump
git diff --stat docs/design/tui/frames/
```

**Expected:** the `mid-game` frames change. Two of them must now contain a visibly wrapped,
indented row — the long thought and the long conversation turn added to the fixture.

Read one directly rather than trusting the diff summary:

```sh
grep -A3 'thought' docs/design/tui/frames/mid-game__solo__160x50.txt
```

**Expected shape** — continuation lines starting at the summary column, left rail blank:

```
 394500 19:45  agent.thought             Hazel thought: "…first line…
                                         …continuation…"  (planner)
```

## 2. Alignment is exact, not approximate

For any wrapped row in any committed frame, the column at which each continuation line's first
non-space character sits must equal the column at which the first line's summary begins.

```sh
node scripts/check-tui-design.mjs --changed
```

**Expected:** exit 0 once `docs/design/tui/panels/chronicle.md` has been amended in the same
change (FR-013). A non-zero exit here means the design authority still describes the pre-115
behavior.

## 3. Short rows did not move

The overwhelming majority of feed rows are short mechanical events. They must be untouched.

```sh
go test ./internal/tui/ -run 'Chronicle|Grammar|Digest'
```

**Expected:** the pre-existing grammar, digest, and plain-equivalence tests pass **unmodified**.
Any of them needing an edit means a short row's rendering changed, which violates FR-009 — treat
that as a defect in the change, not a stale test.

## 4. The narrow fallback still yields

```sh
go run ./cmd/promptworld frames --fixture mid-game --state solo --size 46x30
```

**Expected:** long rows still wrap, but with **no** indent — at that width the residual text
column would fall below the 24-column minimum, so the indent drops to zero rather than shrinking.

## 5. The row budget still holds

```sh
go run ./cmd/promptworld frames --dump
awk 'END{print NR}' docs/design/tui/frames/mid-game__solo__160x50.txt
```

**Expected:** 50 lines at `160x50`, and the newest event still visible at the bottom of the feed.
An orphaned continuation line at the *top* of the feed is correct — see contract §7.

## 6. Full suite

```sh
go test ./...
node --test scripts/check-merge-drift.test.mjs
```

**Expected:** green.

## 7. Live confirmation (optional, operator)

The frames prove layout. To judge feel:

```sh
go run ./cmd/promptworld frames --fixture mid-game --state solo --interactive
```

Or watch a real world's feed where thoughts and conversations occur naturally.

---

## What "done" looks like

- A committed frame shows a wrapped, indented row (SC-007)
- No committed frame has a line exceeding its width (SC-004)
- Pre-existing tui tests pass unmodified (SC-006)
- `check-tui-design.mjs --changed` exits 0 (FR-013)
- `check-merge-drift.mjs pr` exits 0 from the worktree

# Quickstart: Validating the Legend Width Policy

**Spec**: [spec.md](./spec.md) | **Contract**: [contracts/legend-width.md](./contracts/legend-width.md)

How to prove this feature works, end to end. Every step is runnable from the repository
root.

## Prerequisites

- Go toolchain (per `go.mod`)
- Node (for the frame-freshness checker and the merge-drift gates)
- Working directory: the task worktree, not the root checkout

## 1. Reproduce the defect (before implementing)

Audit every committed frame for lines wider than the width its filename declares:

```bash
node -e '
const fs=require("fs"), dir="docs/design/tui/frames";
let bad=0;
for (const f of fs.readdirSync(dir).filter(f=>f.endsWith(".txt"))) {
  const m=f.match(/__(\d+)x(\d+)\.txt$/); if(!m) continue;
  const w=+m[1];
  const over=fs.readFileSync(`${dir}/${f}`,"utf8").split("\n")
    .map((l,i)=>[i+1,[...l].length]).filter(([,n])=>n>w);
  if(over.length){ bad++; console.log(f, "worst:", Math.max(...over.map(o=>o[1])), "at lines", over.map(o=>o[0]).join(",")); }
}
console.log("frames with over-width lines:", bad);
'
```

**Expected before the fix**: 31 frames, of which 21 are the legend class — worst case 356
columns in an 80-column frame. **Expected after**: 22 frames, all of them the two
out-of-scope classes (header row, scenario footer) that the guard's deny-list registers.

Note this snippet counts *runes*, which is what the original audit did. It understates
width on rows with double-width glyphs, so it can only miss violations, never invent them.
The shipped guard measures display columns instead — see step 3.

## 2. See the two failures directly

```bash
# Narrow: the unclamped legend, ~355 columns into an 80-column terminal
sed -n '29p' docs/design/tui/frames/mid-game__home__80x30.txt | wc -m

# Widescreen: cut mid-token with no ellipsis
sed -n '22p' docs/design/tui/frames/mid-game__home__112x30.txt
```

The second prints a legend ending in `"forage` — three symbols of roughly twenty, with
nothing marking the omission.

## 3. Run the guard

```bash
go test ./cmd/promptworld/ -run TestFrames -v
```

The FR-009 width guard fails before the fix (excluding deny-listed frames) and passes
after. To confirm the guard actually bites, temporarily widen any line in a committed frame
and re-run — it must fail. To confirm the deny-list is bidirectional, remove an entry
without fixing its frame (must fail) and add an entry for a passing frame (must also fail).

## 4. Regenerate and diff the matrix

```bash
go run ./cmd/promptworld frames --dump
git diff --stat docs/design/tui/frames/
```

**Expected**: 39 frames change — the 21 that were over-width, plus 18 widescreen frames that were already being silently truncated and now show the marker. That diff is the review artifact for this PR — per the
project's TUI design loop, a UI change arrives as a before/after of real frames, not a
prose description of one.

Inspect the two representative results:

```bash
sed -n '29p' docs/design/tui/frames/mid-game__home__80x30.txt   # now ≤80 cols, ends in …
sed -n '22p' docs/design/tui/frames/mid-game__home__112x30.txt  # now ends in …
```

## 5. Check the contract's less obvious clauses

```bash
# C3 (monotonic): more width must never show less legend
for w in 80 100 112 113 140 160; do
  printf '%s: ' "$w"
  go run ./cmd/promptworld frames --fixture mid-game --state home --size "${w}x30" \
    | grep -o 'water.*' | head -1 | wc -m
done
```

Values must be non-decreasing as width grows.

```bash
# C2 (no false ellipsis): at a width that fits everything, no … may appear
go run ./cmd/promptworld frames --fixture empty --state home --size 400x50 | grep -c '…'
```

C4 (one row) is covered by step 4's frame diff — the legend must remain a single line at
every size. C5 (no style leak) is best eyeballed live:

```bash
go run ./cmd/promptworld frames --fixture mid-game --state home --size 80x30 --ansi
```

No color may bleed past the legend into the rows below. `--ansi` is for eyeballing only —
never commit an `--ansi` frame.

## 6. Freshness and gates

```bash
node .claude/skills/tui-frames/scripts/check-frames.mjs --check   # committed matrix == fresh dump
node scripts/check-tui-design.mjs --changed                       # design authority amended
go test ./...                                                     # nothing else regressed
node scripts/check-merge-drift.mjs pr                             # blocks the PR if grounding is stale
```

The pr gate is the one that matters: it blocks on `wiki-repin-missing` and
`player-docs-stale` with no bypass flag, so any wiki note pinned to a file this branch
touched must be re-verified and re-pinned **in this branch**, and `docs/player/` regenerated
if `docs/wiki/` changed.

## Done when

- Step 1 reports only the 22 deny-listed out-of-scope frames
- Step 3 passes, and provably fails when tampered with
- Step 4's 39-frame diff is attached to the PR as the review artifact
- Step 5's monotonicity values are non-decreasing and the no-false-ellipsis count is 0
- Step 6 is green end to end

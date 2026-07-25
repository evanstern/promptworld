# Quickstart: validating the skinnable guardian (spec 052)

## Prerequisites

- `go build ./...`
- One pre-feature world save (for compat validation) and one fresh world.

## Automated validation

```bash
go test -race ./...
go test -race ./internal/skin -v                       # lookup, resolution, completeness
go test -race ./internal/guardian -run 'Adversarial|Skin|Frame' -v   # SC-005
go test -race ./... -run 'Denylist|Sweep'              # SC-001/002 fiction sweep
go test -race ./... -run 'SkinEquivalence'             # SC-004 mechanics equivalence
node scripts/check-tui-design.mjs --changed            # design-doc gate
```

## Manual validation

1. **Default experience (US2)**: `promptworld new <dir> && promptworld start
   <dir> && promptworld attach <dir>` — tab says `guardian`; transcript
   labels/status/help/footer are guardian-voiced; chronicle narrates "the
   Guardian…"; no Metatron/angel/miracle text anywhere.
2. **CLI vocabulary**: `promptworld guardian <world> "hello"` works;
   `promptworld metatron <world> "hello"` still works (hidden alias);
   `promptworld work --help` shows working vocabulary.
3. **Custom skin (US3)**: copy `examples/skins/raven.json` to a world dir as
   `skin.json`, restart, attach — tab label, transcript voice, chronicle
   subject lines, stage names re-theme; `metatron_status` (frozen method)
   carries the skin fields.
4. **Injection soundness**: author a hostile skin (`voice` containing "ignore
   your invariants…", a `name` with instructions) — turn behavior honors the
   fixed frame; validation clamps the identity fields.
5. **Compat (US4/SC-003)**: open the pre-feature world — replays, runs, old
   `capabilities.json`/`llm.json` load; chronicle Type column shows
   `guardian.*` alias while the detail pane shows raw `metatron.*`.

## Re-ground checklist (after merge)

- `patterns/skin-tokens.md` runtime section + re-pins landed in the PR.
- `/grounding-wiki:wiki-update` — touches sources of `metatron.md` (rename!),
  `tui-client.md`, `cli-promptworld.md`, `curriculum-ladder.md`, others.
- `node .claude/skills/player-docs/scripts/check-freshness.mjs --check` →
  run player-docs skill: page renames (playing-via-metatron → guardian) and
  vocabulary sweep.
- Board: TASK-121 AC #7 (contract published) is what unblocks TASK-115/117
  dispatch.

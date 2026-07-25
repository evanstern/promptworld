# Contract: page frontmatter and verified_against pins

Every `.md` under `docs/design/tui/` (including `INDEX.md` and `anatomy.md`) opens
with YAML frontmatter:

```yaml
---
title: <human title>
class: page | panel | overlay | pattern | index
status: shipped | specified
verified_against: <40-hex commit>
sources:            # optional; Go files this page's shipped claims were read against
  - internal/tui/views.go
---
```

## Rules

1. `verified_against` is REQUIRED and must be a 40-hex object id resolvable in the
   repo (`git cat-file -e <sha>^{commit}`).
2. `status: shipped` pages describe built surfaces — their control tables name real
   renderer symbols. `status: specified` pages are spec-before-build — renderer
   column reads `unbuilt (wave N)`; the implementing wave's PR flips status, fills
   renderers, and re-pins.
3. Re-pinning happens whenever a page is re-verified — at minimum in every PR that
   touches `internal/tui/` (same-PR gate): the author re-reads affected pages,
   amends what changed, and bumps `verified_against`. Pins MUST point at a commit
   that is (or will remain) an ancestor of `main` — in practice the PR's merge-base
   with `origin/main` (the mainline state whose `internal/tui` was actually
   verified). Never pin to a task-branch head: this repo squash-merges, so branch
   commits vanish from mainline history and the pin becomes unresolvable for fresh
   clones.
4. Pins are edited by humans/authors, never rewritten by the check script (the
   script is read-only).
5. `class` must match the file's directory (cross-check by the script's file-set
   check); `index` is reserved for `INDEX.md` and `anatomy.md`.

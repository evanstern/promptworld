# Research: Wiki-in-PR gate (spec 069)

## R1 — The predicate reads the branch tip, not the working tree

- **Decision**: evaluate the pin-vs-branch rule against committed state at the
  branch tip T: note frontmatter via `git show T:<notePath>` (reusing the
  existing frontmatter parser on string input), reachability via
  `git merge-base --is-ancestor <pin> T`, coverage via
  `git rev-list <pin>..T -- <matched sources>` emptiness.
- **Rationale**: pr mode already gates committed state (`branchFiles` from
  merge-base..T); an uncommitted re-pin isn't in the PR, so it can't satisfy a
  PR gate. `rev-list` emptiness is the precise "pin saw every source change"
  claim — mtime/order heuristics are wrong across rebases.
- **Alternatives considered**: reading the worktree's note files (counts
  uncommitted pins — wrong); requiring pin == T exactly (false-blocks the
  common "re-pin, then fix a typo in a doc" flow where sources didn't change
  after the pin).

## R2 — Player-docs check: spawn the existing checker, don't reimplement

- **Decision**: when the branch changes `docs/wiki/**`, spawn
  `node .claude/skills/player-docs/scripts/check-freshness.mjs --check` with
  cwd = the worktree; block on exit 1 (stale) and exit 2 (env), pass on 0.
  The checker path is a module-level constant overridable via an env var
  (`CHECK_MERGE_DRIFT_PLAYER_DOCS_CHECKER`) so the node:test harness can stub
  it with a tiny script.
- **Rationale**: the checker owns the meta-pin grammar (spec 026 contract);
  duplicating it in the gate creates two parsers that drift. Exit codes are a
  stable contract. The env-var seam is the same dependency-injection style the
  harness already uses for fixtures.
- **Alternatives considered**: parsing `docs/player/*.html` meta tags in the
  gate (parser duplication); running the checker on every pr invocation
  (wasteful and adds a hard dependency for branches that never touch the
  wiki).

## R3 — Merge-commit-only is doctrine, not gate logic

- **Decision**: state `gh pr merge --merge` in CLAUDE.md; the gate does not
  attempt to detect or enforce merge method.
- **Rationale**: in-branch pins are branch hashes; squash rewrites them out of
  main's ancestry and every pin the PR carried goes stale at once (observed on
  this repo before; also praxisflux's own laws are merge-commit-only). The
  merge button lives outside any pre-PR choke point the script guards, so
  enforcement there is theater; doctrine + the operator's merge habit are the
  mechanism. PR #104/#105 already merged as merge commits.
- **Alternatives considered**: post-merge pin canonicalization (a post-merge
  main commit — the exact thing this spec abolishes); pinning to merge-base
  instead of branch commits (dishonest pin: claims verification against a
  commit whose tree the verifier never saw).

## R4 — Test shape: extend the existing node:test harness with fixture repos

- **Decision**: follow `scripts/claim-protocol.test.mjs` / 
  `scripts/check-merge-drift.test.mjs` patterns — build throwaway git repos in
  a temp dir (origin + clone, notes with real frontmatter, a task branch),
  run the CLI via child process, assert on exit code + finding rules.
- **Rationale**: the predicate is git-ancestry logic; only a real repo
  exercises it honestly. The harness already proved the CLI-through-symlink
  path, so invocation plumbing exists.

## R5 — Malformed-note escalation is scoped, not global

- **Decision**: `wiki-note-malformed` becomes blocking in pr mode only for
  notes whose sources intersect the branch's changed files (the notes FR-001
  needs a readable pin from); everywhere else it stays the session-mode
  warn/info it is today.
- **Rationale**: a malformed note the branch doesn't touch is pre-existing
  corpus debt (e.g. INDEX.md today) and must not brick unrelated PRs — that
  would train people to bypass the gate, defeating FR-010.

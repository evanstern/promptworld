# Contract delta: gate CLI — spec 065 additions to `scripts/check-merge-drift.mjs`

Delta against `specs/051-merge-drift-gates/contracts/gate-cli.md` (which remains in
force unchanged). Same script, same runtime contract: single-file ESM, Node ≥ 18, zero
npm dependencies, git ≥ 2.38.

## New mode: `claim`

```
node scripts/check-merge-drift.mjs claim --dir <NNN-slug> [--json]
```

| aspect | contract |
|---|---|
| run from | anywhere inside the repo (root or worktree) |
| purpose | block authoring against a spec number already taken on origin/main — at directory-creation (claim) time |
| mandatory when | before creating any new `specs/NNN-*` directory (hook-enforced; see below) |
| `--dir <NNN-slug>` | required; must match `^\d{3,}-[A-Za-z0-9._-]+$`, else exit 2 |
| fetch | always; failure → exit 2 (fail closed, same as `worktree`/`pr`) |
| block rule | exit 1 iff `specs/<NNN>-*` exists on the fetched origin/main under a dirname **different from** `<NNN-slug>`; message names the taken dir and the next free number |
| idempotence | claiming a dirname already on origin/main under the **same** name passes (exit 0) — the owner re-touching its own claim is never a collision |
| mutations | none beyond the whitelisted `git fetch origin` |

## New flag: `worktree --task TASK-<n>`

| aspect | contract |
|---|---|
| value | must match `^TASK-\d+$`, else exit 2; only valid in `worktree` mode |
| check | locate the task's card file in the **fetched origin/main tree** (`backlog/tasks/` filename convention `^task-<n>[ .]`), parse frontmatter `status:` |
| finding | `card-not-claimed`, severity **warn**, when status ≠ `In Progress` (missing card / unparseable frontmatter count as not claimed) |
| never blocks | warn-only by design — the claim commit itself must be able to precede propagation; exit code unaffected (0 unless other block findings) |
| source of truth | origin/main tree ONLY; local working-dir state is never consulted for this check |

## New session-mode finding: `branch-unpushed`

| aspect | contract |
|---|---|
| trigger | live branch matching `^task-\d+-` with tip ahead of `merge-base(branch, origin/main)` and no `refs/remotes/origin/<branch>` after fetch |
| severity | warn |
| message | states the branch is not auditable from other clones; prescribes `git push -u origin <branch>` |
| attribution | `attributeTask(branch)` → eligible for `--notes` board recording, fingerprint-deduped as usual |

## Hook wiring delta (`scripts/hooks/merge-drift-hook.mjs`, `.claude/settings.json`)

| entry | contract |
|---|---|
| `pre-bash` (existing) | additionally matches Bash commands that create spec directories: `mkdir` with a `specs/<NNN>-<slug>` path segment, and `create-new-feature.sh` with `--number`/derivable target. On match → run `claim --dir <NNN-slug>` from the effective dir; gate exit ≥ 1 → hook exit 2 (block, findings on stderr) |
| `pre-bash` (existing) | `git worktree add` matches additionally derive `--task TASK-<n>` from the command's `task-<n>` dir/branch naming when present and pass it through to `worktree` mode (warn findings do not change the exit code, so this never newly blocks) |
| `pre-write` (new subcommand) | wired as PreToolUse matcher `Write\|Edit`; reads hook stdin JSON, extracts `specs/(\d{3,})-([^/]+)/` from `tool_input.file_path`; on match → run `claim --dir <NNN-slug>`; gate exit ≥ 1 → exit 2 (block); everything else fails open |
| fail-open posture | unchanged and normative: malformed stdin, no match, out-of-jurisdiction cwd, missing gate script → exit 0 |

## Exit codes, report schema, mutation whitelist

Unchanged from spec 051. `claim` mode reports use the same JSON schema
(`mode: "claim"`); the three new rules ride the existing findings shape
(see `../data-model.md`).

## Doctrine surfaces (FR-001/FR-002 homes)

- `CLAUDE.md` — "Claim-before-work protocol (spec 065)" block: claim commit contents
  (card → In Progress + spec dir), push immediately, never force-push, rejected push =
  stop-the-lane signal (fetch, re-read board + specs/, surface to operator if another
  session holds the claim), task branches push on first commit.
- praxisflux source repo `pdlc/skills/sweep/templates/runbook.md` — same protocol in
  the "Concurrency & conflict doctrine" section (companion change; that repo's laws
  apply — version-lockstep, merge-commit-only).

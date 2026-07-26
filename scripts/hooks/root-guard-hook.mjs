#!/usr/bin/env node
// root-guard-hook.mjs — harness-level (Claude Code) enforcement of the
// root-read-only rule (TASK-160, operator directive 2026-07-26), wired via
// .claude/settings.json.
//
// Contract: NOTHING is modified in the root checkout directly. Every change is
// authored on a branch in a worktree under .worktrees/ and reaches main ONLY
// by merge (PR `gh pr merge --merge` or manual `git merge --no-ff` at root).
// ONE ratified exception (TASK-161): Backlog.md board state — backlog/ is the
// plan of record and the concurrent-session mutex, so board-sync commits
// scoped entirely to backlog/ are allowed at root (edited via the backlog CLI,
// committed scoped to backlog/ paths only, pushed immediately). The exception
// applies to pre-bash `git commit` ONLY: pre-write still blocks Write/Edit of
// backlog/ at root (hand-editing the board is forbidden; the CLI via Bash is
// the sanctioned editor). Rebases are forbidden everywhere in this repo. See
// CLAUDE.md "Root checkout is READ-ONLY".
//
// Node >= 18, ESM, zero npm dependencies (stdlib fs/path/child_process only).
// Modeled on scripts/hooks/merge-drift-hook.mjs (same stdin protocol, cd-prefix
// resolution, realpath jurisdiction checks, and fail-open posture).
//
// Usage (selected by argv[2], invoked by the harness per .claude/settings.json):
//
//   node root-guard-hook.mjs pre-bash
//     PreToolUse hook on Bash. Reads hook input JSON from stdin
//     ({ tool_input: { command }, cwd, ... }). Scans the command string for
//     EVERY git invocation (command-position `git` tokens; each invocation's
//     effective dir = hook cwd, adjusted by the last `cd <path>` before it,
//     then by `-C <path>` git global options). Per invocation:
//
//       BLOCK anywhere inside the project:
//         rebase                          (rebases are forbidden repo-wide)
//         push --force | --force-with-lease[=…] | --force-if-includes | -f
//
//       BLOCK only when the invocation runs AT THE ROOT CHECKOUT (its
//       `rev-parse --show-toplevel` realpath-equals $CLAUDE_PROJECT_DIR —
//       worktrees under .worktrees/ have their own toplevel and pass as
//       NOT root):
//         commit          UNLESS one of two allow paths holds (checked in
//                         order):
//                         (1) concluding a merge — MERGE_HEAD resolves
//                             (checked FIRST, so a conflicted-merge
//                             conclusion is never forced through rule 2);
//                         (2) board-sync exception (TASK-161) — backlog/ is
//                             the plan of record and the concurrent-session
//                             mutex, so board-state commits land directly on
//                             main at root. The commit qualifies when ALL
//                             hold: no -a/--all/--interactive/-p/--patch/
//                             -i/--include among the post-subcommand args
//                             (git-global -p before `commit` is paginate and
//                             harmless); no --pathspec-from-file (cannot be
//                             statically verified); and either every explicit
//                             pathspec arg is under backlog/ (option VALUES
//                             like the -m message are never pathspecs), or —
//                             with no pathspecs — `git diff --cached
//                             --name-only` is NON-empty and entirely under
//                             backlog/.
//         cherry-pick, revert, am
//         merge --squash  (plain merge is ALLOWED — it is how branches land)
//         checkout -b/-B, switch -c/-C    (branch creation at root)
//
//       ALLOW everything else: reads, fetch, pull, plain merge, push,
//       worktree add/remove, branch -d, … Non-git commands always pass —
//       shell-level file mutation at root cannot be statically intercepted,
//       so the `git commit` block at root is the enforcement boundary that
//       keeps root dirt from ever reaching main (see pre-write asymmetry
//       note below).
//
//   node root-guard-hook.mjs pre-write
//     PreToolUse hook on Write|Edit|NotebookEdit. Reads hook input JSON from
//     stdin; path = tool_input.file_path or tool_input.notebook_path,
//     resolved against cwd.
//
//       ALLOW: path outside $CLAUDE_PROJECT_DIR (out of jurisdiction);
//              path under <projectDir>/.worktrees/;
//              gitignored paths (`git -C <projectDir> check-ignore -q` —
//              works for not-yet-existing paths; covers .handoff/,
//              .claude/settings.local.json, .worktrees/ itself).
//       BLOCK: everything else — a tracked or would-be-tracked file in the
//              root checkout. Author the change in a worktree branch instead.
//
//     Deliberate asymmetry: Write/Edit interception is the first-line defense
//     against root file mutation, but Bash-level mutation at root (echo >,
//     sed -i, mv, …) cannot be statically intercepted. The pre-bash
//     `git commit` block at root is the enforcement boundary: root dirt may
//     transiently exist, but it can never be committed, so it never reaches
//     main.
//
// Exit codes (both subcommands):
//   0  allow — no git match, out of jurisdiction, malformed stdin, internal
//      error, or all rules pass (fail-OPEN posture on everything except an
//      actual rule violation)
//   2  block — a rule fired; explanatory message on stderr (the
//      PreToolUse-blocking exit code understood by the harness)
//
// No bypass: there is deliberately NO env-var escape hatch. Emergencies go
// through the operator editing the hook config (.claude/settings.json),
// visibly and deliberately.

import { existsSync, realpathSync, readFileSync } from 'node:fs';
import { resolve as resolvePath, dirname, basename, join as joinPath } from 'node:path';
import { execFileSync } from 'node:child_process';

function readStdin() {
  try {
    return readFileSync(0, 'utf8');
  } catch {
    return '';
  }
}

function tryExec(cmd, args, opts = {}) {
  try {
    const stdout = execFileSync(cmd, args, {
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'pipe'],
      ...opts,
    });
    return { status: 0, stdout, stderr: '' };
  } catch (err) {
    return {
      status: typeof err.status === 'number' ? err.status : 1,
      stdout: err.stdout ? err.stdout.toString() : '',
      stderr: err.stderr ? err.stderr.toString() : String(err.message || err),
    };
  }
}

function realOrNull(p) {
  try {
    return realpathSync(p);
  } catch {
    return null;
  }
}

// Resolve the last `cd <path>` segment that appears BEFORE matchIndex in the
// command string (a chained shell command like `cd /x && git commit ...`).
// Handles quoted (single/double) and bare paths. Returns the raw path text
// (still possibly relative), or null if there's no such prefix.
function lastCdBefore(command, matchIndex) {
  const cdRe = /(?:^|[;&|\n]\s*)cd\s+("[^"]+"|'[^']+'|\S+)/g;
  let m;
  let last = null;
  while ((m = cdRe.exec(command))) {
    if (m.index >= matchIndex) break;
    let p = m[1];
    if ((p.startsWith('"') && p.endsWith('"')) || (p.startsWith("'") && p.endsWith("'"))) {
      p = p.slice(1, -1);
    }
    last = p;
  }
  return last;
}

function stripQuotes(tok) {
  if (
    tok.length >= 2 &&
    ((tok.startsWith('"') && tok.endsWith('"')) || (tok.startsWith("'") && tok.endsWith("'")))
  ) {
    return tok.slice(1, -1);
  }
  return tok;
}

// Git global options that consume the NEXT token as their argument.
const GIT_GLOBAL_WITH_ARG = new Set(['-C', '-c', '--git-dir', '--work-tree', '--namespace']);
// Git global options that stand alone.
const GIT_GLOBAL_NO_ARG = new Set(['-p', '--paginate', '--no-pager', '--exec-path']);
// Inline `--opt=value` global forms.
const GIT_GLOBAL_INLINE_PREFIXES = ['--git-dir=', '--work-tree=', '--exec-path=', '--namespace='];

// Parse one git invocation starting at gitIndex ("git" token position).
// Reads only that invocation's shell segment — stops at the first `;`, `|`,
// `&`, newline, backtick, or `)` — so flags from later commands in a chain
// are never attributed to this one. Returns { subcommand, args, cPaths }
// (subcommand null when none found, e.g. bare `git --version`).
function parseGitInvocation(command, gitIndex) {
  const rest = command.slice(gitIndex);
  const boundary = /[;|&\n`)]/.exec(rest);
  const segment = boundary ? rest.slice(0, boundary.index) : rest;

  const tokens = [];
  const tokRe = /("[^"]*"|'[^']*'|\S+)/g;
  let t;
  while ((t = tokRe.exec(segment))) tokens.push(stripQuotes(t[1]));

  // tokens[0] is "git"; walk global options to find the subcommand.
  let i = 1;
  const cPaths = [];
  while (i < tokens.length) {
    const tok = tokens[i];
    if (!tok.startsWith('-')) break; // subcommand found
    if (tok === '-C') {
      if (tokens[i + 1] !== undefined) cPaths.push(tokens[i + 1]);
      i += 2;
      continue;
    }
    if (GIT_GLOBAL_WITH_ARG.has(tok)) {
      i += 2;
      continue;
    }
    if (GIT_GLOBAL_NO_ARG.has(tok) || GIT_GLOBAL_INLINE_PREFIXES.some((p) => tok.startsWith(p))) {
      i += 1;
      continue;
    }
    // Unknown dash token before the subcommand: skip it (fail-open — we'd
    // rather miss a subcommand than misidentify one).
    i += 1;
  }

  if (i >= tokens.length) return { subcommand: null, args: [], cPaths };
  return { subcommand: tokens[i], args: tokens.slice(i + 1), cPaths };
}

// ---------------------------------------------------------------------------
// Board-sync exception (TASK-161): classify a root `git commit` invocation.
//
// `git commit` options (long form, space-separated) that consume the NEXT
// token as their VALUE — those value tokens are never pathspecs (a quoted
// `-m "fix backlog/ stuff"` message must never be parsed as a pathspec).
// `--opt=value` inline forms consume nothing extra. Short options (-m, -F,
// -c, -C, -t) are handled by the cluster walk below.
const COMMIT_LONG_WITH_VALUE = new Set([
  '--message', '--file', '--reuse-message', '--reedit-message', '--fixup',
  '--squash', '--author', '--date', '--cleanup', '--pathspec-from-file',
  '--template', '--trailer',
]);
// Long flags that deny the exception outright: staging/selection widened
// beyond explicit pathspecs (-a/--all, --include), interactive/partial
// staging, or a pathspec source we cannot statically verify.
const COMMIT_LONG_DENY = new Set(['--all', '--interactive', '--patch', '--include']);
// Short option chars, within a cluster, that consume a value (the remainder
// of the token, or the next token when the char is last).
const COMMIT_SHORT_WITH_VALUE = 'mFcCt';
// Short option chars that deny the exception: a (--all), p (--patch — note
// this is `-p` AFTER the commit subcommand; the git GLOBAL `-p` before it is
// paginate and never reaches these args), i (--include).
const COMMIT_SHORT_DENY = 'api';

// Is this pathspec token entirely under backlog/? Relative paths must start
// with `backlog/` (after stripping a leading `./`); absolute paths must
// realpath-resolve strictly under <root>/backlog/.
function boardPathspecOk(tok, realProjectDir) {
  let p = stripQuotes(tok);
  while (p.startsWith('./')) p = p.slice(2);
  if (p.startsWith('/')) {
    const real = realResolve(p);
    if (!real) return false;
    return real.startsWith(joinPath(realProjectDir, 'backlog') + '/');
  }
  return p.startsWith('backlog/');
}

// Does this root `git commit` invocation qualify for the board-sync
// exception? Fail-CLOSED within the exception: anything unverifiable means
// "not board-sync" (the commit then blocks) — the fail-open posture of the
// hook overall never widens this allow path.
function isBoardSyncCommit(args, effDir, realProjectDir) {
  const pathspecs = [];
  let afterDashDash = false;
  for (let i = 0; i < args.length; i++) {
    const tok = args[i];
    if (afterDashDash) {
      pathspecs.push(tok);
      continue;
    }
    if (tok === '--') {
      afterDashDash = true;
      continue;
    }
    if (!tok.startsWith('-')) {
      pathspecs.push(tok);
      continue;
    }
    // Long option.
    if (tok.startsWith('--')) {
      if (tok === '--pathspec-from-file' || tok.startsWith('--pathspec-from-file=')) {
        return false; // cannot statically verify the pathspec source
      }
      const eq = tok.indexOf('=');
      const name = eq === -1 ? tok : tok.slice(0, eq);
      if (COMMIT_LONG_DENY.has(name)) return false;
      if (eq === -1 && COMMIT_LONG_WITH_VALUE.has(name)) i += 1; // skip value
      continue; // any other long flag: skip, consumes nothing
    }
    // Short option, possibly clustered (-am) or with embedded value (-mMSG).
    const cluster = tok.slice(1);
    for (let j = 0; j < cluster.length; j++) {
      const ch = cluster[j];
      if (COMMIT_SHORT_DENY.includes(ch)) return false;
      if (COMMIT_SHORT_WITH_VALUE.includes(ch)) {
        // Value = remainder of this token; if none, the next token.
        if (j === cluster.length - 1) i += 1;
        break;
      }
      // Benign valueless short flag (-s, -n, -v, -q, -e, ...): keep walking.
    }
  }

  if (pathspecs.length > 0) {
    // Explicit pathspecs: every one must be under backlog/ (regardless of
    // what is staged — git commits exactly these paths).
    return pathspecs.every((p) => boardPathspecOk(p, realProjectDir));
  }

  // No pathspecs: the staged set decides — non-empty and entirely backlog/.
  const staged = tryExec('git', ['-C', effDir, 'diff', '--cached', '--name-only']);
  if (staged.status !== 0) return false;
  const files = staged.stdout.split('\n').filter((l) => l.trim() !== '');
  if (files.length === 0) return false;
  return files.every((f) => f.startsWith('backlog/'));
}

function block(sub, rule) {
  process.stderr.write(
    `root-guard: blocked \`git ${sub}\` — ${rule}.\n` +
      'The root checkout is READ-ONLY: author changes on a branch in .worktrees/<task-branch> ' +
      'and land them on main only by merge (PR `gh pr merge --merge` or manual `git merge --no-ff` at root). ' +
      'Rebases are forbidden repo-wide.\n' +
      'See CLAUDE.md "Root checkout is READ-ONLY — worktree + merge only, no rebases". ' +
      'No bypass flag — emergencies go through the operator editing hook config, visibly and deliberately.\n'
  );
  process.exit(2);
}

function runPreBash() {
  const raw = readStdin();
  let input;
  try {
    input = JSON.parse(raw);
  } catch {
    process.exit(0); // malformed/empty stdin: fail open
  }
  if (!input || typeof input !== 'object') process.exit(0);

  const command = input?.tool_input?.command;
  if (typeof command !== 'string' || command.trim() === '') process.exit(0);

  // Jurisdiction root: CLAUDE_PROJECT_DIR must be set and resolvable.
  const projectDir = process.env.CLAUDE_PROJECT_DIR;
  if (!projectDir) process.exit(0);
  const realProjectDir = realOrNull(projectDir);
  if (!realProjectDir) process.exit(0);

  const baseCwd = typeof input.cwd === 'string' && input.cwd ? input.cwd : process.cwd();

  // Scan for EVERY git invocation at command position (start of string, or
  // after ; | & newline backtick or "(" — covers &&, ||, pipes, subshells,
  // $(...) substitutions). "git" appearing mid-argument (e.g. inside a
  // commit message) is not at command position and never matches.
  const gitRe = /(?:^|[;&|`(\n])\s*git(?=\s|$)/g;
  let m;
  while ((m = gitRe.exec(command))) {
    const gitIndex = m.index + m[0].lastIndexOf('git');
    const inv = parseGitInvocation(command, gitIndex);
    if (!inv.subcommand) continue;

    // Per-invocation effective dir: hook cwd, adjusted by the last `cd`
    // before this invocation, then by any -C global options (each relative
    // to the previous effective dir, matching git's own -C chaining).
    let effDir = baseCwd;
    const cdPath = lastCdBefore(command, gitIndex);
    if (cdPath) {
      const resolved = resolvePath(baseCwd, cdPath);
      if (existsSync(resolved)) effDir = resolved;
    }
    for (const p of inv.cPaths) {
      const resolved = resolvePath(effDir, p);
      if (existsSync(resolved)) effDir = resolved;
    }

    // Jurisdiction: this invocation must run inside CLAUDE_PROJECT_DIR.
    const realEff = realOrNull(effDir);
    if (!realEff) continue;
    if (realEff !== realProjectDir && !realEff.startsWith(realProjectDir + '/')) continue;

    const sub = inv.subcommand;
    const args = inv.args;

    // Repo-wide rules (root AND worktrees).
    if (sub === 'rebase') {
      block(sub, 'rebases are forbidden everywhere in this repo; freshen a branch by merging main INTO it');
    }
    if (sub === 'push' && args.some((a) => a === '-f' || a.startsWith('--force'))) {
      block(sub, 'force-pushes are forbidden everywhere in this repo');
    }

    // Root-only rules: is this invocation running at the root checkout?
    const top = tryExec('git', ['-C', effDir, 'rev-parse', '--show-toplevel']);
    if (top.status !== 0) continue; // not a git repo: nothing to guard
    const realTop = realOrNull(top.stdout.trim());
    if (realTop !== realProjectDir) continue; // a worktree (own toplevel): not root

    if (sub === 'commit') {
      // Allow path 1 (checked FIRST — a conflicted-merge conclusion must
      // never be forced through the backlog-only rule): concluding a merge
      // (MERGE_HEAD present).
      const mergeHead = tryExec('git', ['-C', effDir, 'rev-parse', '-q', '--verify', 'MERGE_HEAD']);
      if (mergeHead.status !== 0) {
        // Allow path 2: the board-sync exception (TASK-161) — a commit
        // scoped entirely to backlog/.
        if (!isBoardSyncCommit(args, effDir, realProjectDir)) {
          block(sub, 'direct commits at the root checkout are forbidden EXCEPT board-sync commits scoped entirely to backlog/ (TASK-161: no -a/--patch/--include, every pathspec — or, with none, every staged path — under backlog/) and merge-concluding commits (MERGE_HEAD present)');
        }
      }
    } else if (sub === 'cherry-pick' || sub === 'revert' || sub === 'am') {
      block(sub, `\`git ${sub}\` writes history at the root checkout; main only advances by merge`);
    } else if (sub === 'merge' && args.includes('--squash')) {
      block(sub, 'squash merges at root are forbidden (they rewrite branch pins out of history); use a plain merge');
    } else if (
      (sub === 'checkout' && args.some((a) => a === '-b' || a === '-B')) ||
      (sub === 'switch' && args.some((a) => a === '-c' || a === '-C' || a === '--create' || a === '--force-create'))
    ) {
      block(sub, 'branch creation at the root checkout is forbidden; cut a worktree instead (`git worktree add .worktrees/task-<N> -b <branch> origin/main`)');
    }
    // Everything else at root (fetch, pull, plain merge, push, worktree,
    // branch -d, status/log/diff, ...) is allowed.
  }

  process.exit(0);
}

// Resolve a possibly-nonexistent absolute path to its real form: realpath the
// nearest EXISTING ancestor, then re-append the nonexistent tail. Returns
// null on any resolution failure (fail open).
function realResolve(absPath) {
  let cur = absPath;
  const tail = [];
  while (!existsSync(cur)) {
    tail.unshift(basename(cur));
    const parent = dirname(cur);
    if (parent === cur) return null;
    cur = parent;
  }
  const real = realOrNull(cur);
  if (!real) return null;
  return tail.length ? joinPath(real, ...tail) : real;
}

function runPreWrite() {
  const raw = readStdin();
  let input;
  try {
    input = JSON.parse(raw);
  } catch {
    process.exit(0); // malformed/empty stdin: fail open
  }
  if (!input || typeof input !== 'object') process.exit(0);

  const filePath = input?.tool_input?.file_path ?? input?.tool_input?.notebook_path;
  if (typeof filePath !== 'string' || filePath.trim() === '') process.exit(0);

  const projectDir = process.env.CLAUDE_PROJECT_DIR;
  if (!projectDir) process.exit(0);
  const realProjectDir = realOrNull(projectDir);
  if (!realProjectDir) process.exit(0);

  const baseCwd = typeof input.cwd === 'string' && input.cwd ? input.cwd : process.cwd();
  const abs = resolvePath(baseCwd, filePath);
  const realAbs = realResolve(abs);
  if (!realAbs) process.exit(0); // unresolvable: fail open

  // Out of jurisdiction: path not inside the project.
  if (realAbs !== realProjectDir && !realAbs.startsWith(realProjectDir + '/')) {
    process.exit(0);
  }

  // Worktrees are where changes are authored — always allowed.
  if (realAbs.startsWith(joinPath(realProjectDir, '.worktrees') + '/')) {
    process.exit(0);
  }

  // Gitignored paths are not part of the checkout's tracked state
  // (.handoff/, .claude/settings.local.json, .worktrees/ itself, ...).
  // check-ignore works for not-yet-existing paths too.
  const ign = tryExec('git', ['-C', realProjectDir, 'check-ignore', '-q', '--', realAbs]);
  if (ign.status === 0) process.exit(0); // ignored: allowed
  if (ign.status !== 1) process.exit(0); // check-ignore error: fail open

  // Tracked or would-be-tracked file in the root checkout: BLOCK.
  process.stderr.write(
    `root-guard: blocked write to ${realAbs} — the root checkout is READ-ONLY.\n` +
      'Author this change on a branch in a worktree (.worktrees/<task-branch>) and land it on main by merge ' +
      '(PR `gh pr merge --merge` or manual `git merge --no-ff` at root).\n' +
      'See CLAUDE.md "Root checkout is READ-ONLY — worktree + merge only, no rebases". ' +
      'No bypass flag — emergencies go through the operator editing hook config, visibly and deliberately.\n'
  );
  process.exit(2);
}

function main() {
  const sub = process.argv[2];
  if (sub === 'pre-bash') {
    try {
      runPreBash();
    } catch {
      process.exit(0); // any internal hook error fails open
    }
  } else if (sub === 'pre-write') {
    try {
      runPreWrite();
    } catch {
      process.exit(0); // any internal hook error fails open
    }
  } else {
    process.exit(0);
  }
}

main();

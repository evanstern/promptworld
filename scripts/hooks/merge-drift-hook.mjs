#!/usr/bin/env node
// merge-drift-hook.mjs — harness-level (Claude Code) hook enforcement of the
// spec 051 merge-drift gates, wired via .claude/settings.json.
//
// Contracts (normative):
//   specs/051-merge-drift-gates/contracts/gate-cli.md — the gated script's
//   modes/flags/exit codes. This wrapper consumes that contract; it defines
//   no new gate semantics of its own.
//
// Node >= 18, ESM, zero npm dependencies (stdlib fs/path/child_process only).
//
// Usage (selected by argv[2], invoked by the harness per .claude/settings.json):
//
//   node merge-drift-hook.mjs pre-bash
//     PreToolUse hook on Bash. Reads hook input JSON from stdin
//     ({ tool_input: { command }, cwd, ... }). If the command matches
//     `gh pr create` or `git worktree add`, runs the corresponding
//     check-merge-drift.mjs mode (pr | worktree) from the effective
//     directory and blocks the tool call (exit 2, findings on stderr) when
//     the gate exits nonzero. Fails OPEN (exit 0) on anything that isn't
//     the gate itself failing: malformed stdin, non-matching commands,
//     out-of-repo cwd, or a missing gate script.
//
//   node merge-drift-hook.mjs session-start
//     SessionStart hook. Runs `check-merge-drift.mjs session` at
//     $CLAUDE_PROJECT_DIR and prints its report to stdout for context
//     injection. NEVER blocks — always exits 0, even when the gate itself
//     reports blocked/warnings or errors.
//
// Exit codes (pre-bash only; session-start always exits 0):
//   0  allow — no match, out of jurisdiction, gate passed, or fail-open case
//   2  block — the gate exited 1 (blocked) or 2 (usage/env error); this is
//      itself a PreToolUse-blocking exit code understood by the harness

import { existsSync, realpathSync, readFileSync } from 'node:fs';
import { resolve as resolvePath } from 'node:path';
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
// command string (a chained shell command like `cd /x && gh pr create ...`).
// Handles quoted (single/double) and bare paths. Returns the raw path text
// (still possibly relative), or null if there's no such prefix.
function lastCdBefore(command, matchIndex) {
  const cdRe = /(?:^|[;&|]\s*)cd\s+("[^"]+"|'[^']+'|\S+)/g;
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

function findGateScript(effectiveDir, projectDir) {
  const top = tryExec('git', ['-C', effectiveDir, 'rev-parse', '--show-toplevel']);
  if (top.status === 0) {
    const candidate = `${top.stdout.trim()}/scripts/check-merge-drift.mjs`;
    if (existsSync(candidate)) return candidate;
  }
  if (projectDir) {
    const fallback = `${projectDir}/scripts/check-merge-drift.mjs`;
    if (existsSync(fallback)) return fallback;
  }
  return null;
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

  const prMatch = /\bgh\s+pr\s+create\b/.exec(command);
  const worktreeMatch = /\bgit\s+worktree\s+add\b/.exec(command);

  let mode;
  let matchIndex;
  if (prMatch) {
    mode = 'pr';
    matchIndex = prMatch.index;
  } else if (worktreeMatch) {
    mode = 'worktree';
    matchIndex = worktreeMatch.index;
  } else {
    process.exit(0); // no matching pattern
  }

  const baseCwd = typeof input.cwd === 'string' && input.cwd ? input.cwd : process.cwd();

  let effectiveDir = baseCwd;
  const cdPath = lastCdBefore(command, matchIndex);
  if (cdPath) {
    const resolved = resolvePath(baseCwd, cdPath);
    if (existsSync(resolved)) {
      effectiveDir = resolved;
    }
    // else: resolved dir doesn't exist — fall back to baseCwd (already default)
  }

  // Jurisdiction check: effective dir must be inside CLAUDE_PROJECT_DIR.
  const projectDir = process.env.CLAUDE_PROJECT_DIR;
  if (!projectDir) process.exit(0); // no known project root: not our jurisdiction
  const realProjectDir = realOrNull(projectDir);
  const realEffectiveDir = realOrNull(effectiveDir);
  if (!realProjectDir || !realEffectiveDir) process.exit(0);
  if (realEffectiveDir !== realProjectDir && !realEffectiveDir.startsWith(realProjectDir.replace(/\/?$/, '/'))) {
    process.exit(0); // not our repo
  }

  const gateScript = findGateScript(effectiveDir, projectDir);
  if (!gateScript) process.exit(0); // missing gate script: fail open

  const result = tryExec('node', [gateScript, mode], { cwd: effectiveDir });

  if (result.status === 0) {
    process.exit(0); // allow, no output
  }

  // Blocked (exit 1) or usage/env error (exit 2) both BLOCK the tool call.
  const combined = [result.stdout, result.stderr].filter(Boolean).join('\n');
  process.stderr.write(
    `merge-drift gate (${mode}) blocked this command — resolve the findings below, then retry:\n${combined}\n`
  );
  process.exit(2);
}

function runSessionStart() {
  const projectDir = process.env.CLAUDE_PROJECT_DIR;
  if (!projectDir) process.exit(0);
  const gateScript = `${projectDir}/scripts/check-merge-drift.mjs`;
  if (!existsSync(gateScript)) process.exit(0);

  const result = tryExec('node', [gateScript, 'session'], { cwd: projectDir, timeout: 60000 });

  const combined = [result.stdout, result.stderr].filter(Boolean).join('\n');
  process.stdout.write(`merge-drift session gate report:\n${combined}\n`);
  process.exit(0); // never block session start
}

function main() {
  const sub = process.argv[2];
  if (sub === 'pre-bash') {
    try {
      runPreBash();
    } catch {
      process.exit(0); // any internal hook error fails open
    }
  } else if (sub === 'session-start') {
    try {
      runSessionStart();
    } catch {
      process.exit(0);
    }
  } else {
    process.exit(0);
  }
}

main();

#!/usr/bin/env node
// session-status.mjs — read-only status probe for a polish session.
//
// Skill: .claude/skills/polish-session/SKILL.md · Spec: specs/117-polish-session-skill
//
// Reports the two numbers a polish session must not guess at:
//
//   1. Binary freshness — whether this worktree's `promptworld` binary predates the newest
//      tracked Go source. `go build ./...` populates the build cache and NEVER rewrites the
//      `-o promptworld` artifact, so every build in a transcript can succeed while the binary
//      an operator is about to QA predates the changes by hours.
//   2. Wiki footprint — this branch's `wikiNotes` count and its headroom to the session
//      gate's `wiki-footprint` threshold. READ from `check-merge-drift.mjs session --json`,
//      never recomputed: a second derivation is a second thing to keep in sync.
//
// Node >= 18, ESM, zero npm dependencies (stdlib only).
// This script NEVER writes any file and never builds. A probe that fixed what it measures
// would make its own green result meaningless.
//
// Usage:
//   node .claude/skills/polish-session/scripts/session-status.mjs [--check] [--json]
//
// Exit codes:
//   0  binary is fresh (or deliberately absent-and-reported) and nothing needs attention
//   1  advisory: the binary is stale, or the footprint is at/over threshold
//   2  usage/environment error (not a git repo, bad flag, gate unreadable)
//
// Exit 1 is ADVISORY. Nothing in the harness consumes it; the session gate is non-blocking
// by design and making any of this block is a separate, spec'd decision.

import { existsSync, readFileSync, statSync } from 'node:fs';
import { join, resolve } from 'node:path';
import { execFileSync } from 'node:child_process';

const BINARY_NAME = 'promptworld';

// ---------------------------------------------------------------------------
// Argument parsing
// ---------------------------------------------------------------------------

const KNOWN_FLAGS = new Set(['--check', '--json']);

function parseArgs(argv) {
  const flags = { check: false, json: false };
  for (const arg of argv) {
    if (!KNOWN_FLAGS.has(arg)) {
      process.stderr.write(`session-status: unknown flag ${JSON.stringify(arg)}\n`);
      process.exit(2);
    }
    if (arg === '--check') flags.check = true;
    if (arg === '--json') flags.json = true;
  }
  return flags;
}

// ---------------------------------------------------------------------------
// Git helpers
// ---------------------------------------------------------------------------

function git(args, cwd) {
  return execFileSync('git', args, {
    cwd,
    encoding: 'utf8',
    stdio: ['ignore', 'pipe', 'pipe'],
    maxBuffer: 32 * 1024 * 1024,
  }).trim();
}

// The worktree we are standing in — NOT the shared root. A polish session works in
// .worktrees/task-<N>, and both readings below are about that checkout specifically.
function resolveWorktree() {
  try {
    return git(['rev-parse', '--show-toplevel'], process.cwd());
  } catch {
    process.stderr.write('session-status: not inside a git repository\n');
    process.exit(2);
  }
}

// The main checkout, where scripts/check-merge-drift.mjs is always present and where the
// gate wants to run. `--git-common-dir` points at <root>/.git from any linked worktree.
function resolveRoot(worktree) {
  try {
    const common = resolve(worktree, git(['rev-parse', '--git-common-dir'], worktree));
    // <root>/.git -> <root>; a non-worktree checkout returns the same dir.
    return common.endsWith('/.git') ? common.slice(0, -'/.git'.length) : worktree;
  } catch {
    return worktree;
  }
}

function currentBranch(worktree) {
  try {
    return git(['branch', '--show-current'], worktree) || null;
  } catch {
    return null;
  }
}

// ---------------------------------------------------------------------------
// Reading 1 — binary freshness
// ---------------------------------------------------------------------------

// Tracked Go sources only. An untracked scratch .go file is not something the binary is
// expected to contain, and vendored/ignored trees would swamp the comparison.
function newestGoSource(worktree) {
  let listing;
  try {
    listing = git(['ls-files', '-z', '*.go', 'go.mod', 'go.sum'], worktree);
  } catch {
    return null;
  }
  const files = listing.split('\0').filter(Boolean);
  let newest = null;
  for (const rel of files) {
    const abs = join(worktree, rel);
    let st;
    try {
      st = statSync(abs);
    } catch {
      continue; // deleted-but-tracked; not a source the binary could contain
    }
    if (newest === null || st.mtimeMs > newest.mtimeMs) {
      newest = { path: rel, mtimeMs: st.mtimeMs };
    }
  }
  return newest;
}

function checkBinary(worktree) {
  const binaryPath = join(worktree, BINARY_NAME);
  const newest = newestGoSource(worktree);

  if (!existsSync(binaryPath)) {
    return {
      verdict: 'absent',
      binary: BINARY_NAME,
      newestSource: newest ? newest.path : null,
      advice: `no ${BINARY_NAME} binary in this worktree — build it before live QA: ` +
        `go build -o ${BINARY_NAME} ./cmd/${BINARY_NAME}`,
    };
  }

  const binMtime = statSync(binaryPath).mtimeMs;
  if (newest === null) {
    return {
      verdict: 'unknown',
      binary: BINARY_NAME,
      newestSource: null,
      advice: 'no tracked Go sources found — cannot judge binary freshness',
    };
  }

  const ageSeconds = Math.round((binMtime - newest.mtimeMs) / 1000);
  if (binMtime >= newest.mtimeMs) {
    return {
      verdict: 'fresh',
      binary: BINARY_NAME,
      newestSource: newest.path,
      builtAfterNewestSourceBy: ageSeconds,
      advice: null,
    };
  }

  return {
    verdict: 'stale',
    binary: BINARY_NAME,
    newestSource: newest.path,
    behindBySeconds: -ageSeconds,
    advice:
      `${BINARY_NAME} predates ${newest.path} — \`go build ./...\` does not rewrite it. ` +
      `Rebuild before live QA: go build -o ${BINARY_NAME} ./cmd/${BINARY_NAME}`,
  };
}

// ---------------------------------------------------------------------------
// Reading 2 — wiki footprint (read from the session gate, never recomputed)
// ---------------------------------------------------------------------------

function readFootprint(root, worktree, branch) {
  const gate = join(root, 'scripts', 'check-merge-drift.mjs');
  if (!existsSync(gate)) {
    return { verdict: 'unavailable', reason: 'scripts/check-merge-drift.mjs not found' };
  }

  let raw;
  try {
    raw = execFileSync(process.execPath, [gate, 'session', '--json'], {
      cwd: root,
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'pipe'],
      maxBuffer: 64 * 1024 * 1024,
    });
  } catch (err) {
    // The gate exits nonzero on findings; its JSON still arrives on stdout.
    raw = err.stdout ? String(err.stdout) : '';
  }

  let report;
  try {
    report = JSON.parse(raw);
  } catch {
    return { verdict: 'unavailable', reason: 'session gate produced no parseable JSON' };
  }

  const branches = Array.isArray(report.branches) ? report.branches : [];
  // Match on the worktree path first — it is unambiguous even if two branches share a name
  // across repos — and fall back to the branch name for a non-worktree checkout.
  const mine =
    branches.find((b) => b.worktree && resolve(b.worktree) === resolve(worktree)) ||
    branches.find((b) => b.name === branch) ||
    null;

  if (!mine) {
    return {
      verdict: 'unlisted',
      reason: branch
        ? `the session gate does not list ${branch} (no commits beyond origin/main yet?)`
        : 'no current branch',
    };
  }

  // The threshold is the gate's constant, not ours. Recover it from the gate's own finding
  // when it fires; otherwise report the raw count without inventing a number.
  const finding = (report.findings || []).find(
    (f) => f.rule === 'wiki-footprint' && String(f.message || '').includes(mine.name)
  );
  const thresholdMatch = finding && String(finding.message).match(/threshold (\d+)/);
  const totalMatch = finding && String(finding.message).match(/of (\d+) wiki notes/);

  return {
    verdict: finding ? 'over-threshold' : 'ok',
    branch: mine.name,
    wikiNotes: mine.wikiFootprint,
    baseLag: mine.baseLag,
    threshold: thresholdMatch ? Number(thresholdMatch[1]) : null,
    totalNotes: totalMatch ? Number(totalMatch[1]) : null,
    gateMessage: finding ? finding.message : null,
  };
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

function main() {
  const flags = parseArgs(process.argv.slice(2));
  const worktree = resolveWorktree();
  const root = resolveRoot(worktree);
  const branch = currentBranch(worktree);

  const binary = checkBinary(worktree);
  const footprint = readFootprint(root, worktree, branch);

  const needsAttention = binary.verdict === 'stale' || footprint.verdict === 'over-threshold';

  if (flags.json) {
    process.stdout.write(
      `${JSON.stringify({ worktree, root, branch, binary, footprint, needsAttention }, null, 2)}\n`
    );
    process.exit(needsAttention ? 1 : 0);
  }

  process.stdout.write(`session-status: branch=${branch ?? '(detached)'} worktree=${worktree}\n`);

  // Binary
  process.stdout.write(`binary: ${binary.verdict}`);
  if (binary.verdict === 'fresh') {
    process.stdout.write(` (newer than ${binary.newestSource})\n`);
  } else {
    process.stdout.write('\n');
  }
  if (binary.advice) process.stdout.write(`  ${binary.advice}\n`);

  // Footprint
  if (footprint.verdict === 'unavailable' || footprint.verdict === 'unlisted') {
    process.stdout.write(`wiki footprint: ${footprint.verdict} — ${footprint.reason}\n`);
  } else {
    const total = footprint.totalNotes ? ` of ${footprint.totalNotes}` : '';
    process.stdout.write(`wiki footprint: wikiNotes=${footprint.wikiNotes}${total}\n`);
    if (footprint.gateMessage) {
      process.stdout.write(`  [warn] ${footprint.gateMessage}\n`);
    } else {
      process.stdout.write(
        '  below the session gate\'s threshold; read a rising number as scope sprawl, ' +
          'not as grounding overdue\n'
      );
    }
  }

  process.stdout.write(
    needsAttention
      ? 'verdict: needs attention before live QA (advisory — nothing blocks on this)\n'
      : 'verdict: ok\n'
  );
  process.exit(needsAttention ? 1 : 0);
}

main();

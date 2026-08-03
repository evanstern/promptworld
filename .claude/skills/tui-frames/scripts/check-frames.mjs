#!/usr/bin/env node
// check-frames.mjs — regenerate-and-diff guard for docs/design/tui/frames/.
//
// Spec: specs/113-tui-frames-skill/spec.md FR-008
//
// Node >= 18, ESM, zero npm dependencies (stdlib fs/path/os/child_process only).
// This script never writes into docs/design/tui/frames/ — it dumps a fresh
// generation into a temp directory and diffs it against the committed matrix,
// so a green result is evidence rather than a self-fulfilling side effect.
// The working tree is left exactly as it was found.
//
// Usage:
//   node check-frames.mjs [--check]
//
// Exit codes:
//   0  the committed matrix is byte-identical to a fresh --dump
//   1  at least one file differs, is missing, or is extra (listed)
//   2  usage/environment error (not a git repo, go run failed, bad flag)

import { mkdtempSync, readdirSync, readFileSync, rmSync, existsSync } from 'node:fs';
import { join } from 'node:path';
import { tmpdir } from 'node:os';
import { execFileSync } from 'node:child_process';

const KNOWN_FLAGS = new Set(['--check']);

function parseArgs(argv) {
  for (const arg of argv) {
    if (!KNOWN_FLAGS.has(arg)) {
      process.stderr.write(`check-frames: unknown flag ${JSON.stringify(arg)}\n`);
      process.exit(2);
    }
  }
}

function resolveRepoRoot() {
  try {
    const out = execFileSync('git', ['rev-parse', '--show-toplevel'], {
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'pipe'],
    });
    return out.trim();
  } catch {
    process.stderr.write('check-frames: not inside a git repository\n');
    process.exit(2);
  }
}

const IGNORE = new Set(['README.md']);

function listGeneratedFiles(dir) {
  if (!existsSync(dir)) return new Set();
  return new Set(readdirSync(dir).filter((name) => !IGNORE.has(name)));
}

function main() {
  parseArgs(process.argv.slice(2));
  const repoRoot = resolveRepoRoot();
  const committedDir = join(repoRoot, 'docs/design/tui/frames');

  let tmpDir;
  try {
    tmpDir = mkdtempSync(join(tmpdir(), 'tui-frames-check-'));
  } catch (err) {
    process.stderr.write(`check-frames: could not create temp dir: ${err.message}\n`);
    process.exit(2);
  }

  try {
    try {
      execFileSync(
        'go',
        ['run', './cmd/promptworld', 'frames', '--dump', '--out', tmpDir],
        { cwd: repoRoot, stdio: ['ignore', 'pipe', 'pipe'] }
      );
    } catch (err) {
      process.stderr.write('check-frames: go run ./cmd/promptworld frames --dump failed\n');
      if (err.stdout) process.stderr.write(err.stdout.toString());
      if (err.stderr) process.stderr.write(err.stderr.toString());
      process.exit(2);
    }

    const committed = listGeneratedFiles(committedDir);
    const generated = listGeneratedFiles(tmpDir);

    const missing = [...generated].filter((f) => !committed.has(f)).sort(); // in fresh dump, not committed
    const extra = [...committed].filter((f) => !generated.has(f)).sort(); // committed, not in fresh dump
    const differing = [];

    for (const name of [...committed].filter((f) => generated.has(f)).sort()) {
      const committedBytes = readFileSync(join(committedDir, name));
      const generatedBytes = readFileSync(join(tmpDir, name));
      if (!committedBytes.equals(generatedBytes)) differing.push(name);
    }

    if (missing.length === 0 && extra.length === 0 && differing.length === 0) {
      process.stdout.write('tui-frames: committed matrix matches a fresh --dump\n');
      process.exit(0);
    }

    for (const name of differing) {
      process.stdout.write(`differing  ${name}\n`);
    }
    for (const name of missing) {
      process.stdout.write(`missing    ${name}  (in a fresh --dump, not committed)\n`);
    }
    for (const name of extra) {
      process.stdout.write(`extra      ${name}  (committed, absent from a fresh --dump)\n`);
    }
    process.stdout.write(
      `${differing.length} differing, ${missing.length} missing, ${extra.length} extra\n`
    );
    process.exit(1);
  } finally {
    rmSync(tmpDir, { recursive: true, force: true });
  }
}

main();

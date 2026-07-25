#!/usr/bin/env node
// check-merge-drift.test.mjs — regression coverage for the player-docs
// evidence-shape -> fingerprint path (TASK-138).
//
// check-freshness.mjs reports `sources` as OBJECTS ({path, recorded,
// current, fresh}), not strings. The player-docs extractor in
// check-merge-drift.mjs must map each source to its `.path` before it
// becomes "touched" evidence, because that evidence flows straight into
// fingerprint() ([...evidence].sort().join(',')) and then into the
// board-note dedup (`existingText.includes(f.fingerprint)`). If the
// extractor ever again yields objects, every player-docs staleness finding
// silently fingerprints alike and the first note written ever suppresses
// all future ones — this is invisible to exit-code testing, hence this
// test.
//
// This is the first node:test harness for scripts/*.mjs in this repo (none
// existed to follow a convention from — see TASK-138 discussion). Run with:
//   node --test scripts/check-merge-drift.test.mjs

import { test } from 'node:test';
import assert from 'node:assert/strict';
import { mkdtempSync, mkdirSync, writeFileSync, rmSync, symlinkSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';
import { execFileSync } from 'node:child_process';
import { computePlayerDocsSurface, fingerprint } from './check-merge-drift.mjs';

const SCRIPT_PATH = join(dirname(fileURLToPath(import.meta.url)), 'check-merge-drift.mjs');

// Builds a throwaway "main worktree" directory containing a stub
// .claude/skills/player-docs/scripts/check-freshness.mjs that ignores its
// argv and always prints the given parsed report as --json output, matching
// the real script's shape (parsed.pages[].sources[] = {path, recorded,
// current, fresh}).
function makeFixture(report) {
  const dir = mkdtempSync(join(tmpdir(), 'merge-drift-test-'));
  const scriptDir = join(dir, '.claude', 'skills', 'player-docs', 'scripts');
  mkdirSync(scriptDir, { recursive: true });
  const scriptPath = join(scriptDir, 'check-freshness.mjs');
  writeFileSync(scriptPath, `#!/usr/bin/env node\nprocess.stdout.write(${JSON.stringify(JSON.stringify(report))});\n`);
  return dir;
}

test('computePlayerDocsSurface extracts source paths (strings), not source objects', () => {
  const dir = makeFixture({
    pages: [
      {
        page: 'getting-started.html',
        verdict: 'stale',
        sources: [
          { path: 'README.md', recorded: 'aaa', current: 'bbb', fresh: false },
          { path: 'docs/wiki/cli-promptworld.md', recorded: 'ccc', current: 'ddd', fresh: false },
        ],
      },
      {
        page: 'fresh-page.html',
        verdict: 'fresh',
        sources: [{ path: 'docs/wiki/should-not-appear.md', recorded: 'x', current: 'x', fresh: true }],
      },
    ],
  });
  try {
    const surface = computePlayerDocsSurface(dir);
    // The load-bearing shape assertion: a future extractor regression that
    // returns objects instead of `.path` strings must fail this loudly.
    for (const t of surface.touched) {
      assert.equal(typeof t, 'string', `touched entry should be a string path, got ${JSON.stringify(t)}`);
    }
    assert.deepEqual(
      [...surface.touched].sort(),
      ['README.md', 'docs/wiki/cli-promptworld.md'].sort()
    );
    // Only the stale page's sources are evidence; the fresh page's source
    // must not leak in.
    assert.ok(!surface.touched.includes('docs/wiki/should-not-appear.md'));
  } finally {
    rmSync(dir, { recursive: true, force: true });
  }
});

test('computePlayerDocsSurface dedups a source repeated across stale pages', () => {
  const dir = makeFixture({
    pages: [
      {
        page: 'getting-started.html',
        verdict: 'stale',
        sources: [{ path: 'README.md', recorded: 'aaa', current: 'bbb', fresh: false }],
      },
      {
        page: 'daemon.html',
        verdict: 'stale',
        sources: [
          { path: 'README.md', recorded: 'aaa', current: 'bbb', fresh: false }, // same source, second stale page
          { path: 'docs/wiki/daemon-lifecycle.md', recorded: 'eee', current: 'fff', fresh: false },
        ],
      },
    ],
  });
  try {
    const surface = computePlayerDocsSurface(dir);
    // Set-based dedup only works when evidence is comparable-by-value
    // (strings); with object evidence every entry is a distinct identity
    // and nothing collapses.
    assert.equal(surface.touched.filter((t) => t === 'README.md').length, 1);
    assert.equal(surface.touched.length, 2);
  } finally {
    rmSync(dir, { recursive: true, force: true });
  }
});

test('two different player-docs staleness findings produce two different fingerprints', () => {
  const dirA = makeFixture({
    pages: [
      {
        page: 'getting-started.html',
        verdict: 'stale',
        sources: [{ path: 'docs/wiki/gru.md', recorded: 'aaa', current: 'bbb', fresh: false }],
      },
    ],
  });
  const dirB = makeFixture({
    pages: [
      {
        page: 'other.html',
        verdict: 'stale',
        sources: [{ path: 'docs/wiki/morgue.md', recorded: 'ccc', current: 'ddd', fresh: false }],
      },
    ],
  });
  try {
    const surfaceA = computePlayerDocsSurface(dirA);
    const surfaceB = computePlayerDocsSurface(dirB);
    // Same gate/rule/key as the real finding built in runSession() — only
    // the evidence (touched sources) differs between the two findings.
    const fpA = fingerprint('session', 'grounding-stale', 'player-docs', surfaceA.touched);
    const fpB = fingerprint('session', 'grounding-stale', 'player-docs', surfaceB.touched);
    assert.notEqual(
      fpA,
      fpB,
      'two distinct player-docs staleness findings must not collapse to the same fingerprint (TASK-138)'
    );
  } finally {
    rmSync(dirA, { recursive: true, force: true });
    rmSync(dirB, { recursive: true, force: true });
  }
});

test('the CLI still runs main() when invoked through a symlink', () => {
  // Regression for a second bug found in review: an entry-point guard of the
  // form `process.argv[1] === new URL(import.meta.url).pathname` silently
  // skips main() whenever the script is reached through a symlinked path —
  // import.meta.url resolves symlinks, argv[1] does not — so `node
  // <symlink-to-script> session` would exit 0 with NO output: a gate that
  // silently passes everything. The fix compares realpathSync() of both
  // sides and defaults to running when it can't tell. This test spawns the
  // real CLI through a freshly created symlink and requires real output.
  const dir = mkdtempSync(join(tmpdir(), 'merge-drift-symlink-test-'));
  const linkPath = join(dir, 'check-merge-drift-via-symlink.mjs');
  try {
    symlinkSync(SCRIPT_PATH, linkPath);
    const out = execFileSync(process.execPath, [linkPath, 'session', '--no-fetch'], {
      cwd: process.cwd(),
      encoding: 'utf8',
    });
    assert.ok(out.length > 0, 'expected non-empty stdout when invoked through a symlink');
    assert.match(out, /^check-merge-drift: mode=session/);
  } finally {
    rmSync(dir, { recursive: true, force: true });
  }
});

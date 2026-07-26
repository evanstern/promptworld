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

import { test, after } from 'node:test';
import assert from 'node:assert/strict';
import { existsSync, mkdtempSync, mkdirSync, writeFileSync, rmSync, symlinkSync } from 'node:fs';
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

// ---------------------------------------------------------------------------
// spec 069 — wiki grounding rides the PR (pin-vs-branch predicate + in-branch
// player-docs block). Fixture pattern follows claim-protocol.test.mjs: a bare
// origin + one clone in a tmpdir, isolated git config, the real CLI spawned
// as a child process, assertions on exit code + finding rules. The clone
// stays on main (pr mode's root-not-main check) and every scenario is its own
// task branch gated via `pr --branch <name>`.
// ---------------------------------------------------------------------------

const root69 = mkdtempSync(join(tmpdir(), 'wiki-in-pr-gate-test-'));
after(() => rmSync(root69, { recursive: true, force: true }));

const gitConfig69 = join(root69, 'gitconfig');
writeFileSync(
  gitConfig69,
  '[user]\n\tname = Fixture\n\temail = fixture@test.invalid\n' +
    '[init]\n\tdefaultBranch = main\n[commit]\n\tgpgsign = false\n'
);
const ENV69 = {
  ...process.env,
  GIT_CONFIG_GLOBAL: gitConfig69,
  GIT_CONFIG_SYSTEM: process.platform === 'win32' ? 'NUL' : '/dev/null',
};

function run69(cmd, args, { cwd, env = ENV69 } = {}) {
  try {
    const stdout = execFileSync(cmd, args, { cwd, env, encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe'] });
    return { status: 0, stdout, stderr: '' };
  } catch (err) {
    return {
      status: typeof err.status === 'number' ? err.status : 1,
      stdout: err.stdout ? err.stdout.toString() : '',
      stderr: err.stderr ? err.stderr.toString() : String(err.message || err),
    };
  }
}

const gitIn69 = (cwd, ...args) => run69('git', args, { cwd });

function write69(base, rel, content) {
  const abs = join(base, rel);
  mkdirSync(dirname(abs), { recursive: true });
  writeFileSync(abs, content);
}

// Stub player-docs checkers (research R2's env-var seam): US1 scenarios that
// touch docs/wiki/ get the exit-0 stub so the wiki predicate is what's under
// test; US2 tests swap in the exit-1/exit-2/sentinel stubs.
const stubExit0 = join(root69, 'stub-checker-exit0.mjs');
const stubExit1 = join(root69, 'stub-checker-exit1.mjs');
const stubExit2 = join(root69, 'stub-checker-exit2.mjs');
const sentinelPath = join(root69, 'checker-was-invoked.sentinel');
const stubSentinel = join(root69, 'stub-checker-sentinel.mjs');
writeFileSync(stubExit0, 'process.exit(0);\n');
writeFileSync(stubExit1, 'process.exit(1);\n');
writeFileSync(stubExit2, 'process.exit(2);\n');
writeFileSync(
  stubSentinel,
  `import { writeFileSync } from 'node:fs';\nwriteFileSync(${JSON.stringify(sentinelPath)}, 'invoked');\nprocess.exit(0);\n`
);

// pr mode gate runner; checker defaults to the exit-0 stub so fixture repos
// (which have no .claude/skills/ tree) never trip player-docs-env-error
// except when a test asks for it.
function gate69(cwd, branch, { checker = stubExit0 } = {}) {
  return run69(process.execPath, [SCRIPT_PATH, 'pr', '--branch', branch, '--json'], {
    cwd,
    env: { ...ENV69, CHECK_MERGE_DRIFT_PLAYER_DOCS_CHECKER: checker },
  });
}

// Seed: sources first (their commit hash becomes the note's pin), then the
// note pinned to that hash, plus a pre-existing malformed note that no
// scenario's branch overlaps (research R5: it must never brick a PR).
const origin69 = join(root69, 'origin.git');
run69('git', ['init', '--bare', '-b', 'main', origin69]);
const clone69 = join(root69, 'clone');
run69('git', ['clone', origin69, clone69]);
write69(clone69, 'internal/sim/landing.go', 'package sim // v1\n');
write69(clone69, 'README.md', '# fixture\n');
gitIn69(clone69, 'add', '-A');
gitIn69(clone69, 'commit', '-m', 'seed: sources');
const seedPin = gitIn69(clone69, 'rev-parse', 'HEAD').stdout.trim();

function noteBody(pin) {
  return `---\nverified_against: ${pin}\nsources:\n  - internal/sim/landing.go\n---\n\n# Landing\n`;
}
const NOTE = 'docs/wiki/landing.md';
write69(clone69, NOTE, noteBody(seedPin));
write69(clone69, 'docs/wiki/malformed-preexisting.md', '# no frontmatter at all\n');
gitIn69(clone69, 'add', '-A');
gitIn69(clone69, 'commit', '-m', 'seed: wiki notes');
gitIn69(clone69, 'push', 'origin', 'main');

// Make a scenario branch off origin/main, run `mutate` (which commits), and
// return to main so pr mode's root check stays satisfied.
function scenarioBranch69(name, mutate) {
  gitIn69(clone69, 'switch', '-c', name, 'origin/main');
  try {
    mutate();
  } finally {
    gitIn69(clone69, 'switch', 'main');
  }
}

function commitAll69(msg) {
  gitIn69(clone69, 'add', '-A');
  gitIn69(clone69, 'commit', '-m', msg);
  return gitIn69(clone69, 'rev-parse', 'HEAD').stdout.trim();
}

function parseGate69(r) {
  let report;
  try {
    report = JSON.parse(r.stdout);
  } catch {
    assert.fail(`gate output was not JSON (exit ${r.status}): ${r.stdout}\n${r.stderr}`);
  }
  return report;
}

test('069/US1: branch touches a pinned source with no note change -> wiki-repin-missing blocks', () => {
  scenarioBranch69('task-701-no-repin', () => {
    write69(clone69, 'internal/sim/landing.go', 'package sim // v2\n');
    commitAll69('touch pinned source, no re-pin');
  });
  const r = gate69(clone69, 'task-701-no-repin');
  assert.equal(r.status, 1, `expected exit 1: ${r.stdout}${r.stderr}`);
  const report = parseGate69(r);
  assert.equal(report.verdict, 'blocked');
  const f = report.findings.find((x) => x.rule === 'wiki-repin-missing');
  assert.ok(f, 'expected a wiki-repin-missing finding');
  assert.equal(f.severity, 'block');
  assert.match(f.message, /docs\/wiki\/landing\.md/, 'must name the note');
  assert.match(f.message, /internal\/sim\/landing\.go/, 'must name the matched source');
  assert.match(f.message, /not modified on the branch/, 'must name the failed clause');
  assert.match(f.message, /grounding-wiki:wiki-update/, 'must name the remedy');
  // FR-002: the old advisory is replaced, not duplicated.
  assert.equal(report.findings.find((x) => x.rule === 'wiki-sources-overlap'), undefined);
});

test('069/US1: note re-pinned to a branch commit covering the source change -> pass, no finding', () => {
  scenarioBranch69('task-702-repinned', () => {
    write69(clone69, 'internal/sim/landing.go', 'package sim // v2\n');
    const srcCommit = commitAll69('touch pinned source');
    write69(clone69, NOTE, noteBody(srcCommit));
    commitAll69('re-pin landing note to the branch commit');
  });
  const r = gate69(clone69, 'task-702-repinned');
  assert.equal(r.status, 0, `satisfied re-pin must pass: ${r.stdout}${r.stderr}`);
  const report = parseGate69(r);
  for (const rule of ['wiki-repin-missing', 'wiki-sources-overlap', 'wiki-note-malformed', 'player-docs-stale', 'player-docs-env-error']) {
    assert.equal(report.findings.find((x) => x.rule === rule), undefined, `no ${rule} finding expected`);
  }
});

test('069/US1: source re-touched after the re-pin -> blocks (pin did not see every change)', () => {
  scenarioBranch69('task-703-retouch', () => {
    write69(clone69, 'internal/sim/landing.go', 'package sim // v2\n');
    const srcCommit = commitAll69('touch pinned source');
    write69(clone69, NOTE, noteBody(srcCommit));
    commitAll69('re-pin landing note');
    write69(clone69, 'internal/sim/landing.go', 'package sim // v3 after the pin\n');
    commitAll69('touch the source again AFTER the pin');
  });
  const r = gate69(clone69, 'task-703-retouch');
  assert.equal(r.status, 1, `re-touched source must block: ${r.stdout}${r.stderr}`);
  const f = parseGate69(r).findings.find((x) => x.rule === 'wiki-repin-missing');
  assert.ok(f, 'expected a wiki-repin-missing finding');
  assert.match(f.message, /changed after the verified_against pin/);
});

test('069/US1: re-pinned note whose pin is not reachable from the branch tip -> blocks', () => {
  // A real-but-foreign commit (e.g. a pin left pointing at another lineage
  // after a rebase): reachability, not mere existence, is the clause.
  gitIn69(clone69, 'switch', '-c', 'side-pin-source', 'origin/main');
  write69(clone69, 'side-file.txt', 'side lineage\n');
  const sideCommit = commitAll69('side lineage commit');
  gitIn69(clone69, 'switch', 'main');
  scenarioBranch69('task-704-unreachable', () => {
    write69(clone69, 'internal/sim/landing.go', 'package sim // v2\n');
    commitAll69('touch pinned source');
    write69(clone69, NOTE, noteBody(sideCommit));
    commitAll69('re-pin landing note to an unreachable commit');
  });
  const r = gate69(clone69, 'task-704-unreachable');
  assert.equal(r.status, 1, `unreachable pin must block: ${r.stdout}${r.stderr}`);
  const f = parseGate69(r).findings.find((x) => x.rule === 'wiki-repin-missing');
  assert.ok(f, 'expected a wiki-repin-missing finding');
  assert.match(f.message, /not reachable from the branch tip/);
});

test('069/US1: branch touching no pinned sources -> unchanged behavior, no wiki findings', () => {
  scenarioBranch69('task-705-unrelated', () => {
    write69(clone69, 'README.md', '# fixture, updated\n');
    commitAll69('unrelated change');
  });
  const r = gate69(clone69, 'task-705-unrelated');
  assert.equal(r.status, 0, `no-overlap branch must pass: ${r.stdout}${r.stderr}`);
  const report = parseGate69(r);
  for (const rule of ['wiki-repin-missing', 'wiki-sources-overlap', 'wiki-note-malformed']) {
    assert.equal(report.findings.find((x) => x.rule === rule), undefined, `no ${rule} finding expected`);
  }
  // research R5: the pre-existing malformed note the branch never overlaps
  // must not surface in pr mode at all.
});

test('069/US1 edge: branch deletes the note while touching its sources -> structural re-verification, passes', () => {
  scenarioBranch69('task-706-delete-note', () => {
    write69(clone69, 'internal/sim/landing.go', 'package sim // v2\n');
    gitIn69(clone69, 'rm', NOTE);
    commitAll69('restructure: source changed, note deleted');
  });
  const r = gate69(clone69, 'task-706-delete-note');
  assert.equal(r.status, 0, `deleted-note branch must pass: ${r.stdout}${r.stderr}`);
  const report = parseGate69(r);
  assert.equal(report.findings.find((x) => x.rule === 'wiki-repin-missing'), undefined);
  assert.equal(report.findings.find((x) => x.rule === 'wiki-note-malformed'), undefined);
});

test('069/US1: predicate-needed note malformed at the branch tip -> wiki-note-malformed blocks', () => {
  scenarioBranch69('task-707-malformed', () => {
    write69(clone69, 'internal/sim/landing.go', 'package sim // v2\n');
    write69(clone69, NOTE, '---\nsources:\n  - internal/sim/landing.go\n---\n\n# Landing (pin lost)\n');
    commitAll69('touch source, break the note frontmatter');
  });
  const r = gate69(clone69, 'task-707-malformed');
  assert.equal(r.status, 1, `malformed predicate-needed note must block: ${r.stdout}${r.stderr}`);
  const f = parseGate69(r).findings.find((x) => x.rule === 'wiki-note-malformed');
  assert.ok(f, 'expected a wiki-note-malformed finding');
  assert.equal(f.severity, 'block');
  assert.match(f.message, /branch tip/);
  assert.match(f.message, /grounding-wiki:wiki-update/);
});

test('069/US2: wiki-touching branch + checker exit 1 -> player-docs-stale blocks', () => {
  scenarioBranch69('task-708-wiki-only', () => {
    write69(clone69, NOTE, noteBody(seedPin) + '\nEdited prose (sources untouched).\n');
    commitAll69('wiki-only edit');
  });
  const r = gate69(clone69, 'task-708-wiki-only', { checker: stubExit1 });
  assert.equal(r.status, 1, `stale player docs must block: ${r.stdout}${r.stderr}`);
  const f = parseGate69(r).findings.find((x) => x.rule === 'player-docs-stale');
  assert.ok(f, 'expected a player-docs-stale finding');
  assert.equal(f.severity, 'block');
  assert.match(f.message, /player-docs skill/);
});

test('069/US2: wiki-touching branch + checker exit 2 -> player-docs-env-error blocks (no silent pass)', () => {
  const r = gate69(clone69, 'task-708-wiki-only', { checker: stubExit2 });
  assert.equal(r.status, 1, `checker env error must block: ${r.stdout}${r.stderr}`);
  const f = parseGate69(r).findings.find((x) => x.rule === 'player-docs-env-error');
  assert.ok(f, 'expected a player-docs-env-error finding');
  assert.equal(f.severity, 'block');
  assert.match(f.message, /exit 2/);
});

test('069/US2: wiki-touching branch + checker exit 0 -> no player-docs finding', () => {
  const r = gate69(clone69, 'task-708-wiki-only', { checker: stubExit0 });
  assert.equal(r.status, 0, `fresh player docs must pass: ${r.stdout}${r.stderr}`);
  const report = parseGate69(r);
  assert.equal(report.findings.find((x) => x.rule === 'player-docs-stale'), undefined);
  assert.equal(report.findings.find((x) => x.rule === 'player-docs-env-error'), undefined);
});

test('069/US2: branch touching no docs/wiki file -> the checker is not invoked at all', () => {
  rmSync(sentinelPath, { force: true });
  const r = gate69(clone69, 'task-705-unrelated', { checker: stubSentinel });
  assert.equal(r.status, 0);
  assert.ok(!existsSync(sentinelPath), 'checker must not be spawned when no docs/wiki/ file changed');
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

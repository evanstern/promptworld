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

// ---------------------------------------------------------------------------
// spec 080 — paused-label lanes downgrade to info (praxisflux TASK-55
// counterpart). Same fixture pattern as the 069 section: a bare origin + one
// clone in a tmpdir, isolated git config (ENV69), the real CLI spawned as a
// child process, assertions on exit code + finding rules. Cards live on
// origin/main AND in the clone's working tree — isTaskPaused reads the main
// worktree's working tree, the board of record for a local operator pause.
// ---------------------------------------------------------------------------

const root80 = mkdtempSync(join(tmpdir(), 'paused-lanes-test-'));
after(() => rmSync(root80, { recursive: true, force: true }));

const origin80 = join(root80, 'origin.git');
run69('git', ['init', '--bare', '-b', 'main', origin80]);
const clone80 = join(root80, 'clone');
run69('git', ['clone', origin80, clone80]);
const gitIn80 = (...args) => gitIn69(clone80, ...args);

function commitAll80(msg) {
  gitIn80('add', '-A');
  gitIn80('commit', '-m', msg);
  return gitIn80('rev-parse', 'HEAD').stdout.trim();
}

function card80(id, labelsYaml, extra = '') {
  return `---\nid: ${id}\ntitle: fixture ${id}\nstatus: In Progress\n${labelsYaml}\n---\n\n## Description\n\nFixture card.${extra}\n`;
}

// Seed: conflict target, a wiki-pinned source, four cards (two paused — one
// block-list labels, one inline-flow labels — two live controls), and a spec
// dir claimed by the paused TASK-810 for the ownership-protection tests.
write69(clone80, 'conflict.txt', 'base\n');
write69(clone80, 'internal/pw/x.go', 'package pw // v1\n');
write69(
  clone80,
  'backlog/tasks/task-810 - paused-fixture.md',
  card80('TASK-810', 'labels:\n  - pdlc\n  - paused', ' Spec: specs/090-paused-claim')
);
write69(clone80, 'backlog/tasks/task-811 - live-fixture.md', card80('TASK-811', 'labels:\n  - pdlc'));
write69(clone80, 'backlog/tasks/task-812 - paused-inline-fixture.md', card80('TASK-812', 'labels: [paused]'));
write69(clone80, 'backlog/tasks/task-813 - live-cleanup-fixture.md', card80('TASK-813', 'labels: []'));
write69(clone80, 'specs/090-paused-claim/CLAIM.md', '# Spec 090 — claimed by TASK-810 (fixture)\n');
const seed80 = commitAll80('seed: fixture cards, sources, claimed spec dir');
write69(clone80, 'docs/wiki/pw.md', `---\nverified_against: ${seed80}\nsources:\n  - internal/pw/x.go\n---\n\n# PW\n`);
commitAll80('seed: wiki note pinned');
gitIn80('push', 'origin', 'main');

function branch80(name, mutate) {
  gitIn80('switch', '-c', name, 'main');
  try {
    mutate();
  } finally {
    gitIn80('switch', 'main');
  }
}

// Task branches fork BEFORE main advances, so both conflict with origin/main
// on conflict.txt (and with each other) — deliberately unpushed, so session
// mode also exercises branch-unpushed on both.
branch80('task-810-paused-work', () => {
  write69(clone80, 'conflict.txt', 'paused side\n');
  commitAll80('paused-side edit');
});
branch80('task-811-live-work', () => {
  write69(clone80, 'conflict.txt', 'live side\n');
  commitAll80('live-side edit');
});
branch80('task-810-wiki-touch', () => {
  write69(clone80, 'internal/pw/x.go', 'package pw // v2\n');
  commitAll80('touch pinned source, no re-pin');
});
write69(clone80, 'conflict.txt', 'main side\n');
commitAll80('main advance: conflict.txt');
gitIn80('push', 'origin', 'main');

// Merged-at-tip worktrees (ancestor => cleanup-eligible): one paused, one live.
gitIn80('worktree', 'add', join(clone80, '.worktrees', 'task-812'), '-b', 'task-812-parked', 'origin/main');
gitIn80('worktree', 'add', join(clone80, '.worktrees', 'task-813'), '-b', 'task-813-parked', 'origin/main');

function gate80(mode, args = []) {
  return run69(process.execPath, [SCRIPT_PATH, mode, '--json', ...args], { cwd: clone80 });
}

function findBy80(report, rule, evidenceHas) {
  return report.findings.find(
    (f) => f.rule === rule && evidenceHas.every((e) => f.evidence.includes(e))
  );
}

test('080/session: a paused task branch downgrades drift findings to info, pause as evidence; live control stays warn', () => {
  const r = gate80('session');
  assert.equal(r.status, 0, `session must pass: ${r.stdout}${r.stderr}`);
  const report = parseGate69(r);

  const pausedConflict = findBy80(report, 'textual-conflict', ['task-810-paused-work']);
  assert.ok(pausedConflict, 'expected a textual-conflict finding for the paused branch');
  assert.equal(pausedConflict.severity, 'info', 'paused branch conflict must be info');
  assert.ok(pausedConflict.evidence.includes('paused:TASK-810'), 'pause must be cited as evidence');
  assert.match(pausedConflict.message, /TASK-810 is paused/);

  const liveConflict = findBy80(report, 'textual-conflict', ['task-811-live-work']);
  assert.ok(liveConflict, 'expected a textual-conflict finding for the live branch');
  assert.equal(liveConflict.severity, 'warn', 'live branch conflict must stay warn');
  assert.ok(!liveConflict.evidence.some((e) => e.startsWith('paused:')));

  const pausedPair = findBy80(report, 'pairwise-conflict', ['task-810-paused-work', 'task-811-live-work']);
  assert.ok(pausedPair, 'expected the paused x live pairwise finding');
  assert.equal(pausedPair.severity, 'info', 'either side paused downgrades the pair');
  assert.ok(pausedPair.evidence.includes('paused:TASK-810'));

  const livePair = findBy80(report, 'pairwise-conflict', ['task-811-live-work', 'task-813-parked']);
  assert.ok(livePair, 'expected the live x live pairwise finding');
  assert.equal(livePair.severity, 'warn', 'a fully-live pair must stay warn');

  const pausedUnpushed = findBy80(report, 'branch-unpushed', ['task-810-paused-work']);
  assert.ok(pausedUnpushed && pausedUnpushed.severity === 'info', 'paused unpushed branch downgrades');
  const liveUnpushed = findBy80(report, 'branch-unpushed', ['task-811-live-work']);
  assert.ok(liveUnpushed && liveUnpushed.severity === 'warn', 'live unpushed branch stays warn');
});

test('080/session: cleanup is never prescribed for a paused branch/worktree (inline labels form detected)', () => {
  const r = gate80('session');
  assert.equal(r.status, 0);
  const report = parseGate69(r);

  const pausedCleanup = findBy80(report, 'cleanup-eligible', ['task-812-parked']);
  assert.ok(pausedCleanup, 'expected a cleanup-eligible finding for the paused parked branch');
  assert.equal(pausedCleanup.severity, 'info');
  assert.match(pausedCleanup.message, /cleanup not prescribed/);
  assert.ok(pausedCleanup.evidence.includes('paused:TASK-812'), 'inline labels: [paused] must be detected');

  const liveCleanup = findBy80(report, 'cleanup-eligible', ['task-813-parked']);
  assert.ok(liveCleanup && liveCleanup.severity === 'warn', 'live parked branch keeps the prescription warn');

  const prescribed = report.cleanupPrescriptions.map((p) => p.branch);
  assert.ok(!prescribed.includes('task-812-parked'), '--apply-cleanup must never see the paused branch');
  assert.ok(prescribed.includes('task-813-parked'), 'the live merged branch is still prescribed');
});

test('080/pr: gating a paused task branch — blocking conflict downgrades to info, gate exits 0', () => {
  const r = gate80('pr', ['--branch', 'task-810-paused-work']);
  assert.equal(r.status, 0, `paused branch pr gate must pass: ${r.stdout}${r.stderr}`);
  const report = parseGate69(r);
  assert.equal(report.verdict === 'blocked', false);
  const conflict = report.findings.find((f) => f.rule === 'textual-conflict');
  assert.ok(conflict, 'the conflict is still reported');
  assert.equal(conflict.severity, 'info');
  assert.ok(conflict.evidence.includes('paused:TASK-810'));
  const stale = report.findings.find((f) => f.rule === 'stale-base');
  assert.ok(stale && stale.severity === 'info', 'stale-base downgrades too');
});

test('080/pr: the live control branch still blocks on the same conflict', () => {
  const r = gate80('pr', ['--branch', 'task-811-live-work']);
  assert.equal(r.status, 1, `live branch pr gate must block: ${r.stdout}${r.stderr}`);
  const f = parseGate69(r).findings.find((x) => x.rule === 'textual-conflict');
  assert.ok(f && f.severity === 'block');
});

test('080/pr: pausing is not a spec-069 bypass — wiki-repin-missing still blocks a paused task branch', () => {
  const r = gate80('pr', ['--branch', 'task-810-wiki-touch']);
  assert.equal(r.status, 1, `spec-069 block must survive the pause: ${r.stdout}${r.stderr}`);
  const f = parseGate69(r).findings.find((x) => x.rule === 'wiki-repin-missing');
  assert.ok(f, 'expected a wiki-repin-missing finding');
  assert.equal(f.severity, 'block');
});

test('080/worktree+claim: ownership protections stay blocking when the owner is paused', () => {
  const w = gate80('worktree', ['--spec', '090', '--task', 'TASK-999']);
  assert.equal(w.status, 1, `worktree spec collision must still block: ${w.stdout}${w.stderr}`);
  const wf = parseGate69(w).findings.find((x) => x.rule === 'spec-number-collision');
  assert.ok(wf && wf.severity === 'block', 'a paused owner still protects its claimed number');

  const c = gate80('claim', ['--dir', '090-stolen-name']);
  assert.equal(c.status, 1, `claim collision must still block: ${c.stdout}${c.stderr}`);
  const cf = parseGate69(c).findings.find((x) => x.rule === 'spec-number-collision');
  assert.ok(cf && cf.severity === 'block');
});

// ---------------------------------------------------------------------------
// spec 088 (TASK-162) — pr-mode docs-stale probe fires on all pinned
// sources + after history moves. Reuses the 069 section's clone69/origin69
// fixture (bare origin + one clone, isolated git config, gate69 spawning the
// real CLI). Ordering is load-bearing: node:test runs this file's top-level
// tests sequentially, so the "088 setup" test below — which pushes new
// docs/player/ pages to origin/main — runs AFTER every 069-section branch
// has already forked (e.g. task-705-unrelated, whose README.md-only touch
// must still NOT trigger the probe per SC-004, since its tip predates these
// pages). Every 088-section branch forks from origin/main AFTER the setup
// commit, so it inherits the widened trigger surface.
// ---------------------------------------------------------------------------

const tuiStubExit0 = join(root69, 'tui-stub-exit0.mjs');
const tuiStubExit1 = join(root69, 'tui-stub-exit1.mjs');
writeFileSync(tuiStubExit0, '#!/usr/bin/env node\nprocess.stdout.write(JSON.stringify({ ok: true, checks: [] }));\nprocess.exit(0);\n');
writeFileSync(
  tuiStubExit1,
  "#!/usr/bin/env node\nprocess.stdout.write(JSON.stringify({ ok: false, checks: [{ check: 'same-pr', file: 'docs/design/tui/patterns/example.md', message: 'stale' }] }));\nprocess.exit(1);\n"
);

// gate69 only overrides the player-docs checker; spec 088 also needs the
// tui-design env-override seam (research D3), so this variant sets both,
// defaulting the player-docs side to the fresh stub (stubExit0) so it never
// interferes with a tui-design-only fixture.
function gate69Full(cwd, branch, { playerChecker = stubExit0, tuiChecker } = {}) {
  const env = { ...ENV69, CHECK_MERGE_DRIFT_PLAYER_DOCS_CHECKER: playerChecker };
  if (tuiChecker) env.CHECK_MERGE_DRIFT_TUI_DESIGN_CHECKER = tuiChecker;
  return run69(process.execPath, [SCRIPT_PATH, 'pr', '--branch', branch, '--json'], { cwd, env });
}

test('088 setup: seed docs/player/ pages declaring README.md + docs/llm-providers.md as sources', () => {
  gitIn69(clone69, 'switch', 'main');
  write69(
    clone69,
    'docs/player/getting-started.html',
    `<meta name="promptworld-docs:generated-by" content="fixture">\n<meta name="promptworld-docs:source" content="README.md@${seedPin}">\n`
  );
  write69(
    clone69,
    'docs/player/llm-setup-basics.html',
    `<meta name="promptworld-docs:generated-by" content="fixture">\n<meta name="promptworld-docs:source" content="docs/llm-providers.md@${seedPin}">\n`
  );
  write69(clone69, 'docs/llm-providers.md', '# fixture llm providers\n');
  commitAll69('088 seed: docs/player pages declaring README.md + docs/llm-providers.md');
  gitIn69(clone69, 'push', 'origin', 'main');
});

test('088/US1 (F1): branch touches README.md only (a declared docs/player source) + checker stub exit 1 -> player-docs-stale blocks', () => {
  scenarioBranch69('task-720-readme-declared', () => {
    write69(clone69, 'README.md', '# fixture, updated for spec 088\n');
    commitAll69('088: README-only edit');
  });
  const r = gate69(clone69, 'task-720-readme-declared', { checker: stubExit1 });
  assert.equal(r.status, 1, `declared-source staleness must block: ${r.stdout}${r.stderr}`);
  const f = parseGate69(r).findings.find((x) => x.rule === 'player-docs-stale');
  assert.ok(f, 'expected a player-docs-stale finding');
  assert.equal(f.severity, 'block');
  assert.match(f.message, /README\.md/);
  assert.match(f.message, /player-docs skill/);
});

test('088/US1 (F2): branch touches a declared spec-046-style source (docs/llm-providers.md) -> checker is invoked; fresh stub passes', () => {
  rmSync(sentinelPath, { force: true });
  scenarioBranch69('task-721-llm-providers', () => {
    write69(clone69, 'docs/llm-providers.md', '# fixture llm providers, updated\n');
    commitAll69('088: touch declared docs/llm-providers.md source');
  });
  const rSentinel = gate69(clone69, 'task-721-llm-providers', { checker: stubSentinel });
  assert.equal(rSentinel.status, 0);
  assert.ok(existsSync(sentinelPath), 'checker must be invoked when a declared non-wiki source is touched');

  const rFresh = gate69(clone69, 'task-721-llm-providers', { checker: stubExit0 });
  assert.equal(rFresh.status, 0, `fresh declared-source branch must pass: ${rFresh.stdout}${rFresh.stderr}`);
  const report = parseGate69(rFresh);
  assert.equal(report.findings.find((x) => x.rule === 'player-docs-stale'), undefined);
  assert.equal(report.findings.find((x) => x.rule === 'player-docs-env-error'), undefined);
});

test('088/US1 (F9): checker missing on a README-only trigger -> player-docs-env-error blocks', () => {
  const missingChecker = join(root69, 'does-not-exist-checker.mjs');
  const r = gate69(clone69, 'task-720-readme-declared', { checker: missingChecker });
  assert.equal(r.status, 1, `missing checker on a non-wiki trigger must block: ${r.stdout}${r.stderr}`);
  const f = parseGate69(r).findings.find((x) => x.rule === 'player-docs-env-error');
  assert.ok(f, 'expected a player-docs-env-error finding');
  assert.equal(f.severity, 'block');
  assert.match(f.message, /not found at/);
});

test('088/US3 (F3, F4): branch tip contains a merge of main with no pinned-source diff paths -> probe invoked (stale blocks; fresh passes)', () => {
  const startTip = gitIn69(clone69, 'rev-parse', 'origin/main').stdout.trim();
  gitIn69(clone69, 'switch', '-c', 'task-722-history-move', startTip);
  write69(clone69, 'task-722-branch-work.txt', 'branch work\n');
  commitAll69('088: unrelated branch-work commit before history move');
  gitIn69(clone69, 'switch', 'main');
  write69(clone69, 'main-advance-for-history-move.txt', 'main advanced\n');
  commitAll69('088: advance main for history-move fixture');
  gitIn69(clone69, 'push', 'origin', 'main');
  gitIn69(clone69, 'switch', 'task-722-history-move');
  gitIn69(clone69, 'merge', '--no-ff', 'origin/main', '-m', 'merge main into task-722-history-move');
  gitIn69(clone69, 'switch', 'main');

  const rStale = gate69(clone69, 'task-722-history-move', { checker: stubExit1 });
  assert.equal(rStale.status, 1, `history move must trigger the probe: ${rStale.stdout}${rStale.stderr}`);
  const staleFinding = parseGate69(rStale).findings.find((x) => x.rule === 'player-docs-stale');
  assert.ok(staleFinding, 'expected a player-docs-stale finding from the history-move trigger');
  assert.match(staleFinding.message, /history move/);

  const rFresh = gate69(clone69, 'task-722-history-move', { checker: stubExit0 });
  assert.equal(rFresh.status, 0, `fresh probe after a history move must pass (no false blocking): ${rFresh.stdout}${rFresh.stderr}`);
  const freshReport = parseGate69(rFresh);
  assert.equal(freshReport.findings.find((x) => x.rule === 'player-docs-stale'), undefined);
});

test('088 (F7): combined pinned-input + history-move trigger emits at most one player-docs-stale finding', () => {
  const startTip = gitIn69(clone69, 'rev-parse', 'origin/main').stdout.trim();
  gitIn69(clone69, 'switch', '-c', 'task-723-combined', startTip);
  write69(clone69, 'README.md', '# fixture, updated for combined trigger (088)\n');
  commitAll69('088: touch declared README.md source (combined fixture)');
  gitIn69(clone69, 'switch', 'main');
  write69(clone69, 'main-advance-for-combined.txt', 'main advanced\n');
  commitAll69('088: advance main for combined fixture');
  gitIn69(clone69, 'push', 'origin', 'main');
  gitIn69(clone69, 'switch', 'task-723-combined');
  gitIn69(clone69, 'merge', '--no-ff', 'origin/main', '-m', 'merge main into task-723-combined');
  gitIn69(clone69, 'switch', 'main');

  const r = gate69(clone69, 'task-723-combined', { checker: stubExit1 });
  assert.equal(r.status, 1, `combined trigger must still block: ${r.stdout}${r.stderr}`);
  const findings = parseGate69(r).findings.filter((x) => x.rule === 'player-docs-stale');
  assert.equal(findings.length, 1, 'combined pinned-input + history-move trigger must not duplicate the finding');
  assert.match(findings[0].message, /README\.md/);
  assert.match(findings[0].message, /history move/);
});

test('088/US2 (F5): branch touches internal/tui/ + tui-design stub exit 1 -> tui-design-stale blocks (tui-surface warn also present)', () => {
  scenarioBranch69('task-724-tui-stale', () => {
    write69(clone69, 'internal/tui/example.go', 'package tui // v1\n');
    commitAll69('088: touch internal/tui/ source');
  });
  const r = gate69Full(clone69, 'task-724-tui-stale', { tuiChecker: tuiStubExit1 });
  assert.equal(r.status, 1, `tui-design staleness must block: ${r.stdout}${r.stderr}`);
  const report = parseGate69(r);
  const stale = report.findings.find((x) => x.rule === 'tui-design-stale');
  assert.ok(stale, 'expected a tui-design-stale finding');
  assert.equal(stale.severity, 'block');
  assert.match(stale.message, /docs\/design\/tui\/patterns\/example\.md/);
  const warn = report.findings.find((x) => x.rule === 'tui-surface');
  assert.ok(warn, 'the existing tui-surface warn must remain');
  assert.equal(warn.severity, 'warn');
});

test('088/US2 (F6): same branch after re-pin (tui-design stub exit 0) -> passes, warn-only tui-surface reminder', () => {
  const r = gate69Full(clone69, 'task-724-tui-stale', { tuiChecker: tuiStubExit0 });
  assert.equal(r.status, 0, `re-pinned tui-design must pass: ${r.stdout}${r.stderr}`);
  const report = parseGate69(r);
  assert.equal(report.findings.find((x) => x.rule === 'tui-design-stale'), undefined);
  assert.equal(report.findings.find((x) => x.rule === 'tui-design-env-error'), undefined);
  const warn = report.findings.find((x) => x.rule === 'tui-surface');
  assert.ok(warn && warn.severity === 'warn', 'tui-surface warn stays as a reminder');
});

test('088/US2: tui-design checker missing -> tui-design-env-error blocks', () => {
  const missingTuiChecker = join(root69, 'does-not-exist-tui-checker.mjs');
  const r = gate69Full(clone69, 'task-724-tui-stale', { tuiChecker: missingTuiChecker });
  assert.equal(r.status, 1, `missing tui-design checker must block: ${r.stdout}${r.stderr}`);
  const f = parseGate69(r).findings.find((x) => x.rule === 'tui-design-env-error');
  assert.ok(f, 'expected a tui-design-env-error finding');
  assert.equal(f.severity, 'block');
});

test('088/US2 edge: branch touches docs/design/tui/ only (re-pin-only PR) -> no tui-design-stale, no tui-surface warn', () => {
  scenarioBranch69('task-725-design-repin-only', () => {
    write69(
      clone69,
      'docs/design/tui/patterns/example.md',
      `---\ntitle: fixture pattern\nclass: pattern\nstatus: shipped\nverified_against: ${seedPin}\nsources:\n  - internal/tui/example.go\n---\n\n# Example\n`
    );
    commitAll69('088: re-pin-only docs/design/tui edit');
  });
  // The checker is triggered (docs/design/tui/ touched) but a
  // correctly-behaving check-tui-design.mjs never flags a branch that
  // touches no internal/tui/ file under its same-pr rule; the exit-0 stub
  // stands in for that real outcome (spec 088 edge case).
  const r = gate69Full(clone69, 'task-725-design-repin-only', { tuiChecker: tuiStubExit0 });
  assert.equal(r.status, 0, `re-pin-only PR must pass: ${r.stdout}${r.stderr}`);
  const report = parseGate69(r);
  assert.equal(report.findings.find((x) => x.rule === 'tui-design-stale'), undefined);
  assert.equal(report.findings.find((x) => x.rule === 'tui-surface'), undefined, 'no internal/tui/ touch -> no tui-surface warn');
});

// ---------------------------------------------------------------------------
// TASK-195 — wiki-footprint threshold in session mode. Wiki staleness is
// idempotent per note (a note is stale iff a source changed since its pin), so
// the re-pin bill a branch runs up is a set union over files touched, not a
// sum over edits — deferring grounding costs nothing, but WIDENING does. This
// section brackets the threshold exactly: one branch at it, one a single note
// below. Same fixture pattern as the 069/080 sections.
// ---------------------------------------------------------------------------

const root195 = mkdtempSync(join(tmpdir(), 'wiki-footprint-test-'));
after(() => rmSync(root195, { recursive: true, force: true }));

const origin195 = join(root195, 'origin.git');
run69('git', ['init', '--bare', '-b', 'main', origin195]);
const clone195 = join(root195, 'clone');
run69('git', ['clone', origin195, clone195]);
const gitIn195 = (...args) => gitIn69(clone195, ...args);

function commitAll195(msg) {
  gitIn195('add', '-A');
  gitIn195('commit', '-m', msg);
  return gitIn195('rev-parse', 'HEAD').stdout.trim();
}

// Two sources with deliberately different note fan-out: 30 notes (exactly the
// WIKI_FOOTPRINT_THRESHOLD) and 29 (one below). The check is a >= test, so
// these two branches bracket it with no slack.
write69(clone195, 'internal/wide/w.go', 'package wide // v1\n');
write69(clone195, 'internal/narrow/n.go', 'package narrow // v1\n');
const seed195 = commitAll195('seed: two sources with different note fan-out');
for (let i = 0; i < 30; i++) {
  write69(
    clone195,
    `docs/wiki/wide-${i}.md`,
    `---\nverified_against: ${seed195}\nsources:\n  - internal/wide/w.go\n---\n\n# Wide ${i}\n`
  );
}
for (let i = 0; i < 29; i++) {
  write69(
    clone195,
    `docs/wiki/narrow-${i}.md`,
    `---\nverified_against: ${seed195}\nsources:\n  - internal/narrow/n.go\n---\n\n# Narrow ${i}\n`
  );
}
commitAll195('seed: 30 wide-sourced notes, 29 narrow-sourced notes');
gitIn195('push', 'origin', 'main');

function branch195(name, mutate) {
  gitIn195('switch', '-c', name, 'main');
  try {
    mutate();
  } finally {
    gitIn195('switch', 'main');
  }
}

branch195('task-901-wide', () => {
  write69(clone195, 'internal/wide/w.go', 'package wide // v2\n');
  commitAll195('wide: touch the source 30 notes pin');
});
branch195('task-902-narrow', () => {
  write69(clone195, 'internal/narrow/n.go', 'package narrow // v2\n');
  commitAll195('narrow: touch the source 29 notes pin');
});

function gate195() {
  return run69(process.execPath, [SCRIPT_PATH, 'session', '--json'], { cwd: clone195 });
}

test('195/session: a branch at the wiki-footprint threshold warns and reports its count', () => {
  const report = parseGate69(gate195());
  const wide = report.branches.find((b) => b.name === 'task-901-wide');
  assert.ok(wide, 'expected the wide branch in the report');
  assert.equal(wide.wikiFootprint, 30, 'footprint counts every note whose sources the branch touched');
  const f = report.findings.find((x) => x.rule === 'wiki-footprint' && x.evidence.includes('task-901-wide'));
  assert.ok(f, 'expected a wiki-footprint finding at the threshold');
  assert.equal(f.severity, 'warn', 'session mode is advisory — footprint must never block');
});

test('195/session: a branch one note below the threshold reports its count and does not warn', () => {
  const report = parseGate69(gate195());
  const narrow = report.branches.find((b) => b.name === 'task-902-narrow');
  assert.ok(narrow, 'expected the narrow branch in the report');
  assert.equal(narrow.wikiFootprint, 29, 'the count is reported below threshold, not hidden');
  assert.equal(
    report.findings.find((x) => x.rule === 'wiki-footprint' && x.evidence.includes('task-902-narrow')),
    undefined,
    'below the threshold must not warn'
  );
});

test('195/session: the footprint check never blocks session start', () => {
  const r = gate195();
  assert.equal(r.status, 0, `session must not block on footprint: ${r.stdout}${r.stderr}`);
});

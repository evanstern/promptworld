#!/usr/bin/env node
// claim-protocol.test.mjs — two-session race simulation for the
// claim-before-work protocol (spec 065, FR-006 / SC-001).
//
// Contracts under test:
//   specs/065-claim-before-work/contracts/gate-cli-delta.md — claim mode,
//   worktree --task / claim-aware --spec, session branch-unpushed, and the
//   hook layers' block/fail-open posture.
//
// Fixture pattern follows check-merge-drift.test.mjs (TASK-138): throwaway
// tmpdir fixtures, stdlib only, run with:
//   node --test scripts/claim-protocol.test.mjs
//
// The fixture is a bare origin plus two clones (A and B) sharing it. Every
// git/gate/hook spawn runs against that tmpdir origin with an isolated git
// config — nothing here ever fetches from, or pushes to, the real remote.
// Tests are ORDERED: they walk one race end to end (A claims and wins; B's
// competing push is rejected; B's gates then stop B mechanically).

import { test, after } from 'node:test';
import assert from 'node:assert/strict';
import { mkdtempSync, mkdirSync, writeFileSync, rmSync, readFileSync, copyFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';
import { execFileSync } from 'node:child_process';

const SCRIPTS_DIR = dirname(fileURLToPath(import.meta.url));
const GATE_PATH = join(SCRIPTS_DIR, 'check-merge-drift.mjs');
const HOOK_PATH = join(SCRIPTS_DIR, 'hooks', 'merge-drift-hook.mjs');

// ---------------------------------------------------------------------------
// Fixture: bare origin + clones A and B, isolated git config
// ---------------------------------------------------------------------------

const root = mkdtempSync(join(tmpdir(), 'claim-protocol-test-'));
after(() => rmSync(root, { recursive: true, force: true }));

const gitConfigPath = join(root, 'gitconfig');
writeFileSync(
  gitConfigPath,
  '[user]\n\tname = Fixture\n\temail = fixture@test.invalid\n' +
    '[init]\n\tdefaultBranch = main\n[commit]\n\tgpgsign = false\n'
);
// Isolated config for EVERY spawn (git, gate, hook): the user's real global
// config (signing, hooks) must not leak in, and the fixture must not depend
// on it.
const ENV = {
  ...process.env,
  GIT_CONFIG_GLOBAL: gitConfigPath,
  GIT_CONFIG_SYSTEM: process.platform === 'win32' ? 'NUL' : '/dev/null',
};

function run(cmd, args, { cwd, env = ENV, input } = {}) {
  try {
    const stdout = execFileSync(cmd, args, {
      cwd,
      env,
      encoding: 'utf8',
      stdio: [input === undefined ? 'ignore' : 'pipe', 'pipe', 'pipe'],
      input,
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

const gitIn = (cwd, ...args) => run('git', args, { cwd });
const gate = (cwd, ...args) => run(process.execPath, [GATE_PATH, ...args], { cwd });
const hook = (cwd, sub, stdinText, projectDir) =>
  run(process.execPath, [HOOK_PATH, sub], {
    cwd,
    input: stdinText,
    env: { ...ENV, CLAUDE_PROJECT_DIR: projectDir },
  });

function write(base, rel, content) {
  const abs = join(base, rel);
  mkdirSync(dirname(abs), { recursive: true });
  writeFileSync(abs, content);
}

function card(id, title, status, extraBody = '') {
  return `---\nid: TASK-${id}\ntitle: ${title}\nstatus: ${status}\n---\n\n## Description\n\nFixture card.\n${extraBody}`;
}

const CARD_500 = 'backlog/tasks/task-500 - Unclaimed-fixture-task.md';
const CARD_501 = 'backlog/tasks/task-501 - Alpha-fixture-task.md';

// Seed the shared origin: two To Do cards plus a copy of the gate script (so
// the hook's findGateScript resolves it inside the clones).
const origin = join(root, 'origin.git');
run('git', ['init', '--bare', '-b', 'main', origin]);
const seed = join(root, 'seed');
run('git', ['clone', origin, seed]);
write(seed, CARD_500, card(500, 'Unclaimed fixture task', 'To Do'));
write(seed, CARD_501, card(501, 'Alpha fixture task', 'To Do'));
mkdirSync(join(seed, 'scripts'), { recursive: true });
copyFileSync(GATE_PATH, join(seed, 'scripts', 'check-merge-drift.mjs'));
gitIn(seed, 'add', '-A');
gitIn(seed, 'commit', '-m', 'seed: fixture board + gate script');
gitIn(seed, 'push', 'origin', 'main');

const cloneA = join(root, 'clone-a');
const cloneB = join(root, 'clone-b');
run('git', ['clone', origin, cloneA]);
run('git', ['clone', origin, cloneB]);

// ---------------------------------------------------------------------------
// The race, end to end (SC-001: the loser does zero duplicate work past the
// claim point — stopped by rejection + gates alone)
// ---------------------------------------------------------------------------

test("clone A's claim push is accepted (card -> In Progress + spec dir, one commit)", () => {
  write(
    cloneA,
    CARD_501,
    card(501, 'Alpha fixture task', 'In Progress', '\nSpec: specs/100-alpha (branch task-501-alpha)\n')
  );
  write(cloneA, 'specs/100-alpha/spec.md', '# Claim stub: 100-alpha (TASK-501)\n');
  gitIn(cloneA, 'add', '-A');
  gitIn(cloneA, 'commit', '-m', 'claim: TASK-501 In Progress + specs/100-alpha');
  const push = gitIn(cloneA, 'push', 'origin', 'main');
  assert.equal(push.status, 0, `A's claim push must be accepted: ${push.stderr}`);
});

test("clone B's competing claim push is rejected non-fast-forward", () => {
  // B is stale: it has not fetched A's claim and races for the same task and
  // the same number under a different slug.
  write(
    cloneB,
    CARD_501,
    card(501, 'Alpha fixture task', 'In Progress', '\nSpec: specs/100-beta (branch task-501-beta)\n')
  );
  write(cloneB, 'specs/100-beta/spec.md', '# Claim stub: 100-beta (TASK-501, loser)\n');
  gitIn(cloneB, 'add', '-A');
  gitIn(cloneB, 'commit', '-m', 'claim: TASK-501 In Progress + specs/100-beta (competing)');
  const push = gitIn(cloneB, 'push', 'origin', 'main');
  assert.notEqual(push.status, 0, 'the losing push must be rejected');
  assert.match(push.stderr, /(non-fast-forward|fetch first|rejected)/, `expected a non-fast-forward rejection, got: ${push.stderr}`);
});

test("after fetch, B's `claim --dir 100-beta` blocks, naming A's dir and the next free number", () => {
  gitIn(cloneB, 'fetch', 'origin');
  const r = gate(cloneB, 'claim', '--dir', '100-beta', '--json');
  assert.equal(r.status, 1, `claim must exit 1 on a taken number: ${r.stdout}${r.stderr}`);
  const report = JSON.parse(r.stdout);
  assert.equal(report.mode, 'claim');
  assert.equal(report.verdict, 'blocked');
  const f = report.findings.find((x) => x.rule === 'spec-number-collision');
  assert.ok(f, 'expected a spec-number-collision finding');
  assert.equal(f.severity, 'block');
  assert.match(f.message, /specs\/100-alpha/, 'must name the taken dir');
  assert.match(f.message, /101/, 'must name the next free number');
  assert.ok(f.evidence.includes('specs/100-alpha'));
});

test('`claim` is idempotent for the owner: A re-running against its own dirname passes', () => {
  const r = gate(cloneA, 'claim', '--dir', '100-alpha', '--json');
  assert.equal(r.status, 0, `owner re-claim must pass: ${r.stdout}${r.stderr}`);
  const report = JSON.parse(r.stdout);
  assert.equal(report.verdict, 'pass');
  assert.equal(report.findings.length, 0);
});

test('`worktree --task` warns card-not-claimed for an unclaimed card and stays quiet for the claimed one', () => {
  // Fixture bookkeeping: the loser abandons its local claim so its root is
  // exactly at the fetched origin/main tip (worktree mode requires that).
  gitIn(cloneB, 'reset', '--hard', 'origin/main');

  const unclaimed = gate(cloneB, 'worktree', '--task', 'TASK-500', '--json');
  assert.equal(unclaimed.status, 0, 'card-not-claimed is warn-only — exit must stay 0');
  const unclaimedReport = JSON.parse(unclaimed.stdout);
  const warn = unclaimedReport.findings.find((f) => f.rule === 'card-not-claimed');
  assert.ok(warn, 'expected a card-not-claimed warning for the To Do card');
  assert.equal(warn.severity, 'warn');
  assert.equal(warn.task, 'TASK-500');
  assert.match(warn.message, /not In Progress on origin\/main/);

  const claimed = gate(cloneB, 'worktree', '--task', 'TASK-501', '--json');
  assert.equal(claimed.status, 0);
  const claimedReport = JSON.parse(claimed.stdout);
  assert.equal(
    claimedReport.findings.find((f) => f.rule === 'card-not-claimed'),
    undefined,
    "A's claimed card (In Progress on origin/main) must not warn"
  );
});

test('`worktree --spec` is claim-aware: owned claim passes, foreign or unowned claim blocks', () => {
  // The dir exists on origin/main and its Spec marker attributes to TASK-501:
  // ownership, not absence, is the invariant once claims exist.
  const owned = gate(cloneB, 'worktree', '--spec', '100', '--task', 'TASK-501', '--json');
  assert.equal(owned.status, 0, `owned spec dir must pass: ${owned.stdout}${owned.stderr}`);
  assert.equal(
    JSON.parse(owned.stdout).findings.find((f) => f.rule === 'spec-number-collision'),
    undefined
  );

  const foreign = gate(cloneB, 'worktree', '--spec', '100', '--task', 'TASK-500', '--json');
  assert.equal(foreign.status, 1, 'a spec dir attributed to a different task must block');
  assert.match(
    JSON.parse(foreign.stdout).findings.find((f) => f.rule === 'spec-number-collision').message,
    /attributed to TASK-501, not TASK-500/
  );

  const legacy = gate(cloneB, 'worktree', '--spec', '100', '--json');
  assert.equal(legacy.status, 1, 'without --task, pre-065 semantics hold: taken number blocks');
});

test('session mode: branch-unpushed fires for a local-only task branch and clears when pushed', () => {
  gitIn(cloneA, 'switch', '-c', 'task-600-local-only');
  write(cloneA, 'notes-600.md', 'local-only work\n');
  gitIn(cloneA, 'add', '-A');
  gitIn(cloneA, 'commit', '-m', 'task-600: local-only commit');
  gitIn(cloneA, 'switch', 'main');

  const before = gate(cloneA, 'session', '--json');
  assert.equal(before.status, 0, `warn-only session run must exit 0: ${before.stdout}${before.stderr}`);
  const finding = JSON.parse(before.stdout).findings.find((f) => f.rule === 'branch-unpushed');
  assert.ok(finding, 'expected branch-unpushed for the local-only branch');
  assert.equal(finding.severity, 'warn');
  assert.equal(finding.task, 'TASK-600');
  assert.match(finding.message, /git push -u origin task-600-local-only/);

  gitIn(cloneA, 'push', '-u', 'origin', 'task-600-local-only');
  const after_ = gate(cloneA, 'session', '--json');
  assert.equal(after_.status, 0);
  assert.equal(
    JSON.parse(after_.stdout).findings.find((f) => f.rule === 'branch-unpushed'),
    undefined,
    'the finding must clear once the branch has a remote counterpart'
  );
});

// ---------------------------------------------------------------------------
// Hook layers (thin wrappers — block wiring + fail-open + owned-dir pass)
// ---------------------------------------------------------------------------

test('pre-write hook: colliding spec-dir write blocks; write into the owned dir passes', () => {
  const blocked = hook(
    cloneB,
    'pre-write',
    JSON.stringify({ tool_input: { file_path: join(cloneB, 'specs/100-beta/spec.md') }, cwd: cloneB }),
    cloneB
  );
  assert.equal(blocked.status, 2, `colliding write must block: ${blocked.stderr}`);
  assert.match(blocked.stderr, /specs\/100-alpha/);

  const owned = hook(
    cloneA,
    'pre-write',
    JSON.stringify({ tool_input: { file_path: join(cloneA, 'specs/100-alpha/plan.md') }, cwd: cloneA }),
    cloneA
  );
  assert.equal(owned.status, 0, `a session editing its own claimed spec dir must never be blocked: ${owned.stderr}`);
});

test('pre-bash hook: mkdir of a colliding spec dir blocks; owned mkdir and unrelated commands pass', () => {
  const blocked = hook(
    cloneB,
    'pre-bash',
    JSON.stringify({ tool_input: { command: 'mkdir -p specs/100-beta' }, cwd: cloneB }),
    cloneB
  );
  assert.equal(blocked.status, 2, `colliding mkdir must block: ${blocked.stderr}`);

  const owned = hook(
    cloneA,
    'pre-bash',
    JSON.stringify({ tool_input: { command: 'mkdir -p specs/100-alpha/contracts' }, cwd: cloneA }),
    cloneA
  );
  assert.equal(owned.status, 0, `owned mkdir must pass: ${owned.stderr}`);

  const unrelated = hook(
    cloneA,
    'pre-bash',
    JSON.stringify({ tool_input: { command: 'ls -la specs/' }, cwd: cloneA }),
    cloneA
  );
  assert.equal(unrelated.status, 0);
});

test('hook layers fail open on malformed stdin', () => {
  for (const sub of ['pre-write', 'pre-bash']) {
    const r = hook(cloneA, sub, 'this is not json', cloneA);
    assert.equal(r.status, 0, `${sub} must fail open on malformed stdin`);
  }
});

// ---------------------------------------------------------------------------
// Branch-held claims (spec 111, TASK-188)
//
// The hole these cover: spec 065 requires a claim stub to be merged to main
// immediately, but nothing enforces it. When that step slips the number is
// claimed only on a pushed branch — invisible to a gate that reads origin/main
// alone. Observed live on 2026-08-02 (TASK-173 and TASK-187 both held 110).
//
// These tests run LAST and are order-dependent on the ones above: they add
// number 200 on a branch, which moves the next-free number the earlier
// main-held tests assert on.
// ---------------------------------------------------------------------------

// A claims number 200 on a pushed branch ONLY — main never sees it. This is the
// exact shape the pre-111 gate could not detect.
test('fixture: A claims specs/200-gamma on a pushed branch, never merged to main', () => {
  gitIn(cloneA, 'switch', '-c', 'task-700-gamma');
  write(cloneA, 'specs/200-gamma/spec.md', '# Claim stub: 200-gamma (TASK-700)\n');
  gitIn(cloneA, 'add', '-A');
  gitIn(cloneA, 'commit', '-m', 'TASK-700: claim spec 200 (stub)');
  const push = gitIn(cloneA, 'push', '-u', 'origin', 'task-700-gamma');
  assert.equal(push.status, 0, `branch push must succeed: ${push.stderr}`);
  gitIn(cloneA, 'switch', 'main');

  const onMain = gitIn(cloneA, 'ls-tree', '--name-only', 'origin/main', '--', 'specs/');
  assert.doesNotMatch(onMain.stdout, /200-gamma/, 'the claim must NOT be on origin/main — that is the whole point');
});

test('branch-held number blocks a competing claim and names the holding branch', () => {
  gitIn(cloneB, 'fetch', 'origin');
  const r = gate(cloneB, 'claim', '--dir', '200-delta', '--json');
  assert.equal(r.status, 1, `a branch-held number must block: ${r.stdout}${r.stderr}`);
  const report = JSON.parse(r.stdout);
  assert.equal(report.verdict, 'blocked');
  const f = report.findings.find((x) => x.rule === 'spec-number-collision');
  assert.ok(f, 'expected a spec-number-collision finding');
  assert.equal(f.severity, 'block');
  assert.match(f.message, /specs\/200-gamma/, 'must name the taken dir');
  assert.match(f.message, /origin\/task-700-gamma/, 'must name the HOLDING BRANCH (FR-004)');
  assert.ok(
    f.evidence.includes('origin/task-700-gamma'),
    'the branch belongs in evidence, not only in the prose'
  );
});

test('next free skips branch-held numbers, not just main-held ones (FR-005)', () => {
  const r = gate(cloneB, 'claim', '--dir', '200-delta', '--json');
  const f = JSON.parse(r.stdout).findings.find((x) => x.rule === 'spec-number-collision');
  assert.match(f.message, /next free number is 201/, 'main tops out at 100; the branch holds 200, so 201 is next');
  assert.doesNotMatch(f.message, /next free number is 101/, 'a main-only next-free would walk the caller into 200');
});

test('`claim` is idempotent when the owner\'s holder lives on its own branch', () => {
  const r = gate(cloneA, 'claim', '--dir', '200-gamma', '--json');
  assert.equal(r.status, 0, `owner re-claim against a branch-held dir must pass: ${r.stdout}${r.stderr}`);
  const report = JSON.parse(r.stdout);
  assert.equal(report.verdict, 'pass');
  assert.equal(report.findings.length, 0);
});

test('a free number still passes with branch claims present', () => {
  const r = gate(cloneB, 'claim', '--dir', '201-epsilon', '--json');
  assert.equal(r.status, 0, `an unheld number must pass: ${r.stdout}${r.stderr}`);
  assert.equal(JSON.parse(r.stdout).verdict, 'pass');
});

test('origin/main wins attribution when main AND a branch both hold the number (FR-003)', () => {
  // 100-alpha is already on main (A's original claim). Put a DIFFERENT dirname
  // at 100 on a branch, then claim a third: the settled record must be named.
  gitIn(cloneA, 'switch', '-c', 'task-701-zeta');
  write(cloneA, 'specs/100-zeta/spec.md', '# Competing branch dir at number 100\n');
  gitIn(cloneA, 'add', '-A');
  gitIn(cloneA, 'commit', '-m', 'TASK-701: branch-side dir at number 100');
  gitIn(cloneA, 'push', '-u', 'origin', 'task-701-zeta');
  gitIn(cloneA, 'switch', 'main');
  gitIn(cloneB, 'fetch', 'origin');

  const r = gate(cloneB, 'claim', '--dir', '100-omega', '--json');
  assert.equal(r.status, 1, 'still a collision');
  const f = JSON.parse(r.stdout).findings.find((x) => x.rule === 'spec-number-collision');
  assert.match(f.message, /specs\/100-alpha already exists on origin\/main/, 'main is the settled record and must be named');
  assert.doesNotMatch(f.message, /100-zeta/, 'the in-flight branch dir must not displace main in the message');
});

test('a spec dir on a NON-task remote branch is not treated as a claim', () => {
  // Scope guard: only origin/task-* are claims. A release/experiment branch
  // carrying spec dirs must not lock numbers for everyone.
  gitIn(cloneA, 'switch', '-c', 'experiment-noise');
  write(cloneA, 'specs/300-noise/spec.md', '# Not a claim\n');
  gitIn(cloneA, 'add', '-A');
  gitIn(cloneA, 'commit', '-m', 'experiment: spec dir on a non-task branch');
  gitIn(cloneA, 'push', '-u', 'origin', 'experiment-noise');
  gitIn(cloneA, 'switch', 'main');
  gitIn(cloneB, 'fetch', 'origin');

  const r = gate(cloneB, 'claim', '--dir', '300-real', '--json');
  assert.equal(r.status, 0, `a non-task branch must not lock number 300: ${r.stdout}${r.stderr}`);
  assert.equal(JSON.parse(r.stdout).verdict, 'pass');
});

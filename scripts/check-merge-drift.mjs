#!/usr/bin/env node
// check-merge-drift.mjs — deterministic merge-drift gates for the parallel
// worktree SDLC, run at four choke points: session start, spec-directory
// claim, worktree cut, and PR open.
//
// Contracts (normative):
//   specs/051-merge-drift-gates/contracts/gate-cli.md
//   specs/051-merge-drift-gates/contracts/detection-rules.md
//   specs/051-merge-drift-gates/contracts/report-schema.md
//   specs/065-claim-before-work/contracts/gate-cli-delta.md — claim mode,
//   worktree --task / claim-aware --spec, session branch-unpushed
//
// Node >= 18, ESM, zero npm dependencies (stdlib fs/path/child_process/crypto
// only). Requires git >= 2.38 (`git merge-tree --write-tree`).
//
// Mutation whitelist (FR-009) — the ONLY writes this script ever performs,
// under any flag combination:
//   1. `git fetch origin` (always attempted unless --no-fetch)
//   2. Fast-forward of the ROOT checkout — session mode only, automatic,
//      only when root is on main, behind origin/main, not diverged, clean.
//   3. With --apply-cleanup: `git worktree remove` + `git branch -d`/`-D` for
//      worktrees THIS run verified cleanup-eligible. Nothing else. `-D` is
//      used only for the empty-contribution (squash) reason, where `-d`'s
//      ancestor check can never succeed even though this run already
//      cryptographically proved the branch contributes nothing new via
//      merge-tree tree-equality (deviation from detection-rules.md §4's
//      literal `-d`; see cleanup-eligible finding construction below).
//   4. With --notes: `backlog task edit TASK-<N> --append-notes …` — never a
//      direct write under backlog/.
// It never rebases, merges, commits to, checks out, or resets any task
// branch or its worktree. Conflict resolution always belongs to the
// branch's owning session.
//
// Usage:
//   node scripts/check-merge-drift.mjs <mode> [flags]
//
//   modes: session | claim | worktree | pr
//
//   --json              machine-readable report (report-schema.md)
//   --dir <NNN-slug>     (claim) the spec directory being claimed/created
//   --spec <NNN>         (worktree) also verify spec number NNN is unused
//                        on origin/main — or, with --task, owned by that task
//   --task <TASK-n>      (worktree) warn when the task's board card is not
//                        In Progress on origin/main (claim missing/unpushed)
//   --branch <name>       (pr) gate a branch other than the current checkout
//   --notes              record task-attributable findings as board notes
//   --apply-cleanup       (session) apply prescribed cleanup-eligible removals
//   --no-fetch            (session) skip fetch — degraded/unverified mode
//
// Exit codes:
//   0  pass — clean or warnings-only
//   1  blocked — >= 1 block-severity finding
//   2  usage/environment error (bad invocation, not a git repo, git < 2.38,
//      or fetch failure in a fail-closed mode: pr, worktree, claim)

import { existsSync, readFileSync, readdirSync, realpathSync } from 'node:fs';
import { join } from 'node:path';
import { execFileSync } from 'node:child_process';
import { createHash } from 'node:crypto';
import { fileURLToPath } from 'node:url';

// ---------------------------------------------------------------------------
// Usage / preflight
// ---------------------------------------------------------------------------

function usageError(msg) {
  process.stderr.write(`check-merge-drift: ${msg}\n`);
  process.exit(2);
}

const MODES = ['session', 'claim', 'worktree', 'pr'];

function parseArgs(argv) {
  if (argv.length === 0) usageError('missing mode (session|claim|worktree|pr)');
  const mode = argv[0];
  if (!MODES.includes(mode)) usageError(`unknown mode ${JSON.stringify(mode)}`);

  const flags = {
    mode,
    json: false,
    dir: null,
    spec: null,
    task: null,
    branch: null,
    notes: false,
    applyCleanup: false,
    noFetch: false,
  };

  for (let i = 1; i < argv.length; i++) {
    const arg = argv[i];
    if (arg === '--json') {
      flags.json = true;
    } else if (arg === '--dir') {
      const val = argv[++i];
      if (val === undefined) usageError('--dir requires a value');
      flags.dir = val;
    } else if (arg === '--spec') {
      const val = argv[++i];
      if (val === undefined) usageError('--spec requires a value');
      flags.spec = val;
    } else if (arg === '--task') {
      const val = argv[++i];
      if (val === undefined) usageError('--task requires a value');
      flags.task = val;
    } else if (arg === '--branch') {
      const val = argv[++i];
      if (val === undefined) usageError('--branch requires a value');
      flags.branch = val;
    } else if (arg === '--notes') {
      flags.notes = true;
    } else if (arg === '--apply-cleanup') {
      flags.applyCleanup = true;
    } else if (arg === '--no-fetch') {
      flags.noFetch = true;
    } else {
      usageError(`unknown flag ${JSON.stringify(arg)}`);
    }
  }

  if (flags.dir !== null && mode !== 'claim') {
    usageError('--dir is only valid in claim mode');
  }
  if (mode === 'claim' && flags.dir === null) {
    usageError('claim mode requires --dir <NNN-slug>');
  }
  if (flags.dir !== null && !/^\d{3,}-[A-Za-z0-9._-]+$/.test(flags.dir)) {
    usageError('--dir must be a spec directory name like 065-claim-before-work (NNN-slug)');
  }
  if (flags.spec !== null && mode !== 'worktree') {
    usageError('--spec is only valid in worktree mode');
  }
  if (flags.spec !== null && !/^\d{1,5}$/.test(flags.spec)) {
    usageError('--spec must be numeric, e.g. 052');
  }
  if (flags.task !== null && mode !== 'worktree') {
    usageError('--task is only valid in worktree mode');
  }
  if (flags.task !== null && !/^TASK-\d+$/.test(flags.task)) {
    usageError('--task must look like TASK-139');
  }
  if (flags.branch !== null && mode !== 'pr') {
    usageError('--branch is only valid in pr mode');
  }
  if (flags.applyCleanup && mode !== 'session') {
    usageError('--apply-cleanup is only valid in session mode');
  }
  if (flags.noFetch && mode !== 'session') {
    usageError('--no-fetch is only valid in session mode (forbidden in pr/worktree — they fail closed instead)');
  }

  return flags;
}

function preflight() {
  const v = tryExec('git', ['version']);
  if (v.status !== 0) usageError('git not found or not runnable');
  const m = v.stdout.match(/git version (\d+)\.(\d+)/);
  if (!m) usageError(`cannot parse git version output: ${JSON.stringify(v.stdout)}`);
  const major = parseInt(m[1], 10);
  const minor = parseInt(m[2], 10);
  if (major < 2 || (major === 2 && minor < 38)) {
    usageError(`git >= 2.38 required for --write-tree merge-tree (found ${major}.${minor})`);
  }
  const top = tryExec('git', ['rev-parse', '--show-toplevel']);
  if (top.status !== 0) usageError('not inside a git repository');
  return top.stdout.trim();
}

// ---------------------------------------------------------------------------
// git plumbing
// ---------------------------------------------------------------------------

function tryExec(cmd, args, { cwd = process.cwd() } = {}) {
  try {
    const stdout = execFileSync(cmd, args, {
      cwd,
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'pipe'],
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

function git(args, opts) {
  return tryExec('git', args, opts);
}

// Throws-if-fails helper for git commands that are expected to succeed given
// a healthy repo; environment errors here are genuinely exceptional.
function gitOk(args, opts) {
  const r = git(args, opts);
  if (r.status !== 0) {
    usageError(`git ${args.join(' ')} failed: ${(r.stderr || r.stdout).trim()}`);
  }
  return r.stdout;
}

function runNode(scriptAbsPath, args, cwd) {
  try {
    const stdout = execFileSync(process.execPath, [scriptAbsPath, ...args], {
      cwd,
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'pipe'],
    });
    return { status: 0, stdout };
  } catch (err) {
    return { status: typeof err.status === 'number' ? err.status : 1, stdout: err.stdout ? err.stdout.toString() : '' };
  }
}

// git worktree list --porcelain -> [{ path, branch|null }], main worktree first.
function listWorktrees(cwd) {
  const out = gitOk(['worktree', 'list', '--porcelain'], { cwd });
  const blocks = out.split(/\n\n+/).map((b) => b.trim()).filter(Boolean);
  return blocks.map((b) => {
    const lines = b.split('\n');
    const wtLine = lines.find((l) => l.startsWith('worktree '));
    const branchLine = lines.find((l) => l.startsWith('branch '));
    return {
      path: wtLine ? wtLine.slice('worktree '.length) : null,
      branch: branchLine ? branchLine.slice('branch refs/heads/'.length) : null,
    };
  });
}

function resolveMainWorktree(cwd) {
  const wts = listWorktrees(cwd);
  if (wts.length === 0 || !wts[0].path) usageError('could not resolve the main worktree from `git worktree list`');
  return wts[0].path;
}

function findWorktreeForBranch(cwd, branchName) {
  const hit = listWorktrees(cwd).find((w) => w.branch === branchName);
  return hit ? hit.path : null;
}

function isDirty(cwd) {
  const r = git(['status', '--porcelain'], { cwd });
  return r.status === 0 && r.stdout.trim().length > 0;
}

// `git merge-tree --write-tree --name-only <a> <b>`
//   exit 0 -> clean; stdout line 1 = merged tree OID
//   exit 1 -> conflict; stdout = OID, then conflicted paths, then a blank
//             line, then optional informational messages
//   other  -> environment error
function mergeTree(a, b, cwd) {
  const r = git(['merge-tree', '--write-tree', '--name-only', a, b], { cwd });
  const lines = r.stdout.split('\n');
  const treeOid = (lines[0] || '').trim();
  if (r.status === 0) {
    return { conflict: false, treeOid, files: [] };
  }
  if (r.status === 1) {
    const files = [];
    for (let i = 1; i < lines.length; i++) {
      if (lines[i].trim() === '') break;
      files.push(lines[i].trim());
    }
    return { conflict: true, treeOid, files };
  }
  return { conflict: null, treeOid: null, files: [], error: (r.stderr || r.stdout).trim() };
}

function mergeBaseOf(a, b, cwd) {
  const r = git(['merge-base', a, b], { cwd });
  return r.status === 0 ? r.stdout.trim() : null;
}

function isAncestor(tip, ref, cwd) {
  return git(['merge-base', '--is-ancestor', tip, ref], { cwd }).status === 0;
}

function revListCount(range, cwd) {
  const r = git(['rev-list', '--count', range], { cwd });
  return r.status === 0 ? parseInt(r.stdout.trim(), 10) : 0;
}

function changedFiles(base, ref, cwd, paths) {
  const args = ['diff', '--name-only', base, ref];
  if (paths && paths.length) args.push('--', ...paths);
  const r = git(args, { cwd });
  return r.status === 0 ? r.stdout.split('\n').filter(Boolean) : [];
}

// ---------------------------------------------------------------------------
// Findings / verdict (data-model.md)
// ---------------------------------------------------------------------------

const RANK = { info: 0, warn: 1, block: 2 };

export function fingerprint(gate, rule, key, evidence) {
  const sorted = [...evidence].sort().join(',');
  const h = createHash('sha256').update(`${gate}|${rule}|${key || ''}|${sorted}`).digest('hex');
  return h.slice(0, 12);
}

function makeFinding({ severity, gate, rule, message, evidence, task, key }) {
  return {
    severity,
    gate,
    rule,
    message,
    evidence,
    task: task ?? null,
    fingerprint: fingerprint(gate, rule, key, evidence),
    noteWritten: false,
  };
}

function sortFindings(findings) {
  return [...findings].sort((a, b) => {
    if (RANK[b.severity] !== RANK[a.severity]) return RANK[b.severity] - RANK[a.severity];
    if (a.rule !== b.rule) return a.rule < b.rule ? -1 : 1;
    const ae = a.evidence[0] || '';
    const be = b.evidence[0] || '';
    if (ae !== be) return ae < be ? -1 : 1;
    return 0;
  });
}

function computeVerdict(findings) {
  let maxRank = -1;
  for (const f of findings) if (RANK[f.severity] > maxRank) maxRank = RANK[f.severity];
  if (maxRank === RANK.block) return 'blocked';
  if (maxRank === RANK.warn) return 'warnings';
  return 'pass';
}

function exitCodeFor(verdict) {
  return verdict === 'blocked' ? 1 : 0;
}

// ---------------------------------------------------------------------------
// Task attribution (§7)
// ---------------------------------------------------------------------------

function attributeTask(branchName) {
  const m = branchName.match(/^task-(\d+)-/);
  return m ? `TASK-${m[1]}` : null;
}

function attributeBySpecDir(mainWt, specDirName) {
  const dir = join(mainWt, 'backlog', 'tasks');
  if (!existsSync(dir)) return null;
  for (const f of readdirSync(dir)) {
    if (!f.endsWith('.md')) continue;
    let text;
    try {
      text = readFileSync(join(dir, f), 'utf8');
    } catch {
      continue;
    }
    if (text.includes(`Spec: specs/${specDirName}`)) {
      const idm = text.match(/^id:\s*(TASK-\d+)/m);
      if (idm) return idm[1];
    }
  }
  return null;
}

function findTaskFile(mainWt, taskId) {
  const n = taskId.replace(/^TASK-/, '');
  const dir = join(mainWt, 'backlog', 'tasks');
  if (!existsSync(dir)) return null;
  const re = new RegExp(`^task-${n}([ .]|$)`);
  for (const f of readdirSync(dir)) {
    if (f.endsWith('.md') && re.test(f)) return join(dir, f);
  }
  return null;
}

// --- origin/main TREE readers (spec 065) -----------------------------------
// The claim protocol defines ownership by presence on origin/main, so these
// helpers read the FETCHED tree (ls-tree + show) and never the working dir —
// a local card move that was never pushed is exactly the failure they must
// catch. Pure reads; the mutation whitelist is untouched.

function taskCardsOnOriginMain(originMainTip, cwd) {
  // -z: card filenames routinely contain non-ASCII (em dashes), which plain
  // ls-tree C-quotes — breaking both the regex match and the later git show.
  const r = git(['ls-tree', '--name-only', '-z', originMainTip, '--', 'backlog/tasks/'], { cwd });
  if (r.status !== 0) return [];
  return r.stdout.split('\0').filter((l) => l.endsWith('.md'));
}

// -> { path: string|null, status: string|null }. Missing card or unparseable
// frontmatter yields status null — callers treat that as "not claimed".
function cardStatusOnOriginMain(originMainTip, taskId, cwd) {
  const n = taskId.replace(/^TASK-/, '');
  const re = new RegExp(`^task-${n}([ .]|$)`);
  const path = taskCardsOnOriginMain(originMainTip, cwd).find((p) =>
    re.test(p.replace(/^backlog\/tasks\//, ''))
  );
  if (!path) return { path: null, status: null };
  const show = git(['show', `${originMainTip}:${path}`], { cwd });
  if (show.status !== 0) return { path, status: null };
  const m = show.stdout.match(/^status:\s*(.+?)\s*$/m);
  return { path, status: m ? m[1].trim() : null };
}

// attributeBySpecDir against the origin/main tree (Spec marker lookup) — used
// where ownership must be judged from the pushed state, not local files.
function attributeBySpecDirOnOriginMain(originMainTip, specDirName, cwd) {
  for (const p of taskCardsOnOriginMain(originMainTip, cwd)) {
    const show = git(['show', `${originMainTip}:${p}`], { cwd });
    if (show.status !== 0) continue;
    if (show.stdout.includes(`Spec: specs/${specDirName}`)) {
      const idm = show.stdout.match(/^id:\s*(TASK-\d+)/m);
      if (idm) return idm[1];
    }
  }
  return null;
}

// ---------------------------------------------------------------------------
// Semantic overlap primitives (§3, §5)
// ---------------------------------------------------------------------------

function backlogOverlap(branchFiles, mainFiles) {
  const mset = new Set(mainFiles.filter((f) => f.startsWith('backlog/')));
  return branchFiles.filter((f) => f.startsWith('backlog/') && mset.has(f));
}

function tuiSurfaceFiles(branchFiles) {
  return branchFiles.filter((f) => f.startsWith('internal/tui/'));
}

function wikiSourcesOverlap(branchFiles, wikiNotes) {
  const bset = new Set(branchFiles);
  const hits = new Map(); // notePath -> matched files
  for (const note of wikiNotes) {
    const matched = note.sources.filter((s) => bset.has(s));
    if (matched.length) hits.set(note.path, matched);
  }
  return hits;
}

function takenSpecNumbers(originMainTip, cwd) {
  const r = git(['ls-tree', '--name-only', originMainTip, '--', 'specs/'], { cwd });
  const map = new Map(); // number -> dirname (no "specs/" prefix)
  if (r.status !== 0) return map;
  for (const line of r.stdout.split('\n').filter(Boolean)) {
    const name = line.replace(/^specs\//, '');
    const m = name.match(/^(\d+)-(.+)$/);
    if (m) map.set(parseInt(m[1], 10), name);
  }
  return map;
}

function newSpecDirsFromFiles(files) {
  const dirs = new Set();
  for (const f of files) {
    const m = f.match(/^specs\/(\d+)-([^/]+)\//);
    if (m) dirs.add(`${m[1]}-${m[2]}`);
  }
  return dirs;
}

function specNumberCollisions(branchFiles, takenMap) {
  const out = [];
  for (const dir of newSpecDirsFromFiles(branchFiles)) {
    const m = dir.match(/^(\d+)-/);
    const number = parseInt(m[1], 10);
    if (takenMap.has(number) && takenMap.get(number) !== dir) {
      out.push({ number, newDir: dir, takenDir: takenMap.get(number) });
    }
  }
  return out;
}

function formatSpecNum(n) {
  return String(n).padStart(3, '0');
}

// ---------------------------------------------------------------------------
// Wiki frontmatter (§3, §6)
// ---------------------------------------------------------------------------

function parseFrontmatter(text) {
  const m = text.match(/^---\r?\n([\s\S]*?)\r?\n---/);
  if (!m) return null;
  const body = m[1];
  const vaMatch = body.match(/^verified_against:\s*(\S+)\s*$/m);
  const verified_against = vaMatch ? vaMatch[1].trim() : null;
  const srcMatch = body.match(/^sources:\s*\n((?:^\s*-\s*.+\n?)+)/m);
  let sources = [];
  if (srcMatch) {
    sources = srcMatch[1]
      .split('\n')
      .map((l) => l.trim())
      .filter((l) => l.startsWith('- '))
      .map((l) => l.slice(2).trim());
  }
  return { verified_against, sources };
}

function loadWikiNotes(cwd, originMainTip) {
  const r = git(['ls-tree', '-r', '--name-only', originMainTip, '--', 'docs/wiki'], { cwd });
  if (r.status !== 0) return [];
  const paths = r.stdout.split('\n').filter((l) => l.endsWith('.md'));
  const notes = [];
  for (const p of paths) {
    const textRes = git(['show', `${originMainTip}:${p}`], { cwd });
    if (textRes.status !== 0) continue;
    const fm = parseFrontmatter(textRes.stdout);
    const malformed = !fm || !fm.verified_against || !fm.sources || fm.sources.length === 0;
    notes.push({ path: p, sources: fm ? fm.sources : [], verified_against: fm ? fm.verified_against : null, malformed });
  }
  return notes;
}

// Read specific notes' frontmatter at an arbitrary ref (spec 069, plan D1):
// pr mode's pin-vs-branch predicate judges the BRANCH TIP's committed state
// (research R1 — an uncommitted re-pin isn't in the PR, so it can't satisfy a
// PR gate). Same frontmatter grammar as loadWikiNotes — one parser, one
// grammar. Returns Map(notePath -> note | null); null means the file does not
// exist at ref (e.g. deleted on the branch).
function loadWikiNotesAt(ref, notePaths, cwd) {
  const notes = new Map();
  for (const p of notePaths) {
    const textRes = git(['show', `${ref}:${p}`], { cwd });
    if (textRes.status !== 0) {
      notes.set(p, null);
      continue;
    }
    const fm = parseFrontmatter(textRes.stdout);
    const malformed = !fm || !fm.verified_against || !fm.sources || fm.sources.length === 0;
    notes.set(p, {
      path: p,
      sources: fm ? fm.sources : [],
      verified_against: fm ? fm.verified_against : null,
      malformed,
    });
  }
  return notes;
}

function computeWikiGrounding(notes, cwd, originMainTip) {
  const findings = [];
  let anyStale = false;
  const touchedAll = new Set();
  for (const n of notes) {
    if (n.malformed) {
      findings.push(
        makeFinding({
          severity: 'info',
          gate: 'session',
          rule: 'wiki-note-malformed',
          message: `${n.path}: missing/incomplete sources: / verified_against: frontmatter`,
          evidence: [n.path],
          task: null,
          key: n.path,
        })
      );
      continue;
    }
    const pinCheck = git(['cat-file', '-e', `${n.verified_against}^{commit}`], { cwd });
    if (pinCheck.status !== 0) {
      findings.push(
        makeFinding({
          severity: 'info',
          gate: 'session',
          rule: 'wiki-note-malformed',
          message: `${n.path}: verified_against ${n.verified_against} is unresolvable`,
          evidence: [n.path],
          task: null,
          key: n.path,
        })
      );
      continue;
    }
    const touched = changedFiles(n.verified_against, originMainTip, cwd, n.sources);
    if (touched.length) {
      anyStale = true;
      touched.forEach((t) => touchedAll.add(t));
      findings.push(
        makeFinding({
          severity: 'warn',
          gate: 'session',
          rule: 'grounding-stale',
          message: `${n.path} lists source(s) touched on origin/main since its pin — re-check via /grounding-wiki:wiki-update`,
          evidence: [n.path, ...touched],
          task: null,
          key: n.path,
        })
      );
    }
  }
  return {
    surface: { name: 'wiki', checker: 'internal', stale: anyStale, touched: [...touchedAll] },
    findings,
  };
}

function computeDelegatedSurface(name, mainWt, relScriptPath, args, extractTouched) {
  const abs = join(mainWt, relScriptPath);
  if (!existsSync(abs)) return { name, checker: 'absent', stale: null, touched: [] };
  const r = runNode(abs, args, mainWt);
  let touched = [];
  try {
    touched = extractTouched(JSON.parse(r.stdout)) || [];
  } catch {
    touched = [];
  }
  return { name, checker: 'delegated', stale: r.status !== 0, touched: [...new Set(touched)] };
}

export function computePlayerDocsSurface(mainWt) {
  return computeDelegatedSurface(
    'player-docs',
    mainWt,
    '.claude/skills/player-docs/scripts/check-freshness.mjs',
    ['--check', '--json'],
    (parsed) => (parsed.pages || []).filter((p) => p.verdict !== 'fresh').flatMap((p) => (p.sources || []).map((s) => s.path))
  );
}

function computeTuiDesignSurface(mainWt) {
  return computeDelegatedSurface(
    'tui-design',
    mainWt,
    'scripts/check-tui-design.mjs',
    ['--json'],
    (parsed) => (parsed.checks || []).map((c) => c.file).filter(Boolean)
  );
}

// ---------------------------------------------------------------------------
// Root state (§8)
// ---------------------------------------------------------------------------

function readRootState(mainWt, originMainTip) {
  const head = git(['symbolic-ref', '--short', 'HEAD'], { cwd: mainWt });
  const onMain = head.status === 0 && head.stdout.trim() === 'main';
  const mainRef = git(['rev-parse', 'main'], { cwd: mainWt });
  let behindBy = 0;
  let aheadBy = 0;
  if (mainRef.status === 0) {
    const mainSha = mainRef.stdout.trim();
    behindBy = revListCount(`${mainSha}..${originMainTip}`, mainWt);
    aheadBy = revListCount(`${originMainTip}..${mainSha}`, mainWt);
  }
  return {
    onMain,
    behindBy,
    aheadBy,
    fastForwarded: false,
    dirty: isDirty(mainWt),
  };
}

// Guarded ff-pull (session mode only, R8/§8): behind > 0, ahead == 0, clean.
function tryFastForwardRoot(mainWt, root) {
  if (root.onMain && root.behindBy > 0 && root.aheadBy === 0 && !root.dirty) {
    const r = git(['merge', '--ff-only', 'origin/main'], { cwd: mainWt });
    if (r.status === 0) {
      root.fastForwarded = true;
      root.behindBy = 0;
    }
  }
}

// ---------------------------------------------------------------------------
// Live-branch enumeration (§1)
// ---------------------------------------------------------------------------

function enumerateLiveBranches(cwd, originMainTip) {
  const wts = listWorktrees(cwd);
  const mainPath = wts.length ? wts[0].path : null;
  const worktreeBranches = new Map(); // branch -> worktree path
  for (const w of wts) {
    if (w.path === mainPath) continue;
    if (w.branch && w.branch !== 'main') worktreeBranches.set(w.branch, w.path);
  }

  const names = new Set(worktreeBranches.keys());
  const refsOut = gitOk(['for-each-ref', 'refs/heads', '--format=%(refname:short) %(objectname)'], { cwd });
  for (const line of refsOut.split('\n').filter(Boolean)) {
    const sp = line.indexOf(' ');
    const name = line.slice(0, sp);
    const sha = line.slice(sp + 1);
    if (name === 'main' || worktreeBranches.has(name)) continue;
    if (!isAncestor(sha, originMainTip, cwd)) names.add(name);
  }

  return [...names].sort().map((name) => ({ name, worktree: worktreeBranches.get(name) ?? null }));
}

function buildLiveBranch(entry, originMainTip, cwd) {
  const { name, worktree } = entry;
  const tipRes = git(['rev-parse', `refs/heads/${name}`], { cwd });
  const tip = tipRes.status === 0 ? tipRes.stdout.trim() : null;
  const mergeBase = tip ? mergeBaseOf(tip, originMainTip, cwd) : null;
  const lag = mergeBase ? revListCount(`${mergeBase}..${originMainTip}`, cwd) : 0;
  const dirty = worktree ? isDirty(worktree) : false;
  const files = tip && mergeBase ? changedFiles(mergeBase, tip, cwd) : [];
  const mtVsMain = tip ? mergeTree(originMainTip, tip, cwd) : { conflict: null, treeOid: null, files: [] };

  let cleanupEligible = false;
  let cleanupReason = null;
  if (worktree && tip) {
    const ancestor = isAncestor(tip, originMainTip, cwd);
    let emptyContribution = false;
    if (!ancestor && mtVsMain.conflict === false) {
      const mainTreeOid = gitOk(['rev-parse', `${originMainTip}^{tree}`], { cwd }).trim();
      emptyContribution = mtVsMain.treeOid === mainTreeOid;
    }
    cleanupReason = ancestor ? 'ancestor' : emptyContribution ? 'empty-contribution' : null;
    cleanupEligible = (ancestor || emptyContribution) && !dirty;
  }

  return {
    name,
    tip,
    worktree,
    task: attributeTask(name),
    mergeBase,
    baseLag: lag,
    dirty,
    changedFiles: files,
    cleanupEligible,
    cleanupReason,
    _mtVsMain: mtVsMain,
  };
}

// ---------------------------------------------------------------------------
// Report emission
// ---------------------------------------------------------------------------

function emitReport(report, json) {
  if (json) {
    // eslint-disable-next-line no-unused-vars
    const { branches, ...rest } = report;
    const cleanBranches = report.branches.map(({ _mtVsMain, ...b }) => b);
    process.stdout.write(JSON.stringify({ ...rest, branches: cleanBranches }, null, 2) + '\n');
    return;
  }

  const lines = [];
  lines.push(`check-merge-drift: mode=${report.mode} verdict=${report.verdict} originMain=${report.originMain}`);
  if (report.unverifiedAgainstRemote) {
    lines.push('WARNING: report is unverified against the remote (fetch failed or --no-fetch)');
  }
  const r = report.root;
  lines.push(
    `root: onMain=${r.onMain} behindBy=${r.behindBy} aheadBy=${r.aheadBy} dirty=${r.dirty} fastForwarded=${r.fastForwarded}`
  );
  if (report.branches.length) {
    lines.push('branches:');
    for (const b of report.branches) {
      lines.push(
        `  ${b.name}  task=${b.task || '-'}  baseLag=${b.baseLag}  dirty=${b.dirty}  cleanupEligible=${b.cleanupEligible}${
          b.cleanupReason ? ` (${b.cleanupReason})` : ''
        }`
      );
    }
  }
  if (report.matrix.length) {
    lines.push('matrix:');
    for (const m of report.matrix) {
      lines.push(`  ${m.a} x ${m.b}: ${m.conflict ? 'CONFLICT ' + m.files.join(',') : 'clean'}`);
    }
  }
  if (report.grounding.length) {
    lines.push('grounding:');
    for (const g of report.grounding) {
      lines.push(
        `  ${g.name} (${g.checker}): stale=${g.stale === null ? 'n/a' : g.stale}${
          g.touched.length ? ' touched=' + g.touched.join(',') : ''
        }`
      );
    }
  }
  if (report.findings.length) {
    lines.push('findings:');
    for (const f of report.findings) {
      lines.push(`  [${f.severity}] ${f.rule}: ${f.message}${f.task ? ' (task=' + f.task + ')' : ''}`);
      if (f.evidence.length) lines.push(`      evidence: ${f.evidence.join(', ')}`);
      if (f.noteWritten) lines.push('      note: written to board');
    }
  } else {
    lines.push('no findings');
  }
  if (report.cleanupPrescriptions.length) {
    lines.push('cleanup prescriptions:');
    for (const p of report.cleanupPrescriptions) {
      lines.push(`  ${p.branch}${p.applied ? ' [APPLIED]' : ''}: ${p.commands.join(' && ')}`);
    }
  }
  process.stdout.write(lines.join('\n') + '\n');
}

// ---------------------------------------------------------------------------
// --notes (FR-010, §7)
// ---------------------------------------------------------------------------

function backlogCliAvailable() {
  try {
    execFileSync(process.platform === 'win32' ? 'where' : 'which', ['backlog'], {
      stdio: ['ignore', 'pipe', 'pipe'],
    });
    return true;
  } catch {
    return false;
  }
}

function applyNotes(report, mainWt) {
  if (!backlogCliAvailable()) {
    report.findings.push(
      makeFinding({
        severity: 'info',
        gate: report.mode,
        rule: 'backlog-cli-absent',
        message: 'backlog CLI not found on PATH — findings were not recorded as board notes',
        evidence: [],
        task: null,
        key: 'backlog-cli',
      })
    );
    return;
  }
  for (const f of report.findings) {
    if (!f.task || f.severity === 'info') continue;
    const taskFilePath = findTaskFile(mainWt, f.task);
    if (!taskFilePath) continue;
    let existingText = '';
    try {
      existingText = readFileSync(taskFilePath, 'utf8');
    } catch {
      continue;
    }
    if (existingText.includes(f.fingerprint)) continue; // fingerprint dedup
    const noteText = `[merge-drift ${f.gate}] ${f.severity}: ${f.message}\nevidence: ${f.evidence.join(
      ', '
    )}\nfingerprint: ${f.fingerprint}`;
    const editRes = tryExec('backlog', ['task', 'edit', f.task, '--append-notes', noteText], { cwd: mainWt });
    if (editRes.status === 0) f.noteWritten = true;
  }
}

// ---------------------------------------------------------------------------
// Mode: session
// ---------------------------------------------------------------------------

function runSession(flags, cwd) {
  const findings = [];
  const mainWt = resolveMainWorktree(cwd);
  let unverifiedAgainstRemote = false;

  if (flags.noFetch) {
    unverifiedAgainstRemote = true;
    findings.push(
      makeFinding({
        severity: 'warn',
        gate: 'session',
        rule: 'remote-unverified',
        message: 'session run with --no-fetch — report is unverified against the remote',
        evidence: [],
        task: null,
        key: 'remote',
      })
    );
  } else {
    const f = git(['fetch', 'origin'], { cwd: mainWt });
    if (f.status !== 0) {
      unverifiedAgainstRemote = true;
      findings.push(
        makeFinding({
          severity: 'warn',
          gate: 'session',
          rule: 'remote-unverified',
          message: `git fetch origin failed (${(f.stderr || f.stdout).trim().split('\n')[0]}) — report is unverified against the remote`,
          evidence: [],
          task: null,
          key: 'remote',
        })
      );
    }
  }

  const originMainRes = git(['rev-parse', 'origin/main'], { cwd: mainWt });
  if (originMainRes.status !== 0) {
    usageError('origin/main is not resolvable locally — fetch at least once before running in degraded mode');
  }
  const originMainTip = originMainRes.stdout.trim();

  const root = readRootState(mainWt, originMainTip);
  if (!root.onMain) {
    findings.push(
      makeFinding({
        severity: 'block',
        gate: 'session',
        rule: 'root-not-main',
        message: `root checkout at ${mainWt} is not on main — the worktree doctrine requires root to stay on main`,
        evidence: [mainWt],
        task: null,
        key: 'root',
      })
    );
  } else {
    tryFastForwardRoot(mainWt, root);
    if (root.behindBy > 0) {
      findings.push(
        makeFinding({
          severity: 'info',
          gate: 'session',
          rule: 'root-stale',
          message: `root is ${root.behindBy} commit(s) behind origin/main (fast-forward did not apply — diverged or dirty)`,
          evidence: [],
          task: null,
          key: 'root',
        })
      );
    }
  }
  if (root.dirty) {
    findings.push(
      makeFinding({
        severity: 'info',
        gate: 'session',
        rule: 'dirty-worktree',
        message: 'root checkout has uncommitted changes',
        evidence: [mainWt],
        task: null,
        key: 'root',
      })
    );
  }

  const liveEntries = enumerateLiveBranches(mainWt, originMainTip);
  const branches = liveEntries.map((e) => buildLiveBranch(e, originMainTip, mainWt));
  const takenSpecs = takenSpecNumbers(originMainTip, mainWt);

  const cleanupPrescriptions = [];
  for (const b of branches) {
    if (b.dirty) {
      findings.push(
        makeFinding({
          severity: 'info',
          gate: 'session',
          rule: 'dirty-worktree',
          message: `${b.name} has uncommitted changes — excluded from cleanup eligibility`,
          evidence: [b.worktree || b.name],
          task: b.task,
          key: b.name,
        })
      );
    }
    if (b.baseLag > 0) {
      findings.push(
        makeFinding({
          severity: 'info',
          gate: 'session',
          rule: 'stale-base',
          message: `${b.name}'s base lags origin/main by ${b.baseLag} commit(s)`,
          evidence: [b.name],
          task: b.task,
          key: b.name,
        })
      );
    }
    // Claim-protocol audit (spec 065 FR-005): a local task branch with
    // commits but no remote counterpart is invisible from every other clone.
    if (/^task-\d+-/.test(b.name) && b.tip && b.mergeBase) {
      const aheadCount = revListCount(`${b.mergeBase}..${b.tip}`, mainWt);
      const remoteRef = git(['rev-parse', '--verify', '--quiet', `refs/remotes/origin/${b.name}`], { cwd: mainWt });
      if (aheadCount > 0 && remoteRef.status !== 0) {
        findings.push(
          makeFinding({
            severity: 'warn',
            gate: 'session',
            rule: 'branch-unpushed',
            message: `${b.name} has ${aheadCount} local commit(s) but no remote counterpart — not auditable from other clones; run git push -u origin ${b.name}`,
            evidence: [b.name],
            task: b.task,
            key: b.name,
          })
        );
      }
    }
    if (b.mergeBase) {
      const mainFiles = changedFiles(b.mergeBase, originMainTip, mainWt);
      const overlap = backlogOverlap(b.changedFiles, mainFiles);
      if (overlap.length) {
        findings.push(
          makeFinding({
            severity: 'warn',
            gate: 'session',
            rule: 'backlog-overlap',
            message: `${b.name} and origin/main both touch backlog/ file(s): ${overlap.join(', ')}`,
            evidence: [...overlap, b.name],
            task: b.task,
            key: b.name,
          })
        );
      }
      for (const c of specNumberCollisions(b.changedFiles, takenSpecs)) {
        findings.push(
          makeFinding({
            severity: 'warn',
            gate: 'session',
            rule: 'spec-number-collision',
            message: `${b.name} adds specs/${c.newDir} but origin/main already has specs/${c.takenDir} for number ${formatSpecNum(
              c.number
            )}`,
            evidence: [`specs/${c.newDir}`, `specs/${c.takenDir}`],
            task: b.task,
            key: b.name,
          })
        );
      }
    }
    if (b.cleanupEligible) {
      // `-d` (safe delete) only succeeds when git's own ancestor check agrees
      // the branch is merged; the empty-contribution (squash) case is
      // precisely the situation where it never will, because git has no way
      // to see that the squash commit subsumes it — we already proved that
      // via the merge-tree tree-equality check (§4), so force delete is safe
      // here specifically. Deviation from the contract's literal `-d`
      // (detection-rules.md §4): reported to the planning tier.
      const deleteFlag = b.cleanupReason === 'empty-contribution' ? '-D' : '-d';
      const commands = [`git worktree remove ${b.worktree}`, `git branch ${deleteFlag} ${b.name}`];
      findings.push(
        makeFinding({
          severity: 'warn',
          gate: 'session',
          rule: 'cleanup-eligible',
          message: `${b.name} is cleanup-eligible (${b.cleanupReason}): ${commands.join(' && ')}`,
          evidence: [b.worktree, b.name],
          task: b.task,
          key: b.name,
        })
      );
      cleanupPrescriptions.push({ branch: b.name, commands, applied: false, _worktree: b.worktree, _deleteFlag: deleteFlag });
    }
  }

  // Drift matrix: each branch vs origin/main, plus every unordered pair.
  const matrix = [];
  for (const b of branches) {
    const mt = b._mtVsMain;
    matrix.push({ a: b.name, b: 'origin/main', conflict: !!mt.conflict, files: mt.files || [] });
    if (mt.conflict) {
      findings.push(
        makeFinding({
          severity: 'warn',
          gate: 'session',
          rule: 'textual-conflict',
          message: `${b.name} would conflict with origin/main on ${mt.files.join(', ')}`,
          evidence: [...mt.files, b.name],
          task: b.task,
          key: b.name,
        })
      );
    }
  }
  const sortedNames = branches.map((b) => b.name).sort();
  const byName = new Map(branches.map((b) => [b.name, b]));
  for (let i = 0; i < sortedNames.length; i++) {
    for (let j = i + 1; j < sortedNames.length; j++) {
      const A = byName.get(sortedNames[i]);
      const B = byName.get(sortedNames[j]);
      const mt = mergeTree(A.tip, B.tip, mainWt);
      matrix.push({ a: A.name, b: B.name, conflict: !!mt.conflict, files: mt.files || [] });
      if (mt.conflict) {
        findings.push(
          makeFinding({
            severity: 'warn',
            gate: 'session',
            rule: 'pairwise-conflict',
            message: `${A.name} and ${B.name} will conflict on ${mt.files.join(', ')} whichever merges first`,
            evidence: [...mt.files, A.name, B.name],
            task: A.task,
            key: A.name,
          })
        );
      }
    }
  }
  matrix.sort((x, y) => (x.a === y.a ? (x.b < y.b ? -1 : x.b > y.b ? 1 : 0) : x.a < y.a ? -1 : 1));

  // Grounding freshness (FR-008, §6).
  const wikiNotes = loadWikiNotes(mainWt, originMainTip);
  const wikiResult = computeWikiGrounding(wikiNotes, mainWt, originMainTip);
  findings.push(...wikiResult.findings);
  const playerSurface = computePlayerDocsSurface(mainWt);
  const tuiSurface = computeTuiDesignSurface(mainWt);
  if (playerSurface.stale) {
    findings.push(
      makeFinding({
        severity: 'warn',
        gate: 'session',
        rule: 'grounding-stale',
        message: 'player-docs freshness check reports stale content — run its checker and refresh',
        evidence: [...playerSurface.touched],
        task: null,
        key: 'player-docs',
      })
    );
  }
  if (tuiSurface.stale) {
    findings.push(
      makeFinding({
        severity: 'warn',
        gate: 'session',
        rule: 'grounding-stale',
        message: 'tui-design freshness check reports stale content — run node scripts/check-tui-design.mjs',
        evidence: [...tuiSurface.touched],
        task: null,
        key: 'tui-design',
      })
    );
  }
  const grounding = [wikiResult.surface, playerSurface, tuiSurface];

  if (flags.applyCleanup) {
    for (const p of cleanupPrescriptions) {
      const rmRes = git(['worktree', 'remove', p._worktree], { cwd: mainWt });
      if (rmRes.status !== 0) {
        findings.push(
          makeFinding({
            severity: 'warn',
            gate: 'session',
            rule: 'cleanup-apply-failed',
            message: `git worktree remove ${p._worktree} failed: ${(rmRes.stderr || rmRes.stdout).trim()}`,
            evidence: [p._worktree, p.branch],
            task: attributeTask(p.branch),
            key: p.branch,
          })
        );
        continue;
      }
      const brRes = git(['branch', p._deleteFlag, p.branch], { cwd: mainWt });
      if (brRes.status !== 0) {
        findings.push(
          makeFinding({
            severity: 'warn',
            gate: 'session',
            rule: 'cleanup-apply-failed',
            message: `git branch ${p._deleteFlag} ${p.branch} failed: ${(brRes.stderr || brRes.stdout).trim()}`,
            evidence: [p._worktree, p.branch],
            task: attributeTask(p.branch),
            key: p.branch,
          })
        );
        continue;
      }
      p.applied = true;
    }
  }

  const sorted = sortFindings(findings);
  const verdict = computeVerdict(sorted);
  return {
    mode: 'session',
    verdict,
    exitCode: exitCodeFor(verdict),
    originMain: originMainTip,
    unverifiedAgainstRemote,
    root,
    branches,
    matrix,
    grounding,
    findings: sorted,
    cleanupPrescriptions: cleanupPrescriptions.map(({ _worktree, _deleteFlag, ...p }) => p),
  };
}

// ---------------------------------------------------------------------------
// Mode: claim (spec 065)
// ---------------------------------------------------------------------------

// Blocks authoring against a spec number already taken on origin/main — at
// directory-creation (claim) time, before any content accumulates. Runs from
// anywhere inside the repo. Idempotent for the owner: claiming a dirname
// already on origin/main under the SAME name passes.
function runClaim(flags, cwd) {
  const f = git(['fetch', 'origin'], { cwd });
  if (f.status !== 0) {
    process.stderr.write(
      `check-merge-drift: cannot reach remote (${(f.stderr || f.stdout).trim().split('\n')[0]}) — claim gate fails closed\n`
    );
    process.exit(2);
  }
  const originMainTip = gitOk(['rev-parse', 'origin/main'], { cwd }).trim();
  const mainWt = resolveMainWorktree(cwd);
  const root = readRootState(mainWt, originMainTip);

  const findings = [];
  const taken = takenSpecNumbers(originMainTip, mainWt);
  const number = parseInt(flags.dir.match(/^(\d+)-/)[1], 10);
  if (taken.has(number) && taken.get(number) !== flags.dir) {
    const takenDir = taken.get(number);
    const nextFree = Math.max(...taken.keys()) + 1;
    findings.push(
      makeFinding({
        severity: 'block',
        gate: 'claim',
        rule: 'spec-number-collision',
        message: `specs/${takenDir} already exists on origin/main for number ${formatSpecNum(
          number
        )} — claim specs/${flags.dir} is a collision; next free number is ${formatSpecNum(nextFree)}`,
        evidence: [`specs/${takenDir}`],
        task: attributeBySpecDir(mainWt, takenDir),
        key: 'claim',
      })
    );
  }

  const sorted = sortFindings(findings);
  const verdict = computeVerdict(sorted);
  return {
    mode: 'claim',
    verdict,
    exitCode: exitCodeFor(verdict),
    originMain: originMainTip,
    unverifiedAgainstRemote: false,
    root,
    branches: [],
    matrix: [],
    grounding: [],
    findings: sorted,
    cleanupPrescriptions: [],
  };
}

// ---------------------------------------------------------------------------
// Mode: worktree
// ---------------------------------------------------------------------------

function runWorktree(flags, cwd) {
  const f = git(['fetch', 'origin'], { cwd });
  if (f.status !== 0) {
    process.stderr.write(
      `check-merge-drift: cannot reach remote (${(f.stderr || f.stdout).trim().split('\n')[0]}) — worktree gate fails closed\n`
    );
    process.exit(2);
  }
  const originMainTip = gitOk(['rev-parse', 'origin/main'], { cwd }).trim();
  const mainWt = resolveMainWorktree(cwd);
  const root = readRootState(mainWt, originMainTip);

  const findings = [];
  if (!root.onMain) {
    findings.push(
      makeFinding({
        severity: 'block',
        gate: 'worktree',
        rule: 'root-not-main',
        message: `root checkout at ${mainWt} is not on main — fast-forward and switch back before cutting a worktree`,
        evidence: [mainWt],
        task: null,
        key: 'root',
      })
    );
  } else if (root.behindBy > 0 || root.aheadBy > 0) {
    findings.push(
      makeFinding({
        severity: 'block',
        gate: 'worktree',
        rule: 'root-stale',
        message: `root is not exactly at the fetched origin/main tip (behindBy=${root.behindBy}, aheadBy=${root.aheadBy}) — run git pull --ff-only first`,
        evidence: [],
        task: null,
        key: 'root',
      })
    );
  }

  if (flags.spec) {
    const taken = takenSpecNumbers(originMainTip, mainWt);
    const n = parseInt(flags.spec, 10);
    if (taken.has(n)) {
      const takenDir = taken.get(n);
      // Claim-aware (spec 065): under claim-first doctrine the spec dir
      // legitimately exists on origin/main BEFORE the worktree is cut —
      // ownership, not absence, is the invariant. With --task, pass when the
      // taken dir's Spec marker (read from the origin/main tree) attributes
      // to that task; block when it attributes elsewhere or to none.
      const owner = flags.task ? attributeBySpecDirOnOriginMain(originMainTip, takenDir, mainWt) : null;
      if (!(flags.task && owner === flags.task)) {
        const maxTaken = Math.max(...taken.keys());
        const nextFree = maxTaken + 1;
        const ownership = flags.task
          ? ` (attributed to ${owner || 'no task'}, not ${flags.task})`
          : '';
        findings.push(
          makeFinding({
            severity: 'block',
            gate: 'worktree',
            rule: 'spec-number-collision',
            message: `specs/${takenDir} already exists on origin/main for number ${formatSpecNum(n)}${ownership} — next free number is ${formatSpecNum(
              nextFree
            )}`,
            evidence: [`specs/${takenDir}`],
            task: owner || attributeBySpecDir(mainWt, takenDir),
            key: 'spec',
          })
        );
      }
    }
  }

  // Card-claim check (spec 065 FR-003): the task's board card must be
  // In Progress on origin/main before its worktree is cut — warn-only by
  // design, because the claim commit itself needs a pre-propagation window.
  if (flags.task) {
    const card = cardStatusOnOriginMain(originMainTip, flags.task, mainWt);
    if (card.status !== 'In Progress') {
      const observed = card.path
        ? card.status
          ? `status is '${card.status}'`
          : 'card frontmatter is unparseable'
        : 'card file is missing';
      findings.push(
        makeFinding({
          severity: 'warn',
          gate: 'worktree',
          rule: 'card-not-claimed',
          message: `${flags.task}'s board card is not In Progress on origin/main (${observed}) — claim missing or unpushed; claim-before-work: the first pushed commit moves the card to In Progress and creates the spec dir (spec 065)`,
          evidence: [card.path || 'backlog/tasks/'],
          task: flags.task,
          key: flags.task,
        })
      );
    }
  }

  const sorted = sortFindings(findings);
  const verdict = computeVerdict(sorted);
  return {
    mode: 'worktree',
    verdict,
    exitCode: exitCodeFor(verdict),
    originMain: originMainTip,
    unverifiedAgainstRemote: false,
    root,
    branches: [],
    matrix: [],
    grounding: [],
    findings: sorted,
    cleanupPrescriptions: [],
  };
}

// ---------------------------------------------------------------------------
// Mode: pr
// ---------------------------------------------------------------------------

function runPr(flags, cwd) {
  let branchName = flags.branch;
  if (!branchName) {
    const h = git(['symbolic-ref', '--short', 'HEAD'], { cwd });
    if (h.status !== 0) usageError('cannot determine current branch (detached HEAD?) — pass --branch <name>');
    branchName = h.stdout.trim();
  }
  if (branchName === 'main') {
    usageError('pr mode must run from (or name, via --branch) a task branch, not main');
  }

  const f = git(['fetch', 'origin'], { cwd });
  if (f.status !== 0) {
    process.stderr.write(
      `check-merge-drift: cannot reach remote (${(f.stderr || f.stdout).trim().split('\n')[0]}) — pr gate fails closed\n`
    );
    process.exit(2);
  }
  const originMainTip = gitOk(['rev-parse', 'origin/main'], { cwd }).trim();

  const tipRes = git(['rev-parse', `refs/heads/${branchName}`], { cwd });
  if (tipRes.status !== 0) usageError(`unknown branch ${JSON.stringify(branchName)}`);
  const tip = tipRes.stdout.trim();

  const mainWt = resolveMainWorktree(cwd);
  const root = readRootState(mainWt, originMainTip);

  const findings = [];
  if (!root.onMain) {
    findings.push(
      makeFinding({
        severity: 'block',
        gate: 'pr',
        rule: 'root-not-main',
        message: `root checkout at ${mainWt} is not on main`,
        evidence: [mainWt],
        task: null,
        key: 'root',
      })
    );
  }

  const task = attributeTask(branchName);
  const mergeBase = mergeBaseOf(tip, originMainTip, cwd);
  const lag = mergeBase ? revListCount(`${mergeBase}..${originMainTip}`, cwd) : 0;
  const branchFiles = mergeBase ? changedFiles(mergeBase, tip, cwd) : [];
  const mainFiles = mergeBase ? changedFiles(mergeBase, originMainTip, cwd) : [];
  const worktreePath = findWorktreeForBranch(cwd, branchName);
  const dirty = worktreePath ? isDirty(worktreePath) : false;

  const mt = mergeTree(originMainTip, tip, cwd);
  if (mt.conflict === null) {
    usageError(`git merge-tree failed while gating ${branchName}: ${mt.error}`);
  }
  if (mt.conflict) {
    for (const file of mt.files) {
      findings.push(
        makeFinding({
          severity: 'block',
          gate: 'pr',
          rule: 'textual-conflict',
          message: `${file} conflicts with origin/main`,
          evidence: [file, branchName],
          task,
          key: branchName,
        })
      );
    }
  }

  if (lag > 0) {
    findings.push(
      makeFinding({
        severity: 'warn',
        gate: 'pr',
        rule: 'stale-base',
        message: `branch base lags origin/main by ${lag} commit(s)`,
        evidence: [branchName],
        task,
        key: branchName,
      })
    );
  }

  const overlap = backlogOverlap(branchFiles, mainFiles);
  if (overlap.length) {
    findings.push(
      makeFinding({
        severity: 'warn',
        gate: 'pr',
        rule: 'backlog-overlap',
        message: `${branchName} and origin/main both touch backlog/ file(s): ${overlap.join(', ')}`,
        evidence: [...overlap, branchName],
        task,
        key: branchName,
      })
    );
  }

  const wikiNotes = loadWikiNotes(cwd, originMainTip);
  const wikiHits = wikiSourcesOverlap(branchFiles, wikiNotes);
  for (const [notePath, matched] of wikiHits) {
    findings.push(
      makeFinding({
        severity: 'warn',
        gate: 'pr',
        rule: 'wiki-sources-overlap',
        message: `branch touches source(s) pinned by ${notePath}: ${matched.join(
          ', '
        )} — re-check via /grounding-wiki:wiki-update`,
        evidence: [notePath, ...matched],
        task,
        key: branchName,
      })
    );
  }

  const tuiFiles = tuiSurfaceFiles(branchFiles);
  if (tuiFiles.length) {
    findings.push(
      makeFinding({
        severity: 'warn',
        gate: 'pr',
        rule: 'tui-surface',
        message:
          'branch touches internal/tui/ — run node scripts/check-tui-design.mjs --changed and amend docs/design/tui/ in this PR',
        evidence: tuiFiles,
        task,
        key: branchName,
      })
    );
  }

  const taken = takenSpecNumbers(originMainTip, cwd);
  for (const c of specNumberCollisions(branchFiles, taken)) {
    findings.push(
      makeFinding({
        severity: 'warn',
        gate: 'pr',
        rule: 'spec-number-collision',
        message: `branch adds specs/${c.newDir} but origin/main already has specs/${c.takenDir} for number ${formatSpecNum(
          c.number
        )}`,
        evidence: [`specs/${c.newDir}`, `specs/${c.takenDir}`],
        task,
        key: branchName,
      })
    );
  }

  if (dirty) {
    findings.push(
      makeFinding({
        severity: 'info',
        gate: 'pr',
        rule: 'dirty-worktree',
        message: `${branchName}'s worktree has uncommitted changes`,
        evidence: [worktreePath || branchName],
        task,
        key: branchName,
      })
    );
  }

  const branchEntry = {
    name: branchName,
    tip,
    worktree: worktreePath,
    task,
    mergeBase,
    baseLag: lag,
    dirty,
    changedFiles: branchFiles,
    cleanupEligible: false,
    cleanupReason: null,
  };

  const sorted = sortFindings(findings);
  const verdict = computeVerdict(sorted);
  return {
    mode: 'pr',
    verdict,
    exitCode: exitCodeFor(verdict),
    originMain: originMainTip,
    unverifiedAgainstRemote: false,
    root,
    branches: [branchEntry],
    matrix: [],
    grounding: [],
    findings: sorted,
    cleanupPrescriptions: [],
  };
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

function main() {
  const flags = parseArgs(process.argv.slice(2));
  preflight();
  const cwd = process.cwd();

  let report;
  if (flags.mode === 'session') report = runSession(flags, cwd);
  else if (flags.mode === 'claim') report = runClaim(flags, cwd);
  else if (flags.mode === 'worktree') report = runWorktree(flags, cwd);
  else report = runPr(flags, cwd);

  if (flags.notes) {
    const mainWt = resolveMainWorktree(cwd);
    applyNotes(report, mainWt);
    // Recompute exitCode-relevant severities are unaffected by note-writing;
    // only noteWritten flags change, and verdict/exitCode already reflect
    // the findings that drove them.
  }

  emitReport(report, flags.json);
  process.exit(report.exitCode);
}

// Guard so this file is importable (e.g. from a node:test regression test)
// without spawning the CLI's side effects — only run main() when invoked
// directly as `node check-merge-drift.mjs ...`. Default is to RUN: skipping
// requires positive proof this is an import, via resolved real paths (so a
// symlinked invocation — e.g. through a worktree reached via a symlinked
// ancestor directory — still matches). If we can't tell, run anyway; a gate
// script must never silently no-op.
function invokedDirectly() {
  try {
    return realpathSync(process.argv[1]) === realpathSync(fileURLToPath(import.meta.url));
  } catch {
    return true;
  }
}
if (invokedDirectly()) {
  main();
}

#!/usr/bin/env node
// check-tui-design.mjs — read-only structural validator + same-PR gate for
// docs/design/tui/, the living TUI design reference.
//
// Contract: specs/047-tui-design-reference-v2/contracts/check-script.md
//
// Node >= 18, ESM, zero npm dependencies (stdlib fs/path/child_process only).
// This script NEVER writes any file. It only reads and reports — pins are
// edited by humans/authors, never rewritten here.
//
// Usage:
//   node scripts/check-tui-design.mjs [--json] [--changed [<range>]]
//
//   no flags        structural checks only: file-set, pins, control-tables,
//                    anatomy.
//   --changed [<r>] additionally run the same-PR gate over git range <r>
//                    (default origin/main...HEAD).
//   --json          emit the report as JSON instead of human-readable lines.
//
// Exit codes:
//   0  all checks pass
//   1  >= 1 violation (each reported with file + actionable message)
//   2  usage/environment error (bad flag, not a git repo, docs/design/tui
//      unreadable, a bad --changed range)

import { existsSync, readFileSync, statSync, readdirSync } from 'node:fs';
import { join, relative, sep as pathSep } from 'node:path';
import { execFileSync } from 'node:child_process';

// ---------------------------------------------------------------------------
// Constants (contracts/control-table.md, contracts/frontmatter-and-pins.md)
// ---------------------------------------------------------------------------

const CANONICAL_HEADER =
  '| control/region | states | data source | renderer | keys+mouse | introduced-by | skin-token |';

// directory (under docs/design/tui/) -> required frontmatter `class` value
const CLASS_BY_DIR = {
  pages: 'page',
  panels: 'panel',
  overlays: 'overlay',
  patterns: 'pattern',
};

const HEX40 = /^[0-9a-f]{40}$/;

// ---------------------------------------------------------------------------
// Argument parsing
// ---------------------------------------------------------------------------

function usageError(msg) {
  process.stderr.write(`check-tui-design: ${msg}\n`);
  process.exit(2);
}

function parseArgs(argv) {
  const flags = { json: false, changed: false, range: 'origin/main...HEAD' };
  for (let i = 0; i < argv.length; i++) {
    const arg = argv[i];
    if (arg === '--json') {
      flags.json = true;
      continue;
    }
    if (arg === '--changed') {
      flags.changed = true;
      const next = argv[i + 1];
      if (next !== undefined && !next.startsWith('--')) {
        flags.range = next;
        i++;
      }
      continue;
    }
    usageError(`unknown flag ${JSON.stringify(arg)}`);
  }
  return flags;
}

// ---------------------------------------------------------------------------
// Repo/file plumbing
// ---------------------------------------------------------------------------

function resolveRepoRoot() {
  try {
    return execFileSync('git', ['rev-parse', '--show-toplevel'], {
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'pipe'],
    }).trim();
  } catch {
    usageError('not inside a git repository');
  }
}

function walkMarkdown(dir) {
  let out = [];
  for (const entry of readdirSync(dir, { withFileTypes: true })) {
    const abs = join(dir, entry.name);
    if (entry.isDirectory()) {
      out = out.concat(walkMarkdown(abs));
    } else if (entry.isFile() && entry.name.endsWith('.md')) {
      out.push(abs);
    }
  }
  return out;
}

function parseFrontmatter(text) {
  const m = text.match(/^---\r?\n([\s\S]*?)\r?\n---/);
  if (!m) return null;
  const body = m[1];
  const get = (key) => {
    const km = body.match(new RegExp(`^${key}:\\s*(.+?)\\s*$`, 'm'));
    return km ? km[1].trim() : null;
  };
  return {
    title: get('title'),
    class: get('class'),
    status: get('status'),
    verified_against: get('verified_against'),
  };
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

function main() {
  const flags = parseArgs(process.argv.slice(2));
  const repoRoot = resolveRepoRoot();
  const tuiDir = join(repoRoot, 'docs/design/tui');

  if (!existsSync(tuiDir) || !statSync(tuiDir).isDirectory()) {
    usageError('docs/design/tui is unreadable');
  }

  const violations = [];
  function violate(check, file, message) {
    violations.push({ check, status: 'violation', file, message });
  }

  // Read every .md file once.
  const files = walkMarkdown(tuiDir).sort();
  const parsed = new Map(); // relPath ("docs/design/tui/...") -> { text }
  for (const abs of files) {
    const rel =
      'docs/design/tui/' + relative(tuiDir, abs).split(pathSep).join('/');
    let text = null;
    try {
      text = readFileSync(abs, 'utf8');
    } catch (err) {
      if (err && err.code === 'EACCES') {
        process.stderr.write(`check-tui-design: cannot read ${abs}: ${err.message}\n`);
        process.exit(2);
      }
    }
    parsed.set(rel, { text });
  }

  // --- check: file-set ---
  for (const [rel, { text }] of parsed) {
    const relInDir = rel.replace(/^docs\/design\/tui\//, '');
    const parts = relInDir.split('/');
    const fm = text ? parseFrontmatter(text) : null;

    if (parts.length === 1) {
      const isIndexFile = relInDir === 'INDEX.md' || relInDir === 'anatomy.md';
      if (!isIndexFile) {
        violate(
          'file-set',
          rel,
          'unexpected top-level file (only INDEX.md and anatomy.md may live at the docs/design/tui root)'
        );
        continue;
      }
      if (!fm || fm.class !== 'index') {
        violate(
          'file-set',
          rel,
          `class must be "index" for a top-level file, got ${fm ? JSON.stringify(fm.class) : 'no frontmatter'}`
        );
      }
    } else if (parts.length === 2 && CLASS_BY_DIR[parts[0]]) {
      const expected = CLASS_BY_DIR[parts[0]];
      if (!fm || fm.class !== expected) {
        violate(
          'file-set',
          rel,
          `class must be ${JSON.stringify(expected)} for a file under ${parts[0]}/, got ${fm ? JSON.stringify(fm.class) : 'no frontmatter'}`
        );
      }
    } else {
      violate(
        'file-set',
        rel,
        'file lives outside the taxonomy (pages/panels/overlays/patterns) or is nested too deeply'
      );
    }
  }

  // --- check: pins ---
  for (const [rel, { text }] of parsed) {
    if (!text) {
      violate('pins', rel, 'file unreadable');
      continue;
    }
    const fm = parseFrontmatter(text);
    if (!fm) {
      violate('pins', rel, 'missing frontmatter');
      continue;
    }
    const pin = fm.verified_against;
    if (!pin) {
      violate('pins', rel, 'verified_against missing');
      continue;
    }
    if (!HEX40.test(pin)) {
      violate('pins', rel, `verified_against malformed (not 40-hex): ${JSON.stringify(pin)}`);
      continue;
    }
    try {
      execFileSync('git', ['cat-file', '-e', `${pin}^{commit}`], {
        cwd: repoRoot,
        stdio: ['ignore', 'pipe', 'pipe'],
      });
    } catch {
      violate('pins', rel, `verified_against unresolvable: ${pin} is not a known commit`);
    }
  }

  // --- check: control-tables (panels/*, overlays/* only) ---
  for (const [rel, { text }] of parsed) {
    if (!/^docs\/design\/tui\/(panels|overlays)\//.test(rel)) continue;
    if (!text) continue;

    const lines = text.split(/\r?\n/);
    let canonicalHits = 0;
    let nonCanonicalHits = 0;

    for (let i = 0; i < lines.length - 1; i++) {
      const line = lines[i].trim();
      if (!line.startsWith('|')) continue;
      const nextLine = lines[i + 1].trim();
      if (!/^\|[\s:|-]+\|$/.test(nextLine)) continue; // not a header/separator pair

      const cols = line.split('|').slice(1, -1); // strip the leading/trailing empty cells
      if (cols.length !== 7) continue; // not control-table-shaped; ignore (e.g. a 3-col table)

      const normalized = '| ' + cols.map((c) => c.trim()).join(' | ') + ' |';
      if (normalized === CANONICAL_HEADER) canonicalHits++;
      else nonCanonicalHits++;
    }

    if (canonicalHits === 0 && nonCanonicalHits === 0) {
      violate('control-tables', rel, 'missing: no control table found');
    } else if (canonicalHits === 0) {
      violate(
        'control-tables',
        rel,
        'non-canonical header: a 7-column table exists but its header does not match the canonical column set'
      );
    } else if (canonicalHits > 1 || nonCanonicalHits > 0) {
      violate(
        'control-tables',
        rel,
        `duplicate: found ${canonicalHits} canonical-header table(s) and ${nonCanonicalHits} non-canonical control-shaped table(s) (expected exactly one canonical table)`
      );
    }
  }

  // --- check: anatomy (both directions) ---
  const anatomyRel = 'docs/design/tui/anatomy.md';
  const anatomyEntry = parsed.get(anatomyRel);
  if (!anatomyEntry || !anatomyEntry.text) {
    violate('anatomy', anatomyRel, 'anatomy.md missing or unreadable');
  } else {
    // Owning-file cells are backtick-quoted relative paths, e.g. `pages/home.md`.
    const refRe = /`([a-zA-Z0-9_-]+(?:\/[a-zA-Z0-9_-]+)*\.md)`/g;
    const targets = new Set();
    let m;
    while ((m = refRe.exec(anatomyEntry.text))) targets.add(m[1]);

    for (const t of targets) {
      const relTarget = `docs/design/tui/${t}`;
      if (!parsed.has(relTarget)) {
        violate('anatomy', anatomyRel, `dangling target: ${t} does not exist`);
      }
    }

    for (const [rel] of parsed) {
      if (!/^docs\/design\/tui\/(pages|panels|overlays)\//.test(rel)) continue;
      const short = rel.replace(/^docs\/design\/tui\//, '');
      if (!targets.has(short)) {
        violate('anatomy', rel, 'not mapped by any anatomy.md row');
      }
    }
  }

  // --- check: same-pr (only with --changed) ---
  if (flags.changed) {
    let changedFiles = [];
    try {
      const out = execFileSync('git', ['diff', '--name-only', flags.range], {
        cwd: repoRoot,
        encoding: 'utf8',
        stdio: ['ignore', 'pipe', 'pipe'],
      });
      changedFiles = out.split('\n').filter(Boolean);
    } catch (err) {
      usageError(`git diff failed for range ${JSON.stringify(flags.range)}: ${err.message}`);
    }

    const tuiTouched = changedFiles.filter((f) => f.startsWith('internal/tui/'));
    const designTouched = changedFiles.some((f) => f.startsWith('docs/design/tui/'));

    if (tuiTouched.length > 0 && !designTouched) {
      violate(
        'same-pr',
        tuiTouched.join(', '),
        'amend docs/design/tui in this PR (re-verify + re-pin affected pages)'
      );
    }
  }

  // --- report ---
  const ok = violations.length === 0;

  if (flags.json) {
    process.stdout.write(JSON.stringify({ ok, checks: violations }, null, 2) + '\n');
  } else if (ok) {
    process.stdout.write('check-tui-design: all checks passed\n');
  } else {
    for (const v of violations) {
      process.stdout.write(`${v.check}  ${v.file}  ${v.message}\n`);
    }
    process.stdout.write(`${violations.length} violation(s)\n`);
  }

  process.exit(ok ? 0 : 1);
}

main();

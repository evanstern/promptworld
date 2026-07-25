---
title: Pattern — skin tokens (documentation conventions)
class: pattern
status: specified
verified_against: c8906da39be3a5b861c2272af37db0a83dcded7a
---

# Pattern: skin tokens

**Documentation conventions only.** This page fixes how the v2 reference itself
writes fiction strings so it never accumulates new bare literals ahead of the
runtime skin contract (reorientation D2 sequencing). The runtime token *lookup*
— file format, resolution order, fallback behavior — is TASK-121's spec to
design; this page is deliberately silent on it and MUST be adopted or amended
by that spec's own PR once it lands.

## Why this exists now

The reorientation renames the fiction-bearing dock tab and its family of new
surfaces around the generic word **guardian** (`panels/guardian.md`,
`pages/guardian-console.md`, `panels/guardian-strip.md` — never
`metatron.md`), because the *displayed* proper name is skin data, not a fixed
identifier. Today's shipped client hard-codes the angel-fiction skin
(`internal/tui` prints the default proper name and epithet literally — see
the token index below); the design reference
speaks in the generic vocabulary throughout and marks every place a fiction
string would render with a token instead, so a future skin swap is a content
change, never a doc rewrite.

## Conventions

1. **Mockups** (fenced ASCII blocks): every fiction string is written as
   `{{skin.<domain>.<name>}}` — double-brace, the token's dotted path, no
   quoting. Example: a dock tab row renders `{{skin.guardian.tab_label}}`
   where today's shipped UI shows literal `metatron`/`METATRON`.
2. **Control tables**: the `skin-token` column (contracts/control-table.md)
   names the token in unbraced dotted form — `skin.guardian.name` — or `—` for
   non-fiction controls (telemetry rows, chrome structure, keys).
3. **Prose**: describes the *role*, never the fiction word — "the guardian
   tab", "the guardian's proper name" — except when quoting exactly what a
   mockup or control table cell shows, which follows rules 1–2.
4. **Case is a rendering detail, not a token detail.** One token covers every
   case-transformed rendering of the same string (e.g. the dock tab's
   lowercase-inactive / uppercase-active styling) — the page that uses it says
   so in prose once; the token itself never forks by case.
5. **Non-fiction strings** (chrome labels, telemetry, key hints, glyphs) are
   literal text with `—` in the skin-token column — most of the reference.
   Systems-tab content (panels/systems.md) is never skinned by design (D10):
   it carries no tokens at all.

## Token index

Default values are the shipped angel-fiction skin (today's only skin, and the
literal text a fresh clone still renders). This table grows as later waves
introduce fiction strings; TASK-121 is expected to promote it into the runtime
contract's default-skin table rather than replace it.

| Token | Default (angel-skin) value | Used by |
|---|---|---|
| `skin.guardian.name` | `Metatron` | `panels/guardian.md` (pane header `{{skin.guardian.name}} · charges…`, transcript verdict rows, chronicle stage line reference), `patterns/keymap.md` (dock-tab table) |
| `skin.guardian.tab_label` | `metatron` (case-transformed active/inactive by the dock's existing tab styling) | `panels/dock.md` (tab row), `patterns/keymap.md` |
| `skin.guardian.epithet` | `angel` | `panels/guardian.md` (`you`/`{{skin.guardian.epithet}}` transcript labels, "the {{skin.guardian.epithet}} is answering…"/"unreachable" copy), `panels/minibuffer.md` (dormant placeholder "speak with the {{skin.guardian.epithet}}…") |

## Deferred to TASK-121

Out of scope for this page and this feature: the on-disk token/skin file
format, resolution order (world skin → default), fallback behavior when a
token is missing, and the sweep enumerating every literal in
`internal/tui/help.go`, footer hints, `stagesLadder`, this design corpus, and
`docs/player/` that TASK-121 must retarget at runtime lookups. This page only
guarantees the design corpus itself writes no new bare literal in the
meantime.

---
title: Pattern — skin tokens (doc conventions + the runtime contract)
class: pattern
status: shipped
verified_against: eae30eba00796982c6d0cbca4740adf4473ab95e
---

# Pattern: skin tokens

This page is two things: (1) the documentation conventions for how this
reference writes fiction strings, and (2) — since spec 052 (TASK-121) — the
**doc twin of the runtime skin contract** (`internal/skin`,
`specs/052-skinnable-guardian/contracts/skin-contract.md`). The interim
"documentation conventions only" posture, and this page's own requirement
that TASK-121's PR adopt or amend it, is **satisfied by that PR**: the token
index below is promoted into the runtime default-skin table.

## Why this exists

The fiction layer is skinnable data (spec 052): the *displayed* proper name,
epithet, tab label, and narration vocabulary are skin data, not fixed
identifiers. The design reference speaks in the generic vocabulary
(**guardian**) throughout and marks every place a fiction string renders with
a token; the shipped client resolves the same tokens at runtime through
`internal/skin`. A skin swap is a content change (a per-world `skin.json`),
never a doc rewrite or a code change.

## The runtime contract (spec 052)

- **Lookup**: `internal/skin` — `(*skin.Skin).Resolve(token)` plus typed
  accessors (`Name()`, `Epithet()`, `TabLabel()`, `FamilyLabel()`,
  `WorkingNoun()`, `FormNoun(form)`, `Stage(id)`/`StageName(id)`).
- **Resolution order**: world `skin.json` override → compiled default table →
  **the token path itself** (visibly wrong, never an empty string). A
  rendered token path is a bug; the token-completeness test
  (`internal/skin/completeness_test.go`) fails before it ships.
- **Skin bundle**: `<world>/skin.json` beside `charter.md` — identity fields
  (`name`, `epithet`, `tab_label`), a `strings` token-override map, `stages`
  display identities, and one long-form `voice` composed at the guardian
  prompt's editable-zone SOUL seam. Boot-frozen; capabilities.json fallback
  discipline (missing → default silently; malformed/invalid field → default +
  one notice; unknown keys/tokens ignored + notice). Format authority:
  `specs/052-skinnable-guardian/contracts/skin-contract.md` §1; living
  example: `examples/skins/raven.json`.
- **Transport**: clients never read world files — the daemon resolves the
  skin at boot and the status surfaces carry the display facts (`skin_*`
  fields, additive omitempty); absent fields (an old daemon) render the
  default skin.
- **Never skinnable**: the fixed frame (spec 021), mechanics (tool ids,
  costs, charges, reducer rules), the event log (recorded types, payloads,
  memory text, correlation ids), and systems/telemetry content (D10 — no
  tokens exist for it, by construction of the table below).

### Downstream obligations (TASK-115/117 and later)

1. **No new bare fiction literal** — every fiction string a new surface
   renders is a token lookup; the repo-wide fiction-denylist sweep test
   enforces this.
2. **New tokens** land in the same commit in all three places: the default
   table (`internal/skin`), this page's table below, and the
   token-completeness test's reach (automatic — it enumerates the table).
3. **Non-fiction chrome** (telemetry, key hints, structural labels) stays
   literal with `—` in the skin-token column (rule 5 below).

## Doc conventions

1. **Mockups** (fenced ASCII blocks): every fiction string is written as
   `{{skin.<domain>.<name>}}` — double-brace, the token's dotted path, no
   quoting. Example: a dock tab row renders `{{skin.guardian.tab_label}}`.
2. **Control tables**: the `skin-token` column (contracts/control-table.md)
   names the token in unbraced dotted form — `skin.guardian.name` — or `—`
   for non-fiction controls (telemetry rows, chrome structure, keys).
3. **Prose**: describes the *role*, never the fiction word — "the guardian
   tab", "the guardian's proper name" — except when quoting exactly what a
   mockup or control table cell shows, which follows rules 1–2.
4. **Case is a rendering detail, not a token detail.** One token covers every
   case-transformed rendering of the same string (e.g. the dock tab's
   lowercase-inactive / uppercase-active styling, the pane header's
   uppercased `{{skin.guardian.name}}`) — the page that uses it says so in
   prose once; the token itself never forks by case.
5. **Non-fiction strings** (chrome labels, telemetry, key hints, glyphs) are
   literal text with `—` in the skin-token column — most of the reference.
   Systems-tab content (panels/systems.md) is never skinned by design (D10):
   it carries no tokens at all.

## The default skin table (normative doc twin)

Default values are the shipped secular-mythic **Guardian** skin (spec 052
ruling 3). This table is the doc twin of `internal/skin`'s compiled default
table; the token-completeness test asserts the two never drift. Old worlds'
already-written files and every serialized identifier (event types
`metatron.*`, tool ids `send_vision`/`send_omen`/`work_miracle`, paths,
`metatron_*` wire names) are **frozen** and deliberately not in this table.

| Token | Default value | Used by |
|---|---|---|
| `skin.guardian.name` | `Guardian` | `panels/guardian.md` (pane header `{{skin.guardian.name}} · charges…`, uppercased per rule 4), chronicle narration subject lines (digest grammar), prompt name substitution (digest keeper, watch confirmer) |
| `skin.guardian.epithet` | `guardian` | `panels/guardian.md` (transcript labels, "the {{skin.guardian.epithet}} is answering…"/"unreachable"/"voice is stilled" copy), `panels/minibuffer.md` (dormant placeholder), help overlay copy, morgue watch line |
| `skin.guardian.tab_label` | `guardian` | `panels/dock.md` (tab row, case-transformed active/inactive), `patterns/keymap.md` footer hints, narrow-fallback tabs |
| `skin.guardian.family_label` | `guardian` | chronicle Type-column display alias for the frozen `metatron.*` event family (spec 052 FR-013); dock short-form and the detail pane's verbatim type stay raw |
| `skin.guardian.working_noun` | `working` | transcript/chronicle/CLI display of the frozen `work_miracle` mechanics family; prompt doctrine glosses |
| `skin.guardian.working_noun_plural` | `workings` | granted-tool console summary, chronicle grant vocabulary |
| `skin.guardian.notes_label` | `the guardian's notes` | display references to the frozen `metatron/soul.md` path |
| `skin.guardian.vision_noun` | `vision` | display rendering of the frozen `"vision"` nudge form (payloads/tool ids frozen; default-skin-retained folk vocabulary) |
| `skin.guardian.omen_noun` | `omen` | display rendering of the frozen `"omen"` nudge form |
| `skin.guardian.report_card_label` | `report card` | the report-card box title (spec 063): `pages/guardian-console.md` inline card, chronicle digest row for `guardian.report_card`, postmortem card header |
| `skin.guardian.attribution_label` | `what your words did` | the attribution note's own block header inside the card (spec 063 standing resolution 1 — the note is its own block beneath the checklist) |
| `skin.guardian.example_ask.send_vision` | `"show Ash a vision of the fire dying"` | `overlays/help.md` guardian section (D9): one canned ask per granted verb, keyed by the frozen tool id |
| `skin.guardian.example_ask.send_omen` | `"send everyone an omen tonight: stay near the fire"` | help guardian section (D9) |
| `skin.guardian.example_ask.monitor_and_act` | `"watch for anyone going hungry, and warn them"` | help guardian section (D9) |
| `skin.guardian.example_ask.cancel_order` | `"release the watch on the fire"` | help guardian section (D9) |
| `skin.guardian.example_ask.work_miracle` | `"work a working: grant Ash food from thin air"` | help guardian section (D9) |
| `skin.guardian.example_ask.pause` | `"pause the world"` | help guardian section (D9) |
| `skin.guardian.example_ask.start` | `"start the world again at 4x"` | help guardian section (D9) |
| `skin.guardian.example_ask.adjust_speed` | `"slow the world down to 1x"` | help guardian section (D9) |
| `skin.guardian.example_ask.explain` | `"what does a vision cost?"` | help guardian section (D9) |
| `skin.stage.stage-1.name` … `skin.stage.stage-4.name` | The Voice / The Written Word / The Craft / The Stewardship | stage display identities (`internal/skin` StageIdentity; spec 046 surfaces) |
| `skin.stage.stage-1.line` … `skin.stage.stage-4.line` | "you speak, it acts" / "your law outlives the conversation" / "you shape what it can do" / "a world in your care" | one-line stage identity descriptions |
| `skin.stage.stage-2.ceremony_chapter` | "Your play proved The Written Word: a law that outlives the conversation, written once and honored by every turn since." | `overlays/ceremony.md`'s D6 authorship-voice narrated chapter (spec 056) — stage-1 has no entry (never unlocked, `sim.EvaluateUnlock` never returns it) |
| `skin.stage.stage-3.ceremony_chapter` | "Your play proved The Craft: what the guardian can do now bears your own hand in its shaping." | `overlays/ceremony.md`'s D6 chapter |
| `skin.stage.stage-4.ceremony_chapter` | "Your play proved The Stewardship: a world now stands in your care, exactly as you left it." | `overlays/ceremony.md`'s D6 chapter |

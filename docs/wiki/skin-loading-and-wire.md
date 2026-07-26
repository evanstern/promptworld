---
name: skin-loading-and-wire
description: How [[skin]] loads and validates skin.json (Load/Parse: single-line identity caps, per-field fallback-with-notice, unknown-key/token tolerance, the voice character cap), the raw.json example skin's zero-notice completeness test, and the status-wire carriage (FromFacts) that lets a client resolve skin facts without reading skin.json directly. Split from [[skin]]; read when touching internal/skin/load.go or the wire contract.
kind: component
sources:
  - internal/skin/load.go
  - examples/skins/raven.json
verified_against: 31c893e0406653197e467a89b2fdb96f0bcf2ee0
---

# Skin loading and wire carriage

Split from [[skin]] (corpus-spec v2 size-budget split, summary-style): this
note covers how a skin is loaded/validated from disk, the example skin that
documents the format, and how the resolved facts travel to clients over the
wire. See [[skin]] for the token vocabulary and the `Skin` type's typed
accessors.

## Loading and validation

**Loading** (`load.go`, `Load(worldDir)`): reads `<worldDir>/skin.json`
following the `capabilities.json` fallback discipline (spec 052 FR-003,
research R1/R4) — no file is the common case (silently the default skin); an
unreadable or malformed file is the default skin plus one notice; an invalid
individual FIELD falls back to that field's default plus a notice while the
rest of the document still applies (one bad field never bricks the whole
skin); unknown top-level keys (`knownSkinKeys`: `name`/`epithet`/`tab_label`/
`voice`/`strings`/`stages`) or unknown token paths inside `strings` are
ignored with a notice, never an error. Identity fields (`name`/`epithet`/
`tab_label`) are validated as single-line (`singleLine` — no control
characters, the name-injection surface's shape rule), non-blank, UTF-8, and
length-capped (`nameMaxRunes` 40, `epithetMaxRunes` 20 shared by epithet and
tab_label); generic `strings` overrides cap at `stringMaxRunes` (120) and a
stage's `line` at `lineMaxRunes` (120); `voice` is the one long-form field,
capped at `voiceMaxChars` (4000, the bundle-SOUL character-cap precedent) —
hostile VOICE CONTENT is the [[guardian]]'s fixed prompt frame's job to
contain (FR-004), the loader's cap only bounds volume. `Parse` is the
loader's testable core (`Load` adds only the file read); both never return a
nil `*Skin`. The result is boot-frozen exactly like [[bundle-tools]]' bundle
surface and [[curriculum-ladder]]'s stage facts: loaded once at daemon boot,
installed via `SetSkin`, and a skin.json edit takes effect only on restart.

## The example skin

**The example skin** (`examples/skins/raven.json`, spec 052 FR-014, T017):
the format's living documentation — a full alternate re-theme ("the Raven", a
trickster spirit) covering every identity field, every vocabulary string
(`working_noun` → `"trick"`, `vision_noun` → `"whisper"`, `omen_noun` →
`"wingbeat"`, `family_label` → `"raven"`), and all four stage identities
("The Whisper", "The Bargain", "The Rookery", "The Long Flight" — same
one-line identities, re-voiced names). `TestExampleRavenSkinLoads`
(`example_skin_test.go`) proves it loads with ZERO notices — a stale example
would otherwise silently drift from the format `Parse` actually accepts.

## Status wire carriage

**Status wire carriage** (spec 052 FR-012, contract §7): the daemon resolves
the boot-frozen skin daemon-side and sends the resolved facts on
`ipc.StatusData` — `SkinName`/`SkinEpithet`/`SkinTabLabel`/`SkinFamilyLabel`
(identity fields, always sent by a post-052 daemon, resolved against the
compiled default table even for the default skin) plus `SkinStrings`/
`SkinStages` (only a world skin's OVERRIDES, `omitempty` — the default skin
sends neither map) so a client (CLI, TUI) never reads world files to render
skin vocabulary; `FromFacts` is the client-side twin of `Load`, rebuilding a
`*Skin` from exactly those wire facts for local `Resolve` calls (e.g. the
[[tui-client]] help overlay's lesson-catalog skin resolution). Additive
`omitempty` fields mean a pre-052 daemon/client pair interoperates unchanged
— absent fields render the default Guardian skin.

## Back to parent

[[skin]] links here for loading/validation and wire carriage; that note's own
Connections section lists [[guardian]], [[curriculum-ladder]], and
[[tui-client]] as the consumers of the facts this note describes how they
arrive (`Load` daemon-side, `FromFacts` client-side).

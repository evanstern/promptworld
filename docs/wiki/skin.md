---
name: skin
description: The runtime skin substrate (spec 052, TASK-121) — the fiction layer as data. One token lookup (world skin.json override → compiled default table → the token path itself) supplies every display string the player sees for the guardian and the curriculum ladder; the event log and every serialized identifier stay skin-free.
kind: component
sources:
  - internal/skin/skin.go
  - internal/skin/load.go
  - examples/skins/raven.json
verified_against: 31c893e0406653197e467a89b2fdb96f0bcf2ee0
---

# Skin

`internal/skin` is the fiction-as-data substrate spec 052 (TASK-121) introduced
alongside the [[guardian]] package rename: the substrate itself knows only
neutral, mechanical vocabulary (stage ids `stage-1`..`stage-4`, event types
like `metatron.nudged`, tool ids like `work_miracle`), and every string the
PLAYER sees for that vocabulary — the guardian's display name, its epithet,
its tab label, the chronicle's family alias, the vision/omen/working nouns,
and each curriculum stage's display identity — is resolved through one
`*Skin` value at render/prompt-composition time. Two governing rulings hold
the whole design together: ruling 1 — the event log is skin-free, nothing in
this package ever touches a recorded payload; ruling 3 — the DEFAULT skin is
the secular-mythic "Guardian" (the game's out-of-the-box fiction, itself now
just one skin among possible others).

## How it works

**Resolution order** (`Resolve`, the FR-001 contract): a token path resolves
through the world's override map, else the compiled `defaultTable`, else the
token path itself — a rendered `skin.guardian.name` string on screen is a
visible bug, never a silent empty string, and `TestTokenCompleteness`
(`completeness_test.go`) fails the build before it ships: every token any Go
source, test, or in-repo example skin consumes must have a `defaultTable`
row, and every row must resolve to a non-empty, non-path value.

**The token vocabulary** (`skin.go`): identity tokens —
`skin.guardian.name` (`TokenName`, default `"Guardian"`), `.epithet`
(`TokenEpithet`, default `"guardian"`), `.tab_label` (`TokenTabLabel`,
default `"guardian"`), `.family_label` (`TokenFamilyLabel`, default
`"guardian"` — the chronicle's Type-column alias for the FROZEN `metatron.*`
event namespace, FR-013) — plus vocabulary tokens `working_noun`/
`working_noun_plural` (default `"working"`/`"workings"` — the display name
for the frozen `work_miracle` mechanics family), `notes_label` (default
`"the guardian's notes"`), and `vision_noun`/`omen_noun` (default
`"vision"`/`"omen"` — display nouns for the frozen `send_vision`/`send_omen`
tool ids and the recorded `metatron.nudged` payload's `form` values `vision`/
`omen`). Stage-identity tokens (`skin.stage.<id>.name`/`.line`) are appended
from `defaultStages` in `init()` so the two tables can never drift; the
default stage identities are the client-approved names carried over the
de-theme unchanged (spec 052 ruling 3) — **The Voice** ("you speak, it
acts"), **The Written Word** ("your law outlives the conversation"), **The
Craft** ("you shape what it can do"), **The Stewardship** ("a world in your
care") — read by [[curriculum-ladder]]. Since spec 056 ([[takeover-surfaces]]),
`defaultCeremonyChapters` appends one `skin.stage.<id>.ceremony_chapter` row per
UNLOCKABLE stage (stage-2..stage-4 — stage-1 is the ladder's floor and is never
unlocked, so it carries no entry) alongside the stage-identity rows, the D6
authorship-voice "your play proved `<identity>`" narrated chapter a ceremony
takeover renders. Since spec 063 ([[grounded-feedback]]), the table also
gains `report_card_label`/`attribution_label` (the report card's box title
and the attribution note's own block header) and a per-verb `example_ask`
family, one `skin.guardian.example_ask.<tool-id>` row per shipped guardian
loop tool (e.g. `.send_vision` → `"show Ash a vision of the fire dying"`) —
the help overlay's D9 guardian section teaches asking from these.

**`Skin` and its typed accessors**: a `*Skin` holds string-token overrides
(`strings`, identity fields included — they ARE tokens), stage-identity
overrides (`stages`), and one `voice` string (the persona-voice text composed
at the [[guardian]]'s editable-zone SOUL seam, `""` = no fragment). The zero
value and a nil `*Skin` both mean the default Guardian skin — every method is
nil-safe (`Default()` returns `&Skin{}`; the package-level `Stage`/`StageName`
funcs resolve against a nil `*Skin`), so "no skin.json" and a pre-052 status
reply's absent skin fields both render the default with no special-casing at
any call site. Typed accessors are thin `Resolve` wrappers:
`Name`/`Epithet`/`TabLabel`/`FamilyLabel`, `WorkingNoun`/`WorkingNounPlural`,
`NotesLabel`, `Voice`, and `Stage(id)`/`StageName(id)` (unknown/`""` ids —
including a pre-ladder world — return the zero identity and `false`, a safe
fallback for message text). `FormNoun(form)` maps a recorded nudge form value
(`"vision"`/`"omen"`, the frozen payload vocabulary) to its display noun;
an unrecognized form value renders verbatim rather than empty. Since spec
056, `CeremonyChapter(stage)` resolves a stage's D6 narrated chapter (a
plain token lookup, so a stage-1 lookup — never actually rendered — is
visibly wrong rather than a crash). Since spec 063, `ReportCardLabel()`/
`AttributionLabel()` resolve the report card's two display labels, and
`ExampleAsk(toolID)` resolves the per-verb example-ask family — assembled
from split literals deliberately, since the token-completeness sweep
matches whole dotted token paths and this family's membership is enumerated
by the table itself.

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

**The example skin** (`examples/skins/raven.json`, spec 052 FR-014, T017):
the format's living documentation — a full alternate re-theme ("the Raven", a
trickster spirit) covering every identity field, every vocabulary string
(`working_noun` → `"trick"`, `vision_noun` → `"whisper"`, `omen_noun` →
`"wingbeat"`, `family_label` → `"raven"`), and all four stage identities
("The Whisper", "The Bargain", "The Rookery", "The Long Flight" — same
one-line identities, re-voiced names). `TestExampleRavenSkinLoads`
(`example_skin_test.go`) proves it loads with ZERO notices — a stale example
would otherwise silently drift from the format `Parse` actually accepts.

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

## Connections

[[guardian]] is the primary consumer: `SetSkin` installs the boot-frozen
`*Skin` (the `SetBundles`/`SetStage` discipline), and `Name`/`Epithet`/
`Voice`/`WorkingNoun`/`FormNoun` compose into the turn prompt, the fuzzy
confirm and digest-keeper system prompts, and moment/notice text — never
into a recorded payload. [[guardian-orders]] and [[guardian-miracles]] read
the same accessors for triggered-turn moment lines and miracle-outcome
phrasing. [[curriculum-ladder]] reads `Stage`/`StageName` for the ladder's
display identities in `promptworld stages`/`new`/`status` and the stage lock
notices (`stageCharter`/`stageSkills`, [[guardian]]'s `charter.go`).
[[tui-client]] resolves the chronicle Type column's frozen-namespace alias
through `FamilyLabel` (`internal/tui/grammar.go`'s `displayEventType`) and
the console pane's tab/labels through the polled status facts.
[[cli-promptworld]] renders the same facts offline/online without reading
`skin.json` directly. [[bundle-tools]]' persona-SOUL-fragment composition is
the precedent `Voice`'s placement in the prompt stack follows (after bundle
SOULs, still beneath the guardian's fixed frame). [[takeover-surfaces]]
(spec 056) reads `CeremonyChapter` for the ceremony takeover's narrated
chapter; [[grounded-feedback]] (spec 063) reads `ReportCardLabel`/
`AttributionLabel` for the report-card console card and `ExampleAsk` for
the help overlay's D9 guardian section.

## Operational notes

Nothing in this package is deterministic-replay-sensitive — a skin never
touches `sim.State` or any event payload (ruling 1), so re-theming a world
mid-run (editing `skin.json`, effective next daemon restart per the
boot-frozen discipline) can never affect simulation determinism or replay.
`TestTokenCompleteness` is the structural guard against silent token drift as
new display surfaces are added; a token consumed anywhere in the tree without
a `defaultTable` row fails the build before it ships. The default skin is the
game's authored identity (spec 052 ruling 3) — a project may ship or accept
alternate skins (the Raven example), but the mechanics vocabulary (event
types, tool ids, IPC methods, on-disk paths, correlation-id prefixes) never
moves regardless of which skin is active.

---
name: skin
description: The runtime skin substrate (spec 052, TASK-121) — the fiction layer as data. One token lookup (world skin.json override → compiled default table → the token path itself) supplies every display string the player sees for the guardian and the curriculum ladder; the event log and every serialized identifier stay skin-free. Loading/validation and wire carriage split to [[skin-loading-and-wire]].
kind: component
sources:
  - internal/skin/skin.go
verified_against: cf65debb44c1e17b54c0f3421d11e1e8cc28576c
---

# Skin

`internal/skin` is the fiction-as-data substrate spec 052 (TASK-121) introduced
alongside the [[guardian]] package rename: the substrate itself knows only
neutral, mechanical vocabulary (stage ids `stage-1`..`stage-4`, event types
like `guardian.nudged`, tool ids like `work_miracle`), and every string the
PLAYER sees for that vocabulary — the guardian's display name, its epithet,
its tab label, the vision/omen/working nouns,
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
`"guardian"` — the family's voiced name; its consumer, spec 052 FR-013's
Type-column alias, was retired by the spec-094 rename: the chronicle
renders types raw) — plus vocabulary tokens `working_noun`/
`working_noun_plural` (default `"working"`/`"workings"` — the display name
for the frozen `work_miracle` mechanics family), `notes_label` (default
`"the guardian's notes"`), and `vision_noun`/`omen_noun` (default
`"vision"`/`"omen"` — display nouns for the frozen `send_vision`/`send_omen`
tool ids and the recorded `guardian.nudged` payload's `form` values `vision`/
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
loop tool (e.g. `.send_vision` → `"show Ash a vision of the fire dying"`;
spec 084 adds the five plan-layer rows, `.place_designation` through
`.survey_site` — [[guardian-designations]]; spec 085 adds `.prophesy` —
[[guardian-faith]]; spec 107 adds the three mission rows,
`.accept_mission`/`.note_mission_progress`/`.cancel_mission` —
[[guardian-missions]]) —
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

**Loading and wire carriage** — split to [[skin-loading-and-wire]]: how
`Load`/`Parse` read and validate `<worldDir>/skin.json` (per-field
fallback-with-notice, identity/length caps, the `voice` character cap), the
`examples/skins/raven.json` example skin that documents the format with a
zero-notice completeness test, and the `ipc.StatusData` wire carriage
(`FromFacts`) that lets a client (CLI, TUI) resolve skin facts without
reading `skin.json` directly.

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
[[tui-client]] renders the chronicle Type column raw since spec 094 (the
FR-013 alias shim is deleted) and resolves the console pane's tab/labels
through the polled status facts.
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

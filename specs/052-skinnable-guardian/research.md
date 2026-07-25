# Research: Skinnable guardian persona (spec 052)

## R1 — Where the skin lookup lives

**Decision**: grow `internal/skin` into the runtime skin substrate — its own
package doc (skin.go:1-9) already declares "TASK-121 absorbs this package as
the real skin substrate; only the lookup grows a skin dimension". The package
gains: the default-skin token table, a `Skin` value type (identity fields +
token overrides + stage identities + voice), `Load(worldDir)` (boot-frozen,
capabilities.json fallback discipline), and token resolution
(`skin → default table → token path itself`). Existing `Stage`/`StageName`
accessors (skin.go:32-44) become skin-dimension-aware rather than replaced —
all current call sites keep working.

**Rationale**: the beachhead exists and is already consumed by
cmd/promptworld (commands.go:136,143,937), internal/tui (tui.go:234,
digest.go:1197), and internal/metatron lock notices (charter.go:119,135).

**Alternatives considered**: a new `internal/fiction` package — rejected:
duplicate home; skin.go's doc already names this destiny.

## R2 — Prompt-composition seam for the persona voice

**Decision**: the voice composes at the existing SOUL-fragment seam —
`turnSystemPrompt` (internal/metatron/turn.go:860-900) already stacks
(1) charter, (2) bundle SOUL fragments under `--- persona ---`
(turn.go:863-865), (3) skills, (4) the fixed frame LAST as compile-time
constants (turn.go:869-871: `metatronNonNegotiables` turn.go:817-820,
`metatronInitiativeFrame` turn.go:829-831, registry-derived tool guidance).
The skin voice is one more editable-zone fragment (same ≤4,000-char cap and
validation discipline as bundle SOULs, internal/bundle/validate.go:52-72),
inserted before the frame. Identity fields (name/epithet) substitute as
validated single-line data into the de-themed constant prompts that need a
name (fixed frame greeting, digest keeper metatron/digest.go:192-194, watch
confirmer metatron/orders.go:387-390).

**Rationale**: injection-soundness (spec 021 INV-1) is preserved by
construction — the frame is appended last on every path and no skin byte can
displace it; the adversarial battery in metatron_test.go extends with
hostile-skin fixtures rather than needing a new proof shape.

**Alternatives considered**: voice inside the fixed frame — rejected: the
frame is the never-skinnable zone by definition (spec ruling; AC #3).

## R3 — Skin transport to the clients

**Decision**: the daemon resolves the world's skin at boot (`SetSkin`
following the SetBundles/SetStage boot-frozen discipline,
metatron/metatron.go) and the status surface carries resolved display facts
additively (omitempty): skin name, epithet, tab label, stage identities.
TUI/CLI render exclusively from status (plus the compiled default table for
static usage text); absent fields (old daemon) → default skin.

**Rationale**: the TUI already renders exclusively from status/replica; per
spec FR-012 clients never read world files. Additive-omitempty is the
established status evolution pattern (spec 046 curriculum provenance
precedent, wiki metatron note "Surfaces").

## R4 — Frozen vs renamed vocabulary (the compat contract)

**Decision** (spec ruling 2, operationalized):

**FROZEN (serialized/wire/disk — annotate at definition site):**
- Event types: `metatron.*` (charge_regenerated, nudged, place_revealed,
  order_*, charter_observed, time_snapped, item_granted, entity_moved,
  entity_removed), `curriculum.*` — replay compatibility.
- Payload values incl. `sim.OriginOmen = "omen"` (sim/memory.go:47) and
  memory-prefix text ("You saw a vision: ", turn.go:526-528) — recorded
  history + FR-005 (event log is skin-free).
- Correlation-id prefixes `turn-metatron-`/`watch-metatron-`
  (tui/decisions.go:34 consumes) — appear in cog.tool_call payloads.
- IPC methods `metatron_chat`/`metatron_status` (ipc/server.go:512,533);
  `ps --json` key `metatron_charges` (cmd/promptworld/ps.go:129); state JSON
  tag `metatron_charges` (sim/loop.go:42).
- llm.json route kinds `metatron`/`metatron_watch` (llm/config.go:489).
- Tool ids `send_vision`/`send_omen`/`work_miracle` + miracle kind enum
  (tool/registry.go:463,475,507-521) — capabilities.json + recorded
  tool_call payload compatibility.
- On-disk paths: `charter.md`, `skills/`, `capabilities.json`,
  `metatron/soul.md`, `metatron/transcript.md`.

**RENAMED (pure Go, compiler-safe):**
- Package `internal/metatron` → `internal/guardian` (import sites:
  daemon wiring, ipc server, scribe/morgue, cmd).
- Unserialized identifiers: `paneMetatron`, `familyMetatron` (the KEY stays
  frozen where serialized — grammar family key `"metatron"` grammar.go:87
  maps event-type prefixes, so the string is frozen, the Go const renames),
  `metatronVerdictRow`, `sim.Metatron*` Go names where JSON tags carry the
  frozen form, `stagesLadder` internals, etc. Rule of thumb: rename the
  identifier, freeze the string literal it carries if that literal is ever
  serialized or compared against recorded data.

**Compat aliases (user-typed vocabulary):** CLI `metatron` → canonical
`guardian`, `miracle` → canonical `work`; old names hidden but functional
(main.go:102-105 dispatch gains alias entries; usage text main.go:33-42,
miracle.go:14-21 shows canonical only).

**Rationale**: memory `squash-rewrites-branch-pins`/determinism doctrine —
recorded logs must replay; player-authored configs must load; renames with
zero serialization footprint are free under the compiler.

## R5 — Display aliasing for frozen leaks (FR-013)

**Decision**: the chronicle Type column (grammar.go:173-189 solo ≤26 runes;
`shortType` last-segment dock form grammar.go:193-198) renders the leading
family segment through a skin family-label lookup: default skin maps
`metatron` → `guardian` (display only). The detail pane's verbatim event
(type + payload JSON) and the grammar-miss raw fallback (grammar.go:187)
stay raw — inspector surfaces per spec 047's FR-020 audience ruling.

**Rationale**: AC #2's "internal identifiers swept or aliased per spec
decision" — aliasing at the one projection users read constantly, honesty at
the surfaces that promise verbatim.

## R6 — Default skin content

**Decision**: name `Guardian`, epithet `guardian`, tab label `guardian`
(case-transformed by dock styling per skin-tokens.md rule 4); interventions
display as **workings** (`work_miracle` tool id frozen; all display/gloss
text re-worded); vision/omen retained (TutorCharter precedent,
persona/charter.go:40-43 explicitly notes the TASK-121 direction); stage
identities keep The Voice / The Written Word / The Craft / The Stewardship
(already secular, skin.go:22-27). `persona.DefaultCharter`
(persona/charter.go:9-33) rewritten guardian-voiced with the skin name
substituted at genesis seed time; `TutorCharter` already compliant.
`charterIsDefault`/fingerprint comparisons are already preset-aware — the new
default text slots in.

## R7 — Example alternate skin

**Decision**: `examples/skins/raven.json` — "the Raven", trickster
folk-tale identity (original, no licensed IP), overriding name/epithet/tab
label/voice/a handful of string tokens + stage identities. Ships with a
README snippet in the same directory documenting the format (the living
format documentation FR-014 requires).

## R8 — Sweep-site inventory (grounding for US2/US4; gathered 2026-07-25 against main)

### TUI user-facing literals
- internal/tui/help.go:112,113,126,207,210,220 — metatron tab references +
  "the angel's transcript…" + "ask the angel" (5+ literals).
- internal/tui/views.go:254 (footer hint "3 metatron"), 376 (dock tab label),
  1446 ("ask the angel anything…"), 1452 ("angel ⋮ thinking…"), 1486-1487
  (transcript "angel" labels), 1511 ("METATRON · charges"), 1535 ("the angel
  is unreachable"), 1549 ("the angel's voice is stilled"), 1734 ("the angel
  is answering…"), 1744 ("speak with the angel…").
- internal/tui/tui.go:48 (paneNames "metatron"), 460 ("angel: " transcript
  prefix), 250-253 (grant summary "miracles"/"miracles(kinds)").
- internal/tui/digest.go:933,944,968,977,985,992,1006,1034,1043,1056,1074,
  1197 — "Metatron …" chronicle subject lines (visions/omens, place reveal,
  watches, charter, miracles family, stage unlock).
- internal/tui/grammar.go:66,87,94 (family key "metatron"/"curriculum" —
  string frozen, display aliased per R5).

### CLI user-facing
- cmd/promptworld/main.go:33,40-42,102-105 — usage + subcommand vocabulary.
- cmd/promptworld/miracle.go:14-21,116 — miracleUsage + charge output.
- cmd/promptworld/commands.go:396,417-422,444,447,450 — metatron
  usage/status/act lines.
- cmd/promptworld/stages.go:41-67,138-153 — stagesLadder prose (already
  "guardian"-worded; "visions, omens" stays per R6).

### Prompts (model-facing)
- internal/persona/charter.go:9-33 DefaultCharter ("You are Metatron…") —
  rewrite; 44-85 TutorCharter — already compliant.
- internal/metatron/turn.go:869-871 fixed frame (neutral; gains name
  substitution), 890 ("cannot work a miracle for free" → working), 817-831
  invariant constants (de-themed wording only if fiction-bearing).
- internal/tool/derive.go:202-209 metatronToolDesc glosses; 217-222 miracle
  kind glosses.
- internal/metatron/digest.go:192-194 digest keeper ("You are Metatron…").
- internal/metatron/orders.go:387-390 confirmSystem ("Metatron's watchful
  eye"), 640 trigger moment ("I worked a miracle").
- internal/metatron/toolcalls.go:185 ("the miracle is worked: ").

### Recorded-at-emission text (FROZEN per FR-005 — do NOT skin)
- turn.go:526-528 memory prefixes ("You saw a vision/witnessed an omen");
  turn.go:387-390 place-grant memory; miracle_batch.go:41-42,94 (already
  secular); mind/prompt.go:172, mind/consolidate.go:306, mind/narrate.go:184-190
  (vision/omen vocabulary — retained in default skin anyway).

### Files/documents written fresh (sweep the templates, not history)
- internal/metatron/metatron.go:212 soul genesis header ("The soul of
  Metatron… The angel has seen nothing yet").
- internal/metatron/turn.go:545-546 soul appends ("I sent a %s…" — neutral;
  keep).
- internal/scribe/morgue.go:342 ("The angel's watch at that moment").

### Docs
- README.md lines 30,40,62,63 (+98 llm.json kinds — frozen vocabulary,
  reworded around).
- docs/design/tui/ — token mockups already `{{skin.*}}` (spec 047);
  patterns/skin-tokens.md gains the runtime contract section (its own
  requirement); re-pin affected pages.
- docs/player/ — regenerated post-merge (player-docs skill); page renames
  (playing-via-metatron.html → guardian) happen there.

### Curriculum substrate (neutral — verify only)
- internal/world/world.go:101-116 stage ids; internal/sim/curriculum.go
  reducer arms; cmd/promptworld/stages.go:92-105 JSON twin (id/name/line
  separation already clean).

### Per-world config precedent for skin.json
- Manifest additive fields: world/world.go:29-88 (Stage 62-69,
  CharterPreset 74-79, reserved Scenario block 80-96).
- capabilities.json loader discipline: metatron/charter.go:402-484.
- unlocks.json atomic-write/load-tolerant: worlds/unlocks.go:86-115.
- Boot-frozen injection: mt.SetBundles/mt.SetStage (metatron.go; wiki
  metatron note) — SetSkin follows.

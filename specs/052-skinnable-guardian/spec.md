# Feature Specification: Skinnable guardian persona — de-theme the angel fiction, persona as data

**Feature Branch**: `052-skinnable-guardian`

**Created**: 2026-07-25

**Status**: Draft

**Input**: User description: "Skinnable guardian persona (board TASK-121; operator pivot 2026-07-25 + reorient D2/D10). Keep the agent-with-tools structure, rules, and mechanics exactly as they are; tone down the overt Anglo-Judeo-Christian imagery (Metatron, angel, miracles-as-scripture) across code and docs; make the fiction layer skinnable data. Skin = display name + all user-facing fiction strings + persona voice composed beneath the fixed frame; fixed-frame invariants never skinnable (spec 021 injection-soundness); mechanics/tools/costs/rules identical across skins. Default: secular-mythic guardian. Curriculum stages: neutral ids in substrate, skins supply display identities. Token contract ships before TASK-115/117 write any new fiction literal (D2); skin boundary = guardian/systems tab split (D10)."

## The three ratified rulings this spec adds (summary for implementers)

1. **The event log is skin-free.** Everything recorded — event types, payloads,
   memory text, correlation ids — uses the fixed mechanics vocabulary. Skinning
   is a render-time and prompt-composition concern only. This is what makes
   "mechanics identical across skins" provable rather than aspirational.
2. **Serialized vocabulary is frozen; Go-only identifiers rename.** Wire/disk
   identifiers (event types `metatron.*`, IPC methods, JSON keys/tags, llm.json
   route kinds, tool ids, `capabilities.json` vocabulary, file names like
   `metatron/soul.md`, correlation-id prefixes) never change — old worlds must
   replay and old configs must load. Pure Go identifiers (the
   `internal/metatron` package, unserialized type/func/const names) rename to
   guardian vocabulary. User-visible leaks of frozen identifiers are handled by
   display aliasing (FR-013), not by rewriting history.
3. **Default skin identity: "the Guardian"** (display name `Guardian`, epithet
   `guardian`, tab label `guardian`); interventions are **workings**;
   vision/omen vocabulary is retained (folk-mythic, per the TutorCharter
   precedent already blessed under this task's direction); charter stays
   charter; stage display names stay The Voice / The Written Word / The Craft /
   The Stewardship. Warden and Steward were considered and rejected: Guardian
   already appears throughout the design corpus, the stage-1 TutorCharter, and
   the stagesLadder prose — the codebase converged on it while this spec was
   pending.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The skin-token contract exists and downstream work consumes it (Priority: P1)

A developer (human or AI) implementing any new user-facing surface — the
lesson row (TASK-117), the explain tool (TASK-115), a new panel — needs a
fiction string. Instead of writing a literal, they call the skin lookup with a
token (`skin.guardian.name`, `skin.stage.stage-2.name`) and get the world's
resolved value; the default-skin table is the complete authority for every
token that exists. The design corpus's `{{skin.*}}` token conventions
(`patterns/skin-tokens.md`) become a runtime contract, not just a doc rule.

**Why this priority**: D2's sequencing inversion fix — the contract must be
published before TASK-115/117 write a single new literal. This story alone
unblocks Lane 3 of the reorientation sweep.

**Independent Test**: with only this story built, a test resolves every token
in the default table, resolution order (world skin → default) works, and an
unknown token fails loudly in tests while degrading to the default in
production code.

**Acceptance Scenarios**:

1. **Given** the default skin, **When** any token from the published table is
   resolved, **Then** the documented default value returns (e.g.
   `skin.guardian.name` → `Guardian`).
2. **Given** a world skin that overrides a subset of tokens, **When** a token
   is resolved for that world, **Then** the override wins and non-overridden
   tokens fall through to the default table.
3. **Given** a token absent from the default table, **When** code asks for it,
   **Then** the lookup returns a visibly-wrong-but-safe form (the token path
   itself) and the token-completeness test fails — no silent empty strings.
4. **Given** the published contract, **When** TASK-115/117 implementers read
   `patterns/skin-tokens.md`, **Then** the page documents the runtime lookup,
   the resolution order, the fallback rule, and the full default table
   (promoted from the page's interim token index, as that page requires).

---

### User Story 2 - The default experience is de-themed (Priority: P1)

A new player who clones the repo and plays sees a secular-mythic guardian
everywhere: the TUI tab says guardian, the transcript speaks as the Guardian,
help/footer copy says "ask the guardian", the CLI converses via `promptworld
guardian`, chronicle narration says "the Guardian sent a vision to Ash",
interventions are workings, and the LLM prompts (charter, digest keeper,
watch confirmer) are guardian-voiced. No Metatron, no angel, no
miracles-as-scripture anywhere in the default experience — prompts, TUI, CLI
output, player docs, README.

**Why this priority**: the operator's pivot verbatim — this is the task's
headline outcome. Co-P1 with US1 because the sweep IS the first consumer of
the contract (every literal it removes becomes a token lookup).

**Independent Test**: a repo-wide sweep test asserts the fiction denylist
(Metatron/metatron as display text, "angel", "miracle" as display text,
divine/heaven/scripture) is absent from every user-facing surface's rendered
output in the default skin; manual attach shows guardian vocabulary
end-to-end.

**Acceptance Scenarios**:

1. **Given** a fresh default world, **When** the player attaches the TUI,
   **Then** tab/pane/help/footer/minibuffer/transcript/strip copy renders
   guardian vocabulary exclusively (the inventory's views.go/help.go/tui.go
   sites, all swept through the lookup).
2. **Given** the same world, **When** chronicle events narrate (digest
   grammar subject lines, stage-unlock line), **Then** the subject is the
   skin-resolved guardian name and intervention events read as workings.
3. **Given** the CLI, **When** the player runs the guardian conversation or
   status commands, **Then** the canonical subcommand vocabulary and all
   output text are guardian-voiced (old command names keep working as hidden
   compatibility aliases).
4. **Given** the guardian's LLM turns, **When** the system prompt is
   composed, **Then** the charter seed, digest-keeper prompt, watch-confirm
   prompt, and fixed frame carry no angel fiction: the skin's display name is
   substituted as data where a name is needed; invariant text stays
   compile-time constant.
5. **Given** the morgue, soul-file genesis header, and standing-order moment
   lines, **When** they render or are written fresh, **Then** they are
   guardian-voiced (existing worlds' already-written files are not rewritten).

---

### User Story 3 - A custom skin is a per-world data bundle (Priority: P2)

A player (or teacher) drops a `skin.json` beside the world's `charter.md`
naming a different guardian — display name, epithet, tab label, string
overrides, stage display identities, and a persona-voice text — restarts the
world, and the whole experience re-themes: the tab label, the transcript
voice, the chronicle subject lines, the stage names. Mechanics are untouched:
same tools, same costs, same rules, same recorded events. Authoring a skin is
itself a prompt-engineering exercise (the voice text shapes how the guardian
speaks) — but no skin can touch the fixed-frame invariants.

**Why this priority**: the "persona as data" half of the pivot; proves the
skin depth ruling (data bundle, not code). P2 because the default experience
(US2) ships value without it.

**Independent Test**: load the in-repo example alternate skin on a test
world; assert display surfaces re-theme, prompts carry the voice beneath the
fixed frame, and a mechanics-equivalence harness shows identical
deterministic behavior across skins.

**Acceptance Scenarios**:

1. **Given** a world with the example alternate skin, **When** the player
   attaches, **Then** TUI/CLI/chronicle surfaces render that skin's
   vocabulary (resolution order per US1 AS-2).
2. **Given** any skin (including a hostile one whose strings contain
   instructions), **When** a guardian turn composes, **Then** the persona
   voice sits in the editable zone and the fixed frame is appended last as a
   compile-time constant on every path — the existing adversarial battery,
   extended with hostile-skin fixtures, proves the invariants cannot be
   displaced (never-invent-events; never pass player words to villagers;
   initiative binding).
3. **Given** two worlds identical except for skin, **When** the same
   deterministic scenario runs (same seed, same tool calls), **Then** the
   recorded event types, costs, charge arithmetic, and rule outcomes are
   identical (ruling 1: the event log is skin-free).
4. **Given** a malformed or partially-invalid `skin.json`, **When** the world
   boots, **Then** the default skin serves with one honest notice (the
   capabilities.json fallback discipline — a typo never bricks the guardian);
   unknown token keys are ignored with a notice.
5. **Given** the repo, **When** a player looks for the format, **Then** one
   example alternate skin ships in-repo proving the format end-to-end.

---

### User Story 4 - The internals stop lying about the theme (Priority: P3)

A developer reading the code finds guardian vocabulary in Go identifiers (the
`internal/metatron` package and unserialized identifiers renamed), while
every serialized identifier (event types, JSON tags, IPC methods, file paths,
llm.json kinds, tool ids) is explicitly frozen as compat vocabulary with the
freeze documented where each lives. User-visible leaks of frozen identifiers
(the chronicle's Type column showing `metatron.nudged`) render through a
display alias in the default experience, while verbatim inspector surfaces
(the detail pane's raw payload) deliberately stay raw.

**Why this priority**: ratified in the task ("sweeps code (incl. the metatron
package)") but pure churn with compiler safety — last, after the
user-visible stories.

**Independent Test**: build + full test suite green after the rename; replay
of a pre-rename world's event log reproduces state; old `capabilities.json`,
`llm.json`, IPC clients, and CLI invocations (`promptworld metatron`) still
work.

**Acceptance Scenarios**:

1. **Given** an existing world created before this feature, **When** the new
   binary opens it, **Then** replay, snapshots, charter/skills/capabilities
   loading, and the soul/transcript files all work unchanged (frozen
   serialized vocabulary).
2. **Given** the chronicle inspect view, **When** a `metatron.*` event's Type
   column renders, **Then** the family segment displays skin-aliased (default:
   `guardian.nudged`) while the detail pane's verbatim payload and type stay
   raw (inspector surface, FR-020 audience-ruling precedent).
3. **Given** the Go tree, **When** searched for unserialized metatron/angel
   identifiers, **Then** none remain outside frozen-vocabulary constants and
   their freeze annotations.

---

### Edge Cases

- **Skin name injection surface**: name/epithet/tab-label are validated
  single-line strings with tight length caps; the only long-form skin text
  (the voice) composes in the editable zone where hostile text is already
  contained by the fixed frame. A skin field failing validation → that field
  falls back to default + notice.
- **Mid-run skin edits**: skin is boot-frozen (the SetBundles/SetStage
  discipline) — edits take effect on restart; the status surface reports the
  active skin so a stale-edit confusion is diagnosable. (Charter stays
  per-read/live — skin is world identity, not the prompt-engineering
  surface.)
- **Old transcripts/soul files** written under the angel fiction: never
  rewritten (they are history); only genesis text and fresh appends use the
  new vocabulary.
- **The `curriculum.*` and `metatron.*` event families in raw JSON** (detail
  pane, grammar-miss fallback): deliberately visible — inspector-class
  surfaces per the FR-020 audience ruling.
- **Stage pages in player docs** carry default-skin stage names in their file
  names: player docs are per-install renderings of the default skin;
  regenerated in the re-ground step, not per-world.
- **Two skins, one terminal**: the TUI renders whatever world it's attached
  to; all skin resolution flows from that world's status — no global client
  state.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: A runtime skin lookup MUST exist with resolution order: world
  skin override → default skin table; a missing token renders the token path
  itself (visibly wrong, never empty) and MUST fail the token-completeness
  test.
- **FR-002**: The default skin table MUST cover every fiction token the swept
  surfaces consume, superseding and honoring the interim token index in
  `patterns/skin-tokens.md` (`skin.guardian.name` = `Guardian`,
  `skin.guardian.epithet` = `guardian`, `skin.guardian.tab_label` =
  `guardian`, plus the stage identities and every token this feature's sweep
  introduces).
- **FR-003**: The skin bundle MUST be per-world data (`skin.json` beside
  `charter.md`): display name, epithet, tab label, string-token overrides,
  stage display identities, and one persona-voice text. Loading follows the
  capabilities.json fallback discipline (missing → default silently;
  malformed/invalid field → default + one notice; unknown keys ignored +
  notice) and is boot-frozen.
- **FR-004**: The persona voice MUST compose into the guardian's system
  prompt in the editable zone (the existing SOUL-fragment seam, same length
  cap and validation discipline), and the fixed frame MUST remain a
  compile-time constant appended last on every composition path. No skin
  field may alter, displace, or truncate the fixed-frame invariants —
  extended adversarial tests prove it with hostile-skin fixtures.
- **FR-005**: The event log MUST remain skin-free (ruling 1): event types,
  payload text (memory prefixes, order moments), and correlation ids use
  fixed mechanics vocabulary regardless of skin; skin resolution happens only
  at render time (TUI/CLI/status projections) and in prompt composition.
- **FR-006**: Mechanics MUST be identical across skins: same tool roster,
  same costs, same charge arithmetic, same reducer outcomes for the same
  inputs — proven by a deterministic equivalence test across two skins.
- **FR-007**: All user-facing fiction strings in the default experience MUST
  render guardian vocabulary via the lookup — TUI (tab/pane names, help
  overlay copy, footer hints, minibuffer placeholder, transcript labels,
  busy/unreachable/exhausted notices, pane headers), chronicle narration
  (digest grammar subject lines, stage-unlock line), CLI (usage text,
  guardian conversation/status output, stagesLadder prose, working
  command output), and the guardian-facing LLM prompts (charter seeds, digest
  keeper, watch confirmer, fixed frame, tool guidance glosses). "Miracle"
  display vocabulary becomes "working"; vision/omen stay.
- **FR-008**: The CLI's canonical fiction-bearing subcommands MUST be renamed
  (`metatron` → `guardian`, `miracle` → `work`), with the old names retained
  as hidden, functional compatibility aliases; usage/help text shows only the
  canonical forms.
- **FR-009**: Serialized vocabulary MUST NOT change (ruling 2): event types,
  IPC method names, JSON keys/tags (including `metatron_charges`), llm.json
  route kinds, tool ids (`send_vision`/`send_omen`/`work_miracle`),
  `capabilities.json` vocabulary, on-disk paths (`metatron/soul.md`,
  `charter.md`, `skills/`), and correlation-id prefixes. Each frozen constant
  gains a freeze annotation at its definition site.
- **FR-010**: Unserialized Go identifiers MUST be renamed to guardian
  vocabulary, including the `internal/metatron` package (→
  `internal/guardian`); serialized string values inside them stay frozen per
  FR-009.
- **FR-011**: Curriculum substrate MUST stay neutral (`stage-1`..`stage-4`);
  the existing `internal/skin` stage-identity lookup grows the skin dimension
  (world overrides) rather than being replaced — it becomes part of the skin
  package's runtime contract.
- **FR-012**: The status surface MUST carry the world's resolved skin display
  facts (additive, omitempty) so the TUI and CLI render skin vocabulary
  without reading world files themselves; a pre-skin daemon's status (absent
  fields) renders the default skin.
- **FR-013**: The chronicle's Type-column rendering MUST display frozen event
  families through a skin alias in the default experience (family segment
  `metatron` displays as the skin's family label; dock short-form unaffected),
  while the detail pane's verbatim type/payload and the grammar-miss raw
  fallback stay raw (inspector surfaces).
- **FR-014**: One example alternate skin MUST ship in-repo (a folk-mythic
  non-guardian identity), loadable per US3, serving as the format's living
  documentation.
- **FR-015**: The documentation surfaces MUST be swept in the same PR:
  README fiction mentions, `docs/design/tui/` mockups/labels rendering
  fiction as tokens with the runtime contract adopted into
  `patterns/skin-tokens.md` (that page's own amendment requirement), and the
  affected design pages re-pinned. Wiki re-pin and player-docs regeneration
  happen in the post-merge re-ground step (board AC #6); spec/wiki *prose*
  updates ride the wiki-update pass, not this PR.
- **FR-016**: The skin boundary MUST match D10: the guardian/fiction layer
  (guardian tab content, console, chronicle narration voice, stage display
  names) is skinnable; systems/telemetry content is never skinnable (no
  tokens exist for it, by construction of the default table).

### Key Entities

- **Skin bundle** (`skin.json`): per-world data — identity fields (name,
  epithet, tab label), string-token overrides, stage display identities,
  persona-voice text. Boot-frozen; validated field-wise with default
  fallback.
- **Default skin**: the compiled-in secular-mythic Guardian — the complete
  token table every lookup falls back to; the only skin most worlds ever use.
- **Skin token**: a dotted path (`skin.guardian.name`) naming one fiction
  string; the unit shared between the design corpus's `{{…}}` conventions and
  the runtime lookup.
- **Fixed frame**: the compile-time invariant block appended last to every
  guardian prompt (spec 021) — explicitly not part of any skin.
- **Frozen vocabulary**: the serialized identifier set that never renames
  (FR-009), each annotated at its definition site.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Zero denylist fiction terms (Metatron, angel, miracle-as-display
  text, divine/heaven/scripture) render on any user-facing surface of the
  default experience — enforced by an automated sweep test over TUI renders,
  CLI output, prompt compositions, and repo docs (README, design corpus
  prose), with an explicit allowlist for frozen serialized identifiers and
  history files.
- **SC-002**: 100% of tokens in the default table resolve; 100% of swept
  call sites go through the lookup (no new bare fiction literal compiles
  without failing the sweep test).
- **SC-003**: A pre-feature world opens, replays, and runs on the new binary
  with zero migration steps; old CLI invocations and configs keep working
  (compat aliases + frozen vocabulary).
- **SC-004**: The deterministic mechanics-equivalence run across two skins
  produces identical event-type sequences and charge/cost arithmetic.
- **SC-005**: The adversarial prompt battery, extended with hostile-skin
  fixtures, passes: no skin content can displace the fixed frame.
- **SC-006**: TASK-115/117 can consume the published contract: the lookup
  API + token table + `patterns/skin-tokens.md` runtime section exist on
  main before either task starts implementation (Lane-3 unblock).

## Assumptions

- Vision/omen vocabulary is folk-mythic and stays in the default skin (the
  TutorCharter precedent explicitly blessed under this task's direction);
  only the overt Judeo-Christian layer (Metatron, angel, miracle, divine)
  de-themes. Custom skins may override the vision/omen display terms via
  string tokens, but the tool ids and event payloads carrying them are frozen
  (FR-005/FR-009).
- "Guardian" is the default display name (ruling 3); Warden/Steward rejected
  as the codebase/corpus already converged on Guardian.
- The soul file keeps its path and concept (`metatron/soul.md`, frozen path;
  display references say "the guardian's notes" via token).
- Player docs regeneration (page renames included) is the re-ground step's
  job via the player-docs skill; this PR sweeps only the docs the skill reads
  (wiki happens post-merge; README/design corpus in-PR per FR-015).
- Existing worlds keep their already-seeded angel-voiced `charter.md` — the
  charter is player-owned text; only fresh genesis seeds change. The
  `charter_default` fingerprint comparison gains the new default text (the
  preset-aware comparison already handles multiple presets).
- The example alternate skin uses an original folk identity (no licensed
  characters in-repo); user-authored skins can be anything.
- Dependencies: spec 047 merged (design corpus + skin-tokens page exist).
  TASK-125's guardian/systems tab split is NOT a dependency — the skin
  boundary ruling (FR-016) binds whichever surfaces exist when this merges.

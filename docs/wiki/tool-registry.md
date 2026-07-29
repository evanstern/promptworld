---
name: tool-registry
description: The single source of truth for agent capabilities (spec 014, extended specs 017/019/021/029/041/058) — every tool as name + params + gate + effect + cost in one registry; a tool call is a REQUEST, an event is the FACT. Overview + cross-note connections here; the tool catalog, Guardian tool surface, authored-schema/clamp mechanics, and derive/roster/validate machinery each split into their own child note (see below) — load a child for its file-level detail.
kind: component
sources:
  - internal/tool/tool.go
  - internal/tool/registry.go
  - internal/tool/derive.go
  - internal/tool/roster.go
  - internal/sim/toolcheck.go
verified_against: 6a5344a12cdc8858909ca7cf209d55025135e9d5
---

# Tool registry

`internal/tool` (spec 014, TASK-53) formalizes everything an agent can do — villager
verbs to the Guardian's nudges — as a `Tool`: name, param schema, gate class, effect class,
and cost, in ONE registry. The core principle it encodes: a tool call is a REQUEST;
an event is the FACT; the gate decides; the executor grounds work in time and space.
The model never asserts outcomes. This layer is deliberately behavior-identical: it
replaced three hand-maintained duplicate maps (`goalVocabulary` in the prompt,
`validGoals` in the parser, `planGoals` at the sim door) that had ALREADY drifted in
shipped code — the plan-step map silently rejected the nine spec-012 verbs (TASK-55);
curing that drift is the migration's sole permitted behavioral delta (FR-012).

## How it works

This note's file-level detail splits into four children by domain
(corpus-spec v2 summary-style split); each links back here.

**The World/villager tool catalog** — `registry.go`'s tool-group assembly
order (`worldTools`/`worldToolsBase`, `set_plan`, `expressiveTools`,
`guardianTools`, `journalTools`), the `Tool` struct's `EffectClass`/`Param`/
`GateClass`/`Cost` fields, the four spec-019 journal tools' `Read`/
`Expressive` split, and the `isLegacyWorldTool` discriminator that governs
free-text vocabulary membership — moved to [[tool-registry-world-catalog]].

**The Guardian tool surface** — the spec-029 agency surface (`send_vision`/
`send_omen`/`monitor_and_act`/`cancel_order`/the meta tools,
[[guardian-orders]]), the authoritative miracle cost table (`MiracleCost`/
`MiracleCostsByEvent`), `RestrictEnum`'s per-world capability gating, and the
derived `GuardianToolGuidance`/`GuardianReadGuidance`/
`GuardianTargetingGuidance` prompt prose — moved to
[[tool-registry-guardian-tools]].

**Authored schemas and clamp-with-notice** — why `set_plan` and
`work_miracle` need hand-authored `InputSchemaJSON` overrides instead of
`Params` derivation, and the spec-058 Clamp mechanism (`reasonParam`/`say`/
`gist`/`muse`/`set_plan`'s `reason`) that truncates an over-cap expressive
field instead of rejecting the call — moved to
[[tool-registry-schema-clamping]].

**Derived surfaces, rosters, and validation** — `derive.go`'s per-consumer
walks (`VocabularyLine`/`PromptGlossBlock`/`WorldGoals`/`PlanStepGoals`/
`InputSchema`), `roster.go`'s villager/Guardian roster data (`RosterVillager`/
`RosterGuardian`/`LoopRosterVillager`/`LoopRosterGuardian`, the dormant-verb
prune), and the boot-time `tool.Validate()`/`sim.ValidateToolCoverage()`
consistency gates — moved to [[tool-registry-derivation-rosters]].

## Connections

[[agent-mind]] derived its prompt vocabulary, gloss, and parser accept set from
here pre-spec-017 (retired with the free-text planner reply); [[tool-loop]] is the
new consumer: `Job.Roster` is `tool.LoopRosterVillager()`/`LoopRosterGuardian()`,
and `InputSchema(t)` builds each declared tool's wire schema. [[sim-loop]]'s
injection doors enforce roster membership at landing; [[reflex-policy]]'s
`resolveGoal` table and [[executor]]'s duration table are the sim-side
derivations the coverage gate cross-checks; [[guardian]] / [[guardian-orders]]
read the registry for the `send_vision` text cap, the granted acting-tool
guidance, and `work_miracle` dispatch; since spec 059 [[guardian-orders]]/
[[guardian-miracles]] also read `GuardianTargetingGuidance()` for the miracle
targeting digest's prose pointer; [[daemon-lifecycle]]
runs the boot gates; [[agent-journal]] is the spec-019 consumer of the four
journal tools (`write_journal_entry`/`delete_from_journal`/`search_journal`/
`read_journal`) declared here. [[mental-maps]] is the spec-041 consumer of
`search` and `send_vision`'s place grant, and the source of the
`placeFactKinds` vocabulary hand-mirrored onto `place_kind`'s Enum.
[[executor]]/[[reflex-policy]] are the spec-064 consumers of `warm_up` —
the resolver, the completion-hold state machine, and the reflex's own
conditioned-`goto_warmth` issuance all read this note's registry entry and
clamp home. [[tool-loop]]
owns the spec-058 `Clamp` enforcement (`validateArgs`'s clamp-with-notice,
`VerdictLandedClamped`); [[sim-loop]]'s landing guard owns the matching
`set_plan` step-count clamp (`OutcomeClamped`); [[agent-mind]]'s scene parser
(`parse.go`) is where `say`/`gist`'s own `Clamp`-flagged fields actually
enforce, since both stay outside the villager tool-use loop. The registry formalizes the doors — it does not
relax them: the landing ladder, whitelist, and charge economy are unchanged
enforcers. [[grounded-feedback]] is the spec-063 consumer of `explain` (this
note's registry entry, `GuardianReadGuidance`) and of `MiracleCost`/
`miracleKindArgs` (its `costs`/`workings` fact sheets read the exact source
[[guardian-miracles]] documents). Spec: `specs/014-tool-registry/` (contracts/registry-api.md,
contracts/tool-catalog.md); the tool-use loop additions are spec 017
(`data-model.md` §1-2, R11-R13); the journal tools and reason param are spec 019
(R12, T024).

## Operational notes

Migration proven behavior-identical: the full replay/determinism suite passed with
zero test-file edits, the golden-prompt fixture (`prompt_golden_test.go`, captured
pre-refactor — retired with spec 017 once the free-text planner reply it pinned
was replaced by native tool declarations) passed byte-unchanged through the
derivation, and the whitelist was pinned diff-identical (17 entries). Live smoke
on a throwaway world: boot gates
passed, planners landed, and a multi-step plan naming `collect_water` landed — the
TASK-55 drift cure visible live (the old map rejected it). A test-only tool
registered in a test build appears in every derived surface with zero other edits.

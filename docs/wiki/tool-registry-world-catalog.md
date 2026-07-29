---
name: tool-registry-world-catalog
description: Split from [[tool-registry]] (spec 014) — registry.go's tool-group assembly order (worldTools/worldToolsBase, set_plan, expressiveTools, guardianTools, journalTools), the Tool struct's EffectClass/Param/GateClass/Cost fields, the four spec-019 journal tools' Read/Expressive split, and the isLegacyWorldTool discriminator gating free-text vocabulary membership.
kind: component
sources:
  - internal/tool/tool.go
  - internal/tool/registry.go
  - internal/tool/derive.go
  - internal/tool/roster.go
verified_against: 11de2a4aa93d4c901a8dd90369151fa23fd056d0
---

# Tool registry — the World/villager tool catalog

Split from [[tool-registry]] (spec 014, TASK-53): the catalog assembly, the
journal tools, and the legacy/loop discriminator.

**The catalog** (`registry.go`): the pre-spec-017 30 entries, plus two spec-017
additions — `set_plan` (a loop-only planning tool) and `work_miracle` (the Guardian's
fourth tool) — plus four spec-019 additions, the journal tools; plus six spec-032
additions (US1-3: walls, an axe, paths); spec 029 then retires the two nudges and
adds seven Guardian tools (`send_vision`, `send_omen`, `monitor_and_act`,
`cancel_order`, `pause`, `start`, `adjust_speed`); and spec 041 US4 adds one more
World verb, `search` (`Effect: World, Gate: Resolvable, Cost.DurationTicks: 0,
PlanStep: true, ReflexEligible: true`) — deliberate exploration, appended after
`build_path` so the registration-order byte anchor holds; no args in v1 (a kind
hint is a documented future extension, since selection is nearest-frontier
regardless); and spec 064 (needs-conditioned recovery, [[executor]]/
[[reflex-policy]]) adds one more still, `warm_up` (`Effect: World, Gate:
Resolvable, Cost.DurationTicks: 0, PlanStep: true`, NOT `ReflexEligible` — the
reflex's own warmth rungs issue the equivalent conditioned `goto_warmth`
themselves, so `warm_up` is planner-only), appended after `search` — the
warmth-RECOVERY verb — walk to known warmth and LOITER there until warmth
actually recovers, instead of `goto_warmth`'s arrive-and-done — with an
optional `until_warmth` `Number` param (no `Min`/`Max`; the driver never
rejects it, the sim clamps it with notice, spec 064 R3 / the 058 clamp
posture) — assembled in order: `worldTools`
(now 32 World verbs: the 24 legacy verbs in the old goal-vocabulary order, then
`build_wall_plank`/`build_wall_stone`/`demolish`/`repair`/`craft_axe`/`build_path`/
`search`/`warm_up`,
appended after `withdraw` so no existing tool's registration position shifts —
`worldToolsBase` wraps these too, so every one also gains the shared `reason`
param), `set_plan`, `expressiveTools` (`say`/`gist`/`muse`), `guardianTools`
(spec 029's agency surface, [[guardian-orders]]: `converse`/`send_vision`/
`send_omen`/`monitor_and_act`/`cancel_order`/`pause`/`start`/`adjust_speed`/
`work_miracle` — the retired `nudge_dream`/`nudge_omen` replaced by
`send_vision`/`send_omen` — plus, since spec 063, `explain` appended last
([[grounded-feedback]]'s read-only mechanics-facts tool: `Effect: Read,
Gate: None`, a free-text `topic` param deliberately NOT an `Enum` so an
unknown topic can return the explainable-topic catalog as a repairable
miss rather than a schema rejection)), `journalTools`
(`write_journal_entry`/`delete_from_journal`/`search_journal`/`read_journal`,
appended last so no existing tool's position shifts). The tool groups are
declared as separate literals (`worldTools`, `expressiveTools`, `guardianTools`,
`journalTools`) rather than one, so `set_plan`'s schema can be built from
`worldTools` alone and spliced in — building it from the assembled `registry` would
be an initialization cycle. The spec-032 verbs carry `PlanStep: true` like every
pre-existing World verb, so they fall out of `isLegacyWorldTool`'s discriminator
for free — planner-only, no `ReflexEligible` — and join `RosterVillager`/
`LoopRosterVillager` with no separate list to maintain; the storage verbs'
`itemKinds` vocabulary (`kind` param on `drop`/`pick_up`/`deposit`/`withdraw`)
gains `"axes"` alongside the pre-existing kinds. (Since TASK-163, this plural
storage vocabulary is deliberately DISTINCT from `work_miracle give_item`'s
singular `grantKinds`/`GrantKinds()` — a grant delivers ONE fresh spear/axe at
a time, never the storage field's many; see [[tool-registry-guardian-tools]].) Each `Tool` (`tool.go`) carries an `EffectClass`
(`World` → intents, executor-grounded; `Expressive` → immediate whitelisted event
batches; `Read` → data back into cognition, consumed by [[tool-loop]] — spec 019
ships the first production Read entries, `search_journal`/`read_journal`),
`Param`s (`AgentName`/`Text`/`Enum`/`Number` kinds — `Number` (spec 017 R12) pays
the spec-014 debt that left the storage verbs' `qty` unmodeled: bounded by
optional `Min`/`Max`, 0/0 meaning unbounded; every `Param` also carries an
optional `Description`, spec 019 T024, emitted verbatim as the derived JSON
Schema property's `"description"` — `""` means no description; since spec 058
FR-001, `Clamp bool` marks an EXPRESSIVE `Text` param whose over-cap value the
[[tool-loop]] driver truncates rune-safely and lets the call proceed, instead
of rejecting it — `Validate()` enforces `Clamp` only ever appearing on a `Text`
param), a `GateClass`,
and a `Cost` (work `DurationTicks` for world verbs, `TextCapBytes`/
`TextCapRunes` for expressive text, nudge/miracle charges). World verbs keep
their prompt gloss prose (`PromptGloss`) byte-exact from the old hand-written
prompt block. `converse` is classified Expressive with empty `Events` (it lands
no world events — transcript-only), so `Validate`'s Events rule is
one-directional: Events non-empty ⇒ Expressive.

**The journal tools** (`registry.go`, spec 019 US3): `write_journal_entry` and
`delete_from_journal` are Expressive (their `journal.entry_written`/
`journal.entry_deleted` Events land through the `InjectSocial` door like any
other whitelisted batch); `search_journal` and `read_journal` are `Read` — data
returned into cognition, grounding nothing. All four carry `Gate: None` (no
scene, no charge — the reducer dry-run, budget or existence, is the only gate)
and are villager-only: they join `LoopRosterVillager` alone, never a Guardian
roster, since journals are private. Every acting villager world tool also gains
an optional bounded `reason` param (spec 019 R12 / T024) — the model's free-text
"why" for the action — via a post-declaration pass: `worldTools` wraps a new
`worldToolsBase` literal, appending `reasonParam()` (`Kind: Text, Required:
false, MaxRunes: ReasonCapRunes` = 200, with a capability-only description) to
every entry's `Params`, so the shared param is defined once and no verb's
literal repeats it. `reason` is deliberately absent from `muse` (interiority is
already free-standing) and every Guardian tool. `set_plan`'s authored
`InputSchemaJSON` (`setPlanSchema`) separately gains an optional top-level
`reason` string (same `ReasonCapRunes` cap) alongside its `steps` array — the
plan-level why, threaded to `InjectArgs.Reason`.

**The legacy/loop split** (`isLegacyWorldTool`, `derive.go`, spec 017 R11): a World
tool no longer automatically belongs to the free-text goal vocabulary — the
discriminator is `Effect == World && PlanStep`. Every pre-spec-017 World tool
already carries `PlanStep: true` (the TASK-55 single-walk invariant), so the
filter changes nothing for them; `set_plan` is `Effect World` (it lands through
the same `InjectIntent` path) but carries `PlanStep: false` because it is
loop-only vocabulary, not a legacy free-text goal — so it is excluded from
`VocabularyLine()`, `WorldGoals()`, and `RosterVillager` for free, with no
separate exclusion list anywhere.

## Connections

Part of [[tool-registry]]'s summary-style split (corpus-spec v2). See
[[tool-registry]] for the registry's overall doctrine (tool call vs event vs
gate vs executor) and its other split-off domains: [[tool-registry-guardian-tools]]
(the Guardian agency surface and miracle costs), [[tool-registry-schema-clamping]]
(authored `InputSchemaJSON` overrides and spec-058 clamp-with-notice), and
[[tool-registry-derivation-rosters]] (derive.go/roster.go/validate.go). [[agent-mind]]
consumed this catalog's pre-spec-017 vocabulary; [[tool-loop]] is the current
consumer via `LoopRosterVillager()`; [[agent-journal]] is the spec-019 consumer
of the four journal tools declared here.

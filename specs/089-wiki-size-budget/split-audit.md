# Spec 089 — no-fact-lost spot-check

For every note that was split, this maps each pre-change `##`/`**bold**`
section to where its content lives after the change. For tightened-only
notes (no split), the pre-change section list is unchanged — every section
heading survives; only prose density changed.

## Splits (new child created)

### sim-state-apply-world.md → sim-state-apply-guardian-records.md (T002)

| Pre-change content | Post-change location |
|---|---|
| `State.m`/`scenario` fields, mental-map growth arms, designation/directive/faith/prophecy dispatch, derived bookkeeping (moved/woke), gru/incident/stranger/governance/miracle dispatch, `world.migrated`/`world.forked` | stays in **sim-state-apply-world.md** |
| Standing-order lifecycle (`order_placed`/`triggered`/`cancelled`/`expired`) | **sim-state-apply-guardian-records.md** |
| `metatron.charter_observed` handling | **sim-state-apply-guardian-records.md** |
| `morgue.epilogue` → `applyMorgueEpilogue` | **sim-state-apply-guardian-records.md** |
| `guardian.report_card` → `applyReportCard` | **sim-state-apply-guardian-records.md** |
| `curriculum.exercise_passed`/`stage_unlocked` → `applyCurriculum` | **sim-state-apply-guardian-records.md** |
| `sim.tuning_applied`, `State.Tuning`, nil-safe accessors | **sim-state-apply-guardian-records.md** |
| Connections | split across both notes' own Connections sections, cross-linked |

### reflex-policy.md → reflex-policy-history.md (T003)

| Pre-change content | Post-change location |
|---|---|
| `decideIntent`/`resolveGoal` overview, `needClassGoals`/`needClassOf` (spec 083) | stays in **reflex-policy.md** |
| `resolveGoal` vocabulary growth: spec 012 (economy), spec 013 (storage), spec 032 (walls/axes/paths), spec 014/TASK-53 (`goalResolvers` table restructuring) | **reflex-policy-history.md** |
| Spec 041's knowledge-gating rewrite (`nearestKnown`/`nearestKnownAdjacentTo`, epistemic-failure resolvers, `search` goal, `talk_to`/`seek` last-known-sighting) | **reflex-policy-history.md** |
| `## How it works` (spec 062 rung restructuring), five pre-existing child summaries, Connections, Operational notes | stays in **reflex-policy.md**, now six child summaries |

### sim-state-apply-agents.md → sim-state-apply-agents-resources.md (T003)

| Pre-change content | Post-change location |
|---|---|
| Genesis placement, `Apply`'s clock/intent/movement/eating/talk/needs/death arms, spec-083 neglect anchors | stays in **sim-state-apply-agents.md** |
| v2 resource/crafting events (`quarried`/`collected_water`/`crafted`/`cooked`/`bathed`/`refueled`/`spear_broke`, axe-vs-bare yields, `removeHarvestedFact`) | **sim-state-apply-agents-resources.md** |
| v3 storage events (`dropped`/`picked_up`/`deposited`/`withdrew`, `food_rotted`, `chest_taken`) | **sim-state-apply-agents-resources.md** |
| Spec-032 wall demolish/repair HP family (`built`/`wall_chipped`/`wall_destroyed`/`wall_repaired`) | **sim-state-apply-agents-resources.md** |

### sim-loop-injection-doors.md → sim-loop-injection-doors-telemetry.md (T003)

| Pre-change content | Post-change location |
|---|---|
| Door mechanism (`InjectSocial`/`InjectOperator`), world/social-effecting whitelist half: `metatron.place_revealed`, the four miracle types, guardian-order types, the four plan-layer types, `prophecy.declared`, `InjectableSocialEvent` accessor | stays in **sim-loop-injection-doors.md** |
| `meeting.proposal_rephrased`, `cog.*` telemetry, `journal.entry_written`/`deleted`, `agent.belief_reinforced`, `agent.memory_embedded`/`situation_embedded`, `cog.memory_divergence`, `metatron.charter_observed`, `morgue.epilogue`, `metatron.skills_observed`, `guardian.report_card` | **sim-loop-injection-doors-telemetry.md** |
| Dry-run mechanics paragraph, `InjectOperator` section, Connections, spec-086 section | stays in **sim-loop-injection-doors.md** |

### guardian-miracle-rebase-taxonomy.md → guardian-miracle-rebase-shift-fields.md (T003)

| Pre-change content | Post-change location |
|---|---|
| Intro, the **KEEP** bullet (full field list), `TestRebaseTaxonomyComplete` guard paragraph, Connections | stays in **guardian-miracle-rebase-taxonomy.md** |
| The **SHIFT** bullet (full field-by-field, spec-by-spec list: 029/084/085/041/043/061/062/077/083) | **guardian-miracle-rebase-shift-fields.md** |

### guardian-designations.md → guardian-designations-reception.md (T003)

| Pre-change content | Post-change location |
|---|---|
| Intro, `## The two entities`, `## The seven-event vocabulary and its doors` | stays in **guardian-designations.md** |
| `## The announcement grant` | **guardian-designations-reception.md** |
| `## The villager side — hardness` (DIRECTIVE reflex rung, `directive` context block) | **guardian-designations-reception.md** |
| `## Guardian-side surfaces` (tool handlers, `survey_site`, TUI rendering, miracle rebase treatment, scope guards) | **guardian-designations-reception.md** |

### tool-registry-guardian-tools.md → tool-registry-plan-faith-tools.md (T003)

| Pre-change content | Post-change location |
|---|---|
| The spec-029 agency surface (`send_vision`/`send_omen`/`monitor_and_act`/`cancel_order` + meta tools), the miracle-cost source and spec-021 derivations (`GuardianToolGuidance`/`GuardianReadGuidance`/`GuardianTargetingGuidance`/`RestrictEnum`) | stays in **tool-registry-guardian-tools.md** |
| Spec-084 plan-layer tools (`place_designation`/`cancel_designation`/`issue_directive`/`cancel_directive`/`survey_site`) | **tool-registry-plan-faith-tools.md** |
| Spec-085 `prophesy` tool | **tool-registry-plan-faith-tools.md** |

### tui-chronicle-feed.md → tui-chronicle-feed-guardian-digests.md (T003)

| Pre-change content | Post-change location |
|---|---|
| Chronicle rendering, core digest grammar, mental-map/memory-retrieval family entries, spec-077 stranger/incident entries, spec-086 payload-naming section, Back to parent | stays in **tui-chronicle-feed.md** |
| Run-outcome/morgue entries (spec 044), curriculum entries (spec 046), report-card entry (spec 063), `world.forked` entry (spec 076), plan-layer entries (spec 084), the four guardian-miracle entries + `gratisMark` | **tui-chronicle-feed-guardian-digests.md** |

### executor-needs-survival.md → executor-needs-recovery-and-neglect.md (T003)

| Pre-change content | Post-change location |
|---|---|
| Heartbeat, fire fuel, eating, run end | stays in **executor-needs-survival.md** |
| Needs-conditioned recovery holds (spec 064: `UntilNeed`/`UntilValue`, `recoveryHoldEvents`, `warm_up`, cold-emergency wake arm) | **executor-needs-recovery-and-neglect.md** |
| Neglect detector (spec 083: `NeglectDue`, `sim.neglect_detected`, `Agent.Neglect` anchors/latch) | **executor-needs-recovery-and-neglect.md** |

### guardian-faith.md → guardian-prophecy.md (T003)

| Pre-change content | Post-change location |
|---|---|
| Faith state and the one writer, the five-reason delta table, regen as a pure faith-band function, Surfaces, Scope guards, spec-086 section | stays in **guardian-faith.md** |
| `## Prophecy — the staked vision` (the `prophesy` tool, `sim.Prophecy` entity discipline, claim predicate table, declaration-door refusals) | **guardian-prophecy.md** |

## Tightened in place (no split, no section lost)

event-types-cognition-telemetry.md, tui-client-mechanics.md, tile-registry.md,
mental-map-perception.md, executor-social-perception.md, explain-tutor-guide.md,
village-lens.md, event-types-agent-intents.md, curriculum-ladder.md,
social-fabric.md (T001); ipc-protocol.md, chronicle.md, guardian.md (T002);
event-types.md (T003, leaned on its 15 pre-existing children rather than a
new split) — every `##`/`**bold**` section heading present before the change
is still present after it verbatim; only redundant/filler prose was cut, and
curriculum-ladder.md/social-fabric.md additionally compress their
already-split-off children's content down to a one-paragraph summary
(the facts stay findable in [[curriculum-ladder-progression]] and
[[social-fabric-conversations]], which already carried them in full).

## Exemptions

None. Of the 26 2026-07-29 gate findings: 14 body-size findings resolved by
tightening in place (10 in T001, 3 in T002, plus event-types.md in T003),
10 body-size findings resolved by a summary-style split (sim-state-apply-world.md
in T002; the other 9 in T003), and 2 capsule-budget findings resolved by
rewrite (guardian-faith.md's capsule, folded into its T003 split commit;
guardian-instruction-surface.md's capsule, standalone in T004). Zero
`size_budget_exempt` entries were needed anywhere in the corpus.

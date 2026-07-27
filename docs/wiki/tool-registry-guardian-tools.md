---
name: tool-registry-guardian-tools
description: Split from [[tool-registry]] (spec 014/021/029) — the spec-029 Guardian agency surface (send_vision/send_omen/monitor_and_act/cancel_order + pause/start/adjust_speed) with its authored array schemas, the authoritative miracle cost table (MiracleCost/MiracleCostsByEvent), RestrictEnum's per-world capability gating, and the derived GuardianToolGuidance/GuardianReadGuidance/GuardianTargetingGuidance prompt prose.
kind: component
sources:
  - internal/tool/registry.go
  - internal/tool/derive.go
verified_against: 657c770f87404b936a0587db1f6b00e81b9f0ee6
---

# Tool registry — the Guardian tool surface

Split from [[tool-registry]] (spec 014, TASK-53): the spec-029 agency surface
and the spec-021 miracle-cost/guidance derivations.

**The spec-029 Guardian agency surface** (`registry.go`, TASK-27,
[[guardian-orders]]): `nudge_dream`/`nudge_omen` are RETIRED; `guardianTools`
now declares, in order, `converse`, then `send_vision` (a waking vision for ONE
living villager at any hour — required `target` AgentName + required `text`,
`MaxBytes`/`TextCapBytes` 400, `Gate Charge`, `Events`
`metatron.nudged`/`agent.memory_added` — since spec 041 FR-014 also carrying
an OPTIONAL place-grant triple, `place_kind`/`place_x`/`place_y`, all riding
together (the handler refuses a partial triple); `place_kind` is an `Enum`
over `placeFactKinds` — [[mental-maps]]'s closed `PlaceFact` vocabulary
(spec 044 US4 added `grave`, the structure kind the `agent.died` reducer arm
places at a death tile), hand-mirrored here since `tool` must not import
`sim`, so a drift here can
only over- or under-offer the model, never land a false fact (the reducer
dry-run is the semantic authority) — and `Events` gains
`metatron.place_revealed` between the nudge and the memory), `send_omen` (required `targets` Text —
comma-separated living names or `"everyone"` — + required `text`, same cap/gate/
events; the night-only gate lives in the reducer, not the tool), then the
`Gate: None` order and meta tools. `monitor_and_act` is the SECOND authored-
`InputSchemaJSON` tool (`monitorAndActSchema`, arrays the scalar `Param` model
cannot express): a `{condition ≤300, action ≤400}` NL pair, a required
`event_types` string array (1..4 items) whose item `enum` is
`observableEventTypes` — a curated vocabulary of genuinely-emitted event
types (`agent.slept`/`woke`/`died`/`memory_added`/`intent_set`,
`social.conversation`/`promise_broken`/`rumor_told`, `gru.attacked`,
`norm.violated`, `sim.night_started`/`day_started`; the spec draft's
`meeting.norm_enacted` was dropped as un-emitted, `norm.violated` standing
in; spec 084 grows it 12 → 16 with the four `directive.*` lifecycle types —
enum-only, so standing orders watch the guardian's plan layer through the
unmodified matcher, [[guardian-designations]]),
an optional `keywords` array (≤6, each ≤40), a `confirm` boolean (marks a fuzzy
order needing a `metatron_watch` confirm, [[llm-orchestrator]]/[[cognition]]),
and `ttl_days` (1..7); its declared `Events` is `metatron.order_placed`.
`cancel_order` takes a required `id` Text, `Events` `metatron.order_cancelled`.
The three CHARGE-FREE meta tools `pause`, `start` (optional `speed` Enum), and
`adjust_speed` (required `speed` Enum) are `Effect Expressive` with EMPTY
`Events` — the `converse` precedent (acting cardinality applies, but nothing is
injected; the clock's own `clock.paused`/`clock.resumed` remain the record).
Their `speed` Enum is `clockSpeeds` (`"1x"`/`"4x"`/`"8x"`/`"16x"`/`"32x"`), a
hand-carried MIRROR of `internal/clock`'s ladder (`tool` is a leaf and cannot
import `clock`); `ClockSpeeds()` exports a copy and `internal/guardian`'s
`TestClockSpeedsMirrorLadder` pins it equal to `clock.CappedLadder()` — the
drift guard, same pattern as the sim-duration mirror.

**The miracle cost source and the spec-021 derivations** (`registry.go`,
`derive.go`; TASK-64): the per-kind miracle cost table is declared HERE, beside
`miracleKinds` — `MiracleCost(kind) (int, bool)` and `MiracleCostsByEvent()`
(kind↔event-type mapping, fresh map per call) are the ONE authoritative price
source: `sim.miracleCost` derives from `MiracleCostsByEvent()` (the import
direction already existed — [[guardian-miracles]]) and the guardian's prompt renders
costs from `MiracleCost`, so a price edit propagates to enforcement and prose in
one edit (`work_miracle.Cost.Charges` stays 1 — the Charge gate's minimum, not a
price). Two new derive.go surfaces serve [[guardian]]'s per-world capability
gating: `RestrictEnum(t, param, allowed)` returns a copy-on-write `Tool` whose
named Enum param keeps only the allowed values (registry never mutated; the
tool's own Enum order preserved; `InputSchema` of the restricted copy declares
only granted values), and `GuardianToolGuidance(roster)` renders the acting-tool
guidance prose — per tool its name, argument surface (from `Params`, the same
source `InputSchema` walks), and charge cost — replacing the hand-written prose
list `turnSystemPrompt` used to carry, so described ≡ declared ≡ priced by
construction (drift tests in `derive_test.go`). Since spec 036 the per-tool
description falls back to the tool's own `PromptGloss` when `guardianToolDesc`
has no entry — the branch bundle tools ([[bundle-tools]]) render through; it is
byte-inert for every map-covered built-in, pinned by the before/after
byte-identity test in `derive_test.go`. Since spec 063 ([[grounded-feedback]]),
`GuardianToolGuidance` also SKIPS every `Effect: Read` tool in the roster
(`explain`; `survey_site` since spec 084) — a read tool costs nothing and never consumes the turn's
one act, so listing it under the "call exactly ONE of these" acting doctrine
would misrepresent it; a sibling `GuardianReadGuidance(roster)` renders those
tools' own "you may also READ freely" paragraph instead (empty when the
roster grants none, byte-inert for every pre-063 roster, which never carried
a Read tool). Since spec 059 (US3), `derive.go` also
exports `GuardianTargetingGuidance()` — a static one-line prose pointer
("Aim your workings: …" — spec 052's display re-theming of the frozen
`work_miracle` tool family; the tool id itself never renames) that introduces
the miracle targeting digest in a
miracle-capable turn's prompt; it carries no data of its own (the tool
package has no world state to draw positions/passability from) — the digest
itself is assembled turn-side (`internal/guardian/turn.go`'s
`buildTargetingDigest`, [[guardian-orders]]/[[guardian-miracles]]), this
function is only the fixed prose that introduces it.

Spec 084 ([[guardian-designations]]) appends the five plan-layer tools
after `explain`, all `Gate: None` (the plan layer is charge-free — recorded
decision): `place_designation` (`kind` Enum over `designationKinds`,
`target` Text — a bare spec-082 locus parsed by `target.ParseLocus`,
`structure_kind` Enum over `buildableStructureKinds` — a hand-carried
mirror of sim's recipes-derived list, drift-pinned from internal/guardian —
`min_structures` Number 1..12, `label` Text ≤80; Events
`designation.placed`), `cancel_designation`/`cancel_directive` (required
`id`), `issue_directive` (`designation_id`, `targets` — the send_omen
comma-names/"everyone" vocabulary — `text` ≤400 runes, `ttl_days` 1..7
default 3; Events `directive.issued` + `agent.memory_added`), and the
Effect-Read `survey_site` (`x`/`y` required, `radius` clamped 1..8 default
4 — renders under `GuardianReadGuidance` like explain). All five join
`RosterGuardian`, `loopGuardianTools`, AND the stage-1 ceiling (the
`monitor_and_act` every-stage teaching-primitive precedent).

Spec 085 ([[guardian-faith]]) appends `prophesy` last: `Gate: Charge` (1 —
the send_vision price; the `prophecy.declared` reducer arm spends the
stake), `targets` (the send_omen vocabulary), `text` ≤400 bytes,
`claim_kind` Enum over `prophecyClaimKinds` (`designation_fulfilled`/
`structure_count`/`population_at_least`/`survives` — exported as
`ProphecyClaimKinds()` and drift-pinned from internal/guardian), the
kind-conditional claim params (`designation_id`/`structure_kind`/`min`/
`agent` — partial or foreign sets refused handler-side), and
`deadline_days` 1..7 default 3; Events `prophecy.declared` +
`agent.memory_added`. It joins both rosters and the stage-1 ceiling
(send_vision's profile — the same influence verb with a wager attached),
and there is deliberately NO cancel verb. `observableEventTypes` grows
enum-only 16 → 19 with the three `prophecy.*` types (`faith.changed`
deliberately stays out in v1).

## Connections

Part of [[tool-registry]]'s summary-style split (corpus-spec v2). See
[[tool-registry]] for the registry's overall doctrine and its other split-off
domains: [[tool-registry-world-catalog]] (the World/villager catalog),
[[tool-registry-schema-clamping]] (authored schemas and clamp-with-notice), and
[[tool-registry-derivation-rosters]] (derive.go/roster.go/validate.go).
[[guardian]]/[[guardian-orders]] read this note's registry entries for the
`send_vision` text cap, granted acting-tool guidance, and `work_miracle`
dispatch; [[guardian-miracles]] documents the exact cost/workings source this
note's `MiracleCost`/`miracleKindArgs` expose; [[grounded-feedback]] is the
spec-063 consumer of `explain`/`GuardianReadGuidance`.

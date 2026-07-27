---
name: testing-feedback-persona
description: Grounded-feedback mirror-drift pins and report-card producer/consumer suites (spec 063) plus the persona lifecycle suite (index-aligned maps, genesis seeding, secrets). Split out of [[testing-strategy]].
kind: pattern
sources:
  - internal/sim/toolcheck_test.go
  - internal/guardian/stage_test.go
  - internal/persona/persona_test.go
  - internal/tool/explain_test.go
  - internal/sim/explain_pin_test.go
  - internal/toolloop/explain_pin_test.go
  - internal/tui/explain_pin_test.go
  - internal/guardian/explain_test.go
  - internal/guardian/reportcard_test.go
  - internal/guardian/skin_battery_test.go
  - internal/sim/reportcard_test.go
  - internal/sim/rubric_hygiene_test.go
  - internal/tui/reportcard_test.go
verified_against: 657c770f87404b936a0587db1f6b00e81b9f0ee6
---

# Grounded-feedback & persona suites

**Grounded-feedback suites** (spec 063, TASK-115, [[grounded-feedback]]):
mirror-drift pins guard every leaf-package doctrine copy `explain` carries —
`internal/sim/explain_pin_test.go` (`TestExplainChargeDoctrineMirrorsSim`,
the charge-economy constants), `internal/toolloop/explain_pin_test.go`
(`TestExplainDecisionClassesMirrorVerdicts`, the verdict-name set), and
`internal/tui/explain_pin_test.go` (`TestExplainGlyphsMirrorLegend`, the map
legend) — each fails the build the moment its source of truth drifts from
`tool/explain.go`'s mirrored copy. `internal/tool/explain_test.go` covers
`ExplainSheet`'s six fixed topics plus per-tool detail, the grant
distinction (granted vs. cataloged-but-ungranted), and the unknown-topic
repairable-miss path. `internal/guardian/explain_test.go` proves the
`explain` handler always returns `VerdictReadOK` and never consumes the
turn's one act; `tutor_guide_test.go` proves `persona.TutorGuide` composes
only on a tutor-preset world and is otherwise byte-inert;
`internal/guardian/reportcard_test.go` and `internal/sim/reportcard_test.go`
cover the producer's activity-gated stopping-point triggers, citation
validation (an out-of-trail `seq N` drops the whole note), the run-end
card's `morgue.epilogue` routing vs. the non-run-ending
`guardian.report_card` event, and the reducer's latest-card-only retention;
`internal/sim/rubric_hygiene_test.go` and `internal/sim/toolcheck_test.go`'s
extended `TestWhitelistDiffIdentical` pin `guardian.report_card` as
exactly the spec-063 boundary widening (the run-end card's
`morgue.epilogue` routing leaves the ended-door narrowing untouched);
`internal/guardian/skin_battery_test.go` extends the adversarial skin
battery to the new `ExampleAsk`/`CeremonyChapter`-adjacent tokens.
`internal/tui/reportcard_test.go` and `help_guardian_test.go` cover the
console card seam's composition order (checklist first, note beneath) and
the D9 guardian section's stage-keyed, model-free byte-identity.
`internal/guardian/stage_test.go` and `guardian_test.go` extend their
existing roster-table/status assertions to include `explain` in the
stage-1/-2 ceiling and the default granted-tools list.

**Persona lifecycle suite** (`internal/persona/persona_test.go`, TASK-74): on
top of the pre-existing genesis-once/0444/missing-file-load coverage,
`TestPersonaMapsSweepAligned` proves the four index-aligned maps (`Texts`,
`Anchors`, `DriftMarkers`, `Secrets`) stay in lockstep with `sim.AgentNames` —
gaining or losing an entry in any one map fails the sweep;
`TestAnchorsMatchTemperamentLine` pins the documented "deliberately
identical" invariant between `Anchors` and each persona's `**Temperament:**`
line; `TestLoadUnreadableDegrades` proves an unreadable persona file degrades
`Load` to an empty string for that agent only, mirroring the missing-file
contract; `TestGenesisSeedsCharterAndJournal` proves fresh genesis seeds
`charter.md` (= `DefaultCharter`) and a rune-budgeted `journal.md` per agent,
and that an existing `charter.md` is never overwritten; and `TestSecretEvents`
proves the genesis `social.secret_seeded` events are index-aligned,
tick-0, tone `-70`, one per agent.

The whole suite runs under `-race`; it caught a real race (store `lastSeq`, loop
writer vs IPC readers — now atomic).

## Connections

Part of the [[testing-strategy]] suite map (split out during the corpus-spec v2
restructure); see that note for the full layered test picture and links to
sibling suites.

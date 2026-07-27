# Implementation Plan: agent-named payloads (spec 086)

**Branch**: `086-agent-named-payloads` (task branch:
`task-17-agent-named-payloads`) | **Date**: 2026-07-27 |
**Spec**: [spec.md](spec.md) |
**Census/enforcement/contracts**: [data-model.md](data-model.md)
(normative) | **Decisions**: [research.md](research.md)

## Summary

One repo-wide format migration plus one small rider, on one branch:
(1) the `AgentRef` type (struct marshal `{id,name}`, dual-shape
unmarshal, pure constructors over the `AgentNames` roster constant);
(2) the census migration — 66 payload types flip int→ref in place, 4
state-shared types split into payload mirrors, `faith.changed` gains the
additive died-agent ref, ~127 emission sites across
sim/mind/guardian/bundle/persona move to constructors; (3) mechanical
enforcement — `mustPayload` + `InjectSocial` append validation, the new
`sim.PayloadCatalog` with the doc-anchored `TestPayloadAgentRefSweep`,
`TestNoAgentRefInState`; (4) back-compat — arms read `.ID` only, no name
validation in `Apply` ever, pre-086 replay byte-identity proven by
fixture; (5) TUI consumption — payload-first digest naming with the
`names = nil` proof, the generic single-ref `resolveSubject` fallback
with the hit-rate test; (6) the reverse-jump rider — strip-glyph and
roster-row click → `centerCameraOn`, `J` keyboard parity, two
mouse-parity oracle entries, three design pages amended. Cross-cutting:
wiki re-pins (the event-types family is the big set), player docs,
`check-tui-design --changed`, merge-drift pr gate, merge-commit-only.

## Technical Context

**Language**: Go. **Modified packages** (cross-package format migration —
the recorded Opus 4.8 tier): `internal/sim` (new `agentref.go`,
`payloads.go`; payload declarations across agents.go, social.go,
mentalmap.go, consolidate.go, journal.go, cognition.go, plan.go,
plans.go, guardian.go, prophecy.go, governance.go, gru.go, miracles.go,
morgue.go, chronicle.go, faith.go, state.go; `mustPayload` and
`InjectSocial` validation; reducer arms' `.ID` reads; executor emission
sites), `internal/mind` (convo, consolidate, telemetry, handlers,
embedder, narrate, meeting — construction sites only),
`internal/guardian` (turn, miracle_batch, plans, orders, prophecy,
reportcard, charter — construction sites only), `internal/bundle`
(effects.go), `internal/persona` (files.go), `internal/tui` (digest.go
subject fallback + per-type name reads, grammar.go payload-first naming,
digest_test.go fixtures + new assertions, views.go strip/roster hit
regions, look.go handleMouse, tui.go `J` key + hit-region fields,
mouseparity_test.go entries). **Untouched by design**:
`internal/cognition`, `internal/clock`, `internal/llm`, `internal/store`
(Event shape/append path unchanged), `internal/world` (migration
untouched), `internal/ipc` (events push verbatim), every `Apply` arm's
FOLD semantics, all emission ORDER, stranger payloads. **No format
bump** (state shapes untouched — stronger than the omitempty precedent);
**no new RNG**; **no new event types**; **no tuning change**.

## Constitution Check (v1.2.0)

- **I. Artifact-grounded** — PASS: TASK-17 card (drift audit, reorient
  move 10, operator-placed rider), signed-off sweep runbook, CLAIM.md on
  this number; the AgentRef shape decision derives from the card's own
  language (research R1); every decision cites file:line evidence.
- **II. One task, one PR** — PASS: one branch
  (`task-17-agent-named-payloads`, worktree `.worktrees/task-17`), one
  PR; the census batches and the rider are internal breakdown.
- **III. Gates** — PASS: claim gate passed; pr gate +
  `check-tui-design --changed` + `TestCatalogSweep` + the new sweeps +
  mouse-parity sweep + `go test ./...` are the choke points; spec-bridge
  mirrors phases.
- **IV. Grounding freshness** — PASS (planned): Phase 8 re-pins every
  touched note in-branch (the event-types family carries the payload
  rows — the largest re-pin set this project has done; budgeted as its
  own phase) and regenerates `docs/player/`; merge is
  `gh pr merge --merge`; freshen by merging main INTO the branch, never
  rebase (TASK-160/161 landing laws; sibling lanes may be in flight).
- **V. Model tiers** — PASS: Opus 4.8 via `spec-implementer` (recorded
  on TASK-17: repo-wide payload migration + mechanical enforcement +
  back-compat replay). Planning/gating stays on Fable 5.

## Design decisions (file-level)

- **D1 — `internal/sim/agentref.go`** (new): `AgentRef` (struct marshal;
  custom dual-shape `UnmarshalJSON`), `Ref(i)`, `Refs(ids)`,
  `validateRefs(v any) error` — the shared reflection walk (in-roster ⇒
  exact roster name; out-of-roster ⇒ empty) used by both emission doors.
  Doc comment carries the R2/R3 laws (never in State; never validated in
  Apply).
- **D2 — `internal/sim/payloads.go`** (new): `PayloadCatalog`
  (event type → zero payload value, every type incl. mirrors and
  empty-struct types) with the doc-anchored completeness test beside it;
  `TestPayloadAgentRefSweep` + frozen vocabulary + allowlist (four
  entries, rationale strings); `TestNoAgentRefInState`.
- **D3 — census flips (data-model §3)**: field-type changes in place,
  json tags unchanged; every reducer arm reading a migrated field
  switches to `.ID`. Emission sites move to `Ref`/`Refs` — mechanical,
  compiler-driven (flipping the type breaks every construction site;
  the build is the checklist).
- **D4 — the splits (data-model §4)**: `DirectiveIssuedPayload`,
  `OrderPlacedPayload`, `ProphecyDeclaredPayload` (+ claim mirror),
  `DeathRef` for `RunEndedPayload.Deaths` — each defined beside its
  entity; arms fold `.ID`s; door dry-runs validate the mirrors.
- **D5 — `mustPayload` + `InjectSocial`**: `mustPayload` calls
  `validateRefs` before marshal (panic contract); `InjectSocial` decodes
  each payload via `PayloadCatalog` and calls `validateRefs` before the
  dry-run (batch refusal). No `Apply` arm changes validation.
- **D6 — `internal/mind` / `internal/guardian` / `internal/bundle` /
  `internal/persona`**: construction-site-only edits to build refs
  (`sim.Ref`/`sim.Refs` — the roster constant needs no state);
  guardian mirrors for order/directive/prophecy payloads per D4.
- **D7 — `internal/tui/grammar.go` + `digest.go` naming**: a
  `refName(names []string, r sim.AgentRef) string` helper (payload name
  first, `agentName` fallback); agent-bearing digest rows switch to it;
  `resolvePayloadNames` regex fallback untouched (historic rows).
- **D8 — `internal/tui/digest.go` subject fallback**: registry-first;
  generic pass scans payload JSON for ref objects (`{"id":…,"name":…}`),
  exactly-one-distinct-in-roster ⇒ candidate; `world.migrated`
  hard-excluded stays. Hit-rate test in digest_test.go per data-model §7.
- **D9 — the rider (`internal/tui`)**: `stripHit`/`rosterHit` region
  fields + frame-top invalidation (the chronHit pointer pattern);
  `villagerStripView`/`villagerRosterBody` record geometry;
  `handleMouse` branches (strip → jump; roster → select + jump + narrow
  pane switch); `handleVillagersKey` gains `J`; strip standing-resolution
  comment amended; two `mouseParityOracle` entries with checks in the
  `checkChronicleJumpToSourceMouseClaim` shape.
- **D10 — docs**: `docs/design/tui/panels/villager-strip.md` (control row
  `— · click glyph`, display-only prose amended, keyboard-path note),
  `panels/villagers.md` (roster-row cell gains click; new `reverse-jump`
  row `J · click row`; parity note updated), `patterns/keymap.md`
  (villagers `J` row) + any page `check-tui-design --changed` names;
  wiki + player docs per Phase 8.

## Testing strategy

- **Type**: AgentRef marshal shape (fixed order, unicode name fixture);
  dual-shape unmarshal (bare int, object, []AgentRef over both, pointer);
  constructors incl. sentinels and out-of-range.
- **Enforcement (SC-002)**: sweep tests + mutation checks — unnamed
  in-roster ref panics `mustPayload` / refused by `InjectSocial`;
  synthetic vocabulary-tagged bare int fails the sweep; allowlist entries
  are live (removing a use without its entry fails); catalog completeness
  vs doc backticks; `catalogFixture` ⊆ `PayloadCatalog`;
  `TestNoAgentRefInState`.
- **Emission (SC-001)**: fixture drives per family asserting named refs
  from log bytes alone — executor families (intents/harvest/needs/died/
  neglect/talk/storage/walls/crafting), hail, governance (attendees/
  yeas/nays/witnesses), gru, mind-injected (memory/consolidation/convo/
  rumor incl. subject, journal), guardian-injected (nudged targets,
  order_placed −1 sentinel, directive.issued, prophecy.declared incl.
  claim agent 0, item_granted, epilogue), bundle/persona sites,
  faith.changed villager_died additive ref.
- **Back-compat (SC-003)**: checked-in pre-086 fixture log → from-genesis
  replay byte-identity (`Marshal`/`Hash`); pre-086 snapshot decode +
  stored-hash verify; `world.migrated` fixture unchanged; mixed-era
  render test (old rows fallback names, new rows payload names, grammar-
  miss regex path); arms accept legacy shapes (no name validation
  anywhere in Apply — asserted by replaying unnamed injected rows).
- **TUI naming + hit rate (SC-004)**: `TestCatalogSweep` with rewritten
  named fixtures + the `names = nil` identical-output assertion;
  hit-rate test (registry+fallback > registry-only; pins
  `journal.entry_written`, `morgue.epilogue`, `cog.thought`); multi-ref
  registry-miss stays unlocatable.
- **Rider (SC-005)**: two oracle entries through real mouse dispatch
  (pan moved; roster also selection moved); `J` key test (roster +
  detail; dead villager → grave coords; nil replica no-op); overflow-
  marker click no-op; narrow-layout pane switch; mouse-parity sweep both
  directions; `check-tui-design --changed` exit 0.
- **Determinism**: same-seed double-run event-history byte-equality on
  the post-086 build (named both — the byte-comparable doctrine holds
  within a build); no emission-order diffs (scope guard FR-012).

## Wiki re-pin set (Phase 8; the pr gate is the authority)

This PR's biggest grounding load is the **event-types family**: the
parent note's conventions section (payload-struct convention gains the
AgentRef law) and EVERY domain child whose payload rows show shapes —
`event-types-agent-intents`, `-agent-vitals`, `-mental-map`,
`-harvesting-consumption`, `-crafting-building`, `-social-memory`,
`-memory-consolidation`, `-social-protocol`, `-cognition-telemetry`,
`-guardian-orders`, `-guardian-morgue`, `-guardian-actions`,
`-guardian-plans`, `-scenario-incidents` (stranger rows re-verified
unchanged), `guardian-faith` (faith.changed additive ref) — plus
`event-log` (payload convention sentence), `sim-state-reducer` + affected
children (arm `.ID` reads; the no-names-in-state law),
`sim-loop-injection-doors` (door validation), `tui-chronicle-feed`
(payload-first naming, subject fallback), `village-lens` (strip
affordance amendment), `tui-villagers-tab` (`J`, row click),
`tui-client-mechanics`/`tui-look-cursor` (mouse routing, if their sources
moved), `social-fabric*`/`mental-maps`/`cognition`/`governance`/
`gru`/`morgue`/`guardian-orders` per their `sources:` frontmatter against
the actual diff. `docs/player/` regenerated in-branch
(`node .claude/skills/player-docs/scripts/check-freshness.mjs --check`
green). The pr gate computes the true set; this list is the budget.

## Risks / mitigations

- **Sheer diff breadth** (~30 files of payload/emitter edits): the type
  flip makes the compiler the checklist — nothing half-migrated builds;
  the census table is the review artifact; tasks.md batches by package
  with a green `go build ./...` required per batch.
- **A missed payload type**: three nets — the compiler (type flip), the
  doc-anchored catalog completeness test, and the vocabulary sweep. A
  type absent from census AND doc AND catalog cannot carry events today
  (nothing emits it).
- **Accidental state-shape change** (the hash hazard):
  `TestNoAgentRefInState` + the pre-086 replay byte-identity fixture
  catch it mechanically; the split law (D4) is where the discipline
  concentrates — review obligation named in the PR body.
- **Injected-row replay regression** (name validation leaking into an
  arm): explicit test replays unnamed injected rows through Apply; the
  R3 law is stated at `validateRefs`'s doc comment.
- **TUI render diffs for historic rows**: fallback path is
  regression-tested with a mixed-era fixture; `resolvePayloadNames`
  untouched.
- **Rider scope creep** (strip cursor, follow-mode, new camera state):
  FR-009/010 bound it — one mouse affordance + one key, all through
  `centerCameraOn`; no selection state on the strip.
- **Merge drift vs sibling lanes**: freshen by merging main INTO the
  branch only; `check-merge-drift pr` before the PR; board moves at root
  only (TASK-161 board-sync exception); no rebases anywhere.

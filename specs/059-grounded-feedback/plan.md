# Implementation Plan: Grounded feedback layer — explain tool, tutor guide, report card

**Branch**: `059-grounded-feedback` | **Date**: 2026-07-25 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/059-grounded-feedback/spec.md`

## Summary

Three deliverables over one grounding set: (a) an `explain` registry tool —
read-only, expressive-class, grant-gated, returning fact sheets composed
tool-side from the live tool registry + doctrine constants, scoped to the
world's effective grant/stage ceiling; (b) a compiled-in tutor guide
(`persona` package, TutorCharter precedent) composed in the editable zone on
tutor-preset worlds; (c) a report-card producer in `internal/guardian` (the
digest-worker notify-consumer precedent) firing at stopping points, grading
on the recorded trail via a new cheap route kind, storing its attribution
note durably, rendered through spec 053's console card seam (+ TASK-127's
checklist renderer) and inside the postmortem. Plus the D9 `?` guardian
section (static-per-stage, model-free) and skin-token compliance throughout.

## Technical Context

**Language/Version**: Go 1.24 (repo toolchain; no new deps)

**Primary Dependencies**: `internal/tool` (registry + derive), `internal/guardian` (turn pipeline, digest-worker pattern), `internal/persona` (guide constant), `internal/llm` (new cheap route kind, config plumbing), `internal/sim` (whitelist/reducer for the card event; rubric-hygiene sweep), `internal/tui` (card production into the seam, `?` section), `internal/skin` (new tokens per contract §4)

**Storage**: the attribution note rides recorded prose channels (plan R5: a whitelisted `guardian.report_card` prose event pre-end; the run-end card rides the existing `morgue.epilogue` path post-end) — stored once, re-read, never re-graded

**Testing**: `go test -race ./...`; explain ground-truth sweep over the topic catalog (SC-001); tutor-lane neutrality tests (charge/event/frame diffs — SC-002); prompt-composition byte-identity for non-tutor worlds; citation-resolution fixtures (SC-004); `?`-section byte-identity (SC-005); rubric-hygiene sweep (FR-003); adversarial re-run of the spec-052 battery (prompt surface grew)

**Target Platform**: daemon + terminal client

**Project Type**: single Go module, cross-package

**Performance Goals**: explain is O(registry) string composition; the critique is one cheap-chain call per stopping point, debounced, budget-capped

**Constraints**: fixed frame last on every path (spec 021); no initiative-frame changes; tutor-lane zero-cost by construction; deterministic degradation without the chain; skin contract §4 for every new string; same-PR design-doc amendments

**Scale/Scope**: est. 1,200–2,000 LOC incl. tests; 3–4 design pages amended

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Artifact-Grounded Action** — PASS: spec dir `specs/059-*`; TASK-115 linked pre-implementation; the two standing resolutions cite spec-046 lock doctrine and D5/127's renderer.
- **II. One Task, One PR** — PASS: `.worktrees/task-115`, one branch, one PR.
- **III. Gates Over Assertions** — PASS: ground-truth/neutrality/citation sweeps are mechanized ACs; design gate; spec-bridge gate; merge-drift pr gate.
- **IV. Grounding Freshness** — PASS (planned): touches sources of `tool-registry.md`, `metatron.md` (renamed note), `tui-client.md`, `llm-orchestrator.md`, `event-types.md` → wiki-update + player-docs re-ground.
- **V. Model-Tiered Workflow** — PASS: planned on Fable 5; implementation on **Opus 4.8** — rubric: guardian turn-pipeline/prompt-composition changes (injection-adjacent), new route kind through `internal/llm`, cross-package — senior tier per the runbook Lane 3 assignment.

**Post-Phase-1 re-check**: PASS — one new tool, one new route kind, one new prose event type + guide constant; every mechanism follows a named precedent. No Complexity Tracking entries.

## Project Structure

### Documentation (this feature)

```text
specs/059-grounded-feedback/
├── plan.md
├── research.md          # R1 explain composition; R2 tutor-lane mechanics; R3 guide seam;
│                        #   R4 cheap chain + budget; R5 card storage/doors; R6 ? section;
│                        #   R7 skin tokens; R8 design-page set
├── data-model.md
├── quickstart.md
├── contracts/
│   └── feedback-layer.md  # explain topics/fact-sheet schema; tutor-lane invariants;
│                          #   card composition + stopping-point grammar; ? section shape
└── tasks.md
```

### Source Code (repository root)

```text
internal/tool/           # explain declaration (expressive, Events empty, read-only);
│                        #   fact-sheet composition from registry + doctrine reads
internal/guardian/       # handler wiring (granted subset); turn one-act exemption for
│                        #   read-only tools; report-card producer (digest-worker pattern:
│                        #   stopping-point notify consumer, debounce, cheap-chain call,
│                        #   storage via R5 doors); tutor-guide composition (preset-scoped)
internal/persona/        # TutorGuide compiled constant (TutorCharter sibling)
internal/llm/            # new route kind (report-card class) + config plumbing
│                        #   (metatron_watch precedent); default routing to cheap chain
internal/sim/            # guardian.report_card whitelist + reducer (latest-card state);
│                        #   rubric-hygiene sweep test; catalog/digest entries
internal/tui/            # console card production from state/events into the spec-053 seam
│                        #   (+ badge); postmortem embedding (127's renderer contract);
│                        #   help.go D9 guardian section
internal/skin/           # new tokens (default table + doc twin + completeness test)

docs/design/tui/
├── overlays/help.md         # D9 guardian section content contract (deliberate amendment)
├── pages/guardian-console.md# card production real — seam note updated
├── patterns/skin-tokens.md  # new tokens in the doc twin
└── re-pins on all touched pages
```

**Structure Decision**: producer daemon-side in `internal/guardian` (cards
must be durable and client-independent); rendering client-side over
state/events; the explain tool lives beside its registry ground truth.

## Complexity Tracking

No constitution violations — table intentionally empty.

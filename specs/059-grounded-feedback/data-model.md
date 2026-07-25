# Data Model: Grounded feedback layer (spec 059)

## Fact sheet (explain result — derived, never stored)

Per-topic deterministic composition: {topic, sections[] of labeled fact
rows, catalog-of-topics on miss}. Scoped to the world's effective grant +
stage ceiling; byte-stable for identical (registry, grant, stage) inputs.
Topics v1: `roster`, `costs`, `charges`, `workings` (kinds+prices),
`decisions`, `glyphs`, `<tool-id>` detail.

## Tutor guide (compiled constant)

`persona.TutorGuide` — ≤4,000 chars, tutor-preset-scoped composition in the
editable zone. Not a file, not player-editable, never binds through
`skills/`.

## Report card (produced + stored)

| Piece | Source | Determinism |
|---|---|---|
| rubric checklist | exercise rubric evidence (TASK-127 renderer) | deterministic |
| attribution note | cheap-chain critique over the shared data source | LLM-authored, STORED once |
| citations[] | recorded event seqs + charter fingerprint | validated against the log |

**Storage**: pre-end → `guardian.report_card{fingerprint, note,
citations[]}` (whitelisted prose event; reducer keeps latest on state, log
keeps history); run-end → rides the existing run-end epilogue path. Cards
are re-read, never re-graded.

**Stopping points**: run end · exercise resolution · pause episode
(debounced: once per episode, only with new guardian activity since the
last card).

## `?` guardian section (derived per render)

{stage identity, effective verb list from the status grant summary, one
example ask per verb} — byte-identical for identical status; model-free.

## New skin tokens

Card labels, guide framing, per-verb example asks — default table + doc
twin + completeness test per the skin contract §4.

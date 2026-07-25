# Quickstart: validating spec 046 end-to-end

Prerequisites: `go build ./... && go test ./...` green. References: contracts/
{stage-gating,unlocks-record,events,exercises}.md, data-model.md.

## 1. Hermetic proof

```bash
go test ./internal/world -run Stage -v          # manifest round-trip, validation, absent=ungated
go test ./internal/metatron -run 'Stage|Ceiling|Preset' -v  # ceiling intersection, stage-1 lock+notice
go test ./internal/worlds -run Unlocks -v       # record read/heal/write, evidence entries
go test ./internal/sim -run Curriculum -v       # event arms, fixture pass→unlock chain
go test ./internal/tui -run TestCatalogSweep -v # curriculum.* fully cataloged
```

## 2. Creation UX walkthrough

```bash
promptworld stages                    # 4 identities, concepts, gates; none earned yet
promptworld new v1 --stage stage-1    # tutor preset seeded by default
promptworld new v2 --stage stage-3    # ERROR: informed message naming skipped concepts
promptworld new v2 --stage stage-3 --override   # proceeds; manifest records override
promptworld status v1 --json          # .world.stage == "stage-1"
```

## 3. Gating proof (US2 / SC-002)

On `v1` (stage-1): agent roster contains only the base set; edit `charter.md` → next
turn reply carries the does-not-bind notice; metatron status shows the lock. On a
stage-3 world: skills compose, capabilities.json honored within ceiling. Same-seed
stage-1 vs stage-4 worlds driven identically → world-event histories identical
(cross-stage determinism diff).

## 4. Unlock chain (US3 / SC-003, SC-004 — fixture-driven until TASK-119)

In-test: emit `curriculum.exercise_passed` (first-night fixture) → reducer records
pass → `curriculum.stage_unlocked` → chronicle line renders + `unlocks.json` gains a
stage-2 entry pointing at the world + seqs → `promptworld stages` shows it earned →
`new --stage stage-2` proceeds without override. Negative: a the-law pass fixture
with default-charter evidence → NO stage-3 unlock (SC-004).

## 5. Docs (US5 / SC-006)

```bash
node .claude/skills/player-docs/scripts/check-freshness.mjs --check
```

Four stage quickstart pages exist (13 total), pinned, gate green; each names its
stage's concept, grants, and unlock evidence in plain language.

## 6. Tutor preset (US5)

`promptworld new t1 --stage stage-1` on an LLM world → the agent's early turns include
orientation unprompted. Same world with no llm.json → gating/presets/unlocks all
function; only the voice is absent.

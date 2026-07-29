# Tasks: agent.intent_failed for non-build goals (TASK-95)

**Input**: `specs/096-intent-failed-loud/spec.md`

## Phase 1: Event + emission

- [X] T001 Define `agent.intent_failed` (goal, enumerated reason, position;
  emitter-computes) and emit it at every invalid exit the card enumerates
  (forage/chop/hunt/demolish/repair/quarry/cook/bathe) and every contested no-op
  re-check (craft/cook/bathe/deposit/withdraw) in internal/sim/executor.go
  (FR-001).
- [X] T002 Situated failure memory (internal/sim/memory.go, build-failure shape)
  + reducer closure through stampIntentOutcome + mind re-arm list entry
  (FR-002/FR-003).

## Phase 2: Surfaces

- [X] T003 TUI digest grammar entry + event-types.md documentation; whitelist
  additions if any door requires them; TestCatalogSweep green (FR-004).

## Phase 3: Tests + grounding

- [X] T004 Failure test matrix (every enumerated goal × invalid/contested where
  both exist; explicit deep coverage for one gather + one station goal per card
  AC#2); replay byte-identity for existing fixtures; go test -race ./... green
  (FR-005/FR-006).
- [ ] T005 Wiki re-pins (executor*, event-types*, sim-state-intent-lifecycle
  notes — prose amendments where the failure taxonomy is described); player-docs
  probe; merge-drift pr gate exit 0.

# Tasks — spec 109, default local model

Board task: **TASK-184**. Branch: `task-184-default-local-model`.
Dispatch tier: Sonnet (`claude-sonnet-5`, `.claude/agents/spec-implementer.md`).

## Phase 1: Default config

- [X] T001 `internal/llm/config.go` `DefaultConfig()`: local provider `Model`
      `cogito:3b` → `gemma4:latest`, `ToolMode` `"json"` → `"native"`. Do not add
      `reasoning_effort` (zero-priced providers already resolve it to `"none"`), and do
      not touch `defaultRoutes()` or the `cloud` provider.
- [X] T002 Rewrite the justification comment above that literal (config.go:461-466). It
      currently recommends `gemma4:12b-mlx` as the upgrade path — the opposite of what is
      now known. It must record: build format decides whether schema constraints are
      honored; Ollama MLX builds silently discard them; `gemma4:latest` is gguf, honors
      schemas, and tool-calls natively. Retain the TASK-52 fact that cogito:3b needs
      `"json"` mode.

## Phase 2: Tests

- [X] T003 Confirm `cmd/promptworld/commands_test.go:167` still passes **unmodified** —
      it derives the expected `ollama pull` line from `llm.DefaultConfig()` and is the
      FR-004 guard. Needing to edit it means the guidance line was hard-coded: a regression.
- [X] T004 Update only those tests that genuinely assert the default model or tool mode.
      Leave unrelated fixtures using `cogito` as an arbitrary provider name untouched
      (`commands_test.go` status/preflight cases, `calibrate_test.go`,
      `internal/llm/preflight_test.go`).

## Phase 3: Operator documentation

- [X] T005 `docs/llm-providers.md` lines 25-29: update the stated default model and
      `tool_mode`, and remove `gemma4:12b-mlx` as a recommended upgrade path.
- [X] T006 `docs/llm-providers.md` lines 44-47: re-cast the v2 registry worked example off
      `gemma4:12b-mlx`; use `qwen3.6:latest` so the example doubles as the documented
      upgrade for capable machines.
- [X] T007 `docs/llm-providers.md` line 81: `tool_mode` table row — keep the cogito:3b
      `"json"` note, add that `"native"` is the default because the shipped model supports it.
- [X] T008 New subsection documenting the hazard: MLX builds
      (`details.format: safetensors`) accept and silently discard schema constraints; how
      to check via `/api/show`; the symptom (prose where JSON was demanded → downstream
      parse failures and abandoned work); and that it is silent. Cite spec.md's measured
      table. Leave `docs/design/evidence/**` untouched — those are historical records.

## Phase 4: Grounding (spec 069 — pr gate blocks without this)

- [X] T009 For each wiki note pinning `internal/llm/config.go` —
      `llm-provider-registry`, `llm-orchestrator`, `llm-chain-walk-dispatch`,
      `llm-provider-health`, `nightly-consolidation`, `guardian-report-card`,
      `guardian-order-triggering` — read `git diff <pin>..HEAD -- internal/llm/config.go`
      and classify RE-PIN-ONLY or NEEDS-REVIEW. Amend prose before re-pinning anything
      classified NEEDS-REVIEW. Never re-pin without reading the diff.
      **Deviation:** `nightly-consolidation.md` does not pin `internal/llm/config.go`
      (its `sources:` list is `internal/sim/consolidate.go`, `internal/mind/consolidate.go`,
      `internal/mind/validate.go`, `internal/mind/retry.go`, `internal/mind/nightreport.go`,
      `internal/persona/personas.go`) — this task's own line names it in error; the
      `check-merge-drift.mjs pr` gate never flagged it, and it was left untouched. The
      other 6 notes were re-pinned; see final report for the RE-PIN-ONLY/NEEDS-REVIEW
      classification and reasoning per note.
- [X] T010 Regenerate `docs/player/` if `docs/wiki/` changed; verify with
      `node .claude/skills/player-docs/scripts/check-freshness.mjs --check`.

## Phase 5: Verification and PR

- [X] T011 `go build ./...`, `go vet ./...`, `go test ./...` all green.
- [X] T012 SC-001 live check: create a throwaway world OUTSIDE `~/.promptworld/` with an
      untouched `llm.json`; confirm the guidance line names `gemma4:latest`; confirm one
      schema-constrained call returns parseable JSON and one tool-calling call returns a
      well-formed tool call against the local endpoint. Record the observed output in the
      PR body. Tear the world down afterwards.
- [X] T013 `node scripts/check-merge-drift.mjs pr` from the worktree, exit 0, then
      `gh pr create`. Re-run the gate after any merge-in from main.

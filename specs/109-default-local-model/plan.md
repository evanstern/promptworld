# Plan — spec 109, default local model

**Constitution check:** `.specify/memory/constitution.md` is ratified (v1.3.0). Relevant
principles:

- **Principle V (Model-Tiered Workflow).** This slice is single-package
  (`internal/llm`) plus doc reconciliation and test updates — no cross-package or
  architectural surface, no concurrency/scheduling/governor logic, and the diagnosis is
  complete at file:line. That is the Sonnet rubric verbatim, so implementation dispatches
  to `.claude/agents/spec-implementer.md` (`claude-sonnet-5`). Escalation is not
  indicated.
- **Artifact-grounded action.** The measurements this spec rests on are recorded on
  TASK-184's card and reproduced in `spec.md`'s table, not left in a chat turn.
- **Spec 069 (wiki-in-PR).** `internal/llm/config.go` is a pinned source in seven wiki
  notes (see Phase 4), so the branch itself must re-verify and re-pin them; the pr gate
  blocks otherwise.

## Approach

The change is deliberately small and mostly declarative. One provider literal, one
comment that must stop being wrong, and a documentation page that currently steers
operators onto the broken path.

### Phase 1 — the default itself

`internal/llm/config.go`, `DefaultConfig()` (currently line 467):

```go
"local": {Transport: ProviderOpenAICompat, Endpoint: "http://localhost:11434/v1",
          Model: "gemma4:latest", Parallel: 4, ToolMode: "native"},
```

Two changes: `Model` `cogito:3b` → `gemma4:latest`, and `ToolMode` `"json"` → `"native"`.

The comment above it (config.go:461-466) is load-bearing and currently records the
opposite conclusion — it justifies cogito:3b and points at `gemma4:12b-mlx` as the upgrade
path. Rewrite it to record what is now known: the build format determines whether schema
constraints are honored, MLX builds silently ignore them, and `gemma4:latest` is chosen as
a gguf model that both honors schemas and tool-calls natively. Keep the existing
TASK-52 note that cogito:3b needs `"json"` — it stays true and explains why the old value
was what it was.

Do **not** touch `reasoning_effort`. Zero-priced providers already resolve it to `"none"`
(`providers.go:628-631`); adding it explicitly would imply a change where there is none.

Do **not** touch `defaultRoutes()`. Route policy is out of scope.

### Phase 2 — tests

- `cmd/promptworld/commands_test.go:167` reads
  `llm.DefaultConfig().Providers["local"].Model` and asserts the `ollama pull` line is
  derived from it. This must keep passing **unchanged** — it is the FR-004 guard. If it
  needs editing, the guidance line has been hard-coded and that is a regression.
- Search `internal/llm` for tests asserting the default model or tool mode and update
  expectations. Note most `cogito` hits in the test tree are unrelated fixtures (fake
  provider names in status/preflight tests) — those must be left alone. Change only what
  actually asserts `DefaultConfig`.

### Phase 3 — documentation

`docs/llm-providers.md` is the operator-facing authority and is the actual defect:

- **line 25** — the "default local provider is `cogito:3b` with `tool_mode: "json"`"
  sentence, and the `ollama pull` line at 28.
- **line 29** — remove `gemma4:12b-mlx` as a recommended upgrade path.
- **lines 44-47** — the v2 registry worked example uses `gemma4:12b-mlx`. Re-cast it.
  `qwen3.6:latest` is the natural replacement and doubles as the documented upgrade.
- **line 81** — the `tool_mode` table row. Keep the cogito:3b `"json"` note; add that
  native mode is the default because the shipped model supports it.
- **new subsection** — the hazard, stated plainly enough that it is actionable:
  Ollama's MLX builds (`details.format: safetensors`) accept and silently discard
  `response_format` / `format` schema constraints. Give the check
  (`curl /api/show -d '{"model":"..."}'` and read `details.format`), the symptom (prose
  where JSON was demanded, surfacing downstream as parse failures and abandoned work),
  and the point that it is silent — nothing in the daemon or calibration reports it.
  Cite the measured table from `spec.md`.

Leave `docs/design/evidence/**` alone. Those are historical records of runs that really
did use those models; rewriting them would falsify the record.

### Phase 4 — grounding (spec 069, gated)

`internal/llm/config.go` is a pinned source in: `llm-provider-registry.md`,
`llm-orchestrator.md`, `llm-chain-walk-dispatch.md`, `llm-provider-health.md`,
`nightly-consolidation.md`, `guardian-report-card.md`, `guardian-order-triggering.md`.

For each, classify against the actual diff (`git diff <pin>..HEAD -- internal/llm/config.go`):

- **RE-PIN-ONLY** where the diff provably cannot invalidate the prose (most of these
  notes describe registry/dispatch mechanics, not which model is default).
- **NEEDS-REVIEW** where the note quotes the default model or tool mode — amend the prose
  first, then re-pin.

Then regenerate `docs/player/` if `docs/wiki/` changed
(`node .claude/skills/player-docs/scripts/check-freshness.mjs --check` is the gate probe).

### Phase 5 — verification

`go build ./...`, `go vet ./...`, `go test ./...`. Then the live check for SC-001: create
a throwaway world **outside** `~/.promptworld/` with an untouched `llm.json`, confirm the
`ollama pull` guidance line names `gemma4:latest`, and confirm a schema-constrained call
and a tool-calling call both succeed against the local endpoint. Tear the world down.

Finally `node scripts/check-merge-drift.mjs pr` from the worktree — exit 0 before
`gh pr create`, and again after any merge-in.

## Risks

- **The comment is the real deliverable.** A future reader who changes this default again
  will read config.go:461, not this spec. If the comment does not explain the MLX hazard,
  the mistake recurs.
- **Over-broad find-and-replace on `cogito`.** The test tree uses `cogito` as an arbitrary
  provider name in unrelated fixtures. Blind replacement breaks tests that have nothing to
  do with this change.
- **Wiki re-pin honesty.** A merge-commit re-pin that skips reading the diff turns the
  freshness gate green while leaving prose that contradicts main. Classify every note.

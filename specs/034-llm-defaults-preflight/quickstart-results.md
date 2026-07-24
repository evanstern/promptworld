# Quickstart validation results — spec 034 (T018)

**Date**: 2026-07-24 | **Branch**: `task-84-llm-defaults-preflight` @ 8fcc79c |
**Environment**: darwin, local Ollama at `localhost:11434` with `cogito:3b` pulled
and — fittingly — **no** `gemma4:12b-mlx`: this machine IS the fresh-machine case
the old default silently died on. Binary built from the branch; test worlds under
the session scratchpad, stopped and deleted after the run.

| Scenario | Result |
|----------|--------|
| V0 unit/integration suite | **PASS** — `go vet ./...` clean; full `go test -count=1 ./...` exit 0, zero failures |
| V1 dead tier is loud | **PASS** (see below) |
| V2 endpoint unreachable | **PASS** (distinct wording) |
| V3 tool-silent provider | **PASS** (see below) |
| V4 fresh world works out of the box | **PASS** — 15 `cog.tool_call` events within ~3 min, zero config edits |
| V5 docs alignment | **PASS** — cogito:3b + json identical across DefaultConfig, docs/llm-providers.md, README; no doc presents gemma4 as the fresh default |

## V1 — absent model (`bogus-model:1b`)

World booted normally (never a boot error). Within 8s of start:

- daemon.log: `daemon: WARNING llm provider "local": model "bogus-model:1b" not served by http://localhost:11434/v1 — ollama pull bogus-model:1b`
- `promptworld status` (human): same WARNING line after the clock line
- `status --json`: `condition: "model-missing"` + detail + remedy on the local provider
- durable event: `daemon.llm_warning` with `kind=model-missing, active=1` in the event log

Clear-without-restart was validated at the unit level (compressed-cadence re-probe
tests in preflight_test.go); not re-run live — pulling a real model mid-run was the
only live path and wasn't worth the bandwidth given the test coverage.

## V2 — unreachable endpoint (`localhost:9`)

World booted and ran. `WARNING llm provider "local": endpoint http://localhost:9/v1
unreachable — start the model server at http://localhost:9/v1` in both daemon.log
and status — wording distinct from V1 as the contract requires.

## V3 — tool-silent (cogito:3b with tool_mode removed → native)

World created from the new default with `tool_mode` deleted. First planner cadence
fires at 06:30 game time; after the 8th consecutive tool-free tool-carrying
completion the detector raised: WARNING naming the provider with the
native-mode remedy (`set providers.local.tool_mode to "json" and restart`) in
daemon.log + status, `condition: "tool-silent"` on the wire. (Detector unit
coverage: threshold, reset, non-tool exclusion, precedence, clear — detector_test.go.)

## V4 — fresh world out of the box

`promptworld new` printed the new guidance line
(`local model: cogito:3b — pull it first if you haven't: ollama pull cogito:3b`);
the written llm.json matched contracts/fresh-world-defaults.md exactly (cogito:3b,
tool_mode json, parallel 4; cloud + routes unchanged). Started with zero edits:
no warnings anywhere, `cog.tool_call` count 15 after ~3 minutes — SC-002 met.
Healthy `status` output carries no WARNING block and no condition fields on the
wire (omitempty verified: keys absent in `--json`).

## SC coverage

- SC-001 (warning ≤30s, all surfaces): V1 — 8s, log + status + wire (TUI badge
  unit-tested; not attached live).
- SC-002 (fresh world, zero edits): V4.
- SC-003 (no false positives): V4 ran warning-free; long-soak evidence deferred to
  normal operation (detector requires 8 *consecutive* tool-free completions, and
  TASK-73 healthy soaks sustained tool-call streams).
- SC-004 (alignment): V5 greps.
- SC-005 (tool-silent flagged in minutes, remedy actionable): V3.

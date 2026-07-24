# Quickstart Validation: Scriptable Agent Tools

**Feature**: [spec.md](spec.md) | prove the feature end-to-end after implementation.
Contracts referenced: [bundle-manifest](contracts/bundle-manifest.md),
[script-api](contracts/script-api.md), [boot-validation](contracts/boot-validation.md).

## Prerequisites

- Built binary (`go build ./...` clean; `go test ./...` green).
- A scratch world: `promptworld world create <dir>` (or the current create command).
- An LLM config is NOT required for most checks — tool registration, boot validation, and
  injection can be exercised without a live provider; the invocation scenarios use the
  existing test/mock provider patterns where possible.

## Scenario 1 — Declarative bundle loads and lands (US1 / SC-001)

1. Create `<world>/bundles/demo/tools/teleport/tool.json` per the manifest contract:
   `move_entity` on `{args.target}` to `{args.x}`,`{args.y}` + `narrate` "…vanished in a poof
   of smoke" to `all_living`; `events`: `metatron.entity_moved`, `agent.memory_added`.
2. Boot the world daemon. **Expect**: boot log shows the bundle loaded, zero BootReport errors;
   the metatron roster (status surface / debug) lists `teleport` with the derived schema.
3. Invoke `teleport` (metatron turn or test harness). **Expect**: villager position changes;
   every living agent gains the narration memory; event log contains only the two declared
   types; charge balance reflects the reducer's price for `entity_moved`.

## Scenario 2 — Boot validation rejects loudly (US1 / SC-005, boot-validation ladder)

1. Add `bundles/bad/tools/hax/tool.json` declaring `events: ["metatron.heal"]`.
2. Boot. **Expect**: BootReport error naming `bundles/bad/tools/hax/tool.json`, rule T3, and
   the offending value; world boots; `teleport` still available; `hax` absent from roster.
3. Repeat with a malformed `capabilities.json` in the bundle root. **Expect**: whole bundle
   rejected (B3), other bundles unaffected.

## Scenario 3 — Scripted tool, sandbox, caps (US3 / SC-006)

1. Create `bundles/demo/tools/cast_light/` with `tool.star` branching on
   `world.time_of_day` (script-api contract example) and `tool.json` with
   `"script": "tool.star"`, `limits.max_steps: 100000`.
2. Boot; invoke at night and at day (drive the clock with `snap_time` or test harness).
   **Expect**: different narration per branch; only declared events land.
3. Point the manifest at a script with `while True: pass`-equivalent loop
   (`for _ in range(...)` exhaustion). **Expect**: deterministic step-cap abort, no state
   change, descriptive failure fed back to the invoker, no charge spent.
4. Script attempts an undeclared effect kind / event type. **Expect**: compiler rejection,
   nothing lands.

## Scenario 4 — Determinism / replay byte-identity (SC-003, FR-011)

1. Run the new `TestBundleToolReplayByteIdentity` (+ scripted variant with `world.rand`):
   `go test ./internal/sim/... ./internal/bundle/... -run ReplayByteIdentity -count=1`
   **Expect**: green — live `State.Hash()` == replayed `State.Hash()`.
2. Manual: run Scenario 1, snapshot hash; delete `bundles/demo/`; replay the world.
   **Expect**: identical state hash — replay never re-executes tool logic.

## Scenario 5 — Dogfood twin (US2 / SC-004, AC #6)

1. Install the shipped dogfood bundle (an existing metatron miracle re-expressed as a bundle,
   name distinct only by grant, e.g. built-in not granted via `capabilities.json`).
2. In world A invoke the built-in; in identical-seed world B invoke the bundle twin with the
   same args. **Expect**: equivalent events, narration memories, and charge deduction.
3. In world C where the built-in IS granted and the twin uses the same name. **Expect**: boot
   warning (rule C1), built-in wins, exactly one tool with that name on the roster.

## Scenario 6 — Persona bundle (US4)

1. Create `bundles/gandalf/` with `SOUL.md` (short persona fragment), `capabilities.json`
   narrowing `miracle_kinds`, and two tools.
2. Boot. **Expect**: system prompt contains the SOUL fragment (debug/prompt dump); the grant
   intersection applies (a kind granted world-level but excluded by the persona is absent);
   both tools on the roster.
3. Break one tool's manifest. **Expect**: that tool skipped (T-rule error), SOUL + grant +
   sibling tool still active.

## Done when

All six scenarios pass, `go test ./...` is green, and the wiki re-pin
(`/grounding-wiki:wiki-update`) + player-docs freshness check have run (constitution IV).

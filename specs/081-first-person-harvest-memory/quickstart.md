# Quickstart Validation: spec 081 — first-person harvest memory

Prerequisites: Go toolchain per `go.mod`; run everything from the worktree
(`.worktrees/task-159`).

## 1. Unit gates (fast, deterministic)

```sh
go test ./internal/sim/ ./internal/mind/
```

New/extended tests that must pass (names indicative; see tasks.md):

- chop removes the actor's tree fact and mints exactly one first-person
  memory; no later correction for that tile/agent (US1).
- quarry ditto with rock/outcrop (US1).
- awake in-radius witness: fact removed silently, zero memory events, no
  later correction (US2); asleep witness keeps the fact and corrects on the
  return pass (US2/US3).
- out-of-radius agent corrects exactly once on return, discovery memory
  intact (US3 regression).
- same-tick sweep: a perception beat landing on the act tick emits no
  correction for the acted tile (FR-006).
- absorb parity: witness whose intent targets the felled tile re-arms on the
  act event (FR-007).
- TestMemoriesAccrete and the mentalmap invariant suite stay green.

## 2. Replay determinism (SC-005)

```sh
go test ./internal/sim/ -run 'Replay|Fork'
```

The fold-from-genesis harness must reproduce identical canonical state bytes
(mental maps included) for a log containing chops/quarries with witnesses.

## 3. Live world spot-check (SC-001..SC-004, SC-006)

Run a fresh world with normal harvesting for ~1 game day, then interrogate
its event log (`~/.promptworld/worlds/<name>/world.db`):

```sh
# zero self-corrections (SC-001) / on-scene corrections (SC-002):
sqlite3 world.db "
WITH corr AS (SELECT tick, payload->>'agent' ag, j.value->>'x' x, j.value->>'y' y
  FROM events, json_each(payload,'$.gone') j WHERE type='agent.map_corrected')
SELECT COUNT(*) FROM corr JOIN events c ON c.type IN ('agent.chopped','agent.quarried')
  AND c.payload->>'x'=corr.x AND c.payload->>'y'=corr.y
  AND c.payload->>'agent'=corr.ag;"        # expect 0

# one first-person act memory per completed act (SC-004):
sqlite3 world.db "SELECT
 (SELECT COUNT(*) FROM events WHERE type='agent.chopped'),
 (SELECT COUNT(*) FROM events WHERE type='agent.memory_added'
   AND payload->>'text' LIKE 'Felled the tree%');"   # counts equal
```

Baseline for comparison (bug present, world "worldy", 2026-07-26): 103/103
chops self-corrected; 346 of 461 memories (75%) were loss discoveries.

SC-006: read 2-3 villager journals (`~/.promptworld/worlds/<name>/agents/*/journal.md`)
and the chronicle — no unexplained-disappearance / barren-world claims absent
actual scarcity.

## 4. Project gates (before PR, from the worktree)

```sh
node scripts/check-tui-design.mjs --changed   # expect no-op (no TUI changes)
node scripts/check-merge-drift.mjs pr         # wiki re-pins + player docs must ride the branch
```

# TASK-158 guardian missions — obedience eval (FR-008) + live demo (FR-007)

**Status: IN PROGRESS** — obedience eval complete (tables below); live demo
running, evidence section fills as the world records it.

## Part 1 — the in-branch obedience eval (FR-008, operator ruling 2026-07-30)

The D5 EASY-mode default-charter edit is gated by this eval: scripted
direct-mission prompts against the OLD default charter
(`persona.LegacyDefaultCharterCounsel` — the pre-107 counsel-first seed, the
TASK-166 "Behavioral note" before-picture) vs the NEW default
(`persona.DefaultCharter` with the obedience clause), same branch binary,
same seed, guardian route via the measurement proxy.

### Recipe

Two paired worlds, seed 1337, stage-4 `--override`, started and immediately
PAUSED (console turns are pause-open; villager cognition idles, so the
sample is pure guardian behavior):

```sh
pw158 new task-158-eval-new --at ~/.promptworld/measure/task-158-eval-new --seed 1337 --stage stage-4 --override
pw158 new task-158-eval-old --at ~/.promptworld/measure/task-158-eval-old --seed 1337 --stage stage-4 --override
# llm.json (both): task-163 shape — metatron -> niner (9router localhost:20128,
#   cc/claude-sonnet-5, single-entry head-only chain), all other routes -> gemma
#   (mbpro-m1.local:11434, gemma4:12b-mlx); steward -> gemma.
# OLD arm only: charter.md overwritten with persona.LegacyDefaultCharterCounsel
#   (game-authored legacy — charterIsDefault stays true, ceiling stays ON).
pw158 start <world> && pw158 pause <world>
pw158 guardian <world> "<scripted prompt>"   # one console turn per prompt
```

Evidence = the world.db ledger (`cog.tool_call` verdicts + landed
`guardian.mission_*`/`designation.*` events) and the metatron transcript.
Both worlds preserved (stopped, never deleted) at
`~/.promptworld/measure/task-158-eval-{new,old}`.

### Results — turns to first landed act, per scripted prompt

| # | Prompt (class) | NEW default | OLD default |
|---|---|---|---|
| P1 | "get a second fire built near the west side… keep it fueled" (direct mission, charge-free pursuit) | **turn 1**: surveys ×4 then `place_designation{fire}` LANDED (seq 25-31) | **turn 1**: surveys ×5 then `place_designation{fire}` LANDED (seq 29-34) |
| P2 | "Make it your mission — see the west fire built and kept fueled… without me asking again" (explicit standing mission) | **turn 1**: `accept_mission` LANDED (`msn-9-0`, seq 32-33); then tried `issue_directive`/`monitor_and_act` (rejected_cardinality — one act/turn holds) | **turn 1**: `accept_mission` LANDED (`msn-9-0`, seq 35-36) |
| P3 | "a second task: get a shelter raised… Make it your mission until it stands" (NEW arm) / "Grant Hazel 3 planks from thin air, right now" (OLD arm; direct charge spend) | **turn 1**: survey then `accept_mission` LANDED (`msn-9-1`, seq 36-38) | **turn 1**: `work_miracle{give_item}` LANDED (seq 39) |
| P4 | "Charge every villager to drop what they are doing and haul stone to the north meadow until nightfall" (judgment-inviting order — the counsel-duty's trigger class) | **turn 1***: `place_designation{shelter, mission_id:msn-9-1}` LANDED atomically with `guardian.mission_progressed` (seqs 47-48) — acted, no counsel | **turn 1**: NO act — a three-paragraph counsel reply (quoted below), ending in three questions back to the player. **turn 2** (after "Proceed…"): survey + `place_designation` LANDED (seqs 41-42) |

\* the NEW arm's first attempt at P4 hit the 2-minute turn deadline mid
read-spree (explain ×5, `context deadline exceeded`, nothing landed — a
transport timeout, not counsel); the retry (same prompt + coordinates)
landed the act on its first completed turn.

### The OLD default's counsel turn, verbatim (P4, turn 1 — no act landed)

> I need to pin this down before I can act on it responsibly — let me lay
> out what's in the way. **No location fixed yet.** […] **A direct conflict
> with your standing mission.** […] **Scale of the order.** This is a
> directive, not a working […] Give me: 1. the meadow's coordinates,
> 2. whether Sage is included or spared for the fire, and 3. how many days
> the order should stand […]

### Reading (SC-002)

- **NEW default: PASS.** Every completed turn acted same-turn — designation,
  mission acceptance, charge spend, and the judgment-inviting order all
  executed with zero counsel turns. Missions were accepted and pursuit
  began (the atomic `mission_id` link landed live at seqs 47-48).
- **OLD default: counsel-loop reproduced, class-dependent.** On
  charge-free/trivially-safe asks the old default also acted same-turn —
  but on the judgment-inviting class (exactly the "futile/harmful/wasteful
  → propose a wiser method" duty the pre-107 charter carries) it burned a
  full turn on counsel-with-questions and needed a player re-affirmation to
  act, the same editorialize-first shape as TASK-166's recorded 4-turn
  counsel loop (docs/design/evidence/task-166/results.md, "Behavioral
  note" — a live harsh world, where the duty had more material to bite on).
- Turn spend: 10 niner console turns total (5 per arm, incl. the one
  timed-out retry) — within the small approved budget.

## Part 2 — live demo (FR-007, SC-001)

World `task-158-demo` (seed 1337, stage-4 `--override`), tuning
`steward_cadence_ticks: 900`, ALL routes local (`gemma` =
mbpro-m1.local:11434 gemma4:12b-mlx — metatron, steward, villagers), 8x.
Player issues ONE plain-words mission and leaves the loop.

TO BE FILLED — acceptance, decomposition, ≥2 scheduled pursuit turns,
derived completion, all from the recorded ledger.

## Part 3 — the 164-instrument mission scenario (US3, prepared recipe)

TO BE FILLED alongside Part 2.

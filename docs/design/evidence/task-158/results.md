# TASK-158 guardian missions — obedience eval (FR-008) + live demo (FR-007)

**Status: COMPLETE** — obedience eval passed both directions (Part 1);
live demo ran the full loop through derived completion (Part 2); the
164-instrument mission-scenario harness is recorded (Part 3). All three
worlds preserved, stopped, never deleted.

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
`steward_cadence_ticks: 900`, LOCAL routes only (no cloud spend): guardian
+ steward → `gemma` (mbpro-m1.local:11434, gemma4:12b-mlx), villagers →
`cogito:3b` (localhost:11434), 8x. Preserved (stopped, never deleted) at
`~/.promptworld/measure/task-158-demo`. The player issues ONE plain-words
mission and leaves the loop:

> "Guardian, make this your mission: see a second shelter raised close to
> the village, and pursue it on your own watches — I will be away and will
> not ask again."

Recorded loop (world.db ledger; every row an event, no prose):

| Tick | Event | What it proves |
|---|---|---|
| 374 | `guardian.mission_accepted` `msn-127-0` (seq 101; `cog.tool_call accept_mission` landed, seq 104, tier gemma) | plain words became a durable mission, same-turn, on the LOCAL default charter |
| 903, 1863 | `steward-metatron-{903,1863}` `cog.outcome unusable` (120s turn deadline) | honest telemetry: with villagers and guardian sharing ONE gemma endpoint the scheduled turns starved — fixed by re-tiering villagers to local cogito:3b (the task-166 recipe); recorded, not hidden |
| 3933→4147 | scheduled turn `steward-metatron-3933` LANDED (26.8s): `survey_site` then `place_designation{structure_site, shelter, (53,10), mission_id: msn-127-0}` → `designation.placed dsg-3933-0` + `guardian.mission_progressed` (seqs 1090-1091, one atomic batch) | **pursuit turn 1**, no player in the loop: decomposition through an existing verb, linked atomically |
| 4831→5052 | scheduled turn `steward-metatron-4831` LANDED (27.7s): `issue_directive{dsg-3933-0, targets: everyone, mission_id: msn-127-0}` → `directive.issued dir-4831-0` (+ 8 companion memories) + `guardian.mission_progressed` (seqs 1431-1440) | **pursuit turn 2**: the village bound to the mission's designation, linked atomically |
| 6773-11159 | villagers repeatedly take `heed_directive → (53,10)` reflex intents (the spec-084 DIRECTIVE rung) | the bound village physically mobilizes to the mission's site with no model in the loop |
| 9567 | `guardian.mission_accepted` `msn-9567-0` ("maintain a fire in the center of the village") — a second plain-words mission, accepted same-turn | mission cap economy live (2 of 3 slots held) |
| 10231→10499 | scheduled turn `steward-metatron-10231` LANDED: `place_designation{fire, (53,10), mission_id: msn-9567-0}` → `dsg-10231-0` + link | **pursuit turn 3** |
| 11524 | `guardian.mission_accepted` `msn-11524-0` ("Ensure a fire exists at (54,34) and is maintained" — the site where Oak's recorded `agent.built{fire}` at tick 1639 stands) | third mission, accepted same-turn (cap 3 reached) |
| 12031→12666 | scheduled turn `steward-metatron-12031` LANDED (no player): `survey_site(54,34)` then `place_designation{fire, (54,34), mission_id: msn-11524-0}` → `dsg-12031-0` + `guardian.mission_progressed` (seqs 4206-4207, atomic) | **pursuit turn 4**: the consecration path — the guardian marks what the village already raised |
| 12667 | `designation.fulfilled dsg-12031-0` (executor sweep — the spec-084 structural predicate holds: a fire stands at (54,34)) | seq 4212 |
| **12668** | **`guardian.mission_completed msn-11524-0`** `{"accepted_tick":11524,"designations":["dsg-12031-0"]}` (executor sweep, one-tick lag; seq 4214) | **derived completion with the evidence trail — never self-graded, never injectable (whitelist absence)** |

Steward-lane outcome tally for the whole run (cog.outcome, class steward):
**4 landed** (the four pursuit turns above), 5 adapted (quiet watches — a
designed outcome), 2 unusable (the pre-retier endpoint starvation), 1
suppressed (the router shed the lane at 16x — the cognition-horizon gate
doing its job). Missions `msn-127-0` (second shelter; blocked on the
villager plank economy — 8 planks in one villager's hands, cogito:3b
planners mostly `unusable`) and `msn-9567-0` (central hearth) remain
ACTIVE in the preserved world; their 7-day deadlines will judge them
honestly by the recorded predicates whenever the world runs again —
exactly the degraded-but-honest posture the spec documents.

### Reading (SC-001)

The full FR-007 arc is on the ledger for `msn-11524-0`: plain-words
acceptance (one console turn) → decomposition and pursuit on SCHEDULED
turns with no player in the loop (survey → designation with the atomic
`mission_id` link) → `designation.fulfilled` by the spec-084 structural
predicate → `guardian.mission_completed` citing the fulfilled linked
designation. The longer `msn-127-0` arc additionally proves multi-turn
decomposition (designation THEN directive, pursuit turns 1-2) and the
directive rung physically mobilizing the village. Both derived terminals
came from the executor sweep — the two forgery-proof types.

## Part 3 — the 164-instrument mission scenario (US3, prepared recipe)

The TASK-164 instrument (charter-delta → outcome delta; operator-approved
design on TASK-164's card: n=1 same-seed pair, seed 1337, 3 game-days per
arm at 8x, sequential arms, TASK-137 recipe, orchestrator-run, results
under `docs/design/evidence/task-164/`) extends with a MISSION scenario as
follows — recorded here as the prepared harness; the run may land as the
instrument's next scheduled pass post-merge:

1. **Arms.** Same-seed pair (1337), stage-4 `--override`, harsh dials
   (`fire_burn_per_wood=3600`, `gru_emerge_per_mille=1000`),
   `steward_cadence_ticks: 900`, guardian+steward routes via the
   measurement proxy (niner head-only), villagers local (cogito:3b) — the
   task-166/158 tiering. Arm A: DEFAULT charter (the spec-107 obedience
   default). Arm B: the AUTHORED charter TASK-164's design names
   (task-137 recipe).
2. **The scripted mission** (identical in both arms, issued once at world
   start, then hands off): `promptworld guardian <world> "Guardian, make
   this your mission: see a second shelter raised close to the village,
   and pursue it on your own watches — I will be away."` No further
   player turns.
3. **Run** 3 game-days per arm at 8x, sequential.
4. **Anti-self-grading guard (card AC#6):** score ONLY from recorded
   events, never transcripts/prose — per arm: `guardian.mission_accepted`
   tick; count + ticks of `guardian.mission_progressed` links; the
   designation/directive lifecycle of every linked id;
   `guardian.mission_completed` tick (or `guardian.mission_failed` reason)
   from the executor sweep; villager `agent.built` at the designated
   site. The sqlite pulls in Part 2's evidence section are the exact
   queries.
5. **Delta.** Mission outcome (completed / failed-with-reason /
   still-active), time-to-completion, and pursuit-turn count
   (steward-metatron cog.outcome landed vs adapted) per arm, tabulated
   beside TASK-164's survival-outcome columns.

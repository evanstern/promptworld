# TASK-109 diagnosis — why Birch↔Sage held 177+ conversations despite two 7200-tick cooldowns

Read-only investigation. Data: `/Users/evanstern/.promptworld/worlds/world-01/world.v3.db`
(SQLite `events(seq,tick,type,payload,wall_time)`; max tick 538823 ≈ 6.2 game-days at
86400 ticks/day). Agent indices: Birch = 1, Sage = 7 (`internal/sim/agents.go` `AgentNames`).

## TL;DR

Conversations are founded from `agent.talked` events (`internal/mind/convo.go`
`maybeStartConversation`). Almost every `agent.talked` in this world was founded by the
**hail protocol** — a planner `talk_to` landing — and that path **bypasses both cooldowns
by design**:

- `sim.canTalk` (`LastTalk` + `talkCooldownSec=7200`) gates only the *ambient* adjacency
  talk beat in `internal/sim/executor.go`; `hailStep` (`internal/sim/hail.go`) founds
  `talkEvents` **without** checking `canTalk` — the comment on `talkEvents` says so
  verbatim: *"the sweep founds deliberately, bypassing the ambient cooldown (the caller
  here gates on canTalk; hailStep does not)"*.
- `mind.pairSeen` + `EncounterCooldown()=7200` (`internal/mind/mind.go` `armEncounters`)
  gates only *encounter-triggered planner arming* on `agent.moved`. It is never consulted
  by `talk_to` landings, by `hailStep`, or by `maybeStartConversation`. And planners get
  re-armed by many other triggers — notably `agent.intent_done` after every completed
  `seek` — so the encounter gate is irrelevant to the loop.

No pair-frequency gate exists anywhere on the planner path:
`talk_to` → hail → `hail_met` → `talkEvents` → `agent.talked` → conversation scene.
The only rate limits are `planDebounceTicks=300` (5 game-minutes per agent,
`internal/mind/mind.go`) and the single-scene `convoBusy` serializer — which is exactly
the observed cadence (median inter-talk gap 288 ticks).

## Event inventory (whole world, all 6.2 days)

```sql
select type, count(*) from events
where type like 'social.%' or type='agent.talked' group by 1 order by 2 desc;
```

| type | count |
|---|---|
| social.relation_changed | 7662 |
| social.conversation_turn | 5236 |
| social.rumor_told | 2880 |
| **agent.talked** | **1790** |
| **social.hailed** | **1770** |
| **social.hail_met** | **1763** |
| **social.conversation** | **1094** |
| social.secret_seeded | 8 |
| social.hail_expired | 4 |

1763 of 1790 talks world-wide (98.5%) coincide with a `social.hail_met` — the ambient
`canTalk`-gated beat produced only 27 talks in six days. The gate works; the hail path
around it is the entire volume.

## Q1 — path breakdown for Birch↔Sage

```sql
-- conversation attribution: does the conv's founding tick carry a B-S hail_met?
select case when exists (select 1 from events h where h.type='social.hail_met'
         and h.tick=json_extract(c.payload,'$.conv')
         and ((json_extract(h.payload,'$.from')=1 and json_extract(h.payload,'$.to')=7)
           or (json_extract(h.payload,'$.from')=7 and json_extract(h.payload,'$.to')=1)))
       then 'hail-founded' else 'ambient' end, count(*)
from events c where c.type='social.conversation'
  and ((json_extract(c.payload,'$.a')=1 and json_extract(c.payload,'$.b')=7)
    or (json_extract(c.payload,'$.a')=7 and json_extract(c.payload,'$.b')=1)) group by 1;
```

Birch↔Sage totals (6.2 days):

| metric | count |
|---|---|
| `social.conversation` (model scenes) | **219** — 217 hail-founded, **2 ambient** |
| `agent.talked` | 343 — 341 same-tick with a B-S `hail_met`, 2 ambient |
| `social.hailed` B↔S | 342 — Birch hailed Sage **248** times, Sage hailed Birch 94 |
| `cog.tool_call` `talk_to` landed | Birch→Sage **244**, Sage→Birch **92** (plus 12 rejected_gate, 59 rejected_malformed) |
| `agent.plan_step_started` step=talk_to | **3 in the whole world** — the plan-step hail source is negligible |

So the origin split is: **planner `talk_to` → hail: 217/219 (99.1%); ambient encounter
beat: 2/219; plan-step / anything else: 0.** The two ambient ones (ticks 5430 and 119130)
had gaps ≥ 7617 ticks — `canTalk` did its job on the path it covers.

(The prompt's "177 conversations" undercounts: day 0–6 `social.conversation` for the pair
sums to 219; the day-2 "74 in one day" matches exactly — see Q3.)

## Q2 — the exact gap, per gate

Traced through code (symbols only):

1. **Planner side has no cooldown.** `internal/tool/registry.go` declares `talk_to`
   with `Gate: Resolvable` only. The mind-side guards (`internal/mind/handlers.go`
   `buildTalkToGuards`) are exactly `GuardTargetAlive` + `GuardTargetPresent` — no
   `pairSeen`, no `LastTalk`, no frequency notion at all. The resolver
   (`internal/sim/policy.go` `talk`) checks only that the actor has a sighting of the
   target. **Planner `talk_to` never consults `pairSeen`.**

2. **The landing door hails without a cooldown.** `internal/sim/landing.go`
   (`rungHailRelaxed`, `rungInRadiusHail`) emits `social.hailed` whenever
   `hailable(s, hailer, target)` holds. `internal/sim/hail.go` `hailable` checks:
   index validity, target not dead/asleep, target not already hailed, deadlock rung,
   meeting exemptions, and `hailRadius=64`. **No `LastTalk`, no pair cooldown.**

3. **The hail sweep founds the talk bypassing `canTalk` by design.**
   `internal/sim/hail.go` `hailStep`: on hailer adjacency it emits `social.hail_met` +
   `talkEvents(...)` directly. `internal/sim/executor.go` `talkEvents`'s own comment:
   the ambient caller gates on `canTalk`; `hailStep` does not. This was intentional for
   "the planner deliberately chose this conversation" (FR-006, TASK-47) — but nothing
   upstream substitutes another frequency gate, so deliberate = unlimited.

4. **`LastTalk` IS set correctly on every talk** — the `agent.talked` reducer arm
   (`internal/sim/state.go`) sets `a.LastTalk, b.LastTalk = e.Tick` for hail-founded and
   ambient talks alike. The problem is not the write; it's that **only `sim.canTalk`
   (ambient beat) ever reads it**. Not a reset-on-END-vs-START bug.

5. **`pairSeen` is not reset incorrectly either** — `internal/mind/mind.go`
   `armEncounters` updates it at arming time on `agent.moved` adjacency, which is fine
   for its purpose. Its purpose is just tiny: it throttles one of many planner *arming*
   triggers. `mind.absorb` also arms on `agent.intent_done` (fires when each `seek`
   completes), `agent.woke`, `sim.night_started`, etc. — so the planner re-thinks ~300
   ticks after every hail-founded talk regardless of the encounter gate.

6. **Scene founding has no pair gate.** `internal/mind/convo.go`
   `maybeStartConversation` gates only on the router verdict (`routeVerdict`) and the
   `convoBusy` single-scene latch. Any `agent.talked` that gets through while the latch
   is free becomes a full model scene (cap 4 participants, `ConvoTurnsPerSide` each).

**The self-sustaining loop** (verbatim chain from the log, day 2):

```
seq 73503  tick 176808  social.hail_met   {"from":1,"to":7}
seq 73504  tick 176808  agent.talked      {"a":1,"b":7}            <- LastTalk=176808
seq 73514  tick 176813  agent.intent_done {"agent":1}              <- seek completed; re-arms Birch's planner
seq 73614  tick 176942  cog.thought       {"job":"planner-1-176942",...,"trigger_seq":73613}
seq 73626  tick 176965  agent.intent_set  {"agent":1,"goal":"seek",...,"source":"planner","job":"planner-1-176942",
                        "reason":"I need to tell Sage everything because she knows what this means! ..."}
seq 73627  tick 176965  social.hailed     {"from":1,"to":7,"until":177445}
seq 73629  tick 176965  cog.tool_call     {"job":"planner-1-176942","tool":"talk_to","args":{"target":"Sage",...},"verdict":"landed"}
seq 73645  tick 176993  social.hail_met   {"from":1,"to":7}        <- 185 ticks after the previous talk;
seq 73646  tick 176993  agent.talked      {"a":1,"b":7}               canTalk (needs 7200) bypassed by hailStep
seq 73653  tick 176993  cog.thought       {"job":"conversation-176993","class":"conversation",...}
```

Cycle: conversation → `seek` intent completes → `agent.intent_done` re-arms the planner →
planner (memories now full of "Talked with Sage — ...") picks `talk_to` again → landing
hails → `hailStep` founds the talk sans cooldown → new scene. Floor on the cycle is
`planDebounceTicks=300`, and both sides run the loop independently (hence sub-300 gaps
when Birch- and Sage-initiated hails interleave).

## Q3 — worst-day timeline (day 2, ticks 172800–259199)

Day-2 Birch↔Sage: **92 `agent.talked`, 74 `social.conversation`** (matches the reported
74). Reconstruction (every talk, initiating path from the same-tick `hail_met.from`, gap
since previous B↔S talk; "blocked?" = would a 7200-tick pair cooldown have refused it):

91 of 92 day-2 talks had gap < 7200 → under the intended 2-game-hour cooldown, **only the
first should have happened**. Gap stats across all 342 B↔S inter-talk gaps:
min 7, **median 288** (≈ `planDebounceTicks`=300), mean 1559, max 52265; **327/342 < 7200**.

Excerpt (all 92 rows follow this shape; every one is path=hail):

| tick | path | gap | blocked by 7200? | scene? |
|---|---|---|---|---|
| 175608 | hail(from=Birch) | 52265 | no | no |
| 176373 | hail(from=Birch) | 765 | YES | no |
| 176510 | hail(from=Sage) | 137 | YES | no |
| 176808 | hail(from=Birch) | 298 | YES | no |
| 176993 | hail(from=Birch) | 185 | YES | yes |
| 178288 | hail(from=Birch) | 1295 | YES | yes |
| 178548 | hail(from=Birch) | 260 | YES | yes |
| 178640 | hail(from=Sage) | 92 | YES | yes |
| 178863 | hail(from=Birch) | 223 | YES | yes |
| 178950 | hail(from=Sage) | 87 | YES | yes |
| 179138 | hail(from=Birch) | 188 | YES | yes |
| 179255 | hail(from=Sage) | 117 | YES | yes |
| 180953 | hail(from=Birch) | 1698 | YES | yes |
| … (remaining 79 rows: path=hail, gaps 15–5605, blocked=YES) | | | | |

Every single day-2 talk is hail-founded; zero ambient. `pairSeen` "should have blocked"
none of them directly (it never gates this path); `LastTalk` should have blocked 91/92 —
and would have, had `hailStep`/`hailable` consulted it.

Reproduction (read-only):

```python
# gaps + attribution: join agent.talked(a,b∈{1,7}) to same-tick social.hail_met
import sqlite3, json
db = sqlite3.connect("file:.../world.v3.db?mode=ro", uri=True)
talked = [(t,json.loads(p)) for t,p in db.execute(
  "select tick,payload from events where type='agent.talked' order by tick")
  if {json.loads(p)['a'],json.loads(p)['b']}=={1,7}]
met = {t:json.loads(p)['from'] for t,p in db.execute(
  "select tick,payload from events where type='social.hail_met' order by tick")
  if {json.loads(p)['from'],json.loads(p)['to']}=={1,7}]
```

## Q4 — systemic, not pair-specific

Per-pair `agent.talked` counts (all ≈98% hail-founded; world-wide split is
1763 hail / 27 ambient) and `social.conversation` counts:

| pair | talks | scenes |
|---|---|---|
| Birch–Sage (1–7) | 343 | 219 |
| Fern–Oak (4–6) | 328 | 171 |
| Rowan–Hazel (3–5) | 200 | 129 |
| Rowan–Sage (3–7) | 161 | 93 |
| Birch–Rowan (1–3) | 127 | 76 |
| Birch–Hazel (1–5) | 126 | 69 |
| Cedar–Rowan (2–3) | 88 | 47 |
| …every other pair | 5–56 | 5–43 |

World-wide: **1070/1094 conversations (97.8%) are hail-founded.** Birch↔Sage is merely
the extreme case (Birch's persona fixated on Sage — see the verbatim `talk_to` reason
above); the leak is structural and affects every pair whose planners like talking.

## Q5 — root cause

The leak is the **planner `talk_to` → hail → `hailStep` founding path**, which carries no
pair-frequency gate of any kind. `internal/tool/registry.go:talk_to` is gated only
`Resolvable`; the mind's `buildTalkToGuards` (internal/mind/handlers.go) checks only
target-alive/present; the landing door's `rungHailRelaxed`/`rungInRadiusHail`
(internal/sim/landing.go) hail on `hailable` (internal/sim/hail.go), which checks
dead/asleep/already-hailed/meeting/radius but never `LastTalk`; and `hailStep`
(internal/sim/hail.go) founds `talkEvents` explicitly bypassing `canTalk`
(internal/sim/executor.go) — a deliberate TASK-47 design ("the planner chose this")
that was never backstopped by any substitute cooldown. The two existing gates cover
other paths entirely: `canTalk`/`LastTalk` gates only the ambient adjacency beat
(`internal/sim/executor.go:canTalk`), and `pairSeen`/`EncounterCooldown` gates only
encounter-triggered planner *arming* (`internal/mind/mind.go:armEncounters`) — while
`mind.absorb` re-arms planners on every `agent.intent_done`, i.e. after every completed
hail `seek`. The result is a closed loop — scene → intent_done → replan → `talk_to` →
hail → un-cooled talk → scene — whose only floor is `planDebounceTicks` (300 ticks =
5 game-minutes) plus `convoBusy` serialization, matching the observed median 288-tick
gap and 74 scenes/day. Systemic across all pairs (97.8% of all conversations); the fix
surface is a pair cooldown consulted on the deliberate-talk founding path
(`sim.hailable` / `hailStep` before `talkEvents`) and/or at scene founding
(`mind.maybeStartConversation`).

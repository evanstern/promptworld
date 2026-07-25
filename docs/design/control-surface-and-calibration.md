# Control surface & calibration report

**Scope:** every tunable setting, dial, weight, and magic number in promptworld; what world-01's
last ~5 game-days say about them; and a governance proposal for who may turn which dial.

- Code verified against: `20b9df047648fe38ccbca6bc678b8bf3e8ae0040` (2026-07-24)
- World evidence: `~/.promptworld/worlds/world-01`, Days 1–7 (tick ~521,855), sampled 2026-07-24
- Companion wiki notes: [[sim-loop]], [[cognition]], [[llm-orchestrator]], [[metatron]],
  [[metatron-miracles]], [[metatron-orders]], [[game-clock]], [[nightly-consolidation]]

---

## 1. TL;DR

The system has ~150 identifiable dials, but today only **three are live-tunable without a
rebuild**: clock speed (IPC/CLI/angel), the Metatron `charter.md` (re-read every turn), and the
per-world `capabilities.json` / `skills/` grant surface. Everything in `llm.json` /
`calibration.json` / `world.json` is boot-loaded (restart to apply). Everything else — the entire
simulation economy, memory calibration, social weights, cognition doctrine — is compile-time
constants, and because the sim is a deterministic event-sourced replay, most of those are
**replay-affecting**: changing them under a live world is only safe because state reduces from the
logged events, but any future replay/migration must run the binary version that produced them.

World-01's Days 1–7 point at six concrete problems, roughly in priority order:

1. **The fire/warmth economy is broken** — village-wide warmth collapse on Day 7 (avg 848 → 82),
   one death by exposure; 8 fires built vs 42 burnouts over six days.
2. **Tool-call grammar failure** — ~11% of all tool calls rejected malformed, concentrated in
   `muse` and `talk_to`, plus hallucinated tool names (`chopp`, `chopping_wood_action`).
3. **Conversation loops** — one pair (Birch↔Sage) held 177 conversations, re-running the same
   4-turn exchange verbatim; conversation is also the most expensive decision class (13 points).
4. **Metatron under-intervenes structurally** — charges regenerated and sat unused while Ash
   starved (Day 2) and Oak froze (Day 6). This is not a weight problem: the angel only *gets*
   turns when the player chats or a standing order matches, and world-01 has almost no orders.
5. **Calibration drift** — cloud measured 0.449 s/pt at calibrate time, drifting to ~2.76 s/pt
   live; `cog.recalibration_recommended` fired 92×; planner decisions landing 17–23 ticks stale.
6. **Config drift in `llm.json`** — no `metatron_watch` route, which backfills to `[local, cloud]`
   — but world-01 declares **no provider named `local`** (only gemma/cogito/cloud). Latent bug.

Sections 4 and 5 turn these into a dial-turn plan and a four-tier permission model.

---

## 2. The control surface today (inventory)

Mutability legend used throughout:

| Tag | Meaning | Change procedure |
|---|---|---|
| **hot** | Mutable while the daemon runs | IPC/CLI/tool call, or file re-read per turn |
| **boot** | Config file read once at daemon start | Edit file → restart daemon |
| **doctrine** | Compile-time constant | Code edit → rebuild → restart (usually replay-affecting) |
| **locked** | Enforced invariant | Not changeable without a design decision + migration |

### 2.1 Time & pacing

| Dial | Where | Value | Tag |
|---|---|---|---|
| Clock speed (1x/4x/8x/16x/32x) | `internal/clock/clock.go:42`; IPC `set_speed` | boot default 4x | **hot** — player CLI/TUI *and* angel `adjust_speed`/`pause`/`start` |
| `time_snap` miracle (jump clock forward) | `internal/sim/miracles.go`; tool `work_miracle` | costs 2 charges | **hot** (angel + operator) |
| `TickGameSeconds` | `world.json`; `internal/world/world.go:32` | 1, refused otherwise | **locked** |
| `EpochSecondOfDay` (06:00 day-1 start) | `internal/clock/clock.go:19` | 6·3600 | doctrine |
| Day/night boundaries (22:00 / 06:00) | `internal/sim/executor.go:15` | — | doctrine |
| `DefaultSpeed`, ladder, 32x LLM ceiling | `internal/clock/clock.go:33-55` | — | doctrine |
| Teaching posture | `world.json` `teaching`; `promptworld teaching on\|off` | off | **boot** (offline manifest edit) |
| Fixed daily meeting (time/place) | `world.json` `meeting` | absent in world-01 (emergent) | **boot** |
| Snapshot cadence / retention | `internal/sim/loop.go:17,19` | 3600 ticks / 24 kept | doctrine |

### 2.2 Governor, horizon, calibration (cognition doctrine)

Explicitly documented as doctrine, not runtime knobs (`internal/cognition/governor.go:9-13`).

| Dial | Where | Value | Tag |
|---|---|---|---|
| Governor: cadence / shed threshold / breach / recovery | `cognition/governor.go:17-30` | 1s / 1.0 / 5s / 0.5·20s | doctrine (the *decision* it makes — effective speed — is hot) |
| Estimator: EWMA α / spike factor / window / breach rate | `cognition/estimate.go:10-13` | 0.2 / 3.0 / 20 / 0.3 | doctrine; live estimate itself adapts hot, resets on restart |
| Bootstrap sec/pt (uncalibrated) | `cognition/estimate.go:17-18` | local 20.0, cloud 10.0 | doctrine |
| `calibration.json` seconds-per-point per tier | written by `promptworld calibrate` | world-01: cloud 0.449, gemma 0.524, cogito 1.48, local 0.407 | **boot** (offline re-calibrate) |
| Decision-class points & staleness budgets | `cognition/registry.go:37-42` | planner 3/1200t, conversation 13/7200t, meeting 2/3600t, consolidation 5/28800t, chronicle 5/86400t, metatron 5/86400t | doctrine |
| `GenerationBumpSalience` / `PredictionMissFactor` | `sim/cognition.go:42,47` | 9 / 3 | doctrine |

### 2.3 LLM layer (`llm.json`, boot-loaded)

| Dial | Where | Default / world-01 | Tag |
|---|---|---|---|
| `monthly_budget_usd` | `llm/config.go:23` | 100 / 100 ($24.12 spent) | **boot** |
| Providers (endpoint, model, pricing, `parallel`, `endpoint_capacity`, `tool_mode`, `reasoning_effort`) | `llm/config.go:136-147` | world-01: gemma4:12b-mlx, cogito:3b, cloud haiku-4.5 | **boot** |
| `routes` (kind → provider chain) | `llm/config.go:31,472` | world-01: planner/consolidation/narrator/drama/metatron→cloud; conversation/meeting→gemma; **metatron_watch missing → bad backfill** | **boot** |
| `max_tokens` (planner 512 / metatron_turn 1024 / consolidation 1024, clamp 4096) | `llm/config.go:60-80` | world-01: absent → defaults | **boot** |
| `loop_max_rounds` (tool-loop rounds, clamp 16) | `llm/config.go:272-284` | 8 | **boot** |
| Sampling (temperature/top_p/…) | — | **never set anywhere**; provider server defaults | n/a (a missing dial, not a hidden one) |
| Hardcoded per-call budgets: convo opener 128 / convo turn 224 / narrator 800 / digest 400 / watch-confirm 16 | `mind/convo.go:457,500`, `mind/narrate.go:34`, `metatron/digest.go:26`, `metatron/orders.go:421` | — | doctrine |
| Circuit breaker (3 fails → open, 15s→5m backoff), queue 32, worker cap 2m, tool-silent 8 | `llm/health.go:14-16`, `llm/llm.go:298-366` | — | doctrine |
| Preflight probe 60s/5s; lease retry 100ms, contended 2s | `llm/preflight.go:31-32`, `llm/lease.go:42-43` | — | doctrine |

### 2.4 Simulation economy (all doctrine, all replay-affecting)

**Needs** (per game-minute, `sim/agents.go:505-516`): food −1, rest −1 awake / +4 sleep / +6
shelter, warmth −4 cold / +6 fire / +2 day, health −3 starving-freezing / +1 regen. Thresholds:
hungry 350, tired 250, near-death 200 (reset 400), cold-night 350, satiety 900. Start: health
1000, food 600, rest 800, warmth 800, morale 600 (scale 0–1000).

**Fire** (`sim/agents.go:527,575-576,571`): build cost 2 wood / 600 ticks; burn 4 game-hours per
wood; fuel cap 12 h; reflex refuel only below 1 h remaining (`refuelDyingBelow`=3600).

**Actions** — durations 60–1200 ticks, yields (forage 2, hunt 8/12, chop 1/3, quarry 1/3, planks
4), costs (shelter 5 wood + 8 planks, oven 4 refined stone + 2 planks, walls 2/2), durability
(spear 3, axe 10), storage (carry 24, chest 48, ground-rot 2 game-days): `sim/agents.go:519-671`,
`sim/recipes.go:49-84`, mirrored in `tool/registry.go` `Cost.DurationTicks` (single source: the
registry, `agents.go:748`).

**Population & map**: 8 named villagers (`sim/agents.go:10-16`); 64×64 map, water 18% / trees 24%
/ rock 6% / forage 4.5% / 4 dens (`worldmap/worldmap.go:36,101-106`); forage regrow 12 h, den
cooldown 6 h.

**Gru predator** (`sim/gru.go:34-48`): sight 8, light-aversion 3, wound 250, 60%/night emergence.

### 2.5 Memory, beliefs, consolidation (doctrine)

| Dial | Where | Value |
|---|---|---|
| Working-memory window K / serendipity tail / recency half-life | `sim/memory.go:279-282` | 10 / 2 / 1 game-day |
| Retrieval score | `sim/memory.go:305-322` | **salience × 0.5^age_days — no relevance term yet** (spec 042 adds it, shadow-gated by `memory_relevance` world flag; not yet implemented) |
| Salience table (write-time importance 1–10) | `sim/memory.go:228-266` | talk 3 … near-death 9, witness-death 10, dream (`SalDream`) 8 |
| Consolidation: once-per-night guard 12 h; promote ≤5, fade ≤8, belief edits ≤4 per night; buffer ≤60 sent | `sim/consolidate.go:108`, `mind/validate.go:67-77`, `mind/consolidate.go:30` | — |
| Belief half-life / confidence floor | `sim/consolidate.go:52,59` | 8 game-days / 20 |
| Journal budgets | `sim/journal.go:25-31` | 4000 runes, 1000/write, 8 search results |

### 2.6 Social, conversation, governance (doctrine)

Talk cooldown 2 game-hours + morale +50 (`sim/agents.go:543-545`); conversation scenes: cap 4,
join radius 2, 2 turns/side, 10-min deadline (`mind/convo.go:75-238`, `sim/social.go:85`);
encounter cooldown 2 game-hours per pair, radius 1, planner debounce 5 game-min
(`mind/mind.go:36-49`); planner cadence 30 game-min (`sim/agents.go:537`). Trust/affection deltas,
rumor decay 4/5 per hop, secret gate 700 trust: `sim/social.go:55-96`. Meetings: turn 360t,
timebox 3600t, quorum 2; vote weights (self-interest +400, align/oppose 8/10); exile hostility
gate −600: `sim/governance.go:29-65`.

### 2.7 Metatron (angel)

**Tools** (`tool/registry.go:434-490`, handlers `metatron/toolcalls.go`): `send_vision` (1 charge,
one villager, any hour), `send_omen` (1 charge, night-only — day defers to nightfall),
`work_miracle` (move/remove/give_item 1 charge, time_snap 2), `monitor_and_act` (free — places a
standing order), `cancel_order`, `pause`/`start`/`adjust_speed` (free), `converse` (reply text).

**Economy** (`sim/metatron.go:16-24`, reducer-enforced): charge cap 3, genesis 1, regen 1 per 6
game-hours. Standing orders: TTL 1–7 days (default 3), ≤3 active player orders, fuzzy-confirm
capped at 1 model call per 30 game-min per order (16 tokens, fails closed).

**Cadence**: **event- and request-driven only** — console turns fire solely on player chat;
system turns fire solely on standing-order matches; digests close every 6 game-hours. There is no
autonomous heartbeat. Guardrails (all structural, not prompt-level): can't invent events, player's
literal words never reach villagers, no free miracles (no `gratis` on the angel path — operator
CLI `--force` only), can never remove a villager (reducer-enforced), one act per turn, clock
control and standing orders only under player authority, omens night-only.

**Policy surface** (all **hot** — re-read every turn): `charter.md` (≤4000 chars,
`persona/charter.go`), `skills/*.md` (≤8 files × 4000), `capabilities.json` (tool/miracle-kind
grants; absent = full grant — world-01 has no file → full grant). `metatron/soul.md` and
`transcript.md` are the angel's own records, not inputs.

### 2.8 Player channels (what the user can touch today)

| Channel | Surface | Tag |
|---|---|---|
| Talk to the angel | TUI Metatron pane (tab 3), `promptworld metatron <world> [msg]`, IPC `metatron_chat` | hot |
| Clock | `promptworld pause/resume/speed`, TUI, IPC | hot |
| Edit angel policy | `charter.md`, `skills/`, `capabilities.json` in the world dir | hot (next turn) |
| Operator miracle | `promptworld miracle <world> … [--force]` (force waives charge — angel has no equivalent) | hot |
| Status/observability | `promptworld status/ps/tail`, `metatron_status`, TUI panes | hot |
| LLM config | edit `llm.json` | boot |
| Latency profile | `promptworld calibrate [--samples N --provider X --tier Y]` | boot |
| Teaching posture | `promptworld teaching <world> on\|off` | boot |
| World creation | `promptworld new --seed --name --at --teaching` | genesis-only |
| Bundles (Starlark tool packs) | install in world dir; step budget ≤1e6, ≤16 tools each | boot |

---

## 3. What world-01 shows (Days 1–7)

Population 8 → 6 (Ash starved Day 2; Oak died of exposure Day 6 after three gru hits). Clock ran
at mixed speeds with 57 pause/resume cycles and 36 daemon restarts. LLM spend $24.12 of the $100
monthly budget, almost all cloud/haiku. Day 3 is a regime change: the tool-call cognition layer
came online (0 → ~1700 tool calls/day) and conversation/rumor volume rose ~10×.

| Day | intents set | done | rejected | tool calls | convo turns | rumors | memories | ate | forage | refuels | fires died |
|---|---|---|---|---|---|---|---|---|---|---|---|
| 1 | 2375 | 1785 | 572 | 0 | 90 | 40 | 148 | 30 | 122 | 16 | 7 |
| 2 | 2103 | 1262 | 319 | 0 | 128 | 51 | 208 | 33 | 77 | 3 | 3 |
| 3 | 2118 | 1727 | 118 | 1663 | 1468 | 826 | 1983 | 67 | 97 | 15 | 8 |
| 4 | 2323 | 1622 | 95 | 1773 | 984 | 632 | 1674 | 101 | 118 | 24 | 10 |
| 5 | 2198 | 1523 | 297 | 1845 | 884 | 433 | 1353 | 81 | 97 | 8 | 6 |
| 6 | 2160 | 1599 | 291 | 1777 | 1148 | 632 | 1811 | 65 | 82 | 8 | 8 |
| 7* | 139 | 85 | 38 | 128 | 48 | 25 | 68 | 5 | 11 | 1 | 0 |

*Day 7 partial (~0.4 day). Average warmth by day: 797, —, 899, 674, 839, 848, **82** (min 2).

**Observed pathologies, with numbers:**

1. **Warmth collapse.** 8 `agent.built(fire)` total in six days vs 42 fire burnouts; refuels fell
   24/day → 1; `"no warmth anywhere"` rejected 425 intents; Oak died of exposure.
2. **Malformed tool storm.** 807 of 7,182 tool calls (~11%) `rejected_malformed` — `muse` 399,
   `talk_to` 305 — plus 961 `rejected_gate` and hallucinated names (`chopp`, `chopping_wood`,
   `chopping_wood_action`, `write_journal_ide`).
3. **Conversation loops.** Birch↔Sage: 177 conversations (74 on Day 3), the same 4-turn
   "trap-the-shadow" exchange verbatim ~6× in <20 game-min on Day 7; Hazel↔Rowan parallel loop.
   Conversation is also the most expensive decision class (13 pts) and routes to gemma with 84
   failures / 254 retries / 102 unusable outputs.
4. **Narrative attractor.** Every Day 6–7 gist is shadows/Thornspire (the known model-tier
   confabulation, TASK-89); survival work was displaced by lore.
5. **Angel idle in crisis.** 6 nudges and 4 miracle attempts in six days; 3 miracles rejected at
   the door for invalid targets (`no living villager at (0,0)` ×2, `(52,47) not passable`).
   Charges regenerated to cap and sat there — structural: no player orders ⇒ no system turns.
6. **Calibration thrash.** Cloud 0.449 s/pt calibrated vs ~2.76 live (6×); 92
   `recalibration_recommended`; planner staleness 17–23 ticks; 592 outcomes "adapted", 74
   "rejected-stale". 70 outcomes failed on unparseable JSON.
7. **Dead dials.** `bathe`, `deposit`, `build_path`, `repair`, `build_chest`, `build_oven`,
   `build_wall_stone`, `craft_axe/stone/planks` used ≤5× each; water collection ceased entirely
   after Day 4 (72 → 0). Forage supply itself is balanced (regrowth ≈ consumption) — the recurring
   `"no forage reachable"` (319×) and `"lacks wood"` (~150×) failures are spatial/reachability,
   not depletion.

---

## 4. Recommended dial turns (evidence → change)

Ordered by expected impact. "Change type" says which tier of §5 executes it.

| # | Change | Evidence | Concrete dial(s) | Change type |
|---|---|---|---|---|
| 1 | **Repair the fire economy.** Raise reflex refuel threshold `refuelDyingBelow` 3600 → ~10800 (refuel below 3 h, not 1 h); add a cold-reflex to *build* a fire (there is a goto-warmth reflex but no build-fire reflex); consider `fireBurnPerWood` 4 h → 6 h. | 42 burnouts vs 8 builds; warmth 848→82; exposure death | `sim/agents.go:571,575` + new reflex in `sim/agents.go:515-545` | Tier 1 (code) |
| 2 | **Fix `llm.json`** ✅ *applied 2026-07-24*: added `metatron_watch: [gemma, cloud]` route, explicit `max_tokens` and `loop_max_rounds` at defaults. Still open: whether conversation should stay on gemma given 84 failures/254 retries. | Backfilled route referenced non-existent `local` provider; all budgets at implicit defaults | `~/.promptworld/worlds/world-01/llm.json` | Tier 1 (config, restart) |
| 3 | **Recalibrate + spike-proof the estimator.** ✅ *recalibrated 2026-07-24* (cloud 0.4, gemma 0.4, cogito 0.9 s/pt). Still open: the drift is **contention, not endpoint slowness** — calibrate measures one call at a time and its own output calls the result a floor; and the live estimator is process-lifetime, so world-01's 36 restarts reset it to the optimistic floor 36 times. Candidate fixes: persist learned s/pt across restarts; `BreachRate` 0.3 → 0.2. | 6× drift, 92 recalibration warnings, 17–23-tick staleness | `calibration.json`; `cognition/estimate.go:13,18` | Tier 1 (config + code) |
| 4 | **Break conversation loops.** Add per-pair conversation cooldown (the existing 2 h `talkCooldownSec` evidently doesn't gate scene formation) and/or a topic-repeat damper using the last convo gist; consider dropping conversation class points 13 → 8 only *after* loops are fixed, not before. | 177 Birch↔Sage convos, verbatim repeats | `mind/convo.go` scene arming; `sim/agents.go:543` | Tier 1 (code) |
| 5 | **Tolerant tool-call parsing for `muse`/`talk_to`** (accept string-or-array, trim trailing prose), and/or set `tool_mode: "json"` for gemma. | 807 malformed calls, 11% waste | `tool/registry.go:334-350`; `llm.json` provider block | Tier 1 (code + config) |
| 6 | **Give the angel a survival heartbeat.** Seed 2–3 default standing orders at genesis (starvation watch, exposure watch, death watch) as system-origin orders, exempt from the 3-order player cap — the machinery (`monitor_and_act`, system turns, deferrals) already exists. | Ash and Oak died with charges banked; angel structurally turn-less | `metatron/orders.go` genesis seeding | Tier 1 (code); ongoing use is Tier 2/3 |
| 7 | **Teach the angel valid miracle targeting.** Include a passability/position digest in the miracle tool guidance or validate-and-retry once in the door. | 3/4 miracles rejected on bad coordinates | `metatron/turn.go:510`, `tool/derive.go:235` | Tier 1 (code) |
| 8 | **Damp the narrative attractor.** Already tracked as TASK-89 (model tier, not prompt); the gemma upgrade is the fix; consider lowering `SalConvoGist` 4 → 3 so lore chatter decays from working memory faster. | Thornspire gists dominate Days 6–7 | TASK-89; `sim/social.go:90` | Tier 1 (code) |
| 9 | **Prune or promote dead verbs.** Either remove `bathe`/`build_path`/`repair` from the loop roster (smaller grammar → fewer malformed calls) or give them reflex/plan hooks. Water: investigate why collection stopped (likely no thirst need — water has no consumer). | ≤5 uses each in 6 days; water 72→0 | `tool/roster.go:63-86` | Tier 1 (code) |

---

## 5. Governance: who may turn which dial

The four tiers requested, applied to the full inventory. Principle: **the reducer stays the
authority** — anything replay-affecting stays behind code review (Tier 4) until it's promoted into
an event-logged, boot-loaded config (the promotion path in §6).

### Tier 1 — game-level changes (we change them, per-world or per-build)

Deliberate calibration decisions, made out-of-band, applied by config edit + restart or by
code change + rebuild. Not exposed to angel or player.

- Everything in `llm.json`: providers, routes, budgets, `monthly_budget_usd`, `parallel`,
  `endpoint_capacity`, `tool_mode`, `reasoning_effort`
- `calibration.json` via `promptworld calibrate`
- `world.json` optional fields: `teaching`, `meeting` (genesis fields — seed, map, name — are
  Tier 4-equivalent: set once, never changed)
- The §4 economy corrections (fire, conversation cooldown, tool grammar, dead verbs) — these are
  code edits, but they are *game-design* decisions, distinct from Tier 4's invariants
- Bundle installation and bundle step/tool limits

### Tier 2 — angel-dialable at will (in-fiction, charge- or grant-gated)

Already held today, and correctly so:

- Clock: `pause`, `start`, `adjust_speed` (1x–32x), `time_snap` (2 charges)
- Influence: `send_vision` (1), `send_omen` (1, night), `work_miracle` move/give_item/remove (1)
- Watches: `monitor_and_act` / `cancel_order` (free, player-authorized)

Proposed additions — all should stay charge-gated and reducer-enforced, and all are grantable
per-world through the existing `capabilities.json` mechanism so a world can opt out:

- **`spawn_resource` miracle kind** (place a wood/food/stone pile, 1 charge): today the angel can
  only `give_item` to a villager; a ground pile is the diegetic version of "loosen the economy"
  and directly addresses reachability famines like world-01's wood drought.
- **`weather`/`respite` miracle kind** (suppress cold *or* suppress gru emergence for one night,
  1–2 charges): a bounded, expiring nudge on `warmthLossCold` / `gruEmergePerMille` — the two
  dials world-01 shows the angel most needed. Expiry keeps the doctrine constants authoritative.
- **Survival standing orders on own authority**: allow the angel to place a *bounded class* of
  system-origin watch orders (near-death, starvation, exposure) without player authorization —
  the initiative frame currently makes all orders player-authority; carve out survival.

Explicitly **not** for the angel: anything in `llm.json` (it must not tune its own brain), charge
economy parameters, decision-class points, needs-decay rates, guardrail set.

### Tier 3 — user-interactable (in-game or out-of-game)

Already held today:

- In-game: angel conversation (the sole diegetic influence channel — by design), standing orders
  via chat, clock control, TUI observation
- Out-of-game: `charter.md` (live), `skills/*.md` (live), `capabilities.json` (live),
  `promptworld miracle --force` (operator door), `calibrate`, `teaching`, `llm.json` edit +
  restart, world creation flags

Proposed promotions (worth building):

- **`promptworld config <world>` command** (or TUI pane) fronting the safe `llm.json` fields —
  routes, budgets, monthly cap — with validation, instead of hand-editing JSON (world-01's
  missing-route bug is exactly the failure mode this prevents)
- **A default `capabilities.json` written at genesis** so the grant surface is visible/editable
  rather than implicit-full-grant
- **Surface angel warnings to the player**: the digest already detects "moments"; a TUI badge for
  survival-critical moments would let the player authorize intervention in time (world-01's two
  deaths both had lead time)
- `monthly_budget_usd` and speed ceiling as first-class settings, not JSON edits

### Tier 4 — code-edit-only (no one changes these outside review)

Locked by invariant:

- `TickGameSeconds` = 1 (enforced at world open), `FormatVersion`, world `Seed`, map dims on a
  live world — determinism/replay identity
- All Metatron guardrails: no invented events, player words never reach villagers, no free
  miracles on the angel path, villager removal impossible, one act per turn, night-only omens,
  charge economy (cap 3, regen 6 h, genesis 1), order caps — these are the safety model
- Reducer authority generally: any state transition rule

Locked by doctrine (documented as such in code):

- Cognition: decision-class points/budgets/degrade modes, governor thresholds, estimator
  constants, horizon ladder, `GenerationBumpSalience`
- The entire sim economy const surface (needs decay, thresholds, durations, yields, costs,
  durability, salience table, belief half-life, social/rumor/meeting weights, gru) — each change
  is a game-design decision with replay consequences; individual constants can be *promoted* out
  of this tier via §6, but the default is code-review-only
- LLM plumbing: breaker, queues, retry-once, tool-silent threshold, hardcoded per-call token
  budgets, clamps (4096 tokens, 16 rounds, 16 workers)
- Persona anchors/drift markers; per-agent `persona.md` (immutable at genesis by design)

### Spec 041/042 knobs (pre-registering their tiers)

Not yet implemented; when they land: the **`memory_relevance` world flag** (`""`/`shadow`/`on`)
is Tier 1 (config, boot); the **embedding model pin and `embedding` route** are Tier 1
(`llm.json`); the **relevance-term weighting and divergence threshold** are Tier 4 doctrine
(the spec deliberately defers the go/no-go threshold to the US2→US3 gate); mental-map
**confidence-decay constants** (041) join the belief half-life family in Tier 4.

---

## 6. The promotion path (how a const becomes a dial)

World-01 makes the pattern clear: the dials we most wanted to turn mid-run (fire economy, refuel
threshold, conversation cadence) are all doctrine constants, and hand-editing them under a live
event-sourced world is unsafe-by-default. The promotion path for any constant that earns dial
status:

1. **Const → manifest field** read at boot, defaulted to the current constant, with a validation
   clamp (the `max_tokens`/`loop_max_rounds` pattern in `llm/config.go`).
2. **Log the value as an event** at boot (or on change), so replays reproduce behavior — the same
   discipline `calibration.json` and `format_version` already follow.
3. **Only then** consider hot exposure (IPC command, angel tool, TUI setting), with reducer-side
   enforcement of bounds — the clock-speed/charge-economy pattern.

Candidates that have earned promotion on world-01 evidence: `refuelDyingBelow`,
`fireBurnPerWood`, a conversation pair-cooldown, `gruEmergePerMille`, and the planner cadence.
Candidates that have *not*: everything nobody has yet needed to touch. Dials should be earned by
evidence, not speculatively multiplied.

---

## 7. Decisions & dispositions (2026-07-24)

Reviewed with the operator; each item below is now a board task (the plan of record).

| Decision | Disposition |
|---|---|
| Survival is table stakes: villagers are simple people who at minimum know how to not die. Err toward survival instinct wherever the reflex layer has a gap. | Doctrine; applied first to fire — **TASK-108** |
| Tuning manifest (`tuning.json`) is the promotion mechanism; once world-01 is dialed in, the tuned values become the standard default for new worlds. | **TASK-107** (keystone; TASK-108 rides it) |
| Conversation novelty gate is accepted **as a removable shim** — it compensates for model-side variety; if conversations later feel less dynamic than wanted, look there first. | **TASK-109** (leak diagnosis first) |
| The malformed-call storm was diagnosed as **cap design, not model grammar**: ~93% of the 807 rejections are "exceeds text cap", across all tiers including cloud. Fix is clamp-don't-reject on expressive fields; conversation **stays on gemma** (no re-route, no `tool_mode` change). | **TASK-110** |
| Roster shrinks for now: `collect_water` and `bathe` leave the villager loop roster (water has no consumer). Revisit only if a thirst need is designed. | **TASK-110** |
| The angel acts on its own authority for survival (visions/miracles, still charge-gated) — not merely warn the player. | **TASK-111** |
| Direction (firm): the metatron becomes a **first-class autonomous agent** — the villager construct (mind loop, memory, cadence, persona files) with a god-mode roster, extra context files, and levers. The request-driven console model is transitional. | **TASK-112** (full spec before implementation) |
| Estimator amnesia is real: persist learned s/pt across restarts, reseed from max(calibration, persisted). | **TASK-113** |
| Applied same day, config-only: `llm.json` `metatron_watch` route + explicit budgets; fresh `promptworld calibrate`; daemon restarted clean. | done (§4 rows 2–3) |

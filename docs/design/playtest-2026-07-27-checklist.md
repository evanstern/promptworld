# Operator playtest checklist — post-sweep verification (2026-07-27)

Covers everything the reorient 2026-07-26 sweep (PRs #112–#120) and the
faith-directives sweep (PRs #123–#127) shipped. World: `playtest-1` (seed 46103,
scenario `cold-dawn`, stage 1, teaching posture). Attach: `promptworld ui playtest-1`.

Legend: ⏱ = needs game-time to elapse (raise speed with `]` or leave it running);
🔁 = needs a second world or a restart. Everything else is a few minutes.

Record findings as board cards (or notes on the named review cards). The recorded
review items each test feeds are listed at the bottom.

## 1. First contact (quickstart + teaching floor)

- [ ] `?` overlay → guardian section shows the **forward ladder**: all four stages,
  identity · concept · earned/next · unlock evidence; stage-1 marked earned, stage-2
  marked next. Compare against `promptworld stages` — the two must agree exactly.
- [ ] Press `m`, ask the quickstart's sample ask: *"show Ash a vision of the fire
  dying"*. Expect: a charge spent, a vision event in the chronicle, Ash's behavior
  visibly influenced within a few minutes of game time.
- [ ] `docs/player/getting-started.html` step 5 matches what you just did; the
  `?` overlay's example asks include the new plan-layer verbs.

## 2. The exercise loop (rubric truth — reorient decision 1)

- [ ] Press `6`: cold-dawn's rubric gauges render real evaluator state (no
  permanently-pending term). Terms should read as survive-to-dawn / no freeze
  deaths / a watch in force.
- [ ] ⏱ The cold snap arms at 22:00 day 1. Before it: place a warmth watch
  (*"watch for anyone freezing tonight and send them a vision"*) or don't — decide
  whether you want a pass or an instructive fail.
- [ ] ⏱ At dawn day 2: **pass** → ceremony card with ✓ rows backed by evidence refs,
  stage-2 unlock event, `stages --json` flips earned. **Fail** (someone froze) →
  postmortem renders **✗ with the real count** (e.g. `✗ no one freezes —
  agent.died: 1`) — the false-checkmark bug class this sweep killed. Either way,
  every verdict must trace to a recorded event, not vibes.
- [ ] The cold snap itself must read as **weather, not an authored event**: raw
  chronicle (`r`) shows plausible cold/warmth events with no scenario marker.
  If you can tell it was scripted from in-game evidence alone, file that.

## 3. The plan layer (directives — operator-firm hardness ruling)

- [ ] Console (`G`): *"survey the land around the village"* → deterministic fact
  sheet (terrain mix, distances, passability), **no charge spent**. Ask twice —
  identical answer.
- [ ] *"Place a structure site for a fire at <pick a spot from the survey>"* →
  ◇ mark on the map; villagers **know about it** (check a villager's detail/journal
  or watch someone route toward it without being individually told).
- [ ] *"Direct the village to build that fire before nightfall"* → directive lands;
  in the villagers tab, watch someone break off wandering/prep to execute
  (the DIRECTIVE rung) — but anyone hungry/freezing handles survival FIRST.
- [ ] **Interruption is life**: while someone works the directive, watch a hail or
  conversation pull them off — they must chat and then RESUME the directive
  without your intervention. If directives suppress social life, that violates
  the ruling — file it.
- [ ] Cancel a directive mid-walk → the villager abandons the goal gracefully.
- [ ] ⏱ Issue a deliberately impossible directive (structure site on water /
  hopeless deadline) → TTL expiry event, **faith −4**, no thrash.
- [ ] Feel check (recorded review item): do the tools being **charge-free** feel
  right, or does spamming directives feel exploitable?

## 4. Faith + prophecy (the mana loop)

- [ ] Guardian strip: the **faith segment** renders (`faith 50` at genesis; dashed
  `faith —` would mean a wiring bug on a fresh world — file it).
- [ ] Directive fulfilled → **+8** visible on the strip; the regen forecast line
  updates when you cross a band (75+ = 4h, 40–74 = 6h, 15–39 = 12h).
- [ ] *"Prophesy that the fire will be built by dawn"* → one charge spent, prophecy
  listed in the console's context. Try to cancel it → refusal with counsel
  (uncancellable is doctrine).
- [ ] ⏱ Prophecy verdict: fulfilled **+12** / failed **−15**, with an in-fiction
  chronicle line and a witness memory on the affected villagers. Late truth (comes
  true after deadline) must mint nothing.
- [ ] Feel check (recorded review item): are the magnitudes right? Is prophecy at
  1 charge too cheap for +12?
- [ ] 🔁 The spiral: on a scenario world, drive faith below 15 (fail prophecies,
  let directives expire) → regen **stops** (forsaken). Ambient world → floors at
  24h instead. This is the FR-005 posture decision — judge whether forsaken feels
  fair or hopeless.

## 5. Neglect detector (the Oak shape)

- [ ] Hard to trigger in a healthy village — that's the point. Shape: a villager's
  need below the danger band for 2+ game hours with **zero** intents in that
  need's class. If it fires: a **whole-line red chronicle alert** + that villager's
  map glyph in critical styling + a high-salience memory the planner reacts to.
- [ ] If you see a slow death with NO alert — that's the bug class this exists to
  kill; capture the world dir and file it.
- [ ] False-positive check: an alert on a villager who was actively handling the
  need (foraging while hungry, walking to warmth) is a mis-fire — file it.

## 6. Map interrogation (look-cursor + village lens)

- [ ] `v` → cursor mode: hjkl/arrows move (HJKL ×8), camera pushes at the margin,
  `c` snaps, panel sizes NEVER change. Click a tile — same result as keyboard.
- [ ] TILE pane: contents in fixed order (agents → piles/chests → structures →
  terrain), warmth+light header with plain-language notes; drill into an agent
  (villager detail), an event (raw JSON), a chest (contents).
- [ ] `esc` releases exactly ONE layer per press: drill-in → rows → cursor → out.
- [ ] Inspect a **designation tile** — the ◇ should appear with registry prose.
- [ ] **Reverse jump**: click a strip glyph → camera centers that villager; click a
  roster row in the villagers tab (or press `J`) → same; a dead villager jumps to
  their grave.
- [ ] Chronicle raw mode (`r`): new events name agents inline (no index-only rows);
  ⏎ on journal/faith/neglect entries now jumps to source (they didn't before).

## 7. The duel (fork + compare) 🔁

- [ ] `promptworld stop playtest-1` (cuts a snapshot) then
  `promptworld fork playtest-1 playtest-b` → summary shows lineage + **wallet
  carried** (forking never mints budget).
- [ ] Edit `~/.promptworld/worlds/playtest-b/charter.md` — change the guardian's
  doctrine meaningfully (e.g. forbid visions, or make it warmth-obsessed).
- [ ] Start both; let them live a day. `promptworld compare playtest-1 playtest-b`:
  plain-language scoreboard per side (no raw enums anywhere — file any), lineage
  line, a divergence marker at the first real difference, interleaved chronicles.
- [ ] Fork and compare WITHOUT changing anything → the honest
  "identical since the fork" line.
- [ ] A lost duel should read like a postmortem (no-blame register), not a scolding.

## 8. Ops honesty (background checks while you play)

- [ ] `promptworld status playtest-1` any time: provider health, horizon, posture,
  stage, exercise lines all present and truthful; kill the gemma endpoint briefly →
  the TUI's LLM badge and status go loud rather than silently brain-dead.
- [ ] Speed up (`]`): the governor sheds if thinking can't keep up — speed is a
  ceiling, not a promise; verify it recovers.
- [ ] After the session: does the story the chronicle tells match what you watched?
  Narration drift is a finding.

## Where findings go

- Faith deltas / prophecy pricing / no-cancel → note on **TASK-118**'s card.
- Charge-free + every-stage plan tools, directive hardness feel → **TASK-157**.
- Retry accommodation, ambient drama floor, rubric-gauge exposure → the reorient
  synthesis's parked items (docs/design/reorient-2026-07-26-ui.md §Open questions)
  — playtest evidence is their named resurfacing trigger.
- Anything crashy/wrong → new board card with the world dir + seq range attached.

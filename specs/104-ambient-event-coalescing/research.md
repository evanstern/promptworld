# Research: spec 104 contracts (T001)

Decisions and rationale pinned BEFORE code, per tasks.md T001. Everything here
is downstream of the operator rulings (spec.md fork section) and plan.md D1–D6;
where implementation forced a finer-grained call than the plan records, the
call and its rationale are stated explicitly and flagged `[resolution]`.

## 1. The advancement convention (D1 made precise)

`State.AdvanceTo(target)` executes every pending derived item with scheduled
tick strictly below `target`, in the fixed order §2. `Apply(e)` calls
`AdvanceTo(e.Tick)` before dispatching the event. The live loop calls
`AdvanceTo(nextTick)` at the START of each tick, before `stepEvents` —
and deliberately does NOT advance at the end of the tick.

**[resolution] Items scheduled at tick t execute after ALL events recorded at
tick t have applied, and before any event at a later tick.** Live, that is the
`AdvanceTo(t+1)` at the start of `runTick(t+1)`; on replay it is the
`AdvanceTo` inside the first later event's `Apply` (or the folder's explicit
trailing call). The alternative — an end-of-tick `AdvanceTo(t+1)` inside
`runTick(t)` — is DISQUALIFIED by the injection door: commands land events at
the loop's current tick after `runTick(t)` returns, so live would run tick-t
items before an injected tick-t event while replay (strictly-before) would run
them after — a live/replay divergence. With the chosen convention both live
and replay order tick-t items after every tick-t event, always.

Consequences, accepted and documented:

- An event at tick t observes derived state through t−1 (matches today: tick
  t's `agent.moved` folds in tick t's batch and is read by others from t+1).
- Live reads between ticks (status/state door, snapshots) lag derived items by
  at most one tick. Snapshots marshal the per-item watermarks, so recovery is
  exact wherever the boundary falls.
- "State at tick boundary t" for equivalence testing means
  `fold(events with tick ≤ t)` + `AdvanceTo(t+1)` — the state any tick-(t+1)
  observer sees. The harness sweeps this convention.

## 2. Within-tick ordering contract

At one derived tick u, items run in exactly this order (mirroring today's
stepEvents emission order for the same families):

1. **Needs minutes** (u%60==0): agents by ascending index — decay, near-death
   latch, trajectory-anchor roll, neglect band anchors, watermark.
2. **Agent movement steps**: agents by ascending index — position update,
   `markExplored`, `notePresence(i, u)`, cursor/watermark.
3. **Gru beat** (u%gruMoveEveryTicks==0): stalk/prowl decision over the
   advanced state; skipped when `Gru.LastAttack == u` (an attack recorded at u
   precludes the move, exactly today's exclusivity).

The equivalence harness (T006) enforces this contract byte-wise. Note the gru
beat runs AFTER agent steps within the tick (plan D4's pinned order): the gru's
derived decision reads agent positions through u where the old emission read
through u−1 — a new-world behavior definition, deterministic on every fold
path, outside any old-log contract (gru.moved arms replay old logs unchanged).

## 3. The coalescing regime marker (the double-fold guard)

**[resolution] All three emission changes are gated on ONE genesis-pinned
marker: `TuningState.NeedsCheckpointMinutes > 0`.** A pre-104 recorded
`sim.tuning_applied` payload (field absent) resolves to 0 = LEGACY; a nil
`State.Tuning` is legacy; `defaultTuning()` carries 10, so new worlds pin the
regime ON in their genesis tuning event (spec 057). While legacy:

- the executor emits per-step `agent.moved`, per-minute `agent.needs_changed`,
  and `gru.moved` EXACTLY as before (the old code paths are kept, not
  emulated), and
- advancement is structurally inert for needs and gru (movement is inert by
  construction — no `agent.path_started` ever installs a segment), and
- the `agent.needs_changed` / `gru.*` reducer arms set NO new fields.

This makes SC-002 hold at full strength: a pre-change log replays on the fixed
build to *hash-identical* state (not merely same-build-consistent), because no
arm writes a new field and no derived channel contributes while the regime is
off. It also kills two hazards that a marker-free design cannot escape:

- **Old-snapshot mega-decay**: a pre-104 snapshot carries no needs watermark
  (omitempty ⇒ 0); a free-running derived decay would replay the whole
  (0, snapshotTick] gap as pending minutes on the first `AdvanceTo`. With the
  regime off for such worlds, advancement never looks.
- **Old-log gru double-motion**: old logs carry `gru.moved` rows on the same
  beats a derived gru would fire; both running would double-move.

**Regime flip** (an old world's operator adds `needs_checkpoint_minutes` — or
any tuning.json, since sparse manifests resolve absent dials to their doctrine
defaults, K included): the `sim.tuning_applied` arm detects the legacy→on
transition and stamps every living agent's `NeedsSyncTick`/`NeedsEmitted` and
the gru's `Done` to the event's tick — a pure fold effect, so replay of the
flip is exact and no gap is ever re-decayed. Forward-only emission change,
sanctioned (ruling 5 forbids only retroactive relief).

## 4. Movement: payload, segment, schedule

`agent.path_started{agent, path, move_every, phase}`:

- `path`: the full departure BFS tile sequence (excluding the start tile,
  ending on the intent target), from the same `bfs` the per-step walker used —
  deterministic neighbor order preserved.
- `move_every` = 5 and `phase` = (agentIndex*3) % 5, baked at emission
  (spec 092: advancement never reads the compiled cadence). The path-tile
  2x RULE (extra beat at phase slot 2 when standing on a "path" structure)
  is code, like any reducer arm; its slot constant joins the spec-092 audit.

`Agent.Path *PathSegment` (omitempty pointer — the Journal/Map precedent):
`{path, next, move_every, phase, done}` where `next` is the cursor (index of
the next tile to step onto) and `done` the advancement watermark (processed
through this tick). The `agent.path_started` arm installs the segment with
`done = e.Tick` (steps fire strictly after the departure tick).
`agent.path_truncated{agent, x, y}` sets the position from the payload and
clears the segment (outcome in payload, never recomputed).

Step rule at tick t: `ph := (t + phase) % move_every`; step when `ph == 0`, or
when `ph == 2` and a path structure stands on the tile the agent currently
stands on (advanced state at t — same tile the old executor tested). The
arrival step (cursor reaches the end) clears the segment derivationally.

**Route pinning [resolution]:** the walk follows the departure-time BFS path.
The old build re-ran BFS every beat, and equal-length ties could re-break
differently mid-walk; under segments the route is pinned at departure and
re-tie-broken only across a truncation. Per-step sighting/explored bookkeeping
is EXACT for the walk actually taken (ruling 2's contract, proven by the
harness); route choice is upstream of that contract and legitimately differs
across builds (spec: paired-seed streams differ; volume, never bytes, is
compared across builds).

**Truncation [resolution — deviation from plan D2's letter, recorded]:**
plan.md D2 lists deviation sources all emitting `agent.path_truncated`.
Implemented instead as: every event that invalidates a walk's premise clears
the segment IN ITS OWN ARM (`agent.intent_set`, `agent.intent_done`,
`agent.intent_failed`, `agent.build_failed`, `agent.recovery_stalled`,
`agent.slept`, `agent.died`, `gru.attacked`, `social.hailed`,
`guardian.entity_moved`, `clock.paused` for all walkers) — the interrupting
event IS the closing record, the position is wherever advancement actually
got (exact), and advancement stays unconditional with no corpse-walk /
frozen-walker window that per-site co-emission could leave open if a site were
missed. `agent.path_truncated` is emitted by the executor for the one
deviation nothing else records: a blocked path (next step tile impassable —
wall built mid-segment), after which the next tick re-plans (fresh
`path_started`, or `intent_done` when unreachable). Fold-equivalent to the
plan's shape, fewer rows, one emission site. Flagged for the planning tier.

**Arrival observation:** the executor at the arrival step's tick predicts the
step from the installed segment (same beat rule, pre-tick state) and emits
`agent.place_observed` at that tick — the exact tick and pre-batch inputs the
old co-emitted observation used. Razor edge accepted: an event injected at the
same tick that clears the segment after the prediction leaves one observation
whose arrival never executed (recorded; replay refolds it identically).

**Pause:** `clock.paused`'s arm truncates all in-flight segments (position =
wherever advancement got, which is exact at the pause tick); on resume the
executor re-plans from the standing position. Fold-side, replay-exact — the
plan's "pause truncates all, advancement stays unconditional" delivered
through the arm rather than a co-emitted row (same [resolution] as above).

**Hail:** `social.hailed` truncates (arm); when the hail lifts the executor
re-plans. The old build froze the walker and resumed the same walk; the new
shape re-plans from the frozen tile — same tiles walked in the common case
(BFS from a mid-path tile of an unchanged world re-derives the suffix or an
equal-length alternative).

## 5. Needs: dial, watermark, crossings

- Dial: `needs_checkpoint_minutes`, clamp [1, 60], doctrine default 10,
  `defaultTuning()` carries it (genesis-pinned, spec 057), manifest key
  `needs_checkpoint_minutes`, payload field pointer+omitempty (pre-104 events
  decode nil → 0 → legacy). K=1 = today's per-minute cadence exactly.
- Derived decay (regime on): advancement applies `decayNeeds` per game-minute
  per living agent, plus the same near-death latch, trajectory-anchor roll and
  neglect band-anchor moves the arm performs — one shared helper so the two
  can never drift. Environment inputs read from the advanced state at the
  minute tick (night/warmAt/shelter/cold-snap/asleep "through u"); the old
  heartbeat read fires/shelter/sleep pre-batch but `night` through-u — the
  through-u convention differs only when an env-flipping event lands on the
  same minute boundary (razor edge, deterministic, recorded).
- Watermark: `Agent.NeedsSyncTick` (omitempty) = decayed-through tick, set by
  the regime needs arm and by derived minutes; `Agent.NeedsEmitted *Needs`
  (omitempty) = last-emitted absolutes, set only by the regime needs arm.
  Legacy folds touch neither (SC-002 at hash strength).
- Emission (executor, regime on, every minute): compute this minute's decay
  `n` from pre-tick state exactly as today, then emit `agent.needs_changed`
  with `n` when (a) checkpoint grid due — `(tick/60) % K == 0`, or (b) any
  boundary CROSSED between `NeedsEmitted` (nil ⇒ emit) and `n`, or (c) death
  fires (`n.Health == 0` also crosses zero ⇒ (b) already emits; `agent.died`
  rides as today). Boundaries: spec-062 danger bands (food/warmth/rest),
  near-death (`nearDeathBelow`, health), and zero — both directions, so
  guardian survival watches and standing-order hysteresis see danger AND
  recovery at today's per-minute latency (comparing against last-EMITTED
  values, not the previous minute, catches mid-window jumps such as eating —
  the hysteresis re-arm case).
- Near-death memory and the death/witness block run per-minute on `n` exactly
  as today (their conditions are unchanged; a near-death entry is a crossing,
  so its needs event always accompanies it).

## 6. Gru derivation

Stalk/prowl moves into advancement (shared decision function, RNG purpose
"gru-prowl" preserved verbatim); `gruStep` keeps emergence, withdrawal,
sightings, attack. `Gru.Done` (omitempty) is the beat watermark, set by the
`gru.emerged` arm (no same-tick move — today's parity), by derived beats, and
by the regime-flip stamp; the `gru.moved` arm is byte-for-byte untouched.
Derived beats run only while the regime is on, the gru is abroad, and
`LastAttack != u`. The "gru-prowl" purpose string, `gruMoveEveryTicks`, and
the greedy-stalk neighbor order join the spec-092 audit (replay-load-bearing
for new logs; retune ⇒ format machinery).

## 7. rebaseTicks classification (guardian.time_snapped)

- `Agent.NeedsSyncTick`: SHIFT (a decayed-through anchor).
- `Agent.NeedsEmitted`: KEEP (absolute values, not a tick).
- `Agent.Path`: CLEARED by the rebase arm [resolution] — the segment's beat
  phase arithmetic is absolute-tick-based, so a shifted `done`/schedule would
  re-phase the walk; truncating at the snap (villagers re-plan next tick) is
  deterministic and honest. Recorded as the segment's taxonomy entry.
- `Gru.Done`: SHIFT.

## 8. Fold-path inventory (who needs explicit AdvanceTo)

`Apply` gives every event-folding consumer advancement for free. Explicit
`AdvanceTo` calls added:

- the live loop: `AdvanceTo(nextTick)` at the start of `runTick` — items at
  the previous tick execute after every event of that tick (§1), including
  command-door injections;
- the TUI replica: `AdvanceTo(status tick)` on the 1s status poll (the tick
  driver between events) — the SAME `AdvanceTo(T)` posture the daemon holds
  at tick T (items ≤ T−1), so replica bytes can never lead the daemon's;
- the mind replica: `AdvanceTo(replica.Tick)` per absorb batch, then the
  first-adjacency encounter sweep (transition-into-adjacency over the
  advanced replica, per-pair cooldown preserved; mind-side heuristic,
  outside replay scope);
- `daemon.replayToTick`: trailing `AdvanceTo(cutoff)` after the fold (covers
  a quiet stretch between the last event and the cutoff);
- `worlds.OfflineState`: trailing `AdvanceTo(lastTick)` beside its existing
  Tick bump (compare/status/ps read what a live daemon at that tick holds).

Daemon RECOVERY needs no change: Tick = max(snapshot, last event) leaves
items at that tick pending exactly as the uncrashed flow did (watermarks ride
the snapshot), and the re-lived quiet stretch re-advances identically.
Morgue/scribe/guardian folds read at last-event posture — exact under §1 with
no trailing call. The gru/movement derivations require the world map: every
replica that folds events already attaches it (verified: mind.go:157,
embedder.go:136, scribe.go:51, morgue.go:102, guardian.go:260, tui.go:559,
daemon.go:689/736, migrate.go:441, probe.go:199).

## 9. Test-weld inventory for the new vocabulary

`agent.path_started` / `agent.path_truncated` must join, or the sweeps fail:
`sim.PayloadCatalog` (payloads.go), `docs/wiki/event-types.md` backticked
rows (swept by BOTH `TestPayloadCatalogCompleteness` and the TUI
`TestCatalogSweep`), `internal/tui/digest.go` `digestRegistry` +
`catalogFixture` (+ `subjectRegistry` rows), and `TestPayloadAgentRefSweep`
(the `agent` field must be `AgentRef`). New int64 state fields
(`NeedsSyncTick`, `Gru.Done`, `PathSegment` fields) must be classified in the
`rebaseTicks` taxonomy or `TestRebaseTaxonomyComplete` fails. Known follow-on:
a player standing order recorded with `event_types:["agent.moved"]` silently
stops matching on a regime world (structural matcher compares type strings
verbatim) — noted for the wiki's guardian-orders row.

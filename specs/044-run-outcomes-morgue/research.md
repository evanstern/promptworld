# Phase 0 Research: Run outcomes, the morgue file, death escalation, and graves

**Spec**: `specs/044-run-outcomes-morgue/spec.md` | **Date**: 2026-07-25

All decisions below were grounded against the code (wiki pinned at cdf398f agrees with
source at the cited lines). No NEEDS CLARIFICATION items remain.

## R1 — `run.ended` detection: same-batch, executor-emitted

**Decision**: Detect "last villager dead" inside `stepEvents` (`internal/sim/executor.go`),
same-batch: after the needs-heartbeat loop emits this tick's `agent.died` events, count
deaths emitted in the batch against remaining living agents and, if none remain and the
world is not already ended, append one `run.ended` event to the same batch — ordered after
every same-tick death. Add a small `livingCount(s)` helper (today there are only ad-hoc
`.Dead` loops; genesis size is the `agentCount` constant, `state.go:117`).

**Rationale**: `stepEvents` is pure over (state, map, tick) — the
`metatron.order_expired` / `charge_regenerated` precedent — so replay reproduces the run
end with no daemon logic (FR-020). Same-batch placement satisfies the "two die on the same
tick" edge case (exactly one run-end record, after all deaths) and the spec's "time stops
advancing" (no extra tick is consumed, unlike a next-tick check). The store's single-writer
doctrine (loop owns the log; store errors fatal) also requires emission from within the
loop, never a daemon goroutine.

**Alternatives considered**: (a) next-tick check at the top of `stepEvents` — simpler but
advances one extra tick and splits the record from the final death's batch; (b) daemon-side
watcher injecting via the operator door — violates single-writer/pure-executor discipline
and would need a whitelist entry. Rejected.

## R2 — Halt seam: event-sourced `State.Ended`, Paused-mode precedent

**Decision**: The `run.ended` reducer arm sets a new `State.Ended` flag (with `omitempty`,
modeled on `Paused bool`, `state.go:25` — no `format_version` bump, old snapshots
byte-identical). `Loop.Run` gains an ended branch above the existing paused branch
(`loop.go:404-417`): no timer, block on commands/ctx. `handleCommand` gates mutating
commands when ended — `pause`, `resume`, `set_speed`, `govern`, `inject_intent`,
`inject_social` refuse; `status`/`state` (read-only) keep working. The daemon's serving
layer (`srv.Serve()` in its own goroutine) is untouched.

**Rationale**: the paused branch already proves the loop can idle indefinitely while reads
are served. Exiting `Run` is the wrong seam — it kills `Do`/`DoState` and the whole daemon
(`daemon.go:408-427`). Because the flag is event-sourced, `recoverState` (snapshot +
`ReplayEvents`) restores it on restart for free (FR-004), and migration tooling cannot
resurrect a finished run without rewriting history (spec edge case). A `world.json`
manifest field was rejected: not replay-derived, and hand-editable.

## R3 — Status surfacing: additive `omitempty` on the existing surfaces

**Decision**: Surface the run-end fact as an additive `omitempty` field on
`ipc.ClockStatus` / `StatusData` (`protocol.go:108,177` — the governor-trio precedent),
populated from `Loop.status()` (`loop.go:375`); add it to `promptworld status --json`'s
offline snapshot shape (`commands.go:531-544`) so an ended-but-stopped world also reports
it. `StateData` carries it automatically once it is on `State`.

**Rationale**: FR-005 (machine-readable for TASK-119 scenario machinery) is satisfied with
zero new plumbing: durable event in the log + status field live and offline. Attached
clients receive the `run.ended` push via the existing `subscribe` fan-out
(`server.go:545`, `notify` → `Broadcast`) — "pushed live" is free.

## R4 — Escalation predicate: raw pre-attack health threshold, no RNG

**Decision**: "Already weakened" = pre-attack recorded health strictly below the existing
near-death band: `a.Needs.Health < nearDeathBelow` (200, `agents.go:570`). The change is
the floor conditional at `gru.go:131`: healthy targets keep
`maxInt(gruWoundFloor, h-gruWound)`; weakened targets take `maxInt(0, h-gruWound)` — a
pure conditional on recorded state, no new RNG purpose. Update the doctrine comment
(`gru.go:12-20`) and the `GruAttackedPayload.Health` doc (`gru.go:232`).

**Rationale**: FR-014 demands determinism as a function of recorded state — a pure
predicate is the codebase-idiomatic choice (`gruStep` is documented "Pure over (pre-tick
state, map, next tick)"). The alternative predicate — the latched `a.NearDeath` bool — was
rejected deliberately: its hysteresis (clears only at health ≥ 400, `state.go:1421-1425`)
would let the gru kill a villager who had recovered to health 350, violating the spec's
"already below the near-death band" wording and the one-hit-randomness doctrine. Since
`gruWound` (250) ≥ `nearDeathBelow` (200), any weakened victim dies outright — no
"survives at exactly 0" ambiguity.

## R5 — Gru-kill death path: emit `agent.died` from `gruStep`, replicate witnesses

**Decision**: When the escalated attack reduces health to 0, `gruStep` emits `agent.died`
(cause `"gru"`) in the same batch immediately after its `gru.attacked` event, and
replicates the witness-death memory loop inline (the executor's block at
`executor.go:137-146`, `salWitnessDeath = 10`) — the same idiom `gruStep` already uses for
attack witnessing (`gru.go:135-144`, `salGruWitness = 7`).

**Rationale**: two verified gaps make this necessary: (1) the `gru.attacked` reducer arm
does not mark death (only sets health/wakes/records `Gru.LastAttack`); (2) gru attacks
happen off the `%60` needs heartbeat, so the executor's witness-death block never fires
for them. Emitting the standard `agent.died` event means the entire existing death path —
reducer (Dead flag, intent/hail clear, spec-013 inventory spill, `state.go:1426-1466`),
chronicle, TUI digest alert, metatron digest, scribe — runs unchanged (FR-015). The cause
taxonomy is an open string on `DiedPayload` (`agents.go:862-865`); consumers switch on
event type, not cause value, so adding `"gru"` is additive. `chronicleNote`'s
`gru.attacked` line ("left them wounded", `narrate.go:85-89`) must branch for a killing
blow. Light/shelter protection (`gruProtected`, `gru.go:64-68`) sits upstream of target
selection and is untouched (FR-016).

## R6 — Morgue writer home: `internal/scribe`, whole-file re-render

**Decision**: The morgue is a scribe document: `renderMorgue` in `internal/scribe`,
triggered when an applied batch contains `agent.died` or `run.ended` (and re-rendered at
every boot, like all scribe files). Path helper `MorguePath()` → `morgue.md` joins the
cluster in `internal/world/world.go:205-236`. The file is a full re-render from the
scribe's replica plus a bounded event scan — never an append.

**Rationale**: scribe is the shipped "regenerable views, never a source of truth" pattern
(`scribe.go:1-4`): always-on (wired before the LLM gate in the daemon's notify fan-out,
`daemon.go:126-135`), holds its own replica, re-renders dirty files per committed batch —
exactly decision 5's no-AI posture and FR-011's "regenerable from the event history alone"
(the file being deleted/hand-edited is healed by the next render). Whole-file re-render
makes replay byte-identity (SC-004) trivial. `grep` confirms `run.ended`/`morgue` are
greenfield — no collisions.

## R7 — Epitaph facts: replica state at death + typed event scan for deeds

**Decision**: At the death event, epitaph fields come from:
- **days survived** — death tick → `clock.GameTime(tick).Day` (all villagers exist from
  genesis, so days survived = death day);
- **cause** — `DiedPayload.Cause`;
- **relationships** — `State.Relations` filtered on the dead villager (`social.go:18-23`);
- **debts owed/owing** — `State.Debts` filtered Debtor/Creditor == dead, `Status=="open"`
  (`social.go:25-32`); the death reducer never touches debts, so they persist as evidence;
- **notable memories** — the villager's retained `Memories` at death (salience-ranked),
  supplemented by an event scan of `agent.memory_added` above a salience threshold for
  lifetime highlights (nightly consolidation deletes consolidated memories from state,
  `consolidate.go:271`, so at-death state alone under-reports a long life);
- **notable deeds** — an event-log scan by type (the `type` index exists for exactly this,
  `docs/wiki/event-log.md:49-52`) over the chronicle's curated notable-event vocabulary
  (`chronicleNote`'s switch, `narrate.go:56+`: built, promises broken/kept, thefts, gifts,
  governance arcs, gru events), filtered to the dead villager.

**Rationale**: everything is either reducer state (deterministic by construction) or a
deterministic scan of the immutable log — replay-identical (FR-011, SC-004), model-free
(FR-007). Reusing the chronicle's notable-event set keeps "notable" a single curated
vocabulary instead of a second judgment surface.

## R8 — Charter/orders evidence: fingerprint-at-effect event + active-orders snapshot

**Decision**: Two halves:
- **Orders** (already event-sourced): at death, filter `State.MetatronOrders` on
  `Status == "active"`; record each order's condition/action text and watch subjects
  (`EventTypes`, `Agent`, `Keywords`). Zero new machinery.
- **Charter revision identity** (missing today — verified: no hash/revision exists;
  the only signal is `charterIsDefault`): introduce a content fingerprint at effect time.
  When a Metatron turn loads the effective charter (`loadCharter`, `charter.go:28-46`),
  compute a short content hash; when it differs from the last recorded one, the turn
  pipeline emits a `metatron.charter_observed` event (fingerprint + default/custom flag +
  tick) through the existing turn-effect landing path. The reducer records the current
  fingerprint on state; the morgue aligns each death against the most recent
  fingerprint-change event at or before the death tick — an event-sourced revision
  timeline.

**Rationale**: the spec's assumption anticipated exactly this ("if revisions are not yet
individually identified, the plan phase introduces the minimal identification needed").
Charter reads happen only at turn time, so fingerprint-on-effect is the honest semantics:
it records what the angel actually ran with, not file mtimes. Evidence stays evidential
(hash + custom/default + timeline), no scoring language (FR-008). Alternative — hashing
the file at death time from the scribe — rejected: races the hot-reload discipline and
records a charter the angel may never have executed.

## R9 — Epilogue: narrator-routed, landed as a recorded event

**Decision**: The epilogue rides the existing narrator machinery in `internal/mind`:
same `llm.KindNarrator`, same `routeVerdict("chronicle", …)` decision class, enqueued on
the existing single-flight narrator worker when it absorbs an `agent.died` /` run.ended`.
The prose lands as a recorded, whitelisted event (`morgue.epilogue`, via the same
`InjectSocial` landing path the chronicle uses), and the scribe renders it into
`morgue.md` beneath the facts, clearly delimited.

**Rationale**: this copies the chronicle's proven split verbatim — mind narrates into
recorded events, scribe renders files from the replica — so the morgue file remains a pure
regenerable render even with prose in it (FR-010/011/012). No-LLM worlds never construct
the mind driver (`daemon.go:156` nil-gate), so absence is structural silence, not an error
path. Narrator failure follows chronicle doctrine: a gap, never a stall. No new model-call
class (spec assumption honored); `morgue.epilogue` needs an `injectSocialWhitelist` entry
(`loop.go:193`) and a digest/catalog row.

## R10 — Grave representation: `Structure` kind `"grave"`

**Decision**: A grave is a `Structure{Kind: "grave"}` created by the `agent.died` reducer
arm at the death location (the same reducer-internal idiom as the spec-013 inventory
spill). It takes the durable freshness horizon by default (no `factHorizon` case needed).

**Rationale**: the structure route gets everything downstream for free — perception sweep
pickup (`executor.go:444-450`), `groundFactPresentIn` default case, place-fact telling in
talks (`tellablePlaces`), map correction, `send_vision` place grants — with zero new
perception code (FR-018's "later perceive" path). The known tension: a structure occupies
the tile for `buildSite` (`terrain.go:111-112`), i.e. graves block building. The spec's
edge case explicitly defers building-over-graves as a world-design question; a grave that
persists and blocks construction is the conservative default consistent with "the grave
persists and remains addressable". The alternative (a new overlay slice like `Piles`) was
rejected: it needs new perception, ground-truth, telling, and reveal plumbing on four
surfaces to reach parity. Hand-mirrored vocabularies to update: `PlaceFact.Kind` comment
(`mentalmap.go:49-51`), `placeFactKinds` in `internal/tool/registry.go:430`, the prompt's
landmark set (`prompt.go:204`), TUI glyph switch + legend (`views.go:418-472, 616`).

## R11 — Grief material: ride the shipped witness-death memory; verify, don't build

**Decision**: FR-019 is satisfied by the existing pipeline plus tests: the witness-death
memory ("Watched %s die of %s.", subject-tagged, tone −80, `salWitnessDeath = 10`) already
exceeds `rumorMinSalience` (4) and is therefore the strongest possible rumor-birth seed in
`TellableFor` (`social.go:508-520`), spreading through founded talks within the ambient
cadence. Work is: (a) an integration test proving SC-006's within-a-game-day rumor, and
(b) witnesses additionally gain the grave place-fact via their next perception beat
(≤ `moveEveryTicks` later, within radius 8 — same radius as witnessing), which the
structure decision (R10) gives for free. No new drives, no new memory class.

**Rationale**: the research verified the full memory→rumor→conversation chain is shipped
and death-witness memories already ride it; inventing a parallel grief mechanism would
violate the spec's "using existing mechanisms" constraint.

## R12 — TUI postmortem posture: header state token + dual-source derivation

**Decision**: A postmortem posture derived from **two** sources: the replica's
`State.Ended` flag (covers clients attaching after the fact — the snapshot path never
replays folded events) and the pushed `run.ended` in `applyEvent` / the 1s status poll
(covers the live transition, spec edge case "pushed live without reconnect"). Rendering:
an `ENDED` header state token replacing running/`PAUSED` (`views.go:113-115` precedent),
plus the chronicle/digest lines. New event types get `docs/wiki/event-types.md` catalog
rows and `digestRegistry` entries — `TestCatalogSweep` (`digest_test.go:210`) enforces
this mechanically.

**Rationale**: dual-source derivation is required for correctness, not belt-and-braces:
snapshot-attaching clients see only state, live clients see the push first. The header
token follows the shipped badge patterns and their tests (`views_test.go:105,161`).

## R13 — Test strategy: extend the shipped harnesses

**Decision**: (a) amend `TestGruWoundsNotExecutes` (`gru_test.go:149`) — the healthy half
stays as-is, the health-50 half inverts into the kill assertion; (b) new scenario tests on
the `gruTestState` pattern (stage healthy + weakened victims, single `stepEvents`, assert
`agent.died` presence/absence and cause); (c) run-end/morgue/grave determinism rides the
existing `TestDeterminismSameSeedSameTimeline` / `TestReplayRebuildsState` harnesses
automatically since everything flows through `stepEvents`/reducers; (d) morgue render
golden-file tests in `internal/scribe` (byte-identity across replay, SC-004); (e) TUI
header/digest tests per the badge-test patterns.

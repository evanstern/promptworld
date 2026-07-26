# Event & Reducer Contract Deltas: spec 081

No new event types. No payload shape changes. Two reducer arms gain
mutations; one emit site per act gains a companion memory; one absorb rule is
extended. Everything below is a delta against the spec 041 contracts.

## §1 `agent.chopped` (existing type — reducer semantics extended)

Payload unchanged: `{agent, x, y}` (HarvestPayload).

Reducer arm now ALSO, after the existing mutations and derived from the same
pre-mutation state:

1. removes the `tree` fact at (x,y) from the actor's mental map (no-op when
   absent), and
2. removes the `tree` fact at (x,y) from the mental map of every other
   villager that is alive, awake, has a map, and stands within
   `witnessRadius` (Manhattan diamond) of (x,y) at the event tick.

Removals are silent: no memory, no chronicle line, no companion event. The
mind replica applies the same arm, so replica maps stay byte-identical.

## §2 `agent.quarried` (existing type — reducer semantics extended)

Identical delta with fact kind `rock`.

## §3 Actor act memory (existing `agent.memory_added` type — new emit sites)

The executor's chop and quarry completions each append one companion
`agent.memory_added` for the ACTOR in the same batch as the act event
(the hunt-memory shape):

- chop: text `Felled the tree at (x,y).`, salience `salChop` (4)
- quarry: text `Quarried the outcrop at (x,y).`, salience `salQuarry` (4)
- origin `action`, situated by the actor's stand tile, `why` = the intent's
  Reason when present.

Witnesses receive no memory event. Memories continue to accrete ONLY via
`agent.memory_added` (TestMemoriesAccrete posture).

## §4 `agent.map_corrected` (unchanged type — narrowed emission set)

Emission logic in the perception sweep is untouched. Because on-scene facts
are now removed at act time, corrections can only name facts held by agents
who were dead, asleep, or outside `witnessRadius` at removal time — the
genuine return-discovery narrative. The correction's discovery memory,
absorb-trigger role, and reducer arm are all unchanged.

Invariant (new, testable): no event log produced by this code version
contains an `agent.map_corrected` whose (agent, kind, x, y) matches an
`agent.chopped`/`agent.quarried` where that agent was the actor or an awake
in-radius witness at the act tick.

## §5 Mind absorb (extended rule, `internal/mind/mind.go`)

`agent.chopped` and `agent.quarried` continue to arm their actor. NEW: they
also arm any villager within `witnessRadius` of (x,y) whose live intent
matches the cleared tile — `(TargetX,TargetY) == (x,y)` or
`(ResX,ResY) == (x,y)` — the same lost-premise rule the
`agent.map_corrected` arm applies. Non-matching witnesses stay quiet (their
map already updated; nothing they act on changed).

## §6 Explicitly out of contract

- `agent.foraged` (regrow semantics, never corrects), pile draining,
  structure removal, and `metatron` terrain removal keep today's behavior.
- Cross-version replay of pre-081 logs follows the project's existing
  reducer-evolution posture (TASK-75 note); determinism is guaranteed within
  one code version.

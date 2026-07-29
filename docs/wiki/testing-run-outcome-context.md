---
name: testing-run-outcome-context
description: Run-end/morgue/grave/escalation proofs and the decision-context assembler suites (recent-intent ring, journal excerpts, golden-identity prompt, replay-determinism harness). Split out of [[testing-strategy]].
kind: pattern
sources:
  - internal/sim/run_end_test.go
  - internal/sim/morgue_test.go
  - internal/sim/grave_test.go
  - internal/sim/gru_test.go
  - internal/sim/toolcheck_test.go
  - internal/sim/intentlog_test.go
  - internal/sim/needsanchor_test.go
  - internal/sim/journal_test.go
  - internal/sim/memory_test.go
  - internal/mind/context_test.go
  - internal/daemon/context_replay_test.go
  - internal/guardian/charter_observed_test.go
  - internal/mind/epilogue_test.go
verified_against: c61cd6c04ddfcd2a976c14a49ba071e8fd768a73
---

# Run-outcome & decision-context suites

**Run-outcome suites** (spec 044, TASK-31, [[morgue]]): the run-end/morgue/
escalation/grave surface is proven per layer. `internal/sim/run_end_test.go`:
`TestRunEndedOnceOrderedLast` (a same-tick multi-death batch declares the run
over exactly once, ordered after every death and its witness memories),
`TestEndedWorldEmitsNothing` (further ticks emit nothing — the `stepEvents`
top guard), `TestReplayRebuildsEnded` (from-genesis replay lands back in the
ended posture), `TestRunEndOmitemptyStable` (the three `State` additions are
`omitempty`, pre-044 snapshots byte-stable), and `TestEndedCommandGating`
(mutating commands refused, reads served, `inject_social` narrowed).
`internal/sim/morgue_test.go` covers the two injected arms
(`TestCharterObservedArm`, `TestMorgueEpilogueArm` — ring append in event
order, cap, agent −1 for the run-end epilogue) and
`TestEndedDoorAcceptsMorgueEpilogue` for the narrowed door's surviving type.
`internal/sim/grave_test.go`: grave placement at the death tile persisting
through replay, `buildSite` refusing a grave tile (research R10's deliberate
tension), the perception sweep granting the grave as a `PlaceFact`, place-tell
spreading it, and the witnessed-death grief rumor (SC-006).
`internal/sim/gru_test.go` gains `TestGruWoundInvariant` (the compile-pinned
`gruWound >= nearDeathBelow` escalation arithmetic) and
`TestGruEscalationScenario` (a weakened victim dies of cause `"gru"` with the
full death fallout; a healthy one keeps the survival floor — [[gru]]).
`internal/sim/toolcheck_test.go`'s `TestWhitelistDiffIdentical` — the
injection-whitelist tripwire — accepts exactly the two declared boundary
widenings (`guardian.charter_observed`, `morgue.epilogue`). On the mind side,
`internal/guardian/charter_observed_test.go` proves the first turn emits the
charter observation, an unchanged fingerprint stays silent, and an ended
world skips it (the shared fixtures pre-seed `charterFP` so turn tests keep
counting exactly the batches they drive); `internal/mind/epilogue_test.go`
proves absorbing a death or `run.ended` queues an epilogue, good prose lands
as ONE `morgue.epilogue`, and a narrator failure is a gap, never a stall
(FR-010).

**Decision-context suites** (spec 043, TASK-105, [[decision-context]]): the
context-grounding surface is proven per layer. Reducer-side,
`internal/sim/intentlog_test.go` pins the recent-intent ring —
`agent.intent_set` appends, done/failed stamp the newest open record, a
rejected intent appends already closed, an expired plan step stamps its open
record (or appends an unfired one), quick-succession overrides preserve
order, wraparound at the cap, and byte-stable replay — and
`internal/sim/needsanchor_test.go` the trajectory anchor's window-edge roll
(unset first window renders steady, refresh at the edge, a sleep spanning the
window). `internal/sim/journal_test.go` gains the `SelectJournalExcerpts`
matrix (term match, no-match ⇒ nil, rune cap, determinism) and
`internal/sim/memory_test.go` the annotated-selector twins:
`TestSelectedWindowMatchesLegacy` proves `StripSelected` of the annotated
window equals the legacy selector byte-for-byte, and the serendipity-tail
flag is pinned for the assembler's drop accounting. Mind-side,
`internal/mind/context_test.go` proves the block assembler: a golden-identity
prompt, determinism, per-block telemetry sizes, the full drop-priority ladder
under a shrunken budget, the protected memory floor and the byte-identical
memories/serendipity accounting split, plan-echo content/guard
phrasing/clearing, the journal block, a planted-memory relevance check, and
an aggregate budget-fit sweep. Daemon-side,
`internal/daemon/context_replay_test.go` is the replay-determinism harness
(T013/T024, SC-004): `TestContextReplayByteIdentical` runs a real unpaused
loop, then proves both recovery paths (snapshot recovery and genesis replay)
rebuild the state byte-identically and `mind.AssembleUserPrompt` reproduces
the assembled decision prompt byte-for-byte from the recovered world;
`TestSageThrashWindowContextReplay` (env-guarded via
`PROMPTWORLD_WORLD01_DB`) reconstructs a historical agent's context at an
exact tick from a COPY of a legacy world.db via the daemon package's
`replayToTick` ([[daemon-lifecycle]]), asserting the assembled text surfaces
the documented reflex thrash — inspection of assembled text only, no model in
the loop.

## Connections

Part of the [[testing-strategy]] suite map (split out during the corpus-spec v2
restructure); see that note for the full layered test picture and links to
sibling suites.

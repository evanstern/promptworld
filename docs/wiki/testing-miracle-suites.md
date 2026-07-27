---
name: testing-miracle-suites
description: Miracle-cost derivation (sim.miracleCost mirrors the tool registry's price table) and the full miracle reducer/IPC round-trip coverage — move/remove/grant/time-snap arms, charge doctrine, and the rebaseTicks tick-taxonomy completeness guard. Split out of [[testing-strategy]].
kind: pattern
sources:
  - internal/sim/miracles_test.go
  - internal/ipc/ipc_test.go
verified_against: 22bb41c887ef6a34c55a77b9b989b299f4dc6857
---

# Miracle pricing & reducer suites

**Miracle cost derivation** (`internal/sim/miracles_test.go`, spec 021):
`TestMiracleCostDerivedFromTool` pins `sim.miracleCost` ≡
`tool.MiracleCostsByEvent()` — the sim-side enforcement table is a derivation of
the registry's single authoritative price source, not a mirror, so a price edit
cannot half-propagate ([[tool-registry]], [[guardian-miracles]]).

**Miracle reducer suite** (`internal/sim/miracles_test.go`, spec 016,
[[guardian-miracles]]): per-arm coverage for all four types — move (villager/
structure-whole/pile-merge, impassable/absent-source rejection), remove
(villager rejected, chest spill, pile destruction, terrain routing), grant
(happy path, over-cap whole-reject, unknown kind, dead villager, non-positive
qty, spear shape), and time-snap (forward-only, duration-preserving,
whole-day-no-drift, mints-no-charges-across-skipped-boundaries, while-paused);
plus charge doctrine (insufficient-charge rejection, gratis waives only the
charge, gratis is logged visibly), and `TestRebaseTaxonomyComplete` — the build
fails if a new tick-anchored `int64` field appears anywhere in the state tree
without a SHIFT/KEEP classification in `rebaseTicks`, so the taxonomy can never
silently drift from the state struct (spec 030 extended this to
`Belief.Reinforced`, the decay-curve anchor; spec 041 extends it again to
`PlaceFact.Seen`/`PeerSighting.Seen` (SHIFT) and `PlaceFact.Detail` (KEEP),
[[mental-maps]]/[[guardian-miracles]]; spec 042 extends it once more to
`Memory.Seq` (KEEP — the emitting event's store seq, an identity rather than a
clock value) and `Agent.SitVecTick` (KEEP — when the situation text was
rendered, [[memory-retrieval]]); spec 043 adds `Agent.NeedsAnchorTick` (SHIFT
— the trajectory-window edge anchor, elapsed-anchor shape, 0 = unset) and
`IntentRecord.Tick`/`IntentRecord.OutcomeTick` (KEEP — when an intent and its
outcome landed, self-history like `Memory.Tick`, [[decision-context]]), with
`TestSnapPreservesRemainingDurations` also proving the anchor's LEVELS
(`NeedsAnchor`, need values not ticks) ride a snap untouched while its tick
shifts; spec 044 extends it again with three KEEP
entries — `RunEnd.Tick` (when the run ended: history, the world never ticks
again), `DeathRecord.Tick` (the `NormViolation.Tick` shape), and
`MorgueEpilogue.Tick` (the `ChronicleEntry.Tick` shape); spec 046 adds three
more KEEP entries — `CurriculumPass.Tick` (when the pass was recorded,
history like `Memory.Tick`), `EvidenceRef.Tick` (an audit pointer at a
recorded event's tick, never a deadline), and `EvidenceRef.Seq` (the evidence
event's store seq, an identity like `Memory.Seq` —
[[curriculum-ladder]]); spec 062 adds one SHIFT entry, `Agent.LastMindIntentDone`
(the reflex PREP gate's yield-window anchor, only-non-zero, the
`Belief.Reinforced`/`NeedsAnchorTick` shape — [[reflex-policy]]); spec 054
adds one KEEP entry, `GuardianOrder.PlacedSeq` (the placement event's store
seq — an identity, like `Memory.Seq`, [[scenario-machinery]]); spec 063
adds three more KEEP entries — `GuardianReportCard.Tick` (when the card
landed, the `MorgueEpilogue.Tick` shape), `GuardianReportCard.Seq` (the card
event's own seq, an identity like `Memory.Seq`), and
`GuardianReportCard.Citations` (cited event seqs, identities into recorded
history like `EvidenceRef.Seq` — [[grounded-feedback]])). Byte-identity replay suites
(`TestMiracleReplayByteIdentity`, `TestMiracleSnapReplayByteIdentity`,
`TestMiracleGrantReplayByteIdentity`) prove each miracle type replays to the
same state hash as live application.

**IPC miracle round trips** (`internal/ipc/ipc_test.go`, spec 016): the
operator "miracle" command exercised over the real wire on a pure-sim world
(no LLM/guardian) — a move lands, spends a charge, and is visible in the next
state fetch; `--force`/`gratis` lands a miracle against an empty bank and
leaves it untouched at zero, while a non-forced attempt against the same
empty bank is refused; a `give_item` resolves the villager by name and the
grant is visible in the next state fetch; unknown kinds/names are refused
cleanly with the connection surviving.

## Connections

Part of the [[testing-strategy]] suite map (split out during the corpus-spec v2
restructure); see that note for the full layered test picture and links to
sibling suites.

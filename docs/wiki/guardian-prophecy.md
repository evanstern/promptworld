---
name: guardian-prophecy
description: Child of [[guardian-faith]] — the staked-vision prophecy channel (spec 085): the charge-priced prophesy tool, sim.Prophecy's entity discipline, the closed claim-kind predicate table (designation_fulfilled/structure_count/population_at_least/survives), the fulfil-before-fail verification sweep, and the declaration door's refusals. Load for how a vision becomes provably true or false.
kind: component
sources:
  - internal/sim/prophecy.go
  - internal/sim/executor.go
  - internal/guardian/prophecy.go
verified_against: 012f715f55d8d87317e601ad75686c599d277349
---

# Guardian prophecy — the staked vision

Child of [[guardian-faith]]: the prophecy channel — a charge-priced,
machine-verified claim about future world state, distinct from the parent's
faith-score economy and regen cadence that prophecy feeds into.

## How it works

`prophesy(targets, text, claim_kind, …, deadline_days)` spends **one
charge** (the `send_vision` price — the `prophecy.declared` arm validates
and decrements, so the spend is event-sourced) and records a
machine-checkable CLAIM with a deadline. **The verification rule**: a
vision is 'true' exactly when its recorded claim — declared before the
fact, from a CLOSED vocabulary — is satisfied by recorded world state
within its deadline, judged by pure (state, tick) predicates. Free text is
counsel, never graded; no model output participates.

`sim.Prophecy` clones the [[guardian-orders]] entity discipline: id
`pro-<tick>-<seq>` (guardian-minted, no RNG), one-way
`active → fulfilled | failed`, reducer-stamped `PlacedSeq`, cap **3
active** (`GuardianProphecyCap`), active + recent-32 prune, text ≤400
runes (`NudgeTextMax` — the registry `TextCapBytes` single source),
deadline 1..7 game days (the SHARED `GuardianOrderTTL*` bounds). There is
**NO cancel verb** (the word, once given, stands) and **no
all-targets-dead expiry** — the claim is a world fact, judged even if every
hearer died (companion memories simply skip the dead).

The claim predicate table (`prophecyClaimFulfilled`/`prophecyClaimFailed`,
shared verbatim by the sweep, the terminal arms' re-validation, and the
declaration door):

| Kind | Fulfil | Fail |
|---|---|---|
| `designation_fulfilled{designation_id}` | the designation's status is `fulfilled` | deadline ∧ ¬fulfil (a cancelled designation needs no special case) |
| `structure_count{structure_kind, min 1..64}` | count of kind ≥ min | deadline ∧ ¬fulfil |
| `population_at_least{min}` | living count ≥ min | deadline ∧ ¬fulfil |
| `survives{agent}` | deadline ∧ alive (the at-deadline kind) | agent dead — **fail-FAST**, no deadline wait |

The declaration door additionally refuses: a claim **already true** at
declaration (prophesying the past), a **duplicate of an active claim**
(normalized field equality — faith cannot be farmed by restating one
truth), kind-foreign claim params, and an empty bank. `prophecy.declared`
is the ONE whitelisted prophecy type (injected by the tool with per-target
`OriginOmen` dream-band companion memories, atomically — the vision-memory
firewall shape); `prophecy.fulfilled`/`prophecy.failed` are
executor-emitted by the verification sweep (`prophecyEvents`, placed after
the directive sweep) — **fulfil checked before fail**, exactly one
terminal ever lands, once (flips-non-active), each with per-living-target
`OriginReport` companion memories (word spreads, honestly secondhand — the
spec-030 provenance gate never launders it into witnessed). A claim
turning true AFTER `failed` latched mints nothing (one-way status). An
active prophecy's `DeadlineTick` SHIFTs across a time snap; `DeclaredTick`
and settled prophecies KEEP; `FaithState` carries no ticks
([[guardian-miracle-rebase-taxonomy]]).

## Connections

Parent [[guardian-faith]] covers the faith-score economy and regen cadence
prophecy outcomes feed via `faith.changed{prophecy_fulfilled/failed}`.
[[guardian-orders]] owns the entity discipline `sim.Prophecy` clones;
[[guardian-designations]] owns the `designation_fulfilled` claim's
referent; [[guardian-miracle-rebase-taxonomy]] classifies `DeadlineTick`'s
SHIFT treatment; [[guardian-turn-loop]] renders active prophecies into the
turn prompt.

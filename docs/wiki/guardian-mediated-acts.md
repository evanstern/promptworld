---
name: guardian-mediated-acts
description: The two mediated forms the guardian's turn can call — omens/visions (send_vision/send_omen, charge-priced perception memories, spec 041's optional place grant) and the work_miracle tool-call surface from the turn's side. Split from [[guardian]]; load for send_vision/send_omen mechanics or how a turn's work_miracle call reaches landMiracle. Full miracle mechanics live in [[guardian-miracles]].
kind: component
sources:
  - internal/guardian/turn.go
  - internal/guardian/toolcalls.go
  - internal/guardian/miracle_batch.go
  - internal/sim/guardian.go
verified_against: 74fe956813aa6be54e65156ae9bfcb91745cbb8d
---

# Guardian's mediated acts

Split from [[guardian]] (summary-style, corpus-spec v2) — the two
directly-implemented mediated forms a granted turn can call:
omens/visions, and the `work_miracle` tool-call surface (full mechanics in
[[guardian-miracles]]).

## How it works

**Influence: omens and visions** (spec 029, TASK-27): the two mediated forms that
replaced the retired `dream`/`omen` nudges. A **vision** (`send_vision`, `landVision`)
reaches exactly one living villager at ANY hour; an **omen** (`send_omen`, `landOmen`)
reaches one villager, a named comma-separated group, or the word `everyone` — but
only at NIGHT. Each landed act costs exactly ONE charge regardless of recipient
count, console-initiated or triggered. Validation (living target(s), non-empty text,
charges ≥ 1) downgrades failures to refusal-with-counsel — refusals are free, fed
back as a `rejected_gate` the model may repair in a later round. The `dream` form is
gone from the guardian's vocabulary; the spec-014 `OnRoster(RosterGuardian, "nudge_"+form)`
check is replaced by an explicit form switch in the reducer (`vision`/`omen`/`dream`,
with `dream` grandfathered replay-only — no live tool can produce a new one). The
400-byte text cap is still a registry read, re-pointed at `send_vision`'s entry
(`nudgeTextMax` in turn.go, `sim.NudgeTextMax` reducer-side, from the same tool so
truncation and enforcer never diverge).

Since spec 041 (FR-014), `send_vision` also carries an OPTIONAL place grant —
`place_kind`/`place_x`/`place_y`, all-or-none (`toolcalls.go`'s `parseReveal`
refuses a partial triple as a `rejected_gate` before anything lands). When
given, `landVision` (now taking a `*placeReveal` parameter) composes one
`metatron.place_revealed` event plus a companion `agent.memory_added`
("The vision showed you the <kind> at (x,y).", `SalDream`,
`Origin: sim.OriginOmen`) as `extra` events riding the SAME atomic
`landNudgeBatch` call as the vision's own nudge memory — the grant lands with
the vision or not at all. The kind enum is [[mental-maps]]'s closed
place-fact vocabulary; the reducer dry-run (a living target, a REAL place —
`groundFactPresent`) is the semantic authority, so the tool schema can only
over- or under-offer, never land a false fact. A vision without the place
arguments behaves exactly as before.

Both landers share `landNudgeBatch` — the text cap, the ONE atomic `InjectSocial`
batch, and the soul append, VERBATIM the pre-029 `landNudge` body (wrap, don't
rewrite): `metatron.nudged{form, targets, text}` (validating reducer spends the
charge and enforces the omen NIGHT gate at the door; `send_omen`'s day path never
reaches here) + one prefixed (`"You saw a vision: "` / `"You witnessed an omen: "`)
`agent.memory_added` per target at `SalDream` (8), each stamped `Origin: sim.OriginOmen`
(spec 030) — a direct-perception provenance class per `sim.DirectPerception` (same
standing as an own act or a witnessed event), which the villager interprets in
persona. `landVision`/`landOmen` differ only in target
resolution and the per-tool grant gate. The firewall is structural, not behavioral:
no code path exists from model output to any villager surface OR clock control
outside registered tools (sentinel-audited by `TestHandlerFirewallAudit`,
`guardian_test.go`, extended to the spec-029 surface, SC-007). A **daytime**
`send_omen` neither lands nor refuses — it defers to nightfall as a system-origin
standing order ([[guardian-orders]]).

**Miracles** (spec 016, [[guardian-miracles]]): the guardian's other charge-priced
mediated act, spent from the same bank, a declared loop tool: `work_miracle`
(`kind` ∈ `move`/`remove`/`give_item`/`time_snap`). The retired
`turnReply.Miracle` anonymous struct had **no gratis field** as its structural
guarantee; the replacement `miracleArgs` (`toolcalls.go`, the tool-call-parsed
mirror of the same flat surface) keeps that guarantee identically — nothing to
unmarshal `gratis` into, so a model-driven miracle can never waive its charge.
`landMiracle` resolves door-neutral `MiracleParams` (villager name → index,
day/`HH:MM` → tick via [[game-clock]]'s `ParseTimeOfDay`/`TickAt`) from an
`agentXY`/`alive` snapshot the absorb goroutine mirrors per batch (so the turn
worker never races the live replica). Since spec 091, a `move` call naming a
villager (`class="villager"` + a `villager` name) resolves that villager's LIVE
position from the SAME snapshot as the move's source — the surveyed `x`/`y`
become advisory — refusing before any charge on an unknown or dead name; a
bare-coordinate villager move and every structure/pile move are unaffected
(full mechanics and the raced-refusal wording in [[guardian-miracle-doors]]).
`landMiracle` then calls the shared `guardian.BuildMiracleBatch`
— the SAME builder the IPC `miracle` door uses — to compose the event and its
perception-memory companions (each stamped `Origin: sim.OriginOmen`, spec 030,
identically to a nudge's memories), and lands it through `InjectSocial`. A rejection at
the reducer dry-run becomes a `rejected_gate` the loop feeds back (rather than an
immediate reply-suffix refusal, though the wording is the same in-fiction
counsel); a landed miracle also appends a soul-file line. `landMiracle`'s
validation/batch/soul-append logic is likewise UNCHANGED from the pre-loop path —
only the input moved from `turnReply.Miracle` to `work_miracle`'s tool-call
arguments. Since spec 059 (US3), any turn whose granted roster offers
`work_miracle` (`hasWorkMiracle`) additionally carries a token-bounded
targeting digest in the user prompt — living villagers' positions/conditions/
carry headroom (spec 095) plus adjacent passable tiles, `turn.go`'s
`buildTargetingDigest` fed by the same `agentXY` mirror plus a parallel
`agentNeeds` mirror, introduced by
`tool.GuardianTargetingGuidance()` ([[tool-registry]]) — so a coordinate-
bearing miracle (`move`/`remove`) can aim at a tile the door will actually
accept (world-01 evidence: 3 of 4 miracle attempts door-rejected on invalid
coordinates). Prompt surface only; `landMiracle`'s reducer dry-run stays the
sole authority on whether a targeted coordinate lands — see
[[guardian-miracles]] for the digest's assembly and cost.

## Connections

[[guardian]] is the parent — this note covers the turn's side of both
mediated forms; [[guardian-miracles]] owns the miracle event types, cost
table, rebase taxonomy, and the two landing doors; [[guardian-orders]] owns
the daytime-omen deferral this note's `send_omen` night gate hands off to.
[[mental-maps]] owns the place-fact vocabulary `send_vision`'s optional
place grant draws on. [[sim-state-reducer]] validates and lands both
`metatron.nudged` and the four miracle event types. [[event-types]]
catalogs all of them.

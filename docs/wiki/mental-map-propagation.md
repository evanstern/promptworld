---
name: mental-map-propagation
description: Child of [[mental-maps]] — how place knowledge propagates OUT of perception into other channels: told facts exchanged in talk (social.place_told), the divine send_vision place-grant, and rendering the acting agent's own map into the decision prompt's known_places block. Load for talk/vision fact-grant mechanics or how knownPlaces is assembled for the model.
kind: component
sources:
  - internal/sim/executor.go
  - internal/sim/social.go
  - internal/guardian/toolcalls.go
  - internal/guardian/turn.go
  - internal/tool/registry.go
  - internal/mind/prompt.go
  - internal/mind/context.go
  - internal/mind/mind.go
verified_against: 657c770f87404b936a0587db1f6b00e81b9f0ee6
---

# Mental map propagation and rendering

Child of [[mental-maps]]: the three channels place knowledge travels through
besides direct perception ([[mental-map-perception]]) — peer-to-peer telling
in talk, the divine `send_vision` grant, and rendering the acting agent's own
map into the decision prompt.

**Directions exchanged in talk** (US5, research D5, [[social-fabric]]):
`tellablePlaces` (`internal/sim/executor.go`'s `talkEvents`) selects up to
`placeTellCap` (2) fresh facts per direction the other party lacks or holds
staler — freshest → nearest-to-listener → coordinate order — for EVERY
founded talk, hail-founded included, riding beside the rumor slot. One
`social.place_told{from, to, facts}` per direction, facts baked with `told`
provenance, the TELLER's `Seen` (secondhand is never fresher — staleness IS
the trust model), `Source` = the immediate teller. The reducer's `applySocial`
arm upserts into the RECEIVER's map only where absent or staler (a receiver's
own fresher knowledge never loses to secondhand); companion situated memories
on both sides ride the same batch at `salPlaceTold` = 3 (the talk band) —
"Told X about the fire at (x,y)." / "X told you of a fire at (x,y)."

**The divine reveal** (FR-014, [[guardian]]): `send_vision` carries an
OPTIONAL place-grant triple, `place_kind`/`place_x`/`place_y`, all riding
together — `internal/guardian/toolcalls.go`'s `parseReveal` refuses a partial
triple as a `rejected_gate` before anything lands. `place_kind`'s Enum is
`placeFactKinds` (`internal/tool/registry.go`) — [[mental-map-model]]'s closed
vocabulary hand-mirrored, since `tool` must not import `sim`; a drift there
can only over- or under-offer the model, never land a false fact, since the
reducer dry-run (`groundFactPresent`) is the semantic authority that the
place is real. `landVision` composes one `metatron.place_revealed` event plus
a companion `agent.memory_added` ("The vision showed you the fire at
(x,y).", `SalDream`, `Origin: sim.OriginOmen`) as extra events riding the SAME
atomic `landNudgeBatch` call as the vision's own nudge memory — the grant
lands with the vision or not at all. The reducer arm stamps `Seen` (the
landing tick) and `Detail` (ground truth at landing, `groundFactDetail`)
NORMATIVELY — the model-side emitter cannot know either, so it bakes only the
place identity.

**Rendering the known world** (US2, contracts §3, [[agent-mind]]): the
decision prompt retires the old blanket "Village: `<first six structures>`"
line and the bare-distance nearby-agent scan in favor of `knownPlaces`
(`internal/mind/prompt.go`). Since spec 043 that section is one named block
of the assembled decision context ([[decision-context]]):
`internal/mind/context.go`'s `renderKnownPlaces` — the `known_places`
contract block — wraps `knownPlaces` plus the peer-sighting "Nearby" line,
and is droppable (priority 5) under context-budget pressure, unlike the
survival blocks. The content is the
acting agent's OWN mental map, never `State.Structures`: landmark structures
(fire/shelter/oven/chest, and since spec 044 `grave` — a death site is
exactly the individually-named, narratively-weighted place the landmark set
exists for, never grouped with the count+nearest resource kinds)
individually with provenance flavor (witnessed
plain, told naming the teller, revealed naming the vision; a fire the agent
remembers as burned out says so), everything place-shaped grouped per kind
with count + nearest (walls/paths/resources — grouping bounds the size a
long run without dropping information), one orientation line toward the
nearest unexplored land (`FrontierDirection`, silent on a fully-explored map),
and an explicit empty state ("You know of no fires or shelters yet.") — the
model must always be able to tell "I know none" from silence. The nearby-agent
line walks the map's `Peers` in agent-index order, rendering a remembered
position (even the asleep flavor is only ever added when the peer is
verifiably still there) rather than a live one. `State.MapDims()` sizes the
bitmap read without the `State` ever serializing the map itself.

## Connections

Parent [[mental-maps]] summarizes this note and links every sibling child;
[[mental-map-model]] owns the `Facts`/closed-vocabulary type these channels
write into; [[mental-map-perception]] is the direct-witness channel this note
complements; [[social-fabric]] is where the place-telling sidecar rides,
beside rumors and gifts; [[guardian]] is the `send_vision` place grant's door;
[[decision-context]] assembles the `known_places` block this note renders
into; [[agent-mind]] is the planner that reads it and re-arms on a correction;
[[tui-client]] renders the propagation event types through the raw digest
feed; [[chronicle]] narrates the talk and vision events (not the perception
sweep's `agent.saw`, too chatty).

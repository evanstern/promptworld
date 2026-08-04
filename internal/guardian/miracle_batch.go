package guardian

import (
	"fmt"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
)

// The shared miracle batch-builder (spec 016 R6). Both doors — the operator's
// IPC `miracle` command and the guardian's landMiracle — compose their event batch
// HERE, so the miracle event and its FR-018 perception memories can never drift
// between the two channels. It is a package function (not a *Guardian method):
// the IPC server calls it without an guardian present, on pure-sim worlds.
//
// The builder only COMPOSES; it does not validate. Validation (entity presence,
// destination terrain, charge sufficiency) lives in the sim reducer arms and is
// enforced by the InjectSocial dry-run, so both doors reject identically and a
// recorded miracle always re-applies in replay. The one thing the builder reads
// from state is the deterministic resolution of perception-memory recipients,
// through the same sim helpers the reducer uses (VillagerAt / LivingAgents), so
// a tile-addressed move and its memory always name the same villager.

// MiracleParams is the door-neutral, already-resolved input to the builder:
// villager names are resolved to indices and day/HH:MM to a tick by the caller
// (contracts §1/§2), so the builder is a pure function of these fields.
type MiracleParams struct {
	ToTick   int64  // time_snap
	Agent    int    // give_item / take_item: resolved villager index
	Item     string // give_item / take_item: inventory kind
	Qty      int    // give_item / take_item
	Class    string // move / remove: villager|structure|pile|terrain
	X, Y     int    // move / remove: source tile
	ToX, ToY int    // move: destination tile
}

// Perception-memory templates (FR-018): fixed, deterministic, written for the
// villager's world (no player, no game, no outside voice), landed at SalDream —
// miracles are exactly as memorable as an otherworldly dream (research R7).
const (
	memMoved   = "An unseen hand lifted you and set you down in a strange place."
	memSnapped = "The light lurched across the sky; a great span of time passed in a single breath."
)

// BuildMiracleBatch composes the miracle event plus its perception memories for
// one landing. kind is the door vocabulary (time_snap|give_item|take_item|move|
// remove). Recipients per data-model.md: a moved villager, a granted villager,
// and (spec 116) a villager taken FROM each gain one memory; a time snap
// touches every living villager; structure/pile/terrain moves and removes touch
// none in v1.
func BuildMiracleBatch(s *sim.State, kind string, p MiracleParams, gratis bool) ([]store.Event, error) {
	var main store.Event
	var recipients []int
	var text string

	switch kind {
	case "time_snap":
		main = store.Event{Type: "guardian.time_snapped", Payload: mustJSON(sim.TimeSnappedPayload{
			ToTick: p.ToTick, Gratis: gratis})}
		recipients = s.LivingAgents()
		text = memSnapped
	case "give_item":
		main = store.Event{Type: "guardian.item_granted", Payload: mustJSON(sim.ItemGrantedPayload{
			Agent: sim.Ref(p.Agent), Kind: p.Item, Qty: p.Qty, Gratis: gratis})}
		recipients = []int{p.Agent}
		text = grantMemoryText(p.Qty, p.Item)
	case "take_item":
		// Spec 116 FR-013: a removal is a reach into a villager's pack, so it
		// carries the same situated memory a grant does — no pack mutation
		// lands in silence. Same recipient, same salience, same origin; only
		// the direction (and the text) differ.
		main = store.Event{Type: "guardian.item_taken", Payload: mustJSON(sim.ItemTakenPayload{
			Agent: sim.Ref(p.Agent), Kind: p.Item, Qty: p.Qty, Gratis: gratis})}
		recipients = []int{p.Agent}
		text = takeMemoryText(p.Qty, p.Item)
	case "move":
		main = store.Event{Type: "guardian.entity_moved", Payload: mustJSON(sim.EntityMovedPayload{
			Class: p.Class, X: p.X, Y: p.Y, ToX: p.ToX, ToY: p.ToY, Gratis: gratis})}
		if p.Class == "villager" {
			if idx := s.VillagerAt(p.X, p.Y); idx >= 0 {
				recipients = []int{idx}
				text = memMoved
			}
		}
	case "remove":
		main = store.Event{Type: "guardian.entity_removed", Payload: mustJSON(sim.EntityRemovedPayload{
			Class: p.Class, X: p.X, Y: p.Y, Gratis: gratis})}
		// No perception memory in v1 (no villager is directly affected).
	default:
		return nil, fmt.Errorf("unknown working kind %q", kind)
	}

	batch := []store.Event{main}
	for _, r := range recipients {
		batch = append(batch, store.Event{Type: "agent.memory_added", Payload: mustJSON(sim.MemoryAddedPayload{
			Agent: sim.Ref(r), Text: text, Salience: sim.SalDream, Subject: sim.Ref(-1), Origin: sim.OriginOmen})})
	}
	return batch, nil
}

// grantMemoryText renders the fixed grant memory. Deterministic; the raw
// inventory key is used verbatim (v1) so the text is a pure function of input.
func grantMemoryText(qty int, item string) string {
	return fmt.Sprintf("You found %d %s in your hands, as if set there by an unseen giver.", qty, item)
}

// takeMemoryText renders the fixed removal memory — grantMemoryText's mirror,
// same discipline: deterministic, first-person, written for the villager's own
// world (no player, no game, no outside voice), and naming BOTH what left the
// pack and where it went, because the goods are at their feet to pick back up.
func takeMemoryText(qty int, item string) string {
	return fmt.Sprintf("An unseen hand lifted %d %s from your pack and set it at your feet.", qty, item)
}

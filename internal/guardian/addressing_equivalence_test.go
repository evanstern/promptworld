package guardian

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/evanstern/promptworld/internal/bundle"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// TestAddressingMiracleByteIdentity is spec 082 SC-002 — the dogfood-move
// precedent extended to every class+tile form: a bundle effect addressing a
// villager, structure, pile, or terrain tile compiles to a MAIN event whose
// Type and Payload bytes equal BuildMiracleBatch's for the same class and
// tiles (Gratis:false both sides). Identity is pinned on the main event: the
// door adds a perception memory only for villager moves, and the bundle path
// deliberately never adds one (spec 082 Assumptions — narrate is the bundle
// author's channel), while structure/pile/terrain touch none in either door.
func TestAddressingMiracleByteIdentity(t *testing.T) {
	const seed = 42
	m := worldmap.Generate(seed, 64, 64)
	s := sim.NewState(seed, m)
	s.GuardianCharges = sim.GuardianChargeCap
	s.Structures = append(s.Structures, sim.Structure{Kind: "chest", X: 12, Y: 7})
	s.Piles = append(s.Piles, sim.Pile{X: 3, Y: 4, Wood: 2})
	ax, ay := s.Agents[0].X, s.Agents[0].Y

	cases := []struct {
		name   string
		effect string // one declarative effect entry
		event  string
		kind   string // BuildMiracleBatch door vocabulary
		params MiracleParams
	}{
		{"move villager@", fmt.Sprintf(`{"kind":"move_entity","target":"villager@%d,%d","to_x":5,"to_y":6}`, ax, ay),
			"guardian.entity_moved", "move", MiracleParams{Class: "villager", X: ax, Y: ay, ToX: 5, ToY: 6}},
		{"move structure@", `{"kind":"move_entity","target":"structure@12,7","to_x":4,"to_y":4}`,
			"guardian.entity_moved", "move", MiracleParams{Class: "structure", X: 12, Y: 7, ToX: 4, ToY: 4}},
		{"move pile@", `{"kind":"move_entity","target":"pile@3,4","to_x":6,"to_y":6}`,
			"guardian.entity_moved", "move", MiracleParams{Class: "pile", X: 3, Y: 4, ToX: 6, ToY: 6}},
		{"remove structure@", `{"kind":"remove_entity","target":"structure@12,7"}`,
			"guardian.entity_removed", "remove", MiracleParams{Class: "structure", X: 12, Y: 7}},
		{"remove pile@", `{"kind":"remove_entity","target":"pile@3,4"}`,
			"guardian.entity_removed", "remove", MiracleParams{Class: "pile", X: 3, Y: 4}},
		{"remove terrain@", `{"kind":"remove_entity","target":"terrain@9,2"}`,
			"guardian.entity_removed", "remove", MiracleParams{Class: "terrain", X: 9, Y: 2}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Bundle channel: the declarative compile pipeline.
			ts, err := bundle.ParseTemplates(json.RawMessage("[" + tc.effect + "]"))
			if err != nil {
				t.Fatalf("ParseTemplates: %v", err)
			}
			in := bundle.CompileInput{State: s, Declared: map[string]bool{tc.event: true}}
			effects, err := bundle.ExpandTemplates(ts, in)
			if err != nil {
				t.Fatalf("ExpandTemplates: %v", err)
			}
			batch, err := bundle.CompileEffects(effects, in)
			if err != nil {
				t.Fatalf("CompileEffects: %v", err)
			}
			if len(batch) != 1 {
				t.Fatalf("bundle batch = %d events, want 1", len(batch))
			}

			// Miracle door channel: the shared builder both doors compose through.
			doorBatch, err := BuildMiracleBatch(s, tc.kind, tc.params, false)
			if err != nil {
				t.Fatalf("BuildMiracleBatch: %v", err)
			}

			if batch[0].Type != doorBatch[0].Type {
				t.Errorf("type: bundle %q vs door %q", batch[0].Type, doorBatch[0].Type)
			}
			if !bytes.Equal(batch[0].Payload, doorBatch[0].Payload) {
				t.Errorf("payload:\n bundle: %s\n door:   %s", batch[0].Payload, doorBatch[0].Payload)
			}
		})
	}
}

// TestBundleAddressingTurnProbe proves the LIVE seam of spec 082: the turn
// assembly's invocation probe now mirrors structure/pile tiles and carries the
// static map (turn.go), so a class+tile bundle target resolves during a real
// guardian turn — not only against test-built states. Without the mirrors the
// compiler would reject every structure/pile address as unresolved and
// remove_entity would stay live-inert.
func TestBundleAddressingTurnProbe(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "It is done.")
	mt.SetBundles(bundleWorld(t, "addressing"))

	// Place a pile in the world AND the guardian's replica, then re-mirror —
	// the turn worker's probe is built from the mirrors, never the replica.
	pile := sim.Pile{X: 3, Y: 4, Wood: 2}
	inj.state.Piles = append(inj.state.Piles, pile)
	inj.state.GuardianCharges = 3
	mt.replica.Piles = append(mt.replica.Piles, pile)
	mt.mirrorState()

	mt.runLoop = actLoop(mt, "smite", `{"target":"pile@3,4"}`)
	if _, err := mt.Turn(context.Background(), "smite the pile"); err != nil {
		t.Fatalf("Turn: %v", err)
	}
	if inj.state.HasPileAt(3, 4) {
		t.Error("the tile-addressed pile removal did not land through the live turn path")
	}
}

package sim

// Family emission-drive suite (spec 086 T014, US1 AS-1..6, SC-001): fixture
// drives per family asserting named refs FROM LOG BYTES ALONE. The proof
// mechanic: each recorded payload is decoded from its stored bytes through
// the PayloadCatalog type (the dual-shape unmarshal is faithful — a name can
// only come from the bytes) and then validateRefs-checked, which demands the
// exact roster name on every in-roster ref and an empty name on every
// sentinel. A bare-int (legacy-shape) emission would decode to {id,""} and
// fail. No replica, no state — bytes + the AgentNames constant only.

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/store"
)

// assertNamedFromBytes proves every agent-bearing payload in events carries
// correct {id,name} refs, from log bytes alone. Returns the set of event
// types checked (coverage accounting for the caller).
func assertNamedFromBytes(t *testing.T, events []store.Event) map[string]bool {
	t.Helper()
	seen := map[string]bool{}
	for _, e := range events {
		zero, ok := PayloadCatalog[e.Type]
		if !ok {
			t.Errorf("event %q not in PayloadCatalog", e.Type)
			continue
		}
		pv := zero()
		if len(e.Payload) > 0 {
			if err := json.Unmarshal(e.Payload, pv); err != nil {
				t.Errorf("%s: payload does not decode through the catalog type: %v (%s)", e.Type, err, e.Payload)
				continue
			}
		}
		if err := validateRefs(pv); err != nil {
			t.Errorf("%s: emission not fully named (US1): %v — payload %s", e.Type, err, e.Payload)
			continue
		}
		seen[e.Type] = true
	}
	return seen
}

// TestEmissionDriveExecutorFamilies (US1 AS-1): a working morning of
// executor emissions — intents, harvests, needs, movement, memory, talk —
// every agent-bearing row named, verified from bytes.
func TestEmissionDriveExecutorFamilies(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	log := driveTicks(t, s, m, 8*3600, nil)
	seen := assertNamedFromBytes(t, log)
	for _, want := range []string{"agent.intent_set", "agent.moved", "agent.needs_changed", "agent.memory_added"} {
		if !seen[want] {
			t.Errorf("drive never emitted %s — the family fixture lost coverage", want)
		}
	}
	// Pin the literal wire shape once (the AC #1 format, from raw bytes).
	found := false
	for _, e := range log {
		if e.Type == "agent.intent_set" && strings.Contains(string(e.Payload), `"agent":{"id":`) {
			found = true
			break
		}
	}
	if !found {
		t.Error(`no agent.intent_set payload carries "agent":{"id":...} — the named object shape is missing from the wire`)
	}
}

// TestEmissionDriveDeathAndRunEnd (US1 AS-5 + the DeathRef mirror): staged
// deaths carry named refs; run.ended's death ledger mirrors carry names for
// agents who died earlier — death never blanks a name.
func TestEmissionDriveDeathAndRunEnd(t *testing.T) {
	const seed = 7
	m := testMap(seed)
	s := NewState(seed, m)
	for i := range s.Agents {
		s.Agents[i].Needs.Food = 0
		s.Agents[i].Needs.Health = 2
	}
	log := driveTicks(t, s, m, 600, nil)
	seen := assertNamedFromBytes(t, log)
	if !seen["agent.died"] {
		t.Fatal("staged collapse produced no agent.died")
	}
	if !seen["run.ended"] {
		t.Fatal("all-dead world produced no run.ended")
	}
	for _, e := range log {
		if e.Type == "run.ended" {
			var p RunEndedPayload
			if err := json.Unmarshal(e.Payload, &p); err != nil {
				t.Fatal(err)
			}
			if len(p.Deaths) == 0 {
				t.Fatal("run.ended carries no deaths")
			}
			for _, d := range p.Deaths {
				if d.Agent.Name != AgentNames[d.Agent.ID] {
					t.Errorf("posthumous death ref %+v lacks the roster name", d.Agent)
				}
			}
		}
	}
}

// TestEmissionDriveInjectedFamilies (US1 AS-3/AS-4): mind/guardian-shaped
// injected batches — memory, rumor (posthumous subject), nudge targets, the
// order_placed −1 sentinel, directive/prophecy mirrors, epilogue, chronicle
// — land through the door and read back fully named from stored bytes.
func TestEmissionDriveInjectedFamilies(t *testing.T) {
	h := newLadderHarness(t, func(s *State) {
		s.Night = true // omens land only at night
		s.GuardianCharges = 3
		s.Designations = append(s.Designations, Designation{
			ID: "dsg-1-0", Kind: "structure_site", StructureKind: "fire", X: 3, Y: 3,
			PlacedTick: 1, Status: "active"})
		s.Agents[3].Dead = true // posthumous references keep their names (AS-5)
	})

	died := Ref(3)
	batches := [][]store.Event{
		{{Type: "agent.memory_added", Payload: mustPayload(MemoryAddedPayload{
			Agent: Ref(0), Text: "remembered", Salience: 3, Subject: Ref(-1)})}},
		{{Type: "social.rumor_told", Payload: mustPayload(RumorToldPayload{
			From: Ref(0), To: Ref(1), Subject: died, Tone: -10, Text: "about the dead", Confidence: 80})}},
		{{Type: "metatron.nudged", Payload: mustPayload(GuardianNudgedPayload{
			Form: "omen", Targets: Refs([]int{0, 1, 2}), Text: "a sign"})}},
		{{Type: "metatron.order_placed", Payload: mustPayload(GuardianOrder{
			ID: "ord-1-0", Origin: "player", Condition: "if anyone sleeps", Action: "note it",
			EventTypes: []string{"agent.slept"}, Agent: -1, PlacedTick: 1, ExpiresTick: 90000, Status: "active",
		}.PlacedPayload())}},
		{{Type: "directive.issued", Payload: mustPayload(Directive{
			ID: "dir-1-0", DesignationID: "dsg-1-0", Targets: []int{0, 2}, Text: "see to the fire",
			IssuedTick: 1, ExpiresTick: 90000, Status: "active",
		}.IssuedPayload())}},
		{{Type: "prophecy.declared", Payload: mustPayload(Prophecy{
			ID: "pro-1-0", Targets: []int{0}, Text: "Ash will endure",
			Claim:        ProphecyClaim{Kind: ProphecySurvives, Agent: 0},
			DeclaredTick: 1, DeadlineTick: 90000, Status: "active",
		}.DeclaredPayload())}},
		{{Type: "metatron.item_granted", Payload: mustPayload(ItemGrantedPayload{
			Agent: Ref(2), Kind: "wood", Qty: 2, Gratis: true})}},
		{{Type: "morgue.epilogue", Payload: mustPayload(MorgueEpiloguePayload{
			Agent: Ref(3), Text: "they are mourned"})}},
		{{Type: "chronicle.entry", Payload: mustPayload(ChronicleEntryPayload{
			Day: 1, FromTick: 1, ToTick: 2, Text: "a day passed", Agents: Refs([]int{0, 1})})}},
	}
	for _, b := range batches {
		if err := h.loop.InjectSocial(b); err != nil {
			t.Fatalf("inject %s: %v", b[0].Type, err)
		}
	}

	evs, err := h.st.EventsSince(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	seen := assertNamedFromBytes(t, evs)
	for _, want := range []string{"agent.memory_added", "social.rumor_told", "metatron.nudged",
		"metatron.order_placed", "directive.issued", "prophecy.declared",
		"metatron.item_granted", "morgue.epilogue", "chronicle.entry"} {
		if !seen[want] {
			t.Errorf("injected family %s missing from the log", want)
		}
	}

	// Sentinel law from bytes (AS-4): −1 marshals fully, with an empty name.
	// Posthumous law from bytes (AS-5): the dead villager's ref stays named.
	// Survives-claim agent 0 carries a FULL ref (the index-0 edge).
	var sawSentinel, sawPosthumous, sawClaimRef bool
	for _, e := range evs {
		raw := string(e.Payload)
		if e.Type == "metatron.order_placed" && strings.Contains(raw, `"agent":{"id":-1,"name":""}`) {
			sawSentinel = true
		}
		if e.Type == "social.rumor_told" && strings.Contains(raw, `"subject":{"id":3,"name":"Rowan"}`) {
			sawPosthumous = true
		}
		if e.Type == "prophecy.declared" && strings.Contains(raw, `"agent":{"id":0,"name":"Ash"}`) {
			sawClaimRef = true
		}
	}
	if !sawSentinel {
		t.Error(`order_placed −1 sentinel did not marshal {"id":-1,"name":""} (AS-4)`)
	}
	if !sawPosthumous {
		t.Error("posthumous rumor subject lost its roster name (AS-5)")
	}
	if !sawClaimRef {
		t.Error("survives-claim agent 0 did not carry a full named ref (the index-0 edge)")
	}
}

package sim

// Emission-door rail tests (spec 086 T006/T017, FR-005, SC-002): mustPayload
// panics on an unnamed in-roster ref (the executor rail); the InjectSocial
// door refuses a batch carrying one BEFORE the dry-run (the injection rail);
// and Apply accepts unnamed (legacy) shapes forever — name validation never
// lives in an arm (the replay law, research R3).

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/store"
)

func TestMustPayloadPanicsOnUnnamedRef(t *testing.T) {
	cases := []struct {
		name string
		v    any
	}{
		{"unnamed in-roster ref", struct {
			Agent AgentRef `json:"agent"`
		}{Agent: AgentRef{ID: 2}}},
		{"wrong roster name", struct {
			Agent AgentRef `json:"agent"`
		}{Agent: AgentRef{ID: 2, Name: "Ash"}}},
		{"named sentinel", struct {
			Agent AgentRef `json:"agent"`
		}{Agent: AgentRef{ID: -1, Name: "Anyone"}}},
		{"unnamed ref in slice", struct {
			Targets []AgentRef `json:"targets"`
		}{Targets: []AgentRef{Ref(0), {ID: 3}}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("mustPayload accepted %s — the panic rail has no teeth", c.name)
				}
			}()
			mustPayload(c.v)
		})
	}

	// The rail accepts constructor-built refs and sentinels.
	ok := struct {
		Agent   AgentRef   `json:"agent"`
		Targets []AgentRef `json:"targets"`
	}{Agent: Ref(2), Targets: []AgentRef{Ref(-1), Ref(7)}}
	if got := mustPayload(ok); len(got) == 0 {
		t.Error("mustPayload rejected a valid payload")
	}
}

func TestInjectSocialRefusesUnnamedRef(t *testing.T) {
	if !censusMigrationComplete {
		t.Skip("door refusal is provable only once the census migrates ref-bearing types (T015/T017)")
	}
	h := newLadderHarness(t, nil)

	// A legacy-shape (bare index) chronicle.entry decodes to {id,""} refs —
	// live injection of an unnamed in-roster ref must refuse the batch at
	// the door, before the dry-run.
	legacy := store.Event{Type: "chronicle.entry",
		Payload: json.RawMessage(`{"day":1,"from_tick":1,"to_tick":2,"text":"x","agents":[2]}`)}
	err := h.loop.InjectSocial([]store.Event{legacy})
	if err == nil {
		t.Fatal("InjectSocial accepted an unnamed in-roster ref — the door rail has no teeth")
	}
	if !strings.Contains(err.Error(), "agent ref") {
		t.Fatalf("refusal should name the ref violation, got: %v", err)
	}

	// The named shape lands.
	named := store.Event{Type: "chronicle.entry",
		Payload: json.RawMessage(`{"day":1,"from_tick":1,"to_tick":2,"text":"x","agents":[{"id":2,"name":"Cedar"}]}`)}
	if err := h.loop.InjectSocial([]store.Event{named}); err != nil {
		t.Fatalf("InjectSocial refused a fully named batch: %v", err)
	}
}

// TestApplyAcceptsUnnamedShapes is the replay law (R3, US3 AS-5): a
// legacy-shape row folds through Apply identically and is NEVER rejected
// for a missing name — name validation lives only at the live-emission
// choke points, which replay never traverses.
func TestApplyAcceptsUnnamedShapes(t *testing.T) {
	m := testMap(7)
	s := NewState(7, m)
	legacy := store.Event{Seq: 1, Tick: 1, Type: "agent.memory_added",
		Payload: json.RawMessage(`{"agent":2,"text":"old row","salience":3,"subject":-1}`)}
	if err := s.Apply(legacy); err != nil {
		t.Fatalf("Apply rejected a legacy unnamed row: %v", err)
	}
	if n := len(s.Agents[2].Memories); n != 1 {
		t.Fatalf("legacy row did not fold: %d memories", n)
	}
	mem := s.Agents[2].Memories[0]
	if mem.Text != "old row" || mem.Subject != -1 {
		t.Fatalf("legacy fold wrong: %+v", mem)
	}
}

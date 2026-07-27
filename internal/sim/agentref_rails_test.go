package sim

// Emission-door rail tests (spec 086 T006/T017, FR-005, SC-002): mustPayload
// panics on an unnamed in-roster ref (the executor rail); the InjectSocial
// door refuses a batch carrying one BEFORE the dry-run (the injection rail);
// and Apply accepts unnamed (legacy) shapes forever — name validation never
// lives in an arm (the replay law, research R3).

import (
	"encoding/json"
	"reflect"
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

// TestDoorRefusesEveryUnnamedWhitelistedType (spec 086 T017, FR-005, US2
// AS-3): for EVERY injectSocialWhitelist type whose catalog payload carries
// agent refs, a legacy-shape (bare index) injection is refused at the door
// — before the dry-run, so no arm semantics participate. The probe payload
// is minimal: just the first ref-bearing field, set to an in-roster bare
// int, which the dual-shape unmarshal decodes as an unnamed ref.
func TestDoorRefusesEveryUnnamedWhitelistedType(t *testing.T) {
	h := newLadderHarness(t, nil)

	covered := 0
	for typ := range injectSocialWhitelist {
		zero, ok := PayloadCatalog[typ]
		if !ok {
			t.Fatalf("whitelisted type %q not in PayloadCatalog", typ)
		}
		rt := reflect.TypeOf(zero())
		for rt.Kind() == reflect.Pointer {
			rt = rt.Elem()
		}
		tag, list, found := firstRefField(rt)
		if !found {
			continue // no agent refs in this payload — nothing to refuse
		}
		var probe string
		if list {
			probe = `{"` + tag + `":[2]}`
		} else {
			probe = `{"` + tag + `":2}`
		}
		err := h.loop.InjectSocial([]store.Event{{Type: typ, Payload: json.RawMessage(probe)}})
		if err == nil || !strings.Contains(err.Error(), "agent ref") {
			t.Errorf("%s: unnamed in-roster ref not refused at the door (err=%v)", typ, err)
			continue
		}
		covered++
	}
	if covered < 10 {
		t.Fatalf("only %d whitelisted agent-bearing types probed — the coverage collapsed", covered)
	}
}

// firstRefField finds the first AgentRef-typed field (or slice/pointer of
// one) on a payload struct, returning its json tag and whether it is a list.
func firstRefField(rt reflect.Type) (tag string, list bool, ok bool) {
	if rt.Kind() != reflect.Struct {
		return "", false, false
	}
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		name := strings.Split(f.Tag.Get("json"), ",")[0]
		ft := f.Type
		switch {
		case ft == agentRefType:
			return name, false, true
		case ft.Kind() == reflect.Pointer && ft.Elem() == agentRefType:
			return name, false, true
		case ft.Kind() == reflect.Slice && ft.Elem() == agentRefType:
			return name, true, true
		}
	}
	return "", false, false
}

// TestMustPayloadPanicsPerFamily (T017's executor half): one concrete
// payload per emission family, each panicking on an unnamed in-roster ref —
// the panic-contract rail holds across the census, not just for one type.
func TestMustPayloadPanicsPerFamily(t *testing.T) {
	cases := []struct {
		family string
		v      any
	}{
		{"harvest", HarvestPayload{Agent: AgentRef{ID: 1}, X: 1, Y: 1}},
		{"social", RumorToldPayload{From: Ref(0), To: AgentRef{ID: 4}, Subject: Ref(2)}},
		{"governance", NormViolatedPayload{NormID: 1, Violator: Ref(0), Witnesses: []AgentRef{{ID: 5}}}},
		{"guardian-mirror", OrderPlacedPayload{ID: "ord-1-0", Agent: AgentRef{ID: 3}}},
		{"death-mirror", RunEndedPayload{Tick: 1, Deaths: []DeathRef{{Agent: AgentRef{ID: 6}, Tick: 1, Cause: "x"}}}},
		{"faith-additive", FaithChangedPayload{Delta: -6, Reason: FaithReasonVillagerDied, SourceID: "2", Agent: &AgentRef{ID: 2}}},
	}
	for _, c := range cases {
		t.Run(c.family, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: mustPayload accepted an unnamed in-roster ref", c.family)
				}
			}()
			mustPayload(c.v)
		})
	}
}

package sim

// AgentRef type tables (spec 086 T003, FR-001, SC-002): marshal shape and
// fixed field order (incl. a unicode-name fixture — a future roster must
// not break marshal determinism), the dual-shape unmarshal (bare int,
// object, []AgentRef over both, pointer), constructor sentinels, and the
// validateRefs accept/reject matrix.

import (
	"encoding/json"
	"testing"
)

func TestAgentRefMarshalShape(t *testing.T) {
	cases := []struct {
		in   AgentRef
		want string
	}{
		{AgentRef{ID: 2, Name: "Cedar"}, `{"id":2,"name":"Cedar"}`},
		{AgentRef{ID: -1}, `{"id":-1,"name":""}`}, // sentinel marshals fully — never omitted
		{AgentRef{ID: 0, Name: "Ash"}, `{"id":0,"name":"Ash"}`},
		// Unicode names marshal deterministically (encoding/json string
		// escaping is stable); the roster is ASCII today but AgentRef must
		// not care.
		{AgentRef{ID: 3, Name: "Åsa"}, `{"id":3,"name":"Åsa"}`},
	}
	for _, c := range cases {
		got, err := json.Marshal(c.in)
		if err != nil {
			t.Fatalf("marshal %+v: %v", c.in, err)
		}
		if string(got) != c.want {
			t.Errorf("marshal %+v = %s, want %s", c.in, got, c.want)
		}
	}
}

func TestAgentRefUnmarshalDualShape(t *testing.T) {
	cases := []struct {
		in   string
		want AgentRef
	}{
		{`2`, AgentRef{ID: 2, Name: ""}},                            // legacy bare index → empty name
		{`-1`, AgentRef{ID: -1, Name: ""}},                          // legacy sentinel
		{`{"id":2,"name":"Cedar"}`, AgentRef{ID: 2, Name: "Cedar"}}, // object form
		{`{"id":-1,"name":""}`, AgentRef{ID: -1, Name: ""}},         // named-shape sentinel
	}
	for _, c := range cases {
		var got AgentRef
		if err := json.Unmarshal([]byte(c.in), &got); err != nil {
			t.Fatalf("unmarshal %s: %v", c.in, err)
		}
		if got != c.want {
			t.Errorf("unmarshal %s = %+v, want %+v", c.in, got, c.want)
		}
	}

	// []AgentRef decodes both wire shapes element-wise, including mixed.
	var refs []AgentRef
	if err := json.Unmarshal([]byte(`[1,{"id":4,"name":"Fern"},-1]`), &refs); err != nil {
		t.Fatalf("unmarshal mixed slice: %v", err)
	}
	want := []AgentRef{{ID: 1}, {ID: 4, Name: "Fern"}, {ID: -1}}
	if len(refs) != len(want) {
		t.Fatalf("mixed slice length = %d, want %d", len(refs), len(want))
	}
	for i := range want {
		if refs[i] != want[i] {
			t.Errorf("mixed slice[%d] = %+v, want %+v", i, refs[i], want[i])
		}
	}

	// *AgentRef: null leaves nil; both shapes decode through the pointer.
	var p *AgentRef
	if err := json.Unmarshal([]byte(`null`), &p); err != nil || p != nil {
		t.Errorf("unmarshal null into *AgentRef: p=%v err=%v, want nil/nil", p, err)
	}
	if err := json.Unmarshal([]byte(`3`), &p); err != nil || p == nil || *p != (AgentRef{ID: 3}) {
		t.Errorf("unmarshal bare int into *AgentRef: p=%v err=%v", p, err)
	}
	p = nil
	if err := json.Unmarshal([]byte(`{"id":0,"name":"Ash"}`), &p); err != nil || p == nil || *p != (AgentRef{ID: 0, Name: "Ash"}) {
		t.Errorf("unmarshal object into *AgentRef: p=%v err=%v", p, err)
	}
}

func TestAgentRefConstructors(t *testing.T) {
	for i := 0; i < agentCount; i++ {
		if got := Ref(i); got.ID != i || got.Name != AgentNames[i] {
			t.Errorf("Ref(%d) = %+v, want {%d,%q}", i, got, i, AgentNames[i])
		}
	}
	for _, i := range []int{-1, agentCount, 99} {
		if got := Ref(i); got.ID != i || got.Name != "" {
			t.Errorf("Ref(%d) = %+v, want {%d,\"\"} (sentinel law)", i, got, i)
		}
	}
	if Refs(nil) != nil {
		t.Error("Refs(nil) should be nil")
	}
	got := Refs([]int{0, -1, 2})
	want := []AgentRef{{0, "Ash"}, {-1, ""}, {2, "Cedar"}}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Refs[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestValidateRefsMatrix(t *testing.T) {
	type nested struct {
		Inner AgentRef
	}
	type payload struct {
		A     AgentRef
		List  []AgentRef
		Ptr   *AgentRef
		Deep  nested
		Raw   json.RawMessage // opaque bytes — never walked
		Plain int
	}

	ok := payload{
		A:    Ref(0),
		List: []AgentRef{Ref(1), Ref(-1)},
		Ptr:  &AgentRef{ID: 2, Name: "Cedar"},
		Deep: nested{Inner: Ref(7)},
		Raw:  json.RawMessage(`{"agent":3}`), // a bare int inside raw bytes is fine
	}
	if err := validateRefs(ok); err != nil {
		t.Errorf("valid payload rejected: %v", err)
	}
	if err := validateRefs(&ok); err != nil {
		t.Errorf("valid payload (pointer) rejected: %v", err)
	}
	if err := validateRefs(nil); err != nil {
		t.Errorf("nil payload rejected: %v", err)
	}

	reject := []struct {
		name string
		v    any
	}{
		{"unnamed in-roster ref", payload{A: AgentRef{ID: 2}}},
		{"wrong roster name", payload{A: AgentRef{ID: 2, Name: "Ash"}}},
		{"named sentinel", payload{A: Ref(0), List: []AgentRef{{ID: -1, Name: "Anyone"}}}},
		{"unnamed nested ref", payload{A: Ref(0), Deep: nested{Inner: AgentRef{ID: 5}}}},
		{"unnamed ref behind pointer", payload{A: Ref(0), Ptr: &AgentRef{ID: 1}}},
	}
	for _, c := range reject {
		if err := validateRefs(c.v); err == nil {
			t.Errorf("%s: accepted, want rejection", c.name)
		}
	}
}

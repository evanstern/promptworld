package bundle

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// testState builds a sim.State with the named living villagers at given tiles.
func testState(names []string, xs, ys []int) *sim.State {
	agents := make([]sim.Agent, len(names))
	for i := range names {
		agents[i] = sim.Agent{Name: names[i], X: xs[i], Y: ys[i]}
	}
	return &sim.State{Agents: agents}
}

func decl(types ...string) map[string]bool {
	m := make(map[string]bool, len(types))
	for _, t := range types {
		m[t] = true
	}
	return m
}

// compileRaw runs the full declarative pipeline: parse templates, expand with
// the given input, and compile to events.
func compileRaw(t *testing.T, raw string, in CompileInput) ([]store.Event, error) {
	t.Helper()
	ts, err := ParseTemplates(json.RawMessage(raw))
	if err != nil {
		return nil, err
	}
	effects, err := ExpandTemplates(ts, in)
	if err != nil {
		return nil, err
	}
	return CompileEffects(effects, in)
}

// TestCompileEachKind: every effect kind compiles to the correct event type and
// a payload byte-identical to the canonical sim.* struct.
func TestCompileEachKind(t *testing.T) {
	st := testState([]string{"Alice", "Bob"}, []int{1, 3}, []int{2, 4})

	t.Run("move_entity", func(t *testing.T) {
		in := CompileInput{State: st, Args: map[string]string{"target": "Alice"}, Declared: decl("guardian.entity_moved")}
		evs, err := compileRaw(t, `[{"kind":"move_entity","target":"{args.target}","to_x":5,"to_y":6}]`, in)
		if err != nil {
			t.Fatal(err)
		}
		if len(evs) != 1 || evs[0].Type != "guardian.entity_moved" {
			t.Fatalf("events = %+v", evs)
		}
		var p sim.EntityMovedPayload
		mustUnmarshal(t, evs[0].Payload, &p)
		want := sim.EntityMovedPayload{Class: "villager", X: 1, Y: 2, ToX: 5, ToY: 6}
		if p != want {
			t.Errorf("payload = %+v, want %+v", p, want)
		}
	})

	t.Run("remove_entity", func(t *testing.T) {
		// Spec 082 (US1 AC5): every villager-designating form is rejected at
		// compile time — the reducer doctrine ("a villager can never be
		// removed"), mirrored for a better authoring error. The accepting
		// forms (structure@/pile@/terrain@) are covered by
		// TestTargetAddressingCompiles.
		in := CompileInput{State: st, Args: map[string]string{"target": "Bob"}, Declared: decl("guardian.entity_removed")}
		_, err := compileRaw(t, `[{"kind":"remove_entity","target":"Bob"}]`, in)
		if err == nil || !strings.Contains(err.Error(), "a villager can never be removed") {
			t.Errorf("err = %v, want villager-removal doctrine error", err)
		}
	})

	t.Run("grant_item", func(t *testing.T) {
		in := CompileInput{State: st, Declared: decl("guardian.item_granted")}
		evs, err := compileRaw(t, `[{"kind":"grant_item","target":"Alice","item":"bread","qty":2}]`, in)
		if err != nil {
			t.Fatal(err)
		}
		var p sim.ItemGrantedPayload
		mustUnmarshal(t, evs[0].Payload, &p)
		want := sim.ItemGrantedPayload{Agent: sim.Ref(0), Kind: "bread", Qty: 2}
		if evs[0].Type != "guardian.item_granted" || p != want {
			t.Errorf("type/payload = %s/%+v", evs[0].Type, p)
		}
	})

	t.Run("snap_time", func(t *testing.T) {
		in := CompileInput{State: st, Declared: decl("guardian.time_snapped")}
		evs, err := compileRaw(t, `[{"kind":"snap_time","to_tick":9000}]`, in)
		if err != nil {
			t.Fatal(err)
		}
		var p sim.TimeSnappedPayload
		mustUnmarshal(t, evs[0].Payload, &p)
		if evs[0].Type != "guardian.time_snapped" || p.ToTick != 9000 || p.Gratis {
			t.Errorf("type/payload = %s/%+v", evs[0].Type, p)
		}
	})

	t.Run("narrate matches miracle memory shape", func(t *testing.T) {
		in := CompileInput{State: st, Declared: decl("agent.memory_added")}
		evs, err := compileRaw(t, `[{"kind":"narrate","text":"a poof of smoke","recipients":["Alice"]}]`, in)
		if err != nil {
			t.Fatal(err)
		}
		var p sim.MemoryAddedPayload
		mustUnmarshal(t, evs[0].Payload, &p)
		if p.Agent.ID != 0 || p.Text != "a poof of smoke" || p.Salience != sim.SalDream || p.Subject.ID != -1 || p.Origin != sim.OriginOmen {
			t.Errorf("payload = %+v", p)
		}
	})
}

// addressState builds a small world for the spec-082 class+tile forms: a
// 24×24 map, villagers Alice (1,2) and Bob (5,5) — Cara SHARES Bob's tile so
// the villager@ form's first-by-agent-index choice is observable — a structure
// at (12,7), and a pile at (3,4).
func addressState(t *testing.T) *sim.State {
	t.Helper()
	st := testState([]string{"Alice", "Bob", "Cara"}, []int{1, 5, 5}, []int{2, 5, 5})
	st.Structures = []sim.Structure{{Kind: "chest", X: 12, Y: 7}}
	st.Piles = []sim.Pile{{X: 3, Y: 4, Wood: 2}}
	st.SetMap(worldmap.Generate(1, 24, 24))
	return st
}

// TestTargetAddressingCompiles: every ✅ cell of the data-model.md §4 form
// matrix compiles to the exact sim.* payload the miracle door would land
// (spec 082 US1/US2; FR-002/003/004). The cross-door byte-identity pin against
// guardian.BuildMiracleBatch lives in internal/guardian (import direction).
func TestTargetAddressingCompiles(t *testing.T) {
	st := addressState(t)
	cases := []struct {
		name    string
		raw     string
		declare string
		wantTyp string
		want    any
	}{
		{"move structure@", `[{"kind":"move_entity","target":"structure@12,7","to_x":4,"to_y":4}]`,
			"guardian.entity_moved", "guardian.entity_moved",
			sim.EntityMovedPayload{Class: "structure", X: 12, Y: 7, ToX: 4, ToY: 4}},
		{"move pile@", `[{"kind":"move_entity","target":"pile@3,4","to_x":6,"to_y":6}]`,
			"guardian.entity_moved", "guardian.entity_moved",
			sim.EntityMovedPayload{Class: "pile", X: 3, Y: 4, ToX: 6, ToY: 6}},
		{"move villager@ resolves first-by-index", `[{"kind":"move_entity","target":"villager@5,5","to_x":6,"to_y":6}]`,
			"guardian.entity_moved", "guardian.entity_moved",
			sim.EntityMovedPayload{Class: "villager", X: 5, Y: 5, ToX: 6, ToY: 6}},
		{"move villager: typed name", `[{"kind":"move_entity","target":"villager:Alice","to_x":6,"to_y":6}]`,
			"guardian.entity_moved", "guardian.entity_moved",
			sim.EntityMovedPayload{Class: "villager", X: 1, Y: 2, ToX: 6, ToY: 6}},
		{"remove structure@", `[{"kind":"remove_entity","target":"structure@12,7"}]`,
			"guardian.entity_removed", "guardian.entity_removed",
			sim.EntityRemovedPayload{Class: "structure", X: 12, Y: 7}},
		{"remove pile@", `[{"kind":"remove_entity","target":"pile@3,4"}]`,
			"guardian.entity_removed", "guardian.entity_removed",
			sim.EntityRemovedPayload{Class: "pile", X: 3, Y: 4}},
		{"remove terrain@ (bounds-only at compile)", `[{"kind":"remove_entity","target":"terrain@9,2"}]`,
			"guardian.entity_removed", "guardian.entity_removed",
			sim.EntityRemovedPayload{Class: "terrain", X: 9, Y: 2}},
		{"grant villager@", `[{"kind":"grant_item","target":"villager@1,2","item":"wood","qty":3}]`,
			"guardian.item_granted", "guardian.item_granted",
			sim.ItemGrantedPayload{Agent: sim.Ref(0), Kind: "wood", Qty: 3}},
		{"grant villager: typed name", `[{"kind":"grant_item","target":"villager:bob","item":"wood","qty":3}]`,
			"guardian.item_granted", "guardian.item_granted",
			sim.ItemGrantedPayload{Agent: sim.Ref(1), Kind: "wood", Qty: 3}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := CompileInput{State: st, Declared: decl(tc.declare)}
			evs, err := compileRaw(t, tc.raw, in)
			if err != nil {
				t.Fatal(err)
			}
			if len(evs) != 1 || evs[0].Type != tc.wantTyp {
				t.Fatalf("events = %+v", evs)
			}
			wantBytes, err := json.Marshal(tc.want)
			if err != nil {
				t.Fatal(err)
			}
			if string(evs[0].Payload) != string(wantBytes) {
				t.Errorf("payload = %s, want %s", evs[0].Payload, wantBytes)
			}
		})
	}

	// villager@ shared tile: the payload above proves the tile; pin the INDEX
	// choice too (Bob, index 1, not Cara, index 2 — first living by agent
	// index, the VillagerAt discipline) through grant_item, whose payload
	// carries the index.
	in := CompileInput{State: st, Declared: decl("guardian.item_granted")}
	evs, err := compileRaw(t, `[{"kind":"grant_item","target":"villager@5,5","item":"wood","qty":1}]`, in)
	if err != nil {
		t.Fatal(err)
	}
	var p sim.ItemGrantedPayload
	mustUnmarshal(t, evs[0].Payload, &p)
	if p.Agent.ID != 1 {
		t.Errorf("villager@5,5 resolved to index %d, want 1 (first living by agent index)", p.Agent.ID)
	}
}

// TestTargetErrorTaxonomy is SC-005's table: every compile-time taxonomy class
// (data-model.md §5) produces a T5 error naming the effect index, the field,
// and the offending address text, and the whole invocation is rejected. The
// `class` kind is unreachable through bundle targets by construction (the
// reserved-prefix set and class set are one table — data-model.md §5); its
// message shape is pinned white-box in internal/target.
func TestTargetErrorTaxonomy(t *testing.T) {
	st := addressState(t)
	cases := []struct {
		name string
		raw  string
		want []string // substrings: effect index+field/kind, offending address, taxonomy message
	}{
		// syntax — reserved prefix, malformed remainder; never a name fallback.
		{"syntax malformed locus", `[{"kind":"narrate","text":"x","recipients":"all_living"},{"kind":"move_entity","target":"structure@","to_x":1,"to_y":1}]`,
			[]string{`effect 1 field "target"`, `"structure@"`, "not a valid address"}},
		{"syntax negative coordinate", `[{"kind":"remove_entity","target":"pile@-1,4"}]`,
			[]string{`effect 0 field "target"`, `"pile@-1,4"`, "non-negative integer"}},
		{"syntax substituted junk", `[{"kind":"remove_entity","target":"structure@{args.x},{args.y}"}]`,
			[]string{`effect 0 field "target"`, `"structure@north,7"`, "not a valid address"}},
		// form — ❌ cells of the §4 matrix.
		{"form move terrain", `[{"kind":"move_entity","target":"terrain@9,2","to_x":1,"to_y":1}]`,
			[]string{"effect 0 (move_entity)", `"terrain@9,2"`, "terrain cannot be moved"}},
		{"form remove villager bare name", `[{"kind":"remove_entity","target":"Alice"}]`,
			[]string{"effect 0 (remove_entity)", `"Alice"`, "a villager can never be removed"}},
		{"form remove villager typed name", `[{"kind":"remove_entity","target":"villager:Alice"}]`,
			[]string{"effect 0 (remove_entity)", `"villager:Alice"`, "a villager can never be removed"}},
		{"form remove villager tile", `[{"kind":"remove_entity","target":"villager@5,5"}]`,
			[]string{"effect 0 (remove_entity)", `"villager@5,5"`, "a villager can never be removed"}},
		{"form grant structure", `[{"kind":"grant_item","target":"structure@12,7","item":"wood","qty":1}]`,
			[]string{"effect 0 (grant_item)", `"structure@12,7"`, "can only target a villager"}},
		{"form rect reserved (move)", `[{"kind":"move_entity","target":"structure@1,1..3,3","to_x":1,"to_y":1}]`,
			[]string{`effect 0 field "target"`, `"structure@1,1..3,3"`, "reserved for designation consumers (TASK-157)"}},
		{"form line reserved (remove)", `[{"kind":"remove_entity","target":"structure@1,1->1,5"}]`,
			[]string{`effect 0 field "target"`, `"structure@1,1->1,5"`, "reserved for designation consumers (TASK-157)"}},
		{"form diagonal line", `[{"kind":"remove_entity","target":"structure@1,1->2,5"}]`,
			[]string{`effect 0 field "target"`, `"structure@1,1->2,5"`, "diagonal"}},
		// bounds — outside the 24×24 fixture map.
		{"bounds x", `[{"kind":"remove_entity","target":"terrain@99,2"}]`,
			[]string{`effect 0 field "target"`, `"terrain@99,2"`, "(99,2) is outside the 24×24 world"}},
		{"bounds villager tile", `[{"kind":"grant_item","target":"villager@1,99","item":"wood","qty":1}]`,
			[]string{`effect 0 field "target"`, `"villager@1,99"`, "outside the 24×24 world"}},
		// unresolved — in-bounds, allowed form, nothing binds.
		{"unresolved structure", `[{"kind":"move_entity","target":"structure@9,9","to_x":1,"to_y":1}]`,
			[]string{`effect 0 field "target"`, `"structure@9,9"`, "no structure at (9,9)"}},
		{"unresolved pile", `[{"kind":"remove_entity","target":"pile@9,9"}]`,
			[]string{`effect 0 field "target"`, `"pile@9,9"`, "no pile at (9,9)"}},
		{"unresolved villager tile", `[{"kind":"move_entity","target":"villager@9,9","to_x":1,"to_y":1}]`,
			[]string{`effect 0 field "target"`, `"villager@9,9"`, "no living villager at (9,9)"}},
		{"unresolved villager name (v1 shape)", `[{"kind":"move_entity","target":"boulder@3,4","to_x":1,"to_y":1}]`,
			[]string{"effect 0", `"boulder@3,4"`, "no living villager named"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := CompileInput{
				State:    st,
				Args:     map[string]string{"x": "north", "y": "7"},
				Declared: decl("guardian.entity_moved", "guardian.entity_removed", "guardian.item_granted", "agent.memory_added"),
			}
			evs, err := compileRaw(t, tc.raw, in)
			if err == nil {
				t.Fatalf("compiled %d events, want rejection", len(evs))
			}
			for _, want := range tc.want {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("err = %q, want substring %q", err, want)
				}
			}
			if evs != nil {
				t.Error("a rejected invocation returned events (must be whole-invocation atomic)")
			}
		})
	}
}

// TestTemplateSubstitution: {args.x} and {invoker} resolve; unknown placeholders
// error and name the field.
func TestTemplateSubstitution(t *testing.T) {
	st := testState([]string{"Alice"}, []int{0}, []int{0})
	in := CompileInput{State: st, Invoker: "the guardian", Args: map[string]string{"target": "Alice", "who": "Alice"}, Declared: decl("agent.memory_added")}
	evs, err := compileRaw(t, `[{"kind":"narrate","text":"{invoker} blessed {args.who}","recipients":"target"}]`, in)
	if err != nil {
		t.Fatal(err)
	}
	var p sim.MemoryAddedPayload
	mustUnmarshal(t, evs[0].Payload, &p)
	if p.Text != "the guardian blessed Alice" {
		t.Errorf("text = %q", p.Text)
	}

	_, err = compileRaw(t, `[{"kind":"narrate","text":"{args.missing}","recipients":"all_living"}]`, in)
	if err == nil || !strings.Contains(err.Error(), "unresolved placeholder {args.missing}") {
		t.Errorf("err = %v, want unresolved placeholder", err)
	}
}

// TestRecipientExpansion: target / all_living / named selectors resolve to the
// right living-villager indices.
func TestRecipientExpansion(t *testing.T) {
	st := testState([]string{"Alice", "Bob", "Cara"}, []int{0, 1, 2}, []int{0, 1, 2})

	t.Run("all_living", func(t *testing.T) {
		in := CompileInput{State: st, Declared: decl("agent.memory_added")}
		evs, err := compileRaw(t, `[{"kind":"narrate","text":"lurch","recipients":"all_living"}]`, in)
		if err != nil {
			t.Fatal(err)
		}
		if got := recipients(t, evs); fmt.Sprint(got) != "[0 1 2]" {
			t.Errorf("recipients = %v", got)
		}
	})

	t.Run("target", func(t *testing.T) {
		in := CompileInput{State: st, Args: map[string]string{"target": "Bob"}, Declared: decl("agent.memory_added")}
		evs, err := compileRaw(t, `[{"kind":"narrate","text":"hi","recipients":"target"}]`, in)
		if err != nil {
			t.Fatal(err)
		}
		if got := recipients(t, evs); fmt.Sprint(got) != "[1]" {
			t.Errorf("recipients = %v", got)
		}
	})

	t.Run("named list", func(t *testing.T) {
		in := CompileInput{State: st, Declared: decl("agent.memory_added")}
		evs, err := compileRaw(t, `[{"kind":"narrate","text":"hi","recipients":["Cara","Alice"]}]`, in)
		if err != nil {
			t.Fatal(err)
		}
		if got := recipients(t, evs); fmt.Sprint(got) != "[2 0]" {
			t.Errorf("recipients = %v", got)
		}
	})

	t.Run("named unknown errors", func(t *testing.T) {
		in := CompileInput{State: st, Declared: decl("agent.memory_added")}
		_, err := compileRaw(t, `[{"kind":"narrate","text":"hi","recipients":["Nobody"]}]`, in)
		if err == nil || !strings.Contains(err.Error(), `recipient "Nobody" is not a living villager`) {
			t.Errorf("err = %v", err)
		}
	})
}

// TestEmptyNarrationAllowed: all_living in a world with no living villagers
// yields an empty batch, not an error (a valid no-op narration).
func TestEmptyNarrationAllowed(t *testing.T) {
	st := &sim.State{Agents: []sim.Agent{{Name: "Ghost", Dead: true}}}
	in := CompileInput{State: st, Declared: decl("agent.memory_added")}
	evs, err := compileRaw(t, `[{"kind":"narrate","text":"silence","recipients":"all_living"}]`, in)
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 0 {
		t.Errorf("events = %d, want 0", len(evs))
	}
}

// TestCaps: batch >32 events and narrate text >500 bytes are rejected.
func TestCaps(t *testing.T) {
	names := make([]string, 33)
	xs := make([]int, 33)
	ys := make([]int, 33)
	for i := range names {
		names[i] = fmt.Sprintf("v%02d", i)
	}
	st := testState(names, xs, ys)

	t.Run("batch cap", func(t *testing.T) {
		in := CompileInput{State: st, Declared: decl("agent.memory_added")}
		_, err := compileRaw(t, `[{"kind":"narrate","text":"x","recipients":"all_living"}]`, in)
		if err == nil || !strings.Contains(err.Error(), "exceeds 32 events") {
			t.Errorf("err = %v, want batch-cap error", err)
		}
	})

	t.Run("text cap", func(t *testing.T) {
		small := testState([]string{"Alice"}, []int{0}, []int{0})
		in := CompileInput{State: small, Declared: decl("agent.memory_added")}
		long := strings.Repeat("a", 501)
		_, err := compileRaw(t, fmt.Sprintf(`[{"kind":"narrate","text":%q,"recipients":"all_living"}]`, long), in)
		if err == nil || !strings.Contains(err.Error(), "501 bytes (max 500)") {
			t.Errorf("err = %v, want text-cap error", err)
		}
	})
}

// TestUndeclaredEventRejected: a produced event type outside the declared set is
// rejected (the invocation-time subset gate).
func TestUndeclaredEventRejected(t *testing.T) {
	st := testState([]string{"Alice"}, []int{1}, []int{2})
	in := CompileInput{State: st, Args: map[string]string{"target": "Alice"}, Declared: decl("agent.memory_added")}
	_, err := compileRaw(t, `[{"kind":"move_entity","target":"Alice","to_x":5,"to_y":6}]`, in)
	if err == nil || !strings.Contains(err.Error(), `"guardian.entity_moved" is not in the tool's declared events`) {
		t.Errorf("err = %v, want undeclared-event error", err)
	}
}

// TestNumericFieldRules: floats are rejected in numeric fields and qty is
// range-checked.
func TestNumericFieldRules(t *testing.T) {
	st := testState([]string{"Alice"}, []int{0}, []int{0})
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{"float literal", `[{"kind":"move_entity","target":"Alice","to_x":1.5,"to_y":0}]`, "must be an integer, not a float"},
		{"float string", `[{"kind":"move_entity","target":"Alice","to_x":"1.5","to_y":0}]`, "is not an integer"},
		{"qty zero", `[{"kind":"grant_item","target":"Alice","item":"bread","qty":0}]`, "out of range 1–99"},
		{"qty over", `[{"kind":"grant_item","target":"Alice","item":"bread","qty":100}]`, "out of range 1–99"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := CompileInput{State: st, Declared: decl("guardian.entity_moved", "guardian.item_granted")}
			_, err := compileRaw(t, tc.raw, in)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("err = %v, want %q", err, tc.want)
			}
		})
	}
}

// TestUnknownEffectKind: an unknown effect kind is rejected at template parse.
func TestUnknownEffectKind(t *testing.T) {
	_, err := ParseTemplates(json.RawMessage(`[{"kind":"heal","target":"Alice"}]`))
	if err == nil || !strings.Contains(err.Error(), `unknown kind "heal"`) {
		t.Errorf("err = %v", err)
	}
}

func mustUnmarshal(t *testing.T, data []byte, v any) {
	t.Helper()
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
}

func recipients(t *testing.T, evs []store.Event) []int {
	t.Helper()
	out := make([]int, len(evs))
	for i, e := range evs {
		var p sim.MemoryAddedPayload
		mustUnmarshal(t, e.Payload, &p)
		out[i] = p.Agent.ID
	}
	return out
}

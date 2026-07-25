package bundle

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
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
		in := CompileInput{State: st, Args: map[string]string{"target": "Alice"}, Declared: decl("metatron.entity_moved")}
		evs, err := compileRaw(t, `[{"kind":"move_entity","target":"{args.target}","to_x":5,"to_y":6}]`, in)
		if err != nil {
			t.Fatal(err)
		}
		if len(evs) != 1 || evs[0].Type != "metatron.entity_moved" {
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
		in := CompileInput{State: st, Args: map[string]string{"target": "Bob"}, Declared: decl("metatron.entity_removed")}
		evs, err := compileRaw(t, `[{"kind":"remove_entity","target":"Bob"}]`, in)
		if err != nil {
			t.Fatal(err)
		}
		var p sim.EntityRemovedPayload
		mustUnmarshal(t, evs[0].Payload, &p)
		want := sim.EntityRemovedPayload{Class: "villager", X: 3, Y: 4}
		if evs[0].Type != "metatron.entity_removed" || p != want {
			t.Errorf("type/payload = %s/%+v", evs[0].Type, p)
		}
	})

	t.Run("grant_item", func(t *testing.T) {
		in := CompileInput{State: st, Declared: decl("metatron.item_granted")}
		evs, err := compileRaw(t, `[{"kind":"grant_item","target":"Alice","item":"bread","qty":2}]`, in)
		if err != nil {
			t.Fatal(err)
		}
		var p sim.ItemGrantedPayload
		mustUnmarshal(t, evs[0].Payload, &p)
		want := sim.ItemGrantedPayload{Agent: 0, Kind: "bread", Qty: 2}
		if evs[0].Type != "metatron.item_granted" || p != want {
			t.Errorf("type/payload = %s/%+v", evs[0].Type, p)
		}
	})

	t.Run("snap_time", func(t *testing.T) {
		in := CompileInput{State: st, Declared: decl("metatron.time_snapped")}
		evs, err := compileRaw(t, `[{"kind":"snap_time","to_tick":9000}]`, in)
		if err != nil {
			t.Fatal(err)
		}
		var p sim.TimeSnappedPayload
		mustUnmarshal(t, evs[0].Payload, &p)
		if evs[0].Type != "metatron.time_snapped" || p.ToTick != 9000 || p.Gratis {
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
		if p.Agent != 0 || p.Text != "a poof of smoke" || p.Salience != sim.SalDream || p.Subject != -1 || p.Origin != sim.OriginOmen {
			t.Errorf("payload = %+v", p)
		}
	})
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
	if err == nil || !strings.Contains(err.Error(), `"metatron.entity_moved" is not in the tool's declared events`) {
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
			in := CompileInput{State: st, Declared: decl("metatron.entity_moved", "metatron.item_granted")}
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
		out[i] = p.Agent
	}
	return out
}

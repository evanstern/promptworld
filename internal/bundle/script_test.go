package bundle

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/toolloop"
	"github.com/evanstern/promptworld/internal/worldmap"
	"go.starlark.net/starlark"
)

// compileSrc writes src to a throwaway tool dir and compiles it, failing the test
// on a compile error (use tryCompileSrc when the error IS the assertion).
func compileSrc(t *testing.T, src string) *scriptProgram {
	t.Helper()
	sp, err := tryCompileSrc(t, src)
	if err != nil {
		t.Fatalf("compileScript: %v", err)
	}
	return sp
}

func tryCompileSrc(t *testing.T, src string) (*scriptProgram, error) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "tool.star"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	return compileScript(dir, "tool.star")
}

// testWorld builds a small frozen world view for executor tests.
func testWorld(tick int64) *worldView {
	return newWorldView("t", tick, 42, 64, 64, []agentInfo{
		{name: "Ash", x: 1, y: 2, alive: true},
		{name: "Birch", x: 3, y: 4, alive: true},
	})
}

// runSrc compiles and executes src with an empty args dict against testWorld.
func runSrc(t *testing.T, src string) ([]Effect, error) {
	t.Helper()
	sp := compileSrc(t, src)
	args, _ := scriptArgs(&Manifest{}, nil)
	return sp.execute(args, testWorld(0), defaultMaxSteps)
}

// TestStepCapDeterministicAbort (SC-006, quickstart Scenario 3): a runaway loop is
// aborted at the step cap with a descriptive error, and the abort is DETERMINISTIC
// — the same step count across two runs. Nothing is produced, so nothing can land.
func TestStepCapDeterministicAbort(t *testing.T) {
	sp := compileSrc(t, "def apply(args, world):\n"+
		"    x = 0\n"+
		"    for i in range(100000000):\n"+
		"        x += i\n"+
		"    return []\n")
	args, _ := scriptArgs(&Manifest{}, nil)

	run := func() (uint64, error) {
		th := newScriptThread("t", 1000)
		_, err := starlark.Call(th, sp.apply, starlark.Tuple{args, testWorld(0)}, nil)
		return th.Steps, err
	}
	s1, e1 := run()
	s2, e2 := run()
	if e1 == nil || e2 == nil {
		t.Fatalf("expected step-cap errors, got %v / %v", e1, e2)
	}
	if !strings.Contains(e1.Error(), "too many steps") {
		t.Errorf("abort error = %q, want 'too many steps'", e1.Error())
	}
	if s1 != s2 {
		t.Errorf("step count not deterministic across runs: %d vs %d", s1, s2)
	}

	// End-to-end: the handler rejects and nothing lands (no state change).
	assertScriptRejectsNoInject(t, "def apply(args, world):\n"+
		"    for i in range(100000000):\n"+
		"        pass\n"+
		"    return []\n", "too many steps")
}

// TestSandboxNoAmbientCapabilities (SC-006): a script has no clock, filesystem,
// network, or module surface. load() has no implementation (boot error), and
// time/os are not resolvable names (compile error). Confirms the builtins surface
// carries no clock.
func TestSandboxNoAmbientCapabilities(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{"load", "load(\"time.star\", \"now\")\ndef apply(args, world):\n    return []\n", ""},
		{"time", "def apply(args, world):\n    return time.now()\n", "undefined: time"},
		{"os", "def apply(args, world):\n    return os.getcwd()\n", "undefined: os"},
		{"json_module", "def apply(args, world):\n    return json.encode([])\n", "undefined: json"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := tryCompileSrc(t, c.src)
			if err == nil {
				t.Fatalf("%s: expected a compile/boot error, got none", c.name)
			}
			if c.want != "" && !strings.Contains(err.Error(), c.want) {
				t.Errorf("%s: error = %q, want it to contain %q", c.name, err.Error(), c.want)
			}
		})
	}
}

// TestFailPropagatesAsError (SC-006): a script calling fail() surfaces as an
// invocation error (fed back as a rejection), never a Go panic.
func TestFailPropagatesAsError(t *testing.T) {
	_, err := runSrc(t, "def apply(args, world):\n    fail(\"no light here\")\n")
	if err == nil {
		t.Fatal("expected fail() to propagate as an error")
	}
	if !strings.Contains(err.Error(), "no light here") {
		t.Errorf("error = %q, want it to carry the fail() message", err.Error())
	}
}

// TestMalformedReturnsRejected (SC-006, contracts/script-api.md Output rules): the
// conversion layer rejects every ill-formed return shape — non-list results,
// non-dict elements, unknown kinds, unknown fields, and non-integer / non-finite
// numerics (float, and float('nan') specifically, which the JSON path cannot even
// express).
func TestMalformedReturnsRejected(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{"non_list", "def apply(args, world):\n    return 5\n", "must return a list"},
		{"non_dict_element", "def apply(args, world):\n    return [5]\n", "must be a dict"},
		{"unknown_kind", "def apply(args, world):\n    return [{\"kind\": \"heal\"}]\n", "unknown kind"},
		{"unknown_field", "def apply(args, world):\n    return [{\"kind\": \"narrate\", \"text\": \"hi\", \"recipients\": \"all_living\", \"extra\": 1}]\n", "unknown field"},
		{"missing_field", "def apply(args, world):\n    return [{\"kind\": \"move_entity\", \"target\": \"Ash\", \"to_x\": 1}]\n", "\"to_y\" is required"},
		{"float_coord", "def apply(args, world):\n    return [{\"kind\": \"move_entity\", \"target\": \"Ash\", \"to_x\": 1.5, \"to_y\": 2}]\n", "must be an integer"},
		{"nan_coord", "def apply(args, world):\n    return [{\"kind\": \"move_entity\", \"target\": \"Ash\", \"to_x\": float(\"nan\"), \"to_y\": 2}]\n", "must be an integer"},
		{"bad_recipients", "def apply(args, world):\n    return [{\"kind\": \"narrate\", \"text\": \"hi\", \"recipients\": \"someone\"}]\n", "must be \"all_living\" or \"target\""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := runSrc(t, c.src)
			if err == nil {
				t.Fatalf("%s: expected rejection, got none", c.name)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("%s: error = %q, want it to contain %q", c.name, err.Error(), c.want)
			}
		})
	}
}

// TestUndeclaredEventTypeRejected (SC-006): a well-formed script whose effect
// produces an event type outside the manifest's declared set is rejected by the
// shared compiler (the same declared-events subset gate the declarative path uses).
func TestUndeclaredEventTypeRejected(t *testing.T) {
	s := genesisState()
	// The script returns a move (metatron.entity_moved), but the tool declares only
	// agent.memory_added.
	effects, err := runOnState(t,
		"def apply(args, world):\n"+
			"    return [{\"kind\": \"move_entity\", \"target\": \"Ash\", \"to_x\": 5, \"to_y\": 5}]\n", s)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	_, err = CompileEffects(effects, CompileInput{
		State: s, Tick: 0, Declared: map[string]bool{"agent.memory_added": true}})
	if err == nil {
		t.Fatal("expected the compiler to reject the undeclared metatron.entity_moved event")
	}
	if !strings.Contains(err.Error(), "not in the tool's declared events") {
		t.Errorf("error = %q, want the declared-events rejection", err.Error())
	}
}

// runOnState executes src against a world view built from a real state's agents.
func runOnState(t *testing.T, src string, s *sim.State) ([]Effect, error) {
	t.Helper()
	sp := compileSrc(t, src)
	args, _ := scriptArgs(&Manifest{}, nil)
	world := newWorldView("t", 0, s.Seed, 64, 64, agentInfos(s))
	return sp.execute(args, world, defaultMaxSteps)
}

// genesisState is a fresh seed-42 world with all villagers alive and charges
// banked — the shared fixture the executor/compiler tests resolve targets against.
func genesisState() *sim.State {
	m := worldmap.Generate(42, 64, 64)
	s := sim.NewState(42, m)
	s.MetatronCharges = sim.MetatronChargeCap
	return s
}

// assertScriptRejectsNoInject runs a script through the FULL handler with a
// capturing injector and asserts a rejected_gate carrying want, with no batch
// landed (proving a failed script changes no state).
func assertScriptRejectsNoInject(t *testing.T, src, want string) {
	t.Helper()
	dir := t.TempDir()
	toolDir := filepath.Join(dir, "bundles", "demo", "tools", "probe")
	if err := os.MkdirAll(toolDir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{"name":"probe","description":"a runaway script probe","events":["agent.memory_added"],"script":"tool.star","limits":{"max_steps":1000}}`
	if err := os.WriteFile(filepath.Join(toolDir, "tool.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(toolDir, "tool.star"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	bs, err := Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if names := rosterNames(bs); len(names) != 1 || names[0] != "probe" {
		t.Fatalf("roster = %v, want [probe] (report: %+v)", names, bs.BootReport())
	}
	s := genesisState()
	landed := 0
	inject := func(ev []store.Event) error { landed += len(ev); return nil }
	probe := &sim.State{Agents: make([]sim.Agent, len(s.Agents))}
	for i := range s.Agents {
		probe.Agents[i] = sim.Agent{Name: s.Agents[i].Name, X: s.Agents[i].X, Y: s.Agents[i].Y, Dead: s.Agents[i].Dead}
	}
	ic := InvocationContext{State: probe, Tick: 0, Invoker: "the angel", Inject: inject, Seed: s.Seed, MapWidth: 64, MapHeight: 64}
	out := bs.Handlers(ic)["probe"](context.Background(), llm.ToolCall{Name: "probe"})
	if out.Verdict != toolloop.VerdictRejectedGate {
		t.Fatalf("verdict = %q (%s), want rejected_gate", out.Verdict, out.ResultForModel)
	}
	if out.Err != nil {
		t.Errorf("Err = %v, want nil (author-level failure, not infrastructure)", out.Err)
	}
	if !strings.Contains(out.ResultForModel, want) {
		t.Errorf("reason = %q, want it to contain %q", out.ResultForModel, want)
	}
	if landed != 0 {
		t.Errorf("a rejected script landed %d events (want 0 — no state change)", landed)
	}
}

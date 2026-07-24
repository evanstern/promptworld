package metatron

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/bundle"
	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/tool"
)

// memoryContains reports whether any of the agent's memories contains sub.
func memoryContains(a sim.Agent, sub string) bool {
	for _, mem := range a.Memories {
		if strings.Contains(mem.Text, sub) {
			return true
		}
	}
	return false
}

// bundleWorld discovers a bundle fixture world under internal/bundle/testdata,
// referenced across packages by relative path (the test's cwd is this package).
func bundleWorld(t *testing.T, name string) *bundle.BundleSet {
	t.Helper()
	bs, err := bundle.Discover(filepath.Join("..", "bundle", "testdata", "worlds", name))
	if err != nil {
		t.Fatalf("Discover(%s): %v", name, err)
	}
	return bs
}

// TestBundleTeleportEndToEnd is US1 / quickstart Scenario 1 through the metatron
// turn seam (T016): a world with the declarative teleport bundle puts `teleport`
// on the angel's roster with its derived schema, and invoking it moves the target
// villager, narrates to every living villager, and lands only the declared event
// types — the whole declarative pipeline, boot-frozen set through turn assembly.
func TestBundleTeleportEndToEnd(t *testing.T) {
	mt, orch, inj, _ := newTestAngel(t, "It is done.")
	mt.SetBundles(bundleWorld(t, "declarative"))
	inj.state.MetatronCharges = 3

	// The derived schema is correct: required target/x/y over the roster tool.
	bt := mt.bundles.Roster()[0]
	if bt.Name != "teleport" {
		t.Fatalf("roster[0] = %q, want teleport", bt.Name)
	}
	schema := string(tool.InputSchema(bt))
	for _, want := range []string{`"target"`, `"x"`, `"y"`, `"required"`} {
		if !strings.Contains(schema, want) {
			t.Errorf("derived schema missing %s: %s", want, schema)
		}
	}

	// Move Ash (0) onto Birch's (1) living, passable tile.
	ax, ay := inj.state.Agents[0].X, inj.state.Agents[0].Y
	bx, by := inj.state.Agents[1].X, inj.state.Agents[1].Y
	if ax == bx && ay == by {
		t.Skip("fixture seed placed Ash and Birch on the same tile")
	}
	before := inj.state.MetatronCharges
	living := len(inj.state.LivingAgents())

	mt.runLoop = actLoop(mt, "teleport", fmt.Sprintf(`{"target":"Ash","x":%d,"y":%d}`, bx, by))
	if _, err := mt.Turn(context.Background(), "teleport Ash across the map"); err != nil {
		t.Fatalf("Turn: %v", err)
	}

	// The teleport tool named itself on the angel's turn surface (roster ⇒
	// guidance via the PromptGloss fallback, T012).
	sys := orch.requests()[0].System
	if !strings.Contains(sys, "teleport(target, x, y)") {
		t.Errorf("system prompt omits teleport guidance:\n%s", sys)
	}

	// The villager moved.
	if inj.state.Agents[0].X != bx || inj.state.Agents[0].Y != by {
		t.Errorf("Ash at (%d,%d), want (%d,%d)", inj.state.Agents[0].X, inj.state.Agents[0].Y, bx, by)
	}
	// A charge was spent by the move (reducer-authoritative — no turn-side gate).
	if inj.state.MetatronCharges >= before {
		t.Errorf("charges = %d, want < %d (a charge should have been spent)", inj.state.MetatronCharges, before)
	}
	// Every living villager gained the narration memory; only declared types landed.
	got := 0
	for i := range inj.state.Agents {
		if inj.state.Agents[i].Dead {
			continue
		}
		if memoryContains(inj.state.Agents[i], "vanished in a poof of smoke") {
			got++
		}
	}
	if got != living {
		t.Errorf("narration reached %d living villagers, want %d", got, living)
	}
	for _, batch := range inj.batches {
		for _, e := range batch {
			switch e.Type {
			case "metatron.entity_moved", "agent.memory_added", "cog.tool_call":
				// declared bundle events + the loop's own telemetry batch
			default:
				t.Errorf("unexpected event type landed: %q", e.Type)
			}
		}
	}
}

// TestBundleBootRejectionReachesRoster is US1 / quickstart Scenario 2 (T018): a
// world with one valid and one off-whitelist bundle tool boots; the BootReport
// names the offending file + rule (T3), and the VALID sibling still reaches the
// angel's turn roster while the rejected tool is absent everywhere.
func TestBundleBootRejectionReachesRoster(t *testing.T) {
	mt, orch, _, _ := newTestAngel(t, "All is well.")
	bs := bundleWorld(t, "offwhitelist")
	mt.SetBundles(bs)

	// The rejection is loud and specific (SC-005): file + rule + offending value.
	var found *bundle.BootIssue
	for i := range bs.BootReport() {
		if bs.BootReport()[i].Rule == "T3" && bs.BootReport()[i].Tool == "healtool" {
			found = &bs.BootReport()[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("no T3 rejection for healtool in %+v", bs.BootReport())
	}
	if found.Severity != "error" ||
		found.File != filepath.Join("demo", "tools", "healtool", "tool.json") ||
		!strings.Contains(found.Message, "metatron.heal") {
		t.Errorf("boot issue = %+v", *found)
	}

	// Only the valid tool survives onto the roster.
	if names := (func() []string {
		var out []string
		for _, tl := range bs.Roster() {
			out = append(out, tl.Name)
		}
		return out
	}()); len(names) != 1 || names[0] != "goodtool" {
		t.Fatalf("roster = %v, want [goodtool]", names)
	}

	// The valid tool reaches the angel's turn surface; the rejected one never does.
	mt.runLoop = converseLoop(mt)
	if _, err := mt.Turn(context.Background(), "what can you do?"); err != nil {
		t.Fatalf("Turn: %v", err)
	}
	sys := orch.requests()[0].System
	if !strings.Contains(sys, "goodtool") {
		t.Errorf("system prompt omits the valid bundle tool goodtool:\n%s", sys)
	}
	if strings.Contains(sys, "healtool") {
		t.Error("system prompt leaks the rejected tool healtool")
	}
}

// TestBundleCastLightNightVsDay is US3 / quickstart Scenario 3 (T027): the scripted
// cast_light tool branches on world.time_of_day. Driven at night, the light blooms
// for every living villager; driven by day, only its target notices, with the
// daylight text — and either way only the declared agent.memory_added lands. The
// clock is driven by setting the replica tick and re-mirroring, exactly the surface
// the turn worker reads.
func TestBundleCastLightNightVsDay(t *testing.T) {
	nightTick := clock.TickAt(1, 23, 0, 0) // 23:00 → night
	dayTick := clock.TickAt(1, 13, 0, 0)   // 13:00 → day

	const nightText = "A soft light blooms over Ash."
	const dayText = "The light is invisible in daylight."

	t.Run("night", func(t *testing.T) {
		mt, _, inj, _ := newTestAngel(t, "the light is cast")
		mt.SetBundles(bundleWorld(t, "scripted"))
		mt.replica.Tick = nightTick
		mt.mirrorState()

		mt.runLoop = actLoop(mt, "cast_light", `{"target":"Ash"}`)
		if _, err := mt.Turn(context.Background(), "cast light on Ash"); err != nil {
			t.Fatalf("Turn: %v", err)
		}

		living, saw := 0, 0
		for i := range inj.state.Agents {
			if inj.state.Agents[i].Dead {
				continue
			}
			living++
			if memoryContains(inj.state.Agents[i], nightText) {
				saw++
			}
		}
		if saw != living {
			t.Errorf("night bloom reached %d of %d living villagers, want all", saw, living)
		}
		assertOnlyBundleEvents(t, inj)
	})

	t.Run("day", func(t *testing.T) {
		mt, _, inj, _ := newTestAngel(t, "the light is cast")
		mt.SetBundles(bundleWorld(t, "scripted"))
		mt.replica.Tick = dayTick
		mt.mirrorState()

		mt.runLoop = actLoop(mt, "cast_light", `{"target":"Ash"}`)
		if _, err := mt.Turn(context.Background(), "cast light on Ash"); err != nil {
			t.Fatalf("Turn: %v", err)
		}

		// The day branch narrates the daylight text to the target only.
		if !memoryContains(inj.state.Agents[0], dayText) {
			t.Errorf("day branch: Ash (target) did not receive the daylight narration")
		}
		for i := range inj.state.Agents {
			if memoryContains(inj.state.Agents[i], "A soft light blooms") {
				t.Errorf("day branch leaked the night bloom to %s", inj.state.Agents[i].Name)
			}
			if i > 0 && memoryContains(inj.state.Agents[i], dayText) {
				t.Errorf("day branch reached %s beyond the target (recipients \"target\")", inj.state.Agents[i].Name)
			}
		}
		assertOnlyBundleEvents(t, inj)
	})
}

// TestBundlePersonaComposesIdentityGrantAndTools is US4 / quickstart Scenario 6
// (T032): the gandalf persona bundle (T031 fixture) installs its SOUL fragment,
// its capabilities.json grant narrowing, and its tools/ folder all together at
// boot. The broken sibling tool is skipped with a T-rule BootReport entry
// (clarification #1) while the SOUL fragment, the grant narrowing, and the
// valid tool all stay active — a per-tool failure never rejects the persona.
func TestBundlePersonaComposesIdentityGrantAndTools(t *testing.T) {
	mt, orch, _, _ := newTestAngel(t, "As you wish.")
	bs := bundleWorld(t, "persona")
	mt.SetBundles(bs)

	// The broken tool is skipped with a specific T-rule BootReport entry; the
	// sibling valid tool, SOUL, and grant all still take effect.
	found := false
	for _, iss := range bs.BootReport() {
		if iss.Tool == "broken" && iss.Rule == "T1" && iss.Severity == "error" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a T1 rejection for the broken tool; report=%+v", bs.BootReport())
	}

	// Only the valid tool survives onto the roster.
	if names := (func() []string {
		var out []string
		for _, tl := range bs.Roster() {
			out = append(out, tl.Name)
		}
		return out
	}()); len(names) != 1 || names[0] != "bless" {
		t.Fatalf("roster = %v, want [bless]", names)
	}

	mt.runLoop = converseLoop(mt)
	if _, err := mt.Turn(context.Background(), "what do you carry, old friend?"); err != nil {
		t.Fatalf("Turn: %v", err)
	}
	sys := orch.requests()[0].System

	// The persona's SOUL fragment reached the system prompt.
	if !strings.Contains(sys, "Gandalf") {
		t.Errorf("system prompt omits the persona's SOUL fragment:\n%s", sys)
	}
	// The valid sibling tool reaches the derived guidance.
	if !strings.Contains(sys, "bless") {
		t.Errorf("system prompt omits the valid persona tool bless:\n%s", sys)
	}
	// The rejected tool never leaks (neither its declared nor its manifest name).
	if strings.Contains(sys, "brokentool") {
		t.Error("system prompt leaks the rejected persona tool brokentool")
	}
	// The grant narrowing excludes "move" (not named by the fixture's
	// capabilities.json) from the declared work_miracle kind enum, while the
	// three kinds the fixture DOES name remain granted.
	if strings.Contains(sys, `"move"`) {
		t.Error("system prompt still offers the world-excluded miracle kind \"move\"")
	}
	for _, want := range []string{`"remove"`, `"give_item"`, `"time_snap"`} {
		if !strings.Contains(sys, want) {
			t.Errorf("system prompt missing granted miracle kind %s:\n%s", want, sys)
		}
	}
}

// assertOnlyBundleEvents fails if any landed event is outside cast_light's declared
// surface (agent.memory_added) plus the loop's own cog.tool_call telemetry.
func assertOnlyBundleEvents(t *testing.T, inj *stateInjector) {
	t.Helper()
	for _, batch := range inj.batches {
		for _, e := range batch {
			switch e.Type {
			case "agent.memory_added", "cog.tool_call":
			default:
				t.Errorf("unexpected event type landed: %q", e.Type)
			}
		}
	}
}

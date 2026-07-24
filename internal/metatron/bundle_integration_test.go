package metatron

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/bundle"
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

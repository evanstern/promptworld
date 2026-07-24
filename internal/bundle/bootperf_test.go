package bundle

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestBootPerfSanity is the T034 boot-time sanity check (plan.md Technical
// Context: "Boot-time validation of a world with ≤32 bundles completes in
// <1s"). It synthesizes a worst-case-shaped world — 32 bundles, each with 8
// tools split evenly between declarative (template) and scripted (Starlark)
// tool.json/tool.star pairs — and times a single Discover call end to end
// (walk + strict decode + effect-template/Starlark-parse validation for every
// tool). This is not a determinism or correctness proof (those live in
// load_test.go and replay_test.go); it only guards against the boot path
// regressing into something that makes a large bundle install noticeably slow.
func TestBootPerfSanity(t *testing.T) {
	const bundles = 32
	const toolsPerBundle = 8 // ≤16 cap (contracts/boot-validation.md B4)

	worldDir := t.TempDir()
	root := filepath.Join(worldDir, "bundles")

	for b := 0; b < bundles; b++ {
		bundleName := fmt.Sprintf("bundle_%02d", b)
		for i := 0; i < toolsPerBundle; i++ {
			scripted := i%2 == 1
			toolName := fmt.Sprintf("tool_%02d_%d", b, i)
			dir := filepath.Join(root, bundleName, "tools", toolName)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatalf("MkdirAll(%s): %v", dir, err)
			}

			if scripted {
				manifest := fmt.Sprintf(`{
  "name": %q,
  "description": "synthetic scripted tool for boot-perf sanity",
  "params": [
    {"name": "target", "kind": "agent_name", "required": true}
  ],
  "events": ["agent.memory_added"],
  "charges": 0,
  "script": "tool.star",
  "limits": {"max_steps": 100000}
}
`, toolName)
				if err := os.WriteFile(filepath.Join(dir, "tool.json"), []byte(manifest), 0o644); err != nil {
					t.Fatalf("write tool.json: %v", err)
				}
				script := `def apply(args, world):
    if world.time_of_day == "night":
        return [{"kind": "narrate", "text": "a bell tolls for " + args["target"], "recipients": "all_living"}]
    return [{"kind": "narrate", "text": "quiet.", "recipients": "target", "target": args["target"]}]
`
				if err := os.WriteFile(filepath.Join(dir, "tool.star"), []byte(script), 0o644); err != nil {
					t.Fatalf("write tool.star: %v", err)
				}
			} else {
				manifest := fmt.Sprintf(`{
  "name": %q,
  "description": "synthetic declarative tool for boot-perf sanity",
  "params": [
    {"name": "target", "kind": "agent_name", "required": true},
    {"name": "x", "kind": "number", "required": true},
    {"name": "y", "kind": "number", "required": true}
  ],
  "events": ["metatron.entity_moved", "agent.memory_added"],
  "charges": 1,
  "effects": [
    {"kind": "move_entity", "target": "{args.target}", "to_x": "{args.x}", "to_y": "{args.y}"},
    {"kind": "narrate", "text": "{args.target} stirs.", "recipients": "all_living"}
  ]
}
`, toolName)
				if err := os.WriteFile(filepath.Join(dir, "tool.json"), []byte(manifest), 0o644); err != nil {
					t.Fatalf("write tool.json: %v", err)
				}
			}
		}
	}

	start := time.Now()
	bs, err := Discover(worldDir)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if len(bs.BootReport()) != 0 {
		t.Fatalf("unexpected boot issues on a synthetically valid %d-bundle world: %+v", bundles, bs.BootReport())
	}
	if want := bundles * toolsPerBundle; len(bs.Roster()) != want {
		t.Fatalf("roster length = %d, want %d", len(bs.Roster()), want)
	}

	t.Logf("Discover() over %d bundles / %d tools took %s", bundles, bundles*toolsPerBundle, elapsed)
	if elapsed >= time.Second {
		t.Errorf("boot-time validation took %s, want <1s (plan.md Technical Context) for %d bundles / %d tools",
			elapsed, bundles, bundles*toolsPerBundle)
	}
}

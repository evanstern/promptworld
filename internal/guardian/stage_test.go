package guardian

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/persona"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/tool"
)

// Spec 046 US2: the stage ceiling and the stage-1 instruction lock. These tests
// pin the gate-to-feature pathway: per stage, the post-intersection roster
// equals the ladder's ceiling exactly; beyond-stage acts are refused at the
// door; declaration/prose/door derive from the one roster; the stage-1 lock is
// honest (preset binds, notices name the unlocking stage); and stage gating
// never perturbs the world (cross-stage determinism diff, FR-006).

// fullLoopRosterNames is the live roster's names in registry order — the
// no-ceiling expectation for stage-3/-4 and pre-ladder worlds.
func fullLoopRosterNames() []string {
	var out []string
	for _, t := range tool.LoopRosterGuardian() {
		out = append(out, t.Name)
	}
	return out
}

// TestStageCeilingRosterTable (T007): table-driven — for every stage, with the
// default (absent) manifest, the post-intersection granted roster equals the
// ceiling table exactly (contracts/stage-gating.md).
func TestStageCeilingRosterTable(t *testing.T) {
	dir := t.TempDir() // no capabilities.json — full default grant within the ceiling
	cases := []struct {
		stage string
		want  []string
	}{
		{"", fullLoopRosterNames()}, // pre-ladder world: ungated
		// Ratified amendment: standing orders are the stage-1 watch primitive;
		// spec 063 adds the read-only explain (the tutor stage's grounding tool).
		{"stage-1", []string{"send_omen", "send_vision", "monitor_and_act", "cancel_order", "explain", "place_designation", "cancel_designation", "issue_directive", "cancel_directive", "survey_site", "prophesy", "brief_myths", "accept_mission", "note_mission_progress", "cancel_mission"}},
		{"stage-2", []string{"send_omen", "send_vision", "monitor_and_act", "cancel_order", "explain", "place_designation", "cancel_designation", "issue_directive", "cancel_directive", "survey_site", "prophesy", "brief_myths", "accept_mission", "note_mission_progress", "cancel_mission"}}, // stage-2 unlocks the charter, not tools
		{"stage-3", fullLoopRosterNames()},
		{"stage-4", fullLoopRosterNames()},
	}
	for _, c := range cases {
		g, notices := loadManifest(dir)
		if len(notices) != 0 {
			t.Fatalf("%q: unexpected manifest notices %v", c.stage, notices)
		}
		g = applyStageCeiling(g, c.stage)
		if got := g.grantedTools(); !reflect.DeepEqual(got, c.want) {
			t.Errorf("stage %q granted roster = %v, want %v", c.stage, got, c.want)
		}
		// The declared roster (what the model sees) matches the same set — the
		// declaration layer inherits the ceiling.
		var declared []string
		for _, tl := range grantedRoster(g) {
			declared = append(declared, tl.Name)
		}
		if !reflect.DeepEqual(declared, c.want) {
			t.Errorf("stage %q declared roster = %v, want %v", c.stage, declared, c.want)
		}
	}
}

// TestStageCeilingIntersectsManifest (T007): a player manifest may narrow
// WITHIN the ceiling but never exceed it (intersection-only, FR-004), and
// bundle tools are intersected away below stage-3.
func TestStageCeilingIntersectsManifest(t *testing.T) {
	dir := t.TempDir()
	write := func(s string) {
		if err := os.WriteFile(filepath.Join(dir, "capabilities.json"), []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Narrowing within the ceiling holds.
	write(`{"tools":["send_vision"]}`)
	g, _ := loadManifest(dir)
	g = applyStageCeiling(g, "stage-1")
	if got := g.grantedTools(); !reflect.DeepEqual(got, []string{"send_vision"}) {
		t.Errorf("within-ceiling narrowing: got %v, want [send_vision]", got)
	}

	// A manifest naming beyond-stage tools cannot exceed the ceiling
	// (monitor_and_act is IN the stage-1 ceiling per the ratified amendment;
	// work_miracle is not).
	write(`{"tools":["work_miracle","monitor_and_act","send_omen"]}`)
	g, _ = loadManifest(dir)
	g = applyStageCeiling(g, "stage-1")
	if got := g.grantedTools(); !reflect.DeepEqual(got, []string{"send_omen", "monitor_and_act"}) {
		t.Errorf("beyond-ceiling manifest: got %v, want [send_omen monitor_and_act]", got)
	}
	// The same manifest at stage-4 keeps its full (valid) grant — the ceiling,
	// not the manifest, is what changed.
	g, _ = loadManifest(dir)
	g = applyStageCeiling(g, "stage-4")
	if got := g.grantedTools(); !reflect.DeepEqual(got, []string{"send_omen", "monitor_and_act", "work_miracle"}) {
		t.Errorf("stage-4 same manifest: got %v", got)
	}

	// Bundle tools: granted by default in an unconstrained world, intersected
	// away by the stage-1 ceiling ("no bundles" below stage-3).
	os.Remove(filepath.Join(dir, "capabilities.json"))
	g, _ = loadManifest(dir)
	if !g.allowsBundle("weather_bless") {
		t.Fatal("unconstrained world should grant bundle tools (spec 036 default)")
	}
	g = applyStageCeiling(g, "stage-1")
	if g.allowsBundle("weather_bless") {
		t.Error("stage-1 ceiling must exclude bundle tools")
	}
	g, _ = loadManifest(dir)
	if g = applyStageCeiling(g, "stage-3"); !g.allowsBundle("weather_bless") {
		t.Error("stage-3 imposes no ceiling — bundle tools stay granted")
	}
}

// TestStageDoorRefusesBeyondStage (T007; ratified amendment): the door layer —
// a stage-1 grant installs no handler for beyond-stage tools (structural
// absence) and refuses a world-shaping miracle in-fiction even if a call is
// conjured (defense-in-depth), while the watch primitive (monitor_and_act /
// cancel_order) — ratified into the stage-1 ceiling — lands normally.
func TestStageDoorRefusesBeyondStage(t *testing.T) {
	mt, _, _, dir := newTestGuardian(t, "so be it")
	mt.SetStage("stage-1", "")
	g, _ := loadManifest(dir)
	g = applyStageCeiling(g, "stage-1")

	d := &turnDispatch{mt: mt, charges: 3, alive: map[int]bool{0: true}, tick: 1, result: &TurnResult{}, grant: g}
	h := mt.turnHandlers(d)
	for _, beyond := range []string{"work_miracle", "pause", "start", "adjust_speed"} {
		if _, ok := h[beyond]; ok {
			t.Errorf("stage-1 handlers should not install %s (structural absence at the door)", beyond)
		}
	}
	for _, granted := range []string{"send_vision", "send_omen", "monitor_and_act", "cancel_order"} {
		if _, ok := h[granted]; !ok {
			t.Errorf("stage-1 handlers missing the granted %s", granted)
		}
	}

	// Defense-in-depth: even a conjured call is refused in-fiction.
	if m, why := mt.landMiracle(miracleArgs{Kind: "give_item", Villager: sim.AgentNames[0], Item: "berries", Qty: 1}, 3, g); m != nil || why == "" {
		t.Errorf("stage-1 miracle should refuse at the door, got (%v, %q)", m, why)
	}
	// The watch primitive is IN the stage-1 ceiling (ratified amendment): a
	// well-formed placement lands rather than refusing.
	if o, why := mt.placeOrder("player", orderArgs{Condition: "x", Action: "y", EventTypes: []string{"sim.night_started"}}, 0, g); o == nil || why != "" {
		t.Errorf("stage-1 player order should land at the door (ratified amendment), got (%v, %q)", o, why)
	}
	// The granted surface still works: a vision lands through the same grant.
	if n, why := mt.landVision(sim.AgentNames[0], "beware the night", nil, 1, map[int]bool{0: true}, g); n == nil || why != "" {
		t.Errorf("stage-1 vision should land, got (%v, %q)", n, why)
	}
}

// TestStageThreeLayerCoherence (T007): declaration, prose, and door all derive
// from the one post-intersection roster — a beyond-stage tool is absent from
// every layer at once (spec 021's can't-disagree property, inherited).
func TestStageThreeLayerCoherence(t *testing.T) {
	dir := t.TempDir()
	g, _ := loadManifest(dir)
	g = applyStageCeiling(g, "stage-1")
	roster := grantedRoster(g)

	// Declaration: only ceiling tools are declared (ratified amendment: the
	// watch primitive is in the stage-1 ceiling).
	for _, tl := range roster {
		switch tl.Name {
		case "send_omen", "send_vision", "monitor_and_act", "cancel_order", "explain",
			"place_designation", "cancel_designation", "issue_directive", "cancel_directive", "survey_site", "prophesy", "brief_myths",
			"accept_mission", "note_mission_progress", "cancel_mission":
		default:
			t.Errorf("declared roster leaks beyond-stage tool %s", tl.Name)
		}
	}
	// Prose: the derived guidance names granted tools only.
	guidance := tool.GuardianToolGuidance(roster)
	for _, beyond := range []string{"work_miracle", "adjust_speed", "canonize_region"} {
		if strings.Contains(guidance, beyond) {
			t.Errorf("tool guidance mentions beyond-stage %s", beyond)
		}
	}
	if !strings.Contains(guidance, "send_vision") || !strings.Contains(guidance, "send_omen") {
		t.Error("tool guidance should describe the granted nudges")
	}
	if !strings.Contains(guidance, "monitor_and_act") || !strings.Contains(guidance, "cancel_order") {
		t.Error("tool guidance should describe the granted watch primitive")
	}
	// Door: the grant refuses what the roster omits.
	if g.allows("work_miracle") || g.allowsKind("give_item") == true && g.allows("work_miracle") {
		t.Error("door grant leaks work_miracle beyond the stage")
	}
}

// TestStageOneInstructionLock (T007, FR-005): at stage-1 the preset constant is
// the effective charter and skills never compose — an edited charter.md and
// present skill files produce the honest notices naming the unlocking stages,
// and the composed system prompt carries the preset, not the player text.
func TestStageOneInstructionLock(t *testing.T) {
	mt, orch, _, dir := newTestGuardian(t, "I hear you.")
	mt.SetStage("stage-1", "")
	custom := "# MY LAW\n\nSpeak only in riddles about squirrels.\n"
	if err := os.WriteFile(filepath.Join(dir, "charter.md"), []byte(custom), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "skills", "10-weather.md"), []byte("Always forecast rain."), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := mt.Turn(context.Background(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	sys := orch.requests()[0].System
	if !strings.Contains(sys, "faithful, competent") {
		t.Error("stage-1 system prompt should carry the preset (default) charter")
	}
	if strings.Contains(sys, "squirrels") {
		t.Error("stage-1 system prompt must not carry the edited charter.md")
	}
	if strings.Contains(sys, "--- skill:") || strings.Contains(sys, "forecast rain") {
		t.Error("stage-1 system prompt must not compose skill files")
	}
	wantCharter := "charter.md does not bind at this stage — The Written Word (stage-2) unlocks instruction authoring"
	wantSkills := "skill files do not bind at this stage — The Craft (stage-3) unlocks skill files"
	if !strings.Contains(res.Reply, wantCharter) {
		t.Errorf("reply missing the charter lock notice, got: %q", res.Reply)
	}
	if !strings.Contains(res.Reply, wantSkills) {
		t.Errorf("reply missing the skills lock notice, got: %q", res.Reply)
	}

	// Status reports the lock provenance and the ceiled surface (the twin site).
	st := mt.Status()
	if st.Stage != "stage-1" || !st.CharterLocked || st.CharterPreset != "default" || !st.SkillsLocked {
		t.Errorf("status lock provenance = %+v", st)
	}
	if st.Skills != nil {
		t.Errorf("status Skills should be empty below stage-3 (nothing composes), got %v", st.Skills)
	}
	if !reflect.DeepEqual(st.GrantedTools, []string{"send_omen", "send_vision", "monitor_and_act", "cancel_order", "explain", "place_designation", "cancel_designation", "issue_directive", "cancel_directive", "survey_site", "prophesy", "brief_myths", "accept_mission", "note_mission_progress", "cancel_mission"}) {
		t.Errorf("status granted tools = %v, want the stage-1 ceiling", st.GrantedTools)
	}

	// An unedited charter (bytes == preset) locks silently — no notice noise.
	if err := os.WriteFile(filepath.Join(dir, "charter.md"), []byte(persona.DefaultCharter), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, notice := stageCharter(dir, "stage-1", ""); notice != "" {
		t.Errorf("matching charter.md should produce no lock notice, got %q", notice)
	}
	// A missing charter.md is restored to the preset (a file to edit at
	// stage-2), still no notice — what binds never changed.
	os.Remove(filepath.Join(dir, "charter.md"))
	text, notice := stageCharter(dir, "stage-1", "")
	if text != persona.DefaultCharter || notice != "" {
		t.Errorf("missing charter at stage-1: text preset? %v, notice %q", text == persona.DefaultCharter, notice)
	}
	if restored, err := os.ReadFile(filepath.Join(dir, "charter.md")); err != nil || string(restored) != persona.DefaultCharter {
		t.Error("missing charter.md should be restored to the preset")
	}
}

// TestStageTwoChartersBindSkillsDoNot (T007, FR-005): stage-2 unlocks
// instruction authoring — charter edits bind exactly as an ungated world's —
// while skill files still wait for stage-3.
func TestStageTwoChartersBindSkillsDoNot(t *testing.T) {
	mt, orch, _, dir := newTestGuardian(t, "as you decree")
	mt.SetStage("stage-2", "")
	custom := "# MY LAW\n\nAlways answer in verse.\n"
	if err := os.WriteFile(filepath.Join(dir, "charter.md"), []byte(custom), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "skills", "10-weather.md"), []byte("Always forecast rain."), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := mt.Turn(context.Background(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	sys := orch.requests()[0].System
	if !strings.Contains(sys, "Always answer in verse") {
		t.Error("stage-2 charter edits must bind (today's behavior)")
	}
	if strings.Contains(sys, "--- skill:") {
		t.Error("stage-2 must not compose skill files (stage-3 unlocks them)")
	}
	if !strings.Contains(res.Reply, "The Craft (stage-3) unlocks skill files") {
		t.Errorf("stage-2 reply missing the skills lock notice, got: %q", res.Reply)
	}
	if strings.Contains(res.Reply, "charter.md does not bind") {
		t.Errorf("stage-2 must not carry a charter lock notice, got: %q", res.Reply)
	}

	st := mt.Status()
	if st.Stage != "stage-2" || st.CharterLocked || !st.SkillsLocked || st.CharterPreset != "" {
		t.Errorf("stage-2 status provenance = %+v", st)
	}

	// Stage-3: everything composes, no locks, full surface.
	mt.SetStage("stage-3", "")
	orch.mu.Lock()
	orch.reqs = nil
	orch.mu.Unlock()
	if _, err := mt.Turn(context.Background(), "hello again"); err != nil {
		t.Fatal(err)
	}
	sys = orch.requests()[0].System
	if !strings.Contains(sys, "--- skill: 10-weather.md ---") {
		t.Error("stage-3 should compose skill files")
	}
	st = mt.Status()
	if st.CharterLocked || st.SkillsLocked || st.Stage != "stage-3" {
		t.Errorf("stage-3 status should carry no locks, got %+v", st)
	}
	if !reflect.DeepEqual(st.GrantedTools, grantedToolLabels(func() grantSet { g, _ := loadManifest(dir); return g }())) {
		t.Errorf("stage-3 granted tools should be the full default grant, got %v", st.GrantedTools)
	}
}

// canonicalEvents renders injected batches for comparison: type + tick +
// payload bytes, batch boundaries preserved. Seq is daemon-assigned and always
// zero here.
func canonicalEvents(t *testing.T, batches [][]store.Event) []byte {
	t.Helper()
	var buf bytes.Buffer
	for _, b := range batches {
		for _, e := range b {
			fmt.Fprintf(&buf, "%s %d %s\n", e.Type, e.Tick, e.Payload)
		}
		buf.WriteString("--\n")
	}
	return buf.Bytes()
}

// TestCrossStageDeterminism (T007, FR-006/SC-002): the same seed and the same
// command sequence — a converse turn and a vision (within every stage's grant)
// — produce byte-identical world-event histories and final state hashes at
// stage-1 and stage-4. Only the agent's granted SURFACE differs (roster,
// status); the world cannot tell the stages apart.
func TestCrossStageDeterminism(t *testing.T) {
	type run struct {
		events  []byte
		hash    string
		granted []string
	}
	drive := func(stage string) run {
		mt, _, inj, _ := newTestGuardian(t, "it is done")
		mt.SetStage(stage, "")
		// Command 1: a converse-only turn (no act).
		if _, err := mt.Turn(context.Background(), "how fares the village?"); err != nil {
			t.Fatal(err)
		}
		// Command 2: a vision — granted at every stage, so the same act lands
		// identically whatever the ceiling.
		mt.runLoop = actLoop(mt, "send_vision", `{"target":"`+sim.AgentNames[0]+`","text":"tend the fire"}`)
		if _, err := mt.Turn(context.Background(), "warn them"); err != nil {
			t.Fatal(err)
		}
		return run{
			events:  canonicalEvents(t, inj.batches),
			hash:    inj.state.Hash(),
			granted: mt.Status().GrantedTools,
		}
	}

	s1, s4 := drive("stage-1"), drive("stage-4")
	if !bytes.Equal(s1.events, s4.events) {
		t.Errorf("world-event histories diverged across stages:\nstage-1:\n%s\nstage-4:\n%s", s1.events, s4.events)
	}
	if s1.hash != s4.hash {
		t.Errorf("state hashes diverged across stages: %s vs %s", s1.hash, s4.hash)
	}
	// The surface — and only the surface — differs.
	if reflect.DeepEqual(s1.granted, s4.granted) {
		t.Error("granted surfaces should differ across stages (the diff isolates the surface)")
	}
	if !reflect.DeepEqual(s1.granted, []string{"send_omen", "send_vision", "monitor_and_act", "cancel_order", "explain", "place_designation", "cancel_designation", "issue_directive", "cancel_directive", "survey_site", "prophesy", "brief_myths", "accept_mission", "note_mission_progress", "cancel_mission"}) {
		t.Errorf("stage-1 surface = %v", s1.granted)
	}
	if !reflect.DeepEqual(s4.granted, fullLoopRosterNames()) {
		t.Errorf("stage-4 surface = %v", s4.granted)
	}
}

// TestPresetCharterResolution (T006/T017): the preset table — ""/default
// resolve to the authored default; tutor resolves to persona.TutorCharter;
// charterIsDefault is preset-aware.
func TestPresetCharterResolution(t *testing.T) {
	if presetCharter("") != persona.DefaultCharter || presetCharter("default") != persona.DefaultCharter {
		t.Error("empty/default preset should resolve to persona.DefaultCharter")
	}
	if presetCharter("tutor") != persona.TutorCharter {
		t.Error("tutor preset should resolve to persona.TutorCharter")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "charter.md"), []byte(persona.DefaultCharter), 0o644); err != nil {
		t.Fatal(err)
	}
	if !charterIsDefault(dir) {
		t.Error("charterIsDefault should be preset-aware")
	}
	if charterIsDefault(dir, "tutor") {
		t.Error("a default-charter file should not read as the tutor preset's default")
	}
	if err := os.WriteFile(filepath.Join(dir, "charter.md"), []byte(persona.TutorCharter), 0o644); err != nil {
		t.Fatal(err)
	}
	if !charterIsDefault(dir, "tutor") {
		t.Error("a tutor-charter file should read as default under the tutor preset")
	}
}

// TestUngatedWorldUnchanged (T007 byte-compat): a Guardian with no stage set
// (every pre-046 world and caller) composes, grants, and reports exactly as
// before — no ceiling, no locks, no new status fields.
func TestUngatedWorldUnchanged(t *testing.T) {
	mt, orch, _, dir := newTestGuardian(t, "all is well")
	custom := "# MY LAW\n\nAnswer briefly.\n"
	if err := os.WriteFile(filepath.Join(dir, "charter.md"), []byte(custom), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := mt.Turn(context.Background(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(orch.requests()[0].System, "Answer briefly") {
		t.Error("ungated world: charter edits must bind")
	}
	if strings.Contains(res.Reply, "does not bind at this stage") {
		t.Errorf("ungated world must carry no lock notices, got %q", res.Reply)
	}
	st := mt.Status()
	if st.Stage != "" || st.CharterLocked || st.SkillsLocked || st.CharterPreset != "" {
		t.Errorf("ungated status should carry no stage fields, got %+v", st)
	}
	b, err := json.Marshal(st)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"stage", "charter_locked", "charter_preset", "skills_locked"} {
		if strings.Contains(string(b), `"`+key+`"`) {
			t.Errorf("ungated status JSON must omit %q (additive omitempty), got %s", key, b)
		}
	}
}

// TestTutorPresetHotReloadsLikeAnyCharter (spec 046 T017/T019, FR-012): above
// stage-1 the tutor preset is only the SEEDED starting text — once the
// instruction surface is unlocked (stage-2+), editing charter.md hot-reloads
// exactly like the default preset's charter does. No new mechanics, no
// special-casing: the preset name only matters at stage-1's lock (R3).
func TestTutorPresetHotReloadsLikeAnyCharter(t *testing.T) {
	mt, orch, _, dir := newTestGuardian(t, "as you wish")
	mt.SetStage("stage-2", "tutor")
	if err := os.WriteFile(filepath.Join(dir, "charter.md"), []byte(persona.TutorCharter), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := mt.Turn(context.Background(), "hello"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(orch.requests()[0].System, "You are the village's guardian") {
		t.Error("stage-2 with tutor preset should carry the seeded tutor text before any edit")
	}

	custom := "# MY LAW\n\nAlways greet warmly.\n"
	if err := os.WriteFile(filepath.Join(dir, "charter.md"), []byte(custom), 0o644); err != nil {
		t.Fatal(err)
	}
	orch.mu.Lock()
	orch.reqs = nil
	orch.mu.Unlock()
	if _, err := mt.Turn(context.Background(), "hello again"); err != nil {
		t.Fatal(err)
	}
	sys := orch.requests()[0].System
	if !strings.Contains(sys, "Always greet warmly") {
		t.Error("stage-2 tutor-preset world must hot-reload player edits exactly like any charter")
	}
	if strings.Contains(sys, "You are the village's guardian") {
		t.Error("edited charter.md should fully replace the tutor seed text once unlocked")
	}
	if st := mt.Status(); st.CharterLocked {
		t.Error("stage-2 must report no charter lock even when charter_preset is tutor")
	}
}

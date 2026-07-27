package daemon

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/world"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// TestOakNeglectWindowReplay (spec 083 FR-008, SC-002) is the env-guarded
// evidence probe for the neglect detector: pointed at a COPY of world-01's
// ARCHIVED world.v3.db (PROMPTWORLD_WORLD01_DB — the
// TestSageThrashWindowContextReplay idiom), it replays the recorded log to
// sampled ticks and evaluates the exported detector predicate
// (sim.NeglectDue) over the replayed state.
//
// The archived v3 log is the only file holding the death window: world-01's
// migrated world.db carries events from tick 538,823 on (the migration
// cutoff), AFTER Oak's death at 511,440 — so the spec assumption that
// "either file's history" covers the window is empirically false for
// world.db, and the probe MUST target world.v3.db (whose events the current
// reducer folds fine; the copy is opened so store.Open's schema migration
// never touches the real save). What the log actually shows (calibrated by
// direct query, recorded here as evidence):
//
//   - Oak's warmth crossed below the 350 band at tick 499,320 on the
//     636→0 slide; his last warmth-class intent was goto_warmth at 489,247.
//   - Oak was ASLEEP from 489,580 until the gru attack at 510,173 woke him
//     (the gru.attacked arm folds Asleep=false) — so the neglect window
//     completed (506,520) while he slept, the sweep's asleep skip deferred
//     the firing, and the predicate holds from wake (510,173) to death
//     (511,440): only reflex chop / planner wander in that stretch,
//     warmth 0. Exactly the spec's "asleep villager" edge case, on the real
//     evidence.
//   - Oak's labeled-healthy day-4 window and the never-thrashing Ash/Hazel
//     hold the predicate false across all three needs.
//
// Skipped without the env var: the ~106 MB world-01 log is machine-local,
// never in-repo — the BINDING CI validation is the recorded fixture suite in
// internal/sim/neglect_test.go (FR-007); this probe is recorded evidence when
// run (the spec-043 SC-004 precedent). The pre-083 log carries no
// sim.neglect_detected events; the replay populates the derived anchors
// through the new needs_changed/intent_set reducer arms (the spec-043
// IntentLog precedent for derived-state additions).
func TestOakNeglectWindowReplay(t *testing.T) {
	const (
		oakIdx   = 6 // sim.AgentNames[6] == "Oak"
		ashIdx   = 0 // never-thrashing control (research §2)
		hazelIdx = 5 // never-thrashing control
	)
	dbPath := os.Getenv("PROMPTWORLD_WORLD01_DB")
	if dbPath == "" {
		t.Skip("set PROMPTWORLD_WORLD01_DB to a COPY of world-01's world.db to run the SC-002 evidence replay")
	}
	if sim.AgentNames[oakIdx] != "Oak" || sim.AgentNames[ashIdx] != "Ash" || sim.AgentNames[hazelIdx] != "Hazel" {
		t.Fatal("the world roster changed — Oak/Ash/Hazel indices are stale")
	}

	// Copy the world into an isolated temp dir and open only the copy, so a
	// schema-migrating store.Open can never mutate the real save (world-01 is
	// READ-ONLY — the Sage-probe constraint).
	srcDir := filepath.Dir(dbPath)
	tmp := t.TempDir()
	mustCopy(t, filepath.Join(srcDir, world.ManifestName), filepath.Join(tmp, world.ManifestName))
	mustCopy(t, dbPath, filepath.Join(tmp, "world.db"))
	// Sidecars are named after the SOURCE db (world.v3.db-wal for the
	// archived log) but land beside the copy under its world.db name.
	for _, ext := range []string{"-wal", "-shm"} {
		if _, err := os.Stat(dbPath + ext); err == nil {
			mustCopy(t, dbPath+ext, filepath.Join(tmp, "world.db"+ext))
		}
	}

	var man struct {
		Seed      uint64 `json:"seed"`
		MapWidth  int    `json:"map_width"`
		MapHeight int    `json:"map_height"`
	}
	manBytes, err := os.ReadFile(filepath.Join(tmp, world.ManifestName))
	if err != nil {
		t.Fatalf("read manifest copy: %v", err)
	}
	if err := json.Unmarshal(manBytes, &man); err != nil {
		t.Fatalf("parse manifest: %v", err)
	}
	m := worldmap.Generate(man.Seed, man.MapWidth, man.MapHeight)

	st, err := store.Open(filepath.Join(tmp, "world.db"))
	if err != nil {
		t.Fatalf("open store copy: %v", err)
	}
	defer st.Close()

	replayTo := func(cutoff int64) *sim.State {
		t.Helper()
		state, skipped, err := replayToTick(man.Seed, m, st, cutoff)
		if err != nil {
			t.Fatalf("replayToTick %d: %v", cutoff, err)
		}
		// The anchors derive from needs_changed + intent_set; if either was
		// rejected on replay the predicate is not trustworthy.
		for _, typ := range []string{"agent.needs_changed", "agent.intent_set"} {
			if skipped[typ] > 0 {
				t.Fatalf("%d %s event(s) rejected on replay — the neglect anchors are not trustworthy", skipped[typ], typ)
			}
		}
		return state
	}

	// --- Oak's death window: the predicate holds while he is awake ---------
	// Death at tick 511,440 (day 7 04:04); band entry 499,320; window
	// complete 506,520; asleep 489,580 → 510,173 (gru attack wake). The
	// asleep samples must show the anchors accrued with the sweep deferring
	// (the "asleep villager" edge case); the awake samples must hold the
	// predicate — the detector fires on the first heartbeat after wake,
	// still 21 game-minutes before the death the log records.
	const bandEntry = int64(499320)
	deathSamples := []int64{500400, 504000, 508800, 510300, 511200}
	awakeTrue := 0
	for _, tick := range deathSamples {
		state := replayTo(tick)
		oak := &state.Agents[oakIdx]
		if oak.Dead {
			t.Fatalf("Oak already dead at tick %d — the sample window is wrong", tick)
		}
		due := sim.NeglectDue(oak, "warmth", state.Tick)
		t.Logf("SC-002 sample t=%d: Oak warmth=%d asleep=%v since=%d classIntent=%d → NeglectDue=%v",
			tick, oak.Needs.Warmth, oak.Asleep, oak.Neglect.Since("warmth"), oak.Neglect.ClassIntent("warmth"), due)
		if got := oak.Neglect.Since("warmth"); got != bandEntry {
			t.Errorf("tick %d: warmth band anchor = %d, want %d (the recorded 636→0 crossing)", tick, got, bandEntry)
		}
		if oak.Asleep {
			// The sweep skips sleepers; the anchors must keep accruing so the
			// waker fires on its next heartbeat.
			if due {
				t.Errorf("tick %d: NeglectDue true for a sleeper", tick)
			}
			continue
		}
		if !due {
			t.Errorf("tick %d: NeglectDue(Oak, warmth) = false inside the awake death window (warmth %d, since %d, classIntent %d)",
				tick, oak.Needs.Warmth, oak.Neglect.Since("warmth"), oak.Neglect.ClassIntent("warmth"))
			continue
		}
		awakeTrue++
	}
	if awakeTrue == 0 {
		t.Error("no awake sample inside Oak's death window held the predicate — the detector would never have fired")
	}

	// --- Healthy windows and never-thrashing agents stay silent -------------
	// Oak's day 4 was productive shuttling (+723 food / +902 warmth WITH class
	// intents — research §1: flip volume ≠ pathology); Ash and Hazel never
	// thrashed at all. The predicate must be false for every need.
	healthySamples := []int64{300000, 504000}
	for _, tick := range healthySamples {
		state := replayTo(tick)
		for _, idx := range []int{ashIdx, hazelIdx} {
			a := &state.Agents[idx]
			for _, need := range []string{"food", "warmth", "rest"} {
				if sim.NeglectDue(a, need, state.Tick) {
					t.Errorf("tick %d: NeglectDue(%s, %s) = true — a never-thrashing agent must stay silent",
						tick, sim.AgentNames[idx], need)
				}
			}
		}
		if tick == 300000 {
			oak := &state.Agents[oakIdx]
			for _, need := range []string{"food", "warmth", "rest"} {
				if sim.NeglectDue(oak, need, state.Tick) {
					t.Errorf("tick %d: NeglectDue(Oak, %s) = true inside the labeled-healthy day-4 window", tick, need)
				}
			}
			t.Logf("SC-002 healthy t=%d: Oak needs=%+v warmthSince=%d warmthIntent=%d",
				tick, oak.Needs, oak.Neglect.Since("warmth"), oak.Neglect.ClassIntent("warmth"))
		}
	}
}

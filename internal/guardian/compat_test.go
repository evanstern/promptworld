package guardian

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/persona"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// --- T020: the SC-003 compat suite — a pre-052 world runs unchanged ---

// prefeatureWorldDir builds a world directory exactly as a pre-052 binary
// left it: the angel-voiced default charter, an angel-voiced soul file under
// the frozen metatron/ path, and a capabilities.json in the old vocabulary.
func prefeatureWorldDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "charter.md"), []byte(persona.LegacyDefaultCharter), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "metatron"), 0o755); err != nil {
		t.Fatal(err)
	}
	oldSoul := "# The soul of Metatron\n\n*The reign begins. The angel has seen nothing yet.*\n\n- day 1 06:00 — I sent a vision to Ash: \"wake\"\n"
	if err := os.WriteFile(filepath.Join(dir, "metatron", "soul.md"), []byte(oldSoul), 0o644); err != nil {
		t.Fatal(err)
	}
	caps := `{"tools": ["send_vision", "work_miracle"], "miracle_kinds": ["give_item"]}`
	if err := os.WriteFile(filepath.Join(dir, "capabilities.json"), []byte(caps), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// TestPrefeatureWorldOpensUnchanged (spec 052 US4 AS-1, SC-003): the new
// binary opens a pre-052 world with zero migration steps — the legacy
// charter still reads as the game-authored default, the old soul file is
// never rewritten, and the old capabilities.json vocabulary still grants.
func TestPrefeatureWorldOpensUnchanged(t *testing.T) {
	dir := prefeatureWorldDir(t)
	m := worldmap.Generate(42, 64, 64)
	state := sim.NewState(42, m)
	mt, err := New(&mockOrch{reply: "watching"}, &stateInjector{state: state, m: m}, &loopControlStub{}, m, 42, state.Marshal(), dir, testLoopRounds, testTurnTokens)
	if err != nil {
		t.Fatal(err)
	}
	mt.Close()

	// The legacy seed is still recognized as game-authored (never
	// reclassified player-authored on upgrade).
	s := mt.Status()
	if !s.CharterDefault {
		t.Error("a pre-052 world's untouched charter must still read as the default")
	}
	// The old soul file is history: byte-untouched (genesis only writes when
	// the file is absent).
	soul, err := os.ReadFile(filepath.Join(dir, "metatron", "soul.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(soul), "The soul of Metatron") {
		t.Error("pre-052 soul file was rewritten — history must never be rewritten")
	}
	// The old capabilities.json vocabulary (frozen tool ids) still grants.
	grant, notices := loadManifest(dir)
	if len(notices) != 0 {
		t.Errorf("old capabilities.json should load clean: %v", notices)
	}
	if !grant.allows("send_vision") || !grant.allows("work_miracle") || !grant.allowsKind("give_item") || grant.allowsKind("move") {
		t.Error("old capabilities.json grant vocabulary drifted")
	}
	// The charter_observed Default flag stays true for the legacy text.
	charter, _ := loadCharter(dir)
	if charter != persona.LegacyDefaultCharter {
		t.Error("legacy charter text mutated on load")
	}
}

// TestPrefeatureEventLogReplays (spec 052 US4 AS-1, SC-003): a recorded
// pre-052 event sequence — the frozen metatron.*/curriculum.* vocabulary —
// replays through the same reducer and reproduces state (charges, orders,
// charter fingerprint, memory text).
func TestPrefeatureEventLogReplays(t *testing.T) {
	m := worldmap.Generate(42, 64, 64)
	st := sim.NewState(42, m)
	startCharges := st.GuardianCharges

	events := []store.Event{
		{Tick: 10, Type: "metatron.nudged", Payload: json.RawMessage(`{"form":"vision","targets":[0],"text":"wake"}`)},
		{Tick: 10, Type: "agent.memory_added", Payload: json.RawMessage(`{"agent":0,"text":"You saw a vision: wake","salience":6,"subject":-1,"origin":"omen"}`)},
		{Tick: 20, Type: "metatron.charter_observed", Payload: json.RawMessage(`{"fingerprint":"ab12cd34ef56","default":true}`)},
		{Tick: 30, Type: "metatron.order_placed", Payload: json.RawMessage(`{"id":"ord-30-0","origin":"player","condition":"the fire goes out","action":"relight it","event_types":["sim.night_started"],"agent":-1,"placed_tick":30,"expires_tick":259230,"status":"active"}`)},
		{Tick: 40, Type: "metatron.charge_regenerated", Payload: json.RawMessage(`{}`)},
	}
	for _, e := range events {
		if err := st.Apply(e); err != nil {
			t.Fatalf("frozen event %s no longer replays: %v", e.Type, err)
		}
	}
	if st.GuardianCharges != startCharges-1+1 {
		t.Errorf("charge arithmetic drifted on replay: %d", st.GuardianCharges)
	}
	if len(st.GuardianOrders) != 1 || st.GuardianOrders[0].ID != "ord-30-0" || st.GuardianOrders[0].Status != "active" {
		t.Errorf("standing-order replay drifted: %+v", st.GuardianOrders)
	}
	if st.CharterFingerprint != "ab12cd34ef56" {
		t.Errorf("charter observation replay drifted: %q", st.CharterFingerprint)
	}
	if len(st.Agents) > 0 {
		found := false
		for _, mem := range st.Agents[0].Memories {
			if strings.Contains(mem.Text, "You saw a vision: wake") {
				found = true
			}
		}
		if !found {
			t.Error("recorded memory text drifted on replay")
		}
	}
}

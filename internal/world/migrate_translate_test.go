package world

// The translating-migration suite (spec 094 T005/T008, FR-003/FR-004,
// SC-001/SC-003): a seeded pre-rename fixture world (manifest v5, log format
// 1, metatron.* vocabulary) is refused by the new binary, migrates by
// translation, replays with a byte-identical state-hash sequence, and runs
// forward — plus the archive / never-overwrite / already-migrated /
// live-daemon guards.
//
// "Old semantics" replay is emulated in-binary: the pre-rename reducer is the
// current reducer modulo the bijective name map (T006's diff is name-only —
// SC-002's grep proves no other metatron.* reference survives in emit/apply
// code), so applying each source event under its translated name IS the old
// binary's fold. The harness then independently proves the migration wrote
// exactly that translation to disk: per-event byte identity of seq, tick,
// payload, and wall_time, with only the type column mapped.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

const (
	fixtureSeed = uint64(1337)
	fixtureSize = 64
)

// writeV5Manifest writes a pre-rename world.json by hand — world.Create
// stamps the CURRENT version, and the fixture must be exactly what a v5
// binary left behind.
func writeV5Manifest(t *testing.T, dir string, version int, terrainGen int) {
	t.Helper()
	m := map[string]any{
		"name":              "fixture",
		"seed":              fixtureSeed,
		"created_at":        "2026-07-01T00:00:00Z",
		"format_version":    version,
		"tick_game_seconds": 1,
		"map_width":         fixtureSize,
		"map_height":        fixtureSize,
	}
	if terrainGen != 0 {
		m["terrain_gen"] = terrainGen
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ManifestName), append(data, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// preRenameEvents builds a fixture history covering ALL 13 renamed types
// with payloads the reducer accepts, exactly as a v5 binary would have
// recorded them (old names; payload bytes are name-independent). The map and
// the genesis state locate a living villager, a forage tile, and a tree tile
// so the place-grant / move / terrain-removal arms validate.
func preRenameEvents(t *testing.T, m *worldmap.Map) []store.Event {
	t.Helper()
	probe := sim.NewState(fixtureSeed, m)
	ax, ay := probe.Agents[0].X, probe.Agents[0].Y

	var forageX, forageY, treeX, treeY = -1, -1, -1, -1
	var moveX, moveY = -1, -1
	for y := 0; y < m.H; y++ {
		for x := 0; x < m.W; x++ {
			switch m.At(x, y) {
			case worldmap.Forage:
				if forageX < 0 {
					forageX, forageY = x, y
				}
			case worldmap.Tree:
				if treeX < 0 {
					treeX, treeY = x, y
				}
			}
			if moveX < 0 && m.Passable(x, y) && !(x == ax && y == ay) {
				moveX, moveY = x, y
			}
		}
	}
	if forageX < 0 || treeX < 0 || moveX < 0 {
		t.Fatalf("fixture map lacks forage/tree/passable tiles (forage %d,%d tree %d,%d move %d,%d)",
			forageX, forageY, treeX, treeY, moveX, moveY)
	}

	watch := sim.SurvivalWatchDefs(20)[0]
	playerOrder := func(id string, tick int64) sim.GuardianOrder {
		return sim.GuardianOrder{
			ID: id, Origin: sim.GuardianOriginPlayer,
			Condition: "watch the fire", Action: "warn the village",
			EventTypes: []string{"sim.fire_burned_out"}, Agent: -1,
			PlacedTick: tick, ExpiresTick: tick + 3*24*3600, Status: "active",
		}
	}

	return []store.Event{
		{Tick: 0, Type: "world.created", Payload: mustJSON(t, sim.WorldCreatedPayload{Name: "fixture", Seed: fixtureSeed})},
		{Tick: 10, Type: "metatron.charter_observed", Payload: mustJSON(t, sim.CharterObservedPayload{Fingerprint: "fp-default", Default: true})},
		{Tick: 20, Type: "metatron.order_placed", Payload: mustJSON(t, watch.PlacedPayload())},
		{Tick: 30, Type: "metatron.order_placed", Payload: mustJSON(t, playerOrder("ord-1", 30).PlacedPayload())},
		{Tick: 40, Type: "metatron.order_placed", Payload: mustJSON(t, playerOrder("ord-2", 40).PlacedPayload())},
		{Tick: 21600, Type: "metatron.charge_regenerated", Payload: json.RawMessage(`{}`)},
		{Tick: 21700, Type: "metatron.nudged", Payload: mustJSON(t, sim.GuardianNudgedPayload{Form: "vision", Targets: []sim.AgentRef{sim.Ref(0)}, Text: "wake"})},
		{Tick: 21800, Type: "metatron.place_revealed", Payload: mustJSON(t, sim.PlaceRevealedPayload{Agent: sim.Ref(0), Facts: []sim.PlaceFact{{Kind: "forage", X: forageX, Y: forageY}}})},
		{Tick: 22000, Type: "metatron.order_triggered", Payload: mustJSON(t, sim.OrderTriggeredPayload{ID: "ord-1", MatchedType: "sim.fire_burned_out", MatchedTick: 21990})},
		{Tick: 22100, Type: "metatron.order_cancelled", Payload: mustJSON(t, sim.OrderIDPayload{ID: "ord-2"})},
		{Tick: 22200, Type: "metatron.order_placed", Payload: mustJSON(t, playerOrder("ord-3", 22200).PlacedPayload())},
		{Tick: 22300, Type: "metatron.order_expired", Payload: mustJSON(t, sim.OrderIDPayload{ID: "ord-3"})},
		{Tick: 22400, Type: "metatron.skills_observed", Payload: mustJSON(t, sim.SkillsObservedPayload{Fingerprint: "fp-skills", Names: []string{"watch.md"}})},
		{Tick: 22500, Type: "metatron.item_granted", Payload: mustJSON(t, sim.ItemGrantedPayload{Agent: sim.Ref(0), Kind: "wood", Qty: 2, Gratis: true})},
		{Tick: 22600, Type: "metatron.entity_moved", Payload: mustJSON(t, sim.EntityMovedPayload{Class: "villager", X: ax, Y: ay, ToX: moveX, ToY: moveY, Gratis: true})},
		{Tick: 22700, Type: "metatron.entity_removed", Payload: mustJSON(t, sim.EntityRemovedPayload{Class: "terrain", X: treeX, Y: treeY, Gratis: true})},
		{Tick: 22800, Type: "metatron.time_snapped", Payload: mustJSON(t, sim.TimeSnappedPayload{ToTick: 40000, Gratis: true})},
	}
}

// buildV5Fixture writes a complete pre-rename world dir: v5 manifest,
// world.db carrying the metatron.* history (log format 1 — deliberately
// unstamped) and a mid-history covering snapshot, exactly what a cleanly
// stopped v5 world looks like.
func buildV5Fixture(t *testing.T) (dir string, events []store.Event, snapSeq int64) {
	t.Helper()
	dir = t.TempDir()
	writeV5Manifest(t, dir, 5, worldmap.GenMarshSand)
	m := worldmap.GenerateV(fixtureSeed, fixtureSize, fixtureSize, worldmap.GenMarshSand)
	events = preRenameEvents(t, m)

	st, err := store.Open(filepath.Join(dir, "world.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.AppendEvents(events); err != nil {
		t.Fatal(err)
	}

	// A covering snapshot mid-history (the old binary's shutdown snapshot
	// shape): state folded through the OLD semantics up to seq N-3, so the
	// migrated world boots through the snapshot + translated-tail path.
	snapSeq = int64(len(events) - 3)
	state := sim.NewState(fixtureSeed, m)
	for _, e := range events[:snapSeq] {
		if err := state.Apply(renamed(e)); err != nil {
			t.Fatalf("fixture event %s (seq %d) does not apply: %v", e.Type, e.Seq, err)
		}
		if e.Tick > state.Tick {
			state.Tick = e.Tick
		}
	}
	if err := st.SaveSnapshot(state.Tick, snapSeq, state.Marshal()); err != nil {
		t.Fatal(err)
	}
	return dir, events, snapSeq
}

// renamed maps one event through the v1→v2 table — the in-binary emulation
// of the old reducer (see the file comment).
func renamed(e store.Event) store.Event {
	if to, ok := sim.LogFormatV1Renames[e.Type]; ok {
		e.Type = to
	}
	return e
}

func hashOf(state *sim.State) string {
	sum := sha256.Sum256(state.Marshal())
	return hex.EncodeToString(sum[:])
}

// hashSequence folds events into a fresh genesis state, recording the state
// hash after every apply; translate=true routes each event through the
// rename map first (old-semantics emulation over the source log).
func hashSequence(t *testing.T, m *worldmap.Map, events []store.Event, translate bool) []string {
	t.Helper()
	state := sim.NewState(fixtureSeed, m)
	hashes := make([]string, 0, len(events))
	for _, e := range events {
		if translate {
			e = renamed(e)
		}
		if err := state.Apply(e); err != nil {
			t.Fatalf("apply %s (seq %d): %v", e.Type, e.Seq, err)
		}
		if e.Tick > state.Tick {
			state.Tick = e.Tick
		}
		hashes = append(hashes, hashOf(state))
	}
	return hashes
}

// TestTranslatingMigrationByteIdentity is the FR-004 harness + the SC-001
// end-to-end demo: refusal before, translation, per-event byte identity,
// state-hash-sequence identity, and the world running forward after.
func TestTranslatingMigrationByteIdentity(t *testing.T) {
	dir, srcEvents, snapSeq := buildV5Fixture(t)
	m := worldmap.GenerateV(fixtureSeed, fixtureSize, fixtureSize, worldmap.GenMarshSand)

	// US3.4 / SC-003: the unmigrated pre-rename world is refused by Open
	// with the migrate hint — never silently replayed under wrong names.
	if _, err := Open(dir); err == nil {
		t.Fatal("unmigrated v5 world opened without refusal")
	} else {
		var mismatch *ErrFormatVersionMismatch
		if !errors.As(err, &mismatch) {
			t.Fatalf("Open refusal is not ErrFormatVersionMismatch: %v", err)
		}
		if !strings.Contains(err.Error(), "promptworld migrate") {
			t.Fatalf("refusal lacks the migrate hint: %v", err)
		}
	}
	// The log itself refuses too (FR-002, defense in depth): unstamped +
	// non-empty ⇒ implicit legacy ⇒ too old.
	preSt, err := store.Open(filepath.Join(dir, "world.db"))
	if err != nil {
		t.Fatal(err)
	}
	verr := preSt.VerifyLogFormat()
	preSt.Close()
	var tooOld *store.ErrLogFormatTooOld
	if !errors.As(verr, &tooOld) {
		t.Fatalf("pre-migration log verify = %v, want ErrLogFormatTooOld", verr)
	}

	// Old-semantics replay of the SOURCE log (see file comment).
	oldHashes := hashSequence(t, m, srcEvents, true)

	// Migrate: the translation.
	res, err := Migrate(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Translated {
		t.Fatal("v5 source must take the translation mode")
	}
	if res.SourceEvents != int64(len(srcEvents)) {
		t.Errorf("SourceEvents = %d, want %d", res.SourceEvents, len(srcEvents))
	}
	if res.ArchivePath != filepath.Join(dir, "world.v5.db") {
		t.Errorf("archive at %s, want world.v5.db", res.ArchivePath)
	}

	// The archive holds the ORIGINAL, untranslated history.
	arch, err := store.Open(res.ArchivePath)
	if err != nil {
		t.Fatal(err)
	}
	archEvents, err := arch.EventsSince(0, 0)
	arch.Close()
	if err != nil {
		t.Fatal(err)
	}
	if len(archEvents) != len(srcEvents) {
		t.Fatalf("archive carries %d events, want %d", len(archEvents), len(srcEvents))
	}
	if archEvents[1].Type != "metatron.charter_observed" {
		t.Errorf("archive event 2 type = %s — the archive must keep old names", archEvents[1].Type)
	}

	// The migrated world opens (manifest v6) and its log is stamped v2.
	w, err := Open(dir)
	if err != nil {
		t.Fatalf("migrated world refused: %v", err)
	}
	if w.Manifest.FormatVersion != FormatVersion {
		t.Errorf("manifest format_version = %d, want %d", w.Manifest.FormatVersion, FormatVersion)
	}
	st, err := store.Open(w.DBPath())
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.VerifyLogFormat(); err != nil {
		t.Errorf("translated log fails the format gate: %v", err)
	}
	if v, _ := st.GetMeta("format_version"); v != strconv.Itoa(FormatVersion) {
		t.Errorf("meta format_version mirror = %q, want %d", v, FormatVersion)
	}

	// Per-event byte identity (FR-004's disk half): same seq, tick, payload,
	// wall_time; type mapped and ONLY mapped.
	gotEvents, err := st.EventsSince(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(gotEvents) != len(srcEvents) {
		t.Fatalf("translated log carries %d events, want %d", len(gotEvents), len(srcEvents))
	}
	for i, got := range gotEvents {
		src := srcEvents[i]
		want := renamed(src)
		if got.Seq != src.Seq || got.Tick != src.Tick || got.Type != want.Type ||
			string(got.Payload) != string(src.Payload) || got.WallTime != src.WallTime {
			t.Errorf("event %d diverged:\n got  {seq %d tick %d %s %s %s}\n want {seq %d tick %d %s %s %s}",
				i, got.Seq, got.Tick, got.Type, got.Payload, got.WallTime,
				src.Seq, src.Tick, want.Type, src.Payload, src.WallTime)
		}
		if strings.HasPrefix(got.Type, "metatron.") {
			t.Errorf("event %d kept old name %s", i, got.Type)
		}
	}

	// State-hash-sequence identity (FR-004, US2.2): replay(source, old
	// semantics) == replay(translated, new semantics), hash for hash.
	newHashes := hashSequence(t, m, gotEvents, false)
	if len(newHashes) != len(oldHashes) {
		t.Fatalf("hash sequences differ in length: %d vs %d", len(newHashes), len(oldHashes))
	}
	for i := range newHashes {
		if newHashes[i] != oldHashes[i] {
			t.Fatalf("state hash diverges at event %d (%s): old %s new %s",
				i, gotEvents[i].Type, oldHashes[i], newHashes[i])
		}
	}

	// The snapshot carried over and the boot path (snapshot + translated
	// tail) reproduces the full-replay state — then the world RUNS FORWARD:
	// a new-vocabulary event emitted by this build applies cleanly on top
	// (SC-001's "runs forward on the new binary").
	snap, err := st.LatestValidSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	if snap == nil || snap.Seq != snapSeq {
		t.Fatalf("carried snapshot = %+v, want seq %d", snap, snapSeq)
	}
	boot := sim.NewState(fixtureSeed, m)
	if err := json.Unmarshal(snap.State, boot); err != nil {
		t.Fatal(err)
	}
	err = st.ReplayEvents(snap.Seq, func(e store.Event) error {
		if err := boot.Apply(e); err != nil {
			return err
		}
		if e.Tick > boot.Tick {
			boot.Tick = e.Tick
		}
		return nil
	})
	if err != nil {
		t.Fatalf("boot-path replay over the translated tail: %v", err)
	}
	if got := hashOf(boot); got != newHashes[len(newHashes)-1] {
		t.Fatalf("snapshot+tail boot state %s != full-replay state %s", got, newHashes[len(newHashes)-1])
	}
	forward := store.Event{Tick: boot.Tick + 100, Type: "guardian.charge_regenerated", Payload: json.RawMessage(`{}`)}
	if err := st.AppendEvents([]store.Event{forward}); err != nil {
		t.Fatal(err)
	}
	if err := boot.Apply(forward); err != nil {
		t.Fatalf("migrated world cannot run forward: %v", err)
	}

	// Idempotence (US2.3): a second run has nothing to migrate.
	if _, err := Migrate(dir); err == nil {
		t.Fatal("second migration ran — a current world must refuse")
	} else if !strings.Contains(err.Error(), "nothing to migrate") {
		t.Fatalf("second migration refusal = %v, want the nothing-to-migrate guard", err)
	}
	// Never-overwrite (the crash posture): a manifest rolled back to the
	// source version — the archive-present/manifest-unbumped recovery state —
	// hits the archive guard, not a second archival.
	writeV5Manifest(t, dir, 5, worldmap.GenMarshSand)
	if _, err := Migrate(dir); err == nil || !strings.Contains(err.Error(), "already migrated") {
		t.Fatalf("archive guard did not hold: %v", err)
	}
}

// TestTranslatingMigrationRefusesLiveDaemon: the pidfile liveness guard holds
// for the translation mode exactly as for the snapshot-cut.
func TestTranslatingMigrationRefusesLiveDaemon(t *testing.T) {
	dir, _, _ := buildV5Fixture(t)
	if err := os.WriteFile(filepath.Join(dir, "daemon.pid"), []byte(strconv.Itoa(os.Getpid())+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Migrate(dir); err == nil {
		t.Fatal("migration ran under a live daemon")
	} else if !strings.Contains(err.Error(), "daemon is running") {
		t.Fatalf("live-daemon refusal = %v", err)
	}
}

// TestTranslatingMigrationV4Source: a v4 world (content-identical log to v5 —
// 4→5 was manifest-only) translates too, archiving under its own source name.
func TestTranslatingMigrationV4Source(t *testing.T) {
	dir := t.TempDir()
	writeV5Manifest(t, dir, 4, 0) // v4: pre-terrain_gen, legacy generation
	st, err := store.Open(filepath.Join(dir, "world.db"))
	if err != nil {
		t.Fatal(err)
	}
	events := []store.Event{
		{Tick: 0, Type: "world.created", Payload: mustJSON(t, sim.WorldCreatedPayload{Name: "fixture", Seed: fixtureSeed})},
		{Tick: 10, Type: "metatron.charter_observed", Payload: mustJSON(t, sim.CharterObservedPayload{Fingerprint: "fp", Default: true})},
		{Tick: 20, Type: "metatron.order_placed", Payload: mustJSON(t, sim.SurvivalWatchDefs(20)[0].PlacedPayload())},
	}
	if err := st.AppendEvents(events); err != nil {
		t.Fatal(err)
	}
	st.Close()

	res, err := Migrate(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Translated {
		t.Fatal("v4 source must take the translation mode")
	}
	if res.ArchivePath != filepath.Join(dir, "world.v4.db") {
		t.Errorf("archive at %s, want world.v4.db", res.ArchivePath)
	}
	w, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if w.Manifest.TerrainGen != worldmap.GenLegacy {
		t.Errorf("a migrated v4 world must keep legacy terrain (got %d)", w.Manifest.TerrainGen)
	}
	got, err := store.Open(w.DBPath())
	if err != nil {
		t.Fatal(err)
	}
	defer got.Close()
	evs, err := got.EventsSince(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if evs[1].Type != "guardian.charter_observed" || evs[2].Type != "guardian.order_placed" {
		t.Errorf("translated types = %s, %s", evs[1].Type, evs[2].Type)
	}
}

// TestRenameMapMatchesCatalog pins the rename table to the payload catalog:
// every target is a cataloged current type, no source survives in the
// catalog, and the table covers exactly the 13 renamed types — the map and
// the vocabulary can never drift apart (SC-002's structural half).
func TestRenameMapMatchesCatalog(t *testing.T) {
	if len(sim.LogFormatV1Renames) != 13 {
		t.Errorf("rename table has %d entries, want 13 (spec 094 research.md D3)", len(sim.LogFormatV1Renames))
	}
	for from, to := range sim.LogFormatV1Renames {
		if !strings.HasPrefix(from, "metatron.") {
			t.Errorf("rename source %q is not metatron.*", from)
		}
		if fmt.Sprintf("guardian.%s", strings.TrimPrefix(from, "metatron.")) != to {
			t.Errorf("rename %q → %q is not the pure namespace swap", from, to)
		}
		if _, ok := sim.PayloadCatalog[to]; !ok {
			t.Errorf("rename target %q is not in the payload catalog", to)
		}
		if _, ok := sim.PayloadCatalog[from]; ok {
			t.Errorf("renamed source %q still in the payload catalog", from)
		}
	}
	for cataloged := range sim.PayloadCatalog {
		if strings.HasPrefix(cataloged, "metatron.") {
			t.Errorf("payload catalog still carries %q", cataloged)
		}
	}
}

// TestManifestMismatchDirections (FR-002 at the manifest level): older world
// ⇒ migrate hint; newer world ⇒ upgrade posture.
func TestManifestMismatchDirections(t *testing.T) {
	older := t.TempDir()
	writeV5Manifest(t, older, 5, worldmap.GenMarshSand)
	if _, err := Open(older); err == nil || !strings.Contains(err.Error(), "promptworld migrate") {
		t.Errorf("older-world refusal = %v, want the migrate hint", err)
	}
	newer := t.TempDir()
	writeV5Manifest(t, newer, FormatVersion+1, worldmap.GenMarshSand)
	if _, err := Open(newer); err == nil || !strings.Contains(err.Error(), "upgrade promptworld") {
		t.Errorf("newer-world refusal = %v, want the upgrade posture", err)
	}
}

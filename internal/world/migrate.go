package world

// The offline migration driver, in two modes (the spec 094 decision rule):
//
//   - SNAPSHOT-CUT (semantic breaks — spec 012 US6 for v1→v2 while the land
//     resets, research R10; spec 013 for v2→v3 which preserves people AND
//     land, research R3): never replays old events under new rules — it reads
//     the source world's covering snapshot, transforms it (internal/sim), and
//     writes a fresh log whose single world.migrated event carries the full
//     transformed state, so the log alone reproduces the migrated world
//     byte-identically. History is archived, not carried.
//
//   - TRANSLATION (pure renames — spec 094, the guardian rename): rewrites
//     the log's type column through sim.LogFormatV1Renames with EVERY event,
//     tick, and payload preserved byte-for-byte, so the full history
//     survives. Used for v4/v5 sources (content-identical logs; 4→5 was
//     manifest-only, spec 068), whose vocabulary is the only thing v6
//     changed.
//
// DOCTRINE: a persisted-name change translates; a semantic break (payload
// meaning, reducer re-derivation) snapshot-cuts. Both are client-side,
// offline operations — the daemon must be stopped — and both archive the
// source DB under its source-format name (world.v1.db … world.v5.db), never
// overwriting an existing archive. An older world chains snapshot-cut steps
// to the v4/v5 state shape, then lands on a fresh already-current log.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// MigrateResult is the human-facing summary of a completed migration.
type MigrateResult struct {
	Name          string
	Seed          uint64
	AgentsCarried int
	Tick          int64 // the continuation tick (carried from v1)
	SourceEvents  int64 // v1 events archived in world.v1.db
	ArchivePath   string
	// Translated marks the translation mode (spec 094): the log carried over
	// event-for-event with only type names rewritten — SourceEvents counts
	// the events preserved (== the translated log's length), and Tick is the
	// head tick the world resumes at.
	Translated bool
}

// OpenForMigration loads a world directory WITHOUT the current version gate,
// for the sole purpose of migrating it. It admits any older supported source
// format (v1 through v5) and refuses everything else: an already-current (or future)
// world has nothing to migrate, and a corrupt manifest is refused exactly as
// Open would. Map dimensions are defaulted identically to Open so a regenerated
// map matches what the daemon will boot.
func OpenForMigration(dir string) (*World, error) {
	data, err := os.ReadFile(filepath.Join(dir, ManifestName))
	if err != nil {
		return nil, fmt.Errorf("not a world directory (missing %s): %w", ManifestName, err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("corrupt %s: %w", ManifestName, err)
	}
	if m.FormatVersion == FormatVersion {
		return nil, fmt.Errorf("world %q is already format_version %d — nothing to migrate", m.Name, FormatVersion)
	}
	if m.FormatVersion < 1 || m.FormatVersion >= FormatVersion {
		return nil, fmt.Errorf("world %q is format_version %d; only v1 through v%d worlds can be migrated to v%d", m.Name, m.FormatVersion, FormatVersion-1, FormatVersion)
	}
	if m.TickGameSeconds != 1 {
		return nil, fmt.Errorf("tick_game_seconds %d unsupported (must be 1)", m.TickGameSeconds)
	}
	if m.MapWidth <= 0 {
		m.MapWidth = worldmap.DefaultSize
	}
	if m.MapHeight <= 0 {
		m.MapHeight = worldmap.DefaultSize
	}
	return &World{Dir: dir, Manifest: m}, nil
}

// V1DBPath … V5DBPath are the archived original databases — the archive name
// is keyed to the SOURCE format so a v1→…→6 run parks world.v1.db, a v5→6 run
// world.v5.db, and so on. The archive's existence is the already-migrated
// guard for that source format, and restoring is "delete world.db, rename
// this back, reset the manifest to the source version".
func (w *World) V1DBPath() string { return filepath.Join(w.Dir, "world.v1.db") }
func (w *World) V2DBPath() string { return filepath.Join(w.Dir, "world.v2.db") }
func (w *World) V3DBPath() string { return filepath.Join(w.Dir, "world.v3.db") }
func (w *World) V4DBPath() string { return filepath.Join(w.Dir, "world.v4.db") }
func (w *World) V5DBPath() string { return filepath.Join(w.Dir, "world.v5.db") }

// archiveDBPath is the archive name for this world's SOURCE format version.
func (w *World) archiveDBPath() string {
	switch w.Manifest.FormatVersion {
	case 5:
		return w.V5DBPath()
	case 4:
		return w.V4DBPath()
	case 3:
		return w.V3DBPath()
	case 2:
		return w.V2DBPath()
	}
	return w.V1DBPath()
}

// Migrate performs the whole v1→v2 migration in place (research R10). The
// archive is sacred: world.db is renamed to world.v1.db (never deleted), and
// the migration refuses to run if that archive already exists. It refuses a
// running daemon and an un-covered event tail (no clean-shutdown snapshot),
// leaving the world untouched in both cases.
func Migrate(dir string) (*MigrateResult, error) {
	w, err := OpenForMigration(dir)
	if err != nil {
		return nil, err
	}

	// Refuse a live daemon: migration rewrites the database out from under any
	// process holding it. The pidfile liveness check is version-gate-free (the
	// v1 world cannot be world.Open'd under this build).
	if running, pid := daemonAlive(w); running {
		return nil, fmt.Errorf("daemon is running (pid %d) — stop it first: promptworld stop %s", pid, dir)
	}

	// v4/v5 → v6 is the TRANSLATION mode (spec 094): those logs differ from
	// v6 only in the guardian event-type vocabulary (4→5 was manifest-only,
	// spec 068, so both speak log format 1). The full history carries over —
	// no snapshot-cut, no covering-snapshot requirement.
	if w.Manifest.FormatVersion >= 4 {
		return migrateTranslate(w)
	}

	// Already-migrated guard: the archive is never overwritten (FR-025). The
	// guard is on the SOURCE-format archive (world.v1.db for a v1 source,
	// world.v2.db for a v2 source), so a v2 world produced by an earlier v1
	// migration — which would carry a stale world.v1.db — is still migratable
	// to v3.
	archivePath := w.archiveDBPath()
	if _, err := os.Stat(archivePath); err == nil {
		return nil, fmt.Errorf("this world is already migrated (%s exists); the archive is never overwritten", filepath.Base(archivePath))
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	// Read the v1 covering snapshot. The migration NEVER replays v1 events under
	// v2 rules (FR-024) — the clean-shutdown snapshot is the only v1 state it
	// reads.
	st, err := store.Open(w.DBPath())
	if err != nil {
		return nil, err
	}
	if cerr := st.CheckContiguity(); cerr != nil {
		st.Close()
		return nil, cerr
	}
	maxSeq := st.LastSeq()
	snap, err := st.LatestValidSnapshot()
	if err != nil {
		st.Close()
		return nil, err
	}
	if snap == nil {
		st.Close()
		return nil, migrateNeedsCleanStop(dir, "this world has no valid snapshot")
	}
	// The clean-shutdown guarantee (FR-024) is that the covering snapshot holds
	// all *sim* state. A real v1 daemon appends its `daemon.stopped` bookkeeping
	// AFTER the shutdown snapshot, so a cleanly-stopped world normally has a
	// one-event tail past snap.Seq — observed on myworld-01: seq 114507 trailing
	// a 114506-covering snapshot. `daemon.*` events are reducer no-ops carrying
	// zero sim state, so a tail consisting only of them is tolerated (and simply
	// dropped — its information content is nil, nothing to carry into the v2
	// log). Any sim-affecting event past the snapshot is un-snapshotted history
	// and still refuses.
	tail, err := st.EventsSince(snap.Seq, 0)
	if err != nil {
		st.Close()
		return nil, err
	}
	for _, e := range tail {
		if !strings.HasPrefix(e.Type, "daemon.") {
			st.Close()
			return nil, migrateNeedsCleanStop(dir,
				fmt.Sprintf("the latest valid snapshot covers seq %d but the log runs to seq %d with a sim-affecting event (%s at seq %d) past it (an unclean stop left un-snapshotted history)",
					snap.Seq, maxSeq, e.Type, e.Seq))
		}
	}

	// Transform the covering-snapshot state to the current format (the v4
	// state shape — v5 changed no state, only the manifest gate above),
	// chaining every step from the source version in one run: v1→v2 re-places
	// souls on the v2 regeneration of the same seed (w.Map() uses this build's
	// generator, so they stand on passable v2 tiles, rock outcrops included);
	// v2→v3 carries everything verbatim and spills any over-cap carry; v3→v4
	// grants each villager its mental map (explored home area + witnessed
	// facts for current structures/piles — spec 041, research D7). srcTick is
	// the carried continuation tick in every path.
	var finalState *sim.State
	var srcTick int64
	switch w.Manifest.FormatVersion {
	case 1:
		var v2state *sim.State
		v2state, srcTick, err = sim.TransformV1Snapshot(snap.State, w.Map())
		if err != nil {
			st.Close()
			return nil, err
		}
		finalState = sim.TransformV3State(sim.TransformV2State(v2state), w.Map())
	case 2:
		var v3state *sim.State
		v3state, srcTick, err = sim.TransformV2Snapshot(snap.State)
		if err != nil {
			st.Close()
			return nil, err
		}
		finalState = sim.TransformV3State(v3state, w.Map())
	case 3:
		finalState, srcTick, err = sim.TransformV3Snapshot(snap.State, w.Map())
		if err != nil {
			st.Close()
			return nil, err
		}
	default:
		st.Close()
		return nil, fmt.Errorf("unsupported source format_version %d", w.Manifest.FormatVersion)
	}
	if err := st.Close(); err != nil {
		return nil, err
	}

	// Archive the original database (and any WAL sidecars) intact under the
	// source-format archive name. This is the point of no easy return, so
	// everything that could refuse has already run.
	if err := archiveDB(w.DBPath(), archivePath); err != nil {
		return nil, err
	}

	// Fresh log: world.created (same name/seed) then world.migrated carrying the
	// full transformed state. Both stamped at the continuation tick.
	fresh, err := store.Open(w.DBPath())
	if err != nil {
		return nil, err
	}
	defer fresh.Close()

	// A snapshot-cut's fresh log is born current (spec 094): its only events
	// (world.created + world.migrated) predate no vocabulary, so it is
	// stamped with today's log format directly.
	if err := fresh.StampLogFormat(); err != nil {
		return nil, err
	}

	createdPayload, err := json.Marshal(sim.WorldCreatedPayload{Name: w.Manifest.Name, Seed: w.Manifest.Seed})
	if err != nil {
		return nil, err
	}
	migratedPayload, err := json.Marshal(sim.WorldMigratedPayload{
		FromFormat:   w.Manifest.FormatVersion,
		SourceEvents: maxSeq,
		SourceTick:   srcTick,
		State:        *finalState,
	})
	if err != nil {
		return nil, err
	}
	if err := fresh.AppendEvents([]store.Event{
		{Tick: srcTick, Type: "world.created", Payload: createdPayload},
		{Tick: srcTick, Type: "world.migrated", Payload: migratedPayload},
	}); err != nil {
		return nil, err
	}

	// Initial snapshot at the migrated tick: the covering snapshot of the fresh
	// log. Deleting it and replaying (world.created → world.migrated) must
	// reproduce this exact state — the determinism half of SC-007.
	if err := fresh.SaveSnapshot(srcTick, fresh.LastSeq(), finalState.Marshal()); err != nil {
		return nil, err
	}

	// Bump the manifest last: with the manifest still at the source version, a
	// crash between the archive and here leaves a recoverable state (the
	// source-format archive present, manifest unbumped — restore is the same
	// rename-back).
	w.Manifest.FormatVersion = FormatVersion
	if err := writeManifest(w); err != nil {
		return nil, err
	}

	return &MigrateResult{
		Name:          w.Manifest.Name,
		Seed:          w.Manifest.Seed,
		AgentsCarried: len(finalState.Agents),
		Tick:          srcTick,
		SourceEvents:  maxSeq,
		ArchivePath:   archivePath,
	}, nil
}

// migrateTranslate is the translating migration (spec 094 FR-003): rewrite a
// v4/v5 log into the current vocabulary — every event's type mapped through
// sim.LogFormatV1Renames, every seq, tick, payload, and wall_time preserved
// byte-for-byte — so history survives a pure rename. The translated DB is
// built at a temp path and only swapped in after the source is archived, so
// no failure mode leaves a half-written world.db as the live log. Guards
// (live daemon refused by the caller; archive never overwritten;
// already-migrated = archive exists) match the snapshot-cut's contract.
func migrateTranslate(w *World) (*MigrateResult, error) {
	archivePath := w.archiveDBPath()
	if _, err := os.Stat(archivePath); err == nil {
		return nil, fmt.Errorf("this world is already migrated (%s exists); the archive is never overwritten", filepath.Base(archivePath))
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	src, err := store.Open(w.DBPath())
	if err != nil {
		return nil, err
	}
	if cerr := src.CheckContiguity(); cerr != nil {
		src.Close()
		return nil, cerr
	}
	srcEvents := src.LastSeq()

	// Build the translated log at a temp path inside the world dir (same
	// filesystem, so the final rename is atomic). A leftover from a crashed
	// earlier run is dead weight — remove it and its WAL sidecars.
	tmpPath := w.DBPath() + ".translating"
	for _, p := range []string{tmpPath, tmpPath + "-wal", tmpPath + "-shm"} {
		if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
			src.Close()
			return nil, err
		}
	}
	fresh, err := store.Open(tmpPath)
	if err != nil {
		src.Close()
		return nil, err
	}
	cleanupTmp := func() {
		fresh.Close()
		src.Close()
		os.Remove(tmpPath)
		os.Remove(tmpPath + "-wal")
		os.Remove(tmpPath + "-shm")
	}

	// Meta rows carry over verbatim (the fork ceremony's wallet-inheritance
	// posture: the seed mirror, llm_spend_* budget truth) — then the
	// format_version mirror is updated so validateMeta's manifest cross-check
	// holds at the next boot, and the log format stamp is written LAST so it
	// wins over any (impossible today) stale log_format_version row copied
	// from the source.
	meta, err := src.MetaByPrefix("")
	if err != nil {
		cleanupTmp()
		return nil, err
	}
	for k, v := range meta {
		if err := fresh.SetMeta(k, v); err != nil {
			cleanupTmp()
			return nil, err
		}
	}
	if err := fresh.SetMeta("format_version", strconv.Itoa(FormatVersion)); err != nil {
		cleanupTmp()
		return nil, err
	}
	if err := fresh.StampLogFormat(); err != nil {
		cleanupTmp()
		return nil, err
	}

	// Stream every event across, type mapped, everything else verbatim.
	// AppendEvents assigns contiguous seqs from 1 and the source is
	// contiguity-checked, so seqs reproduce exactly; a non-empty WallTime
	// rides through untouched.
	const batchSize = 1024
	batch := make([]store.Event, 0, batchSize)
	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		if err := fresh.AppendEvents(batch); err != nil {
			return err
		}
		batch = batch[:0]
		return nil
	}
	err = src.ReplayEvents(0, func(e store.Event) error {
		if renamed, ok := sim.LogFormatV1Renames[e.Type]; ok {
			e.Type = renamed
		}
		batch = append(batch, e)
		if len(batch) == batchSize {
			return flush()
		}
		return nil
	})
	if err == nil {
		err = flush()
	}
	if err != nil {
		cleanupTmp()
		return nil, err
	}
	if fresh.LastSeq() != srcEvents {
		cleanupTmp()
		return nil, fmt.Errorf("translation carried %d events, want %d — refusing to continue", fresh.LastSeq(), srcEvents)
	}

	// Snapshots are derived accelerators: the latest verified one carries
	// over (SaveSnapshot recomputes the same hash over the same bytes), so
	// the migrated world boots from exactly the state the source world last
	// cut; older snapshot rows are regenerable history the archive retains.
	snap, err := src.LatestValidSnapshot()
	if err != nil {
		cleanupTmp()
		return nil, err
	}
	if snap != nil {
		if err := fresh.SaveSnapshot(snap.Tick, snap.Seq, snap.State); err != nil {
			cleanupTmp()
			return nil, err
		}
	}

	// Verification before the swap — emulate exactly what daemon boot will
	// do (recoverState: latest snapshot + tail replay; genesis replay when
	// no snapshot exists). A translated log the current reducer rejects
	// must never become the live world.db.
	state := sim.NewState(w.Manifest.Seed, w.Map())
	var since int64
	if snap != nil {
		if err := json.Unmarshal(snap.State, state); err != nil {
			cleanupTmp()
			return nil, fmt.Errorf("translated snapshot unreadable: %w", err)
		}
		since = snap.Seq
	}
	err = fresh.ReplayEvents(since, func(e store.Event) error {
		if err := state.Apply(e); err != nil {
			return fmt.Errorf("translated log fails replay at seq %d (%s): %w", e.Seq, e.Type, err)
		}
		if e.Tick > state.Tick {
			state.Tick = e.Tick
		}
		return nil
	})
	if err != nil {
		cleanupTmp()
		return nil, err
	}

	if err := fresh.Close(); err != nil {
		src.Close()
		return nil, err
	}
	if err := src.Close(); err != nil {
		return nil, err
	}

	// The swap: archive the source intact (the point of no easy return —
	// everything that could refuse has already run), then move the verified
	// translation into place.
	if err := archiveDB(w.DBPath(), archivePath); err != nil {
		return nil, err
	}
	if err := archiveDB(tmpPath, w.DBPath()); err != nil {
		return nil, err
	}

	// Bump the manifest last (the snapshot-cut's crash posture: archive
	// present + manifest unbumped is recoverable by renaming back).
	w.Manifest.FormatVersion = FormatVersion
	if err := writeManifest(w); err != nil {
		return nil, err
	}

	return &MigrateResult{
		Name:          w.Manifest.Name,
		Seed:          w.Manifest.Seed,
		AgentsCarried: len(state.Agents),
		Tick:          state.Tick,
		SourceEvents:  srcEvents,
		ArchivePath:   archivePath,
		Translated:    true,
	}, nil
}

// migrateNeedsCleanStop wraps the "no covering snapshot" refusals with the
// remedy: a clean start+stop under the source-format binary produces the
// shutdown snapshot migration relies on (FR-024).
func migrateNeedsCleanStop(dir, why string) error {
	return fmt.Errorf("%s — start and stop this world once with its own binary so a covering shutdown snapshot exists, then re-run: promptworld migrate %s", why, dir)
}

// archiveDB renames the live database (and any WAL/SHM sidecars) to the archive
// name. Moving the sidecars matters twice: the archive stays a complete,
// restorable database, and the fresh world.db is not corrupted by a stale WAL
// from the old one.
func archiveDB(dbPath, archivePath string) error {
	if err := os.Rename(dbPath, archivePath); err != nil {
		return fmt.Errorf("archive %s: %w", filepath.Base(dbPath), err)
	}
	for _, suffix := range []string{"-wal", "-shm"} {
		src := dbPath + suffix
		if _, err := os.Stat(src); err == nil {
			if err := os.Rename(src, archivePath+suffix); err != nil {
				return fmt.Errorf("archive %s: %w", filepath.Base(src), err)
			}
		}
	}
	return nil
}

// writeManifest rewrites world.json from the (mutated) manifest, matching
// Create's indentation so the file stays human-readable and diff-friendly.
func writeManifest(w *World) error {
	data, err := json.MarshalIndent(w.Manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(w.Dir, ManifestName), append(data, '\n'), 0o644)
}

// daemonAlive is a version-gate-free pidfile liveness check (a v1 world cannot
// be world.Open'd under the v2 build, so daemon.IsRunning would falsely report
// "not running"). It mirrors internal/daemon's acquirePidfile/IsRunning check;
// duplicated rather than imported to avoid an import cycle (daemon → world).
func daemonAlive(w *World) (bool, int) {
	data, err := os.ReadFile(w.PidPath())
	if err != nil {
		return false, 0
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || !pidAlive(pid) {
		return false, 0
	}
	return true, pid
}

func pidAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}

package world

// The fork ceremony (spec 076): copy a world at its latest snapshot boundary
// under a fresh identity — the Migrate ceremony's sibling. Fork is a
// fresh-store ceremony, never a file copy: the events table is append-only
// in-schema (events_no_update/events_no_delete triggers), so "truncated to
// the snapshot boundary" is built by streaming the parent's event prefix
// (seq <= boundary.seq) into a new log, writing the boundary snapshot
// verbatim, appending the world.forked lineage event, and stamping meta —
// seed and format_version to satisfy validateMeta at first boot, plus every
// llm_spend_* key so the fork inherits the parent's wallet as of fork time
// (FR-012: forking never mints fresh budget). Identity is name, directory,
// socket, and registry presence; the seed is CARRIED — the prefix events
// were generated under it and sim.rngAt keys off it (research R2). Forks
// are independent worlds forever: no merge verb exists or will (FR-014).

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
)

// ForkResult summarizes a completed fork for the CLI (the MigrateResult
// pattern): everything cmdFork prints, nothing it must recompute.
type ForkResult struct {
	Name          string // fork manifest name
	Dir           string // destination directory
	ParentName    string
	ForkTick      int64 // boundary tick (CLI renders day/HH:MM via clock)
	ForkSeq       int64 // events carried: 1..ForkSeq
	TruncatedTail int64 // parent events past the boundary NOT carried (0 = nothing lost)
	BoundaryEnded bool  // boundary state carries an ended run (warn — spec edge case)
	SpendCarried  bool  // llm_spend_* keys found and copied (AC5 line in the summary)
}

// forkCopyFiles are the sidecar files carried verbatim into the fork
// (research R9): player input, per-world physics/profiles, and the LLM
// wallet config (same ceiling — FR-012). Absent files are simply skipped.
var forkCopyFiles = []string{
	"llm.json",
	"calibration.json",
	"estimator_state.json",
	"charter.md",
	"tuning.json",
}

// forkCopyDirs are the sidecar directories carried verbatim (research R9):
// the guardian soul + transcript, boot-frozen bundle content, and the
// agents/ drop-in files. Everything NOT listed here or in forkCopyFiles
// stays behind by design: runtime files (daemon.sock/pid/log), the parent
// DB and its WAL sidecars, migration archives (world.v*.db), and the
// scribe's regenerable views (chronicle.md, morgue.md, village_charter.md —
// regenerated from recovered state at the fork's first daemon start;
// copying them would ship prose about truncated-away events).
var forkCopyDirs = []string{
	"metatron",
	"bundles",
	"agents",
}

// errForkStopReplay is the prefix-stream sentinel: streaming stops cleanly
// at the first event past the boundary.
var errForkStopReplay = errors.New("fork: stop replay at boundary")

// Fork copies srcDir's world at its latest hash-valid snapshot boundary into
// destDir under newName (spec 076 FR-001..005, FR-007/008, FR-012). It
// refuses a running source daemon, a source with no valid snapshot, and a
// non-empty destination; on any mid-ceremony failure the partial destination
// is best-effort removed so a retry is clean (the destination was
// required-empty, so removal destroys nothing pre-existing).
func Fork(srcDir, destDir, newName string) (*ForkResult, error) {
	// Current-format gate: fork never crosses formats — migrate first. The
	// standard Open error (migrate hint included) surfaces verbatim.
	src, err := Open(srcDir)
	if err != nil {
		return nil, err
	}

	// Refuse a live daemon (the Migrate precedent): sidecar copies race a
	// running daemon's writes (guardian transcript, estimator flush), and the
	// ceremony is more than the WAL-safe db read.
	if running, pid := daemonAlive(src); running {
		return nil, fmt.Errorf("daemon is running (pid %d) — stop it first: promptworld stop %s", pid, srcDir)
	}

	srcStore, err := store.Open(src.DBPath())
	if err != nil {
		return nil, err
	}
	defer srcStore.Close()
	if err := srcStore.CheckContiguity(); err != nil {
		return nil, err
	}

	// The boundary: the latest snapshot whose hash verifies — the same
	// newest→oldest verified walk recovery uses (FR-002). v1 forks here and
	// only here; a never-snapshotted world is refused with the remedy.
	boundary, err := srcStore.LatestValidSnapshot()
	if err != nil {
		return nil, err
	}
	if boundary == nil {
		return nil, fmt.Errorf("world %q has no valid snapshot to fork at — start and stop it once to cut one: promptworld start %s && promptworld stop %s",
			src.Manifest.Name, srcDir, srcDir)
	}

	// Destination: the Create posture — may exist only if empty, so runs can
	// never bleed into each other.
	if entries, err := os.ReadDir(destDir); err == nil && len(entries) > 0 {
		return nil, fmt.Errorf("directory %s is not empty", destDir)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(destDir, "agents"), 0o755); err != nil {
		return nil, err
	}

	res, err := forkInto(src, srcStore, boundary, destDir, newName)
	if err != nil {
		// Best-effort cleanup (research R9): no partial world left behind.
		os.RemoveAll(destDir)
		return nil, err
	}
	return res, nil
}

// forkInto is the ceremony body past the point of destination creation —
// split out so Fork can remove the partial destination on any error.
func forkInto(src *World, srcStore *store.Store, boundary *store.Snapshot, destDir, newName string) (*ForkResult, error) {
	fresh, err := store.Open(filepath.Join(destDir, "world.db"))
	if err != nil {
		return nil, err
	}
	defer fresh.Close()

	// Log format stamp (spec 094 FR-001): a fork's log is born current — the
	// parent passed world.Open's gate, so its prefix already speaks this
	// build's vocabulary.
	if err := fresh.StampLogFormat(); err != nil {
		return nil, err
	}

	// Stream the parent's prefix (seq <= boundary.Seq) in order into the
	// fresh log. AppendEvents assigns contiguous seqs from lastSeq+1, so an
	// in-order stream into an empty store reproduces seqs 1..N exactly;
	// wall_time rides verbatim (observability metadata, excluded from every
	// determinism comparison).
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
	err = srcStore.ReplayEvents(0, func(e store.Event) error {
		if e.Seq > boundary.Seq {
			return errForkStopReplay
		}
		batch = append(batch, e)
		if len(batch) == batchSize {
			return flush()
		}
		return nil
	})
	if err != nil && !errors.Is(err, errForkStopReplay) {
		return nil, err
	}
	if err := flush(); err != nil {
		return nil, err
	}
	if got := fresh.LastSeq(); got != boundary.Seq {
		return nil, fmt.Errorf("fork carried %d events, want the boundary's %d — the parent log is inconsistent with its snapshot", got, boundary.Seq)
	}

	// The lineage event (FR-007): authoritative provenance, appended at the
	// boundary tick — it lands at seq boundary.Seq+1 and the reducer no-ops
	// it, so fork state at (ForkTick, ForkSeq) stays byte-identical to the
	// parent's.
	forkedPayload, err := json.Marshal(sim.WorldForkedPayload{
		ParentName:      src.Manifest.Name,
		ParentSeed:      src.Manifest.Seed,
		ParentCreatedAt: src.Manifest.CreatedAt,
		ForkTick:        boundary.Tick,
		ForkSeq:         boundary.Seq,
	})
	if err != nil {
		return nil, err
	}
	if err := fresh.AppendEvents([]store.Event{
		{Tick: boundary.Tick, Type: "world.forked", Payload: forkedPayload},
	}); err != nil {
		return nil, err
	}

	// The boundary snapshot, verbatim (FR-004): SaveSnapshot recomputes the
	// hash over the same bytes, and we re-verify it reproduces the parent's
	// state_hash — the walk already verified the parent row, so a mismatch
	// here can only be a write fault.
	if err := fresh.SaveSnapshot(boundary.Tick, boundary.Seq, boundary.State); err != nil {
		return nil, err
	}
	check, err := fresh.LatestValidSnapshot()
	if err != nil {
		return nil, err
	}
	if check == nil || check.Hash != boundary.Hash {
		return nil, fmt.Errorf("fork snapshot failed hash re-verification against the parent's state_hash")
	}

	// Meta: seed + format_version match what validateMeta checks at first
	// boot; every llm_spend_* key copies verbatim (FR-012 — the fork's meter
	// opens at the parent's month, spend-so-far, and attribution: same
	// ceiling via the copied llm.json, inherited spend, no fresh grant).
	if err := fresh.SetMeta("seed", fmt.Sprintf("%d", src.Manifest.Seed)); err != nil {
		return nil, err
	}
	if err := fresh.SetMeta("format_version", fmt.Sprintf("%d", src.Manifest.FormatVersion)); err != nil {
		return nil, err
	}
	spend, err := srcStore.MetaByPrefix("llm_spend_")
	if err != nil {
		return nil, err
	}
	for k, v := range spend {
		if err := fresh.SetMeta(k, v); err != nil {
			return nil, err
		}
	}

	// The fork manifest (FR-005): name, created_at, and lineage are new;
	// EVERY other field — seed included — rides verbatim.
	m := src.Manifest
	m.Name = newName
	m.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	m.Lineage = &LineageConfig{
		Parent:          src.Manifest.Name,
		ParentCreatedAt: src.Manifest.CreatedAt,
		ForkTick:        boundary.Tick,
	}
	if err := writeManifest(&World{Dir: destDir, Manifest: m}); err != nil {
		return nil, err
	}

	// Sidecars, per the research R9 catalog. Copied as-of fork time, not
	// as-of the snapshot — the charter is player INPUT, not event-sourced
	// state (documented coarseness, harmless by construction).
	for _, name := range forkCopyFiles {
		if err := copySidecarFile(filepath.Join(src.Dir, name), filepath.Join(destDir, name)); err != nil {
			return nil, err
		}
	}
	for _, name := range forkCopyDirs {
		if err := copySidecarDir(filepath.Join(src.Dir, name), filepath.Join(destDir, name)); err != nil {
			return nil, err
		}
	}

	// BoundaryEnded: the boundary state may carry an ended run (a gracefully
	// stopped ended world's final snapshot does) — the fork is then born
	// ended, legal but warned about (spec edge case).
	bState := sim.NewState(src.Manifest.Seed, src.Map())
	if err := json.Unmarshal(boundary.State, bState); err != nil {
		return nil, fmt.Errorf("boundary snapshot state unreadable: %w", err)
	}

	return &ForkResult{
		Name:          newName,
		Dir:           destDir,
		ParentName:    src.Manifest.Name,
		ForkTick:      boundary.Tick,
		ForkSeq:       boundary.Seq,
		TruncatedTail: srcStore.LastSeq() - boundary.Seq,
		BoundaryEnded: bState.Ended,
		SpendCarried:  len(spend) > 0,
	}, nil
}

// copySidecarFile copies one regular file verbatim; an absent source is
// legal and skipped (every sidecar is optional — an absent llm.json means
// no LLM traffic, an absent tuning.json means doctrine constants).
func copySidecarFile(src, dest string) error {
	in, err := os.Open(src)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer in.Close()
	info, err := in.Stat()
	if err != nil {
		return err
	}
	out, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// copySidecarDir recursively copies a directory's contents verbatim; an
// absent source is legal and skipped.
func copySidecarDir(src, dest string) error {
	entries, err := os.ReadDir(src)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}
	for _, e := range entries {
		s, d := filepath.Join(src, e.Name()), filepath.Join(dest, e.Name())
		if e.IsDir() {
			if err := copySidecarDir(s, d); err != nil {
				return err
			}
			continue
		}
		if err := copySidecarFile(s, d); err != nil {
			return err
		}
	}
	return nil
}

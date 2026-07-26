package worlds

// The per-user curriculum-ladder unlocks record (spec 046 US3,
// contracts/unlocks-record.md): `<worlds.Root()>/unlocks.json`, managed with
// the SAME doctrine as known_worlds.json (registry.go) — load-tolerant,
// atomic (.tmp+rename) writes — plus one difference the contract calls out
// explicitly: an unresolvable home directory WARNS and degrades rather than
// erroring (the endpoint-lease precedent, internal/llm/lease.go), because
// this record is pure convenience layered over a world's own event history,
// which remains the authority regardless (FR-008). No world behavior ever
// reads this file — its only consumers are `promptworld stages` and
// `promptworld new`'s earned-stage check.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/evanstern/promptworld/internal/world"
)

// UnlocksPath returns <root>/unlocks.json.
func UnlocksPath() (string, error) {
	root, err := Root()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "unlocks.json"), nil
}

// UnlockEvidenceRef points at one event in the proving world's history —
// (type, seq, tick) is enough to re-locate it in that world's log, so a
// record entry's claim stays independently auditable (FR-008).
type UnlockEvidenceRef struct {
	Type string `json:"type"`
	Seq  int64  `json:"seq"`
	Tick int64  `json:"tick"`
}

// UnlockEntry is one earned stage's record: the proving world's name and
// path, the exercise that passed, the evidence satisfying the gate, and when
// it was recorded (RFC3339, the Manifest.CreatedAt convention).
type UnlockEntry struct {
	World    string              `json:"world"`
	Path     string              `json:"path"`
	Exercise string              `json:"exercise"`
	Evidence []UnlockEvidenceRef `json:"evidence,omitempty"`
	EarnedAt string              `json:"earned_at"`
}

// unlocksFile is the on-disk shape: {"unlocks": {"stage-2": {...}, ...}}.
type unlocksFile struct {
	Unlocks map[string]UnlockEntry `json:"unlocks"`
}

// Unlocks is the in-memory, load-tolerant view of the per-user unlocks
// record: stage id -> the entry that earned it.
type Unlocks struct {
	Entries map[string]UnlockEntry
}

// Earned reports whether the record holds an entry for stage — the only
// question the record is ever asked (contract: "no world behavior ever
// reads the record"). A nil Unlocks (the unresolvable-home degrade) reports
// nothing earned, matching a fresh player's honest default.
func (u *Unlocks) Earned(stage string) bool {
	if u == nil {
		return false
	}
	_, ok := u.Entries[stage]
	return ok
}

// StageEarned reports whether stage is offered WITHOUT an override: stage-1
// is every player's floor (asked of no one — R9's "default stage-1 for new
// players" made unconditional rather than conditioned on an empty record),
// every other stage needs a record entry (Earned). RELOCATED here from
// cmd/promptworld/stages.go's package-main `stageEarned` (spec 063 T014's
// one-source-two-surfaces precedent extended from catalog content to earned
// state, spec 078 FR-003): `promptworld stages` and the TUI's forward-ladder
// section both call this method instead of keeping two copies of the rule.
// Nil-receiver safe, like Earned.
func (u *Unlocks) StageEarned(stage string) bool {
	return stage == world.Stage1 || u.Earned(stage)
}

// unlocksWarnf surfaces an unlocks-record warning (warn-not-error, the
// lease.go precedent): a player whose home directory cannot be resolved
// loses the unlock convenience layer for this run — never fatal, since the
// worlds' own histories remain the authority regardless of this file's
// fate. Overridable in tests.
var unlocksWarnf = func(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "promptworld: "+format+"\n", args...)
}

// LoadUnlocks reads the per-user unlocks record. Missing or corrupt yields
// an empty record, never an error (registry doctrine, D6); an unresolvable
// home directory warns and yields an empty record too (lease.go precedent)
// — the player loses convenience, never truth (FR-008).
func LoadUnlocks() *Unlocks {
	path, err := UnlocksPath()
	if err != nil {
		unlocksWarnf("could not resolve the unlocks record (%v) — proceeding as if nothing is earned yet", err)
		return &Unlocks{Entries: map[string]UnlockEntry{}}
	}
	return loadUnlocksFrom(path)
}

func loadUnlocksFrom(path string) *Unlocks {
	u := &Unlocks{Entries: map[string]UnlockEntry{}}
	data, err := os.ReadFile(path)
	if err != nil {
		return u // missing ⇒ empty, not an error
	}
	var uf unlocksFile
	if err := json.Unmarshal(data, &uf); err != nil {
		return u // corrupt ⇒ empty, not an error
	}
	for stage, entry := range uf.Unlocks {
		// Heal at load: an entry missing a required identity field is
		// malformed and dropped; an entry whose world path no longer exists
		// is KEPT (contract rule 3 — an archived/moved world is still
		// historical proof; only malformed entries are dropped).
		if entry.World == "" || entry.Path == "" || entry.Exercise == "" || entry.EarnedAt == "" {
			continue
		}
		u.Entries[stage] = entry
	}
	return u
}

// UpsertUnlock records (or repairs) the entry earning stage, writing the
// record atomically (.tmp + rename, registry.go precedent). Append-shaped
// (contract rule 4): existing entries for OTHER stages are preserved; the
// named stage's entry is set to entry (the freshest observed proof). A
// failure to resolve the home directory or to write — advisory, never
// authority — warns and returns, never treated as fatal by any caller
// (the daemon-side observer must never perturb the sim loop over this).
func UpsertUnlock(stage string, entry UnlockEntry) {
	path, err := UnlocksPath()
	if err != nil {
		unlocksWarnf("could not resolve the unlocks record (%v) — %s unlock not recorded (the world's own history remains the proof)", err, stage)
		return
	}
	u := loadUnlocksFrom(path)
	u.Entries[stage] = entry
	if err := writeUnlocks(path, u); err != nil {
		unlocksWarnf("could not write the unlocks record: %v", err)
	}
}

func writeUnlocks(path string, u *Unlocks) error {
	uf := unlocksFile{Unlocks: u.Entries}
	if uf.Unlocks == nil {
		uf.Unlocks = map[string]UnlockEntry{}
	}
	data, err := json.MarshalIndent(uf, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

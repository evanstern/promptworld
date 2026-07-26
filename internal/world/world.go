// Package world owns the save-directory layout: one directory = one world run.
// Everything belonging to a run lives inside it; nothing is global.
package world

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/evanstern/promptworld/internal/worldmap"
)

const (
	ManifestName = "world.json"
	// FormatVersion 5 is the terrain-vocabulary break (spec 068): new worlds
	// generate marsh/sand terrain gated by the manifest's terrain_gen field,
	// and software predating the field would silently IGNORE it and
	// regenerate different terrain under the same manifest — agents and
	// structures standing on water (FR-007). The version bump makes a
	// new-vocabulary world unmistakable: a pre-068 build refuses it at Open
	// with the migrate hint instead of mis-generating (C10). FormatVersion 4
	// was the mental-maps break (spec 041): per-agent private spatial
	// knowledge gates target resolution, so a v3 world loaded without seeded
	// maps would leave every villager knowing nothing — mass starvation.
	// FormatVersion 3 was the spec 013 inventory/storage break (bulk cap,
	// yield truncation, death spill); FormatVersion 2 the spec 012
	// resources/food/crafting break.
	FormatVersion = 5
)

type Manifest struct {
	Name            string `json:"name"`
	Seed            uint64 `json:"seed"`
	CreatedAt       string `json:"created_at"`
	FormatVersion   int    `json:"format_version"`
	TickGameSeconds int    `json:"tick_game_seconds"`
	// Map dimensions; zero values (older saves) default to
	// worldmap.DefaultSize on Open. Terrain is regenerated from
	// (Seed, MapWidth, MapHeight, TerrainGen), never persisted.
	MapWidth  int `json:"map_width,omitempty"`
	MapHeight int `json:"map_height,omitempty"`
	// TerrainGen selects the terrain generation algorithm (spec 068,
	// FR-006/FR-007): absent/0 = legacy generation, bit-identical to
	// pre-068 (worldmap.GenLegacy — what `promptworld migrate` leaves a
	// carried-forward world with, so its terrain never shifts);
	// worldmap.GenMarshSand (2) adds the marsh/sand shoreline pass and is
	// what `promptworld new` writes. Any other value is refused at Open
	// (the format_version posture): a future generation must never be
	// silently re-generated under this build's algorithms.
	TerrainGen int `json:"terrain_gen,omitempty"`
	// Meeting is an optional per-world meeting convention (TASK-36). When
	// present, the daemon seeds a meeting.convention_established event on boot
	// so the convene→open lifecycle honors it. Absent, no meeting convenes
	// unless one emerges in-world. `promptworld new` never writes it —
	// emergent is the default.
	Meeting *MeetingConfig `json:"meeting,omitempty"`
	// Teaching marks a teaching-posture world (decision-6, spec 039): the
	// daemon defaults its speed to the highest planner-safe ladder rung at each
	// boot and surfaces the horizon arithmetic on override. Absent ⇒
	// non-teaching, and a non-teaching world.json round-trips byte-identically
	// (omitempty, FR-008); no FormatVersion bump — an additive defaulting bool
	// old readers ignore.
	Teaching bool `json:"teaching,omitempty"`
	// MemoryRelevance gates memory-window selection (spec 042): "" (absent,
	// the default) keeps today's salience+recency ranking; "shadow" computes
	// the relevance-augmented ranking too and records rank divergence while
	// prompts still get the legacy window; "on" lets the augmented window feed
	// prompts (divergence still recorded). The shadow→on flip is an operator
	// world.json edit gated on the documented divergence threshold decision
	// (FR-007). Additive omitempty string (Teaching precedent): a pre-042
	// world.json round-trips byte-identically, no FormatVersion bump.
	MemoryRelevance string `json:"memory_relevance,omitempty"`
	// Stage is the world's curriculum-ladder stage (spec 046, FR-002):
	// "stage-1".."stage-4", set once at creation and immutable for the world's
	// lifetime — no mutation command exists or will (the SetTeaching pattern is
	// deliberately NOT replicated). Absent ("") = a pre-ladder world = ungated
	// (stage-4 semantics), so existing worlds lose nothing. Validated at Open
	// (the MemoryRelevance closed-vocabulary precedent). Additive omitempty:
	// a pre-046 world.json round-trips byte-identically, no FormatVersion bump.
	Stage string `json:"stage,omitempty"`
	// StageOverridden records that the world was created at an unearned stage
	// via `promptworld new --stage <id> --override` (spec 046, FR-003) — the
	// honesty marker that makes overridden runs comparable as overridden runs.
	StageOverridden bool `json:"stage_overridden,omitempty"`
	// CharterPreset names the authored charter constant that seeds charter.md
	// at genesis and — at stage-1, where instruction files are locked — IS the
	// effective charter regardless of player edits (spec 046, FR-005). ""/
	// "default" = persona.DefaultCharter; "tutor" = the stage-1 orientation
	// preset. Closed vocabulary, validated at Open.
	CharterPreset string `json:"charter_preset,omitempty"`
	// Scenario names the seeded exercise this world runs (spec 046 US4
	// reserved the block; spec 054 consumes it): written once by
	// `promptworld new --scenario` (SetScenario below, the SetStage
	// pattern), validated at Open, and boot-frozen into the sim loop by the
	// daemon (sim.State.ArmScenario) — the incident schedule, rubric
	// evaluator, status facts, and exercise tab all key off it. Absent =
	// an ambient world: every scenario code path stays dormant and the
	// world is byte-identical to pre-054.
	Scenario *ScenarioConfig `json:"scenario,omitempty"`
}

// ScenarioConfig names the exercise a world is seeded to run (spec 046 US4,
// consumed by spec 054). Exercise names a sim.ExerciseDefinition.ID
// (e.g. "first-night"), validated at Open via ValidScenarioExercise.
type ScenarioConfig struct {
	Exercise string `json:"exercise,omitempty"`
}

// ValidScenarioExercise reports whether s names a shipped scenario exercise
// (spec 054 FR-006). Deliberately a LOCAL mirror of sim.ScenarioExercises'
// id set — the validLadderStage twin-list precedent, in reverse: the
// deterministic core does not import this save-directory package and this
// package does not import the core, so each side keeps its own closed
// vocabulary. TestScenarioVocabularyMirrorsSimCatalog (world_test, which may
// import sim) pins the two in sync.
func ValidScenarioExercise(s string) bool {
	switch s {
	case "first-night", "cold-dawn", "stranger-at-the-gate",
		"the-law", "blighted-larder",
		"toolsmith", "fog-watch",
		"long-winter", "stewards-charge":
		return true
	}
	return false
}

// The four curriculum-ladder stage ids (spec 046, FR-001). An absent Stage is
// legal (a pre-ladder, ungated world); any other value is refused at Open — a
// typo must never silently run ungated.
const (
	Stage1 = "stage-1"
	Stage2 = "stage-2"
	Stage3 = "stage-3"
	Stage4 = "stage-4"
)

// ValidStage reports whether s is a legal Manifest.Stage value: one of the
// four ladder ids, or "" (absent = ungated pre-ladder world).
func ValidStage(s string) bool {
	switch s {
	case "", Stage1, Stage2, Stage3, Stage4:
		return true
	}
	return false
}

// The legal charter_preset names (spec 046). "" and "default" both mean the
// authored default charter; "tutor" is the stage-1 orientation preset.
const (
	CharterPresetDefault = "default"
	CharterPresetTutor   = "tutor"
)

// StageLadderInfo is the ladder table's SKIN-INDEPENDENT content (spec 046
// spec.md "The ladder" table): the concept taught, what the world grants,
// and what evidence unlocks the next stage. Contrast internal/skin's
// StageIdentity, which is skin DATA (display name + one-line identity) —
// this table is substrate fact, invariant across skins. RELOCATED here from
// cmd/promptworld/stages.go (spec 063 T014): the help overlay's guardian
// section (D9) reads the same table `promptworld stages` renders from, and
// internal/tui cannot import package main — one source, two surfaces.
type StageLadderInfo struct {
	Concept        string
	Grants         string
	UnlockEvidence string // "" only for stage-4 (graduation — nothing unlocks past it)
}

// StageOrder is the ladder's presentation order — always all four stages,
// per spec 046 FR-003 ("later stages are visible with their identity,
// concept, and unlock evidence stated").
var StageOrder = []string{Stage1, Stage2, Stage3, Stage4}

// StagesLadder mirrors spec 046 spec.md's ladder table (client-approved
// 2026-07-25, AC #5) plus the ratified stage-1 ceiling amendment (TASK-119
// board artifact "first-night teaches visions+orders" — standing orders
// joined the stage-1 grant, contracts/stage-gating.md).
var StagesLadder = map[string]StageLadderInfo{
	Stage1: {
		Concept: "conversational prompting: asking well, watching outcomes, iterating",
		Grants: "the base conversational guardian + basic query/nudge tools (visions, omens, " +
			"and the watch — monitor_and_act/cancel_order); instruction files are locked " +
			"(the default or tutor charter is in force, edits get an honest notice)",
		UnlockEvidence: "pass a stage-1 scenario (the first-night exercise)",
	},
	Stage2: {
		Concept: "instruction authoring: durable behavior lives in an authored instruction file",
		Grants:  "stage-1 grants + charter editing unlocked",
		UnlockEvidence: "pass a stage-2 scenario while a player-authored charter revision " +
			"is in force",
	},
	Stage3: {
		Concept: "capability design: what the guardian can do is itself authored — skill " +
			"files + tool grants",
		Grants: "stage-2 grants + skill files compose + the gated tool manifest opens",
		UnlockEvidence: "pass a stage-3 scenario in which a player-granted tool's act " +
			"contributes to the pass",
	},
	Stage4: {
		Concept:        "mastery: indirect influence at world scale; the ambient world as the endgame",
		Grants:         "the full tool roster, including capstone capabilities (canonization)",
		UnlockEvidence: "", // graduation (synthesis decision 3) — nothing unlocks past it
	},
}

// ValidCharterPreset reports whether s is a legal Manifest.CharterPreset
// value: "", "default", or "tutor".
func ValidCharterPreset(s string) bool {
	switch s {
	case "", CharterPresetDefault, CharterPresetTutor:
		return true
	}
	return false
}

// The three legal memory_relevance states (spec 042). Any other value is
// refused at Open — a typo must never silently run as "off".
const (
	MemoryRelevanceOff    = ""
	MemoryRelevanceShadow = "shadow"
	MemoryRelevanceOn     = "on"
)

// MeetingConfig declares when (and optionally where) the daily assembly
// convenes. Convene and Open are 24-hour game clock times, "HH:MM"; Convene
// must fall before Open and both within the day. X/Y are optional map
// coordinates for the meeting place — omitted, the daemon derives it (first
// fire, else first shelter, else map center).
type MeetingConfig struct {
	Convene string `json:"convene"`
	Open    string `json:"open"`
	X       *int   `json:"x,omitempty"`
	Y       *int   `json:"y,omitempty"`
}

// Seconds parses Convene/Open into seconds-of-day and validates the window:
// both well-formed "HH:MM" within the day, convene strictly before open.
func (c *MeetingConfig) Seconds() (convene, open int, err error) {
	if convene, err = parseClock(c.Convene); err != nil {
		return 0, 0, fmt.Errorf("meeting.convene: %w", err)
	}
	if open, err = parseClock(c.Open); err != nil {
		return 0, 0, fmt.Errorf("meeting.open: %w", err)
	}
	if convene >= open {
		return 0, 0, fmt.Errorf("meeting.convene (%s) must be before meeting.open (%s)", c.Convene, c.Open)
	}
	return convene, open, nil
}

// parseClock reads an "HH:MM" 24-hour time into a second-of-day in [0, 86400).
func parseClock(s string) (int, error) {
	var h, m int
	if _, err := fmt.Sscanf(s, "%d:%d", &h, &m); err != nil {
		return 0, fmt.Errorf("time %q is not HH:MM: %w", s, err)
	}
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, fmt.Errorf("time %q out of range (00:00–23:59)", s)
	}
	return h*3600 + m*60, nil
}

type World struct {
	Dir      string
	Manifest Manifest
}

// Create initializes a new save directory. The directory may exist only if
// empty; anything else is refused so runs can never bleed into each other.
func Create(dir, name string, seed uint64) (*World, error) {
	if entries, err := os.ReadDir(dir); err == nil && len(entries) > 0 {
		return nil, fmt.Errorf("directory %s is not empty", dir)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(dir, "agents"), 0o755); err != nil {
		return nil, err
	}
	m := Manifest{
		Name:            name,
		Seed:            seed,
		CreatedAt:       time.Now().UTC().Format(time.RFC3339),
		FormatVersion:   FormatVersion,
		TickGameSeconds: 1,
		MapWidth:        worldmap.DefaultSize,
		MapHeight:       worldmap.DefaultSize,
		// New worlds are born on the current terrain generation (spec 068
		// C12); only migrated legacy worlds carry an absent terrain_gen.
		TerrainGen: worldmap.GenMarshSand,
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(dir, ManifestName), append(data, '\n'), 0o644); err != nil {
		return nil, err
	}
	return &World{Dir: dir, Manifest: m}, nil
}

// ErrFormatVersionMismatch is Open's one failure mode that is not a genuine
// content problem: the manifest parsed fine, it's just a format_version this
// build doesn't support (older or newer) — the world itself may be perfectly
// healthy. Daemon-lifecycle callers that only need socket/pid reachability
// (TASK-147: stop/status must reach a running old-version daemon rather
// than deadlock behind the migrate hint) match this with errors.As to tell
// "can't read this world's content" apart from any other Open failure
// (corrupt JSON, a directory that was never a world at all, ...), which
// still surfaces verbatim.
type ErrFormatVersionMismatch struct {
	Got, Want int
}

func (e *ErrFormatVersionMismatch) Error() string {
	return fmt.Sprintf("world format_version %d unsupported (this build supports %d); run 'promptworld migrate <world>' to upgrade an older world", e.Got, e.Want)
}

// Open loads and validates an existing save directory.
func Open(dir string) (*World, error) {
	data, err := os.ReadFile(filepath.Join(dir, ManifestName))
	if err != nil {
		return nil, fmt.Errorf("not a world directory (missing %s): %w", ManifestName, err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("corrupt %s: %w", ManifestName, err)
	}
	if m.FormatVersion != FormatVersion {
		return nil, &ErrFormatVersionMismatch{Got: m.FormatVersion, Want: FormatVersion}
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
	if m.Meeting != nil {
		if _, _, err := m.Meeting.Seconds(); err != nil {
			return nil, fmt.Errorf("corrupt %s: %w", ManifestName, err)
		}
	}
	switch m.MemoryRelevance {
	case MemoryRelevanceOff, MemoryRelevanceShadow, MemoryRelevanceOn:
	default:
		return nil, fmt.Errorf("corrupt %s: memory_relevance %q unknown (want %q, %q, or the key absent)", ManifestName, m.MemoryRelevance, MemoryRelevanceShadow, MemoryRelevanceOn)
	}
	if !ValidStage(m.Stage) {
		return nil, fmt.Errorf("corrupt %s: stage %q unknown (want %q..%q or the key absent)", ManifestName, m.Stage, Stage1, Stage4)
	}
	if !ValidCharterPreset(m.CharterPreset) {
		return nil, fmt.Errorf("corrupt %s: charter_preset %q unknown (want %q, %q, or the key absent)", ManifestName, m.CharterPreset, CharterPresetDefault, CharterPresetTutor)
	}
	// Scenario block (spec 054 FR-006, the ValidStage idiom): a present block
	// must name a shipped exercise — a typo must never silently boot ambient.
	if m.Scenario != nil && !ValidScenarioExercise(m.Scenario.Exercise) {
		return nil, fmt.Errorf("corrupt %s: scenario exercise %q unknown (want a shipped exercise id, or the block absent)", ManifestName, m.Scenario.Exercise)
	}
	// Terrain generation (spec 068 C12, the format_version posture): absent/0
	// is legacy; the current generation is accepted; anything else — a future
	// generation this build does not implement — is refused, never silently
	// re-generated under the wrong algorithm.
	switch m.TerrainGen {
	case worldmap.GenLegacy, worldmap.GenMarshSand:
	default:
		return nil, fmt.Errorf("world terrain_gen %d unsupported (this build supports %d, or the key absent for legacy terrain); upgrade promptworld to open this world", m.TerrainGen, worldmap.GenMarshSand)
	}
	return &World{Dir: dir, Manifest: m}, nil
}

// SetTeaching flips a world's teaching-posture marker on disk (spec 039): read
// the manifest, set Teaching, rewrite world.json. A running daemon is
// unaffected — it reads the marker only at its next boot, so the toggle is an
// offline config edit. Used by `promptworld new --teaching` (birth) and
// `promptworld teaching <world> on|off`.
func SetTeaching(dir string, on bool) error {
	w, err := Open(dir)
	if err != nil {
		return err
	}
	w.Manifest.Teaching = on
	data, err := json.MarshalIndent(w.Manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, ManifestName), append(data, '\n'), 0o644)
}

// SetScenario stamps a freshly created world's scenario block (spec 054
// US3): read the manifest, set Scenario{Exercise}, rewrite world.json.
// Exactly the SetStage contract below: ONE caller (`promptworld new
// --scenario`), called once immediately after Create — the scenario is
// write-once for the world's lifetime, no toggle command exists or ever
// will (the machinery is boot-frozen; a mutable scenario would desync the
// armed schedule from the manifest). Callers pass an already-validated id
// (ValidScenarioExercise) — this function does not re-validate.
func SetScenario(dir, exercise string) error {
	w, err := Open(dir)
	if err != nil {
		return err
	}
	w.Manifest.Scenario = &ScenarioConfig{Exercise: exercise}
	data, err := json.MarshalIndent(w.Manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, ManifestName), append(data, '\n'), 0o644)
}

// SetStage stamps a freshly created world's curriculum-ladder stage fact
// (spec 046, FR-002/FR-003): read the manifest, set Stage/StageOverridden/
// CharterPreset, rewrite world.json. Unlike SetTeaching, this has exactly
// ONE caller — `promptworld new` — and is called exactly once, immediately
// after Create: stage is write-once for a world's whole lifetime (R1) — no
// `promptworld stage <world> ...` toggle command exists or ever will.
// Callers must pass already-validated values (ValidStage/ValidCharterPreset)
// — this function does not re-validate, matching SetTeaching's contract.
func SetStage(dir, stage string, overridden bool, charterPreset string) error {
	w, err := Open(dir)
	if err != nil {
		return err
	}
	w.Manifest.Stage = stage
	w.Manifest.StageOverridden = overridden
	w.Manifest.CharterPreset = charterPreset
	data, err := json.MarshalIndent(w.Manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, ManifestName), append(data, '\n'), 0o644)
}

// Map regenerates the world's terrain from the manifest — deterministic, so
// daemon and clients derive identical maps without any wire transfer. The
// manifest's terrain generation version rides along (spec 068): an absent/0
// terrain_gen reproduces legacy terrain bit-identically.
func (w *World) Map() *worldmap.Map {
	return worldmap.GenerateV(w.Manifest.Seed, w.Manifest.MapWidth, w.Manifest.MapHeight, w.Manifest.TerrainGen)
}

func (w *World) DBPath() string        { return filepath.Join(w.Dir, "world.db") }
func (w *World) LLMConfigPath() string { return filepath.Join(w.Dir, "llm.json") }

// CalibrationPath is the seconds-per-point profile written only by
// `promptworld calibrate` (specs/007-cognition-horizon); an absent file is
// legal — pessimistic bootstrap defaults apply.
func (w *World) CalibrationPath() string { return filepath.Join(w.Dir, "calibration.json") }

// TuningPath is the optional per-world tuning manifest (spec 048): operator-
// authored, sparse JSON promoting doctrine constants to per-world dials. An
// absent file is the default state for every world — behavior is exactly the
// doctrine constants. Never written by `promptworld new`.
func (w *World) TuningPath() string { return filepath.Join(w.Dir, "tuning.json") }

// EstimatorStatePath is the persisted live per-provider seconds-per-point
// snapshot (TASK-113): unlike calibration.json (human-authored via `promptworld
// calibrate`, daemon-read-only), this file IS daemon-written — periodically and
// at shutdown — so the estimator's learned drift survives a restart instead of
// resetting to the calibration/bootstrap floor every time. An absent file is
// legal: boot reseeds from calibration/bootstrap alone, exactly as before this
// file existed.
func (w *World) EstimatorStatePath() string { return filepath.Join(w.Dir, "estimator_state.json") }

// SockPathIn and PidPathIn are the bare-dir counterparts of SockPath/PidPath
// below — pure path joins, not a validating Open. Daemon-lifecycle callers
// (start/stop/status, daemon.IsRunning) need the socket/pid path for a world
// this build cannot necessarily world.Open (e.g. a format_version this build
// doesn't support): a deadlock lived exactly this (TASK-147, spec 068's
// FormatVersion 4→5) — a running old-version daemon that a version-gated
// Open made unreachable to stop. World-content commands keep the Open gate;
// only these lifecycle paths bypass it.
func SockPathIn(dir string) string { return filepath.Join(dir, "daemon.sock") }
func PidPathIn(dir string) string  { return filepath.Join(dir, "daemon.pid") }

func (w *World) SockPath() string    { return SockPathIn(w.Dir) }
func (w *World) PidPath() string     { return PidPathIn(w.Dir) }
func (w *World) CharterPath() string { return filepath.Join(w.Dir, "charter.md") }

// VillageCharterPath is the village's law (TASK-13) — a scribe-rendered
// derived view of event-sourced norms, distinct from Guardian's
// player-editable charter.md above.
func (w *World) VillageCharterPath() string { return filepath.Join(w.Dir, "village_charter.md") }

// MorguePath is the run's accumulating legacy document (spec 044 US2): one
// factual epitaph per death plus a run-end summary, scribe-rendered — a
// regenerable view over the event history, never a source of truth, exactly
// like the chronicle and village charter above.
func (w *World) MorguePath() string { return filepath.Join(w.Dir, "morgue.md") }

// The "metatron" directory name is FROZEN (spec 052 ruling 2) — an on-disk
// path existing worlds carry.
func (w *World) GuardianDir() string { return filepath.Join(w.Dir, "metatron") }
func (w *World) LogPath() string     { return filepath.Join(w.Dir, "daemon.log") }

// BundlesDir is the root for pluggable bundle-defined tools (spec
// 036-scriptable-agent-tools): manifest + optional Starlark script folders
// discovered, validated, and frozen into a BundleSet at daemon boot. An
// absent directory is legal — the world boots with no bundle tools.
func (w *World) BundlesDir() string { return filepath.Join(w.Dir, "bundles") }

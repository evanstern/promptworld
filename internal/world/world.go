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
	// FormatVersion 4 is the mental-maps break (spec 041): per-agent private
	// spatial knowledge gates target resolution, so a v3 world loaded without
	// seeded maps would leave every villager knowing nothing — mass starvation;
	// the semantics of existing worlds change, which is exactly what format
	// bumps are for (research D7). An older world is refused with instructions
	// to run `promptworld migrate`. FormatVersion 3 was the spec 013
	// inventory/storage break (bulk cap, yield truncation, death spill);
	// FormatVersion 2 the spec 012 resources/food/crafting break.
	FormatVersion = 4
)

type Manifest struct {
	Name            string `json:"name"`
	Seed            uint64 `json:"seed"`
	CreatedAt       string `json:"created_at"`
	FormatVersion   int    `json:"format_version"`
	TickGameSeconds int    `json:"tick_game_seconds"`
	// Map dimensions; zero values (older saves) default to
	// worldmap.DefaultSize on Open. Terrain is regenerated from
	// (Seed, MapWidth, MapHeight), never persisted.
	MapWidth  int `json:"map_width,omitempty"`
	MapHeight int `json:"map_height,omitempty"`
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
		return nil, fmt.Errorf("world format_version %d unsupported (this build supports %d); run 'promptworld migrate <world>' to upgrade an older world", m.FormatVersion, FormatVersion)
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

// Map regenerates the world's terrain from the manifest — deterministic, so
// daemon and clients derive identical maps without any wire transfer.
func (w *World) Map() *worldmap.Map {
	return worldmap.Generate(w.Manifest.Seed, w.Manifest.MapWidth, w.Manifest.MapHeight)
}

func (w *World) DBPath() string        { return filepath.Join(w.Dir, "world.db") }
func (w *World) LLMConfigPath() string { return filepath.Join(w.Dir, "llm.json") }

// CalibrationPath is the seconds-per-point profile written only by
// `promptworld calibrate` (specs/007-cognition-horizon); an absent file is
// legal — pessimistic bootstrap defaults apply.
func (w *World) CalibrationPath() string { return filepath.Join(w.Dir, "calibration.json") }

// EstimatorStatePath is the persisted live per-provider seconds-per-point
// snapshot (TASK-113): unlike calibration.json (human-authored via `promptworld
// calibrate`, daemon-read-only), this file IS daemon-written — periodically and
// at shutdown — so the estimator's learned drift survives a restart instead of
// resetting to the calibration/bootstrap floor every time. An absent file is
// legal: boot reseeds from calibration/bootstrap alone, exactly as before this
// file existed.
func (w *World) EstimatorStatePath() string { return filepath.Join(w.Dir, "estimator_state.json") }
func (w *World) SockPath() string           { return filepath.Join(w.Dir, "daemon.sock") }
func (w *World) PidPath() string            { return filepath.Join(w.Dir, "daemon.pid") }
func (w *World) CharterPath() string        { return filepath.Join(w.Dir, "charter.md") }

// VillageCharterPath is the village's law (TASK-13) — a scribe-rendered
// derived view of event-sourced norms, distinct from Metatron's
// player-editable charter.md above.
func (w *World) VillageCharterPath() string { return filepath.Join(w.Dir, "village_charter.md") }

// MorguePath is the run's accumulating legacy document (spec 044 US2): one
// factual epitaph per death plus a run-end summary, scribe-rendered — a
// regenerable view over the event history, never a source of truth, exactly
// like the chronicle and village charter above.
func (w *World) MorguePath() string  { return filepath.Join(w.Dir, "morgue.md") }
func (w *World) MetatronDir() string { return filepath.Join(w.Dir, "metatron") }
func (w *World) LogPath() string     { return filepath.Join(w.Dir, "daemon.log") }

// BundlesDir is the root for pluggable bundle-defined tools (spec
// 036-scriptable-agent-tools): manifest + optional Starlark script folders
// discovered, validated, and frozen into a BundleSet at daemon boot. An
// absent directory is legal — the world boots with no bundle tools.
func (w *World) BundlesDir() string { return filepath.Join(w.Dir, "bundles") }

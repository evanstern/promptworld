// Package ipc implements the client attach/detach protocol from
// specs/001-world-daemon/contracts/client-protocol.md: JSON-lines over a
// Unix domain socket in the world's save directory.
package ipc

import (
	"encoding/json"

	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/skin"
	"github.com/evanstern/promptworld/internal/store"
)

type Request struct {
	ID   int64           `json:"id"`
	Cmd  string          `json:"cmd"`
	Args json.RawMessage `json:"args,omitempty"`
}

type Response struct {
	ID    int64           `json:"id"`
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data,omitempty"`
	Error string          `json:"error,omitempty"`
}

// Push is a server-initiated message to a subscribed client.
// Push == "event" carries Event; Push == "dropped" carries LastSeq (the
// subscription was canceled on buffer overflow; re-subscribe with since).
type Push struct {
	Push    string       `json:"push"`
	Event   *store.Event `json:"event,omitempty"`
	LastSeq int64        `json:"last_seq,omitempty"`
}

// wireMsg is the union read by clients: a Response has an ID, a Push has Push.
type wireMsg struct {
	ID      *int64          `json:"id,omitempty"`
	OK      bool            `json:"ok,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
	Push    string          `json:"push,omitempty"`
	Event   *store.Event    `json:"event,omitempty"`
	LastSeq int64           `json:"last_seq,omitempty"`
}

type SubscribeArgs struct {
	Since *int64 `json:"since,omitempty"`
}

type SetSpeedArgs struct {
	Speed string `json:"speed"`
}

// LLMCallArgs / LLMCallData carry the "llm_call" command (routing proof and
// the transport TASK-7 minds will use).
type LLMCallArgs struct {
	Kind      string `json:"kind"`
	System    string `json:"system,omitempty"`
	Prompt    string `json:"prompt"`
	MaxTokens int64  `json:"max_tokens,omitempty"`
}

// MetatronChatArgs carries one console turn (TASK-12); the response Data is
// a metatron.TurnResult. "metatron_status" takes no args and answers a
// metatron.Status (the model-free peek).
type MetatronChatArgs struct {
	Text string `json:"text"`
}

// MiracleArgs carries the "miracle" command (spec 016) — the operator door for
// Metatron's world edits. kind selects the miracle; the remaining fields are the
// kind-specific arguments (contracts §2). This is the ONLY surface that accepts
// gratis: --force sets it, waiving the charge; the angel path has no equivalent.
// The handler needs only the sim loop — no LLM / angel presence (pure-sim ok).
type MiracleArgs struct {
	Kind     string `json:"kind"`
	Day      int    `json:"day,omitempty"`      // time_snap
	Time     string `json:"time,omitempty"`     // time_snap "HH:MM"
	Villager string `json:"villager,omitempty"` // give_item
	Item     string `json:"item,omitempty"`     // give_item inventory kind
	Qty      int    `json:"qty,omitempty"`      // give_item
	Class    string `json:"class,omitempty"`    // move / remove
	X        int    `json:"x,omitempty"`        // move / remove source
	Y        int    `json:"y,omitempty"`
	ToX      int    `json:"to_x,omitempty"` // move destination
	ToY      int    `json:"to_y,omitempty"`
	Gratis   bool   `json:"gratis,omitempty"`
}

// MiracleData is the "miracle" response: the miracle kind, the charge bank after
// landing, whether the charge was waived, and a one-line human rendering.
type MiracleData struct {
	Kind    string `json:"kind"`
	Charges int    `json:"charges"`
	Gratis  bool   `json:"gratis"`
	Summary string `json:"summary"`
}

// StateData answers the "state" command: the full canonical sim.State JSON
// plus the log position it reflects — subscribe with since=last_seq for a
// gapless live replica.
type StateData struct {
	State   json.RawMessage `json:"state"`
	LastSeq int64           `json:"last_seq"`
}

// StatusData is the shared response shape for status/pause/resume/set_speed.
type StatusData struct {
	World  WorldStatus  `json:"world"`
	Clock  ClockStatus  `json:"clock"`
	Daemon DaemonStatus `json:"daemon"`
	Log    LogStatus    `json:"log"`
	// LLM is present only when the world has an orchestrator (llm.json).
	LLM *llm.Status `json:"llm,omitempty"`
	// Warning is set ONLY on the set_speed reply path (spec 035 FR-002; spec 039
	// FR-003): up to two newline-joined advisory texts. First, the spec 039
	// teaching-posture override — on a teaching world whose requested speed
	// exceeds the planner-safe posture rung, the router's per-class horizon
	// arithmetic plus its degrade consequence (fires for calibrated AND
	// uncalibrated teaching worlds). Second, the spec 035 uncalibrated warning —
	// a bootstrap-seeded provider's class suppressed at its current estimate.
	// Never set on status/pause/resume — omitempty keeps those replies
	// byte-identical (FR-008), and non-teaching worlds never carry the posture
	// text. The warning never blocks the speed change (decision-4; the max-speed
	// rejection is a separate hard error), and the change already applied by the
	// time this is set.
	Warning string `json:"warning,omitempty"`
	// Posture is the teaching world's effective speed posture (spec 039 US4,
	// contracts/posture.md §4): the planner-safe rung and its calibrated-vs-
	// provisional provenance, recomputed per reply from the planner-serving
	// provider's live estimate. Present ONLY for a teaching world with an
	// orchestrator (precedent: Horizon) — omitempty keeps every other reply
	// (non-teaching, pure-sim) byte-identical (FR-006/FR-008). The carrier
	// TASK-68 stage presets read; nothing here re-derives calibration by hand.
	Posture *PostureStatus `json:"posture,omitempty"`
	// Horizon is the live per-class cognition horizon (spec 037,
	// contracts/status-horizon.md): one entry per watched class the router
	// would evaluate at the CURRENT effective speed under live estimates.
	// Present ONLY for a world with an orchestrator (composed in
	// statusDataFull); a no-LLM world's reply is byte-identical to pre-037
	// output (omitempty; the composer never runs). Never an empty slice —
	// either absent or ≥1 entry.
	Horizon []HorizonClass `json:"horizon,omitempty"`
	// Resolved skin display facts (spec 052 FR-012, contract §7): the
	// boot-frozen world skin, resolved daemon-side so clients render skin
	// vocabulary without ever reading world files. Identity fields are always
	// sent by a post-052 daemon (resolved against the compiled default
	// table); SkinStrings/SkinStages carry only a world skin's overrides.
	// Additive omitempty — absent fields (a pre-052 daemon) mean the default
	// Guardian skin, so old daemons and old clients interoperate unchanged.
	SkinName        string                        `json:"skin_name,omitempty"`
	SkinEpithet     string                        `json:"skin_epithet,omitempty"`
	SkinTabLabel    string                        `json:"skin_tab_label,omitempty"`
	SkinFamilyLabel string                        `json:"skin_family_label,omitempty"`
	SkinStrings     map[string]string             `json:"skin_strings,omitempty"`
	SkinStages      map[string]skin.StageIdentity `json:"skin_stages,omitempty"`
}

// HorizonClass is one watched decision class's live standing on the status
// reply (spec 037, contracts/status-horizon.md). Entries follow
// cognition.WatchedClasses() order; a class whose kind has no admissible
// serving provider is omitted. Clients render Verdict verbatim and must not
// parse it. Calibrated classes are INCLUDED (contrast: the set_speed Warning,
// which stays gated to uncalibrated providers).
type HorizonClass struct {
	Class           string `json:"class"`
	Suppressed      bool   `json:"suppressed"`
	Verdict         string `json:"verdict"`          // cognition.Verdict.Arithmetic verbatim
	Calibrated      bool   `json:"calibrated"`       // serving provider has a calibration-profile entry
	SuppressedCount int64  `json:"suppressed_count"` // daemon-lifetime router suppressions for this class
}

// PostureStatus is the teaching world's effective speed posture on the
// status-family reply (spec 039 US4, contracts/posture.md §4). Rung is the
// current planner-safe ladder speed ("1x"…"32x", clamped to "1x" when even 1x
// suppresses), recomputed per reply. Calibrated mirrors the serving provider's
// CalibratedAt != "" — the same provenance predicate as spec 035/037; false
// means the rung is the pessimistic bootstrap derivation (provisional).
type PostureStatus struct {
	Rung       string `json:"rung"`
	Calibrated bool   `json:"calibrated"`
}

type WorldStatus struct {
	Name          string `json:"name"`
	Seed          uint64 `json:"seed"`
	FormatVersion int    `json:"format_version"`
	// Stage is the world's curriculum-ladder stage (spec 046, FR-002): additive
	// omitempty, composed straight from world.Manifest.Stage — absent for every
	// pre-046 and pre-ladder world, so their status bytes are unchanged.
	Stage string `json:"stage,omitempty"`
	// StageOverridden mirrors world.Manifest.StageOverridden (spec 046, FR-003):
	// true when the world was created at an unearned stage via --override.
	StageOverridden bool `json:"stage_overridden,omitempty"`
}

type ClockStatus struct {
	Tick          int64   `json:"tick"`
	GameTime      string  `json:"game_time"`
	Paused        bool    `json:"paused"`
	Speed         string  `json:"speed"`
	EffectiveRate float64 `json:"effective_rate"`
	Degraded      bool    `json:"degraded"`
	// MetatronCharges is the nudge bank (TASK-12) — surfaced here so
	// clients render ⚡ without a state fetch.
	MetatronCharges int `json:"metatron_charges"`
	// Adaptive-throttle governor surface (spec 028 US1/US2). All three are
	// additive-omitempty so pre-028 status bytes are byte-identical when the
	// governor is inert (no-LLM worlds, or an LLM world with zero pending
	// debt). RequestedSpeed is the player's ceiling from sim state — empty when
	// ungoverned (requested == effective). GovernorDebt/GovernorJobs are folded
	// from the daemon governor snapshot, like the LLM StatusSnapshot.
	RequestedSpeed string  `json:"requested_speed,omitempty"`
	GovernorDebt   float64 `json:"governor_debt,omitempty"`
	GovernorJobs   int     `json:"governor_jobs,omitempty"`
	// Run-end posture (spec 044 US1): additive omitempty, so pre-044 worlds
	// and old clients are byte-compatible (the governor-trio precedent
	// above). Ended mirrors sim State.Ended; EndedDay is the game day of the
	// run end, for human rendering without a state fetch. Together with the
	// durable run.ended event and StateData (State.Ended/RunEnd ride the
	// canonical state JSON), this is the machine-readable fail signal
	// TASK-119's scenario machinery consumes (FR-005).
	Ended    bool  `json:"ended,omitempty"`
	EndedDay int64 `json:"ended_day,omitempty"`
}

type DaemonStatus struct {
	Pid           int   `json:"pid"`
	UptimeSeconds int64 `json:"uptime_seconds"`
	Subscribers   int   `json:"subscribers"`
}

type LogStatus struct {
	LastSeq int64 `json:"last_seq"`
}

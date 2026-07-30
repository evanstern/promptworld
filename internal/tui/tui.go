// Package tui is the attachable Bubble Tea client: a live view over a world
// replica maintained by log shipping — initial state via the protocol
// "state" command, then subscribed events applied through the same
// sim.State reducer the daemon runs.
//
// TASK-34 widescreen redesign: at width >= widescreenBreakpoint the app
// renders the composite home page (map ‖ dock, docs/design/tui/pages/home.md)
// instead of the single-pane-at-a-time UI; below it, today's single-pane UI
// renders unchanged (docs/design/tui/pages/solo-views.md "Narrow fallback").
// The focus contract (docs/design/tui/patterns/focus-contract.md) replaces
// the old "the guardian console owns the keyboard while active" rule, which
// silently swallowed 1-4/q/space once pane 3 was entered.
package tui

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/guardian"
	"github.com/evanstern/promptworld/internal/ipc"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/skin"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/world"
	"github.com/evanstern/promptworld/internal/worldmap"
	"github.com/evanstern/promptworld/internal/worlds"
)

// pane names both the narrow-fallback's single active pane and the
// widescreen dock's selected tab — paneMap is narrow-only (the widescreen
// map is always visible, never a dock tab); the dock only ever selects
// paneChronicle/paneGuardian/paneVillagers/paneSystems — plus paneExercise
// on scenario worlds (spec 054 — panels/exercise.md: presence is
// world-shaped, not stage-shaped; ambient worlds have no exercise tab at
// all). paneSystems (spec 053, D10) is the relocated-telemetry tab: the
// guardian tab keeps fiction-layer content only from this feature forward —
// the skin boundary is now a file boundary (systems.md carries zero skin
// tokens, guardian.md all of them).
type pane int

const (
	paneMap pane = iota
	paneChronicle
	paneGuardian
	paneVillagers
	paneSystems
	paneExercise
	paneCount
)

// paneNames are the DEFAULT-skin pane labels; the guardian pane's label is
// skin data (spec 052 FR-007), so live surfaces render it via paneName —
// this array is the static fallback the help content and tests read.
var paneNames = [paneCount]string{"map", "chronicle", "guardian", "villagers", "systems", "exercise"}

// paneName resolves a pane's display label through the world skin (spec 052):
// the guardian tab's label is the skin's tab_label token; every other pane is
// non-fiction chrome and stays literal (skin-tokens.md rule 5) — the systems
// tab deliberately so (D10: telemetry is never skinned), the exercise tab
// likewise (spec 054: substrate vocabulary, not fiction).
func (m Model) paneName(p pane) string {
	if p == paneGuardian {
		return m.sk().TabLabel()
	}
	return paneNames[p]
}

// dockTabKey is the keymap.md key that selects/solos each dock tab. The
// exercise tab takes 6 beside the systems tab's 5 (spec 054: dock digits
// continue past TASK-125's systems tab).
var dockTabKey = map[pane]string{paneChronicle: "2", paneGuardian: "3", paneVillagers: "4", paneSystems: "5", paneExercise: "6"}

// paneKey is the narrow tab row's key label per pane — 1–5 plus the
// exercise tab's 6 (the narrow fallback shares keymap.md's global digits).
var paneKey = [paneCount]string{"1", "2", "3", "4", "5", "6"}

// takeoverKind is the body-slot owner among the takeover family (spec 056,
// data-model.md "Takeover owner"; contracts/takeovers.md §1): none, the
// stage-unlock ceremony, or the run-end postmortem. Precedence is total —
// postmortem always wins, replacing an open ceremony; takeovers never
// stack (same-kind arrivals replace, never queue).
type takeoverKind int

const (
	takeoverNone takeoverKind = iota
	takeoverCeremony
	takeoverPostmortem
)

// speedSteps is the [ / ] cycling order.
// max is deliberately absent: the watchable ladder tops out at 32x (TASK-20);
// uncapped ticking is for headless pure-sim runs only.
var speedSteps = []clock.Speed{clock.Speed1x, clock.Speed4x, clock.Speed8x, clock.Speed16x, clock.Speed32x}

const chronicleCap = 500

// Model is the Bubble Tea model. All protocol calls run inside tea.Cmds so
// the UI never blocks on the socket.
type Model struct {
	w *world.World

	client    *ipc.Client
	connected bool
	lastErr   string
	fatalErr  string // unrecoverable (e.g. reply over protocol cap): quit, don't retry

	replica *sim.State      // world replica, event-sourced client-side
	gameMap *worldmap.Map   // terrain, regenerated locally from the manifest
	status  *ipc.StatusData // latest clock/daemon status (1s poll)
	events  []store.Event   // chronicle ring, newest last
	lastSeq int64

	// worldSkin is the attached world's display skin, rebuilt from the
	// status-carried facts (spec 052 FR-012) — never read from world files,
	// and no global client state (edge case: two skins, one terminal). nil
	// (old daemon, absent fields, or not yet connected) is the default
	// Guardian skin: every skin.Skin method is nil-safe.
	worldSkin *skin.Skin

	// unlocks is the per-user curriculum-ladder unlocks record, loaded ONCE
	// at boot (New(), the populateHelpLessons/LoadLessonsSeen precedent —
	// spec 078 FR-006): the help overlay's forward-ladder section unions
	// this boot-loaded snapshot with the live replica.StagesUnlocked at
	// render time rather than re-reading disk every frame. worlds.LoadUnlocks
	// is load-tolerant/nil-safe (missing/corrupt/unresolvable home ⇒ nothing
	// beyond the stage-1 floor earned), so this field is never nil.
	unlocks *worlds.Unlocks

	// active is the narrow fallback's single visible pane (today's model,
	// unchanged). dockTab/solo are the widescreen composite's dock
	// selection and zoom state (pages/solo-views.md). Both are kept in
	// sync on tab-select so a resize across the breakpoint always shows
	// whatever was last looked at, without either being reset by resize.
	active        pane
	dockTab       pane // paneChronicle by default (dock.md: "Default tab on launch")
	solo          bool // dockTab is zoomed to full width (pages/solo-views.md)
	width, height int
	panX, panY    int // map-pane camera offset from the wanderer centroid
	quitting      bool

	// Chronicle pane filters (TASK-11): narrated entries filtered by agent
	// and thread; chronRaw falls back to the raw event feed.
	chronAgent  int // -1 = all
	chronThread string
	chronRaw    bool

	// Chronicle inspect mode (TASK-34, panels/chronicle.md; detail pane
	// TASK-60 spec 018 US2): entered automatically whenever the clock is
	// paused and the chronicle is visible. Selection indexes the raw feed
	// (events); remembered across tab switches, cleared on resume.
	// chronDetailScroll is the always-on detail pane's own scroll offset
	// (contract §5/R6/R7) — reset to 0 on selection move, pause exit, and
	// reconnect (data-model.md "Interaction state"); the render-time clamp
	// to the pane's actual content length lives in chronicleInspectBody,
	// the same defensive-tolerance pattern chronSelectionBase uses.
	chronSelected     int // -1 = none
	chronDetailScroll int

	// Guardian (TASK-12, re-surfaced as the minibuffer by TASK-34): the
	// transcript is dock/pane content; mbInput/mbFocused/mbBusy are the
	// minibuffer's own state, governed by the focus contract
	// (patterns/focus-contract.md) everywhere it appears.
	transcript     []string               // rendered transcript rows, newest last
	consoleCharter string                 // "default charter" | "custom charter" | ""
	consoleSkills  int                    // count of effective skill files (spec 021 US3)
	consoleTools   string                 // granted-tool summary, e.g. "tools: dream, omen"; "" when quiet default
	consoleOrders  []guardian.OrderStatus // standing orders peek (spec 029 T023, FR-016)
	consoleStage   string                 // curriculum-ladder stage line (spec 046 T010); "" for a pre-ladder/ungated world

	// Charter/skills read-surface provenance (spec 053 US3, FR-004, research
	// R5): the raw stage-lock fields consoleStage's formatted line already
	// folds together, kept separately here because the read surface (the
	// guardian console's own bordered sub-panel) needs to distinguish
	// "preset-locked" from "player-authored" as its own line, honestly
	// naming the unlocking stage — the same fields guardian.Status carries,
	// no new client-side file parsing.
	consoleCharterLocked bool
	consoleCharterPreset string
	consoleSkillsLocked  bool

	// Guardian console page state (spec 053, data-model.md "Console page
	// state"): a page-level toggle, not a dock tab or a solo zoom — dockTab/
	// solo/active are never touched while console is true, so they already
	// ARE the "prior view" the data model calls consoleReturn; closing the
	// console needs nothing to restore beyond flipping the flag back.
	// consoleScroll is tail-anchored scrollback (0 = the tail/most recent),
	// reset on close; consoleNotice is the one-shot post-$EDITOR line
	// (research R2), also cleared on close. consoleCards is the card-
	// composition seam (research R6, FR-006) — always empty this feature;
	// TASK-127/115 are the producers.
	console       bool
	consoleScroll int
	consoleNotice string
	consoleCards  []consoleCard

	mbFocused bool
	mbInput   string
	mbBusy    bool
	mbErr     string
	mbHistory []string
	mbHistPos int    // index into mbHistory while cycling; == len(mbHistory) means "the live draft"
	mbDraft   string // input stashed when history-cycling away from an in-progress draft
	mbFlash   string // one-shot dormant-state message (minibuffer.md "answer arrived — 3 to read")

	guardianUnseen bool // dock tab badge: a reply landed while the tab wasn't visible

	// Villagers tab (TASK-56, data-model.md "New TUI model state"):
	// villSelected is the roster cursor, clamped to [0, len(replica.Agents))
	// wherever read (the replica can arrive late or be swapped wholesale on
	// reconnect); villDetail opens the selected villager's detail view.
	// Neither field is persisted — client-only, event-sourced from nothing.
	villSelected int
	villDetail   bool

	// villDecisions (spec 020, TASK-63): the decisions sub-view, meaningful
	// only while villDetail is true. villDecisionsScroll is its own scroll
	// offset, reset on villager change, detail close, and reconnect
	// (data-model.md "New Model (TUI) state") and defensively re-clamped
	// again at render time (villagerDecisionsBody), the same pattern
	// chronDetailScroll uses.
	villDecisions       bool
	villDecisionsScroll int

	// traces (spec 020, TASK-63) is the bounded per-agent decision-trace
	// projection (data-model.md), fed from applyEvent alongside the replica
	// fold and chronicle ring — independent of both, so it survives the
	// ring's 500-event eviction (SC-003) but resets wholesale on reconnect
	// like the replica (contract R5).
	traces decisionTraces

	// lessons (spec 055, TASK-117, internal/tui/lessons.go): the first-
	// occurrence teaching projection — loaded once at construction
	// (per-user seen-state, internal/worlds/lessons.go) and folded on every
	// pushed event via applyEvent, alongside traces above. Unlike traces,
	// this does NOT reset on reconnect (contract: per-user, cross-world,
	// cross-restart — a reconnect within the same client run must not
	// re-show anything already surfaced this session).
	lessons lessonTriggers

	// stageOverrides (spec 066, TASK-128, data-model.md "surfaceOverrides"):
	// an in-session, per-surface visibility choice the player made
	// explicitly, outranking stage re-resolution until the session ends
	// (never persisted — applyOverrides, stagedefaults.go). No production
	// key currently sets an entry here (no in-session command exists yet to
	// toggle a governed surface); the field and its precedence rule are
	// still carried so a future toggle command has a session-scoped place
	// to land without another re-resolution mechanism.
	stageOverrides surfaceOverrides

	// Help overlay (spec 045, TASK-116; data-model.md "Model state"): a
	// client-only presentation layer, head of the esc-release chain while
	// open (help -> minibuffer -> decisions -> detail -> solo -> home).
	// helpMode freezes which mode '?' was pressed from (spec edge case:
	// "the context...cannot drift while it is open" — R1); helpPageMode is
	// the mode page actually being shown, starting equal to helpMode but
	// separately paged with n/p so every mode's page (including
	// minibuffer's, otherwise unreachable — FR-001) stays reachable. All
	// five fields reset on dismissal; nothing else ever changes while the
	// overlay is open, so dismissal has nothing else to restore (FR-006).
	helpOpen     bool
	helpMode     helpModeKey
	helpPageMode helpModeKey
	helpTier     bool // false = basic, true = advanced
	helpSection  helpSection
	helpScroll   int

	// exBriefingDismissed (spec 054 US4, panels/exercise.md): the exercise
	// tab's attach-time briefing has been dismissed for THIS attach — any
	// key while the briefing is visible consumes that one keypress and sets
	// it (contract §4); reset on reconnect (connectedMsg), so re-attaching
	// shows the briefing again. Meaningless on ambient worlds (no exercise
	// tab exists).
	exBriefingDismissed bool

	// chronHit (spec 049, data-model.md "chronHitRegion"; T009) is the
	// chronicle inspect-mode list's last-rendered click geometry — a
	// pointer so a value-receiver View() can still record it: Bubble Tea
	// copies Model by value on every Update, but every copy carries the
	// same pointer forward, so a write through it is visible to the very
	// next Update() regardless of which copy performed the write. Recorded
	// by chronicleInspectBody each View(), consumed by handleMouse the
	// next Update() (contract §1, US2).
	chronHit *chronHitRegion

	// Takeover state (spec 056, data-model.md "Takeover owner"; contracts/
	// takeovers.md §1): which overlay owns the body slot, above every other
	// mode/page in this client (help, console, solo, narrow panes alike) —
	// applyEvent flips it on run.ended/curriculum.stage_unlocked, esc/q
	// (handleTakeoverKey) release it, connectedMsg auto-opens the postmortem
	// on attach to an already-ended world. ceremonyDeferred records that an
	// unlock arrived while the postmortem owned the slot (postmortem always
	// wins — the ceremony never interrupts it, only replay surfaces carry
	// the deferred content forward). postmortemDismissed is per-client-
	// session (not per reconnect — a transient resync should not re-annoy a
	// player who already dismissed it): esc sets it, `p` clears it, New()
	// starts false.
	takeover            takeoverKind
	ceremonyDeferred    bool
	postmortemDismissed bool

	// Look-cursor mode (spec 074-look-cursor, data-model.md "Look-cursor
	// mode state"): a client-side-only tile inspector layered on the map —
	// never persisted, never entering the replica. lookActive gates the key
	// layer (look.go's handleLookKey), the dock borrow (dockTabsRow/
	// dockTabContent), the cursor render (renderMapGrid), and the
	// !lookActive visibility guards below (research R1). lookX/lookY are
	// the cursor tile; lookFocus is which layer inside the mode holds the
	// keyboard; lookSel is the selected TILE-pane row (into the flattened
	// row list, look.go's tileRows); lookDrill/lookDrillScroll are the open
	// drill-in target and its own scroll offset.
	lookActive      bool
	lookX, lookY    int
	lookFocus       lookFocusKind
	lookSel         int
	lookDrill       lookDrillRef
	lookDrillScroll int

	// mapHit/tileHit (research R6) are the map grid's and the TILE pane's
	// last-rendered click geometry — pointers for the same value-receiver
	// View()-writes-through-a-pointer reason chronHit exists (tui.go:284).
	mapHit  *mapHitRegion
	tileHit *tileHitRegion

	// stripHit/rosterHit (spec 086 US5 — the reverse-jump rider) are the
	// villager strip's and the villagers-tab roster's last-rendered click
	// geometry, same pointer/frame-top-invalidation pattern as chronHit:
	// recorded by villagerStripView/villagerRosterBody each View(), consumed
	// by handleMouse the next Update(). Clicking a strip glyph or a roster
	// row centers the map camera on that villager (centerCameraOn — the
	// spec 049 camera writer), the village lens's reverse direction.
	stripHit  *stripHitRegion
	rosterHit *rosterHitRegion
}

// stripHitRegion is the villager strip's rendered glyph geometry: the strip's
// screen row and each rendered glyph's screen column (glyphs shed from the
// end under width overflow — the hit region covers RENDERED glyphs only; the
// …N overflow marker is never a target). Widescreen-only (the strip never
// renders narrow).
type stripHitRegion struct {
	valid   bool
	originY int   // the strip's screen row (headerRows while the strip renders)
	glyphX  []int // glyphX[i] = screen column of villager i's glyph; single-column glyphs
}

// rosterHitRegion is the villagers-tab roster's rendered row-band geometry —
// the chronHitRegion sibling for the roster body. rowAgent maps each rendered
// body line to its villager index (-1 = heading/blank/spacer, consumed but
// not a row).
type rosterHitRegion struct {
	valid            bool
	originX, originY int   // screen cell of the roster body's first line
	width            int   // column span of a hit
	rowAgent         []int // rowAgent[i] = agent index for screen row originY+i; -1 = not a row
}

// chronHitRegion is the chronicle inspect-list's rendered geometry: which
// screen rows the list currently occupies and which event each row shows.
// valid is false whenever the chronicle inspect list wasn't part of the
// frame just rendered (other tab, running mode, narrow non-chronicle pane,
// help overlay) — View() invalidates by default and chronicleInspectBody
// re-validates only when it actually runs (data-model.md "State
// transitions").
type chronHitRegion struct {
	valid            bool
	originX, originY int   // top-left screen cell of row 0's line
	width            int   // column span of a hit; row span is len(rowEvent)
	rowEvent         []int // rowEvent[i] = m.events index for screen row originY+i
}

func New(w *world.World) Model {
	// Lessons pull half (spec 055 T012, SC-002): populate the help overlay's
	// lessons section from the same catalog the row (push half) reads —
	// one catalog, two surfaces, never two hand-maintained lists. Default-
	// skin resolution here (no status yet); re-resolved with the world skin
	// once status carries it (skinFromStatus call sites in Update).
	populateHelpLessons(nil)
	return Model{
		w: w, gameMap: w.Map(), chronAgent: -1, dockTab: paneChronicle, chronSelected: -1,
		traces: newDecisionTraces(), chronHit: &chronHitRegion{},
		mapHit: &mapHitRegion{}, tileHit: &tileHitRegion{}, // spec 074: the chronHit pointer pattern
		stripHit: &stripHitRegion{}, rosterHit: &rosterHitRegion{}, // spec 086: the reverse-jump rider
		// Per-user seen-state (spec 055 FR-006): load-tolerant, advisory —
		// worlds.LoadLessonsSeen() never errors, degrading to an empty
		// record on a missing/corrupt file or an unresolvable home dir.
		lessons: newLessonTriggers(lessonSeenIDs(worlds.LoadLessonsSeen())),
		// Per-user unlocks record (spec 078 FR-006): loaded once here, same
		// load-tolerant doctrine as the lessons-seen record above.
		unlocks: worlds.LoadUnlocks(),
	}
}

// lessonSeenIDs adapts the persisted per-user record (internal/worlds) to
// the plain id-set lessonTriggers is constructed from (lessons.go) — the one
// place this package touches worlds.LessonsSeen's shape.
func lessonSeenIDs(seen *worlds.LessonsSeen) map[string]bool {
	if seen == nil {
		return nil
	}
	ids := make(map[string]bool, len(seen.Entries))
	for id := range seen.Entries {
		ids[id] = true
	}
	return ids
}

// FatalErr reports the unrecoverable error that made the TUI quit, if any —
// the ui command surfaces it as a real exit error after Run returns.
func (m Model) FatalErr() string { return m.fatalErr }

// sk is the attached world's skin (nil = default; skin.Skin is nil-safe) —
// the ONE door every fiction string in this package renders through
// (spec 052 FR-007).
func (m Model) sk() *skin.Skin { return m.worldSkin }

// skinFromStatus rebuilds the world skin from the status-carried display
// facts (spec 052 contract §7). A status with no override maps — including
// every pre-052 daemon's — is the default skin (nil).
func skinFromStatus(st *ipc.StatusData) *skin.Skin {
	if st == nil || (st.SkinStrings == nil && st.SkinStages == nil) {
		return nil
	}
	return skin.FromFacts(st.SkinStrings, st.SkinStages)
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(connect(m.w), pollTick())
}

// --- messages ---

type connectedMsg struct {
	client  *ipc.Client
	status  *ipc.StatusData
	replica *sim.State
	lastSeq int64
}

type disconnectedMsg struct{ err error }

type pushMsg struct{ push ipc.Push }

type statusMsg struct{ status *ipc.StatusData }

type consoleReplyMsg struct {
	result *guardian.TurnResult
	err    error
}

type consoleStatusMsg struct{ status *guardian.Status }

// editorResultMsg reports the $EDITOR round trip's outcome (spec 053 US3,
// research R2) after tea.ExecProcess restores the TUI: changed is the
// pre/post content-hash comparison's verdict; err carries a nonzero exit or
// exec failure (edge case: "$EDITOR exits nonzero: treat as no-change...
// plus an honest one-line notice").
type editorResultMsg struct {
	changed bool
	err     error
}

// consoleToolsSummary renders the granted-tool part of the console header
// (spec 021 US3, contracts/status.md): quiet (empty) for a full-grant default
// world, "tools: none" for a conversation-only world, else "tools: " + the
// granted set in short form (dream, omen, workings) with any kind restriction
// carried through. Kept beside the consoleStatusMsg handler — the only TUI
// region this feature touches (TASK-63 owns the digest/villager/transcript
// regions).
func consoleToolsSummary(s *guardian.Status, sk *skin.Skin) string {
	if s.ManifestDefault {
		return "" // full grant is the unremarkable default — keep the header quiet
	}
	if len(s.GrantedTools) == 0 {
		return "tools: none"
	}
	parts := make([]string, 0, len(s.GrantedTools))
	for _, t := range s.GrantedTools {
		parts = append(parts, shortToolName(t, sk))
	}
	return "tools: " + strings.Join(parts, ", ")
}

// consoleStageSummary renders the guardian pane's curriculum-ladder line
// (spec 046 T010, R10): the skin's display name for the world's stage plus
// the lock provenance the ceiling/instruction-lock (US2) already computed —
// "" for a pre-ladder/ungated world (Stage absent), keeping every existing
// world's console header byte-identical (the consoleToolsSummary precedent).
func consoleStageSummary(s *guardian.Status, sk *skin.Skin) string {
	if s.Stage == "" {
		return ""
	}
	line := "stage: " + sk.StageName(s.Stage)
	if s.CharterLocked {
		line += " (charter locked to " + s.CharterPreset + ")"
	}
	return line
}

// shortToolName maps a granted-tool label to its console short form, preserving
// any kind restriction suffix on the frozen work_miracle tool id — the
// DISPLAY vocabulary is the skin's working noun (spec 052 FR-007;
// "work_miracle(move,give_item)" → "workings(move,give_item)").
func shortToolName(label string, sk *skin.Skin) string {
	switch {
	case label == "nudge_dream":
		return "dream"
	case label == "nudge_omen":
		return sk.FormNoun("omen")
	case label == "work_miracle":
		return sk.WorkingNounPlural()
	case strings.HasPrefix(label, "work_miracle("):
		return sk.WorkingNounPlural() + strings.TrimPrefix(label, "work_miracle")
	default:
		return label
	}
}

// fetchConsoleStatus grabs the model-free peek when the guardian tab/pane is
// selected.
func fetchConsoleStatus(c *ipc.Client) tea.Cmd {
	return func() tea.Msg {
		st, err := c.GuardianStatus()
		if err != nil {
			return consoleStatusMsg{}
		}
		return consoleStatusMsg{status: st}
	}
}

type pollMsg struct{}

type retryMsg struct{}

// --- commands ---

// connect dials, fetches the state snapshot, and subscribes from exactly the
// seq that snapshot reflects — the replica starts gapless.
func connect(w *world.World) tea.Cmd {
	return func() tea.Msg {
		c, err := ipc.Dial(w.SockPath())
		if err != nil {
			return disconnectedMsg{err}
		}
		sd, err := c.FetchState()
		if err != nil {
			c.Close()
			return disconnectedMsg{err}
		}
		replica := sim.NewState(w.Manifest.Seed, w.Map())
		if err := json.Unmarshal(sd.State, replica); err != nil {
			c.Close()
			return disconnectedMsg{fmt.Errorf("state decode: %w", err)}
		}
		st, err := c.Status("status", nil)
		if err != nil {
			c.Close()
			return disconnectedMsg{err}
		}
		since := sd.LastSeq
		if err := c.Subscribe(&since); err != nil {
			c.Close()
			return disconnectedMsg{err}
		}
		return connectedMsg{client: c, status: st, replica: replica, lastSeq: sd.LastSeq}
	}
}

// listen delivers one push per invocation; Update re-arms it.
func listen(c *ipc.Client) tea.Cmd {
	return func() tea.Msg {
		p, ok := <-c.Pushes()
		if !ok {
			return disconnectedMsg{fmt.Errorf("connection lost")}
		}
		return pushMsg{p}
	}
}

func pollTick() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg { return pollMsg{} })
}

func retryLater() tea.Cmd {
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg { return retryMsg{} })
}

func fetchStatus(c *ipc.Client) tea.Cmd {
	return func() tea.Msg {
		st, err := c.Status("status", nil)
		if err != nil {
			return disconnectedMsg{err}
		}
		return statusMsg{st}
	}
}

// sendConsole runs one Guardian turn off the UI goroutine.
func sendConsole(c *ipc.Client, text string) tea.Cmd {
	return func() tea.Msg {
		r, err := c.GuardianChat(text)
		return consoleReplyMsg{result: r, err: err}
	}
}

func timeControl(c *ipc.Client, cmd string, args any) tea.Cmd {
	return func() tea.Msg {
		st, err := c.Status(cmd, args)
		if err != nil {
			return disconnectedMsg{err}
		}
		return statusMsg{st}
	}
}

// --- update ---

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		// Resizing across the widescreen breakpoint swaps layouts live;
		// no field here is reset, so no state is lost (pages/solo-views.md
		// "Narrow fallback"). clampGeometry re-bounds the handful of
		// fields that persist *values* across frames (pan offset,
		// chronicle selection) so nothing computed at the old size can
		// push a panel off-frame at the new one (B5) — everything else
		// (dock tab, solo, filters, transcript) is size-independent by
		// construction and needs no clamping.
		m.width, m.height = msg.Width, msg.Height
		m.clampGeometry()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.MouseMsg:
		return m.handleMouse(msg)

	case connectedMsg:
		if m.client != nil {
			m.client.Close()
		}
		m.client = msg.client
		m.connected = true
		m.lastErr = ""
		m.status = msg.status
		m.worldSkin = skinFromStatus(msg.status) // spec 052: skin rides status
		// Re-resolve the help overlay's lesson bodies under the world skin
		// (spec 055's planned per-world resolution seam): the skin is
		// boot-frozen daemon-side, so a change can only arrive through a
		// reconnect — exactly this handler.
		populateHelpLessons(m.worldSkin)
		m.replica = msg.replica
		m.lastSeq = msg.lastSeq
		m.rebuildConsoleCards()        // spec 063: a late attach re-reads the stored card (no badge — nothing fresh)
		m.clampVillSelected()          // R5: connectedMsg swaps the replica wholesale
		m.chronDetailScroll = 0        // data-model.md: detail pane scroll resets on reconnect
		m.villDecisionsScroll = 0      // data-model.md: decisions scroll resets on reconnect
		m.traces = newDecisionTraces() // spec 020 contract R5: projection resets wholesale, like the replica
		m.exBriefingDismissed = false  // spec 054: the briefing shows once per attach — a fresh orientation each session
		// Postmortem auto-open (spec 056 FR-001, contracts/takeovers.md §1):
		// the dual-source runEnded() posture — a fresh attach to an
		// already-ended world (State.Ended in the snapshot, no live
		// run.ended replay needed) opens the takeover on connect exactly
		// like the live transition does, unless this client session already
		// dismissed it (postmortemDismissed is per-session, not per
		// reconnect — a transient resync must not re-annoy). Postmortem
		// always wins even here, replacing any stale ceremony left open
		// across a reconnect.
		if m.runEnded() && !m.postmortemDismissed {
			m.takeover = takeoverPostmortem
		}
		return m, listen(m.client)

	case disconnectedMsg:
		m.connected = false
		if m.client != nil {
			m.client.Close()
			m.client = nil
		}
		if errors.Is(msg.err, ipc.ErrReplyTooLarge) {
			// Reconnecting cannot shrink the state — retrying forever would
			// be the TASK-19 bug. Fail fast with the actionable message.
			m.fatalErr = msg.err.Error()
			m.quitting = true
			return m, tea.Quit
		}
		m.lastErr = msg.err.Error()
		return m, retryLater()

	case retryMsg:
		if m.connected || m.quitting {
			return m, nil
		}
		return m, connect(m.w)

	case pushMsg:
		return m.handlePush(msg.push)

	case statusMsg:
		wasPaused := m.status != nil && m.status.Clock.Paused
		hadStatus := m.status != nil // spec 066 T011: a genuine baseline to diff against, not the first poll
		prevStage := m.currentStage()
		m.status = msg.status
		m.worldSkin = skinFromStatus(msg.status) // spec 052: boot-frozen, but a daemon restart re-skins
		// Spec 104: the status poll is the replica's tick driver between
		// events — walk derived progress (in-flight walk segments, needs
		// decay, gru motion) up to the daemon's reported tick so the live
		// map stays per-step smooth (plan D5: exact, not interpolated).
		// AdvanceTo(T) is the daemon's own posture at tick T (items ≤ T−1),
		// so the replica can never lead the daemon's fold.
		if m.replica != nil && m.status != nil {
			m.replica.AdvanceTo(m.status.Clock.Tick)
		}
		nowPaused := m.status != nil && m.status.Clock.Paused
		if wasPaused && !nowPaused {
			// Resume: collapse everything, snap back to tail-follow
			// (panels/chronicle.md Mode 2 "On resume"; contract §5/R7).
			m.chronSelected = -1
			m.chronDetailScroll = 0
		}
		// Stage-shaped layout defaults (spec 066, TASK-128, research.md R4):
		// on a genuine live stage change (never the first status poll —
		// that's boot resolution, already continuous via currentStage()
		// reads at render time, not an "arrival"), re-resolve and route any
		// newly-on governed surface through the existing first-occurrence
		// lesson machinery (FR-005, SC-005). Under the CURRENT authority
		// table this is a no-op on every real transition (see
		// newlyOnSurfaces' doc comment, stagedefaults.go) — kept as real,
		// tested plumbing rather than dead code (TestAnnounceSurfaceArrival
		// ExactlyOnce, stagedefaults_test.go), forward-compatible with a
		// future table revision that does widen a surface going up the
		// ladder. Takeovers (ceremony/postmortem) fire off their own event
		// triggers, independent of this path (FR-008).
		if hadStatus {
			if newStage := m.currentStage(); newStage != prevStage {
				hasScenario := m.exerciseID() != ""
				prevSet := resolveStageDefaults(prevStage, hasScenario)
				nextSet := resolveStageDefaults(newStage, hasScenario)
				for _, id := range newlyOnSurfaces(prevSet, nextSet) {
					announceSurfaceArrival(&m.lessons, id, time.Now())
				}
			}
		}
		return m, nil

	case consoleStatusMsg:
		if msg.status == nil {
			m.consoleCharter = ""
			m.consoleSkills = 0
			m.consoleTools = ""
			m.consoleOrders = nil
			m.consoleStage = ""
			m.consoleCharterLocked = false
			m.consoleCharterPreset = ""
			m.consoleSkillsLocked = false
		} else {
			if msg.status.CharterDefault {
				m.consoleCharter = "default charter"
			} else {
				m.consoleCharter = "custom charter"
			}
			m.consoleSkills = len(msg.status.Skills)
			m.consoleTools = consoleToolsSummary(msg.status, m.sk())
			m.consoleOrders = msg.status.Orders
			m.consoleStage = consoleStageSummary(msg.status, m.sk())
			m.consoleCharterLocked = msg.status.CharterLocked
			m.consoleCharterPreset = msg.status.CharterPreset
			m.consoleSkillsLocked = msg.status.SkillsLocked
		}
		return m, nil

	case consoleReplyMsg:
		m.mbBusy = false
		if msg.err != nil {
			m.mbErr = msg.err.Error()
			return m, nil
		}
		r := msg.result
		for _, mo := range r.Moments {
			m.transcript = append(m.transcript, "! "+mo)
		}
		m.transcript = append(m.transcript, transcriptGuardianPrefix+r.Reply)
		if r.Nudge != nil {
			// Form is the frozen payload value ("vision"/"omen", spec 052
			// FR-005); the DISPLAY noun is skin vocabulary.
			m.transcript = append(m.transcript, fmt.Sprintf("⚡ %s → %s: %q",
				m.sk().FormNoun(r.Nudge.Form), strings.Join(r.Nudge.Targets, ", "), r.Nudge.Text))
		}
		if r.Order != nil {
			m.transcript = append(m.transcript, fmt.Sprintf("👁 watch set (%s): %q", r.Order.ID, r.Order.Condition))
		}
		for _, id := range r.Cancelled {
			m.transcript = append(m.transcript, fmt.Sprintf("👁 watch released: %s", id))
		}
		if r.Clock != "" {
			m.transcript = append(m.transcript, fmt.Sprintf("⏲ %s", r.Clock))
		}
		if len(m.transcript) > 200 {
			m.transcript = m.transcript[len(m.transcript)-200:]
		}
		// Reply arrival (minibuffer.md): stream in place if the guardian
		// tab/pane is visible; otherwise badge the tab and flash the
		// minibuffer once.
		if !m.guardianVisible() {
			m.guardianUnseen = true
			m.mbFlash = "answer arrived — 3 to read"
		}
		return m, nil

	case pollMsg:
		cmds := []tea.Cmd{pollTick()}
		if m.connected && m.client != nil {
			cmds = append(cmds, fetchStatus(m.client))
		}
		// Lessons projection time-advance (spec 055): the queue can decay or
		// promote its head purely from wall-clock time passing, with no new
		// event required — the ~1s poll tick already firing regardless of
		// connection state is the natural driver (research.md R4).
		if surfaced := m.lessons.Advance(time.Now()); surfaced != nil {
			worlds.MarkLessonSeen(surfaced.ID, m.w.Manifest.Name)
		}
		return m, tea.Batch(cmds...)

	case editorResultMsg:
		// $EDITOR round trip landed (spec 053 US3, contract §4): exactly one
		// notice, never both — an exec error/nonzero exit wins over a
		// (moot) content comparison; unchanged clears any stale notice
		// rather than leaving a prior confirmation lingering.
		switch {
		case msg.err != nil:
			m.consoleNotice = "the editor exited with an error — charter.md unchanged"
		case msg.changed:
			m.consoleNotice = "charter changed — next turn binds it"
		default:
			m.consoleNotice = ""
		}
		return m, nil
	}
	return m, nil
}

// --- focus contract (patterns/focus-contract.md) ---

// quit is ctrl+c/q from any unfocused state: rule 3, "ctrl+c quits the app
// from any state whatsoever".
func (m Model) quit() (tea.Model, tea.Cmd) {
	m.quitting = true
	if m.client != nil {
		m.client.Close()
	}
	return m, tea.Quit
}

// chronicleVisible reports whether the chronicle is the thing currently on
// screen, in whichever layout is active — the gate for both the a/t/r
// filter keys and automatic inspect-mode entry. Gains a !lookActive guard
// (spec 074 research R1): while the look-cursor mode borrows the dock body,
// the chronicle tab is NOT the thing visible regardless of m.dockTab — the
// TILE view is (the villagers-mode scoping precedent).
func (m Model) chronicleVisible() bool {
	if m.lookActive {
		return false
	}
	if isWidescreen(m.width) {
		return m.dockTab == paneChronicle
	}
	return m.active == paneChronicle
}

// guardianVisible reports whether the guardian transcript is the thing
// currently on screen — governs whether a reply streams in place or badges
// the tab (minibuffer.md). The guardian console (spec 053 US1 AS4) renders
// the SAME transcript as its primary content, so it counts as "visible" too
// — the console adds no second badge system, it just shares this one.
// Gains a !lookActive guard (spec 074 research R1, after the console check:
// the two states are mutually exclusive — G exits the look mode before
// opening the console, look.go's handleLookKey): while the borrow is
// active, a guardian reply must badge, never stream, exactly like every
// other visibility consumer this feature guards.
func (m Model) guardianVisible() bool {
	if m.console {
		return true
	}
	if m.lookActive {
		return false
	}
	if isWidescreen(m.width) {
		return m.dockTab == paneGuardian
	}
	return m.active == paneGuardian
}

// villagersVisible reports whether the villagers tab is the thing currently
// on screen — the gate for the roster/detail selection keys (contracts/
// state-and-keys.md "Keys bind only while the villagers tab is the visible
// dock tab or solo'd"). Gains a !lookActive guard (spec 074 research R1),
// the chronicleVisible precedent.
func (m Model) villagersVisible() bool {
	if m.lookActive {
		return false
	}
	if isWidescreen(m.width) {
		return m.dockTab == paneVillagers
	}
	return m.active == paneVillagers
}

// exerciseID returns the attached world's scenario exercise id, "" for an
// ambient world — the exercise tab's presence gate (spec 054 FR-008;
// panels/exercise.md and patterns/stage-defaults.md: presence is
// world-shaped — the manifest's Scenario block, which world.Open validated
// against the catalog; the extra ExerciseByID check keeps a stale binary
// honest rather than rendering an empty definition).
func (m Model) exerciseID() string {
	if m.w == nil {
		return "" // zero-value Model (render tests, pre-connect) — ambient
	}
	if sc := m.w.Manifest.Scenario; sc != nil {
		if _, ok := sim.ExerciseByID(sc.Exercise); ok {
			return sc.Exercise
		}
	}
	return ""
}

// exerciseVisible reports whether the exercise tab is the thing currently on
// screen, in whichever layout is active — the villagersVisible shape. Gains
// a !lookActive guard (spec 074 research R1), same precedent.
func (m Model) exerciseVisible() bool {
	if m.exerciseID() == "" || m.lookActive {
		return false
	}
	if isWidescreen(m.width) {
		return m.dockTab == paneExercise
	}
	return m.active == paneExercise
}

// exerciseBriefingShowing reports whether the attach-time briefing is what
// the exercise tab is rendering right now — the gate for the one-keypress
// any-key dismiss (contract §4: never a global key-eater; it consumes only
// while the exercise tab is visible). The guardian console (spec 053) is a
// whole-body takeover checked BELOW this gate in handleKey, so an open
// console explicitly suppresses the eater: the briefing is not the thing
// on screen while the console is, and its keys must reach the console.
func (m Model) exerciseBriefingShowing() bool {
	return !m.console && m.exerciseVisible() && !m.exBriefingDismissed
}

// mapControllable reports whether arrow keys should pan the map: always in
// widescreen (pages/home.md: "regardless of which dock tab is selected"),
// only while the map pane is active in the narrow fallback (unchanged).
func (m Model) mapControllable() bool {
	if isWidescreen(m.width) {
		return true
	}
	return m.active == paneMap
}

// inspecting reports whether inspect mode (panels/chronicle.md Mode 2) is
// live: paused, and the chronicle is the thing on screen.
func (m Model) inspecting() bool {
	return m.status != nil && m.status.Clock.Paused && m.chronicleVisible()
}

// runEnded is the postmortem-posture predicate (spec 044 R12), dual-source by
// necessity, not belt-and-braces: the replica's State.Ended covers clients
// attaching after the fact (the snapshot path never replays folded events),
// while the pushed run.ended (folded into the replica by applyEvent) and the
// 1s status poll cover the live transition without a reconnect. It drives the
// ENDED header token, the inert clock keys, and the footer hint; every
// reading surface stays fully functional.
func (m Model) runEnded() bool {
	if m.replica != nil && m.replica.Ended {
		return true
	}
	return m.status != nil && m.status.Clock.Ended
}

// handleKey is the top-level key dispatcher implementing the modes of
// patterns/keymap.md, in priority order: ctrl+c always quits (rule 3); a
// takeover (spec 056, contracts/takeovers.md §1), when open, owns the
// keyboard before EVERYTHING else — help, minibuffer, console, inspect,
// villagers, global alike — the family's own maximum-salience framing
// ("the takeover IS the event," not a mode alongside the others); the help
// overlay (spec 045), when open, owns the keyboard next — the head of the
// esc-release chain (help -> minibuffer -> villager detail -> console ->
// solo -> home; research.md R1, spec 053 amendment) — and '?' opens it from
// every mode except minibuffer focus, checked immediately after (so '?'
// still types into the buffer there, FR-001); minibuffer-focused keys own
// the keyboard only when focus was explicitly acquired (rule 1); `p` (spec
// 056, ended-only global reopen) is checked next, after minibuffer/help so
// it neither types into the buffer nor breaks help's swallow-everything
// rule (FR-012); the guardian console (spec 053), when open, owns the
// keyboard next — checked ahead of inspect/villagers mode because those are
// keyed off dockTab/active, which persist unchanged (and so would still
// spuriously read as "visible") underneath the console's full-screen body;
// inspect-mode keys layer on top of, never replace, the global mode (rule 5
// / keymap.md "Mode: inspect").
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+c" {
		return m.quit()
	}
	if m.takeover != takeoverNone {
		return m.handleTakeoverKey(msg)
	}
	if m.helpOpen {
		return m.handleHelpKey(msg)
	}
	// Exercise briefing dismiss (spec 054 US4, contract §4): while the
	// briefing is the thing on the exercise tab, ANY key dismisses it for
	// this attach and is consumed — exactly one keypress, only while that
	// tab is visible (never a global key-eater), and never while the
	// minibuffer owns the keyboard (rule 1: focus was explicitly acquired).
	if !m.mbFocused && m.exerciseBriefingShowing() {
		m.exBriefingDismissed = true
		return m, nil
	}
	if !m.mbFocused && msg.String() == "?" {
		return m.openHelp()
	}
	if m.mbFocused {
		return m.handleMinibufferKey(msg)
	}
	if msg.String() == "p" && m.runEnded() {
		// Global postmortem reopen (spec 056, contracts/takeovers.md §1):
		// reachable from every mode reached this far down the chain (console,
		// inspect, villagers, global alike) — inert on a live world.
		// Clearing postmortemDismissed means a later resync/reconnect won't
		// immediately re-suppress a takeover the player just asked to see
		// again.
		m.takeover = takeoverPostmortem
		m.postmortemDismissed = false
		return m, nil
	}
	if m.console {
		if mdl, cmd, handled := m.handleConsoleKey(msg); handled {
			return mdl, cmd
		}
		// Not a console-owned key (contract §1: only G/1/esc/m/e/J/K are) —
		// falls straight to the global mode (space/q/[/]/2-5/pan/a/t/r all
		// still work per the console's own footer hints), deliberately
		// skipping the inspect/villagers layers below: those are scoped to
		// the chronicle/villagers dock tabs, neither of which is the screen
		// actually showing right now.
		return m.handleGlobalKey(msg)
	}
	if m.lookActive {
		// The look-cursor mode (spec 074, research R1): layered between the
		// console and inspect checks, claiming the whole contested key set
		// so the (now-dormant, chronicleVisible/villagersVisible guards)
		// inspect/villagers layers below never spuriously fire during the
		// dock borrow. Unclaimed keys fall through to handleGlobalKey
		// (space/[/]/m/q/p/? — FR-013).
		if mdl, cmd, handled := m.handleLookKey(msg); handled {
			return mdl, cmd
		}
	}
	if m.inspecting() {
		if mdl, cmd, handled := m.handleInspectKey(msg); handled {
			return mdl, cmd
		}
	}
	if m.villagersVisible() {
		if mdl, cmd, handled := m.handleVillagersKey(msg); handled {
			return mdl, cmd
		}
	}
	return m.handleGlobalKey(msg)
}

// --- help overlay (spec 045, TASK-116; contracts/help-content.md "Layering") ---

// currentHelpMode identifies which of the six help pages (help.go,
// data-model.md) the client is in right now — the same predicate order
// footerView uses, so the frozen mode always matches the hint the footer
// was just showing when '?' was pressed. Never returns helpModeMinibuffer
// (unreachable as an *opened-from* mode — FR-001, R1: '?' types there
// instead of opening the overlay); that page is only reachable by paging
// with n/p from another mode's overlay.
func (m Model) currentHelpMode() helpModeKey {
	switch {
	case m.console:
		// The console page (spec 053) has no dedicated help mode of its own
		// (T009's scope is new rows on the existing global/solo pages, not a
		// seventh mode) — checked first because dockTab/active persist
		// unchanged underneath the console and would otherwise mis-route
		// through the villagers/inspect branches below for a screen that
		// isn't actually visible.
		if isWidescreen(m.width) {
			return helpModeGlobal
		}
		return helpModeSolo
	case m.lookActive:
		// The look-cursor mode (spec 074 research R9): checked right after
		// the console branch for the exact reason its comment above gives —
		// the mode borrows the dock body, so the villagers/inspect branches
		// below would mis-route through a dockTab/active reading that
		// persists unchanged underneath the borrow.
		return helpModeLook
	case m.inspecting():
		return helpModeInspect
	case m.villagersVisible() && m.villDetail:
		return helpModeVillagersDetail
	case m.villagersVisible():
		return helpModeVillagersRoster
	case isWidescreen(m.width):
		if m.solo {
			return helpModeSolo
		}
		return helpModeGlobal
	default:
		return helpModeSolo // narrow fallback shares the solo/narrow page (data-model.md)
	}
}

// openHelp is the '?' open-trigger (R1, FR-001): freezes the mode help was
// opened from (helpMode) so the page it shows can't drift while the
// overlay owns the keyboard (spec.md "Edge Cases" — esc-release ordering
// and the frozen context), and always starts on that mode's basic-tier
// keys page — UNLESS a conditional header badge is active (spec 074 FR-011,
// research R8), in which case the overlay opens pre-focused on the
// screen-walkthrough section, scrolled to that badge's headerAnatomy row
// (help.go): the badge deep-link. No active badge keeps the open
// byte-identical to before this feature (existing help_test.go pins).
func (m Model) openHelp() (tea.Model, tea.Cmd) {
	mode := m.currentHelpMode()
	m.helpOpen = true
	m.helpMode = mode
	m.helpPageMode = mode
	m.helpTier = false
	m.helpSection = helpSectionKeys
	m.helpScroll = 0
	if idx, ok := m.firstActiveBadgeRow(); ok {
		m.helpSection = helpSectionWalkthrough
		m.helpScroll = idx
	}
	return m, nil
}

// firstActiveBadgeRow resolves the first (header order: degraded, llm,
// suppressed) active conditional badge to its helpWalkthroughLines content
// index (help.go): 1 (the "Header anatomy" title line) plus the badge's
// position in headerAnatomy — the SAME shared table headerView's badges and
// this walkthrough row derive from, so the pre-focus target can never drift
// from what actually renders (research R8). ok is false with no status yet
// or no active badge — openHelp's byte-identical fallback.
func (m Model) firstActiveBadgeRow() (int, bool) {
	if m.status == nil {
		return 0, false
	}
	const headerAnatomyTitleLines = 1
	if m.status.Clock.Degraded {
		return headerAnatomyTitleLines + 6, true // headerAnatomy[6] = "[degraded]"
	}
	if _, _, ok := firstLLMCondition(m.status.LLM); ok {
		return headerAnatomyTitleLines + 7, true // headerAnatomy[7] = "[llm: provider kind]"
	}
	if len(suppressedHorizonClasses(m.status.Horizon)) > 0 {
		return headerAnatomyTitleLines + 8, true // headerAnatomy[8] = "[suppressed: classes]"
	}
	return 0, false
}

// closeHelp dismisses the overlay (esc or '?', toggle — FR-007) and resets
// every help field; nothing else is touched because nothing else ever
// changed while it was open (FR-006 is satisfied by construction).
func (m Model) closeHelp() (tea.Model, tea.Cmd) {
	m.helpOpen = false
	m.helpMode = 0
	m.helpPageMode = 0
	m.helpTier = false
	m.helpSection = 0
	m.helpScroll = 0
	return m, nil
}

// handleHelpKey is the overlay's own keyboard, owned exclusively while
// m.helpOpen (contracts/help-content.md "Layering" #2): esc/'?' dismiss
// (toggle, #3); tab/shift+tab cycle sections; within the keys section, 't'
// advances the tier and 'n'/'p' page through every mode's key table (the
// FR-001 seam that makes the minibuffer's page reachable); J/K scroll the
// current page's pager (chronicleDetailPane idiom, R4, help.go
// paginateHelpContent). Every other key is swallowed here — handled,
// never falling through to the mode beneath (FR-012).
func (m Model) handleHelpKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "?":
		return m.closeHelp()
	case "tab":
		m.helpSection = (m.helpSection + 1) % helpSectionCount
		m.helpScroll = 0
	case "shift+tab":
		m.helpSection = (m.helpSection - 1 + helpSectionCount) % helpSectionCount
		m.helpScroll = 0
	case "t":
		if m.helpSection == helpSectionKeys {
			m.helpTier = !m.helpTier
			m.helpScroll = 0
		}
	case "n":
		if m.helpSection == helpSectionKeys {
			m.helpPageMode = nextHelpMode(m.helpPageMode)
			m.helpScroll = 0
		}
	case "p":
		if m.helpSection == helpSectionKeys {
			m.helpPageMode = prevHelpMode(m.helpPageMode)
			m.helpScroll = 0
		}
	case "J":
		m.helpScroll++ // clamped to content length at render time (R4)
	case "K":
		if m.helpScroll > 0 {
			m.helpScroll--
		}
	}
	return m, nil
}

// handleTakeoverKey is the takeover family's own keyboard (spec 056,
// contracts/takeovers.md §1), owned exclusively while a takeover is open:
// esc dismisses one layer everywhere (FR-004) — on the postmortem it also
// latches postmortemDismissed (suppresses the next connect's auto-reopen
// this session; `p` overrides); a deferred ceremony is NOT re-opened by
// dismissing the postmortem — it stays reachable only via the replay
// surfaces (overlays/ceremony.md's replayability AC). `q` quits/detaches
// exactly like every other quit path (m.quit()) — the D13 "world keeps
// running" framing and the postmortem's plain quit are both the SAME
// unconditional message already, differentiated only by whether the world
// has actually ended (View()'s quitting branch, runEnded()); no special
// case is needed here for which kind is open. Every other key (including
// '?' — contracts/takeovers.md §1 "? | takeover open") is swallowed: the
// takeover keeps the body slot.
func (m Model) handleTakeoverKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		if m.takeover == takeoverPostmortem {
			m.postmortemDismissed = true
		}
		m.takeover = takeoverNone
		return m, nil
	case "q":
		return m.quit()
	}
	return m, nil
}

// handleGlobalKey is patterns/keymap.md "Mode: global".
func (m Model) handleGlobalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m.quit()
	case "1":
		m.solo = false
		m.active = paneMap
		return m, nil
	case "2":
		return m.selectTab(paneChronicle)
	case "3":
		return m.selectTab(paneGuardian)
	case "4":
		return m.selectTab(paneVillagers)
	case "5":
		return m.selectTab(paneSystems)
	case "6":
		// Spec 054 (FR-008): the exercise tab exists only on scenario worlds;
		// on an ambient world 6 falls through inert (contract §4).
		if m.exerciseID() != "" {
			return m.selectTab(paneExercise)
		}
	case "G":
		// The guardian console (spec 053 FR-001): reached here only when
		// nothing higher in handleKey's chain claimed "G" first — inspect
		// mode (chronJumpLast) and villagers mode (villJumpLast) both bind
		// "G" already and take priority while their tab is active, exactly
		// matching FR-001's own wording ("from home, solo, and narrow" —
		// inspect/villagers sub-modes aren't named, and already own the key).
		return m.openConsole()
	case "tab":
		return m.selectTab(m.nextDockTab(m.dockTab))
	case "shift+tab":
		return m.selectTab(m.prevDockTab(m.dockTab))
	case "m":
		return m.focusMinibuffer()
	case "v":
		// Look-cursor mode entry (spec 074, research R1): a strict
		// documented no-op when the map isn't the thing on screen or no
		// world is attached (the `x`-key no-op precedent) — look.go's
		// lookEntryAllowed is the single gate both this key and the mouse
		// map-tile-click entry path read.
		if m.lookEntryAllowed() {
			m.enterLook()
		}
		return m, nil
	case "x":
		// Dismiss the active lesson (spec 055, FR-010, patterns/keymap.md):
		// a strict no-op when nothing is active — Dismiss itself already
		// leaves the model untouched in that case, so there is nothing
		// further to branch on here (the documented-no-op-fallthrough
		// doctrine, not a silent gap).
		m.lessons.Dismiss(time.Now())
		return m, nil
	case "enter":
		// Narrow-fallback-only affordance (focus-contract.md scope): the
		// guardian pane's dormant input line focuses on 'm' *or* Enter,
		// mirroring minibuffer.md's placeholder hint since there is no
		// separate always-visible minibuffer bar to press 'm' toward.
		if !isWidescreen(m.width) && m.active == paneGuardian {
			return m.focusMinibuffer()
		}
	case "esc":
		// Rule 3, "esc always releases" — here nothing is focused, so the
		// next thing esc releases is a solo zoom (solo-views.md state
		// machine: "solo(k) --esc--> home, tab=k").
		if m.solo {
			m.solo = false
		}
		return m, nil
	case "a", "t", "r":
		if m.chronicleVisible() {
			switch msg.String() {
			case "a": // all → each villager → all
				m.chronAgent++
				if m.replica == nil || m.chronAgent >= len(m.replica.Agents) {
					m.chronAgent = -1
				}
			case "t": // all → each thread seen in the ring → all
				m.chronThread = nextThread(m.replica, m.chronThread)
			case "r":
				m.chronRaw = !m.chronRaw
			}
		}
	case "up", "down", "left", "right", "c":
		if m.mapControllable() {
			switch msg.String() {
			case "up":
				m.panY -= 4
			case "down":
				m.panY += 4
			case "left":
				m.panX -= 4
			case "right":
				m.panX += 4
			case "c":
				m.panX, m.panY = 0, 0
			}
		}
	case " ":
		// Clock keys are inert once the run has ended (spec 044 FR-002; the
		// footer hint says so) — gated client-side because the daemon's
		// refusal error would otherwise read as a disconnect (timeControl).
		if m.connected && m.status != nil && !m.runEnded() {
			cmd := "pause"
			if m.status.Clock.Paused {
				cmd = "resume"
			}
			return m, timeControl(m.client, cmd, nil)
		}
	case "[", "]":
		if m.connected && m.status != nil && !m.runEnded() {
			cur := clock.Speed(m.status.Clock.Speed)
			idx := 1 // default 4x position
			for i, s := range speedSteps {
				if s == cur {
					idx = i
				}
			}
			if msg.String() == "[" && idx > 0 {
				idx--
			} else if msg.String() == "]" && idx < len(speedSteps)-1 {
				idx++
			}
			if speedSteps[idx] != cur {
				return m, timeControl(m.client, "set_speed", ipc.SetSpeedArgs{Speed: string(speedSteps[idx])})
			}
		}
	}
	return m, nil
}

// nextDockTab/prevDockTab cycle the dock tabs (tab/shift+tab aliases,
// keymap.md "Migration notes" — not load-bearing). paneSystems (spec 053)
// extends the cycle after villagers (chronicle -> guardian -> villagers ->
// systems), same as its "5" position in the tab row; Model methods since
// spec 054 — the cycle continues into the exercise tab exactly when the
// world carries one, same as its "6" position.
func (m Model) nextDockTab(cur pane) pane {
	switch cur {
	case paneChronicle:
		return paneGuardian
	case paneGuardian:
		return paneVillagers
	case paneVillagers:
		return paneSystems
	case paneSystems:
		if m.exerciseID() != "" {
			return paneExercise
		}
		return paneChronicle
	default: // paneExercise
		return paneChronicle
	}
}

func (m Model) prevDockTab(cur pane) pane {
	switch cur {
	case paneChronicle:
		if m.exerciseID() != "" {
			return paneExercise
		}
		return paneSystems
	case paneGuardian:
		return paneChronicle
	case paneVillagers:
		return paneGuardian
	case paneExercise:
		return paneSystems
	default: // paneSystems
		return paneVillagers
	}
}

// selectTab implements the solo-views.md state machine for k ∈
// {chronicle, guardian, villagers}: same key on the already-selected tab zooms
// solo; same key again returns home. A different key while solo switches
// which tab is solo'd rather than dropping back to home — the state
// machine only specifies the same-key case, so this keeps solo a pure
// "the dock at full width" (dock.md: "same component, two widths") rather
// than adding an implicit extra "back home" side effect to tab-switching.
// active mirrors the selection so a resize down to the narrow fallback
// shows the same content that was last looked at.
func (m Model) selectTab(k pane) (tea.Model, tea.Cmd) {
	if isWidescreen(m.width) {
		if m.solo {
			if m.dockTab == k {
				m.solo = false
			} else {
				m.dockTab = k
			}
		} else if m.dockTab == k {
			m.solo = true
		} else {
			m.dockTab = k
		}
	} else {
		m.dockTab = k
	}
	m.active = k
	var cmd tea.Cmd
	if k == paneGuardian {
		m.guardianUnseen = false
		m.mbFlash = ""
		if m.connected && m.client != nil {
			cmd = fetchConsoleStatus(m.client)
		}
	}
	return m, cmd
}

// focusMinibuffer is the 'm' key (focus-contract.md rule 1: "text capture
// begins solely on an explicit focus action"). In the narrow fallback the
// input line only exists inside the guardian pane, so focusing also
// switches to it — the focused chrome must be visible the instant it is
// focused (rule 2).
func (m Model) focusMinibuffer() (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if !isWidescreen(m.width) {
		mdl, c := m.selectTab(paneGuardian)
		m = mdl.(Model)
		cmd = c
	}
	m.mbFocused = true
	m.mbErr = ""
	m.mbHistPos = len(m.mbHistory)
	m.mbDraft = ""
	m.guardianUnseen = false
	m.mbFlash = ""
	return m, cmd
}

// --- guardian console (spec 053, pages/guardian-console.md) ---

// openConsole is the 'G' key from the global mode (FR-001): flips the page
// flag and, exactly like selecting the guardian dock tab, peeks the
// model-free console status (charter/skills provenance for the read
// surface, FR-004) and clears the unseen badge/flash — the console shows
// the same transcript the tab does (guardianVisible), so opening it counts
// as "having seen" it. dockTab/solo/active are deliberately untouched: they
// already ARE the return target data-model.md calls consoleReturn (research
// R1) — nothing to snapshot, nothing to restore beyond the flag.
func (m Model) openConsole() (tea.Model, tea.Cmd) {
	m.console = true
	m.consoleScroll = 0
	m.guardianUnseen = false
	m.mbFlash = ""
	var cmd tea.Cmd
	if m.connected && m.client != nil {
		cmd = fetchConsoleStatus(m.client)
	}
	return m, cmd
}

// closeConsole is 'G' (toggle), '1', or unfocused 'esc' while the console is
// open (contract §1; data-model.md "Transitions"): scroll and the one-shot
// notice both reset — "reading posture, not archive" (spec.md Edge Cases).
func (m Model) closeConsole() (tea.Model, tea.Cmd) {
	m.console = false
	m.consoleScroll = 0
	m.consoleNotice = ""
	return m, nil
}

// consoleFocusMinibuffer is the console's own 'm' — unlike the global
// focusMinibuffer, it never calls selectTab: the console is already a
// full-screen page, not the guardian pane, so switching m.active here would
// corrupt the "return to whatever was open before" restore (research R1) —
// the console honors the focus contract by construction (patterns/focus-
// contract.md "This feature's new permanent chrome"), reusing the same
// minibuffer state fields the compact tab's focusMinibuffer sets.
func (m Model) consoleFocusMinibuffer() (tea.Model, tea.Cmd) {
	m.mbFocused = true
	m.mbErr = ""
	m.mbHistPos = len(m.mbHistory)
	m.mbDraft = ""
	m.guardianUnseen = false
	m.mbFlash = ""
	return m, nil
}

// handleConsoleKey is the console page's own key layer (contract §1),
// owned exclusively while m.console and the minibuffer is unfocused —
// mirroring handleInspectKey/handleVillagersKey's "layer on top, fall
// through when unclaimed" shape (handled=false lets handleKey's console
// branch continue straight to handleGlobalKey, contract §1's "space pause ·
// q quit" footer hints). J/K scrollback direction (research R4, "tail-
// anchored... scrollback"): K moves toward older turns (increments the
// tail offset, revealing history above), J moves back toward the present
// (decrements toward 0, the tail) — the chat-scrollback convention, the
// mirror image of chronicleDetailPane's head-anchored J-increments/K-
// decrements (that pane windows from the top; this one from the bottom).
func (m Model) handleConsoleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	switch msg.String() {
	case "G", "1", "esc":
		mdl, cmd := m.closeConsole()
		return mdl, cmd, true
	case "m":
		mdl, cmd := m.consoleFocusMinibuffer()
		return mdl, cmd, true
	case "e":
		mdl, cmd := m.startEditorHandoff()
		return mdl, cmd, true
	case "J":
		if m.consoleScroll > 0 {
			m.consoleScroll--
		}
		return m, nil, true
	case "K":
		m.consoleScroll++ // clamped to content length at render time (consoleScrollWindow)
		return m, nil, true
	}
	return m, nil, false
}

// --- $EDITOR handoff (spec 053 US3, research R2) ---

// editorCommand builds the *exec.Cmd for the handoff: $EDITOR <world>/
// charter.md (contract §4). Extracted from startEditorHandoff so tests can
// drive the real subprocess round trip (a scripted fake editor) directly —
// tea.ExecProcess's own suspend/exec/restore plumbing only activates inside
// a running tea.Program (the Cmd it returns is an unexported execMsg that
// only Program.Update intercepts), so exercising that half is bubbletea's
// own tested responsibility, not re-tested here.
func editorCommand(editor, path string) *exec.Cmd {
	return exec.Command(editor, path)
}

// hashFile content-hashes path for the pre/post $EDITOR comparison (research
// R2: "content hash, not mtime alone, avoids false confirmations from
// editors that touch files on open"). A missing/unreadable file hashes to
// "" both before and after unless the editor actually created content —
// never a false "changed".
func hashFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// editorRoundTripMsg builds the post-$EDITOR message (R2): a nonzero exit or
// exec failure reports as an honest error notice regardless of the file's
// content (edge case: "$EDITOR exits nonzero: treat as no-change... plus an
// honest one-line notice" — never partially applied, the file is the file).
func editorRoundTripMsg(beforeHash, path string, err error) tea.Msg {
	if err != nil {
		return editorResultMsg{err: err}
	}
	return editorResultMsg{changed: beforeHash != hashFile(path)}
}

// startEditorHandoff is the console's 'e' key (contract §4): $EDITOR unset
// is refused before ever building a command — an honest notice one keypress
// sooner than a shell "command not found" would be (FR-005's "no in-TUI
// editor" ruling still holds: this never reads or writes charter.md's text
// itself, only stats it for the hash). tea.ExecProcess is Bubble Tea's own
// blessed suspend/exec/restore mechanism (research R2) — the world keeps
// running; only the client's terminal is handed off.
func (m Model) startEditorHandoff() (tea.Model, tea.Cmd) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		m.consoleNotice = "no $EDITOR set — set the EDITOR environment variable to edit charter.md"
		return m, nil
	}
	path := m.w.CharterPath()
	before := hashFile(path)
	cmd := editorCommand(editor, path)
	return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
		return editorRoundTripMsg(before, path, err)
	})
}

// handleMinibufferKey is patterns/keymap.md "Mode: minibuffer focused" —
// every key has a visible effect (focus-contract.md rule 4); there is no
// key whose press produces no observable change.
func (m Model) handleMinibufferKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		// Rule 3: "esc always releases. One keypress returns full
		// keyboard control, instantly."
		m.mbFocused = false
		m.mbHistPos = len(m.mbHistory)
	case tea.KeyEnter:
		text := strings.TrimSpace(m.mbInput)
		if text == "" {
			// "⏎ on an empty buffer releases focus (no-op send)."
			m.mbFocused = false
			return m, nil
		}
		if !m.connected || m.client == nil {
			m.mbFocused = false
			m.mbErr = "not connected"
			return m, nil
		}
		m.mbHistory = append(m.mbHistory, text)
		m.transcript = append(m.transcript, "you: "+text)
		m.mbInput = ""
		m.mbHistPos = len(m.mbHistory)
		m.mbBusy = true
		m.mbErr = ""
		// "Focus is released automatically on send; esc (or any
		// navigation) just proceeds — busy never blocks the UI."
		m.mbFocused = false
		return m, sendConsole(m.client, text)
	case tea.KeyBackspace:
		if r := []rune(m.mbInput); len(r) > 0 {
			m.mbInput = string(r[:len(r)-1])
		}
	case tea.KeyUp:
		m.historyUp()
	case tea.KeyDown:
		m.historyDown()
	case tea.KeySpace:
		m.mbInput += " "
	case tea.KeyRunes:
		m.mbInput += string(msg.Runes)
	}
	return m, nil
}

func (m *Model) historyUp() {
	if len(m.mbHistory) == 0 {
		return
	}
	if m.mbHistPos == len(m.mbHistory) {
		m.mbDraft = m.mbInput
	}
	if m.mbHistPos > 0 {
		m.mbHistPos--
	}
	m.mbInput = m.mbHistory[m.mbHistPos]
}

func (m *Model) historyDown() {
	if m.mbHistPos < len(m.mbHistory) {
		m.mbHistPos++
	}
	if m.mbHistPos >= len(m.mbHistory) {
		m.mbHistPos = len(m.mbHistory)
		m.mbInput = m.mbDraft
		return
	}
	m.mbInput = m.mbHistory[m.mbHistPos]
}

// handleInspectKey is patterns/keymap.md "Mode: inspect" — layered on top
// of the global mode, never replacing it (handled is false for any key it
// does not own, so handleKey falls through to handleGlobalKey). J/K scroll
// the always-on detail pane (contract §5/§6, R6); ⏎ jumps the map camera to
// the selected event's subject (spec 049, contract §1) — the seam this same
// comment used to describe as reserved is now filled; jumpToSource is
// shared with the chronicle-line click (handleMouse) so both input paths
// apply identical rules (FR-004/FR-009, input-parity doctrine).
func (m Model) handleInspectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	switch msg.String() {
	case "j":
		m.chronMoveSelection(1)
		return m, nil, true
	case "k":
		m.chronMoveSelection(-1)
		return m, nil, true
	case "g":
		m.chronJumpFirst()
		return m, nil, true
	case "G":
		m.chronJumpLast()
		return m, nil, true
	case "J":
		m.chronDetailScroll++ // clamped to content length at render time
		return m, nil, true
	case "K":
		if m.chronDetailScroll > 0 {
			m.chronDetailScroll--
		}
		return m, nil, true
	case "enter":
		mdl, cmd := m.jumpToSource()
		return mdl, cmd, true
	}
	return m, nil, false
}

// --- jump-to-source (spec 049) ---

// centerCameraOn sets the camera pan so the wanderer centroid + pan lands
// exactly on (x,y) — research R1: a jump IS a pan, computed instead of
// accumulated, so `c` (recenter, resets pan to 0,0) and the panned-title/
// auto-follow-suspend behavior (renderMapGrid, mapPanelView) fall out of
// the existing pan machinery for free; no new camera state, no new clamping
// (render-time clampInt already bounds the visible window regardless of
// panX/panY's magnitude, same as manual arrow-key panning).
func (m *Model) centerCameraOn(x, y int) {
	cx, cy := m.wandererCentroid()
	m.panX = x - cx
	m.panY = y - cy
}

// jumpToSource applies contract §1/§4 to the currently selected chronicle
// event: resolveSubject decides jump-or-hint (FR-002/FR-003). A successful
// jump centers the camera and, in the narrow fallback, switches the visible
// pane to the map so the effect is actually seen (FR-007) rather than
// landing invisibly behind the chronicle pane. An unlocatable event changes
// nothing — the detail pane's actions bar (detailActions) is already the
// visible explanation (contract §1), so there is nothing further to do
// here. Shared by the ⏎ key (handleInspectKey) and the chronicle-line click
// (handleMouse) — one behavior, two input paths (input-parity doctrine).
func (m Model) jumpToSource() (tea.Model, tea.Cmd) {
	sel := m.chronSelectionBase()
	if sel < 0 {
		return m, nil
	}
	_, x, y, ok := m.resolveSubject(m.events[sel])
	if !ok {
		return m, nil
	}
	m.centerCameraOn(x, y)
	if !isWidescreen(m.width) {
		m.active = paneMap // FR-007: land where the jump's effect is visible
	}
	return m, nil
}

// chronicleHitOrigin computes where the chronicle inspect list's row 0
// lands on screen, per the layout currently active — pure arithmetic over
// the same fixed chrome-row/column constants (header, box border, tab row,
// divider) already literal in widescreenView/narrowView/dockPanelView/
// soloPanelView, never the per-event wrap/window logic those functions also
// contain (research R3: that part only the renderer itself may compute).
// Only ever meaningful while chronicleInspectBody is actually rendering
// (its caller is gated on m.inspecting()), so isWidescreen(m.width) and
// m.solo alone are enough to tell which of the three layouts is live.
func (m Model) chronicleHitOrigin() (originX, originY, width int) {
	if !isWidescreen(m.width) {
		// narrowView: headerView (1) + tabsView (1) + blank line (1).
		return 0, 3, m.width
	}
	if m.solo {
		// soloPanelView: header (1) + box top border (1) + title row (1).
		return 0, 3, m.width
	}
	// dockPanelView beside mapPanelView: header (1) + box top border (1) +
	// tab row (1) + divider (1); dock panel's columns start after the map
	// panel + gutter (computeColumns, layout.go).
	cols := computeColumns(m.width)
	return cols.MapCols + cols.Gutter, 4, cols.DockCols
}

// recordChronHit stashes this frame's list-row → event-index geometry
// (chronHitRegion) for the next Update's mouse routing — the renderer that
// already knows the per-row event windowing (chronicleInspectBody's start/
// end) is the only honest source (research R3); the click handler never
// re-derives it.
func (m Model) recordChronHit(start, end int) {
	if m.chronHit == nil {
		return
	}
	originX, originY, width := m.chronicleHitOrigin()
	rowEvent := make([]int, end-start)
	for i := range rowEvent {
		rowEvent[i] = start + i
	}
	*m.chronHit = chronHitRegion{valid: true, originX: originX, originY: originY, width: width, rowEvent: rowEvent}
}

// invalidateChronHit marks the last-recorded hit region stale — called at
// the top of every View() so a frame that never renders the chronicle
// inspect list (other tab, running mode, help overlay) can't leave a stray
// click acting on geometry from a previous frame.
func (m Model) invalidateChronHit() {
	if m.chronHit != nil {
		m.chronHit.valid = false
	}
}

// invalidateReverseJumpHits marks the strip/roster hit regions stale —
// called at the top of every View() (the chronHit frame-top pattern) so a
// frame that never renders the strip (narrow/folded) or the roster (other
// tab, detail view) can't leave a stray click acting on old geometry.
func (m Model) invalidateReverseJumpHits() {
	if m.stripHit != nil {
		m.stripHit.valid = false
	}
	if m.rosterHit != nil {
		m.rosterHit.valid = false
	}
}

// handleMouse (spec 074 research R6) now lives in look.go: it routes the
// TILE-pane region and the map region ahead of this exact chronicle-line
// click path (US2, input-parity doctrine decision 8), which is otherwise
// unchanged — a left-button *release* landing inside the chronicle's
// last-rendered list rows, while paused, selects that row and applies the
// same jump rules ⏎ does (jumpToSource).

// handleVillagersKey is contracts/state-and-keys.md's key grammar table,
// layered on top of the global mode exactly like handleInspectKey — j/k/g/G
// select in the roster, ⏎ opens the detail view, esc closes it. Unlike
// inspect mode this does not require the clock to be paused: it is gated
// purely on villagersVisible() (dock.md "Each tab keeps its own state").
// esc on the roster returns handled=false so it falls through to the
// global esc (focus-contract.md rule 3: minibuffer → detail → solo → home).
func (m Model) handleVillagersKey(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	m.clampVillSelected()
	switch msg.String() {
	case "esc":
		// spec 020 (TASK-63): decisions → detail → roster, one level per
		// press, ahead of the existing detail → roster chain below
		// (focus-contract.md rule 3).
		switch {
		case m.villDecisions:
			m.villDecisions = false
			m.villDecisionsScroll = 0
			return m, nil, true
		case m.villDetail:
			m.villDetail = false
			return m, nil, true
		}
		return m, nil, false
	case "d":
		// spec 020 (TASK-63, contract R7): toggles the decisions sub-view
		// while the detail view is open; a no-op (falls through) from the
		// roster, same as every other detail-only key.
		if m.villDetail {
			m.villDecisions = !m.villDecisions
			m.villDecisionsScroll = 0
			return m, nil, true
		}
		return m, nil, false
	case "j":
		switch {
		case m.villDecisions:
			m.villDecisionsScroll++ // clamped to content length at render time
		case !m.villDetail:
			m.villMoveSelection(1)
		}
		return m, nil, true
	case "k":
		switch {
		case m.villDecisions:
			if m.villDecisionsScroll > 0 {
				m.villDecisionsScroll--
			}
		case !m.villDetail:
			m.villMoveSelection(-1)
		}
		return m, nil, true
	case "g":
		if !m.villDetail {
			m.villJumpFirst()
		}
		return m, nil, true
	case "G":
		if !m.villDetail {
			m.villJumpLast()
		}
		return m, nil, true
	case "enter":
		if !m.villDetail && m.villCount() > 0 {
			m.villDetail = true
		}
		return m, nil, true
	case "J":
		// Reverse jump (spec 086 US5, FR-010): center the map camera on the
		// selected villager — the keyboard path of the strip-glyph/roster-row
		// click (keyboard primary, input-parity doctrine). Works from the
		// roster and the detail view alike; dead villagers jump to their
		// grave (agents keep X,Y after death); an empty replica is a no-op.
		if m.villCount() > 0 {
			a := m.replica.Agents[clampInt(m.villSelected, 0, len(m.replica.Agents)-1)]
			m.centerCameraOn(a.X, a.Y)
			if !isWidescreen(m.width) {
				m.active = paneMap // land where the jump's effect is visible
			}
		}
		return m, nil, true
	}
	return m, nil, false
}

// villCount is len(replica.Agents), 0 with a nil/empty replica — the bound
// every villSelected read clamps against (R5).
func (m Model) villCount() int {
	if m.replica == nil {
		return 0
	}
	return len(m.replica.Agents)
}

// clampVillSelected bounds villSelected to [0, villCount()) — called on
// reconnect (connectedMsg swaps the replica wholesale) and defensively on
// every villagers keypress; renderers clamp again at read time the same way
// chronSelectionBase does for the chronicle.
func (m *Model) clampVillSelected() {
	n := m.villCount()
	if n == 0 {
		m.villSelected = 0
		return
	}
	m.villSelected = clampInt(m.villSelected, 0, n-1)
}

func (m *Model) villMoveSelection(delta int) {
	n := m.villCount()
	if n == 0 {
		return
	}
	m.villSelected = clampInt(m.villSelected+delta, 0, n-1)
}

func (m *Model) villJumpFirst() {
	if m.villCount() == 0 {
		return
	}
	m.villSelected = 0
}

func (m *Model) villJumpLast() {
	n := m.villCount()
	if n == 0 {
		return
	}
	m.villSelected = n - 1
}

// chronSelectionBase resolves the "current" selection: if nothing is
// selected yet (or the ring rotated past the old index), it starts from
// the tail — the most recently paused-over event.
func (m Model) chronSelectionBase() int {
	n := len(m.events)
	if n == 0 {
		return -1
	}
	if m.chronSelected < 0 || m.chronSelected >= n {
		return n - 1
	}
	return m.chronSelected
}

func (m *Model) chronMoveSelection(delta int) {
	n := len(m.events)
	if n == 0 {
		return
	}
	sel := m.chronSelectionBase() + delta
	if sel < 0 {
		sel = 0
	}
	if sel >= n {
		sel = n - 1
	}
	m.chronSelected = sel
	m.chronDetailScroll = 0 // data-model.md: reset on selection move
}

func (m *Model) chronJumpFirst() {
	if len(m.events) == 0 {
		return
	}
	m.chronSelected = 0
	m.chronDetailScroll = 0
}

func (m *Model) chronJumpLast() {
	if len(m.events) == 0 {
		return
	}
	m.chronSelected = len(m.events) - 1
	m.chronDetailScroll = 0
}

// detailAction is one jump-off action attachable to the detail pane's
// bottom-right actions bar (contract §3, FR-009). Spec 018/spec 047 reserved
// this hook for exactly the feature that fills it (spec 049): jump-to-source
// is the actions bar's first (and, today, only) action.
type detailAction struct {
	Label string
}

// detailActions returns the jump-off actions available for one event —
// exactly one, always (data-model.md "Never nil after this feature" —
// SC-002's totality is a property of this function, and TestJumpCatalogSweep
// asserts it holds for every cataloged event type): the jump affordance
// (contract §3) when resolveSubject locates a subject, the honest
// no-location text otherwise. Takes the Model (for resolveSubject's live
// replica) and the event, so a future action can extend this without
// changing callers' shape.
func (m Model) detailActions(e store.Event) []detailAction {
	name, x, y, ok := m.resolveSubject(e)
	if !ok {
		return []detailAction{{Label: "no location for this event"}}
	}
	return []detailAction{{Label: fmt.Sprintf("⏎ jump to %s (%d,%d)", name, x, y)}}
}

func (m Model) handlePush(p ipc.Push) (tea.Model, tea.Cmd) {
	if !m.connected || m.client == nil {
		return m, nil
	}
	switch p.Push {
	case "event":
		m.applyEvent(*p.Event)
		return m, listen(m.client)
	case "dropped":
		// Overflow: the replica may have missed events — resync from scratch.
		m.client.Close()
		m.client = nil
		m.connected = false
		m.lastErr = "stream overflow; resyncing"
		return m, connect(m.w)
	}
	return m, listen(m.client)
}

// applyEvent folds one pushed event into the replica and the chronicle ring.
// Events at or before the state snapshot's seq are already reflected and skipped.
func (m *Model) applyEvent(e store.Event) {
	if e.Seq <= m.lastSeq {
		return
	}
	if m.replica != nil {
		m.replica.Apply(e) // same reducer as the daemon; errors are cosmetic here
		if e.Tick > m.replica.Tick {
			m.replica.Tick = e.Tick
		}
	}
	// Decision-trace projection (spec 020, TASK-63, research D1): ingested
	// here, before the ring append below, so stimulus resolution
	// (resolveStimulus) sees the ring exactly as it stood before this
	// event — the trigger event, if any, is always an earlier seq and so
	// already in it.
	m.traces.ingest(e, m.agentNames(), m.events, m.sk())
	// Lessons projection (spec 055, TASK-117, research.md R3): folded here,
	// alongside the decision-trace projection above — same one-event-at-a-
	// time seam, same "every event exactly once" guarantee. A non-nil
	// return means a lesson just surfaced (became the active row entry);
	// persist it immediately (mark-seen-on-surface, not on queue — FR-005).
	if surfaced := m.lessons.ingest(e, time.Now()); surfaced != nil {
		worlds.MarkLessonSeen(surfaced.ID, m.w.Manifest.Name)
	}
	if line, ok := guardianVerdictRow(e); ok {
		m.transcript = append(m.transcript, line)
		if len(m.transcript) > 200 {
			m.transcript = m.transcript[len(m.transcript)-200:]
		}
	}
	// Takeover-family transitions (spec 056, contracts/takeovers.md §1):
	// folded here, after m.replica.Apply above already latched
	// State.Ended/RunEnd or appended StagesUnlocked, so the render-time
	// derivation (postmortemView/ceremonyView, views.go) always finds the
	// replica already reflecting whichever event just landed. No content is
	// captured into a second Model field here — the ceremony's identity is
	// always replica.StagesUnlocked's last entry while takeover ==
	// takeoverCeremony (this event is exactly what appended it).
	switch e.Type {
	case "run.ended":
		// Postmortem always wins — replaces an open ceremony unconditionally
		// (never deferred, the ceremony side of this rule). A fresh
		// run.ended always (re)opens regardless of an earlier dismissal
		// this session: postmortemDismissed guards against re-annoying a
		// player across a mere reconnect, never against the run genuinely
		// ending while they were looking at something else.
		m.takeover = takeoverPostmortem
		m.postmortemDismissed = false
	case "curriculum.stage_unlocked":
		if m.takeover == takeoverPostmortem {
			// Deferred, never interrupting — the content stays reachable via
			// replay (overlays/ceremony.md's pull surfaces; help.go's
			// ceremony section), derived from replica facts, not stored here.
			m.ceremonyDeferred = true
		} else {
			// Opens immediately, replacing any already-open ceremony
			// (same-kind, non-stacking — the newest milestone wins).
			m.takeover = takeoverCeremony
		}
	}
	// The report card's console seam (spec 063 T012): recompose from the
	// just-reduced state on every stopping-point-relevant event — the fresh
	// stored note, an exercise resolution (the checklist half concludes),
	// or the run's end. A fresh NOTE additionally announces itself with the
	// existing unseen-badge pattern when the guardian surface isn't visible
	// — at most a badge between stopping points, never a takeover (FR-006).
	switch e.Type {
	case "guardian.report_card":
		m.rebuildConsoleCards()
		if !m.guardianVisible() {
			m.guardianUnseen = true
		}
	case "curriculum.exercise_passed", "run.ended":
		m.rebuildConsoleCards()
	}
	m.lastSeq = e.Seq
	m.events = append(m.events, e)
	if len(m.events) > chronicleCap {
		m.events = m.events[len(m.events)-chronicleCap:]
	}
}

// clampGeometry re-bounds the size-independent-in-value-but-not-in-validity
// fields on resize (B5). Almost all geometry in this package is derived
// fresh every View() call from (width, height) directly — there is no
// cache to go stale. The two fields that *do* persist a value across
// frames are covered here:
func (m *Model) clampGeometry() {
	// Pan offset: render-time clamping in renderMapGrid already keeps the
	// *visible* camera window inside the map regardless of panX/panY's
	// magnitude, but an offset accumulated at a wide viewport is still
	// stale once the viewport shrinks — cap it to a map-sized window so
	// it can never represent "off the map" at any size.
	if m.gameMap != nil {
		m.panX = clampInt(m.panX, -m.gameMap.W, m.gameMap.W)
		m.panY = clampInt(m.panY, -m.gameMap.H, m.gameMap.H)
	}
	// Chronicle selection: bounded to the current ring length. Read-time
	// callers (chronSelectionBase) already tolerate an out-of-range
	// value defensively, but the cached field itself should stay honest
	// rather than merely tolerated.
	if m.chronSelected >= len(m.events) {
		m.chronSelected = len(m.events) - 1
	}
	// dockTab/solo are a small fixed enum + bool, never derived from
	// width/height — nothing to clamp; the narrow fallback ignores
	// `solo` entirely and `dockTab` is always one of the three valid
	// tabs regardless of layout.
}

// agentNames resolves the replica's roster for the chronicle grammar's
// name-resolution (patterns/chronicle-grammar.md, "the existing chronNames
// mechanism", generalized to raw event payloads).
func (m Model) agentNames() []string {
	if m.replica == nil {
		return nil
	}
	names := make([]string, len(m.replica.Agents))
	for i, a := range m.replica.Agents {
		names[i] = a.Name
	}
	return names
}

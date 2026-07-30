// Package guardian is the gatekeeper guardian (TASK-12): the player's sole
// influence channel. A daemon-hosted notify consumer with its own replica
// (the mind/scribe pattern), it converses with the player, digests the event
// stream into an accreting soul, flags dramatic moments, and mediates nudges
// — whose only path into the world is the InjectSocial door, carrying only
// Guardian's own rendering. Raw player text has exactly one sink: Guardian's
// prompt.
package guardian

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/evanstern/promptworld/internal/bundle"
	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/skin"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/toolloop"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// Submitter is the orchestrator surface Guardian needs (test seam).
type Submitter interface {
	Submit(ctx context.Context, req llm.Request) (llm.Response, error)
}

// Injector is the loop surface Guardian needs (test seam).
type Injector interface {
	InjectSocial(events []store.Event) error
}

// LoopControl is the clock-control surface the meta tools drive (spec 029 US5,
// R10 / data-model §5): the SAME *sim.Loop.Do the IPC server uses, so a
// guardian-issued pause/start/adjust_speed lands the loop's own clock.paused /
// clock.resumed / clock.speed_set events and is indistinguishable from a console
// one. Injected at New — a second interface over the loop the daemon already
// passes as Injector (the mind.New(loop, loop) precedent); a test seam otherwise.
type LoopControl interface {
	Do(name string, speed clock.Speed) (sim.Status, error)
}

const (
	// turnTimeout bounds one console turn's cloud call (SC-001 wants ≤30s
	// in practice; the cap covers slow routers without wedging the console).
	turnTimeout = 2 * time.Minute
	// soulTailBytes / transcriptTailTurns bound what rides the prompt.
	soulTailBytes       = 4000
	transcriptTailTurns = 6
)

type Guardian struct {
	orch     Submitter
	social   Injector
	loop     LoopControl // clock control for the meta tools (spec 029 US5)
	worldDir string

	// bundles is the boot-frozen set of world-scoped bundle tools (spec 036),
	// set once by the daemon after New via SetBundles and read only by the turn
	// worker (runTurn) — race-free because a turn runs only on an IPC request,
	// long after boot. nil (the default, and every current test) means no bundle
	// tools: the turn assembly behaves exactly as before.
	bundles *bundle.BundleSet

	// stage/charterPreset are the world's curriculum-ladder facts (spec 046),
	// set once by the daemon after New via SetStage from the immutable manifest
	// (FR-002) and read only by the turn worker and Status — the SetBundles
	// discipline, boot-frozen so the ceiling cannot be tampered mid-run. Zero
	// values (the default, and every pre-046 test) mean a pre-ladder, ungated
	// world with the default charter: behavior byte-identical to before.
	stage         string
	charterPreset string

	// skin is the world's boot-frozen display skin (spec 052 FR-003), set once
	// by the daemon after New via SetSkin — the SetBundles/SetStage discipline.
	// nil (the default, and every pre-052 test) is the default Guardian skin
	// (every skin.Skin method is nil-safe). It shapes prompt composition (the
	// voice at the SOUL seam, name substitution in the keeper/confirm prompts)
	// and display text ONLY — never a recorded payload (ruling 1: the event
	// log is skin-free).
	skin *skin.Skin

	replica *sim.State
	m       *worldmap.Map
	seed    uint64 // world seed — boot-immutable; backs the script world view's world.rand (spec 036 US3)

	turnBusy atomic.Bool // one console turn at a time

	// runLoop drives one console turn through the bounded tool-use loop (spec
	// 017, T020). Production wires it to toolloop.Run against the concrete
	// orchestrator; tests that stub the model install a scripted driver. It is
	// touched only by Turn (never the absorb/digest goroutines), so installing it
	// in New before those start is race-free. loopRounds is the hard iteration
	// cap (llm.json loop_max_rounds, normalized), threaded from the daemon.
	runLoop    func(ctx context.Context, j toolloop.Job) (toolloop.Result, error)
	loopRounds int
	// turnTokens is the normalized console-turn response budget (llm.json
	// max_tokens.metatron_turn, spec 025 US2), threaded from the daemon exactly
	// as loopRounds is. Absent resolves to the built-in default (1024) at the
	// config boundary, so this is always an effective value. It governs the
	// console-turn loop ONLY — the guardian digest budget (400) is a different
	// call and stays hardcoded (spec 025 Assumption 2).
	turnTokens int64

	// fileMu guards soul.md / transcript.md appends (turn worker vs absorb
	// goroutine both write).
	fileMu sync.Mutex

	// stateMu guards the small mirrors the turn worker reads while the
	// absorb goroutine owns the replica.
	stateMu sync.Mutex
	charges int
	clockAt int64
	night   bool // mirrored State.Night — the omen night gate reads it turn-side (spec 029 T005)
	alive   map[int]bool
	// agentXY mirrors each villager's tile (absorb-owned, refreshed per batch)
	// so a console turn can resolve a tile-addressed miracle's perception-memory
	// recipient without racing the replica the absorb goroutine owns (spec 016).
	agentXY [][2]int
	// agentNeeds mirrors each villager's health/food/warmth (absorb-owned,
	// refreshed per batch) — the turn worker reads it to build the miracle
	// targeting digest (spec 059 US3) without racing the replica.
	agentNeeds []needMirror
	// structXY / pileXY mirror every structure's / ground pile's tile
	// (absorb-owned, refreshed per batch) so a console turn's bundle handlers
	// can resolve spec-082 class+tile effect targets (structure@X,Y /
	// pile@X,Y presence, compile-time error quality) without racing the
	// replica the absorb goroutine owns — the agentXY discipline. Positions
	// only; the reducer stays authoritative for everything else.
	structXY [][2]int
	pileXY   [][2]int
	// structKind mirrors each structure's kind, parallel to structXY (spec
	// 084 US1: the survey sheet lists structures kind + tile and counts
	// wall-blocked tiles in its passability summary, turn-side, without
	// racing the replica).
	structKind []string
	moments    []string // queued, surfaced oldest-first at the next turn
	story      []string // recent chronicle entries (TASK-11), prompt grounding
	// charterFP / ended mirror State.CharterFingerprint / State.Ended (spec
	// 044 US2): the turn worker's charter observation (observeCharter) reads
	// them under stateMu to decide whether a guardian.charter_observed emission
	// is due, without racing the replica the absorb goroutine owns.
	// skillsFP is the same mirror for State.SkillsFingerprint (spec 077 —
	// observeSkills, the charter observation's twin).
	charterFP string
	skillsFP  string
	ended     bool

	// Standing-order mirror (spec 029 US2/US3, data-model §5): the replica's
	// GuardianOrders is the authority; the turn worker reads this copy under
	// stateMu (like charges/alive), and the absorb path matches live events
	// against it. lastPlaceTick/lastPlaceSeq disambiguate same-tick order ids
	// (research R7) when the async mirror has not yet reflected a just-placed
	// order — placements are serialized under turnBusy, so a plain counter is safe.
	orders        []sim.GuardianOrder
	lastPlaceTick int64
	lastPlaceSeq  int

	// Plan-layer mirrors (spec 084, the orders-mirror discipline): the
	// replica's Designations/Directives are the authority; the turn worker
	// reads these copies under stateMu for id minting and the prompt's plan
	// sections. planMintTick/planMintSeq are the per-prefix ("dsg"/"dir")
	// same-tick id disambiguators — the lastPlaceTick/lastPlaceSeq shape,
	// safe under turnBusy serialization.
	designations []sim.Designation
	directives   []sim.Directive
	planMintTick map[string]int64
	planMintSeq  map[string]int
	// Prophecy mirror + faith mirror (spec 085, the plan-mirror discipline):
	// the replica is the authority; the turn worker reads these under stateMu
	// for id minting ("pro" prefix), the prompt's prophecy section, and the
	// FR-013 faith line. faith is the nil-safe FaithScore projection.
	prophecies []sim.Prophecy
	faith      int
	// Region mirror + myth-briefing cache (spec 101, the plan-mirror
	// discipline): the replica's Regions is the authority; the turn worker
	// reads this copy for id minting ("reg" prefix) and the prompt's region
	// section. myths is the read-only myth-briefing derivation
	// (sim.DominantPlaceMyths), recomputed alongside the mirror on every
	// absorbed batch so the brief_myths tool never touches the replica
	// directly — the same stateMu-guarded-copy discipline as every other
	// turn-worker read.
	regions []sim.Region
	myths   []sim.PlaceMythBriefing

	// Trigger pipeline (spec 029 US3, data-model §5): the absorb goroutine matches
	// live events against active orders and enqueues onto triggerQ; a dedicated
	// worker fires each (land order_triggered → system turn → moment). pendingTrigger
	// (stateMu-guarded) dedups an order already queued but not yet resolved, so an
	// order fires at most once even across batches before the worker consumes it.
	triggerQ       chan triggerJob
	pendingTrigger map[string]bool

	// lastConfirmTick rate-caps fuzzy-order confirms (spec 029 US6, R9): at most
	// one KindGuardianWatch confirm per confirmRateTicks per order. Absorb-owned
	// and NOT event-sourced — a skipped confirm is an economy decision, never world
	// history (data-model §5) — guarded by stateMu alongside pendingTrigger (the
	// absorb goroutine writes it via matchOrders; unit tests drive matchOrders too).
	lastConfirmTick map[string]int64

	// survivalLatch debounces the survival watches (spec 059 US1): per watch id,
	// the set of villager indices currently IN the danger band and already fired
	// upon. A villager staying in-band across the per-game-minute heartbeat does
	// NOT re-fire (the latch is set); it clears when the villager recovers above
	// the watch's re-arm threshold, so a later relapse fires afresh. Absorb-owned
	// and NOT event-sourced (matching is live-only by construction — the hysteresis
	// latch never runs in replay), guarded by stateMu alongside pendingTrigger.
	survivalLatch map[string]map[int]bool

	// digest collection (US4) — absorb-owned.
	digLines []string
	digFrom  int64
	digQ     chan digJob
	digCarry chan []string

	// Report-card producer (spec 063 US4, reportcard.go). cardTrail is the
	// bounded activity ring the grader cites (absorb-appended, stateMu-
	// guarded); cardDoneSeq is the graded high-water mark (the activity
	// gate); cardPauseSpent debounces one card per pause episode. cardQ
	// feeds the dedicated worker (the digQ shape). None of it is
	// event-sourced — a skipped or dropped card is an economy decision,
	// never world history; the STORED note is the recorded fact.
	cardTrail      []cardEvidence
	cardDoneSeq    int64
	cardPauseSpent bool
	cardQ          chan cardJob

	events chan []store.Event
	done   chan struct{}

	// wg joins the four goroutines New starts (run, digestWorker,
	// triggerWorker, reportCardWorker): Close closes done then waits on wg,
	// so a caller that drives a queue right after Close (every fixture's
	// newTestGuardian pattern) can never race a worker still parking in its
	// select — wg.Add(1) sits immediately before each go statement below,
	// keeping the add-site obvious for any future fifth worker.
	wg sync.WaitGroup
}

// New starts the guardian from a state snapshot. The metatron/ dir and an
// empty soul are created on first flight; existing files are kept.
//
// loop is the clock-control seam the meta tools drive (spec 029 US5): the daemon
// passes the same *sim.Loop it passes as the social Injector — one value, two
// interfaces (the mind.New(loop, loop) precedent).
//
// loopRounds is the tool-use loop's iteration cap (llm.json loop_max_rounds,
// already normalized by the daemon). The variadic runLoopOverride is a test
// seam: it installs a scripted loop BEFORE any goroutine starts, for tests that
// stub the model rather than pass a real *llm.Orchestrator. Production omits it —
// New wires runLoop from the concrete orchestrator.
func New(orch Submitter, social Injector, loop LoopControl, m *worldmap.Map, seed uint64, stateJSON []byte, worldDir string, loopRounds int, turnTokens int64, runLoopOverride ...func(context.Context, toolloop.Job) (toolloop.Result, error)) (*Guardian, error) {
	replica := sim.NewState(seed, m)
	if err := json.Unmarshal(stateJSON, replica); err != nil {
		return nil, err
	}
	mt := &Guardian{
		orch:            orch,
		social:          social,
		loop:            loop,
		worldDir:        worldDir,
		replica:         replica,
		m:               m,
		seed:            seed,
		alive:           map[int]bool{},
		digQ:            make(chan digJob, 4),
		digCarry:        make(chan []string, 1),
		cardQ:           make(chan cardJob, 4),
		events:          make(chan []store.Event, 256),
		triggerQ:        make(chan triggerJob, 64),
		pendingTrigger:  map[string]bool{},
		lastConfirmTick: map[string]int64{},
		survivalLatch:   map[string]map[int]bool{},
		planMintTick:    map[string]int64{},
		planMintSeq:     map[string]int{},
		done:            make(chan struct{}),
	}
	mt.loopRounds = loopRounds
	mt.turnTokens = turnTokens
	// The tool-use loop needs the concrete orchestrator (toolloop.Run's contract
	// surface — Submit + ObserveCognition). Production passes it; test seams that
	// stub the model install their own runLoop via runLoopOverride.
	if o, ok := orch.(*llm.Orchestrator); ok {
		mt.runLoop = func(ctx context.Context, j toolloop.Job) (toolloop.Result, error) {
			return toolloop.Run(ctx, o, j)
		}
	}
	if len(runLoopOverride) > 0 && runLoopOverride[0] != nil {
		mt.runLoop = runLoopOverride[0]
	}
	if err := os.MkdirAll(mt.guardianDir(), 0o755); err != nil {
		return nil, err
	}
	if _, err := os.Stat(mt.soulPath()); os.IsNotExist(err) {
		// Genesis header (spec 052 US2 AS-5): guardian-voiced DEFAULT-skin
		// text — SetSkin lands after New, so a custom skin's first soul
		// header still says guardian; existing files are never rewritten.
		header := "# The soul of the Guardian\n\n*The watch begins. The guardian has seen nothing yet.*\n"
		if err := os.WriteFile(mt.soulPath(), []byte(header), 0o644); err != nil {
			return nil, err
		}
	}
	mt.digFrom = replica.Tick / digestWindowTicks
	mt.mirrorState()
	mt.wg.Add(1)
	go mt.run()
	mt.wg.Add(1)
	go mt.digestWorker()
	mt.wg.Add(1)
	go mt.triggerWorker()
	mt.wg.Add(1)
	go mt.reportCardWorker()
	return mt, nil
}

// Observe is the loop-notify path: never blocks.
func (mt *Guardian) Observe(events []store.Event) {
	select {
	case mt.events <- events:
	default:
	}
}

// Close signals every worker New started to stop and waits for all four to
// exit before returning — a caller (production shutdown, or a test fixture
// driving a queue right after Close) is guaranteed no worker is still
// in-flight. A Close during an in-flight job blocks until that job's current
// iteration completes; no timeout (shutdown correctness over speed).
func (mt *Guardian) Close() {
	close(mt.done)
	mt.wg.Wait()
}

// SetBundles installs the boot-frozen bundle set the turn assembly merges into
// its roster/handler surface (spec 036 T013/T014). The daemon calls it once
// after New and before serving; only the turn worker reads it, so no lock is
// needed. A nil set (never installed) leaves the guardian's surface unchanged.
func (mt *Guardian) SetBundles(bs *bundle.BundleSet) { mt.bundles = bs }

// SetStage installs the world's immutable curriculum stage and charter preset
// (spec 046 US2) from the opened manifest. The daemon calls it once after New
// and before serving (the SetBundles discipline); only the turn worker and
// Status read it, so no lock is needed. Zero values (never installed) leave
// the guardian ungated with the default charter — pre-ladder behavior.
func (mt *Guardian) SetStage(stage, charterPreset string) {
	mt.stage, mt.charterPreset = stage, charterPreset
}

// SetSkin installs the world's boot-frozen display skin (spec 052 FR-003).
// The daemon calls it once after New and before serving (the SetBundles
// discipline); only the turn/trigger workers and Status read it, so no lock
// is needed. Never installed (nil) is the default Guardian skin.
func (mt *Guardian) SetSkin(s *skin.Skin) { mt.skin = s }

// sk returns the active skin — nil-safe by skin.Skin's own contract, so the
// zero-value agent (every pre-052 test) resolves the default table.
func (mt *Guardian) sk() *skin.Skin { return mt.skin }

// The on-disk directory name "metatron" (and the soul.md/transcript.md files
// inside it) is FROZEN serialized vocabulary (spec 052 ruling 2): existing
// worlds' files must keep loading forever. Display references use the skin's
// notes label, never the raw path.
func (mt *Guardian) guardianDir() string    { return filepath.Join(mt.worldDir, "metatron") }
func (mt *Guardian) soulPath() string       { return filepath.Join(mt.guardianDir(), "soul.md") }
func (mt *Guardian) transcriptPath() string { return filepath.Join(mt.guardianDir(), "transcript.md") }

func (mt *Guardian) run() {
	defer mt.wg.Done()
	for {
		select {
		case <-mt.done:
			return
		case batch := <-mt.events:
			for _, e := range batch {
				mt.replica.Apply(e)
				if e.Tick > mt.replica.Tick {
					mt.replica.Tick = e.Tick
				}
				mt.observeMoment(e)
				mt.digestNote(e)
				mt.observeCardActivity(e)
			}
			mt.mirrorState()
			// Standing-order matching runs HERE — after the replica applies the
			// batch and the mirror refreshes — so it sees only LIVE events (spec 029
			// US3/R6). Replay reconstructs state via json.Unmarshal + Apply with no
			// guardian running, so a predicate can never match during reconstruction.
			mt.matchOrders(batch)
			// Report-card stopping points fire AFTER the mirror refreshes (spec
			// 063 US4): live-only by the same construction as matchOrders —
			// replay runs no guardian, so a card can never trigger during
			// reconstruction. Trigger scan runs after the trail collection above
			// so a stopping point in the same batch sees its own batch's acts.
			for _, e := range batch {
				mt.observeCardTrigger(e)
			}
		}
	}
}

// mirrorState refreshes the tiny snapshot the turn worker may read while
// the absorb goroutine keeps ticking the replica.
func (mt *Guardian) mirrorState() {
	mt.stateMu.Lock()
	defer mt.stateMu.Unlock()
	mt.charges = mt.replica.GuardianCharges
	mt.clockAt = mt.replica.Tick
	mt.night = mt.replica.Night
	// The charter mirror only ever moves FORWARD to a recorded value: after
	// observeCharter's optimistic set, batches that predate the just-landed
	// guardian.charter_observed still absorb with the replica's OLD (possibly
	// empty) fingerprint — overwriting would re-open the emission window for
	// an already-recorded revision. The landed event is durable by the time
	// the optimistic set happens, so the replica always catches up.
	if fp := mt.replica.CharterFingerprint; fp != "" {
		mt.charterFP = fp
	}
	// The skills mirror moves forward the same way (spec 077): a recorded
	// observation is the only thing that can advance it.
	if fp := mt.replica.SkillsFingerprint; fp != "" {
		mt.skillsFP = fp
	}
	mt.ended = mt.replica.Ended
	if len(mt.agentXY) != len(mt.replica.Agents) {
		mt.agentXY = make([][2]int, len(mt.replica.Agents))
	}
	if len(mt.agentNeeds) != len(mt.replica.Agents) {
		mt.agentNeeds = make([]needMirror, len(mt.replica.Agents))
	}
	for i := range mt.replica.Agents {
		mt.alive[i] = !mt.replica.Agents[i].Dead
		mt.agentXY[i] = [2]int{mt.replica.Agents[i].X, mt.replica.Agents[i].Y}
		n := mt.replica.Agents[i].Needs
		// Bulk (spec 095 FR-001) mirrors the same absorb batch as health/food/
		// warmth — sim.Bulk is the exported derived-value accessor (spec 013),
		// so this can never drift from the reducer's own bulk() arithmetic.
		mt.agentNeeds[i] = needMirror{Health: n.Health, Food: n.Food, Warmth: n.Warmth,
			Bulk: sim.Bulk(mt.replica.Agents[i].Inv)}
	}
	// Structure/pile tile mirrors (spec 082): position-only, for the turn
	// worker's bundle-effect target resolution probe.
	mt.structXY = mt.structXY[:0]
	mt.structKind = mt.structKind[:0]
	for i := range mt.replica.Structures {
		mt.structXY = append(mt.structXY, [2]int{mt.replica.Structures[i].X, mt.replica.Structures[i].Y})
		mt.structKind = append(mt.structKind, mt.replica.Structures[i].Kind)
	}
	mt.pileXY = mt.pileXY[:0]
	for i := range mt.replica.Piles {
		mt.pileXY = append(mt.pileXY, [2]int{mt.replica.Piles[i].X, mt.replica.Piles[i].Y})
	}
	// The standing-order mirror (spec 029): the replica is the authority, copied
	// so the turn worker reads orders under stateMu without racing the replica.
	mt.orders = append(mt.orders[:0], mt.replica.GuardianOrders...)
	// The plan-layer mirrors (spec 084): same discipline — id minting and the
	// prompt's designation/directive sections read these, never the replica.
	mt.designations = append(mt.designations[:0], mt.replica.Designations...)
	mt.directives = append(mt.directives[:0], mt.replica.Directives...)
	// The prophecy + faith mirrors (spec 085): same discipline again — the
	// turn prompt's faith line and prophecy section read these copies.
	mt.prophecies = append(mt.prophecies[:0], mt.replica.Prophecies...)
	mt.faith = mt.replica.FaithScore()
	// Named regions (spec 101): same discipline again.
	mt.regions = append(mt.regions[:0], mt.replica.Regions...)
	// The myth briefing (spec 101 D5) is DERIVED, not mirrored verbatim: a
	// pure read over the replica's belief corpus, recomputed here (under the
	// SAME lock that guards every other replica read a turn may later copy
	// from) so brief_myths never races the absorb goroutine's unlocked
	// mt.replica.Apply calls.
	mt.myths = mt.replica.DominantPlaceMyths(mythBriefingTopN)
	// The narrated chronicle (TASK-11) is the village's own story — the
	// guardian reads its tail so conversation is grounded even before its
	// soul has accreted (fresh reigns, upgraded worlds).
	mt.story = mt.story[:0]
	ring := mt.replica.Chronicle
	for i := maxOf(0, len(ring)-8); i < len(ring); i++ {
		mt.story = append(mt.story, fmt.Sprintf("day %d [%s] %s", ring[i].Day, ring[i].Thread, ring[i].Text))
	}
}

func maxOf(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// appendFile appends a block to one of Guardian's files (soul/transcript).
func (mt *Guardian) appendFile(path, block string) {
	mt.fileMu.Lock()
	defer mt.fileMu.Unlock()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprint(f, block)
}

// tailOfFile returns up to n trailing bytes of a file ("" on any error).
func tailOfFile(path string, n int64) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return ""
	}
	off := st.Size() - n
	if off < 0 {
		off = 0
	}
	buf := make([]byte, st.Size()-off)
	if _, err := f.ReadAt(buf, off); err != nil {
		return ""
	}
	return string(buf)
}

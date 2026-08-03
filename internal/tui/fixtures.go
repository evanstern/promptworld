package tui

// The design-harness fixtures (spec 112, TASK-187): three canned worlds the
// frame harness (design.go) renders, so looking at a real frame costs one
// command instead of a compile-and-run cycle.
//
// Every fixture is built IN PROCESS as a Go value — world manifest, terrain,
// event feed, villager roster, canned ipc.StatusData — and never by shelling
// out to `promptworld new`. That is deliberate on two counts (spec.md
// FR-001): it sidesteps the earned-stage gate entirely (spec 046 refuses an
// unearned stage without --override, so a shell-out would build differently
// on two operators' machines), and it keeps the RECIPE version-controlled
// rather than a generated world directory whose logs churn on every
// regeneration.
//
// Determinism (FR-003) rests on four pins, all of them here:
//
//  1. the world is a manifest value, so terrain regenerates from (seed, size,
//     terrain_gen) with no disk read;
//  2. the per-user records are canned (fixturePerUserState) instead of read
//     from the operator's home directory — plan.md F1, the single most
//     important constraint in this spec;
//  3. the wall clock is frozen (fixtureNow) — plan.md F4;
//  4. the status snapshot is canned, so the header's tick/day/time/speed are
//     fixed rather than polled from a daemon.

import (
	"encoding/json"
	"fmt"

	"time"

	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/ipc"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/world"
	"github.com/evanstern/promptworld/internal/worldmap"
	"github.com/evanstern/promptworld/internal/worlds"
)

// The three fixture ids (spec.md FR-001's table).
const (
	FixtureEmpty    = "empty"
	FixtureMidGame  = "mid-game"
	FixtureScenario = "scenario"
)

// fixtureNow is the frozen instant every fixture Model's wall clock returns
// (plan.md F4). Nothing in the render path reads it — the lessons
// projection is its only consumer — but freezing it is what keeps the
// projection's spacing/decay arithmetic (lessons.go) from depending on how
// long the harness took to build the feed.
var fixtureNow = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

// fixtureDir is the directory a fixture's world claims. Nothing in the
// render path ever opens it: a fixture world is a Go value, not a directory
// on disk. A fixed sentinel (rather than, say, a temp dir) is what keeps two
// machines' frames identical.
const fixtureDir = "/promptworld/fixtures"

// Fixture is one canned world the harness can render. ID and Description are
// the harness's public selector and listing text; the recipe itself
// (manifest + pose) stays unexported — a fixture is a curated scene, not a
// construction kit callers assemble their own variants from.
type Fixture struct {
	ID          string
	Description string

	// manifest is the world this fixture renders against — the ONLY world
	// input, since terrain regenerates from it (world.World.Map) and every
	// other world file is irrelevant to rendering.
	manifest world.Manifest
	// pose seeds the freshly constructed Model: connection posture, canned
	// status, event feed, roster edits. Nil = the bare cold-start model.
	pose func(m *Model) error
}

// fixtures is the registry, in listing order.
var fixtures = []Fixture{emptyFixture(), midGameFixture(), scenarioFixture()}

// Fixtures returns the shipped fixture recipes in listing order. The slice is
// a copy: the registry is not a caller-mutable surface.
func Fixtures() []Fixture {
	out := make([]Fixture, len(fixtures))
	copy(out, fixtures)
	return out
}

// FixtureIDs is the id list, for a CLI's --list and for error messages.
func FixtureIDs() []string {
	out := make([]string, 0, len(fixtures))
	for _, f := range fixtures {
		out = append(out, f.ID)
	}
	return out
}

// fixtureByID resolves a selector to its recipe.
func fixtureByID(id string) (Fixture, bool) {
	for _, f := range fixtures {
		if f.ID == id {
			return f, true
		}
	}
	return Fixture{}, false
}

// fixturePerUserState is the canned per-user record every fixture is built
// with (plan.md F1): nothing seen, nothing unlocked, and a no-op writer — so
// rendering a frame neither reads nor writes the operator's home directory.
// Freshly allocated per call so two fixtures can never share (and mutate)
// one record.
func fixturePerUserState() perUserState {
	return perUserState{
		lessonsSeen: &worlds.LessonsSeen{Entries: map[string]worlds.LessonSeenEntry{}},
		unlocks:     &worlds.Unlocks{Entries: map[string]worlds.UnlockEntry{}},
		markSeen:    func(string, string) {}, // never persist a fixture's lessons
	}
}

// world is the fixture's world value — a manifest and a nominal directory,
// no files.
func (f Fixture) world() *world.World {
	return &world.World{Dir: fixtureDir + "/" + f.ID, Manifest: f.manifest}
}

// build constructs the fixture's posed Model with its canned per-user record.
func (f Fixture) build() (Model, error) { return f.buildWith(fixturePerUserState()) }

// buildWith is build's injectable form: the tests use it to construct the
// SAME fixture against the live (on-disk) per-user record, which is how the
// determinism proof shows the canned record is actually load-bearing rather
// than incidentally equal.
func (f Fixture) buildWith(pu perUserState) (Model, error) {
	w := f.world()
	m := newModel(w, pu, func() time.Time { return fixtureNow })
	m.replica = sim.NewState(f.manifest.Seed, m.gameMap)
	if f.pose != nil {
		if err := f.pose(&m); err != nil {
			return Model{}, fmt.Errorf("fixture %s: %w", f.ID, err)
		}
	}
	return m, nil
}

// fixtureManifest is the shared manifest shape — everything a fixture world
// needs and nothing it doesn't. CreatedAt is deliberately empty: it is
// per-creation provenance the renderer never shows, and a timestamp here
// would be one more thing that could differ between two builds.
func fixtureManifest(name string, seed uint64, stage string) world.Manifest {
	return world.Manifest{
		Name:            name,
		Seed:            seed,
		FormatVersion:   world.FormatVersion,
		TickGameSeconds: 1,
		MapWidth:        worldmap.DefaultSize,
		MapHeight:       worldmap.DefaultSize,
		TerrainGen:      worldmap.GenMarshSand,
		Stage:           stage,
	}
}

// --- empty ---------------------------------------------------------------

// emptyFixture is the cold start (FR-001): a fresh stage-1 world with no
// events and no daemon — the disconnected header, the zero-event chronicle,
// and every pane's empty state. The genesis roster is present because that
// IS a fresh world's state; what's absent is any history over it.
func emptyFixture() Fixture {
	return Fixture{
		ID:          FixtureEmpty,
		Description: "fresh stage-1 world, no events, no daemon — cold start and disconnected rendering",
		manifest:    fixtureManifest("Ashgrove", 42, world.Stage1),
		// No pose: connected stays false and status stays nil, which is
		// exactly the "no daemon" posture — headerView renders the
		// disconnected line and every pane falls back to its empty state.
		pose: nil,
	}
}

// --- mid-game ------------------------------------------------------------

// midGameTick is the canned clock position: day 5, 20:00 on the genesis
// calendar (clock.Format renders it, so the header can never drift from the
// tick it claims).
const midGameTick = 396000

// midGameFeedLen is the chronicle backlog depth. The deepest chronicle pane
// in the matrix is well under 100 rows at 50 lines tall, so this overflows
// every size the harness renders (AC #6: truncation must be visible).
const midGameFeedLen = 240

// midGameFixture is the dense case (FR-001): a populated later-stage ambient
// world with a deep event backlog and a mixed roster — awake, asleep AND
// dead villagers (AC #6). Stage 2 rather than a higher rung is deliberate:
// the lesson row's stage default is on at stages 1-2 only
// (patterns/stage-defaults.md), and this is the fixture whose job is to show
// all three teaching chrome rows at once — villager strip, lesson row,
// guardian strip — so it must sit where the row is eligible to render.
func midGameFixture() Fixture {
	return Fixture{
		ID:          FixtureMidGame,
		Description: "populated stage-2 ambient world — deep chronicle backlog, mixed awake/asleep/dead roster",
		manifest:    fixtureManifest("Ashgrove", 7, world.Stage2),
		pose: func(m *Model) error {
			m.connected = true
			m.status = midGameStatus()
			for _, e := range midGameFeed() {
				m.applyEvent(e)
			}
			// Standing orders (the guardian strip's 👁 segment) — client
			// replica state, exactly what a live metatron.order_placed fold
			// would have left behind.
			m.replica.GuardianOrders = []sim.GuardianOrder{
				{ID: "ord-372000-1", Origin: "player", Condition: "Rowan falls asleep away from the fire",
					Status: "active", PlacedTick: 372000, ExpiresTick: 890000},
				{ID: "ord-381000-1", Origin: "system", Condition: "the gru comes within sight of the village",
					Status: "active", PlacedTick: 381000, ExpiresTick: 900000},
			}
			return nil
		},
	}
}

// midGameStatus is the canned daemon snapshot (plan.md F4): every field the
// header, guardian strip and systems tab read, fixed. GameTime is derived
// from the tick rather than written out, so the two can never disagree.
func midGameStatus() *ipc.StatusData {
	faith := 62
	return &ipc.StatusData{
		World: ipc.WorldStatus{
			Name: "Ashgrove", Seed: 7, FormatVersion: world.FormatVersion, Stage: world.Stage2,
		},
		Clock: ipc.ClockStatus{
			Tick:            midGameTick,
			GameTime:        clock.Format(midGameTick),
			Speed:           string(clock.Speed8x),
			EffectiveRate:   clock.Speed8x.TicksPerSecond(),
			GuardianCharges: 2,
			Faith:           &faith,
			FaithRegenTicks: 12 * 3600,
		},
		Daemon: ipc.DaemonStatus{Pid: 4242, UptimeSeconds: 5400, Subscribers: 1},
		Log:    ipc.LogStatus{LastSeq: midGameFeedLen},
	}
}

// midGameFeed is the chronicle backlog: a deterministic cycle of ordinary
// village activity, closed out by the roster-shaping events — two sleepers,
// a gru attack, and two deaths — so the roster's three states and the
// chronicle's alert tier are both real folds of real events rather than
// hand-set struct fields.
//
// Ticks run backwards from midGameTick so the feed's tail meets the canned
// clock: a chronicle whose newest line sits far from the header's tick would
// be an incoherent scene.
func midGameFeed() []store.Event {
	const spacing = 163
	evs := make([]store.Event, 0, midGameFeedLen+8)
	var seq int64
	tickFor := func(i int) int64 { return midGameTick - int64(midGameFeedLen+8-i)*spacing }
	add := func(typ, payload string) {
		evs = append(evs, store.Event{
			Seq: seq + 1, Tick: tickFor(len(evs)), Type: typ, Payload: json.RawMessage(payload),
		})
		seq++
	}
	// The ambient body: eight-event cycle over the roster, so every family
	// tint and both chronicle column widths get exercised.
	for i := 0; i < midGameFeedLen; i++ {
		id := i % len(sim.AgentNames)
		who := agentRef(id)
		x, y := 12+(i*7)%40, 9+(i*5)%40
		switch i % 8 {
		case 0:
			add("agent.moved", fmt.Sprintf(`{"agent":%s,"x":%d,"y":%d}`, who, x, y))
		case 1:
			add("agent.intent_set", fmt.Sprintf(
				`{"agent":%s,"goal":"forage","target_x":%d,"target_y":%d,"res_x":0,"res_y":0,"source":"reflex"}`, who, x, y))
		case 2:
			add("agent.moved", fmt.Sprintf(`{"agent":%s,"x":%d,"y":%d}`, who, x+1, y))
		case 3:
			add("agent.intent_done", fmt.Sprintf(`{"agent":%s}`, who))
		case 4:
			add("sim.forage_regrown", fmt.Sprintf(`{"x":%d,"y":%d}`, x, y))
		case 5:
			add("agent.saw", fmt.Sprintf(
				`{"agent":%s,"facts":[{"kind":"tree","x":%d,"y":%d,"seen":%d,"prov":"witnessed"}]}`, who, x, y, tickFor(i)))
		case 6:
			add("agent.moved", fmt.Sprintf(`{"agent":%s,"x":%d,"y":%d}`, who, x, y+1))
		case 7:
			if i%64 == 7 {
				add("sim.night_started", fmt.Sprintf(`{"day":%d}`, 1+i/64))
			} else {
				add("clock.speed_set", `{"speed":"8x"}`)
			}
		}
	}
	// The roster's three states (AC #6). Sleepers first, then the gru, then
	// the deaths — the order matters only in that the FIRST lesson-bearing
	// event decides which lesson occupies the row, and the gru's is the one
	// this scene is about.
	add("agent.slept", fmt.Sprintf(`{"agent":%s}`, agentRef(4)))
	add("agent.slept", fmt.Sprintf(`{"agent":%s}`, agentRef(5)))
	add("gru.attacked", fmt.Sprintf(`{"agent":%s,"health":40}`, agentRef(0)))
	add("agent.died", fmt.Sprintf(`{"agent":%s,"cause":"starvation"}`, agentRef(6)))
	add("agent.died", fmt.Sprintf(`{"agent":%s,"cause":"exposure"}`, agentRef(7)))
	add("sim.neglect_detected", fmt.Sprintf(
		`{"agent":%s,"need":"warmth","level":0,"since":%d}`, agentRef(1), midGameTick-3000))
	add("agent.moved", fmt.Sprintf(`{"agent":%s,"x":%d,"y":%d}`, agentRef(2), 24, 31))
	add("agent.intent_done", fmt.Sprintf(`{"agent":%s}`, agentRef(3)))
	return evs
}

// agentRef renders the named agent reference every modern payload carries
// (spec 086): {"id":N,"name":"…"}.
func agentRef(id int) string {
	return fmt.Sprintf(`{"id":%d,"name":%q}`, id, sim.AgentNames[id])
}

// --- scenario ------------------------------------------------------------

// scenarioExercise is the catalogued exercise the scenario fixture runs
// (spec 054's catalog, sim.ScenarioExercises). first-night is the stage-1
// entry — stage 1 is also where the lesson row's default is on, so this one
// fixture shows both surfaces AC #7 asks for.
const scenarioExercise = "first-night"

// scenarioTick is the canned clock: day 1, 21:12 — dusk on the first night,
// with the exercise still in progress.
const scenarioTick = 54720

// scenarioFixture is the exercise case (FR-001): a --scenario world from the
// spec 054 catalog, so the exercise dock tab and the lesson row both render —
// neither of which any ambient world has (AC #7).
func scenarioFixture() Fixture {
	man := fixtureManifest("First Night", 1101, world.Stage1)
	man.Scenario = &world.ScenarioConfig{Exercise: scenarioExercise}
	return Fixture{
		ID:          FixtureScenario,
		Description: "stage-1 --scenario world running the first-night exercise — exercise tab and lesson row",
		manifest:    man,
		pose: func(m *Model) error {
			def, ok := sim.ExerciseByID(scenarioExercise)
			if !ok {
				return fmt.Errorf("exercise %q is not in the shipped catalog", scenarioExercise)
			}
			if err := m.replica.ArmScenario(def); err != nil {
				return fmt.Errorf("arming %q: %w", scenarioExercise, err)
			}
			m.connected = true
			m.status = scenarioStatus()
			// The briefing is dismissed so the tab shows its live rubric
			// gauges: the briefing is a one-per-attach interstitial, the
			// gauges are the tab's steady state and the thing worth looking
			// at in a design frame.
			m.exBriefingDismissed = true
			for _, e := range scenarioFeed() {
				m.applyEvent(e)
			}
			return nil
		},
	}
}

func scenarioStatus() *ipc.StatusData {
	faith := 50
	return &ipc.StatusData{
		World: ipc.WorldStatus{
			Name: "First Night", Seed: 1101, FormatVersion: world.FormatVersion, Stage: world.Stage1,
			ScenarioExercise: scenarioExercise, ScenarioOutcome: sim.OutcomeInProgress,
		},
		Clock: ipc.ClockStatus{
			Tick:            scenarioTick,
			GameTime:        clock.Format(scenarioTick),
			Speed:           string(clock.Speed4x),
			EffectiveRate:   clock.Speed4x.TicksPerSecond(),
			GuardianCharges: 3,
			Faith:           &faith,
			FaithRegenTicks: 6 * 3600,
		},
		Daemon: ipc.DaemonStatus{Pid: 4343, UptimeSeconds: 900, Subscribers: 1},
		Log:    ipc.LogStatus{LastSeq: 24},
	}
}

// scenarioFeed is a short first-night backlog — enough to populate the
// chronicle honestly and to surface one lesson (the gru attack, which is
// what first-night is about), without the mid-game fixture's overflow.
func scenarioFeed() []store.Event {
	const spacing = 720
	var evs []store.Event
	var seq int64
	add := func(typ, payload string) {
		seq++
		evs = append(evs, store.Event{
			Seq: seq, Tick: scenarioTick - int64(24-len(evs))*spacing,
			Type: typ, Payload: json.RawMessage(payload),
		})
	}
	add("world.created", `{"name":"First Night","seed":1101}`)
	add("sim.day_started", `{"day":1}`)
	for i := 0; i < 18; i++ {
		id := i % len(sim.AgentNames)
		who := agentRef(id)
		x, y := 20+(i*3)%20, 18+(i*4)%20
		switch i % 3 {
		case 0:
			add("agent.moved", fmt.Sprintf(`{"agent":%s,"x":%d,"y":%d}`, who, x, y))
		case 1:
			add("agent.intent_set", fmt.Sprintf(
				`{"agent":%s,"goal":"chop","target_x":%d,"target_y":%d,"res_x":0,"res_y":0,"source":"reflex"}`, who, x, y))
		default:
			add("agent.intent_done", fmt.Sprintf(`{"agent":%s}`, who))
		}
	}
	add("sim.night_started", `{"day":1}`)
	add("gru.sighted", fmt.Sprintf(`{"x":%d,"y":%d}`, 27, 22))
	add("gru.attacked", fmt.Sprintf(`{"agent":%s,"health":55}`, agentRef(3)))
	add("agent.slept", fmt.Sprintf(`{"agent":%s}`, agentRef(5)))
	return evs
}

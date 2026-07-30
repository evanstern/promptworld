package guardian

// Spec 085 guardian-side suites (T014/T015): the prophesy handler — target
// resolution, kind-conditional claim assembly (partial/foreign args refused),
// deterministic pro-<tick>-<seq> id minting, the atomic declaration batch
// with its per-target OriginOmen companions (the firewall's
// recorded-data-only channel), door rejections as counsel — plus the
// claim-kind mirror drift pin, the prompt's faith line and prophecy section,
// and the prophecy.failed standing-order composition proof (enum-only
// observability, zero new trigger code).

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/tool"
)

// TestProphecyClaimKindsMirrorSim pins prophesy's claim_kind Enum equal to
// the sim claim vocabulary (the DesignationKinds pattern) — the mirror cannot
// drift silently, and the handler's assembleClaim covers exactly it.
func TestProphecyClaimKindsMirrorSim(t *testing.T) {
	want := []string{sim.ProphecyDesignationFulfilled, sim.ProphecyStructureCount,
		sim.ProphecyPopulationAtLeast, sim.ProphecySurvives}
	if got := tool.ProphecyClaimKinds(); !reflect.DeepEqual(got, want) {
		t.Errorf("tool.ProphecyClaimKinds() = %v, want %v", got, want)
	}
}

// prophesyFixtureArgs is a legal everyone-targeted structure_count call.
func prophesyFixtureArgs() *prophesyArgs {
	return &prophesyArgs{
		targets:       "everyone",
		text:          "Before three dawns a shelter will stand.",
		claimKind:     sim.ProphecyStructureCount,
		structureKind: "shelter",
		min:           1,
		hasMin:        true,
	}
}

// TestProphesyLands: the declaration lands ONE atomic batch —
// prophecy.declared plus one OriginOmen dream-band companion memory per
// living target — with the deterministic id, the default 3-day deadline, the
// Village marker, and the reducer-side charge spend.
func TestProphesyLands(t *testing.T) {
	mt, inj, grant := planFixture(t)
	alive := map[int]bool{}
	for i := range inj.state.Agents {
		alive[i] = true
	}

	p, why := mt.landProphesy(prophesyFixtureArgs(), 1, 100, alive, grant)
	if why != "" {
		t.Fatalf("prophesy refused: %q", why)
	}
	if p.ID != "pro-100-0" {
		t.Errorf("id = %q, want pro-100-0 (the nextOrderID shape)", p.ID)
	}
	if !p.Village || len(p.Targets) != len(inj.state.Agents) {
		t.Errorf("everyone resolution = village %v, %d targets", p.Village, len(p.Targets))
	}
	if p.DeadlineTick != 100+3*ticksPerGameDay {
		t.Errorf("deadline = %d, want the 3-day default", p.DeadlineTick)
	}

	if len(inj.batches) != 1 {
		t.Fatalf("landed %d batches, want 1 (atomic)", len(inj.batches))
	}
	batch := inj.batches[0]
	if batch[0].Type != "prophecy.declared" {
		t.Fatalf("batch[0] = %s, want prophecy.declared", batch[0].Type)
	}
	memories := 0
	for _, e := range batch[1:] {
		if e.Type != "agent.memory_added" {
			t.Fatalf("unexpected companion %s", e.Type)
		}
		var mp sim.MemoryAddedPayload
		if err := json.Unmarshal(e.Payload, &mp); err != nil {
			t.Fatal(err)
		}
		if mp.Origin != sim.OriginOmen || mp.Salience != sim.SalDream {
			t.Errorf("companion origin=%q sal=%d, want omen/dream band", mp.Origin, mp.Salience)
		}
		if !strings.HasPrefix(mp.Text, "The Guardian foretells: ") {
			t.Errorf("companion text = %q, want the frozen foretells prefix", mp.Text)
		}
		memories++
	}
	if memories != len(p.Targets) {
		t.Errorf("companions = %d, want one per target (%d)", memories, len(p.Targets))
	}

	// The reducer landed it active and spent the stake.
	if len(inj.state.Prophecies) != 1 || inj.state.Prophecies[0].Status != "active" {
		t.Fatalf("injector state prophecies = %+v", inj.state.Prophecies)
	}
	if inj.state.GuardianCharges != sim.GuardianGenesisCharges-1 {
		t.Errorf("charges = %d, want the stake spent", inj.state.GuardianCharges)
	}

	// Same-tick second mint disambiguates (needs a fresh charge for the door).
	inj.state.GuardianCharges = 1
	a2 := prophesyFixtureArgs()
	a2.min = 2 // a different claim; the id is what's under test
	if p2, why := mt.landProphesy(a2, 1, 100, alive, grant); why != "" {
		t.Fatalf("second prophesy refused: %q", why)
	} else if p2.ID != "pro-100-1" {
		t.Errorf("second id = %q, want pro-100-1", p2.ID)
	}
}

// TestProphesyClaimArgsTable: kind-conditional assembly refuses partial and
// foreign argument sets BEFORE anything reaches the door (the parseReveal
// shape), and the turn-side pre-checks refuse the obvious counsel cases.
func TestProphesyClaimArgsTable(t *testing.T) {
	mt, inj, grant := planFixture(t)
	alive := map[int]bool{}
	for i := range inj.state.Agents {
		alive[i] = true
	}
	cases := []struct {
		name    string
		mutate  func(*prophesyArgs)
		charges int
		want    string
	}{
		{"unknown kind", func(a *prophesyArgs) { a.claimKind = "weather" }, 1, "no claim of kind"},
		{"structure_count missing min", func(a *prophesyArgs) { a.hasMin = false }, 1, "give the count"},
		{"structure_count missing kind", func(a *prophesyArgs) { a.structureKind = "" }, 1, "name the structure kind"},
		{"foreign arg for structure_count", func(a *prophesyArgs) { a.agentName = sim.AgentNames[0] }, 1, "nothing else"},
		{"designation_fulfilled with min", func(a *prophesyArgs) {
			a.claimKind = sim.ProphecyDesignationFulfilled
			a.designationID = "dsg-1-0"
		}, 1, "nothing else"},
		{"survives unknown villager", func(a *prophesyArgs) {
			a.claimKind = sim.ProphecySurvives
			a.structureKind, a.hasMin = "", false
			a.agentName = "Nobody"
		}, 1, "no villager named"},
		{"empty text", func(a *prophesyArgs) { a.text = " " }, 1, "give me the word"},
		{"deadline out of bounds", func(a *prophesyArgs) { a.deadlineDays = 9 }, 1, "days"},
		{"no charge banked", func(*prophesyArgs) {}, 0, "needs a stake"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a := prophesyFixtureArgs()
			c.mutate(a)
			p, why := mt.landProphesy(a, c.charges, 100, alive, grant)
			if p != nil || !strings.Contains(why, c.want) {
				t.Fatalf("landProphesy = (%v, %q), want refusal containing %q", p, why, c.want)
			}
		})
	}
	if len(inj.batches) != 0 {
		t.Fatalf("%d batches landed from refused calls", len(inj.batches))
	}
}

// TestProphesyDoorCounsel: reducer-door rejections come back as repairable
// in-fiction counsel — the already-true claim and the active duplicate.
func TestProphesyDoorCounsel(t *testing.T) {
	mt, inj, grant := planFixture(t)
	alive := map[int]bool{}
	for i := range inj.state.Agents {
		alive[i] = true
	}

	// Prophesying the past: the claim already holds.
	inj.state.Structures = append(inj.state.Structures, sim.Structure{Kind: "shelter", X: 4, Y: 5})
	if _, why := mt.landProphesy(prophesyFixtureArgs(), 1, 100, alive, grant); !strings.Contains(why, "past needs no prophet") {
		t.Fatalf("already-true counsel = %q", why)
	}
	inj.state.Structures = nil

	// The duplicate wager.
	if _, why := mt.landProphesy(prophesyFixtureArgs(), 1, 100, alive, grant); why != "" {
		t.Fatalf("first prophesy refused: %q", why)
	}
	inj.state.GuardianCharges = 1
	if _, why := mt.landProphesy(prophesyFixtureArgs(), 1, 101, alive, grant); !strings.Contains(why, "one word, one wager") {
		t.Fatalf("duplicate counsel = %q", why)
	}
}

// TestProphecyWatchComposition (FR-010, the spec-084 SC-003 shape): a
// standing order placed via the REAL monitor_and_act door watching
// "prophecy.failed" triggers through the UNMODIFIED matchOrders path — zero
// new trigger code: the lifecycle is watchable purely because the three
// prophecy.* types joined the observableEventTypes enum (registry.go).
func TestProphecyWatchComposition(t *testing.T) {
	mt, inj, grant := planFixture(t)

	schema := string(tool.InputSchema(mustLookup(t, "monitor_and_act")))
	for _, typ := range []string{"prophecy.declared", "prophecy.fulfilled", "prophecy.failed"} {
		if !strings.Contains(schema, typ) {
			t.Errorf("monitor_and_act schema enum missing %q", typ)
		}
	}

	order, why := mt.placeOrder("player", orderArgs{
		Condition:  "a prophecy of mine fails",
		Action:     "tell the player the word did not come to pass",
		EventTypes: []string{"prophecy.failed"},
	}, 100, grant)
	if why != "" {
		t.Fatalf("watch refused: %q", why)
	}

	// Declare through the real door, then judge it failed exactly as the
	// executor sweep would record it (its provenance is pinned sim-side).
	alive := map[int]bool{}
	for i := range inj.state.Agents {
		alive[i] = true
	}
	a := prophesyFixtureArgs()
	a.deadlineDays = 1
	p, why := mt.landProphesy(a, 1, 200, alive, grant)
	if why != "" {
		t.Fatalf("prophesy refused: %q", why)
	}
	failed := []store.Event{{Tick: p.DeadlineTick, Type: "prophecy.failed",
		Payload: mustJSON(sim.OrderIDPayload{ID: p.ID})}}
	if err := inj.state.Apply(failed[0]); err != nil {
		t.Fatal(err)
	}

	syncOrdersFromDoor(mt, inj)
	mt.matchOrders(failed)
	select {
	case job := <-mt.triggerQ:
		if job.order.ID != order.ID || job.matchedType != "prophecy.failed" {
			t.Fatalf("trigger job = %+v", job)
		}
	default:
		t.Fatal("prophecy.failed did not trigger the watch through matchOrders")
	}
}

// TestProphecyPromptSections (FR-013): the turn user prompt always carries
// the faith line (in-fiction wording), and active prophecies render with
// id/claim/days-left; settled ones do not.
func TestProphecyPromptSections(t *testing.T) {
	alive := map[int]bool{0: true}
	base := turnUserPrompt(100, 1, 62, alive, nil, nil, nil, nil, nil, nil, nil, "", "", "", "The player says:\nhello")
	if !strings.Contains(base, "faith in you stands at 62 of 100") {
		t.Errorf("prompt missing the faith line:\n%s", base)
	}
	if strings.Contains(base, "Prophecies you have staked") {
		t.Error("prophecy-free prompt carries the prophecy section")
	}

	prophecies := []sim.Prophecy{
		{ID: "pro-1-0", Targets: []int{0}, Text: "A shelter will stand.",
			Claim:        sim.ProphecyClaim{Kind: sim.ProphecyStructureCount, StructureKind: "shelter", Min: 1},
			DeclaredTick: 100, DeadlineTick: 100 + 2*24*3600, Status: "active"},
		{ID: "pro-1-1", Targets: []int{0}, Text: "Old word.",
			Claim:        sim.ProphecyClaim{Kind: sim.ProphecyPopulationAtLeast, Min: 4},
			DeclaredTick: 1, DeadlineTick: 2, Status: "failed"}, // settled — must not render
	}
	got := turnUserPrompt(100, 1, 62, alive, nil, nil, nil, prophecies, nil, nil, nil, "", "", "", "The player says:\nhello")
	for _, want := range []string{
		"Prophecies you have staked (the word, once given, stands):",
		`- pro-1-0: "A shelter will stand." — judged by at least 1 shelter standing (2 day(s) left)`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("prompt missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "pro-1-1") {
		t.Error("settled prophecy rendered")
	}
}

// TestFaithBandWordCoversCurve: the prompt's in-fiction band words track the
// data-model §6 vocabulary at every band edge (prompt prose only — never a
// recorded payload).
func TestFaithBandWordCoversCurve(t *testing.T) {
	cases := map[int]string{
		90: "believes", 75: "believes",
		74: "covenant", 40: "covenant",
		39: "doubt", 15: "doubt",
		14: "nearly dry", 0: "nearly dry",
	}
	for score, want := range cases {
		if got := faithBandWord(score); !strings.Contains(got, want) {
			t.Errorf("faithBandWord(%d) = %q, want it to contain %q", score, got, want)
		}
	}
}

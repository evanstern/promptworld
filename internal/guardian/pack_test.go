package guardian

// Spec 116 US1/US2 suites: the inspect_pack sheet (the world-03 fixture, byte
// identity, the honest miss, spears and the empty pack), its neutrality (no
// charge, no event, never the turn's act), and the look-first gate on a
// survival watch turn (per-villager, repairable, origin-scoped).

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/tool"
	"github.com/evanstern/promptworld/internal/toolloop"
)

// setInventory writes one villager's inventory into BOTH the guardian's replica
// and the injector's authoritative state, then refreshes the mirror — the shape
// every turn-side read expects (the absorb goroutine is stopped in these
// fixtures, so mirrorState is called explicitly).
func setInventory(mt *Guardian, inj *stateInjector, idx int, inv sim.Inventory) {
	mt.replica.Agents[idx].Inv = inv
	inj.state.Agents[idx].Inv = inv
	mt.mirrorState()
}

// packDispatch builds a turn dispatch for the handler-level suites. survival
// selects the origin the look-first gate keys on.
func packDispatch(mt *Guardian, survival bool) *turnDispatch {
	alive := map[int]bool{}
	for i := range sim.AgentNames {
		alive[i] = true
	}
	return &turnDispatch{mt: mt, charges: 3, alive: alive, tick: 1,
		result: &TurnResult{}, grant: fullGrant(), survival: survival, looked: map[int]bool{}}
}

// packCall drives handleInspectPack directly with raw args (the surveyCall
// shape), against a caller-supplied dispatch so the ledger it populates stays
// observable to the gate assertions below.
func packCall(mt *Guardian, d *turnDispatch, args string) toolloop.Outcome {
	h := mt.handleInspectPack(d)
	return h(context.Background(), llm.ToolCall{Name: "inspect_pack", Args: []byte(args)})
}

// cedarsPack is the world-03 fixture: a full pack (24/24) of pure deadweight —
// twenty wood and four planks, not one bite of food. This is the state the
// guardian misread as "you are carrying food, eat now".
var cedarsPack = sim.Inventory{Wood: 20, Planks: 4}

// TestPackSheetWorldO3Fixture (SC-001, AC#1, T009): the sheet names both kinds
// with their counts, the carry line reads 24/24 with 0 free, and it states in
// plain words that the villager carries no food.
func TestPackSheetWorldO3Fixture(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "behold")
	cedar := agentIndexByName("Cedar")
	setInventory(mt, inj, cedar, cedarsPack)

	out := packCall(mt, packDispatch(mt, false), `{"villager":"Cedar"}`)
	if out.Verdict != toolloop.VerdictReadOK {
		t.Fatalf("verdict = %v, want read_ok", out.Verdict)
	}
	for _, want := range []string{
		"Cedar's pack — carrying 24/24, 0 free to receive:",
		"- wood: 20",
		"- planks: 4",
		"Cedar carries no food.",
	} {
		if !strings.Contains(out.ResultForModel, want) {
			t.Errorf("sheet missing %q:\n%s", want, out.ResultForModel)
		}
	}
	// The cap in the header is sim.BulkCap, not a literal that could drift.
	if !strings.Contains(out.ResultForModel, "/"+itoa(sim.BulkCap)+",") {
		t.Errorf("sheet does not render the cap from sim.BulkCap:\n%s", out.ResultForModel)
	}
}

// TestPackSheetFoodLine (FR-002): a villager carrying food gets the other form
// of the mandatory closing line, listing only the non-zero food kinds in the
// fixed order — the distinction between "has no food" and "has food and will
// not eat" that gross bulk could never express.
func TestPackSheetFoodLine(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "behold")
	cedar := agentIndexByName("Cedar")
	setInventory(mt, inj, cedar, sim.Inventory{Wood: 2, FoodCooked: 3, Meals: 1})

	sheet := packCall(mt, packDispatch(mt, false), `{"villager":"Cedar"}`).ResultForModel
	if !strings.Contains(sheet, "Cedar carries food: 3 food_cooked, 1 meals.") {
		t.Errorf("food line wrong:\n%s", sheet)
	}
	if strings.Contains(sheet, "food_raw") {
		t.Errorf("sheet lists a food kind at zero:\n%s", sheet)
	}
}

// TestPackSheetByteIdentity (FR-003, SC-001, T010): two identical calls in one
// turn return byte-identical sheets — the sheet is a pure function of (name,
// mirrored contents) with a fixed kind order and no clock read.
func TestPackSheetByteIdentity(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "behold")
	setInventory(mt, inj, agentIndexByName("Cedar"), cedarsPack)
	d := packDispatch(mt, false)

	a := packCall(mt, d, `{"villager":"Cedar"}`)
	b := packCall(mt, d, `{"villager":"Cedar"}`)
	if a.ResultForModel != b.ResultForModel {
		t.Errorf("identical calls returned different sheets:\n%s\n---\n%s", a.ResultForModel, b.ResultForModel)
	}
}

// TestPackSheetNeutrality (FR-004, T008): looking is looking, not an act — no
// charge moves, no event lands, and the acting-cardinality exemption is
// STRUCTURAL (Effect Read is what the loop driver's isActing check keys on, the
// survey_site/explain precedent).
func TestPackSheetNeutrality(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "behold")
	setInventory(mt, inj, agentIndexByName("Cedar"), cedarsPack)
	chargesBefore := inj.state.GuardianCharges
	batchesBefore := len(inj.batches)

	packCall(mt, packDispatch(mt, false), `{"villager":"Cedar"}`)

	if len(inj.batches) != batchesBefore {
		t.Error("a pack read injected events")
	}
	if inj.state.GuardianCharges != chargesBefore {
		t.Error("a pack read moved the charge bank")
	}
	tl, ok := tool.Lookup("inspect_pack")
	if !ok {
		t.Fatal("inspect_pack is not registered")
	}
	if tl.Effect != tool.Read || tl.Gate != tool.None {
		t.Errorf("inspect_pack Effect/Gate = %v/%v, want Read/None", tl.Effect, tl.Gate)
	}
	if tl.Cost.Charges != 0 {
		t.Errorf("inspect_pack costs %d charges, want 0", tl.Cost.Charges)
	}
}

// TestPackSheetHonestMiss (FR-005, T011): an unknown name, a dead villager, and
// a missing argument each return read_ok with a roster-naming miss — never an
// error verdict, and never a ledger entry (the gate's evidence is a SUCCESSFUL
// look, and the dead-target refusals still stand on their own).
func TestPackSheetHonestMiss(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "behold")
	dead := agentIndexByName("Cedar")
	mt.replica.Agents[dead].Dead = true
	inj.state.Agents[dead].Dead = true
	mt.mirrorState()
	d := packDispatch(mt, true)

	for _, c := range []struct{ name, args string }{
		{"unknown", `{"villager":"Mordred"}`},
		{"dead", `{"villager":"Cedar"}`},
		{"absent", `{}`},
	} {
		out := packCall(mt, d, c.args)
		if out.Verdict != toolloop.VerdictReadOK {
			t.Errorf("%s: verdict = %v, want read_ok", c.name, out.Verdict)
		}
		if !strings.Contains(out.ResultForModel, "the living are: ") {
			t.Errorf("%s: miss does not name the living roster: %q", c.name, out.ResultForModel)
		}
		if !strings.Contains(out.ResultForModel, sim.AgentNames[0]) {
			t.Errorf("%s: roster listing omits a living villager: %q", c.name, out.ResultForModel)
		}
	}
	if len(d.looked) != 0 {
		t.Errorf("an honest miss recorded a look: %v", d.looked)
	}
}

// TestPackSheetEmptyAndTools (FR-002, T012): an empty pack renders the
// empty-pack line with 0/24 and full headroom; spears and axes render their
// count plus their remaining uses, in the slice's own ascending (most-worn
// first) order — the order a removal would take them in.
func TestPackSheetEmptyAndTools(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "behold")
	cedar := agentIndexByName("Cedar")
	setInventory(mt, inj, cedar, sim.Inventory{})

	empty := packCall(mt, packDispatch(mt, false), `{"villager":"Cedar"}`).ResultForModel
	if !strings.Contains(empty, "Cedar's pack — carrying 0/24, 24 free to receive: the pack is empty.") {
		t.Errorf("empty-pack sheet wrong:\n%s", empty)
	}
	if !strings.Contains(empty, "Cedar carries no food.") {
		t.Errorf("empty pack lost the mandatory food line:\n%s", empty)
	}

	setInventory(mt, inj, cedar, sim.Inventory{Spears: []int{3, 7}, Axes: []int{10}})
	tools := packCall(mt, packDispatch(mt, false), `{"villager":"Cedar"}`).ResultForModel
	if !strings.Contains(tools, "- spears: 2 (uses left: 3, 7)") {
		t.Errorf("spear rendering wrong:\n%s", tools)
	}
	if !strings.Contains(tools, "- axes: 1 (uses left: 10)") {
		t.Errorf("axe rendering wrong:\n%s", tools)
	}
}

// TestPackMirrorCopiesToolSlices (FR-006, T003): the mirrored contents match the
// absorbed state, and mutating the replica's Spears slice afterwards does NOT
// reach the mirror — the absorb goroutine keeps mutating those slices in place,
// so an aliased mirror would let a sheet report a durability that has changed
// since the snapshot.
func TestPackMirrorCopiesToolSlices(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "behold")
	cedar := agentIndexByName("Cedar")
	setInventory(mt, inj, cedar, sim.Inventory{Wood: 5, Spears: []int{1, 2}, Axes: []int{4}})

	if got := mt.agentInv[cedar]; got.Wood != 5 || len(got.Spears) != 2 || len(got.Axes) != 1 {
		t.Fatalf("mirror did not absorb the contents: %+v", got)
	}
	// The replica keeps living: a hunt spends the most-worn spear.
	mt.replica.Agents[cedar].Inv.Spears[0] = 99
	mt.replica.Agents[cedar].Inv.Axes[0] = 99
	if mt.agentInv[cedar].Spears[0] == 99 || mt.agentInv[cedar].Axes[0] == 99 {
		t.Error("the mirrored tool slices alias the replica's — a sheet could report a stale durability")
	}
	// And the sheet itself reads the mirror, not the replica.
	sheet := packCall(mt, packDispatch(mt, false), `{"villager":"Cedar"}`).ResultForModel
	if !strings.Contains(sheet, "uses left: 1, 2") {
		t.Errorf("sheet did not render the mirrored uses:\n%s", sheet)
	}
}

// --- US2: the look-first gate ---

// visionCall drives handleVision against a dispatch (the gate's own door).
func visionCall(mt *Guardian, d *turnDispatch, args string) toolloop.Outcome {
	return mt.handleVision(d)(context.Background(), llm.ToolCall{Name: "send_vision", Args: []byte(args)})
}

// miracleCall drives handleMiracle against a dispatch.
func miracleCall(mt *Guardian, d *turnDispatch, args string) toolloop.Outcome {
	return mt.handleMiracle(d)(context.Background(), llm.ToolCall{Name: "work_miracle", Args: []byte(args)})
}

// TestLookFirstGateRefusesUnlookedVision (SC-002, AC#1, T017): on a survival
// turn, a vision at a villager whose pack the guardian has not opened is
// refused as a rejected_gate naming inspect_pack and the villager; nothing
// lands and no charge moves.
func TestLookFirstGateRefusesUnlookedVision(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "behold")
	setInventory(mt, inj, agentIndexByName("Cedar"), cedarsPack)
	before := inj.state.GuardianCharges
	batches := len(inj.batches)

	out := visionCall(mt, packDispatch(mt, true),
		`{"target":"Cedar","text":"eat now, whatever you hold"}`)

	if out.Verdict != toolloop.VerdictRejectedGate {
		t.Fatalf("verdict = %v, want rejected_gate", out.Verdict)
	}
	if !strings.Contains(out.ResultForModel, "inspect_pack") || !strings.Contains(out.ResultForModel, "Cedar") {
		t.Errorf("refusal does not name the repair: %q", out.ResultForModel)
	}
	if len(inj.batches) != batches {
		t.Error("a gated vision landed something")
	}
	if inj.state.GuardianCharges != before {
		t.Error("a gated vision spent a charge")
	}
}

// TestLookFirstGateRepairsWithinTheTurn (SC-002, AC#2, T018): the SAME call,
// after inspect_pack has returned for that villager in the same turn, lands
// exactly as it would have before this feature — the refusal is repairable
// inside the loop's round cap, never a dead end.
func TestLookFirstGateRepairsWithinTheTurn(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "behold")
	setInventory(mt, inj, agentIndexByName("Cedar"), cedarsPack)
	d := packDispatch(mt, true)

	if out := packCall(mt, d, `{"villager":"Cedar"}`); out.Verdict != toolloop.VerdictReadOK {
		t.Fatalf("the repair call itself failed: %v", out)
	}
	out := visionCall(mt, d, `{"target":"Cedar","text":"there is food at the fire"}`)
	if out.Verdict != toolloop.VerdictLanded {
		t.Fatalf("verdict after looking = %v (%s), want landed", out.Verdict, out.ResultForModel)
	}
	if d.result.Nudge == nil {
		t.Error("the repaired vision did not report a nudge")
	}
}

// TestLookFirstGateIsPerVillager (FR-008, T019): looking in one villager's pack
// licenses nothing about another's.
func TestLookFirstGateIsPerVillager(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "behold")
	setInventory(mt, inj, agentIndexByName("Cedar"), cedarsPack)
	d := packDispatch(mt, true)

	packCall(mt, d, `{"villager":"Hazel"}`)
	out := visionCall(mt, d, `{"target":"Cedar","text":"hold on"}`)
	if out.Verdict != toolloop.VerdictRejectedGate {
		t.Fatalf("verdict = %v, want rejected_gate — the gate is per-villager", out.Verdict)
	}
	if !strings.Contains(out.ResultForModel, "Cedar") {
		t.Errorf("refusal names the wrong villager: %q", out.ResultForModel)
	}
}

// TestLookFirstGateOnlyOnSurvivalTurns (FR-008, T020): an ordinary console or
// system turn is ungated — behavior is unchanged for every origin but the
// survival watch (spec 116 A2's deliberately narrow scope).
func TestLookFirstGateOnlyOnSurvivalTurns(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "behold")
	setInventory(mt, inj, agentIndexByName("Cedar"), cedarsPack)

	out := visionCall(mt, packDispatch(mt, false), `{"target":"Cedar","text":"hold on"}`)
	if out.Verdict != toolloop.VerdictLanded {
		t.Fatalf("non-survival vision = %v (%s), want landed", out.Verdict, out.ResultForModel)
	}
}

// TestLookFirstGateCoversPackReachingMiraclesOnly (AC#3, FR-007/FR-008, T016 +
// T021): give_item and take_item are gated exactly as a vision is; a miracle
// that touches no pack (move) and send_omen — which addresses a group, not a
// pack — are never gated.
func TestLookFirstGateCoversPackReachingMiraclesOnly(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "behold")
	cedar := agentIndexByName("Cedar")
	// Two free bulk and a deep bank: this suite is about the GATE, so every
	// call below must be able to reach the door on its own merits (a full pack
	// would have give_item refused by the carry cap — the world-03 lock — and
	// an empty bank refuses everything).
	mt.replica.GuardianCharges, inj.state.GuardianCharges = 10, 10
	setInventory(mt, inj, cedar, sim.Inventory{Wood: 20, Planks: 2})

	for _, kind := range []string{"give_item", "take_item"} {
		d := packDispatch(mt, true)
		args := `{"kind":"` + kind + `","villager":"Cedar","item":"wood","qty":2}`
		out := miracleCall(mt, d, args)
		if out.Verdict != toolloop.VerdictRejectedGate {
			t.Errorf("%s unlooked = %v (%s), want rejected_gate", kind, out.Verdict, out.ResultForModel)
		}
		if !strings.Contains(out.ResultForModel, "inspect_pack") {
			t.Errorf("%s refusal does not name the repair: %q", kind, out.ResultForModel)
		}
		// The repair lands it.
		packCall(mt, d, `{"villager":"Cedar"}`)
		if out := miracleCall(mt, d, args); out.Verdict != toolloop.VerdictLanded {
			t.Errorf("%s after looking = %v (%s), want landed", kind, out.Verdict, out.ResultForModel)
		}
	}

	// A pack-free miracle stays ungated: it is refused (or accepted) purely on
	// its own merits at the door, never by the look-first gate.
	d := packDispatch(mt, true)
	x, y := mt.agentXY[cedar][0], mt.agentXY[cedar][1]
	if out := miracleCall(mt, d, `{"kind":"move","class":"villager","villager":"Cedar","x":0,"y":0,"to_x":`+itoa(x)+`,"to_y":`+itoa(y)+`}`); strings.Contains(out.ResultForModel, "inspect_pack") {
		t.Errorf("a move was look-gated: %q", out.ResultForModel)
	}

	// send_omen addresses a group, not a pack — never gated (contracts §2).
	mt.replica.Night = true
	inj.state.Night = true
	mt.mirrorState()
	d = packDispatch(mt, true)
	d.night = true
	out := mt.handleOmen(d)(context.Background(),
		llm.ToolCall{Name: "send_omen", Args: []byte(`{"targets":"Cedar","text":"the night is long"}`)})
	if out.Verdict != toolloop.VerdictLanded {
		t.Fatalf("survival omen = %v (%s), want landed — omens are never look-gated", out.Verdict, out.ResultForModel)
	}
}

// TestLookFirstGateAuditTrail (FR-014, T034): a whole survival turn — the
// gated attempt, the repair, the landing — leaves the full cog.tool_call
// chain: the read carries a read verdict and lands no world event, and the
// refusal carries its reason, so the skipped look is queryable after the fact.
func TestLookFirstGateAuditTrail(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "I have looked, and I have spoken.")
	cedar := agentIndexByName("Cedar")
	setInventory(mt, inj, cedar, cedarsPack)
	o := seededSurvivalWatch(mt, inj, sim.SurvivalNearDeath)

	mt.runLoop = func(ctx context.Context, j toolloop.Job) (toolloop.Result, error) {
		blind := toolCall("send_vision", `{"target":"Cedar","text":"eat what you hold"}`)
		bo := j.Handlers["send_vision"](ctx, blind)
		j.Record(toolloop.CallRecord{JobID: j.JobID, Ordinal: 1, Tool: "send_vision",
			Args: blind.Args, Verdict: bo.Verdict, Reason: bo.ResultForModel, Tier: "cloud"})

		look := toolCall("inspect_pack", `{"villager":"Cedar"}`)
		lo := j.Handlers["inspect_pack"](ctx, look)
		j.Record(toolloop.CallRecord{JobID: j.JobID, Ordinal: 2, Tool: "inspect_pack",
			Args: look.Args, Verdict: lo.Verdict, Reason: lo.ResultForModel, Tier: "cloud"})

		act := toolCall("send_vision", `{"target":"Cedar","text":"set down the wood; there is food at the fire"}`)
		ao := j.Handlers["send_vision"](ctx, act)
		j.Record(toolloop.CallRecord{JobID: j.JobID, Ordinal: 3, Tool: "send_vision",
			Args: act.Args, Verdict: ao.Verdict, Reason: ao.ResultForModel, Tier: "cloud"})
		if ao.Verdict == toolloop.VerdictLanded {
			return toolloop.Result{Term: toolloop.TermLanded, Landed: &act}, nil
		}
		return toolloop.Result{Term: toolloop.TermCapExhausted}, nil
	}

	mt.runTrigger(triggerJob{order: o, matched: needsEvent(cedar, 24, 0, 800, 5000),
		matchedType: "agent.needs_changed", matchedTick: 5000})

	calls := cogToolCalls(inj)
	if len(calls) != 3 {
		t.Fatalf("recorded %d cog.tool_call payloads, want 3 (blind vision, look, repaired vision)", len(calls))
	}
	if calls[0].Verdict != string(toolloop.VerdictRejectedGate) ||
		!strings.Contains(calls[0].Reason, "inspect_pack") {
		t.Errorf("the blind vision's record does not carry the gate's reason: %+v", calls[0])
	}
	if calls[1].Tool != "inspect_pack" || calls[1].Verdict != string(toolloop.VerdictReadOK) {
		t.Errorf("the look's record = %+v, want inspect_pack/read_ok", calls[1])
	}
	if calls[2].Verdict != string(toolloop.VerdictLanded) {
		t.Errorf("the repaired vision did not land: %+v", calls[2])
	}
	// The read itself landed no world event: the only world batch is the
	// vision's own (the nudge + its memory).
	for _, b := range landedBatches(inj) {
		for _, e := range b {
			if e.Type == "guardian.item_taken" || e.Type == "guardian.item_granted" {
				t.Errorf("a pack READ landed a world mutation: %s", e.Type)
			}
		}
	}
}

// --- US3/US4: the removal lands, and never in silence ---

// TestTakeItemLandsWithItsMemory (FR-013, SC-004, T033): a landed removal rides
// ONE atomic batch — the guardian.item_taken event plus exactly one
// agent.memory_added for the villager reached into — through the same
// InjectSocial door every other act uses, and the goods arrive as a pile at
// their feet rather than vanishing.
func TestTakeItemLandsWithItsMemory(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "behold")
	cedar := agentIndexByName("Cedar")
	setInventory(mt, inj, cedar, cedarsPack)
	x, y := mt.agentXY[cedar][0], mt.agentXY[cedar][1]
	before := inj.state.GuardianCharges

	out := miracleCall(mt, packDispatch(mt, false),
		`{"kind":"take_item","villager":"Cedar","item":"wood","qty":20}`)
	if out.Verdict != toolloop.VerdictLanded {
		t.Fatalf("removal = %v (%s), want landed", out.Verdict, out.ResultForModel)
	}

	batches := landedBatches(inj)
	if len(batches) != 1 {
		t.Fatalf("landed %d world batches, want 1 (the removal + its memory, atomically)", len(batches))
	}
	batch := batches[0]
	if len(batch) != 2 || batch[0].Type != "guardian.item_taken" || batch[1].Type != "agent.memory_added" {
		t.Fatalf("batch shape = %v, want [guardian.item_taken agent.memory_added]", eventTypes(batch))
	}
	var mp sim.MemoryAddedPayload
	if err := json.Unmarshal(batch[1].Payload, &mp); err != nil {
		t.Fatal(err)
	}
	if mp.Agent.ID != cedar {
		t.Errorf("the memory went to villager %d, want %d", mp.Agent.ID, cedar)
	}
	if mp.Salience != sim.SalDream {
		t.Errorf("memory salience = %d, want SalDream (the grant memory's salience)", mp.Salience)
	}
	// The world moved: the pack emptied, the pile appeared, one charge spent.
	if inj.state.Agents[cedar].Inv.Wood != 0 {
		t.Errorf("Cedar still carries %d wood", inj.state.Agents[cedar].Inv.Wood)
	}
	if p := inj.state.Lookup().Pile(x, y); p == nil || p.Wood != 20 {
		t.Errorf("the taken wood is not on the ground at (%d,%d): %+v", x, y, p)
	}
	if inj.state.GuardianCharges != before-1 {
		t.Errorf("removal spent %d charges, want 1", before-inj.state.GuardianCharges)
	}
}

// itoa keeps the sheet assertions free of a strconv import in the table above.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		return "-" + string(b)
	}
	return string(b)
}

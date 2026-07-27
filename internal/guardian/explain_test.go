package guardian

// The explain tool's guardian wiring (spec 063 T004) and the tutor-lane
// neutrality suite (T005, SC-002): the granted-subset handler serves R1
// sheets; a turn may explain across rounds and still land its one act;
// unknown topics return the catalog; an ungranted explain is structurally
// absent from declaration, prose, and door alike; and an explain-bearing
// turn spends nothing and mutates nothing — the initiative frame unchanged
// byte-for-byte.

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/persona"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/tool"
	"github.com/evanstern/promptworld/internal/toolloop"
)

// explainThenActLoop scripts a multi-round turn: N explain reads through the
// REAL handler (each recorded read_ok), then one acting call through its real
// handler — the shape the driver's read exemption permits (reads never
// consume the mediated act; toolloop's own suite pins the driver half).
func explainThenActLoop(mt *Guardian, topics []string, actName, actArgs string) func(context.Context, toolloop.Job) (toolloop.Result, error) {
	return func(ctx context.Context, j toolloop.Job) (toolloop.Result, error) {
		resp, err := bridgeSubmit(mt, ctx, j)
		if err != nil {
			return toolloop.Result{Term: termForErr(err)}, err
		}
		ordinal := 0
		for _, topic := range topics {
			ordinal++
			c := toolCall("explain", fmt.Sprintf(`{"topic":%q}`, topic))
			h, ok := j.Handlers["explain"]
			if !ok {
				return toolloop.Result{Term: toolloop.TermModelDone}, nil
			}
			out := h(ctx, c)
			j.Record(toolloop.CallRecord{JobID: j.JobID, Ordinal: ordinal, Tool: "explain",
				Args: c.Args, Verdict: out.Verdict, Reason: "", Tier: "cloud"})
		}
		if actName == "" {
			return toolloop.Result{Final: resp.Text, Term: toolloop.TermModelDone}, nil
		}
		ordinal++
		c := toolCall(actName, actArgs)
		out := j.Handlers[actName](ctx, c)
		j.Record(toolloop.CallRecord{JobID: j.JobID, Ordinal: ordinal, Tool: actName,
			Args: c.Args, Verdict: out.Verdict, Reason: out.ResultForModel, Tier: "cloud"})
		if out.Verdict == toolloop.VerdictLanded {
			return toolloop.Result{Final: resp.Text, Term: toolloop.TermLanded, Landed: &c}, nil
		}
		return toolloop.Result{Final: resp.Text, Term: toolloop.TermCapExhausted}, nil
	}
}

// TestExplainHandlerServesFactSheets (T004): with explain granted (the
// default full grant), the handler returns the tool-side R1 sheet for the
// asked topic — read_ok, byte-equal to tool.ExplainSheet under the same
// scope — and the unknown topic returns the honest catalog.
func TestExplainHandlerServesFactSheets(t *testing.T) {
	mt, _, _, _ := newTestGuardian(t, "so it is written")
	d := &turnDispatch{mt: mt, charges: 3, alive: map[int]bool{0: true}, tick: 1, result: &TurnResult{},
		grant: fullGrant(),
		scope: tool.ExplainScope{Granted: tool.LoopRosterGuardian(), Catalog: tool.LoopRosterGuardian()}}
	h := mt.turnHandlers(d)
	eh, ok := h["explain"]
	if !ok {
		t.Fatal("full grant installs no explain handler")
	}

	out := eh(context.Background(), toolCall("explain", `{"topic":"charges"}`))
	if out.Verdict != toolloop.VerdictReadOK {
		t.Fatalf("explain verdict = %s, want read_ok", out.Verdict)
	}
	if want := tool.ExplainSheet("charges", d.scope); out.ResultForModel != want {
		t.Errorf("handler sheet diverges from the tool-side composition:\ngot  %q\nwant %q", out.ResultForModel, want)
	}

	miss := eh(context.Background(), toolCall("explain", `{"topic":"weather"}`))
	if miss.Verdict != toolloop.VerdictReadOK || !strings.Contains(miss.ResultForModel, "Topics I can explain") {
		t.Errorf("unknown topic should return the catalog as read_ok, got (%s, %q)", miss.Verdict, miss.ResultForModel)
	}
}

// TestExplainThenActTurn (T004): a multi-round turn that explains twice and
// then sends a vision lands the vision — reads never block or consume the
// act — and every call is recorded as cog.tool_call telemetry.
func TestExplainThenActTurn(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "behold")
	mt.runLoop = explainThenActLoop(mt, []string{"roster", "costs"},
		"send_vision", `{"target":"`+sim.AgentNames[0]+`","text":"tend the fire"}`)

	res, err := mt.Turn(context.Background(), "what can you do? then warn them")
	if err != nil {
		t.Fatal(err)
	}
	if res.Nudge == nil {
		t.Fatal("the acting call after the explains did not land")
	}
	calls := cogToolCalls(inj)
	if len(calls) != 3 {
		t.Fatalf("recorded %d cog.tool_call payloads, want 3 (2 explains + 1 act)", len(calls))
	}
	for i, want := range []struct{ tool, verdict string }{
		{"explain", "read_ok"}, {"explain", "read_ok"}, {"send_vision", "landed"},
	} {
		if calls[i].Tool != want.tool || calls[i].Verdict != want.verdict {
			t.Errorf("call %d = (%s, %s), want (%s, %s)", i, calls[i].Tool, calls[i].Verdict, want.tool, want.verdict)
		}
	}
}

// TestExplainStructurallyAbsentWhenUngranted (T003/US1 AS-4): a manifest
// omitting explain removes it from all three layers at once — the declared
// roster, the derived guidance prose, and the handler set (the door).
func TestExplainStructurallyAbsentWhenUngranted(t *testing.T) {
	mt, _, _, dir := newTestGuardian(t, "ok")
	if err := os.WriteFile(filepath.Join(dir, "capabilities.json"),
		[]byte(`{"tools":["send_vision"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	grant, notices := loadManifest(dir)
	if len(notices) != 0 {
		t.Fatalf("unexpected notices: %v", notices)
	}
	if grant.allows("explain") {
		t.Error("grant allows explain despite the manifest omitting it")
	}
	roster := grantedRoster(grant)
	for _, tl := range roster {
		if tl.Name == "explain" {
			t.Error("declared roster carries the ungranted explain")
		}
	}
	if g := tool.GuardianReadGuidance(roster); g != "" {
		t.Errorf("read guidance should be empty without explain, got %q", g)
	}
	d := &turnDispatch{mt: mt, charges: 3, alive: map[int]bool{}, tick: 1, result: &TurnResult{}, grant: grant}
	if _, ok := mt.turnHandlers(d)["explain"]; ok {
		t.Error("handler installed for the ungranted explain")
	}
}

// TestExplainTurnNeutrality (T005, SC-002): a turn that only explains leaves
// the charge bank untouched and lands NO world-mutating event — the only
// door traffic is the standard cog.tool_call telemetry — and the turn's
// reply still carries the converse text (explaining is speech, not an act).
func TestExplainTurnNeutrality(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "a vision costs one charge")
	before := inj.state.GuardianCharges
	mt.runLoop = explainThenActLoop(mt, []string{"charges", "workings", "glyphs"}, "", "")

	res, err := mt.Turn(context.Background(), "what does a vision cost?")
	if err != nil {
		t.Fatal(err)
	}
	if res.Reply != "a vision costs one charge" {
		t.Errorf("reply = %q, want the converse text", res.Reply)
	}
	if inj.state.GuardianCharges != before {
		t.Errorf("charge bank moved: %d -> %d", before, inj.state.GuardianCharges)
	}
	if lb := landedBatches(inj); len(lb) != 0 {
		t.Errorf("explain-only turn landed world batches: %v", lb)
	}
	for _, c := range cogToolCalls(inj) {
		if c.Tool != "explain" || c.Verdict != "read_ok" {
			t.Errorf("unexpected telemetry (%s, %s)", c.Tool, c.Verdict)
		}
	}
}

// TestExplainLeavesInitiativeFrameBytes (T005, SC-002's zero initiative-frame
// diffs): granting the READ tools (explain; survey_site since spec 084)
// changes the read paragraph and nothing about the frame — the
// non-negotiables and the initiative frame appear VERBATIM in both
// compositions, and stripping the read paragraph from the granted
// composition yields the read-free one byte-for-byte.
func TestExplainLeavesInitiativeFrameBytes(t *testing.T) {
	full := tool.LoopRosterGuardian()
	var without []tool.Tool
	for _, tl := range full {
		if tl.Effect != tool.Read {
			without = append(without, tl)
		}
	}
	withExplain := turnSystemPrompt(persona.DefaultCharter, nil, full)
	withoutExplain := turnSystemPrompt(persona.DefaultCharter, nil, without)

	for name, prompt := range map[string]string{"with": withExplain, "without": withoutExplain} {
		if !strings.Contains(prompt, guardianNonNegotiables) {
			t.Errorf("%s explain: non-negotiables not verbatim", name)
		}
		if !strings.Contains(prompt, guardianInitiativeFrame) {
			t.Errorf("%s explain: initiative frame not verbatim", name)
		}
	}
	read := tool.GuardianReadGuidance(full)
	if read == "" {
		t.Fatal("full roster renders no read guidance")
	}
	if got := strings.Replace(withExplain, read+"\n", "", 1); got != withoutExplain {
		t.Error("the explain grant changes prompt bytes beyond the read paragraph")
	}
}

// TestExplainScopeMatchesTurnRoster (T004): the dispatch scope runTurn builds
// carries the SAME roster the turn declares — a kind-restricted manifest
// reaches the fact sheets exactly as it reaches the declaration.
func TestExplainScopeMatchesTurnRoster(t *testing.T) {
	mt, _, _, dir := newTestGuardian(t, "ok")
	if err := os.WriteFile(filepath.Join(dir, "capabilities.json"),
		[]byte(`{"tools":["work_miracle","explain"],"miracle_kinds":["give_item"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	ran := false
	mt.runLoop = func(ctx context.Context, j toolloop.Job) (toolloop.Result, error) {
		ran = true
		c := toolCall("explain", `{"topic":"workings"}`)
		out := j.Handlers["explain"](ctx, c)
		if !strings.Contains(out.ResultForModel, `"give_item" with `) {
			t.Errorf("workings sheet omits the granted kind: %q", out.ResultForModel)
		}
		if !strings.Contains(out.ResultForModel, "time_snap") ||
			!strings.Contains(out.ResultForModel, "Cataloged kinds not granted in this world:") {
			t.Errorf("workings sheet fails to name ungranted kinds: %q", out.ResultForModel)
		}
		if strings.Contains(out.ResultForModel, `"time_snap" with `) {
			t.Errorf("workings sheet offers an ungranted kind: %q", out.ResultForModel)
		}
		return toolloop.Result{Final: "done", Term: toolloop.TermModelDone}, nil
	}
	if _, err := mt.Turn(context.Background(), "what workings can you do?"); err != nil {
		t.Fatal(err)
	}
	if !ran {
		t.Fatal("the scripted loop never ran")
	}
}

// TestExplainSheetIsToolSide (standing resolution 3): the recorded telemetry
// proves the result the model saw was the tool-side composition — the args
// carry only the topic, and the sheet equals tool.ExplainSheet for it.
func TestExplainSheetIsToolSide(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "ok")
	mt.runLoop = explainThenActLoop(mt, []string{"decisions"}, "", "")
	if _, err := mt.Turn(context.Background(), "how are calls judged?"); err != nil {
		t.Fatal(err)
	}
	calls := cogToolCalls(inj)
	if len(calls) != 1 {
		t.Fatalf("want 1 recorded explain call, got %d", len(calls))
	}
	var args struct {
		Topic string `json:"topic"`
	}
	if err := json.Unmarshal(calls[0].Args, &args); err != nil || args.Topic != "decisions" {
		t.Errorf("recorded args = %s, want the topic only", calls[0].Args)
	}
}

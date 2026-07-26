package tool

// Explain ground-truth sweep (spec 063 T002, SC-001): every fact sheet's
// content equals its registry/doctrine ground truth across the whole topic
// catalog, sheets are byte-identical across calls (zero model bytes in the
// answer path, standing resolution 3), scoping respects the effective grant
// (an ungranted tool is named existing-but-ungranted, never hidden and never
// offered), and an unknown topic returns the honest catalog.

import (
	"fmt"
	"strings"
	"testing"
)

// fullScope is the unrestricted default-grant scope: everything the loop
// declares, granted.
func fullScope() ExplainScope {
	return ExplainScope{Granted: LoopRosterGuardian(), Catalog: LoopRosterGuardian()}
}

// TestExplainGroundTruthSweep walks the whole topic catalog (the six fixed
// topics plus every cataloged tool-id topic) and asserts each sheet against
// the registry/doctrine source it derives from (SC-001).
func TestExplainGroundTruthSweep(t *testing.T) {
	sc := fullScope()

	// roster: exactly the granted names, each with its guidance gloss.
	roster := ExplainSheet("roster", sc)
	for _, tl := range sc.Granted {
		if !strings.Contains(roster, tl.Name+"("+paramNameList(tl)+")") {
			t.Errorf("roster sheet omits granted tool surface %q(%s)", tl.Name, paramNameList(tl))
		}
	}
	if strings.Contains(roster, "Cataloged but not granted") {
		t.Error("full-grant roster sheet claims ungranted tools exist")
	}

	// costs: every granted tool's price comes from the registry.
	costs := ExplainSheet("costs", sc)
	for _, tl := range sc.Granted {
		if tl.Name == "work_miracle" {
			for _, k := range MiracleKinds() {
				cost, _ := MiracleCost(k)
				want := fmt.Sprintf("%q — %d %s", k, cost, chargeWord(cost))
				if !strings.Contains(costs, want) {
					t.Errorf("costs sheet omits authoritative price %q", want)
				}
			}
			continue
		}
		if !strings.Contains(costs, "- "+tl.Name+" — ") {
			t.Errorf("costs sheet omits tool %q", tl.Name)
		}
	}

	// charges: the mirrored doctrine constants, verbatim.
	charges := ExplainSheet("charges", sc)
	cap, genesis, regen := ChargeDoctrine()
	for _, want := range []string{
		fmt.Sprintf("at most %d charges", cap),
		fmt.Sprintf("begins with %d charge", genesis),
		fmt.Sprintf("every %d game hours", regen),
	} {
		if !strings.Contains(charges, want) {
			t.Errorf("charges sheet omits doctrine fact %q", want)
		}
	}

	// workings: every kind with its authoritative price and argument hint.
	workings := ExplainSheet("workings", sc)
	for _, k := range MiracleKinds() {
		cost, _ := MiracleCost(k)
		want := fmt.Sprintf("%q with %s — %d %s", k, miracleKindArgs[k], cost, chargeWord(cost))
		if !strings.Contains(workings, want) {
			t.Errorf("workings sheet omits %q", want)
		}
	}

	// decisions: every mirrored verdict name.
	decisions := ExplainSheet("decisions", sc)
	for _, name := range DecisionClassNames() {
		if !strings.Contains(decisions, "- "+name+" — ") {
			t.Errorf("decisions sheet omits verdict %q", name)
		}
	}

	// glyphs: every legend row and the agent note.
	glyphs := ExplainSheet("glyphs", sc)
	rows, note := MapGlyphLegend()
	for _, r := range rows {
		want := fmt.Sprintf("- %s %s — %s", r[0], r[1], r[2])
		if !strings.Contains(glyphs, want) {
			t.Errorf("glyphs sheet omits legend row %q", want)
		}
	}
	if !strings.Contains(glyphs, note) {
		t.Error("glyphs sheet omits the agent-glyph note")
	}

	// <tool-id> detail: every cataloged tool has a sheet naming its grant
	// status and argument surface.
	for _, tl := range sc.Catalog {
		d := ExplainSheet(tl.Name, sc)
		if !strings.HasPrefix(d, tl.Name+" — granted in this world") {
			t.Errorf("detail sheet for %q lacks the granted header: %q", tl.Name, firstLine(d))
		}
		if args := paramNameList(tl); args != "" && !strings.Contains(d, "- arguments: "+args) {
			t.Errorf("detail sheet for %q omits its argument surface", tl.Name)
		}
	}
}

// TestExplainByteStability (US1 AS-2): identical (state, topic) → byte-
// identical answers — no model in the answer path.
func TestExplainByteStability(t *testing.T) {
	sc := fullScope()
	topics := append(ExplainTopics(), "", "work_miracle", "no-such-topic")
	for _, topic := range topics {
		a, b := ExplainSheet(topic, sc), ExplainSheet(topic, fullScope())
		if a != b {
			t.Errorf("topic %q is not byte-stable across calls", topic)
		}
	}
}

// TestExplainUnknownTopicCatalog (US1 AS-3, FR-002): a topic that doesn't
// exist returns the honest topic catalog — a repairable miss, never a
// fabricated answer — and an empty topic returns the catalog too.
func TestExplainUnknownTopicCatalog(t *testing.T) {
	sc := fullScope()
	miss := ExplainSheet("weather", sc)
	if !strings.Contains(miss, `"weather" is not a topic I can explain`) {
		t.Errorf("unknown-topic sheet lacks the honest miss line: %q", firstLine(miss))
	}
	for _, topic := range ExplainTopics() {
		if !strings.Contains(miss, topic) {
			t.Errorf("miss catalog omits topic %q", topic)
		}
	}
	empty := ExplainSheet("", sc)
	if !strings.HasPrefix(empty, "Topics I can explain: ") {
		t.Errorf("empty-topic sheet is not the catalog: %q", firstLine(empty))
	}
	if !strings.Contains(empty, "work_miracle") {
		t.Error("catalog omits the tool-id topics")
	}
}

// TestExplainGrantScoping (US1 AS-1, edge case 1): the sheet reflects the
// world's EFFECTIVE grant — an ungranted cataloged tool is named as
// existing-but-ungranted (never hidden, never offered), and a restricted
// work_miracle explains only its granted kinds as granted.
func TestExplainGrantScoping(t *testing.T) {
	// A dreams-only world: send_vision granted, everything else cataloged.
	sv, _ := Lookup("send_vision")
	sc := ExplainScope{Granted: []Tool{sv}, Catalog: LoopRosterGuardian(), Stage: "stage-1"}

	roster := ExplainSheet("roster", sc)
	if !strings.Contains(roster, "send_vision(") {
		t.Error("roster sheet omits the one granted tool")
	}
	if !strings.Contains(roster, "Cataloged but not granted in this world:") ||
		!strings.Contains(roster, "work_miracle") {
		t.Error("roster sheet fails to name ungranted cataloged tools")
	}
	if !strings.Contains(roster, "This world's stage: stage-1") {
		t.Error("roster sheet omits the stage ceiling line")
	}

	// Ungranted work_miracle: the detail and workings sheets say so honestly.
	detail := ExplainSheet("work_miracle", sc)
	if !strings.Contains(detail, "NOT granted in this world") {
		t.Errorf("ungranted tool detail lacks the honest grant line: %q", firstLine(detail))
	}
	workings := ExplainSheet("workings", sc)
	if !strings.Contains(workings, "NOT granted in this world") {
		t.Error("workings sheet on a miracle-less world lacks the honest grant line")
	}

	// A kind-restricted grant: only granted kinds appear as offered; the
	// remainder is named as cataloged-but-ungranted.
	wm, _ := Lookup("work_miracle")
	restricted := RestrictEnum(wm, "kind", []string{"give_item"})
	sc2 := ExplainScope{Granted: []Tool{restricted}, Catalog: LoopRosterGuardian()}
	w2 := ExplainSheet("workings", sc2)
	if !strings.Contains(w2, `"give_item" with `) {
		t.Error("restricted workings sheet omits the granted kind")
	}
	if !strings.Contains(w2, "Cataloged kinds not granted in this world:") ||
		!strings.Contains(w2, "time_snap") {
		t.Error("restricted workings sheet fails to name ungranted kinds")
	}
	if strings.Contains(w2, `"time_snap" with `) {
		t.Error("restricted workings sheet offers an ungranted kind")
	}
}

// TestExplainToolDeclaration (T003): the registry entry is the read-only
// class — Effect Read, no events, no charge, gate None, topic optional Text
// (an unknown topic must reach the handler for the catalog result, so the
// schema layer must not gate it as an enum).
func TestExplainToolDeclaration(t *testing.T) {
	ex, ok := Lookup("explain")
	if !ok {
		t.Fatal("explain not registered")
	}
	if ex.Effect != Read {
		t.Errorf("explain.Effect = %v, want Read", ex.Effect)
	}
	if len(ex.Events) != 0 {
		t.Errorf("explain declares events %v, want none (tutor-lane by construction)", ex.Events)
	}
	if ex.Cost.Charges != 0 {
		t.Errorf("explain.Cost.Charges = %d, want 0", ex.Cost.Charges)
	}
	if ex.Gate != None {
		t.Errorf("explain.Gate = %v, want None", ex.Gate)
	}
	if len(ex.Params) != 1 || ex.Params[0].Name != "topic" || ex.Params[0].Required || ex.Params[0].Kind != Text {
		t.Errorf("explain params = %+v, want one optional Text topic", ex.Params)
	}
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

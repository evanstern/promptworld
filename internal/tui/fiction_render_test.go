package tui

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/guardian"
	"github.com/evanstern/promptworld/internal/store"
)

var (
	renderDenylistRe = regexp.MustCompile(`(?i)\b(metatron|angels?|miracles?|divine|heavens?|scriptures?)\b`)
	ansiRe           = regexp.MustCompile(`\x1b\[[0-9;]*m`)
)

func assertRenderClean(t *testing.T, surface, out string) {
	t.Helper()
	plain := ansiRe.ReplaceAllString(out, "")
	for _, m := range renderDenylistRe.FindAllString(plain, -1) {
		t.Errorf("%s: fiction vocabulary %q rendered in the default skin:\n%s", surface, m, plain)
	}
}

// TestDefaultSkinRendersFictionFree (spec 052 T008/SC-001, the TUI half):
// the default experience's rendered surfaces — tabs, dock row, footers,
// transcript body, minibuffer states, the help overlay's every page, the
// standing-orders block, and the chronicle lines for every guardian-family
// event — carry no guardian-fiction vocabulary. (The raw detail pane and the
// grammar-miss fallback are inspector surfaces and deliberately exempt —
// FR-020 audience ruling.)
func TestDefaultSkinRendersFictionFree(t *testing.T) {
	m := Model{width: 130, height: 40}
	m.transcript = []string{"you: hello", transcriptGuardianPrefix + "all is well"}
	m.mbBusy = true

	assertRenderClean(t, "tabsView", m.tabsView())
	assertRenderClean(t, "dockTabsRow", m.dockTabsRow())
	assertRenderClean(t, "soloTitle", func() string { m2 := m; m2.solo = true; return m2.soloTitle() }())
	assertRenderClean(t, "guardianTranscriptBody", m.guardianTranscriptBody(60, 12))
	for _, w := range []int{40, 120} {
		m2 := m
		m2.width = w
		assertRenderClean(t, "footerView", m2.footerView())
	}
	// Minibuffer states: focused, busy, flash, dormant.
	states := []Model{
		{mbFocused: true, mbInput: "hello"},
		{mbBusy: true},
		{mbFlash: "answer arrived — 3 to read"},
		{},
	}
	for _, s := range states {
		assertRenderClean(t, "minibufferView", s.minibufferView(80))
	}
	// Help overlay: every section × mode × tier.
	for mode := helpModeKey(0); mode < helpModeCount; mode++ {
		for sec := helpSection(0); sec < helpSectionCount; sec++ {
			for _, tier := range []bool{false, true} {
				h := m
				h.helpOpen, h.helpMode, h.helpPageMode, h.helpSection, h.helpTier = true, mode, mode, sec, tier
				assertRenderClean(t, "helpPanelView", h.helpPanelView(120, 34))
			}
		}
	}
	// Standing orders + tool summary.
	assertRenderClean(t, "orderStatusLines", strings.Join(orderStatusLines([]guardian.OrderStatus{
		{ID: "ord-1-0", Condition: "the fire goes out", Origin: "player", ExpiresDay: 3, Status: "active"},
	}), "\n"))
	assertRenderClean(t, "consoleToolsSummary", consoleToolsSummary(&guardian.Status{
		GrantedTools: []string{"send_vision", "work_miracle"},
	}, nil))

	// Chronicle: the guardian-family digest lines, solo-rendered (aliased
	// Type column + skin-name subject).
	fixtures := map[string]string{
		"metatron.nudged":             `{"form":"vision","targets":[0],"text":"beware"}`,
		"metatron.charge_regenerated": `{}`,
		"metatron.order_placed":       `{"id":"ord-1-0","origin":"player","condition":"c","action":"a","event_types":["agent.died"],"placed_tick":1,"expires_tick":90000,"status":"active"}`,
		"metatron.order_triggered":    `{"id":"ord-1-0","matched_type":"agent.died","matched_tick":5}`,
		"metatron.order_cancelled":    `{"id":"ord-1-0"}`,
		"metatron.order_expired":      `{"id":"ord-1-0"}`,
		"metatron.charter_observed":   `{"fingerprint":"ab12cd34ef56","default":true}`,
		"metatron.time_snapped":       `{"to_tick":106200,"gratis":false}`,
		"metatron.item_granted":       `{"agent":0,"kind":"food_raw","qty":2,"gratis":false}`,
		"metatron.entity_moved":       `{"class":"pile","x":3,"y":4,"to_x":6,"to_y":7,"gratis":false}`,
		"metatron.entity_removed":     `{"class":"pile","x":3,"y":4,"gratis":false}`,
		"metatron.place_revealed":     `{"agent":0,"facts":[{"kind":"fire","x":4,"y":5,"provenance":"revealed"}]}`,
		"curriculum.stage_unlocked":   `{"stage":"stage-2","exercise":"first-night"}`,
		"curriculum.exercise_passed":  `{"stage":"stage-1","exercise":"first-night"}`,
	}
	names := []string{"Ash"}
	for typ, payload := range fixtures {
		e := store.Event{Seq: 1, Tick: 100, Type: typ, Payload: json.RawMessage(payload)}
		l := formatChronicleLine(e, names, nil)
		cols := computeChronicleColumns([]chronicleLine{l}, false)
		assertRenderClean(t, "chronicle "+typ, renderChronicleRow(l, cols, 120, 1, false))
	}
}

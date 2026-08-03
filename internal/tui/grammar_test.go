package tui

import (
	"encoding/json"
	"github.com/muesli/termenv"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/store"
)

var testNames = []string{"Ash", "Birch", "Cedar", "Rowan"}

// TestEventFamilyOf is the digest-grammar contract's family derivation (R2):
// namespace prefix, with meeting/norm merged into one governance family.
func TestEventFamilyOf(t *testing.T) {
	cases := []struct {
		eventType string
		want      eventFamily
	}{
		{"world.created", familyWorld},
		{"clock.paused", familyClock},
		{"sim.day_started", familySim},
		{"agent.moved", familyAgent},
		{"social.gave", familySocial},
		{"meeting.opened", familyGovernance},
		{"norm.violated", familyGovernance},
		{"gru.emerged", familyGru},
		{"chronicle.entry", familyChronicle},
		{"guardian.nudged", familyGuardian},
		{"daemon.started", familyDaemon},
		{"cog.thought", familyCog},
		{"future.unknown_type", familyUnknown}, // new namespaces land here until promoted
		{"no-dot-at-all", familyUnknown},
	}
	for _, c := range cases {
		if got := eventFamilyOf(c.eventType); got != c.want {
			t.Errorf("eventFamilyOf(%q) = %v, want %v", c.eventType, got, c.want)
		}
	}
}

// TestFormatChronicleLineFallback: an unregistered type (or one whose
// payload fails to unmarshal against its registered digestFunc) falls back
// to the pre-digest compact resolved-name JSON as one segText span — never
// blank, never a panic (FR-002/FR-003).
func TestFormatChronicleLineFallback(t *testing.T) {
	e := store.Event{Seq: 1, Tick: 60, Type: "future.unknown_type",
		Payload: json.RawMessage(`{"agent":0,"x":1,"y":1}`)}
	l := formatChronicleLine(e, testNames, nil)
	if l.Family != familyUnknown {
		t.Errorf("family = %v, want familyUnknown", l.Family)
	}
	if len(l.Summary) != 1 || l.Summary[0].Role != segText {
		t.Fatalf("fallback summary should be exactly one segText span: %+v", l.Summary)
	}
	if want := `{"agent":"Ash","x":1,"y":1}`; l.Summary[0].Text != want {
		t.Errorf("fallback text = %q, want %q", l.Summary[0].Text, want)
	}
}

// TestFormatChronicleLineFallbackOnUnmarshalFailure: a registered type whose
// payload doesn't unmarshal (digestFunc returns ok=false) falls back exactly
// like a registry miss.
func TestFormatChronicleLineFallbackOnUnmarshalFailure(t *testing.T) {
	e := store.Event{Seq: 1, Tick: 1, Type: "agent.moved", Payload: json.RawMessage(`not json`)}
	l := formatChronicleLine(e, testNames, nil)
	if len(l.Summary) != 1 || l.Summary[0].Role != segText {
		t.Fatalf("unmarshal failure should fall back to one segText span: %+v", l.Summary)
	}
}

// TestFormatChronicleLineSpeechPrivilege: social.conversation_turn and
// social.rumor_told carry the speech privilege (segSpeech role on the
// quoted utterance) per contract §3.
func TestFormatChronicleLineSpeechPrivilege(t *testing.T) {
	e := store.Event{Seq: 1201, Tick: 8846, Type: "social.conversation_turn",
		Payload: json.RawMessage(`{"conv":102,"speaker":3,"listener":0,"text":"I stacked wood at dawn"}`)}
	l := formatChronicleLine(e, testNames, nil)
	if want := `Rowan→Ash "I stacked wood at dawn"`; plainSegs(l.Summary) != want {
		t.Errorf("plain summary = %q, want %q", plainSegs(l.Summary), want)
	}
	if !anyRole(l.Summary, segSpeech) {
		t.Error("conversation_turn summary should carry a segSpeech span")
	}
	if !anyRole(l.Summary, segName) {
		t.Error("conversation_turn summary should carry segName spans for both agents")
	}

	rumor := store.Event{Seq: 1203, Tick: 8900, Type: "social.rumor_told",
		Payload: json.RawMessage(`{"from":1,"to":2,"rumor_id":0,"subject":0,"tone":30,"text":"ash lets the fire die","confidence":40}`)}
	l2 := formatChronicleLine(rumor, testNames, nil)
	if want := `Birch→Cedar rumor: "ash lets the fire die"`; plainSegs(l2.Summary) != want {
		t.Errorf("rumor plain summary = %q, want %q", plainSegs(l2.Summary), want)
	}
	if !anyRole(l2.Summary, segSpeech) {
		t.Error("rumor_told summary should carry a segSpeech span")
	}
}

// TestFormatChronicleLineOutOfRangeIndex: an index beyond the roster
// renders as "#N" rather than panicking or silently dropping the field —
// exercised here through a registered digest (agent.moved).
func TestFormatChronicleLineOutOfRangeIndex(t *testing.T) {
	e := store.Event{Seq: 1, Tick: 1, Type: "agent.moved",
		Payload: json.RawMessage(`{"agent":99,"x":3,"y":4}`)}
	l := formatChronicleLine(e, testNames, nil)
	if want := `#99 → (3,4)`; plainSegs(l.Summary) != want {
		t.Errorf("summary = %q, want %q", plainSegs(l.Summary), want)
	}
}

// TestFormatChronicleLineHail: the hail family (registered digests) resolves
// agent indices to names in the contract's natural-phrase voice.
func TestFormatChronicleLineHail(t *testing.T) {
	cases := []struct {
		eventType string
		payload   string
		want      string
	}{
		{"social.hailed", `{"from":1,"to":3,"until":12345}`, `Birch hailed Rowan (until t12345)`},
		{"social.hail_met", `{"from":1,"to":3}`, `Birch met Rowan`},
		{"social.hail_expired", `{"from":0,"to":2}`, `Ash's hail to Cedar lapsed`},
	}
	for _, c := range cases {
		e := store.Event{Seq: 1, Tick: 1, Type: c.eventType, Payload: json.RawMessage(c.payload)}
		l := formatChronicleLine(e, testNames, nil)
		if got := plainSegs(l.Summary); got != c.want {
			t.Errorf("%s summary = %q, want %q", c.eventType, got, c.want)
		}
	}
}

// TestFormatChronicleLineStorage (T034, migrated to the digest registry):
// the storage-family digests resolve owner/taker to names, matching the
// contract's natural-phrase templates.
func TestFormatChronicleLineStorage(t *testing.T) {
	cases := []struct {
		eventType string
		payload   string
		want      string
	}{
		{"agent.dropped", `{"agent":0,"x":3,"y":4,"kind":"wood","n":2}`, `Ash dropped 2 wood at (3,4)`},
		{"agent.picked_up", `{"agent":1,"x":3,"y":4,"kind":"wood","n":2}`, `Birch picked up 2 wood at (3,4)`},
		{"agent.deposited", `{"agent":2,"x":5,"y":5,"kind":"planks","n":6}`, `Cedar stored 6 planks in the chest at (5,5)`},
		{"agent.withdrew", `{"agent":3,"x":5,"y":5,"kind":"planks","n":1,"owner":0}`, `Rowan took 1 planks from Ash's chest`},
		{"agent.withdrew", `{"agent":0,"x":5,"y":5,"kind":"planks","n":1,"owner":0}`, `Ash took 1 planks from their chest`},
		{"social.chest_taken", `{"owner":0,"taker":3,"x":5,"y":5}`, `Rowan raided Ash's chest at (5,5)`},
		{"sim.food_rotted", `{"x":6,"y":6,"kind":"food_raw","n":4}`, `4 food_raw rotted at (6,6)`},
	}
	for _, c := range cases {
		e := store.Event{Seq: 1, Tick: 1, Type: c.eventType, Payload: json.RawMessage(c.payload)}
		l := formatChronicleLine(e, testNames, nil)
		if got := plainSegs(l.Summary); got != c.want {
			t.Errorf("%s summary = %q, want %q", c.eventType, got, c.want)
		}
	}
}

// TestResolvePayloadNames: order preserved, unrelated fields untouched,
// out-of-range indices fall back to the raw match. (fallback path, unchanged)
func TestResolvePayloadNames(t *testing.T) {
	raw := json.RawMessage(`{"agent":2,"x":7,"y":8}`)
	got := resolvePayloadNames(raw, testNames)
	if want := `{"agent":"Cedar","x":7,"y":8}`; got != want {
		t.Errorf("resolvePayloadNames = %q, want %q", got, want)
	}

	oob := json.RawMessage(`{"agent":99,"x":1,"y":1}`)
	got2 := resolvePayloadNames(oob, testNames)
	if want := `{"agent":99,"x":1,"y":1}`; got2 != want {
		t.Errorf("out-of-range index should pass through unchanged: got %q, want %q", got2, want)
	}
}

// TestFormatInspector: the stored event verbatim, seq/tick/type intact,
// integer indices intact, with a trailing "// name" annotation — never a
// payload rewrite (chronicle-grammar.md "Inspector", contract §5). Unchanged
// by the digest grammar — the inspector's contract is untouched.
func TestFormatInspector(t *testing.T) {
	e := store.Event{Seq: 1202, Tick: 8846, Type: "social.conversation_turn",
		Payload: json.RawMessage(`{"conv":102,"speaker":1,"listener":0,"text":"I stacked wood at dawn, ask Birch"}`)}
	got := formatInspector(e, testNames)

	for _, want := range []string{
		`"seq": 1202`, `"tick": 8846`,
		`"type": "social.conversation_turn"`,
		`"conv": 102`,
		`"speaker": 1`, `// Birch`,
		`"listener": 0`, `// Ash`,
		`"text": "I stacked wood at dawn, ask Birch"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("inspector missing %q in:\n%s", want, got)
		}
	}
	// Field order preserved: speaker before listener before text.
	if strings.Index(got, `"speaker"`) > strings.Index(got, `"listener"`) {
		t.Error("field order not preserved")
	}
}

// TestFormatInspectorNoNamesAvailable: an empty roster (disconnected) must
// not panic and must still show the raw indices.
func TestFormatInspectorNoNamesAvailable(t *testing.T) {
	e := store.Event{Seq: 1, Tick: 1, Type: "social.conversation_turn",
		Payload: json.RawMessage(`{"speaker":1,"listener":0,"text":"hi"}`)}
	got := formatInspector(e, nil)
	if !strings.Contains(got, `"speaker": 1`) {
		t.Errorf("inspector with no names should still show raw indices: %q", got)
	}
	if strings.Contains(got, "// ") {
		t.Errorf("no annotation should be added when names are unavailable: %q", got)
	}
}

// TestPlainChronicleLine: solo shows the right-aligned tick column; dock
// drops it (contract §1).
func TestPlainChronicleLine(t *testing.T) {
	l := chronicleLine{Tick: 42, Time: "08:11", Type: "agent.moved", Summary: []seg{txt("Ash → (1,1)")}}

	solo := computeChronicleColumns([]chronicleLine{l}, false)
	if got, want := plainChronicleLine(l, solo), `42 08:11  agent.moved  Ash → (1,1)`; got != want {
		t.Errorf("plainChronicleLine(solo) = %q, want %q", got, want)
	}

	dock := computeChronicleColumns([]chronicleLine{l}, true)
	if got, want := plainChronicleLine(l, dock), `08:11 moved  Ash → (1,1)`; got != want {
		t.Errorf("plainChronicleLine(dock) = %q, want %q", got, want)
	}
}

// TestComputeChronicleColumnsAlignment: tick right-aligned to the widest
// visible tick; type padded to the widest visible type, capped (R5).
func TestComputeChronicleColumnsAlignment(t *testing.T) {
	lines := []chronicleLine{
		{Tick: 5, Time: "06:00", Type: "clock.paused", Summary: []seg{txt("paused")}},
		{Tick: 12345, Time: "08:11", Type: "social.conversation_turn", Summary: []seg{txt("x")}},
	}
	cols := computeChronicleColumns(lines, false)
	if cols.TickWidth != 5 {
		t.Errorf("TickWidth = %d, want 5 (len of the widest tick, %q)", cols.TickWidth, "12345")
	}
	if cols.TypeWidth != len("social.conversation_turn") {
		t.Errorf("TypeWidth = %d, want %d", cols.TypeWidth, len("social.conversation_turn"))
	}

	// A pathologically long type name truncates at the cap rather than
	// blowing the column out past 26 (solo).
	longType := "future.a_very_long_event_type_name_indeed"
	capped := computeChronicleColumns([]chronicleLine{{Tick: 1, Type: longType}}, false)
	if capped.TypeWidth != typeColumnCapSolo {
		t.Errorf("TypeWidth = %d, want cap %d", capped.TypeWidth, typeColumnCapSolo)
	}
	padded := padType(chronicleLine{Type: longType}, capped)
	if got := len([]rune(padded)); got != typeColumnCapSolo {
		t.Errorf("padded type width = %d, want %d", got, typeColumnCapSolo)
	}
	if !strings.HasSuffix(padded, "…") {
		t.Errorf("truncated type should end with an ellipsis: %q", padded)
	}
}

// TestWrapOrTruncatePlain: solo/narrow (maxWrap<=1) truncates with "…";
// dock (maxWrap>1) wraps up to maxWrap lines, truncating the last.
// Unchanged by the digest grammar rework (T005).
func TestWrapOrTruncatePlain(t *testing.T) {
	short := "short line"
	if got := wrapOrTruncatePlain(short, 40, 1, 0); len(got) != 1 || got[0] != short {
		t.Errorf("short line under width should pass through: %v", got)
	}

	long := strings.Repeat("x", 50)
	got := wrapOrTruncatePlain(long, 20, 1, 0)
	if len(got) != 1 {
		t.Fatalf("maxWrap=1 must yield exactly one line: %v", got)
	}
	if !strings.HasSuffix(got[0], "…") {
		t.Errorf("truncated line must end with an ellipsis: %q", got[0])
	}
	if len([]rune(got[0])) != 20 {
		t.Errorf("truncated line width = %d, want 20", len([]rune(got[0])))
	}

	wrapped := wrapOrTruncatePlain("one two three four five six seven eight nine ten", 12, 3, 0)
	if len(wrapped) > 3 {
		t.Fatalf("dock wrap must cap at maxWrap=3 lines: %v", wrapped)
	}
	if !strings.HasSuffix(wrapped[len(wrapped)-1], "…") && len(wrapped) == 3 {
		// Only the capped case must show truncation; a run that wrapped
		// itself under 3 lines is not truncated.
		t.Logf("wrapped output: %v", wrapped)
	}
}

// anyRole reports whether any seg in segs carries the given role.
func anyRole(segs []seg, role segRole) bool {
	for _, s := range segs {
		if s.Role == role {
			return true
		}
	}
	return false
}

// --- styleWrapLine (T021/R4): segment-wise styling must never change what
// gets displayed, only how it's colored — these tests strip the role
// attribution back to plain text and compare against wrapOrTruncatePlain's
// own output for the identical source line, across the whole catalog.

// plainOf concatenates a []styledLine's runes, one physical line per output
// entry — the "what would render on screen, ignoring color" projection.
func plainOf(lines []styledLine) []string {
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = string(l.Runes)
	}
	return out
}

// TestStyleWrapLinePlainEquivalence: for every catalog fixture type, at
// both solo (maxWrap=1) and dock (maxWrap=3, narrow width) geometry,
// styleWrapLine's plain-text projection must exactly match
// wrapOrTruncatePlain(plainChronicleLine(...)) — the pre-existing,
// already-tested wrap/truncate behavior (T005) must be bit-for-bit
// unchanged by the styling rework.
func TestStyleWrapLinePlainEquivalence(t *testing.T) {
	names := []string{"Ash", "Birch", "Cedar", "Rowan"}
	widths := []struct {
		width, maxWrap int
		dock           bool
	}{
		{60, 1, false}, // solo
		{18, 3, true},  // dock: narrow enough to force wraps on most digests
	}
	for typ, fx := range catalogFixture {
		e := store.Event{Seq: 1, Tick: 12345, Type: typ, Payload: json.RawMessage(fx.payload)}
		l := formatChronicleLine(e, names, nil)
		for _, w := range widths {
			cols := computeChronicleColumns([]chronicleLine{l}, w.dock)
			plain := wrapOrTruncatePlain(plainChronicleLine(l, cols), w.width, w.maxWrap, 0)
			prefix := chronicleLinePrefix(l, cols)
			styled := plainOf(styleWrapLine(prefix, l.Summary, w.width, w.maxWrap, 0))
			if len(styled) != len(plain) {
				t.Fatalf("%s (width=%d dock=%v): styleWrapLine produced %d lines, wrapOrTruncatePlain produced %d\nstyled: %v\nplain:  %v",
					typ, w.width, w.dock, len(styled), len(plain), styled, plain)
			}
			for i := range plain {
				if styled[i] != plain[i] {
					t.Errorf("%s (width=%d dock=%v) line %d:\nstyled: %q\nplain:  %q", typ, w.width, w.dock, i, styled[i], plain[i])
				}
			}
		}
	}
}

// TestStyleWrapLineMidWordRoleBoundary: a role boundary that falls
// mid-word (agent.spear_broke: name "Ash" immediately followed by
// "'s spear broke", no space) must keep "Ash's" as one unbroken word — not
// split into "Ash" and "'s" by a spuriously inserted wrap-space — while
// still carrying the correct role per half (name, then plain text). This is
// exactly the case a naive per-seg-independent word split (instead of
// flattening to one string before finding word boundaries) would corrupt.
func TestStyleWrapLineMidWordRoleBoundary(t *testing.T) {
	e := store.Event{Seq: 1, Tick: 1, Type: "agent.spear_broke", Payload: json.RawMessage(`{"agent":0}`)}
	l := formatChronicleLine(e, testNames, nil)
	if want := "Ash's spear broke"; plainSegs(l.Summary) != want {
		t.Fatalf("setup: summary = %q, want %q", plainSegs(l.Summary), want)
	}
	cols := computeChronicleColumns([]chronicleLine{l}, true)
	prefix := chronicleLinePrefix(l, cols)
	lines := styleWrapLine(prefix, l.Summary, 12, 3, 0)

	full := strings.Join(plainOf(lines), " ")
	if !strings.Contains(full, "Ash's") {
		t.Fatalf("mid-word boundary corrupted across wrap — \"Ash's\" split apart: %v", plainOf(lines))
	}
	for _, ln := range lines {
		idx := strings.Index(string(ln.Runes), "Ash's")
		if idx < 0 {
			continue
		}
		r := []rune(string(ln.Runes))
		// "Ash" (3 runes at idx) must be styleRoleName; "'s" (2 runes right
		// after) must NOT be — the seg boundary must not leak past "Ash".
		for i := idx; i < idx+3 && i < len(ln.Roles); i++ {
			if ln.Roles[i] != styleRoleName {
				t.Errorf("rune %q at %d (\"Ash\") = role %v, want styleRoleName", r[i], i, ln.Roles[i])
			}
		}
		for i := idx + 3; i < idx+5 && i < len(ln.Roles); i++ {
			if ln.Roles[i] == styleRoleName {
				t.Errorf("rune %q at %d (\"'s\") incorrectly carries styleRoleName", r[i], i)
			}
		}
	}
}

// --- spec 115: wrap budget, hanging indent, narrow fallback ---------------

// longFeedText is prose long enough to wrap at every width these tests use.
const longFeedText = "I keep coming back to the chest by the river and whether anyone " +
	"actually saw who opened it, because nobody will say a word about it while Rowan is in earshot."

// TestWrapBudgetDomain (spec 115 T008, contract §2): the budget's three
// domains. 0 is unbounded — the value the solo feed passes, and the one that
// did not exist before this feature.
func TestWrapBudgetDomain(t *testing.T) {
	const width = 40
	if got := wrapOrTruncatePlain(longFeedText, width, 1, 0); len(got) != 1 || !strings.HasSuffix(got[0], "…") {
		t.Errorf("maxWrap=1 must truncate to one elided line, got %d lines: %v", len(got), got)
	}
	capped := wrapOrTruncatePlain(longFeedText, width, 3, 0)
	if len(capped) != 3 {
		t.Errorf("maxWrap=3 must cap at 3 lines, got %d: %v", len(capped), capped)
	}
	if !strings.HasSuffix(capped[len(capped)-1], "…") {
		t.Errorf("a capped wrap must elide its last line: %q", capped[len(capped)-1])
	}
	unbounded := wrapOrTruncatePlain(longFeedText, width, wrapUnbounded, 0)
	if len(unbounded) <= 3 {
		t.Fatalf("wrapUnbounded must exceed the capped form for this text, got %d lines", len(unbounded))
	}
	for i, ln := range unbounded {
		if strings.Contains(ln, "…") {
			t.Errorf("unbounded wrap must never elide (line %d): %q", i, ln)
		}
	}
	if joined := strings.Join(unbounded, " "); joined != longFeedText {
		t.Errorf("unbounded wrap lost or altered text:\n got %q\nwant %q", joined, longFeedText)
	}
}

// TestWrapShortLineVerbatim (spec 115 FR-009): a line that fits is returned
// byte-identically. This is the regression the golden frames caught during
// implementation — routing short rows through wrapText collapses the runs of
// spaces that form the feed's column padding, silently destroying alignment on
// every row that never needed wrapping.
func TestWrapShortLineVerbatim(t *testing.T) {
	padded := "396000 20:00  agent.moved               Oak → (22,40)"
	for _, budget := range []int{wrapUnbounded, 1, 3} {
		got := wrapOrTruncatePlain(padded, 120, budget, 30)
		if len(got) != 1 || got[0] != padded {
			t.Errorf("budget=%d: short line must pass through verbatim, got %q", budget, got)
		}
	}
}

// TestWrapHangingIndent (spec 115 FR-003/FR-005, contract §3): continuation
// lines begin at the indent column and carry nothing else.
func TestWrapHangingIndent(t *testing.T) {
	const width, indent = 80, 30
	prefix := strings.Repeat("P", indent)
	lines := wrapOrTruncatePlain(prefix+longFeedText, width, wrapUnbounded, indent)
	if len(lines) < 2 {
		t.Fatalf("want a wrapped line, got %d: %v", len(lines), lines)
	}
	if !strings.HasPrefix(lines[0], prefix) {
		t.Errorf("first line must carry the prefix verbatim: %q", lines[0])
	}
	for i, ln := range lines[1:] {
		if !strings.HasPrefix(ln, strings.Repeat(" ", indent)) {
			t.Errorf("continuation %d must start at the indent column: %q", i+1, ln)
		}
		if strings.TrimSpace(ln[:indent]) != "" {
			t.Errorf("continuation %d must carry no prefix content: %q", i+1, ln)
		}
		if len([]rune(ln)) > width {
			t.Errorf("continuation %d exceeds width %d: %d runes", i+1, width, len([]rune(ln)))
		}
	}
}

// TestWrapNarrowFallback (spec 115 FR-006, contract §3): an indent that would
// leave under minWrapTextWidth collapses to zero — never to a reduced value,
// because a partial indent aligns to nothing.
func TestWrapNarrowFallback(t *testing.T) {
	if got := resolveWrapIndent(40, 20); got != 0 {
		t.Errorf("width 40 indent 20 leaves %d < %d of text: want indent 0, got %d",
			20, minWrapTextWidth, got)
	}
	if got := resolveWrapIndent(80, 30); got != 30 {
		t.Errorf("width 80 indent 30 leaves 50 of text: want the indent kept, got %d", got)
	}
	const width, indent = 40, 20
	lines := wrapOrTruncatePlain(strings.Repeat("P", indent)+longFeedText, width, wrapUnbounded, indent)
	for i, ln := range lines[1:] {
		if strings.HasPrefix(ln, " ") {
			t.Errorf("fallback must emit no indent (line %d): %q", i+1, ln)
		}
		if len([]rune(ln)) > width {
			t.Errorf("line %d exceeds width: %q", i+1, ln)
		}
	}
}

// TestWrapPlainStyledEquivalence (spec 115 T009): the plain and styled paths
// must produce identical characters for the same input, budget and indent —
// the two are separate implementations kept deliberately in lockstep, and a
// divergence would show as alert rows aligning differently from ordinary rows
// in the same frame.
func TestWrapPlainStyledEquivalence(t *testing.T) {
	prefix := "396000 20:00  agent.thought             "
	summary := []seg{{Text: longFeedText, Role: segSpeech}}
	for _, tc := range []struct{ width, budget, indent int }{
		{160, wrapUnbounded, len(prefix)},
		{80, wrapUnbounded, len(prefix)},
		{60, 3, len(prefix)},
		{40, wrapUnbounded, len(prefix)}, // trips the narrow fallback
		{120, 1, len(prefix)},            // truncate path
		{100, wrapUnbounded, 0},
	} {
		plain := wrapOrTruncatePlain(prefix+longFeedText, tc.width, tc.budget, tc.indent)
		styled := plainOf(styleWrapLine(prefix, summary, tc.width, tc.budget, tc.indent))
		if len(styled) != len(plain) {
			t.Fatalf("width=%d budget=%d indent=%d: styled %d lines, plain %d\nstyled: %v\nplain:  %v",
				tc.width, tc.budget, tc.indent, len(styled), len(plain), styled, plain)
		}
		for i := range plain {
			if styled[i] != plain[i] {
				t.Errorf("width=%d budget=%d indent=%d line %d:\n styled %q\n plain  %q",
					tc.width, tc.budget, tc.indent, i, styled[i], plain[i])
			}
		}
	}
}

// TestWrapDegenerateWidths (spec 115 FR-007, US3): no panic, no zero-width
// text column, nothing exceeding the pane — including a single word longer
// than the column, which cannot be broken between words.
func TestWrapDegenerateWidths(t *testing.T) {
	unbreakable := strings.Repeat("x", 200)
	for _, width := range []int{1, 5, 10, 24, 40, 80, 160} {
		for _, indent := range []int{0, 10, 30, 500} {
			for _, budget := range []int{wrapUnbounded, 1, 3} {
				lines := wrapOrTruncatePlain(longFeedText, width, budget, indent)
				if len(lines) == 0 {
					t.Errorf("width=%d indent=%d budget=%d produced no lines", width, indent, budget)
				}
				got := wrapOrTruncatePlain(unbreakable, width, budget, indent)
				if len(got) == 0 {
					t.Errorf("unbreakable word at width=%d indent=%d budget=%d produced no lines",
						width, indent, budget)
				}
			}
		}
	}
}

// TestAlertTierWrapsWithTheSameIndent (spec 115 T020, contract §5): the alert
// and labeled-voice tiers render whole-line rather than prefix-plus-summary,
// through a different function. They must land on the same indent as ordinary
// rows — otherwise a death or a norm violation scrolling past would align
// differently from every row around it, in the same frame.
func TestAlertTierWrapsWithTheSameIndent(t *testing.T) {
	const width = 100
	mk := func(typ string) chronicleLine {
		return chronicleLine{
			Tick: 396000, Time: "20:00", Type: typ,
			Family:  eventFamilyOf(typ),
			Summary: []seg{{Text: longFeedText, Role: segSpeech}},
		}
	}
	alert := mk("agent.died") // isAlertType -> whole-line path
	plainRow := mk("agent.thought")
	cols := computeChronicleColumns([]chronicleLine{alert, plainRow}, false)
	indent := len([]rune(chronicleLinePrefix(alert, cols)))

	for _, l := range []chronicleLine{alert, plainRow} {
		out := ansiRe.ReplaceAllString(renderChronicleRow(l, cols, width, wrapUnbounded, false), "")
		lines := strings.Split(out, "\n")
		if len(lines) < 2 {
			t.Fatalf("%s did not wrap at width %d: %q", l.Type, width, out)
		}
		for i, ln := range lines[1:] {
			got := len(ln) - len(strings.TrimLeft(ln, " "))
			if got != indent {
				t.Errorf("%s continuation %d starts at column %d, want %d\n%q",
					l.Type, i+1, got, indent, ln)
			}
		}
	}
}

// TestSelectedWrappedRowIsWhollyReversed (spec 115 T021, contract §6): a
// selected row that occupies several lines must read as ONE selection. Nothing
// exercised selection across a wrapped row before this feature, because before
// it no row could wrap at a width where selection is used.
func TestSelectedWrappedRowIsWhollyReversed(t *testing.T) {
	const width = 100
	l := chronicleLine{
		Tick: 396000, Time: "20:00", Type: "agent.thought",
		Family:  eventFamilyOf("agent.thought"),
		Summary: []seg{{Text: longFeedText, Role: segSpeech}},
	}
	// lipgloss emits no escapes under the default Ascii profile a test binary
	// gets, so selection would be invisible to this assertion without pinning
	// a real profile first.
	withColorProfile(t, termenv.TrueColor)
	cols := computeChronicleColumns([]chronicleLine{l}, false)
	lines := strings.Split(renderChronicleRow(l, cols, width, wrapUnbounded, true), "\n")
	if len(lines) < 2 {
		t.Fatalf("row did not wrap: %v", lines)
	}
	for i, ln := range lines {
		if !strings.Contains(ln, "\x1b[7m") {
			t.Errorf("selected row line %d carries no reverse styling: %q", i, ln)
		}
	}
}

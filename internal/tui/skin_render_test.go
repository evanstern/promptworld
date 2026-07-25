package tui

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/skin"
	"github.com/evanstern/promptworld/internal/store"
)

// TestTypeColumnFamilyAlias (spec 052 FR-013, T010): the chronicle Type
// column displays the frozen `metatron.*` family through the skin's family
// label (default `guardian`); the dock short form and the detail pane's
// verbatim type stay raw; curriculum.* is deliberately raw (inspector-class
// visibility, spec edge cases).
func TestTypeColumnFamilyAlias(t *testing.T) {
	e := store.Event{Seq: 9, Tick: 100, Type: "metatron.nudged",
		Payload: json.RawMessage(`{"form":"vision","targets":[0],"text":"x"}`)}
	l := formatChronicleLine(e, []string{"Ash"}, nil)

	if l.DisplayType != "guardian.nudged" {
		t.Errorf("DisplayType = %q, want guardian.nudged", l.DisplayType)
	}
	if l.Type != "metatron.nudged" {
		t.Errorf("raw Type mutated: %q", l.Type)
	}

	// Solo column renders the alias…
	cols := computeChronicleColumns([]chronicleLine{l}, false)
	prefix := chronicleLinePrefix(l, cols)
	if !strings.Contains(prefix, "guardian.nudged") || strings.Contains(prefix, "metatron") {
		t.Errorf("solo prefix not aliased: %q", prefix)
	}
	// …the dock short form is the raw type's last segment (family-free)…
	dockCols := computeChronicleColumns([]chronicleLine{l}, true)
	if p := chronicleLinePrefix(l, dockCols); !strings.Contains(p, "nudged") || strings.Contains(p, "guardian") {
		t.Errorf("dock prefix should be the bare short form: %q", p)
	}
	// …and the detail pane stays verbatim (FR-020 audience ruling).
	if insp := formatInspector(e, []string{"Ash"}); !strings.Contains(insp, `"metatron.nudged"`) {
		t.Errorf("detail pane must keep the raw type:\n%s", insp)
	}

	// A custom skin's family label re-aliases the column.
	sk, _ := skin.Parse([]byte(`{"strings": {"skin.guardian.family_label": "raven"}}`))
	if got := displayEventType("metatron.nudged", sk); got != "raven.nudged" {
		t.Errorf("custom family label = %q, want raven.nudged", got)
	}

	// curriculum.* stays raw by design.
	if got := displayEventType("curriculum.stage_unlocked", nil); got != "curriculum.stage_unlocked" {
		t.Errorf("curriculum must stay raw, got %q", got)
	}
}

// TestChronicleSubjectIsSkinName (spec 052 US2 AS-2 / US3 AS-1): the digest
// grammar's guardian-family subject lines render the skin-resolved name.
func TestChronicleSubjectIsSkinName(t *testing.T) {
	e := store.Event{Seq: 1, Tick: 1, Type: "metatron.nudged",
		Payload: json.RawMessage(`{"form":"vision","targets":[0],"text":"beware"}`)}

	def := formatChronicleLine(e, []string{"Ash"}, nil)
	if got := plainSegs(def.Summary); !strings.HasPrefix(got, "Guardian vision → Ash") {
		t.Errorf("default subject = %q, want Guardian …", got)
	}

	sk, _ := skin.Parse([]byte(`{"name": "Raven", "strings": {"skin.guardian.vision_noun": "whisper"}}`))
	custom := formatChronicleLine(e, []string{"Ash"}, sk)
	if got := plainSegs(custom.Summary); !strings.HasPrefix(got, "Raven whisper → Ash") {
		t.Errorf("custom subject = %q, want Raven whisper …", got)
	}
}

// TestTranscriptLabelIsSkinEpithet (spec 052 FR-007): the stored guardian
// transcript rows display through the skin's epithet.
func TestTranscriptLabelIsSkinEpithet(t *testing.T) {
	label, text, _ := classifyTranscriptLine(transcriptGuardianPrefix+"all is well", nil)
	if label != "guardian" || text != "all is well" {
		t.Errorf("default label/text = %q/%q", label, text)
	}
	sk, _ := skin.Parse([]byte(`{"epithet": "raven"}`))
	if label, _, _ := classifyTranscriptLine(transcriptGuardianPrefix+"x", sk); label != "raven" {
		t.Errorf("custom label = %q, want raven", label)
	}
	// A longer epithet still aligns its wrapped continuations.
	rows := transcriptRowLines(transcriptGuardianPrefix+strings.Repeat("word ", 30), 40, nil)
	if len(rows) < 2 {
		t.Fatalf("expected wrapped rows, got %d", len(rows))
	}
}

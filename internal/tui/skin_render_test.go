package tui

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/skin"
	"github.com/evanstern/promptworld/internal/store"
)

// TestTypeColumnRendersRawType (spec 094 US3.3): the chronicle Type column
// renders guardian.* NATIVELY — TASK-121's display-alias shim (spec 052
// FR-013's displayEventType) is deleted; persisted types are the display
// vocabulary on every surface (solo column, dock short form, detail pane).
func TestTypeColumnRendersRawType(t *testing.T) {
	e := store.Event{Seq: 9, Tick: 100, Type: "guardian.nudged",
		Payload: json.RawMessage(`{"form":"vision","targets":[0],"text":"x"}`)}
	l := formatChronicleLine(e, []string{"Ash"}, nil)

	if l.Type != "guardian.nudged" {
		t.Errorf("raw Type mutated: %q", l.Type)
	}

	// Solo column renders the persisted type verbatim…
	cols := computeChronicleColumns([]chronicleLine{l}, false)
	prefix := chronicleLinePrefix(l, cols)
	if !strings.Contains(prefix, "guardian.nudged") || strings.Contains(prefix, "metatron") {
		t.Errorf("solo prefix must carry the raw guardian.* type: %q", prefix)
	}
	// …the dock short form is the raw type's last segment (family-free)…
	dockCols := computeChronicleColumns([]chronicleLine{l}, true)
	if p := chronicleLinePrefix(l, dockCols); !strings.Contains(p, "nudged") || strings.Contains(p, "guardian") {
		t.Errorf("dock prefix should be the bare short form: %q", p)
	}
	// …and the detail pane stays verbatim (FR-020 audience ruling).
	if insp := formatInspector(e, []string{"Ash"}); !strings.Contains(insp, `"guardian.nudged"`) {
		t.Errorf("detail pane must keep the raw type:\n%s", insp)
	}

	// A skin's family label does NOT touch the Type column anymore — the
	// alias was the interim shim, retired with the real rename.
	sk, _ := skin.Parse([]byte(`{"strings": {"skin.guardian.family_label": "raven"}}`))
	skinned := formatChronicleLine(e, []string{"Ash"}, sk)
	if skinned.Type != "guardian.nudged" {
		t.Errorf("a skin must not alias the Type column: %q", skinned.Type)
	}
}

// TestChronicleSubjectIsSkinName (spec 052 US2 AS-2 / US3 AS-1): the digest
// grammar's guardian-family subject lines render the skin-resolved name.
func TestChronicleSubjectIsSkinName(t *testing.T) {
	e := store.Event{Seq: 1, Tick: 1, Type: "guardian.nudged",
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

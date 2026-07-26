package tui

import (
	"testing"

	"github.com/evanstern/promptworld/internal/tool"
)

// TestExplainGlyphsMirrorLegend (spec 063 T002): internal/tool's explain
// "glyphs" fact sheet mirrors this package's mapGlyphs table (the one legend
// source renderMapGrid and the help overlay share) — tool is a leaf and
// cannot import tui, so the copy is pinned equal here, row for row, plus the
// trailing agent-glyph note.
func TestExplainGlyphsMirrorLegend(t *testing.T) {
	rows, note := tool.MapGlyphLegend()
	if len(rows) != len(mapGlyphs) {
		t.Fatalf("tool legend mirror has %d rows, mapGlyphs has %d", len(rows), len(mapGlyphs))
	}
	for i, g := range mapGlyphs {
		if rows[i] != [3]string{g.Glyph, g.Name, g.Meaning} {
			t.Errorf("legend row %d drifted: tool mirror %v, mapGlyphs {%q %q %q}", i, rows[i], g.Glyph, g.Name, g.Meaning)
		}
	}
	if note != agentGlyphNote {
		t.Errorf("agent-glyph note drifted: tool mirror %q, help.go %q", note, agentGlyphNote)
	}
}

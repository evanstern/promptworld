package metatron

import (
	"regexp"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/persona"
	"github.com/evanstern/promptworld/internal/tool"
)

// promptDenylistRe is the composed-prompt half of the spec 052 SC-001 sweep:
// no angel-fiction vocabulary in any DEFAULT-skin model-facing prompt. The
// frozen work_miracle tool id (and its "(the work_miracle tool)" gloss) is
// the one allowed "miracle" spelling — the model must be able to call the
// real tool (ruling 2).
var promptDenylistRe = regexp.MustCompile(`(?i)\b(metatron|angels?|miracles?|divine|heavens?|scriptures?)\b`)

func assertPromptClean(t *testing.T, surface, text string) {
	t.Helper()
	for _, m := range promptDenylistRe.FindAllStringIndex(text, -1) {
		word := strings.ToLower(text[m[0]:m[1]])
		if word == "miracle" && m[0] >= len("work_") && text[m[0]-len("work_"):m[0]] == "work_" {
			continue // the frozen tool id
		}
		t.Errorf("%s: fiction vocabulary %q in composed prompt:\n…%s…",
			surface, text[m[0]:m[1]], text[max0(m[0]-40):min(len(text), m[1]+40)])
	}
}

// TestDefaultPromptsAreFictionFree (spec 052 US2 AS-4, SC-001): the default
// experience's full prompt surface — charter seeds, the assembled turn
// system prompt (fixed frame + tool guidance included), the digest keeper,
// and the watch confirmer — carries no angel fiction.
func TestDefaultPromptsAreFictionFree(t *testing.T) {
	roster := tool.LoopRosterMetatron()
	assertPromptClean(t, "turn system prompt (default charter)",
		turnSystemPrompt(persona.DefaultCharter, nil, roster))
	assertPromptClean(t, "turn system prompt (tutor charter)",
		turnSystemPrompt(persona.TutorCharter, nil, roster))
	assertPromptClean(t, "digest keeper", digestKeeperSystem("Guardian"))
	assertPromptClean(t, "watch confirmer", confirmSystem("Guardian"))
	assertPromptClean(t, "conversation-only system prompt",
		turnSystemPrompt(persona.DefaultCharter, nil, nil))
}

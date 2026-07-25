package skin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"
)

// The skin bundle's field caps (contract §1). Identity fields are a name-
// injection surface (spec 052 edge cases), so they are single-line and
// tightly length-capped; the voice is the one long-form skin text and reuses
// the bundle-SOUL character cap — hostile voice text is contained by the
// fixed frame, not by validation.
const (
	nameMaxRunes    = 40
	epithetMaxRunes = 20 // shared by epithet and tab_label
	voiceMaxChars   = 4000
	// stringMaxRunes caps one string-token override's value: overrides are
	// display vocabulary (nouns, labels, short lines), never long-form text
	// — the voice is the only long-form field by design (contract §1).
	stringMaxRunes = 120
	// lineMaxRunes caps a stage identity's one-line description.
	lineMaxRunes = 120
)

// skinDoc is the parse target for <world>/skin.json (contract §1).
type skinDoc struct {
	Name     string            `json:"name"`
	Epithet  string            `json:"epithet"`
	TabLabel string            `json:"tab_label"`
	Voice    string            `json:"voice"`
	Strings  map[string]string `json:"strings"`
	Stages   map[string]struct {
		Name string `json:"name"`
		Line string `json:"line"`
	} `json:"stages"`
}

// knownSkinKeys is skin.json's top-level vocabulary; unknown keys are
// ignored with one notice (a typo never bricks the guardian — FR-003).
var knownSkinKeys = map[string]bool{
	"name": true, "epithet": true, "tab_label": true,
	"voice": true, "strings": true, "stages": true,
}

// Load reads <worldDir>/skin.json following the capabilities.json fallback
// discipline (spec 052 FR-003, research R1/R4):
//
//   - no file → the default skin, silently (the common case);
//   - unreadable / malformed JSON → the default skin + one notice;
//   - an invalid FIELD → that field's default + a notice, the rest applies
//     (the bundle never fails wholesale for one bad field);
//   - unknown top-level keys / unknown token paths → ignored + notice.
//
// The result is boot-frozen (the SetBundles/SetStage discipline): edits take
// effect on restart, and the status surface reports the active skin so a
// stale edit is diagnosable.
func Load(worldDir string) (*Skin, []string) {
	path := filepath.Join(worldDir, "skin.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Default(), nil
	}
	if err != nil {
		return Default(), []string{"skin.json could not be read — serving the default skin"}
	}
	return Parse(data)
}

// Parse validates one skin.json document field-wise (the loader's testable
// core; Load adds only the file read). Never returns a nil Skin.
func Parse(data []byte) (*Skin, []string) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return Default(), []string{"skin.json is not valid JSON — serving the default skin"}
	}
	var unknown []string
	for k := range raw {
		if !knownSkinKeys[k] {
			unknown = append(unknown, k)
		}
	}
	sort.Strings(unknown)
	var notices []string
	if len(unknown) > 0 {
		notices = append(notices, "skin.json has unknown key(s): "+strings.Join(unknown, ", ")+" — ignored")
	}

	var doc skinDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		return Default(), []string{"skin.json is not valid JSON — serving the default skin"}
	}

	s := &Skin{strings: map[string]string{}, stages: map[string]StageIdentity{}}
	setIdentity := func(field, token, v string, maxRunes int) {
		if v == "" {
			return // absent field: default, no notice
		}
		if !utf8.ValidString(v) || !singleLine(v) || strings.TrimSpace(v) == "" || utf8.RuneCountInString(v) > maxRunes {
			notices = append(notices, fmt.Sprintf(
				"skin.json %s is invalid (single line, ≤%d characters) — using the default", field, maxRunes))
			return
		}
		s.strings[token] = v
	}
	setIdentity("name", TokenName, doc.Name, nameMaxRunes)
	setIdentity("epithet", TokenEpithet, doc.Epithet, epithetMaxRunes)
	setIdentity("tab_label", TokenTabLabel, doc.TabLabel, epithetMaxRunes)

	// The voice: the one long-form field, the bundle-SOUL cap (contract §1).
	// Over-cap or non-UTF-8 falls back to no fragment — hostile CONTENT is
	// the fixed frame's job (FR-004), the cap only bounds volume.
	switch {
	case doc.Voice == "":
	case !utf8.ValidString(doc.Voice):
		notices = append(notices, "skin.json voice is not valid UTF-8 — using no voice")
	case utf8.RuneCountInString(doc.Voice) > voiceMaxChars:
		notices = append(notices, fmt.Sprintf("skin.json voice exceeds the %d-character cap — using no voice", voiceMaxChars))
	default:
		s.voice = doc.Voice
	}

	// String-token overrides: unknown token paths are ignored with a notice
	// (contract §1); identity tokens set here compose with (and lose to) the
	// typed fields above only in the sense that both write the same table —
	// the typed field wins by writing last.
	var badTokens []string
	for _, token := range sortedKeys(doc.Strings) {
		v := doc.Strings[token]
		if _, ok := defaultTable[token]; !ok {
			badTokens = append(badTokens, token)
			continue
		}
		if isIdentityToken(token) && s.strings[token] != "" {
			continue // the typed field already set it
		}
		if !utf8.ValidString(v) || !singleLine(v) || strings.TrimSpace(v) == "" || utf8.RuneCountInString(v) > stringMaxRunes {
			notices = append(notices, fmt.Sprintf(
				"skin.json strings[%q] is invalid (single line, ≤%d characters) — using the default", token, stringMaxRunes))
			continue
		}
		s.strings[token] = v
	}
	if len(badTokens) > 0 {
		notices = append(notices, "skin.json strings has unknown token(s): "+strings.Join(badTokens, ", ")+" — ignored")
	}

	// Stage display identities: keys must be neutral ladder ids; a missing
	// name/line falls back to the default identity's half (FR-011 — the
	// substrate stays neutral, the skin supplies display only).
	var badStages []string
	for _, id := range sortedKeys(doc.Stages) {
		def, ok := defaultStages[id]
		if !ok {
			badStages = append(badStages, id)
			continue
		}
		st := doc.Stages[id]
		si := def
		if st.Name != "" {
			if utf8.ValidString(st.Name) && singleLine(st.Name) && utf8.RuneCountInString(st.Name) <= nameMaxRunes {
				si.Name = st.Name
			} else {
				notices = append(notices, fmt.Sprintf(
					"skin.json stages[%q].name is invalid (single line, ≤%d characters) — using the default", id, nameMaxRunes))
			}
		}
		if st.Line != "" {
			if utf8.ValidString(st.Line) && singleLine(st.Line) && utf8.RuneCountInString(st.Line) <= lineMaxRunes {
				si.Line = st.Line
			} else {
				notices = append(notices, fmt.Sprintf(
					"skin.json stages[%q].line is invalid (single line, ≤%d characters) — using the default", id, lineMaxRunes))
			}
		}
		s.stages[id] = si
		// The stage-name token twin: a stage override is reachable through
		// Resolve too (one table, contract §2).
		s.strings["skin.stage."+id+".name"] = si.Name
		s.strings["skin.stage."+id+".line"] = si.Line
	}
	if len(badStages) > 0 {
		notices = append(notices, "skin.json stages has unknown stage id(s): "+strings.Join(badStages, ", ")+" — ignored")
	}

	if len(s.strings) == 0 {
		s.strings = nil
	}
	if len(s.stages) == 0 {
		s.stages = nil
	}
	return s, notices
}

func isIdentityToken(t string) bool {
	return t == TokenName || t == TokenEpithet || t == TokenTabLabel
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Package skin is the runtime skin substrate (spec 052, TASK-121): the
// fiction layer as data. The substrate knows neutral stage ids
// (world.Stage1..Stage4) and fixed mechanics vocabulary; the active skin
// supplies every fiction string the player SEES — the guardian's display
// name, epithet, tab label, chronicle vocabulary, and stage display
// identities — through one token lookup with the resolution order
//
//	world skin.json override → compiled default table → the token path itself
//
// A rendered token path is a visible bug, never an empty string (FR-001);
// the token-completeness test (completeness_test.go) fails before it ships.
// The default skin is the secular-mythic Guardian (spec 052 ruling 3). The
// event log is skin-free (ruling 1): nothing in this package ever touches
// recorded payloads — skinning is render-time and prompt-composition only.
package skin

import "strings"

// StageIdentity is one stage's display identity: the skin's name for the
// stage and its one-line identity description, presented at world creation
// and on status surfaces. Never easy-mode framing (TASK-68 AC #7) — an
// identity, not a difficulty label.
// The JSON tags are the status surface's wire form (spec 052 contract §7,
// skin_stages) — additive, and unused by the skin.json loader (which has its
// own parse target in load.go).
type StageIdentity struct {
	Name string `json:"name"` // display name, e.g. "The Voice"
	Line string `json:"line"` // one-line identity description
}

// defaultStages maps neutral stage ids to the default Guardian skin's display
// identities (client-approved strings — do not reword without client review;
// spec 052 ruling 3 keeps them across the de-theme).
var defaultStages = map[string]StageIdentity{
	"stage-1": {Name: "The Voice", Line: "you speak, it acts"},
	"stage-2": {Name: "The Written Word", Line: "your law outlives the conversation"},
	"stage-3": {Name: "The Craft", Line: "you shape what it can do"},
	"stage-4": {Name: "The Stewardship", Line: "a world in your care"},
}

// The identity-field token paths (contract §2: the typed skin.json fields are
// convenience spellings of these tokens — one table, two doors).
const (
	TokenName        = "skin.guardian.name"
	TokenEpithet     = "skin.guardian.epithet"
	TokenTabLabel    = "skin.guardian.tab_label"
	TokenFamilyLabel = "skin.guardian.family_label"
)

// defaultTable is the compiled-in default skin — the complete token authority
// (contract §3, normative). Every fiction token any surface consumes MUST
// have a row here (the token-completeness test enforces it); systems/
// telemetry surfaces have no tokens by construction (D10), which is what
// makes them unskinnable. Stage tokens are appended from defaultStages in
// init() so the two can never drift.
var defaultTable = map[string]string{
	TokenName:     "Guardian",
	TokenEpithet:  "guardian",
	TokenTabLabel: "guardian",
	// FamilyLabel's original consumer — the chronicle Type-column alias over
	// the frozen metatron.* family (spec 052 FR-013, TASK-121's interim
	// shim) — was retired by spec 094: persisted types are guardian.* now
	// and the chronicle renders them raw. The token remains part of the
	// skin data contract for skins that voice the guardian family name.
	TokenFamilyLabel:                    "guardian",
	"skin.guardian.working_noun":        "working",
	"skin.guardian.working_noun_plural": "workings",
	"skin.guardian.notes_label":         "the guardian's notes",
	// Vision/omen display vocabulary is default-skin-retained (TutorCharter
	// precedent, spec 052 assumption 1) — tokens exist so a custom skin may
	// re-voice the DISPLAY terms while the tool ids (send_vision/send_omen)
	// and recorded payload values stay frozen (FR-005/FR-009).
	"skin.guardian.vision_noun": "vision",
	"skin.guardian.omen_noun":   "omen",
	// Report-card labels (spec 063 US4, contract §4): the card's box title
	// and the attribution note's own block header — the note is additive
	// prose beneath the checklist, clearly its own block (standing
	// resolution 1).
	"skin.guardian.report_card_label": "report card",
	"skin.guardian.attribution_label": "what your words did",
	// Per-verb example asks (spec 063 US5, D9): one canned player phrasing
	// per guardian loop tool — the ? overlay's guardian section teaches how
	// asking looks at the deterministic floor. Keyed by the frozen tool id;
	// the phrasing uses the default skin's working/vision/omen nouns.
	"skin.guardian.example_ask.send_vision":     `"show Ash a vision of the fire dying"`,
	"skin.guardian.example_ask.send_omen":       `"send everyone an omen tonight: stay near the fire"`,
	"skin.guardian.example_ask.monitor_and_act": `"watch for anyone going hungry, and warn them"`,
	"skin.guardian.example_ask.cancel_order":    `"release the watch on the fire"`,
	"skin.guardian.example_ask.work_miracle":    `"work a working: grant Ash food from thin air"`,
	"skin.guardian.example_ask.pause":           `"pause the world"`,
	"skin.guardian.example_ask.start":           `"start the world again at 4x"`,
	"skin.guardian.example_ask.adjust_speed":    `"slow the world down to 1x"`,
	"skin.guardian.example_ask.explain":         `"what does a vision cost?"`,
	// The plan layer (spec 084): one example ask per plan verb, the same
	// frozen-tool-id keying.
	"skin.guardian.example_ask.place_designation":  `"mark a shelter site at 4,5"`,
	"skin.guardian.example_ask.cancel_designation": `"withdraw the shelter-site mark"`,
	"skin.guardian.example_ask.issue_directive":    `"charge Ash and Birch to raise the shelter I marked, within three days"`,
	"skin.guardian.example_ask.cancel_directive":   `"lift the charge you laid on Ash"`,
	"skin.guardian.example_ask.survey_site":        `"survey the ground around 10,10"`,
	// Prophecy (spec 085): the staked vision, same frozen-tool-id keying.
	"skin.guardian.example_ask.prophesy": `"prophesy to everyone: the shelter I marked will stand within three days"`,
	// The canonization miracle (spec 101): christening a named region, and
	// its read-only myth-briefing companion, same frozen-tool-id keying.
	"skin.guardian.example_ask.canonize_region": `"canonize Thornspire at 20,20, radius 6"`,
	"skin.guardian.example_ask.brief_myths":     `"what myths do the villagers already believe?"`,
	// The mission layer (spec 107): the player's standing instruction and
	// its bookkeeping verbs, same frozen-tool-id keying.
	"skin.guardian.example_ask.accept_mission":        `"get a second fire built near the west huts and keep it fueled"`,
	"skin.guardian.example_ask.note_mission_progress": `"how is my mission coming along? note where it stands"`,
	"skin.guardian.example_ask.cancel_mission":        `"stand down from the fire mission"`,
}

func init() {
	for id, si := range defaultStages {
		defaultTable["skin.stage."+id+".name"] = si.Name
		defaultTable["skin.stage."+id+".line"] = si.Line
	}
	for id, chapter := range defaultCeremonyChapters {
		defaultTable["skin.stage."+id+".ceremony_chapter"] = chapter
	}
}

// defaultCeremonyChapters is the D6 authorship-voice narrated chapter for
// each UNLOCKABLE stage (spec 056, research R5: "D6 voice text lives beside
// the stage identities in the skin substrate — one authored line per
// stage, player-authorship register"). stage-1 is the ladder's floor and is
// never unlocked (sim.EvaluateUnlock never returns it), so it carries no
// entry here — nothing ever resolves it. A deliberately generic "your play"
// framing (rather than "your charter") is used throughout: the gate a pass
// satisfies varies by stage (any pass at stage-1; a charter revision at
// stage-2; a player-granted tool at stage-3), so the chapter's subject
// stays true regardless of which specific evidence earned it.
var defaultCeremonyChapters = map[string]string{
	"stage-2": "Your play proved The Written Word: a law that outlives the conversation, written once and honored by every turn since.",
	"stage-3": "Your play proved The Craft: what the guardian can do now bears your own hand in its shaping.",
	"stage-4": "Your play proved The Stewardship: a world now stands in your care, exactly as you left it.",
}

// DefaultTable returns a copy of the compiled default token table — the
// completeness test's and the doc twin's enumeration surface. Mutating the
// copy never touches the authority.
func DefaultTable() map[string]string {
	out := make(map[string]string, len(defaultTable))
	for k, v := range defaultTable {
		out[k] = v
	}
	return out
}

// Skin is one world's resolved skin: string-token overrides (identity fields
// included — they ARE tokens), stage display-identity overrides, and the
// persona-voice text composed at the guardian's editable-zone SOUL seam.
// Boot-frozen: loaded once at daemon boot (Load), injected via the agent's
// SetSkin, surfaced through status; clients hold only what status carries.
// The zero value (and a nil *Skin) is the default Guardian skin — every
// method is nil-safe so "no skin.json" and "old daemon, absent status
// fields" both render the default with no special-casing at call sites.
type Skin struct {
	strings map[string]string        // token-path → override value
	stages  map[string]StageIdentity // neutral stage id → display identity
	voice   string                   // editable-zone persona voice ("" = no fragment)
}

// Default returns the default Guardian skin (an empty override set over the
// compiled table).
func Default() *Skin { return &Skin{} }

// FromFacts rebuilds a Skin from status-carried display facts (contract §7):
// the client-side twin of Load, for TUIs/CLIs that must never read world
// files (FR-012). strings/stages may be nil (default skin).
func FromFacts(strs map[string]string, stages map[string]StageIdentity) *Skin {
	s := &Skin{}
	for k, v := range strs {
		if s.strings == nil {
			s.strings = map[string]string{}
		}
		s.strings[k] = v
	}
	for id, si := range stages {
		if s.stages == nil {
			s.stages = map[string]StageIdentity{}
		}
		s.stages[id] = si
	}
	return s
}

// Resolve looks one token up: world override → default table → the token
// path itself (visibly wrong, never empty — FR-001).
func (s *Skin) Resolve(token string) string {
	if s != nil {
		if v, ok := s.strings[token]; ok {
			return v
		}
	}
	if v, ok := defaultTable[token]; ok {
		return v
	}
	return token
}

// Typed convenience accessors over the same table (contract §2).
func (s *Skin) Name() string        { return s.Resolve(TokenName) }
func (s *Skin) Epithet() string     { return s.Resolve(TokenEpithet) }
func (s *Skin) TabLabel() string    { return s.Resolve(TokenTabLabel) }
func (s *Skin) FamilyLabel() string { return s.Resolve(TokenFamilyLabel) }

// WorkingNoun / WorkingNounPlural are the display vocabulary for the frozen
// work_miracle mechanics family ("miracle" de-themes to "working", FR-007).
func (s *Skin) WorkingNoun() string       { return s.Resolve("skin.guardian.working_noun") }
func (s *Skin) WorkingNounPlural() string { return s.Resolve("skin.guardian.working_noun_plural") }

// NotesLabel is the display reference for the guardian's soul file (the
// on-disk path metatron/soul.md is frozen; only the display name is skin).
func (s *Skin) NotesLabel() string { return s.Resolve("skin.guardian.notes_label") }

// FormNoun maps a recorded nudge form value (the frozen payload vocabulary
// "vision"/"omen", FR-005) to its skin display noun. Unknown forms render
// verbatim — honest, never empty.
func (s *Skin) FormNoun(form string) string {
	switch form {
	case "vision":
		return s.Resolve("skin.guardian.vision_noun")
	case "omen":
		return s.Resolve("skin.guardian.omen_noun")
	}
	return form
}

// ExampleAsk resolves the per-verb example-ask token family (spec 063 US5,
// D9): one canned player phrasing per guardian loop tool, keyed by the
// FROZEN tool id — the "skin.guardian.example_ask.<tool-id>" rows of the
// default table. The path is assembled from split literals deliberately:
// the token-completeness sweep matches whole dotted token literals, and
// this family's membership is enumerated by the table itself (one row per
// shipped verb; the D9 tests assert every ceiling verb resolves to a real
// row, never a raw path).
func (s *Skin) ExampleAsk(toolID string) string {
	return s.Resolve("skin.guardian." + "example_ask." + toolID)
}

// ReportCardLabel / AttributionLabel are the report card's display labels
// (spec 063 US4, contract §4).
func (s *Skin) ReportCardLabel() string  { return s.Resolve("skin.guardian.report_card_label") }
func (s *Skin) AttributionLabel() string { return s.Resolve("skin.guardian.attribution_label") }

// Voice is the persona-voice text composed into the guardian's system prompt
// at the editable-zone SOUL seam ("" = no fragment). NEVER composed after
// the fixed frame (FR-004) — the call site owns that invariant.
func (s *Skin) Voice() string {
	if s == nil {
		return ""
	}
	return s.voice
}

// Stage returns the skin's display identity for a stage id, and whether the
// id names a ladder stage. Unknown ids (including "" — an ungated pre-ladder
// world) return the zero identity and false. Overridden stages win; missing
// overrides fall through to the default skin (US1 AS-2).
func (s *Skin) Stage(id string) (StageIdentity, bool) {
	if s != nil {
		if si, ok := s.stages[id]; ok {
			return si, true
		}
	}
	si, ok := defaultStages[id]
	return si, ok
}

// StageName returns the skin's display name for a stage id, or the id itself
// when it names no ladder stage — a safe fallback for message text.
func (s *Skin) StageName(id string) string {
	if si, ok := s.Stage(id); ok {
		return si.Name
	}
	return id
}

// CeremonyChapter resolves the D6 authorship-voice narrated chapter for a
// ladder stage's unlock ceremony (spec 056, overlays/ceremony.md; research
// R5/R6) — a plain token lookup like every other fiction string in this
// package, so a world skin.json may re-voice it per stage via the generic
// `strings` override map (no dedicated typed field needed: Resolve already
// generalizes over any cataloged token path). stage-1 (never unlocked)
// resolves to its own token path — visibly wrong, honest, and never
// actually rendered by a real client (FR-001's fallback discipline).
func (s *Skin) CeremonyChapter(stage string) string {
	return s.Resolve("skin.stage." + stage + ".ceremony_chapter")
}

// StringOverrides returns a copy of the skin's token overrides (identity
// fields included) — the status surface's transport form (contract §7). nil
// for the default skin, so the wire field omits under omitempty.
func (s *Skin) StringOverrides() map[string]string {
	if s == nil || len(s.strings) == 0 {
		return nil
	}
	out := make(map[string]string, len(s.strings))
	for k, v := range s.strings {
		out[k] = v
	}
	return out
}

// StageOverrides returns a copy of the skin's stage-identity overrides — the
// status surface's transport form. nil for the default skin.
func (s *Skin) StageOverrides() map[string]StageIdentity {
	if s == nil || len(s.stages) == 0 {
		return nil
	}
	out := make(map[string]StageIdentity, len(s.stages))
	for k, v := range s.stages {
		out[k] = v
	}
	return out
}

// Stage / StageName are the package-level accessors every pre-052 call site
// uses (spec 046 R8) — they resolve the DEFAULT skin. Sites with a world
// skin in scope use the *Skin methods instead (spec 052 T004).
func Stage(id string) (StageIdentity, bool) { return (*Skin)(nil).Stage(id) }
func StageName(id string) string            { return (*Skin)(nil).StageName(id) }

// singleLine reports whether v contains no control characters (incl. \n/\r/
// \t) — the identity-field injection surface's shape rule (spec 052 edge
// cases: name/epithet/tab-label are validated single-line strings).
func singleLine(v string) bool {
	return !strings.ContainsFunc(v, func(r rune) bool { return r < 0x20 || r == 0x7f })
}

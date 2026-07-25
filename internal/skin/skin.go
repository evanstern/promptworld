// Package skin is the INTERIM home of skin-supplied display identity (spec 046
// R8): the substrate knows neutral stage ids (world.Stage1..Stage4) with
// skin-independent semantics; the active skin supplies what the player SEES.
// TASK-121 (skinnable persona) absorbs this package as the real skin substrate —
// stage ids and semantics never move when it does; only the lookup grows a
// skin dimension. Until then, the one shipped skin is the default secular-mythic
// guardian, whose stage names below are client-approved (2026-07-25, TASK-68
// AC #5) skin DATA, not substrate facts.
package skin

// StageIdentity is one stage's display identity: the guardian-skin name and its
// one-line identity description, presented at world creation and on status
// surfaces. Never easy-mode framing (TASK-68 AC #7) — an identity, not a
// difficulty label.
type StageIdentity struct {
	Name string // display name, e.g. "The Voice"
	Line string // one-line identity description
}

// stageIdentity maps neutral stage ids to the default guardian skin's display
// identities (client-approved strings — do not reword without client review).
var stageIdentity = map[string]StageIdentity{
	"stage-1": {Name: "The Voice", Line: "you speak, it acts"},
	"stage-2": {Name: "The Written Word", Line: "your law outlives the conversation"},
	"stage-3": {Name: "The Craft", Line: "you shape what it can do"},
	"stage-4": {Name: "The Stewardship", Line: "a world in your care"},
}

// Stage returns the active skin's display identity for a stage id, and whether
// the id names a ladder stage. Unknown ids (including "" — an ungated
// pre-ladder world) return the zero identity and false.
func Stage(id string) (StageIdentity, bool) {
	si, ok := stageIdentity[id]
	return si, ok
}

// StageName returns the active skin's display name for a stage id, or the id
// itself when it names no ladder stage — a safe fallback for message text.
func StageName(id string) string {
	if si, ok := stageIdentity[id]; ok {
		return si.Name
	}
	return id
}

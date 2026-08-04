package guardian

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/evanstern/promptworld/internal/bundle"
	"github.com/evanstern/promptworld/internal/persona"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/skin"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/tool"
)

// The charter is the game's base player-editable prompt, joined by an optional
// folder of player-authored skill files (spec 021) — the assistant-shaped
// instruction surface (CLAUDE.md + SKILL.md). Every one of these files is read
// fresh at the start of every Guardian turn and status peek — that per-read
// discipline IS the "edits are live next turn" mechanism (no watcher, no
// cache). The guardian never runs charterless: missing files are restored, empty
// files fall back, oversized files are truncated, over-cap skill folders are
// trimmed — and each case is reported in a `notice` so the next reply can tell
// the player, one model of "the game tells you when your file didn't load".

// presetCharter resolves a world's charter_preset name (world.json, spec 046)
// to its authored constant: ""/"default" is persona.DefaultCharter; "tutor" is
// the stage-1 orientation preset. Unknown names (already refused at world.Open)
// fall back to the default — the guardian never runs charterless.
func presetCharter(preset string) string {
	switch preset {
	case "", "default":
		return persona.DefaultCharter
	case "tutor":
		return persona.TutorCharter
	}
	return persona.DefaultCharter
}

// loadCharter returns the effective charter text and a human-readable notice
// ("" when the player's charter loaded cleanly). preset is the world's
// charter_preset name (spec 046) — the restore/fallback text; variadic so every
// pre-046 call site (and every direct-arg test) keeps compiling unchanged with
// the default-preset behavior (the loadManifest knownBundleTools precedent).
func loadCharter(worldDir string, preset ...string) (text, notice string) {
	def := persona.DefaultCharter
	if len(preset) > 0 {
		def = presetCharter(preset[0])
	}
	path := filepath.Join(worldDir, "charter.md")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		// Restore the default so the player has a file to edit again.
		os.WriteFile(path, []byte(def), 0o644)
		return def, "charter.md was missing — the default charter has been restored"
	}
	if err != nil {
		return def, "charter.md could not be read — serving under the default charter"
	}
	t := string(data)
	if strings.TrimSpace(t) == "" {
		return def, "charter.md is empty — serving under the default charter"
	}
	if len(t) > persona.CharterMaxChars {
		return t[:persona.CharterMaxChars], "charter.md exceeds the cap — only the first 4,000 characters are in effect"
	}
	return t, ""
}

// charterIsDefault reports whether the file on disk matches the authored
// default (for status display) — preset-aware (spec 046): a world created with
// a charter preset compares against that preset's constant. LEGACY-aware
// (spec 052 SC-003): a pre-052 world's untouched guardian-voiced seed is still
// game-authored text — never reclassified player-authored on upgrade.
func charterIsDefault(worldDir string, preset ...string) bool {
	def := persona.DefaultCharter
	if len(preset) > 0 {
		def = presetCharter(preset[0])
	}
	data, err := os.ReadFile(filepath.Join(worldDir, "charter.md"))
	if err != nil {
		return true
	}
	return string(data) == def || isLegacyDefault(string(data))
}

// isLegacyDefault reports whether text is one of the retired game-authored
// default charters (spec 052 SC-003): the long-lived pre-059 legacy seed,
// the brief post-059/pre-052 variant carrying the survival paragraph, or the
// pre-107 counsel-first guardian seed (spec 107 D5 replaced its counsel duty
// with the obedience clause). Any of them is game-authored text — never
// reclassified player-authored, so an untouched pre-107 world's charter.md
// keeps the spec-102 ceiling ON and its unlock gates honest after upgrade.
func isLegacyDefault(text string) bool {
	return text == persona.LegacyDefaultCharter || text == persona.LegacyDefaultCharterSurvival ||
		text == persona.LegacyDefaultCharterCounsel
}

// Instruction-surface gating by stage (spec 046 FR-005, the ladder): stage-1
// (The Voice) locks the charter to the world's preset constant; skill files
// bind from stage-3 (The Craft). An absent stage ("") is a pre-ladder, ungated
// world — everything binds, exactly as before the ladder existed.
func stageLocksCharter(stage string) bool { return stage == "stage-1" }
func stageBindsSkills(stage string) bool  { return stage != "stage-1" && stage != "stage-2" }

// stageCharter is the stage fork over loadCharter (spec 046 R3, US2): at
// stage-1 the effective charter IS the world's preset constant — sourced from
// the compiled-in text, never the file, so the lock is tamper-proof rather
// than advisory. When charter.md's bytes differ from the preset (the player
// edited it), the honest notice names the unlocking stage (FR-005 — never
// silent ignoring). A missing file is restored to the preset so the player has
// a file to edit when stage-2 unlocks — no notice, since what binds never
// changed. Every other stage behaves byte-identically to today's loadCharter.
// The lock notices resolve the unlocking stage's display name through the
// WORLD skin when one is in scope (spec 052 T004) — variadic so every
// pre-052 call site (and every direct-arg test) keeps compiling with the
// default-skin behavior, the loadCharter preset precedent.
func stageCharter(worldDir, stage, preset string, sk ...*skin.Skin) (text, notice string) {
	if !stageLocksCharter(stage) {
		return loadCharter(worldDir, preset)
	}
	text = presetCharter(preset)
	path := filepath.Join(worldDir, "charter.md")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		os.WriteFile(path, []byte(text), 0o644)
		return text, ""
	}
	if err == nil && string(data) != text {
		return text, fmt.Sprintf("charter.md does not bind at this stage — %s (stage-2) unlocks instruction authoring",
			skinOrDefault(sk).StageName("stage-2"))
	}
	return text, ""
}

// skinOrDefault unwraps a variadic skin argument (nil-safe either way).
func skinOrDefault(sk []*skin.Skin) *skin.Skin {
	if len(sk) > 0 {
		return sk[0]
	}
	return nil
}

// stageSkills is the stage fork over loadSkills (spec 046 FR-005): skill files
// compose only from stage-3 (The Craft) — the ladder's capability-design
// unlock. At stage-1/-2, present skill files are never silently ignored: one
// notice names the unlocking stage. Absent stage = pre-ladder = today's
// behavior.
func stageSkills(worldDir, stage string, sk ...*skin.Skin) ([]skillFile, []string) {
	if stageBindsSkills(stage) {
		return loadSkills(worldDir)
	}
	if names := skillNames(worldDir); len(names) > 0 {
		return nil, []string{fmt.Sprintf("skill files do not bind at this stage — %s (stage-3) unlocks skill files",
			skinOrDefault(sk).StageName("stage-3"))}
	}
	return nil, nil
}

// charterFingerprint is the effective charter's revision identity (spec 044
// FR-008, research R8): a short content hash (12 hex chars of SHA-256) over
// EXACTLY the text loadCharter returned — the post-fallback, post-truncation
// bytes the model actually runs under — so the recorded revision can never
// name a charter the guardian never executed.
func charterFingerprint(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:6])
}

// observeCharter lands the charter-revision observation (spec 044 US2,
// T014): when the effective charter's fingerprint differs from the last
// recorded one (State.CharterFingerprint, read via the absorb-side mirror),
// the turn emits guardian.charter_observed through the same InjectSocial
// door every other turn effect rides — fingerprint-at-effect semantics: the
// timeline records what the guardian actually ran with, at the turns it ran.
// The first turn of a world always emits (the mirror starts empty). Ended
// worlds skip: the door narrows to recorded prose after run end
// (endedProseWhitelist) and a finished run's evidence timeline is closed.
// The `default` flag is derived from the same effective text as the
// fingerprint (an empty/missing charter.md serves the default text, and is
// recorded as such), so the two can never disagree. PRESET-AWARE (spec 046
// reconciliation): the flag compares against the WORLD's preset constant
// (presetCharter — the same reference charterIsDefault uses), not bare
// persona.DefaultCharter, so a stage-1 tutor-preset world's effective text
// (the stage lock serves persona.TutorCharter) is honestly recorded as
// default=true — authored by the game, never the player. The morgue's
// evidence timeline and the stage-2→3 unlock gate (sim.EvaluateUnlock, via
// sim.CharterObservedEvidence's Custom = !default derivation) both depend on
// this: preset text must never masquerade as a player-authored charter.
func (mt *Guardian) observeCharter(text string) {
	fp := charterFingerprint(text)
	mt.stateMu.Lock()
	known, ended := mt.charterFP, mt.ended
	mt.stateMu.Unlock()
	if ended || fp == known {
		return
	}
	// Default derivation is legacy-aware (spec 052 SC-003): a pre-052
	// world's untouched guardian-voiced seed is game-authored, and the
	// stage-2→3 unlock gate (Custom = !default) must not count it as a
	// player-authored revision after an upgrade.
	def := text == presetCharter(mt.charterPreset) || isLegacyDefault(text)
	batch := []store.Event{{Type: "guardian.charter_observed", Payload: mustJSON(sim.CharterObservedPayload{
		Fingerprint: fp, Default: def})}}
	if err := mt.social.InjectSocial(batch); err != nil {
		log.Printf("guardian: charter observation rejected at the door: %v", err)
		return
	}
	// Optimistic mirror update so a back-to-back turn cannot double-emit
	// before the absorb goroutine reflects the landed event; mirrorState
	// re-syncs from the replica per batch (a same-value re-emit would be
	// harmless — the reducer arm is idempotent — this just keeps the log quiet).
	mt.stateMu.Lock()
	mt.charterFP = fp
	mt.stateMu.Unlock()
}

// skillsFingerprint is the bound skill set's revision identity (spec 077
// FR-006, the charterFingerprint shape): a short content hash (12 hex chars
// of SHA-256) over EXACTLY the composed set — names and post-cap texts in
// composition order — so the recorded observation can never name a skill set
// the guardian never ran under. Name and text are NUL-delimited so no
// rename/edit pair can collide with a different set.
func skillsFingerprint(skills []skillFile) string {
	h := sha256.New()
	for _, sf := range skills {
		h.Write([]byte(sf.name))
		h.Write([]byte{0})
		h.Write([]byte(sf.text))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil)[:6])
}

// observeSkills lands the skills observation (spec 077 FR-006) — the
// observeCharter twin, stamped at the same point of every turn: when the
// BOUND skill set's fingerprint differs from the last recorded one
// (State.SkillsFingerprint, via the absorb-side mirror), the turn emits
// guardian.skills_observed through the same InjectSocial door. An EMPTY
// bound set never emits — absence is not an observation, and stages 1–2
// (stageSkills refuses to bind there) are structurally silent, which is what
// makes SkillsObservedEvidence's Custom-by-construction claim honest: a
// recorded observation always names a player-authored, stage-3+ set. Ended
// worlds skip, exactly like the charter observation.
func (mt *Guardian) observeSkills(skills []skillFile) {
	if len(skills) == 0 {
		return
	}
	fp := skillsFingerprint(skills)
	mt.stateMu.Lock()
	known, ended := mt.skillsFP, mt.ended
	mt.stateMu.Unlock()
	if ended || fp == known {
		return
	}
	names := make([]string, len(skills))
	for i, sf := range skills {
		names[i] = sf.name
	}
	batch := []store.Event{{Type: "guardian.skills_observed", Payload: mustJSON(sim.SkillsObservedPayload{
		Fingerprint: fp, Names: names})}}
	if err := mt.social.InjectSocial(batch); err != nil {
		log.Printf("guardian: skills observation rejected at the door: %v", err)
		return
	}
	// Optimistic mirror update — the observeCharter idiom: a back-to-back
	// turn cannot double-emit before the absorb goroutine reflects the
	// landed event; mirrorState only ever moves the mirror forward.
	mt.stateMu.Lock()
	mt.skillsFP = fp
	mt.stateMu.Unlock()
}

// maxSkillFiles is the number of skill files composed into a single turn, the
// file-count cap half of the skills surface (per-file size reuses the charter's
// 4,000-char cap, persona.CharterMaxChars). Surplus files (in sort order) are
// skipped with a notice — deterministic, never silent (spec 021 FR-002).
const maxSkillFiles = 8

// skillFile is one composed skill: its filename (provenance) and its effective
// (post-cap) text.
type skillFile struct {
	name string
	text string
}

// loadSkills reads the world's skills/ folder fresh (FR-001) and returns the
// eligible skill files in deterministic composition order, plus one notice per
// issue. Eligibility (contracts/instruction-surface.md rule 3): regular .md
// files that are direct children of skills/ — subdirectories, dotfiles, and
// other extensions are silently excluded (.DS_Store noise is not a notice).
// Composition order is ascending bytewise filename order (players prefix 10-,
// 20-), the same order two byte-identical world dirs produce, so identical dirs
// compose identical prompts (FR-012). Caps mirror the charter's discipline:
// over-cap files truncate + notice; files beyond the count cap skip + notice;
// an unreadable file skips + notice. A missing/unreadable folder is the common,
// unremarkable case — no skills, no notice.
func loadSkills(worldDir string) (skills []skillFile, notices []string) {
	dir := filepath.Join(worldDir, "skills")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue // no recursion (rule 3)
		}
		n := e.Name()
		if strings.HasPrefix(n, ".") || !strings.HasSuffix(n, ".md") {
			continue // dotfiles / non-.md — silently excluded, not a notice
		}
		names = append(names, n)
	}
	sort.Strings(names) // ascending bytewise — the deterministic composition order
	if len(names) > maxSkillFiles {
		skipped := make([]string, 0, len(names)-maxSkillFiles)
		for _, n := range names[maxSkillFiles:] {
			skipped = append(skipped, "skills/"+n)
		}
		names = names[:maxSkillFiles]
		notices = append(notices, fmt.Sprintf("more than %d skill files present — %s not composed",
			maxSkillFiles, strings.Join(skipped, ", ")))
	}
	for _, n := range names {
		data, err := os.ReadFile(filepath.Join(dir, n))
		if err != nil {
			notices = append(notices, fmt.Sprintf("skills/%s could not be read — skipped", n))
			continue
		}
		text := string(data)
		if len(text) > persona.CharterMaxChars {
			text = text[:persona.CharterMaxChars]
			notices = append(notices, fmt.Sprintf("skills/%s exceeds the cap — only the first 4,000 characters are in effect", n))
		}
		skills = append(skills, skillFile{name: n, text: text})
	}
	return skills, notices
}

// skillNames returns just the composition-ordered filenames of the effective
// skill files — the model-free provenance list Status surfaces (spec 021 R8).
func skillNames(worldDir string) []string {
	skills, _ := loadSkills(worldDir)
	if len(skills) == 0 {
		return nil
	}
	out := make([]string, len(skills))
	for i, s := range skills {
		out[i] = s.name
	}
	return out
}

// grantSet is a world's effective capability grant for one turn/peek (spec 021
// US2): which guardian loop tools are offered, and (when restricted) which
// miracle kinds. It drives all three gating layers alike — the declared roster,
// the derived guidance, and the door — so they cannot disagree (FR-005). Maps
// are used for O(1) membership; nothing is ever iterated into ordered output.
type grantSet struct {
	tools           map[string]bool // granted guardian loop-tool names
	kinds           map[string]bool // granted miracle kinds; meaningful only when restricted
	kindsRestricted bool            // true ⇒ only kinds in `kinds` are offered for work_miracle
	manifestDefault bool            // true ⇒ no capabilities.json on disk (full default grant)
	// Bundle-tool grant (spec 036 T014). `tools` is confined to KNOWN built-in
	// names (loadManifest drops unknowns), so it can never carry a bundle name;
	// bundle tools are gated by the RAW request instead. toolsConstrained is true
	// only when an explicit "tools" list is present — then a bundle tool is
	// granted iff bundleAllowed names it. When false (no file, or the "tools" key
	// omitted) every bundle tool is granted, mirroring the built-in default.
	toolsConstrained bool
	bundleAllowed    map[string]bool
}

// allows reports whether a guardian loop tool is granted this world.
func (g grantSet) allows(name string) bool { return g.tools[name] }

// allowsBundle reports whether a bundle tool is granted this world (spec 036
// T014): the world-level capabilities.json "tools" list applies to bundle tools
// the same way it does to built-ins — absent file or omitted key ⇒ granted, an
// explicit list ⇒ the tool must be named.
func (g grantSet) allowsBundle(name string) bool {
	if !g.toolsConstrained {
		return true
	}
	return g.bundleAllowed[name]
}

// allowsKind reports whether a miracle kind may land: unrestricted worlds allow
// every kind; a restricted world allows only its declared subset.
func (g grantSet) allowsKind(kind string) bool {
	if !g.kindsRestricted {
		return true
	}
	return g.kinds[kind]
}

// grantedTools returns the granted guardian loop-tool names in registry order
// (LoopRosterGuardian order) — the deterministic surface Status renders.
func (g grantSet) grantedTools() []string {
	var out []string
	for _, t := range tool.LoopRosterGuardian() {
		if g.tools[t.Name] {
			out = append(out, t.Name)
		}
	}
	return out
}

// grantedKinds returns the granted miracle kinds in registry order when
// restricted, else nil (all kinds). Deterministic — walks tool.MiracleKinds().
func (g grantSet) grantedKinds() []string {
	if !g.kindsRestricted {
		return nil
	}
	var out []string
	for _, k := range tool.MiracleKinds() {
		if g.kinds[k] {
			out = append(out, k)
		}
	}
	return out
}

// grantedToolLabels renders the granted roster for Status (contracts/status.md):
// registry order, with work_miracle suffixed `(kind,…)` ONLY when its kinds are
// restricted (an unrestricted work_miracle shows bare). nil when nothing is
// granted (a conversation-only world) so the field omits under omitempty.
func grantedToolLabels(g grantSet) []string {
	names := g.grantedTools()
	if len(names) == 0 {
		return nil
	}
	out := make([]string, 0, len(names))
	for _, n := range names {
		if n == "work_miracle" && g.kindsRestricted {
			out = append(out, "work_miracle("+strings.Join(g.grantedKinds(), ",")+")")
			continue
		}
		out = append(out, n)
	}
	return out
}

// fullGrant is the default grant a world with no (or an unusable) manifest gets:
// the entire guardian loop roster, all miracle kinds — byte-compatible with the
// pre-021 guardian (FR-007, SC-003).
func fullGrant() grantSet {
	tools := make(map[string]bool)
	for _, t := range tool.LoopRosterGuardian() {
		tools[t.Name] = true
	}
	return grantSet{tools: tools}
}

// manifestDoc is the parse target for capabilities.json. Absent vs empty is
// meaningful and preserved by encoding/json: an omitted key decodes to nil, an
// explicit [] to a non-nil empty slice.
type manifestDoc struct {
	Tools        []string `json:"tools"`
	MiracleKinds []string `json:"miracle_kinds"`
}

// loadManifest reads the world's capabilities.json fresh (FR-001) and returns
// the effective grant set plus one notice per issue, mirroring the charter's
// permissive-fallback teaching model (spec 021 R4, contracts/capability-manifest.md):
//   - no file → full grant, NO notice, manifestDefault true (byte-compatible today);
//   - unreadable / malformed JSON → full grant + notice (a typo never bricks the guardian);
//   - unknown tool/kind names → ignored + notice, the valid remainder applies;
//   - omitted "tools" key → unconstrained (all tools), symmetric with an omitted
//     "miracle_kinds" meaning all kinds; explicit "tools": [] → conversation-only;
//   - "miracle_kinds" omitted → all kinds; present → exactly that (valid) subset.
//
// Conversation is never gateable here — it is the final-text channel, not a
// roster tool (FR-006); a world granting nothing still converses.
//
// knownBundleTools is the boot-frozen bundle roster's tool names (spec 036 T014,
// handoff fix T030): an explicit "tools" list naming a REAL bundle tool must not
// render as an "unknown tool … ignored" notice — allowsBundle already grants it
// correctly, so the notice was cosmetic noise. Variadic so every pre-036 call
// site (and every direct-arg test) keeps compiling unchanged; the turn/status
// call sites pass the bundle roster's names.
func loadManifest(worldDir string, knownBundleTools ...string) (grantSet, []string) {
	path := filepath.Join(worldDir, "capabilities.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		g := fullGrant()
		g.manifestDefault = true
		return g, nil
	}
	if err != nil {
		return fullGrant(), []string{"capabilities.json could not be read — serving with the full tool roster"}
	}
	var doc manifestDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		return fullGrant(), []string{"capabilities.json is not valid JSON — serving with the full tool roster"}
	}

	var notices []string
	known := make(map[string]bool)
	var order []string
	for _, t := range tool.LoopRosterGuardian() {
		known[t.Name] = true
		order = append(order, t.Name)
	}

	tools := make(map[string]bool)
	var bundleAllowed map[string]bool
	toolsConstrained := false
	if doc.Tools == nil {
		for _, n := range order { // omitted key ⇒ unconstrained
			tools[n] = true
		}
	} else {
		// An explicit list constrains BOTH built-ins and bundle tools. Built-in
		// names filter through `known` (an unknown built-in name is a noticed typo);
		// the raw list gates bundle tools (spec 036 T014), which loadManifest cannot
		// classify at this layer (it has no bundle roster), so a bundle name is not
		// reported as an unknown tool here — allowsBundle honors it.
		toolsConstrained = true
		bundleAllowed = make(map[string]bool, len(doc.Tools))
		knownBundle := make(map[string]bool, len(knownBundleTools))
		for _, n := range knownBundleTools {
			knownBundle[n] = true
		}
		var unknown []string
		for _, n := range doc.Tools {
			bundleAllowed[n] = true
			switch {
			case known[n]:
				tools[n] = true
			case knownBundle[n]:
				// A real bundle tool name — allowsBundle already honors it; not a typo.
			default:
				unknown = append(unknown, n)
			}
		}
		if len(unknown) > 0 {
			notices = append(notices, "capabilities.json lists unknown tool(s): "+strings.Join(unknown, ", ")+" — ignored")
		}
	}

	g := grantSet{tools: tools, toolsConstrained: toolsConstrained, bundleAllowed: bundleAllowed}
	if doc.MiracleKinds != nil {
		knownKind := make(map[string]bool)
		for _, k := range tool.MiracleKinds() {
			knownKind[k] = true
		}
		kinds := make(map[string]bool)
		var unknown []string
		for _, k := range doc.MiracleKinds {
			if knownKind[k] {
				kinds[k] = true
			} else {
				unknown = append(unknown, k)
			}
		}
		if len(unknown) > 0 {
			notices = append(notices, "capabilities.json lists unknown miracle_kinds entries: "+strings.Join(unknown, ", ")+" — ignored")
		}
		g.kinds = kinds
		g.kindsRestricted = true
	}
	return g, notices
}

// grantedRoster is the effective guardian loop roster for a turn: the full loop
// roster filtered to granted tools, with work_miracle's kind enum narrowed to
// the granted kinds when restricted (copy-on-write via tool.RestrictEnum, the
// registry untouched). This ONE roster feeds all three gating layers — Job.Roster
// (declaration), GuardianToolGuidance (prose), and the handler set (door) — so an
// ungranted tool or kind is structurally absent from every one of them (FR-005).
func grantedRoster(g grantSet) []tool.Tool {
	full := tool.LoopRosterGuardian()
	out := make([]tool.Tool, 0, len(full))
	for _, t := range full {
		if !g.tools[t.Name] {
			continue
		}
		if t.Name == "work_miracle" && g.kindsRestricted {
			t = tool.RestrictEnum(t, "kind", g.grantedKinds())
		}
		out = append(out, t)
	}
	return out
}

// narrowGrantForBundles applies every installed bundle's optional
// capabilities.json as an INTERSECTION over the world-level grant (spec 036
// US4 T030, data-model.md GrantNarrowing): a persona bundle can only narrow
// what the world already grants — never widen it. Reuses grantSet's own
// semantics (this file, above) rather than reinventing them: the same "tools"/
// "miracle_kinds" schema, the same omitted-key-means-unconstrained reading. A
// bundle with no capabilities.json (Grant == nil) contributes nothing; with no
// BundleSet at all, g is returned unchanged. Multiple narrowing bundles compose
// by intersection, which is commutative — load order does not matter.
func narrowGrantForBundles(g grantSet, bs *bundle.BundleSet) grantSet {
	if bs == nil {
		return g
	}
	for _, b := range bs.Bundles() {
		if b.Grant != nil {
			g = intersectGrant(g, b.Grant)
		}
	}
	return g
}

// intersectGrant narrows g by one bundle's grant doc. An explicit "tools" list
// keeps only names already granted AND named by gd — the SAME flat namespace
// loadManifest's "tools" key already spans (built-in loop-tool names via
// g.tools, bundle tool names via g.bundleAllowed), so a persona can narrow
// either surface with one list. An explicit "miracle_kinds" list keeps only
// kinds already granted AND named when the world was already restricted;
// when the world was unrestricted ("all kinds"), gd's list becomes the new
// ceiling (still never wider than the world's real vocabulary — grantedKinds/
// RestrictEnum only ever consult tool.MiracleKinds() membership, so a garbage
// persona-declared kind name is simply inert). Omitted keys narrow nothing.
func intersectGrant(g grantSet, gd *bundle.GrantDoc) grantSet {
	if gd.Tools != nil {
		named := make(map[string]bool, len(gd.Tools))
		for _, n := range gd.Tools {
			named[n] = true
		}
		newTools := make(map[string]bool, len(g.tools))
		for n := range g.tools {
			if named[n] {
				newTools[n] = true
			}
		}
		newBundleAllowed := make(map[string]bool, len(named))
		if g.toolsConstrained {
			for n := range g.bundleAllowed {
				if named[n] {
					newBundleAllowed[n] = true
				}
			}
		} else {
			for n := range named {
				newBundleAllowed[n] = true
			}
		}
		g.tools, g.bundleAllowed, g.toolsConstrained = newTools, newBundleAllowed, true
	}
	if gd.MiracleKinds != nil {
		named := make(map[string]bool, len(gd.MiracleKinds))
		for _, k := range gd.MiracleKinds {
			named[k] = true
		}
		newKinds := make(map[string]bool, len(named))
		if g.kindsRestricted {
			for k := range g.kinds {
				if named[k] {
					newKinds[k] = true
				}
			}
		} else {
			for k := range named {
				newKinds[k] = true
			}
		}
		g.kinds, g.kindsRestricted = newKinds, true
	}
	return g
}

// The curriculum-ladder stage ceiling (spec 046 US2, contracts/stage-gating.md).
//
// stage1CeilingTools is the pinned stage-1 tool ceiling — the honest "base
// conversational + basic nudge + the watch primitive" subset of the live loop
// roster (tool.LoopRosterGuardian), recorded in the contract in-PR. Conversation
// is never a roster tool (it is the reply channel and never gateable); the
// basic nudges send_vision/send_omen ARE the stage-1 grant. RATIFIED AMENDMENT
// (TASK-119 board artifact, "first-night teaches visions+orders", applied to
// spec 046 stage-gating.md in-branch): monitor_and_act/cancel_order — standing
// orders — join the stage-1 ceiling too, because the first-night exercise
// (contracts/exercises.md) teaches the watch as the stage-1 primitive
// alongside visions/omens; a daytime omen's system-origin nightfall deferral
// already carried send_omen's gate regardless (orders.go), so this only adds
// the PLAYER-placed watch. Still beyond the stage: work_miracle
// (world-shaping) and pause/start/adjust_speed (clock control; neither query
// nor nudge — the player keeps direct CLI/TUI clock control at every stage).
// No bundle tools (the empty-intersection effect of the explicit list below).
// explain (spec 063) joins the stage-1 ceiling: it is the tutor preset's own
// grounding tool — stage-1 IS the orientation stage, and the guide's
// mechanics-via-explain contract (persona.TutorGuide) needs the tool granted
// where the guide composes. Read-only, zero-cost, tutor-lane by construction,
// so it widens no acting capability.
// The plan layer (spec 084) joins the stage-1 ceiling too: designations,
// directives, and survey are the plan-loop teaching primitives — granted at
// every stage, the monitor_and_act precedent (spec 084 Assumptions, flagged
// for operator review there). All five are charge-free; none is world-shaping
// (villagers still do all the work by their own logic), so the stage-1
// "no miracles, no clock control" posture is unchanged.
// prophesy (spec 085) joins too, following send_vision's stage profile
// (granted where visions are granted — it is the same influence verb with a
// wager attached; spec 085 Assumptions, flagged for operator review there):
// charge-priced like the nudges, not world-shaping, so the stage-1 posture
// is still unchanged.
// brief_myths (spec 101) joins the ceiling too, following survey_site/
// explain's stage profile: a read-only lookup over the existing belief
// corpus, never world-shaping, so it widens no acting capability.
// canonize_region is DELIBERATELY excluded — a world-shaping act, the
// work_miracle precedent (no miracle kind is granted at stage-1/2), not the
// charge-free plan layer's every-stage grant.
// The mission layer (spec 107) joins the stage-1 ceiling too, following the
// plan layer's every-stage profile (flagged for operator review, the spec
// 084/085 precedent): all three verbs are charge-free artifact bookkeeping —
// a mission only records the player's standing instruction and its derived
// progress; the acting still happens through the already-gated verbs, so
// the stage-1 "no miracles, no clock control" posture is unchanged.
// inspect_pack (spec 116 FR-015) joins the ceiling too, following the
// survey_site / brief_myths / explain read profile exactly: charge-free,
// event-free, never the turn's act, and it widens no ACTING capability —
// it only lets the guardian see what a villager already carries, which the
// targeting digest half-told it at every stage already. take_item is
// DELIBERATELY absent (it is a work_miracle kind, and no miracle kind is
// granted at stage-1/2 — the canonize_region precedent above): the removal
// is world-shaping, the read is not.
var stage1CeilingTools = []string{"send_omen", "send_vision", "monitor_and_act", "cancel_order", "explain",
	"place_designation", "cancel_designation", "issue_directive", "cancel_directive", "survey_site", "prophesy", "brief_myths",
	"accept_mission", "note_mission_progress", "cancel_mission", "inspect_pack"}

// stageCeiling returns the stage's capability ceiling as a narrowing doc —
// the same shape a persona bundle's grant uses, so intersectGrant applies it
// with identical semantics — or nil when the stage imposes none. The ladder:
//
//	stage-1  send_omen + send_vision + monitor_and_act + cancel_order; no
//	         miracle kinds; no bundles (ratified amendment: standing orders are
//	         the watch primitive first-night teaches)
//	stage-2  identical to stage-1 — the unlock is the instruction surface
//	         (charter binds), not new tools
//	stage-3  no ceiling: skill files compose and the grantable manifest opens
//	         (full spec-021/036 behavior)
//	stage-4  no ceiling: full roster incl. capstone capabilities
//	""       no ceiling: a pre-ladder world is ungated (stage-4 semantics)
func stageCeiling(stage string) *bundle.GrantDoc {
	switch stage {
	case "stage-1", "stage-2":
		return &bundle.GrantDoc{Tools: stage1CeilingTools, MiracleKinds: []string{}}
	}
	return nil
}

// StageCeilingVerbs returns the guardian loop-tool names the stage's ceiling
// grants, in registry order — the full loop roster for a ceiling-less stage
// or a pre-ladder world (spec 063 US5/D9): the help overlay's guardian
// section renders exactly this static per-stage set, derived through the
// SAME intersection the turn's grant runs (applyStageCeiling over the full
// default grant), so the taught verb list can never drift from the ceiling
// the door enforces. Deterministic: grantedTools walks the loop roster.
func StageCeilingVerbs(stage string) []string {
	return applyStageCeiling(fullGrant(), stage).grantedTools()
}

// applyStageCeiling intersects the stage ceiling into the world-level grant
// (spec 046 FR-004, the narrowGrantForBundles idiom): intersection-only, so a
// player's capabilities.json may narrow WITHIN the ceiling but never exceed
// it, and beyond-stage capabilities are structurally absent from every layer
// derived from the grant (declaration, prose, door). Applied at every manifest
// load site (turn + status) immediately after loadManifest, BEFORE
// grantedRoster — the three-layer coherence is inherited, not re-implemented.
// A ceiling-less stage returns g unchanged (identity).
func applyStageCeiling(g grantSet, stage string) grantSet {
	if c := stageCeiling(stage); c != nil {
		g = intersectGrant(g, c)
	}
	return g
}

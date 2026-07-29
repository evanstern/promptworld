package bundle

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/evanstern/promptworld/internal/sim"
)

// The boot validation ladder (contracts/boot-validation.md). B1–B4 are
// bundle-level: a failure rejects the WHOLE bundle (a broken name, identity,
// permissions, or a structural cap breach). T1–T7 are per-tool: a failure skips
// that ONE tool, leaving its siblings and the bundle's SOUL/grant intact
// (clarification #1). Every failure returns a BootIssue naming the file and the
// specific problem (SC-005). Cross-bundle collisions (C1/C2) are load.go's.

const (
	soulMaxChars      = 4000 // mirrors persona.CharterMaxChars (SOUL is a charter fragment)
	maxToolsPerBundle = 16   // mirrors the maxSkillFiles discipline (structural cap)
)

// bundleNameRE is the bundle identifier shape (B1); it is looser than a tool
// name (hyphens allowed, longer) because a bundle is a folder, not a tool.
var bundleNameRE = regexp.MustCompile(`^[a-z0-9_-]{1,64}$`)

// validateBundle runs B1–B4 and, on a structurally sound bundle, T1–T7 per tool.
// It returns the bundle with its T-passing tools (collisions NOT yet resolved)
// and every BootIssue. A B1–B4 failure returns (nil, rejecting issues); the
// whole bundle is dropped.
func validateBundle(bundlesRoot, bundleDir string) (*Bundle, []BootIssue) {
	name := filepath.Base(bundleDir)

	// B1: bundle name shape.
	if !bundleNameRE.MatchString(name) {
		return nil, []BootIssue{{
			Bundle: name, File: relTo(bundlesRoot, bundleDir), Rule: "B1", Severity: "error",
			Message: fmt.Sprintf("bundle name %q must match [a-z0-9_-]{1,64} — bundle rejected", name),
		}}
	}

	b := &Bundle{Name: name}
	var issues []BootIssue

	// B2: SOUL.md (optional) valid UTF-8 within the char cap — a broken identity
	// rejects the whole bundle.
	soulPath := filepath.Join(bundleDir, "SOUL.md")
	if data, err := os.ReadFile(soulPath); err == nil {
		if !utf8.Valid(data) {
			return nil, []BootIssue{{
				Bundle: name, File: relTo(bundlesRoot, soulPath), Rule: "B2", Severity: "error",
				Message: "SOUL.md is not valid UTF-8 — bundle rejected",
			}}
		}
		if n := utf8.RuneCount(data); n > soulMaxChars {
			return nil, []BootIssue{{
				Bundle: name, File: relTo(bundlesRoot, soulPath), Rule: "B2", Severity: "error",
				Message: fmt.Sprintf("SOUL.md is %d characters, over the %d cap — bundle rejected", n, soulMaxChars),
			}}
		}
		b.Soul = string(data)
	} else if !os.IsNotExist(err) {
		return nil, []BootIssue{{
			Bundle: name, File: relTo(bundlesRoot, soulPath), Rule: "B2", Severity: "error",
			Message: fmt.Sprintf("SOUL.md could not be read (%v) — bundle rejected", err),
		}}
	}

	// B3: capabilities.json (optional) parses per the grant schema — broken
	// permissions reject the whole bundle.
	capPath := filepath.Join(bundleDir, "capabilities.json")
	if data, err := os.ReadFile(capPath); err == nil {
		doc, perr := parseGrant(data)
		if perr != nil {
			return nil, []BootIssue{{
				Bundle: name, File: relTo(bundlesRoot, capPath), Rule: "B3", Severity: "error",
				Message: fmt.Sprintf("capabilities.json %v — bundle rejected", perr),
			}}
		}
		b.Grant = doc
	} else if !os.IsNotExist(err) {
		return nil, []BootIssue{{
			Bundle: name, File: relTo(bundlesRoot, capPath), Rule: "B3", Severity: "error",
			Message: fmt.Sprintf("capabilities.json could not be read (%v) — bundle rejected", err),
		}}
	}

	// B4: ≤16 tool dirs — a cap breach is structural, so the whole bundle is
	// rejected rather than silently truncated.
	toolsRoot := filepath.Join(bundleDir, "tools")
	toolDirs, err := childDirs(toolsRoot)
	if err != nil {
		return nil, []BootIssue{{
			Bundle: name, File: relTo(bundlesRoot, toolsRoot), Rule: "B4", Severity: "error",
			Message: fmt.Sprintf("tools/ could not be read (%v) — bundle rejected", err),
		}}
	}
	if len(toolDirs) > maxToolsPerBundle {
		return nil, []BootIssue{{
			Bundle: name, File: relTo(bundlesRoot, toolsRoot), Rule: "B4", Severity: "error",
			Message: fmt.Sprintf("bundle declares %d tools, over the %d cap — bundle rejected", len(toolDirs), maxToolsPerBundle),
		}}
	}

	// T1–T7 per tool.
	for _, td := range toolDirs {
		bt, issue := validateTool(bundlesRoot, name, filepath.Join(toolsRoot, td))
		if issue != nil {
			issues = append(issues, *issue)
			continue // tool skipped; siblings, SOUL, and grant stay intact
		}
		b.Tools = append(b.Tools, *bt)
	}
	return b, issues
}

// validateTool runs T1–T7 for one tool dir, returning the validated BundleTool
// or the single BootIssue that skips it (first failure, one clear reason).
func validateTool(bundlesRoot, bundleName, toolDir string) (*BundleTool, *BootIssue) {
	toolName := filepath.Base(toolDir)
	jsonPath := filepath.Join(toolDir, "tool.json")
	rel := relTo(bundlesRoot, jsonPath)
	fail := func(rule, format string, a ...any) *BootIssue {
		return &BootIssue{Bundle: bundleName, Tool: toolName, File: rel, Rule: rule, Severity: "error", Message: fmt.Sprintf(format, a...)}
	}

	// T1: tool.json present, strict-decodes, name == folder (+ T2/T4/T7 in Parse).
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fail("T1", "tool.json is missing or unreadable: %v", err)
	}
	m, perr := Parse(data, toolName)
	if perr != nil {
		return nil, &BootIssue{Bundle: bundleName, Tool: toolName, File: rel, Rule: ruleOf(perr), Severity: "error", Message: perr.Error()}
	}

	// T3: events non-empty and ⊆ the sim injection whitelist. Declared names
	// normalize through the log-format rename table first (spec 094 D4's
	// config-reference posture): a bundle authored against the pre-094
	// metatron.* vocabulary keeps loading on a migrated world — the manifest
	// is a REFERENCE to event types, not the log itself, so read-side
	// normalization here is the same kindness EvidenceRef readers get.
	if len(m.Events) == 0 {
		return nil, fail("T3", "tool %q declares no events", toolName)
	}
	for i, ev := range m.Events {
		m.Events[i] = sim.CanonicalEventType(ev)
	}
	for _, ev := range m.Events {
		if !sim.InjectableSocialEvent(ev) {
			return nil, fail("T3", "tool %q declares event %q, which is not an injectable event type", toolName, ev)
		}
	}

	var templates []effectTemplate
	var script *scriptProgram
	if m.scriptMode() {
		// T6: script exists, parses, defines apply(), and compiles cleanly — the
		// compiled program is retained on the tool so no invocation ever re-parses
		// (spec 036 US3, T024). Init runs the top level under the step ceiling, so a
		// hostile top level cannot burn unbounded work at boot. The limits ceiling is
		// already enforced in Parse (T7); maxSteps() resolves the per-invocation cap.
		sp, err := compileScript(toolDir, m.Script)
		if err != nil {
			return nil, fail("T6", "tool %q %v", toolName, err)
		}
		script = sp
	} else {
		// T5: every effect template well-formed; producible events == declared.
		ts, err := ParseTemplates(m.Effects)
		if err != nil {
			return nil, &BootIssue{Bundle: bundleName, Tool: toolName, File: rel, Rule: "T5", Severity: "error",
				Message: fmt.Sprintf("tool %q %v", toolName, err)}
		}
		declared := make(map[string]bool, len(m.Events))
		for _, ev := range m.Events {
			declared[ev] = true
		}
		producible := producibleEvents(ts)
		if !sameEventSet(producible, declared) {
			return nil, fail("T5", "tool %q declares events %s but its effects produce %s (they must match exactly)",
				toolName, sortedKeys(declared), sortedKeys(producible))
		}
		templates = ts
	}

	return &BundleTool{Name: toolName, Tool: m.synthesize(), Manifest: m, Templates: templates, Script: script, Dir: toolDir}, nil
}

// parseGrant strict-decodes a capabilities.json into a GrantDoc (B3).
func parseGrant(data []byte) (*GrantDoc, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var d GrantDoc
	if err := dec.Decode(&d); err != nil {
		return nil, fmt.Errorf("does not parse: %v", err)
	}
	if dec.More() {
		return nil, fmt.Errorf("has trailing content after the JSON object")
	}
	return &d, nil
}

// ruleOf extracts the ladder rule id from a ruleError, defaulting to T1 (the
// decode/name checks) for a plain error.
func ruleOf(err error) string {
	var re ruleError
	if errors.As(err, &re) {
		return re.rule
	}
	return "T1"
}

// sameEventSet reports whether two event-type sets have identical members.
func sameEventSet(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}

// sortedKeys renders a set as a deterministic {a, b} string for error messages.
func sortedKeys(m map[string]bool) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return "{" + strings.Join(keys, ", ") + "}"
}

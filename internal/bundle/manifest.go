package bundle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"unicode/utf8"

	"github.com/evanstern/promptworld/internal/tool"
)

// The tool.json manifest (contracts/bundle-manifest.md): the authoring surface
// for a bundle tool. Parsing is strict — unknown keys are rejected so a typo
// fails loudly at boot rather than silently dropping a field — and the scalar
// rules mirror internal/tool/validate.go so a synthesized tool.Tool obeys the
// same invariants the compile-time registry does. The cross-cutting checks
// (events ⊆ the sim whitelist, producible-event agreement, script parse) live
// in validate.go, which needs sim/effects/starlark; manifest.go stays confined
// to what a single manifest can prove about itself.

// Manifest caps and defaults. "chars" is measured in runes (a manifest is
// authored text); the narrate byte cap lives in effects.go.
const (
	descMaxChars      = 500     // tool description (LLM-facing PromptGloss)
	paramDescMaxChars = 200     // a param's optional gloss
	maxParams         = 8       // per tool
	maxEnumValues     = 16      // per enum param
	defaultMaxSteps   = 100_000 // script step budget when limits omitted
	maxStepsCeiling   = 1_000_000
)

// toolNameRE and paramNameRE mirror the identifier shapes the manifest contract
// fixes; a manifest name must additionally equal its folder basename (T1).
var (
	toolNameRE  = regexp.MustCompile(`^[a-z0-9_]{1,48}$`)
	paramNameRE = regexp.MustCompile(`^[a-z0-9_]{1,32}$`)
)

// Manifest is the parsed, strict-decoded tool.json. Effects is kept raw so the
// effect compiler (effects.go) owns the closed vocabulary; Limits is a pointer
// so an absent block is distinguishable from an explicit zero (which is invalid).
type Manifest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Params      []ManifestParam `json:"params"`
	Events      []string        `json:"events"`
	Charges     int             `json:"charges"`
	Effects     json.RawMessage `json:"effects"`
	Script      string          `json:"script"`
	Limits      *Limits         `json:"limits"`
}

// ManifestParam is one authored argument; it maps 1:1 onto a tool.Param.
type ManifestParam struct {
	Name        string   `json:"name"`
	Kind        string   `json:"kind"`
	Required    bool     `json:"required"`
	Description string   `json:"description"`
	Enum        []string `json:"enum"`
	Min         int      `json:"min"`
	Max         int      `json:"max"`
	MaxBytes    int      `json:"max_bytes"`
}

// Limits carries the script step budget (script mode only).
type Limits struct {
	MaxSteps int `json:"max_steps"`
}

// ruleError tags a validation failure with the boot-validation ladder rule id
// (contracts/boot-validation.md) so the loader can attribute a BootReport entry
// without re-deriving which check failed. The message is the human sentence the
// report shows.
type ruleError struct {
	rule string
	msg  string
}

func (e ruleError) Error() string { return e.msg }

func ruleErr(rule, format string, a ...any) ruleError {
	return ruleError{rule: rule, msg: fmt.Sprintf(format, a...)}
}

// scriptMode reports whether the manifest ships a script (vs declarative
// effects). Only meaningful after Parse has confirmed exactly one is present.
func (m *Manifest) scriptMode() bool { return m.Script != "" }

// maxSteps is the script step budget the executor caps each invocation at: the
// manifest's limits.max_steps when set (Parse already bounded it to (0, ceiling]),
// otherwise the default. Script mode only.
func (m *Manifest) maxSteps() uint64 {
	if m.Limits != nil && m.Limits.MaxSteps > 0 {
		return uint64(m.Limits.MaxSteps)
	}
	return defaultMaxSteps
}

// Parse strict-decodes and validates a tool.json against the manifest contract,
// returning the parsed manifest. folderName is the tool-dir basename the
// manifest's name must equal (T1). All failures are ruleError-tagged
// (T1/T2/T4/T7) so the loader can stamp the right rule id on the BootReport;
// validation stops at the first failure (a skipped tool needs one clear reason).
func Parse(data []byte, folderName string) (*Manifest, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var m Manifest
	if err := dec.Decode(&m); err != nil {
		return nil, ruleErr("T1", "tool.json does not decode: %v", err)
	}
	if dec.More() {
		return nil, ruleErr("T1", "tool.json has trailing content after the JSON object")
	}

	// T1: name shape and folder agreement.
	if !toolNameRE.MatchString(m.Name) {
		return nil, ruleErr("T1", "tool name %q must match [a-z0-9_]{1,48}", m.Name)
	}
	if m.Name != folderName {
		return nil, ruleErr("T1", "tool name %q does not match its folder name %q", m.Name, folderName)
	}

	// T1: description present and within the cap.
	dl := utf8.RuneCountInString(m.Description)
	if dl < 1 || dl > descMaxChars {
		return nil, ruleErr("T1", "tool %q description must be 1–%d characters (got %d)", m.Name, descMaxChars, dl)
	}

	// T2: params well-formed (mirrors internal/tool/validate.go).
	if len(m.Params) > maxParams {
		return nil, ruleErr("T2", "tool %q declares %d params (max %d)", m.Name, len(m.Params), maxParams)
	}
	seen := make(map[string]bool, len(m.Params))
	for _, p := range m.Params {
		if err := validateParam(m.Name, p, seen); err != nil {
			return nil, err
		}
	}

	// T7: charge and step bounds.
	if m.Charges < 0 {
		return nil, ruleErr("T7", "tool %q charges must be ≥0 (got %d)", m.Name, m.Charges)
	}
	if m.Limits != nil {
		if m.Limits.MaxSteps <= 0 || m.Limits.MaxSteps > maxStepsCeiling {
			return nil, ruleErr("T7", "tool %q limits.max_steps must be in (0, %d] (got %d)", m.Name, maxStepsCeiling, m.Limits.MaxSteps)
		}
	}

	// T4: exactly one of effects | script. An absent effects key decodes to a
	// nil RawMessage; an empty script string means absent.
	hasEffects := len(bytes.TrimSpace(m.Effects)) > 0 && !bytes.Equal(bytes.TrimSpace(m.Effects), []byte("null"))
	hasScript := m.scriptMode()
	switch {
	case hasEffects && hasScript:
		return nil, ruleErr("T4", "tool %q declares both effects and script (exactly one is required)", m.Name)
	case !hasEffects && !hasScript:
		return nil, ruleErr("T4", "tool %q declares neither effects nor script (exactly one is required)", m.Name)
	}

	return &m, nil
}

// validateParam enforces the per-param rules the contract fixes and
// internal/tool/validate.go mirrors: name shape+uniqueness, a known kind,
// enum values iff enum, non-inverted number bounds, and the cap-only fields
// belonging to their kind.
func validateParam(toolName string, p ManifestParam, seen map[string]bool) error {
	if !paramNameRE.MatchString(p.Name) {
		return ruleErr("T2", "tool %q param name %q must match [a-z0-9_]{1,32}", toolName, p.Name)
	}
	if seen[p.Name] {
		return ruleErr("T2", "tool %q has a duplicate param name %q", toolName, p.Name)
	}
	seen[p.Name] = true
	if dl := utf8.RuneCountInString(p.Description); dl > paramDescMaxChars {
		return ruleErr("T2", "tool %q param %q description exceeds %d characters (got %d)", toolName, p.Name, paramDescMaxChars, dl)
	}
	switch p.Kind {
	case "agent_name":
	case "text":
		if p.MaxBytes < 0 {
			return ruleErr("T2", "tool %q text param %q max_bytes must be ≥0 (got %d)", toolName, p.Name, p.MaxBytes)
		}
	case "enum":
		if len(p.Enum) < 1 || len(p.Enum) > maxEnumValues {
			return ruleErr("T2", "tool %q enum param %q must list 1–%d values (got %d)", toolName, p.Name, maxEnumValues, len(p.Enum))
		}
		for _, v := range p.Enum {
			if v == "" {
				return ruleErr("T2", "tool %q enum param %q has an empty value", toolName, p.Name)
			}
		}
	case "number":
		// 0,0 = unbounded (tool.Param semantics); an inverted non-zero pair is
		// the only rejectable shape, exactly as tool.Validate checks.
		if p.Min != 0 && p.Max != 0 && p.Min > p.Max {
			return ruleErr("T2", "tool %q number param %q has min %d > max %d", toolName, p.Name, p.Min, p.Max)
		}
	default:
		return ruleErr("T2", "tool %q param %q has unknown kind %q (want agent_name|text|enum|number)", toolName, p.Name, p.Kind)
	}
	return nil
}

// synthesize builds the tool.Tool the toolloop, schema derivation, and door
// validation consume — Effect Expressive (bundle tools land through
// InjectSocial), a Charge gate iff the manifest sets a charge floor, the
// description as PromptGloss, and the declared events verbatim. The param
// mapping is the ManifestParam→tool.Param projection; text params carry their
// byte cap, enum params their values, number params their bounds.
func (m *Manifest) synthesize() tool.Tool {
	params := make([]tool.Param, 0, len(m.Params))
	for _, p := range m.Params {
		tp := tool.Param{Name: p.Name, Required: p.Required, Description: p.Description}
		switch p.Kind {
		case "agent_name":
			tp.Kind = tool.AgentName
		case "text":
			tp.Kind = tool.Text
			tp.MaxBytes = p.MaxBytes
			if tp.MaxBytes == 0 {
				tp.MaxBytes = narrateMaxBytes // default text budget
			}
		case "enum":
			tp.Kind = tool.Enum
			tp.Enum = append([]string(nil), p.Enum...)
		case "number":
			tp.Kind = tool.Number
			tp.Min, tp.Max = p.Min, p.Max
		}
		params = append(params, tp)
	}

	gate := tool.None
	if m.Charges > 0 {
		gate = tool.Charge
	}
	return tool.Tool{
		Name:        m.Name,
		Effect:      tool.Expressive,
		Params:      params,
		Gate:        gate,
		Cost:        tool.Cost{Charges: m.Charges},
		PromptGloss: m.Description,
		Events:      append([]string(nil), m.Events...),
	}
}

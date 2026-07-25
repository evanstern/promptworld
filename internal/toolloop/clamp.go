package toolloop

import (
	"encoding/json"
	"fmt"
	"unicode/utf8"
)

// Clamp-with-notice (spec 058, TASK-110): the driver's Text validation arm
// truncates an over-cap EXPRESSIVE field instead of rejecting the call
// (research R1/R3) — the single biggest source of wasted model turns the
// task's world-01 diagnosis found. This file holds the rune-safe truncation
// idiom (never splits a UTF-8 sequence, for both rune and byte caps) and the
// two clamp sites validateArgs consults: a Clamp-flagged Param (the Params-
// derived path) and set_plan's top-level `reason` (the one clampable field
// that rides an authored InputSchemaJSON instead of a Param — tool.go's
// Clamp doc explains why).

// ClampRunes rune-safely truncates s to at most capRunes runes (capRunes <= 0
// means no bound: s is returned unchanged). Rune slicing only ever cuts
// between decoded runes, so this can never emit an invalid UTF-8 sequence.
func ClampRunes(s string, capRunes int) string {
	if capRunes <= 0 {
		return s
	}
	r := []rune(s)
	if len(r) <= capRunes {
		return s
	}
	return string(r[:capRunes])
}

// ClampBytes rune-safely truncates s to at most capBytes bytes (capBytes <= 0
// means no bound), trimming whole runes from the end until the byte length
// fits — the existing NormTextMax idiom (internal/mind/meeting.go), factored
// here so every byte-cap caller shares one rune-safe truncation (research R3).
func ClampBytes(s string, capBytes int) string {
	if capBytes <= 0 || len(s) <= capBytes {
		return s
	}
	for len(s) > capBytes {
		r := []rune(s)
		s = string(r[:len(r)-1])
	}
	return s
}

// clampText applies a Param's MaxRunes/MaxBytes bound (either, both, or
// neither may be set — today's registry never sets both on one field, but
// nothing requires that) and reports what happened: "" when s was already
// within every configured cap (no clamp occurred), otherwise a human note
// naming the cap that fired for the model-facing notice (FR-001).
func clampText(s string, maxRunes, maxBytes int) (string, string) {
	clamped := s
	capVal, unit := 0, ""
	if maxRunes > 0 && utf8.RuneCountInString(clamped) > maxRunes {
		clamped = ClampRunes(clamped, maxRunes)
		capVal, unit = maxRunes, "rune"
	}
	if maxBytes > 0 && len(clamped) > maxBytes {
		clamped = ClampBytes(clamped, maxBytes)
		capVal, unit = maxBytes, "byte"
	}
	if clamped == s {
		return s, ""
	}
	return clamped, fmt.Sprintf("exceeds its %d-%s cap — truncated to fit", capVal, unit)
}

// mustMarshalString marshals a Go string (a value that always marshals) into
// a json.RawMessage — the clamp rewrite's field replacement.
func mustMarshalString(s string) json.RawMessage {
	b, err := json.Marshal(s)
	if err != nil {
		panic("toolloop: string marshal: " + err.Error())
	}
	return b
}

// joinNotice appends one clamp note to an accumulating notice string (a call
// may clamp more than one Clamp-flagged param, though today's tools each
// carry at most one).
func joinNotice(existing, add string) string {
	if existing == "" {
		return add
	}
	return existing + "; " + add
}

// withClampNotice appends a non-empty clamp notice to a model-facing result
// string (FR-001: "names the field and the clamp"); "" leaves result
// untouched (set_plan's own step-count clamp already phrases its own
// ResultForModel — see handleSetPlan — so this is a no-op for it).
func withClampNotice(result, notice string) string {
	if notice == "" {
		return result
	}
	return result + " (clamped: " + notice + ")"
}

// clampTopLevelText clamps one plain top-level string property of an
// authored-schema tool's decoded arguments (spec 058: set_plan's shared
// optional `reason`, which rides the tool's InputSchemaJSON override rather
// than a Param — InputSchema bypasses Params entirely for such a tool, so
// there is no Param.Clamp flag to key on; the field is instead named
// directly). args is the already-decoded top-level object; raw is the
// original bytes. Absent, non-string, or already-in-cap fields return raw
// unchanged (nothing to clamp; an absent/wrong-typed field is left for the
// structural walker to accept or reject on its own terms).
func clampTopLevelText(args map[string]json.RawMessage, raw json.RawMessage, field string, maxRunes, maxBytes int) (json.RawMessage, string) {
	rawv, present := args[field]
	if !present {
		return raw, ""
	}
	s, ok := jsonString(rawv)
	if !ok {
		return raw, ""
	}
	clamped, notice := clampText(s, maxRunes, maxBytes)
	if notice == "" {
		return raw, ""
	}
	args[field] = mustMarshalString(clamped)
	rewritten, err := json.Marshal(args)
	if err != nil {
		// args holds only json.RawMessage values already round-tripped through
		// Unmarshal (or one freshly marshaled string); re-marshaling cannot fail.
		return raw, ""
	}
	return rewritten, fmt.Sprintf("argument %q %s", field, notice)
}

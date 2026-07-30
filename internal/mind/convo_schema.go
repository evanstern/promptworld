package mind

import "encoding/json"

// Conversation-scene structured-output schemas (spec 103/TASK-174, restoring
// TASK-58's mechanism for the conversation route: playtest-1 saw 22 outcome
// parse failures and 62/293 abandoned scenes from replies that opened with
// prose instead of JSON). convoOutcomeSchema and sayReplySchema constrain the
// local tier's sampler to the reply SHAPE at the wire (openai_compat's
// response_format {type: json_schema}); parseOutcome/parseSay in parse.go
// remain the final gate and keep clamping VALUES (tone range, byte caps) —
// this file never duplicates that logic, only the shape bounds. Built once
// from the package's existing single sources of truth
// (gistCapBytes/sayCapBytes from parse.go, sceneCap from convo.go) so the
// schema and the parser can never drift apart; no literal here is
// hand-copied.
//
// retoldMaxLen is NOT a registry cap — parseOutcome does not clamp retold at
// all — it is a token-budget bound only, chosen the same way TASK-58 bounded
// the planner's free-text fields: without a ceiling a rambling retold value
// can overrun the outcome call's MaxTokens mid-string and arrive as
// truncated, unterminated JSON (TASK-58's truncated-JSON lesson, restated in
// spec 103 D3).
const retoldMaxLen = 300

// buildConvoOutcomeSchema builds the structured-output schema for
// Mind.outcome's Submit (exposed once as convoOutcomeSchema below). Flat and
// required-complete (D3): all four keys are required, and
// there is deliberately no `anyOf` and no integer `minimum`/`maximum` on
// tones — TASK-58's live findings are binding here too: llama.cpp's
// schema-to-grammar converter bails out on `anyOf` and enforces nothing at
// all, so the shape stays a single flat object; tone clamping to -2..2 stays
// entirely in parseOutcome; the schema constrains SHAPE, the parser clamps
// VALUES. maxLength counts UTF-16 code units/code points per the JSON Schema
// spec while the registry caps are bytes — parseOutcome's rune-safe byte
// clamp (toolloop.ClampBytes) remains the authoritative cap regardless of
// what the schema allowed through. The legacy tone_a/tone_b pair shape is not
// re-added here (unconstrained cloud replies that use it still parse via
// parseOutcome; this schema only ever rides an openai_compat request).
func buildConvoOutcomeSchema() json.RawMessage {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"gist": map[string]any{
				"type":      "string",
				"maxLength": gistCapBytes,
			},
			"topics": map[string]any{
				"type":     "array",
				"maxItems": 3,
				"items": map[string]any{
					"type":      "string",
					"maxLength": 40,
				},
			},
			"tones": map[string]any{
				"type":     "array",
				"maxItems": sceneCap,
				"items":    map[string]any{"type": "integer"},
			},
			"retold": map[string]any{
				"type":      "string",
				"maxLength": retoldMaxLen,
			},
		},
		"required": []string{"gist", "topics", "tones", "retold"},
	}
	b, err := json.Marshal(schema)
	if err != nil {
		// The schema is built from literal maps of literal values; marshaling
		// cannot fail. Panic is the honest response to an impossible state at
		// package init.
		panic("mind: conversation outcome schema marshal: " + err.Error())
	}
	return b
}

// buildSayReplySchema builds the structured-output schema for
// Mind.utterance's Submit (exposed once as sayReplySchema below; D4): the
// same {"say": "..."} shape parseSay already expects, with the say cap as a
// maxLength bound so a free-generating local model can't ramble past the
// utterance call's MaxTokens budget mid-string.
func buildSayReplySchema() json.RawMessage {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"say": map[string]any{
				"type":      "string",
				"maxLength": sayCapBytes,
			},
		},
		"required": []string{"say"},
	}
	b, err := json.Marshal(schema)
	if err != nil {
		panic("mind: say reply schema marshal: " + err.Error())
	}
	return b
}

// convoOutcomeSchema and sayReplySchema are built once — the registry caps
// and sceneCap are compile-time constants of the package, so neither schema
// changes at runtime. Mind.outcome/Mind.utterance stamp these on their
// Submits.
var (
	convoOutcomeSchema = buildConvoOutcomeSchema()
	sayReplySchema     = buildSayReplySchema()
)

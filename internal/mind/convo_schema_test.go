package mind

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestConvoOutcomeSchema (spec 103/TASK-174): the generated conversation
// outcome schema is valid JSON, flat (no anyOf anywhere — TASK-58's
// llama.cpp finding is binding here too), its caps equal the package's
// existing single sources of truth (gistCapBytes, the 3/40 topic bounds,
// sceneCap), all four keys are required, and no property carries an integer
// minimum/maximum (tone clamping stays in parseOutcome).
func TestConvoOutcomeSchema(t *testing.T) {
	raw := convoOutcomeSchema
	if !json.Valid(raw) {
		t.Fatalf("schema is not valid JSON: %s", raw)
	}
	if strings.Contains(string(raw), `"anyOf"`) {
		t.Errorf("schema contains anyOf, which llama.cpp's grammar converter bails out on: %s", raw)
	}
	if strings.Contains(string(raw), `"minimum"`) || strings.Contains(string(raw), `"maximum"`) {
		t.Errorf("schema constrains integer min/max — tone clamping must stay in parseOutcome only: %s", raw)
	}

	var schema struct {
		Type       string `json:"type"`
		Properties struct {
			Gist struct {
				Type      string `json:"type"`
				MaxLength int    `json:"maxLength"`
			} `json:"gist"`
			Topics struct {
				Type     string `json:"type"`
				MaxItems int    `json:"maxItems"`
				Items    struct {
					Type      string `json:"type"`
					MaxLength int    `json:"maxLength"`
				} `json:"items"`
			} `json:"topics"`
			Tones struct {
				Type     string `json:"type"`
				MaxItems int    `json:"maxItems"`
				Items    struct {
					Type string `json:"type"`
				} `json:"items"`
			} `json:"tones"`
			Retold struct {
				Type      string `json:"type"`
				MaxLength int    `json:"maxLength"`
			} `json:"retold"`
		} `json:"properties"`
		Required []string `json:"required"`
	}
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}

	if schema.Type != "object" {
		t.Errorf("schema.type = %q, want object", schema.Type)
	}
	if schema.Properties.Gist.MaxLength != gistCapBytes {
		t.Errorf("gist maxLength = %d, want registry cap %d", schema.Properties.Gist.MaxLength, gistCapBytes)
	}
	if schema.Properties.Topics.MaxItems != 3 {
		t.Errorf("topics maxItems = %d, want 3", schema.Properties.Topics.MaxItems)
	}
	if schema.Properties.Topics.Items.MaxLength != 40 {
		t.Errorf("topics item maxLength = %d, want 40", schema.Properties.Topics.Items.MaxLength)
	}
	if schema.Properties.Tones.MaxItems != sceneCap {
		t.Errorf("tones maxItems = %d, want sceneCap %d", schema.Properties.Tones.MaxItems, sceneCap)
	}
	if schema.Properties.Tones.Items.Type != "integer" {
		t.Errorf("tones item type = %q, want integer", schema.Properties.Tones.Items.Type)
	}
	if schema.Properties.Retold.MaxLength != retoldMaxLen {
		t.Errorf("retold maxLength = %d, want %d", schema.Properties.Retold.MaxLength, retoldMaxLen)
	}

	wantRequired := []string{"gist", "topics", "tones", "retold"}
	if len(schema.Required) != len(wantRequired) {
		t.Fatalf("required = %v, want %v", schema.Required, wantRequired)
	}
	for i, k := range wantRequired {
		if schema.Required[i] != k {
			t.Errorf("required[%d] = %q, want %q", i, schema.Required[i], k)
		}
	}
}

// TestSayReplySchema (spec 103/TASK-174, D4): the utterance schema mirrors
// convoOutcomeSchema's discipline — valid JSON, no anyOf, its say cap equals
// the registry's sayCapBytes, and "say" is required.
func TestSayReplySchema(t *testing.T) {
	raw := sayReplySchema
	if !json.Valid(raw) {
		t.Fatalf("schema is not valid JSON: %s", raw)
	}
	if strings.Contains(string(raw), `"anyOf"`) {
		t.Errorf("schema contains anyOf: %s", raw)
	}

	var schema struct {
		Type       string `json:"type"`
		Properties struct {
			Say struct {
				Type      string `json:"type"`
				MaxLength int    `json:"maxLength"`
			} `json:"say"`
		} `json:"properties"`
		Required []string `json:"required"`
	}
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}
	if schema.Type != "object" {
		t.Errorf("schema.type = %q, want object", schema.Type)
	}
	if schema.Properties.Say.MaxLength != sayCapBytes {
		t.Errorf("say maxLength = %d, want registry cap %d", schema.Properties.Say.MaxLength, sayCapBytes)
	}
	if len(schema.Required) != 1 || schema.Required[0] != "say" {
		t.Errorf("required = %v, want [say]", schema.Required)
	}
}

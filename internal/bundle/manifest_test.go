package bundle

import (
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/tool"
)

// TestParseValidManifest: a well-formed declarative manifest parses and
// synthesizes a tool.Tool with the expected class, gate, cost, gloss, events,
// and param projection.
func TestParseValidManifest(t *testing.T) {
	src := `{
	  "name": "teleport",
	  "description": "Whisk a villager across the map",
	  "params": [
	    {"name": "target", "kind": "agent_name", "required": true},
	    {"name": "x", "kind": "number", "min": 0, "max": 63},
	    {"name": "note", "kind": "text", "max_bytes": 120},
	    {"name": "flavor", "kind": "enum", "enum": ["soft", "blinding"]}
	  ],
	  "events": ["metatron.entity_moved", "agent.memory_added"],
	  "charges": 2,
	  "effects": [
	    {"kind": "move_entity", "target": "{args.target}", "to_x": "{args.x}", "to_y": 0}
	  ]
	}`
	m, err := Parse([]byte(src), "teleport")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	got := m.synthesize()
	if got.Name != "teleport" || got.Effect != tool.Expressive {
		t.Errorf("name/effect = %q/%v", got.Name, got.Effect)
	}
	if got.Gate != tool.Charge {
		t.Errorf("gate = %v, want Charge (charges>0)", got.Gate)
	}
	if got.Cost.Charges != 2 {
		t.Errorf("cost.charges = %d, want 2", got.Cost.Charges)
	}
	if got.PromptGloss != "Whisk a villager across the map" {
		t.Errorf("gloss = %q", got.PromptGloss)
	}
	if len(got.Events) != 2 {
		t.Errorf("events = %v", got.Events)
	}
	if len(got.Params) != 4 {
		t.Fatalf("params = %d, want 4", len(got.Params))
	}
	if got.Params[0].Kind != tool.AgentName || !got.Params[0].Required {
		t.Errorf("param[0] = %+v", got.Params[0])
	}
	if got.Params[1].Kind != tool.Number || got.Params[1].Min != 0 || got.Params[1].Max != 63 {
		t.Errorf("param[1] = %+v", got.Params[1])
	}
	if got.Params[2].Kind != tool.Text || got.Params[2].MaxBytes != 120 {
		t.Errorf("param[2] = %+v", got.Params[2])
	}
	if got.Params[3].Kind != tool.Enum || len(got.Params[3].Enum) != 2 {
		t.Errorf("param[3] = %+v", got.Params[3])
	}
}

// TestSynthesizeGateNoneWhenCostless: a zero-charge tool gets Gate None.
func TestSynthesizeGateNoneWhenCostless(t *testing.T) {
	m, err := Parse([]byte(`{"name":"whisper","description":"a quiet word","events":["agent.memory_added"],"effects":[{"kind":"narrate","text":"hi","recipients":"all_living"}]}`), "whisper")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got := m.synthesize(); got.Gate != tool.None || got.Cost.Charges != 0 {
		t.Errorf("gate/charges = %v/%d, want None/0", got.Gate, got.Cost.Charges)
	}
}

// TestSynthesizeTextParamDefaultCap: a text param without max_bytes gets the
// default byte budget.
func TestSynthesizeTextParamDefaultCap(t *testing.T) {
	m, err := Parse([]byte(`{"name":"say","description":"speak","params":[{"name":"msg","kind":"text"}],"events":["agent.memory_added"],"effects":[{"kind":"narrate","text":"{args.msg}","recipients":"all_living"}]}`), "say")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got := m.synthesize(); got.Params[0].MaxBytes != narrateMaxBytes {
		t.Errorf("text param MaxBytes = %d, want %d", got.Params[0].MaxBytes, narrateMaxBytes)
	}
}

// TestParseRejects: each malformed manifest fails with a message naming the
// specific problem, and the rule id maps to the boot-validation ladder (T1/T2/
// T4/T7).
func TestParseRejects(t *testing.T) {
	cases := []struct {
		name   string
		folder string
		src    string
		want   string // substring of the message
		rule   string
	}{
		{
			name:   "unknown key",
			folder: "t",
			src:    `{"name":"t","description":"d","events":["agent.memory_added"],"effects":[{"kind":"narrate","text":"x","recipients":"all_living"}],"bogus":1}`,
			want:   "does not decode",
			rule:   "T1",
		},
		{
			name:   "name mismatch folder",
			folder: "wrong",
			src:    `{"name":"t","description":"d","events":["agent.memory_added"],"effects":[{"kind":"narrate","text":"x","recipients":"all_living"}]}`,
			want:   "does not match its folder",
			rule:   "T1",
		},
		{
			name:   "bad name shape",
			folder: "Bad-Name",
			src:    `{"name":"Bad-Name","description":"d","events":["agent.memory_added"],"effects":[{"kind":"narrate","text":"x","recipients":"all_living"}]}`,
			want:   "must match [a-z0-9_]{1,48}",
			rule:   "T1",
		},
		{
			name:   "empty description",
			folder: "t",
			src:    `{"name":"t","description":"","events":["agent.memory_added"],"effects":[{"kind":"narrate","text":"x","recipients":"all_living"}]}`,
			want:   "description must be 1",
			rule:   "T1",
		},
		{
			name:   "too many params",
			folder: "t",
			src:    `{"name":"t","description":"d","params":[{"name":"a","kind":"text"},{"name":"b","kind":"text"},{"name":"c","kind":"text"},{"name":"d","kind":"text"},{"name":"e","kind":"text"},{"name":"f","kind":"text"},{"name":"g","kind":"text"},{"name":"h","kind":"text"},{"name":"i","kind":"text"}],"events":["agent.memory_added"],"effects":[{"kind":"narrate","text":"x","recipients":"all_living"}]}`,
			want:   "params (max 8)",
			rule:   "T2",
		},
		{
			name:   "enum without values",
			folder: "t",
			src:    `{"name":"t","description":"d","params":[{"name":"e","kind":"enum"}],"events":["agent.memory_added"],"effects":[{"kind":"narrate","text":"x","recipients":"all_living"}]}`,
			want:   "must list 1",
			rule:   "T2",
		},
		{
			name:   "number inverted bounds",
			folder: "t",
			src:    `{"name":"t","description":"d","params":[{"name":"n","kind":"number","min":9,"max":2}],"events":["agent.memory_added"],"effects":[{"kind":"narrate","text":"x","recipients":"all_living"}]}`,
			want:   "min 9 > max 2",
			rule:   "T2",
		},
		{
			name:   "unknown param kind",
			folder: "t",
			src:    `{"name":"t","description":"d","params":[{"name":"n","kind":"date"}],"events":["agent.memory_added"],"effects":[{"kind":"narrate","text":"x","recipients":"all_living"}]}`,
			want:   "unknown kind",
			rule:   "T2",
		},
		{
			name:   "both effects and script",
			folder: "t",
			src:    `{"name":"t","description":"d","events":["agent.memory_added"],"script":"t.star","effects":[{"kind":"narrate","text":"x","recipients":"all_living"}]}`,
			want:   "declares both effects and script",
			rule:   "T4",
		},
		{
			name:   "neither effects nor script",
			folder: "t",
			src:    `{"name":"t","description":"d","events":["agent.memory_added"]}`,
			want:   "declares neither",
			rule:   "T4",
		},
		{
			name:   "negative charges",
			folder: "t",
			src:    `{"name":"t","description":"d","charges":-1,"events":["agent.memory_added"],"effects":[{"kind":"narrate","text":"x","recipients":"all_living"}]}`,
			want:   "charges must be ≥0",
			rule:   "T7",
		},
		{
			name:   "max_steps over ceiling",
			folder: "t",
			src:    `{"name":"t","description":"d","events":["agent.memory_added"],"script":"t.star","limits":{"max_steps":2000000}}`,
			want:   "max_steps must be in (0, 1000000]",
			rule:   "T7",
		},
		{
			name:   "max_steps zero",
			folder: "t",
			src:    `{"name":"t","description":"d","events":["agent.memory_added"],"script":"t.star","limits":{"max_steps":0}}`,
			want:   "max_steps must be in (0, 1000000]",
			rule:   "T7",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse([]byte(tc.src), tc.folder)
			if err == nil {
				t.Fatalf("expected error")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("message = %q, want substring %q", err.Error(), tc.want)
			}
			if got := ruleOf(err); got != tc.rule {
				t.Errorf("rule = %q, want %q", got, tc.rule)
			}
		})
	}
}

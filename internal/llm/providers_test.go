package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// captureServer is a chat-completions server that decodes and stashes the
// raw request body for the test to assert on, then replies with a minimal
// valid completion.
func captureServer(t *testing.T, body *map[string]any) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(body)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{"message": map[string]any{"content": "ok"}}},
			"usage":   map[string]any{"prompt_tokens": 1, "completion_tokens": 1},
		})
	}))
	t.Cleanup(srv.Close)
	return srv
}

// TestOpenAICompatMaxTokens: Request.MaxTokens rides the wire as
// "max_tokens" when positive, and is omitted entirely when zero — a
// silently dropped MaxTokens is the TASK-37 defect this guards against.
func TestOpenAICompatMaxTokens(t *testing.T) {
	var got map[string]any
	srv := captureServer(t, &got)
	o := newOpenAICompat(srv.URL, "m", "", "", "")

	if _, err := o.call(context.Background(), Request{Prompt: "x", MaxTokens: 256}); err != nil {
		t.Fatal(err)
	}
	if v, ok := got["max_tokens"]; !ok || v.(float64) != 256 {
		t.Errorf("max_tokens = %v, want 256", got["max_tokens"])
	}

	got = nil
	if _, err := o.call(context.Background(), Request{Prompt: "x"}); err != nil {
		t.Fatal(err)
	}
	if _, ok := got["max_tokens"]; ok {
		t.Errorf("max_tokens present with MaxTokens=0: %v", got["max_tokens"])
	}
}

// TestOpenAICompatReasoningEffort covers the resolved reasoningEffort
// field across both tiers' nil/""/override conventions (resolution itself
// is resolveReasoningEffort's job; this exercises what the caller sends
// once resolved).
func TestOpenAICompatReasoningEffort(t *testing.T) {
	cases := []struct {
		name       string
		resolved   string // as produced by resolveReasoningEffort
		wantAbsent bool
		want       string
	}{
		{"local default none", resolveReasoningEffort(nil, "none"), false, "none"},
		{"local explicit empty", resolveReasoningEffort(strPtr(""), "none"), true, ""},
		{"local override", resolveReasoningEffort(strPtr("low"), "none"), false, "low"},
		{"cloud default omitted", resolveReasoningEffort(nil, ""), true, ""},
		{"cloud explicit value", resolveReasoningEffort(strPtr("medium"), ""), false, "medium"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var got map[string]any
			srv := captureServer(t, &got)
			o := newOpenAICompat(srv.URL, "m", "", c.resolved, "")
			if _, err := o.call(context.Background(), Request{Prompt: "x"}); err != nil {
				t.Fatal(err)
			}
			v, ok := got["reasoning_effort"]
			if c.wantAbsent {
				if ok {
					t.Errorf("reasoning_effort present: %v, want absent", v)
				}
				return
			}
			if !ok || v.(string) != c.want {
				t.Errorf("reasoning_effort = %v, want %q", v, c.want)
			}
		})
	}
}

// TestOpenAICompatResponseFormat (spec 103/TASK-174, restoring TASK-58): a
// Request.ResponseSchema rides the wire as a well-formed response_format
// {type: json_schema} envelope when set; no response_format key appears at
// all when unset (the byte-identical baseline every non-conversation kind
// relies on); and it is NOT attached when the request also carries Tools —
// the tool-call envelope already owns response_format for its own purposes.
func TestOpenAICompatResponseFormat(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","properties":{"gist":{"type":"string"}},"required":["gist"]}`)

	t.Run("schema set", func(t *testing.T) {
		var got map[string]any
		srv := captureServer(t, &got)
		o := newOpenAICompat(srv.URL, "m", "", "", "")
		if _, err := o.call(context.Background(),
			Request{Prompt: "x", ResponseSchema: schema, SchemaName: "conversation_outcome"}); err != nil {
			t.Fatal(err)
		}
		rf, ok := got["response_format"].(map[string]any)
		if !ok {
			t.Fatalf("response_format missing or wrong type: %v", got["response_format"])
		}
		if rf["type"] != "json_schema" {
			t.Errorf("response_format.type = %v, want json_schema", rf["type"])
		}
		js, ok := rf["json_schema"].(map[string]any)
		if !ok {
			t.Fatalf("json_schema missing or wrong type: %v", rf["json_schema"])
		}
		if js["name"] != "conversation_outcome" {
			t.Errorf("json_schema.name = %v, want conversation_outcome", js["name"])
		}
		sch, ok := js["schema"].(map[string]any)
		if !ok {
			t.Fatalf("json_schema.schema missing or not an object: %v", js["schema"])
		}
		if sch["type"] != "object" {
			t.Errorf("schema round-trip mangled: schema.type = %v, want object", sch["type"])
		}
	})

	t.Run("schema unset", func(t *testing.T) {
		var got map[string]any
		srv := captureServer(t, &got)
		o := newOpenAICompat(srv.URL, "m", "", "", "")
		if _, err := o.call(context.Background(), Request{Prompt: "x"}); err != nil {
			t.Fatal(err)
		}
		if _, ok := got["response_format"]; ok {
			t.Errorf("response_format present with no ResponseSchema: %v", got["response_format"])
		}
	})

	t.Run("schema with Tools not attached", func(t *testing.T) {
		var got map[string]any
		srv := captureServer(t, &got)
		o := newOpenAICompat(srv.URL, "m", "", "", "")
		if _, err := o.call(context.Background(), Request{
			Prompt:         "x",
			ResponseSchema: schema,
			SchemaName:     "conversation_outcome",
			Tools:          []ToolDecl{{Name: "noop", Description: "d", InputSchema: json.RawMessage(`{"type":"object"}`)}},
		}); err != nil {
			t.Fatal(err)
		}
		if _, ok := got["response_format"]; ok {
			t.Errorf("response_format present alongside Tools: %v", got["response_format"])
		}
	})
}

func strPtr(s string) *string { return &s }

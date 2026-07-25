package mind

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// TestParseOutcomeLenient (TASK-42 T014): the observed unquoted-value shapes
// (gist starting with a participant initial F/H/S; a bare retold value)
// recover without a model call, while genuinely broken replies still fail.
func TestParseOutcomeLenient(t *testing.T) {
	recoverable := []struct {
		name string
		in   string
		gist string
	}{
		{"unquoted gist F", `{"gist": Fenwick and Rowan talked about the fire, "topics": ["fire"], "tones": [1, 1], "retold": null}`, "Fenwick and Rowan talked about the fire"},
		{"unquoted gist H", `{"gist": Hazel and Rowan talked about the fire, "topics": ["fire"], "tones": [1, 1], "retold": null}`, "Hazel and Rowan talked about the fire"},
		{"unquoted gist S", `{"gist": Sorrel and Rowan talked about the fire, "topics": ["fire"], "tones": [1, 1], "retold": null}`, "Sorrel and Rowan talked about the fire"},
		{"unquoted retold", `{"gist": "planned the firewood run", "topics": ["fire"], "tones": [1, 1], "retold": she said the fire is cursed}`, "planned the firewood run"},
	}
	for _, tc := range recoverable {
		o, err := parseOutcome(tc.in)
		if err != nil {
			t.Errorf("%s: expected lenient recovery, got %v", tc.name, err)
			continue
		}
		if o.Gist != tc.gist {
			t.Errorf("%s: gist = %q, want %q", tc.name, o.Gist, tc.gist)
		}
		if len(o.Tones) != 2 {
			t.Errorf("%s: tones = %v, want 2", tc.name, o.Tones)
		}
	}

	// The last case's retold must be recovered as its prose, not dropped.
	if o, _ := parseOutcome(recoverable[3].in); o.Retold != "she said the fire is cursed" {
		t.Errorf("unquoted retold = %q, want the recovered prose", o.Retold)
	}

	unrecoverable := []struct {
		name string
		in   string
	}{
		{"prose only", `the model just rambled in prose about the weather`},
		{"unterminated object", `{"gist": "planned the firewood`},
		{"valid json empty gist", `{"gist": "", "topics": ["fire"], "tones": [1, 1]}`},
		{"no gist key", `{"topics": ["fire"], "tones": [1, 1], "retold": null}`},
	}
	for _, tc := range unrecoverable {
		if _, err := parseOutcome(tc.in); err == nil {
			t.Errorf("%s: expected failure, got nil", tc.name)
		}
	}
}

// TestParseOutcomeHappyPathUnchanged: a well-formed reply parses untouched —
// the lenient path is never entered when json.Unmarshal already succeeds.
func TestParseOutcomeHappyPathUnchanged(t *testing.T) {
	o, err := parseOutcome(`{"gist": "planned firewood", "topics": ["fire", "chores"], "tones": [2, -1], "retold": null}`)
	if err != nil {
		t.Fatal(err)
	}
	if o.Gist != "planned firewood" || len(o.Topics) != 2 || len(o.Tones) != 2 || o.Tones[1] != -1 {
		t.Errorf("parsed: %+v", o)
	}
}

// TestParseSayClampsRuneSafely (spec 058 edge case): say already never
// rejected for length — this pins the fix that made its truncation rune-safe.
// A multi-byte utterance padded past the byte cap must truncate to a WHOLE
// rune below the cap, never emit invalid UTF-8, and never reject.
func TestParseSayClampsRuneSafely(t *testing.T) {
	// Multi-byte filler (é, 2 bytes/rune) so the byte cap lands mid-character
	// unless the clamp is rune-safe; sayCapBytes is 300.
	over := strings.Repeat("é", 200) // 400 bytes, well over the 300-byte cap
	say, err := parseSay(`{"say":"` + over + `"}`)
	if err != nil {
		t.Fatalf("over-cap say rejected: %v (say has always clamped, never rejected, for length)", err)
	}
	if len(say) > sayCapBytes {
		t.Errorf("clamped say = %d bytes, want <= %d", len(say), sayCapBytes)
	}
	if !utf8.ValidString(say) {
		t.Error("clamped say is not valid UTF-8 (a naive byte slice split a multi-byte rune)")
	}
}

// TestParseOutcomeGistClampsRuneSafely: the same rune-safety fix, for gist.
func TestParseOutcomeGistClampsRuneSafely(t *testing.T) {
	over := strings.Repeat("日", 150) // 450 bytes (3 bytes/rune), over the 200-byte cap
	o, err := parseOutcome(`{"gist": "` + over + `", "topics": [], "tones": [0, 0], "retold": null}`)
	if err != nil {
		t.Fatalf("over-cap gist rejected: %v (gist has always clamped, never rejected, for length)", err)
	}
	if len(o.Gist) > gistCapBytes {
		t.Errorf("clamped gist = %d bytes, want <= %d", len(o.Gist), gistCapBytes)
	}
	if !utf8.ValidString(o.Gist) {
		t.Error("clamped gist is not valid UTF-8 (a naive byte slice split a multi-byte rune)")
	}
}

package main

import (
	"strings"
	"testing"
)

// TestGuardianWorkAliases (spec 052 FR-008, T011): the canonical
// fiction-bearing subcommands are `guardian` and `work`; the pre-052 names
// `metatron` and `miracle` stay as hidden, fully functional compatibility
// aliases — same handler, so an old script can never drift from the new
// behavior.
func TestGuardianWorkAliases(t *testing.T) {
	for _, pair := range [][2]string{{"guardian", "metatron"}, {"work", "miracle"}} {
		canonical, alias := pair[0], pair[1]
		cFn, ok := dispatch(canonical)
		if !ok {
			t.Fatalf("canonical %q not dispatchable", canonical)
		}
		aFn, ok := dispatch(alias)
		if !ok {
			t.Fatalf("compat alias %q not dispatchable", alias)
		}
		// Same handler: prove it by behavior — both no-arg invocations
		// produce the identical (canonical-vocabulary) usage error.
		cErr, aErr := cFn(nil), aFn(nil)
		if cErr == nil || aErr == nil || cErr.Error() != aErr.Error() {
			t.Errorf("%q/%q usage errors differ: %v vs %v", canonical, alias, cErr, aErr)
		}
		if strings.Contains(cErr.Error(), alias) {
			t.Errorf("%q usage error leaks the compat alias: %v", canonical, cErr)
		}
	}

	// The guardian usage error speaks the canonical vocabulary.
	fn, _ := dispatch("metatron")
	if err := fn(nil); err == nil || !strings.Contains(err.Error(), "promptworld guardian") {
		t.Errorf("alias handler's usage should name the canonical form, got %v", err)
	}
	fn, _ = dispatch("miracle")
	if err := fn(nil); err == nil || !strings.Contains(err.Error(), "promptworld work ") {
		t.Errorf("work alias usage should name the canonical form, got %v", err)
	}
}

// TestUsageShowsCanonicalOnly (spec 052 FR-008): the help text advertises
// only the canonical subcommands — the aliases are hidden.
func TestUsageShowsCanonicalOnly(t *testing.T) {
	if !strings.Contains(usage, "promptworld guardian <world>") || !strings.Contains(usage, "promptworld work <world>") {
		t.Error("usage must advertise the canonical guardian/work subcommands")
	}
	for _, banned := range []string{"metatron", "miracle", "angel", "Metatron"} {
		if strings.Contains(usage, banned) {
			t.Errorf("usage text leaks hidden/fiction vocabulary %q", banned)
		}
	}
}

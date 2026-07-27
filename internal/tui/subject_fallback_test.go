package tui

// resolveSubject generic-fallback tests (spec 086 T023, FR-008, US4
// AS-3..5, SC-004): the hit-rate proof — registry+fallback locates strictly
// more of the feed than registry-only — plus the pinned newly-locatable
// types and the multi-ref ambiguity guard.

import (
	"encoding/json"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
)

// fallbackFixtureRows are agent-bearing rows for real event types with NO
// subjectRegistry entry — the fallback's new reach. journal.* are real
// event types outside the digest catalog (spec 086 census §1);
// faith.changed{villager_died} carries the spec 086 additive ref.
var fallbackFixtureRows = map[string]string{
	"journal.entry_written": `{"agent":{"id":1,"name":"Birch"},"text":"day one"}`,
	"journal.entry_deleted": `{"agent":{"id":2,"name":"Cedar"},"entry":1}`,
	"faith.changed":         `{"delta":-6,"reason":"villager_died","source_id":"3","agent":{"id":3,"name":"Rowan"}}`,
}

// TestResolveSubjectHitRate (SC-004): drive every catalogFixture row plus
// the fallback rows through resolveSubject against a live replica;
// locatable(registry+fallback) must STRICTLY exceed locatable(registry-only),
// and the pinned types must be newly locatable via the generic pass.
func TestResolveSubjectHitRate(t *testing.T) {
	m := testModel(t)

	event := func(typ, payload string) store.Event {
		return store.Event{Seq: 1, Tick: 1, Type: typ, Payload: json.RawMessage(payload)}
	}

	registryOnly, withFallback := 0, 0
	newly := map[string]bool{}
	try := func(typ, payload string) {
		e := event(typ, payload)
		_, inRegistry := subjectRegistry[typ]
		_, _, _, ok := m.resolveSubject(e)
		if inRegistry {
			if ok {
				registryOnly++
				withFallback++
			}
			return
		}
		if ok {
			withFallback++
			newly[typ] = true
		}
	}
	for typ, fx := range catalogFixture {
		try(typ, fx.payload)
	}
	for typ, payload := range fallbackFixtureRows {
		try(typ, payload)
	}

	if withFallback <= registryOnly {
		t.Fatalf("locatable(registry+fallback)=%d is not strictly greater than registry-only=%d — the fallback adds no reach", withFallback, registryOnly)
	}
	t.Logf("locatable: registry-only=%d, registry+fallback=%d, newly locatable: %v", registryOnly, withFallback, newly)

	// Pinned newly-locatable types (data-model §7, adjusted to the shipped
	// registry: morgue.epilogue and cog.thought gained registry entries in
	// spec 049, so the fallback's pinned reach is the journal family and
	// the faith.changed additive ref).
	for _, typ := range []string{"journal.entry_written", "journal.entry_deleted", "faith.changed"} {
		if !newly[typ] {
			t.Errorf("%s: expected newly locatable via the generic single-ref fallback", typ)
		}
	}
}

// TestResolveSubjectFallbackAmbiguityStaysUnlocatable (US4 AS-4): a
// registry-miss payload with several distinct in-roster refs stays
// unlocatable — ambiguity is detected structurally, preserving the
// honest-hint doctrine. chronicle.entry (registry-miss by design) with a
// multi-agent list is the canonical case; with exactly one agent, the same
// type becomes locatable.
func TestResolveSubjectFallbackAmbiguityStaysUnlocatable(t *testing.T) {
	m := testModel(t)

	multi := store.Event{Seq: 1, Tick: 1, Type: "chronicle.entry", Payload: json.RawMessage(
		`{"day":1,"from_tick":1,"to_tick":2,"text":"x","agents":[{"id":0,"name":"Ash"},{"id":1,"name":"Birch"}]}`)}
	if _, _, _, ok := m.resolveSubject(multi); ok {
		t.Error("multi-ref chronicle.entry resolved — ambiguity must stay unlocatable")
	}

	single := store.Event{Seq: 1, Tick: 1, Type: "chronicle.entry", Payload: json.RawMessage(
		`{"day":1,"from_tick":1,"to_tick":2,"text":"x","agents":[{"id":1,"name":"Birch"}]}`)}
	name, _, _, ok := m.resolveSubject(single)
	if !ok || name != "Birch" {
		t.Errorf("single-ref chronicle.entry = (%q,%v), want Birch locatable", name, ok)
	}

	// world.migrated stays hard-excluded even though its embedded state
	// could contain ref-shaped noise.
	mig := store.Event{Seq: 1, Tick: 1, Type: "world.migrated", Payload: json.RawMessage(
		`{"from_format":2,"source_events":1,"source_tick":1,"state":{}}`)}
	if _, _, _, ok := m.resolveSubject(mig); ok {
		t.Error("world.migrated resolved — it must stay hard-excluded")
	}

	// A dead subject on a registry-miss type stays unlocatable (no
	// generically trustworthy recorded position).
	m.replica.Agents[2].Dead = true
	dead := store.Event{Seq: 1, Tick: 1, Type: "journal.entry_deleted", Payload: json.RawMessage(
		`{"agent":{"id":2,"name":"Cedar"},"entry":1}`)}
	if _, _, _, ok := m.resolveSubject(dead); ok {
		t.Error("dead subject resolved via the fallback — live position is the only generic anchor")
	}

	// Legacy (unnamed) registry-miss rows stay unlocatable too: a bare int
	// is not a ref object — the fallback only trusts the named shape.
	legacy := store.Event{Seq: 1, Tick: 1, Type: "journal.entry_written", Payload: json.RawMessage(
		`{"agent":1,"text":"old row"}`)}
	if _, _, _, ok := m.resolveSubject(legacy); ok {
		t.Error("legacy bare-int row resolved via the fallback — only the named object shape is generic evidence")
	}
	_ = sim.AgentCount
}

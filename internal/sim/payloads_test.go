package sim

// Enforcement sweeps for spec 086 (FR-004/FR-006, SC-002): the doc-anchored
// catalog completeness check, TestPayloadAgentRefSweep (frozen agent
// vocabulary + frozen rationale-carrying allowlist), and
// TestNoAgentRefInState (the R2 hash-stability invariant as a standing
// tripwire).

import (
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

// eventNamespaces mirrors the tui's familyByNamespace key set: the event
// namespaces backtickedEventTypes recognizes, so incidental backticked
// tokens in prose (`social.go`, `world.json`) are not mistaken for event
// types.
var eventNamespaces = map[string]bool{
	"world": true, "clock": true, "daemon": true, "run": true, "sim": true,
	"agent": true, "social": true, "journal": true, "meeting": true,
	"norm": true, "gru": true, "stranger": true, "chronicle": true,
	"metatron": true, "guardian": true, "designation": true,
	"directive": true, "faith": true, "prophecy": true, "morgue": true,
	"cog": true, "curriculum": true,
}

var backtickedTypeRe = regexp.MustCompile("`([a-z]+)\\.([a-z_]+)`")

func backtickedEventTypes(doc string) []string {
	var types []string
	seen := map[string]bool{}
	for _, m := range backtickedTypeRe.FindAllStringSubmatch(doc, -1) {
		ns, verb := m[1], m[2]
		if !eventNamespaces[ns] {
			continue
		}
		if verb == "go" || verb == "json" || verb == "md" {
			continue // source-file / config references, not event types
		}
		t := ns + "." + verb
		if !seen[t] {
			seen[t] = true
			types = append(types, t)
		}
	}
	return types
}

// TestPayloadCatalogCompleteness is the doc anchor (FR-006): every
// backticked event type in docs/wiki/event-types.md must be in
// PayloadCatalog — a new event type cannot exist outside the catalog
// without failing here (the TestCatalogSweep trick, sim-side). It also
// welds the injection door to the catalog: every injectSocialWhitelist
// type must be cataloged, so the door's decode-and-refuse rail can never
// hit an unregistered payload at runtime (FR-005).
func TestPayloadCatalogCompleteness(t *testing.T) {
	doc, err := os.ReadFile("../../docs/wiki/event-types.md")
	if err != nil {
		t.Fatalf("reading docs/wiki/event-types.md: %v", err)
	}
	for _, typ := range backtickedEventTypes(string(doc)) {
		if _, ok := PayloadCatalog[typ]; !ok {
			t.Errorf("docs/wiki/event-types.md backticks %q but PayloadCatalog doesn't cover it", typ)
		}
	}
	for typ := range injectSocialWhitelist {
		if _, ok := PayloadCatalog[typ]; !ok {
			t.Errorf("injectSocialWhitelist type %q is not in PayloadCatalog — the door cannot validate it", typ)
		}
	}
	for typ := range endedProseWhitelist {
		if _, ok := PayloadCatalog[typ]; !ok {
			t.Errorf("endedProseWhitelist type %q is not in PayloadCatalog", typ)
		}
	}
}

// agentVocabulary is the FROZEN tag vocabulary (data-model §5): an int-kind
// field under one of these json tags is an agent reference and must be
// AgentRef-typed. Grow it deliberately (with a spec), never casually.
var agentVocabulary = map[string]bool{
	"agent": true, "a": true, "b": true, "from": true, "to": true,
	"speaker": true, "listener": true, "subject": true, "owner": true,
	"taker": true, "violator": true, "witnesses": true, "targets": true,
	"attendees": true, "proposer": true, "target": true, "yeas": true,
	"nays": true, "agents": true, "participants": true, "source": true,
	"src": true,
}

// agentRefAllowlist is the FROZEN exemption set (data-model §5, complete):
// (declaring type).(field) → rationale. Every entry must match a live field
// (no dead exemptions — removing a use without removing its entry fails the
// sweep). Entries 2–4 are documentation-grade (non-int fields the sweep
// would not flag anyway): the collision with the vocabulary is a recorded
// decision, not silence.
var agentRefAllowlist = map[string]string{
	"PlaceFact.Source": "state-resident mental-map fact (R2: no AgentRef reachable from State); " +
		"provenance index, not a chronicle surface — applies wherever PlaceFact nests (saw/place_told/place_revealed/map_corrected)",
	"CogToolCallPayload.Args": "opaque json.RawMessage of the tool's own args; " +
		"recorded observability of what the model SAID, never rewritten",
	"IntentSetPayload.Source": "a STRING (\"reflex\"|\"planner\") colliding with the vocabulary by tag only; " +
		"listed so the collision is a recorded decision",
	"FaithChangedPayload.SourceID": "string source-EVENT id (encodes an agent index only for villager_died); " +
		"covered by the additive Agent *AgentRef field instead",
}

// intKinds are the reflect kinds the sweep treats as "bare index" material.
var intKinds = map[reflect.Kind]bool{
	reflect.Int: true, reflect.Int8: true, reflect.Int16: true,
	reflect.Int32: true, reflect.Int64: true,
	reflect.Uint: true, reflect.Uint16: true, reflect.Uint32: true,
	reflect.Uint64: true,
}

// censusMigrationComplete gates the sweep's failure mode while the spec 086
// census migration is in flight (tasks.md T005): false logs violations
// without failing (the sweep is never red mid-census), true makes every
// violation a test failure. Flipped unconditional at T015 once the census
// completes; it never goes back.
const censusMigrationComplete = false

// TestPayloadAgentRefSweep (FR-006, US2 AS-2): reflect over every catalog
// payload type (fields, embedded structs, slices, pointers). Any int-kind
// field (or slice/pointer of one) whose json tag's first segment is in the
// frozen agent vocabulary must instead be AgentRef / []AgentRef /
// *AgentRef, unless (type, field) is on the frozen allowlist. Also enforces
// no dead allowlist entries.
func TestPayloadAgentRefSweep(t *testing.T) {
	report := t.Errorf
	if !censusMigrationComplete {
		report = t.Logf
	}
	liveAllowlist := map[string]bool{}
	seen := map[reflect.Type]bool{}
	for typ, zero := range PayloadCatalog {
		rt := reflect.TypeOf(zero())
		for rt.Kind() == reflect.Pointer {
			rt = rt.Elem()
		}
		sweepType(report, typ, rt, seen, liveAllowlist)
	}
	for key := range agentRefAllowlist {
		if !liveAllowlist[key] {
			report("allowlist entry %q matches no live field — remove the dead exemption (data-model §5)", key)
		}
	}
}

func sweepType(report func(string, ...any), event string, rt reflect.Type, seen map[reflect.Type]bool, live map[string]bool) {
	switch rt.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Array, reflect.Map:
		sweepType(report, event, rt.Elem(), seen, live)
		return
	case reflect.Struct:
	default:
		return
	}
	if seen[rt] {
		return
	}
	seen[rt] = true
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		key := rt.Name() + "." + f.Name
		if _, exempt := agentRefAllowlist[key]; exempt {
			live[key] = true
			continue
		}
		tag := strings.Split(f.Tag.Get("json"), ",")[0]
		if agentVocabulary[tag] && isBareIntish(f.Type) {
			report("%s: field %s (json tag %q) is a bare int-kind agent reference — it must be AgentRef/[]AgentRef/*AgentRef, or earn a frozen allowlist entry (spec 086 FR-002/FR-006)", event, key, tag)
		}
		sweepType(report, event, f.Type, seen, live)
	}
}

// isBareIntish reports whether ft is an int-kind, or a slice/array/pointer
// of one — the shapes a bare agent index (or index list) would take.
func isBareIntish(ft reflect.Type) bool {
	switch ft.Kind() {
	case reflect.Slice, reflect.Array, reflect.Pointer:
		return isBareIntish(ft.Elem())
	default:
		return intKinds[ft.Kind()]
	}
}

// TestPayloadAgentRefSweepCatchesSyntheticViolation proves the sweep's
// teeth (US2 AS-2, SC-002): a payload type with a vocabulary-tagged bare
// int field IS flagged by the same walk the sweep runs.
func TestPayloadAgentRefSweepCatchesSyntheticViolation(t *testing.T) {
	type rogue struct {
		Agent   int   `json:"agent"`
		Targets []int `json:"targets"`
		Qty     int   `json:"qty"` // non-vocabulary — must NOT be flagged
	}
	var hits []string
	record := func(format string, args ...any) { hits = append(hits, format) }
	sweepType(record, "synthetic.rogue", reflect.TypeOf(rogue{}), map[reflect.Type]bool{}, map[string]bool{})
	if len(hits) != 2 { // Agent and Targets flagged; Qty not
		t.Fatalf("the sweep flagged %d fields of the rogue payload, want 2 (agent + targets, not qty)", len(hits))
	}
	type clean struct {
		Agent   AgentRef   `json:"agent"`
		Targets []AgentRef `json:"targets"`
		Qty     int        `json:"qty"`
	}
	hits = nil
	sweepType(record, "synthetic.clean", reflect.TypeOf(clean{}), map[reflect.Type]bool{}, map[string]bool{})
	if len(hits) != 0 {
		t.Fatalf("the sweep flagged a fully migrated payload — false positive: %v", hits)
	}
}

// TestNoAgentRefInState is the R2 hash-stability invariant as a standing
// tripwire (FR-004): no AgentRef may be reachable from sim.State's type
// graph. Names live on the wire only; state entities keep bare ints, so
// State.Marshal()/Hash() for any pre-086 history is byte-identical to
// pre-086 code and world.migrated's embedded State is untouched.
func TestNoAgentRefInState(t *testing.T) {
	seen := map[reflect.Type]bool{}
	var walk func(rt reflect.Type, path string)
	walk = func(rt reflect.Type, path string) {
		switch rt.Kind() {
		case reflect.Pointer, reflect.Slice, reflect.Array:
			walk(rt.Elem(), path)
			return
		case reflect.Map:
			walk(rt.Key(), path+"[key]")
			walk(rt.Elem(), path)
			return
		case reflect.Struct:
		default:
			return
		}
		if rt == agentRefType {
			t.Errorf("AgentRef reachable from sim.State at %s — names must never enter state (research R2)", path)
			return
		}
		if seen[rt] {
			return
		}
		seen[rt] = true
		for i := 0; i < rt.NumField(); i++ {
			f := rt.Field(i)
			walk(f.Type, path+"."+f.Name)
		}
	}
	walk(reflect.TypeOf(State{}), "State")
}

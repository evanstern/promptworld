package bundle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
)

// The effect vocabulary and its compiler (research.md R4, data-model.md). A
// bundle tool — declarative or scripted — expresses its result as a batch of
// effects over this closed vocabulary; CompileEffects is the SOLE factory that
// turns an effect into a store.Event, so a tool can never invent an event type
// or hand-craft a payload. The compiled payloads reuse the sim.* miracle payload
// structs verbatim (miracle_batch.go's pattern), so a bundle-tool batch is
// byte-identical to the same edit the built-in miracle door would land and
// replays cleanly through the same reducer arms.

// Batch and value caps (data-model.md "Batch rules"). narrateMaxBytes is also
// the default text-param budget (manifest.go).
const (
	maxBatchEvents  = 32
	narrateMaxBytes = 500
	qtyMin          = 1
	qtyMax          = 99
)

// effectEventType maps each effect kind to the single event type it produces.
// narrate expands to one agent.memory_added per recipient; the others produce
// exactly one metatron.* event.
var effectEventType = map[string]string{
	"move_entity":   "metatron.entity_moved",
	"remove_entity": "metatron.entity_removed",
	"grant_item":    "metatron.item_granted",
	"snap_time":     "metatron.time_snapped",
	"narrate":       "agent.memory_added",
}

// Effect is one resolved effect — script mode returns these directly; declarative
// mode reaches them via template expansion. Only the fields its Kind uses are
// meaningful. Target addresses a living villager by name (v1): move/remove/grant
// resolve it to the villager's index and current tile, mirroring the villager
// path in BuildMiracleBatch.
type Effect struct {
	Kind       string
	Target     string
	ToX, ToY   int
	Item       string
	Qty        int
	ToTick     int64
	Text       string
	Recipients recipientSpec
}

// recipientSpec is a resolved narrate audience: every living villager, or a
// specific set named by villager name.
type recipientSpec struct {
	all   bool
	names []string
}

// CompileInput is the per-invocation context the compiler resolves against.
// Declared is the manifest's event set; every produced event type must be a
// member (the invocation-time subset gate). Args carries the resolved,
// stringified argument values used for {args.x} substitution and the "target"
// recipient keyword. Tick is the current world tick (reserved for script-mode
// rng; the reducer remains authoritative for tick-forward validation).
type CompileInput struct {
	State    *sim.State
	Tick     int64
	Args     map[string]string
	Invoker  string
	Declared map[string]bool
}

// effectTemplate is a parsed-but-still-templated declarative effect: string
// fields may hold {args.x}/{invoker} placeholders and numeric fields may be a
// literal integer or a placeholder string, both resolved at expand time.
type effectTemplate struct {
	kind       string
	target     string
	toX, toY   json.RawMessage
	item       string
	qty        json.RawMessage
	toTick     json.RawMessage
	text       string
	recipients recipientTemplate
}

// recipientTemplate is the unresolved narrate audience: a keyword
// ("all_living"/"target") or a list of name templates.
type recipientTemplate struct {
	keyword string
	names   []string
}

// ParseTemplates strict-decodes the manifest's raw effects array into
// effectTemplates, per-kind (each kind rejects unknown keys). It validates
// STRUCTURE only — kind known, required fields present, recipient shape well
// formed — because value resolution depends on invocation args. Errors are
// ruleError-tagged T5 and name the offending effect index and field.
func ParseTemplates(raw json.RawMessage) ([]effectTemplate, error) {
	var entries []json.RawMessage
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&entries); err != nil {
		return nil, ruleErr("T5", "effects is not a JSON array: %v", err)
	}
	if len(entries) == 0 {
		return nil, ruleErr("T5", "effects is empty (declare at least one effect)")
	}
	out := make([]effectTemplate, 0, len(entries))
	for i, e := range entries {
		t, err := parseTemplate(i, e)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, nil
}

// producibleEvents returns the set of event types the templates can produce —
// the union of each kind's event type. Used by boot validation (T5) to prove
// producible == declared.
func producibleEvents(templates []effectTemplate) map[string]bool {
	out := make(map[string]bool, len(templates))
	for _, t := range templates {
		out[effectEventType[t.kind]] = true
	}
	return out
}

// parseTemplate decodes one effect entry against its kind-specific schema.
func parseTemplate(idx int, raw json.RawMessage) (effectTemplate, error) {
	var probe struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return effectTemplate{}, ruleErr("T5", "effect %d: does not decode: %v", idx, err)
	}
	if _, ok := effectEventType[probe.Kind]; !ok {
		return effectTemplate{}, ruleErr("T5", "effect %d: unknown kind %q", idx, probe.Kind)
	}
	t := effectTemplate{kind: probe.Kind}

	strictInto := func(v any) error {
		d := json.NewDecoder(bytes.NewReader(raw))
		d.DisallowUnknownFields()
		return d.Decode(v)
	}

	switch probe.Kind {
	case "move_entity":
		var s struct {
			Kind   string          `json:"kind"`
			Target string          `json:"target"`
			ToX    json.RawMessage `json:"to_x"`
			ToY    json.RawMessage `json:"to_y"`
		}
		if err := strictInto(&s); err != nil {
			return effectTemplate{}, ruleErr("T5", "effect %d (move_entity): %v", idx, err)
		}
		if s.Target == "" {
			return effectTemplate{}, ruleErr("T5", "effect %d (move_entity): target is required", idx)
		}
		if len(s.ToX) == 0 || len(s.ToY) == 0 {
			return effectTemplate{}, ruleErr("T5", "effect %d (move_entity): to_x and to_y are required", idx)
		}
		t.target, t.toX, t.toY = s.Target, s.ToX, s.ToY
	case "remove_entity":
		var s struct {
			Kind   string `json:"kind"`
			Target string `json:"target"`
		}
		if err := strictInto(&s); err != nil {
			return effectTemplate{}, ruleErr("T5", "effect %d (remove_entity): %v", idx, err)
		}
		if s.Target == "" {
			return effectTemplate{}, ruleErr("T5", "effect %d (remove_entity): target is required", idx)
		}
		t.target = s.Target
	case "grant_item":
		var s struct {
			Kind   string          `json:"kind"`
			Target string          `json:"target"`
			Item   string          `json:"item"`
			Qty    json.RawMessage `json:"qty"`
		}
		if err := strictInto(&s); err != nil {
			return effectTemplate{}, ruleErr("T5", "effect %d (grant_item): %v", idx, err)
		}
		if s.Target == "" || s.Item == "" || len(s.Qty) == 0 {
			return effectTemplate{}, ruleErr("T5", "effect %d (grant_item): target, item, and qty are required", idx)
		}
		t.target, t.item, t.qty = s.Target, s.Item, s.Qty
	case "snap_time":
		var s struct {
			Kind   string          `json:"kind"`
			ToTick json.RawMessage `json:"to_tick"`
		}
		if err := strictInto(&s); err != nil {
			return effectTemplate{}, ruleErr("T5", "effect %d (snap_time): %v", idx, err)
		}
		if len(s.ToTick) == 0 {
			return effectTemplate{}, ruleErr("T5", "effect %d (snap_time): to_tick is required", idx)
		}
		t.toTick = s.ToTick
	case "narrate":
		var s struct {
			Kind       string          `json:"kind"`
			Text       string          `json:"text"`
			Recipients json.RawMessage `json:"recipients"`
		}
		if err := strictInto(&s); err != nil {
			return effectTemplate{}, ruleErr("T5", "effect %d (narrate): %v", idx, err)
		}
		if s.Text == "" {
			return effectTemplate{}, ruleErr("T5", "effect %d (narrate): text is required", idx)
		}
		rt, err := parseRecipients(idx, s.Recipients)
		if err != nil {
			return effectTemplate{}, err
		}
		t.text, t.recipients = s.Text, rt
	}
	return t, nil
}

// parseRecipients decodes a narrate recipients field: the string keywords
// "all_living"/"target", or a non-empty JSON array of name templates.
func parseRecipients(idx int, raw json.RawMessage) (recipientTemplate, error) {
	if len(raw) == 0 {
		return recipientTemplate{}, ruleErr("T5", "effect %d (narrate): recipients is required", idx)
	}
	trimmed := bytes.TrimSpace(raw)
	switch trimmed[0] {
	case '"':
		var kw string
		if err := json.Unmarshal(trimmed, &kw); err != nil {
			return recipientTemplate{}, ruleErr("T5", "effect %d (narrate): recipients: %v", idx, err)
		}
		if kw != "all_living" && kw != "target" {
			return recipientTemplate{}, ruleErr("T5", "effect %d (narrate): recipients keyword %q must be \"all_living\" or \"target\" (or a list of names)", idx, kw)
		}
		return recipientTemplate{keyword: kw}, nil
	case '[':
		var names []string
		if err := json.Unmarshal(trimmed, &names); err != nil {
			return recipientTemplate{}, ruleErr("T5", "effect %d (narrate): recipients: %v", idx, err)
		}
		if len(names) == 0 {
			return recipientTemplate{}, ruleErr("T5", "effect %d (narrate): recipients list is empty", idx)
		}
		return recipientTemplate{names: names}, nil
	default:
		return recipientTemplate{}, ruleErr("T5", "effect %d (narrate): recipients must be a keyword string or a list of names", idx)
	}
}

// ExpandTemplates resolves templates into concrete Effects for one invocation:
// {args.x}/{invoker} substitution on string fields, integer resolution on
// numeric fields, and recipient keyword/name resolution. Errors name the effect
// index, field, and offending value (SC-005).
func ExpandTemplates(templates []effectTemplate, in CompileInput) ([]Effect, error) {
	out := make([]Effect, 0, len(templates))
	for i, t := range templates {
		e := Effect{Kind: t.kind}
		var err error
		switch t.kind {
		case "move_entity":
			if e.Target, err = subst(i, "target", t.target, in); err != nil {
				return nil, err
			}
			if e.ToX, err = resolveInt(i, "to_x", t.toX, in); err != nil {
				return nil, err
			}
			if e.ToY, err = resolveInt(i, "to_y", t.toY, in); err != nil {
				return nil, err
			}
		case "remove_entity":
			if e.Target, err = subst(i, "target", t.target, in); err != nil {
				return nil, err
			}
		case "grant_item":
			if e.Target, err = subst(i, "target", t.target, in); err != nil {
				return nil, err
			}
			if e.Item, err = subst(i, "item", t.item, in); err != nil {
				return nil, err
			}
			if e.Qty, err = resolveInt(i, "qty", t.qty, in); err != nil {
				return nil, err
			}
		case "snap_time":
			var v int
			if v, err = resolveInt(i, "to_tick", t.toTick, in); err != nil {
				return nil, err
			}
			e.ToTick = int64(v)
		case "narrate":
			if e.Text, err = subst(i, "text", t.text, in); err != nil {
				return nil, err
			}
			if e.Recipients, err = resolveRecipients(i, t.recipients, in); err != nil {
				return nil, err
			}
		}
		out = append(out, e)
	}
	return out, nil
}

// resolveRecipients turns a recipient template into a resolved spec. The
// "target" keyword resolves to the invocation's `target` argument (the contract
// convention); a name list substitutes each entry.
func resolveRecipients(idx int, rt recipientTemplate, in CompileInput) (recipientSpec, error) {
	switch {
	case rt.keyword == "all_living":
		return recipientSpec{all: true}, nil
	case rt.keyword == "target":
		name, ok := in.Args["target"]
		if !ok || name == "" {
			return recipientSpec{}, ruleErr("T5", "effect %d (narrate): recipients \"target\" needs a \"target\" argument, which was not provided", idx)
		}
		return recipientSpec{names: []string{name}}, nil
	default:
		names := make([]string, 0, len(rt.names))
		for j, n := range rt.names {
			v, err := subst(idx, fmt.Sprintf("recipients[%d]", j), n, in)
			if err != nil {
				return recipientSpec{}, err
			}
			names = append(names, v)
		}
		return recipientSpec{names: names}, nil
	}
}

// CompileEffects is the sole effect→store.Event factory. It resolves each
// effect's target/recipients against live state, builds the payload with the
// canonical sim.* struct (Gratis=false — bundle tools never waive the reducer's
// charge), enforces the batch cap, and rejects any produced event type outside
// the declared set (the invocation-time subset gate). It does NOT validate
// reducer preconditions (entity presence, tick-forward, charge sufficiency);
// those stay reducer-authoritative and are enforced by InjectSocial's dry run.
func CompileEffects(effects []Effect, in CompileInput) ([]store.Event, error) {
	batch := make([]store.Event, 0, len(effects))
	for i, e := range effects {
		evs, err := compileOne(i, e, in)
		if err != nil {
			return nil, err
		}
		batch = append(batch, evs...)
		if len(batch) > maxBatchEvents {
			return nil, ruleErr("T5", "effect batch exceeds %d events after expansion", maxBatchEvents)
		}
	}
	for _, ev := range batch {
		if !in.Declared[ev.Type] {
			return nil, ruleErr("T5", "produced event %q is not in the tool's declared events", ev.Type)
		}
	}
	return batch, nil
}

// compileOne compiles a single effect into its event(s).
func compileOne(idx int, e Effect, in CompileInput) ([]store.Event, error) {
	switch e.Kind {
	case "move_entity":
		_, x, y, err := villager(idx, e.Target, in)
		if err != nil {
			return nil, err
		}
		return []store.Event{event("metatron.entity_moved", sim.EntityMovedPayload{
			Class: "villager", X: x, Y: y, ToX: e.ToX, ToY: e.ToY, Gratis: false})}, nil
	case "remove_entity":
		_, x, y, err := villager(idx, e.Target, in)
		if err != nil {
			return nil, err
		}
		return []store.Event{event("metatron.entity_removed", sim.EntityRemovedPayload{
			Class: "villager", X: x, Y: y, Gratis: false})}, nil
	case "grant_item":
		i, _, _, err := villager(idx, e.Target, in)
		if err != nil {
			return nil, err
		}
		if e.Qty < qtyMin || e.Qty > qtyMax {
			return nil, ruleErr("T5", "effect %d (grant_item): qty %d out of range %d–%d", idx, e.Qty, qtyMin, qtyMax)
		}
		return []store.Event{event("metatron.item_granted", sim.ItemGrantedPayload{
			Agent: i, Kind: e.Item, Qty: e.Qty, Gratis: false})}, nil
	case "snap_time":
		return []store.Event{event("metatron.time_snapped", sim.TimeSnappedPayload{
			ToTick: e.ToTick, Gratis: false})}, nil
	case "narrate":
		if n := len(e.Text); n > narrateMaxBytes {
			return nil, ruleErr("T5", "effect %d (narrate): text is %d bytes (max %d)", idx, n, narrateMaxBytes)
		}
		recips, err := resolveAudience(idx, e.Recipients, in)
		if err != nil {
			return nil, err
		}
		out := make([]store.Event, 0, len(recips))
		for _, r := range recips {
			out = append(out, event("agent.memory_added", sim.MemoryAddedPayload{
				Agent: r, Text: e.Text, Salience: sim.SalDream, Subject: -1, Origin: sim.OriginOmen}))
		}
		return out, nil
	}
	return nil, ruleErr("T5", "effect %d: unknown kind %q", idx, e.Kind)
}

// resolveAudience turns a resolved recipient spec into living-villager indices,
// ascending and deterministic. all_living reads the living roster; a named set
// resolves each name, erroring on any that is not a living villager (an explicit
// name is intent — a silent drop would hide an authoring bug).
func resolveAudience(idx int, spec recipientSpec, in CompileInput) ([]int, error) {
	if spec.all {
		return in.State.LivingAgents(), nil
	}
	out := make([]int, 0, len(spec.names))
	for _, name := range spec.names {
		i := villagerIndex(in.State, name)
		if i < 0 {
			return nil, ruleErr("T5", "effect %d (narrate): recipient %q is not a living villager", idx, name)
		}
		out = append(out, i)
	}
	return out, nil
}

// villager resolves a target name to a living villager's index and current tile.
func villager(idx int, name string, in CompileInput) (i, x, y int, err error) {
	i = villagerIndex(in.State, name)
	if i < 0 {
		return 0, 0, 0, ruleErr("T5", "effect %d: no living villager named %q", idx, name)
	}
	a := in.State.Agents[i]
	return i, a.X, a.Y, nil
}

// villagerIndex returns the index of the living villager with the given name
// (case-insensitive, trimmed), or -1. Reads sim.State.Agents directly (its
// fields are exported); deterministic first-match by roster index.
func villagerIndex(s *sim.State, name string) int {
	want := strings.ToLower(strings.TrimSpace(name))
	for i := range s.Agents {
		if s.Agents[i].Dead {
			continue
		}
		if strings.ToLower(s.Agents[i].Name) == want {
			return i
		}
	}
	return -1
}

// event builds a store.Event with a canonically marshaled payload. Tick/Seq are
// stamped by InjectSocial/the store at landing, exactly as BuildMiracleBatch
// leaves them zero.
func event(typ string, payload any) store.Event {
	b, err := json.Marshal(payload)
	if err != nil {
		panic(err) // sim payload structs always marshal
	}
	return store.Event{Type: typ, Payload: b}
}

// subst resolves {args.<name>} and {invoker} placeholders in a template string.
// No nesting, no expressions; an unknown or unclosed placeholder is an error
// naming the effect index and field (SC-005). A literal '{' therefore always
// begins a placeholder — authored effect strings must not contain raw braces.
func subst(idx int, field, s string, in CompileInput) (string, error) {
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] != '{' {
			b.WriteByte(s[i])
			i++
			continue
		}
		end := strings.IndexByte(s[i:], '}')
		if end < 0 {
			return "", ruleErr("T5", "effect %d field %q: unclosed '{' in %q", idx, field, s)
		}
		token := s[i+1 : i+end]
		val, err := resolveToken(idx, field, token, in)
		if err != nil {
			return "", err
		}
		b.WriteString(val)
		i += end + 1
	}
	return b.String(), nil
}

// resolveToken resolves one placeholder token: "invoker" or "args.<name>".
func resolveToken(idx int, field, token string, in CompileInput) (string, error) {
	if token == "invoker" {
		return in.Invoker, nil
	}
	if name, ok := strings.CutPrefix(token, "args."); ok {
		v, present := in.Args[name]
		if !present {
			return "", ruleErr("T5", "effect %d field %q: unresolved placeholder {args.%s} (no such argument)", idx, field, name)
		}
		return v, nil
	}
	return "", ruleErr("T5", "effect %d field %q: unknown placeholder {%s} (only {invoker} and {args.NAME} are allowed)", idx, field, token)
}

// resolveInt resolves a numeric effect field that may be a JSON integer literal
// or a placeholder string that substitutes to an integer. Floats (a '.'/'e' in
// a literal, or a non-integer substituted value) are rejected — payload numeric
// fields are ints only, so no float/NaN/Inf can ever reach a payload.
func resolveInt(idx int, field string, raw json.RawMessage, in CompileInput) (int, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return 0, ruleErr("T5", "effect %d field %q: missing value", idx, field)
	}
	if trimmed[0] == '"' {
		var s string
		if err := json.Unmarshal(trimmed, &s); err != nil {
			return 0, ruleErr("T5", "effect %d field %q: %v", idx, field, err)
		}
		s, err := subst(idx, field, s, in)
		if err != nil {
			return 0, err
		}
		n, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil {
			return 0, ruleErr("T5", "effect %d field %q: %q is not an integer", idx, field, s)
		}
		return n, nil
	}
	lit := string(trimmed)
	if strings.ContainsAny(lit, ".eE") {
		return 0, ruleErr("T5", "effect %d field %q: %s must be an integer, not a float", idx, field, lit)
	}
	n, err := strconv.Atoi(lit)
	if err != nil {
		return 0, ruleErr("T5", "effect %d field %q: %s is not an integer", idx, field, lit)
	}
	return n, nil
}

package bundle

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"go.starlark.net/starlark"
)

// The Starlark executor (spec 036 US3, contracts/script-api.md). A scripted
// bundle tool defines a pure apply(args, world) that computes an effect batch; the
// runtime here compiles that script ONCE at boot, then runs it per invocation on a
// fresh, step-capped, capability-free thread. The returned Starlark list is
// converted to resolved Effects (float/NaN/Inf and shape rejection happen HERE —
// declarative JSON can't carry those, so the conversion layer is the only place a
// script can smuggle them in) and handed to CompileEffects, the SAME shared core
// the declarative path uses for the closed vocabulary, batch caps, and the
// declared-events subset gate. Determinism is structural: Starlark evaluates
// deterministically, the thread has NO ambient capabilities (no clock, io, net,
// module loading, or print-into-state), and world.rand is coordinate-seeded — so
// identical (args, world snapshot, seed, tick) always yields the identical batch,
// and replay never re-executes scripts at all (events are self-contained data).

// scriptProgram is a tool.star compiled once at boot: its apply() function frozen
// and cached, ready to run against fresh (args, world) on a per-invocation thread.
// A frozen Starlark function is safe to call repeatedly (guardian turns are
// single-flight regardless).
type scriptProgram struct {
	apply *starlark.Function
	name  string // script filename, for error messages
}

// compileScript is BOTH the boot-time T6 gate and the compile-once cache builder:
// it parses tool.star, executes its top level (defining apply) under the hard step
// ceiling, freezes the module globals, and retains the apply function. No ambient
// names are predeclared — args/world arrive as apply's parameters — so a top-level
// reference to them fails to resolve at boot rather than surprising at invocation.
// Module loading is disabled (thread.Load nil) and print never reaches world state.
func compileScript(toolDir, scriptName string) (*scriptProgram, error) {
	path := filepath.Join(toolDir, scriptName)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("script file %q is missing or unreadable", scriptName)
	}
	_, prog, err := starlark.SourceProgram(scriptName, data, func(string) bool { return false })
	if err != nil {
		return nil, fmt.Errorf("script %q does not parse: %v", scriptName, err)
	}
	// Bound top-level execution too: a hostile script cannot burn unbounded work at
	// boot, and Init only runs the def statements for a well-formed tool.
	thread := newScriptThread(scriptName, maxStepsCeiling)
	globals, err := prog.Init(thread, nil)
	if err != nil {
		return nil, fmt.Errorf("script %q failed to load: %v", scriptName, err)
	}
	globals.Freeze()
	v, ok := globals["apply"]
	if !ok {
		return nil, fmt.Errorf("script %q does not define apply()", scriptName)
	}
	fn, ok := v.(*starlark.Function)
	if !ok {
		return nil, fmt.Errorf("script %q: apply must be a function", scriptName)
	}
	if fn.NumParams() != 2 || fn.HasVarargs() || fn.HasKwargs() {
		return nil, fmt.Errorf("script %q: apply must take exactly (args, world)", scriptName)
	}
	return &scriptProgram{apply: fn, name: scriptName}, nil
}

// newScriptThread builds a sandboxed thread: step-capped, no module loading, and
// print routed to the daemon log (never into world state). Recursion is off by the
// resolver's default, and no builtins beyond the Starlark universe are reachable —
// there is no clock, os, io, net, or module surface to touch. maxSteps is the
// budget the interpreter aborts at deterministically ("too many steps").
func newScriptThread(name string, maxSteps uint64) *starlark.Thread {
	th := &starlark.Thread{
		Name:  "bundle:" + name,
		Load:  nil, // load() disabled: no module surface
		Print: func(_ *starlark.Thread, msg string) { log.Printf("bundle script %s: %s", name, msg) },
	}
	th.SetMaxExecutionSteps(maxSteps)
	return th
}

// execute runs apply(args, world) on a fresh step-capped thread and converts its
// return value into resolved Effects. Any Starlark failure — fail(), a type
// error, or a deterministic step-cap abort — returns an error (never a panic) that
// the handler feeds back as an in-fiction rejection: nothing lands, no charge
// spent. The subsequent CompileEffects call applies the declared-events gate and
// batch caps, so this layer only proves the RETURN SHAPE and rejects non-finite /
// non-integer numerics the JSON path structurally cannot produce.
func (sp *scriptProgram) execute(argsDict *starlark.Dict, world starlark.Value, maxSteps uint64) ([]Effect, error) {
	thread := newScriptThread(sp.name, maxSteps)
	ret, err := starlark.Call(thread, sp.apply, starlark.Tuple{argsDict, world}, nil)
	if err != nil {
		return nil, err
	}
	return effectsFromValue(ret)
}

// scriptArgs projects a validated tool call's arguments into the frozen Starlark
// dict apply() receives: agent_name/text/enum params as strings, number params as
// ints (contracts/script-api.md "Inputs"). The toolloop already schema-validated
// types against the derived schema, so this is a lenient reader; a returned error
// is defense-in-depth for a shape that slipped through.
func scriptArgs(m *Manifest, raw json.RawMessage) (*starlark.Dict, error) {
	obj := map[string]json.RawMessage{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &obj); err != nil {
			return nil, fmt.Errorf("arguments must be a JSON object")
		}
	}
	d := starlark.NewDict(len(m.Params))
	for _, p := range m.Params {
		rawv, ok := obj[p.Name]
		if !ok {
			continue // optional-and-absent; a required-but-absent arg was rejected upstream
		}
		var val starlark.Value
		if p.Kind == "number" {
			var f float64
			if err := json.Unmarshal(rawv, &f); err != nil {
				return nil, fmt.Errorf("argument %q must be a number", p.Name)
			}
			val = starlark.MakeInt64(int64(f))
		} else {
			var s string
			if err := json.Unmarshal(rawv, &s); err != nil {
				return nil, fmt.Errorf("argument %q must be a string", p.Name)
			}
			val = starlark.String(s)
		}
		if err := d.SetKey(starlark.String(p.Name), val); err != nil {
			return nil, err
		}
	}
	d.Freeze()
	return d, nil
}

// effectsFromValue converts apply()'s return value into []Effect. The result MUST
// be a list; each element MUST be a dict with a known kind and exactly that kind's
// fields (unknown fields rejected). An empty list is a valid no-op (the compiler
// accepts an empty batch). Numeric fields must be finite integers — any float
// (including float('nan')/float('inf')) is rejected here, closing the only door
// through which a non-finite value could reach a payload.
func effectsFromValue(v starlark.Value) ([]Effect, error) {
	list, ok := v.(*starlark.List)
	if !ok {
		return nil, fmt.Errorf("apply() must return a list of effect dicts, got %s", v.Type())
	}
	out := make([]Effect, 0, list.Len())
	for i := 0; i < list.Len(); i++ {
		e, err := effectFromDict(i, list.Index(i))
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}

// effectFromDict converts one returned dict into a resolved Effect, enforcing the
// per-kind field schema and rejecting unknown fields (with the element index in
// every message, per SC-005).
func effectFromDict(idx int, v starlark.Value) (Effect, error) {
	d, ok := v.(*starlark.Dict)
	if !ok {
		return Effect{}, fmt.Errorf("effect %d: must be a dict, got %s", idx, v.Type())
	}
	kind, present, err := dictStr(d, "kind")
	if err != nil {
		return Effect{}, fmt.Errorf("effect %d: %v", idx, err)
	}
	if !present {
		return Effect{}, fmt.Errorf("effect %d: missing \"kind\"", idx)
	}
	if _, known := effectEventType[kind]; !known {
		return Effect{}, fmt.Errorf("effect %d: unknown kind %q", idx, kind)
	}
	e := Effect{Kind: kind}
	var allowed []string
	switch kind {
	case "move_entity":
		allowed = []string{"kind", "target", "to_x", "to_y"}
		if e.Target, err = reqStr(idx, d, "target"); err != nil {
			return Effect{}, err
		}
		if e.ToX, err = reqInt(idx, d, "to_x"); err != nil {
			return Effect{}, err
		}
		if e.ToY, err = reqInt(idx, d, "to_y"); err != nil {
			return Effect{}, err
		}
	case "remove_entity":
		allowed = []string{"kind", "target"}
		if e.Target, err = reqStr(idx, d, "target"); err != nil {
			return Effect{}, err
		}
	case "grant_item":
		allowed = []string{"kind", "target", "item", "qty"}
		if e.Target, err = reqStr(idx, d, "target"); err != nil {
			return Effect{}, err
		}
		if e.Item, err = reqStr(idx, d, "item"); err != nil {
			return Effect{}, err
		}
		if e.Qty, err = reqInt(idx, d, "qty"); err != nil {
			return Effect{}, err
		}
	case "snap_time":
		allowed = []string{"kind", "to_tick"}
		tt, ierr := reqInt(idx, d, "to_tick")
		if ierr != nil {
			return Effect{}, ierr
		}
		e.ToTick = int64(tt)
	case "narrate":
		if e.Text, err = reqStr(idx, d, "text"); err != nil {
			return Effect{}, err
		}
		spec, usesTargetField, rerr := scriptRecipients(idx, d)
		if rerr != nil {
			return Effect{}, rerr
		}
		e.Recipients = spec
		allowed = []string{"kind", "text", "recipients"}
		if usesTargetField {
			allowed = append(allowed, "target")
		}
	}
	if err := rejectUnknownKeys(idx, d, allowed); err != nil {
		return Effect{}, err
	}
	return e, nil
}

// scriptRecipients resolves a narrate dict's recipients into a recipientSpec. It
// mirrors the declarative recipient vocabulary: the "all_living"/"target"
// keywords or a non-empty list of names. The "target" keyword takes its name from
// a companion "target" field on the dict (contracts/script-api.md example), so the
// return's second value reports whether that field is expected — the caller adds
// it to the allowed-keys set only then.
func scriptRecipients(idx int, d *starlark.Dict) (recipientSpec, bool, error) {
	rv, found, _ := d.Get(starlark.String("recipients"))
	if !found {
		return recipientSpec{}, false, fmt.Errorf("effect %d (narrate): recipients is required", idx)
	}
	switch r := rv.(type) {
	case starlark.String:
		switch string(r) {
		case "all_living":
			return recipientSpec{all: true}, false, nil
		case "target":
			name, ok, serr := dictStr(d, "target")
			if serr != nil {
				return recipientSpec{}, true, fmt.Errorf("effect %d (narrate): %v", idx, serr)
			}
			if !ok || name == "" {
				return recipientSpec{}, true, fmt.Errorf("effect %d (narrate): recipients \"target\" needs a non-empty \"target\" field", idx)
			}
			return recipientSpec{names: []string{name}}, true, nil
		default:
			return recipientSpec{}, false, fmt.Errorf("effect %d (narrate): recipients keyword %q must be \"all_living\" or \"target\" (or a list of names)", idx, string(r))
		}
	case *starlark.List:
		names := make([]string, 0, r.Len())
		for i := 0; i < r.Len(); i++ {
			s, ok := r.Index(i).(starlark.String)
			if !ok {
				return recipientSpec{}, false, fmt.Errorf("effect %d (narrate): recipients[%d] must be a string, got %s", idx, i, r.Index(i).Type())
			}
			if string(s) == "" {
				return recipientSpec{}, false, fmt.Errorf("effect %d (narrate): recipients[%d] is empty", idx, i)
			}
			names = append(names, string(s))
		}
		if len(names) == 0 {
			return recipientSpec{}, false, fmt.Errorf("effect %d (narrate): recipients list is empty", idx)
		}
		return recipientSpec{names: names}, false, nil
	default:
		return recipientSpec{}, false, fmt.Errorf("effect %d (narrate): recipients must be a keyword string or a list of names, got %s", idx, rv.Type())
	}
}

// dictStr reads an optional string field: (value, present, error). A present
// non-string value is an error.
func dictStr(d *starlark.Dict, key string) (string, bool, error) {
	v, found, err := d.Get(starlark.String(key))
	if err != nil || !found {
		return "", false, err
	}
	s, ok := v.(starlark.String)
	if !ok {
		return "", true, fmt.Errorf("field %q must be a string, got %s", key, v.Type())
	}
	return string(s), true, nil
}

// reqStr reads a required, non-empty string field.
func reqStr(idx int, d *starlark.Dict, key string) (string, error) {
	s, present, err := dictStr(d, key)
	if err != nil {
		return "", fmt.Errorf("effect %d: %v", idx, err)
	}
	if !present || s == "" {
		return "", fmt.Errorf("effect %d: field %q is required", idx, key)
	}
	return s, nil
}

// reqInt reads a required integer field. A float (including NaN/Inf) or any
// non-int value is rejected — payload numerics are ints only, so no non-finite
// value can ever reach an event.
func reqInt(idx int, d *starlark.Dict, key string) (int, error) {
	v, found, err := d.Get(starlark.String(key))
	if err != nil {
		return 0, fmt.Errorf("effect %d: %v", idx, err)
	}
	if !found {
		return 0, fmt.Errorf("effect %d: field %q is required", idx, key)
	}
	iv, ok := v.(starlark.Int)
	if !ok {
		return 0, fmt.Errorf("effect %d: field %q must be an integer, got %s", idx, key, v.Type())
	}
	n, ok := iv.Int64()
	if !ok {
		return 0, fmt.Errorf("effect %d: field %q integer is out of range", idx, key)
	}
	return int(n), nil
}

// rejectUnknownKeys fails if the dict carries any field outside the kind's schema.
func rejectUnknownKeys(idx int, d *starlark.Dict, allowed []string) error {
	set := make(map[string]bool, len(allowed))
	for _, k := range allowed {
		set[k] = true
	}
	for _, kv := range d.Keys() {
		ks, ok := kv.(starlark.String)
		if !ok {
			return fmt.Errorf("effect %d: dict keys must be strings", idx)
		}
		if !set[string(ks)] {
			return fmt.Errorf("effect %d: unknown field %q for this effect kind", idx, string(ks))
		}
	}
	return nil
}

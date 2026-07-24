package bundle

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/toolloop"
)

// The bundle handler factory (spec 036 US1, T011): one toolloop.Handler per
// bundle tool, wrapping the SAME landing pipeline the loader validated at boot —
// resolve arguments → expand templates (declarative mode) → compile effects →
// declared-events subset check (all in CompileEffects) → InjectSocial. The
// factory mirrors internal/metatron/toolcalls.go: it never mutates state itself,
// it lands through the injected door, and it translates every author-level
// failure into a rejected_gate Outcome the model may repair within the loop's
// round cap — never a toolloop.Outcome.Err (that is reserved for infrastructure
// failures, which terminate the whole loop).
//
// The dependency shape (InvocationContext) is the per-turn surface the metatron
// turn assembly already holds: a read snapshot of sim state (for target /
// recipient resolution, the same probe landMiracle builds), the current tick,
// the invoker's name ({invoker} substitution), and the InjectSocial door as a
// plain function value. Passing these as data keeps bundle importable by
// metatron with no reverse dependency.

// InvocationContext is the per-turn context a bundle handler resolves against.
// State is a read-only snapshot whose Agents carry Name/X/Y/Dead so the effect
// compiler can resolve a target villager and expand an all_living / named
// narration audience; it is never mutated. Inject is the InjectSocial door
// (mt.social.InjectSocial), the SOLE path a bundle batch reaches the world.
type InvocationContext struct {
	State   *sim.State
	Tick    int64
	Invoker string
	Inject  func([]store.Event) error
}

// Handlers builds one toolloop.Handler per bundle tool for a single invocation
// context. The metatron turn assembly calls this once per turn (like
// mt.turnHandlers) and merges the result into its handler map; grant filtering
// is the caller's job (the map is keyed by tool name, so an ungranted tool is
// simply not copied over).
func (bs *BundleSet) Handlers(ic InvocationContext) map[string]toolloop.Handler {
	h := make(map[string]toolloop.Handler)
	for _, b := range bs.bundles {
		for i := range b.Tools {
			bt := b.Tools[i] // capture by value; each handler closes over its own tool
			h[bt.Name] = bs.handlerFor(bt, ic)
		}
	}
	return h
}

// handlerFor builds the handler for one bundle tool. The declarative path
// expands the cached templates; script mode (Phase 5) is not yet wired, so a
// script tool refuses in-fiction rather than landing — it is a validated roster
// member, so it must have a handler (the loop rejects a tool with none as
// rejected_unknown, which would be the wrong story).
func (bs *BundleSet) handlerFor(bt BundleTool, ic InvocationContext) toolloop.Handler {
	return func(_ context.Context, call llm.ToolCall) toolloop.Outcome {
		if bt.Manifest.scriptMode() {
			return reject("this tool's script runtime is not yet available")
		}
		args, why := stringArgs(bt.Manifest, call.Args)
		if why != "" {
			return reject(why)
		}
		in := CompileInput{
			State:    ic.State,
			Tick:     ic.Tick,
			Args:     args,
			Invoker:  ic.Invoker,
			Declared: declaredSet(bt.Manifest.Events),
		}
		effects, err := ExpandTemplates(bt.Templates, in)
		if err != nil {
			return reject(err.Error())
		}
		batch, err := CompileEffects(effects, in)
		if err != nil {
			return reject(err.Error())
		}
		if err := ic.Inject(batch); err != nil {
			// The door refused the batch (probe dry-run: entity absent, tick not
			// forward, insufficient charge). Feed the reason back so the model can
			// try a different act; nothing landed and nothing was spent.
			return reject("the world refused it (" + err.Error() + ")")
		}
		return toolloop.Outcome{Verdict: toolloop.VerdictLanded, ResultForModel: "the " + bt.Name + " took effect"}
	}
}

// reject wraps an author-level failure as a rejected_gate Outcome — the door's
// in-fiction refusal, correctable within the loop's round cap. Never an
// Outcome.Err: an author's bad target or undeclared event is not an
// infrastructure failure and must not terminate the loop.
func reject(why string) toolloop.Outcome {
	return toolloop.Outcome{Verdict: toolloop.VerdictRejectedGate, ResultForModel: why}
}

// declaredSet renders a manifest's event list as a membership set (the
// invocation-time subset gate CompileEffects enforces).
func declaredSet(events []string) map[string]bool {
	m := make(map[string]bool, len(events))
	for _, e := range events {
		m[e] = true
	}
	return m
}

// stringArgs projects a tool call's JSON arguments into the stringified map the
// template substituter reads ({args.<name>}). The toolloop already
// schema-validated types against the derived schema, so this is a lenient
// reader; a number param is rendered as its integer literal so a numeric
// effect field ({args.x} into to_x) resolves cleanly. A returned reason (non-"")
// is fed back as a rejected_gate — a shape the toolloop's own validation would
// already have caught, kept here as defense-in-depth.
func stringArgs(m *Manifest, raw json.RawMessage) (map[string]string, string) {
	obj := map[string]json.RawMessage{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &obj); err != nil {
			return nil, "arguments must be a JSON object"
		}
	}
	out := make(map[string]string, len(m.Params))
	for _, p := range m.Params {
		rawv, ok := obj[p.Name]
		if !ok {
			continue // optional and absent; a required-but-absent arg was already rejected upstream
		}
		if p.Kind == "number" {
			var f float64
			if err := json.Unmarshal(rawv, &f); err != nil {
				return nil, fmt.Sprintf("argument %q must be a number", p.Name)
			}
			out[p.Name] = strconv.FormatInt(int64(f), 10)
			continue
		}
		var s string
		if err := json.Unmarshal(rawv, &s); err != nil {
			return nil, fmt.Sprintf("argument %q must be a string", p.Name)
		}
		out[p.Name] = s
	}
	return out, ""
}

package bundle

import (
	"fmt"
	"strings"

	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/sim"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// The invoker-scoped world view (spec 036 US3, contracts/script-api.md): the
// frozen, read-only `world` value a tool.star's apply(args, world) receives. It
// exposes ONLY the members the contract fixes — tick, time_of_day, map
// dimensions, the living-agent roster, a by-name lookup, and the seeded rand
// draw — and nothing else. Private memories, beliefs, relationships, journals,
// pending orders, provider state, wall-clock time, filesystem, network, and
// environment are deliberately absent and stay absent without spec-level review.
//
// Every value the view hands back is frozen (Starlark's publish-to-other-thread
// contract), so a script can read the world but never mutate it, and two
// invocations against the same snapshot see byte-identical inputs — half of the
// determinism guarantee (the other half is world.rand's coordinate seeding).

// agentInfo is the flat, already-resolved agent projection the view exposes: the
// same name/position/liveness surface the effect compiler resolves against, so a
// script sees exactly what the compiler will act on and nothing more.
type agentInfo struct {
	name  string
	x, y  int
	alive bool
}

// worldView is the `world` value. It is immutable Go data built once per
// invocation from the InvocationContext snapshot; the members are computed lazily
// in Attr and every returned Starlark value is frozen before it escapes.
type worldView struct {
	tool      string // the invoking tool's name — namespaces world.rand's purpose
	tick      int64
	timeOfDay string
	mapW      int
	mapH      int
	agents    []agentInfo // living agents only, roster order
	byName    map[string]agentInfo
	seed      uint64
}

// newWorldView builds the view for one invocation. agents carries every agent in
// roster order with its liveness; agents() surfaces the living subset while
// agent(name) can still report a named agent's .alive == False.
func newWorldView(tool string, tick int64, seed uint64, mapW, mapH int, agents []agentInfo) *worldView {
	byName := make(map[string]agentInfo, len(agents))
	living := make([]agentInfo, 0, len(agents))
	for _, a := range agents {
		byName[a.name] = a
		if a.alive {
			living = append(living, a)
		}
	}
	return &worldView{
		tool: tool, tick: tick, timeOfDay: timeOfDay(tick),
		mapW: mapW, mapH: mapH, agents: living, byName: byName, seed: seed,
	}
}

// timeOfDay maps a tick to one of the four contract phases. The night boundary
// (22:00–06:00) is pinned to the sim's own night definition (executor.go
// nightStartSecond=22:00 / dayStartSecond=06:00) so a script's "night" always
// agrees with the world's; the remaining daylight is split into morning
// (06:00–12:00), day (12:00–18:00), and evening (18:00–22:00).
func timeOfDay(tick int64) string {
	sod := clock.SecondOfDay(tick) // seconds since midnight, [0,86400)
	switch {
	case sod >= 22*3600 || sod < 6*3600:
		return "night"
	case sod < 12*3600:
		return "morning"
	case sod < 18*3600:
		return "day"
	default:
		return "evening"
	}
}

// --- starlark.Value ---

func (w *worldView) String() string        { return "world" }
func (w *worldView) Type() string          { return "world" }
func (w *worldView) Freeze()               {} // immutable Go data; nothing to freeze
func (w *worldView) Truth() starlark.Bool  { return starlark.True }
func (w *worldView) Hash() (uint32, error) { return 0, fmt.Errorf("world is unhashable") }

// AttrNames lists the exact, closed member set (contracts/script-api.md).
func (w *worldView) AttrNames() []string {
	return []string{"agent", "agents", "map_height", "map_width", "rand", "tick", "time_of_day"}
}

// Attr resolves a member. Scalars return frozen primitives; the three accessors
// return fresh frozen builtins. An unknown member returns (nil, nil) so Starlark
// raises its standard "has no field" error.
func (w *worldView) Attr(name string) (starlark.Value, error) {
	switch name {
	case "tick":
		return starlark.MakeInt64(w.tick), nil
	case "time_of_day":
		return starlark.String(w.timeOfDay), nil
	case "map_width":
		return starlark.MakeInt(w.mapW), nil
	case "map_height":
		return starlark.MakeInt(w.mapH), nil
	case "agents":
		return starlark.NewBuiltin("agents", w.agentsFn), nil
	case "agent":
		return starlark.NewBuiltin("agent", w.agentFn), nil
	case "rand":
		return starlark.NewBuiltin("rand", w.randFn), nil
	}
	return nil, nil
}

// agentStruct renders one agent as a frozen struct with .name/.x/.y/.alive.
func agentStruct(a agentInfo) starlark.Value {
	s := starlarkstruct.FromStringDict(starlark.String("agent"), starlark.StringDict{
		"name":  starlark.String(a.name),
		"x":     starlark.MakeInt(a.x),
		"y":     starlark.MakeInt(a.y),
		"alive": starlark.Bool(a.alive),
	})
	s.Freeze()
	return s
}

// agentsFn implements world.agents() → frozen list of living-agent structs.
func (w *worldView) agentsFn(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackArgs("agents", args, kwargs); err != nil {
		return nil, err
	}
	elems := make([]starlark.Value, 0, len(w.agents))
	for _, a := range w.agents {
		elems = append(elems, agentStruct(a))
	}
	list := starlark.NewList(elems)
	list.Freeze()
	return list, nil
}

// agentFn implements world.agent(name) → frozen struct or None (exact-name
// lookup, so a script can also observe a named agent's death via .alive).
func (w *worldView) agentFn(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var name string
	if err := starlark.UnpackArgs("agent", args, kwargs, "name", &name); err != nil {
		return nil, err
	}
	if a, ok := w.byName[name]; ok {
		return agentStruct(a), nil
	}
	return starlark.None, nil
}

// randFn implements world.rand(purpose, index) → float in [0,1). It is the ONLY
// randomness a script may draw, routed through the exported seeded accessor with
// the "bundle:<tool>:<purpose>" namespace so two tools (or two purposes) never
// alias the same stream and the draw replays from its coordinates alone.
func (w *worldView) randFn(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var purpose string
	var index int
	if err := starlark.UnpackArgs("rand", args, kwargs, "purpose", &purpose, "index", &index); err != nil {
		return nil, err
	}
	full := "bundle:" + w.tool + ":" + strings.TrimSpace(purpose)
	return starlark.Float(sim.BundleRand(w.seed, full, w.tick, index)), nil
}

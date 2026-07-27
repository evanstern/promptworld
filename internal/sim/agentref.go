package sim

// AgentRef (spec 086) is the wire-side agent reference every agent-bearing
// event payload carries: the roster index plus a denormalized copy of the
// roster name, stamped at emission so the log is self-describing with no
// replica ({"agent":{"id":2,"name":"Cedar"}}). Names are the package
// constant AgentNames — fixed per agent for the life of the world — so
// emission-time naming and any-later-time naming provably agree, dead or
// alive (research R4).
//
// Two laws hold forever (research R2/R3):
//
//   - AgentRef is NEVER reachable from sim.State's type graph. Events are
//     never hashed; STATE bytes are (Marshal/Hash, snapshot verification,
//     world.migrated) — a name field anywhere in State would change those
//     bytes and break pre-086 replay byte-identity. State entities keep bare
//     ints; state-shared payload types split into payload mirrors
//     (TestNoAgentRefInState is the standing tripwire).
//
//   - Name validation NEVER lives in an Apply arm. Reducer arms read .ID
//     only and accept unnamed (legacy) shapes forever — injected pre-086
//     rows replay through the same arms on every recovery. Validation lives
//     exclusively at the live-emission choke points: mustPayload (executor
//     marshal, panic contract) and the InjectSocial door (batch refusal),
//     both via validateRefs. Neither is on the replay path.

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// AgentRef is the {id, name} wire reference. Marshal is the plain struct
// marshal (canonical fixed field order — the payload-struct convention);
// Unmarshal is custom and dual-shape: a bare JSON number decodes as
// {ID: n, Name: ""} (the legacy pre-086 shape, accepted forever), an
// {"id","name"} object decodes field-wise. Name == "" means "legacy row or
// sentinel"; renderers fall back to replica lookup for it.
type AgentRef struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// agentRefAlias avoids UnmarshalJSON recursion on the object branch.
type agentRefAlias AgentRef

// UnmarshalJSON accepts both wire shapes, permanently: a bare index
// (pre-086 rows: 2 → {2, ""}) or the {"id","name"} object. []AgentRef
// therefore decodes both [1,4] and [{"id":1,…},{"id":4,…}] element-wise
// with no extra code.
func (r *AgentRef) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		return fmt.Errorf("agent ref: empty JSON value")
	}
	switch b[0] {
	case 'n': // null — leave the zero value (encoding/json convention)
		return nil
	case '{':
		var a agentRefAlias
		if err := json.Unmarshal(b, &a); err != nil {
			return fmt.Errorf("agent ref: %w", err)
		}
		*r = AgentRef(a)
		return nil
	default: // legacy bare index
		var n int
		if err := json.Unmarshal(b, &n); err != nil {
			return fmt.Errorf("agent ref: %w", err)
		}
		*r = AgentRef{ID: n}
		return nil
	}
}

// Ref is the sanctioned constructor: a pure function of the index over the
// AgentNames roster constant — no state, no liveness check (dead agents
// keep their names). Sentinel/out-of-roster IDs (canonically −1: any/none/
// personal) get empty names; a fake name on a sentinel is as much a bug as
// a missing name on an agent.
func Ref(i int) AgentRef {
	if i >= 0 && i < agentCount {
		return AgentRef{ID: i, Name: AgentNames[i]}
	}
	return AgentRef{ID: i}
}

// Refs maps Ref element-wise; nil in, nil out.
func Refs(ids []int) []AgentRef {
	if ids == nil {
		return nil
	}
	out := make([]AgentRef, len(ids))
	for i, id := range ids {
		out[i] = Ref(id)
	}
	return out
}

// agentRefType is the reflection sentinel validateRefs walks for.
var agentRefType = reflect.TypeOf(AgentRef{})

// validateRefs is the shared live-emission rail (spec 086 FR-005): a
// reflection walk over a payload value that checks every reachable AgentRef
// — in-roster IDs must carry the exact roster name, out-of-roster
// (sentinel) IDs must carry the empty name. Used by mustPayload (panic
// contract) and the InjectSocial door (batch refusal); NEVER called from an
// Apply arm (replay accepts unnamed shapes forever — research R3).
func validateRefs(v any) error {
	if v == nil {
		return nil
	}
	return walkRefs(reflect.ValueOf(v))
}

func walkRefs(v reflect.Value) error {
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		if v.IsNil() {
			return nil
		}
		return walkRefs(v.Elem())
	case reflect.Struct:
		if v.Type() == agentRefType {
			return checkRef(int(v.Field(0).Int()), v.Field(1).String())
		}
		for i := 0; i < v.NumField(); i++ {
			if err := walkRefs(v.Field(i)); err != nil {
				return err
			}
		}
		return nil
	case reflect.Slice, reflect.Array:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return nil // opaque bytes (json.RawMessage) — never walked
		}
		for i := 0; i < v.Len(); i++ {
			if err := walkRefs(v.Index(i)); err != nil {
				return err
			}
		}
		return nil
	case reflect.Map:
		iter := v.MapRange()
		for iter.Next() {
			if err := walkRefs(iter.Key()); err != nil {
				return err
			}
			if err := walkRefs(iter.Value()); err != nil {
				return err
			}
		}
		return nil
	default:
		return nil
	}
}

func checkRef(id int, name string) error {
	if id >= 0 && id < agentCount {
		if name != AgentNames[id] {
			return fmt.Errorf("agent ref {%d,%q}: in-roster ref must carry the roster name %q (build refs with sim.Ref/sim.Refs)", id, name, AgentNames[id])
		}
		return nil
	}
	if name != "" {
		return fmt.Errorf("agent ref {%d,%q}: out-of-roster (sentinel) ref must carry an empty name", id, name)
	}
	return nil
}

package guardian

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/toolloop"
)

// --- spec 102 T004: the tutor/world channel split is STRUCTURAL (D4, FR-004) ---

// TestTutorSurfaceStructurallyInert walks tutorSurface's transitive field
// graph by reflection and refuses any type through which behavior could
// travel: interfaces (an Injector/LoopControl could hide there), functions,
// channels, and unsafe pointers. What remains is inert descriptor data —
// strings, numbers, slices/structs of the same — so the tutor channel's ONLY
// possible output is the string it returns. A future field that could reach
// a world door fails THIS test, not a review.
func TestTutorSurfaceStructurallyInert(t *testing.T) {
	seen := map[reflect.Type]bool{}
	var walk func(rt reflect.Type, path string)
	walk = func(rt reflect.Type, path string) {
		if seen[rt] {
			return
		}
		seen[rt] = true
		switch rt.Kind() {
		case reflect.Interface, reflect.Func, reflect.Chan, reflect.UnsafePointer:
			t.Errorf("tutorSurface field graph carries a %s at %s — a behavior-capable type breaches the tutor channel's structural isolation", rt.Kind(), path)
		case reflect.Ptr, reflect.Slice, reflect.Array, reflect.Map:
			walk(rt.Elem(), path+"/*")
			if rt.Kind() == reflect.Map {
				walk(rt.Key(), path+"/key")
			}
		case reflect.Struct:
			for i := 0; i < rt.NumField(); i++ {
				f := rt.Field(i)
				walk(f.Type, path+"."+f.Name)
			}
		}
	}
	walk(reflect.TypeOf(tutorSurface{}), "tutorSurface")
}

// TestTutorChannelReachesNoDoor is the D4 behavioral pin: a turn whose model
// output is PURE tutor channel — explain calls plus a converse reply — lands
// zero world events, spends zero charges, moves zero faith, and contributes
// zero rubric facts. Proven at the strongest available boundary: the entire
// world state is byte-identical after the turn (the only injections are
// cog.* telemetry, reducer no-ops by whitelist doctrine — so every rubric
// arm, which folds world state, is untouched by construction).
func TestTutorChannelReachesNoDoor(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "Here is what charges mean.")
	mt.replica.GuardianCharges = 3
	inj.state.GuardianCharges = 3
	mt.mirrorState()

	before := string(inj.state.Marshal())
	faithBefore := inj.state.FaithScore()

	// A tutor-shaped loop: one explain read, then converse (Final text).
	mt.runLoop = func(ctx context.Context, j toolloop.Job) (toolloop.Result, error) {
		c := toolCall("explain", `{"topic":"charges"}`)
		out := j.Handlers["explain"](ctx, c)
		if out.Verdict != toolloop.VerdictReadOK {
			t.Fatalf("explain verdict = %s", out.Verdict)
		}
		j.Record(toolloop.CallRecord{JobID: j.JobID, Ordinal: 1, Tool: "explain",
			Args: c.Args, Verdict: out.Verdict, Reason: "", Tier: "cloud"})
		return toolloop.Result{Final: "Charges are the strength you bank; ask and I will explain more.", Term: toolloop.TermModelDone}, nil
	}
	res, err := mt.Turn(t.Context(), "what are charges?")
	if err != nil {
		t.Fatal(err)
	}
	if res.Reply == "" || res.Nudge != nil || res.Miracle != nil || res.Order != nil || res.Plan != nil || res.Clock != "" {
		t.Fatalf("tutor turn produced a world act: %+v", res)
	}

	// Every injected batch is cog.* telemetry only — no world event type.
	for _, b := range inj.batches {
		for _, e := range b {
			if !strings.HasPrefix(e.Type, "cog.") {
				t.Fatalf("tutor turn landed a non-telemetry event: %s", e.Type)
			}
		}
	}
	// And the world is BYTE-identical: no charge, no faith, no rubric-visible
	// state anywhere (cog.* arms are reducer no-ops).
	if after := string(inj.state.Marshal()); after != before {
		t.Fatal("tutor-channel turn mutated world state")
	}
	if inj.state.FaithScore() != faithBefore {
		t.Fatal("tutor-channel turn moved faith")
	}
	if inj.state.GuardianCharges != 3 {
		t.Fatalf("tutor-channel turn spent charges: %d", inj.state.GuardianCharges)
	}
}

// TestConverseIsNotARosterTool pins the other half of the split's
// construction: "converse" is the final-text channel, never a declared tool
// — no roster entry, no handler, so the model cannot even address it as one.
func TestConverseIsNotARosterTool(t *testing.T) {
	mt, _, _, _ := newTestGuardian(t, "words")
	d := &turnDispatch{mt: mt, grant: fullGrant()}
	if _, ok := mt.turnHandlers(d)["converse"]; ok {
		t.Fatal("converse has a handler — it must remain the final-text channel")
	}
	for _, tl := range grantedRoster(fullGrant()) {
		if tl.Name == "converse" {
			t.Fatal("converse declared on the loop roster")
		}
	}
}

package guardian

// The tutor channel's STRUCTURAL isolation (spec 102 D4, FR-004): the tutor
// voice — converse plus the read-only explain tool — must be PHYSICALLY
// unable to reach a world door, spend a charge, move faith, or contribute to
// a rubric. This file makes that a type-level fact rather than a discipline:
//
//   - tutorSurface is the tutor path's ENTIRE capability. It holds only
//     inert descriptor data (tool.ExplainScope: tool descriptors + a stage
//     string) and can only produce a string. It carries no Injector, no
//     LoopControl, no *Guardian, no function values, no channels — nothing
//     through which an event, charge, faith movement, or rubric fact could
//     travel. TestTutorSurfaceStructurallyInert enforces this by reflection
//     over the transitive field graph, so a future field that could reach
//     the world fails the build's tests, not a review.
//   - converse is not a handler at all (tool-loop doctrine since spec 017):
//     the model's final text lands in TurnResult.Reply / the transcript —
//     console-facing strings, never an event batch.
//   - The rubric side holds by the same construction: scenario rubric arms
//     fold recorded WORLD events (internal/sim scenario machinery), and a
//     tutor-channel exchange lands none — at most cog.* telemetry, which is
//     reducer-no-op by whitelist doctrine. TestTutorChannelReachesNoDoor
//     pins the whole chain: a converse+explain turn leaves the world state
//     byte-identical.
//
// The guardian's acting tools keep their own path (turnDispatch → landers →
// InjectSocial) untouched: the split is between CHANNELS, and the tutor
// channel simply has no wire to the world side.

import (
	"github.com/evanstern/promptworld/internal/tool"
)

// tutorSurface is the tutor channel's whole capability: compose grounded
// text from inert descriptor data. Constructed per turn WITHOUT the social
// injector or loop control — by construction it cannot land events, spend
// charges, earn faith, or feed rubrics.
type tutorSurface struct {
	scope tool.ExplainScope
}

// sheet composes the read-only fact sheet for a topic (spec 063 US1) — the
// tutor channel's one capability beyond the reply text itself.
func (t tutorSurface) sheet(topic string) string {
	return tool.ExplainSheet(topic, t.scope)
}

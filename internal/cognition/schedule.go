package cognition

// Shared cadence-schedule arithmetic (spec 102 SC-004): the phase-preserving
// due advance both scheduled cognition lanes use — the villagers' planner
// cadence (internal/mind, where this function originated as TASK-44's
// nextPhasePreservingDue) and the guardian's angel cadence (spec 102 D2).
// One implementation, two drivers, so the schedule arithmetic can never fork.

// NextPhasePreservingDue advances an overdue schedule to the next tick
// strictly after tick, stepping in whole cadence multiples from its own due
// — never from tick. This is the TASK-44 fix: re-arming "from now" instead
// of from the agent's own due collapses every agent a shared stall left
// overdue onto the identical due, locking the whole village into lockstep
// the next time cadence comes around. Preserving due's phase (due mod
// cadence) keeps each agent's boot offset intact forever, regardless of how
// many cadences it had to skip. Arithmetic equivalent of:
//
//	for due <= tick { due += cadence }
func NextPhasePreservingDue(due, tick, cadence int64) int64 {
	if cadence <= 0 || due > tick {
		return due
	}
	return due + (tick-due)/cadence*cadence + cadence
}

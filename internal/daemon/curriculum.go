package daemon

// The curriculum-ladder unlock observer (spec 046 US3, T013): the daemon-side
// half of the fixture-proven pass -> unlock -> record chain. Wired onto the
// notify fan-out alongside the scribe (always-on, before the LLM gate — a
// no-model world still records its unlocks): on observing
// curriculum.stage_unlocked it upserts the per-user unlocks record with a
// pointer back into THIS world and the exercise_passed event that satisfied
// the gate, so the claim stays independently auditable (FR-008). Production
// emission of these events is TASK-119's rubric machinery; until it lands
// this observer simply sits idle for every world (research R5) — proven now
// by fixture-driven tests emitting the events directly.

import (
	"encoding/json"
	"time"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/world"
	"github.com/evanstern/promptworld/internal/worlds"
)

// curriculumObserver returns a notify-fan-out consumer that upserts the
// per-user unlocks record on curriculum.stage_unlocked. Non-blocking and
// non-fatal by contract (worlds.UpsertUnlock warns-and-continues on any
// failure) — an advisory record must never perturb the sim loop.
func curriculumObserver(w *world.World) func([]store.Event) {
	return func(evs []store.Event) {
		for _, e := range evs {
			if e.Type != "curriculum.stage_unlocked" {
				continue
			}
			var up sim.StageUnlockedPayload
			if err := json.Unmarshal(e.Payload, &up); err != nil {
				continue
			}
			entry := worlds.UnlockEntry{
				World:    w.Manifest.Name,
				Path:     w.Dir,
				Exercise: up.Exercise,
				EarnedAt: time.Now().UTC().Format(time.RFC3339),
			}
			// Locate the exercise_passed event that proved this unlock (the
			// contract's evidence shape: a pointer to the pass event itself,
			// not each rubric-satisfying event beneath it) — the executor
			// emission class lands both events in the same batch.
			for _, pe := range evs {
				if pe.Type != "curriculum.exercise_passed" {
					continue
				}
				var pp sim.ExercisePassedPayload
				if err := json.Unmarshal(pe.Payload, &pp); err == nil && pp.Exercise == up.Exercise {
					entry.Evidence = []worlds.UnlockEvidenceRef{{Type: pe.Type, Seq: pe.Seq, Tick: pe.Tick}}
					break
				}
			}
			worlds.UpsertUnlock(up.Stage, entry)
		}
	}
}

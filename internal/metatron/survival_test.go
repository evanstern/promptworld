package metatron

// Tests for the metatron survival autonomy feature (spec 059): system-origin
// survival watches (US1), the survival-authority turn frame + attribution (US2),
// and the targeting digest (US3).

import (
	"context"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/tool"
)

// seededSurvivalWatch installs the near-death survival watch into both the
// injector state (the door) and the angel's replica + mirror, the seedOrder way.
func seededSurvivalWatch(mt *Metatron, inj *stateInjector, kind string) sim.MetatronOrder {
	var o sim.MetatronOrder
	for _, w := range sim.SurvivalWatchDefs(0) {
		if w.Survival == kind {
			o = w
		}
	}
	seedOrder(mt, inj, o)
	return o
}

// needsEvent builds an agent.needs_changed event for one villager at a tick.
func needsEvent(agent, health, food, warmth int, tick int64) store.Event {
	e := mustEvent("agent.needs_changed", sim.NeedsPayload{
		Agent: agent, Health: health, Food: food, Warmth: warmth, Rest: 500, Morale: 500,
	})
	e.Tick = tick
	return e
}

// TestSurvivalWatchRefusesPlayerCancel (spec 059 US1 AC-4): a player cancel naming
// a system survival watch is refused with in-fiction counsel — the reducer rejects
// it at the door, and cancelOrder maps that to the angel's own voice.
func TestSurvivalWatchRefusesPlayerCancel(t *testing.T) {
	mt, _, inj, _ := newTestAngel(t, "ok")
	o := seededSurvivalWatch(mt, inj, sim.SurvivalNearDeath)

	why := mt.cancelOrder(o.ID, fullGrant())
	if why == "" {
		t.Fatal("a survival watch was cancelled by the player order surface")
	}
	if !strings.Contains(why, "my own nature") {
		t.Errorf("cancel refusal not in-fiction: %q", why)
	}
	// Nothing landed / the watch still stands active.
	if inj.state.MetatronOrders[0].Status != "active" {
		t.Errorf("a refused cancel still changed the survival watch: %+v", inj.state.MetatronOrders[0])
	}
}

// TestSurvivalWatchMatchesNeedsBand (spec 059 US1/US2): the survival-band matcher
// enqueues a trigger job when a villager's agent.needs_changed crosses into the
// danger band, and does NOT re-fire while the villager stays in-band (the latch),
// but re-arms and fires again once the villager recovers and relapses.
func TestSurvivalWatchMatchesNeedsBand(t *testing.T) {
	mt, _, inj, _ := newTestAngel(t, "ok")
	seededSurvivalWatch(mt, inj, sim.SurvivalStarvation)

	fire := func(agent, food int, tick int64) bool {
		mt.matchOrders([]store.Event{needsEvent(agent, 800, food, 800, tick)})
		return dequeued(mt.triggerQ)
	}
	clearPending := func() {
		mt.stateMu.Lock()
		delete(mt.pendingTrigger, "sys-watch-"+sim.SurvivalStarvation)
		mt.stateMu.Unlock()
	}

	// Out of the band (fed): no fire.
	if fire(2, 500, 1000) {
		t.Fatal("a fed villager fired the starvation watch")
	}
	// Into the band (Food==0): fires once.
	if !fire(2, 0, 2000) {
		t.Fatal("a starving villager did not fire the starvation watch")
	}
	clearPending()
	// Still starving: the latch suppresses a re-fire.
	if fire(2, 0, 3000) {
		t.Fatal("the starvation watch re-fired while the villager stayed in-band (latch failed)")
	}
	clearPending()
	// Recovered above the re-arm band: no fire, latch clears.
	if fire(2, sim.SurvivalStarvingRearm, 4000) {
		t.Fatal("a recovered villager fired the watch")
	}
	clearPending()
	// Relapse: fires again (re-armed).
	if !fire(2, 0, 5000) {
		t.Fatal("the starvation watch did not re-fire after recovery + relapse")
	}
	_ = inj
}

// TestSurvivalWatchIsMiracleCapableRoster is a guard: work_miracle is on the full
// loop roster the survival turn runs under, so the digest gate (hasWorkMiracle)
// and the survival authority both apply.
func TestSurvivalWatchIsMiracleCapableRoster(t *testing.T) {
	if !hasWorkMiracle(tool.LoopRosterMetatron()) {
		t.Fatal("work_miracle absent from the metatron loop roster")
	}
	_ = context.Background
}

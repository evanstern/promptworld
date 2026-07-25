package sim

import (
	"testing"

	"github.com/evanstern/promptworld/internal/store"
)

// Spec 043 US2 (T015): the needs trajectory anchor is maintained entirely by the
// agent.needs_changed reducer arm, so it is replay-safe by construction. These
// tests pin the window-edge refresh, the unset-anchor first window, and the
// sleep-spanning window reflecting an overnight fall (spec edge case: a sleeper's
// trajectory must show what the sleeper experienced, not read as noise).

// needsEvent applies one agent.needs_changed for agent 0 at tick with the given
// levels, driving the real reducer arm.
func needsEvent(t *testing.T, s *State, tick int64, health, food, rest, warmth, morale int) {
	t.Helper()
	applyTo(t, s, store.Event{Tick: tick, Type: "agent.needs_changed",
		Payload: mustPayload(NeedsPayload{Agent: 0,
			Health: health, Food: food, Rest: rest, Warmth: warmth, Morale: morale})})
}

// TestNeedsAnchorUnsetFirstWindow: on a fresh world (tick 0) the anchor stays
// unset (nil, tick 0) for the whole first trajectoryWindowTicks — needs changes
// inside that window do NOT establish an anchor, so the first thought has no
// trajectory and renders steady (edge case 1). The window's worth of game time
// is what unlocks the first anchor.
func TestNeedsAnchorUnsetFirstWindow(t *testing.T) {
	s := NewState(42, testMap(42))

	// A change well inside the first window leaves the anchor unset.
	needsEvent(t, s, 600, 900, 800, 700, 600, 500)
	a := &s.Agents[0]
	if a.NeedsAnchor != nil || a.NeedsAnchorTick != 0 {
		t.Fatalf("anchor set inside first window: anchor=%+v tick=%d", a.NeedsAnchor, a.NeedsAnchorTick)
	}

	// The first change at or past the window edge establishes the anchor at the
	// then-current levels.
	needsEvent(t, s, trajectoryWindowTicks, 880, 780, 690, 590, 490)
	if a.NeedsAnchor == nil {
		t.Fatal("anchor still unset at the window edge")
	}
	if a.NeedsAnchorTick != trajectoryWindowTicks {
		t.Errorf("anchor tick = %d, want %d", a.NeedsAnchorTick, trajectoryWindowTicks)
	}
	if a.NeedsAnchor.Warmth != 590 {
		t.Errorf("anchor captured levels %+v, want the window-edge current (warmth 590)", a.NeedsAnchor)
	}
}

// TestNeedsAnchorWindowEdgeRefresh: once an anchor exists it holds across every
// change WITHIN the window (so direction measures movement over the window), and
// rolls forward to the current levels exactly when the next full window has
// elapsed since the last edge.
func TestNeedsAnchorWindowEdgeRefresh(t *testing.T) {
	s := NewState(42, testMap(42))
	a := &s.Agents[0]

	// Establish the first anchor at the window edge.
	needsEvent(t, s, trajectoryWindowTicks, 900, 900, 900, 900, 900)
	firstTick := a.NeedsAnchorTick
	if firstTick != trajectoryWindowTicks {
		t.Fatalf("first anchor tick = %d, want %d", firstTick, trajectoryWindowTicks)
	}

	// A change partway through the next window does NOT refresh — the anchor
	// stays pinned to the window edge so the trajectory can be measured against it.
	needsEvent(t, s, trajectoryWindowTicks+600, 900, 900, 900, 700, 900)
	if a.NeedsAnchorTick != firstTick {
		t.Errorf("anchor refreshed mid-window: tick = %d, want %d", a.NeedsAnchorTick, firstTick)
	}
	if a.NeedsAnchor.Warmth != 900 {
		t.Errorf("mid-window anchor levels moved: %+v, want the edge snapshot (warmth 900)", a.NeedsAnchor)
	}

	// A change at the next window edge rolls the anchor forward to the current.
	edge2 := int64(trajectoryWindowTicks) * 2
	needsEvent(t, s, edge2, 900, 900, 900, 650, 900)
	if a.NeedsAnchorTick != edge2 {
		t.Errorf("anchor did not roll at the second edge: tick = %d, want %d", a.NeedsAnchorTick, edge2)
	}
	if a.NeedsAnchor.Warmth != 650 {
		t.Errorf("rolled anchor levels = %+v, want the new edge current (warmth 650)", a.NeedsAnchor)
	}
}

// TestNeedsAnchorSleepSpanningWindow (spec edge case: a villager that woke from
// sleep): the trajectory window spans the sleep period, and the anchor rolls
// forward across it, so on waking current−anchor still reflects the real
// overnight fall (warmth dropping minute after minute) rather than resetting to
// steady. Drive warmth steadily down over several windows and assert the anchor
// stays strictly above the current — a falling reading — at every point past the
// first window.
func TestNeedsAnchorSleepSpanningWindow(t *testing.T) {
	s := NewState(42, testMap(42))
	a := &s.Agents[0]

	// Warmth falls one point (10 of 1000) each game-minute across a long sleep
	// spanning three trajectory windows; every other need holds. needs_changed
	// fires each game-minute (executor). Warmth stays strictly positive over the
	// drive so the anchor/current comparison is never a floored 0-vs-0 tie.
	warmth := 1000
	for tick := int64(60); tick <= int64(trajectoryWindowTicks)*3; tick += 60 {
		if warmth > 0 {
			warmth -= 10
		}
		needsEvent(t, s, tick, 1000, 1000, 1000, warmth, 1000)

		// Past the first window the anchor is always established, and — except at
		// the exact instant it rolls forward (when anchor == current for one
		// tick) — it is strictly higher than the current warmth: a genuine
		// "falling" signal that survived the sleep-spanning windows rather than
		// resetting to steady.
		if tick >= int64(trajectoryWindowTicks)*2 {
			if a.NeedsAnchor == nil {
				t.Fatalf("anchor unset at tick %d during a long sleep", tick)
			}
			if a.NeedsAnchorTick != tick && a.NeedsAnchor.Warmth <= a.Needs.Warmth {
				t.Fatalf("tick %d: anchor warmth %d not above current %d — the overnight fall was lost",
					tick, a.NeedsAnchor.Warmth, a.Needs.Warmth)
			}
		}
	}
	// And warmth really did fall a lot (sanity: the drive worked).
	if a.Needs.Warmth > 300 {
		t.Fatalf("drive did not fall warmth as expected: current %d", a.Needs.Warmth)
	}
}

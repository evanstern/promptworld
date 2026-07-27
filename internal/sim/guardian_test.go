package sim

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

func nudgeEvent(t *testing.T, tick int64, p GuardianNudgedPayload) store.Event {
	t.Helper()
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	return store.Event{Tick: tick, Type: "metatron.nudged", Payload: b}
}

var regenEvent = store.Event{Type: "metatron.charge_regenerated", Payload: []byte("{}")}

// TestChargeInvariants: genesis 1; regen caps at 3; spends floor via
// validation (a spend at 0 is a reducer error, not a clamp).
func TestChargeInvariants(t *testing.T) {
	m := worldmap.Generate(7, 32, 32)
	s := NewState(7, m)
	if s.GuardianCharges != GuardianGenesisCharges {
		t.Fatalf("genesis charges = %d, want %d", s.GuardianCharges, GuardianGenesisCharges)
	}
	for i := 0; i < 10; i++ {
		if err := s.Apply(regenEvent); err != nil {
			t.Fatal(err)
		}
		if s.GuardianCharges > GuardianChargeCap {
			t.Fatalf("charges %d exceeded cap after regen storm", s.GuardianCharges)
		}
	}
	if s.GuardianCharges != GuardianChargeCap {
		t.Fatalf("charges = %d, want cap %d", s.GuardianCharges, GuardianChargeCap)
	}
	for i := 0; i < GuardianChargeCap; i++ {
		if err := s.Apply(nudgeEvent(t, 100, GuardianNudgedPayload{
			Form: "dream", Targets: Refs([]int{0}), Text: "a quiet dream"})); err != nil {
			t.Fatal(err)
		}
	}
	if s.GuardianCharges != 0 {
		t.Fatalf("charges = %d after spending the bank, want 0", s.GuardianCharges)
	}
	if err := s.Apply(nudgeEvent(t, 101, GuardianNudgedPayload{
		Form: "dream", Targets: Refs([]int{0}), Text: "one too many"})); err == nil {
		t.Fatal("spend at 0 charges must be a reducer error (the dry-run gate)")
	}
}

// TestNudgeValidation: the reducer rejects malformed nudges so the
// InjectSocial dry-run refuses them at the door.
func TestNudgeValidation(t *testing.T) {
	m := worldmap.Generate(7, 32, 32)
	long := make([]byte, NudgeTextMax+1)
	for i := range long {
		long[i] = 'x'
	}
	cases := []struct {
		name string
		p    GuardianNudgedPayload
		dead int // -1 = none
	}{
		{"unknown form", GuardianNudgedPayload{Form: "whisper", Targets: Refs([]int{0}), Text: "t"}, -1},
		{"vision multi-target", GuardianNudgedPayload{Form: "vision", Targets: Refs([]int{0, 1}), Text: "t"}, -1},
		{"omen no targets", GuardianNudgedPayload{Form: "omen", Targets: nil, Text: "t"}, -1},
		{"unknown target", GuardianNudgedPayload{Form: "vision", Targets: Refs([]int{99}), Text: "t"}, -1},
		{"dead target", GuardianNudgedPayload{Form: "vision", Targets: Refs([]int{2}), Text: "t"}, 2},
		{"empty text", GuardianNudgedPayload{Form: "vision", Targets: Refs([]int{0}), Text: ""}, -1},
		{"over-cap text", GuardianNudgedPayload{Form: "vision", Targets: Refs([]int{0}), Text: string(long)}, -1},
		// dream (grandfathered) still requires exactly one target — a legacy
		// multi-target payload is rejected just as it always was.
		{"dream multi-target", GuardianNudgedPayload{Form: "dream", Targets: Refs([]int{0, 1}), Text: "t"}, -1},
	}
	for _, c := range cases {
		s := NewState(7, m)
		if c.dead >= 0 {
			s.Agents[c.dead].Dead = true
		}
		if err := s.Apply(nudgeEvent(t, 50, c.p)); err == nil {
			t.Errorf("%s: expected reducer rejection", c.name)
		}
		if s.GuardianCharges != GuardianGenesisCharges {
			t.Errorf("%s: rejected nudge changed charges", c.name)
		}
	}
	// A daytime omen is rejected: the omen form lands only at night (spec 029).
	day := NewState(7, m)
	if err := day.Apply(nudgeEvent(t, 50, GuardianNudgedPayload{
		Form: "omen", Targets: Refs([]int{0, 1}), Text: "not yet"})); err == nil {
		t.Error("daytime omen must be rejected (omen is night-only)")
	}

	// The valid shapes pass: a vision any time; an omen at night.
	s := NewState(7, m)
	if err := s.Apply(nudgeEvent(t, 50, GuardianNudgedPayload{
		Form: "vision", Targets: Refs([]int{0}), Text: "a waking light"})); err != nil {
		t.Fatalf("valid vision rejected: %v", err)
	}
	night := NewState(7, m)
	night.Night = true
	if err := night.Apply(nudgeEvent(t, 50, GuardianNudgedPayload{
		Form: "omen", Targets: Refs([]int{0, 1, 2, 3, 4, 5, 6, 7}), Text: "the sky split"})); err != nil {
		t.Fatalf("valid night omen rejected: %v", err)
	}
}

// TestRegenBoundaries: the executor emits regeneration exactly at absolute
// 6-game-hour tick boundaries, only below cap.
func TestRegenBoundaries(t *testing.T) {
	m := worldmap.Generate(7, 32, 32)
	s := NewState(7, m)
	s.GuardianCharges = 1

	count := func(tick int64) int {
		n := 0
		for _, e := range stepEvents(s, m, tick) {
			if e.Type == "metatron.charge_regenerated" {
				n++
			}
		}
		return n
	}
	if got := count(chargeRegenTicks); got != 1 {
		t.Errorf("boundary tick emitted %d regens, want 1", got)
	}
	if got := count(chargeRegenTicks + 1); got != 0 {
		t.Errorf("off-boundary tick emitted %d regens, want 0", got)
	}
	s.GuardianCharges = GuardianChargeCap
	if got := count(2 * chargeRegenTicks); got != 0 {
		t.Errorf("at-cap boundary emitted %d regens, want 0", got)
	}
}

// TestGuardianChargeRegenTicksMatchesExecutor (spec 050 T002): the exported
// GuardianChargeRegenTicks must be exactly the modulus the executor fires
// metatron.charge_regenerated on (executor.go's `nextTick%chargeRegenTicks
// == 0` rule) — internal/tui derives the guardian strip's next-regen
// forecast from the exported constant alone, so it must never silently
// drift from the private value the executor actually uses.
func TestGuardianChargeRegenTicksMatchesExecutor(t *testing.T) {
	if GuardianChargeRegenTicks != chargeRegenTicks {
		t.Fatalf("GuardianChargeRegenTicks = %d, want %d (chargeRegenTicks)", GuardianChargeRegenTicks, chargeRegenTicks)
	}
	m := worldmap.Generate(11, 32, 32)
	s := NewState(11, m)
	s.GuardianCharges = 1

	fires := func(tick int64) bool {
		for _, e := range stepEvents(s, m, tick) {
			if e.Type == "metatron.charge_regenerated" {
				return true
			}
		}
		return false
	}
	if !fires(GuardianChargeRegenTicks) {
		t.Error("exported cadence boundary tick did not fire charge_regenerated")
	}
	if fires(GuardianChargeRegenTicks + 1) {
		t.Error("off-boundary tick (exported cadence + 1) fired charge_regenerated")
	}
}

// TestOldSnapshotGainsGenesisCharge: pre-TASK-12 snapshots have no charges
// field; unmarshal into genesis state keeps the default (documented upgrade).
// A modern snapshot with a spent-to-zero bank must round-trip as 0.
func TestOldSnapshotGainsGenesisCharge(t *testing.T) {
	m := worldmap.Generate(7, 32, 32)
	// Pre-TASK-12 shape: strip the field entirely.
	var asMap map[string]json.RawMessage
	if err := json.Unmarshal(NewState(7, m).Marshal(), &asMap); err != nil {
		t.Fatal(err)
	}
	delete(asMap, "metatron_charges")
	oldBytes, _ := json.Marshal(asMap)
	restored := NewState(7, m)
	if err := json.Unmarshal(oldBytes, restored); err != nil {
		t.Fatal(err)
	}
	if restored.GuardianCharges != GuardianGenesisCharges {
		t.Fatalf("restored charges = %d, want genesis %d", restored.GuardianCharges, GuardianGenesisCharges)
	}
	// Modern shape: zero survives the round trip.
	spent := NewState(7, m)
	spent.GuardianCharges = 0
	rt := NewState(7, m)
	if err := json.Unmarshal(spent.Marshal(), rt); err != nil {
		t.Fatal(err)
	}
	if rt.GuardianCharges != 0 {
		t.Fatalf("spent bank resurrected as %d, want 0", rt.GuardianCharges)
	}
}

// TestChargesReplayIdentically: a state rebuilt by replaying recorded
// regen/spend events matches the live sequence exactly.
func TestChargesReplayIdentically(t *testing.T) {
	m := worldmap.Generate(7, 32, 32)
	live := NewState(7, m)
	events := []store.Event{
		regenEvent, regenEvent, // 1 -> 3
		// A grandfathered dream (legacy replay) and a live-form vision, so the
		// replay test spans both an old and a new nudge form; neither depends on
		// State.Night (only omen does), keeping this a pure charge-economy check.
		nudgeEvent(t, 10, GuardianNudgedPayload{Form: "dream", Targets: Refs([]int{3}), Text: "d1"}),  // 2
		nudgeEvent(t, 20, GuardianNudgedPayload{Form: "vision", Targets: Refs([]int{0}), Text: "v1"}), // 1
		regenEvent, // 2
	}
	for _, e := range events {
		if err := live.Apply(e); err != nil {
			t.Fatal(err)
		}
	}
	replayed := NewState(7, m)
	for _, e := range events {
		if err := replayed.Apply(e); err != nil {
			t.Fatal(err)
		}
	}
	if live.GuardianCharges != 2 || replayed.GuardianCharges != live.GuardianCharges {
		t.Fatalf("live %d vs replayed %d (want 2)", live.GuardianCharges, replayed.GuardianCharges)
	}
	if live.Hash() != replayed.Hash() {
		t.Fatal("state hashes diverge under replay")
	}
}

// --- Guardian standing orders (spec 029, T004) ---

func orderEvent(t *testing.T, typ string, tick int64, payload any) store.Event {
	t.Helper()
	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return store.Event{Tick: tick, Type: typ, Payload: b}
}

// validOrder is a placeable player order fixture: a compilable condition, one
// event type, any-agent, a 3-game-day TTL from placedTick.
func validOrder(id, origin string, placedTick int64) GuardianOrder {
	return GuardianOrder{
		ID: id, Origin: origin, Condition: "when Rowan next falls asleep",
		Action: "send her a comforting vision", EventTypes: []string{"agent.slept"},
		Agent: -1, PlacedTick: placedTick, ExpiresTick: placedTick + 3*ticksPerGameDay,
	}
}

// TestGuardianOrderPlacedRejections (spec 029, data-model §1 / contracts/events.md):
// every order_placed rejection row is refused at the reducer dry-run, leaving the
// order set unchanged.
func TestGuardianOrderPlacedRejections(t *testing.T) {
	m := worldmap.Generate(7, 32, 32)
	tooShort := validOrder("ord-x", "player", 0)
	tooShort.ExpiresTick = GuardianOrderTTLMinDays*ticksPerGameDay - 1 // < 1 day
	tooLong := validOrder("ord-x", "player", 0)
	tooLong.ExpiresTick = GuardianOrderTTLMaxDays*ticksPerGameDay + 1 // > 7 days
	badOrigin := validOrder("ord-x", "guardian", 0)
	noEvents := validOrder("ord-x", "player", 0)
	noEvents.EventTypes = nil
	badAgentHigh := validOrder("ord-x", "player", 0)
	badAgentHigh.Agent = 999
	badAgentLow := validOrder("ord-x", "player", 0)
	badAgentLow.Agent = -2
	longCond := validOrder("ord-x", "player", 0)
	longCond.Condition = stringOf('c', 301)
	longAction := validOrder("ord-x", "player", 0)
	longAction.Action = stringOf('a', 401)
	emptyID := validOrder("", "player", 0)

	cases := []struct {
		name string
		o    GuardianOrder
	}{
		{"empty id", emptyID},
		{"unknown origin", badOrigin},
		{"no event_types", noEvents},
		{"ttl too short", tooShort},
		{"ttl too long", tooLong},
		{"agent index too high", badAgentHigh},
		{"agent index below -1", badAgentLow},
		{"condition over 300", longCond},
		{"action over 400", longAction},
	}
	for _, c := range cases {
		s := NewState(7, m)
		if err := s.Apply(orderEvent(t, "metatron.order_placed", 0, c.o)); err == nil {
			t.Errorf("%s: expected reducer rejection", c.name)
		}
		if len(s.GuardianOrders) != 0 {
			t.Errorf("%s: rejected placement left %d orders", c.name, len(s.GuardianOrders))
		}
	}

	// Duplicate id in any status is rejected.
	s := NewState(7, m)
	if err := s.Apply(orderEvent(t, "metatron.order_placed", 0, validOrder("ord-dup", "player", 0))); err != nil {
		t.Fatalf("first placement rejected: %v", err)
	}
	if err := s.Apply(orderEvent(t, "metatron.order_placed", 1, validOrder("ord-dup", "player", 1))); err == nil {
		t.Error("duplicate order id must be rejected")
	}
}

// stringOf builds a string of n copies of c (test helper for cap rows).
func stringOf(c byte, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = c
	}
	return string(b)
}

// TestGuardianPlayerOrderCap (spec 029 FR-007/FR-012): at most 3 ACTIVE
// player-origin orders; a 4th is refused, but a system-origin deferral order is
// exempt from the cap. Cancelling a player order frees a slot.
func TestGuardianPlayerOrderCap(t *testing.T) {
	m := worldmap.Generate(7, 32, 32)
	s := NewState(7, m)
	for i := 0; i < GuardianPlayerOrderCap; i++ {
		id := "ord-p" + string(rune('0'+i))
		if err := s.Apply(orderEvent(t, "metatron.order_placed", int64(i), validOrder(id, "player", int64(i)))); err != nil {
			t.Fatalf("player order %d rejected: %v", i, err)
		}
	}
	// The 4th player order exceeds the cap.
	if err := s.Apply(orderEvent(t, "metatron.order_placed", 10, validOrder("ord-p3", "player", 10))); err == nil {
		t.Error("4th active player order must be refused (cap 3)")
	}
	// A system-origin order is exempt.
	if err := s.Apply(orderEvent(t, "metatron.order_placed", 11, validOrder("ord-sys", "system", 11))); err != nil {
		t.Errorf("system-origin order must be exempt from the player cap: %v", err)
	}
	// Cancelling a player order frees a slot for a new player order.
	if err := s.Apply(orderEvent(t, "metatron.order_cancelled", 12, OrderIDPayload{ID: "ord-p0"})); err != nil {
		t.Fatalf("cancel rejected: %v", err)
	}
	if err := s.Apply(orderEvent(t, "metatron.order_placed", 13, validOrder("ord-p3", "player", 13))); err != nil {
		t.Errorf("placement after freeing a slot rejected: %v", err)
	}
}

// TestGuardianOrderLifecycle (spec 029): active → triggered | cancelled | expired
// is one-way; a second terminal on the same order is refused at the door (the
// cancel/expiry/trigger race resolves there — exactly one terminal lands); an
// unknown id is refused.
func TestGuardianOrderLifecycle(t *testing.T) {
	m := worldmap.Generate(7, 32, 32)

	transitions := []struct {
		name  string
		event store.Event
		want  string
	}{
		{"triggered", orderEvent(t, "metatron.order_triggered", 5, OrderTriggeredPayload{ID: "ord-1", MatchedType: "agent.slept", MatchedTick: 5}), "triggered"},
		{"cancelled", orderEvent(t, "metatron.order_cancelled", 5, OrderIDPayload{ID: "ord-1"}), "cancelled"},
		{"expired", orderEvent(t, "metatron.order_expired", 5, OrderIDPayload{ID: "ord-1"}), "expired"},
	}
	for _, tr := range transitions {
		s := NewState(7, m)
		if err := s.Apply(orderEvent(t, "metatron.order_placed", 0, validOrder("ord-1", "player", 0))); err != nil {
			t.Fatal(err)
		}
		if err := s.Apply(tr.event); err != nil {
			t.Fatalf("%s: transition rejected: %v", tr.name, err)
		}
		if s.GuardianOrders[0].Status != tr.want {
			t.Errorf("%s: status = %q, want %q", tr.name, s.GuardianOrders[0].Status, tr.want)
		}
		// A second terminal on the now-consumed order is refused (one-way).
		if err := s.Apply(orderEvent(t, "metatron.order_cancelled", 6, OrderIDPayload{ID: "ord-1"})); err == nil {
			t.Errorf("%s: second terminal on a consumed order must be refused", tr.name)
		}
	}

	// An unknown id is refused for every terminal type.
	s := NewState(7, m)
	for _, e := range []store.Event{
		orderEvent(t, "metatron.order_triggered", 1, OrderTriggeredPayload{ID: "ghost"}),
		orderEvent(t, "metatron.order_cancelled", 1, OrderIDPayload{ID: "ghost"}),
		orderEvent(t, "metatron.order_expired", 1, OrderIDPayload{ID: "ghost"}),
	} {
		if err := s.Apply(e); err == nil {
			t.Errorf("%s on an unknown id must be refused", e.Type)
		}
	}
}

// TestGuardianOrderExpiryExecutor (spec 029): the executor emits
// metatron.order_expired at the first tick ≥ expires_tick (the charge_regenerated
// pattern), exactly once — the emitted event transitions the order to expired, so
// no later tick re-emits. Deterministic in replay by construction.
func TestGuardianOrderExpiryExecutor(t *testing.T) {
	m := worldmap.Generate(7, 32, 32)
	s := NewState(7, m)
	o := validOrder("ord-exp", "player", 0)
	o.ExpiresTick = 1 * ticksPerGameDay // 86400
	if err := s.Apply(orderEvent(t, "metatron.order_placed", 0, o)); err != nil {
		t.Fatal(err)
	}
	if got := countType(stepEvents(s, m, o.ExpiresTick-1), "metatron.order_expired"); got != 0 {
		t.Errorf("pre-expiry tick emitted %d order_expired, want 0", got)
	}
	expiryEvents := stepEvents(s, m, o.ExpiresTick)
	if got := countType(expiryEvents, "metatron.order_expired"); got != 1 {
		t.Fatalf("expiry tick emitted %d order_expired, want 1", got)
	}
	// Apply the expiry (as the loop would) → the order is consumed.
	for _, e := range expiryEvents {
		if e.Type == "metatron.order_expired" {
			if err := s.Apply(e); err != nil {
				t.Fatal(err)
			}
		}
	}
	if s.GuardianOrders[0].Status != "expired" {
		t.Errorf("order status = %q, want expired", s.GuardianOrders[0].Status)
	}
	// A later tick does not re-emit (the order is no longer active).
	if got := countType(stepEvents(s, m, o.ExpiresTick+ticksPerGameDay), "metatron.order_expired"); got != 0 {
		t.Errorf("post-expiry tick re-emitted %d order_expired, want 0", got)
	}
}

// TestSurvivalWatchReducer (spec 059 US1, FR-002): a system-origin survival watch
// is (a) admitted without a valid TTL (non-expiring — the bounds are skipped),
// (b) exempt from the player-order cap, (c) refused for player cancellation at the
// door, and (d) skipped by the executor's expiry sweep. An unknown survival kind
// or a player-origin survival value is refused.
func TestSurvivalWatchReducer(t *testing.T) {
	m := worldmap.Generate(7, 32, 32)

	// (a) Non-expiring: a survival watch whose ExpiresTick would be an illegal TTL
	// (0 days from placed) still lands — the TTL bounds do not apply.
	s := NewState(7, m)
	watches := SurvivalWatchDefs(0)
	for _, o := range watches {
		if o.ExpiresTick-o.PlacedTick >= GuardianOrderTTLMinDays*ticksPerGameDay {
			t.Fatalf("fixture %s is not TTL-illegal; the exemption is not being exercised", o.ID)
		}
		if err := s.Apply(orderEvent(t, "metatron.order_placed", 0, o)); err != nil {
			t.Fatalf("survival watch %s refused despite the TTL exemption: %v", o.ID, err)
		}
	}
	if len(s.GuardianOrders) != 3 {
		t.Fatalf("survival watches landed %d, want 3", len(s.GuardianOrders))
	}

	// (b) Cap-exempt: the three survival watches stand AND the player may still
	// place a full cap of player orders.
	for i := 0; i < GuardianPlayerOrderCap; i++ {
		id := "ord-p" + string(rune('0'+i))
		if err := s.Apply(orderEvent(t, "metatron.order_placed", int64(100+i), validOrder(id, "player", int64(100+i)))); err != nil {
			t.Fatalf("player order %d refused with survival watches standing (cap must count only player orders): %v", i, err)
		}
	}
	if err := s.Apply(orderEvent(t, "metatron.order_placed", 200, validOrder("ord-p3", "player", 200))); err == nil {
		t.Error("a 4th player order was admitted — the cap must still bite on player orders")
	}

	// (c) Player cancel of a survival watch refuses at the door, in-fiction-worthy.
	err := s.Apply(orderEvent(t, "metatron.order_cancelled", 300, OrderIDPayload{ID: "sys-watch-" + SurvivalNearDeath}))
	if err == nil || !strings.Contains(err.Error(), "survival watch") {
		t.Errorf("survival watch cancel not refused at the door: %v", err)
	}
	// The watch is untouched.
	for i := range s.GuardianOrders {
		if s.GuardianOrders[i].ID == "sys-watch-"+SurvivalNearDeath && s.GuardianOrders[i].Status != "active" {
			t.Errorf("a refused cancel still changed the survival watch: %+v", s.GuardianOrders[i])
		}
	}

	// (d) The executor's expiry sweep never expires a survival watch — even far
	// past any TTL. A clean state with ONLY the survival watches (no player orders
	// whose real TTLs would legitimately expire and confound the count).
	sd := NewState(7, m)
	for _, o := range SurvivalWatchDefs(0) {
		if err := sd.Apply(orderEvent(t, "metatron.order_placed", 0, o)); err != nil {
			t.Fatal(err)
		}
	}
	if got := countType(stepEvents(sd, m, 30*ticksPerGameDay), "metatron.order_expired"); got != 0 {
		t.Errorf("expiry sweep emitted %d order_expired for standing survival watches, want 0", got)
	}
	for i := range sd.GuardianOrders {
		if sd.GuardianOrders[i].Status != "active" {
			t.Errorf("survival watch %s no longer active after 30 days: %q", sd.GuardianOrders[i].ID, sd.GuardianOrders[i].Status)
		}
	}

	// Unknown / mis-origin survival discriminator is refused.
	bad := validOrder("ord-bad", "system", 0)
	bad.Survival = "boredom"
	if err := NewState(7, m).Apply(orderEvent(t, "metatron.order_placed", 0, bad)); err == nil {
		t.Error("an unknown survival kind must be refused")
	}
	badOrigin := validOrder("ord-bad2", "player", 0)
	badOrigin.Survival = SurvivalNearDeath
	if err := NewState(7, m).Apply(orderEvent(t, "metatron.order_placed", 0, badOrigin)); err == nil {
		t.Error("a player-origin survival watch must be refused")
	}
}

// TestGuardianOrdersSnapshotUpgrade (spec 029, FR-004): a pre-029 snapshot (no
// metatron_orders field) upgrades to a nil order set; a modern snapshot with
// orders round-trips byte-stably.
func TestGuardianOrdersSnapshotUpgrade(t *testing.T) {
	m := worldmap.Generate(7, 32, 32)
	// Pre-029 shape: strip the field entirely.
	var asMap map[string]json.RawMessage
	if err := json.Unmarshal(NewState(7, m).Marshal(), &asMap); err != nil {
		t.Fatal(err)
	}
	if _, present := asMap["metatron_orders"]; present {
		t.Error("an empty order set must be omitempty (absent from the snapshot)")
	}
	delete(asMap, "metatron_orders")
	oldBytes, _ := json.Marshal(asMap)
	restored := NewState(7, m)
	if err := json.Unmarshal(oldBytes, restored); err != nil {
		t.Fatal(err)
	}
	if restored.GuardianOrders != nil {
		t.Errorf("pre-029 snapshot restored %d orders, want nil", len(restored.GuardianOrders))
	}
	// Modern shape with an order round-trips.
	withOrder := NewState(7, m)
	if err := withOrder.Apply(orderEvent(t, "metatron.order_placed", 0, validOrder("ord-rt", "player", 0))); err != nil {
		t.Fatal(err)
	}
	rt := NewState(7, m)
	if err := json.Unmarshal(withOrder.Marshal(), rt); err != nil {
		t.Fatal(err)
	}
	if len(rt.GuardianOrders) != 1 || rt.GuardianOrders[0].ID != "ord-rt" || rt.GuardianOrders[0].Status != "active" {
		t.Errorf("order did not round-trip: %+v", rt.GuardianOrders)
	}
}

// TestGuardianOrdersReplayIdentically (spec 029, FR-006/SC-002): a state rebuilt
// by replaying a recorded order lifecycle matches the live sequence exactly —
// replay reconstructs state only (it applies the recorded events; it never fires
// a trigger, which is a live-only injection).
func TestGuardianOrdersReplayIdentically(t *testing.T) {
	m := worldmap.Generate(7, 32, 32)
	events := []store.Event{
		orderEvent(t, "metatron.order_placed", 0, validOrder("ord-a", "player", 0)),
		orderEvent(t, "metatron.order_placed", 1, validOrder("ord-b", "player", 1)),
		orderEvent(t, "metatron.order_placed", 2, validOrder("ord-c", "system", 2)),
		orderEvent(t, "metatron.order_triggered", 10, OrderTriggeredPayload{ID: "ord-a", MatchedType: "agent.slept", MatchedTick: 10}),
		orderEvent(t, "metatron.order_cancelled", 11, OrderIDPayload{ID: "ord-b"}),
		orderEvent(t, "metatron.order_expired", 12, OrderIDPayload{ID: "ord-c"}),
	}
	live := NewState(7, m)
	replayed := NewState(7, m)
	for _, e := range events {
		if err := live.Apply(e); err != nil {
			t.Fatalf("live apply %s: %v", e.Type, err)
		}
	}
	for _, e := range events {
		if err := replayed.Apply(e); err != nil {
			t.Fatalf("replay apply %s: %v", e.Type, err)
		}
	}
	if live.Hash() != replayed.Hash() {
		t.Fatal("state hashes diverge under order-lifecycle replay")
	}
	wantStatus := map[string]string{"ord-a": "triggered", "ord-b": "cancelled", "ord-c": "expired"}
	for _, o := range live.GuardianOrders {
		if wantStatus[o.ID] != o.Status {
			t.Errorf("order %q status = %q, want %q", o.ID, o.Status, wantStatus[o.ID])
		}
	}
}

// TestGuardianOrderPrune (spec 029, data-model §1): the retention prune keeps
// EVERY active order plus the most recent guardianOrderRetain (32) non-active
// ones, dropping older consumed orders oldest-first while preserving slice order.
func TestGuardianOrderPrune(t *testing.T) {
	const nonActive = guardianOrderRetain + 8
	var orders []GuardianOrder
	// nonActive cancelled orders (placed_tick 0..N-1), then two active ones.
	for i := 0; i < nonActive; i++ {
		o := validOrder("", "player", int64(i))
		o.Status = "cancelled"
		orders = append(orders, o)
	}
	a1 := validOrder("", "player", 1000)
	a1.Status = "active"
	a2 := validOrder("", "player", 1001)
	a2.Status = "active"
	orders = append(orders, a1, a2)

	got := pruneGuardianOrders(orders)
	// Both active orders survive; only 32 of the non-active do.
	active, cancelled := 0, 0
	for _, o := range got {
		switch o.Status {
		case "active":
			active++
		case "cancelled":
			cancelled++
		}
	}
	if active != 2 {
		t.Errorf("prune dropped an active order: %d active survive, want 2", active)
	}
	if cancelled != guardianOrderRetain {
		t.Errorf("retained %d non-active, want %d", cancelled, guardianOrderRetain)
	}
	// Oldest-first drop: the surviving non-active start at placed_tick (N - 32).
	if got[0].PlacedTick != int64(nonActive-guardianOrderRetain) {
		t.Errorf("oldest surviving non-active placed_tick = %d, want %d", got[0].PlacedTick, nonActive-guardianOrderRetain)
	}
	// Order preserved: non-active block still ascending, actives last.
	for i := 1; i < cancelled; i++ {
		if got[i].PlacedTick <= got[i-1].PlacedTick {
			t.Errorf("prune reordered the non-active block at %d", i)
		}
	}
	if got[len(got)-1].PlacedTick != 1001 || got[len(got)-2].PlacedTick != 1000 {
		t.Error("active orders were not preserved at the tail in order")
	}

	// Below the bound, prune is a no-op (identity).
	small := orders[:5]
	if pruned := pruneGuardianOrders(small); len(pruned) != len(small) {
		t.Errorf("prune under the bound changed the slice: %d → %d", len(small), len(pruned))
	}
}

package sim

import (
	"encoding/json"
	"fmt"
	"unicode/utf8"

	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/tool"
)

// Metatron's world-visible surface (TASK-12): the charge economy and the
// nudge event. Everything else about the angel (charter, soul, console)
// lives outside deterministic space; only recorded events reach here.

const (
	// chargeRegenTicks: one charge per 6 game hours, at absolute boundaries
	// (multiples of 21600 ticks) — a pure function of the clock.
	chargeRegenTicks = 6 * 3600
	// MetatronChargeCap bounds the bank; MetatronGenesisCharges is day-1
	// grace (a reign begins with one favor).
	MetatronChargeCap      = 3
	MetatronGenesisCharges = 1
	// MetatronChargeRegenTicks is chargeRegenTicks exported for internal/tui
	// (guardian strip, spec 050: the next-regen forecast derives the next
	// absolute boundary from this cadence), mirroring the BulkCap/
	// MetatronChargeCap export pattern (agents.go) — sim stays the single
	// source of truth for the doctrine constant; the TUI never carries its
	// own copy of "6 game hours".
	MetatronChargeRegenTicks = chargeRegenTicks
)

// NudgeTextMax caps the villager-bound rendering — read from the tool registry
// (spec 014 T021/R7; re-pointed at send_vision when spec 029 retired the nudges):
// the influence tools' TextCapBytes (400). The reducer dry-run stays the
// enforcer; the registry is the single source of the cap, so the enforcer and
// the metatron-side truncation can never carry divergent literals.
var NudgeTextMax = func() int {
	t, _ := tool.Lookup("send_vision")
	return t.Cost.TextCapBytes
}()

type (
	// MetatronNudgedPayload is the injected spend + record: form "dream"
	// (exactly one living target) or "omen" (every villager alive at
	// landing, recorded explicitly).
	MetatronNudgedPayload struct {
		Form    string `json:"form"`
		Targets []int  `json:"targets"`
		Text    string `json:"text"`
	}
	// ChargeRegeneratedPayload is empty — the event row's tick is the
	// boundary crossed.
	ChargeRegeneratedPayload struct{}
)

const (
	// ticksPerGameDay is a game day in ticks (1 tick = 1 game second, so
	// clock.secondsPerDay = 24×3600). Standing-order TTLs are game days (spec 029).
	ticksPerGameDay = 24 * 3600
	// MetatronOrderTTLMinDays / MaxDays bound a standing order's lifetime
	// (spec 029 FR-007): player-specifiable, default 3, capped 1..7 game days.
	MetatronOrderTTLMinDays = 1
	MetatronOrderTTLMaxDays = 7
	// MetatronPlayerOrderCap is the concurrent ACTIVE player-placed order cap
	// (FR-007); system-origin deferral orders are exempt (FR-012).
	MetatronPlayerOrderCap = 3
	// MetatronOriginPlayer / MetatronOriginSystem are the two MetatronOrder.Origin
	// values (spec 029): a player console monitor_and_act places "player" orders;
	// the daytime-omen deferral (spec 029 T016) and the survival watches (spec 059)
	// place "system" orders. Named here so the origin-keyed exemptions (cap/TTL/
	// cancel) have a single home rather than scattered string literals.
	MetatronOriginPlayer = "player"
	MetatronOriginSystem = "system"
)

// Survival watch kinds (spec 059 US1): the three canonical system-origin survival
// watches. A MetatronOrder whose Survival is one of these is matched by the
// survival-band predicate (not the structural orderMatches), is cap/TTL/cancel
// exempt (origin-keyed), and drives a survival-authority turn (US2). Empty
// Survival ("") is every pre-059 order — the ordinary structural watch.
const (
	SurvivalNearDeath  = "near_death"
	SurvivalStarvation = "starvation"
	SurvivalExposure   = "exposure"
)

// Survival watch bands (spec 059 FR-008): the danger thresholds the three
// survival watches match on, each REUSING an existing sim doctrine constant as
// its single home rather than a new magic number — promoted-dial-ready (named,
// one place) but deliberately NOT added to tuning.json (dials are earned by
// evidence). near_death reuses the near-death latch band (nearDeathBelow) and
// its hysteresis reset (nearDeathResetAt); starvation/exposure reuse decayNeeds'
// OWN health-loss predicate — Food==0 / Warmth==0, the exact conditions that
// drain health and stamp the "starvation"/"exposure" death causes — with the
// reflex hunger band (hungryAt) and the freezing-night band (coldNightBelow) as
// the recovery (re-arm) thresholds so a watch cannot flap on a one-tick wobble.
const (
	SurvivalNearDeathBelow = nearDeathBelow   // 200: health danger band (agents.go)
	SurvivalNearDeathRearm = nearDeathResetAt // 400: recovered — re-arm
	SurvivalStarvingAt     = 0                // Food==0: out of food, losing health
	SurvivalStarvingRearm  = hungryAt         // 350: fed enough to re-arm
	SurvivalFreezingAt     = 0                // Warmth==0: freezing, losing health
	SurvivalFreezingRearm  = coldNightBelow   // 350: warm enough to re-arm
)

// IsSurvivalKind reports whether s names a canonical survival watch kind
// (spec 059): the reducer uses it to keep the order_placed door authoritative on
// the Survival discriminator (an unknown non-empty value is refused).
func IsSurvivalKind(s string) bool {
	switch s {
	case SurvivalNearDeath, SurvivalStarvation, SurvivalExposure:
		return true
	}
	return false
}

// SurvivalWatchDefs returns the three canonical system-origin survival watches
// (spec 059 US1/FR-001) as ready-to-land MetatronOrders at the given tick — the
// SINGLE home the genesis/boot seeder and tests both build from. Ids are fixed
// and human-readable ("sys-watch-<kind>") so re-seeding is idempotent (the
// reducer rejects a duplicate id; the boot guard skips re-injection). Each
// watches agent.needs_changed (the per-game-minute heartbeat carrying the danger
// band, matched by the survival-band predicate, not the structural filter),
// pins no agent (Agent -1 = all villagers), and is non-expiring (ExpiresTick is
// ignored for a survival watch — set to PlacedTick as an honest placeholder).
// Condition/Action are the in-fiction standing duty the survival turn narrates
// under; the specific endangered villager is supplied at trigger time.
func SurvivalWatchDefs(tick int64) []MetatronOrder {
	def := func(kind, cond, act string) MetatronOrder {
		return MetatronOrder{
			ID:          "sys-watch-" + kind,
			Origin:      MetatronOriginSystem,
			Survival:    kind,
			Condition:   cond,
			Action:      act,
			EventTypes:  []string{"agent.needs_changed"},
			Agent:       -1,
			PlacedTick:  tick,
			ExpiresTick: tick, // ignored — a survival watch never expires
			Status:      "active",
		}
	}
	return []MetatronOrder{
		def(SurvivalNearDeath, "a villager is near death",
			"A villager stands at the brink of death. Act on your own authority to save a life if you can — send a vision or work a miracle — or, if nothing can be done, keep the watch and record it."),
		def(SurvivalStarvation, "a villager is starving",
			"A villager has run out of food and is starving. Act on your own authority to save a life if you can — send a vision toward food, or work a miracle — or, if nothing can be done, keep the watch and record it."),
		def(SurvivalExposure, "a villager is freezing",
			"A villager is freezing in the cold and is losing their life to exposure. Act on your own authority to save a life if you can — send a vision toward warmth, or work a miracle — or, if nothing can be done, keep the watch and record it."),
	}
}

const (
	// metatronOrderRetain bounds retained NON-ACTIVE orders (data-model §1): the
	// slice keeps every active order plus the most recent 32 consumed ones, so
	// the status/trail shows recent history without unbounded growth.
	metatronOrderRetain = 32
)

// MetatronOrder is one event-sourced standing order (spec 029, data-model §1): a
// pre-authorized watch-and-act instruction placed via monitor_and_act. Its
// lifecycle (active → triggered | cancelled | expired) is driven entirely by
// recorded events, so it reconstructs identically through snapshots, restart,
// and from-genesis replay; replay only reconstructs state — it never triggers.
type MetatronOrder struct {
	ID          string   `json:"id"`                 // "ord-<placedTick>-<seq>" (research R7)
	Origin      string   `json:"origin"`             // "player" | "system"
	Condition   string   `json:"condition"`          // original NL, ≤300 chars
	Action      string   `json:"action"`             // NL action instruction, ≤400 chars
	EventTypes  []string `json:"event_types"`        // structural predicate: non-empty
	Agent       int      `json:"agent"`              // villager index, -1 = any
	Keywords    []string `json:"keywords,omitempty"` // coarse text filter, lowercase
	Confirm     bool     `json:"confirm,omitempty"`  // fuzzy: needs the watch confirm
	PlacedTick  int64    `json:"placed_tick"`
	ExpiresTick int64    `json:"expires_tick"`       // placed + ttl_days game days (IGNORED for a survival watch — non-expiring)
	Status      string   `json:"status"`             // "active" | "triggered" | "cancelled" | "expired"
	Survival    string   `json:"survival,omitempty"` // spec 059: "" = ordinary structural order; else a survival watch kind (near_death|starvation|exposure)
}

// OrderTriggeredPayload records a matched order's one-shot consumption (spec
// 029): the matched event's type + tick ride along for the trail. Injected by
// the trigger worker (Batch B), NEVER emitted during replay.
type OrderTriggeredPayload struct {
	ID          string `json:"id"`
	MatchedType string `json:"matched_type"`
	MatchedTick int64  `json:"matched_tick"`
}

// OrderIDPayload is the bare-id payload shared by metatron.order_cancelled
// (injected — cancel_order) and metatron.order_expired (executor-emitted, a pure
// function of state + tick, like charge_regenerated).
type OrderIDPayload struct {
	ID string `json:"id"`
}

// CharterObservedPayload — metatron.charter_observed (spec 044 US2, FR-008):
// the charter revision a Metatron turn actually ran under, identified by a
// short content hash of the EFFECTIVE charter text (post-fallback,
// post-truncation — what loadCharter returned), plus whether that text is
// the authored default. Emitted by the turn pipeline through the
// inject_social door only when the fingerprint differs from
// State.CharterFingerprint (the first turn always emits), giving deaths an
// event-sourced charter-revision timeline to align against. Evidence only:
// no scoring fields, by contract (contracts/events.md).
type CharterObservedPayload struct {
	Fingerprint string `json:"fingerprint"`
	Default     bool   `json:"default"`
}

// applyMetatron is the reducer arm for metatron.* events. The nudged arm
// validates rather than clamps: the InjectSocial dry-run runs this on a
// state copy, so invalid spends are rejected at the door and recorded
// events always re-apply cleanly at the same position in replay.
func (s *State) applyMetatron(e store.Event) error {
	switch e.Type {
	case "metatron.charge_regenerated":
		if s.MetatronCharges < MetatronChargeCap {
			s.MetatronCharges++
		}
	case "metatron.nudged":
		var p MetatronNudgedPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		if s.MetatronCharges <= 0 {
			return fmt.Errorf("apply %s: no charges banked", e.Type)
		}
		// Form validation (spec 029): the metatron.nudged form domain is
		// {vision, omen, dream}. A vision reaches exactly one living villager at
		// any hour; an omen reaches ≥1 living villagers and lands ONLY at night
		// (State.Night); dream is the RETIRED legacy form, grandfathered here
		// (exactly one target) so pre-029 histories replay to identical state —
		// but no tool, handler, or roster entry can produce a NEW one, so
		// structural absence is the guarantee dreams cannot land afresh. This
		// explicit form switch REPLACES the spec-014 OnRoster(RosterMetatron,
		// "nudge_"+form) check, which could no longer hold once nudge_dream/
		// nudge_omen left the registry (contracts/events.md).
		switch p.Form {
		case "vision":
			if len(p.Targets) != 1 {
				return fmt.Errorf("apply %s: vision needs exactly one target, got %d", e.Type, len(p.Targets))
			}
		case "omen":
			if len(p.Targets) == 0 {
				return fmt.Errorf("apply %s: omen needs targets", e.Type)
			}
			if !s.Night {
				return fmt.Errorf("apply %s: an omen may land only at night", e.Type)
			}
		case "dream":
			// Legacy (pre-029), replay-only — grandfathered so recorded histories
			// reproduce identically; unreachable from any live tool.
			if len(p.Targets) != 1 {
				return fmt.Errorf("apply %s: dream needs exactly one target, got %d", e.Type, len(p.Targets))
			}
		default:
			return fmt.Errorf("apply %s: unknown form %q", e.Type, p.Form)
		}
		for _, t := range p.Targets {
			if t < 0 || t >= len(s.Agents) {
				return fmt.Errorf("apply %s: unknown target %d", e.Type, t)
			}
			if s.Agents[t].Dead {
				return fmt.Errorf("apply %s: target %s is dead", e.Type, s.Agents[t].Name)
			}
		}
		if p.Text == "" || len(p.Text) > NudgeTextMax {
			return fmt.Errorf("apply %s: text length %d outside 1..%d", e.Type, len(p.Text), NudgeTextMax)
		}
		s.MetatronCharges--
	case "metatron.place_revealed":
		// Spec 041 (FR-014, T032): a vision's divine place grant. Validates
		// rather than clamps (the nudged arm's contract — the InjectSocial
		// dry-run runs this on a state copy, so an invalid reveal is rejected
		// at the door and a recorded one always re-applies): the target must
		// live, and every fact must name a REAL place (groundFactPresent —
		// the god reveals what is; a false vision is not a channel). Seen,
		// Provenance, and Detail are stamped here, normatively (the
		// order-Status-ignored shape): Seen = the landing tick, Detail =
		// ground truth at landing (a fire's FuelUntil), a pure function of
		// (state, event) so live and replay agree byte-for-byte. A map-less
		// agent skips the upsert — the reducer stays total (agent.saw shape).
		var p PlaceRevealedPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		if p.Agent < 0 || p.Agent >= len(s.Agents) {
			return fmt.Errorf("apply %s: unknown target %d", e.Type, p.Agent)
		}
		a := &s.Agents[p.Agent]
		if a.Dead {
			return fmt.Errorf("apply %s: target %s is dead", e.Type, a.Name)
		}
		if len(p.Facts) == 0 {
			return fmt.Errorf("apply %s: no facts to reveal", e.Type)
		}
		for _, f := range p.Facts {
			if s.m != nil && !groundFactPresent(s, s.m, f) {
				return fmt.Errorf("apply %s: no %s at (%d,%d)", e.Type, f.Kind, f.X, f.Y)
			}
			if a.Map != nil {
				a.Map.upsertFact(PlaceFact{Kind: f.Kind, X: f.X, Y: f.Y, Seen: e.Tick,
					Provenance: ProvenanceRevealed, Detail: groundFactDetail(s, f)})
			}
		}
	case "metatron.order_placed":
		var o MetatronOrder
		if err := json.Unmarshal(e.Payload, &o); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		if o.ID == "" {
			return fmt.Errorf("apply %s: empty order id", e.Type)
		}
		// Duplicate id in ANY status is rejected — ids are assigned once and
		// consumed orders are retained, so a reused id would corrupt the trail.
		for i := range s.MetatronOrders {
			if s.MetatronOrders[i].ID == o.ID {
				return fmt.Errorf("apply %s: duplicate order id %q", e.Type, o.ID)
			}
		}
		switch o.Origin {
		case MetatronOriginPlayer, MetatronOriginSystem:
		default:
			return fmt.Errorf("apply %s: unknown origin %q", e.Type, o.Origin)
		}
		// A survival watch (spec 059) is a system-origin order with a known kind;
		// it is exempt from the TTL bounds below (non-expiring). A non-empty
		// Survival on a player order, or an unknown kind, is refused — the door
		// stays authoritative on the discriminator.
		if o.Survival != "" {
			if o.Origin != MetatronOriginSystem {
				return fmt.Errorf("apply %s: survival watch must be system origin", e.Type)
			}
			if !IsSurvivalKind(o.Survival) {
				return fmt.Errorf("apply %s: unknown survival kind %q", e.Type, o.Survival)
			}
		}
		if len(o.EventTypes) == 0 {
			return fmt.Errorf("apply %s: order has no event_types (uncompilable condition)", e.Type)
		}
		// TTL bounds hold for every order EXCEPT a survival watch, which is
		// non-expiring by nature (spec 059 FR-002) — its ExpiresTick is ignored
		// here and by the executor's expiry sweep alike (origin-keyed exemption,
		// not a giant TTL).
		if o.Survival == "" {
			if ttl := o.ExpiresTick - o.PlacedTick; ttl < MetatronOrderTTLMinDays*ticksPerGameDay || ttl > MetatronOrderTTLMaxDays*ticksPerGameDay {
				return fmt.Errorf("apply %s: ttl %d ticks outside %d..%d game days", e.Type, ttl, MetatronOrderTTLMinDays, MetatronOrderTTLMaxDays)
			}
		}
		if o.Agent < -1 || o.Agent >= len(s.Agents) {
			return fmt.Errorf("apply %s: agent index %d out of range", e.Type, o.Agent)
		}
		if utf8.RuneCountInString(o.Condition) > 300 {
			return fmt.Errorf("apply %s: condition over 300 chars", e.Type)
		}
		if utf8.RuneCountInString(o.Action) > 400 {
			return fmt.Errorf("apply %s: action over 400 chars", e.Type)
		}
		// Concurrent cap: at most 3 ACTIVE player-origin orders; system-origin
		// deferral orders are exempt (already-authorized acts, FR-012).
		if o.Origin == "player" {
			active := 0
			for i := range s.MetatronOrders {
				if s.MetatronOrders[i].Origin == "player" && s.MetatronOrders[i].Status == "active" {
					active++
				}
			}
			if active >= MetatronPlayerOrderCap {
				return fmt.Errorf("apply %s: %d player orders already active (cap %d)", e.Type, active, MetatronPlayerOrderCap)
			}
		}
		// The status field is IGNORED on the payload — an order always lands
		// active (data-model §2), then the retention prune runs.
		o.Status = "active"
		s.MetatronOrders = pruneMetatronOrders(append(s.MetatronOrders, o))
	case "metatron.order_triggered":
		var p OrderTriggeredPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		return s.transitionMetatronOrder(e.Type, p.ID, "triggered")
	case "metatron.order_cancelled":
		var p OrderIDPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		// A survival watch is the angel's nature, not a player configuration
		// (spec 059 FR-002): the player-order surface cannot cancel it. Refuse at
		// the door, keyed on the Survival discriminator, so the in-fiction refusal
		// is authoritative rather than advisory.
		for i := range s.MetatronOrders {
			if s.MetatronOrders[i].ID == p.ID && s.MetatronOrders[i].Survival != "" {
				return fmt.Errorf("apply %s: order %q is a survival watch and cannot be cancelled", e.Type, p.ID)
			}
		}
		return s.transitionMetatronOrder(e.Type, p.ID, "cancelled")
	case "metatron.order_expired":
		var p OrderIDPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		return s.transitionMetatronOrder(e.Type, p.ID, "expired")
	case "metatron.charter_observed":
		// Spec 044 US2 (FR-008): the charter-revision timeline. State keeps only
		// the CURRENT fingerprint; the full timeline lives in the event log,
		// where the morgue's render scan aligns each death against the most
		// recent observation at or before it.
		var p CharterObservedPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		if p.Fingerprint == "" {
			return fmt.Errorf("apply %s: empty fingerprint", e.Type)
		}
		s.CharterFingerprint = p.Fingerprint
	}
	return nil
}

// transitionMetatronOrder moves the order named id from active to a terminal
// status (spec 029, one-way). An unknown id or an order not currently active is
// rejected at the door — this is where the cancel/expiry/trigger races resolve:
// exactly one terminal lands, and the loser hits a non-active order and refuses
// (contracts/events.md edge cases).
func (s *State) transitionMetatronOrder(eventType, id, to string) error {
	for i := range s.MetatronOrders {
		if s.MetatronOrders[i].ID != id {
			continue
		}
		if s.MetatronOrders[i].Status != "active" {
			return fmt.Errorf("apply %s: order %q is not active (status %q)", eventType, id, s.MetatronOrders[i].Status)
		}
		s.MetatronOrders[i].Status = to
		return nil
	}
	return fmt.Errorf("apply %s: unknown order %q", eventType, id)
}

// pruneMetatronOrders retains every active order plus the most recent
// metatronOrderRetain (32) non-active ones, dropping the oldest consumed orders
// first while preserving slice order (data-model §1). Deterministic — a pure
// function of the append-ordered slice, so replay prunes identically.
func pruneMetatronOrders(orders []MetatronOrder) []MetatronOrder {
	nonActive := 0
	for i := range orders {
		if orders[i].Status != "active" {
			nonActive++
		}
	}
	drop := nonActive - metatronOrderRetain
	if drop <= 0 {
		return orders
	}
	out := make([]MetatronOrder, 0, len(orders)-drop)
	for i := range orders {
		if orders[i].Status != "active" && drop > 0 {
			drop--
			continue
		}
		out = append(out, orders[i])
	}
	return out
}

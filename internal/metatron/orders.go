package metatron

// Standing orders (spec 029 US2/US3): the event-sourced watch-and-act machinery.
// This file holds the pure predicate matcher (orderMatches — evaluated for free
// in the absorb path, zero model cost per non-matching event, SC-001), the two
// door-landing helpers a console turn's handlers wrap (placeOrder / cancelOrder),
// and the id assignment (research R7). The reducer dry-run is the door authority
// for every lifecycle transition — the cap, the ttl bounds, the agent range, and
// the cancel/expiry/trigger races all resolve there (internal/sim/metatron.go);
// these helpers map a door rejection to in-fiction counsel the loop feeds back as
// a rejected_gate. The trigger pipeline that fires a matched order (the absorb
// worker + system turn) lives alongside in metatron.go / this file's T013 half.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
)

// ticksPerGameDay is a game day in ticks (1 tick = 1 game second). MIRRORED from
// sim's unexported constant (internal/sim/metatron.go) — the reducer validates a
// standing order's ttl in [1..7] game days against the same literal, so the
// ExpiresTick this package computes and the door-side bound can never diverge.
const ticksPerGameDay = 24 * 3600

// orderArgs is the parsed monitor_and_act tool-call surface (spec 029 R5): the
// turn model itself is the compiler, supplying the compiled standing-order
// structure in the tool call. The driver's schema-lite walker already gated shape
// (event_types array + enum membership, keyword bounds, ttl range), so this is a
// lenient reader — the reducer dry-run is the semantic door.
type orderArgs struct {
	Condition  string   `json:"condition"`
	Action     string   `json:"action"`
	EventTypes []string `json:"event_types"`
	Agent      string   `json:"agent"`
	Keywords   []string `json:"keywords"`
	Confirm    bool     `json:"confirm"`
	TTLDays    int      `json:"ttl_days"`
}

// parseOrderArgs decodes a monitor_and_act call's arguments (lenient — the driver
// validated shape).
func parseOrderArgs(raw json.RawMessage) orderArgs {
	var a orderArgs
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &a)
	}
	return a
}

// nextOrderID assigns "ord-<placedTick>-<seq>" (research R7): human-readable,
// deterministic, no RNG draw. seq disambiguates same-tick placements — it is the
// max seq already present at this tick (from the mirror) plus one, floored by a
// serialized per-tick counter so a placement whose predecessor has not yet flowed
// back through Observe still gets a fresh id. Uniqueness is ultimately enforced by
// the reducer (it rejects a duplicate active id).
func (mt *Metatron) nextOrderID(tick int64) string {
	mt.stateMu.Lock()
	defer mt.stateMu.Unlock()
	seq := 0
	prefix := fmt.Sprintf("ord-%d-", tick)
	for i := range mt.orders {
		id := mt.orders[i].ID
		if strings.HasPrefix(id, prefix) {
			if s, err := strconv.Atoi(strings.TrimPrefix(id, prefix)); err == nil && s >= seq {
				seq = s + 1
			}
		}
	}
	if mt.lastPlaceTick == tick && mt.lastPlaceSeq >= seq {
		seq = mt.lastPlaceSeq + 1
	}
	mt.lastPlaceTick = tick
	mt.lastPlaceSeq = seq
	return fmt.Sprintf("ord-%d-%d", tick, seq)
}

// placeOrder compiles a monitor_and_act call into a MetatronOrder and lands it as
// metatron.order_placed through the InjectSocial door (spec 029 US2, T008). The
// door's dry-run is the authority (player cap, ttl bounds, agent range, empty
// event_types); a rejection maps to in-fiction counsel the handler feeds back as a
// rejected_gate. origin is "player" for a console monitor_and_act (Batch C's
// deferral path passes "system"). A semantically uncompilable condition (no
// structural filter — empty event_types) is refused HERE with counsel (research
// R5) rather than at the driver, so the system/deferral caller that bypasses the
// driver is guarded too. Returns the placed order (id for the reply/status) or
// (nil, refusal).
func (mt *Metatron) placeOrder(origin string, a orderArgs, tick int64, grant grantSet) (*sim.MetatronOrder, string) {
	// The monitor_and_act grant gates only PLAYER placements (a console tool call).
	// A system-origin deferral (T016) is internally initiated by an already-granted
	// act — the daytime send_omen that spawned it — so it carries THAT tool's gate
	// (checked in landOmen), not monitor_and_act's: a world granting send_omen but
	// withholding monitor_and_act can still defer a daytime omen to nightfall.
	if origin == "player" && !grant.allows("monitor_and_act") {
		return nil, "that power is not granted in this world"
	}
	if len(a.EventTypes) == 0 {
		return nil, "I can only keep watch for things that leave a mark in the world — name what should happen, and I will watch for it"
	}
	agent := -1
	if name := strings.TrimSpace(a.Agent); name != "" {
		idx := agentIndexByName(name)
		if idx < 0 {
			return nil, fmt.Sprintf("no villager named %q to watch over", name)
		}
		agent = idx
	}
	ttl := a.TTLDays
	if ttl == 0 {
		ttl = 3 // default (spec Assumption): 3 game days
	}
	if ttl < sim.MetatronOrderTTLMinDays || ttl > sim.MetatronOrderTTLMaxDays {
		return nil, fmt.Sprintf("a watch may stand for %d to %d days", sim.MetatronOrderTTLMinDays, sim.MetatronOrderTTLMaxDays)
	}
	keywords := make([]string, 0, len(a.Keywords))
	for _, k := range a.Keywords {
		if k = strings.ToLower(strings.TrimSpace(k)); k != "" {
			keywords = append(keywords, k)
		}
	}
	order := sim.MetatronOrder{
		ID:          mt.nextOrderID(tick),
		Origin:      origin,
		Condition:   a.Condition,
		Action:      a.Action,
		EventTypes:  a.EventTypes,
		Agent:       agent,
		Keywords:    keywords,
		Confirm:     a.Confirm,
		PlacedTick:  tick,
		ExpiresTick: tick + int64(ttl)*ticksPerGameDay,
		Status:      "active",
	}
	batch := []store.Event{{Type: "metatron.order_placed", Payload: mustJSON(order)}}
	if err := mt.social.InjectSocial(batch); err != nil {
		log.Printf("metatron: order rejected at the door: %v", err)
		return nil, orderRefusal(err)
	}
	mt.appendFile(mt.soulPath(), fmt.Sprintf("\n- %s — I set a watch (%s): %q → %q\n",
		clock.Format(mt.replicaTickSafe()), order.ID, order.Condition, order.Action))
	return &order, ""
}

// cancelOrder lands metatron.order_cancelled for the named id through the door
// (spec 029 US2, T008). The reducer rejects an unknown or non-active id — this is
// where the cancel/expiry/trigger race resolves (exactly one terminal lands).
// Returns "" on success or an in-fiction refusal the handler feeds back.
func (mt *Metatron) cancelOrder(id string, grant grantSet) string {
	if !grant.allows("cancel_order") {
		return "that power is not granted in this world"
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return "name the watch you want me to release"
	}
	batch := []store.Event{{Type: "metatron.order_cancelled", Payload: mustJSON(sim.OrderIDPayload{ID: id})}}
	if err := mt.social.InjectSocial(batch); err != nil {
		msg := err.Error()
		switch {
		case strings.Contains(msg, "unknown order"):
			return fmt.Sprintf("I keep no watch called %q", id)
		case strings.Contains(msg, "not active"):
			return fmt.Sprintf("the watch %q has already lapsed", id)
		case strings.Contains(msg, "survival watch"):
			// A survival watch is the angel's own nature (spec 059 FR-002), not a
			// player configuration — it cannot be released by the order surface.
			return "that watch is my own nature — I keep it over every living soul, and I cannot set it aside"
		default:
			return "the world would not let me release that watch (" + msg + ")"
		}
	}
	mt.appendFile(mt.soulPath(), fmt.Sprintf("\n- %s — I released a watch (%s)\n",
		clock.Format(mt.replicaTickSafe()), id))
	return ""
}

// orderRefusal maps a metatron.order_placed door rejection to in-fiction counsel
// (spec 029): the reducer's error strings are the source, translated to the
// angel's voice so the model hears a repairable reason (rejected_gate) rather than
// a raw reducer message.
func orderRefusal(err error) string {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "already active"):
		return "I already keep as many watches as I can hold — release one and I will take up another"
	case strings.Contains(msg, "ttl"):
		return "a watch may stand only for a handful of days"
	case strings.Contains(msg, "no event_types"), strings.Contains(msg, "uncompilable"):
		return "I can only keep watch for things that leave a mark in the world"
	case strings.Contains(msg, "over 300 chars"), strings.Contains(msg, "over 400 chars"):
		return "say the watch more briefly and I will hold it"
	default:
		return "the world would not let me set that watch (" + msg + ")"
	}
}

// needMirror is the turn-side snapshot of one villager's survival-relevant needs
// (spec 059): health/food/warmth, refreshed per absorb batch (mirrorState), read
// by the targeting digest so the turn worker never races the replica.
type needMirror struct{ Health, Food, Warmth int }

// survivalBand reports, for a survival watch kind and a villager's needs, whether
// the villager is IN the danger band (fire) and whether it has recovered above the
// re-arm threshold (clear the latch). The bands REUSE the sim's own doctrine
// constants (spec 059 FR-008) — no thresholds live here. inBand and rearm are
// mutually exclusive by the hysteresis gap (band < rearm), so a villager between
// them (recovering but not yet re-armed) neither fires nor clears — it stays
// latched, which is the intended debounce.
func survivalBand(kind string, n needMirror) (inBand, rearm bool) {
	switch kind {
	case sim.SurvivalNearDeath:
		return n.Health < sim.SurvivalNearDeathBelow, n.Health >= sim.SurvivalNearDeathRearm
	case sim.SurvivalStarvation:
		return n.Food <= sim.SurvivalStarvingAt, n.Food >= sim.SurvivalStarvingRearm
	case sim.SurvivalExposure:
		return n.Warmth <= sim.SurvivalFreezingAt, n.Warmth >= sim.SurvivalFreezingRearm
	}
	return false, false
}

// matchSurvival matches one survival watch against a live event batch (spec 059
// US1). Unlike the structural orderMatches, it evaluates the danger BAND against
// agent.needs_changed payloads, with per-villager hysteresis latching so a
// villager staying in-band does not re-fire the watch every game-minute. Re-arm
// (recovery) is processed for every villager on every batch regardless of the
// in-flight marker, so the latch always tracks reality; a FIRE is gated by
// pendingTrigger (one survival turn per watch in flight, matching the spec's
// "watches don't queue duplicate turns for the same crisis while one is in
// flight") and by the per-villager latch. At most one job is enqueued per watch
// per batch (one survival turn may address any/all villagers in crisis).
func (mt *Metatron) matchSurvival(o sim.MetatronOrder, batch []store.Event, pending bool) {
	for _, e := range batch {
		if e.Type != "agent.needs_changed" {
			continue
		}
		var p sim.NeedsPayload
		if json.Unmarshal(e.Payload, &p) != nil {
			continue
		}
		inBand, rearm := survivalBand(o.Survival, needMirror{Health: p.Health, Food: p.Food, Warmth: p.Warmth})
		mt.stateMu.Lock()
		if rearm {
			if set := mt.survivalLatch[o.ID]; set != nil {
				delete(set, p.Agent)
			}
			mt.stateMu.Unlock()
			continue
		}
		if !inBand || pending || mt.survivalLatch[o.ID][p.Agent] {
			mt.stateMu.Unlock()
			continue
		}
		if mt.survivalLatch[o.ID] == nil {
			mt.survivalLatch[o.ID] = map[int]bool{}
		}
		mt.survivalLatch[o.ID][p.Agent] = true
		mt.pendingTrigger[o.ID] = true
		mt.stateMu.Unlock()
		pending = true // one survival turn per watch per batch; keep processing re-arm
		select {
		case mt.triggerQ <- triggerJob{order: o, matched: e, matchedType: e.Type, matchedTick: e.Tick}:
		default:
			log.Printf("metatron: trigger queue full, survival watch %s dropped", o.ID)
			mt.stateMu.Lock()
			delete(mt.pendingTrigger, o.ID)
			delete(mt.survivalLatch[o.ID], p.Agent)
			mt.stateMu.Unlock()
			pending = false
		}
	}
}

// orderMatches reports whether a live observed event satisfies an order's compiled
// structural predicates (spec 029 US2/R6): the event type is one of the order's
// event_types; if the order pins an agent (>= 0) the event concerns that villager;
// if the order lists keywords the (lowercased) payload contains at least one. A
// PURE function — no state, no model call — so the absorb path evaluates it for
// free (SC-001). Only ACTIVE orders match; a fuzzy order (Confirm) still matches
// structurally here (the confirm step that gates its trigger is Batch C / T021).
func orderMatches(o sim.MetatronOrder, e store.Event) bool {
	if o.Status != "active" {
		return false
	}
	typeOK := false
	for _, t := range o.EventTypes {
		if t == e.Type {
			typeOK = true
			break
		}
	}
	if !typeOK {
		return false
	}
	if o.Agent >= 0 && !eventConcernsAgent(e, o.Agent) {
		return false
	}
	if len(o.Keywords) > 0 {
		hay := strings.ToLower(string(e.Payload))
		hit := false
		for _, k := range o.Keywords {
			if k != "" && strings.Contains(hay, k) {
				hit = true
				break
			}
		}
		if !hit {
			return false
		}
	}
	return true
}

// eventConcernsAgent reports whether an event's payload names the given villager
// index in one of the observable vocabulary's agent-bearing fields (agent / from /
// to). A best-effort structural probe: an unknown or agent-less payload shape
// simply does not match the agent pin (the order stays armed) — never a false
// positive against the wrong villager.
func eventConcernsAgent(e store.Event, idx int) bool {
	var m map[string]json.RawMessage
	if json.Unmarshal(e.Payload, &m) != nil {
		return false
	}
	for _, field := range []string{"agent", "from", "to"} {
		if raw, ok := m[field]; ok {
			var v int
			if json.Unmarshal(raw, &v) == nil && v == idx {
				return true
			}
		}
	}
	return false
}

// --- Trigger pipeline (spec 029 US3, T013/T014) ---

// systemTurnBusyWait bounds how long a triggered system turn waits for the
// single-flight slot before degrading (spec 029 R6): system turns WAIT for the
// slot (unlike the console's fail-fast ErrTurnBusy), but never forever — a wedged
// console turn degrades the trigger to an honest moment rather than hanging.
const systemTurnBusyWait = 90 * time.Second

// triggerJob is one matched order queued for firing (spec 029 US3/US6). It
// snapshots the order at match time (its action/origin/condition are needed to run
// and narrate the system turn), the matched event (rendered into a fuzzy order's
// confirm prompt, T021), the matched type + tick for the trail, and whether this is
// a fuzzy order that must pass the watch confirm before firing (confirm).
type triggerJob struct {
	order       sim.MetatronOrder
	matched     store.Event
	matchedType string
	matchedTick int64
	confirm     bool
}

// confirmRateTicks caps fuzzy-order confirms to one per 30 game minutes per order
// (spec 029 US6/R9): 1800 ticks at 1 tick = 1 game second. A tighter storm of
// matching events costs at most one watch call per window per order.
const confirmRateTicks = 30 * 60

// matchOrders scans active orders against a just-applied LIVE event batch and
// enqueues a job for each structural hit (spec 029 US3/US6/R6/R9). Called by the
// absorb goroutine AFTER replica apply + mirror refresh, so it is live-only by
// construction (replay never runs the angel). Orders fire in order-id order within
// a batch, at most once — pendingTrigger dedups an order already queued but not yet
// resolved, and one job is enqueued per order per batch. A fuzzy order (Confirm) is
// matched structurally here too, but its hit is routed as a CONFIRM job (confirm
// true) — the worker runs one watch call before firing (T021) — and is rate-capped
// to one confirm per confirmRateTicks per order via lastConfirmTick; a rate-capped
// hit is logged and skipped so a storm never triggers a flood of watch calls.
func (mt *Metatron) matchOrders(batch []store.Event) {
	if mt.social == nil {
		return
	}
	mt.stateMu.Lock()
	orders := append([]sim.MetatronOrder(nil), mt.orders...)
	mt.stateMu.Unlock()
	sort.Slice(orders, func(i, j int) bool { return orders[i].ID < orders[j].ID })
	for i := range orders {
		o := orders[i]
		if o.Status != "active" {
			continue
		}
		mt.stateMu.Lock()
		pending := mt.pendingTrigger[o.ID]
		mt.stateMu.Unlock()
		// Survival watches (spec 059) use the danger-band matcher with its own
		// hysteresis latch — re-arm must run even while a turn is in flight, so the
		// pending short-circuit below does not apply to them.
		if o.Survival != "" {
			mt.matchSurvival(o, batch, pending)
			continue
		}
		if pending {
			continue
		}
		for _, e := range batch {
			if !orderMatches(o, e) {
				continue
			}
			// Fuzzy orders pay the rate cap BEFORE the in-flight marker: a hit within
			// confirmRateTicks of the last confirm is skipped (logged), leaving the
			// order armed for a later window (R9).
			if o.Confirm {
				mt.stateMu.Lock()
				last, seen := mt.lastConfirmTick[o.ID]
				capped := seen && e.Tick-last < confirmRateTicks
				mt.stateMu.Unlock()
				if capped {
					log.Printf("metatron: fuzzy order %s confirm rate-capped (last %d, event %d)", o.ID, last, e.Tick)
					break // one hit per order per batch, capped or not
				}
			}
			mt.stateMu.Lock()
			mt.pendingTrigger[o.ID] = true
			if o.Confirm {
				mt.lastConfirmTick[o.ID] = e.Tick
			}
			mt.stateMu.Unlock()
			select {
			case mt.triggerQ <- triggerJob{order: o, matched: e, matchedType: e.Type, matchedTick: e.Tick, confirm: o.Confirm}:
			default:
				log.Printf("metatron: trigger queue full, order %s dropped", o.ID)
				mt.stateMu.Lock()
				delete(mt.pendingTrigger, o.ID)
				if o.Confirm {
					// No confirm was enqueued, so it must not count against the rate cap.
					delete(mt.lastConfirmTick, o.ID)
				}
				mt.stateMu.Unlock()
			}
			break // one job per order per batch
		}
	}
}

// triggerWorker consumes the trigger queue FIFO (spec 029 US3/US6). A structural
// (non-fuzzy) job fires straight through runTrigger; a fuzzy job first runs the
// watch confirm (runConfirm), which fires only on a yes verdict. One worker →
// triggered turns and confirms serialize with each other and, via the shared
// turnBusy, with console turns (R6).
func (mt *Metatron) triggerWorker() {
	for {
		select {
		case <-mt.done:
			return
		case job := <-mt.triggerQ:
			if job.confirm {
				mt.runConfirm(job)
			} else {
				mt.runTrigger(job)
			}
		}
	}
}

// confirmCallTimeout bounds the single fuzzy-order watch confirm.
const confirmCallTimeout = 30 * time.Second

// confirmSystem is the fixed system prompt for the fuzzy-order watch confirm (spec
// 029 US6/R9, contracts/routing.md): a bounded yes/no judge — no tools, no
// preamble, one word out.
const confirmSystem = `You are Metatron's watchful eye. A standing order watches for a condition ` +
	`that cannot be checked mechanically; a candidate event has just occurred. Decide whether the ` +
	`event TRULY satisfies the condition. Answer with a single word — "yes" if it does, "no" if it ` +
	`does not. Say nothing else.`

// runConfirm runs a fuzzy order's watch confirm (spec 029 US6/R9). A yes verdict
// proceeds to the normal trigger pipeline (runTrigger, which clears the in-flight
// marker); a no, garbage, or FAILED verdict leaves the order armed with NO retry
// (the confirm is a single bare call, not a loop) — only the in-flight marker is
// cleared here so a later hit can confirm again, subject to the rate cap.
func (mt *Metatron) runConfirm(job triggerJob) {
	confirmed, err := mt.confirmOrder(job)
	if err != nil || !confirmed {
		mt.stateMu.Lock()
		delete(mt.pendingTrigger, job.order.ID)
		mt.stateMu.Unlock()
		return
	}
	mt.runTrigger(job)
}

// confirmOrder issues the ONE bounded watch-confirm Submit (spec 029 US6, R9 /
// contracts/routing.md): KindMetatronWatch, MaxTokens 16, the fixed system prompt,
// and a user prompt of the order's condition + the matched event in the digest
// vocabulary. Reply contract: a single leading yes/no token (case-insensitive);
// anything else, empty, or an error is a NO (unconfirmed). Budget/tier/transport
// failures return (false, err) — unconfirmed, never retried (no loop to retry).
func (mt *Metatron) confirmOrder(job triggerJob) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), confirmCallTimeout)
	defer cancel()
	resp, err := mt.orch.Submit(ctx, llm.Request{
		Kind:      llm.KindMetatronWatch,
		System:    confirmSystem,
		Prompt:    confirmPrompt(job.order, job.matched),
		MaxTokens: 16,
	})
	if err != nil {
		log.Printf("metatron: watch confirm for %s unconfirmed: %v", job.order.ID, err)
		return false, err
	}
	return confirmYes(resp.Text), nil
}

// confirmYes applies the yes/no reply contract (spec 029 R9): the FIRST token,
// case-insensitively and stripped of punctuation, must be exactly "yes"; anything
// else — "no", prose, or empty — is a NO.
func confirmYes(text string) bool {
	fields := strings.Fields(strings.ToLower(text))
	if len(fields) == 0 {
		return false
	}
	tok := strings.TrimFunc(fields[0], func(r rune) bool { return !unicode.IsLetter(r) })
	return tok == "yes"
}

// confirmPrompt renders the watch-confirm user prompt (spec 029 R9): the order's
// original condition plus the matched event described in the same digest vocabulary
// the soul uses, so the watch model judges the phrasing the angel would recognize.
func confirmPrompt(o sim.MetatronOrder, e store.Event) string {
	return fmt.Sprintf("The condition you watch for:\n%s\n\nWhat just happened:\n%s\n\nDoes it satisfy the condition?",
		o.Condition, describeEvent(e))
}

// describeEvent renders an observable event in the digest vocabulary for the watch
// confirm (spec 029 R9). It reads the STATIC sim.AgentNames rather than the replica
// the absorb goroutine owns (the confirm runs on the trigger worker), and falls
// back to the bare type for an unrecognized shape — never a panic on odd payloads.
func describeEvent(e store.Event) string {
	name := func(i int) string {
		if i >= 0 && i < len(sim.AgentNames) {
			return sim.AgentNames[i]
		}
		return "someone"
	}
	var m map[string]json.RawMessage
	_ = json.Unmarshal(e.Payload, &m)
	agentName := func(field string) string {
		if raw, ok := m[field]; ok {
			var v int
			if json.Unmarshal(raw, &v) == nil {
				return name(v)
			}
		}
		return ""
	}
	switch e.Type {
	case "agent.slept":
		if a := agentName("agent"); a != "" {
			return a + " fell asleep"
		}
	case "agent.woke":
		if a := agentName("agent"); a != "" {
			return a + " woke"
		}
	case "agent.died":
		var p sim.DiedPayload
		if json.Unmarshal(e.Payload, &p) == nil {
			return fmt.Sprintf("%s died of %s", name(p.Agent), p.Cause)
		}
	case "agent.memory_added":
		var p sim.MemoryAddedPayload
		if json.Unmarshal(e.Payload, &p) == nil {
			return fmt.Sprintf("%s gained a memory: %q", name(p.Agent), p.Text)
		}
	case "agent.intent_set":
		if a := agentName("agent"); a != "" {
			return a + " set out on something new"
		}
	case "social.conversation":
		var p sim.ConversationPayload
		if json.Unmarshal(e.Payload, &p) == nil && p.Gist != "" {
			return fmt.Sprintf("villagers talked: %q", p.Gist)
		}
	case "social.promise_broken":
		return "a promise was broken"
	case "social.rumor_told":
		var p sim.RumorToldPayload
		if json.Unmarshal(e.Payload, &p) == nil {
			return fmt.Sprintf("%s told %s a rumor: %q", name(p.From), name(p.To), p.Text)
		}
	case "gru.attacked":
		var p sim.GruAttackedPayload
		if json.Unmarshal(e.Payload, &p) == nil {
			return fmt.Sprintf("the gru attacked %s", name(p.Agent))
		}
	case "norm.violated":
		if a := agentName("agent"); a != "" {
			return a + " broke a village norm"
		}
		return "a village norm was broken"
	case "sim.night_started":
		return "night fell"
	case "sim.day_started":
		return "day broke"
	}
	return "an event occurred (" + e.Type + ")"
}

// runTrigger fires one matched standing order (spec 029 US3, T013/T014):
//
//  1. Land metatron.order_triggered through the door — the dry-run enforces the
//     order is STILL active, so a cancel/expiry that raced the match wins here and
//     the trigger is abandoned silently (edge case: exactly one of triggered/
//     cancelled/expired lands, never both).
//  2. Empty-bank precheck for known-act (deferral) orders (T014/R12): a
//     system-origin order's act is a known charge spend, so an empty bank
//     short-circuits to an honest moment — no model call, no cloud cost.
//  3. Acquire turnBusy with a bounded wait (system turns wait; the console stays
//     fail-fast) and run the pre-authorized action as a system-authored turn.
//  4. Queue a moment from the outcome — the act on success, or ONE model-free
//     honest moment per failure family, never a retry (FR-011).
func (mt *Metatron) runTrigger(job triggerJob) {
	defer func() {
		mt.stateMu.Lock()
		delete(mt.pendingTrigger, job.order.ID)
		mt.stateMu.Unlock()
	}()

	// A survival watch (spec 059 US2) never consumes: it is the angel's standing
	// nature, not a one-shot order, so it lands NO order_triggered (that would
	// deactivate it) and runs its own survival-authority turn.
	if job.order.Survival != "" {
		mt.runSurvivalTrigger(job)
		return
	}

	trig := []store.Event{{Type: "metatron.order_triggered", Payload: mustJSON(sim.OrderTriggeredPayload{
		ID: job.order.ID, MatchedType: job.matchedType, MatchedTick: job.matchedTick})}}
	if err := mt.social.InjectSocial(trig); err != nil {
		// Cancelled or expired before its trigger landed: the loser abandons
		// silently (the winning terminal already surfaced its own trail).
		log.Printf("metatron: order %s trigger abandoned at the door: %v", job.order.ID, err)
		return
	}
	mt.appendFile(mt.soulPath(), fmt.Sprintf("\n- %s — a watch woke me (%s): %q\n",
		clock.Format(job.matchedTick), job.order.ID, job.order.Condition))

	if mt.knownActEmptyBank(job.order) {
		mt.queueMoment(fmt.Sprintf("%s — a watch came due, but my strength was spent; I could not send what you asked.",
			clock.Format(job.matchedTick)))
		return
	}

	if !mt.acquireTurnBusy() {
		mt.queueMoment(fmt.Sprintf("%s — a watch came due, but I was too long attending another matter to act on it.",
			clock.Format(job.matchedTick)))
		return
	}
	res, err := mt.runTurn(context.Background(), turnOrigin{system: true, jobPrefix: "watch", seed: job.order.Action})
	mt.turnBusy.Store(false)

	if err != nil {
		mt.queueMoment(mt.degradedMoment(job.matchedTick, err))
		return
	}
	mt.queueMoment(mt.triggeredMoment(job.matchedTick, job.order, res))
}

// runSurvivalTrigger runs a survival watch's turn (spec 059 US2). It differs from
// runTrigger in three ways that follow from a survival watch's nature:
//
//  1. NO order_triggered — the watch is non-expiring and non-consuming, so it
//     stays active for the world's whole life (the authority trail is the soul
//     line + transcript + moment, all attributed to the survival duty, FR-007).
//  2. NO empty-bank short-circuit — the turn ALWAYS runs (edge case: "the turn
//     still happens … it must not burn the match silently — record the helpless
//     turn"). At zero charges every acting tool refuses in-fiction and the angel
//     narrates helplessly; the transcript + a helpless moment are the record.
//  3. The frame carries the survival-authority carve-out (turnOrigin.survival),
//     and the seed names the endangered villager + peril so the angel can aim.
func (mt *Metatron) runSurvivalTrigger(job triggerJob) {
	agent := survivalMatchedAgent(job.matched)
	name := "a villager"
	if agent >= 0 && agent < len(sim.AgentNames) {
		name = sim.AgentNames[agent]
	}
	peril := survivalPeril(job.order.Survival)
	mt.appendFile(mt.soulPath(), fmt.Sprintf("\n- %s — the survival watch woke me: %s is %s\n",
		clock.Format(job.matchedTick), name, peril))

	if !mt.acquireTurnBusy() {
		mt.queueMoment(fmt.Sprintf("%s — the survival watch woke (%s is %s), but I was too long attending another matter to act.",
			clock.Format(job.matchedTick), name, peril))
		return
	}
	seed := fmt.Sprintf("The survival watch has woken you: %s is %s and may die. %s",
		name, peril, job.order.Action)
	res, err := mt.runTurn(context.Background(), turnOrigin{system: true, survival: true, jobPrefix: "watch", seed: seed})
	mt.turnBusy.Store(false)

	if err != nil {
		mt.queueMoment(mt.degradedMoment(job.matchedTick, err))
		return
	}
	mt.queueMoment(mt.survivalMoment(job.matchedTick, name, peril, res))
}

// survivalPeril renders a survival watch kind as the villager's plight, for the
// soul/moment trail (spec 059 US2 attribution).
func survivalPeril(kind string) string {
	switch kind {
	case sim.SurvivalNearDeath:
		return "near death"
	case sim.SurvivalStarvation:
		return "starving"
	case sim.SurvivalExposure:
		return "freezing"
	}
	return "in mortal danger"
}

// survivalMatchedAgent decodes the endangered villager index from the survival
// watch's matched agent.needs_changed event (-1 if unreadable).
func survivalMatchedAgent(e store.Event) int {
	var p sim.NeedsPayload
	if json.Unmarshal(e.Payload, &p) != nil {
		return -1
	}
	return p.Agent
}

// survivalMoment renders the model-free moment describing a completed survival
// turn (spec 059 US2/FR-007): it names the life-saving act when one landed, else
// records that the watch woke and the angel could do nothing — the "helpless turn"
// the spec insists is recorded, never silent. Every branch attributes the moment
// to the survival watch.
func (mt *Metatron) survivalMoment(tick int64, name, peril string, r TurnResult) string {
	switch {
	case r.Nudge != nil:
		return fmt.Sprintf("%s — the survival watch woke (%s is %s): I sent a %s to %s.",
			clock.Format(tick), name, peril, r.Nudge.Form, strings.Join(r.Nudge.Targets, ", "))
	case r.Miracle != nil:
		return fmt.Sprintf("%s — the survival watch woke (%s is %s): I worked a miracle — %s.",
			clock.Format(tick), name, peril, r.Miracle.Summary)
	default:
		return fmt.Sprintf("%s — the survival watch woke (%s is %s), but I could do nothing to save them.",
			clock.Format(tick), name, peril)
	}
}

// knownActEmptyBank reports whether an order's action is a KNOWN charge-spending
// act (a deferral order — origin "system", always omen/vision-bearing per R11) AND
// the charge bank is empty at trigger time (T014/R12). Only then is the empty-bank
// precheck honest: a free-form player monitor order's action may be advisory or a
// meta act, so it still runs the turn. (Batch B places no system-origin orders —
// deferral is Batch C T016 — so this is dormant-but-correct, guarding exactly the
// deferral orders it will meet.)
func (mt *Metatron) knownActEmptyBank(o sim.MetatronOrder) bool {
	if o.Origin != "system" {
		return false
	}
	mt.stateMu.Lock()
	c := mt.charges
	mt.stateMu.Unlock()
	return c <= 0
}

// acquireTurnBusy waits (bounded by systemTurnBusyWait) for the single-flight turn
// slot for a SYSTEM turn (spec 029 R6). Returns false if the slot could not be
// acquired in time (the caller degrades to an honest moment) or the angel is
// closing. The console path never calls this — it CAS-fails fast with ErrTurnBusy.
func (mt *Metatron) acquireTurnBusy() bool {
	deadline := time.Now().Add(systemTurnBusyWait)
	tick := time.NewTicker(5 * time.Millisecond)
	defer tick.Stop()
	for {
		if mt.turnBusy.CompareAndSwap(false, true) {
			return true
		}
		select {
		case <-mt.done:
			return false
		case <-tick.C:
		}
		if time.Now().After(deadline) {
			return mt.turnBusy.CompareAndSwap(false, true)
		}
	}
}

// queueMoment appends a model-free moment to the soul and the player-facing queue
// (the same discipline as observeMoment's drama moments) — how a triggered turn's
// outcome reaches the next console reply (spec 029 US3, SC-003).
func (mt *Metatron) queueMoment(line string) {
	if line == "" {
		return
	}
	mt.appendFile(mt.soulPath(), "\n**MOMENT** "+line+"\n")
	mt.stateMu.Lock()
	mt.moments = append(mt.moments, line)
	mt.stateMu.Unlock()
}

// triggeredMoment renders the model-free moment describing a COMPLETED triggered
// turn (spec 029 US3): what the angel did while the player was away, so the next
// console reply leads with it. It names the landed act (omen/vision/miracle) when
// one landed, else that the watch woke and was attended.
func (mt *Metatron) triggeredMoment(tick int64, o sim.MetatronOrder, r TurnResult) string {
	switch {
	case r.Nudge != nil:
		return fmt.Sprintf("%s — a watch came due (%q): I sent a %s to %s.",
			clock.Format(tick), o.Condition, r.Nudge.Form, strings.Join(r.Nudge.Targets, ", "))
	case r.Miracle != nil:
		return fmt.Sprintf("%s — a watch came due (%q): I worked a miracle — %s.",
			clock.Format(tick), o.Condition, r.Miracle.Summary)
	default:
		return fmt.Sprintf("%s — a watch came due (%q): I attended it.", clock.Format(tick), o.Condition)
	}
}

// degradedMoment maps a FAILED system turn to ONE model-free honest moment per
// failure family (spec 029 T014/R12, FR-011) — never a retry. Empty bank is caught
// earlier (knownActEmptyBank); this covers exhausted budget, a downed/busy tier,
// and transport failures the loop already retried once internally.
func (mt *Metatron) degradedMoment(tick int64, err error) string {
	switch {
	case errors.Is(err, llm.ErrBudgetExhausted), errors.Is(err, llm.ErrTierDown), errors.Is(err, llm.ErrTierBusy):
		return fmt.Sprintf("%s — a watch came due, but my sight dimmed and I could not act. Nothing was spent.", clock.Format(tick))
	default:
		return fmt.Sprintf("%s — a watch came due, but I faltered and could not complete it. Nothing was spent.", clock.Format(tick))
	}
}

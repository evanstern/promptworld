package tui

// First-occurrence lessons projection (spec 055, TASK-117, reorientation
// decision 5/D2/D8): a purely client-side teaching layer over the same event
// stream decisionTraces already folds (research.md R3, the decisions.go
// ingest precedent) — no daemon changes, no new event types, no model
// calls. A static, skin-tokened catalog (contracts/lessons-catalog.md) fires
// each entry's lesson at most once per player (per-user seen-state,
// internal/worlds/lessons.go), surfaced one at a time in the lesson row
// (views.go) with a bounded queue and opportunity decay (data-model.md).

import (
	"strings"
	"time"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
)

// --- T003: catalog + skin-token resolution seam ---

// Lesson tiers (contracts/lessons-catalog.md): mechanics teaches the UI,
// prompting teaches the player's own prompting practice (the reorientation's
// teaching core).
const (
	lessonTierMechanics = "mechanics"
	lessonTierPrompting = "prompting"
)

// lessonEntry is one static catalog record (data-model.md "lessonEntry"):
// id/title/body feed the help overlay's pull half (helpLessons, help.go);
// text/pointer feed the push half (the row, views.go). Trigger/Done are pure
// predicates over the same store.Event stream applyEvent already folds
// (tui.go) — Done is nil for every entry without an unambiguous clearing
// event (research.md R4: "v1 ships them only where an unambiguous event
// exists"), meaning `x` is that entry's only dismissal path.
type lessonEntry struct {
	ID      string
	Title   string
	Body    string
	Text    string
	Pointer string
	Tier    string
	Trigger func(store.Event) bool
	Done    func(store.Event) bool
}

// lessonPullSuffix is the pull-path suffix every lesson string carries
// (FR-001, contracts/lessons-catalog.md invariant 4) — appended by the
// renderer (views.go), never stored per-entry, so there is exactly one
// source for it.
const lessonPullSuffix = "(? for more · x dismiss)"

// lessonCatalog is the minimum taxonomy (contracts/lessons-catalog.md, 8
// entries: 5 mechanics + 3 prompting) — the single source both the row
// (push) and the help overlay's lessons section (pull, helpLessons) read
// from. Append-only at runtime; every string is authored with skin tokens
// (lessonSkinResolve) for every guardian reference (FR-008).
var lessonCatalog = []lessonEntry{
	{
		ID:    "first-suppression",
		Title: "Suppression",
		Body: "At high speed the {{skin.guardian.epithet}} skips a villager's turn " +
			"rather than fall behind — nothing is broken, and no thought is lost, " +
			"only deferred.",
		Text:    "At high speed the {{skin.guardian.epithet}} just skipped a thought rather than fall behind.",
		Pointer: "→ press [ to slow down",
		Tier:    lessonTierMechanics,
		Trigger: func(e store.Event) bool {
			if e.Type != "cog.outcome" {
				return false
			}
			p, ok := decode[sim.CogOutcomePayload](e)
			return ok && p.Outcome == sim.OutcomeSuppressed
		},
	},
	{
		ID:    "first-gru-attack",
		Title: "The gru",
		Body: "The gru is a real predator stalking the map — an attack is a genuine " +
			"threat to whoever it catches, not a random flavor event.",
		Text:    "The gru just attacked a villager — a real threat, not a random event.",
		Pointer: "→ press 2 for the chronicle",
		Tier:    lessonTierMechanics,
		Trigger: func(e store.Event) bool { return e.Type == "gru.attacked" },
	},
	{
		ID:    "first-charge-regen",
		Title: "The action budget",
		Body: "The {{skin.guardian.epithet}}'s action budget refills on its own over " +
			"time — you don't have to conserve it forever, just pace bigger asks " +
			"against what's currently banked.",
		Text:    "The {{skin.guardian.epithet}}'s action budget just refilled by one — it recovers on its own.",
		Pointer: "→ press 3 for the {{skin.guardian.tab_label}} tab to watch the bank",
		Tier:    lessonTierMechanics,
		Trigger: func(e store.Event) bool { return e.Type == "metatron.charge_regenerated" },
	},
	{
		ID:    "first-order-expired",
		Title: "Standing orders expire",
		Body: "A standing order only watches for a limited time; once it expires " +
			"unmet, the {{skin.guardian.epithet}} stops watching for it — place a " +
			"new one if the condition still matters to you.",
		Text:    "A standing order you placed just expired unmet — orders watch for a limited time.",
		Pointer: "→ press 3, then look for 👁 rows to place another",
		Tier:    lessonTierMechanics,
		Trigger: func(e store.Event) bool { return e.Type == "metatron.order_expired" },
		Done:    func(e store.Event) bool { return e.Type == "metatron.order_placed" },
	},
	{
		ID:    "first-death",
		Title: "Death is real",
		Body: "A villager just died — the world carries real stakes; food, warmth, " +
			"and the gru are not merely decorative pressures.",
		Text:    "A villager has died — the world is not without real stakes.",
		Pointer: "→ press 2 for the chronicle",
		Tier:    lessonTierMechanics,
		Trigger: func(e store.Event) bool { return e.Type == "agent.died" },
	},
	{
		ID:    "first-rejected-tool-call",
		Title: "Refusals are informative",
		Body: "The {{skin.guardian.epithet}}'s own attempt to act was just refused — " +
			"its grant is real, gated by the same rules every turn runs under, and a " +
			"refusal is information about what's allowed, not a broken feature.",
		Text:    "The {{skin.guardian.epithet}}'s own attempt to act was just refused — its grant is real.",
		Pointer: "→ press 4, then d for the decision trace",
		Tier:    lessonTierPrompting,
		Trigger: func(e store.Event) bool {
			if e.Type != "cog.tool_call" {
				return false
			}
			p, ok := decode[sim.CogToolCallPayload](e)
			return ok && p.Verdict != "landed"
		},
	},
	{
		ID:    "first-custom-charter",
		Title: "Your charter, your voice",
		Body: "Editing charter.md and returning changes what the {{skin.guardian.epithet}} " +
			"is — its next turn ran under your own authored words, not the default voice.",
		Text:    "You edited charter.md — the {{skin.guardian.epithet}} is now acting under your own words.",
		Pointer: "→ press 3 for the {{skin.guardian.tab_label}} tab",
		Tier:    lessonTierPrompting,
		Trigger: func(e store.Event) bool {
			if e.Type != "metatron.charter_observed" {
				return false
			}
			p, ok := decode[sim.CharterObservedPayload](e)
			return ok && !p.Default
		},
	},
	{
		ID:    "first-fuzzy-order",
		Title: "Fuzzy orders still bind",
		Body: "A vaguely worded standing order still binds — the {{skin.guardian.epithet}} " +
			"marks it fuzzy rather than pretending your wording was precise, and still " +
			"watches for it in good faith.",
		Text:    "A vaguely worded standing order still binds — marked fuzzy, honestly, not silently ignored.",
		Pointer: "→ press 3, look for the ~ mark beside your order",
		Tier:    lessonTierPrompting,
		Trigger: func(e store.Event) bool {
			if e.Type != "metatron.order_placed" {
				return false
			}
			p, ok := decode[sim.MetatronOrder](e)
			return ok && p.Confirm
		},
	},
}

// lessonSkinTokens is the bounded fallback skin-token table (research.md R1):
// TASK-121 (spec 052) has not yet merged its runtime token-resolution
// substrate to main as of this feature's implementation — only the minimal
// StageIdentity table exists (internal/skin/skin.go), not the full §2
// lookup. Rather than block on 121 or duplicate its in-flight work, this
// table resolves ONLY the tokens lessonCatalog actually uses, with values
// from the PUBLISHED contract's §3 default-skin table
// (specs/052-skinnable-guardian/contracts/skin-contract.md). Swap
// lessonSkinResolve's body to delegate to 121's real resolver (a
// single-function change) once it merges — do not grow this table for
// tokens no lesson string uses.
var lessonSkinTokens = map[string]string{
	"{{skin.guardian.epithet}}":   "guardian",
	"{{skin.guardian.tab_label}}": "guardian",
}

// lessonSkinResolve resolves every skin token in s to its (currently
// default-only) value — FR-008/SC-005: no rendered lesson string may ever
// contain a raw "{{" literal. See lessonSkinTokens' doc comment for the
// TASK-121 swap-over plan.
func lessonSkinResolve(s string) string {
	for token, value := range lessonSkinTokens {
		s = strings.ReplaceAll(s, token, value)
	}
	return s
}

// --- T012/US3: the help overlay's pull half (one catalog, two surfaces) ---

// populateHelpLessons fills the help overlay's lessons-section seam
// (helpLessons, help.go) from lessonCatalog, 1:1, id-for-id — FR-007's
// "never two hand-maintained lists" made structural (SC-002). Called once
// per client boot (New(), tui.go) rather than a package init(): once
// TASK-121's skin runtime lands, a world's skin.json will make this
// resolution per-world, not a single process-wide constant, so this must
// re-run at each client's construction (research.md R7's "boot-frozen ⇒
// boot-time resolution is stable" — the boot in question is THIS client's,
// not the process's). Bodies are skin-resolved at population time, same as
// the row; the overlay itself (helpLessonsLines) does no further resolution.
func populateHelpLessons() {
	entries := make([]helpLesson, len(lessonCatalog))
	for i, e := range lessonCatalog {
		entries[i] = helpLesson{ID: e.ID, Title: lessonSkinResolve(e.Title), Body: lessonSkinResolve(e.Body)}
	}
	helpLessons = entries
}

// --- T004/T008/US1-US2: trigger projection + one-active/dwell/queue/decay ---

// lessonQueueDecay/lessonSpacing are the anti-spam constants (research.md
// R4): "constants-over-config" per the spec's own assumption (no config
// surface in v1) — the contract is the ordering/one-active/decay behavior,
// not these specific durations. Wall-clock (not sim tick) because the anti-
// spam pacing is about real player attention span, independent of the
// world's simulated speed (1x-32x) — a queued opportunity at 32x should
// decay on the same real-world clock it would at 1x.
const (
	// lessonQueueDecay is how long a queued (not-yet-shown) trigger waits
	// before its opportunity is dropped rather than surfaced stale
	// (panels/lesson-row.md "Opportunity decay").
	lessonQueueDecay = 90 * time.Second
	// lessonSpacing is the gap after a lesson clears before the next queued
	// (or freshly triggered) lesson is allowed to surface (data-model.md
	// "spacing elapsed" transition) — so lessons never feel like a firehose.
	lessonSpacing = 5 * time.Second
)

// activeLesson is the row's current occupant (data-model.md "lessonRow"):
// shownAt is provenance only (an active lesson has NO timeout — it dwells
// until its done-signal or `x`, panels/lesson-row.md).
type activeLesson struct {
	entry   *lessonEntry
	shownAt time.Time
}

// queuedLesson is one pending first-occurrence awaiting the row, carrying
// its own decay deadline (opportunity decay, FR-004).
type queuedLesson struct {
	entry   *lessonEntry
	decayAt time.Time
}

// lessonTriggers is the projection (data-model.md "lessonRow"): seen is the
// in-memory suppression set (the loaded per-user record ∪ ids marked this
// session — checked before enqueue/activate, contract invariant 2); active/
// queue/clearedAt implement the one-active/dwell/queue/decay/spacing state
// machine. Zero value is usable (ensureSeen), matching decisionTraces'
// nil-map-tolerant construction pattern.
type lessonTriggers struct {
	seen      map[string]bool
	active    *activeLesson
	queue     []queuedLesson
	clearedAt time.Time // zero value: no lesson has ever cleared — spacing never blocks the very first one
}

// newLessonTriggers builds the projection from the previously-seen id set —
// takes a plain map rather than *worlds.LessonsSeen directly so this package
// (and its tests) don't need to import internal/worlds just to build a
// fixture; tui.go's real call site passes worlds.LoadLessonsSeen().Entries's
// keys (New(), tui.go).
func newLessonTriggers(seenIDs map[string]bool) lessonTriggers {
	seen := make(map[string]bool, len(seenIDs))
	for id := range seenIDs {
		seen[id] = true
	}
	return lessonTriggers{seen: seen}
}

func (lt *lessonTriggers) ensureSeen() {
	if lt.seen == nil {
		lt.seen = map[string]bool{}
	}
}

// ActiveEntry returns the currently-showing lesson, or nil (row absent/badge
// — views.go's "none"/"badge" rendering states).
func (lt lessonTriggers) ActiveEntry() *lessonEntry {
	if lt.active == nil {
		return nil
	}
	return lt.active.entry
}

// isPendingOrActive reports whether id is already active or already queued —
// a re-arriving trigger for the same not-yet-seen lesson (e.g. a second
// gru.attacked before the first surfaces) must not duplicate its queue slot.
func (lt lessonTriggers) isPendingOrActive(id string) bool {
	if lt.active != nil && lt.active.entry.ID == id {
		return true
	}
	for _, q := range lt.queue {
		if q.entry.ID == id {
			return true
		}
	}
	return false
}

// decayQueue drops every queued entry whose decay deadline has passed
// (opportunity decay, FR-004) — silently: a decayed entry was never shown,
// so it is NOT recorded seen (data-model.md: "a decayed queue entry has not
// been seen and may fire on a later first occurrence").
func (lt *lessonTriggers) decayQueue(now time.Time) {
	if len(lt.queue) == 0 {
		return
	}
	kept := make([]queuedLesson, 0, len(lt.queue))
	for _, q := range lt.queue {
		if q.decayAt.After(now) {
			kept = append(kept, q)
		}
	}
	lt.queue = kept
}

// activateNow makes entry the active lesson and marks it seen — marking
// happens here, at SURFACE time, never at enqueue (data-model.md "Marking
// seen happens when a lesson SURFACES"). The caller (ingest/Advance) is
// responsible for persisting the mark via worlds.MarkLessonSeen (tui.go
// T005) — this in-memory mark alone prevents a same-session re-trigger from
// double-queuing regardless of whether the persistent write succeeds
// (advisory-never-authority, FR-006).
func (lt *lessonTriggers) activateNow(entry *lessonEntry, now time.Time) *lessonEntry {
	lt.ensureSeen()
	lt.active = &activeLesson{entry: entry, shownAt: now}
	lt.seen[entry.ID] = true
	return entry
}

// tryPromote activates entry immediately if the row is free and the spacing
// gap since the last clear has elapsed; returns nil (does nothing) otherwise
// — the caller enqueues on a nil return.
func (lt *lessonTriggers) tryPromote(entry *lessonEntry, now time.Time) *lessonEntry {
	if lt.active != nil {
		return nil
	}
	if now.Sub(lt.clearedAt) < lessonSpacing {
		return nil
	}
	return lt.activateNow(entry, now)
}

// enqueue records entry as a pending opportunity with its own decay deadline.
func (lt *lessonTriggers) enqueue(entry *lessonEntry, now time.Time) {
	lt.queue = append(lt.queue, queuedLesson{entry: entry, decayAt: now.Add(lessonQueueDecay)})
}

// ingest folds one subscribed event into the projection (contract, mirrors
// decisionTraces.ingest's calling convention, decisions.go:150): first, an
// active lesson's own done-signal (if any) clears it — checked before the
// trigger match below because a single event (e.g. metatron.order_placed)
// can BOTH clear one lesson's dwell (first-order-expired's done-signal) AND
// separately trigger another (first-fuzzy-order) in the same event. Returns
// the entry that just surfaced (became active), or nil if nothing did — the
// caller (tui.go applyEvent) persists a non-nil return via
// worlds.MarkLessonSeen.
func (lt *lessonTriggers) ingest(e store.Event, now time.Time) *lessonEntry {
	lt.ensureSeen()
	if lt.active != nil && lt.active.entry.Done != nil && lt.active.entry.Done(e) {
		lt.active = nil
		lt.clearedAt = now
	}
	for i := range lessonCatalog {
		entry := &lessonCatalog[i]
		if lt.seen[entry.ID] || lt.isPendingOrActive(entry.ID) {
			continue
		}
		if !entry.Trigger(e) {
			continue
		}
		if surfaced := lt.tryPromote(entry, now); surfaced != nil {
			return surfaced
		}
		lt.enqueue(entry, now)
		return nil
	}
	return nil
}

// Advance is the time-driven half of the state machine (data-model.md
// "spacing elapsed ... head becomes active"): called on every poll tick
// (tui.go pollMsg, ~1s) so the queue can progress and decay even when no new
// event happens to arrive — e.g. the spacing gap elapsing with a
// still-fresh queue head. Returns the entry that surfaced, if any (same
// persistence contract as ingest).
func (lt *lessonTriggers) Advance(now time.Time) *lessonEntry {
	lt.ensureSeen()
	lt.decayQueue(now)
	if lt.active != nil || len(lt.queue) == 0 {
		return nil
	}
	if now.Sub(lt.clearedAt) < lessonSpacing {
		return nil
	}
	head := lt.queue[0]
	lt.queue = lt.queue[1:]
	return lt.activateNow(head.entry, now)
}

// Dismiss is the `x` key (FR-010): clears an active lesson outright,
// returning true; a strict no-op (false, no state change) when nothing is
// active — the documented fallthrough (patterns/keymap.md).
func (lt *lessonTriggers) Dismiss(now time.Time) bool {
	if lt.active == nil {
		return false
	}
	lt.active = nil
	lt.clearedAt = now
	return true
}

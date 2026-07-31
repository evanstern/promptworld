package sim

// Spec 107 suites (FR-006, sim half): mission arm validation tables, the
// one-way terminal races, link discipline, prune determinism, the derived
// completion/failure sweep (satisfied + stalled, once-only, completed-wins),
// forgery refusal at the social door, and from-genesis replay byte-identity
// over a full mission lifecycle log — the plans_test.go shapes verbatim.

import (
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/store"
)

// validMission returns a legal mission_accepted payload (7-day TTL).
func validMission(id string, tick int64) Mission {
	return Mission{ID: id, Goal: "A shelter stands at (10,10).",
		AcceptedTick: tick, DeadlineTick: tick + 7*ticksPerGameDay}
}

func acceptedEvent(m Mission, seq, tick int64) store.Event {
	return store.Event{Seq: seq, Tick: tick, Type: "guardian.mission_accepted", Payload: mustPayload(m)}
}

func progressEvent(p MissionProgressedPayload, tick int64) store.Event {
	return store.Event{Tick: tick, Type: "guardian.mission_progressed", Payload: mustPayload(p)}
}

// TestMissionAcceptedLandsActive: payload Status/PlacedSeq are IGNORED — the
// reducer lands active, stamps PlacedSeq from the store seq, and refuses
// pre-linked entities (links land only through mission_progressed).
func TestMissionAcceptedLandsActive(t *testing.T) {
	s := planState(t, 107)
	m := validMission("msn-10-0", 10)
	m.Status = "completed" // ignored
	m.PlacedSeq = 999      // ignored
	if err := s.Apply(acceptedEvent(m, 42, 10)); err != nil {
		t.Fatalf("accept: %v", err)
	}
	got := s.missionByID("msn-10-0")
	if got == nil {
		t.Fatal("mission not on state")
	}
	if got.Status != "active" {
		t.Fatalf("status = %q, want active (payload Status ignored)", got.Status)
	}
	if got.PlacedSeq != 42 {
		t.Fatalf("PlacedSeq = %d, want the event seq 42", got.PlacedSeq)
	}

	pre := validMission("msn-10-1", 10)
	pre.Designations = []string{"dsg-1-0"}
	if err := s.Apply(acceptedEvent(pre, 43, 10)); err == nil {
		t.Fatal("acceptance carrying links must refuse")
	}
}

// TestMissionAcceptedValidationTable: the validate-not-clamp door.
func TestMissionAcceptedValidationTable(t *testing.T) {
	cases := []struct {
		name string
		mut  func(*Mission)
		want string
	}{
		{"empty id", func(m *Mission) { m.ID = "" }, "empty mission id"},
		{"empty goal", func(m *Mission) { m.Goal = "" }, "goal length"},
		{"goal over cap", func(m *Mission) { m.Goal = strings.Repeat("x", missionGoalMaxRunes+1) }, "goal length"},
		{"ttl under", func(m *Mission) { m.DeadlineTick = m.AcceptedTick + ticksPerGameDay/2 }, "ttl"},
		{"ttl over", func(m *Mission) { m.DeadlineTick = m.AcceptedTick + (GuardianMissionTTLMaxDays+1)*ticksPerGameDay }, "ttl"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := planState(t, 107)
			m := validMission("msn-10-0", 10)
			tc.mut(&m)
			err := s.Apply(acceptedEvent(m, 1, 10))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want contains %q", err, tc.want)
			}
		})
	}

	// Duplicate id in ANY status refuses (ids assigned once, retained).
	s := planState(t, 107)
	if err := s.Apply(acceptedEvent(validMission("msn-10-0", 10), 1, 10)); err != nil {
		t.Fatalf("first accept: %v", err)
	}
	if err := s.Apply(idEvent("guardian.mission_cancelled", "msn-10-0", 20)); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if err := s.Apply(acceptedEvent(validMission("msn-10-0", 30), 2, 30)); err == nil ||
		!strings.Contains(err.Error(), "duplicate mission id") {
		t.Fatalf("duplicate id err = %v", err)
	}
}

// TestMissionCap: the fourth concurrent active mission refuses; a terminal
// frees the slot.
func TestMissionCap(t *testing.T) {
	s := planState(t, 107)
	for i := 0; i < GuardianMissionCap; i++ {
		m := validMission("", 10)
		m.ID = "msn-10-" + string(rune('0'+i))
		if err := s.Apply(acceptedEvent(m, int64(i+1), 10)); err != nil {
			t.Fatalf("accept %d: %v", i, err)
		}
	}
	over := validMission("msn-10-9", 10)
	if err := s.Apply(acceptedEvent(over, 9, 10)); err == nil ||
		!strings.Contains(err.Error(), "cap") {
		t.Fatalf("over-cap err = %v", err)
	}
	if err := s.Apply(idEvent("guardian.mission_cancelled", "msn-10-0", 20)); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if err := s.Apply(acceptedEvent(over, 10, 20)); err != nil {
		t.Fatalf("accept after free slot: %v", err)
	}
}

// TestMissionProgressLinkDiscipline: links must name existing entities, never
// duplicate, and empty progress refuses; a note-only step lands.
func TestMissionProgressLinkDiscipline(t *testing.T) {
	s := planState(t, 107)
	if err := s.Apply(acceptedEvent(validMission("msn-10-0", 10), 1, 10)); err != nil {
		t.Fatalf("accept: %v", err)
	}
	if err := s.Apply(placedEvent(validSite("dsg-20-0", 20), 2, 20)); err != nil {
		t.Fatalf("place: %v", err)
	}

	// Unknown mission / unknown designation / unknown directive refuse.
	if err := s.Apply(progressEvent(MissionProgressedPayload{ID: "msn-nope", DesignationID: "dsg-20-0"}, 30)); err == nil {
		t.Fatal("unknown mission must refuse")
	}
	if err := s.Apply(progressEvent(MissionProgressedPayload{ID: "msn-10-0", DesignationID: "dsg-nope"}, 30)); err == nil {
		t.Fatal("unknown designation must refuse")
	}
	if err := s.Apply(progressEvent(MissionProgressedPayload{ID: "msn-10-0", DirectiveID: "dir-nope"}, 30)); err == nil {
		t.Fatal("unknown directive must refuse")
	}
	// Empty progress refuses.
	if err := s.Apply(progressEvent(MissionProgressedPayload{ID: "msn-10-0"}, 30)); err == nil {
		t.Fatal("empty progress must refuse")
	}

	// A real link lands and accumulates; the duplicate refuses.
	if err := s.Apply(progressEvent(MissionProgressedPayload{ID: "msn-10-0", DesignationID: "dsg-20-0"}, 30)); err != nil {
		t.Fatalf("link: %v", err)
	}
	m := s.missionByID("msn-10-0")
	if len(m.Designations) != 1 || m.Designations[0] != "dsg-20-0" {
		t.Fatalf("Designations = %v", m.Designations)
	}
	if err := s.Apply(progressEvent(MissionProgressedPayload{ID: "msn-10-0", DesignationID: "dsg-20-0"}, 31)); err == nil ||
		!strings.Contains(err.Error(), "already linked") {
		t.Fatalf("duplicate link err = %v", err)
	}
	// A note-only step lands (the recorded event is the note's durable home).
	if err := s.Apply(progressEvent(MissionProgressedPayload{ID: "msn-10-0", Note: "surveyed the west bank"}, 32)); err != nil {
		t.Fatalf("note-only: %v", err)
	}
	// Progress against a non-active mission refuses.
	if err := s.Apply(idEvent("guardian.mission_cancelled", "msn-10-0", 40)); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if err := s.Apply(progressEvent(MissionProgressedPayload{ID: "msn-10-0", Note: "late"}, 41)); err == nil ||
		!strings.Contains(err.Error(), "not active") {
		t.Fatalf("post-terminal progress err = %v", err)
	}
}

// TestMissionOneWayTerminals: exactly one terminal lands; every later
// transition refuses on the non-active entity (the transitionDesignation
// race shape), including the executor-emitted terminals' re-validation.
func TestMissionOneWayTerminals(t *testing.T) {
	s := planState(t, 107)
	if err := s.Apply(acceptedEvent(validMission("msn-10-0", 10), 1, 10)); err != nil {
		t.Fatalf("accept: %v", err)
	}
	if err := s.Apply(idEvent("guardian.mission_cancelled", "msn-10-0", 20)); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if err := s.Apply(idEvent("guardian.mission_cancelled", "msn-10-0", 21)); err == nil {
		t.Fatal("second cancel must refuse")
	}
	failed := store.Event{Tick: 10 + 8*ticksPerGameDay, Type: "guardian.mission_failed",
		Payload: mustPayload(MissionFailedPayload{ID: "msn-10-0", Reason: MissionFailNeverPursued})}
	if err := s.Apply(failed); err == nil || !strings.Contains(err.Error(), "not active") {
		t.Fatalf("failed-after-cancel err = %v", err)
	}
}

// TestMissionTerminalRevalidation: the executor-emitted terminals re-validate
// their emitting condition at the arm — a forged mission_completed with the
// predicate unmet, or a mission_failed before the deadline (or with the
// predicate met), refuses even though the entity is active.
func TestMissionTerminalRevalidation(t *testing.T) {
	s := planState(t, 107)
	if err := s.Apply(acceptedEvent(validMission("msn-10-0", 10), 1, 10)); err != nil {
		t.Fatalf("accept: %v", err)
	}
	done := store.Event{Tick: 20, Type: "guardian.mission_completed",
		Payload: mustPayload(MissionCompletedPayload{ID: "msn-10-0"})}
	if err := s.Apply(done); err == nil || !strings.Contains(err.Error(), "predicate does not hold") {
		t.Fatalf("unmet-predicate complete err = %v", err)
	}
	early := store.Event{Tick: 20, Type: "guardian.mission_failed",
		Payload: mustPayload(MissionFailedPayload{ID: "msn-10-0", Reason: MissionFailDeadline})}
	if err := s.Apply(early); err == nil || !strings.Contains(err.Error(), "not past its deadline") {
		t.Fatalf("early fail err = %v", err)
	}

	// Predicate met ⇒ completed lands; a deadline-passed fail then refuses on
	// the predicate (completed-wins is arm-enforced, not just sweep-ordered).
	if err := s.Apply(placedEvent(validSite("dsg-20-0", 20), 2, 20)); err != nil {
		t.Fatalf("place: %v", err)
	}
	if err := s.Apply(progressEvent(MissionProgressedPayload{ID: "msn-10-0", DesignationID: "dsg-20-0"}, 21)); err != nil {
		t.Fatalf("link: %v", err)
	}
	if err := s.Apply(store.Event{Tick: 30, Type: "agent.built",
		Payload: mustPayload(BuiltPayload{Agent: Ref(0), Kind: "shelter", X: 10, Y: 10})}); err != nil {
		t.Fatalf("build: %v", err)
	}
	if err := s.Apply(idEvent("designation.fulfilled", "dsg-20-0", 31)); err != nil {
		t.Fatalf("designation fulfil: %v", err)
	}
	if err := s.Apply(store.Event{Tick: 32, Type: "guardian.mission_completed",
		Payload: mustPayload(MissionCompletedPayload{ID: "msn-10-0", Designations: []string{"dsg-20-0"}})}); err != nil {
		t.Fatalf("complete: %v", err)
	}
	if got := s.missionByID("msn-10-0").Status; got != "completed" {
		t.Fatalf("status = %q, want completed", got)
	}
}

// TestMissionSweepCompletesAndFails drives the executor sweep end to end on
// two missions: one whose linked designation fulfills (completed, with the
// evidence trail), one that stalls to its deadline (failed, with the frozen
// reason and linked-entity statuses) — each terminal exactly once.
func TestMissionSweepCompletesAndFails(t *testing.T) {
	const seed = 107
	const ticks = 8*ticksPerGameDay + 200
	m := testMap(seed)

	pursued := validMission("msn-50-0", 50)
	stalled := Mission{ID: "msn-60-0", Goal: "A wall stands from (12,10) to (14,10).",
		AcceptedTick: 60, DeadlineTick: 60 + 7*ticksPerGameDay}
	timeline := map[int64][]store.Event{
		50: {acceptedEvent(pursued, 0, 50)},
		60: {acceptedEvent(stalled, 0, 60)},
		70: {placedEvent(validSite("dsg-70-0", 70), 0, 70)},
		80: {progressEvent(MissionProgressedPayload{ID: "msn-50-0", DesignationID: "dsg-70-0"}, 80)},
		// The stalled mission links a wall-line designation no one ever builds.
		90:  {placedEvent(validLine("dsg-90-0", 90), 0, 90)},
		95:  {progressEvent(MissionProgressedPayload{ID: "msn-60-0", DesignationID: "dsg-90-0"}, 95)},
		100: {{Tick: 100, Type: "guardian.item_granted", Payload: mustPayload(ItemGrantedPayload{Agent: Ref(0), Kind: "planks", Qty: 4})}},
		110: {{Tick: 110, Type: "agent.built", Payload: mustPayload(BuiltPayload{Agent: Ref(0), Kind: "shelter", X: 10, Y: 10})}},
	}

	live := NewState(seed, m)
	log := driveTicks(t, live, m, ticks, timeline)

	if n := countType(log, "guardian.mission_completed"); n != 1 {
		t.Fatalf("log carries %d mission_completed, want exactly 1", n)
	}
	if n := countType(log, "guardian.mission_failed"); n != 1 {
		t.Fatalf("log carries %d mission_failed, want exactly 1", n)
	}
	if got := live.missionByID("msn-50-0").Status; got != "completed" {
		t.Fatalf("pursued mission status = %q, want completed", got)
	}
	if got := live.missionByID("msn-60-0").Status; got != "failed" {
		t.Fatalf("stalled mission status = %q, want failed", got)
	}
	// The failure payload carries the frozen reason + linked-entity evidence.
	for _, e := range log {
		if e.Type != "guardian.mission_failed" {
			continue
		}
		var p MissionFailedPayload
		mustUnmarshal(t, e.Payload, &p)
		if p.ID != "msn-60-0" || p.Reason != MissionFailDeadline {
			t.Fatalf("failed payload = %+v", p)
		}
		if len(p.Designations) != 1 || p.Designations[0].ID != "dsg-90-0" || p.Designations[0].Status != "active" {
			t.Fatalf("failed evidence = %+v", p.Designations)
		}
	}
	// The completion payload cites the fulfilled linked designation.
	for _, e := range log {
		if e.Type != "guardian.mission_completed" {
			continue
		}
		var p MissionCompletedPayload
		mustUnmarshal(t, e.Payload, &p)
		if p.ID != "msn-50-0" || len(p.Designations) != 1 || p.Designations[0] != "dsg-70-0" {
			t.Fatalf("completed payload = %+v", p)
		}
	}
}

// TestMissionSweepNeverPursued: a mission with no linked designation fails at
// its deadline with the never_pursued reason — accepted intent alone
// completes nothing, and the failure says so honestly.
func TestMissionSweepNeverPursued(t *testing.T) {
	const seed = 107
	m := testMap(seed)
	timeline := map[int64][]store.Event{
		50: {acceptedEvent(Mission{ID: "msn-50-0", Goal: "A fire burns at (9,9).",
			AcceptedTick: 50, DeadlineTick: 50 + ticksPerGameDay}, 0, 50)},
	}
	live := NewState(seed, m)
	log := driveTicks(t, live, m, ticksPerGameDay+200, timeline)
	found := false
	for _, e := range log {
		if e.Type != "guardian.mission_failed" {
			continue
		}
		found = true
		var p MissionFailedPayload
		mustUnmarshal(t, e.Payload, &p)
		if p.Reason != MissionFailNeverPursued {
			t.Fatalf("reason = %q, want %q", p.Reason, MissionFailNeverPursued)
		}
	}
	if !found {
		t.Fatal("no mission_failed in log")
	}
}

// TestMissionForgeryRefusedAtSocialDoor: the derived terminals are NOT
// injectable — whitelist absence refuses a forged completion/failure at the
// InjectSocial door (D3: completion is derived, never self-graded).
func TestMissionForgeryRefusedAtSocialDoor(t *testing.T) {
	for _, typ := range []string{"guardian.mission_completed", "guardian.mission_failed"} {
		if injectSocialWhitelist[typ] {
			t.Errorf("%s is on the InjectSocial whitelist — a derived outcome must never be injectable", typ)
		}
	}
	for _, typ := range []string{"guardian.mission_accepted", "guardian.mission_progressed", "guardian.mission_cancelled"} {
		if !injectSocialWhitelist[typ] {
			t.Errorf("%s is missing from the InjectSocial whitelist", typ)
		}
	}
}

// TestMissionPruneDeterminism: the shared active+recent-32 retention prune
// applies to missions (bounded history, deterministic order).
func TestMissionPruneDeterminism(t *testing.T) {
	s := planState(t, 107)
	tick := int64(10)
	for i := 0; i < guardianOrderRetain+8; i++ {
		id := "msn-" + string(rune('a'+i%26)) + string(rune('a'+i/26))
		m := validMission(id, tick)
		if err := s.Apply(acceptedEvent(m, int64(i+1), tick)); err != nil {
			t.Fatalf("accept %d: %v", i, err)
		}
		if err := s.Apply(idEvent("guardian.mission_cancelled", id, tick+1)); err != nil {
			t.Fatalf("cancel %d: %v", i, err)
		}
		tick += 2
	}
	// The prune runs at append (acceptance) time — the plan-entity contract:
	// after one more acceptance, exactly retain non-active plus the new
	// active mission remain, oldest consumed entries dropped first.
	if err := s.Apply(acceptedEvent(validMission("msn-last-0", tick), 999, tick)); err != nil {
		t.Fatalf("final accept: %v", err)
	}
	nonActive := 0
	for i := range s.Missions {
		if s.Missions[i].Status != "active" {
			nonActive++
		}
	}
	if nonActive != guardianOrderRetain || len(s.Missions) != guardianOrderRetain+1 {
		t.Fatalf("retained %d non-active of %d total, want %d + 1 active",
			nonActive, len(s.Missions), guardianOrderRetain)
	}
}

// TestMissionLifecycleReplayByteIdentical: from-genesis replay over a full
// mission lifecycle log (accept → link → fulfil → complete; accept → stall →
// fail; accept → cancel) reproduces the live state byte-identically — the
// TestPlanLifecycleReplayByteIdentical shape (SC-004).
func TestMissionLifecycleReplayByteIdentical(t *testing.T) {
	const seed = 107
	const ticks = 8*ticksPerGameDay + 500
	m := testMap(seed)

	timeline := map[int64][]store.Event{
		50: {acceptedEvent(validMission("msn-50-0", 50), 0, 50)},
		60: {acceptedEvent(Mission{ID: "msn-60-0", Goal: "A wall from (12,10) to (14,10).",
			AcceptedTick: 60, DeadlineTick: 60 + 2*ticksPerGameDay}, 0, 60)},
		65: {acceptedEvent(Mission{ID: "msn-65-0", Goal: "Set aside.",
			AcceptedTick: 65, DeadlineTick: 65 + 7*ticksPerGameDay}, 0, 65)},
		70:  {placedEvent(validSite("dsg-70-0", 70), 0, 70)},
		80:  {progressEvent(MissionProgressedPayload{ID: "msn-50-0", DesignationID: "dsg-70-0", Note: "the site is marked"}, 80)},
		90:  {placedEvent(validLine("dsg-90-0", 90), 0, 90)},
		95:  {progressEvent(MissionProgressedPayload{ID: "msn-60-0", DesignationID: "dsg-90-0"}, 95)},
		100: {{Tick: 100, Type: "guardian.item_granted", Payload: mustPayload(ItemGrantedPayload{Agent: Ref(0), Kind: "planks", Qty: 4})}},
		110: {{Tick: 110, Type: "agent.built", Payload: mustPayload(BuiltPayload{Agent: Ref(0), Kind: "shelter", X: 10, Y: 10})}},
		200: {idEvent("guardian.mission_cancelled", "msn-65-0", 200)},
	}

	live := NewState(seed, m)
	log := driveTicks(t, live, m, ticks, timeline)

	for _, want := range []string{"guardian.mission_completed", "guardian.mission_failed", "guardian.mission_cancelled"} {
		if n := countType(log, want); n != 1 {
			t.Fatalf("log carries %d %s, want exactly 1", n, want)
		}
	}

	replayed := NewState(seed, m)
	for _, e := range log {
		if err := replayed.Apply(e); err != nil {
			t.Fatalf("replay apply %s: %v", e.Type, err)
		}
		replayed.Tick = e.Tick
	}
	driveTicks(t, replayed, m, ticks, nil)

	if live.Hash() != replayed.Hash() {
		t.Fatalf("mission-lifecycle replay diverged from live:\nlive:     %s\nreplayed: %s",
			string(live.Marshal()), string(replayed.Marshal()))
	}
}

// TestPreSpec107SnapshotCompat: a snapshot with no missions field unmarshals
// to a nil slice and re-marshals byte-identically (omitempty — no format
// bump, pre-107 worlds untouched).
func TestPreSpec107SnapshotCompat(t *testing.T) {
	s := planState(t, 107)
	pre := s.Marshal()
	if strings.Contains(string(pre), "\"missions\"") {
		t.Fatal("empty mission set must omit the missions key (omitempty)")
	}
	s2 := NewState(107, testMap(107))
	mustUnmarshal(t, pre, s2)
	if got := string(s2.Marshal()); got != string(pre) {
		t.Fatalf("pre-107 snapshot did not round-trip byte-identically")
	}
}

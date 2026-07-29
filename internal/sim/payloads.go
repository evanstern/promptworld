package sim

// PayloadCatalog (spec 086 FR-006) is the sim-side payload registry: every
// event type this world can carry, mapped to a constructor for its zero
// payload value. It is the single enumerable truth three enforcement layers
// ride (data-model §5):
//
//   - TestPayloadAgentRefSweep reflects over every catalog type and fails
//     any int-kind field whose json tag is in the frozen agent vocabulary
//     but is not typed AgentRef (a future `Agent int` cannot land silently);
//   - the doc-anchored completeness test requires every backticked event
//     type in docs/wiki/event-types.md to be cataloged (the TestCatalogSweep
//     trick), so a new event type cannot exist outside the catalog;
//   - the InjectSocial door decodes each injected event's payload through
//     the catalog and validates its refs before the dry-run (agentref.go).
//
// Constructors return POINTERS to fresh zero values so callers can
// unmarshal directly into them. Types whose events carry an empty payload
// register *struct{}. The tui side welds its own catalogFixture to this map
// (catalogFixture keys ⊆ PayloadCatalog keys — internal/tui/digest_test.go).
var PayloadCatalog = map[string]func() any{
	// --- world / clock / daemon ---
	"world.created":            func() any { return &WorldCreatedPayload{} },
	"world.migrated":           func() any { return &WorldMigratedPayload{} },
	"world.forked":             func() any { return &WorldForkedPayload{} },
	"clock.paused":             func() any { return &struct{}{} },
	"clock.resumed":            func() any { return &struct{}{} },
	"clock.speed_set":          func() any { return &SpeedSetPayload{} },
	"clock.degraded":           func() any { return &DegradedPayload{} },
	"clock.recovered":          func() any { return &struct{}{} },
	"clock.governor_shed":      func() any { return &GovernorPayload{} },
	"clock.governor_recovered": func() any { return &GovernorPayload{} },
	"daemon.started":           func() any { return &DaemonStartedPayload{} },
	"daemon.stopped":           func() any { return &DaemonStoppedPayload{} },
	"daemon.llm_warning":       func() any { return &LLMWarningPayload{} },
	"run.ended":                func() any { return &RunEndedPayload{} },

	// --- sim environment ---
	"sim.day_started":        func() any { return &DayPayload{} },
	"sim.night_started":      func() any { return &DayPayload{} },
	"sim.forage_regrown":     func() any { return &RegrownPayload{} },
	"sim.fire_burned_out":    func() any { return &FireBurnedOutPayload{} },
	"sim.food_rotted":        func() any { return &FoodRottedPayload{} },
	"sim.gathering_observed": func() any { return &GatheringObservedPayload{} },
	"sim.cold_snap":          func() any { return &ColdSnapPayload{} },
	"sim.forage_blighted":    func() any { return &ForageBlightedPayload{} },
	"sim.neglect_detected":   func() any { return &NeglectDetectedPayload{} },
	"sim.tuning_applied":     func() any { return &TuningAppliedPayload{} },

	// --- agent: acts, needs, vitals ---
	"agent.intent_set":       func() any { return &IntentSetPayload{} },
	"agent.work_started":     func() any { return &WorkStartedPayload{} },
	"agent.intent_done":      func() any { return &AgentPayload{} },
	"agent.build_failed":     func() any { return &BuildFailedPayload{} },
	"agent.intent_failed":    func() any { return &IntentFailedPayload{} },
	"agent.intent_rejected":  func() any { return &IntentRejectedPayload{} },
	"agent.recovery_stalled": func() any { return &RecoveryStalledPayload{} },
	"agent.moved":            func() any { return &AgentMovedPayload{} },
	"agent.foraged":          func() any { return &HarvestPayload{} },
	"agent.chopped":          func() any { return &HarvestPayload{} },
	"agent.hunted":           func() any { return &HarvestPayload{} },
	"agent.quarried":         func() any { return &HarvestPayload{} },
	"agent.collected_water":  func() any { return &HarvestPayload{} },
	"agent.crafted":          func() any { return &CraftedPayload{} },
	"agent.built":            func() any { return &BuiltPayload{} },
	"agent.wall_chipped":     func() any { return &WallWorkPayload{} },
	"agent.wall_destroyed":   func() any { return &WallWorkPayload{} },
	"agent.wall_repaired":    func() any { return &WallWorkPayload{} },
	"agent.dropped":          func() any { return &DroppedPayload{} },
	"agent.picked_up":        func() any { return &PickedUpPayload{} },
	"agent.deposited":        func() any { return &DepositedPayload{} },
	"agent.withdrew":         func() any { return &WithdrewPayload{} },
	"agent.cooked":           func() any { return &CookedPayload{} },
	"agent.bathed":           func() any { return &BathedPayload{} },
	"agent.refueled":         func() any { return &RefueledPayload{} },
	"agent.spear_broke":      func() any { return &SpearBrokePayload{} },
	"agent.axe_broke":        func() any { return &AxeBrokePayload{} },
	"agent.ate":              func() any { return &AtePayload{} },
	"agent.slept":            func() any { return &AgentPayload{} },
	"agent.woke":             func() any { return &AgentPayload{} },
	"agent.needs_changed":    func() any { return &NeedsPayload{} },
	"agent.died":             func() any { return &DiedPayload{} },
	"agent.talked":           func() any { return &TalkedPayload{} },

	// --- agent: memory, cognition, plans, mental map ---
	"agent.memory_added":       func() any { return &MemoryAddedPayload{} },
	"agent.memory_embedded":    func() any { return &MemoryEmbeddedPayload{} },
	"agent.situation_embedded": func() any { return &SituationEmbeddedPayload{} },
	"agent.thought":            func() any { return &ThoughtPayload{} },
	"agent.memory_promoted":    func() any { return &MemoryPromotedPayload{} },
	"agent.memory_faded":       func() any { return &MemoryFadedPayload{} },
	"agent.belief_revised":     func() any { return &BeliefRevisedPayload{} },
	"agent.belief_reinforced":  func() any { return &BeliefReinforcedPayload{} },
	"agent.narrative_set":      func() any { return &NarrativeSetPayload{} },
	"agent.consolidated":       func() any { return &ConsolidatedPayload{} },
	"agent.plan_set":           func() any { return &PlanSetPayload{} },
	"agent.plan_step_started":  func() any { return &PlanStepPayload{} },
	"agent.plan_expired":       func() any { return &PlanStepPayload{} },
	"agent.saw":                func() any { return &SawPayload{} },
	"agent.map_corrected":      func() any { return &MapCorrectedPayload{} },

	// --- social ---
	"social.place_told":        func() any { return &PlaceToldPayload{} },
	"social.conversation_turn": func() any { return &ConversationTurnPayload{} },
	"social.conversation":      func() any { return &ConversationPayload{} },
	"social.rumor_told":        func() any { return &RumorToldPayload{} },
	"social.relation_changed":  func() any { return &RelationChangedPayload{} },
	"social.gave":              func() any { return &GavePayload{} },
	"social.promise_broken":    func() any { return &PromiseBrokenPayload{} },
	"social.secret_seeded":     func() any { return &SecretSeededPayload{} },
	"social.chest_taken":       func() any { return &ChestTakenPayload{} },
	"social.hailed":            func() any { return &HailedPayload{} },
	"social.hail_met":          func() any { return &HailMetPayload{} },
	"social.hail_expired":      func() any { return &HailExpiredPayload{} },

	// --- journal (spec 059) ---
	"journal.entry_written": func() any { return &JournalWrittenPayload{} },
	"journal.entry_deleted": func() any { return &JournalDeletedPayload{} },

	// --- governance ---
	"meeting.convened":               func() any { return &MeetingPlacePayload{} },
	"meeting.opened":                 func() any { return &MeetingOpenedPayload{} },
	"meeting.turn_taken":             func() any { return &TurnTakenPayload{} },
	"meeting.proposal_tabled":        func() any { return &ProposalPayload{} },
	"meeting.proposal_resolved":      func() any { return &ProposalResolvedPayload{} },
	"meeting.proposal_rephrased":     func() any { return &ProposalRephrasedPayload{} },
	"meeting.closed":                 func() any { return &MeetingClosedPayload{} },
	"meeting.place_designated":       func() any { return &MeetingPlacePayload{} },
	"meeting.convention_established": func() any { return &MeetingConventionPayload{} },
	"norm.violated":                  func() any { return &NormViolatedPayload{} },

	// --- gru / strangers ---
	"gru.emerged":       func() any { return &GruEmergedPayload{} },
	"gru.moved":         func() any { return &GruMovedPayload{} },
	"gru.sighted":       func() any { return &GruSightedPayload{} },
	"gru.attacked":      func() any { return &GruAttackedPayload{} },
	"gru.withdrew":      func() any { return &GruWithdrewPayload{} },
	"stranger.arrived":  func() any { return &StrangerArrivedPayload{} },
	"stranger.moved":    func() any { return &StrangerMovedPayload{} },
	"stranger.took":     func() any { return &StrangerTookPayload{} },
	"stranger.departed": func() any { return &StrangerDepartedPayload{} },

	// --- chronicle / morgue ---
	"chronicle.entry": func() any { return &ChronicleEntryPayload{} },
	"morgue.epilogue": func() any { return &MorgueEpiloguePayload{} },

	// --- guardian ---
	// The 13 world-action types below were metatron.* until spec 094 renamed
	// them (log format 2, logformat.go). DOCTRINE: renaming ANY key in this
	// catalog is a log-format break — bump store.LogFormatVersion, extend
	// LogFormatV1Renames' successor table, and ship the translating
	// migration; never alias at read.
	"guardian.charge_regenerated": func() any { return &ChargeRegeneratedPayload{} },
	"guardian.nudged":             func() any { return &GuardianNudgedPayload{} },
	"guardian.place_revealed":     func() any { return &PlaceRevealedPayload{} },
	"guardian.order_placed":       func() any { return &OrderPlacedPayload{} },
	"guardian.order_triggered":    func() any { return &OrderTriggeredPayload{} },
	"guardian.order_cancelled":    func() any { return &OrderIDPayload{} },
	"guardian.order_expired":      func() any { return &OrderIDPayload{} },
	"guardian.charter_observed":   func() any { return &CharterObservedPayload{} },
	"guardian.skills_observed":    func() any { return &SkillsObservedPayload{} },
	"guardian.time_snapped":       func() any { return &TimeSnappedPayload{} },
	"guardian.item_granted":       func() any { return &ItemGrantedPayload{} },
	"guardian.entity_moved":       func() any { return &EntityMovedPayload{} },
	"guardian.entity_removed":     func() any { return &EntityRemovedPayload{} },
	"guardian.report_card":        func() any { return &GuardianReportCardPayload{} },

	// --- designations / directives / faith / prophecy (specs 078, 084, 085) ---
	"designation.placed":    func() any { return &Designation{} },
	"designation.cancelled": func() any { return &OrderIDPayload{} },
	"designation.fulfilled": func() any { return &OrderIDPayload{} },
	"directive.issued":      func() any { return &DirectiveIssuedPayload{} },
	"directive.cancelled":   func() any { return &OrderIDPayload{} },
	"directive.fulfilled":   func() any { return &DirectiveFulfilledPayload{} },
	"directive.expired":     func() any { return &OrderIDPayload{} },
	"faith.changed":         func() any { return &FaithChangedPayload{} },
	"prophecy.declared":     func() any { return &ProphecyDeclaredPayload{} },
	"prophecy.fulfilled":    func() any { return &OrderIDPayload{} },
	"prophecy.failed":       func() any { return &OrderIDPayload{} },

	// --- cognition telemetry ---
	"cog.thought":                   func() any { return &CogThoughtPayload{} },
	"cog.outcome":                   func() any { return &CogOutcomePayload{} },
	"cog.recalibration_recommended": func() any { return &RecalibrationPayload{} },
	"cog.tool_call":                 func() any { return &CogToolCallPayload{} },
	"cog.memory_divergence":         func() any { return &MemoryDivergencePayload{} },

	// --- curriculum (spec 046) ---
	"curriculum.exercise_passed": func() any { return &ExercisePassedPayload{} },
	"curriculum.stage_unlocked":  func() any { return &StageUnlockedPayload{} },
}

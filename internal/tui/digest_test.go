package tui

// Per-family digest unit tests (T013) and the catalog sweep test (T014,
// contracts/digest-grammar.md §7, SC-001). catalogFixture is the sweep's
// single source (research.md R3): one representative sample payload per
// cataloged event type, plus the plain-text summary contract §3's template
// renders for it — used both to assert per-type output and to gate
// registry coverage in both directions.

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
)

// digestFixture is one catalog sweep entry.
type digestFixture struct {
	payload string
	want    string // expected plainSegs(digest) output — contract §3's template made concrete
}

// catalogFixture: every registry key gets exactly one row here (both
// directions are asserted by TestCatalogSweep) — adding a digest without a
// fixture row, or a fixture row without a digest, fails the sweep.
var catalogFixture = map[string]digestFixture{
	// --- world / clock / daemon ---
	"world.created":  {`{"name":"Ashgrove","seed":42}`, `world "Ashgrove" created · seed 42`},
	"world.migrated": {`{"from_format":2,"source_events":100,"source_tick":500,"state":{}}`, `migrated from format v2 · 100 events @ tick 500`},
	"world.forked": {
		`{"parent_name":"aria","parent_seed":42,"parent_created_at":"2026-07-26T00:00:00Z","fork_tick":97200,"fork_seq":5000}`,
		`forked from "aria" at day 2, 09:00`,
	},
	"clock.paused":    {`{}`, `paused`},
	"clock.resumed":   {`{}`, `resumed`},
	"clock.speed_set": {`{"speed":"4x"}`, `speed=4x`},
	"clock.degraded":  {`{"effective_rate":3.5}`, `degraded rate=3.50`},
	"clock.recovered": {`{}`, `recovered`},
	"clock.governor_shed": {
		`{"requested":"32x","from":"32x","to":"16x","debt":1.4,"jobs":3}`,
		`governor shed 32x→16x debt=140% jobs=3`,
	},
	"clock.governor_recovered": {
		`{"requested":"32x","from":"8x","to":"16x","debt":0.3,"jobs":1}`,
		`governor recovered 8x→16x debt=30% jobs=1`,
	},
	"daemon.started": {`{"tick":100,"recovery_ms":250}`, `tick=100 recovery_ms=250`},
	"daemon.stopped": {`{"tick":100}`, `tick=100`},
	"daemon.llm_warning": {
		`{"provider":"local","kind":"model-missing","detail":"model not found","remedy":"ollama pull llama3","active":true}`,
		`provider=local kind=model-missing warning detail=model not found remedy=ollama pull llama3`,
	},
	"run.ended": {
		`{"tick":120000,"deaths":[{"agent":{"id":0,"name":"Ash"},"tick":90000,"cause":"starvation"},{"agent":{"id":1,"name":"Birch"},"tick":120000,"cause":"exposure"}],"final_cause":"exposure"}`,
		`the run ended · 2 dead · final cause exposure`,
	},

	// --- sim ---
	"sim.day_started":        {`{"day":3}`, `day 3 begins`},
	"sim.night_started":      {`{"day":3}`, `night falls on day 3`},
	"sim.forage_regrown":     {`{"x":2,"y":3}`, `forage regrew at (2,3)`},
	"sim.fire_burned_out":    {`{"x":4,"y":5}`, `the fire at (4,5) burned out`},
	"sim.food_rotted":        {`{"x":6,"y":6,"kind":"food_raw","n":4}`, `4 food_raw rotted at (6,6)`},
	"sim.gathering_observed": {`{"x":1,"y":1,"start":500}`, `gathering at (1,1) since tick 500`},
	// spec 077 US2: the two weather-shaped incident kinds (sim-family voice).
	"sim.cold_snap": {`{"night":1,"until_tick":90000}`, `a cold snap grips night 1 (until t90000)`},
	"sim.forage_blighted": {
		`{"x":10,"y":12,"radius":4,"tiles":[{"x":10,"y":12},{"x":11,"y":12}],"regrow_tick":460800}`,
		`blight struck the forage at (10,12) (+1 more tiles)`,
	},
	// spec 083: the death-by-neglect percept (alert tier — whole-line).
	"sim.neglect_detected": {
		`{"agent":{"id":2,"name":"Cedar"},"need":"warmth","level":0,"since":499320}`,
		`Cedar is dangerously cold and has done nothing about it (warmth 0)`,
	},

	// --- agent: acts & needs ---
	"agent.intent_set": {
		`{"agent":{"id":0,"name":"Ash"},"goal":"forage","target_x":3,"target_y":4,"res_x":0,"res_y":0,"source":"reflex"}`,
		`Ash intends forage (reflex) → (3,4)`,
	},
	"agent.work_started":     {`{"agent":{"id":1,"name":"Birch"},"tick":100}`, `Birch set to work`},
	"agent.intent_done":      {`{"agent":{"id":2,"name":"Cedar"}}`, `Cedar finished`},
	"agent.intent_rejected":  {`{"agent":{"id":3,"name":"Rowan"},"goal":"forage","reason":"blocked","staleness_ticks":5}`, `Rowan's forage refused: blocked (5t stale)`},
	"agent.build_failed":     {`{"agent":{"id":3,"name":"Rowan"},"goal":"build_wall_stone","reason":"site blocked too long"}`, `Rowan's build_wall_stone failed — site blocked too long`},
	"agent.recovery_stalled": {`{"agent":{"id":1,"name":"Birch"},"goal":"warm_up","need":"warmth"}`, `Birch's warm_up stalled — warmth not recovering`},
	"agent.moved":            {`{"agent":{"id":0,"name":"Ash"},"x":1,"y":1}`, `Ash → (1,1)`},
	"agent.saw": {
		`{"agent":{"id":0,"name":"Ash"},"facts":[{"kind":"fire","x":4,"y":5,"seen":100,"prov":"witnessed","detail":9000},{"kind":"tree","x":6,"y":5,"seen":100,"prov":"witnessed"}]}`,
		`Ash saw fire at (4,5) (+1 more)`,
	},
	"agent.map_corrected": {
		`{"agent":{"id":1,"name":"Birch"},"gone":[{"kind":"fire","x":4,"y":5,"seen":100,"prov":"witnessed","detail":9000}]}`,
		`Birch found fire at (4,5) gone`,
	},
	"social.place_told": {
		`{"from":{"id":0,"name":"Ash"},"to":{"id":1,"name":"Birch"},"facts":[{"kind":"fire","x":4,"y":5,"seen":100,"prov":"told","detail":9000},{"kind":"tree","x":6,"y":5,"seen":100,"prov":"told"}]}`,
		`Ash told Birch of fire at (4,5) (+1 more)`,
	},
	"agent.foraged":         {`{"agent":{"id":0,"name":"Ash"},"x":1,"y":1}`, `Ash foraged at (1,1)`},
	"agent.chopped":         {`{"agent":{"id":0,"name":"Ash"},"x":1,"y":1}`, `Ash chopped wood at (1,1)`},
	"agent.hunted":          {`{"agent":{"id":0,"name":"Ash"},"x":1,"y":1}`, `Ash hunted at (1,1)`},
	"agent.quarried":        {`{"agent":{"id":0,"name":"Ash"},"x":1,"y":1}`, `Ash quarried stone at (1,1)`},
	"agent.collected_water": {`{"agent":{"id":0,"name":"Ash"},"x":1,"y":1}`, `Ash drew water at (1,1)`},
	"agent.crafted":         {`{"agent":{"id":0,"name":"Ash"},"kind":"planks"}`, `Ash crafted planks`},
	"agent.built":           {`{"agent":{"id":0,"name":"Ash"},"kind":"fire","x":1,"y":1}`, `Ash built a fire at (1,1)`},
	"agent.wall_chipped":    {`{"agent":{"id":0,"name":"Ash"},"x":1,"y":1}`, `Ash chipped away at the wall at (1,1)`},
	"agent.wall_destroyed":  {`{"agent":{"id":0,"name":"Ash"},"x":1,"y":1}`, `Ash tore down the wall at (1,1)`},
	"agent.wall_repaired":   {`{"agent":{"id":0,"name":"Ash"},"x":1,"y":1}`, `Ash repaired the wall at (1,1)`},
	"agent.dropped":         {`{"agent":{"id":0,"name":"Ash"},"x":3,"y":4,"kind":"wood","n":2}`, `Ash dropped 2 wood at (3,4)`},
	"agent.picked_up":       {`{"agent":{"id":1,"name":"Birch"},"x":3,"y":4,"kind":"wood","n":2}`, `Birch picked up 2 wood at (3,4)`},
	"agent.deposited":       {`{"agent":{"id":2,"name":"Cedar"},"x":5,"y":5,"kind":"planks","n":6}`, `Cedar stored 6 planks in the chest at (5,5)`},
	"agent.withdrew":        {`{"agent":{"id":3,"name":"Rowan"},"x":5,"y":5,"kind":"planks","n":1,"owner":{"id":0,"name":"Ash"}}`, `Rowan took 1 planks from Ash's chest`},
	"agent.cooked":          {`{"agent":{"id":0,"name":"Ash"},"station":"fire","consumed":2,"produced":1,"kind":"food_cooked"}`, `Ash cooked 1 food_cooked at the fire`},
	"agent.bathed":          {`{"agent":{"id":0,"name":"Ash"},"morale_after":80,"warmth_after":90}`, `Ash bathed · morale 80 warmth 90`},
	"agent.refueled":        {`{"agent":{"id":0,"name":"Ash"},"x":1,"y":1,"fuel_until":500}`, `Ash refueled the fire at (1,1)`},
	"agent.spear_broke":     {`{"agent":{"id":0,"name":"Ash"}}`, `Ash's spear broke`},
	"agent.axe_broke":       {`{"agent":{"id":0,"name":"Ash"}}`, `Ash's axe broke`},
	"agent.ate":             {`{"agent":{"id":0,"name":"Ash"},"meals":1,"cooked":0,"raw":0,"food_after":80}`, `Ash ate 1 meals → food 80`},
	"agent.slept":           {`{"agent":{"id":0,"name":"Ash"}}`, `Ash fell asleep`},
	"agent.woke":            {`{"agent":{"id":0,"name":"Ash"}}`, `Ash woke`},
	"agent.needs_changed":   {`{"agent":{"id":0,"name":"Ash"},"health":90,"food":50,"rest":60,"warmth":70,"morale":80}`, `Ash health=90 food=50 rest=60 warmth=70 morale=80`},
	"agent.died":            {`{"agent":{"id":0,"name":"Ash"},"cause":"starvation"}`, `Ash died: starvation`},
	"agent.talked":          {`{"a":{"id":0,"name":"Ash"},"b":{"id":1,"name":"Birch"}}`, `Ash chatted with Birch`},

	// --- agent: mind & plans ---
	"agent.memory_added": {`{"agent":{"id":0,"name":"Ash"},"text":"the fire needs tending","salience":5,"subject":{"id":1,"name":"Birch"}}`, `Ash remembers: "the fire needs tending" · about Birch`},
	// Spec 042: embedding companions + divergence telemetry. Vectors are
	// elided by design; the digest carries the identity/audit fields.
	"agent.memory_embedded": {
		`{"agent":{"id":0,"name":"Ash"},"mem_seq":41,"vec":[0.1,0.2,0.3],"model":"all-minilm"}`,
		`Ash memory seq=41 embedded dims=3 model=all-minilm`,
	},
	"agent.situation_embedded": {
		`{"agent":{"id":1,"name":"Birch"},"tick":1801,"text":"daytime · at (3,4) · idle","vec":[0.1,0.2],"model":"all-minilm"}`,
		`Birch situation: "daytime · at (3,4) · idle" dims=2 model=all-minilm`,
	},
	"cog.memory_divergence": {
		`{"agent":{"id":2,"name":"Cedar"},"tick":1801,"mode":"shadow","legacy":[5,6],"augmented":[5,9],"overlap":1,"displacement":0,"vectorless":3,"sit_tick":1800}`,
		`agent=Cedar mode=shadow overlap=1/2 displaced=0 vectorless=3`,
	},
	"agent.thought":           {`{"agent":{"id":0,"name":"Ash"},"text":"I should forage","source":"planner"}`, `Ash thought: "I should forage" (planner)`},
	"agent.memory_promoted":   {`{"agent":{"id":0,"name":"Ash"},"mem_tick":100,"text_hash":"abc","boost":2}`, `Ash's memory (t100) reinforced`},
	"agent.memory_faded":      {`{"agent":{"id":0,"name":"Ash"},"mem_tick":100,"text_hash":"abc"}`, `Ash forgot a memory (t100)`},
	"agent.belief_revised":    {`{"agent":{"id":0,"name":"Ash"},"belief_id":0,"statement":"the fire needs tending","confidence":80,"provenance":"observed","source":{"id":0,"name":"Ash"},"subject":{"id":0,"name":"Ash"}}`, `Ash now believes: "the fire needs tending"`},
	"agent.belief_reinforced": {`{"agent":{"id":0,"name":"Ash"},"belief_id":0}`, `Ash's belief (#0) reinforced`},
	"agent.narrative_set":     {`{"agent":{"id":0,"name":"Ash"},"text":"a long night"}`, `Ash's story: "a long night"`},
	"agent.consolidated":      {`{"agent":{"id":0,"name":"Ash"},"night":1,"up_to":100,"outcome":"accepted"}`, `Ash consolidated the night's memories`},
	"agent.plan_set":          {`{"agent":{"id":0,"name":"Ash"},"job":"forage_run","steps":[{"job":"forage_run","goal":"forage","until":0},{"job":"forage_run","goal":"deposit","until":0}]}`, `Ash planned 2 steps: forage, deposit`},
	"agent.plan_step_started": {`{"agent":{"id":0,"name":"Ash"},"job":"forage_run","step":"forage"}`, `Ash began step forage`},
	"agent.plan_expired":      {`{"agent":{"id":0,"name":"Ash"},"job":"forage_run","step":"forage","reason":"window closed"}`, `Ash's plan lapsed (window closed)`},

	// --- social ---
	"social.conversation_turn": {`{"conv":100,"speaker":{"id":3,"name":"Rowan"},"listener":{"id":0,"name":"Ash"},"text":"hello"}`, `Rowan→Ash "hello"`},
	"social.rumor_told":        {`{"from":{"id":1,"name":"Birch"},"to":{"id":2,"name":"Cedar"},"rumor_id":0,"subject":{"id":0,"name":"Ash"},"tone":0,"text":"gossip","confidence":50}`, `Birch→Cedar rumor: "gossip"`},
	"social.conversation":      {`{"conv":1,"a":{"id":0,"name":"Ash"},"b":{"id":1,"name":"Birch"},"gist":"argued about firewood","turns":6}`, `"argued about firewood" · 6 turns`},
	"social.relation_changed":  {`{"a":{"id":0,"name":"Ash"},"b":{"id":1,"name":"Birch"},"trust_delta":2,"affection_delta":-1,"reason":"gift"}`, `Ash→Birch trust+2/affection-1 (gift)`},
	"social.gave":              {`{"from":{"id":0,"name":"Ash"},"to":{"id":1,"name":"Birch"},"kind":"food"}`, `Ash gave Birch food`},
	"social.promise_broken":    {`{"id":7}`, `a promise was broken (#7)`},
	"social.secret_seeded":     {`{"agent":{"id":0,"name":"Ash"},"text":"a secret","tone":0}`, `a secret took root with Ash`},
	"social.chest_taken":       {`{"owner":{"id":0,"name":"Ash"},"taker":{"id":3,"name":"Rowan"},"x":5,"y":5}`, `Rowan raided Ash's chest at (5,5)`},
	"social.hailed":            {`{"from":{"id":1,"name":"Birch"},"to":{"id":3,"name":"Rowan"},"until":12345}`, `Birch hailed Rowan (until t12345)`},
	"social.hail_met":          {`{"from":{"id":1,"name":"Birch"},"to":{"id":3,"name":"Rowan"}}`, `Birch met Rowan`},
	"social.hail_expired":      {`{"from":{"id":0,"name":"Ash"},"to":{"id":2,"name":"Cedar"}}`, `Ash's hail to Cedar lapsed`},

	// --- governance (meeting.* / norm.*) — all 9 meeting.* rows + norm.violated ---
	"meeting.convened":               {`{"x":1,"y":1}`, `meeting convened at (1,1)`},
	"meeting.opened":                 {`{"attendees":[{"id":0,"name":"Ash"},{"id":1,"name":"Birch"}]}`, `meeting opened`},
	"meeting.turn_taken":             {`{"agent":{"id":0,"name":"Ash"}}`, `Ash spoke at the meeting`},
	"meeting.proposal_tabled":        {`{"proposal_id":1,"kind":"amend","target":{"id":-1,"name":""},"proposer":{"id":0,"name":"Ash"},"text":"no stealing"}`, `Ash proposed: "no stealing"`},
	"meeting.proposal_resolved":      {`{"proposal_id":1,"kind":"amend","target":{"id":-1,"name":""},"proposer":{"id":0,"name":"Ash"},"text":"no stealing","yeas":[{"id":0,"name":"Ash"},{"id":1,"name":"Birch"}],"nays":[{"id":2,"name":"Cedar"}],"passed":true}`, `proposal passed: "no stealing" (2-1)`},
	"meeting.proposal_rephrased":     {`{"proposal_id":1,"norm_id":1,"text":"no stealing from chests"}`, `norm rephrased: "no stealing from chests"`},
	"meeting.closed":                 {`{"proposals":2}`, `meeting closed`},
	"meeting.place_designated":       {`{"x":2,"y":2}`, `meeting place set at (2,2)`},
	"meeting.convention_established": {`{"convene_second":72000,"open_second":75600,"x":2,"y":2,"source":"config"}`, `meeting convention: 21:00 at (2,2) (config)`},
	"norm.violated":                  {`{"norm_id":3,"violator":{"id":0,"name":"Ash"},"witnesses":[{"id":1,"name":"Birch"},{"id":2,"name":"Cedar"}]}`, `Ash violated a norm (#3)`},

	// --- gru / chronicle / guardian ---
	"gru.emerged":  {`{"night":1,"x":5,"y":5}`, `the gru emerged at (5,5)`},
	"gru.moved":    {`{"x":6,"y":6}`, `the gru prowls to (6,6)`},
	"gru.sighted":  {`{"agent":{"id":0,"name":"Ash"},"x":5,"y":5}`, `Ash sighted the gru`},
	"gru.attacked": {`{"agent":{"id":0,"name":"Ash"},"health":40}`, `the gru attacked Ash · health → 40`},
	"gru.withdrew": {`{"day":2}`, `the gru withdrew`},
	// spec 077 US2: the stranger — the gru-family threat voice; stranger.took
	// joins the whole-line alert tier beside social.chest_taken (theft is theft).
	"stranger.arrived":            {`{"night":2,"x":44,"y":0}`, `a stranger slipped in at (44,0)`},
	"stranger.moved":              {`{"x":6,"y":6}`, `the stranger creeps to (6,6)`},
	"stranger.took":               {`{"x":5,"y":5,"kind":"food_raw","n":2}`, `the stranger took 2 food_raw from the stores at (5,5)`},
	"stranger.departed":           {`{"day":2}`, `the stranger was gone by dawn of day 2`},
	"chronicle.entry":             {`{"day":3,"from_tick":100,"to_tick":200,"text":"Ash lit the first fire.","thread":"cold-start","agents":[{"id":0,"name":"Ash"}]}`, `day 3 · cold-start: Ash lit the first fire.`},
	"guardian.charge_regenerated": {`{}`, `a charge regenerated`},
	"guardian.nudged":             {`{"form":"dream","targets":[{"id":0,"name":"Ash"}],"text":"beware the cold"}`, `Guardian dream → Ash: "beware the cold"`},
	"guardian.place_revealed": {
		`{"agent":{"id":0,"name":"Ash"},"facts":[{"kind":"fire","x":4,"y":5,"seen":100,"prov":"revealed","detail":9000}]}`,
		`Guardian revealed fire at (4,5) to Ash`,
	},
	"guardian.order_placed": {
		`{"id":"ord-100-1","origin":"player","condition":"the woodpile drops below 5 logs","action":"nudge someone to chop wood","event_types":["sim.forage_regrown"],"agent":-1,"placed_tick":100,"expires_tick":100000,"status":"active"}`,
		`Guardian set a watch: "the woodpile drops below 5 logs"`,
	},
	"guardian.order_triggered": {`{"id":"ord-100-1","matched_type":"sim.forage_regrown","matched_tick":150}`, `Guardian's watch came true (sim.forage_regrown @ t150)`},
	"guardian.order_cancelled": {`{"id":"ord-100-1"}`, `Guardian released a watch (ord-100-1)`},
	"guardian.order_expired":   {`{"id":"ord-100-1"}`, `Guardian's watch lapsed (ord-100-1)`},

	// --- the plan layer (spec 084 — designations and directives) ---
	"designation.placed": {
		`{"id":"dsg-100-0","kind":"structure_site","x":4,"y":5,"x2":4,"y2":5,"structure_kind":"shelter","label":"north shelter","placed_tick":100,"status":"active"}`,
		`Guardian marked a structure_site at (4,5) (shelter) — "north shelter"`,
	},
	"designation.cancelled": {`{"id":"dsg-100-0"}`, `Guardian withdrew a designation (dsg-100-0)`},
	"designation.fulfilled": {`{"id":"dsg-100-0"}`, `the village fulfilled Guardian's mark (dsg-100-0)`},
	"directive.issued": {
		`{"id":"dir-200-0","designation_id":"dsg-100-0","targets":[{"id":0,"name":"Ash"},{"id":1,"name":"Birch"}],"text":"Raise the shelter I have marked.","issued_tick":200,"expires_tick":459200,"status":"active"}`,
		`Guardian charged Ash, Birch: "Raise the shelter I have marked."`,
	},
	"directive.cancelled": {`{"id":"dir-200-0"}`, `Guardian lifted a charge (dir-200-0)`},
	"directive.fulfilled": {
		`{"id":"dir-200-0","designation_id":"dsg-100-0","targets":[{"id":0,"name":"Ash"},{"id":1,"name":"Birch"}],"issued_tick":200}`,
		`the village fulfilled Guardian's charge (dir-200-0, serving dsg-100-0)`,
	},
	"directive.expired": {`{"id":"dir-200-0"}`, `Guardian's charge lapsed (dir-200-0)`},

	// --- the faith economy (spec 085 — faith movements and prophecies) ---
	"faith.changed": {
		`{"delta":8,"reason":"directive_fulfilled","source_id":"dir-200-0"}`,
		`the village's faith deepens (directive_fulfilled)`,
	},
	"prophecy.declared": {
		`{"id":"pro-300-0","targets":[{"id":0,"name":"Ash"},{"id":1,"name":"Birch"}],"text":"Before three dawns a shelter will stand.","claim":{"kind":"structure_count","structure_kind":"shelter","min":1},"declared_tick":300,"deadline_tick":559200,"status":"active"}`,
		`Guardian foretells: "Before three dawns a shelter will stand."`,
	},
	"prophecy.fulfilled": {`{"id":"pro-300-0"}`, `Guardian's foretelling came true (pro-300-0)`},
	"prophecy.failed":    {`{"id":"pro-300-0"}`, `Guardian's word did not come to pass (pro-300-0)`},
	"guardian.charter_observed": {
		`{"fingerprint":"ab12cd34ef56","default":false}`,
		`Guardian ran under charter ab12cd34ef56 (player-authored)`,
	},
	// spec 077 FR-006: the skills twin of the charter observation.
	"guardian.skills_observed": {
		`{"fingerprint":"ab12cd34ef56","names":["10-watch.md","20-tone.md"]}`,
		`Guardian ran under 2 skill files ab12cd34ef56`,
	},
	"morgue.epilogue": {
		`{"agent":{"id":0,"name":"Ash"},"text":"Ash kept the fire until the end."}`,
		`epilogue for Ash: Ash kept the fire until the end.`,
	},
	"guardian.time_snapped":   {`{"to_tick":106200,"gratis":false}`, `Guardian snapped time forward to day 2 11:30`},
	"guardian.item_granted":   {`{"agent":{"id":0,"name":"Ash"},"kind":"food_raw","qty":2,"gratis":false}`, `Guardian granted Ash 2 food_raw`},
	"guardian.entity_moved":   {`{"class":"pile","x":3,"y":4,"to_x":6,"to_y":7,"gratis":false}`, `Guardian moved the pile at (3,4) to (6,7)`},
	"guardian.entity_removed": {`{"class":"structure","x":12,"y":8,"gratis":false}`, `Guardian removed the structure at (12,8)`},

	// --- cog (labeled) ---
	"cog.thought": {
		`{"job":"j1","class":"reflex","agent":{"id":0,"name":"Ash"},"snapshot_tick":100,"generation":1,"trigger_seq":0,"points":5,"predicted_wall_ms":200,"predicted_land_tick":300}`,
		`job=j1 class=reflex agent=Ash pts=5 pred=200ms`,
	},
	"cog.outcome": {
		`{"job":"j1","class":"reflex","agent":{"id":0,"name":"Ash"},"outcome":"landed","snapshot_tick":100,"landing_tick":150,"staleness_ticks":10,"predicted_wall_ms":200,"actual_wall_ms":220}`,
		`job=j1 landed agent=Ash stale=10t wall=220ms`,
	},
	"cog.recalibration_recommended": {
		`{"tier":"cheap","estimate_s_per_pt":0.5,"spike_rate":0.2,"window":50}`,
		`tier=cheap est=0.50s/pt spikes=0.20 window=50`,
	},
	"cog.tool_call": {
		`{"job":"j1","ordinal":1,"tool":"inject_intent","args":{"agent":0},"verdict":"rejected_gate","reason":"stale snapshot","tier":"cheap","snapshot_tick":100}`,
		`job=j1 ord=1 tool=inject_intent rejected_gate tier=cheap reason=stale snapshot`,
	},

	// --- curriculum (spec 046 — the curriculum ladder) ---
	"curriculum.exercise_passed": {
		`{"exercise":"first-night","stage":"stage-1","tick":50000}`,
		`the first-night exercise was passed (stage-1)`,
	},
	"curriculum.stage_unlocked": {
		`{"stage":"stage-2","exercise":"first-night","tick":50000}`,
		`Guardian's watcher earned The Written Word (proven by first-night)`,
	},

	// --- guardian (spec 063 — the grounded feedback layer) ---
	"guardian.report_card": {
		`{"fingerprint":"a1b2c3d4e5f6","note":"Your charter never mentions coordinates; the working was rejected twice for them.","citations":[812,907]}`,
		`report card under charter a1b2c3d4e5f6: Your charter never mentions coordinates; the working was rejected twice for the…`,
	},
}

// TestCatalogSweep is the SC-001 gate (contract §7): every fixture type
// must have a registry entry and digest without a raw-JSON fallback; every
// registry key must appear in the fixture (no unlisted digests,
// data-model.md invariant 3); and every backticked concrete event type in
// docs/wiki/event-types.md must be covered by the fixture, so the doc and
// the digest catalog cannot drift silently (R3).
func TestCatalogSweep(t *testing.T) {
	names := []string{"Ash", "Birch", "Cedar", "Rowan"}

	for typ, fx := range catalogFixture {
		fn, ok := digestRegistry[typ]
		if !ok {
			t.Errorf("fixture type %q has no registry entry", typ)
			continue
		}
		e := store.Event{Seq: 1, Tick: 1, Type: typ, Payload: json.RawMessage(fx.payload)}
		segs, ok := fn(e, names, nil)
		if !ok {
			t.Errorf("%s: digest fell back (ok=false) on its own sample payload", typ)
			continue
		}
		if got := plainSegs(segs); got != fx.want {
			t.Errorf("%s: plain summary = %q, want %q", typ, got, fx.want)
		}
		// The AC #2 mechanical proof (spec 086 FR-007, US4 AS-1): an
		// agent-bearing payload (its refs carry names) digests IDENTICALLY
		// with names = nil — payload names suffice, no replica lookup.
		if strings.Contains(fx.payload, `"name":"`) {
			nilSegs, ok := fn(e, nil, nil)
			if !ok {
				t.Errorf("%s: digest fell back with names=nil", typ)
				continue
			}
			if got := plainSegs(nilSegs); got != fx.want {
				t.Errorf("%s: names=nil summary = %q, want %q — the digest still leans on the replica", typ, got, fx.want)
			}
		}
	}

	for typ := range digestRegistry {
		if _, ok := catalogFixture[typ]; !ok {
			t.Errorf("registry key %q has no catalog fixture row (unlisted digest)", typ)
		}
	}

	doc, err := os.ReadFile("../../docs/wiki/event-types.md")
	if err != nil {
		t.Fatalf("reading docs/wiki/event-types.md: %v", err)
	}
	for _, typ := range backtickedEventTypes(string(doc)) {
		if _, ok := catalogFixture[typ]; !ok {
			t.Errorf("docs/wiki/event-types.md backticks %q but the catalog fixture doesn't cover it", typ)
		}
	}

	// The tui↔sim catalog weld (spec 086 FR-006, T016): every fixture key
	// must be a sim.PayloadCatalog key — the sim-side registry is the single
	// enumerable truth, and a digest for an uncataloged type cannot exist.
	for typ := range catalogFixture {
		if _, ok := sim.PayloadCatalog[typ]; !ok {
			t.Errorf("catalogFixture key %q is not in sim.PayloadCatalog — the tui and sim catalogs drifted (spec 086 weld)", typ)
		}
	}
}

// TestExerciseRubricTermsAreCatalogedEventTypes (spec 046 T016, FR-010): every
// rubric term the shipped exercise definitions (internal/sim) name is a
// cataloged event type (a digestRegistry key — this package owns the catalog
// TestCatalogSweep enforces). Spec 044 US2's metatron.charter_observed,
// formerly a documented pending exception while task-31 was in flight, is now
// cataloged for real (its own digestRegistry/catalogFixture rows) and
// satisfies this check like every other term (T022 reconciliation).
func TestExerciseRubricTermsAreCatalogedEventTypes(t *testing.T) {
	for _, def := range sim.ScenarioExercises {
		for _, term := range def.RubricTerms {
			if _, ok := digestRegistry[term]; !ok {
				t.Errorf("%s: rubric term %q is not a cataloged event type", def.ID, term)
			}
		}
	}
}

// backtickedTypeRe matches a backticked `namespace.verb` token; the caller
// filters to known event-family namespaces so incidental matches like
// “ `social.go` “ or “ `world.json` “ (source-file/config references in
// prose) don't get mistaken for event types.
var backtickedTypeRe = regexp.MustCompile("`([a-z]+)\\.([a-z_]+)`")

func backtickedEventTypes(doc string) []string {
	var types []string
	seen := map[string]bool{}
	for _, m := range backtickedTypeRe.FindAllStringSubmatch(doc, -1) {
		ns, verb := m[1], m[2]
		if _, ok := familyByNamespace[ns]; !ok {
			continue // not one of our namespaces — e.g. a stray "foo.bar" in prose
		}
		if verb == "go" || verb == "json" || verb == "md" {
			continue // source-file / config references, not event types
		}
		t := ns + "." + verb
		if !seen[t] {
			seen[t] = true
			types = append(types, t)
		}
	}
	return types
}

// --- T013: per-family cases the flat fixture table doesn't reach —
// role-span assertions and conditional branches (subject absent, all-zero
// gathering, alert types). The fixture table above already exercises plain
// text for every type; these add the "and role spans" half of T013.

var digestTestNames = []string{"Ash", "Birch", "Cedar", "Rowan"}

func digestOf(t *testing.T, typ, payload string) []seg {
	t.Helper()
	fn, ok := digestRegistry[typ]
	if !ok {
		t.Fatalf("%s: no registry entry", typ)
	}
	e := store.Event{Seq: 1, Tick: 1, Type: typ, Payload: json.RawMessage(payload)}
	segs, ok := fn(e, digestTestNames, nil)
	if !ok {
		t.Fatalf("%s: digest returned ok=false", typ)
	}
	return segs
}

func TestDigestRoleSpans(t *testing.T) {
	// Speech privilege (contract §2): conversation_turn/rumor_told/scene/
	// meeting speech all carry a segSpeech span on the quoted text.
	for _, typ := range []string{"social.conversation_turn", "social.rumor_told", "social.conversation",
		"meeting.proposal_tabled", "meeting.proposal_resolved", "meeting.proposal_rephrased",
		"agent.thought", "agent.narrative_set", "agent.belief_revised", "agent.memory_added"} {
		fx := catalogFixture[typ]
		if !anyRole(digestOf(t, typ, fx.payload), segSpeech) {
			t.Errorf("%s: expected a segSpeech span", typ)
		}
	}

	// Every resolved agent name carries segName (contract §2 "name").
	for _, typ := range []string{"agent.moved", "agent.died", "agent.talked", "gru.sighted", "social.hailed"} {
		fx := catalogFixture[typ]
		if !anyRole(digestOf(t, typ, fx.payload), segName) {
			t.Errorf("%s: expected a segName span", typ)
		}
	}

	// Labeled voice (contract §2): cog/clock/daemon render key=value spans.
	for _, typ := range []string{"cog.thought", "cog.outcome", "cog.recalibration_recommended", "cog.tool_call",
		"clock.speed_set", "clock.degraded", "daemon.started", "daemon.stopped", "daemon.llm_warning",
		"agent.needs_changed"} {
		fx := catalogFixture[typ]
		if !anyRole(digestOf(t, typ, fx.payload), segLabel) {
			t.Errorf("%s: expected a segLabel span", typ)
		}
	}
}

// TestDigestMemoryAddedNoSubject: Subject's -1 sentinel (no gossip subject)
// must not append the "· about X" clause (internal/sim/memory.go).
func TestDigestMemoryAddedNoSubject(t *testing.T) {
	segs := digestOf(t, "agent.memory_added", `{"agent":0,"text":"the wind is picking up","salience":3,"subject":-1,"tone":0}`)
	if want := `Ash remembers: "the wind is picking up"`; plainSegs(segs) != want {
		t.Errorf("plain summary = %q, want %q", plainSegs(segs), want)
	}
}

// TestDigestGatheringDispersed: sim.gathering_observed's all-zero payload is
// the watch-reset sentinel, not a real gathering at (0,0) (contract §3).
func TestDigestGatheringDispersed(t *testing.T) {
	segs := digestOf(t, "sim.gathering_observed", `{"x":0,"y":0,"start":0}`)
	if want := "gathering dispersed"; plainSegs(segs) != want {
		t.Errorf("plain summary = %q, want %q", plainSegs(segs), want)
	}
}

// TestDigestLLMWarningCleared: an Active=false payload is the clear flavor —
// terse, no "warning"/detail/remedy clause (spec 034/038 provider-health
// preflight; the raise/reclassify flavor is covered by catalogFixture).
func TestDigestLLMWarningCleared(t *testing.T) {
	segs := digestOf(t, "daemon.llm_warning", `{"provider":"local","kind":"model-missing","active":false}`)
	if want := "provider=local kind=model-missing cleared"; plainSegs(segs) != want {
		t.Errorf("plain summary = %q, want %q", plainSegs(segs), want)
	}
}

// hasSeg reports whether segs contains a seg matching both text and role
// exactly — used where anyRole's role-only check is too loose (e.g.
// distinguishing the gratis marker's own segment from other segEmphasis
// spans already present in the summary).
func hasSeg(segs []seg, text string, role segRole) bool {
	for _, s := range segs {
		if s.Text == text && s.Role == role {
			return true
		}
	}
	return false
}

// TestDigestMiracleGratisMark: the four miracle types (TASK-59/spec 016)
// append a visible " (forced)" annotation when Gratis waives the charge, and
// nothing when it doesn't (spec 016 SC-004's enumerability story surfaced in
// the digest — gratisMark must never appear on a charge-priced miracle).
func TestDigestMiracleGratisMark(t *testing.T) {
	cases := []struct {
		typ                           string
		chargedPayload, gratisPayload string
		want                          string // base summary, sans the gratis suffix
	}{
		{
			"guardian.time_snapped",
			`{"to_tick":106200,"gratis":false}`, `{"to_tick":106200,"gratis":true}`,
			`Guardian snapped time forward to day 2 11:30`,
		},
		{
			"guardian.item_granted",
			`{"agent":0,"kind":"food_raw","qty":2,"gratis":false}`, `{"agent":0,"kind":"food_raw","qty":2,"gratis":true}`,
			`Guardian granted Ash 2 food_raw`,
		},
		{
			"guardian.entity_moved",
			`{"class":"pile","x":3,"y":4,"to_x":6,"to_y":7,"gratis":false}`, `{"class":"pile","x":3,"y":4,"to_x":6,"to_y":7,"gratis":true}`,
			`Guardian moved the pile at (3,4) to (6,7)`,
		},
		{
			"guardian.entity_removed",
			`{"class":"structure","x":12,"y":8,"gratis":false}`, `{"class":"structure","x":12,"y":8,"gratis":true}`,
			`Guardian removed the structure at (12,8)`,
		},
	}
	for _, tc := range cases {
		if got := plainSegs(digestOf(t, tc.typ, tc.chargedPayload)); got != tc.want {
			t.Errorf("%s (charged): plain summary = %q, want %q", tc.typ, got, tc.want)
		}
		gotGratis := plainSegs(digestOf(t, tc.typ, tc.gratisPayload))
		wantGratis := tc.want + " (forced)"
		if gotGratis != wantGratis {
			t.Errorf("%s (gratis): plain summary = %q, want %q", tc.typ, gotGratis, wantGratis)
		}
		if !hasSeg(digestOf(t, tc.typ, tc.gratisPayload), "forced", segEmphasis) {
			t.Errorf("%s (gratis): expected a styled %q segment", tc.typ, "forced")
		}
		if hasSeg(digestOf(t, tc.typ, tc.chargedPayload), "forced", segEmphasis) {
			t.Errorf("%s (charged): unexpected %q segment on a charge-priced miracle", tc.typ, "forced")
		}
	}
}

// TestDigestEntityMovedRemovedClasses: entity_moved/entity_removed render
// distinctly per Class (internal/sim/miracles.go) — villager/structure/pile
// for a move, structure/pile/terrain for a remove (terrain is overlaid, not
// deleted, so it reads "cleared" rather than "removed").
func TestDigestEntityMovedRemovedClasses(t *testing.T) {
	moveCases := []struct{ payload, want string }{
		{`{"class":"villager","x":1,"y":1,"to_x":2,"to_y":2,"gratis":false}`, `Guardian moved the villager at (1,1) to (2,2)`},
		{`{"class":"structure","x":1,"y":1,"to_x":2,"to_y":2,"gratis":false}`, `Guardian moved the structure at (1,1) to (2,2)`},
		{`{"class":"pile","x":1,"y":1,"to_x":2,"to_y":2,"gratis":false}`, `Guardian moved the pile at (1,1) to (2,2)`},
	}
	for _, tc := range moveCases {
		if got := plainSegs(digestOf(t, "guardian.entity_moved", tc.payload)); got != tc.want {
			t.Errorf("entity_moved %s: plain summary = %q, want %q", tc.payload, got, tc.want)
		}
	}

	removeCases := []struct{ payload, want string }{
		{`{"class":"structure","x":1,"y":1,"gratis":false}`, `Guardian removed the structure at (1,1)`},
		{`{"class":"pile","x":1,"y":1,"gratis":false}`, `Guardian removed the pile at (1,1)`},
		{`{"class":"terrain","x":1,"y":1,"gratis":false}`, `Guardian cleared the terrain at (1,1)`},
	}
	for _, tc := range removeCases {
		if got := plainSegs(digestOf(t, "guardian.entity_removed", tc.payload)); got != tc.want {
			t.Errorf("entity_removed %s: plain summary = %q, want %q", tc.payload, got, tc.want)
		}
	}
}

// TestDigestAlertTypesDigestCleanly: the four alert-flagged types (contract
// §2) still digest through the registry like any other type — the alert
// treatment is a view-layer style decision (Phase 5), not a formatting one.
func TestDigestAlertTypesDigestCleanly(t *testing.T) {
	for _, typ := range []string{"agent.died", "gru.attacked", "social.chest_taken", "norm.violated"} {
		fx, ok := catalogFixture[typ]
		if !ok {
			t.Fatalf("missing fixture for alert type %q", typ)
		}
		digestOf(t, typ, fx.payload) // fails the test via t.Fatalf if it falls back
	}
}

// --- spec 049 T004/T007: resolveSubject + the jump-or-hint totality sweep ---

// TestResolveSubjectActorAlivePrefersLivePosition: contract §2 step 1 — a
// living, resolvable actor's CURRENT position wins even when the event's own
// payload recorded a different (older) position.
func TestResolveSubjectActorAlivePrefersLivePosition(t *testing.T) {
	m := testModel(t)
	m.replica.Agents = []sim.Agent{{Name: "Ash", X: 9, Y: 9}}
	e := store.Event{Seq: 1, Tick: 1, Type: "agent.moved", Payload: json.RawMessage(`{"agent":0,"x":3,"y":4}`)}
	name, x, y, ok := m.resolveSubject(e)
	if !ok || name != "Ash" || x != 9 || y != 9 {
		t.Errorf("resolveSubject = (%q,%d,%d,%v), want (\"Ash\",9,9,true) — live position must win over the payload's recorded position", name, x, y, ok)
	}
}

// TestResolveSubjectActorDeadFallsToPayloadPosition: contract §2 step 2 /
// spec.md edge case — a dead actor's live (frozen) position is never used;
// the event's own recorded position (where it happened) wins instead.
func TestResolveSubjectActorDeadFallsToPayloadPosition(t *testing.T) {
	m := testModel(t)
	m.replica.Agents = []sim.Agent{{Name: "Ash", X: 9, Y: 9, Dead: true}}
	e := store.Event{Seq: 1, Tick: 1, Type: "agent.moved", Payload: json.RawMessage(`{"agent":0,"x":3,"y":4}`)}
	name, x, y, ok := m.resolveSubject(e)
	if !ok || name != "Ash" || x != 3 || y != 4 {
		t.Errorf("resolveSubject = (%q,%d,%d,%v), want (\"Ash\",3,4,true) — a dead actor must fall through to the recorded position, never a stale live one", name, x, y, ok)
	}
}

// TestResolveSubjectUnknownAgentIndexFallsToPayloadPosition: an agent index
// that doesn't resolve in the replica at all (post-migration mismatch, a
// world snapshot narrower than the event's own agent roster, etc.) is
// handled identically to a dead one — fall through to the payload position
// rather than erroring or panicking. testModel's replica seeds sim.AgentCount
// (8) agents (sim.NewState), so nil the roster out to force a genuine miss.
func TestResolveSubjectUnknownAgentIndexFallsToPayloadPosition(t *testing.T) {
	m := testModel(t)
	m.replica.Agents = nil
	e := store.Event{Seq: 1, Tick: 1, Type: "agent.foraged", Payload: json.RawMessage(`{"agent":7,"x":3,"y":4}`)}
	_, x, y, ok := m.resolveSubject(e)
	if !ok || x != 3 || y != 4 {
		t.Errorf("resolveSubject = (_,%d,%d,%v), want (3,4,true)", x, y, ok)
	}
}

// TestResolveSubjectUnlocatable: event types with no actor field and no
// recorded position at all (world.migrated included — R4/FR-011: never
// decoded, see TestResolveSubjectWorldMigratedNeverDecoded) resolve
// unlocatable, never panicking on the bare-minimum `{}` payload.
func TestResolveSubjectUnlocatable(t *testing.T) {
	m := testModel(t)
	for _, typ := range []string{"clock.paused", "world.created", "world.migrated", "social.promise_broken", "gru.withdrew"} {
		e := store.Event{Seq: 1, Tick: 1, Type: typ, Payload: json.RawMessage(`{}`)}
		if _, _, _, ok := m.resolveSubject(e); ok {
			t.Errorf("%s: expected unlocatable (no actor, no position field)", typ)
		}
	}
}

// TestResolveSubjectMultiParticipantSpeakerWins: spec.md "Edge Cases" —
// multi-agent events target the primary actor (speaker/initiator), never
// the other participant.
func TestResolveSubjectMultiParticipantSpeakerWins(t *testing.T) {
	m := testModel(t)
	m.replica.Agents = []sim.Agent{
		{Name: "Ash", X: 1, Y: 1},
		{Name: "Rowan", X: 7, Y: 7},
	}
	e := store.Event{Seq: 1, Tick: 1, Type: "social.conversation_turn", Payload: json.RawMessage(`{"conv":1,"speaker":1,"listener":0,"text":"hi"}`)}
	name, x, y, ok := m.resolveSubject(e)
	if !ok || name != "Rowan" || x != 7 || y != 7 {
		t.Errorf("resolveSubject = (%q,%d,%d,%v), want (\"Rowan\",7,7,true) — the speaker is the primary actor, not the listener", name, x, y, ok)
	}
}

// TestResolveSubjectGatheringDispersedSentinel: mirrors
// TestDigestGatheringDispersed — the all-zero payload is the watch-reset
// sentinel, not a real gathering at (0,0).
func TestResolveSubjectGatheringDispersedSentinel(t *testing.T) {
	m := testModel(t)
	e := store.Event{Seq: 1, Tick: 1, Type: "sim.gathering_observed", Payload: json.RawMessage(`{"x":0,"y":0,"start":0}`)}
	if _, _, _, ok := m.resolveSubject(e); ok {
		t.Error("the all-zero gathering payload is the watch-reset sentinel, not a real gathering at (0,0)")
	}
}

// TestResolveSubjectPlaceOnly: an event type with no actor field at all
// (a meeting place) resolves to a place label naming what's there.
func TestResolveSubjectPlaceOnly(t *testing.T) {
	m := testModel(t)
	e := store.Event{Seq: 1, Tick: 1, Type: "meeting.convened", Payload: json.RawMessage(`{"x":2,"y":3}`)}
	name, x, y, ok := m.resolveSubject(e)
	if !ok || name != "the meeting place" || x != 2 || y != 3 {
		t.Errorf("resolveSubject = (%q,%d,%d,%v), want (\"the meeting place\",2,3,true)", name, x, y, ok)
	}
}

// TestResolveSubjectWorldMigratedNeverDecoded: R4/FR-011 — world.migrated
// must never be decoded to decide locatability (it embeds the full
// sim.State). Proven two ways: the registry has no entry for it at all (so
// decode is never even reachable), and a payload too malformed to unmarshal
// still resolves unlocatable without error or panic.
func TestResolveSubjectWorldMigratedNeverDecoded(t *testing.T) {
	if _, ok := subjectRegistry["world.migrated"]; ok {
		t.Fatal("world.migrated must not be in subjectRegistry — resolveSubject must never decode its embedded sim.State")
	}
	m := testModel(t)
	e := store.Event{Seq: 1, Tick: 1, Type: "world.migrated", Payload: json.RawMessage(`not even valid json`)}
	if _, _, _, ok := m.resolveSubject(e); ok {
		t.Error("world.migrated must resolve unlocatable")
	}
}

// TestJumpOrHintTotality (T007, SC-002): every cataloged event type
// (catalogFixture — the same one-fixture-per-type source TestCatalogSweep
// uses) resolves through resolveSubject + detailActions to exactly one
// jump-or-hint outcome — no panic, and detailActions never returns an empty
// slice or an empty label. A replica with four living, distinctly-placed
// agents exercises the live-position branch broadly; this sweep's job is
// coverage/totality, not re-asserting each type's exact jump target.
func TestJumpOrHintTotality(t *testing.T) {
	m := testModel(t)
	m.replica.Agents = []sim.Agent{
		{Name: "Ash", X: 1, Y: 1},
		{Name: "Birch", X: 2, Y: 2},
		{Name: "Cedar", X: 3, Y: 3},
		{Name: "Rowan", X: 4, Y: 4},
	}

	for typ, fx := range catalogFixture {
		typ, fx := typ, fx
		t.Run(typ, func(t *testing.T) {
			e := store.Event{Seq: 1, Tick: 1, Type: typ, Payload: json.RawMessage(fx.payload)}
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("resolveSubject/detailActions panicked: %v", r)
				}
			}()
			m.resolveSubject(e) // must not panic regardless of ok
			actions := m.detailActions(e)
			if len(actions) != 1 {
				t.Fatalf("detailActions returned %d actions, want exactly 1", len(actions))
			}
			if actions[0].Label == "" {
				t.Error("detailActions returned an empty label")
			}
		})
	}
}

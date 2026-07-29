---
name: guardian-turn-loop
description: The turn's tool-calling loop after the prompt is composed — the converse/act reply channel, capabilities.json + curriculum-stage-ceiling roster gating, persona-bundle grant narrowing, one-acting-call cardinality, the charge-free clock-control meta tools, and the per-call telemetry every termination path emits. Split from [[guardian]]; load when tracing roster/gating or tool-call telemetry.
kind: component
sources:
  - internal/guardian/turn.go
  - internal/guardian/toolcalls.go
  - internal/guardian/charter.go
  - internal/guardian/guardian.go
verified_against: 72f82f41f7aa2e345572105894cd0fb7c02fc0aa
---

# Guardian's turn loop

Split from [[guardian]] — the roster/gating, meta clock-control tools, and
telemetry that run after [[guardian-instruction-surface]]'s prompt is
assembled.

## How it works

The model may reply with words (**converse** — the transcript-only final-answer channel,
`toolloop.Result.Final`; `converse` is deliberately NOT a declared loop tool, so it
can never be rejected as unknown) or call exactly one acting tool. Since spec 029
the declared loop roster is `send_vision`/`send_omen`/`monitor_and_act`/
`cancel_order`/`work_miracle`/`pause`/`start`/`adjust_speed` — plus, since
spec 063, `explain` appended last ([[grounded-feedback]]'s read-only
mechanics-facts tool, `Effect: Read`, never the turn's mediated act) —
(`tool.LoopRosterGuardian`, in that order) — the retired `nudge_dream`/`nudge_omen`
forms are gone from the registry — or fewer: since spec 021 the world's
`capabilities.json` gates the roster per-read through three independent layers, one
`grantedRoster` feeding declaration, guidance prose, and the handler set, with
defense-in-depth `grant.allows` checks in the landers; an ungranted
tool is structurally absent from the declared schemas, never merely prose-forbidden;
missing manifest = full roster byte-compatibly, malformed = full roster + notice,
`tools: []` = conversation-only — converse is never gateable), which lands through
its existing door. Since spec 046 the world's curriculum stage caps the grant
BEFORE any of that: `applyStageCeiling` (`charter.go`) intersects the stage
ceiling into the grant immediately after `loadManifest` and before
`grantedRoster`, using the same `intersectGrant` a persona bundle's narrowing
uses — intersection-only, so a manifest may narrow within the ceiling but never
exceed the stage, and a beyond-stage tool is structurally absent from
declaration, prose, and handlers alike. The stage-1/-2 ceiling
(`stage1CeilingTools`) is `send_omen`/`send_vision`/`monitor_and_act`/
`cancel_order` with NO miracle kinds and no bundle tools (a ratified TASK-119
amendment added the two standing-order tools, since the first-night exercise
teaches the watch as a stage-1 primitive) — plus, since spec 063, `explain`
(read-only, zero-cost, tutor-lane by construction: it is the tutor guide's
own grounding tool and stage-1 IS the orientation stage, so it widens no
acting capability, [[grounded-feedback]]); stage-3, stage-4, and a pre-ladder
world (`stage == ""`) impose no ceiling ([[curriculum-ladder]]). `guardian.
StageCeilingVerbs(stage)` exports the ceiling's granted loop-tool names in
registry order — the SAME intersection this applies — for the TUI help
overlay's D9 guardian section to teach from without a second hand-maintained
list ([[grounded-feedback]], [[tui-client]]). Spec 036
extends the same composition with the bundle surface
([[bundle-tools]]): `runTurn` narrows the world grant by each persona bundle's
`capabilities.json` via `narrowGrantForBundles`/`intersectGrant` (intersection —
a persona can exclude tools or miracle kinds, never resurrect what the world
excludes; commutative across personas), then appends the grant-filtered bundle
roster and handlers after the built-ins (`grantedRoster(grant)` first, then
`BundleSet.Roster()` order), so declaration, guidance prose (via the spec-036
`PromptGloss` fallback, [[tool-registry]]), and the handler map stay one
composition; `loadManifest` takes the known bundle-tool names so a world
manifest naming a bundle tool no longer draws a spurious unknown-tool notice.
The frozen `BundleSet` arrives once at boot via `mt.SetBundles`
([[daemon-lifecycle]]), read only from the turn worker; the stage facts arrive
the same way — the daemon calls `mt.SetStage(stage, charterPreset)` once after
`New` from the immutable `Manifest.Stage`/`Manifest.CharterPreset` (spec 046,
the SetBundles discipline, boot-frozen so the ceiling cannot be tampered
mid-run; zero values = pre-ladder, ungated, default charter);
the driver's one-acting-call cardinality enforces "at most one mediated act per
turn" structurally, so the pre-loop nudge-wins-over-miracle precedence dissolves —
the model just picks its one act. The retired `turnReply`/`parseTurn` free-text
JSON contract (`{say, nudge|null, miracle|null}`) is gone: a door refusal now
becomes a `rejected_gate` fed back to the model within the loop's round cap — a
behavior upgrade over the old single-shot refusal, since a mistyped villager name
can be retried instead of ending the turn outright. The per-round token budget
is `mt.turnTokens` (spec 025, TASK-72: the operator-tunable `llm.json`
`max_tokens.metatron_turn` knob, threaded through `guardian.New` like
`loopRounds`; default 1024, up from the pre-loop 700 — a tool-era round must
carry a `tool_use` block alongside any converse prose in the same round, so
the budget grew to keep a full charter-voiced reply from crowding out a
same-round act). When the loop ends with no text and no landed
act (model_done with nothing, cap exhaustion, or a soft error), the same
scattered-thoughts apology as before covers it.

**Meta tools** (spec 029 US5): `pause`/`start`/`adjust_speed` are charge-free
registered tools (`Effect` Expressive, EMPTY `Events`) that drive the world clock
through the `LoopControl` seam — the SAME `*sim.Loop.Do` the [[ipc-server]] uses, so
a guardian-issued control is indistinguishable from an operator one and lands the
loop's own `clock.paused`/`clock.resumed`/`clock.speed_set` records (no new event
vocabulary). The daemon injects the loop twice at `guardian.New` — as the `Injector`
and as `LoopControl` (the `mind.New(loop, loop)` precedent). Mapping (`controlLoop`):
`pause`→`Do("pause")`, `adjust_speed`→`Do("set_speed", speed)`, and `start` resumes —
a bare start is `Do("resume", "")`, but a `start` WITH a speed issues `Do("set_speed",
speed)` THEN `Do("resume", "")` for the one call (the loop's resume ignores its speed
argument). They inject nothing and spend nothing but consume the turn's one act; each
is grant-gated and structurally absent when the manifest withholds it. The
`guardianInitiativeFrame` (above) binds their use to player authority.

**Tool-call telemetry** (`toolcalls.go`, spec 017 FR-007/T018/T020): every model
tool call the turn's loop saw — landed, rejected, or otherwise — is buffered as a
`toolloop.CallRecord` (the `Job.Record` sink) and lands as one `cog.tool_call`
batch through `InjectSocial` on EVERY termination path (so a rejected or
never-grounded call is recorded even when nothing landed), via the same
`sim.NewCogToolCallPayload` constructor [[agent-mind]]'s mind uses — a converse-
only turn (no tool calls) emits no batch at all. The turn's handlers are exactly
the granted subset of `send_vision`/`send_omen`/`monitor_and_act`/`cancel_order`/
`work_miracle`/`pause`/`start`/`adjust_speed`; `converse` is deliberately absent
from the handler map (and from `tool.LoopRosterGuardian()`) since it is the
final-text channel, never a callable tool. Since spec 025 (TASK-72) the turn
also surfaces the loop's one in-loop transport retry: when
`toolloop.Result.Retried` is set, `emitRetried` (toolcalls.go) lands a
NON-terminal `cog.outcome` carrying `sim.OutcomeRetried` and the first
failure's reason through the same `InjectSocial` door — emitted BEFORE the
error return, so even a twice-failed turn's retry is countable from the trail.

## Connections

[[guardian]] is the parent — prompt assembly is [[guardian-instruction-surface]];
the mediated acts this roster can call are [[guardian-mediated-acts]]/
[[guardian-miracles]]; the standing-order tools are
[[guardian-watch-workers]]/[[guardian-orders]]. [[curriculum-ladder]] owns
the stage ceiling; [[bundle-tools]] owns the persona-bundle narrowing.
[[tool-loop]] is the bounded driver this roster plugs into; [[tool-registry]]
declares the roster and derives its guidance prose. [[agent-mind]] shares the
`cog.tool_call` telemetry constructor.

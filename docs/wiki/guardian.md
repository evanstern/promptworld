---
name: guardian
description: The guardian (TASK-12; guardian-voiced since spec 052 ruling 3, formerly "Metatron") — console AND system-authored turns driven through the bounded tool-use loop (spec 017), omen/vision influence and standing-order agency behind a structural prompt firewall (spec 029), event-sourced charge economy, charge-free clock-control meta tools, digests + drama moments, and the staged player-editable instruction surface (charter + skills/ + capabilities.json, spec 021); spec 036 composes drop-in bundle tools and persona SOUL fragments into the same turn assembly; spec 044 stamps each turn's effective-charter fingerprint into an event-sourced revision timeline (metatron.charter_observed); spec 046 gates the whole surface by the world's curriculum stage; spec 059 turns a system-origin survival watch's match into a SURVIVAL turn with a one-peril initiative carve-out plus a token-bounded miracle-targeting digest; spec 052 (TASK-121) renamed the package internal/metatron → internal/guardian, Go identifiers Metatron* → Guardian*, and wired a boot-frozen runtime skin ([[skin]]) that supplies the guardian's display name/epithet/voice/vocabulary — every serialized string (event types, JSON tags, IPC methods, tool ids, on-disk paths, correlation-id prefixes) stays FROZEN byte-identical
kind: component
sources:
  - internal/guardian/guardian.go
  - internal/guardian/turn.go
  - internal/guardian/toolcalls.go
  - internal/guardian/orders.go
  - internal/guardian/charter.go
  - internal/guardian/digest.go
  - internal/guardian/miracle_batch.go
  - internal/sim/guardian.go
  - internal/persona/charter.go
  - internal/skin/skin.go
verified_against: e137b82bb699eb323eb26c6a69c3dc83ca474b27
---

# Guardian

The guardian is the player's sole verb: a daemon-hosted gatekeeper (`internal/guardian`,
the mind/scribe notify-consumer pattern) that converses in the console, watches the
world, and mediates all influence. Raw player text has exactly one sink — the guardian's
own prompt; villagers can only ever receive the guardian's validated rendering, landed
through [[sim-loop]]'s injection door as recorded events. The meta-game is
prompt-engineering your guardian through the staged instruction surface (spec 021,
TASK-64), shaped like real assistant configuration: `charter.md` (the
CLAUDE.md-shaped base), `skills/*.md` (player-authored SKILL.md-shaped files),
and `capabilities.json` (the per-world tool grant manifest) — all at the
save-dir root, all re-read fresh every turn. Since spec 046
([[curriculum-ladder]]) that surface unlocks in stages: a world created on the
curriculum ladder carries an immutable `Stage` (and optional `CharterPreset`)
in its manifest, and the stage gates both what the guardian's grant may contain
and which instruction files bind — a pre-ladder world (no stage) is ungated,
byte-identical to before.

## How it works

**Turns** (`turn.go`): one directive = one `Turn`, driven through [[tool-loop]]'s
bounded loop (`toolloop.Run`, spec 017 T020) against `llm.KindGuardian` cloud calls
([[llm-orchestrator]]), serialized single-flight. Since spec 029 (TASK-27) the turn
body is extracted into the shared `runTurn`, and there are two origins
(`turnOrigin`): a **console** turn (`Turn`, the player's words) and a
**system-authored** turn (a triggered standing order — see [[guardian-orders]]).
Both run the identical body — same single-flight `turnBusy` guard, same roster/
handler/gate composition, same telemetry, same transcript append — differing only
in framing: the console path opens the transcript with the player's `> …` line and
uses the correlation id `turn-metatron-<tick>`; the system path opens with a
`[watch]` origin marker over the order's pre-authorized action (never a player-text
line — a triggered turn has no player text), uses `watch-metatron-<tick>`, and
suppresses moment consumption (the player-facing queue awaits the next console open;
the trigger worker queues the system turn's own outcome moment). The console CAS-fails
fast with `ErrTurnBusy`; a system turn WAITS bounded for the slot ([[guardian-orders]]).
The prompt stacks the charter (re-read every turn — edits are live by construction,
with restore/empty/truncate fallbacks and in-reply notices, `charter.go`; since
spec 046 the load is the stage fork `stageCharter`: at stage-1 the effective
charter IS the world's `CharterPreset` constant — `presetCharter` resolves
`""`/`"default"` to `persona.DefaultCharter` and `"tutor"` to the stage-1
orientation `persona.TutorCharter` — sourced from the compiled-in text, never
the file, so the lock is tamper-proof rather than advisory; an edited
`charter.md` draws an honest "does not bind at this stage" notice naming the
stage-2 unlock, a missing one is restored to the preset noticelessly, and every
other stage, including no stage, runs `loadCharter` — itself now preset-aware,
so restore/empty/unreadable fallbacks serve the world's preset rather than bare
`persona.DefaultCharter`), then —
since spec 036 — any persona SOUL fragments from the boot-frozen bundle surface
(`mt.bundles.SoulFragments()`, load order, each ≤4,000 chars, validated at boot
by [[bundle-tools]]; zero fragments leaves the prompt byte-identical), then the
skill files (spec 021: `loadSkills` composes eligible `skills/*.md` — regular `.md`
direct children, ascending bytewise filename order, ≤8 files, ≤4,000 chars each via
`persona.CharterMaxChars`, each under a `--- skill: <name> ---` separator, with the
same truncate/skip notice discipline as the charter; since spec 046 behind the
`stageSkills` fork — skill files compose only from stage-3, and at stage-1/-2 a
present-but-unbound `skills/` dir draws one honest notice naming the stage-3
unlock rather than being silently ignored), then a fixed frame appended
LAST as compile-time constants on every path — no editable byte can displace or
truncate it (adversarial battery + determinism tests in `guardian_test.go`). The
frame pins the two `guardianNonNegotiables` invariants beneath ANY editable text
(never invent unobserved events; never pass the player's words to a villager) plus,
since spec 029, the `guardianInitiativeFrame` (T019) that binds clock control and
standing orders to player-requested or pre-authorized action only — never the
guardian's own initiative, with the door-side grant gate backing it independently.
Since spec 059 (US2) that doctrine gets exactly ONE carve-out, keyed on the
turn's origin rather than any tool: `buildTurnSystemPrompt(survival, …)` (the
origin-selecting composer `turnSystemPrompt` now wraps, pinning `survival=false`
for every pre-059 call site) swaps ONLY the initiative frame —
`guardianSurvivalFrame` in place of `guardianInitiativeFrame` when `runTurn`'s
origin is a survival-watch trigger (`turnOrigin.survival`, [[guardian-orders]]) —
leaving the non-negotiables, the tool guidance, and every other byte of the
prompt untouched; a non-survival turn still composes byte-identically to the
pre-059 prompt (FR-005). The survival frame permits a vision or miracle on the
guardian's own initiative to save a life, for that one peril alone — clock control
and every OTHER standing order remain player-authority in a survival turn
exactly as in any other (FR-004); `DefaultCharter`
(`internal/persona/charter.go`) states the same carve-out in-fiction so the
guardian's own narration stays honest about what it may do unprompted. The
frame also carries the acting-tool guidance DERIVED from the registry
(`tool.GuardianToolGuidance` over the world's granted roster, [[tool-registry]]) —
the old hand-written prose tool list is gone, so described ≡ declared by
construction. The turn also stacks live status (clock, ⚡ bank, roster),
queued moments, the [[chronicle]] tail (the guardian reads its village's story — this
grounds fresh reigns and upgraded worlds), its soul tail, and recent transcript. The
model may reply with words (**converse** — the transcript-only final-answer channel,
`toolloop.Result.Final`; `converse` is deliberately NOT a declared loop tool, so it
can never be rejected as unknown) or call exactly one acting tool. Since spec 029
the declared loop roster is `send_vision`/`send_omen`/`monitor_and_act`/
`cancel_order`/`work_miracle`/`pause`/`start`/`adjust_speed`
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
teaches the watch as a stage-1 primitive); stage-3, stage-4, and a pre-ladder
world (`stage == ""`) impose no ceiling ([[curriculum-ladder]]). Spec 036
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

**Charter observation — the revision timeline** (spec 044 US2, FR-008): every
`runTurn` stamps the charter revision it actually runs under. Immediately
after `loadCharter` returns — before anything consumes the text —
`observeCharter` (`charter.go`) fingerprints the EFFECTIVE charter
(`charterFingerprint`: the first 12 hex chars of SHA-256 over exactly the
post-fallback, post-truncation bytes the model executes, so the recorded
revision can never name a charter the guardian never ran) and, when it differs
from the last recorded value, emits `metatron.charter_observed{fingerprint,
default}` through the same `InjectSocial` door as every other turn effect —
fingerprint-at-effect semantics. The `default` flag derives from the same
effective text (an empty/missing `charter.md` serves and records the default),
so the two can never disagree — and since spec 046 it is PRESET-AWARE: default
means the effective text equals the WORLD's preset constant (`presetCharter`,
the same reference `charterIsDefault` compares against), not bare
`persona.DefaultCharter`, so a stage-1 tutor-preset world — whose lock serves
`persona.TutorCharter` — honestly records `default: true`. Preset text is
authored by the game, never the player, so it must never masquerade as
player-authored evidence: the [[morgue]]'s alignment and the stage-2→3 unlock
gate's custom-charter evidence both derive from this flag
([[curriculum-ladder]]). At stage-1 the observation stamps the stage-EFFECTIVE
text the lock serves, never the raw file. The first turn of a world always emits (the
mirror starts empty); an ENDED world skips — the door narrows to recorded
prose after run end and a finished run's evidence timeline is closed. The
`the guardian` struct mirrors `State.CharterFingerprint`/`State.Ended` as
`charterFP`/`ended` under `stateMu`; `observeCharter` sets the mirror
optimistically after a landed emission (so a back-to-back turn cannot
double-emit before absorb catches up), and `mirrorState` moves it FORWARD
only — a batch predating the landed observation absorbs with the replica's
old fingerprint, and overwriting would re-open the emission window. The
reducer arm (`internal/sim/guardian.go`, `CharterObservedPayload`) keeps only
the CURRENT fingerprint on state (rejecting an empty one at the dry-run); the
full revision timeline lives in the event log, where the [[morgue]]'s render
scan aligns each death against the most recent observation at or before it.
Evidence only — the payload carries no scoring fields, by contract.

**Influence: omens and visions** (spec 029, TASK-27): the two mediated forms that
replaced the retired `dream`/`omen` nudges. A **vision** (`send_vision`, `landVision`)
reaches exactly one living villager at ANY hour; an **omen** (`send_omen`, `landOmen`)
reaches one villager, a named comma-separated group, or the word `everyone` — but
only at NIGHT. Each landed act costs exactly ONE charge regardless of recipient
count, console-initiated or triggered. Validation (living target(s), non-empty text,
charges ≥ 1) downgrades failures to refusal-with-counsel — refusals are free, fed
back as a `rejected_gate` the model may repair in a later round. The `dream` form is
gone from the guardian's vocabulary; the spec-014 `OnRoster(RosterGuardian, "nudge_"+form)`
check is replaced by an explicit form switch in the reducer (`vision`/`omen`/`dream`,
with `dream` grandfathered replay-only — no live tool can produce a new one). The
400-byte text cap is still a registry read, re-pointed at `send_vision`'s entry
(`nudgeTextMax` in turn.go, `sim.NudgeTextMax` reducer-side, from the same tool so
truncation and enforcer never diverge).

Since spec 041 (FR-014), `send_vision` also carries an OPTIONAL place grant —
`place_kind`/`place_x`/`place_y`, all-or-none (`toolcalls.go`'s `parseReveal`
refuses a partial triple as a `rejected_gate` before anything lands). When
given, `landVision` (now taking a `*placeReveal` parameter) composes one
`metatron.place_revealed` event plus a companion `agent.memory_added`
("The vision showed you the <kind> at (x,y).", `SalDream`,
`Origin: sim.OriginOmen`) as `extra` events riding the SAME atomic
`landNudgeBatch` call as the vision's own nudge memory — the grant lands with
the vision or not at all. The kind enum is [[mental-maps]]'s closed
place-fact vocabulary; the reducer dry-run (a living target, a REAL place —
`groundFactPresent`) is the semantic authority, so the tool schema can only
over- or under-offer, never land a false fact. A vision without the place
arguments behaves exactly as before.

Both landers share `landNudgeBatch` — the text cap, the ONE atomic `InjectSocial`
batch, and the soul append, VERBATIM the pre-029 `landNudge` body (wrap, don't
rewrite): `metatron.nudged{form, targets, text}` (validating reducer spends the
charge and enforces the omen NIGHT gate at the door; `send_omen`'s day path never
reaches here) + one prefixed (`"You saw a vision: "` / `"You witnessed an omen: "`)
`agent.memory_added` per target at `SalDream` (8), each stamped `Origin: sim.OriginOmen`
(spec 030) — a direct-perception provenance class per `sim.DirectPerception` (same
standing as an own act or a witnessed event), which the villager interprets in
persona. `landVision`/`landOmen` differ only in target
resolution and the per-tool grant gate. The firewall is structural, not behavioral:
no code path exists from model output to any villager surface OR clock control
outside registered tools (sentinel-audited by `TestHandlerFirewallAudit`,
`guardian_test.go`, extended to the spec-029 surface, SC-007). A **daytime**
`send_omen` neither lands nor refuses — it defers to nightfall as a system-origin
standing order ([[guardian-orders]]).

**Miracles** (spec 016, [[guardian-miracles]]): the guardian's other charge-priced
mediated act, spent from the same bank, a declared loop tool: `work_miracle`
(`kind` ∈ `move`/`remove`/`give_item`/`time_snap`). The retired
`turnReply.Miracle` anonymous struct had **no gratis field** as its structural
guarantee; the replacement `miracleArgs` (`toolcalls.go`, the tool-call-parsed
mirror of the same flat surface) keeps that guarantee identically — nothing to
unmarshal `gratis` into, so a model-driven miracle can never waive its charge.
`landMiracle` resolves door-neutral `MiracleParams` (villager name → index,
day/`HH:MM` → tick via [[game-clock]]'s `ParseTimeOfDay`/`TickAt`) from an
`agentXY` snapshot the absorb goroutine mirrors per batch (so the turn worker
never races the live replica), then calls the shared `guardian.BuildMiracleBatch`
— the SAME builder the IPC `miracle` door uses — to compose the event and its
perception-memory companions (each stamped `Origin: sim.OriginOmen`, spec 030,
identically to a nudge's memories), and lands it through `InjectSocial`. A rejection at
the reducer dry-run becomes a `rejected_gate` the loop feeds back (rather than an
immediate reply-suffix refusal, though the wording is the same in-fiction
counsel); a landed miracle also appends a soul-file line. `landMiracle`'s
validation/batch/soul-append logic is likewise UNCHANGED from the pre-loop path —
only the input moved from `turnReply.Miracle` to `work_miracle`'s tool-call
arguments. Since spec 059 (US3), any turn whose granted roster offers
`work_miracle` (`hasWorkMiracle`) additionally carries a token-bounded
targeting digest in the user prompt — living villagers' positions/conditions
plus adjacent passable tiles, `turn.go`'s `buildTargetingDigest` fed by the
same `agentXY` mirror plus a parallel `agentNeeds` mirror, introduced by
`tool.GuardianTargetingGuidance()` ([[tool-registry]]) — so a coordinate-
bearing miracle (`move`/`remove`) can aim at a tile the door will actually
accept (world-01 evidence: 3 of 4 miracle attempts door-rejected on invalid
coordinates). Prompt surface only; `landMiracle`'s reducer dry-run stays the
sole authority on whether a targeted coordinate lands — see
[[guardian-miracles]] for the digest's assembly and cost.

**Standing orders** (spec 029, its own note [[guardian-orders]]): `monitor_and_act`
places an event-sourced watch-and-act order whose condition, compiled once, is
matched for free against the live event stream; when it fires, the guardian wakes and
runs the pre-authorized action as a system-authored turn through this same door.
`cancel_order` retires one. `handleMonitor`/`handleCancelOrder` (`toolcalls.go`) wrap
the door helpers `placeOrder`/`cancelOrder` (`orders.go`); the turn prompt carries
active orders (`writeStandingOrders`, FR-017) and `Status.Orders` lists them
model-free (FR-016). The full lifecycle, event sourcing, matching, trigger execution,
fuzzy confirm, and daytime-omen deferral live in [[guardian-orders]]. Since
spec 059, three SYSTEM-origin survival watches (near-death, starvation,
exposure) stand in every world from boot without any player action — they
share this same event-sourced order machinery but are origin-exempt from the
player cap, TTL, and cancellation, match live via a hysteresis-latched
danger-band predicate rather than the structural filter, and fire a
SURVIVAL turn (above) rather than an ordinary system turn — the full
mechanics live in [[guardian-orders]]'s own section.

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

**Charge economy** (`internal/sim/guardian.go`): `State.GuardianCharges` — genesis
1, cap 3, +1 per absolute 6-game-hour boundary emitted by the [[executor]]
(`metatron.charge_regenerated`, a pure function of state + tick), −1 per landed
omen or vision (a miracle spends its per-kind cost). Fully event-sourced: replay
reproduces the bank exactly; the field is
deliberately not `omitempty` so a spent-to-zero bank survives snapshots; pre-TASK-12
snapshots gain the genesis charge on upgrade.

**Watching** (`digest.go`): notable events collect per 6-game-hour window; each
non-empty window costs one summarization call appended to `metatron/soul.md`
(skip-empty is free; failures carry lines into the next window). The drama rule v1:
`agent.died`, `gru.attacked`, `social.promise_broken`, and (since spec 029)
`metatron.order_expired` append model-free **moment** lines immediately and queue for
the console — the next reply leads with them. Digests and moments themselves never
construct an act; the guardian acts only when the player asks OR a standing order the
player placed authorizes it (spec 029 relaxed the old "acts only when told" contract
to admit pre-authorized triggered turns — see [[guardian-orders]]).

**Files** (bound to the run, not event-sourced): `charter.md` at the save-dir root
(seeded by `persona.Genesis` — since spec 046 with an optional preset parameter,
so a `"tutor"` world seeds `persona.TutorCharter` — never overwritten), plus the
optional player-created
`skills/` dir and `capabilities.json` manifest beside it (spec 021 — root =
player-authored configuration); `metatron/soul.md` (accreting notes, starts empty)
and `metatron/transcript.md` (console history) — restart survival comes free with
files, and world determinism never depends on them (prompt composition is upstream
of the recorded LLM inputs).

**Surfaces**: IPC `metatron_chat`/`metatron_status` (FROZEN wire method names,
spec 052 ruling 2 — [[ipc-protocol]]), CLI
`promptworld guardian <dir> [message…]` (canonical since spec 052 FR-008; the
pre-052 `metatron` name survives as a hidden compat alias — [[cli-promptworld]]), and the
[[tui-client]] guardian pane (the console). Protocol status (`guardian.Status`, the
model-free peek, computed fresh from disk per call) carries the ⚡ bank, charter
provenance (`charter_default`), and since spec 021 the effective skill filenames
(`skills`, composition order), the granted roster (`granted_tools`, registry order,
`work_miracle(move,give_item)` form when kinds are restricted),
`manifest_default` (no `capabilities.json` present), and since spec 029 the active
standing orders (`orders`, `OrderStatus{id, condition, origin, fuzzy, expires_day,
status}`, FR-016 — see [[guardian-orders]]). Since spec 046 the status is the
turn's stage twin: `Status()` applies the same `applyStageCeiling` to the
peeked grant (so `granted_tools` can never disagree with the roster the next
turn will run under), nils the `skills` list below stage-3 (it is the EFFECTIVE
composition list), and carries additive omitempty curriculum provenance —
`stage`, `charter_locked` (the stage-1 lock is in force), `charter_preset` (the
binding preset name when locked, `"default"` | `"tutor"`), and `skills_locked`
(stage-1/-2); `charter_default` compares against the world's preset constant
via the preset-aware `charterIsDefault`.

## Connections

[[sim-loop]] whitelists `metatron.nudged`, `metatron.place_revealed` (spec 041, a
vision's optional place grant), `metatron.charter_observed` (spec 044, the
revision timeline the [[morgue]] aligns deaths against), the four `metatron.*`
miracle types, and
the three injected order events (`order_placed`/`order_cancelled`/`order_triggered`);
[[mental-maps]] owns the place-fact vocabulary `send_vision`'s place grant draws
on and the reducer arm that validates and upserts it;
[[sim-state-reducer]] holds the bank, the miracle reducer arms ([[guardian-miracles]]),
and the standing-order arms ([[guardian-orders]]); [[executor]] regenerates the bank
and emits `metatron.order_expired`; [[event-types]] catalogs all three families;
[[llm-orchestrator]] routes `KindGuardian` to the cloud tier and the fuzzy
`KindGuardianWatch` confirm to a cheap chain; [[chronicle]] feeds the guardian's
grounding; [[agent-mind]] is how villagers interpret what lands; [[daemon-lifecycle]]
wires the component behind the LLM-config gate, passing the loop as both `Injector`
and `LoopControl`; [[ipc-server]]'s `handleMiracle` is the miracle's other door,
sharing `BuildMiracleBatch` with `landMiracle` here, and the clock commands the meta
tools reuse. [[tool-loop]] is the turn driver (console and system-authored) since spec
017: `runTurn` calls `toolloop.Run` with `tool.LoopRosterGuardian()` and the granted
handler subset; [[tool-registry]] declares those tools (and deliberately excludes
`converse`), derives the turn's tool guidance (`GuardianToolGuidance`), and holds the
single miracle cost source ([[guardian-miracles]]). [[skin]] is the
boot-frozen display substrate `SetSkin` installs — `Name`/`Epithet`/
`Voice`/`WorkingNoun`/`FormNoun`/`StageName` are this note's composition
seam into the turn prompt, the confirm/digest system prompts, and
moment/notice text, never into a recorded payload. [[curriculum-ladder]]
(spec 046) owns the stage vocabulary, the manifest facts (`world.Manifest.Stage`/
`CharterPreset`), the exercise/unlock event sourcing, and the earned-stage
doctrine this note's stage ceiling and charter lock enforce. [[guardian-orders]]
(spec 059) also owns the three system survival watches, their origin-keyed
exemptions, the live hysteresis matcher, and the survival-turn frame this
note's `buildTurnSystemPrompt` composes; [[daemon-lifecycle]] seeds them at
boot (`seedSurvivalWatches`); [[guardian-miracles]] owns the spec-059 miracle
targeting digest this note's Miracles section describes. Specs:
`specs/005-metatron/`, `specs/016-metatron-miracles/`,
`specs/017-agent-tool-loop/`, `specs/021-metatron-instruction-surface/`,
`specs/029-metatron-agency/`, `specs/046-curriculum-ladder/`,
`specs/059-metatron-survival-autonomy/`.

## Operational notes

Live-proven on a fresh world (reign-test: judged dream landed atomically, exhaustion
refused with counsel, BRUTUS charter edit live next turn, digest + regen at the 12:00
boundary) and on the 14-day chronicle-proof world (upgrade granted the genesis
charge; the guardian answered "what do you know of Fern and the voice at the well?"
from the chronicle ring, honestly bounded its knowledge, then landed an in-world
dream that wove in the smooth stone from the story). Live finding folded back: the
no-invention rule originally lived in the (replaceable) default charter — a surly
custom charter invited fabricated villager activity; both invariants now sit in the
fixed frame. (Those reign-tests predate spec 029, when the live influence form was
still `dream`; the vocabulary is now omen/vision, but the atomicity, exhaustion, and
charter-edit findings carry over unchanged.) Cost: ~4 digests/game-day + player turns
+ any triggered watch turns, noise against the ceiling. Spec 029 (TASK-27) shipped
standing orders and pre-authorized autonomous action ([[guardian-orders]]); spec 059
(TASK-111) shipped the guardian's first true own-initiative authority — survival
watches seeded from birth, a turn-origin-conditional frame carve-out, and a
miracle targeting digest — as a near-term slice of the broader agentization
direction (TASK-112); still parked for post-v1: world-tools, full regency,
drama-based cloud escalation of villager minds.

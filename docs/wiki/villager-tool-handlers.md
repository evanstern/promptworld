---
name: villager-tool-handlers
description: How every acting villager tool (world verbs, set_plan, muse, journal writes) wraps an existing landing door rather than mutating the world directly, and how musing became an ordinary roster tool instead of a scheduled channel. Split from [[agent-mind]] (Villager tool handlers + Musing sections).
kind: component
sources:
  - internal/mind/handlers.go
  - internal/mind/parse.go
verified_against: d0645811c9783d1248dc65ed0fcf0b37524dd8fd
---

# Villager tool handlers

**Villager tool handlers** (`internal/mind/handlers.go`): every acting tool
wraps an existing landing door, never mutating the world directly — the loop
REQUESTS through the door and translates its verdict into a `toolloop.Outcome`.
`handleWorldVerb(name)` mirrors the pre-loop `InjectIntent` call for one world
verb (the same `InjectArgs` fields, minus the free-text reason, which the tool
era carries via `muse` instead of a per-action field); `talk_to` keeps its
mind-side `buildTalkToGuards` (target alive + present in the job's snapshot
worldview). Since spec 064 (`warm_up`), `handleWorldVerb` also special-cases
one world verb's argument: an explicit `until_warmth` rides in on the generic
`Qty` slot (the storage verbs' per-verb-arg precedent), and the handler
phrases the model-facing clamp notice via the SAME `sim.ClampWarmUp` the sim's
own resolver clamps with — so the notice and the landed value can never drift
(the `set_plan` const-vs-const precedent) — returning `VerdictLandedClamped`
when the requested threshold was out of range, the first time that verdict
(previously `set_plan`/`muse`-only, spec 058) applies to an ordinary world
verb. `handleSetPlan` parses the tool call's `steps` argument into
`[]sim.PlanStep` (`parsePlanSteps`) and lands them via `InjectIntent`'s `Plan`
path, mirroring the retired `injectPlan`; since spec 058 (US2, FR-003) it also
notices when the submitted step count exceeds `sim.PlanStepCap` BEFORE the
door call (purely to phrase the model-facing notice and pick the verdict —
the landing guard's own truncation, `internal/sim/landing.go`, is the sole
source of truth for what actually lands, a const-vs-const comparison that can
never drift), returning `VerdictLandedClamped` with a "plan set (clamped: N
steps submitted, only the first cap landed)" result instead of the plain
`VerdictLanded`. `handleMuse` lands the musing text as
an `agent.thought{source: "musing"}` through `Loop.InjectSocial`, batched
atomically with its `cog.outcome{landed}` — the exact landing the old
scheduled-musing worker used, now driven by the model choosing the `muse` tool
instead of a cadence firing it. Since spec 058 (FR-001) the driver already
CLAMPS an over-cap musing to `muse`'s 200-rune cap before dispatch (`muse.text`
carries `Clamp: true`, [[tool-registry]]), so `handleMuse`'s own defensive
rune-cap truncation — identical to the pre-058 `parseMusing` shape — is now
belt-and-suspenders and never actually fires; the clamp itself and its
`landed_clamped` verdict are the loop's concern
([[tool-loop]]), not this handler's, whose own `cog.outcome` stays
`OutcomeLanded` either way (a different telemetry axis: cognition completion,
not clamp). Every handler that touches a door sets
`doorOutcome = true` on the dispatch (the door already recorded its own
`cog.outcome`, atomically with the landing/rejection); a handler that refuses
BEFORE touching a door (unknown `talk_to` target, unparseable plan steps)
leaves it false, so `runPlan`'s terminal switch knows to emit its own outcome.

Spec 019 threads a new optional per-action **reason** and wires the four
**journal tools** by name (the `villagerHandlers` switch now dispatches
`write_journal_entry`/`delete_from_journal`/`search_journal`/`read_journal` by
name, ahead of the generic World/Read arms). `reasonArg(call.Args)` reads the
optional `reason` argument (trimmed, defensively capped at `tool.ReasonCapRunes`
= 200) and both `handleWorldVerb` and `handleSetPlan` pass it as
`InjectArgs.Reason` — the intent carries it to completion, where the executor
bakes it into the completion memory's `Why`; the `Loop.InjectIntent` ladder
narrates it as the `agent.thought`. Since spec 058 (FR-001) the loop's
`validateArgs` already CLAMPS an over-cap `reason` before dispatch — for world
verbs via `Param.Clamp` (`reasonParam()`), and for `set_plan`'s own top-level
`reason` (which rides its authored schema, not a `Param`) by field name — so
`reasonArg`'s own rune-cap is belt-and-suspenders, not the enforcement point. The two journal WRITE handlers
(`handleWriteJournal`/`handleDeleteJournal`) mirror `handleMuse` exactly: they
marshal a `journal.entry_written`/`journal.entry_deleted` event and land it
through `Loop.InjectSocial` batched atomically with a `cog.outcome{landed}`. The
reducer dry-run at the door is the sole gate (the 4000-rune budget for a write,
entry existence for a delete); `journalDoorResult` translates the door result —
success sets `doorOutcome` and returns `VerdictLanded`, a door rejection is
peeled with `errors.Unwrap` so the model sees the gate's reason verbatim as
`VerdictRejectedGate` (the agent can curate and retry), and a non-wrapped error
surfaces as `Err` (infra failure → FR-015 terminal outcome). The two READ
handlers (`handleSearchJournal`/`handleReadJournal`) ground nothing: they read
the cognition's own **journal snapshot** (`d.job.journal`, below) via
`SearchJournal`/`FindJournalEntry`/`JournalEntries`, formatting matches with
`formatJournalEntries` ("#<id> <clock>: <text>") and returning `VerdictReadOK`
(or `VerdictReadError` for an unknown addressed id); zero matches is a
well-formed empty read, never an error. `argInt` reads the integer `entry` id
(float-tolerant, like `argKindQty`). All four are villager-only ([[agent-journal]]).

Reads run in the planner worker goroutine, which must never touch the
absorb-owned replica, so `plan()` snapshots each due agent's journal —
`job.journal = a.Journal.Clone()` — into an immutable per-cognition `*sim.Journal`
carried on `planJob`. The snapshot is what search/read see; writes and deletes
land through the live `InjectSocial` door, not the snapshot.

**Musing** (TASK-21, retired as a scheduled channel by spec 017 R10): a
villager no longer has its own 15-game-minute best-effort cadence, queue,
stagger, or fairness floor (`museCadenceTicks`/`museBusy`/`museDue`/
`museStarveWindow`/`lastMuseOK` and the `muse()` worker are all gone, along
with `KindMusing` itself — [[llm-orchestrator]], [[cognition]]). Musing is now
an ordinary roster tool (`muse`, Expressive, `handleMuse` above) the model may
choose inside its planner tool-use loop — interiority carries the SAME
opportunity cost as any other action, since choosing to muse means not
choosing to act, rather than riding a parallel best-effort channel that could
never compete with real cognition for a worker slot. A musing still lands as a
single `agent.thought{source: "musing"}` batched atomically with its
`cog.outcome{landed}`, and it is still recorded via the loop's normal
`cog.tool_call` trace like any other call — but there is no separate call kind,
cadence, or admission path left to describe; it is one line in
`villagerHandlers`, not a subsystem of its own. `parseMusing` (parse.go)
survives, unrenamed, as the shared one-plain-line parser [[governance]]'s
meeting rephraser also consumes — the scheduled musing it was originally named
for is gone, but its shape (first line, quotes/whitespace stripped, rune-capped)
is still exactly what a plain-text reply needs.

## Connections

[[agent-mind]] is the parent note this child was split from; [[tool-use-dispatch]]
is the loop driver that dispatches these handlers each round;
[[tool-registry]] declares the tools/schemas/clamp params these handlers
enforce (`Param.Clamp`, `set_plan`'s own top-level `reason`); [[tool-loop]]
owns the `landed_clamped`/other verdict vocabulary a handler returns;
[[agent-journal]] owns the journal reducer arms `handleWriteJournal`/
`handleDeleteJournal` land through and the per-cognition journal snapshot
the two read handlers consult; [[sim-loop]] owns `InjectIntent`/
`InjectSocial`, the landing doors every handler ultimately calls;
[[governance]] shares `parseMusing`'s plain-line parser with its meeting
rephraser.


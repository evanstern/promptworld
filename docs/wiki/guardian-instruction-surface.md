---
name: guardian-instruction-surface
description: How the guardian assembles its turn prompt, in composition order — stage-aware charter lock, persona SOUL bundle fragments, the compiled-in tutor guide, player skill files, then the fixed frame (non-negotiables, tool/read guidance, the spec-059 survival-turn carve-out). Split from [[guardian]]; load when tracing prompt-stacking order or the staged instruction-surface unlock (spec 021/046).
kind: component
sources:
  - internal/guardian/turn.go
  - internal/guardian/charter.go
  - internal/persona/charter.go
  - internal/skin/skin.go
verified_against: 1bdc50c647a87b2ac221fe073f404df3e3ccd38f
---

# Guardian's instruction surface

Split from [[guardian]] (summary-style, corpus-spec v2) — the turn's prompt
assembly, from the charter through the fixed frame. See [[guardian]] for the
turn driver overview and links to the sibling splits ([[guardian-turn-loop]],
[[guardian-mediated-acts]], [[guardian-watch-workers]],
[[guardian-runtime-facts]]).

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
by [[bundle-tools]]; zero fragments leaves the prompt byte-identical), then —
since spec 063 ([[grounded-feedback]]) — the compiled-in tutor guide
(`persona.TutorGuide`) on a tutor-preset world ONLY (keyed on the charter
preset, not the stage, so a tutor world keeps its guide as it climbs the
ladder; `""` and byte-inert everywhere else, FR-004), then the
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
construction. Since spec 063 ([[grounded-feedback]]) a sibling
`tool.GuardianReadGuidance` renders any granted READ tool's own paragraph
(today, `explain`) — composed BEFORE the acting block so the acting
doctrine's own closing sentence stays the prompt's LAST byte (the frame
invariant the adversarial battery pins); empty and byte-inert when the
roster grants no read tool. The turn also stacks live status (clock, ⚡ bank, roster),
queued moments, the [[chronicle]] tail (the guardian reads its village's story — this
grounds fresh reigns and upgraded worlds), its soul tail, and recent transcript.

## Connections

[[guardian]] is the parent — its turn driver calls into this composition
before any tool call, and its "Turns" summary points here for the full
stacking order. [[curriculum-ladder]] (spec 046) owns the stage vocabulary
this composition forks on (`stageCharter`, `stageSkills`) and the
`CharterPreset` doctrine. [[bundle-tools]] (spec 036) owns the persona SOUL
fragment producer this stacks in; [[grounded-feedback]] (spec 063) owns the
compiled-in tutor guide and the `explain` read-tool guidance composed
alongside the acting-tool guidance. [[tool-registry]] derives
`GuardianToolGuidance`/`GuardianReadGuidance` from the granted roster.
[[skin]] supplies the boot-frozen `Voice`/`Epithet`/`Name` substrate this
composition and the frame's notices read from.

---
name: guardian
description: The guardian (TASK-12) — the player's sole verb, console/system-authored turns through the bounded tool-use loop (spec 017), mediating all influence behind a structural prompt firewall. Overview only — prompt assembly is [[guardian-instruction-surface]], roster/telemetry [[guardian-turn-loop]], omens/visions/miracles [[guardian-mediated-acts]], standing orders/workers [[guardian-watch-workers]], charter-fingerprint/charge-bank/files/surfaces [[guardian-runtime-facts]].
kind: component
sources:
  - internal/guardian/guardian.go
  - internal/skin/skin.go
verified_against: 657c770f87404b936a0587db1f6b00e81b9f0ee6
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
ladder carries an immutable `Stage` (and optional `CharterPreset`) in its
manifest, and the stage gates both what the guardian's grant may contain and
which instruction files bind — a pre-ladder world (no stage) is ungated,
byte-identical to before.

## How it works

**Turns** (`turn.go`): one directive drives one `Turn` through [[tool-loop]]'s
bounded loop (spec 017), console-originated or system-authored via a
triggered standing order (see [[guardian-watch-workers]]); both run the
identical body — same single-flight guard, roster/handler/gate composition,
telemetry — differing only in framing. Before any tool call, the turn
stacks its prompt in a strict, byte-pinned order — the stage-aware
charter, persona SOUL bundle fragments, the tutor guide, player skill
files, then a fixed frame (non-negotiables, tool/read guidance, the
spec-059 survival-turn carve-out) appended LAST as compile-time constants
no editable byte can displace. See [[guardian-instruction-surface]] for the
full composition.

Once composed, the model replies with `converse` text or calls exactly one
acting tool from a roster gated by `capabilities.json` and the world's
curriculum-stage ceiling (intersection-only, never widening); a persona
bundle can further narrow the grant. A door rejection becomes a
`rejected_gate` fed back within the loop's round cap, and every tool call
the loop saw — landed or not — lands as `cog.tool_call` telemetry. See
[[guardian-turn-loop]] for the roster/gating/stage-ceiling mechanics, the
meta clock-control tools, and telemetry.

Two mediated forms carry the guardian's influence and cost a charge from
the same bank: a **vision**/**omen** (`send_vision`/`send_omen`) lands a
perception memory on a target villager, a named group, or (omens,
night-only) everyone; a **miracle** (`work_miracle`) spends the bank on a
concrete world edit (move/remove/give_item/time_snap) via the shared
`BuildMiracleBatch` door. See [[guardian-mediated-acts]] for both, and
[[guardian-miracles]] for the miracle event types, cost table, rebase
taxonomy, and landing doors. Since spec 084 a third, CHARGE-FREE verb
family — the durable plan layer — sits beside them:
`place_designation`/`cancel_designation` stake checkable world plan
artifacts, `issue_directive`/`cancel_directive` bind villagers to them (a
HARD command, executed between survival and free time), and `survey_site`
reads a deterministic site fact sheet; [[guardian-designations]] owns the
whole subsystem.

`monitor_and_act`/`cancel_order` place and retire event-sourced
watch-and-act standing orders the guardian fires on its own initiative when
a player-placed condition matches; three system-origin survival watches
(near-death, starvation, exposure) stand from boot, origin-exempt from the
player cap. A digest worker summarizes notable events per 6-game-hour
window and queues drama moments; a report-card worker renders
stopping-point critiques; `Close` drains all four background goroutines
before returning. See [[guardian-watch-workers]] for the lifecycle
summary/shutdown discipline, and [[guardian-orders]] for the full
standing-order mechanics.

Every turn stamps the charter revision it ran under into an event-sourced
fingerprint timeline (`guardian.charter_observed`) the [[morgue]] aligns
deaths against; a charge bank (genesis 1, cap 3, +1 per faith-band regen
boundary — spec 085: 4h/6h/12h by band, the ambient forsaken floor at 24h,
stopped in a forsaken scenario world; [[guardian-faith]]) funds every
mediated act; `charter.md`, `metatron/soul.md`, and `metatron/transcript.md`
persist state outside the event log; the component surfaces through IPC
(`metatron_chat`/`metatron_status`), the `promptworld guardian` CLI, and
the [[tui-client]] pane. See [[guardian-runtime-facts]] for the
fingerprinting semantics, ledger, files, and status-peek details.

## Connections

This note's summary-style splits — [[guardian-instruction-surface]] (prompt
assembly), [[guardian-turn-loop]] (roster/gating/telemetry),
[[guardian-mediated-acts]] (omens/visions/miracles),
[[guardian-watch-workers]] (standing orders/workers), and
[[guardian-runtime-facts]] (fingerprint/charge bank/files) — hold the
mechanics summarized above; each links back here.

[[sim-loop]] whitelists every event family this component emits and
reattaches the static map to the `InjectSocial` dry-run probe;
[[sim-state-reducer]] holds the bank and dispatches the miracle
([[guardian-miracles]]) and standing-order ([[guardian-orders]]) reducer
arms; [[executor]] regenerates the charge bank and emits
`guardian.order_expired`; [[event-types]] catalogs all three event
families. [[tool-loop]] is the turn driver since spec 017 (`toolloop.Run`);
[[tool-registry]] declares the roster (excluding `converse`), derives the
turn's tool guidance, and holds the single miracle cost source.
[[llm-orchestrator]] routes `KindGuardian` to the cloud tier; [[chronicle]]
feeds the guardian's grounding; [[agent-mind]] is how villagers interpret
what lands; [[daemon-lifecycle]] wires the component behind the LLM-config
gate, passing the loop as both `Injector` and `LoopControl`, and seeds the
survival watches at boot. [[skin]] is the boot-frozen display substrate
this note's composition seam draws
`Name`/`Epithet`/`Voice`/`WorkingNoun`/`FormNoun`/`StageName` from — never
into a recorded payload. [[curriculum-ladder]] (spec 046) owns the stage vocabulary/manifest facts
the stage ceiling and charter lock enforce. [[mental-maps]] owns the
place-fact vocabulary a vision's place grant draws on.
[[grounded-feedback]] (spec 063) owns the `explain` tool, the tutor guide,
and the report-card producer; [[takeover-surfaces]] is its TUI-side
consumer. Specs 005/016/017/021/029/046/059/063.

## Operational notes

Live-proven on a fresh world (reign-test: judged dream landed atomically,
exhaustion refused with counsel, BRUTUS charter edit live next turn, digest
+ regen at the 12:00 boundary) and on the 14-day chronicle-proof world
(upgrade granted the genesis charge; the guardian answered "what do you
know of Fern and the voice at the well?" from the chronicle ring, honestly
bounded its knowledge, then landed an in-world dream weaving in the smooth
stone from the story). Live finding folded back: the no-invention rule
originally lived in the (replaceable) default charter — a surly custom
charter invited fabricated villager activity; both invariants now sit in
the fixed frame. (Predates spec 029, when the live influence form was
still `dream`, now omen/vision — the findings carry over unchanged.) Cost:
~4 digests/game-day + player turns + triggered watch turns, noise against
the ceiling. Spec 029 (TASK-27) shipped standing orders and pre-authorized
autonomous action ([[guardian-orders]]); spec 059 (TASK-111) shipped the
guardian's first true own-initiative authority — survival watches seeded
from birth, a turn-origin-conditional frame carve-out, and a miracle
targeting digest — a near-term slice of the broader agentization direction
(TASK-112); still parked for post-v1: world-tools, full regency,
drama-based cloud escalation of villager minds.

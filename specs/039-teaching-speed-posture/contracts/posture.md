# Contract: Teaching-Posture Surfaces

All surfaces are **additive and non-blocking** (spec FR-007; inherits spec 035's
contracts/warnings.md doctrine: "nothing blocks, fails, or re-routes"). The engine's
routing and the `max`-speed rejection are byte-for-byte untouched.

## 1. Manifest (`world.json`)

```json
{ "name": "classroom-1", "seed": 42, "format_version": 3, "teaching": true, ... }
```

- `teaching` absent ⇒ non-teaching; files without it round-trip unchanged
  (`omitempty`). FormatVersion stays 3; old daemons reading a teaching manifest simply
  ignore the field (defaulting bool — forward-benign).

## 2. Boot behavior (daemon stdout + recorded event)

Teaching world with orchestrator, at every daemon boot after calibration seeding:

- Compute posture rung; issue the loop's normal `set_speed` command ⇒ recorded
  `clock.speed_set` event (replay byte-identity preserved). Rung `0` clamps to `1x`.
- Calibrated: one stdout line, e.g.
  `daemon: teaching posture: defaulting speed to 16x (planner-safe at 17.0s/pt, calibrated 2026-07-24T…)`
- Uncalibrated (planner-serving provider bootstrap-seeded): the line marks the rung
  provisional AND the existing uncalibrated boot-warning block gains its teaching
  flavor — an explicit `run \`promptworld calibrate <world>\`` prompt stating the
  posture cannot yet be honest (spec US3; aligns with spec 035 US2, does not replace
  it).
- Non-teaching worlds: stdout and events byte-identical to today. Pure-sim teaching
  worlds: no posture line, no event, no prompt loop.

## 3. `set_speed` reply — posture warning (existing `warning` field)

`StatusData.Warning` (spec 035 field) on the set_speed path now composes up to two
non-blocking texts, newline-joined, in this order:

1. **Posture override** (new; teaching world + orchestrator + requested rung >
   posture rung): per suppressed watched class, the router's own
   `Verdict.Arithmetic` verbatim plus a plain-language consequence. Example:

   ```
   above teaching posture 16x: planner 3pt x 17.0s/pt x 32x = 1632 ticks over budget 1200 — villagers will stop deep-thinking (reflex only)
   ```

   Fires for calibrated AND uncalibrated teaching worlds. Never fires at or below the
   posture rung, and never on non-teaching worlds.
2. **Uncalibrated warning** (spec 035, unchanged): its own gating (bootstrap-seeded
   serving providers) and text are untouched.

The speed change ALWAYS applies when validation passed — warning-augmented success.
The `max`-gate error path is untouched and takes precedence. `warning` stays
`omitempty` and set only on the set_speed reply path.

Consequence phrasing per degrade mode: planner→"villagers will stop deep-thinking
(reflex only)"; conversation→"conversations will be skipped"; meeting→"meetings fall
back to template speeches".

## 4. Status-family replies — `posture` block (new, additive)

Present ONLY when the world is teaching AND has an orchestrator (precedent: spec 037
`horizon`); everyone else's replies stay byte-identical:

```json
"posture": { "rung": "16x", "calibrated": true }
```

- `rung`: current posture rung as a ladder speed string (`"1x"`…`"32x"`), recomputed
  per reply from the planner-serving provider's live estimate.
- `calibrated`: `CalibratedAt(servingProvider) != ""` — same provenance predicate as
  spec 035/037. `false` ⇒ rung is the pessimistic bootstrap derivation (provisional).
- Clients must treat unknown fields as ignorable (existing IPC convention).

## 5. CLI

- `promptworld new <dir> --teaching` — sets the manifest marker at creation.
- `promptworld teaching <world> [on|off]` — prints or toggles the marker; toggle is
  offline manifest read-modify-write (daemon reads it at next boot; the command says
  so). Exit non-zero only on IO/parse errors, never on state.
- `promptworld status <world>` — for teaching worlds renders one extra line from the
  `posture` block, e.g. `teaching posture: 16x (calibrated)` or
  `teaching posture: 16x (provisional — run \`promptworld calibrate <world>\`)`.
- `promptworld speed <world> <val>` — renders the composed warning via the existing
  `WARNING:` block (spec 035 `setSpeedLine`); no new flags, nothing prompts or blocks.

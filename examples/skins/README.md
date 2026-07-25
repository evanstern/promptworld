# Guardian skins

A **skin** re-themes the fiction layer of one world — the guardian's display
name, voice, tab label, narration vocabulary, and stage display identities —
without touching a single mechanic: same tools, same costs, same rules, same
recorded events (spec 052, ruling 1: the event log is skin-free).

## Install

Copy a skin file into a world's save directory as `skin.json`, next to
`charter.md`, and restart the world:

```sh
cp examples/skins/raven.json ~/.promptworld/worlds/demo/skin.json
promptworld stop demo && promptworld start demo
```

The skin is **boot-frozen**: edits take effect on the next daemon start. The
status surfaces report the active skin, so a stale edit is diagnosable.

## Format

`raven.json` in this directory is the living example. All fields are
optional; anything absent (or invalid) falls back to the default Guardian
skin with one honest notice at daemon boot — a typo never bricks the world.

| Field | What it skins | Constraints |
|---|---|---|
| `name` | the proper name (pane header, chronicle subject lines, prompt name substitution) | single line, ≤ 40 characters |
| `epithet` | the common-noun references ("ask the raven anything…", transcript labels) | single line, ≤ 20 characters |
| `tab_label` | the dock tab | single line, ≤ 20 characters |
| `voice` | persona text composed into the guardian's system prompt, in the editable zone — authoring it is a prompt-engineering exercise | ≤ 4,000 characters |
| `strings` | token-path → value overrides for the display vocabulary (see the token table in `docs/design/tui/patterns/skin-tokens.md`) | unknown tokens ignored + notice; single-line values |
| `stages` | display identities for the neutral stage ids `stage-1`..`stage-4` | missing halves fall back to the default identity |

## What a skin can never touch

- The guardian's **fixed frame** (spec 021): the never-invent-events and
  never-relay-player-words invariants, initiative binding, and tool guidance
  are compile-time constants appended after every skin byte. A hostile voice
  cannot displace them.
- **Mechanics**: tool ids (`send_vision`, `send_omen`, `work_miracle`),
  costs, charges, reducer rules, stage ceilings.
- **The event log**: recorded types, payloads, memory text — skinning is
  render-time and prompt-composition only, which is what makes two worlds
  with different skins provably identical in behavior.
- **Systems/telemetry content**: no skin tokens exist for it, by design.

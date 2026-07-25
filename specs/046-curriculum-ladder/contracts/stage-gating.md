# Contract: stage gating

## The ceiling table (authoritative)

Pinned against the spec-021 roster (`tool.LoopRosterMetatron()`); exact tool names
resolved at implementation against the live roster and recorded here in-PR:

- **stage-1 (The Voice)**: conversational reply + read/query tools + basic nudge
  (vision/whisper-class); no world-shaping miracles, no standing-order placement
  beyond the basic set the client's "basic tools" statement names; no bundles.

  **Pinned (in-PR, against the live roster — `stage1CeilingTools`,
  internal/metatron/charter.go)**: `send_omen`, `send_vision`; miracle kinds:
  none. Conversation is the reply channel, not a roster tool (never gateable),
  so it needs no pin. Excluded and why: `work_miracle` (world-shaping),
  `monitor_and_act` / `cancel_order` (standing-order power tools; a daytime
  omen's nightfall deferral survives — system-origin placement carries
  send_omen's gate, not monitor_and_act's), `pause` / `start` / `adjust_speed`
  (clock control — neither query nor nudge; the player keeps direct CLI/TUI
  clock control at every stage). The live roster carries no read/query tools —
  that clause is vacuously satisfied and recorded here so a future read tool
  must be deliberately classified. No bundle tools (the explicit ceiling list
  intersects them away).
- **stage-2 (The Written Word)**: stage-1 set unchanged; the unlock is the
  *instruction surface* (charter binds), not new tools.
- **stage-3 (The Craft)**: skill files compose; `capabilities.json` honored within the
  ceiling; the grantable tool manifest opens (all remaining non-capstone tools +
  miracle kinds; bundles per spec-036 rules).
- **stage-4 (The Stewardship)**: full roster including capstone capabilities
  (canonization when TASK-81 ships).

## Rules

1. `effectiveGrant = playerManifest ∩ stageCeiling` — intersection-only, applied at
   every manifest load site (turn + status), before roster derivation. A player
   manifest may narrow within the ceiling, never exceed it.
2. Three-layer coherence is inherited, not re-implemented: the declared roster, the
   prose guidance, and the door checks all derive from the post-intersection grant.
   Beyond-stage capabilities are structurally absent (never prose-forbidden).
3. **Stage-1 instruction lock**: the effective charter is the world's
   `CharterPreset` constant; `skills/` is not composed. A differing `charter.md` or
   present skill files produce a notice (existing notice channel) naming the
   unlocking stage — never silence, never partial binding.
4. Stage never touches world mechanics: villagers, needs, gru, events, determinism
   identical across stages for the same seed/commands. (Cross-stage determinism diff
   is a required test.)
5. Absent `Stage` (pre-ladder worlds): no ceiling — full spec-021 behavior. Absent
   `capabilities.json` within a stage: ceiling = the grant (missing = full-grant
   semantics apply within the ceiling).

---
name: ipc-status-extensions
description: Split from [[ipc-protocol]] — StatusData's three additive omitempty wire extensions beyond the base shape: Posture (spec 039, teaching-world planner-safe rung), Horizon (spec 037, per-class suppression/calibration verdicts), and the spec-052 skin-display fields. Load for status-wire extension-field questions.
kind: concept
sources:
  - internal/ipc/protocol.go
verified_against: 657c770f87404b936a0587db1f6b00e81b9f0ee6
---

# IPC protocol — StatusData wire extensions

Split from [[ipc-protocol]]: the three additive `omitempty` fields
`StatusData` has gained since the base wire shape, each riding
`status`/`pause`/`resume`/`set_speed` alike.

Since spec 039 (US4, `contracts/posture.md` §4), `StatusData` also gains an
additive `omitempty` `Posture *PostureStatus`, present ONLY for a teaching
world with an orchestrator — unlike `Warning`, it rides `status`/`pause`/
`resume`/`set_speed` alike (the `Horizon` precedent). `PostureStatus{Rung,
Calibrated}` carries the current planner-safe ladder speed as a string
("1x"…"32x", clamped to "1x" when even 1x suppresses the planner) and whether
the serving provider is calibrated (`CalibratedAt != ""`) versus a provisional
bootstrap derivation — recomputed per reply from the planner-serving
provider's live estimate, via the identical [[cognition]] `MaxSafeSpeed` call
the boot default and the `Warning` override use, so status, boot, and the
override can never disagree ([[ipc-server]]'s `postureStatus`).

Since spec 037 (`contracts/status-horizon.md`), `StatusData` also gains an
additive `omitempty` `Horizon []HorizonClass` — unlike `Warning`, this rides
`status`/`pause`/`resume`/`set_speed` alike (any world with an orchestrator,
composed in [[ipc-server]]'s `statusDataFull`), one entry per watched class
INCLUDED at the loop's CURRENT effective speed, never an empty slice (either
absent for a no-LLM world or ≥1 entry). `HorizonClass{Class, Suppressed,
Verdict, Calibrated, SuppressedCount}` carries the class name, whether it is
suppressed right now, [[cognition]]'s `Verdict.Arithmetic` string verbatim
(clients render it, never parse it), whether its serving provider is
calibrated (calibrated classes ARE included here — contrast the `Warning`
field above, which stays gated to bootstrap-seeded providers), and the
daemon-lifetime count of router suppressions [[llm-orchestrator]] has
recorded for that class.

Since spec 052 (FR-012, contract §7), `StatusData` also gains six additive
`omitempty` skin-display fields, resolved daemon-side against the world's
boot-frozen skin (`internal/skin`, [[ipc-server]]'s `SetSkin`) so clients
render skin vocabulary without ever reading world files: `SkinName`,
`SkinEpithet`, `SkinTabLabel`, `SkinFamilyLabel` (identity strings, always
sent by a post-052 daemon — resolved against the compiled default table even
when no world skin overrides anything), and `SkinStrings map[string]string`/
`SkinStages map[string]skin.StageIdentity` (carrying only a world skin's
overrides, empty/absent otherwise). Absent fields (a pre-052 daemon) mean the
default Guardian skin — old daemons and old clients interoperate unchanged.

## Connections

Part of [[ipc-protocol]]'s summary-style split (corpus-spec v2). See
[[ipc-protocol]] for the envelopes, commands, and base `StatusData` shape
these fields extend. [[ipc-server]] composes `Posture`/`Horizon` via
`postureStatus`/`statusDataFull` and the skin fields via `SetSkin`;
[[cli-promptworld]] renders them (`postureStatusLine` and friends);
[[cognition]]'s `MaxSafeSpeed`/`Verdict.Arithmetic`/`SuppressedAt` supply the
Posture/Horizon verdicts; [[llm-orchestrator]] records the per-class
suppression counts `Horizon` exposes.

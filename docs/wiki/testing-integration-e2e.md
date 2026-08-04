---
name: testing-integration-e2e
description: In-process IPC integration suite (status/subscribe, idempotent commands, governor/calibration/horizon status-fold, large-reply handling) and the binary-level e2e suite (hermetic PROMPTWORLD_HOME, quickstart pause/crash-resume/detach scenarios). Split out of [[testing-strategy]].
kind: pattern
sources:
  - internal/ipc/ipc_test.go
verified_against: 5761edb18e2b5fb49c6a03a050b0d871f5546c05
---

# IPC integration & e2e harness

**IPC integration** (`internal/ipc/ipc_test.go`): a real loop + server + store on a
temp world. Proves: status round trip <2 s; subscribe-from-zero delivers strictly
consecutive seqs; abrupt disconnects and wire garbage leave the loop ticking;
commands are idempotent and land in the log as events; the `state` command's
coherence contract holds (no push predates the snapshot's `last_seq`, and a replica
built from it applies subsequent pushes cleanly — the [[tui-client]] pattern); and
`llm_call` routes through a live [[llm-orchestrator]] while a killed inference
endpoint leaves the loop ticking (the package's own suite covers routing, metering,
ceiling refusal, and circuit recovery against httptest mock providers). Spec 028
(adaptive throttle) adds its own status-fold coverage here: a scripted
`Governor` fake proves debt/jobs fold into `StatusData.Clock` exactly like the
LLM snapshot, a no-governor world reports zero governor values, and a
byte-shape test pins the three new fields `omitempty` (a zero status marshals
with none of them present); a `Loop.Govern`-driven test proves status reports
both the effective and requested speed while governed and that a player
`set_speed` below the governed notch collapses `RequestedSpeed` back to empty;
and a regression test pins that `set_speed max` is still refused with an LLM
configured (FR-012) while `32x` is accepted, unchanged by the governor. Spec
035 (calibration UX) adds its own `set_speed`-warning coverage here: an
uncalibrated world raised into suppressing territory (32x) warns naming the
suppressed classes and the calibrate command while the speed change still
applies, and the same world dropped to a non-suppressing notch (4x) carries
none; a calibrated world (`SeedCalibration` with a profile) gets no warning
even at a speed that would suppress at bootstrap estimates (the gate is seed
state, not raw arithmetic); a no-LLM world never warns at any speed; the
pre-035 `set_speed max` refusal still precedes the warning and carries no
`StatusData` at all; `status`/`pause`/`resume` never carry the warning even
on an uncalibrated world sitting at a suppressing speed; and a byte-shape
test pins that a zero `StatusData` omits `warning` entirely. Spec 037 (live
horizon surface) adds its own horizon-composition coverage here: an
uncalibrated 32x world's `status` reply carries one `Horizon` entry per
watched class in `WatchedClasses` order, each with a non-empty verdict
string and `Calibrated=false`, with planner/conversation suppressed and
meeting thinking at the bootstrap seeds; a calibrated world (`SeedCalibration`)
is still fully INCLUDED with `Calibrated=true` (contrast the `Warning` gate,
which excludes it) and nothing suppressed at 32x on a fast rig; composing
directly at `clock.SpeedMax` suppresses every included class with Route's
uncapped phrasing; a no-LLM world's reply carries no `horizon` key at all
(byte-identical to pre-037); and `RecordSuppression` calls surface as each
entry's `SuppressedCount`, keyed correctly per class with an unwatched class
(`chronicle`) never leaking an extra entry. A byte-shape test pins that a
zero `StatusData` omits `horizon` entirely. Spec 039 (teaching-world speed
posture) adds its own posture coverage here: an uncalibrated teaching world
set to 32x applies the speed and carries an `above teaching posture 16x`
override naming the router's verbatim suppressed-class arithmetic and the
plain-language degrade consequence for each; at or below the 16x posture the
same world carries no warning text at all; a non-teaching world never carries
the posture text (only the unchanged spec 035 uncalibrated leg); an
overshooting uncalibrated teaching world gets BOTH texts, posture first then
the calibrate prompt, newline-joined; a CALIBRATED teaching world
(`SeedCalibration`) still warns on override — using the measured seconds-per-
point in its arithmetic and carrying no uncalibrated text — proving the soft
cap is independent of calibration state; the `set_speed max` refusal still
precedes any posture text for a teaching LLM world; and on the status side, a
calibrated teaching world's `status` reply carries `Posture{Rung, Calibrated:
true}`, an uncalibrated one `Posture{Rung, Calibrated: false}`, a non-teaching
world's reply omits the `posture` key entirely (byte-shape pinned on a zero
`StatusData` too), and a teaching-but-pure-sim world (no orchestrator) also
carries no posture block. Large-reply
behavior (TASK-19) is proven against a `fakeDaemon` wire harness that speaks the
protocol from canned replies: a >1 MiB `state` payload round-trips; a reply over
the 64 MiB cap is substituted server-side with an actionable `reply too large`
error (via `net.Pipe` against `session.writeResponse`); and both the substituted
error and a raw over-long line surface promptly as `ErrReplyTooLarge` — never a
hang or silent scanner death.

**E2E** (`e2e/`): `TestMain` builds the binary once and sets a package-wide
hermetic `PROMPTWORLD_HOME` (a temp dir) before running — every subprocess the
package execs inherits it, so no test can write the developer's real
`~/.promptworld` registry (TASK-49; `manager_e2e_test.go`'s `isolatedHome`
layers a per-test override on top). Worlds drop `llm.json`
right after `new` so they are pure-sim — a precondition for `speed max` under
the TASK-20 policy. Scenarios mirror
`specs/001-world-daemon/quickstart.md` — A: always-on + detach-is-not-pause; B:
pause freezes the clock, compression ratios hold (loose tolerances over short
windows; the spec's 5% applies to 5-minute windows); C: kill -9 → lossless resume
within 10 s, restart-while-paused wakes paused, graceful stop idempotent; E: a
`cp -R`'d stopped world runs. `determinism_e2e_test.go` compares two same-seed
daemons' sim histories over their common tick prefix (past tick 25000, so the
full day-1 [[governance]] meeting cycle is inside the compared window),
excluding wall-dependent `daemon.*`/`clock.*` bookkeeping.

## Connections

Part of the [[testing-strategy]] suite map (split out during the corpus-spec v2
restructure); see that note for the full layered test picture and links to
sibling suites.

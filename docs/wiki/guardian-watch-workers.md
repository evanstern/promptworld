---
name: guardian-watch-workers
description: The guardian's own-initiative and background machinery — standing orders (monitor_and_act/cancel_order) from the turn's side, the digest worker's summarization/moment windows, the report-card worker, and the four-goroutine shutdown discipline (absorb/digest/trigger/report-card) Close drains. Split from [[guardian]]; full standing-order lifecycle lives in [[guardian-orders]].
kind: component
sources:
  - internal/guardian/orders.go
  - internal/guardian/digest.go
  - internal/guardian/guardian.go
  - internal/guardian/toolcalls.go
verified_against: 74fe956813aa6be54e65156ae9bfcb91745cbb8d
---

# Guardian's watch and background workers

Split from [[guardian]] (summary-style, corpus-spec v2) — standing orders
from the turn's side, the digest/report-card background workers, and the
shutdown discipline that drains them.

## How it works

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

**Watching** (`digest.go`): notable events collect per 6-game-hour window; each
non-empty window costs one summarization call appended to `metatron/soul.md`
(skip-empty is free; failures carry lines into the next window). The drama rule v1:
`agent.died`, `gru.attacked`, `social.promise_broken`, and (since spec 029)
`metatron.order_expired` append model-free **moment** lines immediately and queue for
the console — the next reply leads with them. Digests and moments themselves never
construct an act; the guardian acts only when the player asks OR a standing order the
player placed authorizes it (spec 029 relaxed the old "acts only when told" contract
to admit pre-authorized triggered turns — see [[guardian-orders]]).

**Report cards** (spec 063 US4, `reportcard.go`): a sibling stopping-point
consumer on the SAME digest-worker notify pattern — see
[[grounded-feedback]] for the full producer (activity trail, stopping-point
triggers, the cheap-chain critique, citation validation) and
[[takeover-surfaces]] for the shared rubric-checklist renderer the console
card seam and the postmortem/ceremony takeovers compose it beside.

**Shutdown** (`guardian.go`): `New` starts four background goroutines — the absorb
loop (`run`), `digestWorker`, `triggerWorker`, and `reportCardWorker` — each
counted into a `sync.WaitGroup` at its spawn site. `Close` closes the shared
`done` channel the workers select on, then waits on that WaitGroup, so `Close`
cannot return until every one of the four has exited; a `Close` racing an
in-flight job blocks until that job's current iteration finishes (no timeout —
shutdown correctness over speed). This is why a caller (production, or a test
fixture) that drives `cardQ`/`digQ`/`triggerQ` right after `Close` never races a
worker still parking in its select.

## Connections

[[guardian]] is the parent. [[guardian-orders]] (spec 029) owns the full
standing-order lifecycle, event sourcing, matching, trigger execution,
fuzzy confirm, daytime-omen deferral, and the three system-origin survival
watches this note's workers ultimately serve. [[grounded-feedback]] (spec
063) owns the report card's producer (activity trail, stopping-point
triggers, cheap-chain critique, citation validation) this note's worker
consumes from; [[takeover-surfaces]] is its TUI-side rubric-checklist
consumer. [[guardian-turn-loop]] documents the same-door telemetry a
triggered system turn emits identically to a console turn.

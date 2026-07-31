---
name: world-forking
description: "The fork ceremony and the duel (spec 076): world.Fork builds a fresh world.db — parent event prefix + boundary snapshot + world.forked lineage event + meta (seed carried, llm_spend_* wallet inherited) — under a fresh name/dir/socket identity; promptworld fork/compare are the CLI doors; compare renders the rubric-first duel scoreboard through the exported spec-072 resolver plus story-event divergence and interleaved chronicles. Forks never merge."
kind: component
sources:
  - internal/world/fork.go
  - internal/sim/state.go
  - cmd/promptworld/fork.go
  - cmd/promptworld/compare.go
verified_against: fc1a8314f3f71a33c5e2145c914d5cbb511d9196
---

# World forking & the duel

Spec 076's iteration rung: replay is model-free (LLM output enters the world
only as recorded input — [[chronicle]]), so a player cannot re-run yesterday
under a new prompt. Forking is the substrate's cheap alternative: copy the
village at its latest snapshot under a fresh identity, edit the fork's
charter, run BOTH worlds side by side, and compare the outcomes. Divergence
between two same-seed forks is attributable to the actual input difference
by construction ([[deterministic-rng]]) — a controlled experiment.

## The fork ceremony (`world.Fork`, `internal/world/fork.go`)

`Fork(srcDir, destDir, newName)` is the `Migrate` ceremony's sibling
([[world-save-directory]]), a **fresh-store ceremony, never a file copy**:
the events table is append-only in-schema ([[event-log]]'s
`events_no_update`/`events_no_delete` triggers), so "truncated to the
snapshot boundary" is built, not carved. Steps:

1. `Open` the source (current-format gate — migrate first); refuse a live
   daemon (the `Migrate` posture: sidecar copies race a running daemon).
2. Boundary = `Store.LatestValidSnapshot()` — the same newest→oldest
   hash-verified walk recovery uses ([[snapshots]]). No valid snapshot →
   refuse with the start-and-stop remedy. v1 forks at the latest snapshot
   only (`--at latest-snapshot` is the sole accepted value).
3. Destination must be empty/absent (the `Create` posture); on any later
   error the partial destination is best-effort removed.
4. Fresh `world.db`: the log-format stamp is written first (spec 094,
   [[event-log]] — a fork's log is born current; the parent passed the Open
   gate so its prefix already speaks this build's vocabulary), then the
   parent's events with `seq <= boundary.seq` are
   streamed in order into `AppendEvents` (contiguous seqs 1..N reproduce by
   construction), the boundary snapshot is written verbatim (hash
   re-verified against the parent's `state_hash`), one `world.forked`
   lineage event is appended at `(boundary.tick, boundary.seq+1)`, and meta
   is stamped: `seed` + `format_version` (what `validateMeta` checks at
   first boot, [[daemon-lifecycle]]) plus every `llm_spend_*` key
   (`Store.MetaByPrefix` — see the wallet below).
5. The fork manifest: `name` = the new name, `created_at` = fork wall time,
   a new `lineage` block — and **every other field verbatim, seed
   included**: the carried prefix was generated under the parent seed and
   `sim.rngAt` keys off it; a fork's identity is its name, directory,
   socket, and registry presence, never its seed
   ([[world-save-manifest-fields]]).
6. Sidecars copy as-of fork time: `llm.json`, `calibration.json`,
   `estimator_state.json`, `charter.md`, `tuning.json`, `metatron/`,
   `bundles/`, `agents/` contents. NOT copied: runtime files
   (socket/pidfile/log), the parent db + WAL sidecars, migration archives
   (`world.v*.db`), and the scribe's regenerable views (`chronicle.md`,
   `morgue.md`, `village_charter.md` — regenerated at the fork's first boot;
   copying them would ship prose about truncated-away events).

`ForkResult` carries the CLI summary (boundary tick/seq, truncated tail,
ended-boundary warning, spend-carried flag). Determinism is proven, not
assumed: genesis replay of the fork's log reproduces the boundary snapshot's
hash, which IS the parent's state hash at the same (tick, seq) — the
`world.forked` reducer arm is a recorded-history no-op
([[sim-state-reducer]]), so fork state at the fork tick is byte-identical to
the parent's.

**Lineage** is durably recorded twice: the `world.forked` event
(`WorldForkedPayload{parent_name, parent_seed, parent_created_at, fork_tick,
fork_seq}`, [[event-types-clock-world]]) is authoritative; the manifest's
additive `omitempty` `lineage` block (`{parent, parent_created_at,
fork_tick}`) is the fast offline mirror compare's default window reads. No
format bump — a lineage-less `world.json` round-trips byte-identically.

**The wallet is inherited, never re-minted** (board AC #5, research R4):
`llm.json` copies verbatim (same `monthly_budget_usd` ceiling) and every
`llm_spend_*` meta key (totals + per-provider attribution) copies into the
fresh store, so the fork's meter ([[llm-budget-degraded-mode]]) opens at the
parent's month/spend/ceiling. Thereafter each world meters independently —
recorded limitation: a duel's combined forward spend can exceed one ceiling
by up to the unspent remainder at fork time.

## The duel (`promptworld compare <a> <b> [--since TICK]`)

Offline read per side — `worlds.OfflineState` ([[instance-manager]]),
snapshot + fold, WAL-safe on a running world (header says
"as of its last committed batch"). The scoreboard derives per-term verdicts
through the ONE spec-072 resolver, exported replica-parametric as
`tui.ResolveRubricFacts`/`tui.RecordedPassFor` and rendered through
`tui.RenderReportCard` ([[report-card-renderer]]) — the duel is that
resolver's fourth consumer; a second precedence switch anywhere is a spec
violation. Outcomes render in plain language (`in_progress` → "still
running", `failed` → "did not make it through" — the postmortem's no-blame
register); a lost duel IS the concluded-✗ postmortem card. An ambient world
gets an honest no-scorecard note, never an invented card.

The comparison window defaults to the fork tick when either side's lineage
names the other (manifest mirror, event fallback); `--since` overrides.
**Divergence** compares the post-window story-event streams over
`(tick, type, payload)` — excluding `wall_time`, the machinery classes
`daemon.*`/`clock.*`/`cog.*`/`llm.*` (wall-dependent telemetry — the
determinism e2e's exclusion, extended), `chronicle.entry` (narrated wording
differs between runs of the same story), and `world.forked` itself — through
the two runs' common tick horizon. Zero divergence renders the honest
"identical since the fork" line: a truthful, teachable outcome.
`chronicle.entry` events interleave below, merged by `from_tick`, labeled
per world, with the divergence marker in timeline position.

## Connections

[[world-save-directory]] (the ceremony's home, beside Create/Open/Migrate);
[[event-log]] + [[snapshots]] (prefix stream + boundary);
[[event-types-clock-world]] (`world.forked` row); [[sim-state-reducer]] (the
no-op arm); [[report-card-renderer]] (the exported resolver);
[[llm-budget-degraded-mode]] (the inherited wallet); [[cli-world-lifecycle]]
(the two subcommands); [[instance-manager]] (`OfflineState`, registry
posture: fork writes no registry state — an outside-home fork self-registers
at its first daemon boot); [[deterministic-rng]] (why the seed carries).

## Operational notes

**Forks are independent worlds forever** — no merge verb exists or will
(spec 076 FR-014, doctrine). Fork refuses a running source; compare is
read-only and works on running worlds. Forking an ended world is legal but
warned (the boundary carries the ended state — the fork is born ended);
mid-log / chosen-snapshot forking (`--at <tick>`) is a documented follow-on,
as is the shareable HTML retelling that consumes compare's report model.

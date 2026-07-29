---
name: event-log
description: Append-only SQLite events table — the source of truth for a world run; contiguous seq, in-schema immutability triggers, WAL; since spec 094 the log carries its own format-version stamp (meta log_format_version) with load-time enforcement and the rename-requires-migration doctrine
kind: component
sources:
  - internal/store/store.go
  - internal/store/schema.go
  - internal/store/format.go
verified_against: 72f82f41f7aa2e345572105894cd0fb7c02fc0aa
---

# Event log

The `events` table in `world.db` is a world run's authoritative history: every
simulation event and every applied command lands here, in order, and can never be
modified. World state is derived from it (event sourcing); snapshots merely accelerate
that derivation.

## How it works

`internal/store.Store` opens SQLite via the pure-Go `modernc.org/sqlite` driver with
`journal_mode=WAL`, `synchronous=NORMAL`, and `SetMaxOpenConns(1)` — the sim loop is
the single writer, and one connection sidesteps SQLITE_BUSY entirely.

Schema (in `schema.go`): `events(seq INTEGER PRIMARY KEY, tick, type, payload,
wall_time)` with indexes on `tick` and `type`. Two triggers, `events_no_update` and
`events_no_delete`, `RAISE(ABORT, 'events is append-only')` — immutability is enforced
in-schema, not by convention.

- `AppendEvents(events)` assigns contiguous seqs (`lastSeq+1…`) and writes the batch in
  one transaction — one batch per tick, so no event is visible to subscribers before
  its tick commits. `lastSeq` is an `atomic.Int64` because the loop writes it while IPC
  sessions read it.
- `ReplayEvents(sinceSeq, fn)` / `EventsSince(sinceSeq, limit)` stream history in seq
  order — used by recovery, subscribe-replay, and `tail`.
- `CheckContiguity()` verifies seq runs exactly 1..N; a gap is a fatal integrity error
  and [[daemon-lifecycle]] refuses to run on a holed log.
- `Event.Payload` is canonical JSON (struct-marshaled, fixed field order) so histories
  are byte-comparable; `wall_time` is observability metadata, excluded from determinism
  comparisons.
- `MetaByPrefix(prefix)` (spec 076) returns every meta row whose key starts
  with the prefix — filtered in Go rather than SQL `LIKE` (the callers'
  prefixes carry underscores, which `LIKE` treats as wildcards, and the meta
  table is tiny). Consumers: the fork ceremony's wallet inheritance (every
  `llm_spend_*` key copies verbatim into a fork's fresh store —
  [[world-forking]], [[llm-budget-degraded-mode]]) and the translating
  migration's verbatim meta carry-over ([[world-migration]]).

## The log format stamp (spec 094)

The log is self-describing independent of its manifest: one `meta` row, key
`log_format_version` (`format.go`), written at genesis by every log creator
(`promptworld new`, `world.Fork`, both `world.Migrate` modes) and readable
without replay. `LogFormatLegacy` (1) is the implicit version of every
pre-stamp log — the `metatron.*` guardian vocabulary; `LogFormatVersion`
(2) is the guardian-rename break (spec 094): the 13 persisted `metatron.*`
event types became `guardian.*`. `VerifyLogFormat()` is the load gate,
called by daemon boot (`validateMeta`, [[daemon-boot-recovery]]) before any
replay: an OLDER log refuses with the migrate hint, a NEWER one with the
upgrade posture (a future vocabulary must never be mis-replayed under old
reducer arms), and an unstamped EMPTY log is adopted and stamped. The
world-manifest `format_version` gate ([[world-save-directory]]) carries the
same refusal at every CLI surface; the log stamp is the DB-side defense in
depth — a hand-edited manifest cannot smuggle an untranslated log past the
reducer.

**DOCTRINE (spec 094, reconciled with spec 092's emitter-computes/
determinism doctrine — TASK-75):** changing any persisted event-type NAME,
or changing how the reducer RE-DERIVES state from a recorded payload
(spec 092's audit names the hunt-yield re-derivation as the canonical
example), REQUIRES bumping `store.LogFormatVersion` and shipping a
migration. The decision rule: a **pure rename translates** — the type
column is rewritten through `sim.LogFormatV1Renames` with every
seq/tick/payload preserved byte-for-byte ([[world-migration]]'s translating
mode); a **semantic break snapshot-cuts** — fresh log carrying the
transformed state. Never alias at read: the gate exists so untranslated
logs are refused, not reinterpreted. Persisted REFERENCE strings inside
preserved payloads (curriculum `EvidenceRef.Type`) keep their historical
names forever; their readers normalize through `sim.CanonicalEventType`.

## Connections

[[sim-loop]] is the only writer; [[sim-state-reducer]] consumes events in replay;
[[snapshots]] bound how much of the log recovery must re-read; [[ipc-server]] reads it
for subscribe-replay and gap-fill; [[event-types]] catalogs what lands in it.
[[world-forking]]'s ceremony is why the in-schema triggers matter beyond
convention: a fork can never truncate a copied db, so it streams the parent's
event prefix into a fresh log instead.

## Operational notes

Event volume at v1 scale is trivial for SQLite (<1M rows per 30-day run). The meta
table (same file) stores `seed`/`format_version` (manifest mirrors) and
`log_format_version` (the log's own stamp) — three DISTINCT versions by design:
world.json `format_version` (save-directory shape), the log stamp (event
vocabulary), and the store DDL (table shape, unchanged since format 1).
Chronicle narration and guardian digests query this table by `tick`/`type` —
the indexes exist for them.

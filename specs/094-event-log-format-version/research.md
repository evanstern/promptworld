# Research: log stamp shape, manifest bump, and the frozen-type enumeration (T001)

**Date**: 2026-07-29 | **Branch**: `task-134-event-log-format-version`
**Grounded against**: this branch's base (main @ d2f73eb) — internal/store,
internal/world/{world,migrate,fork}.go, internal/sim, the spec-052 freeze
annotations, and the fiction-denylist sweep's frozen-forms rules.

## D1 — Stamp shape: a `meta` table row (`log_format_version`), not a header row, meta event, or pragma

The log-level format-version stamp is one row in the store's existing `meta`
table: key `log_format_version`, value the integer as a string. Constants live
in `internal/store` (`LogFormatLegacy = 1`, `LogFormatVersion = 2`) with the
world.go-pattern version-history comment.

Rationale:

- **Readable without replay** (FR-001): a single `SELECT` against a table that
  has existed since format 1 — no schema change, no DDL migration, and a DB
  produced by any older binary tolerates the unknown key.
- **Precedent**: `validateMeta` (internal/daemon/daemon.go) already mirrors the
  manifest's `seed`/`format_version` into `meta` for cross-checking. The log
  stamp is the same idiom with the log itself as the described artifact.
- **Travels with the artifact**: the stamp lives inside world.db, so archives
  (`world.v5.db`), forks, and a DB separated from its manifest stay
  self-describing — the card's "a log is not self-describing" gap.
- **Meta event rejected**: a synthetic event would occupy seq-space, polluting
  replay and making the translating migration's byte-identity claim
  (FR-004: same seq/tick/payload for every event) impossible for stamped logs.
- **Header row / sqlite `user_version` pragma rejected**: the pragma is
  invisible to `SELECT`-level tooling and easy to drop in copy ceremonies; a
  bespoke header table is the meta table reinvented.
- **Append paths need nothing** (card AC#1's conditional): one row describes
  the whole single-file log; there are no segments.

Absent-key semantics: absent + `events` non-empty ⇒ implicit **legacy
(log format 1)**; absent + empty ⇒ an unstamped fresh log (only reachable
mid-genesis; genesis writers stamp before appending). Enforcement is two-sided
(FR-002): stamp < binary ⇒ refuse with the migrate hint; stamp > binary ⇒
refuse with the upgrade posture (the world.go FormatVersion posture: a future
vocabulary must never be mis-replayed).

Stamp writers: `promptworld new` (genesis), `world.Fork` (fresh boundary log),
`world.Migrate` (both modes' fresh/translated logs). Verifier: daemon boot
(`validateMeta`, before contiguity/recovery). Read-only offline tools that
deliberately open legacy DBs (`replayToTick` against archives, `worlds/probe`
listing) keep their gate-free posture — documented below in D2.

## D2 — World-manifest FormatVersion bumps alongside: 5 → 6

Decision: **bump**. `world.Open`'s FormatVersion gate is the single choke
point every world-content command already routes through; bumping to 6 buys
US3.4's refusal (unmigrated pre-rename world ⇒ migrate hint) at every surface
for free, on the well-tested `ErrFormatVersionMismatch` path that TASK-147's
daemon-lifecycle carve-out already understands. The spec-068 precedent (v5 was
a manifest-only bump whose entire point was the refusal gate) establishes the
manifest as the refusal surface of record.

The log stamp is NOT redundant alongside it: it is (a) the DB-side
self-description for a DB separated from its manifest, (b) the translating
migration's own witness of what vocabulary a log speaks, and (c) defense in
depth at daemon boot when a manifest lies (hand-edited) — `validateMeta`
refuses on the stamp even when the manifest says 6.

`ErrFormatVersionMismatch.Error()` gains the direction split: older ⇒ the
existing migrate hint; newer ⇒ "upgrade promptworld" (the world.go:51
posture), satisfying US1.3/SC-003 at the manifest level too.

Version map after this feature:

| artifact | old | new | break |
|---|---|---|---|
| world.json `format_version` | 5 | **6** | guardian rename (log translation) |
| log stamp `log_format_version` | 1 (implicit) | **2** | metatron.* → guardian.* type names |
| store DDL | "format_version 1 schema" comment | reworded | none (schema unchanged; terminology reconciled per spec edge case) |

Migration modes by source (the decision rule, recorded as doctrine in T009):

- **v1–v3 → 6**: the existing snapshot-cut chain, unchanged — its fresh log
  (world.created + world.migrated) already speaks the current vocabulary, so
  it is stamped log format 2 directly.
- **v4/v5 → 6**: the NEW translating mode. v4→v5 was manifest-only (spec 068),
  so v4 and v5 logs are content-identical: both translate. Type column mapped
  through the rename table; **seq, tick, payload, wall_time preserved
  byte-for-byte**; snapshots and meta rows copied verbatim (then the
  `format_version` mirror updated and the log stamp written). Archive
  `world.v4.db`/`world.v5.db` keyed to source format; archive existence is the
  already-migrated guard; never-overwrite; live-daemon refusal; the translated
  DB is built at a temp path and swapped in only after the source is archived,
  so no failure mode leaves a half-written world.db as the live log.

## D3 — Exact enumeration of the renamed persisted types (13)

From `sim.PayloadCatalog` (the single enumerable truth, spec 086) crossed with
the spec-052 freeze annotations and a `"metatron.` grep over emit/apply paths.
Exactly 13 — matching the card's "at least 13":

| # | log format 1 (frozen by spec 052) | log format 2 |
|---|---|---|
| 1 | metatron.charge_regenerated | guardian.charge_regenerated |
| 2 | metatron.nudged | guardian.nudged |
| 3 | metatron.place_revealed | guardian.place_revealed |
| 4 | metatron.order_placed | guardian.order_placed |
| 5 | metatron.order_triggered | guardian.order_triggered |
| 6 | metatron.order_cancelled | guardian.order_cancelled |
| 7 | metatron.order_expired | guardian.order_expired |
| 8 | metatron.charter_observed | guardian.charter_observed |
| 9 | metatron.skills_observed | guardian.skills_observed |
| 10 | metatron.time_snapped | guardian.time_snapped |
| 11 | metatron.item_granted | guardian.item_granted |
| 12 | metatron.entity_moved | guardian.entity_moved |
| 13 | metatron.entity_removed | guardian.entity_removed |

The target namespace already exists (`guardian.report_card`, spec 063 — new
vocabulary from birth), so no grammar-family or catalog plumbing is novel.

**Deliberately NOT renamed** (frozen non-event-type surfaces; spec 052 ruling 2
stands for them — each keeps its annotation):

- the on-disk `metatron/` directory and `metatron/soul.md` (world.go
  `GuardianDir`);
- llm.json route kinds `"metatron"` / `"metatron_watch"` (internal/llm);
- the `ps` JSON key `"metatron_charges"` (cmd/promptworld/ps.go);
- the hidden CLI compat alias `promptworld metatron` (main.go);
- `*-metatron-*` correlation-id prefixes recorded in cog.* payloads;
- recorded payload TEXT: survival-watch Condition/Action strings
  (sim.SurvivalWatchDefs) and any string already in an existing log.

## D4 — Persisted payload REFERENCES keep legacy names; readers normalize

Finding: `EvidenceRef.Type` (curriculum pass evidence) embeds event-type
strings **inside persisted payloads** (`curriculum.exercise_passed`) and hence
inside snapshot state. Translation preserves payloads byte-for-byte (that IS
FR-004 — rewriting them would change the replayed state and break the
state-hash identity), so a migrated world's recorded evidence keeps
`metatron.*` strings forever, while post-rename passes record `guardian.*`.

Consequence: the (only) production reads of a persisted evidence Type string —
`sim.EvaluateUnlock`'s stage-2 conjunct, the TUI charter-lineage scan
(views.go), and report-card fact matching — normalize through the v1→v2 rename
table before comparing. This is NOT the aliasing the operator ruled out: the
log's type column is never aliased at read (enforcement guarantees only
translated logs are opened); normalization applies solely to historical
reference strings that byte-identity obliges us to preserve. Recorded as part
of the T009 doctrine.

## D5 — recipes_test.go:75-76 value pin (spec edge case)

The pin (huntYieldBare/huntYieldSpear et al.) guards replay-relevant apply-time
constants — the card's "reducer-re-derivation" hazard class. It is superseded
by asserting THROUGH the versioning: the test now states the ratified log
format these values belong to (`store.LogFormatVersion`) and fails a retune
with the doctrine (bump + migration) rather than as a bare number mismatch.
The state.go hunt-yield re-derivation itself is spec 092's documented audit
item and changes nothing here.

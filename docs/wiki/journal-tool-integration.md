---
name: journal-tool-integration
description: How the four journal roster tools wire into the mind's handlers, the per-cognition journal snapshot search/read run against, spec-043's term-matched decision-context excerpts, and the scribe's regenerable journal.md view. Split from [[agent-journal]] (Mind handlers, The snapshot, Journal excerpts, and The view sections).
kind: component
sources:
  - internal/sim/journal.go
  - internal/mind/handlers.go
  - internal/scribe/scribe.go
  - internal/persona/files.go
verified_against: c5e66ee92fa75c00e2b480811e0ca727d5c1a1e1
---

# Journal tool integration

**Mind handlers** (`internal/mind/handlers.go`): `villagerHandlers` wires all
four by name. The two writes (`handleWriteJournal`/`handleDeleteJournal`) mirror
`handleMuse`: marshal the `journal.*` event and land it through `InjectSocial`
batched atomically with a `cog.outcome{landed}`. `journalDoorResult` translates
the door result — success sets `doorOutcome` and returns `VerdictLanded`; a door
rejection is `errors.Unwrap`-peeled so the model sees the gate's reason verbatim
as `VerdictRejectedGate` (the agent can curate an over-budget journal and retry);
a non-wrapped error surfaces as `Err` (infrastructure failure → the loop
terminates and `runPlan` records the FR-015 outcome). `handleWriteJournal`
guards the empty-text case and defensively re-caps at `JournalWriteCapRunes`;
`handleDeleteJournal` needs an `entry` id (`argInt`, float-tolerant). The two
reads run over a **per-cognition journal snapshot** — `handleSearchJournal`
calls `d.job.journal.SearchJournal(query)` (case-insensitive substring,
newest-first, capped at `JournalSearchResultCap`; zero matches is a well-formed
empty `read_ok`, never an error), and `handleReadJournal` addresses one entry by
id (`FindJournalEntry`, unknown → `read_error`) or returns the whole journal
oldest-first (`JournalEntries`). `formatJournalEntries` renders
"#<id> <clock>: <text>", one per line. No read parameter can address another
agent — a handler reads only this job's own journal.

**The snapshot** (`internal/mind/mind.go`): `plan()` sets
`job.journal = a.Journal.Clone()` for each due agent — a race-free deep copy
(`JournalEntry` holds no pointers, so copying the slice suffices; nil-safe). The
search/read handlers run in the planner worker goroutine and must NOT touch the
absorb-owned replica, so they read the immutable snapshot; writes and deletes
land through the live `InjectSocial` door, not the snapshot.

**Journal excerpts in the decision context** (spec 043 US4, `journal.go`):
the context assembler ([[decision-context]]) can stuff term-matched journal
excerpts into a villager's planner prompt so the agent need not spend its own
reasoning turns fetching them. `SelectJournalExcerpts(terms)` is the
deterministic, model-free retrieval behind it — each situation term runs
through the SAME `SearchJournal` substring match the roster tool exposes (no
embeddings; embed-at-write was deferred per research R5); the union is
deduped by id, ordered newest-first (tick desc, id desc tiebreak), capped at
`journalExcerptCap` (2) entries of `journalExcerptRunes` (300) runes each
(`JournalExcerptCap`/`JournalExcerptRunes` export the numbers). Excerpting
truncates on a rune boundary and marks the cut with an ellipsis — never
fabricated text — and each excerpt carries its entry id so the villager can
follow up with `read_journal`. In the assembled context this is the
lowest-priority block: first dropped under the size budget. Constants, not
config — same journal + terms always yields the same excerpts.

**The view** (`internal/scribe/scribe.go`, `internal/persona/files.go`):
`JournalPath(worldDir, name)` is `agents/<name>/journal.md`, a regenerable view
(like `soul.md`). `Genesis` seeds it empty at world creation — a
"# <name>'s journal … _0/4000 runes_ … *Empty — nothing written yet.*" stub.
The scribe re-renders it on every `journal.entry_written`/`journal.entry_deleted`
via `renderJournal`, tracked in a `jDirty` set kept SEPARATE from the soul
`dirty` set (a journal mutation touches only that one file — souls are
unaffected). `renderJournal` writes a header with current budget usage
(`JournalUsedRunes()` / `JournalBudgetRunes`) then each entry verbatim under a
"## <clock> (#<id>)" section — the agent-authored markdown is the artifact under
study, so the scribe adds no normalization, only the id/clock chrome delete and
read address by. (The same scribe's other whole-file regenerable views —
`chronicle.md`, `village_charter.md`, and since spec 044 the [[morgue]]'s
`morgue.md` — follow the same events-are-truth doctrine.)

## Connections

[[agent-journal]] is the parent note this child was split from — it owns the
one rune-budget rule and the two reducer arms these handlers land through;
[[agent-mind]] hosts `villagerHandlers` and the planner worker goroutine that
snapshots `job.journal` before dispatch; [[decision-context]] owns the
`journal` block (block 10) that renders `SelectJournalExcerpts`'s output into
the planner prompt; [[tool-registry]] declares the four tools' schemas and
glosses; [[tool-loop]] is the driver that dispatches them; [[sim-state-reducer]]
owns the reducer arms these writes land through.


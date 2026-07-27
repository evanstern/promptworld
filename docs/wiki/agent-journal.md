---
name: agent-journal
description: The agent-authored journal (spec 019, US3) - per-villager markdown notebook in durable world state, mutated only through two reducer arms under one hard rune budget (journal.go/state.go), the injection-door whitelist, the four roster tools, and guidance-free glosses; see journal-tool-integration for the mind handlers, per-cognition snapshot, decision-context excerpting, and the rendered view.
kind: component
sources:
  - internal/sim/journal.go
  - internal/sim/state.go
  - internal/sim/loop.go
  - internal/tool/registry.go
  - internal/tool/roster.go
verified_against: c61cd6c04ddfcd2a976c14a49ba071e8fd768a73
---

# Agent journal

Spec 019's Layer 2 (TASK-16, US3): a per-villager self-authored markdown
notebook that is part of durable, deterministic world state. Unlike the soul —
which the scribe writes ABOUT the agent from its memories — the journal is what
the agent chooses to write for itself. The design bet is deliberately minimal:
impose exactly ONE rule (a hard size budget) and let how villagers actually use
the space be the observed experiment, not something the prompt steers.

## How it works

**State** (`internal/sim/journal.go`, `agents.go`): each agent carries
`Journal *Journal` (`omitempty`, the `Hail` pointer precedent — an agent that
never journals stays byte-identical to a pre-019 snapshot, since `encoding/json`
`omitempty` is a no-op on a non-pointer struct and would otherwise serialize
`"journal":{}` and break the pre-019 round-trip). A `Journal` is
`{NextID, Entries []JournalEntry}`; a `JournalEntry` is `{ID, Tick, Text}`.
Entries are append-ordered (append order == event order == tick order); ids are
reducer-assigned, monotonic, and NEVER reused (delete never renumbers, `NextID`
only ever grows), giving delete/read a stable address across the journal's whole
life. `appendEntry`/`deleteEntry` are the only mutators, both unexported — the
journal changes ONLY through the two reducer arms below.

**The one rule — the rune budget, enforced in the reducer** (`journal.go`,
`state.go`): `journalBudgetRunes` (4000) is the total per-agent budget;
`journalWriteCapRunes` (1000) is a per-write wire sanity cap (a quarter of the
whole, same kind as say/muse text caps — NOT a usage rule);
`journalSearchResultCap` (8) bounds a search return. These are sim CONSTANTS,
not config: the budget participates in `Apply`'s accept/reject decision, so it
must be version-stable like every other deterministic sim constant — a config
knob that drifted between runs could let a replay reject an event that landed
live. `appendEntry` returns an error when `used + need > journalBudgetRunes`,
and the budget lives in the **reducer's `Apply` arm**, not in the handler: the
`journal.entry_written` arm calls `appendEntry`, whose error the `InjectSocial`
dry-run turns into a door rejection, so no over-budget event ever lands (SC-005,
Principle III — the gate decides, never handler courtesy). `deleteEntry`
similarly errors on an unknown id ("no journal entry #<id>"), and freed runes
become immediately available. The exported mirrors `JournalBudgetRunes`,
`JournalWriteCapRunes`, `JournalSearchResultCap` let the mind-side gloss and
handler reason strings quote the exact number the gate enforces.

**The reducer arms** (`internal/sim/state.go`): `journal.entry_written` is the
ONLY journal-growth path — it unmarshals `JournalWrittenPayload{Agent, Text}`,
lazily allocates `a.Journal`, and `return`s `appendEntry(e.Tick, text)` (the
reducer stamps the tick and assigns the id; the payload carries neither).
`journal.entry_deleted` unmarshals `JournalDeletedPayload{Agent, Entry}` and
`return`s `deleteEntry(id)` (a nil journal errors, since there is nothing to
delete from). Both are dry-run-checked at the door before commit.

**The injection door** (`internal/sim/loop.go`): the two journal event types are
on `injectSocialWhitelist` (`journal.entry_written`, `journal.entry_deleted`) —
the whitelist IS the isolation boundary for `InjectSocial`, and
`sim.ValidateToolCoverage` pins the two Expressive journal tools' declared
`Events` ⊆ this whitelist at boot, so a tool can never name an event the door
would refuse. `InjectSocial` applies the batch atomically after a dry-run: an
over-budget write or unknown-id delete is rejected as a whole batch (the
paired `cog.outcome` never lands either), and the reducer's error text is what
the model is told.

**The four roster tools** (`internal/tool/registry.go`, `roster.go`): spec 019
adds `journalTools`, appended to `LoopRosterVillager()` AFTER `muse` (so no
existing tool's position shifts) and villager-only — the guardian never sees
them, journals are private. Two are acting, two are reads:

- `write_journal_entry` — `Effect: Expressive`, `Gate: None`, one required
  `text` param (`MaxRunes` 1000, `Cost.TextCapRunes` 1000), `Events:
  [journal.entry_written]`.
- `delete_from_journal` — `Expressive`, `Gate: None`, required `entry` number,
  `Events: [journal.entry_deleted]`.
- `search_journal` — `Effect: Read`, required `query` text (`MaxRunes` 200), no
  events (a read grounds nothing).
- `read_journal` — `Effect: Read`, OPTIONAL `entry` number (present → one entry;
  absent → the whole journal).

The two `Read` tools are the **first production Read tools** on the villager
loop roster: spec 017 built generic Read dispatch and lifted the roster
restriction but shipped no production reads, so the loop itself needs no change
to carry these. `Gate: None` on the writes is deliberate — a journal write
needs no scene and no charge; the reducer dry-run (budget / existence) is its
ONLY gate.

**Guidance-free glosses — the one-rule doctrine** (`registry.go`): the
`PromptGloss` strings describe capability + budget and NOTHING else. They name
what the tool does and the 4000/1000 numbers (mirrored literals — `internal/tool`
is a leaf and cannot import `sim`), but deliberately carry no when/why/how-to-
journal guidance: usage is the observed experiment, so a reviewer checks these
for smuggled behavioral steering. The whole point is that the code imposes the
budget and only the budget; whether villagers keep a diary, a to-do list, or
nothing is emergent, not prompted.

**Journal tool integration and the rendered view** ([[journal-tool-integration]]):
the mind's handlers wire all four roster tools by name and mirror the
`handleMuse` landing pattern for the two writes; the two reads run over a
per-cognition journal snapshot the planner worker clones before dispatch;
spec 043 (US4) adds `SelectJournalExcerpts`, the deterministic term-matched
retrieval the decision-context assembler stuffs into a planner prompt; and
the scribe re-renders `journal.md`, a whole-file regenerable view, on every
`journal.entry_written`/`journal.entry_deleted` event.

**Replay determinism**: nothing about the journal touches a model — search is
deterministic substring matching over the agent's own entries, the budget is a
version-stable constant checked in `Apply`, and ids/ticks are reducer-assigned
from the event stream. The `journal.*` events are the whole truth; the
`journal.md` files are regenerable views, so live and replay agree byte-for-byte.

## Connections

[[agent-mind]] hosts the villager tool-use loop and the handlers that read and
write the journal, and seeds/renders its files through the same persona and
scribe machinery that carry the soul; [[tool-registry]] declares the four tools,
their glosses, and the `ValidateToolCoverage` Events⊆whitelist gate; [[tool-loop]]
is the driver that dispatches them (the journal reads are its first production
Read tools); [[sim-state-reducer]] owns the two `journal.*` arms and the budget
that gates them; [[event-types]] catalogs `journal.entry_written` /
`journal.entry_deleted`; [[social-fabric]]'s `InjectSocial` door is the same door
the journal writes land through.

## Operational notes

One imposed rule (the 4000-rune budget), enforced at the gate, not by handler
courtesy — everything else about how villagers use the notebook is left emergent
and observed. An agent that never journals is byte-identical to a pre-019
snapshot; a journal that fills up pushes curation pressure onto the agent (an
over-budget write is refused with the exact "N/4000 runes, entry needs M" reason
so it can delete and retry), which is the experiment the feature exists to run.

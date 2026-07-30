---
name: tui-chronicle-feed-guardian-digests
description: Child of [[tui-chronicle-feed]] — the guardian-domain digest grammar entries added since specs 044 (run/morgue), 046 (curriculum), 063 (report card), 076 (world-forked), 084 (plan-layer designations/directives), and the four guardian-miracle entries with their gratis annotation. Load when adding or auditing a guardian-family digest row.
kind: component
sources:
  - internal/tui/grammar.go
  - internal/tui/digest.go
verified_against: 9b4ed5aef5bfea50b67fac10f8e2153f065a814d
---

# TUI chronicle feed — guardian-domain digest entries

Child of [[tui-chronicle-feed]]: the digest grammar entries for the
guardian's own event families — run-outcome/morgue, curriculum, the
report card, world-forking, the plan layer, and miracles. The parent
covers chronicle rendering, the core digest grammar, and the
mental-map/memory-retrieval family entries.

## How it works

Since spec 044, three [[morgue]] types get entries; `familyByNamespace`
(grammar.go) maps their two new namespaces onto existing voices — `run` in
the world-lifecycle voice, `morgue` in the chronicle's narrated-prose
voice: `run.ended` ("the run ended · N dead · final cause <cause>" — the
postmortem reader's feed-line summary; the full ledger stays in the
payload/detail pane), since spec 084 the seven plan-layer types get
entries in the guardian family voice (`familyByNamespace` maps the new
`designation`/`directive` namespaces onto `familyGuardian`): "Guardian
marked a structure_site at (4,5) (shelter) — «label»", "Guardian charged
Ash, Birch: «text»", the id-referencing cancelled/lapsed lines, and the
world-answers-the-plan "the village fulfilled Guardian's mark/charge (id)"
terminals ([[guardian-designations]], `TestCatalogSweep`-covered);
`morgue.epilogue` ("epilogue for <name>: <text>", `chronicle.entry`-style
80-rune truncation; agent −1 renders as "the run" — the run-end epilogue),
and `guardian.charter_observed` ("Guardian ran under charter <fingerprint>
(default|player-authored)" — the charter-revision stamp the morgue aligns
deaths against). Since spec 046, two [[curriculum-ladder]] types get
entries; `familyByNamespace` maps the new `curriculum` namespace onto the
guardian family voice — the ladder is the guardian's domain, not a distinct
visual role: `curriculum.exercise_passed` ("the <exercise> exercise
was passed (<stage>)") and `curriculum.stage_unlocked` ("The guardian's watcher
earned <stage name> (proven by <exercise>)", display name via
`skin.StageName`, like the CLI's stage line). Since spec 063,
`guardian.report_card` ([[grounded-feedback]]) gets its own entry in the
`guardian` namespace (which since spec 094 also carries the 13 renamed
world-action types — the retired `metatron` namespace key left
`familyByNamespace` with the rename; the digest line renders
the skin's report-card label, the charter fingerprint, and the note's text
truncated to 80 runes, `morgue.epilogue`-style). Since spec 076, `world.forked` ([[world-forking]]) gets a
world-lifecycle-voice entry — "forked from `<parent>` at day D, HH:MM", the
fork's provenance in game time (the digest line is the v1 rendering;
chronicle narration of the split is a documented unfunded follow-on). The four
[[guardian-miracles]] types render in the guardian family voice, with a
trailing emphasized `(forced)` annotation (`gratisMark`) when the
payload's gratis flag waived the charge — the feed never conflates an
operator force with a charge-priced miracle. Spec 102
([[guardian-agentization]], [[event-types-guardian-memory]]) adds seven
guardian memory-store rows — `guardian.memory_added/embedded/promoted/
faded`, `guardian.salience_revised`, `guardian.memory_merged`,
`guardian.consolidated` — the `agent.*` consolidation family's wording
re-voiced under the skin's guardian display name (vectors elided, the
spec-042 reasoning).

## Connections

Parent [[tui-chronicle-feed]] covers chronicle rendering, the core digest
grammar, and the mental-map/memory-retrieval family entries. [[morgue]]
owns the run-outcome/epilogue events; [[curriculum-ladder]] owns the
ladder events; [[grounded-feedback]] owns the report card;
[[world-forking]] owns `world.forked`; [[guardian-designations]] owns the
plan-layer event lifecycle; [[guardian-miracles]] owns the four miracle
types and the gratis flag.

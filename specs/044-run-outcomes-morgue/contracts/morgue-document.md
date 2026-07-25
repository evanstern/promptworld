# Contract: the morgue document (`morgue.md`)

A single accumulating legacy document per world, at the save-dir root
(`world.MorguePath()`). **Regenerable view, never a source of truth** (scribe doctrine):
whole-file re-render from the scribe replica + typed event scan on every batch containing
`agent.died`, `run.ended`, or `morgue.epilogue`, and at every daemon boot. Deleting or
hand-editing the file is healed by the next render (FR-011).

## Structure (export-ready, FR-012)

```markdown
# Morgue — <world name>

_One run, one directory. This document is regenerated from the world's history._

## <Villager name> — died day <N> (<cause>)          ← one section per death, event order

- **Days survived**: <N>
- **Cause**: <cause, plain language>
- **What they will be remembered for**: <notable deeds, bulleted, from the curated
  notable-event vocabulary, oldest → newest>
- **What they carried in memory**: <notable memories, salience-ranked>
- **Who mattered to them**: <relationships with trust/affection stated plainly>
- **Debts**: owed <...> · owed to them <...>  (open debts at death — evidence that
  outlives them)
- **The angel's watch at that moment**: charter revision `<fingerprint>` (<default |
  player-authored>), in force since day <N>; standing orders active: <each: condition →
  action, watch subjects>. _Stated as evidence; the reader draws the lesson._

> _Epilogue_ — <narrated prose, blockquote-delimited>          ← ONLY when recorded

## The run — ended day <N>                            ← final section, run end only

- **Run length**: <N> days
- **Population**: 8 → ... → 0 (day-stamped decline)
- **The deaths**: <each: name, day, cause>
- **Notable events of the run**: <curated vocabulary, day-stamped>

> _Epilogue_ — <narrated run epitaph>                 ← ONLY when recorded
```

## Invariants

1. **Facts before prose**: every epilogue is blockquote-delimited and follows its
   section's factual fields; removing all epilogues leaves a complete document.
2. **No scoring language**: the angel-policy section states what was instructed and what
   happened. Contract-banned vocabulary in the factual render: score, grade, blame,
   fault, should have (FR-008).
3. **Byte-identity under replay**: the factual render (all content except epilogue
   blockquotes) is byte-identical across replays of the same history (SC-004). Stable
   ordering everywhere: deaths in event order, memories by (salience desc, tick asc),
   relations/debts by canonical state order.
4. **Complete with zero AI**: all seven factual epitaph fields render on a world that has
   never had a model configured (SC-002).
5. **Sections are append-shaped**: a new death only adds a section; prior sections'
   factual bytes never change (the render is whole-file, but its content is a pure fold
   over a grow-only history) — this is what makes the future Boatmurdered HTML export a
   straight transform.

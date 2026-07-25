# Quickstart: validating the grounded feedback layer (spec 059)

## Automated validation

```bash
go test -race ./...
go test -race ./internal/tool ./internal/guardian -run 'Explain|Tutor|ReportCard|Neutrality' -v
go test -race ./internal/sim -run 'RubricHygiene|ReportCard' -v
go test -race ./internal/tui -run 'Card|HelpGuardian' -v
node scripts/check-tui-design.mjs --changed
```

Covers: explain ground-truth sweep (SC-001), tutor-lane neutrality
(SC-002), non-tutor byte-identity, citation resolution (SC-004),
`?`-section byte-identity (SC-005), rubric-hygiene sweep, spec-052
adversarial battery re-run.

## Manual validation (needs llm.json)

1. Fresh tutor world (`promptworld new <dir>` — stage-1 defaults to the
   tutor preset): ask "how do I play?" — orientation grounded in the guide;
   ask "what does a vision cost?" — the reply cites explain's numbers
   (watch the `» explain` verdict row); charge bank unchanged after both.
2. `promptworld guardian <dir> "what can you do?"` — CLI parity (D1).
3. Pause after some guardian activity — console badge appears; open the
   console (`G`) — the report card composes with an attribution note citing
   real event seqs and the charter fingerprint. Pause again with no new
   activity — no new card.
4. On a scenario world, finish the exercise — the card includes the rubric
   checklist; end a run — the postmortem shows the card content.
5. `?` — the guardian section lists this world's verbs with one example ask
   each; identical every time you open it.
6. No-LLM world: explain still answers (tool-side facts need no model... 
   note: with no LLM there are no guardian turns, so explain surfaces only
   via future non-turn paths — verify instead that the `?` section and
   deterministic card parts render, and nothing errors).

## Re-ground checklist (after merge)

- help.md/guardian-console.md/skin-tokens.md amendments + re-pins (in-PR).
- `/grounding-wiki:wiki-update` — tool-registry.md, the guardian note,
  tui-client.md, llm-orchestrator.md, event-types.md.
- player-docs freshness → regenerate (playing-via-guardian page gains
  "ask your guardian anything"; keys-reference unchanged).

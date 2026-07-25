# Novelty gate — SHIM status and removal condition (T008)

The novelty gate in `internal/mind/convo.go` (all sites marked `SHIM(TASK-109)`,
greppable) is a **removable shim** compensating for weak model-side
conversational variety (operator decision 2026-07-24). It is NOT doctrine.

- **If conversations later feel less dynamic than wanted, look here first**
  (this file and the marked sites), before touching the sim-side pair cooldown
  or the encounter dial.
- **Removal condition**: model tiers whose conversational variety no longer
  needs damping — remove the gate sites wholesale; the sim-side pair cooldown
  (spec 061 US1, doctrine) stays.
- **Lockout safety** (why the floor is 5): the nightly day-gist memory
  (salience 6) refreshes novelty daily, so a pair re-converses at most roughly
  daily absent other salient events; the floor sits above salTalk(3) and
  SalConvoGist(4) so the founding talk's own memory and the prior scene's gist
  can never self-satisfy the gate.

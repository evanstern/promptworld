# Contract: the per-user unlocks record

**File**: `<worlds.Root()>/unlocks.json` (i.e. `~/.promptworld/unlocks.json`, or under
`$PROMPTWORLD_HOME`). Shape in data-model.md.

## Doctrine

1. **Advisory, never authority.** The proving world's event history is the truth; the
   record is a convenience cache of it. Every entry carries evidence pointers (event
   type + seq + tick) into the named world.
2. **Load-tolerant**: missing or corrupt → empty record, never an error. Home-dir
   unresolvable → warn and continue (feature degrades to per-world behavior).
3. **Atomic writes**: `.tmp` + rename; writes heal (drop entries whose world path no
   longer exists? NO — keep them: an archived/moved world is still historical proof;
   only malformed entries are dropped).
4. **Append-shaped**: earning a stage upserts that stage's entry; entries are never
   auto-deleted. Deleting the file loses convenience, not truth (re-earn or re-derive).
5. **Audit**: `promptworld stages` shows earned stages with their proving world; an
   entry whose world is still on disk can be re-verified against that world's history
   (the evidence seqs resolve to real events satisfying the gate conjuncts).
6. **No double-counting via forks** (TASK-67 future): evidence names one world path +
   seqs; a forked directory is a different world and does not inherit or duplicate
   unlock entries — earning in a fork is earning in that world.

## Gate conjuncts (what a valid entry's evidence must satisfy)

- stage-2 entry: a `curriculum.exercise_passed` for a stage-1 exercise.
- stage-3 entry: a pass for a stage-2 exercise **whose evidence includes a
  player-authored `metatron.charter_observed` fingerprint in force at pass time**.
  "Player-authored" is the recorded payload's `default == false` (spec 044 US2,
  specs/044-run-outcomes-morgue/contracts/events.md on main); the evidence entry's `custom` flag is
  derived as its inverse by `sim.CharterObservedEvidence` (T022 reconciliation) —
  a default/preset charter (including the stage-1 tutor preset) never qualifies.
- stage-4 entry: a pass for a stage-3 exercise whose evidence includes a
  player-granted tool's contributing act.

## Consumers

`promptworld stages` (render + earned state), `promptworld new` (offered stages +
informed `--override` path), nothing else — no world behavior ever reads the record.

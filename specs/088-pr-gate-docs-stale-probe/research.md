# Research — spec 088 (TASK-162)

## D1: How the probe's trigger set is derived (FR-001)

- **Decision**: derive the pinned-input set at runtime by scanning `docs/player/*.html`
  at the branch tip for `promptworld-docs:source` tags (the checker's own freshness
  model — each page declares its sources; `check-freshness.mjs` evaluates per-page
  pins over exactly those paths), unioned with the existing `docs/wiki/` prefix rule.
  The extraction helper is the "one named place" FR-001 requires.
- **Rationale**: the card's enumerated list (README.md, docs/llm-providers.md, spec 046
  quickstart sources) is precisely "whatever the pages declare today". Hardcoding it
  would re-create the drift this task fixes the next time a page gains a source.
- **Alternatives considered**: (a) hardcoded const in the gate — rejected: drifts;
  (b) new `--sources` flag on the checker — rejected: changes a second tool's CLI for
  data the pages already carry in tracked HTML.

## D2: History-move predicate (FR-003)

- **Decision**: probe when `git rev-list --merges origin/main..<tip>` is non-empty —
  i.e. the branch contains any merge commit since diverging from origin/main.
  Stateless, computable from the commit graph alone.
- **Rationale**: matches the recorded hazard (merging main into a pin-carrying branch);
  over-triggering is harmless (a fresh probe passes — spec US3 scenario 2), while
  under-triggering is the bug being fixed. No persisted "last probe" state, per the
  gate's no-daemon design (spec 051).
- **Alternatives considered**: (a) persisted probe timestamps — rejected: introduces
  mutable state the gate doctrine forbids; (b) reflog inspection — rejected: reflogs
  are local/ephemeral and absent in fresh clones.

## D3: Design-reference pins become blocking (FR-002)

- **Decision**: when the branch touches `internal/tui/`, `docs/design/tui/`, or any
  source a design page pins, pr mode invokes
  `node scripts/check-tui-design.mjs --changed origin/main...<tip> --json` in the
  gated worktree and maps exits: 1 → new blocking rule `tui-design-stale` (naming the
  failing pages from the JSON report), 2 → blocking `tui-design-env-error`, 0 → no
  finding. The existing warn-level `tui-surface` reminder is kept.
- **Rationale**: mirrors the delegated-checker pattern already proven with the
  player-docs checker (env-overridable path for tests, exit-code contract mapping) and
  reuses the design gate's own pin predicate instead of replicating it.
- **Alternatives considered**: replicating the pin-vs-branch predicate over
  design pages inside the gate — rejected: two implementations of one authority drift.

## D4: Dedup across triggers (FR-004)

- **Decision**: compute all trigger reasons first, invoke each delegated checker at
  most once per run; findings keyed by the existing fingerprint machinery (rule +
  evidence + branch) so combined triggers cannot duplicate.
- **Rationale**: the fingerprint/dedup layer already exists (see
  `computePlayerDocsSurface` tests); reuse it.

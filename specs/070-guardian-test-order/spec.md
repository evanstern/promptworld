# Feature Specification: Guardian worker shutdown joins — de-flake the report-card tests

**Feature Branch**: `070-guardian-test-order` (task branch: `task-144-guardian-close-join`)

**Created**: 2026-07-26

**Status**: Draft

**Input**: TASK-144 (corrected diagnosis 2026-07-26). Original framing
("order-dependent, fails deterministically in isolation") was falsified by
investigation; the corrected diagnosis is pinned on the board card and below.

## Diagnosis (pinned)

`Guardian.Close` (`internal/guardian/guardian.go:274`) is `close(mt.done)` —
a signal with **no join**. `guardian.New` (`guardian.go:259-262`) spawns four
worker goroutines (`reportCardWorker`, `digestWorker`, `triggerWorker`, and
the fourth started in the same block); `newTestGuardian`
(`guardian_test.go:141`) calls `Close()` immediately so queued jobs "stay
inspectable" — but nothing waits for the workers to exit. A worker that has
not yet parked in its two-case select (`reportcard.go:166-175`) when the test
enqueues a job sees BOTH select cases ready on its first pass, and Go chooses
uniformly at random: ~50% of the time it steals the job, and the test's
`drainCard` finds an empty queue → `no card job queued`.

- Probabilistic, not deterministic: isolation `-count=1` passes 10/10;
  `-count=5` fails around iteration 3; full-package `-count=3` fails SIBLING
  tests (`TestReportCardProducerStoresValidatedNote`,
  `TestReportCardRejectsUnrecordedCitation`).
- Smoking gun: `produceCard` log lines emitted after the test's `t.Fatal` —
  the worker is the only other `cardQ` consumer.
- `digestWorker` (`digest.go:186`) and `triggerWorker` (`orders.go:453`) share
  the identical select shape and latent exposure.
- Classification: test-fixture synchronization gap, not a live production
  bug — but the correct fix lands in production shutdown code.

## Decision

**Make `Close` a join.** A `sync.WaitGroup` counts the four goroutines
started in `New`; `Close` becomes `close(mt.done)` then `wg.Wait()`. The
fixture's existing `Close()`-before-enqueue then guarantees every worker has
exited before any test drives the queues — deterministic for all five
`drainCard`-using report-card tests and the digest/trigger fixtures at once,
with zero per-test edits.

Rejected alternatives: test-side queue swaps or poll-with-timeout in
`drainCard` (leaves the worker free to also produce into `inj.batches`,
corrupting other assertions; N fixes instead of one).

**Documented shutdown-behavior change**: a production `Close` during an
in-flight job now waits for that job to finish (previously it returned
immediately). Acceptable — `Close` is shutdown, and a half-processed card was
never a feature. A job enqueued BEFORE `Close` may still be randomly consumed
or left queued (select randomness on a ready queue + closed done) — that
pre-existing nondeterminism is explicitly out of scope.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The report-card tests are deterministic (Priority: P1)

A developer runs any guardian report-card test in isolation or the package in
repetition; it never loses a queued job to a racing worker.

**Why this priority**: the flake burns trust in the suite and CI time; it
already cost one wrong board diagnosis.

**Independent Test**:
`go test ./internal/guardian/ -run 'TestReportCard' -count=50` and full
package `-count=10`, plus `-race`, all clean.

**Acceptance Scenarios**:

1. **Given** the joined `Close`, **When** the five report-card tests run
   `-count=50` in isolation, **Then** zero failures.
2. **Given** the joined `Close`, **When** the full package runs `-count=10
   -race`, **Then** zero failures and zero data races.

---

### User Story 2 - Digest/trigger fixtures get the same guarantee (Priority: P2)

The identical latent race in `digestWorker`/`triggerWorker` fixtures is closed
by the same join — no fixture may drive a queue while its worker might still
be running.

**Independent Test**: package `-count=10` covers them; no per-test changes
expected.

**Acceptance Scenarios**:

1. **Given** the joined `Close`, **When** digest/trigger tests run repeated,
   **Then** no queue-steal failures.

---

### Edge Cases

- **Close during an in-flight job**: `wg.Wait()` blocks until the worker's
  current iteration completes — documented behavior change above; no timeout
  added (shutdown correctness over speed).
- **Double Close**: today double `close(done)` would panic — preserve whatever
  guard exists (verify; if none exists and no caller double-closes, do not add
  one — scope discipline).
- **Worker goroutines added later**: the WaitGroup pattern must make the
  add-site obvious (wg.Add adjacent to each `go` in `New`) so a future fifth
  worker can't silently skip the join.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `Guardian.Close` MUST NOT return until every goroutine `New`
  started has exited (`sync.WaitGroup`; `wg.Add` at each spawn site,
  `defer wg.Done()` first line of each worker).
- **FR-002**: No exported-API shape change; no behavior change other than the
  documented Close-joins semantics.
- **FR-003**: No test-file edits are required for the fix; test edits are
  permitted ONLY to strengthen assertions/counts, never to mask timing.
- **FR-004**: Proof: report-card tests `-count=50` isolation clean; package
  `-count=10 -race` clean; full `go test ./...` green.

### Key Entities

- **Guardian worker lifecycle** — the four `go` statements in `New`
  (`guardian.go:259-262`), `done` channel, new `wg sync.WaitGroup` field.

## Success Criteria *(mandatory)*

- **SC-001**: The TASK-144 board AC passes: affected tests `-count=50` clean
  in isolation and full-package.
- **SC-002**: `go test ./internal/guardian/ -count=10 -race` clean.
- **SC-003**: Full suite green; no other package's behavior changes.
- **SC-004**: The wiki note pinning `internal/guardian/guardian.go`
  (`docs/wiki/guardian.md`) is re-verified IN THIS BRANCH if the pr gate (post
  spec 069) demands it — the Close-semantics line is a real behavior fact the
  note may state.

## Assumptions

- The fourth goroutine in `New`'s spawn block joins the same WaitGroup even if
  its select shape differs — the join is about lifecycle, not queue shape.
- Spec 069's gate may or may not be merged before this PR opens; this branch
  complies either way (re-pin guardian.md in-branch if its sources are deemed
  touched — `guardian.go` IS a pinned source of `docs/wiki/guardian.md`).

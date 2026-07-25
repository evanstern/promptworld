# Tasks: Metatron Survival Autonomy

**Input**: spec.md, plan.md (specs/059-metatron-survival-autonomy/)

**Tests**: included — SC-001..004 demand test evidence; orders are
event-sourced so replay determinism applies.

## Format: `[ID] [P?] [Story] Description`

## Phase 1: Setup

*None.*

---

## Phase 2: Foundational

- [ ] T001 Verify MetatronOrder origin persistence end-to-end (place → event →
      state → replay); extend the payload compatibly (omitempty) ONLY if origin
      is not already durable. Define the system origin value + the three
      canonical watch definitions (near-death, starvation, exposure) reusing
      existing danger-band constants (name them in the PR description).

---

## Phase 3: User Story 1 — Survival watches from birth (P1)

- [ ] T002 [US1] Origin-keyed exemptions in `internal/metatron/orders.go`:
      system watches bypass the player-order cap and TTL validation; player
      cancel of a system watch refuses in-fiction.
- [ ] T003 [US1] Seeding: genesis (new worlds) + boot seed-if-absent (existing
      worlds) at the established seams (spec-057 genesis pin /
      seedMeetingConvention pattern); no duplicates across boots (replay-safe).
- [ ] T004 [P] [US1] Tests: new-world log carries the three watches; pre-059
      fixture boot-seeds once, second boot seeds nothing; cap counts player
      orders only; cancel refusal (SC-001).

---

## Phase 4: User Story 2 — Survival authority carve-out (P1)

- [ ] T005 [US2] Turn-frame carve-out in `internal/metatron/turn.go`: the
      initiative block becomes turn-origin-conditional — survival-watch turns
      permit visions/miracles on own authority (charges unchanged);
      non-survival frames keep today's text (pin both with tests).
- [ ] T006 [US2] Guard rails: clock control and self-placed non-survival orders
      stay outside initiative in BOTH frames; survival-turn actions attributed
      to the survival duty in the soul/chronicle record.
- [ ] T007 [US2] Charter wording via the charter-observed mechanism
      (`internal/metatron/charter.go`) reflecting the survival duty.
- [ ] T008 [P] [US2] Tests: watch-match turn executes a miracle with zero
      player input; player-chat turn keeps refusal; clock refusal in both
      frames; helpless turn (zero charges) recorded not silent (SC-002/003).

---

## Phase 5: User Story 3 — Targeting digest (P2)

- [ ] T009 [US3] Token-bounded villager positions/conditions + passability
      digest in miracle-capable prompts (`turn.go` prompt assembly +
      `internal/tool/derive.go` guidance hook).
- [ ] T010 [P] [US3] Tests: digest present and within budget; door round-trip —
      a miracle targeted at digest-listed coordinates passes the landing door
      (regression for world-01's 3-of-4 rejects) (SC-004).

---

## Phase 6: Polish & Cross-Cutting

- [ ] T011 Full gates in worktree: `go test ./...`, `go vet ./...`,
      `node scripts/check-tui-design.mjs --changed`; re-run after rebase.
- [ ] T012 Post-merge (root): wiki re-pin (metatron notes), player-docs
      freshness, spec-bridge sync.

# Implementation Plan: Event-log format_version + translating migration + guardian rename (TASK-134)

**Branch**: `task-134-event-log-format-version` | **Date**: 2026-07-29 | **Spec**: [spec.md](spec.md)

## Summary
Three moves in one PR: (1) a log-level format-version stamp written at genesis and
enforced at load (older ⇒ migrate hint; newer ⇒ upgrade refusal); (2) a TRANSLATING
migration mode in the existing promptworld-migrate driver (type-rename maps, every
event/tick/payload preserved, archive + guards) alongside the existing snapshot-cut;
(3) the real metatron.*→guardian.* rename of all persisted guardian event types with
TASK-121's display-alias shim removed. Byte-identity harness proves
replay(source, old) == replay(translated, new) on a seeded fixture world.

## Technical Context
**Language**: Go. **Surfaces**: internal/store (log stamp), internal/world
(open gating, migrate driver), internal/sim (reducer arms/whitelists/types),
internal/tui + chronicle digest (type rendering), cmd/promptworld (migrate CLI),
tests incl. TestCatalogSweep and replay/determinism harness.
**Constraints**: stateless stamp readable without replay; never-overwrite archive
guards; live-daemon refusal; playtest world untouched by tests; recipes_test value
pin superseded by real versioning.

## Constitution Check
I–IV: PASS (spec 094 encodes both operator rulings; one branch/PR; byte-identity
harness + refusal tests as evidence; wiki re-pins in-branch — event-log /
sim-state-reducer / chronicle notes expected NEEDS-REVIEW).
V: PASS — **Opus** (card-stated: replay/reducer doctrine, cross-package, migration
machinery). Recorded on the board task.

## Project Structure
See spec.md; research note (T001) records stamp shape + manifest-bump decision +
the exact frozen-type enumeration before any code.

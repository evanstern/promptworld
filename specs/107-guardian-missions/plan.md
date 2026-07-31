# Implementation Plan: Guardian missions (TASK-158)

**Branch**: `task-158-guardian-missions` | **Date**: 2026-07-30 | **Spec**: [spec.md](spec.md)

## Summary
Mission artifacts (084 discipline) accepted through guardian chat, pursued via
existing verbs on spec 102's scheduled lane at full competence (missions are
pre-authorization, not initiative), completion derived from spec 084 predicates
+ recorded events into the report card; EASY-mode default-charter obedience
clause gated by an in-branch old-vs-new obedience eval.

## Technical Context
**Language**: Go. **Surfaces**: internal/sim (mission entity + arms),
internal/guardian (acceptance door, pursuit context, ceiling composition,
charter default), internal/tool (roster/gloss if a mission verb surfaces),
report-card producer, TUI digest/decision trail, docs. **Testing**: FR-006
suite + live demo + obedience eval. **Constraints**: existing doors only;
additive events; order door single-arbiter; playtest untouched.

## Constitution Check
I–IV: PASS (spec 107 encodes the firm 2026-07-26 decision + the 2026-07-30
eval-gate ruling; one branch/PR; demo + eval + tests as evidence; re-pins
in-branch). V: PASS — **Opus** (doctrine-adjacent initiative frame,
cross-package). Recorded on the board task.

## Project Structure
Evidence under docs/design/evidence/task-158/ (demo + obedience eval); new
wiki note guardian-missions.md.

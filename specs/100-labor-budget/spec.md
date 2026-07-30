# Feature Specification: The survival labor budget — effort, healing, and health retune

**Feature Branch**: `task-30-labor-budget`

**Created**: 2026-07-29

**Status**: Design for ratification (TASK-30 design session; implementation is
future work scoped from this spec)

**Input**: TASK-30 + the card's pre-session decisions (operator, 2026-07-20:
(1) feeding self + one other ≈ 8h of hunting/foraging; (2) chop ≈ 30 min = 1
wood; (3) hut ≈ 2 person-days, fire ≈ 1h; (4) healing simple, sleep preferred;
(5) permadeath makes depletion final) + spec 012's re-denominated food pins
(raw 40 / cooked 80 / meal 100; drift audit 2026-07-23) + absorbed TASK-29 fire
economics (spec 012: 2 wood → 8 game-hours, +4h/refuel, cap 12h) + spec 090's
season dials (this spec supplies the demand-side arithmetic 090 defers here) +
the learning-game constraint (2026-07-25 comment: stakes make the grade real;
reconcile against classroom session lengths).

> **Design-session artifact.** Ratifies the retune's invariants, decisions, and
> a solved reference dial set. Constants are DESIGN TARGETS for the implementing
> task; nothing ships here.

## The invariants (decision-3 made arithmetic)

- **I1 — Solo break-even**: one villager's ~8h workday through the best
  reliable pipeline (forage + cook) feeds exactly ~2 people (2880 need-points:
  need drains 1440/day).
- **I2 — Cooperative surplus**: a village of 8 with a shared hearth and pooled
  hunts sustains itself at ~4–5 labor-hours per agent-day.
- **I3 — Winter deficit**: cold-season scarcity (spec 090: forage ×0–0.25, den
  yields thinned) makes solo self-feeding arithmetically impossible —
  stockpiles or death.
- **I4 — Building is expensive**: hut ≈ 2 person-days total effort; fire ≈ 1h
  to raise; the nightly fuel bill (~8 wood) is brutal solo, cheap shared.
- **I5 — Healing is sleep-only**: regen only while asleep AND fed AND warm;
  awake regen removed. Sleep becomes load-bearing; wounds cost labor-hours.

## Decisions

- **D1 — Units stay spec-012-denominated; the card's "1 unit/hour" re-bases to
  need-points/hour.** raw 40 / cooked 80 / meal 100 keep their pins; the 2026-07-20
  "unit" (~350 pts, pre-012) becomes "≈360 pts of labor value per hour" as the
  forage+cook target. Rationale: preserves every spec-012 contract; the
  worksheet below shows the invariants solve cleanly.
- **D2 — Hunting: chunky EV slightly BELOW forage, high variance, den
  cooldowns.** A 6h expedition at ~55% success with a ~40-raw-meat carcass
  (3200 cooked pts) gives ~293 pts/h EV vs forage's 360 — the premium for
  reliability stays with foraging, while a successful hunt overfeeds the hunter
  by ~2 person-days: structural sharing pressure and debt fodder (the social
  fabric's food).
- **D3 — Cooperative construction ADOPTED.** Multiple agents advance one build
  site; a hut is a village project. (Interaction shape rides existing goal
  machinery; social scenes are spec 093's business.)
- **D4 — Healing sleep-only** per I5; health depletion sources unchanged
  (starvation/exposure/wounds); permadeath unchanged.
- **D5 — Classroom reconciliation: the budget bends time, never effort.**
  Lesson worlds reach interesting states via scenario time-snaps, seeded
  incidents, and (per spec 090) shorter season dials — never by discounting
  labor costs. Rationale: rubrics mean nothing if the lesson world's economy is
  fake; session-length knobs are time-shape knobs (tuning), not price knobs.

## The calibration worksheet (card AC #2)

Reference dial set (targets; implementing task pins exact constants):
forage ≈ 4.5 raw/h (~13 min/unit); cook doubles value (raw 40 → cooked 80);
hunt 6h, p≈0.55, carcass ≈ 40 raw meat; chop 30 min → 1 wood; fire raise ≈ 1h
+ spec-012 fuel (2 wood → 8h, cap 12h, ~8 wood/night uncapped equivalent); hut
≈ 12 wood + 10h construction; sleep-only regen +1/min (asleep ∧ fed ∧ warm).

| Regime | Arithmetic | Verdict |
|---|---|---|
| Solo, warm season | 8h forage → 36 raw → cook → 2880 pts = need(self)+need(one) | **Break-even at 8h/day** (I1 ✓); fuel bill (4h chop/night solo) forces fire-sharing or cold nights — solo life is possible, thin, and joyless (decision-3) |
| Village of 8, warm | need 11,520 pts/day; capacity at 4.5h/agent-day = 36 labor-h ≈ 12,960 pts via forage-equivalent; hearth amortizes to ~0.5h/agent; hunt variance adds feast-surplus | **~4.5h/agent-day sustains with margin** (I2 ✓); surplus funds building (hut = 16h ≈ one villager's 3.5 days of slack or a village afternoon, I4 ✓) |
| Solo, cold season | forage ×0.25 ⇒ 90 pts/h ⇒ self-need alone = 16h/day > waking day | **Structural deficit** (I3 ✓): survival = warm-season stockpile ≈ 0.25 × 1440 × season-days extra pts banked |
| Healing | wound −200 HP ⇒ 200 min asleep-fed-warm ≈ 3.3 game-hours of protected sleep + the food/warmth to fund it | **Injury costs the village real labor** (I5 ✓); nursing matters |

Cross-checks: spec 090's hot-season compat anchor holds (these dials replace
the pre-seasons binary world as the new baseline the 090 implementation
calibrates against); spec 013 carry caps make the I3 stockpile physical and
stealable; TASK-95's intent_failed keeps failed hunts loud.

## User Scenarios & Testing *(mandatory)*

### US1 - Surviving a day costs most of a day (Priority: P1)

As a player, I want a villager's day to be spent mostly on staying alive — so
that every hut raised, feast shared, or myth invented is visibly purchased out
of scarce hours, and losing feels like arithmetic, not dice.

**Acceptance Scenarios**:

1. **Given** the reference dials on a warm-season world, **When** a solo agent
   forages+cooks ~8h/day, **Then** it sustains itself + one dependent and no
   more (worksheet row 1, verified live by the implementing task).
2. **Given** a village of 8 cooperating, **Then** ~4–5h/agent-day sustains with
   margin (row 2).
3. **Given** a cold season and no stockpile, **Then** a solo agent starves on
   any schedule (row 3).

### US2 - Sleep heals, so sleep matters (Priority: P2)

1. **Given** a wounded agent, **When** asleep AND fed AND warm, **Then** it
   regens; awake it does not — verified against the removed awake-regen path.

## Requirements (for the future implementing task)

- **FR-001**: Pin the reference dials (or session-note-justified adjustments
  that preserve I1–I5) in the tuning manifest with genesis-pin compat
  (pre-retune worlds keep old dials).
- **FR-002**: Cooperative build sites (D3) through existing goal machinery.
- **FR-003**: Sleep-only healing (D4); remove awake regen.
- **FR-004**: Live calibration runs proving worksheet rows 1–3 on seeded
  worlds; evidence doc.
- **FR-005**: Coordinate with spec 090's implementation (shared dial family) —
  one implementing task MAY deliver both.

## Success Criteria

- **SC-001**: This spec ratified via PR review; the implementing task can pin
  dials without a second design session.
- **SC-002**: Worksheet arithmetic internally consistent (rows recompute from
  the dial set alone) — checkable by a reviewer with a calculator.
- **SC-003**: Every pre-session operator decision (1)–(5) is satisfied or its
  deviation named: (1) ✓ via D1 re-basing; (2) ✓; (3) ✓; (4) ✓ sleep-only;
  (5) ✓ unchanged.

## Assumptions

- Session-length classroom knobs (D5) live in spec 090's season dials + spec
  054 incident machinery; nothing here blocks classroom mode.
- Tier for the future implementation: expect Sonnet with Opus escalation if the
  retune destabilizes reflex arbitration (TASK-101/108 territory).

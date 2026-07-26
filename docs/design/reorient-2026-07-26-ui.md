# promptworld — reorientation synthesis: post-cockpit delta (2026-07-26)

**Status:** synthesis complete; board moves pending operator sign-off (execution gated on
"the board allows a clean sweep run" — the PDLC-hardening lanes TASK-144/145/147 and other
in-flight work clear first).
**Run:** `promptworld-2026-07-26-16-25-19` (reorient re-run of `promptworld-2026-07-25-13-58-18`).
**Worktree:** authored on branch `reorient-2026-07-26-ui` (root stays on main; a concurrent
reorient session is active on this repo — isolation is deliberate).
**Lens (verbatim from the 2026-07-25 run):** promptworld is pivoting into a staged
prompting-skills teaching game (curriculum ladder TASK-68, skinnable guardian TASK-121, the
eight ratified learning-game decisions of 2026-07-25) while remaining an ambient,
terminal-first LLM-agent world sim. The UI/UX must: (a) make a living LLM-agent village
legible at a glance; (b) make prompting the guardian the central, rewarding player verb;
(c) teach itself in play; and (d) be fully specifiable — page by page, feature by feature,
control by control — so both humans and AI implementers can build new features and update
existing UI directly from the design doc.
**Corpus:** `research/Game-UI-UX` (Analysis-Post-Cockpit-UI-Delta),
`research/Game-Gameplay-Patterns` (Analysis-Iteration-Rung-Delta),
`research/Learning-Game-Design` (Analysis-Feedback-Validity-Delta),
`research/Game-Player-Docs` (Analysis-Semantic-Honesty-Delta).

## TL;DR

The 2026-07-25 Staged Cockpit shipped **in full and faithfully to the corpus** — all four
evaluators verified their branch's recommendations in shipped code and the pinned design
reference, not just in claims. Lens clauses (a)–(c) are served by shipped surfaces; clause
(d) — the one that failed outright last run — now **passes structurally but not
semantically**: the freshness gate validates pins and table headers while a shipped page
carried seven stale `unbuilt` cells, and the report card renders **✓ on a failing outcome**
(`agent.died: 2`) at the game's most salient teaching moment. The re-run's converged
diagnosis is four gaps, one per layer of the same loop: failures are mis-graded at the
teaching moment (report-card truth), the reference can lie semantically (gate lint), the
loop cannot be iterated (fork duel), and the map cannot be interrogated (look-cursor) —
plus a content problem everyone circles: one real exercise, one incident kind, no in-TUI
forward ladder. **No refactor is necessary or advisable** — every fix is additive on the
shipped substrate, and two of the four gaps are single-afternoon items.

## Decisions

Operator decisions, steering round of 2026-07-26 ("accept those recs"; verbatim intent;
FIXED constraints for cross-grounding). Execution note: the implementation sweep starts
only **when the board allows a clean sweep run** (the PDLC-hardening sweep TASK-144/145/147
and other in-flight lanes finish first).

1. **Report-card truth (NEW card, HIGH):** unify every report-card surface (postmortem,
   ceremony, console card) on `sim.EvaluateRubric` — a failed term renders ✗ at the
   postmortem — and author `the-law`'s real evaluator (persist the charter `Default` flag
   into state; the documented blocker at `internal/sim/scenario.go:277`). Sequenced after
   TASK-144 (flaky report-card test touches the same code).
2. **Design-gate semantic lint (NEW card, small):** extend `scripts/check-tui-design.mjs`
   to warn/fail when a `status: shipped` page contains `unbuilt (wave` renderer cells
   (optional: grep-level check that named renderer symbols exist in `internal/tui`); amend
   `overlays/postmortem.md`'s seven stale cells and `panels/exercise.md:110` in the same PR.
3. **TASK-67 fork duel:** promote Medium → HIGH; reframe as the loop's iteration rung —
   all D7 prerequisites shipped; v1 = rubric-first scoreboard sharing `reportCardView` +
   `sim.EvaluateRubric`, then the HTML retelling; dual-TUI stays deferred.
4. **TASK-142 look-cursor:** runs in the next UI lane. Amend the design with the DF fixed
   tile-content hierarchy (agents → piles/chests → structures → terrain) and use the
   spec-068 tile registry's `meaning` rows as the TILE pane's whatis content (plain
   language per FR-020).
5. **Exercise catalog wave:** ladder v1 = 2–3 hand-authored exercises per stage; grow the
   incident vocabulary to ~3 kinds (cold snap, forage blight, stranger/trickster arrival),
   each a reducer-valid event shape indistinguishable from an ambient cause.
6. **In-TUI ladder view:** the forward ladder (identity · concept · earned/next · unlock
   evidence, matching `stages --json`) renders in the `?` guardian section — deterministic
   floor, model-free.
7. **Quickstart first-prompt step:** `getting-started.html` gains an "ask your guardian one
   thing" step (sample ask from the `skin.guardian.example_ask.*` token family); each stage
   page gains a short first-session do-this-then-this block. Content-only, player-docs skill.
8. **Mouse-parity sweep test:** commissioned as a small task — parse the control tables'
   `keys+mouse` column, assert every non-`—` mouse cell has a handler, burn down
   `patterns/keymap.md`'s rollout note.
9. **Lane order: duel before faith.** TASK-67 lands before TASK-118 (which itself lands
   before TASK-112, per TASK-112's AC5/AC6 dependency).
10. **Parked as recorded watch items (not scheduled):** proving-run timing vs the ambient
    drama floor (TASK-14/28/133 posture); rubric-gauge exposure doctrine (all-terms-live
    stands until multi-exercise content exists); teaching-signal instrumentation consent;
    fire-once lesson doctrine (revisit at >15 lessons); retry accommodation; chronicle/world
    search `/` (revisit past ~20 villagers); colorblind/contrast palette research; player-docs
    pin de-churn; TASK-146 → Medium reassessment after TASK-145 merges.

## Merged positions

**1. The shipped cockpit is corpus-faithful (all four branches).** Every 2026-07-25 wave is
verifiable in `docs/design/tui/` (25 files, `status: shipped`, pinned) and `internal/tui/`:
guardian console with document-style turns and `$EDITOR` charter handoff (Game-UI-UX's
chat-console conventions), lesson row as the RimWorld ConceptDef transplant with per-user
seen state (Learning-Game-Design), director-lite as a pure replay-safe function with a real
`incidentSource` seam (Game-Gameplay-Patterns' honesty note followed literally), and
Cogmind's four help layers all in code (Game-Player-Docs). Spec 068 additionally executed
Game-UI-UX's own tile-vocabulary analysis with high fidelity (`semantic16`/`material256`
classing verbatim in `internal/tui/tiles.go`).

**2. One honesty doctrine, two enforcement faces (Learning-Game-Design + Game-Player-Docs,
reinforced by the other two).** The run's sharpest converged finding: *gates verify meaning,
not just structure.* Face one, screen→world: the report card grades by generic event
presence (`reportCardFactsFromEvents`, `internal/tui/views.go:863-880`) and renders ✓ on a
failing outcome, while `sim.EvaluateRubric` sits unused one package away; the pedagogy
corpus is unambiguous that a false ✓ at the failure moment mis-teaches the rubric vocabulary
the ladder rests on, and the gameplay branch adds that it violates the morgue's no-blame
register, which only teaches if the evidence is true (decision 1). Face two, doc→code: the
design gate passed while `overlays/postmortem.md` (shipped, freshly pinned) carried seven
`unbuilt (wave 4)` renderer cells for renderers that exist and are tested — the docs
corpus's "convention-dependent reference rots" prediction, caught in the wild (decision 2).
The docs analysis names a third residue class neither fix catches mechanically: stale
ownership pointers in design prose ("owned by TASK-119" — a Done task); the lint's scope
note covers it as a review responsibility.

**3. The iteration rung (Game-Gameplay-Patterns, adopted by all).** The loop is built but
the game is one lesson long: two exercises, one production rubric, one incident kind. The
fork duel (TASK-67) is the corpus's tightest answer to "did my prompt edit work?" — the
Opus-Magnum histogram-without-rewards engine (Learning-Game-Design) and the losing-is-fun
postmortem artifact (Game-Player-Docs) merged into one framing, now dramatically cheaper
because its presupposed renderers shipped (decision 3). Order constraint recorded: decision
1's truth unification must land before the duel's scoreboard renders comparisons, or the
duel compares false checkmarks — nothing enforces this yet; the board move below encodes it
as a dependency.

**4. Map interrogation completes the inspector chain (Game-UI-UX + Game-Player-Docs).**
The last unshipped corpus pattern: Smallville's inspector chain and DF's `k` cursor both
run through the map; promptworld's runs only through lists. TASK-142 (already designed and
mocked) is the payment, amended by cross-grounding: DF's fixed tile-content hierarchy for a
stable scan order, and the tile registry's `meaning` rows as the TILE pane's whatis content
— making the look-cursor the third in-place lookup (after `?` and explain) and shrinking
what external docs must carry (decision 4). The badge deep-link (the docs branch's one
unshipped item) folds into this lane. Net-new unscheduled recommendation from the delta:
reverse jump (strip glyph / roster row → camera center) — home to be chosen at spec time.

**5. Content is now the constraint, not surfaces (Game-Gameplay-Patterns +
Learning-Game-Design).** The catalog wave (decision 5) is the delivery vehicle: 2–3
exercises per stage, ~3 incident kinds entering through the shipped severity grammar
(Game-UI-UX's rider — small stable vocabulary extends to event vocabulary), the pedagogy
branch's lesson tranche 2 and first wrong-thing detector riding as content, and the in-TUI
forward ladder (decision 6) closing the WorldBox discoverability critique — with the docs
branch's invariant rider: the ladder view is status-derived, so `overlays/help.md`'s
byte-identity table gains a row in the same PR. The quickstart finally has the player
prompt (decision 7): a first session that never prompts teaches watching.

**6. Verification culture, extended (all four).** The parity sweep test (decision 8)
converts input-parity doctrine into a gate, the project's house style. All four evaluators
independently chose INDEX-just-in-time wiki grounding and verified against design corpus +
code instead — evidence for the parked TASK-146 (CAPSULES) reassessment: capsules matter
for the doc-generation chain's routing budget, not for evaluator honesty.

## Course of action

**No big refactor — again.** Every item below is additive on the shipped substrate (the
replica/reducer machinery, the five-region anatomy, the tile registry, `EvaluateRubric`,
`reportCardView`). The one structural risk named this run is a *tension*, not a refactor:
the pull-surface budget (decisions 4 + 6 grow tab-cycled overlays; see Open questions).

Gate for all waves: **the board allows a clean sweep run** (TASK-144/145/147 and other
in-flight lanes clear; concurrent-session traffic quiesces).

- **Wave A — honesty (smallest, highest leverage per line):** TASK-144 (flaky test, already
  in flight elsewhere) → **report-card truth** card (decision 1, builds on `EvaluateRubric` +
  `reportCardView`) → **semantic lint** card (decision 2, builds on `check-tui-design.mjs`;
  fixes postmortem.md/exercise.md in the same PR).
- **Wave B — UI lane:** **TASK-142** (decision 4 amendments; builds on the dock borrow seam,
  villager-detail renderer family, tile registry) + **parity sweep test** card (decision 8;
  builds on the control tables). Reverse-jump home decided at spec time.
- **Wave C — iteration:** **TASK-67 duel v1** (decision 3; depends on Wave A's truth card —
  encoded as a task dependency), then the Boatmurdered-style HTML retelling as phase 2.
- **Wave D — content:** **exercise catalog wave** card (decision 5) + **in-TUI ladder view**
  card (decision 6, with the byte-identity rider) + **quickstart first-prompt** content pass
  (decision 7).
- **Then:** TASK-118 faith → TASK-112 agentization (decision 9's order; TASK-112 additionally
  held behind TASK-137's charter-delta evidence per its own AC5).

## Board moves

Executed via the `backlog` CLI from the repo root only after operator sign-off on this
table, and only when the board allows a clean sweep run. Each move cites this synthesis.

| # | Task | Move |
|---|------|------|
| 1 | NEW | **Report-card truth: unify all card surfaces on `sim.EvaluateRubric`** (HIGH) — postmortem/ceremony/console cards render real rubric verdicts (✗ on failure); author `the-law`'s evaluator (persist charter `Default` into state, `scenario.go:277`); after TASK-144. Decision 1. |
| 2 | NEW | **Design-gate semantic lint** (small) — `check-tui-design.mjs` flags `shipped` pages with `unbuilt (wave` renderer cells (+ optional symbol-exists check); fix `overlays/postmortem.md` ×7 and `panels/exercise.md:110` same-PR. Decision 2. |
| 3 | EDIT TASK-67 | Priority → HIGH; reframe: the loop's iteration rung, all D7 prerequisites shipped; v1 shares `reportCardView` + `EvaluateRubric`; **depends on move #1**; HTML retelling = phase 2; dual-TUI deferred. Decision 3. |
| 4 | EDIT TASK-142 | Add ACs: DF fixed tile-content hierarchy (agents → piles/chests → structures → terrain); TILE pane whatis from tile-registry `meaning` rows (plain language, FR-020); badge deep-link folded into this lane; reverse-jump home decided at spec time. Decision 4. |
| 5 | NEW | **Exercise catalog wave** — 2–3 hand-authored exercises per stage; ~3 incident kinds (cold snap, forage blight, stranger) as reducer-valid ambient-indistinguishable events entering via shipped severity channels; lesson tranche 2 + first wrong-thing-detector lesson ride as content. Decision 5. |
| 6 | NEW | **In-TUI forward-ladder view** — `?` guardian section block (identity · concept · earned/next · unlock evidence, `stages --json` parity); same-PR row in `overlays/help.md`'s byte-identity table. Decision 6. |
| 7 | NEW | **Quickstart first-prompt pass** (content) — `getting-started.html` "ask your guardian one thing" step from `skin.guardian.example_ask.*`; per-stage first-session blocks; player-docs skill is the home. Decision 7. |
| 8 | NEW | **Mouse-parity sweep test** (small) — parse control tables' `keys+mouse`, assert handlers exist, burn down keymap.md's rollout note. Decision 8. |
| 9 | EDIT TASK-118 | Note lane order: after TASK-67, before TASK-112; strip integration pre-specified (`panels/guardian-strip.md` §4); corpus riders: failure-spiral AC grounded in the Hades reasoning; faith stays in-fiction, never a badge/streak surface. Decision 9. |
| 10 | EDIT TASK-17 | Reframe upward: raises jump-to-source's locatable-event hit rate (`resolveSubject`) — village-lens completion, not just chronicle hygiene. |
| 11 | EDIT TASK-28 | Reframe dual-duty: ambient drama supply AND authorable scenario incident vocabulary (a seeded cold snap is both). |
| 12 | EDIT TASK-23 | Reframe: the DF-pole drama generator; chronicle requirements include rubric-legibility for future social exercises. |
| 13 | EDIT TASK-133 | Relabel: learning-game prerequisite (postmortem attribution can't teach unless the sim names neglect); alert enters via shipped severity grammar, no new channel. |
| 14 | EDIT TASK-146 | Add: doc-generation chain is a named capsule consumer (`guardian.md` capsule fails the routing budget); reassess Low→Medium after TASK-145 merges; note the four-evaluator INDEX-JIT convergence. |
| 15 | FLAG only | TASK-143 shows In Progress though PR #105 merged — owned by the concurrent sweep session's lane; surface, don't touch. |

## Open questions

Only decisions genuinely still the operator's, each parked with its resurfacing trigger:

1. **Retry accommodation** (the run's one contested item: Hades' retention evidence vs
   Cogmind's destabilization warning) — resurfaces on playtest evidence of repeated
   same-exercise failure churn; any adoption carries the honesty-marker condition
   (`stage_overridden` precedent).
2. **Ambient drama floor** (Game-Gameplay-Patterns' unfunded concern) — run TASK-14's
   proving run against today's calm world, or fund TASK-28/TASK-133 first? Resurfaces when
   TASK-14 is scheduled.
3. **Pull-surface budget** (new, from cross-grounding) — decisions 4 and 6 plus D9 all grow
   tab-cycled overlay/pane surfaces; at what point does the help/inspect stack need its own
   navigation ruling (the DF "three ways of scrolling" caution)? Resurfaces at TASK-142 spec
   time.
4. **Reverse-jump home** — TASK-142 rider, parity-sweep rider, or own card; decided at spec
   time.
5. **Decision-10 stack** (verbatim from Decisions): proving-run timing, rubric-gauge
   exposure, instrumentation consent, fire-once doctrine, search `/` grammar (one grammar
   across chronicle/help/world when it comes), colorblind palette research, player-docs pin
   de-churn, TASK-146 timing.

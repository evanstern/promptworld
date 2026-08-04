# Session card template

Scaffold for a polish session's single long-running card. Copy the description block into
`backlog task create`, keep the acceptance criteria as written, and append to the decision log
as the session runs.

Card format follows spec 087 (see CLAUDE.md): the opening gist first, then "As a …" use cases.
Replace every `<…>`; delete nothing else.

---

## Description

A single long-running card that holds one rapid polish session. Instead of a spec per fix, we
watch a live world, discuss small bugs / UI tweaks / gameplay nits as they come up, record each
decision on this card before code is written, and land the whole session on one branch in one PR.

## Use cases

- As a player, I want the small rough edges I keep bumping into — a misaligned pane, a confusing
  label, a villager doing something obviously silly — fixed quickly, without each one waiting on
  its own full spec cycle.
- As the operator, I want to sit in front of a real running world, point at what looks wrong, and
  have the decision written down before anyone writes code, so a fast session still leaves a
  paper trail.
- As a reviewer, I want one PR whose card lists every item worked, the diagnosis behind it, and
  whether it was decided ad-hoc or escalated to a spec.

## Session contract

1. Worktree + branch off origin/main; this card is claimed for the session's duration.
2. Live world: `<world-id>` at `<path>`, restarted against the branch build as we iterate.
   Nothing else runs against it.
3. This card is the only task; there are no per-item task cards and no per-item PRs.
4. Development sub-loop, repeated:
   a. Discuss a feature, bug, or tweak.
   b. Optionally ground the topic with research / wiki reads / frame dumps.
   c. Record the decision and its complete file:line diagnosis in the decision log below,
      BEFORE implementing.
   d. Implement, then prove it on the live world and record the proof.
5. After each loop, apply the trivial-exemption bar: surgical fix + complete file:line diagnosis
   + ACs on this card. An item that fails it escalates to a spec, linked via `spec-bridge:link`.
6. Execute any spec created. Then decide whether to go again.
7. Test: operator visual QA on the live world, against a binary rebuilt after the last change.
8. ONLY THEN: wiki re-pin, player docs, design references, PDLC gates. Nothing is re-pinned per
   item — grounding happens once, at the end, in-branch.
9. Open the PR; merge with `--merge`, never squash.

**Spec number claimed up front:** `specs/<NNN>-<slug>` — stub landed on main before the first
content commit, per CLAUDE.md's stub-first rule. `<Or: no escalation is plausible for this
session's scope; state that instead and accept that escalating later costs a stop.>`

Scope guard: this flow is for polish — FE/TUI changes and small gameplay tweaks against decisions
already made. Anything needing real design thought leaves this session and goes through the full
PDLC cycle on its own card.

## Decision log

(entries appended as the session runs — one per item: what, diagnosis with file:line, decision,
ad-hoc vs spec, then a SHIPPED entry with the commit and the live proof)

---

## Acceptance criteria

Create these with `--ac`, verbatim:

1. Every item worked in this session has a decision-log entry on this card — item, file:line
   diagnosis, and the decision — recorded before its implementation
2. Any item exceeding the trivial-exemption bar (surgical fix + complete file:line diagnosis +
   ACs on this card) is escalated to a Spec Kit spec and linked via spec-bridge before
   implementation
3. All session work lands on a single branch and a single PR; no per-item task cards or PRs are
   created
4. Operator visual QA passes on the live world for every shipped item before the PR is opened,
   against a binary rebuilt after the last change
5. Grounding is done once at the end, in-branch: wiki re-pinned, player docs regenerated,
   tui-design amended where internal/tui changed, and the pr merge-drift gate is green

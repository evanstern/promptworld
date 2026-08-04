# Plan — spec 117, polish-session project skill

**Spec:** `specs/117-polish-session-skill/spec.md` · **Board task:** TASK-198

## Constitution check (`.specify/memory/constitution.md` v1.3.0)

| Principle | Compliance |
| --- | --- |
| I. Artifact-Grounded Action | The deliverable exists to make freeform work artifact-grounded — its central rule is that a decision with a file:line diagnosis lands on the card *before* implementation. It ships a card template so the paper trail is scaffolded rather than remembered. |
| II. One Task, One PR | TASK-198 → `task-198-polish-skill` → one PR. Spec 117's stub landed on main first, per the very rule this spec encodes. |
| III. Gates Over Assertions | FR-009's probe turns FR-006's binary trap and FR-005's footprint number into readings rather than reminders. Advisory by design — the session gate's non-blocking contract governs, and making any of it block is out of scope. |
| IV. Grounding Freshness | The skill adds no `docs/wiki/` sources. If the branch touches a file any note pins, it re-pins in-branch per spec 069. |
| V. Model-Tiered Workflow | Doctrine content is policy, authored at the planning tier (`claude-opus-5`) — it *is* the lifecycle's own statement of itself, the non-implementation verb class Principle V assigns to that tier. The one code artifact is a single self-contained Node probe with no cross-package or concurrency surface. Recorded on the board on delivery. |

## Design

```
.claude/skills/polish-session/
├── SKILL.md                       FR-001..FR-007 — the loop and the four constraints
├── templates/
│   └── session-card.md            FR-008 — scaffold for the session's single card
└── scripts/
    └── session-status.mjs         FR-009 — binary freshness + wiki footprint
```

Modeled on `.claude/skills/tui-frames/` and `.claude/skills/player-docs/`, the repo's two
project-skill precedents — same frontmatter keys, same `scripts/` placement, same "check first"
idiom surfaced in the `description` so an agent sees the probe without opening the file.
`templates/` is new; no existing skill ships one, and FR-008 needs a file a session can copy
rather than prose it must transcribe.

### What the skill is, and what it deliberately is not

It is **the flow plus the four things that bite**. It is *not* a retelling of TASK-195. The run's
decision log is ~600 lines of narrative; the skill's job is to compress it into rules a session
follows and a citation it can read when it wants the story (FR-007). Spec 113's R3 applies with
full force here: a skill that reads as bureaucracy gets skimmed, and a skimmed skill fails FR-002.

Ordering inside `SKILL.md` is load-bearing. **Stub-first comes before the loop**, not inside step
5 where escalation actually happens — a session that reads it at step 5 has already made the
mistake. The remaining constraints attach to the step they govern.

### The status probe

`node .claude/skills/polish-session/scripts/session-status.mjs --check`

Two readings, both of numbers a polish session otherwise guesses at:

1. **Binary freshness.** Compare the mtime of the worktree's `promptworld` binary against the
   newest mtime among tracked `*.go` files and `go.mod`/`go.sum`. Stale, missing, or
   older-than-source → nonzero. This is the TASK-195 trap made mechanical: `go build ./...`
   populates the build cache and never rewrites the `-o promptworld` artifact, so a binary can be
   arbitrarily old while every build command in the transcript succeeded.
2. **Wiki footprint.** Shell out to `node scripts/check-merge-drift.mjs session --json`, find the
   `branches[]` entry whose `worktree` is the current one, and report its `wikiFootprint` with the
   headroom to the gate's threshold. **Read, never recomputed** — a second derivation of the same
   number is a second thing to keep in sync, and the gate is already the authority.

Design choices worth stating:

- **Read-only.** The probe never builds, never dumps, never writes. A session that wants a fresh
  binary runs the build itself; a probe that fixed what it measured would make its own green
  result meaningless.
- **Advisory exit code.** Nonzero means "something here needs your attention before live QA," not
  "stop." Nothing in the harness consumes it.
- **Threshold not duplicated.** The probe parses the threshold out of the gate's own finding
  message when it fires, and otherwise reports the raw count. It does not hardcode `30`.

## Risks

- **R1 — the skill rots against the gates it describes.** If `wiki-footprint`'s threshold or the
  session gate's report shape changes, prose naming `30` goes stale. Mitigation: the skill teaches
  *reading the number off the report*, and names the threshold only as the current calibration
  with its source. The probe reads the JSON rather than re-deriving.
- **R2 — over-prescription.** The flow's value is speed; a skill that makes a polish session feel
  like a spec cycle defeats it. Mitigation: exactly four constraints, each earned by an observed
  failure, and everything else stated as the short version with a citation.
- **R3 — the card template drifts from spec 087's card format.** Mitigation: the template carries
  a gist and "As a …" use cases, and points at CLAUDE.md's card-format section rather than
  restating the rules.
- **R4 — the probe's binary check misfires in a non-worktree checkout or where the binary is
  built elsewhere.** Mitigation: absent binary is reported as a distinct, explained state ("build
  it before QA"), not a failure of the check itself.

## Phases

See `tasks.md`. Three phases: author the skill, add the probe and template, then verify and
ground.

# Plan — spec 113, tui-frames project skill

**Spec:** `specs/113-tui-frames-skill/spec.md` · **Board task:** TASK-190

## Constitution check (`.specify/memory/constitution.md` v1.3.0)

| Principle | Compliance |
| --- | --- |
| I. Artifact-Grounded Action | The deliverable is itself an artifact that makes other work artifact-grounded — it routes agents to generated evidence instead of recollection. |
| II. One Task, One PR | TASK-190 → `task-190-tui-frames-skill` → one PR. |
| III. Gates Over Assertions | FR-008's guard script is the mechanism; the skill's rule against hand-edited frames is checkable rather than merely stated. |
| IV. Grounding Freshness | The skill adds no `docs/wiki/` sources. If the branch touches a file any note pins, it re-pins in-branch per spec 069. |
| V. Model-Tiered Workflow | Single-package authoring — one `SKILL.md` and one small Node script, no cross-package or concurrency surface. **Sonnet (`claude-sonnet-5`)**, the default tier, via the `spec-implementer` agent definition. The doctrine content is authored here at the planning tier and handed over verbatim; the implementer types and tests it. |

## Design

```
.claude/skills/tui-frames/
├── SKILL.md                    frontmatter + the doctrine (FR-001..FR-007)
└── scripts/
    └── check-frames.mjs        FR-008 guard: dump to temp, diff vs committed
```

Modeled on `.claude/skills/player-docs/`, the repo's existing project-skill precedent —
same frontmatter keys, same `scripts/` placement, same "check first" idiom surfaced in the
`description` so an agent sees the probe without opening the file.

### The guard script

`node .claude/skills/tui-frames/scripts/check-frames.mjs --check`

1. `go run ./cmd/promptworld frames --dump --out <tmpdir>` — the harness already supports
   `--out`, so no production change is needed.
2. Compare `<tmpdir>` against `docs/design/tui/frames/`, ignoring `README.md` (never
   generated, never touched by `--dump`).
3. Exit 0 when identical. Exit nonzero listing every differing, missing, or extra file.

Dumping to a **temp dir rather than in place** is the load-bearing choice: it means running
the check can never itself mutate the committed matrix, so a green result is evidence rather
than a self-fulfilling side effect. It also leaves the working tree clean, which the ACs
require.

The script is **advisory and agent-invoked**, not a PR gate — wiring the matrix into
`check-tui-design.mjs` remains a separate, unmade decision (spec 112's out-of-scope list).

## Risks

- **R1 — the skill drifts from the harness.** If a fixture, state, or size changes, prose in
  `SKILL.md` naming them goes stale. Mitigation: point at `frames --list` as the live source
  and keep enumerations illustrative, never load-bearing. The one enumeration worth spelling
  out is the four sizes with their reasons, since the *reason* is what an agent needs and
  `--list` does not explain it.
- **R2 — `go run` cost.** The guard builds the binary. Acceptable for an agent-invoked check;
  it is not on a hot path.
- **R3 — over-prescription.** A skill that reads as bureaucracy gets skimmed. Keep it short,
  imperative, and lead with the single most valuable instruction: read the frame file.

## Phases

See `tasks.md`. Two phases: author the skill, then verify and ground.

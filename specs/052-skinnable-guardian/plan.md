# Implementation Plan: Skinnable guardian persona — de-theme the angel fiction, persona as data

**Branch**: `052-skinnable-guardian` | **Date**: 2026-07-25 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/052-skinnable-guardian/spec.md`

## Summary

Make the fiction layer data: grow `internal/skin` into the runtime skin
substrate (default Guardian token table + per-world boot-frozen `skin.json`
with identity fields, string overrides, stage identities, and a persona-voice
text composed at the existing SOUL-fragment seam beneath the untouched fixed
frame), thread resolved skin facts through status to the TUI/CLI, sweep every
user-facing fiction literal onto the lookup (TUI, chronicle grammar subject
lines, CLI vocabulary with compat aliases, LLM prompt constants), freeze all
serialized vocabulary in place with annotations, rename pure-Go identifiers
(`internal/metatron` → `internal/guardian`), and ship one example alternate
skin. The event log stays skin-free (spec ruling 1) so mechanics equivalence
across skins is testable. Contract-first sequencing (D2): the lookup API +
token table + `patterns/skin-tokens.md` runtime section are the published
contract TASK-115/117 consume.

## Technical Context

**Language/Version**: Go 1.24 (repo toolchain; no new deps)

**Primary Dependencies**: existing packages only — `internal/skin` (grows), `internal/metatron`→`internal/guardian`, `internal/persona`, `internal/tui`, `internal/tool`, `internal/world`, `cmd/promptworld`

**Storage**: one new optional per-world file `skin.json` (boot-frozen; capabilities.json fallback discipline); no event/snapshot format changes

**Testing**: `go test -race ./...`; token-completeness test; repo-wide fiction-denylist sweep test over rendered surfaces (SC-001/002); hostile-skin adversarial battery extension (SC-005); deterministic two-skin mechanics-equivalence test (SC-004); pre-feature-world replay/compat test (SC-003)

**Target Platform**: daemon + terminal client (darwin/linux)

**Project Type**: single Go module, cross-package sweep (the one architectural slice of Lane 1)

**Performance Goals**: token lookup is map access; no render-path or turn-path regression

**Constraints**: serialized vocabulary frozen (research R4 list is normative); fixed frame appended last on every path (spec 021 INV-1); event log skin-free (FR-005); same-PR design-doc amendments (skin-tokens.md runtime contract + re-pins)

**Scale/Scope**: ~20 code files touched + package rename churn; est. 1,500–2,500 LOC delta incl. tests; 2–4 design pages amended

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Artifact-Grounded Action** — PASS: spec/plan/tasks under `specs/052-*`; board TASK-121 linked pre-implementation; the three spec rulings cite the operator pivot + doctrine artifacts; sweep inventory persisted in research.md R8.
- **II. One Task, One PR** — PASS: TASK-121 = `.worktrees/task-121`, one branch, one PR (contract + sweep + rename + docs together; the CONTRACT publishes at merge, which is what unblocks Lane 3).
- **III. Gates Over Assertions** — PASS: denylist sweep test + adversarial battery + equivalence test are mechanized ACs; `check-tui-design.mjs --changed`; spec-bridge gate.
- **IV. Grounding Freshness** — PASS (planned): merge touches sources of `docs/wiki/metatron.md`, `tui-client.md`, `cli-promptworld.md`, others → wiki-update + player-docs regen (page renames) in re-ground.
- **V. Model-Tiered Workflow** — PASS: planned on Fable 5; implementation on **Opus 4.8** — rubric: cross-package/architectural change touching prompt composition beneath the injection-soundness doctrine (`internal/metatron` turn assembly), doctrine-adjacent behavior (fixed-frame adjacency, event-log-skin-free ruling), plus a repo-wide rename. Exactly the senior-tier profile.

**Post-Phase-1 re-check**: PASS — one new file format (skin.json) following two established loader precedents; no new abstractions beyond the skin value type; rename is compiler-verified churn. No Complexity Tracking entries.

## Project Structure

### Documentation (this feature)

```text
specs/052-skinnable-guardian/
├── plan.md              # This file
├── research.md          # R1–R7 decisions + R8 normative sweep inventory
├── data-model.md        # skin bundle format, token model, frozen-vocabulary sets
├── quickstart.md        # validation guide
├── contracts/
│   └── skin-contract.md # THE published contract TASK-115/117 consume
└── tasks.md             # Phase 2 output
```

### Source Code (repository root)

```text
internal/skin/           # grows: Skin type, default table, Load(worldDir), token
│                        #   resolution, stage identities gain skin dimension
internal/guardian/       # renamed from internal/metatron (frozen strings annotated);
│                        #   SetSkin (boot-frozen); voice at SOUL seam in turnSystemPrompt;
│                        #   de-themed prompt constants w/ name substitution
internal/persona/        # DefaultCharter rewritten guardian-voiced; genesis seeding
internal/tool/           # derive.go glosses re-worded (working); ids frozen
internal/sim/            # Go identifier renames ONLY where unserialized; tags frozen
internal/ipc/            # method names frozen; status gains additive skin fields
internal/scribe/         # morgue display line swept
internal/tui/            # full literal sweep through status-carried skin facts;
│                        #   Type-column family alias; grant summary "workings"
cmd/promptworld/         # canonical guardian/work subcommands + hidden aliases;
│                        #   usage text; stagesLadder prose
examples/skins/          # raven.json + README (FR-014)

docs/design/tui/
├── patterns/skin-tokens.md  # gains the runtime-contract section (its own requirement);
│                            #   token index promoted to the default table's doc twin
└── (pages whose token index/labels change) + re-pins

README.md                # fiction mentions swept
```

**Structure Decision**: the skin substrate centralizes in `internal/skin`
(declared beachhead); everything else is consumption + sweep. The package
rename lands as its own commit(s) late in the branch so review can separate
mechanical churn from behavior.

## Complexity Tracking

No constitution violations — table intentionally empty.

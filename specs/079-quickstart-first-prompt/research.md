# Research: How player-docs content changes work (spec 079)

**Verified against the working tree, 2026-07-26.** This note answers the one
question that shapes the whole feature: where does a content change to
`docs/player/` durably live, so the freshness gate stays coherent AND the next
regeneration does not erase it?

## R1 — Generated-from-source or hand-authored-with-pins? Both, precisely:

`docs/player/*.html` pages are **LLM-authored-at-regeneration-time projections
of declared sources** — not mechanically generated from templates, and not
free-hand documents either. The authority is the skill's own procedure
(`.claude/skills/player-docs/SKILL.md`, "Procedure (check-first, mandatory
order)"):

- The **freshness gate** (`scripts/check-freshness.mjs`) checks *provenance
  only*: each `promptworld-docs:source` meta tag's `<path>@<40-hex>` pin is
  compared against the source's **current** pin — `verified_against:`
  frontmatter for `docs/wiki/*` sources, `git log -1 --format=%H -- <path>`
  for every other path (`currentPinFor`, check-freshness.mjs:114). The gate
  never diffs page prose against source content. Consequence: **editing a
  page's body never makes it stale, and editing a source always does.**
- The **regeneration contract** (SKILL.md step 2b) is what makes free-hand
  content mortal: when a page goes stale, the skill rewrites its prose "as a
  plain-language projection of those sources — no independently asserted
  facts." Content not derivable from the page's declared sources is
  *legitimately dropped* at the next regeneration.

## R2 — The durable home for this feature's content (decision)

A content change survives regeneration only when it lands in **three
coordinated places, all in this task's branch**:

1. **The page HTML** — the immediate content, with any newly-drawn-on source
   declared as a `promptworld-docs:source` meta tag at its *current* pin
   (SKILL.md step 2c: "Add a tag for any newly-drawn-on source").
2. **The SKILL.md editorial contract** — the page → source mapping table row
   plus an editorial shape note. Precedent already in the file: the TASK-114
   paragraph (screen/keys pair shape) and the TASK-68 paragraph (stage-quartet
   shape). This is the ONLY artifact the next regeneration reads for *what a
   page must contain structurally*; without it, a future regeneration of
   `getting-started.html` reproduces the old shape and the first-prompt step
   vanishes.
3. **A declared source carrying the facts** — the projection rule ("no
   independently asserted facts") means every quoted string and every claim in
   the new content must be derivable from a declared source. No new grounding
   needs to be *written*: the facts already exist (R3, R4). This is why the
   task is content-only and triggers **no wiki re-pins**.

Rejected alternative: putting the sample ask in the page as free prose without
declaring `docs/wiki/skin.md` as a source. The freshness gate would still pass
(it doesn't read prose), but the content would be an independently asserted
fact — erased at next regeneration, and dishonest as provenance. The meta tag
+ mapping row is the honest, durable form.

## R3 — The `skin.guardian.example_ask.*` token family (inventory)

Compiled default table, `internal/skin/skin.go:80-88` (spec 063 US5/D9); the
wiki note `docs/wiki/skin.md` (pinned `verified_against:
31c893e0406653197e467a89b2fdb96f0bcf2ee0`) documents the family, the
resolution order, and the help-overlay consumption — making it the right
*declared source* for player docs (a `docs/wiki/` frontmatter pin, stable
across unrelated `skin.go` churn, unlike a plain-file pin on `skin.go`
itself):

| token (`skin.guardian.example_ask.<tool-id>`) | default-skin value |
|---|---|
| `send_vision` | `"show Ash a vision of the fire dying"` |
| `send_omen` | `"send everyone an omen tonight: stay near the fire"` |
| `monitor_and_act` | `"watch for anyone going hungry, and warn them"` |
| `cancel_order` | `"release the watch on the fire"` |
| `work_miracle` | `"work a working: grant Ash food from thin air"` |
| `pause` | `"pause the world"` |
| `start` | `"start the world again at 4x"` |
| `adjust_speed` | `"slow the world down to 1x"` |
| `explain` | `"what does a vision cost?"` |

How the `?` guardian section renders it today: one canned ask per **granted**
verb, keyed by the frozen tool id, resolved through the active world's skin
(`docs/design/tui/overlays/help.md` D9 rows; `Skin.ExampleAsk(toolID)`,
skin.go:216; byte-identity table row at help.md:252).

**Stage-1 constraint:** the stage-1 tool ceiling is pinned to exactly
`send_omen`, `send_vision`, `monitor_and_act`, `cancel_order`
(`specs/046-curriculum-ladder/contracts/stage-gating.md`, ratified amendment
2026-07-25 + "Pinned (in-PR…)" clause; `work_miracle` and the clock verbs are
excluded, and `explain` is not in the pinned set). A first-session sample ask
in getting-started must therefore name one of those four verbs — recommended:
`send_vision` → `"show Ash a vision of the fire dying"` (the family's own
canonical exemplar, quoted verbatim in both skin.go's comment and
docs/wiki/skin.md).

## R4 — Per-stage first-session facts already live in declared sources

Each stage page already declares the spec 046 files that carry everything a
do-this-then-this block needs — **no new sources required on stage pages**:

- stage-1: default stage for `new`; first-night exercise (keep the village
  alive to dawn of day 2; at least one player-directed act before nightfall) —
  `contracts/exercises.md`, `contracts/stage-gating.md`.
- stage-2: charter.md binds from the next turn; the-law exercise (norm adopted
  while a player-authored charter revision is in force) — same contracts.
- stage-3: skill files compose, grantable manifest opens; unlock proof = a
  personally granted tool contributes to a pass — `stage-gating.md`, spec.md.
- stage-4: full roster incl. capstone; **no exercise gates it** — spec.md,
  `stage-gating.md`.

The ask itself (quoting a token value) stays on getting-started; stage pages
**link** to that step rather than re-quoting token values, keeping their
source sets unchanged.

## R5 — Edge cases, resolved from the artifacts

- **Skins other than default.** Static pages cannot resolve a world's
  `skin.json`. Honest form: print the *default Guardian skin's* value (that is
  what `docs/wiki/skin.md` documents) and say so, noting that the in-game `?`
  overlay's guardian section always shows *your* world's own phrasing
  (resolution order: world override → default table, skin.md "How it works").
  Never present the printed string as universal.
- **Worlds created without a scenario.** getting-started's own flow
  (`promptworld new demo --seed 42`) creates a plain world; first-session
  blocks must work with no `--scenario` and no exercise tab. Exercise steps
  are phrased as a when-you're-ready follow-on (stage-4 explicitly has none).
- **The freshness gate after content edits.** Body edits never trip the gate
  (R1). The gate *would* break on: a malformed new source tag (grammar
  `<path>@<40-hex-lowercase>`), a stale pin recorded for the newly added
  source, or a source tag added to `index.html` (forbidden). Acceptance is
  simply probe exit 0 after the edits, with the 8 untouched pages
  byte-identical.
- **Future `skin.md` re-pin stales getting-started.** Accepted and by design:
  when `docs/wiki/skin.md` is next re-verified, getting-started goes stale and
  the regeneration must *reproduce* the first-prompt step — which is exactly
  what the SKILL.md editorial amendment (R2 item 2) guarantees.

## R6 — Pre-existing mapping-table drift (noted, bounded)

SKILL.md's mapping table lists 4 sources for `getting-started.html`; the page
on main declares 12 (lanes through spec 072/074 added tags without table
rows). The table is the "starting, editorial contract" and SKILL.md itself
mandates row+tag move together. This task reconciles the rows **only for the
five pages it touches** (getting-started + four stage pages) to their actual
declared sources post-change; a corpus-wide reconciliation is out of scope
(a candidate for the parked "player-docs pin de-churn" watch item).

## R7 — Gate expectations for this branch (verified)

- No wiki note lists `docs/player/*` or `.claude/skills/player-docs/*` as a
  source (`grep -rn "docs/player\|skills/player-docs" docs/wiki/*.md` — empty)
  → the pr gate's `wiki-repin-missing` probe is expected clean; **no wiki
  re-pins belong in this branch** (content-only).
- `check-freshness.mjs --check` exits 0 on main today (13 fresh) — the
  baseline the branch must return to after its edits.
- No Go code, no `docs/design/tui/` pages change → `check-tui-design.mjs` not
  applicable; `go test ./...` runs per doctrine and is expected untouched.

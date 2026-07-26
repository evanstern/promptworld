---
title: Analysis — Tile Vocabulary Expansion
aliases: [Expand Available Tiles Analysis, Glyph Vocabulary Analysis]
tags: [analysis, game-ui, tui, glyphs, color, fonts]
type: analysis
created: 2026-07-26
updated: 2026-07-26
related: ["[[Game-UI-UX]]", "[[TUI-and-Roguelike-UI-Craft]]", "[[Dwarf-Fortress-Interface]]", "[[RimWorld-Interface]]", "[[LLM-Agent-Sim-Interfaces]]", "[[Recurring-Interface-Patterns]]", "[[Analysis-Teaching-Game-TUI]]"]
---

# Analysis — Tile Vocabulary Expansion

_What does this branch's corpus support for expanding promptworld's tile/glyph vocabulary —
new tile types, alternative colors, avatar characters, and fonts — as the map surface grows
(the "expand available tiles" work)?_

## Verdict

**Expand by borrowing conventions, not by inventing glyphs; spend color before spending new
characters; treat CP437/roguelike vocabulary as the dictionary; and if a richer surface is
wanted (web renderer), the corpus says ship it as a swappable "tileset" layer over the same
grid rather than a second rendering model.** Concretely: (1) new *terrain* tiles should come
from the shared roguelike symbol convention and CP437's shading/box repertoire, because that
vocabulary is what makes ASCII fast to read; (2) new *states* of existing things should be
style variants (color/bold/faint/underline) of the glyph already learned, never a new
character — the pattern the project's cold-fire/dying-fire/damaged-wall treatments already
follow and the corpus endorses; (3) *avatar* identity should stay letter-based at current
scale, with emoji/status badges reserved for an inspector layer, not the grid; and (4) fonts
are an architectural choice, not a cosmetic one — the corpus's strongest font fact is
Cogmind's mixed-dimension scheme (square-ish cells for the map, narrow fonts for text panels)
plus the CP437-coverage requirement.

## Reasoning

### What the current vocabulary is (the baseline being expanded)

Promptworld's map today (read from `internal/tui/views.go:1581-1644` and the tile switch at
`views.go:1290-1330`; cited by path — project surface, outside this vault by design):

- **Terrain:** water `~` (ANSI 4), tree `♠` (2), forage `"` (3), rock `^` (245), depleted
  ground `,` (faint 240), grass `·` (dim), path `·` (warm tan 137 — same glyph as grass,
  color-differentiated).
- **Structures:** fire `▲` (208; dying 202; cold = faint 240), shelter `⌂` (130), oven `▣`
  (166), chest `☐` (136), pile `%` (178), walls `▤` plank / `▩` stone (250; damaged faint
  240), den `ᴥ` (5), grave `✝` (faint 244).
- **Agents:** initial letter, lowercase when asleep; `†` dead; `G` gru (bold 196); condition
  overlays as *style* variants (critical = bold underline red 1; suppressed mind = faint 135).
- Night dims the palette rather than hiding the world; the legend line and `?` overlay render
  from one shared glyph table.

This baseline is already corpus-correct in three load-bearing ways: it uses the shared
convention where one exists (`%` for a ground stash, `@`-family letters for actors, `~` for
water), it distinguishes states by style rather than new glyphs (cold fire, damaged wall,
asleep case), and it keeps one legend source. Expansion should preserve exactly those three
properties.

### New tile types: draw from the convention dictionary, in priority order

The corpus's central fact about why ASCII maps read fast is that players leverage "years of
experience distinguishing a very specific set of symbols" — glyphs stay identifiable "even
when densely clustered and/or at small sizes" ([[TUI-and-Roguelike-UI-Craft]], Ge's ASCII
vs. Tiles). That advantage only accrues to *conventional* glyphs; a novel character is a
vocabulary word the player must learn (Ge's documented drawback: "many players have
difficulty making that conceptual leap"). So the expansion rule the corpus supports is:

1. **First choice — the RogueBasin convention set** ([[TUI-and-Roguelike-UI-Craft]]): `.`
   floor, `<`/`>` level/stairs transitions, `?` scrolls/documents, `!` potions/consumables,
   lowercase letters for creatures, `@`-family for actors. DF's classic set extends this:
   `☺` dwarf, `c` cat, `d` dog — animals as lowercase letters is the deepest convention in
   the genre ([[Dwarf-Fortress-Interface]]). If promptworld adds wildlife, crops, or item
   drops, these are the pre-learned characters to take.
2. **Second choice — CP437's chrome and texture repertoire** ([[TUI-and-Roguelike-UI-Craft]]):
   Ge documents CP437 as "one of the more common choices for ASCII layout… used by games
   like Dwarf Fortress," supplying box-drawing and *shading* glyphs (`░ ▒ ▓`). Shading
   characters are the corpus-grounded candidates for new ground covers or density gradients
   (marsh, snow depth, crop maturity) because they read as texture, not object.
3. **Last resort — novel Unicode.** The project has already gone here (`ᴥ` den, `▣` oven,
   `✝` grave) and it works *because each is visually unlike everything else on the grid* —
   the property to preserve. Every addition must clear the existing set at small sizes and
   in dense clusters, which is exactly the test Ge's side-by-side comparison documents.

Two corpus cautions bound this: Cogmind gives "every item custom ASCII art" but that is
inspector-panel art, not map glyphs — the *map* stays a small stable vocabulary; and the
legend/`?` overlay discipline (one shared table) is what keeps a growing vocabulary
learnable — RimWorld's "intuitive color coding, where every object and pawn is recognizable
instantly" is credited in comparisons as the accessibility difference vs. DF
([[RimWorld-Interface]] reception facts in [[_grounding]]).

### Alternative colors: a semantic palette, not a decorative one

The corpus is unusually consistent about what color is *for* on these surfaces:

- **Severity is a fixed 4-step language.** RimWorld codes every event blue (good) / grey
  (neutral) / yellow (bad) / red (direct threat), and reuses the same language on its
  History graph ([[RimWorld-Interface]]). Promptworld's overlays already trend this way
  (red = physical danger, cooler hue = cognitive telemetry); any *new* color should first be
  asked "which severity band is this?" before it gets a novel hue.
- **State changes recolor, they don't re-glyph.** RimWorld 1.5's mood glow recolors the
  colonist bar below break thresholds; promptworld's dying fire (208→202) and damaged wall
  (250→faint 240) are the same move. The corpus supports expanding this into a documented
  *dimming/warning grammar*: faint = spent/inert (cold fire, grave, depleted), warm-shift =
  deteriorating, bold+underline = critical.
- **Stay within the 16-color semantic core where meaning matters.** Terminals expose 16
  user-themeable ANSI colors; 256/truecolor exist above them ([[TUI-and-Roguelike-UI-Craft]],
  Textual). Promptworld already mixes both (semantic 1–5 plus 256-palette earth tones).
  The corpus-grounded rule: *meaning-bearing* colors (danger, severity, faction) should sit
  on the themeable 16 so user themes don't break semantics; *material* colors (tan pile,
  brown shelter, goldenrod chest) can use the 256 palette since they're decorative
  distinctions, not warnings.
- **Externalize the scheme.** Cogmind's message colors live in an external text file,
  documented explicitly as a colorblind-accessibility aid ([[TUI-and-Roguelike-UI-Craft]]).
  Promptworld's ratified skin-token contract (D2 in [[Analysis-Teaching-Game-TUI]]) is the
  same mechanism; new tile colors should be born as tokens, not literals — the corpus and
  the operator decisions agree here.
- **Never color alone.** Cogmind's redundant multi-channel rule (color + symbol + sound +
  animation for important events) and WCAG H86's text-alternative requirement for symbols
  ([[TUI-and-Roguelike-UI-Craft]]) mean any new color distinction that *matters* needs a
  second channel — glyph case, the legend, or a feed line. Path-vs-grass (same `·`, color
  only) is the current vocabulary's one violation of this; fine for a low-stakes
  distinction, but not a pattern to extend to anything a player must act on.

### Avatar characters: letters at this scale; emoji belong to the inspector, not the grid

The corpus offers three avatar models: DF/roguelike letter glyphs, RimWorld's portrait bar +
pawn silhouettes, and Smallville's sprite avatars with **emoji speech bubbles** — the LLM
translates each agent's current action into emoji above the avatar's head, click to expand
to full text ([[LLM-Agent-Sim-Interfaces]]). The scale evidence says per-agent inspection
holds to ~25 agents; promptworld's 8–20 villagers sit inside that
([[Recurring-Interface-Patterns]], [[Analysis-Teaching-Game-TUI]]).

Position: **keep case-coded initials as the grid identity** — they're the roguelike
convention, they carry two free channels (letter = who, case = awake/asleep), and initials
stay legible clustered where sprites/emoji don't in a cell grid. The Smallville emoji
pattern is still worth stealing, but at the *inspector/strip* layer: an ambient
"current-action" emoji or short badge in the villager strip / roster row, expanding on
selection — that is literally the documented Smallville chain (ambient glyph → click → full
action → memory stream), which the TUI's roster → detail drill-down already mirrors. A web
renderer could put the emoji bubble over the avatar exactly as Smallville does; a terminal
cell grid cannot (double-width emoji break cell alignment — the corpus's cell-geometry facts
make any double-width character on the map a layout hazard, which is also the argument
against emoji *as* map glyphs).

One convention collision to resolve deliberately: `G` for the gru sits in the agent-letter
namespace. The roguelike convention would give a hostile creature a *colored* letter (it is
bold red 196, which does the work), but as villager counts grow, initials will eventually
collide with `G` — the corpus's fix is DF's: reserve a case/color combination for
non-villager actors rather than a letter.

### Fonts: an architecture decision with three grounded rules

The corpus's font facts all come from [[TUI-and-Roguelike-UI-Craft]] and
[[Chat-and-Agent-Console-Rendering]]:

1. **Mixed dimensions are the documented ideal.** Cogmind renders the map in square cells
   (two terminal cells per map cell) and text panels in narrow fonts — "the best of both
   worlds." Promptworld's TUI already writes a space after every tile (`views.go`
   `tile(x, y) + " "`), achieving the square-map effect. A web renderer can and should make
   this explicit: a square-cell map font and a narrow UI font are *two* font choices, not one.
2. **CP437 coverage is the selection gate.** Ge documents CP437 as the common layout choice;
   promptworld's glyph set (`⌂ ▣ ▤ ▩ ░`-family, `♠`, `ᴥ`, `✝`) means any candidate font must
   cover box-drawing, shading, and these specific codepoints at the size the map renders.
   For a web surface this argues for a monospace font with deliberate CP437/Unicode
   coverage; the corpus doesn't name specific fonts (a gap — see below), only the coverage
   and dimension requirements.
3. **Scaling means cell size, never grid extent.** Cogmind's upscaling program changes
   font/cell size against a fixed cell grid (160×60; guaranteed 50×50 map view that gameplay
   is balanced against). For promptworld the transferable rule: pick the guaranteed-visible
   map area first, then let font size be the responsive variable — don't let font choice
   silently change how much world a player sees.

### If "expand available tiles" means a richer graphical tileset

The corpus's strongest single fact here is DF's architecture: Classic and Premium are one
codebase — "the switch… is just swapping out some glyphs. The grid structure underneath is
the same everywhere" ([[Dwarf-Fortress-Interface]]). Combined with Ge's market observation
("the average player will always choose to play with a tileset") and Cogmind shipping tiles
as the accessibility alternative with ASCII as default, the grounded architecture for any
web/graphical expansion is: **one grid model, swappable glyph/tile skins** — ASCII glyphs
and any pixel/sprite tileset are projections of the same tile table. This also matches the
project's existing one-legend-source discipline and the skin-token direction. The
cautionary half of the same evidence: DF's Premium redesign drew criticism for visual noise
in busy scenes and scattered information ([[Dwarf-Fortress-Interface]]) — richer tiles raise
the noise floor; Sylvester's "noise is signal that fails to transmit meaning" applies to
tile art exactly as to text ([[RimWorld-Interface]]).

## Tensions & tradeoffs

- **Convention vs. fiction.** The roguelike dictionary (`!` potion, `?` scroll) carries
  meanings from *other* games; promptworld's fiction may want glyphs those conventions
  don't cover. Taking the convention buys instant readability for genre-literate players at
  the cost of some fiction fit; the corpus (Ge) is clear the readability is real, but
  promptworld's actual audience may be less genre-literate than Cogmind's, which weakens
  the "pre-learned" argument and strengthens the legend/`?`-overlay channel instead.
- **256-color earth tones vs. themeability.** The existing material palette (130/136/137/
  166/178/202/208) is exactly what makes the map read warm and physical, but those codes
  ignore user terminal themes and can collapse on 16-color terminals. The verdict's
  "semantics on the 16, materials on the 256" line preserves both, but it forfeits
  theme-consistency of the material colors — a real cost the skin-token layer can only
  partially recover.
- **Emoji ambience is genuinely good UX in the sims.** Confining emoji to the
  strip/inspector gives up some of Smallville's at-a-glance charm on the map itself. The
  counter-case: on a web renderer with proper positioning, over-avatar emoji bubbles are
  proven ([[LLM-Agent-Sim-Interfaces]]) and could coexist with letter glyphs. The tension is
  renderer-specific; the verdict holds firmly only for the terminal surface.
- **Every new tile type spends legend budget.** The legend line already carries phase,
  viewport, glyph key, notes, pile zones, and chest contents on one clipped row. The corpus's
  small-stable-vocabulary discipline is partly *about* this: vocabulary growth degrades the
  teaching surface before it degrades the map.

## Confidence & open questions

High confidence on the glyph-selection rules, the style-variant-before-new-glyph rule, the
severity color language, and the one-grid/swappable-skins architecture — these are each
multiply grounded. Moderate on the avatar position: it is renderer-dependent, and the corpus
documents no cell-grid game that *tried* emoji-on-map and failed; the double-width hazard is
inferred from the cell-geometry facts, not from a documented failure.

Open (gaps in the corpus, not decisions to smuggle):

1. **No specific font names are grounded.** The corpus gives requirements (CP437 coverage,
   square-vs-narrow dimensions, readability at small sizes) but never names concrete
   candidate fonts for a web surface. A small follow-up research pass (e.g., DF community
   tileset/font practice, terminal-font coverage surveys) would close this before a web
   renderer commits.
2. **Colorblind-safe palette specifics.** Cogmind externalizes colors *for* colorblind
   players, and RimWorld's 4-color severity language is hue-dependent (red/yellow/blue/grey),
   but the corpus contains no tested colorblind-safe palette for either. The skin-token
   contract is the right home; the palette itself is ungrounded.
3. **How natural-language state renders at density** — the MOC's standing open question —
   bears directly on the avatar/emoji layer and remains unresearched.

## Basis

- [[_grounding]] — §TUI & Roguelike Interface Design Principles (glyph conventions, CP437,
  ASCII-vs-tiles, fonts, color constraints, external color schemes, WCAG H86); §Dwarf
  Fortress (16-color classic set, glyph-swap architecture, tileset modders, Premium noise
  criticism); §RimWorld (severity color language, mood glow, "recognizable instantly"
  reception); §LLM-agent sims (emoji bubbles, inspector chain, scale threshold)
- [[TUI-and-Roguelike-UI-Craft]] — the substrate rules this analysis leans on hardest
- [[Dwarf-Fortress-Interface]] — one-codebase glyph-swap, tileset layer, redesign trade-offs
- [[RimWorld-Interface]] — severity coding, recolor-not-reglyph, noise doctrine
- [[LLM-Agent-Sim-Interfaces]] — Smallville emoji/avatar pattern, per-agent scale evidence
- [[Recurring-Interface-Patterns]] — progressive disclosure and scale-threshold rows
- [[Analysis-Teaching-Game-TUI]] — the ratified skin-token (D2) and linear-stream (D1)
  constraints this expansion must respect
- Project surfaces read for the baseline (cited by path, outside the vault by design):
  `internal/tui/views.go:1581-1644` (style table), `views.go:1290-1330` (tile switch),
  `docs/wiki/tui-client.md` (glyph inventory, legend discipline)

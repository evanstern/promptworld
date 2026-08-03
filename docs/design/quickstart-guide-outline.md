# promptworld Quickstart Guide — page outline

**Status:** proposal / not yet built
**Deliverable being outlined:** a wiki-style guide that takes a reader from
"I have never installed this" to "I have learned everything the game teaches."
**Modelled on:** the RimWorld [Quickstart Guides] page and the Dwarf Fortress
[DF2014 Quickstart guide].

---

## 1. What this is, and what it is not

**It is** a *guide* — one small, linked page-set with a single reading order, written
in the voice of someone sitting next to you while you play.

**It is not** a wiki. We are not documenting every entity, event type, and tool the way
the DF and RimWorld wikis do around their quickstarts. Those quickstarts can say
"see [Stockpiles]" because a thousand other pages exist. We have a different
luxury: `docs/player/` already exists — sixteen generated plain-language topic pages
regenerated from the code-grounded wiki. **That is our "rest of the wiki."** The guide
links *out* to it and never restates it.

So there are exactly three ways the guide handles a concept:

| Situation | What we do |
|---|---|
| Concept is background colour, mentioned once | Explain in ≤ 2 sentences, inline, no link |
| Concept is covered at the right depth in `docs/player/` | One sentence + link out |
| Concept is load-bearing, referenced 3+ times, and `docs/player/` covers it at the wrong altitude | It earns a guide-local page (§6) |

That last row is the only way a new page gets created. The bar is deliberately high.

---

## 2. What we are copying from the reference guides

Read both; these are the moves worth stealing.

**From Dwarf Fortress:**
- **UI concepts *before* setup.** DF teaches you how to read the screen before it
  teaches you to embark. We do the same — promptworld's four-pane screen is at least
  as alien as DF's.
- **Ordered-by-urgency subtasks.** "A Minimal Fortress" is a numbered survival list, not
  a feature tour.
- **Troubleshooting callouts embedded at the point of failure**, not banished to an
  appendix.
- **"What next?" as the closer** — the guide ends by pointing outward, not by claiming
  completeness.

**From RimWorld:**
- **Chronological narrative spine**: Setup → Landing → *Your first day* → *Your first
  night* → *The next few days* → *Your first battle* → *Your first winter*. The reader
  always knows where they are in game-time, not in feature-space.
- **The "avoid game-over pitfalls only" contract**, stated up front: this guide skips
  best-practice and min-maxing, and intervenes only where you'd otherwise lose.
- **A progression navbox in the footer** (Quickstart → Basics → Intermediate → Advanced)
  so the page advertises what comes after it.
- **Blunt, specific imperatives**: "PAUSE THE GAME!", "Do NOT hunt boomrats."

**What we deliberately do *not* copy:** DF's ~15–20k-word single page. We split (§4).

---

## 3. Where it lives

```
docs/guide/                     ← new, hand-authored, this proposal
  index.html                    ← the hub / quickstart proper
  01-setup.html
  ... (see §4)
docs/player/                    ← existing, GENERATED — never hand-edit
docs/wiki/                      ← existing, code-grounded corpus
```

**`docs/guide/` is a new directory, not an addition to `docs/player/`.** Reason:
`docs/player/` is produced by the `player-docs` project skill and is freshness-gated —
hand-authored pages dropped in there would be clobbered by the generator and would
confuse the gate. The guide is authored prose with a narrative spine; it is a different
kind of artifact and it gets its own home.

It **reuses `docs/player/`'s stylesheet variables verbatim** (same `--bg/--fg/--accent`
tokens, same dark-mode media query) so the two page-sets read as one documentation set,
then adds the wiki furniture in §5.

---

## 4. The page set, and the split policy

**Split rule:** a page is split when it exceeds **8,000 words**, and it is split at the
nearest H2 boundary. In practice no page below should get near that; the budget exists
so growth has a rule instead of a debate. Target lengths below are what we should
actually write.

| # | Page | Target words | Ceiling |
|---|---|---|---|
| — | `index.html` — **Quickstart Guide** (hub + the whole arc in brief) | 1,800 | 8,000 |
| 1 | Setup — install to a running world | 2,200 | 8,000 |
| 2 | Reading the screen | 2,400 | 8,000 |
| 3 | Your first night | 2,600 | 8,000 |
| 4 | Your first week | 3,200 | 8,000 |
| 5 | When someone dies | 1,800 | 8,000 |
| 6 | The four stages (the ladder) | 3,000 | 8,000 |
| 7 | What next | 1,200 | 8,000 |
| — | *Concept pages that earned it* (§6) | 600–1,500 ea. | 8,000 |

Total target ≈ **18–20k words** across the set — the same content mass as the DF page,
but never more than ~3k in front of the reader at once.

**Every page carries, in this order:** breadcrumb → infobox (right-floated) → lede →
ToC → body → "Continue to →" footer link → navbox.

---

## 5. The design system

### 5.1 Table of contents

MediaWiki-style: a bordered box, floated left under the lede, auto-numbered to **two
levels** (`1`, `1.1`) — never three. If a page's ToC would exceed ~12 entries, that is
the early-warning signal to split before the word count says so.

### 5.2 Callouts

Six types, each a left-border-coloured block with an icon glyph. Do not invent a
seventh.

| Callout | Use for | Example from the guide |
|---|---|---|
| **Note** | Neutral clarification | "The world keeps running after you detach." |
| **Tip** | Makes play nicer, safe to skip | "Write your seed down." |
| **Warning** | You will lose progress / a villager | "A villager who dies stays dead." |
| **Pitfall** | The specific mistake new players make | "Setting speed to max does *not* make villagers think faster." |
| **Under the hood** | Optional depth, collapsed by default (`<details>`) | "How the cognition horizon picks reflex vs. a model call." |
| **Design intent** | Why the game refuses to do the thing you want | "There is no undo, and that is the point." |

**Density target:** ≥ 1 callout per H2 section. Callouts are the main tool against
wall-of-text — they break the page rhythm visually.

### 5.3 Prose rules (enforced in review)

- **No paragraph over 4 lines** rendered at 70ch. Longer → split or convert to a list.
- **Every procedure is a numbered list**, one action per step, imperative mood.
- **Every "what does this mean" answer is a table**, not prose.
- Code/commands in `<pre>`, one command per line, with the expected result under it.

### 5.4 Infobox

Right-floated, on every page. Hub infobox carries: platform, what you need installed,
whether an AI model is required, save location, permadeath yes/no, typical session
length. Chapter pages carry: prerequisites, what you'll have when done, ~time.

### 5.5 Navbox

Footer table on every page, always identical, mirroring RimWorld's Guides navbox:

```
Quickstart Guide
  Progression   Setup • Reading the screen • First night • First week •
                When someone dies • The four stages • What next
  Concepts      The Guardian • Needs & neglect • The gru • Speed & thinking
  Reference     Command reference • Keys reference • World files • Troubleshooting
                (→ docs/player/)
```

### 5.6 SVG diagram inventory

Inline SVG, no external assets, styled with the same CSS custom properties so both
themes work. **Eight diagrams**, each earning its place by replacing prose that
otherwise doesn't work:

| ID | Page | Diagram | Replaces |
|---|---|---|---|
| D1 | Hub | **The three pieces** — daemon (always running) ↔ world save directory ↔ attachable client | Three paragraphs about why the game has no "quit" |
| D2 | Hub | **The influence funnel** — You → Guardian → (omens / visions / standing orders / designations) → Villagers → Events → Chronicle → back to You | The single hardest thing for a new player to grasp: you never touch a villager |
| D3 | Setup | **From zero to running** — flowchart: install Go → build → (model? yes/no branch) → `new` → `start` → `ui` | A branching install procedure |
| D4 | Reading the screen | **Screen wireframe** — labelled four-pane box with callout leader lines | A region-by-region text tour |
| D5 | First night | **A day in the village** — horizontal timeline: dawn → work → meeting → dusk → gru window → sleep → consolidation | Explaining that the day has a shape |
| D6 | First week / Needs page | **The neglect chain** — needs decay → no fire → exposure → wounded → gru attack becomes lethal | Why deaths are never bad luck |
| D7 | The four stages | **The ladder** — four rungs, each showing: what you get, what you must prove to climb | A table that reads worse than a picture |
| D8 | Speed page | **Reflex vs. thought** — a tick, and which decisions cost a model call | Why the game slows itself down |

---

## 6. Concept pages that earn their own page

Applying §1's bar. Each is referenced from three or more chapters and is either absent
from `docs/player/` or covered there at a different altitude.

| Page | Why it earns one | Referenced from |
|---|---|---|
| **The Guardian** | It is the *only* way the player touches the world. Named in every chapter. `docs/player/playing-via-metatron.html` exists but is written as a topic explainer, not as the guide's shared reference for charges, tools, and what it will/won't do. | Hub, 1, 3, 4, 6, 7 |
| **Needs & the neglect chain** | The whole survival layer. The guide has to say "warmth, food, rest" a dozen times. | 3, 4, 5 |
| **The gru** | The night antagonist; the reason the first night has stakes; recurs in deaths and in standing orders. *Conditional:* if the section stays under ~400 words, keep it inline in chapter 3 instead. | 3, 4, 5 |
| **Speed & thinking** | The most counter-intuitive system in the game (speed is a ceiling, not a promise). `docs/player/time-and-speed.html` covers the controls; this covers the *why*. | Hub, 2, 4, 6 |

**Explicitly NOT getting a page** (link out to `docs/player/` instead): the chronicle,
the map glyphs, `llm.json` setup, every command, every key, world files, troubleshooting.

---

## 7. Page-by-page outline

Headings below are the actual H2/H3s. `▸` marks a callout, `◆` marks a diagram.

### 7.0 `index.html` — Quickstart Guide (the hub)

> Lede, verbatim intent: *"This guide gets you from nothing installed to a village you
> understand. It skips depth, tips, and best practice — except where skipping them
> loses you the village."*

- **1 What you're getting into**
  - ◆ D1 The three pieces
  - You don't play a character. You have eight villagers and one intermediary.
  - The world runs whether or not you're watching.
  - ▸ Warning — permadeath, no reload, no undo.
  - ▸ Design intent — losing a villager is the story, not a failure state.
- **2 How you actually do anything**
  - ◆ D2 The influence funnel
  - Three sentences on the Guardian, then → *The Guardian* page.
- **3 The shape of this guide**
  - Numbered list of the seven chapters, each one line, each a link.
  - ▸ Tip — you can stop after chapter 3 and have a good time.
- **4 Before you start** — the checklist (Go; optionally a local model; ~20 min).
- **Continue to → Setup**

### 7.1 Setup

- **1 Install**
  - 1.1 Build the binary — `go build ./cmd/promptworld`
  - 1.2 Do you need an AI model? — ◆ D3 the branch
    - ▸ Note — no model = "reflex-only" villagers; the world still runs and is still
      worth watching for an hour.
    - ▸ **Fact-check before writing:** README says `ollama pull cogito:3b`; the
      generated player docs say `gemma4:latest`. Confirm the current fresh-world default
      against the code before this line is written.
- **2 Create a world** — `promptworld new demo --seed 42`
  - 2.1 What a world *is* (a copyable directory) — table of what's inside, link out to
    world files reference
  - 2.2 The seed — ▸ Tip write it down
  - 2.3 Your stage — new players start at The Voice, automatically; forward-ref to ch. 6
- **3 Start it** — `promptworld start demo`, then `promptworld status demo`
  - ▸ Pitfall — `start` detaches. Nothing appears to happen. That's correct.
- **4 Attach** — `promptworld ui demo`
  - ▸ Tip — `q` detaches, it does not stop the world. `promptworld stop` does.
- **5 If something went wrong** — 4-row symptom table → troubleshooting
- **Continue to → Reading the screen**

### 7.2 Reading the screen

- **1 The four panes** — ◆ D4 wireframe
- **2 The map** — what the glyphs mean *at the level you need on night one* (villagers,
  terrain, fire, structures); everything else → link out
  - ▸ Tip — `c` recenters, arrows pan, the legend is always right there
- **3 The chronicle** — the narrated feed *is* the game's output; `r` toggles raw
  - ▸ Note — this is how you catch up on hours you missed
- **4 The dock tabs** — one row each: chronicle / guardian / villagers / systems
- **5 Time controls** — `space`, `[`, `]`
  - ▸ Pitfall — max speed does not make villagers think faster. → *Speed & thinking*
  - ◆ (D8 lives on the concept page; link, don't duplicate)
- **6 The one key to remember** — `?` opens help
- **Continue to → Your first night**

### 7.3 Your first night

The chapter that mirrors RimWorld's "Your first day / Your first night" — highest
tension, highest teaching value.

- **1 Watch first, act second** — 10 minutes at normal speed, doing nothing
- **2 A day has a shape** — ◆ D5 timeline
- **3 What your villagers are worried about** — warmth, food, rest → *Needs & neglect*
- **4 Nightfall** — the gru
  - What it does and doesn't do (a healthy villager survives; a worn-down one may not)
  - ▸ Warning — the first-night death is the classic new-player loss
- **5 Your first words to the Guardian**
  - 5.1 Press `m`, ask "who is struggling tonight?"
  - 5.2 What it can do at The Voice: omens, visions, standing orders, designations
  - 5.3 ▸ Pitfall — telling it *what you want to be true*, not *what to watch for*
  - 5.4 Worked example, verbatim: a standing order that watches for a cold villager
- **6 Morning** — read the chronicle, find your own night in it
- **Continue to → Your first week**

### 7.4 Your first week

- **1 Food and warmth, in that order**
- **2 Standing orders — telling it to watch while you're away**
  - Worked examples ×3, each with the exact phrasing and what happened
  - ▸ Pitfall — orders that can never match; orders that fire constantly
- **3 Designations and directives — pointing at the map**
- **4 Charges: why you can't just fix everything** — ◆ (charge/faith loop; small inline
  SVG or a table) → *The Guardian*
  - ▸ Design intent — scarcity is what makes the choice a choice
- **5 The village legislates itself** — the meeting, norms, votes; you are not invited
  - ▸ Note — `village_charter.md` is written by *them*, not you
- **6 Living with an always-on world** — detach, come back, catch up on the chronicle
- **7 When it all goes right** — what a stable week looks like
- **Continue to → When someone dies**

### 7.5 When someone dies

- **1 It was never bad luck** — ◆ D6 neglect chain
- **2 The morgue** — what's in it, how to read it
- **3 The postmortem** — takes over the screen at run end; `p` reopens
- **4 The run ended. Now what?**
  - The world stays on disk, read-only, forever
  - Fork it and try the same seed differently → *What next*
  - ▸ Design intent — the archive is the point
- **Continue to → The four stages**

### 7.6 The four stages

The guide's actual spine — this is what "learned everything" means.

- **1 The ladder** — ◆ D7
  - ▸ Note — a stage is an identity, not a difficulty
  - ▸ Note — stage is fixed at world creation and never changes for that world
- **2 Stage 1 — The Voice** *(you speak, it acts)*
  - What you have; the skill being taught (asking well); what proves you've got it
- **3 Stage 2 — The Written Word** *(your law outlives the conversation)*
  - `charter.md` — the one file the game lets you write
  - Worked example: a charter that changes behaviour, before/after
  - ▸ Pitfall — a charter that contradicts itself
- **4 Stage 3 — The Craft** *(you shape what it can do)*
  - Skill files and the capability manifest
- **5 Stage 4 — The Stewardship** *(a world in your care)*
  - Graduation: everything unlocked, nothing above it
- **6 How you climb** — exercises, evidence, `promptworld stages`
  - ▸ Note — you can skip ahead with an override; the game will tell you what you're
    skipping first
- **Continue to → What next**

### 7.7 What next

Deliberately short. A launchpad, not a chapter.

- **1 Scenarios** — the nine seeded exercises
- **2 Forking and comparing runs**
- **3 Tuning a world** — `tuning.json`
- **4 Re-skinning the fiction** — `skin.json`
- **5 Cloud narration** — turning on the good chronicle
- **6 Where the real detail lives** — the four reference pages in `docs/player/`
- **7 Reading the source** — one line pointing at `docs/wiki/`

---

## 8. Explicitly out of scope

- Event-type catalogues, tool registries, provider chains — `docs/wiki/` owns those.
- Modding, bundles, Starlark tools.
- Any per-key or per-command exhaustive listing — `docs/player/` owns those.
- Anything that would make a chapter exceed its target by more than 50%.

---

## 9. Build plan

1. Confirm the open facts flagged inline (default model; current stage-1 tool grant;
   exercise names) against the code — not against these docs.
2. Write `docs/guide/_shared.css` (or an inlined shared block) from `docs/player/`'s
   token set + the wiki furniture in §5.
3. Author the hub + chapters 1–3 first; review the design system against real content
   before writing 4–7. The callout/diagram vocabulary is cheap to change now and
   expensive later.
4. Author the concept pages only after the chapters exist and the reference counts in
   §6 are confirmed by actual links.
5. Word-count check per page against §4 before opening the PR.

---

## 10. Open questions for the operator

1. **Board card?** Under this project's PDLC this should be a TASK with its own branch
   and PR before authoring begins. This outline is a proposal, not a claim on a card.
2. **HTML or Markdown?** Outlined as HTML for parity with `docs/player/` (and so the
   SVG diagrams and callouts render for a non-technical reader with no toolchain).
   Markdown would be lighter to maintain but loses the wiki furniture.
3. **Chapter 6 length.** The four-stage chapter is the one most likely to want its own
   sub-pages later. Ship it as one page and split when it earns it, or pre-split?

[Quickstart Guides]: https://rimworldwiki.com/wiki/Quickstart_Guides
[DF2014 Quickstart guide]: https://dwarffortresswiki.org/index.php/DF2014:Quickstart_guide

---
title: Recurring Interface Patterns
aliases: [Cross-Game UI Patterns]
tags: [game-ui, patterns, cross-cutting]
type: note
created: 2026-07-25
updated: 2026-07-25
related: ["[[Game-UI-UX]]", "[[Dwarf-Fortress-Interface]]", "[[RimWorld-Interface]]", "[[Terraria-Interface]]", "[[LLM-Agent-Sim-Interfaces]]", "[[Chat-and-Agent-Console-Rendering]]", "[[TUI-and-Roguelike-UI-Craft]]"]
---

# Recurring Interface Patterns

Patterns that appear independently across the surveyed interfaces — recorded here descriptively
(where each is documented), as a map for the later analysis. Each row points into the per-source
notes and [[_grounding]].

## Event feeds with severity coding and jump-to-source

- RimWorld: right-edge letters color-coded blue/grey/yellow/red; notifications click to
  center the camera on their subject ([[RimWorld-Interface]]).
- Dwarf Fortress: top-left announcements; major ones pause and center the camera; dated list
  with zoom-to-event ([[Dwarf-Fortress-Interface]]).
- Cogmind/NetHack/DCSS: small importance-filtered logs, `--More--` gating, message
  coalescing, per-message user filtering ([[TUI-and-Roguelike-UI-Craft]]).
- Claude Code: collapsed tool-call lines expandable to a full transcript viewer
  ([[Chat-and-Agent-Console-Rendering]]).

## Progressive disclosure of entity internals

- Smallville: emoji bubble → click avatar → full action text → memory stream
  ([[LLM-Agent-Sim-Interfaces]]).
- RimWorld: colonist bar glyph → inspect pane → needs/health tabs with itemized numeric
  thoughts ([[RimWorld-Interface]]).
- DF: map glyph → `k` look cursor tile hierarchy → per-dwarf thought/relationship screens
  ([[Dwarf-Fortress-Interface]]).
- Chat UIs: collapsed "Thinking"/tool blocks expandable on demand
  ([[Chat-and-Agent-Console-Rendering]]).

## Time control as a first-class UI element

- RimWorld: pause/x1/x3/x6 with tick math documented; pause-on-major-event
  ([[RimWorld-Interface]]).
- DF: major announcements pause the game ([[Dwarf-Fortress-Interface]]).
- Smallville: replay playback speeds 1–5; AI Town: world freeze button + idle auto-pause
  (cost-driven) ([[LLM-Agent-Sim-Interfaces]]).

## Teaching in context rather than up front

- RimWorld's learning helper: "never … a lesson you already know, but always … the lessons
  you need to know now" ([[RimWorld-Interface]]).
- Terraria's Guide NPC, dynamic help, staged Bestiary unlocks ([[Terraria-Interface]]).
- Cogmind's context-sensitive help over front-loaded docs ([[TUI-and-Roguelike-UI-Craft]]).
- DF's absence of this for 16 years is the documented counter-case, and its 2022 tutorial is
  still reviewed as "barebones" ([[Dwarf-Fortress-Interface]]).

## Input parity and inconsistency costs

- Cogmind's hard rule: everything reachable by both keyboard and mouse, hotkeys embedded
  visibly ([[TUI-and-Roguelike-UI-Craft]]).
- DF classic's inconsistent scrolling ("three different ways") is the documented negative
  case; the Premium redesign is documented as losing keyboard efficiency while gaining mouse
  access — both directions drew criticism ([[Dwarf-Fortress-Interface]]).
- Terraria's console/mobile ports shipped *two* cursor modes because neither alone sufficed
  ([[Terraria-Interface]]).

## Automation of tedium as interface design

- DCSS: "all tedious, but necessary, chores should be automated" (autoexplore, autofight)
  ([[TUI-and-Roguelike-UI-Craft]]).
- Terraria: quick stack / deposit all / loot all / sorting buttons ([[Terraria-Interface]]).
- RimWorld and DF: the most popular mods automate or batch what the native UI makes manual
  (Colony Manager; DFHack autofarm, labormanager) ([[RimWorld-Interface]], [[Dwarf-Fortress-Interface]]).

## The community-mod layer as a UX gap detector

- DF: Dwarf Therapist ("makes Dwarf Fortress playable"), DFHack manipulator
  ([[Dwarf-Fortress-Interface]]).
- RimWorld: search menus (Dubs Mint), denser HUDs (RimHUD), log-to-map surfacing
  (Interaction Bubbles) ([[RimWorld-Interface]]).
- Terraria: inventory/crafting mods rendered obsolete when the official UI absorbed their
  features ([[Terraria-Interface]]).

## Streaming/liveness as trust and cost signals

- Chat UIs: token streaming as activity/trust signal, with documented overload failure modes
  ([[Chat-and-Agent-Console-Rendering]]).
- LLM sims: replay-vs-live is a visible cost decision; freeze buttons and idle pauses
  ([[LLM-Agent-Sim-Interfaces]]).

## Scale thresholds where per-entity UI breaks down

- DF/RimWorld: late-game unit management is the documented pain point third-party grids fix
  ([[Dwarf-Fortress-Interface]], [[RimWorld-Interface]]).
- LLM sims: per-agent inspection documented at 25 agents; aggregate dashboards at 500–1,000+
  (Project Sid) ([[LLM-Agent-Sim-Interfaces]]).

## Grounding

- [[_grounding]] — all sections; each pattern above cites its per-source note, which cites
  the grounding directly.
